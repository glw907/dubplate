package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/serve"
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

// runDoctor runs every doctor probe, then prints the rows in wall
// order. It returns an error when [config.Load] refuses environ or any
// printed row reports a failure.
func runDoctor(ctx context.Context, environ []string, reg *config.Registry, stdout io.Writer) error {
	// Load and the staging rule run apart from loadConfig so a nesting
	// lands on the staging row as "invalid" and the report carries on
	// with the config Load resolved.
	cfg, loadErr := config.Load(environ, reg)
	if loadErr == nil {
		loadErr = intake.ValidateStaging(cfg)
	}
	statuses := rowStatuses(loadErr)
	report := doctorReport{failed: loadErr != nil}

	// The store is inspected first and its row prints later: the admin
	// email row reads the roster standing the same inspection reports.
	storeStatus, storeFailed, storeReport, storeChecked := store.DoctorRow(ctx, environ)

	report.addRegistryRows(environ, reg, statuses, storeReport, storeChecked)

	// Each fixed row prints exactly one line whatever its state, so a
	// store or a config doctor cannot vouch for is as reportable as a
	// missing key.
	report.addChecked("store", storeStatus, storeFailed)
	report.addProbeRows(ctx, environ, reg, cfg, storeReport, storeChecked)

	// One row per registered setting, then one per orphan the store
	// holds under a key nothing declares.
	settingsRows, settingsFailed := settingsDoctorRows(environ, reg, storeReport, storeChecked)
	report.rows = append(report.rows, settingsRows...)
	report.failed = report.failed || settingsFailed

	for _, row := range report.rows {
		if err := printRow(stdout, row.key, row.status); err != nil {
			return err
		}
	}
	if report.failed {
		return errors.New("configuration is incomplete")
	}
	return nil
}

// doctorReport collects the doctor rows in print order, and whether any
// of them reports a failure. Rows print once every probe has run.
type doctorReport struct {
	rows   []doctorRow
	failed bool
}

func (r *doctorReport) add(key, status string) {
	r.rows = append(r.rows, doctorRow{key, status})
}

func (r *doctorReport) addChecked(key, status string, failed bool) {
	r.add(key, status)
	r.failed = r.failed || failed
}

// addRegistryRows adds the rows reg declares: the owned wall, the
// reserved, integration, threshold, and identity-source classes, and a
// row for each load error no declared row claims. None of them marks
// the report failed, since the load error behind statuses already has.
func (r *doctorReport) addRegistryRows(environ []string, reg *config.Registry, statuses map[string]string, storeReport store.Report, storeChecked bool) {
	for _, row := range reg.OwnedWall() {
		status := cmp.Or(statuses[row.Name], "ok")
		if row.Name == adminEmailKey {
			status = auth.AdminAddressRowStatus(status, storeReport, storeChecked)
		}
		r.add(row.Name, status)
	}

	for _, class := range reg.ReservedClasses() {
		r.add(class.Name, fmt.Sprintf("disabled (arrives at rung %d)", class.Rung))
	}

	for _, class := range reg.IntegrationClasses() {
		status := "disabled"
		if enabled, err := class.Enabled(environ); err == nil && enabled {
			status = "enabled"
		}
		r.add(class.Name, status)
	}

	for _, th := range reg.Thresholds() {
		r.add(th.Key, thresholdStatus(environ, statuses, th))
	}

	for _, src := range reg.IdentitySources() {
		r.add(src.Name, identitySourceStatus(environ, src, statuses))
	}

	// A load error keyed by a name no printed row claims would leave the
	// exit status unexplained, so each prints on a line of its own.
	for _, key := range unclaimedLoadErrors(reg, statuses) {
		r.add(key, statuses[key])
	}
}

// addProbeRows adds the fixed rows that probe the store's members, the
// auth leaf, the backups, the unidecode table, the network boundary,
// and the music volume, in that order.
func (r *doctorReport) addProbeRows(ctx context.Context, environ []string, reg *config.Registry, cfg *config.Config, storeReport store.Report, storeChecked bool) {
	membersStatus, membersFailed := auth.MembersRowStatus(storeReport, storeChecked)
	r.addChecked(membersDoctorRow, membersStatus, membersFailed)

	authDirStatus, authDirFailed := auth.DirRowStatus(cfg, auth.DirRowOptions{Geteuid: os.Geteuid})
	r.addChecked(authDirDoctorRow, authDirStatus, authDirFailed)

	for _, row := range backupFactsRows(time.Now(), storeReport, storeChecked) {
		r.add(row.key, row.status)
	}
	backupRemoteStatus, backupRemoteFailed := backup.RemoteRowStatus(ctx, environ, backup.RemoteRowOptions{FeaturesTimeout: backupFeaturesTimeout})
	r.addChecked(backupRemoteRow, backupRemoteStatus, backupRemoteFailed)

	unidecodeStateDir, _ := config.Lookup(environ, "DUBPLATE_STATE_DIR")
	unidecodeStatus, unidecodeFailed := unidecode.RowStatus(unidecodeStateDir)
	r.addChecked(unidecodeDoctorRow, unidecodeStatus, unidecodeFailed)

	clientIPStatus, clientIPFailed := serve.ClientIPRowStatus(cfg)
	r.addChecked(clientIPDoctorRow, clientIPStatus, clientIPFailed)
	webStatus, webFailed := webapp.RowStatus(cfg, webRowResolve(environ))
	r.addChecked(webDoctorRow, webStatus, webFailed)
	ifaceAddrs, ifaceErr := net.InterfaceAddrs()
	bindStatus, bindFailed := serve.BindAddrsRowStatus(environ, ifaceAddrs, ifaceErr)
	r.addChecked(bindAddrsDoctorRow, bindStatus, bindFailed)
	probeStatus, probeFailed := serve.BoundaryProbeRowStatus(ctx, environ, reg, cfg)
	r.addChecked(boundaryProbeDoctorRow, probeStatus, probeFailed)

	var collectionRoots [][]byte
	if storeChecked && storeReport.Facts != nil {
		collectionRoots = storeReport.Facts.CollectionRoots
	}
	volumeStatus, volumeFailed := intake.VolumeRowStatus(cfg, collectionRoots)
	r.addChecked(musicVolumeDoctorRow, volumeStatus, volumeFailed)
}

// unclaimedLoadErrors returns, sorted, the DUBPLATE_ keys statuses
// reports that no owned row, threshold, or identity source reads its
// status from.
func unclaimedLoadErrors(reg *config.Registry, statuses map[string]string) []string {
	claimed := make(map[string]bool)
	for _, row := range reg.OwnedWall() {
		claimed[row.Name] = true
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

// adminEmailKey is the owned wall row naming the operator's address.
const adminEmailKey = "DUBPLATE_ADMIN_EMAIL"

// The boundary rows, fixed rather than declared in the registry.
const (
	clientIPDoctorRow      = "client ip"
	webDoctorRow           = "web"
	bindAddrsDoctorRow     = "bind addresses"
	boundaryProbeDoctorRow = "boundary probe"
)

// unidecodeDoctorRow is the row reporting the unidecode table.
const unidecodeDoctorRow = "unidecode"

// The backup rows, apart from the integration-class loop's "backup"
// row, which reads enabled or disabled.
const (
	backupRemoteRow      = "backup remote"
	backupLastSuccessRow = "backup last success"
	backupWALSizeRow     = "backup wal size"
	backupFreeSpaceRow   = "backup free space"
	backupLocalCopiesRow = "backup local copies"
)

// musicVolumeDoctorRow is the row reporting the music volume: staging,
// the library root, and the collection roots under it.
const musicVolumeDoctorRow = "music volume"

// authDirDoctorRow is the row reporting the auth leaf serve opens.
const authDirDoctorRow = "auth dir"

// membersDoctorRow is the row reporting the members the store holds.
const membersDoctorRow = "members"

// doctorRow is one doctor line: a printRow key and its status text.
type doctorRow struct {
	key, status string
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
