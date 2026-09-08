package tagio

import "slices"

// Field names one of TagSet's 69 writable members, spelled the way
// corpus/README.md's handoff section and INV:315-385 both spell it
// (mediafile's own field name, snake_case). It is the identifier
// mapping's rows key on and the value WriteOrder yields.
type Field string

// The 69 writable fields, one constant per TagSet member. Names come
// from corpus/README.md:2381 onward, the handoff section C1 names as
// the single source of record; INV:315-385 supplies each field's
// Vorbis keys and mediafile citation, never its name.
const (
	FieldAcoustIDFingerprint Field = "acoustid_fingerprint"
	FieldAcoustIDID          Field = "acoustid_id"
	FieldAlbum               Field = "album"
	FieldAlbumartist         Field = "albumartist"
	FieldAlbumartistCredit   Field = "albumartist_credit"
	FieldAlbumartistSort     Field = "albumartist_sort"
	FieldAlbumartists        Field = "albumartists"
	FieldAlbumartistsCredit  Field = "albumartists_credit"
	FieldAlbumartistsSort    Field = "albumartists_sort"
	FieldAlbumdisambig       Field = "albumdisambig"
	FieldAlbumstatus         Field = "albumstatus"
	FieldAlbumtype           Field = "albumtype"
	FieldAlbumtypes          Field = "albumtypes"
	FieldArrangers           Field = "arrangers"
	FieldArtist              Field = "artist"
	FieldArtistCredit        Field = "artist_credit"
	FieldArtistSort          Field = "artist_sort"
	FieldArtists             Field = "artists"
	FieldArtistsCredit       Field = "artists_credit"
	FieldArtistsSort         Field = "artists_sort"
	FieldAsin                Field = "asin"
	FieldBarcode             Field = "barcode"
	FieldBpm                 Field = "bpm"
	FieldCatalognum          Field = "catalognum"
	FieldComments            Field = "comments"
	FieldComp                Field = "comp"
	FieldComposerSort        Field = "composer_sort"
	FieldComposers           Field = "composers"
	FieldCountry             Field = "country"
	FieldYear                Field = "year"
	FieldMonth               Field = "month"
	FieldDay                 Field = "day"
	FieldDisc                Field = "disc"
	FieldDisctitle           Field = "disctitle"
	FieldDisctotal           Field = "disctotal"
	FieldEncoder             Field = "encoder"
	FieldGenres              Field = "genres"
	FieldGrouping            Field = "grouping"
	FieldInitialKey          Field = "initial_key"
	FieldIsrc                Field = "isrc"
	FieldLabel               Field = "label"
	FieldLanguage            Field = "language"
	FieldLyricists           Field = "lyricists"
	FieldLyrics              Field = "lyrics"
	FieldMBAlbumartistid     Field = "mb_albumartistid"
	FieldMBAlbumartistids    Field = "mb_albumartistids"
	FieldMBAlbumid           Field = "mb_albumid"
	FieldMBArtistid          Field = "mb_artistid"
	FieldMBArtistids         Field = "mb_artistids"
	FieldMBReleasegroupid    Field = "mb_releasegroupid"
	FieldMBReleasetrackid    Field = "mb_releasetrackid"
	FieldMBTrackid           Field = "mb_trackid"
	FieldMBWorkid            Field = "mb_workid"
	FieldMedia               Field = "media"
	FieldOriginalYear        Field = "original_year"
	FieldOriginalMonth       Field = "original_month"
	FieldOriginalDay         Field = "original_day"
	FieldR128AlbumGain       Field = "r128_album_gain"
	FieldR128TrackGain       Field = "r128_track_gain"
	FieldRemixers            Field = "remixers"
	FieldRgAlbumGain         Field = "rg_album_gain"
	FieldRgAlbumPeak         Field = "rg_album_peak"
	FieldRgTrackGain         Field = "rg_track_gain"
	FieldRgTrackPeak         Field = "rg_track_peak"
	FieldScript              Field = "script"
	FieldSubtitle            Field = "subtitle"
	FieldTitle               Field = "title"
	FieldTrack               Field = "track"
	FieldTracktotal          Field = "tracktotal"
)

// TagSet is tagio's own vocabulary for the 69 fields mediafile 0.17.0
// (the version corpus/manifest.json pins) writes to a FLAC container:
// the intersection of beets' Item._media_fields and
// MediaFile.sorted_fields(), matching corpus/README.md's handoff
// section key for key (C1). No consumer's type appears here (ruling
// 3); a reader or writer that wants a beets Item or a matcher type
// converts at its own boundary.
//
// Every member carries a presence bit rather than a Go zero value,
// because a Go string's zero value and a Go slice's zero value are
// the same in Go and are not the same on disk: MediaFile.update
// deletes a key on None and sets it on everything else, so a string
// field's empty value writes a comment with an empty value while a
// list field's empty value writes no comment at all (C2). A string
// member is *string: nil is absent, a non-nil pointer to "" is
// present-but-empty, and a non-nil pointer to a non-empty string is a
// value. A list member is a plain slice, not a pointer: a nil slice
// is absent and a non-nil zero-length slice is present-but-empty,
// which is idiomatic Go's own nil/non-nil-empty distinction rather
// than a pointer indirection, so EmptyList's non-nil []string{} carries
// meaning a bare append never would. A numeric or bool member is a
// pointer (*int, *float64, *bool) with only two reachable states, nil
// absent and a value, because mediafile's ledger records no "empty
// int" or "empty bool": zero is a value like any other.
//
// Both nil-vs-empty distinctions are read and written directly, with
// no second lookup: a caller checking presence reads the one field
// TagSet already exposes for it.
type TagSet struct {
	AcoustIDFingerprint *string
	AcoustIDID          *string
	Album               *string
	Albumartist         *string
	AlbumartistCredit   *string
	AlbumartistSort     *string
	Albumartists        []string
	AlbumartistsCredit  []string
	AlbumartistsSort    []string
	Albumdisambig       *string
	Albumstatus         *string
	Albumtype           *string
	Albumtypes          []string
	Arrangers           []string
	Artist              *string
	ArtistCredit        *string
	ArtistSort          *string
	Artists             []string
	ArtistsCredit       []string
	ArtistsSort         []string
	Asin                *string
	Barcode             *string
	Bpm                 *int
	Catalognum          *string
	Comments            *string
	Comp                *bool
	ComposerSort        *string
	Composers           []string
	Country             *string
	Year                *int
	Month               *int
	Day                 *int
	Disc                *int
	Disctitle           *string
	Disctotal           *int
	Encoder             *string
	Genres              []string
	Grouping            *string
	InitialKey          *string
	Isrc                *string
	Label               *string
	Language            *string
	Lyricists           []string
	Lyrics              *string
	MBAlbumartistid     *string
	MBAlbumartistids    []string
	MBAlbumid           *string
	MBArtistid          *string
	MBArtistids         []string
	MBReleasegroupid    *string
	MBReleasetrackid    *string
	MBTrackid           *string
	MBWorkid            *string
	Media               *string
	OriginalYear        *int
	OriginalMonth       *int
	OriginalDay         *int
	R128AlbumGain       *float64
	R128TrackGain       *float64
	Remixers            []string
	RgAlbumGain         *float64
	RgAlbumPeak         *float64
	RgTrackGain         *float64
	RgTrackPeak         *float64
	Script              *string
	Subtitle            *string
	Title               *string
	Track               *int
	Tracktotal          *int
}

// EmptyString returns a present-but-empty value for a TagSet string
// member, a non-nil pointer distinct from the nil pointer that means
// absent (C2).
func EmptyString() *string {
	s := ""
	return &s
}

// EmptyList returns a present-but-empty value for a TagSet list
// member, a non-nil zero-length slice distinct from the nil slice
// that means absent (C2). go-conventions prefers a nil slice by
// default; this is the one case its own exception names, because the
// nil/empty distinction is the fact this function exists to carry.
func EmptyList() []string {
	return []string{}
}

// Style is one Vorbis comment key a field's value writes to (or, for
// a read-only style, only reads from).
type Style struct {
	// Key is the Vorbis comment field name, uppercase per RFC 9639
	// section 8.6 convention.
	Key string
	// Writable reports whether a write emits this style. A read-only
	// style is read by mediafile and deleted when the field is
	// deleted, but a writer never emits it (INV:312-313).
	Writable bool
}

// FieldMapping is one row of Mapping: a field's Vorbis keys in
// mediafile declaration order, each marked writable or read-only, and
// the mediafile line the row was read from (C16).
type FieldMapping struct {
	Field    Field
	Styles   []Style
	Citation string
}

// mapping is the 69-key table, one row per TagSet member, in the
// fixed order MediaFile.sorted_fields() writes them (C3): lexicographic
// by field name, except year, month and day (and their original_
// counterparts) sort as if named date0, date1 and date2, so a date
// triple writes in year-month-day order rather than alphabetical
// day-month-year order (mediafile:322-330, the pinned venv's own
// _field_sort_name). That substitution is also what resolves the
// three colliding pairs: sorted_fields()'s plain lexicographic order
// already places albumtype before albumtypes, mb_artistid before
// mb_artistids, and mb_albumartistid before mb_albumartistids, because
// each singular name is a strict prefix of its plural and a prefix
// sorts first, so the plural writes last and wins the shared key
// (spec:401-410, INV surprise 4).
//
// mapping is unexported so that no caller outside this file can
// reorder or mutate the table WriteOrder and Serialize both walk; a
// caller that needs the rows reads them from Mapping, which hands
// back a copy.
var mapping = []FieldMapping{
	{FieldAcoustIDFingerprint, []Style{{"ACOUSTID_FINGERPRINT", true}}, "mediafile:817"},
	{FieldAcoustIDID, []Style{{"ACOUSTID_ID", true}}, "mediafile:823"},
	{FieldAlbum, []Style{{"ALBUM", true}}, "mediafile:410"},
	{FieldAlbumartist, []Style{{"ALBUM ARTIST", true}, {"ALBUM_ARTIST", true}, {"ALBUMARTIST", true}}, "mediafile:542,543,544"},
	{FieldAlbumartistCredit, []Style{{"ALBUMARTIST_CREDIT", true}}, "mediafile:732"},
	{FieldAlbumartistSort, []Style{{"ALBUMARTISTSORT", true}}, "mediafile:588"},
	{FieldAlbumartists, []Style{{"ALBUMARTISTS", true}, {"ALBUM_ARTISTS", true}, {"ALBUM ARTISTS", false}}, "mediafile:554,555,556"},
	{FieldAlbumartistsCredit, []Style{{"ALBUMARTISTS_CREDIT", true}}, "mediafile:738"},
	{FieldAlbumartistsSort, []Style{{"ALBUMARTISTS_SORT", true}}, "mediafile:744"},
	{FieldAlbumdisambig, []Style{{"MUSICBRAINZ_ALBUMCOMMENT", true}}, "mediafile:680"},
	{FieldAlbumstatus, []Style{{"RELEASESTATUS", true}, {"MUSICBRAINZ_ALBUMSTATUS", true}}, "mediafile:665,666"},
	{FieldAlbumtype, []Style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, "mediafile:562,563"},
	{FieldAlbumtypes, []Style{{"RELEASETYPE", true}, {"MUSICBRAINZ_ALBUMTYPE", true}}, "mediafile:562,563"},
	{FieldArrangers, []Style{{"ARRANGER", true}}, "mediafile:444"},
	{FieldArtist, []Style{{"ARTIST", true}}, "mediafile:398"},
	{FieldArtistCredit, []Style{{"ARTIST_CREDIT", true}}, "mediafile:714"},
	{FieldArtistSort, []Style{{"ARTISTSORT", true}}, "mediafile:581"},
	{FieldArtists, []Style{{"ARTISTS", true}}, "mediafile:404"},
	{FieldArtistsCredit, []Style{{"ARTISTS_CREDIT", true}}, "mediafile:720"},
	{FieldArtistsSort, []Style{{"ARTISTS_SORT", true}}, "mediafile:726"},
	{FieldAsin, []Style{{"ASIN", true}}, "mediafile:594"},
	{FieldBarcode, []Style{{"BARCODE", true}, {"UPC", false}, {"EAN/UPN", false}, {"EAN", false}, {"UPN", false}}, "mediafile:616,617,618,619,620"},
	{FieldBpm, []Style{{"BPM", true}}, "mediafile:528"},
	{FieldCatalognum, []Style{{"CATALOGNUMBER", true}, {"CATALOGID", false}, {"DISCOGS_CATALOG", false}}, "mediafile:604,605,606"},
	{FieldComments, []Style{{"DESCRIPTION", true}, {"COMMENT", true}}, "mediafile:514,515"},
	{FieldComp, []Style{{"COMPILATION", true}}, "mediafile:535"},
	{FieldComposerSort, []Style{{"COMPOSERSORT", true}}, "mediafile:438"},
	{FieldComposers, []Style{{"COMPOSER", true}}, "mediafile:431"},
	{FieldCountry, []Style{{"RELEASECOUNTRY", true}}, "mediafile:659"},
	{FieldYear, []Style{{"DATE", true}, {"YEAR", true}}, "mediafile:693,688,690"},
	{FieldMonth, []Style{{"DATE", true}, {"YEAR", true}}, "mediafile:694,688,690"},
	{FieldDay, []Style{{"DATE", true}, {"YEAR", true}}, "mediafile:695,688,690"},
	{FieldDisc, []Style{{"DISC", true}, {"DISCNUMBER", true}}, "mediafile:480,481"},
	{FieldDisctitle, []Style{{"DISCSUBTITLE", true}}, "mediafile:632"},
	{FieldDisctotal, []Style{{"DISCTOTAL", true}, {"DISCC", true}, {"TOTALDISCS", true}}, "mediafile:488,489,490"},
	{FieldEncoder, []Style{{"ENCODEDBY", true}, {"ENCODER", true}}, "mediafile:638,639"},
	{FieldGenres, []Style{{"GENRE", true}}, "mediafile:416"},
	{FieldGrouping, []Style{{"GROUPING", true}}, "mediafile:452"},
	{FieldInitialKey, []Style{{"INITIALKEY", true}}, "mediafile:888"},
	{FieldIsrc, []Style{{"ISRC", true}}, "mediafile:626"},
	{FieldLabel, []Style{{"LABEL", true}, {"PUBLISHER", true}}, "mediafile:574,575"},
	{FieldLanguage, []Style{{"LANGUAGE", true}}, "mediafile:651"},
	{FieldLyricists, []Style{{"LYRICIST", true}}, "mediafile:424"},
	{FieldLyrics, []Style{{"LYRICS", true}}, "mediafile:504"},
	{FieldMBAlbumartistid, []Style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, "mediafile:801"},
	{FieldMBAlbumartistids, []Style{{"MUSICBRAINZ_ALBUMARTISTID", true}}, "mediafile:801"},
	{FieldMBAlbumid, []Style{{"MUSICBRAINZ_ALBUMID", true}}, "mediafile:782"},
	{FieldMBArtistid, []Style{{"MUSICBRAINZ_ARTISTID", true}}, "mediafile:788"},
	{FieldMBArtistids, []Style{{"MUSICBRAINZ_ARTISTID", true}}, "mediafile:788"},
	{FieldMBReleasegroupid, []Style{{"MUSICBRAINZ_RELEASEGROUPID", true}}, "mediafile:809"},
	{FieldMBReleasetrackid, []Style{{"MUSICBRAINZ_RELEASETRACKID", true}}, "mediafile:770"},
	{FieldMBTrackid, []Style{{"MUSICBRAINZ_TRACKID", true}}, "mediafile:764"},
	{FieldMBWorkid, []Style{{"MUSICBRAINZ_WORKID", true}}, "mediafile:776"},
	{FieldMedia, []Style{{"MEDIA", true}}, "mediafile:673"},
	{FieldOriginalYear, []Style{{"ORIGINALDATE", true}}, "mediafile:706,702"},
	{FieldOriginalMonth, []Style{{"ORIGINALDATE", true}}, "mediafile:707,702"},
	{FieldOriginalDay, []Style{{"ORIGINALDATE", true}}, "mediafile:708,702"},
	{FieldR128AlbumGain, []Style{{"R128_ALBUM_GAIN", true}}, "mediafile:881"},
	{FieldR128TrackGain, []Style{{"R128_TRACK_GAIN", true}}, "mediafile:874"},
	{FieldRemixers, []Style{{"REMIXER", true}}, "mediafile:750"},
	{FieldRgAlbumGain, []Style{{"REPLAYGAIN_ALBUM_GAIN", true}}, "mediafile:846"},
	{FieldRgAlbumPeak, []Style{{"REPLAYGAIN_ALBUM_PEAK", true}}, "mediafile:864"},
	{FieldRgTrackGain, []Style{{"REPLAYGAIN_TRACK_GAIN", true}}, "mediafile:836"},
	{FieldRgTrackPeak, []Style{{"REPLAYGAIN_TRACK_PEAK", true}}, "mediafile:856"},
	{FieldScript, []Style{{"SCRIPT", true}}, "mediafile:645"},
	{FieldSubtitle, []Style{{"SUBTITLE", true}}, "mediafile:457"},
	{FieldTitle, []Style{{"TITLE", true}}, "mediafile:392"},
	{FieldTrack, []Style{{"TRACK", true}, {"TRACKNUMBER", true}}, "mediafile:463,464"},
	{FieldTracktotal, []Style{{"TRACKTOTAL", true}, {"TRACKC", true}, {"TOTALTRACKS", true}}, "mediafile:471,472,473"},
}

// WriteOrder returns the 69 field identifiers in mapping's fixed
// write order (C3). It walks mapping rather than a map, so calling it
// repeatedly always yields the same sequence; a Go map's range order
// is randomized per range statement and would make this
// non-deterministic across calls.
func WriteOrder() []Field {
	order := make([]Field, len(mapping))
	for i, row := range mapping {
		order[i] = row.Field
	}
	return order
}

// Mapping returns a copy of the 69-row field-to-Vorbis-key table, in
// MediaFile.sorted_fields() write order (C3). Each call clones the
// underlying rows, so mutating the returned slice never reorders or
// changes what WriteOrder and Serialize see.
func Mapping() []FieldMapping {
	return slices.Clone(mapping)
}
