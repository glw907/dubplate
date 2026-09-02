# Uploader copy contract (Pass B final)

Status: drafted 2026-09-01 from `passb-copy-deck.md` (all 7 revision items folded),
brief at `docs/content-briefs/uploader-copy.md`. Independent critic pass folded
2026-09-01 (8 of 9 findings; the ninth asked for a review-state screen, which the
custody ruling removed from this surface; `pushed back` in history opens the
no-match copy in section 5, which carries the fix paths). Content-review gate:
PASS 2026-09-01 (fresh context, all four hard gates clear). This file is the string authority for
U10/U11; the mockups remain layout reference only. `{curly}` tokens are values injected
from `/api/config` or session data, never literals. Rules carried from the deck: no em
dashes; no "This [thing] couldn't be [verbed]" shape; engineering state names never
appear in copy; Private Recordings explains itself wherever it is offered; no-match
copy gives concrete fix paths.

## Conductor rulings folded here (plan corrections)

1. **`Serve.SiteName`** joins config and `/api/config` (U1's escalation rule fired: the
   masthead needs a config-driven name and the inventory had none). Non-secret; rides
   the env template as `DUBPLATE_SERVE_SITE_NAME`. Lands in U9.
2. **Reject wire format gains `label`**: a 1-to-3-word chip string per reason,
   `internal/reject.Label()`, added to `reasonWire` and the golden tests. Keeps U10's
   per-track chips API-fed without a client lookup table. Lands in U9 (backend half)
   or U10's first dispatch; backward-compatible field add.
3. **Taxonomy amendment**: `ReasonUnpackFailed.Friendly()` becomes "This archive
   wouldn't unpack as a ZIP." (the shipped "This archive couldn't be unpacked." is the
   banned passive shape). Email template rides along via the existing cross-asset test.
4. **Item 7 ruling (importing step)**: no fifth step. After a confident match, step 3
   completes and step 4 ("In your library") becomes current with sublabel "adding it
   now". The 5s poll live-transitions the page to the in-library card; no reload.
5. **Item 2 ruling (singular/plural)**: taxonomy strings are per-file by contract and
   stay singular; the count-aware sentence is the client-composed alert body. The
   reason hint (Friendly + NextStep) renders once per distinct failing reason, visually
   attached to the flagged rows.

## 1. Idle / selection

Desktop:
- Masthead: `{site_name}` / subhead: "add music to the library"
- Format panel: "This library is **lossless FLAC**, CD quality or better. Upload a
  whole album: a .zip or its files, up to **{max_album_size}**. Cover art is
  optional. Include one if you have it; the library finds one otherwise." *(one bold
  fact per sentence, critic finding 9)*
- Drop target: "Drop an album anywhere on this page" / "or"
- Primary button: "Choose files"
- Hint: "One album at a time. Albums go in whole."

iPad / iPhone:
- Subhead: "add music"
- Format panel: "From a phone or iPad, a **.zip** of the album is easiest. Picking the
  files works too. Cover art is optional; the library finds one if you don't include
  it." *(item 6 folded: the art line joins the mobile panel)*
- Primary button: "Choose from Files"
- Hint: "To pick several songs at once: tap **⋯** in the picker, then **Select**."
- Secondary button: "What can I upload?" opens the detail sheet:
  - "Lossless FLAC, CD quality or better. MP3, AAC, and other lossy files can't go in."
  - "Whole albums only, up to {max_album_size}. Send a .zip of the album, or pick the
    files themselves."
  - "Cover art is optional. If the album comes with art, it's used. If not, the library
    finds a cover on its own."
  - "The checks only look at audio files. Booklets and other extras ride along."

Collection picker (renders only when config lists more than one collection):
- Label: "Collection"; options are the configured collection names verbatim. Selecting
  the private collection swaps the format panel for the private-lane contract panel
  (section 6) and shows the tag form before upload can start.

## 2. Uploading

- Per-track chips: "looks good" (passed) / "sending" (in flight) / "waiting" (not yet
  started) / "stopped" (a connection failure only, transfer failed, awaiting retry;
  ratified post-build 2026-09-01: a failed row must never read as not-yet-started). A
  server-side rejection with no parseable reason chips section 4's own label fallback,
  "didn't pass", instead of "stopped" (ratified post-build 2026-09-01, U13 fold F6.8): a
  permanently refused file must never read as a paused one.
- Progress readout: "{acked_bytes} confirmed" / "of {total_bytes}"
- Hint: "Switching away pauses the upload; coming back resumes it. If the page
  closes, pick it up from Your uploads." *(critic finding 2: a killed tab loses the
  File object, so silent resume-on-return is only true for backgrounding; the
  history path is the honest route for a full close)*
- Button: "Keep my screen awake" (shown only where Wake Lock is supported)
- ZIP path, stated before upload starts: "We check the tracks once the zip finishes
  uploading."
- Slow-connection downshift note: "Slow connection. Switching to smaller pieces so
  progress sticks."
- Retry budget exhausted (client-side stall, manual retry): heading "Upload stopped";
  body "It stopped moving, but nothing is lost. Check your connection and try again;
  it picks up where it left off." Button: "Retry".
- Signed-out (401 or HTML-where-JSON): heading "Signed out"; body "Your sign-in
  expired. Reload to sign in again, then pick the upload back up from Your uploads."
  Button: "Reload". *(critic finding 2: reload loses the File object; resume runs
  through history plus re-selection, never silently)*

## 3. Journey spine (DaisyUI steps, four steps)

| Step | Label | Sublabel |
|---|---|---|
| 1 | "Uploaded" | "all {n} tracks received" |
| 2 | "Checked" | "every file is CD-quality FLAC" |
| 3 | "Matching" | "checking MusicBrainz, just a moment" |
| 4 | "In your library" | pending: "ready to play" / current: "adding it now" |

- On confident match, below the spine: "Matched: {artist} - {album} ({year}). You can
  close this page now."
- In-library card (live transition, no reload): heading "In your library"; body
  "{artist} - {album} is ready."; primary button "Open in Navidrome"; secondary button
  "Upload another album" (ratified post-build 2026-09-01, U13 fold F6.11).
- ZIP-path checking sublabel (step 2 current): "unpacking and checking the tracks"
- If checking itself fails to finish -- a network drop or server error probing the
  outcome, distinct from a MusicBrainz-specific "no match" -- (ratified post-build
  2026-09-01, U13 fold F6.9): heading "Couldn't finish checking"; body "Something went
  wrong finishing the check. Your album is already uploaded and waiting; no need to
  send it again." Button: "Try again".

## 4. Rejected (format failure)

- Alert heading: "Couldn't add this album"
- Alert body, client-composed, count-aware. All failures share one reason: "{n} tracks
  are {label}, and albums go in whole." (singular: "One track is {label}, and albums
  go in whole.") Mixed reasons: "{n} tracks didn't pass the checks, and albums go in
  whole."
- Per-track chips: "fine" (passed) / `{label}` from the wire format (failed)
- Reason hint under the flagged rows: `{friendly}` then "**Next:** {next_step}", both
  verbatim from the API.
- At-selection pre-flight (nothing uploaded yet): same composition plus the closing
  sentence "Nothing has uploaded yet." Multi-file path retry re-sends only the failing
  files; ZIP path retry copy: "Retrying re-uploads the zip."
- Primary button: "Try another album"
- Fallback body when the wire carries no detail (ratified post-build 2026-09-01):
  "The album didn't pass its checks."
- Label fallback when a file's own wire verdict carries no `label` at all (ratified
  post-build 2026-09-01, U13 fold F6.8): "didn't pass", `badge-error`.

Chip labels (the new `label` field; authority is `internal/reject`):
bad-extension "not FLAC" / symlink "shortcut file" / corrupt "damaged in transit" /
below-cd-quality "below CD quality" / too-large "too big" / unpack-failed "wouldn't
unpack" / unsafe-path "unsafe file path" / no-space "library is full" /
too-many-sessions "too many uploads"

## 5. No MusicBrainz match (item 1 folded)

- Alert heading: "No MusicBrainz match"
- Body, para 1: "MusicBrainz didn't find a match for this album, so it didn't go into
  the main library. You can upload it to **Private Recordings** instead. That
  collection is for music MusicBrainz doesn't know, and it shows up in Navidrome as
  its own section."
- Body, para 2: "If it's a released album, a fix usually works: check the artist and
  title spelling and try again, or add the release to MusicBrainz first."
- Buttons: primary "Upload to Private Recordings"; secondary "Fix the names and
  retry"; secondary "Add it to MusicBrainz".
- Probe unavailable (MusicBrainz down): heading "Can't check MusicBrainz right now";
  body "The matching service isn't answering. Your album is already uploaded and
  waiting; no need to send it again. Try again in a little while." Button: "Try
  again". *(critic finding 1: this state renders post-upload, post-check; the
  session holds pre-promotion per U5, but the bytes are on the server)*

## 6. Private Recordings lane (item 3 folded)

- Masthead: "Private Recordings" / subhead: "your tags, its own section"
- Contract panel: "No MusicBrainz matching here. Private Recordings is for music
  MusicBrainz doesn't know: this album keeps **the tags you enter** and goes into its
  own section in Navidrome, separate from the main library."
- Tag form label: "Tag this album"; track rows labeled "{nn} (from filename)"
- Primary button: "Upload to Private Recordings"

## 7. Queued (item 4 folded)

- Alert heading: "Waiting its turn"
- Own earlier session: "Your earlier album is still importing. This one starts as soon
  as it finishes."
- Library busy (another run holds the lock): "The library is finishing another import.
  This one starts right after."

## 8. Stalled (item 5 folded)

- Alert heading: "Taking longer than it should"
- Body: "Something behind the scenes is slow. The admin gets notified automatically
  when this happens, and this page keeps checking on its own. Nothing needs to be
  redone." *(critic findings 4 and 7: present-tense system behavior instead of a
  completed-event claim that can be false at first render; no bare "safe" assurance)*

## 9. History ("Your uploads")

- Section label: "Your uploads"
- Status chips by state: uploading "uploading · resume" (with resume affordance) /
  checking "checking" / matching "matching" / importing "adding now" / in-library
  "in library" / pushed-back "pushed back" / rejected "wrong format" / queued
  "waiting its turn" / stalled "slow going" *(critic finding 6: distinct at a glance
  from importing's "adding now")* / review "needs a look" *(ratified post-build
  2026-09-01: admin-timer-path albums can still surface this ledger state in
  history even though the uploader flow never files review)*
- Empty state: "Nothing here yet."
- Reachability note (ratified 2026-09-01): the shipped ledger has no `matching`
  state (`tagging` renders as "adding now"), so the "matching" chip is reserved
  vocabulary until the ledger carries it; not a defect.
- In-flight resumables list first.

## Out of scope here

The roster-absent warm page shipped in Pass A (`warmRosterAbsentPage`); U13's
adversarial verbiage review covers it alongside everything above. Rejection email
wording is the shared taxonomy contract and changes only via ruling 3.
