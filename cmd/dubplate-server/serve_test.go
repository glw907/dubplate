package main

import (
	"bytes"
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// productionRegistry builds a registry exactly as newRootCmd does, so
// this file's own tests exercise the same collision and derivation
// surface the binary does, on registry.go:24-31's own precedent.
func productionRegistry() *config.Registry {
	reg := config.NewRegistry()
	registerShadowIntegration(reg)
	registerShadowThresholds(reg)
	registerServeConfig(reg)
	return reg
}

func TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass(t *testing.T) {
	reg := productionRegistry()

	want := map[string]struct {
		kind  config.ThresholdKind
		deflt string
	}{
		"DUBPLATE_SERVE_BIND_ADDRS": {config.KindString, "127.0.0.1"},
		"DUBPLATE_SERVE_PORT":       {config.KindInt, "8080"},
		"DUBPLATE_SERVE_PREFIX":     {config.KindString, "/dubplate"},
	}

	got := make(map[string]config.Threshold)
	for _, th := range reg.Thresholds() {
		got[th.Key] = th
	}
	for key, w := range want {
		th, ok := got[key]
		if !ok {
			t.Errorf("Thresholds() missing %s", key)
			continue
		}
		if th.Kind != w.kind {
			t.Errorf("%s.Kind = %v, want %v", key, th.Kind, w.kind)
		}
		if th.Default != w.deflt {
			t.Errorf("%s.Default = %q, want %q", key, th.Default, w.deflt)
		}
	}

	for _, class := range reg.ReservedClasses() {
		if class.Name == "serve" {
			t.Error(`ReservedClasses() still carries the "serve" class, want it retired`)
		}
	}
}

func TestResolveServeConfig(t *testing.T) {
	tests := []struct {
		name       string
		environ    []string
		wantErr    error
		wantAddrs  []string
		wantPort   int
		wantPrefix string
	}{
		{
			name:       "defaults",
			environ:    nil,
			wantAddrs:  []string{"127.0.0.1"},
			wantPort:   8080,
			wantPrefix: "/dubplate",
		},
		{
			name: "configured",
			environ: []string{
				"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,::1",
				"DUBPLATE_SERVE_PORT=9090",
				"DUBPLATE_SERVE_PREFIX=/library",
			},
			wantAddrs:  []string{"127.0.0.1", "::1"},
			wantPort:   9090,
			wantPrefix: "/library",
		},
		{
			name:    "invalid prefix",
			environ: []string{"DUBPLATE_SERVE_PREFIX=no-leading-separator"},
			wantErr: serve.ErrInvalidPrefix,
		},
		{
			name:    "unparseable bind list",
			environ: []string{"DUBPLATE_SERVE_BIND_ADDRS=not-an-address"},
			wantErr: serve.ErrInvalidBindAddr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, addrs, port, err := resolveServeConfig(tt.environ)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("resolveServeConfig() error = %v, want it to wrap %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveServeConfig() error = %v, want nil", err)
			}
			if !slices.Equal(addrs, tt.wantAddrs) {
				t.Errorf("addrs = %v, want %v", addrs, tt.wantAddrs)
			}
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
			if prefix.String() != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix.String(), tt.wantPrefix)
			}
		})
	}
}

// TestNewServeCmd_RefusesInvalidConfigWithoutListening proves the verb
// validates the wall and resolves configuration, refusing before ever
// calling serve.Run: an unsatisfied owned wall, an invalid prefix, or
// an unparseable bind list each surfaces through RunE with the
// underlying sentinel intact, and no case opens a listener, since all
// three refusals precede resolution's own return.
//
// Each case runs under an already-canceled context so a regression
// reports as a failure rather than a hang: should a refusal stop
// firing, serve.Run returns at once instead of serving until the
// gate's own timeout.
func TestNewServeCmd_RefusesInvalidConfigWithoutListening(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		wantErr error
	}{
		{
			name:    "unsatisfied owned wall",
			environ: dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
			wantErr: config.ErrMissing,
		},
		{
			name:    "invalid prefix",
			environ: append(baseEnviron(), "DUBPLATE_SERVE_PREFIX=no-leading-separator"),
			wantErr: serve.ErrInvalidPrefix,
		},
		{
			name:    "unparseable bind list",
			environ: append(baseEnviron(), "DUBPLATE_SERVE_BIND_ADDRS=not-an-address"),
			wantErr: serve.ErrInvalidBindAddr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := newServeCmd(tt.environ, productionRegistry(), &stdout, &stderr)

			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			err := cmd.ExecuteContext(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want it to wrap %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewServeCmd_WiredIntoRootCmd proves newServeCmd's line sits
// inside newRootCmd's AddCommand block, beside newShadowCmd's: a verb
// constructed but never added is unreachable through the command tree
// cobra dispatches, and root.Find alone would not tell the two apart
// from a verb reachable only because some other command re-added it.
//
// The second half proves the wiring carries the production registry
// and the process environment rather than a construction that only
// looks right: dispatching "serve" through the root command with an
// invalid prefix in the environment refuses with the resolution
// sentinel, which can only happen if the added command resolved
// configuration. No listener opens, since resolution fails first.
func TestNewServeCmd_WiredIntoRootCmd(t *testing.T) {
	var serveCmd, shadowCmd bool
	for _, child := range newRootCmd().Commands() {
		switch child.Use {
		case "serve":
			serveCmd = true
		case "shadow":
			shadowCmd = true
		}
	}
	if !serveCmd {
		t.Error("newRootCmd().Commands() carries no serve verb, want it added beside shadow")
	}
	if !shadowCmd {
		t.Fatal("newRootCmd().Commands() carries no shadow verb; the AddCommand block this test reads against has moved")
	}

	// newRootCmd reads os.Environ() at construction, so the process
	// environment is this half's input. A developer's shell carries
	// DUBPLATE_* names of its own, and the wall would refuse them
	// before resolution ever ran; clearing each to the empty string is
	// what the wall reads as unset (config.wall.get's own convention).
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "DUBPLATE_") {
			t.Setenv(name, "")
		}
	}
	for _, kv := range baseEnviron() {
		name, value, _ := strings.Cut(kv, "=")
		t.Setenv(name, value)
	}
	t.Setenv("DUBPLATE_SERVE_PREFIX", "no-leading-separator")

	root := newRootCmd()
	root.SetArgs([]string{"serve"})
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	if err := root.Execute(); !errors.Is(err, serve.ErrInvalidPrefix) {
		t.Fatalf("Execute([]string{\"serve\"}) error = %v, want it to wrap %v", err, serve.ErrInvalidPrefix)
	}
}

// TestDoctorCmd_CarriesServeThresholdsAndNoServeReservedRow proves the
// net row change registerServeConfig plus the keys.go deletion makes:
// minus one reserved class, plus three thresholds.
func TestDoctorCmd_CarriesServeThresholdsAndNoServeReservedRow(t *testing.T) {
	reg := productionRegistry()
	lines := doctorLines(t, reg)

	for _, key := range []string{"DUBPLATE_SERVE_BIND_ADDRS", "DUBPLATE_SERVE_PORT", "DUBPLATE_SERVE_PREFIX"} {
		found := false
		for _, line := range lines {
			if strings.HasPrefix(line, key) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("doctor output missing a row for %s", key)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "serve ") || strings.HasPrefix(line, "serve\t") {
			t.Errorf("doctor output still carries a serve reserved-class row: %q", line)
		}
	}
}
