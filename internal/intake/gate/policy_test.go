package gate

import (
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

func TestLookupFormat(t *testing.T) {
	t.Run("a configured name", func(t *testing.T) {
		row, ok := LookupFormat("flac")
		if !ok {
			t.Fatal("LookupFormat(flac) found no row")
		}
		if row.Name != "flac" || row.Format != tagio.FormatFLAC {
			t.Errorf("row = %+v, want name flac and format FormatFLAC", row)
		}
		if !slices.Equal(row.Extensions, []string{".flac"}) {
			t.Errorf("extensions = %q, want [.flac]", row.Extensions)
		}
	})

	t.Run("a name no row carries", func(t *testing.T) {
		if row, ok := LookupFormat("mp3"); ok {
			t.Errorf("LookupFormat(mp3) = %+v, want none", row)
		}
	})
}

// TestFormatTableNamesSupportedFormats holds the table against
// tagio's registry: a row whose format tagio cannot read would pass
// the allowlist and fail at the first file.
func TestFormatTableNamesSupportedFormats(t *testing.T) {
	supported := tagio.SupportedFormats()
	seen := map[string]bool{}
	for _, row := range formats {
		if !slices.Contains(supported, row.Format) {
			t.Errorf("row %q names format %d, outside tagio.SupportedFormats() %v", row.Name, row.Format, supported)
		}
		if seen[row.Name] {
			t.Errorf("name %q appears twice", row.Name)
		}
		seen[row.Name] = true
		if len(row.Extensions) == 0 {
			t.Errorf("row %q carries no extension", row.Name)
		}
	}
}
