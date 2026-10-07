package conformance

import (
	"context"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"
)

func newTestChecker(t *testing.T) *Checker {
	t.Helper()
	doc, err := os.ReadFile("testdata/contract.json")
	if err != nil {
		t.Fatal(err)
	}
	c, err := New(doc)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func pointers(r Result) []string {
	var out []string
	for _, v := range r.Violations {
		out = append(out, v.Pointer)
	}
	return out
}

func TestCheck(t *testing.T) {
	const jsonCT = "application/json"
	tests := []struct {
		name        string
		method      string
		path        string
		status      int
		contentType string
		body        string
		want        Outcome
		wantPtrs    []string // exact, sorted; checked only for Violations
	}{
		{
			name: "valid success body", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body: `{"subsonic-response":{"status":"ok","version":"1","items":[{"id":"a","name":"x"}]}}`,
			want: Valid,
		},
		{
			name: "valid failure body", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body: `{"subsonic-response":{"status":"failed","version":"1","error":{"code":40}}}`,
			want: Valid,
		},
		{
			name: "content type with parameters", method: http.MethodGet, path: "/rest/items", status: 200,
			contentType: "application/json; charset=utf-8",
			body:        `{"subsonic-response":{"status":"ok","version":"1","items":[]}}`,
			want:        Valid,
		},
		{
			name: "required array removed", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body:     `{"subsonic-response":{"status":"ok","version":"1"}}`,
			want:     Invalid,
			wantPtrs: []string{"/subsonic-response/items"},
		},
		{
			name: "numeric id", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body:     `{"subsonic-response":{"status":"ok","version":"1","items":[{"id":5}]}}`,
			want:     Invalid,
			wantPtrs: []string{"/subsonic-response/items/0/id"},
		},
		{
			name: "two sibling scalars numeric", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body:     `{"subsonic-response":{"status":"ok","version":"1","items":[{"id":5,"name":6}]}}`,
			want:     Invalid,
			wantPtrs: []string{"/subsonic-response/items/0/id", "/subsonic-response/items/0/name"},
		},
		{
			name: "failure branch only when the body says failed", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body:     `{"subsonic-response":{"status":"failed","version":"1","error":{"message":"m"}}}`,
			want:     Invalid,
			wantPtrs: []string{"/subsonic-response/error/code"},
		},
		{
			name: "no readable status reports every branch", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT,
			body: `{"subsonic-response":{"version":"1"}}`,
			want: Invalid,
			// "status" is missing from both branches and is reported once.
			wantPtrs: []string{"/subsonic-response/error", "/subsonic-response/items", "/subsonic-response/status"},
		},
		{name: "endpoint absent", method: http.MethodGet, path: "/rest/nothere", status: 200, contentType: jsonCT, body: `{}`, want: NoRoute},
		{name: "method absent", method: http.MethodPost, path: "/rest/items", status: 200, contentType: jsonCT, body: `{}`, want: NoRoute},
		{name: "undeclared status", method: http.MethodGet, path: "/rest/items", status: 500, contentType: jsonCT, body: `{}`, want: UndeclaredStatus},
		{name: "empty content entry", method: http.MethodGet, path: "/rest/emptycontent", status: 200, contentType: jsonCT, body: `{}`, want: NoSchema},
		{name: "media type without schema", method: http.MethodGet, path: "/rest/noschema", status: 200, contentType: jsonCT, body: `{}`, want: NoSchema},
		{name: "301 is never validated", method: http.MethodGet, path: "/rest/redirect", status: 301, contentType: jsonCT, body: `{}`, want: NoSchema},
		{name: "304 is never validated", method: http.MethodGet, path: "/rest/items", status: 304, contentType: jsonCT, body: ``, want: NoSchema},
		{name: "HEAD is never validated", method: http.MethodHead, path: "/rest/items", status: 200, contentType: jsonCT, body: `{}`, want: NoSchema},
		{
			name: "empty body", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT, body: ``,
			want: Invalid, wantPtrs: []string{""},
		},
		{
			name: "body not JSON", method: http.MethodGet, path: "/rest/items", status: 200, contentType: jsonCT, body: `<subsonic-response/>`,
			want: Invalid, wantPtrs: []string{""},
		},
		{
			name: "wrong content type", method: http.MethodGet, path: "/rest/items", status: 200, contentType: "text/xml",
			body: `{"subsonic-response":{"status":"ok","version":"1","items":[]}}`,
			want: Invalid, wantPtrs: []string{""},
		},
	}
	c := newTestChecker(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Check(context.Background(), tt.method, tt.path, tt.status, tt.contentType, []byte(tt.body))
			if got.Outcome != tt.want {
				t.Fatalf("outcome = %q, want %q (violations %+v)", got.Outcome, tt.want, got.Violations)
			}
			if tt.want != Invalid {
				if len(got.Violations) != 0 {
					t.Errorf("violations %+v on outcome %q", got.Violations, got.Outcome)
				}
				return
			}
			if p := pointers(got); !slices.Equal(p, tt.wantPtrs) {
				t.Errorf("pointers = %q, want %q (violations %+v)", p, tt.wantPtrs, got.Violations)
			}
		})
	}
}

func TestCheckViolationReasons(t *testing.T) {
	c := newTestChecker(t)
	tests := []struct {
		name        string
		contentType string
		body        string
		wantReason  string
	}{
		{"empty body", "application/json", ``, "empty body"},
		{"body not JSON", "application/json", `<a/>`, "not JSON"},
		{"wrong content type", "text/xml", `{}`, "text/xml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.Check(context.Background(), http.MethodGet, "/rest/items", 200, tt.contentType, []byte(tt.body))
			if len(got.Violations) != 1 || !strings.Contains(got.Violations[0].Reason, tt.wantReason) {
				t.Errorf("violations = %+v, want one naming %q", got.Violations, tt.wantReason)
			}
		})
	}
}

func TestNewRejectsBadDocuments(t *testing.T) {
	tests := []struct {
		name string
		doc  string
	}{
		{"empty", ``},
		{"not a document", `[1,2]`},
		{"missing paths", `{"openapi":"3.0.0","info":{"title":"t","version":"1"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := New([]byte(tt.doc)); err == nil {
				t.Error("New accepted the document")
			}
		})
	}
}
