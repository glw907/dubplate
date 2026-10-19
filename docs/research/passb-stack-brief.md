# Pass B / Task U9: Frontend Stack Brief

Dubplate uploader web UI. Svelte 5 (runes) + Vite SPA, DaisyUI 5 + Tailwind 4, embedded
into the Go binary via `go:embed`, served under `/upload/`.

## 1. Dependency set

Versions verified against the npm registry (`registry.npmjs.org`) on 2026-09-01.

| Package | Version | Notes |
|---|---|---|
| `svelte` | `5.57.0` | Runes mode, no peer deps of its own. |
| `vite` | `8.2.2` | Current major; Rolldown-backed bundler landed in the 8.0 line. |
| `@sveltejs/vite-plugin-svelte` | `7.3.0` | Peer deps: `vite: ^8.0.0-beta.7 \|\| ^8.0.0`, `svelte: ^5.46.4`. Both satisfied by the pins above. |
| `tailwindcss` | `4.3.3` | CSS-first config (no `tailwind.config.js` required). |
| `@tailwindcss/vite` | `4.3.3` | Peer dep `vite: ^5.2.0 \|\| ^6 \|\| ^7 \|\| ^8` — compatible with Vite 8. |
| `daisyui` | `5.7.25` | Zero third-party runtime deps as of v5 (npm package shrank 4.7MB -> 1.8MB). Consumed as a Tailwind CSS plugin, not a JS import. |
| `tus-js-client` | `4.3.1` | Stable `latest` dist-tag. A `5.0.0-pre2` exists under the `next` tag — do not use it, it is pre-release and unpublished as stable. |

**Not SvelteKit — plain Vite SPA is correct.** SvelteKit adds a filesystem router, SSR/
prerendering, adapters, and a Node-oriented build output (`.svelte-kit/`, adapter-specific
`build/`) — none of which apply here: this is a single route embedded as static assets in a
Go binary with no server-side rendering, no routing beyond client-side state, and no Node
runtime at request time. A plain Vite SPA build (`vite build`) emits a flat `dist/` of
hashed static assets that `go:embed` can point at directly. Pulling in SvelteKit would mean
fighting its adapter model to force a static, base-path-relative single bundle it wasn't
designed to produce. Confirmed: use `@sveltejs/vite-plugin-svelte` directly in `vite.config.ts`,
not `@sveltejs/kit`.

**Known incompatibility note (Tailwind 4 + DaisyUI 5 wiring).** DaisyUI 5 dropped its old
`tailwind.config.js` `plugins: [require("daisyui")]` registration (that's the v3/v4-Tailwind
pattern). Under Tailwind 4 it's wired entirely from CSS with `@plugin "daisyui";` inside the
stylesheet that also has `@import "tailwindcss";`. Using the old JS-config registration style
with Tailwind 4 either no-ops or errors depending on setup — don't mix the two config models.
No version-range incompatibility exists between the pins above; the incompatibility class is
purely "old DaisyUI/Tailwind wiring pattern used with new Tailwind."

## 2. tus-js-client integration

**Endpoint.** Point the client's `endpoint` option at the tusd creation endpoint mounted
under `/upload`, e.g. `/upload/files/` (tusd's default sub-path is `/files/`, reachable here
via `-base-path` or a reverse-proxy rewrite so the public URL is `/upload/files/`). Use a
relative URL so it works identically in dev (proxied) and prod (embedded).

```js
new tus.Upload(file, {
  endpoint: '/upload/files/',
  chunkSize: 8 * 1024 * 1024, // match server-side max PATCH body config, see below
  retryDelays: [0, 3000, 8000, 15000, 30000],
  storeFingerprintForResuming: true,
  metadata: { filename: file.name, filetype: file.type },
  onError: (error) => { /* surface to UI */ },
  onProgress: (bytesUploaded, bytesTotal) => { /* progress bar */ },
  onSuccess: () => { /* mark journey step complete */ },
});
```

**Chunk size interplay with the 8MB server config.** `tus-js-client`'s `chunkSize` default is
`Infinity` — it tries to PATCH the whole file in one request unless told otherwise. Since the
server (tusd) is configured with an 8MB max PATCH body, `chunkSize` **must** be set explicitly
to `8 * 1024 * 1024` (or smaller) client-side, or every PATCH past the first 8MB will be
rejected by tusd. Setting it exactly to the server's limit maximizes chunk size (fewer round
trips) while staying under the ceiling; going smaller only helps if you want more frequent
progress/resume checkpoints on flaky networks (iPad/cellular is the case that argues for it —
consider 4MB there to bound how much re-upload a mid-chunk failure costs).

**Resume behavior.** With `storeFingerprintForResuming: true` (the default), the client
computes a fingerprint from file name, size, type, and last-modified time and stores the
resulting upload URL in `localStorage` (`canStoreURLs` gates this — it's `false` in contexts
without `localStorage`, e.g. some private-browsing modes). On a fresh `new tus.Upload()` call
for a file matching a stored fingerprint, the client issues a `HEAD` to the stored URL to read
`Upload-Offset` and resumes from there instead of re-creating the upload. Practical
implication: reusing the same `File` object (same name/size/mtime) across a page reload is
what makes resume work — a re-picked file with different mtime breaks the fingerprint match.

**Location header handling.** On the initial `POST` to the creation endpoint, tusd returns a
`Location` header pointing at the new upload resource. Per the tus protocol this may be a
relative or absolute URI; tus-js-client resolves it against the request URL and immediately
issues a confirming `HEAD` against it before starting PATCHes — this is automatic, nothing to
wire up manually. One thing worth checking at implementation time: if tusd sits behind a
reverse proxy that rewrites the public path (e.g. `/files/` internally vs `/upload/files/`
externally), tusd needs `-behind-proxy` plus correct `X-Forwarded-Host`/`X-Forwarded-Proto`
(and if the path itself is rewritten, `X-Forwarded-Prefix` or equivalent path handling) so the
`Location` header it emits is externally resolvable — otherwise the client will try to HEAD/
PATCH a URL the browser can't reach.

**Retry/backoff overrides.** Default `retryDelays` is `[0, 1000, 3000, 5000]` (4 attempts,
maxes out at 5s between retries, ~9s total wall time before giving up). For large dubplate
files over consumer/venue Wi-Fi this is thin. Recommend widening to something like
`[0, 3000, 8000, 15000, 30000]` (longer tail, more attempts) so transient drops during a
multi-minute upload don't surface as a hard failure to the user. Pair with `onShouldRetry` if
finer control over which HTTP statuses retry is needed (e.g. don't retry on 4xx auth
failures).

**Safari/iPad gotchas.**
- **Background-tab/backgrounded-app throttling.** iOS Safari (and WKWebView-based apps)
  suspends or throttles JS timers and network activity when a tab is backgrounded or the
  device locks, which can stall or drop an in-flight PATCH mid-chunk. This is a platform
  constraint, not something tus-js-client controls — the mitigation is the resume behavior
  above (small-ish `chunkSize`, real `retryDelays`) plus UI messaging telling users to keep
  the tab foregrounded during upload, not a code fix.
- **Memory with many/large files.** There are long-standing tus-js-client issues on iOS around
  large in-memory buffers when Safari's video-compression pipeline creates a new Blob for
  large video/audio uploads — this can produce `NSMallocException`-style crashes on files in
  the several-hundred-MB range on memory-constrained iPads. For dubplate-sized audio files
  this is a lower risk than video, but keep chunked reads (never load the whole file into
  memory at once — `tus.Upload` does this correctly by default via `Blob.slice`) and avoid
  holding multiple large `File`/`Blob` references live simultaneously if a multi-file queue is
  added later.
- **Fingerprint invalidation on iOS media re-encoding.** If iOS Safari transcodes/compresses a
  picked media file before upload (observed for video; less consistently for audio), a resumed
  session can get a new file object with a different fingerprint, silently starting a fresh
  upload instead of resuming. Worth a defensive check: if a resume attempt results in offset 0
  and the file size doesn't match what was previously reported, don't treat it as a bug — it's
  expected iOS behavior in this class of issue.

## 3. SSE client pattern

Native `EventSource` can only issue a `GET` and cannot set custom headers or a request body,
but our sessions are POST-opened (session created via `POST`, journey/history events streamed
after). Standard pattern for this:

1. `POST` to the session-creation endpoint (e.g. `/upload/journey/stream/session`), body
   carries whatever context the server needs to build the session; response is JSON containing
   a `sessionId` (or equivalent token).
2. Open `new EventSource('/upload/journey/stream?sessionId=' + sessionId)` — the session id
   travels in the URL query string, since `EventSource` cannot carry it in a header or body.
   The server correlates the GET to the session it already created via POST.
3. Treat `EventSource` as the steady-state transport; it handles reconnection natively.

```js
async function openJourneyStream(uploadId) {
  const res = await fetch('/upload/journey/stream/session', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ uploadId }),
  });
  const { sessionId } = await res.json();

  const es = new EventSource(`/upload/journey/stream?sessionId=${sessionId}`);
  es.addEventListener('journey', (e) => handleJourneyEvent(JSON.parse(e.data)));
  es.onerror = () => { /* EventSource auto-reconnects; surface a transient "reconnecting" state */ };
  return es;
}
```

**Last-Event-ID reconnection.** `EventSource` tracks the `id:` field of the last received
event automatically and, on any reconnect (network drop, tab wake, server restart), sends it
back as the `Last-Event-ID` request header on the reconnecting `GET`. No client code is needed
to make this work — it's built into the spec — but the **server must**: (a) assign a
monotonic/orderable `id:` to every SSE event it emits, (b) read `Last-Event-ID` on incoming
GETs and replay any missed events from that point, and (c) keep enough history (or a
replayable log keyed by session) to satisfy that replay for at least the reconnect window. The
15s heartbeat should also carry an `id:` (or the id of the last real event) so a reconnect
right after a heartbeat-only period doesn't lose position. `EventSource`'s default retry delay
is set by the server via the `retry:` field (in ms) if the default (browser-chosen, typically
~3s) isn't desired.

**Fetch-based fallback (if headers are ever needed).** If a future requirement needs custom
headers (auth bearer token instead of a cookie, etc.) that `EventSource` can't set, replace it
with a `fetch()` + `ReadableStream` reader parsing the `text/event-stream` format by hand (or a
small library like `@microsoft/fetch-event-source`, which explicitly exists to give SSE
semantics — including auto-retry and `Last-Event-ID` handling — over `fetch`). Only worth
introducing if the header requirement is real; native `EventSource` is simpler and sufficient
for a session-id-in-URL design.

## 4. Vite config: base `/upload/` + `go:embed`

```ts
// vite.config.ts
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

export default defineConfig({
  base: '/upload/',
  plugins: [svelte(), tailwindcss()],
  build: {
    outDir: 'dist',       // go:embed target, see below
    emptyOutDir: true,
    assetsDir: 'assets',  // default; hashed chunks land under dist/assets/
  },
  server: {
    proxy: {
      // dev: forward API/tus/SSE calls to the running Go server on :8090
      '/upload/files': { target: 'http://localhost:8090', changeOrigin: true },
      '/upload/journey': { target: 'http://localhost:8090', changeOrigin: true },
      '/upload/config': { target: 'http://localhost:8090', changeOrigin: true },
    },
  },
});
```

**`build.outDir` convention.** Point it at a directory the Go module embeds directly, e.g.
`frontend/dist`, with a sibling `embed.go`:

```go
package frontend

import "embed"

//go:embed dist/*
var Assets embed.FS
```

Serve it with `http.StripPrefix("/upload/", http.FileServer(http.FS(sub)))` where `sub` is
`fs.Sub(Assets, "dist")` — matching `base: '/upload/'` so asset URLs emitted in `index.html`
(`/upload/assets/index-<hash>.js`, etc.) resolve correctly both in dev (Vite dev server itself
respects `base`) and prod (embedded, served under the same prefix).

**Asset hashing.** Vite's default production build content-hashes every emitted JS/CSS/asset
filename (`[name]-[hash].js`) and rewrites all references in the built `index.html` and
chunk graph automatically — this is on by default, no config needed, and it's what makes
`go:embed` + long-lived cache headers on the Go server safe (a code change always produces a
new filename, so cache-forever is fine for hashed assets; only `index.html` itself should be
served with no-cache).

**Dev proxy for local dev against `:8090`.** The `server.proxy` block above forwards any
`/upload/*` API path to the running Go binary while Vite serves the SPA shell and HMR on its
own dev port. This lets `npm run dev` be used against a real backend without CORS
complications (same-origin from the browser's perspective, since Vite proxies server-side).
For the tus and SSE endpoints specifically, plain path-based proxying is sufficient — neither
protocol needs special dev-proxy handling (tus is plain HTTP POST/PATCH/HEAD; SSE is a plain
long-lived GET, and Vite's proxy passes streaming responses through).

## 5. DaisyUI theming to match Navidrome

**DaisyUI 5 custom theme syntax (Tailwind 4 CSS-first, `@plugin` directive).** No
`tailwind.config.js` — themes are declared inline in the stylesheet that imports Tailwind:

```css
/* app.css */
@import "tailwindcss";
@plugin "daisyui" {
  themes: navidrome-light --default, navidrome-dark --prefersdark;
}

@plugin "daisyui/theme" {
  name: "navidrome-light";
  default: true;
  color-scheme: light;

  --color-base-100: oklch(100% 0 0);        /* page background */
  --color-base-200: oklch(98% 0 0);
  --color-base-300: oklch(94% 0 0);
  --color-base-content: oklch(20% 0 0);      /* body text, ~#000 */
  --color-primary: oklch(45% 0.19 265);      /* Navidrome accent, ~#3f51b5 */
  --color-primary-content: oklch(98% 0.01 265);
  --color-secondary: oklch(55% 0.2 240);     /* focus/link accent, ~#0085ff */
  --color-secondary-content: oklch(98% 0.01 240);
  --color-neutral: oklch(30% 0.02 265);
  --color-neutral-content: oklch(98% 0 0);
  /* info/success/warning/error left at DaisyUI defaults unless a specific
     Navidrome equivalent is confirmed against the live app */
}

@plugin "daisyui/theme" {
  name: "navidrome-dark";
  prefersdark: true;
  color-scheme: dark;

  --color-base-100: oklch(26% 0 0);          /* ~#424242 card/paper */
  --color-base-200: oklch(22% 0 0);
  --color-base-300: oklch(18% 0 0);
  --color-base-content: oklch(96% 0 0);      /* ~#eee body text */
  --color-primary: oklch(85% 0.08 240);      /* ~#90caf9 Navidrome dark accent */
  --color-primary-content: oklch(20% 0.02 240);
  --color-secondary: oklch(55% 0.2 240);     /* ~#0085ff link/focus accent, shared with light */
  --color-secondary-content: oklch(98% 0.01 240);
  --color-neutral: oklch(90% 0 0);
  --color-neutral-content: oklch(20% 0 0);
}
```

Source colors pulled from Navidrome's own Material-UI theme files
(`ui/src/themes/{light,dark}.js` in `navidrome/navidrome`): dark theme primary `#90caf9`,
dark paper/surface `#424242`; light theme secondary `#3f51b5`, shared focus/link accent
`#0085ff` in both themes. These are close reads of the MUI palette, not an exact port — treat
the OKLCH values above as a starting point to refine visually against the running Navidrome
UI, not a pixel-exact match; a couple of Navidrome's fields (its default light `background`)
aren't explicitly set in the theme file and fall through to MUI defaults, so confirm those two
against a live screenshot before calling the palette final.

**`--default` / `--prefersdark` semantics.** `--default` on a theme in the `themes:` list
picks the theme applied when no explicit `data-theme` is set and the OS preference doesn't
match any `--prefersdark` theme. `--prefersdark` marks a theme as the one DaisyUI switches to
automatically when `prefers-color-scheme: dark` matches and no manual override is stored —
this is DaisyUI's own built-in system-preference wiring, so a separate manual
`prefers-color-scheme` media query isn't needed for the automatic case.

**Manual theme switching interaction with `prefers-color-scheme`.** DaisyUI applies whichever
theme has `--prefersdark` automatically only until something sets `data-theme` explicitly on
`<html>` (or a persisted override is applied on load). Standard pattern: on page load, check
`localStorage` for a saved user choice; if present, set `document.documentElement.dataset.theme`
to it (overrides system preference); if absent, do nothing and let `--prefersdark`/`--default`
resolve it from the OS setting via `prefers-color-scheme`. A manual toggle control writes the
choice to `localStorage` and sets `data-theme` directly; there's no need to also listen for
`prefers-color-scheme` changes once a manual choice exists — only re-derive from system
preference if the user later clears their override.
