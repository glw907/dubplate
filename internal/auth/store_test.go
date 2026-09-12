package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpen_CreatesDirectoriesWithPermissions(t *testing.T) {
	stateDir := t.TempDir()

	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	for _, sub := range []string{"", "nonce", "epoch"} {
		dir := filepath.Join(StoreDir(stateDir), sub)
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("stat %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("%s is not a directory", dir)
		}
		if got := info.Mode().Perm(); got != 0o700 {
			t.Errorf("%s: mode = %o, want 0700", dir, got)
		}
	}

	// A record file Write creates is 0600, exercised through the
	// package's own write path rather than a permission the test states
	// by hand.
	path, err := store.NoncePath(RecordID("fixed-vector"))
	if err != nil {
		t.Fatalf("NoncePath: %v", err)
	}
	if err := store.Write(path, []byte("x")); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("record file mode = %o, want 0600", got)
	}
}

func TestRecordID_Vector(t *testing.T) {
	// Checked-in vector: sha256("rung5b-task7-fixed-vector") in hex,
	// computed independently of this package so the test cannot pass by
	// echoing the code under test.
	const (
		input = "rung5b-task7-fixed-vector"
		want  = "ba0bf54aed001a32c633dea11f7949ee0165cc30c8175ac317162197eabc51b5"
	)

	got := RecordID(input)
	if got != want {
		t.Errorf("RecordID(%q) = %s, want %s", input, got, want)
	}
	if len(got) != 64 {
		t.Errorf("RecordID(%q) length = %d, want 64", input, len(got))
	}
}

func TestValidID(t *testing.T) {
	tests := []struct {
		name string
		id   string
		want bool
	}{
		{"empty", "", false},
		{"dot", ".", false},
		{"dotdot", "..", false},
		{"forward slash", "a/b", false},
		{"backslash", `a\b`, false},
		{"valid digest", RecordID("some-value"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validID(tt.id); got != tt.want {
				t.Errorf("validID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// TestConfinement writes through every entry point this package offers
// (Open's directory creation, NoncePath/EpochPath's path derivation, and
// Write's atomic write) and proves, by walking the filesystem from
// stateDir, that no file lands outside root. It is the static argument
// the unit's ReadWritePaths grant rests on at 5d.
func TestConfinement(t *testing.T) {
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	root := StoreDir(stateDir)

	// An address whose local part carries "..": hashed first, so the
	// derived id is always a safe 64-character digest and the write
	// through NoncePath succeeds.
	const address = "a..b@example.com"
	id := RecordID(address)
	noncePath, err := store.NoncePath(id)
	if err != nil {
		t.Fatalf("NoncePath(%s): %v", id, err)
	}
	epochPath, err := store.EpochPath(id)
	if err != nil {
		t.Fatalf("EpochPath(%s): %v", id, err)
	}
	if err := store.Write(noncePath, []byte("nonce")); err != nil {
		t.Fatalf("write nonce record: %v", err)
	}
	if err := store.Write(epochPath, []byte("epoch")); err != nil {
		t.Fatalf("write epoch record: %v", err)
	}

	// An attacker-shaped id is refused before any file is created.
	const attackerID = "../../evil"
	if _, err := store.NoncePath(attackerID); err == nil {
		t.Fatalf("NoncePath(%q): want error, got nil", attackerID)
	}
	if _, err := store.EpochPath(attackerID); err == nil {
		t.Fatalf("EpochPath(%q): want error, got nil", attackerID)
	}

	var found []string
	err = filepath.WalkDir(stateDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		found = append(found, path)
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			t.Errorf("file %s escaped root %s", path, root)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", stateDir, err)
	}
	if len(found) != 2 {
		t.Fatalf("found %d files under stateDir, want 2: %v", len(found), found)
	}

	// The raw address never appears in a created filename: the other
	// half of R34's argument that the id is a digest, not the value.
	for _, path := range found {
		name := filepath.Base(path)
		if name == address {
			t.Errorf("filename %s carries the raw address", name)
		}
	}
}
