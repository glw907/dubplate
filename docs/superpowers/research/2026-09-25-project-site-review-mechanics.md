# Project site spec review: mechanics and feasibility

Target: `docs/superpowers/specs/2026-09-25-project-site-design.md` at `e91f528`. Lens: does
every mechanism behave as stated, verified from source or docs. Sources read: dubplate at
`e91f528` (git objects only), cairn-cms at `73e09d7b`, ecxc-ski at `7817b36`, daisyUI 5.7.46's
`theme/index.js` (dubplate `web/node_modules`), Cloudflare docs. Findings are ranked by
consequence. Six of the spec's mechanisms check out (listed at the end).

## Findings

### M1 (major): the theme files do not move "with no translation step"

Location: spec lines 92-94 ("A cairn site's theme redeclares the same DaisyUI theme blocks over
its chassis, so the token files move as files, with no translation step") and 108-110.

The files can be copied verbatim. But they don't produce a themed site without a site-side
adapter, for four reasons.

1. **No default or dark-preference flag.** dubplate's blocks declare only `name` and tokens.
   daisyUI's plugin gives a block the `:root` selector only on `default: true`, and emits the
   dark-media rule only on `prefersdark: true`:
   `if (isDefault) { selector = \`:where(${root}),${selector}\` }` and
   `if (prefersdark) { ... ":root:not([data-theme])" ... }`
   (`web/node_modules/daisyui/theme/index.js`). Otherwise the tokens apply only under
   `[data-theme="dubplate-label"]`. The cairn template relies on those flags with `data-theme`
   unset: "No cookie leaves `data-theme` unset, and theme.css's own prefers-color-scheme block
   picks the system scheme" (ecxc-ski `src/app.html:16-17`; Waymark `theme.css:97-99,145-147`
   declares `default: true` / `prefersdark: true`). dubplate's app avoids this by stamping
   `data-theme="__DUBPLATE_THEME__"` server-side (`web/src/app.html:2`) and by `system.css`,
   which the spec says probably doesn't travel. Copied as they are, a first-time visitor gets
   no dubplate tokens at all.
2. **Theme names are hardcoded in the template.** The no-flash script matches
   `cairn-site-theme=(cairn-dark|cairn)` (`templates/waymark/src/app.html:12`), and
   `SiteHeader.svelte:67-73` types `'cairn' | 'cairn-dark'`. These are site edits, and they are
   small ones.
3. **The font roles differ.** The cairn chassis and theme read `--font-display`,
   `--font-body`, and `--font-mono` from `@theme` (`src/chassis/tokens.css:42-44`, ecxc-ski
   `theme.css` `@theme` block). dubplate's faces are `--font-sans`, `--font-serif`, and
   `--font-mono` inside the daisyUI blocks (`web/src/lib/themes/dubplate.css`). So the site's
   prose and chrome keep the Waymark faces unless it maps display to serif and body to sans.
4. **The ink names differ.** cairn's callouts and code highlighting read
   `--cairn-<status>-ink` (Waymark `theme.css:267-270`, recipe step 6 in ecxc-ski `theme.css`).
   dubplate's inks are `--ink-primary`, `--ink-info`, and so on (`inks.css`). Without a
   mapping, the site shows Waymark's inks next to dubplate's fills.

Also, the "travel" list (line 109: "The palette, ink, and font files travel") doesn't name
`dubplate.css`, the match axis. That file carries the radius, border, depth, noise, and all
three font stacks. Leave it behind and the composed themes lose their geometry and faces. It
moves cleanly, since it is plain `@plugin "daisyui/theme"` blocks with no `@theme` or
`@utility`. `surface.css` and `inks.css` are `@layer base` attribute-selector blocks, which are
harmless in the site.

Fold: say the token files move verbatim and unedited, and hash-lock them. Then say the site
owns one adapter file (site-owned, outside the lock) that does four things: it stamps or flags
the default and dark theme, it renames the toggle's two names, it maps `--font-display` and
`--font-body` onto dubplate's faces, and it maps `--cairn-*-ink` onto `--ink-*`. Add
`dubplate.css` to the travel list. The one-way drift rule still holds, because the adapter holds
only references (`var(--ink-info)`) and no values.

### M2 (major): the reject generator's source doesn't exist in the greenfield tree

Location: spec lines 58 and 79-80 ("The reject page builds from `internal/reject`, whose
wording is already contract").

At `e91f528`, `internal/reject` exists only as `_salvage/internal/reject/reject.go`. The only
greenfield packages are `atomicfile auth backup config engine harness identity notify ops
serve`. CLAUDE.md is explicit about salvage: it is "never imported, and retired at launch". The
build spec lists `internal/reject/` only as a planned leaf (`2026-09-02-greenfield-server-build.md:434`).
The salvage shape can't be walked without help, for two reasons. First, `Reason` is an `iota`
enum and `Friendly()` is a `switch` with no enumerator. Second, one case is parameterized: "the
bad-extension case falls back to a generic phrasing here; RenderEmail names the actual extension
instead". Under "a 0.8 forward hook is named in the rung plan and built by the rung that first
consumes it", this generator is a real consumer. Nothing routes its need to the rung that builds
`internal/reject`.

Fold: say "the reject copy package the build spec plans (`internal/reject`)". Also route a hook
to the forward-hooks list: an enumerable reason set (for example `reject.All()` or an iterator)
with a canonical form for each parameterized message. The rung that builds the package owns the
hook, and the site generator is its named consumer. The generator can't be built until that
rung lands. Say so in Sequence, because Phase A's trigger (cairn's docs reset) can fire first.

### M3 (major): the config registry can't yield a reference page as it stands

Location: spec lines 59 and 80-81 ("Every config key" / "builds from the config key registry").

The registry is walkable: `(*Registry).OwnedWall`, `ReservedClasses`, `IntegrationClasses`,
`Thresholds`, `IdentitySources`, and `Superseded` (`internal/config/registry.go:157-469`). But
only `Threshold` carries prose (`Key, Kind, Default, Doc`, `thresholds.go:45-49`). `WallRow` is
`{Name, Pending}`, `IdentitySource` and `IntegrationClass` are names and key lists, and none
carries a description or an example. A generated page would list names with no meanings. Also,
the production registry is assembled by `newProductionRegistry()`, which is unexported in
`package main` (`cmd/dubplate-server/main.go`). A generator outside that package can't build it
without an export ahead of a caller, which the house rule forbids. Last, the configuration homes
spec gives four homes. The env wall is one of them, so "every config key" is ambiguous.

Fold: (a) the generator lives in `cmd/dubplate-server`, as a hidden subcommand or a golden-file
test, and calls `newProductionRegistry()`. (b) Add a `Doc` field to each declaration family as
a forward hook with this generator as its consumer. **OWNER FORK** on scope. Option 1: the
reference covers the env wall only, with a pointer to where the other homes are set.
Option 2: it covers all four homes. Recommend option 1, since it is the only home a single
registry describes today.

### M4 (major): the spec never says where the generated pages live at the tag

Location: spec lines 47-48 ("Generated pages are built from dubplate's code and vendored the
same way") and 96-99.

The sync "exports the owned files from a dubplate release tag". An export (`git archive` or
`git show`) carries only committed files. So either the generated Markdown is committed in
dubplate, and a dubplate gate leg regenerates it and diffs it so the tag can't carry stale
output, or the site's sync script runs the Go generators against the checkout. The second
option needs a Go toolchain in the site repo's workflow and breaks the "vendored the same way"
claim. Without a freshness leg, "can't drift from the code" (line 79) is unenforced.

Fold: the generated pages are committed under `docs/guide/generated/` (or similar), and a
`scripts/check.sh` leg regenerates them and fails on a diff. The sync then treats them like
every other owned file.

### M5 (major): `npx create-cairn-site` doesn't resolve today, and the docs-reset trigger doesn't publish it

Location: spec lines 11-12, 39-40, and 124.

`packages/create-cairn-site/package.json` is at `"version": "0.0.0"`, and
`npm view create-cairn-site` returns `404 Not Found`. cairn's ROADMAP treats publication as a
separate decision: "Trigger: before `create-cairn-site` publishes, which is the same decision the
release cut already owes" (`ROADMAP.md:413`). It also says "The template repo's pre-release
notice comes off when `create-cairn-site` publishes" (`:569`). The docs reset entry
(`ROADMAP.md:285+`) doesn't include a publish. A "stranger" trial that starts with `npx` fails
at step one, or quietly borrows a local tarball, which is exactly the off-path step line 116
forbids.

Fold: the trigger becomes "the docs reset has closed **and** `create-cairn-site` is on npm".
Alternatively, the trial record logs the tarball route as a known, pre-registered deviation.

### M6 (minor): the admin track path is being redrawn, and its audience doesn't cover this build

Location: spec lines 114-116 ("follows cairn's admin track literally, starting from
`docs/admin/README.md`").

The docs reset spec says "The track structure is reopened (2026-09-23)" and names six
audiences, among them a "site operator" and a "site designer" (`2026-09-23-docs-reset-design.md:119-124`).
So `docs/admin/README.md` may not exist when the trigger fires. The current admin track also
says "You never write code here" (`docs/admin/README.md:3-5`). Phase A writes code: the theme
adapter (M1), a vendored-Markdown route, a sync script, and a hash gate. Those are designer or
extender jobs.

Fold: name the trial by audience ("the site operator's and site designer's entry points as the
reset publishes them"), not by path. State that the code steps are expected to leave the
operator track, so they aren't logged as friction by default.

### M7 (minor): no dubplate release tag exists for Phase A to pin

Location: spec lines 96-97 and 126-127.

`git tag -l` lists only `pre-greenfield`. Phase A builds the site and the sync script before any
0.x or 0.8 tag exists, so "exports the owned files from a dubplate release tag" has no input.

Fold: the sync accepts any commit-ish and records it in the lock. The gate passes a non-release
pin in Phase A, and Phase B's preconditions require a release tag in the lock.

### M8 (minor): the noindex mechanism needs one sentence

Location: spec line 127 ("a preview deploy on a `workers.dev` address marked `noindex`").

This is achievable, but it isn't automatic on the address Phase A would use. Cloudflare says
"Cloudflare adds this header automatically to `workers.dev` Preview URLs", meaning
`<preview>-<worker>.<subdomain>.workers.dev`, and not the production `workers.dev` route. A cairn
site is a Worker with `hooks.server.ts` (ecxc-ski), so the header or a
`<meta name="robots" content="noindex">` works. It must be host-conditional, so the header
disappears on the custom domain at Phase B. Don't add the path to the `robotsResponse`
`disallow` list, because a crawler barred by robots.txt never sees the noindex.

Fold: "an `X-Robots-Tag: noindex` header set in the site's server hook for `*.workers.dev`
hosts only".

### M9 (minor): vendored pages sit outside cairn's link and sitemap machinery

Location: spec lines 46-48.

A vendored guide can be rendered: cairn exports `parseMarkdown` and `createRenderer`
(`src/lib/index.ts:53,107`), and ecxc-ski's `render.ts` composes one. Three things don't come
for free, though. First, relative `.md` links and images written for reading in the dubplate
repo break on the site. Second, ecxc-ski's sitemap adds non-concept routes by hand: "None of
those three are concept entries, so the site resolver cannot see them; they are added here by
hand" (`src/routes/sitemap.xml/+server.ts`). Third, `cairn:` links and `verifyReferences` don't
cover the vendored pages.

Fold: the vendored-page route rewrites guide-relative links to site routes, and the sitemap
lists the vendored and generated routes. Site-owned pages link to guides by plain path.

### M10 (minor): the command generator's source is package main, and cobra/doc is prior art

Location: spec lines 60 and 81-82.

The cobra tree exists and can be walked. But `newRootCmd` is unexported in
`cmd/dubplate-server/main.go`, and `cmd/harness` is a second cobra tree that must be excluded.
The binary is `dubplate-server` today, so the generated page names whatever 0.8 ships, as line
86 already says for `init` and `upgrade`. `github.com/spf13/cobra/doc`'s `GenMarkdownTree` is
the obvious adopt candidate. Its man-page file imports `go-md2man`, which would join `go.sum`.

Fold: the generator lives beside M3's in `cmd/dubplate-server` and walks `newRootCmd()`. The
Phase A plan weighs `cobra/doc` as a prior-art row.

## Checked and sound

- **Publish commits to main.** A cairn site's config names `owner`, `repo`, and
  `branch: 'main'` (ecxc-ski `src/theme/cairn.config.ts:77-79`). Drafts ride `cairn/*` branches
  and publish to main (`src/lib/content/reference-index.ts:7`). A separate repo keeps those
  commits out of dubplate's tree, as line 41 says.
- **The editor can't touch vendored files.** Editable content is the concept `dir`s
  (`src/content/posts`, `src/content/pages`, ecxc-ski `cairn.config.ts:42,64`). A vendor
  directory outside them is invisible to the admin.
- **A site route can render vendored Markdown.** The `parseMarkdown` and `createRenderer`
  exports cover it (see M9 for the loose ends).
- **Exporting from a private local checkout works.** `git archive` and `git show <tag>:<path>`
  run against the local object store, so repo visibility doesn't matter. Committed vendor files
  mean Cloudflare's build never needs dubplate access.
- **The copied blocks don't break the site's Tailwind.** `dubplate.css`, `label.css`, and
  `label-night.css` contain only `@plugin "daisyui/theme"` blocks, and dubplate's `app.css`
  holds the only `@import 'tailwindcss'`. The site's chassis already activates the daisyUI
  plugin with `themes: false` (`tokens.css:36-37`), so the copied blocks compose. The
  quoted-font-stack workaround in `dubplate.css` carries over unchanged.
- **The font files need care, but not a translation.** `fonts.css` uses root-absolute
  `url('/fonts/*.woff2')`, so the sync must write the three `.woff2` files, and their OFL texts,
  to the site's `static/fonts/` (hash-locked there). A vendor directory elsewhere would 404 the
  faces. Record this in the plan.
