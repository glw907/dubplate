package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
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
// short error, leaving main as the sole stderr writer. Exit status
// follows [config.Load]'s own verdict on environ: 0 when the owned
// wall is satisfied, 1 otherwise.
func newDoctorCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Report the configuration wall",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
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
			// through, so a duplicate address, a duplicate
			// streamer_username, or an unparseable record is reported
			// here too. [auth.ErrRosterMissing] is not reported here:
			// R37's own runbook orders a fresh install as deploy, then
			// doctor, then the CLI bootstrap that creates the file, so
			// "no roster yet" is doctor's expected pre-bootstrap state
			// rather than a broken configuration. A registry with no
			// roster-backed identity source, or a roster that has not
			// been created yet, prints no extra row, so R10's
			// line-count arithmetic stays keyed to reg alone whenever
			// there is nothing new to report. It names no value
			// (Resolution R-5b6): [auth.OpenRoster]'s own errors
			// already carry only a record's position and the path,
			// never its address or principal.
			rosterFailed := false
			if path, ok := lookupEnviron(environ, "DUBPLATE_ROSTER_FILE"); ok {
				if _, err := auth.OpenRoster(path); err != nil && !errors.Is(err, auth.ErrRosterMissing) {
					rosterFailed = true
					if err := printRow(stdout, "roster", "invalid: "+err.Error()); err != nil {
						return err
					}
				}
			}

			if loadErr != nil || rosterFailed {
				return errors.New("configuration is incomplete")
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
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
// inactive when the selector names another source, unsatisfied when it
// is the selected one and either [config.Load] flagged one of its
// Required or Optional rows or src's own Validate hook refuses the
// values those rows resolve to, and active otherwise. Optional is
// consulted alongside Required (review M4): a short rotation-overlap
// secret is one of the source's Optional keys, and until this
// consultation it never reflected in this row, however plainly the
// operator's own terminal showed the secret's key rejected by name
// elsewhere. Validate is rerun directly (review M4) because a
// relational refusal, such as R31's retention-vs-expiry comparison, is
// keyed against a threshold's own name rather than one of src's
// Required or Optional keys, so the per-key sweep above cannot see it.
// It names no value, because a source's required set carries secrets
// and doctor's table is read on a terminal (Resolution R-5b6).
func identitySourceStatus(environ []string, src config.IdentitySource, statuses map[string]string) string {
	if !src.Selected(environ) {
		return "inactive"
	}
	for _, key := range src.Required {
		if statuses[key] != "" {
			return "unsatisfied"
		}
	}
	for _, key := range src.Optional {
		if statuses[key] != "" {
			return "unsatisfied"
		}
	}
	if src.Validate != nil {
		if err := src.Validate(identitySourceValues(environ, src), environ); err != nil {
			return "unsatisfied"
		}
	}
	return "active"
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
