package shadow

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// openTempStore opens a Store rooted at a fresh t.TempDir() subdirectory,
// so a test never writes anywhere else on the filesystem.
func openTempStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenStore(filepath.Join(t.TempDir(), "shadow"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	return s
}

func TestStartClockWritesOnceAndNeverRewrites(t *testing.T) {
	s := openTempStore(t)
	first := time.Date(2026, 9, 10, 12, 0, 0, 0, time.UTC)
	wrote, err := s.StartClock(first)
	if err != nil || !wrote {
		t.Fatalf("StartClock = %v, %v; want true, nil", wrote, err)
	}
	later := first.Add(72 * time.Hour)
	wrote, err = s.StartClock(later)
	if err != nil {
		t.Fatalf("StartClock: %v", err)
	}
	if wrote {
		t.Error("StartClock reported a second write; the clock start is written once")
	}
	got, ok, err := s.ClockStart()
	if err != nil || !ok {
		t.Fatalf("ClockStart = %v, %v, %v", got, ok, err)
	}
	if !got.Equal(first) {
		t.Errorf("ClockStart = %v, want %v; a rewrite moves the launch date silently", got, first)
	}
}

func TestClockStartBeforeAnyStart(t *testing.T) {
	s := openTempStore(t)
	got, ok, err := s.ClockStart()
	if err != nil {
		t.Fatalf("ClockStart: %v", err)
	}
	if ok {
		t.Errorf("ClockStart = %v, true; want false before StartClock", got)
	}
}

func TestStoreRoundTripsEveryArtifact(t *testing.T) {
	s := openTempStore(t)

	obs := &Observation{
		ID:            "shadow-aaaaaaaaaaaaaaaa",
		Album:         "example uploader/example album",
		MatcherCommit: "deadbeef",
		EvalDate:      time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC),
		Capture: Capture{
			Uploader: "example uploader",
			Basename: "example album",
			Digest:   "aaaaaaaaaaaaaaaa",
			Items:    []matcher.ItemFields{{Artist: "Example Artist"}},
		},
		Class: ClassAgreement,
	}
	obs2 := &Observation{
		ID:       "shadow-bbbbbbbbbbbbbbbb",
		Album:    "example uploader/second album",
		Capture:  Capture{Digest: "bbbbbbbbbbbbbbbb"},
		Class:    ClassGoAppliesOracleSkipped,
		EvalDate: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC),
	}
	// Inserted out of id order, so a filesystem-order read would return
	// them in the wrong order and the lexical-sort assertion below would
	// fail if Observations trusted os.ReadDir instead.
	if err := s.PutObservation(obs2); err != nil {
		t.Fatalf("PutObservation(obs2): %v", err)
	}
	if err := s.PutObservation(obs); err != nil {
		t.Fatalf("PutObservation(obs): %v", err)
	}

	gotObs, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(gotObs) != 2 {
		t.Fatalf("Observations returned %d records, want 2", len(gotObs))
	}
	if gotObs[0].ID != obs.ID || gotObs[1].ID != obs2.ID {
		t.Errorf("Observations order = [%s, %s], want [%s, %s] (lexical id order)",
			gotObs[0].ID, gotObs[1].ID, obs.ID, obs2.ID)
	}
	if gotObs[0].Album != obs.Album || !gotObs[0].EvalDate.Equal(obs.EvalDate) ||
		gotObs[0].Capture.Items[0].Artist != obs.Capture.Items[0].Artist {
		t.Errorf("Observations()[0] = %+v, want a field-for-field match with %+v", gotObs[0], obs)
	}

	pass1 := &PassRecord{ID: "20260910T090000.000000000Z", Captured: 3, Divergences: map[string]int{ClassAgreement: 2}}
	pass2 := &PassRecord{ID: "20260910T080000.000000000Z", Captured: 1}
	if err := s.PutPass(pass1); err != nil {
		t.Fatalf("PutPass(pass1): %v", err)
	}
	if err := s.PutPass(pass2); err != nil {
		t.Fatalf("PutPass(pass2): %v", err)
	}
	gotPasses, err := s.Passes()
	if err != nil {
		t.Fatalf("Passes: %v", err)
	}
	if len(gotPasses) != 2 || gotPasses[0].ID != pass2.ID || gotPasses[1].ID != pass1.ID {
		t.Fatalf("Passes order = %v, want [%s, %s] (lexical id order)", gotPasses, pass2.ID, pass1.ID)
	}
	if gotPasses[1].Captured != pass1.Captured || gotPasses[1].Divergences[ClassAgreement] != 2 {
		t.Errorf("Passes()[1] = %+v, want a field-for-field match with %+v", gotPasses[1], pass1)
	}

	body := []byte(`{"ok":true}`)
	if err := s.PutResponse(obs.ID, 1, body); err != nil {
		t.Fatalf("PutResponse: %v", err)
	}
	gotBody, err := s.Response(obs.ID, "001.json")
	if err != nil {
		t.Fatalf("Response: %v", err)
	}
	if string(gotBody) != string(body) {
		t.Errorf("Response = %q, want %q", gotBody, body)
	}

	reports := []ConsumedReport{
		{File: "report-abc.json", StartedAt: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), EndedAt: time.Date(2026, 9, 1, 1, 0, 0, 0, time.UTC)},
	}
	if err := s.PutConsumedReports(reports); err != nil {
		t.Fatalf("PutConsumedReports: %v", err)
	}
	gotReports, err := s.ConsumedReports()
	if err != nil {
		t.Fatalf("ConsumedReports: %v", err)
	}
	if len(gotReports) != 1 || gotReports[0].File != reports[0].File || !gotReports[0].StartedAt.Equal(reports[0].StartedAt) {
		t.Errorf("ConsumedReports = %+v, want %+v", gotReports, reports)
	}

	pendingData := []byte(`{"stage":"pending"}`)
	if err := s.PutPending(obs.ID, pendingData); err != nil {
		t.Fatalf("PutPending: %v", err)
	}
	ids, err := s.PendingIDs()
	if err != nil {
		t.Fatalf("PendingIDs: %v", err)
	}
	if len(ids) != 1 || ids[0] != obs.ID {
		t.Fatalf("PendingIDs = %v, want [%s]", ids, obs.ID)
	}
	if err := s.RemovePending(obs.ID); err != nil {
		t.Fatalf("RemovePending: %v", err)
	}
	ids, err = s.PendingIDs()
	if err != nil {
		t.Fatalf("PendingIDs after remove: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("PendingIDs after RemovePending = %v, want empty", ids)
	}

	resolvedData := []byte(`{"stage":"resolved"}`)
	if err := s.PutResolved(obs.ID, resolvedData); err != nil {
		t.Fatalf("PutResolved: %v", err)
	}
	resolvedIDs, err := s.ResolvedIDs()
	if err != nil {
		t.Fatalf("ResolvedIDs: %v", err)
	}
	if len(resolvedIDs) != 1 || resolvedIDs[0] != obs.ID {
		t.Fatalf("ResolvedIDs = %v, want [%s]", resolvedIDs, obs.ID)
	}
}

func TestStoreWritesNothingOutsideItsRoot(t *testing.T) {
	tmp := t.TempDir()
	root := filepath.Join(tmp, "shadow")
	s, err := OpenStore(root)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	if _, err := s.StartClock(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("StartClock: %v", err)
	}
	if err := s.PutObservation(&Observation{ID: "shadow-cccccccccccccccc"}); err != nil {
		t.Fatalf("PutObservation: %v", err)
	}
	if err := s.PutResponse("shadow-cccccccccccccccc", 1, []byte("{}")); err != nil {
		t.Fatalf("PutResponse: %v", err)
	}
	if err := s.PutCaseFile("shadow-cccccccccccccccc", []byte("{}")); err != nil {
		t.Fatalf("PutCaseFile: %v", err)
	}
	if err := s.PutPending("shadow-cccccccccccccccc", []byte("{}")); err != nil {
		t.Fatalf("PutPending: %v", err)
	}
	if err := s.PutResolved("shadow-cccccccccccccccc", []byte("{}")); err != nil {
		t.Fatalf("PutResolved: %v", err)
	}
	if err := s.PutPass(&PassRecord{ID: "20260901T000000.000000000Z"}); err != nil {
		t.Fatalf("PutPass: %v", err)
	}
	if err := s.PutConsumedReports([]ConsumedReport{{File: "report-abc.json"}}); err != nil {
		t.Fatalf("PutConsumedReports: %v", err)
	}

	// The attacker-shaped id: a caller that ever built an id with ".."
	// must be refused rather than silently escaping root.
	if err := s.PutObservation(&Observation{ID: "../../escaped"}); err == nil {
		t.Fatal("PutObservation with a \"..\"-bearing id: got nil error, want an error")
	}
	if s.HasObservation("../../escaped") {
		t.Error("HasObservation(\"../../escaped\") = true, want false")
	}

	err = filepath.WalkDir(tmp, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil || strings.HasPrefix(rel, "..") {
			t.Errorf("wrote outside store root: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}

func TestStoreLeavesNoTempFile(t *testing.T) {
	s := openTempStore(t)

	if _, err := s.StartClock(time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)); err != nil {
		t.Fatalf("StartClock: %v", err)
	}
	if err := s.PutObservation(&Observation{ID: "shadow-dddddddddddddddd"}); err != nil {
		t.Fatalf("PutObservation: %v", err)
	}
	if err := s.PutResponse("shadow-dddddddddddddddd", 1, []byte("{}")); err != nil {
		t.Fatalf("PutResponse: %v", err)
	}
	if err := s.PutCaseFile("shadow-dddddddddddddddd", []byte("{}")); err != nil {
		t.Fatalf("PutCaseFile: %v", err)
	}
	if err := s.PutPending("shadow-dddddddddddddddd", []byte("{}")); err != nil {
		t.Fatalf("PutPending: %v", err)
	}
	if err := s.PutResolved("shadow-dddddddddddddddd", []byte("{}")); err != nil {
		t.Fatalf("PutResolved: %v", err)
	}
	if err := s.PutPass(&PassRecord{ID: "20260901T000000.000000000Z"}); err != nil {
		t.Fatalf("PutPass: %v", err)
	}
	if err := s.PutConsumedReports([]ConsumedReport{{File: "report-abc.json"}}); err != nil {
		t.Fatalf("PutConsumedReports: %v", err)
	}

	err := filepath.WalkDir(s.root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.Contains(d.Name(), ".tmp-") {
			t.Errorf("leaked temp file: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
}

func TestHasObservationIsTheIdempotenceCheck(t *testing.T) {
	s := openTempStore(t)
	const id = "shadow-eeeeeeeeeeeeeeee"
	if s.HasObservation(id) {
		t.Errorf("HasObservation(%q) = true before PutObservation, want false", id)
	}
	if err := s.PutObservation(&Observation{ID: id}); err != nil {
		t.Fatalf("PutObservation: %v", err)
	}
	if !s.HasObservation(id) {
		t.Errorf("HasObservation(%q) = false after PutObservation, want true", id)
	}
}

func TestPassIDCollisionSecondWriteWins(t *testing.T) {
	// The pass id derives from the now parameter given to a caller
	// outside this package; the store itself only proves that two
	// PutPass calls sharing one id collide with the second write
	// winning, rather than the store inventing a disambiguator.
	s := openTempStore(t)
	const id = "20260901T000000.000000000Z"
	if err := s.PutPass(&PassRecord{ID: id, Captured: 1}); err != nil {
		t.Fatalf("PutPass(first): %v", err)
	}
	if err := s.PutPass(&PassRecord{ID: id, Captured: 2}); err != nil {
		t.Fatalf("PutPass(second): %v", err)
	}
	passes, err := s.Passes()
	if err != nil {
		t.Fatalf("Passes: %v", err)
	}
	if len(passes) != 1 {
		t.Fatalf("Passes returned %d records, want 1 (the second write wins)", len(passes))
	}
	if passes[0].Captured != 2 {
		t.Errorf("Passes()[0].Captured = %d, want 2 (the second write wins)", passes[0].Captured)
	}
}
