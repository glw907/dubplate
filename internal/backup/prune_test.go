package backup_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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
// the run, a pre-migrate-v2.db dated after its start, and operator
// files. After a successful run pre-migrate-v1.db alone is gone: a
// pre-restore copy holds the rows its restore did not bring back, which
// no backup captured, so no prune removes it.
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
	preRestore := "pre-restore-" + store.FormatStamp(time.Now().Add(-2*time.Hour)) + ".db"
	plantLocalCopy(t, snapDir, preRestore, old)
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
	want := []string{"notes.txt", "other.db", "pre-migrate-v2.db", preRestore, ref}
	slices.Sort(want)
	if !slices.Equal(names, want) {
		t.Errorf("local snapshots/ holds %q, want %q", names, want)
	}
}

// TestPrune_LocalCopyNames: prune removes a local file exactly when
// the store names it a pre-migration copy, so a name the store would
// never write, such as a zero-padded version, is left alone.
func TestPrune_LocalCopyNames(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{name: "pre-migrate-v1.db", want: true},
		{name: "pre-migrate-v12.db", want: true},
		{name: "pre-migrate-v01.db", want: false},
		{name: "pre-migrate-v+1.db", want: false},
		{name: "pre-migrate-v0.db", want: false},
		{name: "pre-migrate-v-1.db", want: false},
		{name: "pre-migrate-v.db", want: false},
		{name: "pre-migrate-v1.db.tmp", want: false},
		{name: "pre-restore-20260101T000000Z.db", want: false},
	}
	remote, _ := localRemote(t)
	dir := newStateDir(t)
	h := openWrite(t, dir)
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")
	snapDir := filepath.Join(dir, "store", "snapshots")
	old := time.Now().Add(-time.Hour)
	for _, tt := range tests {
		plantLocalCopy(t, snapDir, tt.name, old)
	}

	_, st := runBackup(t, h, remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	names := localSnapshotNames(t, dir)
	for _, tt := range tests {
		if got := store.IsPreMigrateCopy(tt.name); got != tt.want {
			t.Errorf("store.IsPreMigrateCopy(%q) = %v, want %v", tt.name, got, tt.want)
		}
		if removed := !slices.Contains(names, tt.name); removed != tt.want {
			t.Errorf("prune removed %s: %v, want %v", tt.name, removed, tt.want)
		}
	}
}

// TestPrune_KeepsPreRestoreCopy is a real restore followed by a backup
// that starts after the pre-restore copy's mtime: the backup prunes an
// older pre-migration copy but keeps the pre-restore copy, and the
// restore told the operator where the copy is and that only they
// remove it.
func TestPrune_KeepsPreRestoreCopy(t *testing.T) {
	remote, _ := localRemote(t)
	src := newStateDir(t)
	hs := openWrite(t, src)
	file, err := hs.BackupCopy(deadline(t), slog.New(slog.DiscardHandler), "0123456789abcdef0123456789abcdef")
	if err != nil {
		t.Fatalf("BackupCopy: %v", err)
	}
	dir := newStateDir(t)
	var out bytes.Buffer
	err = backup.Restore(deadline(t), backup.RestoreConfig{
		StateDir:    dir,
		BusyTimeout: testBusy,
		Now:         time.Now(),
		FromFile:    file,
		Out:         &out,
		Logger:      slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("Restore(): %v", err)
	}
	var preRestore string
	for _, n := range localSnapshotNames(t, dir) {
		if strings.HasPrefix(n, "pre-restore-") {
			preRestore = n
		}
	}
	if preRestore == "" {
		t.Fatalf("local snapshots/ holds %q, no pre-restore copy", localSnapshotNames(t, dir))
	}
	snapDir := filepath.Join(dir, "store", "snapshots")
	wantOut := "the replaced store is kept at " + filepath.Join(snapDir, preRestore) + " until you remove it by hand"
	if !strings.Contains(out.String(), wantOut) {
		t.Errorf("Restore printed %q, want it to hold %q", out.String(), wantOut)
	}

	old := time.Now().Add(-time.Hour)
	if err := os.Chtimes(filepath.Join(snapDir, preRestore), old, old); err != nil {
		t.Fatal(err)
	}
	plantLocalCopy(t, snapDir, "pre-migrate-v1.db", old)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")
	_, st := runBackup(t, openWrite(t, dir), remote, lib)
	wantState(t, st, ops.JobSucceeded, "")

	names := localSnapshotNames(t, dir)
	if !slices.Contains(names, preRestore) {
		t.Errorf("local snapshots/ holds %q after the backup, want the pre-restore copy %s kept", names, preRestore)
	}
	if slices.Contains(names, "pre-migrate-v1.db") {
		t.Errorf("local snapshots/ holds %q after the backup, want pre-migrate-v1.db pruned", names)
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

// TestPrune_SnapshotKeepReader proves the remote keep consumer: a
// snapshot keep of 3, over five prior snapshots plus the run's own,
// keeps only the run's own and the two newest of the five. The
// default keep of 14 would have kept all six, so this fails unless
// Job.Run honors the injected reader over the retired package
// constant.
func TestPrune_SnapshotKeepReader(t *testing.T) {
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

	keep := func(context.Context) (int, error) { return 3, nil }
	_, st := runBackupWithBackup(t, h, remote, lib, keep, defaultGuardFraction, defaultArchiveRetention)
	wantState(t, st, ops.JobSucceeded, "")

	got := treeFiles(t, filepath.Join(remoteDir, "snapshots"))
	if len(got) != 3 {
		t.Fatalf("snapshots/ holds %d files, want 3 (the run's own and the 2 newest of 5 seeded): %q", len(got), got)
	}
	for _, stamp := range seeded[:2] {
		if !slices.Contains(got, stamp+".db") {
			t.Errorf("snapshots/ lost the newest seeded stamp %s, want it kept under a keep of 3", stamp)
		}
	}
	for _, stamp := range seeded[2:] {
		if slices.Contains(got, stamp+".db") {
			t.Errorf("snapshots/ kept older stamp %s, want it pruned under a keep of 3", stamp)
		}
	}
}

// TestPrune_ArchiveRetentionReader proves the archive retention
// consumer: an archive directory 200h past its move is purged under a
// configured retention of 168h. The default retention of 2160h would
// have kept it, so this fails unless Job.Run honors the injected
// reader over the retired package constant.
func TestPrune_ArchiveRetentionReader(t *testing.T) {
	remote, remoteDir := localRemote(t)
	now := time.Now()
	oldStamp := store.FormatStamp(now.Add(-200 * time.Hour))
	seedArchiveDir(t, remoteDir, oldStamp)

	h := openWrite(t, newStateDir(t))
	plantTracks(t, h, 1)
	lib := t.TempDir()
	writeTree(t, lib, "a/one.flac")

	retention := func(context.Context) (time.Duration, error) { return 168 * time.Hour, nil }
	_, st := runBackupWithBackup(t, h, remote, lib, defaultSnapshotKeep, defaultGuardFraction, retention)
	wantState(t, st, ops.JobSucceeded, "")

	if got := archiveDirNames(t, remoteDir); slices.Contains(got, oldStamp) {
		t.Errorf("archive/ kept %s, want it purged under an archive retention of 168h", oldStamp)
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
