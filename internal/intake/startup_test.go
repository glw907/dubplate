package intake

import (
	"bytes"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
)

// TestStartRefusedAtTheLease is AC 27's first clause: while another
// process holds the intake lease, Start refuses naming it, and no
// recovery runs: an orphaned submissions/ directory and a checking
// submission whose bytes are missing are both left as they were.
func TestStartRefusedAtTheLease(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	tracks := twoTrackAlbum(t)
	c := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	if err := os.Remove(f.path(tusOf(c.files[0].id))); err != nil {
		t.Fatal(err)
	}
	orphan := "submissions/" + wireID(store.NewID()) + "/x/leftover"
	f.write(t, orphan, []byte("bytes"))

	held, err := f.h.TakeKindLease("intake")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = held.Release() }()

	err = f.in.Start(t.Context())
	if !errors.Is(err, store.ErrLeaseBusy) {
		t.Fatalf("Start under a held lease = %v, want ErrLeaseBusy", err)
	}
	if !f.exists(t, orphan) {
		t.Error("the orphaned directory was removed before the lease was taken")
	}
	if got := stateOf(t, f.h, c.id); got != "checking" {
		t.Errorf("state = %q, want checking: reconcile ran before the lease", got)
	}
}

// TestStartupExpiresWhatARestoreLeftWithoutBytes is AC 27's restored
// store: an accepted and a checking submission over empty staging both
// expire with a reason naming their files, which ends their
// reservations, and a re-upload of the accepted tracks is not refused
// as a twin.
func TestStartupExpiresWhatARestoreLeftWithoutBytes(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)
	track := cdTrack(7, 1).encode(t)
	accepted := f.checkingWith(t, m, tracksOf(track), track)
	f.in.check(deadline(t), accepted.id)
	if got := stateOf(t, f.h, accepted.id); got != "accepted" {
		t.Fatalf("fixture state %q, want accepted", got)
	}
	tracks := twoTrackAlbum(t)
	checking := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	if err := os.RemoveAll(f.path("tus")); err != nil {
		t.Fatal(err)
	}

	if err := f.in.startup(deadline(t)); err != nil {
		t.Fatalf("startup: %v", err)
	}

	for _, c := range []created{accepted, checking} {
		if got := stateOf(t, f.h, c.id); got != "expired" {
			t.Errorf("state = %q, want expired", got)
		}
		if _, reason := lastEvent(t, f.h, c.id); !strings.Contains(reason, c.files[0].path) {
			t.Errorf("expiry reason %q names no missing file", reason)
		}
	}
	if use, err := memberUseOf(t, f.h, m); err != nil || use != 0 {
		t.Errorf("member use = %d, %v after both expired, want 0", use, err)
	}

	again := f.checkingWith(t, m, tracksOf(track), track)
	f.in.check(deadline(t), again.id)
	if got := stateOf(t, f.h, again.id); got != "accepted" {
		t.Errorf("re-upload = %q with verdicts %v, want accepted", got, fileVerdicts(t, f.h, again.id))
	}
	assertStatesMatchEvents(t, f.h)
}

// TestStartupReconcilesArchivesByTheirStagedSet is AC 27's two ZIPs and
// the truncated member: a ZIP past its verdict is tested against its
// members and never its archive, by size as well as presence, and a ZIP
// in check-unfinished keeps its archive.
func TestStartupReconcilesArchivesByTheirStagedSet(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)

	accepted := twoTrackZIP(t)
	acc := f.checkingWith(t, m, zipped(gate.UnitTrack, accepted), accepted)
	f.in.check(deadline(t), acc.id)
	if got := stateOf(t, f.h, acc.id); got != "accepted" || f.exists(t, tusOf(acc.files[0].id)) {
		t.Fatalf("fixture: accepted ZIP in %q, archive still staged %v", got, f.exists(t, tusOf(acc.files[0].id)))
	}
	accMembers := membersOf(t, f.h, acc.id)

	truncated := twoTrackZIP(t)
	tr := f.checkingWith(t, m, zipped(gate.UnitAlbum, truncated), truncated)
	f.in.check(deadline(t), tr.id)
	if got := stateOf(t, f.h, tr.id); got != "matching" {
		t.Fatalf("fixture: ZIP album in %q, want matching", got)
	}
	member := xOf(tr.id, membersOf(t, f.h, tr.id)["02.flac"])
	whole := f.read(t, member)
	f.write(t, member, whole[:len(whole)-10])

	held := twoTrackZIP(t)
	cu := f.checkingWith(t, m, zipped(gate.UnitAlbum, held), held)
	f.s.freeBytes = func(string) (uint64, error) { return 0, errors.New("statfs: input/output error") }
	f.in.check(deadline(t), cu.id)
	f.s.freeBytes = func(string) (uint64, error) { return plentyFree, nil }
	if got := stateOf(t, f.h, cu.id); got != "check-unfinished" {
		t.Fatalf("fixture: ZIP in %q, want check-unfinished", got)
	}

	before := map[string][]byte{}
	for path, id := range accMembers {
		before[path] = f.read(t, xOf(acc.id, id))
	}

	if err := f.in.startup(deadline(t)); err != nil {
		t.Fatalf("startup: %v", err)
	}

	if got := stateOf(t, f.h, tr.id); got != "expired" {
		t.Errorf("a matching ZIP with a truncated member = %q, want expired", got)
	}
	if got := stateOf(t, f.h, acc.id); got != "accepted" {
		t.Errorf("an accepted ZIP whose archive is gone = %q, want accepted", got)
	}
	for path, id := range accMembers {
		if !bytes.Equal(f.read(t, xOf(acc.id, id)), before[path]) {
			t.Errorf("accepted member %s changed or gone", path)
		}
	}
	if got := stateOf(t, f.h, cu.id); got != "check-unfinished" {
		t.Errorf("a check-unfinished ZIP = %q, want check-unfinished", got)
	}
	if !bytes.Equal(f.read(t, tusOf(cu.files[0].id)), held) {
		t.Error("a check-unfinished ZIP's archive changed or gone")
	}
	assertStatesMatchEvents(t, f.h)
}

// TestStartupCleansUpAndReceives covers the startup pass's other steps:
// a submissions/ directory with no row and a rejected submission's
// leftover bytes go; an unreceived upload whose .info is torn loses both
// files, so the client's HEAD gets 404 and creates it again; and an
// upload whose last byte landed with no receipt is received, moving its
// submission to checking (AC 11's restart clause).
func TestStartupCleansUpAndReceives(t *testing.T) {
	f := newIntakeFixture(t)
	m := member(t, f.h)

	orphan := "submissions/" + wireID(store.NewID())
	f.write(t, orphan+"/x/leftover", []byte("bytes"))

	tracks := twoTrackAlbum(t)
	tracks[0] = flipped(tracks[0])
	rejected := f.checkingWith(t, m, albumOf(tracks...), tracks...)
	f.in.check(deadline(t), rejected.id)
	if got := stateOf(t, f.h, rejected.id); got != "rejected" {
		t.Fatalf("fixture state %q, want rejected", got)
	}
	f.stageUpload(t, rejected.files[0].id, tracks[0])

	torn := mustCreate(t, f.s, m, album(10), testLimits())
	f.write(t, tusOf(torn.files[0].id), []byte("abc"))
	f.write(t, tusOf(torn.files[0].id)+".info", nil)

	whole := mustCreate(t, f.s, m, album(5), testLimits())
	f.stageUpload(t, whole.files[0].id, []byte("abcde"))

	if err := f.in.startup(deadline(t)); err != nil {
		t.Fatalf("startup: %v", err)
	}

	if f.exists(t, orphan) {
		t.Error("a submissions/ directory with no row survived")
	}
	if f.exists(t, tusOf(rejected.files[0].id)) || f.exists(t, tusOf(rejected.files[0].id)+".info") {
		t.Error("a rejected submission's leftover bytes survived")
	}
	if f.exists(t, tusOf(torn.files[0].id)) || f.exists(t, tusOf(torn.files[0].id)+".info") {
		t.Error("an upload with a torn .info kept its files")
	}
	if got := stateOf(t, f.h, torn.id); got != "uploading" {
		t.Errorf("the torn upload's submission = %q, want uploading", got)
	}
	if at := receivedAt(t, f.s, whole.files[0].id); at == nil {
		t.Error("a whole upload with no receipt is still unreceived")
	}
	if got := stateOf(t, f.h, whole.id); got != "checking" {
		t.Errorf("its submission = %q, want checking", got)
	}
	assertStatesMatchEvents(t, f.h)
}
