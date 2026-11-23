package backup_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/backup"
)

// probeOpts is the bound the command fills the row with.
var probeOpts = backup.RemoteRowOptions{FeaturesTimeout: 10 * time.Second}

// cryptWarning is the text a remote that is not an rclone crypt remote
// carries, written out so the tests do not read it from the code they
// check.
const cryptWarning = "the backup remote is not wrapped in an rclone crypt remote; a snapshot carries member data unencrypted at the provider"

// TestBackupRemoteRowStatus is the golden test for the row's name,
// colon, and credential checks: every case here is pure string and
// environment logic, run against a constructed environ with no rclone
// on PATH at all, since none of these states reaches the features
// probe.
func TestBackupRemoteRowStatus(t *testing.T) {
	on := func(kv ...string) []string {
		return append([]string{"DUBPLATE_BACKUP_ENABLED=true"}, kv...)
	}

	tests := []struct {
		name       string
		environ    []string
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "flag off: unavailable, not a failure",
			environ:    nil,
			wantStatus: "unavailable: backup integration disabled",
			wantFailed: false,
		},
		{
			name:       "no colon",
			environ:    on("DUBPLATE_BACKUP_REMOTE=relative/path"),
			wantStatus: "invalid remote: want <name>:<path>",
			wantFailed: true,
		},
		{
			name:       "empty name",
			environ:    on("DUBPLATE_BACKUP_REMOTE=:dir"),
			wantStatus: "invalid remote: want <name>:<path>",
			wantFailed: true,
		},
		{
			name:       "a character outside rclone's remote-name alphabet",
			environ:    on("DUBPLATE_BACKUP_REMOTE=my-remote:dir"),
			wantStatus: `invalid remote: name "my-remote" holds "-", and rclone remote names "can only contain letters, digits, or the _ (underscore) character"`,
			wantFailed: true,
		},
		{
			name:       "empty path is accepted, and the check proceeds to _TYPE",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_TYPE is not set",
			wantFailed: true,
		},
		{
			name:       "_TYPE unset",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:dir"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_TYPE is not set",
			wantFailed: true,
		},
		{
			name:       "crypt _REMOTE unset",
			environ:    on("DUBPLATE_BACKUP_REMOTE=name:dir", "RCLONE_CONFIG_NAME_TYPE=crypt"),
			wantStatus: "invalid remote: RCLONE_CONFIG_NAME_REMOTE is not set",
			wantFailed: true,
		},
		{
			name: "crypt wrapped remote's _TYPE unset",
			environ: on("DUBPLATE_BACKUP_REMOTE=name:dir",
				"RCLONE_CONFIG_NAME_TYPE=crypt",
				"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub"),
			wantStatus: "invalid remote: wrapped remote wrapped: RCLONE_CONFIG_WRAPPED_TYPE is not set",
			wantFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("PATH", t.TempDir()) // no rclone reachable: these states never look for it
			status, failed := backup.RemoteRowStatus(context.Background(), tt.environ, probeOpts)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// TestBackupRemoteRowStatus_RcloneNotFound proves the "rclone not
// found" row: a valid, fully-credentialed remote whose PATH holds no
// rclone binary at all.
func TestBackupRemoteRowStatus_RcloneNotFound(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	environ := []string{
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=name:dir",
		"RCLONE_CONFIG_NAME_TYPE=local",
	}
	status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
	if status != "rclone not found" {
		t.Errorf("status = %q, want %q", status, "rclone not found")
	}
	if !failed {
		t.Error("failed = false, want true")
	}
}

// TestBackupRemoteRowStatus_NoSubprocessWhenFlagOff proves the flag-off
// state runs no subprocess: the stand-in alone on PATH records no
// invocation.
func TestBackupRemoteRowStatus_NoSubprocessWhenFlagOff(t *testing.T) {
	s := useStandin(t, standinConfig{})
	status, failed := backup.RemoteRowStatus(context.Background(), nil, probeOpts)
	if status != "unavailable: backup integration disabled" {
		t.Errorf("status = %q", status)
	}
	if failed {
		t.Error("failed = true, want false")
	}
	if lines := s.lines(t); len(lines) != 0 {
		t.Errorf("stand-in recorded %d line(s), want 0: %v", len(lines), lines)
	}
}

// TestBackupRemoteRowStatus_Features is the golden test for the
// features probe's own states, each run through the rclone stand-in
// rather than a real backend.
func TestBackupRemoteRowStatus_Features(t *testing.T) {
	t.Run("crypt _REMOTE a plain path skips the wrapped check; the probe decides", func(t *testing.T) {
		useStandin(t, standinConfig{backendStdout: `{"Features":{"Copy":false,"Move":true}}`})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=/plain/path",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if status != "ok" {
			t.Errorf("status = %q, want %q (no wrapped remote's _TYPE is set, so the wrapped check must have been skipped)", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("crypt _REMOTE a connection string skips the wrapped check", func(t *testing.T) {
		useStandin(t, standinConfig{backendStdout: `{"Features":{"Copy":false,"Move":true}}`})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=:local,type=vfat:/mnt",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if status != "ok" {
			t.Errorf("status = %q, want %q", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("Copy and Move both false", func(t *testing.T) {
		useStandin(t, standinConfig{backendStdout: `{"Features":{"Copy":false,"Move":false}}`})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=http",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if !strings.Contains(status, "lacks server-side move and copy") {
			t.Errorf("status = %q, want it to name the backend and rclone's --backup-dir requirement", status)
		}
		if !strings.Contains(status, "name") {
			t.Errorf("status = %q, want it to name the backend", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a nonzero exit before the deadline reports the last stderr line", func(t *testing.T) {
		useStandin(t, standinConfig{
			exit:          "backend=1",
			backendStderr: "2026/09/24 12:00:00 NOTICE: some rclone context\n401 bad_auth_token\n",
		})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if status != "401 bad_auth_token" {
			t.Errorf("status = %q, want the last non-blank stderr line %q", status, "401 bad_auth_token")
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a deadline hit reads not checked, whatever the exit", func(t *testing.T) {
		useStandin(t, standinConfig{sleep: "backend", sigint: "honor"})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		short := backup.RemoteRowOptions{FeaturesTimeout: 30 * time.Millisecond}
		begun := time.Now()
		status, failed := backup.RemoteRowStatus(context.Background(), environ, short)
		// The stand-in sleeps for a minute, so only the option's bound
		// can end the probe this soon.
		if took := time.Since(begun); took > 10*time.Second {
			t.Errorf("the probe took %v, want it ended by its %v bound", took, short.FeaturesTimeout)
		}
		if status != "not checked: rclone backend features timed out" {
			t.Errorf("status = %q, want %q", status, "not checked: rclone backend features timed out")
		}
		if failed {
			t.Error("failed = true, want false: a deadline hit never fails doctor")
		}
	})

	t.Run("a non-crypt type carries the encryption warning, never a failure", func(t *testing.T) {
		useStandin(t, standinConfig{backendStdout: `{"Features":{"Copy":true,"Move":false}}`})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=s3",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if want := "ok; " + cryptWarning; status != want {
			t.Errorf("status = %q, want %q", status, want)
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})

	t.Run("a crypt type carries no warning", func(t *testing.T) {
		useStandin(t, standinConfig{backendStdout: `{"Features":{"Copy":false,"Move":true}}`})
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:dir",
			"RCLONE_CONFIG_NAME_TYPE=crypt",
			"RCLONE_CONFIG_NAME_REMOTE=wrapped:sub",
			"RCLONE_CONFIG_WRAPPED_TYPE=local",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if status != "ok" {
			t.Errorf("status = %q, want %q", status, "ok")
		}
		if failed {
			t.Error("failed = true, want false")
		}
	})
}

// TestBackupRemoteRowStatus_EndToEnd runs RemoteRowStatus against the
// real rclone binary rather than the stand-in: a valid local remote, an
// http remote at a closed loopback port (both features false), and a
// crypt remote wrapping an undefined remote.
func TestBackupRemoteRowStatus_EndToEnd(t *testing.T) {
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}

	t.Run("a valid local remote", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2ELOCAL_TYPE", "local")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2elocal:" + t.TempDir(),
			"RCLONE_CONFIG_E2ELOCAL_TYPE=local",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		// local is not crypt, so the encryption warning is appended.
		if !strings.HasPrefix(status, "ok; ") {
			t.Errorf("status = %q, want it to start with %q", status, "ok; ")
		}
		if failed {
			t.Errorf("failed = true, want false: %s", status)
		}
	})

	t.Run("an http remote at a closed loopback port reports both features false", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2EHTTP_TYPE", "http")
		t.Setenv("RCLONE_CONFIG_E2EHTTP_URL", "http://127.0.0.1:9/")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2ehttp:dir",
			"RCLONE_CONFIG_E2EHTTP_TYPE=http",
			"RCLONE_CONFIG_E2EHTTP_URL=http://127.0.0.1:9/",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if !strings.Contains(status, "lacks server-side move and copy") {
			t.Errorf("status = %q, want it to report both features false", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})

	t.Run("a crypt remote whose wrapped remote is undefined", func(t *testing.T) {
		t.Setenv("RCLONE_CONFIG_E2ECRYPT_TYPE", "crypt")
		t.Setenv("RCLONE_CONFIG_E2ECRYPT_REMOTE", "e2eundefined:sub")
		environ := []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=e2ecrypt:dir",
			"RCLONE_CONFIG_E2ECRYPT_TYPE=crypt",
			"RCLONE_CONFIG_E2ECRYPT_REMOTE=e2eundefined:sub",
		}
		status, failed := backup.RemoteRowStatus(context.Background(), environ, probeOpts)
		if status != "invalid remote: wrapped remote e2eundefined: RCLONE_CONFIG_E2EUNDEFINED_TYPE is not set" {
			t.Errorf("status = %q", status)
		}
		if !failed {
			t.Error("failed = false, want true")
		}
	})
}
