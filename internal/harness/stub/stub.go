// Package stub replays a [cassette.Case]'s captured MusicBrainz
// responses over a real loopback HTTP server, so pinned beets can point
// its musicbrainz.host at a fixed, offline fixture set instead of the
// live API. A request the case's log does not cover gets a 404 and is
// recorded, never silently made up.
package stub

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
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
// It answers from the case's committed request log, which is an
// expectation: a request the log does not cover is a [Miss], and a
// conformance run leaving a non-empty miss list fails.
type Server struct {
	listener  net.Listener
	http      *http.Server
	responses map[key][]byte

	// log is the case's request list verbatim, kept beside the
	// response map because [Server.Unused] reports in that list's own
	// order and a map cannot carry it.
	log []cassette.Request

	mu     sync.Mutex
	misses []Miss
	hit    map[key]bool
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

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("stub: listen: %w", err)
	}

	s := &Server{
		listener:  ln,
		responses: responses,
		log:       slices.Clone(c.Requests),
		hit:       make(map[key]bool, len(c.Requests)),
	}
	s.http = &http.Server{Handler: s}
	go func() {
		// http.Server.Serve returns ErrServerClosed once Close stops the
		// listener; that is the expected exit, not a failure to report.
		_ = s.http.Serve(ln)
	}()
	return s, nil
}

// ServeHTTP replays the response filed under the request's canonical
// path+query, or records a [Miss] and answers 404 if the log lacks it.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathQuery := cassette.CanonicalizePathQuery(r.URL.EscapedPath() + "?" + r.URL.RawQuery)
	k := key{method: r.Method, pathQuery: pathQuery}

	if body, ok := s.responses[k]; ok {
		s.recordHit(k)
		writeResponse(w, body)
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

func (s *Server) recordHit(k key) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hit[k] = true
}

// Misses returns every unmatched request the server has received so
// far, in the order it received them.
func (s *Server) Misses() []Miss {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.misses)
}

// Unused returns the logged requests the run never issued, the
// complement of [Server.Misses], in the case's own log order.
//
// Misses alone cannot catch a client that issues FEWER requests than
// the case's log records: an unissued request leaves no trace at all,
// so a truncated flow passes a zero-miss check silently. A
// conformance run therefore fails on a non-empty result here, because
// a run that skipped a lookup took a different flow than pinned beets
// took.
func (s *Server) Unused() []cassette.Request {
	s.mu.Lock()
	defer s.mu.Unlock()
	var unused []cassette.Request
	for _, req := range s.log {
		if !s.hit[key{method: req.Method, pathQuery: req.PathQuery}] {
			unused = append(unused, req)
		}
	}
	return unused
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
