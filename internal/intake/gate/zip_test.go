package gate

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"context"
	"errors"
	"hash/crc32"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

const stagedArchive = "tus/00000000000000000000000000000001"

// zipFixture builds an archive entry by entry. Every entry is stored, so a
// test can place bytes by offset.
type zipFixture struct {
	t   *testing.T
	buf bytes.Buffer
	w   *zip.Writer
}

func newZIPFixture(t *testing.T) *zipFixture {
	t.Helper()
	f := &zipFixture{t: t}
	f.w = zip.NewWriter(&f.buf)
	return f
}

func (f *zipFixture) file(name, data string) *zipFixture {
	f.t.Helper()
	w, err := f.w.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
	if err != nil {
		f.t.Fatalf("create %q: %v", name, err)
	}
	if _, err := io.WriteString(w, data); err != nil {
		f.t.Fatalf("write %q: %v", name, err)
	}
	return f
}

// raw writes an entry with a header and body the caller controls, so a test
// can declare a size the body does not match.
func (f *zipFixture) raw(h *zip.FileHeader, body []byte) *zipFixture {
	f.t.Helper()
	w, err := f.w.CreateRaw(h)
	if err != nil {
		f.t.Fatalf("create raw %q: %v", h.Name, err)
	}
	if _, err := w.Write(body); err != nil {
		f.t.Fatalf("write raw %q: %v", h.Name, err)
	}
	return f
}

func (f *zipFixture) bytes() []byte {
	f.t.Helper()
	if err := f.w.Close(); err != nil {
		f.t.Fatalf("close zip: %v", err)
	}
	return f.buf.Bytes()
}

// stageZIP writes data as the staged archive and returns a root on the staging
// directory.
func stageZIP(t *testing.T, data []byte) *os.Root {
	t.Helper()
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "tus"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, stagedArchive), data, 0o644); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return root
}

func zipPolicy() Policy {
	return Policy{MaxAlbumBytes: 1 << 20, MaxFiles: 100, Allowlist: []string{"flac"}}
}

func walkStaged(t *testing.T, data []byte, p Policy, tracks bool) (*os.Root, *ZIPWalk, error) {
	t.Helper()
	root := stageZIP(t, data)
	w, err := WalkZIP(root, stagedArchive, p, tracks)
	if w != nil {
		t.Cleanup(func() { _ = w.Close() })
	}
	return root, w, err
}

func wantRefusal(t *testing.T, err error, code reject.Code, path string) {
	t.Helper()
	var ref *ZIPRefusal
	if !errors.As(err, &ref) {
		t.Fatalf("err = %v, want a *ZIPRefusal for %s", err, code)
	}
	if ref.Reason.Code != code || string(ref.Path) != path {
		t.Errorf("refusal = %s at %q, want %s at %q", ref.Reason.Code, ref.Path, code, path)
	}
}

func wantCouldNotRun(t *testing.T, err error, cause error) {
	t.Helper()
	var ref *ZIPRefusal
	if err == nil || errors.As(err, &ref) {
		t.Fatalf("err = %v, want a could-not-run error, never a verdict", err)
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Errorf("err = %v, want it to wrap %v", err, cause)
	}
}

func wantNoExtraction(t *testing.T, root *os.Root, dir string) {
	t.Helper()
	if _, err := root.Stat(dir); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stat %s = %v, want it absent", dir, err)
	}
}

func TestWalkZIPRefusals(t *testing.T) {
	symlink := &zip.FileHeader{Name: "link.flac", Method: zip.Store}
	symlink.SetMode(fs.ModeSymlink | 0o777)

	tests := []struct {
		name  string
		zip   func(t *testing.T) []byte
		cap   int64
		code  reject.Code
		where string
	}{
		{"a symlink member", func(t *testing.T) []byte {
			f := newZIPFixture(t).file("01.flac", "a")
			w, err := f.w.CreateHeader(symlink)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.WriteString(w, "01.flac"); err != nil {
				t.Fatal(err)
			}
			return f.bytes()
		}, 0, reject.UnsafePath, "link.flac"},
		{"an absolute name", func(t *testing.T) []byte {
			return newZIPFixture(t).file("/etc/passwd.flac", "a").bytes()
		}, 0, reject.UnsafePath, "/etc/passwd.flac"},
		{"a dotdot name", func(t *testing.T) []byte {
			return newZIPFixture(t).file("../x.flac", "a").bytes()
		}, 0, reject.UnsafePath, "../x.flac"},
		{"a NUL name", func(t *testing.T) []byte {
			return newZIPFixture(t).file("a\x00b.flac", "a").bytes()
		}, 0, reject.UnsafePath, "a\x00b.flac"},
		{"a backslash name", func(t *testing.T) []byte {
			return newZIPFixture(t).file(`a\b.flac`, "a").bytes()
		}, 0, reject.UnsafePath, `a\b.flac`},
		{"an empty name", func(t *testing.T) []byte {
			return newZIPFixture(t).file("", "a").bytes()
		}, 0, reject.UnsafePath, ""},
		{"a directory entry with content", func(t *testing.T) []byte {
			// The writer refuses content under a directory name, so the entry is
			// written as a file and renamed in place, same length.
			data := newZIPFixture(t).file("01.flac", "a").file("dirx", "hello").bytes()
			return bytes.ReplaceAll(data, []byte("dirx"), []byte("dir/"))
		}, 0, reject.UnsafePath, "dir/"},
		{"a duplicate name", func(t *testing.T) []byte {
			return newZIPFixture(t).file("01.flac", "a").file("02.flac", "b").file("01.flac", "c").bytes()
		}, 0, reject.UnsafePath, "01.flac"},
		{"a declared total over the cap", func(t *testing.T) []byte {
			return newZIPFixture(t).file("01.flac", "aaaaaa").file("02.flac", "bbbbbb").bytes()
		}, 11, reject.TooLarge, ""},
		{"a file that is not a ZIP", func(t *testing.T) []byte {
			return []byte("this is not a zip archive at all")
		}, 0, reject.UnpackFailed, ""},
		{"no audio member", func(t *testing.T) []byte {
			return newZIPFixture(t).file("cover.jpg", "a").file("notes.txt", "b").bytes()
		}, 0, reject.NoAudio, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := zipPolicy()
			if tt.cap > 0 {
				p.MaxAlbumBytes = tt.cap
			}
			root, w, err := walkStaged(t, tt.zip(t), p, false)
			if w != nil {
				t.Fatalf("WalkZIP returned a walk, want a refusal")
			}
			wantRefusal(t, err, tt.code, tt.where)
			wantNoExtraction(t, root, "submissions")
		})
	}
}

func TestWalkZIPSumsAtTheCap(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", "aaaaa").file("02.flac", "bbbbb").bytes()
	p := zipPolicy()
	p.MaxAlbumBytes = 10
	_, w, err := walkStaged(t, data, p, false)
	if err != nil {
		t.Fatalf("a total equal to the cap: %v", err)
	}
	if w.Total != 10 {
		t.Errorf("Total = %d, want 10", w.Total)
	}
}

func TestWalkZIPOverflowingSizesAreTooLarge(t *testing.T) {
	huge := func(name string) *zip.FileHeader {
		return &zip.FileHeader{Name: name, Method: zip.Store, UncompressedSize64: 1 << 63, CompressedSize64: 0}
	}
	data := newZIPFixture(t).raw(huge("01.flac"), nil).raw(huge("02.flac"), nil).bytes()
	_, _, err := walkStaged(t, data, zipPolicy(), false)
	wantRefusal(t, err, reject.TooLarge, "")
}

func TestWalkZIPAlbum(t *testing.T) {
	data := newZIPFixture(t).
		file("Artist/", "").
		file("Artist/Album/01 Track.FLAC", "AAAA").
		file("Artist/Album/02.flac", "BBBBBB").
		file("Artist/Album/cover.jpg", "CC").
		bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	for _, m := range w.Members {
		paths = append(paths, string(m.Path))
	}
	want := []string{"Artist/Album/01 Track.FLAC", "Artist/Album/02.flac", "Artist/Album/cover.jpg"}
	if !slices.Equal(paths, want) {
		t.Errorf("members = %q, want %q (the directory entry is not a member)", paths, want)
	}
	if w.Total != 12 || len(w.Extras) != 0 {
		t.Errorf("Total = %d, Extras = %q, want 12 and none", w.Total, w.Extras)
	}

	const dir = "submissions/s1/x"
	if err := w.Extract(context.Background(), root, dir, []string{"f1", "f2", "f3"}); err != nil {
		t.Fatal(err)
	}
	for name, content := range map[string]string{"f1": "AAAA", "f2": "BBBBBB", "f3": "CC"} {
		got, err := os.ReadFile(filepath.Join(rootDir(t, root), dir, name))
		if err != nil || string(got) != content {
			t.Errorf("%s = %q, %v, want %q", name, got, err, content)
		}
	}
	entries, err := fs.ReadDir(root.FS(), dir)
	if err != nil || len(entries) != 3 {
		t.Errorf("x/ holds %d entries, %v, want exactly the 3 members", len(entries), err)
	}
}

func TestWalkZIPSeparateTracksKeepsAudioOnly(t *testing.T) {
	data := newZIPFixture(t).
		file("01.flac", "AAAA").
		file("scan/cover.pdf", "PPPPPPPP").
		file("02.flac", "BB").
		bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Members) != 2 || string(w.Members[0].Path) != "01.flac" || string(w.Members[1].Path) != "02.flac" {
		t.Errorf("members = %v, want the two tracks", w.Members)
	}
	if len(w.Extras) != 1 || string(w.Extras[0]) != "scan/cover.pdf" {
		t.Errorf("extras = %q, want [scan/cover.pdf]", w.Extras)
	}
	if w.Total != 6 {
		t.Errorf("Total = %d, want 6: an extra is never written, so it never counts", w.Total)
	}
	if err := w.Extract(context.Background(), root, "submissions/s1/x", []string{"f1", "f2"}); err != nil {
		t.Fatal(err)
	}
	entries, err := fs.ReadDir(root.FS(), "submissions/s1/x")
	if err != nil || len(entries) != 2 {
		t.Errorf("x/ holds %d entries, %v, want the 2 tracks", len(entries), err)
	}
}

func TestWalkZIPSeparateTracksWithoutATrack(t *testing.T) {
	data := newZIPFixture(t).file("scan.pdf", "P").bytes()
	_, _, err := walkStaged(t, data, zipPolicy(), true)
	wantRefusal(t, err, reject.NoAudio, "")
}

func TestWalkZIPKeepsNonUTF8NameBytes(t *testing.T) {
	const name = "caf\xe9/01.flac"
	data := newZIPFixture(t).file(name, "AAAA").bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Members) != 1 || !bytes.Equal(w.Members[0].Path, []byte(name)) {
		t.Fatalf("members = %v, want the name bytes unchanged", w.Members)
	}
	if err := w.Extract(context.Background(), root, "submissions/s1/x", []string{"f1"}); err != nil {
		t.Fatal(err)
	}
}

// extractFaults are archives that open and walk cleanly and fail in
// archive/zip during extraction, each after a good member.
func TestExtractMemberFaultsAreUnpackFailed(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 100)
	var deflated bytes.Buffer
	fw, err := flate.NewWriter(&deflated, flate.BestCompression)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fw.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := fw.Close(); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		bad  *zip.FileHeader
		data []byte
	}{
		{"data past the declared size",
			&zip.FileHeader{Name: "02.flac", Method: zip.Deflate, CRC32: crc32.ChecksumIEEE(body[:10]), CompressedSize64: uint64(deflated.Len()), UncompressedSize64: 10},
			deflated.Bytes()},
		{"a wrong CRC",
			&zip.FileHeader{Name: "02.flac", Method: zip.Store, CRC32: crc32.ChecksumIEEE([]byte("hello")) ^ 1, CompressedSize64: 5, UncompressedSize64: 5},
			[]byte("hello")},
		{"data short of the declared size",
			&zip.FileHeader{Name: "02.flac", Method: zip.Store, CRC32: crc32.ChecksumIEEE([]byte("hello")), CompressedSize64: 5, UncompressedSize64: 10},
			[]byte("hello")},
		{"an unsupported method",
			&zip.FileHeader{Name: "02.flac", Method: 99, CRC32: 1, CompressedSize64: 5, UncompressedSize64: 5},
			[]byte("hello")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := newZIPFixture(t).file("01.flac", "good").raw(tt.bad, tt.data).bytes()
			root, w, err := walkStaged(t, data, zipPolicy(), false)
			if err != nil {
				t.Fatalf("the walk must pass, the fault is in the member: %v", err)
			}
			const dir = "submissions/s1/x"
			err = w.Extract(context.Background(), root, dir, []string{"f1", "f2"})
			wantRefusal(t, err, reject.UnpackFailed, "")
			wantNoExtraction(t, root, dir)
		})
	}
}

// failingReaderAt serves data except in [from, to), where every read fails
// with EIO, the shape of a disk fault under the staged archive.
type failingReaderAt struct {
	data     []byte
	from, to int64
}

func (r *failingReaderAt) ReadAt(p []byte, off int64) (int, error) {
	if off < r.to && off+int64(len(p)) > r.from {
		return 0, syscall.EIO
	}
	return bytes.NewReader(r.data).ReadAt(p, off)
}

func TestZIPReadErrorIsCouldNotRun(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", strings.Repeat("g", 3000)).bytes()
	size := int64(len(data))
	const bodyAt = int64(30 + len("01.flac"))

	t.Run("mid-member", func(t *testing.T) {
		w, err := walkZIP(&failingReaderAt{data: data, from: bodyAt, to: bodyAt + 4}, size, zipPolicy(), false)
		if err != nil {
			t.Fatal(err)
		}
		root := stageZIP(t, nil)
		err = w.Extract(context.Background(), root, "submissions/s1/x", []string{"f1"})
		wantCouldNotRun(t, err, syscall.EIO)
		wantNoExtraction(t, root, "submissions/s1/x")
	})
	t.Run("during the walk", func(t *testing.T) {
		_, err := walkZIP(&failingReaderAt{data: data, from: 0, to: size}, size, zipPolicy(), false)
		wantCouldNotRun(t, err, syscall.EIO)
	})
	t.Run("in the central directory only", func(t *testing.T) {
		cd := size - 22 - 46 - int64(len("01.flac"))
		_, err := walkZIP(&failingReaderAt{data: data, from: cd, to: size - 22}, size, zipPolicy(), false)
		wantCouldNotRun(t, err, syscall.EIO)
	})
}

func TestRecordedReaderAt(t *testing.T) {
	t.Run("a negative offset is the archive's fault", func(t *testing.T) {
		r := &recordedReaderAt{r: bytes.NewReader([]byte("abc"))}
		if _, err := r.ReadAt(make([]byte, 1), -1); !errors.Is(err, zip.ErrFormat) {
			t.Errorf("ReadAt(-1) = %v, want zip.ErrFormat", err)
		}
		if r.err != nil {
			t.Errorf("recorded %v, want nothing: an offset the archive chose is no server fault", r.err)
		}
	})
	t.Run("EOF is not recorded", func(t *testing.T) {
		r := &recordedReaderAt{r: bytes.NewReader([]byte("abc"))}
		_, _ = r.ReadAt(make([]byte, 10), 0)
		if r.err != nil {
			t.Errorf("recorded %v, want nothing", r.err)
		}
	})
	t.Run("the first other error is kept", func(t *testing.T) {
		r := &recordedReaderAt{r: &failingReaderAt{from: 0, to: 10}}
		_, _ = r.ReadAt(make([]byte, 1), 0)
		r.r = bytes.NewReader(nil)
		_, _ = r.ReadAt(make([]byte, 1), 5)
		if !errors.Is(r.err, syscall.EIO) {
			t.Errorf("recorded %v, want EIO to persist", r.err)
		}
	})
}

type failingWriter struct{ err error }

func (w failingWriter) Write([]byte) (int, error) { return 0, w.err }

func TestCopyMemberWriteErrorIsCouldNotRun(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", "good bytes").bytes()
	w, err := walkZIP(bytes.NewReader(data), int64(len(data)), zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	err = w.copyMember(failingWriter{syscall.ENOSPC}, w.Members[0].file)
	wantCouldNotRun(t, err, syscall.ENOSPC)
}

func TestExtractDestinationFaultKeepsTheArchive(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}
	data := newZIPFixture(t).file("01.flac", "good bytes").bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	dir := rootDir(t, root)
	if err := os.Mkdir(filepath.Join(dir, "submissions"), 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(filepath.Join(dir, "submissions"), 0o755) })

	err = w.Extract(context.Background(), root, "submissions/s1/x", []string{"f1"})
	wantCouldNotRun(t, err, fs.ErrPermission)
	if _, err := root.Stat(stagedArchive); err != nil {
		t.Errorf("archive: %v, want it kept", err)
	}
}

func TestExtractClearsAStaleDirectory(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", "good bytes").bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	const dir = "submissions/s1/x"
	if err := root.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := root.WriteFile(dir+"/stale", []byte("left by a crashed run"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := w.Extract(context.Background(), root, dir, []string{"f1"}); err != nil {
		t.Fatal(err)
	}
	if _, err := root.Stat(dir + "/stale"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("stale file: %v, want it removed", err)
	}
	if _, err := root.Stat(dir + "/f1"); err != nil {
		t.Errorf("member: %v", err)
	}
}

func TestExtractCancelled(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", "good bytes").bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err = w.Extract(ctx, root, "submissions/s1/x", []string{"f1"})
	wantCouldNotRun(t, err, context.Canceled)
	wantNoExtraction(t, root, "submissions/s1/x")
}

func TestExtractNeedsOneNamePerMember(t *testing.T) {
	data := newZIPFixture(t).file("01.flac", "a").file("02.flac", "b").bytes()
	root, w, err := walkStaged(t, data, zipPolicy(), false)
	if err != nil {
		t.Fatal(err)
	}
	err = w.Extract(context.Background(), root, "submissions/s1/x", []string{"f1"})
	wantCouldNotRun(t, err, nil)
	wantNoExtraction(t, root, "submissions/s1/x")
}

func TestWalkZIPMissingArchiveIsCouldNotRun(t *testing.T) {
	root := stageZIP(t, nil)
	_, err := WalkZIP(root, "tus/absent", zipPolicy(), false)
	wantCouldNotRun(t, err, fs.ErrNotExist)
}

// rootDir returns the directory a root was opened on.
func rootDir(t *testing.T, root *os.Root) string {
	t.Helper()
	return root.Name()
}
