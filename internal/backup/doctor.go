package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/store"
)

// featuresWaitDelay bounds how long the features probe's child has to
// exit after its SIGINT before it is killed, the same bound
// [os/exec.Cmd.WaitDelay] gives rclone's children in rclone.go.
const featuresWaitDelay = 5 * time.Second

// cryptWarning is what a backup remote not wrapped in an rclone crypt
// remote earns: it carries every snapshot, addresses included,
// unencrypted at the provider.
const cryptWarning = "the backup remote is not wrapped in an rclone crypt remote; a snapshot carries member data unencrypted at the provider"

// featuresTimeout bounds the `rclone backend features` probe behind
// the backup remote row.
const featuresTimeout = 10 * time.Second

// The backup rows, apart from the integration class's "backup" row,
// which reads enabled or disabled.
const (
	lastSuccessRowName = "backup last success"
	walSizeRowName     = "backup wal size"
	freeSpaceRowName   = "backup free space"
	localCopiesRowName = "backup local copies"
	remoteRowName      = "backup remote"
)

// DoctorRows renders the backup's five doctor rows: the four read-only
// facts the store inspection into report read, then the remote. last
// is the newest succeeded job of each kind the same inspection read,
// and now is the time the last success's age is measured from.
//
// Every facts row prints, reading "not checked" when report carries no
// Facts: the store row never ran [store.Inspect], or it left Facts nil
// for a store this binary cannot read as current: absent, busy, mid-restore, not a store,
// refused, newer, or older below the version that added the jobs
// table. Only the remote row can fail.
func DoctorRows(ctx context.Context, environ []string, now time.Time, report store.Report, last map[string]ops.JobSuccess) []doctor.Row {
	rows := factsRows(now, report, last)
	remote := doctor.Row{Name: remoteRowName}
	remote.Status, remote.Failed = remoteStatus(ctx, environ, featuresTimeout)
	return append(rows, remote)
}

func factsRows(now time.Time, report store.Report, last map[string]ops.JobSuccess) []doctor.Row {
	if report.Facts == nil {
		return []doctor.Row{
			{Name: lastSuccessRowName, Status: "not checked"},
			{Name: walSizeRowName, Status: "not checked"},
			{Name: freeSpaceRowName, Status: "not checked"},
			{Name: localCopiesRowName, Status: "not checked"},
		}
	}
	f := report.Facts
	return []doctor.Row{
		{Name: lastSuccessRowName, Status: lastSuccessStatus(now, last)},
		{Name: walSizeRowName, Status: formatBytes(f.WALSize)},
		{Name: freeSpaceRowName, Status: freeSpaceStatus(f)},
		{Name: localCopiesRowName, Status: localCopiesStatus(f)},
	}
}

// lastSuccessStatus renders the backup kind's newest succeeded record
// out of last: "never" when it has none, otherwise its age rounded to
// minutes, with its final progress text appended when that text names
// a skipped or failed prune (anything but finishedMessage, the clean
// case).
func lastSuccessStatus(now time.Time, last map[string]ops.JobSuccess) string {
	success, ok := last[ops.KindBackup]
	if !ok {
		return "never"
	}
	status := formatAge(now.Sub(success.Finished)) + " ago"
	if success.Progress != "" && success.Progress != finishedMessage {
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

// remoteStatus renders the backup remote row: unavailable while the
// flag is off, without running any subprocess; the parse, _TYPE, and
// crypt checks doctor alone runs, since [Job.Run] reads no
// environment; and the features probe last, under timeout. It calls
// [ParseRemote], the same validator the job's first phase calls, so
// the two never disagree about a well-formed remote.
func remoteStatus(ctx context.Context, environ []string, timeout time.Duration) (status string, failed bool) {
	enabled, _ := Class.Enabled(environ)
	if !enabled {
		return "unavailable: backup integration disabled", false
	}

	remoteValue, _ := config.Lookup(environ, RemoteKey)
	if _, err := ParseRemote(remoteValue); err != nil {
		return remoteParseError(remoteValue), true
	}
	name, _, _ := strings.Cut(remoteValue, ":")
	upper := strings.ToUpper(name)

	typeKey := "RCLONE_CONFIG_" + upper + "_TYPE"
	remoteType, ok := config.Lookup(environ, typeKey)
	if !ok {
		return "invalid remote: " + typeKey + " is not set", true
	}

	if remoteType == "crypt" {
		if status, failed, blocked := cryptCheck(environ, upper); blocked {
			return status, failed
		}
	}

	bin, err := exec.LookPath("rclone")
	if err != nil {
		return "rclone not found", true
	}

	status, failed = featuresStatus(ctx, bin, name, timeout)
	if remoteType != "crypt" {
		status += "; " + cryptWarning
	}
	return status, failed
}

// remoteParseError renders [ParseRemote]'s refusal of value as one of
// the two shapes the row takes: the generic "want name:path" for no
// colon or an empty name, or the offending character quoted against
// rclone's rule for a name holding one.
func remoteParseError(value string) string {
	name, _, hasColon := strings.Cut(value, ":")
	if !hasColon || name == "" {
		return "invalid remote: want <name>:<path>"
	}
	if r, ok := firstInvalidRemoteNameRune(name); ok {
		return fmt.Sprintf("invalid remote: name %q holds %q, and rclone remote names \"can only contain letters, digits, or the _ (underscore) character\"", name, string(r))
	}
	// Unreachable given ParseRemote's rule (name and hasColon
	// already passed), kept so this function never returns an empty
	// status for a value ParseRemote refuses for a reason this switch
	// does not yet know.
	return "invalid remote: " + value
}

// firstInvalidRemoteNameRune returns the first rune in name outside
// rclone's remote-name alphabet (ASCII letters, digits, underscore).
func firstInvalidRemoteNameRune(name string) (rune, bool) {
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			continue
		}
		return r, true
	}
	return 0, false
}

// cryptCheck runs the crypt remote's two checks against the remote
// named by upper: that RCLONE_CONFIG_<upper>_REMOTE is set, and, when
// its value names another remote rather than a plain path or a
// connection string, that the named remote's _TYPE is set. blocked
// is true only when a check refuses; status and failed are meaningful
// only then. A plain path (no colon) or a connection string (a leading
// colon) skips the wrapped check, leaving the features probe to decide.
func cryptCheck(environ []string, upper string) (status string, failed, blocked bool) {
	remoteKey := "RCLONE_CONFIG_" + upper + "_REMOTE"
	wrapped, ok := config.Lookup(environ, remoteKey)
	if !ok {
		return "invalid remote: " + remoteKey + " is not set", true, true
	}
	wrappedName, _, hasColon := strings.Cut(wrapped, ":")
	if !hasColon || wrappedName == "" {
		return "", false, false
	}
	wrappedTypeKey := "RCLONE_CONFIG_" + strings.ToUpper(wrappedName) + "_TYPE"
	if _, ok := config.Lookup(environ, wrappedTypeKey); !ok {
		return fmt.Sprintf("invalid remote: wrapped remote %s: %s is not set", wrappedName, wrappedTypeKey), true, true
	}
	return "", false, false
}

// featuresJSON is the shape of `rclone backend features`'s stdout this
// probe reads: everything else in the object is backend-specific and
// unread.
type featuresJSON struct {
	Features struct {
		Copy bool
		Move bool
	}
}

// featuresStatus runs `rclone backend features <name>: --low-level-retries 1`
// under timeout and renders its outcome: "not checked: ..." on a
// deadline hit, whatever the child's exit (the deadline takes
// precedence, the same rule [classify] takes for the backup job's
// child); the child's last stderr line on any other nonzero exit; and,
// on success, a refusal naming name when neither Copy nor Move is true,
// since rclone's sync then has no backend to satisfy --backup-dir.
func featuresStatus(ctx context.Context, bin, name string, timeout time.Duration) (status string, failed bool) {
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(probeCtx, bin, "backend", "features", name+":", "--low-level-retries", "1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = featuresWaitDelay
	err := cmd.Run()

	if probeCtx.Err() != nil {
		return "not checked: rclone backend features timed out", false
	}
	if err != nil {
		if line := lastNonBlankLine(stderr.String()); line != "" {
			return line, true
		}
		return fmt.Sprintf("rclone backend features %s: %v", name, err), true
	}

	var parsed featuresJSON
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		return fmt.Sprintf("rclone backend features %s: could not parse its output: %v", name, err), true
	}
	if !parsed.Features.Copy && !parsed.Features.Move {
		return fmt.Sprintf(`backend %s lacks server-side move and copy: rclone's manual requires one for --backup-dir ("the remote in use must support server-side move or copy")`, name), true
	}
	return "ok", false
}

// lastNonBlankLine returns the last non-blank line of s, trimmed, or
// "" when s holds none.
func lastNonBlankLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for _, i := range slices.Backward(lines) {
		if line := strings.TrimSpace(i); line != "" {
			return line
		}
	}
	return ""
}
