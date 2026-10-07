package fixture

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/engine/tagio"
)

// Build writes l's library under target, creating directories as
// needed. It returns the library digest: the hex sha256 over the sorted
// lines "path, size, sha256" of every file written, each field
// tab-separated.
func (l *Library) Build(target string) (string, error) {
	var written []string
	put := func(rel string, data []byte) (string, error) {
		full := filepath.Join(target, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return "", fmt.Errorf("create directory: %w", err)
		}
		if err := atomicfile.WriteFile(full, data, 0o644); err != nil {
			return "", fmt.Errorf("write %s: %w", rel, err)
		}
		written = append(written, rel)
		return full, nil
	}
	for _, a := range l.manifest.Albums {
		if a.Cover != nil {
			if _, err := put(a.Cover.Path, l.seeds[a.Cover.Seed]); err != nil {
				return "", err
			}
		}
		for _, t := range a.Tracks {
			full, err := put(t.Path, l.seeds[t.Seed])
			if err != nil {
				return "", err
			}
			var pictures []tagio.Picture
			if t.Picture != "" {
				pictures = []tagio.Picture{{Type: frontCover, Data: l.seeds[t.Picture]}}
			}
			if err := tagio.Write(full, t.Tags, pictures, 0o644); err != nil {
				return "", fmt.Errorf("tag %s: %w", t.Path, err)
			}
		}
	}

	slices.Sort(written)
	h := sha256.New()
	for _, rel := range written {
		b, err := os.ReadFile(filepath.Join(target, filepath.FromSlash(rel)))
		if err != nil {
			return "", fmt.Errorf("read back %s: %w", rel, err)
		}
		// hash.Hash.Write never returns an error.
		_, _ = fmt.Fprintf(h, "%s\t%d\t%x\n", rel, len(b), sha256.Sum256(b))
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
