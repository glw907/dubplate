package grade

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
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
		ID: id, Area: "getIndexes", Divergence: "by-design", Sources: []string{"dubplate", "navidrome"},
		Oracle: "o", Dubplate: "d", Evidence: "e", Test: "t", State: StateOwed, Slice: slice,
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
		{name: "owed row with three sources", rows: []Row{
			func() Row {
				r := owedRow("both", "slice 3")
				r.Sources = []string{"dubplate", "navidrome", "gonic"}
				return r
			}(),
		}},
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
			func() Row { r := owedRow("owed-no-divergence", "slice 3"); r.Divergence = ""; return r }(),
			func() Row { r := owedRow("owed-odd-divergence", "slice 3"); r.Divergence = "odd"; return r }(),
			func() Row { r := owedRow("owed-no-test", "slice 3"); r.Test = ""; return r }(),
			func() Row { r := owedRow("owed-with-class", "slice 3"); r.Class = ClassVolatile; return r }(),
			func() Row { r := activeRow("active-no-class", ClassVolatile); r.Class = ""; return r }(),
			func() Row {
				r := owedRow("owed-four-sources", "slice 3")
				r.Sources = []string{"dubplate", "navidrome", "gonic", "melodee"}
				return r
			}(),
			func() Row {
				r := owedRow("owed-no-dubplate", "slice 3")
				r.Sources = []string{"navidrome", "gonic"}
				return r
			}(),
			func() Row {
				r := activeRow("active-three-sources", ClassVolatile)
				r.Sources = []string{"dubplate", "navidrome", "gonic"}
				return r
			}(),
		}, wantIs: ErrInvalidLedger, want: []string{
			`row "good": duplicate id`, `unknown class "weird"`, "active row has an empty pattern",
			"active row names no session", `session "gone"`, "empty version for gonic",
			`uses * inside segment "na*me"`, "contract row names no source",
			`contract source "dubplate" is outside`, "owed row names no slice",
			`source "gonic" on a volatile row`,
			`row "owed-no-divergence": owed row has divergence ""`, `row "owed-odd-divergence": owed row has divergence "odd"`,
			`row "owed-no-test": empty test`, `row "owed-with-class": owed row carries class "volatile"`,
			`row "active-no-class": active row names no class`,
			`row "owed-four-sources": sources`, `row "owed-no-dubplate": owed row sources`,
			`row "active-three-sources": sources`,
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

// recordPath is the dated record of divergences that the committed ledger
// and the assignment's agreements transcribe.
const recordPath = "../../../../docs/superpowers/research/2026-10-06-slice-1-divergence-ledger.md"

// recordOracles maps the record's Oracle column to the oracle sources it
// names; dubplate heads every list.
var recordOracles = map[string][]string{
	"Both":      {dubplate, "navidrome", "gonic"},
	"Navidrome": {dubplate, "navidrome"},
	"gonic":     {dubplate, "gonic"},
}

// retiredRows holds the record rows the committed ledger no longer carries,
// each with the ruling that retired it.
var retiredRows = map[string]string{
	"L-35": "ruling A: the transcoding extension is not declared at 0.8",
	"L-52": "ruling C: no dubplate grade compares the folder view with an oracle",
}

// ruledTexts holds the four cells whose text departs from the
// record. Each has one channel, the ledger file.
var ruledTexts = []struct{ id, column, ruling, want string }{
	{"L-34", "dubplate", "ruling A",
		"The `transcoding` extension is not declared at 0.8; both endpoints answer error 0 as unserved paths"},
	{"L-51", "dubplate", "ruling C",
		"An empty `getIndexes` and error 70 from `getMusicDirectory`, graded against the contract and the spec, never against an oracle"},
	{"L-51", "test", "ruling C",
		"Slice 3: an empty `getIndexes` and error 70 from `getMusicDirectory`, asserted against the spec"},
	{"L-54", "dubplate", "ruling A",
		"The launch subset without `transcoding`: `formPost`, `transcodeOffset`, and `apiKeyAuthentication` from slice 6; the list equals what is implemented at each slice's close"},
}

// untokenedSlices holds the rows whose Test owed cell has no slice token,
// each with the slice it lands in.
var untokenedSlices = map[string]string{
	"L-01": "slice 3", "L-02": "slice 3", "L-03": "slice 3", "L-04": "slice 3", "L-05": "slice 3",
	"L-07": "slice 3", "L-08": "slice 3", "L-12": "slice 3", "L-14": "slice 3", "L-15": "slice 3",
	"L-16": "slice 3", "L-23": "slice 3", "L-27": "slice 3", "L-34": "slice 3", "L-37": "slice 3",
	"L-51": "slice 3", "L-54": "slice 3",
	"L-24": "slice 4", "L-25": "slice 4", "L-26": "slice 4",
	"L-17": "slice 5",
	"A-01": "slice 3", "A-02": "slice 3", "A-04": "slice 3", "A-11": "slice 3", "A-12": "slice 3",
	"A-05": "slice 4",
	"A-09": "slice 5",
}

var sliceToken = regexp.MustCompile(`(?i)\bslice (\d+)`)

// recordSlice computes a record row's slice from its Test owed cell: the
// first slice token, else the list.
func recordSlice(id, test string) (string, error) {
	listed, inList := untokenedSlices[id]
	if m := sliceToken.FindStringSubmatch(test); m != nil {
		if inList {
			return "", fmt.Errorf("%s: listed as untokened but its cell names %q", id, m[0])
		}
		return "slice " + m[1], nil
	}
	if !inList {
		return "", fmt.Errorf("%s: no slice token and not in the list", id)
	}
	return listed, nil
}

// recordCells returns the cells of every record table line whose first
// cell starts with prefix, and the line itself.
func recordCells(t *testing.T, prefix string, width int) (cells [][]string, lines []string) {
	t.Helper()
	for line := range strings.Lines(string(mustRead(t, recordPath))) {
		line = strings.TrimSuffix(line, "\n")
		if !strings.HasPrefix(line, "| "+prefix) {
			continue
		}
		cs := strings.Split(strings.TrimSuffix(strings.TrimPrefix(line, "| "), " |"), " | ")
		if len(cs) != width {
			t.Fatalf("record line has %d cells, want %d: %.60s", len(cs), width, line)
		}
		cells, lines = append(cells, cs), append(lines, line)
	}
	return cells, lines
}

func tableLine(cells ...string) string { return "| " + strings.Join(cells, " | ") + " |" }

// parseRecordRow reads the record's ledger columns (id, area, divergence,
// oracle, oracle behavior, dubplate behavior, evidence, test owed) into an
// owed row, leaving its slice to the caller.
func parseRecordRow(t *testing.T, c []string) Row {
	t.Helper()
	sources, ok := recordOracles[c[3]]
	if !ok {
		t.Fatalf("%s: unknown oracle cell %q", c[0], c[3])
	}
	return Row{
		ID: c[0], Area: c[1], Divergence: c[2], Sources: sources,
		Oracle: c[4], Dubplate: c[5], Evidence: c[6], Test: c[7], State: StateOwed,
	}
}

// renderRecordRow writes r back as the record's table line.
func renderRecordRow(r Row) string {
	var oracle string
	switch others := slices.DeleteFunc(slices.Clone(r.Sources), func(s string) bool { return s == dubplate }); len(others) {
	case 2:
		oracle = "Both"
	case 1:
		oracle = map[string]string{"navidrome": "Navidrome", "gonic": "gonic"}[others[0]]
	}
	return tableLine(r.ID, r.Area, r.Divergence, oracle, r.Oracle, r.Dubplate, r.Evidence, r.Test)
}

func parseAgreement(c []string) agreementEntry {
	return agreementEntry{ID: c[0], Area: c[1], Behavior: c[2], Evidence: c[3], Test: c[4]}
}

func renderAgreement(a agreementEntry) string {
	return tableLine(a.ID, a.Area, a.Behavior, a.Evidence, a.Test)
}

func recordRows(t *testing.T) []Row {
	t.Helper()
	cells, _ := recordCells(t, "L-", 8)
	rows := make([]Row, len(cells))
	for i, c := range cells {
		rows[i] = parseRecordRow(t, c)
	}
	return rows
}

func TestRecordTablesRoundTrip(t *testing.T) {
	cells, lines := recordCells(t, "L-", 8)
	if len(lines) != 54 {
		t.Fatalf("parsed %d divergence rows, want 54", len(lines))
	}
	for i, c := range cells {
		if got := renderRecordRow(parseRecordRow(t, c)); got != lines[i] {
			t.Errorf("%s does not round trip:\n got %s\nwant %s", c[0], got, lines[i])
		}
	}
	cells, lines = recordCells(t, "A-", 5)
	if len(lines) != 12 {
		t.Fatalf("parsed %d agreement rows, want 12", len(lines))
	}
	for i, c := range cells {
		if got := renderAgreement(parseAgreement(c)); got != lines[i] {
			t.Errorf("%s does not round trip:\n got %s\nwant %s", c[0], got, lines[i])
		}
	}
}

// expectedRow is the committed form of a record row: the slice computed,
// the ruled texts in place of the record's.
func expectedRow(rec Row) (Row, error) {
	slice, err := recordSlice(rec.ID, rec.Test)
	if err != nil {
		return Row{}, err
	}
	rec.Slice = slice
	for _, rt := range ruledTexts {
		if rt.id != rec.ID {
			continue
		}
		switch rt.column {
		case "dubplate":
			rec.Dubplate = rt.want
		case "test":
			rec.Test = rt.want
		}
	}
	return rec, nil
}

// transcriptionProblems checks the ledger rows against the record rows:
// every record id is a ledger row (by its id, or as id/oracle) or retired,
// an owed row equals the record's expected form, and an active row names
// sources the record row names.
func transcriptionProblems(rows, record []Row, retired map[string]string) []string {
	var ps []string
	byBase := map[string][]Row{}
	for _, r := range rows {
		base, _, _ := strings.Cut(r.ID, "/")
		byBase[base] = append(byBase[base], r)
		if _, gone := retired[base]; gone {
			ps = append(ps, fmt.Sprintf("%s: carried though retired", r.ID))
		}
	}
	for _, rec := range record {
		got := byBase[rec.ID]
		if _, gone := retired[rec.ID]; gone {
			continue
		}
		if len(got) == 0 {
			ps = append(ps, fmt.Sprintf("%s: neither committed nor retired", rec.ID))
			continue
		}
		want, err := expectedRow(rec)
		if err != nil {
			ps = append(ps, err.Error())
			continue
		}
		for _, r := range got {
			if r.State == StateActive {
				for _, src := range r.Sources {
					if !slices.Contains(rec.Sources, src) {
						ps = append(ps, fmt.Sprintf("%s: source %s is outside the record's %v", r.ID, src, rec.Sources))
					}
				}
				continue
			}
			for _, col := range []struct {
				name      string
				got, want any
			}{
				{"area", r.Area, want.Area}, {"divergence", r.Divergence, want.Divergence},
				{"sources", r.Sources, want.Sources}, {"oracle", r.Oracle, want.Oracle},
				{"dubplate", r.Dubplate, want.Dubplate}, {"evidence", r.Evidence, want.Evidence},
				{"test", r.Test, want.Test}, {"slice", r.Slice, want.Slice},
			} {
				if fmt.Sprint(col.got) != fmt.Sprint(col.want) {
					ps = append(ps, fmt.Sprintf("%s: %s = %v, want %v", r.ID, col.name, col.got, col.want))
				}
			}
		}
	}
	return ps
}

// transcribe returns the committed form of every record row not retired.
func transcribe(t *testing.T, record []Row) []Row {
	t.Helper()
	var rows []Row
	for _, rec := range record {
		if _, gone := retiredRows[rec.ID]; gone {
			continue
		}
		r, err := expectedRow(rec)
		if err != nil {
			t.Fatal(err)
		}
		rows = append(rows, r)
	}
	return rows
}

func TestTranscriptionChecks(t *testing.T) {
	record := recordRows(t)
	if len(record) != 54 {
		t.Fatalf("record has %d rows, want 54", len(record))
	}
	tests := []struct {
		name    string
		edit    func(rows []Row) []Row
		retired map[string]string
		want    []string // fragments; nil means no problem
	}{
		{name: "faithful transcription"},
		{name: "retired list omits L-35", retired: map[string]string{"L-52": "c"}, want: []string{"L-35: neither committed nor retired"}},
		{name: "retired row carried", edit: func(rows []Row) []Row { return append(rows, owedRow("L-35", "slice 5")) },
			want: []string{"L-35: carried though retired"}},
		{name: "row dropped", edit: func(rows []Row) []Row { return rows[1:] }, want: []string{"L-01: neither committed nor retired"}},
		{name: "Both mapped to one oracle", edit: func(rows []Row) []Row {
			rows[0].Sources = []string{dubplate, "navidrome"}
			return rows
		}, want: []string{"L-01: sources"}},
		{name: "test column dropped", edit: func(rows []Row) []Row { rows[1].Test = ""; return rows }, want: []string{"L-02: test"}},
		{name: "slice moved", edit: func(rows []Row) []Row { rows[0].Slice = "slice 4"; return rows },
			want: []string{"L-01: slice = slice 4, want slice 3"}},
		{name: "ruled text reverted", edit: func(rows []Row) []Row {
			for i, r := range rows {
				if r.ID == "L-51" {
					rows[i].Dubplate = record[50].Dubplate
				}
			}
			return rows
		}, want: []string{"L-51: dubplate"}},
		{name: "active pair from the row's sources", edit: func(rows []Row) []Row {
			r := activeRow("L-01/navidrome", ClassContract)
			r.Sources, r.Source = []string{dubplate, "navidrome"}, "navidrome"
			return append(rows, r)
		}},
		{name: "active pair against the other oracle", edit: func(rows []Row) []Row {
			r := activeRow("L-07/navidrome", ClassContract) // L-07 names gonic alone
			r.Sources, r.Source = []string{dubplate, "navidrome"}, "navidrome"
			return append(rows, r)
		}, want: []string{"L-07/navidrome: source navidrome is outside"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := transcribe(t, record)
			if tt.edit != nil {
				rows = tt.edit(rows)
			}
			retired := retiredRows
			if tt.retired != nil {
				retired = tt.retired
			}
			got := transcriptionProblems(rows, record, retired)
			if tt.want == nil {
				if len(got) != 0 {
					t.Fatalf("problems = %v, want none", got)
				}
				return
			}
			joined := strings.Join(got, "\n")
			for _, w := range tt.want {
				if !strings.Contains(joined, w) {
					t.Errorf("problems lack %q:\n%s", w, joined)
				}
			}
		})
	}
}

func TestRuledTextsDifferFromTheRecord(t *testing.T) {
	record := recordRows(t)
	for _, rt := range ruledTexts {
		i := slices.IndexFunc(record, func(r Row) bool { return r.ID == rt.id })
		if i < 0 {
			t.Fatalf("%s is not in the record", rt.id)
		}
		recorded := map[string]string{"dubplate": record[i].Dubplate, "test": record[i].Test}[rt.column]
		if recorded == "" || recorded == rt.want {
			t.Errorf("%s %s (%s): record text %q does not differ from the ruled text", rt.id, rt.column, rt.ruling, recorded)
		}
	}
}

func TestCommittedLedger(t *testing.T) {
	l, err := LoadLedger("../data/ledger.json", "../data/sessions")
	if err != nil {
		t.Fatal(err)
	}
	record := recordRows(t)
	if got, want := len(l.Rows), len(record)-len(retiredRows); got != want {
		t.Errorf("ledger has %d rows, want %d", got, want)
	}
	for _, r := range l.Rows {
		if !slices.Contains(r.Sources, dubplate) {
			t.Errorf("row %q does not name dubplate as a source: %v", r.ID, r.Sources)
		}
		if r.Divergence == "" || r.Test == "" {
			t.Errorf("row %q lacks a divergence or a test", r.ID)
		}
	}
	for _, p := range transcriptionProblems(l.Rows, record, retiredRows) {
		t.Error(p)
	}
	wantOwed := map[string]int{}
	for _, rec := range record {
		if _, gone := retiredRows[rec.ID]; gone {
			continue
		}
		want, err := expectedRow(rec)
		if err != nil {
			t.Fatal(err)
		}
		wantOwed[want.Slice]++
	}
	gotOwed := map[string]int{}
	for _, r := range l.Rows {
		if r.State == StateOwed {
			gotOwed[r.Slice]++
		}
	}
	if !maps.Equal(gotOwed, wantOwed) {
		t.Errorf("owed rows by slice = %v, want %v", gotOwed, wantOwed)
	}
}

func TestCommittedAgreements(t *testing.T) {
	cells, _ := recordCells(t, "A-", 5)
	got := committedAssign(t).Agreements
	if len(got) != len(cells) {
		t.Fatalf("assignment holds %d agreements, record has %d", len(got), len(cells))
	}
	for i, c := range cells {
		want := parseAgreement(c)
		slice, err := recordSlice(want.ID, want.Test)
		if err != nil {
			t.Fatal(err)
		}
		want.Slice = slice
		if got[i] != want {
			t.Errorf("agreement %s = %+v, want %+v", want.ID, got[i], want)
		}
	}
}
