// Package gate enforces dubplate's format and quality policy over a
// staged upload: extension, symlink, decode-integrity, and CD-quality
// checks. It takes a plain [Policy] struct built by [internal/config]
// rather than importing it, so gate stays usable on its own.
package gate

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/reject"
)

// audioExt is the set of case-insensitive extensions Check recognizes as
// audio content. A file outside this set (album art, .cue sheets, .nfo)
// is out of scope and never inspected or reported. A file inside this
// set that Policy.AllowedExt does not name fails as
// [reject.ReasonBadExtension].
var audioExt = map[string]bool{
	".flac": true, ".mp3": true, ".m4a": true, ".ogg": true,
	".opus": true, ".wav": true, ".aiff": true,
}

// Policy configures one [Check] run. AllowedExt entries carry a leading
// dot ("`.flac`") and are matched case-insensitively.
type Policy struct {
	AllowedExt       []string
	RequireDecode    bool
	MinSampleRate    int
	MinBitsPerSample int
	// MetaflacPath is the metaflac binary Check runs to read a FLAC's
	// sample rate and bit depth; a test seam, since production always
	// sets it to "metaflac" resolved on PATH.
	MetaflacPath string
}

// Verdict is one staged file's outcome. OK is false unless Reason is
// [reject.ReasonNone].
type Verdict struct {
	Path   string
	OK     bool
	Reason reject.Reason
}

// Check walks every regular file and symlink under root and returns one
// Verdict per audio file found (an extension in [audioExt]); a
// non-audio file is skipped entirely and never appears in the result.
//
// A symlink fails as [reject.ReasonSymlink] regardless of its
// extension, checked before any extension logic runs: a symlink can
// point anywhere the beets process can read, or (from a maliciously
// crafted archive member) at a path that does not even exist yet. An
// audio file whose extension p does not allow fails as
// [reject.ReasonBadExtension] without further inspection. An allowed
// file is decoded with flacPath -t when p.RequireDecode is set (failing
// as [reject.ReasonCorrupt] on a bad decode), then always checked
// against p.MinSampleRate and p.MinBitsPerSample via p.MetaflacPath,
// failing as [reject.ReasonBelowCDQuality] on either a below-floor
// reading or a metaflac that produces none.
func Check(ctx context.Context, root string, p Policy, flacPath string) ([]Verdict, error) {
	var verdicts []Verdict

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Type()&fs.ModeSymlink != 0 {
			verdicts = append(verdicts, Verdict{Path: path, Reason: reject.ReasonSymlink})
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !audioExt[ext] {
			return nil
		}
		if !slices.ContainsFunc(p.AllowedExt, func(a string) bool { return strings.EqualFold(a, ext) }) {
			verdicts = append(verdicts, Verdict{Path: path, Reason: reject.ReasonBadExtension})
			return nil
		}

		if p.RequireDecode && !decodes(ctx, flacPath, path) {
			verdicts = append(verdicts, Verdict{Path: path, Reason: reject.ReasonCorrupt})
			return nil
		}

		rate, bps, err := sampleInfo(ctx, p.MetaflacPath, path)
		if err != nil || rate < p.MinSampleRate || bps < p.MinBitsPerSample {
			verdicts = append(verdicts, Verdict{Path: path, Reason: reject.ReasonBelowCDQuality})
			return nil
		}

		verdicts = append(verdicts, Verdict{Path: path, OK: true})
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}
	return verdicts, nil
}

// decodes reports whether flacPath -t's full decode test passes for
// path, catching a corrupt or truncated upload the way a header-only
// probe would miss.
func decodes(ctx context.Context, flacPath, path string) bool {
	return exec.CommandContext(ctx, flacPath, "-t", path).Run() == nil
}

// sampleInfo runs metaflacPath --show-sample-rate --show-bps against
// path and parses its two-line output.
func sampleInfo(ctx context.Context, metaflacPath, path string) (rate, bps int, err error) {
	out, err := exec.CommandContext(ctx, metaflacPath, "--show-sample-rate", "--show-bps", path).Output()
	if err != nil {
		return 0, 0, fmt.Errorf("metaflac %s: %w", path, err)
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("metaflac %s: unexpected output %q", path, out)
	}
	rate, err = strconv.Atoi(strings.TrimSpace(lines[0]))
	if err != nil {
		return 0, 0, fmt.Errorf("metaflac %s: parse sample rate: %w", path, err)
	}
	bps, err = strconv.Atoi(strings.TrimSpace(lines[1]))
	if err != nil {
		return 0, 0, fmt.Errorf("metaflac %s: parse bits per sample: %w", path, err)
	}
	return rate, bps, nil
}
