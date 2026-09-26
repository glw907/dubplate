package backup

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/engine/store"
)

// latestSnapshot names, in place of a stamp, the newest snapshot on the
// remote.
const latestSnapshot = "latest"

// A source is the remote a restore reads its store and its library
// from, through a transport whose child inherits the restore's store
// lock.
type source struct {
	t      transport
	remote Remote
}

// A snapshot is one object under the remote's snapshots/: the stamp it
// is named by, and its size in bytes, -1 when the backend reports
// none.
type snapshot struct {
	stamp string
	size  int64
}

// resolve decides which snapshot a restore reads from want, a stamp or
// [latestSnapshot]. It refuses a stamp that does not parse before any
// rclone run, and otherwise decides from snapshots/'s listing alone,
// as [pickSnapshot] does: a copyto of a missing object ends on an
// error that names no path.
func (s source) resolve(ctx context.Context, want string, older bool) (snapshot, error) {
	if want != latestSnapshot {
		if _, err := store.ParseStamp(want); err != nil {
			return snapshot{}, fmt.Errorf("restore snapshot: %w", err)
		}
	}
	dir := s.remote.join("snapshots")
	entries, err := s.t.list(ctx, dir, false)
	if err != nil {
		return snapshot{}, err
	}
	return pickSnapshot(entries, dir, want, older)
}

// pickSnapshot picks want from entries, the listing of the remote path
// dir. The latest snapshot is the greatest stamp among the names that
// parse as <stamp>.db, whatever order the listing gives: stamps sort
// lexicographically in time order. It refuses a stamp the listing
// lacks, naming its remote path; latest on a listing with no
// snapshot, naming dir; and a stamp older than the latest unless older
// is set, naming the latest.
func pickSnapshot(entries []remoteEntry, dir, want string, older bool) (snapshot, error) {
	sizes := make(map[string]int64)
	var latest string
	for _, e := range entries {
		stamp, ok := e.snapshotStamp()
		if !ok {
			continue
		}
		sizes[stamp] = e.Size
		latest = max(latest, stamp)
	}

	if want == latestSnapshot {
		if latest == "" {
			return snapshot{}, fmt.Errorf("no snapshot under %s", dir)
		}
		return snapshot{stamp: latest, size: sizes[latest]}, nil
	}
	size, ok := sizes[want]
	if !ok {
		return snapshot{}, fmt.Errorf("no snapshot %s/%s.db in the remote's listing", dir, want)
	}
	if want != latest && !older {
		return snapshot{}, fmt.Errorf("snapshot %s is not the latest snapshot %s: restore an older snapshot with --older", want, latest)
	}
	return snapshot{stamp: want, size: size}, nil
}

// fetch copies snap from the remote into sess's stage and sets the
// stage's mode to 0600. It first checks staging/ has room for snap,
// skipping the check with a line to out when the backend reports no
// size.
//
// copyto downloads to a temporary file and renames it into place, so
// the stage is a new file at whatever mode the umask leaves (0644
// under 022), and a stage created beforehand would be replaced; the
// chmod, not a pre-created file, is what makes the stage private. The
// leaf's directories are 0700, so nothing reads it in between. The
// fetch then syncs the stage and staging/, as [syncStage] does. A
// failed or canceled fetch removes the stage.
func (s source) fetch(ctx context.Context, sess *store.RestoreSession, out io.Writer, snap snapshot) error {
	src := s.remote.join("snapshots/" + snap.stamp + ".db")
	if snap.size < 0 {
		if _, err := fmt.Fprintf(out, "skipping the store free-space check: %s reports no size\n", src); err != nil {
			return fmt.Errorf("fetch %s: %w", src, err)
		}
	} else if err := sess.CheckSpace(snap.size); err != nil {
		return fmt.Errorf("fetch %s: %w", src, err)
	}

	stage := sess.StagePath()
	if err := s.t.run(ctx, nil, "copyto", src, stage); err != nil {
		if f, ok := errors.AsType[*failure](err); ok {
			err = fmt.Errorf("%s: %w", f.code, err)
		}
		return errors.Join(fmt.Errorf("fetch %s: %w", src, err), removeStage(stage))
	}
	if err := os.Chmod(stage, 0o600); err != nil {
		return errors.Join(fmt.Errorf("fetch %s: %w", src, err), removeStage(stage))
	}
	if err := syncStage(stage); err != nil {
		return errors.Join(fmt.Errorf("fetch %s: %w", src, err), removeStage(stage))
	}
	return nil
}

// syncStage flushes the stage rclone wrote, and its directory entry, to
// disk. rclone's local backend offers no fsync, and the swap syncs only
// directories, so without it a power cut after the swap could leave
// db/ naming a file whose blocks were never written.
func syncStage(stage string) error {
	// O_RDWR, since a read-only open of a FIFO planted at the stage
	// path would block until a writer came.
	f, err := os.OpenFile(stage, os.O_RDWR, 0)
	if err != nil {
		return fmt.Errorf("sync the stage: %w", err)
	}
	if err := errors.Join(f.Sync(), f.Close()); err != nil {
		return fmt.Errorf("sync the stage: %w", err)
	}
	d, err := os.Open(filepath.Dir(stage))
	if err != nil {
		return fmt.Errorf("sync staging: %w", err)
	}
	if err := errors.Join(d.Sync(), d.Close()); err != nil {
		return fmt.Errorf("sync staging: %w", err)
	}
	return nil
}

// removeStage removes the stage a fetch left, if any.
func removeStage(stage string) error {
	if err := os.Remove(stage); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("remove stage: %w", err)
	}
	return nil
}

// A treeSize is rclone size --json's measure of a remote tree. Bytes
// is a floor: some backends do not always report a file's size, and
// rclone counts each such file as empty and in Sizeless.
type treeSize struct {
	Count    int64 `json:"count"`
	Bytes    int64 `json:"bytes"`
	Sizeless int64 `json:"sizeless"`
}

// librarySize measures the remote's library/. rclone exits 3 on an
// absent directory, so a caller lists the remote's root first.
func (s source) librarySize(ctx context.Context) (treeSize, error) {
	dir := s.remote.join("library")
	var stdout bytes.Buffer
	if err := s.t.run(ctx, &stdout, "size", "--json", dir); err != nil {
		return treeSize{}, fmt.Errorf("measure %s: %w", dir, err)
	}
	size, err := parseTreeSize(stdout.Bytes())
	if err != nil {
		return treeSize{}, fmt.Errorf("measure %s: %w", dir, err)
	}
	return size, nil
}

// parseTreeSize parses rclone size --json's output.
func parseTreeSize(b []byte) (treeSize, error) {
	var size treeSize
	if err := json.Unmarshal(b, &size); err != nil {
		return treeSize{}, fmt.Errorf("parse rclone size: %w", err)
	}
	return size, nil
}

// remoteWarning returns a warning naming both remotes when configured,
// the remote a restore reads, is not referenced, the remote the backup
// reference names, as rclone reads each; it returns "" when they
// agree or referenced is empty. Before the store carries an identity,
// a differing remote is the one sign that a snapshot may belong to
// another store.
func remoteWarning(configured Remote, referenced string) string {
	if referenced == "" {
		return ""
	}
	if ref, err := ParseRemote(referenced); err == nil && ref.join("") == configured.join("") {
		return ""
	}
	return fmt.Sprintf("warning: restoring from %s, but the backup reference names %s", configured.join(""), referenced)
}
