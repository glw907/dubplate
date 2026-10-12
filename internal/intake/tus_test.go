package intake

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
)

// The prefix and public URL every tus fixture mounts under.
const (
	tusPrefix    = "/dub"
	tusPublicURL = "https://music.example"
)

// tusFaults is what a tus fixture's requests read from the disk and the
// settings, changed between requests: the free-space read, the floor
// setting, and receipt's sync. The server reads them on its own
// goroutines, so every access holds mu.
type tusFaults struct {
	mu       sync.Mutex
	free     uint64
	freeErr  error
	floor    int64
	floorErr error
	sync     func(path string) error
}

func (f *tusFaults) set(change func(f *tusFaults)) {
	f.mu.Lock()
	defer f.mu.Unlock()
	change(f)
}

func (f *tusFaults) freeBytes(string) (uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.free, f.freeErr
}

func (f *tusFaults) readFloor(context.Context) (int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.floor, f.floorErr
}

func (f *tusFaults) syncPath(path string) error {
	f.mu.Lock()
	sync := f.sync
	f.mu.Unlock()
	if sync == nil {
		return nil
	}
	return sync(path)
}

// tusFixture is a tus mount over a fresh store and staging directory,
// served over HTTP. Its verifier admits the member a request names in
// the memberHeader header, standing in for the session cookie.
type tusFixture struct {
	h      *store.Handle
	in     *Intake
	s      *submissions
	faults *tusFaults
	url    string
	client *http.Client
}

// memberHeader carries the test verifier's member. failVerify in it
// makes the verifier fail as a broken session store would.
const (
	memberHeader = "X-Test-Member"
	failVerify   = "fail"
)

// errVerifierBroken is the test verifier's failure; its text must never
// reach a response.
var errVerifierBroken = errors.New("session store unreadable at /var/lib/secret")

func testVerifier(_ http.ResponseWriter, r *http.Request) (identity.UserID, error) {
	v := r.Header.Get(memberHeader)
	switch v {
	case "":
		return identity.UserID{}, ErrUnauthenticated
	case failVerify:
		return identity.UserID{}, errVerifierBroken
	}
	b, err := hex.DecodeString(v)
	if err != nil || len(b) != 16 {
		return identity.UserID{}, ErrUnauthenticated
	}
	return identity.UserID(b), nil
}

func newTusFixture(t *testing.T) *tusFixture {
	t.Helper()
	f := newIntakeFixture(t)
	h, s := f.h, f.s
	faults := &tusFaults{free: plentyFree, floor: 1 << 30}
	s.freeBytes = faults.freeBytes
	s.sync = faults.syncPath
	mount, err := f.in.TusMount(TusConfig{
		Prefix:    tusPrefix,
		PublicURL: tusPublicURL,
		Verify:    testVerifier,
		Floor:     faults.readFloor,
		TusLog:    io.Discard,
	})
	if err != nil {
		t.Fatalf("tus mount: %v", err)
	}
	srv := httptest.NewServer(mount)
	t.Cleanup(srv.Close)
	return &tusFixture{h: h, in: f.in, s: s, faults: faults, url: srv.URL, client: srv.Client()}
}

// tusReply is one response, its body read.
type tusReply struct {
	status int
	header http.Header
	body   string
}

// tusRequest is one request to the mount. path is under the shell's
// prefix; member is the session's member, or nil for none.
type tusRequest struct {
	method string
	path   string
	member *identity.UserID
	header map[string]string
	body   []byte
}

func (f *tusFixture) do(t *testing.T, req tusRequest) tusReply {
	t.Helper()
	r, err := http.NewRequestWithContext(deadline(t), req.method, f.url+req.path, bytes.NewReader(req.body))
	if err != nil {
		t.Fatal(err)
	}
	if req.body == nil {
		r.Body, r.ContentLength = http.NoBody, 0
	}
	if req.member != nil {
		r.Header.Set(memberHeader, hex.EncodeToString(req.member[:]))
	}
	for k, v := range req.header {
		if k == "Host" {
			r.Host = v
			continue
		}
		r.Header.Set(k, v)
	}
	resp, err := f.client.Do(r)
	if err != nil {
		t.Fatalf("%s %s: %v", req.method, req.path, err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return tusReply{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

// uploadPath is a file's tus URL path under the shell's prefix.
func uploadPath(file []byte) string {
	return tusPrefix + "/tus/" + hex.EncodeToString(file)
}

// metadata is an Upload-Metadata header naming each key's value, base64
// encoded as tus requires.
func metadata(pairs ...string) string {
	var parts []string
	for i := 0; i < len(pairs); i += 2 {
		parts = append(parts, pairs[i]+" "+base64.StdEncoding.EncodeToString([]byte(pairs[i+1])))
	}
	return strings.Join(parts, ",")
}

// createRequest is the tus creation for file of submission c declaring
// length bytes, with extra headers merged over the protocol's own.
func createRequest(m identity.UserID, c created, file int, length int64, extra map[string]string) tusRequest {
	header := map[string]string{
		"Tus-Resumable":   "1.0.0",
		"Upload-Length":   strconv.FormatInt(length, 10),
		"Upload-Metadata": metadata("submission", hex.EncodeToString(c.id), "file", hex.EncodeToString(c.files[file].id)),
	}
	maps.Copy(header, extra)
	return tusRequest{method: http.MethodPost, path: tusPrefix + "/tus/", member: &m, header: header}
}

// patchRequest appends body to file at offset.
func patchRequest(m identity.UserID, file []byte, offset int64, body []byte) tusRequest {
	return tusRequest{method: http.MethodPatch, path: uploadPath(file), member: &m, body: body, header: map[string]string{
		"Tus-Resumable": "1.0.0",
		"Upload-Offset": strconv.FormatInt(offset, 10),
		"Content-Type":  "application/offset+octet-stream",
	}}
}

func headRequest(m identity.UserID, path string) tusRequest {
	return tusRequest{method: http.MethodHead, path: path, member: &m, header: map[string]string{"Tus-Resumable": "1.0.0"}}
}

// mustStatus fails the test unless reply answered want.
func mustStatus(t *testing.T, what string, reply tusReply, want int) {
	t.Helper()
	if reply.status != want {
		t.Fatalf("%s: status %d, body %q; want %d", what, reply.status, reply.body, want)
	}
}

// startUpload creates c's first file for m and appends sent, the file's
// first bytes, so the upload holds them.
func (f *tusFixture) startUpload(t *testing.T, m identity.UserID, c created, size int64, sent []byte) {
	t.Helper()
	mustStatus(t, "create", f.do(t, createRequest(m, c, 0, size, nil)), http.StatusCreated)
	if len(sent) > 0 {
		mustStatus(t, "patch", f.do(t, patchRequest(m, c.files[0].id, 0, sent)), http.StatusNoContent)
	}
}

// tusFile is the bytes tusd holds for file, or nil when it holds none.
func (f *tusFixture) tusFile(t *testing.T, file []byte) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(f.s.staging, "tus", hex.EncodeToString(file)))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (f *tusFixture) infoExists(t *testing.T, file []byte) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(f.s.staging, "tus", hex.EncodeToString(file)+".info"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	return err == nil
}

// TestTusMountIsConfiguredClosed is AC 7 over a created upload holding
// bytes: no CORS header on any answer, downloads and termination
// unmounted, only creation advertised, a deferred length refused, and
// a Location built from the public URL whatever Host the request names.
func TestTusMountIsConfiguredClosed(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(5, 7), testLimits())
	origin := map[string]string{"Origin": "https://elsewhere.example", "Host": "elsewhere.example"}

	created := f.do(t, createRequest(m, c, 0, 5, origin))
	mustStatus(t, "create", created, http.StatusCreated)
	if got, want := created.header.Get("Location"), tusPublicURL+tusPrefix+"/tus/"+hex.EncodeToString(c.files[0].id); got != want {
		t.Errorf("Location = %q, want %q", got, want)
	}
	mustStatus(t, "patch", f.do(t, patchRequest(m, c.files[0].id, 0, []byte("abc"))), http.StatusNoContent)

	replies := map[string]tusReply{"create": created}
	for _, method := range []string{http.MethodGet, http.MethodDelete} {
		header := map[string]string{"Tus-Resumable": "1.0.0", "Origin": origin["Origin"], "Host": origin["Host"]}
		reply := f.do(t, tusRequest{method: method, path: uploadPath(c.files[0].id), member: &m, header: header})
		if reply.status != http.StatusMethodNotAllowed {
			t.Errorf("%s on an upload: status %d, want 405", method, reply.status)
		}
		replies[method] = reply
	}
	if got := f.tusFile(t, c.files[0].id); string(got) != "abc" {
		t.Errorf("upload holds %q after GET and DELETE, want %q", got, "abc")
	}

	options := f.do(t, tusRequest{method: http.MethodOptions, path: tusPrefix + "/tus/", member: &m, header: origin})
	if got := options.header.Get("Tus-Extension"); got != "creation,creation-with-upload" {
		t.Errorf("Tus-Extension = %q, want exactly creation,creation-with-upload", got)
	}
	replies["OPTIONS"] = options

	deferred := createRequest(m, c, 1, 7, origin)
	delete(deferred.header, "Upload-Length")
	deferred.header["Upload-Defer-Length"] = "1"
	reply := f.do(t, deferred)
	if reply.status != http.StatusNotImplemented {
		t.Errorf("Upload-Defer-Length: status %d, body %q; want 501", reply.status, reply.body)
	}
	if f.infoExists(t, c.files[1].id) {
		t.Error("a deferred-length creation wrote an upload")
	}
	replies["deferred"] = reply

	for name, reply := range replies {
		for key := range reply.header {
			if strings.HasPrefix(key, "Access-Control-") {
				t.Errorf("%s answered %s", name, key)
			}
		}
	}
}

// storedInfo is the part of tusd's .info file the tests read.
type storedInfo struct {
	ID       string
	Size     int64
	MetaData map[string]string
}

// TestTusCreateKeepsNoClientMetadata is AC 10: a creation carrying extra
// keys and a forged uploader stores none of them and takes the file id
// as its tus id; a creation declaring one byte more than the manifest,
// and a second creation for a file whose upload exists, each answer 400
// before anything is written.
func TestTusCreateKeepsNoClientMetadata(t *testing.T) {
	f := newTusFixture(t)
	m, other := member(t, f.h), member(t, f.h)
	c := mustCreate(t, f.s, m, album(4, 6), testLimits())

	forged := createRequest(m, c, 0, 4, map[string]string{"Upload-Metadata": metadata(
		"submission", hex.EncodeToString(c.id), "file", hex.EncodeToString(c.files[0].id),
		"uploader", hex.EncodeToString(other[:]), "filename", "../../escape.flac")})
	mustStatus(t, "create with extra metadata", f.do(t, forged), http.StatusCreated)

	raw, err := os.ReadFile(filepath.Join(f.s.staging, "tus", hex.EncodeToString(c.files[0].id)+".info"))
	if err != nil {
		t.Fatalf("read .info: %v", err)
	}
	var info storedInfo
	if err := json.Unmarshal(raw, &info); err != nil {
		t.Fatalf("parse .info: %v", err)
	}
	if info.ID != hex.EncodeToString(c.files[0].id) || info.Size != 4 {
		t.Errorf(".info id %q size %d, want the file id %x and size 4", info.ID, info.Size, c.files[0].id)
	}
	if len(info.MetaData) != 0 {
		t.Errorf(".info keeps client metadata %v, want none", info.MetaData)
	}
	if bytes.Contains(raw, []byte(hex.EncodeToString(other[:]))) || bytes.Contains(raw, []byte("escape")) {
		t.Errorf(".info carries a forged value: %s", raw)
	}

	tooLong := f.do(t, createRequest(m, c, 1, 7, nil))
	assertRefusal(t, "one byte over the declared size", tooLong, http.StatusBadRequest, "invalid_request")
	if f.infoExists(t, c.files[1].id) {
		t.Error("a creation over the declared size wrote .info")
	}

	mustStatus(t, "patch", f.do(t, patchRequest(m, c.files[0].id, 0, []byte("ab"))), http.StatusNoContent)
	again := f.do(t, createRequest(m, c, 0, 4, nil))
	assertRefusal(t, "a second creation for one file", again, http.StatusBadRequest, "invalid_request")
	if got := f.tusFile(t, c.files[0].id); string(got) != "ab" {
		t.Errorf("upload holds %q after a refused second creation, want %q", got, "ab")
	}
}

// TestTusReceipt is AC 11's server half over a two-file submission:
// each finished upload is received, the last moves the submission to
// checking, a HEAD after receipt answers the full offset, a HEAD 404 on
// a file with no upload is followed by a create that succeeds, and a
// lost create's retry is refused because the upload exists.
func TestTusReceipt(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(5, 3), testLimits())
	drain(f.s)

	head := f.do(t, headRequest(m, uploadPath(c.files[1].id)))
	mustStatus(t, "HEAD before any create", head, http.StatusNotFound)

	f.startUpload(t, m, c, 5, []byte("ab"))
	mustStatus(t, "last patch", f.do(t, patchRequest(m, c.files[0].id, 2, []byte("cde"))), http.StatusNoContent)
	if at := receivedAt(t, f.s, c.files[0].id); at == nil {
		t.Fatal("the finished first file is not received")
	}
	if got := stateOf(t, f.h, c.id); got != "uploading" {
		t.Fatalf("state %q after one of two files, want uploading", got)
	}

	// The second file goes up in its creation request.
	withBody := createRequest(m, c, 1, 3, map[string]string{"Content-Type": "application/offset+octet-stream"})
	withBody.body = []byte("xyz")
	mustStatus(t, "create with upload", f.do(t, withBody), http.StatusCreated)
	if got := stateOf(t, f.h, c.id); got != "checking" {
		t.Fatalf("state %q after the last file, want checking", got)
	}
	if len(f.s.wake) != 1 {
		t.Errorf("wake holds %d pokes, want 1", len(f.s.wake))
	}

	for i, size := range []int64{5, 3} {
		reply := f.do(t, headRequest(m, uploadPath(c.files[i].id)))
		mustStatus(t, "HEAD after receipt", reply, http.StatusOK)
		if got := reply.header.Get("Upload-Offset"); got != strconv.FormatInt(size, 10) {
			t.Errorf("file %d: Upload-Offset %q after receipt, want %d", i, got, size)
		}
	}

	assertRefusal(t, "a create for a received file", f.do(t, createRequest(m, c, 0, 5, nil)), http.StatusBadRequest, "invalid_request")
	if got := f.tusFile(t, c.files[0].id); string(got) != "abcde" {
		t.Errorf("received file holds %q, want %q", got, "abcde")
	}
	assertStatesMatchEvents(t, f.h)
}

// TestTusLostCreateRetry is AC 11's lost 201: the retried creation for
// a file whose upload exists answers 400, and the client finds the
// upload by HEAD instead.
func TestTusLostCreateRetry(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(4), testLimits())
	mustStatus(t, "create", f.do(t, createRequest(m, c, 0, 4, nil)), http.StatusCreated)

	assertRefusal(t, "the lost create's retry", f.do(t, createRequest(m, c, 0, 4, nil)), http.StatusBadRequest, "invalid_request")
	head := f.do(t, headRequest(m, uploadPath(c.files[0].id)))
	mustStatus(t, "HEAD", head, http.StatusOK)
	if got := head.header.Get("Upload-Offset"); got != "0" {
		t.Errorf("Upload-Offset = %q, want 0", got)
	}
}

// TestTusMountRefusesAPublicURLWithoutSchemeAndHost: tusd builds a
// Location from the base path alone only when it carries a scheme, and
// otherwise from the request's Host, so a public URL lacking either is
// refused when the mount is built.
func TestTusMountRefusesAPublicURLWithoutSchemeAndHost(t *testing.T) {
	f := newIntakeFixture(t)
	for _, publicURL := range []string{"//localhost:8080", "music.example", "", "https://"} {
		_, err := f.in.TusMount(TusConfig{
			Prefix: tusPrefix, PublicURL: publicURL, Verify: testVerifier,
			Floor: func(context.Context) (int64, error) { return 0, nil }, TusLog: io.Discard,
		})
		if err == nil {
			t.Errorf("TusMount(%q) built a mount, want a refusal", publicURL)
		}
	}
}

// TestTusAdmitsAFileBetweenTheCaps is P7's MaxSize row: an album cap
// raised after the mount was built admitted a 5 GiB file, past the
// default cap, and tusd must not answer its creation 413.
func TestTusAdmitsAFileBetweenTheCaps(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	const size = 5 << 30
	c := mustCreate(t, f.s, m, album(size), testLimits())

	mustStatus(t, "create between the caps", f.do(t, createRequest(m, c, 0, size, nil)), http.StatusCreated)
}
