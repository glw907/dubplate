package grade

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
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

// errCollectionMismatch marks a session recorded over other collections
// than the grader was given.
var errCollectionMismatch = errors.New("session was recorded over other collections")

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
	// Manifest is the fixture manifest the assignment's track shape is
	// checked against. It is also the one collection a session was served
	// from when Collections is empty.
	Manifest string
	// Collections is the ordered list of collections a graded session was
	// served from, with the names the recording gave them. The grader's
	// library is their union in list order.
	Collections []fixture.Collection
	// Assignment is the oracle assignment, which decides the oracle
	// tier's requests whenever a graded source is dubplate.
	Assignment string
}

// Grader grades recorded sessions against one ledger, one OpenAPI
// document, and one fixture library.
type Grader struct {
	ledger   *Ledger
	contract *contract
	lib      *library
	assign   *assignment
	// digest is the combined library digest of the collections, which
	// every graded session must carry.
	digest string
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
	cols := cfg.Collections
	if len(cols) == 0 {
		cols = []fixture.Collection{{Name: fixture.DefaultCollectionName, Manifest: cfg.Manifest}}
	}
	built, digest, err := buildCollections(cols)
	if err != nil {
		return nil, err
	}
	shape := built[0].Library
	if len(cfg.Collections) > 0 {
		if shape, err = fixture.Load(cfg.Manifest); err != nil {
			return nil, err
		}
	}
	a, err := loadAssignment(cfg.Assignment, shape.Tracks())
	if err != nil {
		return nil, err
	}
	var tracks []fixture.Track
	names := make([]string, len(built))
	for i, b := range built {
		tracks = append(tracks, b.Library.Tracks()...)
		names[i] = b.Name
	}
	return &Grader{ledger: l, contract: c, lib: newLibrary(tracks, names), assign: a, digest: digest}, nil
}

// buildCollections builds cols with [fixture.BuildCollections]. Build is
// the only source of a library digest, so the libraries are written to a
// directory that is removed before it returns.
func buildCollections(cols []fixture.Collection) ([]fixture.BuiltCollection, string, error) {
	root, err := os.MkdirTemp("", "dubplate-grade-")
	if err != nil {
		return nil, "", fmt.Errorf("create build directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(root) }()
	return fixture.BuildCollections(root, cols)
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
	// skips is set when one graded source is dubplate, and says which
	// requests the assignment kept out of the oracle tier.
	skips *skips
}

// skips counts, by endpoint, the requests a grade of dubplate against one
// oracle did not send to the oracle tier.
type skips struct {
	oracle string
	// notAssigned holds the endpoints the assignment does not grant to
	// the oracle, and unmapped the endpoints it lacks.
	notAssigned, unmapped map[string]int
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
// When one source is dubplate, the assignment decides which requests the
// oracle tier compares: a request to an endpoint the assignment does not
// grant to the other source, or lacks, is graded by the contract tier
// alone and counted in the report. A pair of oracles never consults the
// assignment.
//
// Grade refuses, with no Report, a source the session lacks, two sources
// recorded over different libraries, naming both digests, a session
// recorded over other collections than the grader's, naming the session's
// digest and the collections', and a ledger
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
	if left != g.digest {
		return nil, fmt.Errorf("%w: session %s carries library digest %s, collections %q combine to %s",
			errCollectionMismatch, name, left, g.lib.collections, g.digest)
	}
	var scopes []Row
	for _, r := range g.ledger.Rows {
		if classRules[r.Class].subtree && r.applies(name, pair) {
			scopes = append(scopes, r)
		}
	}
	oracleView, skipped := g.assigned(s, pair)
	findings := g.oracleTier(oracleView, pair, scopes)
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
	return &Report{findings: findings, session: name, sources: pair, ledger: res, unvalidated: unvalidated, skips: skipped}, err
}

// assigned returns the view of s the oracle tier compares. For a pair
// that names dubplate it holds only the requests the assignment grants to
// the pair's oracle, and the skips report counts the rest; any other pair
// gets s whole and no report.
func (g *Grader) assigned(s *session.Session, pair [2]string) (*session.Session, *skips) {
	var oracleSource string
	switch {
	case pair[0] == dubplate:
		oracleSource = pair[1]
	case pair[1] == dubplate:
		oracleSource = pair[0]
	default:
		return s, nil
	}
	sk := &skips{oracle: oracleSource, notAssigned: map[string]int{}, unmapped: map[string]int{}}
	view := *s
	view.Requests = nil
	for _, req := range s.Requests {
		e, ok := g.assign.lookup(req.Endpoint)
		switch {
		case !ok:
			sk.unmapped[req.Endpoint]++
		case !e.Oracle.grants(oracleSource):
			sk.notAssigned[req.Endpoint]++
		default:
			view.Requests = append(view.Requests, req)
		}
	}
	return &view, sk
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

	folders := [2]map[string]string{folderCollections(s, pair[0]), folderCollections(s, pair[1])}
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
			ca := identCtx{endpoint: w.req.Endpoint, param: d.param, folders: folders[0]}
			cb := ca
			cb.folders = folders[1]
			d.compare("", "", w.trees[0], w.trees[1], ca, cb)
			found = append(found, d.found...)
		}
	}
	return found
}

// folderCollections returns the collection each music folder of source
// stands for, by folder id, as the session's collection resolutions
// record it.
func folderCollections(s *session.Session, source string) map[string]string {
	m := map[string]string{}
	for key, bySource := range s.Resolutions {
		if name, ok := strings.CutPrefix(key, "collection:"); ok && bySource[source] != "" {
			m[bySource[source]] = name
		}
	}
	return m
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
	if r.skips != nil {
		fmt.Fprintf(&b, "not assigned to %s:\n", r.skips.oracle)
		for _, endpoint := range slices.Sorted(maps.Keys(r.skips.notAssigned)) {
			fmt.Fprintf(&b, "  %s: %d\n", endpoint, r.skips.notAssigned[endpoint])
		}
		b.WriteString("unmapped:\n")
		for _, endpoint := range slices.Sorted(maps.Keys(r.skips.unmapped)) {
			fmt.Fprintf(&b, "  %s: %d\n", endpoint, r.skips.unmapped[endpoint])
		}
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
