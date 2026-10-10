package grade

import (
	"errors"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/session"
)

const (
	collectionA = "../data/fixture/collection-a.json"
	collectionB = "../data/fixture/collection-b.json"
)

// TestGradeRefusesASessionRecordedOverOtherCollections grades the seed
// session, recorded over one library, with collection lists that combine to
// other digests.
func TestGradeRefusesASessionRecordedOverOtherCollections(t *testing.T) {
	seed, err := session.Load(seedSession)
	if err != nil {
		t.Fatal(err)
	}
	recorded := seed.Sources["navidrome"].LibraryDigest

	tests := []struct {
		name        string
		collections []Collection
		combined    string // the digest the refusal must name beside the recorded one
	}{
		{
			name:        "two collections",
			collections: []Collection{{"first", collectionA}, {"second", collectionB}},
			combined:    "7e342ed0efcdacc17d1653d0a5f265b6e13062e07f9f4b3eec631a6f929ca433",
		},
		{
			name:        "another single collection",
			collections: []Collection{{"first", collectionA}},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := seedConfig(seedLedger)
			cfg.Collections = tc.collections
			g, err := New(cfg)
			if err != nil {
				t.Fatal(err)
			}
			r, err := g.Grade(t.Context(), sessionsIn+"/seed", pairNG)
			if !errors.Is(err, errCollectionMismatch) {
				t.Fatalf("error = %v, want one wrapping errCollectionMismatch", err)
			}
			if r != nil {
				t.Error("a refused grade returned a report")
			}
			if !strings.Contains(err.Error(), recorded) || !strings.Contains(err.Error(), g.digest) {
				t.Errorf("error %q must name the recorded digest %s and the collections' %s", err, recorded, g.digest)
			}
			if tc.combined != "" && g.digest != tc.combined {
				t.Errorf("combined digest = %s, want %s", g.digest, tc.combined)
			}
		})
	}
}
