// Package reject defines the quality-gate rejection taxonomy: the reasons a
// staged file can fail the gate, their plain-language wording for a
// contributor's rejection email, and the rendered email itself. The wording
// is contract, shared verbatim with the uploader page and upload-check.
package reject

import (
	"encoding/json"
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
	// ReasonUnsafePath marks a ZIP central-directory entry internal/intake's
	// CheckZIP refuses to ever extract: a symlink member, an absolute path,
	// or a path whose ".." components would resolve outside the album
	// directory. Authorized for this check alone.
	ReasonUnsafePath
	// ReasonNoSpace marks an upload session internal/serve's admission
	// control refused because accepting it would cross the staging
	// filesystem's free-space floor. Authorized for that check alone.
	ReasonNoSpace
	// ReasonTooManySessions marks an upload session internal/serve's
	// admission control refused because its contributor already has
	// MaxSessionsPerUser sessions in flight. Authorized for that check
	// alone.
	ReasonTooManySessions
)

// Friendly returns the plain-language explanation for r, matching the shell
// layer's friendly_reason() wording verbatim. Reason alone carries no file
// path, so the bad-extension case falls back to a generic phrasing here;
// RenderEmail names the actual extension instead, via friendlyLine.
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
		return "This archive wouldn't unpack as a ZIP."
	case ReasonUnsafePath:
		return "This archive contains a file path the library can't safely use."
	case ReasonNoSpace:
		return "The library doesn't have enough free space to accept this upload right now."
	case ReasonTooManySessions:
		return "You already have the maximum number of uploads in progress."
	default:
		return ""
	}
}

// NextStep returns the plain-language action to take for r, rendered
// alongside Friendly() on both the uploader page and in a rejection email
// so a contributor never sees an explanation without a way forward.
func (r Reason) NextStep() string {
	switch r {
	case ReasonBadExtension:
		return "Convert it to FLAC and upload it again."
	case ReasonSymlink:
		return "Replace it with the real file and upload it again."
	case ReasonCorrupt:
		return "Upload it again from the original file."
	case ReasonBelowCDQuality:
		return "Get a CD-quality or better FLAC and upload that instead."
	case ReasonTooLarge:
		return "Split it into smaller uploads, or ask the admin to add it by hand."
	case ReasonUnpackFailed:
		return "Repack it as a standard ZIP file and upload it again."
	case ReasonUnsafePath:
		return "Repack the album as a plain ZIP with real files only, then upload it again."
	case ReasonNoSpace:
		return "Try again in a little while, or let the admin know if this keeps happening."
	case ReasonTooManySessions:
		return "Wait for one of your other uploads to finish, then try again."
	default:
		return ""
	}
}

// Label returns r's one-to-three-word chip string, rendered on the
// uploader page's per-track rows without a client-side lookup table
// (passb-copy-final.md section 4 is the wording authority).
func (r Reason) Label() string {
	switch r {
	case ReasonBadExtension:
		return "not FLAC"
	case ReasonSymlink:
		return "shortcut file"
	case ReasonCorrupt:
		return "damaged in transit"
	case ReasonBelowCDQuality:
		return "below CD quality"
	case ReasonTooLarge:
		return "too big"
	case ReasonUnpackFailed:
		return "wouldn't unpack"
	case ReasonUnsafePath:
		return "unsafe file path"
	case ReasonNoSpace:
		return "library is full"
	case ReasonTooManySessions:
		return "too many uploads"
	default:
		return ""
	}
}

// code returns r's stable wire identifier: lowercase, hyphenated, and
// independent of Reason's underlying int value, so a client never has to
// track this package's enum ordering.
func (r Reason) code() string {
	switch r {
	case ReasonBadExtension:
		return "bad-extension"
	case ReasonSymlink:
		return "symlink"
	case ReasonCorrupt:
		return "corrupt"
	case ReasonBelowCDQuality:
		return "below-cd-quality"
	case ReasonTooLarge:
		return "too-large"
	case ReasonUnpackFailed:
		return "unpack-failed"
	case ReasonUnsafePath:
		return "unsafe-path"
	case ReasonNoSpace:
		return "no-space"
	case ReasonTooManySessions:
		return "too-many-sessions"
	default:
		return "none"
	}
}

// codeToReason inverts Reason.code, so [Reason.UnmarshalJSON] can decode a
// wire code back to its enum value without hardcoding the table twice.
var codeToReason = map[string]Reason{
	ReasonNone.code():            ReasonNone,
	ReasonBadExtension.code():    ReasonBadExtension,
	ReasonSymlink.code():         ReasonSymlink,
	ReasonCorrupt.code():         ReasonCorrupt,
	ReasonBelowCDQuality.code():  ReasonBelowCDQuality,
	ReasonTooLarge.code():        ReasonTooLarge,
	ReasonUnpackFailed.code():    ReasonUnpackFailed,
	ReasonUnsafePath.code():      ReasonUnsafePath,
	ReasonNoSpace.code():         ReasonNoSpace,
	ReasonTooManySessions.code(): ReasonTooManySessions,
}

// reasonWire is the JSON shape [Reason.MarshalJSON] emits and
// [Reason.UnmarshalJSON] reads: a stable code plus the rendered copy
// strings, so a decoder never needs this package's Go-side enum or wording
// table to display a rejection.
type reasonWire struct {
	Code     string `json:"code"`
	Friendly string `json:"friendly"`
	NextStep string `json:"next_step"`
	Label    string `json:"label"`
}

// MarshalJSON encodes r as {"code","friendly","next_step","label"} rather
// than its underlying int, so the wire format survives a future reordering
// of the Reason constants and a client never has to special-case a bare
// integer.
func (r Reason) MarshalJSON() ([]byte, error) {
	return json.Marshal(reasonWire{Code: r.code(), Friendly: r.Friendly(), NextStep: r.NextStep(), Label: r.Label()})
}

// UnmarshalJSON decodes r from the {"code","friendly","next_step","label"}
// shape [Reason.MarshalJSON] emits. Only "code" is trusted; the copy
// strings are always recomputed from r's own tables, never taken off the
// wire, so a decoder can't be made to display wording this package didn't
// write.
func (r *Reason) UnmarshalJSON(data []byte) error {
	var wire reasonWire
	if err := json.Unmarshal(data, &wire); err != nil {
		return fmt.Errorf("reject: unmarshal reason: %w", err)
	}
	reason, ok := codeToReason[wire.Code]
	if !ok {
		return fmt.Errorf("reject: unknown reason code %q", wire.Code)
	}
	*r = reason
	return nil
}

// Rejection pairs one staged file with the reason it failed the quality
// gate.
type Rejection struct {
	Path   string
	Reason Reason
}

// extName maps a lowercase file extension, without its leading dot, to the
// label the bad-extension rejection line names it by. It reproduces the
// shell layer's friendly_reason() name table verbatim; an extension outside
// the table renders uppercased.
func extName(ext string) string {
	switch strings.ToLower(ext) {
	case "mp3":
		return "MP3"
	case "m4a":
		return "AAC"
	case "ogg":
		return "Ogg Vorbis"
	case "opus":
		return "Opus"
	case "wav":
		return "WAV"
	case "aiff":
		return "AIFF"
	default:
		return strings.ToUpper(ext)
	}
}

// friendlyLine returns the rejection-email line for rj. For
// ReasonBadExtension it names the file's actual extension, taken from
// rj.Path, since Friendly's Reason-only signature can't carry it; every
// other reason renders its Friendly() wording unchanged.
func friendlyLine(rj Rejection) string {
	if rj.Reason != ReasonBadExtension {
		return rj.Reason.Friendly()
	}
	ext := strings.TrimPrefix(filepath.Ext(rj.Path), ".")
	return "This is an " + extName(ext) + " file, and the library keeps lossless FLAC."
}

// RenderEmail fills the rejection-email template at templatePath and
// returns the rendered body. The template carries three literal
// placeholders -- "{first name}", "{album or file name}", and "{one reason
// line per file, from the mapping below}" -- which RenderEmail replaces
// with firstName, albumName, and one "<basename>: <friendly reason> <next
// step>" line per rejection, in the order given, so the email carries the
// same Friendly and NextStep wording the API JSON does. Geoff owns the
// copy in the template; RenderEmail never rewords it.
func RenderEmail(templatePath, firstName, albumName string, rejections []Rejection) (string, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("read reject email template %s: %w", templatePath, err)
	}

	lines := make([]string, len(rejections))
	for i, rj := range rejections {
		lines[i] = filepath.Base(rj.Path) + ": " + friendlyLine(rj) + " " + rj.Reason.NextStep()
	}

	body := string(data)
	body = strings.ReplaceAll(body, "{first name}", firstName)
	body = strings.ReplaceAll(body, "{album or file name}", albumName)
	body = strings.ReplaceAll(body, "{one reason line per file, from the mapping below}", strings.Join(lines, "\n"))
	return body, nil
}
