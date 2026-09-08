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
			reg := config.NewRegistry()
			registerShadowIntegration(reg)
			registerShadowThresholds(reg)

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(tt.environ, reg, &stdout, &stderr)
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

			for _, row := range reg.OwnedWall() {
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

			for _, class := range reg.ReservedClasses() {
				wantLine := fmt.Sprintf("%-32s disabled (arrives at rung %d)", class.Name, class.Rung)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before reserved class %s", class.Name)
				}
				if lines[li] != wantLine {
					t.Errorf("reserved row %d = %q, want %q", li, lines[li], wantLine)
				}
				li++
			}

			for _, class := range reg.IntegrationClasses() {
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

			for _, th := range reg.Thresholds() {
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
	cmd := newDoctorCmd(baseEnviron(), config.NewRegistry(), &stdout, &stderr)
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

	// Resolution R15: reserved classes print in table order, not rung
	// order, and rung 13 (decision and review queues) precedes rung 12
	// (streamer) in that table. TestDoctorCmd itself cannot pin this: it
	// builds its expected lines by iterating config.ReservedClasses()
	// itself, so a registry that sorted both the production output and
	// the test's own expectations would still agree. This assertion pins
	// the order against the raw stdout instead.
	out := stdout.String()
	queuesIdx := strings.Index(out, "decision and review queues")
	streamerIdx := strings.Index(out, "streamer")
	if queuesIdx < 0 || streamerIdx < 0 {
		t.Fatalf("stdout missing a reserved class row: queues idx %d, streamer idx %d", queuesIdx, streamerIdx)
	}
	if queuesIdx >= streamerIdx {
		t.Errorf("decision and review queues printed at %d, streamer at %d; want queues before streamer (R15)", queuesIdx, streamerIdx)
	}
}

// TestDoctorCmd_ShadowRows proves the duration threshold renders as a
// duration rather than "invalid:", and the shadow integration row
// reflects its flag state.
func TestDoctorCmd_ShadowRows(t *testing.T) {
	reg := config.NewRegistry()
	registerShadowIntegration(reg)
	registerShadowThresholds(reg)

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	wantFloor := fmt.Sprintf("%-32s %s\n", "DUBPLATE_SHADOW_FLOOR", "720h0m0s")
	if !strings.Contains(stdout.String(), wantFloor) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantFloor)
	}

	wantBudget := fmt.Sprintf("%-32s %s\n", "DUBPLATE_SHADOW_BUDGET", "60")
	if !strings.Contains(stdout.String(), wantBudget) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantBudget)
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
	cmd = newDoctorCmd(enabledEnviron, reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	wantEnabled := fmt.Sprintf("%-32s %s\n", "shadow", "enabled")
	if !strings.Contains(stdout.String(), wantEnabled) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantEnabled)
	}
}

// TestDoctorCmd_FloatThreshold proves a registered KindFloat threshold
// prints its effective value in the thresholds block, in the same
// pinned "%-32s %s" form every other row uses, and that the value comes
// through the same round-tripping parser thresholdStatus itself calls
// (config.ParseFloat), rather than a captured golden string.
func TestDoctorCmd_FloatThreshold(t *testing.T) {
	reg := config.NewRegistry()
	reg.AddThreshold(config.Threshold{
		Key:     "DUBPLATE_TEST_STRONG_REC_THRESH",
		Kind:    config.KindFloat,
		Default: "0.04",
		Doc:     "a float between 0 and 1",
	})

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(baseEnviron(), reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	th := reg.Thresholds()[0]
	wantLine := fmt.Sprintf("%-32s %s\n", th.Key, thresholdStatus(baseEnviron(), th))
	if !strings.Contains(stdout.String(), wantLine) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantLine)
	}
	if !strings.Contains(stdout.String(), fmt.Sprintf("%-32s %s\n", "DUBPLATE_TEST_STRONG_REC_THRESH", "0.04")) {
		t.Errorf("stdout = %q, want the float row to print 0.04", stdout.String())
	}
}
