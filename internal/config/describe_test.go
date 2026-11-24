package config

import (
	"strings"
	"testing"
)

// testDescription is the description every test declaration carries.
const testDescription = "a test key"

// testKey declares name with a description and the given sensitivity.
func testKey(name string, sensitivity Sensitivity) KeyDecl {
	return KeyDecl{Key: name, Description: testDescription, Sensitivity: sensitivity}
}

// describedFamilies lists, per declaration family, a function that
// registers one key built from the description and sensitivity it is
// handed.
// Every other key the declaration carries is fully described, so a
// panic can only come from the one under test.
var describedFamilies = []struct {
	name string
	add  func(r *Registry, key, description string, sensitivity Sensitivity)
}{
	{"wall row", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddWallRow(WallRow{Key: key, Description: description, Sensitivity: sensitivity})
	}},
	{"threshold", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddThreshold(Threshold{Key: key, Kind: KindInt, Default: "1", Description: description, Sensitivity: sensitivity})
	}},
	{"setting", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddSetting(Setting{
			Key: key, Kind: KindInt, Default: "1", Doc: "an integer", Apply: ApplyImmediately,
			Floor: "0", Ceiling: "9", Description: description, Sensitivity: sensitivity,
		})
	}},
	{"integration flag", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddIntegrationClass(IntegrationClass{
			Name: "test", Flag: KeyDecl{Key: key, Description: description, Sensitivity: sensitivity},
			Required: []KeyDecl{testKey("DUBPLATE_TEST_OTHER", Sensitive)},
		})
	}},
	{"integration required key", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddIntegrationClass(IntegrationClass{
			Name: "test", Flag: testKey("DUBPLATE_TEST_FLAG", NotSensitive),
			Required: []KeyDecl{{Key: key, Description: description, Sensitivity: sensitivity}},
		})
	}},
	{"identity source required key", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddIdentitySource(IdentitySource{
			Name: "test", Selector: "DUBPLATE_TEST_SELECTOR",
			Required: []KeyDecl{{Key: key, Description: description, Sensitivity: sensitivity}},
		})
	}},
	{"identity source optional key", func(r *Registry, key, description string, sensitivity Sensitivity) {
		r.AddIdentitySource(IdentitySource{
			Name: "test", Selector: "DUBPLATE_TEST_SELECTOR",
			Optional: []KeyDecl{{Key: key, Description: description, Sensitivity: sensitivity}},
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
				sensitivity Sensitivity
				want        string
			}{
				{"empty description", "", NotSensitive, "description"},
				{"blank description", " \t", Sensitive, "description"},
				{"zero sensitivity", "a test key", 0, "without a named sensitivity"},
				{"unnamed sensitivity", "a test key", Sensitive + 1, "without a named sensitivity"},
			}
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					msg := panicMessage(t, func() { fam.add(NewRegistry(), key, tt.description, tt.sensitivity) })
					if !strings.Contains(msg, key) || !strings.Contains(msg, tt.want) {
						t.Errorf("panic = %q, want it to name %s and %q", msg, key, tt.want)
					}
				})
			}
			for _, s := range []Sensitivity{NotSensitive, Sensitive} {
				fam.add(NewRegistry(), key, "a test key", s)
			}
		})
	}
}

func TestOwnedWall_RowsAreDescribedAndMarked(t *testing.T) {
	if len(ownedWall) == 0 {
		t.Fatal("ownedWall is empty")
	}
	for _, row := range ownedWall {
		if strings.TrimSpace(row.Description) == "" {
			t.Errorf("%s: no description", row.Key)
		}
		if row.Sensitivity != NotSensitive && row.Sensitivity != Sensitive {
			t.Errorf("%s: sensitivity = %v, want a named sensitivity", row.Key, row.Sensitivity)
		}
	}
}

func TestSensitivity_String(t *testing.T) {
	tests := []struct {
		sensitivity Sensitivity
		want        string
	}{
		{NotSensitive, "not sensitive"},
		{Sensitive, "sensitive"},
		{0, "Sensitivity(0)"},
		{7, "Sensitivity(7)"},
	}
	for _, tt := range tests {
		if got := tt.sensitivity.String(); got != tt.want {
			t.Errorf("Sensitivity(%d).String() = %q, want %q", int(tt.sensitivity), got, tt.want)
		}
	}
}
