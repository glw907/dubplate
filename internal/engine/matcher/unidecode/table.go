package unidecode

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
)

// Table is a decoded or built Unidecode transliteration table: a block
// map keyed by codepoint>>8, each entry either a mapped string
// (possibly empty) or nil for "Unidecode has no mapping", the same
// None/empty distinction Unidecode's own table lookup makes.
type Table struct {
	version string
	sweepLo uint32
	sweepHi uint32
	digest  [32]byte
	blocks  map[uint16][]*string
}

// Parse verifies data's sha256 against the pinned table checksum, then
// decodes it. It refuses any table this package's own extraction did
// not produce, whether truncated, hand-edited, or from a different
// Unidecode release.
func Parse(data []byte) (*Table, error) {
	sum := sha256.Sum256(data)
	got := hex.EncodeToString(sum[:])
	if got != tableSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrTableChecksum, got, tableSHA256)
	}
	return decode(data)
}

// Lookup reproduces Unidecode's _get_repl_str for one codepoint: the
// ok return distinguishes "no mapping" (Unidecode's None) from an
// explicit empty-string mapping.
func (t *Table) Lookup(r rune) (repl string, ok bool) {
	if r < 0x80 {
		return string(r), true
	}
	if r > unidecodeMaxCodepoint {
		return "", false
	}
	section := uint16(r >> 8)
	position := int(r & 0xff)
	entries, ok := t.blocks[section]
	if !ok || position >= len(entries) {
		return "", false
	}
	entry := entries[position]
	if entry == nil {
		return "", false
	}
	return *entry, true
}

// encode writes t in this package's binary format (all multi-byte
// integers big-endian):
//
//	magic        4 bytes   "UNTB"
//	version_len  1 byte
//	version      version_len bytes, UTF-8
//	sweep_lo     4 bytes   uint32, first codepoint the digest covers
//	sweep_hi     4 bytes   uint32, last codepoint the digest covers (inclusive)
//	digest       32 bytes  sha256 over every (codepoint, replacement) pair in [sweep_lo, sweep_hi]
//	block_count  4 bytes   uint32
//	blocks       block_count repetitions of:
//	    section      2 bytes   uint16, codepoint >> 8
//	    entry_count  2 bytes   uint16, entries in this block
//	    entries      entry_count repetitions of:
//	        length   2 bytes   uint16, noneSentinel means "no mapping"
//	        bytes    length bytes, UTF-8 (omitted when length is noneSentinel)
//
// Blocks are written in ascending section order, so encode is
// deterministic regardless of the map's own iteration order.
func (t *Table) encode() []byte {
	var buf bytes.Buffer
	buf.WriteString("UNTB")

	versionBytes := []byte(t.version)
	buf.WriteByte(byte(len(versionBytes)))
	buf.Write(versionBytes)

	var u32 [4]byte
	binary.BigEndian.PutUint32(u32[:], t.sweepLo)
	buf.Write(u32[:])
	binary.BigEndian.PutUint32(u32[:], t.sweepHi)
	buf.Write(u32[:])
	buf.Write(t.digest[:])

	sections := slices.Sorted(maps.Keys(t.blocks))
	binary.BigEndian.PutUint32(u32[:], uint32(len(sections)))
	buf.Write(u32[:])

	var u16 [2]byte
	for _, section := range sections {
		entries := t.blocks[section]
		binary.BigEndian.PutUint16(u16[:], section)
		buf.Write(u16[:])
		binary.BigEndian.PutUint16(u16[:], uint16(len(entries)))
		buf.Write(u16[:])
		for _, entry := range entries {
			if entry == nil {
				binary.BigEndian.PutUint16(u16[:], noneSentinel)
				buf.Write(u16[:])
				continue
			}
			entryBytes := []byte(*entry)
			binary.BigEndian.PutUint16(u16[:], uint16(len(entryBytes)))
			buf.Write(u16[:])
			buf.Write(entryBytes)
		}
	}
	return buf.Bytes()
}

// tableCursor reads the table's fixed binary layout left to right,
// erroring on truncation rather than panicking on an out-of-range
// slice.
type tableCursor struct {
	data []byte
	pos  int
}

func (c *tableCursor) bytes(n int) ([]byte, error) {
	if c.pos+n > len(c.data) {
		return nil, fmt.Errorf("truncated at offset %d: need %d bytes, have %d", c.pos, n, len(c.data)-c.pos)
	}
	b := c.data[c.pos : c.pos+n]
	c.pos += n
	return b, nil
}

func (c *tableCursor) uint8() (uint8, error) {
	b, err := c.bytes(1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (c *tableCursor) uint16() (uint16, error) {
	b, err := c.bytes(2)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(b), nil
}

func (c *tableCursor) uint32() (uint32, error) {
	b, err := c.bytes(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(b), nil
}

// decode reads data in encode's format. It is called only after the
// caller has verified data's sha256 against the pinned table checksum,
// so its own errors surface only against synthetic bytes in tests.
func decode(data []byte) (*Table, error) {
	c := &tableCursor{data: data}

	magic, err := c.bytes(4)
	if err != nil {
		return nil, fmt.Errorf("read magic: %w", err)
	}
	if string(magic) != "UNTB" {
		return nil, fmt.Errorf("bad magic %q", magic)
	}

	versionLen, err := c.uint8()
	if err != nil {
		return nil, fmt.Errorf("read version length: %w", err)
	}
	versionBytes, err := c.bytes(int(versionLen))
	if err != nil {
		return nil, fmt.Errorf("read version: %w", err)
	}

	sweepLo, err := c.uint32()
	if err != nil {
		return nil, fmt.Errorf("read sweep lo: %w", err)
	}
	sweepHi, err := c.uint32()
	if err != nil {
		return nil, fmt.Errorf("read sweep hi: %w", err)
	}
	digestBytes, err := c.bytes(32)
	if err != nil {
		return nil, fmt.Errorf("read digest: %w", err)
	}

	blockCount, err := c.uint32()
	if err != nil {
		return nil, fmt.Errorf("read block count: %w", err)
	}

	blocks := make(map[uint16][]*string, blockCount)
	for i := range blockCount {
		section, err := c.uint16()
		if err != nil {
			return nil, fmt.Errorf("read block %d section: %w", i, err)
		}
		entryCount, err := c.uint16()
		if err != nil {
			return nil, fmt.Errorf("read block %d entry count: %w", i, err)
		}
		entries := make([]*string, entryCount)
		for j := range entries {
			length, err := c.uint16()
			if err != nil {
				return nil, fmt.Errorf("read block %d entry %d length: %w", i, j, err)
			}
			if length == noneSentinel {
				continue
			}
			entryBytes, err := c.bytes(int(length))
			if err != nil {
				return nil, fmt.Errorf("read block %d entry %d: %w", i, j, err)
			}
			entry := string(entryBytes)
			entries[j] = &entry
		}
		blocks[section] = entries
	}

	if c.pos != len(c.data) {
		return nil, fmt.Errorf("%d trailing byte(s) after last block", len(c.data)-c.pos)
	}

	var digest [32]byte
	copy(digest[:], digestBytes)
	return &Table{
		version: string(versionBytes),
		sweepLo: sweepLo,
		sweepHi: sweepHi,
		digest:  digest,
		blocks:  blocks,
	}, nil
}
