package config

import "testing"

func TestOwnedWall_IdentitySourceRow(t *testing.T) {
	rows := OwnedWall()

	var found *WallRow
	for i, row := range rows {
		if row.Name == "identity source" {
			found = &rows[i]
			break
		}
	}
	if found == nil {
		t.Fatal("OwnedWall() has no identity source row")
	}
	if found.Pending != "pending (arrives at rung 5)" {
		t.Errorf("identity source row Pending = %q, want %q", found.Pending, "pending (arrives at rung 5)")
	}
}

func TestOwnedWall_RequiredNowRows(t *testing.T) {
	want := map[string]bool{
		"DUBPLATE_LIBRARY_ROOT": true,
		"DUBPLATE_STATE_DIR":    true,
		"DUBPLATE_SITE_NAME":    true,
		"DUBPLATE_ADMIN_EMAIL":  true,
	}

	got := make(map[string]bool)
	for _, row := range OwnedWall() {
		if row.Pending == "" {
			got[row.Name] = true
		}
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

func TestOwnedWall_ReturnsCopy(t *testing.T) {
	rows := OwnedWall()
	rows[0].Name = "mutated"

	if ownedWall[0].Name == "mutated" {
		t.Fatal("OwnedWall() leaked the backing array; a caller can corrupt the package table")
	}
}

func TestReservedClassOf(t *testing.T) {
	tests := []struct {
		key       string
		wantClass string
		wantRung  int
		wantFound bool
	}{
		{key: "DUBPLATE_R2_BUCKET", wantClass: "backup targets", wantRung: 6, wantFound: true},
		{key: "DUBPLATE_SMTP_HOST", wantClass: "auth and identity", wantRung: 5, wantFound: true},
		{key: "DUBPLATE_NAVIDROME_URL", wantClass: "streamer", wantRung: 12, wantFound: true},
		{key: "DUBPLATE_UPLOADS_DIR", wantClass: "intake", wantRung: 7, wantFound: true},
		{key: "DUBPLATE_PING_IMPORT_URL", wantClass: "monitoring", wantRung: 9, wantFound: true},
		{key: "DUBPLATE_LIBRARY_ROOT", wantFound: false},
		{key: "DUBPLATE_NONSENSE", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			class, ok := reservedClassOf(tt.key)
			if ok != tt.wantFound {
				t.Fatalf("reservedClassOf(%q) found = %v, want %v", tt.key, ok, tt.wantFound)
			}
			if !ok {
				return
			}
			if class.Name != tt.wantClass {
				t.Errorf("reservedClassOf(%q) class = %q, want %q", tt.key, class.Name, tt.wantClass)
			}
			if class.Rung != tt.wantRung {
				t.Errorf("reservedClassOf(%q) rung = %d, want %d", tt.key, class.Rung, tt.wantRung)
			}
		})
	}
}

func TestReservedClasses_ReturnsCopy(t *testing.T) {
	classes := ReservedClasses()
	classes[0].Name = "mutated"

	if reservedClasses[0].Name == "mutated" {
		t.Fatal("ReservedClasses() leaked the backing array; a caller can corrupt the package table")
	}
}

func TestSupersessionOf(t *testing.T) {
	tests := []struct {
		key       string
		wantFound bool
	}{
		{key: "DUBPLATE_LIBRARY_DIR", wantFound: true},
		{key: "DUBPLATE_SERVE_SITE_NAME", wantFound: true},
		{key: "DUBPLATE_BEETS_DB", wantFound: true},
		{key: "DUBPLATE_TUNNEL_TOKEN", wantFound: true},
		{key: "DUBPLATE_LIBRARY_ROOT", wantFound: false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			_, ok := supersessionOf(tt.key)
			if ok != tt.wantFound {
				t.Errorf("supersessionOf(%q) found = %v, want %v", tt.key, ok, tt.wantFound)
			}
		})
	}
}
