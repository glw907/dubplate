package xmlrule

import (
	"encoding/json"
	"flag"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the additions golden")

// The committed OpenAPI bundle, pinned by the grader beside its sum.
const bundlePath = "../data/openapi/openapi.json"

// shelfRenames is the unit schema's rename table: its envelope and its
// fault type have OpenAPI counterparts of other names.
var shelfRenames = map[string]string{"Envelope": "EnvelopeBase", "Fault": "Problem"}

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

// systemRunner returns the workstation's checked xmllint, failing when it
// is absent or not the pinned version.
func systemRunner(t *testing.T) *runner {
	t.Helper()
	r, err := systemXmllint()
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// shelfValidator builds a validator over the unit schema and document,
// validating with run.
func shelfValidator(t *testing.T, run *runner) *Validator {
	t.Helper()
	path, err := filepath.Abs("testdata/shelf.xsd")
	if err != nil {
		t.Fatal(err)
	}
	v, err := newValidator(readFile(t, path), path, readFile(t, "testdata/shelf.openapi.json"), shelfRenames, run)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// realValidator builds a validator over the cached XSD and the committed
// bundle, through New.
func realValidator(t *testing.T) *Validator {
	t.Helper()
	v, err := New(readFile(t, bundlePath))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

// TestAdditionsGolden pins each XSD type's additions at the pinned XSD
// and bundle.
func TestAdditionsGolden(t *testing.T) {
	v := realValidator(t)
	got := make(map[string][]string)
	for name, typ := range v.schema.types {
		// A type with no additions is pinned as [], never null.
		got[name] = append([]string{}, slices.Sorted(maps.Keys(typ.additions))...)
	}
	data, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	golden := "testdata/additions.golden.json"
	if *update {
		if err := os.WriteFile(golden, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if want := readFile(t, golden); string(data) != string(want) {
		t.Errorf("additions differ from %s; rerun with -update and review the diff", golden)
	}
}

func TestAdditionsUnit(t *testing.T) {
	v := shelfValidator(t, systemRunner(t))
	tests := []struct {
		typ  string
		want []string
	}{
		{"Envelope", []string{"kind"}}, // through the rename table
		{"Fault", []string{"hint"}},    // through the rename table
		{"Record", []string{"gloss", "notes"}},
		{"Sleeve", []string{"sheen"}},
		{"Track", []string{"mood"}},
		{"Labelled", nil}, // no OpenAPI counterpart
	}
	for _, tt := range tests {
		t.Run(tt.typ, func(t *testing.T) {
			got := slices.Sorted(maps.Keys(v.schema.types[tt.typ].additions))
			if !slices.Equal(got, tt.want) {
				t.Errorf("additions = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExtensionChain proves a type carries what every level of its
// extension chain declares.
func TestExtensionChain(t *testing.T) {
	s, err := parseSchema(readFile(t, "testdata/shelf.xsd"))
	if err != nil {
		t.Fatal(err)
	}
	rec := s.types["Record"]
	for _, attr := range []string{"id", "label", "size"} {
		if !rec.attrs[attr] {
			t.Errorf("Record lacks attribute %s", attr)
		}
	}
	if rec.children["track"] != "Track" {
		t.Errorf("Record's track child has type %q, want Track", rec.children["track"])
	}
}

func TestRenameMissingSchema(t *testing.T) {
	path, err := filepath.Abs("testdata/shelf.xsd")
	if err != nil {
		t.Fatal(err)
	}
	renames := map[string]string{"Envelope": "EnvelopeBase", "Fault": "Grievance"}
	_, err = newValidator(readFile(t, path), path, readFile(t, "testdata/shelf.openapi.json"), renames, systemRunner(t))
	if err == nil {
		t.Fatal("newValidator succeeded; want an error naming the missing schema")
	}
	for _, want := range []string{"Grievance", "Fault"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("err = %v, want it to name %s", err, want)
		}
	}
}
