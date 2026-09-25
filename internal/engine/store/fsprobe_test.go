package store

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// ext4SuperMagic (linux/magic.h) is the "opens" control value in
// TestProbeFilesystem: every named refusal or warning constant beside
// one filesystem this package must never refuse.
const ext4SuperMagic = 0xEF53

func TestProbeFilesystem(t *testing.T) {
	tests := []struct {
		name        string
		magic       int64
		wantVerdict fsVerdict
	}{
		{name: "NFS refuses", magic: nfsSuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "CIFS refuses", magic: cifsSuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "SMB2 refuses", magic: smb2SuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "9p refuses", magic: v9fsMagic, wantVerdict: fsRefuseMultiHost},
		{name: "smbfs refuses", magic: smbSuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "ceph refuses", magic: cephSuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "AFS (first magic.h constant) refuses", magic: afsSuperMagic, wantVerdict: fsRefuseMultiHost},
		{name: "AFS (second magic.h constant) refuses", magic: afsFsMagic, wantVerdict: fsRefuseMultiHost},
		{name: "FUSE warns", magic: fuseSuperMagic, wantVerdict: fsWarnFUSE},
		{name: "ext4 opens", magic: ext4SuperMagic, wantVerdict: fsOpen},
	}

	// A real leaf on t.TempDir(), so the seam is the only thing each
	// subtest varies.
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	if err := l.create(stateDir, os.Geteuid()); err != nil {
		t.Fatalf("create: %v", err)
	}
	target := l.probeTarget()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := statfsFunc
			t.Cleanup(func() { statfsFunc = orig })
			statfsFunc = func(_ string, buf *syscall.Statfs_t) error {
				buf.Type = tt.magic
				return nil
			}

			verdict, magic, err := probeFilesystem(target)
			if err != nil {
				t.Fatalf("probeFilesystem: %v", err)
			}
			if verdict != tt.wantVerdict {
				t.Errorf("probeFilesystem verdict = %v, want %v", verdict, tt.wantVerdict)
			}
			if magic != tt.magic {
				t.Errorf("probeFilesystem magic = %#x, want %#x", magic, tt.magic)
			}
		})
	}
}

func TestProbeFilesystem_RealTempDirOpens(t *testing.T) {
	// t.TempDir() sits on tmpfs or btrfs on this workstation (the
	// mechanics review measured both), neither of which this package
	// refuses or warns on.
	verdict, _, err := probeFilesystem(t.TempDir())
	if err != nil {
		t.Fatalf("probeFilesystem: %v", err)
	}
	if verdict != fsOpen {
		t.Errorf("probeFilesystem(t.TempDir()) verdict = %v, want fsOpen", verdict)
	}
}

func TestProbeTarget(t *testing.T) {
	stateDir := t.TempDir()
	l := newLeaf(stateDir)

	if got, want := l.probeTarget(), stateDir; got != want {
		t.Errorf("probeTarget with no leaf = %q, want the leaf's parent %q", got, want)
	}

	if err := os.MkdirAll(l.root, leafDirPerm); err != nil {
		t.Fatal(err)
	}
	if got, want := l.probeTarget(), l.root; got != want {
		t.Errorf("probeTarget with a leaf and no db/ = %q, want the leaf %q", got, want)
	}

	if err := os.MkdirAll(l.dbDir, leafDirPerm); err != nil {
		t.Fatal(err)
	}
	if got, want := l.probeTarget(), l.dbDir; got != want {
		t.Errorf("probeTarget with db/ present = %q, want db/ %q", got, want)
	}
}

func TestPreCreate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target")

	if err := preCreate(path); err != nil {
		t.Fatalf("preCreate: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if info.IsDir() {
		t.Errorf("%s: is a directory, want a regular file", path)
	}
	if size := info.Size(); size != 0 {
		t.Errorf("preCreate: file size = %d, want 0", size)
	}
}

func TestPreCreate_UmaskCannotWiden0600(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	dir := t.TempDir()
	path := filepath.Join(dir, "target")
	if err := preCreate(path); err != nil {
		t.Fatalf("preCreate: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", path, err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("preCreate under umask 0022: mode = %o, want 0600", mode)
	}
}

func TestPreCreate_ExistingTargetRefusedUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "target")
	if err := os.WriteFile(path, []byte("keep me"), 0o600); err != nil {
		t.Fatal(err)
	}

	err := preCreate(path)
	if err == nil {
		t.Fatal("preCreate: want a refusal for an existing target, got nil")
	}
	if !errors.Is(err, fs.ErrExist) {
		t.Errorf("preCreate error %v does not wrap fs.ErrExist", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep me" {
		t.Errorf("preCreate touched the existing file: content = %q", got)
	}
}
