# Uploader UX best practices (researched 2026-08-31, for the dubplate uploader plan)

Sonnet research dispatch, web-sourced with citations. Feeds the uploader
plan's Phase P brief and U9-U11 criteria.

## 1. Drag-and-drop vs file-picker across devices

Drag-and-drop is desktop-only; there is no native file-drag gesture on
iOS/Android. Modern uploaders (Uppy Dashboard, Google web properties) treat
the drop zone as progressive enhancement over a big tappable "select files"
button opening the OS picker - the button is the primary target on touch.
Folder upload (webkitdirectory) works on desktop Chrome/Firefox/Edge and
Android but is NOT implemented in iOS/iPadOS Safari (2025-2026; no
showDirectoryPicker / File System Access API either). "Select a folder of
FLACs" is impossible on iPhone/iPad; only multi-file selection or a ZIP.

Recommendation: ZIP upload is the primary, universal path on every
platform; folder-select is a secondary desktop/Android convenience via
webkitdirectory, feature-detected and hidden (not disabled) on iOS Safari.
Message iPhone/iPad users explicitly toward multi-select or "zip it first".

Sources: uppy.io/docs/dashboard, MDN webkitDirectory,
developer.apple.com/forums/thread/816515, caniuse.com/input-file-directory

## 2. iOS/iPadOS picker and memory specifics

Files-app picker: multi-file selection works, third-party providers
(Dropbox/Drive/OneDrive) expose individual files but folder selection from
them is greyed out; only local/iCloud folders select as folders. Mobile
Safari has a hard memory ceiling that crashes the page around ~100MB
(iPhone SE) to ~200MB (iPad) when JS holds large in-memory buffers
(FileReader whole-reads, data URLs). FileReader chunked reads have a
documented iOS bug: reads silently stop after 60 seconds
(WebkitBlobResource error 1). Any client that pre-reads a whole 1-2GB
album ZIP will crash on an iPhone; tus-js-client's Blob.slice() chunking
avoids this - never touch the whole buffer (no client-side full-file
hashing or preview decode).

Recommendation: rely on tus-js-client's native chunked Blob.slice upload;
never pre-read whole files on iOS; "files from Drive/Dropbox" supported,
"folders from Drive/Dropbox" not.

Sources: support.apple.com/en-tm/102238,
developer.apple.com/forums/thread/120257,
lapcatsoftware.com/articles/2026/1/7.html,
developer.apple.com/forums/thread/686192

## 3. Progress and long-running upload UX

Best practice is a clean state machine with distinct UI per phase:
selecting -> uploading (byte-accurate progress) -> processing (staged or
indeterminate) -> done/needs-review/rejected. Client shows real transfer
progress; after bytes finish, the async pipeline is polled/streamed by ID,
never faked into one bar. Distinguish TRANSFER progress from DURABLE
server-acknowledged offset - tus resumability only covers what the server
persisted. iOS Safari suspends JS when a tab backgrounds: uploads
pause-and-resume-on-return, not silent-continue; a web app cannot get true
OS background transfer. Wake Lock (all major browsers 2025) prevents
screen-dim while foregrounded but releases on backgrounding - it is a
"keep the screen on" nicety, never a background-upload promise.

Recommendation: byte progress during upload; labeled staged indicator for
the pipeline with honest copy ("this can take a few minutes"); state
plainly that closing/backgrounding pauses the upload and it resumes on
return.

Sources: LogRocket async-workflow UI patterns, buildo.com tus post,
github.com/tus/tus-js-client/issues/142, MDN Screen Wake Lock API

## 4. Error and rejection UX

Batch items render as independently-stated rows, each with its own status
and specific inline error - never generic "upload failed". Failed items
individually retryable without re-uploading successes. For all-or-nothing
verdicts (album as verdict unit): still enumerate every file's individual
result even though the batch outcome is singular - show exactly which
tracks triggered the rejection and why, then state plainly the album is
held, framed as a property of the files, never a judgment of the uploader.
Error copy: what happened, why, what to do next - one sentence each.

Recommendation: per-track checklist inside the album card (pass/fail per
file, reject-taxonomy wording), one calm summary line, a single batch
retry affordance, zero blaming language.

Sources: saasui.design file-upload patterns, oneuptime.com bulk-API
partial success, standardbeagle.com trust-through-UX

## 5. Responsive layout patterns

Desktop: spacious drop zone + queue table. Mobile: single-column stack of
full-width cards (tables don't reflow; horizontal scroll in queues is an
anti-pattern). DaisyUI 5 fits directly: table with overflow-x-auto above
the breakpoint, cards below; drawer/modal serve as bottom-sheet-style
panels. Touch targets: WCAG 2.2 SC 2.5.8 (AA) requires 24x24 CSS px
minimum; Apple HIG 44x44pt; Material 48x48dp. Practitioners design to
44px as the de facto floor.

Recommendation: queue as a DaisyUI card list, full-width single-column
below md:, table above; every interactive control at least 44x44 px at
every breakpoint.

Sources: daisyui.com/components/table, WCAG 2.5.8 guides, LogRocket touch
target sizes

## 6. Accessibility for upload flows

aria-live="polite" regions for progress and status changes (W3C
ARIA25/ARIA27); announce at meaningful increments (10-25%), not every
percent; completion is a distinct final announcement. Live regions exist
empty in the DOM at load, populated later - never injected fresh. When an
item enters the queue, announce it (or move focus) so keyboard/SR users
are not stranded. prefers-reduced-motion: reduce disables animated
progress fills for instant value jumps (WCAG 2.3.3).

Recommendation: one polite live region for stage transitions and per-file
changes, throttled to significant deltas; all progress animation behind a
prefers-reduced-motion guard.

Sources: W3C ARIA27, uploadcare.com accessibility guide, MDN
prefers-reduced-motion

## 7. Recent tus/Uppy developments

The tus protocol is mid-standardization at the IETF
(draft-ietf-httpbis-resumable-upload, draft -12, Transloadit + Apple
co-authors, targeting RFC; expiry Jan 2027). tus-js-client 4.3.x speaks
classic tus 1.0; the draft protocol sits behind its `protocol` option
(matches the spec's thin-seam note). Uppy's Golden Retriever
(localStorage metadata + IndexedDB small files + optional Service Worker)
does NOT survive a real browser crash for large files - only refresh or
closed tab. Plain tus resume (re-attach to the same upload URL) is the
reliable guarantee.

Recommendation: no client-side file caching for multi-hundred-MB files;
persist the tus upload anchor (server-side per verified email, per the
spec) and set the expectation that a fully-closed browser requires
re-selecting the file - tus resumes the TRANSFER, not the file handle.

Sources: datatracker.ietf.org draft-ietf-httpbis-resumable-upload,
tus.io IETF blog post, uppy.io Golden Retriever docs
