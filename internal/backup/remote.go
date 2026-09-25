package backup

import (
	"fmt"
	"path"
	"strings"
)

// A Remote is a backup destination as rclone names it: a configured
// remote's name and a path under that remote.
type Remote struct {
	name string
	path string
}

// ParseRemote parses s as name:path, splitting at the first colon.
// The name must be one or more ASCII letters, digits, and underscores.
// It refuses a value with no colon, which rclone would read as a local
// path. The path may be empty, which rclone reads as the remote's
// root.
func ParseRemote(s string) (Remote, error) {
	name, p, ok := strings.Cut(s, ":")
	if !ok {
		return Remote{}, fmt.Errorf("backup remote %q: no colon: want name:path", s)
	}
	if !validName(name) {
		return Remote{}, fmt.Errorf("backup remote %q: name %q is not ASCII letters, digits, and underscores", s, name)
	}
	return Remote{name: name, path: p}, nil
}

// join returns the destination sub names under r. It joins the path
// part with [path.Join], never by concatenation, so an empty path gives
// name:sub rather than name:/sub, which rclone reads from the backend's
// root.
func (r Remote) join(sub string) string {
	return r.name + ":" + path.Join(r.path, sub)
}

func validName(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' {
			continue
		}
		return false
	}
	return true
}
