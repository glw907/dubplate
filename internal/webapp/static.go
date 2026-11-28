package webapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"path"
	"strings"
	"time"
)

// svgCSP is the policy every SVG response carries. A browser opening an
// SVG directly runs it as a document, script and all; the policy
// refuses every fetch and script and sandboxes it, keeping only the
// inline styles an SVG's own presentation may use.
const svgCSP = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// staticFile is one embedded file served at an exact path: its bytes,
// its Content-Type, and its strong ETag, all fixed at startup so no
// request byte reaches them.
type staticFile struct {
	data        []byte
	contentType string
	etag        string
}

// newStaticFile returns data as a staticFile of contentType, its ETag
// a strong SHA-256 hash of data.
func newStaticFile(data []byte, contentType string) staticFile {
	sum := sha256.Sum256(data)
	return staticFile{data: data, contentType: contentType, etag: `"` + hex.EncodeToString(sum[:]) + `"`}
}

// staticFilesHandler returns an http.Handler serving GET and HEAD
// only, at exactly the request paths files holds, and answering
// [http.NotFoundHandler]'s 404 on every other path, with no directory
// listing. Every file is served no-cache, since its name stays put
// across builds while its bytes may not.
//
// Content-Type is set on w before [http.ServeContent] runs, so no host
// mime table can override it (Go's built-in extension table carries no
// ".woff2" entry, and its fallback reads host files that differ
// between the workstation and the box). The ETag lets a revalidation
// answer 304: "If the caller has set w's ETag header formatted per RFC
// 7232, section 2.3, ServeContent uses it to handle requests using
// If-Match, If-None-Match, or If-Range" (go doc
// net/http.ServeContent), which holds even though every response
// carries a zero modtime.
func staticFilesHandler(files map[string]staticFile) http.Handler {
	notFound := http.NotFoundHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowGetHead(w, r) {
			return
		}
		file, ok := files[r.URL.Path]
		if !ok {
			notFound.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", file.contentType)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", file.etag)
		if file.contentType == "image/svg+xml" {
			w.Header().Set("Content-Security-Policy", svgCSP)
		}
		http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, bytes.NewReader(file.data))
	})
}

// fontsHandler returns a [staticFilesHandler] serving, as
// font/woff2, every .woff2 file fontsFS holds at its root, each at
// exactly routePrefix followed by a separator and its name. Every
// other path under it (routePrefix itself, a directory such as
// fontsFS's own "licenses/", a top-level file that is not a .woff2
// such as FONTS.md, a trailing slash, and a name fontsFS does not
// hold) answers [http.NotFoundHandler]'s 404. It refuses when fontsFS
// holds no .woff2 file, the same shape [isPlaceholderOnly] refuses on
// for the whole bundle.
func fontsHandler(fontsFS fs.FS, routePrefix string) (http.Handler, error) {
	faces, err := readStaticDir(fontsFS, ".", ".woff2", "font/woff2", routePrefix)
	if err != nil {
		return nil, fmt.Errorf("webapp: fonts: %w", err)
	}
	if len(faces) == 0 {
		return nil, errors.New("no .woff2 file embedded under web/build/fonts/")
	}
	return staticFilesHandler(faces), nil
}

// iconsHandler returns a [staticFilesHandler] serving the mark:
// the favicon in its two forms under prefix, the two files the
// documents' icon links name, and every .svg file under the build's
// mark directory under the mark's own route. It refuses when either
// favicon or every mark file is missing, since a document would then
// link a file the server cannot answer.
func iconsHandler(build fs.FS, prefix string, names RouteNames) (http.Handler, error) {
	files, err := readStaticDir(build, names.Mark, ".svg", "image/svg+xml", prefix+"/"+names.Mark)
	if err != nil {
		return nil, fmt.Errorf("webapp: mark: %w", err)
	}
	if len(files) == 0 {
		return nil, errors.New("no .svg file embedded under web/build/mark/")
	}
	for _, favicon := range []struct{ name, contentType string }{
		{names.FaviconSVG, "image/svg+xml"},
		{names.FaviconPNG, "image/png"},
	} {
		data, err := fs.ReadFile(build, favicon.name)
		if err != nil {
			return nil, fmt.Errorf("webapp: favicon: %w", err)
		}
		files[prefix+"/"+favicon.name] = newStaticFile(data, favicon.contentType)
	}
	return staticFilesHandler(files), nil
}

// assetsHandler returns an http.Handler serving GET and HEAD only,
// looking up each request path as an exact file under root:
// routePrefix is the exact path a caller mounted root at (for example
// "<prefix>/_app"), and a request path is served only when it is
// routePrefix followed by a separator and a name root holds as a
// regular file. Every other shape (routePrefix itself, a path under it naming a directory, and
// a path naming no file) answers [http.NotFoundHandler]'s 404, with no
// directory listing and no redirect, unlike [http.FileServerFS].
//
// A file under root's immutable directory is served public and
// immutable for a year: the build names each one by its content's
// hash, so a name never changes content. Every other file keeps a
// stable name across builds and is served no-cache.
func assetsHandler(root fs.FS, routePrefix string) http.Handler {
	notFound := http.NotFoundHandler()
	prefixSlash := routePrefix + "/"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowGetHead(w, r) {
			return
		}
		if !strings.HasPrefix(r.URL.Path, prefixSlash) {
			notFound.ServeHTTP(w, r)
			return
		}
		name := strings.TrimPrefix(r.URL.Path, prefixSlash)
		data, err := fs.ReadFile(root, name)
		if err != nil {
			notFound.ServeHTTP(w, r)
			return
		}
		cacheControl := "no-cache"
		if strings.HasPrefix(name, "immutable/") {
			cacheControl = "public, max-age=31536000, immutable"
		}
		w.Header().Set("Cache-Control", cacheControl)
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(data))
	})
}

// stylesheetHandler returns an http.Handler serving GET and HEAD only,
// answering body as text/css at exactly route and the shell's 404 at
// every other path under it. body is fixed when the handler is built,
// so the response reads nothing from the request. The response is
// no-cache, since the stylesheet's content changes at a restart while
// its name does not.
func stylesheetHandler(route string, body []byte) http.Handler {
	serve := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !allowGetHead(w, r) {
			return
		}
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = w.Write(body)
	})
	return exactPaths(map[string]http.Handler{route: serve}, http.NotFoundHandler())
}

// allowGetHead reports whether r's method is GET or HEAD, and answers
// 405 with Allow: GET, HEAD when it is not, so the caller has only to
// return. HEAD needs no separate branch: net/http's own server discards
// a handler's written body on a HEAD request while still computing
// Content-Length, so a caller can always write the full body.
func allowGetHead(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	w.Header().Set("Allow", "GET, HEAD")
	http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
	return false
}

// readStaticDir returns every regular file directly under dir in fsys
// whose name ends in suffix, as a staticFile of contentType keyed by
// routePrefix followed by a separator and its name. Subdirectories are
// skipped.
func readStaticDir(fsys fs.FS, dir, suffix, contentType, routePrefix string) (map[string]staticFile, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	files := make(map[string]staticFile, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("%q: %w", entry.Name(), err)
		}
		files[routePrefix+"/"+entry.Name()] = newStaticFile(data, contentType)
	}
	return files, nil
}
