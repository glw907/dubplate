// Package web embeds the uploader's compiled static assets into the
// dubplate binary. `npm run build` in web/ writes index.html and a
// hashed assets/ directory into dist/, plus a copy of
// web/public/placeholder.txt (Vite copies public/ verbatim on every
// build); dist/placeholder.txt is a second, separately committed copy
// of that same file so this package still builds on a fresh checkout
// that has never run the web build (a bare, empty dist/ fails
// //go:embed at compile time).
package web

import "embed"

// Assets is dist/'s content, rooted at "dist" ("dist/index.html",
// "dist/assets/<hash>.js", ...). cmd/dubplate/serve.go passes it to
// internal/serve, which strips the "dist" prefix via [io/fs.Sub] itself
// rather than this package doing so, since that is the one place able to
// return the error a missing directory would produce instead of panicking
// at program init.
//
//go:embed all:dist
var Assets embed.FS
