// Package importer orchestrates one dubplate import run: staging each
// contributor's quiescent inbox uploads, gating them for format and
// quality, running one beet import over the survivors, filing anything
// beets or the gate rejected, notifying contributors and the admin, and
// nudging Navidrome to rescan. It replaces scripts/music-import.
//
// Three outcomes govern one run, per the design spec's error doctrine: a
// per-run failure (a held lock aside) returns a wrapped error so main can
// exit nonzero and let systemd's OnFailure= chain page the admin; a
// per-item failure (one contributor's bad archive, one unmovable file) is
// recorded in the run report and never aborts the run; and a held lock is
// healthy, pinging success and returning nil so an overlapping run never
// looks like a failure.
package importer

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/archive"
	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/beets"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/gate"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ping"
	"github.com/glw907/dubplate/internal/reject"
	"github.com/glw907/dubplate/internal/review"
)

// ErrLowSpace marks a run aborted because the staging filesystem had less
// free space than [config.Thresholds.FreeFloorGB], checked before a
// single archive is extracted or a single file is staged.
var ErrLowSpace = errors.New("importer: free space below floor")

// ErrSessionNotFound marks a --session (or nudge-triggered) import whose
// session ID matches no directory under InboxDir: the trigger file
// outlived its promotion, or named a session that was never promoted.
var ErrSessionNotFound = errors.New("importer: session not found")

// ErrMissingSidecar marks a --session import target with no
// session.json: [Engine.promote] (internal/serve) always writes one
// before the atomic rename into the inbox, so its absence means the
// directory did not come from the uploader -- report the failure rather
// than guess at a collection.
var ErrMissingSidecar = errors.New("importer: session missing sidecar")

// pingSlug is the check name a run's success ping resolves through
// [config.Config.PingURLs], matching the shell layer's `ping_hc import`.
const pingSlug = "import"

// reportFilePrefix and reportFileSuffix bound one run's report filename,
// "report-<runid>.json" under [config.Paths.StateDir] -- one file per
// run, not one shared, overwritten file, so nudge-consume's
// session-scoped runs and the timer's poll runs never race each other's
// report the way the earlier fixed-name design did. [Importer.sweepReports]
// matches this same pattern to find every report file a run leaves
// behind. The journey NEVER reads these (plan-time decision): they are
// monitoring artifacts, and [ledger.Ledger] is the uploader pass's own
// source of truth.
const (
	reportFilePrefix = "report-"
	reportFileSuffix = ".json"
)

// sessionSidecarName is session.json's fixed name, matching
// internal/serve's [Engine.promote]: the collection's transport from
// serve into the importer, and the session key's durable home.
const sessionSidecarName = "session.json"

// gib is one gibibyte, matching the shell layer's df-derived GiB
// thresholds (avail_kb/1024/1024).
const gib = 1 << 30

// The FLAC-only quality floor: at least CD quality, matching
// scripts/lib.sh's quality_gate exactly.
const (
	cdMinSampleRate    = 44100
	cdMinBitsPerSample = 16
)

// ReportVersion is the run report's schema version, bumped whenever a
// field is added or its meaning changes. A consumer reading a report
// from disk checks this before trusting the shape.
//
// v2 (U8): the report gains RunID and each [AlbumOutcome] gains
// SessionID, and the report itself moved from one shared, overwritten
// file to per-run "report-<runid>.json" files.
const ReportVersion = 2

// Outcome is one contributor upload's disposition at the end of a run.
type Outcome string

const (
	// OutcomeImported is a beets-tagged upload now in the library.
	OutcomeImported Outcome = "imported"
	// OutcomeReview is an upload beets could not confidently auto-tag,
	// filed for a human to look at.
	OutcomeReview Outcome = "review"
	// OutcomeQuarantined is an upload that failed the quality gate (or a
	// beets skip) and could not be filed to the ordinary review tree, so
	// it landed in the quarantine fallback instead -- or, in the rare
	// case neither move succeeded, stayed in staging and may still have
	// been swept into the library by the beet import that followed, the
	// shell layer's own accepted residual risk.
	OutcomeQuarantined Outcome = "quarantined"
	// OutcomeRejected is an upload that failed the quality gate and was
	// filed to review with a friendly, per-file explanation emailed to
	// its uploader.
	OutcomeRejected Outcome = "rejected"
)

// AlbumOutcome is one contributor upload's provenance for a single run:
// where it came from, what happened to it, and -- once beets has renamed
// it into the library -- where it landed. LibraryPaths is populated only
// for [OutcomeImported], resolved from the beets database rather than
// inferred from the filesystem: the tag log records only skips, and
// beets renames every file it imports, so the database is the only
// surviving record of where an import landed.
type AlbumOutcome struct {
	Uploader  string  `json:"uploader"`
	SourceDir string  `json:"source_dir"`
	Outcome   Outcome `json:"outcome"`
	// SessionID identifies the uploader session this album came through,
	// read from its session.json sidecar (v2, U8). Empty for a leftover
	// staged before the uploader shipped, or a synthetic orphan unit
	// [Importer.deriveOrphanUnit] built from a bare staging-tree path --
	// neither carries a sidecar to read one from.
	SessionID    string   `json:"session_id,omitempty"`
	LibraryPaths []string `json:"library_paths,omitempty"`
	Collection   string   `json:"collection,omitempty"`
	// Error carries a per-item failure's detail (an unmovable file, a
	// beets-database match that never turned up). Outcome still reflects
	// the best available disposition; Error explains why it is uncertain.
	Error string `json:"error,omitempty"`
}

// RunReport is the versioned record of one import run, written to the
// state directory on every exit path -- including a held lock -- so the
// absence of a report is itself a monitoring event, not a journey state
// (spec: 2026-08-31-dubplate-uploader-design.md).
type RunReport struct {
	Version int `json:"version"`
	// RunID identifies this report's file, "report-<runid>.json" under
	// [config.Paths.StateDir] (v2, U8): not derived from StartedAt, since
	// several runs sharing a test's frozen clock must not collide on one
	// path.
	RunID     string    `json:"run_id"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	// LockHeld is true when this run exited immediately because another
	// run already held the import lock; every field below is zero.
	LockHeld bool `json:"lock_held"`
	// Error carries a per-run failure's message; empty on a clean run
	// (a held lock included -- that path is healthy, not an error).
	Error  string         `json:"error,omitempty"`
	Albums []AlbumOutcome `json:"albums"`
}

// Importer holds one run's dependencies: every collaborator package the
// earlier tasks built. Each is a concrete type with its own test seam;
// nothing here is an interface.
type Importer struct {
	Cfg    *config.Config
	Roster config.Roster
	Ping   *ping.Client
	Mailer *notify.Mailer
	Beets  *beets.Runner
	// Ledger is the contributions ledger the importer appends its
	// terminal events to (StateTagging, StateReview, StateRejected --
	// serve owns every other state, per the plan's one-appender-per-state
	// rule). A nil Ledger disables appending entirely rather than
	// panicking: every production caller sets it, and unit tests
	// exercising gate or beet logic alone need not build one.
	Ledger *ledger.Ledger

	// FlacPath, MetaflacPath, and SqlitePath are the gate's decode and
	// quality-probe binaries and the beets-database query tool; test
	// seams, production sets all three resolved on PATH.
	FlacPath     string
	MetaflacPath string
	SqlitePath   string

	// Now is the clock seam Run reads for the report timestamps, the
	// review filer's collision suffix, the quiescence cutoff, and the
	// beets-database "added since" window. Production sets it to
	// time.Now; tests pin it to a fixed instant.
	Now func() time.Time
}

// albumUnit is one contributor upload as staged: a single top-level
// entry under staging/<uploader>/, tracked from the moment it lands in
// staging through whatever the gate, beets, or a filing failure decides
// for it. It is dubplate's "album" for provenance purposes -- the same
// granularity scripts/music-import used for its own per-album rollups.
type albumUnit struct {
	uploader  string
	sourceDir string // the uploader's inbox directory
	stagedDir string // the entry's path under staging/<uploader>/

	// sessionID and email come from the unit's session.json sidecar, read
	// once its staged directory is settled and the sidecar removed from
	// disk -- readAndConsumeSidecar for the poll path's collectUnits,
	// stageSession for a --session-targeted run (which removes it only
	// after the move into staging has actually succeeded). Both are
	// empty for a unit with no sidecar: a leftover staged before the
	// uploader shipped, or a synthetic orphan from deriveOrphanUnit.
	sessionID string
	email     string

	result       Outcome
	collection   string
	libraryPaths []string
	errMsg       string
}

func (u *albumUnit) toReport() AlbumOutcome {
	return AlbumOutcome{
		Uploader:     u.uploader,
		SourceDir:    u.sourceDir,
		Outcome:      u.result,
		SessionID:    u.sessionID,
		LibraryPaths: u.libraryPaths,
		Collection:   u.collection,
		Error:        u.errMsg,
	}
}

// Run executes one poll-mode import run over every contributor's
// quiescent inbox uploads. It writes a [RunReport] to the state
// directory on every exit path via a deferred write, so a caller need
// only inspect the returned error to decide main's exit code -- the
// report itself is the operational record. The return is named because
// the deferred report write reads it.
func (im *Importer) Run(ctx context.Context) (runErr error) {
	l, ok, err := im.acquireOrRecordHeld(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer func() {
		if err := l.Release(); err != nil {
			slog.Error("importer: release import lock", "err", err)
		}
	}()

	report, finish := im.startRun()
	var units []*albumUnit
	defer func() { finish(units, &runErr) }()

	if err := im.prepareRunDirs(); err != nil {
		return err
	}

	dateStr := report.StartedAt.Format("2006-01-02")

	units, err = im.stageAllContributors(ctx)
	if err != nil {
		return err
	}

	return im.runPipeline(ctx, &units, im.defaultCollection(), dateStr)
}

// RunSession is `dubplate import --session <id>`'s entry point: it
// imports exactly sessionID's promoted directory, bypassing the
// ordinary per-contributor quiescence check entirely -- the caller
// already knows, from the CLI flag, exactly which directory is ready.
// It acquires the import lock the same non-blocking, held-is-healthy
// way [Run] does, then delegates to [Importer.ImportSessionLocked].
func (im *Importer) RunSession(ctx context.Context, sessionID string) error {
	l, ok, err := im.acquireOrRecordHeld(ctx)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	defer func() {
		if err := l.Release(); err != nil {
			slog.Error("importer: release import lock", "err", err)
		}
	}()
	return im.ImportSessionLocked(ctx, sessionID)
}

// ImportSessionLocked imports exactly sessionID's promoted directory,
// writing its own per-run report on every exit path the same way [Run]
// does. The caller must already hold the import lock: [RunSession]
// acquires it itself for a direct CLI invocation; `dubplate
// nudge-consume`'s own loop acquires it once via [lock.AcquireBlocking]
// for a whole batch of triggered sessions and calls this directly per
// session, never through RunSession -- a second flock(2) attempt from
// the same process on the same lock file would report itself as held
// and skip the work rather than doing it.
func (im *Importer) ImportSessionLocked(ctx context.Context, sessionID string) (runErr error) {
	report, finish := im.startRun()
	var units []*albumUnit
	defer func() { finish(units, &runErr) }()

	if err := im.prepareRunDirs(); err != nil {
		return err
	}

	srcDir, err := findSessionDir(im.Cfg.Paths.InboxDir, sessionID)
	if err != nil {
		return err
	}
	unit, err := im.stageSession(srcDir, sessionID)
	if err != nil {
		return err
	}
	units = []*albumUnit{unit}

	collection := unit.collection
	if collection == "" {
		collection = im.defaultCollection()
	}

	return im.runPipeline(ctx, &units, collection, report.StartedAt.Format("2006-01-02"))
}

// prepareRunDirs creates every directory a run writes into that might
// not exist yet, and checks free space -- shared by [Run] and
// [Importer.ImportSessionLocked], both of which need it before staging
// anything.
func (im *Importer) prepareRunDirs() error {
	for _, dir := range []string{im.Cfg.Paths.StagingDir, im.Cfg.Paths.ReviewDir, im.Cfg.Paths.StateDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("create directory %s: %w", dir, err)
		}
	}
	// Checked here, before a single archive is extracted or a single
	// file is staged: staging can itself consume the last of the free
	// space it is meant to guard.
	return im.checkFreeSpace()
}

// runPipeline runs the quality gate, the beet import (skipped entirely
// when nothing passed the gate), the ledger append for every unit this
// run settled, and the best-effort tail (empty-directory cleanup, the
// Navidrome scan trigger, the review/quarantine sweep, the run-report
// sweep, and the success ping) -- the shared body of [Run] and
// [Importer.ImportSessionLocked], which differ only in how units was
// staged and which collection this run's beet import applies.
func (im *Importer) runPipeline(ctx context.Context, units *[]*albumUnit, collection, dateStr string) error {
	if err := im.gateAndQuarantine(ctx, units, dateStr); err != nil {
		return err
	}

	if pending := pendingUnits(*units); len(pending) > 0 {
		if err := im.beetImportAndFileSkips(ctx, units, pending, dateStr, collection); err != nil {
			return err
		}
	}

	im.appendLedgerEvents(*units)

	if err := cleanEmptyStagingDirs(im.Cfg.Paths.StagingDir); err != nil {
		slog.Warn("importer: clean empty staging directories", "err", err)
	}

	// Best-effort margin: a failed scan trigger never fails the run.
	// Navidrome's own daily Scanner.Schedule backstop still runs.
	if err := beets.TriggerScan(ctx, im.Cfg.Navidrome.URL, im.Cfg.Navidrome.AdminUser, im.Cfg.Navidrome.AdminPassword); err != nil {
		slog.Warn("importer: trigger navidrome scan", "err", err)
	}

	sweeper := &review.Filer{ReviewDir: im.Cfg.Paths.ReviewDir, QuarantineDir: im.Cfg.Paths.QuarantineDir, Now: im.Now}
	if err := sweeper.Sweep(time.Duration(im.Cfg.Thresholds.RetentionDays) * 24 * time.Hour); err != nil {
		slog.Warn("importer: sweep review and quarantine", "err", err)
	}
	if err := im.sweepReports(im.Cfg.Thresholds.ReportRetention); err != nil {
		slog.Warn("importer: sweep run reports", "err", err)
	}

	im.pingSuccess(ctx)
	return nil
}

// acquireOrRecordHeld attempts the import lock. On [lock.ErrHeld] --
// another run, or an interactive `dubplate beet` session, already holds
// it -- that is a healthy overlap, not a failure (package doc's error
// doctrine): it writes a minimal LockHeld report itself, pings success,
// and returns ok false so the caller returns immediately without
// writing a second report of its own. Any other acquire error is
// returned unmodified. On success it returns the held lock and ok true;
// the caller releases it.
func (im *Importer) acquireOrRecordHeld(ctx context.Context) (l *lock.Lock, ok bool, err error) {
	l, err = lock.Acquire(lock.Path(im.Cfg.Paths.StateDir))
	if errors.Is(err, lock.ErrHeld) {
		report, finish := im.startRun()
		report.LockHeld = true
		var noErr error
		finish(nil, &noErr)
		im.pingSuccess(ctx)
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("acquire import lock: %w", err)
	}
	return l, true, nil
}

// startRun begins one run's report and returns a finish function the
// caller defers immediately: it stamps EndedAt, records runErr's
// message, appends every unit's outcome, and writes the versioned
// report to a fresh per-run path -- on every exit, a held lock and a
// per-item failure included, so the report's mere absence is itself a
// monitoring event.
func (im *Importer) startRun() (report *RunReport, finish func(units []*albumUnit, runErr *error)) {
	runID, err := newRunID()
	if err != nil {
		// crypto/rand failing is a host-level fault no report can
		// usefully record; a fixed fallback keeps the run itself going
		// rather than aborting an import over a report filename.
		slog.Error("importer: generate run id", "err", err)
		runID = "unknown"
	}
	report = &RunReport{Version: ReportVersion, RunID: runID, StartedAt: im.Now()}
	finish = func(units []*albumUnit, runErr *error) {
		report.EndedAt = im.Now()
		if *runErr != nil {
			report.Error = (*runErr).Error()
		}
		for _, u := range units {
			report.Albums = append(report.Albums, u.toReport())
		}
		if err := im.writeReport(report); err != nil {
			slog.Error("importer: write run report", "err", err)
		}
	}
	return report, finish
}

// newRunID returns a fresh, filename-safe identifier for one run's
// report. It is random rather than time-derived: several runs sharing a
// test's frozen [Importer.Now] clock -- the ten-run history-survival
// test included -- must never collide on the same report path.
func newRunID() (string, error) {
	var raw [8]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return "", fmt.Errorf("generate run id: %w", err)
	}
	return hex.EncodeToString(raw[:]), nil
}

// checkFreeSpace aborts the run with a wrapped [ErrLowSpace] when the
// staging filesystem has less free space than the configured floor.
func (im *Importer) checkFreeSpace() error {
	avail, err := archive.FreeSpace(im.Cfg.Paths.StagingDir)
	if err != nil {
		return fmt.Errorf("check free space on %s: %w", im.Cfg.Paths.StagingDir, err)
	}
	floor := int64(im.Cfg.Thresholds.FreeFloorGB) * gib
	if avail < floor {
		return fmt.Errorf("%s has %d bytes free, below the %d byte floor: %w", im.Cfg.Paths.StagingDir, avail, floor, ErrLowSpace)
	}
	return nil
}

// archiveLimits builds the [archive.Limits] every extraction in this run
// shares, from the same thresholds [Importer.checkFreeSpace] enforces at
// the run level.
func (im *Importer) archiveLimits() archive.Limits {
	return archive.Limits{
		MaxExtractedBytes: int64(im.Cfg.Thresholds.ArchiveCapGB) * gib,
		MinFreeBytes:      int64(im.Cfg.Thresholds.FreeFloorGB) * gib,
	}
}

// pingSuccess reports this run's success. A ping failure is best-effort
// margin: logged, never returned, per the doctrine that monitoring must
// never take down the pipeline it monitors.
func (im *Importer) pingSuccess(ctx context.Context) {
	if err := im.Ping.Success(ctx, im.Cfg.PingURLs[pingSlug]); err != nil {
		slog.Warn("importer: success ping", "err", err)
	}
}

// filerFor builds the [review.Filer] one contributor's filing (a
// quality-gate failure or a beets skip) files into on dateStr.
func (im *Importer) filerFor(dateStr, uploader string) *review.Filer {
	return &review.Filer{
		ReviewDir:     filepath.Join(im.Cfg.Paths.ReviewDir, dateStr, uploader),
		QuarantineDir: filepath.Join(im.Cfg.Paths.QuarantineDir, dateStr, uploader),
		Now:           im.Now,
	}
}

// reportPath returns runID's report file path under the state directory.
func (im *Importer) reportPath(runID string) string {
	return filepath.Join(im.Cfg.Paths.StateDir, reportFilePrefix+runID+reportFileSuffix)
}

// writeReport marshals report as versioned JSON and writes it atomically
// to its own per-run path (report.RunID): every run leaves its own
// file behind rather than overwriting the last one, held-lock runs
// included, so nudge-consume's session-scoped runs and the timer's poll
// runs never race each other's report. [Importer.sweepReports] is what
// keeps these from accumulating forever.
func (im *Importer) writeReport(report *RunReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal run report: %w", err)
	}
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		return fmt.Errorf("create state directory %s: %w", im.Cfg.Paths.StateDir, err)
	}
	path := im.reportPath(report.RunID)
	if err := atomicfile.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write run report %s: %w", path, err)
	}
	return nil
}

// sweepReports keeps at most retention of the newest report-*.json
// files under the state directory, deleting the rest by ascending
// modification time -- so an unattended box never accumulates one
// report file per timer tick, per nudge-consume session, forever. A
// non-positive retention is treated as "keep everything": Load never
// defaults it below 1, but a caller-supplied 0 must not delete every
// report just written.
func (im *Importer) sweepReports(retention int) error {
	if retention <= 0 {
		return nil
	}
	matches, err := filepath.Glob(filepath.Join(im.Cfg.Paths.StateDir, reportFilePrefix+"*"+reportFileSuffix))
	if err != nil {
		return fmt.Errorf("list run reports: %w", err)
	}
	if len(matches) <= retention {
		return nil
	}

	type reportFile struct {
		path    string
		modTime time.Time
	}
	files := make([]reportFile, 0, len(matches))
	for _, path := range matches {
		info, err := os.Stat(path)
		if err != nil {
			slog.Warn("importer: stat run report during sweep", "path", path, "err", err)
			continue
		}
		files = append(files, reportFile{path: path, modTime: info.ModTime()})
	}
	slices.SortFunc(files, func(a, b reportFile) int { return a.modTime.Compare(b.modTime) })

	excess := len(files) - retention
	for _, f := range files[:excess] {
		if err := os.Remove(f.path); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Warn("importer: remove stale run report", "path", f.path, "err", err)
		}
	}
	return nil
}

// defaultCollection returns [config.Config.Collections]'s first entry,
// the poll path's beet-import collection when nothing more specific
// applies, or "" when no collection is configured at all (an empty
// collection omits --set entirely -- see [beets.Runner.Import]).
func (im *Importer) defaultCollection() string {
	if len(im.Cfg.Collections) > 0 {
		return im.Cfg.Collections[0]
	}
	return ""
}

// appendLedgerEvents appends one terminal ledger event for every unit
// this run settled to imported, review, or rejected -- the vocabulary
// the importer owns (plan's one-appender-per-state rule; Quarantined is
// an infra-failure edge case with no ledger event, since it carries no
// confident state to report). A unit with no session.json sidecar
// (sessionID empty -- a leftover staged before the uploader shipped, or
// a synthetic orphan) is skipped: there is no session to record
// against. A nil [Importer.Ledger] disables this entirely.
func (im *Importer) appendLedgerEvents(units []*albumUnit) {
	if im.Ledger == nil {
		return
	}
	now := im.Now()
	for _, u := range units {
		if u.sessionID == "" {
			continue
		}
		var state ledger.State
		switch u.result {
		case OutcomeImported:
			state = ledger.StateTagging
		case OutcomeReview:
			state = ledger.StateReview
		case OutcomeRejected:
			state = ledger.StateRejected
		default:
			continue
		}
		event := ledger.Event{
			SessionID:    u.sessionID,
			Email:        u.email,
			Album:        albumFromStagedDir(u.stagedDir),
			Collection:   u.collection,
			State:        state,
			At:           now,
			LibraryPaths: u.libraryPaths,
		}
		if err := im.Ledger.Append(event); err != nil {
			slog.Error("importer: append ledger event", "session_id", u.sessionID, "state", state, "err", err)
		}
	}
}

// albumFromStagedDir returns the ledger [ledger.Event.Album] value for a
// unit staged under "<sessionid>__<album>": the part after the first
// "__", matching exactly how [Engine.promote] (internal/serve) names
// the directory. A name with no "__" (should not happen for a unit that
// carries a sessionID) falls back to the whole base name rather than an
// empty string.
func albumFromStagedDir(stagedDir string) string {
	base := filepath.Base(stagedDir)
	if _, album, ok := strings.Cut(base, "__"); ok {
		return album
	}
	return base
}

// readSidecar reads dir's session.json, if present, without removing
// it. A missing sidecar is not an error -- a leftover staged before the
// uploader shipped, or a directory [Importer.deriveOrphanUnit] derived
// synthetically, carries none.
func readSidecar(dir string) (sc sessionSidecar, ok bool, err error) {
	path := filepath.Join(dir, sessionSidecarName)
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return sessionSidecar{}, false, nil
	}
	if err != nil {
		return sessionSidecar{}, false, fmt.Errorf("read session sidecar %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &sc); err != nil {
		return sessionSidecar{}, false, fmt.Errorf("parse session sidecar %s: %w", path, err)
	}
	return sc, true, nil
}

// readAndConsumeSidecar reads dir's session.json, if present, and
// removes it: its only job is carrying the session's identity and
// collection from serve into the importer, and it must not survive to
// confuse [dirIsEmptyOrGone]'s "everything beets moved out" check or
// beets' own directory walk. Only for a directory already staged in
// place ([collectUnits]): a directory still mid-move must not have its
// sidecar consumed before the move can be retried -- see [stageSession].
func readAndConsumeSidecar(dir string) (sc sessionSidecar, ok bool, err error) {
	sc, ok, err = readSidecar(dir)
	if err != nil || !ok {
		return sc, ok, err
	}
	path := filepath.Join(dir, sessionSidecarName)
	if err := os.Remove(path); err != nil {
		return sessionSidecar{}, false, fmt.Errorf("remove consumed session sidecar %s: %w", path, err)
	}
	return sc, true, nil
}

// sessionSidecar mirrors internal/serve's on-disk session.json shape
// field for field: the collection's transport from serve into the
// importer, and the session key's durable home.
type sessionSidecar struct {
	Version    int       `json:"version"`
	SessionID  string    `json:"session_id"`
	Email      string    `json:"email"`
	DirName    string    `json:"dir_name"`
	Album      string    `json:"album"`
	Collection string    `json:"collection"`
	CreatedAt  time.Time `json:"created_at"`
}

// findSessionDir returns the promoted directory for sessionID under
// inboxDir: [Engine.promote] (internal/serve) names it
// "<dirname>/<sessionid>__<album>", one directory per contributor
// beneath inboxDir, so every contributor's top level is scanned for a
// prefix match.
func findSessionDir(inboxDir, sessionID string) (string, error) {
	contributors, err := os.ReadDir(inboxDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
		}
		return "", fmt.Errorf("read inbox %s: %w", inboxDir, err)
	}

	prefix := sessionID + "__"
	for _, c := range contributors {
		if !c.IsDir() {
			continue
		}
		userDir := filepath.Join(inboxDir, c.Name())
		entries, err := os.ReadDir(userDir)
		if err != nil {
			return "", fmt.Errorf("read inbox directory %s: %w", userDir, err)
		}
		for _, e := range entries {
			if e.IsDir() && strings.HasPrefix(e.Name(), prefix) {
				return filepath.Join(userDir, e.Name()), nil
			}
		}
	}
	return "", fmt.Errorf("%w: %s", ErrSessionNotFound, sessionID)
}

// stageSession moves srcDir -- the entry [findSessionDir] located --
// from the inbox into staging and returns the resulting [albumUnit].
// The sidecar is only read here, then moved along with the rest of
// srcDir by [review.Move] and removed from its destination once the
// move has actually succeeded: consuming it before the move (as
// [readAndConsumeSidecar] does for a directory already staged in
// place) would leave a failed move's srcDir permanently unrecoverable
// -- present in the inbox but missing the sidecar every retry needs.
// Unlike stageUser, a promoted session directory is never an archive:
// [Engine.promote] (internal/serve) always renames a real,
// already-assembled directory into place, so no extraction step
// applies here. A missing sidecar is [ErrMissingSidecar]: every
// session [Engine.promote] creates carries one, so its absence means
// srcDir did not come from the uploader.
func (im *Importer) stageSession(srcDir, sessionID string) (*albumUnit, error) {
	sc, ok, err := readSidecar(srcDir)
	if err != nil {
		return nil, fmt.Errorf("read session sidecar for %s: %w", sessionID, err)
	}
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrMissingSidecar, sessionID)
	}

	destDir := filepath.Join(im.Cfg.Paths.StagingDir, sc.DirName)
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return nil, fmt.Errorf("create staging directory %s: %w", destDir, err)
	}
	dest := filepath.Join(destDir, filepath.Base(srcDir))
	if err := review.Move(srcDir, dest); err != nil {
		return nil, fmt.Errorf("stage session %s: %w", sessionID, err)
	}
	sidecarPath := filepath.Join(dest, sessionSidecarName)
	if err := os.Remove(sidecarPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("remove consumed session sidecar %s: %w", sidecarPath, err)
	}

	return &albumUnit{
		uploader:   sc.DirName,
		sourceDir:  filepath.Dir(srcDir),
		stagedDir:  dest,
		sessionID:  sc.SessionID,
		email:      sc.Email,
		collection: sc.Collection,
	}, nil
}

// --- staging ----------------------------------------------------------

// stageAllContributors walks every top-level entry of the inbox,
// quiescence-checks and stages each contributor's uploads in turn, and
// returns one [albumUnit] per staged top-level entry across every
// contributor. A contributor whose inbox is still changing, or whose
// staging hits an isolated error, is skipped for this run -- logged, not
// fatal -- and simply picked up again on a later run.
func (im *Importer) stageAllContributors(ctx context.Context) ([]*albumUnit, error) {
	contributors, err := listContributorDirs(im.Cfg.Paths.InboxDir)
	if err != nil {
		return nil, fmt.Errorf("read inbox %s: %w", im.Cfg.Paths.InboxDir, err)
	}

	var units []*albumUnit
	for _, user := range contributors {
		inboxUserDir := filepath.Join(im.Cfg.Paths.InboxDir, user)

		cutoff := im.Now().Add(-time.Duration(im.Cfg.Thresholds.QuiescentMin) * time.Minute)
		quiescent, err := quiescentSince(inboxUserDir, cutoff)
		if err != nil {
			slog.Warn("importer: check quiescence", "uploader", user, "err", err)
			continue
		}
		if !quiescent {
			slog.Info("importer: inbox still changing; skipping this run", "uploader", user)
			continue
		}

		hasEntries, err := dirHasEntries(inboxUserDir)
		if err != nil {
			slog.Warn("importer: read inbox", "uploader", user, "err", err)
			continue
		}
		if !hasEntries {
			continue
		}

		stagingUserDir := filepath.Join(im.Cfg.Paths.StagingDir, user)
		if err := im.stageUser(ctx, user, inboxUserDir, stagingUserDir); err != nil {
			slog.Error("importer: stage uploads", "uploader", user, "err", err)
			continue
		}

		userUnits, err := collectUnits(user, inboxUserDir, stagingUserDir)
		if err != nil {
			slog.Error("importer: collect staged units", "uploader", user, "err", err)
			continue
		}
		units = append(units, userUnits...)
	}
	return units, nil
}

// stageUser moves one contributor's quiescent inbox entries into
// destDir. An archive entry is first expanded into its own private
// scratch subdirectory and only then has its top-level members moved
// into destDir under a collision-safe name, so two archives in the same
// batch can never step on each other. A rejected or unmovable entry is
// logged and left for a later run -- one contributor's bad upload never
// aborts another's.
func (im *Importer) stageUser(ctx context.Context, user, srcDir, destDir string) error {
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("create staging directory %s: %w", destDir, err)
	}
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return fmt.Errorf("read inbox directory %s: %w", srcDir, err)
	}

	for _, e := range entries {
		name := e.Name()
		src := filepath.Join(srcDir, name)

		if !e.IsDir() && isArchiveName(name) {
			if err := im.stageArchive(ctx, src, destDir); err != nil {
				slog.Warn("importer: rejected archive; left in inbox for review", "uploader", user, "archive", src, "err", err)
			}
			continue
		}
		if err := review.Move(src, filepath.Join(destDir, name)); err != nil {
			slog.Warn("importer: could not stage upload", "uploader", user, "entry", src, "err", err)
		}
	}
	return nil
}

// stageArchive extracts archivePath into its own scratch subdirectory of
// destDir, then moves each extracted top-level entry into destDir under
// a collision-safe name and removes the archive from the inbox. The
// scratch directory is always removed, whether extraction succeeded or
// not.
func (im *Importer) stageArchive(ctx context.Context, archivePath, destDir string) error {
	scratch, err := os.MkdirTemp(destDir, ".extract-*")
	if err != nil {
		return fmt.Errorf("create extraction scratch directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	if err := archive.Extract(ctx, archivePath, scratch, im.archiveLimits()); err != nil {
		return err
	}

	entries, err := os.ReadDir(scratch)
	if err != nil {
		return fmt.Errorf("read extracted %s: %w", scratch, err)
	}
	for _, e := range entries {
		src := filepath.Join(scratch, e.Name())
		dst := stagingDestPath(destDir, e.Name())
		if err := review.Move(src, dst); err != nil {
			return fmt.Errorf("move extracted entry %s: %w", src, err)
		}
	}
	return os.Remove(archivePath)
}

// collectUnits builds one [albumUnit] per top-level entry now staged
// under stagingUserDir, reading and consuming each directory entry's
// session.json sidecar (if present -- a loose file entry, from before
// the uploader shipped, carries none and is left as-is).
func collectUnits(user, inboxUserDir, stagingUserDir string) ([]*albumUnit, error) {
	entries, err := os.ReadDir(stagingUserDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read staged directory %s: %w", stagingUserDir, err)
	}
	units := make([]*albumUnit, 0, len(entries))
	for _, e := range entries {
		stagedDir := filepath.Join(stagingUserDir, e.Name())
		u := &albumUnit{
			uploader:  user,
			sourceDir: inboxUserDir,
			stagedDir: stagedDir,
		}
		if e.IsDir() {
			sc, ok, err := readAndConsumeSidecar(stagedDir)
			if err != nil {
				slog.Warn("importer: read session sidecar", "staged_dir", stagedDir, "err", err)
			} else if ok {
				u.sessionID = sc.SessionID
				u.email = sc.Email
				u.collection = sc.Collection
			}
		}
		units = append(units, u)
	}
	return units, nil
}

// --- quality gate -------------------------------------------------------

// gateAndQuarantine runs the quality gate over the whole staging tree
// and files every failing [albumUnit] into review (or quarantine, on a
// review-filing failure), emailing a friendly per-album explanation to
// each matched uploader and one technical summary to the admin. Only a
// gate infrastructure failure (the walk itself erroring) returns an
// error; an individual album's rejection is per-item and never aborts
// the run. units is a pointer because a verdict matching no unit staged
// this run -- a leftover album, per [Importer.deriveOrphanUnit] -- is
// appended to it so the run report still carries the album's outcome.
func (im *Importer) gateAndQuarantine(ctx context.Context, units *[]*albumUnit, dateStr string) error {
	policy := gate.Policy{
		AllowedExt:       []string{".flac"},
		RequireDecode:    true,
		MinSampleRate:    cdMinSampleRate,
		MinBitsPerSample: cdMinBitsPerSample,
		MetaflacPath:     im.MetaflacPath,
	}
	verdicts, err := gate.Check(ctx, im.Cfg.Paths.StagingDir, policy, im.FlacPath)
	if err != nil {
		return fmt.Errorf("run quality gate: %w", err)
	}

	failures := make(map[*albumUnit][]gate.Verdict)
	var order []*albumUnit
	for _, v := range verdicts {
		if v.OK {
			continue
		}
		u := findUnit(*units, v.Path)
		if u == nil {
			orphan, ok := im.deriveOrphanUnit(v.Path)
			if !ok {
				slog.Warn("importer: quality-gate verdict matched no staged upload", "path", v.Path)
				continue
			}
			*units = append(*units, orphan)
			u = orphan
		}
		if _, seen := failures[u]; !seen {
			order = append(order, u)
		}
		failures[u] = append(failures[u], v)
	}
	if len(order) == 0 {
		return nil
	}

	var adminSummary strings.Builder
	for _, u := range order {
		im.quarantineOne(ctx, u, failures[u], dateStr, &adminSummary)
	}

	subject := fmt.Sprintf("Dubplate: import quality gate failures (%s)", dateStr)
	body := "The following uploads failed the quality gate today (see each item for where it landed):\n\n" + adminSummary.String()
	if err := im.Mailer.Send(ctx, im.Cfg.AdminEmail, subject, body); err != nil {
		slog.Warn("importer: send admin quality-gate summary", "err", err)
	}
	return nil
}

// quarantineOne files one quality-gate-failed unit and appends its
// contribution to the admin's technical summary.
func (im *Importer) quarantineOne(ctx context.Context, u *albumUnit, failed []gate.Verdict, dateStr string, adminSummary *strings.Builder) {
	albumKey := u.uploader + "/" + filepath.Base(u.stagedDir)

	dest, err := im.filerFor(dateStr, u.uploader).MoveToReview(u.stagedDir)
	if err != nil {
		// Geoff's summary in the shell layer's own words: this must never
		// go silent, since neither move succeeding is exactly the case
		// that most needs a human's attention -- the album is still in
		// staging and the beet import that follows may ingest it anyway.
		u.result = OutcomeQuarantined
		u.errMsg = err.Error()
		fmt.Fprintf(adminSummary, "%s: COULD NOT MOVE BY ANY MEANS -- still in staging and beet import may ingest it; manual removal required: %v\n", albumKey, err)
		return
	}
	if strings.HasPrefix(dest, im.Cfg.Paths.QuarantineDir) {
		// The review move itself failed and only the quarantine fallback
		// landed it. The friendly per-file detail and reject email below
		// are for albums actually filed to review, so this album gets
		// only the shell's own one-line note -- Geoff's promised "see
		// each item for where it landed" -- and nothing else.
		u.result = OutcomeQuarantined
		fmt.Fprintf(adminSummary, "%s: could not move; excluded from import (see %s)\n", albumKey, dest)
		return
	}
	u.result = OutcomeRejected

	var rejections []reject.Rejection
	var rawLines strings.Builder
	for _, v := range failed {
		rejections = append(rejections, reject.Rejection{Path: v.Path, Reason: v.Reason})
		fmt.Fprintf(&rawLines, "%s\t%s\n", v.Path, v.Reason.Friendly())
	}
	fmt.Fprintf(adminSummary, "%s:\n%s", albumKey, rawLines.String())

	email, hasEmail := im.Roster.EmailFor(u.uploader)
	firstName, hasName := im.Roster.FirstName(u.uploader)
	switch {
	case !hasEmail:
		slog.Info("importer: no email on file for uploader; skipping their quality-gate notice", "uploader", u.uploader)
	case !hasName:
		slog.Info("importer: no contributor name on file for uploader; skipping their quality-gate notice", "uploader", u.uploader)
	default:
		body, err := reject.RenderEmail(im.Cfg.RejectEmailTemplate, firstName, filepath.Base(u.stagedDir), im.Cfg.AdminEmail, rejections)
		if err != nil {
			slog.Warn("importer: render reject email", "uploader", u.uploader, "err", err)
			return
		}
		if err := im.Mailer.Send(ctx, email, "Your music upload needs a different format", body); err != nil {
			slog.Warn("importer: send reject email", "uploader", u.uploader, "err", err)
		}
	}
}

// deriveOrphanUnit builds a synthetic [albumUnit] for a quality-gate
// verdict or beets skip whose path falls under StagingDir but matches no
// unit staged this run -- a leftover from an earlier or aborted run, or a
// contributor this run skipped for non-quiescence. music-import at HEAD
// derives the same album key from the staging-relative path
// (rel=${file#"$staging_root"/}, user=${rel%%/*}) and files it
// regardless of provenance: staging can hold leftovers, and beet import
// scans the whole staging tree, so a leftover that fails the gate must
// still be filed or a later run's beet import may ingest it. It reports
// ok false only when path is not under StagingDir at all, the one case
// that stays a bare warning.
func (im *Importer) deriveOrphanUnit(path string) (unit *albumUnit, ok bool) {
	rel, err := filepath.Rel(im.Cfg.Paths.StagingDir, path)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, false
	}
	parts := strings.SplitN(rel, string(filepath.Separator), 3)
	if len(parts) < 2 {
		return nil, false
	}
	uploader, album := parts[0], parts[1]
	return &albumUnit{
		uploader:  uploader,
		sourceDir: filepath.Join(im.Cfg.Paths.InboxDir, uploader),
		stagedDir: filepath.Join(im.Cfg.Paths.StagingDir, uploader, album),
	}, true
}

// --- beet import and skip handling --------------------------------------

// pendingUnits returns the units the quality gate left untouched: those
// still without a result, which is exactly what remains inside the
// staging tree for beet import to see.
func pendingUnits(units []*albumUnit) []*albumUnit {
	var pending []*albumUnit
	for _, u := range units {
		if u.result == "" {
			pending = append(pending, u)
		}
	}
	return pending
}

// beetImportAndFileSkips runs one beet import over the whole staging
// tree (which, after gateAndQuarantine, holds exactly the pending
// units) with collection set via beets' "--set collection=" flexible
// attribute ([beets.Runner.Import]'s doc explains the routing), files
// every skipped path the way a quality-gate failure is filed, and
// resolves each remaining pending unit's outcome: imported (its staged
// directory is now empty or gone) or left behind in review. A beet
// import subprocess failure is a per-run failure -- the shell layer's
// own `beet import` call runs unguarded under `set -e`, so this
// preserves that behavior rather than degrading it to a per-item one.
// units is the same pointer gateAndQuarantine received, so a skip
// matching no pending unit can still be appended for the run report.
func (im *Importer) beetImportAndFileSkips(ctx context.Context, units *[]*albumUnit, pending []*albumUnit, dateStr, collection string) error {
	preImport := im.Now()
	result, err := im.Beets.Import(ctx, im.Cfg.Paths.StagingDir, collection)
	postImport := im.Now()
	if err != nil {
		return fmt.Errorf("beet import %s: %w", im.Cfg.Paths.StagingDir, err)
	}

	skipUsers := im.fileSkips(units, pending, result.Skipped, dateStr)
	if len(skipUsers) > 0 {
		im.notifySkips(ctx, skipUsers, dateStr)
	}

	imported := im.resolveImported(pending, collection)
	if len(imported) == 0 {
		return nil
	}

	since := float64(preImport.UnixNano()) / 1e9
	until := float64(postImport.UnixNano()) / 1e9
	groups, err := beets.ResolveLibraryPaths(ctx, im.SqlitePath, im.Cfg.BeetsDB, since, until)
	if err != nil {
		slog.Error("importer: resolve library paths", "err", err)
		return nil
	}
	// Positional match, in staging order, against the database's album
	// order: one shared beet invocation carries no per-uploader join key,
	// so this is the best correlation available without splitting the
	// import per contributor (which would itself break "one beet import"
	// parity with the shell layer). A mismatched count leaves the extra
	// units' LibraryPaths empty rather than guessing.
	for i, u := range imported {
		if i < len(groups) {
			u.libraryPaths = groups[i].Paths
		} else {
			slog.Warn("importer: no beets-database match for imported album", "uploader", u.uploader, "staged_dir", u.stagedDir)
		}
	}
	return nil
}

// fileSkips relocates every path beets declined to auto-tag into review
// (or quarantine, on a review-filing failure), marks the owning unit's
// outcome, and returns the set of affected uploaders. units is a pointer
// because a skip matching no pending unit -- a leftover album staged in
// an earlier run, per [Importer.deriveOrphanUnit] -- is appended to it
// so the run report still carries the album's outcome. Filing is local
// I/O only, so this takes no context.
func (im *Importer) fileSkips(units *[]*albumUnit, pending []*albumUnit, skipped []string, dateStr string) map[string]bool {
	skipUsers := make(map[string]bool)
	for _, path := range skipped {
		u := findUnit(pending, path)
		if u == nil {
			u = findUnit(*units, path)
		}
		if u == nil {
			orphan, ok := im.deriveOrphanUnit(path)
			if !ok {
				slog.Warn("importer: skipped path matched no staged upload", "path", path)
				continue
			}
			*units = append(*units, orphan)
			u = orphan
		}

		dest, err := im.filerFor(dateStr, u.uploader).MoveToReview(path)
		if err != nil {
			u.result = OutcomeQuarantined
			u.errMsg = err.Error()
			slog.Error("importer: could not file skipped path", "path", path, "err", err)
			continue
		}
		if strings.HasPrefix(dest, im.Cfg.Paths.QuarantineDir) {
			u.result = OutcomeQuarantined
		} else {
			u.result = OutcomeReview
		}
		skipUsers[u.uploader] = true
	}
	return skipUsers
}

// resolveImported settles every pending unit still without a result:
// imported if its staged directory is now empty or gone (beets moved
// everything out of it), left in review otherwise -- a partial skip that
// took some but not all of a unit's files, the one shape a beets tag log
// can produce that this album-as-a-unit model does not cleanly resolve.
func (im *Importer) resolveImported(pending []*albumUnit, collection string) []*albumUnit {
	var imported []*albumUnit
	for _, u := range pending {
		if u.result != "" {
			continue
		}
		empty, err := dirIsEmptyOrGone(u.stagedDir)
		if err != nil {
			slog.Warn("importer: check staged upload after import", "staged_dir", u.stagedDir, "err", err)
			continue
		}
		if !empty {
			u.result = OutcomeReview
			u.errMsg = "left in staging after import"
			continue
		}
		u.result = OutcomeImported
		u.collection = collection
		imported = append(imported, u)
	}
	return imported
}

// notifySkips emails each affected uploader on file and one aggregate
// summary to the admin, matching move_skips_and_notify at HEAD.
func (im *Importer) notifySkips(ctx context.Context, skipUsers map[string]bool, dateStr string) {
	users := slices.Sorted(maps.Keys(skipUsers))
	reviewRoot := filepath.Join(im.Cfg.Paths.ReviewDir, dateStr)

	for _, user := range users {
		email, ok := im.Roster.EmailFor(user)
		if !ok {
			slog.Info("importer: no email on file for uploader; skipping their notice", "uploader", user)
			continue
		}
		body := fmt.Sprintf("Some of your uploaded music could not be auto-tagged and was moved to %s for the admin to review.", filepath.Join(reviewRoot, user))
		if err := im.Mailer.Send(ctx, email, "Dubplate: some uploads need review", body); err != nil {
			slog.Warn("importer: send skip notice", "uploader", user, "err", err)
		}
	}

	subject := fmt.Sprintf("Dubplate: import skipped items (%s)", dateStr)
	body := fmt.Sprintf("Contributors with skipped uploads today: %s. See %s.", strings.Join(users, " "), reviewRoot)
	if err := im.Mailer.Send(ctx, im.Cfg.AdminEmail, subject, body); err != nil {
		slog.Warn("importer: send admin skip summary", "err", err)
	}
}

// --- small filesystem helpers -------------------------------------------

// listContributorDirs returns the sorted names of every directory
// directly under inboxDir, so a run always processes contributors in the
// same deterministic order.
func listContributorDirs(inboxDir string) ([]string, error) {
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var users []string
	for _, e := range entries {
		if e.IsDir() {
			users = append(users, e.Name())
		}
	}
	slices.Sort(users)
	return users, nil
}

// dirHasEntries reports whether dir contains anything below itself,
// matching the shell layer's `find dir -mindepth 1 -print -quit`.
func dirHasEntries(dir string) (bool, error) {
	found := false
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path == dir {
			return nil
		}
		found = true
		return fs.SkipAll
	})
	return found, err
}

// quiescentSince reports whether every file under dir has an mtime at or
// before cutoff. Checking file mtimes rather than the directory's own
// mtime matters: appending to an existing file two levels down updates
// only that file's mtime, never a parent directory's, since no directory
// entry is added or removed.
func quiescentSince(dir string, cutoff time.Time) (bool, error) {
	quiescent := true
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if info.ModTime().After(cutoff) {
			quiescent = false
			return fs.SkipAll
		}
		return nil
	})
	return quiescent, err
}

// archiveSuffixes are the extensions [isArchiveName] recognizes,
// matching the shell layer's is_archive_name exactly.
var archiveSuffixes = []string{".zip", ".tar.gz", ".tgz", ".tar.bz2", ".tbz2", ".tar"}

// isArchiveName reports whether name is a filename beets can't read
// directly and must be unpacked first.
func isArchiveName(name string) bool {
	lower := strings.ToLower(name)
	for _, suffix := range archiveSuffixes {
		if strings.HasSuffix(lower, suffix) {
			return true
		}
	}
	return false
}

// stagingDestPath returns dir joined with name if nothing is there yet,
// or a numerically disambiguated path otherwise, so two archives in one
// batch extracting a same-named top-level entry land beside each other
// rather than merging or clobbering.
func stagingDestPath(dir, name string) string {
	dest := filepath.Join(dir, name)
	if !pathExists(dest) {
		return dest
	}
	ext := filepath.Ext(name)
	stem := strings.TrimSuffix(name, ext)
	for n := 2; ; n++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s.%d%s", stem, n, ext))
		if !pathExists(candidate) {
			return candidate
		}
	}
}

func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// findUnit returns the unit whose staged directory is path itself or an
// ancestor of it, or nil if none matches.
func findUnit(units []*albumUnit, path string) *albumUnit {
	for _, u := range units {
		if path == u.stagedDir || strings.HasPrefix(path, u.stagedDir+string(filepath.Separator)) {
			return u
		}
	}
	return nil
}

// dirIsEmptyOrGone reports whether path no longer exists, or exists as a
// directory with nothing left in it -- beets' move-on-import having
// taken everything out. A loose file still present (not a directory)
// reports false: it was not moved, so it was not imported.
func dirIsEmptyOrGone(path string) (bool, error) {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	if !info.IsDir() {
		return false, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	return len(entries) == 0, nil
}

// cleanEmptyStagingDirs removes every now-empty directory under root,
// cascading bottom-up so a directory left empty only after its own
// children are removed is still caught in the same pass -- matching
// `find "$STAGING_DIR" -mindepth 1 -type d -empty -delete`. root itself
// is never removed.
func cleanEmptyStagingDirs(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			if err := removeIfEmptyRecursive(filepath.Join(root, e.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func removeIfEmptyRecursive(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			if err := removeIfEmptyRecursive(filepath.Join(dir, e.Name())); err != nil {
				return err
			}
		}
	}
	entries, err = os.ReadDir(dir)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return os.Remove(dir)
	}
	return nil
}
