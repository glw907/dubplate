package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// newDoctorCmd builds the doctor subcommand: it reports the config
// wall's owned rows, reserved classes, integration classes, and
// registered thresholds, one line each in "%-32s %s" form, deriving
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

			// IntegrationClasses stays a package-level read: rung 4's
			// class has not migrated onto the registry yet (see
			// internal/config's integrationClassOf doc), matching
			// config.Load's own read path above.
			for _, class := range config.IntegrationClasses() {
				status := "disabled"
				if enabled, err := class.Enabled(environ); err == nil && enabled {
					status = "enabled"
				}
				if err := printRow(stdout, class.Name, status); err != nil {
					return err
				}
			}

			for _, th := range reg.Thresholds() {
				if err := printRow(stdout, th.Key, thresholdStatus(environ, th)); err != nil {
					return err
				}
			}

			if loadErr != nil {
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
func rowStatuses(err error) map[string]string {
	statuses := make(map[string]string)
	if err == nil {
		return statuses
	}

	joined, ok := err.(interface{ Unwrap() []error })
	if !ok {
		return statuses
	}
	for _, sub := range joined.Unwrap() {
		key, reason := splitKeyReason(sub)
		if key == "" {
			continue
		}
		if errors.Is(sub, config.ErrMissing) {
			statuses[key] = "missing"
			continue
		}
		statuses[key] = "invalid: " + reason
	}
	return statuses
}

// thresholdStatus resolves t's effective value against environ,
// reporting a parse failure the same way an owned row reports one:
// "invalid: " plus the reason.
func thresholdStatus(environ []string, t config.Threshold) string {
	switch t.Kind {
	case config.KindDuration:
		d, err := config.ParseDuration(environ, t)
		if err != nil {
			_, reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return d.String()
	default:
		v, err := config.ParseInt(environ, t)
		if err != nil {
			_, reason := splitKeyReason(err)
			return "invalid: " + reason
		}
		return strconv.Itoa(v)
	}
}

// splitKeyReason splits err's message on its first "key: reason"
// separator, the shape every [config] load and parse error uses. A
// message with no such separator returns an empty key and the whole
// message as reason.
func splitKeyReason(err error) (key, reason string) {
	key, reason, ok := strings.Cut(err.Error(), ": ")
	if !ok {
		return "", err.Error()
	}
	return key, reason
}
