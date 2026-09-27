package unidecode

import (
	"errors"
	"testing"
)

// buildTestTable returns a small Table whose blocks and Lookup rows
// are known values, independent of the real Unidecode pin, so
// encode/decode and Lookup can be tested without the real wheel.
func buildTestTable() *Table {
	return &Table{
		version: "9.9.9",
		sweepLo: 0,
		sweepHi: 0x10ffff,
		digest:  [32]byte{1, 2, 3},
		blocks: map[uint16][]*string{
			0x01: {new("a"), nil, new("")},
			0x30: {new("b")},
		},
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	want := buildTestTable()
	got, err := decode(encode(want))
	if err != nil {
		t.Fatalf("decode(encode(table)): %v", err)
	}
	if got.version != want.version {
		t.Errorf("version = %q, want %q", got.version, want.version)
	}
	if got.sweepLo != want.sweepLo || got.sweepHi != want.sweepHi {
		t.Errorf("sweep = [%d, %d], want [%d, %d]", got.sweepLo, got.sweepHi, want.sweepLo, want.sweepHi)
	}
	if got.digest != want.digest {
		t.Errorf("digest = %x, want %x", got.digest, want.digest)
	}
	if len(got.blocks) != len(want.blocks) {
		t.Fatalf("blocks = %d, want %d", len(got.blocks), len(want.blocks))
	}
	for section, wantEntries := range want.blocks {
		gotEntries, ok := got.blocks[section]
		if !ok {
			t.Fatalf("section %#x missing after round trip", section)
		}
		if len(gotEntries) != len(wantEntries) {
			t.Fatalf("section %#x = %d entries, want %d", section, len(gotEntries), len(wantEntries))
		}
		for i := range wantEntries {
			g, w := gotEntries[i], wantEntries[i]
			if (g == nil) != (w == nil) {
				t.Fatalf("section %#x entry %d = %v, want %v", section, i, g, w)
			}
			if g != nil && *g != *w {
				t.Fatalf("section %#x entry %d = %q, want %q", section, i, *g, *w)
			}
		}
	}
}

func TestDecodeTruncated(t *testing.T) {
	full := encode(buildTestTable())
	for cut := range full {
		if cut == 0 {
			continue
		}
		if _, err := decode(full[:cut]); err == nil {
			t.Fatalf("decode(truncated at %d) = nil error, want a truncation error", cut)
		}
	}
}

func TestDecodeTrailingBytes(t *testing.T) {
	full := append(encode(buildTestTable()), 0x00)
	if _, err := decode(full); err == nil {
		t.Fatal("decode(table with trailing byte) = nil error, want a trailing-byte error")
	}
}

func TestParseRefusesChecksumMismatch(t *testing.T) {
	if _, err := Parse(encode(buildTestTable())); !errors.Is(err, ErrTableChecksum) {
		t.Fatalf("Parse(unpinned table) error = %v, want ErrTableChecksum", err)
	}
}

func TestTableLookup(t *testing.T) {
	table := buildTestTable()

	tests := []struct {
		name     string
		r        rune
		wantRepl string
		wantOK   bool
	}{
		{"ASCII passthrough", 'A', "A", true},
		{"mapped entry", 0x0100, "a", true},
		{"explicit empty mapping", 0x0102, "", true},
		{"None entry", 0x0101, "", false},
		{"section with no module", 0x0200, "", false},
		{"position past block end", 0x3005, "", false},
		{"above the maximum codepoint", 0xf0000, "", false},
		{"at the maximum codepoint, unmapped section", unidecodeMaxCodepoint, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repl, ok := table.Lookup(tt.r)
			if repl != tt.wantRepl || ok != tt.wantOK {
				t.Errorf("Lookup(%#x) = %q, %v, want %q, %v", tt.r, repl, ok, tt.wantRepl, tt.wantOK)
			}
		})
	}
}
