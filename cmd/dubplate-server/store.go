package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
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

// registerStoreConfig registers the store's own thresholds against
// reg. newProductionRegistry calls it after registerAuthConfig.
func registerStoreConfig(reg *config.Registry) {
	reg.AddThreshold(storeBusyTimeoutThreshold)
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
func storeRow(ctx context.Context, environ []string) (status string, failed bool) {
	stateDir, ok := lookupEnviron(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return "not checked", false
	}

	d, parseErr, rangeErr := storeBusyTimeout(environ)
	if parseErr != nil {
		return "not checked", true
	}
	if rangeErr != nil {
		return "not checked: " + rangeErr.Error(), true
	}

	r := store.Inspect(ctx, store.InspectOptions{StateDir: stateDir, BusyTimeout: d})
	return storeRowStatus(r)
}

// storeRowStatus renders r as the "store" row's status text, and
// reports whether r's state exits nonzero: absent,
// older, and current do not; every other state does.
func storeRowStatus(r store.Report) (status string, failed bool) {
	switch r.State {
	case store.StateAbsent:
		status = "absent"
	case store.StateRestoreInProgress:
		status, failed = "restore in progress: finish it with restore --resume", true
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
