package conform

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// TestGradeCaseRequestsFiresOnAWrongLookupPath proves the zero-miss
// rule: a hand-authored fixture logs one release lookup at a path the
// real client never builds (a bogus inc= value), so the client's
// actual request goes unresolved and stub.Server records it as a
// [stub.Miss]. A gradeCaseRequests that swallowed the miss list would
// pass this fixture silently; this test is what catches that.
func TestGradeCaseRequestsFiresOnAWrongLookupPath(t *testing.T) {
	dir := filepath.Join("testdata", "request", "miss-fixture")
	_, err := gradeCaseRequests(context.Background(), dir)
	if err == nil {
		t.Fatal("expected gradeCaseRequests to fire on a fixture whose logged lookup path the client never builds")
	}
	if !strings.Contains(err.Error(), "unresolved request") {
		t.Errorf("error does not name the miss: %v", err)
	}
}

// TestDistinctReleaseLookupMBIDsDedupsRepeatedMBID proves mutation 5's
// underlying mechanism: a hand-authored log naming the same release
// mbid twice yields one distinct lookup to issue, while
// countLookupRequests still counts both log rows. That divergence is
// exactly what gradeCaseRequests' per-case issued-count assertion
// compares, so an extractor that silently dropped every mbid after the
// first would produce the same shape of mismatch this fixture proves
// the two helpers can disagree on.
func TestDistinctReleaseLookupMBIDsDedupsRepeatedMBID(t *testing.T) {
	c := &cassette.Case{
		Requests: []cassette.Request{
			{Method: "GET", PathQuery: releaseLookupPrefix + "aaaa?fmt=json&inc=x", Response: "a.json"},
			{Method: "GET", PathQuery: releaseLookupPrefix + "aaaa?fmt=json&inc=y", Response: "b.json"},
		},
	}

	mbids := distinctReleaseLookupMBIDs(c)
	if len(mbids) != 1 || mbids[0] != "aaaa" {
		t.Fatalf("distinctReleaseLookupMBIDs = %v, want exactly one distinct mbid", mbids)
	}
	if got := countLookupRequests(c); got != 2 {
		t.Fatalf("countLookupRequests = %d, want 2: the divergence from one distinct mbid is what the per-case issued-count assertion catches", got)
	}
}

// TestCountSearchRequestsIgnoresLookupRows proves the search/lookup
// discriminator: a release-lookup PathQuery ("/ws/2/release/<mbid>?...")
// never satisfies searchPathPrefix ("/ws/2/release?..."), so a case
// carrying only lookups counts zero searches rather than miscounting a
// lookup as a search.
func TestCountSearchRequestsIgnoresLookupRows(t *testing.T) {
	c := &cassette.Case{
		Requests: []cassette.Request{
			{Method: "GET", PathQuery: releaseLookupPrefix + "aaaa?fmt=json", Response: "a.json"},
		},
	}
	if got := countSearchRequests(c); got != 0 {
		t.Fatalf("countSearchRequests = %d, want 0: a lookup row must never count as a search", got)
	}
}

// TestRequestConstructionGradesCleanly is the pass's own instrument:
// the gate-resident sweep over the whole committed corpus.
func TestRequestConstructionGradesCleanly(t *testing.T) {
	root := repoRoot(t)

	start := time.Now()
	summary, err := GradeRequests(root)
	elapsed := time.Since(start)
	t.Logf("graded %d cases (%d lookups, %d searches, %d with no search) in %s",
		summary.Cases, summary.Lookups, summary.Searches, summary.CasesNoSearch, elapsed)

	if err != nil {
		t.Fatalf("GradeRequests: %v", err)
	}
	if summary.Cases < RequestCaseFloor {
		t.Errorf("graded %d cases, floor is %d: a case went missing", summary.Cases, RequestCaseFloor)
	}
	if summary.Lookups < RequestLookupFloor {
		t.Errorf("issued %d release lookups, floor is %d", summary.Lookups, RequestLookupFloor)
	}
	if summary.Searches < RequestSearchFloor {
		t.Errorf("issued %d searches, floor is %d", summary.Searches, RequestSearchFloor)
	}
	// The identity holds only because tag_album issues at most one
	// search per case, which is true of every committed case today
	// (measured: 164 at 0 searches, 3 at 0, 26+4+9+9 at exactly 1). A
	// case carrying two searches is itself the finding, so the message
	// says so rather than reading as an accounting bug (fold m24).
	if summary.Cases != summary.Searches+summary.CasesNoSearch {
		t.Errorf("case accounting does not close: %d graded, %d with a search, %d without; a case logging two searches would break this identity and is itself the finding",
			summary.Cases, summary.Searches, summary.CasesNoSearch)
	}
	// Reported, not asserted: no committed case carries an Extra entry
	// today, so there is no expectation to compare against. Consumed()'s
	// own firing proof lives in internal/harness/stub's testdata.
	for _, r := range summary.ExtraConsumed {
		t.Logf("extra channel answered %s %s from %s", r.Method, r.PathQuery, r.Response)
	}
}
