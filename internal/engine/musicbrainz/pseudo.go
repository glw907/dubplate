package musicbrainz

import "github.com/glw907/dubplate/internal/engine/matcher"

// pseudoReleaseStatus is the exact status string album_for_id compares
// against (beetsplug/musicbrainz.py:785, "Pseudo-Release").
const pseudoReleaseStatus = "Pseudo-Release"

// PseudoReleaseTarget reports the release id whose AlbumInfo must be
// merged into rel's, and whether rel is a pseudo-release at all,
// mirroring album_for_id's own check (beetsplug/musicbrainz.py:785-798):
// rel.Status is exactly "Pseudo-Release" and some release relation
// carries type "transl-tracklisting" with direction "backward".
//
// The fetch of that release is the client's work; this predicate names
// the target and MergePseudoRelease merges over two already-built
// AlbumInfo values.
func PseudoReleaseTarget(rel *Release) (releaseID string, ok bool) {
	if rel.Status != pseudoReleaseStatus {
		return "", false
	}

	for _, r := range rel.Relations {
		if isTranslSource(r) {
			return r.Release.ID, true
		}
	}
	return "", false
}

// isTranslSource reports whether r is a release relation of type
// "transl-tracklisting" pointing backward, the link from a pseudo-release
// to the release it translates. beets reads these through the
// release_relations group it regroups out of the generic relations array
// (beetsplug/_utils/musicbrainz.py:759-772); filtering by target type
// here gives the same relations in the same response order.
func isTranslSource(r Relation) bool {
	return r.TargetType == "release" && r.Type == "transl-tracklisting" && r.Direction == "backward"
}

// MergePseudoRelease folds an actual release's own fields into a
// pseudo-release candidate, mirroring _merge_pseudo_and_actual_album
// (beetsplug/musicbrainz.py:203-240): the pseudo supplies identity and
// titles, the actual supplies the physical and date fields. It returns a
// new value; beets starts from pseudo.copy(), which is a deepcopy
// (autotag/hooks.py:93-95), and nothing here mutates either argument.
//
// Beets moves fifteen keys. Seven of them are in matcher.AlbumInfo and
// are the merge's entire verdict-moving effect: Media, Mediums,
// Country, CatalogNum, Year, OriginalYear, Label. The other eight
// (month, day, original_month, original_day, barcode, asin, style,
// genres) are outside the vocabulary. Album, Artist, VA, AlbumDisambig,
// AlbumID, DataSource and Tracks stay the pseudo's.
//
// The copy is shallow: merged.Tracks shares the pseudo's backing array,
// so a caller that mutates a merged track element writes through to the
// pseudo candidate. A caller that retains both values must clone Tracks.
//
// Both arguments are total values, so the KeyError beets can raise on
// actual[k] (musicbrainz.py:220) has no expression here.
func MergePseudoRelease(pseudo, actual *matcher.AlbumInfo) *matcher.AlbumInfo {
	merged := *pseudo
	merged.Media = actual.Media
	merged.Mediums = actual.Mediums
	merged.Country = actual.Country
	merged.CatalogNum = actual.CatalogNum
	merged.Year = actual.Year
	merged.OriginalYear = actual.OriginalYear
	merged.Label = actual.Label
	return &merged
}
