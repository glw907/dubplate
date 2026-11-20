package store

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Filesystem magic numbers (linux/magic.h, read 2026-09-23) that make
// [probeFilesystem] refuse: every one spreads writes across more than
// one host, which WAL requires all stay on (sqlite.org/wal.html
// section 1).
const (
	nfsSuperMagic  = 0x6969
	cifsSuperMagic = 0xFF534D42
	smb2SuperMagic = 0xFE534D42
	v9fsMagic      = 0x01021997
	smbSuperMagic  = 0x517B // the retired smbfs driver
	cephSuperMagic = 0x00c36400
	afsSuperMagic  = 0x5346414F // the first of magic.h's two AFS constants
	afsFsMagic     = 0x6B414653 // the second
)

// fuseSuperMagic is the one magic number [probeFilesystem] warns on
// rather than refuses: a FUSE backend may or may not span hosts, and
// the workstation's own t.TempDir() and the box's local disk both
// avoid it, so the warning is conservative rather than a refusal.
const fuseSuperMagic = 0x65735546

// fsVerdict is probeFilesystem's classification of one filesystem.
type fsVerdict int

const (
	fsOpen fsVerdict = iota
	fsWarnFUSE
	fsRefuseMultiHost
)

// statfsFunc is the seam a test replaces to exercise every
// classification without mounting a real network filesystem. Every
// production caller uses the package default, syscall.Statfs.
var statfsFunc = syscall.Statfs

// probeFilesystem classifies the filesystem holding path by its
// f_type, following a symlink if path is one. It returns the raw magic
// number alongside the verdict so a refusal can name the type.
func probeFilesystem(path string) (fsVerdict, int64, error) {
	var st syscall.Statfs_t
	if err := statfsFunc(path, &st); err != nil {
		return fsRefuseMultiHost, 0, fmt.Errorf("statfs %s: %w", path, err)
	}
	switch st.Type {
	case nfsSuperMagic, cifsSuperMagic, smb2SuperMagic, v9fsMagic,
		smbSuperMagic, cephSuperMagic, afsSuperMagic, afsFsMagic:
		return fsRefuseMultiHost, st.Type, nil
	case fuseSuperMagic:
		return fsWarnFUSE, st.Type, nil
	default:
		return fsOpen, st.Type, nil
	}
}

// probeTarget picks db/ when it exists, else the leaf, else the leaf's
// parent, since the database's own directory is what WAL depends on
// and statfs follows a symlinked db/ to its target.
func (l leaf) probeTarget() string {
	if _, err := os.Stat(l.dbDir); err == nil {
		return l.dbDir
	}
	if _, err := os.Stat(l.root); err == nil {
		return l.root
	}
	return filepath.Dir(l.root)
}

// preCreate makes an empty file at path at mode 0600, whatever the
// process umask, and refuses without touching an existing file. It is
// the only way a database or snapshot file comes to exist: no DSN can
// create one.
func preCreate(path string) error {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("pre-create %s: %w", path, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("pre-create %s: close: %w", path, err)
	}
	return nil
}
