package unidecode

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

// wantTableSHA256 is the sha256 tools/unidecode-table/generate.py
// reported for the table it wrote from the pinned Unidecode 1.4.0
// release, computed independently of this package's own tableSHA256
// pin: a build that happens to agree with a wrong pin still fails this
// test.
const wantTableSHA256 = "c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85"

const (
	wantBlockCount = 190
	wantEntryCount = 48597
)

func TestBuildIsByteIdenticalToThePin(t *testing.T) {
	wheel := loadCachedWheel(t)

	built, err := Build(wheel)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	sum := sha256.Sum256(built)
	if got := hex.EncodeToString(sum[:]); got != wantTableSHA256 {
		t.Fatalf("Build sha256 = %s, want %s", got, wantTableSHA256)
	}

	table, err := decode(built)
	if err != nil {
		t.Fatalf("decode the built table: %v", err)
	}
	if len(table.blocks) != wantBlockCount {
		t.Fatalf("built table carries %d blocks, want %d", len(table.blocks), wantBlockCount)
	}
	var totalEntries int
	for _, entries := range table.blocks {
		totalEntries += len(entries)
	}
	if totalEntries != wantEntryCount {
		t.Fatalf("built table carries %d entries, want %d", totalEntries, wantEntryCount)
	}
}

// TestBuildParseLookupMatchesThePin builds and parses the table from
// the cached wheel and asserts Lookup against rows the Python
// reference (unidecode's x020.py and translit_test.go's
// TestTransliterateRune rows) reports for the same codepoints, so a
// defect that agrees with wantTableSHA256 and wantEntryCount but
// mistranslates a lookup path (the ASCII fast path, an explicit empty
// mapping, a None entry, a codepoint above the block table, or one
// beyond unidecodeMaxCodepoint) still fails here.
func TestBuildParseLookupMatchesThePin(t *testing.T) {
	wheel := loadCachedWheel(t)

	built, err := Build(wheel)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	table, err := Parse(built)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	tests := []struct {
		name     string
		r        rune
		wantRepl string
		wantOK   bool
	}{
		{"ascii passthrough below 0x80", 'A', "A", true},
		{"latin-1 e-acute (U+00E9) maps to plain e", 0xe9, "e", true},
		{"explicit empty mapping (U+0080) is present, not None", 0x80, "", true},
		{"unmapped codepoint inside a mapped section (U+0241) is None", 0x241, "", false},
		{"hiragana a (U+3042) maps to plain a", 0x3042, "a", true},
		{"codepoint with no shipped block at all (U+10000) is None", 0x10000, "", false},
		{"beyond unidecode's 0xeffff cutoff (U+F0000) is None", 0xf0000, "", false},
		{"line separator (U+2028) maps to a newline", 0x2028, "\n", true},
		{"paragraph separator (U+2029) maps to two newlines", 0x2029, "\n\n", true},
		{"quadruple prime (U+2057) maps to four apostrophes", 0x2057, "''''", true},
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

func TestBuildRefusesWheelChecksumMismatch(t *testing.T) {
	wheel := loadCachedWheel(t)
	corrupted := append([]byte(nil), wheel...)
	corrupted[0] ^= 0xff

	_, err := Build(corrupted)
	if !errors.Is(err, errWheelChecksum) {
		t.Fatalf("Build(flipped byte) error = %v, want errWheelChecksum", err)
	}
}
