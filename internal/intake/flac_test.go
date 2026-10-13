package intake

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// flacBlock is the samples per channel in every generated frame.
const flacBlock = 1024

// flacSpec describes one generated stereo FLAC: verbatim frames of
// seeded PCM, and the tags and picture it carries.
type flacSpec struct {
	rate    uint32
	bits    uint8
	blocks  int
	seed    uint64
	tags    [][2]string
	picture int // bytes of one PICTURE block's data; 0 carries none
}

// cdTrack is a clean 16-bit, 44.1 kHz track numbered n, from seed.
func cdTrack(seed uint64, n int) flacSpec {
	return flacSpec{rate: 44100, bits: 16, blocks: 4, seed: seed, tags: [][2]string{{"TRACKNUMBER", strconv.Itoa(n)}}}
}

// encode writes s as a FLAC file through a real file, so the encoder,
// given an io.WriteSeeker, fills STREAMINFO's MD5 and sample count on
// Close, and returns its bytes.
func (s flacSpec) encode(t testing.TB) []byte {
	t.Helper()
	rng := rand.New(rand.NewPCG(s.seed, s.seed^0x9e3779b97f4a7c15))
	limit := int32(1) << (s.bits - 1)
	pcm := [2][]int32{make([]int32, s.blocks*flacBlock), make([]int32, s.blocks*flacBlock)}
	for ch := range pcm {
		for i := range pcm[ch] {
			pcm[ch][i] = rng.Int32N(2*limit) - limit
		}
	}

	path := filepath.Join(t.TempDir(), "track.flac")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	var blocks []*meta.Block
	if len(s.tags) > 0 {
		blocks = append(blocks, &meta.Block{Type: meta.TypeVorbisComment, Length: 1,
			Body: &meta.VorbisComment{Vendor: "intake-test", Tags: s.tags}})
	}
	if s.picture > 0 {
		blocks = append(blocks, &meta.Block{Type: meta.TypePicture, Length: 1,
			Body: &meta.Picture{Type: 3, MIME: "image/png", Data: make([]byte, s.picture)}})
	}
	info := &meta.StreamInfo{BlockSizeMin: flacBlock, BlockSizeMax: flacBlock, SampleRate: s.rate, NChannels: 2, BitsPerSample: s.bits}
	enc, err := flac.NewEncoder(f, info, blocks...)
	if err != nil {
		t.Fatal(err)
	}
	enc.EnablePredictionAnalysis(false)
	for b := range s.blocks {
		fr := &frame.Frame{HasFixedBlockSize: true, BlockSize: flacBlock, SampleRate: s.rate,
			Channels: frame.ChannelsLR, BitsPerSample: s.bits}
		for ch := range pcm {
			fr.Subframes = append(fr.Subframes, &frame.Subframe{Pred: frame.PredVerbatim, NSamples: flacBlock,
				Samples: pcm[ch][b*flacBlock : (b+1)*flacBlock]})
		}
		if err := enc.WriteFrame(fr); err != nil {
			t.Fatal(err)
		}
	}
	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.blocks == 0 {
		// A stream with no frame leaves the block sizes zero, which
		// STREAMINFO forbids; real encoders write their nominal size.
		copy(data[8:], []byte{0x10, 0x00, 0x10, 0x00})
	}
	return data
}

// withStreamInfoBits returns a copy of data with STREAMINFO's bits per
// sample rewritten. The eight bytes at payload [10:18] pack the rate (20
// bits), channels-1 (3), bits-1 (5), and samples (36).
func withStreamInfoBits(data []byte, bits uint64) []byte {
	out := bytes.Clone(data)
	pack := out[18:26]
	v := binary.BigEndian.Uint64(pack)
	binary.BigEndian.PutUint64(pack, v&^(0x1f<<36)|(bits-1)<<36)
	return out
}

// withHeaderMD5 returns a copy of data with STREAMINFO's MD5 replaced.
func withHeaderMD5(data, sum []byte) []byte {
	out := bytes.Clone(data)
	copy(out[26:42], sum)
	return out
}

// flipped returns data with its middle byte inverted, inside a frame.
func flipped(data []byte) []byte {
	out := bytes.Clone(data)
	out[len(out)/2] ^= 0xff
	return out
}

// zipMember is one entry of a generated archive.
type zipMember struct {
	name string
	data []byte
}

// zipOfMembers stores each member uncompressed in a new archive.
func zipOfMembers(t testing.TB, members ...zipMember) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, m := range members {
		f, err := w.CreateHeader(&zip.FileHeader{Name: m.name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := f.Write(m.data); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
