package gate

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

// track is one generated album member: its position tags, the totals it
// carries (0 for none), and a seed that gives it audio no other track has.
type track struct {
	disc, no     string // raw tag text; "" omits the tag
	total, discs int
}

// build generates the album's files, runs each through CheckFile, and
// returns the results with the clips for key arithmetic. A file that does
// not pass fails the test: the position cases need every file to pass.
func build(t *testing.T, tracks []track) ([]FileResult, []*clip) {
	t.Helper()
	var results []FileResult
	var clips []*clip
	for i, tr := range tracks {
		var tags [][2]string
		if tr.disc != "" {
			tags = append(tags, [2]string{"DISCNUMBER", tr.disc})
		}
		if tr.no != "" {
			tags = append(tags, [2]string{"TRACKNUMBER", tr.no})
		}
		if tr.total != 0 {
			tags = append(tags, [2]string{"TRACKTOTAL", strconv.Itoa(tr.total)})
		}
		if tr.discs != 0 {
			tags = append(tags, [2]string{"DISCTOTAL", strconv.Itoa(tr.discs)})
		}
		c := cd(uint64(100+i), 1, tags...)
		res := check(t, c.encode(t), testPolicy(), UnitAlbum)
		if res.Verdict != "" {
			t.Fatalf("track %d verdict = %q, want a pass", i, res.Verdict)
		}
		results = append(results, res)
		clips = append(clips, c)
	}
	return results, clips
}

func tr(disc, no string) track { return track{disc: disc, no: no} }

func TestCheckAlbum_Positions(t *testing.T) {
	tests := []struct {
		name   string
		tracks []track
		want   reject.Code
	}{
		{"a single disc, no disc tag", []track{tr("", "1"), tr("", "2"), tr("", "3")}, ""},
		{"a two-disc album restarting on disc 2", []track{tr("1", "1"), tr("1", "2"), tr("2", "1"), tr("2", "2")}, ""},
		{"disc 0 reads as disc 1", []track{tr("0", "1"), tr("0", "2")}, ""},
		{"a matching TRACKTOTAL and DISCTOTAL", []track{{no: "1", total: 3, discs: 1}, {no: "2", total: 3, discs: 1}, {no: "3", total: 3, discs: 1}}, ""},
		{"a duplicate position", []track{tr("", "1"), tr("", "1"), tr("", "2")}, reject.DuplicateTrackPosition},
		{"a duplicate on disc 2", []track{tr("1", "1"), tr("2", "1"), tr("2", "1")}, reject.DuplicateTrackPosition},
		{"a gap", []track{tr("", "1"), tr("", "3")}, reject.IncompleteAlbum},
		{"numbering from 2", []track{tr("", "2"), tr("", "3")}, reject.IncompleteAlbum},
		{"a gap on disc 2", []track{tr("1", "1"), tr("2", "1"), tr("2", "3")}, reject.IncompleteAlbum},
		{"a disc gap", []track{tr("1", "1"), tr("3", "1")}, reject.IncompleteAlbum},
		{"discs starting at 2", []track{tr("2", "1"), tr("2", "2")}, reject.IncompleteAlbum},
		{"a TRACKTOTAL above the count", []track{{no: "1", total: 4}, {no: "2", total: 4}, {no: "3", total: 4}}, reject.IncompleteAlbum},
		{"a TRACKTOTAL below the count", []track{{no: "1", total: 2}, {no: "2", total: 2}, {no: "3", total: 2}}, reject.IncompleteAlbum},
		{"a DISCTOTAL above the discs", []track{{no: "1", discs: 2}, {no: "2", discs: 2}}, reject.IncompleteAlbum},
		{"a DISCTOTAL below the discs", []track{{disc: "1", no: "1", discs: 1}, {disc: "2", no: "1", discs: 1}}, reject.IncompleteAlbum},
		{"tracks that disagree on a total", []track{{no: "1", total: 3}, {no: "2", total: 3}, {no: "3", total: 4}}, reject.IncompleteAlbum},
		{"tracks that disagree on the disc total", []track{{disc: "1", no: "1", discs: 2}, {disc: "2", no: "1", discs: 3}}, reject.IncompleteAlbum},
		{"one track carrying a total and the rest none", []track{{no: "1", total: 3}, tr("", "2"), tr("", "3")}, ""},
		{"a total on one disc that is not that disc's count", []track{{disc: "1", no: "1", total: 2}, {disc: "1", no: "2", total: 2}, {disc: "2", no: "1", total: 5}, {disc: "2", no: "2", total: 5}}, reject.IncompleteAlbum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, _ := build(t, tt.tracks)
			got := CheckAlbum(results, nil)
			if got.Verdict != tt.want {
				t.Errorf("verdict = %q, want %q", got.Verdict, tt.want)
			}
			if tt.want != "" && got.Key != nil {
				t.Errorf("a refused album carries key %x, want none", got.Key)
			}
		})
	}
}

// TestCheckAlbum_Key holds the album key to its definition: SHA-256 over
// the track count (8 bytes, big endian) and the audio keys in (disc, track)
// order, written here from the clips' own samples.
func TestCheckAlbum_Key(t *testing.T) {
	tracks := []track{tr("2", "1"), tr("1", "2"), tr("1", "1"), tr("2", "2")}
	results, clips := build(t, tracks)
	want := sha256.New()
	_ = binary.Write(want, binary.BigEndian, uint64(4))
	for _, i := range []int{2, 1, 0, 3} { // disc 1 track 1, disc 1 track 2, disc 2 track 1, disc 2 track 2
		want.Write(audioMD5(16, clips[i].pcm()))
	}
	got := CheckAlbum(results, nil)
	if got.Verdict != "" {
		t.Fatalf("verdict = %q, want a pass", got.Verdict)
	}
	if !bytes.Equal(got.Key, want.Sum(nil)) {
		t.Errorf("album key = %x, want %x", got.Key, want.Sum(nil))
	}
	// The key is a function of the set, not of the order files arrive in.
	slices.Reverse(results)
	if again := CheckAlbum(results, nil); !bytes.Equal(again.Key, got.Key) {
		t.Errorf("key changed with arrival order: %x vs %x", again.Key, got.Key)
	}
}

func TestCheckAlbum_Warnings(t *testing.T) {
	plain := []track{tr("", "1"), tr("", "2")}
	results, _ := build(t, plain)
	covered := slices.Clone(results)
	covered[1].HasPicture = true

	// Album-wide totals on a two-disc album: each disc has two tracks, and
	// every file says four.
	wide := []track{{disc: "1", no: "1", total: 4}, {disc: "1", no: "2", total: 4}, {disc: "2", no: "1", total: 4}, {disc: "2", no: "2", total: 4}}
	wideResults, _ := build(t, wide)

	tests := []struct {
		name    string
		results []FileResult
		extras  []string
		want    []reject.Code
	}{
		{"no picture and no cover image", results, []string{"booklet.pdf", "a.cue"}, []reject.Code{reject.NoCoverArt}},
		{"a picture block on one file", covered, nil, nil},
		{"a cover image among the extras", results, []string{"scans/Cover.JPG"}, nil},
		{"an album-wide total across discs", wideResults, []string{"cover.png"}, []reject.Code{reject.TrackTotalDiffers}},
		{"both", wideResults, nil, []reject.Code{reject.NoCoverArt, reject.TrackTotalDiffers}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckAlbum(tt.results, tt.extras)
			if got.Verdict != "" {
				t.Fatalf("verdict = %q, want a pass (a warning never blocks)", got.Verdict)
			}
			if !slices.Equal(got.Warnings, tt.want) {
				t.Errorf("warnings = %v, want %v", got.Warnings, tt.want)
			}
		})
	}
}

// TestAlbum_NoTagIsRewritten is AC 21's "no tag is rewritten": the files of a
// passing album are byte-identical after the gate has judged every one.
func TestAlbum_NoTagIsRewritten(t *testing.T) {
	dir := t.TempDir()
	root, err := os.OpenRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	var staged []StagedFile
	var before [][]byte
	for i, tags := range [][][2]string{
		{{"DISCNUMBER", "0"}, {"TRACKNUMBER", "1/2"}},
		{{"DISCNUMBER", "0"}, {"TRACKNUMBER", "2/2"}},
	} {
		data := cd(uint64(300+i), 2, tags...).encode(t)
		name := "f" + strconv.Itoa(i)
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
		staged = append(staged, StagedFile{Root: root, Name: name, Path: name + ".flac"})
		before = append(before, data)
	}
	var results []FileResult
	for _, f := range staged {
		res, err := CheckFile(t.Context(), f, testPolicy(), UnitAlbum)
		if err != nil || res.Verdict != "" {
			t.Fatalf("verdict %q, error %v, want a pass", res.Verdict, err)
		}
		results = append(results, res)
	}
	if got := CheckAlbum(results, nil); got.Verdict != "" {
		t.Fatalf("album verdict = %q, want a pass", got.Verdict)
	}
	for i, f := range staged {
		after, err := os.ReadFile(filepath.Join(dir, f.Name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(after, before[i]) {
			t.Errorf("file %d changed during the gate", i)
		}
	}
}

// TestAlbum_VinylSidesAreMissingTrackNumbers shows the per-file half of "the album is
// the verdict unit": the worker refuses the album on the first non-empty
// Verdict, so the gate must hand it each file's own.
func TestAlbum_VinylSidesAreMissingTrackNumbers(t *testing.T) {
	for _, no := range []string{"A1", "A2", "B1"} {
		res := check(t, cd(400, 1, [2]string{"TRACKNUMBER", no}).encode(t), testPolicy(), UnitAlbum)
		if res.Verdict != reject.MissingTrackNumber {
			t.Errorf("TRACKNUMBER=%s verdict = %q, want missing-track-number", no, res.Verdict)
		}
	}
}

// TestSeparateTracks_PerFileVerdicts is AC 22's gate half: three files, one
// corrupt, give two passes and one corrupt, with no track number needed and
// no album check run.
func TestSeparateTracks_PerFileVerdicts(t *testing.T) {
	good1, good2 := cd(500, 4).encode(t), cd(501, 4).encode(t)
	bad := flipByte(cd(502, 4).encode(t), 600)
	var got []reject.Code
	for _, data := range [][]byte{good1, bad, good2} {
		got = append(got, check(t, data, testPolicy(), UnitTrack).Verdict)
	}
	if want := []reject.Code{"", reject.Corrupt, ""}; !slices.Equal(got, want) {
		t.Errorf("verdicts = %q, want %q", got, want)
	}
}
