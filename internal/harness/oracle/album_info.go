//go:build oracle

// album_info.go is the Go half of ruling E's derived AlbumInfo ledger:
// it sets up the same replay stub and rendered profile a derivation runs
// under, then hands derive_album_info.py the release ids the case's own
// request log fetches by id. It carries the oracle tag because it
// launches Python; the gate-resident half of the analysis is
// internal/harness/analysis's ledger checks.
package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// releaseLookupPrefix is the WS/2 path a release lookup by id starts
// with. The search endpoint is /ws/2/release? and never matches it, so
// a search request contributes no ledger row: T1's translation contract
// establishes that a search hit reaches album_info through album_for_id
// anyway, once per id.
const releaseLookupPrefix = "/ws/2/release/"

// albumInfoInput is the envelope derive_album_info.py reads: the case's
// pinned date plus the release ids to translate, in log order.
type albumInfoInput struct {
	CaseID   string   `json:"case_id"`
	EvalDate string   `json:"eval_date"`
	Lookups  []string `json:"lookups"`
}

// DeriveAlbumInfo runs derive_album_info.py against one committed case
// and returns the ledger document it prints, one row per release-lookup
// request in the case's own log, unparsed. Which cases are subjects is
// the caller's decision; this driver only owns the mechanism.
func DeriveAlbumInfo(ctx context.Context, c *cassette.Case, opts Options) ([]byte, error) {
	lookups := releaseLookups(c)
	if len(lookups) == 0 {
		return nil, fmt.Errorf("oracle: album info %s: case logs no release lookup to translate", c.ID)
	}

	dir := opts.Dir
	if dir == "" {
		var err error
		dir, err = caseDir(c)
		if err != nil {
			return nil, err
		}
	}

	srv, err := stub.New(c, dir)
	if err != nil {
		return nil, fmt.Errorf("oracle: album info %s: start stub: %w", c.ID, err)
	}
	defer func() { _ = srv.Close() }()

	scratch, err := os.MkdirTemp("", "dubplate-album-info-*")
	if err != nil {
		return nil, fmt.Errorf("oracle: album info %s: scratch dir: %w", c.ID, err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	profile, err := renderProfile(c.Profile, scratch, filepath.Join(scratch, "library.db"), srv.Addr())
	if err != nil {
		return nil, err
	}
	profilePath := filepath.Join(scratch, "config.yaml")
	if err := os.WriteFile(profilePath, profile, 0o644); err != nil {
		return nil, fmt.Errorf("oracle: album info %s: write rendered profile: %w", c.ID, err)
	}

	input, err := json.Marshal(albumInfoInput{CaseID: c.ID, EvalDate: c.EvalDate, Lookups: lookups})
	if err != nil {
		return nil, fmt.Errorf("oracle: album info %s: encode input: %w", c.ID, err)
	}
	inputPath := filepath.Join(scratch, "input.json")
	if err := os.WriteFile(inputPath, input, 0o644); err != nil {
		return nil, fmt.Errorf("oracle: album info %s: write input: %w", c.ID, err)
	}

	raw, err := runPython(ctx, opts, "derive_album_info.py", profilePath, inputPath)
	if err != nil {
		return nil, fmt.Errorf("oracle: album info %s: %w", c.ID, err)
	}
	return raw, nil
}

// releaseLookups returns the release ids c's own request log fetches by
// id, in log order and without repeats. It selects the derivation's work
// where analysis.CheckAlbumInfoLedger's own walk validates the result,
// so the two never share a decision.
func releaseLookups(c *cassette.Case) []string {
	var mbids []string
	seen := make(map[string]bool, len(c.Requests))
	for _, req := range c.Requests {
		rest, ok := strings.CutPrefix(req.PathQuery, releaseLookupPrefix)
		if !ok {
			continue
		}
		mbid, _, _ := strings.Cut(rest, "?")
		if mbid == "" || seen[mbid] {
			continue
		}
		seen[mbid] = true
		mbids = append(mbids, mbid)
	}
	return mbids
}
