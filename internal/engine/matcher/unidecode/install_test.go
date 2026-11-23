package unidecode

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// cachedWheelPath returns the workstation's cached wheel, which the
// install tests pass as the --from file, and fails when it is absent.
func cachedWheelPath(t *testing.T) string {
	t.Helper()
	loadCachedWheel(t)
	path, err := WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}
	return path
}

// denyNetwork returns options whose client reaches a server that fails
// the test on any request, so a call that should never download fails
// instead of opening a connection.
func denyNetwork(t *testing.T) InstallOptions {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected request: this call should reach the network never")
	}))
	t.Cleanup(server.Close)
	client, _ := newTestClient(t, server)
	return InstallOptions{Client: client, Geteuid: os.Geteuid}
}

// foreign is a Geteuid that names a uid other than the process's.
func foreign() int { return os.Geteuid() + 1 }

// installTestTable installs a valid table into stateDir by running
// [Install] against the cached wheel, never by copying bytes.
func installTestTable(t *testing.T, stateDir string) {
	t.Helper()
	var stdout bytes.Buffer
	if err := Install(context.Background(), stateDir, cachedWheelPath(t), &stdout, denyNetwork(t)); err != nil {
		t.Fatalf("install test table into %s: %v", stateDir, err)
	}
}

func TestUnidecodeInstall_RefusesAbsentStateDir(t *testing.T) {
	stateDir := filepath.Join(t.TempDir(), "absent")
	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, denyNetwork(t))
	if err == nil || !strings.Contains(err.Error(), stateDir) {
		t.Fatalf("error = %v, want a refusal naming %s", err, stateDir)
	}
}

// TestUnidecodeInstall_RefusesForeignStateDirOwner proves the state
// directory's ownership refusal: a dropped owner check would let this
// pass instead.
func TestUnidecodeInstall_RefusesForeignStateDirOwner(t *testing.T) {
	opts := denyNetwork(t)
	opts.Geteuid = foreign

	var stdout bytes.Buffer
	err := Install(context.Background(), t.TempDir(), "", &stdout, opts)
	if err == nil || !strings.Contains(err.Error(), "with the server's environment file loaded") {
		t.Fatalf("error = %v, want a refusal naming the environment-file fix", err)
	}
}

// TestUnidecodeInstall_OwnershipChecksPrecedeAlreadyInstalled installs
// a valid table as the real owner, then makes every later ownership
// check see a foreign uid: the refusal must still fire, so the checks
// run before the already-installed check reads the table at all.
func TestUnidecodeInstall_OwnershipChecksPrecedeAlreadyInstalled(t *testing.T) {
	stateDir := t.TempDir()
	installTestTable(t, stateDir)

	opts := denyNetwork(t)
	opts.Geteuid = foreign
	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, opts)
	if err == nil {
		t.Fatal("error = nil, want a refusal: ownership checks must run before the already-installed check")
	}
	if strings.Contains(stdout.String(), "already installed") {
		t.Errorf("stdout = %q, want no already-installed message", stdout.String())
	}
}

// TestUnidecodeInstall_RefusesForeignLeafOwner proves the leaf's
// ownership refusal apart from the state directory's: Geteuid answers
// the state directory's check with the real owner, since the test
// process created it, and only the leaf's check sees a foreign uid.
func TestUnidecodeInstall_RefusesForeignLeafOwner(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, "unidecode")
	if err := os.Mkdir(leafPath, 0o755); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}

	real := os.Geteuid()
	calls := 0
	opts := denyNetwork(t)
	opts.Geteuid = func() int {
		calls++
		if calls == 1 {
			return real
		}
		return real + 1
	}

	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, opts)
	if err == nil || !strings.Contains(err.Error(), leafPath) || !strings.Contains(err.Error(), "with the server's environment file loaded") {
		t.Fatalf("error = %v, want a refusal naming %s and the environment-file fix", err, leafPath)
	}
}

func TestUnidecodeInstall_RefusesBadLeafShape(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T, leafPath string)
		want  string
	}{
		{
			name: "leaf is a file",
			setup: func(t *testing.T, leafPath string) {
				t.Helper()
				if err := os.WriteFile(leafPath, []byte("not a directory"), 0o644); err != nil {
					t.Fatalf("write file at leaf path: %v", err)
				}
			},
			want: "not a directory, want a directory",
		},
		{
			name: "leaf has the wrong mode",
			setup: func(t *testing.T, leafPath string) {
				t.Helper()
				if err := os.Mkdir(leafPath, 0o700); err != nil {
					t.Fatalf("mkdir leaf: %v", err)
				}
			},
			want: "mode 0700, want 0755",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			leafPath := filepath.Join(stateDir, "unidecode")
			tt.setup(t, leafPath)

			var stdout bytes.Buffer
			err := Install(context.Background(), stateDir, "", &stdout, denyNetwork(t))
			if err == nil {
				t.Fatal("error = nil, want a refusal")
			}
			if !strings.Contains(err.Error(), leafPath) || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want it to name %s and say %q", err, leafPath, tt.want)
			}
		})
	}
}

// TestUnidecodeInstall_RefusesUnreadableLeaf proves a leaf whose stat
// fails for a reason other than absence refuses, naming the leaf,
// rather than reading as absent and writing a table beside it.
func TestUnidecodeInstall_RefusesUnreadableLeaf(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions, so the leaf's stat cannot be made to fail this way")
	}
	stateDir := t.TempDir()
	if err := os.Mkdir(filepath.Join(stateDir, "unidecode"), 0o755); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}
	if err := os.Chmod(stateDir, 0o000); err != nil {
		t.Fatalf("chmod state directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(stateDir, 0o700) })

	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, denyNetwork(t))
	if err == nil || !strings.Contains(err.Error(), filepath.Join(stateDir, "unidecode")) {
		t.Fatalf("error = %v, want a refusal naming the leaf", err)
	}
	if !errors.Is(err, fs.ErrPermission) {
		t.Errorf("error = %v, want it to wrap fs.ErrPermission", err)
	}
}

// TestUnidecodeInstall_AlreadyInstalledMakesNoRequest proves the
// idempotent path: one that fetches anyway fails this on the request
// count.
func TestUnidecodeInstall_AlreadyInstalledMakesNoRequest(t *testing.T) {
	stateDir := t.TempDir()
	installTestTable(t, stateDir)

	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("unexpected request when a valid table is already installed")
	}))
	defer server.Close()
	client, rt := newTestClient(t, server)

	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, InstallOptions{Client: client, Geteuid: os.Geteuid})
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}
	if rt.requests != 0 {
		t.Errorf("requests = %d, want 0", rt.requests)
	}
	want := "unidecode: already installed at " + filepath.Join(stateDir, "unidecode", TableFileName) + " (version " + Version + ")\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

// TestUnidecodeInstall_ReplacesInvalidTable proves an invalid table on
// disk is rebuilt and replaced, and the old bytes never survive a
// success.
func TestUnidecodeInstall_ReplacesInvalidTable(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, "unidecode")
	if err := os.MkdirAll(leafPath, 0o755); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}
	path := filepath.Join(leafPath, TableFileName)
	oldBytes := []byte("not a valid table")
	if err := os.WriteFile(path, oldBytes, 0o644); err != nil {
		t.Fatalf("write invalid table: %v", err)
	}

	var stdout bytes.Buffer
	if err := Install(context.Background(), stateDir, cachedWheelPath(t), &stdout, denyNetwork(t)); err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if bytes.Equal(got, oldBytes) {
		t.Fatal("table bytes unchanged, want the invalid table replaced")
	}
	if _, err := Parse(got); err != nil {
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
			err := Install(context.Background(), stateDir, from, &stdout, denyNetwork(t))
			if err == nil || !strings.Contains(err.Error(), from) {
				t.Fatalf("error = %v, want a refusal naming %s", err, from)
			}
			if _, statErr := os.Stat(filepath.Join(stateDir, "unidecode")); !errors.Is(statErr, fs.ErrNotExist) {
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
	err = Install(context.Background(), stateDir, fromPath, &stdout, denyNetwork(t))
	if err == nil || !strings.Contains(err.Error(), "exceeds maximum size") || !strings.Contains(err.Error(), fromPath) {
		t.Fatalf("error = %v, want a size refusal naming %s", err, fromPath)
	}
	if _, statErr := os.Stat(filepath.Join(stateDir, "unidecode")); !errors.Is(statErr, fs.ErrNotExist) {
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
	err := Install(context.Background(), stateDir, fromPath, &stdout, denyNetwork(t))
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") || !strings.Contains(err.Error(), fromPath) {
		t.Fatalf("error = %v, want a checksum refusal naming %s", err, fromPath)
	}
	if !strings.Contains(err.Error(), WheelSHA256) {
		t.Errorf("error = %q, want it to name the pinned sha256 %s", err.Error(), WheelSHA256)
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
	client, _ := newTestClient(t, server)

	stateDir := t.TempDir()
	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, InstallOptions{Client: client, Geteuid: os.Geteuid})
	if err == nil {
		t.Fatal("error = nil, want a download refusal")
	}
	for _, want := range []string{WheelURL, WheelSHA256, "--from"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err.Error(), want)
		}
	}
	if _, statErr := os.Stat(filepath.Join(stateDir, "unidecode")); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("leaf exists after a refused download: %v", statErr)
	}
}

// TestUnidecodeInstall_FailedDownloadLeavesExistingFileUntouched
// proves a failed download, attempting to replace an invalid table,
// never removes what was already there.
func TestUnidecodeInstall_FailedDownloadLeavesExistingFileUntouched(t *testing.T) {
	stateDir := t.TempDir()
	leafPath := filepath.Join(stateDir, "unidecode")
	if err := os.MkdirAll(leafPath, 0o755); err != nil {
		t.Fatalf("mkdir leaf: %v", err)
	}
	path := filepath.Join(leafPath, TableFileName)
	existing := []byte("stale but present")
	if err := os.WriteFile(path, existing, 0o644); err != nil {
		t.Fatalf("write existing table: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	client, _ := newTestClient(t, server)

	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, InstallOptions{Client: client, Geteuid: os.Geteuid})
	if err == nil {
		t.Fatal("error = nil, want a download refusal")
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if !bytes.Equal(got, existing) {
		t.Error("table bytes changed after a failed download")
	}
}

// TestUnidecodeInstall_Success proves the success path end to end
// through the network: the leaf's mode is set explicitly rather than
// left to the umask, the table file is 0644, the written table
// round-trips through Parse, and the output names the path and
// version.
func TestUnidecodeInstall_Success(t *testing.T) {
	wheel := loadCachedWheel(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(wheel)
	}))
	defer server.Close()
	client, rt := newTestClient(t, server)

	stateDir := t.TempDir()

	// This test runs serially (this package runs no test in parallel):
	// a restrictive umask proves the leaf's 0755 mode is set by an
	// explicit os.Chmod, not left to what os.Mkdir produces under the
	// umask in effect.
	oldMask := syscall.Umask(0o027)
	defer syscall.Umask(oldMask)

	var stdout bytes.Buffer
	err := Install(context.Background(), stateDir, "", &stdout, InstallOptions{Client: client, Geteuid: os.Geteuid})
	if err != nil {
		t.Fatalf("Install() error = %v, want nil", err)
	}
	if rt.requests != 1 {
		t.Errorf("requests = %d, want 1", rt.requests)
	}

	leafPath := filepath.Join(stateDir, "unidecode")
	leafInfo, err := os.Stat(leafPath)
	if err != nil {
		t.Fatalf("stat leaf: %v", err)
	}
	if mode := leafInfo.Mode().Perm(); mode != 0o755 {
		t.Errorf("leaf mode = %04o, want 0755", mode)
	}

	path := filepath.Join(leafPath, TableFileName)
	tableInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat table: %v", err)
	}
	if mode := tableInfo.Mode().Perm(); mode != 0o644 {
		t.Errorf("table mode = %04o, want 0644", mode)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read table: %v", err)
	}
	if _, err := Parse(data); err != nil {
		t.Errorf("Parse(installed table): %v", err)
	}

	want := "unidecode: installed " + path + " (version " + Version + ")\n"
	if got := stdout.String(); got != want {
		t.Errorf("stdout = %q, want %q", got, want)
	}
}

// TestUnidecodeInstall_CreatesNothingUnderStateDirOnRefusal proves the
// state directory is still empty after each refusal that runs before
// any write.
func TestUnidecodeInstall_CreatesNothingUnderStateDirOnRefusal(t *testing.T) {
	tests := []struct {
		name string
		// setup returns the options and the --from path the row needs.
		setup func(t *testing.T) (InstallOptions, string)
	}{
		{"foreign state directory owner", func(t *testing.T) (InstallOptions, string) {
			opts := denyNetwork(t)
			opts.Geteuid = foreign
			return opts, ""
		}},
		{"absent --from file", func(t *testing.T) (InstallOptions, string) {
			return denyNetwork(t), filepath.Join(t.TempDir(), "missing.whl")
		}},
		{"--from directory", func(t *testing.T) (InstallOptions, string) {
			return denyNetwork(t), t.TempDir()
		}},
		{"--from file that is not the pinned wheel", func(t *testing.T) (InstallOptions, string) {
			path := filepath.Join(t.TempDir(), "wrong.whl")
			if err := os.WriteFile(path, []byte("not the pinned wheel"), 0o644); err != nil {
				t.Fatalf("write scratch file: %v", err)
			}
			return denyNetwork(t), path
		}},
		{"refused download", func(t *testing.T) (InstallOptions, string) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			}))
			t.Cleanup(server.Close)
			client, _ := newTestClient(t, server)
			return InstallOptions{Client: client, Geteuid: os.Geteuid}, ""
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			opts, from := tt.setup(t)

			var stdout bytes.Buffer
			if err := Install(context.Background(), stateDir, from, &stdout, opts); err == nil {
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
