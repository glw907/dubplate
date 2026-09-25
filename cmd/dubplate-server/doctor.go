package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// newDoctorCmd builds the doctor subcommand: it reports the config
// wall's owned rows, reserved classes, integration classes,
// registered thresholds, and declared identity sources, one line each
// in "%-32s %s" form and one block per family in that order, deriving
// every row from reg so no rung hard-codes a count. That table is the
// command's human output; on an unsatisfied wall RunE returns a single
// short error, leaving main as the sole stderr writer. Exit status is 1
// when [config.Load] refuses environ, the roster row fails to open or
// names no admin, or the store row reports a failing state; 0
// otherwise.
func newDoctorCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Report the configuration wall",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, loadErr := config.Load(environ, reg)
			statuses := rowStatuses(loadErr)

			for _, row := range reg.OwnedWall() {
				status := cmp.Or(row.Pending, statuses[row.Name], "ok")
				if err := printRow(stdout, row.Name, status); err != nil {
					return err
				}
			}

			for _, class := range reg.ReservedClasses() {
				status := fmt.Sprintf("disabled (arrives at rung %d)", class.Rung)
				if err := printRow(stdout, class.Name, status); err != nil {
					return err
				}
			}

			for _, class := range reg.IntegrationClasses() {
				status := "disabled"
				if enabled, err := class.Enabled(environ); err == nil && enabled {
					status = "enabled"
				}
				if err := printRow(stdout, class.Name, status); err != nil {
					return err
				}
			}

			for _, th := range reg.Thresholds() {
				if err := printRow(stdout, th.Key, thresholdStatus(environ, statuses, th)); err != nil {
					return err
				}
			}

			for _, src := range reg.IdentitySources() {
				if err := printRow(stdout, src.Name, identitySourceStatus(environ, src, statuses)); err != nil {
					return err
				}
			}

			// The roster row runs R3's principal-collision check in
			// doctor as well as at startup and at write time: it calls
			// the same [auth.OpenRoster] the magic-link source loads
			// through, so a missing file, a duplicate address, a
			// duplicate streamer_username, or an unparseable record is
			// reported here too. A missing file is a failure like any
			// other, since R2 makes it a startup refusal: the spec's
			// runbook order is install, render the environment, run the
			// bootstrap verb that creates the roster, then start the
			// unit, so doctor run before the verb correctly reports the
			// deployment as not yet startable. The row prints only while
			// a selected identity source requires DUBPLATE_ROSTER_FILE
			// and the roster fails to open, so R10's line-count
			// arithmetic stays keyed to reg alone whenever there is
			// nothing to report. It names no value (Resolution R-5b6):
			// [auth.OpenRoster]'s own errors carry only a record's
			// position and the path, never its address or principal, and
			// a multi-record refusal is folded onto the one line. A
			// roster that loads but names no admin fails the same row,
			// since nobody could then change it but by hand, and the row
			// names the command that recovers it.
			rosterFailed := false
			if path, ok := selectedRosterPath(environ, reg); ok {
				if status, failed := rosterRowStatus(path); failed {
					rosterFailed = true
					if err := printRow(stdout, "roster", status); err != nil {
						return err
					}
				}
			}

			// The store row runs the inspect mode after every other
			// row, always printing exactly one line, absent state
			// included: a store the doctor run cannot vouch for is as
			// reportable as a missing config key.
			storeStatus, storeFailed := storeRow(cmd.Context(), environ)
			if err := printRow(stdout, "store", storeStatus); err != nil {
				return err
			}

			if loadErr != nil || rosterFailed || storeFailed {
				return errors.New("configuration is incomplete")
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// selectedRosterPath returns DUBPLATE_ROSTER_FILE's value when a
// source reg declares is selected in environ and requires that key.
// It reports false when no roster-backed source is selected or the key
// is unset; the source's own row already reports the unset key.
func selectedRosterPath(environ []string, reg *config.Registry) (string, bool) {
	const rosterKey = "DUBPLATE_ROSTER_FILE"
	for _, src := range reg.IdentitySources() {
		if src.Selected(environ) && slices.Contains(src.Required, rosterKey) {
			return lookupEnviron(environ, rosterKey)
		}
	}
	return "", false
}

// rosterRowStatus opens the roster at path and reports the roster
// row's status and whether the row fails: an [auth.OpenRoster] refusal
// rendered by rosterStatus, or a roster that loads but names no admin.
func rosterRowStatus(path string) (string, bool) {
	roster, err := auth.OpenRoster(path)
	if err != nil {
		return rosterStatus(err), true
	}
	members, err := roster.List()
	if err != nil {
		return rosterStatus(err), true
	}
	if !slices.ContainsFunc(members, func(m auth.Member) bool { return m.Admin }) {
		return "no admin; run dubplate-server member add --admin", true
	}
	return "", false
}

// rosterStatus renders an [auth.OpenRoster] refusal as one doctor
// status: "missing" for an absent file, the wall's own word for an
// unset row, and otherwise "invalid: " plus the error with every
// newline an [errors.Join] of several bad records carries folded to
// "; ", so the refusal stays a single row.
func rosterStatus(err error) string {
	if errors.Is(err, auth.ErrRosterMissing) {
		return "missing"
	}
	return "invalid: " + strings.ReplaceAll(err.Error(), "\n", "; ")
}

// printRow writes one doctor line in the pinned "%-32s %s" form.
func printRow(w io.Writer, key, status string) error {
	_, err := fmt.Fprintf(w, "%-32s %s\n", key, status)
	return err
}

// rowStatuses parses the joined error [config.Load] returns into
// per-key status text: "missing" for a row [config.ErrMissing] flags,
// or "invalid: " plus the row's own reason for anything else. A key
// with no entry is unmentioned in err and therefore satisfied.
//
// It flattens nested joins to any depth (review M4): an identity
// source's own Validate hook joins its own failures into one error
// before [config.Load] joins that alongside every other row's, so a
// hook that names two bad keys at once must not collapse into one
// entry naming the first key and swallowing the second. flattenErrors
// does the recursion; a leaf whose own message still carries an
// embedded newline, which no hook in this codebase produces but a
// future one could by joining strings without errors.Join, is split
// per line as a second, cheaper net.
func rowStatuses(err error) map[string]string {
	statuses := make(map[string]string)
	if err == nil {
		return statuses
	}

	for _, leaf := range flattenErrors(err) {
		missing := errors.Is(leaf, config.ErrMissing)
		for line := range strings.SplitSeq(leaf.Error(), "\n") {
			key, reason, ok := strings.Cut(line, ": ")
			if !ok || key == "" {
				continue
			}
			if missing {
				statuses[key] = "missing"
				continue
			}
			statuses[key] = "invalid: " + reason
		}
	}
	return statuses
}

// flattenErrors returns every leaf of err's error tree: err itself when
// it does not implement Unwrap() []error, or the flattened leaves of
// every error it wraps when it does. [errors.Join] returns exactly that
// interface, so this recursion is the whole of what unwraps its own
// nested calls (review M4).
func flattenErrors(err error) []error {
	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return []error{err}
	}
	var leaves []error
	for _, sub := range joined.Unwrap() {
		leaves = append(leaves, flattenErrors(sub)...)
	}
	return leaves
}

// thresholdStatus resolves t's effective value against environ,
// reporting a parse failure the same way an owned row reports one:
// "invalid: " plus the reason. It consults statuses first (review
// M4): a threshold's own Parse* call only ever sees that threshold's
// own value, so a relational refusal a hook raises against it, such as
// R31's retention-vs-expiry comparison, is invisible to Parse* and
// must be read from the flattened [rowStatuses] map instead, the same
// way an owned wall row already is.
//
// Absent such an entry, the switch names every registered kind
// explicitly; its default arm never fires against a threshold the
// registry accepted, because [config.Registry.AddThreshold] refuses an
// unnamed kind at registration, but it reports the registration bug in
// its own words rather than falling through to ParseInt, which would
// misreport a mis-declared threshold as an environment problem
// (inventory surprise 11).
func thresholdStatus(environ []string, statuses map[string]string, t config.Threshold) string {
	if status, ok := statuses[t.Key]; ok {
		return status
	}
	switch t.Kind {
	case config.KindInt:
		v, err := config.ParseInt(environ, t)
		if err != nil {
			reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return strconv.Itoa(v)
	case config.KindDuration:
		d, err := config.ParseDuration(environ, t)
		if err != nil {
			reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return d.String()
	case config.KindFloat:
		f, err := config.ParseFloat(environ, t)
		if err != nil {
			reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		// 'g' with precision -1 prints the shortest decimal that
		// round-trips back to f through strconv.ParseFloat, which is
		// the same parser config.ParseFloat reads a set value through.
		return strconv.FormatFloat(f, 'g', -1, 64)
	case config.KindString:
		v, err := config.ParseString(environ, t)
		if err != nil {
			reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return v
	default:
		return "invalid: registered with an unnamed threshold kind, a registration bug"
	}
}

// identitySourceStatus reports how src stands against environ:
// inactive when the selector names another source, "unsatisfied: "
// plus every failing key when it is the selected one and either
// [config.Load] flagged one of its Required or Optional rows or src's
// own Validate hook refuses the values those rows resolve to, and
// active otherwise. Optional is consulted alongside Required (review
// M4): a short rotation-overlap secret is one of the source's Optional
// keys, and until this consultation it never reflected in this row,
// however plainly the operator's own terminal showed the secret's key
// rejected by name elsewhere. Validate is rerun directly (review M4)
// because a relational refusal, such as R31's retention-vs-expiry
// comparison, is keyed against a threshold's own name rather than one
// of src's Required or Optional keys, so the per-key sweep above cannot
// see it; its failing keys are read through [sourceKeyError] via
// errors.As, never by parsing the rendered message, so a reason that
// itself carries a colon cannot be mistaken for a key. It names no
// value, because a source's required set carries secrets and doctor's
// table is read on a terminal (Resolution R-5b6).
func identitySourceStatus(environ []string, src config.IdentitySource, statuses map[string]string) string {
	if !src.Selected(environ) {
		return "inactive"
	}

	var failing []string
	fail := func(key, word string) {
		line := key + " " + word
		if !slices.Contains(failing, line) {
			failing = append(failing, line)
		}
	}

	for _, key := range src.Required {
		if status := statuses[key]; status != "" {
			fail(key, failingWord(status))
		}
	}
	for _, key := range src.Optional {
		if status := statuses[key]; status != "" {
			fail(key, failingWord(status))
		}
	}

	validateFailed := false
	if src.Validate != nil {
		if err := src.Validate(identitySourceValues(environ, src), environ); err != nil {
			validateFailed = true
			for _, leaf := range flattenErrors(err) {
				if keyed, ok := errors.AsType[*sourceKeyError](leaf); ok {
					fail(keyed.key, "invalid")
				}
			}
		}
	}

	if len(failing) == 0 {
		// A Validate refusal that named no key through
		// [sourceKeyError] must still read unsatisfied: every
		// refusal validateMagicLinkSource can return is keyed today,
		// but a future hook's unkeyed error must never fall through
		// to "active" and hide a real refusal from the table.
		if validateFailed {
			return "unsatisfied"
		}
		return "active"
	}
	return "unsatisfied: " + strings.Join(failing, ", ")
}

// failingWord renders one per-key status as the single word doctor's
// source line carries beside the key: "missing" for [rowStatuses]'s
// "missing" entry, "invalid" for its "invalid: " prefixed entries.
func failingWord(status string) string {
	if status == "missing" {
		return "missing"
	}
	return "invalid"
}

// identitySourceValues resolves src's Required and Optional keys from
// environ, the same construction [config]'s own resolveIdentitySource
// uses against its wall: an unset or empty value is omitted from the
// map rather than recorded as an empty string.
func identitySourceValues(environ []string, src config.IdentitySource) map[string]string {
	values := make(map[string]string, len(src.Required)+len(src.Optional))
	for _, key := range src.Required {
		if v, ok := lookupEnviron(environ, key); ok {
			values[key] = v
		}
	}
	for _, key := range src.Optional {
		if v, ok := lookupEnviron(environ, key); ok {
			values[key] = v
		}
	}
	return values
}

// lookupEnviron returns name's value out of environ (os.Environ()
// form), a later entry winning over an earlier one for the same name,
// and treats an empty value as unset, [config]'s own wall.get rule.
func lookupEnviron(environ []string, name string) (string, bool) {
	value, ok := "", false
	for _, kv := range environ {
		if key, v, found := strings.Cut(kv, "="); found && key == name {
			value, ok = v, true
		}
	}
	if !ok || value == "" {
		return "", false
	}
	return value, true
}

// splitKeyReason strips err's leading "key: " prefix, the shape every
// [config] load and parse error uses, and returns the reason that
// follows it. A message with no such separator is returned whole.
func splitKeyReason(err error) (reason string) {
	_, reason, ok := strings.Cut(err.Error(), ": ")
	if !ok {
		return err.Error()
	}
	return reason
}
