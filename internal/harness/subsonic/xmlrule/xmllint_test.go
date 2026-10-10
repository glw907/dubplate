package xmlrule

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// The test binary stands in for xmllint when it runs under fakeVersion:
// it reports that libxml version, and for a validation it copies the file
// named by fakeStderr to stderr and exits with fakeExit.
const (
	fakeVersion = "XMLRULE_FAKE_XMLLINT"
	fakeStderr  = "XMLRULE_FAKE_STDERR"
	fakeExit    = "XMLRULE_FAKE_EXIT"
)

func TestMain(m *testing.M) {
	if os.Getenv(fakeVersion) != "" {
		os.Exit(fakeXmllint())
	}
	os.Exit(m.Run())
}

func fakeXmllint() int {
	// The version line goes to stderr and nothing to stdout, as the
	// real xmllint writes it.
	if slices.Contains(os.Args[1:], "--version") {
		fmt.Fprintf(os.Stderr, "xmllint: using libxml version %s\n   compiled with: Schemas\n", os.Getenv(fakeVersion))
		return 0
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	if data, err := os.ReadFile(os.Getenv(fakeStderr)); err == nil {
		_, _ = os.Stderr.Write(data)
	}
	code, _ := strconv.Atoi(os.Getenv(fakeExit))
	return code
}

// fakeRunner puts the test binary on an otherwise empty PATH as xmllint,
// reporting version, and returns the runner newRunner builds over it.
func fakeRunner(t *testing.T, version, stderr string, exit int) (*runner, error) {
	t.Helper()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.Symlink(exe, filepath.Join(dir, "xmllint")); err != nil {
		t.Fatal(err)
	}
	stderrFile := filepath.Join(dir, "stderr")
	if err := os.WriteFile(stderrFile, []byte(stderr), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv(fakeVersion, version)
	t.Setenv(fakeStderr, stderrFile)
	t.Setenv(fakeExit, strconv.Itoa(exit))
	return newRunner(context.Background(), "xmllint")
}

func TestNewRunnerAbsent(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	_, err := newRunner(context.Background(), "xmllint")
	if !errors.Is(err, exec.ErrNotFound) {
		t.Fatalf("err = %v, want exec.ErrNotFound", err)
	}
	if !strings.Contains(err.Error(), "xmllint") {
		t.Errorf("err = %v, want it to name xmllint", err)
	}
}

func TestNewRunnerVersion(t *testing.T) {
	t.Run("another version is refused", func(t *testing.T) {
		_, err := fakeRunner(t, "21210", "", 0)
		if !errors.Is(err, errVersion) {
			t.Fatalf("err = %v, want errVersion", err)
		}
		path := filepath.Join(os.Getenv("PATH"), "xmllint")
		for _, want := range []string{"21210", pinnedLibxml, path, "brew unpin libxml2", "brew pin libxml2"} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("err = %v, want it to name %q", err, want)
			}
		}
	})
	t.Run("the pinned version is accepted", func(t *testing.T) {
		r, err := fakeRunner(t, pinnedLibxml, "", 0)
		if err != nil {
			t.Fatal(err)
		}
		if want := filepath.Join(os.Getenv("PATH"), "xmllint"); r.path != want {
			t.Errorf("path = %s, want %s", r.path, want)
		}
	})
}

func TestValidateToolFailure(t *testing.T) {
	tests := []struct {
		name   string
		stderr string
		exit   int
	}{
		{"exit 4", "-:1: parser error : Premature end of data in tag a line 1\n", 4},
		{"exit 5", "", 5},
		{"exit 3 with no validity error", "- fails to validate\n", 3},
		{"exit 3 with an unrecognized first line", "-:1: parser error : boom\n- fails to validate\n", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := fakeRunner(t, pinnedLibxml, tt.stderr, tt.exit)
			if err != nil {
				t.Fatal(err)
			}
			errs, err := r.validate(context.Background(), "unused.xsd", []byte("<a/>"))
			if !errors.Is(err, ErrToolFailure) {
				t.Fatalf("errs, err = %v, %v; want ErrToolFailure", errs, err)
			}
		})
	}
}

// TestValidateLongStderr floods stderr past 64 KiB, where
// ExitError.Stderr keeps only a prefix and a suffix.
func TestValidateLongStderr(t *testing.T) {
	const n = 1000
	var b strings.Builder
	for i := range n {
		fmt.Fprintf(&b, "-:%d: Schemas validity error : Element '{urn:x}song', attribute 'a%04d': "+
			"'x' is not a valid value of the atomic type 'xs:int'.\n", i+1, i)
	}
	b.WriteString("- fails to validate\n")
	if b.Len() <= 64<<10 {
		t.Fatalf("stderr is %d bytes, want more than 64 KiB", b.Len())
	}
	r, err := fakeRunner(t, pinnedLibxml, b.String(), 3)
	if err != nil {
		t.Fatal(err)
	}
	errs, err := r.validate(context.Background(), "unused.xsd", []byte("<a/>"))
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != n {
		t.Fatalf("got %d errors, want %d", len(errs), n)
	}
	for i, e := range errs {
		if want := fmt.Sprintf("a%04d", i); e.line != i+1 || e.attr != want {
			t.Fatalf("errs[%d] = line %d attr %q, want line %d attr %q", i, e.line, e.attr, i+1, want)
		}
	}
}

func TestParseValidity(t *testing.T) {
	tests := []struct {
		name    string
		stderr  string
		want    []schemaError
		wantErr bool
	}{
		{
			name: "two errors and the trailer",
			stderr: "-:2: Schemas validity error : Element '{urn:x}album', attribute 'songCount': 'x' is not a valid value of the atomic type 'xs:int'.\n" +
				"-:3: Schemas validity error : Element '{urn:x}song': The attribute 'id' is required but missing.\n" +
				"- fails to validate\n",
			want: []schemaError{
				{2, "songCount", "Element '{urn:x}album', attribute 'songCount': 'x' is not a valid value of the atomic type 'xs:int'."},
				{3, "id", "Element '{urn:x}song': The attribute 'id' is required but missing."},
			},
		},
		{
			name:   "an error naming no attribute",
			stderr: "-:4: Schemas validity error : Element '{urn:x}roles': This element is not expected.\n- fails to validate\n",
			want:   []schemaError{{4, "", "Element '{urn:x}roles': This element is not expected."}},
		},
		{
			name: "a wrapped message joins its error",
			stderr: "-:1: Schemas validity error : Element '{urn:x}song', attribute 'track': 'x\n" +
				"y' is not a valid value of the atomic type 'xs:int'.\n- fails to validate\n",
			want: []schemaError{{1, "track", "Element '{urn:x}song', attribute 'track': 'x\ny' is not a valid value of the atomic type 'xs:int'."}},
		},
		{
			name:    "an unrecognized line before any error",
			stderr:  "warning: something\n-:1: Schemas validity error : Element 'a': bad.\n- fails to validate\n",
			wantErr: true,
		},
		{
			name:    "an unrecognized line after the trailer",
			stderr:  "-:1: Schemas validity error : Element 'a': bad.\n- fails to validate\nstray\n",
			wantErr: true,
		},
		{
			name:    "no error at all",
			stderr:  "- fails to validate\n",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseValidity(tt.stderr)
			if tt.wantErr {
				if !errors.Is(err, ErrToolFailure) {
					t.Fatalf("got %v, %v; want ErrToolFailure", got, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("got %#v\nwant %#v", got, tt.want)
			}
		})
	}
}
