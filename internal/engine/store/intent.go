package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// intentFormat is the one restore-intent record format this binary
// reads and writes. A record carrying any other number is refused
// rather than guessed at.
const intentFormat = 1

// intentFilePerm is the restore-intent record's mode, and the
// reconciliation report's: the report lists every library path.
const intentFilePerm = 0o600

// reportPrefix and reportSuffix frame a reconciliation report's name
// in the leaf, around the restore's stamp.
const (
	reportPrefix = "restore-report-"
	reportSuffix = ".txt"
)

// intentState is the flow step a restore-intent record was written
// after, which decides what finishes the restore.
type intentState string

const (
	// intentSwap is written before the swap renames any file.
	intentSwap intentState = "swap"

	// intentLibrary is written after the swap when the library half
	// is still owed.
	intentLibrary intentState = "library"

	// intentReconcile is written when only the reconciliation report
	// is owed on a store at the binary's version.
	intentReconcile intentState = "reconcile"

	// intentReconcilePending is written when the restored store is
	// below the binary's version, so only serve, which migrates it,
	// may finish the restore.
	intentReconcilePending intentState = "reconcile-pending"
)

// intentRecord is the restore-intent record in the leaf: what a
// restore has done and what it still owes. Every leaf file it names is
// a basename whose directory its field fixes.
type intentRecord struct {
	Format int         `json:"format"`
	State  intentState `json:"state"`

	// Stamp is the restore's own stamp, in FormatStamp's layout. It
	// names the report and every discarded file.
	Stamp string `json:"stamp"`

	Source intentSource `json:"source"`

	// Incoming is the staged store's basename in staging/.
	Incoming string `json:"incoming"`

	// PreRestore is the current store's copy's basename in
	// snapshots/, empty when no current store existed or when
	// DiscardCurrent moves the current files aside instead.
	PreRestore string `json:"pre_restore"`

	DiscardCurrent bool `json:"discard_current"`

	// Library is set when the library half is owed.
	Library *libraryOwed `json:"library"`

	// Version is the incoming store's user_version when it was
	// verified, kept for the report and refusal text.
	Version int `json:"version"`
}

// intentSource is where a restore read its snapshot: a stamp on a
// remote, or a local file.
type intentSource struct {
	Remote   string `json:"remote,omitempty"`
	Snapshot string `json:"snapshot,omitempty"`
	File     string `json:"file,omitempty"`
}

// libraryOwed is the library half of a restore.
type libraryOwed struct {
	// Root is the library root, resolved once, as an absolute path.
	Root         string `json:"root"`
	IntoNonempty bool   `json:"into_nonempty"`
}

// decodeIntent parses b as a restore-intent record, refusing an empty
// body, a key the record does not carry, anything after the one
// object, and every record validate refuses.
func decodeIntent(b []byte) (intentRecord, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return intentRecord{}, errors.New("the record is empty")
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	var r intentRecord
	if err := dec.Decode(&r); err != nil {
		return intentRecord{}, fmt.Errorf("parse: %w", err)
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return intentRecord{}, errors.New("parse: data after the record's object")
	}
	return r, r.validate()
}

// validate refuses a record this binary cannot act on: another
// format, an unknown state, a stamp FormatStamp did not write, a leaf
// name that is empty where it is required or that could leave its
// directory, a library state with no absolute root, and a record
// owing both a pre-restore copy and a discard.
func (r intentRecord) validate() error {
	if r.Format != intentFormat {
		return fmt.Errorf("format %d, want %d", r.Format, intentFormat)
	}
	switch r.State {
	case intentSwap, intentLibrary, intentReconcile, intentReconcilePending:
	default:
		return fmt.Errorf("unknown state %q", r.State)
	}
	if _, err := ParseStamp(r.Stamp); err != nil {
		return fmt.Errorf("field stamp: %w", err)
	}
	if r.Incoming == "" {
		return errors.New("field incoming is empty")
	}
	if err := checkBasename("incoming", r.Incoming); err != nil {
		return err
	}
	if err := checkBasename("pre_restore", r.PreRestore); err != nil {
		return err
	}
	if r.DiscardCurrent && r.PreRestore != "" {
		return fmt.Errorf("field pre_restore names %s beside discard_current", r.PreRestore)
	}
	if r.State == intentLibrary && r.Library == nil {
		return errors.New("field library is empty in state library")
	}
	if r.Library != nil && !filepath.IsAbs(r.Library.Root) {
		return fmt.Errorf("field library.root %q is not an absolute path", r.Library.Root)
	}
	return nil
}

// checkBasename refuses name, the value of field, when it holds a
// separator or is "." or "..", either of which would reach outside the
// directory the field fixes. An empty name passes: a caller that
// requires one checks for it.
func checkBasename(field, name string) error {
	if strings.ContainsRune(name, filepath.Separator) || name == "." || name == ".." {
		return fmt.Errorf("field %s %q is not a file name", field, name)
	}
	return nil
}

// readIntent returns l's restore-intent record, or nil when there is
// none. A record it cannot read or decode is refused, naming the path
// and the command that owns the record.
func (l leaf) readIntent() (*intentRecord, error) {
	b, err := os.ReadFile(l.restoreIntentPath)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, l.unreadableIntent(err)
	}
	r, err := decodeIntent(b)
	if err != nil {
		return nil, l.unreadableIntent(err)
	}
	return &r, nil
}

// unreadableIntent is the refusal of a record readIntent cannot use.
func (l leaf) unreadableIntent(err error) error {
	return refusal(StateRestoreInProgress, "restore-intent record %s is unreadable: %v: finish the restore with restore --resume", l.restoreIntentPath, err)
}

// checkIntent refuses when l holds a restore-intent record, in any
// state, naming what finishes the restore.
func (l leaf) checkIntent() error {
	r, err := l.readIntent()
	if err != nil || r == nil {
		return err
	}
	return r.refusal(l)
}

// refusal is how every mode but serve reports r: a reconcile-pending
// record names serve, and every other state names restore --resume.
func (r *intentRecord) refusal(l leaf) error {
	if r.State == intentReconcilePending {
		return refusal(StateRestoreInProgress, "restore-intent record %s (state %s): start serve to finish the restore", l.restoreIntentPath, r.State)
	}
	return refusal(StateRestoreInProgress, "restore-intent record %s (state %s): finish the restore with restore --resume", l.restoreIntentPath, r.State)
}

// writeIntent writes r as l's restore-intent record, replacing any
// record there, at intentFilePerm. It refuses a record decodeIntent
// would refuse, writing nothing.
func (l leaf) writeIntent(r intentRecord) error {
	if err := r.validate(); err != nil {
		return fmt.Errorf("write restore-intent record: %w", err)
	}
	return atomicfile.WriteStream(l.restoreIntentPath, intentFilePerm, func(w io.Writer) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return enc.Encode(r)
	})
}

// removeIntent removes l's restore-intent record and syncs the leaf,
// so the removal survives a power cut.
func (l leaf) removeIntent() error {
	if err := os.Remove(l.restoreIntentPath); err != nil {
		return fmt.Errorf("remove restore-intent record: %w", err)
	}
	return syncDir(l.root)
}

// reportPath is the reconciliation report's path for the restore
// stamped stamp.
func (l leaf) reportPath(stamp string) string {
	return filepath.Join(l.root, reportPrefix+stamp+reportSuffix)
}

// writeReport writes report as the reconciliation report of the
// restore stamped stamp, replacing one an earlier attempt wrote.
func (l leaf) writeReport(stamp string, report []byte) error {
	return atomicfile.WriteStream(l.reportPath(stamp), intentFilePerm, func(w io.Writer) error {
		_, err := w.Write(report)
		return err
	})
}
