package matcher

import (
	"encoding/json"
	"os"
	"slices"
	"testing"
	"time"
)

// TestAddAlbumYearPreferredOriginalYear grades branch 1, the
// preferred.original_year path, off a hand-authored table
// (testdata/album_year_preferred_original.json) because pinned
// beets' corpus carries no case that sets preferred.original_year
// (spec, M8: one of the three corpus-unreachable surfaces). Every
// "want" is hand-computed from contract 9's own formula:
// AddRatio(|year - original|, |evalYear - original|), original =
// originalYear or 1889.
func TestAddAlbumYearPreferredOriginalYear(t *testing.T) {
	type fixture struct {
		Name         string  `json:"name"`
		Year         int     `json:"year"`
		OriginalYear int     `json:"originalYear"`
		EvalYear     int     `json:"evalYear"`
		Want         float64 `json:"want"`
	}

	raw, err := os.ReadFile("testdata/album_year_preferred_original.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixtures []fixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/album_year_preferred_original.json carries no fixtures")
	}

	tun := DefaultTunables()
	tun.Preferred.OriginalYear = true

	for _, f := range fixtures {
		t.Run(f.Name, func(t *testing.T) {
			dist := NewDistance(tun)
			info := &AlbumInfo{Year: f.Year, OriginalYear: f.OriginalYear}
			evalDate := time.Date(f.EvalYear, time.January, 1, 0, 0, 0, 0, time.UTC)
			addAlbumYear(dist, tun, Likelies{}, info, nil, nil, evalDate)

			got := dist.penalties["year"]
			if want := []float64{f.Want}; !slices.Equal(got, want) {
				t.Errorf("addAlbumYear(year=%d, originalYear=%d, evalYear=%d) added %v, want %v",
					f.Year, f.OriginalYear, f.EvalYear, got, want)
			}
		})
	}
}

func TestAddAlbumYear(t *testing.T) {
	tun := DefaultTunables()
	evalDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		likelyYear   int
		infoYear     int
		originalYear int
		want         []float64
	}{
		{"declines when the item set carries no likely year", 0, 1975, 0, nil},
		{"declines when the candidate carries no release year", 1980, 0, 0, nil},
		{"matching release years add zero", 1975, 1975, 0, []float64{0}},
		{"matching original years add zero even with a different release year", 1970, 1975, 1970, []float64{0}},
		{"a candidate with an original year ratios plainly", 1980, 1975, 1970, []float64{5.0 / 55.0}},
		{"no original year is a flat penalty", 1980, 1975, 0, []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Year: tt.likelyYear}
			info := &AlbumInfo{Year: tt.infoYear, OriginalYear: tt.originalYear}
			addAlbumYear(dist, tun, likelies, info, nil, nil, evalDate)

			got := dist.penalties["year"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumYear(likely=%d, year=%d, original=%d) added %v, want %v",
					tt.likelyYear, tt.infoYear, tt.originalYear, got, tt.want)
			}
		})
	}
}

// TestAddAlbumYearPreferredOriginalYearTakesPrecedence proves branch
// 1 is tried before branch 2: with Preferred.OriginalYear set, even a
// likely-year match that branch 2 would score zero for is ignored.
func TestAddAlbumYearPreferredOriginalYearTakesPrecedence(t *testing.T) {
	tun := DefaultTunables()
	tun.Preferred.OriginalYear = true
	evalDate := time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC)

	dist := NewDistance(tun)
	likelies := Likelies{Year: 1975}
	info := &AlbumInfo{Year: 1975, OriginalYear: 1970}
	addAlbumYear(dist, tun, likelies, info, nil, nil, evalDate)

	got := dist.penalties["year"]
	want := []float64{5.0 / 55.0}
	if !slices.Equal(got, want) {
		t.Errorf("addAlbumYear with Preferred.OriginalYear added %v, want %v (branch 1, not branch 2's zero)", got, want)
	}
}
