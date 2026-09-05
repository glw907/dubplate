// Package stub replays a [cassette.Case]'s captured MusicBrainz
// responses over a real loopback HTTP server, so pinned beets can point
// its musicbrainz.host at a fixed, offline fixture set instead of the
// live API. A request the case's log does not cover gets a 404 and is
// recorded, never silently made up.
//
// A [cassette.Case.Extra] entry is answered too, but never counted as a
// miss: it is a recall-extension fixture (U8), a response the case
// captured that no logged request consumes. Because an Extra entry is
// by construction a response the current flow does not request,
// answering from it never changes what a caller pointed at
// [github.com/glw907/dubplate/internal/harness/oracle.Derive] receives,
// so the second map is inert for every derivation this pass can run:
// zero of the corpus's committed cases carry an extra key.
package stub

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// key identifies one logged request the way the incoming HTTP request is
// matched against it: method plus the canonical path+query form.
type key struct {
	method    string
	pathQuery string
}

// Miss is one request the case's log did not cover: the method and
// canonical path+query the server actually received, in the shape a
// follow-up capture can act on directly. JSON tags match this
// package's on-disk miss-file contract, documented in
// corpus/README.md: [internal/harness/oracle.Derive] writes a case's
// misses in this shape, and `harness capture --case <dir>` decodes
// them straight into this type.
type Miss struct {
	Method    string `json:"method"`
	PathQuery string `json:"path_query"`
}

// Server replays one [cassette.Case] over HTTP on an ephemeral loopback
// port. Close it when done; it does not stop itself.
//
// It answers from two maps. The request map is the case's committed
// log, which is an expectation: a request it does not cover is a
// [Miss], and a conformance run leaving a non-empty miss list fails.
// The extra map is the case's recall-extension fixtures
// ([cassette.Case.Extra], U8): answering from it is neither a miss nor
// a failure, and every such answer is reported through [Server.Consumed]
// so a future candidate-widening improvement is graded on the shared
// subset rather than escaping the harness (rung 3 spec, the B2 fold).
// The request map is consulted first; [cassette.LoadCase] rejects a
// case whose two maps would share a key, so the order is a stated rule
// rather than a tiebreak.
type Server struct {
	listener  net.Listener
	http      *http.Server
	responses map[key][]byte
	extra     map[key]extraEntry

	mu       sync.Mutex
	misses   []Miss
	consumed []cassette.Request
}

// extraEntry is one recall-extension response and the log entry that
// names it, kept together so [Server.Consumed] reports the entry
// rather than a reconstructed shape.
type extraEntry struct {
	request cassette.Request
	body    []byte
}

// New reads c's referenced response files from dir (the case's own
// directory, the same one [cassette.LoadCase] was given) and starts
// serving them on 127.0.0.1 at an OS-assigned port.
func New(c *cassette.Case, dir string) (*Server, error) {
	responses := make(map[key][]byte, len(c.Requests))
	for _, req := range c.Requests {
		body, err := os.ReadFile(filepath.Join(dir, req.Response))
		if err != nil {
			return nil, fmt.Errorf("stub: read response %q: %w", req.Response, err)
		}
		responses[key{method: req.Method, pathQuery: req.PathQuery}] = body
	}

	extra := make(map[key]extraEntry, len(c.Extra))
	for _, req := range c.Extra {
		body, err := os.ReadFile(filepath.Join(dir, req.Response))
		if err != nil {
			return nil, fmt.Errorf("stub: read extra response %q: %w", req.Response, err)
		}
		extra[key{method: req.Method, pathQuery: req.PathQuery}] = extraEntry{request: req, body: body}
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("stub: listen: %w", err)
	}

	s := &Server{listener: ln, responses: responses, extra: extra}
	s.http = &http.Server{Handler: s}
	go func() {
		// http.Server.Serve returns ErrServerClosed once Close stops the
		// listener; that is the expected exit, not a failure to report.
		_ = s.http.Serve(ln)
	}()
	return s, nil
}

// ServeHTTP replays the response filed under the request's canonical
// path+query: from the request map first, then the extra map, or
// records a [Miss] and answers 404 if neither carries it.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathQuery := canonicalize(r.URL.EscapedPath(), r.URL.RawQuery)
	k := key{method: r.Method, pathQuery: pathQuery}

	if body, ok := s.responses[k]; ok {
		writeResponse(w, body)
		return
	}
	if entry, ok := s.extra[k]; ok {
		s.recordConsumed(entry.request)
		writeResponse(w, entry.body)
		return
	}
	s.recordMiss(r.Method, pathQuery)
	w.WriteHeader(http.StatusNotFound)
}

func writeResponse(w http.ResponseWriter, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.Write(body) //nolint:errcheck // best-effort write to a replay stub; the client sees a truncated body either way
}

func (s *Server) recordMiss(method, pathQuery string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.misses = append(s.misses, Miss{Method: method, PathQuery: pathQuery})
}

func (s *Server) recordConsumed(req cassette.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consumed = append(s.consumed, req)
}

// Misses returns every unmatched request the server has received so
// far, in the order it received them.
func (s *Server) Misses() []Miss {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.misses)
}

// Consumed returns every [cassette.Case.Extra] entry the server has
// answered so far, in the order it answered them. It is deliberately
// distinct from [Server.Misses]: an Extra hit is a reportable event,
// never a failure.
func (s *Server) Consumed() []cassette.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.consumed)
}

// Addr returns the server's host:port, suitable for a beets
// musicbrainz.host override.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Close shuts the server down and releases its port.
func (s *Server) Close() error {
	return s.http.Close()
}

// canonicalize mirrors cassette's canonicalization rule, documented in
// corpus/README.md: sort the raw, percent-encoded query's "&"-joined
// key=value pairs byte-wise; the path, including any trailing slash,
// and each pair's own encoding are left untouched. path must be the
// request's raw, percent-encoded form (r.URL.EscapedPath(), never the
// already-decoded r.URL.Path): a case logs the wire bytes it captured,
// and Go's request parsing decodes Path eagerly, so keying on it would
// silently fail to match a request whose path segment needed encoding.
// Duplicated here rather than called from cassette because Task 1's
// cassette package ships frozen and keeps the rule unexported.
func canonicalize(path, rawQuery string) string {
	if rawQuery == "" {
		return path
	}
	pairs := strings.Split(rawQuery, "&")
	slices.Sort(pairs)
	return path + "?" + strings.Join(pairs, "&")
}
