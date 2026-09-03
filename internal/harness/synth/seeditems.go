package synth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// ProjectItems reads doc, a WS/2 release document's raw bytes (a
// captured case's responses/001.json, verbatim), and derives the
// baseline item set a real rip of that release would carry: one item
// per track, release order, mirroring [buildItems]'s own default shape
// but sourced from a parsed captured document instead of a
// [ReleaseSpec]. overrides then perturbs that baseline through the
// exact same [applyOverrides] vocabulary (field, missing, extra) a
// synthetic spec uses, so the mistagged and partial-rip axes reach
// their divergence the same way every other case does.
//
// This is the sanctioned reader of a captured response
// (corpus/README.md, "The seed-items tool"): a captured document
// carries per-track fields (title, length, recording id) no MBID or
// axis label can invent, and this is the only code path in the corpus
// tooling that opens one.
func ProjectItems(doc []byte, overrides []Override, allowAlbumIDOverride bool) ([]cassette.ItemFields, error) {
	var rel capturedRelease
	if err := json.Unmarshal(doc, &rel); err != nil {
		return nil, fmt.Errorf("synth: seed-items: parse release document: %w", err)
	}
	if rel.ID == "" {
		return nil, errors.New("synth: seed-items: release document: id: missing")
	}
	if len(rel.Media) == 0 {
		return nil, errors.New("synth: seed-items: release document: media: empty")
	}

	albumArtist := creditString(rel.ArtistCredit)
	year := releaseYear(rel.Date)
	multiDisc := len(rel.Media) > 1
	comp := releaseIsVariousArtists(rel.ArtistCredit)

	var items []cassette.ItemFields
	for mi, m := range rel.Media {
		for _, tr := range m.Tracks {
			title := tr.Title
			if title == "" {
				title = tr.Recording.Title
			}
			artist := trackArtist(tr, albumArtist)
			item := cassette.ItemFields{
				"album":       rel.Title,
				"albumartist": albumArtist,
				"artist":      artist,
				"length":      float64(tr.Length) / 1000,
				"mb_albumid":  rel.ID,
				"mb_trackid":  tr.Recording.ID,
				"title":       title,
				"track":       int(tr.Position),
				"year":        year,
			}
			if comp {
				item["comp"] = true
			}
			if multiDisc {
				item["disc"] = mi + 1
			}
			items = append(items, item)
		}
	}

	return applyOverrides(items, overrides, rel.ID, allowAlbumIDOverride)
}

// flexPosition decodes a WS/2 track or medium "position" field: the
// live MusicBrainz API renders it as a JSON number, but
// internal/harness/synth's own built documents (marshalRelease, via
// wsTrack.Position) render it as a JSON string, and both shapes appear
// in the committed corpus. The decoder accepts either rather than
// picking a side.
type flexPosition int

// UnmarshalJSON implements [json.Unmarshaler].
func (p *flexPosition) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*p = flexPosition(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("position: neither a number nor a string: %w", err)
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("position: %q is not an integer: %w", s, err)
	}
	*p = flexPosition(parsed)
	return nil
}

// capturedRelease is the WS/2 release document decode shape
// [ProjectItems] and [ExtractFacts] both read: the fields a baseline
// item set needs plus the structural facts Task 3's fact-verified axes
// confirm (Status, ReleaseGroup, Aliases). Every other field a maximal
// `inc=` lookup returns (genres, cover-art-archive, tags, ...) is
// simply ignored by encoding/json rather than modeled here.
type capturedRelease struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	Date         string               `json:"date"`
	Status       string               `json:"status"`
	ArtistCredit []capturedCredit     `json:"artist-credit"`
	Media        []capturedMedium     `json:"media"`
	ReleaseGroup capturedReleaseGroup `json:"release-group"`
	Aliases      []capturedAlias      `json:"aliases"`
}

// capturedReleaseGroup is a release's parent release-group, read for
// its own first-release-date: absent an `original_date` override, a
// release's year and its group's first-release-date agree, and a
// year-penalty case's whole point is the two diverging
// (corpus/README.md, "The synth tool").
type capturedReleaseGroup struct {
	FirstReleaseDate string `json:"first-release-date"`
}

// capturedAlias is one documented alias, on either a release itself
// (rare; a release's own top-level `aliases`) or an artist-credit
// entry's artist. Type and Locale decode to "" when the document
// carries a JSON null for either, matching a real WS/2 document's own
// shape for an alias the API has not classified.
type capturedAlias struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Locale string `json:"locale"`
}

// capturedCredit is one artist-credit entry: a display name, and the
// joinphrase WS/2 inserts before the next entry ("Marvin Gaye" + " & "
// + "Tammi Terrell"), which [creditString] reassembles into the same
// single display string beets itself builds from a multi-entry credit.
type capturedCredit struct {
	Name       string         `json:"name"`
	JoinPhrase string         `json:"joinphrase"`
	Artist     capturedArtist `json:"artist"`
}

// capturedArtist is one artist-credit entry's artist object.
type capturedArtist struct {
	ID      string          `json:"id"`
	Aliases []capturedAlias `json:"aliases"`
}

// capturedMedium is one disc's track list.
type capturedMedium struct {
	Tracks []capturedTrack `json:"tracks"`
}

// capturedTrack is one WS/2 track. Title is the track-level title the
// live API returns; a document built by [marshalRelease] always leaves
// it empty, so [ProjectItems] falls back to Recording.Title, matching
// beets' own get_tracks_from_medium behavior.
type capturedTrack struct {
	Position     flexPosition      `json:"position"`
	Title        string            `json:"title"`
	Length       int               `json:"length"` // milliseconds
	ArtistCredit []capturedCredit  `json:"artist-credit"`
	Recording    capturedRecording `json:"recording"`
}

// capturedRecording is a track's nested recording.
type capturedRecording struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	ArtistCredit []capturedCredit `json:"artist-credit"`
}

// creditString reassembles credits into the one display string beets
// builds from a WS/2 artist-credit array: each entry's name followed
// by its own joinphrase, concatenated in order. "" for an empty credit
// list.
func creditString(credits []capturedCredit) string {
	var b strings.Builder
	for _, c := range credits {
		b.WriteString(c.Name)
		b.WriteString(c.JoinPhrase)
	}
	return b.String()
}

// releaseIsVariousArtists reports whether credits' lead artist carries
// MusicBrainz's well-known Various Artists id, the same signal
// [singleArtistCredit] writes for a synthetic VA release and the sole
// one beets' musicbrainz plugin reads to set AlbumInfo.va.
func releaseIsVariousArtists(credits []capturedCredit) bool {
	return len(credits) > 0 && credits[0].Artist.ID == variousArtistsMBID
}

// trackArtist returns tr's own performer credit string: its
// track-level credit when the live API populates one, else its
// recording's credit (the shape a document built by [marshalRelease]
// always uses), else fallback (the release's own albumartist).
func trackArtist(tr capturedTrack, fallback string) string {
	if name := creditString(tr.ArtistCredit); name != "" {
		return name
	}
	if name := creditString(tr.Recording.ArtistCredit); name != "" {
		return name
	}
	return fallback
}

// Facts is the structural fact set a captured release document
// carries that Task 3's fact-verified axes (reissue/remaster,
// multi-disc/box-set, non-Latin, two-release) confirm through
// `harness seed-items --facts` rather than by reading responses/
// directly (corpus/README.md, "The seed-items tool"; the plan's
// Global Constraints).
type Facts struct {
	Date                  string
	ReleaseGroupFirstDate string
	Status                string
	DiscTrackCounts       []int
	ReleaseAliases        []FactAlias
	ArtistCreditAliases   []ArtistAliases
}

// ArtistAliases is one artist-credit entry's own documented aliases,
// under the entry's display name. [ExtractFacts] omits an entry whose
// artist carries none, so every entry in this slice has at least one
// alias to report.
type ArtistAliases struct {
	Artist  string
	Aliases []FactAlias
}

// FactAlias is one documented alias: a name plus its type and locale,
// when the release document carries them ("" otherwise).
type FactAlias struct {
	Name   string
	Type   string
	Locale string
}

// ExtractFacts reads doc, a captured release document's raw bytes, and
// returns its structural facts: the release date, its release-group's
// own first-release-date, its status, each disc's track count, and its
// documented aliases (the release's own top-level aliases plus each
// artist-credit entry's artist aliases, where the document carries
// either). ExtractFacts is the read-only half of the seed-items reader
// ([ProjectItems] is the write half): both parse the same captured
// document, but ExtractFacts never touches Input.Items, so
// `harness seed-items --facts` can run against a case whose items are
// already projected without perturbing it.
func ExtractFacts(doc []byte) (*Facts, error) {
	var rel capturedRelease
	if err := json.Unmarshal(doc, &rel); err != nil {
		return nil, fmt.Errorf("synth: seed-items: parse release document: %w", err)
	}

	discs := make([]int, len(rel.Media))
	for i, m := range rel.Media {
		discs[i] = len(m.Tracks)
	}

	var artistAliases []ArtistAliases
	for _, c := range rel.ArtistCredit {
		if len(c.Artist.Aliases) == 0 {
			continue
		}
		artistAliases = append(artistAliases, ArtistAliases{
			Artist:  c.Name,
			Aliases: convertAliases(c.Artist.Aliases),
		})
	}

	return &Facts{
		Date:                  rel.Date,
		ReleaseGroupFirstDate: rel.ReleaseGroup.FirstReleaseDate,
		Status:                rel.Status,
		DiscTrackCounts:       discs,
		ReleaseAliases:        convertAliases(rel.Aliases),
		ArtistCreditAliases:   artistAliases,
	}, nil
}

// convertAliases converts a decoded document's own alias shape into
// [FactAlias], the facts-package's stable, format-ready shape.
func convertAliases(aliases []capturedAlias) []FactAlias {
	out := make([]FactAlias, len(aliases))
	for i, a := range aliases {
		out[i] = FactAlias(a)
	}
	return out
}

// FormatFacts renders f as the deterministic text `harness seed-items
// --facts` prints to stdout. Every fact prints on its own line so a
// golden-file diff isolates exactly which fact changed; "aliases: none"
// prints as an explicit single line rather than an omitted section,
// so a document that documents no alias at all is never
// indistinguishable from one ExtractFacts failed to read.
func FormatFacts(f *Facts) string {
	var b strings.Builder
	fmt.Fprintf(&b, "release date: %s\n", f.Date)
	fmt.Fprintf(&b, "release-group first-release-date: %s\n", f.ReleaseGroupFirstDate)
	fmt.Fprintf(&b, "release status: %s\n", f.Status)
	fmt.Fprintf(&b, "discs: %d\n", len(f.DiscTrackCounts))
	for i, n := range f.DiscTrackCounts {
		fmt.Fprintf(&b, "  disc %d: %d tracks\n", i+1, n)
	}
	writeAliases(&b, f)
	return b.String()
}

// writeAliases writes f's aliases section: a single "aliases: none"
// line when the document documents no alias at all (neither the
// release itself nor any artist-credit entry), else a header followed
// by the release's own alias list (or "none") and one indented block
// per artist-credit entry that carries at least one alias.
func writeAliases(b *strings.Builder, f *Facts) {
	if len(f.ReleaseAliases) == 0 && len(f.ArtistCreditAliases) == 0 {
		b.WriteString("aliases: none\n")
		return
	}

	b.WriteString("aliases:\n")
	b.WriteString("  release:\n")
	writeAliasList(b, "    ", f.ReleaseAliases)
	for _, ac := range f.ArtistCreditAliases {
		fmt.Fprintf(b, "  artist-credit %q:\n", ac.Artist)
		writeAliasList(b, "    ", ac.Aliases)
	}
}

// writeAliasList writes one alias per line at indent, each carrying
// its type and/or locale in parentheses when the document names them;
// an empty list writes a single "none" line rather than nothing, the
// same explicit-absence rule [writeAliases] applies at the section
// level.
func writeAliasList(b *strings.Builder, indent string, aliases []FactAlias) {
	if len(aliases) == 0 {
		fmt.Fprintf(b, "%snone\n", indent)
		return
	}
	for _, a := range aliases {
		b.WriteString(indent)
		b.WriteString(a.Name)
		var parts []string
		if a.Type != "" {
			parts = append(parts, "type: "+a.Type)
		}
		if a.Locale != "" {
			parts = append(parts, "locale: "+a.Locale)
		}
		if len(parts) > 0 {
			b.WriteString(" (")
			b.WriteString(strings.Join(parts, ", "))
			b.WriteString(")")
		}
		b.WriteString("\n")
	}
}
