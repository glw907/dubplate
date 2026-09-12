// Command close-evidence prints a pass close task's evidence block, measured
// from the tree instead of reconstructed by an agent.
//
// Usage:
//
//	go run ./tools/close-evidence -range <a>..<b> [-pkgs <patterns>] [-repo <dir>]
//
// The range is any revision range git accepts, usually the pass's first commit
// parent through its merge. The patterns are the package patterns to time,
// comma separated, defaulting to the whole module.
//
// The block reports four numbers: the wall time of one uncached `go test` over
// the named packages, the test files the range changed, whether the range
// touched internal/serve or go.mod, and how many of its commits are refactors.
// Paste the block into the close task's evidence section as printed. A number
// produced here cannot be misremembered, which is what the fix rounds on the
// two measured closes were spent discovering.
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

func main() {
	cfg := config{}
	flag.StringVar(&cfg.revRange, "range", "", "git revision range, for example 4787e91..HEAD")
	flag.StringVar(&cfg.pkgs, "pkgs", "./...", "comma separated go package patterns to time")
	flag.StringVar(&cfg.repo, "repo", ".", "repository directory")
	flag.Parse()

	if err := run(context.Background(), cfg, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type config struct {
	revRange string
	pkgs     string
	repo     string
}

// refactorSubject matches the commit subjects counted as refactors. The close
// block reports the match count and the subjects, so a miscount is visible.
var refactorSubject = regexp.MustCompile(`(?i)\b(refactor|simplif)`)

func run(ctx context.Context, cfg config, out io.Writer) error {
	if cfg.revRange == "" {
		return errors.New("range is required")
	}
	patterns := splitPatterns(cfg.pkgs)
	if len(patterns) == 0 {
		return errors.New("pkgs names no package")
	}

	tests, err := timeTests(ctx, cfg.repo, patterns)
	if err != nil {
		return err
	}
	changedTests, err := git(ctx, cfg.repo, "diff", "--stat", cfg.revRange, "--", "*_test.go")
	if err != nil {
		return err
	}
	serve, err := changedPaths(ctx, cfg.repo, cfg.revRange, "internal/serve")
	if err != nil {
		return err
	}
	gomod, err := changedPaths(ctx, cfg.repo, cfg.revRange, "go.mod")
	if err != nil {
		return err
	}
	subjects, err := commitSubjects(ctx, cfg.repo, cfg.revRange)
	if err != nil {
		return err
	}

	var b strings.Builder
	fmt.Fprintf(&b, "## Close evidence: %s\n\n", cfg.revRange)
	fmt.Fprintf(&b, "Produced by tools/close-evidence.\n\n")

	fmt.Fprintf(&b, "### Test wall time\n\n")
	fmt.Fprintf(&b, "`go test -count=1 %s`: %s, %s\n\n", strings.Join(patterns, " "), tests.wall.Round(time.Millisecond), tests.result)
	if len(tests.packages) == 0 {
		fmt.Fprintf(&b, "No package reported an elapsed time.\n\n")
	} else {
		fmt.Fprintf(&b, "| package | elapsed |\n| --- | --- |\n")
		for _, p := range tests.packages {
			fmt.Fprintf(&b, "| %s | %s |\n", p.name, p.elapsed)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "### Test files changed\n\n")
	b.WriteString(block(changedTests, "No test file changed in the range."))

	fmt.Fprintf(&b, "### Surfaces touched\n\n")
	fmt.Fprintf(&b, "- internal/serve: %s\n", touched(serve))
	fmt.Fprintf(&b, "- go.mod: %s\n\n", touched(gomod))

	refactors := filterSubjects(subjects)
	fmt.Fprintf(&b, "### Refactor commits\n\n")
	fmt.Fprintf(&b, "%d of %d commits match refactor or simplif.\n", len(refactors), len(subjects))
	for _, s := range refactors {
		fmt.Fprintf(&b, "- %s\n", s)
	}

	_, err = io.WriteString(out, b.String())
	return err
}

// testRun is one `go test` invocation over the named patterns.
type testRun struct {
	wall     time.Duration
	result   string
	packages []packageTime
}

type packageTime struct {
	name    string
	elapsed string
}

// timeTests runs the packages once with caching off, so the wall time measures
// work rather than a cache hit. A failing suite is reported, not an error: the
// close block should say the suite was red rather than print nothing.
func timeTests(ctx context.Context, repo string, patterns []string) (testRun, error) {
	args := append([]string{"test", "-count=1"}, patterns...)
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = repo
	start := time.Now()
	output, err := cmd.CombinedOutput()
	run := testRun{
		wall:     time.Since(start),
		result:   "pass",
		packages: parsePackageTimes(string(output)),
	}
	if err != nil {
		if _, ok := errors.AsType[*exec.ExitError](err); !ok {
			return testRun{}, fmt.Errorf("run go test: %w", err)
		}
		run.result = "FAIL"
	}
	return run, nil
}

// parsePackageTimes reads the per-package elapsed times off `go test`'s summary
// lines. A cached or test-free package reports no duration and is skipped.
func parsePackageTimes(output string) []packageTime {
	var times []packageTime
	for line := range strings.SplitSeq(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 3 || fields[0] != "ok" {
			continue
		}
		if !strings.HasSuffix(fields[2], "s") {
			continue
		}
		times = append(times, packageTime{name: fields[1], elapsed: fields[2]})
	}
	return times
}

func commitSubjects(ctx context.Context, repo, revRange string) ([]string, error) {
	return gitLines(ctx, repo, "log", "--format=%h %s", revRange)
}

func filterSubjects(subjects []string) []string {
	var matched []string
	for _, s := range subjects {
		if refactorSubject.MatchString(s) {
			matched = append(matched, s)
		}
	}
	return matched
}

func changedPaths(ctx context.Context, repo, revRange, path string) ([]string, error) {
	return gitLines(ctx, repo, "diff", "--name-only", revRange, "--", path)
}

func touched(paths []string) string {
	if len(paths) == 0 {
		return "unchanged"
	}
	return fmt.Sprintf("changed (%d files): %s", len(paths), strings.Join(paths, ", "))
}

func git(ctx context.Context, repo string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repo
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return string(out), nil
}

// gitLines runs git and returns the non-empty lines of its output.
func gitLines(ctx context.Context, repo string, args ...string) ([]string, error) {
	out, err := git(ctx, repo, args...)
	if err != nil {
		return nil, err
	}
	return lines(out), nil
}

func lines(s string) []string {
	var got []string
	for line := range strings.SplitSeq(strings.TrimRight(s, "\n"), "\n") {
		if line != "" {
			got = append(got, line)
		}
	}
	return got
}

// block renders command output as a fenced section, or the empty note when the
// command printed nothing.
func block(out, empty string) string {
	if strings.TrimSpace(out) == "" {
		return empty + "\n\n"
	}
	return "```\n" + strings.TrimRight(out, "\n") + "\n```\n\n"
}

func splitPatterns(s string) []string {
	var patterns []string
	for p := range strings.SplitSeq(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			patterns = append(patterns, p)
		}
	}
	return patterns
}
