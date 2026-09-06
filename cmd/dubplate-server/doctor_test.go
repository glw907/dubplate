package main

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

func TestDoctorCmd(t *testing.T) {
	tests := []struct {
		name         string
		environ      []string
		wantErr      bool
		rowOverrides map[string]string // owned row name -> expected status, overriding "ok"
	}{
		{
			name:    "full valid wall",
			environ: baseEnviron(),
			wantErr: false,
		},
		{
			name:    "missing two keys",
			environ: dropKey(dropKey(baseEnviron(), "DUBPLATE_LIBRARY_ROOT"), "DUBPLATE_SITE_NAME"),
			wantErr: true,
			rowOverrides: map[string]string{
				"DUBPLATE_LIBRARY_ROOT": "missing",
				"DUBPLATE_SITE_NAME":    "missing",
			},
		},
		{
			name: "bad ADMIN_EMAIL",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
				"DUBPLATE_ADMIN_EMAIL=not-an-address"),
			wantErr: true,
			rowOverrides: map[string]string{
				"DUBPLATE_ADMIN_EMAIL": "invalid: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registerShadowThresholds()

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(tt.environ, &stdout, &stderr)
			cmd.SetOut(&stdout)
			err := cmd.Execute()

			if tt.wantErr {
				if err == nil {
					t.Fatal("Execute() error = nil, want error")
				}
				if got := err.Error(); got != "configuration is incomplete" {
					t.Errorf("Execute() error = %q, want %q", got, "configuration is incomplete")
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty: doctor's table is stdout-only, main prints the returned error", got)
			}

			lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
			li := 0

			for _, row := range config.OwnedWall() {
				want := "ok"
				switch {
				case row.Pending != "":
					want = row.Pending
				case tt.rowOverrides[row.Name] != "":
					want = tt.rowOverrides[row.Name]
				}
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before owned row %s", row.Name)
				}
				gotLine := lines[li]
				wantPrefix := fmt.Sprintf("%-32s %s", row.Name, want)
				if !strings.HasPrefix(gotLine, wantPrefix) {
					t.Errorf("owned row %d = %q, want prefix %q", li, gotLine, wantPrefix)
				}
				li++
			}

			for _, class := range config.ReservedClasses() {
				wantLine := fmt.Sprintf("%-32s disabled (arrives at rung %d)", class.Name, class.Rung)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before reserved class %s", class.Name)
				}
				if lines[li] != wantLine {
					t.Errorf("reserved row %d = %q, want %q", li, lines[li], wantLine)
				}
				li++
			}

			for _, class := range config.IntegrationClasses() {
				status := "disabled"
				if enabled, err := class.Enabled(tt.environ); err == nil && enabled {
					status = "enabled"
				}
				wantLine := fmt.Sprintf("%-32s %s", class.Name, status)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before integration class %s", class.Name)
				}
				if lines[li] != wantLine {
					t.Errorf("integration row %d = %q, want %q", li, lines[li], wantLine)
				}
				li++
			}

			for _, th := range config.Thresholds() {
				want := thresholdStatus(tt.environ, th)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before threshold row %s", th.Key)
				}
				wantPrefix := fmt.Sprintf("%-32s %s", th.Key, want)
				if !strings.HasPrefix(lines[li], wantPrefix) {
					t.Errorf("threshold row %d = %q, want prefix %q", li, lines[li], wantPrefix)
				}
				li++
			}

			if li != len(lines) {
				t.Errorf("stdout has %d extra line(s) beyond owned+reserved+threshold rows: %v", len(lines)-li, lines[li:])
			}
		})
	}
}

// TestDoctorCmd_LiteralSpotCheck hard-codes two known lines so the
// table-driven test above cannot go vacuous if OwnedWall ever returned
// an empty table.
func TestDoctorCmd_LiteralSpotCheck(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	want := fmt.Sprintf("%-32s %s\n", "DUBPLATE_ADMIN_EMAIL", "ok")
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}

	wantIdentity := fmt.Sprintf("%-32s %s\n", "identity source", "pending (arrives at rung 5)")
	if !strings.Contains(stdout.String(), wantIdentity) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantIdentity)
	}
}

// TestDoctorCmd_ShadowRows proves the duration threshold renders as a
// duration rather than "invalid:", and the shadow integration row
// reflects its flag state.
func TestDoctorCmd_ShadowRows(t *testing.T) {
	registerShadowThresholds()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	wantFloor := fmt.Sprintf("%-32s %s\n", "DUBPLATE_SHADOW_FLOOR", "720h0m0s")
	if !strings.Contains(stdout.String(), wantFloor) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantFloor)
	}

	wantDisabled := fmt.Sprintf("%-32s %s\n", "shadow", "disabled")
	if !strings.Contains(stdout.String(), wantDisabled) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantDisabled)
	}

	enabledEnviron := append(baseEnviron(),
		"DUBPLATE_SHADOW_ENABLED=true",
		"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=/var/lib/incumbent",
		"DUBPLATE_SHADOW_REVIEW_DIR=/srv/review",
	)
	stdout.Reset()
	cmd = newDoctorCmd(enabledEnviron, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	wantEnabled := fmt.Sprintf("%-32s %s\n", "shadow", "enabled")
	if !strings.Contains(stdout.String(), wantEnabled) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantEnabled)
	}
}
