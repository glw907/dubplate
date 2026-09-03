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
type Server struct {
	listener  net.Listener
	http      *http.Server
	responses map[key][]byte

	mu     sync.Mutex
	misses []Miss
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

	s := &Server{listener: ln, responses: responses}
	s.http = &http.Server{Handler: s}
	go func() {
		// http.Server.Serve returns ErrServerClosed once Close stops the
		// listener; that is the expected exit, not a failure to report.
		_ = s.http.Serve(ln)
	}()
	return s, nil
}

// ServeHTTP replays the response filed under the request's canonical
// path+query, or records a [Miss] and answers 404.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pathQuery := canonicalize(r.URL.EscapedPath(), r.URL.RawQuery)
	body, ok := s.responses[key{method: r.Method, pathQuery: pathQuery}]
	if !ok {
		s.recordMiss(r.Method, pathQuery)
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(body) //nolint:errcheck // best-effort write to a replay stub; the client sees a truncated body either way
}

func (s *Server) recordMiss(method, pathQuery string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.misses = append(s.misses, Miss{Method: method, PathQuery: pathQuery})
}

// Misses returns every unmatched request the server has received so
// far, in the order it received them.
func (s *Server) Misses() []Miss {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.misses)
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
