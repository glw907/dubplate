package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/store"
)

// buildZeroByteStoreLeaf creates a leaf under stateDir shaped as
// internal/store's own leaf (store.lock, db/dubplate.db at
// zero bytes, snapshots/, staging/), so the inspect mode classifies it
// as an interrupted create rather than a real store.
// cmd/dubplate-server has no caller into store's own unexported leaf
// builder, so this mirrors its four names directly; a change to that
// shape is store's own package, not this one, to notice.
func buildZeroByteStoreLeaf(t *testing.T, stateDir string) {
	t.Helper()
	root := filepath.Join(stateDir, "store")
	for _, dir := range []string{root, filepath.Join(root, "db"), filepath.Join(root, "snapshots"), filepath.Join(root, "staging")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "store.lock"), nil, 0o600); err != nil {
		t.Fatalf("write store.lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "db", "dubplate.db"), nil, 0o600); err != nil {
		t.Fatalf("write dubplate.db: %v", err)
	}
}

// seedSettingRow writes key as a literal settings row of source
// ("env_seed" or "admin") with a matching seed value, and bumps the
// generation, in one transaction: production writes a settings row
// only through the seed apply or, for an admin row, rung 14's write
// path, so a test seeds one directly by literal SQL.
func seedSettingRow(t *testing.T, h *store.Handle, key, value, source string) {
	t.Helper()
	if source != "env_seed" && source != "admin" {
		t.Fatalf("seed setting %s: source %q, want env_seed or admin", key, source)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		q := fmt.Sprintf(`INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES (?, ?, '%s', ?, 'fixture', 1)`, source)
		if _, err := tx.Exec(ctx, q, key, value, value); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE settings_generation SET generation = generation + 1 WHERE singleton = 1`)
		return err
	})
	if err != nil {
		t.Fatalf("seed setting %s: %v", key, err)
	}
}
