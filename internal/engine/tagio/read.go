package tagio

import (
	"strconv"
	"strings"
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
	return Tags(c), nil
}

// Tags is [project] exported over an already-parsed Container, for a
// caller that needs one file's tags and another Container field, such
// as StreamInfo, off a single [Parse]. The import reader is the
// intended consumer; no caller exists yet. Read stays the
// sniffing, path-taking entry point for a caller that
// wants only the tags; Tags never opens a file, never sniffs a
// format, and is not a second entry point.
func Tags(c Container) TagSet {
	return project(c)
}

// project reads all 69 fields out of c's VORBIS_COMMENT block the way
// mediafile 0.17.0's field descriptors read them (C6).
//
// A field's styles are walked in declaration order and the first
// truthy one wins, truthy being Python's test on the fetched string
// before any cast: "0" is truthy and only "" is not, so TRACK=0
// beside TRACKNUMBER=5 reads 0 rather than 5. When no style is
// truthy, what survives is the LAST style's value rather than the
// first, absent included (mediafile/fields.py:48-54), which is why
// catalognum and barcode read back absent for the empty value a write
// left on disk: their trailing styles are read-only and a file rarely
// carries them (C5).
//
// An empty list and an absent key are indistinguishable here, and the
// ambiguity is on disk rather than in this code: a list field holding
// no values writes no comment at all, so both states leave identical
// bytes and both project to a nil slice (C2, measured). An absent key
// projects to a nil member rather than to a typed zero, because
// safe_cast of an absent value is absent for every type.
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
// 9639 section 8.6 requires of every field-name evaluation (C8), so a
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
