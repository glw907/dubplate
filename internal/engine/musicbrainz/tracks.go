package musicbrainz

import (
	"github.com/glw907/dubplate/internal/engine/matcher"
)

// dataTrackTitle is the recording title beets filters out unconditionally,
// regardless of tun.IgnoreDataTracks (musicbrainz.py:573, T5). It is a
// literal comparison, no strip, no case fold.
const dataTrackTitle = "[data track]"

// builtTrack is the matcher's TrackInfo plus the one field the
// translation needs and the vocabulary deliberately does not carry: the
// medium's own format string, from which AlbumInfo.Media is derived
// (beetsplug/musicbrainz.py:654-658). TrackInfo declares itself "the
// eight fields the track-level distance path reads ... and nothing
// else" (matcher/info.go:69-72), and per-track media is never read by
// that path, so it stays here rather than widening the vocabulary (R2).
//
// Info.Index and Info.DataSource are left zero here: Index is assigned
// by extractTracks across all valid media, and DataSource is the one
// literal every AlbumInfo and TrackInfo share, which Task 6's
// TranslateAlbum assigns once rather than repeating it per track.
type builtTrack struct {
	Info   matcher.TrackInfo
	Format *string
}

// tracksFromMedium yields one builtTrack per surviving track on m,
// mirroring get_tracks_from_medium (beetsplug/musicbrainz.py:555-600):
// assembledTracks' fixed order, with data tracks joining only when
// tun.IgnoreDataTracks is false; then the two filters, the
// "[data track]" title drop that runs regardless of IgnoreDataTracks
// and the video drop under tun.IgnoreVideoTracks; then the three
// track-level overrides. Index is left zero here and assigned by the
// caller across all valid media.
func tracksFromMedium(m Medium, tun matcher.Tunables) []builtTrack {
	var built []builtTrack
	for _, tr := range assembledTracks(m, !tun.IgnoreDataTracks) {
		if tr.Recording.Title == dataTrackTitle {
			continue
		}
		if tun.IgnoreVideoTracks && tr.Recording.Video {
			continue
		}
		built = append(built, buildTrack(m, tr, tun))
	}
	return built
}

// buildTrack applies T6's three track-level overrides to tr's recording
// and returns the resulting builtTrack. The length and artist-credit
// overrides are truthiness, not nil-ness (a track length of 0 falls
// back to the recording length exactly as a null does); the title
// override additionally guards on preferredAlias, which is why it reads
// recording.Aliases a second time through keyWithPreferredAlias.
func buildTrack(m Medium, tr Track, tun matcher.Tunables) builtTrack {
	recording := tr.Recording

	length := recording.Length
	if tr.Length != nil && *tr.Length != 0 {
		length = tr.Length
	}
	var lengthMS float64
	if length != nil {
		lengthMS = *length
	}

	title := recording.Title
	if tr.Title != "" && preferredAlias(recording.Aliases, tun) == nil {
		title = tr.Title
	}
	title = keyWithPreferredAlias(recording.Aliases, title, tun)

	artist := parseArtistCredits(resolvedTrackCredit(tr), tun).Artist

	return builtTrack{
		Info: matcher.TrackInfo{
			Title:       title,
			Artist:      artist,
			Length:      lengthMS / 1000.0,
			MediumIndex: int(tr.Position),
			Medium:      m.Position,
			TrackID:     recording.ID,
		},
		Format: m.Format,
	}
}

// extractTracks runs tracksFromMedium over every medium of rel whose
// format is not in tun.IgnoredMedia, then assigns Index 1-based across
// all surviving tracks in medium order (musicbrainz.py:645-646).
//
// It does not compute AlbumInfo.Mediums: that is len(rel.Media), the
// UNFILTERED count (musicbrainz.py:659), and it is deliberately not
// derived here, because a reader who saw both numbers produced by one
// function would reasonably assume they agree.
func extractTracks(rel *Release, tun matcher.Tunables) []builtTrack {
	ignored := make(map[string]struct{}, len(tun.IgnoredMedia))
	for _, format := range tun.IgnoredMedia {
		ignored[format] = struct{}{}
	}

	var all []builtTrack
	for _, m := range rel.Media {
		if m.Format != nil {
			if _, skip := ignored[*m.Format]; skip {
				continue
			}
		}
		all = append(all, tracksFromMedium(m, tun)...)
	}

	for i := range all {
		all[i].Info.Index = i + 1
	}
	return all
}

// deriveMedia returns AlbumInfo.Media for the given built tracks: the
// single format when exactly one distinct format survives, and
// otherwise the literal "Media" (musicbrainz.py:654-658). Zero tracks
// and mixed formats both land on the sentinel, which is truthy and
// which the media distance component therefore scores as if it were a
// media name (autotag/distance.py:455).
//
// A single medium whose format is null is NOT a sentinel route: it is
// the one distinct value beets' set collects, so pop() returns None,
// which this port maps to "" (R1's zero-value mapping), and which is
// falsy and declines the media penalty (fold, B5). A null format
// alongside any other distinct format IS a sentinel route, because None
// still counts as one of the distinct values.
func deriveMedia(tracks []builtTrack) string {
	formats := make(map[string]struct{})
	var sawNilFormat bool
	for _, t := range tracks {
		if t.Format == nil {
			sawNilFormat = true
			continue
		}
		formats[*t.Format] = struct{}{}
	}

	distinct := len(formats)
	if sawNilFormat {
		distinct++
	}
	switch {
	case distinct == 1 && sawNilFormat:
		return ""
	case distinct == 1:
		for format := range formats {
			return format
		}
	}
	return "Media"
}
