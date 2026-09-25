package serve

import (
	"bytes"
	"fmt"
	"html"
	"io/fs"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ThemeCookieName is the theme preference cookie's name, following
// __Host-dubplate_session (internal/auth/session.go): its __Host-
// prefix is enforced by the browser rather than by this server, so the
// cookie is stored only when it carries Path=/, no Domain, and Secure.
// Only 5c-iii's toggle writes it; this package only reads it.
const ThemeCookieName = "__Host-dubplate_theme"

// RewriteDocument returns raw with the exactly-one occurrence of
// placeholder replaced by composed, the R4.5 rewrite's whole edit: no
// other byte, an inline <script> body included, changes. name
// identifies the document in the two states the rewrite refuses:
// placeholder absent, and placeholder present more than once. Either
// shape is a build defect a caller must refuse startup on, never guess
// past, since a second occurrence could sit inside a script body a
// blind replace would then corrupt.
func RewriteDocument(name string, raw []byte, placeholder, composed string) ([]byte, error) {
	needle := []byte(placeholder)
	switch n := bytes.Count(raw, needle); n {
	case 0:
		return nil, fmt.Errorf("serve: document %q: placeholder %q not found", name, placeholder)
	case 1:
		return bytes.Replace(raw, needle, []byte(composed), 1), nil
	default:
		return nil, fmt.Errorf("serve: document %q: placeholder %q appears %d times, want exactly 1", name, placeholder, n)
	}
}

// BuildDocumentVariants rewrites raw once per entry in palettes and
// returns one body per palette. The rewrite makes two edits and no
// other: it inserts one <link rel="stylesheet"> per entry of tierHrefs,
// in order, directly after the last stylesheet link the build wrote,
// and substitutes placeholder with match+"-"+palette (the composed
// theme name). Anchoring on the last link keeps every tier after all of
// the bundle's own CSS however many stylesheets a page links, so a
// tier declaration wins every tie of equal specificity by document
// order. A document with no stylesheet link is refused, naming it,
// since the tiers would then have nothing to follow.
//
// The variants are computed once, at startup; [NewDocument] never
// touches raw again, so no request byte can reach a served document.
// name identifies the document in a returned error.
func BuildDocumentVariants(name string, raw []byte, placeholder, match string, palettes, tierHrefs []string) (map[string][]byte, error) {
	linked, err := insertTierLinks(name, raw, tierHrefs)
	if err != nil {
		return nil, err
	}
	variants := make(map[string][]byte, len(palettes))
	for _, palette := range palettes {
		body, err := RewriteDocument(name, linked, placeholder, match+"-"+palette)
		if err != nil {
			return nil, err
		}
		variants[palette] = body
	}
	return variants, nil
}

// insertTierLinks returns raw with a stylesheet link per href inserted
// directly after the end of raw's last stylesheet link tag.
func insertTierLinks(name string, raw []byte, hrefs []string) ([]byte, error) {
	anchor := -1
	for _, loc := range linkTagRE.FindAllIndex(raw, -1) {
		if isStylesheetLink(raw[loc[0]:loc[1]]) {
			anchor = loc[1]
		}
	}
	if anchor < 0 {
		return nil, fmt.Errorf("serve: document %q: no stylesheet link to place the tier links after", name)
	}

	var links bytes.Buffer
	for _, href := range hrefs {
		fmt.Fprintf(&links, `<link href="%s" rel="stylesheet">`, html.EscapeString(href))
	}
	out := make([]byte, 0, len(raw)+links.Len())
	out = append(out, raw[:anchor]...)
	out = append(out, links.Bytes()...)
	return append(out, raw[anchor:]...), nil
}

// NewStylesheet returns an http.Handler serving GET and HEAD only,
// answering body as text/css. body is fixed when the handler is built,
// so the response reads nothing from the request. The response is
// no-cache, since the stylesheet's content changes at a restart while
// its name does not.
func NewStylesheet(body []byte) http.Handler {
	return methodGetHead(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(body)
	}))
}

// DocumentConfig is what [NewDocument] serves a document from.
type DocumentConfig struct {
	// Variants holds one rewritten body per stampable palette, as
	// [BuildDocumentVariants] returns them.
	Variants map[string][]byte

	// DefaultPalette names the variant a request without a usable theme
	// cookie is answered with. It must be a key of Variants.
	DefaultPalette string

	// CacheControl is the document's Cache-Control value.
	CacheControl string

	// CSP is the document's Content-Security-Policy value, as
	// [DocumentCSP] builds it.
	CSP string
}

// NewDocument returns an http.Handler serving GET and HEAD only,
// answering the variant named by the [ThemeCookieName] cookie when its
// value is an exact key of cfg.Variants, and the default variant
// otherwise. cfg.DefaultPalette must be a key of cfg.Variants: that
// pairing is the caller's own startup wiring, never a per-request
// condition, so a mismatch panics rather than serving an empty body.
//
// A served document carries cfg.CacheControl, cfg.CSP, and Vary:
// Cookie, since the variant depends on the theme cookie and a shared
// cache must not answer one visitor with another's. A 405 carries
// none of them.
//
// The handler writes no Set-Cookie, on any of these paths, and logs
// nothing about the cookie's value: every response is byte-identical
// to one of the bodies [BuildDocumentVariants] computed at startup, so
// no request byte reaches a served document.
func NewDocument(cfg DocumentConfig) http.Handler {
	if _, ok := cfg.Variants[cfg.DefaultPalette]; !ok {
		panic(fmt.Sprintf("serve: NewDocument: default palette %q has no variant", cfg.DefaultPalette))
	}
	return methodGetHead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		palette := cfg.DefaultPalette
		if cookie, err := r.Cookie(ThemeCookieName); err == nil {
			if _, ok := cfg.Variants[cookie.Value]; ok {
				palette = cookie.Value
			}
		}
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", cfg.CacheControl)
		h.Set("Content-Security-Policy", cfg.CSP)
		h.Set("Vary", "Cookie")
		_, _ = w.Write(cfg.Variants[palette])
	}))
}

// LogRequests wraps next, writing one record to logger for each
// request once next has answered it: the method, the status, and the
// path. The path is bounded as the shell's refusal record bounds it,
// and the query string is never logged, since a document's URL may
// carry a sign-in token there.
func LogRequests(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture := &statusCapture{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(capture, r)
		logger.Info("serve: document request",
			"method", r.Method,
			"path", boundedPath(r.URL.Path),
			"status", capture.status,
		)
	})
}

// NewAssets returns an http.Handler serving GET and HEAD only, looking
// up each request path as an exact file under root: routePrefix is the
// exact path a caller mounted root at (for example "<prefix>/_app"),
// and a request path is served only when it is routePrefix followed by
// a separator and a name root holds as a regular file. Every other
// shape -- routePrefix itself, a path under it naming a directory, and
// a path naming no file -- answers [Terminator]'s 404, with no
// directory listing and no redirect, unlike [http.FileServerFS]
// (R-5cii21).
//
// A file under root's immutable directory is served public and
// immutable for a year: the build names each one by its content's
// hash, so a name never changes content. Every other file keeps a
// stable name across builds and is served no-cache.
func NewAssets(root fs.FS, routePrefix string) http.Handler {
	terminator := Terminator()
	prefixSlash := routePrefix + "/"
	return methodGetHead(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, prefixSlash) {
			terminator.ServeHTTP(w, r)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, prefixSlash)
		data, err := fs.ReadFile(root, name)
		if err != nil {
			terminator.ServeHTTP(w, r)
			return
		}
		cacheControl := "no-cache"
		if strings.HasPrefix(name, "immutable/") {
			cacheControl = "public, max-age=31536000, immutable"
		}
		w.Header().Set("Cache-Control", cacheControl)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	}))
}

// methodGetHead answers 405 with Allow: GET, HEAD to any method but
// GET and HEAD, and hands everything else to next unchanged. HEAD
// needs no separate branch: net/http's own server discards a
// handler's written body on a HEAD request while still computing
// Content-Length, so next can always write the full body.
func methodGetHead(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// linkTagRE finds every <link ...> opening tag in a built document.
// hrefAttrRE and relAttrRE then search inside one matched tag's own
// bytes for its href and rel attributes independently, rather than
// assuming either attribute's order: the probe's built links carry
// href before rel (Task 0), and a regex anchored to one order would
// silently find nothing on the other.
var (
	linkTagRE  = regexp.MustCompile(`(?i)<link\b[^>]*>`)
	hrefAttrRE = regexp.MustCompile(`(?i)\bhref\s*=\s*"([^"]*)"`)
	relAttrRE  = regexp.MustCompile(`(?i)\brel\s*=\s*"([^"]*)"`)
)

// StylesheetHrefs returns the href of every <link rel="stylesheet">
// tag in doc, in document order. It reads the built bundle's own
// attribute-quoting shape (kit always double-quotes), the fixture
// state every caller of this function runs against.
func StylesheetHrefs(doc []byte) []string {
	var hrefs []string
	for _, tag := range linkTagRE.FindAll(doc, -1) {
		if !isStylesheetLink(tag) {
			continue
		}
		href := hrefAttrRE.FindSubmatch(tag)
		if href == nil {
			continue
		}
		hrefs = append(hrefs, string(href[1]))
	}
	return hrefs
}

// isStylesheetLink reports whether tag, one <link> opening tag, carries
// rel="stylesheet".
func isStylesheetLink(tag []byte) bool {
	rel := relAttrRE.FindSubmatch(tag)
	return rel != nil && string(rel[1]) == "stylesheet"
}

// dataThemeSelectorRE matches one [data-theme=<name>] compound
// selector, quoted or not, capturing the name alone: the closing "]"
// anchors the match, so a prefix of a real name (matched as a
// substring rather than the whole bracket expression) can never pass.
// It is applied to raw CSS text rather than a parsed selector tree, so
// a name inside an :is(...) list, which a minifier may produce
// (R-5cii12), is found the same as a standalone selector.
var dataThemeSelectorRE = regexp.MustCompile(`\[data-theme=(?:"([a-zA-Z0-9-]+)"|([a-zA-Z0-9-]+))\]`)

// ThemeSelectors returns the set of every name css declares a
// [data-theme=<name>] compound selector for, quoted and unquoted forms
// both.
func ThemeSelectors(css []byte) map[string]struct{} {
	names := make(map[string]struct{})
	for _, m := range dataThemeSelectorRE.FindAllSubmatch(css, -1) {
		name := m[1]
		if len(name) == 0 {
			name = m[2]
		}
		names[string(name)] = struct{}{}
	}
	return names
}
