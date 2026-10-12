package gate

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand/v2"
	"os"
	"path/filepath"
	"testing"

	"github.com/mewkiz/flac"
	"github.com/mewkiz/flac/frame"
	"github.com/mewkiz/flac/meta"
)

// blockSize is the samples per channel in every generated frame. A
// generated clip holds a whole number of them, because the encoder
// writes the smallest block it saw into STREAMINFO and a final short
// block under 16 samples would make that value invalid.
const blockSize = 1024

// clip describes one generated FLAC: PCM from a seeded generator, so a
// test can name the exact samples, and the tags and picture to carry.
type clip struct {
	rate     uint32
	bits     uint8
	channels int
	blocks   int // frames of blockSize samples per channel
	seed     uint64
	tags     [][2]string
	picture  int // bytes of each PICTURE block's data; 0 carries none
	pictures int // PICTURE blocks to carry when picture > 0; 0 means one
	samples  [][]int32
}

// pcm returns the clip's samples, one slice per channel, generating them
// from the seed on first use.
func (c *clip) pcm() [][]int32 {
	if c.samples != nil {
		return c.samples
	}
	if c.channels == 0 {
		c.channels = 2
	}
	rng := rand.New(rand.NewPCG(c.seed, c.seed^0x9e3779b97f4a7c15))
	limit := int32(1) << (c.bits - 1)
	c.samples = make([][]int32, c.channels)
	for ch := range c.samples {
		c.samples[ch] = make([]int32, c.blocks*blockSize)
		for i := range c.samples[ch] {
			c.samples[ch][i] = rng.Int32N(2*limit) - limit
		}
	}
	return c.samples
}

// audioMD5 is RFC 9639 section 8.2's checksum over samples, computed here
// from the sample slice alone: interleaved per sample, signed, little-endian,
// each sample sign-extended to whole bytes.
func audioMD5(bits uint8, samples [][]int32) []byte {
	h := md5.New()
	width := (int(bits) + 7) / 8
	if len(samples) == 0 {
		return h.Sum(nil)
	}
	for i := range samples[0] {
		for _, ch := range samples {
			v := uint32(ch[i])
			for b := range width {
				h.Write([]byte{byte(v >> (8 * b))})
			}
		}
	}
	return h.Sum(nil)
}

// encode writes the clip as a FLAC file and returns its bytes. The file
// goes through a real file so the encoder, given an io.WriteSeeker, fills
// STREAMINFO's MD5 and sample count on Close (encode.go).
func (c *clip) encode(t testing.TB) []byte {
	t.Helper()
	pcm := c.pcm()
	path := filepath.Join(t.TempDir(), "clip.flac")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	info := &meta.StreamInfo{
		BlockSizeMin: blockSize, BlockSizeMax: blockSize,
		SampleRate: c.rate, NChannels: uint8(c.channels), BitsPerSample: c.bits,
	}
	var blocks []*meta.Block
	if len(c.tags) > 0 {
		blocks = append(blocks, &meta.Block{
			Type: meta.TypeVorbisComment, Length: 1,
			Body: &meta.VorbisComment{Vendor: "gate-test", Tags: c.tags},
		})
	}
	if c.picture > 0 {
		for range max(c.pictures, 1) {
			blocks = append(blocks, &meta.Block{
				Type: meta.TypePicture, Length: 1,
				Body: &meta.Picture{Type: 3, MIME: "image/png", Data: make([]byte, c.picture)},
			})
		}
	}
	enc, err := flac.NewEncoder(f, info, blocks...)
	if err != nil {
		t.Fatal(err)
	}
	enc.EnablePredictionAnalysis(false) // keep subframes verbatim, so a test can place bytes
	for b := range c.blocks {
		fr := &frame.Frame{
			HasFixedBlockSize: true, BlockSize: blockSize, SampleRate: c.rate,
			Channels: channelLayout(c.channels), BitsPerSample: c.bits}
		for ch := range c.channels {
			fr.Subframes = append(fr.Subframes, &frame.Subframe{
				Pred:     frame.PredVerbatim,
				NSamples: blockSize,
				Samples:  pcm[ch][b*blockSize : (b+1)*blockSize],
			})
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
	return data
}

func channelLayout(n int) frame.Channels {
	if n == 1 {
		return frame.ChannelsMono
	}
	return frame.ChannelsLR
}

// cd returns a clean 16-bit, 44.1 kHz stereo clip of the given blocks.
func cd(seed uint64, blocks int, tags ...[2]string) *clip {
	return &clip{rate: 44100, bits: 16, channels: 2, blocks: blocks, seed: seed, tags: tags}
}

// streamInfoPayload is the offset of STREAMINFO's 34-byte payload in a file
// the encoder wrote with STREAMINFO first: "fLaC", then a 4-byte block
// header.
const streamInfoPayload = 8

// patchStreamInfo returns a copy of data with STREAMINFO's bits per sample,
// sample count, and MD5 rewritten, each only when its argument is non-nil.
// The eight-byte pack at payload [10:18] is rate (20 bits), channels-1 (3),
// bits-1 (5), samples (36), RFC 9639 section 8.2.
func patchStreamInfo(data []byte, bits, samples *uint64, md5sum []byte) []byte {
	out := append([]byte(nil), data...)
	pack := out[streamInfoPayload+10 : streamInfoPayload+18]
	v := binary.BigEndian.Uint64(pack)
	if bits != nil {
		v = v&^(0x1f<<36) | (*bits-1)<<36
	}
	if samples != nil {
		v = v&^0xfffffffff | *samples
	}
	binary.BigEndian.PutUint64(pack, v)
	if md5sum != nil {
		copy(out[streamInfoPayload+18:streamInfoPayload+34], md5sum)
	}
	return out
}

// withID3v2Prefix puts a minimal ID3v2.3 header (an empty tag) before data.
func withID3v2Prefix(data []byte) []byte {
	return append([]byte{'I', 'D', '3', 3, 0, 0, 0, 0, 0, 0}, data...)
}

// id3v1Record is a 128-byte ID3v1 tag: "TAG" and zeros.
func id3v1Record() []byte {
	rec := make([]byte, 128)
	copy(rec, "TAG")
	return rec
}

// chanceTAG returns a clean 16-bit stereo clip whose last 128 bytes begin
// "TAG" inside its final frame's right-channel samples. The layout, in a
// verbatim L/R frame: the 2-byte CRC-16 closes the file, and the right
// channel's samples are big-endian pairs before it, so the pair 62 from
// the last starts exactly 128 bytes from the end (2 + 2*(62+1) = 128).
func chanceTAG(t testing.TB, seed uint64) (data []byte, c *clip) {
	t.Helper()
	c = cd(seed, 2)
	pcm := c.pcm()
	right := pcm[1]
	last := len(right) - 1
	right[last-62] = 0x5441 // 'T' 'A'
	right[last-61] = 0x4700 // 'G' and a free byte
	data = c.encode(t)
	if got := string(data[len(data)-128 : len(data)-125]); got != "TAG" {
		t.Fatalf("fixture's last 128 bytes begin %q, want TAG", got)
	}
	return data, c
}

// flipByte returns data with the byte at off inverted.
func flipByte(data []byte, off int) []byte {
	out := append([]byte(nil), data...)
	out[off] ^= 0xff
	return out
}
