package intake

import (
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// A codeProducer drives one fixture through a create, the tus path, or
// the gate, and returns every reject code the run recorded or answered.
type codeProducer func(t *testing.T) []reject.Code

// createCodes runs one create of n under lim after setup, and returns
// the codes its refusal carried.
func createCodes(n newSubmission, setup func(t *testing.T, f *intakeFixture, m identity.UserID, lim *limits)) codeProducer {
	return func(t *testing.T) []reject.Code {
		f := newIntakeFixture(t)
		m := member(t, f.h)
		lim := testLimits()
		if setup != nil {
			setup(t, f, m, &lim)
		}
		_, err := f.s.create(deadline(t), m, n, lim)
		var manifest *manifestRefusal
		var limit *limitRefusal
		var codes []reject.Code
		switch {
		case errors.As(err, &manifest):
			for _, r := range manifest.refusal.Submission {
				codes = append(codes, r.Code)
			}
			for _, r := range manifest.refusal.Files {
				codes = append(codes, r.Reason.Code)
			}
		case errors.As(err, &limit):
			codes = append(codes, limit.code)
		}
		return codes
	}
}

// gateCodes stages the submission build makes, runs the worker's check,
// and returns the verdict reason and every file verdict it recorded.
func gateCodes(build func(t *testing.T) (newSubmission, [][]byte)) codeProducer {
	return func(t *testing.T) []reject.Code {
		f := newIntakeFixture(t)
		m := member(t, f.h)
		n, data := build(t)
		c := f.checkingWith(t, m, n, data...)
		f.in.check(deadline(t), c.id)
		return recordedCodes(t, f, c)
	}
}

func recordedCodes(t *testing.T, f *intakeFixture, c created) []reject.Code {
	t.Helper()
	var codes []reject.Code
	if _, reason := lastEvent(t, f.h, c.id); reason != "" {
		codes = append(codes, reject.Code(reason))
	}
	for _, code := range failedFiles(t, f.h, c.id) {
		codes = append(codes, reject.Code(code))
	}
	return codes
}

// oneFile is a one-file album of spec's bytes, passed through edit when
// it is not nil.
func oneFile(spec flacSpec, edit func([]byte) []byte) func(t *testing.T) (newSubmission, [][]byte) {
	return func(t *testing.T) (newSubmission, [][]byte) {
		data := spec.encode(t)
		if edit != nil {
			data = edit(data)
		}
		return albumOf(data), [][]byte{data}
	}
}

// albumOfSpecs is a loose album of each spec's bytes.
func albumOfSpecs(specs ...flacSpec) func(t *testing.T) (newSubmission, [][]byte) {
	return func(t *testing.T) (newSubmission, [][]byte) {
		var data [][]byte
		for _, s := range specs {
			data = append(data, s.encode(t))
		}
		return albumOf(data...), data
	}
}

// TestEveryBlockingCodeIsProduced is AC 17's keyed table: every blocking
// reason in the set is produced by a create, tus-path, or gate fixture,
// and a code the set holds that the table lacks fails the test. symlink
// is exempt with its reason.
func TestEveryBlockingCodeIsProduced(t *testing.T) {
	exempt := map[reject.Code]string{
		reject.Symlink: "a ZIP symlink member answers unsafe-path; symlink is kept for rung 9's timer path",
	}
	tagged := func(tags ...[2]string) flacSpec {
		s := cdTrack(60, 1)
		s.tags = tags
		return s
	}
	table := map[reject.Code]codeProducer{
		reject.BadExtension: createCodes(newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles,
			Source: "cd", Files: []gate.ManifestFile{{Path: "01.mp3", Size: 10}}}}, nil),
		reject.UnsafePath: createCodes(newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles,
			Source: "cd", Files: []gate.ManifestFile{{Path: "../01.flac", Size: 10}}}}, nil),
		reject.PathTooDeep: createCodes(newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles,
			Source: "cd", Files: []gate.ManifestFile{{Path: strings.Repeat("d/", 9) + "01.flac", Size: 10}}}}, nil),
		reject.NoAudio: createCodes(newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles,
			Source: "cd", Files: []gate.ManifestFile{{Path: "cover.jpg", Size: 10}}}}, nil),
		reject.MissingSource: createCodes(newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles,
			Files: []gate.ManifestFile{{Path: "01.flac", Size: 10}}}}, nil),
		reject.TooManyFiles: createCodes(album(10, 10), func(_ *testing.T, _ *intakeFixture, _ identity.UserID, lim *limits) {
			lim.policy.MaxFiles = 1
		}),
		reject.TooLarge: createCodes(album(10, 10), func(_ *testing.T, _ *intakeFixture, _ identity.UserID, lim *limits) {
			lim.policy.MaxAlbumBytes = 15
		}),
		reject.UploadsOff: createCodes(album(10), func(t *testing.T, f *intakeFixture, m identity.UserID, _ *limits) {
			override(t, f.h, m, nil, 0, nil)
		}),
		reject.TooManySessions: createCodes(album(10), func(t *testing.T, f *intakeFixture, m identity.UserID, lim *limits) {
			lim.sessionsPerMember = 1
			mustCreate(t, f.s, m, album(10), *lim)
		}),
		reject.OverQuota: createCodes(album(10), func(t *testing.T, f *intakeFixture, m identity.UserID, _ *limits) {
			override(t, f.h, m, 5, nil, nil)
		}),
		reject.NoSpace: createCodes(album(10), func(_ *testing.T, f *intakeFixture, _ identity.UserID, lim *limits) {
			lim.freeSpaceFloor = plentyFree
		}),
		reject.CouldNotCheck: func(t *testing.T) []reject.Code {
			f := newTusFixture(t)
			m := member(t, f.h)
			c := mustCreate(t, f.s, m, album(4), testLimits())
			f.faults.set(func(d *tusFaults) { d.freeErr = errDiskOnFire })
			reply := f.do(t, createRequest(m, c, 0, 4, nil))
			if reply.status != http.StatusServiceUnavailable {
				t.Errorf("tus create with a failing statfs = %d, want 503", reply.status)
			}
			code, _, _ := strings.Cut(reply.body, ":")
			return []reject.Code{reject.Code(code)}
		},

		reject.Corrupt:         gateCodes(oneFile(cdTrack(61, 1), flipped)),
		reject.BelowCDQuality:  gateCodes(oneFile(flacSpec{rate: 22050, bits: 16, blocks: 2, seed: 62, tags: [][2]string{{"TRACKNUMBER", "1"}}}, nil)),
		reject.UnsupportedFLAC: gateCodes(oneFile(cdTrack(63, 1), func(b []byte) []byte { return withStreamInfoBits(b, 18) })),
		reject.SourceDamaged: gateCodes(oneFile(cdTrack(64, 1), func(b []byte) []byte {
			return withHeaderMD5(b, []byte("0123456789abcdef"))
		})),
		reject.EmptyTrack:   gateCodes(oneFile(flacSpec{rate: 44100, bits: 16, blocks: 0, seed: 65, tags: [][2]string{{"TRACKNUMBER", "1"}}}, nil)),
		reject.TagsTooLarge: gateCodes(oneFile(tagged([2]string{"TRACKNUMBER", "1"}, [2]string{"COMMENT", strings.Repeat("x", 5000)}), nil)),
		reject.PictureTooLarge: gateCodes(oneFile(flacSpec{rate: 44100, bits: 16, blocks: 2, seed: 66, picture: 70000,
			tags: [][2]string{{"TRACKNUMBER", "1"}}}, nil)),
		reject.ID3Tags: gateCodes(oneFile(cdTrack(67, 1), func(b []byte) []byte {
			return append([]byte{'I', 'D', '3', 3, 0, 0, 0, 0, 0, 0}, b...)
		})),
		reject.MissingTrackNumber:     gateCodes(oneFile(tagged(), nil)),
		reject.DuplicateTrackPosition: gateCodes(albumOfSpecs(cdTrack(68, 1), cdTrack(69, 1))),
		reject.IncompleteAlbum:        gateCodes(albumOfSpecs(cdTrack(70, 1), cdTrack(71, 3))),
		reject.UnpackFailed: gateCodes(func(*testing.T) (newSubmission, [][]byte) {
			garbage := []byte(strings.Repeat("not a zip ", 20))
			return zipped(gate.UnitAlbum, garbage), [][]byte{garbage}
		}),
		reject.AlreadyInLibrary: func(t *testing.T) []reject.Code {
			f := newIntakeFixture(t)
			m := member(t, f.h)
			track := cdTrack(72, 1).encode(t)
			first := f.checkingWith(t, m, tracksOf(track), track)
			f.in.check(deadline(t), first.id)
			twin := f.checkingWith(t, m, tracksOf(track), track)
			f.in.check(deadline(t), twin.id)
			return recordedCodes(t, f, twin)
		},
	}

	for _, r := range reject.All() {
		if !r.Blocking {
			continue
		}
		if reason, ok := exempt[r.Code]; ok {
			if reason == "" {
				t.Errorf("%s is exempt with no reason", r.Code)
			}
			continue
		}
		produce, ok := table[r.Code]
		if !ok {
			t.Errorf("the reject set holds %s, which the table does not produce", r.Code)
			continue
		}
		t.Run(string(r.Code), func(t *testing.T) {
			if got := produce(t); !slices.Contains(got, r.Code) {
				t.Errorf("the fixture produced %q, want %s", got, r.Code)
			}
		})
	}
}
