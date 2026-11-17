package tagio

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// KeyWrite is one field's contribution to a write: the field's ledger
// name, which is also its TagSet member's JSON tag; every Vorbis key
// the field can ever touch; and, when the field carries a value, the
// ordered comments a write appends. Deletes lists a read-only style
// only when the field itself is absent, because read_only blocks a
// set but not a delete (INV:238-240); Writes covers writable styles
// only. Task 7 applies a KeyWrite's Deletes before its Writes, in the
// order Serialize returns them, which is what resolves the three
// colliding pairs the plural's own way (C3).
type KeyWrite struct {
	Field   string
	Deletes []string
	Writes  []Comment
}

// Serialize turns ts into one KeyWrite per field, in the field table's
// write order. An absent member deletes every key its field owns; a
// present one, empty included, deletes the writable keys and writes
// its comments in their place. The year, month and day members compose
// into one KeyWrite named for the year, as do the three original_
// members, because one DATE or ORIGINALDATE comment carries all three
// components.
func Serialize(ts TagSet) []KeyWrite {
	out := make([]KeyWrite, 0, len(fields))
	for _, f := range fields {
		if f.codec.write == nil {
			continue
		}
		writes, present := f.codec.write(&ts, f.styles)
		if !present {
			out = append(out, KeyWrite{Field: f.name, Deletes: allKeys(f.styles)})
			continue
		}
		out = append(out, KeyWrite{Field: f.name, Deletes: writableKeys(f.styles), Writes: writes})
	}
	return out
}

// allKeys lists every key styles records, writable and read-only
// alike. A delete of an absent field removes all of them: deleting
// barcode removes UPC, EAN/UPN, EAN and UPN along with BARCODE.
func allKeys(styles []style) []string {
	keys := make([]string, len(styles))
	for i, s := range styles {
		keys[i] = s.key
	}
	return keys
}

// writableKeys lists the writable keys in styles, the set a present
// value replaces.
func writableKeys(styles []style) []string {
	var keys []string
	for _, s := range styles {
		if s.writable {
			keys = append(keys, s.key)
		}
	}
	return keys
}

// repeatComment writes value once under every writable style.
func repeatComment(styles []style, value string) []Comment {
	var out []Comment
	for _, s := range styles {
		if s.writable {
			out = append(out, Comment{Name: s.key, Value: value})
		}
	}
	return out
}

// listComments writes every element of values under each writable
// style in turn, in list order.
func listComments(styles []style, values []string) []Comment {
	var out []Comment
	for _, s := range styles {
		if !s.writable {
			continue
		}
		for _, v := range values {
			out = append(out, Comment{Name: s.key, Value: v})
		}
	}
	return out
}

// formatBool writes comp as "1" or "0", never "true" or "false"
// (mediafile:535, StorageStyle.serialize).
func formatBool(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// formatGain writes a ReplayGain gain to two decimal places, the
// default float_places, with the " dB" suffix that is the only suffix
// in the FLAC table (mediafile:836, :846).
func formatGain(v float64) string {
	return fmt.Sprintf("%.2f dB", v)
}

// formatPeak writes a ReplayGain peak to six decimal places, with no
// suffix (mediafile:856, :864).
func formatPeak(v float64) string {
	return fmt.Sprintf("%.6f", v)
}

// formatQ8 writes an R128 gain as an integer string, the value shifted
// by fraction_bits=8: round(value * 2**8) (mediafile:874, :881;
// __init__.py:870-884). r128_track_gain = -1.5 writes
// R128_TRACK_GAIN=-384.
func formatQ8(v float64) string {
	return strconv.FormatInt(int64(math.Round(v*256)), 10)
}

// dateComments composes the DATE (or ORIGINALDATE) comment and, when
// styles carries a YEAR style, the bare-year fallback, from a year,
// month and day (mediafile/fields.py:182-201). The year is zero-padded
// to four digits, the month is appended zero-padded to two digits only
// when it is nonzero, and the day only when both the month and the day
// are. The caller passes an absent month or day as zero: Write takes a
// whole tag set rather than mediafile's read-modify-write update, so
// there is no stored month or day for an absent one to keep.
func dateComments(styles []style, y, m, d int) []Comment {
	date := fmt.Sprintf("%04d", y)
	if m != 0 {
		date += fmt.Sprintf("-%02d", m)
		if d != 0 {
			date += fmt.Sprintf("-%02d", d)
		}
	}

	var out []Comment
	for _, s := range styles {
		if !s.writable {
			continue
		}
		switch s.key {
		case "YEAR":
			out = append(out, Comment{Name: s.key, Value: strconv.Itoa(y)})
		default: // DATE, ORIGINALDATE
			out = append(out, Comment{Name: s.key, Value: date})
		}
	}
	return out
}

// valueOr returns 0 for a nil *int and the pointed-to value otherwise.
func valueOr(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

// intPrefixPattern and floatPrefixPattern reproduce safe_cast's
// leading-numeric-prefix rule exactly (mediafile/utils/type_conversion.py:4-60).
var (
	intPrefixPattern   = regexp.MustCompile(`^[+-]?[0-9]+`)
	floatPrefixPattern = regexp.MustCompile(`^[+-]?([0-9]+\.?[0-9]*|[0-9]*\.[0-9]+)`)
)

// safeCastInt takes s's leading [+-]?[0-9]+ prefix and parses it,
// yielding 0 when s carries no such prefix (safe_cast to int,
// INV:253-260). It never sees an absent value: a caller that has not
// found a truthy style has nothing to cast.
func safeCastInt(s string) int {
	prefix := intPrefixPattern.FindString(strings.TrimSpace(s))
	if prefix == "" {
		return 0
	}
	n, err := strconv.Atoi(prefix)
	if err != nil {
		return 0
	}
	return n
}

// safeCastFloat takes s's leading numeric prefix and parses it,
// yielding 0.0 when s carries no such prefix (safe_cast to float).
// The suffix on rg_track_gain and rg_album_gain is stripped by the
// caller before this runs, though safe_cast's own prefix match would
// stop at " dB" either way.
func safeCastFloat(s string) float64 {
	prefix := floatPrefixPattern.FindString(strings.TrimSpace(s))
	if prefix == "" {
		return 0
	}
	f, err := strconv.ParseFloat(prefix, 64)
	if err != nil {
		return 0
	}
	return f
}

// safeCastBool is bool(int(val)) with a parse failure falling to
// false (safe_cast to bool). Unlike safeCastInt, mediafile's int()
// constructor here takes the whole trimmed string rather than a
// leading prefix, so "1x" is a parse failure and yields false.
func safeCastBool(s string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return false
	}
	return n != 0
}

// replaceInvalidUTF8 returns s with each invalid UTF-8 code point
// replaced by U+FFFD, matching mutagen's errors='replace' decode
// (ruling 10). Python's utf-8 codec follows the Unicode "maximal
// subpart" rule: an ill-formed sequence collapses its longest
// well-formed prefix (the lead byte plus whatever continuation bytes
// already matched) into a single U+FFFD, then resumes scanning at the
// byte that broke the sequence. unicode/utf8's own DecodeRuneInString
// does not have these semantics: it always reports size 1 on an
// invalid encoding, where mutagen/_vorbis.py:109 collapses the whole
// maximal subpart instead, so this scans the byte structure explicitly
// via utf8SequenceSpan rather than leaning on the stdlib decoder.
// Values are valid UTF-8 by
// construction on write, so this only ever runs on read.
func replaceInvalidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		validLen, invalidLen := utf8SequenceSpan(s[i:])
		if validLen > 0 {
			b.WriteString(s[i : i+validLen])
			i += validLen
			continue
		}
		b.WriteRune(utf8.RuneError)
		i += invalidLen
	}
	return b.String()
}

// utf8SequenceSpan looks at s's leading bytes and reports either the
// byte length of a well-formed UTF-8 sequence starting there
// (validLen, with invalidLen zero) or the byte length of the maximal
// ill-formed subpart starting there (invalidLen, with validLen zero):
// the lead byte plus however many of its required continuation bytes
// are present and in range before one is missing or out of range.
// utf8SequenceSpan never sees an empty s; replaceInvalidUTF8's loop
// only calls it while i < len(s).
func utf8SequenceSpan(s string) (validLen, invalidLen int) {
	b0 := s[0]
	if b0 < 0x80 {
		return 1, 0
	}

	var contRanges [][2]byte
	switch {
	case b0 >= 0xC2 && b0 <= 0xDF:
		contRanges = [][2]byte{{0x80, 0xBF}}
	case b0 == 0xE0:
		contRanges = [][2]byte{{0xA0, 0xBF}, {0x80, 0xBF}}
	case b0 >= 0xE1 && b0 <= 0xEC:
		contRanges = [][2]byte{{0x80, 0xBF}, {0x80, 0xBF}}
	case b0 == 0xED:
		contRanges = [][2]byte{{0x80, 0x9F}, {0x80, 0xBF}}
	case b0 >= 0xEE && b0 <= 0xEF:
		contRanges = [][2]byte{{0x80, 0xBF}, {0x80, 0xBF}}
	case b0 == 0xF0:
		contRanges = [][2]byte{{0x90, 0xBF}, {0x80, 0xBF}, {0x80, 0xBF}}
	case b0 >= 0xF1 && b0 <= 0xF3:
		contRanges = [][2]byte{{0x80, 0xBF}, {0x80, 0xBF}, {0x80, 0xBF}}
	case b0 == 0xF4:
		contRanges = [][2]byte{{0x80, 0x8F}, {0x80, 0xBF}, {0x80, 0xBF}}
	default:
		// A stray continuation byte (0x80-0xBF) or a byte that can
		// never lead a sequence (0xC0, 0xC1, 0xF5-0xFF): a maximal
		// subpart of length 1, on its own.
		return 0, 1
	}

	for i, r := range contRanges {
		if i+1 >= len(s) || s[i+1] < r[0] || s[i+1] > r[1] {
			return 0, i + 1
		}
	}
	return len(contRanges) + 1, 0
}
