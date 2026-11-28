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
// per record kind.
const leafDirName = "auth"

// nonceSubdir and epochSubdir are the two record subdirectories Open
// creates beneath the leaf. They double as [recordID]'s
// domain-separation labels.
const (
	nonceSubdir = "nonce"
	epochSubdir = "epoch"
)

// dirPerm and filePerm are this leaf's directory and record-file
// permissions: identity state is private, so the leaf states its own
// permission in code rather than leaving it to the unit's umask.
const (
	dirPerm  = 0o700
	filePerm = 0o600
)

func dirPath(stateDir string) string {
	return filepath.Join(stateDir, leafDirName)
}

// A Dir is the auth leaf, the directory under the state directory that
// holds the auth package's durable records.
// The nonce store and the epoch store are built on the record path Dir
// derives, so neither can disagree about where a record lives or what
// its name may contain.
type Dir struct {
	root string
}

// Open creates the leaf and both record subdirectories under stateDir
// if they do not already exist, and returns a Dir rooted there.
//
// It refuses to start when the leaf itself already has a mode wider
// than dirPerm: os.MkdirAll is a no-op on an existing directory
// whatever its mode, so a leaf created once at a wider mode, by a
// hand-run command or an earlier deploy, would otherwise stay wide
// forever. Record names are digests, but a listable directory is still
// a listing of how many links are outstanding.
func Open(stateDir string) (*Dir, error) {
	s := &Dir{root: dirPath(stateDir)}
	for _, sub := range []string{nonceSubdir, epochSubdir} {
		if err := os.MkdirAll(s.subdir(sub), dirPerm); err != nil {
			return nil, fmt.Errorf("open auth dir: create %s: %w", sub, err)
		}
	}
	info, err := os.Stat(s.root)
	if err != nil {
		return nil, fmt.Errorf("open auth dir: stat leaf: %w", err)
	}
	if mode := info.Mode().Perm(); mode&^dirPerm != 0 {
		return nil, fmt.Errorf("open auth dir: leaf %s has mode %o, want at most %o", s.root, mode, dirPerm)
	}
	return s, nil
}

func (s *Dir) subdir(sub string) string {
	return filepath.Join(s.root, sub)
}

// noncePath returns the record path for id inside the nonce
// subdirectory, after refusing an id validID rejects.
func (s *Dir) noncePath(id string) (string, error) {
	return s.recordPath(nonceSubdir, id)
}

// epochPath returns the record path for id inside the epoch
// subdirectory, after refusing an id validID rejects.
func (s *Dir) epochPath(id string) (string, error) {
	return s.recordPath(epochSubdir, id)
}

func (s *Dir) recordPath(sub, id string) (string, error) {
	if !validID(id) {
		return "", fmt.Errorf("record path: invalid id: length %d, %s", len(id), idRefusalReason(id))
	}
	return filepath.Join(s.root, sub, id), nil
}

// recordID derives a record's filename from kind and value: the hex of
// the SHA-256 digest of the two together, untruncated. kind
// domain-separates the digest the same way macInput length-prefixes a
// MAC's purpose label, so a nonce record and an epoch record can never
// share a name. Callers pass [nonceSubdir] or [epochSubdir] as kind.
func recordID(kind, value string) string {
	buf := appendField(nil, []byte(kind))
	buf = appendField(buf, []byte(value))
	sum := sha256.Sum256(buf)
	return hex.EncodeToString(sum[:])
}

// validID reports whether id is safe to use as a single path segment
// under root. Ids are digest-derived ([recordID]), so anything else is a
// bug rather than input, and this check makes it a loud one rather than
// a path traversal.
func validID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	return !strings.ContainsAny(id, `/\`)
}

// idRefusalReason names why validID refused id without printing id
// itself, since a refused id is not known to be a digest and this
// package's errors never carry a credential.
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
