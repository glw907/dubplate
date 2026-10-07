package grade

import (
	"encoding/json/v2"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const (
	navidromeVersion = "0.64.2 (10114574)"
	gonicVersion     = "0.22.0"
)

var pairNG = [2]string{"navidrome", "gonic"}

// writeLedger writes rows to a ledger file and a sessions directory
// holding the named session directories, and returns both paths.
func writeLedger(t *testing.T, rows []Row, sessions ...string) (ledger, sessionsDir string) {
	t.Helper()
	dir := t.TempDir()
	sessionsDir = filepath.Join(dir, "sessions")
	for _, s := range sessions {
		if err := os.MkdirAll(filepath.Join(sessionsDir, s), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	raw, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	ledger = filepath.Join(dir, "ledger.json")
	if err := os.WriteFile(ledger, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	return ledger, sessionsDir
}

// activeRow returns a valid active row for the navidrome and gonic pair on
// the session "s1" at getAlbum, /album/name.
func activeRow(id string, class Class) Row {
	r := Row{
		ID: id, Area: "getAlbum", Class: class, Sources: []string{"navidrome", "gonic"},
		Oracle: "o", Dubplate: "d", Evidence: "e", State: StateActive,
		Pattern:  Pattern{Endpoint: "getAlbum", Format: "json", Path: "album/name"},
		Session:  "s1",
		Versions: map[string]string{"navidrome": navidromeVersion, "gonic": gonicVersion},
	}
	if class == ClassContract {
		r.Source = "gonic"
	}
	return r
}

func owedRow(id, slice string) Row {
	return Row{
		ID: id, Area: "getIndexes", Class: ClassByConstruction, Sources: []string{"dubplate", "navidrome"},
		Oracle: "o", Dubplate: "d", Evidence: "e", State: StateOwed, Slice: slice,
	}
}

func testSession() *session.Session {
	return &session.Session{Sources: map[string]session.Source{
		"navidrome": {ServerVersion: navidromeVersion},
		"gonic":     {ServerVersion: gonicVersion},
	}}
}

func finding(kind Kind, path string, sources ...string) Finding {
	if sources == nil {
		sources = []string{"navidrome", "gonic"}
	}
	return Finding{
		Request: "getAlbum.json", Endpoint: "getAlbum", Format: client.JSON,
		Sources: sources, Kind: kind, Path: path, Values: make([]string, len(sources)),
	}
}

var allKinds = []Kind{
	KindValue, KindType, KindPresence, KindOrder, KindIdentity,
	KindMissingSource, KindParse, KindContract,
}

func TestLoadLedgerStates(t *testing.T) {
	good := activeRow("good", ClassVolatile)
	tests := []struct {
		name    string
		rows    []Row
		raw     string // when set, written instead of rows
		missing bool
		want    []string // every fragment the error must contain; nil means success
		wantIs  error
	}{
		{name: "empty ledger is valid", rows: []Row{}},
		{name: "valid mixed ledger", rows: []Row{good, owedRow("owed", "slice 3")}},
		{name: "absent ledger", missing: true, wantIs: fs.ErrNotExist, want: []string{"read ledger"}},
		{name: "malformed", raw: `[{"id": `, want: []string{"decode ledger"}},
		{name: "unknown member", raw: `[{"id":"x","bogus":1}]`, want: []string{"decode ledger", "bogus"}},
		{name: "many bad entries all reported", rows: []Row{
			good,
			good, // duplicate id
			func() Row { r := activeRow("unknown-class", "weird"); return r }(),
			func() Row { r := activeRow("empty-pattern", ClassVolatile); r.Pattern = Pattern{}; return r }(),
			func() Row { r := activeRow("no-session", ClassVolatile); r.Session = ""; return r }(),
			func() Row { r := activeRow("absent-session", ClassVolatile); r.Session = "gone"; return r }(),
			func() Row { r := activeRow("empty-version", ClassVolatile); r.Versions["gonic"] = ""; return r }(),
			func() Row { r := activeRow("bad-star", ClassVolatile); r.Pattern.Path = "album/na*me"; return r }(),
			func() Row { r := activeRow("contract-no-source", ClassContract); r.Source = ""; return r }(),
			func() Row { r := activeRow("contract-outside", ClassContract); r.Source = "dubplate"; return r }(),
			func() Row { r := owedRow("owed-no-slice", ""); return r }(),
			func() Row { r := activeRow("source-on-volatile", ClassVolatile); r.Source = "gonic"; return r }(),
		}, wantIs: ErrInvalidLedger, want: []string{
			`row "good": duplicate id`, `unknown class "weird"`, "active row has an empty pattern",
			"active row names no session", `session "gone"`, "empty version for gonic",
			`uses * inside segment "na*me"`, "contract row names no source",
			`contract source "dubplate" is outside`, "owed row names no slice",
			`source "gonic" on a volatile row`,
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, sessions := writeLedger(t, tt.rows, "s1")
			switch {
			case tt.raw != "":
				if err := os.WriteFile(path, []byte(tt.raw), 0o644); err != nil {
					t.Fatal(err)
				}
			case tt.missing:
				path = filepath.Join(t.TempDir(), "nope.json")
			}
			l, err := LoadLedger(path, sessions)
			if tt.want == nil {
				if err != nil {
					t.Fatalf("LoadLedger: %v", err)
				}
				if len(l.Rows) != len(tt.rows) {
					t.Fatalf("loaded %d rows, want %d", len(l.Rows), len(tt.rows))
				}
				return
			}
			if err == nil {
				t.Fatal("LoadLedger succeeded")
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("error %v is not %v", err, tt.wantIs)
			}
			for _, frag := range tt.want {
				if !strings.Contains(err.Error(), frag) {
					t.Errorf("error lacks %q:\n%v", frag, err)
				}
			}
		})
	}
}

func TestApplyEmptyLedgerLeavesEveryFindingUnledgered(t *testing.T) {
	path, sessions := writeLedger(t, []Row{}, "s1")
	l, err := LoadLedger(path, sessions)
	if err != nil {
		t.Fatal(err)
	}
	var found []Finding
	for _, k := range allKinds {
		found = append(found, finding(k, "album/name"))
	}
	res, err := l.Apply(testSession(), "s1", pairNG, found)
	if err != nil {
		t.Fatal(err)
	}
	for i, m := range res.Matches {
		if m != "" {
			t.Errorf("finding %s matched %q on an empty ledger", found[i].Kind, m)
		}
	}
}

// TestApplyClassSuppression applies one row per class and one finding per
// kind at the row's path, so each cell of the class-by-kind table is a
// distinct finding the row either takes or leaves.
func TestApplyClassSuppression(t *testing.T) {
	suppressed := map[Class][]Kind{
		ClassByConstruction: {KindValue, KindType, KindPresence, KindOrder, KindIdentity},
		ClassUnoracled:      {KindValue, KindType, KindPresence, KindOrder, KindIdentity},
		ClassVolatile:       {KindValue},
		ClassContract:       {KindContract},
	}
	for class, kinds := range suppressed {
		for _, kind := range allKinds {
			t.Run(string(class)+"/"+string(kind), func(t *testing.T) {
				l := &Ledger{Rows: []Row{activeRow("r", class)}}
				f := finding(kind, "album/name")
				if kind == KindContract {
					f = finding(kind, "album/name", "gonic")
				}
				res, err := l.Apply(testSession(), "s1", pairNG, []Finding{f})
				want := slices.Contains(kinds, kind)
				if got := res.Matches[0] == "r"; got != want {
					t.Errorf("suppressed = %t, want %t", got, want)
				}
				if want != (err == nil) {
					t.Errorf("err = %v, want stale only when the row matched nothing", err)
				}
			})
		}
	}
}

func TestApplyOwedRowSuppressesNothing(t *testing.T) {
	owed := owedRow("owed", "slice 3")
	owed.Pattern = Pattern{Endpoint: "getAlbum", Format: "json", Path: "album/name"}
	// Everything but its state would make the row apply and match.
	owed.Session = "s1"
	owed.Versions = map[string]string{"navidrome": navidromeVersion}
	l := &Ledger{Rows: []Row{owed, owedRow("owed2", "slice 3"), owedRow("owed3", "slice 6")}}
	res, err := l.Apply(testSession(), "s1", [2]string{"dubplate", "navidrome"},
		[]Finding{finding(KindValue, "album/name", "dubplate", "navidrome")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Matches[0] != "" {
		t.Errorf("owed row matched: %q", res.Matches[0])
	}
	if _, applied := res.Counts["owed"]; applied {
		t.Error("owed row was applied")
	}
	if got := res.Owed; got["slice 3"] != 2 || got["slice 6"] != 1 || len(got) != 2 {
		t.Errorf("Owed = %v, want slice 3 twice and slice 6 once", got)
	}
}

func TestApplyRowMatchesOnlyItsPair(t *testing.T) {
	tests := []struct {
		name    string
		pair    [2]string
		session string
		finding Finding
		rows    []Row
		matched bool
	}{
		{"same pair either order", [2]string{"gonic", "navidrome"}, "s1", finding(KindValue, "album/name"), []Row{activeRow("r", ClassVolatile)}, true},
		{"other pair at the same path", [2]string{"dubplate", "gonic"}, "s1", finding(KindValue, "album/name", "dubplate", "gonic"), []Row{activeRow("r", ClassVolatile)}, false},
		{"other session", pairNG, "s2", finding(KindValue, "album/name"), []Row{activeRow("r", ClassVolatile)}, false},
		{"contract row on an oracle leaves dubplate's finding", [2]string{"dubplate", "gonic"}, "s1",
			finding(KindContract, "album/name", "dubplate"), []Row{activeRow("r", ClassContract)}, false},
		{"contract row for the same source on the same pair", pairNG, "s1",
			finding(KindContract, "album/name", "gonic"), []Row{activeRow("r", ClassContract)}, true},
		{"contract row leaves the other source's finding", pairNG, "s1",
			finding(KindContract, "album/name", "navidrome"), []Row{activeRow("r", ClassContract)}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &Ledger{Rows: tt.rows}
			sess := testSession()
			sess.Sources["dubplate"] = session.Source{ServerVersion: "x"}
			res, _ := l.Apply(sess, tt.session, tt.pair, []Finding{tt.finding})
			if got := res.Matches[0] == "r"; got != tt.matched {
				t.Errorf("matched = %t, want %t", got, tt.matched)
			}
		})
	}
}

func TestApplyStaleRow(t *testing.T) {
	l := &Ledger{Rows: []Row{activeRow("live", ClassVolatile), func() Row {
		r := activeRow("dead", ClassVolatile)
		r.Pattern.Path = "album/year"
		return r
	}()}}
	// A non-empty findings list that one row matches and the other does not.
	res, err := l.Apply(testSession(), "s1", pairNG, []Finding{finding(KindValue, "album/name")})
	if !errors.Is(err, ErrStaleRow) || !strings.Contains(err.Error(), `"dead"`) || strings.Contains(err.Error(), `"live"`) {
		t.Fatalf("error = %v, want a stale error naming only dead", err)
	}
	if res.Counts["live"] != 1 || res.Counts["dead"] != 0 {
		t.Errorf("Counts = %v", res.Counts)
	}
	// A row on another session is not stale for this one.
	other := activeRow("elsewhere", ClassVolatile)
	other.Session = "s2"
	l = &Ledger{Rows: []Row{activeRow("live", ClassVolatile), other}}
	if _, err := l.Apply(testSession(), "s1", pairNG, []Finding{finding(KindValue, "album/name")}); err != nil {
		t.Errorf("row on another session reported: %v", err)
	}
}

func TestApplyVersionPrivileged(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Row)
		sess   func(*session.Session)
		bad    bool
	}{
		{"equal", func(*Row) {}, func(*session.Session) {}, false},
		{"ledger version is a prefix of the session's", func(r *Row) { r.Versions["navidrome"] = "0.64.2" }, func(*session.Session) {}, true},
		{"session version is a prefix of the ledger's", func(*Row) {}, func(s *session.Session) { s.Sources["gonic"] = session.Source{ServerVersion: "0.22"} }, true},
		{"session recorded no version", func(*Row) {}, func(s *session.Session) { delete(s.Sources, "navidrome") }, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := activeRow("r", ClassVolatile)
			tt.mutate(&row)
			sess := testSession()
			tt.sess(sess)
			_, err := (&Ledger{Rows: []Row{row}}).Apply(sess, "s1", pairNG, []Finding{finding(KindValue, "album/name")})
			if got := errors.Is(err, ErrVersionMismatch); got != tt.bad {
				t.Errorf("mismatch = %t (err %v), want %t", got, err, tt.bad)
			}
		})
	}
}

func TestLoadLedgerRefusesEmptyVersionBeforeApply(t *testing.T) {
	row := activeRow("r", ClassVolatile)
	row.Versions["navidrome"] = ""
	path, sessions := writeLedger(t, []Row{row}, "s1")
	if _, err := LoadLedger(path, sessions); err == nil || !strings.Contains(err.Error(), "empty version for navidrome") {
		t.Errorf("error = %v", err)
	}
}

func TestPatternWildcard(t *testing.T) {
	tests := []struct {
		pattern, path  string
		exact, subtree bool
	}{
		{"album/song/*/title", "album/song/3/title", true, true},
		{"album/song/*/title", "album/song/12/title", true, true},
		{"album/song/*/title", "album/song/extra/title", false, false}, // a named segment
		{"album/song/*/title", "album/song//title", false, false},
		{"album/song/*/title", "album/song/3", false, false},
		{"album/song/*", "album/song/3/title", false, true}, // beneath the pattern
		{"album/song", "album/song/3/title", false, true},
		{"album/song", "album/songs", false, false}, // segment boundary
		{"album/song", "album", false, false},
		{"album/song/3", "album/song/3", true, true},
	}
	for _, tt := range tests {
		if got := pathCovers(tt.pattern, tt.path, false); got != tt.exact {
			t.Errorf("pathCovers(%q, %q, exact) = %t, want %t", tt.pattern, tt.path, got, tt.exact)
		}
		if got := pathCovers(tt.pattern, tt.path, true); got != tt.subtree {
			t.Errorf("pathCovers(%q, %q, subtree) = %t, want %t", tt.pattern, tt.path, got, tt.subtree)
		}
	}
}

// TestRowScopeByClass writes every class's row at a parent path and grades
// one finding of a kind the class suppresses at a child path. Only the
// surface classes reach beneath their path, so a volatile or contract row
// written one level too high leaves the child difference reported.
func TestRowScopeByClass(t *testing.T) {
	tests := []struct {
		class      Class
		kind       Kind
		sources    []string
		childTaken bool
	}{
		{ClassByConstruction, KindValue, nil, true},
		{ClassUnoracled, KindPresence, nil, true},
		{ClassVolatile, KindValue, nil, false},
		{ClassContract, KindContract, []string{"gonic"}, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.class), func(t *testing.T) {
			row := activeRow("parent", tt.class)
			row.Pattern.Path = "album"
			at := func(path string) Finding { return finding(tt.kind, path, tt.sources...) }
			res, err := (&Ledger{Rows: []Row{row}}).Apply(testSession(), "s1", pairNG, []Finding{at("album"), at("album/name")})
			if err != nil {
				t.Fatal(err)
			}
			if res.Matches[0] != "parent" {
				t.Errorf("finding at the row's own path unledgered")
			}
			if got := res.Matches[1] == "parent"; got != tt.childTaken {
				t.Errorf("child finding taken = %t, want %t", got, tt.childTaken)
			}
		})
	}
}

func TestPatternFormatAndEndpoint(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Row)
		f      Finding
		want   bool
	}{
		{"any covers xml", func(r *Row) { r.Pattern.Format = "any" }, func() Finding { f := finding(KindValue, "album/name"); f.Format = client.XML; return f }(), true},
		{"json row skips xml", func(*Row) {}, func() Finding { f := finding(KindValue, "album/name"); f.Format = client.XML; return f }(), false},
		{"other endpoint", func(*Row) {}, func() Finding { f := finding(KindValue, "album/name"); f.Endpoint = "getSong"; return f }(), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := activeRow("r", ClassVolatile)
			tt.mutate(&row)
			res, _ := (&Ledger{Rows: []Row{row}}).Apply(testSession(), "s1", pairNG, []Finding{tt.f})
			if got := res.Matches[0] == "r"; got != tt.want {
				t.Errorf("matched = %t, want %t", got, tt.want)
			}
		})
	}
}

func TestApplyCountsEveryOverlappingRow(t *testing.T) {
	broad := activeRow("broad", ClassByConstruction)
	broad.Pattern.Path = "album"
	l := &Ledger{Rows: []Row{broad, activeRow("narrow", ClassVolatile)}}
	res, err := l.Apply(testSession(), "s1", pairNG, []Finding{finding(KindValue, "album/name"), finding(KindValue, "album/year")})
	if err != nil {
		t.Fatal(err)
	}
	if res.Counts["broad"] != 2 || res.Counts["narrow"] != 1 {
		t.Errorf("Counts = %v, want broad 2 and narrow 1", res.Counts)
	}
	if res.Matches[0] != "broad" || res.Matches[1] != "broad" {
		t.Errorf("Matches = %v, want the first row in ledger order", res.Matches)
	}
}

// TestEverySessionAnActiveRowNamesIsStaleChecked grades each session the
// ledger's active rows name, found from the rows themselves, so a new
// session joins the check without an edit here.
func TestEverySessionAnActiveRowNamesIsStaleChecked(t *testing.T) {
	r1, r2 := activeRow("a", ClassVolatile), activeRow("b", ClassVolatile)
	r2.Session = "s2"
	path, sessions := writeLedger(t, []Row{r1, r2}, "s1", "s2")
	l, err := LoadLedger(path, sessions)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, r := range l.Rows {
		if r.State == StateActive && !slices.Contains(names, r.Session) {
			names = append(names, r.Session)
		}
	}
	if !slices.Equal(names, []string{"s1", "s2"}) {
		t.Fatalf("sessions = %v", names)
	}
	for _, name := range names {
		// One finding that no row covers: every row on the session is stale.
		_, err := l.Apply(testSession(), name, pairNG, []Finding{finding(KindValue, "album/other")})
		if !errors.Is(err, ErrStaleRow) {
			t.Errorf("session %s: error = %v, want a stale row", name, err)
		}
	}
}

func TestCommittedLedger(t *testing.T) {
	l, err := LoadLedger("../data/ledger.json", "../data/sessions")
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Rows) == 0 {
		t.Fatal("committed ledger is empty")
	}
	owed := map[string]int{}
	for _, r := range l.Rows {
		if !slices.Contains(r.Sources, dubplate) {
			t.Errorf("row %q does not name dubplate as a source: %v", r.ID, r.Sources)
		}
		if r.State == StateOwed {
			owed[r.Slice]++
		}
	}
	if owed["slice 3"] != 3 || owed["slice 6"] != 1 || len(owed) != 2 {
		t.Errorf("owed rows by slice = %v, want slice 3 three and slice 6 one", owed)
	}
}
