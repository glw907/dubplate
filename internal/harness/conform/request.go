package conform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// gradeUserAgent identifies the offline conformance client. The stub
// never reads a request header, so this value is a NewClient
// requirement rather than a graded fact; it is an invented,
// non-site-specific literal per CLAUDE.md's universality rule.
const gradeUserAgent = "dubplate-conform/0 ( offline-conformance-run )"

// Floors for the request census, measured off the tree at 2026-09-04.
// They are floors rather than exact counts because lane E1 adds cases:
// a shrink means a case or a request went missing, which is the failure
// worth catching, while growth is the corpus doing its job.
const (
	RequestCaseFloor   = 215
	RequestLookupFloor = 303
	RequestSearchFloor = 48
)

// searchPathPrefix is the WS/2 path a release search starts with; a
// release lookup's path always carries a trailing mbid segment before
// its own "?", so the two prefixes never collide (mirrors
// internal/harness/analysis's own releaseLookups discriminator,
// album_info.go:248-251).
const searchPathPrefix = "/ws/2/release?"

// RequestGradeSummary reports what one sweep actually compared, not what
// it visited.
type RequestGradeSummary struct {
	Cases         int
	Lookups       int
	Searches      int
	CasesNoSearch int
	ExtraConsumed []cassette.Request
}

// caseRequestResult is one case's contribution to a RequestGradeSummary,
// counting only requests actually issued (a failed client.Release call
// never increments Lookups), so a mutation that breaks every lookup is
// visible both as a Misses() report and as an issued-count shortfall.
type caseRequestResult struct {
	lookups       int
	searches      int
	extraConsumed []cassette.Request
}

// GradeRequests replays every committed case's request construction
// through the Go client against a stub built from that case's own log,
// and reports what it compared.
//
// It grades request CONSTRUCTION, never the request SET. The set is a
// function of the ID-match early return, _recommendation and
// _add_candidate, which are 3d's; this sweep reads each case's log to
// learn WHICH requests to issue and builds every issued request from the
// case's own items and this engine's own constants, so nothing about a
// request's bytes comes from the expectation it is graded against. 3d
// replaces the decision and inherits the stub, the zero-miss rule, and
// the client unchanged.
func GradeRequests(root string) (RequestGradeSummary, error) {
	casesRoot := filepath.Join(root, "corpus", "cases")
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		return RequestGradeSummary{}, fmt.Errorf("read corpus/cases: %w", err)
	}

	ctx := context.Background()
	var summary RequestGradeSummary
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		result, caseErr := gradeCaseRequests(ctx, filepath.Join(casesRoot, e.Name()))
		if caseErr != nil {
			// A case that failed contributes its errors and nothing
			// else: the counters below report what was compared, and a
			// case that could not be loaded, stubbed or fully replayed
			// was not.
			errs = append(errs, caseErr)
			continue
		}
		summary.Cases++
		summary.Lookups += result.lookups
		summary.Searches += result.searches
		if result.searches == 0 {
			summary.CasesNoSearch++
		}
		summary.ExtraConsumed = append(summary.ExtraConsumed, result.extraConsumed...)
	}

	return summary, errors.Join(errs...)
}

// gradeCaseRequests replays one case's request construction: a client
// built against a stub of the case's own log, one release lookup per
// distinct mbid the log names, and one search if and only if the log
// carries a search row. It keeps issuing every planned request even
// after one fails, so a systemic mutation (a wrong include set, say)
// reports its full miss count in one run rather than stopping at the
// first.
func gradeCaseRequests(ctx context.Context, dir string) (result caseRequestResult, err error) {
	var errs []error
	defer func() {
		err = errors.Join(errs...)
	}()

	c, loadErr := cassette.LoadCase(dir)
	if loadErr != nil {
		errs = append(errs, fmt.Errorf("case at %s: load: %w", dir, loadErr))
		return result, nil
	}

	srv, startErr := stub.New(c, dir)
	if startErr != nil {
		errs = append(errs, fmt.Errorf("case %s: start stub: %w", c.ID, startErr))
		return result, nil
	}
	defer func() {
		if closeErr := srv.Close(); closeErr != nil {
			errs = append(errs, fmt.Errorf("case %s: close stub: %w", c.ID, closeErr))
		}
	}()

	client, clientErr := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   "http://" + srv.Addr(),
		UserAgent: gradeUserAgent,
		Budget:    musicbrainz.Unmetered(),
	})
	if clientErr != nil {
		errs = append(errs, fmt.Errorf("case %s: build client: %w", c.ID, clientErr))
		return result, nil
	}

	tun, tunErr := mergeProfile(matcher.DefaultTunables(), c.Profile)
	if tunErr != nil {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, tunErr))
		return result, nil
	}

	mbids := distinctReleaseLookupMBIDs(c)
	for _, mbid := range mbids {
		if _, releaseErr := client.Release(ctx, mbid); releaseErr != nil {
			errs = append(errs, fmt.Errorf("case %s: release %s: %w", c.ID, mbid, releaseErr))
			continue
		}
		result.lookups++
	}
	if expected := countLookupRequests(c); result.lookups != expected {
		errs = append(errs, fmt.Errorf("case %s: issued %d release lookups, log carries %d", c.ID, result.lookups, expected))
	}

	expectedSearches := countSearchRequests(c)
	if expectedSearches > 0 {
		items := make([]matcher.ItemFields, len(c.Input.Items))
		for i, raw := range c.Input.Items {
			items[i] = matcher.ProjectItemFields(raw)
		}
		search := musicbrainz.AlbumSearchFor(items, tun)
		if _, searchErr := client.SearchReleases(ctx, search); searchErr != nil {
			// The search error is the whole failure. The issued-count
			// check below would restate it as a second error and inflate
			// the case's failure count.
			errs = append(errs, fmt.Errorf("case %s: search: %w", c.ID, searchErr))
		} else {
			result.searches++
			if result.searches != expectedSearches {
				errs = append(errs, fmt.Errorf("case %s: issued %d searches, log carries %d", c.ID, result.searches, expectedSearches))
			}
		}
	}

	for _, m := range srv.Misses() {
		errs = append(errs, fmt.Errorf("case %s: unresolved request %s %s", c.ID, m.Method, m.PathQuery))
	}
	result.extraConsumed = srv.Consumed()

	return result, nil
}

// distinctReleaseLookupMBIDs collects the release mbids c's request log
// fetches by id, in log order and with duplicates dropped: R1's seam
// only reads the log to learn WHICH requests to issue, never their
// bytes, so a case whose log names one mbid twice issues one lookup
// against it rather than two.
func distinctReleaseLookupMBIDs(c *cassette.Case) []string {
	seen := make(map[string]bool, len(c.Requests))
	var mbids []string
	for _, req := range c.Requests {
		rest, ok := strings.CutPrefix(req.PathQuery, releaseLookupPrefix)
		if !ok {
			continue
		}
		mbid, _, _ := strings.Cut(rest, "?")
		if mbid == "" || seen[mbid] {
			continue
		}
		seen[mbid] = true
		mbids = append(mbids, mbid)
	}
	return mbids
}

// countLookupRequests counts c's own logged release-lookup requests,
// undeduplicated, the anti-silent-skip baseline
// distinctReleaseLookupMBIDs' issued count is checked against.
func countLookupRequests(c *cassette.Case) int {
	n := 0
	for _, req := range c.Requests {
		if strings.HasPrefix(req.PathQuery, releaseLookupPrefix) {
			n++
		}
	}
	return n
}

// countSearchRequests counts c's own logged release-search requests.
func countSearchRequests(c *cassette.Case) int {
	n := 0
	for _, req := range c.Requests {
		if strings.HasPrefix(req.PathQuery, searchPathPrefix) {
			n++
		}
	}
	return n
}
