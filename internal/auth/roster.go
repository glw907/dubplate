package auth

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

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

	// ErrRosterAdminChannelForbidden reports that AddMember refused to
	// set the admin flag. Every identity.Channel is a network assertion,
	// a service token included, and the spec makes the command-line
	// verb the only path to an admin, so AddMember refuses the flag on
	// all of them. The admin grant arrives together with that verb.
	ErrRosterAdminChannelForbidden = errors.New("auth: admin flag refused on this channel")
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
//
// The file may change under a Roster: another process writes it, or an
// operator edits it by hand. Every Lookup and List checks the file and
// re-reads it when it has changed, and every write merges into the
// copy on disk, never the copy in memory.
type Roster struct {
	path string

	// mu guards snap. A writer holds it only to swap in the copy it
	// has just written, never while it waits for another writer, so a
	// Lookup never queues behind a lock another process holds.
	mu   sync.Mutex
	snap *rosterSnapshot
}

// rosterSnapshot is one read of the roster file: the members it
// carried, the stat of the descriptor they were read through, and when
// the read began.
type rosterSnapshot struct {
	members   []Member
	byAddress map[string]int
	info      os.FileInfo
	readAt    time.Time
}

// modTimeTrustWindow is how far a snapshot's read must trail the file's
// modification time before an unchanged stat is trusted to mean
// unchanged content. An in-place edit that keeps the inode and the size
// inside one timestamp tick keeps the modification time as well, so a
// read taken within one tick of that time cannot tell the two contents
// apart. One second covers the coarsest common filesystem timestamp.
const modTimeTrustWindow = time.Second

// rosterLockPollInterval is how often a writer retries the directory
// lock while another writer holds it. flock(2) offers no wait a context
// can cancel, so the writer polls a non-blocking attempt instead.
const rosterLockPollInterval = 25 * time.Millisecond

// errRosterBusy reports that a roster write gave up waiting for the
// directory lock because its context ended first.
var errRosterBusy = errors.New("auth: roster busy")

// OpenRoster is the read-open (R1, R2): it reads and validates the
// roster file at path, and never creates it, so a fresh install with
// no roster fails closed rather than answering an empty one. A
// missing, unreadable, or unparseable file, and a record failing
// validation, are each a distinct, path-naming error.
func OpenRoster(path string) (*Roster, error) {
	snap, err := readRoster(path)
	if err != nil {
		return nil, err
	}
	return &Roster{path: path, snap: snap}, nil
}

// CreateRoster is the write-open: it creates path at 0600 carrying an
// empty, schema-versioned roster when path is absent, and refuses with
// ErrRosterExists when path already exists, so a second bootstrap can
// never overwrite a first roster. The existence check and the write run
// under the same directory lock every roster write takes, so of two
// concurrent calls exactly one creates the file. It creates no
// directory: the deployment owns that.
func CreateRoster(ctx context.Context, path string) (*Roster, error) {
	unlock, err := lockRosterDir(ctx, filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	defer unlock()

	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("%w: %s", ErrRosterExists, path)
	} else if !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnreadable, path, err)
	}
	if err := writeRoster(path, nil); err != nil {
		return nil, err
	}
	return OpenRoster(path)
}

// readRoster opens path, stats the open descriptor, and reads the
// content from that same descriptor, so the stat a snapshot caches
// always describes the content it caches. Reading by path and then
// statting the path could pair the old content with a file renamed in
// between, and every later check would then match the new file while
// the cache served the old one.
func readRoster(path string) (*rosterSnapshot, error) {
	// The clock is read before the file is, so any edit the read
	// misses happened after readAt and cannot share a modification time
	// with the content read once readAt trails it by the trust window.
	readAt := time.Now()
	f, err := os.Open(path)
	switch {
	case errors.Is(err, os.ErrNotExist):
		return nil, fmt.Errorf("%w: %s", ErrRosterMissing, path)
	case err != nil:
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnreadable, path, err)
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("%w: %s: %v", ErrRosterUnreadable, path, err)
	}
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("%w: %s: read: %v", ErrRosterUnreadable, path, err)
	}
	members, byAddress, err := parseRoster(path, data)
	if err != nil {
		return nil, err
	}
	return &rosterSnapshot{members: members, byAddress: byAddress, info: info, readAt: readAt}, nil
}

// parseRoster decodes and validates data as loaded from path. The
// decoding is strict: member names match case-sensitively, an unknown
// or duplicated name is refused, and so is trailing data or invalid
// UTF-8. A file that decodes cleanly must still carry the current
// schema version and a members array, since null, {} and a missing or
// null members all decode without error.
func parseRoster(path string, data []byte) ([]Member, map[string]int, error) {
	var file rosterFile
	if err := json.Unmarshal(data, &file, json.RejectUnknownMembers(true)); err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %v", ErrRosterUnparseable, path, err)
	}
	if file.SchemaVersion != rosterSchemaVersion {
		return nil, nil, fmt.Errorf("%w: %s: schema_version is %d, want %d", ErrRosterUnparseable, path, file.SchemaVersion, rosterSchemaVersion)
	}
	if file.Members == nil {
		return nil, nil, fmt.Errorf("%w: %s: members is absent or null", ErrRosterUnparseable, path)
	}
	members, byAddress, err := validateRoster(file.Members)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: %s: %w", ErrRosterRecordInvalid, path, err)
	}
	return members, byAddress, nil
}

// writeRoster writes members to path through atomicfile, so a reader
// sees the old file or the new one and never a partial one.
func writeRoster(path string, members []Member) error {
	// Members is never nil, so the file carries "members": [] rather
	// than null, which the loader refuses.
	records := make([]memberRecord, 0, len(members))
	for _, m := range members {
		records = append(records, memberRecord(m))
	}
	file := rosterFile{SchemaVersion: rosterSchemaVersion, Members: records}
	if err := atomicfile.WriteStream(path, 0o600, func(w io.Writer) error {
		return json.MarshalWrite(w, file)
	}); err != nil {
		return fmt.Errorf("auth: write roster %s: %w", path, err)
	}
	return nil
}

// lockRosterDir takes an exclusive flock(2) on dir, the directory that
// holds the roster file, retrying until ctx ends, and returns the
// function that releases it. The lock is on the directory because every
// write replaces the file by rename: a lock on the file would sit on
// the inode the rename retires, and a writer opening the path afterward
// would lock the new inode unopposed. flock treats two descriptors for
// one file independently, so two writers in one process exclude each
// other exactly as two processes do.
func lockRosterDir(ctx context.Context, dir string) (func(), error) {
	d, err := os.Open(dir)
	if err != nil {
		return nil, fmt.Errorf("auth: open roster directory %s: %v", dir, err)
	}
	unlock := func() { _ = d.Close() }
	for {
		err := syscall.Flock(int(d.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return unlock, nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) && !errors.Is(err, syscall.EINTR) {
			unlock()
			return nil, fmt.Errorf("auth: lock roster directory %s: %v", dir, err)
		}
		select {
		case <-ctx.Done():
			unlock()
			return nil, fmt.Errorf("%w: %s: %v", errRosterBusy, dir, context.Cause(ctx))
		case <-time.After(rosterLockPollInterval):
		}
	}
}

// update is the roster's one read-modify-write, and every write goes
// through it. It takes the directory lock, re-reads and validates the
// file from disk, hands apply that disk copy, writes the members apply
// returns, re-reads the written file, and swaps that read in as the
// cache before the lock is released. A refusal from apply writes
// nothing.
func (r *Roster) update(ctx context.Context, apply func(disk *rosterSnapshot) ([]Member, error)) error {
	unlock, err := lockRosterDir(ctx, filepath.Dir(r.path))
	if err != nil {
		return err
	}
	defer unlock()

	disk, err := readRoster(r.path)
	if err != nil {
		return err
	}
	next, err := apply(disk)
	if err != nil {
		return err
	}
	if err := writeRoster(r.path, next); err != nil {
		return err
	}
	written, err := readRoster(r.path)
	if err != nil {
		return err
	}

	r.mu.Lock()
	r.snap = written
	r.mu.Unlock()
	return nil
}

// current returns the roster as the file holds it now. It serves the
// cached snapshot only when the file's stat still matches it and the
// snapshot's read trails the file's modification time by the trust
// window; otherwise it re-reads. A failed re-read drops the cache and
// returns the failure rather than the last good copy, which may carry
// the very member a hand edit just removed. The caller holds r.mu.
func (r *Roster) current() (*rosterSnapshot, error) {
	if r.snap != nil && r.snap.matches(r.path) {
		return r.snap, nil
	}
	snap, err := readRoster(r.path)
	if err != nil {
		r.snap = nil
		return nil, err
	}
	r.snap = snap
	return snap, nil
}

// matches reports whether the file at path is, as far as its stat can
// tell, the file s was read from: the same device and inode, the same
// size, and the same modification time, with s's read taken far enough
// after that time for the time to be trusted.
func (s *rosterSnapshot) matches(path string) bool {
	if s.readAt.Sub(s.info.ModTime()) < modTimeTrustWindow {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return os.SameFile(info, s.info) &&
		info.Size() == s.info.Size() &&
		info.ModTime().Equal(s.info.ModTime())
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
		member, err := rec.validate()
		if err != nil {
			errs = append(errs, fmt.Errorf("member %d: %w", i, err))
			continue
		}
		if _, dup := byAddress[member.Address]; dup {
			errs = append(errs, fmt.Errorf("member %d: duplicate address", i))
			continue
		}
		if member.StreamerUsername != "" {
			if _, dup := seenStreamer[member.StreamerUsername]; dup {
				errs = append(errs, fmt.Errorf("member %d: duplicate streamer_username", i))
				continue
			}
			seenStreamer[member.StreamerUsername] = i
		}

		byAddress[member.Address] = len(members)
		members = append(members, member)
	}
	if len(errs) > 0 {
		return nil, nil, errors.Join(errs...)
	}
	return members, byAddress, nil
}

// validate validates the fields of one record that do not depend on
// the rest of the roster: the address parses and normalizes, no field
// carries a control character, and every field is bounded. It is the
// field-level half of B9's list; the duplicate checks that need the
// rest of the roster stay in validateRoster's loop and in AddMember's
// locked re-check, since only those two callers know what "the rest of
// the roster" currently is.
func (rec memberRecord) validate() (Member, error) {
	normalized, err := normalizeMemberAddress(rec.Address)
	if err != nil {
		return Member{}, errors.New("address does not parse as a login address")
	}
	if containsControl(rec.DisplayName) {
		return Member{}, errors.New("display name carries a control character")
	}
	if len(rec.DisplayName) > maxDisplayNameLength {
		return Member{}, fmt.Errorf("display name exceeds %d bytes", maxDisplayNameLength)
	}
	if rec.StreamerUsername != "" {
		if !isASCIIPrintable(rec.StreamerUsername) {
			return Member{}, errors.New("streamer_username is not ASCII-printable")
		}
		if len(rec.StreamerUsername) > maxStreamerUsernameLength {
			return Member{}, fmt.Errorf("streamer_username exceeds %d bytes", maxStreamerUsernameLength)
		}
	}
	return Member{
		Address:          normalized,
		DisplayName:      rec.DisplayName,
		StreamerUsername: rec.StreamerUsername,
		Admin:            rec.Admin,
	}, nil
}

// containsControl reports whether s carries a C0 or C1 control
// character or DEL. C1 matters as much as C0: U+009B is a
// single-character CSI that a terminal honors like ESC [.
func containsControl(s string) bool {
	return strings.ContainsFunc(s, unicode.IsControl)
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
// (spec:1161-1163). It re-reads the file first when the file has
// changed, and a re-read that fails returns its load error, never
// ErrMemberNotFound, so a caller can tell a broken roster from an
// unknown address.
func (r *Roster) Lookup(address string) (Member, error) {
	// The parse error is dropped rather than wrapped: net/mail quotes
	// the unparsed remainder of its input, which would carry an address
	// into an error this package promises never names one.
	normalized, err := normalizeMemberAddress(address)
	if err != nil {
		return Member{}, fmt.Errorf("%w: address does not parse", ErrMemberNotFound)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	snap, err := r.current()
	if err != nil {
		return Member{}, err
	}
	idx, ok := snap.byAddress[normalized]
	if !ok {
		return Member{}, ErrMemberNotFound
	}
	return snap.members[idx], nil
}

// List returns a copy of every member on the roster, in file order,
// re-reading the file first when it has changed. A re-read that fails
// returns its load error.
func (r *Roster) List() ([]Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	snap, err := r.current()
	if err != nil {
		return nil, err
	}
	return slices.Clone(snap.members), nil
}

// AddMember validates candidate against B9's list, then, holding the
// roster's directory lock, re-checks candidate's address and
// streamer_username against the file as it is on disk, appends
// candidate to that disk copy, writes the whole roster atomically,
// re-reads it, and logs one audit record naming actor, channel,
// candidate.Address, candidate.Admin, and clientIP. The lock wait ends
// with an error when ctx does. It is the roster's only member write.
//
// Every check runs before the write, so a candidate that fails
// validation, fails the collision re-check, or sets Admin changes no
// byte on disk.
//
// AddMember refuses candidate.Admin on every channel with
// ErrRosterAdminChannelForbidden. Each channel is an identity asserted
// over the network, a service token as much as a cookie or a header,
// and the command-line verb is the only path to an admin. That verb
// brings the admin grant with it; until then no caller can mint one.
func (r *Roster) AddMember(ctx context.Context, candidate Member, actor string, channel identity.Channel, clientIP string, logger *slog.Logger) error {
	if candidate.Admin {
		return fmt.Errorf("%w: channel %s", ErrRosterAdminChannelForbidden, channel)
	}

	member, err := memberRecord(candidate).validate()
	if err != nil {
		return fmt.Errorf("%w: %s: %w", ErrRosterRecordInvalid, r.path, err)
	}

	err = r.update(ctx, func(disk *rosterSnapshot) ([]Member, error) {
		if _, dup := disk.byAddress[member.Address]; dup {
			return nil, fmt.Errorf("%w: %s: duplicate address", ErrRosterRecordInvalid, r.path)
		}
		if member.StreamerUsername != "" {
			for _, existing := range disk.members {
				if existing.StreamerUsername == member.StreamerUsername {
					return nil, fmt.Errorf("%w: %s: duplicate streamer_username", ErrRosterRecordInvalid, r.path)
				}
			}
		}
		return append(slices.Clone(disk.members), member), nil
	})
	if err != nil {
		return err
	}

	logger.Info("roster member added",
		"actor", actor,
		"channel", channel.String(),
		"address", member.Address,
		"admin", member.Admin,
		"client_ip", clientIP,
	)
	return nil
}
