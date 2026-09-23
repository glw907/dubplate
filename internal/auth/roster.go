package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/identity"
)

// rosterSchemaVersion is the roster file's own schema version
// (R-5b19): a version field costs one line today and gives rung 6's
// migration something to read, rather than a bare array that leaves
// the first schema change guessing what an unversioned file means.
const rosterSchemaVersion = 1

// maxDisplayNameLength bounds a display name in bytes: 200 bytes
// comfortably covers a legal name in any script, with room for a
// honorific or a nickname, well short of turning the field into free
// text.
const maxDisplayNameLength = 200

// maxStreamerUsernameLength bounds streamer_username in bytes: 64
// bytes clears every common platform's own username cap (GitHub 39,
// Discord 32) with headroom, while still refusing an attempt to stuff
// arbitrary text into what R3 requires as a stable principal.
const maxStreamerUsernameLength = 64

// Roster load and lookup sentinels. A refusal names the check that
// failed and the file's path, never a record's own address or
// principal (doc.go): the one exception is the audit record Task 5
// writes.
var (
	// ErrRosterMissing reports that OpenRoster's path does not exist.
	// The read-open fails closed here (R2) rather than answering an
	// empty roster: a fail-open roster lets a deployment start and
	// refuse every login as unknown, which a member reads as a broken
	// login rather than a broken configuration this error names as
	// such. This supersedes the salvage loader's fail-open branch
	// (_salvage/internal/config/contributors.go:36-40).
	ErrRosterMissing = errors.New("auth: roster file missing")

	// ErrRosterUnreadable reports that path exists but could not be
	// read, for example a permission error.
	ErrRosterUnreadable = errors.New("auth: roster file unreadable")

	// ErrRosterUnparseable reports that path's content is not valid
	// JSON in the roster's schema.
	ErrRosterUnparseable = errors.New("auth: roster file unparseable")

	// ErrRosterRecordInvalid reports that a parsed roster file failed
	// validation: a duplicate address, a duplicate streamer_username, a
	// control character, an out-of-bounds field, or an address
	// net/mail cannot parse.
	ErrRosterRecordInvalid = errors.New("auth: roster record invalid")

	// ErrRosterExists reports that CreateRoster's target already has a
	// file. The write-open refuses an existing file (R37) so a second
	// roster can never be created over a first.
	ErrRosterExists = errors.New("auth: roster file already exists")

	// ErrMemberNotFound reports that Lookup's address matched no member
	// once normalized.
	ErrMemberNotFound = errors.New("auth: member not found")
)

// Member is one roster entry (R1): a login address, a display name, an
// optional stable principal, and the admin flag. The admin flag is the
// whole of rung 5's authorization vocabulary (spec:1111-1115): a later
// rung reaching for a role model is adding one, not extending one.
type Member struct {
	// Address is the member's login address, normalized by
	// normalizeMemberAddress.
	Address string
	// DisplayName is the member's human-readable name.
	DisplayName string
	// StreamerUsername is the member's stable principal: ASCII-printable
	// and unique across the roster when set (R3). Empty means unset.
	StreamerUsername string
	// Admin marks the member as an operator.
	Admin bool
}

// rosterFile is the roster file's on-disk shape (R-5b19).
type rosterFile struct {
	SchemaVersion int            `json:"schema_version"`
	Members       []memberRecord `json:"members"`
}

// memberRecord is one roster file entry as read from disk, before
// validation promotes it to a Member.
type memberRecord struct {
	Address          string `json:"address"`
	DisplayName      string `json:"display_name"`
	StreamerUsername string `json:"streamer_username,omitempty"`
	Admin            bool   `json:"admin"`
}

// Roster is the rung 5 member store (R1): one type opens the file,
// validates it, answers a lookup, and lists the roster. No caller
// outside this file knows the format is JSON, which is what makes
// rung 6's substitution a change of one constructor.
type Roster struct {
	path      string
	members   []Member
	byAddress map[string]int
}

// OpenRoster is the read-open (R1, R2): it reads and validates the
// roster file at path, and never creates it, so a fresh install with
// no roster fails closed rather than answering an empty one. A
// missing, unreadable, or unparseable file, and a record failing
// validation, are each a distinct, path-naming error.
func OpenRoster(path string) (*Roster, error) {
	data, err := os.ReadFile(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("%w: %s", ErrRosterMissing, path)
	case err != nil:
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnreadable, path, err)
	}
	return parseRoster(path, data)
}

// CreateRoster is the write-open (R37, R-5b20): it creates path at
// 0600 carrying an empty, schema-versioned roster when path is absent,
// and refuses when path already exists, so dubplate-server member add
// --admin can bootstrap a fresh install without a second constructor
// ever overwriting a first roster. It creates no directory:
// deploy.sh's pre-create and the serve unit's ReadWritePaths grant own
// that.
func CreateRoster(path string) (*Roster, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrRosterExists, path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnreadable, path, err)
	}

	// Members is []memberRecord{} rather than nil so the written file
	// carries "members": [] instead of "members": null.
	empty := rosterFile{SchemaVersion: rosterSchemaVersion, Members: []memberRecord{}}
	data, err := json.Marshal(empty)
	if err != nil {
		return nil, fmt.Errorf("auth: marshal empty roster: %w", err)
	}
	if err := atomicfile.WriteFile(path, data, 0o600); err != nil {
		return nil, fmt.Errorf("auth: create roster %s: %w", path, err)
	}
	return OpenRoster(path)
}

// parseRoster unmarshals and validates data as loaded from path.
func parseRoster(path string, data []byte) (*Roster, error) {
	var file rosterFile
	if err := json.Unmarshal(data, &file); err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnparseable, path, err)
	}
	members, byAddress, err := validateRoster(file.Members)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrRosterRecordInvalid, path, err)
	}
	return &Roster{path: path, members: members, byAddress: byAddress}, nil
}

// normalizeMemberAddress is the store's single resolution point (R1,
// S20:367-373): the only place a login address becomes the identity
// value the roster keys, compares, and looks members up by. Rung 6
// substitutes a durable surrogate id here without touching a caller
// (2.0 hook 5).
func normalizeMemberAddress(address string) (string, error) {
	return identity.NormalizeAddress(address)
}

// validateRoster promotes records to Members, applying B9's list: each
// address is parsed and normalized through normalizeMemberAddress and
// unique after normalization, no field carries a control character,
// every field is bounded, and streamer_username, when present, is
// ASCII-printable and unique across the roster. A record's own address
// or principal never appears in a returned error, only its position.
func validateRoster(records []memberRecord) ([]Member, map[string]int, error) {
	members := make([]Member, 0, len(records))
	byAddress := make(map[string]int, len(records))
	seenStreamer := make(map[string]int, len(records))
	var errs []error

	for i, rec := range records {
		normalized, err := normalizeMemberAddress(rec.Address)
		if err != nil {
			errs = append(errs, fmt.Errorf("member %d: address does not parse as a login address", i))
			continue
		}
		if _, dup := byAddress[normalized]; dup {
			errs = append(errs, fmt.Errorf("member %d: duplicate address", i))
			continue
		}
		if containsControl(rec.DisplayName) {
			errs = append(errs, fmt.Errorf("member %d: display name carries a control character", i))
			continue
		}
		if len(rec.DisplayName) > maxDisplayNameLength {
			errs = append(errs, fmt.Errorf("member %d: display name exceeds %d bytes", i, maxDisplayNameLength))
			continue
		}
		if rec.StreamerUsername != "" {
			if !isASCIIPrintable(rec.StreamerUsername) {
				errs = append(errs, fmt.Errorf("member %d: streamer_username is not ASCII-printable", i))
				continue
			}
			if len(rec.StreamerUsername) > maxStreamerUsernameLength {
				errs = append(errs, fmt.Errorf("member %d: streamer_username exceeds %d bytes", i, maxStreamerUsernameLength))
				continue
			}
			if _, dup := seenStreamer[rec.StreamerUsername]; dup {
				errs = append(errs, fmt.Errorf("member %d: duplicate streamer_username", i))
				continue
			}
			seenStreamer[rec.StreamerUsername] = i
		}

		byAddress[normalized] = len(members)
		members = append(members, Member{
			Address:          normalized,
			DisplayName:      rec.DisplayName,
			StreamerUsername: rec.StreamerUsername,
			Admin:            rec.Admin,
		})
	}
	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}
	return members, byAddress, nil
}

// containsControl reports whether s carries an ASCII control character
// or DEL.
func containsControl(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f })
}

// isASCIIPrintable reports whether every byte of s is an ASCII
// printable character (0x20 through 0x7e), which also refuses any
// multi-byte UTF-8 sequence, since every continuation byte falls
// outside that range.
func isASCIIPrintable(s string) bool {
	for i := 0; i < len(s); i++ {
		if b := s[i]; b < 0x20 || b > 0x7e {
			return false
		}
	}
	return true
}

// Lookup resolves address to its roster member, matching
// case-insensitively on the normalized address, with normalization
// happening before any keying, logging, or rate-limit lookup
// (spec:1161-1163).
func (r *Roster) Lookup(address string) (Member, error) {
	normalized, err := normalizeMemberAddress(address)
	if err != nil {
		return Member{}, fmt.Errorf("%w: %v", ErrMemberNotFound, err)
	}
	idx, ok := r.byAddress[normalized]
	if !ok {
		return Member{}, ErrMemberNotFound
	}
	return r.members[idx], nil
}

// List returns a copy of every member on the roster, in file order.
func (r *Roster) List() []Member {
	out := make([]Member, len(r.members))
	copy(out, r.members)
	return out
}
