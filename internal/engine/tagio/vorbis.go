package tagio

import (
	"encoding/binary"
	"fmt"
	"strings"
)

// Comment is one VORBIS_COMMENT entry, its field name and value cut
// apart at the first '=' the way RFC 9639 section 8.6 defines a
// comment.
type Comment struct {
	Name  string
	Value string
}

// VorbisComment is the decoded VORBIS_COMMENT block: the vendor
// string and every comment as a sequence, not a map, because a key
// may repeat and a map would keep only one occurrence.
type VorbisComment struct {
	Vendor   string
	Comments []Comment
}

// decodeVorbisComment decodes a VORBIS_COMMENT payload. The block is
// little-endian throughout: a vendor length and string, a comment
// count, then each comment's length and its NAME=VALUE bytes.
func decodeVorbisComment(payload []byte) (VorbisComment, error) {
	vendor, rest, err := vorbisField(payload)
	if err != nil {
		return VorbisComment{}, fmt.Errorf("vendor string: %w", err)
	}
	if len(rest) < 4 {
		return VorbisComment{}, fmt.Errorf("VORBIS_COMMENT carries no comment count, %d bytes remain", len(rest))
	}
	count := binary.LittleEndian.Uint32(rest[:4])
	rest = rest[4:]

	var comments []Comment
	for i := range count {
		field, remainder, err := vorbisField(rest)
		if err != nil {
			return VorbisComment{}, fmt.Errorf("comment %d: %w", i, err)
		}
		rest = remainder
		name, value, found := strings.Cut(string(field), "=")
		if !found {
			return VorbisComment{}, fmt.Errorf("comment %d: %q carries no '='", i, field)
		}
		if err := validateFieldName(name); err != nil {
			return VorbisComment{}, fmt.Errorf("comment %d: %w", i, err)
		}
		comments = append(comments, Comment{Name: name, Value: value})
	}
	return VorbisComment{Vendor: string(vendor), Comments: comments}, nil
}

// validateFieldName refuses a comment field name carrying a code point
// outside U+0020 through U+007E, RFC 9639 section 8.6's bound, which
// admits U+007E where mutagen's is_valid_key does not
// (mutagen/_vorbis.py:25-41).
func validateFieldName(name string) error {
	for _, r := range name {
		if r < 0x20 || r > 0x7e {
			return fmt.Errorf("comment field name %q carries a code point outside U+0020-U+007E", name)
		}
	}
	return nil
}

// vorbisField splits one length-prefixed field off the front of b,
// checking the declared length against what remains before any slice
// is taken.
func vorbisField(b []byte) (field, rest []byte, err error) {
	if len(b) < 4 {
		return nil, nil, fmt.Errorf("length field wants 4 bytes, %d remain", len(b))
	}
	size := binary.LittleEndian.Uint32(b[:4])
	if uint64(4)+uint64(size) > uint64(len(b)) {
		return nil, nil, fmt.Errorf("field declares %d bytes, %d remain", size, len(b)-4)
	}
	return b[4 : 4+size], b[4+size:], nil
}

// encodeVorbisComment lays out a VORBIS_COMMENT payload in the shape
// decodeVorbisComment reads.
func encodeVorbisComment(vendor string, comments []Comment) []byte {
	buf := appendLengthPrefixed(binary.LittleEndian, nil, vendor)
	buf = binary.LittleEndian.AppendUint32(buf, uint32(len(comments)))
	for _, c := range comments {
		buf = appendLengthPrefixed(binary.LittleEndian, buf, c.Name+"="+c.Value)
	}
	return buf
}
