package store

import (
	"errors"
	"io/fs"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"testing"
)

func TestNewLeaf(t *testing.T) {
	stateDir := "/var/lib/dubplate"
	l := newLeaf(stateDir)

	want := leaf{
		root:              "/var/lib/dubplate/store",
		lockPath:          "/var/lib/dubplate/store/store.lock",
		dbDir:             "/var/lib/dubplate/store/db",
		dbPath:            "/var/lib/dubplate/store/db/dubplate.db",
		snapshotsDir:      "/var/lib/dubplate/store/snapshots",
		stagingDir:        "/var/lib/dubplate/store/staging",
		restoreIntentPath: "/var/lib/dubplate/store/restore-intent",
	}
	if l != want {
		t.Errorf("newLeaf(%q) = %+v, want %+v", stateDir, l, want)
	}
}

// assertLeafShape fails the test unless every directory createLeaf
// makes exists at leafDirPerm and store.lock exists as a regular file.
func assertLeafShape(t *testing.T, l leaf) {
	t.Helper()
	for _, dir := range []string{l.root, l.dbDir, l.snapshotsDir, l.stagingDir} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Errorf("%s: not a directory", dir)
		}
		if mode := info.Mode().Perm(); mode != leafDirPerm {
			t.Errorf("%s: mode %o, want %o", dir, mode, leafDirPerm)
		}
	}
	info, err := os.Stat(l.lockPath)
	if err != nil {
		t.Fatalf("stat %s: %v", l.lockPath, err)
	}
	if info.IsDir() {
		t.Errorf("%s: is a directory, want a regular file", l.lockPath)
	}
}

func TestCreateLeaf_FirstRunAndIdempotentRetry(t *testing.T) {
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	euid := os.Geteuid()

	if err := createLeaf(l, stateDir, euid); err != nil {
		t.Fatalf("first createLeaf: %v", err)
	}
	assertLeafShape(t, l)

	if err := createLeaf(l, stateDir, euid); err != nil {
		t.Fatalf("second createLeaf: %v", err)
	}
	assertLeafShape(t, l)
}

func TestCreateLeaf_CompletesAPartialCreate(t *testing.T) {
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	euid := os.Geteuid()

	// A crash after the leaf and store.lock land but before db/,
	// snapshots/, or staging/ do.
	if err := os.MkdirAll(l.root, leafDirPerm); err != nil {
		t.Fatalf("plant partial leaf: %v", err)
	}
	f, err := os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, lockFilePerm)
	if err != nil {
		t.Fatalf("plant partial store.lock: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	if err := createLeaf(l, stateDir, euid); err != nil {
		t.Fatalf("createLeaf over a partial leaf: %v", err)
	}
	assertLeafShape(t, l)
}

func TestCreateLeaf_OwnerMismatchRefusesAndCreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	mismatchEuid := os.Geteuid() + 1

	err := createLeaf(l, stateDir, mismatchEuid)
	if err == nil {
		t.Fatal("createLeaf: want a refusal for a mismatched euid, got nil")
	}
	if !strings.Contains(err.Error(), "run as that user") {
		t.Errorf("createLeaf error %q does not name the fix", err)
	}
	if _, statErr := os.Stat(l.root); !errors.Is(statErr, fs.ErrNotExist) {
		t.Errorf("createLeaf: want no leaf on refusal, stat returned %v", statErr)
	}
}

// TestCreateLeaf_OwnerCheckRunsOnARetryToo proves the owner check runs
// every time createLeaf is called, not only when the leaf is absent:
// a second call against an already-created leaf, with a mismatched
// euid, must still refuse.
func TestCreateLeaf_OwnerCheckRunsOnARetryToo(t *testing.T) {
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	euid := os.Geteuid()

	if err := createLeaf(l, stateDir, euid); err != nil {
		t.Fatalf("first createLeaf: %v", err)
	}

	mismatchEuid := euid + 1
	if err := createLeaf(l, stateDir, mismatchEuid); err == nil {
		t.Fatal("createLeaf: want a refusal on retry with a mismatched euid, got nil")
	}
}

func TestCheckOwner(t *testing.T) {
	dir := t.TempDir()
	info, err := os.Stat(dir)
	if err != nil {
		t.Fatal(err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		t.Fatal("no unix stat info for a temp directory")
	}
	realUID := int(st.Uid)

	tests := []struct {
		name       string
		euid       int
		wantErr    bool
		wantSubstr string
	}{
		{name: "matching euid", euid: realUID, wantErr: false},
		{name: "mismatched euid", euid: realUID + 1, wantErr: true, wantSubstr: "run as that user"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkOwner(info, tt.euid)
			if tt.wantErr && err == nil {
				t.Fatal("checkOwner: want an error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("checkOwner: want no error, got %v", err)
			}
			if tt.wantSubstr != "" && !strings.Contains(err.Error(), tt.wantSubstr) {
				t.Errorf("checkOwner error %q does not contain %q", err, tt.wantSubstr)
			}
		})
	}
}

func TestOwnerName(t *testing.T) {
	cur, err := user.Current()
	if err != nil {
		t.Skipf("no current user info: %v", err)
	}
	uid, err := strconv.Atoi(cur.Uid)
	if err != nil {
		t.Fatal(err)
	}

	got := ownerName(uint32(uid))
	if got != cur.Username {
		t.Errorf("ownerName(%d) = %q, want %q", uid, got, cur.Username)
	}
}

func TestOwnerName_UnknownUID(t *testing.T) {
	const unknownUID = 4294967294 // near uint32's top; no passwd entry
	got := ownerName(unknownUID)
	want := "uid:4294967294"
	if got != want {
		t.Errorf("ownerName(%d) = %q, want %q", unknownUID, got, want)
	}
}
