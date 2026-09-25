# Photo curation round one implementation plan

> **For agentic workers:** each task below runs as a chain: an implementer or search agent, then
> the check named in its acceptance, then (for code) one `diff-reviewer`. Tasks specify outcomes,
> constraints, and acceptance criteria, never implementation code.

**Goal:** stand up the curation tooling, run round one of the photo pool through Geoff's first
sorting sitting, and put the first outreach letters in Geoff's hands.

**Architecture:** data under `photos/` (one WebP and one JSON record per picked image, plus
`photos/meta/`), one dependency-free Node check script under `tools/photos/`, a sorting page
published as a private claude.ai artifact whose rounds live in its database, and letter drafts
under the research directory. No Go, nothing under `web/` or `cmd/`, no `Makefile` or `scripts/`
edit in this plan.

**Tech stack:** Node 24 (`node:fs`, `node:crypto`, `node:test` only), ImageMagick 7.1.2, cwebp
1.6.0, `curl`, `flock`, the Artifact tool with its database and asset capabilities, the
`ArtifactData` tool.

**Spec:** `docs/superpowers/specs/2026-09-25-photo-curation-design.md`, revision 2.3 (revision
2.2 at `90a9ad1`, approved by Geoff 2026-09-25, plus his two plan-review rulings of the same
day). Every task reads the spec sections it names; the spec wins on any conflict with this plan.
Fold record: `docs/superpowers/research/2026-09-25-photo-curation-plan-fold.md`.

**Lane:** the photo document lane, in the main tree beside lane 1 (5c-iii, a code lane). Token
ceiling 4.5M (raised from 3.5M at the fold; the fold record gives the arithmetic); the 80% flag
is 3.6M. Checkpoint after Task 4 and before any question to Geoff. Execution mode: per-task Agent
dispatches (below six sequential tasks; Task 3 is already drafted).

**Owner rulings built in (Geoff, 2026-09-25):**
- People and minors, "Flag, you decide": a search agent flags a candidate that carries a Commons
  personality-rights template or shows an apparent minor in focus, the sorting page shows the
  flag, and Geoff's keep or pass decides. No blanket exclusion.
- Approvals, "Batch, no pruning": Geoff approves each round's uploads in one burst, and passed
  images are never deleted from the page's store, so no delete prompts.

## Global constraints

- Never touch `web/`, `cmd/`, `internal/`, `Makefile`, `scripts/`, or the 5c-iii plan. Lane 1
  owns them. Never stage a path outside the task's Files list; lane 1's warm changes share this
  tree.
- Never run `scripts/check.sh`. This lane's checks are `node --test tools/photos/check.test.mjs`
  and `node tools/photos/check.mjs integrity` (Task 1). A directory argument
  (`node --test tools/photos/`) fails on Node 24 with "Cannot find module" (review probe, PM1).
- Commits: in ONE Bash call, `git add -N -- <new paths>` then `git commit -- <paths>`, then
  compare `git show --stat HEAD` with the path list. If the commit fails, run
  `git reset -q -- <new paths>` in the same call before any retry, so no intent-to-add entry is
  left for another lane's `git commit -a` to sweep in (PM12 probes). Retry on `.git/index.lock`
  or `cannot lock ref 'HEAD'`; never delete a lock. Only the conductor commits, except where a
  task says otherwise.
- Encoded images enter `photos/` only inside the same Bash call that commits them (Task 6).
  Before that they wait under `photos/.candidates/`, which is ignored, so they never join lane
  1's scoped change set.
- STATUS changes only at lane open and lane close, through this lane's registry row, and only
  when `git status -- docs/STATUS.md` is clean. Checkpoints go to the round log.
- Every HTTP request's User-Agent names dubplate and takes its contact from
  `$DUBPLATE_HARNESS_CONTACT`. Reference that variable only inside the curl command line; never
  write or echo its value into a file, a record, a log, or a report.
- Wikimedia and Flickr etiquette. The Wikitech Robot policy says: "If unauthenticated, keep the
  concurrency of your requests to 1 at a time, and below 5 requests per second overall", and for
  `upload.wikimedia.org`, "Always keep a total concurrency of at most 2" (quoted in PM3). So:
  - every Commons `api.php` or `index.php` request runs under
    `flock /tmp/dubplate-photos-wm-api.lock curl --retry 5 ...`;
  - every `upload.wikimedia.org` download runs under `flock /tmp/dubplate-photos-wm-media.lock`
    with `--limit-rate 2M`;
  - every Flickr request runs under `flock /tmp/dubplate-photos-flickr.lock`.
  curl's manual says `--retry` covers HTTP 429 and 503 and that "curl complies with the
  Retry-After: response header" (PM3). An agent that still meets 429 or 403 after the retries
  stops and reports; it never loops.
- License evidence is fetched raw with `curl`, never through a summarizing fetch tool. A raw page
  is saved to the agent's candidate directory and never read whole into context: the agent
  extracts the lines it needs with `grep` or `sed` and quotes the matched line into `verified`.
  - Commons: the license and review templates from `action=raw` wikitext. When the license sits
    in a transcluded template (`{{Anefo}}`, a Bundesarchiv template), also fetch
    `Template:<name>` with `action=raw` and quote the license line inside it.
  - Commons sizes come from one batched `api.php?action=query&prop=imageinfo&iiprop=size|sha1|url`
    request of up to 50 titles, so an under-1000px file is dropped before any download.
  - Flickr: quote the photo model's `"license":<n>` together with the license URL rendered for
    that photo, and map `<n>` through the table in `vocab.json`. Never quote from the page's
    site-wide `"license-models"` table, which lists every license with "All Rights Reserved"
    first (PC9 probe). For a Flickr-origin Commons file, the FlickreviewR template stays the
    evidence of record.
- Shipped license classes: public domain (also PD in the US, on a base from `vocab.json`'s
  list), CC0, CC BY, CC BY-SA, Free Art License. NC or permission-only goes to the pack. A term
  missing from `vocab.json` (a license, a base, a scene) drops the candidate with reason `vocab`
  and the proposed term in the report; an agent never maps to a near term.
- Laundering and deletion. A Commons page carrying a deletion, copyright-violation,
  no-permission, or no-source template (`{{delete}}`, `{{copyvio}}`, `{{No permission since}}`,
  `{{No source since}}`) or an open deletion request is dropped with `license`. A Flickr-origin
  candidate's account is checked against Commons' "Questionable Flickr images" list and its other
  uploads for mixed authorship. An "own work" claim on a work dated more than about 20 years
  before its upload states in `verified` why the uploader is plausibly the author, or the
  candidate is dropped.
- Inputs: only JPEG, PNG, TIFF, and WebP sources, confirmed with `magick identify` before
  encoding. Anything else (PDF, SVG) is dropped with reason `format`.
- Large originals: download the original when it is under about 40MB. Above that, download a
  named rendition of at least 1600px, set `file_url` to it, and note in `verified` that
  `source_sha256` covers that rendition.
- Size floor: 1200px long edge; 1000 to 1199px only with `size_exception`; under 1000px never.
  Measured on the source's long edge. Never upscale.
- Encode: `magick -limit memory 512MiB -limit map 1GiB <src> -auto-orient -resize '1600x1600>'
  <tmp>.png` then `cwebp -q 80 -m 6 -metadata icc`. The limits bound memory and leave the output
  unchanged (PM13: a 48MP JPEG peaked at 718536 KB RSS without them). The 600px thumbnail is the
  same encode with `600x600>`. Temporary files stay inside the agent's candidate directory.
- Prose (letters, logs) follows the writing-voice output style; letters use plain terms and no
  record-shop puns.

## Review focus

1. A Commons file whose license template sits inside a transcluded template: the agent quotes
   the template it found and the license line inside it, and the record carries a license ID
   from the list.
2. The same photo reached through Commons and through its Flickr original: `sheet` and
   `integrity` both dedupe on `origin_id`. Task 1's fixtures include this pair.
3. A JSON record that fails to parse, or a `.webp` with no record: integrity fails with the file
   named. Task 1's fixtures include both.
4. A sorting page opened on a phone with 150 thumbnails: the page stays usable and saves each
   choice as it is made. Task 2's acceptance renders a 150-candidate fixture at 390px.
5. A pick whose candidate file changed or vanished after the sitting: round close returns it
   unpicked, never commits a substitute. Task 6's close checks `source_sha256` first.

---

### Task 0: Lane preflight (conductor, inline) - DONE at `fae78c3`

Landed: the STATUS lane registry row, `/photos/.candidates/` in `.gitignore` and the
`.gitleaks.toml` allowlist, the identity spec's pointer line, and ROADMAP's `photo-curation`
entry. Carried in STATUS for the batch after lane 1 closes: the `scripts/check.sh` integrity
line and the 5c-iii plan's Ruling 3 erratum. Do not re-run.

### Task 1: Vocabulary and check script

**Files:** create `photos/meta/vocab.json` (populated), `photos/meta/waivers.json`,
`photos/meta/rejected.json`, `photos/meta/pack.json` (the last three empty), `tools/photos/check.mjs`,
`tools/photos/check.test.mjs`, `tools/photos/testdata/`.

**Implementer:** `dubplate-implementer`, model `sonnet`, effort `high`. It reads the spec's Pool
size and balance, Categories, Storage, Licensing, and Process sections in full, and
`~/.claude/skills/ts-conventions/SKILL.md` for comment rules. Deliverable count at dispatch:
four modes, one populated vocabulary, one fixture table.

**Dispatch text (verbatim):** "Your gate is
`cd /var/home/glw907/Projects/dubplate && CAIRN_GATE_LANE=light cairn-run-gate 'node --test tools/photos/check.test.mjs'`.
This replaces workflow step 5's `bash scripts/check.sh` form. Never run `scripts/check.sh`.
Re-issue the same command on exit 75 until it prints the gate exit status." The lane choice
rests on the tool's header: "CAIRN_GATE_LANE=light is for a gate that launches no browser ...:
it queues on its own lock, so it never waits behind a browser gate" (`~/.local/bin/cairn-run-gate`,
lines 17 to 19). The tool has no `--help` flag; its header is its usage text.

**Reviewer:** `diff-reviewer` reads `git show <sha> -- tools/photos photos/meta`. Uncommitted
changes elsewhere belong to lane 1 and are not this task.

**After accept (conductor):** dispatch `code-simplifier` over `tools/photos/` with
`.claude/instructions/simplifier-brief.md` pasted verbatim, re-run the gate above, and commit the
refinements with the Global constraints' form.

**Interfaces:**
- `integrity [--root <dir>]`: reads the top-level `*.json` and `*.webp` of the root and
  `meta/*.json`, and ignores `.candidates/`. Exits 0 clean, 1 on any failure with one line per
  failure naming the file and rule, 2 on unreadable input (a missing `vocab.json`).
- `balance [--root <dir>] [--json]`: runs over the picked pool only. Exits 0 with no unwaived
  breach, 1 on a breach, 2 on unreadable input, and prints no JSON on exit 2. `--json` carries
  every count, band, and breach, plus the derivation maps (scene to family, country code to
  region) and the picked pool's normalized artist and landmark names and IDs.
- `candidate <file.json | dir>`: validates candidate records so search agents can self-check. On
  a directory it also reports repeats across the records it read, and a key collision with picked,
  rejected, or pack records.
- `sheet [--root <dir>] --out <dir>`: reads every `.candidates/*/*.json`, drops pack records,
  dedupes across agents on `source_id`, `origin_id`, and `source_sha256` (first found wins, the
  rest listed as `duplicate` drops), sets the second-image flag against the picked keys and the
  other candidates' normalized names, adds each candidate's derived family, region, and primary
  subject, orders by `screen` score, and writes one JSON document per candidate, ready for an
  `ArtifactData` batch.

**The vocabulary (`vocab.json`), populated, never empty:**
- Scenes with their family, subjects, shots, formats, surfaces, and regions with their ISO
  3166-1 country codes, all from the spec's Categories section.
- The license list: ID, display name, legal code URL, and class (`shipped` or `pack`). It holds
  the spec's shipped classes plus every license the research files cite (found by grep over
  `docs/superpowers/research/2026-09-24-identity-photos/`, the grep and its hits in the report),
  ported versions such as `LicenseRef-CC-BY-SA-3.0-NL` included.
- A Flickr license-number table mapping `"license":<n>` to a license ID, with the URL of the
  Flickr page it was taken from recorded beside it. A number no raw source confirms stays out.
- The public-domain bases, each naming the fact it requires: publication before 1931 (the
  rolling 95-year term); US federal government work; US publication before 1978 without notice;
  US publication 1978 to 1989 without notice and with no registration within five years.
  "No known copyright restrictions" and a PDM mark alone are not bases (PD4).
- Reason codes: `passed`, `blurred`, `too-small`, `license`, `duplicate`, `unavailable`,
  `vocab`, `format`.
- People flags: `personality-rights`, `apparent-minor`.

**The candidate record:** the picked table's fields minus `sha256`, `width`, `height`, `status`,
and `withdrawn`, plus `screen`, `source_width`, `source_height` (from `magick identify` on the
source), `pack` where set, and `flags` where any apply. `candidate` requires the sibling source
file to hash to `source_sha256`, the sibling `<id>.thumb.webp` and `<id>.preview.webp` to exist,
the size floor and `size_exception` rule on the source long edge, `pack: true` exactly when the
license's class is `pack`, and `source_id` matching `^(commons|flickr|archive):[A-Za-z0-9_.~:@+-]{1,190}$`,
which keeps it inside `ArtifactData`'s `doc_id` pattern (PM10).

**Acceptance:**
- Every rule in the spec's Integrity and Balance bullets is implemented, with the waiver keys the
  spec lists, the normalized-name dedupe (case folded, leading "the" dropped), and "no JSON file
  fails to parse".
- The gate passes. The tests are one table with one failing case per Integrity rule, per Balance
  rule, and per `candidate` rule above, each asserting the named reason, not merely a nonzero
  exit. The table includes at least: a duplicate artist; an NC-licensed pick; an NC artwork under
  a CC BY scan; an empty floor cell; a Commons record and its Flickr original sharing an
  `origin_id`; an unparsable record; a `.webp` without a record; a picked record at 1100px
  without `size_exception`; a picked record at 900px; a `public_domain` on an unlisted base; a
  hash mismatch; a missing thumbnail; an NC license without `pack`; a bad `source_id`. It also
  proves the waiver path: a waived breach passes, the same waiver with a wrong key fails, and a
  waiver on an unknown `rule` fails integrity. A passing pool exercises every field and holds a
  `.candidates/a/` directory with a record and a `.webp` that integrity must ignore. Each of the
  three exit codes has a case.
- Each failing case is proven able to fail: the implementer reports, per case, the one-line
  change that makes it pass.
- On the real, empty `photos/`: `integrity` exits 0; `balance` exits 1, reports every share as
  `0 of 0` with the status `short` (never NaN), and lists every floor cell, landmark floor, and
  surface floor as short; `--json` is well-formed.
- Comments follow the ts-conventions skill's comment rules.

### Task 2: Sorting page

**Files:** create `tools/photos/sorter.html` (the artifact's source, committed so every round
reuses it) and `tools/photos/testdata/sorter-fixture.json`.

**Conductor first:** load `artifact-design` and `artifact-capabilities`. Confirm the `db` and
`assets` capabilities in the roster the skill serves; if either is missing, stop and tell Geoff.
Pass the implementer the exact `db` and `assets` calls, the collection and document-id rules, the
page contract, and the capability declaration, all quoted from the skill. The implementer never
guesses the runtime API.

**Implementer:** `general-purpose`, model `sonnet`, effort `high`. It publishes the page itself
(it wrote the file, and the Artifact tool publishes only a file its caller has read) and returns
the URL.

**Outcome:** the page the spec's Process section describes, reading everything from its database:
- One round document per round, written only by the conductor's side through `ArtifactData`: the
  round number, `closed`, and the balance JSON. One candidate document per candidate (from
  `sheet`), each with its caption fields, flags, and the thumbnail and preview asset URLs. One
  choice document per candidate, keyed by `source_id` and carrying the round number, written only
  by the page. The page has no control that writes `closed`.
- A thumbnail grid in score order with captions (subject, place, year, license); the
  `size_exception` line, the second-image flag, and any people flag shown beside the image;
  `Y`/`N`/arrows/`Space`, clicks and taps; `Space` opens the 1600px preview, the exact encode
  round close commits; filter chips (family, subject, region); a count bar seeded from the round's
  balance JSON that adds keeps client-side; a closed round shows its state and refuses changes.
  Pack candidates never appear (`sheet` drops them).
- When `claude.use("db")` resolves to `null` (the runtime does so "after 10 s when no viewer
  answers" outside claude.ai, PM8), the page falls back to an in-memory store seeded from a local
  fixture file, so it can be rendered headless.

**Acceptance:**
- A `general-purpose` Sonnet render agent serves the page locally with the fallback store and a
  150-candidate fixture whose IDs are namespaced (`fixture:1` ...), then renders it with
  `chromium-shot` at desktop width and with `--mobile` (390x844). It drives the keyboard and tap
  paths with Playwright loaded read-only from a scratchpad script (`web/node_modules/playwright`,
  1.63.0; it installs nothing and writes nothing under `web/`). `Y`, `N`, and a tap each save a
  choice; a closed fixture round refuses both keys and shows its state. It returns the
  screenshots, which the conductor reads before Geoff sees anything.
- One functional pass on the published page by the conductor, under the separate round id
  `test`: seed a few fixture candidates through `ArtifactData`, make a keep and a pass, confirm
  them with an `ArtifactData` `list`, set `closed` and confirm the page refuses a change, then
  delete every `test` document and confirm a `list` shows none.
- It follows the artifact contract (tokens on `:root`, dark mode, no external origins, 16px
  gutter) and loads nothing outside its own files.
- Plain labels; no record-shop wording.

### Task 3: Outreach drafts - DRAFTED at `b843519`, folded at the plan fold

The five drafts and `outreach/README.md` exist. The plan fold added the plain CC BY sentence,
described the photo pack as planned and undated, and marked the Charlie Phillips draft `blocked`
(no MX record for `charliephillipsarchive.com`; the fallback route is noted in the draft).
Remaining: the conductor dispatches one `prose-voice-reviewer` read over the five drafts (a
subagent cannot dispatch it), folds its ranked findings, and hands the drafts to Geoff in the
same message as the sitting notice (Task 6). No contact is made; Geoff sends.

### Task 4: Round one search (after Task 1)

**Agents:** seven in parallel, `general-purpose`, model `sonnet`, effort `high`, each writing
only to its own `photos/.candidates/<agent>/`, each with the runaway guard armed at about 2MB (the
dispatch-mechanics memory's figure for research readers):

1. `research-a`: re-verify `photo-candidates.md`, `photo-candidates-2.md`, `photos-reggae-deep.md`,
   `photos-reggae-retry.md`.
2. `research-b`: re-verify `photos-crate-deep.md`, `photos-places.md`.
3. `research-c`: re-verify `photos-africa.md`, `photos-asia-europe-pacific.md`,
   `photos-mena-sa-latam.md`.
4. `research-d1`: re-verify the first half of `flickr-search-unverified.tsv`'s rows (every row is
   unverified; expect a high drop rate).
5. `research-d2`: the second half.
6. `hiphop`: new search for the hip-hop family, every era, go-go, graffiti, and b-boy scenes, with
   a landmark sweep.
7. `flyers`: new search for show flyers and zine covers, DC hardcore first, with the artwork's own
   license found or the candidate dropped. A no-notice public-domain claim rests on the scan of
   the notice-free published copy. A flyer built around another photographer's image needs that
   photographer's license too, or it is dropped; an embedded photograph is noted in `screen`.

Each agent's brief carries the vocabulary (`photos/meta/vocab.json`), the Global constraints, the
spec's Categories, Storage, and Licensing sections, the rejected ledger, and the other agents'
territories. Each agent writes one small fetch-and-encode script in its own directory, run under
the Global constraints' locks, and spends its turns on judgment rather than per-file tool calls.
Each writes a source, a `<id>.thumb.webp`, and a `<id>.preview.webp` (the 1600px encode) per
kept candidate.

People flags: an agent sets `flags` to `personality-rights` when the Commons wikitext carries a
personality-rights template (quoted in `verified`), and to `apparent-minor` when an apparent minor
is in focus. It never drops a candidate for either.

**Acceptance per agent:**
- Every kept candidate passes `node tools/photos/check.mjs candidate` on the agent's directory.
- Every drop is in its `drops.json` with keys and a reason code.
- The agent reports: kept, dropped by reason, pack count, flagged count, counts by family and
  subject, proposed vocabulary terms, and its own token spend.
- The conductor ends Task 4 with a checkpoint in the round log
  (`docs/superpowers/research/photo-curation-rounds/round-1.md`): the agents' counts and the spend
  so far.

### Task 5: License spot-check (after Task 4)

**Reader:** `general-purpose`, model `opus`, effort `high`, fresh context. The conductor passes a
seed; the reader draws the spec's sample (every public-domain, flyer, and zine cover record, plus
20% of the rest, at least five) by a stated deterministic method over the `sheet` output, and
logs the seed, the method, the pool size, and the computed sample size. It checks each record
through the same raw channel and the same Flickr rule as the search agents, applies the
laundering and deletion checks, and writes its results to the round log.

**Acceptance:** each sampled record is logged as `match`, `mismatch`, or `changed-since-grant`
(a later change on a live Flickr page, logged per the spec, not a mismatch), with the date. Any
mismatch sends the whole round back for re-verification before Task 6, per the spec.

### Task 6: Round open, first sitting, and round close

**Round open.** The conductor tells Geoff the uploads are coming (about 12 approval prompts: 150
thumbnails and 150 previews in calls of 25), so he approves them in one burst. It then dispatches
a `general-purpose` Sonnet round-open agent that runs `balance --json` and `sheet`, reads the
artifact once (`action: read`, which the Artifact tool requires before an upload), uploads the
thumbnails and previews, and seeds the round and candidate documents through `ArtifactData`
`batch` with the URLs the upload results return. It returns counts. The conductor opens the page
and reads one screen at desktop and phone width, then sends Geoff one message: the sitting is
ready, and the five letters are ready to read and send (two touchpoints, one message). Passed
images are never deleted from the store.

**Close.**
1. When Geoff says he is done, the conductor sets `closed` on the round document.
2. The conductor dispatches a `general-purpose` Sonnet close agent at `high`, which loads
   `ArtifactData`, reads the closed round's choices, and carries out spec close steps 2 and 3
   into `photos/.candidates/_close/`, never into `photos/`:
   - a keep applies only where the candidate file's hash equals `source_sha256`; a missing file
     or mismatch goes back to Geoff unpicked and into the round log;
   - it mints each slug and refuses one whose `.webp` or `.json` already exists; encodes;
     writes `sha256`, `width`, and `height`; strips `screen`, `flags`, `source_width`, and
     `source_height`;
   - it runs the MusicBrainz step at one request per second, recording an MBID only when exactly
     one result matches the normalized name and its area or active years fit the record;
     otherwise the ID stays absent, the name key applies, and the name is logged;
   - it writes the waiver a flagged keep carries (the spec's rule), and never any other: a
     duplicate that first appears after the MBID step goes back to Geoff unpicked;
   - it merges passes and drops into `rejected.json` and moves pack records to `pack.json`
     (tracked meta files, edited in place); unsorted candidates stay in `.candidates/` for the
     next round and go nowhere;
   - it runs `integrity --root` over a staging copy of the pool plus the new pairs.
   It returns: kept, returned unpicked with reasons, waivers written, the MBID table, the path
   list, and the integrity output.
3. The conductor rules on anything returned, then commits in one Bash call: move the staged pairs
   into `photos/`, run `integrity`, commit with the Global constraints' form, and check
   `git show --stat HEAD`; on any failure it resets the new paths and moves the pairs back.

**Acceptance:** `check.mjs integrity` passes on the committed pool; the round log lists the
committed `source_id`s beside the round's keep documents and the two sets are equal (an empty set
is valid and logged as zero); `git show --stat HEAD` matches the path list; the round log carries
the counts, the balance report, and round two's targets (the gaps the report shows).

## Later rounds

Round two onward repeats Tasks 4 through 6 with agents aimed at the balance report's gaps and a
landmark sweep per family. Each later round gets its own short plan section in this file, written
at the previous round's close. Letters that return a grant join a round as ordinary candidates.
