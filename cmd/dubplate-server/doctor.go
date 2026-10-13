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
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

// newDoctorCmd builds the doctor subcommand: it reports the config
// wall's owned rows, reserved classes, integration classes,
// registered thresholds, and declared identity sources, one line each
// in "%-32s %s" form and one block per family in that order, deriving
// every row from reg so no rung hard-codes a count. That table is the
// command's human output; on an unsatisfied wall RunE returns a single
// short error, leaving main as the sole stderr writer. Exit status is 1
// when [config.Load] refuses environ, the members row finds a store
// with no live admin, the store row reports a failing state, the auth
// store row finds a leaf serve would refuse or could not write, the backup remote row fails,
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
			// Load and the staging nesting rule run apart rather than
			// through loadConfig, so the nesting lands on the staging
			// row as "invalid" and the report carries on with the
			// config Load resolved.
			cfg, loadErr := config.Load(environ, reg)
			if loadErr == nil {
				loadErr = resolveStaging(cfg)
			}
			statuses := rowStatuses(loadErr)

			// Every family appends its rows, in print order, and ORs its
			// own failure into failed. One loop prints them once every
			// probe has run, so a row is added in one place.
			var rows []doctorRow
			failed := loadErr != nil
			add := func(key, status string) {
				rows = append(rows, doctorRow{key, status})
			}
			addChecked := func(key, status string, rowFailed bool) {
				add(key, status)
				failed = failed || rowFailed
			}

			// The store is inspected before any row is added, though its
			// row prints later: the admin email row's status reads the
			// roster standing the same inspection reports, so doctor
			// still inspects once and the print order is unchanged.
			storeStatus, storeFailed, storeReport, storeChecked := storeRow(cmd.Context(), environ)

			for _, row := range reg.OwnedWall() {
				status := cmp.Or(statuses[row.Name], "ok")
				if row.Name == adminEmailKey {
					status = adminEmailRowStatus(status, storeReport, storeChecked)
				}
				add(row.Name, status)
			}

			for _, class := range reg.ReservedClasses() {
				add(class.Name, fmt.Sprintf("disabled (arrives at rung %d)", class.Rung))
			}

			for _, class := range reg.IntegrationClasses() {
				status := "disabled"
				if enabled, err := class.Enabled(environ); err == nil && enabled {
					status = "enabled"
				}
				add(class.Name, status)
			}

			for _, th := range reg.Thresholds() {
				add(th.Key, thresholdStatus(environ, statuses, th))
			}

			for _, src := range reg.IdentitySources() {
				add(src.Name, identitySourceStatus(environ, src, statuses))
			}

			// A load error keyed by a name no printed row claims (a
			// superseded, reserved-class, or unknown key) would leave
			// the exit status unexplained, so each prints on a line of
			// its own, and only when one exists.
			for _, key := range unclaimedLoadErrors(reg, statuses) {
				add(key, statuses[key])
			}

			// The store row follows every other row, always printing
			// exactly one line, absent state included: a store the
			// doctor run cannot vouch for is as reportable as a missing
			// config key.
			addChecked("store", storeStatus, storeFailed)

			// The members row prints from the report the store row
			// read, so doctor inspects once, and always prints exactly
			// one line.
			membersStatus, membersFailed := membersRowStatus(storeReport, storeChecked)
			addChecked(membersDoctorRow, membersStatus, membersFailed)

			// The auth store row follows the store row and always
			// prints exactly one line: it reads "not checked" when
			// Load refused environ, since the state directory is then
			// unknown.
			authStoreStatus, authStoreFailed := authStoreRowStatus(cfg)
			addChecked(authStoreDoctorRow, authStoreStatus, authStoreFailed)

			// The backup rows (R-6bi13, R-6bi14, Ruling 2) print
			// unconditionally, the store row's own precedent: the facts
			// rows read "not checked" rather than disappear when the
			// store cannot supply them, and the remote row reads
			// "unavailable" rather than disappear while the flag is off.
			for _, row := range backupFactsRows(time.Now(), storeReport, storeChecked) {
				add(row.key, row.status)
			}
			backupRemoteStatus, backupRemoteFailed := backupRemoteRowStatus(cmd.Context(), environ)
			addChecked(backupRemoteRow, backupRemoteStatus, backupRemoteFailed)

			// The unidecode row follows the backup rows, the store
			// row's own precedent: it always prints exactly one line,
			// since the matcher can never score without a verified
			// table.
			unidecodeStatus, unidecodeFailed := unidecodeRowStatus(environ)
			addChecked(unidecodeDoctorRow, unidecodeStatus, unidecodeFailed)

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
			addChecked(clientIPDoctorRow, clientIPStatus, clientIPFailed)
			webStatus, webFailed := webRowStatus(environ, cfg)
			addChecked(webDoctorRow, webStatus, webFailed)
			ifaceAddrs, ifaceErr := net.InterfaceAddrs()
			bindStatus, bindFailed := bindAddrsRowStatus(environ, ifaceAddrs, ifaceErr)
			addChecked(bindAddrsDoctorRow, bindStatus, bindFailed)
			probeStatus, probeFailed := boundaryProbeRowStatus(cmd.Context(), environ, reg, cfg)
			addChecked(boundaryProbeDoctorRow, probeStatus, probeFailed)

			// The music volume row closes the fixed-count rows: serve's
			// startup check on staging and the library root, and where
			// each collection root the store lists sits.
			volumeStatus, volumeFailed := musicVolumeRowStatus(cfg, storeReport, storeChecked)
			addChecked(musicVolumeDoctorRow, volumeStatus, volumeFailed)

			// The settings rows print one per registered setting, in
			// declaration order, after the boundary rows, and one row
			// per orphan the store holds under a key no rung still
			// declares.
			settingsRows, settingsFailed := settingsDoctorRows(environ, reg, storeReport, storeChecked)
			rows = append(rows, settingsRows...)
			failed = failed || settingsFailed

			for _, row := range rows {
				if err := printRow(stdout, row.key, row.status); err != nil {
					return err
				}
			}
			if failed {
				return errors.New("configuration is incomplete")
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// unclaimedLoadErrors returns, in sorted order, the DUBPLATE_ keys
// statuses reports for which no row doctor prints is responsible: the
// owned wall rows, the thresholds, and the identity sources' selector
// and keys each read their status from statuses, so every other key reaches
// the operator only through the list this returns.
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
// The environment only seeds the first admin: the roster alone decides
// who is one, so the row warns when the address no longer names a live
// admin.
const adminEmailKey = "DUBPLATE_ADMIN_EMAIL"

// The two warnings the admin email row can carry. A removed member and
// an address no member holds share the first, since one action fixes
// both: member add with --admin reactivates a removed row, and the
// variable may instead name a live admin. A live non-admin has no verb
// to promote it, since member add refuses a live address, so the only
// action is to name a live admin.
const (
	adminEmailNotMember = "warning: this address is not a live member; run dubplate-server member add <address> --admin, or set " + adminEmailKey + " to a live admin"
	adminEmailNotAdmin  = "warning: this address is a live member without the admin flag; set " + adminEmailKey + " to a live admin"
)

// adminEmailRowStatus renders the admin email row from the wall's own
// status and the report the store row read. The wall's status stands
// unless it is "ok". An "ok" stays when the store cannot speak for the
// roster (not checked, not vouched for, absent, older than the users
// table) and when the roster holds no live member, which the members
// row already fails. Otherwise the roster standing decides: a live
// admin is "ok", and any other standing is a warning in the row's
// status text, which never fails doctor.
func adminEmailRowStatus(wall string, r store.Report, checked bool) string {
	switch {
	case wall != "ok", !checked, !storeVouches(r.State):
		return wall
	case r.Facts == nil, r.StoreVersion < store.UsersVersion, r.Facts.Members == 0:
		return wall
	}
	switch r.Standing {
	case store.StandingNone, store.StandingRemoved:
		return adminEmailNotMember
	case store.StandingMember:
		return adminEmailNotAdmin
	}
	return wall
}

// musicVolumeDoctorRow is the row reporting the music volume: staging,
// the library root, and the collection roots under it.
const musicVolumeDoctorRow = "music volume"

// musicVolumeRowStatus reports what serve's startup check finds of
// staging and the library root, and where each collection root the
// store lists sits, and whether the row fails. A directory not created
// yet is named without failing doctor, as on a fresh box before
// deploy.sh runs, and a collection root not created yet reads "not
// created", since rung 9 creates them. Any other refusal the startup
// check makes fails the row, as does a collection root outside the
// library root's mount. A nil cfg reads "not checked".
func musicVolumeRowStatus(cfg *config.Config, r store.Report, checked bool) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	stagingDir, library := staging.Read(cfg.WallValue(stagingKey)), staging.Read(cfg.LibraryRoot)
	var absent []string
	for _, d := range []struct {
		role string
		dir  staging.Dir
	}{{"staging directory", stagingDir}, {"library root", library}} {
		if errors.Is(d.dir.Err, fs.ErrNotExist) {
			absent = append(absent, d.role+" "+d.dir.Path+" does not exist")
		}
	}
	if len(absent) > 0 {
		return strings.Join(absent, "; "), false
	}
	if err := staging.Check(stagingDir, library); err != nil {
		return err.Error(), true
	}

	parts := []string{"ok: staging and the library root share one mount"}
	if !checked || r.Facts == nil {
		return parts[0], false
	}
	for _, root := range r.Facts.CollectionRoots {
		path := filepath.Join(cfg.LibraryRoot, string(root))
		d := staging.Read(path)
		switch {
		case errors.Is(d.Err, fs.ErrNotExist):
			parts = append(parts, "collection root "+path+" not created")
		case d.Err != nil:
			parts, failed = append(parts, "collection root: "+d.Err.Error()), true
		case d.StatxErr != nil:
			parts, failed = append(parts, "collection root: "+d.StatxErr.Error()), true
		case d.Dev != library.Dev || d.Mask&unix.STATX_MNT_ID == 0 || d.MountID != library.MountID:
			parts, failed = append(parts, "collection root "+path+" sits outside the library root's mount"), true
		default:
			parts = append(parts, "collection root "+path+" on the library root's mount")
		}
	}
	return strings.Join(parts, "; "), failed
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

// doctorRow is one doctor line: a printRow key and its status text.
type doctorRow struct {
	key, status string
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
	for _, key := range config.KeyNames(slices.Concat(src.Required, src.Optional)) {
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
	for _, key := range config.KeyNames(src.Required) {
		if v, ok := lookupEnviron(environ, key); ok {
			values[key] = v
		}
	}
	for _, key := range config.KeyNames(src.Optional) {
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
