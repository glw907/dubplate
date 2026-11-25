package config

import "testing"

func TestOwnedWall_RequiredNowRows(t *testing.T) {
	want := map[string]bool{
		"DUBPLATE_LIBRARY_ROOT": true,
		"DUBPLATE_STATE_DIR":    true,
		"DUBPLATE_SITE_NAME":    true,
		"DUBPLATE_ADMIN_EMAIL":  true,
	}

	got := make(map[string]bool)
	for _, row := range ownedWall {
		got[row.Key] = true
	}

	for name := range want {
		if !got[name] {
			t.Errorf("OwnedWall() missing required-now row %s", name)
		}
	}
	for name := range got {
		if !want[name] {
			t.Errorf("OwnedWall() has unexpected required-now row %s", name)
		}
	}
}

// Both table-driven tests below (TestReservedClassOf and
// TestSupersessionOf) name their fixture field input rather than key. A
// struct field named key sitting on the same line as a value whose
// entropy clears gitleaks' generic-api-key threshold trips the rule,
// because the rule needs both halves: a key-like field name and a
// high-entropy neighbor. The same discipline applies to any
// declaration file that writes a key literal beside a field named Key.

// DUBPLATE_SMTP_HOST no longer has a row here: rung 5 retired the "auth
// and identity" reserved class it belonged to and declared it instead
// as the magic-link identity source's own required key
// (internal/auth). internal/config's own NewRegistry
// never sees that declaration (the rung 1b registry contract), so its
// classification as "identity source" is proved against the production
// registry instead, in cmd/dubplate-server/templatekeys_test.go's
// TestClassifyKey_SMTPHostMovedToIdentitySource.
func TestReservedClassOf(t *testing.T) {
	tests := []struct {
		input     string
		wantClass string
		wantFound bool
	}{
		{input: "DUBPLATE_R2_BUCKET", wantFound: false},
		{input: "DUBPLATE_REVIEW_DIR", wantClass: "decision and review queues", wantFound: true},
		{input: "DUBPLATE_UPLOADS_DIR", wantFound: false},
		{input: "DUBPLATE_TRIGGER_DIR", wantFound: false},
		{input: "DUBPLATE_STAGING_DIR", wantFound: false},
		{input: "DUBPLATE_INBOX_DIR", wantClass: "pipeline", wantFound: true},
		{input: "DUBPLATE_PING_IMPORT_URL", wantClass: "monitoring", wantFound: true},
		{input: "DUBPLATE_LIBRARY_ROOT", wantFound: false},
		{input: "DUBPLATE_NONSENSE", wantFound: false},
	}

	reg := NewRegistry()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			class, ok := reg.reservedClassOf(tt.input)
			if ok != tt.wantFound {
				t.Fatalf("reservedClassOf(%q) found = %v, want %v", tt.input, ok, tt.wantFound)
			}
			if !ok {
				return
			}
			if class.Name != tt.wantClass {
				t.Errorf("reservedClassOf(%q) class = %q, want %q", tt.input, class.Name, tt.wantClass)
			}
		})
	}
}

// TestSupersessionOf also names its fixture field input, for the
// gitleaks reason stated above TestReservedClassOf.
func TestSupersessionOf(t *testing.T) {
	tests := []struct {
		input     string
		wantFound bool
	}{
		{input: "DUBPLATE_LIBRARY_DIR", wantFound: true},
		{input: "DUBPLATE_SERVE_SITE_NAME", wantFound: true},
		{input: "DUBPLATE_BEETS_DB", wantFound: true},
		{input: "DUBPLATE_TUNNEL_TOKEN", wantFound: true},
		{input: "DUBPLATE_R2_ACCOUNT_ID", wantFound: true},
		{input: "DUBPLATE_R2_BUCKET", wantFound: true},
		{input: "DUBPLATE_R2_REMOTE", wantFound: true},
		{input: "DUBPLATE_R2_ACCESS_KEY_ID", wantFound: true},
		{input: "DUBPLATE_R2_SECRET_ACCESS_KEY", wantFound: true},
		{input: "DUBPLATE_BACKUP_STAGING", wantFound: true},
		{input: "DUBPLATE_LIBRARY_ROOT", wantFound: false},
	}

	reg := NewRegistry()
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			_, ok := reg.supersessionOf(tt.input)
			if ok != tt.wantFound {
				t.Errorf("supersessionOf(%q) found = %v, want %v", tt.input, ok, tt.wantFound)
			}
		})
	}
}

// TestIntegrationClassesCoverEveryKey proves no integration key
// is declared without appearing in its own class: every name a class
// claims (its Flag and its Required rows) resolves
// back to that same class through integrationClassOf, which is what
// rejectNamespace's acceptance path depends on. Retargeted at the
// registry's integration family (Resolution R5), driven by
// fixtureIntegrationClass (config_test.go), the fixture this package's
// own tests declare.
func TestIntegrationClassesCoverEveryKey(t *testing.T) {
	reg := fixtureRegistry()
	classes := reg.IntegrationClasses()
	if len(classes) != 2 {
		t.Fatalf("IntegrationClasses len = %d, want 2", len(classes))
	}
	c := classes[0]
	if c.Name != fixtureIntegrationClass.Name || c.Flag != fixtureIntegrationClass.Flag {
		t.Errorf("class = %+v, want the fixture class", c)
	}
	if len(c.Required) != 2 {
		t.Errorf("Required = %d, want 2", len(c.Required))
	}
	names := KeyNames(append([]KeyDecl{c.Flag}, c.Required...))
	for _, name := range names {
		if got, ok := reg.integrationClassOf(name); !ok || got.Name != c.Name {
			t.Errorf("integrationClassOf(%s) = %v, %v; want the fixture class", name, got.Name, ok)
		}
	}
}
