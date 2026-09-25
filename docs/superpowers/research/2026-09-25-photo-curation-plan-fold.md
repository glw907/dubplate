# Photo curation round one: plan fold record

Fold of the three round-one plan reviews (contract `PC`, mechanics `PM`, domain risk `PD`) into
`docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`, against the spec
`docs/superpowers/specs/2026-09-25-photo-curation-design.md`, now revision 2.3. Folder: Opus 5.5,
2026-09-25, from commit `b843519`. Each finding was checked against the tree, the tools, or the
review's own probe before it was folded. A finding is folded when it prevents real harm: a wrong
license reaching the repo, a pick landing on the wrong image, a broken command, a blown ceiling,
or a Wikimedia block. It is refused when its fix costs more than the risk it removes.

## Rulings built in

- **PD7, people and minors (Geoff, 2026-09-25): "Flag, you decide."** Agents flag a Commons
  personality-rights template or an apparent minor in focus in a new candidate `flags` field
  (closed list in `vocab.json`), the page shows it, and Geoff's keep or pass decides. Spec:
  Process paragraph and the third candidate-only field in Storage (stripped at commit like
  `screen`). Plan: Owner rulings, Task 1 vocabulary and candidate record, Task 2 outcome, Task 4.
- **PM9, approvals (Geoff, 2026-09-25): "Batch, no pruning."** Spec: the pruning sentence in
  Process replaced with the ruling. Plan: Owner rulings, Task 6 round open (one announced burst
  of about 12 upload approvals; nothing deleted from the store).
- **Spec status** bumped to revision 2.3, naming both rulings and this record. The Storage
  sentence "two fields of their own" became "three" so the spec does not contradict its own
  Process paragraph; that is the only edit outside the ruling lines and status.
- **Conductor rulings** applied as given: Task 0 marked done at `fae78c3`; PD5 downgraded; PM6
  close agent; PM8 fallback store, headless render, and `test` round; PM2 dispatch text;
  PD1 and PD2 on the letters.

## Token ceiling

Raised from 3.5M to **4.5M** (80% flag at 3.6M). PD6 estimates 2.5M to 3.3M for the plan as
written, already 94% of 3.5M at its top end, so the 80% flag would fire inside Task 4. The fold
then adds work: a fourth `check.mjs` mode and a fixture table roughly twice as long (Task 1), a
render agent and a functional pass (Task 2), a seventh search agent from splitting research-d
(PM17), a round-open agent, and a close agent (PM6). By the same per-agent rates that is about
0.5M to 0.7M more, for about 3.0M to 4.0M. 4.5M keeps the flag above the expected top end.
STATUS's lane row still says 3.5M; the conductor updates it.

## Contract lens (PC)

| ID | Disposition |
|---|---|
| PC1 | Folded, merged with PM1: `node --test tools/photos/check.test.mjs` in Global constraints and Task 1's gate. |
| PC2 | Folded, merged with PM7: Task 1 Interfaces (exit 0, 1, 2; no JSON on 2) and the empty-pool acceptance (`0 of 0`, `short`, exit 1; integrity exit 0). PC2's three-code form taken over PM7's because it separates a crash from a report. |
| PC3 | Folded, merged with PM5: `balance --json` emits the derivation maps and picked keys; `sheet` adds derived fields and the second-image flag against pool and candidates; a post-MBID duplicate goes back to Geoff unpicked, and no unmade waiver is written (Task 6 close). |
| PC4 | Folded, merged with PM4: the round, candidate, and choice documents in Task 2's outcome; `closed` written only from the conductor's side; reopening clears it. |
| PC5 | Folded, merged with PM8 per the conductor's ruling: 150-candidate `fixture:` render at desktop and 390px, closed-state refusal, and the `test` round deleted after (Task 2 acceptance). |
| PC6 | Folded, merged with PD12: agents write a 1600px `preview.webp`; `Space` opens it; it is the encode round close commits (Task 2, Task 4, Task 6). |
| PC7 | Folded: the candidate record's field list, `source_width`/`source_height`, sibling hash, thumbnail and preview presence, size floor on the source edge, and `pack` iff pack class (Task 1). The fixtures are in the table. |
| PC8 | Folded into `sheet` (merged with PM5): cross-agent dedupe on `source_id`, `origin_id`, `source_sha256`, repeats listed as `duplicate`. |
| PC9 | Folded: Flickr evidence is the photo model's `"license":<n>` with the rendered URL, mapped through a `vocab.json` table with its source URL; never `license-models` (Global constraints; Task 1 vocabulary). |
| PC10 | Folded: one table-driven case per rule, the three waiver cases, and a `.candidates/` fixture integrity must ignore (Task 1 acceptance). Cheap in `node:test`, and the waiver is the script's only owner-authority field. |
| PC11 | Folded: `vocab.json` populated from the research files' cited licenses plus the shipped classes; reason codes gain `unavailable` and `vocab`; an off-list term drops with `vocab` and is proposed in the report (Global constraints; Task 1). |
| PC12 | Folded: seeded, logged sample; whole-round re-verify per the spec; `changed-since-grant` result (Task 5). The reader draws with the conductor's seed so the conductor stays thin. |
| PC13 | Folded, merged with PD13: unsorted candidates carry over; mismatches logged; slug-unused refusal; MBID only on one fitting match (Task 6 close). |
| PC14 | Folded: fetch `Template:<name>` raw and quote its license line (Global constraints; review focus 1). |
| PC15 | Folded, merged with PM14: the conductor dispatches `code-simplifier` after Task 1's accept, re-runs the gate, and commits (Task 1). |
| PC16 | Pack hiding folded (`sheet` drops pack records; Task 2 outcome). Pruning superseded by Geoff's PM9 ruling. |
| PC17 | Folded: Task 2's conductor-first step confirms `db` and `assets` in the roster and stops for Geoff if either is missing. |
| PC18 | Folded: the round log lists committed `source_id`s beside the keep documents, and the sets are equal (Task 6 acceptance). |

## Mechanics lens (PM)

| ID | Disposition |
|---|---|
| PM1 | Folded, merged with PC1. |
| PM2 | Folded per the conductor's ruling: Task 1's dispatch text names `CAIRN_GATE_LANE=light cairn-run-gate 'node --test tools/photos/check.test.mjs'` (with the `cd` prefix the agent's workflow uses), replacing step 5, and the reviewer reads `git show <sha> -- tools/photos photos/meta`. The ts-conventions path is in the dispatch. Verification: `cairn-run-gate` has no `--help` flag. The fold's `cairn-run-gate --help` call treated `--help` as a gate string and queued a heavy-lane run behind lane 1's gate. It was stopped and its state directory (`6a2fd0842b4e6faa`) removed. The plan quotes the tool's header, lines 17 to 19, instead. |
| PM3 | Folded, merged with PD10: per-host `flock` locks, `curl --retry 5`, `--limit-rate 2M` on the media host, batched `imageinfo`, stop on repeated 429 or 403 (Global constraints). A Flickr lock is added on the same pattern. |
| PM4 | Folded, merged with PC4: candidates come from `db` documents, and thumbnails and previews are uploaded with URLs seeded from the upload results (Task 2, Task 6). The `files`-republish alternative was declined because it changes the page per round. |
| PM5 | Folded: `sheet` is Task 1's fourth mode, with the deliverable count stated at dispatch. |
| PM6 | Folded per the conductor's ruling: a Sonnet close agent does close steps 2 and 3 into `.candidates/_close/`; the conductor sets `closed`, rules on returns, and runs the commit call (Task 6). A Sonnet round-open agent is added on the same reasoning, which keeps the conductor from grinding 12 uploads and a seed batch. |
| PM7 | Folded, merged with PC2. |
| PM8 | Folded per the conductor's ruling: in-memory fallback on a `null` `db`, headless render by a subagent using `chromium-shot` and Playwright read-only from `web/node_modules` (1.63.0, confirmed present), then one conductor pass under round `test`, deleted after (Task 2). |
| PM9 | Settled by Geoff's ruling (batch, no pruning). The page-side prune button is not built. |
| PM10 | Folded: `candidate` enforces the `source_id` pattern that fits `ArtifactData`'s `doc_id` (Task 1). |
| PM11 | Folded: integrity reads the top level and `meta/` only and ignores `.candidates/`; `candidate` on a directory reports collisions with picked, rejected, and pack keys (Task 1). |
| PM12 | Folded: `add -N` and `commit --` in one Bash call, `git reset -q --` on failure, and retry on either lock message (Global constraints). |
| PM13 | Folded: `-limit memory 512MiB -limit map 1GiB` in the encode, temp files inside the agent's directory (Global constraints). The limits change memory use, not output, so the spec's pinned encode still holds. |
| PM14 | Merged with PC15. |
| PM15 | Folded: the conductor dispatches `prose-voice-reviewer` (Task 3); the Task 2 agent publishes, and the round-open agent runs `action: read` before its uploads (Task 6). |
| PM16 | Folded: the contact variable appears only on the curl command line and is never written or echoed (Global constraints). |
| PM17 | Folded: one fetch-and-encode script per agent, and research-d split into d1 and d2 (Task 4); the ceiling is raised to cover it. |

## Domain risk lens (PD)

| ID | Disposition |
|---|---|
| PD1 | Folded into the drafts per the conductor's ruling. The fold re-checked every route: each address still appears on the page the draft cites; the image and gallery URLs return 200; Beckman's Wikipedia article lists no death date; Lesser has no article, and no death record turned up. Each draft header records the check. `charliephillipsarchive.com` still has no MX record, so the Phillips draft and its README row are `blocked`, with the Akehurst fallback noted. The Lesser letter states no license fact about her archive, so the unverified NC-ND claim never reaches her. |
| PD2 | Folded: every letter says "CC BY lets anyone reuse and adapt a photo, commercially too, as long as they credit you" (or "credit it" for institutions), and describes the pack as planned, undated, and a separate optional download. All five letters are under 250 words counting the subject line, contain no em dash, and never say "non-commercial". |
| PD3 | Folded: deletion and no-permission template scan, the Questionable Flickr images check, and the old own-work plausibility rule (Global constraints; Task 5 applies them). |
| PD4 | Folded: the public-domain bases list, each with its required fact, and the two explicit non-bases (Task 1 vocabulary); the no-notice evidence rule is in the flyers brief (Task 4). |
| PD5 | Downgraded to minor per the conductor's ruling: `dubplate-implementer` gates before committing, and the scoped change set covers uncommitted work plus `HEAD~1`, so a lane 1 task is scanned before a later photo commit lands, and the merge ritual's full gate is unscoped. The cheap parts are folded: checkpoints go to the round log, STATUS changes only at lane open and close, and encoded images enter `photos/` only in the commit's Bash call (Global constraints; Task 6). The segment-boundary commit rule is refused as not needed. |
| PD6 | Folded: raw pages saved and grepped, never read whole; `imageinfo` sizes before download; per-agent spend reported; runaway guard at about 2MB per agent (Global constraints; Task 4). It drives the ceiling change above. |
| PD7 | Settled by Geoff's ruling (flag, you decide). |
| PD8 | Merged, already met by the drafts: the British Culture Archive letter asks "you and the contributors", and the DC letter asks who holds the rights and to pass the note along. No change. |
| PD9 | Folded: the letters' read is a second touchpoint in the same message as the sitting notice, and the upload approvals come as one announced burst (Task 6). |
| PD10 | Merged with PM3. |
| PD11 | Folded in part: a format allowlist checked with `magick identify`, with others dropped as `format` (Global constraints; new reason code). The explicit coder prefix is refused: the identify check already rejects a disguised PDF or SVG, and a per-file prefix adds a failure mode for no added safety. |
| PD12 | Merged with PC6. |
| PD13 | Merged with PC13. |
| PD14 | Folded into the flyers brief (Task 4). |
| PD15 | Folded: originals over about 40MB are taken as a named rendition of at least 1600px, with `file_url` naming it and `verified` noting what `source_sha256` covers (Global constraints). |
| N1 | Folded: Task 0 marked done at `fae78c3`. |
| N2 | Refused for this plan, because it is no round-one change. The conductor carries it to the release checklist: a public repo exposes withdrawn blobs in history. |
| N3 | Folded: Global constraints forbid staging outside the task's Files list. |

## Counts

53 findings across three reviews: 18 PC, 17 PM, and 15 PD plus 3 notes. Two findings (PD7, PM9)
are settled by owner rulings. 43 are folded, some in part (PC16, PD5 as downgraded, PD11). 7 are
merged into another finding's fold (PM1, PM7, PM14, PD8, PD10, PD12, PD13). 1 is refused (N2, routed
to the release checklist). Refused in part: PD5's segment-boundary rule, PD11's coder prefix, and
PM4's `files` alternative. PM9's prune button is moot under the ruling.

## Left for the conductor

- Update the STATUS lane row's ceiling from 3.5M to 4.5M at the next registry touch.
- Dispatch the `prose-voice-reviewer` read on the five drafts (Task 3).
- Decide with Geoff whether the Phillips letter goes through Nicky Akehurst.
- Carry N2 (withdrawn blobs in git history) to the release checklist.
