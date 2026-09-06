package shadow

import "github.com/glw907/dubplate/internal/engine/matcher"

// itemBag projects fields back into the beets field-name bag that
// cassette.Case's Input.Items carries, which is ProjectItemFields' own
// input shape (matcher/info.go:189-211).
//
// The names are spelled here because matcher's own field-name constants
// are unexported (info.go:156-176). Two decoders over one field bag
// drift silently (info.go:148-150), so the round-trip test in
// flac_test.go is what keeps these nineteen names in step with that
// table rather than a comment asking a reader to check.
func itemBag(fields matcher.ItemFields) map[string]any {
	return map[string]any{
		"artist":        fields.Artist,
		"album":         fields.Album,
		"albumartist":   fields.AlbumArtist,
		"year":          fields.Year,
		"disctotal":     fields.DiscTotal,
		"mb_albumid":    fields.MBAlbumID,
		"label":         fields.Label,
		"barcode":       fields.Barcode,
		"catalognum":    fields.CatalogNum,
		"country":       fields.Country,
		"media":         fields.Media,
		"albumdisambig": fields.AlbumDisambig,
		"data_source":   fields.DataSource,
		"length":        fields.Length,
		"title":         fields.Title,
		"track":         fields.Track,
		"mb_trackid":    fields.MBTrackID,
		"disc":          fields.Disc,
		"comp":          fields.Comp,
	}
}
