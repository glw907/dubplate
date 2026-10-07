package grade

import (
	"encoding/json/v2"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
)

// errInvalidAssignment marks an assignment the loader refuses; the error
// carries every problem found.
var errInvalidAssignment = errors.New("invalid assignment")

// Source names of the two oracles in a session.
const (
	sourceNavidrome = "navidrome"
	sourceGonic     = "gonic"
)

// notServed is the slice of an endpoint dubplate answers with error 0.
const notServed = "not-served-0.8"

// oracle says which recorded server grades an endpoint beyond the
// contract tier, which every endpoint takes.
type oracle string

const (
	// oracleBoth: Navidrome is primary and gonic secondary.
	oracleBoth oracle = "both"
	// oracleNavidrome: Navidrome alone grades the endpoint.
	oracleNavidrome oracle = "navidrome"
	// oracleGonic: gonic alone grades the endpoint.
	oracleGonic oracle = "gonic"
	// oracleContractOnly: neither oracle is a fair grader.
	oracleContractOnly oracle = "contract-only"
)

// grants reports whether o lets the oracle source grade an endpoint.
func (o oracle) grants(source string) bool {
	switch o {
	case oracleBoth:
		return source == sourceNavidrome || source == sourceGonic
	case oracleNavidrome:
		return source == sourceNavidrome
	case oracleGonic:
		return source == sourceGonic
	}
	return false
}

// endpointEntry assigns one endpoint of the pinned OpenAPI document.
type endpointEntry struct {
	Endpoint string `json:"endpoint"`
	Oracle   oracle `json:"oracle"`
	// Slice is the slice that first serves the endpoint, or [notServed].
	Slice  string `json:"slice"`
	Reason string `json:"reason"`
}

// areaEntry is one cross-cutting class of behavior the harness tests.
type areaEntry struct {
	Name   string `json:"name"`
	Slice  string `json:"slice"`
	Reason string `json:"reason"`
}

// agreementEntry is a behavior dubplate shares with both oracles. It
// suppresses nothing; its test is owed.
type agreementEntry struct {
	ID       string `json:"id"`
	Area     string `json:"area"`
	Behavior string `json:"behavior"`
	Evidence string `json:"evidence"`
	Test     string `json:"test"`
	// Slice is the slice whose plan builds the test.
	Slice string `json:"slice"`
}

// shapeEntry names the manifest album ("album artist/album") that carries
// one album shape.
type shapeEntry struct {
	Shape  string `json:"shape"`
	Album  string `json:"album"`
	Reason string `json:"reason"`
}

// unusedShape is an album shape the library cannot hold.
type unusedShape struct {
	Shape  string `json:"shape"`
	Reason string `json:"reason"`
}

// shapes holds, for each album shape an entry may name, the test that an
// album's tracks carry it.
var shapes = map[string]func([]fixture.Track) bool{
	"compilation": func(ts []fixture.Track) bool {
		return slices.ContainsFunc(ts, func(t fixture.Track) bool { return t.AlbumArtist == "Various Artists" })
	},
	"multi-disc": func(ts []fixture.Track) bool {
		return slices.ContainsFunc(ts, func(t fixture.Track) bool { return t.Disc > 1 })
	},
	"collaboration credit": func(ts []fixture.Track) bool {
		return slices.ContainsFunc(ts, func(t fixture.Track) bool { return len(t.Artists) > 1 })
	},
	"unicode": func(ts []fixture.Track) bool {
		return slices.ContainsFunc(ts, func(t fixture.Track) bool {
			return strings.ContainsFunc(t.AlbumArtist+t.Album+t.Title, func(r rune) bool { return r > unicode.MaxASCII })
		})
	},
}

// assignment is the loaded oracle assignment and area list. Only its file
// sets an endpoint's oracle.
type assignment struct {
	Endpoints     []endpointEntry  `json:"endpoints"`
	Areas         []areaEntry      `json:"areas"`
	Agreements    []agreementEntry `json:"agreements"`
	AlbumShapes   []shapeEntry     `json:"album_shapes"`
	NotApplicable []unusedShape    `json:"not_applicable"`

	byEndpoint map[string]endpointEntry
}

// loadAssignment reads and validates the assignment at path against the
// albums tracks declares. Every problem found is reported in one error.
func loadAssignment(path string, tracks []fixture.Track) (*assignment, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read assignment: %w", err)
	}
	var a assignment
	if err := json.Unmarshal(raw, &a, json.RejectUnknownMembers(true)); err != nil {
		return nil, fmt.Errorf("decode assignment %s: %w", path, err)
	}
	if len(a.Endpoints) == 0 {
		return nil, fmt.Errorf("%w: %s: no endpoints", errInvalidAssignment, path)
	}
	if err := errors.Join(a.problems(tracks)...); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errInvalidAssignment, path, err)
	}
	a.byEndpoint = make(map[string]endpointEntry, len(a.Endpoints))
	for _, e := range a.Endpoints {
		a.byEndpoint[e.Endpoint] = e
	}
	return &a, nil
}

func (a *assignment) problems(tracks []fixture.Track) []error {
	var ps []error
	bad := func(format string, args ...any) { ps = append(ps, fmt.Errorf(format, args...)) }
	seen := map[string]bool{}
	for i, e := range a.Endpoints {
		label := fmt.Sprintf("endpoint %d", i+1)
		if e.Endpoint != "" {
			label = fmt.Sprintf("endpoint %q", e.Endpoint)
		}
		switch {
		case e.Endpoint == "":
			bad("%s: empty name", label)
		case seen[e.Endpoint]:
			bad("%s: listed twice", label)
		}
		seen[e.Endpoint] = true
		if !slices.Contains([]oracle{oracleBoth, oracleNavidrome, oracleGonic, oracleContractOnly}, e.Oracle) {
			bad("%s: unknown oracle %q", label, e.Oracle)
		}
		if e.Slice == "" {
			bad("%s: empty slice", label)
		}
		if e.Reason == "" {
			bad("%s: empty reason", label)
		}
		if e.Slice == notServed && e.Oracle != oracleContractOnly {
			bad("%s: not served at launch, so oracle %q must be %q", label, e.Oracle, oracleContractOnly)
		}
	}
	areas := map[string]bool{}
	for _, ar := range a.Areas {
		switch {
		case ar.Name == "":
			bad("area with no name")
		case areas[ar.Name]:
			bad("area %q: listed twice", ar.Name)
		}
		areas[ar.Name] = true
		if ar.Slice == "" || ar.Reason == "" {
			bad("area %q: empty slice or reason", ar.Name)
		}
	}
	agreements := map[string]bool{}
	for i, ag := range a.Agreements {
		label := fmt.Sprintf("agreement %d", i+1)
		switch {
		case ag.ID == "":
			bad("%s: empty id", label)
		case agreements[ag.ID]:
			bad("agreement %q: listed twice", ag.ID)
		}
		agreements[ag.ID] = true
		if ag.ID != "" {
			label = fmt.Sprintf("agreement %q", ag.ID)
		}
		for _, c := range []struct{ name, value string }{
			{"area", ag.Area}, {"behavior", ag.Behavior}, {"evidence", ag.Evidence}, {"test", ag.Test}, {"slice", ag.Slice},
		} {
			if c.value == "" {
				bad("%s: empty %s", label, c.name)
			}
		}
	}
	albums := map[string][]fixture.Track{}
	for _, t := range tracks {
		key := t.AlbumArtist + "/" + t.Album
		albums[key] = append(albums[key], t)
	}
	for _, s := range a.AlbumShapes {
		has, known := shapes[s.Shape]
		ts, inLibrary := albums[s.Album]
		switch {
		case !known:
			bad("album shape %q: unknown shape", s.Shape)
		case !inLibrary:
			bad("album shape %q: album %q is not in the manifest", s.Shape, s.Album)
		case !has(ts):
			bad("album shape %q: album %q lacks the shape", s.Shape, s.Album)
		}
		if s.Reason == "" {
			bad("album shape %q: empty reason", s.Shape)
		}
	}
	for _, s := range a.NotApplicable {
		if s.Shape == "" || s.Reason == "" {
			bad("not-applicable shape %q: empty shape or reason", s.Shape)
		}
	}
	return ps
}

// lookup returns the entry for endpoint, and whether the assignment
// has one.
func (a *assignment) lookup(endpoint string) (endpointEntry, bool) {
	e, ok := a.byEndpoint[endpoint]
	return e, ok
}
