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
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
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

// filesBasePath is where the tus protocol is mounted, agreed with U7's
// routing and U9's dev proxy.
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

	now        func() time.Time
	rename     func(oldpath, newpath string) error
	onPromoted func(sessionID, dest string)
}

// Options overrides Engine's test seams; every field defaults to its
// production behavior when left zero.
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
		now:        opts.Now,
		rename:     opts.Rename,
		onPromoted: opts.OnPromoted,
	}

	tusCfg := thandler.Config{
		StoreComposer: composer,
		BasePath:      filesBasePath,
		// MaxSize bounds one tus resource: the whole album for a lone-ZIP
		// session, or one member file for a multi-file session. It is not
		// an aggregate cap across a multi-file session's several uploads;
		// nothing in the tus protocol offers that hook, so the album-total
		// cap for that path is enforced by Decide's callers rejecting an
		// oversized member instead.
		MaxSize:                    cfg.Upload.MaxAlbumBytes,
		DisableDownload:            true,
		DisableConcatenation:       true,
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

// preCreate is tusd's PreUploadCreateCallback: it stamps the verified
// email into the upload's metadata, resolves or mints the upload's
// session, and enforces every pre-create acceptance rule -- session
// ownership, album and collection validity, the ExpectedFiles slot
// count, and free-space admission -- before tusd ever creates the
// upload resource.
func (e *Engine) preCreate(hook thandler.HookEvent) (thandler.HTTPResponse, thandler.FileInfoChanges, error) {
	email, ok := EmailFromContext(hook.Context)
	if !ok {
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-unauthenticated", "no verified identity", http.StatusUnauthorized)
	}
	if hook.Upload.SizeIsDeferred {
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-deferred-length", "deferred-length uploads are not supported", http.StatusBadRequest)
	}

	meta := hook.Upload.MetaData
	filename := sanitizeFilename(meta["filename"])
	sessionID := meta["session_id"]

	var sess *Session
	switch {
	case sessionID != "":
		if !validSessionID(sessionID) {
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.ErrNotFound
		}
		found, ok := e.sessions.get(sessionID)
		if !ok || found.Email != email {
			// Unknown and cross-user both answer 404: an existence oracle
			// would let a prober enumerate other contributors' sessions.
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.ErrNotFound
		}
		sess = found

	case isZipName(filename):
		album, err := slugifyAlbum(meta["album"])
		if err != nil {
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-bad-album", err.Error(), http.StatusBadRequest)
		}
		collection := meta["collection"]
		if !validCollection(collection, e.cfg.Collections) {
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-bad-collection", "unknown collection", http.StatusBadRequest)
		}
		if e.sessions.countActive(email) >= e.cfg.Upload.MaxSessionsPerUser {
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, reasonError(reject.ReasonTooManySessions, http.StatusTooManyRequests)
		}
		created, err := e.sessions.create(email, album, collection, 1, e.cfg.Collections, e.now())
		if err != nil {
			return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, fmt.Errorf("create implicit session: %w", err)
		}
		sess = created

	default:
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-session-required", "a standalone upload must be a ZIP, or must name an existing session", http.StatusBadRequest)
	}

	slot, ok := sess.reserve()
	if !ok {
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, thandler.NewError("dubplate-session-full", "session already has its expected number of files", http.StatusConflict)
	}

	avail, err := archive.FreeSpace(e.cfg.Paths.UploadsDir)
	if err != nil {
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, fmt.Errorf("check free space: %w", err)
	}
	floor := int64(e.cfg.Thresholds.FreeFloorGB) * bytesPerGiB
	if avail-hook.Upload.Size < floor {
		return thandler.HTTPResponse{}, thandler.FileInfoChanges{}, reasonError(reject.ReasonNoSpace, http.StatusInsufficientStorage)
	}

	changes := thandler.FileInfoChanges{
		ID: sess.ID + "-" + strconv.Itoa(slot),
		MetaData: thandler.MetaData{
			"email":      email,
			"session_id": sess.ID,
			"filename":   filename,
		},
	}
	return thandler.HTTPResponse{}, changes, nil
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
		return thandler.HTTPResponse{}, fmt.Errorf("serve: session %s vanished before finish", sessionID)
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
// [Session.completeOnce] decides and promotes or discards; every other
// caller reads back the same outcome via [Session.waitOutcome].
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
	switch {
	case verdict.OK:
		dest, err := e.promote(sess)
		if err != nil {
			// A mechanical promotion failure (a bad rename, a permissions
			// fault) is not a content rejection: the staged directory is
			// left in place for a retry or manual recovery rather than
			// discarded, since throwing away a fully validated upload
			// because promotion's plumbing hiccuped would be a worse
			// outcome than leaving it staged.
			slog.Error("promote session", "session_id", sess.ID, "err", err)
		} else {
			out.Promoted = true
			out.Dest = dest
			if e.onPromoted != nil {
				e.onPromoted(sess.ID, dest)
			}
		}
	default:
		e.discard(sess)
	}
	sess.setOutcome(out)
	return out
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

// SweepOnce removes every session created before now minus
// [config.UploadConfig.PartialTTL] that never completed: its assembled
// session directory (if any files had already landed) and every raw,
// still-in-flight tus blob its reserved slots produced (named
// "<session-id>-<slot>" by [Engine.preCreate], which is what makes a
// glob by that prefix find them). It returns how many sessions it swept.
func (e *Engine) SweepOnce(now time.Time) int {
	cutoff := now.Add(-e.cfg.Upload.PartialTTL)
	targets := e.sessions.expired(cutoff)
	for _, sess := range targets {
		e.sweepOne(sess)
		e.sessions.delete(sess.ID)
	}
	return len(targets)
}

func (e *Engine) sweepOne(sess *Session) {
	dir := filepath.Join(e.cfg.Paths.UploadsDir, sess.ID)
	if err := os.RemoveAll(dir); err != nil {
		slog.Error("sweep session directory", "session_id", sess.ID, "err", err)
	}

	matches, err := filepath.Glob(filepath.Join(e.cfg.Paths.UploadsDir, sess.ID+"-*"))
	if err != nil {
		slog.Error("sweep glob stray uploads", "session_id", sess.ID, "err", err)
		return
	}
	for _, m := range matches {
		if err := os.RemoveAll(m); err != nil {
			slog.Error("sweep stray upload", "path", m, "err", err)
		}
	}
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
// client never special-cases which path completed its session.
type finishResponse struct {
	Promoted bool                 `json:"promoted"`
	Files    []intake.FileVerdict `json:"files"`
}

// finishHTTPResponse renders out as the tus-hook HTTPResponse a
// completing PATCH returns.
func finishHTTPResponse(out sessionOutcome) thandler.HTTPResponse {
	body, err := json.Marshal(finishResponse{Promoted: out.Promoted, Files: out.Verdict.Files})
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
