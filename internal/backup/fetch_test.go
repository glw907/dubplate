package backup

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// Three stamps in the store's layout, oldest first, and one no
// snapshot on any fixture remote carries.
const (
	stampOld    = "20260101T000000.000000Z"
	stampMid    = "20260102T000000.000000Z"
	stampNew    = "20260103T000000.000000Z"
	stampAbsent = "20251231T000000.000000Z"
)

// A fetchStandin is the stand-in rclone standin_test.go's TestMain runs
// whenever the test binary is executed under that name, as this
// package's own tests see it: the environment it reads and the record
// it appends to.
type fetchStandin struct {
	record string
	pid    string
}

// useFetchStandin puts the stand-in alone on PATH, exiting per exit
// ("copyto=1"), sleeping in the one subcommand sleep names, and
// exiting 130 on a SIGINT, and returns its rclone path and its files.
// The environment names are the ones standin_test.go declares.
func useFetchStandin(t *testing.T, exit, sleep string) (string, fetchStandin) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("test binary: %v", err)
	}
	bin := filepath.Join(t.TempDir(), "rclone")
	if err := os.Symlink(exe, bin); err != nil {
		t.Fatalf("link the stand-in: %v", err)
	}
	files := t.TempDir()
	s := fetchStandin{record: filepath.Join(files, "record"), pid: filepath.Join(files, "pid")}
	t.Setenv("DUBPLATE_STANDIN_RECORD", s.record)
	t.Setenv("DUBPLATE_STANDIN_PID", s.pid)
	t.Setenv("DUBPLATE_STANDIN_EXIT", exit)
	t.Setenv("DUBPLATE_STANDIN_SLEEP", sleep)
	t.Setenv("DUBPLATE_STANDIN_SIGINT", "honor")
	t.Setenv("DUBPLATE_STANDIN_LOCK_STAGE", "")
	return bin, s
}

// lines returns every line the stand-in recorded, none when it never
// ran.
func (s fetchStandin) lines(t *testing.T) []string {
	t.Helper()
	data, err := os.ReadFile(s.record)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatalf("read stand-in record: %v", err)
	}
	return strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
}

// subcommands returns the subcommand of each run the stand-in
// recorded, in order.
func (s fetchStandin) subcommands(t *testing.T) []string {
	t.Helper()
	var subs []string
	for _, line := range s.lines(t) {
		argv, ok := strings.CutPrefix(line, `argv ["`)
		if !ok {
			continue
		}
		sub, _, _ := strings.Cut(argv, `"`)
		subs = append(subs, sub)
	}
	return subs
}

// cancelOnStart cancels once the stand-in has written its pid, which
// it does after installing its SIGINT handler, and kills it at
// cleanup whether or not it still runs.
func (s fetchStandin) cancelOnStart(t *testing.T, cancel context.CancelFunc) {
	t.Helper()
	t.Cleanup(func() {
		if data, err := os.ReadFile(s.pid); err == nil {
			if pid, err := strconv.Atoi(string(data)); err == nil {
				_ = syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	})
	go func() {
		defer cancel()
		for end := time.Now().Add(20 * time.Second); time.Now().Before(end); time.Sleep(10 * time.Millisecond) {
			if _, err := os.Stat(s.pid); err == nil {
				return
			}
		}
	}()
}

// rcloneRemote configures a local-backend rclone remote named
// fetchtest over a fresh directory and returns rclone's path, the
// remote over that directory, and the directory.
func rcloneRemote(t *testing.T) (string, Remote, string) {
	t.Helper()
	bin, err := exec.LookPath("rclone")
	if err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}
	conf := filepath.Join(t.TempDir(), "rclone.conf")
	if err := os.WriteFile(conf, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("RCLONE_CONFIG", conf)
	t.Setenv("RCLONE_CONFIG_FETCHTEST_TYPE", "local")
	dir := t.TempDir()
	remote, err := ParseRemote("fetchtest:" + dir)
	if err != nil {
		t.Fatal(err)
	}
	return bin, remote, dir
}

// seedRemoteSnapshots writes body under dir's snapshots/ as
// <stamp>.db for each stamp, and as each name in extra.
func seedRemoteSnapshots(t *testing.T, dir, body string, stamps []string, extra ...string) {
	t.Helper()
	snaps := filepath.Join(dir, "snapshots")
	if err := os.MkdirAll(snaps, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range slices.Concat(stampNames(stamps), extra) {
		if err := os.WriteFile(filepath.Join(snaps, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func stampNames(stamps []string) []string {
	var names []string
	for _, s := range stamps {
		names = append(names, s+".db")
	}
	return names
}

// openSession opens a new restore session over a fresh state
// directory, closed at cleanup.
func openSession(t *testing.T) *store.RestoreSession {
	t.Helper()
	sess, err := store.OpenRestore(t.TempDir(), store.RestoreOptions{
		BusyTimeout: 5 * time.Second,
		Now:         time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC),
		Logger:      slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenRestore: %v", err)
	}
	t.Cleanup(func() {
		if err := sess.Close(); err != nil {
			t.Errorf("close session: %v", err)
		}
	})
	return sess
}

// withUmask sets the process umask to mask for the test.
func withUmask(t *testing.T, mask int) {
	t.Helper()
	old := syscall.Umask(mask)
	t.Cleanup(func() { syscall.Umask(old) })
}

// TestPickSnapshot pins the listing's rules without a remote: latest
// is the greatest stamp that parses, whatever order the listing gives
// and whatever else it holds; an absent stamp names its remote path;
// an older stamp needs --older, naming the latest; and the chosen
// object's size is the listing's.
func TestPickSnapshot(t *testing.T) {
	const dir = "name:backups/snapshots"
	outOfOrder := []remoteEntry{
		{Name: stampMid + ".db", Size: 20},
		{Name: stampNew + ".db", Size: 30},
		{Name: "99999999T999999.999999Z.db", Size: 90},
		{Name: stampOld + ".db", Size: 10},
		{Name: "notes.db", Size: 40},
		{Name: "20270101T000000.000000Z", IsDir: true, Size: -1},
	}
	tests := []struct {
		name     string
		entries  []remoteEntry
		want     string
		older    bool
		wantSnap snapshot
		wantErr  []string
	}{
		{
			name:     "latest out of order, beside unparseable names",
			entries:  outOfOrder,
			want:     latestSnapshot,
			wantSnap: snapshot{stamp: stampNew, size: 30},
		},
		{
			name:     "the latest stamp by name",
			entries:  outOfOrder,
			want:     stampNew,
			wantSnap: snapshot{stamp: stampNew, size: 30},
		},
		{
			name:    "an older stamp without --older",
			entries: outOfOrder,
			want:    stampMid,
			wantErr: []string{stampMid, "is not the latest snapshot " + stampNew, "--older"},
		},
		{
			name:     "an older stamp with --older",
			entries:  outOfOrder,
			want:     stampOld,
			older:    true,
			wantSnap: snapshot{stamp: stampOld, size: 10},
		},
		{
			name:    "an absent stamp",
			entries: outOfOrder,
			want:    stampAbsent,
			wantErr: []string{"no snapshot name:backups/snapshots/" + stampAbsent + ".db"},
		},
		{
			name:    "latest on an empty listing",
			want:    latestSnapshot,
			wantErr: []string{"no snapshot under name:backups/snapshots"},
		},
		{
			name:     "a size the backend does not report",
			entries:  []remoteEntry{{Name: stampNew + ".db", Size: -1}},
			want:     latestSnapshot,
			wantSnap: snapshot{stamp: stampNew, size: -1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pickSnapshot(tt.entries, dir, tt.want, tt.older)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("pickSnapshot() = %+v, want a refusal", got)
				}
				for _, sub := range tt.wantErr {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("refusal %q does not name %q", err, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("pickSnapshot(): %v", err)
			}
			if got != tt.wantSnap {
				t.Errorf("pickSnapshot() = %+v, want %+v", got, tt.wantSnap)
			}
		})
	}
}

// TestResolve_RealRemote is the remote half of spec criterion 9
// against a real rclone local remote seeded with three snapshots and
// an unparseable notes.db: an absent stamp names its remote path,
// latest on an empty remote names snapshots/, and an older stamp
// without --older names the latest. A stamp that does not parse is
// refused before any rclone run, so its row points the transport at
// a binary that does not exist.
func TestResolve_RealRemote(t *testing.T) {
	bin, remote, dir := rcloneRemote(t)
	seedRemoteSnapshots(t, dir, "12345", []string{stampNew, stampOld, stampMid}, "notes.db")
	_, emptyRemote, _ := rcloneRemote(t)
	snaps := "fetchtest:" + dir + "/snapshots"
	tests := []struct {
		name     string
		bin      string
		remote   Remote
		want     string
		older    bool
		wantSnap snapshot
		wantErr  []string
	}{
		{name: "latest", want: latestSnapshot, wantSnap: snapshot{stamp: stampNew, size: 5}},
		{name: "the latest by name", want: stampNew, wantSnap: snapshot{stamp: stampNew, size: 5}},
		{name: "older with --older", want: stampOld, older: true, wantSnap: snapshot{stamp: stampOld, size: 5}},
		{name: "older without --older", want: stampOld, wantErr: []string{"latest snapshot " + stampNew, "--older"}},
		{name: "an absent stamp", want: stampAbsent, wantErr: []string{"no snapshot " + snaps + "/" + stampAbsent + ".db"}},
		{name: "latest on an empty remote", remote: emptyRemote, want: latestSnapshot, wantErr: []string{"no snapshot under " + emptyRemote.join("snapshots")}},
		{name: "a stamp that does not parse", bin: "/nonexistent/rclone", want: "yesterday", wantErr: []string{`"yesterday"`, "20060102T150405.000000Z"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := source{t: transport{bin: cmp.Or(tt.bin, bin)}, remote: remote}
			if tt.remote != (Remote{}) {
				src.remote = tt.remote
			}
			got, err := src.resolve(context.Background(), tt.want, tt.older)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("resolve() = %+v, want a refusal", got)
				}
				for _, sub := range tt.wantErr {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("refusal %q does not name %q", err, sub)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("resolve(): %v", err)
			}
			if got != tt.wantSnap {
				t.Errorf("resolve() = %+v, want %+v", got, tt.wantSnap)
			}
		})
	}
}

// TestResolve_AbsentStampRunsNoCopyto pins that an absent stamp is
// decided from the listing alone: the stand-in lists an absent
// snapshots/ (lsjson exits 3) and records no run but that one.
func TestResolve_AbsentStampRunsNoCopyto(t *testing.T) {
	bin, s := useFetchStandin(t, "lsjson=3", "")
	remote, err := ParseRemote("name:dir")
	if err != nil {
		t.Fatal(err)
	}
	src := source{t: transport{bin: bin}, remote: remote}
	if _, err := src.resolve(context.Background(), stampAbsent, false); err == nil {
		t.Fatal("resolve() of an absent stamp succeeded, want a refusal")
	}
	if got := s.subcommands(t); !slices.Equal(got, []string{"lsjson"}) {
		t.Errorf("the stand-in ran %q, want only the one lsjson", got)
	}
}

// TestResolve_EmptyPathRemote is the empty-path remote row: name:
// reads the backend's root, which for a local remote is the working
// directory, so its snapshots/ is name:snapshots and never
// name:/snapshots.
func TestResolve_EmptyPathRemote(t *testing.T) {
	bin, _, dir := rcloneRemote(t)
	seedRemoteSnapshots(t, dir, "x", []string{stampOld, stampNew})
	t.Chdir(dir)
	remote, err := ParseRemote("fetchtest:")
	if err != nil {
		t.Fatal(err)
	}
	got, err := source{t: transport{bin: bin}, remote: remote}.resolve(context.Background(), latestSnapshot, false)
	if err != nil {
		t.Fatalf("resolve() on name:: %v", err)
	}
	if got.stamp != stampNew {
		t.Errorf("resolve() on name: = %+v, want stamp %s", got, stampNew)
	}
}

// TestList_Canceled is the canceled listing row: a SIGINT-honoring
// stand-in sleeping in lsjson, canceled once it runs, records one
// SIGINT to its own process, and the listing's error wraps the
// context's.
func TestList_Canceled(t *testing.T) {
	bin, s := useFetchStandin(t, "", "lsjson")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	s.cancelOnStart(t, cancel)

	_, err := transport{bin: bin}.list(ctx, "name:dir/snapshots", false)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("list() = %v, want an error wrapping context.Canceled", err)
	}
	n := 0
	for _, line := range s.lines(t) {
		if line == "SIGINT" {
			n++
		}
	}
	if n != 1 {
		t.Errorf("the stand-in recorded %d SIGINTs, want 1: %q", n, s.lines(t))
	}
}

// TestFetch_RealRemote is the fetch through a real rclone under umask
// 022: copyto renames a new file into place at 0644 there (measured),
// so the stage reads 0600 only because the fetch chmods it, and it
// holds the remote object's bytes. The transport inherits the
// session's lock, as a restore's does.
func TestFetch_RealRemote(t *testing.T) {
	withUmask(t, 0o022)
	bin, remote, dir := rcloneRemote(t)
	seedRemoteSnapshots(t, dir, "snapshot bytes", []string{stampNew})
	sess := openSession(t)
	src := source{t: transport{bin: bin, inherit: sess.DupFiles}, remote: remote}

	snap, err := src.resolve(context.Background(), latestSnapshot, false)
	if err != nil {
		t.Fatalf("resolve(): %v", err)
	}
	var out bytes.Buffer
	if err := src.fetch(context.Background(), sess, &out, snap); err != nil {
		t.Fatalf("fetch(): %v", err)
	}
	info, err := os.Stat(sess.StagePath())
	if err != nil {
		t.Fatalf("stat stage: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Errorf("stage mode = %v, want 0600", got)
	}
	data, err := os.ReadFile(sess.StagePath())
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "snapshot bytes" {
		t.Errorf("stage holds %q, want the remote object's bytes", data)
	}
	if out.Len() != 0 {
		t.Errorf("fetch printed %q for an object with a size, want nothing", out.String())
	}
}

// TestFetch_SizeChecks is the size rows: an object whose size is -1
// skips the store's free-space check with a printed line naming it,
// and is fetched, where the check itself refuses a size of -1; an
// object larger than any filesystem is refused by the check before
// any rclone run.
func TestFetch_SizeChecks(t *testing.T) {
	bin, remote, dir := rcloneRemote(t)
	seedRemoteSnapshots(t, dir, "x", []string{stampNew})
	object := remote.join("snapshots/" + stampNew + ".db")

	t.Run("size -1", func(t *testing.T) {
		sess := openSession(t)
		src := source{t: transport{bin: bin, inherit: sess.DupFiles}, remote: remote}
		var out bytes.Buffer
		if err := src.fetch(context.Background(), sess, &out, snapshot{stamp: stampNew, size: -1}); err != nil {
			t.Fatalf("fetch(): %v", err)
		}
		want := "skipping the store free-space check: " + object + " reports no size\n"
		if out.String() != want {
			t.Errorf("fetch printed %q, want %q", out.String(), want)
		}
		if _, err := os.Stat(sess.StagePath()); err != nil {
			t.Errorf("stage after a fetch of a sizeless object: %v", err)
		}
	})

	t.Run("too large", func(t *testing.T) {
		sess := openSession(t)
		src := source{t: transport{bin: "/nonexistent/rclone"}, remote: remote}
		err := src.fetch(context.Background(), sess, &bytes.Buffer{}, snapshot{stamp: stampNew, size: 1 << 60})
		if err == nil || !strings.Contains(err.Error(), "fetch "+object) || strings.Contains(err.Error(), "nonexistent") {
			t.Errorf("fetch() = %v, want the free-space refusal naming %s and no rclone run", err, object)
		}
	})
}

// TestFetch_FailureRemovesStage is the failed and canceled fetch rows:
// a stand-in copyto that exits 1, or that sleeps until the fetch is
// canceled, beside a stage planted as a transfer's leftover. The
// fetch names rclone's exit and last stderr line, or wraps the
// cancel, and removes the stage either way.
func TestFetch_FailureRemovesStage(t *testing.T) {
	tests := []struct {
		name       string
		exit       string
		sleep      string
		cancel     bool
		wantErr    []string
		wantCanc   bool
		wantSIGINT int
	}{
		{
			name:    "rclone exits 1",
			exit:    "copyto=1",
			wantErr: []string{"fetch name:dir/snapshots/" + stampNew + ".db", "rclone-exit-1", "stand-in copyto exits 1"},
		},
		{
			name:       "canceled mid-transfer",
			sleep:      "copyto",
			cancel:     true,
			wantCanc:   true,
			wantSIGINT: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bin, s := useFetchStandin(t, tt.exit, tt.sleep)
			sess := openSession(t)
			remote, err := ParseRemote("name:dir")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(sess.StagePath(), []byte("partial"), 0o644); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if tt.cancel {
				s.cancelOnStart(t, cancel)
			}

			src := source{t: transport{bin: bin, inherit: sess.DupFiles}, remote: remote}
			err = src.fetch(ctx, sess, &bytes.Buffer{}, snapshot{stamp: stampNew, size: 7})
			if err == nil {
				t.Fatal("fetch() succeeded, want a failure")
			}
			for _, sub := range tt.wantErr {
				if !strings.Contains(err.Error(), sub) {
					t.Errorf("fetch() = %q, does not name %q", err, sub)
				}
			}
			if got := errors.Is(err, context.Canceled); got != tt.wantCanc {
				t.Errorf("fetch() = %v, canceled %v, want %v", err, got, tt.wantCanc)
			}
			if _, err := os.Lstat(sess.StagePath()); !errors.Is(err, os.ErrNotExist) {
				t.Errorf("stage after a failed fetch: %v, want it removed", err)
			}
			n := 0
			for _, line := range s.lines(t) {
				if line == "SIGINT" {
					n++
				}
			}
			if n != tt.wantSIGINT {
				t.Errorf("the stand-in recorded %d SIGINTs, want %d", n, tt.wantSIGINT)
			}
		})
	}
}

// TestParseTreeSize pins rclone size --json's shape, as measured on
// v1.74.3: the byte floor and the count of objects with no size.
func TestParseTreeSize(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    treeSize
		wantErr bool
	}{
		{name: "measured", in: `{"count":2,"bytes":20004096,"sizeless":0}`, want: treeSize{Count: 2, Bytes: 20004096, Sizeless: 0}},
		{name: "sizeless objects", in: `{"count":5,"bytes":100,"sizeless":3}`, want: treeSize{Count: 5, Bytes: 100, Sizeless: 3}},
		{name: "not JSON", in: `2 files`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTreeSize([]byte(tt.in))
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseTreeSize(%s) error = %v, want error %v", tt.in, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("parseTreeSize(%s) = %+v, want %+v", tt.in, got, tt.want)
			}
		})
	}
}

// TestLibrarySize_RealRemote measures a real local remote's library/
// of two files, 20000000 and 4096 bytes, as rclone size --json does.
func TestLibrarySize_RealRemote(t *testing.T) {
	bin, remote, dir := rcloneRemote(t)
	lib := filepath.Join(dir, "library", "a")
	if err := os.MkdirAll(lib, 0o755); err != nil {
		t.Fatal(err)
	}
	for name, size := range map[string]int64{"one.flac": 20000000, "two.flac": 4096} {
		f, err := os.Create(filepath.Join(lib, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := f.Truncate(size); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	}
	got, err := source{t: transport{bin: bin}, remote: remote}.librarySize(context.Background())
	if err != nil {
		t.Fatalf("librarySize(): %v", err)
	}
	if want := (treeSize{Count: 2, Bytes: 20004096}); got != want {
		t.Errorf("librarySize() = %+v, want %+v", got, want)
	}
}

// TestRemoteWarning pins the wrong-archive warning: none when the
// configured remote is the one the backup reference names, as rclone
// reads each, or when no reference names one; otherwise a line naming
// both.
func TestRemoteWarning(t *testing.T) {
	configured, err := ParseRemote("b2:family/backups")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name       string
		referenced string
		want       string
	}{
		{name: "the same remote", referenced: "b2:family/backups"},
		{name: "the same remote with a trailing slash", referenced: "b2:family/backups/"},
		{name: "no reference", referenced: ""},
		{
			name:       "another path",
			referenced: "b2:family/old",
			want:       "warning: restoring from b2:family/backups, but the backup reference names b2:family/old",
		},
		{
			name:       "a reference that does not parse",
			referenced: "no-colon",
			want:       "warning: restoring from b2:family/backups, but the backup reference names no-colon",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remoteWarning(configured, tt.referenced); got != tt.want {
				t.Errorf("remoteWarning(%q) = %q, want %q", tt.referenced, got, tt.want)
			}
		})
	}
}
