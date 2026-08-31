package reject

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReasonFriendly(t *testing.T) {
	tests := []struct {
		name   string
		reason Reason
		want   string
	}{
		{"none", ReasonNone, ""},
		{"bad extension", ReasonBadExtension, "This isn't a FLAC file, and the library keeps lossless FLAC."},
		{"symlink", ReasonSymlink, "This looks like a symlink or shortcut rather than a real file, and the library only accepts real files."},
		{"corrupt", ReasonCorrupt, "This file didn't survive the upload intact. Uploading it again usually fixes this."},
		{"below CD quality", ReasonBelowCDQuality, "This FLAC is below CD quality."},
		{"too large", ReasonTooLarge, "This upload is larger than the library allows."},
		{"unpack failed", ReasonUnpackFailed, "This archive couldn't be unpacked."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.Friendly(); got != tt.want {
				t.Errorf("Friendly() = %q, want %q", got, tt.want)
			}
		})
	}
}

// writeTemplate writes a template file with the three real placeholders,
// mirroring config/reject-email.txt's shape without depending on its exact
// prose.
func writeTemplate(t *testing.T) string {
	t.Helper()
	const tmpl = "Hi {first name},\n\n{album or file name}\n{one reason line per file, from the mapping below}\n\nBye.\n"
	path := filepath.Join(t.TempDir(), "reject-email.txt")
	if err := os.WriteFile(path, []byte(tmpl), 0o644); err != nil {
		t.Fatalf("write template: %v", err)
	}
	return path
}

func TestRenderEmail(t *testing.T) {
	t.Run("fills all three placeholders", func(t *testing.T) {
		path := writeTemplate(t)
		got, err := RenderEmail(path, "Alice", "Bad Album", []Rejection{
			{Path: "alice/Bad Album/track2.mp3", Reason: ReasonBadExtension},
		})
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		for _, want := range []string{
			"Hi Alice,",
			"Bad Album",
			"track2.mp3: This isn't a FLAC file, and the library keeps lossless FLAC.",
		} {
			if !strings.Contains(got, want) {
				t.Errorf("output missing %q; got %q", want, got)
			}
		}
		if strings.Contains(got, "{first name}") || strings.Contains(got, "{album or file name}") ||
			strings.Contains(got, "{one reason line per file, from the mapping below}") {
			t.Errorf("a placeholder survived unfilled: %q", got)
		}
	})

	t.Run("multiple rejections render one line each in order", func(t *testing.T) {
		path := writeTemplate(t)
		got, err := RenderEmail(path, "Alice", "Bad Album", []Rejection{
			{Path: "alice/Bad Album/track1.flac", Reason: ReasonBelowCDQuality},
			{Path: "alice/Bad Album/track2.mp3", Reason: ReasonBadExtension},
		})
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		wantLines := "track1.flac: This FLAC is below CD quality.\n" +
			"track2.mp3: This isn't a FLAC file, and the library keeps lossless FLAC."
		if !strings.Contains(got, wantLines) {
			t.Errorf("reason lines out of order or malformed; got %q, want substring %q", got, wantLines)
		}
	})

	// Bash's ${var//pat/repl} treats an unquoted "&" in the replacement as
	// "insert the matched text" (the shell layer's documented quoting
	// hazard, confirmed on bash 5.3). Go's strings.ReplaceAll has no such
	// special case, but the behavior stays pinned here so a future rewrite
	// of RenderEmail can't reintroduce it.
	t.Run("an ampersand in the album name renders literally", func(t *testing.T) {
		path := writeTemplate(t)
		got, err := RenderEmail(path, "Alice", "Sturm & Drang", nil)
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		if !strings.Contains(got, "Sturm & Drang") {
			t.Errorf("album name not rendered literally; got %q", got)
		}
		if strings.Contains(got, "album or file name") {
			t.Errorf("placeholder text leaked into output instead of the album name: %q", got)
		}
	})

	t.Run("no rejections renders an empty reason block", func(t *testing.T) {
		path := writeTemplate(t)
		got, err := RenderEmail(path, "Alice", "Bad Album", nil)
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		if strings.Contains(got, "{one reason line per file") {
			t.Errorf("reason placeholder left unfilled: %q", got)
		}
	})

	t.Run("missing template returns a wrapped not-exist error", func(t *testing.T) {
		_, err := RenderEmail(filepath.Join(t.TempDir(), "does-not-exist.txt"), "Alice", "Bad Album", nil)
		if err == nil {
			t.Fatal("RenderEmail: want error, got nil")
		}
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("RenderEmail: err = %v, want wrapped os.ErrNotExist", err)
		}
	})

	t.Run("renders the real deployed template", func(t *testing.T) {
		got, err := RenderEmail("../../config/reject-email.txt", "Alice", "Bad Album", []Rejection{
			{Path: "track2.mp3", Reason: ReasonBadExtension},
		})
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		for _, want := range []string{"Hi Alice,", "Bad Album", "track2.mp3: This isn't a FLAC file"} {
			if !strings.Contains(got, want) {
				t.Errorf("output missing %q; got %q", want, got)
			}
		}
	})
}
