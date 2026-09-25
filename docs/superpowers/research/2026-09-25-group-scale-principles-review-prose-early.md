# Group-scale principles spec: early prose and fact review

A `prose-voice-reviewer` read of revision 1 (`1d6bc6a`), run before the four adversarial lenses
and saved by the conductor as an input to the fold. The final prose review runs again after the
fold. Line numbers refer to revision 1.

## Blockers (facts and contradictions)

- **B1, :119-120.** "the documented API admits music only through intake" is false. The
  configuration homes spec says a mirror "consumes disk with no intake verdict per album"
  (CH:621-622) and lists "Mirrors bypass per-album intake verdicts" as an accepted risk (CH:653).
  :114-116 omits the laptop's contribute and hand-over verbs (CH:614-628). Proposed: the uploader
  and contribute admit through intake one album at a time; a mirror lands a collection with no
  per-album verdict, bounded by the admin's approval of the member's first mirror and the quota;
  none of these paths fetches anything.
- **B2, :108-110, :115, table :212 and :218.** ROADMAP:113-116 says 0.9's client scope is
  unspecified; the configuration homes spec lists the TUI as unscheduled (~CH:708). The draft
  asserts 0.9 pairs through the device credential and that the mirror is "scheduled with the local
  client", and marks both "planned" against its own :198 definition. Proposed: say only that a
  local client pairs through the configuration homes device credential and the cap counts it;
  add a "designed, unscheduled" status and use it on both rows.
- **B3, table :205.** Marks "login required for all audio" as "exists" for the 0.8 surface, while
  the draft's own Conflicts item 1 says the streamer's Shares serve audio on a public token path.
  Proposed: "exists at rung 5; new for 0.8 once the shares erratum lands".
- **B4, :33-34 and :15-16.** Website copy states buy links and the wishlist as present fact; both
  are undecided or unscheduled, and the source question waits on ruling 2. Proposed: state only
  what ships, or mark planned items in the future tense.

## Warnings

- **W1, :172-174.** "services that credit artists and help listeners find them" is unsupported,
  and scrobbling routes no money. Proposed: "Scrobbling is planned" on the streamer's 2.1 list,
  or move the paragraph.
- **W2, :231-252.** The erratum inventory misses four Shares sites: streamer:507 (hook 15's share
  rows), streamer:42 (revision 2 ruling 12, "share management"), streamer:384 (endpoint list),
  CH:705 (`createShare` in the projector test list). Item 3 is called an addition while :231 says
  all three "conflict". Item 2 drops the playlist zip and `.m3u8`, which the streamer adopted
  "matching Navidrome" and its DAP section leans on (streamer:408); add "unless the study rung
  finds a roster client that depends on the playlist zip".
- **W3, table :208 and :213.** The build spec has no "principles" section; use its "Out of scope"
  (build:997) as a new entry.
- **W4, :143-144 and :218.** The admin approves a member's first mirror (CH:621), not each
  collection's; record the source answer when the member creates each mirror.
- **W5, :170.** "tells the member who added it" is ambiguous; IDEAS.md says intake "tells the
  asker".
- **W6, :26.** "A family is the typical group" breaks the group-not-family ruling; use "a group of
  people who know each other, such as a family or a few close friends".
- **W7, :31, :40-41, :43-44.** Jargon for a non-developer ("tags every album against MusicBrainz",
  "open a collection", "stream one track at a time"); plain rewrites proposed.
- **W8, :141.** Setup-colon payoff; split into two sentences.
- **W9, :122-124.** Aphorism ending in "a shared shelf", a coined metaphor the plain-terms ruling
  bars; state the point directly.
- **W10, :62-63.** The "can't stop a determined operator" caveat appears a third time; cut or
  fold.

## Suggestions (optional)

- :155-156: cite the streamer's "Per-user state" section (streamer:309) for the stream-start
  event, not hook 4.
- :101: the source has members mint a named app password from a device page, not at sign-in.
- :104, :127: state that both limits are settings, overridable per member, and whether sensitive.
- :239: give dates rather than "ratified later" (streamer revisions 2.3 and 2.4, 2026-09-24,
  postdate configuration homes' ratification, 2026-09-23).
- :160: the heading "Money points back at artists" is slightly cute; "Members are pointed toward
  buying".
- Table :223: "Ruling 5" isn't a place; "The first web album page (ruling 5)".
- :265-268: "one click" repeats in ruling 2.
- :28: 0.8 and 0.9 are FLAC-only; say "lossless (FLAC) downloads from stores such as Bandcamp or
  Qobuz".
- Repeated "no X, no Y, and no Z" triads and bold lead-ins: optional.

tellgrader (docs register): 0 findings.
