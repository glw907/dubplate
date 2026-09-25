// Package web embeds the SvelteKit build's static output into the
// dubplate-server binary. `npm run build` writes a hashed `_app/`
// directory and a prerendered document per route into build/, plus a
// copy of web/static/placeholder.txt (adapter-static copies static/
// verbatim on every build); build/placeholder.txt is a second,
// separately committed copy of that same file, since adapter-static
// clears build/ on every run and a fresh checkout that has never run
// the web build would otherwise fail //go:embed at compile time.
package web

import "embed"

// Assets is build/'s content, rooted at "build" ("build/login.html",
// "build/_app/immutable/<hash>.js", ...). cmd/dubplate-server strips
// the "build" prefix via [io/fs.Sub] itself, rather than this package
// doing so, since that is the one place able to return an error a
// placeholder-only tree should produce instead of panicking at program
// init.
//
//go:embed all:build
var Assets embed.FS
