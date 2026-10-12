package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/settings"
)

// The production settings these tests plant rows for.
const (
	retentionKey = "DUBPLATE_JOB_RETENTION"
	keepKey      = "DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP"
	expiryKey    = "DUBPLATE_MAGIC_LINK_EXPIRY"
	nonceKey     = "DUBPLATE_NONCE_RETENTION"
)

// runSettings runs the settings verb with args against environ and
// returns what it wrote to stdout beside its error.
func runSettings(t *testing.T, environ []string, args ...string) (stdout string, err error) {
	t.Helper()
	var out bytes.Buffer
	cmd := newSettingsCmd(environ, productionRegistry(), &out, io.Discard)
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), err
}

// plantRow inserts a settings row for key into the closed store file at
// path. The source is spelled in the SQL text, as the table's check
// constraint reads it.
func plantRow(t *testing.T, path, key, value, source string) {
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

// currentStore returns environ, whose state directory holds a current
// store, and the store file's path.
func currentStore(t *testing.T) (environ []string, stateDir, path string) {
	t.Helper()
	environ = serveAuthEnviron(t)
	stateDir = environValue(environ, "DUBPLATE_STATE_DIR")
	path = olderStore(t, stateDir, storeReport(t, t.TempDir()).BinaryVersion)
	return environ, stateDir, path
}

// asOperator makes the verbs record "operator" as the operating-system
// user for the length of the test.
func asOperator(t *testing.T) {
	t.Helper()
	orig := osUserLookupID
	t.Cleanup(func() { osUserLookupID = orig })
	osUserLookupID = func(string) (*user.User, error) { return &user.User{Username: "operator"}, nil }
}

// requireServeRefuses fails t unless starting serve over environ is
// refused with an error naming every part.
func requireServeRefuses(t *testing.T, environ []string, parts ...string) {
	t.Helper()
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err == nil || shell != nil {
		t.Fatalf("serve start = %v, %v, want a refusal", shell, err)
	}
	for _, p := range parts {
		if !strings.Contains(err.Error(), p) {
			t.Errorf("serve refusal = %q, want it to contain %q", err, p)
		}
	}
}

// requireServeStarts fails t unless serve starts over environ.
func requireServeStarts(t *testing.T, environ []string) {
	t.Helper()
	if shell, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil || shell == nil {
		t.Fatalf("serve start = %v, %v, want a shell", shell, err)
	}
}

// TestSettingsResetCmd_ResolvesEachRefusal is AC14's first sentence:
// on a current store, `settings reset` clears each refusal the seed
// rule raises over an admin row, and serve then starts. Fixture: one
// populated admin row per case, serve refusing first, so a reset that
// changed nothing leaves serve refusing; the audit row is the
// command-line reset's.
func TestSettingsResetCmd_ResolvesEachRefusal(t *testing.T) {
	tests := []struct {
		name  string
		value string
		// serveEnv is the environment serve starts under, and resetEnv
		// the one the verb runs under (a key the environment sets is
		// never reset).
		serveEnv  []string
		resetEnv  []string
		wantRefus string
		// wantRow is the row left after the reset and the start.
		wantRow settingRow
		wantHas bool
	}{
		{
			name:      "a stored value the binary cannot parse",
			value:     "soon",
			wantRefus: "invalid value",
		},
		{
			name:      "a value outside the binary's bounds",
			value:     "100000h",
			wantRefus: "ceiling",
		},
		{
			name:      "an environment that contradicts the row, reset with the key unset",
			value:     "720h",
			serveEnv:  []string{retentionKey + "=48h"},
			wantRefus: "unset " + retentionKey,
			wantRow:   settingRow{"48h", "env-seed"},
			wantHas:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asOperator(t)
			environ, stateDir, path := currentStore(t)
			plantRow(t, path, retentionKey, tt.value, "admin")
			serveEnviron := append(slices.Clip(environ), tt.serveEnv...)
			requireServeRefuses(t, serveEnviron, retentionKey, tt.wantRefus, "settings reset "+retentionKey)
			_, baseAudit, baseGen := readSeedState(t, stateDir)

			stdout, err := runSettings(t, environ, "reset", retentionKey)
			if err != nil {
				t.Fatalf("reset error = %v, want nil", err)
			}
			if !strings.Contains(stdout, retentionKey) {
				t.Errorf("stdout = %q, want it to name %s", stdout, retentionKey)
			}

			rows, audit, gen := readSeedState(t, stateDir)
			if _, has := rows[retentionKey]; has {
				t.Errorf("row after reset = %v, want it deleted", rows[retentionKey])
			}
			if got, want := strings.Join(audit[len(baseAudit):], "|"), "reset operator command-line "+retentionKey; got != want {
				t.Errorf("new audit rows = %q, want %q", got, want)
			}
			if gen != baseGen+1 {
				t.Errorf("generation = %d, want %d", gen, baseGen+1)
			}

			requireServeStarts(t, serveEnviron)
			rows, _, _ = readSeedState(t, stateDir)
			if got, has := rows[retentionKey]; has != tt.wantHas || got != tt.wantRow {
				t.Errorf("row after serve = %v (present %v), want %v (present %v)", got, has, tt.wantRow, tt.wantHas)
			}
		})
	}
}

// TestSettingsResetCmd_DeletesRowTheEnvironmentContradicts proves the
// verb deletes an admin row for a key its own environment sets, as the
// design says the command line does where the web refuses. Fixture: an
// admin row the environment contradicts, so a verb that refused it
// would leave the row and fail the test.
func TestSettingsResetCmd_DeletesRowTheEnvironmentContradicts(t *testing.T) {
	environ, stateDir, path := currentStore(t)
	plantRow(t, path, retentionKey, "720h", "admin")
	environ = append(environ, retentionKey+"=48h")

	if _, err := runSettings(t, environ, "reset", retentionKey); err != nil {
		t.Fatalf("reset: %v", err)
	}

	rows, audit, _ := readSeedState(t, stateDir)
	if got, has := rows[retentionKey]; has {
		t.Errorf("row after the reset = %v, want it deleted", got)
	}
	if len(audit) != 1 {
		t.Errorf("audit after the reset = %v, want one row", audit)
	}
}

// TestSettingsResetCmd_TwoOutOfBoundsRows is AC14's second sentence: a
// store holding two out-of-bounds admin rows takes a reset of each, one
// at a time, and serve starts only after the second. Fixture: two
// populated admin rows, each refusing alone.
func TestSettingsResetCmd_TwoOutOfBoundsRows(t *testing.T) {
	environ, stateDir, path := currentStore(t)
	plantRow(t, path, retentionKey, "100000h", "admin")
	plantRow(t, path, keepKey, "500", "admin")
	requireServeRefuses(t, environ, "ceiling")

	if _, err := runSettings(t, environ, "reset", retentionKey); err != nil {
		t.Fatalf("first reset: %v", err)
	}
	requireServeRefuses(t, environ, keepKey)
	if rows, _, _ := readSeedState(t, stateDir); len(rows) != 1 {
		t.Fatalf("rows after the first reset = %v, want only the keep row", rows)
	}

	if _, err := runSettings(t, environ, "reset", keepKey); err != nil {
		t.Fatalf("second reset: %v", err)
	}
	requireServeStarts(t, environ)
}

// TestSettingsResetCmd_RuleRefusal proves the verb runs under the
// validator: resetting the expiry restores its 15m default, above the
// 10m retention, which breaks the nonce rule, so the reset is refused
// and nothing changes. Fixture: an admin expiry of 5m and an admin
// retention of 10m, a pair serve starts under.
func TestSettingsResetCmd_RuleRefusal(t *testing.T) {
	environ, stateDir, path := currentStore(t)
	plantRow(t, path, expiryKey, "5m", "admin")
	plantRow(t, path, nonceKey, "10m", "admin")
	requireServeStarts(t, environ)
	rowsBefore, auditBefore, genBefore := readSeedState(t, stateDir)

	_, err := runSettings(t, environ, "reset", expiryKey)

	if !errors.Is(err, settings.ErrInvalidValue) {
		t.Fatalf("reset error = %v, want it to wrap ErrInvalidValue", err)
	}
	if !strings.Contains(err.Error(), "nonce retention covers the link expiry") {
		t.Errorf("reset error = %q, want it to name the rule", err)
	}
	rows, audit, gen := readSeedState(t, stateDir)
	if fmt.Sprint(rows) != fmt.Sprint(rowsBefore) || fmt.Sprint(audit) != fmt.Sprint(auditBefore) || gen != genBefore {
		t.Errorf("store changed by a refused reset: rows %v -> %v, audit %v -> %v, generation %d -> %d",
			rowsBefore, rows, auditBefore, audit, genBefore, gen)
	}
}

// TestSettingsResetCmd_EnvSeedRowStays proves a reset over an env-seed
// row writes nothing and says which source is in force, and that the
// refusal such a row raises does not point at the verb. Fixture: an
// env-seed row that does not parse.
func TestSettingsResetCmd_EnvSeedRowStays(t *testing.T) {
	environ, stateDir, path := currentStore(t)
	plantRow(t, path, retentionKey, "soon", "env-seed")
	requireServeRefuses(t, environ, retentionKey, "stored env-seed value refused")
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err == nil || strings.Contains(err.Error(), "settings reset") {
		t.Fatalf("serve start = %v, %v, want a refusal that does not name settings reset", shell, err)
	}
	_, auditBefore, genBefore := readSeedState(t, stateDir)

	stdout, err := runSettings(t, environ, "reset", retentionKey)

	if err != nil {
		t.Fatalf("reset error = %v, want nil", err)
	}
	if !strings.Contains(stdout, "env-seed") {
		t.Errorf("stdout = %q, want it to name the env-seed source in force", stdout)
	}
	rows, audit, gen := readSeedState(t, stateDir)
	if got := rows[retentionKey]; got != (settingRow{"soon", "env-seed"}) {
		t.Errorf("row after reset = %v, want the env-seed row kept", got)
	}
	if len(audit) != len(auditBefore) || gen != genBefore {
		t.Errorf("audit %v -> %v, generation %d -> %d, want nothing written", auditBefore, audit, genBefore, gen)
	}
}

// TestSettingsResetCmd_RefusesUnknownKeys proves the verb answers a key
// that is no store-homed setting with the validator's refusal.
func TestSettingsResetCmd_RefusesUnknownKeys(t *testing.T) {
	environ, _, _ := currentStore(t)
	for _, key := range []string{"DUBPLATE_NO_SUCH_SETTING", "DUBPLATE_STATE_DIR"} {
		_, err := runSettings(t, environ, "reset", key)
		if !errors.Is(err, settings.ErrNotSetting) {
			t.Errorf("reset %s error = %v, want it to wrap ErrNotSetting", key, err)
		}
	}
	if _, err := runSettings(t, environ, "reset"); err == nil {
		t.Error("reset with no key error = nil, want an argument refusal")
	}
}

// TestSettingsResetCmd_StoreStates is AC14's refusal sentence: the verb
// refuses an older store, a newer store, a restore intent, and an
// absent store with OpenWrite's own words, as member does, and leaves
// the store as it was.
func TestSettingsResetCmd_StoreStates(t *testing.T) {
	head := storeReport(t, t.TempDir()).BinaryVersion
	reconcilePending := `{"format": 1, "state": "reconcile-pending", "stamp": "20260925T120000.000000Z",
"source": {"remote": "remote:dubplate", "snapshot": "20260924T030000.000000Z"},
"incoming": "restore-20260925T120000.000000Z.db", "pre_restore": "pre-restore-20260925T120000.000000Z.db",
"discard_current": false, "library": null, "version": 2}`
	tests := []struct {
		name    string
		setup   func(t *testing.T, stateDir string)
		wantErr string
	}{
		{
			name:    "absent",
			setup:   func(*testing.T, string) {},
			wantErr: "start serve first",
		},
		{
			name: "older than the binary",
			setup: func(t *testing.T, stateDir string) {
				plantRow(t, olderStore(t, stateDir, head-1), keepKey, "500", "admin")
			},
			wantErr: "start serve to migrate",
		},
		{
			name: "newer than the binary",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, head)
				plantRow(t, path, keepKey, "500", "admin")
				execStoreSQL(t, path, fmt.Sprintf("PRAGMA user_version = %d", head+1))
			},
			wantErr: fmt.Sprintf("version %d is newer than this binary's version %d", head+1, head),
		},
		{
			name: "a restore-intent record",
			setup: func(t *testing.T, stateDir string) {
				plantRow(t, olderStore(t, stateDir, head), keepKey, "500", "admin")
				if err := os.WriteFile(filepath.Join(stateDir, "store", "restore-intent"), []byte(reconcilePending), 0o600); err != nil {
					t.Fatalf("write the restore-intent record: %v", err)
				}
			},
			wantErr: "restore-intent",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := serveAuthEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			tt.setup(t, stateDir)
			db := storeDBPath(stateDir)
			var before [32]byte
			if _, err := os.Stat(db); err == nil {
				before = fileSum(t, db)
			}

			stdout, err := runSettings(t, environ, "reset", keepKey)

			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("reset error = %v, want it to contain %q", err, tt.wantErr)
			}
			if stdout != "" {
				t.Errorf("stdout = %q, want nothing beside a refusal", stdout)
			}
			if _, statErr := os.Stat(db); statErr == nil && fileSum(t, db) != before {
				t.Error("the refused reset changed the store file")
			}
		})
	}
}

// TestSettingsResetCmd_ServingHandleReadsTheDefault proves a reset
// reaches a process that holds the store open: an Accessor over a
// serve-mode handle reads the admin value, the verb resets it from its
// own handle, and the Accessor then reads the default. Fixture: an
// admin row of 50, written through the serving handle.
func TestSettingsResetCmd_ServingHandleReadsTheDefault(t *testing.T) {
	environ := serveAuthEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	serving, err := store.OpenServe(t.Context(), stateDir, store.ServeOptions{
		BusyTimeout:   testBackupBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	t.Cleanup(func() { _ = serving.Close() })
	reg := productionRegistry()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	author := settings.Author{Actor: "member:00112233445566778899aabbccddeeff", Channel: "cookie"}
	if _, err := settings.Set(ctx, serving, reg, nil, author, nil, keepKey, "50", time.Now()); err != nil {
		t.Fatalf("Set: %v", err)
	}
	acc := settings.NewAccessor(serving, reg)
	if got, err := acc.Int(ctx, storeLocalSnapshotKeepSetting); err != nil || got != 50 {
		t.Fatalf("accessor before the reset = %d, %v, want 50", got, err)
	}

	if _, err := runSettings(t, environ, "reset", keepKey); err != nil {
		t.Fatalf("reset: %v", err)
	}

	if got, err := acc.Int(ctx, storeLocalSnapshotKeepSetting); err != nil || got != 3 {
		t.Errorf("accessor after the reset = %d, %v, want the default 3", got, err)
	}
}

// TestSettingsResetCmd_AuditRow proves the verb's audit row: the
// command-line actor, the command-line channel, the key, the replaced
// value, and a NULL new value. Fixture: an admin row of 50.
func TestSettingsResetCmd_AuditRow(t *testing.T) {
	asOperator(t)
	environ, stateDir, path := currentStore(t)
	plantRow(t, path, keepKey, "50", "admin")

	if _, err := runSettings(t, environ, "reset", keepKey); err != nil {
		t.Fatalf("reset: %v", err)
	}

	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	var kind, actor, channel, oldValue string
	var newIsNull bool
	err = h.QueryRow(t.Context(), `SELECT kind, actor, channel, old_value, new_value IS NULL FROM settings_audit WHERE key = ?`, keepKey).
		Scan(&kind, &actor, &channel, &oldValue, &newIsNull)
	if err != nil {
		t.Fatalf("read the audit row: %v", err)
	}
	if kind != "reset" || actor != "operator" || channel != "command-line" || oldValue != "50" || !newIsNull {
		t.Errorf("audit row = %q %q %q old %q new NULL %v, want reset operator command-line old 50 new NULL",
			kind, actor, channel, oldValue, newIsNull)
	}
}

// TestPrepareServe_OlderStoreOverAdminRowNamesTheRecovery is AC14's
// last sentence: serve's pre-open refusal of an older store over an
// admin row names the previous-release recovery, and the same refusal
// over an environment value alone does not. Fixture: an older store
// holding an admin keep row outside its bounds, then one holding none
// beside a keep value in the environment that does not parse.
func TestPrepareServe_OlderStoreOverAdminRowNamesTheRecovery(t *testing.T) {
	const recovery = "redeploy the previous release"
	t.Run("an admin row", func(t *testing.T) {
		environ := serveAuthEnviron(t)
		stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
		plantKeepRow(t, olderStore(t, stateDir, 4), "500", "admin")

		requireServeRefuses(t, environ, keepKey, "ceiling", recovery, "settings reset", "then upgrade")
	})
	t.Run("an environment value alone", func(t *testing.T) {
		environ := append(serveAuthEnviron(t), keepKey+"=many")
		olderStore(t, environValue(environ, "DUBPLATE_STATE_DIR"), 4)

		shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
		if err == nil || shell != nil || strings.Contains(err.Error(), recovery) {
			t.Fatalf("serve start = %v, %v, want a refusal that does not name the recovery", shell, err)
		}
	})
}
