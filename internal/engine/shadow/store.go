package shadow

import (
	"cmp"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// A Store is the shadow's durable state under one directory, rooted at
// <state dir>/shadow. It is the unit's only writable path, so every
// method here writes inside root and nowhere else, and every write goes
// through atomicfile.WriteFile: a reader never sees a partial file, and
// the observe pass and a hand-run report never contend.
type Store struct {
	root string
}

const (
	observationsDir = "observations"
	passesDir       = "passes"
	pendingDir      = "pending"
	resolvedDir     = "resolved"

	observationFile = "observation.json"
	caseFile        = "case.json"
	responsesDir    = "responses"
	startedAtFile   = "started-at"
	consumedReports = "consumed-reports.json"
)

// OpenStore creates root and its four subdirectories if they do not
// exist and returns a Store over them. Directories are 0755 and files
// 0644, per go-conventions.
func OpenStore(root string) (*Store, error) {
	for _, sub := range []string{observationsDir, passesDir, pendingDir, resolvedDir} {
		if err := os.MkdirAll(filepath.Join(root, sub), 0o755); err != nil {
			return nil, fmt.Errorf("open store: create %s: %w", sub, err)
		}
	}
	return &Store{root: root}, nil
}

// validID reports whether id is safe to use as a single path segment
// under root. Ids are digest-derived (Resolution R5), so anything else
// is a bug rather than input, and this check makes it a loud one rather
// than a path traversal.
func validID(id string) bool {
	if id == "" || id == "." || id == ".." {
		return false
	}
	return !strings.ContainsAny(id, `/\`)
}

// ClockStart reports the recorded clock start and whether one exists.
func (s *Store) ClockStart() (time.Time, bool, error) {
	data, err := os.ReadFile(filepath.Join(s.root, startedAtFile))
	if os.IsNotExist(err) {
		return time.Time{}, false, nil
	}
	if err != nil {
		return time.Time{}, false, fmt.Errorf("read clock start: %w", err)
	}
	t, err := time.Parse(time.RFC3339Nano, strings.TrimSpace(string(data)))
	if err != nil {
		return time.Time{}, false, fmt.Errorf("parse clock start: %w", err)
	}
	return t, true, nil
}

// StartClock writes started-at if and only if no clock start exists.
// It reports whether this call wrote it. The clock start is written
// once and never rewritten: launch date is this date plus the floor, so
// a rewrite would silently move the launch date.
func (s *Store) StartClock(now time.Time) (bool, error) {
	_, ok, err := s.ClockStart()
	if err != nil {
		return false, err
	}
	if ok {
		return false, nil
	}
	data := []byte(now.Format(time.RFC3339Nano) + "\n")
	if err := atomicfile.WriteFile(filepath.Join(s.root, startedAtFile), data, 0o644); err != nil {
		return false, fmt.Errorf("write clock start: %w", err)
	}
	return true, nil
}

// PutObservation writes obs to observations/<obs.ID>/observation.json.
func (s *Store) PutObservation(obs *Observation) error {
	if !validID(obs.ID) {
		return fmt.Errorf("put observation: invalid id %q", obs.ID)
	}
	dir := filepath.Join(s.root, observationsDir, obs.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("put observation %s: %w", obs.ID, err)
	}
	data, err := json.MarshalIndent(obs, "", "  ")
	if err != nil {
		return fmt.Errorf("put observation %s: %w", obs.ID, err)
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, observationFile), data, 0o644); err != nil {
		return fmt.Errorf("put observation %s: %w", obs.ID, err)
	}
	return nil
}

// Observations returns every stored observation, in lexical id order.
// That is the STORE's read order and not the pass's match order: an id
// is a digest prefix (Resolution R5), so id order says nothing about
// capture order, and Run matches in (uploader, basename) capture order
// instead (Resolution R18).
func (s *Store) Observations() ([]*Observation, error) {
	root := filepath.Join(s.root, observationsDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("list observations: %w", err)
	}
	var obs []*Observation
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, e.Name(), observationFile))
		if err != nil {
			return nil, fmt.Errorf("read observation %s: %w", e.Name(), err)
		}
		var o Observation
		if err := json.Unmarshal(data, &o); err != nil {
			return nil, fmt.Errorf("decode observation %s: %w", e.Name(), err)
		}
		obs = append(obs, &o)
	}
	slices.SortFunc(obs, func(a, b *Observation) int { return cmp.Compare(a.ID, b.ID) })
	return obs, nil
}

// HasObservation reports whether an observation directory exists for id,
// which is the capture's idempotence check (contract S7).
func (s *Store) HasObservation(id string) bool {
	if !validID(id) {
		return false
	}
	_, err := os.Stat(filepath.Join(s.root, observationsDir, id, observationFile))
	return err == nil
}

// PutResponse writes one recorded body to
// observations/<obsID>/responses/<n>.json, named %03d from 1, matching
// the corpus's own case-directory convention.
func (s *Store) PutResponse(obsID string, n int, body []byte) error {
	if !validID(obsID) {
		return fmt.Errorf("put response: invalid observation id %q", obsID)
	}
	dir := filepath.Join(s.root, observationsDir, obsID, responsesDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("put response %s/%03d: %w", obsID, n, err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%03d.json", n))
	if err := atomicfile.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("put response %s/%03d: %w", obsID, n, err)
	}
	return nil
}

// Response reads one recorded body back, for the re-score and for
// reevaluate.
func (s *Store) Response(obsID, name string) ([]byte, error) {
	if !validID(obsID) {
		return nil, fmt.Errorf("response: invalid observation id %q", obsID)
	}
	name = filepath.Base(name)
	data, err := os.ReadFile(filepath.Join(s.root, observationsDir, obsID, responsesDir, name))
	if err != nil {
		return nil, fmt.Errorf("response %s/%s: %w", obsID, name, err)
	}
	return data, nil
}

// PutCaseFile writes the marshalled cassette-shaped fixture.
// internal/engine/shadow may import no harness package (contract S19),
// so it takes bytes cmd/dubplate-server produced rather than a
// cassette.Case.
func (s *Store) PutCaseFile(obsID string, data []byte) error {
	if !validID(obsID) {
		return fmt.Errorf("put case file: invalid observation id %q", obsID)
	}
	dir := filepath.Join(s.root, observationsDir, obsID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("put case file %s: %w", obsID, err)
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, caseFile), data, 0o644); err != nil {
		return fmt.Errorf("put case file %s: %w", obsID, err)
	}
	return nil
}

// PutPending, RemovePending and PutResolved manage the two staged-row
// directories. PutPending takes marshalled bytes for the same reason
// PutCaseFile does.
func (s *Store) PutPending(obsID string, data []byte) error {
	return s.putStagedRow(pendingDir, obsID, data)
}

func (s *Store) RemovePending(obsID string) error {
	if !validID(obsID) {
		return fmt.Errorf("remove pending: invalid observation id %q", obsID)
	}
	err := os.Remove(filepath.Join(s.root, pendingDir, obsID+".json"))
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove pending %s: %w", obsID, err)
	}
	return nil
}

func (s *Store) PutResolved(obsID string, data []byte) error {
	return s.putStagedRow(resolvedDir, obsID, data)
}

func (s *Store) putStagedRow(sub, obsID string, data []byte) error {
	if !validID(obsID) {
		return fmt.Errorf("put %s row: invalid observation id %q", sub, obsID)
	}
	path := filepath.Join(s.root, sub, obsID+".json")
	if err := atomicfile.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("put %s row %s: %w", sub, obsID, err)
	}
	return nil
}

// PendingIDs returns every staged pending observation id.
func (s *Store) PendingIDs() ([]string, error) {
	return s.stagedRowIDs(pendingDir)
}

// ResolvedIDs returns every staged resolved observation id.
func (s *Store) ResolvedIDs() ([]string, error) {
	return s.stagedRowIDs(resolvedDir)
}

// Resolved reads back resolved/<obsID>.json, mirroring the pending and
// observation readers: a caller that only checked ResolvedIDs never saw
// what Reevaluate actually recorded there.
func (s *Store) Resolved(obsID string) (*ResolvedRecord, error) {
	if !validID(obsID) {
		return nil, fmt.Errorf("resolved: invalid observation id %q", obsID)
	}
	data, err := os.ReadFile(filepath.Join(s.root, resolvedDir, obsID+".json"))
	if err != nil {
		return nil, fmt.Errorf("resolved %s: %w", obsID, err)
	}
	var rec ResolvedRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return nil, fmt.Errorf("decode resolved %s: %w", obsID, err)
	}
	return &rec, nil
}

func (s *Store) stagedRowIDs(sub string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(s.root, sub))
	if err != nil {
		return nil, fmt.Errorf("list %s: %w", sub, err)
	}
	var ids []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ids = append(ids, strings.TrimSuffix(e.Name(), ".json"))
	}
	slices.Sort(ids)
	return ids, nil
}

// PutPass writes one pass record.
func (s *Store) PutPass(p *PassRecord) error {
	if !validID(p.ID) {
		return fmt.Errorf("put pass: invalid id %q", p.ID)
	}
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("put pass %s: %w", p.ID, err)
	}
	path := filepath.Join(s.root, passesDir, p.ID+".json")
	if err := atomicfile.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("put pass %s: %w", p.ID, err)
	}
	return nil
}

// Passes returns every pass record, in lexical id order.
func (s *Store) Passes() ([]*PassRecord, error) {
	root := filepath.Join(s.root, passesDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("list passes: %w", err)
	}
	var passes []*PassRecord
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read pass %s: %w", e.Name(), err)
		}
		var p PassRecord
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("decode pass %s: %w", e.Name(), err)
		}
		passes = append(passes, &p)
	}
	slices.SortFunc(passes, func(a, b *PassRecord) int { return cmp.Compare(a.ID, b.ID) })
	return passes, nil
}

// ConsumedReports and PutConsumedReports read and replace the ledger
// that makes a swept report a named condition rather than silence
// (contract S12).
func (s *Store) ConsumedReports() ([]ConsumedReport, error) {
	data, err := os.ReadFile(filepath.Join(s.root, consumedReports))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read consumed reports: %w", err)
	}
	var rs []ConsumedReport
	if err := json.Unmarshal(data, &rs); err != nil {
		return nil, fmt.Errorf("decode consumed reports: %w", err)
	}
	return rs, nil
}

func (s *Store) PutConsumedReports(rs []ConsumedReport) error {
	data, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return fmt.Errorf("put consumed reports: %w", err)
	}
	if err := atomicfile.WriteFile(filepath.Join(s.root, consumedReports), data, 0o644); err != nil {
		return fmt.Errorf("put consumed reports: %w", err)
	}
	return nil
}

// Size returns the store's own footprint in bytes, the report's sizing
// row (rung 4 spec, risk 5).
func (s *Store) Size() (int64, error) {
	var total int64
	err := filepath.WalkDir(s.root, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		total += info.Size()
		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("size store: %w", err)
	}
	return total, nil
}
