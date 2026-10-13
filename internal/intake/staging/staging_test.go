package staging

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
)

// dir builds a Dir that reads cleanly: device 7, mount 40, the mount
// id returned.
func dir(path, resolved string) Dir {
	return Dir{Path: path, Resolved: resolved, Dev: 7, Mask: mntIDMask, MountID: 40}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name     string
		staging  Dir
		library  Dir
		wantNil  bool
		contains []string
	}{
		{
			name:    "all match",
			staging: dir("/srv/music/staging", "/srv/music/staging"),
			library: dir("/srv/music/library", "/srv/music/library"),
			wantNil: true,
		},
		{
			name:     "staging absent",
			staging:  Dir{Path: "/srv/music/staging", Err: fmt.Errorf("stat: %w", fs.ErrNotExist)},
			library:  dir("/srv/music/library", "/srv/music/library"),
			contains: []string{"staging directory /srv/music/staging does not exist"},
		},
		{
			name:     "staging not writable",
			staging:  Dir{Path: "/srv/music/staging", Err: errors.New("write access to /srv/music/staging: read-only file system")},
			library:  dir("/srv/music/library", "/srv/music/library"),
			contains: []string{"staging directory", "/srv/music/staging", "write access"},
		},
		{
			name:     "library root absent",
			staging:  dir("/srv/music/staging", "/srv/music/staging"),
			library:  Dir{Path: "/srv/music/library", Err: fmt.Errorf("stat: %w", fs.ErrNotExist)},
			contains: []string{"library root /srv/music/library does not exist"},
		},
		{
			name:     "library inside staging after symlink resolution",
			staging:  dir("/srv/music/staging", "/srv/music/staging"),
			library:  dir("/srv/music/link", "/srv/music/staging/inner"),
			contains: []string{"nest", "/srv/music/staging", "/srv/music/staging/inner"},
		},
		{
			name:     "staging inside library after symlink resolution",
			staging:  dir("/srv/music/link", "/srv/music/library/stage"),
			library:  dir("/srv/music/library", "/srv/music/library"),
			contains: []string{"nest", "/srv/music/library/stage"},
		},
		{
			name:     "both resolve to one directory",
			staging:  dir("/a", "/srv/music"),
			library:  dir("/b", "/srv/music"),
			contains: []string{"nest"},
		},
		{
			name:    "sibling prefix is not nesting",
			staging: dir("/srv/music/stage", "/srv/music/stage"),
			library: dir("/srv/music/stage-two", "/srv/music/stage-two"),
			wantNil: true,
		},
		{
			name:     "devices differ",
			staging:  dir("/srv/music/staging", "/srv/music/staging"),
			library:  Dir{Path: "/srv/music/library", Resolved: "/srv/music/library", Dev: 8, Mask: mntIDMask, MountID: 40},
			contains: []string{"different devices", "device 7", "device 8"},
		},
		{
			name:     "bind mount keeps the device but not the mount",
			staging:  dir("/srv/music/staging", "/srv/music/staging"),
			library:  Dir{Path: "/srv/music/library", Resolved: "/srv/music/library", Dev: 7, Mask: mntIDMask, MountID: 41},
			contains: []string{"one mount", "mount 40", "mount 41"},
		},
		{
			name:     "statx fails",
			staging:  Dir{Path: "/srv/music/staging", Resolved: "/srv/music/staging", Dev: 7, StatxErr: errors.New("statx /srv/music/staging: function not implemented")},
			library:  dir("/srv/music/library", "/srv/music/library"),
			contains: []string{"statx /srv/music/staging", "function not implemented"},
		},
		{
			name:     "mask lacks the mount id",
			staging:  dir("/srv/music/staging", "/srv/music/staging"),
			library:  Dir{Path: "/srv/music/library", Resolved: "/srv/music/library", Dev: 7, Mask: 0x7ff, MountID: 40},
			contains: []string{"library root", "STATX_MNT_ID", "0x7ff"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Check(tt.staging, tt.library)
			if tt.wantNil {
				if err != nil {
					t.Fatalf("Check = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Check = nil, want a refusal")
			}
			for _, s := range tt.contains {
				if !strings.Contains(err.Error(), s) {
					t.Errorf("Check = %q, want it to contain %q", err, s)
				}
			}
		})
	}
}
