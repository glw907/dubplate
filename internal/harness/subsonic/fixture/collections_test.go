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
	a := namedDigest{"first", strings.Repeat("a", 64)}
	b := namedDigest{"second", strings.Repeat("b", 64)}
	if got := combineDigests([]namedDigest{a}); got != a.digest {
		t.Errorf("one collection digest = %s, want the library's own %s", got, a.digest)
	}
	both := combineDigests([]namedDigest{a, b})
	if len(both) != 64 || both == a.digest || both == b.digest {
		t.Errorf("combined digest = %q, want a fresh 64-hex digest", both)
	}
	tests := []struct {
		name  string
		other []namedDigest
	}{
		{"names swapped over the same libraries", []namedDigest{{"second", a.digest}, {"first", b.digest}}},
		{"order reversed", []namedDigest{b, a}},
		{"a library changed", []namedDigest{a, {"second", strings.Repeat("c", 64)}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := combineDigests(tt.other); got == both {
				t.Errorf("digest unchanged: %s", got)
			}
		})
	}
	// Frozen so a change to the line format is a decision, not an accident.
	if got, want := combineDigests([]namedDigest{a, b}), "c527b2d9b3160095209e88e3623c63b683e47fb2f7419097402966fc9d9cea4b"; got != want {
		t.Errorf("combined digest = %s, want %s", got, want)
	}
}

func TestBuildCollections(t *testing.T) {
	// The two committed collections under these names, as the
	// music-folders session records them.
	const twoDigest = "7e342ed0efcdacc17d1653d0a5f265b6e13062e07f9f4b3eec631a6f929ca433"
	for _, name := range []string{DefaultCollectionName, "renamed"} {
		one, digest, err := BuildCollections(t.TempDir(), []Collection{{name, committedManifest}})
		if err != nil {
			t.Fatal(err)
		}
		if len(one) != 1 || one[0].Name != name || len(one[0].Library.Tracks()) != 29 {
			t.Errorf("%s: built %+v, want one collection of 29 tracks", name, one)
		}
		if digest != seedLibraryDigest {
			t.Errorf("%s: digest = %s, want the seed library's own %s", name, digest, seedLibraryDigest)
		}
	}

	root := t.TempDir()
	cols := []Collection{{"first", collectionManifests[0]}, {"second", collectionManifests[1]}}
	two, digest, err := BuildCollections(root, cols)
	if err != nil {
		t.Fatal(err)
	}
	if digest != twoDigest {
		t.Errorf("two collections digest = %s, want %s", digest, twoDigest)
	}
	if len(two) != 2 || two[0].Name != "first" || two[1].Name != "second" || two[0].Dir == two[1].Dir {
		t.Fatalf("built %+v, want first and second in distinct directories", two)
	}
	for _, c := range two {
		if !strings.HasPrefix(c.Dir, root) || len(readTree(t, c.Dir)) == 0 {
			t.Errorf("%s: no library built under the root at %s", c.Name, c.Dir)
		}
	}
	_, swapped, err := BuildCollections(t.TempDir(), []Collection{{"second", cols[0].Manifest}, {"first", cols[1].Manifest}})
	if err != nil {
		t.Fatal(err)
	}
	if swapped == digest {
		t.Errorf("the same libraries under swapped names digest alike: %s", swapped)
	}

	_, _, err = BuildCollections(t.TempDir(), []Collection{{"gone", root + "/absent.json"}})
	if err == nil || !strings.Contains(err.Error(), `collection "gone"`) {
		t.Errorf("missing manifest error = %v, want one naming the collection", err)
	}
}
