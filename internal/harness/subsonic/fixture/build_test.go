package fixture

import (
	"bytes"
	"crypto/sha256"
	"encoding/json/v2"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

const committedManifest = "../data/fixture/manifest.json"

// readTree maps every regular file under root, by slash path, to its bytes.
func readTree(t *testing.T, root string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		b, err := os.ReadFile(p)
		files[filepath.ToSlash(rel)] = b
		return err
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	return files
}

func TestBuildIsDeterministic(t *testing.T) {
	lib, err := load(committedManifest)
	if err != nil {
		t.Fatalf("load committed manifest: %v", err)
	}
	want := 0
	for _, a := range lib.manifest.Albums {
		want += len(a.Tracks)
		if a.Cover != nil {
			want++
		}
	}
	if want < 20 {
		t.Fatalf("committed manifest declares %d files, want at least 20", want)
	}

	dirA, dirB := t.TempDir(), t.TempDir()
	digestA, err := Build(committedManifest, dirA)
	if err != nil {
		t.Fatalf("first build: %v", err)
	}
	digestB, err := Build(committedManifest, dirB)
	if err != nil {
		t.Fatalf("second build: %v", err)
	}
	filesA, filesB := readTree(t, dirA), readTree(t, dirB)
	if len(filesA) != want {
		t.Errorf("built %d files, manifest declares %d", len(filesA), want)
	}
	if !reflect.DeepEqual(filesA, filesB) {
		t.Errorf("two builds differ")
	}
	if digestA != digestB {
		t.Errorf("digests differ: %s and %s", digestA, digestB)
	}

	var lines []string
	for rel, b := range filesA {
		lines = append(lines, fmt.Sprintf("%s\t%d\t%x\n", rel, len(b), sha256.Sum256(b)))
	}
	slices.Sort(lines)
	if got := fmt.Sprintf("%x", sha256.Sum256([]byte(strings.Join(lines, "")))); digestA != got {
		t.Errorf("digest = %s, want %s over the sorted path, size, sha256 lines", digestA, got)
	}
}

func TestBuildRoundTripsEveryTrack(t *testing.T) {
	lib, err := load(committedManifest)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	dir := t.TempDir()
	if _, err := Build(committedManifest, dir); err != nil {
		t.Fatalf("build: %v", err)
	}
	tracks := 0
	for _, a := range lib.manifest.Albums {
		for _, tr := range a.Tracks {
			tracks++
			path := filepath.Join(dir, filepath.FromSlash(tr.Path))
			got, err := tagio.Read(path)
			if err != nil {
				t.Errorf("%s: read: %v", tr.Path, err)
				continue
			}
			// Read reports an absent list member as empty, so the
			// comparison is over what each side would write back.
			if !reflect.DeepEqual(tagio.Serialize(got), tagio.Serialize(tr.Tags)) {
				t.Errorf("%s: tags read back differ from the manifest\n got %+v\nwant %+v", tr.Path, got, tr.Tags)
			}

			seedPath := filepath.Join("..", "data", "fixture", "seeds", tr.Seed)
			seed, err := tagio.Parse(seedPath)
			if err != nil {
				t.Fatalf("parse seed %s: %v", tr.Seed, err)
			}
			built, err := tagio.Parse(path)
			if err != nil {
				t.Errorf("%s: parse: %v", tr.Path, err)
				continue
			}
			if built.VorbisComment.Vendor != seed.VorbisComment.Vendor || !strings.HasPrefix(seed.VorbisComment.Vendor, "reference libFLAC") {
				t.Errorf("%s: vendor %q, want the seed's %q", tr.Path, built.VorbisComment.Vendor, seed.VorbisComment.Vendor)
			}

			wantPictures := 0
			if tr.Picture != "" {
				wantPictures = 1
			}
			if len(built.Pictures) != wantPictures {
				t.Errorf("%s: %d pictures, want %d", tr.Path, len(built.Pictures), wantPictures)
			} else if wantPictures == 1 && !bytes.Equal(built.Pictures[0].Data, lib.seeds[tr.Picture]) {
				t.Errorf("%s: embedded picture differs from seed %s", tr.Path, tr.Picture)
			}
		}
	}
	if tracks < 20 {
		t.Errorf("round-tripped %d tracks, want at least 20", tracks)
	}
}

func TestCommittedFixtureStaysUnderBudget(t *testing.T) {
	var total int64
	err := filepath.WalkDir("../data/fixture", func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		info, err := d.Info()
		total += info.Size()
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if total == 0 || total >= 512*1024 {
		t.Errorf("committed fixture bytes = %d, want between 1 and %d", total, 512*1024-1)
	}
}

// stripVorbisComment rewrites a FLAC file without its VORBIS_COMMENT block.
func stripVorbisComment(t *testing.T, src string) []byte {
	t.Helper()
	c, err := tagio.Parse(src)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(src)
	if err != nil {
		t.Fatal(err)
	}
	var keep []tagio.Block
	for _, b := range c.Blocks {
		if b.Type != vorbisCommentBlock {
			keep = append(keep, b)
		}
	}
	out := []byte("fLaC")
	for i, b := range keep {
		head := b.Type
		if i == len(keep)-1 {
			head |= 0x80
		}
		out = append(out, head, byte(len(b.Payload)>>16), byte(len(b.Payload)>>8), byte(len(b.Payload)))
		out = append(out, b.Payload...)
	}
	return append(out, raw[c.FrameOffset:]...)
}

func TestLoadStates(t *testing.T) {
	good, err := os.ReadFile("../data/fixture/seeds/tone-220.flac")
	if err != nil {
		t.Fatal(err)
	}
	noComment := stripVorbisComment(t, "../data/fixture/seeds/tone-220.flac")
	bent := slices.Clone(good)
	bent[len(bent)-1] ^= 0xff
	sum := func(b []byte) string { return fmt.Sprintf("%x", sha256.Sum256(b)) }
	seedDecl := func(name string, b []byte) seedEntry { return seedEntry{Name: name, SHA256: sum(b)} }
	one := func(p, seed string) albumEntry {
		return albumEntry{Tracks: []trackEntry{{Path: p, Seed: seed}}}
	}

	tests := []struct {
		name  string
		files map[string][]byte // seed files written under seeds/
		raw   string            // manifest bytes when set, else m is marshalled
		m     *manifest
		want  []string // fragments every one of which the error names
	}{
		{
			name: "absent manifest",
			raw:  "\x00absent",
			want: []string{"read manifest", "manifest.json"},
		},
		{
			name: "empty",
			m:    &manifest{},
			want: []string{"manifest declares no albums"},
		},
		{
			name: "malformed",
			raw:  `{"seeds": [`,
			want: []string{"parse manifest"},
		},
		{
			name: "unknown member",
			raw:  `{"seeds": [], "albums": [], "surprise": 1}`,
			want: []string{"parse manifest", "surprise"},
		},
		{
			name:  "seed hash differs",
			files: map[string][]byte{"a.flac": bent},
			m: &manifest{
				Seeds:  []seedEntry{seedDecl("a.flac", good)},
				Albums: []albumEntry{one("x.flac", "a.flac")},
			},
			want: []string{`seed "a.flac": sha256`},
		},
		{
			name:  "seed without a VORBIS_COMMENT block",
			files: map[string][]byte{"n.flac": noComment},
			m: &manifest{
				Seeds:  []seedEntry{seedDecl("n.flac", noComment)},
				Albums: []albumEntry{one("x.flac", "n.flac")},
			},
			want: []string{`seed "n.flac": no VORBIS_COMMENT block`},
		},
		{
			name:  "many bad",
			files: map[string][]byte{"ok.flac": good, "bent.flac": bent, "n.flac": noComment},
			m: &manifest{
				Seeds: []seedEntry{
					seedDecl("ok.flac", good), seedDecl("bent.flac", good),
					seedDecl("n.flac", noComment), seedDecl("gone.flac", good),
				},
				Albums: []albumEntry{{
					Cover: &fileEntry{Path: "../cover.jpg", Seed: "ok.flac"},
					Tracks: []trackEntry{
						{Path: "dup.flac", Seed: "ok.flac"},
						{Path: "dup.flac", Seed: "ok.flac"},
						{Path: "/abs.flac", Seed: "ok.flac"},
						{Path: "m1.flac", Seed: "gone.flac"},
						{Path: "m2.flac", Seed: "undeclared.flac"},
						{Path: "b.flac", Seed: "bent.flac"},
						{Path: "n.flac", Seed: "n.flac"},
					},
				}},
			},
			want: []string{
				`path "dup.flac" is declared twice`,
				`path "/abs.flac"`,
				`path "../cover.jpg"`,
				`seed "gone.flac": `,
				`seed "undeclared.flac" is not available`,
				`seed "bent.flac": sha256`,
				`seed "n.flac": no VORBIS_COMMENT block`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "seeds"), 0o755); err != nil {
				t.Fatal(err)
			}
			for n, b := range tt.files {
				if err := os.WriteFile(filepath.Join(dir, "seeds", n), b, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			path := filepath.Join(dir, "manifest.json")
			switch {
			case tt.raw == "\x00absent":
			case tt.raw != "":
				if err := os.WriteFile(path, []byte(tt.raw), 0o644); err != nil {
					t.Fatal(err)
				}
			default:
				b, err := json.Marshal(tt.m)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, b, 0o644); err != nil {
					t.Fatal(err)
				}
			}

			target := filepath.Join(t.TempDir(), "lib")
			digest, err := Build(path, target)
			if err == nil {
				t.Fatalf("Build succeeded with digest %s, want an error", digest)
			}
			for _, frag := range tt.want {
				if !strings.Contains(err.Error(), frag) {
					t.Errorf("error does not name %q:\n%v", frag, err)
				}
			}
			if tt.name != "absent manifest" && tt.raw == "" && !errors.Is(err, ErrInvalid) {
				t.Errorf("error is not ErrInvalid: %v", err)
			}
			if _, statErr := os.Stat(target); !errors.Is(statErr, fs.ErrNotExist) {
				t.Errorf("target was touched by a refused build: %v", statErr)
			}
		})
	}
}

func TestTracks(t *testing.T) {
	tracks, err := Tracks(committedManifest)
	if err != nil {
		t.Fatal(err)
	}
	if len(tracks) != 29 {
		t.Errorf("got %d tracks, want 29", len(tracks))
	}
	byPath := map[string]Track{}
	for _, tr := range tracks {
		byPath[tr.Path] = tr
	}
	tests := []struct {
		path string
		want Track
	}{
		{"Borealis Trio/Double Feature/Disc 2/01 Reprise.flac", Track{
			AlbumArtist: "Borealis Trio", Album: "Double Feature", Title: "Reprise",
			Artists: []string{"Borealis Trio"}, Disc: 2, Number: 1,
		}},
		{"Alpha Artist/Single Disc/03 Together.flac", Track{
			AlbumArtist: "Alpha Artist", Album: "Single Disc", Title: "Together",
			Artists: []string{"Alpha Artist feat. Guest Singer", "Alpha Artist", "Guest Singer"}, Disc: 1, Number: 3,
		}},
		{"Delta Duo/Greatest Hits/02 Last Light.flac", Track{
			AlbumArtist: "Delta Duo", Album: "Greatest Hits", Title: "Last Light",
			Artists: []string{"Delta Duo"}, Number: 2,
		}},
		{"Loose/untagged.flac", Track{}},
	}
	for _, tt := range tests {
		got, ok := byPath[tt.path]
		if !ok {
			t.Errorf("no track %s", tt.path)
			continue
		}
		tt.want.Path = tt.path
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%s:\n got %+v\nwant %+v", tt.path, got, tt.want)
		}
	}
	if _, err := Tracks(filepath.Join(t.TempDir(), "manifest.json")); err == nil || !strings.Contains(err.Error(), "read manifest") {
		t.Errorf("absent manifest: err = %v", err)
	}
}
