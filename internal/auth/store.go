package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// leafDirName is the one directory under the state directory the auth
// package writes. The nonce store and the epoch store share it, so the
// serve unit's ReadWritePaths grant is one static path rather than one
// per record kind (R-5b11).
const leafDirName = "auth"

// nonceSubdir and epochSubdir are the two record subdirectories Open
// creates beneath the leaf (R-5b11). They double as [RecordID]'s
// domain-separation labels, so the one pair of names serves both
// purposes rather than a second constant apiece.
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
//
// It refuses to start when the leaf itself already has a mode wider
// than dirPerm: os.MkdirAll is a no-op on an existing directory
// whatever its mode, so a leaf created once at a wider mode, by a
// hand-run command or an earlier deploy, would otherwise stay wide
// forever. Record names are digests, but a listable directory is still
// a listing of how many links are outstanding (review m3).
func Open(stateDir string) (*Store, error) {
	s := &Store{root: StoreDir(stateDir)}
	for _, sub := range []string{nonceSubdir, epochSubdir} {
		if err := os.MkdirAll(s.subdir(sub), dirPerm); err != nil {
			return nil, fmt.Errorf("open auth store: create %s: %w", sub, err)
		}
	}
	info, err := os.Stat(s.root)
	if err != nil {
		return nil, fmt.Errorf("open auth store: stat leaf: %w", err)
	}
	if mode := info.Mode().Perm(); mode&^dirPerm != 0 {
		return nil, fmt.Errorf("open auth store: leaf %s has mode %o, want at most %o", s.root, mode, dirPerm)
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
		return "", fmt.Errorf("record path: invalid id: length %d, %s", len(id), idRefusalReason(id))
	}
	return filepath.Join(s.root, sub, id), nil
}

// RecordID derives a record's filename from kind and value: the hex of
// the SHA-256 digest of the two together, untruncated (R-5b12). kind
// domain-separates the digest the same way macInput already
// length-prefixes a MAC's own purpose label, so the nonce store and
// the epoch store keying two different values can never collide even
// if a future caller placed both kinds of record in one directory
// (review m12). Callers pass [nonceSubdir] or [epochSubdir] as kind.
func RecordID(kind, value string) string {
	buf := appendField(nil, []byte(kind))
	buf = appendField(buf, []byte(value))
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])
}

// validID reports whether id is safe to use as a single path segment
// under root. Ids are digest-derived (RecordID), so anything else is a
// bug rather than input, and this check makes it a loud one rather than
// a path traversal.
func validID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	return !strings.ContainsAny(id, `/\`)
}

// idRefusalReason names why validID refused id, without ever printing
// id itself: every current caller passes a RecordID digest, but a
// future caller that skipped it could pass a raw credential, and this
// package's doc promises no error here ever carries one (review m7,
// internal/auth/doc.go).
func idRefusalReason(id string) string {
	switch {
	case id == "":
		return "empty"
	case id == ".":
		return "the current-directory name"
	case id == "..":
		return "the parent-directory name"
	case strings.ContainsAny(id, `/\`):
		return "contains a path separator"
	default:
		return "not a valid single path segment"
	}
}
