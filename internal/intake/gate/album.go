package gate

import (
	"cmp"
	"crypto/sha256"
	"encoding/binary"
	"slices"

	"github.com/glw907/dubplate/internal/reject"
)

// AlbumResult is the verdict on a whole album. Verdict is empty for a pass,
// and Key is set only then. Warnings never block.
type AlbumResult struct {
	Verdict  reject.Code
	Key      []byte // SHA-256, or nil
	Warnings []reject.Code
}

// imageExtensions are the extensions that count as a cover image among an
// album's extras.
var imageExtensions = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

// CheckAlbum runs the per-album checks over files, each of which passed
// [CheckFile] for [UnitAlbum], and over the member paths of the album's
// extras.
//
// Positions must be unique, and run 1 to n on each disc and discs 1 to m,
// whatever the files' numbering restarts. A TRACKTOTAL must equal its
// disc's track count, or, on a multi-disc album, the whole album's; a
// DISCTOTAL must equal the disc count; and files must agree on each. The
// album-wide TRACKTOTAL passes with a warning. No check rewrites a tag.
func CheckAlbum(files []FileResult, extras []string) AlbumResult {
	sorted := slices.SortedFunc(slices.Values(files), func(a, b FileResult) int {
		return cmp.Or(cmp.Compare(a.Disc, b.Disc), cmp.Compare(a.Track, b.Track))
	})
	for i := 1; i < len(sorted); i++ {
		if sorted[i].Disc == sorted[i-1].Disc && sorted[i].Track == sorted[i-1].Track {
			return AlbumResult{Verdict: reject.DuplicateTrackPosition}
		}
	}

	discs := discTracks(sorted)
	totalsWide, ok := completeness(sorted, discs)
	if !ok {
		return AlbumResult{Verdict: reject.IncompleteAlbum}
	}

	res := AlbumResult{Key: albumKey(sorted)}
	if !slices.ContainsFunc(files, func(f FileResult) bool { return f.HasPicture }) &&
		!slices.ContainsFunc(extras, func(name string) bool { return slices.Contains(imageExtensions, extension(name)) }) {
		res.Warnings = append(res.Warnings, reject.NoCoverArt)
	}
	if totalsWide {
		res.Warnings = append(res.Warnings, reject.TrackTotalDiffers)
	}
	return res
}

// discTracks groups position-sorted files by disc, in disc order.
func discTracks(sorted []FileResult) [][]FileResult {
	var discs [][]FileResult
	for _, f := range sorted {
		if n := len(discs); n == 0 || discs[n-1][0].Disc != f.Disc {
			discs = append(discs, nil)
		}
		discs[len(discs)-1] = append(discs[len(discs)-1], f)
	}
	return discs
}

// completeness reports whether the album has no break and no total that
// contradicts what was uploaded, and whether any disc carried the
// album-wide total instead of its own.
func completeness(sorted []FileResult, discs [][]FileResult) (wide, ok bool) {
	for i, disc := range discs {
		if disc[0].Disc != i+1 || disc[len(disc)-1].Track != len(disc) {
			return false, false
		}
	}
	if total, agreed := agreedTotal(sorted, func(f FileResult) int { return f.DiscTotal }); !agreed || (total != 0 && total != len(discs)) {
		return false, false
	}
	for _, disc := range discs {
		total, agreed := agreedTotal(disc, func(f FileResult) int { return f.TrackTotal })
		switch {
		case !agreed:
			return false, false
		case total == 0 || total == len(disc):
		case len(discs) > 1 && total == len(sorted):
			wide = true
		default:
			return false, false
		}
	}
	return wide, true
}

// agreedTotal returns the one non-zero value field takes across files, 0
// when none carries it, and false when two carry different values.
func agreedTotal(files []FileResult, field func(FileResult) int) (total int, agreed bool) {
	for _, f := range files {
		switch v := field(f); {
		case v == 0 || v == total:
		case total == 0:
			total = v
		default:
			return 0, false
		}
	}
	return total, true
}

// albumKey is the SHA-256 over the track count, as eight big-endian bytes,
// and the audio keys in (disc, track) order.
func albumKey(sorted []FileResult) []byte {
	h := sha256.New()
	h.Write(binary.BigEndian.AppendUint64(nil, uint64(len(sorted))))
	for _, f := range sorted {
		h.Write(f.AudioKey)
	}
	return h.Sum(nil)
}
