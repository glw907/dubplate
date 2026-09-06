package shadow

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// flacSignature opens every FLAC stream, ahead of the metadata blocks.
const flacSignature = "fLaC"

// The two metadata block types this reader looks inside. Every other
// type is skipped by its declared length.
const (
	blockStreamInfo    = 0
	blockVorbisComment = 4
)

// streamInfoSize is STREAMINFO's fixed payload width. A shorter block
// is a malformed container rather than a block to read defensively.
const streamInfoSize = 34

// streamInfo is STREAMINFO's decoded content, of which the shadow reads
// only the two quantities the track length divides. Channels and
// BitsPerSample are decoded because they share the same eight-byte pack
// and pin the reader's offsets to a real file (flac_test.go's committed
// template assertion).
type streamInfo struct {
	SampleRate    uint32
	Channels      uint8
	BitsPerSample uint8
	TotalSamples  uint64
}

// length is the track length in seconds, which is zero for a stream
// declaring no sample rate rather than a division by zero.
func (si streamInfo) length() float64 {
	if si.SampleRate == 0 {
		return 0
	}
	return float64(si.TotalSamples) / float64(si.SampleRate)
}

// flacFile is one container's decoded metadata: the stream's own
// description and its Vorbis comments, keyed upper case with the first
// occurrence of a key kept. Later occurrences are dropped at parse time
// because mediafile reads a comment's first value alone
// (mediafile/storage/base.py:77-84), so nothing downstream can see one.
type flacFile struct {
	info streamInfo
	tags map[string]string
}

// readFLAC reads one FLAC file's matcher input: the nineteen fields
// matcher.ItemFields carries (info.go:124-152), read one way, for
// scoring only.
//
// It is a provisional reader. Rung 8 owns tag I/O and the full field
// surface in internal/engine/tagio (internal/engine/tag/doc.go:8-11);
// rung 4 builds this because waiting on that lane would spend calendar
// on the one rung whose calendar IS the launch date, and rung 8's plan
// carries the task of absorbing it. It writes no tag, reads no format
// but FLAC, and reads exactly two of the container's metadata blocks:
// STREAMINFO (type 0) for the track length, and VORBIS_COMMENT (type 4)
// for everything else. Every other block is skipped by its declared
// length.
//
// An absent or unparseable tag takes the field's Go zero value, which is
// ProjectItemFields' own typed-default rule (info.go:178-188) and beets'
// Item.get behavior. A malformed container is an error and never a
// partly filled bag, because a capture that cannot read every entry it
// listed is discarded whole rather than scored short.
func readFLAC(path string) (matcher.ItemFields, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return matcher.ItemFields{}, fmt.Errorf("read flac: %w", err)
	}
	file, err := parseFLAC(data)
	if err != nil {
		return matcher.ItemFields{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return file.fields(), nil
}

// parseFLAC decodes data's signature and metadata blocks.
func parseFLAC(data []byte) (flacFile, error) {
	if len(data) < len(flacSignature) || string(data[:len(flacSignature)]) != flacSignature {
		return flacFile{}, fmt.Errorf("not a FLAC stream: signature %q", firstBytes(data, len(flacSignature)))
	}

	file := flacFile{tags: make(map[string]string)}
	for off := len(flacSignature); ; {
		if off+4 > len(data) {
			return flacFile{}, fmt.Errorf("metadata block header at offset %d runs past the end", off)
		}
		last := data[off]&0x80 != 0
		blockType := data[off] & 0x7f
		length := int(data[off+1])<<16 | int(data[off+2])<<8 | int(data[off+3])
		off += 4
		if off+length > len(data) {
			return flacFile{}, fmt.Errorf("metadata block of type %d at offset %d declares %d bytes, %d remain", blockType, off, length, len(data)-off)
		}
		payload := data[off : off+length]
		off += length

		switch blockType {
		case blockStreamInfo:
			info, err := parseStreamInfo(payload)
			if err != nil {
				return flacFile{}, err
			}
			file.info = info
		case blockVorbisComment:
			if err := file.readVorbisComment(payload); err != nil {
				return flacFile{}, err
			}
		}
		if last {
			return file, nil
		}
	}
}

// parseStreamInfo decodes STREAMINFO's 34-byte payload. The layout is
// mutagen's own write order (mutagen/flac.py:233-257): two 16-bit block
// sizes, two 24-bit frame sizes, then an eight-byte pack at [10:18]
// holding the sample rate (20 bits), the channel count less one (3),
// the sample width less one (5) and the total sample count (36), then
// the audio MD5 at [18:34]. The last eight bytes of the payload are the
// tail of that digest and not the pack.
func parseStreamInfo(payload []byte) (streamInfo, error) {
	if len(payload) != streamInfoSize {
		return streamInfo{}, fmt.Errorf("STREAMINFO carries %d bytes, want %d", len(payload), streamInfoSize)
	}
	pack := binary.BigEndian.Uint64(payload[10:18])
	return streamInfo{
		SampleRate:    uint32(pack >> 44),
		Channels:      uint8(pack>>41&0x7) + 1,
		BitsPerSample: uint8(pack>>36&0x1f) + 1,
		TotalSamples:  pack & 0xFFFFFFFFF,
	}, nil
}

// readVorbisComment decodes a VORBIS_COMMENT payload into f.tags. The
// block is little-endian throughout: a vendor length and string, a
// comment count, then each comment's length and its KEY=value bytes.
// Keys fold to upper case because mutagen's own lookup is
// case-insensitive.
func (f *flacFile) readVorbisComment(payload []byte) error {
	_, rest, err := vorbisField(payload)
	if err != nil {
		return fmt.Errorf("vendor string: %w", err)
	}
	if len(rest) < 4 {
		return fmt.Errorf("VORBIS_COMMENT carries no comment count, %d bytes remain", len(rest))
	}
	count := binary.LittleEndian.Uint32(rest[:4])
	rest = rest[4:]

	for i := range count {
		comment, remainder, err := vorbisField(rest)
		if err != nil {
			return fmt.Errorf("comment %d: %w", i, err)
		}
		rest = remainder
		key, value, found := strings.Cut(string(comment), "=")
		if !found {
			continue
		}
		key = strings.ToUpper(key)
		if _, seen := f.tags[key]; !seen {
			f.tags[key] = value
		}
	}
	return nil
}

// vorbisField splits one length-prefixed field off the front of b,
// returning the field and what follows it.
func vorbisField(b []byte) (field, rest []byte, err error) {
	if len(b) < 4 {
		return nil, nil, fmt.Errorf("length field wants 4 bytes, %d remain", len(b))
	}
	size := int(binary.LittleEndian.Uint32(b[:4]))
	if size < 0 || 4+size > len(b) {
		return nil, nil, fmt.Errorf("field declares %d bytes, %d remain", size, len(b)-4)
	}
	return b[4 : 4+size], b[4+size:], nil
}

// fields projects the decoded container into the matcher's input, one
// row per mediafile MediaField of pinned mediafile 0.17.0, the version
// corpus/manifest.json pins. Where a field declares several storage
// styles they are listed in mediafile's own declaration order, because
// MediaField.__get__ takes the first style yielding a truthy value
// (mediafile/fields.py:48-54). Read-only styles are read here because
// this reader only reads.
//
// DataSource has no row: a grep of pinned mediafile for data_source
// returns nothing, since beets keeps it as a database flexible
// attribute rather than in the file, so an inbox file always projects
// the empty string ProjectItemFields' typed-default rule already models
// (info.go:115-123).
func (f flacFile) fields() matcher.ItemFields {
	return matcher.ItemFields{
		Title:         f.tag("TITLE"),                                         // mediafile:392
		Artist:        f.tag("ARTIST"),                                        // mediafile:398
		Album:         f.tag("ALBUM"),                                         // mediafile:410
		AlbumArtist:   f.tag("ALBUM ARTIST", "ALBUM_ARTIST", "ALBUMARTIST"),   // mediafile:542-544
		Track:         castInt(f.tag("TRACK", "TRACKNUMBER")),                 // mediafile:463-464
		Disc:          castInt(f.tag("DISC", "DISCNUMBER")),                   // mediafile:480-481
		DiscTotal:     castInt(f.tag("DISCTOTAL", "DISCC", "TOTALDISCS")),     // mediafile:488-490
		Comp:          castBool(f.tag("COMPILATION")),                         // mediafile:535
		Label:         f.tag("LABEL", "PUBLISHER"),                            // mediafile:574-575
		CatalogNum:    f.tag("CATALOGNUMBER", "CATALOGID", "DISCOGS_CATALOG"), // mediafile:604-606
		Barcode:       f.tag("BARCODE", "UPC", "EAN/UPN", "EAN", "UPN"),       // mediafile:616-620
		Country:       f.tag("RELEASECOUNTRY"),                                // mediafile:659
		Media:         f.tag("MEDIA"),                                         // mediafile:673
		AlbumDisambig: f.tag("MUSICBRAINZ_ALBUMCOMMENT"),                      // mediafile:680
		MBTrackID:     f.tag("MUSICBRAINZ_TRACKID"),                           // mediafile:764
		MBAlbumID:     f.tag("MUSICBRAINZ_ALBUMID"),                           // mediafile:782
		Year:          f.year(),                                               // mediafile:685-693
		Length:        f.info.length(),                                        // mediafile:892-895
	}
}

// tag returns the first key holding a non-empty value. An empty value
// falls through to the next style and a "0" does not, because
// mediafile's truthiness test runs on the fetched string before
// safe_cast sees it, and "0" is a truthy Python string.
func (f flacFile) tag(keys ...string) string {
	for _, key := range keys {
		if v := f.tags[key]; v != "" {
			return v
		}
	}
	return ""
}

// year is DATE's year component with YEAR as the fallback style
// (mediafile/__init__.py:685-693). _get_date_tuple strips anything from
// a T or a space onward, splits on - or /, and reaches for the year
// style only when that first component is empty
// (mediafile/fields.py:156-172), so a DATE whose year does not parse
// yields zero rather than falling through to YEAR.
func (f flacFile) year() int {
	year := dateYear(f.tags["DATE"])
	if year == "" {
		year = f.tags["YEAR"]
	}
	n, err := strconv.Atoi(strings.TrimSpace(year))
	if err != nil {
		return 0
	}
	return n
}

// dateYear returns the year component of a Vorbis DATE value.
func dateYear(date string) string {
	if i := strings.IndexAny(date, "Tt "); i >= 0 {
		date = date[:i]
	}
	if i := strings.IndexAny(date, "-/"); i >= 0 {
		date = date[:i]
	}
	return date
}

// castInt reproduces mediafile's safe_cast for an int
// (mediafile/utils/type_conversion.py:15-27): the leading [+-]?[0-9]+
// of the trimmed value, or zero when there is none, so TRACKNUMBER=3/12
// is 3 and TRACKNUMBER=A is 0.
func castInt(s string) int {
	s = strings.TrimSpace(s)
	end := 0
	if end < len(s) && (s[end] == '+' || s[end] == '-') {
		end++
	}
	digits := end
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == digits {
		return 0
	}
	// A value wider than an int is beyond what beets' own SQLite
	// integer column holds, so it takes the zero value here rather
	// than a truncation.
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0
	}
	return n
}

// castBool reproduces mediafile's safe_cast for a bool
// (mediafile/utils/type_conversion.py:29-34), which is bool(int(val))
// with a ValueError falling to false, so COMPILATION=1 is true and
// COMPILATION=yes is false.
func castBool(s string) bool {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	return err == nil && n != 0
}

// firstBytes returns at most n bytes of b, for an error message about a
// file that may hold anything at all.
func firstBytes(b []byte, n int) []byte {
	return b[:min(len(b), n)]
}
