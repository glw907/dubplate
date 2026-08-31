// Package reject defines the quality-gate rejection taxonomy: the reasons a
// staged file can fail the gate, their plain-language wording for a
// contributor's rejection email, and the rendered email itself. The wording
// is contract, shared verbatim with the uploader page and upload-check.
package reject

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Reason identifies why a staged file failed the quality gate. ReasonNone
// is the zero value, so a passing verdict carries no spurious reason.
type Reason int

const (
	ReasonNone Reason = iota
	ReasonBadExtension
	ReasonSymlink
	ReasonCorrupt
	ReasonBelowCDQuality
	ReasonTooLarge
	ReasonUnpackFailed
)

// Friendly returns the plain-language explanation shown to the contributor
// who uploaded the rejected file. It matches the shell layer's
// friendly_reason() wording verbatim, with one deliberate narrowing: the
// shell named the file's actual extension ("This is an MP3 file, ..."),
// which Reason can't reproduce since it carries no file path. The trailing
// clause that wording shares across every extension survives unchanged.
func (r Reason) Friendly() string {
	switch r {
	case ReasonBadExtension:
		return "This isn't a FLAC file, and the library keeps lossless FLAC."
	case ReasonSymlink:
		return "This looks like a symlink or shortcut rather than a real file, and the library only accepts real files."
	case ReasonCorrupt:
		return "This file didn't survive the upload intact. Uploading it again usually fixes this."
	case ReasonBelowCDQuality:
		return "This FLAC is below CD quality."
	case ReasonTooLarge:
		return "This upload is larger than the library allows."
	case ReasonUnpackFailed:
		return "This archive couldn't be unpacked."
	default:
		return ""
	}
}

// Rejection pairs one staged file with the reason it failed the quality
// gate.
type Rejection struct {
	Path   string
	Reason Reason
}

// RenderEmail fills the rejection-email template at templatePath and
// returns the rendered body. The template carries three literal
// placeholders -- "{first name}", "{album or file name}", and "{one reason
// line per file, from the mapping below}" -- which RenderEmail replaces
// with firstName, albumName, and one "<basename>: <friendly reason>" line
// per rejection, in the order given. Geoff owns the copy in the template;
// RenderEmail never rewords it.
func RenderEmail(templatePath, firstName, albumName string, rejections []Rejection) (string, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read reject email template %s: %w", templatePath, err)
	}

	lines := make([]string, len(rejections))
	for i, rj := range rejections {
		lines[i] = filepath.Base(rj.Path) + ": " + rj.Reason.Friendly()
	}

	body := string(data)
	body = strings.ReplaceAll(body, "{first name}", firstName)
	body = strings.ReplaceAll(body, "{album or file name}", albumName)
	body = strings.ReplaceAll(body, "{one reason line per file, from the mapping below}", strings.Join(lines, "\n"))
	return body, nil
}
