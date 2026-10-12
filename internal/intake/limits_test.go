package intake

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"sync"
	"testing"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// override writes a member_overrides row; a nil argument stores NULL.
func override(t *testing.T, h *store.Handle, user identity.UserID, quota, mayUpload, sessionCap any) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO member_overrides (user_id, quota_bytes, may_upload, upload_concurrency_cap, updated_at)
			VALUES (?, ?, ?, ?, 1)`, user[:], quota, mayUpload, sessionCap)
		return err
	})
	if err != nil {
		t.Fatalf("write override: %v", err)
	}
}

func charge(t *testing.T, h *store.Handle, user identity.UserID, bytes int64) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO member_quota (user_id, charged_bytes, updated_at) VALUES (?, ?, 1)`, user[:], bytes)
		return err
	})
	if err != nil {
		t.Fatalf("write member_quota: %v", err)
	}
}

func setGroupTotal(t *testing.T, h *store.Handle, bytes int64) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE group_quota SET physical_total_bytes = ?`, bytes)
		return err
	})
	if err != nil {
		t.Fatalf("write group_quota: %v", err)
	}
}

func countSubmissions(t *testing.T, h *store.Handle) int {
	t.Helper()
	var n int
	if err := h.QueryRow(deadline(t), `SELECT count(*) FROM submissions`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// limitCode returns the code of a limit refusal, and fails the test on
// any other error.
func limitCode(t *testing.T, err error) reject.Code {
	t.Helper()
	if err == nil {
		return ""
	}
	var lr *limitRefusal
	if !errors.As(err, &lr) {
		t.Fatalf("error = %v, want a limit refusal", err)
	}
	return lr.code
}

func TestCreateWritesTheSubmission(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader, other := member(t, h), member(t, h)

	n := newSubmission{
		manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles, Source: "vinyl", Files: []gate.ManifestFile{
			{Path: "Disc 1/01.flac", Size: 100},
			{Path: "Disc 1/02.FLAC", Size: 200},
			{Path: "scans/front.jpg", Size: 30},
		}},
		sourceNote: "from the shop on 4th",
		storeURL:   "https://example.invalid/album",
	}
	c := mustCreate(t, s, uploader, n, testLimits())

	var gotUploader, coll []byte
	var unit, container, st, source, note, url string
	var declared, reserved, createdAt, updatedAt int64
	err := h.QueryRow(deadline(t), `SELECT uploader, collection_id, verdict_unit, container, state, source,
		source_note, store_url, declared_bytes, reserved_bytes, created_at, updated_at FROM submissions WHERE id = ?`, c.id).
		Scan(&gotUploader, &coll, &unit, &container, &st, &source, &note, &url, &declared, &reserved, &createdAt, &updatedAt)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(gotUploader, uploader[:]) || bytes.Equal(gotUploader, other[:]) {
		t.Errorf("uploader = %x, want the caller %x", gotUploader, uploader)
	}
	if unit != "album" || container != "files" || st != "uploading" || source != "vinyl" {
		t.Errorf("row = %s, %s, %s, %s; want album, files, uploading, vinyl", unit, container, st, source)
	}
	if note != "from the shop on 4th" || url != "https://example.invalid/album" {
		t.Errorf("note %q, url %q", note, url)
	}
	if declared != 330 || reserved != 330 {
		t.Errorf("declared %d, reserved %d; want 330 each", declared, reserved)
	}
	if createdAt != testNow || updatedAt != testNow {
		t.Errorf("stamps %d, %d; want %d", createdAt, updatedAt, testNow)
	}
	if got := readCollection(t, h, coll); got.name != "Library" || got.route != "default" {
		t.Errorf("collection = %q on %q, want Library on default", got.name, got.route)
	}

	rows, err := h.Query(deadline(t), `SELECT id, rel_path, size, role, received_at FROM submission_files
		WHERE submission_id = ? ORDER BY rel_path`, c.id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	type file struct {
		path, role string
		size       int64
	}
	var files []file
	for rows.Next() {
		var id, path []byte
		var f file
		var received *int64
		if err := rows.Scan(&id, &path, &f.size, &f.role, &received); err != nil {
			t.Fatal(err)
		}
		if received != nil {
			t.Errorf("%s received at create", path)
		}
		i := slices.IndexFunc(c.files, func(cf createdFile) bool { return cf.path == string(path) })
		if i < 0 || !bytes.Equal(c.files[i].id, id) {
			t.Errorf("%s: row id %x is not the id create returned", path, id)
		}
		f.path = string(path)
		files = append(files, f)
	}
	want := []file{
		{path: "Disc 1/01.flac", role: "audio", size: 100},
		{path: "Disc 1/02.FLAC", role: "audio", size: 200},
		{path: "scans/front.jpg", role: "extra", size: 30},
	}
	if !slices.Equal(files, want) {
		t.Errorf("files = %+v, want %+v", files, want)
	}
	if got := eventsOf(t, h, c.id); !slices.Equal(got, []string{"uploading"}) {
		t.Errorf("events = %q, want one uploading", got)
	}

	// A second album reuses the Library; separate tracks and a ZIP of
	// them take the unlisted collection, and the archive's role is the
	// server's.
	second := mustCreate(t, s, other, album(5), testLimits())
	trk := mustCreate(t, s, other, zipOf(gate.UnitTrack, 9), testLimits())
	var secondColl, trackColl []byte
	var role string
	if err := h.QueryRow(deadline(t), `SELECT collection_id FROM submissions WHERE id = ?`, second.id).Scan(&secondColl); err != nil {
		t.Fatal(err)
	}
	if err := h.QueryRow(deadline(t), `SELECT s.collection_id, f.role FROM submissions s
		JOIN submission_files f ON f.submission_id = s.id WHERE s.id = ?`, trk.id).Scan(&trackColl, &role); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(secondColl, coll) {
		t.Error("a second album made a second collection")
	}
	if got := readCollection(t, h, trackColl); got.name != "Unlisted music" || role != "archive" {
		t.Errorf("tracks landed in %q with role %q, want Unlisted music and archive", got.name, role)
	}
	assertStatesMatchEvents(t, h)
}

// TestCreateStoresEverySource is AC 14's store half: each of the seven
// answers is stored, and an absent, empty, or unknown one is refused
// with missing-source.
func TestCreateStoresEverySource(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader := member(t, h)
	for _, source := range []string{"bought", "cd", "vinyl", "free-download", "own-recording", "gift", "other"} {
		n := album(10)
		n.manifest.Source = source
		c := mustCreate(t, s, uploader, n, testLimits())
		var got string
		if err := h.QueryRow(deadline(t), `SELECT source FROM submissions WHERE id = ?`, c.id).Scan(&got); err != nil {
			t.Fatal(err)
		}
		if got != source {
			t.Errorf("stored source %q, want %q", got, source)
		}
	}
	stored := countSubmissions(t, h)
	for _, source := range []string{"", "Bought", "shop"} {
		n := album(10)
		n.manifest.Source = source
		_, err := s.create(deadline(t), uploader, n, testLimits())
		var mr *manifestRefusal
		if !errors.As(err, &mr) || !slices.ContainsFunc(mr.refusal.Submission, func(r reject.Reason) bool {
			return r.Code == reject.MissingSource
		}) {
			t.Errorf("source %q: error = %v, want a manifest refusal with missing-source", source, err)
		}
	}
	if n := countSubmissions(t, h); n != stored {
		t.Errorf("%d submissions after the refusals, want %d", n, stored)
	}
}

// TestCreateAnswersEveryManifestReason holds that a refused manifest
// reaches no limit and writes nothing, and that a malformed one is an
// error the seam can tell apart.
func TestCreateAnswersEveryManifestReason(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader := member(t, h)
	override(t, h, uploader, nil, 0, nil) // uploads off: a limit the manifest must precede

	n := newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles, Source: "", Files: []gate.ManifestFile{
		{Path: "../01.flac", Size: 1},
		{Path: "02.mp3", Size: 1},
		{Path: "03.flac", Size: 1},
	}}}
	_, err := s.create(deadline(t), uploader, n, testLimits())
	var mr *manifestRefusal
	if !errors.As(err, &mr) {
		t.Fatalf("error = %v, want a manifest refusal", err)
	}
	var files []string
	for _, f := range mr.refusal.Files {
		files = append(files, string(f.Path)+" "+string(f.Reason.Code))
	}
	if want := []string{"../01.flac unsafe-path", "02.mp3 bad-extension"}; !slices.Equal(files, want) {
		t.Errorf("file reasons = %q, want %q", files, want)
	}
	if len(mr.refusal.Submission) != 1 || mr.refusal.Submission[0].Code != reject.MissingSource {
		t.Errorf("submission reasons = %+v, want missing-source alone", mr.refusal.Submission)
	}

	dup := album(1, 1)
	dup.manifest.Files[1].Path = dup.manifest.Files[0].Path
	if _, err := s.create(deadline(t), uploader, dup, testLimits()); !errors.Is(err, gate.ErrMalformedManifest) {
		t.Errorf("duplicate path: error = %v, want gate.ErrMalformedManifest", err)
	}
	if n := countSubmissions(t, h); n != 0 {
		t.Errorf("%d submissions written by refused creates, want 0", n)
	}
}

// TestCreateReadsTheCollectionsAllowlist narrows the bound by the route
// collection's own allowlist, as the effective allowlist is defined.
func TestCreateReadsTheCollectionsAllowlist(t *testing.T) {
	h, s := newTestSubmissions(t)
	uploader := member(t, h)
	mustCreate(t, s, uploader, album(1), testLimits())
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `UPDATE collections SET admission_format_allowlist = 'wav' WHERE intake_route = 'default'`)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.create(deadline(t), uploader, album(1), testLimits())
	var mr *manifestRefusal
	if !errors.As(err, &mr) || len(mr.refusal.Files) != 1 || mr.refusal.Files[0].Reason.Code != reject.BadExtension {
		t.Fatalf("error = %v, want bad-extension for a format the collection does not admit", err)
	}
}

// TestCreateLimits runs each limit in its order: the first failing one
// answers, and a free-space read that failed answers only once the
// four before it pass.
func TestCreateLimits(t *testing.T) {
	failedRead := errors.New("statfs: input/output error")
	tests := []struct {
		name string
		// setup prepares the member and the store and returns the limits.
		setup   func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits
		free    uint64
		freeErr error
		size    int64
		want    reject.Code
	}{
		{name: "uploads off, and over every other limit", size: 900, free: 10, freeErr: failedRead, want: reject.UploadsOff,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, 1, 0, 1)
				return testLimits()
			}},
		{name: "may_upload NULL is allowed", size: 10, free: plentyFree,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, nil, nil, nil)
				return testLimits()
			}},
		{name: "the override's session cap", size: 10, free: plentyFree, want: reject.TooManySessions,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, nil, 1, 1)
				mustCreate(t, s, m, album(1), testLimits())
				return testLimits()
			}},
		{name: "the setting's session cap when the override is NULL", size: 10, free: plentyFree, want: reject.TooManySessions,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, nil, 1, nil)
				mustCreate(t, s, m, album(1), testLimits())
				mustCreate(t, s, m, album(1), testLimits())
				lim := testLimits()
				lim.sessionsPerMember = 2
				return lim
			}},
		{name: "a closed upload frees its session", size: 10, free: plentyFree,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				inState(t, s, m, stateChecking, 1)
				inState(t, s, m, stateCancelled, 1)
				lim := testLimits()
				lim.sessionsPerMember = 1
				return lim
			}},
		{name: "member quota, charged bytes counted", size: 41, free: plentyFree, want: reject.OverQuota,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, 100, nil, nil)
				charge(t, h, m, 50)
				mustCreate(t, s, m, album(10), testLimits())
				return testLimits()
			}},
		{name: "member quota exactly met", size: 40, free: plentyFree,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, 100, nil, nil)
				charge(t, h, m, 50)
				mustCreate(t, s, m, album(10), testLimits())
				return testLimits()
			}},
		{name: "member quota over a failed free-space read", size: 101, freeErr: failedRead, want: reject.OverQuota,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				override(t, h, m, 100, nil, nil)
				return testLimits()
			}},
		{name: "group cap counts every member's open reservation", size: 300, free: plentyFree, want: reject.OverQuota,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				setGroupTotal(t, h, 500)
				inState(t, s, member(t, h), stateAccepted, 201)
				lim := testLimits()
				lim.groupQuotaBytes = 1000
				return lim
			}},
		{name: "group cap ignores a closed reservation", size: 300, free: plentyFree,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				setGroupTotal(t, h, 500)
				inState(t, s, member(t, h), stateRejected, 201)
				lim := testLimits()
				lim.groupQuotaBytes = 1000
				return lim
			}},
		{name: "group cap of zero is no cap", size: 300, free: plentyFree,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				setGroupTotal(t, h, 5000)
				return testLimits()
			}},
		{name: "the floor counts other uploads' unwritten bytes", size: 300, free: 1000, want: reject.NoSpace,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				// 100 of this upload's 150 are received, so 50 are unwritten.
				c := mustCreate(t, s, member(t, h), album(100, 50), testLimits())
				if err := s.receive(deadline(t), c.files[0].id); err != nil {
					t.Fatal(err)
				}
				lim := testLimits()
				lim.freeSpaceFloor = 651
				return lim
			}},
		{name: "the floor exactly met", size: 300, free: 1000,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				c := mustCreate(t, s, member(t, h), album(100, 50), testLimits())
				if err := s.receive(deadline(t), c.files[0].id); err != nil {
					t.Fatal(err)
				}
				lim := testLimits()
				lim.freeSpaceFloor = 650
				return lim
			}},
		{name: "a failed free-space read once the others pass", size: 10, freeErr: failedRead, want: reject.CouldNotCheck,
			setup: func(t *testing.T, h *store.Handle, s *submissions, m identity.UserID) limits {
				return testLimits()
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s := newTestSubmissions(t)
			m := member(t, h)
			lim := tt.setup(t, h, s, m)
			s.freeBytes = func(string) (uint64, error) { return tt.free, tt.freeErr }
			before := countSubmissions(t, h)

			_, err := s.create(deadline(t), m, album(tt.size), lim)

			if got := limitCode(t, err); got != tt.want {
				t.Fatalf("limit = %q, want %q", got, tt.want)
			}
			wantCount := before + 1
			if tt.want != "" {
				wantCount = before
			}
			if n := countSubmissions(t, h); n != wantCount {
				t.Errorf("%d submissions, want %d", n, wantCount)
			}
			if tt.want == reject.CouldNotCheck && !errors.Is(err, failedRead) {
				t.Errorf("could-not-check is %v, want it to carry the read's own error", err)
			}
			if wantCount > 0 {
				assertStatesMatchEvents(t, h)
			}
		})
	}
}

// raceCreates runs two creates for uploaders a and b whose free-space
// reads each block until both have reached them, so both creates hold
// their read before either Update runs. It returns both errors.
func raceCreates(t *testing.T, s *submissions, a, b identity.UserID, n newSubmission, lim limits, free uint64) [2]error {
	t.Helper()
	var reached sync.WaitGroup
	reached.Add(2)
	s.freeBytes = func(string) (uint64, error) {
		reached.Done()
		reached.Wait()
		return free, nil
	}
	var errs [2]error
	var done sync.WaitGroup
	for i, uploader := range []identity.UserID{a, b} {
		done.Go(func() {
			_, errs[i] = s.create(deadline(t), uploader, n, lim)
		})
	}
	done.Wait()
	return errs
}

// TestConcurrentCreatesCannotOvershoot is AC 15's race for each of the
// member cap, the group cap, and the floor: two creates each fit, both
// reach the free-space read before either writes, and together they
// exceed the limit, so exactly one commits.
func TestConcurrentCreatesCannotOvershoot(t *testing.T) {
	tests := []struct {
		name string
		// setup returns the two racing members and the limits.
		setup func(t *testing.T, h *store.Handle, s *submissions) (a, b identity.UserID, lim limits)
		size  int64
		free  uint64
		want  reject.Code
	}{
		{name: "member cap", size: 60, free: plentyFree, want: reject.OverQuota,
			setup: func(t *testing.T, h *store.Handle, s *submissions) (identity.UserID, identity.UserID, limits) {
				m := member(t, h)
				override(t, h, m, 100, nil, nil)
				return m, m, testLimits()
			}},
		{name: "group cap with an accepted reservation in the sum", size: 400, free: plentyFree, want: reject.OverQuota,
			setup: func(t *testing.T, h *store.Handle, s *submissions) (identity.UserID, identity.UserID, limits) {
				setGroupTotal(t, h, 100)
				inState(t, s, member(t, h), stateAccepted, 300)
				lim := testLimits()
				lim.groupQuotaBytes = 1000
				return member(t, h), member(t, h), lim
			}},
		{name: "floor", size: 500, free: 1000, want: reject.NoSpace,
			setup: func(t *testing.T, h *store.Handle, s *submissions) (identity.UserID, identity.UserID, limits) {
				lim := testLimits()
				lim.freeSpaceFloor = 200
				return member(t, h), member(t, h), lim
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s := newTestSubmissions(t)
			a, b, lim := tt.setup(t, h, s)
			before := countSubmissions(t, h)

			errs := raceCreates(t, s, a, b, album(tt.size), lim, tt.free)

			var codes []reject.Code
			for _, err := range errs {
				codes = append(codes, limitCode(t, err))
			}
			slices.Sort(codes)
			if want := []reject.Code{"", tt.want}; !slices.Equal(codes, want) {
				t.Errorf("outcomes = %q, want one commit and one %q", codes, tt.want)
			}
			if n := countSubmissions(t, h); n != before+1 {
				t.Errorf("%d submissions, want %d", n, before+1)
			}
			assertStatesMatchEvents(t, h)
		})
	}
}

// TestUncappedMembersAreNotBoundByAnothersCap is AC 15's NULL-cap
// clause: a member with no override row, one whose cap is NULL, and one
// with no member_quota row each create more than the capped member's
// quota, and every create commits.
func TestUncappedMembersAreNotBoundByAnothersCap(t *testing.T) {
	h, s := newTestSubmissions(t)
	capped := member(t, h)
	override(t, h, capped, 100, nil, nil)
	charge(t, h, capped, 0)

	noOverride, nullCap, noQuotaRow := member(t, h), member(t, h), member(t, h)
	charge(t, h, noOverride, 10)
	override(t, h, nullCap, nil, 1, nil)
	charge(t, h, nullCap, 10)
	override(t, h, noQuotaRow, nil, nil, nil)

	for _, m := range []identity.UserID{noOverride, nullCap, noQuotaRow} {
		for range 3 {
			mustCreate(t, s, m, album(60), testLimits())
		}
	}
	if n := countSubmissions(t, h); n != 9 {
		t.Errorf("%d submissions, want 9", n)
	}
	assertStatesMatchEvents(t, h)
}

// TestUseCountsOpenSubmissions is AC 16's first clause: with one
// submission per state, each reserving a distinct power of two, a
// member's use is their charged bytes plus exactly the open ones.
func TestUseCountsOpenSubmissions(t *testing.T) {
	h, s := newTestSubmissions(t)
	m := member(t, h)
	charge(t, h, m, 1000)
	for i, st := range states {
		inState(t, s, m, st, int64(1)<<i)
	}

	var use int64
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		var err error
		use, err = memberUse(ctx, tx, m, nil)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	// uploading 1, checking 2, check-unfinished 4, matching 8, accepted 16;
	// rejected 32, cancelled 64, and expired 128 stop counting.
	if use != 1000+1+2+4+8+16 {
		t.Errorf("use = %d, want 1031", use)
	}

	override(t, h, m, 1100, nil, nil)
	if _, err := s.create(deadline(t), m, album(70), testLimits()); limitCode(t, err) != reject.OverQuota {
		t.Errorf("a create one byte over the quota: error = %v, want over-quota", err)
	}
	mustCreate(t, s, m, album(69), testLimits())
	assertStatesMatchEvents(t, h)
}

// zipChecking creates a ZIP submission of archive size for m and
// receives the archive, so it sits in checking awaiting its extraction.
func zipChecking(t *testing.T, s *submissions, m identity.UserID, unit gate.Unit, size int64) created {
	t.Helper()
	c := mustCreate(t, s, m, zipOf(unit, size), testLimits())
	receiveAll(t, s, c)
	return c
}

func reservedOf(t *testing.T, h *store.Handle, id []byte) int64 {
	t.Helper()
	var n int64
	if err := h.QueryRow(deadline(t), `SELECT reserved_bytes FROM submissions WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// TestReserveExtracted is AC 16's ZIP clauses: a deflated archive that
// fits whose extracted total does not is rejected before extraction,
// and a passing one's reservation becomes its extracted total.
func TestReserveExtracted(t *testing.T) {
	failedRead := errors.New("statfs: input/output error")
	tests := []struct {
		name    string
		quota   any
		floor   int64
		free    uint64
		freeErr error
		other   int64 // another open upload's unwritten bytes, 0 for none
		want    reject.Code
		wantErr error
		state   state
		// reserved is the reservation after the call.
		reserved int64
	}{
		{name: "fits", quota: 50, free: plentyFree, state: stateChecking, reserved: 30},
		{name: "member cap", quota: 25, free: plentyFree, want: reject.OverQuota, state: stateRejected, reserved: 10},
		{name: "floor, counting another upload's unwritten bytes", floor: 900, free: 1000, other: 71,
			want: reject.NoSpace, state: stateRejected, reserved: 10},
		{name: "floor met", floor: 900, free: 1000, other: 70, state: stateChecking, reserved: 30},
		{name: "a failed read", free: 0, freeErr: failedRead, wantErr: gate.ErrCouldNotRun, state: stateChecking, reserved: 10},
		{name: "member cap over a failed read", quota: 25, freeErr: failedRead, want: reject.OverQuota, state: stateRejected, reserved: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, s := newTestSubmissions(t)
			m := member(t, h)
			override(t, h, m, tt.quota, nil, nil)
			c := zipChecking(t, s, m, gate.UnitAlbum, 10)
			if tt.other > 0 {
				mustCreate(t, s, member(t, h), album(tt.other), testLimits())
			}
			lim := testLimits()
			lim.freeSpaceFloor = tt.floor
			s.freeBytes = func(string) (uint64, error) { return tt.free, tt.freeErr }

			code, err := s.reserveExtracted(deadline(t), c.id, 30, lim)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
			if code != tt.want {
				t.Errorf("code = %q, want %q", code, tt.want)
			}
			if got := stateOf(t, h, c.id); got != string(tt.state) {
				t.Errorf("state = %q, want %q", got, tt.state)
			}
			if got := reservedOf(t, h, c.id); got != tt.reserved {
				t.Errorf("reserved = %d, want %d", got, tt.reserved)
			}
			if tt.state == stateRejected {
				var reason string
				if err := h.QueryRow(deadline(t), `SELECT reason FROM submission_events WHERE submission_id = ?
					ORDER BY seq DESC LIMIT 1`, c.id).Scan(&reason); err != nil {
					t.Fatal(err)
				}
				if reason != string(tt.want) {
					t.Errorf("rejected with reason %q, want %q", reason, tt.want)
				}
			}
			assertStatesMatchEvents(t, h)
		})
	}
}
