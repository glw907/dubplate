package tagio

import "strconv"

// style is one Vorbis comment key a field's value writes to or, for a
// read-only style, only reads from. A read-only style is read by
// mediafile and deleted when the field is deleted, but a writer never
// emits it.
type style struct {
	key      string
	writable bool
}

// field is one row of the field table: a TagSet member's ledger name,
// which is also its JSON tag; its Vorbis keys in mediafile declaration
// order; and the codec that reads the member from a file's comments
// and writes it back.
type field struct {
	name   string
	styles []style
	codec  codec
}

// codec reads one TagSet member from a file's comments and writes it
// back. write reports present false for an absent member, which a
// write turns into a delete of every key the field owns.
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
			return listComments(styles, []string{format(*v)}), true
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
	{"acoustid_fingerprint", []style{{"ACOUSTID_FINGERPRINT", true}}, text(func(t *TagSet) **string { return &t.AcoustIDFingerprint })},                               // mediafile:817
	{"acoustid_id", []style{{"ACOUSTID_ID", true}}, text(func(t *TagSet) **string { return &t.AcoustIDID })},                                                          // mediafile:823
	{"album", []style{{"ALBUM", true}}, text(func(t *TagSet) **string { return &t.Album })},                                                                           // mediafile:410
	{"albumartist", []style{{"ALBUM ARTIST", true}, {"ALBUM_ARTIST", true}, {"ALBUMARTIST", true}}, text(func(t *TagSet) **string { return &t.Albumartist })},         // mediafile:542, :543, :544
	{"albumartist_credit", []style{{"ALBUMARTIST_CREDIT", true}}, text(func(t *TagSet) **string { return &t.AlbumartistCredit })},                                     // mediafile:732
	{"albumartist_sort", []style{{"ALBUMARTISTSORT", true}}, text(func(t *TagSet) **string { return &t.AlbumartistSort })},                                            // mediafile:588
	{"albumartists", []style{{"ALBUMARTISTS", true}, {"ALBUM_ARTISTS", true}, {"ALBUM ARTISTS", false}}, list(func(t *TagSet) *[]string { return &t.Albumartists })},  // mediafile:554, :555, :556
	{"albumartists_credit", []style{{"ALBUMARTISTS_CREDIT", true}}, list(func(t *TagSet) *[]string { return &t.AlbumartistsCredit })},                                 // mediafile:738
	{"albumartists_sort", []style{{"ALBUMARTISTS_SORT", true}}, list(func(t *TagSet) *[]string { return &t.AlbumartistsSort })},                                       // mediafile:744
	{"albumdisambig", []style{{"MUSICBRAINZ_ALBUMCOMMENT", true}}, text(func(t *TagSet) **string { return &t.Albumdisambig })},                                        // mediafile:680
	{"albumstatus", []style{{"RELEASESTATUS", true}, {"MUSICBRAINZ_ALBUMSTATUS", true}}, text(func(t *TagSet) **string { return &t.Albumstatus })},                    // mediafile:665, :666
	{"albumtype", []style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, text(func(t *TagSet) **string { return &t.Albumtype })},                            // mediafile:562, :563
	{"albumtypes", []style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, list(func(t *TagSet) *[]string { return &t.Albumtypes })},                         // mediafile:562, :563
	{"arrangers", []style{{"ARRANGER", true}}, list(func(t *TagSet) *[]string { return &t.Arrangers })},                                                               // mediafile:444
	{"artist", []style{{"ARTIST", true}}, text(func(t *TagSet) **string { return &t.Artist })},                                                                        // mediafile:398
	{"artist_credit", []style{{"ARTIST_CREDIT", true}}, text(func(t *TagSet) **string { return &t.ArtistCredit })},                                                    // mediafile:714
	{"artist_sort", []style{{"ARTISTSORT", true}}, text(func(t *TagSet) **string { return &t.ArtistSort })},                                                           // mediafile:581
	{"artists", []style{{"ARTISTS", true}}, list(func(t *TagSet) *[]string { return &t.Artists })},                                                                    // mediafile:404
	{"artists_credit", []style{{"ARTISTS_CREDIT", true}}, list(func(t *TagSet) *[]string { return &t.ArtistsCredit })},                                                // mediafile:720
	{"artists_sort", []style{{"ARTISTS_SORT", true}}, list(func(t *TagSet) *[]string { return &t.ArtistsSort })},                                                      // mediafile:726
	{"asin", []style{{"ASIN", true}}, text(func(t *TagSet) **string { return &t.Asin })},                                                                              // mediafile:594
	{"barcode", []style{{"BARCODE", true}, {"UPC", false}, {"EAN/UPN", false}, {"EAN", false}, {"UPN", false}}, text(func(t *TagSet) **string { return &t.Barcode })}, // mediafile:616, :617, :618, :619, :620
	{"bpm", []style{{"BPM", true}}, integer(func(t *TagSet) **int { return &t.Bpm })},                                                                                 // mediafile:528
	{"catalognum", []style{{"CATALOGNUMBER", true}, {"CATALOGID", false}, {"DISCOGS_CATALOG", false}}, text(func(t *TagSet) **string { return &t.Catalognum })},       // mediafile:604, :605, :606
	{"comments", []style{{"DESCRIPTION", true}, {"COMMENT", true}}, text(func(t *TagSet) **string { return &t.Comments })},                                            // mediafile:514, :515
	{"comp", []style{{"COMPILATION", true}}, boolean(func(t *TagSet) **bool { return &t.Comp })},                                                                      // mediafile:535
	{"composer_sort", []style{{"COMPOSERSORT", true}}, text(func(t *TagSet) **string { return &t.ComposerSort })},                                                     // mediafile:438
	{"composers", []style{{"COMPOSER", true}}, list(func(t *TagSet) *[]string { return &t.Composers })},                                                               // mediafile:431
	{"country", []style{{"RELEASECOUNTRY", true}}, text(func(t *TagSet) **string { return &t.Country })},                                                              // mediafile:659
	{"year", []style{{"DATE", true}, {"YEAR", true}}, date( // mediafile:693, :688, :690
		func(t *TagSet) **int { return &t.Year },
		func(t *TagSet) **int { return &t.Month },
		func(t *TagSet) **int { return &t.Day },
	)},
	// The month and day rows carry neither styles nor a codec: the year
	// row reads and writes the one DATE comment all three share.
	{"month", nil, codec{}}, // mediafile:694
	{"day", nil, codec{}},   // mediafile:695
	{"disc", []style{{"DISC", true}, {"DISCNUMBER", true}}, integer(func(t *TagSet) **int { return &t.Disc })},                                 // mediafile:480, :481
	{"disctitle", []style{{"DISCSUBTITLE", true}}, text(func(t *TagSet) **string { return &t.Disctitle })},                                     // mediafile:632
	{"disctotal", []style{{"DISCTOTAL", true}, {"DISCC", true}, {"TOTALDISCS", true}}, integer(func(t *TagSet) **int { return &t.Disctotal })}, // mediafile:488, :489, :490
	{"encoder", []style{{"ENCODEDBY", true}, {"ENCODER", true}}, text(func(t *TagSet) **string { return &t.Encoder })},                         // mediafile:638, :639
	{"genres", []style{{"GENRE", true}}, list(func(t *TagSet) *[]string { return &t.Genres })},                                                 // mediafile:416
	{"grouping", []style{{"GROUPING", true}}, text(func(t *TagSet) **string { return &t.Grouping })},                                           // mediafile:452
	{"initial_key", []style{{"INITIALKEY", true}}, text(func(t *TagSet) **string { return &t.InitialKey })},                                    // mediafile:888
	{"isrc", []style{{"ISRC", true}}, text(func(t *TagSet) **string { return &t.Isrc })},                                                       // mediafile:626
	{"label", []style{{"LABEL", true}, {"PUBLISHER", true}}, text(func(t *TagSet) **string { return &t.Label })},                               // mediafile:574, :575
	{"language", []style{{"LANGUAGE", true}}, text(func(t *TagSet) **string { return &t.Language })},                                           // mediafile:651
	{"lyricists", []style{{"LYRICIST", true}}, list(func(t *TagSet) *[]string { return &t.Lyricists })},                                        // mediafile:424
	{"lyrics", []style{{"LYRICS", true}}, text(func(t *TagSet) **string { return &t.Lyrics })},                                                 // mediafile:504
	{"mb_albumartistid", []style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, text(func(t *TagSet) **string { return &t.MBAlbumartistid })},           // mediafile:801
	{"mb_albumartistids", []style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, list(func(t *TagSet) *[]string { return &t.MBAlbumartistids })},        // mediafile:801
	{"mb_albumid", []style{{"MUSICBRAINZ_ALBUMID", true}}, text(func(t *TagSet) **string { return &t.MBAlbumid })},                             // mediafile:782
	{"mb_artistid", []style{{"MUSICBRAINZ_ARTISTID", true}}, text(func(t *TagSet) **string { return &t.MBArtistid })},                          // mediafile:788
	{"mb_artistids", []style{{"MUSICBRAINZ_ARTISTID", true}}, list(func(t *TagSet) *[]string { return &t.MBArtistids })},                       // mediafile:788
	{"mb_releasegroupid", []style{{"MUSICBRAINZ_RELEASEGROUPID", true}}, text(func(t *TagSet) **string { return &t.MBReleasegroupid })},        // mediafile:809
	{"mb_releasetrackid", []style{{"MUSICBRAINZ_RELEASETRACKID", true}}, text(func(t *TagSet) **string { return &t.MBReleasetrackid })},        // mediafile:770
	{"mb_trackid", []style{{"MUSICBRAINZ_TRACKID", true}}, text(func(t *TagSet) **string { return &t.MBTrackid })},                             // mediafile:764
	{"mb_workid", []style{{"MUSICBRAINZ_WORKID", true}}, text(func(t *TagSet) **string { return &t.MBWorkid })},                                // mediafile:776
	{"media", []style{{"MEDIA", true}}, text(func(t *TagSet) **string { return &t.Media })},                                                    // mediafile:673
	{"original_year", []style{{"ORIGINALDATE", true}}, date( // mediafile:706, :702
		func(t *TagSet) **int { return &t.OriginalYear },
		func(t *TagSet) **int { return &t.OriginalMonth },
		func(t *TagSet) **int { return &t.OriginalDay },
	)},
	// As with month and day, the original_year row carries these two.
	{"original_month", nil, codec{}}, // mediafile:707
	{"original_day", nil, codec{}},   // mediafile:708
	{"r128_album_gain", []style{{"R128_ALBUM_GAIN", true}}, q8(func(t *TagSet) **float64 { return &t.R128AlbumGain })},                              // mediafile:881
	{"r128_track_gain", []style{{"R128_TRACK_GAIN", true}}, q8(func(t *TagSet) **float64 { return &t.R128TrackGain })},                              // mediafile:874
	{"remixers", []style{{"REMIXER", true}}, list(func(t *TagSet) *[]string { return &t.Remixers })},                                                // mediafile:750
	{"rg_album_gain", []style{{"REPLAYGAIN_ALBUM_GAIN", true}}, gain(func(t *TagSet) **float64 { return &t.RgAlbumGain })},                          // mediafile:846
	{"rg_album_peak", []style{{"REPLAYGAIN_ALBUM_PEAK", true}}, peak(func(t *TagSet) **float64 { return &t.RgAlbumPeak })},                          // mediafile:864
	{"rg_track_gain", []style{{"REPLAYGAIN_TRACK_GAIN", true}}, gain(func(t *TagSet) **float64 { return &t.RgTrackGain })},                          // mediafile:836
	{"rg_track_peak", []style{{"REPLAYGAIN_TRACK_PEAK", true}}, peak(func(t *TagSet) **float64 { return &t.RgTrackPeak })},                          // mediafile:856
	{"script", []style{{"SCRIPT", true}}, text(func(t *TagSet) **string { return &t.Script })},                                                      // mediafile:645
	{"subtitle", []style{{"SUBTITLE", true}}, text(func(t *TagSet) **string { return &t.Subtitle })},                                                // mediafile:457
	{"title", []style{{"TITLE", true}}, text(func(t *TagSet) **string { return &t.Title })},                                                         // mediafile:392
	{"track", []style{{"TRACK", true}, {"TRACKNUMBER", true}}, integer(func(t *TagSet) **int { return &t.Track })},                                  // mediafile:463, :464
	{"tracktotal", []style{{"TRACKTOTAL", true}, {"TRACKC", true}, {"TOTALTRACKS", true}}, integer(func(t *TagSet) **int { return &t.Tracktotal })}, // mediafile:471, :472, :473
}
