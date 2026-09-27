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

// geteuid is the seam the install verb reads the effective uid
// through, internal/engine/store/leaf.go's own precedent. A test
// swaps it to exercise an owner mismatch without running as a
// different real user, restoring it with t.Cleanup; this package runs
// no test in parallel, which makes the swap safe.
var geteuid = os.Geteuid

// unidecodeTablePath returns the table file the install verb writes
// and both loads below read, under stateDir's unidecode leaf.
func unidecodeTablePath(stateDir string) string {
	return filepath.Join(stateDir, unidecodeLeafName, unidecode.TableFileName)
}

// loadUnidecodeTable reads and verifies the table at stateDir's
// unidecode leaf. It is the one load doctor's row and every scoring
// verb share: it touches the network never, and creates nothing.
func loadUnidecodeTable(stateDir string) (*unidecode.Table, error) {
	data, err := os.ReadFile(unidecodeTablePath(stateDir))
	if err != nil {
		return nil, err
	}
	return unidecode.Parse(data)
}

// installUnidecodeTransliterator loads stateDir's table and installs
// its lookup as the matcher's process-wide transliterator. A
// scoring verb calls it after its own configuration refusals and
// before shadow.OpenStore, so a missing or invalid table never leaves
// a half-started store; doctor never calls it.
func installUnidecodeTransliterator(stateDir string) error {
	table, err := loadUnidecodeTable(stateDir)
	if err != nil {
		return fmt.Errorf("load transliteration table at %s: %w; run dubplate-server %s",
			unidecodeTablePath(stateDir), err, unidecodeInstallVerbName)
	}
	matcher.SetTransliterator(table.Lookup)
	return nil
}

// unidecodeRowStatus renders doctor's "unidecode" row: "not checked"
// when DUBPLATE_STATE_DIR is unset (the wall row already fails), "ok
// (<version>)" for a table Parse accepts, "missing: ..." when the read
// fails because the state directory, leaf, or file is absent, and
// "invalid: ..." for every other failure (a directory in the file's
// place, permission denied, a checksum or format refusal). It calls
// loadUnidecodeTable alone.
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
func newUnidecodeCmd(environ []string, client *http.Client, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "unidecode",
		Short:         "Manage the pinned Unidecode transliteration table",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.SetErr(stderr)
	cmd.AddCommand(newUnidecodeInstallCmd(environ, client, stdout, stderr))
	return cmd
}

// newUnidecodeInstallCmd builds the install verb: it reads
// DUBPLATE_STATE_DIR alone, never config.Load, because setup runs it
// before the rest of the wall may be rendered.
func newUnidecodeInstallCmd(environ []string, client *http.Client, stdout, stderr io.Writer) *cobra.Command {
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
			return runUnidecodeInstall(cmd.Context(), client, environ, from, stdout)
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
func runUnidecodeInstall(ctx context.Context, client *http.Client, environ []string, from string, stdout io.Writer) error {
	stateDir, ok := lookupEnviron(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return errors.New("unidecode install: DUBPLATE_STATE_DIR is not set")
	}

	stateInfo, err := os.Stat(stateDir)
	if err != nil {
		return fmt.Errorf("unidecode install: state directory %s: %w", stateDir, err)
	}
	if err := checkUnidecodeOwner(stateInfo, stateDir); err != nil {
		return err
	}

	leafPath := filepath.Join(stateDir, unidecodeLeafName)
	leafInfo, statErr := os.Stat(leafPath)
	leafExists := statErr == nil
	if statErr != nil && !errors.Is(statErr, fs.ErrNotExist) {
		return fmt.Errorf("unidecode install: leaf %s: %w", leafPath, statErr)
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

	wheel, err := obtainUnidecodeWheel(ctx, client, from)
	if err != nil {
		return err
	}
	table, err := unidecode.Build(wheel)
	if err != nil {
		return err
	}

	// MkdirAll is a no-op on a directory that already exists, at
	// whatever mode it already carries, so an explicit chmod is the
	// only way this holds under any umask.
	if err := os.MkdirAll(leafPath, unidecodeLeafPerm); err != nil {
		return fmt.Errorf("unidecode install: create leaf %s: %w", leafPath, err)
	}
	if err := os.Chmod(leafPath, unidecodeLeafPerm); err != nil {
		return fmt.Errorf("unidecode install: chmod leaf %s: %w", leafPath, err)
	}
	if err := atomicfile.WriteFile(tablePath, table, unidecodeTablePerm); err != nil {
		return fmt.Errorf("unidecode install: write table %s: %w", tablePath, err)
	}

	written, err := os.ReadFile(tablePath)
	if err != nil {
		return fmt.Errorf("unidecode install: read back table %s: %w", tablePath, err)
	}
	if _, err := unidecode.Parse(written); err != nil {
		return fmt.Errorf("unidecode install: verify written table %s: %w", tablePath, err)
	}

	_, err = fmt.Fprintf(stdout, "unidecode: installed %s (version %s)\n", tablePath, unidecode.Version)
	return err
}

// obtainUnidecodeWheel returns from's bytes when from is set, or
// downloads the pinned wheel over client otherwise. A download refusal
// is wrapped with the offline guidance: the wheel's URL, its sha256,
// and --from.
func obtainUnidecodeWheel(ctx context.Context, client *http.Client, from string) ([]byte, error) {
	if from != "" {
		return readUnidecodeWheelFile(from)
	}
	wheel, err := unidecode.Fetch(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("unidecode install: %w; download the wheel from %s (sha256 %s) and retry with --from, or pass --from directly",
			err, unidecode.WheelURL, unidecode.WheelSHA256)
	}
	return wheel, nil
}

// readUnidecodeWheelFile reads path as an offline wheel, capped at
// unidecode.MaxWheelSize plus one byte so an oversized file refuses
// before this function ever hashes it.
func readUnidecodeWheelFile(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("unidecode install: --from %s: %w", path, err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("unidecode install: --from %s: is a directory, want a file", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("unidecode install: --from %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	data, err := io.ReadAll(io.LimitReader(f, unidecode.MaxWheelSize+1))
	if err != nil {
		return nil, fmt.Errorf("unidecode install: --from %s: %w", path, err)
	}
	if len(data) > unidecode.MaxWheelSize {
		return nil, unidecode.ErrWheelTooLarge
	}
	return data, nil
}

// checkUnidecodeOwner refuses when info's owning uid differs from
// geteuid(), naming the owner and the runuser fix, leaf.go's own
// checkOwner precedent.
func checkUnidecodeOwner(info fs.FileInfo, path string) error {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("unidecode install: %s: cannot determine its owner", path)
	}
	if int(st.Uid) == geteuid() {
		return nil
	}
	name := unidecodeOwnerName(st.Uid)
	return fmt.Errorf("unidecode install: %s owned by %s: run as that user: runuser -u %s -- dubplate-server %s",
		path, name, name, unidecodeInstallVerbName)
}

// checkUnidecodeLeafShape refuses when info, the leaf's own stat, is
// not a unidecodeLeafPerm directory: a regular file at path, or a
// directory at any other mode.
func checkUnidecodeLeafShape(info fs.FileInfo, path string) error {
	if !info.IsDir() {
		return fmt.Errorf("unidecode install: leaf %s: not a directory, want a directory: remove it and rerun dubplate-server %s",
			path, unidecodeInstallVerbName)
	}
	if mode := info.Mode().Perm(); mode != unidecodeLeafPerm {
		return fmt.Errorf("unidecode install: leaf %s: mode %04o, want %04o: fix its mode and rerun dubplate-server %s",
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
