package main

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
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
	"github.com/spf13/cobra"
)

// unidecodeInstallVerbName is the verb this file's own refusals and
// doctor's row name as the fix for a missing or invalid table.
const unidecodeInstallVerbName = "unidecode install"

// unidecodeDoctorRow is the doctor row's name.
const unidecodeDoctorRow = "unidecode"

// unidecodeLeafName is the one directory under the state directory the
// installed table lives in.
const unidecodeLeafName = "unidecode"

// unidecodeLeafPerm and unidecodeTablePerm are the leaf's and the
// table file's modes: the shadow leaf's own tier, since the table is
// public data whose integrity comes from its checksum at every load,
// never from its mode.
const (
	unidecodeLeafPerm  = 0o755
	unidecodeTablePerm = 0o644
)

// unidecodeInstallTimeout bounds the install verb's wheel download.
const unidecodeInstallTimeout = 30 * time.Second

// geteuid is the effective-uid seam; a test swaps it to exercise an owner mismatch.
var geteuid = os.Geteuid

// unidecodeClient is the install verb's download client; a test swaps
// it for one that reaches a stub server.
var unidecodeClient = &http.Client{Timeout: unidecodeInstallTimeout}

// unidecodeTablePath returns the table file the install verb writes
// and both loads below read, under stateDir's unidecode leaf.
func unidecodeTablePath(stateDir string) string {
	return filepath.Join(stateDir, unidecodeLeafName, unidecode.TableFileName)
}

// loadUnidecodeTable reads and verifies the table at stateDir's
// unidecode leaf. It is the one load doctor's row and every scoring
// verb share: it touches the network never, and creates nothing.
func loadUnidecodeTable(stateDir string) (*unidecode.Table, error) {
	path := unidecodeTablePath(stateDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	table, err := unidecode.Parse(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return table, nil
}

// setUnidecodeTransliterator loads stateDir's table and installs its
// lookup as the matcher's transliterator. A scoring verb calls it
// before shadow.OpenStore, so a missing or invalid table never leaves
// a half-started store.
func setUnidecodeTransliterator(stateDir string) error {
	table, err := loadUnidecodeTable(stateDir)
	if err != nil {
		return fmt.Errorf("load transliteration table: %w; run dubplate-server %s",
			err, unidecodeInstallVerbName)
	}
	matcher.SetTransliterator(table.Lookup)
	return nil
}

// unidecodeRowStatus renders doctor's "unidecode" row: "not checked"
// when DUBPLATE_STATE_DIR is unset (the wall row already fails), "ok
// (<version>)" for a table Parse accepts, "missing: ..." when the read
// fails because the state directory, leaf, or file is absent, and
// "invalid: ..." for every other failure (a directory in the file's
// place, permission denied, a checksum or format refusal).
func unidecodeRowStatus(environ []string) (status string, failed bool) {
	stateDir, ok := lookupEnviron(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return "not checked", false
	}
	_, err := loadUnidecodeTable(stateDir)
	if err == nil {
		return fmt.Sprintf("ok (%s)", unidecode.Version), false
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Sprintf("missing: run dubplate-server %s", unidecodeInstallVerbName), true
	}
	return fmt.Sprintf("invalid: %s: run dubplate-server %s", err, unidecodeInstallVerbName), true
}

// newUnidecodeCmd builds the unidecode subcommand's parent: today it
// carries the install verb alone.
func newUnidecodeCmd(environ []string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "unidecode",
		Short:         "Manage the pinned Unidecode transliteration table",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetErr(stderr)
	cmd.AddCommand(newUnidecodeInstallCmd(environ, stdout, stderr))
	return cmd
}

// newUnidecodeInstallCmd builds the install verb: it reads
// DUBPLATE_STATE_DIR alone, never config.Load, because setup runs it
// before the rest of the wall may be rendered.
func newUnidecodeInstallCmd(environ []string, stdout, stderr io.Writer) *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the pinned Unidecode transliteration table",
		Long: fmt.Sprintf(
			"Downloads Unidecode %s's pinned wheel from %s (sha256 %s) and\n"+
				"builds its transliteration table into the state directory. Offline,\n"+
				"download the wheel yourself and pass --from its path.",
			unidecode.Version, unidecode.WheelURL, unidecode.WheelSHA256),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := runUnidecodeInstall(cmd.Context(), environ, from, stdout); err != nil {
				return fmt.Errorf("unidecode install: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "install from a local wheel file instead of downloading")
	cmd.SetErr(stderr)
	return cmd
}

// runUnidecodeInstall runs the install verb's contract, every
// refusal below writing nothing: DUBPLATE_STATE_DIR unset; the state
// directory absent or owned by another user; the unidecode leaf, when
// it already exists, not a unidecodeLeafPerm directory owned by the
// effective user. Only past every refusal does it check whether a
// valid table is already installed, so a root run against a valid
// table still refuses. Otherwise it obtains the wheel, builds the
// table, creates the leaf, and writes the table atomically, reading it
// back through Parse before reporting success.
func runUnidecodeInstall(ctx context.Context, environ []string, from string, stdout io.Writer) error {
	stateDir, ok := lookupEnviron(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return errors.New("DUBPLATE_STATE_DIR is not set")
	}

	stateInfo, err := os.Stat(stateDir)
	if err != nil {
		return fmt.Errorf("state directory %s: %w", stateDir, err)
	}
	if err := checkUnidecodeOwner(stateInfo, stateDir); err != nil {
		return err
	}

	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	leafInfo, statErr := os.Stat(leafPath)
	leafExists := statErr == nil
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("leaf %s: %w", leafPath, statErr)
	}
	if leafExists {
		if err := checkUnidecodeLeafShape(leafInfo, leafPath); err != nil {
			return err
		}
		if err := checkUnidecodeOwner(leafInfo, leafPath); err != nil {
			return err
		}
	}

	tablePath := filepath.Join(leafPath, unidecode.TableFileName)
	if leafExists {
		if existing, err := os.ReadFile(tablePath); err == nil {
			if _, err := unidecode.Parse(existing); err == nil {
				_, err := fmt.Fprintf(stdout, "unidecode: already installed at %s (version %s)\n", tablePath, unidecode.Version)
				return err
			}
		}
	}

	table, err := obtainUnidecodeTable(ctx, from)
	if err != nil {
		return err
	}

	if !leafExists {
		// Mkdir's mode is masked by the umask, so the chmod sets it exactly.
		if err := os.Mkdir(leafPath, unidecodeLeafPerm); err != nil {
			return fmt.Errorf("create leaf %s: %w", leafPath, err)
		}
		if err := os.Chmod(leafPath, unidecodeLeafPerm); err != nil {
			return fmt.Errorf("chmod leaf %s: %w", leafPath, err)
		}
	}
	if err := atomicfile.WriteFile(tablePath, table, unidecodeTablePerm); err != nil {
		return fmt.Errorf("write table %s: %w", tablePath, err)
	}

	written, err := os.ReadFile(tablePath)
	if err != nil {
		return fmt.Errorf("read back table %s: %w", tablePath, err)
	}
	if _, err := unidecode.Parse(written); err != nil {
		return fmt.Errorf("verify written table %s: %w", tablePath, err)
	}

	_, err = fmt.Fprintf(stdout, "unidecode: installed %s (version %s)\n", tablePath, unidecode.Version)
	return err
}

// obtainUnidecodeTable builds the table from the wheel at from when
// from is set, or from the pinned wheel downloaded over unidecodeClient
// otherwise. A download refusal carries the offline guidance: the
// wheel's URL, its sha256, and --from.
func obtainUnidecodeTable(ctx context.Context, from string) ([]byte, error) {
	if from != "" {
		wheel, err := readUnidecodeWheelFile(from)
		if err == nil {
			var table []byte
			if table, err = unidecode.Build(wheel); err == nil {
				return table, nil
			}
		}
		return nil, fmt.Errorf("--from %s: %w", from, err)
	}
	wheel, err := unidecode.Fetch(ctx, unidecodeClient)
	if err != nil {
		return nil, fmt.Errorf("%w; download the wheel from %s (sha256 %s) and retry with --from, or pass --from directly",
			err, unidecode.WheelURL, unidecode.WheelSHA256)
	}
	return unidecode.Build(wheel)
}

// readUnidecodeWheelFile reads path as an offline wheel through
// unidecode.ReadWheel, which caps the read and checks the checksum.
func readUnidecodeWheelFile(path string) ([]byte, error) {
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
	return unidecode.ReadWheel(f)
}

// checkUnidecodeOwner refuses when info's owning uid differs from
// geteuid(), naming the owner and the runuser fix, leaf.go's own
// checkOwner precedent.
func checkUnidecodeOwner(info fs.FileInfo, path string) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("%s: cannot determine its owner", path)
	}
	if int(st.Uid) == geteuid() {
		return nil
	}
	name := unidecodeOwnerName(st.Uid)
	return fmt.Errorf("%s owned by %s: run as that user: runuser -u %s -- dubplate-server %s",
		path, name, name, unidecodeInstallVerbName)
}

// checkUnidecodeLeafShape refuses when info, the leaf's own stat, is
// not a unidecodeLeafPerm directory: a regular file at path, or a
// directory at any other mode.
func checkUnidecodeLeafShape(info fs.FileInfo, path string) error {
	if !info.IsDir() {
		return fmt.Errorf("leaf %s: not a directory, want a directory: remove it and rerun dubplate-server %s",
			path, unidecodeInstallVerbName)
	}
	if mode := info.Mode().Perm(); mode != unidecodeLeafPerm {
		return fmt.Errorf("leaf %s: mode %04o, want %04o: fix its mode and rerun dubplate-server %s",
			path, mode, unidecodeLeafPerm, unidecodeInstallVerbName)
	}
	return nil
}

// unidecodeOwnerName resolves uid to a username, falling back to
// "uid:<n>" when the uid has no passwd entry.
func unidecodeOwnerName(uid uint32) string {
	u, err := user.LookupId(strconv.FormatUint(uint64(uid), 10))
	if err != nil {
		return fmt.Sprintf("uid:%d", uid)
	}
	return u.Username
}
