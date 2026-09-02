package beets

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestParseSkips(t *testing.T) {
	tests := []struct {
		name string
		data string
		want []string
	}{
		{
			name: "no skip lines",
			data: "import log with no skips\nsomething else\n",
			want: nil,
		},
		{
			name: "single skip",
			data: "skip /staging/alice/Weird Album\n",
			want: []string{"/staging/alice/Weird Album"},
		},
		{
			name: "semicolon-separated paths on one skip line",
			data: "skip /staging/alice/Album A; /staging/alice/Album B\n",
			want: []string{"/staging/alice/Album A", "/staging/alice/Album B"},
		},
		{
			name: "multiple skip lines",
			data: "skip /staging/alice/Album A\nskip /staging/bob/Album C\n",
			want: []string{"/staging/alice/Album A", "/staging/bob/Album C"},
		},
		{
			name: "no trailing newline on final line",
			data: "skip /staging/alice/Album A",
			want: []string{"/staging/alice/Album A"},
		},
		{
			name: "empty part between semicolons is dropped",
			data: "skip /staging/alice/Album A; ; /staging/alice/Album B\n",
			want: []string{"/staging/alice/Album A", "/staging/alice/Album B"},
		},
		{
			name: "empty log",
			data: "",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSkips([]byte(tt.data))
			if !equalSlices(got, tt.want) {
				t.Errorf("parseSkips(%q) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// writeStubBeet writes an executable shell script at dir/beet and returns
// its path, the BeetPath test seam.
func writeStubBeet(t *testing.T, dir, script string) string {
	t.Helper()
	path := filepath.Join(dir, "beet")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write stub beet: %v", err)
	}
	return path
}

func TestRunnerImport(t *testing.T) {
	tests := []struct {
		name       string
		script     string
		configPath string
		wantSkip   []string
		wantErr    bool
	}{
		{
			name:     "clean import: no skip lines",
			script:   `: # nothing to log`,
			wantSkip: nil,
		},
		{
			name: "single skip",
			script: `
logfile=""
prev=""
for arg in "$@"; do
  if [ "$prev" = "-l" ]; then logfile="$arg"; fi
  prev="$arg"
done
echo "skip /staging/alice/Weird Album" >> "$logfile"
`,
			wantSkip: []string{"/staging/alice/Weird Album"},
		},
		{
			name: "beet exits nonzero",
			script: `
echo "beet: something went wrong" >&2
exit 1
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			stateDir := t.TempDir()
			beetPath := writeStubBeet(t, binDir, tt.script)

			r := &Runner{BeetPath: beetPath, ConfigPath: tt.configPath, StateDir: stateDir}
			result, err := r.Import(context.Background(), "/staging", "")

			if tt.wantErr {
				if err == nil {
					t.Fatalf("Import() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("Import() unexpected error: %v", err)
			}
			if !equalSlices(result.Skipped, tt.wantSkip) {
				t.Errorf("Import() Skipped = %v, want %v", result.Skipped, tt.wantSkip)
			}
			if result.LogPath == "" {
				t.Fatal("Import() LogPath is empty")
			}
			if filepath.Dir(result.LogPath) != stateDir {
				t.Errorf("Import() LogPath = %s, want it under %s", result.LogPath, stateDir)
			}
			if _, err := os.Stat(result.LogPath); err != nil {
				t.Errorf("Import() LogPath %s not persisted: %v", result.LogPath, err)
			}
		})
	}
}

// TestRunnerImportArgs pins the exact beet invocation: -c only when
// ConfigPath is set, then "import -q -l <tag log>", then --set
// collection=<c> only when collection is non-empty, then <dir> last.
func TestRunnerImportArgs(t *testing.T) {
	tests := []struct {
		name       string
		configPath string
		collection string
		wantPrefix []string
		wantSuffix []string
	}{
		{
			name:       "with config path, default collection",
			configPath: "/etc/beets/config.yaml",
			collection: "",
			wantPrefix: []string{"-c", "/etc/beets/config.yaml", "import", "-q", "-l"},
			wantSuffix: []string{"/staging"},
		},
		{
			name:       "without config path, default collection",
			configPath: "",
			collection: "",
			wantPrefix: []string{"import", "-q", "-l"},
			wantSuffix: []string{"/staging"},
		},
		{
			name:       "non-default collection sets the flexible attribute",
			configPath: "",
			collection: "archive",
			wantPrefix: []string{"import", "-q", "-l"},
			wantSuffix: []string{"--set", "collection=archive", "/staging"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			stateDir := t.TempDir()
			argsLog := filepath.Join(binDir, "args.log")
			script := fmt.Sprintf("for a in \"$@\"; do printf '%%s\\n' \"$a\"; done > \"%s\"\n", argsLog)
			beetPath := writeStubBeet(t, binDir, script)

			r := &Runner{BeetPath: beetPath, ConfigPath: tt.configPath, StateDir: stateDir}
			if _, err := r.Import(context.Background(), "/staging", tt.collection); err != nil {
				t.Fatalf("Import() unexpected error: %v", err)
			}

			raw, err := os.ReadFile(argsLog)
			if err != nil {
				t.Fatalf("read args log: %v", err)
			}
			got := strings.Split(strings.TrimRight(string(raw), "\n"), "\n")

			if len(got) < len(tt.wantPrefix)+1+len(tt.wantSuffix) {
				t.Fatalf("beet args = %v, too short to hold prefix %v plus <log> plus suffix %v", got, tt.wantPrefix, tt.wantSuffix)
			}
			for i, want := range tt.wantPrefix {
				if got[i] != want {
					t.Errorf("beet args[%d] = %q, want %q (full args: %v)", i, got[i], want, got)
				}
			}
			gotSuffix := got[len(got)-len(tt.wantSuffix):]
			if !equalSlices(gotSuffix, tt.wantSuffix) {
				t.Errorf("beet args suffix = %v, want %v (full args: %v)", gotSuffix, tt.wantSuffix, got)
			}
		})
	}
}

func TestTriggerScan(t *testing.T) {
	t.Run("builds startScan URL with u/t/s/v/c and never exposes the password", func(t *testing.T) {
		var captured *url.URL
		var method string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			captured = r.URL
			method = r.Method
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		const user, password = "admin", "s3cret"
		if err := TriggerScan(context.Background(), server.URL, user, password); err != nil {
			t.Fatalf("TriggerScan() unexpected error: %v", err)
		}

		if method != http.MethodGet {
			t.Errorf("request method = %s, want GET", method)
		}
		if captured == nil {
			t.Fatal("server never received a request")
		}
		if got := captured.Path; got != "/rest/startScan" {
			t.Errorf("request path = %s, want /rest/startScan", got)
		}

		q := captured.Query()
		if got := q.Get("u"); got != user {
			t.Errorf("u = %q, want %q", got, user)
		}
		if got := q.Get("v"); got != subsonicVersion {
			t.Errorf("v = %q, want %q", got, subsonicVersion)
		}
		if got := q.Get("c"); got != subsonicClient {
			t.Errorf("c = %q, want %q", got, subsonicClient)
		}

		salt := q.Get("s")
		if !regexp.MustCompile(`^[0-9a-f]{16}$`).MatchString(salt) {
			t.Errorf("s = %q, want 16 lowercase hex chars", salt)
		}
		token := q.Get("t")
		if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(token) {
			t.Errorf("t = %q, want 32 lowercase hex chars", token)
		}
		sum := md5.Sum([]byte(password + salt))
		if want := hex.EncodeToString(sum[:]); token != want {
			t.Errorf("t = %q, want md5(password+salt) = %q", token, want)
		}

		if strings.Contains(captured.String(), password) {
			t.Errorf("request URL %q contains the raw password", captured.String())
		}
	})

	t.Run("fresh salt on every call", func(t *testing.T) {
		var salts []string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			salts = append(salts, r.URL.Query().Get("s"))
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		for range 2 {
			if err := TriggerScan(context.Background(), server.URL, "admin", "s3cret"); err != nil {
				t.Fatalf("TriggerScan() unexpected error: %v", err)
			}
		}

		if len(salts) != 2 {
			t.Fatalf("server received %d requests, want 2", len(salts))
		}
		if salts[0] == salts[1] {
			t.Errorf("salt was reused across calls: both %q", salts[0])
		}
	})

	t.Run("a non-2xx response is returned as an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		err := TriggerScan(context.Background(), server.URL, "admin", "s3cret")
		if err == nil {
			t.Fatal("TriggerScan() error = nil, want an error for a 500 response")
		}
	})

	t.Run("an unreachable server is returned as an error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		unreachable := server.URL
		server.Close() // closed before use: nothing is listening on this port anymore

		err := TriggerScan(context.Background(), unreachable, "admin", "s3cret")
		if err == nil {
			t.Fatal("TriggerScan() error = nil, want an error for an unreachable server")
		}
	})
}

// TestResolveLibraryPaths runs against a real sqlite3 binary and a
// minimal items table shaped like beets' own schema, proving the query
// itself -- not just its parsing -- against the tool it actually shells
// out to.
func TestResolveLibraryPaths(t *testing.T) {
	sqlitePath, err := exec.LookPath("sqlite3")
	if err != nil {
		t.Fatalf("sqlite3 not on PATH: it is a required test dependency, install it")
	}

	dbPath := filepath.Join(t.TempDir(), "library.db")
	initSQL := `
CREATE TABLE items (id INTEGER PRIMARY KEY, album_id INTEGER, path TEXT, added REAL);
INSERT INTO items VALUES (1, 10, '/library/Artist A/Album A/01.flac', 100.0);
INSERT INTO items VALUES (2, 10, '/library/Artist A/Album A/02.flac', 100.5);
INSERT INTO items VALUES (3, 11, '/library/Artist B/Album B/01.flac', 200.0);
INSERT INTO items VALUES (4, 0, '/library/Singles/Artist C - Track.flac', 300.0);
INSERT INTO items VALUES (5, 99, '/library/Old/before-the-window.flac', 1.0);
`
	if out, err := exec.Command(sqlitePath, dbPath, initSQL).CombinedOutput(); err != nil {
		t.Fatalf("seed database: %v: %s", err, out)
	}

	groups, err := ResolveLibraryPaths(context.Background(), sqlitePath, dbPath, 50, 300)
	if err != nil {
		t.Fatalf("ResolveLibraryPaths() unexpected error: %v", err)
	}

	want := []AlbumPaths{
		{AlbumID: 0, Paths: []string{"/library/Singles/Artist C - Track.flac"}},
		{AlbumID: 10, Paths: []string{"/library/Artist A/Album A/01.flac", "/library/Artist A/Album A/02.flac"}},
		{AlbumID: 11, Paths: []string{"/library/Artist B/Album B/01.flac"}},
	}
	if len(groups) != len(want) {
		t.Fatalf("ResolveLibraryPaths() = %+v, want %d groups", groups, len(want))
	}
	for i, g := range groups {
		if g.AlbumID != want[i].AlbumID || !equalSlices(g.Paths, want[i].Paths) {
			t.Errorf("group[%d] = %+v, want %+v", i, g, want[i])
		}
	}
}
