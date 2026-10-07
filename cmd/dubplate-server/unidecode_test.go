package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
)

// loadCachedUnidecodeWheel reads the workstation's cached copy of the
// pinned wheel, filled by `make unidecode-wheel`. It fails rather than
// skips when the cache is empty, since Go's test cache does not track
// the wheel file and a stale "ok" would hide the gap.
func loadCachedUnidecodeWheel(t *testing.T) (path string, wheel []byte) {
	t.Helper()
	path, err := unidecode.WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}
	wheel, err = os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached wheel at %s: %v; run make unidecode-wheel", path, err)
	}
	return path, wheel
}

// installTestUnidecodeTable installs a valid table into stateDir by
// running the install verb's own logic against --from the
// workstation's cached wheel, never by copying bytes: the leaf a
// doctor test needs already installed.
func installTestUnidecodeTable(t *testing.T, stateDir string) {
	t.Helper()
	path, _ := loadCachedUnidecodeWheel(t)
	denyUnidecodeNetwork(t)
	var stdout bytes.Buffer
	environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}
	if err := runUnidecodeInstall(context.Background(), environ, path, &stdout); err != nil {
		t.Fatalf("install test unidecode table into %s: %v", stateDir, err)
	}
}

// unidecodeRewritingTransport asserts that every request targets
// unidecode.WheelURL, counts the requests it sees, and forwards a
// clone to a test server standing in for PyPI. Any other host fails
// the test instead of reaching the network: every download in a test
// goes through a rewriting transport.
type unidecodeRewritingTransport struct {
	t        *testing.T
	server   *httptest.Server
	requests int
}

func (rt *unidecodeRewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.t.Helper()
	if req.URL.String() != unidecode.WheelURL {
		rt.t.Fatalf("request to %s, want %s", req.URL, unidecode.WheelURL)
	}
	rt.requests++

	target, err := url.Parse(rt.server.URL)
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme = target.Scheme
	clone.URL.Host = target.Host
	clone.Host = target.Host
	return http.DefaultTransport.RoundTrip(clone)
}

// useUnidecodeServer points unidecodeClient at server through a
// rewriting transport for the test, and returns the transport so a test
// can read its request count.
func useUnidecodeServer(t *testing.T, server *httptest.Server) *unidecodeRewritingTransport {
	t.Helper()
	rt := &unidecodeRewritingTransport{t: t, server: server}
	orig := unidecodeClient
	unidecodeClient = &http.Client{Transport: rt}
	t.Cleanup(func() { unidecodeClient = orig })
	return rt
}

// denyUnidecodeNetwork points unidecodeClient at a server whose handler
// fails the test on any request, so a call that should reach the network
// never fails the test instead of opening a real connection.
func denyUnidecodeNetwork(t *testing.T) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected request: this call should reach the network never")
	}))
	t.Cleanup(server.Close)
	useUnidecodeServer(t, server)
}

func TestUnidecodeInstall_RefusesWithoutStateDir(t *testing.T) {
	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), nil, "", &stdout)
	if err == nil || !strings.Contains(err.Error(), "DUBPLATE_STATE_DIR") {
		t.Fatalf("error = %v, want a refusal naming DUBPLATE_STATE_DIR", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

func TestUnidecodeInstall_RefusesAbsentStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "absent")
	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil || !strings.Contains(err.Error(), stateDir) {
		t.Fatalf("error = %v, want a refusal naming %s", err, stateDir)
	}
}

// TestUnidecodeInstall_RefusesForeignStateDirOwner proves the state
// directory's own ownership refusal: a dropped owner check would let
// this pass instead.
func TestUnidecodeInstall_RefusesForeignStateDirOwner(t *testing.T) {
	orig := geteuid
	geteuid = func() int { return os.Geteuid() + 1 }
	t.Cleanup(func() { geteuid = orig })

	stateDir := t.TempDir()
	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil || !strings.Contains(err.Error(), "with the server's environment file loaded") {
		t.Fatalf("error = %v, want a refusal naming the environment-file fix", err)
	}
}

// TestUnidecodeInstall_OwnershipChecksPrecedeAlreadyInstalled installs
// a valid table as the real owner, then makes every subsequent
// ownership check see a foreign uid: the refusal must still fire, so
// the ownership checks run before the already-installed check reads
// the table at all.
func TestUnidecodeInstall_OwnershipChecksPrecedeAlreadyInstalled(t *testing.T) {
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)

	orig := geteuid
	geteuid = func() int { return os.Geteuid() + 1 }
	t.Cleanup(func() { geteuid = orig })

	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil {
		t.Fatal("error = nil, want a refusal: ownership checks must run before the already-installed check")
	}
	if strings.Contains(stdout.String(), "already installed") {
		t.Errorf("stdout = %q, want no already-installed message", stdout.String())
	}
}

// TestUnidecodeInstall_RefusesForeignLeafOwner proves the leaf's own
// ownership refusal, independent of the state directory's: geteuid
// answers the state directory's own check with the real owner, since
// the test process created it, and only the leaf's check sees a
// foreign uid, the same split internal/engine/store's read path
// (Inspect) and write path (leaf.create) draw between the two checks.
func TestUnidecodeInstall_RefusesForeignLeafOwner(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	if err := os.Mkdir(leafPath, unidecodeLeafPerm); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}

	real := os.Geteuid()
	calls := 0
	orig := geteuid
	geteuid = func() int {
		calls++
		if calls == 1 {
			return real
		}
		return real + 1
	}
	t.Cleanup(func() { geteuid = orig })

	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil || !strings.Contains(err.Error(), leafPath) || !strings.Contains(err.Error(), "with the server's environment file loaded") {
		t.Fatalf("error = %v, want a refusal naming %s and the environment-file fix", err, leafPath)
	}
}

func TestUnidecodeInstall_RefusesBadLeafShape(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, leafPath string)
	}{
		{
			name: "leaf is a file",
			setup: func(t *testing.T, leafPath string) {
				t.Helper()
				if err := os.WriteFile(leafPath, []byte("not a directory"), 0o644); err != nil {
					t.Fatalf("write file at leaf path: %v", err)
				}
			},
		},
		{
			name: "leaf has the wrong mode",
			setup: func(t *testing.T, leafPath string) {
				t.Helper()
				if err := os.Mkdir(leafPath, 0o700); err != nil {
					t.Fatalf("mkdir leaf: %v", err)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			leafPath := filepath.Join(stateDir, unidecodeLeafName)
			tt.setup(t, leafPath)

			var stdout bytes.Buffer
			denyUnidecodeNetwork(t)
			err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
			if err == nil {
				t.Fatal("error = nil, want a refusal")
			}
			if !strings.Contains(err.Error(), leafPath) {
				t.Errorf("error = %v, want it to name %s", err, leafPath)
			}
		})
	}
}

// TestUnidecodeInstall_AlreadyInstalledMakesNoRequest proves the
// idempotent path: an idempotent path that fetches anyway fails this
// on the request count.
func TestUnidecodeInstall_AlreadyInstalledMakesNoRequest(t *testing.T) {
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected request when a valid table is already installed")
	}))
	defer server.Close()
	rt := useUnidecodeServer(t, server)

	var stdout bytes.Buffer
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if rt.requests != 0 {
		t.Errorf("requests = %d, want 0", rt.requests)
	}
	if !strings.Contains(stdout.String(), "already installed") {
		t.Errorf("stdout = %q, want it to mention already installed", stdout.String())
	}
}

// TestUnidecodeInstall_ReplacesInvalidTable proves an invalid table on
// disk is rebuilt and replaced, and the old bytes never survive a
// success.
func TestUnidecodeInstall_ReplacesInvalidTable(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	if err := os.MkdirAll(leafPath, unidecodeLeafPerm); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}
	tablePath := filepath.Join(leafPath, unidecode.TableFileName)
	oldBytes := []byte("not a valid table")
	if err := os.WriteFile(tablePath, oldBytes, unidecodeTablePerm); err != nil {
		t.Fatalf("write invalid table: %v", err)
	}

	wheelPath, _ := loadCachedUnidecodeWheel(t)
	denyUnidecodeNetwork(t)

	var stdout bytes.Buffer
	if err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, wheelPath, &stdout); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	got, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if bytes.Equal(got, oldBytes) {
		t.Fatal("table bytes unchanged, want the invalid table replaced")
	}
	if _, err := unidecode.Parse(got); err != nil {
		t.Fatalf("Parse(written table): %v", err)
	}
}

func TestUnidecodeInstall_RefusesBadFromPath(t *testing.T) {
	tests := []struct {
		name string
		from func(t *testing.T) string
	}{
		{"absent", func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing.whl") }},
		{"a directory", func(t *testing.T) string { return t.TempDir() }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			from := tt.from(t)

			var stdout bytes.Buffer
			denyUnidecodeNetwork(t)
			err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, from, &stdout)
			if err == nil || !strings.Contains(err.Error(), from) {
				t.Fatalf("error = %v, want a refusal naming %s", err, from)
			}
			if _, statErr := os.Stat(filepath.Join(stateDir, unidecodeLeafName)); !errors.Is(statErr, fs.ErrNotExist) {
				t.Errorf("leaf exists after a refused install: %v", statErr)
			}
		})
	}
}

// TestUnidecodeInstall_RefusesOversizedFromFile proves the --from cap
// refuses before the file is hashed: a cap check that hashed first
// would refuse on checksum instead, which the message check below
// would catch. The file is one byte past the 1 MiB cap.
func TestUnidecodeInstall_RefusesOversizedFromFile(t *testing.T) {
	stateDir := t.TempDir()
	fromPath := filepath.Join(t.TempDir(), "oversized.whl")
	f, err := os.Create(fromPath)
	if err != nil {
		t.Fatalf("create scratch file: %v", err)
	}
	if err := f.Truncate(1<<20 + 1); err != nil {
		t.Fatalf("truncate scratch file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close scratch file: %v", err)
	}

	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err = runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, fromPath, &stdout)
	if err == nil || !strings.Contains(err.Error(), "exceeds maximum size") || !strings.Contains(err.Error(), fromPath) {
		t.Fatalf("error = %v, want a size refusal naming %s", err, fromPath)
	}
	if _, statErr := os.Stat(filepath.Join(stateDir, unidecodeLeafName)); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("leaf exists after a refused install: %v", statErr)
	}
}

func TestUnidecodeInstall_RefusesWrongFromFile(t *testing.T) {
	stateDir := t.TempDir()
	fromPath := filepath.Join(t.TempDir(), "wrong.whl")
	if err := os.WriteFile(fromPath, []byte("not the pinned wheel"), 0o644); err != nil {
		t.Fatalf("write scratch file: %v", err)
	}

	var stdout bytes.Buffer
	denyUnidecodeNetwork(t)
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, fromPath, &stdout)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") || !strings.Contains(err.Error(), fromPath) {
		t.Fatalf("error = %v, want a checksum refusal naming %s", err, fromPath)
	}
	if !strings.Contains(err.Error(), unidecode.WheelSHA256) {
		t.Errorf("error = %q, want it to name the pinned sha256 %s", err.Error(), unidecode.WheelSHA256)
	}
}

// TestUnidecodeInstall_DownloadRefusalCarriesOfflineGuidance proves a
// failed download writes nothing and names the wheel's URL, its
// sha256, and --from.
func TestUnidecodeInstall_DownloadRefusalCarriesOfflineGuidance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()
	useUnidecodeServer(t, server)

	stateDir := t.TempDir()
	var stdout bytes.Buffer
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil {
		t.Fatal("error = nil, want a download refusal")
	}
	for _, want := range []string{unidecode.WheelURL, unidecode.WheelSHA256, "--from"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err.Error(), want)
		}
	}
	if _, statErr := os.Stat(filepath.Join(stateDir, unidecodeLeafName)); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("leaf exists after a refused download: %v", statErr)
	}
}

// TestUnidecodeInstall_FailedDownloadLeavesExistingFileUntouched
// proves a failed download, attempting to replace an invalid table,
// never removes what was already there.
func TestUnidecodeInstall_FailedDownloadLeavesExistingFileUntouched(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	if err := os.MkdirAll(leafPath, unidecodeLeafPerm); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}
	tablePath := filepath.Join(leafPath, unidecode.TableFileName)
	existing := []byte("stale but present")
	if err := os.WriteFile(tablePath, existing, unidecodeTablePerm); err != nil {
		t.Fatalf("write existing table: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	useUnidecodeServer(t, server)

	var stdout bytes.Buffer
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err == nil {
		t.Fatal("error = nil, want a download refusal")
	}
	got, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if !bytes.Equal(got, existing) {
		t.Error("table bytes changed after a failed download")
	}
}

// TestUnidecodeInstall_Success proves the success row end to end
// through the network path: the leaf's mode is set explicitly rather
// than left to the umask, the table file is 0644, the
// written table round-trips through Parse, and the output names the
// path and version.
func TestUnidecodeInstall_Success(t *testing.T) {
	_, wheel := loadCachedUnidecodeWheel(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(wheel)
	}))
	defer server.Close()
	rt := useUnidecodeServer(t, server)

	stateDir := t.TempDir()

	// This test runs serially (this package runs no test in
	// parallel): a restrictive umask proves the leaf's 0755 mode is
	// set by an explicit os.Chmod, not left to whatever os.Mkdir's own
	// call produces under the umask in effect.
	oldMask := syscall.Umask(0o027)
	defer syscall.Umask(oldMask)

	var stdout bytes.Buffer
	err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, "", &stdout)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if rt.requests != 1 {
		t.Errorf("requests = %d, want 1", rt.requests)
	}

	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	leafInfo, err := os.Stat(leafPath)
	if err != nil {
		t.Fatalf("stat leaf: %v", err)
	}
	if mode := leafInfo.Mode().Perm(); mode != unidecodeLeafPerm {
		t.Errorf("leaf mode = %04o, want %04o", mode, unidecodeLeafPerm)
	}

	tablePath := filepath.Join(leafPath, unidecode.TableFileName)
	tableInfo, err := os.Stat(tablePath)
	if err != nil {
		t.Fatalf("stat table: %v", err)
	}
	if mode := tableInfo.Mode().Perm(); mode != unidecodeTablePerm {
		t.Errorf("table mode = %04o, want %04o", mode, unidecodeTablePerm)
	}

	data, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if _, err := unidecode.Parse(data); err != nil {
		t.Errorf("Parse(installed table): %v", err)
	}

	out := stdout.String()
	if !strings.Contains(out, tablePath) || !strings.Contains(out, unidecode.Version) {
		t.Errorf("stdout = %q, want it to name the path %q and version %q", out, tablePath, unidecode.Version)
	}
}

// TestUnidecodeInstall_CreatesNothingUnderStateDirOnRefusal proves the
// state directory is still empty after each refusal that runs before
// any write.
func TestUnidecodeInstall_CreatesNothingUnderStateDirOnRefusal(t *testing.T) {
	tests := []struct {
		name string
		// setup swaps what the row needs and returns the --from path.
		setup func(t *testing.T) string
	}{
		{"foreign state directory owner", func(t *testing.T) string {
			denyUnidecodeNetwork(t)
			orig := geteuid
			geteuid = func() int { return os.Geteuid() + 1 }
			t.Cleanup(func() { geteuid = orig })
			return ""
		}},
		{"absent --from file", func(t *testing.T) string {
			denyUnidecodeNetwork(t)
			return filepath.Join(t.TempDir(), "missing.whl")
		}},
		{"--from directory", func(t *testing.T) string {
			denyUnidecodeNetwork(t)
			return t.TempDir()
		}},
		{"--from file that is not the pinned wheel", func(t *testing.T) string {
			denyUnidecodeNetwork(t)
			path := filepath.Join(t.TempDir(), "wrong.whl")
			if err := os.WriteFile(path, []byte("not the pinned wheel"), 0o644); err != nil {
				t.Fatalf("write scratch file: %v", err)
			}
			return path
		}},
		{"refused download", func(t *testing.T) string {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			t.Cleanup(server.Close)
			useUnidecodeServer(t, server)
			return ""
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			from := tt.setup(t)

			var stdout bytes.Buffer
			if err := runUnidecodeInstall(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir}, from, &stdout); err == nil {
				t.Fatal("error = nil, want a refusal")
			}

			after, err := os.ReadDir(stateDir)
			if err != nil {
				t.Fatalf("ReadDir: %v", err)
			}
			if len(after) != 0 {
				t.Fatalf("state directory listing = %v, want empty", after)
			}
		})
	}
}

// TestUnidecodeInstallCmd_NamesTheVerbOnce proves the verb's prefix
// is added at the command boundary, once, to a refusal from below it.
func TestUnidecodeInstallCmd_NamesTheVerbOnce(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := newUnidecodeInstallCmd(nil, &stdout, &stderr)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil || err.Error() != "unidecode install: DUBPLATE_STATE_DIR is not set" {
		t.Fatalf("error = %v, want %q", err, "unidecode install: DUBPLATE_STATE_DIR is not set")
	}
}

// TestDoctorCmd_UnidecodeRow pins the row's four states and the
// command's exit for each, against baseEnviron's otherwise-satisfied
// wall so the row's own contribution is what moves the exit.
func TestDoctorCmd_UnidecodeRow(t *testing.T) {
	tests := []struct {
		name         string
		setStateDir  bool
		setup        func(t *testing.T, stateDir string)
		wantContains []string
		wantFail     bool
	}{
		{
			// The row itself reports no failure, but the wall's own
			// DUBPLATE_STATE_DIR row is unset too, and that failure is
			// what moves the exit here.
			name:         "not checked when DUBPLATE_STATE_DIR is unset",
			setStateDir:  false,
			wantContains: []string{"not checked"},
			wantFail:     true,
		},
		{
			name:         "ok when a valid table is installed",
			setStateDir:  true,
			setup:        func(t *testing.T, stateDir string) { installTestUnidecodeTable(t, stateDir) },
			wantContains: []string{fmt.Sprintf("ok (%s)", unidecode.Version)},
		},
		{
			name:         "missing when no table was ever installed",
			setStateDir:  true,
			wantContains: []string{"missing: run dubplate-server unidecode install"},
			wantFail:     true,
		},
		{
			name:        "invalid when the table fails Parse",
			setStateDir: true,
			setup: func(t *testing.T, stateDir string) {
				leafPath := filepath.Join(stateDir, unidecodeLeafName)
				if err := os.MkdirAll(leafPath, unidecodeLeafPerm); err != nil {
					t.Fatalf("mkdir leaf: %v", err)
				}
				tablePath := filepath.Join(leafPath, unidecode.TableFileName)
				if err := os.WriteFile(tablePath, []byte("not a table"), unidecodeTablePerm); err != nil {
					t.Fatalf("write invalid table: %v", err)
				}
			},
			wantContains: []string{"invalid: ", "run dubplate-server unidecode install"},
			wantFail:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, stateDir)
			}
			environ := dropKey(baseEnviron(), "DUBPLATE_STATE_DIR")
			if tt.setStateDir {
				environ = append(environ, "DUBPLATE_STATE_DIR="+stateDir)
			}

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, config.NewRegistry(), &stdout, &stderr)
			err := cmd.Execute()

			if tt.wantFail && err == nil {
				t.Fatal("Execute() error = nil, want a refusal")
			}
			if !tt.wantFail && err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}

			wantPrefix := fmt.Sprintf("%-32s", unidecodeDoctorRow)
			var gotLine string
			for line := range strings.SplitSeq(stdout.String(), "\n") {
				if strings.HasPrefix(line, wantPrefix) {
					gotLine = line
				}
			}
			if gotLine == "" {
				t.Fatalf("stdout has no unidecode row: %q", stdout.String())
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(gotLine, want) {
					t.Errorf("unidecode row = %q, want it to contain %q", gotLine, want)
				}
			}
		})
	}
}

// TestDoctorCmd_UnidecodeRowCreatesNothing proves doctor's row calls
// the load half alone: an empty state directory is still empty after a
// doctor run reports its table missing.
func TestDoctorCmd_UnidecodeRowCreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	before, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir before: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd([]string{"DUBPLATE_STATE_DIR=" + stateDir}, config.NewRegistry(), &stdout, &stderr)
	_ = cmd.Execute()

	after, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir after: %v", err)
	}
	if len(before) != 0 || len(after) != 0 {
		t.Fatalf("state directory listing before = %v, after = %v, want both empty", before, after)
	}
}
