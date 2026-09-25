# Photo curation round one implementation plan

> **For agentic workers:** each task below runs as a chain: an implementer or search agent, then
> the check named in its acceptance, then (for code) one `diff-reviewer`. Tasks specify outcomes,
> constraints, and acceptance criteria, never implementation code.

**Goal:** stand up the curation tooling, run round one of the photo pool through Geoff's first
sorting sitting, and put the first outreach letters in Geoff's hands.

**Architecture:** data under `photos/` (one WebP and one JSON record per picked image, plus
`photos/meta/`), one dependency-free Node check script under `tools/photos/`, a sorting page
published as a private claude.ai artifact, and letter drafts under the research directory. No Go,
nothing under `web/` or `cmd/`, no `Makefile` or `scripts/` edit in this plan.

**Tech stack:** Node 24 (`node:fs`, `node:crypto`, `node:test` only), ImageMagick 7.1.2, cwebp
1.6.0, `curl`, the Artifact tool with its database capability.

**Spec:** `docs/superpowers/specs/2026-09-25-photo-curation-design.md`, revision 2.2 (`90a9ad1`),
approved by Geoff 2026-09-25. Every task reads the spec sections it names; the spec wins on any
conflict with this plan.

**Lane:** the photo document lane, in the main tree beside lane 1 (5c-iii, a code lane). Token ceiling
3.5M; checkpoint after Task 4 and before any question to Geoff. Execution mode: per-task Agent
dispatches (below six sequential tasks; Task 3 runs in parallel with Tasks 1 and 2).

## Global constraints

- Never touch `web/`, `cmd/`, `internal/`, `Makefile`, `scripts/`, or the 5c-iii plan. Lane 1
  owns them.
- Never run `scripts/check.sh`. This lane's check is `node tools/photos/check.mjs` (Task 1).
- Commits: `git add -N -- <new paths>` then `git commit -- <paths>`, then compare
  `git show --stat HEAD` with the path list. Retry on `.git/index.lock`; never delete it. Only the
  conductor commits, except where a task says otherwise.
- Every HTTP request's User-Agent names dubplate and takes its contact from
  `$DUBPLATE_HARNESS_CONTACT`. No other personal address in any request.
- License evidence is fetched raw with `curl` (Commons `action=raw`, Flickr page HTML), never
  through a summarizing fetch tool.
- Shipped license classes: public domain (also PD in the US), CC0, CC BY, CC BY-SA, Free Art
  License. NC or permission-only goes to the pack.
- Size floor: 1200px long edge; 1000 to 1199px only with `size_exception`; under 1000px never.
  Never upscale.
- Encode: `magick <src> -auto-orient -resize '1600x1600>' <tmp>.png` then
  `cwebp -q 80 -m 6 -metadata icc`; thumbnails the same with `600x600>`.
- Prose (letters, logs) follows the writing-voice output style; letters use plain terms and no
  record-shop puns.

## Review focus

1. A Commons file whose license template sits inside a transcluded template (for example an
   Anefo or Bundesarchiv institutional template): the agent must quote the template it found and
   the check must still find a shipped license ID. Task 3's agents record the raw template text.
2. The same photo reached through Commons and through its Flickr original: the second find must
   dedupe on `origin_id`. Task 1's fixtures include this pair.
3. A JSON record that fails to parse, or a `.webp` with no record: integrity must fail with the
   file named. Task 1's fixtures include both.
4. A sorting page opened on a phone with 150 thumbnails: the page must stay usable and save each
   choice as it is made. Task 2's acceptance includes a phone-width render.
5. A pick whose candidate file changed or vanished after the sitting: round close must return it
   unpicked, never commit a substitute. Task 6's procedure checks `source_sha256` first.

---

### Task 0: Lane preflight (conductor, inline)

**Files:** `docs/STATUS.md` (lane registry row only), `.gitignore`, `.gitleaks.toml`,
`docs/superpowers/specs/2026-09-24-visual-identity-design.md` (one pointer line), `ROADMAP.md`
(the `photo-curation` entry).

**Outcome:** the lane is registered, the candidate home is ignored and unscanned, and the two
ratified documents this lane may touch carry their errata.

**Acceptance:**
- STATUS's lane registry carries a row for this document lane naming this plan and its tree.
  Before editing, `git status -- docs/STATUS.md` shows no uncommitted change from another session.
- `.gitignore` carries `/photos/.candidates/`; `.gitleaks.toml`'s allowlist carries
  `'''photos/\.candidates/.*'''` with a comment in the file's existing style.
- The identity spec's Photography section carries one pointer line to the curation spec; its
  Curation rules, Licensing manifest list, and "unmodified" sentence are left as ratified, with the
  pointer naming the curation spec as the later ruling (the fold record's errata list).
- ROADMAP's `photo-curation` entry says about 300, rung 7 consumes it, and the owner touchpoints
  are the sorting sittings and sending letters.
- Queued for the batch after lane 1 closes, and recorded in STATUS's carried items: the
  `scripts/check.sh` integrity line, and the 5c-iii plan's Ruling 3 erratum routing the photo
  column to rung 7.

### Task 1: Vocabulary and check script

**Files:** create `photos/meta/vocab.json`, `photos/meta/waivers.json`,
`photos/meta/rejected.json`, `photos/meta/pack.json` (empty arrays or objects),
`tools/photos/check.mjs`, `tools/photos/check.test.mjs`, `tools/photos/testdata/`.

**Implementer:** `dubplate-implementer`, model `sonnet`, effort `high`. It reads the spec's
Pool size and balance, Categories, Storage, and Licensing sections in full. It does not run
`scripts/check.sh`; its gate is `node --test tools/photos/` and it commits its own files with the
Global constraints' commit form.

**Interfaces:**
- Produces `node tools/photos/check.mjs integrity [--root <dir>]` and
  `node tools/photos/check.mjs balance [--root <dir>] [--json]`. `--root` defaults to `photos/`.
  Integrity exits 0 or 1 and prints one line per failure naming the file and rule. Balance prints
  the report (or JSON with every count, band, and breach, for the sorting page's bar) and exits 1
  on any unwaived breach.
- Produces `node tools/photos/check.mjs candidate <file.json>`: validates one candidate record
  (closed lists, required fields, no `status`, `screen` present) so search agents can self-check.
- Produces the vocabulary shape the other tasks read: scenes with their family, subjects, shots,
  formats, surfaces, regions with their ISO 3166-1 country codes, public-domain bases, reason
  codes, and the license list (ID, display name, legal code URL, shipped or pack class).

**Acceptance:**
- Every rule in the spec's Integrity bullet and Balance bullet is implemented, including the
  waiver keys the spec lists, the normalized-name dedupe (case folded, leading "the" dropped), and
  "no JSON file fails to parse".
- `node --test tools/photos/` passes, and each fixture below is a failing case the test asserts
  fails for the named reason (not merely nonzero): a duplicate artist; an NC-licensed pick; an NC
  artwork under a CC BY scan; an empty floor cell (balance); a Commons record and its Flickr
  original sharing an `origin_id`; an unparsable record; a `.webp` without a record; a picked
  record at 1100px without `size_exception`; a picked record at 900px. One passing fixture pool
  exercises every field.
- Each failing fixture is proven able to fail: the implementer reports, per fixture, the one-line
  change that makes it pass.
- `integrity` and `balance` pass on the real, still-empty `photos/` (an empty pool is a valid
  pool; balance reports every share and cell as short without erroring).
- Comments follow the ts-conventions skill's comment rules (JSDoc on exported functions, none
  restating code).

### Task 2: Sorting page

**Files:** create `tools/photos/sorter.html` (the artifact's source, committed so every round
reuses it).

**Conductor first:** load `artifact-design` and `artifact-capabilities`, and pass the
implementer the exact database and asset-store calls, the page contract, and the capability
declaration. The implementer never guesses the runtime API.

**Implementer:** `general-purpose`, model `sonnet`, effort `high`. The conductor publishes.

**Outcome:** the page the spec's Process section describes: a thumbnail grid in score order with
captions (subject, place, year, license), `size_exception` and second-image flags, `Y`/`N`/arrows/
`Space`, clicks and taps, filter chips (family, subject, region), a count bar seeded from
`check.mjs balance --json`, choices saved per candidate keyed by `source_id`, and a closed-round
state that refuses changes. Candidates and the balance JSON are read from files the conductor
uploads with each round, so the page itself never changes between rounds.

**Acceptance:**
- Rendered and read by the conductor at desktop and 390px widths with a 20-candidate fixture
  before Geoff sees it; keyboard and tap paths both save and reload correctly.
- It follows the artifact contract (tokens on `:root`, dark mode, no external origins, 16px
  gutter) and loads nothing outside its own files.
- Plain labels; no record-shop wording.

### Task 3: Outreach drafts (parallel with Tasks 1 and 2)

**Files:** create `docs/superpowers/research/2026-09-24-identity-photos/outreach/` with one draft
per recipient: Beth Lesser, the British Culture Archive's People's Archive, the DC Punk Archive at
DC Public Library, Charlie Phillips, Janette Beckman; and `outreach/README.md` listing each draft's
status (drafted, approved, sent, answered).

**Drafter:** `general-purpose`, model `opus`, effort `high`, with the spec's Outreach section and
`photographer-outreach.md`.

**Acceptance:**
- Each letter, in Geoff's voice, introduces dubplate as a free, open-source music server for a
  group of friends under the MIT license; never says "non-commercial"; names specific images (from
  the research, each with its URL) or, where none is known, a specific body of work; asks for
  CC BY 4.0 and for the license to be posted somewhere checkable; offers the optional photo pack
  as the fallback; and runs under 250 words.
- Each draft carries, above the letter, the recipient's contact route with its source URL, and
  the images asked for.
- No contact is made. Geoff sends.
- A `prose-voice-reviewer` read runs on the five drafts before Geoff sees them.

### Task 4: Round one search (after Task 1)

**Agents:** six in parallel, `general-purpose`, model `sonnet`, effort `high`, each writing only
to its own `photos/.candidates/<agent>/`:

1. `research-a`: re-verify `photo-candidates.md`, `photo-candidates-2.md`, `photos-reggae-deep.md`,
   `photos-reggae-retry.md`.
2. `research-b`: re-verify `photos-crate-deep.md`, `photos-places.md`.
3. `research-c`: re-verify `photos-africa.md`, `photos-asia-europe-pacific.md`,
   `photos-mena-sa-latam.md`.
4. `research-d`: re-verify `flickr-search-unverified.tsv` (every row is unverified; expect a high
   drop rate).
5. `hiphop`: new search for the hip-hop family, every era, go-go, graffiti, and b-boy scenes, with
   a landmark sweep.
6. `flyers`: new search for show flyers and zine covers, DC hardcore first, with the artwork's own
   license found or the candidate dropped.

Each agent's brief carries the vocabulary (`photos/meta/vocab.json`), the Global constraints, the
spec's Categories, Storage (candidate fields, encode, size floor), and Licensing sections, the
rejected ledger, and the other agents' territories so they do not overlap.

**Acceptance per agent:**
- Every kept candidate has its source file, its 600px thumbnail, and a record that passes
  `node tools/photos/check.mjs candidate`; `size_exception` where 1000 to 1199px; `pack: true`
  where NC or permission-only.
- Every drop is in its `drops.json` with keys and a reason code.
- The agent reports: kept, dropped by reason, pack count, and counts by family and subject.
- The conductor ends Task 4 with a checkpoint: STATUS updated, spend recorded.

### Task 5: License spot-check (after Task 4)

**Reader:** `general-purpose`, model `opus`, effort `high`, fresh context. It checks the spec's
sample (every public-domain, flyer, and zine cover record, plus 20% of the rest, at least five)
through the same raw channel and writes the round log at
`docs/superpowers/research/photo-curation-rounds/round-1.md`.

**Acceptance:** each sampled record logged as match or mismatch with the date. Any mismatch sends
that agent's whole batch back for re-verification before Task 6.

### Task 6: First sitting and round close

**Conductor steps:**
1. Run `check.mjs balance --json` over the pool plus candidates, upload thumbnails and the
   candidate list to the sorting page in batches of 25, and publish. Tell Geoff the round is
   ready (the one expected attended touchpoint).
2. When Geoff says he is done, close the round in the spec's five steps, in order: close in the
   database; apply keeps only where the candidate file's hash equals `source_sha256`; mint slugs,
   encode, write hashes and dimensions, strip `screen`, run the serialized MusicBrainz step at one
   request per second, write waivers, merge passes and drops into `rejected.json`, move pack
   records to `pack.json`; run integrity; commit with the Global constraints' form.
3. Dispatch `code-simplifier` once over `tools/photos/` with `.claude/instructions/simplifier-brief.md`
   pasted verbatim, before the round's first commit touching those files if Task 1 has not
   already had it.

**Acceptance:** `check.mjs integrity` passes on the committed pool; every committed image was
kept by Geoff and matches its `source_sha256`; `git show --stat HEAD` matches the path list; the
round log carries the counts and the balance report; STATUS names round two's targets (the gaps
the balance report shows).

## Later rounds

Round two onward repeats Tasks 4 through 6 with agents aimed at the balance report's gaps and a
landmark sweep per family. Each later round gets its own short plan section in this file, written
at the previous round's close. Letters that return a grant join a round as ordinary candidates.
