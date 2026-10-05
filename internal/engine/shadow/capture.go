package shadow

import (
	"bytes"
	"cmp"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// A captureResult is one pass's walk of the observe directory.
type captureResult struct {
	Captures     []Capture
	Excluded     []excludedDir // multi-album directories, with their group counts
	PartialReads []string      // "<uploader>/<basename>" of every discarded capture
}

// An excludedDir is a directory the capture could not resolve to one
// album, kept so the report prints a count with a reason rather than
// silence.
type excludedDir struct {
	Uploader string
	Basename string
	Reason   string
	Groups   int // for ExcludedMultiAlbum, the number of (albumartist, album) groups found
}

// capture walks observeDir/<uploader>/<basename>/ and reads one fixture
// per album directory.
//
// The walk is lexical, matching os.ReadDir's own order and the
// incumbent's listContributorDirs (_salvage/internal/importer/importer.go:1238-1254),
// so a budget-capped pass reaches albums in a stated order rather than
// an accidental one.
//
// Three rules, each of which costs an album rather than risking a lie:
//
// A capture that cannot read every entry it listed is DISCARDED
// (contract S8). The importer moves an album out of the inbox with
// os.Rename or a copy-then-remove fallback (_salvage/internal/review/move.go:24-40),
// so a walk can list four tracks and take ENOENT on the fourth, and
// once the album is gone there is no next pass to supersede a partial
// fixture with a whole one. review.Move renames the album directory
// itself, so the same race can take ENOENT on the whole directory: the
// walk lists a basename and then fails to os.ReadDir it. That directory
// is recorded as a partial read too, and the walk moves on to the next
// one rather than failing the pass.
//
// A directory whose items group to more than one album is EXCLUDED with
// its group count (contract S9). The incumbent ran group_albums: yes
// (config/beets-config.yaml at the pre-greenfield tag), so beets is free
// to split one uploaded directory into several verdicts, while
// tag.Album takes items []matcher.ItemFields and calls MostCommonTags
// over all of them (internal/engine/tag/tag.go:75). Rung 4 observes
// only what groups to exactly one album, and counts the rest.
//
// The key is (cmp.Or(AlbumArtist, Artist), Album), which is beets' own:
// group(item) returns (item.albumartist or item.artist, item.album)
// (beets/importer/stages.py:102-103, consumed at :112-113). The
// fallback is what handles the estate's real material: a contributor
// ZIP extracts flat (the same file, lines 19-22) and its tracks
// often carry ARTIST and no ALBUMARTIST, which an albumartist-only key
// collapses into one group and observes as one album while beets splits
// it into several verdicts. The fallback is conductor ruling C1 at
// rung 4a's plan fold (spec:444), which fixes the key as beets' own
// rather than the albumartist-alone form the spec first carried.
//
// A directory whose digest already names a stored observation is
// SKIPPED, which is what makes the capture idempotent (contract S7).
func capture(observeDir string, has func(id string) bool, now time.Time) (captureResult, error) {
	var result captureResult

	uploaders, err := os.ReadDir(observeDir)
	if err != nil {
		return captureResult{}, fmt.Errorf("read observe directory: %w", err)
	}

	for _, uploaderEntry := range uploaders {
		if !uploaderEntry.IsDir() {
			continue
		}
		uploader := uploaderEntry.Name()
		uploaderDir := filepath.Join(observeDir, uploader)

		basenames, err := os.ReadDir(uploaderDir)
		if err != nil {
			result.PartialReads = append(result.PartialReads, uploader)
			continue
		}

		for _, basenameEntry := range basenames {
			if !basenameEntry.IsDir() {
				continue
			}
			basename := basenameEntry.Name()
			albumDir := filepath.Join(uploaderDir, basename)

			c, excl, partial := captureAlbum(uploader, basename, albumDir, now)
			switch {
			case partial:
				result.PartialReads = append(result.PartialReads, uploader+"/"+basename)
			case excl != nil:
				result.Excluded = append(result.Excluded, *excl)
			case c != nil:
				if has(observationID(c.Digest)) {
					continue
				}
				result.Captures = append(result.Captures, *c)
			}
		}
	}

	return result, nil
}

// captureAlbum reads one album directory's entries, without descending
// into any subdirectory (the incumbent's own consumed layout is one
// level, importer.go:839-857). It reads only files whose extension is
// ".flac", case-insensitively, because the reader reads no other
// format and a cover image is not a track; every entry, FLAC or not,
// still enters the file set the digest is computed over, because a
// changed cover is a changed upload.
func captureAlbum(uploader, basename, dir string, now time.Time) (capture *Capture, excluded *excludedDir, partial bool) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, nil, true
	}

	var files []CapturedFile
	var items []matcher.ItemFields

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, nil, true
		}
		files = append(files, CapturedFile{
			Name:    entry.Name(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})

		if !strings.EqualFold(filepath.Ext(entry.Name()), ".flac") {
			continue
		}

		fields, err := readFLAC(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, nil, true
		}
		items = append(items, fields)
	}

	if len(items) == 0 {
		return nil, &excludedDir{Uploader: uploader, Basename: basename, Reason: ExcludedNoTracks}, false
	}

	groups := make(map[[2]string]struct{})
	for _, it := range items {
		key := [2]string{cmp.Or(it.AlbumArtist, it.Artist), it.Album}
		groups[key] = struct{}{}
	}
	if len(groups) > 1 {
		return nil, &excludedDir{Uploader: uploader, Basename: basename, Reason: ExcludedMultiAlbum, Groups: len(groups)}, false
	}

	return &Capture{
		Uploader:   uploader,
		Basename:   basename,
		Digest:     digestFor(uploader, basename, files),
		Files:      files,
		Items:      items,
		CapturedAt: now,
	}, nil, false
}

// digestFor computes a capture's identity: the hex SHA-256 over the
// uploader, the basename, and each file's name, size and modification
// time in lexical name order, NUL-separated.
func digestFor(uploader, basename string, files []CapturedFile) string {
	sorted := slices.Clone(files)
	slices.SortFunc(sorted, func(a, b CapturedFile) int { return strings.Compare(a.Name, b.Name) })

	var buf bytes.Buffer
	writeField(&buf, uploader)
	writeField(&buf, basename)
	for _, f := range sorted {
		writeField(&buf, f.Name)
		writeField(&buf, strconv.FormatInt(f.Size, 10))
		writeField(&buf, f.ModTime.Format(time.RFC3339Nano))
	}
	sum := sha256.Sum256(buf.Bytes())
	return hex.EncodeToString(sum[:])
}

// writeField writes s to buf followed by a NUL separator.
func writeField(buf *bytes.Buffer, s string) {
	buf.WriteString(s)
	buf.WriteByte(0)
}

// observationID is "shadow-" plus a digest's first sixteen hex
// characters (Resolution R5). It names the observation directory, the
// pending row and the resolved row.
func observationID(digest string) string {
	const idLen = 16
	return "shadow-" + digest[:idLen]
}
