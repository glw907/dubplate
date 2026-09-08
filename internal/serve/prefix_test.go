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

// callRecorder is one of three distinct recorders TestRoute uses per
// row, so a row cannot pass by reaching the wrong terminal with the
// right status.
type callRecorder struct {
	called bool
}

func recordingHandlerWithStatus(status int, dst *callRecorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst.called = true
		w.WriteHeader(status)
	})
}

// recordingTerminator wraps the package's real [Terminator] rather
// than faking its status, so a row using it still exercises the
// production 404 and not a stand-in.
func recordingTerminator(dst *callRecorder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dst.called = true
		Terminator().ServeHTTP(w, r)
	})
}

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

// terminal names which of the three recorders a row expects to have
// been called.
type terminal int

const (
	terminalRegistered terminal = iota
	terminalTerminator
	terminalFallthrough
)

func TestRoute(t *testing.T) {
	tests := []struct {
		name string
		path string
		// useTerminator selects which handler backs the row's "owned"
		// argument when the path is owned; ignored otherwise.
		useTerminator bool
		want          terminal
		wantStatus    int
	}{
		{
			name:          "the bare prefix enters and, with nothing registered, reaches the terminator",
			path:          testPrefix,
			useTerminator: true,
			want:          terminalTerminator,
			wantStatus:    http.StatusNotFound,
		},
		{
			name:          "the prefix with a trailing separator enters and, with nothing registered, reaches the terminator",
			path:          testPrefix + "/",
			useTerminator: true,
			want:          terminalTerminator,
			wantStatus:    http.StatusNotFound,
		},
		{
			name:          "a path under the prefix enters and reaches a registered route",
			path:          testPrefix + "/nested",
			useTerminator: false,
			want:          terminalRegistered,
			wantStatus:    http.StatusOK,
		},
		{
			name:          "a different case in a later segment still enters, since only the prefix itself is case-checked",
			path:          testPrefix + "/PATH",
			useTerminator: false,
			want:          terminalRegistered,
			wantStatus:    http.StatusOK,
		},
		{
			name:       "a neighbor sharing the prefix's bytes without a segment boundary does not enter (the mountBasePath defect)",
			path:       testPrefix + "foo",
			want:       terminalFallthrough,
			wantStatus: http.StatusOK,
		},
		{
			name:       "a differently cased prefix does not enter",
			path:       "/Anchor",
			want:       terminalFallthrough,
			wantStatus: http.StatusOK,
		},
		{
			name:       "the root path does not enter",
			path:       "/",
			want:       terminalFallthrough,
			wantStatus: http.StatusOK,
		},
		{
			name:       "a path that does not begin with the prefix does not enter",
			path:       "/other",
			want:       terminalFallthrough,
			wantStatus: http.StatusOK,
		},
	}

	prefix, err := NewPrefix(testPrefix)
	if err != nil {
		t.Fatalf("NewPrefix(%q): %v", testPrefix, err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var registered, terminated, fellThrough callRecorder

			var owned http.Handler
			if tt.useTerminator {
				owned = recordingTerminator(&terminated)
			} else {
				owned = recordingHandlerWithStatus(http.StatusOK, &registered)
			}
			next := recordingHandlerWithStatus(http.StatusOK, &fellThrough)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()
			prefix.Route(owned, next).ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			got := map[terminal]bool{
				terminalRegistered:  registered.called,
				terminalTerminator:  terminated.called,
				terminalFallthrough: fellThrough.called,
			}
			for term, called := range got {
				want := term == tt.want
				if called != want {
					t.Fatalf("terminal %v called = %v, want %v (registered=%v terminator=%v fallthrough=%v)",
						term, called, want, registered.called, terminated.called, fellThrough.called)
				}
			}
		})
	}
}
