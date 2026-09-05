package musicbrainz

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

// loadTestdata decodes testdata/name into T. The fixtures decode into
// four different wire types across this package's tests, which is why
// it is generic rather than one loader per type.
func loadTestdata[T any](t *testing.T, name string) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(readTestdata(t, name), &v); err != nil {
		t.Fatalf("decoding testdata/%s: %v", name, err)
	}
	return v
}
