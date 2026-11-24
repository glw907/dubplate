package backup_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// TestSettingsReaders is spec criterion 13's reader half: each of the
// three backup keys' reader returns the registry default while the
// store holds no row, and a stored env_seed row's non-default value
// once one exists. Each stored value sits at or above the floor under
// either answer to Ruling 3.
func TestSettingsReaders(t *testing.T) {
	h := openWrite(t, newStateDir(t))

	reg := config.NewRegistry()
	store.Register(reg)
	backup.Register(reg)
	guardFraction, snapshotKeep, archiveRetention := backup.Readers(settings.NewAccessor(h, reg))

	tests := []struct {
		name        string
		key         string
		seed        string
		read        func() (string, error)
		wantDefault string
		wantSeeded  string
	}{
		{
			name: "backup guard fraction", key: "DUBPLATE_BACKUP_GUARD_FRACTION", seed: "0.9",
			read:        func() (string, error) { v, err := guardFraction(context.Background()); return fmt.Sprint(v), err },
			wantDefault: "0.5", wantSeeded: "0.9",
		},
		{
			name: "backup snapshot keep", key: "DUBPLATE_BACKUP_SNAPSHOT_KEEP", seed: "3",
			read:        func() (string, error) { v, err := snapshotKeep(context.Background()); return fmt.Sprint(v), err },
			wantDefault: "14", wantSeeded: "3",
		},
		{
			name: "backup archive retention", key: "DUBPLATE_BACKUP_ARCHIVE_RETENTION", seed: "168h",
			read:        func() (string, error) { v, err := archiveRetention(context.Background()); return v.String(), err },
			wantDefault: "2160h0m0s", wantSeeded: "168h0m0s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.read(); err != nil || got != tt.wantDefault {
				t.Fatalf("%s with no row = %q, %v, want the default %q", tt.key, got, err, tt.wantDefault)
			}
			seedSettingRow(t, h, tt.key, tt.seed)
			if got, err := tt.read(); err != nil || got != tt.wantSeeded {
				t.Errorf("%s after an env_seed row = %q, %v, want %q", tt.key, got, err, tt.wantSeeded)
			}
		})
	}
}

// seedSettingRow writes key as a literal env_seed settings row and
// bumps the generation, in one transaction: production writes a
// settings row only through the seed apply, so a test seeds one
// directly by literal SQL.
func seedSettingRow(t *testing.T, h *store.Handle, key, value string) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES (?, ?, 'env_seed', ?, 'fixture', 1)`, key, value, value)
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
