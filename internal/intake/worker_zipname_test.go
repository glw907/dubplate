package intake

import (
	"testing"

	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// TestCheckRefusesAMemberNamedLikeItsArchive: the archive's own row carries
// the manifest's path, and a member of the same name would collide with it on
// (submission_id, rel_path) at the verdict commit, which the worker reads as a
// store fault and retries forever. The check must refuse the archive instead,
// so the verdict commits and cleanup takes the bytes.
func TestCheckRefusesAMemberNamedLikeItsArchive(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	archive := zipOfMembers(t,
		zipMember{"01.flac", tracks[0]},
		zipMember{"02.flac", tracks[1]},
		zipMember{"music.zip", []byte("a nested archive")}, // the manifest names the archive music.zip
	)
	c := f.checkingWith(t, m, zipped(gate.UnitAlbum, archive), archive)

	f.in.check(deadline(t), c.id)

	if got := stateOf(t, f.h, c.id); got != "rejected" {
		t.Fatalf("state = %q (%s), want rejected, not wedged in checking", got, eventKinds(t, f.h, c.id))
	}
	if _, reason := lastEvent(t, f.h, c.id); reason != string(reject.UnsafePath) {
		t.Errorf("reason = %q, want unsafe-path", reason)
	}
	if f.exists(t, tusOf(c.files[0].id)) {
		t.Error("a rejected archive is still staged")
	}
	assertStatesMatchEvents(t, f.h)
}
