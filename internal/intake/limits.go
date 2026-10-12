package intake

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// limits are the settings one create or extraction runs under, read
// before its Update: a setting read inside one would be read under the
// write lock. Byte counts are in bytes.
type limits struct {
	policy            gate.Policy
	sessionsPerMember int
	// groupQuotaBytes is the group's cap; zero sets none.
	groupQuotaBytes int64
	freeSpaceFloor  int64
}

// limitRefusal is a create or an extraction refused by its first failing
// limit. For could-not-check, err is the failed free-space read.
type limitRefusal struct {
	code reject.Code
	err  error
}

func (e *limitRefusal) Error() string {
	if e.err != nil {
		return string(e.code) + ": " + e.err.Error()
	}
	return string(e.code)
}

func (e *limitRefusal) Unwrap() error { return e.err }

// freeRead is one read of staging's free space. A failed read is
// carried rather than returned, because it answers only once every limit
// before the floor has passed.
type freeRead struct {
	bytes uint64
	err   error
}

func (s *submissions) readFree() freeRead {
	b, err := s.freeBytes(s.staging)
	return freeRead{bytes: b, err: err}
}

// checkCreate runs the create limits in order over tx, the transaction
// that writes the reservation, so no other create lands between their
// reads and that write: uploads switched off, the session count, then
// the space limits.
func checkCreate(ctx context.Context, tx *store.Tx, uploader identity.UserID, total int64, lim limits, free freeRead) error {
	var mayUpload, sessionCap sql.NullInt64
	err := tx.QueryRow(ctx, `SELECT may_upload, upload_concurrency_cap FROM member_overrides WHERE user_id = ?`,
		uploader[:]).Scan(&mayUpload, &sessionCap)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read member overrides: %w", err)
	}
	if mayUpload.Valid && mayUpload.Int64 == 0 {
		return &limitRefusal{code: reject.UploadsOff}
	}

	sessions := int64(lim.sessionsPerMember)
	if sessionCap.Valid {
		sessions = sessionCap.Int64
	}
	var uploading int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM submissions WHERE uploader = ? AND state = ?`,
		uploader[:], stateUploading).Scan(&uploading); err != nil {
		return fmt.Errorf("count upload sessions: %w", err)
	}
	if uploading >= sessions {
		return &limitRefusal{code: reject.TooManySessions}
	}
	return checkSpace(ctx, tx, uploader, nil, total, lim, free)
}

// checkSpace runs the space limits over tx for total more bytes reserved
// by uploader: the member's quota, the group's, and the floor. except is
// a submission whose own reservation total replaces, left out of every
// sum; nil leaves none out.
func checkSpace(ctx context.Context, tx *store.Tx, uploader identity.UserID, except []byte, total int64, lim limits, free freeRead) error {
	var quota sql.NullInt64
	err := tx.QueryRow(ctx, `SELECT quota_bytes FROM member_overrides WHERE user_id = ?`, uploader[:]).Scan(&quota)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("read member quota: %w", err)
	}
	if quota.Valid {
		use, err := memberUse(ctx, tx, uploader, except)
		if err != nil {
			return err
		}
		if use+total > quota.Int64 {
			return &limitRefusal{code: reject.OverQuota}
		}
	}

	if lim.groupQuotaBytes > 0 {
		var group int64
		err := tx.QueryRow(ctx, `SELECT physical_total_bytes + coalesce((SELECT sum(reserved_bytes) FROM submissions
			WHERE `+openSubmission+` AND id IS NOT ?), 0) FROM group_quota`, except).Scan(&group)
		if err != nil {
			return fmt.Errorf("read group use: %w", err)
		}
		if group+total > lim.groupQuotaBytes {
			return &limitRefusal{code: reject.OverQuota}
		}
	}

	if free.err != nil {
		return &limitRefusal{code: reject.CouldNotCheck, err: free.err}
	}
	// An upload's unwritten bytes are its reservation less its received
	// files, and a checking ZIP's are its reservation, the extracted
	// total once extraction reserves it. Partly written files count in
	// full, which errs toward refusing.
	var unwritten int64
	err = tx.QueryRow(ctx, `SELECT coalesce(sum(CASE
			WHEN s.state = 'uploading' THEN s.reserved_bytes - coalesce((SELECT sum(f.size) FROM submission_files f
				WHERE f.submission_id = s.id AND f.received_at IS NOT NULL), 0)
			WHEN s.state = 'checking' AND s.container = 'zip' THEN s.reserved_bytes
			ELSE 0 END), 0)
		FROM submissions s WHERE s.id IS NOT ?`, except).Scan(&unwritten)
	if err != nil {
		return fmt.Errorf("read unwritten bytes: %w", err)
	}
	available := int64(min(free.bytes, math.MaxInt64))
	if available-unwritten-total < lim.freeSpaceFloor {
		return &limitRefusal{code: reject.NoSpace}
	}
	return nil
}

// memberUse is what counts against uploader's quota: their charged bytes
// plus every open submission's reservation, less except's.
func memberUse(ctx context.Context, tx *store.Tx, uploader identity.UserID, except []byte) (int64, error) {
	var use int64
	err := tx.QueryRow(ctx, `SELECT coalesce((SELECT charged_bytes FROM member_quota WHERE user_id = ?1), 0)
		+ coalesce((SELECT sum(reserved_bytes) FROM submissions
			WHERE uploader = ?1 AND `+openSubmission+` AND id IS NOT ?2), 0)`, uploader[:], except).Scan(&use)
	if err != nil {
		return 0, fmt.Errorf("read member use: %w", err)
	}
	return use, nil
}

// reserveExtracted replaces a checking ZIP's reservation, its archive's
// size, with total, its members' extracted size, re-running the space
// limits with every other submission's bytes counted. A failing limit
// rejects the submission in the same Update and returns its code; a
// failed free-space read is gate.ErrCouldNotRun, which leaves the
// submission for check-unfinished with its archive kept.
func (s *submissions) reserveExtracted(ctx context.Context, id []byte, total int64, lim limits) (reject.Code, error) {
	free := s.readFree()
	var refused reject.Code
	err := s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		var uploader []byte
		var current state
		err := tx.QueryRow(ctx, `SELECT uploader, state FROM submissions WHERE id = ?`, id).Scan(&uploader, &current)
		if errors.Is(err, sql.ErrNoRows) {
			return errNotFound
		}
		if err != nil {
			return fmt.Errorf("read submission: %w", err)
		}
		if current != stateChecking {
			return fmt.Errorf("%w: reserving for a %s submission", errWrongState, current)
		}

		err = checkSpace(ctx, tx, identity.UserID(uploader), id, total, lim, free)
		var lr *limitRefusal
		switch {
		case err == nil:
			if _, err := tx.Exec(ctx, `UPDATE submissions SET reserved_bytes = ?, updated_at = ? WHERE id = ?`, total, at, id); err != nil {
				return fmt.Errorf("write extracted reservation: %w", err)
			}
			return nil
		case errors.As(err, &lr) && lr.code == reject.CouldNotCheck:
			return fmt.Errorf("%w: %w", gate.ErrCouldNotRun, lr.err)
		case errors.As(err, &lr):
			refused = lr.code
			return move(ctx, tx, id, at, stateRejected, string(lr.code), stateChecking)
		}
		return err
	})
	if err != nil {
		return "", err
	}
	return refused, nil
}

// effectiveAllowlist narrows bound, the formats the server admits, to
// those collection coll admits when its admission_format_allowlist
// names any, separated by commas: a collection may narrow the bound and
// never widen it.
func effectiveAllowlist(ctx context.Context, tx *store.Tx, coll []byte, bound []string) ([]string, error) {
	var own sql.NullString
	if err := tx.QueryRow(ctx, `SELECT admission_format_allowlist FROM collections WHERE id = ?`, coll).Scan(&own); err != nil {
		return nil, fmt.Errorf("read collection allowlist: %w", err)
	}
	if !own.Valid {
		return bound, nil
	}
	admits := strings.Split(own.String, ",")
	for i := range admits {
		admits[i] = strings.TrimSpace(admits[i])
	}
	return slices.DeleteFunc(slices.Clone(bound), func(name string) bool {
		return !slices.Contains(admits, name)
	}), nil
}
