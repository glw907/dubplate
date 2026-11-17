package tagio

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
// than a pointer indirection, so a non-nil []string{} carries
// meaning a bare append never would. A numeric or bool member is a
// pointer (*int, *float64, *bool) with only two reachable states, nil
// absent and a value, because mediafile's ledger records no "empty
// int" or "empty bool": zero is a value like any other.
//
// Both nil-vs-empty distinctions are read and written directly, with
// no second lookup: a caller checking presence reads the one field
// TagSet already exposes for it.
//
// Every member carries a JSON tag naming the ledger's own field name,
// which is also its row's name in the field table, with no omitempty (R11): the
// conformance instrument fills a TagSet by unmarshalling a ledger
// row's filtered applied-fields object straight into it and projects
// a read-back TagSet by marshalling it, and every ledger row carries
// all 69 keys, absent values included as explicit nulls, so an
// omitted member would drop an absent field from the round trip.
// TestTagSetJSONTagsTieToFieldNames ties every tag to the field it
// names (TestTagSetJSONTagsCarryNoOmitempty grades the no-omitempty
// half), so a wrong tag fails loudly rather than silently swapping two
// fields' values while both sides of a comparison agree.
type TagSet struct {
	AcoustIDFingerprint *string  `json:"acoustid_fingerprint"`
	AcoustIDID          *string  `json:"acoustid_id"`
	Album               *string  `json:"album"`
	Albumartist         *string  `json:"albumartist"`
	AlbumartistCredit   *string  `json:"albumartist_credit"`
	AlbumartistSort     *string  `json:"albumartist_sort"`
	Albumartists        []string `json:"albumartists"`
	AlbumartistsCredit  []string `json:"albumartists_credit"`
	AlbumartistsSort    []string `json:"albumartists_sort"`
	Albumdisambig       *string  `json:"albumdisambig"`
	Albumstatus         *string  `json:"albumstatus"`
	Albumtype           *string  `json:"albumtype"`
	Albumtypes          []string `json:"albumtypes"`
	Arrangers           []string `json:"arrangers"`
	Artist              *string  `json:"artist"`
	ArtistCredit        *string  `json:"artist_credit"`
	ArtistSort          *string  `json:"artist_sort"`
	Artists             []string `json:"artists"`
	ArtistsCredit       []string `json:"artists_credit"`
	ArtistsSort         []string `json:"artists_sort"`
	Asin                *string  `json:"asin"`
	Barcode             *string  `json:"barcode"`
	Bpm                 *int     `json:"bpm"`
	Catalognum          *string  `json:"catalognum"`
	Comments            *string  `json:"comments"`
	Comp                *bool    `json:"comp"`
	ComposerSort        *string  `json:"composer_sort"`
	Composers           []string `json:"composers"`
	Country             *string  `json:"country"`
	Year                *int     `json:"year"`
	Month               *int     `json:"month"`
	Day                 *int     `json:"day"`
	Disc                *int     `json:"disc"`
	Disctitle           *string  `json:"disctitle"`
	Disctotal           *int     `json:"disctotal"`
	Encoder             *string  `json:"encoder"`
	Genres              []string `json:"genres"`
	Grouping            *string  `json:"grouping"`
	InitialKey          *string  `json:"initial_key"`
	Isrc                *string  `json:"isrc"`
	Label               *string  `json:"label"`
	Language            *string  `json:"language"`
	Lyricists           []string `json:"lyricists"`
	Lyrics              *string  `json:"lyrics"`
	MBAlbumartistid     *string  `json:"mb_albumartistid"`
	MBAlbumartistids    []string `json:"mb_albumartistids"`
	MBAlbumid           *string  `json:"mb_albumid"`
	MBArtistid          *string  `json:"mb_artistid"`
	MBArtistids         []string `json:"mb_artistids"`
	MBReleasegroupid    *string  `json:"mb_releasegroupid"`
	MBReleasetrackid    *string  `json:"mb_releasetrackid"`
	MBTrackid           *string  `json:"mb_trackid"`
	MBWorkid            *string  `json:"mb_workid"`
	Media               *string  `json:"media"`
	OriginalYear        *int     `json:"original_year"`
	OriginalMonth       *int     `json:"original_month"`
	OriginalDay         *int     `json:"original_day"`
	R128AlbumGain       *float64 `json:"r128_album_gain"`
	R128TrackGain       *float64 `json:"r128_track_gain"`
	Remixers            []string `json:"remixers"`
	RgAlbumGain         *float64 `json:"rg_album_gain"`
	RgAlbumPeak         *float64 `json:"rg_album_peak"`
	RgTrackGain         *float64 `json:"rg_track_gain"`
	RgTrackPeak         *float64 `json:"rg_track_peak"`
	Script              *string  `json:"script"`
	Subtitle            *string  `json:"subtitle"`
	Title               *string  `json:"title"`
	Track               *int     `json:"track"`
	Tracktotal          *int     `json:"tracktotal"`
}
