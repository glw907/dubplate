package unidecode

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// installVerb is the command the installer's refusals and [RowStatus]
// name as the fix for a missing or invalid table.
const installVerb = "unidecode install"

// leafName is the one directory under the state directory that the
// installed table lives in.
const leafName = "unidecode"

// leafPerm and tablePerm are the leaf's and the table file's modes:
// world-readable, since the table is public data whose integrity comes
// from its checksum at every load, never from its mode.
const (
	leafPerm  = 0o755
	tablePerm = 0o644
)

// InstallOptions carries what [Install] takes from its caller.
type InstallOptions struct {
	// Client downloads the pinned wheel when no local wheel is given.
	// Its Timeout bounds the download.
	Client *http.Client

	// Geteuid reports the effective uid that the owner checks compare
	// the state directory's and the leaf's owners against.
	Geteuid func() int
}

// tablePath returns the table file that [Install] writes and
// [RowStatus] reads, under stateDir's unidecode leaf.
func tablePath(stateDir string) string {
	return filepath.Join(stateDir, leafName, TableFileName)
}

// Install builds the pinned table into stateDir's unidecode leaf and
// writes a line naming it to stdout. The wheel comes from the file
// at from when from is set, and from the pinned URL over opts.Client
// otherwise.
//
// Every refusal below writes nothing: stateDir absent or owned by
// another user, and a leaf that exists but is not a directory of mode
// 0755 owned by the effective user. Only past those does Install check
// whether a valid table is already installed, so a run as another user
// against a valid table still refuses. Otherwise it obtains the wheel,
// builds the table, creates the leaf, and writes the table atomically,
// reading it back through [Parse] before it reports success.
func Install(ctx context.Context, stateDir, from string, stdout io.Writer, opts InstallOptions) error {
	stateInfo, err := os.Stat(stateDir)
	if err != nil {
		return fmt.Errorf("state directory %s: %w", stateDir, err)
	}
	if err := checkOwner(stateInfo, stateDir, opts.Geteuid); err != nil {
		return err
	}

	leafPath := filepath.Join(stateDir, leafName)
	leafInfo, statErr := os.Stat(leafPath)
	leafExists := statErr == nil
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("leaf %s: %w", leafPath, statErr)
	}
	if leafExists {
		if err := checkLeafShape(leafInfo, leafPath); err != nil {
			return err
		}
		if err := checkOwner(leafInfo, leafPath, opts.Geteuid); err != nil {
			return err
		}
	}

	path := tablePath(stateDir)
	if leafExists {
		if existing, err := os.ReadFile(path); err == nil {
			if _, err := Parse(existing); err == nil {
				_, err := fmt.Fprintf(stdout, "unidecode: already installed at %s (version %s)\n", path, Version)
				return err
			}
		}
	}

	table, err := obtainTable(ctx, from, opts.Client)
	if err != nil {
		return err
	}

	if !leafExists {
		// Mkdir's mode is masked by the umask, so the chmod sets it exactly.
		if err := os.Mkdir(leafPath, leafPerm); err != nil {
			return fmt.Errorf("create leaf %s: %w", leafPath, err)
		}
		if err := os.Chmod(leafPath, leafPerm); err != nil {
			return fmt.Errorf("chmod leaf %s: %w", leafPath, err)
		}
	}
	if err := atomicfile.WriteFile(path, table, tablePerm); err != nil {
		return fmt.Errorf("write table %s: %w", path, err)
	}

	written, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read back table %s: %w", path, err)
	}
	if _, err := Parse(written); err != nil {
		return fmt.Errorf("verify written table %s: %w", path, err)
	}

	_, err = fmt.Fprintf(stdout, "unidecode: installed %s (version %s)\n", path, Version)
	return err
}

// obtainTable builds the table from the wheel at from when from is
// set, or from the pinned wheel downloaded over client otherwise. A
// download refusal carries the offline guidance: the wheel's URL, its
// sha256, and --from.
func obtainTable(ctx context.Context, from string, client *http.Client) ([]byte, error) {
	if from != "" {
		wheel, err := readWheelFile(from)
		if err == nil {
			var table []byte
			if table, err = Build(wheel); err == nil {
				return table, nil
			}
		}
		return nil, fmt.Errorf("--from %s: %w", from, err)
	}
	wheel, err := Fetch(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("%w; download the wheel from %s (sha256 %s) and retry with --from, or pass --from directly",
			err, WheelURL, WheelSHA256)
	}
	return Build(wheel)
}

// readWheelFile reads path as an offline wheel through [ReadWheel],
// which caps the read and checks the checksum.
func readWheelFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if info.IsDir() {
		return nil, errors.New("is a directory, want a file")
	}
	return ReadWheel(f)
}

// checkOwner refuses when info's owning uid differs from geteuid(),
// naming the owner and the environment-file fix: the table is written
// into a directory the server's user owns, so another user's run would
// leave a table the server cannot replace.
func checkOwner(info fs.FileInfo, path string, geteuid func() int) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("%s: cannot determine its owner", path)
	}
	if int(st.Uid) == geteuid() {
		return nil
	}
	name := ownerName(st.Uid)
	return fmt.Errorf("%s owned by %s: run dubplate-server %s as %s with the server's environment file loaded",
		path, name, installVerb, name)
}

// checkLeafShape refuses when info, the leaf's stat, is not a
// directory of mode leafPerm: a regular file at path, or a directory at
// any other mode.
func checkLeafShape(info fs.FileInfo, path string) error {
	if !info.IsDir() {
		return fmt.Errorf("leaf %s: not a directory, want a directory: remove it and rerun dubplate-server %s",
			path, installVerb)
	}
	if mode := info.Mode().Perm(); mode != leafPerm {
		return fmt.Errorf("leaf %s: mode %04o, want %04o: fix its mode and rerun dubplate-server %s",
			path, mode, leafPerm, installVerb)
	}
	return nil
}

// ownerName resolves uid to a username, falling back to "uid:<n>" when
// the uid has no passwd entry.
func ownerName(uid uint32) string {
	u, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return fmt.Sprintf("uid:%d", uid)
	}
	return u.Username
}
