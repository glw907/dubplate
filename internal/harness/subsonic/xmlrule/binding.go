package xmlrule

import (
	"errors"
	"fmt"
	"maps"
	"mime"
	"slices"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// envelopeMember names the envelope: the schema's root element, and the
// property an OpenAPI response schema wraps it in.
const envelopeMember = "subsonic-response"

// binding is what the OpenAPI document says one endpoint answers.
type binding struct {
	// payloads maps each payload member a success may carry to whether
	// the success schema requires it. It is empty for an endpoint whose
	// success carries no payload.
	payloads map[string]bool
	// statuses holds the HTTP statuses the operation declares; anyStatus
	// is set when it declares a default response.
	statuses  map[int]bool
	anyStatus bool
	// binary is set when the operation's success is not an envelope, so
	// only an error envelope arrives as XML.
	binary bool
}

// readBindings reads one binding per /rest/<endpoint> path of doc, from
// the path's GET operation, or its POST when it has no GET.
// envelopeNames holds the names an envelope carries as attributes, which
// are never a payload.
func readBindings(doc *openapi3.T, envelopeNames map[string]bool) (map[string]binding, error) {
	if doc.Paths == nil {
		return nil, errors.New("openapi document declares no paths")
	}
	out := make(map[string]binding)
	for path, item := range doc.Paths.Map() {
		endpoint, ok := strings.CutPrefix(path, "/rest/")
		if !ok {
			continue
		}
		op := item.Get
		if op == nil {
			op = item.Post
		}
		if op == nil || op.Responses == nil {
			continue
		}
		b, err := readBinding(op.Responses, envelopeNames)
		if err != nil {
			return nil, fmt.Errorf("openapi path %s: %w", path, err)
		}
		out[endpoint] = b
	}
	return out, nil
}

func readBinding(responses *openapi3.Responses, envelopeNames map[string]bool) (binding, error) {
	b := binding{statuses: make(map[int]bool), payloads: make(map[string]bool), binary: true}
	for key, ref := range responses.Map() {
		if key == "default" {
			b.anyStatus = true
		} else {
			code, err := strconv.Atoi(key)
			if err != nil {
				return binding{}, fmt.Errorf("response key %q is not a status", key)
			}
			b.statuses[code] = true
		}
		if key != "200" || ref == nil || ref.Value == nil {
			continue
		}
		for _, media := range ref.Value.Content {
			if media.Schema == nil || media.Schema.Value == nil {
				continue
			}
			env, ok := media.Schema.Value.Properties[envelopeMember]
			if !ok || env.Value == nil {
				continue
			}
			b.binary = false
			success := successBranch(env.Value)
			if success == nil {
				return binding{}, fmt.Errorf("%s names no branch whose status admits ok", envelopeMember)
			}
			payloadsOf(success, envelopeNames, b.payloads)
		}
	}
	return b, nil
}

// successBranch returns the oneOf branch of an envelope schema whose
// status enum admits "ok", or the schema itself when it has no oneOf.
func successBranch(env *openapi3.Schema) *openapi3.Schema {
	if len(env.OneOf) == 0 {
		return env
	}
	for _, branch := range env.OneOf {
		if branch.Value != nil && admitsOK(branch.Value) {
			return branch.Value
		}
	}
	return nil
}

func admitsOK(sc *openapi3.Schema) bool {
	if status, ok := sc.Properties["status"]; ok && status.Value != nil && slices.Contains(status.Value.Enum, any("ok")) {
		return true
	}
	for _, part := range sc.AllOf {
		if part.Value != nil && admitsOK(part.Value) {
			return true
		}
	}
	return false
}

// payloadsOf adds the success schema's payload members to out, each
// marked with whether it is required: its properties, following allOf,
// other than the envelope's own names.
func payloadsOf(success *openapi3.Schema, envelopeNames map[string]bool, out map[string]bool) {
	required := requiredOf(success)
	for name := range properties(success) {
		if !envelopeNames[name] {
			out[name] = required[name]
		}
	}
}

// requiredOf returns the names a schema requires, following allOf.
func requiredOf(sc *openapi3.Schema) map[string]bool {
	names := make(map[string]bool)
	for _, name := range sc.Required {
		names[name] = true
	}
	for _, part := range sc.AllOf {
		if part.Value != nil {
			maps.Copy(names, requiredOf(part.Value))
		}
	}
	return names
}

// envelope is what the binding reads from a body: the root's status
// attribute and the names of its child elements, in order.
type envelope struct {
	status   string
	children []string
}

// check returns the binding's violations for one response, each at the
// envelope's ordinal.
func (b binding) check(endpoint string, status int, contentType string, env envelope, choice map[string]bool) []Violation {
	var out []Violation
	add := func(format string, args ...any) {
		out = append(out, Violation{Ordinal: 1, Message: fmt.Sprintf(format, args...), Source: Binding})
	}
	if !b.anyStatus && !b.statuses[status] {
		add("HTTP status %d is not one %s declares", status, endpoint)
	}
	if !xmlMedia(contentType) {
		add("content type %q is not application/xml or text/xml", contentType)
	}
	switch env.status {
	case "failed":
		if !slices.Contains(env.children, "error") {
			add("a failed envelope carries no error element")
		}
	case "ok":
		if b.binary {
			add("%s answers success with a binary body, never an ok envelope", endpoint)
			break
		}
		for _, name := range env.children {
			if choice[name] {
				if _, ok := b.payloads[name]; !ok {
					add("an ok envelope for %s carries %s, which is not its payload", endpoint, name)
				}
			}
		}
		for _, name := range slices.Sorted(maps.Keys(b.payloads)) {
			if b.payloads[name] && !slices.Contains(env.children, name) {
				add("an ok envelope for %s lacks its payload element %s", endpoint, name)
			}
		}
	}
	return out
}

// unvalidated returns the payload element an ok envelope carries that the
// schema's choice does not declare, or "" when there is none.
func (b binding) unvalidated(env envelope, choice map[string]bool) string {
	if env.status != "ok" {
		return ""
	}
	for _, name := range env.children {
		if _, ok := b.payloads[name]; ok && !choice[name] {
			return name
		}
	}
	return ""
}

// xmlMedia reports whether contentType names an XML media type.
func xmlMedia(contentType string) bool {
	media, _, err := mime.ParseMediaType(contentType)
	return err == nil && (media == "application/xml" || media == "text/xml")
}
