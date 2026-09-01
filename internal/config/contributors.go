package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// contributor is one entry in contributors.yaml.
type contributor struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

// contributorsFile is contributors.yaml's top-level shape.
type contributorsFile struct {
	Contributors []contributor `yaml:"contributors"`
}

// Roster maps contributor names, as they appear as inbox directory names,
// to their contributors.yaml entry. The zero Roster is empty and safe to
// query: every lookup method reports absent rather than panicking, so a
// missing contributors.yaml degrades to "nobody on file", never an error.
type Roster struct {
	byName map[string]contributor
}

// LoadContributors parses the contributors.yaml at path. A missing file
// returns an empty [Roster] and a nil error: the retired shell layer's
// load_contributors_json (scripts/music-import) treated "no file" the same
// as "no match", and every caller here preserves that.
func LoadContributors(path string) (Roster, error) {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Roster{}, nil
	}
	if err != nil {
		return Roster{}, fmt.Errorf("read contributors %s: %w", path, err)
	}

	var cf contributorsFile
	if err := yaml.Unmarshal(data, &cf); err != nil {
		return Roster{}, fmt.Errorf("parse contributors %s: %w", path, err)
	}

	byName := make(map[string]contributor, len(cf.Contributors))
	for _, c := range cf.Contributors {
		byName[c.Name] = c
	}
	return Roster{byName: byName}, nil
}

// EmailFor reports the email address on file for name, and whether name
// matched a contributor.
func (r Roster) EmailFor(name string) (string, bool) {
	c, ok := r.byName[name]
	if !ok {
		return "", false
	}
	return c.Email, true
}

// FirstName reports the first word of the roster-confirmed name matching
// name, for the quality-gate rejection email's greeting. The inbox
// directory name IS the matched contributor's full name (contributors.yaml
// is matched on that same "name" field), so this confirms membership and
// then takes the first word of name itself.
func (r Roster) FirstName(name string) (string, bool) {
	if !r.Has(name) {
		return "", false
	}
	first, _, _ := strings.Cut(name, " ")
	return first, true
}

// Has reports whether name matches a contributor.
func (r Roster) Has(name string) bool {
	_, ok := r.byName[name]
	return ok
}
