package grade

import (
	"cmp"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

// dubplate is the source name a session gives the server under test.
const dubplate = "dubplate"

var (
	// ErrStaleRow marks an active row that matched no finding in the
	// session it names.
	ErrStaleRow = errors.New("ledger row matches no finding")
	// ErrVersionMismatch marks an active row adjudicated against an
	// oracle version other than the one the session recorded.
	ErrVersionMismatch = errors.New("ledger row version differs from the session's")
	// ErrInvalidLedger marks a ledger the loader refuses; the error
	// carries every problem found.
	ErrInvalidLedger = errors.New("invalid ledger")
)

// Class says why a difference is accepted, and so what a row suppresses.
type Class string

const (
	// ClassByConstruction: an oracle cannot grade this surface. It
	// suppresses every oracle-tier kind in its scope.
	ClassByConstruction Class = "by-construction"
	// ClassUnoracled: no oracle covers this surface. It suppresses the
	// same kinds as ClassByConstruction.
	ClassUnoracled Class = "unoracled"
	// ClassContract: one source's response violates the OpenAPI document.
	// It suppresses the contract kind for that source alone.
	ClassContract Class = "contract"
	// ClassVolatile: the value varies per run or per server. It
	// suppresses the value kind only, so type and presence still compare.
	ClassVolatile Class = "volatile"
)

// State says whether a row can match yet.
type State string

const (
	// StateActive rows match findings in a recorded session.
	StateActive State = "active"
	// StateOwed rows describe behavior whose test a later slice builds;
	// they match nothing and are counted by slice.
	StateOwed State = "owed"
)

// Pattern selects the findings a row covers: one endpoint, one format
// ("json", "xml", or "any"), and a field path. A "*" segment stands for
// exactly one array-index segment; no other wildcard exists.
//
// A volatile or contract row covers the one field its path names, so a
// row written for a parent never hides a difference in a child. A
// by-construction or unoracled row covers its path and everything
// beneath it, since it names a surface no oracle can grade rather than a
// field.
type Pattern struct {
	Endpoint string `json:"endpoint"`
	Format   string `json:"format"`
	Path     string `json:"path"`
}

// Row is one adjudicated divergence. The first seven fields are the
// columns of the study's Markdown ledger.
type Row struct {
	ID       string   `json:"id"`
	Area     string   `json:"area"`
	Class    Class    `json:"class"`
	Sources  []string `json:"sources"`
	Oracle   string   `json:"oracle"`
	Dubplate string   `json:"dubplate"`
	Evidence string   `json:"evidence"`

	State State `json:"state"`
	// Source names, for a contract row, the source whose response
	// violates the document; it is one of Sources.
	Source string `json:"source,omitempty"`
	// Slice names, for an owed row, the slice whose plan builds its test.
	Slice string `json:"slice,omitempty"`
	// Pattern, Session, and Versions are required of an active row:
	// the findings it covers, the committed session that exercises it,
	// and the serverVersion of each non-dubplate source it was
	// adjudicated against.
	Pattern  Pattern           `json:"pattern,omitzero"`
	Session  string            `json:"session,omitempty"`
	Versions map[string]string `json:"versions,omitempty"`
}

// Ledger is the loaded divergence ledger.
type Ledger struct {
	Rows []Row
}

// LoadLedger reads and validates the ledger at path. An active row's
// session must be a directory under sessionsDir. Every problem found is
// reported in one error.
func LoadLedger(path, sessionsDir string) (*Ledger, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read ledger: %w", err)
	}
	var rows []Row
	if err := json.Unmarshal(raw, &rows, json.RejectUnknownMembers(true)); err != nil {
		return nil, fmt.Errorf("decode ledger %s: %w", path, err)
	}
	var problems []error
	seen := make(map[string]bool, len(rows))
	for i, r := range rows {
		label := fmt.Sprintf("row %d", i+1)
		if r.ID != "" {
			label = fmt.Sprintf("row %q", r.ID)
		}
		if r.ID != "" && seen[r.ID] {
			problems = append(problems, fmt.Errorf("%s: duplicate id", label))
		}
		seen[r.ID] = true
		for _, p := range r.problems(sessionsDir) {
			problems = append(problems, fmt.Errorf("%s: %w", label, p))
		}
	}
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", ErrInvalidLedger, path, err)
	}
	return &Ledger{Rows: rows}, nil
}

func (r Row) problems(sessionsDir string) []error {
	var ps []error
	bad := func(format string, args ...any) { ps = append(ps, fmt.Errorf(format, args...)) }
	for _, c := range []struct{ name, value string }{
		{"id", r.ID}, {"area", r.Area}, {"oracle", r.Oracle}, {"dubplate", r.Dubplate}, {"evidence", r.Evidence},
	} {
		if c.value == "" {
			bad("empty %s", c.name)
		}
	}
	switch r.Class {
	case ClassByConstruction, ClassUnoracled, ClassVolatile, ClassContract:
	default:
		bad("unknown class %q", r.Class)
	}
	if n := len(r.Sources); n < 1 || n > 2 || slices.Contains(r.Sources, "") || (n == 2 && r.Sources[0] == r.Sources[1]) {
		bad("sources %q must be one or two distinct names", r.Sources)
	}
	switch {
	case r.Class == ClassContract && r.Source == "":
		bad("contract row names no source")
	case r.Class == ClassContract && !slices.Contains(r.Sources, r.Source):
		bad("contract source %q is outside the sources %q", r.Source, r.Sources)
	case r.Class != ClassContract && r.Source != "":
		bad("source %q on a %s row", r.Source, r.Class)
	}
	if r.Pattern != (Pattern{}) {
		ps = append(ps, r.Pattern.problems()...)
	}
	switch r.State {
	case StateOwed:
		if r.Slice == "" {
			bad("owed row names no slice")
		}
	case StateActive:
		ps = append(ps, r.activeProblems(sessionsDir)...)
	default:
		bad("unknown state %q", r.State)
	}
	return ps
}

func (p Pattern) problems() []error {
	var ps []error
	if p.Endpoint == "" {
		ps = append(ps, errors.New("pattern has no endpoint"))
	}
	if !slices.Contains([]string{"json", "xml", "any"}, p.Format) {
		ps = append(ps, fmt.Errorf("pattern format %q is not json, xml, or any", p.Format))
	}
	if p.Path == "" {
		return append(ps, errors.New("pattern has no path"))
	}
	for seg := range strings.SplitSeq(p.Path, "/") {
		if seg == "" {
			ps = append(ps, fmt.Errorf("pattern path %q has an empty segment", p.Path))
		}
		if seg != "*" && strings.Contains(seg, "*") {
			ps = append(ps, fmt.Errorf("pattern path %q uses * inside segment %q", p.Path, seg))
		}
	}
	return ps
}

func (r Row) activeProblems(sessionsDir string) []error {
	var ps []error
	if r.Pattern == (Pattern{}) {
		ps = append(ps, errors.New("active row has an empty pattern"))
	}
	if len(r.Sources) != 2 {
		ps = append(ps, errors.New("active row names no pair of sources"))
	}
	switch {
	case r.Session == "":
		ps = append(ps, errors.New("active row names no session"))
	case !filepath.IsLocal(r.Session) || filepath.Base(r.Session) != r.Session:
		ps = append(ps, fmt.Errorf("session %q is not a plain directory name", r.Session))
	default:
		if info, err := os.Stat(filepath.Join(sessionsDir, r.Session)); err != nil || !info.IsDir() {
			ps = append(ps, fmt.Errorf("session %q: %w", r.Session, cmp.Or(err, fs.ErrNotExist)))
		}
	}
	for _, src := range r.Sources {
		if src != dubplate && r.Versions[src] == "" {
			ps = append(ps, fmt.Errorf("empty version for %s", src))
		}
	}
	for _, src := range slices.Sorted(maps.Keys(r.Versions)) {
		if src == dubplate || !slices.Contains(r.Sources, src) {
			ps = append(ps, fmt.Errorf("version for %s, which is no oracle source of the row", src))
		}
	}
	return ps
}

// Result is the ledger applied to one graded session.
type Result struct {
	// Matches holds, for each finding in the order given, the id of the
	// first row that suppresses it, or "" when it is unledgered.
	Matches []string
	// Counts is the number of findings each applied active row matched.
	// Every applied row appears, a stale one with a zero.
	Counts map[string]int
	// Owed is the number of owed rows in the ledger, by slice.
	Owed map[string]int
}

// Apply adjudicates findings from grading the session s, loaded from the
// directory named name, between the two sources in pair. A row applies
// only when it names that session and exactly that pair.
//
// Apply refuses with [ErrVersionMismatch], returning no Result, when an
// applying row's version for an oracle differs from the version s
// recorded for it. When an applying row matches no finding it returns the
// complete Result and an error wrapping [ErrStaleRow], naming each stale
// row.
func (l *Ledger) Apply(s *session.Session, name string, pair [2]string, findings []Finding) (Result, error) {
	res := Result{
		Matches: make([]string, len(findings)),
		Counts:  map[string]int{},
		Owed:    map[string]int{},
	}
	var applied []Row
	var problems []error
	for _, r := range l.Rows {
		if r.State == StateOwed {
			res.Owed[r.Slice]++
			continue
		}
		if !r.applies(name, pair) {
			continue
		}
		applied = append(applied, r)
		res.Counts[r.ID] = 0
		for _, src := range slices.Sorted(maps.Keys(r.Versions)) {
			want := r.Versions[src]
			if got := s.Sources[src].ServerVersion; got != want {
				problems = append(problems, fmt.Errorf("row %q, source %s: ledger %q, session %q: %w",
					r.ID, src, want, got, ErrVersionMismatch))
			}
		}
	}
	if err := errors.Join(problems...); err != nil {
		return Result{}, err
	}
	for i, f := range findings {
		for _, r := range applied {
			if !r.covers(f) {
				continue
			}
			res.Counts[r.ID]++
			if res.Matches[i] == "" {
				res.Matches[i] = r.ID
			}
		}
	}
	for _, r := range applied {
		if res.Counts[r.ID] == 0 {
			problems = append(problems, fmt.Errorf("row %q on session %s: %w", r.ID, name, ErrStaleRow))
		}
	}
	return res, errors.Join(problems...)
}

// applies reports whether r is an active row on the session name and on
// exactly the sources in pair, in either order.
func (r Row) applies(name string, pair [2]string) bool {
	return r.State == StateActive && r.Session == name &&
		slices.Equal(slices.Sorted(slices.Values(r.Sources)), slices.Sorted(slices.Values(pair[:])))
}

// subtree reports whether r's pattern covers everything beneath its path.
func (r Row) subtree() bool {
	return r.Class == ClassByConstruction || r.Class == ClassUnoracled
}

// selects reports whether r's pattern names the field at path in a
// response to endpoint in format.
func (r Row) selects(endpoint string, format client.Format, path string) bool {
	return r.Pattern.Endpoint == endpoint &&
		(r.Pattern.Format == "any" || r.Pattern.Format == string(format)) &&
		pathCovers(r.Pattern.Path, path, r.subtree())
}

// covers reports whether r, an active row on the finding's pair, selects
// and suppresses f.
func (r Row) covers(f Finding) bool {
	if !r.selects(f.Endpoint, f.Format, f.Path) {
		return false
	}
	switch r.Class {
	case ClassByConstruction, ClassUnoracled:
		switch f.Kind {
		case KindValue, KindType, KindPresence, KindOrder, KindIdentity:
			return true
		}
	case ClassVolatile:
		return f.Kind == KindValue
	case ClassContract:
		return f.Kind == KindContract && len(f.Sources) == 1 && f.Sources[0] == r.Source
	}
	return false
}

// pathCovers reports whether path is pattern, or, when subtree is set,
// lies beneath it, segment by segment; a "*" pattern segment covers only
// an all-digit segment.
func pathCovers(pattern, path string, subtree bool) bool {
	ps := strings.Split(pattern, "/")
	segs := strings.Split(path, "/")
	if len(ps) > len(segs) || (!subtree && len(ps) != len(segs)) {
		return false
	}
	for i, p := range ps {
		if p == "*" {
			if !isIndex(segs[i]) {
				return false
			}
			continue
		}
		if p != segs[i] {
			return false
		}
	}
	return true
}

func isIndex(seg string) bool {
	return seg != "" && !strings.ContainsFunc(seg, func(r rune) bool { return r < '0' || r > '9' })
}
