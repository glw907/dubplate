# Uploader prior-art teardown (researched 2026-08-31, for the dubplate uploader plan)

Sonnet research dispatch over product docs, teardowns, and user complaint
threads. The synthesis and anti-patterns sections are design constraints
for the uploader pass; per-product notes carry the evidence.

## Per-product highlights

- **WeTransfer**: the whole page is the drop target; everything
  non-essential hidden until needed; calm, human copy. Complaint ceiling:
  plain-HTTP transfers with no resume die on very large files (user
  reports of 50-60GB failures) - the direct argument for tus.
- **Uppy Dashboard**: per-file progress/error/retry cards PLUS an
  aggregate status bar; pause/resume both per file and batch-wide. The
  Informer pattern: transient toasts for events, a durable per-item list
  for state - a fleeting message is never the only record. Golden
  Retriever's asymmetric recovery: resume what you can, show a "ghost
  file" and re-request only what you can't, never force a full restart.
  (Informer merged into Dashboard as of Uppy 5.0.)
- **Bandcamp** (closest domain match): lossless-only enforced at intake
  (>=16-bit/44.1k) with educational, artist-protective copy explaining WHY
  (server transcodes need a lossless master). Album assembled as a DRAFT
  object - tracks, order, art trickle in; publish is explicit and
  decoupled from "upload finished". FLAC nudged over WAV (half the upload
  time). Complaints: long processing, occasional corrupt output forcing
  re-upload.
- **DistroKid / TuneCore**: exact specs published BEFORE upload (formats,
  bit depth, caps). Complaints: rejection reasons arrive out-of-band
  (email), are hard to relocate later, and fixes can force
  delete-and-reupload of the whole release - both named pain points in
  user threads.
- **Google Photos/Drive mobile**: background upload is bursty and
  OS-gated; the one checkable terminal state ("Backup complete") is the
  signal. Documented complaint class: users ASSUME backup happened when
  it silently stalled - a false "done" is worse than a slow honest
  "still working".
- **Dropbox web**: never silently overwrites - name collision interrupts
  with replace/cancel/keep-both; sync conflicts preserve both copies.
  Data-preservation bias, resolution pushed to the user.
- **Apple iCloud / Files-app flows**: on iOS/iPadOS the native picker and
  Share Sheet ARE the upload UX users know; bespoke in-page pickers with
  novel gestures (WeTransfer's long-press) confuse non-technical users.

## Ten patterns to adopt

1. Drop zone unmissable and forgiving on desktop; native picker deferred
   to entirely on mobile.
2. Per-item state alongside batch state (one track can fail while eleven
   succeed).
3. Transient notices separated from persistent status (Informer pattern).
4. Publish the spec before the user hits a wall (formats, caps, art) -
   reject-then-explain generates support tickets.
5. A partial failure never destroys the whole submission; recover or
   re-request only the broken piece.
6. Crash/interruption recovery automatic, not a lost afternoon.
7. The album is a draft object, not an atomic transaction - review can
   span more than one sitting.
8. Explicit conflict resolution, never silent overwrite; preserve data,
   ask.
9. One unambiguous terminal state checkable from outside the flow ("safe
   to close this tab").
10. Tone: calm, plain-language, blame-free, specific about the fix,
    explaining WHY a constraint exists (Bandcamp's register).

## Anti-patterns (from user complaints)

- Rejection detached from the item it's about (email-only, hard to
  relocate later).
- Ambiguous completion that reads "done" when it isn't (Google Photos
  complaint class).
- Unresumable browser transfer at album scale (WeTransfer failures).
- Full resubmit forced by one bad file/field (DistroKid pain point).
- Non-native mobile selection gestures (WeTransfer iOS long-press).

Verification notes: Uppy Informer's merge into Dashboard (5.0) verified;
DistroKid's live rejection-screen copy could not be fetched directly
(403) - its flow description is from support-article summaries,
directionally accurate, not verbatim UI.

Sources: wetransfer.com/resources, ashrafali.net WeTransfer critique,
uppy.io (Golden Retriever, Informer, Status Bar docs), get.bandcamp.help
(lossless, upload guides), elektronauts.com + kvraudio.com threads,
support.distrokid.com (formats, rejected releases), support.tunecore.com
(upload errors), support.google.com/photos + learngooglephotos.com +
complaint threads, dropboxforum.com collision thread, support.apple.com
iCloud Drive upload guide.
