package fixture

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

// errInvalid marks a manifest Load refuses. The error carries every
// problem found, one per line.
var errInvalid = errors.New("fixture: invalid manifest")

const (
	// vorbisCommentBlock is the FLAC metadata block type of VORBIS_COMMENT
	// (RFC 9639 section 8.1).
	vorbisCommentBlock = 4
	// frontCover is the FLAC PICTURE type of a front cover.
	frontCover = 3
)

type manifest struct {
	Seeds  []seedEntry  `json:"seeds"`
	Albums []albumEntry `json:"albums"`
}

type seedEntry struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type albumEntry struct {
	Artist string       `json:"artist"`
	Title  string       `json:"title"`
	Cover  *fileEntry   `json:"cover,omitempty"`
	Tracks []trackEntry `json:"tracks"`
}

// fileEntry is one image file written verbatim from a seed.
type fileEntry struct {
	Path string `json:"path"`
	Seed string `json:"seed"`
}

type trackEntry struct {
	Path    string       `json:"path"`
	Seed    string       `json:"seed"`
	Tags    tagio.TagSet `json:"tags"`
	Picture string       `json:"picture,omitempty"`
}

// Library is a manifest whose every check has passed, with the seed
// bytes it was checked against.
type Library struct {
	manifest manifest
	seeds    map[string][]byte
}

// Load reads and checks the manifest at path. The seeds are read from a
// seeds directory beside it. A manifest with any problem yields one error
// naming all of them.
func Load(path string) (*Library, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	var m manifest
	if err := json.Unmarshal(data, &m, json.RejectUnknownMembers(true)); err != nil {
		return nil, fmt.Errorf("parse manifest %s: %w", path, err)
	}
	if len(m.Albums) == 0 {
		return nil, fmt.Errorf("%w: manifest declares no albums", errInvalid)
	}
	lib := &Library{manifest: m, seeds: make(map[string][]byte)}
	problems := lib.check(filepath.Join(filepath.Dir(path), "seeds"))
	if len(problems) > 0 {
		return nil, fmt.Errorf("%w: %w", errInvalid, errors.Join(problems...))
	}
	return lib, nil
}

// check reads the declared seeds from dir into l and returns every
// problem with the seeds and with the paths that use them.
func (l *Library) check(dir string) []error {
	var problems []error
	trackSeeds := make(map[string]bool)
	for _, a := range l.manifest.Albums {
		for _, t := range a.Tracks {
			trackSeeds[t.Seed] = true
		}
	}
	declared := make(map[string]bool)
	for _, s := range l.manifest.Seeds {
		if !fs.ValidPath(s.Name) || s.Name == "." {
			problems = append(problems, fmt.Errorf("seed %q: name is not a plain relative path", s.Name))
			continue
		}
		if declared[s.Name] {
			problems = append(problems, fmt.Errorf("seed %q declared twice", s.Name))
			continue
		}
		declared[s.Name] = true
		file := filepath.Join(dir, s.Name)
		b, err := os.ReadFile(file)
		if err != nil {
			problems = append(problems, fmt.Errorf("seed %q: %w", s.Name, err))
			continue
		}
		l.seeds[s.Name] = b
		if sum := sha256.Sum256(b); !strings.EqualFold(hex.EncodeToString(sum[:]), s.SHA256) {
			problems = append(problems, fmt.Errorf("seed %q: sha256 is %x, manifest declares %s", s.Name, sum, s.SHA256))
		}
		if trackSeeds[s.Name] {
			if err := requireVorbisComment(file); err != nil {
				problems = append(problems, fmt.Errorf("seed %q: %w", s.Name, err))
			}
		}
	}

	paths := make(map[string]bool)
	checkPath := func(what, p string) {
		switch {
		case !fs.ValidPath(p) || p == ".":
			problems = append(problems, fmt.Errorf("%s: path %q is not a relative path inside the library root", what, p))
		case paths[p]:
			problems = append(problems, fmt.Errorf("%s: path %q is declared twice", what, p))
		}
		paths[p] = true
	}
	checkSeed := func(what, name string) {
		if _, ok := l.seeds[name]; !ok {
			problems = append(problems, fmt.Errorf("%s: seed %q is not available", what, name))
		}
	}
	for _, a := range l.manifest.Albums {
		if a.Cover != nil {
			what := "cover " + a.Cover.Path
			checkPath(what, a.Cover.Path)
			checkSeed(what, a.Cover.Seed)
		}
		for _, t := range a.Tracks {
			what := "track " + t.Path
			checkPath(what, t.Path)
			checkSeed(what, t.Seed)
			if t.Picture != "" {
				checkSeed(what+" picture", t.Picture)
			}
		}
	}
	return problems
}

// requireVorbisComment refuses a seed that is not FLAC or has no
// VORBIS_COMMENT block, since tagio.Write would then stamp the
// building binary's version into the file.
func requireVorbisComment(path string) error {
	c, err := tagio.Parse(path)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(c.Blocks, func(b tagio.Block) bool { return b.Type == vorbisCommentBlock }) {
		return errors.New("no VORBIS_COMMENT block")
	}
	return nil
}

// Track is one track's identity as the manifest declares it. A tag the
// manifest leaves unset is the zero value.
type Track struct {
	// Path is the track's path relative to the library root.
	Path        string
	AlbumArtist string
	Album       string
	Title       string
	// Artists holds the track's artist tag followed by its multi-valued
	// artists tag.
	Artists []string
	Disc    int
	Number  int
}

// Tracks returns every track l declares, in manifest order. Its length is
// the count a scan of the built library must report.
func (l *Library) Tracks() []Track {
	var out []Track
	for _, a := range l.manifest.Albums {
		for _, t := range a.Tracks {
			tr := Track{
				Path:        t.Path,
				AlbumArtist: deref(t.Tags.Albumartist),
				Album:       deref(t.Tags.Album),
				Title:       deref(t.Tags.Title),
				Disc:        deref(t.Tags.Disc),
				Number:      deref(t.Tags.Track),
			}
			if t.Tags.Artist != nil {
				tr.Artists = append(tr.Artists, *t.Tags.Artist)
			}
			tr.Artists = append(tr.Artists, t.Tags.Artists...)
			out = append(out, tr)
		}
	}
	return out
}

func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
