package main

import (
	"cmp"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
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
// names no admin or sits in a directory the process cannot write, the
// store row reports a failing state, the auth store row finds a leaf
// serve would refuse or could not write, the backup remote row fails,
// the unidecode row finds the transliteration table missing or
// invalid, the client ip row refuses the trusted proxy values, the web
// row refuses what serve startup would refuse, the
// bind addresses row finds a configured address absent or cannot
// check it, the boundary probe row finds a header the strip should
// have removed or an answer it cannot read, or a registered setting's
// row refuses; 0 otherwise.
func newDoctorCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "doctor",
		Short:         "Report the configuration wall",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, loadErr := config.Load(environ, reg)
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

			// The store row runs the inspect mode after every other
			// row, always printing exactly one line, absent state
			// included: a store the doctor run cannot vouch for is as
			// reportable as a missing config key.
			storeStatus, storeFailed, storeReport, storeChecked := storeRow(cmd.Context(), environ)
			if err := printRow(stdout, "store", storeStatus); err != nil {
				return err
			}

			// The members row prints from the report the store row
			// read, so doctor inspects once, and always prints exactly
			// one line.
			membersStatus, membersFailed := membersRowStatus(storeReport, storeChecked)
			if err := printRow(stdout, membersDoctorRow, membersStatus); err != nil {
				return err
			}

			// The auth store row follows the store row and always
			// prints exactly one line: it reads "not checked" when
			// Load refused environ, since the state directory is then
			// unknown.
			authStoreStatus, authStoreFailed := authStoreRowStatus(cfg)
			if err := printRow(stdout, authStoreDoctorRow, authStoreStatus); err != nil {
				return err
			}

			// The backup rows (R-6bi13, R-6bi14, Ruling 2) print
			// unconditionally, the store row's own precedent: the facts
			// rows read "not checked" rather than disappear when the
			// store cannot supply them, and the remote row reads
			// "unavailable" rather than disappear while the flag is off.
			for _, row := range backupFactsRows(time.Now(), storeReport, storeChecked) {
				if err := printRow(stdout, row.key, row.status); err != nil {
					return err
				}
			}
			backupRemoteStatus, backupRemoteFailed := backupRemoteRowStatus(cmd.Context(), environ)
			if err := printRow(stdout, backupRemoteRow, backupRemoteStatus); err != nil {
				return err
			}

			// The unidecode row follows the backup rows, the store
			// row's own precedent: it always prints exactly one line,
			// since the matcher can never score without a verified
			// table.
			unidecodeStatus, unidecodeFailed := unidecodeRowStatus(environ)
			if err := printRow(stdout, unidecodeDoctorRow, unidecodeStatus); err != nil {
				return err
			}

			// The four boundary rows follow the unidecode row and
			// precede the settings rows, so every fixed-count row
			// comes before the one variable-count family. Each
			// prints exactly one line whatever its state: client ip
			// and web read "not checked" when Load refused environ.
			// The bind addresses row calls [net.InterfaceAddrs]
			// once and hands the status function both results. The
			// boundary probe row dials each bind address on the
			// serve port, each dial and answer bounded.
			clientIPStatus, clientIPFailed := clientIPRowStatus(cfg)
			if err := printRow(stdout, clientIPDoctorRow, clientIPStatus); err != nil {
				return err
			}
			webStatus, webFailed := webRowStatus(environ, cfg)
			if err := printRow(stdout, webDoctorRow, webStatus); err != nil {
				return err
			}
			ifaceAddrs, ifaceErr := net.InterfaceAddrs()
			bindStatus, bindFailed := bindAddrsRowStatus(environ, ifaceAddrs, ifaceErr)
			if err := printRow(stdout, bindAddrsDoctorRow, bindStatus); err != nil {
				return err
			}
			probeStatus, probeFailed := boundaryProbeRowStatus(cmd.Context(), environ, reg, cfg)
			if err := printRow(stdout, boundaryProbeDoctorRow, probeStatus); err != nil {
				return err
			}

			// The settings rows print one per
			// registered setting, in declaration order, after the
			// boundary rows, and one row per orphan the store holds
			// under a key no rung still declares.
			settingsRows, settingsFailed := settingsDoctorRows(environ, reg, storeReport, storeChecked)
			for _, row := range settingsRows {
				if err := printRow(stdout, row.key, row.status); err != nil {
					return err
				}
			}

			if loadErr != nil || storeFailed || membersFailed || authStoreFailed || backupRemoteFailed || unidecodeFailed ||
				clientIPFailed || webFailed || bindFailed || probeFailed || settingsFailed {
				return errors.New("configuration is incomplete")
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// authStoreDoctorRow is the row reporting the auth leaf serve opens.
const authStoreDoctorRow = "auth store"

// membersDoctorRow is the row reporting the members the store holds.
const membersDoctorRow = "members"

// membersRowStatus renders the members row from the report the store
// row read, and reports whether the row fails. It decides on the
// report's StoreVersion, never on whether Facts is present, since a
// version 1 store carries no facts while a version 2 store carries
// facts without the users table. A store the store row cannot vouch
// for reads "not read"; the row fails only for a store that holds the
// users table and no live admin, since nobody could then add a member
// through the page.
func membersRowStatus(r store.Report, checked bool) (status string, failed bool) {
	switch {
	case !checked:
		return "not read", false
	case r.State == store.StateAbsent:
		return "no store yet; start serve, then member add --admin", false
	case !storeVouches(r.State):
		return "not read", false
	case r.StoreVersion < store.UsersVersion:
		return fmt.Sprintf("members not recorded before store version %d", store.UsersVersion), false
	case r.Facts == nil:
		return "not read", false
	case r.Facts.Members == 0:
		return "no members; run dubplate-server member add --admin", true
	case r.Facts.Admins == 0:
		return fmt.Sprintf("members: %d, admins: 0; run dubplate-server member add --admin", r.Facts.Members), true
	}
	return fmt.Sprintf("members: %d, admins: %d", r.Facts.Members, r.Facts.Admins), false
}

// authStoreLeafPerm is the widest mode [auth.Open] accepts on the auth
// leaf; a wider leaf refuses serve's start.
const authStoreLeafPerm = 0o700

// authStoreSubdirs are the two record directories [auth.Open] creates
// under the leaf, the nonce store's and the epoch store's.
var authStoreSubdirs = []string{"nonce", "epoch"}

// authStoreRowStatus reports, without writing, whether serve could
// open the auth leaf under cfg's state directory: "ok" when the leaf
// is at most 0700, owned by the effective uid, and each record
// directory is writable; "absent; serve creates it" when the leaf is
// missing and its nearest existing ancestor is writable. Otherwise it
// names every problem and fails: the leaf's mode in the words
// [auth.Open] refuses it with, a foreign owner, or a directory the
// process cannot write, the shape a path outside the unit's
// ReadWritePaths takes. A nil cfg reads "not checked" and adds no
// failure of its own, since Load's own report already fails doctor.
func authStoreRowStatus(cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	leaf := auth.StoreDir(cfg.StateDir)
	info, err := os.Stat(leaf)
	if errors.Is(err, fs.ErrNotExist) {
		if err := checkCreatable(leaf); err != nil {
			return "absent, and serve cannot create it: " + err.Error(), true
		}
		return "absent; serve creates it", false
	}
	if err != nil {
		return err.Error(), true
	}
	if !info.IsDir() {
		return fmt.Sprintf("leaf %s is not a directory", leaf), true
	}

	var problems []string
	if mode := info.Mode().Perm(); mode&^authStoreLeafPerm != 0 {
		problems = append(problems, fmt.Sprintf("leaf %s has mode %o, want at most %o", leaf, mode, authStoreLeafPerm))
	}
	// io/fs.FileInfo documents "Sys() any // underlying data source
	// (can return nil)"; on Linux, os.Stat's carries a *syscall.Stat_t.
	if st, ok := info.Sys().(*syscall.Stat_t); !ok {
		problems = append(problems, fmt.Sprintf("leaf %s: cannot determine its owner", leaf))
	} else if euid := geteuid(); int(st.Uid) != euid {
		problems = append(problems, fmt.Sprintf("leaf %s is owned by uid %d, want uid %d", leaf, st.Uid, euid))
	}
	for _, name := range authStoreSubdirs {
		sub := filepath.Join(leaf, name)
		_, err := os.Stat(sub)
		switch {
		case errors.Is(err, fs.ErrNotExist):
			if err := checkDirWritable(leaf); err != nil {
				problems = append(problems, fmt.Sprintf("%s is absent, and %v", sub, err))
			}
		case err != nil:
			problems = append(problems, err.Error())
		default:
			if err := checkDirWritable(sub); err != nil {
				problems = append(problems, err.Error())
			}
		}
	}
	if len(problems) > 0 {
		return strings.Join(problems, "; "), true
	}
	return "ok", false
}

// checkCreatable reports whether [os.MkdirAll] could create path: it
// walks up to path's nearest existing ancestor and checks that
// ancestor with [checkDirWritable].
func checkCreatable(path string) error {
	dir := filepath.Dir(path)
	for {
		_, err := os.Stat(dir)
		if err == nil {
			return checkDirWritable(dir)
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return err
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return err
		}
		dir = parent
	}
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
// table is read on a terminal: a Required or Optional key's reason is
// printed only when it does not quote the value, through
// [sourceKeyWord].
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

	values := identitySourceValues(environ, src)
	for _, key := range slices.Concat(src.Required, src.Optional) {
		if status := statuses[key]; status != "" {
			fail(key, sourceKeyWord(status, values[key]))
		}
	}

	validateFailed := false
	if src.Validate != nil {
		if err := src.Validate(values, environ); err != nil {
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

// sourceKeyWord renders one per-key status as the text doctor's source
// line carries beside the key: "missing" for [rowStatuses]'s "missing"
// entry, and "invalid" for its "invalid: " prefixed entries, followed
// by the reason in parentheses whenever the reason does not quote
// value, the key's configured value. A source's own keys have no row
// of their own, so this is the one place doctor can show the reason
// check prints; a reason that does quote the value is withheld, since
// the key may carry a secret and doctor's table is read on a terminal.
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
