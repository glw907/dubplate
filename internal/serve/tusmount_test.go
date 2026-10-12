package serve

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake"
)

// seedUpload writes one member with an uploading submission of one
// file of size bytes, and returns the member, the submission id, and
// the file id.
func seedUpload(t *testing.T, h *store.Handle, size int) (identity.UserID, []byte, []byte) {
	t.Helper()
	member, coll, submission, file := store.NewID(), store.NewID(), store.NewID(), store.NewID()
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		for _, stmt := range []struct {
			query string
			args  []any
		}{
			{`INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at)
				VALUES (?, 'member@example.invalid', 'Member', 'member', 1, 1, 1)`, []any{member}},
			{`INSERT INTO collections (id, name, root_path, owner, visibility, writer)
				VALUES (?, 'Library', ?, ?, 'group', 'server')`, []any{coll, []byte(hex.EncodeToString(coll)), member}},
			{`INSERT INTO submissions (id, uploader, collection_id, verdict_unit, container, state, source,
				declared_bytes, reserved_bytes, created_at, updated_at)
				VALUES (?, ?, ?, 'album', 'files', 'uploading', 'cd', ?, ?, 1, 1)`, []any{submission, member, coll, size, size}},
			{`INSERT INTO submission_files (id, submission_id, rel_path, size, role)
				VALUES (?, ?, ?, ?, 'audio')`, []any{file, submission, []byte("01.flac"), size}},
		} {
			if _, err := tx.Exec(ctx, stmt.query, stmt.args...); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed upload: %v", err)
	}
	return identity.UserID(member), submission, file
}

// TestTusPatchOutlivesTheServerTimeouts is AC 9: a PATCH whose body
// trickles in over longer than the server's read and write timeouts,
// in reads spaced well under tusd's NetworkTimeout, completes through
// the production server constructor and the shell's whole middleware.
// tusd extends both deadlines after each read through
// http.ResponseController, which reaches the connection only while
// every writer between it and net/http unwraps.
//
// Both timeouts are lowered to a fifth of the time the body takes, so
// scheduling delay cannot carry the trickle under either.
func TestTusPatchOutlivesTheServerTimeouts(t *testing.T) {
	const (
		lowered = 200 * time.Millisecond
		spacing = 100 * time.Millisecond
		size    = 10
	)
	h, err := store.OpenServe(t.Context(), t.TempDir(), store.ServeOptions{
		BusyTimeout: 5 * time.Second, MigrationWait: 30 * time.Second, Keep: 1, Logger: slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	member, submission, file := seedUpload(t, h, size)

	staging := t.TempDir()
	logger := slog.New(slog.DiscardHandler)
	mount, err := intake.NewTusMount(h, staging, logger, intake.TusConfig{
		Prefix:    testPrefix,
		PublicURL: "https://music.example",
		Verify:    func(http.ResponseWriter, *http.Request) (identity.UserID, error) { return member, nil },
		Floor:     func(context.Context) (int64, error) { return 0, nil },
		TusLog:    io.Discard,
	})
	if err != nil {
		t.Fatal(err)
	}
	shell := newTestShell(t)
	shell.Register("tus", mount)

	srv := httptest.NewUnstartedServer(nil)
	srv.Config = newHTTPServer(shell.Handler(), logger, lowered)
	srv.Config.ReadTimeout = lowered
	srv.Start()
	t.Cleanup(srv.Close)

	create, err := http.NewRequestWithContext(t.Context(), http.MethodPost, srv.URL+testPrefix+"/tus/", http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	create.Header.Set("Sec-Fetch-Site", "same-origin")
	create.Header.Set("Tus-Resumable", "1.0.0")
	create.Header.Set("Upload-Length", strconv.Itoa(size))
	create.Header.Set("Upload-Metadata", "submission "+base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(submission)))+
		",file "+base64.StdEncoding.EncodeToString([]byte(hex.EncodeToString(file))))
	resp, err := srv.Client().Do(create)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("create: status %d, want 201", resp.StatusCode)
	}

	body, trickle := io.Pipe()
	go func() {
		for range size {
			time.Sleep(spacing)
			if _, err := trickle.Write([]byte("x")); err != nil {
				return
			}
		}
		_ = trickle.Close()
	}()
	patch, err := http.NewRequestWithContext(t.Context(), http.MethodPatch, srv.URL+testPrefix+"/tus/"+hex.EncodeToString(file), body)
	if err != nil {
		t.Fatal(err)
	}
	patch.ContentLength = size
	patch.Header.Set("Sec-Fetch-Site", "same-origin")
	patch.Header.Set("Tus-Resumable", "1.0.0")
	patch.Header.Set("Upload-Offset", "0")
	patch.Header.Set("Content-Type", "application/offset+octet-stream")

	start := time.Now()
	resp, err = srv.Client().Do(patch)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("PATCH failed after %s: %v", elapsed, err)
	}
	_ = resp.Body.Close()
	if elapsed < 2*lowered {
		t.Fatalf("the trickle took %s, not past the lowered timeouts; the test proves nothing", elapsed)
	}
	if resp.StatusCode != http.StatusNoContent || resp.Header.Get("Upload-Offset") != strconv.Itoa(size) {
		t.Fatalf("PATCH after %s: status %d offset %q, want 204 and %d", elapsed, resp.StatusCode, resp.Header.Get("Upload-Offset"), size)
	}
	got, err := os.ReadFile(filepath.Join(staging, "tus", hex.EncodeToString(file)))
	if err != nil || len(got) != size {
		t.Fatalf("upload holds %d bytes (%v), want %d", len(got), err, size)
	}
}
