package config

import (
	"strings"
	"testing"
	"time"
)

// withRegistry saves and restores the package-level threshold
// registry around a test, so tests that call Register do not leak
// entries into unrelated tests.
func withRegistry(t *testing.T) {
	t.Helper()
	saved := registeredThresholds
	registeredThresholds = nil
	t.Cleanup(func() { registeredThresholds = saved })
}

func TestRegister_AppearsInThresholds(t *testing.T) {
	withRegistry(t)

	if got := Thresholds(); len(got) != 0 {
		t.Fatalf("Thresholds() before Register = %v, want empty", got)
	}

	first := Threshold{Key: "DUBPLATE_TEST_FIRST", Default: "1", Doc: "a positive integer"}
	second := Threshold{Key: "DUBPLATE_TEST_SECOND", Default: "2", Doc: "a positive integer"}
	Register(first)
	Register(second)

	got := Thresholds()
	if len(got) != 2 {
		t.Fatalf("Thresholds() len = %d, want 2", len(got))
	}
	if got[0] != first || got[1] != second {
		t.Errorf("Thresholds() = %+v, want [%+v %+v] in registration order", got, first, second)
	}
}

func TestThresholds_ReturnsCopy(t *testing.T) {
	withRegistry(t)
	Register(Threshold{Key: "DUBPLATE_TEST_ONE", Default: "1", Doc: "a positive integer"})

	got := Thresholds()
	got[0].Key = "mutated"

	if registeredThresholds[0].Key == "mutated" {
		t.Fatal("Thresholds() leaked the backing array; a caller can corrupt the registry")
	}
}

func TestParseInt(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_QUIESCENT_MIN", Default: "10", Doc: "a positive integer number of minutes"}

	tests := []struct {
		name    string
		environ []string
		want    int
		wantErr bool
		errSub  string
	}{
		{
			name:    "unset uses default",
			environ: nil,
			want:    10,
		},
		{
			name:    "empty value uses default",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN="},
			want:    10,
		},
		{
			name:    "set value parses",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN=30"},
			want:    30,
		},
		{
			name:    "malformed value errors naming key and doc",
			environ: []string{"DUBPLATE_TEST_QUIESCENT_MIN=soon"},
			wantErr: true,
			errSub:  "DUBPLATE_TEST_QUIESCENT_MIN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseInt(tt.environ, th)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseInt() error = nil, want error containing %q", tt.errSub)
				}
				if got := err.Error(); !strings.Contains(got, tt.errSub) || !strings.Contains(got, th.Doc) {
					t.Errorf("ParseInt() error = %q, want it to contain %q and %q", got, tt.errSub, th.Doc)
				}
				return
			}

			if err != nil {
				t.Fatalf("ParseInt() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("ParseInt() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestParseInt_InvalidDefaultIsRegistrationError(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_TEST_BROKEN_DEFAULT", Default: "not-a-number", Doc: "a positive integer"}

	_, err := ParseInt(nil, th)
	if err == nil {
		t.Fatal("ParseInt() error = nil, want error for an unparsable Default")
	}
	if !strings.Contains(err.Error(), th.Key) {
		t.Errorf("ParseInt() error = %q, want it to contain %q", err.Error(), th.Key)
	}
}

func TestParseDuration(t *testing.T) {
	th := Threshold{Key: "DUBPLATE_SHADOW_FLOOR", Kind: KindDuration, Default: "720h", Doc: "a Go duration"}
	tests := []struct {
		name    string
		environ []string
		want    time.Duration
		wantErr bool
	}{
		{name: "unset takes the default", want: 720 * time.Hour},
		{name: "empty takes the default", environ: []string{"DUBPLATE_SHADOW_FLOOR="}, want: 720 * time.Hour},
		{name: "set wins", environ: []string{"DUBPLATE_SHADOW_FLOOR=48h"}, want: 48 * time.Hour},
		{name: "a bad value errors", environ: []string{"DUBPLATE_SHADOW_FLOOR=soon"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDuration(tt.environ, th)
			if tt.wantErr {
				if err == nil {
					t.Fatal("ParseDuration() error = nil, want error")
				}
				if !strings.Contains(err.Error(), th.Key) || !strings.Contains(err.Error(), th.Doc) {
					t.Errorf("ParseDuration() error = %q, want it to contain %q and %q", err.Error(), th.Key, th.Doc)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDuration() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("ParseDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseDurationRejectsABadDefaultAsARegistrationBug(t *testing.T) {
	_, err := ParseDuration(nil, Threshold{Key: "K", Kind: KindDuration, Default: "not a duration"})
	if err == nil || !strings.Contains(err.Error(), "K") {
		t.Errorf("ParseDuration on a bad default = %v; want an error naming the key", err)
	}
}
