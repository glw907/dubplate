package tagio

import (
	"fmt"
	"strconv"
	"strings"
)

// Read sniffs path, then projects its VORBIS_COMMENT block into a
// TagSet through the seam registry's row for the format Sniff
// reports. A format Sniff does not recognize, including a FLAC file
// once the registry has been emptied of its FLAC row, hands back
// [ErrUnsupportedContainer] unchanged; a recognized format still
// hands back [ErrMalformedContainer] unchanged from [Parse], so a
// caller branching on either (C14) branches on the same error
// whether it parsed the container itself or read its tags.
func Read(path string) (TagSet, error) {
	format, err := Sniff(path)
	if err != nil {
		return TagSet{}, err
	}
	return registry[format].read(path)
}

// readFLAC is FLAC's seam registry row for Read: it parses path
// through [Parse] and projects the result the way project always
// has.
func readFLAC(path string) (TagSet, error) {
	c, err := Parse(path)
	if err != nil {
		return TagSet{}, err
	}
	return project(c), nil
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
	for _, row := range mapping {
		projectField(&ts, row, cs)
	}
	return ts
}

// projectField reads one Mapping row into its TagSet member, in the
// encoding that member's Go type takes. It is the read counterpart of
// serializeField and keeps its shape: one case per field rather than
// a reflection-based lookup.
//
// The year, month and day rows are read together from the row that
// owns the DATE style, because one comment carries all three
// components; the month and day rows have nothing left to do when
// their own turn comes.
func projectField(ts *TagSet, row FieldMapping, cs []Comment) {
	switch row.Field {
	case FieldMonth, FieldDay, FieldOriginalMonth, FieldOriginalDay:
		return
	case FieldAcoustIDFingerprint:
		ts.AcoustIDFingerprint = readString(cs, row)
	case FieldAcoustIDID:
		ts.AcoustIDID = readString(cs, row)
	case FieldAlbum:
		ts.Album = readString(cs, row)
	case FieldAlbumartist:
		ts.Albumartist = readString(cs, row)
	case FieldAlbumartistCredit:
		ts.AlbumartistCredit = readString(cs, row)
	case FieldAlbumartistSort:
		ts.AlbumartistSort = readString(cs, row)
	case FieldAlbumartists:
		ts.Albumartists = readList(cs, row)
	case FieldAlbumartistsCredit:
		ts.AlbumartistsCredit = readList(cs, row)
	case FieldAlbumartistsSort:
		ts.AlbumartistsSort = readList(cs, row)
	case FieldAlbumdisambig:
		ts.Albumdisambig = readString(cs, row)
	case FieldAlbumstatus:
		ts.Albumstatus = readString(cs, row)
	case FieldAlbumtype:
		ts.Albumtype = readString(cs, row)
	case FieldAlbumtypes:
		ts.Albumtypes = readList(cs, row)
	case FieldArrangers:
		ts.Arrangers = readList(cs, row)
	case FieldArtist:
		ts.Artist = readString(cs, row)
	case FieldArtistCredit:
		ts.ArtistCredit = readString(cs, row)
	case FieldArtistSort:
		ts.ArtistSort = readString(cs, row)
	case FieldArtists:
		ts.Artists = readList(cs, row)
	case FieldArtistsCredit:
		ts.ArtistsCredit = readList(cs, row)
	case FieldArtistsSort:
		ts.ArtistsSort = readList(cs, row)
	case FieldAsin:
		ts.Asin = readString(cs, row)
	case FieldBarcode:
		ts.Barcode = readString(cs, row)
	case FieldBpm:
		ts.Bpm = readInt(cs, row)
	case FieldCatalognum:
		ts.Catalognum = readString(cs, row)
	case FieldComments:
		ts.Comments = readString(cs, row)
	case FieldComp:
		ts.Comp = readBool(cs, row)
	case FieldComposerSort:
		ts.ComposerSort = readString(cs, row)
	case FieldComposers:
		ts.Composers = readList(cs, row)
	case FieldCountry:
		ts.Country = readString(cs, row)
	case FieldYear:
		ts.Year, ts.Month, ts.Day = readDate(cs, row)
	case FieldDisc:
		ts.Disc = readInt(cs, row)
	case FieldDisctitle:
		ts.Disctitle = readString(cs, row)
	case FieldDisctotal:
		ts.Disctotal = readInt(cs, row)
	case FieldEncoder:
		ts.Encoder = readString(cs, row)
	case FieldGenres:
		ts.Genres = readList(cs, row)
	case FieldGrouping:
		ts.Grouping = readString(cs, row)
	case FieldInitialKey:
		ts.InitialKey = readString(cs, row)
	case FieldIsrc:
		ts.Isrc = readString(cs, row)
	case FieldLabel:
		ts.Label = readString(cs, row)
	case FieldLanguage:
		ts.Language = readString(cs, row)
	case FieldLyricists:
		ts.Lyricists = readList(cs, row)
	case FieldLyrics:
		ts.Lyrics = readString(cs, row)
	case FieldMBAlbumartistid:
		ts.MBAlbumartistid = readString(cs, row)
	case FieldMBAlbumartistids:
		ts.MBAlbumartistids = readList(cs, row)
	case FieldMBAlbumid:
		ts.MBAlbumid = readString(cs, row)
	case FieldMBArtistid:
		ts.MBArtistid = readString(cs, row)
	case FieldMBArtistids:
		ts.MBArtistids = readList(cs, row)
	case FieldMBReleasegroupid:
		ts.MBReleasegroupid = readString(cs, row)
	case FieldMBReleasetrackid:
		ts.MBReleasetrackid = readString(cs, row)
	case FieldMBTrackid:
		ts.MBTrackid = readString(cs, row)
	case FieldMBWorkid:
		ts.MBWorkid = readString(cs, row)
	case FieldMedia:
		ts.Media = readString(cs, row)
	case FieldOriginalYear:
		ts.OriginalYear, ts.OriginalMonth, ts.OriginalDay = readDate(cs, row)
	case FieldR128AlbumGain:
		ts.R128AlbumGain = readQ8(cs, row)
	case FieldR128TrackGain:
		ts.R128TrackGain = readQ8(cs, row)
	case FieldRemixers:
		ts.Remixers = readList(cs, row)
	case FieldRgAlbumGain:
		ts.RgAlbumGain = readGain(cs, row)
	case FieldRgAlbumPeak:
		ts.RgAlbumPeak = readPeak(cs, row)
	case FieldRgTrackGain:
		ts.RgTrackGain = readGain(cs, row)
	case FieldRgTrackPeak:
		ts.RgTrackPeak = readPeak(cs, row)
	case FieldScript:
		ts.Script = readString(cs, row)
	case FieldSubtitle:
		ts.Subtitle = readString(cs, row)
	case FieldTitle:
		ts.Title = readString(cs, row)
	case FieldTrack:
		ts.Track = readInt(cs, row)
	case FieldTracktotal:
		ts.Tracktotal = readInt(cs, row)
	default:
		panic(fmt.Sprintf("tagio: no projection for field %q", row.Field))
	}
}

// readStyles returns the value MediaField.__get__ hands safe_cast:
// the first truthy style's value, or the last style's value when none
// is truthy, along with whether that value exists on disk at all. The
// absent case is not the empty case: an absent key ends the read as a
// nil member, while an empty value is a value a string field keeps.
func readStyles(cs []Comment, row FieldMapping) (string, bool) {
	var (
		value   string
		present bool
	)
	for _, s := range row.Styles {
		value, present = styleValue(cs, s.Key)
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

// readString projects a *string member: safe_cast to str is the
// identity on a value and absent on an absent one.
func readString(cs []Comment, row FieldMapping) *string {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	return &value
}

// readList projects a []string member. ListMediaField.__get__ takes
// every value under the first style holding any and never reaches the
// last-style rule: a field no style holds values for is absent, which
// is the same nil slice an empty list leaves.
func readList(cs []Comment, row FieldMapping) []string {
	for _, s := range row.Styles {
		if values := styleValues(cs, s.Key); len(values) > 0 {
			return values
		}
	}
	return nil
}

// readInt projects a *int member through safe_cast's leading
// [+-]?[0-9]+ prefix, so TRACK=3/12 reads 3 and BPM=fast reads 0.
func readInt(cs []Comment, row FieldMapping) *int {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	n := safeCastInt(value)
	return &n
}

// readBool projects comp's *bool member through bool(int(val)).
func readBool(cs []Comment, row FieldMapping) *bool {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	b := safeCastBool(value)
	return &b
}

// readPeak projects the two ReplayGain peak members, which carry no
// suffix (mediafile:856, :864).
func readPeak(cs []Comment, row FieldMapping) *float64 {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	f := safeCastFloat(value)
	return &f
}

// readGain projects the two ReplayGain gain members, whose styles
// declare the " dB" suffix that is the only suffix in the FLAC table
// (mediafile:836, :846). StorageStyle.deserialize strips a declared
// suffix before safe_cast runs, and only when the value ends with it.
func readGain(cs []Comment, row FieldMapping) *float64 {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	f := safeCastFloat(strings.TrimSuffix(value, " dB"))
	return &f
}

// readQ8 projects the two R128 members, which are Q8 fixed point on
// disk: QNumberField reads the integer and divides it by 2 to the
// fraction_bits, 8 here, so R128_TRACK_GAIN=-384 reads -1.5
// (mediafile:874, :881).
func readQ8(cs []Comment, row FieldMapping) *float64 {
	value, present := readStyles(cs, row)
	if !present {
		return nil
	}
	f := float64(safeCastInt(value)) / 256
	return &f
}

// readDate recomposes a date triple from row's date style, with the
// bare-year fallback style behind it when the row carries one
// (DATE has YEAR; ORIGINALDATE has nothing). The fallback is read
// only when the date's own year component is falsy, so a DATE whose
// year does not parse yields an absent year rather than falling
// through (mediafile/fields.py:156-172).
//
// Each component parses strictly rather than through safe_cast, which
// is why a component that is missing, empty or unparsable is absent
// rather than zero: _get_date_tuple calls int() and answers a
// TypeError or a ValueError with None.
func readDate(cs []Comment, row FieldMapping) (year, month, day *int) {
	value, present := styleValue(cs, row.Styles[0].Key)
	var parts [3]string
	if present {
		for i, part := range splitDate(value) {
			if i == len(parts) {
				break
			}
			parts[i] = part
		}
	}
	if parts[0] == "" && len(row.Styles) > 1 {
		parts[0], _ = styleValue(cs, row.Styles[1].Key)
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
