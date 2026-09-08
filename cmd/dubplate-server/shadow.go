package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/shadow"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// shadowFloorThreshold is the evidence floor rung 15 measures the
// shadow window against: a configurable duration with a 30-day default
// (build spec ruling 1). Provisional, owner ruling 4 open.
var shadowFloorThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_FLOOR",
	Kind:    config.KindDuration,
	Default: "720h",
	Doc:     "a Go duration, for example 720h",
}

// shadowBudgetThreshold caps one observe pass's MusicBrainz spend. An
// album costs one lookup on the ID-consensus early return
// (internal/engine/tag/tag.go:96-118) and, on the search path, one
// search plus SearchLimit lookups plus one pseudo-release fetch per
// candidate (internal/engine/musicbrainz/client.go:320-333), so the
// real per-album bound is eleven rather than seven. Sixty tokens at two
// seconds a token bounds a pass at two minutes.
// Provisional, owner ruling 5 open.
var shadowBudgetThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_BUDGET",
	Kind:    config.KindInt,
	Default: "60",
	Doc:     "an integer request count per pass",
}

// registerShadowThresholds runs once per process, from newRootCmd
// (Resolution R12): config.Register appends without dedup
// (internal/config/thresholds.go:40-42) and newRootCmd runs more than
// once in a test binary.
var registerShadowThresholds = sync.OnceFunc(func() {
	config.Register(shadowFloorThreshold)
	config.Register(shadowBudgetThreshold)
})

// caseJSON marshals one shadow fixture into cassette.Case's shape, with
// Expected nil: internal/engine/shadow may import no harness package
// (contract S19), so cmd/dubplate-server does the translation from the
// engine's own field names into the harness type. PathQuery is copied
// through verbatim rather than re-canonicalized: the fixture's own
// requests are already canonical (cassette's own rule), and a
// caseJSON that canonicalized on the way out would hide a fixture that
// stopped being canonical upstream.
func caseJSON(f shadow.CaseFixture) ([]byte, error) {
	items := make([]cassette.ItemFields, len(f.Items))
	for i, fields := range f.Items {
		items[i] = cassette.ItemFields(fields)
	}
	requests := make([]cassette.Request, len(f.Requests))
	for i, r := range f.Requests {
		requests[i] = cassette.Request{
			Method:    r.Method,
			PathQuery: r.PathQuery,
			Response:  r.Response,
		}
	}
	c := cassette.Case{
		ID:       f.ObservationID,
		Input:    cassette.Input{Items: items},
		Profile:  f.Profile,
		EvalDate: f.EvalDate,
		Requests: requests,
	}
	data, err := json.MarshalIndent(&c, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("case json %s: %w", f.ObservationID, err)
	}
	return data, nil
}

// pendingJSON marshals one shadow pending row into
// analysis.AdjudicationEntry's shape, setting Disposition and
// Adjudicator as constants rather than reading them off the row:
// shadow.PendingRow deliberately carries no Disposition field, since
// disposition is owner-only and every row the shadow writes starts
// pending by adjudicator agent. It validates the entry through
// analysis.CheckAdjudications before returning, so a malformed row
// never reaches disk.
func pendingJSON(row shadow.PendingRow) ([]byte, error) {
	entry := analysis.AdjudicationEntry{
		ID: row.ID,
		Subject: analysis.AdjudicationSubject{
			Album:       row.Album,
			ReleaseMBID: row.ReleaseMBID,
		},
		Suspected:         row.Suspected,
		Citation:          row.Citation,
		ReleaseTruth:      row.ReleaseTruth,
		Disposition:       analysis.Pending,
		Reason:            row.Reason,
		Adjudicator:       analysis.AgentAdjudicator,
		Date:              row.Date,
		MatcherCommit:     row.MatcherCommit,
		MatcherConfigHash: row.MatcherConfigHash,
	}

	ledger := &analysis.AdjudicationLedger{Entries: []analysis.AdjudicationEntry{entry}}
	if err := analysis.CheckAdjudications(nil, ledger); err != nil {
		return nil, fmt.Errorf("pending json %s: %w", row.ID, err)
	}

	data, err := json.MarshalIndent(&entry, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("pending json %s: %w", row.ID, err)
	}
	return data, nil
}

// storeArtifacts writes every fixture in res as an observation's
// case.json and every pending row as a staged pending row, the two
// harness-shaped artifacts contract S19 keeps out of
// internal/engine/shadow.
func storeArtifacts(s *shadow.Store, res *shadow.PassResult) error {
	for _, f := range res.Fixtures {
		data, err := caseJSON(f)
		if err != nil {
			return err
		}
		if err := s.PutCaseFile(f.ObservationID, data); err != nil {
			return err
		}
	}
	for _, row := range res.Pending {
		data, err := pendingJSON(row)
		if err != nil {
			return err
		}
		if err := s.PutPending(row.ID, data); err != nil {
			return err
		}
	}
	return nil
}
