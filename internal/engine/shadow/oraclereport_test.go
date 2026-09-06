package shadow

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"
)

// tagLogSeq gives every writeTagLog call its own file name without
// reading the clock: internal/engine/shadow's own tests are inside the
// architectural walk that forbids time.Now (Global Constraints), so a
// timestamp-named log file is not available here the way the
// incumbent names one at beets.go:104.
var tagLogSeq atomic.Int64

// writeRunReport writes one run report to stateDir under the
// incumbent's own naming rule, "report-<runid>.json".
func writeRunReport(t *testing.T, stateDir string, r oracleRunReport) {
	t.Helper()

	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		t.Fatalf("marshal run report: %v", err)
	}
	path := filepath.Join(stateDir, reportFileName(r.RunID))
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write run report %s: %v", path, err)
	}
}

// writeTagLog writes one tag log carrying a "skip " line for every
// entry in skips, joined the way beets joins several paths on one line
// ("; " separated), one line per call so a test can also exercise
// several tag logs across several calls.
func writeTagLog(t *testing.T, stateDir string, skips []string) {
	t.Helper()

	var data []byte
	if len(skips) > 0 {
		line := "skip " + joinSemicolon(skips) + "\n"
		data = []byte(line)
	}
	name := fmt.Sprintf("import-%d.log", tagLogSeq.Add(1))
	path := filepath.Join(stateDir, name)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write tag log %s: %v", path, err)
	}
}

func joinSemicolon(parts []string) string {
	out := parts[0]
	for _, p := range parts[1:] {
		out += "; " + p
	}
	return out
}

// writeReviewDir creates <reviewDir>/<date>/<uploader>/<name> as an
// empty directory, the shape a filed skip or quarantine leaves behind.
func writeReviewDir(t *testing.T, reviewDir, date, uploader, name string) {
	t.Helper()

	dir := filepath.Join(reviewDir, date, uploader, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir review dir %s: %v", dir, err)
	}
}

func TestReadRunReportsSortsByStartedAt(t *testing.T) {
	dir := t.TempDir()
	writeRunReport(t, dir, oracleRunReport{Version: 2, RunID: "bbb", StartedAt: passNow.Add(time.Hour), EndedAt: passNow.Add(time.Hour + time.Minute)})
	writeRunReport(t, dir, oracleRunReport{Version: 2, RunID: "aaa", StartedAt: passNow, EndedAt: passNow.Add(time.Minute)})

	got, err := readRunReports(dir)
	if err != nil {
		t.Fatalf("readRunReports: %v", err)
	}
	if len(got) != 2 || got[0].RunID != "aaa" || got[1].RunID != "bbb" {
		t.Fatalf("reports = %+v, want [aaa bbb] by StartedAt", got)
	}
}

// TestReadRunReportsToleratesAnUnknownField proves the decoder does
// not set DisallowUnknownFields: the incumbent may add a field, and a
// shadow that refused to read a newer report would go blind rather
// than degrade.
func TestReadRunReportsToleratesAnUnknownField(t *testing.T) {
	dir := t.TempDir()
	raw := `{"version":2,"run_id":"ccc","started_at":"2026-09-01T00:00:00Z","ended_at":"2026-09-01T00:05:00Z","albums":[],"a_field_this_pass_has_never_seen":"value"}`
	if err := os.WriteFile(filepath.Join(dir, "report-ccc.json"), []byte(raw), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}

	got, err := readRunReports(dir)
	if err != nil {
		t.Fatalf("readRunReports: %v", err)
	}
	if len(got) != 1 || got[0].RunID != "ccc" {
		t.Fatalf("reports = %+v, want one report ccc", got)
	}
}

func TestReadTagLogSkipsExtractsBasenames(t *testing.T) {
	dir := t.TempDir()
	writeTagLog(t, dir, []string{"/srv/staging/ada/windrose", "/srv/staging/ada/second"})

	got, err := readTagLogSkips(dir)
	if err != nil {
		t.Fatalf("readTagLogSkips: %v", err)
	}
	if !got["windrose"] || !got["second"] {
		t.Errorf("skips = %v, want windrose and second", got)
	}
}

func TestReviewedBasenamesReadsEveryDatedDirectory(t *testing.T) {
	dir := t.TempDir()
	writeReviewDir(t, dir, "2026-09-01", "ada", "windrose")
	writeReviewDir(t, dir, "2026-09-02", "ada", "second-142530")

	got, err := reviewedBasenames(dir)
	if err != nil {
		t.Fatalf("reviewedBasenames: %v", err)
	}
	if len(got["ada"]) != 2 {
		t.Fatalf("reviewed[ada] = %v, want 2 entries", got["ada"])
	}
}

// TestReconcileNeverReadsTheImportLock proves risk 3's whole mechanism
// stays closed: readRunReportsFS and readTagLogSkipsFS are given a
// recording fs.FS wrapping a real directory that also carries
// import.lock, and the recorded path list must never contain it.
func TestReconcileNeverReadsTheImportLock(t *testing.T) {
	dir := t.TempDir()
	writeRunReport(t, dir, oracleRunReport{Version: 2, RunID: "aaa", StartedAt: passNow, EndedAt: passNow.Add(time.Minute)})
	writeTagLog(t, dir, []string{"/srv/staging/ada/windrose"})
	if err := os.WriteFile(filepath.Join(dir, importLockName), nil, 0o644); err != nil {
		t.Fatalf("write import.lock: %v", err)
	}

	rec := &recordingFS{fsys: os.DirFS(dir)}
	if _, err := readRunReportsFS(rec); err != nil {
		t.Fatalf("readRunReportsFS: %v", err)
	}
	if _, err := readTagLogSkipsFS(rec); err != nil {
		t.Fatalf("readTagLogSkipsFS: %v", err)
	}

	for _, opened := range rec.opened {
		if opened == importLockName {
			t.Fatalf("opened paths = %v, must never contain %s", rec.opened, importLockName)
		}
	}
}

// recordingFS wraps an fs.FS and records every name Open is called
// with, the "recording os.DirFS-shaped seam" the reconcile tests use to
// prove the import lock is opened nowhere.
type recordingFS struct {
	fsys   fs.FS
	opened []string
}

func (r *recordingFS) Open(name string) (fs.File, error) {
	r.opened = append(r.opened, name)
	return r.fsys.Open(name)
}
