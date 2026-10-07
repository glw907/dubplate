package auth

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

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
)

// maxDisplayNameLength bounds a display name in bytes: 200 bytes
// comfortably covers a legal name in any script, with room for a
// honorific or a nickname, well short of turning the field into free
// text.
const maxDisplayNameLength = 200

// maxUsernameLength bounds username in bytes: 64
// bytes clears every common platform's own username cap (GitHub 39,
// Discord 32) with headroom, while still refusing an attempt to stuff
// arbitrary text into what R3 requires as a stable principal.
const maxUsernameLength = 64

// usernamePattern is the shape the write path enforces for
// username: an alphanumeric first character, refusing a
// leading dot or hyphen at no cost to a plausible account name, then
// any run of alphanumerics, '.', '_', '@', '+', or '-'. '@' and '+'
// admit an email-shaped username, and '_' stays allowed since a
// username is matched by case-insensitive equality, never a SQL LIKE
// pattern that would treat it as a wildcard.
var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._@+-]*$`)

// errMemberNotFound reports that Lookup's address matched no live
// member once normalized.
var errMemberNotFound = errors.New("auth: member not found")

// errRosterRecordInvalid is wrapped by every refusal of a member
// write: a duplicate address or username, a field out of bounds, a
// control character, an address net/mail cannot parse, an actor that
// does not meet its channel's rule, or a write that would leave no
// admin.
var errRosterRecordInvalid = errors.New("auth: roster record invalid")

// errAdminGrantForbidden reports that AddMember refused to set the
// admin flag because the change did not come from the command line,
// the one channel that may grant it.
var errAdminGrantForbidden = errors.New("auth: admin flag may be granted only from the command line")

// The refusals below each wrap errRosterRecordInvalid, so a caller
// outside this package branches on that one sentinel. They are
// unexported because only code in this package, such as
// additionOutcome, branches on which one fired, and it may test either
// the specific refusal or errRosterRecordInvalid.
var (
	// errDuplicateAddress reports that a candidate's address matches a
	// live member, once both are normalized.
	errDuplicateAddress = errors.New("auth: duplicate address")

	// errDuplicateUsername reports that a candidate's username matches
	// another row's, live or tombstoned, compared case-insensitively.
	errDuplicateUsername = errors.New("auth: duplicate username")

	// errInvalidMemberAddress reports that a record's address does not
	// normalize as a login address.
	errInvalidMemberAddress = errors.New("auth: address does not parse as a login address")

	// errInvalidDisplayName reports that a record's display name
	// carries a refused rune or exceeds maxDisplayNameLength.
	errInvalidDisplayName = errors.New("auth: display name is invalid")

	// errBlankDisplayName reports that a write candidate's display name
	// is empty or carries nothing but spaces and zero-width joiners, so
	// it names nobody wherever it is shown.
	errBlankDisplayName = errors.New("auth: display name is empty or blank")

	// errInvalidUsername reports that a record's
	// username fails its shape check or exceeds
	// maxUsernameLength.
	errInvalidUsername = errors.New("auth: username is invalid")

	// errMissingUsername reports that a write candidate carries
	// no username, which the write path requires.
	errMissingUsername = errors.New("auth: username is required")

	// errInvalidActor reports that AddMember's actor or channel failed
	// the rule its channel imposes: an unrecognized channel, or an
	// actor that does not meet the shape its channel requires.
	errInvalidActor = errors.New("auth: invalid audit actor or channel")

	// errNoAdmin reports that a write would leave the roster naming no
	// admin, which would leave nobody able to change it but an operator
	// writing the store by hand.
	errNoAdmin = errors.New("auth: the roster must name at least one admin")

	// errActorNotAdmin reports that a change arriving over the network
	// named an actor with no live roster record holding the admin flag,
	// as the write itself reads the roster.
	errActorNotAdmin = errors.New("auth: the actor is not an admin")
)

// Member is one roster entry (R1): a durable id, a login address, a
// display name, a stable principal, and the admin flag. The admin flag
// is the whole of rung 5's authorization vocabulary
// (spec:1111-1115): a later rung reaching for a role model is adding
// one, not extending one.
type Member struct {
	// ID is the member's durable id, minted once by the store when the
	// member is first added. It is zero on a candidate not yet written.
	ID identity.UserID
	// Address is the member's login address, normalized by
	// normalizeMemberAddress.
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

// Roster is the member store (R1): it answers a lookup, lists the
// members, and adds one, over the store's users table.
type Roster struct {
	users *store.Handle
}

// NewRoster returns a roster over the users table of the store h
// opens. Each Lookup and List reads the table's committed state, and
// AddMember writes it in one transaction, so another process's writes
// through its own handle are seen at the next call. h must stay open
// for the roster's life.
func NewRoster(h *store.Handle) *Roster {
	return &Roster{users: h}
}

// normalizeMemberAddress is the roster's single resolution point: the
// only place a login address becomes the value the roster keys,
// compares, and looks members up by.
func normalizeMemberAddress(address string) (string, error) {
	return identity.NormalizeAddress(address)
}

// validateWriteCandidate validates candidate's fields: the address
// parses and normalizes, the display name carries no refused rune, fits
// its bound, and is not blank, and the username is present, fits its
// bound, and matches usernamePattern. It returns the candidate with its
// address normalized.
func validateWriteCandidate(candidate Member) (Member, error) {
	normalized, err := normalizeMemberAddress(candidate.Address)
	if err != nil {
		return Member{}, errInvalidMemberAddress
	}
	if displayNameCarriesRefusedRune(candidate.DisplayName) || len(candidate.DisplayName) > maxDisplayNameLength {
		return Member{}, errInvalidDisplayName
	}
	if len(candidate.Username) > maxUsernameLength {
		return Member{}, errInvalidUsername
	}
	if strings.TrimFunc(candidate.DisplayName, displayNameAdmitsInvisible) == "" {
		return Member{}, errBlankDisplayName
	}
	if candidate.Username == "" {
		return Member{}, errMissingUsername
	}
	if !usernamePattern.MatchString(candidate.Username) {
		return Member{}, errInvalidUsername
	}
	return Member{
		Address:     normalized,
		DisplayName: candidate.DisplayName,
		Username:    candidate.Username,
		Admin:       candidate.Admin,
	}, nil
}

// validateActor checks actor against the rule its channel imposes on
// an audit record's actor, and returns actor in the form the record
// names it. An unrecognized channel refuses outright. A cookie carries
// a login over the network, so its actor must normalize as a login
// address through [identity.NormalizeAddress], and the record names
// the normalized address. The command line names an operating-system
// user, which is no address, so its actor need only be non-empty and
// free of every rune [identity.IsInvisibleRune] names, the ordinary
// space included, and the record names it as given.
func validateActor(actor string, channel identity.Channel) (string, error) {
	switch channel {
	case identity.ChannelCookie:
		normalized, err := identity.NormalizeAddress(actor)
		if err != nil {
			return "", errInvalidActor
		}
		return normalized, nil
	case identity.ChannelCommandLine:
		if actor == "" || strings.ContainsFunc(actor, identity.IsInvisibleRune) {
			return "", errInvalidActor
		}
		return actor, nil
	default:
		return "", errInvalidActor
	}
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
// happening before any keying, logging, or rate-limit lookup
// (spec:1161-1163). A tombstoned row is errMemberNotFound. A store
// error is returned as itself, never as errMemberNotFound, so a caller
// can tell a broken store from an unknown address.
func (r *Roster) Lookup(ctx context.Context, address string) (Member, error) {
	// The parse error is dropped rather than wrapped: net/mail quotes
	// the unparsed remainder of its input, which would carry an address
	// into an error this package promises never names one.
	normalized, err := normalizeMemberAddress(address)
	if err != nil {
		return Member{}, fmt.Errorf("%w: address does not parse", errMemberNotFound)
	}
	row := r.users.QueryRow(ctx, `SELECT `+memberColumns+` FROM users WHERE address = ? AND tombstoned_at IS NULL`, normalized)
	m, err := scanMember(row.Scan)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return Member{}, errMemberNotFound
	case err != nil:
		return Member{}, fmt.Errorf("auth: look up member: %w", err)
	}
	return m, nil
}

// List returns every live member, ordered by when it was first added
// and then by id, so two rows added in one tick still list in one
// stable order. A failed read returns its error.
func (r *Roster) List(ctx context.Context) ([]Member, error) {
	rows, err := r.users.Query(ctx, `SELECT `+memberColumns+` FROM users WHERE tombstoned_at IS NULL ORDER BY created_at, id`)
	if err != nil {
		return nil, fmt.Errorf("auth: list members: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var members []Member
	for rows.Next() {
		m, err := scanMember(rows.Scan)
		if err != nil {
			return nil, fmt.Errorf("auth: list members: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("auth: list members: %w", err)
	}
	return members, nil
}

// AddMember validates actor and channel, then candidate against the
// field rules, then writes candidate and logs one audit record naming
// actor, channel, candidate.Address, candidate.Admin, and clientIP. On
// identity.ChannelCookie the record names actor normalized. On
// identity.ChannelCommandLine the record's client_ip reads "local"
// whatever clientIP holds, so every record keeps one shape.
//
// The write is one store transaction, so ctx must carry a deadline. A
// new address gets a fresh id; an address whose row is tombstoned is
// reactivated under its old id, every field taken from candidate. A
// username matching any other row, live or tombstoned, compared
// case-insensitively, is refused.
//
// Only identity.ChannelCommandLine may set candidate.Admin; any other
// channel is refused with errAdminGrantForbidden before anything is
// read. On any other channel the actor must also hold the admin flag in
// the state the write decides on, read inside the write, so a demotion
// after a caller checked the actor still refuses it. Whatever the
// channel, the table written must hold at least one live admin.
//
// Every check runs before the write lands, so a refused candidate,
// actor, or channel changes nothing.
func (r *Roster) AddMember(ctx context.Context, candidate Member, actor string, channel identity.Channel, clientIP string, logger *slog.Logger) error {
	member, actor, err := checkAddition(candidate, actor, channel)
	if err != nil {
		return err
	}
	if err := r.addUser(ctx, member, actor, channel); err != nil {
		return err
	}
	logAddition(ctx, logger, member, actor, channel, clientIP)
	return nil
}

// checkAddition runs every check on an addition that needs no roster:
// the admin grant's channel, the actor's rule for its channel, and the
// candidate's own fields. It returns the candidate as validated and the
// actor as validateActor returns it.
func checkAddition(candidate Member, actor string, channel identity.Channel) (Member, string, error) {
	if candidate.Admin && channel != identity.ChannelCommandLine {
		return Member{}, "", fmt.Errorf("%w: channel %s", errAdminGrantForbidden, channel)
	}
	actor, err := validateActor(actor, channel)
	if err != nil {
		return Member{}, "", fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	member, err := validateWriteCandidate(candidate)
	if err != nil {
		return Member{}, "", fmt.Errorf("%w: %w", errRosterRecordInvalid, err)
	}
	return member, actor, nil
}

// memberColumns are the users columns a Member is scanned from, in
// scanMember's order.
const memberColumns = `id, address, display_name, username, admin`

// scanMember scans one users row, selected as memberColumns, through
// scan.
func scanMember(scan func(dest ...any) error) (Member, error) {
	var (
		m  Member
		id []byte
	)
	if err := scan(&id, &m.Address, &m.DisplayName, &m.Username, &m.Admin); err != nil {
		return Member{}, err
	}
	if len(id) != len(m.ID) {
		return Member{}, fmt.Errorf("auth: a users row's id is %d bytes, want %d", len(id), len(m.ID))
	}
	m.ID = identity.UserID(id)
	return m, nil
}

// addUser writes member, already validated, into the users table in one
// transaction. Every row the write decides on is read inside that
// transaction, which holds the store's write lock from its start, so a
// concurrent writer in this process or another cannot change the
// answer between the read and the write. In order: an actor on any
// channel but the command line must have a live row holding the admin
// flag; a live row on member's address refuses the add, and a
// tombstoned one is reactivated under its own id with every field,
// the admin flag included, taken from member; any other row, live or
// tombstoned, holding member's username compared case-insensitively
// refuses the add; and the table must hold a live admin once the write
// lands.
func (r *Roster) addUser(ctx context.Context, member Member, actor string, channel identity.Channel) error {
	return r.users.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if channel != identity.ChannelCommandLine {
			// validateActor has already normalized a cookie actor.
			var admin bool
			err := tx.QueryRow(ctx, `SELECT admin FROM users WHERE address = ? AND tombstoned_at IS NULL`, actor).Scan(&admin)
			switch {
			case errors.Is(err, sql.ErrNoRows) || (err == nil && !admin):
				return fmt.Errorf("%w: %w", errRosterRecordInvalid, errActorNotAdmin)
			case err != nil:
				return fmt.Errorf("auth: read the actor's row: %w", err)
			}
		}

		var (
			id         []byte
			tombstoned bool
		)
		err := tx.QueryRow(ctx, `SELECT id, tombstoned_at IS NOT NULL FROM users WHERE address = ?`, member.Address).Scan(&id, &tombstoned)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			// A new address: id stays nil, and the write inserts.
		case err != nil:
			return fmt.Errorf("auth: read the address's row: %w", err)
		case !tombstoned:
			return fmt.Errorf("%w: %w", errRosterRecordInvalid, errDuplicateAddress)
		}

		var taken bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE username = ? COLLATE NOCASE AND address != ?)`,
			member.Username, member.Address).Scan(&taken); err != nil {
			return fmt.Errorf("auth: read the username's rows: %w", err)
		}
		if taken {
			return fmt.Errorf("%w: %w", errRosterRecordInvalid, errDuplicateUsername)
		}

		at := time.Now().UnixMicro()
		if id != nil {
			_, err = tx.Exec(ctx, `UPDATE users SET display_name = ?, username = ?, admin = ?, tombstoned_at = NULL, updated_at = ? WHERE id = ?`,
				member.DisplayName, member.Username, member.Admin, at, id)
		} else {
			_, err = tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)`,
				store.NewID(), member.Address, member.DisplayName, member.Username, member.Admin, at, at)
		}
		if err != nil {
			return fmt.Errorf("auth: write the member's row: %w", err)
		}

		var adminLive bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE admin = 1 AND tombstoned_at IS NULL)`).Scan(&adminLive); err != nil {
			return fmt.Errorf("auth: read the admins: %w", err)
		}
		if !adminLive {
			return fmt.Errorf("%w: %w: add the first admin with dubplate-server member add --admin", errRosterRecordInvalid, errNoAdmin)
		}
		return nil
	})
}

// logAddition writes the audit record for member's addition.
func logAddition(ctx context.Context, logger *slog.Logger, member Member, actor string, channel identity.Channel, clientIP string) {
	if channel == identity.ChannelCommandLine {
		clientIP = "local"
	}
	logger.Log(ctx, slog.LevelInfo, "roster member added",
		"actor", actor,
		"channel", channel.String(),
		"address", member.Address,
		"admin", member.Admin,
		"client_ip", clientIP,
	)
}
