package shadow

import "time"

// A Report is the launch-gate report's computed rows. It is computed
// here and PRINTED in cmd/dubplate-server, in the pinned "%-32s %s"
// row form doctor already uses (cmd/dubplate-server/doctor.go:64-66),
// so an operator reads one table shape across the binary.
//
// Two rows the spec's table names are deliberately absent, because the
// box cannot know them: divergences by owner disposition, and false
// confident matches net of oracle-wrong rulings. Owner dispositions
// live in corpus/analysis/adjudications.json on the workstation, which
// the box never reads. The report prints what the store knows, and the
// disposition join is the launch gate's own, run on the workstation
// against the committed corpus (Resolution R21).
type Report struct {
	ClockStart             time.Time
	ClockStarted           bool
	Floor                  time.Duration
	Elapsed                time.Duration
	ObservedApplied        int
	ObservedSkipped        int
	Excluded               map[string]int
	Unreconciled           map[string]int
	ReportGaps             int
	Deferred               map[string]int
	Divergences            map[string]int
	GoAppliesOracleSkipped int
	StagedPending          int
	Resolved               int
	MatcherCommits         map[string]int
	Passes                 []BudgetRecord
	BudgetTotal            int
	StoreBytes             int64
}

// BuildReport computes the report from the store alone. It reads no
// clock: now is the parameter the elapsed row is measured against, and
// this file calls time.Now nowhere, which is what
// TestEngineNeverReadsTheClock requires.
//
// Excluded, Unreconciled, Deferred, Divergences, report gaps and
// budget are summed across every stored pass record, because those are
// the shapes Run already accumulates per pass (run.go:106-116). The
// applied/skipped split and the per-commit divergence count are not:
// a pass record carries one matcher commit for its whole pass, but
// reevaluate restamps an individual observation's own commit
// (reevaluate.go), so those two rows are computed from the
// observations themselves.
func BuildReport(s *Store, now time.Time, floor time.Duration) (*Report, error) {
	r := &Report{
		Floor:          floor,
		Excluded:       map[string]int{},
		Unreconciled:   map[string]int{},
		Deferred:       map[string]int{},
		Divergences:    map[string]int{},
		MatcherCommits: map[string]int{},
	}

	clockStart, started, err := s.ClockStart()
	if err != nil {
		return nil, err
	}
	r.ClockStart = clockStart
	r.ClockStarted = started
	if started {
		r.Elapsed = now.Sub(clockStart)
	}

	passes, err := s.Passes()
	if err != nil {
		return nil, err
	}
	for _, p := range passes {
		for reason, n := range p.Excluded {
			r.Excluded[reason] += n
		}
		for reason, n := range p.Unreconciled {
			r.Unreconciled[reason] += n
		}
		for reason, n := range p.Deferred {
			r.Deferred[reason] += n
		}
		for class, n := range p.Divergences {
			r.Divergences[class] += n
		}
		r.ReportGaps += p.ReportGaps
		r.Passes = append(r.Passes, p.Budget)
		r.BudgetTotal += p.Budget.Spent
	}
	r.GoAppliesOracleSkipped = r.Divergences[ClassGoAppliesOracleSkipped]

	obs, err := s.Observations()
	if err != nil {
		return nil, err
	}
	for _, o := range obs {
		if o.Class == "" || o.Oracle == nil {
			continue
		}
		if o.Oracle.Signal == "apply" {
			r.ObservedApplied++
		} else {
			r.ObservedSkipped++
		}
		if o.Class != ClassAgreement {
			r.MatcherCommits[o.MatcherCommit]++
		}
	}

	pending, err := s.PendingIDs()
	if err != nil {
		return nil, err
	}
	r.StagedPending = len(pending)

	resolved, err := s.ResolvedIDs()
	if err != nil {
		return nil, err
	}
	r.Resolved = len(resolved)

	size, err := s.Size()
	if err != nil {
		return nil, err
	}
	r.StoreBytes = size

	return r, nil
}
