package beets

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writeProbeStub writes a stub beet that dispatches on the subcommand.
// [Runner.Probe] always invokes beet as "-c <overlay> <subcommand>
// ...", so the subcommand is always the third argument. importBody and
// lsBody are shell fragments run for their respective subcommands; an
// unexpected subcommand fails loudly rather than silently succeeding.
func writeProbeStub(t *testing.T, dir, importBody, lsBody string) string {
	t.Helper()
	script := fmt.Sprintf(`
case "$3" in
import)
%s
;;
ls)
%s
;;
*)
  echo "stub beet: unexpected subcommand $3" >&2
  exit 2
;;
esac
`, importBody, lsBody)
	return writeStubBeet(t, dir, script)
}

// logSkipBody is an import body that writes one beets skip line for the
// probed directory into the tag log beet was handed with -l, which is
// argument six of "-c <overlay> import -q -l <log> <dir>".
const logSkipBody = `echo "skip $7" >> "$6"`

// albumLine renders one line of the probe's `beet ls` output, matching
// [probeListFormat]'s field order and separator.
func albumLine(artist, album, year string) string {
	return strings.Join([]string{artist, album, year}, probeFieldSep)
}

// echoLines is an ls body that prints lines verbatim on stdout.
func echoLines(lines ...string) string {
	if len(lines) == 0 {
		return ":"
	}
	return fmt.Sprintf("printf '%%s\\n' %s", shellQuoteAll(lines))
}

func shellQuoteAll(args []string) string {
	quoted := make([]string, 0, len(args))
	for _, a := range args {
		quoted = append(quoted, "'"+strings.ReplaceAll(a, "'", `'\''`)+"'")
	}
	return strings.Join(quoted, " ")
}

// failLoudly is a body for a subcommand a case must never reach.
const failLoudly = `echo "stub beet: this subcommand should not have run" >&2; exit 3`

func TestRunnerProbe(t *testing.T) {
	tests := []struct {
		name       string
		importBody string
		lsBody     string
		timeout    time.Duration
		want       ProbeResult
		wantErr    bool
		wantErrIs  []error
	}{
		{
			name:       "confident match carries the release metadata",
			importBody: ":",
			lsBody:     echoLines(albumLine("Miles Davis", "Kind of Blue", "1959")),
			want: ProbeResult{
				Confident: true,
				Artist:    "Miles Davis",
				Album:     "Kind of Blue",
				Year:      "1959",
			},
		},
		{
			name:       "confident match with an unknown year reports no year",
			importBody: ":",
			lsBody:     echoLines(albumLine("Unlabelled", "Field Recordings", "0")),
			want: ProbeResult{
				Confident: true,
				Artist:    "Unlabelled",
				Album:     "Field Recordings",
			},
		},
		{
			name:       "a beets skip is a not-confident verdict, never a library read",
			importBody: logSkipBody,
			lsBody:     failLoudly,
			want:       ProbeResult{Detail: detailNoMatch},
		},
		{
			name:       "an empty library is a not-confident verdict",
			importBody: ":",
			lsBody:     echoLines(),
			want:       ProbeResult{Detail: detailNoMatch},
		},
		{
			name:       "two matched albums is a not-confident verdict",
			importBody: ":",
			lsBody: echoLines(
				albumLine("Miles Davis", "Kind of Blue", "1959"),
				albumLine("Bill Evans", "Waltz for Debby", "1961"),
			),
			want: ProbeResult{Detail: detailManyAlbums},
		},
		{
			name:       "beets chatter on stdout is not an album",
			importBody: ":",
			lsBody:     echoLines("No matching release found.", albumLine("Miles Davis", "Kind of Blue", "1959")),
			want: ProbeResult{
				Confident: true,
				Artist:    "Miles Davis",
				Album:     "Kind of Blue",
				Year:      "1959",
			},
		},
		{
			name:       "a failing import is unavailable, not a verdict",
			importBody: `echo "MusicBrainz not reachable" >&2; exit 1`,
			lsBody:     failLoudly,
			wantErr:    true,
			wantErrIs:  []error{ErrProbeUnavailable},
		},
		{
			name:       "a failing library read is unavailable, not a verdict",
			importBody: ":",
			lsBody:     `echo "database is locked" >&2; exit 1`,
			wantErr:    true,
			wantErrIs:  []error{ErrProbeUnavailable},
		},
		{
			name: "an expired context is unavailable, not a verdict",
			// The redirect keeps sleep off beet's own output pipes, so
			// killing the stub really does end the run.
			importBody: `sleep 30 >/dev/null 2>&1`,
			lsBody:     failLoudly,
			timeout:    50 * time.Millisecond,
			wantErr:    true,
			wantErrIs:  []error{ErrProbeUnavailable, context.DeadlineExceeded},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			binDir := t.TempDir()
			stateDir := t.TempDir()
			albumDir := t.TempDir()
			beetPath := writeProbeStub(t, binDir, tt.importBody, tt.lsBody)

			ctx := context.Background()
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}

			r := &Runner{BeetPath: beetPath, StateDir: stateDir}
			got, err := r.Probe(ctx, albumDir)

			switch {
			case tt.wantErr && err == nil:
				t.Fatalf("Probe() error = nil, want an error; result = %+v", got)
			case !tt.wantErr && err != nil:
				t.Fatalf("Probe() unexpected error: %v", err)
			}
			for _, target := range tt.wantErrIs {
				if !errors.Is(err, target) {
					t.Errorf("Probe() error = %v, want it to wrap %v", err, target)
				}
			}
			if !tt.wantErr {
				if got != tt.want {
					t.Errorf("Probe() = %+v, want %+v", got, tt.want)
				}
			} else if got != (ProbeResult{}) {
				t.Errorf("Probe() = %+v on an error, want the zero result", got)
			}

			// The scratch directory holds a throwaway library, a staged
			// config, and a tag log; none of it may outlive the run.
			leftovers, globErr := filepath.Glob(filepath.Join(stateDir, "probe-*"))
			if globErr != nil {
				t.Fatalf("glob probe scratch: %v", globErr)
			}
			if len(leftovers) > 0 {
				t.Errorf("Probe() left scratch directories behind: %v", leftovers)
			}
		})
	}
}

// TestRunnerProbeIsDry pins the invocation the dryness of the probe
// rests on: the argv shape, the hermetic BEETSDIR/HOME environment, the
// site config staged where beets will find it, and every overlay key
// that keeps the run from touching the contributor's files or the real
// library.
func TestRunnerProbeIsDry(t *testing.T) {
	binDir := t.TempDir()
	stateDir := t.TempDir()
	albumDir := t.TempDir()
	capture := filepath.Join(binDir, "capture")
	if err := os.MkdirAll(capture, 0o755); err != nil {
		t.Fatalf("mkdir capture: %v", err)
	}

	siteConfig := filepath.Join(binDir, "site-config.yaml")
	const siteConfigBody = "directory: /srv/music/library\nplugins: musicbrainz fetchart embedart replaygain\n"
	if err := os.WriteFile(siteConfig, []byte(siteConfigBody), 0o644); err != nil {
		t.Fatalf("write site config: %v", err)
	}

	importBody := fmt.Sprintf(`
for a in "$@"; do printf '%%s\n' "$a"; done > %[1]s/args
printf '%%s\n' "$BEETSDIR" "$HOME" > %[1]s/env
cat "$2" > %[1]s/overlay
cat "$BEETSDIR/config.yaml" > %[1]s/staged-config
`, capture)
	beetPath := writeProbeStub(t, binDir, importBody, echoLines())

	r := &Runner{BeetPath: beetPath, ConfigPath: siteConfig, StateDir: stateDir}
	if _, err := r.Probe(context.Background(), albumDir); err != nil {
		t.Fatalf("Probe() unexpected error: %v", err)
	}

	read := func(name string) string {
		t.Helper()
		data, err := os.ReadFile(filepath.Join(capture, name))
		if err != nil {
			t.Fatalf("read captured %s: %v", name, err)
		}
		return string(data)
	}

	args := strings.Split(strings.TrimRight(read("args"), "\n"), "\n")
	if len(args) != 7 {
		t.Fatalf("beet args = %v, want 7 of them", args)
	}
	wantFixed := map[int]string{0: "-c", 2: "import", 3: "-q", 4: "-l", 6: albumDir}
	for i, want := range wantFixed {
		if args[i] != want {
			t.Errorf("beet args[%d] = %q, want %q (full args: %v)", i, args[i], want, args)
		}
	}

	env := strings.Split(strings.TrimRight(read("env"), "\n"), "\n")
	if len(env) != 2 {
		t.Fatalf("captured env = %v, want BEETSDIR and HOME", env)
	}
	beetsDir, home := env[0], env[1]
	if beetsDir == "" {
		t.Fatal("BEETSDIR is empty; the probe would read the ambient user config")
	}
	if home != beetsDir {
		t.Errorf("HOME = %q, want it redirected into the scratch directory %q", home, beetsDir)
	}
	if filepath.Dir(beetsDir) != stateDir {
		t.Errorf("BEETSDIR = %q, want a scratch directory under %q", beetsDir, stateDir)
	}
	if got := filepath.Dir(args[1]); got != beetsDir {
		t.Errorf("overlay config %q is outside the scratch directory %q", args[1], beetsDir)
	}
	if got := filepath.Dir(args[5]); got != beetsDir {
		t.Errorf("tag log %q is outside the scratch directory %q", args[5], beetsDir)
	}

	if got := read("staged-config"); got != siteConfigBody {
		t.Errorf("staged user config = %q, want the site config %q", got, siteConfigBody)
	}

	overlay := read("overlay")
	wantKeys := []string{
		"plugins: musicbrainz",
		"raise_on_error: yes",
		"copy: no",
		"move: no",
		"link: no",
		"hardlink: no",
		"reflink: no",
		"write: no",
		"delete: no",
		"pretend: no",
		"quiet: yes",
		"quiet_fallback: skip",
		"timid: no",
	}
	for _, key := range wantKeys {
		if !strings.Contains(overlay, key) {
			t.Errorf("probe overlay is missing %q:\n%s", key, overlay)
		}
	}
	for _, path := range []string{"probe-library.db", "state.pickle"} {
		if !strings.Contains(overlay, filepath.Join(beetsDir, path)) {
			t.Errorf("probe overlay does not point %s at the scratch directory:\n%s", path, overlay)
		}
	}
}

// TestRunnerProbeMissingConfig proves an unreadable site config is
// reported as an unavailable probe rather than silently degrading to a
// verdict beets never actually reached.
func TestRunnerProbeMissingConfig(t *testing.T) {
	binDir := t.TempDir()
	stateDir := t.TempDir()
	beetPath := writeProbeStub(t, binDir, failLoudly, failLoudly)

	r := &Runner{BeetPath: beetPath, ConfigPath: filepath.Join(binDir, "absent.yaml"), StateDir: stateDir}
	_, err := r.Probe(context.Background(), t.TempDir())
	if !errors.Is(err, ErrProbeUnavailable) {
		t.Fatalf("Probe() error = %v, want it to wrap %v", err, ErrProbeUnavailable)
	}
}

func TestParseProbeAlbums(t *testing.T) {
	tests := []struct {
		name string
		out  string
		want []ProbeResult
	}{
		{
			name: "empty output",
			out:  "",
			want: nil,
		},
		{
			name: "one album",
			out:  albumLine("Miles Davis", "Kind of Blue", "1959") + "\n",
			want: []ProbeResult{{Artist: "Miles Davis", Album: "Kind of Blue", Year: "1959"}},
		},
		{
			name: "no trailing newline",
			out:  albumLine("Miles Davis", "Kind of Blue", "1959"),
			want: []ProbeResult{{Artist: "Miles Davis", Album: "Kind of Blue", Year: "1959"}},
		},
		{
			name: "unknown year",
			out:  albumLine("Miles Davis", "Kind of Blue", "0") + "\n",
			want: []ProbeResult{{Artist: "Miles Davis", Album: "Kind of Blue"}},
		},
		{
			name: "a line with the wrong field count is dropped",
			out:  "Sending event: import_begin\n" + albumLine("Miles Davis", "Kind of Blue", "1959") + "\n",
			want: []ProbeResult{{Artist: "Miles Davis", Album: "Kind of Blue", Year: "1959"}},
		},
		{
			name: "two albums",
			out: albumLine("Miles Davis", "Kind of Blue", "1959") + "\n" +
				albumLine("Bill Evans", "Waltz for Debby", "1961") + "\n",
			want: []ProbeResult{
				{Artist: "Miles Davis", Album: "Kind of Blue", Year: "1959"},
				{Artist: "Bill Evans", Album: "Waltz for Debby", Year: "1961"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProbeAlbums([]byte(tt.out))
			if len(got) != len(tt.want) {
				t.Fatalf("parseProbeAlbums() = %+v, want %+v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseProbeAlbums()[%d] = %+v, want %+v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
