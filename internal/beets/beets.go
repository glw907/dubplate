// Package beets wraps the beet CLI: running quiet imports and parsing
// their tag logs for skipped albums, and triggering Navidrome's Subsonic
// scan endpoint afterward.
package beets

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// subsonicVersion is the Subsonic API version dubplate declares in every
// startScan request, matching the shell layer's pinned value.
const subsonicVersion = "1.16.1"

// subsonicClient is the c= client-name parameter Navidrome logs against
// startScan requests.
const subsonicClient = "musicbox"

// scanTimeout bounds one startScan request.
const scanTimeout = 10 * time.Second

// Runner runs beet imports against one beets library.
type Runner struct {
	// BeetPath is the beet binary to run; a test seam, production sets it
	// to "beet" resolved on PATH.
	BeetPath string
	// ConfigPath is the beets config file, passed as -c. Empty leaves
	// beets to its own config search.
	ConfigPath string
	// StateDir is where each run's tag log is persisted.
	StateDir string
}

// ImportResult is the outcome of one beet import run.
type ImportResult struct {
	// Skipped holds the paths beets declined to auto-tag, one per "skip"
	// tag-log entry.
	Skipped []string
	// LogPath is the persisted tag log under [Runner.StateDir].
	LogPath string
}

// Import runs a quiet beet import over dir, capturing beets' own tag log
// (the "skip <path>[; <path>...]" lines beets writes for anything it
// declines to auto-tag) into a scratch file, then persisting that content
// under r.StateDir via [atomicfile.WriteFile] and re-reading the
// persisted copy to parse. The tag log records only skips: import
// success is never inferred from it, only from the beets database
// afterward.
//
// collection sets the batch's beets flexible attribute via
// "--set collection=<collection>", routing it to a non-default
// [config.Config.Collections] path in config/beets-config.yaml's
// query-prefixed path formats. An empty collection omits --set entirely,
// so a default-collection import writes no flexible attribute at all and
// the default path format's exact, unmigrated layout is untouched.
func (r *Runner) Import(ctx context.Context, dir, collection string) (*ImportResult, error) {
	scratch, err := os.CreateTemp(r.StateDir, "beet-import-*.log")
	if err != nil {
		return nil, fmt.Errorf("create scratch tag log: %w", err)
	}
	scratchPath := scratch.Name()
	if err := scratch.Close(); err != nil {
		_ = os.Remove(scratchPath)
		return nil, fmt.Errorf("close scratch tag log %s: %w", scratchPath, err)
	}
	defer func() { _ = os.Remove(scratchPath) }()

	var args []string
	if r.ConfigPath != "" {
		args = append(args, "-c", r.ConfigPath)
	}
	args = append(args, "import", "-q", "-l", scratchPath)
	if collection != "" {
		args = append(args, "--set", "collection="+collection)
	}
	args = append(args, dir)

	out, err := exec.CommandContext(ctx, r.BeetPath, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("beet import %s: %w: %s", dir, err, strings.TrimSpace(string(out)))
	}

	data, err := os.ReadFile(scratchPath)
	if err != nil {
		return nil, fmt.Errorf("read scratch tag log %s: %w", scratchPath, err)
	}

	logPath := filepath.Join(r.StateDir, fmt.Sprintf("import-%s.log", time.Now().Format("20060102-150405.000000000")))
	if err := atomicfile.WriteFile(logPath, data, 0o644); err != nil {
		return nil, fmt.Errorf("persist tag log: %w", err)
	}

	persisted, err := os.ReadFile(logPath)
	if err != nil {
		return nil, fmt.Errorf("read persisted tag log %s: %w", logPath, err)
	}

	return &ImportResult{Skipped: parseSkips(persisted), LogPath: logPath}, nil
}

// parseSkips extracts skipped paths from beets' tag-log lines, matching
// the shell layer's move_skips_and_notify: a line starting "skip " holds
// one or more paths separated by "; ", each trimmed of at most one
// leading space, empty entries dropped.
func parseSkips(data []byte) []string {
	var skipped []string
	for line := range strings.Lines(string(data)) {
		rest, ok := strings.CutPrefix(strings.TrimRight(line, "\n"), "skip ")
		if !ok {
			continue
		}
		for part := range strings.SplitSeq(rest, ";") {
			trimmed := strings.TrimPrefix(part, " ")
			if trimmed != "" {
				skipped = append(skipped, trimmed)
			}
		}
	}
	return skipped
}

// AlbumPaths is one album's imported item paths, in the order the beets
// database returned them.
type AlbumPaths struct {
	// AlbumID is the beets database's albums.id, or 0 for a singleton
	// import with no album grouping -- each singleton row is its own
	// AlbumPaths rather than being merged into one group by the shared
	// zero ID.
	AlbumID int
	Paths   []string
}

// ResolveLibraryPaths queries the beets database at dbPath, via the
// sqlite3 CLI, for every item beets added between since and until
// (inclusive, [Config.Thresholds]-independent Unix-seconds bounds a
// caller narrows around one beet import call), grouped by album in
// database order. It is the only way a caller learns where beets
// actually put a batch's files: the tag log records only skips, and a
// successful import renames and moves every file, so the database is the
// sole surviving record.
//
// This does not by itself say which uploader or staged directory an
// album came from -- a caller matches [AlbumPaths] back to its own
// pending units positionally, in staging order, since one shared beets
// invocation carries no per-uploader join key.
func ResolveLibraryPaths(ctx context.Context, sqlitePath, dbPath string, since, until float64) ([]AlbumPaths, error) {
	query := fmt.Sprintf(
		"SELECT album_id, path FROM items WHERE added >= %s AND added <= %s ORDER BY album_id, path;",
		strconv.FormatFloat(since, 'f', -1, 64), strconv.FormatFloat(until, 'f', -1, 64),
	)
	out, err := exec.CommandContext(ctx, sqlitePath, "-noheader", "-separator", "\t", dbPath, query).Output()
	if err != nil {
		return nil, fmt.Errorf("query beets database %s: %w", dbPath, err)
	}

	var groups []AlbumPaths
	for line := range strings.Lines(string(out)) {
		line = strings.TrimRight(line, "\n")
		if line == "" {
			continue
		}
		idStr, path, ok := strings.Cut(line, "\t")
		if !ok {
			return nil, fmt.Errorf("query beets database %s: unexpected row %q", dbPath, line)
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("query beets database %s: parse album_id %q: %w", dbPath, idStr, err)
		}

		// A singleton (album_id 0) never merges with another row: each
		// one is its own release, unlike a real album_id, which groups
		// every row sharing it into one AlbumPaths.
		if id != 0 && len(groups) > 0 && groups[len(groups)-1].AlbumID == id {
			last := &groups[len(groups)-1]
			last.Paths = append(last.Paths, path)
			continue
		}
		groups = append(groups, AlbumPaths{AlbumID: id, Paths: []string{path}})
	}
	return groups, nil
}

// TriggerScan requests a library rescan from Navidrome's Subsonic
// startScan endpoint. Subsonic's query-param authentication -- t is
// md5(password+salt), s a fresh random salt per call, u/v/c the
// username, API version, and client name -- goes directly on the request
// URL, since this is a native HTTP call rather than a curl subprocess:
// there is no argv for the credentials to leak through, and no
// --config-file dance to clean up. A failed request is returned to the
// caller, who treats the scan trigger as best-effort margin -- Navidrome's
// own daily Scanner.Schedule backstop still runs regardless.
func TriggerScan(ctx context.Context, navidromeURL, user, password string) error {
	salt := make([]byte, 8)
	if _, err := rand.Read(salt); err != nil {
		return fmt.Errorf("generate scan salt: %w", err)
	}
	saltHex := hex.EncodeToString(salt)
	sum := md5.Sum([]byte(password + saltHex))
	token := hex.EncodeToString(sum[:])

	u, err := url.Parse(strings.TrimSuffix(navidromeURL, "/") + "/rest/startScan")
	if err != nil {
		return fmt.Errorf("parse Navidrome URL %s: %w", navidromeURL, err)
	}
	q := u.Query()
	q.Set("u", user)
	q.Set("t", token)
	q.Set("s", saltHex)
	q.Set("v", subsonicVersion)
	q.Set("c", subsonicClient)
	q.Set("f", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("build scan request: %w", err)
	}

	client := &http.Client{Timeout: scanTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("navidrome scan request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return fmt.Errorf("navidrome scan request: status %s", resp.Status)
	}
	return nil
}
