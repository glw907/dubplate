package main

import (
	"context"
	"fmt"
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
// lock inside SQLite. Its default follows the estate's own streamer:
// Navidrome's default DSN carries _busy_timeout=15000 on the same
// SQLite.
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

// storeBusyTimeout resolves DUBPLATE_STORE_BUSY_TIMEOUT from environ
// and checks it against its range. parseErr is set only when the
// value fails to parse at all, a failure doctor's own threshold row
// already names; rangeErr is set only when it parses but falls outside
// the range, naming the key and the bound it crossed. A caller never
// sees both at once, and d is meaningful only when both are nil.
func storeBusyTimeout(environ []string) (d time.Duration, parseErr, rangeErr error) {
	d, err := config.ParseDuration(environ, storeBusyTimeoutThreshold)
	if err != nil {
		return 0, err, nil
	}
	switch {
	case d < storeBusyTimeoutFloor:
		return d, nil, fmt.Errorf("%s must be at least %dms", storeBusyTimeoutThreshold.Key, storeBusyTimeoutFloorMS)
	case d > storeBusyTimeoutCeiling:
		return d, nil, fmt.Errorf("%s must be at most %dms", storeBusyTimeoutThreshold.Key, storeBusyTimeoutCeilingMS)
	}
	return d, nil, nil
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

	d, parseErr, rangeErr := storeBusyTimeout(environ)
	if parseErr != nil {
		return "not checked", true, store.Report{}, false
	}
	if rangeErr != nil {
		return "not checked: " + rangeErr.Error(), true, store.Report{}, false
	}

	r := store.Inspect(ctx, store.InspectOptions{StateDir: stateDir, BusyTimeout: d})
	status, failed = storeRowStatus(r)
	return status, failed, r, true
}

// storeRowStatus renders r as the "store" row's status text, and
// reports whether r's state exits nonzero: absent,
// older, and current do not; every other state does.
func storeRowStatus(r store.Report) (status string, failed bool) {
	switch r.State {
	case store.StateAbsent:
		status = "absent"
	case store.StateRestoreInProgress:
		status, failed = "restore in progress: "+r.Reason, true
	case store.StateBusy:
		status, failed = "busy: "+storeHolders(r.Holders), true
	case store.StateNotAStore:
		status, failed = "not a store: "+r.Reason, true
	case store.StateRefused:
		status, failed = "refused: "+r.Reason, true
	case store.StateOlder:
		status = fmt.Sprintf("older (v%d, binary v%d): pending %s", r.StoreVersion, r.BinaryVersion, strings.Join(r.Pending, ", "))
	case store.StateNewer:
		status, failed = fmt.Sprintf("newer (v%d, binary v%d)", r.StoreVersion, r.BinaryVersion), true
	case store.StateCurrent:
		status = fmt.Sprintf("current (v%d)", r.StoreVersion)
	default:
		// [store.State]'s doc comment states the set is closed; this
		// arm reports an unrecognized value rather than silently
		// printing an empty status, the same defensive shape
		// thresholdStatus's own default arm takes.
		status, failed = fmt.Sprintf("refused: unrecognized store state %d", r.State), true
	}
	if r.Warning != "" {
		status += "; " + r.Warning
	}
	return status, failed
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

// backupFactRow is one doctor row [backupFactsRows] renders: a
// printRow key and its status text.
type backupFactRow struct {
	key, status string
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
func backupFactsRows(now time.Time, report store.Report, checked bool) []backupFactRow {
	if !checked || report.Facts == nil {
		return []backupFactRow{
			{backupLastSuccessRow, "not checked"},
			{backupWALSizeRow, "not checked"},
			{backupFreeSpaceRow, "not checked"},
			{backupLocalCopiesRow, "not checked"},
		}
	}
	f := report.Facts
	return []backupFactRow{
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
