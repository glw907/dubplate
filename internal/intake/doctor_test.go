package intake

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/store"
	"golang.org/x/sys/unix"
)

// TestDoctorRows_RootsNeedAReadStore proves the volume row lists the
// collection roots only from a store inspection whose readers ran. A
// refused inspection can leave roots in Facts that its report does not
// vouch for. Fixture state: sibling staging and library directories,
// and Facts holding the root "abab", which is not created.
func TestDoctorRows_RootsNeedAReadStore(t *testing.T) {
	f := newVolumeFixture(t)
	facts := Facts{roots: [][]byte{[]byte("abab")}}
	tests := []struct {
		name   string
		report store.Report
		want   string
	}{
		{
			name:   "a refused inspection",
			report: store.Report{State: store.StateRefused},
			want:   "ok: staging and the library root share one mount",
		},
		{
			name:   "an inspection whose readers ran",
			report: store.Report{State: store.StateCurrent, Facts: &store.Facts{}},
			want:   "ok: staging and the library root share one mount; collection root " + f.library + "/abab not created",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows := DoctorRows(f.cfg, tt.report, facts)
			if len(rows) != 1 {
				t.Fatalf("DoctorRows() returned %d rows, want 1", len(rows))
			}
			if rows[0].Status != tt.want || rows[0].Failed {
				t.Errorf("DoctorRows() = %q (failed %v), want %q (not failed)", rows[0].Status, rows[0].Failed, tt.want)
			}
		})
	}
}

// volumeFixture is a staging directory and a library root, siblings
// under one temporary directory, with a config naming them.
type volumeFixture struct {
	staging, library string
	cfg              *config.Config
}

// newVolumeFixture creates both directories and loads the config that
// names them.
func newVolumeFixture(t *testing.T) volumeFixture {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("fixture state: resolve the temporary directory: %v", err)
	}
	f := volumeFixture{staging: filepath.Join(base, "staging"), library: filepath.Join(base, "library")}
	for _, dir := range []string{f.staging, f.library} {
		if err := os.Mkdir(dir, 0o755); err != nil {
			t.Fatalf("fixture state: %v", err)
		}
	}
	f.cfg = loadVolumeConfig(t, f.staging, f.library)
	return f
}

// loadVolumeConfig loads a config whose staging directory and library
// root are the given paths, which need not exist.
func loadVolumeConfig(t *testing.T, staging, library string) *config.Config {
	t.Helper()
	reg := config.NewRegistry()
	Register(reg)
	cfg, err := config.Load([]string{
		"DUBPLATE_LIBRARY_ROOT=" + library,
		"DUBPLATE_STAGING_DIR=" + staging,
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load: %v", err)
	}
	return cfg
}

// deviceOf returns st_dev of path as decimal text, the way the startup
// check words it.
func deviceOf(t *testing.T, path string) string {
	t.Helper()
	var st unix.Stat_t
	if err := unix.Stat(path, &st); err != nil {
		t.Fatalf("fixture state: stat %s: %v", path, err)
	}
	return strconv.FormatUint(uint64(st.Dev), 10)
}

func TestVolumeStatus(t *testing.T) {
	const same = "ok: staging and the library root share one mount"
	// A directory on a mount other than the temporary directory's: the
	// shared-memory tmpfs is a mount of its own on every Linux host.
	const otherMount = "/dev/shm"

	tests := []struct {
		name       string
		setup      func(t *testing.T, f *volumeFixture)
		roots      [][]byte
		want       func(_ *testing.T, f volumeFixture) string
		wantFailed bool
	}{
		{
			name:  "sibling directories and no collection roots",
			setup: func(*testing.T, *volumeFixture) {},
			want:  func(*testing.T, volumeFixture) string { return same },
		},
		{
			name:  "an empty list of roots",
			setup: func(*testing.T, *volumeFixture) {},
			roots: [][]byte{},
			want:  func(*testing.T, volumeFixture) string { return same },
		},
		{
			name: "a staging directory not created yet",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Remove(f.staging); err != nil {
					t.Fatal(err)
				}
			},
			want: func(_ *testing.T, f volumeFixture) string {
				return "staging directory " + f.staging + " does not exist"
			},
		},
		{
			name: "a library root not created yet",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Remove(f.library); err != nil {
					t.Fatal(err)
				}
			},
			want: func(_ *testing.T, f volumeFixture) string { return "library root " + f.library + " does not exist" },
		},
		{
			name: "neither directory created yet",
			setup: func(t *testing.T, f *volumeFixture) {
				for _, dir := range []string{f.staging, f.library} {
					if err := os.Remove(dir); err != nil {
						t.Fatal(err)
					}
				}
			},
			roots: [][]byte{[]byte("abab")},
			want: func(_ *testing.T, f volumeFixture) string {
				return "staging directory " + f.staging + " does not exist; library root " + f.library + " does not exist"
			},
		},
		{
			name: "staging that is a file",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Remove(f.staging); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f.staging, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			want: func(_ *testing.T, f volumeFixture) string {
				return "staging directory: " + f.staging + " is not a directory"
			},
			wantFailed: true,
		},
		{
			name: "staging inside the library root",
			setup: func(t *testing.T, f *volumeFixture) {
				f.staging = filepath.Join(f.library, "staging")
				if err := os.Mkdir(f.staging, 0o755); err != nil {
					t.Fatal(err)
				}
				f.cfg = loadVolumeConfig(t, f.staging, f.library)
			},
			want: func(_ *testing.T, f volumeFixture) string {
				return "staging directory " + f.staging + " and library root " + f.library +
					" nest (resolved to " + f.staging + " and " + f.library + ")"
			},
			wantFailed: true,
		},
		{
			name: "a staging directory on another device",
			setup: func(t *testing.T, f *volumeFixture) {
				f.staging = otherMount
				f.cfg = loadVolumeConfig(t, f.staging, f.library)
			},
			want: func(t *testing.T, f volumeFixture) string {
				return "staging directory /dev/shm (device " + deviceOf(t, otherMount) + ") and library root " +
					f.library + " (device " + deviceOf(t, f.library) + ") sit on different devices"
			},
			wantFailed: true,
		},
		{
			name:  "a collection root not created yet",
			setup: func(*testing.T, *volumeFixture) {},
			roots: [][]byte{[]byte("abab")},
			want: func(_ *testing.T, f volumeFixture) string {
				return same + "; collection root " + filepath.Join(f.library, "abab") + " not created"
			},
		},
		{
			name: "a collection root on the library root's mount",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Mkdir(filepath.Join(f.library, "abab"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			roots: [][]byte{[]byte("abab")},
			want: func(_ *testing.T, f volumeFixture) string {
				return same + "; collection root " + filepath.Join(f.library, "abab") + " on the library root's mount"
			},
		},
		{
			name: "a collection root that is a file",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.WriteFile(filepath.Join(f.library, "abab"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
			},
			roots: [][]byte{[]byte("abab")},
			want: func(_ *testing.T, f volumeFixture) string {
				return same + "; collection root: " + filepath.Join(f.library, "abab") + " is not a directory"
			},
			wantFailed: true,
		},
		{
			name: "a collection root on another mount",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Symlink(otherMount, filepath.Join(f.library, "abab")); err != nil {
					t.Fatal(err)
				}
			},
			roots: [][]byte{[]byte("abab")},
			want: func(_ *testing.T, f volumeFixture) string {
				return same + "; collection root " + filepath.Join(f.library, "abab") + " sits outside the library root's mount"
			},
			wantFailed: true,
		},
		{
			name: "one root off the mount fails the row whatever follows it",
			setup: func(t *testing.T, f *volumeFixture) {
				if err := os.Symlink(otherMount, filepath.Join(f.library, "aaaa")); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(filepath.Join(f.library, "bbbb"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			roots: [][]byte{[]byte("aaaa"), []byte("bbbb"), []byte("cccc")},
			want: func(_ *testing.T, f volumeFixture) string {
				return same +
					"; collection root " + filepath.Join(f.library, "aaaa") + " sits outside the library root's mount" +
					"; collection root " + filepath.Join(f.library, "bbbb") + " on the library root's mount" +
					"; collection root " + filepath.Join(f.library, "cccc") + " not created"
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newVolumeFixture(t)
			tt.setup(t, &f)
			status, failed := volumeStatus(f.cfg, tt.roots)
			if want := tt.want(t, f); status != want {
				t.Errorf("status = %q, want %q", status, want)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}

	t.Run("a refused environment", func(t *testing.T) {
		status, failed := volumeStatus(nil, [][]byte{[]byte("abab")})
		if status != "not checked" || failed {
			t.Errorf("volumeStatus(nil) = %q, %v, want \"not checked\", false", status, failed)
		}
	})
}
