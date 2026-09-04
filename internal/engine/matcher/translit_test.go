package matcher

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"testing"
)

// TestUnidecodeVersion proves the exported stamp and the embedded
// table's own version field cannot drift apart: a table regenerated
// against a different unidecode release without updating the
// constant, or vice versa, fails here rather than silently.
func TestUnidecodeVersion(t *testing.T) {
	table := loadUnidecodeTable()
	if table.version != UnidecodeVersion {
		t.Fatalf("UnidecodeVersion = %q, embedded table carries %q", UnidecodeVersion, table.version)
	}
}

func TestTransliterateRune(t *testing.T) {
	tests := []struct {
		name    string
		r       rune
		wantOK  bool
		wantVal string
	}{
		{name: "ascii passthrough below 0x80", r: 'A', wantOK: true, wantVal: "A"},
		{name: "latin-1 e-acute (U+00E9) maps to plain e", r: 0xe9, wantOK: true, wantVal: "e"},
		{name: "explicit empty mapping (U+0080) is present, not None", r: 0x80, wantOK: true, wantVal: ""},
		{name: "unmapped codepoint inside a mapped section (U+0241) is None", r: 0x241, wantOK: false, wantVal: ""},
		{name: "hiragana a (U+3042) maps to plain a", r: 0x3042, wantOK: true, wantVal: "a"},
		{name: "codepoint with no shipped block at all is None", r: 0x10000, wantOK: false, wantVal: ""},
		{name: "beyond unidecode's 0xeffff cutoff is None", r: 0xf0000, wantOK: false, wantVal: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, gotOK := transliterateRune(tt.r)
			if gotOK != tt.wantOK || gotVal != tt.wantVal {
				t.Errorf("transliterateRune(%U) = %q, %v; want %q, %v", tt.r, gotVal, gotOK, tt.wantVal, tt.wantOK)
			}
		})
	}
}

func TestTransliterate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "already ascii", input: "hello", want: "hello"},
		{name: "cafe with e-acute (U+00E9)", input: "caf" + string(rune(0xe9)), want: "cafe"},
		{
			name:  "an unmapped codepoint contributes nothing",
			input: "a" + string(rune(0x241)) + "b",
			want:  "ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Transliterate(tt.input); got != tt.want {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// sweepDigest hashes transliterateRune's output over [lo, hi] using
// the exact encoding tools/unidecode-table/generate.py's sweep_digest
// uses, so the two sides can only agree if the Go decoder reproduces
// unidecode's behavior for every codepoint in range.
func sweepDigest(lo, hi uint32) [32]byte {
	h := sha256.New()
	var codepointBuf, lengthBuf [4]byte
	for cp := lo; cp <= hi; cp++ {
		binary.BigEndian.PutUint32(codepointBuf[:], cp)
		h.Write(codepointBuf[:])
		repl, ok := transliterateRune(rune(cp))
		if !ok {
			h.Write([]byte{0})
			continue
		}
		h.Write([]byte{1})
		encoded := []byte(repl)
		binary.BigEndian.PutUint32(lengthBuf[:], uint32(len(encoded)))
		h.Write(lengthBuf[:])
		h.Write(encoded)
	}
	var out [32]byte
	copy(out[:], h.Sum(nil))
	return out
}

// TestUnidecodeTableFullSweep is the fidelity proof the plan requires:
// a full codepoint sweep, not a spot check. It recomputes the digest
// from the embedded table and the Go decoder above and compares it to
// the digest tools/unidecode-table/generate.py computed directly
// against pinned unidecode 1.4.0 and baked into the table file, over
// every codepoint from 0 to 0x10FFFF (unidecode's own upper bound of
// what chr() accepts), so a table edit or a regeneration against a
// different unidecode release fails here rather than drifting
// silently. Run with -count=1; it iterates the full sweep every time.
func TestUnidecodeTableFullSweep(t *testing.T) {
	table := loadUnidecodeTable()
	got := sweepDigest(table.sweepLo, table.sweepHi)
	if !bytes.Equal(got[:], table.digest[:]) {
		t.Fatalf("sweep [%#x, %#x] digest = %x, table carries %x", table.sweepLo, table.sweepHi, got, table.digest)
	}
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "both empty", a: "", b: "", want: 0},
		{name: "empty vs non-empty", a: "", b: "abc", want: 3},
		{name: "non-empty vs empty", a: "abc", b: "", want: 3},
		{name: "equal-length identical", a: "kitten", b: "kitten", want: 0},
		{name: "equal-length one substitution", a: "abc", b: "abd", want: 1},
		{name: "equal-length every position differs", a: "abc", b: "xyz", want: 3},
		{name: "classic kitten/sitting", a: "kitten", b: "sitting", want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Levenshtein(tt.a, tt.b); got != tt.want {
				t.Errorf("Levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
