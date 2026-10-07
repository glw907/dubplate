package grade

import (
	"cmp"
	"path"
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
)

// library is the fixture manifest seen as the set of identities a server
// response may name.
type library struct {
	artists map[string]bool
	albums  map[string]bool // keyed by albumKey
	dirs    map[string]bool // every directory a track path passes through
	tracks  []fixture.Track
}

func newLibrary(tracks []fixture.Track) *library {
	l := &library{artists: map[string]bool{}, albums: map[string]bool{}, dirs: map[string]bool{}, tracks: tracks}
	for _, t := range tracks {
		for _, a := range append([]string{t.AlbumArtist}, t.Artists...) {
			if a != "" {
				l.artists[a] = true
			}
		}
		if t.AlbumArtist != "" && t.Album != "" {
			l.albums[albumKey(t.AlbumArtist, t.Album)] = true
		}
		for d := path.Dir(t.Path); d != "."; d = path.Dir(d) {
			l.dirs[d] = true
		}
	}
	return l
}

func albumKey(artist, title string) string { return "album:" + artist + "/" + title }

// entityKind names what an object with an id stands for.
type entityKind string

const (
	entityNone    entityKind = ""
	entityArtist  entityKind = "artist"
	entityAlbum   entityKind = "album"
	entitySong    entityKind = "song"
	entityFolder  entityKind = "folder"
	entityLibrary entityKind = "library"
)

// kindOf classifies the object n held by the member name of a response to
// endpoint. Only an object with an id is an entity. getIndexes lists
// folders under the member the other endpoints use for artists, and a
// directory child is a folder or a song by its isDir flag.
func kindOf(endpoint, member string, n *node) entityKind {
	if n.kind != kindObject || n.fields["id"] == nil {
		return entityNone
	}
	switch member {
	case "artist":
		if endpoint == "getIndexes" {
			return entityFolder
		}
		return entityArtist
	case "artists", "albumArtists":
		return entityArtist
	case "album":
		return entityAlbum
	case "song":
		return entitySong
	case "directory":
		return entityFolder
	case "child":
		if n.field("isDir") == "true" {
			return entityFolder
		}
		return entitySong
	case "musicFolder":
		return entityLibrary
	}
	return entityNone
}

// identity is what an entity resolved to in the manifest. key is set
// when exactly one manifest identity matches; otherwise problem says why
// none was chosen.
type identity struct {
	kind    entityKind
	key     string
	problem string
}

// identCtx carries what an entity's identity depends on beyond its own
// members: the request's endpoint and id parameter, and the album or
// folder enclosing it.
type identCtx struct {
	endpoint string
	param    string
	album    string // albumKey of the enclosing album, or ""
	albumArt string // that album's artist
	albumNm  string // that album's title
	folder   string // directory path of the enclosing folder, or ""
}

// describe is an entity's identity as a finding prints it.
func (id identity) describe(n *node) string {
	if id.kind == entityNone {
		return "not an entity: " + n.render()
	}
	if id.key == "" {
		return id.problem
	}
	return id.key
}

// enter returns the context the members of the entity n, resolved to id,
// see.
func (c identCtx) enter(id identity, n *node) identCtx {
	switch id.kind {
	case entityAlbum:
		c.album, c.albumArt, c.albumNm = id.key, n.field("artist"), n.title()
	case entityFolder:
		c.folder = strings.TrimPrefix(id.key, "folder:")
	}
	return c
}

// resolve finds the manifest identity of the entity n.
func (l *library) resolve(kind entityKind, n *node, ctx identCtx, member string) identity {
	id := identity{kind: kind}
	switch kind {
	case entityArtist:
		name := n.field("name")
		if !l.artists[name] {
			id.problem = "no artist " + strconv.Quote(name) + " in the manifest"
			return id
		}
		id.key = "artist:" + name
	case entityAlbum:
		key := albumKey(n.field("artist"), n.title())
		if !l.albums[key] {
			id.problem = "no " + key + " in the manifest"
			return id
		}
		id.key = key
	case entitySong:
		return l.resolveSong(n, ctx)
	case entityFolder:
		dir := folderPath(n, ctx, member)
		if !l.dirs[dir] {
			id.problem = "no folder " + strconv.Quote(dir) + " in the manifest"
			return id
		}
		id.key = "folder:" + dir
	case entityLibrary:
		// The fixture is one library, so every music folder a server
		// lists for it is that library.
		id.key = "library"
	}
	return id
}

// folderPath reads a folder's directory path: a directory response is
// the folder its request named, a directory child sits under its
// directory, and an index entry is a top-level folder.
func folderPath(n *node, ctx identCtx, member string) string {
	switch member {
	case "directory":
		dir, ok := strings.CutPrefix(ctx.param, "folder:")
		if !ok || path.Base(dir) != n.field("name") {
			return ""
		}
		return dir
	case "child":
		if ctx.folder == "" {
			return ""
		}
		return ctx.folder + "/" + n.field("title")
	}
	return n.field("name")
}

// resolveSong matches a song by album, then track number, then disc
// number when the track number alone leaves several; a song with no track
// number matches by title instead. The album artist is the enclosing
// album's, else the song's own album artist members.
func (l *library) resolveSong(n *node, ctx identCtx) identity {
	id := identity{kind: entitySong}
	album := n.field("album")
	artist := ""
	if ctx.album != "" && ctx.albumNm == album {
		artist = ctx.albumArt
	}
	if artist == "" {
		artist = n.field("displayAlbumArtist")
	}
	if artist == "" {
		if aa := n.fields["albumArtists"]; aa != nil && aa.kind == kindArray && len(aa.items) > 0 {
			artist = aa.items[0].field("name")
		}
	}
	matches := func(pred func(fixture.Track) bool, in []fixture.Track) []fixture.Track {
		return slices.DeleteFunc(slices.Clone(in), func(t fixture.Track) bool { return !pred(t) })
	}
	cands := matches(func(t fixture.Track) bool {
		return t.Album == album && (artist == "" || t.AlbumArtist == artist)
	}, l.tracks)
	if track := n.field("track"); track != "" {
		cands = matches(func(t fixture.Track) bool { return strconv.Itoa(t.Number) == track }, cands)
		if len(cands) > 1 {
			disc := cmp.Or(n.field("discNumber"), "0")
			cands = matches(func(t fixture.Track) bool { return strconv.Itoa(t.Disc) == disc }, cands)
		}
	} else {
		title := n.field("title")
		cands = matches(func(t fixture.Track) bool { return t.Title == title }, cands)
	}
	desc := "song " + strconv.Quote(n.field("title")) + " on " + albumKey(artist, album)
	switch len(cands) {
	case 0:
		id.problem = "no manifest track matches " + desc
	case 1:
		id.key = "song:" + cands[0].Path
	default:
		id.problem = strconv.Itoa(len(cands)) + " manifest tracks match " + desc
	}
	return id
}
