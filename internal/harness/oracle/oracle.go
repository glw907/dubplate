// Package oracle drives beets against a [cassette.Case]: the adapter
// half lives here in Go, the shim half in derive_oracle.py alongside
// it, because the shim is the code that actually knows beets, and a
// future beets bump must be able to edit only this package. The shim
// runs under a uv-pinned venv (tools/venv) that the repository gate
// never requires; [Derive] and [Reconcile] are exercised by the
// oracle-tagged tests `make derive` runs. Derive owns a run: rendering
// the case's config profile against a [stub.Server], launching exactly
// one shim process, and validating the record the shim reports before
// handing it back as a [cassette.Expected].
package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// Options carries the pinned Python venv and, for tests, a shim
// override.
type Options struct {
	Venv string // path to the pinned uv venv; unused when Shim is set
	Shim string // shim executable path; empty means the package's own shim
	Dir  string // case directory to derive against, overriding the corpus/cases/<ID> default; authoritative for any directory, staging or committed
}

// missFileName is the file Derive writes into a case's own directory
// whenever its stub sees a request the case log did not cover: the
// plumbing corpus/README.md documents between a derivation and
// `harness capture --case <dir>`, which reads this file to know
// exactly what to fetch next.
const missFileName = "misses.json"

// metadataSourcePlugins are the beets plugins that supply match
// candidates (beetsplug/musicbrainz.py and its siblings); fetchart,
// embedart, and replaygain never count even if a profile listed them,
// since they are matching-inert.
var metadataSourcePlugins = map[string]bool{
	"musicbrainz": true,
	"discogs":     true,
	"spotify":     true,
	"deezer":      true,
	"beatport":    true,
}

// loadProfile reads the named profile's raw YAML from corpus/profiles
// under the repository root. Tests substitute this var to exercise
// Derive against profile content the checked-in corpus never carries.
var loadProfile = func(name string) ([]byte, error) {
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, "corpus", "profiles", name+".yaml"))
}

// caseDir resolves c's on-disk directory, corpus/cases/<ID> under the
// repository root, the same directory [cassette.LoadCase] read it
// from. Tests substitute this var to point at a scratch fixture
// instead of the checked-in corpus.
var caseDir = func(c *cassette.Case) (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "corpus", "cases", c.ID), nil
}

// repoRoot walks up from the working directory to the nearest
// ancestor holding go.mod, so profile and case lookup works whether
// the caller runs from the repository root (`make derive`) or from a
// package directory (`go test`).
func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("oracle: locate repo root: %w", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("oracle: locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}

// Derive runs the oracle for one case: it renders the case's config
// profile with the stub's MusicBrainz host substituted, starts a
// replay [stub.Server] over the case's captured responses, launches
// exactly one shim process with PYTHONHASHSEED=0 in its environment,
// and decodes and validates the record before returning it.
//
// Derive fails loudly, never silently, when the shim's echoed
// evaluation date disagrees with c.EvalDate, when the shim omits its
// config hash or its thresholds document, or when the rendered
// profile does not load exactly one metadata-source plugin.
func Derive(ctx context.Context, c *cassette.Case, opts Options) (*cassette.Expected, error) {
	raw, err := deriveRaw(ctx, c, opts)
	if err != nil {
		return nil, err
	}

	var exp cassette.Expected
	if err := json.Unmarshal(raw, &exp); err != nil {
		return nil, fmt.Errorf("oracle: derive %s: decode shim output: %w", c.ID, err)
	}
	if exp.EchoedDate != c.EvalDate {
		return nil, fmt.Errorf("oracle: derive %s: shim echoed date %q, case wants %q", c.ID, exp.EchoedDate, c.EvalDate)
	}
	if exp.ConfigHash == "" {
		return nil, fmt.Errorf("oracle: derive %s: shim omitted its config hash", c.ID)
	}
	if err := validateThresholdsPresent(raw); err != nil {
		return nil, fmt.Errorf("oracle: derive %s: %w", c.ID, err)
	}
	return &exp, nil
}

// validateThresholdsPresent proves raw's document carried a
// "thresholds" key at all. cassette.Thresholds decodes to its zero
// value whether the shim omitted the key or reported every threshold
// as zero, so presence is checked against the raw document rather
// than the decoded struct.
func validateThresholdsPresent(raw []byte) error {
	var probe struct {
		Thresholds json.RawMessage `json:"thresholds"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return fmt.Errorf("decode shim output: %w", err)
	}
	if len(probe.Thresholds) == 0 {
		return errors.New("shim omitted its thresholds")
	}
	return nil
}

// deriveRaw runs one shim process for c and returns its stdout
// unparsed, which is what the determinism proof diffs: a Go-decoded
// record would hide Python-side key ordering or float formatting drift.
func deriveRaw(ctx context.Context, c *cassette.Case, opts Options) ([]byte, error) {
	dir := opts.Dir
	if dir == "" {
		var err error
		dir, err = caseDir(c)
		if err != nil {
			return nil, err
		}
	}

	srv, err := stub.New(c, dir)
	if err != nil {
		return nil, fmt.Errorf("oracle: start stub for %s: %w", c.ID, err)
	}
	defer func() { _ = srv.Close() }()

	scratch, err := os.MkdirTemp("", "dubplate-oracle-*")
	if err != nil {
		return nil, fmt.Errorf("oracle: scratch dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	profile, err := renderProfile(c.Profile, scratch, filepath.Join(scratch, "library.db"), srv.Addr())
	if err != nil {
		return nil, err
	}
	profilePath := filepath.Join(scratch, "config.yaml")
	if err := os.WriteFile(profilePath, profile, 0o644); err != nil {
		return nil, fmt.Errorf("oracle: write rendered profile: %w", err)
	}

	inputPath := filepath.Join(scratch, "input.json")
	if err := writeShimInput(inputPath, c); err != nil {
		return nil, err
	}

	raw, err := runShim(ctx, opts, profilePath, inputPath)
	// The stub's miss list is collected regardless of runShim's own
	// outcome: a shim failure is often caused by a request the case
	// log never covered, so the miss file is diagnostic either way.
	// A write failure here only surfaces when the shim itself
	// succeeded; the shim's own error always takes priority.
	if missErr := writeMissFile(dir, srv.Misses()); missErr != nil && err == nil {
		err = fmt.Errorf("write miss file: %w", missErr)
	}
	if err != nil {
		return nil, fmt.Errorf("oracle: derive %s: %w", c.ID, err)
	}
	return raw, nil
}

// writeMissFile records srv's unmatched requests to dir/misses.json,
// or removes any stale file when there were none, so a case
// directory's presence of the file always reflects its most recent
// derivation.
func writeMissFile(dir string, misses []stub.Miss) error {
	path := filepath.Join(dir, missFileName)
	if len(misses) == 0 {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("remove stale miss file: %w", err)
		}
		return nil
	}

	data, err := json.MarshalIndent(misses, "", "  ")
	if err != nil {
		return fmt.Errorf("encode miss file: %w", err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write miss file: %w", err)
	}
	return nil
}

// renderProfile loads the named profile, substitutes its box-path
// placeholders with libDir/libDB and appends a musicbrainz: block
// pointed at host (the Q5 seam: {host, https: no, ratelimit: 100}),
// after checking the profile resolves to exactly one metadata source.
func renderProfile(name, libDir, libDB, host string) ([]byte, error) {
	raw, err := loadProfile(name)
	if err != nil {
		return nil, fmt.Errorf("oracle: load profile %s: %w", name, err)
	}

	rendered := strings.NewReplacer(
		"{{DUBPLATE_LIBRARY_DIR}}", libDir,
		"{{DUBPLATE_LIBRARY_DB}}", libDB,
	).Replace(string(raw))

	if n := countMetadataSources(rendered); n != 1 {
		return nil, fmt.Errorf("oracle: profile %s: rendered plugins line names %d metadata sources, want exactly 1", name, n)
	}

	rendered += fmt.Sprintf("\nmusicbrainz:\n  host: %q\n  https: no\n  ratelimit: 100\n", host)
	return []byte(rendered), nil
}

// countMetadataSources returns how many recognized metadata-source
// plugins profile's plugins: line names. It returns 0 when the
// profile carries no plugins: line at all.
func countMetadataSources(profile string) int {
	for line := range strings.SplitSeq(profile, "\n") {
		rest, ok := strings.CutPrefix(strings.TrimSpace(line), "plugins:")
		if !ok {
			continue
		}
		n := 0
		for p := range strings.FieldsSeq(rest) {
			if metadataSourcePlugins[p] {
				n++
			}
		}
		return n
	}
	return 0
}

// shimInput is the envelope Derive writes for the shim to read: the
// case's synthetic items plus the evaluation date the shim must pin
// its own process clock to.
type shimInput struct {
	CaseID   string                `json:"case_id"`
	EvalDate string                `json:"eval_date"`
	Items    []cassette.ItemFields `json:"items"`
}

func writeShimInput(path string, c *cassette.Case) error {
	data, err := json.Marshal(shimInput{CaseID: c.ID, EvalDate: c.EvalDate, Items: c.Input.Items})
	if err != nil {
		return fmt.Errorf("oracle: encode shim input: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("oracle: write shim input: %w", err)
	}
	return nil
}

// runShim launches exactly one shim process against the rendered
// profile and input, returns its stdout verbatim, and reports the
// shim's stderr on failure.
func runShim(ctx context.Context, opts Options, profilePath, inputPath string) ([]byte, error) {
	cmd, err := shimCommand(ctx, opts, profilePath, inputPath)
	if err != nil {
		return nil, err
	}
	cmd.Env = append(os.Environ(), "PYTHONHASHSEED=0")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("run shim: %w: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

// shimCommand builds the shim invocation: opts.Shim exec'd directly
// when set (tests point it at testdata/fakeshim), or, when empty, the
// venv's python3 running derive_oracle.py. The venv branch is covered
// only by the oracle-tagged tests, since the gate has no venv.
func shimCommand(ctx context.Context, opts Options, args ...string) (*exec.Cmd, error) {
	if opts.Shim != "" {
		return exec.CommandContext(ctx, opts.Shim, args...), nil
	}
	root, err := repoRoot()
	if err != nil {
		return nil, err
	}
	python := filepath.Join(opts.Venv, "bin", "python3")
	script := filepath.Join(root, "internal", "harness", "oracle", "derive_oracle.py")
	return exec.CommandContext(ctx, python, append([]string{script}, args...)...), nil
}
