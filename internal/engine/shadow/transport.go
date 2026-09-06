package shadow

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// recordingTransport records every request's canonical path and query
// and its response body, in the order tag.Album issues them, so an
// observation's stored responses are exactly what the run saw. The
// recorded path_query is already in cassette's canonical form by
// construction, because the client builds every query through
// url.Values.Encode, which sorts by key
// (musicbrainz/client.go:183-200).
type recordingTransport struct {
	base     http.RoundTripper
	requests []RecordedRequest
	bodies   [][]byte
}

// RoundTrip issues req through base, buffers the response body with
// io.ReadAll, closes the original, and hands the caller a fresh
// io.NopCloser(bytes.NewReader(body)), so the client's own io.ReadAll
// (client.go:368) sees the same bytes this transport stored.
func (t *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, fmt.Errorf("shadow: recording response body: %w", err)
	}

	t.requests = append(t.requests, RecordedRequest{Method: req.Method, PathQuery: req.URL.RequestURI()})
	t.bodies = append(t.bodies, body)

	resp.Body = io.NopCloser(bytes.NewReader(body))
	return resp, nil
}

// replayTransport answers from an observation's stored responses and
// REFUSES a miss rather than answering 404.
//
// A 404 on a lookup is swallowed inside tag.Album's lookupAndAdd
// (internal/engine/tag/tag.go:166-170), so a miss answered 404 would
// silently drop a candidate and retire a divergence that never actually
// re-ran. The refusal is 501, which doWithRetry's default arm turns
// into a non-retryable error in one attempt
// (musicbrainz/client.go:381-382), and the miss list below is the
// signal callers branch on rather than the error's shape. A silently
// retired divergence is the one outcome worse than an un-retired one
// (rung 4 spec, reevaluate).
type replayTransport struct {
	bodies map[string][]byte // keyed by request URI
	misses []string
}

// RoundTrip answers req from bodies when its request URI is stored, and
// otherwise records the miss and refuses with 501.
func (t *replayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	uri := req.URL.RequestURI()
	body, ok := t.bodies[uri]
	if !ok {
		t.misses = append(t.misses, uri)
		return &http.Response{
			StatusCode: http.StatusNotImplemented,
			Status:     "501 Not Implemented",
			Proto:      req.Proto,
			ProtoMajor: req.ProtoMajor,
			ProtoMinor: req.ProtoMinor,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewReader(nil)),
			Request:    req,
		}, nil
	}
	return &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Proto:      req.Proto,
		ProtoMajor: req.ProtoMajor,
		ProtoMinor: req.ProtoMinor,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(body)),
		Request:    req,
	}, nil
}
