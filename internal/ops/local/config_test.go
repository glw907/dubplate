package local

import (
	"context"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// TestJobRetention holds the reader runBackup builds for the job
// runner: with no stored row it returns the registry default, and once
// an env-seed row exists it returns that row's value, which sits above
// the 24h floor.
func TestJobRetention(t *testing.T) {
	_, h := newStore(t)

	reg := config.NewRegistry()
	Register(reg)
	retention := JobRetention(settings.NewAccessor(h, reg))

	got, err := retention(t.Context())
	if err != nil || got != 2160*time.Hour {
		t.Fatalf("DUBPLATE_JOB_RETENTION with no row = %v, %v, want the default 2160h0m0s", got, err)
	}

	seedSettingRow(t, h, "DUBPLATE_JOB_RETENTION", "48h")
	got, err = retention(t.Context())
	if err != nil || got != 48*time.Hour {
		t.Errorf("DUBPLATE_JOB_RETENTION after an env-seed row = %v, %v, want 48h0m0s", got, err)
	}
}

// seedSettingRow writes key as a literal env-seed settings row and
// bumps the generation, in one transaction: production writes a
// settings row only through the seed apply, so a test seeds one
// directly by literal SQL.
func seedSettingRow(t *testing.T, h *store.Handle, key, value string) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES (?, ?, 'env-seed', ?, 'fixture', 1)`, key, value, value)
		if err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `UPDATE settings_generation SET generation = generation + 1 WHERE singleton = 1`)
		return err
	})
	if err != nil {
		t.Fatalf("seed setting %s: %v", key, err)
	}
}
