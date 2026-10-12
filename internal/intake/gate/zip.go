package gate

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"strings"

	"github.com/glw907/dubplate/internal/reject"
)

// ZIPRefusal is the verdict that a staged ZIP is refused: the reason, and
// the member's name as the archive spelled it (bytes, not necessarily UTF-8)
// when the reason is about one member. Any other error from this file is
// "could not run": a fault of the server's that keeps the archive.
type ZIPRefusal struct {
	Reason reject.Reason
	Path   []byte
	cause  error
}

func (e *ZIPRefusal) Error() string {
	s := "zip refused: " + string(e.Reason.Code)
	if e.Path != nil {
		s += fmt.Sprintf(" at %q", e.Path)
	}
	if e.cause != nil {
		s += ": " + e.cause.Error()
	}
	return s
}

func (e *ZIPRefusal) Unwrap() error { return e.cause }

// ZIPMember is one member the walk keeps for extraction.
type ZIPMember struct {
	Path []byte // the name as the archive spelled it
	Size int64  // the declared uncompressed size
	file *zip.File
}

// ZIPWalk is a staged ZIP whose central directory passed the walk. Members
// and Total are what [ZIPWalk.Extract] will write; Extras are the names a
// separate-tracks ZIP carried that are not audio, dropped with a warning. It
// holds the archive open until Close.
type ZIPWalk struct {
	Members []ZIPMember
	Extras  [][]byte
	Total   int64

	src    *recordedReaderAt
	closer io.Closer
}

// WalkZIP judges the staged archive name under root without writing a byte.
// It counts the central directory's headers against p.MaxFiles before
// archive/zip reads it, refuses a member the library could not store
// (a symlink or other special file, a name [ValidPath] refuses, a directory
// with content, a duplicate name) with unsafe-path, sums the members'
// declared sizes against p.MaxAlbumBytes (too-large), and requires an audio
// member (no-audio). A separate-tracks archive keeps its audio members only.
//
// The refusals are *[ZIPRefusal]; an archive archive/zip cannot open is
// unpack-failed. A read error on the staged archive is returned as itself.
func WalkZIP(root *os.Root, name string, p Policy, separateTracks bool) (*ZIPWalk, error) {
	f, err := root.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open staged archive: %w", err)
	}
	fi, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("stat staged archive: %w", err)
	}
	w, err := walkZIP(f, fi.Size(), p, separateTracks)
	if err != nil {
		_ = f.Close()
		return nil, err
	}
	w.closer = f
	return w, nil
}

func walkZIP(ra io.ReaderAt, size int64, p Policy, separateTracks bool) (*ZIPWalk, error) {
	src := &recordedReaderAt{r: ra}
	start, err := directoryStart(src, size)
	if err != nil {
		return nil, archiveFault(src, err)
	}
	over := exceedsHeaders(src, size, start, p.MaxFiles)
	if src.err != nil {
		return nil, archiveFault(src, nil)
	}
	if over {
		return nil, &ZIPRefusal{Reason: reject.Of(reject.TooManyFiles)}
	}

	// The names are checked below, so the GODEBUG that makes NewReader refuse
	// some of them adds nothing.
	zr, err := zip.NewReader(src, size)
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil, archiveFault(src, err)
	}

	w := &ZIPWalk{src: src}
	seen := make(map[string]bool, len(zr.File))
	var total uint64
	var overflow, audio bool
	for _, f := range zr.File {
		name, isDir := strings.CutSuffix(f.Name, "/")
		if !ValidPath([]byte(name)) || f.Mode().Type()&^fs.ModeDir != 0 {
			return nil, unsafeMember(f)
		}
		if isDir {
			if f.UncompressedSize64 != 0 {
				return nil, unsafeMember(f)
			}
			continue
		}
		if seen[f.Name] {
			return nil, unsafeMember(f)
		}
		seen[f.Name] = true

		_, isAudio := p.admitted(extension(name))
		if separateTracks && !isAudio {
			w.Extras = append(w.Extras, []byte(f.Name))
			continue
		}
		audio = audio || isAudio
		if f.UncompressedSize64 > math.MaxInt64-total {
			overflow = true
		} else {
			total += f.UncompressedSize64
		}
		w.Members = append(w.Members, ZIPMember{Path: []byte(f.Name), Size: int64(f.UncompressedSize64), file: f})
	}
	if overflow || int64(total) > p.MaxAlbumBytes {
		return nil, &ZIPRefusal{Reason: reject.Of(reject.TooLarge)}
	}
	if !audio {
		return nil, &ZIPRefusal{Reason: reject.Of(reject.NoAudio)}
	}
	w.Total = int64(total)
	return w, nil
}

// Extract writes the members into dir under root, member i as names[i]: a
// name the server minted, never one a member chose, since a ".." in it would
// climb out of dir. It removes dir first, so a re-run starts clean, and on
// any failure removes it again.
//
// Each member and dir are synced before Extract returns: the archive is the
// only other copy, and it goes once the verdict commits.
//
// An error from archive/zip or its decompressor (a CRC mismatch, data past
// or short of the declared size, an unsupported method) is a *[ZIPRefusal]
// with unpack-failed. A read error on the staged archive, a write or sync
// error, and a cancelled ctx are returned as themselves.
func (w *ZIPWalk) Extract(ctx context.Context, root *os.Root, dir string, names []string) (err error) {
	if len(names) != len(w.Members) {
		return fmt.Errorf("extract: %d names for %d members", len(names), len(w.Members))
	}
	defer func() {
		if err != nil {
			_ = root.RemoveAll(dir)
		}
	}()
	if err := root.RemoveAll(dir); err != nil {
		return fmt.Errorf("clear extraction directory: %w", err)
	}
	if err := root.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create extraction directory: %w", err)
	}
	for i, m := range w.Members {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := w.extractMember(root, path.Join(dir, names[i]), m.file); err != nil {
			return err
		}
	}
	for _, d := range []string{dir, path.Dir(dir)} {
		if err := syncDir(root, d); err != nil {
			return err
		}
	}
	return nil
}

func (w *ZIPWalk) extractMember(root *os.Root, name string, f *zip.File) error {
	dst, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create extracted member: %w", err)
	}
	err = w.copyMember(dst, f)
	if err == nil {
		if serr := dst.Sync(); serr != nil {
			err = fmt.Errorf("sync extracted member: %w", serr)
		}
	}
	if cerr := dst.Close(); err == nil && cerr != nil {
		err = fmt.Errorf("close extracted member: %w", cerr)
	}
	return err
}

// copyMember streams one member to dst. The reader stops at the declared
// size, so what reaches dst is bounded by what the walk summed.
func (w *ZIPWalk) copyMember(dst io.Writer, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return archiveFault(w.src, err)
	}
	defer func() { _ = rc.Close() }()
	sink := &recordedWriter{w: dst}
	if _, err := io.Copy(sink, rc); err != nil {
		if sink.err != nil {
			return fmt.Errorf("write extracted member: %w", sink.err)
		}
		return archiveFault(w.src, err)
	}
	return nil
}

// Close releases the staged archive.
func (w *ZIPWalk) Close() error {
	if w.closer == nil {
		return nil
	}
	return w.closer.Close()
}

func syncDir(root *os.Root, name string) error {
	d, err := root.Open(name)
	if err != nil {
		return fmt.Errorf("open directory to sync: %w", err)
	}
	err = d.Sync()
	if cerr := d.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return fmt.Errorf("sync directory: %w", err)
	}
	return nil
}

// archiveFault classifies err from archive/zip by whose fault it was
// (P19): a read error recorded on the staged archive is the server's, and
// anything else the archive's own.
func archiveFault(src *recordedReaderAt, err error) error {
	if src.err != nil {
		return fmt.Errorf("read staged archive: %w", src.err)
	}
	return &ZIPRefusal{Reason: reject.Of(reject.UnpackFailed), cause: err}
}

func unsafeMember(f *zip.File) *ZIPRefusal {
	return &ZIPRefusal{Reason: reject.Of(reject.UnsafePath), Path: []byte(f.Name)}
}

// recordedReaderAt keeps the first read error of the staged archive other
// than io.EOF, so a caller can tell a fault of the disk from a fault of the
// bytes whatever archive/zip returns. It is not safe for concurrent use.
type recordedReaderAt struct {
	r   io.ReaderAt
	err error
}

func (r *recordedReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < 0 {
		// archive/zip reads at offsets the archive chose, which can be
		// negative; the disk never saw a fault.
		return 0, zip.ErrFormat
	}
	n, err := r.r.ReadAt(p, off)
	if err != nil && err != io.EOF && r.err == nil {
		r.err = err
	}
	return n, err
}

// recordedWriter keeps the first write error, which io.Copy returns
// indistinguishable from a read error.
type recordedWriter struct {
	w   io.Writer
	err error
}

func (w *recordedWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	if err != nil && w.err == nil {
		w.err = err
	}
	return n, err
}
