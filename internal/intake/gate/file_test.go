package gate

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"testing/iotest"

	"github.com/mewkiz/flac/frame"

	"github.com/glw907/dubplate/internal/engine/tagio"
	"github.com/glw907/dubplate/internal/reject"
)

// stage writes data under a fresh staging root and returns the staged file
// as the gate takes it: the root, a server-minted name, and the extension
// the member's path carried.
func stage(t *testing.T, data []byte) StagedFile {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f0"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = root.Close() })
	return StagedFile{Root: root, Name: "f0", Path: "t.flac"}
}

func check(t *testing.T, data []byte, p Policy, unit Unit) FileResult {
	t.Helper()
	res, err := CheckFile(t.Context(), stage(t, data), p, unit)
	if err != nil {
		t.Fatalf("CheckFile error = %v, want a verdict", err)
	}
	return res
}

func TestCheckFile_CleanFile(t *testing.T) {
	c := cd(1, 4, [2]string{"TRACKNUMBER", "3"}, [2]string{"DISCNUMBER", "2"})
	data := c.encode(t)
	res := check(t, data, testPolicy(), UnitAlbum)
	if res.Verdict != "" {
		t.Fatalf("verdict = %q, want a pass", res.Verdict)
	}
	wantKey := audioMD5(16, c.pcm())
	if !bytes.Equal(res.AudioKey, wantKey) {
		t.Errorf("audio key = %x, want %x (the test's own MD5 over the samples)", res.AudioKey, wantKey)
	}
	if !bytes.Equal(res.HeaderMD5, wantKey) {
		t.Errorf("header md5 = %x, want the audio key %x", res.HeaderMD5, wantKey)
	}
	if got, want := hex.EncodeToString(res.FileHash), sha256sum(t, data); got != want {
		t.Errorf("file hash = %s, want sha256sum's %s", got, want)
	}
	if res.SampleRate != 44100 || res.BitsPerSample != 16 || res.Channels != 2 || res.Samples != 4*blockSize {
		t.Errorf("stream = %d Hz, %d bit, %d ch, %d samples; want 44100, 16, 2, %d", res.SampleRate, res.BitsPerSample, res.Channels, res.Samples, 4*blockSize)
	}
	if res.Disc != 2 || res.Track != 3 {
		t.Errorf("position = disc %d track %d, want disc 2 track 3", res.Disc, res.Track)
	}
}

// sha256sum is the coreutils binary's digest of data, so the expected hash
// shares no code with the gate's.
func sha256sum(t *testing.T, data []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	out, err := exec.CommandContext(t.Context(), "sha256sum", path).Output()
	if err != nil {
		t.Fatalf("sha256sum: %v", err)
	}
	sum, _, _ := strings.Cut(string(out), " ")
	return sum
}

func TestCheckFile_RetaggedCopyKeepsItsAudioKey(t *testing.T) {
	a := cd(7, 2, [2]string{"TITLE", "one"})
	b := cd(7, 2, [2]string{"TITLE", "a longer second title"})
	ra, rb := check(t, a.encode(t), testPolicy(), UnitTrack), check(t, b.encode(t), testPolicy(), UnitTrack)
	if ra.Verdict != "" || rb.Verdict != "" {
		t.Fatalf("verdicts = %q, %q, want passes", ra.Verdict, rb.Verdict)
	}
	if !bytes.Equal(ra.AudioKey, rb.AudioKey) {
		t.Errorf("audio keys differ across a retag: %x vs %x", ra.AudioKey, rb.AudioKey)
	}
	if bytes.Equal(ra.FileHash, rb.FileHash) {
		t.Error("file hashes equal across a retag; the hash must cover every byte")
	}
}

func TestCheckFile_Verdicts(t *testing.T) {
	clean := cd(2, 4, [2]string{"TRACKNUMBER", "1"})
	cleanBytes := clean.encode(t)
	chance, _ := chanceTAG(t, 3)
	zeroBlocks := (&clip{rate: 44100, bits: 16, channels: 2, blocks: 0, seed: 4}).encode(t)
	// A zero-sample stream carries no frame, so the encoder leaves the
	// block sizes at zero, which STREAMINFO forbids; real encoders write
	// their nominal size.
	copy(zeroBlocks[streamInfoPayload:], []byte{0x10, 0x00, 0x10, 0x00})
	p := testPolicy()
	small := p
	small.MaxTagBytes, small.MaxPictureBytes = 64, 4096

	tests := []struct {
		name string
		data []byte
		ext  string
		p    Policy
		want reject.Code
	}{
		{"clean", cleanBytes, ".flac", p, ""},
		{"a clean file whose last 128 bytes begin TAG by chance", chance, ".flac", p, ""},
		{"a flipped frame byte", flipByte(cleanBytes, len(cleanBytes)/2), ".flac", p, reject.Corrupt},
		{"a truncation", cleanBytes[:len(cleanBytes)-100], ".flac", p, reject.Corrupt},
		{"a header md5 unequal to the decode", patchStreamInfo(cleanBytes, nil, nil, bytes.Repeat([]byte{0x11}, 16)), ".flac", p, reject.SourceDamaged},
		{"a sample count unequal to a nonzero NSamples", patchStreamInfo(cleanBytes, nil, new(uint64(4*blockSize+1)), nil), ".flac", p, reject.SourceDamaged},
		{"22.05 kHz", (&clip{rate: 22050, bits: 16, channels: 2, blocks: 2, seed: 5}).encode(t), ".flac", p, reject.BelowCDQuality},
		{"12 bit", (&clip{rate: 44100, bits: 12, channels: 2, blocks: 2, seed: 6}).encode(t), ".flac", p, reject.BelowCDQuality},
		{"18 bit", patchStreamInfo(cleanBytes, new(uint64(18)), nil, nil), ".flac", p, reject.UnsupportedFLAC},
		{"32 bit", patchStreamInfo(cleanBytes, new(uint64(32)), nil, nil), ".flac", p, reject.UnsupportedFLAC},
		{"20 bit", (&clip{rate: 44100, bits: 20, channels: 2, blocks: 2, seed: 8}).encode(t), ".flac", p, ""},
		{"24 bit", (&clip{rate: 96000, bits: 24, channels: 2, blocks: 2, seed: 9}).encode(t), ".flac", p, ""},
		{"zero samples", zeroBlocks, ".flac", p, reject.EmptyTrack},
		{"an ID3v2 prefix", withID3v2Prefix(cleanBytes), ".flac", p, reject.ID3Tags},
		{"a trailing ID3v1 record", append(append([]byte(nil), cleanBytes...), id3v1Record()...), ".flac", p, reject.ID3Tags},
		{"an all-zero header md5", patchStreamInfo(cleanBytes, nil, nil, make([]byte, 16)), ".flac", p, ""},
		// With no sample count and no header MD5 to catch them, frames that
		// disagree with STREAMINFO would otherwise pass the CD-quality floor.
		{"frames deeper than STREAMINFO says", patchStreamInfo((&clip{rate: 44100, bits: 24, channels: 2, blocks: 2, seed: 30}).encode(t), new(uint64(16)), new(uint64(0)), make([]byte, 16)), ".flac", p, reject.Corrupt},
		{"frames wider than STREAMINFO says", patchStreamInfo(patchChannels(cleanBytes, 1), nil, new(uint64(0)), make([]byte, 16)), ".flac", p, reject.Corrupt},
		{"frames smaller than STREAMINFO's minimum block size", patchStreamInfo(patchMinBlockSize(cleanBytes, 2*blockSize), nil, new(uint64(0)), make([]byte, 16)), ".flac", p, reject.Corrupt},
		{"a rate above 768 kHz", patchRate(cleanBytes, 1_000_000), ".flac", p, reject.UnsupportedFLAC},
		{"176.4 kHz, which the decoder has no test files for", patchRate(cleanBytes, 176400), ".flac", p, reject.UnsupportedFLAC},
		{"frames larger than STREAMINFO's maximum block size", patchStreamInfo(patchMaxBlockSize(cleanBytes, blockSize/2), nil, new(uint64(0)), make([]byte, 16)), ".flac", p, reject.Corrupt},
		{"an extension outside the allowlist", cleanBytes, ".mp3", p, reject.BadExtension},
		{"an empty allowlist", cleanBytes, ".flac", func() Policy { q := p; q.Allowlist = nil; return q }(), reject.BadExtension},
		{"content that is not FLAC", []byte("OggS\x00\x02 not flac at all"), ".flac", p, reject.BadExtension},
		{"a file too short to sniff", []byte("fL"), ".flac", p, reject.BadExtension},
		{"a header cut off after the magic", []byte("fLaC\x00"), ".flac", p, reject.Corrupt},
		{"a tag block over its cap", cd(10, 2, [2]string{"COMMENT", strings.Repeat("x", 200)}).encode(t), ".flac", small, reject.TagsTooLarge},
		{"a picture block over its cap", (&clip{rate: 44100, bits: 16, channels: 2, blocks: 2, seed: 11, picture: 5000}).encode(t), ".flac", p, reject.PictureTooLarge},
		{"a chain over four times the picture cap", (&clip{rate: 44100, bits: 16, channels: 2, blocks: 2, seed: 12, picture: 4000, pictures: 5}).encode(t), ".flac", p, reject.PictureTooLarge},
		{"a chain just under four times the picture cap", (&clip{rate: 44100, bits: 16, channels: 2, blocks: 2, seed: 13, picture: 4000, pictures: 3}).encode(t), ".flac", p, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := stage(t, tt.data)
			f.Path = "t" + tt.ext
			res, err := CheckFile(t.Context(), f, tt.p, UnitTrack)
			if err != nil {
				t.Fatalf("CheckFile error = %v, want a verdict", err)
			}
			if res.Verdict != tt.want {
				t.Errorf("verdict = %q, want %q", res.Verdict, tt.want)
			}
		})
	}
}

// TestCheckFile_Min* run the CD floor from the two settings: a policy that
// raises either bound refuses a file the defaults admit.
func TestCheckFile_FloorsComeFromPolicy(t *testing.T) {
	data := cd(14, 2).encode(t)
	for _, tt := range []struct {
		name  string
		rate  int
		depth int
		want  reject.Code
	}{
		{"defaults admit CD", 44100, 16, ""},
		{"a higher rate floor", 48000, 16, reject.BelowCDQuality},
		{"a higher depth floor", 44100, 24, reject.BelowCDQuality},
	} {
		t.Run(tt.name, func(t *testing.T) {
			p := testPolicy()
			p.MinSampleRate, p.MinBitsPerSample = tt.rate, tt.depth
			if got := check(t, data, p, UnitTrack).Verdict; got != tt.want {
				t.Errorf("verdict = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckFile_AllZeroHeaderMD5IsStoredNull(t *testing.T) {
	c := cd(15, 2)
	res := check(t, patchStreamInfo(c.encode(t), nil, nil, make([]byte, 16)), testPolicy(), UnitTrack)
	if res.Verdict != "" {
		t.Fatalf("verdict = %q, want a pass", res.Verdict)
	}
	if res.HeaderMD5 != nil {
		t.Errorf("header md5 = %x, want nil (an all-zero value is absent, never a value)", res.HeaderMD5)
	}
	if !bytes.Equal(res.AudioKey, audioMD5(16, c.pcm())) {
		t.Errorf("audio key = %x, want the MD5 of the samples", res.AudioKey)
	}
}

func TestCheckFile_Tags(t *testing.T) {
	tests := []struct {
		name      string
		tags      [][2]string
		unit      Unit
		want      reject.Code
		disc, trk int
	}{
		{"plain numbers", [][2]string{{"TRACKNUMBER", "4"}, {"DISCNUMBER", "2"}}, UnitAlbum, "", 2, 4},
		{"track of total", [][2]string{{"TRACKNUMBER", "3/12"}}, UnitAlbum, "", 1, 3},
		{"disc 0 reads as disc 1", [][2]string{{"TRACKNUMBER", "1"}, {"DISCNUMBER", "0"}}, UnitAlbum, "", 1, 1},
		{"a missing disc reads as disc 1", [][2]string{{"TRACKNUMBER", "1"}}, UnitAlbum, "", 1, 1},
		{"a negative disc reads as disc 1", [][2]string{{"TRACKNUMBER", "1"}, {"DISCNUMBER", "-2"}}, UnitAlbum, "", 1, 1},
		{"a vinyl side", [][2]string{{"TRACKNUMBER", "A1"}}, UnitAlbum, reject.MissingTrackNumber, 0, 0},
		{"no track", nil, UnitAlbum, reject.MissingTrackNumber, 0, 0},
		{"track 0", [][2]string{{"TRACKNUMBER", "0"}}, UnitAlbum, reject.MissingTrackNumber, 0, 0},
		{"a negative track", [][2]string{{"TRACKNUMBER", "-1"}}, UnitAlbum, reject.MissingTrackNumber, 0, 0},
		{"separate tracks need no number", nil, UnitTrack, "", 0, 0},
		{"separate tracks ignore a vinyl side", [][2]string{{"TRACKNUMBER", "A1"}}, UnitTrack, "", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := check(t, cd(20, 2, tt.tags...).encode(t), testPolicy(), tt.unit)
			if res.Verdict != tt.want {
				t.Errorf("verdict = %q, want %q", res.Verdict, tt.want)
			}
			if tt.want == "" && tt.unit == UnitAlbum && (res.Disc != tt.disc || res.Track != tt.trk) {
				t.Errorf("position = disc %d track %d, want disc %d track %d", res.Disc, res.Track, tt.disc, tt.trk)
			}
		})
	}
}

func TestCheckFile_RecordsWhatItLearnedBeforeARefusal(t *testing.T) {
	res := check(t, patchStreamInfo(cd(21, 2).encode(t), nil, nil, bytes.Repeat([]byte{0x22}, 16)), testPolicy(), UnitTrack)
	if res.Verdict != reject.SourceDamaged {
		t.Fatalf("verdict = %q, want source-damaged", res.Verdict)
	}
	if res.SampleRate != 44100 || res.AudioKey == nil || res.FileHash == nil {
		t.Errorf("a source-damaged file keeps its stream, audio key, and hash; got rate %d key %x hash %x", res.SampleRate, res.AudioKey, res.FileHash)
	}
}

func TestCheckFile_HasPicture(t *testing.T) {
	with := check(t, (&clip{rate: 44100, bits: 16, channels: 2, blocks: 2, seed: 22, picture: 100}).encode(t), testPolicy(), UnitTrack)
	without := check(t, cd(23, 2).encode(t), testPolicy(), UnitTrack)
	if !with.HasPicture || without.HasPicture {
		t.Errorf("HasPicture = %t with a picture, %t without; want true, false", with.HasPicture, without.HasPicture)
	}
}

// TestCheckFile_AFlippedByteIsNotATruncation holds the corrupt verdict to
// the decoder, not to a size mismatch: the flipped file has the right size.
func TestCheckFile_AFlippedByteIsNotATruncation(t *testing.T) {
	data := cd(24, 4).encode(t)
	flipped := flipByte(data, len(data)/2)
	if len(flipped) != len(data) {
		t.Fatal("fixture changed the file's length")
	}
	if got := check(t, flipped, testPolicy(), UnitTrack).Verdict; got != reject.Corrupt {
		t.Errorf("verdict = %q, want corrupt", got)
	}
}

func TestCheckFile_CancelledContext(t *testing.T) {
	f := stage(t, cd(25, 4).encode(t))
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := CheckFile(ctx, f, testPolicy(), UnitTrack)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want context.Canceled", err)
	}
	if errors.Is(err, ErrCouldNotRun) {
		t.Errorf("a cancelled context is neither a verdict nor a failed check, got %v", err)
	}
	// The same file with a live context passes, so the cancellation is
	// what stopped the first run.
	if _, err := CheckFile(t.Context(), f, testPolicy(), UnitTrack); err != nil {
		t.Errorf("live context error = %v, want none", err)
	}
}

func TestCheckFile_AMissingFileCouldNotRun(t *testing.T) {
	f := stage(t, cd(26, 2).encode(t))
	f.Name = "nowhere"
	_, err := CheckFile(t.Context(), f, testPolicy(), UnitTrack)
	if !errors.Is(err, ErrCouldNotRun) || !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("error = %v, want ErrCouldNotRun wrapping fs.ErrNotExist", err)
	}
}

// TestDecode_ReadErrorIsNeverAVerdict is P19: a reader that fails with EIO
// partway through the stream makes the check "could not run" whatever the
// decoder said, and the same bytes without the fault decode clean.
func TestDecode_ReadErrorIsNeverAVerdict(t *testing.T) {
	data := cd(30, 4).encode(t)
	size := int64(len(data))
	faulty := func(after int) io.Reader {
		return io.MultiReader(bytes.NewReader(data[:after]), iotest.ErrReader(syscall.EIO))
	}

	if _, err := decode(t.Context(), bytes.NewReader(data), size, cdStream, noLimits); err != nil {
		t.Fatalf("decode without the fault: %v, want none", err)
	}
	for _, after := range []int{0, 30, len(data) / 2, len(data) - 1} {
		t.Run(fmt.Sprintf("EIO after %d bytes", after), func(t *testing.T) {
			_, err := decode(t.Context(), faulty(after), size, cdStream, noLimits)
			if !errors.Is(err, ErrCouldNotRun) || !errors.Is(err, syscall.EIO) {
				t.Errorf("error = %v, want ErrCouldNotRun wrapping EIO", err)
			}
			if errors.Is(err, errUndecodable) {
				t.Errorf("error = %v also reads as the file's fault", err)
			}
		})
	}

	t.Run("the same cut, ended cleanly, is the file's fault", func(t *testing.T) {
		_, err := decode(t.Context(), bytes.NewReader(data[:len(data)/2]), size, cdStream, noLimits)
		if !errors.Is(err, errUndecodable) || errors.Is(err, ErrCouldNotRun) {
			t.Errorf("error = %v, want errUndecodable alone", err)
		}
	})
}

func TestVerdictOf(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		want   reject.Code
		wantOK bool
	}{
		{"malformed", tagio.ErrMalformedContainer, reject.Corrupt, true},
		{"malformed, wrapped", fmt.Errorf("read: %w", tagio.ErrMalformedContainer), reject.Corrupt, true},
		{"unsupported", tagio.ErrUnsupportedContainer, reject.BadExtension, true},
		{"tags too large", tagio.ErrTagsTooLarge, reject.TagsTooLarge, true},
		{"picture too large", tagio.ErrPictureTooLarge, reject.PictureTooLarge, true},
		{"a path error", &fs.PathError{Op: "open", Path: "x", Err: syscall.EIO}, "", false},
		{"an unrelated error", errors.New("boom"), "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := verdictOf(tt.err)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("verdictOf(%v) = %q, %t; want %q, %t", tt.err, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

// BenchmarkCheckFile is P13's measurement: one pass of the gate's decode,
// SHA-256, and audio key over a 16-bit, 44.1 kHz stereo FLAC of about
// 200 MB. The frames are one generated clip's, repeated, with STREAMINFO's
// count and MD5 zeroed (unknown) so the whole file decodes and passes.
func BenchmarkCheckFile(b *testing.B) {
	c := cd(40, 1500)
	data := c.encode(b)
	path := filepath.Join(b.TempDir(), "src.flac")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		b.Fatal(err)
	}
	ct, err := tagio.Parse(path)
	if err != nil {
		b.Fatal(err)
	}
	head, frames := data[:ct.FrameOffset], data[ct.FrameOffset:]
	big := append([]byte(nil), head...)
	for len(big) < 200<<20 {
		big = append(big, frames...)
	}
	big = patchStreamInfo(big, nil, new(uint64(0)), make([]byte, 16))
	dir := b.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f0"), big, 0o600); err != nil {
		b.Fatal(err)
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		b.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f := StagedFile{Root: root, Name: "f0", Path: "t.flac"}
	b.SetBytes(int64(len(big)))
	b.ResetTimer()
	for range b.N {
		res, err := CheckFile(b.Context(), f, testPolicy(), UnitTrack)
		if err != nil || res.Verdict != "" {
			b.Fatalf("verdict %q, error %v", res.Verdict, err)
		}
	}
}

// cdStream is the STREAMINFO of the generated CD-quality clips.
var cdStream = tagio.StreamInfo{SampleRate: 44100, Channels: 2, BitsPerSample: 16, MaxBlockSize: blockSize}

// noLimits is a decode bound no test file reaches.
var noLimits = decodeLimits{maxSamples: 1 << 40, pcmAllowance: 1 << 50, pcmPerByte: 1 << 20}

// countingReader counts the bytes read through it.
type countingReader struct {
	r io.Reader
	n int64
}

func (c *countingReader) Read(p []byte) (int, error) {
	n, err := c.r.Read(p)
	c.n += int64(n)
	return n, err
}

// TestDecode_StopsAtTheLimitsOnTheFrame holds the three bounds on the work a
// member's bytes can cost: a stream longer than its own sample count, one
// longer than any track can be, and one that decodes to more PCM than its
// size and the allowance explain each end the decode at the frame that
// crosses the bound, instead of running to the end of the file.
func TestDecode_StopsAtTheLimitsOnTheFrame(t *testing.T) {
	data := cd(40, 40).encode(t)
	tests := []struct {
		name      string
		totalSamp uint64
		lim       decodeLimits
	}{
		{"past the declared sample count", 2 * blockSize, noLimits},
		{"past the longest track", 0, decodeLimits{maxSamples: 2 * blockSize, pcmAllowance: 1 << 50, pcmPerByte: 1 << 20}},
		{"past the decoded work its size explains", 0, decodeLimits{maxSamples: 1 << 40, pcmAllowance: 20_000, pcmPerByte: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			si := cdStream
			si.TotalSamples = tt.totalSamp
			cr := &countingReader{r: bytes.NewReader(data)}
			_, err := decode(t.Context(), cr, int64(len(data)), si, tt.lim)
			if !errors.Is(err, errOverrun) {
				t.Fatalf("decode error = %v, want errOverrun", err)
			}
			if cr.n > int64(len(data))/4 {
				t.Errorf("decode read %d of %d bytes, want it to stop within the first quarter", cr.n, len(data))
			}
		})
	}
}

// TestSubframesAgree holds the check that turns a decoder panic a member can
// trigger into a verdict: mewkiz/flac appends a subframe's warm-up samples
// even when its order exceeds the block size, and Frame.Hash then indexes
// past the shorter subframe.
func TestSubframesAgree(t *testing.T) {
	sub := func(n int) *frame.Subframe { return &frame.Subframe{Samples: make([]int32, n)} }
	frameOf := func(block uint16, subs ...*frame.Subframe) *frame.Frame {
		fr := &frame.Frame{Subframes: subs}
		fr.BlockSize = block
		return fr
	}
	tests := []struct {
		name string
		fr   *frame.Frame
		ok   bool
	}{
		{"every subframe holds the block", frameOf(4, sub(4), sub(4)), true},
		{"a subframe longer than the block", frameOf(1, sub(32), sub(1)), false},
		{"a subframe shorter than the block", frameOf(4, sub(4), sub(3)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := subframesAgree(tt.fr)
			if (err == nil) != tt.ok || (err != nil && !errors.Is(err, errUndecodable)) {
				t.Errorf("subframesAgree = %v, want ok %t (an errUndecodable when not)", err, tt.ok)
			}
		})
	}
}
