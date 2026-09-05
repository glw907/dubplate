package musicbrainz

import (
	"cmp"
	"slices"
)

// GroupRelations splits rels by target type, mirroring the client's own
// regrouping of a generic relations array into <target_type>_relations
// keys (beetsplug/_utils/musicbrainz.py:759-772). The sort is stable so
// within-type order survives, which Python's sorted() gives for free.
//
// The returned map's key iteration order is unobservable and callers
// must not depend on it: every beets consumer reads a single named key
// through .get, so no ordering of the map itself is part of the
// contract. Callers pass a release Validate has accepted, so a relation
// with an empty target type cannot reach here.
//
// The destructive-collision rule beets' own regrouping carries (a
// relations key colliding with a pre-existing <type>_relations key,
// later-wins in both directions) does not apply here and has no
// expression in this function: that rule arises because beets writes
// regrouped keys into the same dict a generic relations key may already
// share a name with. Release has no such field, since its only
// relations field is the generic array, so there is nothing for a
// regrouped group to collide with.
func GroupRelations(rels []Relation) map[string][]Relation {
	sorted := slices.Clone(rels)
	slices.SortStableFunc(sorted, func(a, b Relation) int {
		return cmp.Compare(a.TargetType, b.TargetType)
	})

	groups := make(map[string][]Relation)
	for _, r := range sorted {
		groups[r.TargetType] = append(groups[r.TargetType], r)
	}
	return groups
}
