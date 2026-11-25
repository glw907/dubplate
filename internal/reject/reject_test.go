package reject

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"
)

// adopted holds the nine reasons PCF section 4 and the salvage package
// share, transcribed as literals so the contract cannot drift silently.
var adopted = []struct {
	code     Code
	label    string
	friendly string
	nextStep string
}{
	{"bad_extension", "not FLAC", "This isn't a FLAC file, and the library keeps lossless FLAC.", "Download this album as FLAC and upload that instead."},
	{"symlink", "shortcut file", "This is a shortcut, not the music file itself.", "Replace it with the real file and upload it again."},
	{"corrupt", "damaged in transit", "This file didn't survive the upload intact. Uploading it again usually fixes this.", "Upload it again from the original file."},
	{"below_cd_quality", "below CD quality", "This FLAC is below CD quality.", "Get a CD-quality or better FLAC and upload that instead."},
	{"too_large", "too big", "This upload is larger than the library allows.", "Ask the admin to add this one by hand."},
	{"unpack_failed", "wouldn't unpack", "This archive wouldn't unpack as a ZIP.", "Repack it as a standard ZIP file and upload it again."},
	{"unsafe_path", "unsafe file path", "This archive contains a file path the library can't safely use.", "Repack the album as a plain ZIP with real files only, then upload it again."},
	{"no_space", "library is full", "The library doesn't have enough free space to accept this upload right now.", "Try again in a little while, or let the admin know if this keeps happening."},
	{"too_many_sessions", "too many uploads", "You already have the maximum number of uploads in progress.", "Wait for one of your other uploads to finish, then try again."},
}

func TestAdoptedReasonsEqualPCF(t *testing.T) {
	for _, tt := range adopted {
		t.Run(string(tt.code), func(t *testing.T) {
			r, ok := Lookup(tt.code)
			if !ok {
				t.Fatalf("Lookup(%q) found nothing", tt.code)
			}
			if r.Label != tt.label {
				t.Errorf("Label = %q, want %q", r.Label, tt.label)
			}
			if r.Friendly != tt.friendly {
				t.Errorf("Friendly = %q, want %q", r.Friendly, tt.friendly)
			}
			if r.NextStep != tt.nextStep {
				t.Errorf("NextStep = %q, want %q", r.NextStep, tt.nextStep)
			}
			if !r.Blocking {
				t.Error("Blocking = false, want true")
			}
		})
	}
}

var (
	codeShape = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	// leaks is W8.2's catalogue less the words a reject reason may carry:
	// FLAC (conductor ruling 6) and "tags" in id3_tags alone.
	leaks    = regexp.MustCompile(`(?i)musicbrainz|beets|transcode|matcher|session|roster|epoch|collection|\btags?\b`)
	badShape = regexp.MustCompile(`(?i)couldn't be \w+`)
	warnings = map[Code]bool{"no_cover_art": true, "track_total_differs": true, "extras_dropped": true}
)

func TestEveryReasonCarriesItsFields(t *testing.T) {
	seen := make(map[Code]bool)
	for _, r := range All() {
		t.Run(string(r.Code), func(t *testing.T) {
			if !codeShape.MatchString(string(r.Code)) {
				t.Errorf("code %q is not lowercase snake_case", r.Code)
			}
			if seen[r.Code] {
				t.Errorf("code %q appears twice", r.Code)
			}
			seen[r.Code] = true
			if got := len(strings.Fields(r.Label)); got < 1 || got > 3 {
				t.Errorf("label %q has %d words, want one to three", r.Label, got)
			}
			if r.Friendly == "" {
				t.Error("no friendly sentence")
			}
			if r.NextStep == "" {
				t.Error("no next step")
			}
			if r.Blocking == warnings[r.Code] {
				t.Errorf("Blocking = %v, want %v", r.Blocking, !warnings[r.Code])
			}
			for _, s := range []string{r.Label, r.Friendly, r.NextStep, r.Canonical} {
				if strings.Contains(s, "—") {
					t.Errorf("%q carries an em dash", s)
				}
				if badShape.MatchString(s) {
					t.Errorf("%q has the \"couldn't be\" shape", s)
				}
				if r.Code != "id3_tags" && leaks.MatchString(s) {
					t.Errorf("%q leaks developer vocabulary", s)
				}
			}
		})
	}
}

func TestSetHoldsTwentySixBlockingAndThreeWarnings(t *testing.T) {
	var blocking, warned int
	for _, r := range All() {
		if r.Blocking {
			blocking++
		} else {
			warned++
		}
	}
	if blocking != 26 || warned != 3 {
		t.Errorf("set holds %d blocking and %d non-blocking reasons, want 26 and 3", blocking, warned)
	}
}

func TestNewReasonsNamedBySpec(t *testing.T) {
	for _, code := range []Code{
		"missing_source", "uploads_off", "over_quota", "too_many_files", "path_too_deep",
		"no_audio", "tags_too_large", "picture_too_large", "unsupported_flac", "source_damaged",
		"empty_track", "missing_track_number", "duplicate_track_position", "incomplete_album",
		"already_in_library", "could_not_check", "id3_tags",
	} {
		if r, ok := Lookup(code); !ok || !r.Blocking {
			t.Errorf("Lookup(%q) = %v, %v; want a blocking reason", code, r, ok)
		}
	}
}

func TestID3TagsNamesTheStripCommand(t *testing.T) {
	r := Of("id3_tags")
	if want := "mid3v2 --delete-all"; !strings.Contains(r.NextStep, want) {
		t.Errorf("NextStep = %q, want it to name %q", r.NextStep, want)
	}
	if want := "ID3 tags"; !strings.Contains(r.Friendly, want) {
		t.Errorf("Friendly = %q, want it to tell the member about %q", r.Friendly, want)
	}
}

func TestParameterizedReasons(t *testing.T) {
	for _, r := range All() {
		if (r.Param == "") != (r.Canonical == "") {
			t.Errorf("%s: Param %q and Canonical %q must be set together", r.Code, r.Param, r.Canonical)
		}
		if r.Param != "" && r.Code != "bad_extension" {
			t.Errorf("%s is parameterized; only bad_extension is", r.Code)
		}
	}
	r := Of("bad_extension")
	if r.Param != "formats" {
		t.Errorf("Param = %q, want %q", r.Param, "formats")
	}
	want := "This isn't one of the formats the library keeps (the configured formats)."
	if r.Canonical != want {
		t.Errorf("Canonical = %q, want %q", r.Canonical, want)
	}
}

func TestWithFormats(t *testing.T) {
	tests := []struct {
		name     string
		formats  []string
		label    string
		friendly string
		nextStep string
	}{
		{
			name:     "flac only renders the adopted strings",
			formats:  []string{"flac"},
			label:    "not FLAC",
			friendly: "This isn't a FLAC file, and the library keeps lossless FLAC.",
			nextStep: "Download this album as FLAC and upload that instead.",
		},
		{
			name:     "two formats",
			formats:  []string{"flac", "alac"},
			label:    "wrong format",
			friendly: "This isn't one of the formats the library keeps (FLAC or ALAC).",
			nextStep: "Get a copy in one of these formats and upload that instead.",
		},
		{
			name:     "three formats",
			formats:  []string{"flac", "alac", "wav"},
			label:    "wrong format",
			friendly: "This isn't one of the formats the library keeps (FLAC, ALAC, or WAV).",
			nextStep: "Get a copy in one of these formats and upload that instead.",
		},
		{
			name:     "one other format",
			formats:  []string{"alac"},
			label:    "wrong format",
			friendly: "This isn't one of the formats the library keeps (ALAC).",
			nextStep: "Get a copy in one of these formats and upload that instead.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := WithFormats(tt.formats)
			if r.Code != "bad_extension" || !r.Blocking {
				t.Errorf("Code = %q, Blocking = %v; want bad_extension, true", r.Code, r.Blocking)
			}
			if r.Label != tt.label || r.Friendly != tt.friendly || r.NextStep != tt.nextStep {
				t.Errorf("got (%q, %q, %q), want (%q, %q, %q)", r.Label, r.Friendly, r.NextStep, tt.label, tt.friendly, tt.nextStep)
			}
		})
	}
}

func TestMarshalJSON(t *testing.T) {
	got, err := json.Marshal(Of("too_large"))
	if err != nil {
		t.Fatal(err)
	}
	want := `{"code":"too_large","label":"too big","friendly":"This upload is larger than the library allows.","next_step":"Ask the admin to add this one by hand."}`
	if string(got) != want {
		t.Errorf("Marshal = %s, want %s", got, want)
	}
}

func TestLookup(t *testing.T) {
	if r, ok := Lookup("no-such-code"); ok {
		t.Errorf("Lookup(no-such-code) = %v, true; want false", r)
	}
	if r, ok := Lookup("corrupt"); !ok || r.Code != "corrupt" {
		t.Errorf("Lookup(corrupt) = %v, %v", r, ok)
	}
}

func TestOfPanicsOnAnUnknownCode(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("Of(no-such-code) did not panic")
		}
	}()
	Of("no-such-code")
}

func TestAllReturnsACopy(t *testing.T) {
	first := All()
	first[0].Friendly = "changed"
	if got := All()[0].Friendly; got == "changed" {
		t.Error("mutating All's result changed the set")
	}
}
