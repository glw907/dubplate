package grade

import (
	"cmp"
	"encoding/json/v2"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const committedManifest = "../data/fixture/manifest.json"

// The perturbation bodies: Navidrome's recorded answers to getAlbum for
// Alpha Artist's Single Disc, three songs with a featured guest on the
// third, in both formats.
var (
	albumJSON = filepath.Join(seedSession, "012-navidrome.json")
	albumXML  = filepath.Join(seedSession, "043-navidrome.xml")

	// Gonic's answers that carry the empty shapes the differ must tell
	// apart: getAlbum for the compilation (a null replayGain) and
	// getArtist (an album with releaseTypes [""]).
	gonicMixtape = filepath.Join(seedSession, "015-gonic.json")
	gonicArtist  = filepath.Join(seedSession, "008-gonic.json")
)

func testLibrary(t *testing.T) *library {
	t.Helper()
	lib, err := fixture.Load(committedManifest)
	if err != nil {
		t.Fatal(err)
	}
	return newLibrary(lib.Tracks(), []string{fixture.DefaultCollectionName})
}

// diffBodies runs the oracle tier alone over one getAlbum request whose
// two sources answered left and right.
func diffBodies(t *testing.T, format client.Format, left, right []byte) []Finding {
	t.Helper()
	return diffBodiesIn(t, testLibrary(t), format, left, right)
}

// diffBodiesIn is diffBodies against the given library.
func diffBodiesIn(t *testing.T, lib *library, format client.Format, left, right []byte) []Finding {
	t.Helper()
	return diffEndpoint(t, lib, "getAlbum", "album:Alpha Artist/Single Disc", format, left, right)
}

// diffEndpoint is diffBodiesIn for any endpoint and entity reference.
func diffEndpoint(t *testing.T, lib *library, endpoint, ref string, format client.Format, left, right []byte) []Finding {
	t.Helper()
	ct := map[client.Format]string{client.JSON: "application/json", client.XML: "application/xml"}[format]
	s := &session.Session{Requests: []session.Request{{
		Endpoint: endpoint, Format: format,
		Params: []session.Param{{Name: "id", Value: ref}},
		Responses: []session.Response{
			{Source: "left", Status: 200, ContentType: ct, Body: left},
			{Source: "right", Status: 200, ContentType: ct, Body: right},
		},
	}}}
	g := &Grader{lib: lib}
	return g.oracleTier(s, [2]string{"left", "right"}, nil)
}

// editJSON decodes body, applies edit to the album member, and encodes
// the result. Every number in the album body is an integer, which a
// float64 round trip keeps, so the "re-encoded" row proves the trip
// itself changes nothing the differ sees.
func editJSON(t *testing.T, body []byte, edit func(album map[string]any)) []byte {
	t.Helper()
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	envelope, _ := doc["subsonic-response"].(map[string]any)
	album, ok := envelope["album"].(map[string]any)
	if !ok {
		t.Fatal("body has no album object")
	}
	edit(album)
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func songs(album map[string]any) []map[string]any {
	list, _ := album["song"].([]any)
	out := make([]map[string]any, len(list))
	for i, s := range list {
		out[i], _ = s.(map[string]any)
	}
	return out
}

func replaceOnce(t *testing.T, body []byte, old, new string) []byte {
	t.Helper()
	if n := strings.Count(string(body), old); n != 1 {
		t.Fatalf("%q occurs %d times in the body, want once", old, n)
	}
	return []byte(strings.Replace(string(body), old, new, 1))
}

// replaceFirst is replaceOnce for a literal the body repeats: it changes
// the first occurrence and requires at least one.
func replaceFirst(t *testing.T, body []byte, old, new string) []byte {
	t.Helper()
	if !strings.Contains(string(body), old) {
		t.Fatalf("%q does not occur in the body", old)
	}
	return []byte(strings.Replace(string(body), old, new, 1))
}

// jsonEnvelope and xmlEnvelope wrap payload in a hand-written envelope
// that names its own server, for rows that need a shape no recorded body
// has.
func jsonEnvelope(payload string) []byte {
	return []byte(`{"subsonic-response":{"status":"ok","version":"1.16.1","type":"a",` +
		`"serverVersion":"a-1","openSubsonic":true` + payload + `}}`)
}

func xmlEnvelope(payload string) []byte {
	return []byte(`<subsonic-response xmlns="http://subsonic.org/restapi" status="ok" version="1.16.1" type="a">` +
		payload + `</subsonic-response>`)
}

func failedEnvelope(server, message string, code int) []byte {
	return []byte(`{"subsonic-response":{"status":"failed","version":"1.16.1","type":"` + server +
		`","serverVersion":"` + server + `-1","openSubsonic":true,"error":{"code":` + strconv.Itoa(code) +
		`,"message":` + strconv.Quote(message) + `}}}`)
}

func readBody(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// TestDifferPerturbations compares one real body against itself, then
// against copies that each carry one change, and asserts exactly the one
// finding the change makes.
func TestDifferPerturbations(t *testing.T) {
	jsonBody, xmlBody := readBody(t, albumJSON), readBody(t, albumXML)
	tests := []struct {
		name   string
		format client.Format
		left   func(t *testing.T, body []byte) []byte // nil leaves the left body as recorded
		right  func(t *testing.T, body []byte) []byte
		empty  bool // grade against a library with no tracks
		kind   Kind // "" means no finding at all
		path   string
		// from, when set, is the recorded body to perturb instead of the
		// album's, and endpoint and ref name the request it answered.
		from, endpoint, ref string
	}{
		{name: "unperturbed json", format: client.JSON, right: func(_ *testing.T, b []byte) []byte { return b }},
		{name: "re-encoded json", format: client.JSON, right: func(t *testing.T, b []byte) []byte {
			return editJSON(t, b, func(map[string]any) {})
		}},
		{
			name: "changed value", format: client.JSON, kind: KindValue, path: "subsonic-response/album/year",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { a["year"] = 2002 })
			},
		},
		{
			name: "number literal changed, float64 value kept", format: client.JSON, kind: KindValue,
			path: "subsonic-response/album/songCount",
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"songCount":3,`, `"songCount":3.0,`)
			},
		},
		{
			name: "missing field", format: client.JSON, kind: KindPresence, path: "subsonic-response/album/sortName",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { delete(a, "sortName") })
			},
		},
		{
			name: "extra field", format: client.JSON, kind: KindPresence, path: "subsonic-response/album/extra",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { a["extra"] = "x" })
			},
		},
		{
			name: "string for number", format: client.JSON, kind: KindType, path: "subsonic-response/album/songCount",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { a["songCount"] = "3" })
			},
		},
		{
			name: "singleton for array", format: client.JSON, kind: KindType, path: "subsonic-response/album/artists",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					list, _ := a["artists"].([]any)
					a["artists"] = list[0]
				})
			},
		},
		{
			name: "reordered songs", format: client.JSON, kind: KindOrder, path: "subsonic-response/album/song",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					list, _ := a["song"].([]any)
					list[0], list[1] = list[1], list[0]
				})
			},
		},
		{
			name: "an id that pairs twice", format: client.JSON, kind: KindIdentity, path: "subsonic-response/album/song/0/id",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					s := songs(a)
					s[1]["id"] = s[0]["id"]
				})
			},
		},
		{
			name: "an album the manifest lacks", format: client.JSON, kind: KindIdentity, path: "subsonic-response/album",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { a["name"] = "Nowhere" })
			},
		},
		{
			name: "an album against an empty manifest", format: client.JSON, empty: true,
			kind: KindIdentity, path: "subsonic-response/album",
			right: func(_ *testing.T, b []byte) []byte { return b },
		},
		{
			name: "song removed", format: client.JSON, kind: KindPresence, path: "subsonic-response/album/song/2",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					list, _ := a["song"].([]any)
					a["song"] = list[:2]
				})
			},
		},
		{
			name: "label added", format: client.JSON, kind: KindPresence, path: "subsonic-response/album/recordLabels/0",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { a["recordLabels"] = []any{map[string]any{"name": "Extra"}} })
			},
		},
		{
			name: "song duplicated", format: client.JSON, kind: KindIdentity, path: "subsonic-response/album/song",
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					list, _ := a["song"].([]any)
					a["song"] = append(list, list[0])
				})
			},
		},
		{
			name: "unordered array reordered", format: client.JSON,
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) {
					artists, _ := songs(a)[2]["artists"].([]any)
					artists[0], artists[1] = artists[1], artists[0]
				})
			},
		},
		{
			name: "an album id neither source pairs", format: client.JSON, kind: KindIdentity,
			path: "subsonic-response/album/song/0/albumId",
			left: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { songs(a)[0]["albumId"] = "u1" })
			},
			right: func(t *testing.T, b []byte) []byte {
				return editJSON(t, b, func(a map[string]any) { songs(a)[0]["albumId"] = "u2" })
			},
		},
		{name: "unperturbed xml", format: client.XML, right: func(_ *testing.T, b []byte) []byte { return b }},
		{
			name: "self-closing element", format: client.XML,
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `<releaseDate year="2001"></releaseDate>`, `<releaseDate year="2001"/>`)
			},
		},
		{
			name: "attribute moved to a child element", format: client.XML, kind: KindType, path: "subsonic-response/album/name",
			right: func(t *testing.T, b []byte) []byte {
				b = replaceOnce(t, b, ` name="Single Disc"`, "")
				return replaceOnce(t, b, `<releaseDate year="2001">`, `<name>Single Disc</name><releaseDate year="2001">`)
			},
		},
		{
			name: "serverVersion differs by value", format: client.JSON,
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"serverVersion":"0.64.2 (10114574)"`, `"serverVersion":"9.9.9"`)
			},
		},
		{
			name: "type differs by value", format: client.JSON,
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"type":"navidrome"`, `"type":"gonic"`)
			},
		},
		{
			name: "serverVersion removed", format: client.JSON, kind: KindPresence, path: "subsonic-response/serverVersion",
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"serverVersion":"0.64.2 (10114574)",`, "")
			},
		},
		{
			name: "serverVersion becomes a number", format: client.JSON, kind: KindType, path: "subsonic-response/serverVersion",
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"serverVersion":"0.64.2 (10114574)"`, `"serverVersion":9`)
			},
		},
		{
			name: "version still differs by value", format: client.JSON, kind: KindValue, path: "subsonic-response/version",
			right: func(t *testing.T, b []byte) []byte {
				return replaceOnce(t, b, `"version":"1.16.1"`, `"version":"1.15.0"`)
			},
		},
		{
			name: "xml serverVersion and type differ by value", format: client.XML,
			right: func(t *testing.T, b []byte) []byte {
				b = replaceOnce(t, b, `serverVersion="0.64.2 (10114574)"`, `serverVersion="9.9.9"`)
				return replaceOnce(t, b, `type="navidrome"`, `type="gonic"`)
			},
		},
		{
			name: "failed envelope, messages differ", format: client.JSON,
			left:  func(*testing.T, []byte) []byte { return failedEnvelope("a", "Data not found", 70) },
			right: func(*testing.T, []byte) []byte { return failedEnvelope("b", "not found", 70) },
		},
		{
			name: "failed envelope, codes differ", format: client.JSON, kind: KindValue, path: "subsonic-response/error/code",
			left:  func(*testing.T, []byte) []byte { return failedEnvelope("a", "m", 70) },
			right: func(*testing.T, []byte) []byte { return failedEnvelope("a", "m", 40) },
		},
		{
			name: "failed envelope, message absent", format: client.JSON, kind: KindPresence, path: "subsonic-response/error/message",
			left: func(*testing.T, []byte) []byte { return failedEnvelope("a", "m", 70) },
			right: func(*testing.T, []byte) []byte {
				return []byte(`{"subsonic-response":{"status":"failed","version":"1.16.1","type":"a","serverVersion":"a-1","openSubsonic":true,"error":{"code":70}}}`)
			},
		},
		{
			name: "null against an empty object", format: client.JSON, kind: KindType, path: "subsonic-response/album/song/0/replayGain",
			from: gonicMixtape,
			right: func(t *testing.T, b []byte) []byte {
				return replaceFirst(t, b, `"replayGain":null`, `"replayGain":{}`)
			},
		},
		{
			name: "null against an empty array", format: client.JSON, kind: KindType, path: "subsonic-response/album/song/0/replayGain",
			from: gonicMixtape,
			right: func(t *testing.T, b []byte) []byte {
				return replaceFirst(t, b, `"replayGain":null`, `"replayGain":[]`)
			},
		},
		{
			name: "null against an absent member", format: client.JSON, kind: KindPresence, path: "subsonic-response/album/song/0/replayGain",
			from: gonicMixtape,
			right: func(t *testing.T, b []byte) []byte {
				return replaceFirst(t, b, `,"replayGain":null`, "")
			},
		},
		{
			name: "an empty object against an empty list", format: client.JSON, kind: KindPresence, path: "subsonic-response/playlists/playlist",
			left:  func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{}`) },
			right: func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{"playlist":[]}`) },
		},
		{
			name: "an empty string against no element", format: client.JSON, kind: KindPresence,
			path: "subsonic-response/artist/album/0/releaseTypes/0",
			from: gonicArtist, endpoint: "getArtist", ref: "artist:Delta Duo",
			right: func(t *testing.T, b []byte) []byte {
				return replaceFirst(t, b, `"releaseTypes":[""]`, `"releaseTypes":[]`)
			},
		},
		{
			name: "json escape against a literal ampersand", format: client.JSON,
			left:  func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{"note":"R \u0026 B"}`) },
			right: func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{"note":"R & B"}`) },
		},
		{
			name: "a different string is still a value finding", format: client.JSON, kind: KindValue,
			path:  "subsonic-response/playlists/note",
			left:  func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{"note":"R \u0026 B"}`) },
			right: func(*testing.T, []byte) []byte { return jsonEnvelope(`,"playlists":{"note":"R and B"}`) },
		},
		{
			name: "xml character reference against entity", format: client.XML,
			left: func(*testing.T, []byte) []byte {
				return xmlEnvelope(`<playlists note="say &#34;hi&#34;"/>`)
			},
			right: func(*testing.T, []byte) []byte {
				return xmlEnvelope(`<playlists note="say &quot;hi&quot;"/>`)
			},
		},
		{
			name: "xml declaration present against absent", format: client.XML,
			left: func(*testing.T, []byte) []byte {
				return append([]byte(`<?xml version="1.0" encoding="UTF-8"?>`+"\n"), xmlEnvelope(`<playlists/>`)...)
			},
			right: func(*testing.T, []byte) []byte { return xmlEnvelope(`<playlists/>`) },
		},
		{
			name: "xml prefix against the default namespace", format: client.XML,
			left: func(*testing.T, []byte) []byte {
				return []byte(`<s:subsonic-response xmlns:s="http://subsonic.org/restapi" status="ok" version="1.16.1" type="a"><s:playlists/></s:subsonic-response>`)
			},
			right: func(*testing.T, []byte) []byte { return xmlEnvelope(`<playlists/>`) },
		},
		{
			name: "xml singleton against an array stays one finding", format: client.XML, kind: KindPresence,
			path: "subsonic-response/playlists/playlist/1",
			left: func(*testing.T, []byte) []byte {
				return xmlEnvelope(`<playlists><playlist name="x"/></playlists>`)
			},
			right: func(*testing.T, []byte) []byte {
				return xmlEnvelope(`<playlists><playlist name="x"/><playlist name="y"/></playlists>`)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := jsonBody
			if tt.format == client.XML {
				left = xmlBody
			}
			if tt.from != "" {
				left = readBody(t, tt.from)
			}
			lib := testLibrary(t)
			if tt.empty {
				lib = newLibrary(nil, nil)
			}
			recorded := left
			if tt.left != nil {
				left = tt.left(t, recorded)
			}
			endpoint, ref := cmp.Or(tt.endpoint, "getAlbum"), cmp.Or(tt.ref, "album:Alpha Artist/Single Disc")
			got := diffEndpoint(t, lib, endpoint, ref, tt.format, left, tt.right(t, recorded))
			if tt.kind == "" {
				if len(got) != 0 {
					t.Fatalf("got %d findings, want none: %+v", len(got), got)
				}
				return
			}
			if len(got) != 1 || got[0].Kind != tt.kind || got[0].Path != tt.path {
				t.Fatalf("got %+v\nwant one %s finding at %s", got, tt.kind, tt.path)
			}
		})
	}
}

// TestDifferResolvesEntities edits both sources' bodies alike, so the only
// finding a row can yield is an entity the manifest cannot place.
func TestDifferResolvesEntities(t *testing.T) {
	body := readBody(t, albumJSON)
	tests := []struct {
		name string
		edit func(album map[string]any)
		want []string // "kind path" of every finding
	}{
		{
			name: "a song on an album two artists share, named by neither",
			edit: func(a map[string]any) {
				s := songs(a)[0]
				s["album"] = "Greatest Hits"
				delete(s, "displayAlbumArtist")
				delete(s, "albumArtists")
				delete(s, "discNumber") // both albums' track 1 then agree on disc 0
			},
			want: []string{"identity subsonic-response/album/song/0"},
		},
		{
			name: "an artist the manifest lacks",
			edit: func(a map[string]any) {
				artists, _ := a["artists"].([]any)
				artist, _ := artists[0].(map[string]any)
				artist["name"] = "Nobody"
			},
			want: []string{"identity subsonic-response/album/artists/0"},
		},
		{
			name: "the same song named by its album artist",
			edit: func(a map[string]any) {
				songs(a)[0]["album"] = "Greatest Hits"
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			edited := editJSON(t, body, tt.edit)
			var got []string
			for _, f := range diffBodies(t, client.JSON, edited, edited) {
				got = append(got, string(f.Kind)+" "+f.Path)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("findings %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		name    string
		format  client.Format
		body    string
		path    string
		want    string // the node's rendering at path
		wantErr string
	}{
		{name: "number keeps its literal", format: client.JSON, body: `{"a":2.50}`, path: "a", want: "2.50"},
		{name: "string", format: client.JSON, body: `{"a":"2"}`, path: "a", want: `"2"`},
		{name: "attribute", format: client.XML, body: `<r a="1"/>`, path: "r/a", want: `attribute "1"`},
		{name: "text element", format: client.XML, body: `<r><a>1</a></r>`, path: "r/a", want: `element "1"`},
		{name: "repeated element is an array", format: client.XML, body: `<r><a x="1"/><a x="2"/></r>`, path: "r/a", want: "array of 2"},
		{name: "namespace declaration dropped", format: client.XML, body: `<r xmlns="u" b="1"/>`, path: "r/xmlns", want: "absent"},
		{name: "layout whitespace dropped", format: client.XML, body: "<r>\n  <a x=\"1\"/>\n</r>", path: "r/value", want: "absent"},
		{name: "empty", format: client.JSON, body: " ", wantErr: "empty body"},
		{name: "malformed json", format: client.JSON, body: `{"a":`, wantErr: "EOF"},
		{name: "trailing json", format: client.JSON, body: `{} {}`, wantErr: "after the top-level value"},
		{name: "duplicate json member", format: client.JSON, body: `{"a":1,"a":2}`, wantErr: "duplicate"},
		{name: "attribute and child share a name", format: client.XML, body: `<r a="1"><a/></r>`, wantErr: "both named a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := decode([]byte(tt.body), tt.format)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				if strings.Contains(err.Error(), "%!") {
					t.Errorf("err %q carries a formatting verb error", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got := at(n, tt.path).render(); got != tt.want {
				t.Errorf("at %s: %s, want %s", tt.path, got, tt.want)
			}
		})
	}
}

// TestXMLElementPaths pins the path of every element by ordinal: a
// repeated name is indexed from zero, a single child is not, and indexes
// count only the siblings of that name.
func TestXMLElementPaths(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		want    []string
		wantErr string
	}{
		{"root only", `<r/>`, []string{"r"}, ""},
		{
			"single child is not indexed", `<r><a><b/></a></r>`,
			[]string{"r", "r/a", "r/a/b"}, "",
		},
		{
			"repeated name is indexed", `<r><a/><a/></r>`,
			[]string{"r", "r/a/0", "r/a/1"}, "",
		},
		{
			"other siblings do not shift the index", `<r><a/><x/><a><b/></a></r>`,
			[]string{"r", "r/a/0", "r/x", "r/a/1", "r/a/1/b"}, "",
		},
		{
			"document order is depth first", `<r><a><c/><c/></a><b><c/></b></r>`,
			[]string{"r", "r/a", "r/a/c/0", "r/a/c/1", "r/b", "r/b/c"}, "",
		},
		{"malformed", `<r><a></b></r>`, nil, "closed by"},
		{"no element", ` `, nil, "no root element"},
		{"second root", `<r/><s/>`, nil, "second root"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := xmlElementPaths([]byte(tt.body))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("err = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || !slices.Equal(got, tt.want) {
				t.Errorf("paths = %q, %v; want %q", got, err, tt.want)
			}
		})
	}
}
