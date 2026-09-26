package backup

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// catalogRoot is the root path of the one collection plantCatalog
// creates, under which every track's file lies.
const catalogRoot = "main"

// catalogStore returns a handle, closed at cleanup, on a fresh store.
// With withUser it holds a member and a collection at catalogRoot,
// with one album holding a track at each of rels; without, it holds no
// row at all, since a collection needs an owner.
func catalogStore(t *testing.T, withUser bool, rels ...string) *store.Handle {
	t.Helper()
	h, err := store.OpenServe(ctxFor(t), t.TempDir(), store.ServeOptions{
		BusyTimeout:   testBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	if withUser {
		plantCatalog(t, h, rels...)
	}
	return h
}

// plantCatalog inserts into h's store a member, a collection at
// catalogRoot that the member owns, and an album holding a track at
// each of rels, each rel bound as the bytes it holds.
func plantCatalog(t *testing.T, h *store.Handle, rels ...string) {
	t.Helper()
	err := h.Update(ctxFor(t), func(ctx context.Context, tx *store.Tx) error {
		user, coll, album := store.NewID(), store.NewID(), store.NewID()
		if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, streamer_username, created_at, updated_at)
VALUES (?, 'owner@example.org', 'Owner', 'owner', 1, 1)`, user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (?, 'main', ?, ?, 'group', 'server')`, coll, []byte(catalogRoot), user); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO albums (id, collection_id, created_at, updated_at) VALUES (?, ?, 1, 1)`, album, coll); err != nil {
			return err
		}
		for _, rel := range rels {
			if _, err := tx.Exec(ctx, `INSERT INTO tracks (id, album_id, collection_id, rel_path, created_at, updated_at)
VALUES (?, ?, ?, ?, 1, 1)`, store.NewID(), album, coll, []byte(rel)); err != nil {
				return fmt.Errorf("track %q: %w", rel, err)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("plant the catalog: %v", err)
	}
}

// under returns each of rels under catalogRoot, as a report lists it.
func under(rels ...string) []string {
	var paths []string
	for _, r := range rels {
		paths = append(paths, catalogRoot+"/"+r)
	}
	return paths
}

// TestReconcile is the reconciliation's states, each with the entries
// its report must list. Every tree file sits under catalogRoot unless
// its row says otherwise.
func TestReconcile(t *testing.T) {
	const bad = "caf\xe9.flac" // Latin-1, not UTF-8
	tests := []struct {
		name   string
		tracks []string
		// tree lists the files to create, relative to the root.
		tree []string
		// arrange changes the root after the tree is written.
		arrange      func(t *testing.T, root string)
		absentRoot   bool
		wantUnfiled  []string
		wantUnlisted []string
		wantText     []string
	}{
		{
			name:     "no tracks, empty tree",
			wantText: []string{"tracks without a file: 0\nfiles without a track: 0\n"},
		},
		{
			name:         "no tracks, populated tree",
			tree:         under("a.flac", "b/c.flac"),
			wantUnlisted: under("a.flac", "b/c.flac"),
			wantText:     []string{"files without a track: 2\n", "files without a track:\n  main/a.flac\n  main/b/c.flac\n"},
		},
		{
			name:        "tracks, empty tree",
			tracks:      []string{"x/1.flac", "x/2.flac"},
			wantUnfiled: under("x/1.flac", "x/2.flac"),
			wantText:    []string{"tracks without a file: 2\n", "tracks without a file:\n  main/x/1.flac\n  main/x/2.flac\n"},
		},
		{
			name:         "a row without a file and a file without a row",
			tracks:       []string{"x/1.flac", "x/2.flac"},
			tree:         under("x/1.flac", "x/extra.flac"),
			wantUnfiled:  under("x/2.flac"),
			wantUnlisted: under("x/extra.flac"),
			wantText:     []string{"tracks without a file: 1\nfiles without a track: 1\n"},
		},
		{
			name:        "the root absent",
			tracks:      []string{"x/1.flac"},
			absentRoot:  true,
			wantUnfiled: under("x/1.flac"),
			wantText:    []string{" is absent: every track is listed as without a file\n"},
		},
		{
			name:   "a non-UTF-8 name that matches its row",
			tracks: []string{bad},
			tree:   under(bad),
		},
		{
			name:         "a non-UTF-8 name with no row",
			tree:         under(bad),
			wantUnlisted: under(bad),
			wantText:     []string{"  \"main/caf\\xe9.flac\"\n"},
		},
		{
			name:   "a track whose path is a symlink to a file",
			tracks: []string{"x/link.flac"},
			tree:   under("x/real.flac"),
			arrange: func(t *testing.T, root string) {
				if err := os.Symlink("real.flac", filepath.Join(root, catalogRoot, "x/link.flac")); err != nil {
					t.Fatal(err)
				}
			},
			wantUnfiled:  under("x/link.flac"),
			wantUnlisted: under("x/real.flac"),
		},
		{
			name:   "a symlink with no row",
			tracks: []string{"x/real.flac"},
			tree:   under("x/real.flac"),
			arrange: func(t *testing.T, root string) {
				if err := os.Symlink("real.flac", filepath.Join(root, catalogRoot, "x/stray.flac")); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:   "a track whose path is a directory",
			tracks: []string{"x"},
			tree:   under("x/1.flac"),
			// The directory is not a file for its row, and the file
			// inside it has no row.
			wantUnfiled:  under("x"),
			wantUnlisted: under("x/1.flac"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := catalogStore(t, true, tt.tracks...)
			root := t.TempDir()
			writeTree(t, root, tt.tree...)
			if tt.arrange != nil {
				tt.arrange(t, root)
			}
			if tt.absentRoot {
				root = filepath.Join(root, "gone")
			}
			rep, err := reconcile(ctxFor(t), h, root)
			if err != nil {
				t.Fatalf("reconcile(): %v", err)
			}
			if !slices.Equal(rep.unfiled, tt.wantUnfiled) {
				t.Errorf("tracks without a file %q, want %q", rep.unfiled, tt.wantUnfiled)
			}
			if !slices.Equal(rep.unlisted, tt.wantUnlisted) {
				t.Errorf("files without a track %q, want %q", rep.unlisted, tt.wantUnlisted)
			}
			text := string(rep.render(math.MaxInt))
			for _, want := range tt.wantText {
				if !strings.Contains(text, want) {
					t.Errorf("report text %q, want it to hold %q", text, want)
				}
			}
			if strings.Contains(text, "run member import") {
				t.Errorf("report over a store with a member says %q", text)
			}
		})
	}
}

// TestReconcile_WalkError: a subtree the walk cannot read is named in
// the report with its error, and the files beside it are still
// reconciled.
func TestReconcile_WalkError(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("this test needs a caller a mode-000 directory refuses; root reads it")
	}
	h := catalogStore(t, true)
	root := t.TempDir()
	writeTree(t, root, under("open/a.flac", "shut/b.flac")...)
	shut := filepath.Join(root, catalogRoot, "shut")
	if err := os.Chmod(shut, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(shut, 0o755) })

	rep, err := reconcile(ctxFor(t), h, root)
	if err != nil {
		t.Fatalf("reconcile(): %v", err)
	}
	if want := under("open/a.flac"); !slices.Equal(rep.unlisted, want) {
		t.Errorf("files without a track %q, want %q", rep.unlisted, want)
	}
	want := "unreadable subtree " + shut + ": open " + shut + ": permission denied\n"
	if text := string(rep.render(math.MaxInt)); !strings.Contains(text, want) {
		t.Errorf("report text %q, want it to hold %q", text, want)
	}
}

// TestReconcile_NoUsers is the report over a store whose users table is
// empty, as a snapshot from before the roster moved into the store
// leaves it: it says to run member import.
func TestReconcile_NoUsers(t *testing.T) {
	for _, tt := range []struct {
		name     string
		withUser bool
		want     bool
	}{
		{name: "empty users table", withUser: false, want: true},
		{name: "one member", withUser: true, want: false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := catalogStore(t, tt.withUser)
			rep, err := reconcile(ctxFor(t), h, t.TempDir())
			if err != nil {
				t.Fatalf("reconcile(): %v", err)
			}
			got := strings.Contains(string(rep.render(math.MaxInt)), "the users table is empty: run member import\n")
			if got != tt.want {
				t.Errorf("report %q names member import: %v, want %v", rep.render(math.MaxInt), got, tt.want)
			}
		})
	}
}

// TestReport_Renders: the file carries every entry, and the summary the
// counts and at most the first 20 of each list, then how many more.
func TestReport_Renders(t *testing.T) {
	var rels []string
	for i := range 23 {
		rels = append(rels, fmt.Sprintf("t%02d.flac", i))
	}
	h := catalogStore(t, true, rels...)
	rep, err := reconcile(ctxFor(t), h, t.TempDir())
	if err != nil {
		t.Fatalf("reconcile(): %v", err)
	}

	text, summary := string(rep.render(math.MaxInt)), string(rep.render(summaryLimit))
	for _, s := range []string{text, summary} {
		if !strings.Contains(s, "tracks without a file: 23\nfiles without a track: 0\n") {
			t.Errorf("render %q lacks the counts", s)
		}
	}
	if got := strings.Count(text, "\n  main/"); got != 23 {
		t.Errorf("report file lists %d entries, want 23", got)
	}
	if strings.Contains(text, "more") {
		t.Errorf("report file %q elides entries", text)
	}
	if got := strings.Count(summary, "\n  main/"); got != 20 {
		t.Errorf("summary lists %d entries, want 20", got)
	}
	if !strings.Contains(summary, "  main/t19.flac\n  and 3 more\n") {
		t.Errorf("summary %q, want it to end its list at t19 and count 3 more", summary)
	}
}
