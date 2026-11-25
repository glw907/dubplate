package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/ops/jobs"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
	"github.com/glw907/dubplate/internal/webapp"
	"github.com/spf13/cobra"
)

// newDoctorCmd builds the doctor subcommand: it prints one line per row
// of the config wall, one block per family, deriving every row it can
// from reg. The exit status is 1 when [config.Load] refuses environ or
// any row it prints reports a failure, and 0 otherwise.
func newDoctorCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Report the configuration wall",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runDoctor(cmd.Context(), environ, reg, stdout)
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// runDoctor runs every doctor probe, then prints the rows in order. It
// returns an error when [config.Load] refuses environ or any printed
// row reports a failure.
func runDoctor(ctx context.Context, environ []string, reg *config.Registry, stdout io.Writer) error {
	// loadWall rather than loadConfig, so a refused check lands on its
	// row as "invalid" and the report carries on with the config Load
	// resolved.
	cfg, loadErr := loadWall(environ, reg)

	// One store inspection reads every fact the rows report, each
	// package's reader filling the facts that package owns.
	var (
		roster  member.Facts
		volume  intake.Facts
		stored  settings.Facts
		history jobs.Facts
	)
	storeRow, report := store.Doctor(ctx, environ,
		roster.Reader(environ), volume.Reader(), stored.Reader(), history.Reader())

	wall := registryRows(environ, reg, rowStatuses(loadErr))
	for i, row := range wall {
		wall[i] = member.RefineAdminEmail(row, report, roster)
	}
	stateDir, _ := config.Lookup(environ, "DUBPLATE_STATE_DIR")
	rows := slices.Concat(
		wall,
		[]doctor.Row{storeRow},
		member.DoctorRows(report, roster),
		auth.DoctorRows(cfg),
		backup.DoctorRows(ctx, environ, time.Now(), report, history.LastSuccess),
		unidecode.DoctorRows(stateDir),
		serve.DoctorRows(ctx, environ, reg, cfg),
		webapp.DoctorRows(cfg, webRowResolve(environ)),
		intake.DoctorRows(cfg, volume),
		settings.DoctorRows(environ, reg, report, stored),
	)

	failed := loadErr != nil
	for _, row := range rows {
		if err := printRow(stdout, row.Name, row.Status); err != nil {
			return err
		}
		failed = failed || row.Failed
	}
	if failed {
		return errors.New("configuration is incomplete")
	}
	return nil
}

// registryRows returns the rows reg declares: the owned wall, the
// reserved, integration, threshold, and identity-source classes, and a
// row for each load error no declared row claims. None of them fails,
// since the load error behind statuses already has.
func registryRows(environ []string, reg *config.Registry, statuses map[string]string) []doctor.Row {
	var rows []doctor.Row
	add := func(name, status string) {
		rows = append(rows, doctor.Row{Name: name, Status: status})
	}
	for _, row := range reg.OwnedWall() {
		add(row.Key, cmp.Or(statuses[row.Key], "ok"))
	}

	for _, class := range reg.ReservedClasses() {
		add(class.Name, "not yet available")
	}

	for _, class := range reg.IntegrationClasses() {
		add(class.Name, integrationStatus(environ, class))
	}

	for _, th := range reg.Thresholds() {
		add(th.Key, thresholdStatus(environ, statuses, th))
	}

	for _, src := range reg.IdentitySources() {
		add(src.Name, identitySourceStatus(environ, src, statuses))
	}

	// A load error keyed by a name no printed row claims would leave the
	// exit status unexplained, so each prints on a line of its own.
	for _, key := range unclaimedLoadErrors(reg, statuses) {
		add(key, statuses[key])
	}
	return rows
}

// unclaimedLoadErrors returns, sorted, the DUBPLATE_ keys statuses
// reports that no owned row, threshold, or identity source reads its
// status from.
func unclaimedLoadErrors(reg *config.Registry, statuses map[string]string) []string {
	claimed := make(map[string]bool)
	for _, row := range reg.OwnedWall() {
		claimed[row.Key] = true
	}
	for _, th := range reg.Thresholds() {
		claimed[th.Key] = true
	}
	for _, src := range reg.IdentitySources() {
		for _, key := range slices.Concat([]string{src.Selector}, config.KeyNames(src.Required), config.KeyNames(src.Optional)) {
			claimed[key] = true
		}
	}
	var keys []string
	for key := range statuses {
		if strings.HasPrefix(key, "DUBPLATE_") && !claimed[key] {
			keys = append(keys, key)
		}
	}
	slices.Sort(keys)
	return keys
}

// printRow writes one doctor line in the pinned "%-32s %s" form.
func printRow(w io.Writer, key, status string) error {
	_, err := fmt.Fprintf(w, "%-32s %s\n", key, status)
	return err
}

// rowStatuses maps the joined error [config.Load] returns, and the
// staging check's, to per-key status text: "missing" for a row
// [config.ErrMissing] flags, or "invalid: " plus the row's reason for
// anything else. A leaf with no [config.KeyError] names no row and gets
// no entry, and a key with no entry is satisfied.
//
// It flattens nested joins to any depth: an identity source's Validate
// hook joins its failures before [config.Load] joins them with every
// other row's.
func rowStatuses(err error) map[string]string {
	statuses := make(map[string]string)
	if err == nil {
		return statuses
	}

	for _, leaf := range flattenErrors(err) {
		keyed, ok := errors.AsType[*config.KeyError](leaf)
		if !ok {
			continue
		}
		if errors.Is(keyed, config.ErrMissing) {
			statuses[keyed.Key] = "missing"
			continue
		}
		statuses[keyed.Key] = "invalid: " + keyed.Reason
	}
	return statuses
}

// flattenErrors returns every leaf of err's tree of [errors.Join]
// errors.
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

// thresholdStatus resolves t's effective value against environ, and
// reports a parse failure as "invalid: " plus the reason. It consults
// statuses first, because a relational refusal a hook raises against a
// threshold, such as the retention against expiry comparison, is
// invisible to that threshold's own Parse call.
func thresholdStatus(environ []string, statuses map[string]string, t config.Threshold) string {
	if status, ok := statuses[t.Key]; ok {
		return status
	}
	var (
		value string
		err   error
	)
	switch t.Kind {
	case config.KindInt:
		var v int
		v, err = config.ParseInt(environ, t)
		value = strconv.Itoa(v)
	case config.KindDuration:
		var d time.Duration
		d, err = config.ParseDuration(environ, t)
		value = d.String()
	case config.KindFloat:
		var f float64
		f, err = config.ParseFloat(environ, t)
		value = strconv.FormatFloat(f, 'g', -1, 64)
	case config.KindString:
		value, err = config.ParseString(environ, t)
	default:
		return "invalid: registered with an unnamed threshold kind, a registration bug"
	}
	if err != nil {
		if keyed, ok := errors.AsType[*config.KeyError](err); ok {
			return "invalid: " + keyed.Reason
		}
		return "invalid: " + err.Error()
	}
	return value
}

// integrationStatus reports class as "enabled" or "disabled", or as
// "invalid: " plus the refusal when its flag will not parse, the value
// [config.Load] refuses rather than reading as off.
func integrationStatus(environ []string, class config.IntegrationClass) string {
	enabled, err := class.Enabled(environ)
	switch {
	case err != nil:
		if keyed, ok := errors.AsType[*config.KeyError](err); ok {
			return "invalid: " + keyed.Reason
		}
		return "invalid: " + err.Error()
	case enabled:
		return "enabled"
	default:
		return "disabled"
	}
}

// identitySourceStatus reports how src stands against environ: inactive
// when the selector names another source, "unsatisfied: " plus every
// failing key when it is selected and [config.Load] flagged one of its
// Required or Optional rows or its Validate hook refuses their values,
// and active otherwise. Validate is rerun because a relational refusal
// is keyed against a threshold's name, which the per-key sweep cannot
// see. It names no value, since a source's keys carry secrets and the
// table is read on a terminal; [sourceKeyWord] withholds a reason that
// quotes one.
func identitySourceStatus(environ []string, src config.IdentitySource, statuses map[string]string) string {
	if !src.Selected(environ) {
		return "inactive"
	}

	// Each key is named once, by its first report: the per-key sweep
	// below runs before the Validate rerun, so a key both name keeps
	// the sweep's reason.
	var failing []string
	named := make(map[string]bool)
	fail := func(key, word string) {
		if named[key] {
			return
		}
		named[key] = true
		failing = append(failing, key+" "+word)
	}

	// An unset or empty value stays out of values, as the wall's own
	// identity-source resolution leaves it.
	values := make(map[string]string, len(src.Required)+len(src.Optional))
	for _, key := range config.KeyNames(slices.Concat(src.Required, src.Optional)) {
		if v, ok := config.Lookup(environ, key); ok {
			values[key] = v
		}
		if status := statuses[key]; status != "" {
			fail(key, sourceKeyWord(status, values[key]))
		}
	}

	validateFailed := false
	if src.Validate != nil {
		if err := src.Validate(values, environ); err != nil {
			validateFailed = true
			for _, leaf := range flattenErrors(err) {
				if keyed, ok := errors.AsType[*config.KeyError](leaf); ok {
					fail(keyed.Key, "invalid")
				}
			}
		}
	}

	if len(failing) == 0 {
		// A Validate refusal whose leaves carry no [config.KeyError]
		// names no key, and still reads unsatisfied rather than
		// falling through to "active" and hiding a real refusal from
		// the table.
		if validateFailed {
			return "unsatisfied"
		}
		return "active"
	}
	return "unsatisfied: " + strings.Join(failing, ", ")
}

// sourceKeyWord renders one per-key status as the word the source line
// carries beside the key: "missing", or "invalid" followed by the
// reason in parentheses. A reason that quotes value is withheld, because
// the value may be a secret.
func sourceKeyWord(status, value string) string {
	if status == "missing" {
		return "missing"
	}
	reason, ok := strings.CutPrefix(status, "invalid: ")
	if !ok || reason == "" || (value != "" && strings.Contains(reason, value)) {
		return "invalid"
	}
	return "invalid (" + reason + ")"
}
