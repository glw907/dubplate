package serve

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"regexp"
	"slices"
	"strings"
)

// The states [DocumentCSP] refuses. Each is a build defect: a document
// that reaches a browser in any of them either never boots or runs a
// script the policy was meant to forbid.
var (
	errNoCSPMeta          = errors.New("no content-security-policy meta tag")
	errNoScriptSrc        = errors.New("the CSP meta has no script-src directive")
	errUnsafeInlineScript = errors.New("the CSP meta's script-src allows 'unsafe-inline'")
	errUnhashedScript     = errors.New("an inline script's hash is not in the CSP meta's script-src")
	errUnrecognizedScript = errors.New("a script tag is not the bare <script> the build emits")
)

// The directives every served document's policy carries beside
// script-src. frame-ancestors is among them because a meta policy
// cannot carry it, so only the header can refuse framing. img-src
// admits data: because DaisyUI draws its --noise texture as a data:
// SVG background image. The kit options in web/vite.config.ts write
// the same set into each document's meta, and a test in the serve verb
// fails when the two differ.
const (
	cspDefaultSrc = "default-src 'self'"
	cspStyleSrc   = "style-src 'self' 'unsafe-inline'"
	cspImgSrc     = "img-src 'self' data:"
	cspHardening  = "frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'"
)

var (
	metaTagRE       = regexp.MustCompile(`(?i)<meta\b[^>]*>`)
	httpEquivAttrRE = regexp.MustCompile(`(?i)\bhttp-equiv\s*=\s*"([^"]*)"`)
	contentAttrRE   = regexp.MustCompile(`(?i)\bcontent\s*=\s*"([^"]*)"`)

	// scriptOpenRE matches the start of any script opening tag, so a
	// tag carrying attributes is found and refused rather than missed.
	scriptOpenRE  = regexp.MustCompile(`(?i)<script`)
	scriptCloseRE = regexp.MustCompile(`(?i)</script`)
)

// DocumentCSP returns the Content-Security-Policy header value for doc,
// a document the SvelteKit build prerendered with hashes. The header's
// script-src carries 'self' and every sha256 source the document's own
// CSP meta lists, and the other directives are fixed.
//
// Before building it, DocumentCSP hashes every inline script in doc
// and refuses when one is missing from the meta's list, so a document
// whose bootstrap the policy would block never reaches a browser. It
// also refuses a document with no CSP meta, a meta with no script-src,
// a script-src allowing 'unsafe-inline', and any script tag other than
// a bare <script>, which is the only kind the build emits and the only
// kind this scan reads. name identifies the document in a returned
// error.
func DocumentCSP(name string, doc []byte) (string, error) {
	hashes, err := metaScriptHashes(doc)
	if err != nil {
		return "", fmt.Errorf("serve: document %q: %w", name, err)
	}
	if err := checkInlineScripts(doc, hashes); err != nil {
		return "", fmt.Errorf("serve: document %q: %w", name, err)
	}

	var scriptSrc strings.Builder
	scriptSrc.WriteString("script-src 'self'")
	for _, h := range hashes {
		scriptSrc.WriteString(" '" + h + "'")
	}
	return strings.Join([]string{cspDefaultSrc, scriptSrc.String(), cspStyleSrc, cspImgSrc, cspHardening}, "; "), nil
}

// metaScriptHashes returns the sha256 sources, unquoted, that the
// script-src of doc's CSP meta lists, in the meta's order. The meta's
// content attribute is HTML-escaped by the build, so it is unescaped
// before its directives are read.
func metaScriptHashes(doc []byte) ([]string, error) {
	var policy string
	found := false
	for _, tag := range metaTagRE.FindAll(doc, -1) {
		equiv := httpEquivAttrRE.FindSubmatch(tag)
		if equiv == nil || !strings.EqualFold(string(equiv[1]), "content-security-policy") {
			continue
		}
		if content := contentAttrRE.FindSubmatch(tag); content != nil {
			policy = html.UnescapeString(string(content[1]))
			found = true
			break
		}
	}
	if !found {
		return nil, errNoCSPMeta
	}

	for directive := range strings.SplitSeq(policy, ";") {
		fields := strings.Fields(directive)
		if len(fields) == 0 || !strings.EqualFold(fields[0], "script-src") {
			continue
		}
		sources := fields[1:]
		if slices.ContainsFunc(sources, func(s string) bool { return strings.EqualFold(s, "'unsafe-inline'") }) {
			return nil, errUnsafeInlineScript
		}
		var hashes []string
		for _, s := range sources {
			if h, ok := strings.CutPrefix(s, "'sha256-"); ok && strings.HasSuffix(h, "'") {
				hashes = append(hashes, "sha256-"+strings.TrimSuffix(h, "'"))
			}
		}
		return hashes, nil
	}
	return nil, errNoScriptSrc
}

// checkInlineScripts reports the first script in doc that is not a
// bare <script> or whose body's sha256 is not among hashes.
func checkInlineScripts(doc []byte, hashes []string) error {
	rest := doc
	for {
		loc := scriptOpenRE.FindIndex(rest)
		if loc == nil {
			return nil
		}
		if loc[1] >= len(rest) || rest[loc[1]] != '>' {
			return errUnrecognizedScript
		}
		body := rest[loc[1]+1:]
		end := scriptCloseRE.FindIndex(body)
		if end == nil {
			return fmt.Errorf("%w: it is never closed", errUnrecognizedScript)
		}
		sum := sha256.Sum256(body[:end[0]])
		hash := "sha256-" + base64.StdEncoding.EncodeToString(sum[:])
		if !slices.Contains(hashes, hash) {
			return fmt.Errorf("%w: %s", errUnhashedScript, hash)
		}
		rest = body[end[1]:]
	}
}
