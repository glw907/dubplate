package intake

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/reject"
)

// testSeamLimits are the limits testLimits holds, as the seam's
// settings reader returns them.
func testSeamLimits() Limits {
	l := testLimits()
	return Limits{Policy: l.policy, SessionsPerMember: l.sessionsPerMember}
}

// newTestSeam returns a store with one live admin and a seam over the
// test submissions writer, so its clock is fixed and its free-space
// read answers plentyFree.
func newTestSeam(t *testing.T) (*store.Handle, *Seam) {
	t.Helper()
	h, s := newTestSubmissions(t)
	return h, &Seam{subs: s, limits: func(context.Context) (Limits, error) { return testSeamLimits(), nil }}
}

// actorOf is the network actor a verified session for uploader builds.
func actorOf(uploader identity.UserID) ops.Actor {
	return ops.NetworkActor(identity.Identity{
		Address: "member@example.invalid", UserID: uploader, Principal: "member",
		Source: "magic-link", Channel: identity.ChannelCookie,
	}, time.UnixMicro(testNow))
}

// albumRequest is a create request for a loose-files album of n tracks.
func albumRequest(n int) ops.CreateSubmissionRequest {
	req := ops.CreateSubmissionRequest{VerdictUnit: "album", Container: "files", Source: "vinyl"}
	for i := range n {
		req.Files = append(req.Files, ops.ManifestFile{Path: fmt.Sprintf("%02d.flac", i+1), Size: 100})
	}
	return req
}

func mustCreateVia(t *testing.T, seam *Seam, uploader identity.UserID, req ops.CreateSubmissionRequest) ops.CreateSubmissionResponse {
	t.Helper()
	resp, err := seam.CreateSubmission(deadline(t), actorOf(uploader), req)
	if err != nil {
		t.Fatalf("CreateSubmission: %v", err)
	}
	return resp
}

// opsError returns err as the *ops.Error the wire answers from, and
// fails the test when it is not one.
func opsError(t *testing.T, err error) *ops.Error {
	t.Helper()
	e, ok := errors.AsType[*ops.Error](err)
	if !ok {
		t.Fatalf("error = %v, want an *ops.Error", err)
	}
	return e
}

// wireOf is a submission id's wire form, written out from its bytes.
func wireOf(id []byte) string { return fmt.Sprintf("%x", id) }

func TestSeamCreateTakesTheCallerFromTheActor(t *testing.T) {
	h, seam := newTestSeam(t)
	ann, bo := member(t, h), member(t, h)

	resp := mustCreateVia(t, seam, ann, albumRequest(2))

	id, ok := parseID(resp.SubmissionID)
	if !ok {
		t.Fatalf("submission id %q is not 32 lowercase hex characters", resp.SubmissionID)
	}
	var uploader, collection []byte
	var state string
	if err := h.QueryRow(deadline(t), `SELECT uploader, collection_id, state FROM submissions WHERE id = ?`, id).Scan(&uploader, &collection, &state); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(uploader, ann[:]) || bytes.Equal(uploader, bo[:]) {
		t.Errorf("uploader = %x, want the actor's %x", uploader, ann[:])
	}
	if state != "uploading" {
		t.Errorf("state = %q, want uploading", state)
	}
	var routeCollection []byte
	if err := h.QueryRow(deadline(t), `SELECT id FROM collections WHERE intake_route = 'default'`).Scan(&routeCollection); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(collection, routeCollection) {
		t.Errorf("collection = %x, want the route collection %x", collection, routeCollection)
	}
	if len(resp.Files) != 2 || resp.Files[0].Path != "01.flac" || resp.Files[1].Path != "02.flac" {
		t.Fatalf("files = %+v, want 01.flac and 02.flac", resp.Files)
	}
	for _, f := range resp.Files {
		fid, ok := parseID(f.FileID)
		if !ok {
			t.Fatalf("file id %q is not 32 lowercase hex characters", f.FileID)
		}
		var role string
		if err := h.QueryRow(deadline(t), `SELECT role FROM submission_files WHERE id = ? AND submission_id = ?`, fid, id).Scan(&role); err != nil {
			t.Errorf("file %s: %v", f.Path, err)
		}
		if role != "audio" {
			t.Errorf("file %s role = %q, want audio", f.Path, role)
		}
	}
}

func TestSeamCreateStoresEachSource(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	override(t, h, uploader, nil, nil, 100)
	for _, source := range []string{"bought", "cd", "vinyl", "free-download", "own-recording", "gift", "other"} {
		t.Run(source, func(t *testing.T) {
			req := albumRequest(1)
			req.Source = source
			resp := mustCreateVia(t, seam, uploader, req)
			got, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})
			if err != nil {
				t.Fatal(err)
			}
			if got.Source != source {
				t.Errorf("source = %q, want %q", got.Source, source)
			}
		})
	}
}

func TestSeamCreateRefusals(t *testing.T) {
	tests := []struct {
		name       string
		edit       func(*ops.CreateSubmissionRequest)
		wantSub    []reject.Code
		wantFiles  []string
		wantFileBy reject.Code
	}{
		{"source absent", func(r *ops.CreateSubmissionRequest) { r.Source = "" }, []reject.Code{reject.MissingSource}, nil, ""},
		{"source unknown", func(r *ops.CreateSubmissionRequest) { r.Source = "found" }, []reject.Code{reject.MissingSource}, nil, ""},
		{"no files", func(r *ops.CreateSubmissionRequest) { r.Files = nil }, []reject.Code{reject.NoAudio}, nil, ""},
		{"every failing path is listed", func(r *ops.CreateSubmissionRequest) {
			r.Files = []ops.ManifestFile{
				{Path: "../a.flac", Size: 1}, {Path: "ok.flac", Size: 1}, {Path: "/b.flac", Size: 1},
				{Path: "a/b/c/d/e/f/g/h/i.flac", Size: 1}, {Path: "x\\y.flac", Size: 1},
			}
		}, nil, []string{"../a.flac", "/b.flac", "a/b/c/d/e/f/g/h/i.flac", "x\\y.flac"}, ""},
		{"a format outside the allowlist", func(r *ops.CreateSubmissionRequest) {
			r.Files = append(r.Files, ops.ManifestFile{Path: "03.mp3", Size: 1})
		}, nil, []string{"03.mp3"}, reject.BadExtension},
		{"an album over the size cap", func(r *ops.CreateSubmissionRequest) {
			r.Files = r.Files[:1]
			r.Files[0].Size = 1 << 41
		}, []reject.Code{reject.TooLarge}, nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, seam := newTestSeam(t)
			uploader := member(t, h)
			req := albumRequest(2)
			tt.edit(&req)

			_, err := seam.CreateSubmission(deadline(t), actorOf(uploader), req)

			e := opsError(t, err)
			if e.Code != ops.CodeInvalidRequest || e.Detail != "" {
				t.Fatalf("error = %+v, want invalid_request with no detail", e)
			}
			if e.Rejects == nil || e.Rejects.Submission == nil || e.Rejects.Files == nil {
				t.Fatalf("rejects = %+v, want both lists present, empty when nothing fails", e.Rejects)
			}
			var gotSub []reject.Code
			for _, r := range e.Rejects.Submission {
				gotSub = append(gotSub, r.Code)
			}
			if !slices.Equal(gotSub, tt.wantSub) {
				t.Errorf("submission reasons = %v, want %v", gotSub, tt.wantSub)
			}
			var gotFiles []string
			for _, f := range e.Rejects.Files {
				gotFiles = append(gotFiles, f.Path)
				if tt.wantFileBy != "" && f.Reason.Code != tt.wantFileBy {
					t.Errorf("file %q reason = %q, want %q", f.Path, f.Reason.Code, tt.wantFileBy)
				}
			}
			if !slices.Equal(gotFiles, tt.wantFiles) {
				t.Errorf("failing paths = %q, want %q", gotFiles, tt.wantFiles)
			}
			if n := countSubmissions(t, h); n != 0 {
				t.Errorf("a refused create wrote %d submissions", n)
			}
		})
	}
}

func TestSeamCreateStoreURL(t *testing.T) {
	long := "https://example.com/" + strings.Repeat("a", maxStoreURL-len("https://example.com/"))
	tests := []struct {
		name string
		url  string
		ok   bool
	}{
		{"absent", "", true},
		{"https with a host", "https://store.example/album?x=1", true},
		{"exactly 2048 bytes", long, true},
		{"2049 bytes", long + "a", false},
		{"http", "http://store.example/album", false},
		{"no scheme", "store.example/album", false},
		{"host-less", "https:///album", false},
		{"port only", "https://:443/album", false},
		{"relative", "/album", false},
		{"ftp", "ftp://store.example/album", false},
		{"not a URL", "https://store example/%zz", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, seam := newTestSeam(t)
			uploader := member(t, h)
			req := albumRequest(1)
			req.StoreURL = tt.url

			resp, err := seam.CreateSubmission(deadline(t), actorOf(uploader), req)

			if tt.ok {
				if err != nil {
					t.Fatalf("CreateSubmission: %v", err)
				}
				if resp.SubmissionID == "" {
					t.Error("no submission id")
				}
				return
			}
			e := opsError(t, err)
			if e.Code != ops.CodeInvalidRequest || e.Detail != ops.DetailStoreURL {
				t.Errorf("error = %+v, want invalid_request with detail store_url", e)
			}
			if n := countSubmissions(t, h); n != 0 {
				t.Errorf("a refused create wrote %d submissions", n)
			}
		})
	}
}

func TestSeamCreateClientFaults(t *testing.T) {
	tests := []struct {
		name       string
		edit       func(*ops.CreateSubmissionRequest)
		wantDetail string
	}{
		{"a separate-tracks manifest naming a scan", func(r *ops.CreateSubmissionRequest) {
			r.VerdictUnit = "track"
			r.Files = append(r.Files, ops.ManifestFile{Path: "scan.jpg", Size: 5})
		}, ops.DetailExtra},
		{"a duplicate path", func(r *ops.CreateSubmissionRequest) {
			r.Files = append(r.Files, ops.ManifestFile{Path: "01.flac", Size: 5})
		}, ops.DetailManifest},
		{"a negative size", func(r *ops.CreateSubmissionRequest) { r.Files[0].Size = -1 }, ops.DetailManifest},
		{"an unknown verdict unit", func(r *ops.CreateSubmissionRequest) { r.VerdictUnit = "disc" }, ""},
		{"an empty verdict unit", func(r *ops.CreateSubmissionRequest) { r.VerdictUnit = "" }, ""},
		{"an unknown container", func(r *ops.CreateSubmissionRequest) { r.Container = "tar" }, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, seam := newTestSeam(t)
			uploader := member(t, h)
			req := albumRequest(2)
			tt.edit(&req)

			_, err := seam.CreateSubmission(deadline(t), actorOf(uploader), req)

			e := opsError(t, err)
			if e.Code != ops.CodeInvalidRequest || e.Detail != tt.wantDetail || e.Rejects != nil {
				t.Errorf("error = %+v, want invalid_request, detail %q, no rejects", e, tt.wantDetail)
			}
			if n := countSubmissions(t, h); n != 0 {
				t.Errorf("a refused create wrote %d submissions", n)
			}
		})
	}
}

func TestSeamAZipTracksSubmissionIsAccepted(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	resp := mustCreateVia(t, seam, uploader, ops.CreateSubmissionRequest{
		VerdictUnit: "track", Container: "zip", Source: "cd",
		Files: []ops.ManifestFile{{Path: "music.zip", Size: 500}},
	})
	got, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})
	if err != nil {
		t.Fatal(err)
	}
	if got.VerdictUnit != "track" || got.Container != "zip" || len(got.Files) != 1 || got.Files[0].Path != "music.zip" {
		t.Errorf("submission = %+v, want a one-archive separate-tracks submission", got)
	}
}

func TestSeamCreateLimitRefusals(t *testing.T) {
	tests := []struct {
		name     string
		arrange  func(t *testing.T, h *store.Handle, s *Seam, uploader identity.UserID)
		wantCode ops.Code
		want     reject.Code
	}{
		{"uploads off", func(t *testing.T, h *store.Handle, _ *Seam, u identity.UserID) { override(t, h, u, nil, 0, nil) },
			ops.CodeConflict, reject.UploadsOff},
		{"too many sessions", func(t *testing.T, h *store.Handle, _ *Seam, u identity.UserID) { override(t, h, u, nil, nil, 1) },
			ops.CodeConflict, reject.TooManySessions},
		{"over quota", func(t *testing.T, h *store.Handle, _ *Seam, u identity.UserID) { override(t, h, u, 150, nil, nil) },
			ops.CodeConflict, reject.OverQuota},
		{"no space", func(_ *testing.T, _ *store.Handle, s *Seam, _ identity.UserID) {
			s.subs.freeBytes = func(string) (uint64, error) { return 10, nil }
			s.limits = func(context.Context) (Limits, error) {
				l := testSeamLimits()
				l.FreeSpaceFloor = 1000
				return l, nil
			}
		}, ops.CodeConflict, reject.NoSpace},
		{"free space unreadable", func(_ *testing.T, _ *store.Handle, s *Seam, _ identity.UserID) {
			s.subs.freeBytes = func(string) (uint64, error) { return 0, errors.New("statfs: input/output error") }
		}, ops.CodeUnavailable, reject.CouldNotCheck},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, seam := newTestSeam(t)
			uploader := member(t, h)
			if tt.name == "too many sessions" {
				mustCreateVia(t, seam, uploader, albumRequest(1))
			}
			tt.arrange(t, h, seam, uploader)
			before := countSubmissions(t, h)

			_, err := seam.CreateSubmission(deadline(t), actorOf(uploader), albumRequest(2))

			e := opsError(t, err)
			if e.Code != tt.wantCode {
				t.Fatalf("code = %q, want %q", e.Code, tt.wantCode)
			}
			if e.Rejects == nil || len(e.Rejects.Submission) != 1 || e.Rejects.Submission[0].Code != tt.want || e.Rejects.Files == nil || len(e.Rejects.Files) != 0 {
				t.Errorf("rejects = %+v, want exactly the %s reason and an empty file list", e.Rejects, tt.want)
			}
			if n := countSubmissions(t, h); n != before {
				t.Errorf("a refused create wrote %d submissions", n-before)
			}
		})
	}
}

func TestSeamLimitsFailureIsNotAVerdict(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	seam.limits = func(context.Context) (Limits, error) { return Limits{}, errors.New("setting unreadable") }

	_, err := seam.CreateSubmission(deadline(t), actorOf(uploader), albumRequest(1))

	if err == nil {
		t.Fatal("CreateSubmission with unreadable limits succeeded")
	}
	if _, isOps := errors.AsType[*ops.Error](err); isOps {
		t.Errorf("error = %v, want one that answers internal, never a coded refusal", err)
	}
}

func TestSeamBusyStoreIsRetryable(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	ctx, cancel := context.WithDeadline(t.Context(), time.Now().Add(-time.Second))
	defer cancel()

	_, err := seam.CreateSubmission(ctx, actorOf(uploader), albumRequest(1))

	e := opsError(t, err)
	if e.Code != ops.CodeUnavailable || e.Detail != ops.DetailBusy {
		t.Errorf("error = %+v, want unavailable with detail busy", e)
	}
}

// twoMembers returns a seam over two members with one submission each,
// the first member and both submissions' ids.
func twoMembers(t *testing.T) (h *store.Handle, seam *Seam, ann identity.UserID, annsID, bosID string) {
	t.Helper()
	h, seam = newTestSeam(t)
	ann, bo := member(t, h), member(t, h)
	annsID = mustCreateVia(t, seam, ann, albumRequest(1)).SubmissionID
	bosID = mustCreateVia(t, seam, bo, albumRequest(1)).SubmissionID
	return h, seam, ann, annsID, bosID
}

func TestSeamAnotherMembersIdIsAnUnknownId(t *testing.T) {
	h, seam, ann, annsID, bosID := twoMembers(t)
	unknown := "ffffffffffffffffffffffffffffffff"
	ctx := deadline(t)
	actor := actorOf(ann)
	calls := map[string]func(id string) error{
		"get": func(id string) error {
			_, err := seam.GetSubmission(ctx, actor, ops.GetSubmissionRequest{SubmissionID: id})
			return err
		},
		"retry": func(id string) error {
			_, err := seam.RetrySubmission(ctx, actor, ops.RetrySubmissionRequest{SubmissionID: id})
			return err
		},
		"cancel": func(id string) error {
			_, err := seam.CancelSubmission(ctx, actor, ops.CancelSubmissionRequest{SubmissionID: id})
			return err
		},
	}
	for name, call := range calls {
		t.Run(name, func(t *testing.T) {
			other, missing := opsError(t, call(bosID)), opsError(t, call(unknown))
			if other.Code != ops.CodeNotFound {
				t.Fatalf("another member's id answered %q, want not_found", other.Code)
			}
			if !reflect.DeepEqual(other, missing) {
				t.Errorf("another member's id answered %+v, an unknown id %+v: the bodies must be identical", other, missing)
			}
		})
	}

	if got := stateOf(t, h, mustParse(t, bosID)); got != "uploading" {
		t.Errorf("another member's submission is %q after the refused cancel, want uploading", got)
	}
	if _, err := seam.GetSubmission(ctx, actor, ops.GetSubmissionRequest{SubmissionID: annsID}); err != nil {
		t.Errorf("the caller's own submission: %v", err)
	}
}

func mustParse(t *testing.T, wire string) []byte {
	t.Helper()
	id, ok := parseID(wire)
	if !ok {
		t.Fatalf("%q is not a wire id", wire)
	}
	return id
}

func TestSeamMalformedIdsAreInvalid(t *testing.T) {
	_, seam, ann, _, _ := twoMembers(t)
	for _, id := range []string{"", "abc", "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFF", "00000000000000000000000000000000", "gfffffffffffffffffffffffffffffff"} {
		t.Run(fmt.Sprintf("%q", id), func(t *testing.T) {
			_, err := seam.GetSubmission(deadline(t), actorOf(ann), ops.GetSubmissionRequest{SubmissionID: id})
			if e := opsError(t, err); e.Code != ops.CodeInvalidRequest {
				t.Errorf("code = %q, want invalid_request", e.Code)
			}
		})
	}
}

func TestSeamListsOnlyTheCallersNewestFirst(t *testing.T) {
	h, seam := newTestSeam(t)
	ann, bo := member(t, h), member(t, h)
	override(t, h, ann, nil, nil, 10)
	clock := int64(testNow)
	seam.subs.now = func() time.Time { clock += 1000; return time.UnixMicro(clock) }

	first := mustCreateVia(t, seam, ann, albumRequest(1)).SubmissionID
	mustCreateVia(t, seam, bo, albumRequest(1))
	second := mustCreateVia(t, seam, ann, albumRequest(1)).SubmissionID

	got, err := seam.ListSubmissions(deadline(t), actorOf(ann), ops.ListSubmissionsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	var ids []string
	for _, s := range got.Submissions {
		ids = append(ids, s.SubmissionID)
		if s.State != "uploading" || s.VerdictUnit != "album" || s.CreatedAt.IsZero() || s.UpdatedAt.IsZero() {
			t.Errorf("row = %+v, want an uploading album with its times", s)
		}
	}
	if !slices.Equal(ids, []string{second, first}) {
		t.Errorf("listed %v, want the caller's two, newest %s first", ids, second)
	}

	empty, err := seam.ListSubmissions(deadline(t), actorOf(member(t, h)), ops.ListSubmissionsRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if empty.Submissions == nil || len(empty.Submissions) != 0 {
		t.Errorf("a member with no uploads listed %#v, want an empty non-nil list", empty.Submissions)
	}
}

// A wording is the four strings a reason carries on the wire.
type wording struct{ code, label, friendly, nextStep string }

func equalWording(r *reject.Reason, w wording) bool {
	return r != nil && r.Code == reject.Code(w.code) && r.Label == w.label && r.Friendly == w.friendly && r.NextStep == w.nextStep
}

// TestSeamGetRendersEveryReasonAsItsWording holds that each reason a
// submission carries reaches the wire as the four strings the reject set
// holds, written out here from internal/reject's text.
func TestSeamGetRendersEveryReasonAsItsWording(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	req := albumRequest(3)
	resp := mustCreateVia(t, seam, uploader, req)
	id := mustParse(t, resp.SubmissionID)
	if err := seam.subs.receive(deadline(t), mustParse(t, resp.Files[0].FileID)); err != nil {
		t.Fatal(err)
	}
	for _, f := range resp.Files[1:] {
		if err := seam.subs.receive(deadline(t), mustParse(t, f.FileID)); err != nil {
			t.Fatal(err)
		}
	}
	v := verdict{to: stateRejected, reason: string(reject.IncompleteAlbum), warnings: []gate.FileReason{
		{Reason: reject.Of(reject.NoCoverArt)},
		{Path: []byte("scan.jpg"), Reason: reject.Of(reject.ExtrasDropped)},
	}}
	for i, f := range resp.Files {
		fv := fileVerdict{id: mustParse(t, f.FileID)}
		switch i {
		case 0:
			fv.result.Verdict = reject.Corrupt
		case 1:
			fv.result.Verdict = reject.BelowCDQuality
		}
		v.files = append(v.files, fv)
	}
	if err := seam.subs.commitVerdict(deadline(t), id, v); err != nil {
		t.Fatal(err)
	}

	got, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})
	if err != nil {
		t.Fatal(err)
	}

	corrupt := wording{"corrupt", "damaged in transit",
		"This file didn't survive the upload intact. Uploading it again usually fixes this.",
		"Upload it again from the original file."}
	belowCD := wording{"below-cd-quality", "below CD quality",
		"This FLAC is below CD quality.", "Get a CD-quality or better FLAC and upload that instead."}
	incomplete := wording{"incomplete-album", "album incomplete",
		"This album's track numbers skip or don't match its track count, so it looks incomplete.",
		"Find the missing tracks and upload the whole album together."}
	noCover := wording{"no-cover-art", "no cover",
		"This album has no cover picture.",
		"It can go in without one, or you can cancel, add a picture to the files, and upload again."}
	extras := wording{"extras-dropped", "extras left out",
		"When you upload separate tracks, only the music files are kept, so the other files in the ZIP were left out.",
		"If you wanted them kept, upload the whole album instead."}

	if got.State != "rejected" {
		t.Errorf("state = %q, want rejected", got.State)
	}
	if len(got.Files) != 3 {
		t.Fatalf("files = %d, want 3", len(got.Files))
	}
	if !equalWording(got.Files[0].Verdict, corrupt) ||
		!equalWording(got.Files[1].Verdict, belowCD) || got.Files[2].Verdict != nil {
		t.Errorf("verdicts = %+v, %+v, %+v; want corrupt, below-cd-quality, and null for the pass", got.Files[0].Verdict, got.Files[1].Verdict, got.Files[2].Verdict)
	}
	for _, f := range got.Files {
		if !f.Received {
			t.Errorf("file %s received = false, want true", f.Path)
		}
	}
	if n := len(got.Events); n != 3 || got.Events[0].Kind != "uploading" || got.Events[1].Kind != "checking" || got.Events[2].Kind != "rejected" {
		t.Fatalf("events = %+v, want uploading, checking, rejected in order", got.Events)
	}
	if last := got.Events[2]; last.Reason != "incomplete-album" || !equalWording(last.Reject, incomplete) {
		t.Errorf("rejected event = %+v, want the incomplete-album reason and its wording", last)
	}
	if got.Events[0].Reject != nil || got.Events[0].Reason != "" {
		t.Errorf("uploading event = %+v, want no reason", got.Events[0])
	}
	if len(got.Warnings) != 2 {
		t.Fatalf("warnings = %+v, want 2", got.Warnings)
	}
	if w := got.Warnings[0]; !equalWording(&w.Reason, noCover) || w.Subject != nil {
		t.Errorf("completeness warning = %+v, want no-cover-art with a null subject", w)
	}
	if w := got.Warnings[1]; !equalWording(&w.Reason, extras) || w.Subject == nil || *w.Subject != "scan.jpg" {
		t.Errorf("extras warning = %+v, want extras-dropped naming scan.jpg", w)
	}
}

func TestSeamBadExtensionNamesTheEffectiveAllowlist(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	seam.limits = func(context.Context) (Limits, error) {
		l := testSeamLimits()
		l.Policy.Allowlist = []string{"flac", "mp3"}
		return l, nil
	}
	resp := mustCreateVia(t, seam, uploader, albumRequest(1))
	id := mustParse(t, resp.SubmissionID)
	if err := seam.subs.receive(deadline(t), mustParse(t, resp.Files[0].FileID)); err != nil {
		t.Fatal(err)
	}
	v := verdict{to: stateRejected, files: []fileVerdict{{id: mustParse(t, resp.Files[0].FileID)}}}
	v.files[0].result.Verdict = reject.BadExtension
	if err := seam.subs.commitVerdict(deadline(t), id, v); err != nil {
		t.Fatal(err)
	}

	got, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})
	if err != nil {
		t.Fatal(err)
	}

	want := wording{"bad-extension", "wrong format",
		"This isn't one of the formats the library keeps (FLAC or MP3).",
		"Get a copy in one of these formats and upload that instead."}
	if v := got.Files[0].Verdict; !equalWording(v, want) {
		t.Errorf("verdict = %+v, want %+v", v, want)
	}
}

func TestSeamRendersANonUTF8PathWithTheReplacementCharacter(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	req := albumRequest(1)
	req.Files[0].Path = "caf\xe9.flac"
	resp := mustCreateVia(t, seam, uploader, req)

	got, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})
	if err != nil {
		t.Fatal(err)
	}
	if got.Files[0].Path != "caf�.flac" {
		t.Errorf("path = %q, want %q", got.Files[0].Path, "caf�.flac")
	}
}

// TestSeamAVerdictItCannotRenderIsNotAPass holds that a stored verdict
// outside the reject set fails the read rather than showing as a file
// that passed.
func TestSeamAVerdictItCannotRenderIsNotAPass(t *testing.T) {
	h, seam := newTestSeam(t)
	uploader := member(t, h)
	resp := mustCreateVia(t, seam, uploader, albumRequest(1))
	if err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE submission_files SET verdict = 'no-such-code'`)
		return err
	}); err != nil {
		t.Fatal(err)
	}

	_, err := seam.GetSubmission(deadline(t), actorOf(uploader), ops.GetSubmissionRequest{SubmissionID: resp.SubmissionID})

	if err == nil {
		t.Fatal("GetSubmission rendered an unknown stored verdict")
	}
	if _, isOps := errors.AsType[*ops.Error](err); isOps {
		t.Errorf("error = %v, want a server fault that answers internal", err)
	}
}

// TestSeamActionsFromEveryState holds each of retry and cancel against
// each journey state: the one move the transition table gives it, or
// conflict, with the answer carrying the submission as it then stands.
func TestSeamActionsFromEveryState(t *testing.T) {
	moves := map[string]map[state]state{
		"retry":  {stateCheckUnfinished: stateChecking},
		"cancel": {stateUploading: stateCancelled, stateCheckUnfinished: stateCancelled},
	}
	for action, table := range moves {
		for _, from := range states {
			t.Run(action+" from "+string(from), func(t *testing.T) {
				h, seam := newTestSeam(t)
				uploader := member(t, h)
				c := inState(t, seam.subs, uploader, from, 10, 20)
				wire := wireOf(c.id)

				var gotState string
				var err error
				if action == "retry" {
					var resp ops.RetrySubmissionResponse
					resp, err = seam.RetrySubmission(deadline(t), actorOf(uploader), ops.RetrySubmissionRequest{SubmissionID: wire})
					gotState = resp.State
				} else {
					var resp ops.CancelSubmissionResponse
					resp, err = seam.CancelSubmission(deadline(t), actorOf(uploader), ops.CancelSubmissionRequest{SubmissionID: wire})
					gotState = resp.State
				}

				to, moves := table[from]
				if !moves {
					if e := opsError(t, err); e.Code != ops.CodeConflict {
						t.Fatalf("code = %q, want conflict", e.Code)
					}
					if got := stateOf(t, h, c.id); got != string(from) {
						t.Errorf("state = %q after a refused %s, want %q", got, action, from)
					}
					return
				}
				if err != nil {
					t.Fatalf("%s: %v", action, err)
				}
				if gotState != string(to) || stateOf(t, h, c.id) != string(to) {
					t.Errorf("answer state %q, stored %q, want %q", gotState, stateOf(t, h, c.id), to)
				}
				assertStatesMatchEvents(t, h)
			})
		}
	}
}

// TestLayering holds the admin-api-1 rule: the operation vocabulary and
// its HTTP projection import neither the auth nor the settings package,
// and intake imports no auth.
func TestLayering(t *testing.T) {
	const module = "github.com/glw907/dubplate/internal/"
	tests := []struct {
		pkg       string
		forbidden []string
	}{
		{"../ops", []string{module + "auth", module + "settings"}},
		{"../ops/opshttp", []string{module + "auth", module + "settings"}},
		{".", []string{module + "auth"}},
	}
	for _, tt := range tests {
		t.Run(tt.pkg, func(t *testing.T) {
			out, err := exec.CommandContext(t.Context(), "go", "list", "-deps", tt.pkg).Output()
			if err != nil {
				t.Fatalf("go list -deps %s: %v", tt.pkg, err)
			}
			deps := strings.Fields(string(out))
			if !slices.Contains(deps, module+"ops") && tt.pkg != "../ops" {
				t.Fatalf("go list -deps %s listed no internal/ops dependency: %v", tt.pkg, deps)
			}
			for _, dep := range deps {
				if slices.Contains(tt.forbidden, dep) {
					t.Errorf("%s depends on %s", tt.pkg, dep)
				}
			}
		})
	}
}
