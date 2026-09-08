package main

import (
	"bufio"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// templateEnvPaths names every environment template these tests
// classify against, one entry per adopter-facing template
// scripts/deploy.sh renders (ENV_TEMPLATE at deploy.sh:27,
// SERVER_ENV_TEMPLATE at deploy.sh:28, both rendered by render_template
// at deploy.sh:205, called at deploy.sh:240 and deploy.sh:241). A name
// missing from either fails deploy the same way, which is why both are
// read rather than one.
//
// dubplate.env.template is the frozen pre-pivot template, kept so a
// name the wall still classifies as superseded stays proven reachable
// from a real file rather than only from the superseded table's own
// literals. dubplate-server.env.template is the greenfield template,
// which the completeness direction below reads by name.
var templateEnvPaths = []string{
	"../../env/dubplate.env.template",
	"../../env/dubplate-server.env.template",
}

// TestTemplateKeysClassifiedAgainstProductionRegistry proves R10's
// classification direction: every DUBPLATE_* name any template under
// env/ lists classifies as owned, reserved, superseded, an integration
// class's flag or Required or Optional key, or a registered threshold,
// never unknown. A name that classified as unknown would make a real
// adopter's template-derived env file trip a load error the wall never
// intended.
//
// It lives in this package rather than in internal/config as a
// consequence of the rung 1b registry contract: a rung declares its
// keys in its own file, so internal/config's base registry cannot see
// DUBPLATE_SHADOW_* (rung 4, shadow.go) or DUBPLATE_SERVE_* (rung 5,
// serve.go), and classifying a template there would report every such
// name as unknown. Only the composed production registry knows the
// whole wall, and productionRegistry (serve_test.go) builds it exactly
// as newRootCmd does. Rung 5 is the pass that moved this test, because
// it is the first to declare a name in its own file that a template
// already lists and no reserved class covers.
func TestTemplateKeysClassifiedAgainstProductionRegistry(t *testing.T) {
	reg := productionRegistry()

	for _, path := range templateEnvPaths {
		names := templateKeys(t, path)
		if len(names) == 0 {
			t.Fatalf("%s: templateKeys found no DUBPLATE_* names; template path or format changed", path)
		}
		t.Logf("%s: classified %d DUBPLATE_* names", path, len(names))

		for _, name := range names {
			if classifyKey(reg, name) == "" {
				t.Errorf("%s: %s classifies as unknown; declare it in the rung's own file, or add it to a reserved class or the superseded table", path, name)
			}
		}
	}
}

// TestTemplateKeysCoverEveryOwnedNowRow asserts R10's completeness
// direction: every owned-now row (a wall row [config.Load] requires
// now rather than one carrying a Pending status) appears in
// env/dubplate-server.env.template, the template the greenfield server
// ships. This test never writes the template; a missing row is
// reported, not repaired, since adding a name there mints a value
// outside the age store.
func TestTemplateKeysCoverEveryOwnedNowRow(t *testing.T) {
	const greenfieldTemplate = "../../env/dubplate-server.env.template"
	names := templateKeys(t, greenfieldTemplate)
	if len(names) == 0 {
		t.Fatalf("%s: templateKeys found no DUBPLATE_* names; template path or format changed", greenfieldTemplate)
	}

	for _, row := range productionRegistry().OwnedWall() {
		if row.Pending != "" {
			continue
		}
		if !slices.Contains(names, row.Name) {
			t.Errorf("%s: owned-now row %s does not appear; a rung adding an owned-now row must add it to the greenfield template in the same pass", greenfieldTemplate, row.Name)
		}
	}
}

// classifyKey names the declaration family reg carries name in, or ""
// when reg declares it nowhere. It reads the registry's exported
// accessors rather than internal/config's own unexported lookups,
// which this package cannot reach.
func classifyKey(reg *config.Registry, name string) string {
	for _, row := range reg.OwnedWall() {
		if row.Name == name {
			return "owned wall"
		}
	}
	for _, class := range reg.ReservedClasses() {
		if slices.Contains(class.Keys, name) {
			return "reserved class"
		}
	}
	if _, ok := reg.Superseded()[name]; ok {
		return "superseded"
	}
	for _, class := range reg.IntegrationClasses() {
		if name == class.Flag || slices.Contains(class.Required, name) || slices.Contains(class.Optional, name) {
			return "integration class"
		}
	}
	for _, th := range reg.Thresholds() {
		if th.Key == name {
			return "threshold"
		}
	}
	return ""
}

// templateKeys parses the template at path for every DUBPLATE_* name
// assigned on its own line, in file order.
func templateKeys(t *testing.T, path string) []string {
	t.Helper()

	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open template: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close template: %v", err)
		}
	}()

	var names []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "DUBPLATE_") {
			continue
		}
		name, _, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		names = append(names, name)
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan template: %v", err)
	}
	return names
}
