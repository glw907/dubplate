package webapp

import (
	"errors"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

func TestRowStatus(t *testing.T) {
	loaded := &config.Config{}
	tests := []struct {
		name       string
		cfg        *config.Config
		resolveErr error
		wantCalled bool
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "Load refused the environment",
			cfg:        nil,
			wantStatus: "not checked",
		},
		{
			name:       "a resolution that succeeds",
			cfg:        loaded,
			wantCalled: true,
			wantStatus: "ok",
		},
		{
			name:       "one refusal",
			cfg:        loaded,
			resolveErr: errors.New("DUBPLATE_THEME_PALETTE: \"nope\": no theme selector"),
			wantCalled: true,
			wantStatus: "invalid: DUBPLATE_THEME_PALETTE: \"nope\": no theme selector",
			wantFailed: true,
		},
		{
			name:       "a joined refusal folds onto one line",
			cfg:        loaded,
			resolveErr: errors.Join(errors.New("first"), errors.New("second")),
			wantCalled: true,
			wantStatus: "invalid: first; second",
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got *config.Config
			called := false
			status, failed := rowStatus(tt.cfg, func(cfg *config.Config) error {
				called, got = true, cfg
				return tt.resolveErr
			})
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
			if called != tt.wantCalled {
				t.Errorf("resolve called = %v, want %v", called, tt.wantCalled)
			}
			if called && got != tt.cfg {
				t.Errorf("resolve received %p, want the row's cfg %p", got, tt.cfg)
			}
		})
	}
}
