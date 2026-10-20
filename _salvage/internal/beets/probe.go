// probe.go is the dry beets-matcher probe internal/serve runs against a
// staged upload before promoting it into the inbox: the real matcher and
// the real MusicBrainz lookups, with none of the side effects a live
// import has.
package beets

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// ErrProbeUnavailable marks a probe that could not run at all: the beet
// binary missing, MusicBrainz unreachable, the scratch environment
// unwritable, or the context expiring mid-lookup. It is deliberately
// distinct from a probe that ran cleanly and found no confident match.
// A caller tells the two apart with [errors.Is]: the first is an honest
// retry-later state, the second a considered verdict the contributor
// can act on.
var ErrProbeUnavailable = errors.New("beets: matcher probe could not run")

// ProbeResult is one dry matcher run's verdict over a staged upload
// directory.
type ProbeResult struct {
	// Confident reports whether beets matched the directory to exactly
	// one release, confidently enough that its own quiet mode applied
	// the match rather than falling back to a skip.
	Confident bool
	// Artist, Album, and Year are the matched release's metadata, read
	// back from the throwaway library the probe imported into. They are
	// set only when Confident is true.
	Artist string
	Album  string
	Year   string
	// Detail is a short plain-language note explaining a not-confident
	// verdict, for a caller to log or carry forward. It is empty on a
	// confident match.
	Detail string
}

// detailNoMatch and detailManyAlbums are the notes a not-confident probe
// carries back. The uploader page's own no-match copy is the contract
// (docs/research/passb-copy-final.md, section 5); these
// explain the verdict to whoever reads the log, and never replace it.
const (
	detailNoMatch    = "MusicBrainz didn't have a confident match for these tracks."
	detailManyAlbums = "These files look like more than one album."
)

// probeWaitDelay bounds how long one killed beet invocation may keep
// [Runner.runProbe] waiting on its output pipes after ctx has already
// expired.
const probeWaitDelay = 10 * time.Second

// probeFieldSep separates the fields of one album line in the probe's
// `beet ls` output: an ASCII unit separator, which no artist name,
// album title, or year can itself contain, so no quoting or escaping is
// needed to parse a line back apart.
const probeFieldSep = "\x1f"

// probeListFormat is the `beet ls -a -f` template that reads the matched
// release back out of the probe's throwaway library.
const probeListFormat = "$albumartist" + probeFieldSep + "$album" + probeFieldSep + "$year"

// probeOverlayConfig is the configuration layered on top of the site's
// own beets config for a probe run, with the scratch directory's
// library, item directory, and state file substituted in. Every key
// here exists to make the run dry or hermetic; nothing that shapes
// which release beets picks is set, so the probe's confidence means the
// same thing the importer's later `beet import` will mean.
//
// beets' own --pretend is not the dry mode this needs: pretend replaces
// the whole lookup pipeline with a file listing, so it never consults
// MusicBrainz and could never report a match. A real import into a
// throwaway library with copy, move, link, and write all off is what
// runs the matcher without touching anything.
//
// plugins is pinned to musicbrainz alone. The site config's own list
// carries fetchart, embedart, and replaygain, all of which edit the
// contributor's files or fetch artwork on import; a probe must leave the
// staged directory byte for byte as it found it.
//
// raise_on_error turns a MusicBrainz outage into a nonzero exit instead
// of a logged warning and an empty candidate list, which is what lets
// [Runner.Probe] report ErrProbeUnavailable rather than a false
// not-confident verdict.
//
// quiet_fallback, none_rec_action, and timid are pinned because the
// probe's whole meaning rests on them: a site that set quiet_fallback to
// asis would have the probe "match" every album, and a timid run would
// block forever on input that never comes.
//
// create_backup_before_migrations is off because every probe starts on a
// brand new database, so beets' own schema migrations would otherwise
// copy eleven backup files into the scratch directory for nothing.
const probeOverlayConfig = `library: %s
directory: %s
statefile: %s
plugins: musicbrainz
raise_on_error: yes
create_backup_before_migrations: no
import:
  copy: no
  move: no
  link: no
  hardlink: no
  reflink: no
  write: no
  delete: no
  resume: no
  incremental: no
  from_scratch: no
  pretend: no
  autotag: yes
  quiet: yes
  quiet_fallback: skip
  none_rec_action: skip
  timid: no
  duplicate_action: skip
`

// Probe runs the real beets matcher over dir and reports whether it
// found a confident match, per the custody policy internal/serve's
// session completion enforces. No file under dir is moved, copied, or
// retagged, and the production beets library database is never opened,
// let alone mutated: the run imports into a throwaway library inside a
// scratch directory under r.StateDir, which is removed before Probe
// returns.
//
// r.ConfigPath is staged into that scratch directory as the probe's
// user config, with BEETSDIR pointing beets at it, so the probe sees
// exactly the site's matching settings and nothing from the ambient
// environment; [probeOverlayConfig] then pins only what makes the run
// dry.
//
// ctx bounds both beet invocations. Every failure path returns an error
// wrapping [ErrProbeUnavailable]; a clean run that matched nothing
// returns a zero-Confident result and a nil error.
func (r *Runner) Probe(ctx context.Context, dir string) (ProbeResult, error) {
	scratch, err := os.MkdirTemp(r.StateDir, "probe-*")
	if err != nil {
		return ProbeResult{}, fmt.Errorf("create probe scratch directory in %s: %w: %w", r.StateDir, ErrProbeUnavailable, err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	overlayPath, err := r.writeProbeConfig(scratch)
	if err != nil {
		return ProbeResult{}, err
	}

	logPath := filepath.Join(scratch, "probe.log")
	if _, err := r.runProbe(ctx, scratch, overlayPath, "import", "-q", "-l", logPath, dir); err != nil {
		return ProbeResult{}, err
	}

	// beets logs a skip for anything its quiet mode declined to
	// auto-tag. One skip is enough to push the whole upload back: the
	// album is the verdict unit, so a partial match is not a match.
	tagLog, err := os.ReadFile(logPath)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return ProbeResult{}, fmt.Errorf("read probe tag log %s: %w: %w", logPath, ErrProbeUnavailable, err)
	}
	if len(parseSkips(tagLog)) > 0 {
		return ProbeResult{Detail: detailNoMatch}, nil
	}

	out, err := r.runProbe(ctx, scratch, overlayPath, "ls", "-a", "-f", probeListFormat)
	if err != nil {
		return ProbeResult{}, err
	}
	albums := parseProbeAlbums(out)
	switch len(albums) {
	case 1:
		result := albums[0]
		result.Confident = true
		return result, nil
	case 0:
		return ProbeResult{Detail: detailNoMatch}, nil
	default:
		// One upload is one album (the uploader page's own format rule),
		// so several matched releases is a considered no, not a match to
		// pick a winner from.
		return ProbeResult{Detail: detailManyAlbums}, nil
	}
}

// writeProbeConfig populates scratch with the probe's beets
// configuration -- a copy of r.ConfigPath as the user config beets finds
// through BEETSDIR, plus the [probeOverlayConfig] overlay -- and returns
// the overlay's path, which the caller passes as beet's -c. An empty
// r.ConfigPath stages no user config at all, leaving beets' own
// defaults under the same overlay.
func (r *Runner) writeProbeConfig(scratch string) (string, error) {
	if r.ConfigPath != "" {
		site, err := os.ReadFile(r.ConfigPath)
		if err != nil {
			return "", fmt.Errorf("read beets config %s: %w: %w", r.ConfigPath, ErrProbeUnavailable, err)
		}
		if err := atomicfile.WriteFile(filepath.Join(scratch, "config.yaml"), site, 0o600); err != nil {
			return "", fmt.Errorf("stage probe user config: %w: %w", ErrProbeUnavailable, err)
		}
	}

	overlay := fmt.Sprintf(probeOverlayConfig,
		yamlString(filepath.Join(scratch, "probe-library.db")),
		yamlString(filepath.Join(scratch, "library")),
		yamlString(filepath.Join(scratch, "state.pickle")),
	)
	path := filepath.Join(scratch, "probe-overlay.yaml")
	if err := atomicfile.WriteFile(path, []byte(overlay), 0o600); err != nil {
		return "", fmt.Errorf("write probe overlay config: %w: %w", ErrProbeUnavailable, err)
	}
	return path, nil
}

// yamlString renders s as a YAML double-quoted scalar. A double-quoted
// YAML scalar takes the same backslash escapes JSON does, which is what
// makes [strconv.Quote] the right renderer here rather than a
// hand-rolled one: a scratch path holding a quote, a backslash, or a
// control byte still round-trips.
func yamlString(s string) string {
	return strconv.Quote(s)
}

// runProbe runs one beet invocation inside the probe's scratch
// environment and returns its standard output. BEETSDIR points beets at
// the staged user config; HOME is redirected into the same scratch
// directory so nothing beets or a plugin decides to cache lands outside
// it. Standard error is kept separate from the returned output so a
// warning can never be parsed as an album line.
//
// Any failure is an [ErrProbeUnavailable]: a missing binary, a nonzero
// exit (raise_on_error turns a MusicBrainz outage into one), or ctx
// expiring mid-lookup, which is reported as the context's own error so
// a caller can still match on it.
func (r *Runner) runProbe(ctx context.Context, scratch, overlayPath string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, r.BeetPath, append([]string{"-c", overlayPath}, args...)...)
	cmd.Env = append(os.Environ(), "BEETSDIR="+scratch, "HOME="+scratch)
	// Killing beet on a canceled ctx is not enough on its own: a
	// subprocess of its own that inherited the output pipes would keep
	// Wait blocked long past the deadline the caller set. WaitDelay
	// closes the pipes out from under it instead.
	cmd.WaitDelay = probeWaitDelay
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, fmt.Errorf("beet %s: %w: %w", args[0], ErrProbeUnavailable, ctxErr)
		}
		return nil, fmt.Errorf("beet %s: %w: %w: %s", args[0], ErrProbeUnavailable, err, strings.TrimSpace(stderr.String()))
	}
	return stdout.Bytes(), nil
}

// parseProbeAlbums turns the probe's `beet ls` output into one
// [ProbeResult] per matched release. A line carrying anything other
// than the format's two separators is beets chatter rather than an
// album line, and is dropped. A year of "0" is beets' own "unknown" and
// becomes an empty string rather than a misleading zero.
func parseProbeAlbums(out []byte) []ProbeResult {
	var albums []ProbeResult
	for line := range strings.Lines(string(out)) {
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			continue
		}
		fields := strings.Split(line, probeFieldSep)
		if len(fields) != 3 {
			continue
		}
		year := fields[2]
		if year == "0" {
			year = ""
		}
		albums = append(albums, ProbeResult{Artist: fields[0], Album: fields[1], Year: year})
	}
	return albums
}
