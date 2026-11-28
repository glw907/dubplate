// Package web embeds the SvelteKit build's static output into the
// dubplate-server binary. A copy of static/placeholder.txt is committed
// under build/, so a checkout that has never run `npm run build` still
// compiles.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:build
var embedded embed.FS

// Build is the build's output, rooted at the build directory:
// "login.html", "_app/immutable/<hash>.js", and so on.
var Build = func() fs.FS {
	sub, err := fs.Sub(embedded, "build")
	if err != nil {
		panic(err)
	}
	return sub
}()
