package web

import (
	"io/fs"
	"os"
	"strings"
	"testing"
)

// colorSchemeMeta is the literal meta tag web/src/app.html carries
// (CSS Color Adjustment 1, "the color-scheme <meta>"): its exact
// rendered bytes, since kit emits app.html's own markup outside
// %sveltekit.head% unchanged.
const colorSchemeMeta = `<meta name="color-scheme" content="light dark" />`

// TestNoFlashStaticHalf is R-5cii22's static half: every prerendered
// document carries exactly one color-scheme meta before its first
// stylesheet link, and no stylesheet link carries a media or disabled
// attribute that could defer it past first paint. The rendered half (a
// first-paint sample under each emulated preference) is 5c-iii's.
func TestNoFlashStaticHalf(t *testing.T) {
	buildFS, err := fs.Sub(Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}

	for _, name := range []string{"login.html", "login/link.html"} {
		t.Run(name, func(t *testing.T) {
			data, err := fs.ReadFile(buildFS, name)
			if err != nil {
				t.Fatalf("fs.ReadFile: %v", err)
			}
			body := string(data)

			if n := strings.Count(body, colorSchemeMeta); n != 1 {
				t.Fatalf("%s carries %d color-scheme meta(s), want exactly 1", name, n)
			}
			metaIdx := strings.Index(body, colorSchemeMeta)

			linkIdx := strings.Index(body, `rel="stylesheet"`)
			if linkIdx == -1 {
				t.Fatalf("%s carries no stylesheet link", name)
			}
			if linkIdx < metaIdx {
				t.Errorf("%s's first stylesheet link precedes the color-scheme meta", name)
			}

			for _, tag := range linkTagsWithHref(body) {
				if !strings.Contains(tag, `rel="stylesheet"`) {
					continue
				}
				if strings.Contains(tag, " media=") || strings.Contains(tag, " disabled") {
					t.Errorf("%s's stylesheet tag %q carries media or disabled", name, tag)
				}
			}
		})
	}
}

// linkTagsWithHref returns every <link ...> opening tag in body, a
// small scanner independent of internal/serve's own extractor so
// neither can share a blind spot with the other.
func linkTagsWithHref(body string) []string {
	var tags []string
	for {
		start := strings.Index(body, "<link ")
		if start == -1 {
			return tags
		}
		end := strings.Index(body[start:], ">")
		if end == -1 {
			return tags
		}
		tags = append(tags, body[start:start+end+1])
		body = body[start+end+1:]
	}
}

// TestDaisyUIRootBackgroundRule quotes DaisyUI's own root-color rule
// from the installed package at the pin, rather than a literal this
// file restates: a version bump that changed the declaration would
// otherwise go unnoticed here.
func TestDaisyUIRootBackgroundRule(t *testing.T) {
	data, err := os.ReadFile("node_modules/daisyui/base/rootcolor.css")
	if err != nil {
		t.Fatalf("os.ReadFile: %v", err)
	}
	css := string(data)
	if !strings.Contains(css, "background-color:var(--root-bg)") {
		t.Error("rootcolor.css lost its background-color:var(--root-bg) declaration")
	}
	if !strings.Contains(css, "--root-bg:var(--color-base-100)") {
		t.Error("rootcolor.css lost its --root-bg:var(--color-base-100) declaration")
	}
}

// TestAssets_EmbedsHashedAssetDirectory proves the "all:" embed
// pattern is in force: [io/fs.ReadDir]'s doc states a plain "build"
// pattern excludes files and directories beginning with "_" or ".",
// which would silently drop every file SvelteKit writes under _app/.
// Dropping the "all:" prefix is this task's own named mutation 1.
func TestAssets_EmbedsHashedAssetDirectory(t *testing.T) {
	buildFS, err := fs.Sub(Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	entries, err := fs.ReadDir(buildFS, "_app")
	if err != nil {
		t.Fatalf(`fs.ReadDir(buildFS, "_app"): %v (the "all:" embed pattern may be missing)`, err)
	}
	if len(entries) == 0 {
		t.Fatal(`build/_app holds no entries; run "npm run build" in web/`)
	}
}

// TestAssets_PlaceholderCommitted proves build/placeholder.txt exists
// even inside this test's own embedded tree, the fallback that keeps a
// fresh checkout compiling before the first `npm run build` (R-5cii9).
func TestAssets_PlaceholderCommitted(t *testing.T) {
	buildFS, err := fs.Sub(Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	if _, err := fs.Stat(buildFS, "placeholder.txt"); err != nil {
		t.Fatalf("build/placeholder.txt: %v", err)
	}
}
