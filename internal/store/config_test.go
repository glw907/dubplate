package store

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// TestStoreBusyTimeout_Bounds proves the busy timeout's floor (1ms)
// and ceiling (2147483647ms), each a named unexported constant
// 0s, -1s, and 500us all fail the floor since a positive
// duration under one millisecond truncates to zero; 1000h
// (3600000000ms, the risk review's measured over-range value) fails
// the ceiling; 2147483647ms itself passes.
func TestStoreBusyTimeout_Bounds(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "zero", value: "0s", wantErr: true},
		{name: "negative", value: "-1s", wantErr: true},
		{name: "below one millisecond", value: "500us", wantErr: true},
		{name: "far above the ceiling", value: "1000h", wantErr: true},
		{name: "at the ceiling", value: "2147483647ms", wantErr: false},
		{name: "the default", value: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var environ []string
			if tt.value != "" {
				environ = []string{"DUBPLATE_STORE_BUSY_TIMEOUT=" + tt.value}
			}
			_, err := BusyTimeout(environ)
			if tt.wantErr {
				if !errors.Is(err, ErrOutOfRange) {
					t.Fatalf("BusyTimeout() error = %v, want one wrapping ErrOutOfRange", err)
				}
				if !strings.Contains(err.Error(), "DUBPLATE_STORE_BUSY_TIMEOUT") {
					t.Errorf("error = %q, want it to name DUBPLATE_STORE_BUSY_TIMEOUT", err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("BusyTimeout() error = %v, want nil", err)
			}
		})
	}
}

// TestStoreBusyTimeout_ParseFailure proves a parse failure is distinct
// from a range failure: a caller (storeRow) tells the two apart with
// errors.Is rather than by matching error text.
func TestStoreBusyTimeout_ParseFailure(t *testing.T) {
	environ := []string{"DUBPLATE_STORE_BUSY_TIMEOUT=not-a-duration"}
	_, err := BusyTimeout(environ)
	if err == nil {
		t.Fatal("BusyTimeout() error = nil, want a parse failure")
	}
	if errors.Is(err, ErrOutOfRange) {
		t.Errorf("BusyTimeout() error = %v, want a parse failure that does not wrap ErrOutOfRange", err)
	}
}

// TestRegisterStoreConfig proves the threshold registers with its
// default: doctor's own threshold row derives from
// [config.Registry.Thresholds], not from busyTimeoutThreshold
// directly, so a registration that never landed would leave the row
// silently absent rather than failing loudly.
func TestRegisterStoreConfig(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)

	found := false
	for _, th := range reg.Thresholds() {
		if th.Key == "DUBPLATE_STORE_BUSY_TIMEOUT" {
			found = true
			if th.Default != "15s" {
				t.Errorf("threshold default = %q, want %q", th.Default, "15s")
			}
		}
	}
	if !found {
		t.Fatal("Register did not register DUBPLATE_STORE_BUSY_TIMEOUT")
	}
}

// TestRegisterStoreConfig_Settings proves the store-homed setting
// registers with R-6c11's default.
func TestRegisterStoreConfig_Settings(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)

	want := map[string]string{
		"DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP": "3",
	}
	got := make(map[string]string)
	for _, s := range reg.Settings() {
		got[s.Key] = s.Default
	}
	for key, def := range want {
		if got[key] != def {
			t.Errorf("Register registered %s with default %q, want %q", key, got[key], def)
		}
	}
}

// TestStoreMigrationWait pins the threshold's default and bounds, each
// bound exactly at and one step past.
func TestStoreMigrationWait(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    time.Duration
		wantErr string
	}{
		{name: "default", want: 5 * time.Minute},
		{name: "floor", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=1s"}, want: time.Second},
		{name: "below the floor", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=999ms"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT must be at least 1s"},
		{name: "ceiling", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=24h"}, want: 24 * time.Hour},
		{name: "above the ceiling", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=24h1s"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT must be at most 24h"},
		{name: "unparseable", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=soon"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MigrationWait(tt.environ)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("MigrationWait() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("MigrationWait() = %v, %v, want %v, nil", got, err, tt.want)
			}
		})
	}
}
