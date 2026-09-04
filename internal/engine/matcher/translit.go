package matcher

import (
	_ "embed"
	"encoding/binary"
	"fmt"
	"strings"
	"sync"
)

//go:embed unidecode-1.4.0.tbl
var unidecodeTableData []byte

// UnidecodeVersion is the pinned unidecode release
// tools/unidecode-table/generate.py ran against to produce the
// embedded table. TestUnidecodeVersion checks this constant against
// the table's own version field so the two cannot drift; the
// comparison against corpus/manifest.json's unidecode field is
// conform-side (Task 6), since this package reads no corpus file.
const UnidecodeVersion = "1.4.0"

// unidecodeMaxCodepoint is unidecode's own cutoff
// (unidecode/__init__.py: "codepoint > 0xeffff: no data on characters
// in Private Use Area and above").
const unidecodeMaxCodepoint = 0xeffff

// unidecodeNoneSentinel marks a table entry as unmapped. No shipped
// entry is anywhere near this long (the longest is 13 bytes), so the
// length field is free to reuse as the marker.
const unidecodeNoneSentinel = 0xffff

// unidecodeTable is the decoded embedded table: a block map keyed by
// codepoint>>8, each entry either a mapped string (possibly empty) or
// nil for "unidecode has no mapping", the same None/empty distinction
// unidecode/__init__.py's own table lookup makes.
type unidecodeTable struct {
	version string
	sweepLo uint32
	sweepHi uint32
	digest  [32]byte
	blocks  map[uint16][]*string
}

// loadUnidecodeTable decodes the embedded table once. The embedded
// bytes are generated and committed by this package's own build step
// (tools/unidecode-table/generate.py), never user input, so a parse
// failure is a corrupted build artifact and panics rather than
// threading an error through every caller of transliterateRune.
var loadUnidecodeTable = sync.OnceValue(func() *unidecodeTable {
	table, err := parseUnidecodeTable(unidecodeTableData)
	if err != nil {
		panic(fmt.Sprintf("matcher: embedded unidecode table is corrupt: %v", err))
	}
	return table
})

// tableCursor reads the table's fixed binary layout (translit.go's
// generator counterpart, tools/unidecode-table/generate.py, documents
// the format) left to right, erroring on truncation rather than
// panicking on an out-of-range slice.
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

func parseUnidecodeTable(data []byte) (*unidecodeTable, error) {
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
			if length == unidecodeNoneSentinel {
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
	return &unidecodeTable{
		version: string(versionBytes),
		sweepLo: sweepLo,
		sweepHi: sweepHi,
		digest:  digest,
		blocks:  blocks,
	}, nil
}

// transliterateRune reproduces unidecode's _get_repl_str for one
// codepoint (unidecode/__init__.py): the ok return distinguishes "no
// mapping" (unidecode's None) from an explicit empty-string mapping.
// [Transliterate] folds both into "", the errors="ignore" behavior
// beets' own `from unidecode import unidecode` call uses
// (autotag/distance.py:9,55-56); the distinction only matters to
// translit_test.go's full-sweep fidelity proof.
func transliterateRune(r rune) (repl string, ok bool) {
	if r < 0x80 {
		// unidecode/__init__.py: codepoints below 0x80 are "already
		// ASCII" and pass through unchanged, without consulting any
		// table.
		return string(r), true
	}
	if r < 0 || r > unidecodeMaxCodepoint {
		return "", false
	}
	table := loadUnidecodeTable()
	section := uint16(r >> 8)
	position := int(r & 0xff)
	entries, ok := table.blocks[section]
	if !ok || position >= len(entries) {
		// unidecode ships no data module for this section (or, within
		// a shipped section, no entry reaches this far), the same gap
		// unidecode/__init__.py's own ImportError catch and length
		// check leave as None.
		return "", false
	}
	entry := entries[position]
	if entry == nil {
		return "", false
	}
	return *entry, true
}

// Transliterate reproduces `unidecode(s)`'s default errors="ignore"
// behavior (autotag/distance.py:9,55-56): every rune maps to its
// ASCII approximation, and a rune with no mapping contributes nothing.
func Transliterate(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		repl, _ := transliterateRune(r)
		b.WriteString(repl)
	}
	return b.String()
}

// Levenshtein is jellyfish.levenshtein_distance's plain-unit-cost edit
// distance over Unicode code points: the minimum number of single-rune
// insertions, deletions, and substitutions turning a into b.
func Levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}

	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	curr := make([]int, len(rb)+1)
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}
