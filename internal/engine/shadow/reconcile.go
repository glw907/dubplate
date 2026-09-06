package shadow

import (
	"cmp"
	"fmt"
	"math"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// lengthTolerance is how close two track lengths must be to count as
// the same track for the imported-row content join: beets rewrites
// tags but not audio, so an exact equality is the wrong bar.
const lengthTolerance = 1.0 // seconds

// A reconciliation is a reconciled observation and its oracle signal,
// or the reason it could not be resolved.
type reconciliation struct {
	Capture      Capture
	Oracle       *OracleSignal // nil when Unreconciled or Excluded is set
	Unreconciled string
	Excluded     string
	Superseded   bool
}

type reconcileInput struct {
	Captures      []Capture                                     // every stored, still-unreconciled fixture
	Reports       []oracleRunReport                             // every report-*.json currently on disk; reconcile filters out reports already named in Consumed
	Consumed      []ConsumedReport                              // the ledger from the last pass
	TagLogSkips   map[string]bool                               // basenames beets' tag log recorded as skipped, uploader-agnostic (readTagLogSkips)
	Reviewed      map[string][]string                           // uploader to the basenames filed under the review tree (reviewedBasenames)
	LibraryReadFn func(path string) (matcher.ItemFields, error) // readFLAC, injected so the test needs no library tree
}

type reconcileResult struct {
	Reconciliations []reconciliation
	Consumed        []ConsumedReport
	Gaps            int
}

// reconcile resolves each album outcome in the newly arrived run
// reports to a captured fixture and records the oracle signal
// (contract S10).
//
// It returns the reconciliations, the reports it consumed, and the
// number of gaps it detected. A gap means at least one report was swept
// before this shadow consumed it: the incumbent keeps at most thirty
// reports (importer.go:389, :533-548,
// _salvage/internal/config/config.go:172) and its timer fires *:0/15,
// which is roughly seven and a half hours of history rather than a
// durable ledger.
func reconcile(in reconcileInput) (reconcileResult, error) {
	consumedFiles := make(map[string]bool, len(in.Consumed))
	var newestConsumedEnded time.Time
	for _, c := range in.Consumed {
		consumedFiles[c.File] = true
		if c.EndedAt.After(newestConsumedEnded) {
			newestConsumedEnded = c.EndedAt
		}
	}

	var oldestPresent time.Time
	havePresent := false
	var newReports []oracleRunReport
	for _, r := range in.Reports {
		if !havePresent || r.StartedAt.Before(oldestPresent) {
			oldestPresent = r.StartedAt
			havePresent = true
		}
		if !consumedFiles[reportFileName(r.RunID)] {
			newReports = append(newReports, r)
		}
	}

	gaps := 0
	if len(in.Consumed) > 0 && havePresent && oldestPresent.After(newestConsumedEnded) {
		gaps = 1
	}

	claimed := make([]bool, len(in.Captures))
	var out []reconciliation

	supersedeDuplicateCaptures(in.Captures, claimed, &out)

	partialSkipUploader := make(map[string]bool)
	gateRejectedRemaining := make(map[string]int)
	for _, r := range newReports {
		for _, row := range r.Albums {
			switch row.Outcome {
			case outcomeImported:
				rec, err := resolveImportedRow(row, in.Captures, claimed, in.LibraryReadFn)
				if err != nil {
					return reconcileResult{}, err
				}
				if rec != nil {
					out = append(out, *rec)
				}
			case outcomeQuarantined, outcomeRejected:
				gateRejectedRemaining[row.Uploader]++
			case outcomeReview:
				if row.Error != "" {
					partialSkipUploader[row.Uploader] = true
				}
			}
		}
	}

	resolveSkips(in.Captures, claimed, in.TagLogSkips, in.Reviewed, partialSkipUploader, gateRejectedRemaining, &out)

	// A gate outcome (quarantined or rejected) never named a specific
	// capture: quarantineOne files an OutcomeRejected album into the
	// review tree by basename (_salvage/internal/review/importer.go:1014),
	// so resolveSkips claims the real capture when one matches. Any gate
	// outcome left with no matching capture becomes an anonymous,
	// uploader-only exclusion rather than being dropped.
	gateUploaders := make([]string, 0, len(gateRejectedRemaining))
	for u := range gateRejectedRemaining {
		gateUploaders = append(gateUploaders, u)
	}
	slices.Sort(gateUploaders)
	for _, u := range gateUploaders {
		for range gateRejectedRemaining[u] {
			out = append(out, reconciliation{Capture: Capture{Uploader: u}, Excluded: ExcludedGateRejected})
		}
	}

	if gaps > 0 {
		for i, c := range in.Captures {
			if claimed[i] {
				continue
			}
			if c.CapturedAt.Before(oldestPresent) {
				claimed[i] = true
				out = append(out, reconciliation{Capture: c, Unreconciled: UnreconciledReportSwept})
			}
		}
	}

	consumed := slices.Clone(in.Consumed)
	for _, r := range newReports {
		consumed = append(consumed, ConsumedReport{
			File:      reportFileName(r.RunID),
			StartedAt: r.StartedAt,
			EndedAt:   r.EndedAt,
		})
	}

	return reconcileResult{Reconciliations: out, Consumed: consumed, Gaps: gaps}, nil
}

// supersedeDuplicateCaptures groups captures by (Uploader, Basename)
// and marks every one but the newest by CapturedAt as claimed and
// Superseded (Resolution R5): several unreconciled captures can share
// one album when a re-upload arrives before the incumbent's next run
// reconciles the first.
func supersedeDuplicateCaptures(captures []Capture, claimed []bool, out *[]reconciliation) {
	type key struct{ uploader, basename string }
	groups := make(map[key][]int)
	for i, c := range captures {
		k := key{c.Uploader, c.Basename}
		groups[k] = append(groups[k], i)
	}

	keys := make([]key, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	slices.SortFunc(keys, func(a, b key) int {
		return cmp.Or(cmp.Compare(a.uploader, b.uploader), cmp.Compare(a.basename, b.basename))
	})

	for _, k := range keys {
		idxs := groups[k]
		if len(idxs) < 2 {
			continue
		}
		newest := idxs[0]
		for _, i := range idxs[1:] {
			if captures[i].CapturedAt.After(captures[newest].CapturedAt) {
				newest = i
			}
		}
		for _, i := range idxs {
			if i == newest {
				continue
			}
			claimed[i] = true
			*out = append(*out, reconciliation{Capture: captures[i], Superseded: true})
		}
	}
}

// resolveImportedRow joins one imported outcome to a capture by
// content (contract S10). LibraryPaths is a positional correlation
// rather than a join (importer.go:1119-1136), so the shadow reads the
// first named file and matches back to a still-unclaimed capture under
// the row's own uploader whose items carry the same track title and a
// length within lengthTolerance. An empty LibraryPaths, or a library
// file whose MUSICBRAINZ_ALBUMID is absent, is unreconciled with reason
// UnreconciledNoReleaseID rather than an apply (Resolution R22):
// recording either as an apply against an empty release would
// classify as go-applies-different-release against any Go apply, a
// divergence against a release the oracle never named.
func resolveImportedRow(row oracleAlbumOutcome, captures []Capture, claimed []bool, readFn func(string) (matcher.ItemFields, error)) (*reconciliation, error) {
	if len(row.LibraryPaths) == 0 {
		idx := soleUnclaimedForUploader(captures, claimed, row.Uploader)
		if idx < 0 {
			return nil, nil
		}
		claimed[idx] = true
		return &reconciliation{Capture: captures[idx], Unreconciled: UnreconciledNoReleaseID}, nil
	}

	fields, err := readFn(row.LibraryPaths[0])
	if err != nil {
		return nil, fmt.Errorf("read library file %s: %w", row.LibraryPaths[0], err)
	}

	idx := matchContentCapture(captures, claimed, row.Uploader, fields)
	if idx < 0 {
		return nil, nil
	}
	claimed[idx] = true

	if fields.MBAlbumID == "" {
		return &reconciliation{Capture: captures[idx], Unreconciled: UnreconciledNoReleaseID}, nil
	}
	return &reconciliation{
		Capture: captures[idx],
		Oracle: &OracleSignal{
			Signal:        "apply",
			ReleaseMBID:   fields.MBAlbumID,
			Uploader:      row.Uploader,
			AlbumBasename: captures[idx].Basename,
			ReadFrom:      row.LibraryPaths[0],
		},
	}, nil
}

// matchContentCapture returns the index of the first unclaimed capture
// under uploader carrying a track whose title matches fields.Title and
// whose length is within lengthTolerance of fields.Length, or -1.
func matchContentCapture(captures []Capture, claimed []bool, uploader string, fields matcher.ItemFields) int {
	for i, c := range captures {
		if claimed[i] || c.Uploader != uploader {
			continue
		}
		for _, item := range c.Items {
			if item.Title == fields.Title && math.Abs(item.Length-fields.Length) < lengthTolerance {
				return i
			}
		}
	}
	return -1
}

// soleUnclaimedForUploader returns the index of the single unclaimed
// capture under uploader, or -1 when there is none or more than one:
// an empty LibraryPaths carries no content to join on, so this is the
// only join available, and it is deliberately refused rather than
// guessed when it would be ambiguous.
func soleUnclaimedForUploader(captures []Capture, claimed []bool, uploader string) int {
	found := -1
	for i, c := range captures {
		if claimed[i] || c.Uploader != uploader {
			continue
		}
		if found != -1 {
			return -1
		}
		found = i
	}
	return found
}

// resolveSkips resolves every still-unclaimed capture against the two
// review backstops (contract S10, Resolution R9). The tag log's
// basenames carry no uploader (readTagLogSkips), so a basename shared
// by two uploaders' captures is UnreconciledAmbiguousBasename rather
// than a guess; the review tree is scoped per uploader and never
// ambiguous on that account. A capture that resolves as a skip under
// an uploader partialSkipUploader flags is UnreconciledPartialSkip
// instead of a clean skip (Resolution R4). A capture whose uploader
// still carries an unclaimed gateRejectedRemaining count is Excluded
// rather than emitting a skip signal, but only when that capture
// resolved via the review tree and never appeared in the tag log:
// quarantineOne files a rejected album into the review tree by
// basename (_salvage/internal/review/importer.go:1014), and a
// quarantined or rejected album never reached beets, so it can never
// appear in a tag log. A tag-logged capture claims an ordinary skip
// instead, leaving the gate-rejected count for a review-only capture
// under the same uploader.
func resolveSkips(captures []Capture, claimed []bool, tagLogSkips map[string]bool, reviewed map[string][]string, partialSkipUploader map[string]bool, gateRejectedRemaining map[string]int, out *[]reconciliation) {
	tagLogCandidates := make(map[string][]int)
	for i, c := range captures {
		if claimed[i] {
			continue
		}
		if tagLogSkips[c.Basename] {
			tagLogCandidates[c.Basename] = append(tagLogCandidates[c.Basename], i)
		}
	}

	basenames := make([]string, 0, len(tagLogCandidates))
	for b := range tagLogCandidates {
		basenames = append(basenames, b)
	}
	slices.Sort(basenames)
	for _, basename := range basenames {
		idxs := tagLogCandidates[basename]
		if len(idxs) < 2 {
			continue
		}
		for _, i := range idxs {
			claimed[i] = true
			*out = append(*out, reconciliation{Capture: captures[i], Unreconciled: UnreconciledAmbiguousBasename})
		}
	}

	for i, c := range captures {
		if claimed[i] {
			continue
		}
		viaReview := reviewMatches(reviewed[c.Uploader], c.Basename)
		viaTagLog := len(tagLogCandidates[c.Basename]) == 1
		if !viaReview && !viaTagLog {
			continue
		}
		claimed[i] = true

		if viaReview && !tagLogSkips[c.Basename] && gateRejectedRemaining[c.Uploader] > 0 {
			gateRejectedRemaining[c.Uploader]--
			*out = append(*out, reconciliation{Capture: c, Excluded: ExcludedGateRejected})
			continue
		}

		if partialSkipUploader[c.Uploader] {
			*out = append(*out, reconciliation{Capture: c, Unreconciled: UnreconciledPartialSkip})
			continue
		}

		readFrom := "tag-log"
		if viaReview {
			readFrom = "review-tree"
		}
		*out = append(*out, reconciliation{
			Capture: c,
			Oracle: &OracleSignal{
				Signal:        "skip",
				Uploader:      c.Uploader,
				AlbumBasename: c.Basename,
				ReadFrom:      readFrom,
			},
		})
	}
}
