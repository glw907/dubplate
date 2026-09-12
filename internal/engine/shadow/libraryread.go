package shadow

import (
	"fmt"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/tagio"
)

// readFLAC reads one FLAC file's matcher input: the eighteen fields
// matcher.ItemFields carries from a file (info.go:124-152), read one
// way, for scoring only.
//
// It reads through tagio, rung 8's tag I/O package
// (internal/engine/tagio/doc.go): one tagio.Parse per file for the
// container, tagio.Tags over the parsed Container for the 69-key tag
// set, and Container.StreamInfo for the track length (R1's fold of
// the 8b handoff's second half). An absent or unparseable tag takes
// the field's Go zero value, which is ProjectItemFields' own
// typed-default rule (info.go:178-188) and beets' Item.get behavior.
// A malformed container is an error and never a partly filled bag,
// because a capture that cannot read every entry it listed is
// discarded whole rather than scored short.
func readFLAC(path string) (matcher.ItemFields, error) {
	c, err := tagio.Parse(path)
	if err != nil {
		return matcher.ItemFields{}, fmt.Errorf("parse %s: %w", path, err)
	}
	return projectItemFields(tagio.Tags(c), c.StreamInfo), nil
}

// projectItemFields projects ts and info into the matcher's input, one
// row per mediafile MediaField of pinned mediafile 0.17.0, the
// version corpus/manifest.json pins, mirroring
// [matcher.ProjectItemFields]'s own typed-default rule: an absent
// tagio member takes its Go zero value.
//
// DataSource has no row: a grep of pinned mediafile for data_source
// returns nothing, since beets keeps it as a database flexible
// attribute rather than in the file, so an inbox file always projects
// the empty string ProjectItemFields' typed-default rule already
// models (info.go:115-123).
func projectItemFields(ts tagio.TagSet, info tagio.StreamInfo) matcher.ItemFields {
	return matcher.ItemFields{
		Title:         stringVal(ts.Title),
		Artist:        stringVal(ts.Artist),
		Album:         stringVal(ts.Album),
		AlbumArtist:   stringVal(ts.Albumartist),
		Track:         intVal(ts.Track),
		Disc:          intVal(ts.Disc),
		DiscTotal:     intVal(ts.Disctotal),
		Comp:          boolVal(ts.Comp),
		Label:         stringVal(ts.Label),
		CatalogNum:    stringVal(ts.Catalognum),
		Barcode:       stringVal(ts.Barcode),
		Country:       stringVal(ts.Country),
		Media:         stringVal(ts.Media),
		AlbumDisambig: stringVal(ts.Albumdisambig),
		MBTrackID:     stringVal(ts.MBTrackid),
		MBAlbumID:     stringVal(ts.MBAlbumid),
		Year:          intVal(ts.Year),
		Length:        trackLength(info),
	}
}

// trackLength is the track length in seconds, which is zero for a
// stream declaring no sample rate rather than a division by zero.
func trackLength(info tagio.StreamInfo) float64 {
	if info.SampleRate == 0 {
		return 0
	}
	return float64(info.TotalSamples) / float64(info.SampleRate)
}

// stringVal reads a TagSet *string member as ProjectItemFields' own
// typed-default rule does: an absent member is the empty string.
func stringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// intVal reads a TagSet *int member as ProjectItemFields' own
// typed-default rule does: an absent member is zero.
func intVal(n *int) int {
	if n == nil {
		return 0
	}
	return *n
}

// boolVal reads a TagSet *bool member as ProjectItemFields' own
// typed-default rule does: an absent member is false.
func boolVal(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}
