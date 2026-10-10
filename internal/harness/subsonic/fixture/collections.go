package fixture

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strconv"
)

// DefaultCollectionName names the collection a single-manifest run serves.
const DefaultCollectionName = "music"

// A Collection is one library a run serves: the name a recording gives it
// and the manifest it builds from.
type Collection struct {
	Name, Manifest string
}

// A BuiltCollection is a collection whose library has been loaded and
// built.
type BuiltCollection struct {
	Name string
	// Dir is the library root Build wrote.
	Dir     string
	Library *Library
}

// BuildCollections loads each collection's manifest and builds its library
// into its own directory under root, in list order. It returns the built
// collections and the digest of the whole list: the hex sha256 over each
// collection's quoted name and library digest, one tab-separated line
// apiece, in list order.
//
// A one-collection list's digest is that library's own digest, blind to
// the collection's name, so a single-library recording carries the digest
// [Library.Build] reports.
func BuildCollections(root string, cols []Collection) ([]BuiltCollection, string, error) {
	built := make([]BuiltCollection, 0, len(cols))
	digests := make([]namedDigest, 0, len(cols))
	for i, c := range cols {
		lib, err := Load(c.Manifest)
		if err != nil {
			return nil, "", fmt.Errorf("collection %q: %w", c.Name, err)
		}
		dir := filepath.Join(root, strconv.Itoa(i))
		digest, err := lib.Build(dir)
		if err != nil {
			return nil, "", fmt.Errorf("collection %q: %w", c.Name, err)
		}
		built = append(built, BuiltCollection{Name: c.Name, Dir: dir, Library: lib})
		digests = append(digests, namedDigest{name: c.Name, digest: digest})
	}
	return built, combineDigests(digests), nil
}

type namedDigest struct {
	name, digest string
}

func combineDigests(ds []namedDigest) string {
	if len(ds) == 1 {
		return ds[0].digest
	}
	h := sha256.New()
	for _, d := range ds {
		// hash.Hash.Write never returns an error.
		_, _ = fmt.Fprintf(h, "%q\t%s\n", d.name, d.digest)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
