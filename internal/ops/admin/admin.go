// Package admin implements [ops.Admin] for serve: the member and
// settings operations an admin reaches over the cookie mount, over the
// roster in internal/auth and the settings write path in
// internal/settings.
//
// The [Service] adds what the network channel needs beyond those write
// paths, which the command line shares: the re-authentication gate,
// which holds a sensitive settings write and every member write to a
// session signed in within the window; the sensitive-change notice to
// every live admin; and the mapping of each write's refusal to the
// wire code and detail a page branches on. Whether the actor is still
// an admin is decided inside each write's own transaction, never here.
package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// writeWait bounds how long a write waits for the store's write
// connection inside this process. Another process's write lock is
// bounded by the store's busy timeout alone.
const writeWait = 5 * time.Second

// Config is everything the [Service] needs, resolved by the caller. Every
// field is required.
type Config struct {
	// Store is the store the settings write path writes through. The
	// roster writes through the same store.
	Store    *store.Handle
	Registry *config.Registry
	// Environ is the serving process's environment, as [os.Environ]
	// lists it, which a settings write is checked against.
	Environ []string
	Roster  *auth.Roster

	// Window reads the re-authentication window in force. It is read
	// before each gated write, outside the write's transaction.
	Window auth.Setting[time.Duration]

	// Send delivers one mail. It is notify.Mailer's Send in
	// production.
	Send func(ctx context.Context, to, subject, body string) error

	// Now returns the instant a call is evaluated at.
	Now    func() time.Time
	Logger *slog.Logger
}

// A Service is the admin operations' implementation for serve.
type Service struct {
	cfg      Config
	settings map[string]config.Setting
}

var _ ops.Admin = (*Service)(nil)

// New returns a [Service] over cfg.
func New(cfg Config) *Service {
	byKey := make(map[string]config.Setting)
	for _, s := range cfg.Registry.Settings() {
		byKey[s.Key] = s
	}
	return &Service{cfg: cfg, settings: byKey}
}

// requireFresh is the re-authentication gate: it refuses unless the
// session behind actor was signed in, its anchor, less than the window
// in force before now. A refresh never moves the anchor, so staying
// signed in never renews it, and a local actor's zero anchor is never
// fresh. A window that cannot be read refuses the call rather than
// falling back to a default. It first refuses any actor off the cookie
// channel, since the roster treats a command-line actor as the box
// operator and skips its admin check.
func (s *Service) requireFresh(ctx context.Context, actor identity.Actor) error {
	if actor.Channel() != identity.ChannelCookie {
		return ops.NewForbiddenError("this change needs a member's session")
	}
	window, err := s.cfg.Window(ctx)
	if err != nil {
		return fmt.Errorf("read the re-authentication window: %w", err)
	}
	// An anchor ahead of the clock was minted under a clock since stepped
	// back, so its age is unknown and it is no fresher than a stale one.
	if elapsed := s.cfg.Now().Sub(actor.Anchor()); elapsed < 0 || elapsed >= window {
		return ops.NewReauthenticationRequiredError("this change needs a sign-in from the last " + window.String())
	}
	return nil
}

// memberID returns v as a member's user id, refusing the zero id, which
// names no resolved member. The decoder has already refused any text
// that does not spell an id.
func memberID(v id.ID) (identity.UserID, error) {
	if v.IsZero() {
		return identity.UserID{}, ops.NewInvalidRequestError("user_id must be 32 lowercase hex characters naming a member")
	}
	return identity.UserID(v), nil
}

// refusal maps a write's error to the operation's answer: an [*ops.Error]
// wrapped around err, so the wire carries the code, the reason, and
// the detail while the log keeps the cause. An error no case names is
// the server failing and is returned as is, which answers internal.
func refusal(err error) error {
	var answer *ops.Error
	switch {
	case errors.Is(err, store.ErrBusy), errors.Is(err, context.DeadlineExceeded):
		answer = ops.NewUnavailableError("the server is busy; try again").WithDetail(ops.DetailBusy)
	case errors.Is(err, auth.ErrActorNotAdmin):
		answer = ops.NewForbiddenError("")
	case errors.Is(err, auth.ErrMemberNotFound):
		answer = ops.NewNotFoundError("no member has that user_id")
	case errors.Is(err, auth.ErrNoAdmin):
		answer = ops.NewConflictError("the group must keep at least one admin").WithDetail(ops.DetailLastAdmin)
	case errors.Is(err, auth.ErrDuplicateAddress):
		answer = ops.NewConflictError("a member already signs in with that address").WithDetail(ops.DetailDuplicateAddress)
	case errors.Is(err, auth.ErrDuplicateUsername):
		answer = ops.NewConflictError("that username is taken").WithDetail(ops.DetailDuplicateUsername)
	case errors.Is(err, auth.ErrInvalidMemberAddress):
		answer = ops.NewInvalidRequestError("the address is not one a member can sign in with").WithDetail(ops.DetailInvalidAddress)
	case errors.Is(err, auth.ErrInvalidDisplayName):
		answer = ops.NewInvalidRequestError("the display name is too long or carries a hidden character").WithDetail(ops.DetailInvalidDisplayName)
	case errors.Is(err, auth.ErrBlankDisplayName):
		answer = ops.NewInvalidRequestError("the display name is blank").WithDetail(ops.DetailBlankDisplayName)
	case errors.Is(err, auth.ErrInvalidUsername):
		answer = ops.NewInvalidRequestError("the username is too long or has the wrong shape").WithDetail(ops.DetailInvalidUsername)
	case errors.Is(err, auth.ErrMissingUsername):
		answer = ops.NewInvalidRequestError("a username is required").WithDetail(ops.DetailMissingUsername)
	case errors.Is(err, settings.ErrEnvironmentSets):
		answer = ops.NewConflictError("the server's environment file sets this key").WithDetail(ops.DetailEnvironmentSets)
	case errors.Is(err, settings.ErrNotSetting), errors.Is(err, settings.ErrInvalidValue):
		// The seed rule that phrases these reads store-homed keys
		// alone, so the reason never carries a wall or integration
		// value.
		answer = ops.NewInvalidRequestError(err.Error())
	default:
		return err
	}
	return fmt.Errorf("%w: %w", answer, err)
}
