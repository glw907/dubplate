package session

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const navidromeArtist = "../data/sessions/seed/007-navidrome.json"

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// leakStrings renders the leaks as scanCredentials reports them.
func leakStrings(leaks []credentialLeak) []string {
	var out []string
	for _, l := range leaks {
		out = append(out, l.String())
	}
	return out
}

func TestScanCredentials(t *testing.T) {
	artist := string(readFile(t, navidromeArtist))
	const image = "?size=600"
	if strings.Count(artist, image) != 1 {
		t.Fatalf("the recorded getArtist body holds %d occurrences of %s, want 1", strings.Count(artist, image), image)
	}
	tests := []struct {
		name   string
		format client.Format
		body   string
		want   []string
	}{
		{name: "a recorded body with a plain query", format: client.JSON, body: artist},
		{
			name: "a recorded body with credentials appended", format: client.JSON,
			body: strings.Replace(artist, image, image+"&u=admin&t=tokenvalue&s=saltvalue", 1),
			want: []string{"subsonic-response/artist/artistImageUrl: URL query carries credential parameter u, t, s"},
		},
		{
			name: "api key", format: client.JSON,
			body: `{"a":{"b":["x","http://h/p?apiKey=k1"]}}`,
			want: []string{"a/b/1: URL query carries credential parameter apiKey"},
		},
		{name: "no query", format: client.JSON, body: `{"a":"http://h/x"}`},
		{name: "names that only end in a credential name", format: client.JSON, body: `{"a":"http://h/x?format=mp3&size=9&dt=1&apiKeys=2&password=3"}`},
		{
			name: "a query the parser rejects", format: client.JSON,
			body: `{"a":"http://h/x?u=a&t=x;s=y"}`,
			want: []string{"a: URL query does not parse carries credential parameter u"},
		},
		{
			name: "a rejected query that names nothing", format: client.JSON,
			body: `{"a":"http://h/x?size=1;s=y"}`,
			want: []string{"a: URL query does not parse"},
		},
		{
			name: "a bad escape in the path", format: client.JSON,
			body: `{"a":"/rest/x%zz?p=secret"}`,
			want: []string{"a: URL query carries credential parameter p"},
		},
		{
			name: "one finding per leaf", format: client.JSON,
			body: `{"a":"/x?u=1","b":"/y?t=2","c":"/z?size=3"}`,
			want: []string{"a: URL query carries credential parameter u", "b: URL query carries credential parameter t"},
		},
		{
			name: "an escaped query separator is still read", format: client.JSON,
			body: `{"a":"http://h/x?s=1"}`,
			want: []string{"a: URL query carries credential parameter s"},
		},
		{
			name: "xml attribute and element text", format: client.XML,
			body: `<r xmlns="u"><img src="/x?u=1"/><note>see /y?p=2</note><ok href="/z?size=1"/></r>`,
			want: []string{"r/img/src: URL query carries credential parameter u", "r/note: URL query carries credential parameter p"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := leakStrings(scanCredentials([]byte(tt.body), tt.format))
			if !slices.Equal(got, tt.want) {
				t.Errorf("leaks\n%q\nwant\n%q", got, tt.want)
			}
			for _, g := range got {
				for _, value := range []string{"admin", "tokenvalue", "saltvalue", "secret"} {
					if strings.Contains(g, value) {
						t.Errorf("leak %q repeats the value %q", g, value)
					}
				}
			}
		})
	}
}

const leakingNote = `"note": "http://h/x?p=secretvalue", `

func TestWriteRefusesBodyCredential(t *testing.T) {
	s, err := Load("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	body := &s.Requests[0].Responses[0].Body
	*body = []byte(strings.Replace(string(*body), `"status": "ok"`, leakingNote+`"status": "ok"`, 1))
	dir := filepath.Join(t.TempDir(), "out")
	err = Write(dir, s)
	if !errors.Is(err, errBodyCredential) {
		t.Fatalf("err = %v, want errBodyCredential", err)
	}
	for _, want := range []string{"request 1 (getArtists)", s.Requests[0].Responses[0].Source, "subsonic-response/note", "parameter p"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "secretvalue") {
		t.Errorf("error repeats the credential: %v", err)
	}
	if _, err := os.Stat(dir); err == nil {
		t.Error("refused write created the directory")
	}
}

func TestLoadRefusesBodyCredential(t *testing.T) {
	dir := copyValid(t)
	path := filepath.Join(dir, "001-navidrome.json")
	body := strings.Replace(string(readFile(t, path)), `"status": "ok"`, leakingNote+`"status": "ok"`, 1)
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err := Load(dir)
	if !errors.Is(err, errBodyCredential) {
		t.Fatalf("err = %v, want errBodyCredential", err)
	}
	for _, want := range []string{dir, "request 1 (getArtists)", `source "navidrome"`, "parameter p"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error lacks %q: %v", want, err)
		}
	}
	if strings.Contains(err.Error(), "secretvalue") {
		t.Errorf("error repeats the credential: %v", err)
	}
}
