// Package web embeds the uploader's compiled static assets into the
// dubplate binary. U9's Vite build writes index.html and a hashed
// /assets/ directory into dist/; until then, dist/placeholder.txt is the
// only file here, committed so this package builds before Vite ever
// runs (a bare, empty dist/ fails //go:embed at compile time).
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
