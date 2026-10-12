package admin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
)

// testNow is the instant every seam call here is evaluated at.
var testNow = time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)

// testWindow is the re-authentication window the fixture's reader
// answers. fresh is anchored inside it and stale outside it.
const testWindow = 2 * time.Hour

var (
	fresh = testNow.Add(-time.Hour)
	stale = testNow.Add(-3 * time.Hour)
)

// testBusy is every fixture store's busy timeout: far above any hold a
// concurrency fixture takes, so correct code never answers busy.
const testBusy = 5 * time.Second

// The fixture's members. dave is a removed admin: his row keeps the
// flag under its tombstone.
type member struct {
	id                         identity.UserID
	address, username, display string
	admin, tombstoned          bool
	hex                        string
}

var (
	alice = member{id: identity.UserID{0xa1}, address: "alice@example.com", username: "alice", display: "Alice", admin: true,
		hex: "a1000000000000000000000000000000"}
	bob = member{id: identity.UserID{0xb0}, address: "bob@example.com", username: "bob", display: "Bob", admin: true,
		hex: "b0000000000000000000000000000000"}
	carol = member{id: identity.UserID{0xc0}, address: "carol@example.com", username: "carol", display: "Carol",
		hex: "c0000000000000000000000000000000"}
	dave = member{id: identity.UserID{0xd0}, address: "dave@example.com", username: "dave", display: "Dave", admin: true, tombstoned: true,
		hex: "d0000000000000000000000000000000"}
)

// The test settings' keys: one sensitive, one not.
const (
	keyLifetime = "DUBPLATE_TEST_LIFETIME"
	keyKeep     = "DUBPLATE_TEST_KEEP"
)

var (
	lifetimeSetting = config.Setting{
		Key: keyLifetime, Kind: config.KindDuration, Default: "1h", Doc: "a Go duration",
		Apply: config.ApplyNextOperation, Floor: "1m", Ceiling: "24h",
		Description: "how long a test session lasts", Mark: config.Sensitive,
	}
	keepSetting = config.Setting{
		Key: keyKeep, Kind: config.KindInt, Default: "14", Doc: "an integer",
		Apply: config.ApplyRestart, Floor: "3", Ceiling: "1000",
		Description: "how many test snapshots to keep", Mark: config.NotSensitive,
	}
)

// newRegistry returns a registry declaring the two test settings.
func newRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddSetting(lifetimeSetting)
	reg.AddSetting(keepSetting)
	return reg
}

// A sent is one message the recording mailer took.
type sent struct {
	to, subject, body string
}

// A mailbox records every message a seam sends. err fails each send;
// when gate is non-nil, each send signals entered and then waits for
// gate to close, recording the context's state once released.
type mailbox struct {
	mu      sync.Mutex
	sent    []sent
	err     error
	gate    chan struct{}
	entered chan struct{}

	releasedErr      error
	releasedDeadline bool
}

func (m *mailbox) send(ctx context.Context, to, subject, body string) error {
	if m.gate != nil {
		m.entered <- struct{}{}
		<-m.gate
		m.mu.Lock()
		m.releasedErr = ctx.Err()
		_, m.releasedDeadline = ctx.Deadline()
		m.mu.Unlock()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, sent{to, subject, body})
	return nil
}

// recipients returns every address the mailbox sent to, sorted.
func (m *mailbox) recipients() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var to []string
	for _, s := range m.sent {
		to = append(to, s.to)
	}
	slices.Sort(to)
	return to
}

// A fixture is a seam over a fresh store holding alice, bob, carol, and
// dave, with the pieces a test reads back.
type fixture struct {
	seam  *Seam
	h     *store.Handle
	dir   string
	reg   *config.Registry
	mail  *mailbox
	logs  *bytes.Buffer
	epoch *auth.EpochStore
}

// fixtureOptions vary the fixture's seam: the store's busy timeout and
// the window reader.
type fixtureOptions struct {
	busy   time.Duration
	window auth.Setting[time.Duration]
	reg    *config.Registry
	env    []string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureWith(t, fixtureOptions{})
}

func newFixtureWith(t *testing.T, opts fixtureOptions) *fixture {
	t.Helper()
	if opts.busy == 0 {
		opts.busy = testBusy
	}
	if opts.window == nil {
		opts.window = func(context.Context) (time.Duration, error) { return testWindow, nil }
	}
	if opts.reg == nil {
		opts.reg = newRegistry()
	}
	dir := t.TempDir()
	created, err := store.OpenServe(context.Background(), dir, store.ServeOptions{
		BusyTimeout: testBusy, MigrationWait: 30 * time.Second, Keep: 1, Logger: slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if err := created.Close(); err != nil {
		t.Fatal(err)
	}
	h, err := store.OpenWrite(context.Background(), dir, opts.busy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	insertMembers(t, h, alice, bob, carol, dave)

	authStore, err := auth.Open(t.TempDir())
	if err != nil {
		t.Fatalf("auth.Open: %v", err)
	}
	epoch := auth.OpenEpochStore(authStore)
	logs := &bytes.Buffer{}
	mail := &mailbox{}
	seam := New(Config{
		Store:    h,
		Registry: opts.reg,
		Environ:  opts.env,
		Roster:   auth.NewRoster(h, epoch, func() time.Time { return testNow }),
		Window:   opts.window,
		Send:     mail.send,
		Now:      func() time.Time { return testNow },
		Logger:   slog.New(slog.NewTextHandler(&lockedWriter{w: logs}, nil)),
	})
	return &fixture{seam: seam, h: h, dir: dir, reg: opts.reg, mail: mail, logs: logs, epoch: epoch}
}

// lockedWriter serializes writes to w, so a log written from a goroutine
// a concurrency fixture starts does not race the test's read.
type lockedWriter struct {
	mu sync.Mutex
	w  *bytes.Buffer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

// actor returns the network actor m's session carries, anchored at
// anchor.
func actor(m member, anchor time.Time) ops.Actor {
	return ops.NetworkActor(identity.Identity{
		Address: m.address, UserID: m.id, Principal: m.username, DisplayName: m.display,
		Source: "magic-link", Channel: identity.ChannelCookie,
	}, anchor)
}

// writeCtx returns a context carrying the deadline every store write
// needs.
func writeCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func insertMembers(t *testing.T, h *store.Handle, members ...member) {
	t.Helper()
	exec(t, h, func(ctx context.Context, tx *store.Tx) error {
		for i, m := range members {
			var tombstonedAt any
			if m.tombstoned {
				tombstonedAt = int64(i + 100)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, m.id[:], m.address, m.display, m.username, m.admin, tombstonedAt, i+1, i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

// exec runs fn as one direct store write, failing t on error.
func exec(t *testing.T, h *store.Handle, fn func(ctx context.Context, tx *store.Tx) error) {
	t.Helper()
	if err := h.Update(writeCtx(t), fn); err != nil {
		t.Fatalf("direct store write: %v", err)
	}
}

// execSQL runs one statement as a direct store write.
func execSQL(t *testing.T, h *store.Handle, stmt string, args ...any) {
	t.Helper()
	exec(t, h, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, stmt, args...)
		return err
	})
}

// plantAdminRow writes value as key's admin row from the command line,
// the way an earlier admin write would leave it.
func plantAdminRow(t *testing.T, f *fixture, key, value string) {
	t.Helper()
	by := settings.Author{Actor: "local:test", Channel: identity.ChannelCommandLine.String()}
	if _, err := settings.Set(writeCtx(t), f.h, f.reg, nil, by, nil, key, value, testNow); err != nil {
		t.Fatalf("plant %s=%s: %v", key, value, err)
	}
}

// snapshot renders every row a seam write could change: the users, the
// settings rows, the audit rows, the generation, and the preferences.
// Two equal snapshots prove nothing was written between them.
func snapshot(t *testing.T, h *store.Handle) string {
	t.Helper()
	ctx := context.Background()
	var b strings.Builder
	query := func(q string, cols int) {
		rows, err := h.Query(ctx, q)
		if err != nil {
			t.Fatalf("snapshot %q: %v", q, err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			vals := make([]any, cols)
			ptrs := make([]any, cols)
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			fmt.Fprintln(&b, vals...)
		}
		if err := rows.Err(); err != nil {
			t.Fatal(err)
		}
	}
	query(`SELECT hex(id), address, username, admin, tombstoned_at FROM users ORDER BY id`, 5)
	query(`SELECT key, value, source, seed_value, actor FROM settings ORDER BY key`, 5)
	query(`SELECT kind, actor, channel, key, old_value, new_value FROM settings_audit ORDER BY at, kind`, 6)
	query(`SELECT generation FROM settings_generation`, 1)
	query(`SELECT count(*) FROM member_preferences`, 1)
	return b.String()
}

// auditRow is one settings_audit row as a test compares it.
type auditRow struct {
	kind, actor, channel, key string
	oldValue, newValue        *string
}

func audits(t *testing.T, h *store.Handle) []auditRow {
	t.Helper()
	rows, err := h.Query(context.Background(), `SELECT kind, actor, channel, key, old_value, new_value FROM settings_audit ORDER BY at, rowid`)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var out []auditRow
	for rows.Next() {
		var a auditRow
		if err := rows.Scan(&a.kind, &a.actor, &a.channel, &a.key, &a.oldValue, &a.newValue); err != nil {
			t.Fatal(err)
		}
		out = append(out, a)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// storedValue returns key's settings row value and source, or two empty
// strings when the store holds none.
func storedValue(t *testing.T, h *store.Handle, key string) (value, source string) {
	t.Helper()
	rows, err := h.Query(context.Background(), `SELECT value, source FROM settings WHERE key = ?`, key)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	if rows.Next() {
		if err := rows.Scan(&value, &source); err != nil {
			t.Fatal(err)
		}
	}
	return value, source
}

// userState returns m's admin flag and whether m's row is tombstoned.
func userState(t *testing.T, h *store.Handle, m member) (admin, tombstoned bool) {
	t.Helper()
	if err := h.QueryRow(context.Background(), `SELECT admin, tombstoned_at IS NOT NULL FROM users WHERE id = ?`, m.id[:]).Scan(&admin, &tombstoned); err != nil {
		t.Fatalf("read %s: %v", m.username, err)
	}
	return admin, tombstoned
}

// wantCode fails t unless err carries code and, when detail is set,
// that detail.
func wantCode(t *testing.T, err error, code ops.Code, detail string) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want code %s", code)
	}
	if got := ops.CodeOf(err); got != code {
		t.Fatalf("code = %s (%v), want %s", got, err, code)
	}
	opsErr, ok := errors.AsType[*ops.Error](err)
	if !ok {
		if detail != "" {
			t.Fatalf("err %v carries no *ops.Error, want detail %q", err, detail)
		}
		return
	}
	if opsErr.Detail != detail {
		t.Errorf("detail = %q, want %q", opsErr.Detail, detail)
	}
}

// whileHeld runs stmt on other inside an Update it holds open, starts
// write once the write lock is held, commits stmt after lockHold, and
// returns write's error.
func whileHeld(t *testing.T, other *store.Handle, stmt string, args []any, write func() error) error {
	t.Helper()
	locked := make(chan struct{})
	release := make(chan struct{})
	held := make(chan error, 1)
	go func() {
		held <- other.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
			if _, err := tx.Exec(ctx, stmt, args...); err != nil {
				return err
			}
			close(locked)
			<-release
			return nil
		})
	}()
	select {
	case <-locked:
	case err := <-held:
		t.Fatalf("the held write ended before taking the lock: %v", err)
	}
	done := make(chan error, 1)
	go func() { done <- write() }()
	time.Sleep(lockHold)
	close(release)
	if err := <-held; err != nil {
		t.Fatalf("the held write: %v", err)
	}
	return <-done
}

// lockHold is how long a concurrency fixture holds the write lock:
// well inside testBusy.
const lockHold = 300 * time.Millisecond

// secondHandle opens another write handle on f's store.
func (f *fixture) secondHandle(t *testing.T) *store.Handle {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), f.dir, testBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}
