// Package client issues OpenSubsonic requests and returns each response
// verbatim beside its parsed envelope.
//
// Behavior cites the OpenSubsonic documentation (https://opensubsonic.netlify.app/docs/):
// "Authentication" for the legacy token scheme, "Response" for the envelope in
// both formats, and "Error handling" for the failed-status form.
package client

import (
	"cmp"
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// apiVersion is the protocol version a request declares in its v parameter.
// The docs state that a server must answer a client whose version it supports.
const apiVersion = "1.16.1"

// Format selects the response encoding, sent as the f parameter.
type Format string

// The two response encodings every endpoint offers.
const (
	JSON Format = "json"
	XML  Format = "xml"
)

// Valid reports whether f is one of the two response encodings.
func (f Format) Valid() bool { return f == JSON || f == XML }

// Client issues requests to one server. BaseURL, User, Password, and Name
// must be set.
type Client struct {
	// BaseURL is the server root. It may carry a sub-path; endpoints are
	// requested at BaseURL + "/rest/" + name.
	BaseURL  string
	User     string
	Password string
	// Name is the client identifier sent as the c parameter.
	Name string
	// HTTP is the client requests go through; nil means http.DefaultClient.
	HTTP *http.Client
}

// Envelope holds the fields of the subsonic-response element a caller reads.
type Envelope struct {
	Status        string // "ok" or "failed"
	ServerVersion string // OpenSubsonic servers report it; others leave it empty
	Error         *Error // set exactly when Status is "failed"
}

// Response is one server answer. Body is the verbatim payload.
type Response struct {
	StatusCode int
	Header     http.Header
	Body       []byte
	Envelope   Envelope
}

// Error is a failed envelope's error element. Do returns it, wrapped in
// nothing, so callers match it with errors.As.
type Error struct {
	Code    int
	Message string
}

func (e *Error) Error() string { return fmt.Sprintf("subsonic error %d: %s", e.Code, e.Message) }

// statusError reports an HTTP status other than 200.
type statusError struct{ code int }

func (e *statusError) Error() string { return fmt.Sprintf("http status %d", e.code) }

// malformedError reports a body that is not a subsonic-response envelope in
// the requested format. prefix holds the first bytes of the body.
type malformedError struct {
	format Format
	prefix string
	reason string
}

func (e *malformedError) Error() string {
	return fmt.Sprintf("malformed %s envelope (%s): %q", e.format, e.reason, e.prefix)
}

// errEmptyResponse reports a 200 answer with no body.
var errEmptyResponse = errors.New("empty response")

// token returns the legacy authentication token: the lowercase hex MD5 of the
// password followed by the salt.
func token(password, salt string) string {
	sum := md5.Sum([]byte(password + salt))
	return hex.EncodeToString(sum[:])
}

// newSalt returns a fresh random salt. The docs ask for at least six
// characters; twelve hex digits from six random bytes meets that.
func newSalt() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read salt: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// Do requests the named endpoint with params in the given format. It never
// modifies params. The returned Response is non-nil whenever the server
// answered, including with a non-200 status, an unparsable body, or a failed
// envelope, so a caller can record the verbatim bytes of an error. A failed
// envelope's error is an *Error. An invalid format is refused before any
// request is sent.
func (c *Client) Do(ctx context.Context, endpoint string, params url.Values, format Format) (*Response, error) {
	if !format.Valid() {
		return nil, fmt.Errorf("request %s: unknown format %q", endpoint, format)
	}
	u, err := url.Parse(c.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}
	u = u.JoinPath("rest", endpoint)
	salt, err := newSalt()
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	for k, vs := range params {
		q[k] = append([]string(nil), vs...)
	}
	q.Set("u", c.User)
	q.Set("t", token(c.Password, salt))
	q.Set("s", salt)
	q.Set("v", apiVersion)
	q.Set("c", c.Name)
	q.Set("f", string(format))
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	hr, err := cmp.Or(c.HTTP, http.DefaultClient).Do(req)
	if err != nil {
		return nil, fmt.Errorf("request %s: %w", endpoint, redactURL(err))
	}
	defer func() { _ = hr.Body.Close() }()
	body, err := io.ReadAll(hr.Body)
	if err != nil {
		return nil, fmt.Errorf("read %s response: %w", endpoint, err)
	}
	resp := &Response{StatusCode: hr.StatusCode, Header: hr.Header, Body: body}
	if hr.StatusCode != http.StatusOK {
		return resp, &statusError{code: hr.StatusCode}
	}
	if len(body) == 0 {
		return resp, errEmptyResponse
	}
	resp.Envelope, err = ParseEnvelope(body, format)
	if err != nil {
		return resp, err
	}
	if resp.Envelope.Error != nil {
		return resp, resp.Envelope.Error
	}
	return resp, nil
}

// redactURL strips the request URL, which carries the token and salt, from a
// transport error.
func redactURL(err error) error {
	if ue, ok := errors.AsType[*url.Error](err); ok {
		return ue.Err
	}
	return err
}

type wireError struct {
	Code    int    `json:"code" xml:"code,attr"`
	Message string `json:"message" xml:"message,attr"`
}

type jsonEnvelope struct {
	Response *struct {
		Status        string     `json:"status"`
		ServerVersion string     `json:"serverVersion"`
		Error         *wireError `json:"error"`
	} `json:"subsonic-response"`
}

type xmlEnvelope struct {
	XMLName       xml.Name   `xml:"subsonic-response"`
	Status        string     `xml:"status,attr"`
	ServerVersion string     `xml:"serverVersion,attr"`
	Error         *wireError `xml:"error"`
}

// ParseEnvelope reads the subsonic-response envelope from body, which must be
// in the given format. It is the parse Do applies to every answer.
func ParseEnvelope(body []byte, format Format) (Envelope, error) {
	malformed := func(reason string) error {
		return &malformedError{format: format, prefix: string(body[:min(len(body), 40)]), reason: reason}
	}
	var env Envelope
	var werr *wireError
	switch format {
	case JSON:
		var je jsonEnvelope
		if err := json.Unmarshal(body, &je); err != nil {
			return Envelope{}, malformed(err.Error())
		}
		if je.Response == nil {
			return Envelope{}, malformed("no subsonic-response member")
		}
		env = Envelope{Status: je.Response.Status, ServerVersion: je.Response.ServerVersion}
		werr = je.Response.Error
	case XML:
		var xe xmlEnvelope
		if err := xml.Unmarshal(body, &xe); err != nil {
			return Envelope{}, malformed(err.Error())
		}
		env = Envelope{Status: xe.Status, ServerVersion: xe.ServerVersion}
		werr = xe.Error
	default:
		return Envelope{}, fmt.Errorf("unknown format %q", format)
	}
	switch env.Status {
	case "ok":
	case "failed":
		if werr == nil {
			return Envelope{}, malformed("failed status without an error element")
		}
		env.Error = &Error{Code: werr.Code, Message: werr.Message}
	default:
		return Envelope{}, malformed(fmt.Sprintf("status %q", env.Status))
	}
	return env, nil
}
