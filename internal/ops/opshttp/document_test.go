package opshttp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	"github.com/getkin/kin-openapi/routers/legacy"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
)

// The document and its tests are the one place kin-openapi appears: its
// silent passes and its 2020-12 fallback are closed here so every other
// package stays clear of it. A response is validated through
// openapi3filter with IncludeResponseStatus set, and every other pass
// the library gives silently (a HEAD, a redirect, an answer with no
// content, content with no schema) is a failure of the test.

// A document is the loaded contract with its router.
type document struct {
	doc    *openapi3.T
	router routers.Router
}

// loadDocument loads openapi.yaml. With closed set, every object schema
// a response carries refuses fields it does not declare, so a Go field
// the document omits fails validation; the published document leaves
// responses open, because its version policy tells a client to ignore an
// unknown field.
func loadDocument(t *testing.T, closed bool) *document {
	t.Helper()
	data, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("read the contract document: %v", err)
	}
	doc, err := openapi3.NewLoader().LoadFromData(data)
	if err != nil {
		t.Fatalf("load the contract document: %v", err)
	}
	if err := doc.Validate(t.Context()); err != nil {
		t.Fatalf("the contract document is not a valid OpenAPI document: %v", err)
	}
	doc.Servers = nil
	if closed {
		seen := map[*openapi3.Schema]bool{}
		for _, item := range doc.Paths.Map() {
			for _, op := range item.Operations() {
				for _, resp := range op.Responses.Map() {
					for _, media := range resp.Value.Content {
						closeObjects(media.Schema, seen)
					}
				}
			}
		}
	}
	router, err := legacy.NewRouter(doc)
	if err != nil {
		t.Fatalf("route the contract document: %v", err)
	}
	return &document{doc: doc, router: router}
}

// closeObjects sets additionalProperties false on every object schema
// reachable from ref through properties and items.
func closeObjects(ref *openapi3.SchemaRef, seen map[*openapi3.Schema]bool) {
	if ref == nil || ref.Value == nil || seen[ref.Value] {
		return
	}
	s := ref.Value
	seen[s] = true
	if s.Type.Is("object") {
		no := false
		s.AdditionalProperties = openapi3.AdditionalProperties{Has: &no}
	}
	for _, p := range s.Properties {
		closeObjects(p, seen)
	}
	closeObjects(s.Items, seen)
	for _, group := range []openapi3.SchemaRefs{s.OneOf, s.AnyOf, s.AllOf} {
		for _, branch := range group {
			closeObjects(branch, seen)
		}
	}
}

// operation returns the document's POST operation for name.
func (d *document) operation(t *testing.T, name ops.Op) *openapi3.Operation {
	t.Helper()
	item := d.doc.Paths.Value("/" + string(name))
	if item == nil || item.Post == nil {
		t.Fatalf("the document has no POST /%s", name)
	}
	return item.Post
}

// redirectStatuses are the statuses openapi3filter never validates.
var redirectStatuses = []int{
	http.StatusMovedPermanently, http.StatusNotModified,
	http.StatusTemporaryRedirect, http.StatusPermanentRedirect,
}

// checkResponse validates one answer to the request a client sent for
// name against the document. Every silent pass is an error here.
func (d *document) checkResponse(ctx context.Context, name ops.Op, rec *httptest.ResponseRecorder) error {
	docReq := httptest.NewRequest(http.MethodPost, "/"+string(name), nil)
	route, params, err := d.router.FindRoute(docReq)
	if err != nil {
		return fmt.Errorf("no route for %s: %w", name, err)
	}
	switch {
	case slices.Contains(redirectStatuses, rec.Code):
		return fmt.Errorf("status %d is one openapi3filter never validates", rec.Code)
	case rec.Body.Len() == 0:
		return errors.New("empty body: openapi3filter passes content it cannot read")
	}
	// A status the operation does not declare is left to
	// IncludeResponseStatus below: without it openapi3filter passes it.
	if declared := route.Operation.Responses.Status(rec.Code); declared != nil {
		media := declared.Value.Content.Get("application/json")
		if media == nil || media.Schema == nil {
			return fmt.Errorf("%s %d declares no JSON schema: openapi3filter would pass it silently", name, rec.Code)
		}
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		return fmt.Errorf("Content-Type = %q, want application/json", got)
	}
	in := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{Request: docReq, PathParams: params, Route: route},
		Status:                 rec.Code,
		Header:                 rec.Header(),
		Options:                &openapi3filter.Options{IncludeResponseStatus: true},
	}
	in.SetBodyBytes(rec.Body.Bytes())
	return openapi3filter.ValidateResponse(ctx, in)
}

// checkRequest validates a request body sent to name against the
// document.
func (d *document) checkRequest(ctx context.Context, name ops.Op, body string) error {
	req := httptest.NewRequest(http.MethodPost, "/"+string(name), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	route, params, err := d.router.FindRoute(req)
	if err != nil {
		return fmt.Errorf("no route for %s: %w", name, err)
	}
	return openapi3filter.ValidateRequest(ctx, &openapi3filter.RequestValidationInput{
		Request: req, PathParams: params, Route: route, Options: &openapi3filter.Options{},
	})
}

// operationNames is every operation the mount routes: the keys of its
// route table, so an operation added there without a path in the
// document fails the paths test.
func operationNames(t *testing.T) []ops.Op {
	t.Helper()
	f := newFixture(t, admitAnn())
	names := slices.Collect(maps.Keys(f.m.routes))
	slices.Sort(names)
	return names
}

// TestRoutesCoverThePrivilegeTable proves the mount routes exactly the
// operations package ops declares, each of which carries a privilege,
// so a name the table admits always has a route behind it.
func TestRoutesCoverThePrivilegeTable(t *testing.T) {
	var declared []ops.Op
	for _, file := range []string{"../admin.go", "../submissions.go"} {
		for _, name := range goConstants(t, file, "Op") {
			declared = append(declared, ops.Op(name))
		}
	}
	slices.Sort(declared)
	if !slices.Contains(declared, ops.OpListMembers) || !slices.Contains(declared, ops.OpCancelSubmission) {
		t.Fatalf("read %v from package ops, want the Op constants of both files", declared)
	}
	if got := operationNames(t); !slices.Equal(got, declared) {
		t.Errorf("routes = %v, want %v", got, declared)
	}
	for _, op := range declared {
		if _, ok := ops.PrivilegeOf(op); !ok {
			t.Errorf("PrivilegeOf(%q) answers no privilege", op)
		}
	}
}

// requestSamples is a well-formed request body for each operation.
var requestSamples = map[ops.Op]string{
	ops.OpListMembers:  `{}`,
	ops.OpAddMember:    `{"address":"bo@example.com","display_name":"Bo","username":"bo"}`,
	ops.OpRemoveMember: `{"user_id":"` + testHandle + `"}`,
	ops.OpDemoteMember: `{"user_id":"` + testHandle + `"}`,
	ops.OpListSettings: `{}`,
	ops.OpSetSetting:   `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`,
	ops.OpResetSetting: `{"key":"DUBPLATE_REAUTH_WINDOW"}`,

	ops.OpCreateSubmission: `{"verdict_unit":"album","container":"files","files":[{"path":"a/1.flac","size":10}],"source":"vinyl","source_note":"n","store_url":"https://x.example/a"}`,
	ops.OpListSubmissions:  `{}`,
	ops.OpGetSubmission:    `{"submission_id":"` + testHandle + `"}`,
	ops.OpRetrySubmission:  `{"submission_id":"` + testHandle + `"}`,
	ops.OpCancelSubmission: `{"submission_id":"` + testHandle + `"}`,
}

// TestDocument_PathsMatchTheCatalogBothWays proves every operation the
// mount routes has a POST path in the document and every path names an
// operation (AC1).
func TestDocument_PathsMatchTheCatalogBothWays(t *testing.T) {
	d := loadDocument(t, false)
	want := operationNames(t)
	for _, name := range want {
		if item := d.doc.Paths.Value("/" + string(name)); item == nil || item.Post == nil {
			t.Errorf("operation %q has no POST path in the document", name)
		}
	}
	for path, item := range d.doc.Paths.Map() {
		if name := ops.Op(strings.TrimPrefix(path, "/")); !slices.Contains(want, name) {
			t.Errorf("document path %q names no operation", path)
		}
		if got := len(item.Operations()); got != 1 || item.Post == nil {
			t.Errorf("document path %q carries %d operations, want one POST", path, got)
		}
	}
	if got := d.doc.Paths.Len(); got != len(want) {
		t.Errorf("document has %d paths, want %d", got, len(want))
	}
}

// allowedKeywords are the schema keywords the built-in validator
// enforces, plus the annotations that validate nothing and $ref. A
// 3.1-only keyword outside this set validates nothing once openapi3's
// 2020-12 compile fails on a nested $ref and falls back silently.
var allowedKeywords = []string{
	"type", "const", "enum", "required", "properties", "additionalProperties", "items", "prefixItems",
	"oneOf", "anyOf", "allOf", "not", "minimum", "maximum", "minLength", "maxLength", "pattern",
	"minItems", "maxItems", "uniqueItems", "format",
	"$ref", "description", "title", "example", "deprecated",
}

// allowedFormats are the five formats the built-in validator
// registers; it passes any other silently.
var allowedFormats = []string{"byte", "date", "date-time", "int32", "int64"}

// schemaKeywordFindings walks every schema reachable from ref and
// reports each keyword or format outside the allowed sets.
func schemaKeywordFindings(ref *openapi3.SchemaRef, where string, seen map[*openapi3.Schema]bool) []string {
	if ref == nil || ref.Value == nil || seen[ref.Value] {
		return nil
	}
	s := ref.Value
	seen[s] = true
	var out []string
	rv := reflect.ValueOf(*s)
	for i := range rv.NumField() {
		field := rv.Type().Field(i)
		tag, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if tag == "" || tag == "-" || rv.Field(i).IsZero() {
			continue
		}
		if !slices.Contains(allowedKeywords, tag) {
			out = append(out, fmt.Sprintf("%s: keyword %q", where, tag))
		}
	}
	if len(s.Extensions) > 0 {
		out = append(out, fmt.Sprintf("%s: unknown keys %v", where, s.Extensions))
	}
	if s.Format != "" && !slices.Contains(allowedFormats, s.Format) {
		out = append(out, fmt.Sprintf("%s: format %q", where, s.Format))
	}
	for name, p := range s.Properties {
		out = append(out, schemaKeywordFindings(p, where+"."+name, seen)...)
	}
	out = append(out, schemaKeywordFindings(s.Items, where+"[]", seen)...)
	if s.AdditionalProperties.Schema != nil {
		out = append(out, schemaKeywordFindings(s.AdditionalProperties.Schema, where+".*", seen)...)
	}
	for _, group := range []openapi3.SchemaRefs{s.OneOf, s.AnyOf, s.AllOf, s.PrefixItems} {
		for i, branch := range group {
			out = append(out, schemaKeywordFindings(branch, fmt.Sprintf("%s|%d", where, i), seen)...)
		}
	}
	out = append(out, schemaKeywordFindings(s.Not, where+".not", seen)...)
	return out
}

// documentSchemas lists every schema the document names, components and
// inline, with a label.
func documentSchemas(d *document) map[string]*openapi3.SchemaRef {
	out := map[string]*openapi3.SchemaRef{}
	for name, ref := range d.doc.Components.Schemas {
		out["components.schemas."+name] = ref
	}
	for path, item := range d.doc.Paths.Map() {
		for _, op := range item.Operations() {
			if op.RequestBody != nil {
				for _, media := range op.RequestBody.Value.Content {
					out[path+" request"] = media.Schema
				}
			}
			for status, resp := range op.Responses.Map() {
				for _, media := range resp.Value.Content {
					out[path+" "+status] = media.Schema
				}
			}
		}
	}
	return out
}

// TestDocument_SchemaKeywordsAndFormats proves the document uses only
// keywords the validator enforces and one of its five formats (AC1),
// and that the walk can fail: a schema carrying an if keyword and an
// unregistered format is reported.
func TestDocument_SchemaKeywordsAndFormats(t *testing.T) {
	d := loadDocument(t, false)
	schemas := documentSchemas(d)
	if len(schemas) < 20 {
		t.Fatalf("walked %d schemas, want the document's whole set", len(schemas))
	}
	seen := map[*openapi3.Schema]bool{}
	for where, ref := range schemas {
		for _, finding := range schemaKeywordFindings(ref, where, seen) {
			t.Error(finding)
		}
	}

	bad := &openapi3.SchemaRef{Value: &openapi3.Schema{
		Type:   &openapi3.Types{"string"},
		Format: "email",
		If:     &openapi3.SchemaRef{Value: &openapi3.Schema{Type: &openapi3.Types{"string"}}},
	}}
	findings := schemaKeywordFindings(bad, "synthetic", map[*openapi3.Schema]bool{})
	if len(findings) != 2 {
		t.Errorf("the walk reported %v for an if keyword and an email format, want both", findings)
	}
}

// goConstants returns the string values of the exported constants in
// file whose names start with prefix, read with go/parser so the test
// names no value the code does not.
func goConstants(t *testing.T, file, prefix string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		t.Fatalf("parse %s: %v", file, err)
	}
	var out []string
	ast.Inspect(parsed, func(n ast.Node) bool {
		spec, ok := n.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for i, name := range spec.Names {
			lit, isLit := spec.Values[i].(*ast.BasicLit)
			if strings.HasPrefix(name.Name, prefix) && isLit && lit.Kind == token.STRING {
				v, err := strconv.Unquote(lit.Value)
				if err != nil {
					t.Fatalf("%s: %v", name.Name, err)
				}
				out = append(out, v)
			}
		}
		return true
	})
	return out
}

// enumOf returns the enum values at the named schema's property path.
func enumOf(t *testing.T, d *document, schema string, path ...string) []string {
	t.Helper()
	ref := d.doc.Components.Schemas[schema]
	for _, p := range path {
		ref = ref.Value.Properties[p]
	}
	var out []string
	for _, v := range ref.Value.Enum {
		s, ok := v.(string)
		if !ok {
			t.Fatalf("%s.%v enum value %v is not a string", schema, path, v)
		}
		out = append(out, s)
	}
	return out
}

// sameSet reports whether a and b hold the same strings.
func sameSet(a, b []string) bool {
	x, y := slices.Clone(a), slices.Clone(b)
	slices.Sort(x)
	slices.Sort(y)
	return slices.Equal(x, y)
}

// TestDocument_EnumerationsMatchTheCode proves the wire code enum equals
// the Code constants and the detail enum equals the Detail constants, both
// ways, that remove_member declares its 409, and that the version policy
// names the enumerations it extends (AC1, AC3, AC7).
func TestDocument_EnumerationsMatchTheCode(t *testing.T) {
	d := loadDocument(t, false)

	codes := goConstants(t, "../errors.go", "Code")
	if !slices.Contains(codes, "unauthenticated") {
		t.Fatalf("read %v from errors.go, want the Code constants including unauthenticated", codes)
	}
	if got := enumOf(t, d, "ErrorBody", "error", "code"); !sameSet(got, codes) {
		t.Errorf("code enum = %v, want the Code constants %v", got, codes)
	}

	details := goConstants(t, "../details.go", "Detail")
	if !slices.Contains(details, "last_admin") {
		t.Fatalf("read %v from details.go, want the Detail constants including last_admin", details)
	}
	if got := enumOf(t, d, "ErrorBody", "error", "detail"); !sameSet(got, details) {
		t.Errorf("detail enum = %v, want the Detail constants %v", got, details)
	}

	if d.operation(t, ops.OpRemoveMember).Responses.Status(http.StatusConflict) == nil {
		t.Error("remove_member declares no 409 response")
	}
	policy := d.doc.Info.Description
	for _, phrase := range []string{
		"additive within v1", "wire codes", "error details", "journey vocabulary",
		"treat an enumeration value it does not know as opaque",
	} {
		if !strings.Contains(policy, phrase) {
			t.Errorf("info.description does not say %q", phrase)
		}
	}
	for path, item := range d.doc.Paths.Map() {
		if item.Post.Security != nil && len(*item.Post.Security) != 0 || len(d.doc.Security) != 0 {
			t.Errorf("%s declares a security requirement", path)
		}
	}
}

// jsonFields returns the JSON field names of struct type typ, and which
// of them are always present on the wire.
func jsonFields(typ reflect.Type) (all, required []string) {
	for field := range typ.Fields() {
		name, opts, _ := strings.Cut(field.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			continue
		}
		all = append(all, name)
		if !strings.Contains(opts, "omitempty") {
			required = append(required, name)
		}
	}
	return all, required
}

// TestDocument_ClosedCopyClosesEveryResponseObject proves the loaded
// copy refuses an unknown field on every object schema a response
// carries, the empty answers of remove_member and demote_member
// included: a field added to one of those Go types must fail a
// contract test.
func TestDocument_ClosedCopyClosesEveryResponseObject(t *testing.T) {
	d := loadDocument(t, true)
	var open []string
	seen := map[*openapi3.Schema]bool{}
	var walk func(where string, ref *openapi3.SchemaRef)
	walk = func(where string, ref *openapi3.SchemaRef) {
		if ref == nil || ref.Value == nil || seen[ref.Value] {
			return
		}
		s := ref.Value
		seen[s] = true
		if s.Type.Is("object") && (s.AdditionalProperties.Has == nil || *s.AdditionalProperties.Has) {
			open = append(open, where)
		}
		for name, p := range s.Properties {
			walk(where+"."+name, p)
		}
		walk(where+"[]", s.Items)
	}
	for path, item := range d.doc.Paths.Map() {
		for status, resp := range item.Post.Responses.Map() {
			for _, media := range resp.Value.Content {
				walk(path+" "+status, media.Schema)
			}
		}
	}
	if len(seen) == 0 {
		t.Fatal("walked no response schema")
	}
	if len(open) != 0 {
		t.Errorf("response objects still open in the closed copy: %v", open)
	}
	if empty := d.doc.Components.Schemas["Empty"].Value; empty.AdditionalProperties.Has == nil || *empty.AdditionalProperties.Has {
		t.Error("the Empty response schema is open: a field added to remove_member's answer would pass")
	}
}

// TestDocument_SchemasMatchTheGoTypes proves each wire type's property
// set and required set equal its Go struct's JSON fields, so no Go
// field is missing from the document and none is invented, and that no
// job schema carries a queue position or order (AC7).
func TestDocument_SchemasMatchTheGoTypes(t *testing.T) {
	d := loadDocument(t, false)
	parity := map[string][]any{
		"Empty":                {ops.RemoveMemberResponse{}, ops.DemoteMemberResponse{}},
		"EmptyRequest":         {ops.ListMembersRequest{}, ops.ListSettingsRequest{}, ops.ListSubmissionsRequest{}},
		"Member":               {ops.Member{}},
		"ListMembersResponse":  {ops.ListMembersResponse{}},
		"AddMemberRequest":     {ops.AddMemberRequest{}},
		"AddMemberResponse":    {ops.AddMemberResponse{}},
		"MemberRequest":        {ops.RemoveMemberRequest{}, ops.DemoteMemberRequest{}},
		"SettingEntry":         {ops.SettingEntry{}},
		"ListSettingsResponse": {ops.ListSettingsResponse{}},
		"SetSettingRequest":    {ops.SetSettingRequest{}},
		"ResetSettingRequest":  {ops.ResetSettingRequest{}},
		"SettingWriteResponse": {ops.SetSettingResponse{}, ops.ResetSettingResponse{}},

		"Reason":                   {reject.Reason{}},
		"Rejects":                  {ops.Rejects{}},
		"FileReason":               {ops.FileReason{}},
		"ManifestFile":             {ops.ManifestFile{}},
		"CreateSubmissionRequest":  {ops.CreateSubmissionRequest{}},
		"CreatedFile":              {ops.CreatedFile{}},
		"CreateSubmissionResponse": {ops.CreateSubmissionResponse{}},
		"SubmissionSummary":        {ops.SubmissionSummary{}},
		"ListSubmissionsResponse":  {ops.ListSubmissionsResponse{}},
		"SubmissionRequest":        {ops.GetSubmissionRequest{}, ops.RetrySubmissionRequest{}, ops.CancelSubmissionRequest{}},
		"SubmissionEvent":          {ops.SubmissionEvent{}},
		"SubmissionFile":           {ops.SubmissionFile{}},
		"SubmissionWarning":        {ops.SubmissionWarning{}},
		"SubmissionResponse":       {ops.GetSubmissionResponse{}, ops.RetrySubmissionResponse{}, ops.CancelSubmissionResponse{}},
	}
	// Each of these has no struct of its own: a pattern-checked string, a
	// nullable wrapper, or the error body the next block checks.
	notStructs := []string{"Handle", "UserID", "SubmissionID", "NullableReason", "ErrorBody"}
	for name := range d.doc.Components.Schemas {
		if _, ok := parity[name]; !ok && !slices.Contains(notStructs, name) {
			t.Errorf("document schema %s has no Go type in the parity map", name)
		}
	}
	for schema, goTypes := range parity {
		ref := d.doc.Components.Schemas[schema]
		if ref == nil {
			t.Errorf("document has no schema %s", schema)
			continue
		}
		var props []string
		for name := range ref.Value.Properties {
			props = append(props, name)
		}
		for _, goType := range goTypes {
			all, required := jsonFields(reflect.TypeOf(goType))
			if !sameSet(props, all) {
				t.Errorf("%s properties = %v, want %T's fields %v", schema, props, goType, all)
			}
			if !sameSet(ref.Value.Required, required) {
				t.Errorf("%s required = %v, want %T's always-present fields %v", schema, ref.Value.Required, goType, required)
			}
		}
	}

	errAll, errRequired := jsonFields(reflect.TypeFor[wireError]())
	errSchema := d.doc.Components.Schemas["ErrorBody"].Value.Properties["error"].Value
	var errProps []string
	for name := range errSchema.Properties {
		errProps = append(errProps, name)
	}
	if !sameSet(errProps, errAll) || !sameSet(errSchema.Required, errRequired) {
		t.Errorf("ErrorBody.error = %v required %v, want %v required %v", errProps, errSchema.Required, errAll, errRequired)
	}
}

// A statusCase is how a fake seam produces one declared status.
type statusCase struct {
	status int
	err    error // the seam's answer, nil for 200
}

// errorsByStatus produces each non-success status a seam can answer.
var errorsByStatus = []statusCase{
	{http.StatusBadRequest, ops.NewError(ops.CodeInvalidRequest, "the value is refused").WithDetail(ops.DetailInvalidAddress)},
	{http.StatusForbidden, ops.NewError(ops.CodeReauthenticationRequired, "sign in again")},
	{http.StatusNotFound, ops.NewError(ops.CodeNotFound, "no such member or submission")},
	{http.StatusConflict, ops.NewError(ops.CodeConflict, "the last admin cannot be removed").WithDetail(ops.DetailLastAdmin)},
	{http.StatusInternalServerError, ops.NewError(ops.CodeInternal, "the store failed")},
	{http.StatusServiceUnavailable, ops.NewError(ops.CodeUnavailable, "the write lock is held").WithDetail(ops.DetailBusy)},
	{http.StatusBadRequest, ops.NewError(ops.CodeInvalidRequest, "the manifest was refused").WithRejects(sampleRejects)},
	{http.StatusConflict, ops.NewError(ops.CodeConflict, "uploads are paused").WithRejects(ops.Rejects{
		Submission: []reject.Reason{reject.Of(reject.UploadsOff)}, Files: []ops.FileReason{}})},
	{http.StatusServiceUnavailable, ops.NewError(ops.CodeUnavailable, "the check could not run").WithRejects(ops.Rejects{
		Submission: []reject.Reason{reject.Of(reject.CouldNotCheck)}, Files: []ops.FileReason{}})},
}

// sampleRejects is a refusal carrying a submission reason and a file's.
var sampleRejects = ops.Rejects{
	Submission: []reject.Reason{reject.Of(reject.NoAudio)},
	Files:      []ops.FileReason{{Path: "a/../b.flac", Reason: reject.Of(reject.UnsafePath)}},
}

// exchange drives one request through a mount over fake seams that
// answer err for op (nil for the success every fake already holds), for
// an admin session, and returns the recorded answer.
func exchange(t *testing.T, op ops.Op, body string, err error) *httptest.ResponseRecorder {
	t.Helper()
	f := newFixture(t, grant(ann, ops.PrivilegeAdmin))
	if err != nil {
		switch {
		case f.subs.answers[op] != fakeAnswer{}:
			f.subs.answers[op] = fakeAnswer{err: err}
		default:
			f.admin.answers[op] = fakeAnswer{err: err}
		}
	}
	return f.post(op, body)
}

// TestDocument_EveryDeclaredExchangeValidates proves, for every
// operation, that a well-formed request passes the document's request
// check, the success answer passes its response check, and each other
// status the operation declares is one the projection really answers,
// with a body the document accepts (AC2).
func TestDocument_EveryDeclaredExchangeValidates(t *testing.T) {
	d := loadDocument(t, true)
	ctx := t.Context()
	for _, op := range operationNames(t) {
		t.Run(string(op), func(t *testing.T) {
			if err := d.checkRequest(ctx, op, requestSamples[op]); err != nil {
				t.Errorf("a well-formed request fails the document: %v", err)
			}
			rec := exchange(t, op, requestSamples[op], nil)
			if rec.Code != http.StatusOK {
				t.Fatalf("success answer = %d %s", rec.Code, rec.Body)
			}
			if err := d.checkResponse(ctx, op, rec); err != nil {
				t.Errorf("the success answer fails the document: %v", err)
			}

			declared := d.operation(t, op).Responses.Map()
			answered := map[int]bool{http.StatusOK: true}
			for _, c := range errorsByStatus {
				if _, ok := declared[strconv.Itoa(c.status)]; !ok {
					continue
				}
				rec := exchange(t, op, requestSamples[op], c.err)
				if rec.Code != c.status {
					t.Errorf("a %v answers %d, but the document declares %d for %s", c.err, rec.Code, c.status, op)
					continue
				}
				answered[rec.Code] = true
				if err := d.checkResponse(ctx, op, rec); err != nil {
					t.Errorf("the %d answer fails the document: %v", c.status, err)
				}
			}

			unauth := newFixture(t, func(http.ResponseWriter, *http.Request) (identity.Actor, ops.Privilege, error) {
				return identity.Actor{}, 0, ops.CodeUnauthenticated
			}).post(op, requestSamples[op])
			if err := d.checkResponse(ctx, op, unauth); err != nil {
				t.Errorf("the 401 answer fails the document: %v", err)
			}
			answered[http.StatusUnauthorized] = true

			for status := range declared {
				code, _ := strconv.Atoi(status)
				if !answered[code] {
					t.Errorf("%s declares %d, which no exchange produced", op, code)
				}
			}
		})

	}
}

// TestDocument_RefusalsTheMountMakesItselfValidate proves the answers
// no operation produces (a method other than POST, a content type the
// mount refuses, an unknown operation, a path under api that names no
// v1 operation, a body that is not JSON) carry a body the document's
// error schema accepts.
func TestDocument_RefusalsTheMountMakesItselfValidate(t *testing.T) {
	d := loadDocument(t, true)
	errSchema := d.doc.Components.Schemas["ErrorBody"].Value
	f := newFixture(t, grant(ann, ops.PrivilegeAdmin))
	for name, rec := range map[string]*httptest.ResponseRecorder{
		"a GET":                  f.send(http.MethodGet, mountPrefix+"/api/v1/list_members", ``),
		"a query string":         f.send(http.MethodPost, mountPrefix+"/api/v1/list_members?x=1", `{}`),
		"a body that is no JSON": f.send(http.MethodPost, mountPrefix+"/api/v1/list_members", `not json`),
		"an unknown operation":   f.send(http.MethodPost, mountPrefix+"/api/v1/no_such_operation", `{}`),
		"a path under api/v2":    f.send(http.MethodPost, mountPrefix+"/api/v2/list_members", `{}`),
		"api alone":              f.send(http.MethodPost, mountPrefix+"/api", `{}`),
	} {
		if rec.Code < 400 {
			t.Errorf("%s answered %d, want a refusal", name, rec.Code)
			continue
		}
		var decoded any
		if err := json.Unmarshal(rec.Body.Bytes(), &decoded); err != nil {
			t.Errorf("%s: body %q is not JSON: %v", name, rec.Body, err)
			continue
		}
		if err := errSchema.VisitJSON(decoded, openapi3.EnableJSONSchema2020()); err != nil {
			t.Errorf("%s: body %s fails the error schema: %v", name, rec.Body, err)
		}
	}
}

// TestDocument_MalformedRequestsFailTheDocument proves the request check
// can fail: a request the schema forbids is refused.
func TestDocument_MalformedRequestsFailTheDocument(t *testing.T) {
	d := loadDocument(t, false)
	for _, tt := range []struct {
		op   ops.Op
		body string
	}{
		{ops.OpGetSubmission, `{"submission_id":"XYZ"}`},
		{ops.OpGetSubmission, `{}`},
		{ops.OpAddMember, `{"address":"bo@example.com","display_name":"Bo"}`},
		{ops.OpRemoveMember, `{"user_id":"0123456789ABCDEF0123456789ABCDEF"}`},
		{ops.OpSetSetting, `{"key":"K"}`},
		{ops.OpSetSetting, `{"key":7,"value":"x"}`},
	} {
		if err := d.checkRequest(t.Context(), tt.op, tt.body); err == nil {
			t.Errorf("%s accepted the malformed request %s", tt.op, tt.body)
		}
	}
}

// TestDocument_TheChecksCanFail runs the three proofs AC2 names: each
// breaks the closed document one way and shows the exchange check then
// fails, so a passing exchange means something.
func TestDocument_TheChecksCanFail(t *testing.T) {
	t.Run("a renamed required response field", func(t *testing.T) {
		d := loadDocument(t, true)
		resp := d.doc.Components.Schemas["AddMemberResponse"].Value
		resp.Required = []string{"person"}
		resp.Properties["person"] = resp.Properties["member"]
		delete(resp.Properties, "member")
		rec := exchange(t, ops.OpAddMember, requestSamples[ops.OpAddMember], nil)
		if err := d.checkResponse(t.Context(), ops.OpAddMember, rec); err == nil {
			t.Error("add_member's answer validated against a document that renames member")
		}
	})
	t.Run("a Go response field the document omits", func(t *testing.T) {
		d := loadDocument(t, true)
		errBody := d.doc.Components.Schemas["ErrorBody"].Value.Properties["error"].Value
		delete(errBody.Properties, "reason")
		rec := exchange(t, ops.OpAddMember, requestSamples[ops.OpAddMember], ops.NewError(ops.CodeConflict, "the address is already a member"))
		if rec.Code != http.StatusConflict {
			t.Fatalf("answer = %d, want 409", rec.Code)
		}
		if err := d.checkResponse(t.Context(), ops.OpAddMember, rec); err == nil {
			t.Error("an error body with a reason validated against a closed document that omits it")
		}
	})
	t.Run("an undocumented status", func(t *testing.T) {
		d := loadDocument(t, true)
		d.operation(t, ops.OpAddMember).Responses.Delete("409")
		rec := exchange(t, ops.OpAddMember, requestSamples[ops.OpAddMember], ops.NewError(ops.CodeConflict, "the address is already a member"))
		if err := d.checkResponse(t.Context(), ops.OpAddMember, rec); err == nil {
			t.Error("a 409 validated against an operation that declares no 409")
		}
	})
	t.Run("an open document accepts the omitted field", func(t *testing.T) {
		// The published document leaves responses open; only the loaded
		// copy in these tests is closed. This pins that difference.
		d := loadDocument(t, false)
		errBody := d.doc.Components.Schemas["ErrorBody"].Value.Properties["error"].Value
		delete(errBody.Properties, "reason")
		rec := exchange(t, ops.OpAddMember, requestSamples[ops.OpAddMember], ops.NewError(ops.CodeConflict, "the address is already a member"))
		if err := d.checkResponse(t.Context(), ops.OpAddMember, rec); err != nil {
			t.Errorf("the open document refused an unknown field: %v", err)
		}
	})
	t.Run("a silent pass is a failure", func(t *testing.T) {
		d := loadDocument(t, true)
		empty := httptest.NewRecorder()
		empty.WriteHeader(http.StatusOK)
		if err := d.checkResponse(t.Context(), ops.OpListMembers, empty); err == nil {
			t.Error("an empty answer passed the response check")
		}
		redirect := httptest.NewRecorder()
		redirect.WriteHeader(http.StatusTemporaryRedirect)
		_, _ = io.Copy(redirect, bytes.NewReader([]byte(`{}`)))
		if err := d.checkResponse(t.Context(), ops.OpListMembers, redirect); err == nil {
			t.Error("a redirect passed the response check")
		}
	})
}
