package intake

import (
	"bytes"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// TestCheckPassesAnAlbumToMatching: a clean two-track album, checked,
// moves to matching with every file's verdict NULL, its stream facts and
// keys written, its album key set, the cover warning recorded, and its
// bytes and reservation kept for the probe.
func TestCheckPassesAnAlbumToMatching(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	c := f.checkingWith(t, m, albumOf(tracks...), tracks...)

	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "matching" {
		t.Fatalf("state = %q (%s), want matching", got, eventKinds(t, f.h, c.id))
	}
	if got := failedFiles(t, f.h, c.id); len(got) != 0 {
		t.Errorf("file verdicts = %v, want none", got)
	}
	var keyed int
	if err := f.h.QueryRow(deadline(t), `SELECT count(*) FROM submission_files WHERE submission_id = ? AND audio_key IS NOT NULL
		AND file_hash IS NOT NULL AND sample_rate = 44100 AND track IS NOT NULL`, c.id).Scan(&keyed); err != nil {
		t.Fatal(err)
	}
	if keyed != 2 {
		t.Errorf("%d files carry their stream facts and keys, want 2", keyed)
	}
	if len(albumKeyOf(t, f.h, c.id)) != 32 {
		t.Errorf("album key %x, want 32 bytes", albumKeyOf(t, f.h, c.id))
	}
	var warnings []string
	rows, err := f.h.Query(deadline(t), `SELECT code FROM submission_warnings WHERE submission_id = ?`, c.id)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			t.Fatal(err)
		}
		warnings = append(warnings, code)
	}
	_ = rows.Close()
	if !slices.Equal(warnings, []string{"no-cover-art"}) {
		t.Errorf("warnings = %q, want no-cover-art", warnings)
	}
	for i, file := range c.files {
		if !bytes.Equal(f.read(t, tusOf(file.id)), tracks[i]) {
			t.Errorf("%s: staged bytes changed or gone", file.path)
		}
	}
	if got := reservedBytes(t, f.h, c.id); got != int64(len(tracks[0])+len(tracks[1])) {
		t.Errorf("reservation = %d, want the two files' %d", got, len(tracks[0])+len(tracks[1]))
	}
	assertStatesMatchEvents(t, f.h)
}

// TestCheckRejectsAnAlbumWithOneFailingTrack: one corrupt track rejects
// the album with that file's verdict, and its passing track is not kept.
func TestCheckRejectsAnAlbumWithOneFailingTrack(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	tracks[1] = flipped(tracks[1])
	c := f.checkingWith(t, m, albumOf(tracks...), tracks...)

	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "rejected" {
		t.Fatalf("state = %q, want rejected", got)
	}
	if got := failedFiles(t, f.h, c.id); got["02.flac"] != "corrupt" || len(got) != 1 {
		t.Errorf("file verdicts = %v, want 02.flac corrupt alone", got)
	}
	for _, file := range c.files {
		if f.exists(t, tusOf(file.id)) || f.exists(t, tusOf(file.id)+".info") {
			t.Errorf("%s: a rejected album's upload is still staged", file.path)
		}
	}
	assertStatesMatchEvents(t, f.h)
}

// TestCheckFaultsAreNotVerdicts is AC 24: each fault of the server's
// moves a passing two-track album to check-unfinished with its staged
// bytes and its reservation kept, and a retry once the fault clears
// reaches the album's verdict. The ZIP rows hold the archive and no x/
// while unfinished.
func TestCheckFaultsAreNotVerdicts(t *testing.T) {
	type faultCase struct {
		name string
		zip  bool
		// fault breaks the fixture before the check, or, for a fault
		// during decode, while it stands paused after extraction; clear
		// undoes it.
		fault, clear func(t *testing.T, f *intakeFixture, c created)
		duringDecode bool
	}
	cases := []faultCase{
		{
			name: "a staged path replaced by a directory",
			fault: func(t *testing.T, f *intakeFixture, c created) {
				name := f.path(tusOf(c.files[1].id))
				if err := os.Remove(name); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(name, 0o750); err != nil {
					t.Fatal(err)
				}
			},
			clear: func(t *testing.T, f *intakeFixture, c created) {
				if err := os.Remove(f.path(tusOf(c.files[1].id))); err != nil {
					t.Fatal(err)
				}
				f.write(t, tusOf(c.files[1].id), twoTrackAlbum(t)[1])
			},
		},
		{
			name: "a staged file made unreadable",
			fault: func(t *testing.T, f *intakeFixture, c created) {
				if os.Geteuid() == 0 {
					t.Skip("uid 0 reads a file whatever its mode, so chmod cannot fault the read")
				}
				if err := os.Chmod(f.path(tusOf(c.files[0].id)), 0); err != nil {
					t.Fatal(err)
				}
			},
			clear: func(t *testing.T, f *intakeFixture, c created) {
				if err := os.Chmod(f.path(tusOf(c.files[0].id)), 0o640); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "a free-space read failing at extraction",
			zip:  true,
			fault: func(_ *testing.T, f *intakeFixture, _ created) {
				f.s.freeBytes = func(string) (uint64, error) { return 0, errors.New("statfs: input/output error") }
			},
			clear: func(_ *testing.T, f *intakeFixture, _ created) {
				f.s.freeBytes = func(string) (uint64, error) { return plentyFree, nil }
			},
		},
		{
			name:         "an extracted member that cannot be read during decode",
			zip:          true,
			duringDecode: true,
			fault: func(t *testing.T, f *intakeFixture, c created) {
				second := xOf(c.id, extractedAs(t, f, c, twoTrackAlbum(t)[1]))
				if err := os.Remove(f.path(second)); err != nil {
					t.Fatal(err)
				}
				if err := os.Mkdir(f.path(second), 0o750); err != nil {
					t.Fatal(err)
				}
			},
			clear: func(*testing.T, *intakeFixture, created) {},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newIntakeFixture(t)
			m := member(t, f.h)
			tracks := twoTrackAlbum(t)
			var c created
			var archive []byte
			if tc.zip {
				archive = twoTrackZIP(t)
				c = f.checkingWith(t, m, zipped(gate.UnitAlbum, archive), archive)
			} else {
				c = f.checkingWith(t, m, albumOf(tracks...), tracks...)
			}
			reserved := reservedBytes(t, f.h, c.id)

			if tc.duringDecode {
				xDir := "submissions/" + wireID(c.id) + "/x"
				ctx := newPauseWhen(func() bool { return f.entries(t, xDir) == 2 })
				done := make(chan struct{})
				go func() { defer close(done); f.in.check(ctx, c.id) }()
				<-ctx.paused
				tc.fault(t, f, c)
				close(ctx.release)
				<-done
			} else {
				tc.fault(t, f, c)
				f.in.check(deadline(t), c.id)
			}

			if got := stateOf(t, f.h, c.id); got != "check-unfinished" {
				t.Fatalf("state = %q (%s), want check-unfinished", got, eventKinds(t, f.h, c.id))
			}
			if kind, reason := lastEvent(t, f.h, c.id); reason != "could-not-check" {
				t.Errorf("event %s carries reason %q, want could-not-check", kind, reason)
			}
			if got := reservedBytes(t, f.h, c.id); got != reserved && (!tc.zip || got != extractedTotal(t)) {
				t.Errorf("reservation = %d, want it kept at %d or the extracted %d", got, reserved, extractedTotal(t))
			}
			if tc.zip {
				if !bytes.Equal(f.read(t, tusOf(c.files[0].id)), archive) {
					t.Error("the archive is not kept whole")
				}
				if f.exists(t, "submissions/"+wireID(c.id)+"/x") {
					t.Error("x/ outlived a check that could not run")
				}
				if got := membersOf(t, f.h, c.id); len(got) != 0 {
					t.Errorf("member rows %v were written for an unfinished check", got)
				}
			} else if tc.name != "a staged path replaced by a directory" && !bytes.Equal(f.read(t, tusOf(c.files[1].id)), tracks[1]) {
				t.Error("a staged file changed")
			}
			if got := failedFiles(t, f.h, c.id); len(got) != 0 {
				t.Errorf("file verdicts %v were written for an unfinished check", got)
			}

			tc.clear(t, f, c)
			if err := f.s.retry(deadline(t), m, c.id); err != nil {
				t.Fatalf("retry: %v", err)
			}
			f.in.check(deadline(t), c.id)
			if got := stateOf(t, f.h, c.id); got != "matching" {
				t.Fatalf("state after the retry = %q (%s), want matching", got, eventKinds(t, f.h, c.id))
			}
			assertStatesMatchEvents(t, f.h)
		})
	}
}

// extractedAs is the server-minted id the member holding track's bytes
// extracted to in c's x/.
func extractedAs(t *testing.T, f *intakeFixture, c created, track []byte) []byte {
	t.Helper()
	dir := "submissions/" + wireID(c.id) + "/x"
	list, err := os.ReadDir(f.path(dir))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range list {
		id, ok := parseID(e.Name())
		if !ok {
			t.Fatalf("x/ holds %q, not a minted id", e.Name())
		}
		if bytes.Equal(f.read(t, dir+"/"+e.Name()), track) {
			return id
		}
	}
	t.Fatal("no extracted member holds the track")
	return nil
}

// extractedTotal is the two-track album's members' size: the reservation
// an archive of it holds once extraction reserves it.
func extractedTotal(t *testing.T) int64 {
	t.Helper()
	tracks := twoTrackAlbum(t)
	return int64(len(tracks[0]) + len(tracks[1]))
}

// TestCancellationCommitsNothing stops a check at every point it looks
// at its context, a deploy's SIGTERM landing anywhere: the submission
// stays checking with its staged bytes, never check-unfinished, and a
// second start reaches the verdict an uninterrupted run reaches. For an
// archive this covers a stop mid-extraction and one after extraction
// before the verdict commits (AC 19).
func TestCancellationCommitsNothing(t *testing.T) {
	for _, container := range []string{"files", "zip"} {
		t.Run(container, func(t *testing.T) {
			stage := func(t *testing.T) (*intakeFixture, created, [][]byte) {
				f := newIntakeFixture(t)
				m := member(t, f.h)
				if container == "zip" {
					archive := twoTrackZIP(t)
					return f, f.checkingWith(t, m, zipped(gate.UnitAlbum, archive), archive), [][]byte{archive}
				}
				tracks := twoTrackAlbum(t)
				return f, f.checkingWith(t, m, albumOf(tracks...), tracks...), tracks
			}

			// An uninterrupted run counts the points and fixes the verdict.
			f, c, _ := stage(t)
			count := newStopAfter(1 << 30)
			f.in.check(count, c.id)
			if got := stateOf(t, f.h, c.id); got != "matching" {
				t.Fatalf("uninterrupted run = %q, want matching", got)
			}
			want := albumKeyOf(t, f.h, c.id)
			points := count.points()
			if points < 5 {
				t.Fatalf("the check looked at its context %d times, too few to stop mid-decode", points)
			}

			// Every stop before the verdict commits leaves the submission
			// checking; a stop in the cleanup after the commit finds the
			// verdict already landed, and so does every later one.
			stopped, committed := 0, false
			for n := 1; n <= points; n++ {
				f, c, staged := stage(t)
				reserved := reservedBytes(t, f.h, c.id)
				f.in.check(newStopAfter(n), c.id)

				got := stateOf(t, f.h, c.id)
				if got == "matching" {
					committed = true
					continue
				}
				if committed || got != "checking" {
					t.Fatalf("stop at point %d of %d: state %q (%s), want checking", n, points, got, eventKinds(t, f.h, c.id))
				}
				stopped++
				for i, file := range c.files {
					if !bytes.Equal(f.read(t, tusOf(file.id)), staged[i]) {
						t.Fatalf("stop at point %d: %s changed or gone", n, file.path)
					}
				}
				if got := reservedBytes(t, f.h, c.id); got != reserved && (container != "zip" || got != extractedTotal(t)) {
					t.Fatalf("stop at point %d: reservation %d, want %d or the extracted %d", n, got, reserved, extractedTotal(t))
				}
				if got := failedFiles(t, f.h, c.id); len(got) != 0 {
					t.Fatalf("stop at point %d: verdicts %v landed", n, got)
				}

				g := f.restart(t)
				if err := g.in.Start(t.Context()); err != nil {
					t.Fatalf("restart: %v", err)
				}
				if got := g.await(t, c.id, "checking"); got != "matching" {
					t.Fatalf("stop at point %d: the restart reached %q, want matching", n, got)
				}
				if got := albumKeyOf(t, g.h, c.id); !bytes.Equal(got, want) {
					t.Fatalf("stop at point %d: album key %x, want the uninterrupted %x", n, got, want)
				}
				if err := g.in.Close(); err != nil {
					t.Fatal(err)
				}
				assertStatesMatchEvents(t, g.h)
			}
			if stopped < 8 {
				t.Errorf("only %d of %d stops landed before the verdict committed, too few to reach mid-decode", stopped, points)
			}
		})
	}
}

// TestCheckRejectsAnArchiveItCannotOpen: an archive archive/zip cannot
// read is the member's, unpack-failed on the submission and its archive.
func TestCheckRejectsAnArchiveItCannotOpen(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	garbage := bytes.Repeat([]byte("not a zip "), 50)
	c := f.checkingWith(t, m, zipped(gate.UnitAlbum, garbage), garbage)

	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "rejected" {
		t.Fatalf("state = %q, want rejected", got)
	}
	if _, reason := lastEvent(t, f.h, c.id); reason != string(reject.UnpackFailed) {
		t.Errorf("reason = %q, want unpack-failed", reason)
	}
	if got := failedFiles(t, f.h, c.id); got["music.zip"] != "unpack-failed" {
		t.Errorf("file verdicts = %v, want the archive unpack-failed", got)
	}
	if f.exists(t, tusOf(c.files[0].id)) {
		t.Error("a rejected archive is still staged")
	}
	assertStatesMatchEvents(t, f.h)
}

// TestWorkersCheckEveryQueuedSubmission: four workers started over six
// checking submissions bring each to exactly one verdict.
func TestWorkersCheckEveryQueuedSubmission(t *testing.T) {
	f := newIntakeFixture(t)
	f.in.cfg.CheckWorkers = 4
	m := member(t, f.h)
	var ids [][]byte
	for range 6 {
		tracks := twoTrackAlbum(t)
		ids = append(ids, f.checkingWith(t, m, albumOf(tracks...), tracks...).id)
	}
	if err := f.in.Start(t.Context()); err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if got := f.await(t, id, "checking"); got != "matching" {
			t.Errorf("state = %q, want matching", got)
		}
		if got := eventsOf(t, f.h, id); !slices.Equal(got, []string{"uploading", "checking", "matching"}) {
			t.Errorf("events = %q, want one verdict", got)
		}
	}
}
