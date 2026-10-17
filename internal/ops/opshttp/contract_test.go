package opshttp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/local"
)

// The contract suite runs each seam case twice: in-process, against
// ops/local wrapped in the authorizing layer, and over HTTP, through
// the whole projection over the same runner. Both legs must classify
// to the same wire code. Backup jobs run the real backup body, which
// finds rclone through PATH; TestMain puts a stand-in first on PATH
// that blocks in whatever subcommand it is given until it receives
// SIGINT, so a started backup stays running until it is canceled.

// The environment the stand-in reads, set per test.
const (
	// standinRecordEnv names the file the stand-in appends "SIGINT" to
	// for each SIGINT it receives.
	standinRecordEnv = "DUBPLATE_OPSHTTP_STANDIN_RECORD"

	// standinStartedEnv names a directory the stand-in creates one
	// file in, named by its pid, once its SIGINT handler is installed.
	standinStartedEnv = "DUBPLATE_OPSHTTP_STANDIN_STARTED"
)

// standinLimit bounds a stand-in no test signals, so a leaked one
// cannot outlive the suite by much.
const standinLimit = time.Minute

// The environment a re-executed test binary reads to run a backup as a
// second process over a state directory's store.
const (
	childDirEnv  = "DUBPLATE_OPSHTTP_CHILD_DIR"
	childRootEnv = "DUBPLATE_OPSHTTP_CHILD_ROOT"
)

// contractRemote is every backup's remote. The stand-in never reads
// it; the backup body only parses it.
const contractRemote = "contract:dir"

// contractBusy is every contract store's busy timeout.
const contractBusy = 5 * time.Second

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "rclone" {
		os.Exit(standIn())
	}
	if dir := os.Getenv(childDirEnv); dir != "" {
		os.Exit(runChild(dir, os.Getenv(childRootEnv)))
	}
	os.Exit(runSuite(m))
}

// runSuite links the test binary as rclone into a fresh directory put
// first on PATH, runs the tests, and removes the directory.
func runSuite(m *testing.M) int {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, "test binary:", err)
		return 2
	}
	bin, err := os.MkdirTemp("", "opshttp-standin-")
	if err != nil {
		fmt.Fprintln(os.Stderr, "stand-in directory:", err)
		return 2
	}
	defer func() { _ = os.RemoveAll(bin) }()
	if err := os.Symlink(exe, filepath.Join(bin, "rclone")); err != nil {
		fmt.Fprintln(os.Stderr, "link the stand-in:", err)
		return 2
	}
	if err := os.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH")); err != nil {
		fmt.Fprintln(os.Stderr, "put the stand-in on PATH:", err)
		return 2
	}
	return m.Run()
}

// standIn is the stand-in rclone's whole life: it announces itself,
// then blocks until SIGINT, which it records and answers by exiting
// 130 as rclone does.
func standIn() int {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	if dir := os.Getenv(standinStartedEnv); dir != "" {
		_ = os.WriteFile(filepath.Join(dir, strconv.Itoa(os.Getpid())), nil, 0o600)
	}
	select {
	case <-sigs:
		if f, err := os.OpenFile(os.Getenv(standinRecordEnv), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600); err == nil {
			_, _ = f.WriteString("SIGINT\n")
			_ = f.Close()
		}
		return 130
	case <-time.After(standinLimit):
		return 1
	}
}

// A standin is the files the stand-in writes for one test.
type standin struct {
	record  string
	started string
}

// useStandin gives the stand-ins this test and its children start
// fresh files.
func useStandin(t *testing.T) standin {
	t.Helper()
	dir := t.TempDir()
	s := standin{record: filepath.Join(dir, "record"), started: filepath.Join(dir, "started")}
	if err := os.Mkdir(s.started, 0o700); err != nil {
		t.Fatalf("stand-in start directory: %v", err)
	}
	t.Setenv(standinRecordEnv, s.record)
	t.Setenv(standinStartedEnv, s.started)
	return s
}

// awaitStarted waits until n stand-ins have installed their SIGINT
// handlers.
func (s standin) awaitStarted(t *testing.T, n int) {
	t.Helper()
	end := time.Now().Add(20 * time.Second)
	for {
		entries, err := os.ReadDir(s.started)
		if err != nil {
			t.Fatalf("read stand-in start directory: %v", err)
		}
		if len(entries) >= n {
			return
		}
		if time.Now().After(end) {
			t.Fatalf("%d stand-ins started in 20s, want %d", len(entries), n)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// sigints counts the SIGINTs the stand-ins recorded.
func (s standin) sigints(t *testing.T) int {
	t.Helper()
	data, err := os.ReadFile(s.record)
	if errors.Is(err, os.ErrNotExist) {
		return 0
	}
	if err != nil {
		t.Fatalf("read stand-in record: %v", err)
	}
	return strings.Count(string(data), "SIGINT\n")
}

// A caller is the verified identity a request carries, and whether it
// is an admin: what a session middleware would put in the context.
type caller struct {
	id    identity.Identity
	admin bool
}

var (
	annAdmin  = caller{id: ann, admin: true}
	annMember = caller{id: ann}
)

// withCaller puts c on ctx as the mount's session middleware would:
// a network actor anchored at annAnchor, granted admin to an admin and
// member to anyone else.
func withCaller(ctx context.Context, c caller) context.Context {
	granted := ops.PrivilegeMember
	if c.admin {
		granted = ops.PrivilegeAdmin
	}
	return withAdmitted(ctx, ops.NetworkActor(c.id, annAnchor), granted)
}

// A denyAll authorizer denies every caller.
type denyAll struct{}

func (denyAll) authorize(context.Context, string) (ops.Actor, ops.Privilege, error) {
	return ops.Actor{}, 0, errors.New("every caller is denied")
}

// A setup is what a contract case's fixture needs.
type setup struct {
	// backup turns backup integration on.
	backup bool

	// auth is both legs' authorizer; nil means the contextAuthorizer
	// the mount uses.
	auth authorizer

	// admin is both legs' admin seam; nil means a fake that succeeds
	// at every operation.
	admin *fakeAdmin

	// caller is the identity both legs' requests carry.
	caller caller

	// running starts a backup that blocks in the stand-in, and waits
	// until the stand-in runs.
	running bool

	// holdWrite holds the runner's write pool's one connection open
	// once the fixture is ready.
	holdWrite bool

	// closeStore closes the runner's store handle once the fixture is
	// ready, leaving the runner live.
	closeStore bool
}

// A contract is one case's fixture: a runner over a fresh store, and
// the two legs over it.
type contract struct {
	store   *store.Handle
	runner  *local.Runner
	caller  caller
	seam    *authorizingSeam
	handler http.Handler
	server  *httptest.Server
	log     *syncBuffer
	standin standin
	job     ops.Handle
}

// newStateDir creates a current store under a fresh state directory
// and returns the directory, with no handle left open.
func newStateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	h, err := store.OpenServe(context.Background(), dir, store.ServeOptions{
		BusyTimeout:   contractBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}
	return dir
}

// newLibrary returns a library root holding one file, which a backup
// over an empty catalog accepts.
func newLibrary(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "a", "one.flac"), []byte("flac"), 0o644); err != nil {
		t.Fatal(err)
	}
	return root
}

// bounded returns a context bounded for one call a test makes.
func bounded(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// defaultBackupConfig returns the reader functions [local.Config] and
// [local.BackupConfig] need, each returning R-6c11's default
// literally: DUBPLATE_JOB_RETENTION and DUBPLATE_BACKUP_ARCHIVE_RETENTION
// at 2160h (90 days), DUBPLATE_BACKUP_GUARD_FRACTION at 0.5,
// DUBPLATE_BACKUP_SNAPSHOT_KEEP at 14. This package's fixtures assert
// nothing about these values, so every construction here uses the
// default alone.
func defaultBackupConfig() (
	retention func(context.Context) (time.Duration, error),
	guardFraction func(context.Context) (float64, error),
	snapshotKeep func(context.Context) (int, error),
	archiveRetention func(context.Context) (time.Duration, error),
) {
	return func(context.Context) (time.Duration, error) { return 2160 * time.Hour, nil },
		func(context.Context) (float64, error) { return 0.5, nil },
		func(context.Context) (int, error) { return 14, nil },
		func(context.Context) (time.Duration, error) { return 2160 * time.Hour, nil }
}

// newContract builds s's fixture over a fresh store in dir, or in a
// fresh state directory when dir is empty.
func newContract(t *testing.T, dir string, s setup) *contract {
	t.Helper()
	if dir == "" {
		dir = newStateDir(t)
	}
	c := &contract{caller: s.caller, log: &syncBuffer{}, standin: useStandin(t)}
	h, err := store.OpenWrite(context.Background(), dir, contractBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	c.store = h
	t.Cleanup(func() { _ = h.Close() })

	logger := slog.New(slog.NewTextHandler(c.log, nil))
	retention, guardFraction, snapshotKeep, archiveRetention := defaultBackupConfig()
	r, err := local.New(context.Background(), local.Config{
		Store:     h,
		Retention: retention,
		Backup: local.BackupConfig{
			Enabled: s.backup, Remote: contractRemote, LibraryRoot: newLibrary(t),
			GuardFraction: guardFraction, SnapshotKeep: snapshotKeep, ArchiveRetention: archiveRetention,
		},
		Logger: logger,
	})
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}
	c.runner = r
	t.Cleanup(func() { _ = r.Close() })
	// Only one backup runs at a time, so canceling the live one lets
	// Close return.
	t.Cleanup(func() {
		var id []byte
		if err := h.QueryRow(context.Background(), `SELECT id FROM jobs WHERE state IN ('queued', 'running')`).Scan(&id); err == nil {
			r.CancelOwn(ops.Handle(hex.EncodeToString(id)))
		}
	})

	auth := s.auth
	if auth == nil {
		auth = contextAuthorizer{}
	}
	admin := s.admin
	if admin == nil {
		admin = &fakeAdmin{answers: adminSuccesses()}
	}
	subs := &fakeSubmissions{answers: submissionSuccesses()}
	if c.seam, err = newAuthorizingSeam(r, admin, subs, auth, identity.ChannelCookie); err != nil {
		t.Fatalf("newAuthorizingSeam: %v", err)
	}
	projection, err := newHandler(r, admin, subs, auth, identity.ChannelCookie, logger)
	if err != nil {
		t.Fatalf("newHandler: %v", err)
	}
	c.handler = http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		projection.ServeHTTP(w, req.WithContext(withCaller(req.Context(), c.caller)))
	})
	c.server = httptest.NewServer(c.handler)
	t.Cleanup(c.server.Close)

	if s.running {
		resp, err := r.StartBackup(bounded(t), ops.LocalActor("tester"), ops.StartBackupRequest{})
		if err != nil {
			t.Fatalf("start the running backup: %v", err)
		}
		c.job = resp.JobID
		c.standin.awaitStarted(t, 1)
	}
	if s.holdWrite {
		c.holdWrite(t)
	}
	if s.closeStore {
		if err := h.Close(); err != nil {
			t.Fatalf("close the runner's store: %v", err)
		}
	}
	return c
}

// holdWrite holds the write pool's one connection until cleanup.
func (c *contract) holdWrite(t *testing.T) {
	t.Helper()
	held, release := make(chan struct{}), make(chan struct{})
	done := make(chan error, 1)
	// Update requires a deadline; this one outlasts any test.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	go func() {
		defer cancel()
		done <- c.store.Update(ctx, func(context.Context, *store.Tx) error {
			close(held)
			<-release
			return nil
		})
	}()
	select {
	case <-held:
	case err := <-done:
		t.Fatalf("hold the write connection: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("the write connection was not held in 10s")
	}
	t.Cleanup(func() {
		close(release)
		if err := <-done; err != nil {
			t.Errorf("held write: %v", err)
		}
	})
}

// countJobs counts every job record.
func (c *contract) countJobs(t *testing.T) int {
	t.Helper()
	var n int
	if err := c.store.QueryRow(bounded(t), "SELECT count(*) FROM jobs").Scan(&n); err != nil {
		t.Fatalf("count jobs: %v", err)
	}
	return n
}

// post sends one operation over the server and returns the answer.
func (c *contract) post(t *testing.T, ctx context.Context, op, body string) *httptest.ResponseRecorder {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.server.URL+"/"+op, strings.NewReader(body))
	if err != nil {
		t.Fatalf("build request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.server.Client().Do(req)
	if err != nil {
		t.Fatalf("POST /%s: %v", op, err)
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /%s answer: %v", op, err)
	}
	// A recorder carrying the answer lets errorAnswer check the error
	// shape the same way for both transports.
	rec := httptest.NewRecorder()
	maps.Copy(rec.Header(), resp.Header)
	rec.WriteHeader(resp.StatusCode)
	_, _ = rec.Write(data)
	return rec
}

// serve sends one operation straight to the handler under ctx.
func (c *contract) serve(ctx context.Context, op, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/"+op, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	return rec
}

// jobState reads the job h names over HTTP.
func (c *contract) jobState(t *testing.T, h ops.Handle) ops.JobState {
	t.Helper()
	rec := c.post(t, bounded(t), ops.OpJobStatus, `{"job_id":"`+string(h)+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("job_status %s = %d %s", h, rec.Code, rec.Body)
	}
	var resp ops.JobStatusResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode job_status %q: %v", rec.Body, err)
	}
	return resp.Job.State
}

// awaitEnd polls the job h names over HTTP until it ends.
func (c *contract) awaitEnd(t *testing.T, h ops.Handle) ops.JobState {
	t.Helper()
	end := time.Now().Add(30 * time.Second)
	for {
		st := c.jobState(t, h)
		if st.Terminal() {
			return st
		}
		if time.Now().After(end) {
			t.Fatalf("job %s still %s after 30s", h, st)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// A legs is one case's call on both legs: the in-process call and the
// operation and body the HTTP leg posts.
type legs struct {
	inProcess func(ctx context.Context, s ops.Seam) error
	op        string
	body      string
}

// A contractCase is one row of the contract: a fixture, a call on both
// legs, and the code both must classify to.
type contractCase struct {
	name     string
	want     ops.Code
	sentinel error
	status   int
	setup    setup
	call     func(c *contract) legs

	// cancelAfter, when set, cancels each leg's context that long
	// after its call begins, and sends the HTTP leg straight to the
	// handler, since over a server the client sees only its own
	// context error.
	cancelAfter time.Duration

	// ownReasons marks a row whose legs send different malformed
	// input, so each leg's reason names its own.
	ownReasons bool

	// check makes the row's own assertions on both legs' answers.
	check func(t *testing.T, c *contract, in error, inElapsed time.Duration, out *httptest.ResponseRecorder, outElapsed time.Duration)
}

// wellFormed is a job handle with no record in any fixture.
const wellFormed = "0123456789abcdef0123456789abcdef"

// contractCases is the contract's table: at least one row per wire
// code. TestContractCoversEveryCode holds it to [ops.Codes].
func contractCases() []contractCase {
	startBackup := func(*contract) legs {
		return legs{
			inProcess: func(ctx context.Context, s ops.Seam) error {
				_, err := s.StartBackup(ctx, ops.Actor{}, ops.StartBackupRequest{})
				return err
			},
			op: ops.OpStartBackup, body: `{}`,
		}
	}
	jobStatus := func(h ops.Handle) func(*contract) legs {
		return func(*contract) legs {
			return legs{
				inProcess: func(ctx context.Context, s ops.Seam) error {
					_, err := s.JobStatus(ctx, ops.Actor{}, ops.JobStatusRequest{JobID: h})
					return err
				},
				op: ops.OpJobStatus, body: `{"job_id":"` + string(h) + `"}`,
			}
		}
	}
	noJobs := func(t *testing.T, c *contract, _ error, _ time.Duration, _ *httptest.ResponseRecorder, _ time.Duration) {
		if n := c.countJobs(t); n != 0 {
			t.Errorf("jobs holds %d records after both legs, want 0", n)
		}
	}
	listMembers := func(c *contract) legs {
		return legs{
			inProcess: func(ctx context.Context, _ ops.Seam) error {
				_, err := c.seam.ListMembers(ctx, ops.Actor{}, ops.ListMembersRequest{})
				return err
			},
			op: ops.OpListMembers, body: `{}`,
		}
	}
	// failingAdmin is an admin seam whose list_members answers err.
	failingAdmin := func(err error) *fakeAdmin {
		return &fakeAdmin{answers: map[string]fakeAnswer{ops.OpListMembers: {err: err}}}
	}
	memberRefused := failingAdmin(errors.New("must not be reached"))
	return []contractCase{
		{
			name: "unauthenticated", want: ops.CodeUnauthenticated, sentinel: ops.ErrUnauthenticated, status: http.StatusUnauthorized,
			setup: setup{caller: annAdmin, admin: failingAdmin(ops.NewUnauthenticatedError("sign in again"))},
			call:  listMembers,
		},
		{
			name: "reauthentication_required", want: ops.CodeReauthenticationRequired, sentinel: ops.ErrReauthenticationRequired, status: http.StatusForbidden,
			setup: setup{caller: annAdmin, admin: failingAdmin(ops.NewReauthenticationRequiredError("sign in from this browser"))},
			call:  listMembers,
		},
		{
			name: "forbidden to a member on an admin operation", want: ops.CodeForbidden, sentinel: ops.ErrForbidden, status: http.StatusForbidden,
			setup: setup{caller: annMember, admin: memberRefused},
			call:  listMembers,
			check: func(t *testing.T, _ *contract, _ error, _ time.Duration, _ *httptest.ResponseRecorder, _ time.Duration) {
				t.Helper()
				if n := len(memberRefused.calls); n != 0 {
					t.Errorf("a refused list_members reached the admin seam %d times, want 0", n)
				}
			},
		},
		{
			name: "unavailable", want: ops.CodeUnavailable, sentinel: ops.ErrUnavailable, status: http.StatusServiceUnavailable,
			setup: setup{caller: annAdmin},
			call:  startBackup,
			check: func(t *testing.T, c *contract, in error, _ time.Duration, out *httptest.ResponseRecorder, _ time.Duration) {
				t.Helper()
				// The reason ops/local answers while backup is off.
				const want = "backup integration disabled"
				if e, _ := errors.AsType[*ops.Error](in); e == nil || e.Reason != want {
					t.Errorf("in-process reason of %v, want %q", in, want)
				}
				if _, reason, _ := errorAnswer(t, out); reason != want {
					t.Errorf("HTTP reason = %q, want %q", reason, want)
				}
				noJobs(t, c, in, 0, out, 0)
			},
		},
		{
			name: "not_found", want: ops.CodeNotFound, sentinel: ops.ErrNotFound, status: http.StatusNotFound,
			setup: setup{caller: annAdmin},
			call:  jobStatus(wellFormed),
		},
		{
			name: "conflict", want: ops.CodeConflict, sentinel: ops.ErrConflict, status: http.StatusConflict,
			setup: setup{backup: true, caller: annAdmin, running: true},
			call:  startBackup,
			check: func(t *testing.T, c *contract, in error, _ time.Duration, out *httptest.ResponseRecorder, _ time.Duration) {
				t.Helper()
				if !strings.Contains(in.Error(), string(c.job)) {
					t.Errorf("in-process conflict %q does not name the running job %s", in, c.job)
				}
				if _, reason, _ := errorAnswer(t, out); !strings.Contains(reason, string(c.job)) {
					t.Errorf("HTTP conflict reason %q does not name the running job %s", reason, c.job)
				}
			},
		},
		{
			name: "invalid_request by a malformed handle", want: ops.CodeInvalidRequest, sentinel: ops.ErrInvalidRequest, status: http.StatusBadRequest,
			setup: setup{caller: annAdmin},
			call:  jobStatus("ABC"),
		},
		{
			name: "invalid_request by a malformed body", want: ops.CodeInvalidRequest, sentinel: ops.ErrInvalidRequest, status: http.StatusBadRequest,
			setup: setup{caller: annAdmin}, ownReasons: true,
			call: func(*contract) legs {
				return legs{
					inProcess: func(ctx context.Context, s ops.Seam) error {
						_, err := s.JobStatus(ctx, ops.Actor{}, ops.JobStatusRequest{JobID: "ABC"})
						return err
					},
					op: ops.OpJobStatus, body: `{"job_id":`,
				}
			},
		},
		{
			name: "forbidden by a denying authorizer", want: ops.CodeForbidden, sentinel: ops.ErrForbidden, status: http.StatusForbidden,
			setup: setup{backup: true, auth: denyAll{}, caller: annAdmin},
			call:  startBackup,
			check: noJobs,
		},
		{
			name: "forbidden by privilege", want: ops.CodeForbidden, sentinel: ops.ErrForbidden, status: http.StatusForbidden,
			setup: setup{backup: true, caller: annMember},
			call:  startBackup,
			check: noJobs,
		},
		{
			name: "canceled", want: ops.CodeCanceled, sentinel: ops.ErrCanceled, status: http.StatusServiceUnavailable,
			setup: setup{backup: true, caller: annAdmin, running: true, holdWrite: true},
			call: func(c *contract) legs {
				return legs{
					inProcess: func(ctx context.Context, s ops.Seam) error {
						_, err := s.CancelJob(ctx, ops.Actor{}, ops.CancelJobRequest{JobID: c.job})
						return err
					},
					op: ops.OpCancelJob, body: `{"job_id":"` + string(c.job) + `"}`,
				}
			},
			cancelAfter: 200 * time.Millisecond,
			check: func(t *testing.T, _ *contract, _ error, inElapsed time.Duration, _ *httptest.ResponseRecorder, outElapsed time.Duration) {
				t.Helper()
				// Each leg waited for the held connection until its
				// context was canceled, and no longer: the wait honors
				// the context, and writeWait is 30s.
				for leg, d := range map[string]time.Duration{"in-process": inElapsed, "HTTP": outElapsed} {
					if d < 200*time.Millisecond || d >= 2*time.Second {
						t.Errorf("%s leg answered after %v, want between 200ms and 2s", leg, d)
					}
				}
			},
		},
		{
			name: "internal", want: ops.CodeInternal, sentinel: nil, status: http.StatusInternalServerError,
			setup: setup{caller: annAdmin, closeStore: true},
			call:  jobStatus(wellFormed),
			check: func(t *testing.T, c *contract, in error, _ time.Duration, out *httptest.ResponseRecorder, _ time.Duration) {
				t.Helper()
				// database/sql's answer on a closed *DB (sql.go's
				// errDBClosed).
				const closed = "sql: database is closed"
				if !strings.Contains(in.Error(), closed) {
					t.Errorf("in-process error %q lacks %q", in, closed)
				}
				if _, _, hasReason := errorAnswer(t, out); hasReason {
					t.Errorf("HTTP body %q carries a reason, want none", out.Body)
				}
				if !strings.Contains(c.log.String(), closed) {
					t.Errorf("log %q lacks %q", c.log.String(), closed)
				}
			},
		},
	}
}

// TestContract runs every case on both legs. Each leg's answer must
// classify to the row's code, and the two must agree on it; a coded
// answer's HTTP reason must be the in-process error's reason.
func TestContract(t *testing.T) {
	for _, tt := range contractCases() {
		t.Run(tt.name, func(t *testing.T) {
			c := newContract(t, "", tt.setup)
			l := tt.call(c)

			inCtx, inCancel := context.WithCancel(withCaller(bounded(t), c.caller))
			defer inCancel()
			if tt.cancelAfter > 0 {
				time.AfterFunc(tt.cancelAfter, inCancel)
			}
			begin := time.Now()
			in := l.inProcess(inCtx, c.seam)
			inElapsed := time.Since(begin)

			outCtx, outCancel := context.WithCancel(bounded(t))
			defer outCancel()
			begin = time.Now()
			var out *httptest.ResponseRecorder
			if tt.cancelAfter > 0 {
				time.AfterFunc(tt.cancelAfter, outCancel)
				out = c.serve(outCtx, l.op, l.body)
			} else {
				out = c.post(t, outCtx, l.op, l.body)
			}
			outElapsed := time.Since(begin)

			if in == nil {
				t.Fatalf("in-process %s answered no error, want %s", l.op, tt.want)
			}
			inCode := ops.CodeOf(in)
			if inCode != tt.want {
				t.Errorf("in-process code = %q (%v), want %q", inCode, in, tt.want)
			}
			if tt.sentinel != nil && !errors.Is(in, tt.sentinel) {
				t.Errorf("in-process error %v does not match %v", in, tt.sentinel)
			}
			if out.Code != tt.status {
				t.Errorf("HTTP status = %d, want %d; body %q", out.Code, tt.status, out.Body)
			}
			outCode, outReason, _ := errorAnswer(t, out)
			if outCode != tt.want {
				t.Errorf("HTTP code = %q, want %q", outCode, tt.want)
			}
			if inCode != outCode {
				t.Errorf("legs disagree: in-process %q, HTTP %q", inCode, outCode)
			}
			if tt.want != ops.CodeInternal && !tt.ownReasons {
				var inReason string
				if e, ok := errors.AsType[*ops.Error](in); ok {
					inReason = e.Reason
				}
				if outReason != inReason {
					t.Errorf("HTTP reason = %q, in-process reason = %q", outReason, inReason)
				}
			}
			if tt.check != nil {
				tt.check(t, c, in, inElapsed, out, outElapsed)
			}
		})
	}
}

// TestContractCoversEveryCode fails, naming the code, when a wire code
// package ops declares has no contract case.
func TestContractCoversEveryCode(t *testing.T) {
	covered := make(map[ops.Code]bool)
	for _, tt := range contractCases() {
		covered[tt.want] = true
	}
	for _, code := range ops.Codes() {
		if !covered[code] {
			t.Errorf("wire code %q has no contract case", code)
		}
	}
}

// TestContractActor starts a backup whose body names an actor, in any
// case of the key, and one whose in-process caller passes its own; the
// job records the authorizer's actor every time.
func TestContractActor(t *testing.T) {
	tests := []struct {
		name  string
		start func(t *testing.T, c *contract) ops.Handle
	}{
		{"actor key in the body", func(t *testing.T, c *contract) ops.Handle {
			return c.startOverHTTP(t, `{"actor":"local:root"}`)
		}},
		{"ACTOR key in the body", func(t *testing.T, c *contract) ops.Handle {
			return c.startOverHTTP(t, `{"ACTOR":"local:root"}`)
		}},
		{"in-process caller's own actor", func(t *testing.T, c *contract) ops.Handle {
			resp, err := c.seam.StartBackup(withCaller(bounded(t), c.caller), ops.LocalActor("root"), ops.StartBackupRequest{})
			if err != nil {
				t.Fatalf("StartBackup: %v", err)
			}
			return resp.JobID
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newContract(t, "", setup{backup: true, caller: annAdmin})
			h := tt.start(t, c)
			var actor string
			if err := c.store.QueryRow(bounded(t), "SELECT actor FROM jobs WHERE lower(hex(id)) = ?", string(h)).Scan(&actor); err != nil {
				t.Fatalf("read job %s's actor: %v", h, err)
			}
			if actor != annWire {
				t.Errorf("jobs.actor = %q, want %q", actor, annWire)
			}
		})
	}
}

// startOverHTTP starts a backup over HTTP with body and returns its
// handle.
func (c *contract) startOverHTTP(t *testing.T, body string) ops.Handle {
	t.Helper()
	rec := c.post(t, bounded(t), ops.OpStartBackup, body)
	if rec.Code != http.StatusOK {
		t.Fatalf("start_backup %s = %d %s", body, rec.Code, rec.Body)
	}
	var resp ops.StartBackupResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode start_backup %q: %v", rec.Body, err)
	}
	return resp.JobID
}

// TestContractJobPath starts a backup over HTTP, polls it running,
// cancels it over HTTP, and polls it to its end: canceled, with its
// transfer interrupted once.
func TestContractJobPath(t *testing.T) {
	c := newContract(t, "", setup{backup: true, caller: annAdmin})
	h := c.startOverHTTP(t, `{}`)
	c.standin.awaitStarted(t, 1)

	if st := c.jobState(t, h); st != ops.JobRunning {
		t.Fatalf("job %s polled %s, want running", h, st)
	}
	rec := c.post(t, bounded(t), ops.OpCancelJob, `{"job_id":"`+string(h)+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel_job = %d %s", rec.Code, rec.Body)
	}
	if st := c.awaitEnd(t, h); st != ops.JobCanceled {
		t.Errorf("job %s ended %s, want canceled", h, st)
	}
	if n := c.standin.sigints(t); n != 1 {
		t.Errorf("the stand-in recorded %d SIGINTs, want 1", n)
	}
}

// runChild is a second process's whole life: it opens a runner over
// dir's store, starts a backup of root, writes the job's handle to
// descriptor 3, and exits 0 once the job has ended.
func runChild(dir, root string) int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	fail := func(err error) int {
		logger.Error("child", "err", err)
		return 2
	}
	ctx := context.Background()
	h, err := store.OpenWrite(ctx, dir, contractBusy)
	if err != nil {
		return fail(err)
	}
	defer func() { _ = h.Close() }()
	retention, guardFraction, snapshotKeep, archiveRetention := defaultBackupConfig()
	r, err := local.New(ctx, local.Config{
		Store:     h,
		Retention: retention,
		Backup: local.BackupConfig{
			Enabled: true, Remote: contractRemote, LibraryRoot: root,
			GuardFraction: guardFraction, SnapshotKeep: snapshotKeep, ArchiveRetention: archiveRetention,
		},
		Logger: logger,
	})
	if err != nil {
		return fail(err)
	}
	defer func() { _ = r.Close() }()
	resp, err := r.StartBackup(ctx, ops.LocalActor("child"), ops.StartBackupRequest{})
	if err != nil {
		return fail(err)
	}
	job := resp.JobID
	ready := os.NewFile(3, "ready")
	if _, err := fmt.Fprintln(ready, job); err != nil {
		return fail(err)
	}
	if err := ready.Close(); err != nil {
		return fail(err)
	}
	end := time.Now().Add(time.Minute)
	for time.Now().Before(end) {
		st, err := r.JobStatus(ctx, ops.LocalActor("child"), ops.JobStatusRequest{JobID: job})
		if err != nil {
			return fail(err)
		}
		if st.Job.State.Terminal() {
			return 0
		}
		time.Sleep(20 * time.Millisecond)
	}
	return fail(errors.New("the child's job never ended"))
}

// TestContractCancelsAnotherProcessesJob cancels, over HTTP through a
// runner in this process, a backup a re-executed child runs over the
// same store. The cancel answers the job running; the child's runner
// ends it canceled, interrupting its transfer once.
func TestContractCancelsAnotherProcessesJob(t *testing.T) {
	dir := newStateDir(t)
	s := useStandin(t)

	readyR, readyW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = readyR.Close() }()
	var stderr syncBuffer
	child := exec.Command(os.Args[0], "-test.run=^$")
	child.Env = append(os.Environ(), childDirEnv+"="+dir, childRootEnv+"="+newLibrary(t))
	child.ExtraFiles = []*os.File{readyW}
	child.Stderr = &stderr
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := child.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	_ = readyW.Close()
	exited := make(chan error, 1)
	go func() { exited <- child.Wait() }()
	t.Cleanup(func() {
		_ = syscall.Kill(-child.Process.Pid, syscall.SIGKILL)
		// SIGKILL makes a nonzero exit expected; log it, never fail on it.
		if err := <-exited; err != nil {
			t.Logf("child exit: %v", err)
		}
	})

	line := make(chan string, 1)
	go func() {
		l, _ := bufio.NewReader(readyR).ReadString('\n')
		line <- strings.TrimSpace(l)
	}()
	var h ops.Handle
	select {
	case l := <-line:
		if _, err := ops.ParseHandle(l); err != nil {
			t.Fatalf("child's ready line %q: %v; stderr:\n%s", l, err, stderr.String())
		}
		h = ops.Handle(l)
	case <-time.After(30 * time.Second):
		t.Fatalf("child never became ready; stderr:\n%s", stderr.String())
	}
	s.awaitStarted(t, 1)

	c := newContract(t, dir, setup{caller: annAdmin})
	c.standin = s
	rec := c.post(t, bounded(t), ops.OpCancelJob, `{"job_id":"`+string(h)+`"}`)
	if rec.Code != http.StatusOK {
		t.Fatalf("cancel_job = %d %s", rec.Code, rec.Body)
	}
	var resp ops.CancelJobResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode cancel_job %q: %v", rec.Body, err)
	}
	if resp.Job.ID != h || resp.Job.State != ops.JobRunning {
		t.Errorf("cancel_job answered %s %s, want %s running", resp.Job.ID, resp.Job.State, h)
	}

	select {
	case err := <-exited:
		exited <- err
		if err != nil {
			t.Fatalf("child exited %v; stderr:\n%s", err, stderr.String())
		}
	case <-time.After(30 * time.Second):
		t.Fatalf("child still running 30s after the cancel; stderr:\n%s", stderr.String())
	}
	if st := c.jobState(t, h); st != ops.JobCanceled {
		t.Errorf("the child's job reads %s, want canceled", st)
	}
	if n := s.sigints(t); n != 1 {
		t.Errorf("the child's stand-in recorded %d SIGINTs, want 1", n)
	}
}

// syncBuffer is a bytes.Buffer safe for the concurrent writes a
// server's and a job's loggers make.
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}
