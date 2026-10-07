package grade

import (
	"encoding/json/v2"
	"os"
	"path/filepath"
	"slices"
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
)

func testLibrary(t *testing.T) *library {
	t.Helper()
	lib, err := fixture.Load(committedManifest)
	if err != nil {
		t.Fatal(err)
	}
	return newLibrary(lib.Tracks())
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
	ct := map[client.Format]string{client.JSON: "application/json", client.XML: "application/xml"}[format]
	s := &session.Session{Requests: []session.Request{{
		Endpoint: "getAlbum", Format: format,
		Params: []session.Param{{Name: "id", Value: "album:Alpha Artist/Single Disc"}},
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			left := jsonBody
			if tt.format == client.XML {
				left = xmlBody
			}
			lib := testLibrary(t)
			if tt.empty {
				lib = newLibrary(nil)
			}
			recorded := left
			if tt.left != nil {
				left = tt.left(t, recorded)
			}
			got := diffBodiesIn(t, lib, tt.format, left, tt.right(t, recorded))
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
