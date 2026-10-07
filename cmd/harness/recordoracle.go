package main

import (
	"cmp"
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/harness/subsonic/stack"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
	"github.com/spf13/cobra"
)

var (
	errSpec        = errors.New("bad session spec")
	errUnreachable = errors.New("no earlier response reaches the reference")
	errAmbiguous   = errors.New("reference matches more than one entity")
	errBadRef      = errors.New("malformed reference")
	errOutExists   = errors.New("output directory is not empty")
	errNoVersion   = errors.New("response carries no serverVersion")
)

// requestTimeout bounds one oracle request.
const requestTimeout = 30 * time.Second

// refKinds are the prefixes that mark a spec parameter value as a symbolic
// reference to a fixture entity. Every other value is sent as written.
var refKinds = []string{"artist:", "album:", "song:", "folder:"}

// recordSpec is the committed list of requests a session records.
type recordSpec struct {
	Requests []specRequest `json:"requests"`
}

type specRequest struct {
	Endpoint string            `json:"endpoint"`
	Format   client.Format     `json:"format"`
	Params   map[string]string `json:"params,omitempty"`
}

// sessionRequest returns r as the session's request, parameters sorted.
func (r specRequest) sessionRequest() session.Request {
	req := session.Request{Endpoint: r.Endpoint, Format: r.Format}
	for _, name := range slices.Sorted(maps.Keys(r.Params)) {
		req.Params = append(req.Params, session.Param{Name: name, Value: r.Params[name]})
	}
	return req
}

// loadRecordSpec reads and checks the spec at path; every problem found is
// reported in one error.
func loadRecordSpec(path string) ([]specRequest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read session spec: %w", err)
	}
	var spec recordSpec
	if err := json.Unmarshal(data, &spec, json.RejectUnknownMembers(true)); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errSpec, path, err)
	}
	if len(spec.Requests) == 0 {
		return nil, fmt.Errorf("%w: %s lists no requests", errSpec, path)
	}
	var problems []error
	seen := map[string]bool{}
	for i, r := range spec.Requests {
		at := fmt.Sprintf("request %d (%s)", i+1, r.Endpoint)
		if r.Endpoint == "" || strings.HasSuffix(r.Endpoint, ".view") {
			problems = append(problems, fmt.Errorf("%s: endpoint is not in canonical form", at))
		}
		if r.Format != client.JSON && r.Format != client.XML {
			problems = append(problems, fmt.Errorf("%s: unknown format %q", at, r.Format))
		}
		for _, name := range slices.Sorted(maps.Keys(r.Params)) {
			if session.IsAuthParam(name) {
				problems = append(problems, fmt.Errorf("%s: parameter %q carries a credential", at, name))
			}
		}
		if key := r.sessionRequest().Key(); seen[key] {
			problems = append(problems, fmt.Errorf("%s: duplicate of an earlier request", at))
		} else {
			seen[key] = true
		}
	}
	if err := errors.Join(problems...); err != nil {
		return nil, fmt.Errorf("%w: %s: %w", errSpec, path, err)
	}
	return spec.Requests, nil
}

// runningOracle is one server a recording drives.
type runningOracle struct {
	name        string
	client      *client.Client
	imageDigest string // verified against the running image, never the digest file
}

// recordConfig is one record-oracle run.
type recordConfig struct {
	spec, manifest, out string
	// start runs the oracles on the library under music and returns them
	// with the function that removes them.
	start func(ctx context.Context, music string, tracks int) ([]runningOracle, func(), error)
}

// newRecordOracleCmd builds the record-oracle subcommand: it starts the
// pinned oracles on the fixture library, issues a session spec's requests
// against each, and writes the session.
func newRecordOracleCmd() *cobra.Command {
	var spec, manifest, digests, out string

	cmd := &cobra.Command{
		Use:           "record-oracle",
		Short:         "Record the pinned Navidrome and gonic answers to a session spec",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if spec == "" || manifest == "" || digests == "" || out == "" {
				root, err := repoRoot()
				if err != nil {
					return err
				}
				data := filepath.Join(root, "internal", "harness", "subsonic", "data")
				spec = cmp.Or(spec, filepath.Join(data, "sessions", "seed.spec.json"))
				manifest = cmp.Or(manifest, filepath.Join(data, "fixture", "manifest.json"))
				digests = cmp.Or(digests, filepath.Join(data, "oracles.json"))
				out = cmp.Or(out, filepath.Join(data, "sessions", "seed"))
			}
			return runRecordOracle(cmd.Context(), recordConfig{
				spec: spec, manifest: manifest, out: out,
				start: stackStarter(digests),
			})
		},
	}
	cmd.Flags().StringVar(&spec, "spec", "", "session spec (default: the committed seed spec)")
	cmd.Flags().StringVar(&manifest, "manifest", "", "fixture manifest (default: the committed one)")
	cmd.Flags().StringVar(&digests, "digests", "", "oracle image digest file (default: the committed one)")
	cmd.Flags().StringVar(&out, "out", "", "session directory to create (default: the committed seed session)")
	return cmd
}

// stackStarter returns the production start function: the podman stack,
// each running image checked against its pin.
func stackStarter(digests string) func(context.Context, string, int) ([]runningOracle, func(), error) {
	return func(ctx context.Context, music string, tracks int) ([]runningOracle, func(), error) {
		s, err := stack.Start(ctx, stack.Config{MusicDir: music, Tracks: tracks, DigestFile: digests})
		if err != nil {
			return nil, nil, err
		}
		var runs []runningOracle
		for _, o := range []stack.Oracle{s.Navidrome, s.Gonic} {
			digest, err := s.ImageDigest(ctx, o)
			if err != nil {
				s.Close(ctx)
				return nil, nil, err
			}
			runs = append(runs, runningOracle{
				name: o.Name,
				client: &client.Client{
					BaseURL: o.BaseURL, User: o.User, Password: o.Password,
					Name: "dubplate-harness", HTTP: &http.Client{Timeout: requestTimeout},
				},
				imageDigest: digest,
			})
		}
		return runs, func() { s.Close(ctx) }, nil
	}
}

func runRecordOracle(ctx context.Context, cfg recordConfig) error {
	reqs, err := loadRecordSpec(cfg.spec)
	if err != nil {
		return err
	}
	if entries, err := os.ReadDir(cfg.out); err == nil && len(entries) > 0 {
		return fmt.Errorf("%w: %s", errOutExists, cfg.out)
	}
	tracks, err := fixture.TrackCount(cfg.manifest)
	if err != nil {
		return err
	}
	music, err := os.MkdirTemp("", "dubplate-record-music-")
	if err != nil {
		return fmt.Errorf("create library directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(music) }()
	libraryDigest, err := fixture.Build(cfg.manifest, music)
	if err != nil {
		return err
	}
	oracles, stop, err := cfg.start(ctx, music, tracks)
	if err != nil {
		return err
	}
	defer stop()

	base := &session.Session{}
	for _, r := range reqs {
		base.Requests = append(base.Requests, r.sessionRequest())
	}
	var additions []session.Addition
	var errs []error
	for _, o := range oracles {
		rec := &recorder{oracle: o}
		if err := rec.record(ctx, reqs, base.Requests); err != nil {
			errs = append(errs, err)
			continue
		}
		additions = append(additions, session.Addition{
			Name: o.name,
			Source: session.Source{
				LibraryDigest: libraryDigest, ImageDigest: o.imageDigest, ServerVersion: rec.version,
			},
			Resolutions: rec.resolved,
			Responses:   rec.responses,
		})
	}
	if err := errors.Join(errs...); err != nil {
		return err
	}
	return writeJoined(cfg.out, base, additions)
}

// writeJoined writes base with each addition joined to out, one oracle at a
// time through a scratch directory, and removes out again on failure.
func writeJoined(out string, base *session.Session, additions []session.Addition) error {
	if len(additions) == 0 {
		return errors.New("no oracle recorded")
	}
	scratch, err := os.MkdirTemp("", "dubplate-record-session-")
	if err != nil {
		return fmt.Errorf("create scratch directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(scratch) }()

	s := base
	for i, a := range additions {
		dir := filepath.Join(scratch, fmt.Sprint(i))
		if i == len(additions)-1 {
			dir = out
		}
		if err := session.Join(dir, s, a); err != nil {
			if dir == out {
				_ = os.RemoveAll(out)
			}
			return err
		}
		if s, err = session.Load(dir); err != nil {
			return err
		}
	}
	return nil
}

// recorder drives one oracle through the spec.
type recorder struct {
	oracle    runningOracle
	version   string
	resolved  map[string]string
	responses map[string]session.Response
	seen      []seenResponse
}

// seenResponse is an earlier JSON answer a later reference may resolve
// against, keyed by the id its request carried.
type seenResponse struct {
	endpoint string
	id       string
	body     wireBody
}

// wireBody holds the members of a JSON answer that references resolve
// against.
type wireBody struct {
	Response struct {
		Artists struct {
			Index []struct {
				Artist []wireRef `json:"artist"`
			} `json:"index"`
		} `json:"artists"`
		Indexes struct {
			Index []struct {
				Artist []wireRef `json:"artist"`
			} `json:"index"`
		} `json:"indexes"`
		Artist struct {
			Album []wireRef `json:"album"`
		} `json:"artist"`
		Album struct {
			Song []wireRef `json:"song"`
		} `json:"album"`
		Directory struct {
			Child []wireRef `json:"child"`
		} `json:"directory"`
	} `json:"subsonic-response"`
}

type wireRef struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title"`
	IsDir bool   `json:"isDir"`
}

func (r *recorder) record(ctx context.Context, reqs []specRequest, keyed []session.Request) error {
	r.resolved = map[string]string{}
	r.responses = map[string]session.Response{}
	for i, req := range reqs {
		at := fmt.Errorf("%s: request %d (%s)", r.oracle.name, i+1, req.Endpoint)
		q := url.Values{}
		for name, value := range req.Params {
			id, err := r.resolve(value)
			if err != nil {
				return fmt.Errorf("%w: %w", at, err)
			}
			q.Set(name, id)
		}
		rctx, cancel := context.WithTimeout(ctx, requestTimeout)
		resp, err := r.oracle.client.Do(rctx, req.Endpoint, q, req.Format)
		cancel()
		if _, failed := errors.AsType[*client.Error](err); resp == nil || (err != nil && !failed) {
			return fmt.Errorf("%w: %w", at, err)
		}
		switch v := resp.Envelope.ServerVersion; {
		case v == "":
			return fmt.Errorf("%w: %w", at, errNoVersion)
		case r.version == "":
			r.version = v
		case v != r.version:
			return fmt.Errorf("%w: serverVersion %q, earlier %q", at, v, r.version)
		}
		r.responses[keyed[i].Key()] = session.Response{
			Status:      resp.StatusCode,
			ContentType: resp.Header.Get("Content-Type"),
			Body:        resp.Body,
		}
		if req.Format != client.JSON {
			continue
		}
		var body wireBody
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			return fmt.Errorf("%w: decode answer: %w", at, err)
		}
		r.seen = append(r.seen, seenResponse{endpoint: req.Endpoint, id: q.Get("id"), body: body})
	}
	return nil
}

// resolve returns the wire value for a spec parameter value. A reference
// resolves through the answers recorded so far and lands in the resolution
// map; any other value is returned as written.
func (r *recorder) resolve(value string) (string, error) {
	if !slices.ContainsFunc(refKinds, func(k string) bool { return strings.HasPrefix(value, k) }) {
		return value, nil
	}
	if id, ok := r.resolved[value]; ok {
		return id, nil
	}
	id, err := r.lookup(value)
	if err != nil {
		return "", fmt.Errorf("reference %q: %w", value, err)
	}
	r.resolved[value] = id
	return id, nil
}

// lookup walks the recorded answers from the reference's outermost entity
// inward, so a title that two artists share resolves through the artist the
// reference names and never by the title alone.
func (r *recorder) lookup(ref string) (string, error) {
	kind, rest, _ := strings.Cut(ref, ":")
	switch kind {
	case "artist":
		return one(ref, r.children("getArtists", "", func(b wireBody) []wireRef {
			var all []wireRef
			for _, ix := range b.Response.Artists.Index {
				all = append(all, ix.Artist...)
			}
			return all
		}, rest, func(c wireRef) string { return c.Name }))
	case "album":
		artist, title, ok := strings.Cut(rest, "/")
		if !ok {
			return "", fmt.Errorf("%w: want album:ARTIST/TITLE", errBadRef)
		}
		parent, err := r.lookup("artist:" + artist)
		if err != nil {
			return "", err
		}
		return one(ref, r.children("getArtist", parent, func(b wireBody) []wireRef {
			return b.Response.Artist.Album
		}, title, func(c wireRef) string { return c.Name }))
	case "song":
		parts := strings.SplitN(rest, "/", 3)
		if len(parts) != 3 {
			return "", fmt.Errorf("%w: want song:ARTIST/ALBUM/TITLE", errBadRef)
		}
		parent, err := r.lookup("album:" + parts[0] + "/" + parts[1])
		if err != nil {
			return "", err
		}
		return one(ref, r.children("getAlbum", parent, func(b wireBody) []wireRef {
			return b.Response.Album.Song
		}, parts[2], func(c wireRef) string { return c.Title }))
	case "folder":
		segments := strings.Split(rest, "/")
		id, err := one(ref, r.children("getIndexes", "", func(b wireBody) []wireRef {
			var all []wireRef
			for _, ix := range b.Response.Indexes.Index {
				all = append(all, ix.Artist...)
			}
			return all
		}, segments[0], func(c wireRef) string { return c.Name }))
		if err != nil {
			return "", err
		}
		for _, seg := range segments[1:] {
			id, err = one(ref, r.children("getMusicDirectory", id, func(b wireBody) []wireRef {
				return slices.DeleteFunc(slices.Clone(b.Response.Directory.Child), func(c wireRef) bool { return !c.IsDir })
			}, seg, func(c wireRef) string { return c.Title }))
			if err != nil {
				return "", err
			}
		}
		return id, nil
	}
	return "", fmt.Errorf("%w: kind %q", errBadRef, kind)
}

// children returns the ids of every entry named want among the entries
// list extracts from earlier answers to endpoint. A parent of "" accepts
// any answer; otherwise only answers whose request carried that id count.
func (r *recorder) children(endpoint, parent string, list func(wireBody) []wireRef, want string, name func(wireRef) string) []string {
	var ids []string
	for _, s := range r.seen {
		if s.endpoint != endpoint || (parent != "" && s.id != parent) {
			continue
		}
		for _, c := range list(s.body) {
			if name(c) == want && !slices.Contains(ids, c.ID) {
				ids = append(ids, c.ID)
			}
		}
	}
	return ids
}

func one(ref string, ids []string) (string, error) {
	switch len(ids) {
	case 0:
		return "", fmt.Errorf("%w: %s", errUnreachable, ref)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("%w: %s matches %d entities", errAmbiguous, ref, len(ids))
}
