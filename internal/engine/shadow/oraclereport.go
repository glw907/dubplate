package shadow

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
)

// importLockName is the incumbent's lock file. It is named here so the
// reader functions can refuse it by name: flock on a read-only
// descriptor succeeds on Linux, and the incumbent treats a held lock as
// healthy, writing a LockHeld report, pinging SUCCESS and returning nil
// without importing (_salvage/internal/importer/importer.go:397-413).
// A shadow holding it would stall the box's own import while every
// signal read green, so this package opens it never: readRunReports and
// readTagLogSkips glob "report-*.json" and "import-*.log" by name, and
// neither ever lists or opens importLockName.
const importLockName = "import.lock"

// The incumbent's run-report shape, decoded rather than imported.
// _salvage/ is outside the module and invisible to the go tool, so
// these declarations mirror _salvage/internal/importer/importer.go:134-172
// field for field and tag for tag. They retire at rung 16 with the
// incumbent binary; nothing else in the greenfield server reads them.
//
// The decoder deliberately does NOT set DisallowUnknownFields: the
// incumbent may add a field, and a shadow that refused to read a newer
// report would go blind rather than degrade.
type oracleRunReport struct {
	Version   int                  `json:"version"`
	RunID     string               `json:"run_id"`
	StartedAt time.Time            `json:"started_at"`
	EndedAt   time.Time            `json:"ended_at"`
	LockHeld  bool                 `json:"lock_held"`
	Error     string               `json:"error,omitempty"`
	Albums    []oracleAlbumOutcome `json:"albums"`
}

// oracleAlbumOutcome carries no per-album path, which is the whole basis
// of the basename join (contract S10): SourceDir is the UPLOADER's inbox
// directory (the field at importer.go:136, documented on
// albumUnit.sourceDir at :212, assigned at the uploader level at :766,
// :908 and :1066) and several rows of one report routinely share it.
type oracleAlbumOutcome struct {
	Uploader     string   `json:"uploader"`
	SourceDir    string   `json:"source_dir"`
	Outcome      string   `json:"outcome"`
	SessionID    string   `json:"session_id,omitempty"`
	LibraryPaths []string `json:"library_paths,omitempty"`
	Collection   string   `json:"collection,omitempty"`
	Error        string   `json:"error,omitempty"`
}

// The outcome vocabulary, importer.go:107-125. imported is an apply and
// review is a skip; quarantined and rejected are GATE outcomes that
// never reached beets (importer.go:945, :499-505), so an observation
// carrying one is excluded rather than counted.
const (
	outcomeImported    = "imported"
	outcomeReview      = "review"
	outcomeQuarantined = "quarantined"
	outcomeRejected    = "rejected"
)

// reportFileName is the incumbent's own naming rule
// (reportFilePrefix + RunID + reportFileSuffix, importer.go:76-78,
// :509): reconcile uses it to test a report already named in
// ConsumedReport.File without keeping a second index.
func reportFileName(runID string) string {
	return "report-" + runID + ".json"
}

// readRunReports reads every report-*.json under stateDir, sorted
// ascending by StartedAt (oldest first). It opens no other path under
// stateDir, and in particular it opens import.lock NEVER: the
// incumbent treats a held lock as healthy and returns nil without
// importing (importer.go:397-413), so a shadow holding it would stall
// the box's own import while every signal read green.
func readRunReports(stateDir string) ([]oracleRunReport, error) {
	return readRunReportsFS(os.DirFS(stateDir))
}

// readRunReportsFS is readRunReports over an injected fs.FS, so a test
// can wrap a real directory in a recording seam and assert import.lock
// is never among the paths opened.
func readRunReportsFS(fsys fs.FS) ([]oracleRunReport, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read state directory: %w", err)
	}

	var reports []oracleRunReport
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "report-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read run report %s: %w", name, err)
		}
		var r oracleRunReport
		if err := json.Unmarshal(data, &r); err != nil {
			return nil, fmt.Errorf("decode run report %s: %w", name, err)
		}
		reports = append(reports, r)
	}
	slices.SortFunc(reports, func(a, b oracleRunReport) int { return a.StartedAt.Compare(b.StartedAt) })
	return reports, nil
}

// readTagLogSkips reads every import-*.log under stateDir and returns
// the basenames beets skipped. Tag logs are written atomically
// (_salvage/internal/beets/beets.go:104) and are never swept, which is
// why they are the skip half's durability backstop. The logged paths
// are STAGING paths (importer.go:1096) and stageUser preserves the
// inbox entry's own name (importer.go:844-856), so filepath.Base of a
// skip line is the captured basename.
func readTagLogSkips(stateDir string) (map[string]bool, error) {
	return readTagLogSkipsFS(os.DirFS(stateDir))
}

// readTagLogSkipsFS is readTagLogSkips over an injected fs.FS, for the
// same recording-seam reason readRunReportsFS carries one.
func readTagLogSkipsFS(fsys fs.FS) (map[string]bool, error) {
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		return nil, fmt.Errorf("read state directory: %w", err)
	}

	skips := make(map[string]bool)
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasPrefix(name, "import-") || !strings.HasSuffix(name, ".log") {
			continue
		}
		data, err := fs.ReadFile(fsys, name)
		if err != nil {
			return nil, fmt.Errorf("read tag log %s: %w", name, err)
		}
		for _, path := range parseTagLogSkips(data) {
			skips[filepath.Base(path)] = true
		}
	}
	return skips, nil
}

// parseTagLogSkips extracts skipped paths from beets' tag-log lines,
// matching the incumbent's own parseSkips
// (_salvage/internal/beets/beets.go:115-135): a line starting "skip "
// holds one or more paths separated by "; ", each trimmed of at most
// one leading space, empty entries dropped.
func parseTagLogSkips(data []byte) []string {
	var skipped []string
	for line := range strings.SplitSeq(string(data), "\n") {
		rest, ok := strings.CutPrefix(strings.TrimRight(line, "\r"), "skip ")
		if !ok {
			continue
		}
		for part := range strings.SplitSeq(rest, ";") {
			trimmed := strings.TrimPrefix(part, " ")
			if trimmed != "" {
				skipped = append(skipped, trimmed)
			}
		}
	}
	return skipped
}

// reviewedBasenames reads <reviewDir>/<date>/<uploader>/ and returns,
// per uploader, the album basenames filed there. The match is a suffix
// grammar rather than an equality, because uniqueDestPath appends
// -HHMMSS on a same-day collision and then .2, .3
// (_salvage/internal/review/review.go:76-99). The date component is the
// RUN's StartedAt (importer.go:271), not the album's, so every dated
// directory is read.
func reviewedBasenames(reviewDir string) (map[string][]string, error) {
	dates, err := os.ReadDir(reviewDir)
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read review directory: %w", err)
	}

	result := make(map[string][]string)
	for _, dateEntry := range dates {
		if !dateEntry.IsDir() {
			continue
		}
		dateDir := filepath.Join(reviewDir, dateEntry.Name())
		uploaders, err := os.ReadDir(dateDir)
		if err != nil {
			return nil, fmt.Errorf("read review date directory %s: %w", dateEntry.Name(), err)
		}
		for _, uploaderEntry := range uploaders {
			if !uploaderEntry.IsDir() {
				continue
			}
			uploader := uploaderEntry.Name()
			names, err := os.ReadDir(filepath.Join(dateDir, uploader))
			if err != nil {
				return nil, fmt.Errorf("read review uploader directory %s/%s: %w", dateEntry.Name(), uploader, err)
			}
			for _, nameEntry := range names {
				result[uploader] = append(result[uploader], nameEntry.Name())
			}
		}
	}
	return result, nil
}

// reviewSuffixPattern compiles the collision-suffix grammar S9/R9
// state for basename b: b itself, b-HHMMSS, or b-HHMMSS.N for N >= 2.
func reviewSuffixPattern(basename string) *regexp.Regexp {
	return regexp.MustCompile(`^` + regexp.QuoteMeta(basename) + `(-\d{6}(\.\d+)?)?$`)
}

// reviewMatches reports whether one of names (the entries filed under
// one uploader's dated review directories) resolves basename under the
// suffix grammar.
func reviewMatches(names []string, basename string) bool {
	if len(names) == 0 {
		return false
	}
	pattern := reviewSuffixPattern(basename)
	return slices.ContainsFunc(names, pattern.MatchString)
}
