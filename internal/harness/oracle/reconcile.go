package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// versionProbe prints the installed version of every component a
// derived record's meaning depends on. unidecode's transliteration
// tables, jellyfish's edit distance, and lap's assignment solver all
// sit directly in the distance path (the mechanics study, Q1.4), so a
// silent bump there would change verdicts with nothing in the corpus
// to show for it.
const versionProbe = `import json, importlib.metadata as md
print(json.dumps({n: md.version(n) for n in ("beets", "unidecode", "jellyfish", "lap", "numpy")}))`

// Reconcile compares the manifest's pinned component versions against
// the venv the shim would actually run under, and reports every
// component that disagrees by name. An unpinned manifest field counts
// as a disagreement: the corpus records what beets ran with, and a
// blank is not a version.
//
// Reconcile is derivation-time work, called before a record is
// written, never from the repository gate: it needs the venv, which
// the gate does not have.
func Reconcile(ctx context.Context, m *cassette.Manifest, opts Options) error {
	installed, err := venvVersions(ctx, opts.Venv)
	if err != nil {
		return err
	}

	pinned := []struct {
		component string
		manifest  string
	}{
		{"beets", m.Beets},
		{"unidecode", m.Unidecode},
		{"jellyfish", m.Jellyfish},
		{"lap", m.Lap},
		{"numpy", m.Numpy},
	}

	var disagreements []string
	for _, p := range pinned {
		switch {
		case p.manifest == "":
			disagreements = append(disagreements, fmt.Sprintf("%s: manifest pins nothing, venv has %s", p.component, installed[p.component]))
		case p.manifest != installed[p.component]:
			disagreements = append(disagreements, fmt.Sprintf("%s: manifest pins %s, venv has %s", p.component, p.manifest, installed[p.component]))
		}
	}
	if len(disagreements) > 0 {
		return fmt.Errorf("oracle: reconcile: %s", strings.Join(disagreements, "; "))
	}
	return nil
}

// venvVersions asks the venv's own interpreter what it has installed,
// rather than parsing the requirements file, so an environment that
// drifted from the pins is caught instead of believed.
func venvVersions(ctx context.Context, venv string) (map[string]string, error) {
	cmd := exec.CommandContext(ctx, filepath.Join(venv, "bin", "python3"), "-c", versionProbe)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("oracle: reconcile: probe venv %s: %w: %s", venv, err, stderr.String())
	}

	versions := map[string]string{}
	if err := json.Unmarshal(stdout.Bytes(), &versions); err != nil {
		return nil, fmt.Errorf("oracle: reconcile: parse venv versions: %w", err)
	}
	return versions, nil
}
