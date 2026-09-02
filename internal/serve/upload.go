// Package serve implements dubplate's upload engine: a tusd-backed
// resumable-upload store, the multi-file session bookkeeping that groups
// several tus uploads into one album, and the atomic promotion of a
// completed, passing album into the inbox the importer watches.
//
// Every request this package handles has already passed Cloudflare
// Access at the reverse-proxy layer; [WithEmail] and [EmailFromContext]
// are how the verified identity crosses from that outer enforcement
// (owned by cmd/dubplate/serve.go) into this package's ownership checks.
// A caller that constructs an [Engine] directly, as every test here
// does, must set that context value itself on every request.
package serve

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tus/tusd/v2/pkg/filestore"
	thandler "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/memorylocker"

	"github.com/glw907/dubplate/internal/archive"
	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/gate"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/journey"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/reject"
)

// cdMinSampleRate and cdMinBitsPerSample are the CD-quality floor
// [intake.CheckFLAC] enforces on every uploaded FLAC, matching
// internal/importer's own copy of the same two constants: each package
// that builds a [gate.Policy] owns its own literal floor rather than
// importing a shared one, the pattern internal/checks/uploader.go's
// probeAccess doc comment already names for this codebase.
const (
	cdMinSampleRate    = 44100
	cdMinBitsPerSample = 16
)

// bytesPerGiB converts a [config.Thresholds.FreeFloorGB] value to bytes.
const bytesPerGiB = 1 << 30

// defaultPrivateCollection is the "Private Recordings" collection name
// the custody-policy decision documents as the private lane's default:
// an operator who never sets [Options.PrivateCollection] still gets a
// working, config-named private lane rather than a silently absent one.
// It must also appear in [config.Config.Collections] for
// [validCollection] to admit a session created against it.
const defaultPrivateCollection = "Private Recordings"

// filesBasePath is where the tus protocol is mounted on Engine's own
// mux, agreed with U9's dev proxy: the pattern [Engine.Handler]
// registers and the prefix [http.StripPrefix] removes before the
// request reaches tusd. It is relative to wherever Engine.Handler is
// itself mounted; [Options.TusBasePath] is the absolute counterpart,
// the path tusd advertises to clients.
const filesBasePath = "/files/"

// contextKey namespaces this package's context values.
type contextKey int

const emailContextKey contextKey = iota

// WithEmail returns a copy of ctx carrying email as the verified
// requester identity, as Cloudflare Access confirmed it. The serve
// command's enforcement middleware sets this once per request; every
// ownership check in this package reads it back via [EmailFromContext].
func WithEmail(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, emailContextKey, email)
}

// EmailFromContext reports the verified requester email [WithEmail] set
// on ctx, and whether one was present.
func EmailFromContext(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(emailContextKey).(string)
	return email, ok
}

// TagProbeResult is what a tag-check probe reports about a staged,
// format-passing session, per the custody-policy decision: a beets
// matcher dry run against the session's assembled directory.
// [Engine.completeSession] promotes only when Confident is true; Artist,
// Album, and Year carry the matched metadata for the journey, and Detail
// carries the probe's best guess when it is not.
type TagProbeResult struct {
	Confident bool
	Artist    string
	Album     string
	Year      string
	Detail    string
}

// ErrTagProbeUnavailable is the error the zero-value [Options.TagProbe]
// always returns: an Engine built without a probe holds every main-lane
// session unpromoted rather than promoting an unconfirmed match -- the
// custody policy's fail-closed default, not a bug to work around with a
// permissive stub. It is a default for a test that does not care about
// the probe, never a production state: [NewServer] refuses to build a
// Server at all without a real [NewServerDeps.TagProbe].
var ErrTagProbeUnavailable = errors.New("serve: tag-check probe not configured")

// TagProbeFunc runs the custody-policy tag-check probe against a staged
// session's assembled directory: a dry beets matcher run
// ([github.com/glw907/dubplate/internal/beets.Runner.Probe]) that
// reports whether MusicBrainz confidently matched the album. ctx bounds
// the probe's lookups. A non-nil error means the probe itself could not
// run at all (MusicBrainz unreachable, beet missing) -- distinct from a
// clean not-confident result, since [Engine.completeSession] holds a
// probe-unavailable session with an honest retry-later state rather than
// treating it the same as a considered rejection.
type TagProbeFunc func(ctx context.Context, dir string) (TagProbeResult, error)

// unavailableTagProbe is the zero-value [Options.TagProbe]: see
// [ErrTagProbeUnavailable].
func unavailableTagProbe(context.Context, string) (TagProbeResult, error) {
	return TagProbeResult{}, ErrTagProbeUnavailable
}

// tagProbeTimeout bounds the tag-check probe [Engine.completeSession]
// runs: the contributor is watching the Matching step live, so three
// minutes is the ceiling before completeSession gives up on
// MusicBrainz and falls back to the honest probe-unavailable state
// rather than leaving the request hanging.
const tagProbeTimeout = 3 * time.Minute

// Engine is dubplate's upload engine: the tus protocol handler, the
// in-flight session store, and the promotion logic that turns a
// completed, passing session into an inbox directory the importer picks
// up. The zero Engine is not usable; construct with [NewEngine].
type Engine struct {
	cfg    *config.Config
	roster config.Roster

	sessions *Store
	store    thandler.DataStore
	tus      http.Handler

	checkPolicy gate.Policy

	now               func() time.Time
	rename            func(oldpath, newpath string) error
	onPromoted        func(sessionID, dest string)
	tagProbe          TagProbeFunc
	privateCollection string
	ledger            *ledger.Ledger
}

// Options carries Engine's construction seams: the test seams (a clock,
// a rename, a promotion hook, a tag probe) and the mount wiring
// [NewServer] sets in production. Each field's own doc names what the
// zero value falls back to.
type Options struct {
	// Now stands in for time.Now: session creation timestamps and the
	// sweep's cutoff both read the clock through this seam.
	Now func() time.Time
	// Rename stands in for os.Rename: promotion's one atomic move reads
	// it through this seam so a test can force a promotion failure
	// without needing a real cross-device filesystem.
	Rename func(oldpath, newpath string) error
	// OnPromoted, when set, is called with a session's ID and its final
	// inbox destination immediately after that session's [Engine.promote]
	// rename has already returned successfully. cmd/dubplate/serve.go
	// wires this to write the nudge trigger file; the zero value is a
	// no-op, which every test in this package other than the nudge's own
	// wants.
	OnPromoted func(sessionID, dest string)
	// TagProbe runs the custody-policy tag-check probe (the dry beets
	// matcher run) that every main-lane session must pass before
	// promotion. The zero value is [unavailableTagProbe], which always
	// fails closed with [ErrTagProbeUnavailable] -- correct for a test
	// exercising some other path, never for production, where
	// [NewServer] requires a real probe before it will build a Server at
	// all.
	TagProbe TagProbeFunc
	// PrivateCollection names the private lane's collection (the
	// custody-policy decision's "Private Recordings" lane): a session
	// created against it skips [Options.TagProbe] entirely. The zero
	// value defaults to [defaultPrivateCollection], the documented
	// default name.
	PrivateCollection string
	// TusBasePath sets tusd's own advertised BasePath
	// (github.com/tus/tusd/v2/pkg/handler.Config.BasePath), the prefix
	// of the absolute Location URL a tus client PATCHes back against.
	// tusd bakes that string in verbatim rather than inferring it from
	// the request, so it must name where [Engine.Handler]'s
	// [filesBasePath] route is reachable from outside this process,
	// mount prefix included. The zero value is [filesBasePath] itself,
	// correct only for an Engine.Handler mounted at the process root
	// (every test in this package). [NewServer] mounts it under
	// [basePath] instead and sets this to basePath+filesBasePath.
	TusBasePath string
	// Ledger is the same contributions ledger [JourneyDeps.Ledger] reads:
	// [Engine.SweepOnce] queries it to tell a session the importer has
	// already settled (a terminal event exists) from one it never will
	// (rejected, pushed back, or probe-unavailable -- none of which ever
	// reach the importer). The zero value is nil, which
	// [Engine.settledInLedger] treats as "never settled": every completed
	// session then waits out [config.UploadConfig.PartialTTL] before its
	// Store entry is forgotten, safe for a test that does not exercise
	// this path but wrong for production, where [NewServer] always wires
	// the real ledger.
	Ledger *ledger.Ledger
}

// NewEngine builds an Engine backed by a tus filestore rooted at
// cfg.Paths.UploadsDir. It performs no I/O beyond what
// [filestore.FileStore] itself does at construction (none): the store
// directory must already exist (doctor and serve's startup checks are
// the preconditions, not NewEngine).
func NewEngine(cfg *config.Config, roster config.Roster, opts Options) (*Engine, error) {
	if opts.Now == nil {
		opts.Now = time.Now
	}
	if opts.Rename == nil {
		opts.Rename = os.Rename
	}
	if opts.TagProbe == nil {
		opts.TagProbe = unavailableTagProbe
	}
	if opts.PrivateCollection == "" {
		opts.PrivateCollection = defaultPrivateCollection
	}
	if opts.TusBasePath == "" {
		opts.TusBasePath = filesBasePath
	}

	store := filestore.New(cfg.Paths.UploadsDir)
	locker := memorylocker.New()

	composer := thandler.NewStoreComposer()
	composer.UseCore(store)
	composer.UseTerminater(store)
	composer.UseLocker(locker)

	e := &Engine{
		cfg:      cfg,
		roster:   roster,
		sessions: newStore(),
		store:    store,
		checkPolicy: gate.Policy{
			MinSampleRate:    cdMinSampleRate,
			MinBitsPerSample: cdMinBitsPerSample,
		},
		now:               opts.Now,
		rename:            opts.Rename,
		onPromoted:        opts.OnPromoted,
		tagProbe:          opts.TagProbe,
		privateCollection: opts.PrivateCollection,
		ledger:            opts.Ledger,
	}

	tusCfg := thandler.Config{
		StoreComposer: composer,
		BasePath:      opts.TusBasePath,
		// MaxSize bounds one tus resource: the whole album for a lone-ZIP
		// session, or one member file for a multi-file session. It is not
		// an aggregate cap across a multi-file session's several uploads;
		// nothing in the tus protocol offers that hook, so the album-total
		// cap for that path is enforced by Decide's callers rejecting an
		// oversized member instead.
		MaxSize:              cfg.Upload.MaxAlbumBytes,
		DisableDownload:      true,
		DisableConcatenation: true,
		// serve sits behind cloudflared, which terminates TLS and forwards
		// plain HTTP: without this, tusd stamps http:// into the absolute
		// Location header it hands back, and a tus client's resume PATCH
		// then hits an http->https redirect mid-protocol.
		RespectForwardedHeaders:    true,
		PreUploadCreateCallback:    e.preCreate,
		PreFinishResponseCallback:  e.preFinish,
		PreUploadTerminateCallback: e.preTerminate,
	}
	h, err := thandler.NewHandler(tusCfg)
	if err != nil {
		return nil, fmt.Errorf("build tus handler: %w", err)
	}
	e.tus = h

	return e, nil
}

// PrivateCollection returns the private lane's collection name: either
// [Options.PrivateCollection] or the [defaultPrivateCollection] fallback
// NewEngine applied. GET /api/config serves it so no client has to
// re-derive that fallback from the collection list.
func (e *Engine) PrivateCollection() string {
	return e.privateCollection
}

// Handler returns the full mux Engine serves: the tus protocol under
// /files/, and the session JSON API under /api/session.
func (e *Engine) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/session", e.handleCreateSession)
	mux.HandleFunc("POST /api/session/{id}/finish", e.handleFinish)
	mux.Handle(filesBasePath, http.StripPrefix(strings.TrimSuffix(filesBasePath, "/"), e.filesHandler()))
	return mux
}

// filesHandler wraps the tus protocol handler with ownership enforcement
// and the read/write deadline clearing a large, slow upload needs.
func (e *Engine) filesHandler() http.Handler {
	return e.enforceUploadOwnership(clearBodyDeadlines(e.tus))
}

// clearBodyDeadlines clears the enclosing http.Server's ReadTimeout and
// WriteTimeout for every request under next, so a large, slow PATCH is
// never cut mid-transfer by a deadline sized for the API's short JSON
// routes. tusd owns the tus routes' own transfer deadline
// (handler.Config.NetworkTimeout); the server's ReadHeaderTimeout still
// applies here too, since it is set before a handler ever runs and so
// cannot be cleared from inside one -- it bounds only how long a client
// may take to start sending a request, never the body that follows.
func clearBodyDeadlines(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		_ = rc.SetReadDeadline(time.Time{})
		_ = rc.SetWriteDeadline(time.Time{})
		next.ServeHTTP(w, r)
	})
}

// uploadIDPattern is the shape [Engine.preCreate] actually mints for a
// tus upload ID (sess.ID + "-" + slot): the only shape
// [Engine.enforceUploadOwnership] may hand to the store, since tusd's
// filestore performs no ID validation of its own and joins the ID
// straight onto its storage root with filepath.Join. A percent-encoded
// "../" survives Go's ServeMux (which canonicalizes only the escaped
// path) and StripPrefix (a plain string trim) to reach here still
// decoded in r.URL.Path, so this check -- not tusd's -- is what stops it
// from ever reaching a filesystem join.
var uploadIDPattern = regexp.MustCompile(`^[a-z2-7]{26}-[0-9]+$`)

// enforceUploadOwnership rejects any /files/ request against an upload
// whose stamped owner (set at pre-create) differs from the request's
// verified email, and any request from a caller with no verified
// identity at all. A denial and an unknown ID both answer 404: telling
// the two apart would hand a cross-user prober an existence oracle. The
// bare create path (POST to the base path, no ID yet) has no existing
// upload to own and passes through; its own ownership is established by
// [Engine.preCreate].
func (e *Engine) enforceUploadOwnership(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.Trim(r.URL.Path, "/")
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}
		if !uploadIDPattern.MatchString(id) {
			http.NotFound(w, r)
			return
		}

		email, ok := EmailFromContext(r.Context())
		if !ok {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		upload, err := e.store.GetUpload(r.Context(), id)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		info, err := upload.GetInfo(r.Context())
		if err != nil {
			http.NotFound(w, r)
			return
		}
		if info.MetaData["email"] != email {
			http.NotFound(w, r)
			return
		}

		if sess, found := e.sessions.get(info.MetaData["session_id"]); found {
			w.Header().Set("Upload-Expires", sess.expiresAt(e.cfg.Upload.PartialTTL).UTC().Format(http.TimeFormat))
		}

		next.ServeHTTP(w, r)
	})
}

// preCreate is tusd's PreUploadCreateCallback. It never sends a response
// body of its own, so the whole decision lives in [Engine.admitUpload]
// and this adapter only widens that verdict to the three values tusd's
// hook signature takes.
func (e *Engine) preCreate(hook thandler.HookEvent) (thandler.HTTPResponse, thandler.FileInfoChanges, error) {
	changes, err := e.admitUpload(hook)
	return thandler.HTTPResponse{}, changes, err
}

// admitUpload stamps the verified email into the upload's metadata,
// resolves or mints the upload's session, and enforces every pre-create
// acceptance rule -- session ownership, album and collection validity,
// the ExpectedFiles slot count, and free-space admission -- before tusd
// ever creates the upload resource. A non-nil error refuses the upload
// with that error's own status and body.
func (e *Engine) admitUpload(hook thandler.HookEvent) (thandler.FileInfoChanges, error) {
	email, ok := EmailFromContext(hook.Context)
	if !ok {
		return thandler.FileInfoChanges{}, thandler.NewError("dubplate-unauthenticated", "no verified identity", http.StatusUnauthorized)
	}
	if hook.Upload.SizeIsDeferred {
		return thandler.FileInfoChanges{}, thandler.NewError("dubplate-deferred-length", "deferred-length uploads are not supported", http.StatusBadRequest)
	}

	meta := hook.Upload.MetaData
	filename := sanitizeFilename(meta["filename"])

	sess, err := e.sessionForUpload(email, filename, meta)
	if err != nil {
		return thandler.FileInfoChanges{}, err
	}

	slot, err := sess.reserve(hook.Upload.Size, e.cfg.Upload.MaxAlbumBytes)
	switch {
	case errors.Is(err, errAlbumTooLarge):
		return thandler.FileInfoChanges{}, reasonError(reject.ReasonTooLarge, http.StatusRequestEntityTooLarge)
	case err != nil:
		return thandler.FileInfoChanges{}, thandler.NewError("dubplate-session-full", "session already has its expected number of files", http.StatusConflict)
	}

	avail, err := archive.FreeSpace(e.cfg.Paths.UploadsDir)
	if err != nil {
		return thandler.FileInfoChanges{}, fmt.Errorf("check free space: %w", err)
	}
	floor := int64(e.cfg.Thresholds.FreeFloorGB) * bytesPerGiB
	if avail-hook.Upload.Size < floor {
		return thandler.FileInfoChanges{}, reasonError(reject.ReasonNoSpace, http.StatusInsufficientStorage)
	}

	return thandler.FileInfoChanges{
		ID: sess.ID + "-" + strconv.Itoa(slot),
		MetaData: thandler.MetaData{
			"email":      email,
			"session_id": sess.ID,
			"filename":   filename,
		},
	}, nil
}

// sessionForUpload resolves the session a pre-create belongs to: the
// existing one meta names, or -- for the lone-ZIP path, which never
// calls POST /api/session -- a fresh implicit one minted here. An upload
// that is neither answers a 400, since it has no album to join.
func (e *Engine) sessionForUpload(email, filename string, meta map[string]string) (*Session, error) {
	if sessionID := meta["session_id"]; sessionID != "" {
		if !validSessionID(sessionID) {
			return nil, thandler.ErrNotFound
		}
		sess, ok := e.sessions.get(sessionID)
		if !ok || sess.Email != email {
			// Unknown and cross-user both answer 404: an existence oracle
			// would let a prober enumerate other contributors' sessions.
			return nil, thandler.ErrNotFound
		}
		return sess, nil
	}

	if !isZipName(filename) {
		return nil, thandler.NewError("dubplate-session-required", "a standalone upload must be a ZIP, or must name an existing session", http.StatusBadRequest)
	}

	album, err := slugifyAlbum(meta["album"])
	if err != nil {
		return nil, thandler.NewError("dubplate-bad-album", err.Error(), http.StatusBadRequest)
	}
	collection := meta["collection"]
	if !validCollection(collection, e.cfg.Collections) {
		return nil, thandler.NewError("dubplate-bad-collection", "unknown collection", http.StatusBadRequest)
	}
	if e.sessions.countActive(email) >= e.cfg.Upload.MaxSessionsPerUser {
		return nil, reasonError(reject.ReasonTooManySessions, http.StatusTooManyRequests)
	}
	sess, err := e.sessions.create(email, album, collection, 1, e.cfg.Collections, e.now())
	if err != nil {
		return nil, fmt.Errorf("create implicit session: %w", err)
	}
	return sess, nil
}

// preFinish is tusd's PreFinishResponseCallback: it fires once for every
// individual tus upload that reaches its declared size, whether that is
// the whole album (a lone ZIP) or one member of a multi-file session. It
// assembles the finished upload's content into the session directory,
// records its verdicts, and -- only once every reserved slot has
// reported in -- runs [Engine.completeSession] and folds the outcome
// into this upload's own response, so a lone ZIP's completion IS its
// album's verdict, delivered in the same request.
func (e *Engine) preFinish(hook thandler.HookEvent) (thandler.HTTPResponse, error) {
	sessionID := hook.Upload.MetaData["session_id"]
	sess, ok := e.sessions.get(sessionID)
	if !ok {
		// The Store is in-memory only, so a session a prior process still
		// tracked at PATCH time is gone here after a restart between that
		// PATCH and this finishing one: 410, not a bare 500, so a resuming
		// client can tell "start this album over" from "something broke."
		// True mid-restart resume needs sess reconstructed from persisted
		// metadata this package does not yet keep (F2.4's disk-orphan sweep,
		// [Engine.sweepOrphans], is the reclaim half of that finding; this is
		// the honest-failure half, not a fix for the resume gap itself).
		return thandler.HTTPResponse{}, thandler.NewError("dubplate-session-vanished",
			"this upload's session is no longer available; start the album over", http.StatusGone)
	}

	verdicts, err := e.assembleUpload(hook.Context, sess, hook.Upload)
	if err != nil {
		return thandler.HTTPResponse{}, fmt.Errorf("assemble upload: %w", err)
	}

	if !sess.uploadDone(verdicts) {
		return thandler.HTTPResponse{}, nil
	}

	return finishHTTPResponse(e.completeSession(sess)), nil
}

// preTerminate is tusd's PreUploadTerminateCallback: it denies a DELETE
// against any upload whose stamped owner differs from the caller's
// verified email, as 404 for the same existence-oracle reason
// [Engine.enforceUploadOwnership] answers PATCH and HEAD that way.
func (e *Engine) preTerminate(hook thandler.HookEvent) (thandler.HTTPResponse, error) {
	email, ok := EmailFromContext(hook.Context)
	if !ok || hook.Upload.MetaData["email"] != email {
		return thandler.HTTPResponse{}, thandler.ErrNotFound
	}
	return thandler.HTTPResponse{}, nil
}

// reasonError wraps reason as a [thandler.Error] whose body is reason's
// own {"code","friendly","next_step"} wire encoding -- the admission-
// control kind copy every reject.Reason-carrying rejection in this
// package sends instead of a bare 500.
func reasonError(reason reject.Reason, status int) error {
	body, err := json.Marshal(reason)
	if err != nil {
		body = []byte(`{}`)
	}
	return thandler.Error{
		ErrorCode: "dubplate-reject",
		Message:   reason.Friendly(),
		HTTPResponse: thandler.HTTPResponse{
			StatusCode: status,
			Body:       string(body),
			Header:     thandler.HTTPHeader{"Content-Type": "application/json"},
		},
	}
}

// isZipName reports whether name's extension marks it as the lone-ZIP
// implicit-session upload, case-insensitively.
func isZipName(name string) bool {
	return strings.HasSuffix(strings.ToLower(name), ".zip")
}

// sanitizeFilename reduces raw to a single safe path component: no
// directory traversal, no absolute path, no embedded NUL. It never
// returns an empty string, so a caller can always join the result under
// a directory without a second check.
func sanitizeFilename(raw string) string {
	name := strings.ReplaceAll(raw, `\`, "/")
	if i := strings.LastIndexByte(name, '/'); i >= 0 {
		name = name[i+1:]
	}
	if name == "" || name == "." || strings.ContainsRune(name, 0) {
		return "upload.bin"
	}
	if len(name) > 255 {
		name = name[:255]
	}
	return name
}

// assembleUpload turns one finished tus upload into on-disk album
// content under <UploadsDir>/<sess.ID>/, and returns the verdicts intake
// found. A ZIP is expanded via [archive.Extract] -- isolation and the
// same size caps [intake.CheckZIP] already screened at the header level
// -- and its raw blob discarded; a FLAC that is one member of a
// multi-file session is relocated into the session directory as itself,
// since there is nothing to unpack.
func (e *Engine) assembleUpload(ctx context.Context, sess *Session, info thandler.FileInfo) ([]intake.FileVerdict, error) {
	blobPath := info.Storage[filestore.StorageKeyPath]
	infoPath := info.Storage[filestore.StorageKeyInfoPath]
	filename := sanitizeFilename(info.MetaData["filename"])
	sessionDir := filepath.Join(e.cfg.Paths.UploadsDir, sess.ID)

	cleanupBlob := func() {
		_ = os.Remove(blobPath)
		_ = os.Remove(infoPath)
	}

	if isZipName(filename) {
		f, err := os.Open(blobPath)
		if err != nil {
			return nil, fmt.Errorf("open uploaded zip: %w", err)
		}
		verdicts, err := intake.CheckZIP(f, info.Size, e.checkPolicy)
		_ = f.Close()
		if err != nil {
			cleanupBlob()
			return nil, fmt.Errorf("check uploaded zip: %w", err)
		}

		if err := os.MkdirAll(sessionDir, 0o755); err != nil {
			return nil, fmt.Errorf("create session directory: %w", err)
		}
		limits := archive.Limits{
			MaxExtractedBytes: e.cfg.Upload.MaxAlbumBytes,
			MinFreeBytes:      int64(e.cfg.Thresholds.FreeFloorGB) * bytesPerGiB,
		}
		// archive.Extract dispatches on archivePath's own extension, but a
		// tus blob is stored under its bare upload ID with no extension at
		// all; give it one via a same-directory rename (cheap, same
		// filesystem) rather than teaching archive.Extract to take a
		// caller-supplied format hint.
		zipPath := blobPath + ".zip"
		if err := os.Rename(blobPath, zipPath); err != nil {
			return nil, fmt.Errorf("stage uploaded zip for extraction: %w", err)
		}
		if err := archive.Extract(ctx, zipPath, sessionDir, limits); err != nil {
			_ = os.RemoveAll(sessionDir)
			_ = os.Remove(zipPath)
			_ = os.Remove(infoPath)
			return nil, fmt.Errorf("extract uploaded zip: %w", err)
		}
		_ = os.Remove(zipPath)
		_ = os.Remove(infoPath)
		return verdicts, nil
	}

	f, err := os.Open(blobPath)
	if err != nil {
		return nil, fmt.Errorf("open uploaded file: %w", err)
	}
	verdict, err := intake.CheckFLAC(f, filename, e.checkPolicy)
	_ = f.Close()
	if err != nil {
		cleanupBlob()
		return nil, fmt.Errorf("check uploaded file: %w", err)
	}

	if err := os.MkdirAll(sessionDir, 0o755); err != nil {
		return nil, fmt.Errorf("create session directory: %w", err)
	}
	dest := filepath.Join(sessionDir, filename)
	if _, err := os.Stat(dest); err == nil {
		// Two uploads in one session landed the same sanitized filename;
		// disambiguate rather than silently clobbering the first one.
		dest = filepath.Join(sessionDir, info.ID+"-"+filename)
	}
	if err := os.Rename(blobPath, dest); err != nil {
		return nil, fmt.Errorf("place uploaded file: %w", err)
	}
	_ = os.Remove(infoPath)
	return []intake.FileVerdict{verdict}, nil
}

// completeSession runs a session to its terminal outcome exactly once,
// no matter how many callers race to trigger it (the count-reaching-
// ExpectedFiles path in [Engine.preFinish] and an explicit /finish call
// via [Engine.handleFinish]): the first caller through
// [Session.completeOnce] decides and promotes, holds, or discards; every
// other caller reads back the same outcome via [Session.waitOutcome].
//
// A session that fails its format verdict is discarded outright -- its
// audio itself was rejected, so nothing about a retry would change. A
// session that passes the format verdict but is a private-lane session
// (sess.Collection equals e.privateCollection) skips the tag-check probe
// entirely and promotes as-is. Every other passing session runs the
// custody-policy tag-check probe first: a confident match promotes; an
// unconfident or unavailable probe leaves the session staged and
// unpromoted rather than discarding it, since the audio passed and the
// contributor's fix path (re-upload, MusicBrainz edit, or a switch to
// the private lane) does not require re-sending it.
func (e *Engine) completeSession(sess *Session) sessionOutcome {
	if !sess.completeOnce() {
		out, err := sess.waitOutcome(context.Background())
		if err != nil {
			return sessionOutcome{}
		}
		return out
	}

	verdict := intake.Decide(sess.verdictsSnapshot())
	out := sessionOutcome{Verdict: verdict}
	if !verdict.OK {
		e.discard(sess)
		sess.setOutcome(out, e.now())
		return out
	}

	if sess.Collection == e.privateCollection {
		e.promoteInto(sess, &out, TagProbeResult{})
		sess.setOutcome(out, e.now())
		return out
	}

	ctx, cancel := context.WithTimeout(context.Background(), tagProbeTimeout)
	defer cancel()
	result, err := e.tagProbe(ctx, filepath.Join(e.cfg.Paths.UploadsDir, sess.ID))
	switch {
	case err != nil:
		out.ProbeUnavailable = true
		slog.Error("tag-check probe unavailable", "session_id", sess.ID, "err", err)
	case !result.Confident:
		out.PushedBack = true
		out.Detail = result.Detail
	default:
		e.promoteInto(sess, &out, result)
	}
	sess.setOutcome(out, e.now())
	return out
}

// promoteInto runs [Engine.promote] and folds its result into out,
// carrying match's metadata forward when the promotion succeeds. A
// mechanical promotion failure (a bad rename, a permissions fault) is
// not a content rejection: the staged directory is left in place for a
// retry or manual recovery rather than discarded, since throwing away a
// fully validated, tag-confirmed upload because promotion's plumbing
// hiccuped would be a worse outcome than leaving it staged.
func (e *Engine) promoteInto(sess *Session, out *sessionOutcome, match TagProbeResult) {
	dest, err := e.promote(sess)
	if err != nil {
		slog.Error("promote session", "session_id", sess.ID, "err", err)
		return
	}
	out.Promoted = true
	out.Dest = dest
	out.MatchedArtist = match.Artist
	out.MatchedAlbum = match.Album
	out.MatchedYear = match.Year
	if e.onPromoted != nil {
		e.onPromoted(sess.ID, dest)
	}
}

// sessionSidecar is session.json's on-disk shape: the collection's
// transport into the importer, and the session key's durable home,
// written into the session directory before the one [os.Rename] that
// promotes it into the inbox.
type sessionSidecar struct {
	Version    int       `json:"version"`
	SessionID  string    `json:"session_id"`
	Email      string    `json:"email"`
	DirName    string    `json:"dir_name"`
	Album      string    `json:"album"`
	Collection string    `json:"collection"`
	CreatedAt  time.Time `json:"created_at"`
}

// sessionSidecarVersion is [sessionSidecar]'s schema version.
const sessionSidecarVersion = 1

// promote writes session.json into sess's session directory and moves
// that directory into the inbox with one [Engine.rename] call. The
// rename, not a lock, is the ordering guarantee: session.json exists
// inside the directory before the single rename call ever runs, so the
// importer -- which only ever looks under InboxDir -- can never observe
// a directory that lacks its sidecar, and any failure between writing
// session.json and the rename returning leaves the whole directory
// exactly where it started, under UploadsDir, never partially under the
// inbox. UploadsDir and InboxDir sharing a filesystem (config.Paths'
// doc, checked at startup and by doctor) is what makes this one rename
// atomic rather than a silent cross-device copy-and-delete.
func (e *Engine) promote(sess *Session) (string, error) {
	sessionDir := filepath.Join(e.cfg.Paths.UploadsDir, sess.ID)

	dirName, ok := e.roster.DirNameFor(sess.Email)
	if !ok {
		return "", fmt.Errorf("serve: %s has no roster entry", sess.Email)
	}

	sidecar := sessionSidecar{
		Version:    sessionSidecarVersion,
		SessionID:  sess.ID,
		Email:      sess.Email,
		DirName:    dirName,
		Album:      sess.Album,
		Collection: sess.Collection,
		CreatedAt:  sess.createdAt,
	}
	data, err := json.MarshalIndent(sidecar, "", "  ")
	if err != nil {
		return "", fmt.Errorf("encode session sidecar: %w", err)
	}
	if err := atomicfile.WriteFile(filepath.Join(sessionDir, "session.json"), data, 0o644); err != nil {
		return "", fmt.Errorf("write session sidecar: %w", err)
	}

	destParent := filepath.Join(e.cfg.Paths.InboxDir, dirName)
	if err := os.MkdirAll(destParent, 0o755); err != nil {
		return "", fmt.Errorf("create inbox contributor directory: %w", err)
	}
	dest := filepath.Join(destParent, sess.ID+"__"+sess.Album)

	if err := e.rename(sessionDir, dest); err != nil {
		return "", fmt.Errorf("promote session %s: %w", sess.ID, err)
	}
	return dest, nil
}

// discard removes a rejected session's staged directory: it never
// reaches promote, so it must never linger under UploadsDir forever.
func (e *Engine) discard(sess *Session) {
	dir := filepath.Join(e.cfg.Paths.UploadsDir, sess.ID)
	if err := os.RemoveAll(dir); err != nil {
		slog.Error("discard rejected session", "session_id", sess.ID, "err", err)
	}
}

// settleGrace bounds how long a completed session's [Store] entry can be
// forgotten early, past its usual [config.UploadConfig.PartialTTL] wait:
// long enough for the contributor's 5s journey poll (api.go) to read the
// final state back at least once, once [Engine.settledInLedger] confirms
// the importer has already recorded that same final state durably. A
// session settledInLedger will never confirm -- rejected, pushed back,
// or probe-unavailable, none of which ever reach the importer -- gets no
// early reprieve: api.go:284's "a poll that lands after the tab reloads"
// promise depends on the live Store, the only place those three outcomes
// are ever recorded, surviving until [SweepOnce]'s PartialTTL fallback
// instead. Reclaiming any leftover on-disk directory a completed session
// left behind is [sweepOrphans]' separate, much longer job either way.
const settleGrace = 2 * time.Minute

// settledInLedger reports whether sessionID already carries a terminal
// ledger event (see [journey.Terminal]): the importer's own durable
// settlement record, once it exists, is what lets [SweepOnce] forget the
// session's live [Store] entry after [settleGrace] instead of waiting
// out the full PartialTTL fallback. A nil ledger (no test in this
// package wires one unless it targets this path) reports false, the safe
// default -- [SweepOnce] then only ever forgets the session through the
// PartialTTL fallback, same as before this method existed.
func (e *Engine) settledInLedger(sessionID string) bool {
	if e.ledger == nil {
		return false
	}
	events, err := e.ledger.Session(sessionID)
	if err != nil {
		slog.Error("sweep: read ledger for settled session", "session_id", sessionID, "err", err)
		return false
	}
	return len(events) > 0 && journey.Terminal(events[0].State)
}

// SweepOnce reclaims every session past its retention window, in two
// passes. First, the live [Store]: an abandoned, never-completed session
// older than [config.UploadConfig.PartialTTL] is forgotten entirely; a
// completed session past [settleGrace] is forgotten early only once
// [Engine.settledInLedger] confirms the importer has durably recorded its
// outcome, and every completed session -- settled in the ledger or not --
// is forgotten unconditionally once it too clears PartialTTL, the same
// backstop an abandoned session gets (its own disk content, if any
// survives it, is the second pass's job, not this one's). Second,
// [Engine.sweepOrphans] scans UploadsDir directly rather than the map, so
// it also reclaims what no map entry could ever point it to: a completed
// session's leftover directory once it clears PartialTTL, and -- the only
// version of this sweep that is correct across a restart -- anything a
// prior process's Store forgot the moment it exited. It returns how many
// sessions the first, map-driven pass swept for abandonment.
func (e *Engine) SweepOnce(now time.Time) int {
	cutoff := now.Add(-e.cfg.Upload.PartialTTL)

	targets := e.sessions.expired(cutoff)
	for _, sess := range targets {
		e.sessions.delete(sess.ID)
	}
	for _, sess := range e.sessions.settled(now.Add(-settleGrace)) {
		if e.settledInLedger(sess.ID) {
			e.sessions.delete(sess.ID)
		}
	}
	for _, sess := range e.sessions.settled(cutoff) {
		e.sessions.delete(sess.ID)
	}

	e.sweepOrphans(cutoff)
	return len(targets)
}

// sweepOrphans removes every UploadsDir entry older than cutoff that the
// live session map no longer claims: a session directory, a raw
// still-in-flight tus blob, or that blob's ".info" sidecar (the shapes
// [orphanSessionID] recognizes, all of them named for the session that
// produced them). Driving this from the filesystem rather than the map
// is what makes it correct across a restart -- the map is empty right
// after one, so every surviving entry looks unclaimed here regardless of
// why, and cutoff (not process memory) is what protects anything still
// in flight.
func (e *Engine) sweepOrphans(cutoff time.Time) {
	entries, err := os.ReadDir(e.cfg.Paths.UploadsDir)
	if err != nil {
		slog.Error("sweep read uploads dir", "err", err)
		return
	}
	for _, entry := range entries {
		id := orphanSessionID(entry.Name())
		if id == "" {
			continue
		}
		if _, live := e.sessions.get(id); live {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		path := filepath.Join(e.cfg.Paths.UploadsDir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			slog.Error("sweep orphaned upload path", "path", path, "err", err)
		}
	}
}

// orphanSessionID extracts the session ID a raw UploadsDir entry belongs
// to: the bare session directory itself, or a "<session-id>-<slot>" raw
// tus blob or its ".info" sidecar, the two shapes [Engine.preCreate]
// mints. Anything else -- session.json lives inside a session directory,
// never at UploadsDir's top level, so it never reaches this function --
// reports "".
func orphanSessionID(name string) string {
	base := strings.TrimSuffix(name, ".info")
	if validSessionID(base) {
		return base
	}
	if i := strings.LastIndexByte(base, '-'); i > 0 && validSessionID(base[:i]) {
		return base[:i]
	}
	return ""
}

// RunSweep calls [Engine.SweepOnce] every interval until ctx is done.
// The serve command starts this once at startup; production timing
// comes from the real clock regardless of any [Options.Now] override,
// since only SweepOnce's cutoff arithmetic needs to be controllable for
// a test.
func (e *Engine) RunSweep(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			e.SweepOnce(e.now())
		}
	}
}

// finishResponse is the JSON body both POST /api/session/<id>/finish and
// a lone ZIP's completing PATCH return: the same shape either way, so a
// client never special-cases which path completed its session. Promoted,
// PushedBack, and ProbeUnavailable are mutually exclusive; Detail and the
// Matched* fields are populated only alongside the outcome that sets
// them (see [sessionOutcome]).
type finishResponse struct {
	Promoted         bool                 `json:"promoted"`
	PushedBack       bool                 `json:"pushed_back,omitempty"`
	ProbeUnavailable bool                 `json:"probe_unavailable,omitempty"`
	Detail           string               `json:"detail,omitempty"`
	MatchedArtist    string               `json:"matched_artist,omitempty"`
	MatchedAlbum     string               `json:"matched_album,omitempty"`
	MatchedYear      string               `json:"matched_year,omitempty"`
	Files            []intake.FileVerdict `json:"files"`
}

// finishHTTPResponse renders out as the tus-hook HTTPResponse a
// completing PATCH returns.
func finishHTTPResponse(out sessionOutcome) thandler.HTTPResponse {
	body, err := json.Marshal(finishResponse{
		Promoted:         out.Promoted,
		PushedBack:       out.PushedBack,
		ProbeUnavailable: out.ProbeUnavailable,
		Detail:           out.Detail,
		MatchedArtist:    out.MatchedArtist,
		MatchedAlbum:     out.MatchedAlbum,
		MatchedYear:      out.MatchedYear,
		Files:            out.Verdict.Files,
	})
	if err != nil {
		body = []byte(`{}`)
	}
	return thandler.HTTPResponse{
		StatusCode: http.StatusOK,
		Body:       string(body),
		Header:     thandler.HTTPHeader{"Content-Type": "application/json"},
	}
}

// maxCreateSessionBody bounds POST /api/session's request body: its
// three fields never approach this in practice.
const maxCreateSessionBody = 4096

type createSessionRequest struct {
	Album         string `json:"album"`
	Collection    string `json:"collection"`
	ExpectedFiles int    `json:"expected_files"`
}

type createSessionResponse struct {
	SessionID string `json:"session_id"`
}

// handleCreateSession serves POST /api/session: it registers a new
// multi-file session for the request's verified email and returns its ID.
func (e *Engine) handleCreateSession(w http.ResponseWriter, r *http.Request) {
	email, ok := EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	var req createSessionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxCreateSessionBody)).Decode(&req); err != nil {
		http.Error(w, "bad request body", http.StatusBadRequest)
		return
	}

	if e.sessions.countActive(email) >= e.cfg.Upload.MaxSessionsPerUser {
		writeJSON(w, http.StatusTooManyRequests, reject.ReasonTooManySessions)
		return
	}

	sess, err := e.sessions.create(email, req.Album, req.Collection, req.ExpectedFiles, e.cfg.Collections, e.now())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, createSessionResponse{SessionID: sess.ID})
}

// handleFinish serves POST /api/session/{id}/finish: it forces a
// multi-file session to complete with whatever verdicts have
// accumulated so far, idempotently if the session already completed on
// its own by count.
func (e *Engine) handleFinish(w http.ResponseWriter, r *http.Request) {
	email, ok := EmailFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthenticated", http.StatusUnauthorized)
		return
	}

	id := r.PathValue("id")
	if !validSessionID(id) {
		http.NotFound(w, r)
		return
	}
	sess, ok := e.sessions.get(id)
	if !ok || sess.Email != email {
		http.NotFound(w, r)
		return
	}
	if sess.doneCount() == 0 {
		http.Error(w, "no files uploaded yet", http.StatusBadRequest)
		return
	}

	resp := finishHTTPResponse(e.completeSession(sess))
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write([]byte(resp.Body))
}

// writeJSON encodes v as status's JSON body.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
