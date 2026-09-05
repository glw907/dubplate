package matcher

import (
	"encoding/json"
	"math"
	"math/rand/v2"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// assignmentCost sums the cells a column-indexed solver result selects.
func assignmentCost(costs [][]float64, assigned []int) float64 {
	var total float64
	for col, row := range assigned {
		if row >= 0 {
			total += costs[row][col]
		}
	}
	return total
}

// bruteForceOptimum returns the cheapest full assignment of a square
// matrix by enumerating every permutation, which is tractable only for
// the small n the property test uses.
func bruteForceOptimum(costs [][]float64) float64 {
	n := len(costs)
	used := make([]bool, n)
	best := math.Inf(1)
	var walk func(row int, total float64)
	walk = func(row int, total float64) {
		if row == n {
			best = min(best, total)
			return
		}
		for col := range n {
			if used[col] {
				continue
			}
			used[col] = true
			walk(row+1, total+costs[row][col])
			used[col] = false
		}
	}
	walk(0, 0)
	return best
}

func TestSolveLAPFindsAnOptimum(t *testing.T) {
	tests := []struct {
		name      string
		costs     [][]float64
		wantCost  float64
		wantPairs int // min(rows, cols)
	}{
		{"square, unique optimum", [][]float64{{0, 1}, {1, 0}}, 0, 2},
		{"square, swap is optimal", [][]float64{{1, 0}, {0, 1}}, 0, 2},
		{"more rows than columns", [][]float64{{0, 0.9}, {0.9, 0}, {0.5, 0.5}}, 0, 2},
		{"more columns than rows", [][]float64{{10, 0, 5}, {0, 10, 5}}, 0, 2},
		{"3x3 with a forced detour", [][]float64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}, 5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := solveLAP(tt.costs)
			rows := len(tt.costs)
			cols := len(tt.costs[0])
			if len(got) != cols {
				t.Fatalf("solveLAP returned %d entries, want one per column (%d)", len(got), cols)
			}
			matched := 0
			seen := make(map[int]bool)
			var total float64
			for col, row := range got {
				if row == -1 {
					continue
				}
				if row < 0 || row >= rows {
					t.Fatalf("column %d assigned row %d, out of range [0,%d)", col, row, rows)
				}
				if seen[row] {
					t.Fatalf("row %d assigned to two columns", row)
				}
				seen[row] = true
				matched++
				total += tt.costs[row][col]
			}
			if matched != tt.wantPairs {
				t.Errorf("matched %d pairs, want min(rows, cols) = %d", matched, tt.wantPairs)
			}
			if total != tt.wantCost {
				t.Errorf("assignment cost = %v, want the optimum %v", total, tt.wantCost)
			}
		})
	}
}

func TestSolveLAPMatchesBruteForceOnSmallMatrices(t *testing.T) {
	// Deterministic pseudo-random matrices with small integer costs, so
	// ties are common and an optimum is cheap to enumerate. The
	// assertion is on COST, never on which optimum is chosen: the tie
	// choice is Task 2's, and asserting it here would make this test
	// fail for the right solver and the wrong reason.
	rng := rand.New(rand.NewPCG(1, 2))
	for n := 2; n <= 6; n++ {
		for range 200 {
			costs := make([][]float64, n)
			for i := range costs {
				costs[i] = make([]float64, n)
				for j := range costs[i] {
					costs[i][j] = float64(rng.IntN(4))
				}
			}
			got := assignmentCost(costs, solveLAP(costs))
			want := bruteForceOptimum(costs)
			if got != want {
				t.Fatalf("n=%d: solveLAP cost %v, brute force optimum %v, matrix %v", n, got, want, costs)
			}
		}
	}
}

func TestAssignItemsOrdersPairsByTrackIndex(t *testing.T) {
	// Three items whose optimum is the scramble {item 1 -> track 0,
	// item 2 -> track 1, item 0 -> track 3}: the shape
	// complete-combined-over-scrambled carries. Pairs must come back
	// [(1,0), (2,1), (0,3)], ascending by Track.
	items := []ItemFields{
		{Title: "delta", Track: 1, Length: 100},
		{Title: "alpha", Track: 2, Length: 100},
		{Title: "bravo", Track: 3, Length: 100},
	}
	tracks := []TrackInfo{
		{Title: "alpha", Index: 1, Length: 100},
		{Title: "bravo", Index: 2, Length: 100},
		{Title: "charlie", Index: 3, Length: 100},
		{Title: "delta", Index: 4, Length: 100},
	}
	got := AssignItems(DefaultTunables(), items, tracks)
	want := []Pair{{Item: 1, Track: 0}, {Item: 2, Track: 1}, {Item: 0, Track: 3}}
	if !slices.Equal(got.Pairs, want) {
		t.Errorf("Pairs = %v, want %v (ascending track index, match.py:72-76)", got.Pairs, want)
	}
	if len(got.ExtraItems) != 0 {
		t.Errorf("ExtraItems = %v, want none: three items, four tracks", got.ExtraItems)
	}
	if !slices.Equal(got.ExtraTracks, []int{2}) {
		t.Errorf("ExtraTracks = %v, want [2]", got.ExtraTracks)
	}
}

func TestAssignItemsSortsExtraItemsByDiscTrackTitle(t *testing.T) {
	// Four items, one track. Three items fall out, and their recorded
	// order is (disc, track, title), not input order and not solver
	// order.
	items := []ItemFields{
		{Title: "zulu", Disc: 2, Track: 1, Length: 100},
		{Title: "alpha", Disc: 1, Track: 9, Length: 100},
		{Title: "bravo", Disc: 1, Track: 2, Length: 100},
		{Title: "match", Disc: 1, Track: 1, Length: 100},
	}
	tracks := []TrackInfo{{Title: "match", Index: 1, Length: 100}}
	got := AssignItems(DefaultTunables(), items, tracks)
	if !slices.Equal(got.ExtraItems, []int{2, 1, 0}) {
		t.Errorf("ExtraItems = %v, want [2 1 0]: (1,2,bravo) < (1,9,alpha) < (2,1,zulu)", got.ExtraItems)
	}
}

func TestAssignItemsSortsExtraTracksByIndexTitle(t *testing.T) {
	// One item, four tracks. The three leftover tracks come back in
	// (index, title) order, which here inverts the candidate's own
	// list order.
	items := []ItemFields{{Title: "match", Track: 1, Length: 100}}
	tracks := []TrackInfo{
		{Title: "match", Index: 1, Length: 100},
		{Title: "zulu", Index: 9, Length: 100},
		{Title: "alpha", Index: 4, Length: 100},
		{Title: "bravo", Index: 2, Length: 100},
	}
	got := AssignItems(DefaultTunables(), items, tracks)
	if !slices.Equal(got.ExtraTracks, []int{3, 2, 1}) {
		t.Errorf("ExtraTracks = %v, want [3 2 1]: index 2 < 4 < 9", got.ExtraTracks)
	}
}

func TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease(t *testing.T) {
	// Two items and two tracks whose titles and lengths are symmetric,
	// so the only components that can move the optimum are the track
	// index and the track artist. Under inclArtist FALSE the index
	// component (distance.py:411-413, weight 1.0) makes the IDENTITY
	// the unique optimum, cost 0.0 against the swap's 0.3333. Under
	// inclArtist TRUE the artist component (weight 2.0,
	// tunables.go:272-273) makes the SWAP the unique optimum, 0.25
	// against the identity's 0.4444. So the fixture discriminates the
	// two objectives outright and depends on lapjv's tie rule nowhere.
	// Measured against pinned beets on these four objects. The mutation
	// that proves this test fires: pass true in AssignItems'
	// DistanceForTrack call.
	tun := DefaultTunables()
	items := []ItemFields{
		{Title: "same", Artist: "Ana Reed", Track: 1, Length: 200},
		{Title: "same", Artist: "Cole Vance", Track: 2, Length: 200},
	}
	tracks := []TrackInfo{
		{Title: "same", Artist: "Cole Vance", Index: 1, Length: 200},
		{Title: "same", Artist: "Ana Reed", Index: 2, Length: 200},
	}
	withArtist := AssignItems(tun, items, tracks)
	for _, p := range withArtist.Pairs {
		if cost := DistanceForTrack(tun, items[p.Item], &tracks[p.Track], true).Distance(); cost == 0 {
			t.Fatalf("pair %v scores 0 with the artist included; the fixture no longer discriminates the two objectives", p)
		}
	}
	// The assignment objective ignores the artist entirely, so the
	// index component alone decides and the identity is the unique
	// optimum at cost 0.0 (revision 2, fold B3: revision 1 asserted the
	// swap here, which is the answer an inclArtist-true matrix gives).
	want := []Pair{{Item: 0, Track: 0}, {Item: 1, Track: 1}}
	if !slices.Equal(withArtist.Pairs, want) {
		t.Errorf("Pairs = %v, want %v: the cost matrix must be built with inclArtist false (distance.py:383-385)", withArtist.Pairs, want)
	}
}

func TestAssignItemsMatchesEveryShapeOneSidedly(t *testing.T) {
	// Contract V4 and V3 over every rectangular shape, the empty sides
	// included: min(items, tracks) pairs, strictly ascending Track, and
	// never two non-empty extra lists.
	all := []ItemFields{
		{Title: "alpha", Track: 1, Length: 100},
		{Title: "bravo", Track: 2, Length: 150},
		{Title: "charlie", Track: 3, Length: 200},
	}
	every := []TrackInfo{
		{Title: "bravo", Index: 1, Length: 150},
		{Title: "charlie", Index: 2, Length: 200},
		{Title: "alpha", Index: 3, Length: 100},
		{Title: "delta", Index: 4, Length: 250},
	}
	for nItems := range len(all) + 1 {
		for nTracks := range len(every) + 1 {
			items, tracks := all[:nItems], every[:nTracks]
			got := AssignItems(DefaultTunables(), items, tracks)
			if want := min(nItems, nTracks); len(got.Pairs) != want {
				t.Errorf("%dx%d: %d pairs, want %d", nItems, nTracks, len(got.Pairs), want)
			}
			for k, p := range got.Pairs {
				if k > 0 && p.Track <= got.Pairs[k-1].Track {
					t.Errorf("%dx%d: Pairs %v is not strictly ascending in Track", nItems, nTracks, got.Pairs)
				}
			}
			if len(got.ExtraItems) > 0 && len(got.ExtraTracks) > 0 {
				t.Errorf("%dx%d: ExtraItems %v and ExtraTracks %v are both non-empty, which assign_items cannot produce",
					nItems, nTracks, got.ExtraItems, got.ExtraTracks)
			}
			if want := nItems - len(got.Pairs); len(got.ExtraItems) != want {
				t.Errorf("%dx%d: %d extra items, want %d", nItems, nTracks, len(got.ExtraItems), want)
			}
			if want := nTracks - len(got.Pairs); len(got.ExtraTracks) != want {
				t.Errorf("%dx%d: %d extra tracks, want %d", nItems, nTracks, len(got.ExtraTracks), want)
			}
		}
	}
}

func TestAlbumPairsForKeepsTrackOrder(t *testing.T) {
	// AlbumPairsFor hands DistanceForAlbum its pairing in the
	// assignment's own track order, and points each pair at the
	// candidate's track rather than a copy of it (contract V3).
	items := []ItemFields{
		{Title: "delta", Track: 1, Length: 100},
		{Title: "alpha", Track: 2, Length: 100},
	}
	info := &AlbumInfo{Tracks: []TrackInfo{
		{Title: "alpha", Index: 1, Length: 100},
		{Title: "bravo", Index: 2, Length: 100},
		{Title: "delta", Index: 3, Length: 100},
	}}
	a := AssignItems(DefaultTunables(), items, info.Tracks)
	pairs := AlbumPairsFor(a, items, info)
	if len(pairs) != len(a.Pairs) {
		t.Fatalf("AlbumPairsFor returned %d pairs, want %d", len(pairs), len(a.Pairs))
	}
	for k, p := range a.Pairs {
		if pairs[k].Item != items[p.Item] {
			t.Errorf("pair %d carries item %+v, want %+v", k, pairs[k].Item, items[p.Item])
		}
		if pairs[k].Info != &info.Tracks[p.Track] {
			t.Errorf("pair %d points at %p, want the candidate's own track %p", k, pairs[k].Info, &info.Tracks[p.Track])
		}
	}
	if got := []string{pairs[0].Info.Title, pairs[1].Info.Title}; !slices.Equal(got, []string{"alpha", "delta"}) {
		t.Errorf("track titles %v, want [alpha delta] in track order", got)
	}
}

// assignFixture is one pinned tied case: the cost matrix beets solved,
// the assignment lap 0.5.13 chose among the equal optima, and the
// optimum corpus/analysis/ties.json records for the same case.
type assignFixture struct {
	CaseID      string      `json:"case_id"`
	Rows        int         `json:"rows"`
	Cols        int         `json:"cols"`
	Costs       [][]float64 `json:"costs"`
	Want        []int       `json:"want"`
	OptimalCost float64     `json:"optimal_cost"`
}

// loadAssignFixture reads one testdata/assign fixture and checks its
// own shape, so a truncated or half-written matrix fails as a fixture
// error rather than as a solver divergence.
func loadAssignFixture(t *testing.T, name string) assignFixture {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "assign", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fx assignFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("decode %s: %v", name, err)
	}
	if len(fx.Costs) != fx.Rows || len(fx.Want) != fx.Cols {
		t.Fatalf("%s: %d rows and %d wants, header says %dx%d", name, len(fx.Costs), len(fx.Want), fx.Rows, fx.Cols)
	}
	for i, row := range fx.Costs {
		if len(row) != fx.Cols {
			t.Fatalf("%s: row %d has %d cells, header says %d columns", name, i, len(row), fx.Cols)
		}
	}
	return fx
}

// lapjvRow is one frozen differential row: a matrix and the assignment
// pinned lap returned for it.
type lapjvRow struct {
	Name  string      `json:"name"`
	Costs [][]float64 `json:"costs"`
	Want  []int       `json:"want"`
}

// lapjvTable is testdata/lapjv-optima.json, header and all. Lap, Numpy
// and Command are read rather than ignored so a table generated by a
// different build or a rewritten command cannot pass as this one.
type lapjvTable struct {
	Lap     string     `json:"lap"`
	Numpy   string     `json:"numpy"`
	Command string     `json:"command"`
	Rows    []lapjvRow `json:"rows"`
}

func loadLapjvOptima(t *testing.T) lapjvTable {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "lapjv-optima.json"))
	if err != nil {
		t.Fatalf("read frozen table: %v", err)
	}
	var table lapjvTable
	if err := json.Unmarshal(raw, &table); err != nil {
		t.Fatalf("decode frozen table: %v", err)
	}
	if table.Lap != "0.5.13" || table.Numpy != "2.5.2" {
		t.Fatalf("frozen table was generated against lap %q and numpy %q, want the pins 0.5.13 and 2.5.2", table.Lap, table.Numpy)
	}
	if table.Command == "" {
		t.Fatal("frozen table carries no regeneration command; the provenance header is what makes it re-derivable")
	}
	return table
}

func TestSolveLAPMirrorsLapjvOnThePinnedTiedCases(t *testing.T) {
	// The two cases corpus/analysis/ties.json pins. Both optima cost the
	// same, so an optimality assertion cannot discriminate them; only
	// the exact choice can, and reproducing lap.lapjv's choice is what
	// ladder entry 2's "pinned" disposition requires.
	for _, name := range []string{"should-ask", "should-reject"} {
		t.Run(name, func(t *testing.T) {
			fx := loadAssignFixture(t, name+"-costs.json")
			got := solveLAP(fx.Costs)
			if !slices.Equal(got, fx.Want) {
				t.Errorf("solveLAP chose %v, lap 0.5.13 chose %v; the costs are equal, so this is a pivot-order divergence rather than a wrong optimum", got, fx.Want)
			}
			if c, want := assignmentCost(fx.Costs, got), fx.OptimalCost; math.Abs(c-want) > 1e-12 {
				t.Errorf("assignment cost %v, ties.json records the optimum as %v", c, want)
			}
		})
	}
}

func TestSolveLAPMirrorsTheFrozenLapjvTable(t *testing.T) {
	table := loadLapjvOptima(t)
	if len(table.Rows) < 400 {
		t.Fatalf("frozen table carries %d rows, want at least 400: a shrunken table is a silently weakened check", len(table.Rows))
	}
	square, rect := 0, 0
	for _, row := range table.Rows {
		got := solveLAP(row.Costs)
		if !slices.Equal(got, row.Want) {
			t.Errorf("%s: solveLAP %v, lap %v", row.Name, got, row.Want)
		}
		if len(row.Costs) == len(row.Costs[0]) {
			square++
		} else {
			rect++
		}
	}
	if square == 0 || rect == 0 {
		t.Errorf("table covers %d square and %d rectangular shapes; both are required, because the extension model only shows on rectangular input", square, rect)
	}
}
