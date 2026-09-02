package main

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// newDoctorCmd builds the doctor subcommand: it reports the config
// wall's owned rows, reserved classes, and registered thresholds, one
// line each in "%-32s %s" form, deriving every row from
// internal/config's own tables so no rung hard-codes a count. That
// table is the command's human output; on an unsatisfied wall RunE
// returns a single short error, leaving main as the sole stderr
// writer. Exit status follows [config.Load]'s own verdict on
// environ: 0 when the owned wall is satisfied, 1 otherwise.
func newDoctorCmd(environ []string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Report the configuration wall",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			_, err := config.Load(environ)
			statuses := rowStatuses(err)

			for _, row := range config.OwnedWall() {
				status := "ok"
				switch {
				case row.Pending != "":
					status = row.Pending
				case statuses[row.Name] != "":
					status = statuses[row.Name]
				}
				if werr := printRow(stdout, row.Name, status); werr != nil {
					return werr
				}
			}

			for _, class := range config.ReservedClasses() {
				status := fmt.Sprintf("disabled (arrives at rung %d)", class.Rung)
				if werr := printRow(stdout, class.Name, status); werr != nil {
					return werr
				}
			}

			for _, th := range config.Thresholds() {
				if werr := printRow(stdout, th.Key, thresholdStatus(environ, th)); werr != nil {
					return werr
				}
			}

			if err != nil {
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
	v, err := config.ParseInt(environ, t)
	if err != nil {
		_, reason := splitKeyReason(err)
		return "invalid: " + reason
	}
	return strconv.Itoa(v)
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
