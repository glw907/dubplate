package backup

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"

	"github.com/glw907/dubplate/internal/engine/store"
)

// summaryLimit bounds how many entries of each list the operator's
// terminal names.
const summaryLimit = 20

// A report is a reconciliation of a store's catalog against a library
// tree: the tracks whose file is not a regular file under the root, and
// the regular files under the root that no track names. Every path is
// relative to the root, and holds the catalog's or the file system's
// bytes as they are, valid UTF-8 or not.
type report struct {
	root string

	// absent says why the root was not walked, and is "" when it was.
	absent string

	noUsers  bool
	unfiled  []string
	unlisted []string

	// walkErrs names each subtree the walk could not read.
	walkErrs []string
}

// reconcile reconciles the catalog h reads against the library tree
// under root. A track's file is its collection's root path joined with
// its own relative path under root, and counts only when it is a
// regular file itself, never through a symlink; the walk of root does
// not follow symlinks either, so a symlink is never a file without a
// track.
//
// An absent root, or none, is no error: the report says so and lists
// every track as without a file. A subtree the walk cannot read is
// named in the report, and the rest of the tree is still reconciled.
func reconcile(ctx context.Context, h *store.Handle, root string) (*report, error) {
	r := &report{root: root}
	var users int
	if err := h.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&users); err != nil {
		return nil, fmt.Errorf("reconcile: count users: %w", err)
	}
	r.noUsers = users == 0
	tracks, err := trackPaths(ctx, h)
	if err != nil {
		return nil, err
	}

	switch resolved, err := filepath.EvalSymlinks(root); {
	case root == "":
		r.absent = "no library root is configured"
	case errors.Is(err, fs.ErrNotExist) || errors.Is(err, syscall.ENOTDIR):
		r.absent = "library root " + root + " is absent"
	case err != nil:
		return nil, fmt.Errorf("reconcile: resolve library root %s: %w", root, err)
	default:
		r.root = resolved
	}
	if r.absent != "" {
		r.unfiled = slices.Sorted(maps.Keys(tracks))
		return r, nil
	}

	for p := range tracks {
		info, err := os.Lstat(filepath.Join(r.root, p))
		if err != nil || !info.Mode().IsRegular() {
			r.unfiled = append(r.unfiled, p)
		}
	}
	slices.Sort(r.unfiled)
	if err := r.walk(ctx, tracks); err != nil {
		return nil, err
	}
	return r, nil
}

// trackPaths returns the path of every track's file relative to the
// library root, as bytes read from the catalog.
func trackPaths(ctx context.Context, h *store.Handle) (map[string]struct{}, error) {
	rows, err := h.Query(ctx, `SELECT c.root_path, t.rel_path FROM tracks t JOIN collections c ON c.id = t.collection_id`)
	if err != nil {
		return nil, fmt.Errorf("reconcile: read tracks: %w", err)
	}
	defer func() { _ = rows.Close() }()
	paths := make(map[string]struct{})
	for rows.Next() {
		var rootPath, relPath []byte
		if err := rows.Scan(&rootPath, &relPath); err != nil {
			return nil, fmt.Errorf("reconcile: read a track: %w", err)
		}
		paths[string(rootPath)+"/"+string(relPath)] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reconcile: read tracks: %w", err)
	}
	return paths, nil
}

// walk lists every regular file under r.root that tracks does not
// name. filepath.WalkDir reports a directory it cannot read in a
// second call for it, after which the walk goes on with its siblings.
func (r *report) walk(ctx context.Context, tracks map[string]struct{}) error {
	err := filepath.WalkDir(r.root, func(p string, d fs.DirEntry, err error) error {
		if cerr := ctx.Err(); cerr != nil {
			return cerr
		}
		if err != nil {
			r.walkErrs = append(r.walkErrs, "unreadable subtree "+quotePath(p)+": "+err.Error())
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(r.root, p)
		if err != nil {
			return err
		}
		if _, ok := tracks[rel]; !ok {
			r.unlisted = append(r.unlisted, rel)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("reconcile: walk %s: %w", r.root, err)
	}
	return nil
}

// render renders r: its counts, then at most limit entries of each
// list, one per line, and a count of the rest. The report file carries
// every entry, with a limit of [math.MaxInt]; the operator's terminal
// carries [summaryLimit].
func (r *report) render(limit int) []byte {
	var b bytes.Buffer
	r.head(&b)
	list(&b, "tracks without a file:", r.unfiled, limit)
	list(&b, "files without a track:", r.unlisted, limit)
	return b.Bytes()
}

func (r *report) head(b *bytes.Buffer) {
	if r.absent != "" {
		b.WriteString(r.absent + ": every track is listed as without a file\n")
	} else {
		b.WriteString("library root: " + quotePath(r.root) + "\n")
	}
	fmt.Fprintf(b, "tracks without a file: %d\nfiles without a track: %d\n", len(r.unfiled), len(r.unlisted))
	if r.noUsers {
		b.WriteString("the users table is empty: run member import\n")
	}
	for _, e := range r.walkErrs {
		b.WriteString(e + "\n")
	}
}

// list writes title and at most limit of paths, one per line, and a
// count of the rest. It writes nothing for an empty list.
func list(b *bytes.Buffer, title string, paths []string, limit int) {
	if len(paths) == 0 {
		return
	}
	b.WriteString(title + "\n")
	for _, p := range paths[:min(limit, len(paths))] {
		b.WriteString("  " + quotePath(p) + "\n")
	}
	if rest := len(paths) - limit; rest > 0 {
		fmt.Fprintf(b, "  and %d more\n", rest)
	}
}

// quotePath returns p as it is when it is printable UTF-8, and quoted
// with Go's escapes otherwise, so one line holds exactly one path and a
// name that is not UTF-8 keeps every byte.
func quotePath(p string) string {
	if utf8.ValidString(p) && !strings.ContainsFunc(p, func(r rune) bool { return !unicode.IsPrint(r) }) {
		return p
	}
	return strconv.Quote(p)
}
