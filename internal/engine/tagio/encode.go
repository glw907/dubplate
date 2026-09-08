package tagio

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// KeyWrite is one field's contribution to a write: every Vorbis key
// the field can ever touch and, when the field carries a value, the
// ordered comments a write appends. Deletes lists a read-only style
// only when the field itself is absent, because read_only blocks a
// set but not a delete (INV:238-240); Writes covers writable styles
// only. Task 7 applies a KeyWrite's Deletes before its Writes, in the
// order Serialize returns them, which is what resolves the three
// colliding pairs the plural's own way (C3).
type KeyWrite struct {
	Field   Field
	Deletes []string
	Writes  []Comment
}

// Serialize turns ts into the ordered per-field write instructions
// C2's four states and the package contract's serialization list
// produce. year, month and day compose into one KeyWrite keyed on
// FieldYear, and original_year, original_month and original_day
// compose into one keyed on FieldOriginalYear, because DATE and
// ORIGINALDATE are written once from all three components together
// rather than once per component (the compose-as-a-unit rule).
func Serialize(ts TagSet) []KeyWrite {
	out := make([]KeyWrite, 0, len(Mapping))
	for _, row := range Mapping {
		switch row.Field {
		case FieldMonth, FieldDay, FieldOriginalMonth, FieldOriginalDay:
			continue
		case FieldYear:
			out = append(out, serializeDate(row, ts.Year, ts.Month, ts.Day))
		case FieldOriginalYear:
			out = append(out, serializeDate(row, ts.OriginalYear, ts.OriginalMonth, ts.OriginalDay))
		default:
			out = append(out, serializeField(row, ts))
		}
	}
	return out
}

// serializeField dispatches row to the encoding its TagSet member's
// Go type takes. There is one case per field rather than a
// reflection-based lookup, because tagio takes no dependency and a
// concrete switch is what go-conventions asks for over a runtime type
// test.
func serializeField(row FieldMapping, ts TagSet) KeyWrite {
	switch row.Field {
	case FieldAcoustIDFingerprint:
		return stringField(row, ts.AcoustIDFingerprint)
	case FieldAcoustIDID:
		return stringField(row, ts.AcoustIDID)
	case FieldAlbum:
		return stringField(row, ts.Album)
	case FieldAlbumartist:
		return stringField(row, ts.Albumartist)
	case FieldAlbumartistCredit:
		return stringField(row, ts.AlbumartistCredit)
	case FieldAlbumartistSort:
		return stringField(row, ts.AlbumartistSort)
	case FieldAlbumartists:
		return listField(row, ts.Albumartists)
	case FieldAlbumartistsCredit:
		return listField(row, ts.AlbumartistsCredit)
	case FieldAlbumartistsSort:
		return listField(row, ts.AlbumartistsSort)
	case FieldAlbumdisambig:
		return stringField(row, ts.Albumdisambig)
	case FieldAlbumstatus:
		return stringField(row, ts.Albumstatus)
	case FieldAlbumtype:
		return stringField(row, ts.Albumtype)
	case FieldAlbumtypes:
		return listField(row, ts.Albumtypes)
	case FieldArrangers:
		return listField(row, ts.Arrangers)
	case FieldArtist:
		return stringField(row, ts.Artist)
	case FieldArtistCredit:
		return stringField(row, ts.ArtistCredit)
	case FieldArtistSort:
		return stringField(row, ts.ArtistSort)
	case FieldArtists:
		return listField(row, ts.Artists)
	case FieldArtistsCredit:
		return listField(row, ts.ArtistsCredit)
	case FieldArtistsSort:
		return listField(row, ts.ArtistsSort)
	case FieldAsin:
		return stringField(row, ts.Asin)
	case FieldBarcode:
		return stringField(row, ts.Barcode)
	case FieldBpm:
		return intField(row, ts.Bpm)
	case FieldCatalognum:
		return stringField(row, ts.Catalognum)
	case FieldComments:
		return stringField(row, ts.Comments)
	case FieldComp:
		return boolField(row, ts.Comp)
	case FieldComposerSort:
		return stringField(row, ts.ComposerSort)
	case FieldComposers:
		return listField(row, ts.Composers)
	case FieldCountry:
		return stringField(row, ts.Country)
	case FieldDisc:
		return intField(row, ts.Disc)
	case FieldDisctitle:
		return stringField(row, ts.Disctitle)
	case FieldDisctotal:
		return intField(row, ts.Disctotal)
	case FieldEncoder:
		return stringField(row, ts.Encoder)
	case FieldGenres:
		return listField(row, ts.Genres)
	case FieldGrouping:
		return stringField(row, ts.Grouping)
	case FieldInitialKey:
		return stringField(row, ts.InitialKey)
	case FieldIsrc:
		return stringField(row, ts.Isrc)
	case FieldLabel:
		return stringField(row, ts.Label)
	case FieldLanguage:
		return stringField(row, ts.Language)
	case FieldLyricists:
		return listField(row, ts.Lyricists)
	case FieldLyrics:
		return stringField(row, ts.Lyrics)
	case FieldMBAlbumartistid:
		return stringField(row, ts.MBAlbumartistid)
	case FieldMBAlbumartistids:
		return listField(row, ts.MBAlbumartistids)
	case FieldMBAlbumid:
		return stringField(row, ts.MBAlbumid)
	case FieldMBArtistid:
		return stringField(row, ts.MBArtistid)
	case FieldMBArtistids:
		return listField(row, ts.MBArtistids)
	case FieldMBReleasegroupid:
		return stringField(row, ts.MBReleasegroupid)
	case FieldMBReleasetrackid:
		return stringField(row, ts.MBReleasetrackid)
	case FieldMBTrackid:
		return stringField(row, ts.MBTrackid)
	case FieldMBWorkid:
		return stringField(row, ts.MBWorkid)
	case FieldMedia:
		return stringField(row, ts.Media)
	case FieldR128AlbumGain:
		return r128Field(row, ts.R128AlbumGain)
	case FieldR128TrackGain:
		return r128Field(row, ts.R128TrackGain)
	case FieldRemixers:
		return listField(row, ts.Remixers)
	case FieldRgAlbumGain:
		return rgGainField(row, ts.RgAlbumGain)
	case FieldRgAlbumPeak:
		return rgPeakField(row, ts.RgAlbumPeak)
	case FieldRgTrackGain:
		return rgGainField(row, ts.RgTrackGain)
	case FieldRgTrackPeak:
		return rgPeakField(row, ts.RgTrackPeak)
	case FieldScript:
		return stringField(row, ts.Script)
	case FieldSubtitle:
		return stringField(row, ts.Subtitle)
	case FieldTitle:
		return stringField(row, ts.Title)
	case FieldTrack:
		return intField(row, ts.Track)
	case FieldTracktotal:
		return intField(row, ts.Tracktotal)
	default:
		panic(fmt.Sprintf("tagio: no encoding for field %q", row.Field))
	}
}

// allKeys lists every Vorbis key row.Styles records, writable and
// read-only alike. A delete of an absent field removes all of them
// (INV:238-240): deleting barcode removes UPC, EAN/UPN, EAN and UPN
// along with BARCODE.
func allKeys(row FieldMapping) []string {
	keys := make([]string, len(row.Styles))
	for i, s := range row.Styles {
		keys[i] = s.Key
	}
	return keys
}

// writableKeys lists row.Styles's writable keys only, the set a set
// (as opposed to a delete) touches (C4).
func writableKeys(row FieldMapping) []string {
	var keys []string
	for _, s := range row.Styles {
		if s.Writable {
			keys = append(keys, s.Key)
		}
	}
	return keys
}

// repeatComment writes value under every writable style in row,
// unchanged, one comment per style (C4). A one-element caller (a
// scalar field) gets one comment per style; a caller inside a list
// loop gets one call per element, so the same helper covers both a
// plain string field and a repeated list value.
func repeatComment(row FieldMapping, value string) []Comment {
	var out []Comment
	for _, s := range row.Styles {
		if s.Writable {
			out = append(out, Comment{Name: s.Key, Value: value})
		}
	}
	return out
}

// stringField encodes a *string TagSet member: nil is absent and
// deletes every style; a non-nil value, empty or not, deletes only
// the writable styles and writes one comment per writable style
// holding the value unchanged (C2, C4). This also covers the five
// fields that reach a list-backed key through a single value
// (albumtype, catalognum, language, mb_artistid, mb_albumartistid):
// a one-element list serializes to exactly one comment, which is what
// this function already writes.
func stringField(row FieldMapping, v *string) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, *v)}
}

// listField encodes a []string TagSet member: nil is absent and
// deletes every style; a non-nil slice, empty or not, deletes only
// the writable styles. A non-empty slice writes one comment per
// element under each writable style, in list order; an empty slice
// writes nothing at all (C2).
func listField(row FieldMapping, values []string) KeyWrite {
	if values == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	kw := KeyWrite{Field: row.Field, Deletes: writableKeys(row)}
	for _, s := range row.Styles {
		if !s.Writable {
			continue
		}
		for _, v := range values {
			kw.Writes = append(kw.Writes, Comment{Name: s.Key, Value: v})
		}
	}
	return kw
}

// intField encodes a *int TagSet member. Numeric fields carry no
// empty state (C2): the only two reachable states are nil (absent)
// and a value, formatted through str() the way mediafile's
// StorageStyle.serialize does for anything that is not a float or a
// bool.
func intField(row FieldMapping, v *int) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, strconv.Itoa(*v))}
}

// boolField encodes comp's *bool member as "1" or "0", never
// "true"/"false" (mediafile:535, StorageStyle.serialize).
func boolField(row FieldMapping, v *bool) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	s := "0"
	if *v {
		s = "1"
	}
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, s)}
}

// rgGainField encodes rg_track_gain and rg_album_gain: two decimal
// places, the default float_places, with the " dB" suffix that is the
// only suffix in the FLAC table (mediafile:836, :846).
func rgGainField(row FieldMapping, v *float64) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, fmt.Sprintf("%.2f dB", *v))}
}

// rgPeakField encodes rg_track_peak and rg_album_peak: six decimal
// places, no suffix (mediafile:856, :864).
func rgPeakField(row FieldMapping, v *float64) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, fmt.Sprintf("%.6f", *v))}
}

// r128Field encodes r128_track_gain and r128_album_gain as an integer
// string, the value shifted by fraction_bits=8: round(value * 2**8)
// (mediafile:874, :881; __init__.py:870-884). r128_track_gain = -1.5
// writes R128_TRACK_GAIN=-384.
func r128Field(row FieldMapping, v *float64) KeyWrite {
	if v == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	shifted := int64(math.Round(*v * 256))
	return KeyWrite{Field: row.Field, Deletes: writableKeys(row), Writes: repeatComment(row, strconv.FormatInt(shifted, 10))}
}

// serializeDate composes the DATE (or ORIGINALDATE) and, when row
// carries a YEAR style, the bare-year fallback, from a year/month/day
// triple (mediafile/fields.py:182-201). A nil year is absent and
// deletes every style the row carries, including the YEAR fallback,
// together (`_set_date_tuple`, :146-149, :188-190). A non-nil year,
// zero or not, is a value: the year is zero-padded to four digits,
// the month is appended zero-padded to two digits only when truthy,
// and the day only when both the month and the day are truthy. A nil
// month or day is treated as falsy, the same as zero, because tagio's
// write entry point takes a whole tag set rather than mediafile's
// read-modify-write update (Task 8's deviation note).
func serializeDate(row FieldMapping, year, month, day *int) KeyWrite {
	if year == nil {
		return KeyWrite{Field: row.Field, Deletes: allKeys(row)}
	}
	y, m, d := *year, valueOr(month), valueOr(day)

	date := fmt.Sprintf("%04d", y)
	if m != 0 {
		date += fmt.Sprintf("-%02d", m)
		if d != 0 {
			date += fmt.Sprintf("-%02d", d)
		}
	}

	kw := KeyWrite{Field: row.Field, Deletes: writableKeys(row)}
	for _, s := range row.Styles {
		if !s.Writable {
			continue
		}
		switch s.Key {
		case "YEAR":
			kw.Writes = append(kw.Writes, Comment{Name: s.Key, Value: strconv.Itoa(y)})
		default: // DATE, ORIGINALDATE
			kw.Writes = append(kw.Writes, Comment{Name: s.Key, Value: date})
		}
	}
	return kw
}

// valueOr returns 0 for a nil *int and the pointed-to value
// otherwise, the falsy-when-absent rule serializeDate applies to
// month and day.
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
// (ruling 10). Values are valid UTF-8 by construction on write, so
// this only ever runs on read.
func replaceInvalidUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && size <= 1 {
			b.WriteRune(utf8.RuneError)
			i++
			continue
		}
		b.WriteRune(r)
		i += size
	}
	return b.String()
}
