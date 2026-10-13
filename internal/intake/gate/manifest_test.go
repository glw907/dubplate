package gate

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

// testPolicy is the small policy the tests run under: five files, three
// folders deep, a 1000-byte album.
func testPolicy() Policy {
	return Policy{
		MaxAlbumBytes:    1000,
		MaxFiles:         5,
		MaxPathDepth:     3,
		MaxTagBytes:      4096,
		MaxPictureBytes:  4096,
		MinSampleRate:    44100,
		MinBitsPerSample: 16,
		Allowlist:        []string{"flac"},
	}
}

func album(source string, files ...ManifestFile) Manifest {
	return Manifest{Unit: UnitAlbum, Container: ContainerFiles, Source: source, Files: files}
}

func mf(path string, size int64) ManifestFile { return ManifestFile{Path: path, Size: size} }

// submissionCodes lists the codes of a refusal's submission-level reasons.
func submissionCodes(r Refusal) []reject.Code {
	var codes []reject.Code
	for _, reason := range r.Submission {
		codes = append(codes, reason.Code)
	}
	return codes
}

// fileCodes lists "path=code" for each of a refusal's per-file reasons.
func fileCodes(r Refusal) []string {
	var out []string
	for _, fr := range r.Files {
		out = append(out, string(fr.Path)+"="+string(fr.Reason.Code))
	}
	return out
}

func TestCheckManifest_Passes(t *testing.T) {
	tests := []struct {
		name string
		m    Manifest
	}{
		{"an album with an extra", album("cd", mf("a/01.flac", 100), mf("a/02.FLAC", 100), mf("a/cover.jpg", 50))},
		{"a ZIP", Manifest{Unit: UnitAlbum, Container: ContainerZip, Source: "cd", Files: []ManifestFile{mf("album.zip", 500)}}},
		{"separate tracks", Manifest{Unit: UnitTrack, Container: ContainerFiles, Source: "cd", Files: []ManifestFile{mf("01.flac", 10), mf("02.flac", 10)}}},
		{"separate tracks in a ZIP", Manifest{Unit: UnitTrack, Container: ContainerZip, Source: "cd", Files: []ManifestFile{mf("tracks.zip", 10)}}},
		{"a zero-byte file", album("cd", mf("01.flac", 0))},
		{"exactly at every cap", album("cd", mf("a/b/01.flac", 995), mf("a/b/02.flac", 5), mf("c.jpg", 0), mf("d.jpg", 0), mf("e.jpg", 0))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckManifest(tt.m, testPolicy())
			if err != nil {
				t.Fatalf("CheckManifest error = %v, want none", err)
			}
			if len(got.Submission) != 0 || len(got.Files) != 0 {
				t.Errorf("refusal = %v %v, want none", submissionCodes(got), fileCodes(got))
			}
		})
	}
}

func TestCheckManifest_Sources(t *testing.T) {
	// The seven values the submissions.source column stores, written out
	// here rather than read from the package.
	valid := []string{"bought", "cd", "vinyl", "free-download", "own-recording", "gift", "other"}
	for _, src := range valid {
		t.Run("accepts "+src, func(t *testing.T) {
			got, err := CheckManifest(album(src, mf("01.flac", 1)), testPolicy())
			if err != nil || len(got.Submission) != 0 {
				t.Errorf("source %q: refusal %v, error %v, want none", src, submissionCodes(got), err)
			}
		})
	}
	for _, src := range []string{"", "nope", "CD", " cd"} {
		t.Run("refuses "+src, func(t *testing.T) {
			got, err := CheckManifest(album(src, mf("01.flac", 1)), testPolicy())
			if err != nil {
				t.Fatalf("error = %v, want none", err)
			}
			if want := []reject.Code{reject.MissingSource}; !slices.Equal(submissionCodes(got), want) {
				t.Errorf("submission reasons = %v, want %v", submissionCodes(got), want)
			}
		})
	}
}

// TestCheckManifest_PathsOfEachKind is AC 6's manifest half: a path of
// each refused kind answers unsafe-path.
func TestCheckManifest_PathsOfEachKind(t *testing.T) {
	for _, p := range []string{"/abs.flac", "a//b.flac", "../x.flac", "./x.flac", "a/./b.flac", `a\b.flac`, "a\x00b.flac", "", "a/b/", "a/../b.flac", ".."} {
		t.Run(p, func(t *testing.T) {
			got, err := CheckManifest(album("cd", mf(p, 1), mf("ok.flac", 1)), testPolicy())
			if err != nil {
				t.Fatalf("error = %v, want none", err)
			}
			if want := []string{p + "=unsafe-path"}; !slices.Equal(fileCodes(got), want) {
				t.Errorf("file reasons = %q, want %q", fileCodes(got), want)
			}
		})
	}
}

func TestCheckManifest_Limits(t *testing.T) {
	tests := []struct {
		name     string
		m        Manifest
		wantSub  []reject.Code
		wantFile []string
	}{
		{"depth over the cap", album("cd", mf("a/b/c/d.flac", 1), mf("ok.flac", 1)), nil, []string{"a/b/c/d.flac=path-too-deep"}},
		{"a count over the cap", album("cd", mf("1.flac", 1), mf("2.flac", 1), mf("3.flac", 1), mf("4.flac", 1), mf("5.flac", 1), mf("6.flac", 1)), []reject.Code{reject.TooManyFiles}, nil},
		{"a declared total over the cap", album("cd", mf("1.flac", 600), mf("2.flac", 401)), []reject.Code{reject.TooLarge}, nil},
		{"extras count toward the cap", album("cd", mf("1.flac", 600), mf("scan.jpg", 401)), []reject.Code{reject.TooLarge}, nil},
		{"a sum that would overflow", album("cd", mf("1.flac", 1<<62), mf("2.flac", 1<<62), mf("3.flac", 1<<62)), []reject.Code{reject.TooLarge}, nil},
		{"an audio extension outside the allowlist", album("cd", mf("1.mp3", 1), mf("2.flac", 1)), nil, []string{"1.mp3=bad-extension"}},
		{"an upper-case outside extension", album("cd", mf("1.M4A", 1), mf("2.flac", 1)), nil, []string{"1.M4A=bad-extension"}},
		{"an empty manifest", album("cd"), []reject.Code{reject.NoAudio}, nil},
		{"no files key", Manifest{Unit: UnitAlbum, Container: ContainerFiles, Source: "cd"}, []reject.Code{reject.NoAudio}, nil},
		{"only extras", album("cd", mf("cover.jpg", 1), mf("a.cue", 1)), []reject.Code{reject.NoAudio}, nil},
		{"a ZIP container with a non-zip file", Manifest{Unit: UnitAlbum, Container: ContainerZip, Source: "cd", Files: []ManifestFile{mf("a.flac", 1)}}, []reject.Code{reject.NoAudio}, nil},
		{"a ZIP container with two files", Manifest{Unit: UnitAlbum, Container: ContainerZip, Source: "cd", Files: []ManifestFile{mf("a.zip", 1), mf("b.zip", 1)}}, []reject.Code{reject.NoAudio}, nil},
		{"a ZIP container with none", Manifest{Unit: UnitAlbum, Container: ContainerZip, Source: "cd"}, []reject.Code{reject.NoAudio}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckManifest(tt.m, testPolicy())
			if err != nil {
				t.Fatalf("error = %v, want none", err)
			}
			if !slices.Equal(submissionCodes(got), tt.wantSub) {
				t.Errorf("submission reasons = %v, want %v", submissionCodes(got), tt.wantSub)
			}
			if !slices.Equal(fileCodes(got), tt.wantFile) {
				t.Errorf("file reasons = %q, want %q", fileCodes(got), tt.wantFile)
			}
		})
	}
}

func TestCheckManifest_BadExtensionNamesTheAllowlist(t *testing.T) {
	p := testPolicy()
	got, err := CheckManifest(album("cd", mf("1.mp3", 1), mf("2.flac", 1)), p)
	if err != nil || len(got.Files) != 1 {
		t.Fatalf("refusal = %v, error %v, want one file reason", fileCodes(got), err)
	}
	if want := reject.WithFormats([]string{"flac"}); got.Files[0].Reason != want {
		t.Errorf("reason = %+v, want %+v", got.Files[0].Reason, want)
	}
}

// TestCheckManifest_CollectsEveryReason is the spec's "many bad files": each
// failing path is listed with its own code, and the submission-level reasons
// come with them.
func TestCheckManifest_CollectsEveryReason(t *testing.T) {
	m := album("",
		mf("/abs.flac", 400), mf("a/b/c/d.flac", 400), mf("x.mp3", 400), mf("../up.flac", 400),
		mf("ok.flac", 400), mf("six.jpg", 400),
	)
	got, err := CheckManifest(m, testPolicy())
	if err != nil {
		t.Fatalf("error = %v, want none", err)
	}
	wantFiles := []string{"/abs.flac=unsafe-path", "a/b/c/d.flac=path-too-deep", "x.mp3=bad-extension", "../up.flac=unsafe-path"}
	if !slices.Equal(fileCodes(got), wantFiles) {
		t.Errorf("file reasons = %q, want %q", fileCodes(got), wantFiles)
	}
	wantSub := []reject.Code{reject.TooManyFiles, reject.TooLarge, reject.MissingSource}
	if !slices.Equal(submissionCodes(got), wantSub) {
		t.Errorf("submission reasons = %v, want %v", submissionCodes(got), wantSub)
	}
}

func TestCheckManifest_NonUTF8PathKeptAsBytes(t *testing.T) {
	got, err := CheckManifest(album("cd", mf("caf\xe9/01.mp3", 1), mf("ok.flac", 1)), testPolicy())
	if err != nil || len(got.Files) != 1 {
		t.Fatalf("refusal = %v, error %v, want one file reason", fileCodes(got), err)
	}
	if string(got.Files[0].Path) != "caf\xe9/01.mp3" {
		t.Errorf("path = %q, want the member's bytes", got.Files[0].Path)
	}
}

func TestCheckManifest_MalformedManifestErrors(t *testing.T) {
	tests := []struct {
		name string
		m    Manifest
		want error
	}{
		{"a negative size", album("cd", mf("1.flac", -1)), ErrMalformedManifest},
		{"a duplicate path", album("cd", mf("1.flac", 1), mf("1.flac", 2)), ErrMalformedManifest},
		{"separate tracks naming an extra", Manifest{Unit: UnitTrack, Container: ContainerFiles, Source: "cd", Files: []ManifestFile{mf("1.flac", 1), mf("cover.jpg", 1)}}, ErrExtraInTracks},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckManifest(tt.m, testPolicy())
			if !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
			for _, other := range []error{ErrMalformedManifest, ErrExtraInTracks} {
				if other != tt.want && errors.Is(err, other) {
					t.Errorf("error %v also matches %v; the two must stay distinct", err, other)
				}
			}
			if len(got.Submission) != 0 || len(got.Files) != 0 {
				t.Errorf("refusal = %v %v, want none beside the error", submissionCodes(got), fileCodes(got))
			}
		})
	}
}

func TestCheckManifest_ErrorNamesTheFault(t *testing.T) {
	_, err := CheckManifest(album("cd", mf("dup.flac", 1), mf("dup.flac", 1)), testPolicy())
	if err == nil || !strings.Contains(err.Error(), "dup.flac") {
		t.Errorf("error = %v, want it to name the duplicate path", err)
	}
}
