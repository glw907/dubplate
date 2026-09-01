package reject

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// allReasons lists every non-zero Reason this package defines, kept
// separate from the production code so a reason added to the const block
// without a matching entry here fails TestFriendlyNeverEmpty and
// TestNextStepNeverEmpty instead of silently shipping blank copy.
// TestAllReasonsCoversEnum pins the list's end against the const block
// itself, so a forgotten entry here fails loudly rather than by omission.
var allReasons = []Reason{
	ReasonBadExtension,
	ReasonSymlink,
	ReasonCorrupt,
	ReasonBelowCDQuality,
	ReasonTooLarge,
	ReasonUnpackFailed,
	ReasonUnsafePath,
	ReasonNoSpace,
	ReasonTooManySessions,
}

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
		{"unpack failed", ReasonUnpackFailed, "This archive wouldn't unpack as a ZIP."},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.reason.Friendly(); got != tt.want {
				t.Errorf("Friendly() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestReasonLabel pins each reason's chip string against
// passb-copy-final.md section 4, the wording authority for U10's
// per-track chips.
func TestReasonLabel(t *testing.T) {
	tests := []struct {
		reason Reason
		want   string
	}{
		{ReasonBadExtension, "not FLAC"},
		{ReasonSymlink, "shortcut file"},
		{ReasonCorrupt, "damaged in transit"},
		{ReasonBelowCDQuality, "below CD quality"},
		{ReasonTooLarge, "too big"},
		{ReasonUnpackFailed, "wouldn't unpack"},
		{ReasonUnsafePath, "unsafe file path"},
		{ReasonNoSpace, "library is full"},
		{ReasonTooManySessions, "too many uploads"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.reason.Label(); got != tt.want {
				t.Errorf("Label() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestLabelNeverEmpty mirrors TestFriendlyNeverEmpty for Label: every
// reason a real rejection can carry has a non-empty chip string, and
// ReasonNone (a passing verdict) does not.
func TestLabelNeverEmpty(t *testing.T) {
	for _, r := range allReasons {
		if got := r.Label(); got == "" {
			t.Errorf("Label() for reason %d = %q, want non-empty", r, got)
		}
	}
	if got := ReasonNone.Label(); got != "" {
		t.Errorf("Label() for ReasonNone = %q, want empty string, not a placeholder chip", got)
	}
}

// TestFriendlyNeverEmpty confirms every reason that can appear in a real
// rejection renders non-empty wording; a blank line would leak into a
// contributor's email with no explanation attached. ReasonNone is excluded
// by design, since it's the zero value a passing verdict carries -- Friendly
// renders it as an empty string on purpose rather than a placeholder line.
func TestFriendlyNeverEmpty(t *testing.T) {
	for _, r := range allReasons {
		if got := r.Friendly(); got == "" {
			t.Errorf("Friendly() for reason %d = %q, want non-empty", r, got)
		}
	}
	if got := ReasonNone.Friendly(); got != "" {
		t.Errorf("Friendly() for ReasonNone = %q, want empty string, not a placeholder line", got)
	}
}

// TestNextStepNeverEmpty mirrors TestFriendlyNeverEmpty for NextStep: every
// reason a real rejection can carry must tell the contributor what to do
// next, and ReasonNone (a passing verdict) must not.
func TestNextStepNeverEmpty(t *testing.T) {
	for _, r := range allReasons {
		if got := r.NextStep(); got == "" {
			t.Errorf("NextStep() for reason %d = %q, want non-empty", r, got)
		}
	}
	if got := ReasonNone.NextStep(); got != "" {
		t.Errorf("NextStep() for ReasonNone = %q, want empty string, not a placeholder line", got)
	}
}

// TestAllReasonsCoversEnum pins allReasons' length against the end of the
// Reason const block: the value one past the last entry in allReasons must
// fall through every switch to its default case (empty Friendly, empty
// NextStep, "none" code), the same fallthrough an unregistered Reason
// would hit. If a Reason is ever added to the const block without a
// matching entry in allReasons, that new Reason lands exactly at this
// boundary and this test fails, naming the fix.
func TestAllReasonsCoversEnum(t *testing.T) {
	next := Reason(len(allReasons) + 1)
	if got := next.Friendly(); got != "" {
		t.Errorf("Reason %d (one past allReasons) has non-empty Friendly() = %q; add the new Reason to allReasons in reject_test.go", next, got)
	}
	if got := next.NextStep(); got != "" {
		t.Errorf("Reason %d (one past allReasons) has non-empty NextStep() = %q; add the new Reason to allReasons in reject_test.go", next, got)
	}
	if got := next.Label(); got != "" {
		t.Errorf("Reason %d (one past allReasons) has non-empty Label() = %q; add the new Reason to allReasons in reject_test.go", next, got)
	}
	if got := next.code(); got != "none" {
		t.Errorf("Reason %d (one past allReasons) has code() = %q, want %q; add the new Reason to allReasons in reject_test.go", next, got, "none")
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
			"track2.mp3: This is an MP3 file, and the library keeps lossless FLAC.",
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
		wantLines := "track1.flac: This FLAC is below CD quality. " + ReasonBelowCDQuality.NextStep() + "\n" +
			"track2.mp3: This is an MP3 file, and the library keeps lossless FLAC. " + ReasonBadExtension.NextStep()
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
		for _, want := range []string{"Hi Alice,", "Bad Album", "track2.mp3: This is an MP3 file"} {
			if !strings.Contains(got, want) {
				t.Errorf("output missing %q; got %q", want, got)
			}
		}
	})

	// Ports the retired shell layer's tests/music-import.bats, which pinned
	// friendly_reason()'s per-extension wording literally: the shell names
	// the file's actual extension rather than a generic "not a FLAC file"
	// phrasing.
	t.Run("bad-extension line names the actual extension", func(t *testing.T) {
		path := writeTemplate(t)
		got, err := RenderEmail(path, "alice", "Bad Album", []Rejection{
			{Path: "alice/Bad Album/track2.mp3", Reason: ReasonBadExtension},
		})
		if err != nil {
			t.Fatalf("RenderEmail: %v", err)
		}
		// The reason wording itself.
		if !strings.Contains(got, "This is an MP3 file, and the library keeps lossless FLAC.") {
			t.Errorf("output missing the per-extension reason line; got %q", got)
		}
		// The reason line prefixed with the failing file's basename.
		if !strings.Contains(got, "track2.mp3: This is an MP3 file, and the library keeps lossless FLAC.") {
			t.Errorf("output missing the per-file reason line; got %q", got)
		}
	})

	t.Run("bad-extension line maps every known extension per the shell's name table", func(t *testing.T) {
		tests := []struct {
			path string
			want string
		}{
			{"track.mp3", "This is an MP3 file, and the library keeps lossless FLAC."},
			{"track.m4a", "This is an AAC file, and the library keeps lossless FLAC."},
			{"track.ogg", "This is an Ogg Vorbis file, and the library keeps lossless FLAC."},
			{"track.opus", "This is an Opus file, and the library keeps lossless FLAC."},
			{"track.wav", "This is an WAV file, and the library keeps lossless FLAC."},
			{"track.aiff", "This is an AIFF file, and the library keeps lossless FLAC."},
			{"track.wma", "This is an WMA file, and the library keeps lossless FLAC."},
		}
		for _, tt := range tests {
			path := writeTemplate(t)
			got, err := RenderEmail(path, "Alice", "Bad Album", []Rejection{{Path: tt.path, Reason: ReasonBadExtension}})
			if err != nil {
				t.Fatalf("RenderEmail(%s): %v", tt.path, err)
			}
			if !strings.Contains(got, tt.want) {
				t.Errorf("RenderEmail(%s): missing %q; got %q", tt.path, tt.want, got)
			}
		}
	})
}

// TestReasonMarshalJSON pins the wire shape [Reason.MarshalJSON] emits
// against the literal example the plan names: a bare-integer encode is
// the failure this golden guards against, since a plain `int` marshal
// would silently pass a byte-count check but produce "5" instead of an
// object a client can render.
func TestReasonMarshalJSON(t *testing.T) {
	got, err := json.Marshal(ReasonBelowCDQuality)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	want := `{"code":"below-cd-quality","friendly":"This FLAC is below CD quality.","next_step":"Get a CD-quality or better FLAC and upload that instead.","label":"below CD quality"}`
	if string(got) != want {
		t.Errorf("Marshal(ReasonBelowCDQuality) = %s, want %s", got, want)
	}
}

// TestReasonJSONRoundTrip proves every reason a real rejection can carry,
// plus the zero value, survives a Marshal/Unmarshal round trip unchanged.
func TestReasonJSONRoundTrip(t *testing.T) {
	for _, r := range append([]Reason{ReasonNone}, allReasons...) {
		data, err := json.Marshal(r)
		if err != nil {
			t.Fatalf("Marshal(%d): %v", r, err)
		}
		var got Reason
		if err := json.Unmarshal(data, &got); err != nil {
			t.Fatalf("Unmarshal(%s): %v", data, err)
		}
		if got != r {
			t.Errorf("round trip: Marshal(%d) -> Unmarshal -> %d, want %d", r, got, r)
		}
	}
}

// TestReasonUnmarshalJSONUnknownCode proves a code this package never
// wrote (a future addition a decoder hasn't been rebuilt for) fails
// loudly instead of silently decoding to ReasonNone.
func TestReasonUnmarshalJSONUnknownCode(t *testing.T) {
	var r Reason
	err := json.Unmarshal([]byte(`{"code":"not-a-real-reason","friendly":"x","next_step":"y"}`), &r)
	if err == nil {
		t.Fatal("Unmarshal: want error for an unknown code, got nil")
	}
}

// TestReasonUnmarshalJSONIgnoresWireFriendlyAndNextStep proves Friendly,
// NextStep, and Label are recomputed from the decoded code rather than
// trusted off the wire: a decoder fed forged copy for a real code still
// lands on this package's own wording.
func TestReasonUnmarshalJSONIgnoresWireFriendlyAndNextStep(t *testing.T) {
	var r Reason
	forged := `{"code":"below-cd-quality","friendly":"forged text","next_step":"forged next step","label":"forged label"}`
	if err := json.Unmarshal([]byte(forged), &r); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if r != ReasonBelowCDQuality {
		t.Fatalf("Unmarshal: got %d, want ReasonBelowCDQuality", r)
	}
	if r.Friendly() != ReasonBelowCDQuality.Friendly() {
		t.Errorf("Friendly() = %q, want the package's own wording, not the wire's", r.Friendly())
	}
	if r.Label() != ReasonBelowCDQuality.Label() {
		t.Errorf("Label() = %q, want the package's own wording, not the wire's", r.Label())
	}
}

// TestEmailAndJSONCarryIdenticalStrings is the cross-asset test the plan
// requires: it renders a rejection email and independently marshals the
// same Reason to the API's JSON shape, then proves the email's per-file
// line carries exactly the Friendly and NextStep strings the JSON
// decodes to. A hardcoded divergence between RenderEmail and
// Reason.MarshalJSON -- one asset changed without the other -- fails
// this test.
func TestEmailAndJSONCarryIdenticalStrings(t *testing.T) {
	path := writeTemplate(t)
	rj := Rejection{Path: "alice/Bad Album/track1.flac", Reason: ReasonBelowCDQuality}

	email, err := RenderEmail(path, "Alice", "Bad Album", []Rejection{rj})
	if err != nil {
		t.Fatalf("RenderEmail: %v", err)
	}

	data, err := json.Marshal(rj.Reason)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var wire struct {
		Friendly string `json:"friendly"`
		NextStep string `json:"next_step"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	wantLine := "track1.flac: " + wire.Friendly + " " + wire.NextStep
	if !strings.Contains(email, wantLine) {
		t.Errorf("email does not carry the same Friendly/NextStep strings as the API JSON;\nemail = %q\nwant substring = %q", email, wantLine)
	}
}
