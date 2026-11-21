package admin

import (
	"strings"
	"testing"
)

// TestReauthWindowBelowCap proves the window's correlated rule is
// strict: a window equal to the absolute cap is refused, since the
// gate would then hold nothing a session could still reach, and one a
// second below it is accepted.
func TestReauthWindowBelowCap(t *testing.T) {
	tests := []struct {
		name, window, cap string
		wantErr           []string
		notWant           string
	}{
		{name: "a second below the cap", window: "1h59m59s", cap: "2h"},
		{name: "equal to the cap", window: "2h", cap: "2h",
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "lower", "raise"}},
		{name: "above the cap", window: "3h", cap: "2h",
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "lower", "raise"}},
		{name: "a cap at the window's floor", window: "5m", cap: "5m",
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "raise"}, notWant: "lower"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := reauthWindowBelowCap.Check(map[string]string{
				"DUBPLATE_REAUTH_WINDOW":        tt.window,
				"DUBPLATE_SESSION_ABSOLUTE_CAP": tt.cap,
			})
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("Check = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Check = nil, want a refusal")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("refusal = %q, want it to contain %q", err, want)
				}
			}
			if tt.notWant != "" && strings.Contains(err.Error(), tt.notWant) {
				t.Errorf("refusal = %q, want no %q remedy", err, tt.notWant)
			}
		})
	}
}
