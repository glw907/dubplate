package grade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/opensubsonic/conformance"
)

// openapiFile is the pinned OpenAPI document's name in the data directory.
const openapiFile = "openapi.json"

// pinnedDocSum is the sha256 of the committed OpenAPI document. A
// re-pin replaces the file, its provenance record, and this sum together.
const pinnedDocSum = "cb54c03c33835d132c555863e9771e30dfaa2930312853ca27dfece2ed46bfb6"

// loadContract reads the OpenAPI document in dir, verifies its sha256
// against sum, and returns a checker over it.
func loadContract(dir, sum string) (*conformance.Checker, error) {
	path := filepath.Join(dir, openapiFile)
	doc, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read openapi document: %w", err)
	}
	got := sha256.Sum256(doc)
	if hex.EncodeToString(got[:]) != sum {
		return nil, fmt.Errorf("openapi document %s has sha256 %x, want %s", path, got, sum)
	}
	c, err := conformance.New(doc)
	if err != nil {
		return nil, fmt.Errorf("openapi document %s: %w", path, err)
	}
	return c, nil
}
