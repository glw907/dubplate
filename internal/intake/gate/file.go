package gate

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"

	"github.com/glw907/dubplate/internal/engine/tagio"
	"github.com/glw907/dubplate/internal/reject"
)

// StagedFile is one received audio file, named the way staging keeps it.
type StagedFile struct {
	// Root is the staging directory's root. The gate opens Name through it
	// and hands tagio the joined path, a name no member chooses.
	Root *os.Root
	// Name is the server-minted path of the file under Root.
	Name string
	// Path is the member's relative path; only its extension is read.
	Path string
}

// FileResult is what the gate learned about one file. Verdict is empty for a
// pass and otherwise the first failing check's code; the fields filled
// before that check stay set, so a refused file still records its stream and
// hashes. A nil HeaderMD5 is the stream header's all-zero value, which means
// absent.
type FileResult struct {
	Verdict reject.Code

	SampleRate, BitsPerSample, Channels int
	Samples                             uint64
	AudioKey, HeaderMD5                 []byte // 16 bytes, or nil
	FileHash                            []byte // 32 bytes, or nil
	HasPicture                          bool

	// Disc, Track, TrackTotal, and DiscTotal are read only for an album.
	// Disc is at least 1; a total of 0 is one the tags do not carry.
	Disc, Track, TrackTotal, DiscTotal int
}

// errUndecodable marks a stream the decoder refused while every read
// succeeded: the file's fault, never the server's.
var errUndecodable = errors.New("not a decodable FLAC stream")

// errOverrun says a stream ran past its own sample count or past the longest
// track the gate decodes.
var errOverrun = errors.New("stream longer than it declares or the gate allows")

// CheckFile runs a file's checks in order and stops at the first that
// fails: the container and format, the metadata bounds, the stream format,
// one pass of decode and hashing, and, for an album, the disc and track
// tags. A separate track's tags only prefill a form, so they are never read.
//
// A non-nil error is not a verdict. [ErrCouldNotRun] says the server failed
// to read or run; ctx's own error says the check was cancelled, which it
// notices between frames.
func CheckFile(ctx context.Context, f StagedFile, p Policy, unit Unit) (FileResult, error) {
	var res FileResult
	refuse := func(code reject.Code) (FileResult, error) {
		res.Verdict = code
		return res, nil
	}

	file, err := f.Root.Open(f.Name)
	if err != nil {
		return res, couldNotRun("open", err)
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil {
		return res, couldNotRun("stat", err)
	}
	size := info.Size()

	// Step 1. The ID3 test comes before the sniff, which would misreport a
	// prefixed FLAC as some other format.
	var head [3]byte
	if _, err := io.ReadFull(file, head[:]); err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return res, couldNotRun("read", err)
	}
	if string(head[:]) == "ID3" {
		return refuse(reject.ID3Tags)
	}
	row, ok := p.admitted(extension(f.Path))
	if !ok {
		return refuse(reject.BadExtension)
	}
	path := filepath.Join(f.Root.Name(), f.Name)
	format, err := tagio.Sniff(path)
	if err != nil {
		return verdictOrError(res, err)
	}
	if format != row.Format {
		return refuse(reject.BadExtension)
	}

	// Step 2.
	c, err := tagio.ParseBounded(path, tagio.Bounds{TagBytes: p.MaxTagBytes, PictureBytes: p.MaxPictureBytes})
	if err != nil {
		return verdictOrError(res, err)
	}
	si := c.StreamInfo
	res.SampleRate, res.BitsPerSample, res.Channels = int(si.SampleRate), int(si.BitsPerSample), int(si.Channels)
	res.HasPicture = len(c.Pictures) > 0

	// Step 3.
	if res.SampleRate < p.MinSampleRate || res.BitsPerSample < p.MinBitsPerSample {
		return refuse(reject.BelowCDQuality)
	}
	if !slices.Contains([]int{16, 20, 24}, res.BitsPerSample) {
		return refuse(reject.UnsupportedFLAC)
	}

	// Step 4.
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return res, couldNotRun("rewind", err)
	}
	d, err := decode(ctx, file, size, si, uint64(si.SampleRate)*maxTrackSeconds)
	switch {
	case errors.Is(err, errOverrun):
		return refuse(reject.SourceDamaged)
	case errors.Is(err, errUndecodable):
		tagged, err := endsWithID3v1(file, size)
		if err != nil {
			return res, err
		}
		if tagged {
			return refuse(reject.ID3Tags)
		}
		return refuse(reject.Corrupt)
	case err != nil:
		return res, err
	}
	res.Samples, res.AudioKey, res.FileHash = d.samples, d.audioKey, d.fileHash
	if !isZero(si.MD5[:]) {
		res.HeaderMD5 = si.MD5[:]
	}
	switch {
	case d.samples == 0:
		return refuse(reject.EmptyTrack)
	case res.HeaderMD5 != nil && !bytes.Equal(res.HeaderMD5, d.audioKey):
		return refuse(reject.SourceDamaged)
	case si.TotalSamples != 0 && si.TotalSamples != d.samples:
		return refuse(reject.SourceDamaged)
	}

	// Step 5.
	if unit != UnitAlbum {
		return res, nil
	}
	ts, err := tagio.Read(path)
	if err != nil {
		return verdictOrError(res, err)
	}
	res.Disc, res.Track = max(1, deref(ts.Disc)), deref(ts.Track)
	res.TrackTotal, res.DiscTotal = max(0, deref(ts.Tracktotal)), max(0, deref(ts.Disctotal))
	if res.Track <= 0 {
		return refuse(reject.MissingTrackNumber)
	}
	return res, nil
}

func deref(n *int) int {
	if n == nil {
		return 0
	}
	return *n
}

func isZero(b []byte) bool {
	return !slices.ContainsFunc(b, func(c byte) bool { return c != 0 })
}

// verdictOrError turns a tagio error into the file's verdict when it is
// one, and into a could-not-run error when it is not.
func verdictOrError(res FileResult, err error) (FileResult, error) {
	code, ok := verdictOf(err)
	if !ok {
		return res, couldNotRun("read tags", err)
	}
	res.Verdict = code
	return res, nil
}

// verdictOf classifies a tagio error: the four sentinels are verdicts about
// the file, and anything else, a wrapped *fs.PathError among them, says the
// server could not read it.
func verdictOf(err error) (reject.Code, bool) {
	switch {
	case errors.Is(err, tagio.ErrMalformedContainer):
		return reject.Corrupt, true
	case errors.Is(err, tagio.ErrUnsupportedContainer):
		return reject.BadExtension, true
	case errors.Is(err, tagio.ErrTagsTooLarge):
		return reject.TagsTooLarge, true
	case errors.Is(err, tagio.ErrPictureTooLarge):
		return reject.PictureTooLarge, true
	}
	return "", false
}

func couldNotRun(op string, err error) error {
	return fmt.Errorf("%w: %s: %w", ErrCouldNotRun, op, err)
}

// decoded is the product of one pass over a file.
type decoded struct {
	samples            uint64
	audioKey, fileHash []byte
}

// maxTrackSeconds bounds the audio one file may hold. No track on an album
// runs a day, and the bound caps the decode and MD5 work a stream of
// near-empty frames can cost at a fixed multiple of its declared format.
const maxTrackSeconds = 24 * 60 * 60

// decode makes the one pass over r: a SHA-256 over every byte and a decode
// of every frame into the audio key, checking ctx between frames.
//
// mewkiz/flac checks no frame against STREAMINFO, so decode does: a frame
// whose depth, channel count, rate, or block size disagrees with it is
// undecodable, which keeps Frame.Hash on the depths it implements. The
// pass ends with [errOverrun] at the frame that takes the samples past
// STREAMINFO's own count, when it states one, or past maxSamples.
//
// Neither decoder's error says whose fault it was (mewkiz/flac returns its
// bit reader's error raw), so decode classifies by its own I/O: a read that
// failed with anything but io.EOF is [ErrCouldNotRun] whatever the decoder
// then said, and any other failure, or a stream whose bytes do not add up
// to size, is [errUndecodable].
func decode(ctx context.Context, r io.Reader, size int64, si tagio.StreamInfo, maxSamples uint64) (decoded, error) {
	rec := &recorder{r: r}
	sum := sha256.New()
	stream, err := flac.New(io.TeeReader(rec, sum))
	if err != nil {
		return decoded{}, rec.fault(err)
	}
	key := md5.New()
	var samples uint64
	for {
		if err := ctx.Err(); err != nil {
			return decoded{}, err
		}
		fr, err := stream.ParseNext()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return decoded{}, rec.fault(err)
		}
		if err := disagrees(fr, si); err != nil {
			return decoded{}, err
		}
		samples += uint64(fr.BlockSize)
		if samples > maxSamples || (si.TotalSamples != 0 && samples > si.TotalSamples) {
			return decoded{}, fmt.Errorf("%w: %d samples", errOverrun, samples)
		}
		fr.Hash(key)
	}
	if rec.err != nil {
		return decoded{}, rec.fault(nil)
	}
	if rec.n != size {
		return decoded{}, fmt.Errorf("%w: hashed %d of %d bytes", errUndecodable, rec.n, size)
	}
	return decoded{samples: samples, audioKey: key.Sum(nil), fileHash: sum.Sum(nil)}, nil
}

// disagrees reports how fr contradicts STREAMINFO, as an [errUndecodable]
// error, or nil when it does not. A frame rate of 0 is the header's "take it
// from STREAMINFO".
func disagrees(fr *frame.Frame, si tagio.StreamInfo) error {
	switch {
	case fr.BitsPerSample != si.BitsPerSample:
		return fmt.Errorf("%w: frame has %d bits per sample, STREAMINFO %d", errUndecodable, fr.BitsPerSample, si.BitsPerSample)
	case len(fr.Subframes) != int(si.Channels):
		return fmt.Errorf("%w: frame has %d channels, STREAMINFO %d", errUndecodable, len(fr.Subframes), si.Channels)
	case fr.SampleRate != 0 && fr.SampleRate != si.SampleRate:
		return fmt.Errorf("%w: frame has rate %d, STREAMINFO %d", errUndecodable, fr.SampleRate, si.SampleRate)
	case si.MaxBlockSize != 0 && fr.BlockSize > si.MaxBlockSize:
		return fmt.Errorf("%w: frame has %d samples, STREAMINFO's maximum is %d", errUndecodable, fr.BlockSize, si.MaxBlockSize)
	}
	return nil
}

// recorder counts the bytes it passes and remembers the first read error
// that is not io.EOF.
type recorder struct {
	r   io.Reader
	n   int64
	err error
}

func (r *recorder) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	r.n += int64(n)
	if err != nil && !errors.Is(err, io.EOF) && r.err == nil {
		r.err = err
	}
	return n, err
}

// fault classifies a decoder failure by what the reads did.
func (r *recorder) fault(decodeErr error) error {
	if r.err != nil {
		return couldNotRun("read staged file", r.err)
	}
	return fmt.Errorf("%w: %w", errUndecodable, decodeErr)
}

// endsWithID3v1 reports whether the last 128 bytes of the file begin with
// the ID3v1 identifier. mutagen finds the record the same way (seek 128
// from the end, look for "TAG"), and a decode error is the only reason to
// look, so a valid FLAC whose last bytes happen to spell TAG still passes.
func endsWithID3v1(file *os.File, size int64) (bool, error) {
	if size < 128 {
		return false, nil
	}
	var id [3]byte
	if _, err := file.ReadAt(id[:], size-128); err != nil {
		return false, couldNotRun("read trailer", err)
	}
	return string(id[:]) == "TAG", nil
}
