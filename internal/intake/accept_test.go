package intake

import (
	"bytes"
	"context"
	"maps"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// key is a 16-byte audio key, or a 32-byte album key, filled with b.
func key(n int, b byte) []byte { return bytes.Repeat([]byte{b}, n) }

// lastReason is the reason on a submission's latest event.
func lastReason(t *testing.T, h *store.Handle, id []byte) string {
	t.Helper()
	var reason *string
	if err := h.QueryRow(deadline(t), `SELECT reason FROM submission_events WHERE submission_id = ?
		ORDER BY seq DESC LIMIT 1`, id).Scan(&reason); err != nil {
		t.Fatal(err)
	}
	if reason == nil {
		return ""
	}
	return *reason
}

// fileVerdicts maps each of a submission's file paths to its stored
// verdict, "" for a pass.
func fileVerdicts(t *testing.T, h *store.Handle, id []byte) map[string]string {
	t.Helper()
	rows, err := h.Query(deadline(t), `SELECT rel_path, coalesce(verdict, '') FROM submission_files WHERE submission_id = ?`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	got := map[string]string{}
	for rows.Next() {
		var path []byte
		var v string
		if err := rows.Scan(&path, &v); err != nil {
			t.Fatal(err)
		}
		got[string(path)] = v
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return got
}

// checkedAlbum creates an album for m, receives it, and returns its
// passing verdict to state to carrying albumKey.
func checkedAlbum(t *testing.T, s *submissions, m identity.UserID, to state, albumKey []byte) (created, verdict) {
	t.Helper()
	c := mustCreate(t, s, m, album(10, 20), testLimits())
	receiveAll(t, s, c)
	v := passing(c, to)
	v.albumKey = albumKey
	return c, v
}

// TestAcceptRefusesADuplicateAlbum is AC 23's album clauses: the same
// album accepted twice into one collection is refused the second time,
// one waiting in matching blocks nothing, and the same album accepted
// into the unlisted collection is not a duplicate of the Library's.
func TestAcceptRefusesADuplicateAlbum(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	k := key(32, 0xa1)

	waiting, v := checkedAlbum(t, s, m, stateMatching, k)
	if err := s.commitVerdict(deadline(t), waiting.id, v); err != nil {
		t.Fatal(err)
	}
	first, v := checkedAlbum(t, s, m, stateAccepted, k)
	if err := s.commitVerdict(deadline(t), first.id, v); err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, h, first.id); got != "accepted" {
		t.Fatalf("first album = %q, want accepted past a twin still in matching", got)
	}
	var stored []byte
	if err := h.QueryRow(deadline(t), `SELECT album_key FROM submissions WHERE id = ?`, first.id).Scan(&stored); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(stored, k) {
		t.Errorf("album_key = %x, want %x", stored, k)
	}

	second, v := checkedAlbum(t, s, m, stateAccepted, k)
	if err := s.commitVerdict(deadline(t), second.id, v); err != nil {
		t.Fatal(err)
	}
	if got, reason := stateOf(t, h, second.id), lastReason(t, h, second.id); got != "rejected" || reason != "already-in-library" {
		t.Errorf("second album = %q (%q), want rejected already-in-library", got, reason)
	}

	// The same album on the unlisted route, its collection set by direct
	// write because the album's unlisted route lands with intake-2.
	unlisted, v := checkedAlbum(t, s, m, stateAccepted, k)
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		coll, err := RouteCollection(ctx, tx, s.log, RouteUnlisted)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE submissions SET collection_id = ? WHERE id = ?`, coll, unlisted.id)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.commitVerdict(deadline(t), unlisted.id, v); err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, h, unlisted.id); got != "accepted" {
		t.Errorf("the unlisted twin = %q, want accepted", got)
	}
	assertStatesMatchEvents(t, h)
}

// tracksVerdict creates separate tracks of sizes for m, receives them,
// and returns an accepted verdict whose file i carries audio key keys[i]
// and verdict codes[i].
func tracksVerdict(t *testing.T, s *submissions, m identity.UserID, keys [][]byte, codes []reject.Code, sizes ...int64) (created, verdict) {
	t.Helper()
	c := mustCreate(t, s, m, tracks(sizes...), testLimits())
	receiveAll(t, s, c)
	v := passing(c, stateAccepted)
	for i := range v.files {
		v.files[i].result.AudioKey = keys[i]
		v.files[i].result.Verdict = codes[i]
	}
	return c, v
}

// TestAcceptRefusesDuplicateTracks is AC 22's accept half and AC 23's
// track clause: a track whose audio key equals a passing track of an
// accepted submission in the unlisted collection gets
// already-in-library, the others go on, the reservation trues down to
// what stays, and a submission whose every passing track is refused ends
// rejected.
func TestAcceptRefusesDuplicateTracks(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	k1, k2, k3, k4 := key(16, 1), key(16, 2), key(16, 3), key(16, 4)

	twin, v := tracksVerdict(t, s, m, [][]byte{k1, k2, k4}, []reject.Code{"", "", reject.Corrupt}, 10, 20, 30)
	if err := s.commitVerdict(deadline(t), twin.id, v); err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, h, twin.id); got != "accepted" {
		t.Fatalf("the first twin = %q, want accepted", got)
	}
	if got := reservedOf(t, h, twin.id); got != 30 {
		t.Errorf("first twin reserves %d, want its two passing tracks' 30", got)
	}

	mixed, v := tracksVerdict(t, s, m, [][]byte{k1, k3, k4}, []reject.Code{"", "", reject.Corrupt}, 10, 20, 30)
	if err := s.commitVerdict(deadline(t), mixed.id, v); err != nil {
		t.Fatal(err)
	}
	if got := stateOf(t, h, mixed.id); got != "accepted" {
		t.Errorf("mixed = %q, want accepted with its one new track", got)
	}
	want := map[string]string{"01.flac": "already-in-library", "02.flac": "", "03.flac": "corrupt"}
	if got := fileVerdicts(t, h, mixed.id); !maps.Equal(got, want) {
		t.Errorf("mixed verdicts = %v, want %v (a failing twin's key refuses nothing)", got, want)
	}
	if got := reservedOf(t, h, mixed.id); got != 20 {
		t.Errorf("mixed reserves %d, want the one kept track's 20", got)
	}

	allTwins, v := tracksVerdict(t, s, m, [][]byte{k2, k1}, []reject.Code{"", ""}, 10, 20)
	if err := s.commitVerdict(deadline(t), allTwins.id, v); err != nil {
		t.Fatal(err)
	}
	if got, reason := stateOf(t, h, allTwins.id), lastReason(t, h, allTwins.id); got != "rejected" || reason != "already-in-library" {
		t.Errorf("all twins = %q (%q), want rejected already-in-library", got, reason)
	}
	want = map[string]string{"01.flac": "already-in-library", "02.flac": "already-in-library"}
	if got := fileVerdicts(t, h, allTwins.id); !maps.Equal(got, want) {
		t.Errorf("all-twins verdicts = %v, want %v", got, want)
	}
	assertStatesMatchEvents(t, h)
}

// TestVerdictWritesMembersAndOnlyLowersTheReservation commits a ZIP
// album's verdict: the extracted members get rows with server roles and
// the gate's fields, the warnings land with their subjects, and a
// reservation below the kept total is not raised.
func TestVerdictWritesMembersAndOnlyLowersTheReservation(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	c := zipChecking(t, s, m, gate.UnitAlbum, 10)

	audioID, extraID := store.NewID(), store.NewID()
	v := verdict{
		to:       stateMatching,
		albumKey: key(32, 7),
		members: []memberVerdict{
			{id: audioID, path: []byte("CD1/01 \xff.flac"), size: 25, result: gate.FileResult{
				SampleRate: 44100, BitsPerSample: 16, Channels: 2, Samples: 441000,
				AudioKey: key(16, 9), FileHash: key(32, 8), Disc: 1, Track: 1,
			}},
			{id: extraID, path: []byte("scan.jpg"), size: 5},
		},
		warnings: []gate.FileReason{
			{Reason: reject.Of(reject.NoCoverArt)},
			{Path: []byte("notes.txt"), Reason: reject.Of(reject.ExtrasDropped)},
		},
	}
	if err := s.commitVerdict(deadline(t), c.id, v); err != nil {
		t.Fatal(err)
	}

	if got := stateOf(t, h, c.id); got != "matching" {
		t.Errorf("state = %q, want matching", got)
	}
	if got := reservedOf(t, h, c.id); got != 10 {
		t.Errorf("reserved = %d, want 10: the true-up only lowers", got)
	}
	var role string
	var path, audioKey []byte
	var rate, depth, channels, samples, disc, track int64
	var hasMD5 bool
	err := h.QueryRow(deadline(t), `SELECT rel_path, role, audio_key, header_md5 IS NOT NULL, sample_rate,
		bits_per_sample, channels, samples, disc, track FROM submission_files WHERE id = ?`, audioID).
		Scan(&path, &role, &audioKey, &hasMD5, &rate, &depth, &channels, &samples, &disc, &track)
	if err != nil {
		t.Fatal(err)
	}
	if string(path) != "CD1/01 \xff.flac" || role != "audio" || !bytes.Equal(audioKey, key(16, 9)) || hasMD5 {
		t.Errorf("audio member = %q %s key %x md5 %v", path, role, audioKey, hasMD5)
	}
	if rate != 44100 || depth != 16 || channels != 2 || samples != 441000 || disc != 1 || track != 1 {
		t.Errorf("stream = %d/%d/%d/%d, position %d/%d", rate, depth, channels, samples, disc, track)
	}
	if err := h.QueryRow(deadline(t), `SELECT role FROM submission_files WHERE id = ?`, extraID).Scan(&role); err != nil {
		t.Fatal(err)
	}
	if role != "extra" {
		t.Errorf("scan.jpg role = %q, want extra", role)
	}

	rows, err := h.Query(deadline(t), `SELECT code, coalesce(subject, X'') FROM submission_warnings
		WHERE submission_id = ? ORDER BY code`, c.id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	var warnings []string
	for rows.Next() {
		var code string
		var subject []byte
		if err := rows.Scan(&code, &subject); err != nil {
			t.Fatal(err)
		}
		warnings = append(warnings, code+":"+string(subject))
	}
	if want := []string{"extras-dropped:notes.txt", "no-cover-art:"}; !slices.Equal(warnings, want) {
		t.Errorf("warnings = %q, want %q", warnings, want)
	}
	assertStatesMatchEvents(t, h)
}
