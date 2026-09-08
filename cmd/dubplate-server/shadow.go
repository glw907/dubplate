package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/shadow"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/spf13/cobra"
)

// shadowIntegrationClass is rung 4's integration: its own enable flag,
// its three required-when-enabled paths, and its two thresholds, all
// declared in rung 4's own file rather than internal/config
// (Resolution R5, re-expressing rung 4's own ruling 1 in registry
// terms per Resolution R14). registerShadowIntegration and
// registerShadowThresholds register it, and its two thresholds, onto
// the production registry from newRootCmd.
//
// ruled, conductor 2026-09-08, R5
var shadowIntegrationClass = config.IntegrationClass{
	Name: "shadow",
	Rung: 4,
	Flag: "DUBPLATE_SHADOW_ENABLED",
	Required: []string{
		"DUBPLATE_SHADOW_OBSERVE_DIR",      // the inbox the incumbent consumes
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR", // where the incumbent's run reports and tag logs live
		"DUBPLATE_SHADOW_REVIEW_DIR",       // where a skipped album is filed
	},
	Optional: []string{
		"DUBPLATE_SHADOW_FLOOR",  // threshold, duration, default 720h
		"DUBPLATE_SHADOW_BUDGET", // threshold, int, default 60
	},
}

// ShadowIntegration is the shadow integration's resolved paths, the
// three required-when-enabled rows of shadowIntegrationClass. The
// floor and the request cap are registered thresholds rather than
// fields here, because both carry defaults and shadowFromConfig
// carries no defaulted value.
type ShadowIntegration struct {
	ObserveDir     string
	OracleStateDir string
	ReviewDir      string
}

// shadowFromConfig builds the shadow integration's resolved paths from
// cfg's generic integration-resolution surface, or reports nil when
// DUBPLATE_SHADOW_ENABLED did not resolve true: every adopter who has
// not opted in.
func shadowFromConfig(cfg *config.Config) *ShadowIntegration {
	if !cfg.IntegrationEnabled(shadowIntegrationClass.Name) {
		return nil
	}
	return &ShadowIntegration{
		ObserveDir:     cfg.IntegrationValue(shadowIntegrationClass.Name, "DUBPLATE_SHADOW_OBSERVE_DIR"),
		OracleStateDir: cfg.IntegrationValue(shadowIntegrationClass.Name, "DUBPLATE_SHADOW_ORACLE_STATE_DIR"),
		ReviewDir:      cfg.IntegrationValue(shadowIntegrationClass.Name, "DUBPLATE_SHADOW_REVIEW_DIR"),
	}
}

// shadowFloorThreshold is the evidence floor rung 15 measures the
// shadow window against: a configurable duration with a 30-day default
// (build spec ruling 1). Ratified as owner ruling 4.
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
// Ratified as owner ruling 5.
var shadowBudgetThreshold = config.Threshold{
	Key:     "DUBPLATE_SHADOW_BUDGET",
	Kind:    config.KindInt,
	Default: "60",
	Doc:     "an integer request count per pass",
}

// registerShadowIntegration adds shadowIntegrationClass to reg,
// separately from its two thresholds (registerShadowThresholds): the
// class and the thresholds are two different registry families, and
// each keeps its own single-purpose registration function.
func registerShadowIntegration(reg *config.Registry) {
	reg.AddIntegrationClass(shadowIntegrationClass)
}

// registerShadowThresholds adds the shadow integration's two thresholds
// to reg. It is a plain function rather than the sync.OnceFunc rung 1
// needed (Resolution R12): newRootCmd builds a fresh, independent
// registry on every construction (Resolution R1), so registering onto
// it can never collide with a prior process's own registration the way
// appending to a shared package-level slice could.
//
// ruled, conductor 2026-09-07, R1
func registerShadowThresholds(reg *config.Registry) {
	reg.AddThreshold(shadowFloorThreshold)
	reg.AddThreshold(shadowBudgetThreshold)
}

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

// newShadowCmd builds the shadow subcommand and its three verbs:
// observe runs one capture, reconcile and match pass and is the timer's
// ExecStart; report prints the launch-gate report; reevaluate runs the
// mechanical re-evaluation ratified ruling B's adjudication-load
// reduction rests on.
//
// matcherCommit is a parameter rather than a package variable read
// inside RunE, on the same injection grounds environ already uses: the
// observe verb refuses an empty or "dev" stamp, and a test that cannot
// vary the stamp cannot prove the refusal. newRootCmd passes
// main.matcherCommit, so production behavior is unchanged and no test
// writes a global.
func newShadowCmd(environ []string, reg *config.Registry, matcherCommit string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "shadow",
		Short:         "Run the shadow matcher against the incumbent's own decisions",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetErr(stderr)
	cmd.AddCommand(
		newShadowObserveCmd(environ, reg, matcherCommit, stdout, stderr),
		newShadowReportCmd(environ, reg, stdout, stderr),
		newShadowReevaluateCmd(environ, reg, matcherCommit, stdout, stderr),
	)
	return cmd
}

// newShadowObserveCmd builds the observe verb. Its contract runs in
// order: load config; refuse when shadowFromConfig(cfg) is nil, naming
// DUBPLATE_SHADOW_ENABLED; refuse when matcherCommit is empty or "dev",
// creating no store either way; resolve the floor and the cap through
// the two registered thresholds; compose the User-Agent; open the
// store; call shadow.Run with time.Now(), legal here because cmd/ is
// outside the engine's no-clock walk; write the artifacts through
// storeArtifacts; print a one-line summary.
func newShadowObserveCmd(environ []string, reg *config.Registry, matcherCommit string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "observe",
		Short:         "Run one capture, reconcile and match pass",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			shadowCfg := shadowFromConfig(cfg)
			if shadowCfg == nil {
				return errors.New("shadow observe: DUBPLATE_SHADOW_ENABLED is not set")
			}
			if matcherCommit == "" || matcherCommit == "dev" {
				return fmt.Errorf("shadow observe: refusing to run on an unstamped matcher commit %q", matcherCommit)
			}

			if _, err := config.ParseDuration(environ, shadowFloorThreshold); err != nil {
				return err
			}
			budget, err := config.ParseInt(environ, shadowBudgetThreshold)
			if err != nil {
				return err
			}

			userAgent := fmt.Sprintf("dubplate-server/%s ( %s )", version, cfg.AdminEmail)

			store, err := shadow.OpenStore(shadow.StoreDir(cfg.StateDir))
			if err != nil {
				return err
			}

			res, err := shadow.Run(cmd.Context(), shadow.Config{
				StateDir:       cfg.StateDir,
				ObserveDir:     shadowCfg.ObserveDir,
				OracleStateDir: shadowCfg.OracleStateDir,
				ReviewDir:      shadowCfg.ReviewDir,
				UserAgent:      userAgent,
				MatcherCommit:  matcherCommit,
				BudgetLimit:    budget,
			}, time.Now())
			if err != nil {
				return err
			}

			if err := storeArtifacts(store, res); err != nil {
				return err
			}

			divergences := 0
			for _, n := range res.Pass.Divergences {
				divergences += n
			}
			_, err = fmt.Fprintf(stdout, "shadow observe: captured %d, observed %d (%d agreement(s), %d divergence(s))\n",
				res.Pass.Captured, res.Pass.Observed, res.Pass.Agreements, divergences)
			return err
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// newShadowReportCmd builds the report verb: it prints the launch-gate
// report in doctor's own "%-32s %s" row form, so an operator reads one
// table shape across the binary.
func newShadowReportCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "report",
		Short:         "Print the launch-gate report",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			floor, err := config.ParseDuration(environ, shadowFloorThreshold)
			if err != nil {
				return err
			}

			store, err := shadow.OpenStore(shadow.StoreDir(cfg.StateDir))
			if err != nil {
				return err
			}
			report, err := shadow.BuildReport(store, time.Now(), floor)
			if err != nil {
				return err
			}
			return printShadowReport(stdout, report)
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// newShadowReevaluateCmd builds the reevaluate verb: it reruns every
// observation stamped with a superseded matcher commit against its own
// stored responses, refusing to run on an unstamped build on the same
// grounds observe does (an unstamped commit can never be mechanically
// distinguished from a later one, so "reevaluate" against it would
// either restamp everything or nothing).
func newShadowReevaluateCmd(environ []string, reg *config.Registry, matcherCommit string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "reevaluate",
		Short:         "Re-run stamped observations against their stored responses",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			if matcherCommit == "" || matcherCommit == "dev" {
				return fmt.Errorf("shadow reevaluate: refusing to run on an unstamped matcher commit %q", matcherCommit)
			}

			store, err := shadow.OpenStore(shadow.StoreDir(cfg.StateDir))
			if err != nil {
				return err
			}

			now := time.Now()
			res, err := shadow.Reevaluate(cmd.Context(), shadow.Config{MatcherCommit: matcherCommit}, store, now)
			if err != nil {
				return err
			}
			if err := stageNewlyDiverged(store, res.NewlyDiverged, now); err != nil {
				return err
			}

			_, err = fmt.Fprintf(stdout,
				"shadow reevaluate: retired %d, restamped %d, newly diverged %d, unevaluable %d, skipped %d\n",
				len(res.Retired), len(res.Restamped), len(res.NewlyDiverged), len(res.Unevaluable), len(res.Skipped))
			return err
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// stageNewlyDiverged stages a pending row for every observation
// Reevaluate reports as newly diverged: an agreement that reevaluate
// promoted to a divergence was never staged in the first place, so this
// is the one outcome that needs a fresh row rather than a removal or an
// untouched one. internal/engine/shadow may import no harness package
// (contract S19), so the row's own construction (PendingRowFor) lives
// there and only the marshalling into analysis.AdjudicationEntry's
// shape happens here, exactly as storeArtifacts does for an observe
// pass's own pending rows.
func stageNewlyDiverged(s *shadow.Store, ids []string, now time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	obsList, err := s.Observations()
	if err != nil {
		return err
	}
	byID := make(map[string]*shadow.Observation, len(obsList))
	for _, obs := range obsList {
		byID[obs.ID] = obs
	}
	for _, id := range ids {
		obs, ok := byID[id]
		if !ok {
			return fmt.Errorf("shadow reevaluate: newly diverged observation %s not found", id)
		}
		row, ok := shadow.PendingRowFor(obs, now)
		if !ok {
			return fmt.Errorf("shadow reevaluate: observation %s reported newly diverged but carries no divergence", id)
		}
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

// printShadowReport writes r's rows in the pinned "%-32s %s" form,
// printing every map-valued row in sorted key order so the table is
// reproducible rather than map-iteration-ordered.
func printShadowReport(w io.Writer, r *shadow.Report) error {
	clockStart := "not started"
	if r.ClockStarted {
		clockStart = r.ClockStart.Format(time.RFC3339)
	}
	if err := printRow(w, "clock start", clockStart); err != nil {
		return err
	}
	if err := printRow(w, "floor", r.Floor.String()); err != nil {
		return err
	}
	pct := 0.0
	if r.Floor > 0 {
		pct = 100 * float64(r.Elapsed) / float64(r.Floor)
	}
	if err := printRow(w, "elapsed", fmt.Sprintf("%s (%.1f%% of floor)", r.Elapsed, pct)); err != nil {
		return err
	}
	if err := printRow(w, "observed albums", fmt.Sprintf("%d applied, %d skipped", r.ObservedApplied, r.ObservedSkipped)); err != nil {
		return err
	}

	if err := printCountRows(w, "excluded", r.Excluded); err != nil {
		return err
	}
	if err := printCountRows(w, "unreconciled", r.Unreconciled); err != nil {
		return err
	}
	if err := printRow(w, "report gaps", fmt.Sprintf("%d", r.ReportGaps)); err != nil {
		return err
	}
	if err := printCountRows(w, "deferred", r.Deferred); err != nil {
		return err
	}
	if err := printCountRows(w, "divergences", r.Divergences); err != nil {
		return err
	}
	if err := printRow(w, "go-applies-oracle-skipped", fmt.Sprintf("%d false confident matches (owner disposition joined off-box)", r.GoAppliesOracleSkipped)); err != nil {
		return err
	}
	if err := printRow(w, "divergences by disposition", "joined off-box against corpus/analysis/adjudications.json"); err != nil {
		return err
	}
	if err := printRow(w, "staged pending", fmt.Sprintf("%d", r.StagedPending)); err != nil {
		return err
	}
	if err := printRow(w, "resolved", fmt.Sprintf("%d", r.Resolved)); err != nil {
		return err
	}
	if err := printCountRows(w, "matcher commits", r.MatcherCommits); err != nil {
		return err
	}

	for i, p := range r.Passes {
		label := fmt.Sprintf("budget pass %d", i+1)
		if err := printRow(w, label, fmt.Sprintf("%s %d", p.Name, p.Spent)); err != nil {
			return err
		}
	}
	if err := printRow(w, "budget total", fmt.Sprintf("%d", r.BudgetTotal)); err != nil {
		return err
	}

	return printRow(w, "store size", fmt.Sprintf("%d bytes", r.StoreBytes))
}

// printCountRows prints one line per key in counts, sorted, each
// labeled "<label> <key>": the sorted order is what keeps a map-valued
// row reproducible across runs.
func printCountRows(w io.Writer, label string, counts map[string]int) error {
	for _, k := range slices.Sorted(maps.Keys(counts)) {
		if err := printRow(w, label+" "+k, fmt.Sprintf("%d", counts[k])); err != nil {
			return err
		}
	}
	return nil
}
