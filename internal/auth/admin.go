package auth

import (
	"context"
	"errors"
	"log/slog"
	"mime"
	"net/http"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// rosterAPIWriteWait bounds how long a roster API write waits for the
// roster's directory lock. A server request's context ends only when
// the client goes or the handler returns, so without its own bound a
// write queued behind another process would hold the request open for
// as long as that process holds the lock.
const rosterAPIWriteWait = 5 * time.Second

// Form fields the roster API's POST reads. They are the whole of what a
// POST may carry; the admin flag is not among them.
const (
	displayNameField      = "display_name"
	streamerUsernameField = "streamer_username"
	adminField            = "admin"
)

// rosterFields is the closed set of keys a roster API POST may carry.
var rosterFields = []string{addressField, displayNameField, streamerUsernameField}

// Outcomes the roster API answers a POST with, beside the shared ones.
const (
	outcomeAdded                     = "added"
	outcomeDuplicateAddress          = "duplicate_address"
	outcomeDuplicateStreamerUsername = "duplicate_streamer_username"
	outcomeDuplicatePrincipal        = "duplicate_principal"
	outcomeInvalidDisplayName        = "invalid_display_name"
	outcomeBlankDisplayName          = "blank_display_name"
	outcomeInvalidStreamerUsername   = "invalid_streamer_username"
	outcomeMissingStreamerUsername   = "missing_streamer_username"
	outcomeAdminNotGrantable         = "admin_not_grantable"
	outcomeUnknownField              = "unknown_field"
	outcomeBusy                      = "busy"
	outcomeWriteUnconfirmed          = "write_unconfirmed"
)

// AdminConfig is everything the admin gate and the roster API need,
// resolved by the caller. Every field is required.
type AdminConfig struct {
	// Seam verifies the session a request carries. The admin flag the
	// gate checks is the one the seam read from the roster record.
	Seam Seam

	Roster *Roster

	// ClientIP returns the client address a request is held
	// accountable for, under the shell's trusted-proxy policy.
	ClientIP func(r *http.Request) string

	// Now returns the instant a request is evaluated at.
	Now func() time.Time

	Logger *slog.Logger

	// BodyCap is the byte limit on a POST body, applied before any
	// parser reads it.
	BodyCap int64
}

// NewRosterAPI returns the roster API behind the admin gate. A request
// that is not from a verified admin answers exactly what
// [http.NotFound] answers, so it learns nothing about whether the route
// exists. A verified admin may GET or HEAD the member list and POST one
// new member as a urlencoded form of address, display_name, and
// streamer_username. Those answers are JSON carrying an "outcome" or a
// "members" field. The API answers every path it is mounted on, so the
// mount confines it to the one path it serves.
func NewRosterAPI(cfg AdminConfig) http.Handler {
	return RequireAdmin(cfg, &rosterAPI{cfg: cfg, writeWait: rosterAPIWriteWait})
}

// adminKey is the context key under which the gate passes the
// verified admin to the handler it wraps.
type adminKey struct{}

// verifiedAdmin returns the identity the gate placed in ctx, reporting
// false unless it is a valid identity holding the admin flag.
func verifiedAdmin(ctx context.Context) (identity.Identity, bool) {
	v, ok := ctx.Value(adminKey{}).(VerifiedIdentity)
	if !ok || !v.Identity.Valid() || !v.Admin {
		return identity.Identity{}, false
	}
	return v.Identity, true
}

// RequireAdmin admits a request to next only when its session verifies
// through cfg.Seam as a valid identity holding the admin flag. Every
// other request answers exactly what [http.NotFound] answers and
// carries no cookie, so a gated path tells a non-admin nothing an
// unregistered path would not. next decides every method and path
// check itself, and those answers reach verified admins only.
//
// The session is refreshed only when Sec-Fetch-Site is exactly
// same-origin: the shell's cross-origin check always admits a GET, so
// having been reached says nothing about where a GET came from. The
// seam refreshes a verified non-admin's session as well, so the
// refreshed cookie is set only once the admin check has passed; setting
// it earlier would put a Set-Cookie on a non-admin's 404 that an
// unregistered path never carries.
func RequireAdmin(cfg AdminConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := cfg.Now()
		sameOrigin := r.Header.Get("Sec-Fetch-Site") == "same-origin"
		v, err := cfg.Seam.VerifyIdentity(r, now, sameOrigin)
		var reason string
		switch {
		case err != nil:
			reason = gateRefusalReason(err)
		case !v.Identity.Valid():
			reason = "identity is not a verified login"
		case !v.Admin:
			reason = "not an admin"
		}
		if reason != "" {
			cfg.Logger.Info("admin gate refused", "reason", reason, "client_ip", cfg.ClientIP(r))
			http.NotFound(w, r)
			return
		}
		if v.RefreshedValue != "" {
			http.SetCookie(w, SessionCookie(v.RefreshedValue, now, v.Session.ExpiresAt))
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminKey{}, v)))
	})
}

// gateRefusals are the causes the gate's log names. A seam error's own
// text is never logged: the epoch check wraps the normalizer's error,
// which quotes the address a signed session carries.
var gateRefusals = []error{
	http.ErrNoCookie,
	ErrSessionMalformed,
	ErrSessionMAC,
	ErrSessionCapReached,
	ErrSessionExpired,
	ErrSessionRevoked,
	ErrMemberNotFound,
	ErrRosterMissing,
	ErrRosterUnreadable,
	ErrRosterUnparseable,
	ErrRosterRecordInvalid,
}

// gateRefusalReason names err by the first of gateRefusals it wraps.
func gateRefusalReason(err error) string {
	for _, known := range gateRefusals {
		if errors.Is(err, known) {
			return known.Error()
		}
	}
	return "identity did not verify"
}

// rosterAPI is the handler the gate wraps. writeWait bounds a POST's
// wait for the roster's directory lock.
type rosterAPI struct {
	cfg       AdminConfig
	writeWait time.Duration
}

// rosterListing is a GET's JSON body.
type rosterListing struct {
	Members []rosterEntry `json:"members"`
}

type rosterEntry struct {
	Address          string `json:"address"`
	DisplayName      string `json:"display_name"`
	StreamerUsername string `json:"streamer_username"`
	Admin            bool   `json:"admin"`
}

// ServeHTTP refuses on its own when the gate did not run, so a mount
// that forgot the gate still discloses nothing.
func (a *rosterAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	actor, ok := verifiedAdmin(r.Context())
	if !ok {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		a.list(w, r)
	case http.MethodPost:
		a.add(w, r, actor.Address)
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	}
}

// list answers the whole roster or, when the roster fails to load,
// failed alone, never a partial list.
func (a *rosterAPI) list(w http.ResponseWriter, r *http.Request) {
	members, err := a.cfg.Roster.List()
	if err != nil {
		a.cfg.Logger.Error("roster listing failed", "reason", err.Error(), "client_ip", a.cfg.ClientIP(r))
		writeAnswer(w, http.StatusInternalServerError, answer{Outcome: outcomeFailed})
		return
	}
	listing := rosterListing{Members: make([]rosterEntry, 0, len(members))}
	for _, m := range members {
		listing.Members = append(listing.Members, rosterEntry(m))
	}
	if r.Method == http.MethodHead {
		writeJSONHeader(w, http.StatusOK)
		return
	}
	writeJSON(w, http.StatusOK, listing)
}

// add runs a fixed order: the media type, the body cap while the form
// parses, then the form's keys, and only then the fields, which come
// from the body alone. Every refusal before AddMember writes nothing,
// and AddMember's own refusals change no byte on disk.
func (a *rosterAPI) add(w http.ResponseWriter, r *http.Request, actor string) {
	clientIP := a.cfg.ClientIP(r)
	log := a.cfg.Logger.With("client_ip", clientIP)
	refuse := func(status int, outcome string) {
		log.Info("roster addition refused", "outcome", outcome)
		writeAnswer(w, status, answer{Outcome: outcome})
	}

	// ParseForm reads a body only under this media type and otherwise
	// leaves PostForm empty, which would surface as three missing fields.
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/x-www-form-urlencoded" {
		refuse(http.StatusBadRequest, outcomeMalformed)
		return
	}
	if !readForm(w, r, a.cfg.BodyCap, log) {
		return
	}
	if outcome := formRefusal(r); outcome != "" {
		refuse(http.StatusBadRequest, outcome)
		return
	}

	candidate := Member{
		Address:          r.PostForm.Get(addressField),
		DisplayName:      r.PostForm.Get(displayNameField),
		StreamerUsername: r.PostForm.Get(streamerUsernameField),
	}
	ctx, cancel := context.WithTimeout(r.Context(), a.writeWait)
	defer cancel()
	err = a.cfg.Roster.AddMember(ctx, candidate, actor, identity.ChannelCookie, clientIP, a.cfg.Logger)
	status, outcome := additionOutcome(err)
	if status >= http.StatusInternalServerError {
		log.Error("roster addition failed", "outcome", outcome, "reason", err.Error())
		writeAnswer(w, status, answer{Outcome: outcome})
		return
	}
	if err != nil {
		refuse(status, outcome)
		return
	}
	writeAnswer(w, status, answer{Outcome: outcome})
}

// formRefusal returns the outcome refusing r's parsed form, or "" when
// the form carries each of rosterFields at most once, in the body, and
// nothing else. A repeated key is refused because Get would silently
// take the first. The admin key is refused in the body and the query
// alike, whatever its value. A field in the query is refused so an
// address never rides a URL into an access log.
func formRefusal(r *http.Request) string {
	for _, values := range r.Form {
		if len(values) > 1 {
			return outcomeMalformed
		}
	}
	if r.Form.Has(adminField) {
		return outcomeAdminNotGrantable
	}
	for key := range r.Form {
		if !slices.Contains(rosterFields, key) {
			return outcomeUnknownField
		}
	}
	if len(r.URL.Query()) > 0 {
		return outcomeMalformed
	}
	return ""
}

// additionOutcome maps AddMember's result to a status and an outcome.
// Every validation refusal also wraps ErrRosterRecordInvalid, as does a
// demoted actor, so only the specific sentinels decide a 4xx.
func additionOutcome(err error) (int, string) {
	switch {
	case err == nil:
		return http.StatusCreated, outcomeAdded
	case errors.Is(err, errDuplicateAddress):
		return http.StatusConflict, outcomeDuplicateAddress
	case errors.Is(err, errDuplicateStreamerUsername):
		return http.StatusConflict, outcomeDuplicateStreamerUsername
	case errors.Is(err, errDuplicatePrincipal):
		return http.StatusConflict, outcomeDuplicatePrincipal
	case errors.Is(err, errInvalidMemberAddress):
		return http.StatusBadRequest, outcomeInvalid
	case errors.Is(err, errInvalidDisplayName):
		return http.StatusBadRequest, outcomeInvalidDisplayName
	case errors.Is(err, errBlankDisplayName):
		return http.StatusBadRequest, outcomeBlankDisplayName
	case errors.Is(err, errInvalidStreamerUsername):
		return http.StatusBadRequest, outcomeInvalidStreamerUsername
	case errors.Is(err, errMissingStreamerUsername):
		return http.StatusBadRequest, outcomeMissingStreamerUsername
	case errors.Is(err, errRosterBusy):
		return http.StatusServiceUnavailable, outcomeBusy
	case errors.Is(err, errRosterWriteLanded):
		// The member may be on disk, so a retry could meet
		// duplicate_address; the caller is told to reload first.
		return http.StatusInternalServerError, outcomeWriteUnconfirmed
	default:
		return http.StatusInternalServerError, outcomeFailed
	}
}
