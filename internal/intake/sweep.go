package intake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// stallMailWait bounds the stall mail: the address read and every send.
const stallMailWait = 30 * time.Second

// sweepEvery runs the sweep each SweepInterval until ctx ends.
func (in *Intake) sweepEvery(ctx context.Context) {
	ticker := time.NewTicker(in.cfg.SweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			in.sweep(ctx)
		}
	}
}

// sweep is one pass, in order: receive the uploads whose last byte
// landed with no receipt, expire the uploads idle past the resume
// window and the checks held past the hold, clean up, so an expiry's
// bytes go in the sweep that expired it, wake the workers, so a check a
// store fault left in checking runs again, and record a stall for each
// check past STALL_AFTER, mailing the live admins once the record
// commits. Each step's failure is logged and the next step runs.
func (in *Intake) sweep(ctx context.Context) {
	log := in.cfg.Log
	w, err := in.cfg.Windows(ctx)
	if err != nil {
		log.Error("intake: sweep could not read its windows", "error", err)
		return
	}
	if _, err := in.receiveComplete(ctx); err != nil {
		log.Error("intake: sweep receipt", "error", err)
	}
	if err := in.expireIdle(ctx, w.ResumeWindow); err != nil {
		log.Error("intake: sweep idle expiry", "error", err)
	}
	if err := in.expireHeld(ctx, w.Hold); err != nil {
		log.Error("intake: sweep hold expiry", "error", err)
	}
	if _, err := in.clean(ctx, w.ResumeWindow); err != nil {
		log.Error("intake: sweep cleanup", "error", err)
	}
	in.subs.poke()
	stalled, err := in.markStalled(ctx, w.StallAfter)
	if err != nil {
		log.Error("intake: sweep stall", "error", err)
	}
	if len(stalled) > 0 {
		in.mailStalled(ctx, stalled, w.StallAfter)
	}
}

// expireIdle expires each upload idle past window. Its idle clock is
// the later of its updated_at and its newest tus data file's
// modification time, since a long PATCH writes no row.
func (in *Intake) expireIdle(ctx context.Context, window time.Duration) error {
	cutoff := in.cfg.Now().Add(-window)
	type idleFile struct {
		submission, file []byte
	}
	files, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) (idleFile, error) {
		var f idleFile
		return f, rows.Scan(&f.submission, &f.file)
	}, `SELECT s.id, f.id FROM submissions s JOIN submission_files f ON f.submission_id = s.id
		WHERE s.state = ? AND s.updated_at < ? ORDER BY s.id`, stateUploading, cutoff.UnixMicro())
	if err != nil {
		return fmt.Errorf("read idle uploads: %w", err)
	}

	var order [][]byte
	written := map[string]bool{}
	for _, f := range files {
		key := string(f.submission)
		if _, seen := written[key]; !seen {
			order = append(order, f.submission)
			written[key] = false
		}
		if info, err := in.root.Stat(uploadName(f.file)); err == nil && !info.ModTime().Before(cutoff) {
			written[key] = true
		}
	}
	var errs []error
	for _, id := range order {
		if written[string(id)] {
			continue
		}
		err := in.subs.expire(ctx, id, "idle past the resume window", stateUploading)
		if err != nil && !errors.Is(err, errWrongState) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// expireHeld expires each check that could not run and has waited past
// hold for a retry.
func (in *Intake) expireHeld(ctx context.Context, hold time.Duration) error {
	ids, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) ([]byte, error) {
		var id []byte
		return id, rows.Scan(&id)
	}, `SELECT id FROM submissions WHERE state = ? AND updated_at < ?`, stateCheckUnfinished, in.cfg.Now().Add(-hold).UnixMicro())
	if err != nil {
		return fmt.Errorf("read held checks: %w", err)
	}
	var errs []error
	for _, id := range ids {
		err := in.subs.expire(ctx, id, "held past the hold", stateCheckUnfinished)
		if err != nil && !errors.Is(err, errWrongState) {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// markStalled records a stalled event for each submission in checking
// since before after ago, once per stay: the submission's latest event
// is the checking event that began it until the stall is recorded, so
// the record survives a restart and is never repeated. The events land
// in one Update, and markStalled returns the ids it marked.
func (in *Intake) markStalled(ctx context.Context, after time.Duration) ([][]byte, error) {
	cutoff := in.cfg.Now().Add(-after).UnixMicro()
	var marked [][]byte
	err := in.subs.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		rows, err := tx.Query(ctx, `SELECT e.submission_id FROM submission_events e JOIN submissions s ON s.id = e.submission_id
			WHERE s.state = ?1 AND e.kind = ?1 AND e.at < ?2
			AND e.seq = (SELECT max(seq) FROM submission_events WHERE submission_id = e.submission_id)
			ORDER BY e.at, e.submission_id`, stateChecking, cutoff)
		if err != nil {
			return fmt.Errorf("read long checks: %w", err)
		}
		marked = nil
		for rows.Next() {
			var id []byte
			if err := rows.Scan(&id); err != nil {
				_ = rows.Close()
				return fmt.Errorf("read long check: %w", err)
			}
			marked = append(marked, id)
		}
		if err := rows.Close(); err != nil {
			return fmt.Errorf("read long checks: %w", err)
		}
		for _, id := range marked {
			if err := appendEvent(ctx, tx, id, at, eventStalled, ""); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return marked, nil
}

// mailStalled tells every live admin that the checks stalled have run
// past after. It runs after the stall events commit and outside any
// write, so no writer waits on the mail server, on a context of its own
// that a shutdown does not cut short, since the event that records the
// stall is never repeated. A failure is logged; the event stands.
func (in *Intake) mailStalled(ctx context.Context, stalled [][]byte, after time.Duration) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), stallMailWait)
	defer cancel()
	log := in.cfg.Log.With("stalled", len(stalled))

	admins, err := in.cfg.Admins(ctx)
	if err != nil {
		log.Error("intake: stall mail not sent", "reason", err.Error())
		return
	}
	if len(admins) == 0 {
		log.Warn("intake: stall mail not sent: no live admin")
		return
	}
	ids := make([]string, len(stalled))
	for i, id := range stalled {
		ids[i] = wireID(id)
	}
	subject := "Uploads stalled while being checked"
	body := fmt.Sprintf("%d uploads have been checking for longer than %s. The server is still checking them; "+
		"if they stay stalled, its log names the cause.\n\nSubmissions:\n%s\n", len(stalled), after, strings.Join(ids, "\n"))
	for _, to := range admins {
		if err := in.cfg.Send(ctx, to, subject, body); err != nil {
			log.Error("intake: stall mail failed", "to", to, "reason", err.Error())
		}
	}
}
