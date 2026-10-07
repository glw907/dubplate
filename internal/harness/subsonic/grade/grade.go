package grade

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/harness/subsonic/xmlrule"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
	"github.com/glw907/dubplate/internal/opensubsonic/conformance"
)

// errLibraryMismatch marks a pair of sources that served different
// fixture libraries, whose answers cannot be compared.
var errLibraryMismatch = errors.New("sources served different libraries")

// envelopePath is where a contract finding about a whole body sits: a
// body with no route, an undeclared status, or nothing to validate has no
// narrower place, and the envelope member is one every row can name.
const envelopePath = "subsonic-response"

// Config names the files a grade reads.
type Config struct {
	// Ledger is the divergence ledger, and Sessions the directory holding
	// the sessions its active rows name.
	Ledger, Sessions string
	// OpenAPI is the directory holding the pinned OpenAPI document.
	OpenAPI string
	// Manifest is the fixture manifest every graded session was served
	// from.
	Manifest string
}

// Grader grades recorded sessions against one ledger, one OpenAPI
// document, and one fixture library.
type Grader struct {
	ledger   *Ledger
	contract *contract
	lib      *library
}

// New loads everything cfg names. The OpenAPI document must match the
// pinned sha256.
func New(cfg Config) (*Grader, error) {
	l, err := LoadLedger(cfg.Ledger, cfg.Sessions)
	if err != nil {
		return nil, err
	}
	c, err := loadContract(cfg.OpenAPI, pinnedDocSum)
	if err != nil {
		return nil, err
	}
	lib, err := fixture.Load(cfg.Manifest)
	if err != nil {
		return nil, err
	}
	return &Grader{ledger: l, contract: c, lib: newLibrary(lib.Tracks())}, nil
}

// Report is one graded session: every finding, sorted, and the ledger
// applied to them.
type Report struct {
	findings []Finding // paired by index with ledger.Matches
	session  string
	sources  [2]string
	ledger   Result
	// unvalidated counts the XML bodies the contract tier declared
	// unvalidated, keyed by endpoint and source.
	unvalidated map[string]int
}

// Len returns the number of findings.
func (r *Report) Len() int { return len(r.findings) }

// Unledgered returns the number of findings no ledger row covers.
func (r *Report) Unledgered() int {
	n := 0
	for _, m := range r.ledger.Matches {
		if m == "" {
			n++
		}
	}
	return n
}

// Grade grades the session in dir, comparing the two sources in pair.
// The session's name, the directory's base name, is the name ledger rows
// use.
//
// Grade refuses, with no Report, a source the session lacks, two sources
// recorded over different libraries, naming both digests, and a ledger
// row adjudicated against another oracle version ([ErrVersionMismatch]).
// It also refuses when the XML contract tier fails, wrapping
// [xmlrule.ErrToolFailure], since a report without the tier would pass
// for a clean one.
// When an applied row matches nothing it returns the Report and an error
// wrapping [ErrStaleRow].
func (g *Grader) Grade(ctx context.Context, dir string, pair [2]string) (*Report, error) {
	s, err := session.Load(dir)
	if err != nil {
		return nil, err
	}
	name := filepath.Base(dir)
	var missing []error
	for _, src := range pair {
		if _, ok := s.Sources[src]; !ok {
			missing = append(missing, fmt.Errorf("session %s has no source %q", name, src))
		}
	}
	if err := errors.Join(missing...); err != nil {
		return nil, err
	}
	if pair[0] == pair[1] {
		return nil, fmt.Errorf("grade %s against itself", pair[0])
	}
	left, right := s.Sources[pair[0]].LibraryDigest, s.Sources[pair[1]].LibraryDigest
	if left != right {
		return nil, fmt.Errorf("%w: %s %s, %s %s", errLibraryMismatch, pair[0], left, pair[1], right)
	}
	var scopes []Row
	for _, r := range g.ledger.Rows {
		if r.Class == ClassByConstruction && r.applies(name, pair) {
			scopes = append(scopes, r)
		}
	}
	findings := g.oracleTier(s, pair, scopes)
	contractFound, unvalidated, err := g.contractTier(ctx, s, pair)
	if err != nil {
		return nil, err
	}
	findings = append(findings, contractFound...)
	slices.SortFunc(findings, compareFindings)
	res, err := g.ledger.Apply(s, name, pair, findings)
	if errors.Is(err, ErrVersionMismatch) {
		return nil, err
	}
	return &Report{findings: findings, session: name, sources: pair, ledger: res, unvalidated: unvalidated}, err
}

func compareFindings(a, b Finding) int {
	return cmp.Or(
		cmp.Compare(a.Request, b.Request),
		cmp.Compare(a.Path, b.Path),
		cmp.Compare(a.Kind, b.Kind),
		slices.Compare(a.Sources, b.Sources),
		slices.Compare(a.Values, b.Values),
	)
}

// graded is one request's two answers, decoded.
type graded struct {
	req   session.Request
	trees [2]*node
}

// oracleTier diffs the pair's answers to every request. It decodes every
// answer first, seeds the id pairing from the session's resolution map,
// learns every pair the answers yield, and only then compares, so an id
// first paired by a later request is known to an earlier one.
func (g *Grader) oracleTier(s *session.Session, pair [2]string, scopes []Row) []Finding {
	var found []Finding
	var work []graded
	for _, req := range s.Requests {
		base := Finding{Request: req.Key(), Endpoint: req.Endpoint, Format: req.Format, Sources: pair[:]}
		var resps [2]*session.Response
		for i, src := range pair {
			for k := range req.Responses {
				if req.Responses[k].Source == src {
					resps[i] = &req.Responses[k]
				}
			}
		}
		if resps[0] == nil || resps[1] == nil {
			f := base
			f.Kind, f.Path, f.Values = KindMissingSource, envelopePath, []string{answered(resps[0]), answered(resps[1])}
			found = append(found, f)
			continue
		}
		var w graded
		w.req = req
		var errs [2]error
		for i, r := range resps {
			w.trees[i], errs[i] = decode(r.Body, req.Format)
		}
		if errs[0] != nil || errs[1] != nil {
			f := base
			f.Kind, f.Path, f.Values = KindParse, envelopePath, []string{parseNote(req.Format, errs[0]), parseNote(req.Format, errs[1])}
			found = append(found, f)
			continue
		}
		work = append(work, w)
	}

	ids := newPairing()
	for _, w := range work {
		d := g.differFor(w.req, pair, ids, scopes)
		path := payloadPath(w.trees[0])
		for _, p := range w.req.Params {
			if bySource, ok := s.Resolutions[p.Value]; ok {
				ids.add(d.scope(path), spaceEntity, bySource[pair[0]], bySource[pair[1]])
			}
		}
	}
	for _, learn := range []bool{true, false} {
		for _, w := range work {
			d := g.differFor(w.req, pair, ids, scopes)
			d.learn = learn
			ctx := identCtx{endpoint: w.req.Endpoint, param: d.param}
			d.compare("", "", w.trees[0], w.trees[1], ctx, ctx)
			found = append(found, d.found...)
		}
	}
	return found
}

func (g *Grader) differFor(req session.Request, pair [2]string, ids *pairing, scopes []Row) *differ {
	d := &differ{
		lib: g.lib, ids: ids, scopes: scopes,
		request: req.Key(), endpt: req.Endpoint, format: req.Format, sources: pair,
	}
	for _, p := range req.Params {
		if p.Name == "id" {
			d.param = p.Value
		}
	}
	return d
}

// payloadPath is the path of the one member an answer carries beside the
// envelope's own fields, where the entity a request names sits.
func payloadPath(root *node) string {
	env := root.fields[envelopePath]
	if env == nil || env.kind != kindObject {
		return envelopePath
	}
	var members []string
	for name, n := range env.fields {
		if n.kind == kindObject || n.kind == kindArray {
			members = append(members, name)
		}
	}
	if len(members) != 1 {
		return envelopePath
	}
	return envelopePath + "/" + members[0]
}

func answered(r *session.Response) string {
	if r == nil {
		return "no response"
	}
	return "status " + strconv.Itoa(r.Status)
}

func parseNote(format client.Format, err error) string {
	if err == nil {
		return "parsed"
	}
	return string(format) + ": " + err.Error()
}

// contractTier checks every answer either source of the pair gave against
// the OpenAPI document: JSON bodies through the OpenAPI checker, XML
// bodies through the XML tier. Any outcome but valid is a finding
// carrying that one source, except a body the XML tier declares
// unvalidated, which is tallied by endpoint and source instead. A failed
// XML tier ends the check with an error.
func (g *Grader) contractTier(ctx context.Context, s *session.Session, pair [2]string) ([]Finding, map[string]int, error) {
	var found []Finding
	unvalidated := map[string]int{}
	for _, req := range s.Requests {
		for _, resp := range req.Responses {
			if !slices.Contains(pair[:], resp.Source) {
				continue
			}
			base := Finding{
				Request: req.Key(), Endpoint: req.Endpoint, Format: req.Format,
				Sources: []string{resp.Source}, Kind: KindContract, Path: envelopePath,
			}
			if req.Format != client.JSON {
				fs, skipped, err := g.contractXML(ctx, req, resp, base)
				if err != nil {
					return nil, nil, fmt.Errorf("xml contract check of %s from %s: %w", req.Key(), resp.Source, err)
				}
				if skipped {
					unvalidated[req.Endpoint+" "+resp.Source]++
				}
				found = append(found, fs...)
				continue
			}
			found = append(found, g.contractJSON(ctx, req, resp, base)...)
		}
	}
	return found, unvalidated, nil
}

func (g *Grader) contractJSON(ctx context.Context, req session.Request, resp session.Response, base Finding) []Finding {
	res := g.contract.json.Check(ctx, http.MethodGet, "/rest/"+req.Endpoint, resp.Status, resp.ContentType, resp.Body)
	if res.Outcome != conformance.Invalid {
		if res.Outcome == conformance.Valid {
			return nil
		}
		base.Values = []string{string(res.Outcome)}
		return []Finding{base}
	}
	// A body the contract check judged Invalid may not decode; a nil tree
	// makes at return nil, which renders as the absent value.
	tree, _ := decode(resp.Body, client.JSON)
	var found []Finding
	for _, v := range res.Violations {
		f := base
		if v.Pointer != "" {
			f.Path = pointerPath(v.Pointer)
		}
		f.Values = []string{at(tree, f.Path).render() + ": " + v.Reason}
		found = append(found, f)
	}
	return found
}

// contractXML checks one XML answer. An endpoint the document declares no
// path for is the no-route finding the JSON tier gives, and a binary
// endpoint's non-XML answer has nothing for this tier to check. The
// second result reports a body declared unvalidated. Every other failure
// of the tier is an error.
func (g *Grader) contractXML(ctx context.Context, req session.Request, resp session.Response, base Finding) (found []Finding, unvalidated bool, err error) {
	res, err := g.contract.xml.Check(ctx, req.Endpoint, resp.Status, resp.ContentType, resp.Body)
	switch {
	case errors.Is(err, xmlrule.ErrUnknownEndpoint):
		base.Values = []string{string(conformance.NoRoute)}
		return []Finding{base}, false, nil
	case errors.Is(err, xmlrule.ErrNotXML):
		return nil, false, nil
	case err != nil:
		return nil, false, err
	case res.Outcome == xmlrule.Unvalidated:
		return nil, true, nil
	}
	// A body whose elements cannot be walked leaves every ordinal at the
	// envelope.
	paths, _ := xmlElementPaths(resp.Body)
	tree, _ := decode(resp.Body, client.XML)
	for _, v := range res.Violations {
		f := base
		if v.Source != xmlrule.Document && v.Ordinal >= 2 && v.Ordinal <= len(paths) {
			f.Path = paths[v.Ordinal-1]
		}
		if v.Attr != "" {
			f.Path += "/" + v.Attr
		}
		f.Values = []string{at(tree, f.Path).render() + ": " + v.Message}
		found = append(found, f)
	}
	return found, false, nil
}

// pointerUnescape undoes RFC 6901 segment escaping.
var pointerUnescape = strings.NewReplacer("~1", "/", "~0", "~")

// pointerPath turns an RFC 6901 pointer into a finding path.
func pointerPath(pointer string) string {
	segs := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	for i, s := range segs {
		segs[i] = pointerUnescape.Replace(s)
	}
	return strings.Join(segs, "/")
}

// at returns the node at path in tree, or nil.
func at(tree *node, path string) *node {
	n := tree
	for seg := range strings.SplitSeq(path, "/") {
		switch {
		case n == nil:
			return nil
		case n.kind == kindObject:
			n = n.fields[seg]
		case n.kind == kindArray:
			i, err := strconv.Atoi(seg)
			if err != nil || i < 0 || i >= len(n.items) {
				return nil
			}
			n = n.items[i]
		default:
			return nil
		}
	}
	return n
}

// Format renders the report. Two grades of one session under one ledger
// format to the same bytes.
func (r *Report) Format() []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "session %s: %s against %s\n", r.session, r.sources[0], r.sources[1])
	byKind := map[Kind]int{}
	for _, f := range r.findings {
		byKind[f.Kind]++
	}
	fmt.Fprintf(&b, "findings: %d, unledgered %d\n", len(r.findings), r.Unledgered())
	for _, k := range slices.Sorted(maps.Keys(byKind)) {
		fmt.Fprintf(&b, "  %s: %d\n", k, byKind[k])
	}
	b.WriteString("declared unvalidated xml bodies:\n")
	for _, key := range slices.Sorted(maps.Keys(r.unvalidated)) {
		fmt.Fprintf(&b, "  %s: %d\n", key, r.unvalidated[key])
	}
	b.WriteString("ledger rows matched:\n")
	for _, id := range slices.Sorted(maps.Keys(r.ledger.Counts)) {
		fmt.Fprintf(&b, "  %s: %d\n", id, r.ledger.Counts[id])
	}
	b.WriteString("owed rows by slice:\n")
	for _, slice := range slices.Sorted(maps.Keys(r.ledger.Owed)) {
		fmt.Fprintf(&b, "  %s: %d\n", slice, r.ledger.Owed[slice])
	}
	request := ""
	for i, f := range r.findings {
		if f.Request != request {
			request = f.Request
			fmt.Fprintf(&b, "\n%s\n", request)
		}
		fmt.Fprintf(&b, "  %s %s\n", f.Kind, f.Path)
		for j, src := range f.Sources {
			fmt.Fprintf(&b, "    %s: %s\n", src, f.Values[j])
		}
		fmt.Fprintf(&b, "    ledger: %s\n", cmp.Or(r.ledger.Matches[i], "unledgered"))
	}
	return []byte(b.String())
}
