package xmlrule

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// pinnedLibxml is the libxml version the tier accepts, as xmllint
// --version prints it. Its messages land in goldens, so another build
// is refused rather than met as a golden diff.
const pinnedLibxml = "21504"

var (
	// errVersion reports an xmllint built against another libxml.
	errVersion = errors.New("xmlrule: xmllint version is not the pinned one")

	// ErrToolFailure reports an xmllint run that ended in neither a
	// validation result nor a set of validity errors.
	ErrToolFailure = errors.New("xmlrule: xmllint failed")
)

var (
	versionLine  = regexp.MustCompile(`using libxml version (\d+)`)
	validityLine = regexp.MustCompile(`^-:(\d+): Schemas validity error : (.*)$`)
	// Two message forms name an attribute: a bad value and a missing
	// required one.
	attrNamed   = regexp.MustCompile(`^Element '[^']*', attribute '([^']*)':`)
	attrMissing = regexp.MustCompile(`The attribute '([^']*)' is required but missing`)
)

// runner is an xmllint whose libxml version has been checked.
type runner struct {
	path string
}

// systemXmllint resolves and checks the xmllint on PATH once per
// process. The probe runs under a background context: a once cannot
// carry one caller's cancellation to every later caller.
var systemXmllint = sync.OnceValues(func() (*runner, error) {
	return newRunner(context.Background(), "xmllint")
})

// newRunner resolves name through PATH and checks that it reports the
// pinned libxml version. xmllint --version writes to stderr.
func newRunner(ctx context.Context, name string) (*runner, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return nil, fmt.Errorf("xmlrule: resolve %s: %w; install Homebrew's libxml2 and hold it with brew pin libxml2", name, err)
	}
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, path, "--version")
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("xmlrule: %s --version: %w: %s", path, err, stderr.Bytes())
	}
	m := versionLine.FindSubmatch(stderr.Bytes())
	got := "none"
	if m != nil {
		got = string(m[1])
	}
	if got != pinnedLibxml {
		return nil, fmt.Errorf("%w: %s reports libxml %s, want %s; to move the pin run "+
			"brew unpin libxml2 && brew upgrade libxml2, regenerate the goldens, brew pin libxml2, "+
			"then brew upgrade the dependents the pin held",
			errVersion, path, got, pinnedLibxml)
	}
	return &runner{path: path}, nil
}

// schemaError is one validity error xmllint reports.
type schemaError struct {
	line int
	attr string
	msg  string
}

// validate runs xmllint over doc against the schema at schemaPath and
// returns its validity errors, none for a valid document. Stderr goes to
// the runner's own buffer: ExitError.Stderr keeps only a prefix and a
// suffix of long output.
func (r *runner) validate(ctx context.Context, schemaPath string, doc []byte) ([]schemaError, error) {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, r.path, "--noout", "--schema", schemaPath, "-")
	cmd.Stdin = bytes.NewReader(doc)
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err == nil {
		return nil, nil
	}
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 3 {
		return nil, fmt.Errorf("%w: %w: %s", ErrToolFailure, err, stderr.Bytes())
	}
	return parseValidity(stderr.String())
}

// parseValidity reads xmllint's stderr under exit 3, for a document read
// from stdin, which xmllint names "-". Each validity error is a line of
// the form "-:<line>: Schemas validity error : ", and a
// line that follows one without that form continues its message, since a
// value holding a newline wraps it. The "- fails to validate"
// trailer is expected. Any other line is a tool failure.
func parseValidity(stderr string) ([]schemaError, error) {
	var errs []schemaError
	inError := false
	for line := range strings.Lines(stderr) {
		line = strings.TrimSuffix(line, "\n")
		if m := validityLine.FindStringSubmatch(line); m != nil {
			n, err := strconv.Atoi(m[1])
			if err != nil {
				return nil, fmt.Errorf("%w: line number in %q: %w", ErrToolFailure, line, err)
			}
			errs = append(errs, schemaError{line: n, msg: m[2]})
			inError = true
			continue
		}
		if line == "- fails to validate" {
			inError = false
			continue
		}
		if !inError {
			return nil, fmt.Errorf("%w: unrecognized output %q", ErrToolFailure, line)
		}
		errs[len(errs)-1].msg += "\n" + line
	}
	if len(errs) == 0 {
		return nil, fmt.Errorf("%w: exit 3 with no validity error: %s", ErrToolFailure, stderr)
	}
	for i := range errs {
		errs[i].attr = attributeOf(errs[i].msg)
	}
	return errs, nil
}

// attributeOf returns the attribute an xmllint message names, or "".
func attributeOf(msg string) string {
	if m := attrNamed.FindStringSubmatch(msg); m != nil {
		return m[1]
	}
	if m := attrMissing.FindStringSubmatch(msg); m != nil {
		return m[1]
	}
	return ""
}
