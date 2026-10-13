package intake

import (
	"net/http"
	"slices"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
)

// uploadWhole sends data as one file of a created submission in its
// tus creation request, and fails the test unless it answers 201.
func (f *tusFixture) uploadWhole(t *testing.T, m identity.UserID, submission, file string, data []byte) {
	t.Helper()
	reply := f.do(t, tusRequest{method: http.MethodPost, path: tusPrefix + "/tus/", member: &m, body: data, header: map[string]string{
		"Tus-Resumable":   "1.0.0",
		"Upload-Length":   strconv.Itoa(len(data)),
		"Upload-Metadata": metadata("submission", submission, "file", file),
		"Content-Type":    "application/offset+octet-stream",
	}})
	mustStatus(t, "create with upload", reply, http.StatusCreated)
}

// submitThroughTus creates a submission through the seam, sends each
// file through the tus mount, and returns the created ids.
func (f *tusFixture) submitThroughTus(t *testing.T, seam *Seam, m identity.UserID, req ops.CreateSubmissionRequest, data ...[]byte) ops.CreateSubmissionResponse {
	t.Helper()
	resp := mustCreateVia(t, seam, m, req)
	for i, file := range resp.Files {
		f.uploadWhole(t, m, resp.SubmissionID, file.FileID, data[i])
	}
	return resp
}

// tracksRequest declares separate tracks named 01.flac, 02.flac, ...
func tracksRequest(data ...[]byte) ops.CreateSubmissionRequest {
	req := ops.CreateSubmissionRequest{VerdictUnit: "track", Container: "files", Source: "bought"}
	for i, d := range data {
		req.Files = append(req.Files, ops.ManifestFile{Path: "0" + strconv.Itoa(i+1) + ".flac", Size: int64(len(d))})
	}
	return req
}

// awaitVerdict polls get_submission until the worker has judged id.
func (f *tusFixture) awaitVerdict(t *testing.T, seam *Seam, m identity.UserID, id string) ops.GetSubmissionResponse {
	t.Helper()
	var got ops.GetSubmissionResponse
	waitFor(t, func() bool {
		var err error
		got, err = seam.GetSubmission(deadline(t), actorOf(m), ops.GetSubmissionRequest{SubmissionID: id})
		if err != nil {
			t.Fatal(err)
		}
		return got.State != "uploading" && got.State != "checking"
	})
	return got
}

// collectionName is the name of the collection a submission landed in.
func collectionName(t *testing.T, f *tusFixture, id string) string {
	t.Helper()
	var name string
	err := f.h.QueryRow(deadline(t), `SELECT c.name FROM submissions s JOIN collections c ON c.id = s.collection_id WHERE s.id = ?`,
		mustParse(t, id)).Scan(&name)
	if err != nil {
		t.Fatal(err)
	}
	return name
}

// TestSeparateTracksEndToEnd is AC 22 through every piece this pass
// builds: a create through the seam, each file through the tus mount,
// receipt waking the started worker over the one shared writer, the
// verdict and cleanup, and the result read through get_submission.
func TestSeparateTracksEndToEnd(t *testing.T) {
	f := newTusFixture(t)
	seam := f.in.Seam()
	m := member(t, f.h)
	if err := f.in.Start(t.Context()); err != nil {
		t.Fatal(err)
	}

	t.Run("one corrupt of three", func(t *testing.T) {
		data := [][]byte{cdTrack(31, 1).encode(t), flipped(cdTrack(32, 2).encode(t)), cdTrack(33, 3).encode(t)}
		resp := f.submitThroughTus(t, seam, m, tracksRequest(data...), data...)
		got := f.awaitVerdict(t, seam, m, resp.SubmissionID)

		if got.State != "accepted" {
			t.Fatalf("state = %q, want accepted", got.State)
		}
		if name := collectionName(t, f, resp.SubmissionID); name != "Unlisted music" {
			t.Errorf("collection = %q, want Unlisted music", name)
		}
		for _, ev := range got.Events {
			if ev.Kind == "matching" {
				t.Error("separate tracks passed through matching")
			}
		}
		for i, file := range got.Files {
			corrupt := file.Path == "02.flac"
			if corrupt != (file.Verdict != nil && file.Verdict.Code == "corrupt") {
				t.Errorf("%s: verdict %+v", file.Path, file.Verdict)
			}
			id := mustParse(t, resp.Files[i].FileID)
			if corrupt {
				waitFor(t, func() bool { return f.tusFile(t, id) == nil && !f.infoExists(t, id) })
			} else if f.tusFile(t, id) == nil {
				t.Errorf("%s: a passing track's bytes are gone", file.Path)
			}
		}
	})

	t.Run("three failing", func(t *testing.T) {
		data := [][]byte{flipped(cdTrack(41, 1).encode(t)), flipped(cdTrack(42, 2).encode(t)), flipped(cdTrack(43, 3).encode(t))}
		resp := f.submitThroughTus(t, seam, m, tracksRequest(data...), data...)
		if got := f.awaitVerdict(t, seam, m, resp.SubmissionID); got.State != "rejected" {
			t.Errorf("state = %q, want rejected", got.State)
		}
	})

	t.Run("a ZIP of two tracks and a scan", func(t *testing.T) {
		archive := zipOfMembers(t, zipMember{"01.flac", cdTrack(51, 1).encode(t)}, zipMember{"02.flac", cdTrack(52, 2).encode(t)},
			zipMember{"scan.jpg", []byte("jpeg bytes")})
		req := ops.CreateSubmissionRequest{VerdictUnit: "track", Container: "zip", Source: "cd",
			Files: []ops.ManifestFile{{Path: "music.zip", Size: int64(len(archive))}}}
		resp := f.submitThroughTus(t, seam, m, req, archive)
		got := f.awaitVerdict(t, seam, m, resp.SubmissionID)

		if got.State != "accepted" {
			t.Fatalf("state = %q, want accepted", got.State)
		}
		var paths []string
		for _, file := range got.Files {
			paths = append(paths, file.Path)
		}
		if !slices.Equal(paths, []string{"01.flac", "02.flac", "music.zip"}) {
			t.Errorf("files = %q, want the two tracks and the archive", paths)
		}
		if len(got.Warnings) != 1 || got.Warnings[0].Reason.Code != "extras-dropped" ||
			got.Warnings[0].Subject == nil || *got.Warnings[0].Subject != "scan.jpg" {
			t.Errorf("warnings = %+v, want extras-dropped naming scan.jpg", got.Warnings)
		}
	})
	assertStatesMatchEvents(t, f.h)
}
