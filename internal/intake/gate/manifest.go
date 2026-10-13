package gate

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/reject"
)

// Manifest is what a member declares when creating a submission, before a
// byte is sent.
type Manifest struct {
	Unit      Unit
	Container Container
	// Source is the member's answer to where the music came from.
	Source string
	Files  []ManifestFile
}

// ManifestFile is one declared file: the member's relative path, which need
// not be valid UTF-8, and its size in bytes.
type ManifestFile struct {
	Path string
	Size int64
}

// Refusal is every reason a manifest fails: the submission-level reasons,
// and one reason per failing file. A manifest with neither passes.
type Refusal struct {
	Submission []reject.Reason
	Files      []FileReason
}

// The two faults a client makes, which the seam answers as a malformed
// request rather than as a reject reason.
var (
	// ErrMalformedManifest is a negative size or a duplicate path.
	ErrMalformedManifest = errors.New("malformed manifest")
	// ErrExtraInTracks is a separate-tracks manifest that names a file
	// without an audio extension: the page leaves extras out, so a client
	// that sends one is at fault.
	ErrExtraInTracks = errors.New("separate tracks keep audio files only")
)

// sources are the answers to where the music came from, as the
// submissions.source column stores them.
var sources = []string{"bought", "cd", "vinyl", "free-download", "own-recording", "gift", "other"}

// CheckManifest runs the checks that need no bytes against m, collecting
// every reason rather than stopping at the first. It returns an error, and
// no reasons, for a manifest that is malformed rather than refused.
//
// A file's reason is the first of: unsafe-path, path-too-deep, then, for
// loose files, bad-extension. Extras count toward the album's size.
func CheckManifest(m Manifest, p Policy) (Refusal, error) {
	seen := make(map[string]bool, len(m.Files))
	for _, f := range m.Files {
		switch {
		case f.Size < 0:
			return Refusal{}, fmt.Errorf("%w: %q has a negative size", ErrMalformedManifest, f.Path)
		case seen[f.Path]:
			return Refusal{}, fmt.Errorf("%w: %q appears twice", ErrMalformedManifest, f.Path)
		case m.Unit == UnitTrack && m.Container == ContainerFiles && !isAudioExt(extension(f.Path)):
			return Refusal{}, fmt.Errorf("%w: %q", ErrExtraInTracks, f.Path)
		}
		seen[f.Path] = true
	}

	var r Refusal
	var total int64
	tooLarge, audio := false, 0
	for _, f := range m.Files {
		ext := extension(f.Path)
		if reason, bad := p.pathFault(f.Path, ext, m.Container); bad {
			r.Files = append(r.Files, FileReason{Path: []byte(f.Path), Reason: reason})
		}
		if isAudioExt(ext) {
			audio++
		}
		// Compare against the room left, so a sum past int64 cannot wrap.
		switch {
		case tooLarge:
		case f.Size > p.MaxAlbumBytes-total:
			tooLarge = true
		default:
			total += f.Size
		}
	}

	if len(m.Files) > p.MaxFiles {
		r.Submission = append(r.Submission, reject.Of(reject.TooManyFiles))
	}
	if tooLarge {
		r.Submission = append(r.Submission, reject.Of(reject.TooLarge))
	}
	hasAudio := audio > 0
	if m.Container == ContainerZip {
		hasAudio = len(m.Files) == 1 && extension(m.Files[0].Path) == ".zip"
	}
	if !hasAudio {
		r.Submission = append(r.Submission, reject.Of(reject.NoAudio))
	}
	if !slices.Contains(sources, m.Source) {
		r.Submission = append(r.Submission, reject.Of(reject.MissingSource))
	}
	return r, nil
}

// pathFault returns the reason a declared path is refused, and false when
// it passes. A ZIP's one path is never a music file, so only loose files
// meet the extension check.
func (p Policy) pathFault(name, ext string, c Container) (reject.Reason, bool) {
	switch {
	case !ValidPath([]byte(name)):
		return reject.Of(reject.UnsafePath), true
	case strings.Count(name, "/")+1 > p.MaxPathDepth:
		return reject.Of(reject.PathTooDeep), true
	case c == ContainerFiles && isAudioExt(ext):
		if _, ok := p.admitted(ext); !ok {
			return reject.WithFormats(p.Allowlist), true
		}
	}
	return reject.Reason{}, false
}
