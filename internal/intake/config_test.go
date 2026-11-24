package intake

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// TestIntakeAllowlist: the allowlist splits into format names, and a
// name the gate's format table does not hold, an empty entry among
// them, is refused naming it.
func TestIntakeAllowlist(t *testing.T) {
	const key = "DUBPLATE_INTAKE_FORMAT_ALLOWLIST"
	tests := []struct {
		value string
		want  []string
		// refusal names the entry a refusal must quote.
		refusal string
	}{
		{value: "flac", want: []string{"flac"}},
		{value: " flac ", want: []string{"flac"}},
		{value: "flac,mp3", refusal: `"mp3"`},
		{value: "flac,", refusal: `""`},
		{value: ",", refusal: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			got, err := Allowlist([]string{key + "=" + tt.value})
			if tt.refusal == "" {
				if err != nil || !slices.Equal(got, tt.want) {
					t.Errorf("Allowlist = %q, %v; want %q", got, err, tt.want)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), key) || !strings.Contains(err.Error(), tt.refusal) {
				t.Errorf("Allowlist error = %v, want a refusal naming %s and %s", err, key, tt.refusal)
			}
		})
	}
}

// TestValidateStaging_ReadsTheWallValue proves the overlap rule reads the
// staging row's value: Load accepts a staging directory inside the
// library root, and ValidateStaging refuses it keyed by the row for
// doctor.
func TestValidateStaging_ReadsTheWallValue(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)
	cfg, err := config.Load([]string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STAGING_DIR=/srv/music/library/staging",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load: %v", err)
	}
	err = ValidateStaging(cfg)
	if err == nil {
		t.Fatal("ValidateStaging() error = nil, want the nesting refused: Load alone accepts it")
	}
	if !strings.HasPrefix(err.Error(), "DUBPLATE_STAGING_DIR: ") {
		t.Errorf("ValidateStaging() error = %q, want it keyed by DUBPLATE_STAGING_DIR for doctor's row", err)
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

// newRegistry returns a registry holding intake's settings, which the
// accessor reads defaults and bounds from.
func newRegistry() *config.Registry {
	reg := config.NewRegistry()
	Register(reg)
	return reg
}

// TestReadersReadStoredValues proves each reader returns the stored
// non-default value of its own key, so a reader wired to the wrong
// setting shows.
func TestReadersReadStoredValues(t *testing.T) {
	h := openTestStore(t)
	acc := settings.NewAccessor(h, newRegistry())
	for key, value := range map[string]string{
		"DUBPLATE_INTAKE_MAX_ALBUM_BYTES":     "200000000",
		"DUBPLATE_INTAKE_MAX_FILES":           "11",
		"DUBPLATE_INTAKE_MAX_PATH_DEPTH":      "12",
		"DUBPLATE_INTAKE_MAX_TAG_BYTES":       "5000",
		"DUBPLATE_INTAKE_MAX_PICTURE_BYTES":   "70000",
		"DUBPLATE_INTAKE_MIN_SAMPLE_RATE":     "48000",
		"DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE": "24",
		"DUBPLATE_INTAKE_SESSIONS_PER_MEMBER": "4",
		"DUBPLATE_INTAKE_GROUP_QUOTA_BYTES":   "9000000000",
		"DUBPLATE_INTAKE_FREE_SPACE_FLOOR":    "2000000000",
		"DUBPLATE_INTAKE_RESUME_WINDOW":       "48h",
		"DUBPLATE_INTAKE_HOLD":                "72h",
		"DUBPLATE_INTAKE_STALL_AFTER":         "20m",
		"DUBPLATE_INTAKE_SWEEP_INTERVAL":      "2m",
		"DUBPLATE_INTAKE_CHECK_WORKERS":       "3",
	} {
		seedSettingRow(t, h, key, value)
	}

	lim, err := LimitsFrom(acc, []string{"flac"})(t.Context())
	if err != nil {
		t.Fatalf("LimitsFrom: %v", err)
	}
	if p := lim.Policy; p.MaxAlbumBytes != 200000000 || p.MaxFiles != 11 || p.MaxPathDepth != 12 ||
		p.MaxTagBytes != 5000 || p.MaxPictureBytes != 70000 || p.MinSampleRate != 48000 ||
		p.MinBitsPerSample != 24 || !slices.Equal(p.Allowlist, []string{"flac"}) {
		t.Errorf("LimitsFrom policy = %+v, want the seeded values and the allowlist", p)
	}
	if lim.SessionsPerMember != 4 || lim.GroupQuotaBytes != 9000000000 || lim.FreeSpaceFloor != 2000000000 {
		t.Errorf("LimitsFrom = {sessions %d, quota %d, floor %d}, want {4, 9000000000, 2000000000}",
			lim.SessionsPerMember, lim.GroupQuotaBytes, lim.FreeSpaceFloor)
	}

	w, err := WindowsFrom(acc)(t.Context())
	if err != nil {
		t.Fatalf("WindowsFrom: %v", err)
	}
	if w.ResumeWindow != 48*time.Hour || w.Hold != 72*time.Hour || w.StallAfter != 20*time.Minute {
		t.Errorf("WindowsFrom = %+v, want {48h, 72h, 20m}", w)
	}

	if floor, err := FreeSpaceFloor(acc)(t.Context()); err != nil || floor != 2000000000 {
		t.Errorf("FreeSpaceFloor = %d, %v, want 2000000000", floor, err)
	}
	if interval, err := SweepInterval(t.Context(), acc); err != nil || interval != 2*time.Minute {
		t.Errorf("SweepInterval = %v, %v, want 2m0s", interval, err)
	}
	if workers, err := CheckWorkerCount(t.Context(), acc); err != nil || workers != 3 {
		t.Errorf("CheckWorkerCount = %d, %v, want 3", workers, err)
	}
}

// TestReadersNameTheRefusedKey proves a stored row a reader cannot
// parse fails the read with the key first, and that the readers answer
// nil on success. The stored value is no number or duration, which no
// seed apply would write but a damaged store can hold.
func TestReadersNameTheRefusedKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
		read  func(acc *settings.Accessor) error
		want  string
	}{
		{"limits", "DUBPLATE_INTAKE_MAX_FILES", "many", func(acc *settings.Accessor) error {
			_, err := LimitsFrom(acc, nil)(t.Context())
			return err
		}, `DUBPLATE_INTAKE_MAX_FILES: stored env-seed value refused: DUBPLATE_INTAKE_MAX_FILES: invalid value "many", want an integer count from 1 to 2000, for example 500`},
		{"windows", "DUBPLATE_INTAKE_HOLD", "soon", func(acc *settings.Accessor) error {
			_, err := WindowsFrom(acc)(t.Context())
			return err
		}, `DUBPLATE_INTAKE_HOLD: stored env-seed value refused: DUBPLATE_INTAKE_HOLD: invalid value "soon", want a Go duration string from 24h to 720h, for example 168h`},
		{"free space floor", "DUBPLATE_INTAKE_FREE_SPACE_FLOOR", "lots", func(acc *settings.Accessor) error {
			_, err := FreeSpaceFloor(acc)(t.Context())
			return err
		}, `DUBPLATE_INTAKE_FREE_SPACE_FLOOR: stored env-seed value refused: DUBPLATE_INTAKE_FREE_SPACE_FLOOR: invalid value "lots", want an integer byte count from 1073741824 to 1099511627776, for example 5368709120`},
		{"sweep interval", "DUBPLATE_INTAKE_SWEEP_INTERVAL", "often", func(acc *settings.Accessor) error {
			_, err := SweepInterval(t.Context(), acc)
			return err
		}, `DUBPLATE_INTAKE_SWEEP_INTERVAL: stored env-seed value refused: DUBPLATE_INTAKE_SWEEP_INTERVAL: invalid value "often", want a Go duration string from 1m to 1h, for example 10m`},
		{"check workers", "DUBPLATE_INTAKE_CHECK_WORKERS", "several", func(acc *settings.Accessor) error {
			_, err := CheckWorkerCount(t.Context(), acc)
			return err
		}, `DUBPLATE_INTAKE_CHECK_WORKERS: stored env-seed value refused: DUBPLATE_INTAKE_CHECK_WORKERS: invalid value "several", want an integer count from 1 to 8, for example 1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := openTestStore(t)
			acc := settings.NewAccessor(h, newRegistry())
			if err := tt.read(acc); err != nil {
				t.Fatalf("read with no row: %v, want nil", err)
			}
			seedSettingRow(t, h, tt.key, tt.value)
			err := tt.read(acc)
			if err == nil || err.Error() != tt.want {
				t.Errorf("read error = %v, want %q", err, tt.want)
			}
		})
	}
}
