package intake

import (
	"bytes"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
)

// assertRefusal checks one refusal is in tusd's format: the status, a
// text/plain body that begins with the code, and none of the injected
// faults' text.
func assertRefusal(t *testing.T, what string, reply tusReply, status int, code string) {
	t.Helper()
	if reply.status != status {
		t.Errorf("%s: status %d, body %q; want %d", what, reply.status, reply.body, status)
	}
	if got := reply.header.Get("Content-Type"); got != "text/plain; charset=utf-8" {
		t.Errorf("%s: Content-Type %q, want text/plain; charset=utf-8", what, got)
	}
	if !strings.HasPrefix(reply.body, code+": ") || !strings.HasSuffix(reply.body, "\n") {
		t.Errorf("%s: body %q, want %q then a reason and a newline", what, reply.body, code+": ")
	}
	for _, leak := range []string{errDiskOnFire.Error(), errVerifierBroken.Error(), "/var/lib"} {
		if strings.Contains(reply.body, leak) {
			t.Errorf("%s: body %q leaks %q", what, reply.body, leak)
		}
	}
}

// errDiskOnFire is an injected disk or settings fault.
var errDiskOnFire = errors.New("statfs /srv/music/staging: disk on fire")

// refusedBeforeTusd reports whether the wrapper answered reply itself:
// tusd's middleware stamps Tus-Resumable on every answer it writes.
func refusedBeforeTusd(reply tusReply) bool {
	return reply.header.Get("Tus-Resumable") == ""
}

// TestTusRefusals is every row of the spec's "Refusals on the tus path",
// each asserting its status, text/plain, and the body's code.
func TestTusRefusals(t *testing.T) {
	type fixture struct {
		*tusFixture
		m identity.UserID
		c created
	}
	tests := []struct {
		name   string
		run    func(t *testing.T, f fixture) tusReply
		status int
		code   string
	}{
		{
			name: "no session",
			run: func(t *testing.T, f fixture) tusReply {
				req := createRequest(f.m, f.c, 0, 4, nil)
				req.member = nil
				return f.do(t, req)
			},
			status: http.StatusUnauthorized, code: "unauthenticated",
		},
		{
			name: "the verifier fails",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 4, map[string]string{memberHeader: failVerify}))
			},
			status: http.StatusInternalServerError, code: "internal",
		},
		{
			name: "method override",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 4, map[string]string{"X-HTTP-Method-Override": "DELETE"}))
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "method override on an upload",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, []byte("ab"))
				req := patchRequest(f.m, f.c.files[0].id, 2, []byte("cd"))
				req.method = http.MethodPost
				req.header["X-HTTP-Method-Override"] = "PATCH"
				return f.do(t, req)
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "Upload-Concat",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 4, map[string]string{"Upload-Concat": "partial"}))
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "create: size differs from the declared size",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 3, nil))
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "create: the file already has an upload",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, nil)
				return f.do(t, createRequest(f.m, f.c, 0, 4, nil))
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "create: metadata naming no file",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 4, map[string]string{"Upload-Metadata": metadata("submission", "nope", "file", "nope")}))
			},
			status: http.StatusBadRequest, code: "invalid_request",
		},
		{
			name: "create: over MaxSize",
			run: func(t *testing.T, f fixture) tusReply {
				return f.do(t, createRequest(f.m, f.c, 0, 17179869185, nil))
			},
			status: http.StatusRequestEntityTooLarge, code: "ERR_MAX_SIZE_EXCEEDED",
		},
		{
			name: "create: the floor",
			run: func(t *testing.T, f fixture) tusReply {
				f.faults.set(func(d *tusFaults) { d.free = 1<<30 - 1 })
				return f.do(t, createRequest(f.m, f.c, 0, 4, nil))
			},
			status: http.StatusForbidden, code: "no-space",
		},
		{
			name: "PATCH: free space crosses the floor after the create",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, []byte("ab"))
				f.faults.set(func(d *tusFaults) { d.free = 1<<30 - 1 })
				return f.do(t, patchRequest(f.m, f.c.files[0].id, 2, []byte("cd")))
			},
			status: http.StatusForbidden, code: "no-space",
		},
		{
			name: "create: statfs fails",
			run: func(t *testing.T, f fixture) tusReply {
				f.faults.set(func(d *tusFaults) { d.freeErr = errDiskOnFire })
				return f.do(t, createRequest(f.m, f.c, 0, 4, nil))
			},
			status: http.StatusServiceUnavailable, code: "could-not-check",
		},
		{
			name: "PATCH: statfs fails",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, []byte("ab"))
				f.faults.set(func(d *tusFaults) { d.freeErr = errDiskOnFire })
				return f.do(t, patchRequest(f.m, f.c.files[0].id, 2, []byte("cd")))
			},
			status: http.StatusServiceUnavailable, code: "could-not-check",
		},
		{
			name: "create: the floor setting cannot be read",
			run: func(t *testing.T, f fixture) tusReply {
				f.faults.set(func(d *tusFaults) { d.floorErr = errDiskOnFire })
				return f.do(t, createRequest(f.m, f.c, 0, 4, nil))
			},
			status: http.StatusServiceUnavailable, code: "could-not-check",
		},
		{
			name: "receipt into a closed submission",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, []byte("ab"))
				// The member cancels while the last bytes are written: the
				// cancel commits between the wrapper's check and receipt.
				f.faults.set(func(d *tusFaults) {
					d.sync = func(string) error {
						f.faults.set(func(d *tusFaults) { d.sync = nil })
						return f.s.cancel(deadline(t), f.m, f.c.id)
					}
				})
				return f.do(t, patchRequest(f.m, f.c.files[0].id, 2, []byte("cd")))
			},
			status: http.StatusForbidden, code: "conflict",
		},
		{
			name: "receipt: a store fault",
			run: func(t *testing.T, f fixture) tusReply {
				f.startUpload(t, f.m, f.c, 4, []byte("ab"))
				f.faults.set(func(d *tusFaults) { d.sync = func(string) error { return errDiskOnFire } })
				return f.do(t, patchRequest(f.m, f.c.files[0].id, 2, []byte("cd")))
			},
			status: http.StatusServiceUnavailable, code: "could-not-check",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tf := newTusFixture(t)
			m := member(t, tf.h)
			f := fixture{tusFixture: tf, m: m, c: mustCreate(t, tf.s, m, album(4, 4), testLimits())}
			assertRefusal(t, tt.name, tt.run(t, f), tt.status, tt.code)
		})
	}
}

// TestTusReceiptRefusalsChangeNothing checks the two receipt rows' state
// after the refusal: a closed submission stays closed with the file
// unreceived, and a store fault leaves the file for the sweep to
// receive.
func TestTusReceiptRefusalsChangeNothing(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	c := mustCreate(t, f.s, m, album(4), testLimits())
	f.startUpload(t, m, c, 4, []byte("ab"))
	f.faults.set(func(d *tusFaults) { d.sync = func(string) error { return errDiskOnFire } })
	assertRefusal(t, "receipt fault", f.do(t, patchRequest(m, c.files[0].id, 2, []byte("cd"))),
		http.StatusServiceUnavailable, "could-not-check")
	if at := receivedAt(t, f.s, c.files[0].id); at != nil {
		t.Errorf("received_at = %d after a failed receipt, want NULL", *at)
	}

	// The client's retry HEADs and finds the whole file.
	head := f.do(t, headRequest(m, uploadPath(c.files[0].id)))
	if head.status != http.StatusOK || head.header.Get("Upload-Offset") != "4" {
		t.Errorf("HEAD after a failed receipt: status %d offset %q, want 200 and 4", head.status, head.header.Get("Upload-Offset"))
	}
	assertStatesMatchEvents(t, f.h)
}

// TestTusMemberCannotReachAnothersUpload is AC 8: two members, each with
// an in-flight upload holding bytes. Every request one makes on the
// other's upload answers the unknown-id 404 before tusd sees it, and the
// other's file is byte-identical after each.
func TestTusMemberCannotReachAnothersUpload(t *testing.T) {
	f := newTusFixture(t)
	a, b := member(t, f.h), member(t, f.h)
	ca := mustCreate(t, f.s, a, album(6), testLimits())
	cb := mustCreate(t, f.s, b, album(6), testLimits())
	f.startUpload(t, a, ca, 6, []byte("aaa"))
	f.startUpload(t, b, cb, 6, []byte("bbb"))

	unknown := f.do(t, patchRequest(b, bytes.Repeat([]byte{0xab}, 16), 0, []byte("x")))
	assertRefusal(t, "unknown id", unknown, http.StatusNotFound, "ERR_UPLOAD_NOT_FOUND")

	theirs := ca.files[0].id
	requests := map[string]tusRequest{
		"HEAD":   headRequest(b, uploadPath(theirs)),
		"PATCH":  patchRequest(b, theirs, 3, []byte("bbb")),
		"GET":    {method: http.MethodGet, path: uploadPath(theirs), member: &b},
		"DELETE": {method: http.MethodDelete, path: uploadPath(theirs), member: &b, header: map[string]string{"Tus-Resumable": "1.0.0"}},
	}
	for name, req := range requests {
		reply := f.do(t, req)
		if reply.status != http.StatusNotFound || !refusedBeforeTusd(reply) {
			t.Errorf("%s on another's upload: status %d, Tus-Resumable %q; want the wrapper's 404",
				name, reply.status, reply.header.Get("Tus-Resumable"))
		}
		if req.method != http.MethodHead && reply.body != unknown.body {
			t.Errorf("%s on another's upload: body %q, want the unknown-id body %q", name, reply.body, unknown.body)
		}
		if got := f.tusFile(t, theirs); string(got) != "aaa" {
			t.Fatalf("after %s, the other's upload holds %q, want %q", name, got, "aaa")
		}
	}

	// A creation naming the other's file would truncate it; it answers
	// what a creation naming no file answers.
	forged := createRequest(b, ca, 0, 6, nil)
	reply := f.do(t, forged)
	noFile := f.do(t, createRequest(b, cb, 0, 6, map[string]string{"Upload-Metadata": metadata("submission", "x", "file", "y")}))
	assertRefusal(t, "a creation naming another's file", reply, http.StatusBadRequest, "invalid_request")
	if reply.body != noFile.body {
		t.Errorf("a creation naming another's file answers %q, a creation naming none %q; want one body", reply.body, noFile.body)
	}
	if got := f.tusFile(t, theirs); string(got) != "aaa" {
		t.Errorf("after a forged creation, the other's upload holds %q, want %q", got, "aaa")
	}
	if got := f.tusFile(t, cb.files[0].id); string(got) != "bbb" {
		t.Errorf("the member's own upload holds %q, want %q", got, "bbb")
	}
}

// TestTusDisallowedStatesAnswerTheSharedNotFound is step 4's state rule:
// PATCH only while uploading, HEAD while the submission is open, and a
// closed one's file as unknown as another member's.
func TestTusDisallowedStatesAnswerTheSharedNotFound(t *testing.T) {
	f := newTusFixture(t)
	m := member(t, f.h)
	checking := mustCreate(t, f.s, m, album(2), testLimits())
	f.startUpload(t, m, checking, 2, []byte("ok"))
	cancelled := mustCreate(t, f.s, m, album(4), testLimits())
	f.startUpload(t, m, cancelled, 4, []byte("ab"))
	if err := f.s.cancel(deadline(t), m, cancelled.id); err != nil {
		t.Fatal(err)
	}

	head := f.do(t, headRequest(m, uploadPath(checking.files[0].id)))
	mustStatus(t, "HEAD on a checking submission's file", head, http.StatusOK)

	for name, req := range map[string]tusRequest{
		"PATCH on a checking submission's file":  patchRequest(m, checking.files[0].id, 2, []byte("x")),
		"HEAD on a cancelled submission's file":  headRequest(m, uploadPath(cancelled.files[0].id)),
		"PATCH on a cancelled submission's file": patchRequest(m, cancelled.files[0].id, 2, []byte("cd")),
	} {
		reply := f.do(t, req)
		if reply.status != http.StatusNotFound || !refusedBeforeTusd(reply) {
			t.Errorf("%s: status %d, want the wrapper's 404", name, reply.status)
		}
	}
	if got := f.tusFile(t, cancelled.files[0].id); string(got) != "ab" {
		t.Errorf("cancelled upload holds %q, want %q", got, "ab")
	}
}

// TestTusPathShapes is the States table's path row: each shape answers
// the shared 404 from the wrapper or acts on the authorized id alone.
func TestTusPathShapes(t *testing.T) {
	f := newTusFixture(t)
	m, other := member(t, f.h), member(t, f.h)
	mine := mustCreate(t, f.s, m, album(4), testLimits())
	theirs := mustCreate(t, f.s, other, album(4), testLimits())
	f.startUpload(t, m, mine, 4, []byte("ab"))
	f.startUpload(t, other, theirs, 4, []byte("xy"))
	own, oth := hex.EncodeToString(mine.files[0].id), hex.EncodeToString(theirs.files[0].id)
	base := tusPrefix + "/tus/"

	for _, path := range []string{base + own + "/", base + "/" + own} {
		reply := f.do(t, headRequest(m, path))
		if reply.status != http.StatusOK || reply.header.Get("Upload-Offset") != "2" {
			t.Errorf("HEAD %s: status %d offset %q, want 200 and the own upload's 2", path, reply.status, reply.header.Get("Upload-Offset"))
		}
	}
	for _, path := range []string{
		base + own + "/" + oth,
		base + oth + "/" + own,
		base + own + "%2F" + oth,
		base + strings.ToUpper(own),
		base + oth + "/",
		base + own + "x",
	} {
		for _, req := range []tusRequest{headRequest(m, path), patchRequest(m, nil, 2, []byte("zz"))} {
			req.path = path
			reply := f.do(t, req)
			if reply.status != http.StatusNotFound || !refusedBeforeTusd(reply) {
				t.Errorf("%s %s: status %d, Tus-Resumable %q; want the wrapper's 404",
					req.method, path, reply.status, reply.header.Get("Tus-Resumable"))
			}
		}
	}
	if got := f.tusFile(t, mine.files[0].id); string(got) != "ab" {
		t.Errorf("own upload holds %q, want %q", got, "ab")
	}
	if got := f.tusFile(t, theirs.files[0].id); string(got) != "xy" {
		t.Errorf("other's upload holds %q, want %q", got, "xy")
	}
}
