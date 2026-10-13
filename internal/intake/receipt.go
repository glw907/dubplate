package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// receiptWait bounds receipt's syncs and its write, which outlive the
// request that finished the upload.
const receiptWait = 30 * time.Second

// receive records that file fileID has arrived whole, and moves its
// submission to checking once every file has, poking the worker. It is
// idempotent: a replay for a received file changes nothing and succeeds.
// A file whose submission closed while it uploaded changes nothing and
// returns errWrongState.
//
// tusd writes with no sync and the store commits on another device, so
// the upload's data, its .info, and the tus directory are synced before
// the write that says they exist. The client that sent the last byte
// may drop at once, so receive runs on a context its cancellation does
// not reach.
func (s *submissions) receive(ctx context.Context, fileID []byte) error {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), receiptWait)
	defer cancel()

	tus := filepath.Join(s.staging, "tus")
	data := filepath.Join(tus, hex.EncodeToString(fileID))
	for _, path := range []string{data, data + ".info", tus} {
		if err := s.sync(path); err != nil {
			return fmt.Errorf("sync received upload: %w", err)
		}
	}

	moved := false
	err := s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		var submission []byte
		var received sql.NullInt64
		var current state
		err := tx.QueryRow(ctx, `SELECT f.submission_id, f.received_at, s.state FROM submission_files f
			JOIN submissions s ON s.id = f.submission_id WHERE f.id = ?`, fileID).Scan(&submission, &received, &current)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("%w: no file %x", errNotFound, fileID)
		}
		if err != nil {
			return fmt.Errorf("read submission file: %w", err)
		}
		if received.Valid {
			return nil
		}
		if current != stateUploading {
			return fmt.Errorf("%w: file %x finished in a %s submission", errWrongState, fileID, current)
		}

		if _, err := tx.Exec(ctx, `UPDATE submission_files SET received_at = ? WHERE id = ?`, at, fileID); err != nil {
			return fmt.Errorf("mark file received: %w", err)
		}
		if _, err := tx.Exec(ctx, `UPDATE submissions SET updated_at = ? WHERE id = ?`, at, submission); err != nil {
			return fmt.Errorf("touch submission: %w", err)
		}
		var waiting int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM submission_files WHERE submission_id = ? AND received_at IS NULL`,
			submission).Scan(&waiting); err != nil {
			return fmt.Errorf("count unreceived files: %w", err)
		}
		if waiting > 0 {
			return nil
		}
		moved = true
		return move(ctx, tx, submission, at, stateChecking, "", stateUploading)
	})
	if err != nil {
		return err
	}
	if moved {
		s.poke()
	}
	return nil
}

// syncPath flushes a file or a directory to disk.
func syncPath(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
