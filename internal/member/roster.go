// Package member holds the group's roster: the users table and each
// member's preferences row. It adds, removes, and demotes members under
// the last-admin guard, and answers the roster reads the server makes.
// Ending a removed member's credentials is the [Revoker]'s job, which
// auth's epoch store does.
package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/store"
)

// maxDisplayNameLength bounds a display name in bytes: 200 bytes
// comfortably covers a legal name in any script, with room for a
// honorific or a nickname, well short of turning the field into free
// text.
const maxDisplayNameLength = 200

// maxUsernameLength bounds username in bytes: 64
// bytes clears every common platform's own username cap (GitHub 39,
// Discord 32) with headroom, while still refusing an attempt to stuff
// arbitrary text into what must stay a stable principal.
const maxUsernameLength = 64

// usernamePattern is the shape the write path enforces for
// username: an alphanumeric first character, refusing a
// leading dot or hyphen at no cost to a plausible account name, then
// any run of alphanumerics, '.', '_', '@', '+', or '-'. '@' and '+'
// admit an email-shaped username, and '_' stays allowed since a
// username is matched by case-insensitive equality, never a SQL LIKE
// pattern that would treat it as a wildcard.
var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._@+-]*$`)

// ErrMemberNotFound reports that an address or id matched no live
// member: Lookup's address once normalized, or the id a removal or a
// demotion names.
var ErrMemberNotFound = errors.New("member not found")

// errRosterRecordInvalid is wrapped by every refusal of a member
// write's own content: a duplicate address or username, a field out of
// bounds, a control character, an address net/mail cannot parse, an
// actor that does not meet its channel's rule, or a write that would
// leave no admin.
var errRosterRecordInvalid = errors.New("roster record invalid")

// errAdminGrantForbidden reports that AddMember refused to set the
// admin flag because the change did not come from the command line,
// the one channel that may grant it.
var errAdminGrantForbidden = errors.New("admin flag may be granted only from the command line")

// errInvalidActor reports that a write's [identity.Actor] failed the
// rule its channel imposes: the zero Actor, or an operator name that
// carries an invisible rune.
var errInvalidActor = errors.New("invalid audit actor or channel")

// The refusals below are the ones internal/ops/admin branches on with
// errors.Is to choose a wire code and detail.
var (
	// ErrDuplicateAddress reports that a candidate's address matches a
	// live member, once both are normalized.
	ErrDuplicateAddress = errors.New("duplicate address")

	// ErrDuplicateUsername reports that a candidate's username matches
	// another row's, live or tombstoned, compared case-insensitively.
	ErrDuplicateUsername = errors.New("duplicate username")

	// ErrInvalidMemberAddress reports that a record's address does not
	// normalize as a login address. A write's refusal carries the
	// normalizer's reason beside it, which names the rule the address
	// broke and never quotes the address.
	ErrInvalidMemberAddress = errors.New("address does not parse as a login address")

	// ErrInvalidDisplayName reports that a record's display name
	// carries a refused rune or exceeds maxDisplayNameLength.
	ErrInvalidDisplayName = errors.New("display name is invalid")

	// ErrBlankDisplayName reports that a write candidate's display name
	// is empty or carries nothing but spaces and zero-width joiners, so
	// it names nobody wherever it is shown.
	ErrBlankDisplayName = errors.New("display name is empty or blank")

	// ErrInvalidUsername reports that a record's username fails its
	// shape check or exceeds maxUsernameLength.
	ErrInvalidUsername = errors.New("username is invalid")

	// ErrMissingUsername reports that a write candidate carries no
	// username, which the write path requires.
	ErrMissingUsername = errors.New("username is required")

	// ErrNoAdmin reports that a write would leave the roster naming no
	// live admin, which would leave nobody able to change it but an
	// operator at the command line.
	ErrNoAdmin = errors.New("the roster must name at least one admin")

	// ErrActorNotAdmin reports that a write arriving over the network
	// named an actor with no live roster record holding the admin flag,
	// as the write itself read the roster.
	ErrActorNotAdmin = errors.New("the actor is not an admin")
)

// Member is one roster entry: a durable id, a login address, a
// display name, a stable principal, and the admin flag. The admin flag
// is the whole of the authorization vocabulary: a later change
// reaching for a role model is adding one, not extending one.
type Member struct {
	// ID is the member's durable id, minted once by the store when the
	// member is first added. It is zero on a candidate not yet written.
	ID identity.UserID
	// Address is the member's login address, normalized by
	// [identity.NormalizeAddress].
	Address string
	// DisplayName is the member's human-readable name.
	DisplayName string
	// Username is the member's stable principal: unique across the
	// roster, compared case-insensitively, and shaped by
	// usernamePattern.
	Username string
	// Admin marks the member as an operator.
	Admin bool
}

// A Revoker ends every credential issued for an address up to a time.
// A roster write calls Revoke last inside its transaction, so an error
// rolls the write back. auth's epoch store is the production Revoker.
type Revoker interface {
	Revoke(ctx context.Context, address string, at time.Time) error
}

// Roster is the member store: it answers a lookup, lists the
// members, and adds, removes, and demotes one, over the store's users
// table.
type Roster struct {
	users   *store.Handle
	revoker Revoker
	now     func() time.Time
	logger  *slog.Logger
}

// NewRoster returns a roster over the users table of the store h
// opens. Each Lookup and List reads the table's committed state, and
// each write runs in one transaction, so another process's writes
// through its own handle are seen at the next call. h must stay open
// for the roster's life.
//
// revoker, now, and logger serve the writes alone: a removal and a
// re-add revoke the address through revoker at the time now returns,
// every write stamps its row with it, and each committed write logs
// its audit record through logger. A roster that only reads, such as
// member list's, passes nil for all three and so opens no auth dir.
func NewRoster(h *store.Handle, revoker Revoker, now func() time.Time, logger *slog.Logger) *Roster {
	return &Roster{users: h, revoker: revoker, now: now, logger: logger}
}

// checkActor refuses the zero Actor, which names nobody, and a local
// actor whose username carries a rune [identity.IsInvisibleRune] names,
// the ordinary space included, since the audit record names it as
// given. A network actor needs no check here: its constructor already
// refused every identity without a durable id.
func checkActor(a identity.Actor) error {
	switch a.Channel() {
	case identity.ChannelCookie:
		return nil
	case identity.ChannelCommandLine:
		if strings.ContainsFunc(a.String(), identity.IsInvisibleRune) {
			return errInvalidActor
		}
		return nil
	default:
		return errInvalidActor
	}
}

// authorizeActor is a's in-write admin check: a network actor must be a
// live admin as tx reads the roster, and a local actor passes, since box
// access is the command line's authority.
func authorizeActor(ctx context.Context, tx *store.Tx, a identity.Actor) error {
	if a.Channel() == identity.ChannelCommandLine {
		return nil
	}
	return CheckAdmin(ctx, tx, a.UserID())
}

// EarliestAdmin returns the id of the live admin created first, breaking
// a tie in creation time by the lower id. It reports false when no live
// admin exists. A new collection is owned by this member until the owner
// is changed. A read error comes back bare: the caller adds the prefix
// its operators already read.
func EarliestAdmin(ctx context.Context, tx *store.Tx) (identity.UserID, bool, error) {
	var id identity.UserID
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE admin = 1 AND tombstoned_at IS NULL
		ORDER BY created_at, id LIMIT 1`).Scan(&id)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return identity.UserID{}, false, nil
	case err != nil:
		return identity.UserID{}, false, err
	}
	return id, true, nil
}

// Live counts the members who can sign in: every row of the users table
// that is not tombstoned.
func Live(ctx context.Context, h *store.Handle) (int, error) {
	var n int
	if err := h.QueryRow(ctx, `SELECT count(*) FROM users WHERE tombstoned_at IS NULL`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// CheckAdmin returns nil when id names a live member holding the admin
// flag as tx reads the users table, [ErrActorNotAdmin] when it does not,
// and the store's error when the read fails. A write
// calls it from inside its own transaction, under the store's write
// lock, so a demotion or removal that committed after a caller's earlier
// check still refuses the write.
func CheckAdmin(ctx context.Context, tx *store.Tx, id identity.UserID) error {
	var admin bool
	err := tx.QueryRow(ctx, `SELECT admin FROM users WHERE id = ? AND tombstoned_at IS NULL`, id).Scan(&admin)
	switch {
	case errors.Is(err, sql.ErrNoRows) || (err == nil && !admin):
		return ErrActorNotAdmin
	case err != nil:
		return fmt.Errorf("member: read the actor's row: %w", err)
	}
	return nil
}

// validateWriteCandidate validates candidate's fields: the address
// parses and normalizes, the display name carries no refused rune, fits
// its bound, and is not blank, and the username is present, fits its
// bound, and matches usernamePattern. It returns the candidate with its
// address normalized.
func validateWriteCandidate(candidate Member) (Member, error) {
	normalized, err := identity.NormalizeAddress(candidate.Address)
	if err != nil {
		return Member{}, fmt.Errorf("%w: %w", ErrInvalidMemberAddress, err)
	}
	if displayNameCarriesRefusedRune(candidate.DisplayName) || len(candidate.DisplayName) > maxDisplayNameLength {
		return Member{}, ErrInvalidDisplayName
	}
	if len(candidate.Username) > maxUsernameLength {
		return Member{}, ErrInvalidUsername
	}
	if strings.TrimFunc(candidate.DisplayName, displayNameAdmitsInvisible) == "" {
		return Member{}, ErrBlankDisplayName
	}
	if candidate.Username == "" {
		return Member{}, ErrMissingUsername
	}
	if !usernamePattern.MatchString(candidate.Username) {
		return Member{}, ErrInvalidUsername
	}
	return Member{
		Address:     normalized,
		DisplayName: candidate.DisplayName,
		Username:    candidate.Username,
		Admin:       candidate.Admin,
	}, nil
}

// zeroWidthNonJoiner and zeroWidthJoiner are the two Cf runes a display
// name admits despite [identity.IsInvisibleRune]: both compose a legal
// name in scripts that use them (Persian, Indic ligatures), and neither
// carries the rendering or comparison hazard the rest of Cf does.
const (
	zeroWidthNonJoiner = '\u200c'
	zeroWidthJoiner    = '\u200d'
)

// displayNameAdmitsInvisible reports whether r is one of the invisible
// runes a display name admits: the ordinary space, the zero-width
// joiner and non-joiner, and the variation selectors, which choose a
// glyph for the rune before them (U+FE0F turns a heart into the emoji
// heart) and never stand alone in a name. None of them makes a name
// visible, so a name of nothing else is blank.
func displayNameAdmitsInvisible(r rune) bool {
	switch r {
	case ' ', zeroWidthNonJoiner, zeroWidthJoiner:
		return true
	}
	return unicode.Is(unicode.Variation_Selector, r)
}

// displayNameCarriesRefusedRune reports whether s carries a rune
// [identity.IsInvisibleRune] refuses and [displayNameAdmitsInvisible]
// does not admit back.
func displayNameCarriesRefusedRune(s string) bool {
	return strings.ContainsFunc(s, func(r rune) bool {
		return identity.IsInvisibleRune(r) && !displayNameAdmitsInvisible(r)
	})
}

// Lookup resolves address to its live roster member, matching
// case-insensitively on the normalized address, with normalization
// happening before any keying, logging, or rate-limit lookup.
// A tombstoned row is [ErrMemberNotFound]. A store
// error is returned as itself, never as ErrMemberNotFound, so a caller
// can tell a broken store from an unknown address.
func (r *Roster) Lookup(ctx context.Context, address string) (Member, error) {
	// The parse error is dropped rather than wrapped: net/mail quotes
	// the unparsed remainder of its input, which would carry an address
	// into an error this package promises never names one.
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return Member{}, fmt.Errorf("%w: address does not parse", ErrMemberNotFound)
	}
	row := r.users.QueryRow(ctx, `SELECT `+memberColumns+` FROM users WHERE address = ? AND tombstoned_at IS NULL`, normalized)
	m, err := scanMember(row.Scan)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Member{}, ErrMemberNotFound
	case err != nil:
		return Member{}, fmt.Errorf("member: look up member: %w", err)
	}
	return m, nil
}

// List returns every live member, ordered by when it was first added
// and then by id, so two rows added in one tick still list in one
// stable order. A failed read returns its error.
func (r *Roster) List(ctx context.Context) ([]Member, error) {
	rows, err := r.users.Query(ctx, `SELECT `+memberColumns+` FROM users WHERE tombstoned_at IS NULL ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("member: list members: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var members []Member
	for rows.Next() {
		m, err := scanMember(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("member: list members: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("member: list members: %w", err)
	}
	return members, nil
}

// AddMember validates by, then candidate against the field rules, then
// writes candidate and logs one audit record naming by, its channel,
// candidate.Address, candidate.Admin, and by's client address, which
// reads "local" for an operator.
//
// The write is one store transaction, so ctx must carry a deadline. A
// new address gets a fresh id; an address whose row is tombstoned is
// reactivated under its old id, every field taken from candidate, and
// the address is revoked as a removal revokes it (see
// [Roster.RemoveMember]). No credential issued before the re-add
// verifies after it, even when the removal's revocation was lost.
// A username matching any other row, live or tombstoned, compared
// case-insensitively, is refused.
//
// Only an operator may set candidate.Admin; a member actor is refused
// with errAdminGrantForbidden before anything is read, so a re-add over
// the network returns a removed admin without the flag. A member actor
// must also hold the admin flag in the state the write decides on, read
// inside the write, so a demotion after a caller checked the actor
// still refuses it. Whatever the channel, the table written must hold
// at least one live admin.
//
// Every check runs before the write commits, so a refused candidate or
// actor changes nothing. A nil error returns the member as stored, its
// ID the one the write inserted or reactivated.
func (r *Roster) AddMember(ctx context.Context, candidate Member, by identity.Actor) (Member, error) {
	member, err := checkAddition(candidate, by)
	if err != nil {
		return Member{}, err
	}
	member, err = r.addUser(ctx, member, by)
	if err != nil {
		return Member{}, err
	}
	r.logChange(ctx, "roster member added", by, member.Address, "admin", member.Admin)
	return member, nil
}

// checkAddition runs every check on an addition that needs no roster:
// the admin grant's channel, by's rule for its channel, and the
// candidate's own fields. It returns the candidate as validated.
func checkAddition(candidate Member, by identity.Actor) (Member, error) {
	if candidate.Admin && by.Channel() != identity.ChannelCommandLine {
		return Member{}, fmt.Errorf("%w: channel %s", errAdminGrantForbidden, by.Channel())
	}
	if err := checkActor(by); err != nil {
		return Member{}, fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	member, err := validateWriteCandidate(candidate)
	if err != nil {
		return Member{}, fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	return member, nil
}

// memberColumns are the users columns a Member is scanned from, in
// scanMember's order.
const memberColumns = `id, address, display_name, username, admin`

// scanMember scans one users row, selected as memberColumns, through
// scan.
func scanMember(scan func(dest ...any) error) (Member, error) {
	var m Member
	if err := scan(&m.ID, &m.Address, &m.DisplayName, &m.Username, &m.Admin); err != nil {
		return Member{}, err
	}
	return m, nil
}

// addUser writes member, already validated, into the users table in one
// transaction. Every row the write decides on is read inside that
// transaction, which holds the store's write lock from its start, so a
// concurrent writer in this process or another cannot change the
// answer between the read and the write. In order: by must pass its
// in-write admin check; a live row on member's address refuses the
// add, and a tombstoned one is reactivated under its own id with every
// field, the admin flag included, taken from member; any other row,
// live or tombstoned, holding member's username compared
// case-insensitively refuses the add; the table must hold a live admin
// once the write lands; and a reactivation then revokes the address,
// last, as a removal does. It returns member with the ID the row holds.
func (r *Roster) addUser(ctx context.Context, member Member, by identity.Actor) (Member, error) {
	err := r.users.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if err := authorizeActor(ctx, tx, by); err != nil {
			return err
		}

		var (
			existing   identity.UserID
			tombstoned bool
		)
		err := tx.QueryRow(ctx, `SELECT id, tombstoned_at IS NOT NULL FROM users WHERE address = ?`, member.Address).Scan(&existing, &tombstoned)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// A new address: existing stays zero, and the write inserts.
		case err != nil:
			return fmt.Errorf("member: read the address's row: %w", err)
		case !tombstoned:
			return fmt.Errorf("%w: %w", errRosterRecordInvalid, ErrDuplicateAddress)
		}

		var taken bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = ? COLLATE NOCASE AND address != ?)`,
			member.Username, member.Address).Scan(&taken); err != nil {
			return fmt.Errorf("member: read the username's rows: %w", err)
		}
		if taken {
			return fmt.Errorf("%w: %w", errRosterRecordInvalid, ErrDuplicateUsername)
		}

		now := r.now()
		at := now.UnixMicro()
		reactivate := !existing.IsZero()
		if reactivate {
			member.ID = existing
			_, err = tx.Exec(ctx, `UPDATE users SET display_name = ?, username = ?, admin = ?, tombstoned_at = NULL, updated_at = ? WHERE id = ?`,
				member.DisplayName, member.Username, member.Admin, at, member.ID)
		} else {
			member.ID = identity.UserID(id.New())
			_, err = tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				member.ID, member.Address, member.DisplayName, member.Username, member.Admin, at, at)
		}
		if err != nil {
			return fmt.Errorf("member: write the member's row: %w", err)
		}

		if err := requireLiveAdmin(ctx, tx); err != nil {
			if errors.Is(err, ErrNoAdmin) {
				return fmt.Errorf("%w: add the first admin with dubplate-server member add --admin", err)
			}
			return err
		}
		if !reactivate {
			return nil
		}
		return r.revoker.Revoke(ctx, member.Address, now)
	})
	if err != nil {
		return Member{}, err
	}
	return member, nil
}

// RemoveMember removes the live member id names: it tombstones their
// row with the removal time, deletes their member_preferences row, and
// revokes every credential issued for their address up to now through
// the roster's [Revoker]; then it logs one
// audit record naming by, its channel, the address, and by's client
// address. Their member_overrides and member_quota rows stay: the
// overrides are the admin's policy and return with a re-add, and the
// quota reflects audio that stays.
//
// The write is one store transaction, so ctx must carry a deadline. by
// passes the in-write admin check first, as for [Roster.AddMember]. An
// id with no row, or a tombstoned one, is [ErrMemberNotFound]. A
// removal that would leave no live admin, a sole admin's self-removal
// included, is [ErrNoAdmin]. Every refusal changes nothing.
//
// The revocation is outside the store, so "one transaction" holds by
// ordering: Revoke runs inside the write, after the tombstone and the
// last-admin guard and before the commit. A failed revocation rolls
// the removal back. A commit that fails after one leaves a live member
// signed out, the safe direction.
func (r *Roster) RemoveMember(ctx context.Context, id identity.UserID, by identity.Actor) error {
	if err := checkActor(by); err != nil {
		return fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	var address string
	err := r.users.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if err := authorizeActor(ctx, tx, by); err != nil {
			return err
		}
		target, err := liveMember(ctx, tx, id)
		if err != nil {
			return err
		}
		address = target.Address
		now := r.now()
		if _, err := tx.Exec(ctx, `UPDATE users SET tombstoned_at = ?, updated_at = ? WHERE id = ?`, now.UnixMicro(), now.UnixMicro(), id); err != nil {
			return fmt.Errorf("member: tombstone the member's row: %w", err)
		}
		if _, err := tx.Exec(ctx, `DELETE FROM member_preferences WHERE user_id = ?`, id); err != nil {
			return fmt.Errorf("member: delete the member's preferences: %w", err)
		}
		if err := requireLiveAdmin(ctx, tx); err != nil {
			return err
		}
		return r.revoker.Revoke(ctx, address, now)
	})
	if err != nil {
		return err
	}
	r.logChange(ctx, "roster member removed", by, address)
	return nil
}

// DemoteMember clears the admin flag of the live member id names and
// logs one audit record naming by, its channel, the address, and by's
// client address. Demoting a member who holds no flag writes and logs
// nothing and returns nil. A demotion ends no session: every request
// reads the flag afresh.
//
// The write is one store transaction, so ctx must carry a deadline. by
// passes the in-write admin check first, as for [Roster.AddMember]. An
// id with no row, or a tombstoned one, is [ErrMemberNotFound], and a
// demotion that would leave no live admin is [ErrNoAdmin], read inside
// the write, so of two concurrent demotions of the last two admins the
// second is refused. Every refusal changes nothing.
func (r *Roster) DemoteMember(ctx context.Context, id identity.UserID, by identity.Actor) error {
	if err := checkActor(by); err != nil {
		return fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	var target Member
	err := r.users.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if err := authorizeActor(ctx, tx, by); err != nil {
			return err
		}
		var err error
		if target, err = liveMember(ctx, tx, id); err != nil || !target.Admin {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE users SET admin = 0, updated_at = ? WHERE id = ?`, r.now().UnixMicro(), id); err != nil {
			return fmt.Errorf("member: clear the member's admin flag: %w", err)
		}
		return requireLiveAdmin(ctx, tx)
	})
	if err != nil || !target.Admin {
		return err
	}
	r.logChange(ctx, "roster member demoted", by, target.Address)
	return nil
}

// liveMember reads id's live row through tx, answering
// [ErrMemberNotFound] for an id with no row or a tombstoned one.
func liveMember(ctx context.Context, tx *store.Tx, id identity.UserID) (Member, error) {
	m, err := scanMember(tx.QueryRow(ctx, `SELECT `+memberColumns+` FROM users WHERE id = ? AND tombstoned_at IS NULL`, id).Scan)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Member{}, ErrMemberNotFound
	case err != nil:
		return Member{}, fmt.Errorf("member: read the member's row: %w", err)
	}
	return m, nil
}

// requireLiveAdmin is the last-admin guard. It runs after a write's own
// statements, inside its transaction, and refuses with [ErrNoAdmin]
// unless the users table then holds a live admin.
func requireLiveAdmin(ctx context.Context, tx *store.Tx) error {
	var adminLive bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE admin = 1 AND tombstoned_at IS NULL)`).Scan(&adminLive); err != nil {
		return fmt.Errorf("member: read the admins: %w", err)
	}
	if !adminLive {
		return fmt.Errorf("%w: %w", errRosterRecordInvalid, ErrNoAdmin)
	}
	return nil
}

// logChange writes the audit record for a committed roster write: msg,
// then by, its channel, the address, any further key and value pairs in
// args, and by's client address.
func (r *Roster) logChange(ctx context.Context, msg string, by identity.Actor, address string, args ...any) {
	attrs := append([]any{"actor", by.String(), "channel", by.Channel().String(), "address", address}, args...)
	r.logger.Log(ctx, slog.LevelInfo, msg, append(attrs, "client_ip", by.ClientIP())...)
}
