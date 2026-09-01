package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/importer"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ping"
)

// testNudgeImporter builds an [importer.Importer] whose Mailer points at
// a refused local address -- fails fast, logged and swallowed, the same
// best-effort seam internal/importer's own tests use -- with every
// other collaborator wired to root's fixture tree.
func testNudgeImporter(root string) *importer.Importer {
	return &importer.Importer{
		Cfg: &config.Config{
			AdminEmail: "admin@example.com",
			Paths: config.Paths{
				InboxDir:      filepath.Join(root, "inbox"),
				StagingDir:    filepath.Join(root, "staging"),
				ReviewDir:     filepath.Join(root, "review"),
				QuarantineDir: filepath.Join(root, "quarantine"),
				StateDir:      filepath.Join(root, "state"),
			},
		},
		Roster: config.Roster{},
		Mailer: notify.New(config.SMTPConfig{Host: "127.0.0.1", Port: 1}),
		Ping:   ping.NewClient(),
		Now:    time.Now,
	}
}

// stageSessionFixture writes <inboxDir>/<dirName>/<sessionID>__<album>
// plus its session.json sidecar, with a wrong-extension track: the
// quality gate rejects it before beets is ever invoked, the same
// dependency-free "the run succeeds without a real beet binary" fixture
// internal/importer's own tests use.
func stageSessionFixture(t *testing.T, inboxDir, sessionID, dirName, album string) {
	t.Helper()
	dir := filepath.Join(inboxDir, dirName, sessionID+"__"+album)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "track.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}
	sidecar, err := json.Marshal(map[string]any{
		"version":    1,
		"session_id": sessionID,
		"email":      "alice@example.com",
		"dir_name":   dirName,
		"album":      album,
		"collection": "",
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("marshal sidecar: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "session.json"), sidecar, 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
}

func writeTrigger(t *testing.T, triggerDir, sessionID string) {
	t.Helper()
	if err := os.MkdirAll(triggerDir, 0o755); err != nil {
		t.Fatalf("mkdir trigger dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(triggerDir, sessionID), nil, 0o644); err != nil {
		t.Fatalf("write trigger: %v", err)
	}
}

// TestRunNudgeConsumeDeletesTriggerOnlyAfterSuccess proves the
// delete-ordering contract: a trigger whose import succeeds is removed,
// and a trigger whose import fails for a real reason (no matching
// promoted directory) survives for the path unit's re-arm or the timer
// backstop to retry -- and the batch itself still reports the failure
// rather than swallowing it.
func TestRunNudgeConsumeDeletesTriggerOnlyAfterSuccess(t *testing.T) {
	root := t.TempDir()
	im := testNudgeImporter(root)
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	stageSessionFixture(t, im.Cfg.Paths.InboxDir, "sessOK", "alice", "Album A")
	// sessMissing has a trigger file but no matching promoted directory
	// under InboxDir, so its import fails with ErrSessionNotFound.

	triggerDir := filepath.Join(root, "run", "nudge-triggers")
	writeTrigger(t, triggerDir, "sessOK")
	writeTrigger(t, triggerDir, "sessMissing")

	err := runNudgeConsume(context.Background(), im, triggerDir, lock.Path(im.Cfg.Paths.StateDir), 5*time.Second)
	if err == nil {
		t.Fatal("runNudgeConsume error = nil, want the sessMissing failure surfaced")
	}
	if !errors.Is(err, importer.ErrSessionNotFound) {
		t.Errorf("runNudgeConsume error = %v, want it to wrap ErrSessionNotFound", err)
	}

	if _, err := os.Stat(filepath.Join(triggerDir, "sessOK")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("successful session's trigger survived: err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(triggerDir, "sessMissing")); err != nil {
		t.Errorf("failed session's trigger was deleted despite the import failing: err = %v", err)
	}
}

// TestRunNudgeConsumeLockContentionTimesOut proves the bounded wait:
// another holder keeping the import lock past lockTimeout is a real
// failure (nonzero exit, the OnFailure chain pages the admin), and the
// contended batch never even starts -- every trigger is left untouched
// for the path unit's level-triggered re-arm or the 15-minute timer
// backstop.
func TestRunNudgeConsumeLockContentionTimesOut(t *testing.T) {
	root := t.TempDir()
	im := testNudgeImporter(root)
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	held, err := lock.Acquire(lock.Path(im.Cfg.Paths.StateDir))
	if err != nil {
		t.Fatalf("acquire lock for the test's own holder: %v", err)
	}
	defer func() { _ = held.Release() }()

	triggerDir := filepath.Join(root, "run", "nudge-triggers")
	writeTrigger(t, triggerDir, "sessAny")

	const timeout = 200 * time.Millisecond
	start := time.Now()
	err = runNudgeConsume(context.Background(), im, triggerDir, lock.Path(im.Cfg.Paths.StateDir), timeout)
	if err == nil {
		t.Fatal("runNudgeConsume error = nil, want a lock-acquire timeout")
	}
	if elapsed := time.Since(start); elapsed < timeout {
		t.Errorf("runNudgeConsume returned after %v, want it to have waited out the %v bound", elapsed, timeout)
	}

	if _, err := os.Stat(filepath.Join(triggerDir, "sessAny")); err != nil {
		t.Errorf("trigger removed despite the batch never starting: %v", err)
	}
}

func TestRunNudgeConsumeEmptyOrMissingTriggerDirIsNoop(t *testing.T) {
	root := t.TempDir()
	im := testNudgeImporter(root)

	if err := runNudgeConsume(context.Background(), im, filepath.Join(root, "does-not-exist"), lock.Path(im.Cfg.Paths.StateDir), time.Second); err != nil {
		t.Errorf("runNudgeConsume(missing trigger dir) error = %v, want nil", err)
	}

	empty := filepath.Join(root, "run", "nudge-triggers")
	if err := os.MkdirAll(empty, 0o755); err != nil {
		t.Fatalf("mkdir trigger dir: %v", err)
	}
	if err := runNudgeConsume(context.Background(), im, empty, lock.Path(im.Cfg.Paths.StateDir), time.Second); err != nil {
		t.Errorf("runNudgeConsume(empty trigger dir) error = %v, want nil", err)
	}
}
