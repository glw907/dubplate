package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"

	"github.com/tus/tusd/v2/pkg/handler"
)

// startup is the pass [Intake.Start] runs under the lease before any
// worker, in order: reconcile, clean up, remove the uploads whose .info
// is torn, receive the uploads whose last byte landed with no receipt,
// and wake the workers for every checking submission. It logs its
// counts. A store read that fails refuses the start; an item it cannot
// act on is logged and left for the sweep.
func (in *Intake) startup(ctx context.Context) error {
	w, err := in.cfg.Windows(ctx)
	if err != nil {
		return fmt.Errorf("read intake windows: %w", err)
	}
	expired, err := in.reconcile(ctx)
	if err != nil {
		return err
	}
	removed, err := in.clean(ctx, w.ResumeWindow)
	if err != nil {
		in.cfg.Log.Error("intake: startup cleanup", "error", err)
	}
	torn, err := in.removeTorn(ctx)
	if err != nil {
		return err
	}
	received, err := in.receiveComplete(ctx)
	if err != nil {
		return err
	}
	queued, err := in.queue(ctx)
	if err != nil {
		return fmt.Errorf("read the check queue: %w", err)
	}
	in.subs.poke()
	in.cfg.Log.Info("intake startup pass", "expired", expired, "removed", removed, "torn", torn,
		"received", received, "queued", len(queued))
	return nil
}

// stagedSet selects, for every open submission, the files its staged
// bytes must hold, each with its row's size and whether it is an
// extracted member. Loose files are their received uploads, less the
// failing tracks an accepted separate-tracks submission dropped. An
// archive is its upload until its verdict, and from the verdict that
// leaves checking for matching or accepted, its kept members, never the
// archive again.
const stagedSet = `SELECT s.id, f.id, f.rel_path, f.size, s.container = 'zip' AND f.role != 'archive'
	FROM submissions s JOIN submission_files f ON f.submission_id = s.id
	WHERE ` + openSubmission + ` AND f.verdict IS NULL AND (
		(s.container = 'files' AND f.received_at IS NOT NULL)
		OR (s.container = 'zip' AND s.state IN ('uploading', 'checking', 'check-unfinished')
			AND f.role = 'archive' AND f.received_at IS NOT NULL)
		OR (s.container = 'zip' AND s.state IN ('matching', 'accepted') AND f.role != 'archive'))
	ORDER BY s.id, f.rel_path`

// reconcile expires every open submission one of whose staged files is
// missing or shorter than its row's size, with a reason naming each.
// Staging is never backed up, so a restored store can hold rows whose
// bytes are gone; expiring them ends their reservations and frees their
// keys. A file it cannot stat is logged and its submission left.
func (in *Intake) reconcile(ctx context.Context) (int, error) {
	type stagedFile struct {
		submission, id, path []byte
		size                 int64
		member               bool
	}
	files, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) (stagedFile, error) {
		var f stagedFile
		return f, rows.Scan(&f.submission, &f.id, &f.path, &f.size, &f.member)
	}, stagedSet)
	if err != nil {
		return 0, fmt.Errorf("read staged files: %w", err)
	}

	var order [][]byte
	missing := map[string][]string{}
	unread := map[string]bool{}
	for _, f := range files {
		name := uploadName(f.id)
		if f.member {
			name = path.Join(extractDir(f.submission), hex.EncodeToString(f.id))
		}
		key := string(f.submission)
		info, err := in.root.Stat(name)
		switch {
		case errors.Is(err, fs.ErrNotExist), err == nil && info.Size() < f.size:
			if missing[key] == nil {
				order = append(order, f.submission)
			}
			missing[key] = append(missing[key], strings.ToValidUTF8(string(f.path), "�"))
		case err != nil:
			in.cfg.Log.Error("intake: reconcile could not read a staged file", "submission", wireID(f.submission), "error", err)
			unread[key] = true
		}
	}

	open := slices.DeleteFunc(slices.Clone(states), state.terminal)
	expired := 0
	for _, id := range order {
		if unread[string(id)] {
			continue
		}
		reason := "its staged files are missing or short: " + strings.Join(missing[string(id)], ", ")
		switch err := in.subs.expire(ctx, id, reason, open...); {
		case errors.Is(err, errWrongState):
		case err != nil:
			return expired, fmt.Errorf("expire %x: %w", id, err)
		default:
			expired++
		}
	}
	return expired, nil
}

// unreceivedUploads selects every unreceived upload of an uploading
// submission, with its declared size.
const unreceivedUploads = `SELECT f.id, f.size FROM submission_files f JOIN submissions s ON s.id = f.submission_id
	WHERE s.state = 'uploading' AND f.received_at IS NULL ORDER BY f.id`

type upload struct {
	id   []byte
	size int64
}

func (in *Intake) unreceived(ctx context.Context) ([]upload, error) {
	list, err := queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) (upload, error) {
		var u upload
		return u, rows.Scan(&u.id, &u.size)
	}, unreceivedUploads)
	if err != nil {
		return nil, fmt.Errorf("read unreceived uploads: %w", err)
	}
	return list, nil
}

// removeTorn removes both tus files of each unreceived upload whose
// .info does not parse. tusd writes it with no sync, so a crash can
// leave it torn, and tusd then answers the upload's HEAD with a 500 the
// client cannot get past; with both gone, the HEAD answers 404 and the
// client creates the upload again.
func (in *Intake) removeTorn(ctx context.Context) (int, error) {
	list, err := in.unreceived(ctx)
	if err != nil {
		return 0, err
	}
	removed := 0
	for _, u := range list {
		name := hex.EncodeToString(u.id)
		data, err := in.root.ReadFile(path.Join("tus", name+".info"))
		if errors.Is(err, fs.ErrNotExist) {
			continue
		}
		if err != nil {
			in.cfg.Log.Error("intake: read an upload's .info", "file", name, "error", err)
			continue
		}
		var info handler.FileInfo
		if json.Unmarshal(data, &info) == nil {
			continue
		}
		gone, err := in.removeUpload(ctx, name)
		if err != nil {
			in.cfg.Log.Error("intake: remove an upload with a torn .info", "file", name, "error", err)
		}
		if gone {
			removed++
		}
	}
	return removed, nil
}

// receiveComplete runs receipt for every upload whose data file holds
// its whole declared size with no receipt: a crash or a store fault
// between the last byte and the receipt's commit leaves one. Receipt
// syncs before it reads the row, so the size is checked first.
func (in *Intake) receiveComplete(ctx context.Context) (int, error) {
	list, err := in.unreceived(ctx)
	if err != nil {
		return 0, err
	}
	received := 0
	for _, u := range list {
		info, err := in.root.Stat(uploadName(u.id))
		if err != nil || info.Size() != u.size {
			continue
		}
		switch err := in.subs.receive(ctx, u.id); {
		case errors.Is(err, errWrongState):
		case err != nil:
			in.cfg.Log.Error("intake: receive a whole upload", "file", wireID(u.id), "error", err)
		default:
			received++
		}
	}
	return received, nil
}
