package main

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

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
// facts rows from the same report never inspects the store
// twice: report is the zero value whenever checked is false.
func storeRow(ctx context.Context, environ []string) (status string, failed bool, report store.Report, checked bool) {
	stateDir, ok := config.Lookup(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return "not checked", false, store.Report{}, false
	}

	d, err := store.BusyTimeout(environ)
	switch {
	case errors.Is(err, store.ErrOutOfRange):
		return "not checked: " + err.Error(), true, store.Report{}, false
	case err != nil:
		return "not checked", true, store.Report{}, false
	}

	r := store.Inspect(ctx, store.InspectOptions{StateDir: stateDir, BusyTimeout: d, Address: adminAddress(environ)})
	status, failed = storeRowStatus(r)
	return status, failed, r, true
}

// adminAddress returns DUBPLATE_ADMIN_EMAIL as the roster stores an
// address: parsed and lowercased as [config.Load] reads it. It returns
// "" when the variable is unset or does not parse, which names no
// address for the inspection to look up.
func adminAddress(environ []string) string {
	v, _ := config.Lookup(environ, adminEmailKey)
	addr, err := mail.ParseAddress(v)
	if err != nil {
		return ""
	}
	return strings.ToLower(addr.Address)
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
// the facts [settings.SeedFacts] picks. A store with no settings table yet
// shows each setting at its default and what serve's next start seeds,
// unless the seed would refuse, which reports here as it would there.
// It reports whether any row refuses.
func settingsDoctorRows(environ []string, reg *config.Registry, report store.Report, checked bool) (rows []doctorRow, failed bool) {
	list := reg.Settings()
	facts, ok := settings.SeedFacts(report)
	if !checked || !storeVouches(report.State) || !ok {
		for _, s := range list {
			rows = append(rows, doctorRow{s.Key, settingsNotCheckedStatus(environ, s, report.State, checked)})
		}
		return rows, false
	}

	ev := settings.EvaluateFacts(reg, environ, facts)
	registered := make(map[string]config.Setting, len(list))
	for _, s := range list {
		registered[s.Key] = s
	}
	for _, o := range ev.Outcomes {
		s := registered[o.Key]
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
	if v, ok := config.Lookup(environ, s.Key); ok {
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
	if v, ok := config.Lookup(environ, s.Key); ok {
		status += fmt.Sprintf("; env %q not checked against the store", v)
	}
	return status
}

// settingTraits renders the parenthesized traits doctor prints after a
// setting's value: the value's source, s's apply mode, and the
// sensitive mark when s carries it.
func settingTraits(s config.Setting, source string) string {
	if s.Mark == config.Sensitive {
		return fmt.Sprintf("(source %s, %s, sensitive)", source, s.Apply)
	}
	return fmt.Sprintf("(source %s, %s)", source, s.Apply)
}

// backupFactsRows renders the backup's read-only facts as doctor's four
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
