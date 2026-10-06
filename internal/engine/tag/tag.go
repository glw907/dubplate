package tag

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// ErrNoCandidates reports that the flow admitted nothing: every lookup
// 404'd, the search returned no ids, or every returned release was
// dropped by an admission gate. It is a sentinel because a caller
// branches on it: rung 9's pipeline reports "no match" differently from
// a transport failure.
var ErrNoCandidates = errors.New("no candidates")

// A Proposal is the matcher's answer for one album: beets' Proposal
// (autotag/match.py:247-249) with tag_album's two leading return values
// folded in, plus the action beets' own quiet-mode mapping would take.
type Proposal struct {
	CurArtist      string
	CurAlbum       string
	Candidates     []matcher.Candidate // ascending distance, stable
	Recommendation matcher.Recommendation
	Action         matcher.Action
}

// Options carries tag_album's optional inputs.
//
// SearchIDs is import.search_ids (beets/config_default.yaml:50), which
// selects the explicit-ID branch and takes no early return
// (match.py:417-420). beets' search_artist and search_name overrides are
// a recorded non-port (3c's close): no committed profile sets them and
// the automatic album-match path is rung 3's whole scope, so they never
// join this struct.
type Options struct {
	SearchIDs []string
}

// Album is tag_album (autotag/match.py:383-468): the matcher's whole
// public entry point.
//
// It issues one release lookup per item-consensus album id, scores what
// comes back, and returns WITHOUT searching when that match is strong
// and tun.Timid is false. Otherwise it searches, fans out over the
// returned ids in the search response's order, and returns every
// admitted candidate sorted by distance. The interleaving is
// load-bearing rather than incidental: the request set is exactly what
// internal/harness/conform's tier 1 grades, so a "fetch everything, then
// score" rearrangement would issue requests beets never issued.
//
// Every request Album issues spends one token from src's own budget, so
// a caller caps its MusicBrainz spend by the budget it built the client
// with (musicbrainz.NewClient requires one, client.go:128-130). A
// conformance run passes musicbrainz.Unmetered(), deferred work passes a
// background budget, and the import pipeline passes a batch one.
//
// evalDate is a parameter and never a clock read.
//
// It returns ErrNoCandidates when nothing was admitted. A lookup that
// 404s drops that candidate silently, reproducing album_for_id's return
// None (beetsplug/musicbrainz.py:777-781). Every other client error is
// returned wrapped, including musicbrainz.ErrUnavailable, which a caller can
// branch on to defer the lookup.
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
	tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error) {
	tun = compat.Apply(tun)
	if err := tun.Validate(); err != nil {
		return nil, fmt.Errorf("tag: %w", err)
	}

	likelies, consensus := matcher.MostCommonTags(items)
	cands := matcher.NewCandidates()

	if len(opts.SearchIDs) > 0 {
		// import.search_ids: look each id up and add it, taking NO
		// early return and skipping the search entirely
		// (match.py:417-420).
		for _, id := range opts.SearchIDs {
			if err := lookupAndAdd(ctx, src, cands, tun, items, id, evalDate); err != nil {
				return nil, err
			}
		}
	} else {
		// match_by_id (match.py:252-266): fires only on a unanimous
		// mb_albumid consensus, and beets logs neither early branch
		// because internal/engine has no logger.
		if likelies.MBAlbumID != "" && consensus["mb_albumid"] {
			if err := lookupAndAdd(ctx, src, cands, tun, items, likelies.MBAlbumID, evalDate); err != nil {
				return nil, err
			}

			// The early return: _recommendation is called on the
			// UNSORTED insertion-order slice at match.py:430, and the
			// early return at :432-441 ships that unsorted list.
			// Harmless in practice only because match_by_id yields at
			// most one candidate here, so sorting first could never
			// change the outcome (contract V18).
			rec, err := matcher.Recommend(tun, cands.All())
			if err != nil {
				return nil, fmt.Errorf("tag: %w", err)
			}
			if cands.Len() > 0 && !tun.Timid && rec == matcher.RecStrong {
				action, err := matcher.ActionFor(tun, rec)
				if err != nil {
					return nil, fmt.Errorf("tag: %w", err)
				}
				return &Proposal{
					CurArtist:      likelies.Artist,
					CurAlbum:       likelies.Album,
					Candidates:     cands.All(),
					Recommendation: rec,
					Action:         action,
				}, nil
			}
		}

		// The search (match.py:451-463): vaLikely is computed inside
		// AlbumSearchFor and stays unexported there, because no
		// caller outside internal/engine/musicbrainz needs the
		// no-consensus branch it feeds (3c carry 13c).
		search := musicbrainz.AlbumSearchFor(items, tun)
		ids, err := src.SearchReleases(ctx, search)
		if err != nil {
			return nil, fmt.Errorf("tag: searching: %w", err)
		}
		for _, id := range ids {
			if err := lookupAndAdd(ctx, src, cands, tun, items, id, evalDate); err != nil {
				return nil, err
			}
		}
	}

	sorted := cands.Sorted()
	if len(sorted) == 0 {
		return nil, ErrNoCandidates
	}
	rec, err := matcher.Recommend(tun, sorted)
	if err != nil {
		return nil, fmt.Errorf("tag: %w", err)
	}
	action, err := matcher.ActionFor(tun, rec)
	if err != nil {
		return nil, fmt.Errorf("tag: %w", err)
	}
	return &Proposal{
		CurArtist:      likelies.Artist,
		CurAlbum:       likelies.Album,
		Candidates:     sorted,
		Recommendation: rec,
		Action:         action,
	}, nil
}

// lookupAndAdd fetches one release by MusicBrainz id and admits it into
// cands. A 404 drops the candidate silently, reproducing
// album_for_id's return None (beetsplug/musicbrainz.py:777-781); every
// other client error, including musicbrainz.ErrUnavailable, is returned
// wrapped.
func lookupAndAdd(ctx context.Context, src *musicbrainz.Client, cands *matcher.Candidates,
	tun matcher.Tunables, items []matcher.ItemFields, mbid string, evalDate time.Time) error {
	info, err := src.AlbumForID(ctx, mbid, tun)
	if err != nil {
		if errors.Is(err, musicbrainz.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("tag: looking up %s: %w", mbid, err)
	}
	if _, err := cands.Add(tun, items, info, evalDate); err != nil {
		return fmt.Errorf("tag: %w", err)
	}
	return nil
}
