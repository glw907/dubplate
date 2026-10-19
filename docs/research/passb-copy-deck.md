# Dubplate uploader copy deck

Source: `uploader-mockups.html` (ratified layout reference) cross-checked against
`docs/superpowers/specs/2026-08-31-dubplate-uploader-design.md`, including the
2026-09-01 journey addendum, which supersedes the spec's original journey states
and review path.

## Rules every string in this deck must obey

1. **No em dashes anywhere.** Use a period, comma, or a plain hyphen for a
   title/date pairing (e.g. "The Feelies - Only Life").
2. **Copy must not read as AI-written.** The ratified example of a banned tell:
   *"This album couldn't be taken in."* Any string built on that same
   passive "This [thing] couldn't be [verbed]" shape is suspect, not just that
   literal sentence.
3. **Spec state names are engineering labels and never appear in copy.**
   `pushed-back`, `in-library`, `no-match`, etc. (hyphenated, code-shaped) must
   never leak into user-facing text. Render them as ordinary English instead
   ("pushed back", "in library").
4. **Private-lane copy must explain what Private Recordings is**: a separate
   collection for music MusicBrainz doesn't know, distinct from the main
   library.
5. **MusicBrainz no-match copy must give concrete fix paths**, including
   adding the release to MusicBrainz itself, not just "try again."

Journey state vocabulary, final (spec addendum): `uploading`, `checking`,
`matching`, `importing`, `in-library`, `pushed-back`, `queued`, `stalled`.
Private-lane sessions skip `matching`. These are engineering names only; the
column below maps each to its mocked or spec-described user-facing rendering.

---

## 1. Selection / format requirements (idle screens)

### Screen 1 - Desktop idle
| Component | Text |
|---|---|
| Masthead | "Wright Family Music Server" |
| Masthead subhead | "add music to the library" |
| Format-requirements panel | "This library is **lossless FLAC**, CD quality or better. Upload a whole album: a **.zip** or its files, up to **2 GB**. Cover art is optional. Include one if you have it; the library finds one otherwise." |
| Drop target | "Drop an album anywhere on this page" / "or" |
| Primary button | "Choose files" |
| Hint | "One album at a time. Albums go in whole." |

### Screen 2 - iPad and iPhone selection
| Component | Text |
|---|---|
| Masthead subhead | "add music" (shorter than desktop's "add music to the library" - mobile-width variant) |
| Format-requirements panel | "From a phone or iPad, a **.zip** of the album is easiest. Picking the files works too." |
| Primary button | "Choose from Files" (capitalized "Files" - names the iOS Files app; desktop's parallel button says "Choose files", lowercase, no app reference) |
| Hint | "To pick several songs at once: tap **⋯** in the picker, then **Select**." |
| Secondary button | "What can I upload?" (opens fuller format-requirements detail; the destination copy is not shown in the mockup) |

**Album-art variant note**: the format panel states cover art is optional and
self-supplies ("the library finds one otherwise") only on the desktop screen.
The iPad/iPhone screen's shorter panel omits the art line entirely; if the
mobile panel is meant to carry the same guidance, that's a gap to close before
build, not a contradiction to preserve as a legitimate variant.

---

## 2. Uploading

### Screen 3 - Uploading, per-file verdicts (phone)
| Component | Text |
|---|---|
| Album heading | "The Feelies - Only Life" (artist-title separator is a plain hyphen, not an em dash) |
| Per-track verdict tag, passed | "looks good" (used per file; two instances shown, one per passing track) |
| Per-track verdict tag, in flight | "sending" (the file actively transferring) |
| Per-track verdict tag, queued locally | "waiting" (a file not yet started) |
| Progress readout | "112 MB confirmed" / "of 244 MB" |
| Hint | "Closing this page pauses the upload. It resumes when you return." |
| Secondary button | "Keep my screen awake" |

Variant: the three per-track tags ("looks good" / "sending" / "waiting") are a
single small vocabulary reused across every track row; no singular/plural
inflection needed since each tag describes one file.

---

## 3. Checking / matching / importing / in-library (the journey spine)

### Screen 4 - "The journey" (phone, DaisyUI `steps`)
| Step | Label | Sublabel |
|---|---|---|
| done | "Uploaded" | "all 10 tracks received" |
| done | "Checked" | "every file is CD-quality FLAC" |
| now (current) | "Matching" | "checking MusicBrainz. Stay a moment" |
| pending | "In your library" | "ready to play" |

Below the steps: "Matched: The Feelies - Only Life (1988). You can close this
page now."

Notes:
- Labels are natural-language ("Uploaded", "Checked", "Matching", "In your
  library"), never the hyphenated engineering names ("uploading",
  "checking", "matching", "in-library") - this already follows rule 3
  correctly and should stay the model for any new label.
- Per the journey addendum, a confident match releases the user before
  import/art/scan finish; "In your library" is only supposed to render once
  the link is real. This mock shows the step as still-pending at match time,
  which is consistent, but no mocked screen shows the moment the page
  transitions from "you can close this page" to the finished in-library
  card (screen 6) without the user having reloaded. Flagged below.
- The spine has four visual steps but the addendum's vocabulary has five
  states before in-library (`uploading`, `checking`, `matching`,
  `importing`, `in-library`). "Importing" has no dedicated step in this
  mock; it's implied to happen silently between "Matching" and "In your
  library" per the addendum's "import/art/scan finish alone" language. This
  is a legitimate design choice (not every engineering state needs a visible
  step), not a gap, but worth confirming at build time that the spine text
  doesn't imply "In your library" is reachable directly from "Matching" with
  nothing in between.

### Screen 6 - "In your library" card
| Component | Text |
|---|---|
| Alert heading | "In your library" |
| Alert body | "The Feelies - Only Life is ready." |
| Primary button | "Open in Navidrome" |

---

## 4. Pushed-back / rejected (format failures, in-page)

### Screen 5a - Rejected (format failure, phone)
| Component | Text |
|---|---|
| Alert heading | "Couldn't add this album" |
| Alert body | "Two tracks are MP3, and albums go in whole." |
| Per-track tag, passing | "fine" |
| Per-track tag, failing | "MP3 file" (used on both offending tracks) |
| Hint / next step | "This is an MP3 file, and the library keeps lossless FLAC." / "**Next:** find the album as FLAC and upload it again." |
| Primary button | "Try another album" |

**Singular/plural mismatch, flagged below**: the alert body correctly says
"Two tracks are MP3," but the hint underneath reads "This is an MP3 file"
(singular), even though two tracks failed. See Needs Revision.

---

## 5. Matching / no-match (MusicBrainz)

### Screen 5b - No MusicBrainz match (phone)
| Component | Text |
|---|---|
| Alert heading | "No MusicBrainz match" |
| Alert body, para 1 | "This album couldn't be matched, so it wasn't added to the main library. You can upload it to **Private Recordings** instead. That collection is for music MusicBrainz doesn't know, and it shows up in Navidrome as its own section." |
| Alert body, para 2 (fix paths) | "If it's a released album, a fix usually works: check the artist and title spelling and try again, or add the release to MusicBrainz first." |
| Primary button | "Upload to Private Recordings" |
| Secondary button | "Fix the names and retry" |
| Secondary button | "Add it to MusicBrainz" |

This screen is the one place in the mockup set that satisfies rule 4 (explains
what Private Recordings is) and rule 5 (concrete fix paths, including adding
the release to MusicBrainz) in full. It's also the closest string to the
banned-tell pattern in the whole set. See Needs Revision.

---

## 6. Private Recordings lane

### Screen 5c - Private Recordings lane (phone)
| Component | Text |
|---|---|
| Masthead | "Private Recordings" |
| Masthead subhead | "your tags, its own section" |
| Contract panel | "No MusicBrainz matching here. This album keeps **the tags you enter** and goes into **Private Recordings**, its own section in Navidrome, separate from the main library." |
| Tag form label | "Tag this album" |
| Tag rows (example values, not copy) | Artist: "Duncan W.", Album: "Spring Recital 2026", Year: "2026", "01 (from filename)": "Prelude in C" |
| Primary button | "Upload to Private Recordings" |

Cross-reference to rule 4: this screen's own panel text says "No MusicBrainz
matching here" and "separate from the main library" but doesn't restate "for
music MusicBrainz doesn't know" the way screen 5b does. In the actual flow a
user only reaches 5c after 5b (or after choosing the lane up front), so the
explanation has usually already been read once. Still worth tightening so the
screen is self-sufficient on its own. See Needs Revision.

---

## 7. Queued (not mocked as a screen)

No dedicated screen exists for `queued`. The mockup's closing note describes
it in prose only, as the state where a second upload by the same contributor
(or someone else's import) is ahead in line:

> "the 'waiting on your other upload' state (either 'your earlier album is
> still importing' or 'the library is finishing someone else's import')"

Two candidate sub-strings live inside that note:
- "your earlier album is still importing" (own prior upload is ahead)
- "the library is finishing someone else's import" (another contributor's
  import is ahead)

These are drafting notes, not ratified screen copy - no frame renders them.
Flagged below as needing an actual screen before U9, since `queued` is one of
the eight vocabulary states in the spec addendum.

---

## 8. Stalled (not mocked as a screen)

No dedicated screen or note describes `stalled` at all, not even in prose.
The closest analogue in the mockup set is the history-list status on screen 6:

| Component | Text |
|---|---|
| History row status | "uploading · resume" (attached to "New Order - Power…", a title truncated by the mock's layout, not ratified copy) |

This reads as an interrupted-but-resumable upload, which may or may not be
what `stalled` is meant to cover (versus a queued session, versus a genuinely
stuck import). Flagged below: `stalled` has no ratified user-facing string at
all.

---

## 9. History list (own-uploads ledger)

### Screen 6 - history rows
| Album | Status shown |
|---|---|
| "The Feelies - Only Life" | "in library" |
| "Aim - Hinterland" | "pushed back" |
| "Road Trip Mix" | "wrong format" |
| "New Order - Power…" (truncated in the mock, not ratified) | "uploading · resume" |

Section label above the list: "Your uploads" (uppercase small-caps styling,
not a literal casing instruction).

Notable compliance with rule 3: "in library" and "pushed back" render the
engineering states `in-library` and `pushed-back` as plain two-word English,
un-hyphenated. This is the correct pattern and should be the template for any
new status string using those two states.

---

## 10. Roster gate (not mocked as a screen)

Mentioned only in the mockup's closing note and in the spec:

- Mockup note: "the warm not-on-the-roster page ('ask Geoff to add you')"
- Spec: "An email not in the roster gets a warm 'ask the admin to add you'
  page."

The mockup's version names Geoff directly (site-specific instantiation for
the Wright Family Music Server); the spec's version is the generic
config-driven default ("the admin"). Both are consistent with the
adopter-goal rule that site specifics enter through config only - the site
copy substitutes the configured admin name/contact for "the admin." Not
flagged as a defect; noted so the eventual build doesn't treat "Geoff" as a
literal hardcoded string.

---

## Needs revision before U9

1. **Screen 5b, alert body para 1: "This album couldn't be matched, so it
   wasn't added to the main library."**
   This is structurally the banned-tell pattern itself: "This [thing]
   couldn't be [verbed]," the same shape as the ratified banned example
   ("This album couldn't be taken in"). It's the single highest-risk string
   in the deck.
   *Proposed rewrite*: "MusicBrainz didn't find a match for this album, so it
   didn't go into the main library. You can upload it to **Private
   Recordings** instead. That collection is for music MusicBrainz doesn't
   know, and it shows up in Navidrome as its own section."
   (Keeps the required Private Recordings explanation intact; states the
   actor - MusicBrainz - and the actual event, instead of a passive
   "couldn't be" construction.)

2. **Screen 5a, hint line: "This is an MP3 file, and the library keeps
   lossless FLAC."**
   Singular ("This is an MP3 file") contradicts the alert above it ("Two
   tracks are MP3"). Looks like a generic per-reason template applied once
   regardless of how many files triggered it, which will read as broken to a
   contributor who scrolls up and sees two flagged tracks.
   *Proposed rewrite (plural-aware)*: "These two files are MP3, and the
   library keeps lossless FLAC only." / **Next:** find the album as FLAC and
   upload it again.
   For the single-file case, keep the original: "This file is MP3, and the
   library keeps lossless FLAC only." / **Next:** find the album as FLAC and
   upload it again.
   Needs an actual singular/plural branch in the reason copy, not one fixed
   string reused for both counts.

3. **Screen 5c panel doesn't restate what Private Recordings is.**
   "No MusicBrainz matching here. This album keeps the tags you enter and
   goes into Private Recordings, its own section in Navidrome, separate from
   the main library." never says *why* - that Private Recordings exists for
   music MusicBrainz doesn't know. A contributor who lands here directly
   (lane chosen up front, never having seen 5b) gets the mechanism but not
   the reason.
   *Proposed rewrite*: "No MusicBrainz matching here. Private Recordings is
   for music MusicBrainz doesn't know: this album keeps **the tags you
   enter** and goes into its own section in Navidrome, separate from the
   main library."

4. **`queued` has no ratified screen, only a drafting note.**
   The mockup's closing note ("the 'waiting on your other upload' state...")
   is explicitly marked as "not shown" and its two sub-strings read as
   internal shorthand rather than screen copy: "your earlier album is still
   importing" and "the library is finishing someone else's import" are
   plausible but unstyled, un-framed, and never placed in a component.
   *Proposed direction*: a screen matching the shape of screen 4 (steps) or
   screen 5-series (alert), naming which of the two sub-cases applies, e.g.
   an alert headed "Waiting on another upload" with body either "Your
   earlier album is still importing. This one starts as soon as it
   finishes." or "Someone else's import is finishing first. This one starts
   right after." Needs an actual mock before copy can be finalized, not just
   a proposed sentence.

5. **`stalled` has no ratified copy anywhere, not even in a drafting note.**
   Nothing in the mockup or its closing note addresses this state, despite
   it being one of the eight names in the spec addendum's final vocabulary.
   *Proposed direction*: needs its own screen/alert before U9. At minimum,
   scope the question the copy has to answer: is `stalled` contributor-facing
   ("this upload has stopped moving, here's what to do") or closer to a
   monitoring-only condition that never renders? The spec says absence of a
   run report "is a monitoring event, not a journey state," which suggests
   `stalled` is meant to be user-visible and distinct from that - but the
   copy deck can't specify wording for a state nobody has drafted a sentence
   for yet.

6. **iPad/iPhone format panel (screen 2) drops the album-art guidance.**
   Desktop's panel (screen 1) tells the contributor cover art is optional and
   self-supplied if missing; the mobile panel (screen 2) has no equivalent
   line. If this is deliberate (mobile panel trimmed for space), fine, but it
   should be a stated decision, not a silent gap - a mobile contributor
   currently has no way to learn the art rule from this screen (the "What can
   I upload?" button may cover it, but that destination copy doesn't exist
   yet either).
   *Proposed rewrite (if space allows)*: append to the existing panel: "Cover
   art is optional; the library finds one if you don't include it." If space
   doesn't allow, the "What can I upload?" destination must carry it instead,
   and that destination needs to be drafted before U9.

7. **Screen 4's spine has no visible "importing" step, and no mocked
   transition out of "you can close this page now."**
   Not necessarily wrong (see note in section 3), but nothing in the mockup
   set shows what a contributor sees if they don't close the page and instead
   wait through import/art/scan. Confirm before build whether that's meant to
   be silent (spine just sits at "Matching" until the page is closed and
   reopened to screen 6) or whether it needs its own transitional copy.
