package tagio

import (
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Read sniffs path, then projects its VORBIS_COMMENT block into a
// TagSet. A file Sniff does not recognize hands back Sniff's
// [ErrUnsupportedContainer] unchanged; a FLAC file hands back
// [ErrMalformedContainer] unchanged from [Parse], so a caller
// branching on either branches on the same error whether it parsed
// the container itself or read its tags.
func Read(path string) (TagSet, error) {
	// Sniff first so a file that is not FLAC reports Sniff's error, not
	// the signature refusal Parse would give it.
	if _, err := Sniff(path); err != nil {
		return TagSet{}, err
	}
	c, err := Parse(path)
	if err != nil {
		return TagSet{}, err
	}
	return project(c), nil
}

// project reads all 69 fields out of c's VORBIS_COMMENT block the way
// mediafile 0.17.0's field descriptors read them.
//
// A field's styles are walked in declaration order and the first
// truthy one wins, truthy being Python's test on the fetched string
// before any cast: "0" is truthy and only "" is not, so TRACK=0
// beside TRACKNUMBER=5 reads 0 rather than 5. When no style is
// truthy, what survives is the LAST style's value rather than the
// first, absent included (mediafile/fields.py:48-54), which is why
// catalognum and barcode read back absent for the empty value a write
// left on disk: their trailing styles are read-only and a file rarely
// carries them.
//
// An empty list and an absent key are indistinguishable here, and the
// ambiguity is on disk rather than in this code: a list field holding
// no values writes no comment at all, so both states leave identical
// bytes and both project to a nil slice. An absent key projects to a
// nil member rather than to a typed zero, because safe_cast of an
// absent value is absent for every type.
func project(c Container) TagSet {
	cs := c.VorbisComment.Comments
	var ts TagSet
	for _, f := range fields {
		if f.codec.read != nil {
			f.codec.read(&ts, cs, f.styles)
		}
	}
	return ts
}

// readStyles returns the value MediaField.__get__ hands safe_cast:
// the first truthy style's value, or the last style's value when none
// is truthy, along with whether that value exists on disk at all. The
// absent case is not the empty case: an absent key ends the read as a
// nil member, while an empty value is a value a string field keeps.
func readStyles(cs []Comment, styles []style) (string, bool) {
	var (
		value   string
		present bool
	)
	for _, s := range styles {
		value, present = styleValue(cs, s.key)
		if value != "" {
			break
		}
	}
	return value, present
}

// styleValue returns the first value cs holds under key and whether
// the key is there at all. Names match case insensitively, as RFC
// 9639 section 8.6 requires of every field-name evaluation, so a
// container another tagger wrote in lowercase reads the same as one
// written in the conventional uppercase.
func styleValue(cs []Comment, key string) (string, bool) {
	for _, c := range cs {
		if strings.EqualFold(c.Name, key) {
			return replaceInvalidUTF8(c.Value), true
		}
	}
	return "", false
}

// styleValues returns every value cs holds under key, in on-disk
// order.
func styleValues(cs []Comment, key string) []string {
	var out []string
	for _, c := range cs {
		if strings.EqualFold(c.Name, key) {
			out = append(out, replaceInvalidUTF8(c.Value))
		}
	}
	return out
}

// readList projects a []string member. ListMediaField.__get__ takes
// every value under the first style holding any and never reaches the
// last-style rule: a field no style holds values for is absent, which
// is the same nil slice an empty list leaves.
func readList(cs []Comment, styles []style) []string {
	for _, s := range styles {
		if values := styleValues(cs, s.key); len(values) > 0 {
			return values
		}
	}
	return nil
}

// readGain casts a ReplayGain gain, whose styles declare the " dB"
// suffix that is the only suffix in the FLAC table (mediafile:836,
// :846). StorageStyle.deserialize strips a declared suffix before
// safe_cast runs, and only when the value ends with it.
func readGain(value string) float64 {
	return safeCastFloat(strings.TrimSuffix(value, " dB"))
}

// readQ8 projects the two R128 members, which are Q8 fixed point on
// disk: QNumberField reads the integer and divides it by 2 to the
// fraction_bits, 8 here, so R128_TRACK_GAIN=-384 reads -1.5
// (mediafile:874, :881).
func readQ8(value string) float64 {
	return float64(safeCastInt(value)) / 256
}

// readDate recomposes a date triple from the first of styles, the
// date style, with the bare-year fallback style behind it when there
// is one (DATE has YEAR; ORIGINALDATE has nothing). The fallback is
// read only when the date's own year component is falsy, so a DATE
// whose year does not parse yields an absent year rather than falling
// through (mediafile/fields.py:156-172).
//
// Each component parses strictly rather than through safe_cast, which
// is why a component that is missing, empty or unparsable is absent
// rather than zero: _get_date_tuple calls int() and answers a
// TypeError or a ValueError with None.
func readDate(cs []Comment, styles []style) (year, month, day *int) {
	value, present := styleValue(cs, styles[0].key)
	var parts [3]string
	if present {
		for i, part := range splitDate(value) {
			if i == len(parts) {
				break
			}
			parts[i] = part
		}
	}
	if parts[0] == "" && len(styles) > 1 {
		parts[0], _ = styleValue(cs, styles[1].key)
	}
	return dateComponent(parts[0]), dateComponent(parts[1]), dateComponent(parts[2])
}

// splitDate drops everything from a T or a space onward, then splits
// on hyphens and slashes alike, keeping empty components so a
// "2001--04" leaves its month empty rather than shifting the day into
// it.
func splitDate(value string) []string {
	if i := strings.IndexAny(value, "Tt "); i >= 0 {
		value = value[:i]
	}
	return strings.Split(strings.ReplaceAll(value, "/", "-"), "-")
}

// dateComponent parses one date component, answering nil for anything
// int() would refuse. Python's int() ignores surrounding whitespace,
// which strconv.Atoi does not, so the value is trimmed first.
func dateComponent(part string) *int {
	n, err := strconv.Atoi(strings.TrimSpace(part))
	if err != nil {
		return nil
	}
	return &n
}

// intPrefixPattern and floatPrefixPattern reproduce safe_cast's
// leading-numeric-prefix rule exactly (mediafile/utils/type_conversion.py:4-60).
var (
	intPrefixPattern   = regexp.MustCompile(`^[+-]?[0-9]+`)
	floatPrefixPattern = regexp.MustCompile(`^[+-]?([0-9]+\.?[0-9]*|[0-9]*\.[0-9]+)`)
)

// safeCastInt takes s's leading [+-]?[0-9]+ prefix and parses it,
// yielding 0 when s carries no such prefix (safe_cast to int). It
// never sees an absent value: a caller that has not found a truthy
// style has nothing to cast.
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
// replaced by U+FFFD, matching mutagen's errors='replace' decode.
// Python's utf-8 codec follows the Unicode "maximal subpart" rule: an
// ill-formed sequence collapses its longest well-formed prefix (the
// lead byte plus whatever continuation bytes already matched) into a
// single U+FFFD, then resumes scanning at the byte that broke the
// sequence. unicode/utf8's own DecodeRuneInString does not have these
// semantics: it always reports size 1 on an invalid encoding, where
// mutagen/_vorbis.py:109 collapses the whole maximal subpart instead,
// so this scans the byte structure explicitly via utf8SequenceSpan
// rather than leaning on the stdlib decoder. Values are valid UTF-8 by
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
