package auth

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/store"
	"golang.org/x/sys/unix"
)

// adminEmailKey is the owned wall row naming the operator's address.
// The environment only seeds the first admin: the roster alone decides
// who is one, so the admin email row warns when the address no longer
// names a live admin.
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

// AdminAddressRowStatus renders the admin email row from the wall's own
// status and the report the store row read. The wall's status stands
// unless it is "ok". An "ok" stays when the store cannot speak for the
// roster (not checked, not vouched for, absent, older than the users
// table) and when the roster holds no live member, which the members
// row already fails. Otherwise the roster standing decides: a live
// admin is "ok", and any other standing is a warning in the row's
// status text, which never fails doctor.
func AdminAddressRowStatus(wall string, r store.Report, checked bool) string {
	switch {
	case wall != "ok", !checked, !r.State.Vouches():
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

// MembersRowStatus renders the members row from the report the store
// row read, and reports whether the row fails. It decides on the
// report's StoreVersion, never on whether Facts is present, since a
// version 1 store carries no facts while a version 2 store carries
// facts without the users table. A store the store row cannot vouch
// for reads "not read"; the row fails only for a store that holds the
// users table and no live admin, since nobody could then add a member
// through the page.
func MembersRowStatus(r store.Report, checked bool) (status string, failed bool) {
	switch {
	case !checked:
		return "not read", false
	case r.State == store.StateAbsent:
		return "no store yet; start serve, then member add --admin", false
	case !r.State.Vouches():
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

// DirRowOptions are the process facts [DirRowStatus] reads.
type DirRowOptions struct {
	// Geteuid returns the effective uid the leaf must belong to.
	Geteuid func() int
}

// DirRowStatus reports, without writing, whether serve could open
// the auth leaf under cfg's state directory: "ok" when the leaf is at
// most 0700, owned by the effective uid, and each record directory is
// writable; "absent; serve creates it" when the leaf is missing and
// its nearest existing ancestor is writable. Otherwise it names every
// problem and fails: the leaf's mode in the words [Open] refuses it
// with, a foreign owner, or a directory the process cannot write, the
// shape a path outside the unit's ReadWritePaths takes. A nil cfg reads
// "not checked" and adds no failure of its own, since Load's own
// report already fails doctor.
func DirRowStatus(cfg *config.Config, opts DirRowOptions) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	leaf := DirPath(cfg.StateDir)
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
	if mode := info.Mode().Perm(); mode&^dirPerm != 0 {
		problems = append(problems, fmt.Sprintf("leaf %s has mode %o, want at most %o", leaf, mode, dirPerm))
	}
	// io/fs.FileInfo documents "Sys() any // underlying data source
	// (can return nil)"; on Linux, os.Stat's carries a *syscall.Stat_t.
	if st, ok := info.Sys().(*syscall.Stat_t); !ok {
		problems = append(problems, fmt.Sprintf("leaf %s: cannot determine its owner", leaf))
	} else if euid := opts.Geteuid(); int(st.Uid) != euid {
		problems = append(problems, fmt.Sprintf("leaf %s is owned by uid %d, want uid %d", leaf, st.Uid, euid))
	}
	for _, name := range []string{nonceSubdir, epochSubdir} {
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

// errDirNotWritable reports that this process may not list, create,
// and remove entries in a directory a later write needs.
var errDirNotWritable = errors.New("not writable")

// checkDirWritable asks the kernel whether this process's effective
// uid may read, write, and search dir: faccessat(2) with AT_EACCESS,
// whose man page says it "Perform[s] access checks using the effective
// user and group IDs" and fails with EROFS when "Write permission was
// requested for a file on a read-only filesystem", the shape a
// directory outside the unit's ReadWritePaths takes under
// ProtectSystem=strict. It writes nothing, so it can neither race the
// roster's directory flock nor leave a probe file behind. Read and
// search are asked for beside write because every writer here opens or
// lists the directory too.
func checkDirWritable(dir string) error {
	if err := unix.Faccessat(unix.AT_FDCWD, dir, unix.R_OK|unix.W_OK|unix.X_OK, unix.AT_EACCESS); err != nil {
		return fmt.Errorf("%s is %w: %w", dir, errDirNotWritable, err)
	}
	return nil
}
