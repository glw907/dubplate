package serve

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestShellHandler_SecurityHeadersOnEveryResponse proves the header
// step reaches a response from every step of the chain: a registered
// subtree, the normalization refusal, the echo-method refusal, the
// origin refusal, the shell's 404, and a path outside the prefix. HSTS is
// present on every row with it enabled and absent on every row
// without. Referrer-Policy is the header a refusal cannot supply for
// itself: [http.Error] sets nosniff on its own, so a step moved inside
// the chain still leaves nosniff on a refusal.
func TestShellHandler_SecurityHeadersOnEveryResponse(t *testing.T) {
	requests := []struct {
		name       string
		method     string
		target     string
		wantStatus int
	}{
		{name: "registered subtree", method: http.MethodGet, target: testPrefix + "/page", wantStatus: http.StatusOK},
		{name: "normalization refusal", method: http.MethodGet, target: testPrefix + "/a/../page", wantStatus: http.StatusBadRequest},
		{name: "echo method refusal", method: http.MethodTrace, target: testPrefix + "/page", wantStatus: http.StatusMethodNotAllowed},
		{name: "origin refusal", method: http.MethodPost, target: testPrefix + "/page", wantStatus: http.StatusForbidden},
		{name: "unclaimed under the prefix", method: http.MethodGet, target: testPrefix + "/unknown", wantStatus: http.StatusNotFound},
		{name: "outside the prefix", method: http.MethodGet, target: "/elsewhere", wantStatus: http.StatusNotFound},
	}
	for _, hsts := range []bool{false, true} {
		shell := newTestShell(t)
		shell.Register("page", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("page"))
		}))
		if hsts {
			shell.EnableHSTS()
		}
		handler := shell.Handler()

		wantHSTS := ""
		if hsts {
			wantHSTS = "max-age=31536000"
		}
		for _, rr := range requests {
			t.Run(fmt.Sprintf("hsts=%v/%s", hsts, rr.name), func(t *testing.T) {
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, httptest.NewRequest(rr.method, rr.target, nil))
				if rec.Code != rr.wantStatus {
					t.Fatalf("status = %d, want %d", rec.Code, rr.wantStatus)
				}
				h := rec.Header()
				if got := h.Get("X-Content-Type-Options"); got != "nosniff" {
					t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
				}
				if got := h.Get("Referrer-Policy"); got != "no-referrer" {
					t.Errorf("Referrer-Policy = %q, want no-referrer", got)
				}
				if got := h.Get("Strict-Transport-Security"); got != wantHSTS {
					t.Errorf("hsts %v: Strict-Transport-Security = %q, want %q", hsts, got, wantHSTS)
				}
				if got := h.Get("Content-Security-Policy"); got != "" {
					t.Errorf("Content-Security-Policy = %q, want none from the shell", got)
				}
			})
		}
	}
}

func TestShell_EnableHSTSAfterHandlerPanics(t *testing.T) {
	shell := newTestShell(t)
	shell.Handler()
	defer func() {
		if recover() == nil {
			t.Fatal("EnableHSTS after Handler: want a panic")
		}
	}()
	shell.EnableHSTS()
}
