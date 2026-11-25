package unidecode

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/glw907/dubplate/internal/doctor"
)

// checkInstalled reports whether the table at stateDir's unidecode leaf
// reads and verifies through [Parse]. It returns the read error
// unwrapped, so a missing file satisfies errors.Is(err, fs.ErrNotExist),
// and wraps a parse refusal with the path. It never touches the
// network and creates nothing.
func checkInstalled(stateDir string) error {
	path := tablePath(stateDir)
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if _, err := Parse(data); err != nil {
		return fmt.Errorf("%s: %w", path, err)
	}
	return nil
}

// DoctorRows renders the unidecode row for the state directory
// stateDir, which is "" when the environment does not set it.
func DoctorRows(stateDir string) []doctor.Row {
	row := doctor.Row{Name: "unidecode"}
	row.Status, row.Failed = rowStatus(stateDir)
	return []doctor.Row{row}
}

// rowStatus renders the unidecode row for the state directory
// stateDir: "not checked" when stateDir is empty (the wall row already
// fails), "ok (<version>)" for a table [Parse] accepts, "missing: ..."
// when the read fails because the state directory, the leaf, or the
// file is absent, and "invalid: ..." for every other failure (a
// directory in the file's place, permission denied, a checksum or
// format refusal).
func rowStatus(stateDir string) (status string, failed bool) {
	if stateDir == "" {
		return "not checked", false
	}
	err := checkInstalled(stateDir)
	if err == nil {
		return fmt.Sprintf("ok (%s)", Version), false
	}
	if errors.Is(err, fs.ErrNotExist) {
		return fmt.Sprintf("missing: run dubplate-server %s", installVerb), true
	}
	return fmt.Sprintf("invalid: %s: run dubplate-server %s", err, installVerb), true
}
