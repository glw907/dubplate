package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/atomicfile"
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

	// A record file this package writes through atomicfile.WriteFile
	// (the entry point [NonceStore] and [EpochStore] use) is 0600,
	// exercised through that path rather than a permission the test
	// states by hand.
	path, err := store.NoncePath(RecordID(nonceSubdir, "fixed-vector"))
	if err != nil {
		t.Fatalf("NoncePath: %v", err)
	}
	if err := atomicfile.WriteFile(path, []byte("x"), filePerm); err != nil {
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

// TestOpen_RefusesWideLeaf proves the review m3 fix: a leaf directory
// created wide before Open ever runs is refused rather than silently
// admitted, because os.MkdirAll is a no-op on an existing directory
// whatever its mode.
func TestOpen_RefusesWideLeaf(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(StoreDir(stateDir), 0o755); err != nil {
		t.Fatalf("pre-create wide leaf: %v", err)
	}

	_, err := Open(stateDir)
	if err == nil {
		t.Fatal("Open() error = nil, want a refusal on a leaf wider than 0700")
	}
	if !strings.Contains(err.Error(), StoreDir(stateDir)) {
		t.Errorf("error = %q, want it to name the leaf path", err.Error())
	}
}

// TestOpen_AdmitsNarrowerLeaf proves the refusal is one-sided: a leaf
// already at exactly dirPerm, the mode Open itself would have created
// it at, starts cleanly.
func TestOpen_AdmitsNarrowerLeaf(t *testing.T) {
	stateDir := t.TempDir()
	if err := os.MkdirAll(StoreDir(stateDir), 0o700); err != nil {
		t.Fatalf("pre-create leaf: %v", err)
	}
	if _, err := Open(stateDir); err != nil {
		t.Errorf("Open() error = %v, want nil for a leaf already at dirPerm", err)
	}
}

func TestRecordID_Vector(t *testing.T) {
	// Checked-in vectors: sha256 of the length-prefixed kind and value
	// together (review m12), computed independently of this package so
	// the test cannot pass by echoing the code under test.
	const (
		input     = "rung5b-task7-fixed-vector"
		wantNonce = "8e416bb0b99243386f69bd2cb7b5d1566c8a59beabd823f4fddccc598768dac4"
		wantEpoch = "c333f9dd4b2043340c7aedc69d35f9896dc1f19de703a35373c330035217c9d5"
	)

	if got := RecordID(nonceSubdir, input); got != wantNonce {
		t.Errorf("RecordID(%q, %q) = %s, want %s", nonceSubdir, input, got, wantNonce)
	}
	if got := RecordID(epochSubdir, input); got != wantEpoch {
		t.Errorf("RecordID(%q, %q) = %s, want %s", epochSubdir, input, got, wantEpoch)
	}
	if len(RecordID(nonceSubdir, input)) != 64 {
		t.Errorf("RecordID length = %d, want 64", len(RecordID(nonceSubdir, input)))
	}
}

// TestRecordID_KindSeparatesTheDigest proves the review m12 fix
// directly: the same value under two different kinds never collides,
// which is what lets a future caller place both kinds of record in one
// directory without a silent overwrite.
func TestRecordID_KindSeparatesTheDigest(t *testing.T) {
	const value = "member@example.com"
	if RecordID(nonceSubdir, value) == RecordID(epochSubdir, value) {
		t.Errorf("RecordID(%q, ...) == RecordID(%q, ...), want them to differ", nonceSubdir, epochSubdir)
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
		{"valid digest", RecordID(nonceSubdir, "some-value"), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validID(tt.id); got != tt.want {
				t.Errorf("validID(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

// TestRecordPathRefusalNamesNoValue proves the review m7 fix: an
// invalid id's refusal names its length and its offending character
// class, never the id itself, so a future caller that skipped
// [RecordID] could not leak a raw credential into an error string.
func TestRecordPathRefusalNamesNoValue(t *testing.T) {
	stateDir := t.TempDir()
	store, err := Open(stateDir)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	const secret = "attacker-controlled-nonce-value"
	_, err = store.NoncePath("../" + secret)
	if err == nil {
		t.Fatal("NoncePath() error = nil, want a refusal")
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("error = %q, carries the raw value it must never echo", err.Error())
	}
	if !strings.Contains(err.Error(), "contains a path separator") {
		t.Errorf("error = %q, want it to name the offending character class", err.Error())
	}
}

// TestConfinement writes through every entry point this package offers
// (Open's directory creation, NoncePath/EpochPath's path derivation,
// and atomicfile.WriteFile) and proves, by walking the filesystem from
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
	nonceID := RecordID(nonceSubdir, address)
	epochID := RecordID(epochSubdir, address)
	noncePath, err := store.NoncePath(nonceID)
	if err != nil {
		t.Fatalf("NoncePath(%s): %v", nonceID, err)
	}
	epochPath, err := store.EpochPath(epochID)
	if err != nil {
		t.Fatalf("EpochPath(%s): %v", epochID, err)
	}
	if err := atomicfile.WriteFile(noncePath, []byte("nonce"), filePerm); err != nil {
		t.Fatalf("write nonce record: %v", err)
	}
	if err := atomicfile.WriteFile(epochPath, []byte("epoch"), filePerm); err != nil {
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
