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
)

// featuresWaitDelay bounds how long the features probe's child has to
// exit after its SIGINT before it is killed, the same bound
// [os/exec.Cmd.WaitDelay] gives rclone's children in rclone.go.
const featuresWaitDelay = 5 * time.Second

// cryptWarning is what a backup remote not wrapped in an rclone crypt
// remote earns: it carries every snapshot, addresses included,
// unencrypted at the provider.
const cryptWarning = "the backup remote is not wrapped in an rclone crypt remote; a snapshot carries member data unencrypted at the provider"

// RemoteRowOptions are the process facts [RemoteRowStatus] reads.
type RemoteRowOptions struct {
	// FeaturesTimeout bounds the `rclone backend features` probe.
	FeaturesTimeout time.Duration
}

// RemoteRowStatus renders doctor's "backup remote" row: unavailable
// while the flag is off, without running any subprocess; the parse,
// _TYPE, and crypt checks doctor alone runs, since [Job.Run] reads no
// environment; and the features probe last, under opts.FeaturesTimeout.
// It calls [ParseRemote], the same validator the job's first phase
// calls, so the two never disagree about a well-formed remote.
func RemoteRowStatus(ctx context.Context, environ []string, opts RemoteRowOptions) (status string, failed bool) {
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

	status, failed = featuresStatus(ctx, bin, name, opts.FeaturesTimeout)
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
