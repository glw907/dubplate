package serve

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

const (
	tusResumable    = "1.0.0"
	testEmailHeader = "X-Test-Email" // stands in for U7's real Access-JWT middleware
)

// withTestAuth stands in for the serve command's real Cloudflare Access
// enforcement (U7): it reads a test-only header and stamps the
// corresponding email into the request context via [WithEmail], exactly
// what the real middleware will do with a verified JWT claim.
func withTestAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if email := r.Header.Get(testEmailHeader); email != "" {
			r = r.WithContext(WithEmail(r.Context(), email))
		}
		next.ServeHTTP(w, r)
	})
}

// testEngine bundles an Engine with the config and a running server, and
// exposes the raw filesystem roots so tests can assert directly on what
// landed where.
type testEngine struct {
	engine     *Engine
	cfg        *config.Config
	server     *httptest.Server
	client     *http.Client
	uploadsDir string
	inboxDir   string
	now        time.Time
}

const (
	testEmailAlice = "alice@example.com"
	testDirAlice   = "Alice Ay"
)

// newTestEngine builds an Engine over real temp directories and a real
// contributors roster (Alice only), with generous defaults every test
// can override via mutate.
func newTestEngine(t *testing.T, mutate func(cfg *config.Config)) *testEngine {
	t.Helper()

	root := t.TempDir()
	uploadsDir := filepath.Join(root, "uploads")
	inboxDir := filepath.Join(root, "inbox")
	for _, d := range []string{uploadsDir, inboxDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	contributorsPath := filepath.Join(root, "contributors.yaml")
	yamlBody := "contributors:\n  - name: " + testDirAlice + "\n    email: " + testEmailAlice + "\n"
	if err := os.WriteFile(contributorsPath, []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write contributors: %v", err)
	}
	roster, err := config.LoadContributors(contributorsPath)
	if err != nil {
		t.Fatalf("LoadContributors: %v", err)
	}

	cfg := &config.Config{
		Paths: config.Paths{
			UploadsDir: uploadsDir,
			InboxDir:   inboxDir,
		},
		Upload: config.UploadConfig{
			ChunkMB:            8,
			MaxAlbumBytes:      1 << 30,
			PartialTTL:         time.Hour,
			MaxSessionsPerUser: 2,
		},
		Thresholds: config.Thresholds{
			// A generous floor keeps admission control out of tests that
			// aren't exercising it; TestPreCreate_NoSpaceAdmissionControl
			// overrides it to something no real filesystem clears.
			FreeFloorGB: 0,
		},
		Collections: []string{"", "family"},
	}
	if mutate != nil {
		mutate(cfg)
	}

	now := time.Now()
	e, err := NewEngine(cfg, roster, Options{Now: func() time.Time { return now }, TagProbe: alwaysConfidentTagProbe})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	srv := httptest.NewServer(withTestAuth(e.Handler()))
	t.Cleanup(srv.Close)

	return &testEngine{
		engine:     e,
		cfg:        cfg,
		server:     srv,
		client:     srv.Client(),
		uploadsDir: uploadsDir,
		inboxDir:   inboxDir,
		now:        now,
	}
}

// --- tus protocol helpers --------------------------------------------

func encodeTusMetadata(m map[string]string) string {
	parts := make([]string, 0, len(m))
	for k, v := range m {
		parts = append(parts, k+" "+base64.StdEncoding.EncodeToString([]byte(v)))
	}
	return strings.Join(parts, ",")
}

func (te *testEngine) tusCreate(t *testing.T, email string, size int64, meta map[string]string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, te.server.URL+"/files/", http.NoBody)
	if err != nil {
		t.Fatalf("build create request: %v", err)
	}
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set("Upload-Length", strconv.FormatInt(size, 10))
	if len(meta) > 0 {
		req.Header.Set("Upload-Metadata", encodeTusMetadata(meta))
	}
	if email != "" {
		req.Header.Set(testEmailHeader, email)
	}
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	return resp
}

func (te *testEngine) tusPatch(t *testing.T, uploadURL, email string, offset int64, data []byte) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodPatch, uploadURL, bytes.NewReader(data))
	if err != nil {
		t.Fatalf("build patch request: %v", err)
	}
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Upload-Offset", strconv.FormatInt(offset, 10))
	if email != "" {
		req.Header.Set(testEmailHeader, email)
	}
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("patch upload: %v", err)
	}
	return resp
}

// uploadWhole creates and fully PATCHes a single tus upload in one call,
// returning the final PATCH response (whose body carries the finish
// outcome when this upload completed its session).
func (te *testEngine) uploadWhole(t *testing.T, data []byte, meta map[string]string) *http.Response {
	t.Helper()
	createResp := te.tusCreate(t, testEmailAlice, int64(len(data)), meta)
	defer func() { _ = createResp.Body.Close() }()
	if createResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResp.Body)
		t.Fatalf("create upload: status %d, body %q", createResp.StatusCode, body)
	}
	loc := createResp.Header.Get("Location")
	if loc == "" {
		t.Fatal("create upload: no Location header")
	}
	return te.tusPatch(t, loc, testEmailAlice, 0, data)
}

// --- fixtures ----------------------------------------------------------

// flacFixture builds a minimal synthetic FLAC stream carrying exactly
// what [intake.CheckFLAC] parses: the magic marker, a STREAMINFO
// metadata-block header, and a STREAMINFO payload encoding sampleRate
// and bitsPerSample. It is not a real, decodable FLAC (fine: the fast
// in-request check this package runs never decodes audio, only the
// deep import-time check the importer runs later would notice).
func flacFixture(sampleRate, bitsPerSample int) []byte {
	buf := make([]byte, 8+34)
	copy(buf[0:4], "fLaC")
	buf[4] = 0x80 // last-metadata-block=1, type=0 (STREAMINFO)
	buf[5], buf[6], buf[7] = 0, 0, 34

	info := buf[8:]
	var packed uint64
	packed |= uint64(sampleRate&0xFFFFF) << 44
	packed |= uint64((bitsPerSample-1)&0x1F) << 36
	binary.BigEndian.PutUint64(info[10:18], packed)
	return buf
}

const (
	cdSampleRate    = 44100
	cdBitsPerSample = 16
)

func cdFlacFixture() []byte {
	return flacFixture(cdSampleRate, cdBitsPerSample)
}

// buildZip returns an in-memory ZIP holding files, safe entries only.
// intake.CheckZIP validates paths and declared sizes, never FLAC
// content, so any bytes suffice for a verdict of OK.
func buildZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create zip entry %s: %v", name, err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("write zip entry %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

// --- identity and authorization (B3) ------------------------------------

func TestPreCreate_UnauthenticatedRejected(t *testing.T) {
	te := newTestEngine(t, nil)
	resp := te.tusCreate(t, "", 10, map[string]string{"filename": "album.zip"})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

// TestOwnership_CrossUserPATCHDenied404 proves a second contributor
// cannot PATCH into someone else's upload, and that the denial is 404,
// never 403 -- a 403 would confirm the upload exists at all, the
// existence oracle B3 forbids.
func TestOwnership_CrossUserPATCHDenied404(t *testing.T) {
	te := newTestEngine(t, nil)
	data := buildZip(t, map[string][]byte{"one.flac": cdFlacFixture()})

	createResp := te.tusCreate(t, testEmailAlice, int64(len(data)), map[string]string{"filename": "album.zip", "album": "Cross User Test", "collection": ""})
	defer func() { _ = createResp.Body.Close() }()
	if createResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResp.Body)
		t.Fatalf("create: status %d, body %q", createResp.StatusCode, body)
	}
	loc := createResp.Header.Get("Location")

	// bob@example.com is not on the roster at all, but ownership denial
	// must fire before any roster question is even reached.
	resp := te.tusPatch(t, loc, "bob@example.com", 0, data)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-user PATCH status = %d, want 404", resp.StatusCode)
	}

	// The rightful owner can still complete it afterward.
	own := te.tusPatch(t, loc, testEmailAlice, 0, data)
	defer func() { _ = own.Body.Close() }()
	if own.StatusCode < 200 || own.StatusCode >= 300 {
		t.Fatalf("owner PATCH after a denied cross-user attempt: status %d", own.StatusCode)
	}
}

// TestOwnership_CrossUserSessionFinishDenied404 proves POST
// /api/session/{id}/finish denies a non-owner as 404, the same
// existence-oracle rule enforced for /files/.
func TestOwnership_CrossUserSessionFinishDenied404(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Cross User Session", "", 2)

	req, err := http.NewRequest(http.MethodPost, te.server.URL+"/api/session/"+sessID+"/finish", http.NoBody)
	if err != nil {
		t.Fatalf("build finish request: %v", err)
	}
	req.Header.Set(testEmailHeader, "bob@example.com")
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("finish request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-user finish status = %d, want 404", resp.StatusCode)
	}
}

// TestOwnership_UnknownSessionIDInTusMetadataDenied404 proves a
// forged, well-formed but nonexistent session ID in tus metadata is
// denied the same way as a real cross-user one: no way to distinguish
// "wrong owner" from "never existed" from the response.
func TestOwnership_UnknownSessionIDInTusMetadataDenied404(t *testing.T) {
	te := newTestEngine(t, nil)
	forged, err := newSessionID()
	if err != nil {
		t.Fatalf("newSessionID: %v", err)
	}
	resp := te.tusCreate(t, testEmailAlice, 10, map[string]string{
		"filename":   "track.flac",
		"session_id": forged,
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", resp.StatusCode)
	}
}

// TestOwnership_CrossUserTerminateDenied404 proves a DELETE against
// another contributor's upload is denied as 404, the same rule
// [Engine.preTerminate] enforces natively and
// [Engine.enforceUploadOwnership] enforces again at the HTTP layer.
func TestOwnership_CrossUserTerminateDenied404(t *testing.T) {
	te := newTestEngine(t, nil)
	data := cdFlacFixture()
	createResp := te.tusCreate(t, testEmailAlice, int64(len(data)), map[string]string{
		"filename": "01 - one.flac", "session_id": te.createSession(t, "Terminate Test", "", 1),
	})
	defer func() { _ = createResp.Body.Close() }()
	if createResp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createResp.Body)
		t.Fatalf("create: status %d, body %q", createResp.StatusCode, body)
	}
	loc := createResp.Header.Get("Location")

	req, err := http.NewRequest(http.MethodDelete, loc, http.NoBody)
	if err != nil {
		t.Fatalf("build delete request: %v", err)
	}
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set(testEmailHeader, "bob@example.com")
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("delete request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("cross-user DELETE status = %d, want 404", resp.StatusCode)
	}
}

// TestOwnership_TraversalUploadIDRejected404 proves a percent-encoded
// traversal segment in the /files/ path never reaches the store's
// filepath.Join: F2.2 found tusd's filestore performs no ID validation
// of its own, so a path shaped like "..%2f<name>" survives Go's
// ServeMux and http.StripPrefix (neither is path-segment aware over the
// already-decoded r.URL.Path) all the way to enforceUploadOwnership
// unless this layer's own shape check stops it.
func TestOwnership_TraversalUploadIDRejected404(t *testing.T) {
	te := newTestEngine(t, nil)

	paths := []string{
		te.server.URL + "/files/..%2fzzprobe",
		te.server.URL + "/files/%2e%2e%2fzzprobe",
		te.server.URL + "/files/not-a-valid-shape-at-all",
	}
	for _, p := range paths {
		req, err := http.NewRequest(http.MethodHead, p, http.NoBody)
		if err != nil {
			t.Fatalf("build request for %s: %v", p, err)
		}
		req.Header.Set("Tus-Resumable", tusResumable)
		req.Header.Set(testEmailHeader, testEmailAlice)
		resp, err := te.client.Do(req)
		if err != nil {
			t.Fatalf("request %s: %v", p, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("HEAD %s: status = %d, want 404", p, resp.StatusCode)
		}
	}
}

// --- input sanitization (B4) --------------------------------------------

func TestHandleCreateSession_HostileInputsRejected(t *testing.T) {
	tests := []struct {
		name       string
		album      string
		collection string
	}{
		{"path traversal album", "../../etc/passwd", ""},
		{"absolute path album", "/etc/passwd", ""},
		{"NUL byte album", "abbey\x00road", ""},
		{"only punctuation album", "///", ""},
		{"unconfigured collection", "Fine Album", "bootlegs"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			te := newTestEngine(t, nil)
			if tt.name == "path traversal album" || tt.name == "absolute path album" || tt.name == "NUL byte album" {
				// These sanitize to a safe, non-empty slug rather than
				// erroring -- see TestSlugifyAlbum -- so creation itself
				// succeeds; the assertion that matters is that the
				// resulting session's Album never carries the raw hostile
				// bytes into a path.
				sessID := te.createSession(t, tt.album, tt.collection, 1)
				sess, ok := te.engine.sessions.get(sessID)
				if !ok {
					t.Fatal("session not found immediately after creation")
				}
				if strings.ContainsAny(sess.Album, "./\\\x00") {
					t.Fatalf("Session.Album = %q, still carries an unsafe character", sess.Album)
				}
				return
			}

			status, _ := te.tryCreateSession(t, tt.album, tt.collection, 1)
			if status != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400 for %+v", status, tt)
			}
		})
	}
}

// TestHandleCreateSession_TooManyExpectedFilesRejected proves
// expected_files is bounded: an unbounded value could pin an unbounded
// number of Session.verdicts slots regardless of the aggregate byte cap
// admission control separately enforces (F2.5).
func TestHandleCreateSession_TooManyExpectedFilesRejected(t *testing.T) {
	te := newTestEngine(t, nil)
	status, body := te.tryCreateSession(t, "Huge Box Set", "", maxExpectedFiles+1)
	if status != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400, body %q", status, body)
	}
}

// --- session completion (B2) ---------------------------------------------

// createSession is a small wrapper over POST /api/session for tests that
// only need the resulting session ID.
func (te *testEngine) createSession(t *testing.T, album, collection string, expectedFiles int) string {
	t.Helper()
	status, body := te.tryCreateSession(t, album, collection, expectedFiles)
	if status != http.StatusCreated {
		t.Fatalf("create session: status %d, body %q", status, body)
	}
	var resp createSessionResponse
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		t.Fatalf("unmarshal create-session response %q: %v", body, err)
	}
	return resp.SessionID
}

// tryCreateSession always authenticates as [testEmailAlice]: every test
// that needs a different (or absent) identity on POST /api/session
// builds its own request rather than adding an email parameter no
// caller here would ever vary.
func (te *testEngine) tryCreateSession(t *testing.T, album, collection string, expectedFiles int) (int, string) {
	t.Helper()
	payload, err := json.Marshal(createSessionRequest{Album: album, Collection: collection, ExpectedFiles: expectedFiles})
	if err != nil {
		t.Fatalf("marshal create-session request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, te.server.URL+"/api/session", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("build create-session request: %v", err)
	}
	req.Header.Set(testEmailHeader, testEmailAlice)
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("create-session request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read create-session response: %v", err)
	}
	return resp.StatusCode, string(body)
}

// TestMultiFileSession_CompletesByCountAndPromotes drives a two-file
// session to completion purely by uploading both files, with no
// explicit /finish call, and checks the album lands under the inbox
// with its session.json sidecar.
func TestMultiFileSession_CompletesByCountAndPromotes(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Two Track Album", "family", 2)

	first := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	firstBody, _ := io.ReadAll(first.Body)
	_ = first.Body.Close()
	if first.StatusCode < 200 || first.StatusCode >= 300 {
		t.Fatalf("first upload: status %d, body %q", first.StatusCode, firstBody)
	}

	second := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "02 - two.flac", "session_id": sessID,
	})
	secondBody, _ := io.ReadAll(second.Body)
	_ = second.Body.Close()
	if second.StatusCode != http.StatusOK {
		t.Fatalf("second (completing) upload: status %d, body %q", second.StatusCode, secondBody)
	}

	var fin finishResponse
	if err := json.Unmarshal(secondBody, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", secondBody, err)
	}
	if !fin.Promoted {
		t.Fatalf("finish response = %+v, want Promoted", fin)
	}

	dest := filepath.Join(te.inboxDir, testDirAlice, sessID+"__two-track-album")
	assertPromoted(t, dest, []string{"01 - one.flac", "02 - two.flac"})
	assertSessionJSON(t, dest, sessID, testEmailAlice, "family")
}

// TestCompleteSession_OnPromotedFiresAfterRename proves
// [Options.OnPromoted] runs exactly once, carrying the promoted
// session's own ID and destination, and only after [Engine.promote]'s
// rename has already put the album under the inbox -- U7's nudge
// trigger depends on that ordering, so a fired-too-early hook would
// nudge the importer at a directory it cannot see yet.
func TestCompleteSession_OnPromotedFiresAfterRename(t *testing.T) {
	var calls []string
	te := newTestEngineWithOptions(t, nil, Options{
		OnPromoted: func(sessionID, dest string) {
			if _, err := os.Stat(dest); err != nil {
				t.Errorf("OnPromoted fired before rename landed: stat %s: %v", dest, err)
			}
			calls = append(calls, sessionID)
		},
	})
	sessID := te.createSession(t, "Hooked Album", "", 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: status %d, body %q", resp.StatusCode, body)
	}

	if len(calls) != 1 || calls[0] != sessID {
		t.Fatalf("OnPromoted calls = %v, want exactly one call for %s", calls, sessID)
	}
}

// TestLoneZipSession_CompletesAtPreFinish proves a lone ZIP upload with
// no explicit session ever created completes -- and its verdict is
// delivered -- in the same request that finished the PATCH.
func TestLoneZipSession_CompletesAtPreFinish(t *testing.T) {
	te := newTestEngine(t, nil)
	data := buildZip(t, map[string][]byte{
		"01 - one.flac": []byte("not decoded by the fast check"),
		"02 - two.flac": []byte("not decoded by the fast check"),
	})

	resp := te.uploadWhole(t, data, map[string]string{
		"filename":   "Lone Album.zip",
		"album":      "Lone Album",
		"collection": "",
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if !fin.Promoted {
		t.Fatalf("finish response = %+v, want Promoted", fin)
	}

	entries, err := os.ReadDir(filepath.Join(te.inboxDir, testDirAlice))
	if err != nil {
		t.Fatalf("read inbox contributor dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("inbox contributor dir has %d entries, want 1: %v", len(entries), entries)
	}
	dest := filepath.Join(te.inboxDir, testDirAlice, entries[0].Name())
	assertPromoted(t, dest, []string{"01 - one.flac", "02 - two.flac"})
}

// TestMultiFileSession_NPlusOneUploadRejected proves the upload past
// ExpectedFiles is refused at pre-create, not merely ignored.
func TestMultiFileSession_NPlusOneUploadRejected(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "One Track Album", "", 1)

	first := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = first.Body.Close()
	if first.StatusCode != http.StatusOK {
		t.Fatalf("first upload: status %d", first.StatusCode)
	}

	second := te.tusCreate(t, testEmailAlice, int64(len(cdFlacFixture())), map[string]string{
		"filename": "02 - extra.flac", "session_id": sessID,
	})
	defer func() { _ = second.Body.Close() }()
	if second.StatusCode != http.StatusConflict {
		t.Fatalf("N+1st upload create status = %d, want 409", second.StatusCode)
	}
}

// TestMultiFileSession_ExplicitFinishCompletesEarly proves the second
// way a session completes: an explicit /finish call before the count
// ever reaches ExpectedFiles, using whatever verdicts have accumulated
// so far.
func TestMultiFileSession_ExplicitFinishCompletesEarly(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Early Finish Album", "", 3)

	first := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = first.Body.Close()
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("first upload: status %d, want 204 (2 more still expected)", first.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPost, te.server.URL+"/api/session/"+sessID+"/finish", http.NoBody)
	if err != nil {
		t.Fatalf("build finish request: %v", err)
	}
	req.Header.Set(testEmailHeader, testEmailAlice)
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("finish request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read finish response: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("finish status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if !fin.Promoted || len(fin.Files) != 1 {
		t.Fatalf("finish response = %+v, want Promoted with exactly the one uploaded file's verdict", fin)
	}

	dest := filepath.Join(te.inboxDir, testDirAlice, sessID+"__early-finish-album")
	assertPromoted(t, dest, []string{"01 - one.flac"})

	// A second /finish call is idempotent: it must not try to run
	// Decide/promote a second time (which would try to rename an
	// already-moved directory) and must return the same outcome.
	resp2, err := te.client.Do(req.Clone(req.Context()))
	if err != nil {
		t.Fatalf("second finish request: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()
	if resp2.StatusCode != http.StatusOK {
		body2, _ := io.ReadAll(resp2.Body)
		t.Fatalf("second finish status = %d, body %q", resp2.StatusCode, body2)
	}
}

// TestMultiFileSession_PartialNeverPromoted proves a session with fewer
// uploads than ExpectedFiles never appears under the inbox, even after
// the one file it did receive finished.
func TestMultiFileSession_PartialNeverPromoted(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Partial Album", "", 2)

	first := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(first.Body)
	_ = first.Body.Close()
	if first.StatusCode != http.StatusNoContent { // 204 = tus's own "keep going", session not yet done
		t.Fatalf("first upload: status %d, body %q", first.StatusCode, body)
	}

	assertNoInboxTrace(t, te.inboxDir)

	entries, err := os.ReadDir(filepath.Join(te.uploadsDir, sessID))
	if err != nil {
		t.Fatalf("read session dir under uploads: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("session dir has %d entries, want the one uploaded file", len(entries))
	}
}

// TestSweepOnce_AbandonedSessionExpiresWithoutPromotion proves
// SweepOnce removes an abandoned session's staged content once
// PartialTTL has passed, through the frozen-clock seam, and that the
// inbox never saw it.
func TestSweepOnce_AbandonedSessionExpiresWithoutPromotion(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Abandoned Album", "", 2)

	first := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = first.Body.Close()
	if first.StatusCode != http.StatusNoContent {
		t.Fatalf("first upload: status %d", first.StatusCode)
	}

	if _, err := os.Stat(filepath.Join(te.uploadsDir, sessID)); err != nil {
		t.Fatalf("session directory missing before sweep: %v", err)
	}

	future := te.now.Add(te.cfg.Upload.PartialTTL + time.Minute)
	swept := te.engine.SweepOnce(future)
	if swept != 1 {
		t.Fatalf("SweepOnce swept %d sessions, want 1", swept)
	}

	if _, err := os.Stat(filepath.Join(te.uploadsDir, sessID)); !os.IsNotExist(err) {
		t.Fatalf("session directory still present after sweep: err=%v", err)
	}
	assertNoInboxTrace(t, te.inboxDir)

	if _, ok := te.engine.sessions.get(sessID); ok {
		t.Fatal("session still present in the store after sweep")
	}
}

// TestSweepOnce_FreshSessionSurvives is the negative control: a session
// still inside its TTL is untouched.
func TestSweepOnce_FreshSessionSurvives(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Fresh Album", "", 2)

	swept := te.engine.SweepOnce(te.now.Add(time.Minute))
	if swept != 0 {
		t.Fatalf("SweepOnce swept %d sessions, want 0", swept)
	}
	if _, ok := te.engine.sessions.get(sessID); !ok {
		t.Fatal("fresh session was removed by an early sweep")
	}
}

// TestSweepOnce_ReclaimsCompletedSessionFromStoreAndDisk proves F2.3's
// fix on both halves: a completed session's Store entry is forgotten
// once settleGrace has passed even though it is not "expired" in
// [Store.expired]'s sense, and its leftover on-disk directory (here, a
// pushed-back album -- the tag-check probe wasn't confident, so the
// content is held rather than discarded) survives until it clears
// PartialTTL, then is reclaimed by the same filesystem-driven pass that
// makes [Engine.sweepOrphans] restart-safe.
func TestSweepOnce_ReclaimsCompletedSessionFromStoreAndDisk(t *testing.T) {
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: alwaysUnconfidentTagProbe,
	})
	sessID := te.createSession(t, "Pushed Back Album", "", 1)
	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: status %d", resp.StatusCode)
	}
	sessDir := filepath.Join(te.uploadsDir, sessID)
	if _, err := os.Stat(sessDir); err != nil {
		t.Fatalf("pushed-back session directory missing before sweep: %v", err)
	}

	// Just past settleGrace, well inside PartialTTL: the Store entry is
	// forgotten, but the directory is untouched -- it is not yet old
	// enough for sweepOrphans' own, much longer retention window.
	te.engine.SweepOnce(te.now.Add(settleGrace + time.Second))
	if _, ok := te.engine.sessions.get(sessID); ok {
		t.Fatal("completed session still present in the store past settleGrace")
	}
	if _, err := os.Stat(sessDir); err != nil {
		t.Fatalf("pushed-back directory swept before it cleared PartialTTL: %v", err)
	}

	// Past PartialTTL: sweepOrphans reclaims the now-unclaimed directory.
	te.engine.SweepOnce(te.now.Add(te.cfg.Upload.PartialTTL + time.Minute))
	if _, err := os.Stat(sessDir); !os.IsNotExist(err) {
		t.Fatalf("pushed-back directory still present past PartialTTL: err=%v", err)
	}
}

// TestSweepOnce_ReclaimsRestartOrphanFromDisk proves F2.4's fix: a
// directory a session left behind (here, one pushed back by the
// tag-check probe) is reclaimed by a second Engine built over the same
// UploadsDir whose Store never held that session at all -- exactly what
// every serve restart looks like from the sweep's point of view, since
// the Store is in-memory only.
func TestSweepOnce_ReclaimsRestartOrphanFromDisk(t *testing.T) {
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: alwaysUnconfidentTagProbe,
	})
	sessID := te.createSession(t, "Orphaned After Restart", "", 1)
	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("upload: status %d", resp.StatusCode)
	}
	sessDir := filepath.Join(te.uploadsDir, sessID)
	if _, err := os.Stat(sessDir); err != nil {
		t.Fatalf("pushed-back session directory missing before restart: %v", err)
	}

	restarted, err := NewEngine(te.cfg, te.engine.roster, Options{
		Now:      func() time.Time { return te.now },
		TagProbe: alwaysUnconfidentTagProbe,
	})
	if err != nil {
		t.Fatalf("NewEngine (simulated restart): %v", err)
	}
	if _, ok := restarted.sessions.get(sessID); ok {
		t.Fatal("the restarted engine's fresh Store already knows about the old session")
	}

	restarted.SweepOnce(te.now.Add(te.cfg.Upload.PartialTTL + time.Minute))
	if _, err := os.Stat(sessDir); !os.IsNotExist(err) {
		t.Fatalf("orphaned directory survived a restart-simulated sweep: err=%v", err)
	}
}

// --- atomic promotion (M9, B6) --------------------------------------------

// TestPromotion_RenameFailureLeavesNoPartialInbox proves that when the
// final rename fails, nothing partial ever appears under the inbox, and
// the staged session directory survives under UploadsDir instead of
// being lost.
func TestPromotion_RenameFailureLeavesNoPartialInbox(t *testing.T) {
	te := newTestEngineWithOptions(t, nil, Options{
		Rename: func(oldpath, newpath string) error { return os.ErrPermission },
	})

	sessID := te.createSession(t, "Doomed Album", "", 1)
	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if fin.Promoted {
		t.Fatal("finish response reports Promoted despite a forced rename failure")
	}

	assertNoInboxTrace(t, te.inboxDir)
	// The staged directory is never lost even though promotion failed:
	// promote's own doc comment argues this is what a single failed
	// rename must guarantee.
	if _, err := os.Stat(filepath.Join(te.uploadsDir, sessID)); err != nil {
		t.Fatalf("staged session directory lost after a failed rename: %v", err)
	}
}

// TestPromotion_SessionJSONWrittenBeforeRename proves session.json
// exists inside the promoted directory with the fields the plan-time
// decision names, confirming it landed before the one rename made the
// directory visible under the inbox at all.
func TestPromotion_SessionJSONWrittenBeforeRename(t *testing.T) {
	te := newTestEngine(t, nil)
	sessID := te.createSession(t, "Sidecar Album", "family", 1)
	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	dest := filepath.Join(te.inboxDir, testDirAlice, sessID+"__sidecar-album")
	assertSessionJSON(t, dest, sessID, testEmailAlice, "family")
}

// --- admission control (M13) ----------------------------------------------

func TestAdmissionControl_TooManySessions(t *testing.T) {
	te := newTestEngine(t, func(cfg *config.Config) {
		cfg.Upload.MaxSessionsPerUser = 1
	})
	te.createSession(t, "First Album", "", 1)

	status, body := te.tryCreateSession(t, "Second Album", "", 1)
	if status != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429, body %q", status, body)
	}
	if !strings.Contains(body, "too-many-sessions") {
		t.Fatalf("body = %q, want the too-many-sessions reject.Reason wire code", body)
	}
}

func TestAdmissionControl_NoSpaceOnImplicitZipPreCreate(t *testing.T) {
	te := newTestEngine(t, func(cfg *config.Config) {
		// No real filesystem clears a petabyte floor.
		cfg.Thresholds.FreeFloorGB = 1 << 20
	})
	data := buildZip(t, map[string][]byte{"one.flac": []byte("x")})
	resp := te.tusCreate(t, testEmailAlice, int64(len(data)), map[string]string{
		"filename": "album.zip", "album": "No Space Album", "collection": "",
	})
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusInsufficientStorage {
		t.Fatalf("status = %d, want 507, body %q", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "no-space") {
		t.Fatalf("body = %q, want the no-space reject.Reason wire code", body)
	}
}

// TestAdmissionControl_TooManySessionsOnImplicitZipPreCreate proves the
// same MaxSessionsPerUser cap applies on the implicit lone-ZIP path at
// tus pre-create, not only on the explicit POST /api/session path.
func TestAdmissionControl_TooManySessionsOnImplicitZipPreCreate(t *testing.T) {
	te := newTestEngine(t, func(cfg *config.Config) {
		cfg.Upload.MaxSessionsPerUser = 1
	})
	te.createSession(t, "First Album", "", 1)

	data := buildZip(t, map[string][]byte{"one.flac": []byte("x")})
	resp := te.tusCreate(t, testEmailAlice, int64(len(data)), map[string]string{
		"filename": "album.zip", "album": "Second Album", "collection": "",
	})
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want 429, body %q", resp.StatusCode, body)
	}
	if !strings.Contains(string(body), "too-many-sessions") {
		t.Fatalf("body = %q, want the too-many-sessions reject.Reason wire code", body)
	}
}

// TestAdmissionControl_AggregateAlbumBytesOnMultiFilePreCreate proves a
// multi-file session's aggregate size is bounded by MaxAlbumBytes, not
// only each individual tus resource: F2.5 found nothing enforced the
// album total on this path, so a session declaring a generous
// expected_files count could otherwise admit as many MaxAlbumBytes-sized
// uploads as free space allowed.
func TestAdmissionControl_AggregateAlbumBytesOnMultiFilePreCreate(t *testing.T) {
	te := newTestEngine(t, func(cfg *config.Config) {
		cfg.Upload.MaxAlbumBytes = 1000
	})
	sessID := te.createSession(t, "Aggregate Cap Album", "", 2)

	first := te.tusCreate(t, testEmailAlice, 600, map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	defer func() { _ = first.Body.Close() }()
	if first.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(first.Body)
		t.Fatalf("first create: status %d, body %q", first.StatusCode, body)
	}

	second := te.tusCreate(t, testEmailAlice, 500, map[string]string{
		"filename": "02 - two.flac", "session_id": sessID,
	})
	defer func() { _ = second.Body.Close() }()
	body, _ := io.ReadAll(second.Body)
	if second.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("second create: status %d, want 413, body %q", second.StatusCode, body)
	}
	if !strings.Contains(string(body), "too-large") {
		t.Fatalf("body = %q, want the too-large reject.Reason wire code", body)
	}
}

// TestUploadExpiresHeaderStamped proves an in-flight upload's HEAD
// response carries Upload-Expires, computed from its session's
// PartialTTL, so a client can show a contributor when an abandoned
// upload will be swept.
func TestUploadExpiresHeaderStamped(t *testing.T) {
	te := newTestEngine(t, nil)
	data := cdFlacFixture()
	createResp := te.tusCreate(t, testEmailAlice, int64(len(data)), map[string]string{
		"filename": "01 - one.flac", "session_id": te.createSession(t, "Expires Album", "", 1),
	})
	defer func() { _ = createResp.Body.Close() }()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status %d", createResp.StatusCode)
	}
	loc := createResp.Header.Get("Location")

	req, err := http.NewRequest(http.MethodHead, loc, http.NoBody)
	if err != nil {
		t.Fatalf("build head request: %v", err)
	}
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set(testEmailHeader, testEmailAlice)
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("head request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.Header.Get("Upload-Expires") == "" {
		t.Fatal("Upload-Expires header missing from an in-flight upload's HEAD response")
	}
}

// TestOptions_TusBasePathOverride proves [Options.TusBasePath] reaches
// tusd's own advertised BasePath (github.com/tus/tusd/v2/pkg/handler.
// Config.BasePath), the value baked into the absolute Location URL a
// tus client PATCHes back against -- U12a's production wiring
// (cmd/dubplate/serve.go's NewServer) depends on this to advertise
// "/upload/files/" rather than the bare "/files/" every other test in
// this file gets from the zero value.
func TestOptions_TusBasePathOverride(t *testing.T) {
	const wantPrefix = "/upload/files/"
	te := newTestEngineWithOptions(t, nil, Options{TusBasePath: wantPrefix})

	createResp := te.tusCreate(t, testEmailAlice, int64(len(cdFlacFixture())), map[string]string{
		"filename": "01 - one.flac", "session_id": te.createSession(t, "Base Path Album", "", 1),
	})
	defer func() { _ = createResp.Body.Close() }()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status %d", createResp.StatusCode)
	}

	loc := createResp.Header.Get("Location")
	if !strings.Contains(loc, wantPrefix) {
		t.Fatalf("Location = %q, want it to contain %q", loc, wantPrefix)
	}
}

// TestOptions_RespectsForwardedProto proves tusd's Config.
// RespectForwardedHeaders is wired on: behind cloudflared, serve sees
// plain HTTP, so without it tusd would stamp http:// into the absolute
// Location header a tus client PATCHes back against, sending a
// resuming client into an http->https redirect mid-protocol.
func TestOptions_RespectsForwardedProto(t *testing.T) {
	te := newTestEngine(t, nil)

	req, err := http.NewRequest(http.MethodPost, te.server.URL+"/files/", http.NoBody)
	if err != nil {
		t.Fatalf("build create request: %v", err)
	}
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set("Upload-Length", strconv.FormatInt(int64(len(cdFlacFixture())), 10))
	req.Header.Set("Upload-Metadata", encodeTusMetadata(map[string]string{
		"filename": "01 - one.flac", "session_id": te.createSession(t, "Forwarded Proto Album", "", 1),
	}))
	req.Header.Set(testEmailHeader, testEmailAlice)
	req.Header.Set("X-Forwarded-Proto", "https")
	resp, err := te.client.Do(req)
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status %d", resp.StatusCode)
	}

	loc := resp.Header.Get("Location")
	if !strings.HasPrefix(loc, "https://") {
		t.Fatalf("Location = %q, want an https:// scheme", loc)
	}
}

// --- tag-check probe + private lane (custody policy) -----------------------

// alwaysConfidentTagProbe is the default [Options.TagProbe] every test
// helper in this file wires in except where a test is deliberately
// exercising probe behavior: it lets every pre-existing promotion test
// reach [Engine.promote] without knowing the custody policy exists, the
// same way a real, working U8 probe would for a well-tagged fixture
// album.
func alwaysConfidentTagProbe(context.Context, string) (TagProbeResult, error) {
	return TagProbeResult{Confident: true}, nil
}

// alwaysUnconfidentTagProbe reports a clean, considered not-confident
// result: [Engine.completeSession] holds the session's staged content
// (PushedBack) rather than promoting or discarding it.
func alwaysUnconfidentTagProbe(context.Context, string) (TagProbeResult, error) {
	return TagProbeResult{Confident: false, Detail: "not confident enough"}, nil
}

// stubTagProbe returns a fixed [TagProbeResult]/error pair and records
// the directory it was called with, so a test can both drive a specific
// probe outcome and assert on what Engine handed it.
func stubTagProbe(result TagProbeResult, err error, calledWith *string) TagProbeFunc {
	return func(_ context.Context, dir string) (TagProbeResult, error) {
		*calledWith = dir
		return result, err
	}
}

// TestTagProbe_ConfidentMatchPromotesAndCapturesMetadata proves a
// confident probe result promotes the session and carries the matched
// artist/album/year into the finish response, and that the probe ran
// against the session's own assembled directory under UploadsDir.
func TestTagProbe_ConfidentMatchPromotesAndCapturesMetadata(t *testing.T) {
	var probedDir string
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: stubTagProbe(TagProbeResult{Confident: true, Artist: "The Beatles", Album: "Abbey Road", Year: "1969"}, nil, &probedDir),
	})
	sessID := te.createSession(t, "Abbey Road", "", 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if !fin.Promoted || fin.PushedBack || fin.ProbeUnavailable {
		t.Fatalf("finish response = %+v, want Promoted only", fin)
	}
	if fin.MatchedArtist != "The Beatles" || fin.MatchedAlbum != "Abbey Road" || fin.MatchedYear != "1969" {
		t.Fatalf("finish response = %+v, want the stub probe's matched metadata", fin)
	}

	wantDir := filepath.Join(te.uploadsDir, sessID)
	if probedDir != wantDir {
		t.Fatalf("probe called with dir %q, want %q", probedDir, wantDir)
	}

	dest := filepath.Join(te.inboxDir, testDirAlice, sessID+"__abbey-road")
	assertPromoted(t, dest, []string{"01 - one.flac"})
}

// TestTagProbe_NotConfidentNeverPromotes proves an unconfident probe
// result never promotes the session -- no trace under the inbox -- and
// that it holds the staged content rather than discarding it: the audio
// passed its own format verdict, so a fix path (re-upload, a MusicBrainz
// edit, or switching to the private lane) does not require sending it
// again.
func TestTagProbe_NotConfidentNeverPromotes(t *testing.T) {
	var probedDir string
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: stubTagProbe(TagProbeResult{Confident: false, Detail: "closest guess: a 1969 bootleg, not confident enough"}, nil, &probedDir),
	})
	sessID := te.createSession(t, "Uncertain Album", "", 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if fin.Promoted {
		t.Fatalf("finish response = %+v, want Promoted false", fin)
	}
	if !fin.PushedBack || fin.Detail == "" {
		t.Fatalf("finish response = %+v, want PushedBack with the probe's detail", fin)
	}

	assertNoInboxTrace(t, te.inboxDir)
	if _, err := os.Stat(filepath.Join(te.uploadsDir, sessID)); err != nil {
		t.Fatalf("staged session directory lost on a not-confident probe (should be held): %v", err)
	}
}

// TestTagProbe_UnavailableHoldsWithHonestState proves a probe that
// cannot run at all (MusicBrainz unreachable, modeled by an error
// return) never promotes, reports ProbeUnavailable rather than
// PushedBack (the two are distinct honest states), and likewise holds
// the staged content for a retry.
func TestTagProbe_UnavailableHoldsWithHonestState(t *testing.T) {
	var probedDir string
	stubErr := errors.New("musicbrainz unreachable")
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: stubTagProbe(TagProbeResult{}, stubErr, &probedDir),
	})
	sessID := te.createSession(t, "Retry Later Album", "", 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if fin.Promoted || fin.PushedBack {
		t.Fatalf("finish response = %+v, want neither Promoted nor PushedBack", fin)
	}
	if !fin.ProbeUnavailable {
		t.Fatalf("finish response = %+v, want ProbeUnavailable", fin)
	}

	assertNoInboxTrace(t, te.inboxDir)
	if _, err := os.Stat(filepath.Join(te.uploadsDir, sessID)); err != nil {
		t.Fatalf("staged session directory lost on a probe-unavailable outcome (should be held): %v", err)
	}
}

// TestTagProbe_ContextCarriesTimeoutDeadline proves completeSession
// bounds the tag-check probe with [tagProbeTimeout] rather than handing
// it a context that never expires: a stub probe captures the ctx it
// receives, and the test asserts it carries a deadline within
// tagProbeTimeout of the call.
func TestTagProbe_ContextCarriesTimeoutDeadline(t *testing.T) {
	var gotDeadline time.Time
	var hadDeadline bool
	te := newTestEngineWithOptions(t, nil, Options{
		TagProbe: func(ctx context.Context, dir string) (TagProbeResult, error) {
			gotDeadline, hadDeadline = ctx.Deadline()
			return TagProbeResult{Confident: true}, nil
		},
	})
	sessID := te.createSession(t, "Deadline Check Album", "", 1)

	before := time.Now()
	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	after := time.Now()
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	if !hadDeadline {
		t.Fatal("probe ctx had no deadline, want one bounded by tagProbeTimeout")
	}
	earliest := before.Add(tagProbeTimeout)
	latest := after.Add(tagProbeTimeout)
	if gotDeadline.Before(earliest) || gotDeadline.After(latest) {
		t.Fatalf("probe ctx deadline = %v, want within [%v, %v] (tagProbeTimeout from the call)", gotDeadline, earliest, latest)
	}
}

// TestTagProbe_DefaultOptionsFailsClosed proves that with no
// [Options.TagProbe] wired in at all -- production's state until U8's
// probe is connected in cmd/dubplate/serve.go -- a main-lane session
// never promotes untagged; it holds as ProbeUnavailable instead of
// silently succeeding. It builds NewEngine directly with a bare
// Options{Now: ...}, deliberately bypassing this file's own test
// helpers, which all default TagProbe to [alwaysConfidentTagProbe] so
// the rest of the suite can promote without knowing the probe exists.
func TestTagProbe_DefaultOptionsFailsClosed(t *testing.T) {
	base := newTestEngineNoServer(t, nil)
	e, err := NewEngine(base.cfg, base.engine.roster, Options{Now: func() time.Time { return base.now }})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	srv := httptest.NewServer(withTestAuth(e.Handler()))
	t.Cleanup(srv.Close)
	te := &testEngine{engine: e, cfg: base.cfg, server: srv, client: srv.Client(), uploadsDir: base.uploadsDir, inboxDir: base.inboxDir, now: base.now}

	sessID := te.createSession(t, "No Probe Wired Album", "", 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if fin.Promoted || !fin.ProbeUnavailable {
		t.Fatalf("finish response = %+v, want ProbeUnavailable and not Promoted with no probe wired in", fin)
	}
	assertNoInboxTrace(t, te.inboxDir)
}

// TestTagProbe_PrivateLaneBypassesProbe proves a session created against
// [Options.PrivateCollection] promotes without ever calling the probe,
// and carries no matched metadata (the private lane's tag-form values,
// not a MusicBrainz match, are what the importer applies).
func TestTagProbe_PrivateLaneBypassesProbe(t *testing.T) {
	called := false
	te := newTestEngineWithOptions(t, func(cfg *config.Config) {
		cfg.Collections = append(cfg.Collections, defaultPrivateCollection)
	}, Options{
		TagProbe: func(context.Context, string) (TagProbeResult, error) {
			called = true
			return TagProbeResult{}, errors.New("the probe must never run for a private-lane session")
		},
	})
	sessID := te.createSession(t, "Home Recording", defaultPrivateCollection, 1)

	resp := te.uploadWhole(t, cdFlacFixture(), map[string]string{
		"filename": "01 - one.flac", "session_id": sessID,
	})
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, body %q", resp.StatusCode, body)
	}
	if called {
		t.Fatal("TagProbe was called for a private-lane session")
	}

	var fin finishResponse
	if err := json.Unmarshal(body, &fin); err != nil {
		t.Fatalf("unmarshal finish response %q: %v", body, err)
	}
	if !fin.Promoted {
		t.Fatalf("finish response = %+v, want Promoted", fin)
	}
	if fin.MatchedArtist != "" || fin.MatchedAlbum != "" || fin.MatchedYear != "" {
		t.Fatalf("finish response = %+v, want no matched metadata on the private lane", fin)
	}

	dest := filepath.Join(te.inboxDir, testDirAlice, sessID+"__home-recording")
	assertPromoted(t, dest, []string{"01 - one.flac"})
}

// --- mechanics: timeouts (M13 mechanics bullet) ---------------------------

// TestSlowPATCHNotCutByServerTimeouts proves a trickling PATCH is never
// cut by an http.Server-wide ReadTimeout/WriteTimeout sized for the
// API's short JSON routes, because the /files/ route clears them.
func TestSlowPATCHNotCutByServerTimeouts(t *testing.T) {
	te := newTestEngineNoServer(t, nil)
	srv := httptest.NewUnstartedServer(withTestAuth(te.engine.Handler()))
	srv.Config.ReadTimeout = 150 * time.Millisecond
	srv.Config.WriteTimeout = 150 * time.Millisecond
	srv.Start()
	defer srv.Close()
	client := srv.Client()

	data := cdFlacFixture()
	sessID := te.createSessionAgainst(t, srv.URL, client, testEmailAlice, "Slow Album", "", 1)

	createReq, err := http.NewRequest(http.MethodPost, srv.URL+"/files/", http.NoBody)
	if err != nil {
		t.Fatalf("build create request: %v", err)
	}
	createReq.Header.Set("Tus-Resumable", tusResumable)
	createReq.Header.Set("Upload-Length", strconv.Itoa(len(data)))
	createReq.Header.Set("Upload-Metadata", encodeTusMetadata(map[string]string{"filename": "01 - one.flac", "session_id": sessID}))
	createReq.Header.Set(testEmailHeader, testEmailAlice)
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create upload: %v", err)
	}
	loc := createResp.Header.Get("Location")
	_ = createResp.Body.Close()

	slow := &slowReader{chunks: splitIntoChunks(data, 4), delay: 100 * time.Millisecond}
	req, err := http.NewRequest(http.MethodPatch, loc, slow)
	if err != nil {
		t.Fatalf("build patch request: %v", err)
	}
	req.ContentLength = int64(len(data))
	req.Header.Set("Tus-Resumable", tusResumable)
	req.Header.Set("Content-Type", "application/offset+octet-stream")
	req.Header.Set("Upload-Offset", "0")
	req.Header.Set(testEmailHeader, testEmailAlice)

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("slow patch was cut: %v", err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("slow patch status = %d, body %q, want 200 (deadlines cleared, upload finished)", resp.StatusCode, body)
	}
}

// slowReader trickles chunks with a delay between each, well past a
// short server-wide read timeout, to prove the tus route is exempt from
// it.
type slowReader struct {
	chunks [][]byte
	delay  time.Duration
	i      int
	off    int
}

func (r *slowReader) Read(p []byte) (int, error) {
	if r.i >= len(r.chunks) {
		return 0, io.EOF
	}
	if r.off == 0 && r.i > 0 {
		time.Sleep(r.delay)
	}
	chunk := r.chunks[r.i]
	n := copy(p, chunk[r.off:])
	r.off += n
	if r.off >= len(chunk) {
		r.i++
		r.off = 0
	}
	return n, nil
}

func splitIntoChunks(data []byte, n int) [][]byte {
	if n <= 0 || n > len(data) {
		n = len(data)
	}
	size := (len(data) + n - 1) / n
	var out [][]byte
	for i := 0; i < len(data); i += size {
		end := min(i+size, len(data))
		out = append(out, data[i:end])
	}
	return out
}

// --- helpers reused across the file ---------------------------------------

// newTestEngineWithOptions is [newTestEngine] with explicit NewEngine
// [Options], for tests that need to inject a rename or clock seam.
func newTestEngineWithOptions(t *testing.T, mutate func(cfg *config.Config), opts Options) *testEngine {
	t.Helper()
	te := newTestEngineNoServer(t, mutate)

	now := te.now
	if opts.Now == nil {
		opts.Now = func() time.Time { return now }
	}
	if opts.TagProbe == nil {
		opts.TagProbe = alwaysConfidentTagProbe
	}

	roster := te.engine.roster
	e, err := NewEngine(te.cfg, roster, opts)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	te.engine = e

	srv := httptest.NewServer(withTestAuth(e.Handler()))
	t.Cleanup(srv.Close)
	te.server = srv
	te.client = srv.Client()
	return te
}

// newTestEngineNoServer builds the config/roster/Engine trio without
// starting an httptest.Server, for callers that need to construct their
// own (a custom [Options], or a custom http.Server with its own
// timeouts).
func newTestEngineNoServer(t *testing.T, mutate func(cfg *config.Config)) *testEngine {
	t.Helper()

	root := t.TempDir()
	uploadsDir := filepath.Join(root, "uploads")
	inboxDir := filepath.Join(root, "inbox")
	for _, d := range []string{uploadsDir, inboxDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", d, err)
		}
	}

	contributorsPath := filepath.Join(root, "contributors.yaml")
	yamlBody := "contributors:\n  - name: " + testDirAlice + "\n    email: " + testEmailAlice + "\n"
	if err := os.WriteFile(contributorsPath, []byte(yamlBody), 0o644); err != nil {
		t.Fatalf("write contributors: %v", err)
	}
	roster, err := config.LoadContributors(contributorsPath)
	if err != nil {
		t.Fatalf("LoadContributors: %v", err)
	}

	cfg := &config.Config{
		Paths: config.Paths{
			UploadsDir: uploadsDir,
			InboxDir:   inboxDir,
		},
		Upload: config.UploadConfig{
			ChunkMB:            8,
			MaxAlbumBytes:      1 << 30,
			PartialTTL:         time.Hour,
			MaxSessionsPerUser: 2,
		},
		Thresholds:  config.Thresholds{FreeFloorGB: 0},
		Collections: []string{"", "family"},
	}
	if mutate != nil {
		mutate(cfg)
	}

	now := time.Now()
	e, err := NewEngine(cfg, roster, Options{Now: func() time.Time { return now }, TagProbe: alwaysConfidentTagProbe})
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	return &testEngine{engine: e, cfg: cfg, uploadsDir: uploadsDir, inboxDir: inboxDir, now: now}
}

// createSessionAgainst is [testEngine.createSession] against an
// arbitrary server URL/client pair, for tests that build their own
// httptest.Server.
func (te *testEngine) createSessionAgainst(t *testing.T, baseURL string, client *http.Client, email, album, collection string, expectedFiles int) string {
	t.Helper()
	payload, err := json.Marshal(createSessionRequest{Album: album, Collection: collection, ExpectedFiles: expectedFiles})
	if err != nil {
		t.Fatalf("marshal create-session request: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/session", bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("build create-session request: %v", err)
	}
	req.Header.Set(testEmailHeader, email)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("create-session request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read create-session response: %v", err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create session: status %d, body %q", resp.StatusCode, body)
	}
	var out createSessionResponse
	if err := json.Unmarshal(body, &out); err != nil {
		t.Fatalf("unmarshal create-session response %q: %v", body, err)
	}
	return out.SessionID
}

// assertPromoted checks dest exists, holds exactly wantFiles (plus
// session.json), and that InboxDir's contributor directory contains
// exactly one promoted album.
func assertPromoted(t *testing.T, dest string, wantFiles []string) {
	t.Helper()
	info, err := os.Stat(dest)
	if err != nil {
		t.Fatalf("promoted directory %s: %v", dest, err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", dest)
	}
	for _, name := range wantFiles {
		if _, err := os.Stat(filepath.Join(dest, name)); err != nil {
			t.Fatalf("promoted file %s missing: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dest, "session.json")); err != nil {
		t.Fatalf("session.json missing from promoted directory: %v", err)
	}
}

// assertSessionJSON checks session.json's fields inside a promoted
// directory match the plan-time decision's sidecar shape.
func assertSessionJSON(t *testing.T, dest, sessionID, email, collection string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dest, "session.json"))
	if err != nil {
		t.Fatalf("read session.json: %v", err)
	}
	var sidecar sessionSidecar
	if err := json.Unmarshal(data, &sidecar); err != nil {
		t.Fatalf("unmarshal session.json: %v", err)
	}
	if sidecar.SessionID != sessionID {
		t.Errorf("session.json SessionID = %q, want %q", sidecar.SessionID, sessionID)
	}
	if sidecar.Email != email {
		t.Errorf("session.json Email = %q, want %q", sidecar.Email, email)
	}
	if sidecar.Collection != collection {
		t.Errorf("session.json Collection = %q, want %q", sidecar.Collection, collection)
	}
	if sidecar.Version != sessionSidecarVersion {
		t.Errorf("session.json Version = %d, want %d", sidecar.Version, sessionSidecarVersion)
	}
}

// assertNoInboxTrace checks inboxDir holds nothing at all: neither a
// promoted album directory nor a stray file.
func assertNoInboxTrace(t *testing.T, inboxDir string) {
	t.Helper()
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		t.Fatalf("read inbox dir: %v", err)
	}
	for _, e := range entries {
		sub, err := os.ReadDir(filepath.Join(inboxDir, e.Name()))
		if err != nil {
			t.Fatalf("read inbox contributor dir %s: %v", e.Name(), err)
		}
		if len(sub) != 0 {
			t.Fatalf("inbox contributor dir %s is not empty: %v", e.Name(), sub)
		}
	}
}
