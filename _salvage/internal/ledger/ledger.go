// Package ledger implements the estate's contributions ledger: an
// append-only JSONL history of every session's state transitions,
// written by both serve and the importer, that lets a contributor see
// their uploads after beets renames the files that would otherwise be
// their only trace.
package ledger

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

// State is the one journey-state vocabulary shared end to end: the
// ledger, internal/journey, and the web store all compare against
// these values rather than a locally redefined set.
//
// Each state has exactly one sanctioned appender, so a history read and
// a journey read can never disagree about who last touched a session:
// the importer appends [StateTagging] (carrying the beets-resolved
// LibraryPaths), [StateReview], and [StateRejected]; serve appends
// every upload-side state ([StateUploading], [StateUploaded],
// [StateChecking], [StateQueued]), [StateStalled], and
// [StateInLibrary] once it observes the Navidrome scan complete.
type State string

// The State vocabulary. ValidState is the exhaustive membership test;
// Append rejects any other value.
const (
	StateUploading State = "uploading"
	StateUploaded  State = "uploaded"
	StateChecking  State = "checking"
	StateTagging   State = "tagging"
	StateInLibrary State = "in-library"
	StateReview    State = "review"
	StateRejected  State = "rejected"
	StateQueued    State = "queued"
	StateStalled   State = "stalled"
)

var validStates = map[State]bool{
	StateUploading: true,
	StateUploaded:  true,
	StateChecking:  true,
	StateTagging:   true,
	StateInLibrary: true,
	StateReview:    true,
	StateRejected:  true,
	StateQueued:    true,
	StateStalled:   true,
}

// ValidState reports whether s is a member of the [State] vocabulary.
func ValidState(s State) bool {
	return validStates[s]
}

// Event is one contributions-ledger record: a state transition for a
// session, keyed by the contributor's verified email.
type Event struct {
	SessionID    string    `json:"session_id"`
	Email        string    `json:"email"`
	Album        string    `json:"album"`
	Collection   string    `json:"collection,omitempty"`
	State        State     `json:"state"`
	At           time.Time `json:"at"`
	Detail       string    `json:"detail,omitempty"`
	LibraryPaths []string  `json:"library_paths,omitempty"`
}

// ErrInvalidState is the sentinel [Append] wraps when an Event's State
// fails [ValidState].
var ErrInvalidState = errors.New("ledger: invalid state")

// maxRecordBytes is the JSON-line size Append degrades records to fit,
// never by rejecting or splitting one: see encodeRecord.
const maxRecordBytes = 4096

// maxScanTokenBytes bounds how large a single line readAll will parse,
// well above maxRecordBytes as a defensive margin against a ledger file
// from a future, larger record format.
const maxScanTokenBytes = 1 << 20

const ledgerFilePerm = 0o644

// Ledger is the contributions history at a JSONL file path. A Ledger
// holds no open file handle and no in-process lock: Append, History,
// and Session each open path fresh, which is what lets serve and the
// importer act as independent, uncoordinated appenders against the
// same file.
type Ledger struct {
	path string
}

// Open returns a Ledger backed by path. Open does no I/O: the file is
// created lazily by the first [Ledger.Append], and reading a path with
// no file yet is an empty result, not an error.
func Open(path string) (*Ledger, error) {
	if path == "" {
		return nil, errors.New("ledger: empty path")
	}
	return &Ledger{path: path}, nil
}

// Append writes e as one JSONL record. It opens path fresh with
// O_APPEND so concurrent appenders never interleave: each record
// reaches the file through a single write(2) call, which POSIX
// guarantees an O_APPEND file descriptor lands whole or not at all.
// Append rejects e.State outside the [ValidState] vocabulary; every
// other Event degrades to fit rather than being rejected (see
// encodeRecord).
func (l *Ledger) Append(e Event) error {
	if !ValidState(e.State) {
		return fmt.Errorf("%w: %q", ErrInvalidState, e.State)
	}

	line, err := encodeRecord(e)
	if err != nil {
		return fmt.Errorf("encode ledger record: %w", err)
	}

	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, ledgerFilePerm)
	if err != nil {
		return fmt.Errorf("open ledger %s: %w", l.path, err)
	}
	defer func() { _ = f.Close() }()

	if err := appendLine(f, line); err != nil {
		return fmt.Errorf("append ledger %s: %w", l.path, err)
	}
	return nil
}

// syncWriter is the write surface appendLine needs. *os.File satisfies
// it in production; tests substitute a fake to assert the
// one-write(2)-call guarantee directly.
type syncWriter interface {
	io.Writer
	Sync() error
}

// appendLine writes line to w in a single Write call, then fsyncs
// before returning. One Write call is the interleave guarantee two
// processes appending to the same O_APPEND file depend on: splitting a
// record across two Write calls would let another appender's record
// land between them and corrupt both.
func appendLine(w syncWriter, line []byte) error {
	n, err := w.Write(line)
	if err != nil {
		return err
	}
	if n != len(line) {
		return fmt.Errorf("short write: wrote %d of %d bytes", n, len(line))
	}
	return w.Sync()
}

// encodeRecord marshals e to a newline-terminated JSON line, degrading
// an oversized record deterministically instead of splitting or
// rejecting it: LibraryPaths collapses to the album's common directory
// plus a track count first, then Detail truncates. A record still over
// maxRecordBytes after both steps is returned anyway: a terminal event
// is never lost to a size limit.
func encodeRecord(e Event) ([]byte, error) {
	line, err := marshalRecord(e)
	if err != nil {
		return nil, err
	}
	if len(line) <= maxRecordBytes {
		return line, nil
	}

	degraded := e
	degraded.LibraryPaths = collapseLibraryPaths(e.LibraryPaths)
	line, err = marshalRecord(degraded)
	if err != nil {
		return nil, err
	}

	for len(line) > maxRecordBytes && degraded.Detail != "" {
		degraded.Detail = halveRunes(degraded.Detail)
		line, err = marshalRecord(degraded)
		if err != nil {
			return nil, err
		}
	}
	return line, nil
}

func marshalRecord(e Event) ([]byte, error) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}

// collapseLibraryPaths reduces a full track listing to the album's
// common directory and a track count, the field most likely to blow
// the size ceiling on a multi-disc release.
func collapseLibraryPaths(paths []string) []string {
	if len(paths) == 0 {
		return paths
	}
	return []string{albumRoot(paths), strconv.Itoa(len(paths))}
}

// albumRoot returns the deepest directory common to every path. For a
// multi-disc release whose tracks sit under per-disc subdirectories,
// that is the album directory one level above the discs.
func albumRoot(paths []string) string {
	if len(paths) == 1 {
		return filepath.Dir(paths[0])
	}
	common := strings.Split(filepath.Clean(paths[0]), string(filepath.Separator))
	for _, p := range paths[1:] {
		common = sharedPrefix(common, strings.Split(filepath.Clean(p), string(filepath.Separator)))
	}
	root := strings.Join(common, string(filepath.Separator))
	if root == "" {
		return string(filepath.Separator)
	}
	return root
}

func sharedPrefix(a, b []string) []string {
	n := min(len(a), len(b))
	i := 0
	for i < n && a[i] == b[i] {
		i++
	}
	return a[:i]
}

// halveRunes drops the back half of s by rune count, converging to ""
// within log2(len(s)) calls so encodeRecord's truncation loop always
// terminates.
func halveRunes(s string) string {
	r := []rune(s)
	if len(r) <= 1 {
		return ""
	}
	return string(r[:len(r)/2])
}

// History returns every event recorded for email, newest first. A
// ledger path with no file yet returns an empty slice and a nil
// error.
func (l *Ledger) History(email string) ([]Event, error) {
	events, err := l.readAll()
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, e := range slices.Backward(events) {
		if e.Email == email {
			out = append(out, e)
		}
	}
	return out, nil
}

// Session returns every event recorded for sessionID, newest first. A
// ledger path with no file yet returns an empty slice and a nil
// error.
func (l *Ledger) Session(sessionID string) ([]Event, error) {
	events, err := l.readAll()
	if err != nil {
		return nil, err
	}
	var out []Event
	for _, e := range slices.Backward(events) {
		if e.SessionID == sessionID {
			out = append(out, e)
		}
	}
	return out, nil
}

// readAll parses every well-formed line in the ledger file, in file
// order (oldest first). A line that fails to parse as an Event,
// normally only the final line, torn by a crash mid-write, is logged
// as a warning and skipped rather than aborting the read: a
// contributor's earlier history must survive a crash that tore only
// the most recent record.
func (l *Ledger) readAll() ([]Event, error) {
	f, err := os.Open(l.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("open ledger %s: %w", l.path, err)
	}
	defer func() { _ = f.Close() }()

	var events []Event
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenBytes)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		var e Event
		if err := json.Unmarshal(line, &e); err != nil {
			slog.Warn("ledger: skip unparseable record", "path", l.path, "line", lineNo, "err", err)
			continue
		}
		events = append(events, e)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read ledger %s: %w", l.path, err)
	}
	return events, nil
}
