package store

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
)

// leafDirName is the one directory under the state directory the store
// occupies.
const leafDirName = "store"

// leafDirPerm is the mode every directory in the leaf carries: the leaf
// itself, db/, snapshots/, staging/, and leases/.
const leafDirPerm = 0o700

// lockFilePerm is the mode of store.lock and of every lease file, each
// a sensitive coordination file inside a leafDirPerm directory.
const lockFilePerm = 0o600

// dbFileName is the database file's name inside db/.
const dbFileName = "dubplate.db"

// leaf names every path the store leaf holds.
type leaf struct {
	root              string
	lockPath          string
	dbDir             string
	dbPath            string
	snapshotsDir      string
	stagingDir        string
	leasesDir         string
	restoreIntentPath string
}

// newLeaf returns the leaf rooted at stateDir's "store" subdirectory.
// It builds paths only; it creates nothing.
func newLeaf(stateDir string) leaf {
	root := filepath.Join(stateDir, leafDirName)
	dbDir := filepath.Join(root, "db")
	return leaf{
		root:              root,
		lockPath:          filepath.Join(root, "store.lock"),
		dbDir:             dbDir,
		dbPath:            filepath.Join(dbDir, dbFileName),
		snapshotsDir:      filepath.Join(root, "snapshots"),
		stagingDir:        filepath.Join(root, "staging"),
		leasesDir:         filepath.Join(root, "leases"),
		restoreIntentPath: filepath.Join(root, "restore-intent"),
	}
}

// geteuid is the seam the package reads the effective uid through,
// the form fsprobe.go's statfsFunc already uses for syscall.Statfs. A
// test swaps it to exercise an owner mismatch without running as a
// different real user, and restores it with t.Cleanup; the store's
// no-t.Parallel rule makes that safe. Every production path leaves it
// at its default, os.Geteuid.
var geteuid = os.Geteuid

// create makes l's leaf, store.lock, and its four directories (each
// leafDirPerm), tolerating any of them already existing: a crash can
// stop an earlier create partway, and a second process can race it.
// It refuses, creating nothing, when geteuid() does not own stateDir,
// so a first run as root can never leave a root-owned leaf.
func (l leaf) create(stateDir string) error {
	info, err := os.Stat(stateDir)
	if err != nil {
		return fmt.Errorf("create leaf: stat state directory: %w", err)
	}
	if err := checkOwner(info); err != nil {
		return fmt.Errorf("create leaf: state directory %s: %w", stateDir, err)
	}

	for _, dir := range []string{l.root, l.dbDir, l.snapshotsDir, l.stagingDir, l.leasesDir} {
		if err := os.MkdirAll(dir, leafDirPerm); err != nil {
			return fmt.Errorf("create leaf: make %s: %w", dir, err)
		}
		// MkdirAll is a no-op on a directory that already exists, at
		// whatever mode it already carries, so an explicit chmod is the
		// only way an idempotent retry still guarantees leafDirPerm.
		if err := os.Chmod(dir, leafDirPerm); err != nil {
			return fmt.Errorf("create leaf: chmod %s: %w", dir, err)
		}
	}

	f, err := os.OpenFile(l.lockPath, os.O_CREATE|os.O_RDWR, lockFilePerm)
	if err != nil {
		return fmt.Errorf("create leaf: create %s: %w", l.lockPath, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("create leaf: close %s: %w", l.lockPath, err)
	}
	return nil
}

// checkOwner refuses when info's owning uid differs from geteuid(),
// naming the owner and the fix a caller running as the wrong user
// needs: the owner's account with the server's environment file loaded,
// which a bare su shell does not do.
func checkOwner(info fs.FileInfo) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("check owner: no unix stat info for %s", info.Name())
	}
	if int(st.Uid) == geteuid() {
		return nil
	}
	name := ownerName(st.Uid)
	return fmt.Errorf("owned by %s: run the command as %s with the server's environment file loaded", name, name)
}

// checkLeafShape refuses when info, the leaf's own stat, is not a
// leafDirPerm directory: a regular file at path, or a directory at any
// other mode. It takes info and path as parameters, rather than
// stat'ing path itself, so a caller that already stat'd the leaf (to
// decide it exists at all) does not stat it twice.
func checkLeafShape(info fs.FileInfo, path string) error {
	if !info.IsDir() {
		return fmt.Errorf("leaf %s: not a directory, want a directory at mode %04o", path, leafDirPerm)
	}
	if mode := info.Mode().Perm(); mode != leafDirPerm {
		return fmt.Errorf("leaf %s: mode %04o, want a directory at mode %04o", path, mode, leafDirPerm)
	}
	return nil
}

// ownerName resolves uid to a username, falling back to "uid:<n>" when
// the uid has no passwd entry, the shape [user.LookupId] documents for
// its UnknownUserIdError.
func ownerName(uid uint32) string {
	u, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return fmt.Sprintf("uid:%d", uid)
	}
	return u.Username
}
