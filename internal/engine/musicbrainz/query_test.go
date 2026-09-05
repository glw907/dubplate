package musicbrainz

import (
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// renderRow is one row of the rendering fixture's escape, trim and lower
// tables.
type renderRow struct {
	Name string `json:"name"`
	In   string `json:"in"`
	Want string `json:"want"`
}

// formatRow is one row of the rendering fixture's format table, which
// carries the field the term is wrapped in.
type formatRow struct {
	Name  string `json:"name"`
	Field string `json:"field"`
	In    string `json:"in"`
	Want  string `json:"want"`
}

type renderingFixtures struct {
	Escape []renderRow `json:"escape"`
	Format []formatRow `json:"format"`
	Trim   []renderRow `json:"trim"`
	Lower  []renderRow `json:"lower"`
}

// extraTagRow is one row of the extra-tag fixture: rung 3's tenth
// declared corpus-coverage gap, since no committed profile writes a
// musicbrainz.extra_tags block at all.
type extraTagRow struct {
	Name      string               `json:"name"`
	ExtraTags []string             `json:"extra_tags"`
	Items     []matcher.ItemFields `json:"items"`
	Want      string               `json:"want"`
}

// TestReleaseIncludesJoinsToTheCommittedIncludeSet pins the sixteen-element
// RELEASE_INCLUDES set in the order and joining the client issues it.
// The want string is a literal rather than a read of
// internal/harness/mbrequest's own copy, which the engine could not
// import in any case: the two definitions are independent on purpose and
// the conformance run is what proves they agree.
//
// Mutation: dropping "genres" from releaseIncludes fails this test, and
// misses all 303 committed lookups in the conformance run.
func TestReleaseIncludesJoinsToTheCommittedIncludeSet(t *testing.T) {
	const want = "artists+media+recordings+release-groups+labels+" +
		"artist-credits+aliases+recording-level-rels+work-rels+work-level-rels+" +
		"artist-rels+isrcs+url-rels+release-rels+genres+tags"

	if got := strings.Join(releaseIncludes, "+"); got != want {
		t.Errorf("inc = %q, want %q", got, want)
	}
	if len(releaseIncludes) != 16 {
		t.Errorf("releaseIncludes has %d elements, want 16", len(releaseIncludes))
	}
}

// TestEscapeLucene walks LUCENE_SPECIAL_CHAR_PAT's nineteen characters
// and the eleven punctuation marks notably absent from it, which the
// committed corpus proves reach the wire unescaped.
//
// Mutation: removing "-" from luceneSpecial fails the hyphen row here and
// the arid row in TestAlbumSearchForRendersTheCommittedShapes.
func TestEscapeLucene(t *testing.T) {
	fx := loadTestdata[renderingFixtures](t, "query/rendering.json")
	if len(fx.Escape) != 30 {
		t.Fatalf("escape fixture has %d rows, want 30 (19 escaped plus 11 untouched)", len(fx.Escape))
	}
	for _, tt := range fx.Escape {
		t.Run(tt.Name, func(t *testing.T) {
			if got := escapeLucene(tt.In); got != tt.Want {
				t.Errorf("escapeLucene(%q) = %q, want %q", tt.In, got, tt.Want)
			}
		})
	}
}

// TestFormatSearchTerm proves the order format_search_term applies its
// steps in: an empty or whitespace-only term drops out before anything
// else, and the field wrapper is added after the escape, so the wrapper's
// own colon and parentheses survive while a colon or parenthesis inside
// the term does not.
//
// Mutation: escaping after the wrapper instead of before fails the
// "wrapper's own colon and parens" row.
func TestFormatSearchTerm(t *testing.T) {
	fx := loadTestdata[renderingFixtures](t, "query/rendering.json")
	for _, tt := range fx.Format {
		t.Run(tt.Name, func(t *testing.T) {
			if got := formatSearchTerm(tt.Field, tt.In); got != tt.Want {
				t.Errorf("formatSearchTerm(%q, %q) = %q, want %q", tt.Field, tt.In, got, tt.Want)
			}
		})
	}
}

// TestTrimSearchTerm names the four C0 code points Python's str.strip
// treats as whitespace and Go's unicode.IsSpace does not, plus two the
// two agree on.
//
// Mutation: reverting trimSearchTerm to strings.TrimSpace fails the
// U+001C row.
func TestTrimSearchTerm(t *testing.T) {
	fx := loadTestdata[renderingFixtures](t, "query/rendering.json")
	for _, tt := range fx.Trim {
		t.Run(tt.Name, func(t *testing.T) {
			if got := trimSearchTerm(tt.In); got != tt.Want {
				t.Errorf("trimSearchTerm(%q) = %q, want %q", tt.In, got, tt.Want)
			}
		})
	}
}

// TestLowerSearchTerm records the case-mapping divergence rather than
// closing it: Go lowers U+0130 to a single "i" and CPython to "i" plus
// U+0307, and the query path has no strip to rescue the difference. The
// asserted row is the stable member of the 56-code-point set, the other
// 55 being a Unicode version skew that moves when either toolchain does.
//
// Mutation: strings.ToUpper in place of strings.ToLower fails both rows.
func TestLowerSearchTerm(t *testing.T) {
	fx := loadTestdata[renderingFixtures](t, "query/rendering.json")
	for _, tt := range fx.Lower {
		t.Run(tt.Name, func(t *testing.T) {
			if got := lowerSearchTerm(tt.In); got != tt.Want {
				t.Errorf("lowerSearchTerm(%q) = %q, want %q", tt.In, got, tt.Want)
			}
		})
	}
}

// TestVALikelyTakesEachDisjunctAlone isolates va_likely's three
// disjuncts, each fired by itself against a row where the other two do
// not hold, plus the row where none does.
//
// Mutation: dropping the any(item.comp) disjunct fails the comp row.
func TestVALikelyTakesEachDisjunctAlone(t *testing.T) {
	plain := []matcher.ItemFields{{Artist: "Northline", Album: "Static Coast"}}
	compFlagged := []matcher.ItemFields{
		{Artist: "Northline", Album: "Static Coast"},
		{Artist: "Northline", Album: "Static Coast", Comp: true},
	}
	tests := []struct {
		name      string
		likelies  matcher.Likelies
		consensus map[string]bool
		items     []matcher.ItemFields
		want      bool
	}{
		{
			name:      "no artist consensus",
			likelies:  matcher.Likelies{Artist: "Northline"},
			consensus: map[string]bool{"artist": false},
			items:     plain,
			want:      true,
		},
		{
			name:      "the likely artist names a various-artists placeholder",
			likelies:  matcher.Likelies{Artist: "Unknown"},
			consensus: map[string]bool{"artist": true},
			items:     plain,
			want:      true,
		},
		{
			name:      "one item carries the comp flag",
			likelies:  matcher.Likelies{Artist: "Northline"},
			consensus: map[string]bool{"artist": true},
			items:     compFlagged,
			want:      true,
		},
		{
			name:      "no disjunct holds",
			likelies:  matcher.Likelies{Artist: "Northline"},
			consensus: map[string]bool{"artist": true},
			items:     plain,
			want:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := vaLikely(tt.likelies, tt.consensus, tt.items); got != tt.want {
				t.Errorf("vaLikely() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAlbumCriteriaOrdersTheCorpusReachableShapes covers the two filter
// shapes the committed corpus carries, release plus artist on 43 searches
// and release plus arid on 5, in the order get_album_criteria inserts
// them.
//
// Mutation: reversing the criteria order fails both rows here and both
// rows of TestAlbumSearchForRendersTheCommittedShapes.
func TestAlbumCriteriaOrdersTheCorpusReachableShapes(t *testing.T) {
	likelies := matcher.Likelies{Artist: "Northline", Album: "Static Coast"}
	tests := []struct {
		name string
		isVA bool
		want []criterion
	}{
		{
			name: "release then artist",
			want: []criterion{{"release", "Static Coast"}, {"artist", "Northline"}},
		},
		{
			name: "release then arid when the release is various-artists",
			isVA: true,
			want: []criterion{{"release", "Static Coast"}, {"arid", "89ad4ac3-39f7-470e-963a-56509c546377"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := albumCriteria(likelies, 1, tt.isVA, matcher.DefaultTunables())
			if !slices.Equal(got, tt.want) {
				t.Errorf("albumCriteria() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestAlbumSearchForBuildsTheExtraTagCriteria grades the extra-tag branch,
// rung 3's tenth declared corpus-coverage gap: no committed profile
// writes a musicbrainz block, so FIELDS_TO_MB_KEYS, the tracks and alias
// special cases, the catalognum space strip and the date:(0) trap are all
// unreachable by any corpus case and are graded by this hand-authored
// table alone.
//
// Two mutations: iterating FIELDS_TO_MB_KEYS in sorted key order instead
// of tun.ExtraTags fails the two-tag order row deterministically, because
// the sorted keys put label ahead of media; emitting one criterion per
// ExtraTags element without the de-duplication skip fails the repeated-tag
// row with two tracks terms.
func TestAlbumSearchForBuildsTheExtraTagCriteria(t *testing.T) {
	rows := loadTestdata[[]extraTagRow](t, "query/extra_tags.json")
	if len(rows) == 0 {
		t.Fatal("extra-tag fixture is empty")
	}
	for _, tt := range rows {
		t.Run(tt.Name, func(t *testing.T) {
			tun := matcher.DefaultTunables()
			tun.ExtraTags = tt.ExtraTags
			if got := AlbumSearchFor(tt.Items, tun).Query; got != tt.Want {
				t.Errorf("Query =\n  %q\nwant\n  %q", got, tt.Want)
			}
		})
	}
}

// TestAlbumSearchForRendersTheCommittedShapes renders the two filter
// shapes the corpus carries end to end. Each want is a literal: the
// release clauses are percent-decoded from the tree's own committed
// path_query values (release%3A%28symphony+no.+9%2C+%5C%22choral%5C%22%29
// and release%3A%28frank+zappa%5C%3A+cucamonga%29) and the artist clauses
// are this fixture's own.
//
// The first row also proves the escape set's absences: the comma and the
// period survive while the quotes do not.
func TestAlbumSearchForRendersTheCommittedShapes(t *testing.T) {
	tests := []struct {
		name  string
		items []matcher.ItemFields
		want  string
	}{
		{
			name: "release plus artist",
			items: []matcher.ItemFields{
				{Artist: "Vienna Philharmonic", Album: `Symphony No. 9, "Choral"`},
			},
			want: `release:(symphony no. 9, \"choral\") artist:(vienna philharmonic)`,
		},
		{
			name: "release plus arid, va_likely through the comp flag",
			items: []matcher.ItemFields{
				{Artist: "Frank Zappa", Album: "Frank Zappa: Cucamonga", Comp: true},
			},
			want: `release:(frank zappa\: cucamonga) arid:(89ad4ac3\-39f7\-470e\-963a\-56509c546377)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AlbumSearchFor(tt.items, matcher.DefaultTunables())
			if got.Query != tt.want {
				t.Errorf("Query =\n  %q\nwant\n  %q", got.Query, tt.want)
			}
			if got.Limit != 5 {
				t.Errorf("Limit = %d, want 5 (search_limit's shipped default)", got.Limit)
			}
		})
	}
}

// TestAlbumSearchForDropsEmptyTerms proves both filter points: an empty
// album name drops its own criterion rather than rendering release:()
// and rather than emptying the query.
func TestAlbumSearchForDropsEmptyTerms(t *testing.T) {
	got := AlbumSearchFor([]matcher.ItemFields{{Artist: "Northline"}}, matcher.DefaultTunables()).Query
	if want := "artist:(northline)"; got != want {
		t.Errorf("Query = %q, want %q", got, want)
	}
}
