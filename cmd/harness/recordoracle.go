package main

import (
	"cmp"
	"context"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"maps"
	"net/url"
	"os"
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
	errNoFolder    = errors.New("no music folder lists exactly the collection's artists")
)

// requestTimeout bounds one oracle request.
const requestTimeout = 30 * time.Second

// refKinds are the prefixes that mark a spec parameter value as a symbolic
// reference to a fixture entity. Every other value is sent as written.
var refKinds = []string{"artist:", "album:", "song:", "folder:", "collection:"}

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
	reqs := make([]session.Request, len(spec.Requests))
	for i, r := range spec.Requests {
		reqs[i] = r.sessionRequest()
	}
	if err := session.CheckRequests(reqs); err != nil {
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
	spec, digests, out string
	collections        []collection
}

// startOracles runs the oracles pinned in the digest file on the built
// collections and returns them with the function that removes them. Tests
// swap it.
var startOracles = startStack

// newRecordOracleCmd builds the record-oracle subcommand: it starts the
// pinned oracles on the fixture library, issues a session spec's requests
// against each, and writes the session.
func newRecordOracleCmd() *cobra.Command {
	var spec, manifest, digests, out string
	var collectionFlags []string

	cmd := &cobra.Command{
		Use:           "record-oracle",
		Short:         "Record the pinned Navidrome and gonic answers to a session spec",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			cols, digests, err := resolveCollections(manifest, collectionFlags, digests)
			if err != nil {
				return err
			}
			if spec == "" || out == "" {
				p, err := defaultSubsonicPaths()
				if err != nil {
					return err
				}
				spec = cmp.Or(spec, p.seedSpec)
				out = cmp.Or(out, p.seedSession)
			}
			return runRecordOracle(cmd.Context(), recordConfig{spec: spec, collections: cols, digests: digests, out: out})
		},
	}
	cmd.Flags().StringVar(&spec, "spec", "", "session spec (default: the committed seed spec)")
	cmd.Flags().StringVar(&manifest, "manifest", "", "fixture manifest, the only collection (default: the committed one)")
	cmd.Flags().StringArrayVar(&collectionFlags, "collection", nil, "collection as name=manifest; repeat for several")
	cmd.MarkFlagsMutuallyExclusive("manifest", "collection")
	cmd.Flags().StringVar(&digests, "digests", "", "oracle image digest file (default: the committed one)")
	cmd.Flags().StringVar(&out, "out", "", "session directory to create (default: the committed seed session)")
	return cmd
}

// startStack runs the podman stack and checks each running image against
// its pin.
func startStack(ctx context.Context, digests string, cols []stack.Collection) ([]runningOracle, func(), error) {
	s, err := stack.Start(ctx, stack.Config{Collections: cols, DigestFile: digests})
	if err != nil {
		return nil, nil, err
	}
	var runs []runningOracle
	for _, o := range s.Oracles {
		digest, err := s.ImageDigest(ctx, o)
		if err != nil {
			s.Close(ctx)
			return nil, nil, err
		}
		runs = append(runs, runningOracle{name: o.Name, client: o.Client(requestTimeout), imageDigest: digest})
	}
	return runs, func() { s.Close(ctx) }, nil
}

func runRecordOracle(ctx context.Context, cfg recordConfig) error {
	reqs, err := loadRecordSpec(cfg.spec)
	if err != nil {
		return err
	}
	if entries, err := os.ReadDir(cfg.out); err == nil && len(entries) > 0 {
		return fmt.Errorf("%w: %s", errOutExists, cfg.out)
	}
	root, err := os.MkdirTemp("", "dubplate-record-music-")
	if err != nil {
		return fmt.Errorf("create library directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(root) }()
	built, libraryDigest, err := buildCollections(root, cfg.collections)
	if err != nil {
		return err
	}
	oracles, stop, err := startOracles(ctx, cfg.digests, built)
	if err != nil {
		return err
	}
	defer stop()

	s := &session.Session{}
	for _, r := range reqs {
		s.Requests = append(s.Requests, r.sessionRequest())
	}
	sets, err := albumArtistSets(cfg.collections)
	if err != nil {
		return err
	}
	var additions []session.Addition
	var errs []error
	for _, o := range oracles {
		rec := &recorder{oracle: o, collectionArtists: sets}
		if err := rec.record(ctx, reqs, s.Requests); err != nil {
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
	if len(additions) == 0 {
		return errors.New("no oracle recorded")
	}
	for _, a := range additions {
		if s, err = s.Join(a); err != nil {
			return err
		}
	}
	return session.Write(cfg.out, s)
}

// recorder drives one oracle through the spec.
type recorder struct {
	oracle runningOracle
	// collectionArtists holds each collection's sorted, distinct album
	// artists by collection name.
	collectionArtists map[string][]string
	// folders holds the oracle's music folders, listed once a collection
	// key first asks.
	folders   []musicFolder
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

// musicFolder is one music folder an oracle lists and the distinct artists
// a getArtists filtered by it lists, sorted.
type musicFolder struct {
	id      string
	artists []string
}

// albumArtistSets returns the sorted, distinct album artists of each
// collection's manifest, by collection name.
func albumArtistSets(cols []collection) (map[string][]string, error) {
	sets := make(map[string][]string, len(cols))
	for _, c := range cols {
		lib, err := fixture.Load(c.manifest)
		if err != nil {
			return nil, fmt.Errorf("collection %q: %w", c.name, err)
		}
		var names []string
		for _, t := range lib.Tracks() {
			names = append(names, t.AlbumArtist)
		}
		slices.Sort(names)
		sets[c.name] = slices.Compact(names)
	}
	return sets, nil
}

func (r *recorder) record(ctx context.Context, reqs []specRequest, keyed []session.Request) error {
	r.resolved = map[string]string{}
	r.responses = map[string]session.Response{}
	for i, req := range reqs {
		at := fmt.Errorf("%s: request %d (%s)", r.oracle.name, i+1, req.Endpoint)
		q := url.Values{}
		for name, value := range req.Params {
			id, err := r.resolve(ctx, value)
			if err != nil {
				return fmt.Errorf("%w: %w", at, err)
			}
			q.Set(name, id)
		}
		// Do answers a nil error only with a response, and a failed
		// envelope, an *Error, is an answer the session records.
		resp, err := r.oracle.client.Do(ctx, req.Endpoint, q, req.Format)
		if _, failed := errors.AsType[*client.Error](err); err != nil && (resp == nil || !failed) {
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
func (r *recorder) resolve(ctx context.Context, value string) (string, error) {
	if !slices.ContainsFunc(refKinds, func(k string) bool { return strings.HasPrefix(value, k) }) {
		return value, nil
	}
	if id, ok := r.resolved[value]; ok {
		return id, nil
	}
	var id string
	var err error
	if name, ok := strings.CutPrefix(value, "collection:"); ok {
		id, err = r.collectionFolder(ctx, name)
	} else {
		id, err = r.lookup(value)
	}
	if err != nil {
		return "", fmt.Errorf("reference %q: %w", value, err)
	}
	r.resolved[value] = id
	return id, nil
}

// collectionFolder returns the id of the music folder whose filtered
// getArtists lists exactly the artists of the named collection. It asks the
// oracle itself, since a folder's id and name are the oracle's to choose and
// the spec's order says nothing about them.
func (r *recorder) collectionFolder(ctx context.Context, name string) (string, error) {
	want, ok := r.collectionArtists[name]
	if !ok {
		return "", fmt.Errorf("%w: no --collection names %q", errBadRef, name)
	}
	if r.folders == nil {
		var err error
		if r.folders, err = r.listFolders(ctx); err != nil {
			return "", err
		}
	}
	var ids []string
	for _, f := range r.folders {
		if slices.Equal(f.artists, want) {
			ids = append(ids, f.id)
		}
	}
	return one("collection:"+name, ids, errNoFolder)
}

// listFolders asks the oracle for its music folders and, for each, the
// artists a getArtists filtered by it lists.
func (r *recorder) listFolders(ctx context.Context) ([]musicFolder, error) {
	var listed struct {
		Response struct {
			MusicFolders struct {
				MusicFolder []struct {
					ID jsontext.Value `json:"id"`
				} `json:"musicFolder"`
			} `json:"musicFolders"`
		} `json:"subsonic-response"`
	}
	if err := r.probe(ctx, "getMusicFolders", nil, &listed); err != nil {
		return nil, err
	}
	var folders []musicFolder
	for _, f := range listed.Response.MusicFolders.MusicFolder {
		id := strings.Trim(string(f.ID), `"`)
		var body wireBody
		if err := r.probe(ctx, "getArtists", url.Values{"musicFolderId": {id}}, &body); err != nil {
			return nil, err
		}
		var names []string
		for _, ix := range body.Response.Artists.Index {
			for _, a := range ix.Artist {
				names = append(names, a.Name)
			}
		}
		slices.Sort(names)
		folders = append(folders, musicFolder{id: id, artists: slices.Compact(names)})
	}
	return folders, nil
}

// probe issues a request the session does not record and decodes its JSON
// answer into into.
func (r *recorder) probe(ctx context.Context, endpoint string, q url.Values, into any) error {
	resp, err := r.oracle.client.Do(ctx, endpoint, q, client.JSON)
	if err != nil {
		return fmt.Errorf("probe %s: %w", endpoint, err)
	}
	if err := json.Unmarshal(resp.Body, into); err != nil {
		return fmt.Errorf("probe %s: decode answer: %w", endpoint, err)
	}
	return nil
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
		}, rest, func(c wireRef) string { return c.Name }), errUnreachable)
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
		}, title, func(c wireRef) string { return c.Name }), errUnreachable)
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
		}, parts[2], func(c wireRef) string { return c.Title }), errUnreachable)
	case "folder":
		segments := strings.Split(rest, "/")
		id, err := one(ref, r.children("getIndexes", "", func(b wireBody) []wireRef {
			var all []wireRef
			for _, ix := range b.Response.Indexes.Index {
				all = append(all, ix.Artist...)
			}
			return all
		}, segments[0], func(c wireRef) string { return c.Name }), errUnreachable)
		if err != nil {
			return "", err
		}
		for _, seg := range segments[1:] {
			id, err = one(ref, r.children("getMusicDirectory", id, func(b wireBody) []wireRef {
				return slices.DeleteFunc(slices.Clone(b.Response.Directory.Child), func(c wireRef) bool { return !c.IsDir })
			}, seg, func(c wireRef) string { return c.Title }), errUnreachable)
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

// one returns the single id of ids; no id is the error none.
func one(ref string, ids []string, none error) (string, error) {
	switch len(ids) {
	case 0:
		return "", fmt.Errorf("%w: %s", none, ref)
	case 1:
		return ids[0], nil
	}
	return "", fmt.Errorf("%w: %s matches %d entities", errAmbiguous, ref, len(ids))
}
