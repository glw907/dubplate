// Package conformance checks a recorded JSON response against an OpenAPI
// document.
//
// The caller supplies the document's bytes; the package reads no file and
// embeds none. A check reports one of five outcomes, and only [Valid] is a
// pass. kin-openapi returns nil, which reads as a pass, for a HEAD request,
// for a handful of redirect statuses, and for a response with no schema to
// validate against. The checker classifies each of those itself, before
// kin-openapi runs, so a silent pass is reported as [NoSchema].
package conformance

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"
)

// Outcome is the result class of one check. Only [Valid] is a pass.
type Outcome string

const (
	// Valid means the body satisfies the document's schema.
	Valid Outcome = "valid"
	// Invalid means the body violates the schema, is empty, is not JSON,
	// or arrived under a content type the document does not declare.
	Invalid Outcome = "invalid"
	// NoRoute means the document declares no operation for the method
	// and path.
	NoRoute Outcome = "no-route"
	// UndeclaredStatus means the operation declares neither the status nor
	// a default response.
	UndeclaredStatus Outcome = "undeclared-status"
	// NoSchema means nothing was validated: the response has no content
	// entry or no schema, or the request was a HEAD or the status one
	// kin-openapi never validates.
	NoSchema Outcome = "no-schema"
)

// Violation is one place a body departs from its schema.
type Violation struct {
	// Pointer is the RFC 6901 JSON pointer of the offending value, empty
	// for the document root.
	Pointer string
	Reason  string
}

// Result is the outcome of a check; Violations is set only for [Invalid].
type Result struct {
	Outcome    Outcome
	Violations []Violation
}

// Checker validates responses against one OpenAPI document.
type Checker struct {
	router routers.Router
}

// New loads the OpenAPI document in document. Its servers are cleared, so
// a check routes by path alone.
func New(document []byte) (*Checker, error) {
	doc, err := openapi3.NewLoader().LoadFromData(document)
	if err != nil {
		return nil, fmt.Errorf("load openapi document: %w", err)
	}
	// A server URL of "{url}" consumes input up to the next "/", so a full
	// URL never matches a path such as /rest/ping.
	doc.Servers = nil
	router, err := legacy.NewRouter(doc)
	if err != nil {
		return nil, fmt.Errorf("route openapi document: %w", err)
	}
	return &Checker{router: router}, nil
}

// Check validates one response to the operation at method and path.
// contentType is the response's Content-Type header value.
func (c *Checker) Check(ctx context.Context, method, path string, status int, contentType string, body []byte) Result {
	if method == http.MethodHead || neverValidated(status) {
		return Result{Outcome: NoSchema}
	}
	req := &http.Request{Method: method, URL: &url.URL{Path: path}}
	route, pathParams, err := c.router.FindRoute(req)
	if err != nil {
		return Result{Outcome: NoRoute}
	}
	responses := route.Operation.Responses
	ref := responses.Status(status)
	if ref == nil {
		ref = responses.Default()
	}
	if ref == nil {
		return Result{Outcome: UndeclaredStatus}
	}
	if ref.Value == nil || len(ref.Value.Content) == 0 {
		return Result{Outcome: NoSchema}
	}
	media := ref.Value.Content.Get(contentType)
	if media == nil {
		return invalid(Violation{Reason: fmt.Sprintf("content type %q is not declared for the response", contentType)})
	}
	if media.Schema == nil {
		return Result{Outcome: NoSchema}
	}
	if len(body) == 0 {
		return invalid(Violation{Reason: "empty body"})
	}
	var decoded any
	if err := json.Unmarshal(body, &decoded); err != nil {
		return invalid(Violation{Reason: fmt.Sprintf("body is not JSON: %v", err)})
	}

	input := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request: req, PathParams: pathParams, Route: route,
		},
		Status:  status,
		Header:  http.Header{"Content-Type": []string{contentType}},
		Options: &openapi3filter.Options{MultiError: true, IncludeResponseStatus: true},
	}
	input.SetBodyBytes(body)
	err = openapi3filter.ValidateResponse(ctx, input)
	if err == nil {
		return Result{Outcome: Valid}
	}
	var respErr *openapi3filter.ResponseError
	if !errors.As(err, &respErr) || respErr.Err == nil {
		return invalid(Violation{Reason: err.Error()})
	}
	var out []Violation
	collect(respErr.Err, claimedStatus(decoded), &out)
	if len(out) == 0 {
		out = append(out, Violation{Reason: respErr.Reason})
	}
	slices.SortFunc(out, func(a, b Violation) int {
		return cmp.Or(cmp.Compare(a.Pointer, b.Pointer), cmp.Compare(a.Reason, b.Reason))
	})
	out = slices.Compact(out)
	return Result{Outcome: Invalid, Violations: out}
}

func invalid(v ...Violation) Result { return Result{Outcome: Invalid, Violations: v} }

// claimedStatus returns the status the body's subsonic-response member
// declares, or "" when it declares none.
func claimedStatus(decoded any) string {
	root, _ := decoded.(map[string]any)
	resp, _ := root["subsonic-response"].(map[string]any)
	status, _ := resp["status"].(string)
	return status
}

// neverValidated reports whether kin-openapi skips validation for status.
func neverValidated(status int) bool {
	switch status {
	case http.StatusMovedPermanently, http.StatusNotModified,
		http.StatusTemporaryRedirect, http.StatusPermanentRedirect:
		return true
	}
	return false
}

// collect appends a violation for every leaf of err. A failed oneOf
// reports one error per branch; when the body claims a status, only the
// branches whose status enum holds it are kept, so a success body is not
// also reported against the failure branch.
func collect(err error, status string, out *[]Violation) {
	switch e := err.(type) {
	case openapi3.MultiError:
		for _, sub := range e {
			collect(sub, status, out)
		}
	case *openapi3.SchemaError:
		var branches openapi3.MultiError
		if e.Origin != nil && errors.As(e.Origin, &branches) {
			for _, b := range pickBranches(e.Schema, branches, status) {
				collect(b, status, out)
			}
			return
		}
		*out = append(*out, Violation{Pointer: pointer(e.JSONPointer()), Reason: cmp.Or(e.Reason, e.SchemaField)})
	default:
		*out = append(*out, Violation{Reason: err.Error()})
	}
}

// pickBranches filters the per-branch errors of a failed oneOf. Entry i
// belongs to OneOf[i] only when every branch failed, which holds exactly
// when the counts agree.
func pickBranches(schema *openapi3.Schema, branches openapi3.MultiError, status string) []error {
	if status == "" || schema == nil || len(branches) != len(schema.OneOf) {
		return branches
	}
	var kept []error
	for i, b := range branches {
		if branch := schema.OneOf[i].Value; branch != nil && slices.Contains(statusEnum(branch), any(status)) {
			kept = append(kept, b)
		}
	}
	if len(kept) == 0 {
		return branches
	}
	return kept
}

// statusEnum returns the enum of the schema's status property, looking
// through allOf members.
func statusEnum(s *openapi3.Schema) []any {
	if p := s.Properties["status"]; p != nil && p.Value != nil && len(p.Value.Enum) > 0 {
		return p.Value.Enum
	}
	for _, m := range s.AllOf {
		if m.Value == nil {
			continue
		}
		if enum := statusEnum(m.Value); enum != nil {
			return enum
		}
	}
	return nil
}

var pointerEscaper = strings.NewReplacer("~", "~0", "/", "~1")

func pointer(path []string) string {
	var b strings.Builder
	for _, seg := range path {
		b.WriteByte('/')
		b.WriteString(pointerEscaper.Replace(seg))
	}
	return b.String()
}
