package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
)

// doctorRowName is the name of the "store" row doctor and check print.
const doctorRowName = "store"

// Doctor resolves the store's state from environ, running readers
// against it, and renders it as the "store" row. It reads
// DUBPLATE_STATE_DIR itself, because the store reads no environment
// otherwise, so the row still renders when [config.Load] has already
// failed on a different key. A missing state directory or a busy
// timeout that fails to parse reports "not checked" without opening
// anything; a busy timeout that parses outside its range reports "not
// checked" naming the key and the bound.
//
// It also returns the [Report] it read and whether it ran [Inspect]
// at all (checked), so a caller building further rows from the same
// report never inspects the store twice: report is the zero value, and
// no reader ran, whenever checked is false.
func Doctor(ctx context.Context, environ []string, readers ...Reader) (row doctor.Row, report Report, checked bool) {
	notChecked := doctor.Row{Name: doctorRowName, Status: "not checked"}
	stateDir, ok := config.Lookup(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return notChecked, Report{}, false
	}

	d, err := BusyTimeout(environ)
	switch {
	case errors.Is(err, ErrOutOfRange):
		notChecked.Status, notChecked.Failed = "not checked: "+err.Error(), true
		return notChecked, Report{}, false
	case err != nil:
		notChecked.Failed = true
		return notChecked, Report{}, false
	}

	r := Inspect(ctx, InspectOptions{StateDir: stateDir, BusyTimeout: d, Readers: readers})
	return r.Row(), r, true
}

// Row renders r as the "store" row. The row fails for every state
// [State.Vouches] cannot vouch for.
func (r Report) Row() doctor.Row {
	var status string
	switch r.State {
	case StateAbsent:
		status = "absent"
	case StateRestoreInProgress:
		status = "restore in progress: " + r.Reason
	case StateBusy:
		status = "busy: " + holders(r.Holders)
	case StateNotAStore:
		status = "not a store: " + r.Reason
	case StateRefused:
		status = "refused: " + r.Reason
	case StateOlder:
		status = fmt.Sprintf("older (v%d, binary v%d): pending %s", r.StoreVersion, r.BinaryVersion, strings.Join(r.Pending, ", "))
	case StateNewer:
		status = fmt.Sprintf("newer (v%d, binary v%d)", r.StoreVersion, r.BinaryVersion)
	case StateCurrent:
		status = fmt.Sprintf("current (v%d)", r.StoreVersion)
	default:
		// [State]'s doc comment states the set is closed; this arm
		// reports an unrecognized value rather than silently printing
		// an empty status.
		status = fmt.Sprintf("refused: unrecognized store state %d", r.State)
	}
	if r.Warning != "" {
		status += "; " + r.Warning
	}
	return doctor.Row{Name: doctorRowName, Status: status, Failed: !r.State.Vouches()}
}

// Vouches reports whether doctor vouches for a store in state s:
// absent, older, and current it does; every other state, an
// unrecognized one included, fails the store row and leaves every
// setting row "not checked".
func (s State) Vouches() bool {
	switch s {
	case StateAbsent, StateOlder, StateCurrent:
		return true
	}
	return false
}

// holders renders a busy store's holder pids, or "held by another
// process" when [Report.Holders] names none because /proc/locks could
// not identify them.
func holders(pids []string) string {
	if len(pids) == 0 {
		return "held by another process"
	}
	return "held by pid " + strings.Join(pids, ", ")
}
