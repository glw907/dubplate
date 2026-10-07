package xmlrule

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"strings"
	"testing"
)

const (
	shelfOpen = `<subsonic-response xmlns="urn:example:shelf" status="ok" version="1">`
	realOpen  = `<subsonic-response xmlns="http://subsonic.org/restapi" status="ok" version="1.16.1">`
	closeEnv  = `</subsonic-response>`
)

// seedBody reads one recorded seed response.
func seedBody(t *testing.T, name string) []byte {
	t.Helper()
	return readFile(t, "../data/sessions/seed/"+name)
}

// want is one expected violation; message is a fragment of its text.
type want struct {
	ordinal int
	attr    string
	source  Source
	message string
}

func check(t *testing.T, v *Validator, endpoint string, status int, contentType, body string) Result {
	t.Helper()
	res, err := v.Check(context.Background(), endpoint, status, contentType, []byte(body))
	if err != nil {
		t.Fatalf("Check: %v", err)
	}
	return res
}

func assertViolations(t *testing.T, res Result, wants []want) {
	t.Helper()
	wantOutcome := Invalid
	if len(wants) == 0 {
		wantOutcome = Valid
	}
	if res.Outcome != wantOutcome {
		t.Fatalf("outcome = %s, want %s; violations %+v", res.Outcome, wantOutcome, res.Violations)
	}
	if len(res.Violations) != len(wants) {
		t.Fatalf("got %d violations, want %d: %+v", len(res.Violations), len(wants), res.Violations)
	}
	for i, w := range wants {
		got := res.Violations[i]
		if got.Ordinal != w.ordinal || got.Attr != w.attr || got.Source != w.source || !strings.Contains(got.Message, w.message) {
			t.Errorf("violation %d = %+v, want ordinal %d attr %q source %s message containing %q",
				i, got, w.ordinal, w.attr, w.source, w.message)
		}
	}
}

// TestProjectionScopeUnit uses the unit schema, where item is a Record in
// a crate and a Sleeve in a bin, each with its own additions.
func TestProjectionScopeUnit(t *testing.T) {
	v := shelfValidator(t, systemRunner(t))
	tests := []struct {
		name     string
		endpoint string
		body     string
		wants    []want
	}{
		{
			name:     "a Record addition in a crate is dropped",
			endpoint: "getCrate",
			body:     shelfOpen + `<crate><item id="1" gloss="matte"/></crate>` + closeEnv,
		},
		{
			name:     "a Record addition in a bin is a residual",
			endpoint: "getBin",
			body:     shelfOpen + `<bin><item id="1" gloss="matte"/></bin>` + closeEnv,
			wants:    []want{{3, "gloss", Residual, `"gloss" is declared by neither schema type Sleeve`}},
		},
		{
			name:     "a Sleeve addition in a bin is dropped",
			endpoint: "getBin",
			body:     shelfOpen + `<bin><item id="1" sheen="satin"/></bin>` + closeEnv,
		},
		{
			name:     "a Sleeve addition in a crate is a residual",
			endpoint: "getCrate",
			body:     shelfOpen + `<crate><item id="1" sheen="satin"/></crate>` + closeEnv,
			wants:    []want{{3, "sheen", Residual, "schema type Record"}},
		},
		{
			name:     "every level of the extension chain is declared, and element additions drop their subtree",
			endpoint: "getCrate",
			body: shelfOpen + `<crate><item id="1" label="L" size="2"><notes><note/></notes>` +
				`<track id="t" number="1" mood="calm"/></item></crate>` + closeEnv,
		},
		{
			name:     "additions reached through the rename table are dropped",
			endpoint: "ping",
			body:     `<subsonic-response xmlns="urn:example:shelf" status="failed" version="1" kind="x"><error code="10" hint="h"/>` + closeEnv,
		},
		{
			name:     "a declared attribute with a bad value is xmllint's",
			endpoint: "getCrate",
			body:     shelfOpen + `<crate><item id="1" size="big"/></crate>` + closeEnv,
			wants:    []want{{3, "size", Schema, "'big' is not a valid value"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolations(t, check(t, v, tt.endpoint, 200, "application/xml", tt.body), tt.wants)
		})
	}
}

// TestRealSeedAlbums validates one getAlbum body from each oracle, with
// gonic's residual title and album attributes removed by hand.
func TestRealSeedAlbums(t *testing.T) {
	v := realValidator(t)
	for _, name := range []string{"getalbum-navidrome.xml", "getalbum-gonic.xml"} {
		t.Run(name, func(t *testing.T) {
			body := string(readFile(t, "testdata/"+name))
			assertViolations(t, check(t, v, "getAlbum", 200, "application/xml", body), nil)
		})
	}
}

// TestRealPositionScope separates position-keyed scoping from a scope
// keyed on the element name and from a union over every type sharing it:
// album is a Child under albumList, and mediaType is a Child addition
// while isCompilation is an AlbumID3 one.
func TestRealPositionScope(t *testing.T) {
	v := realValidator(t)
	tests := []struct {
		name  string
		album string
		wants []want
	}{
		{"a Child addition validates", `<album id="1" isDir="true" title="A" mediaType="album"/>`, nil},
		{
			"an AlbumID3 addition is a residual",
			`<album id="1" isDir="true" title="A" isCompilation="true"/>`,
			[]want{{3, "isCompilation", Residual, "schema type Child"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := realOpen + `<albumList>` + tt.album + `<album id="2" isDir="true" title="B"/></albumList>` + closeEnv
			assertViolations(t, check(t, v, "getAlbumList", 200, "application/xml", body), tt.wants)
		})
	}
}

// TestRealBehindResidual proves a residual is dropped before xmllint
// runs: xmllint stops at a parent's first unexpected child, so a kept
// residual would hide the song's bad track.
func TestRealBehindResidual(t *testing.T) {
	v := realValidator(t)
	body := realOpen + `<album id="a" name="n" songCount="1" duration="1" created="2020-01-01T00:00:00Z">` +
		`<sticker colour="red"/><song id="s" isDir="false" title="t" track="x" titel="t"/></album>` + closeEnv
	assertViolations(t, check(t, v, "getAlbum", 200, "application/xml", body), []want{
		{3, "", Residual, `element "sticker"`},
		{4, "track", Schema, "'x' is not a valid value"},
		{4, "titel", Residual, `attribute "titel"`},
	})
}

// TestRealErrorLocation maps each xmllint line back to the element's
// ordinal in the original body.
func TestRealErrorLocation(t *testing.T) {
	v := realValidator(t)
	tests := []struct {
		name     string
		endpoint string
		body     string
		want     want
	}{
		{
			// envelope 1, album 2, releaseDate 3, artists 4, song 5 with
			// two children, song 8.
			name:     "the second song of a one-line body",
			endpoint: "getAlbum",
			body:     strings.Replace(string(readFile(t, "testdata/getalbum-navidrome.xml")), `track="2"`, `track="x"`, 1),
			want:     want{8, "track", Schema, "'x' is not a valid value"},
		},
		{
			// envelope 1, artist 2, two roles, album 5 with three
			// children, album 9.
			name:     "the second album of an indented body",
			endpoint: "getArtist",
			body:     strings.Replace(string(seedBody(t, "038-gonic.xml")), `songCount="3"`, `songCount="x"`, 1),
			want:     want{9, "songCount", Schema, "'x' is not a valid value"},
		},
		{
			name:     "an element after a multi-line text node",
			endpoint: "getGenres",
			body: realOpen + "<genres><genre songCount=\"1\" albumCount=\"1\">Jazz\nFusion\nEra</genre>" +
				`<genre songCount="x" albumCount="1">Rock</genre></genres>` + closeEnv,
			want: want{4, "songCount", Schema, "'x' is not a valid value"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := check(t, v, tt.endpoint, 200, "application/xml", tt.body)
			var got []Violation
			for _, vi := range res.Violations {
				if vi.Source == Schema {
					got = append(got, vi)
				}
			}
			if len(got) != 1 {
				t.Fatalf("got xmllint violations %+v, want one", got)
			}
			w := tt.want
			if got[0].Ordinal != w.ordinal || got[0].Attr != w.attr || !strings.Contains(got[0].Message, w.message) {
				t.Errorf("violation = %+v, want ordinal %d attr %q", got[0], w.ordinal, w.attr)
			}
		})
	}
}

// TestUnlocatedError proves an xmllint line no element opens is kept, at
// ordinal 0.
func TestUnlocatedError(t *testing.T) {
	run, err := fakeRunner(t, pinnedLibxml, "-:999: Schemas validity error : Element 'x': bad.\n- fails to validate\n", 3)
	if err != nil {
		t.Fatal(err)
	}
	v := shelfValidator(t, run)
	body := shelfOpen + `<crate><item id="1"/></crate>` + closeEnv
	assertViolations(t, check(t, v, "getCrate", 200, "application/xml", body), []want{{0, "", Schema, "Element 'x': bad."}})
}

// TestEnvelopeSurvives proves the projection keeps the envelope's status
// and version, which other types' additions also name, and drops its own
// additions.
func TestEnvelopeSurvives(t *testing.T) {
	v := realValidator(t)
	p, err := v.schema.project(readFile(t, "testdata/getalbum-navidrome.xml"))
	if err != nil {
		t.Fatal(err)
	}
	dec := xml.NewDecoder(bytes.NewReader(p.doc))
	var root xml.StartElement
	for {
		tok, err := dec.Token()
		if err != nil {
			t.Fatal(err)
		}
		if s, ok := tok.(xml.StartElement); ok {
			root = s
			break
		}
	}
	got := make(map[string]string)
	for _, a := range root.Attr {
		got[a.Name.Local] = a.Value
	}
	if got["status"] != "ok" || got["version"] != "1.16.1" {
		t.Errorf("projected envelope attributes = %v, want status ok and version 1.16.1", got)
	}
	for _, dropped := range []string{"type", "serverVersion", "openSubsonic"} {
		if _, ok := got[dropped]; ok {
			t.Errorf("projected envelope keeps %s", dropped)
		}
	}
}

func TestBinding(t *testing.T) {
	v := realValidator(t)
	album := string(readFile(t, "testdata/getalbum-navidrome.xml"))
	tests := []struct {
		name        string
		endpoint    string
		status      int
		contentType string
		body        string
		wants       []want
	}{
		{
			name: "getAlbum with no payload", endpoint: "getAlbum", status: 200, contentType: "application/xml",
			body:  realOpen + closeEnv,
			wants: []want{{1, "", Binding, "lacks its payload element album"}},
		},
		{
			name: "getAlbum carrying artists", endpoint: "getAlbum", status: 200, contentType: "application/xml",
			body: realOpen + `<artists ignoredArticles="The"/>` + closeEnv,
			wants: []want{
				{1, "", Binding, "carries artists, which is not its payload"},
				{1, "", Binding, "lacks its payload element album"},
			},
		},
		{
			name: "an ok ping carrying error", endpoint: "ping", status: 200, contentType: "application/xml",
			body:  realOpen + `<error code="10"/>` + closeEnv,
			wants: []want{{1, "", Binding, "carries error, which is not its payload"}},
		},
		{
			name: "a ping with no payload", endpoint: "ping", status: 200, contentType: "application/xml",
			body: realOpen + closeEnv,
		},
		{
			name: "a failed body without error", endpoint: "getAlbum", status: 200, contentType: "application/xml",
			body:  `<subsonic-response xmlns="http://subsonic.org/restapi" status="failed" version="1.16.1"/>`,
			wants: []want{{1, "", Binding, "carries no error element"}},
		},
		{
			name: "a failed body with error", endpoint: "getAlbum", status: 200, contentType: "text/xml",
			body: `<subsonic-response xmlns="http://subsonic.org/restapi" status="failed" version="1.16.1"><error code="70" message="not found"/>` + closeEnv,
		},
		{
			name: "an undeclared status", endpoint: "getAlbum", status: 404, contentType: "application/xml",
			body:  album,
			wants: []want{{1, "", Binding, "HTTP status 404"}},
		},
		{
			name: "a non-XML content type", endpoint: "getAlbum", status: 200, contentType: "application/json",
			body:  album,
			wants: []want{{1, "", Binding, `content type "application/json"`}},
		},
		{
			name: "an XML content type with a charset", endpoint: "getAlbum", status: 200, contentType: "text/xml; charset=utf-8",
			body: album,
		},
		{
			name: "an XML error from a binary endpoint", endpoint: "getCoverArt", status: 200, contentType: "text/xml",
			body: `<subsonic-response xmlns="http://subsonic.org/restapi" status="failed" version="1.16.1"><error code="70"/>` + closeEnv,
		},
		{
			name: "an ok envelope from a binary endpoint", endpoint: "getCoverArt", status: 200, contentType: "text/xml",
			body:  realOpen + closeEnv,
			wants: []want{{1, "", Binding, "binary body"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolations(t, check(t, v, tt.endpoint, tt.status, tt.contentType, tt.body), tt.wants)
		})
	}
}

func TestDeclaredUnvalidated(t *testing.T) {
	v := realValidator(t)
	res := check(t, v, "getOpenSubsonicExtensions", 200, "application/xml", string(seedBody(t, "034-navidrome.xml")))
	if res.Outcome != Unvalidated || res.Element != "openSubsonicExtensions" || len(res.Violations) != 0 {
		t.Errorf("result = %+v, want declared unvalidated naming openSubsonicExtensions", res)
	}
}

func TestDocumentStates(t *testing.T) {
	v := realValidator(t)
	tests := []struct {
		name string
		body string
		want want
	}{
		{"empty body", "", want{0, "", Document, "empty body"}},
		{"unclosed element", realOpen + `<album>`, want{0, "", Document, "malformed XML"}},
		{"mismatched end tag", realOpen + `</album>`, want{0, "", Document, "malformed XML"}},
		{"no root element", "  \n", want{0, "", Document, "no root element"}},
		{"a second root", realOpen + closeEnv + realOpen + closeEnv, want{2, "", Document, "second root"}},
		{"text outside the root", realOpen + closeEnv + "x", want{0, "", Document, "outside the root"}},
		{
			"a prefixed body",
			`<sub:subsonic-response xmlns:sub="http://subsonic.org/restapi" status="ok" version="1.16.1"/>`,
			want{1, "", Document, `prefix "sub" is not supported by the projection`},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertViolations(t, check(t, v, "ping", 200, "application/xml", tt.body), []want{tt.want})
		})
	}
}

func TestCheckErrors(t *testing.T) {
	v := realValidator(t)
	tests := []struct {
		name        string
		endpoint    string
		contentType string
		want        error
	}{
		{"an endpoint the bundle lacks", "getNothing", "application/xml", ErrUnknownEndpoint},
		{"a binary success", "getCoverArt", "image/jpeg", ErrNotXML},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := v.Check(context.Background(), tt.endpoint, 200, tt.contentType, []byte("\xff\xd8"))
			if !errors.Is(err, tt.want) {
				t.Errorf("err = %v, want %v", err, tt.want)
			}
		})
	}
}
