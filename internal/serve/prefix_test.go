package serve

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// testPrefix is a literal this test invents and never the configured
// default ("/dubplate", `spec:1566` via R-A), so the table would fail
// under R-A's other answer too.
const testPrefix = "/anchor"

func TestNewPrefix(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"a valid prefix", testPrefix, false},
		{"missing the leading separator", "anchor", true},
		{"a trailing separator", testPrefix + "/", true},
		{"the bare root path", "/", true},
		{"a doubled separator", "//anchor", true},
		{"a dot segment", testPrefix + "/.", true},
		{"a dot-dot segment", testPrefix + "/..", true},
		{"a backslash", testPrefix + `\x`, true},
		{"a percent-encoded byte", "/%2e%2e", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewPrefix(tt.value)
			if tt.wantErr {
				if !errors.Is(err, ErrInvalidPrefix) {
					t.Fatalf("NewPrefix(%q) error = %v, want errors.Is(err, ErrInvalidPrefix)", tt.value, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewPrefix(%q) unexpected error: %v", tt.value, err)
			}
			if got := p.String(); got != tt.value {
				t.Fatalf("p.String() = %q, want %q", got, tt.value)
			}
		})
	}
}

// TestNewPrefixReportsEachShapeByItsOwnReason pins the message text per
// refused shape, not only the sentinel. url.URL's default path escaping
// escapes a backslash, a space, a tab and a "!" alike, so a validator
// that runs the escaping check first answers every one of them with
// "must carry no percent-encoded byte" and sends an adopter hunting an
// encoding their value never carried. The want strings are literals
// this test owns: deriving them from the validator would make every row
// pass by construction.
func TestNewPrefixReportsEachShapeByItsOwnReason(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"a backslash", `/an\chor`, "must carry no backslash"},
		{"a space", "/an chor", "must carry no whitespace"},
		{"a tab", "/an\tchor", "must carry no whitespace"},
		{"a percent sign", "/an%20chor", "must carry no percent-encoded byte"},
		{"an exclamation mark", "/an!chor", "must carry only bytes a request path admits unescaped"},
		{"a doubled separator", "//anchor", "must carry no doubled separator"},
		{"a dot-dot segment", "/anchor/..", `must carry no "." or ".." segment`},
		{"no leading separator", "anchor", "must begin with a separator"},
		{"a trailing separator", "/anchor/", "must carry no trailing separator"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPrefix(tt.value)
			if !errors.Is(err, ErrInvalidPrefix) {
				t.Fatalf("NewPrefix(%q) error = %v, want errors.Is(err, ErrInvalidPrefix)", tt.value, err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("NewPrefix(%q) error = %q, want it to carry %q", tt.value, err, tt.want)
			}
		})
	}
}

// TestHandlerOwnsOnlySegmentBoundaries registers one subtree and sends
// paths that share its bytes without a segment boundary, so a matcher
// that trimmed the prefix as a plain string would reach the subtree
// for a neighbor such as testPrefix+"echo" instead of answering 404.
func TestHandlerOwnsOnlySegmentBoundaries(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		wantStatus  int
		wantReached bool
	}{
		{"the subtree itself", testPrefix + "/echo", http.StatusOK, true},
		{"a path under the subtree", testPrefix + "/echo/nested", http.StatusOK, true},
		{"a different case in a later segment", testPrefix + "/echo/PATH", http.StatusOK, true},
		{"the bare prefix", testPrefix, http.StatusNotFound, false},
		{"the prefix with a trailing separator", testPrefix + "/", http.StatusNotFound, false},
		{"a neighbor sharing the prefix's bytes", testPrefix + "echo", http.StatusNotFound, false},
		{"a neighbor sharing the subtree's bytes", testPrefix + "/echoes", http.StatusNotFound, false},
		{"a differently cased prefix", "/Anchor/echo", http.StatusNotFound, false},
		{"the root path", "/", http.StatusNotFound, false},
		{"a path outside the prefix", "/other", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shell := newTestShell(t)
			var reached recorder
			shell.Register("echo", recordingHandler(&reached))

			rec := httptest.NewRecorder()
			shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if rec.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if got := reached.header != nil; got != tt.wantReached {
				t.Errorf("subtree reached = %v, want %v", got, tt.wantReached)
			}
		})
	}
}
