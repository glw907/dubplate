package intake

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake/gate"
)

// testClock is the one clock a fixture's writes and sweeps read, moved
// only by the test.
type testClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *testClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *testClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// testWindows are the sweep windows every fixture runs under.
var testWindows = Windows{ResumeWindow: 7 * 24 * time.Hour, Hold: 7 * 24 * time.Hour, StallAfter: 30 * time.Minute}

// sentMail is one stall mail the fake sent.
type sentMail struct {
	to, subject, body string
	// waited is whether a short write on a fresh context had to wait
	// for the store when the mail went: true only when a write
	// callback held the write connection.
	waited bool
}

// mailFake stands in for the admin roster and the mailer. Each send
// runs a short-deadline Update on a fresh context first, which waits
// only while another write callback holds the store.
type mailFake struct {
	h      *store.Handle
	admins []string

	mu   sync.Mutex
	sent []sentMail
}

func (m *mailFake) addresses(context.Context) ([]string, error) { return slices.Clone(m.admins), nil }

func (m *mailFake) send(_ context.Context, to, subject, body string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	err := m.h.Update(ctx, func(context.Context, *store.Tx) error { return nil })
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sent = append(m.sent, sentMail{to: to, subject: subject, body: body, waited: err != nil})
	return nil
}

func (m *mailFake) mails() []sentMail {
	m.mu.Lock()
	defer m.mu.Unlock()
	return slices.Clone(m.sent)
}

// intakeFixture is an Intake over a store holding one live admin and a
// fresh staging directory, its clock the fixture's, its free-space read
// answering plentyFree, and its syncs succeeding without the disk.
type intakeFixture struct {
	h       *store.Handle
	in      *Intake
	s       *submissions
	staging string
	clock   *testClock
	mail    *mailFake
}

func newIntakeFixture(t *testing.T) *intakeFixture {
	t.Helper()
	h := openTestStore(t)
	insertUser(t, h, userSpec{admin: true, createdAt: 1})
	return reopenIntake(t, h, t.TempDir(), &testClock{now: time.UnixMicro(testNow)})
}

// reopenIntake builds a second Intake over h and staging, as a restart
// does: a new process over the same store and directory.
func reopenIntake(t *testing.T, h *store.Handle, staging string, clock *testClock) *intakeFixture {
	t.Helper()
	mail := &mailFake{h: h, admins: []string{"ann@example.invalid", "bo@example.invalid"}}
	in, err := New(Config{
		Store: h, Log: slog.New(slog.DiscardHandler), Staging: staging, Now: clock.Now,
		Limits:        func(context.Context) (Limits, error) { return testSeamLimits(), nil },
		Windows:       func(context.Context) (Windows, error) { return testWindows, nil },
		SweepInterval: time.Hour, CheckWorkers: 1,
		Admins: mail.addresses, Send: mail.send,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	in.subs.freeBytes = func(string) (uint64, error) { return plentyFree, nil }
	in.subs.sync = func(string) error { return nil }
	t.Cleanup(func() { _ = in.Close() })
	return &intakeFixture{h: h, in: in, s: in.subs, staging: staging, clock: clock, mail: mail}
}

// restart closes f's Intake and builds another over the same store,
// staging, and clock.
func (f *intakeFixture) restart(t *testing.T) *intakeFixture {
	t.Helper()
	if err := f.in.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	return reopenIntake(t, f.h, f.staging, f.clock)
}

// path is name under staging, slash-separated as the layout writes it.
func (f *intakeFixture) path(name string) string {
	return filepath.Join(f.staging, filepath.FromSlash(name))
}

func tusOf(id []byte) string { return "tus/" + hex.EncodeToString(id) }

func xOf(submission, member []byte) string {
	return "submissions/" + hex.EncodeToString(submission) + "/x/" + hex.EncodeToString(member)
}

// write puts data at name under staging, creating its directory.
func (f *intakeFixture) write(t *testing.T, name string, data []byte) {
	t.Helper()
	p := f.path(name)
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, data, 0o640); err != nil {
		t.Fatal(err)
	}
}

// read is the bytes at name under staging, or nil when nothing is there.
func (f *intakeFixture) read(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(f.path(name))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func (f *intakeFixture) exists(t *testing.T, name string) bool {
	t.Helper()
	_, err := os.Lstat(f.path(name))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	return err == nil
}

// stageUpload writes data as file's tus upload: its data file and the
// .info tusd keeps beside it.
func (f *intakeFixture) stageUpload(t *testing.T, file []byte, data []byte) {
	t.Helper()
	f.write(t, tusOf(file), data)
	f.write(t, tusOf(file)+".info", fmt.Appendf(nil, `{"ID":%q,"Size":%d,"MetaData":{}}`, hex.EncodeToString(file), len(data)))
}

// albumOf is a loose-files album manifest of one file per data, named
// 01.flac, 02.flac, and so on.
func albumOf(data ...[]byte) newSubmission {
	n := newSubmission{manifest: gate.Manifest{Unit: gate.UnitAlbum, Container: gate.ContainerFiles, Source: "bought"}}
	for i, d := range data {
		n.manifest.Files = append(n.manifest.Files, gate.ManifestFile{Path: fmt.Sprintf("%02d.flac", i+1), Size: int64(len(d))})
	}
	return n
}

// tracksOf is albumOf's separate-tracks twin.
func tracksOf(data ...[]byte) newSubmission {
	n := albumOf(data...)
	n.manifest.Unit = gate.UnitTrack
	return n
}

// zipped is a one-archive submission of unit holding archive.
func zipped(unit gate.Unit, archive []byte) newSubmission {
	return newSubmission{manifest: gate.Manifest{Unit: unit, Container: gate.ContainerZip, Source: "cd",
		Files: []gate.ManifestFile{{Path: "music.zip", Size: int64(len(archive))}}}}
}

// checkingWith creates n for uploader, stages data as its uploads in
// manifest order, and receives each, so the submission is checking.
func (f *intakeFixture) checkingWith(t *testing.T, uploader identity.UserID, n newSubmission, data ...[]byte) created {
	t.Helper()
	c := mustCreate(t, f.s, uploader, n, testLimits())
	for i, file := range c.files {
		f.stageUpload(t, file.id, data[i])
	}
	receiveAll(t, f.s, c)
	if got := stateOf(t, f.h, c.id); got != "checking" {
		t.Fatalf("fixture state %q, want checking", got)
	}
	return c
}

// twoTrackAlbum is a passing two-track album's file bytes.
func twoTrackAlbum(t *testing.T) [][]byte {
	t.Helper()
	return [][]byte{cdTrack(1, 1).encode(t), cdTrack(2, 2).encode(t)}
}

// twoTrackZIP is the same album as one archive.
func twoTrackZIP(t *testing.T) []byte {
	t.Helper()
	tracks := twoTrackAlbum(t)
	return zipOfMembers(t, zipMember{"01.flac", tracks[0]}, zipMember{"02.flac", tracks[1]})
}

// await polls until submission id leaves every state in passing, and
// returns the state it reached.
func (f *intakeFixture) await(t *testing.T, id []byte, passing ...string) string {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		st := stateOf(t, f.h, id)
		if !slices.Contains(passing, st) {
			return st
		}
		if time.Now().After(deadline) {
			t.Fatalf("submission still %q", st)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// reservedBytes is a submission's reservation.
func reservedBytes(t *testing.T, h *store.Handle, id []byte) int64 {
	t.Helper()
	var n int64
	if err := h.QueryRow(deadline(t), `SELECT reserved_bytes FROM submissions WHERE id = ?`, id).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// albumKeyOf is a submission's album key, nil when unset.
func albumKeyOf(t *testing.T, h *store.Handle, id []byte) []byte {
	t.Helper()
	var key []byte
	if err := h.QueryRow(deadline(t), `SELECT album_key FROM submissions WHERE id = ?`, id).Scan(&key); err != nil {
		t.Fatal(err)
	}
	return key
}

// members are an archive's extracted rows, keyed by member path, each
// with its verdict ("" for a pass).
func membersOf(t *testing.T, h *store.Handle, id []byte) map[string][]byte {
	t.Helper()
	rows, err := h.Query(deadline(t), `SELECT rel_path, id FROM submission_files WHERE submission_id = ? AND role != 'archive'`, id)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = rows.Close() }()
	out := map[string][]byte{}
	for rows.Next() {
		var path, fid []byte
		if err := rows.Scan(&path, &fid); err != nil {
			t.Fatal(err)
		}
		out[string(path)] = fid
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// lastEvent is a submission's latest event kind and reason.
func lastEvent(t *testing.T, h *store.Handle, id []byte) (kind, reason string) {
	t.Helper()
	var r *string
	err := h.QueryRow(deadline(t), `SELECT kind, reason FROM submission_events WHERE submission_id = ? ORDER BY seq DESC LIMIT 1`, id).Scan(&kind, &r)
	if err != nil {
		t.Fatal(err)
	}
	if r != nil {
		reason = *r
	}
	return kind, reason
}

// stopAfter is a context whose n-th counted Err call, and every later
// call, reports it cancelled, closing Done at that call: a stop at the
// n-th point the code under test looks. database/sql reads Err from a
// goroutine of its own as each result set closes, at no fixed point in
// the code under test's order, so its calls are not counted.
type stopAfter struct {
	context.Context
	mu    sync.Mutex
	n     int
	calls int
	done  chan struct{}
}

func newStopAfter(n int) *stopAfter {
	return &stopAfter{Context: context.Background(), n: n, done: make(chan struct{})}
}

func (c *stopAfter) Done() <-chan struct{} { return c.done }

func (c *stopAfter) Err() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	pc, _, _, _ := runtime.Caller(1)
	if !strings.HasPrefix(runtime.FuncForPC(pc).Name(), "database/sql.") {
		c.calls++
		if c.calls == c.n {
			close(c.done)
		}
	}
	if c.calls >= c.n {
		return context.Canceled
	}
	return nil
}

// points is how many counted Err calls the context has seen.
func (c *stopAfter) points() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.calls
}

// pauseWhen is a context whose first Err call at which when holds blocks
// until the test releases it, so the test can act while the code under
// test stands at that point. It never reports cancellation.
type pauseWhen struct {
	context.Context
	when    func() bool
	once    sync.Once
	paused  chan struct{}
	release chan struct{}
}

func newPauseWhen(when func() bool) *pauseWhen {
	return &pauseWhen{Context: context.Background(), when: when, paused: make(chan struct{}), release: make(chan struct{})}
}

func (c *pauseWhen) Err() error {
	if c.when() {
		c.once.Do(func() {
			close(c.paused)
			<-c.release
		})
	}
	return nil
}

// entries counts the names in dir under staging, zero when it is absent.
func (f *intakeFixture) entries(t *testing.T, dir string) int {
	t.Helper()
	list, err := os.ReadDir(f.path(dir))
	if errors.Is(err, fs.ErrNotExist) {
		return 0
	}
	if err != nil {
		t.Fatal(err)
	}
	return len(list)
}

// failedFiles maps each file of a submission that holds a verdict to
// its code.
func failedFiles(t *testing.T, h *store.Handle, id []byte) map[string]string {
	t.Helper()
	got := fileVerdicts(t, h, id)
	maps.DeleteFunc(got, func(_, code string) bool { return code == "" })
	return got
}

// memberUseOf is what counts against uploader's quota now.
func memberUseOf(t *testing.T, h *store.Handle, uploader identity.UserID) (int64, error) {
	t.Helper()
	var use int64
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		var err error
		use, err = memberUse(ctx, tx, uploader, nil)
		return err
	})
	return use, err
}

// eventKinds joins a submission's event kinds for a failure message.
func eventKinds(t *testing.T, h *store.Handle, id []byte) string {
	return strings.Join(eventsOf(t, h, id), ",")
}
