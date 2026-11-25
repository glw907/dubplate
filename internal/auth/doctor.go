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
	"github.com/glw907/dubplate/internal/doctor"
	"golang.org/x/sys/unix"
)

// dirRowName is the name of the row reporting the auth leaf serve
// opens.
const dirRowName = "auth dir"

// DoctorRows renders the auth dir row for cfg, which is nil when
// [config.Load] refused the environment.
func DoctorRows(cfg *config.Config) []doctor.Row {
	row := doctor.Row{Name: dirRowName}
	row.Status, row.Failed = dirStatus(cfg, dirOptions{geteuid: os.Geteuid})
	return []doctor.Row{row}
}

// dirOptions are the process facts [dirStatus] reads.
type dirOptions struct {
	// geteuid returns the effective uid the leaf must belong to.
	geteuid func() int
}

// dirStatus reports, without writing, whether serve could open
// the auth leaf under cfg's state directory: "ok" when the leaf is at
// most 0700, owned by the effective uid, and each record directory is
// writable; "absent; serve creates it" when the leaf is missing and
// its nearest existing ancestor is writable. Otherwise it names every
// problem and fails: the leaf's mode in the words [Open] refuses it
// with, a foreign owner, or a directory the process cannot write, the
// shape a path outside the unit's ReadWritePaths takes. A nil cfg reads
// "not checked" and adds no failure of its own, since Load's own
// report already fails doctor.
func dirStatus(cfg *config.Config, opts dirOptions) (status string, failed bool) {
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
	} else if euid := opts.geteuid(); int(st.Uid) != euid {
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
