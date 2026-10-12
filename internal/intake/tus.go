package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/tus/tusd/v2/pkg/filestore"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/memorylocker"
	expslog "golang.org/x/exp/slog"
)

// maxUploadBytes is DUBPLATE_INTAKE_MAX_ALBUM_BYTES's ceiling, which
// bounds tusd's MaxSize. The live cap is the wrong bound: tusd's
// configuration is built once, the cap applies at the next create, and
// a fixed live value would answer 413 to a file a raised cap admitted.
// The create callback holds each upload to the size its create
// declared, which the live cap bounded then.
const maxUploadBytes = 17179869184

// TusConfig is what the tus mount needs from the binary's wiring.
type TusConfig struct {
	// Prefix is the shell's owned prefix, with a leading separator and
	// no trailing one.
	Prefix string
	// PublicURL is DUBPLATE_PUBLIC_URL. Every Location the mount
	// answers is built from it, never from the request's Host.
	PublicURL string
	// Verify admits a request's member.
	Verify Verifier
	// Floor reads DUBPLATE_INTAKE_FREE_SPACE_FLOOR in force, in bytes.
	Floor func(ctx context.Context) (int64, error)
	// TusLog receives tusd's own warnings and errors as JSON: its
	// logger is golang.org/x/exp/slog's, a type log/slog cannot fill.
	TusLog io.Writer
}

// NewTusMount returns the handler for the shell's subtree tus, which
// takes uploads into staging's tus directory for the submissions the
// store holds. The shell hands it the full path, prefix included.
func NewTusMount(h *store.Handle, stagingDir string, log *slog.Logger, cfg TusConfig) (http.Handler, error) {
	return newSubmissions(h, log, stagingDir, time.Now, staging.FreeBytes).tusMount(cfg)
}

// tusMount authorizes every request before tusd sees it (tuswrap.go)
// and holds tusd's two callbacks, which check a creation against the
// store and run receipt when an upload finishes.
type tusMount struct {
	subs   *submissions
	base   string
	verify Verifier
	floor  func(ctx context.Context) (int64, error)
	files  filestore.FileStore
	tusd   http.Handler
}

// tusMount builds the mount over tusd configured closed: no CORS,
// downloads, termination, concatenation, deferred length, experimental
// protocol, or forwarded headers. The composer is built by hand,
// because FileStore.UseIn would also compose the concater, the length
// deferrer, and the content server.
func (s *submissions) tusMount(cfg TusConfig) (http.Handler, error) {
	files := filestore.New(filepath.Join(s.staging, "tus"))
	files.DirModePerm, files.FileModePerm = 0o750, 0o640
	composer := handler.NewStoreComposer()
	composer.UseCore(files)
	memorylocker.New().UseIn(composer)

	m := &tusMount{
		subs:   s,
		base:   cfg.Prefix + "/tus",
		verify: cfg.Verify,
		floor:  cfg.Floor,
		files:  files,
	}
	tusd, err := handler.NewHandler(handler.Config{
		StoreComposer:             composer,
		MaxSize:                   maxUploadBytes,
		BasePath:                  strings.TrimSuffix(cfg.PublicURL, "/") + m.base + "/",
		DisableDownload:           true,
		DisableTermination:        true,
		Cors:                      &handler.CorsConfig{Disable: true},
		Logger:                    expslog.New(expslog.NewJSONHandler(cfg.TusLog, &expslog.HandlerOptions{Level: expslog.LevelWarn})),
		PreUploadCreateCallback:   m.checkCreate,
		PreFinishResponseCallback: m.receive,
	})
	if err != nil {
		return nil, err
	}
	m.tusd = tusd
	return m, nil
}

// checkCreate admits a creation only for an unreceived file of the
// member's own uploading submission that has no upload yet, at exactly
// the size the manifest declared. It names the upload by the file id
// and stores none of the client's metadata, so no key the client sends
// reaches .info. Another member's file answers as a file that does not
// exist.
func (m *tusMount) checkCreate(hook handler.HookEvent) (handler.HTTPResponse, handler.FileInfoChanges, error) {
	ctx := hook.Context
	// The wrapper places the member on every request it hands tusd;
	// were it missing, the zero id would own no file.
	member, _ := ctx.Value(memberKey{}).(identity.UserID)
	submission, ok := parseID(hook.Upload.MetaData["submission"])
	file, okFile := parseID(hook.Upload.MetaData["file"])
	if !ok || !okFile {
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusNoFile
	}

	var size int64
	err := m.subs.store.QueryRow(ctx, `SELECT f.size FROM submission_files f JOIN submissions s ON s.id = f.submission_id
		WHERE f.id = ? AND f.submission_id = ? AND s.uploader = ? AND s.state = ? AND f.received_at IS NULL`,
		file, submission, member[:], stateUploading).Scan(&size)
	if errors.Is(err, sql.ErrNoRows) {
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusNoFile
	}
	if err != nil {
		m.subs.log.Error("tus creation could not read its file", "error", err)
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusCouldNotCheck
	}
	if hook.Upload.Size != size {
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusSizeDiffers
	}

	// The filestore creates with O_TRUNC, so a second creation would
	// empty the upload the first one filled.
	id := hex.EncodeToString(file)
	switch _, err := m.files.GetUpload(ctx, id); {
	case err == nil:
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusUploadExists
	case !errors.Is(err, handler.ErrNotFound):
		m.subs.log.Error("tus creation could not read its upload", "file", id, "error", err)
		return handler.HTTPResponse{}, handler.FileInfoChanges{}, errTusCouldNotCheck
	}
	return handler.HTTPResponse{}, handler.FileInfoChanges{ID: id, MetaData: handler.MetaData{}}, nil
}

// receive runs receipt for an upload that has just finished. Its id is
// a file id checkCreate minted.
func (m *tusMount) receive(hook handler.HookEvent) (handler.HTTPResponse, error) {
	file, _ := hex.DecodeString(hook.Upload.ID)
	switch err := m.subs.receive(hook.Context, file); {
	case errors.Is(err, errWrongState):
		return handler.HTTPResponse{}, errTusClosed
	case err != nil:
		m.subs.log.Error("tus receipt failed", "file", hook.Upload.ID, "error", err)
		return handler.HTTPResponse{}, errTusCouldNotCheck
	}
	return handler.HTTPResponse{}, nil
}
