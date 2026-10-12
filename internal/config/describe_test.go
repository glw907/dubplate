package config

import (
	"strings"
	"testing"
)

// testDescription is the description every test declaration carries.
const testDescription = "a test key"

// testKey declares name with a description and the given mark.
func testKey(name string, mark Sensitivity) KeyDecl {
	return KeyDecl{Name: name, Description: testDescription, Mark: mark}
}

// describedFamilies lists, per declaration family, a function that
// registers one key built from the description and mark it is handed.
// Every other key the declaration carries is fully described, so a
// panic can only come from the one under test.
var describedFamilies = []struct {
	name string
	add  func(r *Registry, key, description string, mark Sensitivity)
}{
	{"wall row", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddWallRow(WallRow{Name: key, Description: description, Mark: mark})
	}},
	{"threshold", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddThreshold(Threshold{Key: key, Kind: KindInt, Default: "1", Description: description, Mark: mark})
	}},
	{"setting", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddSetting(Setting{
			Key: key, Kind: KindInt, Default: "1", Doc: "an integer", Apply: ApplyImmediately,
			Floor: "0", Ceiling: "9", Description: description, Mark: mark,
		})
	}},
	{"integration flag", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddIntegrationClass(IntegrationClass{
			Name: "test", Flag: KeyDecl{Name: key, Description: description, Mark: mark},
			Required: []KeyDecl{testKey("DUBPLATE_TEST_OTHER", Sensitive)},
		})
	}},
	{"integration required key", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddIntegrationClass(IntegrationClass{
			Name: "test", Flag: testKey("DUBPLATE_TEST_FLAG", NotSensitive),
			Required: []KeyDecl{{Name: key, Description: description, Mark: mark}},
		})
	}},
	{"identity source required key", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddIdentitySource(IdentitySource{
			Name: "test", Selector: "DUBPLATE_TEST_SELECTOR",
			Required: []KeyDecl{{Name: key, Description: description, Mark: mark}},
		})
	}},
	{"identity source optional key", func(r *Registry, key, description string, mark Sensitivity) {
		r.AddIdentitySource(IdentitySource{
			Name: "test", Selector: "DUBPLATE_TEST_SELECTOR",
			Optional: []KeyDecl{{Name: key, Description: description, Mark: mark}},
		})
	}},
}

func TestRegistration_PanicsNamingTheKeyForAMissingDescriptionOrMark(t *testing.T) {
	const key = "DUBPLATE_TEST_UNDESCRIBED"
	for _, fam := range describedFamilies {
		t.Run(fam.name, func(t *testing.T) {
			cases := []struct {
				name        string
				description string
				mark        Sensitivity
				want        string
			}{
				{"empty description", "", NotSensitive, "description"},
				{"blank description", " \t", Sensitive, "description"},
				{"zero mark", "a test key", 0, "sensitive mark"},
				{"unnamed mark", "a test key", Sensitive + 1, "sensitive mark"},
			}
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					msg := panicMessage(t, func() { fam.add(NewRegistry(), key, tt.description, tt.mark) })
					if !strings.Contains(msg, key) || !strings.Contains(msg, tt.want) {
						t.Errorf("panic = %q, want it to name %s and %q", msg, key, tt.want)
					}
				})
			}
			for _, mark := range []Sensitivity{NotSensitive, Sensitive} {
				fam.add(NewRegistry(), key, "a test key", mark)
			}
		})
	}
}

func TestOwnedWall_RowsAreDescribedAndMarked(t *testing.T) {
	rows := OwnedWall()
	if len(rows) == 0 {
		t.Fatal("OwnedWall() is empty")
	}
	for _, row := range rows {
		if strings.TrimSpace(row.Description) == "" {
			t.Errorf("%s: no description", row.Name)
		}
		if row.Mark != NotSensitive && row.Mark != Sensitive {
			t.Errorf("%s: mark = %v, want a named mark", row.Name, row.Mark)
		}
	}
}

func TestSensitivity_String(t *testing.T) {
	tests := []struct {
		mark Sensitivity
		want string
	}{
		{NotSensitive, "not sensitive"},
		{Sensitive, "sensitive"},
		{0, "Sensitivity(0)"},
		{7, "Sensitivity(7)"},
	}
	for _, tt := range tests {
		if got := tt.mark.String(); got != tt.want {
			t.Errorf("Sensitivity(%d).String() = %q, want %q", int(tt.mark), got, tt.want)
		}
	}
}
