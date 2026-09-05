package analysis

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// albumInfoTestLookups is how many release-lookup requests each subject
// case logs, counted off the committed corpus at 2026-09-04. The stand-in
// cases below reproduce those counts so the fixture ledger has the same
// fifty rows the real one does, and so the row floor is exercised by a
// ledger that actually reaches it.
var albumInfoTestLookups = map[string]int{
	"clean-match":                             1,
	"media-baseline":                          1,
	"media-uneven-split":                      1,
	"va-exact":                                1,
	"various-artists":                         1,
	"should-reject":                           6,
	"non-latin-pseudo-release":                2,
	"real-nonlatin-zemfira":                   2,
	"real-nonlatin-zemfira-stock":             2,
	"real-nonlatin-girls-generation":          6,
	"real-nonlatin-anna-vissi-essential":      1,
	"real-nonlatin-alla-pugacheva-vol1":       1,
	"real-nonlatin-x-japan-art-of-life":       1,
	"real-nonlatin-utada-first-love":          6,
	"real-nonlatin-utada-first-love-locale":   6,
	"real-nonlatin-hamasaki-ayumi":            5,
	"real-genre-metallica-master-of-puppets":  1,
	"real-compilation-queen-greatest-hits-ii": 1,
	"real-noise-dylan-bootleg":                5,
}

// albumInfoTestMBID is the invented release id the stand-in corpus uses
// for a case's nth lookup. It is not a real MBID: the checker compares
// the ledger's ids against the case's own log and never parses them.
func albumInfoTestMBID(caseID string, n int) string {
	return fmt.Sprintf("%s-release-%d", caseID, n)
}

// fullAlbumInfoCases builds a committed-corpus stand-in carrying every
// AlbumInfoSubjects id, each with its measured number of release-lookup
// requests plus one search request nothing should ever derive a row for.
func fullAlbumInfoCases() map[string]*cassette.Case {
	cases := make(map[string]*cassette.Case, len(AlbumInfoSubjects)+1)
	for _, s := range AlbumInfoSubjects {
		c := &cassette.Case{ID: s.CaseID}
		for n := range albumInfoTestLookups[s.CaseID] {
			c.Requests = append(c.Requests, cassette.Request{
				Method:    "GET",
				PathQuery: "/ws/2/release/" + albumInfoTestMBID(s.CaseID, n) + "?fmt=json&inc=artists",
			})
		}
		c.Requests = append(c.Requests, cassette.Request{
			Method:    "GET",
			PathQuery: "/ws/2/release?fmt=json&limit=5&query=release%3A%28x%29",
		})
		cases[s.CaseID] = c
	}
	// A committed case that is deliberately not a subject, so a row
	// naming it is reported as an extra subject rather than as an
	// unresolvable id.
	cases["real-boxset-mahler-complete"] = &cassette.Case{ID: "real-boxset-mahler-complete"}
	return cases
}

// fullAlbumInfoLedger builds a ledger that satisfies every check, so
// each subtest below can break exactly one thing and attribute the
// failure to it, matching fullFileTagLedger's role for
// CheckFileTagLedger.
func fullAlbumInfoLedger() *AlbumInfoLedger {
	l := &AlbumInfoLedger{ShimVersion: "1", BeetsVersion: "2.13.1"}
	for _, s := range AlbumInfoSubjects {
		for n := range albumInfoTestLookups[s.CaseID] {
			row := AlbumInfoLedgerRow{
				CaseID:      s.CaseID,
				ReleaseMBID: albumInfoTestMBID(s.CaseID, n),
				Album:       LedgerAlbumInfo{Album: new("Album"), Tracks: []LedgerTrackInfo{{Title: new("One")}}},
			}
			// The pseudo half is the case's first lookup: beets fetches
			// the pseudo-release, then the actual one it links to.
			if n == 0 && slices.Contains(AlbumInfoPseudoSubjects, s.CaseID) {
				row.Merged = &LedgerAlbumInfo{Album: new("Album"), Tracks: []LedgerTrackInfo{{Title: new("One")}}}
			}
			l.Rows = append(l.Rows, row)
		}
	}
	return l
}

func TestCheckAlbumInfoLedger(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*AlbumInfoLedger)
		wantHas string // empty means the ledger is accepted
	}{
		{name: "a complete ledger", mutate: func(*AlbumInfoLedger) {}},
		{
			name:    "no shim version",
			mutate:  func(l *AlbumInfoLedger) { l.ShimVersion = "" },
			wantHas: "shim_version: missing",
		},
		{
			name:    "no beets version",
			mutate:  func(l *AlbumInfoLedger) { l.BeetsVersion = "" },
			wantHas: "beets_version: missing",
		},
		{
			name:    "a row pointing at a case that is not committed",
			mutate:  func(l *AlbumInfoLedger) { l.Rows[0].CaseID = "no-such-case" },
			wantHas: `case "no-such-case" is not a committed corpus case`,
		},
		{
			name: "a subject whose rows were all dropped",
			mutate: func(l *AlbumInfoLedger) {
				l.Rows = dropAlbumInfoRows(l.Rows, "real-nonlatin-hamasaki-ayumi", 5)
			},
			wantHas: `missing rows for subject "real-nonlatin-hamasaki-ayumi"`,
		},
		{
			name: "a row for a committed case that is not a subject",
			mutate: func(l *AlbumInfoLedger) {
				l.Rows[0].CaseID = "real-boxset-mahler-complete"
			},
			wantHas: "real-boxset-mahler-complete: not one of the nineteen named AlbumInfo subjects",
		},
		{
			name: "one of should-reject's six lookups left without a row",
			mutate: func(l *AlbumInfoLedger) {
				l.Rows = dropAlbumInfoRows(l.Rows, "should-reject", 1)
			},
			wantHas: `release lookup "should-reject-release-0" has no row; the derivation skipped it`,
		},
		{
			name:    "a row naming a release the case never looked up",
			mutate:  func(l *AlbumInfoLedger) { l.Rows[0].ReleaseMBID = "not-in-the-log" },
			wantHas: `release "not-in-the-log" is not a release lookup in the case's own request log`,
		},
		{
			name:    "a ledger one row below the floor",
			mutate:  func(l *AlbumInfoLedger) { l.Rows = l.Rows[:len(l.Rows)-1] },
			wantHas: "49 rows, want at least 50",
		},
		{
			name: "a merged block on a case that carries no pseudo-release",
			mutate: func(l *AlbumInfoLedger) {
				l.Rows[0].Merged = &LedgerAlbumInfo{Album: new("Album")}
			},
			wantHas: "clean-match: 1 rows carry a merged block, want none",
		},
		{
			name: "the pseudo case's merged block removed",
			mutate: func(l *AlbumInfoLedger) {
				for i := range l.Rows {
					if l.Rows[i].CaseID == "non-latin-pseudo-release" {
						l.Rows[i].Merged = nil
					}
				}
			},
			wantHas: "non-latin-pseudo-release: 0 rows carry a merged block, want exactly 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := fullAlbumInfoLedger()
			tt.mutate(l)
			err := CheckAlbumInfoLedger(fullAlbumInfoCases(), l)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckAlbumInfoLedger: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckAlbumInfoLedger accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// dropAlbumInfoRows removes up to n of the named case's rows, taking the
// case's first lookups first so the reported gap is a named mbid.
func dropAlbumInfoRows(rows []AlbumInfoLedgerRow, caseID string, n int) []AlbumInfoLedgerRow {
	kept := make([]AlbumInfoLedgerRow, 0, len(rows))
	dropped := 0
	for _, row := range rows {
		if row.CaseID == caseID && dropped < n {
			dropped++
			continue
		}
		kept = append(kept, row)
	}
	return kept
}

// TestCheckAlbumInfoLedgerRejectsASubjectWithoutARationale proves the
// Why requirement fires, which no ledger mutation can reach: the field
// lives on the subject list rather than on a row.
func TestCheckAlbumInfoLedgerRejectsASubjectWithoutARationale(t *testing.T) {
	original := AlbumInfoSubjects
	t.Cleanup(func() { AlbumInfoSubjects = original })

	AlbumInfoSubjects = append([]AlbumInfoSubject(nil), original...)
	AlbumInfoSubjects[0].Why = ""

	err := CheckAlbumInfoLedger(fullAlbumInfoCases(), fullAlbumInfoLedger())
	if err == nil {
		t.Fatal("CheckAlbumInfoLedger accepted a subject carrying no rationale")
	}
	if want := `subject "clean-match" carries no rationale`; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not mention %q", err, want)
	}
}

// TestLoadAlbumInfoLedgerParsesTheTestdataFixture proves
// LoadAlbumInfoLedger reads a well-formed, on-disk document covering all
// fifty rows and that the loaded ledger then satisfies
// CheckAlbumInfoLedger against a matching case set, not only the in-code
// fullAlbumInfoLedger.
func TestLoadAlbumInfoLedgerParsesTheTestdataFixture(t *testing.T) {
	l, err := LoadAlbumInfoLedger(filepath.Join("testdata", "album_info", "good", "album_info.json"))
	if err != nil {
		t.Fatalf("LoadAlbumInfoLedger: %v", err)
	}
	if len(l.Rows) != AlbumInfoLedgerRowFloor {
		t.Fatalf("rows = %d, want %d", len(l.Rows), AlbumInfoLedgerRowFloor)
	}
	if err := CheckAlbumInfoLedger(fullAlbumInfoCases(), l); err != nil {
		t.Errorf("CheckAlbumInfoLedger: %v", err)
	}
}

// TestLoadAlbumInfoLedgerKeepsNullDistinctFromEmpty proves the pointer
// convention survives the load: a JSON null and an empty string decode
// to different values, which is the whole reason the ledger's scalars
// are pointers.
func TestLoadAlbumInfoLedgerKeepsNullDistinctFromEmpty(t *testing.T) {
	l, err := LoadAlbumInfoLedger(filepath.Join("testdata", "album_info", "null-and-empty", "album_info.json"))
	if err != nil {
		t.Fatalf("LoadAlbumInfoLedger: %v", err)
	}
	if got := l.Rows[0].Album.Media; got != nil {
		t.Errorf("media = %q, want a JSON null decoded as nil", *got)
	}
	if got := l.Rows[1].Album.Media; got == nil || *got != "" {
		t.Errorf("media = %v, want a non-nil empty string", got)
	}
}

func TestLoadAlbumInfoLedgerRejectsAnInventedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "album_info.json")
	body := `{"shim_version":"1","beets_version":"2.13.1","rows":[],"extra_key":true}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadAlbumInfoLedger(path); err == nil {
		t.Fatal("LoadAlbumInfoLedger accepted a key nothing parses")
	}
}
