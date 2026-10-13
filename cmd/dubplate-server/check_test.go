package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// dbFileState returns the store's database file's SHA-256 sum and mode
// bits, and whether its -wal file holds any frames (absent or 0 bytes
// counts as none), so a test proves check's read-only claim (criterion
// 14): the state before and after a run must match exactly.
func dbFileState(t *testing.T, stateDir string) (sum [sha256.Size]byte, mode os.FileMode, walEmpty bool) {
	t.Helper()
	dbPath := filepath.Join(stateDir, "store", "db", "dubplate.db")
	data, err := os.ReadFile(dbPath)
	if err != nil {
		t.Fatalf("read %s: %v", dbPath, err)
	}
	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat %s: %v", dbPath, err)
	}
	walInfo, err := os.Stat(dbPath + "-wal")
	switch {
	case err == nil:
		walEmpty = walInfo.Size() == 0
	case os.IsNotExist(err):
		walEmpty = true
	default:
		t.Fatalf("stat %s-wal: %v", dbPath, err)
	}
	return sha256.Sum256(data), info.Mode(), walEmpty
}

// TestCheckCmd_SettingsRow_Reseed is spec criterion 14: an env value
// that disagrees with a seeded row prints the disagreement and leaves
// the database file untouched, since check never applies the rule.
// Fixture state: an env-seed row of 14 exists so the run has something
// to compare the env's 20 against; an apply would write here if check
// called it.
func TestCheckCmd_SettingsRow_Reseed(t *testing.T) {
	stateDir := t.TempDir()
	seedBackupStore(t, stateDir)
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	seedSettingRow(t, h, "DUBPLATE_BACKUP_SNAPSHOT_KEEP", "14", "env-seed")
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	beforeSum, beforeMode, beforeWALEmpty := dbFileState(t, stateDir)
	if !beforeWALEmpty {
		t.Fatalf("fixture needs a checkpointed -wal before the first hash")
	}

	environ := replaceKey(authEnviron(t), "DUBPLATE_STATE_DIR", stateDir)
	environ = append(environ, "DUBPLATE_BACKUP_SNAPSHOT_KEEP=20")

	var stdout, stderr bytes.Buffer
	cmd := newCheckCmd(environ, productionRegistry(), &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil: a disagreement is not a refusal", err)
	}
	if !strings.Contains(stdout.String(), "settings: DUBPLATE_BACKUP_SNAPSHOT_KEEP:") {
		t.Errorf("stdout = %q, want a settings line naming the disagreeing key", stdout.String())
	}

	afterSum, afterMode, afterWALEmpty := dbFileState(t, stateDir)
	if afterSum != beforeSum || afterMode != beforeMode {
		t.Errorf("database changed: sum/mode before (%x, %v), after (%x, %v)", beforeSum, beforeMode, afterSum, afterMode)
	}
	if !afterWALEmpty {
		t.Error("the -wal holds frames after check, want it empty: check must never write")
	}
}

// TestCheckCmd_SettingsRow_AdminConflict is Ruling 2's exit-code half:
// an env value that disagrees with an admin-set row refuses, and check
// exits nonzero naming it.
func TestCheckCmd_SettingsRow_AdminConflict(t *testing.T) {
	stateDir := t.TempDir()
	seedBackupStore(t, stateDir)
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	seedSettingRow(t, h, "DUBPLATE_BACKUP_GUARD_FRACTION", "0.9", "admin")
	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	environ := replaceKey(authEnviron(t), "DUBPLATE_STATE_DIR", stateDir)
	environ = append(environ, "DUBPLATE_BACKUP_GUARD_FRACTION=0.3")

	var stdout, stderr bytes.Buffer
	cmd := newCheckCmd(environ, productionRegistry(), &stdout, &stderr)
	err = cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want a refusal: the environment disagrees with the administrator's value")
	}
	if !strings.Contains(stdout.String(), "settings: DUBPLATE_BACKUP_GUARD_FRACTION:") {
		t.Errorf("stdout = %q, want a settings line naming the refusing key", stdout.String())
	}
}

// freshRegistry returns the registry a test builds over baseEnviron
// alone: rung 1's own rows plus the staging row baseEnviron sets, and
// no other rung's. config.NewRegistry on its own would meet the staging
// entry as an unknown key.
func freshRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddWallRow(stagingRow)
	return reg
}

// baseEnviron returns a minimal environ with every owned-now key set
// to a valid value, mirroring internal/config's own test fixture, so a
// case can override or add exactly the entries it needs.
func baseEnviron() []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STAGING_DIR=/srv/music/staging",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}
}

// reservedPort returns a loopback port the test binary reserved once
// (listened on 127.0.0.1:0, recorded, closed) and never listens on
// again, so a doctor run that dials its configured serve port reaches
// no listener, never the host's own 127.0.0.1:8080.
var reservedPort = sync.OnceValue(func() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("reserve a loopback port: %v", err))
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		panic(fmt.Sprintf("reserve a loopback port: address is %T, want *net.TCPAddr", l.Addr()))
	}
	port := addr.Port
	if err := l.Close(); err != nil {
		panic(fmt.Sprintf("release the reserved port: %v", err))
	}
	return port
})

// authEnviron returns baseEnviron plus a full, satisfied magic-link
// identity source: the selector and the source's eight required keys,
// on fixture values visibly synthetic rather than shaped like a real
// secret, so gitleaks' generic-secret rule (a key-like name beside a
// high-entropy neighbor) never fires against a test file. A test
// against productionRegistry (serve_test.go) that expects config.Load
// to succeed needs this, rather than baseEnviron alone, because rung 5
// registers DUBPLATE_IDENTITY_SOURCE as an owned wall row (auth.go).
func authEnviron(t *testing.T) []string {
	t.Helper()
	return append(baseEnviron(),
		"DUBPLATE_IDENTITY_SOURCE=magic-link",
		"DUBPLATE_MAGIC_LINK_SECRET=fixture magic link secret, not a real key",
		"DUBPLATE_SESSION_SECRET=fixture session secret, not a real key",
		"DUBPLATE_PUBLIC_URL=https://dubplate.example.test",
		"DUBPLATE_SMTP_HOST=smtp.example.test",
		"DUBPLATE_SMTP_PORT=587",
		"DUBPLATE_SMTP_USER=dubplate",
		"DUBPLATE_SMTP_PASSWORD=fixture smtp password, not a real key",
		"DUBPLATE_SMTP_FROM=dubplate@example.test",
	)
}

// dropKey returns a copy of environ with any entry named name removed.
func dropKey(environ []string, name string) []string {
	out := make([]string, 0, len(environ))
	for _, kv := range environ {
		n, _, _ := strings.Cut(kv, "=")
		if n == name {
			continue
		}
		out = append(out, kv)
	}
	return out
}

func TestCheckCmd(t *testing.T) {
	tests := []struct {
		name         string
		environ      []string
		tempStateDir bool // replace DUBPLATE_STATE_DIR with t.TempDir(): a case whose store the run inspects must never read the host's own state directory.
		wantOK       bool
		wantStdout   string
		errHas       []string
	}{
		{
			name:         "full valid wall",
			environ:      baseEnviron(),
			tempStateDir: true,
			wantOK:       true,
			wantStdout:   "configuration: ok\nstore: absent\n",
		},
		{
			name:    "missing STATE_DIR",
			environ: dropKey(baseEnviron(), "DUBPLATE_STATE_DIR"),
			wantOK:  false,
			errHas:  []string{"DUBPLATE_STATE_DIR"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := tt.environ
			if tt.tempStateDir {
				environ = replaceKey(environ, "DUBPLATE_STATE_DIR", t.TempDir())
			}

			var stdout, stderr bytes.Buffer
			cmd := newCheckCmd(environ, freshRegistry(), &stdout, &stderr)
			cmd.SetOut(&stdout)
			err := cmd.Execute()

			if tt.wantOK {
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				if got := stdout.String(); got != tt.wantStdout {
					t.Errorf("stdout = %q, want %q", got, tt.wantStdout)
				}
				if stderr.Len() != 0 {
					t.Errorf("stderr = %q, want empty", stderr.String())
				}
				return
			}

			if err == nil {
				t.Fatal("Execute() error = nil, want error")
			}
			if stdout.Len() != 0 {
				t.Errorf("stdout = %q, want empty on failure", stdout.String())
			}
			// check itself must not print the load error: main is the
			// sole stderr writer, so this command's own stderr stays
			// empty and the error surfaces only through the return value.
			if stderr.Len() != 0 {
				t.Errorf("stderr = %q, want empty; check must not print its own error", stderr.String())
			}
			for _, sub := range tt.errHas {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("error = %q, want it to contain %q", err.Error(), sub)
				}
			}
			if strings.Contains(stdout.String(), "Usage:") {
				t.Errorf("stdout = %q, want no usage block", stdout.String())
			}
		})
	}
}

// TestCheckCmd_EmptyEnvironment proves the acceptance criterion pinned
// in the rung 1 plan word for word: an empty environment yields the
// load error alone, with no usage block, printed by main exactly
// once (proved separately by TestRootCmd_UnknownSubcommand in
// main_test.go and the build-and-run probe); check's own stdout and
// stderr stay empty.
func TestCheckCmd_EmptyEnvironment(t *testing.T) {
	reg := freshRegistry()
	_, loadErr := config.Load(nil, reg)
	if loadErr == nil {
		t.Fatal("config.Load(nil) error = nil, want error; fixture assumption broken")
	}

	var stdout, stderr bytes.Buffer
	cmd := newCheckCmd(nil, reg, &stdout, &stderr)
	cmd.SetOut(&stdout)
	err := cmd.Execute()

	if err == nil {
		t.Fatal("Execute() error = nil, want error")
	}
	if err.Error() != loadErr.Error() {
		t.Errorf("Execute() error = %q, want %q", err.Error(), loadErr.Error())
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("stderr = %q, want empty; check must not print its own error", stderr.String())
	}
}

// TestCheckCmd_StoreRow proves the acceptance criterion: check against
// an absent store prints "configuration: ok" then "store: absent" and
// exits 0, and against a zero-byte database prints the not-a-store
// line and exits nonzero. Fixture state: a full valid environment, so
// only the store differs.
func TestCheckCmd_StoreRow(t *testing.T) {
	t.Run("absent store exits zero", func(t *testing.T) {
		environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())

		var stdout, stderr bytes.Buffer
		cmd := newCheckCmd(environ, freshRegistry(), &stdout, &stderr)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil", err)
		}
		want := "configuration: ok\nstore: absent\n"
		if got := stdout.String(); got != want {
			t.Errorf("stdout = %q, want %q", got, want)
		}
	})

	t.Run("zero-byte database exits nonzero", func(t *testing.T) {
		stateDir := t.TempDir()
		buildZeroByteStoreLeaf(t, stateDir)
		environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir)

		var stdout, stderr bytes.Buffer
		cmd := newCheckCmd(environ, freshRegistry(), &stdout, &stderr)
		err := cmd.Execute()
		if err == nil {
			t.Fatal("Execute() error = nil, want a refusal: the database is an unfinished create, not a store")
		}

		out := stdout.String()
		if !strings.HasPrefix(out, "configuration: ok\n") {
			t.Errorf("stdout = %q, want it to start with %q", out, "configuration: ok\n")
		}
		if !strings.Contains(out, "store: not a store: ") {
			t.Errorf("stdout = %q, want a store row naming not a store", out)
		}
	})
}

// TestCheckCmd_BusyTimeoutBounds proves the busy timeout's range check runs
// before the inspect mode: an out-of-range value fails naming the key,
// and no store line prints.
func TestCheckCmd_BusyTimeoutBounds(t *testing.T) {
	tests := []struct {
		name   string
		value  string
		wantOK bool
	}{
		{name: "zero", value: "0s"},
		{name: "negative", value: "-1s"},
		{name: "below one millisecond", value: "500us"},
		{name: "far above the ceiling", value: "1000h"},
		{name: "at the ceiling", value: "2147483647ms", wantOK: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir()),
				"DUBPLATE_STORE_BUSY_TIMEOUT="+tt.value)
			reg := freshRegistry()
			registerStoreConfig(reg)

			var stdout, stderr bytes.Buffer
			cmd := newCheckCmd(environ, reg, &stdout, &stderr)
			err := cmd.Execute()

			if tt.wantOK {
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Execute() error = nil, want a refusal: the busy timeout is out of range")
			}
			if !strings.Contains(err.Error(), "DUBPLATE_STORE_BUSY_TIMEOUT") {
				t.Errorf("error = %q, want it to name DUBPLATE_STORE_BUSY_TIMEOUT", err.Error())
			}
			if strings.Contains(stdout.String(), "store:") {
				t.Errorf("stdout = %q, want no store line: the bound check runs before the inspect mode", stdout.String())
			}
		})
	}
}

// TestCheckAndDoctor_AuthSettingsEveryStoreState proves check and
// doctor refuse an auth value serve's seed would refuse in every store
// state they vouch for, and pass an in-bounds one. Fixture state per
// row: the absent rows have no store leaf at all, so check has no
// facts; the 0003 row's facts mark the settings table absent; the 0001
// row's report carries no facts at all, so a rule keyed on the facts
// alone would evaluate nothing; the current row reaches the stored
// rows. Each refusing row names the key and the bound or the accepted
// form, or both keys of a relation.
func TestCheckAndDoctor_AuthSettingsEveryStoreState(t *testing.T) {
	tests := []struct {
		name    string
		version int // 0: no store; -1: current
		env     []string
		// want maps each key whose row the run must print to every
		// part that row must name.
		want     map[string][]string
		wantFail bool
	}{
		{
			name:     "a current store, a rate out of bounds",
			version:  -1,
			env:      []string{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR=0"},
			want:     map[string][]string{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR": {"refused:", "below the floor 1"}},
			wantFail: true,
		},
		{
			name:     "an absent store, the body cap out of bounds",
			env:      []string{"DUBPLATE_AUTH_REQUEST_BODY_CAP=100"},
			want:     map[string][]string{"DUBPLATE_AUTH_REQUEST_BODY_CAP": {"refused:", "below the floor 512"}},
			wantFail: true,
		},
		{
			name:     "a store at 0003, an expiry that does not parse",
			version:  3,
			env:      []string{"DUBPLATE_MAGIC_LINK_EXPIRY=soon"},
			want:     map[string][]string{"DUBPLATE_MAGIC_LINK_EXPIRY": {"refused:", `invalid value "soon"`, "a Go duration string"}},
			wantFail: true,
		},
		{
			name:     "a store at 0001, a rate out of bounds",
			version:  1,
			env:      []string{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR=601"},
			want:     map[string][]string{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR": {"refused:", "above the ceiling 600"}},
			wantFail: true,
		},
		{
			name: "an absent store, the retention below the default expiry",
			env:  []string{"DUBPLATE_NONCE_RETENTION=10m"},
			want: map[string][]string{
				"DUBPLATE_NONCE_RETENTION":   {"refused:", "DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"},
				"DUBPLATE_MAGIC_LINK_EXPIRY": {"refused:", "DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"},
			},
			wantFail: true,
		},
		{
			name: "an absent store, an in-bounds expiry",
			env:  []string{"DUBPLATE_MAGIC_LINK_EXPIRY=20m"},
			want: map[string][]string{"DUBPLATE_MAGIC_LINK_EXPIRY": {`"20m"`, "seeds"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(doctorEnviron(t), tt.env...)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			switch tt.version {
			case 0:
			case -1:
				seedBackupStore(t, stateDir)
				seedDoctorAdmin(t, environ)
			default:
				olderStore(t, stateDir, tt.version)
			}

			t.Run("check", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := newCheckCmd(environ, productionRegistry(), &stdout, &stderr).Execute()
				if (err != nil) != tt.wantFail {
					t.Fatalf("Execute() error = %v, wantFail %v; stdout %q", err, tt.wantFail, stdout.String())
				}
				for key, parts := range tt.want {
					var line string
					for l := range strings.SplitSeq(stdout.String(), "\n") {
						if strings.HasPrefix(l, "settings: "+key+": ") {
							line = l
						}
					}
					if line == "" {
						t.Fatalf("stdout = %q, want a settings line for %s", stdout.String(), key)
					}
					for _, part := range parts {
						if !strings.Contains(line, part) {
							t.Errorf("line = %q, want it to contain %q", line, part)
						}
					}
				}
			})

			t.Run("doctor", func(t *testing.T) {
				var stdout, stderr bytes.Buffer
				err := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr).Execute()
				if tt.wantFail && err == nil {
					t.Fatalf("Execute() error = nil, want a refusal; stdout %q", stdout.String())
				}
				if !tt.wantFail && err != nil {
					t.Fatalf("Execute() error = %v, want nil; stdout %q", err, stdout.String())
				}
				for key, parts := range tt.want {
					status, ok := doctorKeyStatus(stdout.String(), key)
					if !ok {
						t.Fatalf("stdout = %q, want exactly one row keyed %s", stdout.String(), key)
					}
					for _, part := range parts {
						if !strings.Contains(status, part) {
							t.Errorf("%s row = %q, want it to contain %q", key, status, part)
						}
					}
				}
				if !tt.wantFail {
					status, _ := doctorKeyStatus(stdout.String(), "DUBPLATE_MAGIC_LINK_EXPIRY")
					if want := `env "20m" seeds at serve's next start`; !strings.Contains(status, want) {
						t.Errorf("row = %q, want it to contain %q", status, want)
					}
				}
			})
		})
	}
}
