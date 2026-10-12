package settings

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// testBusy is every test store's busy timeout.
const testBusy = 5 * time.Second

// testNow stamps every row an apply writes in these tests.
var testNow = time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC)

// The test settings' keys.
const (
	keyKeep      = "DUBPLATE_TEST_KEEP"
	keyRetention = "DUBPLATE_TEST_RETENTION"
	keyExpiry    = "DUBPLATE_TEST_EXPIRY"
	keyNonce     = "DUBPLATE_TEST_NONCE_RETENTION"
	keyOrphan    = "DUBPLATE_TEST_ORPHAN"
)

// testDescription is the description every test setting carries.
const testDescription = "a test setting"

// keepSetting is an integer setting shaped like the backup snapshot
// keep: default def, bounds 3 to 1000.
func keepSetting(def string) config.Setting {
	return config.Setting{
		Key:     keyKeep,
		Kind:    config.KindInt,
		Default: def,
		Doc:     "an integer",
		Apply:   config.ApplyNextOperation,
		Floor:   "3",
		Ceiling: "1000",

		Description: testDescription,
		Mark:        config.NotSensitive,
	}
}

// retentionSetting is a duration setting shaped like the job
// retention: default 2160h, bounds 24h to 87600h.
var retentionSetting = config.Setting{
	Key:     keyRetention,
	Kind:    config.KindDuration,
	Default: "2160h",
	Doc:     "a Go duration such as 720h",
	Apply:   config.ApplyNextOperation,
	Floor:   "24h",
	Ceiling: "87600h",

	Description: testDescription,
	Mark:        config.NotSensitive,
}

// newRegistry returns a registry declaring the keep setting with
// default 14, the retention setting, and the expiry setting.
func newRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddSetting(keepSetting("14"))
	reg.AddSetting(retentionSetting)
	reg.AddSetting(expirySetting)
	return reg
}

// ctxFor returns a context carrying the deadline every store write
// needs.
func ctxFor(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

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

// newStore returns a write-mode handle on a fresh current store.
func newStore(t *testing.T) *store.Handle {
	t.Helper()
	return openWrite(t, newStateDir(t))
}

// fixtureRow is a settings row a test writes before the code under
// test runs.
type fixtureRow struct {
	key, value, source, seed string
}

// writeFixture writes rows as literal settings rows and sets the
// generation to fixtureGeneration, in one transaction, as every settings
// writer must. The source is spelled in the SQL text, never bound.
func writeFixture(t *testing.T, h *store.Handle, rows ...fixtureRow) {
	t.Helper()
	err := h.Update(ctxFor(t), func(ctx context.Context, tx *store.Tx) error {
		for _, r := range rows {
			var q string
			switch r.source {
			case "admin":
				q = `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
					VALUES (?, ?, 'admin', ?, 'fixture', 1)`
			case "env-seed":
				q = `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
					VALUES (?, ?, 'env-seed', ?, 'fixture', 1)`
			default:
				return fmt.Errorf("fixture source %q", r.source)
			}
			var seed any
			if r.seed != "" {
				seed = r.seed
			}
			if _, err := tx.Exec(ctx, q, r.key, r.value, seed); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `UPDATE settings_generation SET generation = ? WHERE singleton = 1`, fixtureGeneration)
		return err
	})
	if err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

// storedRow is one settings row's key, value, source, and seed as a
// test reads them back; seed is "<NULL>" for a NULL seed_value.
type storedRow struct {
	key, value, source, seed string
}

// readSettings returns every settings row in key order.
func readSettings(t *testing.T, h *store.Handle) []storedRow {
	t.Helper()
	rs, err := h.Query(context.Background(), `SELECT key, value, source, seed_value FROM settings ORDER BY key`)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	defer func() { _ = rs.Close() }()
	var rows []storedRow
	for rs.Next() {
		var r storedRow
		var seed sql.NullString
		if err := rs.Scan(&r.key, &r.value, &r.source, &seed); err != nil {
			t.Fatalf("scan settings: %v", err)
		}
		r.seed = cmpNull(seed)
		rows = append(rows, r)
	}
	if err := rs.Err(); err != nil {
		t.Fatalf("read settings: %v", err)
	}
	return rows
}

// auditRow is one settings_audit row as a test reads it back; the old
// and new values are "<NULL>" for NULL.
type auditRow struct {
	at                                        int64
	kind, actor, channel, key, oldVal, newVal string
}

// readAudit returns every settings_audit row, oldest first.
func readAudit(t *testing.T, h *store.Handle) []auditRow {
	t.Helper()
	rs, err := h.Query(context.Background(),
		`SELECT at, kind, actor, channel, key, old_value, new_value FROM settings_audit ORDER BY at, key`)
	if err != nil {
		t.Fatalf("read audit: %v", err)
	}
	defer func() { _ = rs.Close() }()
	var rows []auditRow
	for rs.Next() {
		var r auditRow
		var oldVal, newVal sql.NullString
		if err := rs.Scan(&r.at, &r.kind, &r.actor, &r.channel, &r.key, &oldVal, &newVal); err != nil {
			t.Fatalf("scan audit: %v", err)
		}
		r.oldVal, r.newVal = cmpNull(oldVal), cmpNull(newVal)
		rows = append(rows, r)
	}
	if err := rs.Err(); err != nil {
		t.Fatalf("read audit: %v", err)
	}
	return rows
}

// cmpNull renders a nullable column for comparison.
func cmpNull(s sql.NullString) string {
	if !s.Valid {
		return "<NULL>"
	}
	return s.String
}

// readGeneration returns the store's settings generation.
func readGeneration(t *testing.T, h *store.Handle) int64 {
	t.Helper()
	var g int64
	if err := h.QueryRow(context.Background(), `SELECT generation FROM settings_generation WHERE singleton = 1`).Scan(&g); err != nil {
		t.Fatalf("read generation: %v", err)
	}
	return g
}

// countAdmin returns how many settings rows carry the admin source.
func countAdmin(t *testing.T, h *store.Handle) int {
	t.Helper()
	var n int
	if err := h.QueryRow(context.Background(), `SELECT count(*) FROM settings WHERE source = 'admin'`).Scan(&n); err != nil {
		t.Fatalf("count admin rows: %v", err)
	}
	return n
}

// discardLog is a logger that drops every record.
var discardLog = slog.New(slog.DiscardHandler)

// bufferLog returns a logger writing text records into the returned
// buffer.
func bufferLog() (*slog.Logger, *bytes.Buffer) {
	var buf bytes.Buffer
	return slog.New(slog.NewTextHandler(&buf, nil)), &buf
}

// wantErrText fails t unless err is non-nil and its text holds every
// part.
func wantErrText(t *testing.T, err error, parts ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("err = nil, want an error naming %q", parts)
	}
	for _, p := range parts {
		if !strings.Contains(err.Error(), p) {
			t.Errorf("err = %q, want it to name %q", err, p)
		}
	}
}

// errIsNoRows reports whether err wraps sql.ErrNoRows.
func errIsNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }
