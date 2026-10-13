package intake

import (
	"context"
	"fmt"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// A verdict is the gate's judgment on a checking submission, as the
// worker commits it.
//
// Only a final verdict, to matching, accepted, or rejected, carries
// members: commitVerdict inserts their rows whatever the target, and a
// check-unfinished submission's retry extracts again and mints new
// rows, which a second insert under the same paths would refuse. A move
// to accepted needs at least one file the gate passed.
type verdict struct {
	// to is matching, accepted, rejected, or check-unfinished. A move to
	// accepted runs accept, which may turn it to rejected.
	to state
	// reason is the event's reason: an album-level code, or why a check
	// could not run.
	reason   string
	files    []fileVerdict
	members  []memberVerdict
	warnings []gate.FileReason
	// albumKey is set for an album that passed the gate.
	albumKey []byte
}

// fileVerdict is the gate's result for one of the submission's rows.
type fileVerdict struct {
	id     []byte
	result gate.FileResult
}

// memberVerdict is the gate's result for one extracted ZIP member, which
// gets its row only now: id is the server-minted name it extracted to.
type memberVerdict struct {
	id     []byte
	path   []byte
	size   int64
	result gate.FileResult
}

// commitVerdict writes v for checking submission id in one Update: each
// file's result, the members' rows, the warnings, the album key, the
// move, and, when the submission goes on, the reservation trued down to
// the files it keeps. A submission that left checking meanwhile fails
// the move, and none of it lands.
func (s *submissions) commitVerdict(ctx context.Context, id []byte, v verdict) error {
	return s.update(ctx, func(ctx context.Context, tx *store.Tx, at int64) error {
		for _, f := range v.files {
			if err := writeResult(ctx, tx, id, f.id, f.result); err != nil {
				return err
			}
		}
		for _, m := range v.members {
			role := "extra"
			if gate.IsAudio(string(m.path)) {
				role = "audio"
			}
			if _, err := tx.Exec(ctx, `INSERT INTO submission_files (id, submission_id, rel_path, size, role)
				VALUES (?, ?, ?, ?, ?)`, m.id, id, m.path, m.size, role); err != nil {
				return fmt.Errorf("write extracted member: %w", err)
			}
			if err := writeResult(ctx, tx, id, m.id, m.result); err != nil {
				return err
			}
		}
		for _, w := range v.warnings {
			if _, err := tx.Exec(ctx, `INSERT INTO submission_warnings (id, submission_id, code, subject, at)
				VALUES (?, ?, ?, ?, ?)`, store.NewID(), id, string(w.Reason.Code), blob(w.Path), at); err != nil {
				return fmt.Errorf("write warning: %w", err)
			}
		}
		if v.albumKey != nil {
			if _, err := tx.Exec(ctx, `UPDATE submissions SET album_key = ? WHERE id = ?`, v.albumKey, id); err != nil {
				return fmt.Errorf("write album key: %w", err)
			}
		}

		to, reason := v.to, v.reason
		if to == stateAccepted {
			var err error
			if to, reason, err = accept(ctx, tx, id); err != nil {
				return err
			}
		}
		if to == stateMatching || to == stateAccepted {
			_, err := tx.Exec(ctx, `UPDATE submissions SET reserved_bytes = min(reserved_bytes,
				(SELECT coalesce(sum(size), 0) FROM submission_files
					WHERE submission_id = ?1 AND role != 'archive' AND verdict IS NULL)) WHERE id = ?1`, id)
			if err != nil {
				return fmt.Errorf("true up reservation: %w", err)
			}
		}
		return move(ctx, tx, id, at, to, reason, stateChecking)
	})
}

// writeResult stores what the gate learned of one file. A zero field is
// one the gate stopped before reaching, stored NULL.
func writeResult(ctx context.Context, tx *store.Tx, submission, file []byte, r gate.FileResult) error {
	_, err := tx.Exec(ctx, `UPDATE submission_files SET verdict = ?, sample_rate = ?, bits_per_sample = ?, channels = ?,
		samples = ?, audio_key = ?, header_md5 = ?, file_hash = ?, disc = ?, track = ?
		WHERE id = ? AND submission_id = ?`,
		nullable(string(r.Verdict)), nonzero(int64(r.SampleRate)), nonzero(int64(r.BitsPerSample)),
		nonzero(int64(r.Channels)), nonzero(int64(r.Samples)), blob(r.AudioKey), blob(r.HeaderMD5), blob(r.FileHash),
		nonzero(int64(r.Disc)), nonzero(int64(r.Track)), file, submission)
	if err != nil {
		return fmt.Errorf("write file verdict: %w", err)
	}
	return nil
}

// accept refuses duplicates in submission id's target collection as it
// moves to accepted, and returns the state it ends in with that event's
// reason. It reads and writes through the caller's transaction, the one
// that writes the move, so of two twins the second to commit is refused.
//
// An album whose key equals an accepted submission's is rejected. A
// separate track whose audio key equals a passing track of an accepted
// submission gets already-in-library, the others go on, and a
// submission left with no passing track is rejected. Duplicates in
// other collections are allowed.
func accept(ctx context.Context, tx *store.Tx, id []byte) (state, string, error) {
	var coll []byte
	var unit gate.Unit
	if err := tx.QueryRow(ctx, `SELECT collection_id, verdict_unit FROM submissions WHERE id = ?`, id).Scan(&coll, &unit); err != nil {
		return "", "", fmt.Errorf("read submission: %w", err)
	}
	duplicate := string(reject.AlreadyInLibrary)

	if unit == gate.UnitAlbum {
		var twins int
		err := tx.QueryRow(ctx, `SELECT count(*) FROM submissions s JOIN submissions twin
			ON twin.collection_id = s.collection_id AND twin.album_key = s.album_key
			WHERE s.id = ? AND twin.id != s.id AND twin.state = ?`, id, stateAccepted).Scan(&twins)
		if err != nil {
			return "", "", fmt.Errorf("find a twin album: %w", err)
		}
		if twins > 0 {
			return stateRejected, duplicate, nil
		}
		return stateAccepted, "", nil
	}

	_, err := tx.Exec(ctx, `UPDATE submission_files SET verdict = ?1
		WHERE submission_id = ?2 AND role = 'audio' AND verdict IS NULL AND audio_key IN (
			SELECT f.audio_key FROM submission_files f JOIN submissions s ON s.id = f.submission_id
			WHERE s.collection_id = ?3 AND s.state = ?4 AND s.id != ?2 AND f.role = 'audio' AND f.verdict IS NULL)`,
		duplicate, id, coll, stateAccepted)
	if err != nil {
		return "", "", fmt.Errorf("refuse twin tracks: %w", err)
	}
	var kept int
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM submission_files WHERE submission_id = ? AND role = 'audio' AND verdict IS NULL`,
		id).Scan(&kept); err != nil {
		return "", "", fmt.Errorf("count kept tracks: %w", err)
	}
	if kept == 0 {
		return stateRejected, duplicate, nil
	}
	return stateAccepted, "", nil
}

// nonzero stores a zero as NULL.
func nonzero(v int64) any {
	if v == 0 {
		return nil
	}
	return v
}

// blob stores a nil slice as NULL.
func blob(b []byte) any {
	if b == nil {
		return nil
	}
	return b
}
