package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// templateEnvPaths names every environment template these tests
// classify against, one entry per template scripts/deploy.sh renders
// (SERVER_ENV_TEMPLATE, rendered by its render_template).
var templateEnvPaths = []string{
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
// DUBPLATE_BACKUP_* (backup.go) or DUBPLATE_SERVE_*
// (serve.go), and classifying a template there would report every such
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
// direction: every owned wall row (a key [config.Load] requires)
// appears in
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
		if name == class.Flag || slices.Contains(class.Required, name) {
			return "integration class"
		}
	}
	for _, th := range reg.Thresholds() {
		if th.Key == name {
			return "threshold"
		}
	}
	for _, src := range reg.IdentitySources() {
		if name == src.Selector || slices.Contains(src.Required, name) || slices.Contains(src.Optional, name) {
			return "identity source"
		}
	}
	return ""
}

// TestClassifyKey_SMTPHostMovedToIdentitySource proves the transition
// rung 5 makes: DUBPLATE_SMTP_HOST classified as the "auth and
// identity" reserved class before this pass, and now classifies as the
// magic-link identity source's own required key
// (cmd/dubplate-server/auth.go), since that class retired in the same
// commit that declared the source (internal/config/keys.go).
func TestClassifyKey_SMTPHostMovedToIdentitySource(t *testing.T) {
	if got := classifyKey(productionRegistry(), "DUBPLATE_SMTP_HOST"); got != "identity source" {
		t.Errorf(`classifyKey("DUBPLATE_SMTP_HOST") = %q, want "identity source"`, got)
	}
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

// renderTemplate resolves template text the way scripts/deploy.sh's
// render_template does, one line at a time: a blank or '#' line is
// omitted; a line with no '=' is an error; NAME=value keeps value
// verbatim (a value may itself contain '='), unless values names NAME,
// in which case values replaces it; a bare NAME= takes its value from
// values and is an error when values has none. values only ever
// replaces the value of a row the template carries, so a row deleted
// from the template stays deleted. The result is one NAME=value per
// row, in file order.
func renderTemplate(template string, values map[string]string) ([]string, error) {
	var rows []string
	for line := range strings.SplitSeq(template, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, rhs, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("malformed line (no '='): %s", line)
		}
		value, replaced := values[name]
		switch {
		case replaced:
		case rhs != "":
			value = rhs
		default:
			return nil, fmt.Errorf("bare row %s has no value", name)
		}
		rows = append(rows, name+"="+value)
	}
	return rows, nil
}

func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name     string
		template string
		values   map[string]string
		want     []string
		wantErr  string
	}{
		{
			name:     "literal keeps its value",
			template: "A=one\n",
			want:     []string{"A=one"},
		},
		{
			name:     "literal containing equals keeps all of it",
			template: "A=x=y=z\n",
			want:     []string{"A=x=y=z"},
		},
		{
			name:     "bare row takes the fixture value",
			template: "A=\n",
			values:   map[string]string{"A": "fixture"},
			want:     []string{"A=fixture"},
		},
		{
			name:     "fixture replaces a literal's value",
			template: "A=one\n",
			values:   map[string]string{"A": "two"},
			want:     []string{"A=two"},
		},
		{
			name:     "blank and commented rows are omitted",
			template: "\n# A=one\nB=two\n",
			want:     []string{"B=two"},
		},
		{
			name:     "a fixture never appends a row the template lacks",
			template: "A=one\n",
			values:   map[string]string{"B": "two"},
			want:     []string{"A=one"},
		},
		{
			name:     "bare row with no fixture is an error",
			template: "A=\n",
			wantErr:  "bare row A has no value",
		},
		{
			name:     "line with no equals is an error",
			template: "A=one\nnot a row\n",
			wantErr:  "malformed line (no '='): not a row",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := renderTemplate(tt.template, tt.values)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("renderTemplate() error = %v, want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("renderTemplate() error = %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("renderTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

// renderedDoctor renders the greenfield template with fixture values
// for every bare row and for the rows doctor stats or opens (the state
// directory), appends the reserved serve port, and runs doctor against the
// production registry. stateDir carries an installed transliteration
// table and no store. It returns doctor's stdout and Execute's error.
// Only the template's own rows reach doctor, bar the serve port, which
// the template does not carry.
func renderedDoctor(t *testing.T, template string, stateDir string) (string, error) {
	t.Helper()
	secret := strings.Repeat("a", 32)
	environ, err := renderTemplate(template, map[string]string{
		"DUBPLATE_ADMIN_EMAIL":       "admin@example.test",
		"DUBPLATE_MAGIC_LINK_SECRET": secret,
		"DUBPLATE_SESSION_SECRET":    strings.Repeat("b", 32),
		"DUBPLATE_SMTP_HOST":         "smtp.example.test",
		"DUBPLATE_SMTP_PORT":         "587",
		"DUBPLATE_SMTP_USER":         "dubplate",
		"DUBPLATE_SMTP_PASSWORD":     "fixture smtp password, not a real key",
		"DUBPLATE_SMTP_FROM":         "dubplate@example.test",
		"DUBPLATE_STATE_DIR":         stateDir,
	})
	if err != nil {
		t.Fatalf("render template: %v", err)
	}
	environ = append(environ, "DUBPLATE_SERVE_PORT="+strconv.Itoa(reservedPort()))

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
	err = cmd.Execute()
	return stdout.String(), err
}

// TestRenderedTemplateDoctor runs doctor over the greenfield template
// rendered the way scripts/deploy.sh renders it, and pins each row's
// status. A render here is a parity model only: the real proof of the
// render is the deploy script's own --check run against the template,
// which resolves bare rows from the age store this test cannot reach.
func TestRenderedTemplateDoctor(t *testing.T) {
	raw, err := os.ReadFile("../../env/dubplate-server.env.template")
	if err != nil {
		t.Fatalf("read template: %v", err)
	}
	template := string(raw)

	tests := []struct {
		name      string
		template  string
		wantErr   bool
		wantLines []string // each must appear as a whole output line
		wantNone  []string // no output line may start with one of these
	}{
		{
			name:     "the rendered environment has no store yet and exits 0",
			template: template,
			wantLines: []string{
				fmt.Sprintf("%-32s %s", "DUBPLATE_IDENTITY_SOURCE", "ok"),
				fmt.Sprintf("%-32s %s", "magic-link", "active"),
				fmt.Sprintf("%-32s %s", "store", "absent"),
				fmt.Sprintf("%-32s %s", "members", "no store yet; start serve, then member add --admin"),
				fmt.Sprintf("%-32s %s", "unidecode", "ok (1.4.0)"),
				fmt.Sprintf("%-32s %s", "client ip", "direct: the peer must be the browser; a proxy in front must be configured as trusted"),
				fmt.Sprintf("%-32s %s", "web", "ok"),
				fmt.Sprintf("%-32s %s", "bind addresses", "127.0.0.1 present"),
				fmt.Sprintf("%-32s %s", "trusted proxy", "disabled"),
				fmt.Sprintf("%-32s %s", "boundary probe", "127.0.0.1 no listener"),
			},
			wantNone: []string{"roster ", "DUBPLATE_ROSTER_FILE"},
		},
		{
			name:     "a template carrying the retired roster row fails the load with the supersession",
			template: template + "\nDUBPLATE_ROSTER_FILE=/var/lib/dubplate/roster/roster.json\n",
			wantErr:  true,
			wantLines: []string{fmt.Sprintf("%-32s %s", "DUBPLATE_ROSTER_FILE", "invalid: superseded: members live in the store; "+
				"add the first admin with dubplate-server member add ADDRESS --admin")},
		},
		{
			name:      "deleting the trusted proxy flag row stays passing, since the class is off by default",
			template:  deleteTemplateRow(template, "DUBPLATE_TRUSTED_PROXY_ENABLED"),
			wantLines: []string{fmt.Sprintf("%-32s %s", "trusted proxy", "disabled")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			installTestUnidecodeTable(t, stateDir)

			out, err := renderedDoctor(t, tt.template, stateDir)
			if (err != nil) != tt.wantErr {
				t.Fatalf("doctor error = %v, wantErr %v\n%s", err, tt.wantErr, out)
			}
			lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
			for _, want := range tt.wantLines {
				if !slices.Contains(lines, want) {
					t.Errorf("doctor output lacks line %q\n%s", want, out)
				}
			}
			for _, prefix := range tt.wantNone {
				for _, line := range lines {
					if strings.HasPrefix(line, prefix) {
						t.Errorf("doctor output has line %q, want none starting %q", line, prefix)
					}
				}
			}
		})
	}
}

// deleteTemplateRow returns template without the row named name.
func deleteTemplateRow(template, name string) string {
	var kept []string
	for line := range strings.SplitSeq(template, "\n") {
		if strings.HasPrefix(line, name+"=") {
			continue
		}
		kept = append(kept, line)
	}
	return strings.Join(kept, "\n")
}
