package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// Form fields the handlers read. Both pages post a urlencoded form,
// so neither field is ever carried in a query string this package
// reads.
const (
	addressField = "address"
	tokenField   = "token"
)

// linkMailSubject and linkMailBody are the magic-link mail's text. The
// body names no site, no person, and no deployment value, because the
// same text goes out from every install; the link is the whole of what
// differs.
const (
	linkMailSubject = "Your sign-in link"
	linkMailBody    = "Open this link to sign in:\n\n%s\n\nThe link works once and expires soon. If you did not ask to sign in, ignore this message.\n"
)

// Outcomes the handlers answer with, in the "outcome" field of every
// JSON response. The copy a person reads is the page's; these are the
// values a page branches on.
const (
	outcomeSent        = "sent"
	outcomeUnknown     = "unknown"
	outcomeRequested   = "requested"
	outcomeFailed      = "failed"
	outcomeRateLimited = "rate_limited"
	outcomeInvalid     = "invalid_address"
	outcomeTooLarge    = "too_large"
	outcomeMalformed   = "malformed"
	outcomeSignedIn    = "signed_in"
	outcomeRefused     = "refused"
	outcomeSignedOut   = "signed_out"
)

// HandlersConfig is everything [NewHandlers] needs, resolved by the
// caller. Every field is required except PreviousLinkSecret.
type HandlersConfig struct {
	Roster  *Roster
	Nonces  *NonceStore
	Epoch   *EpochStore
	Limiter *Limiter

	// Send delivers one mail. It is notify.Mailer's Send in
	// production.
	Send func(ctx context.Context, to, subject, body string) error

	// ClientIP returns the client address a request is held
	// accountable for, under the shell's trusted-proxy policy. Every
	// per-client limit and every auth event's log record uses it, and
	// the handlers read no forwarding header of their own.
	ClientIP func(r *http.Request) string

	// Now returns the instant a request is evaluated at.
	Now func() time.Time

	Logger *slog.Logger

	// LinkSecret and PreviousLinkSecret sign and verify magic links;
	// PreviousLinkSecret may be empty. SessionSecret signs the session
	// a consumed link mints.
	LinkSecret, PreviousLinkSecret, SessionSecret []byte

	LinkExpiry, SessionLifetime time.Duration

	// PublicOrigin is the configured public URL every mailed link is
	// built from.
	PublicOrigin string

	// BodyCap is the byte limit on every request body, applied before
	// any parser reads it.
	BodyCap int64

	// GenericAnswer selects the generic unknown-address answer: a known
	// address and an unknown one get the same response, sent before any
	// mail is. When false, an unknown address is told so and handed the
	// roster's admin contact, and a known address whose mail fails is
	// told its request failed.
	GenericAnswer bool
}

// Handlers is the three auth endpoints, as values the shell mounts.
// This package registers no route; the method and path each is served
// on are the mount's.
type Handlers struct {
	// RequestLink takes a posted address and mails it a magic link.
	RequestLink http.Handler

	// ConsumeLink takes a posted magic-link token and, when it is
	// valid and unconsumed, sets a session cookie. It accepts no
	// destination: where a signed-in member lands is the page's
	// default, never a value a request supplies.
	ConsumeLink http.Handler

	// SignOut clears the session cookie. It does not revoke the
	// session: sessions are stateless, so a copy of the cookie taken
	// before sign-out keeps working until its own expiry or until the
	// member's revocation epoch is bumped.
	SignOut http.Handler
}

// NewHandlers returns the auth endpoints built on cfg.
func NewHandlers(cfg HandlersConfig) Handlers {
	h := &handlers{cfg: cfg}
	return Handlers{
		RequestLink: http.HandlerFunc(h.requestLink),
		ConsumeLink: http.HandlerFunc(h.consumeLink),
		SignOut:     http.HandlerFunc(h.signOut),
	}
}

type handlers struct {
	cfg HandlersConfig
}

// answer is every response's JSON body.
type answer struct {
	Outcome      string        `json:"outcome"`
	AdminContact *adminContact `json:"admin_contact,omitempty"`
}

type adminContact struct {
	DisplayName string `json:"display_name"`
	Address     string `json:"address"`
}

// requestLink runs a fixed order: cap the body, parse it, normalize
// the address, charge both link-request limits, resolve the address
// against the roster, and only then spend mail.
//
// Under the generic answer the response is written before the mail is
// attempted and does not depend on its result. Waiting would hand a
// stranger the membership the generic answer exists to hide, twice
// over: a known address's failed send would answer differently from an
// unknown address, and a known address would take an SMTP round trip
// longer to answer.
func (h *handlers) requestLink(w http.ResponseWriter, r *http.Request) {
	now := h.cfg.Now()
	clientIP := h.cfg.ClientIP(r)
	log := h.cfg.Logger.With("client_ip", clientIP)

	if !h.readForm(w, r, log) {
		return
	}
	// The normalizer's error can quote its input, so it is dropped
	// rather than logged.
	address, err := identity.NormalizeAddress(r.PostForm.Get(addressField))
	if err != nil {
		log.Info("auth link request refused", "reason", "address does not parse")
		writeAnswer(w, http.StatusBadRequest, answer{Outcome: outcomeInvalid})
		return
	}
	if err := h.cfg.Limiter.AllowLinkRequest(now, address, clientIP); err != nil {
		log.Warn("auth rate limit tripped", "reason", err.Error())
		writeAnswer(w, http.StatusTooManyRequests, answer{Outcome: outcomeRateLimited})
		return
	}
	member, lookupErr := h.cfg.Roster.Lookup(address)
	known := lookupErr == nil

	if h.cfg.GenericAnswer {
		log.Info("auth link requested", "known", known)
		writeAnswer(w, http.StatusOK, answer{Outcome: outcomeRequested})
		if known {
			// The send outlives the request, so it runs under a
			// context the request's end does not cancel. The mail
			// transport bounds every dial and command with its own
			// deadline, which is what bounds the goroutine.
			ctx := context.WithoutCancel(r.Context())
			go func() {
				if err := h.sendLink(ctx, now, member.Address); err != nil {
					h.logSendFailure(log, err)
				}
			}()
		}
		return
	}

	if !known {
		log.Info("auth link requested", "known", false)
		writeAnswer(w, http.StatusOK, answer{Outcome: outcomeUnknown, AdminContact: h.adminContact()})
		return
	}
	if err := h.sendLink(r.Context(), now, member.Address); err != nil {
		h.logSendFailure(log, err)
		if errors.Is(err, ErrRateLimited) {
			writeAnswer(w, http.StatusTooManyRequests, answer{Outcome: outcomeRateLimited})
			return
		}
		writeAnswer(w, http.StatusInternalServerError, answer{Outcome: outcomeFailed})
		return
	}
	log.Info("auth link requested", "known", true)
	writeAnswer(w, http.StatusOK, answer{Outcome: outcomeSent})
}

// sendLink charges the global mail ceiling, mints a link for address,
// and mails it.
func (h *handlers) sendLink(ctx context.Context, now time.Time, address string) error {
	if err := h.cfg.Limiter.AllowLinkMail(now); err != nil {
		return err
	}
	token, _, err := MintLink(h.cfg.LinkSecret, address, now, h.cfg.LinkExpiry)
	if err != nil {
		return err
	}
	link, err := LinkURL(h.cfg.PublicOrigin, token)
	if err != nil {
		return err
	}
	return h.cfg.Send(ctx, address, linkMailSubject, fmt.Sprintf(linkMailBody, link))
}

func (h *handlers) logSendFailure(log *slog.Logger, err error) {
	if errors.Is(err, ErrRateLimited) {
		log.Warn("auth rate limit tripped", "reason", err.Error())
		return
	}
	log.Error("auth link send failed", "reason", err.Error())
}

// adminContact returns the first admin on the roster, or nil when the
// roster names none.
func (h *handlers) adminContact() *adminContact {
	for _, m := range h.cfg.Roster.List() {
		if m.Admin {
			return &adminContact{DisplayName: m.DisplayName, Address: m.Address}
		}
	}
	return nil
}

// consumeLink runs a fixed order, and no step may move: the per-client
// limit, then the token's MAC and expiry, then the roster, then the
// revocation epoch, then the nonce insert, and only then the session.
//
// The limit comes first because a forged token fails at the MAC, so a
// limit anywhere after it never sees the flood it exists to bound. The
// roster and epoch checks precede the insert so a link for a removed or
// revoked member consumes nothing, and the session is minted only once
// the insert has succeeded, so a replay can never mint a second one.
//
// Every refusal after the limit answers the same bytes, whatever the
// cause, and any error at all refuses; the cause reaches the server log
// alone. The log names the path and never the query, which a mailed
// link carries its token in.
func (h *handlers) consumeLink(w http.ResponseWriter, r *http.Request) {
	now := h.cfg.Now()
	clientIP := h.cfg.ClientIP(r)
	log := h.cfg.Logger.With("client_ip", clientIP, "path", r.URL.Path)

	if err := h.cfg.Limiter.AllowConsumption(now, clientIP); err != nil {
		log.Warn("auth rate limit tripped", "reason", err.Error())
		writeAnswer(w, http.StatusTooManyRequests, answer{Outcome: outcomeRateLimited})
		return
	}
	if !h.readForm(w, r, log) {
		return
	}

	value, sess, err := h.consume(now, r.PostForm.Get(tokenField))
	if err != nil {
		log.Info("auth link consumption failed", "reason", err.Error())
		writeAnswer(w, http.StatusBadRequest, answer{Outcome: outcomeRefused})
		return
	}
	http.SetCookie(w, SessionCookie(value, now, sess.ExpiresAt))
	log.Info("auth link consumed")
	writeAnswer(w, http.StatusOK, answer{Outcome: outcomeSignedIn})
}

// consume verifies token and consumes its nonce, returning the signed
// session it mints. Any non-nil error is a refusal.
func (h *handlers) consume(now time.Time, token string) (string, Session, error) {
	tok, err := VerifyLink(h.cfg.LinkSecret, h.cfg.PreviousLinkSecret, token, now)
	if err != nil {
		return "", Session{}, err
	}
	if _, err := h.cfg.Roster.Lookup(tok.Address); err != nil {
		return "", Session{}, err
	}
	if err := h.cfg.Epoch.CheckLink(tok); err != nil {
		return "", Session{}, err
	}
	if err := h.cfg.Nonces.Insert(tok.Nonce, now, tok.ExpiresAt); err != nil {
		return "", Session{}, err
	}
	return MintSession(h.cfg.SessionSecret, tok.Address, now, h.cfg.SessionLifetime)
}

// signOut caps and reads its body like every auth endpoint, then
// clears the session cookie. See [Handlers.SignOut] for what it does
// not do.
func (h *handlers) signOut(w http.ResponseWriter, r *http.Request) {
	log := h.cfg.Logger.With("client_ip", h.cfg.ClientIP(r))
	if !h.readForm(w, r, log) {
		return
	}
	http.SetCookie(w, ClearSessionCookie())
	log.Info("auth signed out")
	writeAnswer(w, http.StatusOK, answer{Outcome: outcomeSignedOut})
}

// readForm caps r's body at the configured limit and parses it,
// answering the request itself and reporting false when either step
// refuses. The cap is installed before the parser runs, so an oversized
// body is refused having been read no further than the cap.
func (h *handlers) readForm(w http.ResponseWriter, r *http.Request, log *slog.Logger) bool {
	r.Body = http.MaxBytesReader(w, r.Body, h.cfg.BodyCap)
	err := r.ParseForm()
	if err == nil {
		return true
	}
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		log.Info("auth request refused", "reason", "body exceeds cap")
		writeAnswer(w, http.StatusRequestEntityTooLarge, answer{Outcome: outcomeTooLarge})
		return false
	}
	log.Info("auth request refused", "reason", "body does not parse")
	writeAnswer(w, http.StatusBadRequest, answer{Outcome: outcomeMalformed})
	return false
}

// writeAnswer writes a as the response with status. Every auth
// response is no-store: one carries a session cookie, and none is a
// document a cache should replay.
func writeAnswer(w http.ResponseWriter, status int, a answer) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	// A failed write means the client has gone, and there is no one
	// left to answer.
	_ = json.NewEncoder(w).Encode(a)
}
