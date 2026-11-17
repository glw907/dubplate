package webapp

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unicode/utf8"
)

// operatorCSSMaxBytes bounds the operator CSS file, 256 KiB: well above
// any hand-written override, and small enough that a wrong path naming
// a large file refuses startup rather than loading it into memory.
const operatorCSSMaxBytes = 256 << 10

// ReadOperatorCSS reads the operator's stylesheet at path. The route
// that serves it is public, so the file's class is checked before its
// bytes are kept: a wrong path would otherwise publish whatever the
// service user can read, an environment file included.
//
// The path must end in ".css", and so must the path it resolves to
// once every symbolic link in it is followed, which is the path
// opened: a link named x.css cannot stand in for another file. The
// resolved path must not lie under stateDir, unless stateDir is "".
// The file is opened once, with O_NOFOLLOW so a link swapped in for
// the file itself after resolution is refused, and with O_NONBLOCK so
// a FIFO cannot hang startup; the class check reads the open
// descriptor, so the file checked is the file read. O_NOFOLLOW reaches
// the last path element alone: a directory on the resolved path
// swapped for a link after resolution is still followed, so the checks
// above hold only while no one else can write a directory on the path.
// The content must be UTF-8 text holding no NUL byte and no more than
// 256 KiB. A returned error names the rule or the reason and never
// carries the path or any byte of the file.
func ReadOperatorCSS(path, stateDir string) ([]byte, error) {
	if !strings.HasSuffix(path, ".css") {
		return nil, errors.New(`the path must end in ".css"`)
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	if !strings.HasSuffix(resolved, ".css") {
		return nil, errors.New(`the path must resolve to a file ending in ".css"`)
	}
	if stateDir != "" {
		resolvedStateDir, err := resolvePath(stateDir)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve the state directory: %w", withoutPath(err))
		}
		if rel, err := filepath.Rel(resolvedStateDir, resolved); err == nil && filepath.IsLocal(rel) {
			return nil, errors.New("the file is under the state directory")
		}
	}

	f, err := os.OpenFile(resolved, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("the path is not a regular file")
	}
	css, err := io.ReadAll(io.LimitReader(f, operatorCSSMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	switch {
	case len(css) > operatorCSSMaxBytes:
		return nil, fmt.Errorf("the file is over %d bytes", operatorCSSMaxBytes)
	case !utf8.Valid(css):
		return nil, errors.New("the file is not valid UTF-8")
	case bytes.IndexByte(css, 0) >= 0:
		return nil, errors.New("the file holds a NUL byte")
	}
	return css, nil
}

// resolvePath returns path made absolute with every symbolic link in
// it followed.
func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// withoutPath returns the error inside err's *fs.PathError, or err
// itself when it carries none, so a reported reason ("no such file or
// directory") keeps its errors.Is identity and drops the path.
func withoutPath(err error) error {
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		return pathErr.Err
	}
	return err
}
