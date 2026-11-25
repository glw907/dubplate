package backup

import (
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/store"
)

// testNow anchors every age computation in this file's goldens, so a
// wall-clock read can never make a test flaky.
var testNow = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

// TestFactsRows is the golden test for the four backup rows the store
// facts feed, each built from a constructed [store.Report] rather than
// a real store: the "not checked" row (checked false, or a report whose
// Facts is nil), WAL size absent and present, free space known and
// unknown, local copies none and several, and last success never, an
// age, and an age with a prune's skipped or failed text appended.
func TestFactsRows(t *testing.T) {
	notChecked := []doctor.Row{
		{Name: "backup last success", Status: "not checked"},
		{Name: "backup wal size", Status: "not checked"},
		{Name: "backup free space", Status: "not checked"},
		{Name: "backup local copies", Status: "not checked"},
	}
	current := func(f store.Facts) store.Report { return store.Report{State: store.StateCurrent, Facts: &f} }
	tests := []struct {
		name   string
		report store.Report
		last   map[string]ops.JobSuccess
		want   []doctor.Row
	}{
		{
			name:   "not checked at all: the store row never ran Inspect",
			report: store.Report{},
			last:   map[string]ops.JobSuccess{ops.KindBackup: {Finished: testNow}},
			want:   notChecked,
		},
		{
			name:   "inspected, but the store carries no Facts (absent, busy, older, newer, not a store, refused, or mid-restore)",
			report: store.Report{State: store.StateAbsent},
			want:   notChecked,
		},
		{
			name:   "no succeeded backup: never, no wal, free known, no copies",
			report: current(store.Facts{FreeBytes: 2048, FreeKnown: true}),
			last:   map[string]ops.JobSuccess{"scan": {Finished: testNow}},
			want: []doctor.Row{
				{Name: "backup last success", Status: "never"},
				{Name: "backup wal size", Status: "0 B"},
				{Name: "backup free space", Status: "2.0 KiB"},
				{Name: "backup local copies", Status: "0 copies"},
			},
		},
		{
			name: "a clean succeeded backup: age alone, wal present, free unknown, several copies",
			report: current(store.Facts{
				WALSize:    4096,
				Copies:     3,
				CopiesSize: 3*1024*1024 + 512*1024,
			}),
			last: map[string]ops.JobSuccess{ops.KindBackup: {Finished: testNow.Add(-95 * time.Minute), Progress: finishedMessage}},
			want: []doctor.Row{
				{Name: "backup last success", Status: "1h35m ago"},
				{Name: "backup wal size", Status: "4.0 KiB"},
				{Name: "backup free space", Status: "unknown"},
				{Name: "backup local copies", Status: "3 copies, 3.5 MiB"},
			},
		},
		{
			name:   "a succeeded backup whose prune was skipped names the reason",
			report: current(store.Facts{}),
			last: map[string]ops.JobSuccess{ops.KindBackup: {
				Finished: testNow.Add(-30 * time.Second), Progress: "prune skipped: the run's stamp is not the newest on the remote",
			}},
			want: []doctor.Row{
				{Name: "backup last success", Status: "1m ago: prune skipped: the run's stamp is not the newest on the remote"},
				{Name: "backup wal size", Status: "0 B"},
				{Name: "backup free space", Status: "unknown"},
				{Name: "backup local copies", Status: "0 copies"},
			},
		},
		{
			name:   "a succeeded backup whose prune failed names the reason",
			report: current(store.Facts{}),
			last: map[string]ops.JobSuccess{ops.KindBackup: {
				Finished: testNow.Add(-25 * time.Hour), Progress: "prune failed: list remote snapshots: exit status 1",
			}},
			want: []doctor.Row{
				{Name: "backup last success", Status: "1d1h ago: prune failed: list remote snapshots: exit status 1"},
				{Name: "backup wal size", Status: "0 B"},
				{Name: "backup free space", Status: "unknown"},
				{Name: "backup local copies", Status: "0 copies"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := factsRows(testNow, tt.report, tt.last)
			if !slices.Equal(got, tt.want) {
				t.Errorf("factsRows() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestFormatBytes proves the binary-unit scale directly, so
// TestFactsRows's goldens are not the only proof a mutated divisor or
// exponent step would fail.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.n); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

// TestFormatAge proves the minute-rounded age rendering across every
// scale it names: minutes alone, hours and minutes, and days and
// hours, plus a future finish time clamped to "0m" rather than a
// negative duration.
func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{-time.Minute, "0m"},
		{30 * time.Second, "1m"},
		{23 * time.Minute, "23m"},
		{95 * time.Minute, "1h35m"},
		{25 * time.Hour, "1d1h"},
	}
	for _, tt := range tests {
		if got := formatAge(tt.d); got != tt.want {
			t.Errorf("formatAge(%s) = %q, want %q", tt.d, got, tt.want)
		}
	}
}
