package settings

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestReaders reads each kind through its Reader and holds a failed
// read to one [*ReadError] whose text names the key once, ahead of the
// reason.
func TestReaders(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h,
		fixtureRow{keyRetention, "720h", "env_seed", "720h"},
		fixtureRow{guard.Key, "0.75", "env_seed", "0.75"})
	reg := newRegistry()
	reg.AddSetting(guard)
	acc := NewAccessor(h, reg)
	ctx := context.Background()

	if got, err := acc.DurationReader(retentionSetting)(ctx); err != nil || got != 720*time.Hour {
		t.Errorf("DurationReader = %v, %v, want 720h", got, err)
	}
	if got, err := acc.FloatReader(guard)(ctx); err != nil || got != 0.75 {
		t.Errorf("FloatReader = %v, %v, want 0.75", got, err)
	}
	if got, err := acc.IntReader(keepSetting("14"))(ctx); err != nil || got != 14 {
		t.Errorf("IntReader = %v, %v, want the default 14", got, err)
	}

	unknown := keepSetting("14")
	unknown.Key = keyOrphan
	_, err := acc.IntReader(unknown)(ctx)
	re, ok := errors.AsType[*ReadError](err)
	if !ok {
		t.Fatalf("IntReader of an undeclared key = %v, want a *ReadError", err)
	}
	if re.Key != keyOrphan {
		t.Errorf("ReadError.Key = %q, want %q", re.Key, keyOrphan)
	}
	if got, want := err.Error(), "DUBPLATE_TEST_ORPHAN: not a registered setting"; got != want {
		t.Errorf("err = %q, want %q", got, want)
	}

	_, err = acc.Int(ctx, retentionSetting)
	if got, want := err.Error(), "DUBPLATE_TEST_RETENTION: a KindDuration setting read as KindInt"; got != want {
		t.Errorf("err = %q, want %q", got, want)
	}
}
