# Brief: uploader UI copy (Pass B, U10/U11)

Finalizes `docs/superpowers/research/passb-copy-deck.md` into the copy contract at
`docs/superpowers/research/passb-copy-final.md`. The deck's 7 revision items get folded;
`queued`, `stalled`, the "What can I upload?" destination, the 401 state, and the
client-side pre-flight get first drafts.

## Verifiable facts

- Library policy: lossless FLAC, CD quality or better (16-bit/44.1kHz floor), enforced
  server-side by `internal/gate`; the page's spec text is generated from `/api/config`,
  never hand-written.
- Size cap: `Upload.MaxAlbumBytes` from config, rendered via `/api/config` (the mock's
  "2 GB" is an example value, not copy).
- Site name: enters through config (`Serve.SiteName`, new field per conductor ruling
  2026-09-01) and `/api/config`; "Wright Family Music Server" is this site's value,
  never a literal.
- Admin contact: `Config.AdminEmail`; copy says "the admin", the page may show the
  address.
- Cover art: optional; used when present, fetched by the pipeline otherwise
  (fetchart/embedart); non-audio files are never inspected by the gate.
- Album is the verdict unit; a partial album never enters; nothing uploads before the
  at-selection pre-flight passes.
- Rejection copy is the `internal/reject` wire contract (code/friendly/next_step, plus a
  short `label` per conductor ruling), shared verbatim with rejection emails; per-file,
  singular by design.
- Closing the page pauses an upload; returning resumes (tus fingerprint + stored URL).
  A 401 mid-flow means the Access session expired; reload re-authenticates and the
  upload resumes.
- A confident MusicBrainz match releases the user before import/art/scan finish; the
  page polls every 5s, so the spine live-updates to the in-library card without reload.
- Private Recordings (config-named): no matching, user tags kept as entered, its own
  Navidrome section; for music MusicBrainz doesn't know.
- `queued` has two causes: the contributor's own earlier session, or the import lock
  held by another run.
- `stalled` is user-visible; the admin notification it references is real (U6 sends it,
  ledger-deduped).
- iPadOS multi-select hides behind the picker's ⋯ → Select gesture; folder-select is
  absent on iOS; ZIP selection works.

## Audience questions

A family member on a phone or iPad, not technical, mid-upload or checking back later:
What can I upload? What's happening right now? Did it fail, and what do I do about it?
Can I close this page? Where is my music now?

## The one next step

Every terminal or error state names exactly one action (retry per path, pick the
private lane, add to MusicBrainz, try another album, reload, or nothing to do).

## Container plan

DaisyUI containers per the ratified mockups: masthead + subhead; format-requirements
panel (card); drop target; buttons (`btn`, short verb phrases); per-track chips
(1 to 3 words); steps spine labels (1 to 3 words) + sublabels (one short clause);
alert heading (one line) + body (1 to 2 sentences); history rows (title + short
status chip); toast/note lines (one sentence).

[ASK: none open — items 4, 5, and 7 are settled by conductor ruling in the final deck;
Geoff review rides the U13 before/after gate.]
