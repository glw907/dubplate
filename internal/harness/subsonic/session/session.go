// Package session defines the on-disk record of what several servers
// answered to one ordered list of OpenSubsonic requests.
//
// A session is one directory: session.json plus one body file per
// (request, source) pair. The body files hold each response's bytes
// verbatim and are never decoded and re-encoded, because encoding/json
// compacts and HTML-escapes a json.RawMessage, which would break byte
// identity; session.json names them and carries no body.
//
// A request names its endpoint in canonical form (no ".view" suffix), its
// format, and its non-auth parameters sorted byte-wise, with entity
// references in symbolic form (a fixture manifest key, never a server's
// id); the resolution map records the id each key resolved to on each
// source. No credential parameter is ever written, in a request or in a
// URL a response body holds, so a session commits safely.
//
// A collection: key in the resolution map names a collection by what an
// oracle itself answers. The recorder resolves it by artist-set equality,
// an exact match of the collection's album artists against the artists a
// getArtists call filtered by each music folder lists, never by the
// folder's name or its order.
package session

import (
	"cmp"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const (
	metaName      = "session.json"
	schemaVersion = 1
)

var (
	errAuthParam      = errors.New("auth parameter in request")
	errSourceExists   = errors.New("source already in session")
	errUnknownRequest = errors.New("response to a request the session lacks")
	errDirNotEmpty    = errors.New("session directory is not empty")
	errBodyCredential = errors.New("credential in response body")
)

// authParams are the parameters that carry or select a credential, which
// no session may hold.
var authParams = []string{"u", "p", "t", "s", "apiKey"}

// Session is one loaded or about-to-be-written session.
type Session struct {
	Sources map[string]Source
	// Resolutions maps a symbolic key to the id it resolved to, per source.
	Resolutions map[string]map[string]string
	Requests    []Request
}

// Source records what one server served and who it was.
type Source struct {
	// LibraryDigest is the digest of the fixture library the source served.
	LibraryDigest string `json:"library_digest"`
	// ImageDigest is the container image digest; an oracle has one and
	// dubplate does not.
	ImageDigest string `json:"image_digest,omitempty"`
	// ServerVersion is the server's own envelope serverVersion, which every
	// one of the source's bodies must repeat.
	ServerVersion string `json:"server_version"`
}

// Param is one non-auth request parameter.
type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// Request is one endpoint call and the answers each source gave.
type Request struct {
	Endpoint  string
	Format    client.Format
	Params    []Param // sorted byte-wise by name, then value
	Responses []Response
}

// Response is one source's answer, with its verbatim body.
type Response struct {
	Source      string
	Status      int
	ContentType string
	Body        []byte
}

// Key identifies a request by endpoint, format, and parameters. A request
// appears once per session, and [Session.Join] addresses answers by it.
func (r Request) Key() string {
	var b strings.Builder
	b.WriteString(r.Endpoint + "." + string(r.Format))
	for _, p := range r.Params {
		b.WriteString("&" + url.QueryEscape(p.Name) + "=" + url.QueryEscape(p.Value))
	}
	return b.String()
}

type wireSession struct {
	SchemaVersion int                          `json:"schema_version"`
	Sources       map[string]Source            `json:"sources"`
	Resolutions   map[string]map[string]string `json:"resolutions,omitempty"`
	Requests      []wireRequest                `json:"requests"`
}

type wireRequest struct {
	Endpoint  string         `json:"endpoint"`
	Format    client.Format  `json:"format"`
	Params    []Param        `json:"params,omitempty"`
	Responses []wireResponse `json:"responses"`
}

type wireResponse struct {
	Source      string `json:"source"`
	Status      int    `json:"status"`
	ContentType string `json:"content_type"`
	Body        string `json:"body"`
}

func bodyName(index int, source string, format client.Format) string {
	return fmt.Sprintf("%03d-%s.%s", index+1, source, format)
}

// Load reads and validates the session in dir. Every problem found is
// reported in one error.
func Load(dir string) (*Session, error) {
	raw, err := os.ReadFile(filepath.Join(dir, metaName))
	if err != nil {
		return nil, fmt.Errorf("read session metadata: %w", err)
	}
	var w wireSession
	if err := json.Unmarshal(raw, &w, json.RejectUnknownMembers(true)); err != nil {
		return nil, fmt.Errorf("decode %s: %w", filepath.Join(dir, metaName), err)
	}
	var problems []error
	if w.SchemaVersion != schemaVersion {
		problems = append(problems, fmt.Errorf("schema_version %d, want %d", w.SchemaVersion, schemaVersion))
	}
	s := &Session{Sources: w.Sources, Resolutions: w.Resolutions}
	referenced := map[string]bool{metaName: true}
	unread := map[[2]int]bool{} // (request, response) whose body was not read
	for i, wr := range w.Requests {
		req := Request{Endpoint: wr.Endpoint, Format: wr.Format, Params: wr.Params}
		for j, resp := range wr.Responses {
			r := Response{Source: resp.Source, Status: resp.Status, ContentType: resp.ContentType}
			switch {
			case !filepath.IsLocal(resp.Body) || filepath.Base(resp.Body) != resp.Body:
				problems = append(problems, fmt.Errorf("request %d: body %q is not a plain file name", i+1, resp.Body))
				unread[[2]int{i, j}] = true
			case referenced[resp.Body]:
				problems = append(problems, fmt.Errorf("request %d: body file %s named twice", i+1, resp.Body))
				unread[[2]int{i, j}] = true
			default:
				referenced[resp.Body] = true
				body, err := os.ReadFile(filepath.Join(dir, resp.Body))
				if err != nil {
					problems = append(problems, fmt.Errorf("request %d: read body: %w", i+1, err))
					unread[[2]int{i, j}] = true
				}
				r.Body = body
			}
			req.Responses = append(req.Responses, r)
		}
		s.Requests = append(s.Requests, req)
	}
	problems = append(problems, s.validate(unread)...)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("list session directory: %w", err)
	}
	for _, e := range entries {
		if !referenced[e.Name()] {
			problems = append(problems, fmt.Errorf("file %s is referenced by no response", e.Name()))
		}
	}
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("session %s: %w", dir, err)
	}
	return s, nil
}

// Write validates s and writes it to dir, creating the directory. It
// refuses a dir that already holds any file, since bodies written over
// another session's would sit under that session's metadata. Bodies land
// first and session.json last, each atomically, so a reader sees a
// complete session or no metadata.
func Write(dir string, s *Session) error {
	if err := errors.Join(s.validate(nil)...); err != nil {
		return fmt.Errorf("session: %w", err)
	}
	switch entries, err := os.ReadDir(dir); {
	case err == nil && len(entries) > 0:
		return fmt.Errorf("%w: %s", errDirNotEmpty, dir)
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("check session directory: %w", err)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create session directory: %w", err)
	}
	w := wireSession{SchemaVersion: schemaVersion, Sources: s.Sources, Resolutions: s.Resolutions}
	for i, req := range s.Requests {
		wr := wireRequest{Endpoint: req.Endpoint, Format: req.Format, Params: req.Params, Responses: []wireResponse{}}
		for _, resp := range slices.SortedFunc(slices.Values(req.Responses), func(a, b Response) int {
			return cmp.Compare(a.Source, b.Source)
		}) {
			name := bodyName(i, resp.Source, req.Format)
			if err := atomicfile.WriteFile(filepath.Join(dir, name), resp.Body, 0o644); err != nil {
				return fmt.Errorf("write body: %w", err)
			}
			wr.Responses = append(wr.Responses, wireResponse{Source: resp.Source, Status: resp.Status, ContentType: resp.ContentType, Body: name})
		}
		w.Requests = append(w.Requests, wr)
	}
	raw, err := json.Marshal(w, json.Deterministic(true), jsontext.WithIndent("  "))
	if err != nil {
		return fmt.Errorf("encode session metadata: %w", err)
	}
	if err := atomicfile.WriteFile(filepath.Join(dir, metaName), append(raw, '\n'), 0o644); err != nil {
		return fmt.Errorf("write session metadata: %w", err)
	}
	return nil
}

// Addition is one source joining a session.
type Addition struct {
	Name   string
	Source Source
	// Resolutions maps each symbolic key to the id this source gave it.
	Resolutions map[string]string
	// Responses holds the source's answers by [Request.Key]. The Source
	// field of each is set from Name.
	Responses map[string]Response
}

// Join returns a copy of s with a added. It refuses a source s already
// holds and a response to a request s does not hold, naming each; s itself
// is unchanged. [Write] validates the result.
func (s *Session) Join(a Addition) (*Session, error) {
	if _, ok := s.Sources[a.Name]; ok {
		return nil, fmt.Errorf("join %q: %w", a.Name, errSourceExists)
	}
	byKey := make(map[string]int, len(s.Requests))
	for i, req := range s.Requests {
		byKey[req.Key()] = i
	}
	var problems []error
	for key := range a.Responses {
		if _, ok := byKey[key]; !ok {
			problems = append(problems, fmt.Errorf("source %q, request %s: %w", a.Name, key, errUnknownRequest))
		}
	}
	if err := errors.Join(problems...); err != nil {
		return nil, err
	}
	joined := &Session{
		Sources:     make(map[string]Source, len(s.Sources)+1),
		Resolutions: make(map[string]map[string]string, len(s.Resolutions)),
		Requests:    slices.Clone(s.Requests),
	}
	maps.Copy(joined.Sources, s.Sources)
	joined.Sources[a.Name] = a.Source
	for key, bySource := range s.Resolutions {
		joined.Resolutions[key] = maps.Clone(bySource)
	}
	for key, id := range a.Resolutions {
		if joined.Resolutions[key] == nil {
			joined.Resolutions[key] = map[string]string{}
		}
		joined.Resolutions[key][a.Name] = id
	}
	for key, resp := range a.Responses {
		i := byKey[key]
		resp.Source = a.Name
		joined.Requests[i].Responses = append(slices.Clone(joined.Requests[i].Responses), resp)
	}
	return joined, nil
}

// CheckRequests reports every problem with reqs as a request list: an
// endpoint not in canonical form, an unknown format, parameters unsorted
// or carrying a credential, and a request that repeats an earlier one.
func CheckRequests(reqs []Request) error {
	return errors.Join(checkRequests(reqs)...)
}

func checkRequests(reqs []Request) []error {
	var problems []error
	add := func(format string, args ...any) { problems = append(problems, fmt.Errorf(format, args...)) }
	seen := map[string]bool{}
	for i, req := range reqs {
		at := requestAt(i, req)
		if req.Endpoint == "" || strings.HasSuffix(req.Endpoint, ".view") {
			add("%s: endpoint is not in canonical form", at)
		}
		if !req.Format.Valid() {
			add("%s: unknown format %q", at, req.Format)
		}
		if !slices.IsSortedFunc(req.Params, func(a, b Param) int {
			return cmp.Or(cmp.Compare(a.Name, b.Name), cmp.Compare(a.Value, b.Value))
		}) {
			add("%s: parameters are not sorted", at)
		}
		for _, p := range req.Params {
			if slices.Contains(authParams, p.Name) {
				add("%s: parameter %q: %w", at, p.Name, errAuthParam)
			}
		}
		if key := req.Key(); seen[key] {
			add("%s: duplicate of an earlier request", at)
		} else {
			seen[key] = true
		}
	}
	return problems
}

func requestAt(i int, req Request) string {
	return "request " + strconv.Itoa(i+1) + " (" + req.Endpoint + ")"
}

// validate reports every structural problem in s. A response keyed in
// unread by (request, response) index has a body Load could not read, and
// its envelope goes unchecked so the read error stands alone.
func (s *Session) validate(unread map[[2]int]bool) []error {
	var problems []error
	add := func(format string, args ...any) { problems = append(problems, fmt.Errorf(format, args...)) }
	if len(s.Requests) == 0 {
		add("no requests")
	}
	for _, name := range slices.Sorted(maps.Keys(s.Sources)) {
		src := s.Sources[name]
		if name == "" || strings.ContainsAny(name, "/\\. ") {
			add("source name %q is not a plain name", name)
		}
		if src.LibraryDigest == "" {
			add("source %q: empty library digest", name)
		}
		if src.ServerVersion == "" {
			add("source %q: empty serverVersion", name)
		}
	}
	problems = append(problems, checkRequests(s.Requests)...)
	used := map[string]bool{}
	for i, req := range s.Requests {
		at := requestAt(i, req)
		for _, p := range req.Params {
			used[p.Value] = true
		}
		answered := map[string]bool{}
		for j, resp := range req.Responses {
			src, declared := s.Sources[resp.Source]
			switch {
			case !declared:
				add("%s: response from undeclared source %q", at, resp.Source)
			case answered[resp.Source]:
				add("%s: two responses from source %q", at, resp.Source)
			}
			answered[resp.Source] = true
			if resp.Status < 100 || resp.ContentType == "" {
				add("%s: source %q: status %d, content type %q", at, resp.Source, resp.Status, resp.ContentType)
			}
			if unread[[2]int{i, j}] {
				continue
			}
			for _, leak := range scanCredentials(resp.Body, req.Format) {
				add("%s: source %q: %s: %w", at, resp.Source, leak, errBodyCredential)
			}
			if !declared || src.ServerVersion == "" {
				continue
			}
			env, err := client.ParseEnvelope(resp.Body, req.Format)
			switch {
			case err != nil:
				add("%s: source %q: body has no readable envelope: %v", at, resp.Source, err)
			case env.ServerVersion != src.ServerVersion:
				add("%s: source %q: body serverVersion %q, metadata %q", at, resp.Source, env.ServerVersion, src.ServerVersion)
			}
		}
	}
	for _, key := range slices.Sorted(maps.Keys(s.Resolutions)) {
		if !used[key] {
			add("resolution key %q is used by no request", key)
		}
		for _, name := range slices.Sorted(maps.Keys(s.Resolutions[key])) {
			if _, ok := s.Sources[name]; !ok {
				add("resolution %q: undeclared source %q", key, name)
			}
			if s.Resolutions[key][name] == "" {
				add("resolution %q: source %q has an empty id", key, name)
			}
		}
	}
	return problems
}
