package intake

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/ops"
)

// openTestStore returns a migrated store under a fresh state directory.
func openTestStore(t *testing.T) *store.Handle {
	t.Helper()
	h, err := store.OpenServe(t.Context(), t.TempDir(), store.ServeOptions{
		BusyTimeout:   5 * time.Second,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// userSpec is one roster row; createdAt is in microseconds, and
// tombstoned is true for a removed member.
type userSpec struct {
	admin      bool
	tombstoned bool
	createdAt  int64
}

func insertUser(t *testing.T, h *store.Handle, u userSpec) []byte {
	t.Helper()
	id := store.NewID()
	var tombstonedAt any
	if u.tombstoned {
		tombstonedAt = u.createdAt + 1
	}
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
			VALUES (?, ?, 'Member', ?, ?, ?, ?, ?)`,
			id, hex.EncodeToString(id)+"@example.invalid", hex.EncodeToString(id), u.admin, tombstonedAt, u.createdAt, u.createdAt)
		return err
	})
	if err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

// routeCollectionOnce runs RouteCollection in its own Update.
func routeCollectionOnce(t *testing.T, h *store.Handle, log *slog.Logger, route Route) (id []byte, err error) {
	t.Helper()
	err = h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		var err error
		id, err = RouteCollection(ctx, tx, log, route)
		return err
	})
	return id, err
}

type collectionRow struct {
	name, visibility, writer, layout, route string
	owner, root                             []byte
}

func readCollection(t *testing.T, h *store.Handle, id []byte) collectionRow {
	t.Helper()
	var c collectionRow
	err := h.QueryRow(deadline(t),
		`SELECT name, owner, root_path, visibility, writer, layout_template, intake_route FROM collections WHERE id = ?`, id).
		Scan(&c.name, &c.owner, &c.root, &c.visibility, &c.writer, &c.layout, &c.route)
	if err != nil {
		t.Fatalf("read collection %x: %v", id, err)
	}
	return c
}

func countRoute(t *testing.T, h *store.Handle, route string) int {
	t.Helper()
	var n int
	if err := h.QueryRow(deadline(t), `SELECT count(*) FROM collections WHERE intake_route = ?`, route).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestRouteCollection(t *testing.T) {
	log := slog.New(slog.DiscardHandler)

	// The roster: a tombstoned admin created before both live admins, a
	// plain member created before every admin, and two live admins with
	// distinct creation times, the later inserted first.
	h := openTestStore(t)
	insertUser(t, h, userSpec{admin: false, createdAt: 50})
	tombstoned := insertUser(t, h, userSpec{admin: true, tombstoned: true, createdAt: 100})
	late := insertUser(t, h, userSpec{admin: true, createdAt: 300})
	early := insertUser(t, h, userSpec{admin: true, createdAt: 200})

	libraryID, err := routeCollectionOnce(t, h, log, RouteDefault)
	if err != nil {
		t.Fatalf("first default: %v", err)
	}
	lib := readCollection(t, h, libraryID)
	if lib.name != "Library" || lib.route != "default" {
		t.Errorf("default collection = %q on route %q, want Library on default", lib.name, lib.route)
	}
	if !bytes.Equal(lib.owner, early) {
		t.Errorf("owner = %x, want the earliest live admin %x (not the tombstoned %x or the later %x)", lib.owner, early, tombstoned, late)
	}
	if lib.writer != "server" || lib.visibility != "group" || lib.layout != "" {
		t.Errorf("writer %q, visibility %q, layout %q; want server, group, empty", lib.writer, lib.visibility, lib.layout)
	}
	if want := hex.EncodeToString(libraryID); string(lib.root) != want {
		t.Errorf("root = %q, want the lowercase hex of the id %q", lib.root, want)
	}

	again, err := routeCollectionOnce(t, h, log, RouteDefault)
	if err != nil {
		t.Fatalf("second default: %v", err)
	}
	if !bytes.Equal(again, libraryID) {
		t.Errorf("second call returned %x, want the existing %x", again, libraryID)
	}
	if n := countRoute(t, h, "default"); n != 1 {
		t.Errorf("%d default collections, want 1", n)
	}

	unlistedID, err := routeCollectionOnce(t, h, log, RouteUnlisted)
	if err != nil {
		t.Fatalf("unlisted: %v", err)
	}
	un := readCollection(t, h, unlistedID)
	if un.name != "Unlisted music" || un.route != "unlisted" || !bytes.Equal(un.owner, early) {
		t.Errorf("unlisted = %q on %q owned by %x, want Unlisted music on unlisted owned by %x", un.name, un.route, un.owner, early)
	}
	if un.writer != "server" || un.visibility != "group" {
		t.Errorf("unlisted writer %q, visibility %q; want server, group", un.writer, un.visibility)
	}
	if bytes.Equal(unlistedID, libraryID) {
		t.Error("the unlisted route returned the default collection")
	}
	if n := countRoute(t, h, "unlisted"); n != 1 {
		t.Errorf("%d unlisted collections, want 1", n)
	}
}

func TestRouteCollectionWithNoLiveAdmin(t *testing.T) {
	var logged bytes.Buffer
	log := slog.New(slog.NewTextHandler(&logged, nil))

	// A tombstoned admin and a plain member leave the roster with no
	// live admin.
	h := openTestStore(t)
	insertUser(t, h, userSpec{admin: true, tombstoned: true, createdAt: 100})
	insertUser(t, h, userSpec{admin: false, createdAt: 200})

	_, err := routeCollectionOnce(t, h, log, RouteDefault)
	var oe *ops.Error
	if !errors.As(err, &oe) || oe.Code != ops.CodeInternal {
		t.Fatalf("error = %v, want an ops internal error", err)
	}
	if !strings.Contains(logged.String(), "no live admin") {
		t.Errorf("log = %q, want a line naming the missing admin", logged.String())
	}
	if n := countRoute(t, h, "default"); n != 0 {
		t.Errorf("%d default collections after the refusal, want 0", n)
	}
}

func TestCheckNoNesting(t *testing.T) {
	h := openTestStore(t)
	admin := insertUser(t, h, userSpec{admin: true, createdAt: 1})
	existing := []string{"a", "x/y"}
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		for _, root := range existing {
			if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer)
				VALUES (?, ?, ?, ?, 'group', 'server')`, store.NewID(), root, []byte(root), admin); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed roots: %v", err)
	}

	tests := []struct {
		name string
		root string
		want bool // true when the root is refused as nested
	}{
		{"inside an existing root", "a/b", true},
		{"the same root", "a", true},
		{"a parent of an existing root", "x", true},
		{"a sibling sharing a name prefix", "ab", false},
		{"a root beside both", "c", false},
		{"a sibling of a nested root", "x/z", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
				return checkNoNesting(ctx, tx, []byte(tt.root))
			})
			if got := errors.Is(err, errNestedRoot); got != tt.want {
				t.Errorf("checkNoNesting(%q) = %v, want nested %t", tt.root, err, tt.want)
			}
			if !tt.want && err != nil {
				t.Errorf("checkNoNesting(%q) = %v, want nil", tt.root, err)
			}
		})
	}
}

// TestRelPathRefusals runs each path through the validator and through
// a direct insert into submission_files.rel_path. The two agree on
// every row but the backslash, which SQL admits and the validator
// refuses.
func TestRelPathRefusals(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantValid  bool
		wantInsert bool
	}{
		{"a plain relative path", "album/01 Track.flac", true, true},
		{"absolute", "/album/01.flac", false, false},
		{"empty component", "a//b", false, false},
		{"dot component", "a/./b", false, false},
		{"dotdot component", "a/../b", false, false},
		{"dotdot", "..", false, false},
		{"NUL", "a\x00b", false, false},
		{"backslash", `a\b`, false, true},
		{"trailing slash", "a/b/", false, false},
		{"empty", "", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := openTestStore(t)
			sub := seedSubmission(t, h, "uploading", []string{"uploading"}, []int{0})

			if got := gate.ValidPath([]byte(tt.path)); got != tt.wantValid {
				t.Errorf("ValidPath = %t, want %t", got, tt.wantValid)
			}
			err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
				_, err := tx.Exec(ctx, `INSERT INTO submission_files (id, submission_id, rel_path, size, role)
					VALUES (?, ?, ?, 1, 'extra')`, store.NewID(), sub, []byte(tt.path))
				return err
			})
			if got := err == nil; got != tt.wantInsert {
				t.Errorf("insert error = %v, want accepted %t", err, tt.wantInsert)
			}
		})
	}
}
