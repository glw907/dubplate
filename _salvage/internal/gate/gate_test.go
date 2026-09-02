package gate_test

import (
	"context"
	"errors"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// TestCheck_SymlinkAnyExtension ports the shell layer's `-L` check, which
// runs before any extension logic and so catches a symlink regardless of
// what it is named -- an audio extension and a non-audio one both fail the
// same way.
func TestCheck_SymlinkAnyExtension(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target-does-not-exist")
	for _, name := range []string{"link.flac", "link.txt"} {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatalf("symlink %s: %v", name, err)
		}
	}

	verdicts, err := gate.Check(context.Background(), root, gate.Policy{AllowedExt: []string{".flac"}}, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	got := map[string]reject.Reason{}
	for _, v := range verdicts {
		got[filepath.Base(v.Path)] = v.Reason
	}
	for _, name := range []string{"link.flac", "link.txt"} {
		if r, ok := got[name]; !ok || r != reject.ReasonSymlink {
			t.Errorf("%s reason = %v (present=%v), want ReasonSymlink", name, r, ok)
		}
	}
}

// TestCheck_WrongExtensionPerPolicy proves an audio extension outside
// Policy.AllowedExt fails as ReasonBadExtension, while a non-audio
// extension (album art, sleeve notes) never appears in the result at all.
func TestCheck_WrongExtensionPerPolicy(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "track.mp3"), []byte("not audio"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "cover.jpg"), []byte("art"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	verdicts, err := gate.Check(context.Background(), root, gate.Policy{AllowedExt: []string{".flac"}}, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 {
		t.Fatalf("verdicts = %+v, want exactly one (cover.jpg is out of scope)", verdicts)
	}
	if verdicts[0].Path != filepath.Join(root, "track.mp3") || verdicts[0].Reason != reject.ReasonBadExtension {
		t.Errorf("verdict = %+v, want track.mp3/ReasonBadExtension", verdicts[0])
	}
}

// TestCheck_AllowedExtCaseInsensitive proves Policy.AllowedExt matches
// case-insensitively, per its doc comment.
func TestCheck_AllowedExtCaseInsensitive(t *testing.T) {
	root := t.TempDir()
	makeFlac(t, filepath.Join(root, "track.flac"), 44100)

	policy := gate.Policy{AllowedExt: []string{".FLAC"}, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: "metaflac"}
	verdicts, err := gate.Check(context.Background(), root, policy, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason == reject.ReasonBadExtension {
		t.Fatalf("verdicts = %+v, want the uppercase .FLAC policy entry to match a lowercase .flac file", verdicts)
	}
}

// TestCheck_CorruptFlacFailsFullDecode ports the bats case "quality_gate
// fails a corrupt/truncated flac" against a truncated-FLAC fixture built
// the same way: a real FLAC encode, then truncated after the fact so the
// header parses but the full decode does not.
func TestCheck_CorruptFlacFailsFullDecode(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "track.flac")
	makeFlac(t, path, 44100)
	if err := os.Truncate(path, 2000); err != nil {
		t.Fatalf("truncate fixture: %v", err)
	}

	policy := gate.Policy{AllowedExt: []string{".flac"}, RequireDecode: true, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: "metaflac"}
	verdicts, err := gate.Check(context.Background(), root, policy, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonCorrupt {
		t.Fatalf("verdicts = %+v, want one ReasonCorrupt verdict", verdicts)
	}
}

// TestCheck_BelowSampleRateFloor ports the bats case "quality_gate fails a
// flac below 44.1kHz" (tests/lib.bats).
func TestCheck_BelowSampleRateFloor(t *testing.T) {
	root := t.TempDir()
	makeFlac(t, filepath.Join(root, "track.flac"), 22050)

	policy := gate.Policy{AllowedExt: []string{".flac"}, RequireDecode: true, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: "metaflac"}
	verdicts, err := gate.Check(context.Background(), root, policy, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonBelowCDQuality {
		t.Fatalf("verdicts = %+v, want one ReasonBelowCDQuality verdict", verdicts)
	}
}

// TestCheck_Below16BitFloor ports the bats case "quality_gate fails an
// 8-bit flac" (tests/lib.bats).
func TestCheck_Below16BitFloor(t *testing.T) {
	root := t.TempDir()
	make8BitFlac(t, filepath.Join(root, "track.flac"))

	policy := gate.Policy{AllowedExt: []string{".flac"}, RequireDecode: true, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: "metaflac"}
	verdicts, err := gate.Check(context.Background(), root, policy, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 || verdicts[0].Reason != reject.ReasonBelowCDQuality {
		t.Fatalf("verdicts = %+v, want one ReasonBelowCDQuality verdict", verdicts)
	}
}

// TestCheck_PassesValidFlacAndIgnoresNonAudioSiblings ports the bats case
// "quality_gate passes a valid 44.1kHz/16-bit flac and ignores non-audio
// siblings": the positive control proving a good file passes and art/notes
// never appear in the result.
func TestCheck_PassesValidFlacAndIgnoresNonAudioSiblings(t *testing.T) {
	root := t.TempDir()
	makeFlac(t, filepath.Join(root, "track.flac"), 44100)
	if err := os.WriteFile(filepath.Join(root, "cover.jpg"), []byte("art"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "notes.nfo"), []byte("notes"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	policy := gate.Policy{AllowedExt: []string{".flac"}, RequireDecode: true, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: "metaflac"}
	verdicts, err := gate.Check(context.Background(), root, policy, "flac")
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	if len(verdicts) != 1 || !verdicts[0].OK || verdicts[0].Reason != reject.ReasonNone {
		t.Fatalf("verdicts = %+v, want exactly one passing verdict and no entries for the siblings", verdicts)
	}
}

// TestCheck_MetaflacPathIsATestSeam exercises the CD-quality floor against
// a stub metaflac rather than a real encode, proving Policy.MetaflacPath is
// substitutable and that Check's error handling (a nonzero exit,
// unparsable output, or a short line count) all fall back to
// ReasonBelowCDQuality rather than a bare error.
func TestCheck_MetaflacPathIsATestSeam(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "track.flac"), []byte("stand-in content, never decoded since RequireDecode is false"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	tests := []struct {
		name   string
		script string
		want   reject.Reason
	}{
		{"above both floors passes", "printf '48000\\n24\\n'", reject.ReasonNone},
		{"below the sample-rate floor fails", "printf '22050\\n16\\n'", reject.ReasonBelowCDQuality},
		{"below the bit-depth floor fails", "printf '44100\\n8\\n'", reject.ReasonBelowCDQuality},
		{"a nonzero exit fails", "exit 1", reject.ReasonBelowCDQuality},
		{"unparsable output fails", "printf 'not-a-number\\n16\\n'", reject.ReasonBelowCDQuality},
		{"one-line output fails", "printf '44100\\n'", reject.ReasonBelowCDQuality},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stub := writeStubMetaflac(t, tt.script)
			policy := gate.Policy{AllowedExt: []string{".flac"}, MinSampleRate: 44100, MinBitsPerSample: 16, MetaflacPath: stub}
			verdicts, err := gate.Check(context.Background(), root, policy, "flac")
			if err != nil {
				t.Fatalf("Check: %v", err)
			}
			if len(verdicts) != 1 {
				t.Fatalf("verdicts = %+v, want exactly one", verdicts)
			}
			if verdicts[0].Reason != tt.want {
				t.Errorf("Reason = %v, want %v", verdicts[0].Reason, tt.want)
			}
			if wantOK := tt.want == reject.ReasonNone; verdicts[0].OK != wantOK {
				t.Errorf("OK = %v, want %v", verdicts[0].OK, wantOK)
			}
		})
	}
}

func TestCheck_ContextCanceled(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "track.flac"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := gate.Check(ctx, root, gate.Policy{AllowedExt: []string{".flac"}}, "flac")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Check err = %v, want wrapped context.Canceled", err)
	}
}

// TestImportsRejectOnly enforces the T6 plan criterion "gate imports
// reject only" at test time: a later change that pulls in another package
// from this module (internal/config, say, breaking gate's standalone
// usability) fails this test rather than surviving unnoticed.
func TestImportsRejectOnly(t *testing.T) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "gate.go", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse gate.go: %v", err)
	}
	for _, imp := range node.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if strings.HasPrefix(path, "github.com/glw907/dubplate/") && path != "github.com/glw907/dubplate/internal/reject" {
			t.Errorf("gate.go imports %s; gate must import internal/reject only among this module's packages", path)
		}
	}
}

// --- audio fixtures ----------------------------------------------------

// makeFlac synthesizes a real, valid FLAC at rate (16-bit), matching
// tests/lib.bats' make_flac helper: ffmpeg and flac are both present on
// this workstation, and the source is fully deterministic.
func makeFlac(t *testing.T, dest string, rate int) {
	t.Helper()
	wav := filepath.Join(t.TempDir(), "gen.wav")
	if out, err := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-ar", strconv.Itoa(rate), "-sample_fmt", "s16", wav).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg: %v: %s", err, out)
	}
	if out, err := exec.Command("flac", "--best", "-f", "-o", dest, wav).CombinedOutput(); err != nil {
		t.Fatalf("flac encode: %v: %s", err, out)
	}
}

// make8BitFlac synthesizes a real, valid 44.1kHz/8-bit FLAC, matching
// tests/lib.bats' inline 8-bit fixture.
func make8BitFlac(t *testing.T, dest string) {
	t.Helper()
	wav := filepath.Join(t.TempDir(), "gen8.wav")
	if out, err := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error", "-y",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=1",
		"-ar", "44100", "-acodec", "pcm_u8", wav).CombinedOutput(); err != nil {
		t.Fatalf("ffmpeg 8-bit: %v: %s", err, out)
	}
	if out, err := exec.Command("flac", "--best", "-f", "-o", dest, wav).CombinedOutput(); err != nil {
		t.Fatalf("flac encode 8-bit: %v: %s", err, out)
	}
}

// writeStubMetaflac writes an executable shell script standing in for
// metaflac, whose body is script.
func writeStubMetaflac(t *testing.T, script string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "metaflac-stub")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script+"\n"), 0o755); err != nil {
		t.Fatalf("write stub metaflac: %v", err)
	}
	return path
}
