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

// newStaticFiles returns an http.Handler serving GET and HEAD only, at
// exactly the request paths files holds, and answering
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
func newStaticFiles(files map[string]staticFile) http.Handler {
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

// newFontsRoute returns a [newStaticFiles] handler serving, as
// font/woff2, every .woff2 file fontsFS holds at its root, each at
// exactly routePrefix followed by a separator and its name. Every
// other path under it (routePrefix itself, a directory such as
// fontsFS's own "licenses/", a top-level file that is not a .woff2
// such as FONTS.md, a trailing slash, and a name fontsFS does not
// hold) answers [http.NotFoundHandler]'s 404. It refuses when fontsFS
// holds no .woff2 file, the same shape [isPlaceholderOnly] refuses on
// for the whole bundle.
func newFontsRoute(fontsFS fs.FS, routePrefix string) (http.Handler, error) {
	faces, err := readStaticDir(fontsFS, ".", ".woff2", "font/woff2", routePrefix)
	if err != nil {
		return nil, fmt.Errorf("webapp: fonts: %w", err)
	}
	if len(faces) == 0 {
		return nil, errors.New("no .woff2 file embedded under web/build/fonts/")
	}
	return newStaticFiles(faces), nil
}

// newIconsRoute returns a [newStaticFiles] handler serving the mark:
// the favicon in its two forms under prefix, the two files the
// documents' icon links name, and every .svg file under the build's
// mark directory under the mark's own route. It refuses when either
// favicon or every mark file is missing, since a document would then
// link a file the server cannot answer.
func newIconsRoute(build fs.FS, prefix string, names RouteNames) (http.Handler, error) {
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
	return newStaticFiles(files), nil
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
