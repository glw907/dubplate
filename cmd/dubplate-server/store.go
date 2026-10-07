package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
)

// storeBusyTimeoutThreshold is DUBPLATE_STORE_BUSY_TIMEOUT: the
// inspect mode's only production reader at this rung, doctor and
// check, bounding how long a read waits on another process's write
// lock inside SQLite. The 15s default is long enough to ride out a
// snapshot or a migration holding the lock, short enough that a stuck
// writer surfaces as a doctor failure rather than a hang.
var storeBusyTimeoutThreshold = config.Threshold{
	Key:     "DUBPLATE_STORE_BUSY_TIMEOUT",
	Kind:    config.KindDuration,
	Default: "15s",
	Doc:     "a Go duration string, for example 15s",
}

// storeBusyTimeoutFloorMS and storeBusyTimeoutCeilingMS bound
// DUBPLATE_STORE_BUSY_TIMEOUT in milliseconds, the unit SQLite's
// _busy_timeout takes. The floor is 1ms: a positive duration under it
// would truncate to zero, no busy wait at all. The ceiling is the
// largest value a probe measured the driver's DSN parser
// (strconv.ParseInt(..., 10, 64) in sqlite.go) reading back unchanged;
// values above it read back as 0.
const (
	storeBusyTimeoutFloorMS   = 1
	storeBusyTimeoutCeilingMS = 2147483647
)

// storeBusyTimeoutFloor and storeBusyTimeoutCeiling are the same bound
// as [time.Duration], compared against the threshold's parsed value at
// nanosecond precision so a value like 500us fails the floor even
// though it would round to a whole millisecond.
var (
	storeBusyTimeoutFloor   = storeBusyTimeoutFloorMS * time.Millisecond
	storeBusyTimeoutCeiling = storeBusyTimeoutCeilingMS * time.Millisecond
)

// storeMigrationWaitThreshold is DUBPLATE_STORE_MIGRATION_WAIT: how
// long serve waits for the store lock before it exits naming the
// holders. A current store waits only on an exclusive holder, a
// restore, since shared locks coexist; a migrating start waits on
// every shared holder, a backup's sync above all, which on a first run
// lasts hours. The 5m default outlasts every member verb and every
// snapshot copy a group-scale store takes and is far shorter than a
// first sync; the unit's restart policy retries a start that gave up.
var storeMigrationWaitThreshold = config.Threshold{
	Key:     "DUBPLATE_STORE_MIGRATION_WAIT",
	Kind:    config.KindDuration,
	Default: "5m",
	Doc:     "a Go duration string, for example 5m",
}

// storeMigrationWaitFloor and storeMigrationWaitCeiling bound
// DUBPLATE_STORE_MIGRATION_WAIT.
const (
	storeMigrationWaitFloor   = time.Second
	storeMigrationWaitCeiling = 24 * time.Hour
)

// jobRetentionSetting is DUBPLATE_JOB_RETENTION: how long a finished
// job record survives past its own kind's newest succeeded record.
// Its default, 2160h (90 days), carries unchanged the value
// ops/local's retired retention constant held.
var jobRetentionSetting = config.Setting{
	Key:     "DUBPLATE_JOB_RETENTION",
	Kind:    config.KindDuration,
	Default: "2160h",
	Doc:     "a Go duration string, for example 2160h",
	Apply:   config.ApplyNextOperation,
	Floor:   "24h",
	Ceiling: "87600h",
}

// storeLocalSnapshotKeepSetting is DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP:
// how many local pre-migration snapshot copies OpenServe's own keep
// bound retains. Its default of 3 is a working value the retired
// package never set: a rollback needs the newest copy, and two older
// copies cover a bad upgrade found a release late, each a full store
// whose size doctor already shows. doctor is its first reader.
var storeLocalSnapshotKeepSetting = config.Setting{
	Key:     "DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP",
	Kind:    config.KindInt,
	Default: "3",
	Doc:     "an integer count of local snapshot copies to keep, for example 3",
	Apply:   config.ApplyNextOperation,
	Floor:   "1",
	Ceiling: "100",
}

// registerStoreConfig registers the store's own thresholds and
// store-homed settings against reg. newProductionRegistry calls it
// after registerAuthConfig.
func registerStoreConfig(reg *config.Registry) {
	reg.AddThreshold(storeBusyTimeoutThreshold)
	reg.AddThreshold(storeMigrationWaitThreshold)
	reg.AddSetting(jobRetentionSetting)
	reg.AddSetting(storeLocalSnapshotKeepSetting)
}

// jobRetentionReader returns the closure over acc that
// [local.Config.Retention] reads before every job start: the
// accessor's typed read of jobRetentionSetting, one per value
// ops/local needs.
func jobRetentionReader(acc *settings.Accessor) func(context.Context) (time.Duration, error) {
	return func(ctx context.Context) (time.Duration, error) {
		return acc.Duration(ctx, jobRetentionSetting)
	}
}

// errOutOfRange marks a threshold that parsed but fell outside its
// range, as opposed to one that did not parse at all.
var errOutOfRange = errors.New("out of range")

// rangeError is an out-of-range refusal whose message is its own
// sentence and whose Unwrap is [errOutOfRange].
type rangeError string

func (e rangeError) Error() string { return string(e) }

func (e rangeError) Unwrap() error { return errOutOfRange }

// storeBusyTimeout resolves DUBPLATE_STORE_BUSY_TIMEOUT from environ
// and checks it against its range. A value that parses outside the
// range is refused wrapping errOutOfRange and naming the key and the
// bound it crossed; a value that fails to parse is returned as
// config.ParseDuration reports it, a failure doctor's own threshold
// row already names.
func storeBusyTimeout(environ []string) (time.Duration, error) {
	d, err := config.ParseDuration(environ, storeBusyTimeoutThreshold)
	if err != nil {
		return 0, err
	}
	switch {
	case d < storeBusyTimeoutFloor:
		return 0, rangeError(fmt.Sprintf("%s must be at least %dms", storeBusyTimeoutThreshold.Key, storeBusyTimeoutFloorMS))
	case d > storeBusyTimeoutCeiling:
		return 0, rangeError(fmt.Sprintf("%s must be at most %dms", storeBusyTimeoutThreshold.Key, storeBusyTimeoutCeilingMS))
	}
	return d, nil
}

// storeMigrationWait resolves DUBPLATE_STORE_MIGRATION_WAIT from
// environ and checks it against its range, naming the key on every
// refusal.
func storeMigrationWait(environ []string) (time.Duration, error) {
	d, err := config.ParseDuration(environ, storeMigrationWaitThreshold)
	if err != nil {
		return 0, err
	}
	switch {
	case d < storeMigrationWaitFloor:
		return 0, fmt.Errorf("%s must be at least %s", storeMigrationWaitThreshold.Key, storeMigrationWaitFloor)
	case d > storeMigrationWaitCeiling:
		return 0, fmt.Errorf("%s must be at most %s", storeMigrationWaitThreshold.Key, storeMigrationWaitCeiling)
	}
	return d, nil
}

// localSnapshotKeep decides [store.ServeOptions.Keep] before serve
// opens the store, from the facts report holds and the registry's own
// setting, and refuses a start that a migration would otherwise make
// unrecoverable.
//
// A stored keep row in its bounds is the keep, even when the
// environment names another value: that value reaches the keep at the
// seed this start applies after its migration. The outcome's own
// value is the value in force after the seed, so it decides only when
// the row is out of bounds, where it is the healed value or a refusal
// naming the key and the bound. With no keep row, or no facts, the
// registry default is the keep and the open decides whether the state
// itself refuses.
//
// An older store whose evaluation refuses anything, whether an
// administrator's row out of bounds or an environment value that does
// not parse, is refused here, before the migration moves it past the
// binary a rollback would run. An older store is evaluated against
// the facts [seedFacts] picks, as check and doctor evaluate it, so one
// whose version predates the settings table, or the jobs table and
// with it any facts at all, is evaluated against the empty table its
// first seed meets. A store that does not migrate gets no such check,
// nor does one that carries no facts, and the seed apply after the
// open decides.
func localSnapshotKeep(report store.Report, environ []string, reg *config.Registry) (int, error) {
	setting := byKey(reg.Settings(), storeLocalSnapshotKeepSetting.Key)
	keep, err := strconv.Atoi(setting.Default)
	if err != nil {
		return 0, fmt.Errorf("%s: default %q: %w", setting.Key, setting.Default, err)
	}
	var facts store.Facts
	switch {
	case report.State == store.StateOlder:
		facts, _ = seedFacts(report)
	case report.Facts != nil:
		facts = *report.Facts
	default:
		return keep, nil
	}

	ev := settings.EvaluateFacts(reg, environ, facts)
	if report.State == store.StateOlder && ev.Err != nil {
		return 0, ev.Err
	}
	for _, o := range ev.Outcomes {
		if o.Key != setting.Key || o.Stored == nil {
			continue
		}
		if setting.Validate(o.Stored.Value) == nil {
			return strconv.Atoi(o.Stored.Value)
		}
		if o.Err != nil {
			return 0, o.Err
		}
		return strconv.Atoi(o.Value)
	}
	return keep, nil
}

// storeRow resolves the store's state from environ and renders it as
// doctor's and check's shared "store" row: the status text and
// whether that state exits nonzero. It reads
// DUBPLATE_STATE_DIR itself, because the store reads no environment,
// so the row still renders when config.Load has already failed on a
// different key. A missing state directory or a busy timeout that
// fails to parse reports "not checked" without opening anything; a
// busy timeout that parses outside its range reports "not checked"
// naming the key and the bound.
//
// It also returns the [store.Report] it read and whether it ran
// [store.Inspect] at all (checked), so a caller building the backup
// facts rows (R-6bi13) from the same report never inspects the store
// twice: report is the zero value whenever checked is false.
func storeRow(ctx context.Context, environ []string) (status string, failed bool, report store.Report, checked bool) {
	stateDir, ok := lookupEnviron(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return "not checked", false, store.Report{}, false
	}

	d, err := storeBusyTimeout(environ)
	switch {
	case errors.Is(err, errOutOfRange):
		return "not checked: " + err.Error(), true, store.Report{}, false
	case err != nil:
		return "not checked", true, store.Report{}, false
	}

	r := store.Inspect(ctx, store.InspectOptions{StateDir: stateDir, BusyTimeout: d})
	status, failed = storeRowStatus(r)
	return status, failed, r, true
}

// storeRowStatus renders r as the "store" row's status text, and
// reports whether r's state exits nonzero: every state [storeVouches]
// cannot vouch for does.
func storeRowStatus(r store.Report) (status string, failed bool) {
	switch r.State {
	case store.StateAbsent:
		status = "absent"
	case store.StateRestoreInProgress:
		status = "restore in progress: " + r.Reason
	case store.StateBusy:
		status = "busy: " + storeHolders(r.Holders)
	case store.StateNotAStore:
		status = "not a store: " + r.Reason
	case store.StateRefused:
		status = "refused: " + r.Reason
	case store.StateOlder:
		status = fmt.Sprintf("older (v%d, binary v%d): pending %s", r.StoreVersion, r.BinaryVersion, strings.Join(r.Pending, ", "))
	case store.StateNewer:
		status = fmt.Sprintf("newer (v%d, binary v%d)", r.StoreVersion, r.BinaryVersion)
	case store.StateCurrent:
		status = fmt.Sprintf("current (v%d)", r.StoreVersion)
	default:
		// [store.State]'s doc comment states the set is closed; this
		// arm reports an unrecognized value rather than silently
		// printing an empty status, the same defensive shape
		// thresholdStatus's own default arm takes.
		status = fmt.Sprintf("refused: unrecognized store state %d", r.State)
	}
	if r.Warning != "" {
		status += "; " + r.Warning
	}
	return status, !storeVouches(r.State)
}

// storeVouches reports whether doctor vouches for a store in state s:
// absent, older, and current it does; every other state, an
// unrecognized one included, fails the store row and leaves every
// setting row "not checked".
func storeVouches(s store.State) bool {
	switch s {
	case store.StateAbsent, store.StateOlder, store.StateCurrent:
		return true
	}
	return false
}

// storeHolders renders a busy store's holder pids, or "held by
// another process" when [store.Report.Holders] names none because
// /proc/locks could not identify them.
func storeHolders(holders []string) string {
	if len(holders) == 0 {
		return "held by another process"
	}
	return "held by pid " + strings.Join(holders, ", ")
}

// settingsDoctorRows renders doctor's one row per registered setting,
// in declaration order, and one row per orphan. Its disposition
// follows the store's own state: a state [storeVouches] refuses
// reports every setting "not checked", never calling into
// internal/settings; every other state evaluates the seed rule against
// the facts [seedFacts] picks. A store with no settings table yet
// shows each setting at its default and what serve's next start seeds,
// unless the seed would refuse, which reports here as it would there.
// It reports whether any row refuses.
func settingsDoctorRows(environ []string, reg *config.Registry, report store.Report, checked bool) (rows []doctorRow, failed bool) {
	list := reg.Settings()
	facts, ok := seedFacts(report)
	if !checked || !storeVouches(report.State) || !ok {
		for _, s := range list {
			rows = append(rows, doctorRow{s.Key, settingsNotCheckedStatus(environ, s, report.State, checked)})
		}
		return rows, false
	}

	ev := settings.EvaluateFacts(reg, environ, facts)
	for _, o := range ev.Outcomes {
		s := byKey(list, o.Key)
		switch {
		case o.Err != nil:
			failed = true
			rows = append(rows, doctorRow{o.Key, o.Report})
		case facts.SettingsAbsent:
			rows = append(rows, doctorRow{o.Key, settingsNoTableStatus(environ, s)})
		default:
			rows = append(rows, doctorRow{o.Key, settingsOutcomeStatus(s, o)})
		}
	}
	for _, orphan := range ev.Orphans {
		rows = append(rows, doctorRow{orphan.Key, "no longer registered; ignored"})
	}
	return rows, failed
}

// seedFacts returns the facts the seed rule is evaluated against for
// a store in report, and false when there are none to evaluate. An
// absent store, and an older one whose version predates the settings
// table, holds no settings row: its facts are an absent table, the
// empty row set serve's first seed meets, so an environment value that
// start would refuse is refused before it. An older store below the
// version that adds the jobs table carries no facts at all, and its
// state alone marks it. Any other state evaluates the facts
// [store.Inspect] read, and a state that read none has nothing to
// evaluate.
func seedFacts(report store.Report) (store.Facts, bool) {
	switch {
	case report.State == store.StateAbsent,
		report.State == store.StateOlder && (report.Facts == nil || report.Facts.SettingsAbsent):
		return store.Facts{SettingsAbsent: true}, true
	case report.Facts != nil:
		return *report.Facts, true
	}
	return store.Facts{}, false
}

// byKey returns the setting named key out of list, or its zero value
// when list carries none, the shape a defensive caller falls back on
// rather than panics on a key [settings.Evaluate] could not have
// invented.
func byKey(list []config.Setting, key string) config.Setting {
	for _, s := range list {
		if s.Key == key {
			return s
		}
	}
	return config.Setting{}
}

// settingsOutcomeStatus renders one from-facts outcome as doctor's
// status for the registered setting s: the value in force, its
// source, and the apply mode, then any note about what changes at
// serve's next start. A refused outcome never reaches it.
func settingsOutcomeStatus(s config.Setting, o settings.Outcome) string {
	value, source := s.Default, "default"
	if o.Stored != nil {
		value, source = o.Stored.Value, o.Stored.Source
	}
	status := strconv.Quote(value) + " " + settingTraits(s, source)
	if note := settingsChangeNote(o); note != "" {
		status += "; " + note
	}
	return status
}

// settingsChangeNote renders what o's change does at serve's next
// start, or "" for an outcome the seed rule leaves unchanged.
func settingsChangeNote(o settings.Outcome) string {
	switch o.Change {
	case settings.Seed:
		return fmt.Sprintf("env %q seeds at serve's next start", o.Env)
	case settings.Reseed:
		return fmt.Sprintf("env %q re-seeds at serve's next start", o.Env)
	case settings.Remove:
		return "returns to the default at serve's next start"
	default:
		return ""
	}
}

// settingsNoTableStatus renders s's row for a store that holds no
// settings table yet, or none of any store at all: the default is in
// force, and an env value the seed accepts seeds at serve's next start.
func settingsNoTableStatus(environ []string, s config.Setting) string {
	status := strconv.Quote(s.Default) + " " + settingTraits(s, "default") + "; store holds no settings yet"
	if v, ok := lookupEnviron(environ, s.Key); ok {
		status += fmt.Sprintf("; env %q seeds at serve's next start", v)
	}
	return status
}

// settingsNotCheckedStatus renders s's row for a store state this
// binary cannot vouch for: the default is shown for reference, since
// no read reached the store's own row, and an env value is named as
// not checked rather than compared. checked names whether storeRow ran
// [store.Inspect] at all; when it did not, no state names why, so the
// row says only "not checked".
func settingsNotCheckedStatus(environ []string, s config.Setting, state store.State, checked bool) string {
	label := "not checked"
	if checked {
		label = fmt.Sprintf("not checked (%s)", state)
	}
	status := strconv.Quote(s.Default) + " " + settingTraits(s, "default") + "; store " + label
	if v, ok := lookupEnviron(environ, s.Key); ok {
		status += fmt.Sprintf("; env %q not checked against the store", v)
	}
	return status
}

// settingTraits renders the parenthesized traits doctor prints after a
// setting's value: the value's source, s's apply mode, and the
// sensitive mark when s carries it.
func settingTraits(s config.Setting, source string) string {
	if s.Sensitive {
		return fmt.Sprintf("(source %s, %s, sensitive)", source, s.Apply)
	}
	return fmt.Sprintf("(source %s, %s)", source, s.Apply)
}

// backupFactsRows renders R-6bi13's read-only facts as doctor's four
// backup rows: the newest succeeded backup's age (or "never"), the
// store's -wal file size, the leaf's free space, and the local
// snapshot copy count and size. Every row prints unconditionally, the
// store row's own precedent (doctor.go's storeRow doc comment): a
// store facts cannot vouch for is as reportable as one that can.
//
// Every row reads "not checked" when checked is false (storeRow never
// ran [store.Inspect]) or report carries no Facts, which
// [store.Inspect] leaves nil for a store this binary cannot read as
// current: absent, busy, mid-restore, not a store, refused, newer, or
// older below the version that added the jobs table.
func backupFactsRows(now time.Time, report store.Report, checked bool) []doctorRow {
	if !checked || report.Facts == nil {
		return []doctorRow{
			{backupLastSuccessRow, "not checked"},
			{backupWALSizeRow, "not checked"},
			{backupFreeSpaceRow, "not checked"},
			{backupLocalCopiesRow, "not checked"},
		}
	}
	f := report.Facts
	return []doctorRow{
		{backupLastSuccessRow, lastSuccessStatus(now, f.LastSuccess)},
		{backupWALSizeRow, formatBytes(f.WALSize)},
		{backupFreeSpaceRow, freeSpaceStatus(f)},
		{backupLocalCopiesRow, localCopiesStatus(f)},
	}
}

// lastSuccessStatus renders the backup kind's newest succeeded record
// out of last: "never" when it has none, otherwise its age rounded to
// minutes, with its final progress text appended when that text names
// a skipped or failed prune (anything but [backup.FinishedMessage],
// the clean case).
func lastSuccessStatus(now time.Time, last map[string]store.JobSuccess) string {
	success, ok := last[ops.KindBackup]
	if !ok {
		return "never"
	}
	status := formatAge(now.Sub(success.Finished)) + " ago"
	if success.Progress != "" && success.Progress != backup.FinishedMessage {
		status += ": " + success.Progress
	}
	return status
}

// formatAge renders d, rounded to whole minutes, as "<N>m" or, past an
// hour or a day, "<N>h<N>m" or "<N>d<N>h": a negative or sub-minute d
// (a finish time in the future, or in the last minute) renders as
// "0m".
func formatAge(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	minutes := int64(d.Round(time.Minute) / time.Minute)
	hours := minutes / 60
	minutes %= 60
	days := hours / 24
	hours %= 24
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, minutes)
	default:
		return fmt.Sprintf("%dm", minutes)
	}
}

// freeSpaceStatus renders f's free-space fact: "unknown" when statfs
// failed, its byte count otherwise.
func freeSpaceStatus(f *store.Facts) string {
	if !f.FreeKnown {
		return "unknown"
	}
	return formatBytes(int64(f.FreeBytes))
}

// localCopiesStatus renders f's local snapshot copy count and total
// size: "0 copies" for none, the count and size otherwise.
func localCopiesStatus(f *store.Facts) string {
	if f.Copies == 0 {
		return "0 copies"
	}
	return fmt.Sprintf("%d copies, %s", f.Copies, formatBytes(f.CopiesSize))
}

// formatBytes renders n in binary units (KiB, MiB, GiB, ...), the same
// scale most disk-usage tools use: a value under 1024 renders as
// "<n> B", and each larger unit carries one decimal place.
func formatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for v := n / unit; v >= unit; v /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
