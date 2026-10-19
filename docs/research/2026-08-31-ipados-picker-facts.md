# iPadOS Safari picker facts (documentary research 2026-08-31)

Attempted documentary closure of uploader gate 3. Three of six questions
closed; the rest need one probe-page run on any family iPad. Sources in
the session transcript; key ones inline.

## Closed by documentation

- webkitdirectory: CONFIRMED absent on iPadOS Safari (WebKit Bugzilla
  271705, open since 2024; MDN BCD's contrary claim is flagged wrong in
  mdn/browser-compat-data#11982). Degrades-to-normal-picker is
  standards-consistent inference.
- accept filtering: CONFIRMED unreliable in BOTH directions (fails to
  filter, or over-filters). Server-side rejection is mandatory - already
  dubplate's design.
- Picker-side size ceiling: none documented at selection/handoff. The
  File API hands a lazy reference; documented large-upload crashes
  (tus-js-client#146 NSMallocException >500MB) are in-page buffering,
  designed out by chunked Blob.slice. No first-person 2GB-from-iPad
  confirmation found; inference from tooling design.

## Design-changing finding

- WebKit UTI bug (root r250410): MIME-wildcard `audio/*` accepts can
  render audio files GREYED OUT and unselectable in Files - all sources
  (local, iCloud, Dropbox) equally (react-dropzone#1039, gradio#4021,
  openradar#19227). Field workaround: extension-based accept lists.
  RULING folded into plan U10: accept values extension-first, never bare
  audio/*. Currency on iPadOS 17/18 unverified - the probe's three
  accept variants settle it empirically.
- Multi-select gesture: on iPadOS the picker hides multi-select behind
  "... -> Select" (Apple dev forums). U10 copy must surface the gesture.

## CLOSED by the prior-art round (same day)

A second research round (prior art + Apple developer docs) closed the
residuals; gate 3 needs NO device:
- audio/* greyout: CONFIRMED ALIVE through iOS 18.3.1 - WebKit bug
  242110 (accept="audio/*" treated as video/*), status NEW since 2022,
  reproduced May 2025 in WordPress/Gutenberg#70119, corroborated by
  react-dropzone#1039 and gradio; no Safari 18.0-18.5 release note
  fixes it. Extension-first accepts are REQUIRED, not defensive.
- ZIP selection: works in production (WeTransfer's mobile-web docs
  assume it; the documented iOS ZIP problem is download-side only).
- Multi-select gesture ("... -> Select"): Apple documents it; no
  shipping uploader's help center does - dubplate's coaching copy
  exceeds prior art.
- Large uploads: no documented iOS byte ceiling anywhere; chunked
  resumable (tus) is the production norm; WeTransfer's "use the app"
  nudge cites reliability, not a platform limit.
- Only unknown left: exact accept strings of auth-walled uploaders
  (Dropbox/Drive/WeTransfer SPAs) - academic, gates nothing.
The published probe (claude.ai artifact "Dubplate Picker Probe") is a
courtesy runtime confirmation only.
