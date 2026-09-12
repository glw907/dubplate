package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// leafDirName is the one directory under the state directory the auth
// package writes. The nonce store and the epoch store share it, so the
// serve unit's ReadWritePaths grant is one static path rather than one
// per record kind (R-5b11).
const leafDirName = "auth"

// nonceSubdir and epochSubdir are the two record subdirectories Open
// creates beneath the leaf (R-5b11).
const (
	nonceSubdir = "nonce"
	epochSubdir = "epoch"
)

// dirPerm and filePerm are this leaf's directory and record-file
// permissions (R10): identity state is not the shadow store's evidence
// tier, so this rung states its own permission in code rather than
// inheriting rung 4's 0755/0644 or leaving it to the unit's umask.
const (
	dirPerm  = 0o700
	filePerm = 0o600
)

// StoreDir joins stateDir to the auth leaf, so the leaf name has one
// home rather than a literal at every call site.
func StoreDir(stateDir string) string {
	return filepath.Join(stateDir, leafDirName)
}

// A Store is the auth leaf's durable state, rooted at
// StoreDir(stateDir). The nonce store and the epoch store are built on
// the record path Store derives, so neither can disagree about where a
// record lives or what its name may contain.
type Store struct {
	root string
}

// Open creates the leaf and both record subdirectories under stateDir
// if they do not already exist, and returns a Store rooted there.
func Open(stateDir string) (*Store, error) {
	s := &Store{root: StoreDir(stateDir)}
	for _, sub := range []string{nonceSubdir, epochSubdir} {
		if err := os.MkdirAll(s.subdir(sub), dirPerm); err != nil {
			return nil, fmt.Errorf("open auth store: create %s: %w", sub, err)
		}
	}
	return s, nil
}

// subdir returns the path of one of the store's record subdirectories.
// It is the one place the leaf and a subdirectory become a directory
// path, so no caller reaches for the root field to build one.
func (s *Store) subdir(sub string) string {
	return filepath.Join(s.root, sub)
}

// NoncePath and EpochPath return the record path for id inside the
// nonce and epoch subdirectories, after refusing an id validID rejects.
// This is the one place a subdirectory and an id become a path, so no
// other file in the package builds one by hand.
func (s *Store) NoncePath(id string) (string, error) {
	return s.recordPath(nonceSubdir, id)
}

func (s *Store) EpochPath(id string) (string, error) {
	return s.recordPath(epochSubdir, id)
}

func (s *Store) recordPath(sub, id string) (string, error) {
	if !validID(id) {
		return "", fmt.Errorf("record path: invalid id %q", id)
	}
	return filepath.Join(s.root, sub, id), nil
}

// Write writes data to path atomically at filePerm, through
// atomicfile.WriteFile. It is the one write path a record whose
// overwrite discipline atomicfile.WriteFile's rename-over-target covers
// can use; a record needing first-writer-wins semantics, such as the
// nonce store's, opens through NoncePath and O_CREATE|O_EXCL directly
// instead, because a rename would silently overwrite it.
func (s *Store) Write(path string, data []byte) error {
	if err := atomicfile.WriteFile(path, data, filePerm); err != nil {
		return fmt.Errorf("write record: %w", err)
	}
	return nil
}

// RecordID derives a record's filename from value: the hex of value's
// full SHA-256 digest, untruncated (R-5b12). One derivation serves both
// the nonce store and the epoch store, so rung 6's re-key onto a
// surrogate id is a change to this one function rather than a rename of
// every record (R34).
func RecordID(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

// validID reports whether id is safe to use as a single path segment
// under root. Ids are digest-derived (RecordID), so anything else is a
// bug rather than input, and this check makes it a loud one rather than
// a path traversal (shape adopted from
// internal/engine/shadow/store.go:66-70).
func validID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	return !strings.ContainsAny(id, `/\`)
}
