package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// The second plain member of the submission fixtures: seedMembers gives
// the member at position 4 the id whose first two bytes are 0xf0 and 4.
const apiBoID = "f0040000000000000000000000000000"

// maxStoreURL is the longest store URL the wire takes, written out.
const maxStoreURL = 2048

// submissionMembers is apiMembers and one more plain member, so two
// members each hold a submission.
var submissionMembers = append(slices.Clone(apiMembers),
	seedMember{address: "bo@example.com", displayName: "Bo", username: "bo"})

// newSubmissionHarness is newAPIHarness over submissionMembers with the
// free-space floor at its 1 GiB minimum, so a test's answer never turns
// on how much room the machine's temporary directory has.
func newSubmissionHarness(t *testing.T) *apiHarness {
	t.Helper()
	a := newAPIHarness(t, submissionMembers, nil)
	writeAdminRows(t, a.h, map[string]string{"DUBPLATE_INTAKE_FREE_SPACE_FLOOR": "1073741824"})
	return a
}

// albumBody is a create_submission body for an album of the named files,
// each 1000 bytes, from source vinyl.
func albumBody(paths ...string) string {
	files := make([]string, len(paths))
	for i, p := range paths {
		files[i] = `{"path":` + strconv.Quote(p) + `,"size":1000}`
	}
	return `{"verdict_unit":"album","container":"files","source":"vinyl","files":[` + strings.Join(files, ",") + `]}`
}

// createdID creates a one-file album as session and returns its id.
func (a *apiHarness) createdID(session string) string {
	a.t.Helper()
	rec := a.call("create_submission", session, albumBody("01.flac"))
	if rec.Code != http.StatusOK {
		a.t.Fatalf("create_submission = %d %s, want 200", rec.Code, rec.Body)
	}
	var resp ops.CreateSubmissionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		a.t.Fatalf("decode %s: %v", rec.Body, err)
	}
	return resp.SubmissionID
}

// errorBody decodes rec's JSON error body.
func errorBody(t *testing.T, rec *httptest.ResponseRecorder) (code, detail string, rejects *ops.Rejects) {
	t.Helper()
	var body struct {
		Error struct {
			Code    string       `json:"code"`
			Detail  string       `json:"detail"`
			Rejects *ops.Rejects `json:"rejects"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("answer %d %.200q is not a JSON error: %v", rec.Code, rec.Body, err)
	}
	return body.Error.Code, body.Error.Detail, body.Error.Rejects
}

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("decode %q: %v", s, err)
	}
	return b
}

// TestAPI_MemberReachesTheSubmissionOperationsAndNoAdminOne proves a
// plain member's session meets all five submission operations and is
// refused every admin and job operation (AC13, AC29).
func TestAPI_MemberReachesTheSubmissionOperationsAndNoAdminOne(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())

	id := a.createdID(member)
	idBody := `{"submission_id":"` + id + `"}`
	for op, body := range map[string]string{
		"list_submissions": `{}`, "get_submission": idBody, "cancel_submission": idBody,
	} {
		if rec := a.call(op, member, body); rec.Code != http.StatusOK {
			t.Errorf("a member's %s = %d %s, want 200", op, rec.Code, rec.Body)
		}
	}
	if rec := a.call("retry_submission", member, idBody); rec.Code != http.StatusConflict {
		t.Errorf("a member's retry of a cancelled submission = %d %s, want 409: the member reaches the operation and the state refuses it", rec.Code, rec.Body)
	}

	for op, body := range map[string]string{
		"list_members": `{}`, "add_member": `{"address":"x@example.com","display_name":"X","username":"x"}`,
		"remove_member": `{"user_id":"` + apiBoID + `"}`, "demote_member": `{"user_id":"` + apiAdminID + `"}`,
		"list_settings": `{}`, "set_setting": `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h"}`,
		"reset_setting": `{"key":"DUBPLATE_REAUTH_WINDOW"}`, "start_backup": `{}`,
		"job_status": `{"job_id":"` + apiBoID + `"}`, "cancel_job": `{"job_id":"` + apiBoID + `"}`,
	} {
		if rec := a.call(op, member, body); rec.Code != http.StatusForbidden {
			t.Errorf("a member's %s = %d %s, want 403", op, rec.Code, rec.Body)
		}
	}

	catalog := a.call("catalog", member, `{}`)
	var list ops.CatalogResponse
	if err := json.Unmarshal(catalog.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode the catalog %s: %v", catalog.Body, err)
	}
	privileges := map[string]ops.Privilege{}
	for _, e := range list.Operations {
		privileges[e.Name] = e.Privilege
	}
	for _, op := range []string{"create_submission", "list_submissions", "get_submission", "retry_submission", "cancel_submission"} {
		if privileges[op] != ops.PrivilegeMember {
			t.Errorf("the catalog lists %s as %v, want member", op, privileges[op])
		}
	}
}

// TestAPI_SignedOutSubmissionCallsAre401 proves a request with no
// session reaches no submission operation (AC29).
func TestAPI_SignedOutSubmissionCallsAre401(t *testing.T) {
	a := newSubmissionHarness(t)
	for _, op := range []string{"create_submission", "list_submissions", "get_submission", "retry_submission", "cancel_submission"} {
		if rec := a.call(op, "", `{}`); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s with no session = %d, want 401", op, rec.Code)
		}
	}
	var n int
	if err := a.h.QueryRow(t.Context(), `SELECT count(*) FROM submissions`).Scan(&n); err != nil || n != 0 {
		t.Errorf("submissions = %d (%v), want none written", n, err)
	}
}

// TestAPI_AnotherMembersSubmissionIsAnUnknownOne proves, with two
// members holding one submission each, that a member's get, retry, and
// cancel on the other's id answer the byte-identical 404 an unknown id
// does, that list shows only the caller's own, and that nothing of the
// other's changed (AC13, AC29).
func TestAPI_AnotherMembersSubmissionIsAnUnknownOne(t *testing.T) {
	a := newSubmissionHarness(t)
	member, bo := a.session("member@example.com", time.Now()), a.session("bo@example.com", time.Now())
	memberID, boID := a.createdID(member), a.createdID(bo)
	unknown := strings.Repeat("ab", 16)

	for _, op := range []string{"get_submission", "retry_submission", "cancel_submission"} {
		t.Run(op, func(t *testing.T) {
			other := a.call(op, member, `{"submission_id":"`+boID+`"}`)
			missing := a.call(op, member, `{"submission_id":"`+unknown+`"}`)
			if other.Code != http.StatusNotFound || missing.Code != http.StatusNotFound {
				t.Fatalf("another member's id = %d, an unknown id = %d, want both 404", other.Code, missing.Code)
			}
			if other.Body.String() != missing.Body.String() {
				t.Errorf("another member's id answered %s, an unknown id %s: they must be identical", other.Body, missing.Body)
			}
			if code, _, _ := errorBody(t, other); code != "not_found" {
				t.Errorf("code = %q, want not_found", code)
			}
		})
	}

	var state string
	if err := a.h.QueryRow(t.Context(), `SELECT state FROM submissions WHERE id = ?`, mustHex(t, boID)).Scan(&state); err != nil || state != "uploading" {
		t.Errorf("another member's submission is %q (%v) after the refused cancel, want uploading", state, err)
	}

	rec := a.call("list_submissions", member, `{}`)
	var list ops.ListSubmissionsResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	if len(list.Submissions) != 1 || list.Submissions[0].SubmissionID != memberID {
		t.Errorf("list_submissions = %+v, want only the caller's %s", list.Submissions, memberID)
	}
	if own := a.call("get_submission", member, `{"submission_id":"`+memberID+`"}`); own.Code != http.StatusOK {
		t.Errorf("the caller's own get = %d %s, want 200", own.Code, own.Body)
	}
}

// TestAPI_CreateNeverTakesAFieldFromTheBody proves a body naming another
// member as uploader, a collection, a state, a reservation, and a role
// changes none of them: the row's uploader is the session's member
// (AC13).
func TestAPI_CreateNeverTakesAFieldFromTheBody(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	body := `{"verdict_unit":"album","container":"files","source":"vinyl","files":[{"path":"01.flac","size":1000,"role":"extra"}],
		"uploader":"` + apiBoID + `","user_id":"` + apiBoID + `","collection_id":"` + apiBoID + `","collection":"` + apiBoID + `",
		"state":"accepted","reserved_bytes":0,"declared_bytes":0,"role":"admin","admin":true,"actor":"member:` + apiBoID + `"}`
	rec := a.call("create_submission", member, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("create_submission = %d %s, want 200", rec.Code, rec.Body)
	}
	var resp ops.CreateSubmissionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	var uploader, collection []byte
	var state string
	var reserved int64
	err := a.h.QueryRow(t.Context(), `SELECT uploader, collection_id, state, reserved_bytes FROM submissions WHERE id = ?`,
		mustHex(t, resp.SubmissionID)).Scan(&uploader, &collection, &state, &reserved)
	if err != nil {
		t.Fatal(err)
	}
	if want := mustHex(t, apiMemberID); !bytes.Equal(uploader, want) {
		t.Errorf("uploader = %x, want the session's member %x", uploader, want)
	}
	var route string
	if err := a.h.QueryRow(t.Context(), `SELECT intake_route FROM collections WHERE id = ?`, collection).Scan(&route); err != nil || route != "default" {
		t.Errorf("collection route = %q (%v), want the default route collection", route, err)
	}
	if state != "uploading" || reserved != 1000 {
		t.Errorf("state %q reserved %d, want uploading and 1000", state, reserved)
	}
	var role string
	if err := a.h.QueryRow(t.Context(), `SELECT role FROM submission_files WHERE submission_id = ?`, mustHex(t, resp.SubmissionID)).Scan(&role); err != nil || role != "audio" {
		t.Errorf("file role = %q (%v), want the server-derived audio", role, err)
	}
}

// TestAPI_RefusedCreateListsEveryFailingPathWithItsWording proves a
// refused create answers 400 with every failing path in error.rejects
// and each reason as the four strings internal/reject holds, written out
// here (AC14, AC29).
func TestAPI_RefusedCreateListsEveryFailingPathWithItsWording(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	body := `{"verdict_unit":"album","container":"files","source":"","files":[
		{"path":"../a.flac","size":1},{"path":"ok.flac","size":1},{"path":"/b.flac","size":1},{"path":"c.mp3","size":1}]}`

	rec := a.call("create_submission", member, body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body)
	}
	code, detail, rejects := errorBody(t, rec)
	if code != "invalid_request" || detail != "" || rejects == nil {
		t.Fatalf("code %q detail %q rejects %v, want invalid_request, no detail, and rejects", code, detail, rejects)
	}
	if len(rejects.Submission) != 1 {
		t.Fatalf("submission reasons = %+v, want only missing-source", rejects.Submission)
	}
	missing := rejects.Submission[0]
	if missing.Code != "missing-source" || missing.Label != "source missing" ||
		missing.Friendly != "This upload doesn't say where the music came from." ||
		missing.NextStep != "Choose where it came from, then upload it again." {
		t.Errorf("missing-source reason = %+v, want PCF's wording", missing)
	}
	got := map[string]string{}
	for _, f := range rejects.Files {
		got[f.Path] = string(f.Reason.Code)
	}
	want := map[string]string{"../a.flac": "unsafe-path", "/b.flac": "unsafe-path", "c.mp3": "bad-extension"}
	if !maps.Equal(got, want) {
		t.Errorf("failing files = %v, want %v", got, want)
	}
	for _, f := range rejects.Files {
		if f.Path == "c.mp3" && (f.Reason.Label != "not FLAC" || f.Reason.Friendly != "This isn't a FLAC file, and the library keeps lossless FLAC." ||
			f.Reason.NextStep != "Download this album as FLAC and upload that instead.") {
			t.Errorf("bad-extension reason = %+v, want PCF's wording", f.Reason)
		}
	}
	var n int
	if err := a.h.QueryRow(t.Context(), `SELECT count(*) FROM submissions`).Scan(&n); err != nil || n != 0 {
		t.Errorf("submissions = %d (%v), want none written", n, err)
	}
}

// TestAPI_SourceAnswers proves a submission without a source, with an
// empty one, and with an unknown one each answer missing-source, and
// each of the seven values is stored and read back (AC14).
func TestAPI_SourceAnswers(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	withSource := func(source string) string {
		return `{"verdict_unit":"album","container":"files","files":[{"path":"01.flac","size":1000}]` + source + `}`
	}
	for name, body := range map[string]string{
		"absent": withSource(``), "empty": withSource(`,"source":""`), "unknown": withSource(`,"source":"found"`),
	} {
		t.Run(name, func(t *testing.T) {
			rec := a.call("create_submission", member, body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body)
			}
			if _, _, rejects := errorBody(t, rec); rejects == nil || len(rejects.Submission) != 1 || rejects.Submission[0].Code != "missing-source" {
				t.Errorf("rejects = %+v, want missing-source", rejects)
			}
		})
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	if err := a.h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO member_overrides (user_id, upload_concurrency_cap, updated_at) VALUES (?, 20, 1)`, mustHex(t, apiMemberID))
		return err
	}); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"bought", "cd", "vinyl", "free-download", "own-recording", "gift", "other"} {
		t.Run(source, func(t *testing.T) {
			rec := a.call("create_submission", member, withSource(`,"source":"`+source+`"`))
			if rec.Code != http.StatusOK {
				t.Fatalf("create_submission = %d %s, want 200", rec.Code, rec.Body)
			}
			var created ops.CreateSubmissionResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
				t.Fatal(err)
			}
			got := a.call("get_submission", member, `{"submission_id":"`+created.SubmissionID+`"}`)
			var status ops.GetSubmissionResponse
			if err := json.Unmarshal(got.Body.Bytes(), &status); err != nil || status.Source != source {
				t.Errorf("get_submission source = %q (%v), want %q", status.Source, err, source)
			}
		})
	}
}

// TestAPI_StoreURL proves a malformed, http, host-less, and 2049-byte
// store URL each answer invalid_request with detail store_url, and a
// 2048-byte one is stored.
func TestAPI_StoreURL(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	long := "https://store.example/" + strings.Repeat("a", maxStoreURL-len("https://store.example/"))
	for name, url := range map[string]string{
		"malformed": "https://store example/%zz", "http": "http://store.example/a", "host-less": "https:///a", "2049 bytes": long + "a",
	} {
		t.Run(name, func(t *testing.T) {
			body := strings.Replace(albumBody("01.flac"), `"source":"vinyl"`, `"source":"vinyl","store_url":`+strconv.Quote(url), 1)
			rec := a.call("create_submission", member, body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body)
			}
			if code, detail, _ := errorBody(t, rec); code != "invalid_request" || detail != "store_url" {
				t.Errorf("code %q detail %q, want invalid_request and store_url", code, detail)
			}
		})
	}
	good := strings.Replace(albumBody("01.flac"), `"source":"vinyl"`, `"source":"vinyl","store_url":"`+long+`"`, 1)
	if rec := a.call("create_submission", member, good); rec.Code != http.StatusOK {
		t.Errorf("a 2048-byte https URL = %d %s, want 200", rec.Code, rec.Body)
	}
}

// TestAPI_ManifestFaultsHaveTheirDetails proves a separate-tracks
// manifest naming a scan answers detail extra, and a duplicate path or
// a negative size detail manifest.
func TestAPI_ManifestFaultsHaveTheirDetails(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	for name, tt := range map[string]struct{ body, detail string }{
		"a scan in separate tracks": {`{"verdict_unit":"track","container":"files","source":"cd","files":[{"path":"01.flac","size":1},{"path":"scan.jpg","size":1}]}`, "extra"},
		"a duplicate path":          {`{"verdict_unit":"album","container":"files","source":"cd","files":[{"path":"01.flac","size":1},{"path":"01.flac","size":1}]}`, "manifest"},
		"a negative size":           {`{"verdict_unit":"album","container":"files","source":"cd","files":[{"path":"01.flac","size":-1}]}`, "manifest"},
	} {
		t.Run(name, func(t *testing.T) {
			rec := a.call("create_submission", member, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d %s, want 400", rec.Code, rec.Body)
			}
			if code, detail, _ := errorBody(t, rec); code != "invalid_request" || detail != tt.detail {
				t.Errorf("code %q detail %q, want invalid_request and %q", code, detail, tt.detail)
			}
		})
	}
}

// TestAPI_CreateBodyCaps proves a 500-path manifest, too large for the
// 4 KiB every other operation takes, creates, and a body over 8 MiB
// answers invalid_request.
func TestAPI_CreateBodyCaps(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	paths := make([]string, 500)
	for i := range paths {
		paths[i] = fmt.Sprintf("disc 1/%04d - a track named long enough to make the manifest outgrow four kibibytes.flac", i)
	}
	body := albumBody(paths...)
	if len(body) <= 4<<10 {
		t.Fatalf("manifest is %d bytes, want over 4 KiB", len(body))
	}
	rec := a.call("create_submission", member, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("a 500-path manifest = %d %.200s, want 200", rec.Code, rec.Body)
	}
	var created ops.CreateSubmissionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil || len(created.Files) != 500 {
		t.Fatalf("created %d files (%v), want 500", len(created.Files), err)
	}

	over := `{"pad":"` + strings.Repeat("x", 8<<20) + `"}`
	rec = a.call("create_submission", member, over)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("a body over 8 MiB = %d, want 400", rec.Code)
	}
	if code, _, _ := errorBody(t, rec); code != "invalid_request" {
		t.Errorf("code = %q, want invalid_request", code)
	}
}

// TestAPI_ARejectedSubmissionCarriesEveryReasonWithItsWording proves a
// rejected submission's file verdicts, its rejected event, and its
// warnings reach the wire as the four strings internal/reject holds, and
// that a pass is a null verdict (AC29).
func TestAPI_ARejectedSubmissionCarriesEveryReasonWithItsWording(t *testing.T) {
	a := newSubmissionHarness(t)
	member := a.session("member@example.com", time.Now())
	id := a.createdID(member)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	if err := a.h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		sid := mustHex(t, id)
		for _, q := range []string{
			`UPDATE submission_files SET verdict = 'corrupt' WHERE submission_id = ?1`,
			`UPDATE submissions SET state = 'rejected' WHERE id = ?1`,
			`INSERT INTO submission_events (id, submission_id, seq, at, kind, reason) VALUES (randomblob(16), ?1, 2, 5, 'rejected', 'corrupt')`,
			`INSERT INTO submission_warnings (id, submission_id, code, subject, at) VALUES (randomblob(16), ?1, 'extras-dropped', CAST('scan.jpg' AS BLOB), 6)`,
		} {
			if _, err := tx.Exec(ctx, q, sid); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}

	rec := a.call("get_submission", member, `{"submission_id":"`+id+`"}`)

	var got struct {
		State  string `json:"state"`
		Events []struct {
			Kind   string            `json:"kind"`
			Reject map[string]string `json:"reject"`
		} `json:"events"`
		Files []struct {
			Verdict map[string]string `json:"verdict"`
		} `json:"files"`
		Warnings []struct {
			Reason  map[string]string `json:"reason"`
			Subject string            `json:"subject"`
		} `json:"warnings"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %s: %v", rec.Body, err)
	}
	corrupt := map[string]string{
		"code": "corrupt", "label": "damaged in transit",
		"friendly":  "This file didn't survive the upload intact. Uploading it again usually fixes this.",
		"next_step": "Upload it again from the original file.",
	}
	extras := map[string]string{
		"code": "extras-dropped", "label": "extras left out",
		"friendly":  "When you upload separate tracks, only the music files are kept, so the other files in the ZIP were left out.",
		"next_step": "If you wanted them kept, upload the whole album instead.",
	}
	if got.State != "rejected" || len(got.Files) != 1 || !maps.Equal(got.Files[0].Verdict, corrupt) {
		t.Errorf("state %q files %+v, want rejected with the corrupt verdict %v", got.State, got.Files, corrupt)
	}
	if n := len(got.Events); n != 2 || got.Events[1].Kind != "rejected" || !maps.Equal(got.Events[1].Reject, corrupt) {
		t.Errorf("events = %+v, want uploading then rejected carrying the corrupt reason", got.Events)
	}
	if len(got.Warnings) != 1 || !maps.Equal(got.Warnings[0].Reason, extras) || got.Warnings[0].Subject != "scan.jpg" {
		t.Errorf("warnings = %+v, want extras-dropped naming scan.jpg", got.Warnings)
	}
}
