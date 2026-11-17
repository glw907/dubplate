package tagio

import "strconv"

// style is one Vorbis comment key a field's value writes to or, for a
// read-only style, only reads from. A read-only style is read by
// mediafile and deleted when the field is deleted, but a writer never
// emits it (INV:312-313).
type style struct {
	key      string
	writable bool
}

// field is one row of the field table: a TagSet member's ledger name,
// which is also its JSON tag; its Vorbis keys in mediafile declaration
// order; the mediafile line those keys were read from; and the codec
// that reads the member from a file's comments and writes it back.
type field struct {
	name   Field
	styles []style
	cite   string
	codec  codec
}

// codec reads one TagSet member from a file's comments and writes it
// back. write reports present false for an absent member, which a
// write turns into a delete of every key the field owns. The month
// and day rows carry a zero codec, because the year row's date codec
// reads and writes all three components of the one comment that
// holds them.
type codec struct {
	read  func(ts *TagSet, cs []Comment, styles []style)
	write func(ts *TagSet, styles []style) (writes []Comment, present bool)
}

// scalar is the codec of a pointer member: the value readStyles picks
// goes through cast on the way in, and the member's value goes through
// format, once under every writable style, on the way out.
func scalar[T any](member func(*TagSet) **T, cast func(string) T, format func(T) string) codec {
	return codec{
		read: func(ts *TagSet, cs []Comment, styles []style) {
			if value, ok := readStyles(cs, styles); ok {
				v := cast(value)
				*member(ts) = &v
			}
		},
		write: func(ts *TagSet, styles []style) ([]Comment, bool) {
			v := *member(ts)
			if v == nil {
				return nil, false
			}
			return repeatComment(styles, format(*v)), true
		},
	}
}

// text is a string member's codec: safe_cast to str is the identity,
// and the value writes unchanged.
func text(member func(*TagSet) **string) codec {
	identity := func(s string) string { return s }
	return scalar(member, identity, identity)
}

// integer is an int member's codec. It reads through safe_cast's
// leading [+-]?[0-9]+ prefix, so TRACK=3/12 reads 3 and BPM=fast
// reads 0, and writes through str(), the way mediafile's
// StorageStyle.serialize formats anything that is not a float or a
// bool.
func integer(member func(*TagSet) **int) codec {
	return scalar(member, safeCastInt, strconv.Itoa)
}

// boolean is comp's codec: bool(int(val)) on the way in, "1" or "0" on
// the way out, never "true" or "false" (mediafile:535).
func boolean(member func(*TagSet) **bool) codec {
	return scalar(member, safeCastBool, formatBool)
}

// gain is a ReplayGain gain member's codec, whose styles carry the
// " dB" suffix (mediafile:836, :846).
func gain(member func(*TagSet) **float64) codec {
	return scalar(member, readGain, formatGain)
}

// peak is a ReplayGain peak member's codec, with no suffix
// (mediafile:856, :864).
func peak(member func(*TagSet) **float64) codec {
	return scalar(member, safeCastFloat, formatPeak)
}

// q8 is an R128 member's codec: Q8 fixed point on disk
// (mediafile:874, :881).
func q8(member func(*TagSet) **float64) codec {
	return scalar(member, readQ8, formatQ8)
}

// list is a list member's codec. A nil slice is absent; a non-nil
// slice writes one comment per element under each writable style, in
// list order, so an empty one writes nothing at all.
func list(member func(*TagSet) *[]string) codec {
	return codec{
		read: func(ts *TagSet, cs []Comment, styles []style) {
			*member(ts) = readList(cs, styles)
		},
		write: func(ts *TagSet, styles []style) ([]Comment, bool) {
			values := *member(ts)
			if values == nil {
				return nil, false
			}
			return listComments(styles, values), true
		},
	}
}

// date is the codec of a year, month and day triple, which one DATE or
// ORIGINALDATE comment carries together; the year decides presence.
func date(year, month, day func(*TagSet) **int) codec {
	return codec{
		read: func(ts *TagSet, cs []Comment, styles []style) {
			*year(ts), *month(ts), *day(ts) = readDate(cs, styles)
		},
		write: func(ts *TagSet, styles []style) ([]Comment, bool) {
			y := *year(ts)
			if y == nil {
				return nil, false
			}
			return dateComments(styles, *y, valueOr(*month(ts)), valueOr(*day(ts))), true
		},
	}
}

// fields is the 69-row table, one row per TagSet member in TagSet's own
// order, which is the fixed order MediaFile.sorted_fields() writes
// them: lexicographic by field name, except year, month and day (and
// their original_ counterparts) sort as if named date0, date1 and
// date2, so a date triple writes in year-month-day order rather than
// alphabetical day-month-year order (mediafile:322-330, the pinned
// venv's own _field_sort_name).
//
// The order also resolves the three colliding pairs, albumtype and
// albumtypes, mb_artistid and mb_artistids, and mb_albumartistid and
// mb_albumartistids: each singular name is a strict prefix of its
// plural and sorts first, so the plural writes last and wins the key
// the two share.
var fields = []field{
	{"acoustid_fingerprint", []style{{"ACOUSTID_FINGERPRINT", true}}, "mediafile:817", text(func(t *TagSet) **string { return &t.AcoustIDFingerprint })},
	{"acoustid_id", []style{{"ACOUSTID_ID", true}}, "mediafile:823", text(func(t *TagSet) **string { return &t.AcoustIDID })},
	{"album", []style{{"ALBUM", true}}, "mediafile:410", text(func(t *TagSet) **string { return &t.Album })},
	{"albumartist", []style{{"ALBUM ARTIST", true}, {"ALBUM_ARTIST", true}, {"ALBUMARTIST", true}}, "mediafile:542,543,544", text(func(t *TagSet) **string { return &t.Albumartist })},
	{"albumartist_credit", []style{{"ALBUMARTIST_CREDIT", true}}, "mediafile:732", text(func(t *TagSet) **string { return &t.AlbumartistCredit })},
	{"albumartist_sort", []style{{"ALBUMARTISTSORT", true}}, "mediafile:588", text(func(t *TagSet) **string { return &t.AlbumartistSort })},
	{"albumartists", []style{{"ALBUMARTISTS", true}, {"ALBUM_ARTISTS", true}, {"ALBUM ARTISTS", false}}, "mediafile:554,555,556", list(func(t *TagSet) *[]string { return &t.Albumartists })},
	{"albumartists_credit", []style{{"ALBUMARTISTS_CREDIT", true}}, "mediafile:738", list(func(t *TagSet) *[]string { return &t.AlbumartistsCredit })},
	{"albumartists_sort", []style{{"ALBUMARTISTS_SORT", true}}, "mediafile:744", list(func(t *TagSet) *[]string { return &t.AlbumartistsSort })},
	{"albumdisambig", []style{{"MUSICBRAINZ_ALBUMCOMMENT", true}}, "mediafile:680", text(func(t *TagSet) **string { return &t.Albumdisambig })},
	{"albumstatus", []style{{"RELEASESTATUS", true}, {"MUSICBRAINZ_ALBUMSTATUS", true}}, "mediafile:665,666", text(func(t *TagSet) **string { return &t.Albumstatus })},
	{"albumtype", []style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, "mediafile:562,563", text(func(t *TagSet) **string { return &t.Albumtype })},
	{"albumtypes", []style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, "mediafile:562,563", list(func(t *TagSet) *[]string { return &t.Albumtypes })},
	{"arrangers", []style{{"ARRANGER", true}}, "mediafile:444", list(func(t *TagSet) *[]string { return &t.Arrangers })},
	{"artist", []style{{"ARTIST", true}}, "mediafile:398", text(func(t *TagSet) **string { return &t.Artist })},
	{"artist_credit", []style{{"ARTIST_CREDIT", true}}, "mediafile:714", text(func(t *TagSet) **string { return &t.ArtistCredit })},
	{"artist_sort", []style{{"ARTISTSORT", true}}, "mediafile:581", text(func(t *TagSet) **string { return &t.ArtistSort })},
	{"artists", []style{{"ARTISTS", true}}, "mediafile:404", list(func(t *TagSet) *[]string { return &t.Artists })},
	{"artists_credit", []style{{"ARTISTS_CREDIT", true}}, "mediafile:720", list(func(t *TagSet) *[]string { return &t.ArtistsCredit })},
	{"artists_sort", []style{{"ARTISTS_SORT", true}}, "mediafile:726", list(func(t *TagSet) *[]string { return &t.ArtistsSort })},
	{"asin", []style{{"ASIN", true}}, "mediafile:594", text(func(t *TagSet) **string { return &t.Asin })},
	{"barcode", []style{{"BARCODE", true}, {"UPC", false}, {"EAN/UPN", false}, {"EAN", false}, {"UPN", false}}, "mediafile:616,617,618,619,620", text(func(t *TagSet) **string { return &t.Barcode })},
	{"bpm", []style{{"BPM", true}}, "mediafile:528", integer(func(t *TagSet) **int { return &t.Bpm })},
	{"catalognum", []style{{"CATALOGNUMBER", true}, {"CATALOGID", false}, {"DISCOGS_CATALOG", false}}, "mediafile:604,605,606", text(func(t *TagSet) **string { return &t.Catalognum })},
	{"comments", []style{{"DESCRIPTION", true}, {"COMMENT", true}}, "mediafile:514,515", text(func(t *TagSet) **string { return &t.Comments })},
	{"comp", []style{{"COMPILATION", true}}, "mediafile:535", boolean(func(t *TagSet) **bool { return &t.Comp })},
	{"composer_sort", []style{{"COMPOSERSORT", true}}, "mediafile:438", text(func(t *TagSet) **string { return &t.ComposerSort })},
	{"composers", []style{{"COMPOSER", true}}, "mediafile:431", list(func(t *TagSet) *[]string { return &t.Composers })},
	{"country", []style{{"RELEASECOUNTRY", true}}, "mediafile:659", text(func(t *TagSet) **string { return &t.Country })},
	{"year", []style{{"DATE", true}, {"YEAR", true}}, "mediafile:693,688,690", date(
		func(t *TagSet) **int { return &t.Year },
		func(t *TagSet) **int { return &t.Month },
		func(t *TagSet) **int { return &t.Day },
	)},
	{"month", []style{{"DATE", true}, {"YEAR", true}}, "mediafile:694,688,690", codec{}},
	{"day", []style{{"DATE", true}, {"YEAR", true}}, "mediafile:695,688,690", codec{}},
	{"disc", []style{{"DISC", true}, {"DISCNUMBER", true}}, "mediafile:480,481", integer(func(t *TagSet) **int { return &t.Disc })},
	{"disctitle", []style{{"DISCSUBTITLE", true}}, "mediafile:632", text(func(t *TagSet) **string { return &t.Disctitle })},
	{"disctotal", []style{{"DISCTOTAL", true}, {"DISCC", true}, {"TOTALDISCS", true}}, "mediafile:488,489,490", integer(func(t *TagSet) **int { return &t.Disctotal })},
	{"encoder", []style{{"ENCODEDBY", true}, {"ENCODER", true}}, "mediafile:638,639", text(func(t *TagSet) **string { return &t.Encoder })},
	{"genres", []style{{"GENRE", true}}, "mediafile:416", list(func(t *TagSet) *[]string { return &t.Genres })},
	{"grouping", []style{{"GROUPING", true}}, "mediafile:452", text(func(t *TagSet) **string { return &t.Grouping })},
	{"initial_key", []style{{"INITIALKEY", true}}, "mediafile:888", text(func(t *TagSet) **string { return &t.InitialKey })},
	{"isrc", []style{{"ISRC", true}}, "mediafile:626", text(func(t *TagSet) **string { return &t.Isrc })},
	{"label", []style{{"LABEL", true}, {"PUBLISHER", true}}, "mediafile:574,575", text(func(t *TagSet) **string { return &t.Label })},
	{"language", []style{{"LANGUAGE", true}}, "mediafile:651", text(func(t *TagSet) **string { return &t.Language })},
	{"lyricists", []style{{"LYRICIST", true}}, "mediafile:424", list(func(t *TagSet) *[]string { return &t.Lyricists })},
	{"lyrics", []style{{"LYRICS", true}}, "mediafile:504", text(func(t *TagSet) **string { return &t.Lyrics })},
	{"mb_albumartistid", []style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, "mediafile:801", text(func(t *TagSet) **string { return &t.MBAlbumartistid })},
	{"mb_albumartistids", []style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, "mediafile:801", list(func(t *TagSet) *[]string { return &t.MBAlbumartistids })},
	{"mb_albumid", []style{{"MUSICBRAINZ_ALBUMID", true}}, "mediafile:782", text(func(t *TagSet) **string { return &t.MBAlbumid })},
	{"mb_artistid", []style{{"MUSICBRAINZ_ARTISTID", true}}, "mediafile:788", text(func(t *TagSet) **string { return &t.MBArtistid })},
	{"mb_artistids", []style{{"MUSICBRAINZ_ARTISTID", true}}, "mediafile:788", list(func(t *TagSet) *[]string { return &t.MBArtistids })},
	{"mb_releasegroupid", []style{{"MUSICBRAINZ_RELEASEGROUPID", true}}, "mediafile:809", text(func(t *TagSet) **string { return &t.MBReleasegroupid })},
	{"mb_releasetrackid", []style{{"MUSICBRAINZ_RELEASETRACKID", true}}, "mediafile:770", text(func(t *TagSet) **string { return &t.MBReleasetrackid })},
	{"mb_trackid", []style{{"MUSICBRAINZ_TRACKID", true}}, "mediafile:764", text(func(t *TagSet) **string { return &t.MBTrackid })},
	{"mb_workid", []style{{"MUSICBRAINZ_WORKID", true}}, "mediafile:776", text(func(t *TagSet) **string { return &t.MBWorkid })},
	{"media", []style{{"MEDIA", true}}, "mediafile:673", text(func(t *TagSet) **string { return &t.Media })},
	{"original_year", []style{{"ORIGINALDATE", true}}, "mediafile:706,702", date(
		func(t *TagSet) **int { return &t.OriginalYear },
		func(t *TagSet) **int { return &t.OriginalMonth },
		func(t *TagSet) **int { return &t.OriginalDay },
	)},
	{"original_month", []style{{"ORIGINALDATE", true}}, "mediafile:707,702", codec{}},
	{"original_day", []style{{"ORIGINALDATE", true}}, "mediafile:708,702", codec{}},
	{"r128_album_gain", []style{{"R128_ALBUM_GAIN", true}}, "mediafile:881", q8(func(t *TagSet) **float64 { return &t.R128AlbumGain })},
	{"r128_track_gain", []style{{"R128_TRACK_GAIN", true}}, "mediafile:874", q8(func(t *TagSet) **float64 { return &t.R128TrackGain })},
	{"remixers", []style{{"REMIXER", true}}, "mediafile:750", list(func(t *TagSet) *[]string { return &t.Remixers })},
	{"rg_album_gain", []style{{"REPLAYGAIN_ALBUM_GAIN", true}}, "mediafile:846", gain(func(t *TagSet) **float64 { return &t.RgAlbumGain })},
	{"rg_album_peak", []style{{"REPLAYGAIN_ALBUM_PEAK", true}}, "mediafile:864", peak(func(t *TagSet) **float64 { return &t.RgAlbumPeak })},
	{"rg_track_gain", []style{{"REPLAYGAIN_TRACK_GAIN", true}}, "mediafile:836", gain(func(t *TagSet) **float64 { return &t.RgTrackGain })},
	{"rg_track_peak", []style{{"REPLAYGAIN_TRACK_PEAK", true}}, "mediafile:856", peak(func(t *TagSet) **float64 { return &t.RgTrackPeak })},
	{"script", []style{{"SCRIPT", true}}, "mediafile:645", text(func(t *TagSet) **string { return &t.Script })},
	{"subtitle", []style{{"SUBTITLE", true}}, "mediafile:457", text(func(t *TagSet) **string { return &t.Subtitle })},
	{"title", []style{{"TITLE", true}}, "mediafile:392", text(func(t *TagSet) **string { return &t.Title })},
	{"track", []style{{"TRACK", true}, {"TRACKNUMBER", true}}, "mediafile:463,464", integer(func(t *TagSet) **int { return &t.Track })},
	{"tracktotal", []style{{"TRACKTOTAL", true}, {"TRACKC", true}, {"TOTALTRACKS", true}}, "mediafile:471,472,473", integer(func(t *TagSet) **int { return &t.Tracktotal })},
}
