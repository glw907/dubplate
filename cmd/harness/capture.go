package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
	"github.com/spf13/cobra"
)

// contactEnvVar names the universal identity input: MusicBrainz's
// terms require every client to advertise a reachable contact, and
// this tool never invents one (never a literal, per CLAUDE.md's
// universality rule).
const contactEnvVar = "DUBPLATE_HARNESS_CONTACT"

// requestInterval is the wire discipline every capture run honors: at
// most one request per second against the live API, regardless of
// MusicBrainz's own configured rate limit.
const requestInterval = time.Second

// mbAPIRoot is the live MusicBrainz WS/2 host; the capture tool always
// talks to it directly; it is oracle.Derive's replay stub, not this
// tool, that ever points at 127.0.0.1.
const mbAPIRoot = "https://musicbrainz.org"

// liveFetchTimeout bounds one live request; beets' own client uses a
// 10s default (mechanics study Q5), generous enough for the maximal
// inc= release lookups this tool issues.
const liveFetchTimeout = 30 * time.Second

// maximalReleaseInclude is the release lookup's inc= parameter,
// mirroring beets' own RELEASE_INCLUDES (mechanics study Q5) so a
// captured case never forces a re-capture the first time a future
// matcher improvement wants a field this run already fetched (U6).
const maximalReleaseInclude = "artists+media+recordings+release-groups+labels+" +
	"artist-credits+aliases+recording-level-rels+work-rels+work-level-rels+" +
	"artist-rels+isrcs+url-rels+release-rels+genres+tags"

// missFileName matches internal/harness/oracle's own constant; the
// two packages agree on it through the file format documented in
// corpus/README.md, not a shared symbol.
const missFileName = "misses.json"

// fetcher performs one live GET against the MusicBrainz API and
// returns the raw response body. liveFetch is the only production
// implementation; tests substitute one that never touches the network
// (no live network in go test, ever).
type fetcher func(ctx context.Context, pathQuery, userAgent string) ([]byte, error)

// doFetch is a test seam, following this repository's existing
// package-var-override convention (internal/harness/oracle's
// loadProfile and caseDir).
var doFetch fetcher = liveFetch

// sleep is the wire-discipline test seam: production waits a real
// second between requests; tests substitute a recording no-op so the
// suite never waits on real time.
var sleep = time.Sleep

// newCaptureCmd builds the capture subcommand: --mbid seeds a case
// from a release id, --case consumes an existing case's miss list.
// Exactly one of the two is required, and neither runs without an
// identified contact.
func newCaptureCmd(environ []string, version string) *cobra.Command {
	var mbid, out, caseDir string

	cmd := &cobra.Command{
		Use:           "capture",
		Short:         "Capture MusicBrainz responses into a corpus case",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			switch {
			case mbid != "" && caseDir != "":
				return errors.New("--mbid and --case are mutually exclusive")
			case mbid != "":
				if out == "" {
					return errors.New("--out is required with --mbid")
				}
			case caseDir == "":
				return errors.New("one of --mbid or --case is required")
			}

			contact, err := requireContact(environ)
			if err != nil {
				return err
			}

			if mbid != "" {
				if err := runCapture(cmd.Context(), version, contact, mbid, out); err != nil {
					return err
				}
				_, err := fmt.Fprintf(cmd.ErrOrStderr(), "seeded case at %s\n", out)
				return err
			}

			n, err := runCaptureMisses(cmd.Context(), version, contact, caseDir)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintf(cmd.ErrOrStderr(), "captured %d miss(es) into %s\n", n, caseDir)
			return err
		},
	}

	cmd.Flags().StringVar(&mbid, "mbid", "", "MusicBrainz release id to seed a case from")
	cmd.Flags().StringVar(&out, "out", "", "case directory to create (with --mbid)")
	cmd.Flags().StringVar(&caseDir, "case", "", "existing case directory whose miss list to capture")
	return cmd
}

// requireContact returns the identified User-Agent contact from
// environ's DUBPLATE_HARNESS_CONTACT entry, or an error naming the
// missing variable: the tool refuses to touch the live API
// unidentified.
func requireContact(environ []string) (string, error) {
	for _, kv := range environ {
		name, value, ok := strings.Cut(kv, "=")
		if ok && name == contactEnvVar && value != "" {
			return value, nil
		}
	}
	return "", fmt.Errorf("%s is not set; the harness refuses to hit MusicBrainz unidentified", contactEnvVar)
}

// userAgent builds the identified User-Agent every live request
// carries, per corpus/README.md's wire discipline.
func userAgent(version, contact string) string {
	return fmt.Sprintf("dubplate-harness/%s ( %s )", version, contact)
}

// runCapture seeds a case at outDir from mbid: it fetches the release
// at beets' own maximal inc= (U6), follows a Pseudo-Release's backward
// transl-tracklisting relation with the same inc= (mechanics study
// Q5), and writes every response verbatim before logging both
// requests in a fresh cassette.Case. The case is otherwise empty:
// Input, Profile, and EvalDate are a human's job once it is seeded.
func runCapture(ctx context.Context, version, contact, mbid, outDir string) error {
	if err := os.MkdirAll(filepath.Join(outDir, "responses"), 0o755); err != nil {
		return fmt.Errorf("create case dir: %w", err)
	}

	ua := userAgent(version, contact)
	var pace pacer

	pathQuery := releaseRequest(mbid)
	pace.wait()
	body, err := doFetch(ctx, pathQuery, ua)
	if err != nil {
		return fmt.Errorf("fetch release %s: %w", mbid, err)
	}

	respName, err := writeResponse(outDir, 1, body)
	if err != nil {
		return err
	}
	requests := []cassette.Request{{Method: http.MethodGet, PathQuery: pathQuery, Response: respName}}

	if followUp, ok := pseudoReleaseFollowUp(body); ok {
		followPathQuery := releaseRequest(followUp)
		pace.wait()
		followBody, err := doFetch(ctx, followPathQuery, ua)
		if err != nil {
			return fmt.Errorf("fetch pseudo-release follow-up %s: %w", followUp, err)
		}
		followRespName, err := writeResponse(outDir, 2, followBody)
		if err != nil {
			return err
		}
		requests = append(requests, cassette.Request{Method: http.MethodGet, PathQuery: followPathQuery, Response: followRespName})
	}

	c := &cassette.Case{ID: filepath.Base(outDir), Requests: requests}
	return c.Write(outDir)
}

// runCaptureMisses loads the existing case at caseDir, fetches exactly
// the requests its most recent derivation's misses.json names, and
// appends each as a logged Request before rewriting case.json and
// removing the now-consumed miss file. It reports how many requests
// it captured; zero with a nil error means there was nothing to do.
func runCaptureMisses(ctx context.Context, version, contact, caseDir string) (int, error) {
	c, err := cassette.LoadCase(caseDir)
	if err != nil {
		return 0, fmt.Errorf("load case: %w", err)
	}

	misses, err := readMisses(caseDir)
	if err != nil {
		return 0, err
	}
	if len(misses) == 0 {
		return 0, nil
	}

	ua := userAgent(version, contact)
	var pace pacer
	next := nextResponseNumber(caseDir)

	for _, miss := range misses {
		pace.wait()
		body, err := doFetch(ctx, miss.PathQuery, ua)
		if err != nil {
			return 0, fmt.Errorf("fetch miss %s %s: %w", miss.Method, miss.PathQuery, err)
		}
		respName, err := writeResponse(caseDir, next, body)
		if err != nil {
			return 0, err
		}
		next++
		c.Requests = append(c.Requests, cassette.Request{Method: miss.Method, PathQuery: miss.PathQuery, Response: respName})
	}

	if err := c.Write(caseDir); err != nil {
		return 0, err
	}
	if err := os.Remove(filepath.Join(caseDir, missFileName)); err != nil && !os.IsNotExist(err) {
		return 0, fmt.Errorf("remove consumed miss file: %w", err)
	}
	return len(misses), nil
}

// readMisses decodes caseDir's miss file, the shape
// internal/harness/oracle.Derive writes ([]stub.Miss, documented in
// corpus/README.md). A missing file means there is nothing to
// capture, not an error.
func readMisses(caseDir string) ([]stub.Miss, error) {
	data, err := os.ReadFile(filepath.Join(caseDir, missFileName))
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read miss file: %w", err)
	}

	var misses []stub.Miss
	if err := json.Unmarshal(data, &misses); err != nil {
		return nil, fmt.Errorf("parse miss file: %w", err)
	}
	return misses, nil
}

// nextResponseNumber returns one past the highest NNN.json response
// file already present under dir, or 1 when dir carries none.
func nextResponseNumber(dir string) int {
	entries, err := os.ReadDir(filepath.Join(dir, "responses"))
	if err != nil {
		return 1
	}
	highest := 0
	for _, e := range entries {
		if n, convErr := strconv.Atoi(strings.TrimSuffix(e.Name(), ".json")); convErr == nil && n > highest {
			highest = n
		}
	}
	return highest + 1
}

// writeResponse writes body verbatim to responses/NNN.json under dir
// (os.WriteFile, never re-encoded, so the file stays byte-identical to
// the wire response) and returns the path a case log entry should
// reference.
func writeResponse(dir string, n int, body []byte) (string, error) {
	name := filepath.Join("responses", fmt.Sprintf("%03d.json", n))
	if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
		return "", fmt.Errorf("write response %s: %w", name, err)
	}
	return name, nil
}

// releaseRequest returns the canonical GET path+query for a maximal
// release lookup: fmt=json forced, inc= carrying beets' own full
// RELEASE_INCLUDES list.
func releaseRequest(mbid string) string {
	return canonicalizeQuery("/ws/2/release/"+mbid, map[string]string{
		"fmt": "json",
		"inc": maximalReleaseInclude,
	})
}

// canonicalizeQuery joins path with params sorted byte-wise by their
// "key=value" pair, matching the cassette package's own
// canonicalization rule (corpus/README.md): params carries values
// already in their final percent-encoded wire form, since a decode-
// then-re-encode step can silently change what a Lucene-escaped search
// query or a "+"-joined inc= list means. Duplicated here rather than
// called from cassette because Task 1's cassette package ships frozen
// and keeps the rule unexported.
func canonicalizeQuery(path string, params map[string]string) string {
	pairs := make([]string, 0, len(params))
	for k, v := range params {
		pairs = append(pairs, k+"="+v)
	}
	slices.Sort(pairs)
	return path + "?" + strings.Join(pairs, "&")
}

// pseudoReleaseFollowUp reports the actual release id a Pseudo-Release
// body links to via a backward transl-tracklisting relation
// (mechanics study Q5, item 7): beets fetches and merges that release
// too, so a capture missing it would replay a 404 the moment pinned
// beets follows it.
func pseudoReleaseFollowUp(body []byte) (mbid string, ok bool) {
	var doc struct {
		Status    string `json:"status"`
		Relations []struct {
			Type      string `json:"type"`
			Direction string `json:"direction"`
			Release   struct {
				ID string `json:"id"`
			} `json:"release"`
		} `json:"relations"`
	}
	if err := json.Unmarshal(body, &doc); err != nil || doc.Status != "Pseudo-Release" {
		return "", false
	}
	for _, rel := range doc.Relations {
		if rel.Type == "transl-tracklisting" && rel.Direction == "backward" && rel.Release.ID != "" {
			return rel.Release.ID, true
		}
	}
	return "", false
}

// pacer enforces requestInterval between successive requests within
// one capture run, never delaying the first.
type pacer struct {
	started bool
}

func (p *pacer) wait() {
	if p.started {
		sleep(requestInterval)
	}
	p.started = true
}

// liveFetch is doFetch's production implementation: a real HTTP GET
// against mbAPIRoot, identified by userAgent, accepting only a 200
// response.
func liveFetch(ctx context.Context, pathQuery, userAgent string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, mbAPIRoot+pathQuery, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: liveFetchTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", pathQuery, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: status %d", pathQuery, resp.StatusCode)
	}
	return body, nil
}
