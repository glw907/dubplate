package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
)

// maxStoreURL is the longest store URL a submission keeps, in bytes.
const maxStoreURL = 2048

// Limits are the intake settings one create runs under.
type Limits struct {
	Policy            gate.Policy
	SessionsPerMember int
	// GroupQuotaBytes is the group's cap; zero sets none.
	GroupQuotaBytes int64
	FreeSpaceFloor  int64
}

func (l Limits) limits() limits {
	return limits{
		policy:            l.Policy,
		sessionsPerMember: l.SessionsPerMember,
		groupQuotaBytes:   l.GroupQuotaBytes,
		freeSpaceFloor:    l.FreeSpaceFloor,
	}
}

// SeamConfig is what [NewSeam] needs.
type SeamConfig struct {
	Store *store.Handle
	Log   *slog.Logger
	// Staging is the directory whose free space the floor guards.
	Staging string
	Now     func() time.Time
	// Limits reads the intake settings in force. It runs before each
	// create, so a changed setting reaches the next one.
	Limits func(ctx context.Context) (Limits, error)
}

// A Seam is the member-facing submission operations over the store. It
// implements [ops.SubmissionSeam], taking the caller only from the
// actor.
type Seam struct {
	subs   *submissions
	limits func(ctx context.Context) (Limits, error)
}

var _ ops.SubmissionSeam = (*Seam)(nil)

// NewSeam returns the seam over cfg.
func NewSeam(cfg SeamConfig) *Seam {
	return &Seam{
		subs:   newSubmissions(cfg.Store, cfg.Log, cfg.Staging, cfg.Now, staging.FreeBytes),
		limits: cfg.Limits,
	}
}

// CreateSubmission implements [ops.SubmissionSeam].
func (s *Seam) CreateSubmission(ctx context.Context, actor ops.Actor, req ops.CreateSubmissionRequest) (ops.CreateSubmissionResponse, error) {
	unit, container := gate.Unit(req.VerdictUnit), gate.Container(req.Container)
	switch {
	case unit != gate.UnitAlbum && unit != gate.UnitTrack:
		return ops.CreateSubmissionResponse{}, ops.NewInvalidRequestError(`verdict_unit must be "album" or "track"`)
	case container != gate.ContainerFiles && container != gate.ContainerZip:
		return ops.CreateSubmissionResponse{}, ops.NewInvalidRequestError(`container must be "files" or "zip"`)
	case !validStoreURL(req.StoreURL):
		return ops.CreateSubmissionResponse{}, ops.NewInvalidRequestError(
			fmt.Sprintf("store_url must be an absolute https URL with a host, at most %d bytes", maxStoreURL)).WithDetail(ops.DetailStoreURL)
	}
	lim, err := s.limits(ctx)
	if err != nil {
		return ops.CreateSubmissionResponse{}, fmt.Errorf("read intake limits: %w", err)
	}

	m := gate.Manifest{Unit: unit, Container: container, Source: req.Source}
	for _, f := range req.Files {
		m.Files = append(m.Files, gate.ManifestFile{Path: f.Path, Size: f.Size})
	}
	c, err := s.subs.create(ctx, actor.UserID(), newSubmission{
		manifest: m, sourceNote: req.SourceNote, storeURL: req.StoreURL,
	}, lim.limits())
	if err != nil {
		return ops.CreateSubmissionResponse{}, createError(err)
	}

	resp := ops.CreateSubmissionResponse{SubmissionID: wireID(c.id), Files: make([]ops.CreatedFile, 0, len(c.files))}
	for _, f := range c.files {
		resp.Files = append(resp.Files, ops.CreatedFile{Path: f.path, FileID: wireID(f.id)})
	}
	return resp, nil
}

// validStoreURL reports whether raw is empty, which stores nothing, or
// an absolute https URL with a host within [maxStoreURL] bytes.
func validStoreURL(raw string) bool {
	if raw == "" {
		return true
	}
	u, err := url.Parse(raw)
	return err == nil && len(raw) <= maxStoreURL && u.Scheme == "https" && u.Hostname() != ""
}

// createError maps what create returns to the wire: a client's malformed
// manifest, its refusals with their reasons, a busy store, or the error
// unchanged for a server fault.
func createError(err error) error {
	var manifest *manifestRefusal
	var limit *limitRefusal
	switch {
	case errors.Is(err, gate.ErrMalformedManifest):
		return ops.NewInvalidRequestError(err.Error()).WithDetail(ops.DetailManifest)
	case errors.Is(err, gate.ErrExtraInTracks):
		return ops.NewInvalidRequestError(err.Error()).WithDetail(ops.DetailExtra)
	case errors.As(err, &manifest):
		rejects := ops.Rejects{Submission: manifest.refusal.Submission, Files: make([]ops.FileReject, 0, len(manifest.refusal.Files))}
		if rejects.Submission == nil {
			rejects.Submission = []reject.Reason{}
		}
		for _, f := range manifest.refusal.Files {
			rejects.Files = append(rejects.Files, ops.FileReject{Path: validUTF8(string(f.Path)), Reason: f.Reason})
		}
		return ops.NewInvalidRequestError("the manifest was refused").WithRejects(rejects)
	case errors.As(err, &limit):
		reason := reject.Of(limit.code)
		rejects := ops.Rejects{Submission: []reject.Reason{reason}, Files: []ops.FileReject{}}
		if limit.code == reject.CouldNotCheck {
			return fmt.Errorf("%w: %w", ops.NewUnavailableError(reason.Friendly).WithRejects(rejects), limit.err)
		}
		return ops.NewConflictError(reason.Friendly).WithRejects(rejects)
	}
	return writeFault(err)
}

// writeFault maps a write the store could not take. A busy store is the
// caller's to retry; any other error is the server failing and answers
// internal.
func writeFault(err error) error {
	if errors.Is(err, store.ErrBusy) || errors.Is(err, context.DeadlineExceeded) {
		return ops.NewUnavailableError("the server is busy; try again").WithDetail(ops.DetailBusy)
	}
	return err
}

// ListSubmissions implements [ops.SubmissionSeam].
func (s *Seam) ListSubmissions(ctx context.Context, actor ops.Actor, _ ops.ListSubmissionsRequest) (ops.ListSubmissionsResponse, error) {
	uploader := actor.UserID()
	list, err := queryAll(ctx, s.subs.store, func(rows *sql.Rows) (ops.SubmissionSummary, error) {
		var id []byte
		var createdAt, updatedAt int64
		var sum ops.SubmissionSummary
		if err := rows.Scan(&id, &sum.State, &sum.VerdictUnit, &createdAt, &updatedAt); err != nil {
			return sum, err
		}
		sum.SubmissionID, sum.CreatedAt, sum.UpdatedAt = wireID(id), micros(createdAt), micros(updatedAt)
		return sum, nil
	}, `SELECT id, state, verdict_unit, created_at, updated_at FROM submissions
		WHERE uploader = ? ORDER BY created_at DESC, id DESC`, uploader[:])
	if err != nil {
		return ops.ListSubmissionsResponse{}, fmt.Errorf("list submissions: %w", err)
	}
	return ops.ListSubmissionsResponse{Submissions: list}, nil
}

// GetSubmission implements [ops.SubmissionSeam].
func (s *Seam) GetSubmission(ctx context.Context, actor ops.Actor, req ops.GetSubmissionRequest) (ops.GetSubmissionResponse, error) {
	id, err := parseWireID(req.SubmissionID)
	if err != nil {
		return ops.GetSubmissionResponse{}, err
	}
	return s.status(ctx, actor.UserID(), id)
}

// RetrySubmission implements [ops.SubmissionSeam].
func (s *Seam) RetrySubmission(ctx context.Context, actor ops.Actor, req ops.RetrySubmissionRequest) (ops.RetrySubmissionResponse, error) {
	status, err := s.act(ctx, actor.UserID(), req.SubmissionID, s.subs.retry)
	return ops.RetrySubmissionResponse(status), err
}

// CancelSubmission implements [ops.SubmissionSeam].
func (s *Seam) CancelSubmission(ctx context.Context, actor ops.Actor, req ops.CancelSubmissionRequest) (ops.CancelSubmissionResponse, error) {
	status, err := s.act(ctx, actor.UserID(), req.SubmissionID, s.subs.cancel)
	return ops.CancelSubmissionResponse(status), err
}

// act runs one of the uploader's actions on the submission wireID names
// and answers the submission as it stands afterwards.
func (s *Seam) act(ctx context.Context, uploader identity.UserID, wire string,
	action func(ctx context.Context, uploader identity.UserID, id []byte) error) (ops.GetSubmissionResponse, error) {
	id, err := parseWireID(wire)
	if err != nil {
		return ops.GetSubmissionResponse{}, err
	}
	if err := action(ctx, uploader, id); err != nil {
		return ops.GetSubmissionResponse{}, actionError(err)
	}
	return s.status(ctx, uploader, id)
}

// actionError maps a refused action to the wire. The submission is the
// caller's own whenever it is found, so naming that its state forbids
// the action leaks nothing.
func actionError(err error) error {
	switch {
	case errors.Is(err, errNotFound):
		return notFound()
	case errors.Is(err, errWrongState):
		return ops.NewConflictError("the submission is not in a state this action applies to")
	}
	return writeFault(err)
}

// notFound is the one answer for an unknown id and another member's.
func notFound() error { return ops.NewNotFoundError("no such submission") }

// parseWireID parses a submission id as the wire carries it.
func parseWireID(wire string) ([]byte, error) {
	id, ok := parseID(wire)
	if !ok {
		return nil, ops.NewInvalidRequestError("submission_id must be 32 lowercase hex characters")
	}
	return id, nil
}

// status reads submission id as uploader's. Another member's submission
// reads as one that does not exist.
func (s *Seam) status(ctx context.Context, uploader identity.UserID, id []byte) (ops.GetSubmissionResponse, error) {
	var out ops.GetSubmissionResponse
	var coll []byte
	var createdAt, updatedAt int64
	err := s.subs.store.QueryRow(ctx, `SELECT state, verdict_unit, container, source, collection_id, created_at, updated_at
		FROM submissions WHERE id = ? AND uploader = ?`, id, uploader[:]).
		Scan(&out.State, &out.VerdictUnit, &out.Container, &out.Source, &coll, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return out, notFound()
	}
	if err != nil {
		return out, fmt.Errorf("read submission: %w", err)
	}
	out.SubmissionID, out.CreatedAt, out.UpdatedAt = wireID(id), micros(createdAt), micros(updatedAt)

	reasons := &reasonReader{seam: s, collection: coll}
	out.Events, err = queryAll(ctx, s.subs.store, func(rows *sql.Rows) (ops.SubmissionEvent, error) {
		var at int64
		var reason sql.NullString
		var ev ops.SubmissionEvent
		if err := rows.Scan(&at, &ev.Kind, &reason); err != nil {
			return ev, err
		}
		ev.At, ev.Reason = micros(at), reason.String
		// An event's reason is a reject code or free text, so text that
		// names no code is no error here.
		if r, ok := reject.Lookup(reject.Code(reason.String)); ok {
			ev.Reject = &r
		}
		return ev, nil
	}, `SELECT at, kind, reason FROM submission_events WHERE submission_id = ? ORDER BY seq`, id)
	if err != nil {
		return out, fmt.Errorf("read submission events: %w", err)
	}

	files, err := queryAll(ctx, s.subs.store, func(rows *sql.Rows) (storedFile, error) {
		var f storedFile
		var fid, path []byte
		if err := rows.Scan(&fid, &path, &f.Size, &f.Received, &f.verdict); err != nil {
			return f, err
		}
		f.FileID, f.Path = wireID(fid), validUTF8(string(path))
		return f, nil
	}, `SELECT id, rel_path, size, received_at IS NOT NULL, verdict FROM submission_files
		WHERE submission_id = ? ORDER BY rel_path`, id)
	if err != nil {
		return out, fmt.Errorf("read submission files: %w", err)
	}
	out.Files = make([]ops.SubmissionFile, 0, len(files))
	for _, f := range files {
		if f.verdict.Valid {
			r, err := reasons.read(ctx, f.verdict.String)
			if err != nil {
				return out, err
			}
			f.Verdict = &r
		}
		out.Files = append(out.Files, f.SubmissionFile)
	}

	warnings, err := queryAll(ctx, s.subs.store, func(rows *sql.Rows) (storedWarning, error) {
		var w storedWarning
		var subject []byte
		if err := rows.Scan(&w.code, &subject); err != nil {
			return w, err
		}
		if subject != nil {
			name := validUTF8(string(subject))
			w.Subject = &name
		}
		return w, nil
	}, `SELECT code, subject FROM submission_warnings WHERE submission_id = ? ORDER BY at, id`, id)
	if err != nil {
		return out, fmt.Errorf("read submission warnings: %w", err)
	}
	out.Warnings = make([]ops.SubmissionWarning, 0, len(warnings))
	for _, w := range warnings {
		if w.Reason, err = reasons.read(ctx, w.code); err != nil {
			return out, err
		}
		out.Warnings = append(out.Warnings, w.SubmissionWarning)
	}
	return out, nil
}

// storedFile and storedWarning carry a row's stored code beside the
// wire value, which is rendered once the query's rows are closed.
type storedFile struct {
	ops.SubmissionFile
	verdict sql.NullString
}

type storedWarning struct {
	ops.SubmissionWarning
	code string
}

// A reasonReader renders the reject codes stored for one submission. The
// format a bad-extension verdict names is the submission's effective
// allowlist, read the first time one is rendered.
type reasonReader struct {
	seam       *Seam
	collection []byte
	allowlist  []string
}

func (r *reasonReader) read(ctx context.Context, code string) (reject.Reason, error) {
	reason, ok := reject.Lookup(reject.Code(code))
	if !ok {
		// A stored verdict this build cannot render must not read as a
		// pass.
		return reject.Reason{}, fmt.Errorf("stored reject code %q is outside the reject set", code)
	}
	if reason.Code != reject.BadExtension {
		return reason, nil
	}
	if r.allowlist == nil {
		lim, err := r.seam.limits(ctx)
		if err != nil {
			return reject.Reason{}, fmt.Errorf("read intake limits: %w", err)
		}
		var own sql.NullString
		if err := r.seam.subs.store.QueryRow(ctx, `SELECT admission_format_allowlist FROM collections WHERE id = ?`,
			r.collection).Scan(&own); err != nil {
			return reject.Reason{}, fmt.Errorf("read collection allowlist: %w", err)
		}
		r.allowlist = narrow(own, lim.Policy.Allowlist)
	}
	return reject.WithFormats(r.allowlist), nil
}

// queryAll scans every row query returns through scan.
func queryAll[T any](ctx context.Context, h *store.Handle, scan func(*sql.Rows) (T, error), query string, args ...any) ([]T, error) {
	rows, err := h.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// wireID renders an id as the wire carries it.
func wireID(id []byte) string { return hex.EncodeToString(id) }

func micros(at int64) time.Time { return time.UnixMicro(at).UTC() }

// validUTF8 renders a stored path or name that need not be UTF-8.
func validUTF8(s string) string { return strings.ToValidUTF8(s, "�") }
