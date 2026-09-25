package local

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// testBusy is every test store's busy timeout.
const testBusy = 5 * time.Second

// testActor is the actor every test start names.
var testActor = ops.LocalActor("tester")

// newStateDir creates a current store under a fresh state directory
// and returns the directory, with no handle left open.
func newStateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	h, err := store.OpenServe(context.Background(), dir, store.ServeOptions{
		BusyTimeout:   testBusy,
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

// openWrite opens dir's store in write mode, closed at cleanup.
func openWrite(t *testing.T, dir string) *store.Handle {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), dir, testBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// newStore returns a write-mode handle on a fresh current store, and
// the store's state directory.
func newStore(t *testing.T) (string, *store.Handle) {
	t.Helper()
	dir := newStateDir(t)
	return dir, openWrite(t, dir)
}

// leasesDir returns the leases/ directory of dir's store.
func leasesDir(dir string) string { return filepath.Join(dir, "store", "leases") }

// leaseNames lists leases/ in dir's store.
func leaseNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(leasesDir(dir))
	if err != nil {
		t.Fatalf("read leases: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// syncBuffer is a bytes.Buffer safe for the concurrent writes a job
// goroutine's logger makes.
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

// newRunnerFor opens a runner over h with backup on and body as the
// backup body, logging to logs when it is non-nil, and closes it at
// cleanup.
func newRunnerFor(t *testing.T, h *store.Handle, logs *syncBuffer, b body) *Runner {
	t.Helper()
	var logger *slog.Logger
	if logs != nil {
		logger = slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.DiscardHandler)
	}
	r, err := newRunner(context.Background(), Config{Store: h, Backup: BackupConfig{Enabled: true}, Logger: logger}, b)
	if err != nil {
		t.Fatalf("newRunner: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// deadline returns a context bounded for a test's store call.
func deadline(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// A record is a job row a test plants or reads. A zero finished or
// started reads as NULL.
type record struct {
	id              []byte
	kind            string
	state           string
	owner           string
	actor           string
	at              int64
	started         int64
	updated         int64
	finished        int64
	cancelRequested bool
	progress        string
	code            string
	message         string
}

// newInstance returns a fresh owner instance no lease file names.
func newInstance() string { return hex.EncodeToString(store.NewID()) }

func nullable(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

// plant inserts rec through the write method, filling an unset id,
// kind, actor, at, and updated.
func plant(t *testing.T, h *store.Handle, rec record) []byte {
	t.Helper()
	if rec.id == nil {
		rec.id = store.NewID()
	}
	if rec.kind == "" {
		rec.kind = ops.KindBackup
	}
	if rec.actor == "" {
		rec.actor = "local:planted"
	}
	if rec.at == 0 {
		rec.at = time.Now().UnixMicro()
	}
	if rec.updated == 0 {
		rec.updated = rec.at
	}
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, started_at, updated_at, finished_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, rec.id, rec.kind, rec.at, rec.state, rec.owner, rec.actor,
			nullable(rec.started), rec.updated, nullable(rec.finished))
		return err
	})
	if err != nil {
		t.Fatalf("plant %s record: %v", rec.state, err)
	}
	return rec.id
}

// readRecord reads the record id names; ok is false when none does.
func readRecord(t *testing.T, h *store.Handle, id []byte) (rec record, ok bool) {
	t.Helper()
	var (
		started, finished       sql.NullInt64
		progress, code, message sql.NullString
	)
	err := h.QueryRow(deadline(t), `SELECT id, kind, state, owner_instance, actor, at, started_at, updated_at, finished_at,
	cancel_requested, progress, failure_code, failure_message FROM jobs WHERE id = ?`, id).Scan(
		&rec.id, &rec.kind, &rec.state, &rec.owner, &rec.actor, &rec.at, &started, &rec.updated, &finished,
		&rec.cancelRequested, &progress, &code, &message)
	if err == sql.ErrNoRows {
		return record{}, false
	}
	if err != nil {
		t.Fatalf("read record %x: %v", id, err)
	}
	rec.started, rec.finished = started.Int64, finished.Int64
	rec.progress, rec.code, rec.message = progress.String, code.String, message.String
	return rec, true
}

// mustRecord reads the record id names, failing when none does.
func mustRecord(t *testing.T, h *store.Handle, id []byte) record {
	t.Helper()
	rec, ok := readRecord(t, h, id)
	if !ok {
		t.Fatalf("record %x absent", id)
	}
	return rec
}

// countRecords counts every job record.
func countRecords(t *testing.T, h *store.Handle) int {
	t.Helper()
	var n int
	if err := h.QueryRow(deadline(t), "SELECT count(*) FROM jobs").Scan(&n); err != nil {
		t.Fatalf("count records: %v", err)
	}
	return n
}

// idOf decodes a handle a test holds.
func idOf(t *testing.T, h ops.Handle) []byte {
	t.Helper()
	id, err := hex.DecodeString(string(h))
	if err != nil {
		t.Fatalf("decode handle %q: %v", h, err)
	}
	return id
}

// waitState polls id's record until it reads want, failing after
// limit.
func waitState(t *testing.T, h *store.Handle, id []byte, want string, limit time.Duration) record {
	t.Helper()
	end := time.Now().Add(limit)
	for {
		rec := mustRecord(t, h, id)
		if rec.state == want {
			return rec
		}
		if time.Now().After(end) {
			t.Fatalf("record %x reads %s after %v, want %s", id, rec.state, limit, want)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// onlyLease reports whether leases/ holds exactly name.
func onlyLease(t *testing.T, dir, name string) bool {
	t.Helper()
	return slices.Equal(leaseNames(t, dir), []string{name})
}
