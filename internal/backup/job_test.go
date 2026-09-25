package backup_test

import (
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/local"
)

// testActor starts every test backup.
var testActor = ops.LocalActor("tester")

// newStateDir creates a current store under a fresh state directory
// and returns the directory, with no handle left open.
func newStateDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	h, err := store.OpenServe(context.Background(), dir, store.ServeOptions{
		BusyTimeout:   testBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}
	return dir
}

// openWrite opens dir's store in write mode, closed at cleanup.
func openWrite(t *testing.T, dir string) *store.Handle {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), dir, testBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// deadline returns a context bounded for a test's store call.
func deadline(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// plantTracks inserts a member, a collection, an album, and n tracks.
func plantTracks(t *testing.T, h *store.Handle, n int) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		user, coll, album := store.NewID(), store.NewID(), store.NewID()
		if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, streamer_username, created_at, updated_at)
VALUES (?, 'owner@example.org', 'Owner', 'owner', 1, 1)`, user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (?, 'main', CAST('main' AS BLOB), ?, 'group', 'server')`, coll, user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO albums (id, collection_id, created_at, updated_at) VALUES (?, ?, 1, 1)`, album, coll); err != nil {
			return err
		}
		for i := range n {
			if _, err := tx.Exec(ctx, `INSERT INTO tracks (id, album_id, collection_id, rel_path, created_at, updated_at)
VALUES (?, ?, ?, CAST(? AS BLOB), 1, 1)`, store.NewID(), album, coll, fmt.Sprintf("t%d.flac", i)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("plant %d tracks: %v", n, err)
	}
}

// writeTree creates each file under dir, with its parent directories,
// holding its own name.
func writeTree(t *testing.T, dir string, files ...string) {
	t.Helper()
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", f, err)
		}
		if err := os.WriteFile(p, []byte(f), 0o644); err != nil {
			t.Fatalf("write %s: %v", f, err)
		}
	}
}

// treeFiles lists the regular files under dir, relative to it; an
// absent dir lists none.
func treeFiles(t *testing.T, dir string) []string {
	t.Helper()
	var files []string
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrNotExist) && p == dir {
			return filepath.SkipAll
		}
		if err != nil {
			return err
		}
		if d.Type().IsRegular() {
			rel, err := filepath.Rel(dir, p)
			if err != nil {
				return err
			}
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
	return files
}

// newRunner opens a production runner over h with backup on, closed at
// cleanup before h.
func newRunner(t *testing.T, h *store.Handle, remote, root string, logs *syncBuffer) *local.Runner {
	t.Helper()
	logger := slog.New(slog.DiscardHandler)
	if logs != nil {
		logger = slog.New(slog.NewTextHandler(logs, nil))
	}
	r, err := local.New(context.Background(), local.Config{
		Store:  h,
		Backup: local.BackupConfig{Enabled: true, Remote: remote, LibraryRoot: root},
		Logger: logger,
	})
	if err != nil {
		t.Fatalf("local.New: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

// start starts a backup through r.
func start(t *testing.T, r *local.Runner) ops.Handle {
	t.Helper()
	resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	if err != nil {
		t.Fatalf("StartBackup: %v", err)
	}
	return resp.JobID
}

// await waits for the job h names to end and returns its status.
func await(t *testing.T, r *local.Runner, h ops.Handle) ops.JobStatus {
	t.Helper()
	end := time.Now().Add(60 * time.Second)
	for {
		resp, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: h})
		if err != nil {
			t.Fatalf("JobStatus: %v", err)
		}
		switch resp.Job.State {
		case ops.JobQueued, ops.JobRunning:
		default:
			return resp.Job
		}
		if time.Now().After(end) {
			t.Fatalf("job %s still %s after 60s", h, resp.Job.State)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// runBackup runs one backup to its end through a runner over h.
func runBackup(t *testing.T, h *store.Handle, remote, root string) (ops.Handle, ops.JobStatus) {
	t.Helper()
	r := newRunner(t, h, remote, root, nil)
	id := start(t, r)
	st := await(t, r, id)
	if err := r.Close(); err != nil {
		t.Fatalf("close runner: %v", err)
	}
	return id, st
}

// wantState fails unless st ended in state with code.
func wantState(t *testing.T, st ops.JobStatus, state ops.JobState, code string) {
	t.Helper()
	if st.State != state || st.FailureCode != code {
		t.Fatalf("job ended %s, code %q, message %q; want %s, code %q", st.State, st.FailureCode, st.FailureMessage, state, code)
	}
}

// backupStages lists staging/backup-* in dir's store.
func backupStages(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "store", "staging"))
	if err != nil {
		t.Fatalf("read staging: %v", err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "backup-") {
			names = append(names, e.Name())
		}
	}
	return names
}

// localRemote configures a real rclone local remote over a fresh
// directory, with no configuration file of the user's in reach, and
// returns the remote and its directory. rclone is a prerequisite: its
// absence fails the test.
func localRemote(t *testing.T) (string, string) {
	t.Helper()
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}
	conf := filepath.Join(t.TempDir(), "rclone.conf")
	if err := os.WriteFile(conf, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RCLONE_CONFIG", conf)
	t.Setenv("RCLONE_CONFIG_BACKUPTEST_TYPE", "local")
	dir := t.TempDir()
	return "backuptest:" + dir, dir
}

// snapshotStamps lists the stamps of the snapshots under a local
// remote's directory, oldest first, failing on any other name.
func snapshotStamps(t *testing.T, remoteDir string) []string {
	t.Helper()
	var stamps []string
	for _, name := range treeFiles(t, filepath.Join(remoteDir, "snapshots")) {
		stamp, ok := strings.CutSuffix(name, ".db")
		if _, err := store.ParseStamp(stamp); !ok || err != nil {
			t.Fatalf("snapshots/ holds %q, not <stamp>.db", name)
		}
		stamps = append(stamps, stamp)
	}
	slices.Sort(stamps)
	return stamps
}

// openCopy opens a store copy read-only.
func openCopy(t *testing.T, path string) *sql.DB {
	t.Helper()
	dsn := (&url.URL{Scheme: "file", Path: path}).String() + "?mode=ro"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		t.Fatalf("open copy: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestBackup_RealRemote is spec criterion 10: every library file lands
// under library/, one object under snapshots/ that passes the
// integrity check at the binary's version, and its metadata names the
// job and when the sync completed.
func TestBackup_RealRemote(t *testing.T) {
	remote, remoteDir := localRemote(t)
	dir := newStateDir(t)
	h := openWrite(t, dir)
	plantTracks(t, h, 2)
	lib := t.TempDir()
	writeTree(t, lib, "artist/one.flac", "artist/two.flac")

	id, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	if got, want := treeFiles(t, filepath.Join(remoteDir, "library")), []string{"artist/one.flac", "artist/two.flac"}; !slices.Equal(got, want) {
		t.Errorf("remote library/ holds %q, want %q", got, want)
	}
	if got := treeFiles(t, filepath.Join(remoteDir, "archive")); len(got) != 0 {
		t.Errorf("remote archive/ holds %q after a first run, want nothing", got)
	}
	stamps := snapshotStamps(t, remoteDir)
	if len(stamps) != 1 {
		t.Fatalf("snapshots/ holds %q, want one", stamps)
	}
	snap := filepath.Join(remoteDir, "snapshots", stamps[0]+".db")
	db := openCopy(t, snap)
	var integrity string
	if err := db.QueryRowContext(deadline(t), "PRAGMA integrity_check").Scan(&integrity); err != nil || integrity != "ok" {
		t.Errorf("snapshot integrity_check = %q, %v; want ok", integrity, err)
	}
	var got, want int
	if err := db.QueryRowContext(deadline(t), "PRAGMA user_version").Scan(&got); err != nil {
		t.Fatalf("snapshot user_version: %v", err)
	}
	if err := h.QueryRow(deadline(t), "PRAGMA user_version").Scan(&want); err != nil {
		t.Fatalf("live user_version: %v", err)
	}
	if got != want || got == 0 {
		t.Errorf("snapshot user_version = %d, want the live store's %d", got, want)
	}

	out, err := exec.Command("rclone", "lsjson", "-M", remote+"/snapshots").Output()
	if err != nil {
		t.Fatalf("rclone lsjson: %v", err)
	}
	var listing []struct {
		Name     string
		Metadata map[string]string
	}
	if err := json.Unmarshal(out, &listing); err != nil {
		t.Fatalf("parse lsjson %q: %v", out, err)
	}
	if len(listing) != 1 {
		t.Fatalf("lsjson lists %d objects, want 1", len(listing))
	}
	md := listing[0].Metadata
	if md["run_id"] != string(id) {
		t.Errorf("run_id = %q, want the job's handle %q", md["run_id"], id)
	}
	if _, err := time.Parse(time.RFC3339, md["sync_completed"]); err != nil {
		t.Errorf("sync_completed %q: %v", md["sync_completed"], err)
	}
	if stages := backupStages(t, dir); len(stages) != 0 {
		t.Errorf("staging/ holds %q after the send, want no stage", stages)
	}
}

// TestBackup_SymlinkRoot is the symlink row: a root reached through a
// symlink to a two-file directory passes the guard, which with zero
// catalog tracks refuses an empty tree, and syncs both files.
func TestBackup_SymlinkRoot(t *testing.T) {
	remote, remoteDir := localRemote(t)
	h := openWrite(t, newStateDir(t))
	base := t.TempDir()
	lib := filepath.Join(base, "lib")
	writeTree(t, lib, "a/one.flac", "a/two.flac")
	link := filepath.Join(base, "link")
	if err := os.Symlink(lib, link); err != nil {
		t.Fatal(err)
	}

	_, st := runBackup(t, h, remote, link)
	wantState(t, st, ops.JobSucceeded, "")
	if got, want := treeFiles(t, filepath.Join(remoteDir, "library")), []string{"a/one.flac", "a/two.flac"}; !slices.Equal(got, want) {
		t.Errorf("remote library/ holds %q, want %q", got, want)
	}
}

// TestBackup_Archive is spec criterion 11: a file deleted between two
// runs lands under the second run's archive/<stamp>/, not deleted.
func TestBackup_Archive(t *testing.T) {
	remote, remoteDir := localRemote(t)
	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 2)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac", "a/two.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")
	if err := os.Remove(filepath.Join(lib, "a", "two.flac")); err != nil {
		t.Fatal(err)
	}
	_, st = runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	stamps := snapshotStamps(t, remoteDir)
	if len(stamps) != 2 {
		t.Fatalf("snapshots/ holds %q, want two runs' snapshots", stamps)
	}
	second := stamps[1]
	if got, want := treeFiles(t, filepath.Join(remoteDir, "archive")), []string{second + "/a/two.flac"}; !slices.Equal(got, want) {
		t.Errorf("remote archive/ holds %q, want %q", got, want)
	}
	if got, want := treeFiles(t, filepath.Join(remoteDir, "library")), []string{"a/one.flac"}; !slices.Equal(got, want) {
		t.Errorf("remote library/ holds %q, want %q", got, want)
	}
}

// TestBackup_ZeroTracksKeepsRemote is spec criterion 13: with zero
// catalog tracks, a remote library/ of two files, and a root holding
// one of them, --max-delete 0 refuses the move into the archive. rclone
// exits 7 and the remote keeps both files with nothing archived.
func TestBackup_ZeroTracksKeepsRemote(t *testing.T) {
	remote, remoteDir := localRemote(t)
	h := openWrite(t, newStateDir(t))
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")
	writeTree(t, filepath.Join(remoteDir, "library"), "a/one.flac", "a/two.flac")
	mtime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range []string{filepath.Join(lib, "a/one.flac"), filepath.Join(remoteDir, "library/a/one.flac")} {
		if err := os.Chtimes(p, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobFailed, "rclone-exit-7")
	if got, want := treeFiles(t, filepath.Join(remoteDir, "library")), []string{"a/one.flac", "a/two.flac"}; !slices.Equal(got, want) {
		t.Errorf("remote library/ holds %q, want both files", got)
	}
	if got := treeFiles(t, filepath.Join(remoteDir, "archive")); len(got) != 0 {
		t.Errorf("remote archive/ holds %q, want nothing", got)
	}
	if got := treeFiles(t, filepath.Join(remoteDir, "snapshots")); len(got) != 0 {
		t.Errorf("snapshots/ holds %q after a failed sync, want nothing", got)
	}
}

// TestBackup_SendsCommittedRows is spec criterion 9's job half: with a
// writer committing through the write method across the run, the
// snapshot sent holds every row committed before the job began.
func TestBackup_SendsCommittedRows(t *testing.T) {
	remote, remoteDir := localRemote(t)
	h := openWrite(t, newStateDir(t))
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	var committed atomic.Int64
	insert := func() error {
		n := committed.Load()
		return h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, streamer_username, created_at, updated_at)
VALUES (?, ?, 'Member', ?, 1, 1)`, store.NewID(), fmt.Sprintf("m%d@example.org", n), fmt.Sprintf("m%d", n))
			return err
		})
	}
	for range 20 {
		if err := insert(); err != nil {
			t.Fatalf("populate: %v", err)
		}
		committed.Add(1)
	}
	stop, done := make(chan struct{}), make(chan error, 1)
	go func() {
		for {
			select {
			case <-stop:
				done <- nil
				return
			default:
			}
			if err := insert(); err != nil {
				done <- err
				return
			}
			committed.Add(1)
		}
	}()
	for committed.Load() < 30 {
		time.Sleep(time.Millisecond)
	}

	r := newRunner(t, h, remote, lib, nil)
	before := committed.Load()
	st := await(t, r, start(t, r))
	close(stop)
	if err := <-done; err != nil {
		t.Fatalf("looping writer: %v", err)
	}
	wantState(t, st, ops.JobSucceeded, "")

	stamps := snapshotStamps(t, remoteDir)
	if len(stamps) != 1 {
		t.Fatalf("snapshots/ holds %q, want one", stamps)
	}
	var got int64
	db := openCopy(t, filepath.Join(remoteDir, "snapshots", stamps[0]+".db"))
	if err := db.QueryRowContext(deadline(t), "SELECT count(*) FROM users").Scan(&got); err != nil {
		t.Fatalf("count users: %v", err)
	}
	if got < before {
		t.Errorf("snapshot holds %d rows, want at least the %d committed before the job began", got, before)
	}
}

// TestBackup_Refusals is spec criterion 12 and the rows that refuse
// before rclone: each ends failed with its code, and the stand-in
// records no invocation.
func TestBackup_Refusals(t *testing.T) {
	tests := []struct {
		name     string
		tracks   int
		root     func(t *testing.T, base string) string
		remote   string
		noRclone bool
		wantCode string
		wantMsg  string
	}{
		{
			name:     "missing root",
			root:     func(_ *testing.T, base string) string { return filepath.Join(base, "absent") },
			wantCode: "library-missing",
		},
		{
			name: "dangling symlink",
			root: func(t *testing.T, base string) string {
				link := filepath.Join(base, "link")
				if err := os.Symlink(filepath.Join(base, "gone"), link); err != nil {
					t.Fatal(err)
				}
				return link
			},
			wantCode: "library-missing",
		},
		{
			name: "root is a regular file",
			root: func(t *testing.T, base string) string {
				writeTree(t, base, "file")
				return filepath.Join(base, "file")
			},
			wantCode: "library-missing",
		},
		{
			name: "empty root, zero tracks",
			root: func(t *testing.T, base string) string {
				lib := filepath.Join(base, "lib")
				if err := os.MkdirAll(filepath.Join(lib, "artist"), 0o755); err != nil {
					t.Fatal(err)
				}
				return lib
			},
			wantCode: "library-empty",
		},
		{
			name:   "one file under four tracks",
			tracks: 4,
			root: func(t *testing.T, base string) string {
				writeTree(t, base, "lib/a/one.flac")
				return filepath.Join(base, "lib")
			},
			wantCode: "library-short",
			wantMsg:  "holds 1 files, fewer than half the catalog's 4 tracks",
		},
		{
			name: "unreadable root",
			root: func(t *testing.T, base string) string {
				writeTree(t, base, "lib/a/one.flac")
				lib := filepath.Join(base, "lib")
				if err := os.Chmod(lib, 0); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(lib, 0o755) })
				return lib
			},
			wantCode: "library-unreadable",
			wantMsg:  "permission denied",
		},
		{
			name:     "remote with no colon",
			root:     func(t *testing.T, base string) string { writeTree(t, base, "lib/a"); return filepath.Join(base, "lib") },
			remote:   "relative/dir",
			wantCode: "remote-invalid",
		},
		{
			name:     "remote name outside the set",
			root:     func(t *testing.T, base string) string { writeTree(t, base, "lib/a"); return filepath.Join(base, "lib") },
			remote:   "my-remote:dir",
			wantCode: "remote-invalid",
		},
		{
			name:     "rclone not on PATH",
			root:     func(_ *testing.T, base string) string { return filepath.Join(base, "absent") },
			noRclone: true,
			wantCode: "rclone-missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantCode == "library-unreadable" && os.Geteuid() == 0 {
				t.Skip("root reads a mode-0 directory")
			}
			s := useStandin(t, standinConfig{})
			if tt.noRclone {
				t.Setenv("PATH", t.TempDir())
			}
			h := openWrite(t, newStateDir(t))
			plantTracks(t, h, tt.tracks)
			remote := cmp.Or(tt.remote, "name:dir")
			_, st := runBackup(t, h, remote, tt.root(t, t.TempDir()))
			wantState(t, st, ops.JobFailed, tt.wantCode)
			if !strings.Contains(st.FailureMessage, tt.wantMsg) {
				t.Errorf("failure message %q does not name %q", st.FailureMessage, tt.wantMsg)
			}
			if argvs := s.invocations(t); len(argvs) != 0 {
				t.Errorf("the stand-in recorded %q, want no invocation", argvs)
			}
		})
	}
}

// TestBackup_Destinations is the join rows: for name:, name:dir, and
// name:dir/, the stand-in's argv carries exactly the table's
// destinations, one stamp across them, the resolved root as the sync's
// source, and --max-delete equal to the catalog's two tracks. A
// successful send is followed by the prune's own listing of
// snapshots/ under the same destination.
func TestBackup_Destinations(t *testing.T) {
	tests := []struct {
		remote string
		prefix string
	}{
		{remote: "name:", prefix: "name:"},
		{remote: "name:dir", prefix: "name:dir/"},
		{remote: "name:dir/", prefix: "name:dir/"},
	}
	for _, tt := range tests {
		t.Run(tt.remote, func(t *testing.T) {
			s := useStandin(t, standinConfig{})
			h := openWrite(t, newStateDir(t))
			plantTracks(t, h, 2)
			lib := t.TempDir()
			writeTree(t, lib, "a/one.flac", "a/two.flac")

			id, st := runBackup(t, h, tt.remote, lib)
			wantState(t, st, ops.JobSucceeded, "")
			argvs := s.invocations(t)
			if len(argvs) != 3 {
				t.Fatalf("the stand-in recorded %q, want sync, copyto, then the prune's listing", argvs)
			}
			sync, copyto, list := argvs[0], argvs[1], argvs[2]
			if want := []string{"lsjson", tt.prefix + "snapshots"}; !slices.Equal(list, want) {
				t.Errorf("prune's listing argv = %q, want %q", list, want)
			}
			if len(sync) != 7 || sync[0] != "sync" || sync[3] != "--backup-dir" || sync[5] != "--max-delete" {
				t.Fatalf("sync argv = %q", sync)
			}
			stamp, ok := strings.CutPrefix(sync[4], tt.prefix+"archive/")
			if _, err := store.ParseStamp(stamp); !ok || err != nil {
				t.Fatalf("--backup-dir %q is not %sarchive/<stamp>", sync[4], tt.prefix)
			}
			wantSync := []string{"sync", lib, tt.prefix + "library", "--backup-dir", tt.prefix + "archive/" + stamp, "--max-delete", "2"}
			if !slices.Equal(sync, wantSync) {
				t.Errorf("sync argv = %q, want %q", sync, wantSync)
			}
			if len(copyto) != 8 || copyto[0] != "copyto" {
				t.Fatalf("copyto argv = %q", copyto)
			}
			completed, ok := strings.CutPrefix(copyto[7], "sync_completed=")
			if _, err := time.Parse(time.RFC3339, completed); !ok || err != nil {
				t.Errorf("copyto's last argument %q is not sync_completed=<RFC 3339>", copyto[7])
			}
			wantCopy := []string{"copyto", copyto[1], tt.prefix + "snapshots/" + stamp + ".db", "-M",
				"--metadata-set", "run_id=" + string(id), "--metadata-set", copyto[7]}
			if !slices.Equal(copyto, wantCopy) {
				t.Errorf("copyto argv = %q, want %q", copyto, wantCopy)
			}
		})
	}
}

// TestBackup_TransferFailures pins the transfer's failure rows against
// the stand-in: a nonzero sync exit is rclone-exit-<N> (spec criterion
// 14 for 7) with rclone's last stderr line as the message; a failed
// send removes the stage; a refused copy is snapshot-failed and sends
// nothing.
func TestBackup_TransferFailures(t *testing.T) {
	tests := []struct {
		name        string
		exit        string
		lockStaging bool
		wantCode    string
		wantMsg     string
		wantSubs    []string
	}{
		{name: "sync exits 7", exit: "sync=7", wantCode: "rclone-exit-7", wantMsg: "stand-in sync exits 7", wantSubs: []string{"sync"}},
		{name: "sync exits 3", exit: "sync=3", wantCode: "rclone-exit-3", wantMsg: "stand-in sync exits 3", wantSubs: []string{"sync"}},
		{name: "send exits 1", exit: "copyto=1", wantCode: "rclone-exit-1", wantMsg: "stand-in copyto exits 1", wantSubs: []string{"sync", "copyto"}},
		{name: "copy refused", lockStaging: true, wantCode: "snapshot-failed", wantMsg: "permission denied", wantSubs: []string{"sync"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := useStandin(t, standinConfig{exit: tt.exit})
			dir := newStateDir(t)
			h := openWrite(t, dir)
			lib := t.TempDir()
			writeTree(t, lib, "a/one.flac")
			if tt.lockStaging {
				staging := filepath.Join(dir, "store", "staging")
				if err := os.Chmod(staging, 0o500); err != nil {
					t.Fatal(err)
				}
				t.Cleanup(func() { _ = os.Chmod(staging, 0o700) })
			}

			_, st := runBackup(t, h, "name:dir", lib)
			wantState(t, st, ops.JobFailed, tt.wantCode)
			if !strings.Contains(st.FailureMessage, tt.wantMsg) {
				t.Errorf("failure message %q does not name %q", st.FailureMessage, tt.wantMsg)
			}
			var subs []string
			for _, argv := range s.invocations(t) {
				subs = append(subs, argv[0])
			}
			if !slices.Equal(subs, tt.wantSubs) {
				t.Errorf("the stand-in ran %q, want %q", subs, tt.wantSubs)
			}
			if stages := backupStages(t, dir); len(stages) != 0 {
				t.Errorf("staging/ holds %q after the run, want no stage", stages)
			}
		})
	}
}

// TestBackup_StageNotRemoved pins the stage-removal row: a stage that
// cannot be removed after the send is logged, and the job still
// succeeds.
func TestBackup_StageNotRemoved(t *testing.T) {
	useStandin(t, standinConfig{lockStage: true})
	dir := newStateDir(t)
	h := openWrite(t, dir)
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "store", "staging"), 0o700) })
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	logs := &syncBuffer{}
	r := newRunner(t, h, "name:dir", lib, logs)
	st := await(t, r, start(t, r))
	wantState(t, st, ops.JobSucceeded, "")
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if got := logs.String(); !strings.Contains(got, "backup stage not removed") {
		t.Errorf("logs %q do not record the stage left behind", got)
	}
	if stages := backupStages(t, dir); len(stages) != 1 {
		t.Errorf("staging/ holds %q, want the one stage left behind", stages)
	}
}

// TestBackup_CancelPrecedence is the cancel precedence rows. A
// SIGINT-honoring stand-in canceled mid-sync receives one SIGINT,
// exits 130, and the job ends canceled, never rclone-exit-130. A
// stand-in the test kills with SIGKILL, with no cancel, ends the job
// failed with rclone-signal-KILL.
func TestBackup_CancelPrecedence(t *testing.T) {
	tests := []struct {
		name      string
		cancel    bool
		wantState ops.JobState
		wantCode  string
		wantExit  int
	}{
		{name: "canceled", cancel: true, wantState: ops.JobCanceled, wantExit: 1},
		{name: "killed", wantState: ops.JobFailed, wantCode: "rclone-signal-KILL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := useStandin(t, standinConfig{sleep: "sync", sigint: "honor"})
			h := openWrite(t, newStateDir(t))
			lib := t.TempDir()
			writeTree(t, lib, "a/one.flac")
			r := newRunner(t, h, "name:dir", lib, nil)
			id := start(t, r)
			pid := s.awaitPID(t)
			killAtCleanup(t, pid)

			if tt.cancel {
				if _, err := r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: id}); err != nil {
					t.Fatalf("CancelJob: %v", err)
				}
			} else if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
				t.Fatalf("kill the stand-in: %v", err)
			}
			wantState(t, await(t, r, id), tt.wantState, tt.wantCode)
			if got := s.count(t, "exit 130"); got != tt.wantExit {
				t.Errorf("the stand-in recorded %d exits of 130, want %d", got, tt.wantExit)
			}
			if got := s.count(t, "SIGINT"); got != tt.wantExit {
				t.Errorf("the stand-in recorded %d SIGINTs, want %d", got, tt.wantExit)
			}
		})
	}
}

// TestBackup_OrphanTransfer is the orphan row. A re-executed child runs
// a backup through a real runner against a stand-in that ignores
// SIGINT and sleeps, and is killed. The stand-in outlives it holding
// the kind lease and the store lock it inherited, so a start from a
// runner built afterward is a conflict naming a transfer still
// running. Once the stand-in exits, a start succeeds.
func TestBackup_OrphanTransfer(t *testing.T) {
	s := useStandin(t, standinConfig{sleep: "sync", sigint: "ignore"})
	dir := newStateDir(t)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	var stderr syncBuffer
	child := exec.Command(os.Args[0], "-test.run=^$")
	child.Env = append(os.Environ(), childDirEnv+"="+dir, childRemoteEnv+"=name:dir", childRootEnv+"="+lib)
	child.Stderr = &stderr
	child.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := child.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Cleanup(func() { _ = child.Process.Kill(); _ = child.Wait() })
	pid := s.awaitPID(t)
	killAtCleanup(t, pid)
	if err := child.Process.Kill(); err != nil {
		t.Fatalf("kill child: %v", err)
	}
	_ = child.Wait()
	if err := syscall.Kill(pid, 0); err != nil {
		t.Fatalf("the stand-in did not outlive the child: %v; child stderr: %s", err, stderr.String())
	}

	h := openWrite(t, dir)
	configureStandin(t, standinConfig{})
	r := newRunner(t, h, "name:dir", lib, nil)
	_, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	if !errors.Is(err, ops.ErrConflict) || !strings.Contains(err.Error(), "a previous run's transfer is still running") {
		t.Fatalf("StartBackup() with the orphan alive = %v, want a conflict naming a transfer still running", err)
	}

	if err := syscall.Kill(pid, syscall.SIGKILL); err != nil {
		t.Fatalf("kill the stand-in: %v", err)
	}
	end := time.Now().Add(20 * time.Second)
	for {
		resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
		if err == nil {
			wantState(t, await(t, r, resp.JobID), ops.JobSucceeded, "")
			return
		}
		if !errors.Is(err, ops.ErrConflict) || time.Now().After(end) {
			t.Fatalf("StartBackup() after the orphan exited = %v, want success", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
