// Package archive safely extracts contributor-uploaded archives natively,
// via [archive/zip] and [archive/tar] rather than shelling out to bsdtar.
// Every extraction refuses path traversal, stays under a caller-set
// extracted-byte budget, and preflights free space before touching disk.
package archive

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/unix"
)

// ErrPathTraversal marks an archive member whose path would resolve
// outside the extraction destination.
var ErrPathTraversal = errors.New("archive: path traversal member")

// ErrTooLarge marks an archive whose extracted content exceeds
// [Limits.MaxExtractedBytes].
var ErrTooLarge = errors.New("archive: extracted size exceeds limit")

// ErrInsufficientSpace marks a destination filesystem with less free
// space than [Limits.MinFreeBytes].
var ErrInsufficientSpace = errors.New("archive: insufficient free space")

// Limits bounds one [Extract] call.
type Limits struct {
	// MaxExtractedBytes caps the total decompressed size of every member
	// combined, checked as bytes are written rather than trusted from
	// archive headers, so a small archive engineered to decompress far
	// past this budget (a zip bomb) is caught mid-write.
	MaxExtractedBytes int64
	// MinFreeBytes is the free space [Extract] requires on destDir's
	// filesystem before it writes anything.
	MinFreeBytes int64
}

// Extract unpacks archivePath into destDir, which the caller creates as
// an isolated, per-archive directory: two archives in the same batch
// must never share a destDir, since a colliding member name would
// silently merge or clobber. Extract recognizes .zip, .tar, .tar.gz,
// .tgz, .tar.bz2, and .tbz2 by extension; any other extension is an
// error.
//
// Every member's resolved path is checked against destDir before it is
// written, and a symlink member is recreated as a symlink rather than
// followed or rejected outright -- rejecting symlinks by extension or
// content is [internal/gate]'s job, run over the extracted tree
// afterward. The free-space preflight and the extracted-byte budget both
// run before [internal/gate] ever sees the result: a caller that ignores
// Extract's error is expected to discard destDir, matching the shell
// layer's unconditional cleanup of its own scratch extraction directory.
func Extract(ctx context.Context, archivePath, destDir string, lim Limits) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	avail, err := FreeSpace(destDir)
	if err != nil {
		return err
	}
	if avail < lim.MinFreeBytes {
		return fmt.Errorf("%s: %d bytes available, need %d: %w", destDir, avail, lim.MinFreeBytes, ErrInsufficientSpace)
	}

	remaining := lim.MaxExtractedBytes
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return extractZip(ctx, archivePath, destDir, &remaining)
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"),
		strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"),
		strings.HasSuffix(lower, ".tar"):
		return extractTar(ctx, archivePath, destDir, &remaining)
	default:
		return fmt.Errorf("%s: unsupported archive format", archivePath)
	}
}

// FreeSpace reports the number of bytes available to an unprivileged
// user on the filesystem holding path, matching df's "Avail" column.
func FreeSpace(path string) (int64, error) {
	var stat unix.Statfs_t
	if err := unix.Statfs(path, &stat); err != nil {
		return 0, fmt.Errorf("statfs %s: %w", path, err)
	}
	return int64(stat.Bavail) * stat.Bsize, nil
}

// extractZip unpacks a zip archive, whose central directory Go reads in
// full before this loop runs -- so unlike extractTar, every member name
// here is already known and safe to range over without further framing.
func extractZip(ctx context.Context, archivePath, destDir string, remaining *int64) error {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("open zip %s: %w", archivePath, err)
	}
	defer func() { _ = zr.Close() }()

	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := extractZipEntry(destDir, f, remaining); err != nil {
			return err
		}
	}
	return nil
}

func extractZipEntry(destDir string, f *zip.File, remaining *int64) error {
	fi := f.FileInfo()
	switch {
	case fi.IsDir():
		return writeDir(destDir, f.Name)
	case fi.Mode()&fs.ModeSymlink != 0:
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open symlink member %s: %w", f.Name, err)
		}
		defer func() { _ = rc.Close() }()
		target, err := io.ReadAll(io.LimitReader(rc, 4096))
		if err != nil {
			return fmt.Errorf("read symlink target of %s: %w", f.Name, err)
		}
		return writeSymlink(destDir, f.Name, string(target))
	default:
		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open member %s: %w", f.Name, err)
		}
		defer func() { _ = rc.Close() }()
		return writeFile(destDir, f.Name, fi.Mode(), rc, remaining)
	}
}

// extractTar unpacks a tar archive (optionally gzip- or bzip2-wrapped),
// whose members are only known one at a time as tr.Next() advances the
// stream -- so each member is validated and written as it arrives,
// rather than precomputed as a batch the way extractZip's central
// directory allows.
func extractTar(ctx context.Context, archivePath, destDir string, remaining *int64) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive %s: %w", archivePath, err)
	}
	defer func() { _ = f.Close() }()

	r, err := tarPayloadReader(archivePath, f)
	if err != nil {
		return err
	}

	tr := tar.NewReader(r)
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("read tar member of %s: %w", archivePath, err)
		}

		switch header.Typeflag {
		case tar.TypeDir:
			err = writeDir(destDir, header.Name)
		case tar.TypeReg:
			err = writeFile(destDir, header.Name, fs.FileMode(header.Mode), tr, remaining)
		case tar.TypeSymlink, tar.TypeLink:
			err = writeSymlink(destDir, header.Name, header.Linkname)
		default:
			// Device nodes, FIFOs, and the like never appear in a
			// contributor's music archive; skip rather than fail the
			// whole batch over an entry nothing downstream will ever read.
		}
		if err != nil {
			return err
		}
	}
}

// tarPayloadReader wraps f in a decompressor matching archivePath's
// extension, or returns f unwrapped for a plain .tar.
func tarPayloadReader(archivePath string, f *os.File) (io.Reader, error) {
	lower := strings.ToLower(archivePath)
	switch {
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		gz, err := gzip.NewReader(f)
		if err != nil {
			return nil, fmt.Errorf("open gzip %s: %w", archivePath, err)
		}
		return gz, nil
	case strings.HasSuffix(lower, ".tar.bz2"), strings.HasSuffix(lower, ".tbz2"):
		return bzip2.NewReader(f), nil
	default:
		return f, nil
	}
}

// safeJoin resolves name against destDir and rejects any result that
// would land outside it, refusing an absolute path or a ".." component
// the same way regardless of how it is spelled.
func safeJoin(destDir, name string) (string, error) {
	target := filepath.Join(destDir, name)
	rel, err := filepath.Rel(destDir, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%s: %w", name, ErrPathTraversal)
	}
	return target, nil
}

func writeDir(destDir, name string) error {
	target, err := safeJoin(destDir, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create directory %s: %w", target, err)
	}
	return nil
}

func writeSymlink(destDir, name, linkTarget string) error {
	target, err := safeJoin(destDir, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", target, err)
	}
	// linkTarget is never resolved or validated: a symlink surviving
	// extraction is [internal/gate]'s to reject outright, regardless of
	// where it points.
	if err := os.Symlink(linkTarget, target); err != nil {
		return fmt.Errorf("create symlink %s: %w", target, err)
	}
	return nil
}

// writeFile writes content to name under destDir, permission bits from
// mode (falling back to 0644 when an archive member carries none), and
// counts every byte written against remaining, failing with ErrTooLarge
// the moment the shared budget across the whole archive is exhausted.
func writeFile(destDir, name string, mode fs.FileMode, content io.Reader, remaining *int64) error {
	target, err := safeJoin(destDir, name)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return fmt.Errorf("create parent directory for %s: %w", target, err)
	}

	perm := mode.Perm()
	if perm == 0 {
		perm = 0o644
	}
	out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("create %s: %w", target, err)
	}

	if _, err := io.Copy(capWriter{dst: out, remaining: remaining}, content); err != nil {
		_ = out.Close()
		return fmt.Errorf("write %s: %w", target, err)
	}
	if err := out.Close(); err != nil {
		return fmt.Errorf("close %s: %w", target, err)
	}
	return nil
}

// capWriter enforces a byte budget shared across every member one
// Extract call writes.
type capWriter struct {
	dst       io.Writer
	remaining *int64
}

func (w capWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > *w.remaining {
		return 0, ErrTooLarge
	}
	n, err := w.dst.Write(p)
	*w.remaining -= int64(n)
	return n, err
}
