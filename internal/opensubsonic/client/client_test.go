package client

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

const (
	okJSON     = `{"subsonic-response":{"status":"ok","version":"1.16.1","type":"navidrome","serverVersion":"0.64.2"}}`
	failedJSON = `{"subsonic-response":{"status":"failed","version":"1.16.1","error":{"code":40,"message":"Wrong username or password"}}}`
	okXML      = `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<subsonic-response xmlns="http://subsonic.org/restapi" status="ok" version="1.16.1" type="gonic" serverVersion="v0.22.0" openSubsonic="true"></subsonic-response>`
	failedXML = `<subsonic-response xmlns="http://subsonic.org/restapi" status="failed" version="1.16.1"><error code="70" message="not found"></error></subsonic-response>`
)

func TestToken(t *testing.T) {
	// The worked vector from the legacy token authentication section of the OpenSubsonic docs.
	if got, want := token("sesame", "c19b2d"), "26719a1196d2a940705a59634eb18eab"; got != want {
		t.Errorf("token = %s, want %s", got, want)
	}
	if token("sesame", "c19b2d") == token("c19b2d", "sesame") {
		t.Error("token ignores argument order")
	}
}

type recorded struct {
	query url.Values
	path  string
}

func serve(t *testing.T, status int, ctype, body string) (*Client, *[]recorded) {
	t.Helper()
	var got []recorded
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, recorded{query: r.URL.Query(), path: r.URL.Path})
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return &Client{BaseURL: srv.URL + "/sub", User: "alice", Password: "sesame", Name: "harness", HTTP: srv.Client()}, &got
}

func TestDoRequest(t *testing.T) {
	c, got := serve(t, 200, "application/json", okJSON)
	params := url.Values{"id": {"al-1"}}
	for range 2 {
		if _, err := c.Do(context.Background(), "getAlbum", params, JSON); err != nil {
			t.Fatal(err)
		}
	}
	if len(*got) != 2 {
		t.Fatalf("server saw %d requests, want 2", len(*got))
	}
	first, second := (*got)[0], (*got)[1]
	if first.path != "/sub/rest/getAlbum" {
		t.Errorf("path = %q, want /sub/rest/getAlbum", first.path)
	}
	q := first.query
	for k, want := range map[string]string{"u": "alice", "v": apiVersion, "c": "harness", "f": "json", "id": "al-1"} {
		if q.Get(k) != want {
			t.Errorf("%s = %q, want %q", k, q.Get(k), want)
		}
	}
	if q.Has("p") {
		t.Error("request carries the plaintext password")
	}
	salt := q.Get("s")
	if len(salt) < 6 {
		t.Errorf("salt %q shorter than six characters", salt)
	}
	sum := md5.Sum([]byte("sesame" + salt))
	if q.Get("t") != hex.EncodeToString(sum[:]) {
		t.Errorf("t = %q does not match md5(password+salt)", q.Get("t"))
	}
	if salt == second.query.Get("s") {
		t.Errorf("two requests share the salt %q", salt)
	}
	if params.Has("u") {
		t.Error("Do mutated the caller's params")
	}
}

func TestDoXMLFormat(t *testing.T) {
	c, got := serve(t, 200, "application/xml", okXML)
	if _, err := c.Do(context.Background(), "ping", nil, XML); err != nil {
		t.Fatal(err)
	}
	if f := (*got)[0].query.Get("f"); f != "xml" {
		t.Errorf("f = %q, want xml", f)
	}
}

func TestDoEnvelope(t *testing.T) {
	tests := []struct {
		name    string
		format  Format
		status  int
		body    string
		want    Envelope
		wantErr func(error) bool
		errText string
	}{
		{name: "json ok", format: JSON, status: 200, body: okJSON,
			want: Envelope{Status: "ok", ServerVersion: "0.64.2"}},
		{name: "xml ok attribute heavy", format: XML, status: 200, body: okXML,
			want: Envelope{Status: "ok", ServerVersion: "v0.22.0"}},
		{name: "json failed", format: JSON, status: 200, body: failedJSON,
			want: Envelope{Status: "failed", Error: &Error{Code: 40, Message: "Wrong username or password"}},
			wantErr: func(err error) bool {
				var e *Error
				return errors.As(err, &e) && e.Code == 40 && e.Message == "Wrong username or password"
			}},
		{name: "xml failed", format: XML, status: 200, body: failedXML,
			want: Envelope{Status: "failed", Error: &Error{Code: 70, Message: "not found"}},
			wantErr: func(err error) bool {
				var e *Error
				return errors.As(err, &e) && e.Code == 70 && e.Message == "not found"
			}},
		{name: "non-200", format: JSON, status: 503, body: okJSON,
			wantErr: func(err error) bool { var e *statusError; return errors.As(err, &e) && e.code == 503 }},
		{name: "empty body", format: JSON, status: 200, body: "",
			wantErr: func(err error) bool { return errors.Is(err, errEmptyResponse) }, errText: "empty response"},
		{name: "json without wrapper", format: JSON, status: 200, body: `{"status":"ok","version":"1.16.1"}`,
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) && e.format == JSON }, errText: `json`},
		{name: "json not json", format: JSON, status: 200, body: "<html>oops</html>",
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) }, errText: "<html>oops"},
		{name: "xml wrong root", format: XML, status: 200, body: `<html status="ok"></html>`,
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) && e.format == XML }, errText: "xml"},
		{name: "xml given json", format: XML, status: 200, body: okJSON,
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) }},
		{name: "unknown status", format: JSON, status: 200, body: `{"subsonic-response":{"status":"weird"}}`,
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) }},
		{name: "failed without error member", format: JSON, status: 200, body: `{"subsonic-response":{"status":"failed"}}`,
			wantErr: func(err error) bool { var e *malformedError; return errors.As(err, &e) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := serve(t, tt.status, "text/plain", tt.body)
			resp, err := c.Do(context.Background(), "ping", nil, tt.format)
			if (err != nil) != (tt.wantErr != nil) {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr != nil)
			}
			if tt.wantErr != nil {
				if !tt.wantErr(err) {
					t.Errorf("error %v does not match the expected type", err)
				}
				if tt.errText != "" && !strings.Contains(err.Error(), tt.errText) {
					t.Errorf("error %q lacks %q", err, tt.errText)
				}
			}
			if resp == nil {
				t.Fatal("nil response; the verbatim body is owed even on error")
			}
			if resp.StatusCode != tt.status || string(resp.Body) != tt.body {
				t.Errorf("response = %d %q, want %d %q", resp.StatusCode, resp.Body, tt.status, tt.body)
			}
			if resp.Header.Get("Content-Type") != "text/plain" {
				t.Errorf("headers not returned: %v", resp.Header)
			}
			if tt.want.Status != "" {
				got := resp.Envelope
				if got.Status != tt.want.Status || got.ServerVersion != tt.want.ServerVersion {
					t.Errorf("envelope = %+v, want %+v", got, tt.want)
				}
				if (got.Error == nil) != (tt.want.Error == nil) || (got.Error != nil && *got.Error != *tt.want.Error) {
					t.Errorf("envelope error = %v, want %v", got.Error, tt.want.Error)
				}
			}
		})
	}
}

func TestDoBadBaseURL(t *testing.T) {
	c := &Client{BaseURL: "://bad", HTTP: http.DefaultClient}
	if _, err := c.Do(context.Background(), "ping", nil, JSON); err == nil {
		t.Fatal("Do accepted an unparsable base URL")
	}
}

func TestDoTransportErrorOmitsCredentials(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	base := srv.URL
	srv.Close()
	c := &Client{BaseURL: base, User: "alice", Password: "sesame", Name: "harness", HTTP: http.DefaultClient}
	_, err := c.Do(context.Background(), "ping", nil, JSON)
	if err == nil {
		t.Fatal("Do succeeded against a closed server")
	}
	for _, leak := range []string{"t=", "s=", "u=alice"} {
		if strings.Contains(err.Error(), leak) {
			t.Errorf("transport error %q leaks %q", err, leak)
		}
	}
}

func TestDoNilHTTPUsesDefaultClient(t *testing.T) {
	c, got := serve(t, 200, "application/json", okJSON)
	c.HTTP = nil
	if _, err := c.Do(context.Background(), "ping", nil, JSON); err != nil {
		t.Fatal(err)
	}
	if len(*got) != 1 {
		t.Errorf("server saw %d requests, want 1", len(*got))
	}
}

func TestDoRefusesUnknownFormatUnsent(t *testing.T) {
	tests := []struct {
		name   string
		format Format
	}{
		{"empty", ""},
		{"unknown", "yaml"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, got := serve(t, 200, "application/json", okJSON)
			resp, err := c.Do(context.Background(), "ping", nil, tt.format)
			if err == nil || !strings.Contains(err.Error(), "unknown format") {
				t.Fatalf("err = %v, want an unknown format refusal", err)
			}
			if resp != nil || len(*got) != 0 {
				t.Errorf("response %v and %d requests sent, want none", resp, len(*got))
			}
		})
	}
}
