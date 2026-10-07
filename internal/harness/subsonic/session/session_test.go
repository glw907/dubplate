package session

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copyValid copies testdata/valid into a fresh directory.
func copyValid(t *testing.T) string {
	t.Helper()
	dst := t.TempDir()
	entries, err := os.ReadDir("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join("testdata/valid", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dst
}

// edit replaces old with repl in the metadata file of dir; old must be present.
func edit(t *testing.T, dir, old, repl string) {
	t.Helper()
	path := filepath.Join(dir, "session.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), old) {
		t.Fatalf("session.json does not contain %q", old)
	}
	if err := os.WriteFile(path, []byte(strings.ReplaceAll(string(data), old, repl)), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestLoadValid(t *testing.T) {
	s, err := Load("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Sources) != 2 || len(s.Requests) != 2 {
		t.Fatalf("sources %d, requests %d, want 2 and 2", len(s.Sources), len(s.Requests))
	}
	if got := s.Resolutions["album:one"]["gonic"]; got != "al-9" {
		t.Errorf("gonic resolution = %q, want al-9", got)
	}
	want, err := os.ReadFile("testdata/valid/001-navidrome.json")
	if err != nil {
		t.Fatal(err)
	}
	var got []byte
	for _, r := range s.Requests[0].Responses {
		if r.Source == "navidrome" {
			got = r.Body
		}
	}
	if !bytes.Equal(got, want) || !bytes.Contains(got, []byte("A & B <x>")) {
		t.Errorf("body not verbatim:\n%q\nwant\n%q", got, want)
	}
}

func TestLoadRefusals(t *testing.T) {
	tests := []struct {
		name string
		prep func(t *testing.T, dir string)
		want []string // every substring the one error must carry
	}{
		{"missing body", func(t *testing.T, d string) { _ = os.Remove(filepath.Join(d, "001-gonic.json")) }, []string{"read body", "001-gonic.json"}},
		{"orphan body", func(t *testing.T, d string) { _ = os.WriteFile(filepath.Join(d, "stray.json"), nil, 0o644) }, []string{"stray.json", "referenced by no response"}},
		{"auth parameter", func(t *testing.T, d string) { edit(t, d, `"name": "id"`, `"name": "t"`) }, []string{`parameter "t"`, "auth parameter"}},
		{"undeclared source", func(t *testing.T, d string) { edit(t, d, `"source": "gonic"`, `"source": "ghost"`) }, []string{`undeclared source "ghost"`}},
		{"empty serverVersion", func(t *testing.T, d string) {
			edit(t, d, `"server_version": "0.21.0"`, `"server_version": ""`)
		}, []string{`source "gonic": empty serverVersion`}},
		{"empty library digest", func(t *testing.T, d string) {
			edit(t, d, `"library_digest": "lib1"`, `"library_digest": ""`)
		}, []string{"empty library digest"}},
		{"unused resolution key", func(t *testing.T, d string) { edit(t, d, `"album:one": {`, `"album:zzz": {`) }, []string{`resolution key "album:zzz" is used by no request`}},
		{"view suffix", func(t *testing.T, d string) { edit(t, d, `"getArtists"`, `"getArtists.view"`) }, []string{"canonical form"}},
		{"empty requests", func(t *testing.T, d string) {
			_ = os.WriteFile(filepath.Join(d, "session.json"), []byte(`{"schema_version":1,"sources":{},"requests":[]}`), 0o644)
		}, []string{"no requests"}},
		{"unknown member", func(t *testing.T, d string) {
			edit(t, d, `"schema_version": 1,`, `"schema_version": 1, "extra": 1,`)
		}, []string{"extra"}},
		{"malformed", func(t *testing.T, d string) {
			_ = os.WriteFile(filepath.Join(d, "session.json"), []byte(`{"schema_`), 0o644)
		}, []string{"decode", "session.json", "offset"}},
		{"absent metadata", func(t *testing.T, d string) { _ = os.Remove(filepath.Join(d, "session.json")) }, []string{"session.json"}},
		{"version hand-edit", func(t *testing.T, d string) {
			edit(t, d, `"server_version": "0.21.0"`, `"server_version": "0.21.1"`)
		},
			[]string{`request 1 (getArtists): source "gonic": body serverVersion "0.21.0", metadata "0.21.1"`, `request 2 (getAlbum): source "gonic": body serverVersion "0.21.0", metadata "0.21.1"`}},
		{"many bad", func(t *testing.T, d string) {
			_ = os.Remove(filepath.Join(d, "001-gonic.json"))
			_ = os.WriteFile(filepath.Join(d, "stray.json"), nil, 0o644)
			edit(t, d, `"name": "id"`, `"name": "p"`)
			edit(t, d, `"library_digest": "lib1"`, `"library_digest": ""`)
			edit(t, d, `"server_version": "0.99.0 (abc)"`, `"server_version": "0.99.1"`)
		}, []string{"001-gonic.json", "stray.json", `parameter "p"`, "empty library digest", `metadata "0.99.1"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := copyValid(t)
			tt.prep(t, dir)
			_, err := Load(dir)
			if err == nil {
				t.Fatal("Load accepted the session")
			}
			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("error lacks %q:\n%v", w, err)
				}
			}
		})
	}
	t.Run("missing body reports only the read", func(t *testing.T) {
		dir := copyValid(t)
		_ = os.Remove(filepath.Join(dir, "001-gonic.json"))
		_, err := Load(dir)
		if err == nil || strings.Contains(err.Error(), "no readable envelope") {
			t.Errorf("err = %v, want the read error without an envelope error", err)
		}
	})
	t.Run("absent directory", func(t *testing.T) {
		_, err := Load(filepath.Join(t.TempDir(), "nope"))
		if !errors.Is(err, os.ErrNotExist) {
			t.Errorf("err = %v, want not-exist", err)
		}
	})
	t.Run("auth sentinel", func(t *testing.T) {
		dir := copyValid(t)
		edit(t, dir, `"name": "id"`, `"name": "apiKey"`)
		if _, err := Load(dir); !errors.Is(err, errAuthParam) {
			t.Errorf("err = %v, want errAuthParam", err)
		}
	})
}

func TestWriteRefusals(t *testing.T) {
	base := func(t *testing.T) *Session {
		s, err := Load("testdata/valid")
		if err != nil {
			t.Fatal(err)
		}
		return s
	}
	for _, name := range []string{"u", "p", "t", "s", "apiKey"} {
		t.Run("auth "+name, func(t *testing.T) {
			s := base(t)
			s.Requests[1].Params = []Param{{name, "x"}, {"zzz", "album:one"}}
			dir := filepath.Join(t.TempDir(), "out")
			if err := Write(dir, s); !errors.Is(err, errAuthParam) {
				t.Errorf("err = %v, want errAuthParam", err)
			}
			if _, err := os.Stat(dir); err == nil {
				t.Error("refused write created the directory")
			}
		})
	}
	t.Run("duplicate request", func(t *testing.T) {
		s := base(t)
		s.Requests = append(s.Requests, s.Requests[0])
		err := Write(filepath.Join(t.TempDir(), "out"), s)
		if err == nil || !strings.Contains(err.Error(), "duplicate of an earlier request") {
			t.Errorf("err = %v", err)
		}
	})
	t.Run("directory holding a session", func(t *testing.T) {
		dir := filepath.Join(t.TempDir(), "out")
		if err := Write(dir, base(t)); err != nil {
			t.Fatal(err)
		}
		s := base(t)
		s.Requests = s.Requests[:1]
		delete(s.Resolutions, "album:one")
		if err := Write(dir, s); !errors.Is(err, errDirNotEmpty) {
			t.Fatalf("err = %v, want errDirNotEmpty", err)
		}
		if got, err := Load(dir); err != nil || len(got.Requests) != 2 {
			t.Errorf("first session disturbed by the refused write: %v", err)
		}
	})
	t.Run("unsorted params", func(t *testing.T) {
		s := base(t)
		s.Requests[1].Params = []Param{{"id", "album:one"}, {"a", "b"}}
		err := Write(filepath.Join(t.TempDir(), "out"), s)
		if err == nil || !strings.Contains(err.Error(), "not sorted") {
			t.Errorf("err = %v", err)
		}
	})
}

func TestRoundTripBytes(t *testing.T) {
	s, err := Load("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if err := Write(out, s); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(entries) {
		t.Fatalf("wrote %d files, want %d", len(got), len(entries))
	}
	for _, e := range entries {
		want, _ := os.ReadFile(filepath.Join("testdata/valid", e.Name()))
		have, err := os.ReadFile(filepath.Join(out, e.Name()))
		if err != nil || !bytes.Equal(have, want) {
			t.Errorf("%s differs after round trip (err %v)", e.Name(), err)
		}
	}
}

func joinAddition(t *testing.T, s *Session) Addition {
	t.Helper()
	a := Addition{
		Name:        "dubplate",
		Source:      Source{LibraryDigest: "lib-dub", ServerVersion: "dev-1"},
		Resolutions: map[string]string{"album:one": "d-7"},
		Responses:   map[string]Response{},
	}
	for _, req := range s.Requests {
		body := []byte(`{"subsonic-response":{"status":"ok","version":"1.16.1","serverVersion":"dev-1"}}`)
		ct := "application/json"
		if req.Format == "xml" {
			body = []byte(`<subsonic-response status="ok" version="1.16.1" serverVersion="dev-1"/>`)
			ct = "application/xml"
		}
		a.Responses[req.Key()] = Response{Status: 200, ContentType: ct, Body: body}
	}
	return a
}

func TestJoin(t *testing.T) {
	s, err := Load("testdata/valid")
	if err != nil {
		t.Fatal(err)
	}
	joined, err := s.Join(joinAddition(t, s))
	if err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "joined")
	if err := Write(out, joined); err != nil {
		t.Fatal(err)
	}
	got, err := Load(out)
	if err != nil {
		t.Fatal(err)
	}
	if d := got.Sources["dubplate"].LibraryDigest; d != "lib-dub" {
		t.Errorf("joined library digest = %q, want lib-dub", d)
	}
	if id := got.Resolutions["album:one"]["dubplate"]; id != "d-7" {
		t.Errorf("joined resolution = %q, want d-7", id)
	}
	if n := len(got.Requests[1].Responses); n != 3 {
		t.Errorf("request 2 holds %d responses, want 3", n)
	}
	if len(s.Sources) != 2 || len(s.Requests[0].Responses) != 2 {
		t.Error("Join modified the session it was given")
	}

	t.Run("source already present", func(t *testing.T) {
		a := joinAddition(t, s)
		a.Name = "gonic"
		got, err := s.Join(a)
		if !errors.Is(err, errSourceExists) || !strings.Contains(err.Error(), "gonic") {
			t.Errorf("err = %v, want errSourceExists naming gonic", err)
		}
		if got != nil {
			t.Error("refused join returned a session")
		}
	})
	t.Run("request the session lacks", func(t *testing.T) {
		a := joinAddition(t, s)
		a.Responses["getSong.json&id=song%3Aq"] = Response{Status: 200, ContentType: "application/json", Body: []byte("{}")}
		got, err := s.Join(a)
		if !errors.Is(err, errUnknownRequest) || !strings.Contains(err.Error(), "getSong.json&id=song%3Aq") || !strings.Contains(err.Error(), "dubplate") {
			t.Errorf("err = %v, want errUnknownRequest naming source and request", err)
		}
		if got != nil {
			t.Error("refused join returned a session")
		}
	})
}
