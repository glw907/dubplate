// Package cassette owns the conformance corpus's fixture schema and its
// on-disk I/O: one case per directory, a committed request log, pinned
// synthetic input, and a derived oracle record.
//
// A case directory holds case.json (this package's own metadata,
// encoded through [Case]'s json tags) plus a responses/ directory of
// verbatim MusicBrainz response bodies. [Request.Response] and
// [Case.Extra] name those files; nothing in this package embeds a
// response body inside case.json, so there is no re-encoding step that
// could corrupt one (encoding/json compacts and HTML-escapes
// json.RawMessage, which would break byte identity). [LoadCase] reads a
// referenced response file with os.ReadFile only to prove it exists and
// is readable; it never decodes the body.
//
// Canonicalization rule: a [Request.PathQuery] is a path plus its raw,
// percent-encoded query string, with the query's "&"-joined key=value
// pairs sorted byte-wise. The rule never decodes and re-encodes a pair:
// three wire facts documented alongside it in corpus/README.md would not
// survive that round trip (fmt=json forced on every request, inc=
// values joining logically with "+" but percent-escaped to "%2B" on the
// wire, and Lucene escapes in search queries). [LoadCase] rejects a case
// whose stored PathQuery is not already in this canonical form, so a
// corpus case can never depend on a lookup silently reordering pairs
// for it.
package cassette

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// caseFileName is the metadata file inside every case directory.
const caseFileName = "case.json"

// Case is one corpus case, loaded from its directory.
type Case struct {
	ID       string    `json:"id"`
	Input    Input     `json:"input"`   // items in pinned order; no file references at 2a
	Profile  string    `json:"profile"` // profile name under corpus/profiles ("estate", "stock")
	EvalDate string    `json:"eval_date"`
	Requests []Request `json:"requests"`           // the committed request log: an expectation, not scaffolding
	Extra    []string  `json:"extra,omitempty"`    // response files marked recall-extension (U8)
	Expected *Expected `json:"expected,omitempty"` // nil until derived
}

// Input is the matcher's input: synthetic beets items, file-free.
type Input struct {
	Items []ItemFields `json:"items"`
}

// ItemFields is the field bag one item carries into matching
// (title, artist, album, track, disc, length, year, mb ids, ...),
// serialized as a flat JSON object; length is set directly.
type ItemFields map[string]any

// Request is one logged oracle request in canonical form.
type Request struct {
	Method    string `json:"method"`
	PathQuery string `json:"path_query"` // path + "?" + query, canonicalized per the README rule
	Response  string `json:"response"`   // response filename within the case dir
}

// Expected is the oracle record the shim emits, decoded.
type Expected struct {
	Recommendation string             `json:"recommendation"` // strong | medium | low | none
	Action         string             `json:"action"`         // apply | skip; executed via beets' own quiet-mode mapping
	EchoedDate     string             `json:"echoed_date"`    // the date the shim observed; must equal EvalDate
	ConfigHash     string             `json:"config_hash"`    // hash of the shim's effective config
	Weights        map[string]float64 `json:"weights"`        // match.distance_weights the run scored under
	Thresholds     Thresholds         `json:"thresholds"`     // match.*_thresh and max_rec the run scored under
	Candidates     []Candidate        `json:"candidates"`     // in beets' final order
	Assignment     map[int]int        `json:"assignment"`     // item index -> track index, best candidate
	ExtraItems     []int              `json:"extra_items"`    // unmatched item indexes, in beets' sort order
	ExtraTracks    []int              `json:"extra_tracks"`   // unmatched track indexes, in beets' sort order
	AppliedFields  map[string]any     `json:"applied_fields"` // post-apply in-memory item fields, best candidate
}

// Thresholds is match's recommendation-boundary config, echoed
// verbatim by the shim. MaxRec is not uniformly numeric like the three
// *_thresh fields: it maps a penalty-key name to the recommendation
// level it caps a match at (for example "missing_tracks": "medium"),
// so a flat map[string]float64 cannot decode it.
type Thresholds struct {
	StrongRecThresh float64           `json:"strong_rec_thresh"`
	MediumRecThresh float64           `json:"medium_rec_thresh"`
	RecGapThresh    float64           `json:"rec_gap_thresh"`
	MaxRec          map[string]string `json:"max_rec"`
}

// Candidate is one scored release candidate.
type Candidate struct {
	Source    string             `json:"source"`   // metadata source name ("MusicBrainz")
	AlbumID   string             `json:"album_id"` // the U8 shared-subset key with Source
	Distance  float64            `json:"distance"`
	Penalties map[string]float64 `json:"penalties"` // per-key weighted distances
	Tracks    []float64          `json:"tracks"`    // per-track distances, assignment order
}

// Manifest pins everything the corpus's meaning depends on.
type Manifest struct {
	SchemaVersion string            `json:"schema_version"`
	Beets         string            `json:"beets"` // version and commit
	Unidecode     string            `json:"unidecode"`
	Jellyfish     string            `json:"jellyfish"`
	Lap           string            `json:"lap"`
	Numpy         string            `json:"numpy"`
	MediaFile     string            `json:"mediafile"` // pass 2d's sixth pin: the file-tags ledger's write/read-back layer
	ShimVersion   string            `json:"shim_version"`
	ProfileHashes map[string]string `json:"profile_hashes"` // profile name -> content hash
}

// LoadCase reads and validates the case directory at dir: case.json
// must parse into a [Case] with a non-empty ID, Profile, and EvalDate;
// every Request's PathQuery must already be in canonical form and name
// a readable response file under dir; and no Extra entry may name a
// response file a Request already consumes.
func LoadCase(dir string) (*Case, error) {
	data, err := os.ReadFile(filepath.Join(dir, caseFileName))
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", caseFileName, err)
	}

	var c Case
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", caseFileName, err)
	}
	if err := c.validate(dir); err != nil {
		return nil, err
	}
	return &c, nil
}

// validate checks case.json's parsed shape and cross-references it
// against dir's response files. It does not derive or reconcile
// anything; that is the oracle package's job at derivation time.
func (c *Case) validate(dir string) error {
	if c.ID == "" {
		return errors.New("case.json: id: missing")
	}
	if c.Profile == "" {
		return errors.New("case.json: profile: missing")
	}
	if c.EvalDate == "" {
		return errors.New("case.json: eval_date: missing")
	}

	consumed := make(map[string]bool, len(c.Requests))
	for i, req := range c.Requests {
		if req.Method == "" {
			return fmt.Errorf("case.json: requests[%d]: method: missing", i)
		}
		if canon := canonicalizePathQuery(req.PathQuery); canon != req.PathQuery {
			return fmt.Errorf("case.json: requests[%d]: path_query %q is not canonical, want %q", i, req.PathQuery, canon)
		}
		if req.Response == "" {
			return fmt.Errorf("case.json: requests[%d]: response: missing", i)
		}
		if err := checkResponseReadable(dir, req.Response); err != nil {
			return fmt.Errorf("case.json: requests[%d]: %w", i, err)
		}
		consumed[req.Response] = true
	}

	for i, extra := range c.Extra {
		if consumed[extra] {
			return fmt.Errorf("case.json: extra[%d]: %s already consumed by a request", i, extra)
		}
		if err := checkResponseReadable(dir, extra); err != nil {
			return fmt.Errorf("case.json: extra[%d]: %w", i, err)
		}
	}

	return nil
}

// checkResponseReadable proves name resolves to a readable file under
// dir; it discards the bytes, since neither [Case] nor its callers keep
// a response body in memory (they read it fresh from disk when needed).
func checkResponseReadable(dir, name string) error {
	if _, err := os.ReadFile(filepath.Join(dir, name)); err != nil {
		return fmt.Errorf("response %q: %w", name, err)
	}
	return nil
}

// Write encodes c to dir/case.json with a deterministic encoder
// (encoding/json sorts every map key it serializes, so the same Case
// always produces the same bytes) and creates dir if needed. Write
// never touches response body files; a case's responses/ directory is
// populated independently, by the capture tool or by whatever built the
// fixture, and referenced by filename only.
func (c *Case) Write(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create case dir: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encode case.json: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(filepath.Join(dir, caseFileName), data, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", caseFileName, err)
	}
	return nil
}

// canonicalizePathQuery sorts pathQuery's raw, percent-encoded query
// pairs byte-wise, leaving the path (including any trailing slash) and
// every pair's own encoding untouched. Two path/query strings that
// differ only in pair order canonicalize to the same string; any other
// difference, including a trailing slash, canonicalizes to a different
// one, so a near-miss variant never silently matches a stored request.
func canonicalizePathQuery(pathQuery string) string {
	path, query, hasQuery := strings.Cut(pathQuery, "?")
	if !hasQuery || query == "" {
		return path
	}
	pairs := strings.Split(query, "&")
	slices.Sort(pairs)
	return path + "?" + strings.Join(pairs, "&")
}

// LoadManifest reads and shape-validates the manifest at path: it must
// parse as a [Manifest] with a non-empty SchemaVersion and Beets pin.
// The dependency and profile pins a live venv must supply
// (Unidecode, Jellyfish, Lap, Numpy, ShimVersion, ProfileHashes) are not
// required here; reconciling them against a real venv is derivation-time
// work in the oracle package's Reconcile.
func LoadManifest(path string) (*Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}

	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parse manifest: %w", err)
	}
	if m.SchemaVersion == "" {
		return nil, errors.New("manifest: schema_version: missing")
	}
	if m.Beets == "" {
		return nil, errors.New("manifest: beets: missing")
	}
	return &m, nil
}
