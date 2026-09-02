package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/importer"
	"github.com/glw907/dubplate/internal/lock"
)

// nudgeTriggerSubdir is the directory under [config.Paths.TriggerDir]
// `dubplate serve` writes nudge trigger files into and this command
// drains. MUST match internal/serve/serve.go's identical unexported
// nudgeTriggerSubdir constant: serve and cmd/dubplate never import each
// other, so this name is the shared, hand-kept contract between the
// writer and the reader, not a value either derives from the other.
// U8's tmpfiles fragment (config/tmpfiles-dubplate.conf) creates it.
const nudgeTriggerSubdir = "nudge-triggers"

// nudgeLockTimeout bounds how long `dubplate nudge-consume` waits for
// the import lock before giving up. dubplate-import-nudge.path is
// level-triggered (DirectoryNotEmpty=): a timeout that leaves trigger
// files in place re-fires the service immediately on the path unit's
// re-arm, and the 15-minute import timer is the backstop regardless, so
// a bounded wait here is safe rather than blocking forever against a
// stuck poll run holding the lock.
const nudgeLockTimeout = 5 * time.Minute

func newNudgeConsumeCmd() *cobra.Command {
	return &cobra.Command{
		Use:    "nudge-consume",
		Short:  "Drain the nudge trigger directory, importing each triggered session under one held lock",
		Args:   cobra.NoArgs,
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			im, err := newImporter()
			if err != nil {
				return err
			}
			triggerDir := filepath.Join(im.Cfg.Paths.TriggerDir, nudgeTriggerSubdir)
			return runNudgeConsume(ctx, im, triggerDir, lock.Path(im.Cfg.Paths.StateDir), nudgeLockTimeout)
		},
	}
}

// runNudgeConsume drains triggerDir: it lists the trigger files present
// at the moment it starts (a trigger written after this snapshot is
// simply picked up by the path unit's own re-arm, not raced against
// here), acquires the import lock ONCE via [lock.AcquireBlocking] for
// the whole batch (review-forced, N3: an exit-and-retry consumer cannot
// observe lock contention, and re-acquiring per session would thrash
// against a concurrent poll run), then imports each triggered session
// in-process under that one held lock via
// [importer.Importer.ImportSessionLocked] -- never through
// [importer.Importer.RunSession], which acquires its own lock and would
// report itself held by this very process. A trigger's file is deleted
// once its import returns nil, or once its import returns
// [importer.ErrSessionNotFound] or [importer.ErrMissingSidecar]: both
// mark the trigger itself as stale (its session was already settled by
// the 15-minute timer backstop, or never came from the uploader) rather
// than an import that might still succeed on retry, so retiring the
// trigger is the only way to stop the level-triggered path unit from
// re-firing on it forever. Every other failure leaves the trigger in
// place for the path unit's level-triggered re-arm or the 15-minute
// timer backstop to retry. A per-session failure never aborts the
// batch -- it is logged and joined into the return so the process still
// exits nonzero overall and the OnFailure chain pages the admin, while
// every other triggered session in the batch still gets its turn.
func runNudgeConsume(ctx context.Context, im *importer.Importer, triggerDir, lockPath string, lockTimeout time.Duration) error {
	entries, err := os.ReadDir(triggerDir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// The path unit fires only on a non-empty directory; a race
			// where it emptied between the fire and this read means
			// nothing is left to do, not a failure.
			return nil
		}
		return fmt.Errorf("read nudge trigger directory %s: %w", triggerDir, err)
	}
	if len(entries) == 0 {
		return nil
	}

	lockCtx, cancel := context.WithTimeout(ctx, lockTimeout)
	defer cancel()
	l, err := lock.AcquireBlocking(lockCtx, lockPath)
	if err != nil {
		return fmt.Errorf("acquire import lock: %w", err)
	}
	defer func() {
		if err := l.Release(); err != nil {
			slog.Error("nudge-consume: release import lock", "err", err)
		}
	}()

	var errs []error
	for _, e := range entries {
		sessionID := e.Name()
		triggerPath := filepath.Join(triggerDir, sessionID)

		err := im.ImportSessionLocked(ctx, sessionID)
		switch {
		case err == nil:
			// fall through to the trigger removal below.
		case errors.Is(err, importer.ErrSessionNotFound), errors.Is(err, importer.ErrMissingSidecar):
			slog.Warn("nudge-consume: stale nudge trigger, retiring", "session_id", sessionID, "err", err)
		default:
			slog.Error("nudge-consume: import session", "session_id", sessionID, "err", err)
			errs = append(errs, fmt.Errorf("import session %s: %w", sessionID, err))
			continue
		}

		if err := os.Remove(triggerPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Error("nudge-consume: remove trigger", "path", triggerPath, "err", err)
			errs = append(errs, fmt.Errorf("remove trigger %s: %w", triggerPath, err))
		}
	}
	return errors.Join(errs...)
}
