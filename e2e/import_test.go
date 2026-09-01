package e2e

import (
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"mime"
	"mime/quotedprintable"
	"net/http"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/importer"
	"github.com/glw907/dubplate/internal/lock"
)

// --- fixture helpers -----------------------------------------------------

// writeBadAlbum stages an album the quality gate must reject: a
// valid FLAC beside a file carrying an .mp3 extension.
func (e *testEnv) writeBadAlbum(user, album string) {
	e.t.Helper()
	dir := filepath.Join(e.inboxDir, user, album)
	writeGoodFlac(e.t, filepath.Join(dir, "track1.flac"))
	if err := os.WriteFile(filepath.Join(dir, "track2.mp3"), []byte("not-really-audio"), 0o644); err != nil {
		e.t.Fatalf("write bad track: %v", err)
	}
	e.ageInbox(user)
}

// readReport reads and parses the run report the last [testEnv.run] call
// wrote: since U8, one report-<runid>.json per run rather than one
// shared, overwritten file, so this finds it by glob -- every e2e test
// using it calls "import" exactly once per testEnv, so exactly one match
// is the invariant being checked here too.
func (e *testEnv) readReport() importer.RunReport {
	e.t.Helper()
	matches, err := filepath.Glob(filepath.Join(e.stateDir, "report-*.json"))
	if err != nil {
		e.t.Fatalf("glob run reports: %v", err)
	}
	if len(matches) != 1 {
		e.t.Fatalf("run report files = %d, want exactly 1: %v", len(matches), matches)
	}
	data, err := os.ReadFile(matches[0])
	if err != nil {
		e.t.Fatalf("read run report: %v", err)
	}
	var report importer.RunReport
	if err := json.Unmarshal(data, &report); err != nil {
		e.t.Fatalf("unmarshal run report %s: %v", data, err)
	}
	return report
}

// findFile searches root recursively for a file named name and returns
// its path, or "" if none matched.
func findFile(root, name string) string {
	var found string
	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || found != "" {
			return nil //nolint:nilerr // best-effort scan; a walk error just ends the search
		}
		if !d.IsDir() && d.Name() == name {
			found = path
		}
		return nil
	})
	return found
}

// parseDelivery decodes one delivered message's subject (undoing RFC
// 2047 encoding when present) and quoted-printable body.
func parseDelivery(t *testing.T, d smtpDelivery) (subject, body string) {
	t.Helper()
	msg, err := mail.ReadMessage(bytes.NewReader(d.data))
	if err != nil {
		t.Fatalf("parse delivered message: %v", err)
	}
	subject, err = (&mime.WordDecoder{}).DecodeHeader(msg.Header.Get("Subject"))
	if err != nil {
		t.Fatalf("decode subject %q: %v", msg.Header.Get("Subject"), err)
	}
	raw, err := io.ReadAll(quotedprintable.NewReader(msg.Body))
	if err != nil {
		t.Fatalf("decode body: %v", err)
	}
	return subject, string(raw)
}

func findDeliveryTo(deliveries []smtpDelivery, addrContains string) (smtpDelivery, bool) {
	for _, d := range deliveries {
		for _, to := range d.to {
			if strings.Contains(to, addrContains) {
				return d, true
			}
		}
	}
	return smtpDelivery{}, false
}

func findDeliveryBySubject(t *testing.T, deliveries []smtpDelivery, subjectContains string) smtpDelivery {
	t.Helper()
	for _, d := range deliveries {
		subject, _ := parseDelivery(t, d)
		if strings.Contains(subject, subjectContains) {
			return d
		}
	}
	t.Fatalf("no delivered message with subject containing %q among %d deliveries", subjectContains, len(deliveries))
	return smtpDelivery{}
}

// --- happy path ------------------------------------------------------------

func TestImportHappyPath(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Good Album")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	if _, err := os.Stat(filepath.Join(e.stagingDir, "alice")); !os.IsNotExist(err) {
		t.Errorf("alice's staging directory survived a clean import: err = %v", err)
	}

	report := e.readReport()
	if report.LockHeld {
		t.Error("report.LockHeld = true, want false")
	}
	if report.Error != "" {
		t.Errorf("report.Error = %q, want empty on a clean run", report.Error)
	}
	if len(report.Albums) != 1 {
		t.Fatalf("report.Albums = %+v, want exactly 1", report.Albums)
	}
	album := report.Albums[0]
	if album.Uploader != "alice" {
		t.Errorf("album.Uploader = %q, want alice", album.Uploader)
	}
	if album.Outcome != importer.OutcomeImported {
		t.Errorf("album.Outcome = %q, want %q", album.Outcome, importer.OutcomeImported)
	}
	if len(album.LibraryPaths) != 1 {
		t.Errorf("album.LibraryPaths = %v, want exactly one resolved path", album.LibraryPaths)
	}

	if mail := e.drainMail(); len(mail) != 0 {
		t.Errorf("mail sent on a clean import = %d, want 0", len(mail))
	}

	select {
	case hit := <-e.pingHits:
		if hit.method != http.MethodGet {
			t.Errorf("ping method = %s, want GET", hit.method)
		}
	default:
		t.Error("no success ping received")
	}

	select {
	case <-e.navidromeHits:
	default:
		t.Error("no navidrome scan request received")
	}
}

// --- beets skip path, a known contributor -----------------------------------

func TestImportSkipPath_KnownUploader(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Weird Album")
	e.env["STUB_BEET_SKIP"] = filepath.Join(e.stagingDir, "alice", "Weird Album", "track1.flac")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	if findFile(e.reviewDir, "track1.flac") == "" {
		t.Fatal("skipped track never landed in review")
	}

	report := e.readReport()
	if len(report.Albums) != 1 || report.Albums[0].Outcome != importer.OutcomeReview {
		t.Fatalf("report.Albums = %+v, want one review outcome", report.Albums)
	}

	mail := e.drainMail()
	if len(mail) != 2 {
		t.Fatalf("mail sent = %d, want 2 (uploader + admin)", len(mail))
	}
	uploaderMail, ok := findDeliveryTo(mail, "alice@example.com")
	if !ok {
		t.Fatal("alice was not emailed")
	}
	subject, body := parseDelivery(t, uploaderMail)
	if subject != "Dubplate: some uploads need review" {
		t.Errorf("uploader subject = %q", subject)
	}
	if !strings.Contains(body, "could not be auto-tagged") {
		t.Errorf("uploader body = %q, want it to explain the skip", body)
	}

	adminMail := findDeliveryBySubject(t, mail, "import skipped items")
	_, adminBody := parseDelivery(t, adminMail)
	if !strings.Contains(adminBody, "alice") {
		t.Errorf("admin summary = %q, want it to name alice", adminBody)
	}
}

// --- quality gate failure, a known contributor ------------------------------

func TestImportQualityGateFailure_KnownUploader(t *testing.T) {
	e := newTestEnv(t)
	e.writeBadAlbum("alice", "Bad Album")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	if findFile(e.reviewDir, "track2.mp3") == "" {
		t.Fatal("bad album never landed in review")
	}

	report := e.readReport()
	if len(report.Albums) != 1 || report.Albums[0].Outcome != importer.OutcomeRejected {
		t.Fatalf("report.Albums = %+v, want one rejected outcome", report.Albums)
	}

	mail := e.drainMail()
	if len(mail) != 2 {
		t.Fatalf("mail sent = %d, want 2 (uploader + admin)", len(mail))
	}
	uploaderMail, ok := findDeliveryTo(mail, "alice@example.com")
	if !ok {
		t.Fatal("alice was not emailed")
	}
	subject, body := parseDelivery(t, uploaderMail)
	if subject != "Your music upload needs a different format" {
		t.Errorf("uploader subject = %q", subject)
	}
	if !strings.Contains(body, "Hi alice,") {
		t.Errorf("uploader body missing greeting: %q", body)
	}

	adminMail := findDeliveryBySubject(t, mail, "import quality gate failures")
	_, adminBody := parseDelivery(t, adminMail)
	if !strings.Contains(adminBody, "track2.mp3") {
		t.Errorf("admin summary = %q, want the failing file named", adminBody)
	}
}

// --- M10: unknown uploader (absent from contributors.yaml) ------------------
//
// "carol" never appears in newTestEnv's contributors.yaml, so her email
// and first name are both absent. These four cases (ported from the
// shell layer's fix wave, tests/music-import.bats) prove the unmatched
// branch changes only the notification -- never where the file lands or
// whether the run succeeds -- and never bleeds into a sibling upload
// from a real contributor in the same run.

func TestImportSkipPath_UnmatchedUploaderOnly(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("carol", "Weird Album")
	e.env["STUB_BEET_SKIP"] = filepath.Join(e.stagingDir, "carol", "Weird Album", "track1.flac")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	if findFile(e.reviewDir, "track1.flac") == "" {
		t.Fatal("carol's skipped track never landed in review")
	}

	mail := e.drainMail()
	if len(mail) != 1 {
		t.Fatalf("mail sent = %d, want exactly 1 (admin only)", len(mail))
	}
	if _, ok := findDeliveryTo(mail, "carol"); ok {
		t.Error("carol was emailed despite being absent from contributors.yaml")
	}
	_, adminBody := parseDelivery(t, mail[0])
	if !strings.Contains(adminBody, "carol") {
		t.Errorf("admin summary = %q, want it to name carol", adminBody)
	}
}

func TestImportSkipPath_KnownAndUnmatchedTogether(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("carol", "Weird Album C")
	e.writeUploadedFlac("alice", "Weird Album A")
	skipC := filepath.Join(e.stagingDir, "carol", "Weird Album C", "track1.flac")
	skipA := filepath.Join(e.stagingDir, "alice", "Weird Album A", "track1.flac")
	e.env["STUB_BEET_SKIP"] = skipC + "\n" + skipA

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	mail := e.drainMail()
	// Alice's own notice plus the admin aggregate: carol never gets a
	// personal message.
	if len(mail) != 2 {
		t.Fatalf("mail sent = %d, want exactly 2", len(mail))
	}
	if _, ok := findDeliveryTo(mail, "carol"); ok {
		t.Error("carol was emailed despite being absent from contributors.yaml")
	}
	if _, ok := findDeliveryTo(mail, "alice@example.com"); !ok {
		t.Error("alice was not emailed her own notice")
	}

	adminMail := findDeliveryBySubject(t, mail, "import skipped items")
	_, adminBody := parseDelivery(t, adminMail)
	if !strings.Contains(adminBody, "alice") || !strings.Contains(adminBody, "carol") {
		t.Errorf("admin summary = %q, want both alice and carol named", adminBody)
	}
}

func TestImportQualityGateFailure_UnmatchedUploaderOnly(t *testing.T) {
	e := newTestEnv(t)
	e.writeBadAlbum("carol", "Bad Album")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	if findFile(e.reviewDir, "track2.mp3") == "" {
		t.Fatal("carol's bad album never landed in review")
	}

	mail := e.drainMail()
	if len(mail) != 1 {
		t.Fatalf("mail sent = %d, want exactly 1 (admin only)", len(mail))
	}
	if _, ok := findDeliveryTo(mail, "carol"); ok {
		t.Error("carol was emailed despite being absent from contributors.yaml")
	}
	_, adminBody := parseDelivery(t, mail[0])
	if !strings.Contains(adminBody, "carol") || !strings.Contains(adminBody, "track2.mp3") {
		t.Errorf("admin summary = %q, want carol and the failing file named", adminBody)
	}
}

func TestImportQualityGateFailure_KnownAndUnmatchedTogether(t *testing.T) {
	e := newTestEnv(t)
	e.writeBadAlbum("alice", "Bad Album A")
	e.writeBadAlbum("carol", "Bad Album C")

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	mail := e.drainMail()
	if len(mail) != 2 {
		t.Fatalf("mail sent = %d, want exactly 2", len(mail))
	}
	if _, ok := findDeliveryTo(mail, "carol"); ok {
		t.Error("carol was emailed despite being absent from contributors.yaml")
	}
	aliceMail, ok := findDeliveryTo(mail, "alice@example.com")
	if !ok {
		t.Fatal("alice was not emailed")
	}
	if subject, _ := parseDelivery(t, aliceMail); subject != "Your music upload needs a different format" {
		t.Errorf("alice's subject = %q", subject)
	}

	adminMail := findDeliveryBySubject(t, mail, "import quality gate failures")
	_, adminBody := parseDelivery(t, adminMail)
	if !strings.Contains(adminBody, "carol/Bad Album C") {
		t.Errorf("admin summary = %q, want carol/Bad Album C named", adminBody)
	}
}

// --- lock contention, free-space abort, and collection routing -------------

func TestImportLockContention(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Good Album")

	if err := os.MkdirAll(e.stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	held, err := lock.Acquire(lock.Path(e.stateDir))
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	defer func() { _ = held.Release() }()

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0 on a held lock: %s", code, output)
	}

	// Nothing was touched: the album is exactly where it was uploaded.
	if _, err := os.Stat(filepath.Join(e.inboxDir, "alice", "Good Album", "track1.flac")); err != nil {
		t.Errorf("uploaded album disturbed despite a held lock: %v", err)
	}

	report := e.readReport()
	if !report.LockHeld {
		t.Error("report.LockHeld = false, want true")
	}
	if len(report.Albums) != 0 {
		t.Errorf("report.Albums = %v, want empty on a held-lock run", report.Albums)
	}

	select {
	case hit := <-e.pingHits:
		if hit.method != http.MethodGet {
			t.Errorf("ping method = %s, want GET", hit.method)
		}
	default:
		t.Error("no success ping sent on a held lock")
	}
}

func TestImportFreeSpaceAbort(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Good Album")
	e.env["DUBPLATE_FREE_FLOOR_GB"] = "1000000000"

	output, code := e.run("import")
	if code == 0 {
		t.Fatalf("import exited 0, want nonzero on a free-space abort: %s", output)
	}

	if _, err := os.Stat(filepath.Join(e.inboxDir, "alice", "Good Album", "track1.flac")); err != nil {
		t.Errorf("uploaded album disturbed despite a free-space abort: %v", err)
	}

	report := e.readReport()
	if report.Error == "" {
		t.Error("report.Error is empty, want the free-space failure recorded")
	}
	if len(report.Albums) != 0 {
		t.Errorf("report.Albums = %v, want empty on an aborted run", report.Albums)
	}

	select {
	case hit := <-e.pingHits:
		t.Errorf("in-process ping sent on a per-run failure: %+v (OnFailure= owns this, not the binary)", hit)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestImportCollectionRouting(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Good Album")
	e.env["DUBPLATE_COLLECTIONS"] = "archive"
	argsLog := filepath.Join(e.root, "beet-args.log")
	e.env["STUB_BEET_ARGS_LOG"] = argsLog

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0: %s", code, output)
	}

	data, err := os.ReadFile(argsLog)
	if err != nil {
		t.Fatalf("read beet args log: %v", err)
	}
	if !strings.Contains(string(data), "--set\ncollection=archive") {
		t.Errorf("beet args = %q, want --set collection=archive", data)
	}

	report := e.readReport()
	if len(report.Albums) != 1 || report.Albums[0].Collection != "archive" {
		t.Errorf("report.Albums = %+v, want collection %q recorded", report.Albums, "archive")
	}
}
