package main

import (
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// TestReauthWindowSetting_Declaration pins the window's declaration in
// the production registry: a sensitive duration applied at the next
// operation, defaulting to GitHub sudo mode's two hours, between a
// magic link's round trip and a day.
func TestReauthWindowSetting_Declaration(t *testing.T) {
	for _, s := range newProductionRegistry().Settings() {
		if s.Key != "DUBPLATE_REAUTH_WINDOW" {
			continue
		}
		if s.Kind != config.KindDuration || s.Default != "2h" || s.Floor != "5m" || s.Ceiling != "24h" ||
			s.Apply != config.ApplyNextOperation || s.Mark != config.Sensitive {
			t.Errorf("DUBPLATE_REAUTH_WINDOW = %+v, want a sensitive next-operation duration, default 2h, 5m to 24h", s)
		}
		return
	}
	t.Fatal("DUBPLATE_REAUTH_WINDOW is not a registered setting")
}

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

// TestPrepareServe_ReauthWindowAtOrAboveTheCap proves serve's start
// refuses a window at or above the absolute cap through the seed rule,
// naming the window, the cap, and the remedies, and starts with one a
// second below it. Each fixture lowers the session lifetime to the cap
// or under, so the lifetime's own rule passes.
func TestPrepareServe_ReauthWindowAtOrAboveTheCap(t *testing.T) {
	tests := []struct {
		name    string
		env     []string
		wantErr []string
		notWant string
	}{
		{
			name: "a window a second below the cap starts",
			env: []string{"DUBPLATE_SESSION_ABSOLUTE_CAP=2h", "DUBPLATE_SESSION_LIFETIME=1h",
				"DUBPLATE_REAUTH_WINDOW=1h59m59s"},
		},
		{
			name: "a window equal to the cap is refused",
			env: []string{"DUBPLATE_SESSION_ABSOLUTE_CAP=2h", "DUBPLATE_SESSION_LIFETIME=1h",
				"DUBPLATE_REAUTH_WINDOW=2h"},
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "lower", "raise"},
		},
		{
			name:    "a cap below the default window is refused",
			env:     []string{"DUBPLATE_SESSION_ABSOLUTE_CAP=1h", "DUBPLATE_SESSION_LIFETIME=1h"},
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "lower", "raise"},
		},
		{
			name:    "a cap at the window's floor leaves raising it",
			env:     []string{"DUBPLATE_SESSION_ABSOLUTE_CAP=5m", "DUBPLATE_SESSION_LIFETIME=1m", "DUBPLATE_REAUTH_WINDOW=5m"},
			wantErr: []string{"DUBPLATE_REAUTH_WINDOW", "DUBPLATE_SESSION_ABSOLUTE_CAP", "raise"},
			notWant: "lower",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := serveAuthEnviron(t)
			for _, kv := range tt.env {
				name, value, _ := strings.Cut(kv, "=")
				environ = replaceKey(environ, name, value)
			}
			shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
			if tt.wantErr == nil {
				if err != nil || shell == nil {
					t.Fatalf("prepare = %v, %v, want a shell", shell, err)
				}
				return
			}
			if err == nil {
				t.Fatal("prepare error = nil, want the window's refusal")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
			if tt.notWant != "" && strings.Contains(err.Error(), tt.notWant) {
				t.Errorf("error = %q, want no %q remedy", err, tt.notWant)
			}
		})
	}
}
