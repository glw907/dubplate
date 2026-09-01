package intake

import (
	"archive/zip"
	"bytes"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/gate"
	"github.com/glw907/dubplate/internal/reject"
)

func cdPolicy() gate.Policy {
	return gate.Policy{MinSampleRate: 44100, MinBitsPerSample: 16}
}

// --- CheckFLAC -----------------------------------------------------------

// TestCheckFLAC_ValidPasses is the positive control: a real 44.1kHz/16-bit
// encode passes at the CD-quality floor.
func TestCheckFLAC_ValidPasses(t *testing.T) {
	data := makeFlacBytes(t, 44100, false)
	v, err := CheckFLAC(bytes.NewReader(data), "track.flac", cdPolicy())
	if err != nil {
		t.Fatalf("CheckFLAC: %v", err)
	}
	if !v.OK || v.Reason != reject.ReasonNone || v.Name != "track.flac" {
		t.Fatalf("verdict = %+v, want a passing verdict for track.flac", v)
	}
}

// TestCheckFLAC_TruncatedFailsCorrupt ports the truncated-FLAC fixture the
// plan names: a real encode cut off before STREAMINFO's 34 bytes fully
// arrive fails as ReasonCorrupt, not a bare error.
func TestCheckFLAC_TruncatedFailsCorrupt(t *testing.T) {
	data := makeFlacBytes(t, 44100, false)
	truncated := data[:20] // shorter than the 4+4+34 CheckFLAC needs
	v, err := CheckFLAC(bytes.NewReader(truncated), "track.flac", cdPolicy())
	if err != nil {
		t.Fatalf("CheckFLAC: %v", err)
	}
	if v.OK || v.Reason != reject.ReasonCorrupt {
		t.Fatalf("verdict = %+v, want ReasonCorrupt", v)
	}
}

// TestCheckFLAC_BelowCDQuality ports the below-CD-quality fixture the plan
// names, at both the sample-rate and bit-depth floors.
func TestCheckFLAC_BelowCDQuality(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"below sample rate", makeFlacBytes(t, 22050, false)},
		{"below bit depth", makeFlacBytes(t, 44100, true)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := CheckFLAC(bytes.NewReader(tt.data), "track.flac", cdPolicy())
			if err != nil {
				t.Fatalf("CheckFLAC: %v", err)
			}
			if v.OK || v.Reason != reject.ReasonBelowCDQuality {
				t.Fatalf("verdict = %+v, want ReasonBelowCDQuality", v)
			}
		})
	}
}

// TestCheckFLAC_BadMagicFailsCorrupt proves a stream that never carried a
// FLAC magic marker at all -- not merely truncated -- is rejected the same
// way as a truncated one.
func TestCheckFLAC_BadMagicFailsCorrupt(t *testing.T) {
	v, err := CheckFLAC(bytes.NewReader(bytes.Repeat([]byte{0}, 64)), "track.flac", cdPolicy())
	if err != nil {
		t.Fatalf("CheckFLAC: %v", err)
	}
	if v.OK || v.Reason != reject.ReasonCorrupt {
		t.Fatalf("verdict = %+v, want ReasonCorrupt", v)
	}
}

// TestCheckFLAC_ReaderErrorReturnsError proves an underlying transport
// error (as opposed to a merely-short stream) surfaces as a Go error, not
// a FileVerdict -- CheckFLAC tried to validate the file and could not,
// rather than validated it and found it wanting.
func TestCheckFLAC_ReaderErrorReturnsError(t *testing.T) {
	wantErr := errors.New("boom")
	v, err := CheckFLAC(failingReader{err: wantErr}, "track.flac", cdPolicy())
	if err == nil {
		t.Fatal("CheckFLAC: want error, got nil")
	}
	if !errors.Is(err, wantErr) {
		t.Errorf("CheckFLAC err = %v, want wrapped %v", err, wantErr)
	}
	if v != (FileVerdict{}) {
		t.Errorf("verdict = %+v, want the zero value on a reader error", v)
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

// --- CheckZIP --------------------------------------------------------------

// TestCheckZIP_TraversalEntryFailsUnsafePath ports the traversal-ZIP-entry
// fixture the plan names.
func TestCheckZIP_TraversalEntryFailsUnsafePath(t *testing.T) {
	data := buildZip(t, map[string][]byte{"../evil.flac": []byte("x")}, nil)
	verdicts, err := CheckZIP(bytes.NewReader(data), int64(len(data)), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonUnsafePath {
		t.Fatalf("verdicts = %+v, want one ReasonUnsafePath verdict", verdicts)
	}
}

// TestCheckZIP_AbsolutePathFailsUnsafePath proves an absolute member name
// fails the same way as a "../" traversal name.
func TestCheckZIP_AbsolutePathFailsUnsafePath(t *testing.T) {
	data := buildZip(t, map[string][]byte{"/etc/passwd": []byte("x")}, nil)
	verdicts, err := CheckZIP(bytes.NewReader(data), int64(len(data)), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonUnsafePath {
		t.Fatalf("verdicts = %+v, want one ReasonUnsafePath verdict", verdicts)
	}
}

// TestCheckZIP_SymlinkEntryFailsUnsafePath proves a symlink member fails
// as ReasonUnsafePath even though its own name is perfectly safe: the
// symlink target is never trusted, matching internal/archive's own
// refusal to follow one.
func TestCheckZIP_SymlinkEntryFailsUnsafePath(t *testing.T) {
	data := buildZip(t, nil, map[string]string{"album/link.flac": "/etc/passwd"})
	verdicts, err := CheckZIP(bytes.NewReader(data), int64(len(data)), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonUnsafePath {
		t.Fatalf("verdicts = %+v, want one ReasonUnsafePath verdict", verdicts)
	}
}

// TestCheckZIP_EntryOverSizeCapFailsTooLarge proves a member whose central
// directory declares more uncompressed content than the per-entry cap
// fails as ReasonTooLarge, an existing reason CheckZIP reuses rather than
// minting a fourth one.
func TestCheckZIP_EntryOverSizeCapFailsTooLarge(t *testing.T) {
	fh := &zip.FileHeader{Name: "huge.flac", Method: zip.Store}
	fh.UncompressedSize64 = maxZipEntryUncompressed + 1

	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	// Deliberately write no matching content: CheckZIP reads the
	// central-directory header's declared size, never the member's actual
	// bytes, so a header lying about a huge size is exactly what this test
	// exercises without an 2 GiB fixture on disk.
	if _, err := zw.CreateRaw(fh); err != nil {
		t.Fatalf("CreateRaw: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close: %v", err)
	}

	verdicts, err := CheckZIP(bytes.NewReader(buf.Bytes()), int64(buf.Len()), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonTooLarge {
		t.Fatalf("verdicts = %+v, want one ReasonTooLarge verdict", verdicts)
	}
}

// TestCheckZIP_DirectoryEntrySkipped proves a directory member never
// produces a verdict at all, matching gate.Check's own directory skip.
func TestCheckZIP_DirectoryEntrySkipped(t *testing.T) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	if _, err := zw.Create("album/"); err != nil {
		t.Fatalf("Create dir entry: %v", err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close: %v", err)
	}

	verdicts, err := CheckZIP(bytes.NewReader(buf.Bytes()), int64(buf.Len()), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 0 {
		t.Fatalf("verdicts = %+v, want none for a directory-only zip", verdicts)
	}
}

// TestCheckZIP_TooManyEntriesReturnsError proves an entry count past the
// cap fails the whole call rather than producing thousands of verdicts
// for one pathological upload.
func TestCheckZIP_TooManyEntriesReturnsError(t *testing.T) {
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)
	for i := range maxZipEntries + 1 {
		if _, err := zw.Create("f" + strconv.Itoa(i)); err != nil {
			t.Fatalf("Create entry %d: %v", i, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close: %v", err)
	}

	_, err := CheckZIP(bytes.NewReader(buf.Bytes()), int64(buf.Len()), cdPolicy())
	if !errors.Is(err, ErrTooManyEntries) {
		t.Fatalf("CheckZIP err = %v, want wrapped ErrTooManyEntries", err)
	}
}

// TestCheckZIP_CleanAlbumPasses ports the clean-album-ZIP fixture the plan
// names: two real FLAC encodes plus a non-audio sibling, all under safe
// names, every entry verdict OK.
func TestCheckZIP_CleanAlbumPasses(t *testing.T) {
	data := buildZip(t, map[string][]byte{
		"Bad Album/01 Track One.flac": makeFlacBytes(t, 44100, false),
		"Bad Album/02 Track Two.flac": makeFlacBytes(t, 44100, false),
		"Bad Album/cover.jpg":         []byte("art"),
	}, nil)

	verdicts, err := CheckZIP(bytes.NewReader(data), int64(len(data)), cdPolicy())
	if err != nil {
		t.Fatalf("CheckZIP: %v", err)
	}
	if len(verdicts) != 3 {
		t.Fatalf("verdicts = %+v, want 3", verdicts)
	}
	for _, v := range verdicts {
		if !v.OK || v.Reason != reject.ReasonNone {
			t.Errorf("verdict for %s = %+v, want OK", v.Name, v)
		}
	}
}

// --- Decide ------------------------------------------------------------

// TestDecide_AlbumAsVerdictUnit proves one failing file rejects the whole
// album even when every other file passed.
func TestDecide_AlbumAsVerdictUnit(t *testing.T) {
	files := []FileVerdict{
		{Name: "01.flac", OK: true},
		{Name: "02.flac", OK: false, Reason: reject.ReasonBelowCDQuality},
		{Name: "03.flac", OK: true},
	}
	got := Decide(files)
	if got.OK {
		t.Errorf("Decide(%+v).OK = true, want false: one failing file must reject the album", files)
	}
	if len(got.Files) != 3 {
		t.Errorf("Decide(%+v).Files = %+v, want all 3 files carried through regardless of verdict", files, got.Files)
	}
}

// TestDecide_AllPassingIsOK is the positive control for TestDecide_AlbumAsVerdictUnit.
func TestDecide_AllPassingIsOK(t *testing.T) {
	files := []FileVerdict{{Name: "01.flac", OK: true}, {Name: "02.flac", OK: true}}
	if got := Decide(files); !got.OK {
		t.Errorf("Decide(%+v).OK = false, want true: every file passed", files)
	}
}

// --- fixtures ------------------------------------------------------------

// makeFlacBytes synthesizes a real, valid FLAC at rate and returns its
// encoded bytes, at 16-bit unless eightBit is set. It mirrors
// internal/gate's own makeFlac/make8BitFlac test helpers: ffmpeg and flac
// are both present on this workstation, and the source is fully
// deterministic.
func makeFlacBytes(t *testing.T, rate int, eightBit bool) []byte {
	t.Helper()
	dir := t.TempDir()
	wav := filepath.Join(dir, "gen.wav")
	flacPath := filepath.Join(dir, "gen.flac")

	args := []string{"-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1", "-ar", strconv.Itoa(rate)}
	if eightBit {
		args = append(args, "-acodec", "pcm_u8")
	} else {
		args = append(args, "-sample_fmt", "s16")
	}
	args = append(args, wav)
	if out, err := exec.Command("ffmpeg", args...).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg: %v: %s", err, out)
	}
	if out, err := exec.Command("flac", "--best", "-f", "-o", flacPath, wav).CombinedOutput(); err != nil {
		t.Fatalf("flac encode: %v: %s", err, out)
	}

	data, err := os.ReadFile(flacPath)
	if err != nil {
		t.Fatalf("read encoded fixture: %v", err)
	}
	return data
}

// buildZip writes files and symlinks (name -> link target) into an
// in-memory ZIP and returns its bytes.
func buildZip(t *testing.T, files map[string][]byte, symlinks map[string]string) []byte {
	t.Helper()
	buf := &bytes.Buffer{}
	zw := zip.NewWriter(buf)

	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("create entry %s: %v", name, err)
		}
		if _, err := w.Write(content); err != nil {
			t.Fatalf("write entry %s: %v", name, err)
		}
	}
	for name, target := range symlinks {
		fh := &zip.FileHeader{Name: name, Method: zip.Store}
		fh.SetMode(fs.ModeSymlink | 0o777)
		w, err := zw.CreateHeader(fh)
		if err != nil {
			t.Fatalf("create symlink entry %s: %v", name, err)
		}
		if _, err := w.Write([]byte(target)); err != nil {
			t.Fatalf("write symlink target %s: %v", name, err)
		}
	}

	if err := zw.Close(); err != nil {
		t.Fatalf("zip Close: %v", err)
	}
	return buf.Bytes()
}
