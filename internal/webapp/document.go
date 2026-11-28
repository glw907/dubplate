package webapp

import (
	"bytes"
	"fmt"
	"html"
	"log/slog"
	"net/http"
	"regexp"
	"slices"
	"strings"
)

// themeCookieName is the theme preference cookie's name, following
// __Host-dubplate_session (internal/auth/session.go): its __Host-
// prefix is enforced by the browser rather than by this server, so the
// cookie is stored only when it carries Path=/, no Domain, and Secure.
// The web app's theme toggle writes it; this package only reads it.
const themeCookieName = "__Host-dubplate_theme"

// themePlaceholder is the token app.html carries on <html data-theme>,
// substituted once per stampable palette at startup. It must match
// web/src/app.html's own literal exactly.
const themePlaceholder = "__DUBPLATE_THEME__"

// authRobots is the robots meta the two sign-in documents carry, so a
// search engine neither lists a sign-in page nor follows a link out of
// one.
const authRobots = "noindex, nofollow"

// document is one document the routes serve: its path inside the build
// tree, the path under the owned prefix it answers at, its
// Cache-Control value, and whether it carries the robots meta. raw and
// csp are empty in the [documents] table; [New] fills them with the
// build's bytes for the document and the CSP header built from them.
type document struct {
	name         string
	route        string
	cacheControl string
	noIndex      bool
	raw          []byte
	csp          string
}

// documents returns every document the build emits and the server
// serves. A document is no-cache by default: the stamp and the tier
// links are fixed at startup, so a cache may keep it and revalidate.
// The consumption document is no-store because its URL carried a
// sign-in token. The members and server settings documents are
// no-store because they answer only behind the admin gate, as the API
// they read from does.
func documents(names RouteNames) []document {
	return []document{
		{name: "login.html", route: "/" + names.Login, cacheControl: "no-cache", noIndex: true},
		{name: "login/link.html", route: "/" + names.Login + "/link", cacheControl: "no-store", noIndex: true},
		{name: "members.html", route: "/" + names.Members, cacheControl: "no-store"},
		{name: "server-settings.html", route: "/" + names.ServerSettings, cacheControl: "no-store"},
	}
}

// rewriteDocument returns raw with the exactly-one occurrence of
// placeholder replaced by composed, the theme stamp's whole edit: no
// other byte, an inline <script> body included, changes. name
// identifies the document in the two states the rewrite refuses:
// placeholder absent, and placeholder present more than once. Either
// shape is a build defect a caller must refuse startup on, never guess
// past, since a second occurrence could sit inside a script body a
// blind replace would then corrupt.
func rewriteDocument(name string, raw []byte, placeholder, composed string) ([]byte, error) {
	needle := []byte(placeholder)
	switch n := bytes.Count(raw, needle); n {
	case 0:
		return nil, fmt.Errorf("webapp: document %q: placeholder %q not found", name, placeholder)
	case 1:
		return bytes.Replace(raw, needle, []byte(composed), 1), nil
	default:
		return nil, fmt.Errorf("webapp: document %q: placeholder %q appears %d times, want exactly 1", name, placeholder, n)
	}
}

// variantConfig is what [buildDocumentVariants] rewrites a document
// with. Every attribute value it carries is escaped with
// html.EscapeString as it is written, so a quote, an angle bracket, or
// an ampersand in an operator's value cannot leave its attribute.
type variantConfig struct {
	// placeholder is the token the build left on <html data-theme>.
	placeholder string

	// match names the theme group; each palette's composed theme name
	// is match+"-"+palette.
	match string

	// palettes names one variant each.
	palettes []string

	// tierHrefs are the stylesheet links that follow the build's own,
	// in order.
	tierHrefs []string

	// siteName is the content of the dubplate-site-name meta every
	// variant carries.
	siteName string

	// robots is the content of the robots meta every variant carries,
	// or "" for none.
	robots string

	// themeColors holds each palette's theme-color metas. A palette
	// with no entry carries none.
	themeColors map[string][]themeColor

	// icons are the icon links every variant carries, in order.
	icons []icon
}

// icon is one <link rel="icon">: its href, its MIME type, and its
// sizes attribute, "" to omit it.
type icon struct {
	href        string
	contentType string
	sizes       string
}

// buildDocumentVariants rewrites raw once per entry in cfg.palettes and
// returns one body per palette. The rewrite makes two edits and no
// other. It substitutes cfg.placeholder with the composed theme name,
// and it inserts the head block directly after the last stylesheet
// link the build wrote: one <link rel="stylesheet"> per tier href, the
// site-name meta, the robots meta when cfg.robots is set, the
// palette's own theme-color metas, and the icon links, in that order.
// Anchoring on the last link keeps every tier after all of the
// bundle's own CSS however many stylesheets a page links, so a tier
// declaration wins every tie of equal specificity by document order,
// and it keeps every insertion in the head and outside every script
// body. A document with no stylesheet link, or whose last one sits
// outside <head>, is refused, naming it, since the tiers would then
// have nothing to follow or would land in the body.
//
// The placeholder is substituted in the build's own bytes before the
// head block goes in, so no operator value (the site name, a prefix in
// an href) can supply a missing placeholder or add a second one; an
// inserted value carrying the placeholder is refused outright.
//
// The variants are computed once, at startup; [documentHandler] never
// touches raw again, so no request byte can reach a served document.
// name identifies the document in a returned error.
func buildDocumentVariants(name string, raw []byte, cfg variantConfig) (map[string][]byte, error) {
	var shared bytes.Buffer
	for _, href := range cfg.tierHrefs {
		fmt.Fprintf(&shared, `<link href="%s" rel="stylesheet">`, html.EscapeString(href))
	}
	fmt.Fprintf(&shared, `<meta name="dubplate-site-name" content="%s">`, html.EscapeString(cfg.siteName))
	if cfg.robots != "" {
		fmt.Fprintf(&shared, `<meta name="robots" content="%s">`, html.EscapeString(cfg.robots))
	}
	var icons bytes.Buffer
	for _, icon := range cfg.icons {
		fmt.Fprintf(&icons, `<link rel="icon" href="%s" type="%s"`, html.EscapeString(icon.href), html.EscapeString(icon.contentType))
		if icon.sizes != "" {
			fmt.Fprintf(&icons, ` sizes="%s"`, html.EscapeString(icon.sizes))
		}
		icons.WriteString(">")
	}

	variants := make(map[string][]byte, len(cfg.palettes))
	for _, palette := range cfg.palettes {
		var head bytes.Buffer
		head.Write(shared.Bytes())
		for _, tc := range cfg.themeColors[palette] {
			head.WriteString(`<meta name="theme-color"`)
			if tc.scheme != "" {
				fmt.Fprintf(&head, ` media="(prefers-color-scheme: %s)"`, html.EscapeString(tc.scheme))
			}
			fmt.Fprintf(&head, ` content="%s">`, html.EscapeString(tc.color))
		}
		head.Write(icons.Bytes())

		stamped, err := rewriteDocument(name, raw, cfg.placeholder, cfg.match+"-"+palette)
		if err != nil {
			return nil, err
		}
		if bytes.Contains(head.Bytes(), []byte(cfg.placeholder)) {
			return nil, fmt.Errorf("webapp: document %q: an inserted head value carries the theme placeholder %q", name, cfg.placeholder)
		}
		anchor, err := headAnchor(name, stamped)
		if err != nil {
			return nil, err
		}
		doc := make([]byte, 0, len(stamped)+head.Len())
		doc = append(doc, stamped[:anchor]...)
		doc = append(doc, head.Bytes()...)
		doc = append(doc, stamped[anchor:]...)
		variants[palette] = doc
	}
	return variants, nil
}

// headAnchor returns the offset just past the end of raw's last
// stylesheet link tag, which must precede raw's first </head>.
func headAnchor(name string, raw []byte) (int, error) {
	anchor := -1
	for _, loc := range linkTagRE.FindAllIndex(raw, -1) {
		if isStylesheetLink(raw[loc[0]:loc[1]]) {
			anchor = loc[1]
		}
	}
	if anchor < 0 {
		return 0, fmt.Errorf("webapp: document %q: no stylesheet link to place the tier links after", name)
	}
	if end := bytes.Index(raw, []byte("</head>")); end < anchor {
		return 0, fmt.Errorf("webapp: document %q: the last stylesheet link sits outside <head>", name)
	}
	return anchor, nil
}

// tierHrefs returns the tier links for the document at docPath inside
// the build tree: the branding stylesheet always, and the operator
// stylesheet when one is configured. Each href is relative to the
// document's own URL, as the build's own links are, so the same
// document resolves them under any prefix.
func tierHrefs(docPath string, names RouteNames, hasOperatorCSS bool) []string {
	root := "./"
	if depth := strings.Count(docPath, "/"); depth > 0 {
		root = strings.Repeat("../", depth)
	}
	hrefs := []string{root + names.Branding}
	if hasOperatorCSS {
		hrefs = append(hrefs, root+names.OperatorCSS)
	}
	return hrefs
}

// linkTagRE finds every <link ...> opening tag in a built document.
// hrefAttrRE and relAttrRE then search inside one matched tag's own
// bytes for its href and rel attributes independently, rather than
// assuming either attribute's order: the build writes href before rel,
// and a regex anchored to one order would silently find nothing on the
// other.
var (
	linkTagRE  = regexp.MustCompile(`(?i)<link\b[^>]*>`)
	hrefAttrRE = regexp.MustCompile(`(?i)\bhref\s*=\s*"([^"]*)"`)
	relAttrRE  = regexp.MustCompile(`(?i)\brel\s*=\s*"([^"]*)"`)
)

// stylesheetHrefs returns the href of every <link rel="stylesheet">
// tag in doc, in document order. It reads the built bundle's own
// attribute-quoting shape (kit always double-quotes), the fixture
// state every caller of this function runs against.
func stylesheetHrefs(doc []byte) []string {
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

// documentConfig is what [documentHandler] serves a document from.
type documentConfig struct {
	// variants holds one rewritten body per stampable palette, as
	// [buildDocumentVariants] returns them.
	variants map[string][]byte

	// defaultPalette names the variant a request without a usable theme
	// cookie is answered with. It must be a key of variants.
	defaultPalette string

	// cacheControl is the document's Cache-Control value.
	cacheControl string

	// csp is the document's Content-Security-Policy value, as
	// [documentCSP] builds it.
	csp string

	// logger takes the record [logDocumentRequest] writes for each
	// answer.
	logger *slog.Logger
}

// documentHandler returns an http.Handler serving GET and HEAD only,
// answering the variant named by the [themeCookieName] cookie when its
// value is an exact key of cfg.variants, and the default variant
// otherwise. cfg.defaultPalette must be a key of cfg.variants: that
// pairing is the caller's own startup wiring, never a per-request
// condition, so a mismatch panics rather than serving an empty body.
//
// A served document carries cfg.cacheControl, cfg.csp, and Vary:
// Cookie, since the variant depends on the theme cookie and a shared
// cache must not answer one visitor with another's. A 405 carries
// none of them. Each answer, the 405 included, writes one record to
// cfg.logger.
//
// The handler writes no Set-Cookie, on any of these paths, and logs
// nothing about the cookie's value: every response is byte-identical
// to one of the bodies [buildDocumentVariants] computed at startup, so
// no request byte reaches a served document.
func documentHandler(cfg documentConfig) http.Handler {
	if _, ok := cfg.variants[cfg.defaultPalette]; !ok {
		panic(fmt.Sprintf("webapp: documentHandler: default palette %q has no variant", cfg.defaultPalette))
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowGetHead(w, r) {
			logDocumentRequest(cfg.logger, r, http.StatusMethodNotAllowed)
			return
		}
		palette := cfg.defaultPalette
		if cookie, err := r.Cookie(themeCookieName); err == nil {
			if _, ok := cfg.variants[cookie.Value]; ok {
				palette = cookie.Value
			}
		}
		h := w.Header()
		h.Set("Content-Type", "text/html; charset=utf-8")
		h.Set("Cache-Control", cfg.cacheControl)
		h.Set("Content-Security-Policy", cfg.csp)
		h.Set("Vary", "Cookie")
		_, _ = w.Write(cfg.variants[palette])
		logDocumentRequest(cfg.logger, r, http.StatusOK)
	})
}

// documentNotFoundHandler returns the 404 a documents' router answers
// a path it holds no document for with, logged as [documentHandler]
// logs its own answers.
func documentNotFoundHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
		logDocumentRequest(logger, r, http.StatusNotFound)
	})
}

// logDocumentRequest writes the one record a document route writes for
// each answer: the method, the status, and the path. The path is
// bounded, and the query string is never logged, since a document's
// URL may carry a sign-in token there. The method is logged only when
// it is one net/http names and as "other" otherwise, since a client may
// send any token as a method, up to the header size limit.
func logDocumentRequest(logger *slog.Logger, r *http.Request, status int) {
	logger.Info("webapp: document request",
		"method", loggedMethod(r.Method),
		"path", boundedPath(r.URL.Path),
		"status", status,
	)
}

// knownMethods is every method net/http names a constant for.
var knownMethods = []string{
	http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch,
	http.MethodDelete, http.MethodConnect, http.MethodOptions, http.MethodTrace,
}

// loggedMethod returns method when it is one of knownMethods and
// "other" otherwise.
func loggedMethod(method string) string {
	if slices.Contains(knownMethods, method) {
		return method
	}
	return "other"
}

// maxLoggedPathBytes bounds the request path a document record
// carries, matching the bound internal/serve puts on its refusal
// record: a path is bounded by nothing but http.Server.MaxHeaderBytes,
// and 256 bytes is enough to identify which page was asked for.
const maxLoggedPathBytes = 256

// boundedPath returns path unchanged when it fits maxLoggedPathBytes,
// and otherwise its first maxLoggedPathBytes bytes with a marker naming
// the truncation and the full length.
func boundedPath(path string) string {
	if len(path) <= maxLoggedPathBytes {
		return path
	}
	return fmt.Sprintf("%s...[truncated, %d bytes]", path[:maxLoggedPathBytes], len(path))
}
