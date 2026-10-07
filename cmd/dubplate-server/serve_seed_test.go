package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/settings"
)

// settingRow is a stored settings row as the seed tests compare it.
type settingRow struct {
	value, source string
}

// readSeedState reads the closed or shared store under stateDir: the
// settings rows by key, the audit rows as "kind actor channel key",
// and the generation.
func readSeedState(t *testing.T, stateDir string) (rows map[string]settingRow, audit []string, generation int64) {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	ctx := context.Background()

	rs, err := h.Query(ctx, "SELECT key, value, source FROM settings")
	if err != nil {
		t.Fatal(err)
	}
	rows = map[string]settingRow{}
	for rs.Next() {
		var key string
		var r settingRow
		if err := rs.Scan(&key, &r.value, &r.source); err != nil {
			t.Fatal(err)
		}
		rows[key] = r
	}
	if err := rs.Err(); err != nil {
		t.Fatal(err)
	}
	_ = rs.Close()

	as, err := h.Query(ctx, "SELECT kind, actor, channel, key FROM settings_audit ORDER BY at, key")
	if err != nil {
		t.Fatal(err)
	}
	for as.Next() {
		var kind, actor, channel, key string
		if err := as.Scan(&kind, &actor, &channel, &key); err != nil {
			t.Fatal(err)
		}
		audit = append(audit, fmt.Sprintf("%s %s %s %s", kind, actor, channel, key))
	}
	if err := as.Err(); err != nil {
		t.Fatal(err)
	}
	_ = as.Close()

	if err := h.QueryRow(ctx, "SELECT generation FROM settings_generation WHERE singleton = 1").Scan(&generation); err != nil {
		t.Fatal(err)
	}
	return rows, audit, generation
}

// plantSettingRow inserts a settings row into the closed store file at
// path. The source is spelled in the SQL text, as the table's check
// constraint reads it.
func plantSettingRow(t *testing.T, path, key, value, source string) {
	t.Helper()
	var q string
	switch source {
	case "admin":
		q = `INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES (?, ?, 'admin', ?, 'fixture', 1)`
	case "env-seed":
		q = `INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES (?, ?, 'env-seed', ?, 'fixture', 1)`
	default:
		t.Fatalf("source %q", source)
	}
	execStoreSQL(t, path, q, key, value, value)
}

// TestPrepareServe_AppliesTheSeed is the seed's start states, each
// through the preparation function the verb calls: when it returns a
// shell, the seed's rows are already committed. Fixture: every store is
// current, created through OpenServe, with the rows each case names, so
// the generation baseline is read from the fixture itself and the
// unchanged case asserts it, never only a nil error.
func TestPrepareServe_AppliesTheSeed(t *testing.T) {
	const (
		retention = "DUBPLATE_JOB_RETENTION"
		keep      = "DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP"
	)
	tests := []struct {
		name      string
		plant     func(t *testing.T, path string)
		env       []string
		wantRows  map[string]settingRow
		wantAudit []string
		wantBump  bool
		wantErr   []string
	}{
		{
			name:      "env, no row: seeded and audited",
			env:       []string{retention + "=720h"},
			wantRows:  map[string]settingRow{retention: {"720h", "env-seed"}},
			wantAudit: []string{"seed system env-seed " + retention},
			wantBump:  true,
		},
		{
			name:     "env equal to the stored env-seed row: nothing written",
			plant:    func(t *testing.T, path string) { plantSettingRow(t, path, retention, "720h", "env-seed") },
			env:      []string{retention + "=720h"},
			wantRows: map[string]settingRow{retention: {"720h", "env-seed"}},
		},
		{
			name:      "env absent, an env-seed row present: removed and audited",
			plant:     func(t *testing.T, path string) { plantSettingRow(t, path, retention, "720h", "env-seed") },
			wantRows:  map[string]settingRow{},
			wantAudit: []string{"remove system env-seed " + retention},
			wantBump:  true,
		},
		{
			name:  "an admin row, env different: refused naming both values and the way out",
			plant: func(t *testing.T, path string) { plantSettingRow(t, path, retention, "720h", "admin") },
			env:   []string{retention + "=48h"},
			wantErr: []string{retention, `"48h"`, `"720h"`,
				"unset " + retention + `, or set it to "720h"`},
			wantRows: map[string]settingRow{retention: {"720h", "admin"}},
		},
		{
			name:     "an admin row outside the bounds, env unset: refused naming the key and the bound",
			plant:    func(t *testing.T, path string) { plantSettingRow(t, path, keep, "500", "admin") },
			wantErr:  []string{keep, "above the ceiling 100"},
			wantRows: map[string]settingRow{keep: {"500", "admin"}},
		},
		{
			name:     "env that does not parse: refused naming the key and the accepted form",
			env:      []string{retention + "=soon"},
			wantErr:  []string{retention, `invalid value "soon"`, "a Go duration"},
			wantRows: map[string]settingRow{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(serveAuthEnviron(t), tt.env...)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			path := olderStore(t, stateDir, storeReport(t, t.TempDir()).BinaryVersion)
			if tt.plant != nil {
				tt.plant(t, path)
			}
			_, baseAudit, baseGen := readSeedState(t, stateDir)

			shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
			if tt.wantErr != nil {
				if err == nil {
					t.Fatal("prepare error = nil, want a refusal")
				}
				if shell != nil {
					t.Error("prepare returned a shell beside its refusal")
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %q, want it to contain %q", err, want)
					}
				}
			} else if err != nil || shell == nil {
				t.Fatalf("prepare = %v, %v, want a shell", shell, err)
			}

			rows, audit, gen := readSeedState(t, stateDir)
			if fmt.Sprint(rows) != fmt.Sprint(tt.wantRows) {
				t.Errorf("settings = %v, want %v", rows, tt.wantRows)
			}
			if got, want := strings.Join(audit[len(baseAudit):], "|"), strings.Join(tt.wantAudit, "|"); got != want {
				t.Errorf("new audit rows = %q, want %q", got, want)
			}
			wantGen := baseGen
			if tt.wantBump {
				wantGen++
			}
			if gen != wantGen {
				t.Errorf("generation = %d, want %d", gen, wantGen)
			}
		})
	}
}

// TestPrepareServe_SeedLogsItsChange proves each committed change logs
// one line, so an operator reading the journal sees what the start did.
func TestPrepareServe_SeedLogsItsChange(t *testing.T) {
	environ := append(serveAuthEnviron(t), "DUBPLATE_JOB_RETENTION=720h")
	var logs bytes.Buffer
	if _, err := buildServeShellWithLogger(t, environ, slog.New(slog.NewTextHandler(&logs, nil))); err != nil {
		t.Fatal(err)
	}
	if got := logs.String(); !strings.Contains(got, "setting seed") || !strings.Contains(got, "DUBPLATE_JOB_RETENTION") {
		t.Errorf("log = %q, want a seed line naming the key", got)
	}
}

// TestPrepareServe_RefusalReleasesTheStore proves a refusal after the
// open (the seed's own refusal, which only the open store can reach)
// closes the handle: an exclusive open, the one restore takes,
// then succeeds. A preparation function that leaks its handle on the
// error path keeps the shared lock and fails this open.
func TestPrepareServe_RefusalReleasesTheStore(t *testing.T) {
	const key = "DUBPLATE_JOB_RETENTION"
	environ := append(serveAuthEnviron(t), key+"=48h")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	path := olderStore(t, stateDir, storeReport(t, t.TempDir()).BinaryVersion)
	plantSettingRow(t, path, key, "720h", "admin")

	if shell, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err == nil || shell != nil {
		t.Fatalf("prepare = %v, %v, want a refusal", shell, err)
	}
	sess, err := store.OpenRestore(stateDir, store.RestoreOptions{
		BusyTimeout: testBackupBusy, Now: time.Now(), Logger: discardTestLogger(),
	})
	if err != nil {
		t.Fatalf("exclusive open after the refusal: %v", err)
	}
	if err := sess.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestPrepareServe_SeedMakesTheStringsTrue renders doctor's settings
// row and backup's evaluation log for an env value that differs from
// the stored env-seed row, before and after the preparation function
// runs. Before, the stored value is in force and each says serve
// applies the env's at its next start; after, the row reads the env's
// value in force and the log is silent, because the start applied it.
func TestPrepareServe_SeedMakesTheStringsTrue(t *testing.T) {
	const key = "DUBPLATE_JOB_RETENTION"
	environ := append(serveAuthEnviron(t), key+"=48h")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	path := olderStore(t, stateDir, storeReport(t, t.TempDir()).BinaryVersion)
	plantSettingRow(t, path, key, "720h", "env-seed")
	reg := productionRegistry()

	doctorRow := func() string {
		t.Helper()
		report := storeReport(t, stateDir)
		rows, failed := settingsDoctorRows(environ, reg, report, true)
		if failed {
			t.Fatalf("doctor rows refuse: %v", rows)
		}
		for _, r := range rows {
			if r.key == key {
				return r.status
			}
		}
		t.Fatalf("no doctor row for %s in %v", key, rows)
		return ""
	}
	backupLog := func() string {
		t.Helper()
		report := storeReport(t, stateDir)
		var logs bytes.Buffer
		logSettingsEvaluation(slog.New(slog.NewTextHandler(&logs, nil)), settings.EvaluateFacts(reg, environ, *report.Facts))
		return logs.String()
	}

	before := doctorRow()
	for _, want := range []string{`"720h" (source env-seed`, `env "48h" re-seeds at serve's next start`} {
		if !strings.Contains(before, want) {
			t.Errorf("doctor row before = %q, want it to contain %q", before, want)
		}
	}
	if got := backupLog(); !strings.Contains(got, "serve applies it at its next start") || !strings.Contains(got, key) {
		t.Errorf("backup log before = %q, want the disagreement warning for %s", got, key)
	}

	if shell, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil || shell == nil {
		t.Fatalf("prepare = %v, %v, want a shell", shell, err)
	}

	after := doctorRow()
	if !strings.Contains(after, `"48h" (source env-seed`) || strings.Contains(after, "serve's next start") {
		t.Errorf("doctor row after = %q, want the env's value in force and no pending note", after)
	}
	if got := backupLog(); got != "" {
		t.Errorf("backup log after = %q, want none", got)
	}
}
