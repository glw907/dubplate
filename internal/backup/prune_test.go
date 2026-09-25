package backup_test

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// seedSnapshots writes an empty file for each of stamps, plus each
// name in extra, under remoteDir's snapshots/.
func seedSnapshots(t *testing.T, remoteDir string, stamps []string, extra ...string) {
	t.Helper()
	dir := filepath.Join(remoteDir, "snapshots")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, stamp := range stamps {
		if err := os.WriteFile(filepath.Join(dir, stamp+".db"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range extra {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// seedArchiveDir creates remoteDir's archive/name/ holding one file,
// so it exists for rclone's lsjson --dirs-only to list.
func seedArchiveDir(t *testing.T, remoteDir, name string) {
	t.Helper()
	dir := filepath.Join(remoteDir, "archive", name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "marker.flac"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// archiveDirNames lists remoteDir's archive/'s immediate
// subdirectories, sorted; an absent archive/ lists none.
func archiveDirNames(t *testing.T, remoteDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(remoteDir, "archive"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read archive: %v", err)
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)
	return names
}

// localSnapshotNames lists the store's local snapshots/ leaf under
// dir, a state directory.
func localSnapshotNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(dir, "store", "snapshots"))
	if err != nil {
		t.Fatalf("read local snapshots: %v", err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	slices.Sort(names)
	return names
}

// plantLocalCopy writes name under dir with content and mtime, making
// dir first if it is not already there.
func plantLocalCopy(t *testing.T, dir, name string, mtime time.Time) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte("x"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(p, mtime, mtime); err != nil {
		t.Fatal(err)
	}
}

// backupRefName returns the one name in names shaped like a backup
// reference, or "" for none.
func backupRefName(names []string) string {
	for _, n := range names {
		if strings.HasPrefix(n, "backup-") && strings.HasSuffix(n, ".ref") {
			return n
		}
	}
	return ""
}

// TestPrune_RealRemote is spec criterion 15: a remote pre-seeded with
// 15 stamped snapshots, an unparseable notes.db, two archive
// directories at 100 and 10 days, and one named keep-me. After a
// successful run the remote holds exactly the run's own snapshot and
// the 13 newest seeded ones, notes.db, the 10-day archive, and
// keep-me; the 100-day archive is purged.
func TestPrune_RealRemote(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	var seeded []string
	for i := 1; i <= 15; i++ {
		seeded = append(seeded, store.FormatStamp(now.Add(-time.Duration(i)*time.Hour)))
	}
	slices.Sort(seeded)
	seedSnapshots(t, remoteDir, seeded, "notes.db")
	tenDay := store.FormatStamp(now.Add(-10 * 24 * time.Hour))
	seedArchiveDir(t, remoteDir, store.FormatStamp(now.Add(-100*24*time.Hour)))
	seedArchiveDir(t, remoteDir, tenDay)
	seedArchiveDir(t, remoteDir, "keep-me")

	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 2)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac", "a/two.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	allNames := treeFiles(t, filepath.Join(remoteDir, "snapshots"))
	if !slices.Contains(allNames, "notes.db") {
		t.Errorf("snapshots/ lost the unparseable notes.db: holds %q", allNames)
	}
	var stamps []string
	for _, n := range allNames {
		if n == "notes.db" {
			continue
		}
		stamp, ok := strings.CutSuffix(n, ".db")
		if !ok {
			t.Fatalf("snapshots/ holds unexpected name %q", n)
		}
		if _, err := store.ParseStamp(stamp); err != nil {
			t.Fatalf("snapshots/ holds %q, not <stamp>.db: %v", n, err)
		}
		stamps = append(stamps, stamp)
	}
	slices.Sort(stamps)
	if len(stamps) != 14 {
		t.Fatalf("snapshots/ holds %d stamps (own + kept seeded), want 14: %q", len(stamps), stamps)
	}
	own, kept := stamps[len(stamps)-1], stamps[:len(stamps)-1]
	wantKept := seeded[2:] // the 13 newest of the 15 seeded, the 2 oldest dropped
	if !slices.Equal(kept, wantKept) {
		t.Errorf("kept seeded stamps = %q, want the 13 newest %q", kept, wantKept)
	}
	if own <= seeded[len(seeded)-1] {
		t.Errorf("own stamp %q is not the newest, want it after every seeded stamp", own)
	}

	wantArchive := []string{"keep-me", tenDay}
	slices.Sort(wantArchive)
	if got := archiveDirNames(t, remoteDir); !slices.Equal(got, wantArchive) {
		t.Errorf("archive/ holds %q, want %q (the 100-day directory purged)", got, wantArchive)
	}
}

// TestPrune_ClockSkew is the clock row: a remote seeded with 15
// snapshots stamped a day in the future. The safety rule finds no
// stamp before the run's own and skips every prune step, so all 16
// snapshots survive and the run still succeeds, naming the skip in
// its progress.
func TestPrune_ClockSkew(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	var seeded []string
	for i := 1; i <= 15; i++ {
		seeded = append(seeded, store.FormatStamp(now.Add(24*time.Hour+time.Duration(i)*time.Minute)))
	}
	seedSnapshots(t, remoteDir, seeded)

	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	if got := treeFiles(t, filepath.Join(remoteDir, "snapshots")); len(got) != 16 {
		t.Errorf("snapshots/ holds %d files, want all 16 (15 seeded + the run's own) to survive the skip", len(got))
	}
	if !strings.Contains(st.Phase, "prune skipped") {
		t.Errorf("progress %q does not name the skip", st.Phase)
	}
}

// TestPrune_FirstRun is the first-run row: a remote with no archive/
// and no snapshots/ before the run. The run succeeds, and since the
// prune has nothing to report, the succeeded record's final progress
// names the run finished rather than repeating the last phase
// ("pruning old snapshots") that backup.Job.Run reported before the
// prune ran.
func TestPrune_FirstRun(t *testing.T) {
	remote, _ := localRemote(t)
	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")
	if st.Phase != backup.FinishedMessage {
		t.Errorf("progress %q, want %q for a clean prune", st.Phase, backup.FinishedMessage)
	}
}

// TestPrune_LocalCopies is spec criterion 16: a local snapshots/ leaf
// holding pre-migrate-v1.db and pre-restore-<stamp>.db dated before
// the run, a pre-migrate-v2.db dated after its start, and an operator
// file notes.txt. After a successful run only pre-migrate-v2.db,
// notes.txt, and the run's own reference survive.
func TestPrune_LocalCopies(t *testing.T) {
	remote, _ := localRemote(t)
	dir := newStateDir(t)
	h := openWrite(t, dir)
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	snapDir := filepath.Join(dir, "store", "snapshots")
	old := time.Now().Add(-time.Hour)
	future := time.Now().Add(time.Hour)
	plantLocalCopy(t, snapDir, "pre-migrate-v1.db", old)
	plantLocalCopy(t, snapDir, "pre-restore-"+store.FormatStamp(time.Now().Add(-2*time.Hour))+".db", old)
	plantLocalCopy(t, snapDir, "pre-migrate-v2.db", future)
	plantLocalCopy(t, snapDir, "notes.txt", old)
	plantLocalCopy(t, snapDir, "other.db", old)

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	names := localSnapshotNames(t, dir)
	ref := backupRefName(names)
	if ref == "" {
		t.Fatalf("local snapshots/ holds %q, no backup reference", names)
	}
	want := []string{"notes.txt", "other.db", "pre-migrate-v2.db", ref}
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Errorf("local snapshots/ holds %q, want %q", names, want)
	}
}

// TestPrune_FailedRunPreservesEverything is the failed-run row,
// mirroring TestBackup_ZeroTracksKeepsRemote's shape: a remote seeded
// with 15 old snapshots and zero catalog tracks, with a root missing
// one remote file. The run fails, every seeded snapshot and local
// copy survives, and no reference is written.
func TestPrune_FailedRunPreservesEverything(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	var seeded []string
	for i := 1; i <= 15; i++ {
		seeded = append(seeded, store.FormatStamp(now.Add(-time.Duration(i)*time.Hour)))
	}
	seedSnapshots(t, remoteDir, seeded)
	writeTree(t, filepath.Join(remoteDir, "library"), "a/one.flac", "a/two.flac")

	dir := newStateDir(t)
	h := openWrite(t, dir)
	snapDir := filepath.Join(dir, "store", "snapshots")
	plantLocalCopy(t, snapDir, "pre-migrate-v1.db", now.Add(-time.Hour))

	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")
	mtime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, p := range []string{filepath.Join(lib, "a/one.flac"), filepath.Join(remoteDir, "library/a/one.flac")} {
		if err := os.Chtimes(p, mtime, mtime); err != nil {
			t.Fatal(err)
		}
	}

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobFailed, "rclone-exit-7")

	if got := treeFiles(t, filepath.Join(remoteDir, "snapshots")); len(got) != 15 {
		t.Errorf("snapshots/ holds %d files after a failed run, want all 15 seeded", len(got))
	}
	if names := localSnapshotNames(t, dir); !slices.Equal(names, []string{"pre-migrate-v1.db"}) {
		t.Errorf("local snapshots/ holds %q, want only the untouched pre-migrate-v1.db", names)
	}
}

// TestPrune_SendFailureNoReference is the send-failure row: a
// stand-in exiting 0 for sync and 1 for copyto. The run fails with
// rclone-exit-1, writes no reference, and runs no prune step.
func TestPrune_SendFailureNoReference(t *testing.T) {
	s := useStandin(t, standinConfig{exit: "copyto=1"})
	dir := newStateDir(t)
	h := openWrite(t, dir)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	_, st := runBackup(t, h, "name:dir", lib)
	wantState(t, st, ops.JobFailed, "rclone-exit-1")

	if ref := backupRefName(localSnapshotNames(t, dir)); ref != "" {
		t.Errorf("local snapshots/ holds a reference %q after a failed send, want none", ref)
	}

	var subs []string
	for _, argv := range s.invocations(t) {
		subs = append(subs, argv[0])
	}
	if want := []string{"sync", "copyto"}; !slices.Equal(subs, want) {
		t.Errorf("the stand-in ran %q, want %q; a send failure must run no prune step", subs, want)
	}
}

// TestPrune_FewSnapshotsKeepsAll is the states table's "14 or fewer,
// own included" row for remote snapshots: a remote seeded with 5
// snapshots stays well under the keep total, so a successful run
// deletes none of them.
func TestPrune_FewSnapshotsKeepsAll(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	var seeded []string
	for i := 1; i <= 5; i++ {
		seeded = append(seeded, store.FormatStamp(now.Add(-time.Duration(i)*time.Hour)))
	}
	seedSnapshots(t, remoteDir, seeded)

	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	got := treeFiles(t, filepath.Join(remoteDir, "snapshots"))
	if len(got) != len(seeded)+1 {
		t.Fatalf("snapshots/ holds %d files, want %d (%d seeded + the run's own): %q", len(got), len(seeded)+1, len(seeded), got)
	}
	for _, stamp := range seeded {
		if !slices.Contains(got, stamp+".db") {
			t.Errorf("snapshots/ lost seeded stamp %s, want every seeded snapshot kept when the total stays at or under the keep count", stamp)
		}
	}
}

// TestPrune_StepFailureLogged is the states table's "any prune step
// fails" row: a read-only archive directory makes rclone purge fail
// to delete the seeded marker inside it. The run still ends
// succeeded, and the failure is logged and folded into the record's
// final progress rather than failing the job.
func TestPrune_StepFailureLogged(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	oldStamp := store.FormatStamp(now.Add(-100 * 24 * time.Hour))
	seedArchiveDir(t, remoteDir, oldStamp)
	archiveDir := filepath.Join(remoteDir, "archive", oldStamp)
	if err := os.Chmod(archiveDir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chmod(archiveDir, 0o755); err != nil {
			t.Fatal(err)
		}
	})

	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")
	if !strings.Contains(st.Phase, "prune failed") {
		t.Errorf("progress %q does not name the failed prune step", st.Phase)
	}

	if _, err := os.Stat(filepath.Join(archiveDir, "marker.flac")); err != nil {
		t.Errorf("marker.flac under the read-only archive directory: %v, want it left in place since the purge could not remove it", err)
	}
}
