package grade

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glw907/dubplate/internal/harness/subsonic/xmlrule"
	"github.com/glw907/dubplate/internal/opensubsonic/conformance"
)

// openapiFile is the pinned OpenAPI document's name in the data directory.
const openapiFile = "openapi.json"

// pinnedDocSum is the sha256 of the committed OpenAPI document. A
// re-pin replaces the file, its provenance record, and this sum together.
const pinnedDocSum = "cb54c03c33835d132c555863e9771e30dfaa2930312853ca27dfece2ed46bfb6"

// contract is the pinned OpenAPI document's two checkers: the JSON
// checker, and the XML tier, which scopes the Subsonic XSD by the same
// document.
type contract struct {
	json *conformance.Checker
	xml  *xmlrule.Validator
}

// loadContract reads the OpenAPI document in dir, verifies its sha256
// against sum, and returns both checkers over it. The XML tier reads the
// cached XSD and needs xmllint at its pinned version.
func loadContract(dir, sum string) (*contract, error) {
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
	x, err := xmlrule.New(doc)
	if err != nil {
		return nil, fmt.Errorf("openapi document %s: %w", path, err)
	}
	return &contract{json: c, xml: x}, nil
}
