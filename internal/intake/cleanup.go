package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/tus/tusd/v2/pkg/handler"
)

// cleanupLockWait bounds cleanup's wait for an upload's tus lock. The
// lock asks its holder to stop, and only a closed submission's upload
// is asked, where stopping is intended; a holder that has not let go by
// then keeps its file for the next sweep.
const cleanupLockWait = 2 * time.Second

// clean removes the bytes outside every submission's staged set, and
// nothing else. The list is closed:
//
//   - all of a terminal submission's bytes;
//   - a submissions/ directory no row names;
//   - the archive of a ZIP past its verdict, in matching or accepted,
//     never in check-unfinished, whose retry extracts from it;
//   - the failing files of an accepted separate-tracks submission;
//   - a tus upload no row names, once older than window.
//
// An open submission's uploads, received or not, and a checking
// archive's x/ are never touched. clean walks what staging holds rather
// than the rows, so its cost follows the bytes on disk and not the
// history, and returns how many uploads and directories it removed. A
// failure leaves its item for the next sweep and the walk goes on.
func (in *Intake) clean(ctx context.Context, window time.Duration) (int, error) {
	removed := 0
	var errs []error
	uploads, err := in.stagedUploads()
	if err != nil {
		return 0, err
	}
	for name, modified := range uploads {
		id, ok := parseID(name)
		if !ok {
			continue
		}
		goes, known, err := in.uploadGoes(ctx, id)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if !known {
			goes = in.cfg.Now().Sub(modified) > window
		}
		if !goes {
			continue
		}
		gone, err := in.removeUpload(ctx, name)
		if err != nil {
			errs = append(errs, err)
		}
		if gone {
			removed++
		}
	}

	dirs, err := fs.ReadDir(in.root.FS(), "submissions")
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return removed, errors.Join(append(errs, fmt.Errorf("list extractions: %w", err))...)
	}
	for _, d := range dirs {
		id, ok := parseID(d.Name())
		if !ok {
			continue
		}
		n, err := in.cleanExtraction(ctx, id)
		if err != nil {
			errs = append(errs, err)
		}
		removed += n
	}
	return removed, errors.Join(errs...)
}

// stagedUploads maps the name of each upload in staging's tus directory
// to the newest modification time of its two files.
func (in *Intake) stagedUploads() (map[string]time.Time, error) {
	entries, err := fs.ReadDir(in.root.FS(), "tus")
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("list uploads: %w", err)
	}
	uploads := make(map[string]time.Time, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		name := strings.TrimSuffix(e.Name(), ".info")
		if info.ModTime().After(uploads[name]) {
			uploads[name] = info.ModTime()
		}
	}
	return uploads, nil
}

// uploadGoes reports whether the closed list takes upload id's tus
// files, and whether any row names it.
func (in *Intake) uploadGoes(ctx context.Context, id []byte) (goes, known bool, err error) {
	var st state
	var unit gate.Unit
	var role string
	var failed bool
	err = in.cfg.Store.QueryRow(ctx, `SELECT s.state, s.verdict_unit, f.role, f.verdict IS NOT NULL
		FROM submission_files f JOIN submissions s ON s.id = f.submission_id WHERE f.id = ?`, id).Scan(&st, &unit, &role, &failed)
	if errors.Is(err, sql.ErrNoRows) {
		return false, false, nil
	}
	if err != nil {
		return false, false, fmt.Errorf("read upload %x: %w", id, err)
	}
	switch {
	case st.terminal():
		return true, true, nil
	case role == "archive":
		return st == stateMatching || st == stateAccepted, true, nil
	}
	return st == stateAccepted && unit == gate.UnitTrack && failed, true, nil
}

// removeUpload removes upload name's data file and .info under its tus
// lock, and reports whether it did. A lock its holder keeps past
// cleanupLockWait leaves both files.
func (in *Intake) removeUpload(ctx context.Context, name string) (bool, error) {
	lock, err := in.locker.NewLock(name)
	if err != nil {
		return false, fmt.Errorf("lock upload %s: %w", name, err)
	}
	wait, cancel := context.WithTimeout(ctx, cleanupLockWait)
	defer cancel()
	switch err := lock.Lock(wait, func() {}); {
	case errors.Is(err, handler.ErrLockTimeout):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("lock upload %s: %w", name, err)
	}
	defer func() { _ = lock.Unlock() }()
	for _, file := range []string{path.Join("tus", name), path.Join("tus", name+".info")} {
		if err := in.root.Remove(file); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return false, fmt.Errorf("remove upload: %w", err)
		}
	}
	return true, nil
}

// cleanExtraction removes what the closed list takes of submission id's
// directory under submissions/: all of it when no row names the
// submission or the submission is terminal, and the failing members of
// an accepted separate-tracks archive. It returns how many directories
// and members it removed.
func (in *Intake) cleanExtraction(ctx context.Context, id []byte) (int, error) {
	var st state
	var unit gate.Unit
	err := in.cfg.Store.QueryRow(ctx, `SELECT state, verdict_unit FROM submissions WHERE id = ?`, id).Scan(&st, &unit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("read submission %x: %w", id, err)
	}
	if err != nil || st.terminal() {
		if err := in.root.RemoveAll(path.Join("submissions", hex.EncodeToString(id))); err != nil {
			return 0, fmt.Errorf("remove extraction: %w", err)
		}
		return 1, nil
	}
	if st != stateAccepted || unit != gate.UnitTrack {
		return 0, nil
	}

	failing, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) ([]byte, error) {
		var member []byte
		return member, rows.Scan(&member)
	}, `SELECT id FROM submission_files WHERE submission_id = ? AND role != 'archive' AND verdict IS NOT NULL`, id)
	if err != nil {
		return 0, fmt.Errorf("read failing members: %w", err)
	}
	removed := 0
	for _, member := range failing {
		err := in.root.Remove(path.Join(extractDir(id), hex.EncodeToString(member)))
		switch {
		case err == nil:
			removed++
		case !errors.Is(err, fs.ErrNotExist):
			return removed, fmt.Errorf("remove failing member: %w", err)
		}
	}
	return removed, nil
}

// cleanSubmission removes what the closed list takes of submission id's
// bytes, at once after its verdict rather than at the next sweep.
func (in *Intake) cleanSubmission(ctx context.Context, id []byte) error {
	uploads, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) ([]byte, error) {
		var file []byte
		return file, rows.Scan(&file)
	}, `SELECT f.id FROM submission_files f JOIN submissions s ON s.id = f.submission_id
		WHERE s.id = ? AND (s.container = 'files' OR f.role = 'archive')`, id)
	if err != nil {
		return fmt.Errorf("read uploads: %w", err)
	}
	for _, file := range uploads {
		goes, _, err := in.uploadGoes(ctx, file)
		if err != nil {
			return err
		}
		if !goes {
			continue
		}
		if _, err := in.removeUpload(ctx, hex.EncodeToString(file)); err != nil {
			return err
		}
	}
	_, err = in.cleanExtraction(ctx, id)
	return err
}
