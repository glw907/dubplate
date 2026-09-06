package matcher

import (
	"cmp"
	"math"
	"slices"
)

// A Pair is one item index matched to one track index.
//
// Indexes rather than values, because beets keys its mapping on object
// identity (id(item), derive_oracle.py:184-189) while Go ItemFields is a
// comparable struct value, so two identical items would collapse into
// one map key. The index form is also what cassette.Expected.Assignment
// grades (cassette.go:84).
type Pair struct {
	Item  int
	Track int
}

// An Assignment is one solved item-to-track matching, in beets' own
// output shape (autotag/match.py:53-81).
//
// Pairs is in ASCENDING TRACK INDEX order, not item order. beets builds
// its mapping by zipping the solver's track-indexed result with the
// track list (match.py:72-76), and distance() then walks that dict to
// build both dist.tracks and the "tracks" penalty list
// (autotag/distance.py:527-530). Float addition is not associative, so
// the order fixes both the graded Tracks vector and the summed
// distance, and a pair list in item order is wrong in two ways at once.
type Assignment struct {
	Pairs       []Pair
	ExtraItems  []int // item indexes, sorted by (disc, track, title)
	ExtraTracks []int // track indexes, sorted by (index, title)
}

// AssignItems solves beets' track assignment (autotag/match.py:53-81):
// the minimum-cost matching between items and a candidate's tracks,
// where cell (i, t) is the normalized track distance between them.
//
// The cost is DistanceForTrack with inclArtist FALSE for every cell,
// unconditionally, including on a various-artists release. beets'
// assignment objective uses track_distance's own default
// (autotag/distance.py:383-385) while the final per-track distances a
// user sees pass album_info.va (distance.py:529). The two objectives
// disagree exactly on VA releases, and no committed corpus case
// discriminates them, so the hand-authored fixture in assign_test.go,
// TestAssignItemsUsesInclArtistFalseOnAVariousArtistsRelease, is what
// proves this line right: the index component makes the identity the
// unique optimum with the artist excluded, and including the artist
// flips the unique optimum to the swap.
//
// Exactly min(len(items), len(tracks)) pairs are returned, so at most
// one of ExtraItems and ExtraTracks is non-empty.
//
// The extras carry beets' own sort keys, (disc, track, title) for items
// and (index, title) for tracks (match.py:77-80). beets sorts a Python
// set of identity-hashed objects, whose pre-sort order derives from
// addresses and lies outside PYTHONHASHSEED's reach, so a committed
// record is reproducible only while its sort key is total over that
// case's extras. The track key is total by construction, because
// album_info renumbers every track 1..n across the release
// (beetsplug/musicbrainz.py:645-646). The item key is total as a fact
// of today's corpus rather than a guarantee: no committed case holds
// two extra items sharing (disc, track, title). Both sorts are stable
// so that the port stays faithful the day one does.
//
// An empty items or tracks slice returns an Assignment with no pairs
// and the whole non-empty side as extras, never a panic. beets never
// reaches that shape: _add_candidate's first gate drops a zero-track
// candidate before assign_items runs (match.py:345-347), and an album
// with no items is not an import. The behavior is defined anyway so a
// later caller outside that gate finds an answer rather than an index
// out of range.
func AssignItems(tun Tunables, items []ItemFields, tracks []TrackInfo) Assignment {
	var a Assignment
	if len(items) > 0 && len(tracks) > 0 {
		costs := make([][]float64, len(items))
		for i := range costs {
			costs[i] = make([]float64, len(tracks))
			for t := range costs[i] {
				costs[i][t] = DistanceForTrack(tun, items[i], &tracks[t], false).Distance()
			}
		}
		for t, item := range solveLAP(costs) {
			if item >= 0 {
				a.Pairs = append(a.Pairs, Pair{Item: item, Track: t})
			}
		}
	}

	matchedItem := make([]bool, len(items))
	matchedTrack := make([]bool, len(tracks))
	for _, p := range a.Pairs {
		matchedItem[p.Item] = true
		matchedTrack[p.Track] = true
	}
	for i, matched := range matchedItem {
		if !matched {
			a.ExtraItems = append(a.ExtraItems, i)
		}
	}
	for t, matched := range matchedTrack {
		if !matched {
			a.ExtraTracks = append(a.ExtraTracks, t)
		}
	}
	slices.SortStableFunc(a.ExtraItems, func(x, y int) int {
		return cmp.Or(
			cmp.Compare(items[x].Disc, items[y].Disc),
			cmp.Compare(items[x].Track, items[y].Track),
			cmp.Compare(items[x].Title, items[y].Title),
		)
	})
	slices.SortStableFunc(a.ExtraTracks, func(x, y int) int {
		return cmp.Or(
			cmp.Compare(tracks[x].Index, tracks[y].Index),
			cmp.Compare(tracks[x].Title, tracks[y].Title),
		)
	})
	return a
}

// AlbumPairsFor projects an assignment onto the pairing DistanceForAlbum
// consumes, preserving track order (contract V3).
func AlbumPairsFor(a Assignment, items []ItemFields, info *AlbumInfo) []AlbumPair {
	pairs := make([]AlbumPair, len(a.Pairs))
	for k, p := range a.Pairs {
		pairs[k] = AlbumPair{Item: items[p.Item], Info: &info.Tracks[p.Track]}
	}
	return pairs
}

// solveLAP returns, for each column, the row assigned to it, or -1.
// It reproduces lap.lapjv(costs, extend_cost=True) (match.py:66):
// a dense Jonker-Volgenant solve over the matrix padded to
// n = max(rows, cols) with ZERO, which is the extension model measured
// against the pinned solver (contract V5) rather than the
// (rows+cols, max+1) recipe derive_ties.py's docstring assumes, which
// disagrees with this lap build on 30% of rectangular ties.
//
// The reproduction is exact rather than merely optimal: a matrix with
// several equally cheap assignments has to come back with the one lap
// 0.5.13 returns, because two committed corpus cases record that choice
// and nothing but the pivot order decides it. The three routines below
// are that solver's own, in its own order, and
// testdata/lapjv-optima.json freezes 404 of its answers as the
// differential evidence.
func solveLAP(costs [][]float64) []int {
	rows := len(costs)
	if rows == 0 {
		return nil
	}
	cols := len(costs[0])
	if cols == 0 {
		return nil
	}
	n := max(rows, cols)

	c := make([][]float64, n)
	for i := range c {
		c[i] = make([]float64, n)
		if i < rows {
			copy(c[i], costs[i])
		}
	}

	x := make([]int, n) // row to column
	y := make([]int, n) // column to row
	v := make([]float64, n)
	free := columnReduce(c, x, y, v)
	// Two augmenting row reduction sweeps, then Dijkstra augmentation
	// for whatever is still free: lap's own schedule, which bounds the
	// cheap sweeps rather than iterating them to a fixed point.
	for range 2 {
		if len(free) == 0 {
			break
		}
		free = augmentingRowReduce(c, free, x, y, v)
	}
	augment(c, free, x, y, v)

	assigned := make([]int, cols)
	for j := range assigned {
		if y[j] < rows {
			assigned[j] = y[j]
		} else {
			assigned[j] = -1
		}
	}
	return assigned
}

// columnReduce assigns each column to its cheapest row, keeps the
// assignment only where that row is one column's unique cheapest, and
// transfers the slack of every uniquely assigned row back into its
// column's dual. It returns the rows left free.
func columnReduce(c [][]float64, x, y []int, v []float64) []int {
	n := len(c)
	unique := make([]bool, n)
	for i := range n {
		x[i] = -1
		unique[i] = true
		v[i] = math.Inf(1)
	}
	for i := range n {
		for j := range n {
			if c[i][j] < v[j] {
				v[j] = c[i][j]
				y[j] = i
			}
		}
	}
	for j := n - 1; j >= 0; j-- {
		i := y[j]
		if x[i] < 0 {
			x[i] = j
			continue
		}
		unique[i] = false
		y[j] = -1
	}

	var free []int
	for i := range n {
		if x[i] < 0 {
			free = append(free, i)
			continue
		}
		if !unique[i] {
			continue
		}
		j := x[i]
		slack := math.Inf(1)
		for j2 := range n {
			if j2 != j {
				slack = min(slack, c[i][j2]-v[j2])
			}
		}
		v[j] -= slack
	}
	return free
}

// augmentingRowReduce runs one sweep of lap's augmenting row reduction
// over the free rows, assigning each to its cheapest reduced column and
// returning the rows its reassignments freed in turn.
func augmentingRowReduce(c [][]float64, free []int, x, y []int, v []float64) []int {
	n := len(c)
	var newFree []int
	// rrCount bounds the requeue-in-place path, so a pathological
	// matrix falls back to plain augmentation instead of cycling here.
	rrCount := 0
	for current := 0; current < len(free); {
		rrCount++
		i := free[current]
		current++

		j1, j2 := 0, -1
		v1, v2 := c[i][0]-v[0], math.Inf(1)
		for j := 1; j < n; j++ {
			red := c[i][j] - v[j]
			if red >= v2 {
				continue
			}
			if red >= v1 {
				v2, j2 = red, j
				continue
			}
			v2, v1 = v1, red
			j2, j1 = j1, j
		}

		i0 := y[j1]
		lowers := v[j1]-(v2-v1) < v[j1]
		switch {
		case rrCount >= current*n:
			if i0 >= 0 {
				newFree = append(newFree, i0)
			}
		case lowers:
			v[j1] -= v2 - v1
			if i0 >= 0 {
				current--
				free[current] = i0
			}
		default:
			if i0 >= 0 && j2 >= 0 {
				j1 = j2
				i0 = y[j1]
			}
			if i0 >= 0 {
				newFree = append(newFree, i0)
			}
		}
		x[i] = j1
		y[j1] = i
	}
	return newFree
}

// augment routes each still-free row down a shortest augmenting path
// (Dijkstra over reduced costs) and flips the path into the assignment.
func augment(c [][]float64, free []int, x, y []int, v []float64) {
	pred := make([]int, len(c))
	for _, i := range free {
		j := findPath(c, i, y, v, pred)
		for row := -1; row != i; {
			row = pred[j]
			y[j] = row
			j, x[row] = x[row], j
		}
	}
}

// findPath returns the unassigned column ending the shortest augmenting
// path out of start, filling pred with the path's predecessors and
// updating the duals of every column the search settled.
func findPath(c [][]float64, start int, y []int, v []float64, pred []int) int {
	n := len(c)
	cols := make([]int, n)
	d := make([]float64, n)
	for j := range n {
		cols[j] = j
		pred[j] = start
		d[j] = c[start][j] - v[j]
	}

	lo, hi, ready := 0, 0, 0
	var minD float64
	final := -1
	for final < 0 {
		if lo == hi {
			ready = lo
			hi = lo + 1
			minD = d[cols[lo]]
			for k := hi; k < n; k++ {
				j := cols[k]
				if d[j] > minD {
					continue
				}
				if d[j] < minD {
					hi = lo
					minD = d[j]
				}
				cols[k] = cols[hi]
				cols[hi] = j
				hi++
			}
			for _, j := range cols[lo:hi] {
				if y[j] < 0 {
					final = j
				}
			}
		}
		if final >= 0 {
			break
		}
		j := cols[lo]
		lo++
		i := y[j]
		h := c[i][j] - v[j] - d[j]
		for k := hi; k < n; k++ {
			j := cols[k]
			red := c[i][j] - v[j] - h
			if red >= d[j] {
				continue
			}
			d[j] = red
			pred[j] = i
			if red != minD {
				continue
			}
			if y[j] < 0 {
				final = j
				break
			}
			cols[k] = cols[hi]
			cols[hi] = j
			hi++
		}
	}

	// The path's own length prices the columns the search settled, and
	// minD is that length whichever way the loop ended: a column taken
	// from the scan's own minimum window carries d[j] == minD, and so
	// does one the augmentation's inner loop found unassigned, since it
	// takes that branch only on red == minD.
	for _, j := range cols[:ready] {
		v[j] += d[j] - minD
	}
	return final
}
