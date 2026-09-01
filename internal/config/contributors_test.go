package config

import (
	"os"
	"path/filepath"
	"testing"
)

const fixtureYAML = `
contributors:
  - name: alice
    email: alice@example.com
    role: contributor
  - name: bob
    email: bob@example.com
    role: contributor
`

func writeFixture(t *testing.T, contents string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "contributors.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func TestLoadContributors_knownUploader(t *testing.T) {
	path := writeFixture(t, fixtureYAML)

	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}

	email, ok := r.EmailFor("alice")
	if !ok || email != "alice@example.com" {
		t.Errorf("EmailFor(alice) = (%q, %v), want (alice@example.com, true)", email, ok)
	}
	first, ok := r.FirstName("alice")
	if !ok || first != "alice" {
		t.Errorf("FirstName(alice) = (%q, %v), want (alice, true)", first, ok)
	}
	if !r.Has("alice") {
		t.Error("Has(alice) = false, want true")
	}
}

// --- unknown uploader (absent from contributors.yaml) ----------------------
//
// Re-expresses tests/music-import.bats's "skip path: an unmatched uploader
// is not emailed" case: uploader_email/uploader_name returned empty for an
// unmatched name, with no error. EmailFor/FirstName/Has report absent the
// same way.

func TestLoadContributors_unknownUploader(t *testing.T) {
	path := writeFixture(t, fixtureYAML)

	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}

	if _, ok := r.EmailFor("carol"); ok {
		t.Error("EmailFor(carol) reported present, want absent")
	}
	if _, ok := r.FirstName("carol"); ok {
		t.Error("FirstName(carol) reported present, want absent")
	}
	if r.Has("carol") {
		t.Error("Has(carol) = true, want false")
	}
}

// --- missing contributors.yaml -----------------------------------------
//
// Re-expresses the shell layer's documented behavior (scripts/music-import,
// lines 111-116: load_contributors_json prints empty output for a missing
// CONTRIBUTORS_FILE, "matching uploader_email/uploader_name's prior
// behavior of treating 'no file' the same as 'no match'"): a missing file
// is an empty roster, not an error.

func TestLoadContributors_missingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "does-not-exist.yaml")

	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil for a missing file", err)
	}
	if r.Has("alice") {
		t.Error("Has(alice) = true on an empty roster, want false")
	}
	if _, ok := r.EmailFor("alice"); ok {
		t.Error("EmailFor(alice) reported present on an empty roster, want absent")
	}
	if _, ok := r.FirstName("alice"); ok {
		t.Error("FirstName(alice) reported present on an empty roster, want absent")
	}
}

func TestLoadContributors_malformedYAML(t *testing.T) {
	path := writeFixture(t, "contributors: [this is not a mapping list")

	if _, err := LoadContributors(path); err == nil {
		t.Fatal("LoadContributors() error = nil, want non-nil for malformed YAML")
	}
}

func TestLoadContributors_multiWordName(t *testing.T) {
	path := writeFixture(t, `
contributors:
  - name: Geoff Wright
    email: geoff-login@907.life
    role: contributor
`)

	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}
	first, ok := r.FirstName("Geoff Wright")
	if !ok || first != "Geoff" {
		t.Errorf("FirstName(\"Geoff Wright\") = (%q, %v), want (Geoff, true)", first, ok)
	}
}

func TestRoster_zeroValueIsEmpty(t *testing.T) {
	var r Roster

	if r.Has("anyone") {
		t.Error("Has(anyone) on zero Roster = true, want false")
	}
	if _, ok := r.EmailFor("anyone"); ok {
		t.Error("EmailFor(anyone) on zero Roster reported present, want absent")
	}
	if _, ok := r.DirNameFor("anyone@example.com"); ok {
		t.Error("DirNameFor(anyone@example.com) on zero Roster reported present, want absent")
	}
}

// --- DirNameFor: the uploader's email-to-directory-name identity index ----

func TestRoster_DirNameForRoundTrip(t *testing.T) {
	path := writeFixture(t, fixtureYAML)
	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}

	email, ok := r.EmailFor("alice")
	if !ok {
		t.Fatal("EmailFor(alice) reported absent, want present")
	}
	dirName, ok := r.DirNameFor(email)
	if !ok || dirName != "alice" {
		t.Errorf("DirNameFor(%q) = (%q, %v), want (alice, true)", email, dirName, ok)
	}
}

func TestRoster_DirNameForCaseInsensitive(t *testing.T) {
	path := writeFixture(t, fixtureYAML)
	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}

	dirName, ok := r.DirNameFor("ALICE@EXAMPLE.COM")
	if !ok || dirName != "alice" {
		t.Errorf("DirNameFor(ALICE@EXAMPLE.COM) = (%q, %v), want (alice, true)", dirName, ok)
	}
}

func TestRoster_DirNameForUnknownEmail(t *testing.T) {
	path := writeFixture(t, fixtureYAML)
	r, err := LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors() error = %v, want nil", err)
	}

	if _, ok := r.DirNameFor("nobody@example.com"); ok {
		t.Error("DirNameFor(nobody@example.com) reported present, want absent")
	}
}
