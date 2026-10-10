package fixture

import (
	"slices"
	"strings"
	"testing"
)

// seedLibraryDigest is the library digest the committed seed manifest
// builds to, as the seed session records it.
const seedLibraryDigest = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"

var collectionManifests = []string{
	"../data/fixture/collection-a.json",
	"../data/fixture/collection-b.json",
}

func TestSeedManifestDigestIsUnchanged(t *testing.T) {
	lib, err := Load(committedManifest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := lib.Build(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if got != seedLibraryDigest {
		t.Errorf("seed library digest = %s, want %s", got, seedLibraryDigest)
	}
}

func TestCollectionManifestsBuildIdentically(t *testing.T) {
	for _, path := range collectionManifests {
		t.Run(path, func(t *testing.T) {
			lib, err := Load(path)
			if err != nil {
				t.Fatal(err)
			}
			dirA, dirB := t.TempDir(), t.TempDir()
			digA, err := lib.Build(dirA)
			if err != nil {
				t.Fatal(err)
			}
			digB, err := lib.Build(dirB)
			if err != nil {
				t.Fatal(err)
			}
			if digA != digB {
				t.Errorf("digests differ: %s, %s", digA, digB)
			}
			a, b := readTree(t, dirA), readTree(t, dirB)
			if len(a) == 0 {
				t.Fatal("built no files")
			}
			for rel, data := range a {
				if string(b[rel]) != string(data) {
					t.Errorf("%s differs between builds", rel)
				}
			}
			if len(a) != len(b) {
				t.Errorf("file counts differ: %d, %d", len(a), len(b))
			}
		})
	}
}

func TestCollectionManifestsShape(t *testing.T) {
	var libs []*Library
	for _, path := range collectionManifests {
		lib, err := Load(path)
		if err != nil {
			t.Fatal(err)
		}
		libs = append(libs, lib)
		if n := len(lib.manifest.Albums); n < 3 || n > 4 {
			t.Errorf("%s declares %d albums, want 3 or 4", path, n)
		}
	}

	paths := make(map[string]bool)
	albums := make(map[string]bool)
	artists := make([]map[string]bool, len(libs))
	for i, lib := range libs {
		artists[i] = make(map[string]bool)
		for _, a := range lib.manifest.Albums {
			key := a.Artist + "/" + a.Title
			if albums[key] {
				t.Errorf("album %q appears in both collections", key)
			}
			albums[key] = true
			artists[i][a.Artist] = true
			if a.Title == "Loose files" {
				t.Errorf("collection %d carries the untagged album", i)
			}
			for _, tr := range a.Tracks {
				if paths[tr.Path] {
					t.Errorf("track path %q appears in both collections", tr.Path)
				}
				paths[tr.Path] = true
				if !slices.Equal(tr.Tags.Genres, []string{"Ambient"}) {
					t.Errorf("track %q genres = %q, want the one shared value", tr.Path, tr.Tags.Genres)
				}
			}
		}
	}
	var shared []string
	for a := range artists[0] {
		if artists[1][a] {
			shared = append(shared, a)
		}
	}
	if len(shared) != 1 {
		t.Errorf("album artists in both collections = %q, want exactly one", shared)
	}
}

func TestCombineDigests(t *testing.T) {
	a := NamedDigest{"first", strings.Repeat("a", 64)}
	b := NamedDigest{"second", strings.Repeat("b", 64)}
	if got := CombineDigests([]NamedDigest{a}); got != a.Digest {
		t.Errorf("one collection digest = %s, want the library's own %s", got, a.Digest)
	}
	both := CombineDigests([]NamedDigest{a, b})
	if len(both) != 64 || both == a.Digest || both == b.Digest {
		t.Errorf("combined digest = %q, want a fresh 64-hex digest", both)
	}
	tests := []struct {
		name  string
		other []NamedDigest
	}{
		{"names swapped over the same libraries", []NamedDigest{{"second", a.Digest}, {"first", b.Digest}}},
		{"order reversed", []NamedDigest{b, a}},
		{"a library changed", []NamedDigest{a, {"second", strings.Repeat("c", 64)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CombineDigests(tt.other); got == both {
				t.Errorf("digest unchanged: %s", got)
			}
		})
	}
	// Frozen so a change to the line format is a decision, not an accident.
	if got, want := CombineDigests([]NamedDigest{a, b}), "c527b2d9b3160095209e88e3623c63b683e47fb2f7419097402966fc9d9cea4b"; got != want {
		t.Errorf("combined digest = %s, want %s", got, want)
	}
}
