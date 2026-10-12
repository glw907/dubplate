package opshttp

import (
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/ops"
)

// mountPrefix is the owned prefix every mount test serves under.
const mountPrefix = "/dub"

// A mountFixture is a mount over fake seams and a counting verifier.
type mountFixture struct {
	mount    http.Handler
	seam     *fakeSeam
	admin    *fakeAdmin
	verified int
	log      *bytes.Buffer
}

// newMountFixture builds a mount whose verifier answers decide.
func newMountFixture(t *testing.T, decide func(w http.ResponseWriter, r *http.Request) (ops.Actor, ops.Privilege, error)) *mountFixture {
	t.Helper()
	f := &mountFixture{
		seam:  &fakeSeam{answers: successes()},
		admin: &fakeAdmin{answers: adminSuccesses()},
		log:   &bytes.Buffer{},
	}
	verify := func(w http.ResponseWriter, r *http.Request) (ops.Actor, ops.Privilege, error) {
		f.verified++
		return decide(w, r)
	}
	m, err := NewMount(f.seam, f.admin, mountPrefix, verify, slog.New(slog.NewTextHandler(f.log, nil)))
	if err != nil {
		t.Fatalf("NewMount: %v", err)
	}
	f.mount = m
	return f
}

// admitAs is a verifier admitting every request as ann with p.
func admitAs(p ops.Privilege) func(http.ResponseWriter, *http.Request) (ops.Actor, ops.Privilege, error) {
	return func(http.ResponseWriter, *http.Request) (ops.Actor, ops.Privilege, error) {
		return ops.NetworkActor(ann, annAnchor), p, nil
	}
}

func (f *mountFixture) do(method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/x", strings.NewReader(body))
	req.URL.Path = path
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	f.mount.ServeHTTP(rec, req)
	return rec
}

// TestMount_OtherAPIPathsAnswerJSONNotFound proves every path under
// <prefix>/api that names no v1 operation answers 404 not_found in the
// JSON error shape, before the session is verified (AC4).
func TestMount_OtherAPIPathsAnswerJSONNotFound(t *testing.T) {
	for _, path := range []string{
		mountPrefix + "/api",
		mountPrefix + "/api/",
		mountPrefix + "/api/v1",
		mountPrefix + "/api/v2/catalog",
		mountPrefix + "/api/v10/catalog",
		mountPrefix + "/api/catalog",
		mountPrefix + "/apix",
	} {
		t.Run(path, func(t *testing.T) {
			f := newMountFixture(t, admitAs(ops.PrivilegeAdmin))
			rec := f.do(http.MethodPost, path, `{}`)
			if rec.Code != http.StatusNotFound {
				t.Fatalf("status = %d, want 404", rec.Code)
			}
			if code, _, _ := errorAnswer(t, rec); code != ops.CodeNotFound {
				t.Errorf("code = %q, want not_found", code)
			}
			if f.verified != 0 {
				t.Errorf("verifier ran %d times before routing, want 0", f.verified)
			}
		})
	}
}

// TestMount_NoSessionAnswers401 proves a verifier refusal wrapping
// ErrUnauthenticated answers 401 unauthenticated and reaches no seam,
// on every operation and on an unknown name under v1.
func TestMount_NoSessionAnswers401(t *testing.T) {
	refuse := func(http.ResponseWriter, *http.Request) (ops.Actor, ops.Privilege, error) {
		return ops.Actor{}, 0, fmt.Errorf("verify: %w", ErrUnauthenticated)
	}
	paths := []string{"catalog", "list_members", "add_member", "set_setting", "start_backup", "no_such_operation"}
	for _, name := range paths {
		t.Run(name, func(t *testing.T) {
			f := newMountFixture(t, refuse)
			rec := f.do(http.MethodPost, mountPrefix+"/api/v1/"+name, `{}`)
			if rec.Code != http.StatusUnauthorized {
				t.Fatalf("status = %d, want 401", rec.Code)
			}
			if code, _, _ := errorAnswer(t, rec); code != ops.CodeUnauthenticated {
				t.Errorf("code = %q, want unauthenticated", code)
			}
			if len(f.seam.calls)+len(f.admin.calls) != 0 {
				t.Errorf("a refused request reached a seam: %v %v", f.seam.calls, f.admin.calls)
			}
		})
	}
}

// TestMount_VerifierFailureAnswers500 proves any other verifier error
// is the server failing: 500 internal with no reason on the wire and
// none of the error's text, which may quote an address, in the log.
func TestMount_VerifierFailureAnswers500(t *testing.T) {
	f := newMountFixture(t, func(http.ResponseWriter, *http.Request) (ops.Actor, ops.Privilege, error) {
		return ops.Actor{}, 0, errors.New("roster lookup failed for ann@example.com")
	})
	rec := f.do(http.MethodPost, mountPrefix+"/api/v1/catalog", `{}`)
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", rec.Code)
	}
	code, reason, hasReason := errorAnswer(t, rec)
	if code != ops.CodeInternal || hasReason && reason != "" {
		t.Errorf("answer = %q %q, want internal and no reason", code, reason)
	}
	if strings.Contains(rec.Body.String(), "ann@example.com") || strings.Contains(f.log.String(), "ann@example.com") {
		t.Errorf("the verifier's error text leaked: body %q log %q", rec.Body.String(), f.log.String())
	}
	if !strings.Contains(f.log.String(), "level=ERROR") {
		t.Errorf("log %q has no Error record", f.log.String())
	}
	if len(f.seam.calls)+len(f.admin.calls) != 0 {
		t.Errorf("a failed verification reached a seam")
	}
}

// TestMount_AdmittedCallerRidesTheContext proves an admitted caller
// reaches the operation as the verifier's actor, whatever the body
// names, and that an admin operation reaches the admin seam.
func TestMount_AdmittedCallerRidesTheContext(t *testing.T) {
	f := newMountFixture(t, admitAs(ops.PrivilegeAdmin))
	rec := f.do(http.MethodPost, mountPrefix+"/api/v1/catalog", `{"actor":"member:ffffffffffffffffffffffffffffffff"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("catalog status = %d, body %s", rec.Code, rec.Body)
	}
	if len(f.seam.calls) != 1 || f.seam.calls[0].actor.String() != annWire {
		t.Fatalf("job seam calls = %v, want one as %s", f.seam.calls, annWire)
	}
	rec = f.do(http.MethodPost, mountPrefix+"/api/v1/set_setting", `{"key":"DUBPLATE_REAUTH_WINDOW","value":"3h","actor":"member:ffffffffffffffffffffffffffffffff"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("set_setting status = %d, body %s", rec.Code, rec.Body)
	}
	if len(f.admin.calls) != 1 || f.admin.calls[0].actor.String() != annWire {
		t.Fatalf("admin seam calls = %v, want one as %s", f.admin.calls, annWire)
	}
	if f.verified != 2 {
		t.Errorf("verifier ran %d times, want once per request", f.verified)
	}
}

// TestMount_MemberReachesOnlyTheCatalog proves a member session meets
// the catalog and is refused every admin operation with forbidden, with
// no admin seam call (AC5).
func TestMount_MemberReachesOnlyTheCatalog(t *testing.T) {
	f := newMountFixture(t, admitAs(ops.PrivilegeMember))
	if rec := f.do(http.MethodPost, mountPrefix+"/api/v1/catalog", `{}`); rec.Code != http.StatusOK {
		t.Fatalf("catalog status = %d, body %s", rec.Code, rec.Body)
	}
	for _, name := range []string{"list_members", "add_member", "remove_member", "demote_member", "list_settings", "set_setting", "reset_setting"} {
		rec := f.do(http.MethodPost, mountPrefix+"/api/v1/"+name, `{}`)
		if rec.Code != http.StatusForbidden {
			t.Errorf("%s status = %d, want 403", name, rec.Code)
		}
		if code, _, _ := errorAnswer(t, rec); code != ops.CodeForbidden {
			t.Errorf("%s code = %q, want forbidden", name, code)
		}
	}
	if len(f.admin.calls) != 0 {
		t.Errorf("a member's call reached the admin seam: %v", f.admin.calls)
	}
}

// TestMount_BearerHeaderAdmitsNobody proves the mount never reads an
// Authorization header: the verifier's answer alone decides, so a
// bearer header with no session is refused and a bogus one beside a
// session changes nothing.
func TestMount_BearerHeaderAdmitsNobody(t *testing.T) {
	sawBearer := false
	f := newMountFixture(t, func(_ http.ResponseWriter, r *http.Request) (ops.Actor, ops.Privilege, error) {
		if r.Header.Get("Authorization") == "" {
			return ops.Actor{}, 0, fmt.Errorf("no cookie: %w", ErrUnauthenticated)
		}
		sawBearer = true
		return ops.Actor{}, 0, fmt.Errorf("bearer is no session: %w", ErrUnauthenticated)
	})
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader(`{}`))
	req.URL.Path = mountPrefix + "/api/v1/catalog"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer 0123456789abcdef0123456789abcdef")
	rec := httptest.NewRecorder()
	f.mount.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	if !sawBearer {
		t.Fatal("the verifier never saw the request's header, so the test proves nothing")
	}
	if len(f.seam.calls) != 0 {
		t.Errorf("a bearer header reached the job seam")
	}
}

// TestMount_UnknownOperationAnswers404AfterVerification proves a name
// under v1 that is no operation answers not_found for a verified
// caller.
func TestMount_UnknownOperationAnswers404AfterVerification(t *testing.T) {
	f := newMountFixture(t, admitAs(ops.PrivilegeAdmin))
	rec := f.do(http.MethodPost, mountPrefix+"/api/v1/no_such_operation", `{}`)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if code, _, _ := errorAnswer(t, rec); code != ops.CodeNotFound {
		t.Errorf("code = %q, want not_found", code)
	}
	if f.verified != 1 {
		t.Errorf("verifier ran %d times, want 1", f.verified)
	}
}

// TestMount_VerifierCanSetCookies proves the verifier is handed the
// response writer, so a refreshed session cookie reaches the answer.
func TestMount_VerifierCanSetCookies(t *testing.T) {
	f := newMountFixture(t, func(w http.ResponseWriter, _ *http.Request) (ops.Actor, ops.Privilege, error) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "refreshed"})
		return ops.NetworkActor(ann, annAnchor), ops.PrivilegeAdmin, nil
	})
	rec := f.do(http.MethodPost, mountPrefix+"/api/v1/catalog", `{}`)
	if got := rec.Header().Get("Set-Cookie"); !strings.HasPrefix(got, "session=refreshed") {
		t.Errorf("Set-Cookie = %q, want the verifier's cookie", got)
	}
}

// TestNewMount_Refusals proves a mount is not built without its parts.
func TestNewMount_Refusals(t *testing.T) {
	verify := func(http.ResponseWriter, *http.Request) (ops.Actor, ops.Privilege, error) {
		return ops.Actor{}, 0, ErrUnauthenticated
	}
	log := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	seam, admin := &fakeSeam{}, &fakeAdmin{}
	for _, tt := range []struct {
		name string
		call func() (http.Handler, error)
	}{
		{"nil verifier", func() (http.Handler, error) { return NewMount(seam, admin, mountPrefix, nil, log) }},
		{"nil job seam", func() (http.Handler, error) { return NewMount(nil, admin, mountPrefix, verify, log) }},
		{"nil admin seam", func() (http.Handler, error) { return NewMount(seam, nil, mountPrefix, verify, log) }},
		{"nil logger", func() (http.Handler, error) { return NewMount(seam, admin, mountPrefix, verify, nil) }},
		{"prefix without a leading separator", func() (http.Handler, error) { return NewMount(seam, admin, "dub", verify, log) }},
		{"prefix with a trailing separator", func() (http.Handler, error) { return NewMount(seam, admin, "/dub/", verify, log) }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if h, err := tt.call(); err == nil || h != nil {
				t.Errorf("NewMount = %v, %v, want a refusal", h, err)
			}
		})
	}
}
