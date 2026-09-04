package matcher

import "time"

// addAlbumYear is distance()'s year component (autotag/distance.py:
// 471-493): three branches, tried in order, and evalDate is an
// explicit parameter rather than a clock read (contract 9) because
// both ratio branches divide by a span ending at its year.
//
//  1. With Preferred.OriginalYear set and a candidate release year,
//     the candidate is graded against its own original year (or 1889,
//     the earliest first gramophone discs, when it carries none),
//     ratioed against the same span up to evalDate's year.
//  2. Otherwise, with both a likely year and a candidate release year:
//     zero penalty when the likely year matches either the release or
//     the original year; a ratioed penalty when the candidate carries
//     an original year; a flat 1.0 otherwise.
//  3. Neither condition holds: decline (no penalty added at all).
//
// preferred.original_year is one of the three corpus-unreachable
// surfaces (spec, M8); branch 1's coverage is this file's own
// testdata/album_year_preferred_original.json table rather than a
// corpus fixture.
func addAlbumYear(dist *Distance, tun Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, evalDate time.Time) {
	evalYear := evalDate.Year()

	if info.Year != 0 && tun.Preferred.OriginalYear {
		original := info.OriginalYear
		if original == 0 {
			original = 1889
		}
		diff := absInt(info.Year - original)
		diffMax := absInt(evalYear - original)
		dist.AddRatio("year", float64(diff), float64(diffMax))
		return
	}

	if likelies.Year == 0 || info.Year == 0 {
		return
	}

	if likelies.Year == info.Year || likelies.Year == info.OriginalYear {
		dist.Add("year", 0)
		return
	}
	if info.OriginalYear != 0 {
		diff := absInt(likelies.Year - info.Year)
		diffMax := absInt(evalYear - info.OriginalYear)
		dist.AddRatio("year", float64(diff), float64(diffMax))
		return
	}
	dist.Add("year", 1)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
