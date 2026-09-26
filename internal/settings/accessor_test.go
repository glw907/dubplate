package settings

import (
	"context"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// TestAccessor_SeesAnotherHandlesWrite opens two write handles on one
// store. A reads and caches the seeded value, B re-seeds it, and A's
// next read must return B's value: the generation B's apply bumped is
// all that tells A its warm cache is stale.
func TestAccessor_SeesAnotherHandlesWrite(t *testing.T) {
	dir := newStateDir(t)
	a, b := openWrite(t, dir), openWrite(t, dir)
	reg := newRegistry()
	keep := keepSetting("14")

	if err := apply(ctxFor(t), b, reg, []string{keyKeep + "=20"}, testNow, discardLog); err != nil {
		t.Fatalf("first apply: %v", err)
	}
	acc := NewAccessor(a, reg)
	got, err := acc.Int(context.Background(), keep)
	if err != nil {
		t.Fatalf("warm read: %v", err)
	}
	if got != 20 {
		t.Fatalf("warm read = %d, want 20", got)
	}

	if err := apply(ctxFor(t), b, reg, []string{keyKeep + "=30"}, testNow, discardLog); err != nil {
		t.Fatalf("second apply: %v", err)
	}
	got, err = acc.Int(context.Background(), keep)
	if err != nil {
		t.Fatalf("read after the write: %v", err)
	}
	if got != 30 {
		t.Errorf("read after another handle's write = %d, want 30", got)
	}
}

// TestAccessor_RefusesBadStoredValue proves a stored value outside the
// bounds or one that does not parse is an error naming the key, never
// the default.
func TestAccessor_RefusesBadStoredValue(t *testing.T) {
	tests := []struct {
		name    string
		row     fixtureRow
		wantErr []string
	}{
		{"out of bounds", fixtureRow{keyKeep, "2000", "admin", "20"}, []string{keyKeep, "above the ceiling 1000"}},
		{"does not parse", fixtureRow{keyKeep, "many", "admin", "20"}, []string{keyKeep, `invalid value "many"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.row)

			got, err := NewAccessor(h, newRegistry()).Int(context.Background(), keepSetting("14"))

			wantErrText(t, err, tt.wantErr...)
			if got != 0 {
				t.Errorf("read = %d alongside the error, want 0", got)
			}
		})
	}
}

// TestAccessor_ReadsPastAnotherKeysBadRow proves a read validates only
// the key it serves: the keep setting reads its value while the
// retention row is unparseable, out of bounds, or an orphan row sits
// beside it.
func TestAccessor_ReadsPastAnotherKeysBadRow(t *testing.T) {
	tests := []struct {
		name string
		row  fixtureRow
	}{
		{"another key unparseable", fixtureRow{keyRetention, "soon", "admin", "48h"}},
		{"another key out of bounds", fixtureRow{keyRetention, "1h", "admin", "48h"}},
		{"an orphan row", fixtureRow{keyOrphan, "7", "env-seed", "7"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.row, fixtureRow{keyKeep, "40", "env-seed", "40"})

			got, err := NewAccessor(h, newRegistry()).Int(context.Background(), keepSetting("14"))
			if err != nil {
				t.Fatalf("read keep: %v", err)
			}
			if got != 40 {
				t.Errorf("read keep = %d, want 40", got)
			}
		})
	}
}

// TestAccessor_MissingGeneration proves a store whose generation row
// is gone fails every read naming settings_generation, even with the
// cache warm, rather than serving the cached value.
func TestAccessor_MissingGeneration(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h, fixtureRow{keyKeep, "40", "env-seed", "40"})
	acc := NewAccessor(h, newRegistry())
	if got, err := acc.Int(context.Background(), keepSetting("14")); err != nil || got != 40 {
		t.Fatalf("warm read = %d, %v, want 40", got, err)
	}

	err := h.Update(ctxFor(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `DELETE FROM settings_generation`)
		return err
	})
	if err != nil {
		t.Fatalf("delete generation: %v", err)
	}

	got, err := acc.Int(context.Background(), keepSetting("14"))
	wantErrText(t, err, keyKeep, "settings_generation")
	if !errIsNoRows(err) {
		t.Errorf("err = %v, want it to wrap sql.ErrNoRows", err)
	}
	if got != 0 {
		t.Errorf("read = %d alongside the error, want 0", got)
	}
}

// guard is a float setting shaped like the backup guard fraction.
var guard = config.Setting{
	Key: "DUBPLATE_TEST_GUARD", Kind: config.KindFloat, Default: "0.5", Doc: "a fraction",
	Apply: config.ApplyNextOperation, Floor: "0.25", Ceiling: "1",
}

// TestAccessor_TypedReads reads each kind through its own method and
// refuses a read under the wrong kind or of an undeclared key.
func TestAccessor_TypedReads(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h,
		fixtureRow{keyRetention, "720h", "env-seed", "720h"},
		fixtureRow{guard.Key, "0.75", "env-seed", "0.75"})
	reg := newRegistry()
	reg.AddSetting(guard)
	acc := NewAccessor(h, reg)
	ctx := context.Background()

	if got, err := acc.Duration(ctx, retentionSetting); err != nil || got != 720*time.Hour {
		t.Errorf("Duration = %v, %v, want 720h", got, err)
	}
	if got, err := acc.Float(ctx, guard); err != nil || got != 0.75 {
		t.Errorf("Float = %v, %v, want 0.75", got, err)
	}
	if got, err := acc.Int(ctx, keepSetting("14")); err != nil || got != 14 {
		t.Errorf("Int = %v, %v, want the default 14", got, err)
	}

	_, err := acc.Int(ctx, retentionSetting)
	wantErrText(t, err, keyRetention, "KindDuration setting read as KindInt")

	unknown := keepSetting("14")
	unknown.Key = keyOrphan
	_, err = acc.Int(ctx, unknown)
	wantErrText(t, err, keyOrphan, "not a registered setting")
}
