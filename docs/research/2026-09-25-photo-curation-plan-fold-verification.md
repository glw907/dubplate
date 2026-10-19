# Photo curation round one: plan fold verification

Fresh verification read of the plan fold at `a2aeb1b`. Plan:
`docs/plans/2026-09-25-photo-curation-round-one.md` (cited as P). Fold record:
`docs/research/2026-09-25-photo-curation-plan-fold.md` (cited as F). Spec revision
2.3 (cited as S). Reviewer: Opus 5.5, 2026-09-25, no part in the fold.

**Verdict:** the fold is sound. The blocker and every major close where the record says, with one
partial (PC4's reopen path). The fold introduced no blocker. Six new majors sit in the seams the
fold created: the page's headless fallback, the live-page proof, the close agent's meta writes,
the `duplicate` drops, the candidate directory's file layout, and the database layout the Task 6
agents need. Each has a fold of a sentence or two. Counts: 0 blockers, 6 majors, 10 minors.

## Probes

All probes ran in the session scratchpad, outside the repo.

- `node --test tools/photos/check.test.mjs` on Node v24.20.0 against a one-test file: exit 0.
  The folded gate string works.
- `flock <file> echo` runs (`/home/linuxbrew/.linuxbrew/bin/flock`, command form accepted).
- `magick -limit memory 512MiB -limit map 1GiB in.jpg -auto-orient -resize '1600x1600>' out.png`
  on a 3000x2000 JPEG: exit 0, 1600x1067 output. The unit spelling `MiB` is accepted.
- `man curl`, `--retry`: "Transient error means either: a timeout, an FTP 4xx response code or
  an HTTP 408, 429, 500, 502, 503, 504, 522 or 524 response code", and "curl complies with the
  Retry-After: response header". The plan's quote (P:67-68) holds.
- `~/.local/bin/cairn-run-gate` lines 17 to 19 match the plan's quote (P:146-148). The gate
  state key is `$PWD` plus the gate string, so the `cd` prefix in the dispatch is load-bearing
  and correct.
- `web/node_modules/playwright` is 1.63.0; `~/.cache/ms-playwright` holds `chromium-1243` and a
  headless shell. `chromium-shot --mobile` is "a 390x844 viewport". Both claims in P:256-257
  hold.
- The runtime contract (artifact-capabilities 0.2.59, `claude.d.ts:16-30`, `assets.d.ts:1-45`,
  `db.d.ts`) and the `ArtifactData` schema (batch "1-50 entries", `doc_id` pattern
  `[A-Za-z0-9_\-.~:@+]{1,200}`) were read directly. PM10's `source_id` pattern fits the
  `doc_id` pattern. S:160 makes the Commons key `commons:<pageid>`, so Commons titles with
  parentheses never reach the key.
- The five letters: each carries "CC BY lets anyone", none says "non-commercial", and each body
  from the subject line runs 236 to 248 words. The one em dash (`beth-lesser.md:16`) is inside a
  quotation from her site in the header notes, not in the letter. Phillips is `blocked` in both
  the draft and `outreach/README.md`.
- Every Task 4 research input exists; `flickr-search-unverified.tsv` has 140 lines.

## Question 1: did each blocker and major close?

| ID | Closed? | Where |
|---|---|---|
| PM1 (blocker) | Yes | P:42-44, P:143; probe above |
| PC1 | Yes | merged with PM1 |
| PC2 | Yes | P:158-162, P:214, P:217-219 |
| PC3 | Yes | P:161-164, P:168-173, P:357-358 |
| PC4 | **In part** | Round, candidate, and choice documents and conductor-only `closed` are at P:238-242. F:45 says "reopening clears it", but the plan has no reopen text (PV13). |
| PC5 | Yes, with PV1 and PV2 | P:253-264 |
| PC6 | Yes | P:245-246, P:303-304, P:337-338 |
| PC7 | Yes | P:192-198, P:207-210 |
| PC8 | Yes, with PV4 | P:168-171 |
| PC9 | Yes | P:78-82, P:182-183 |
| PC10 | Yes | P:204-214 |
| PC11 | Yes | P:84-86, P:178-181, P:188-189 |
| PM2 | Yes | P:142-150 |
| PM3 | Yes, see PV15 | P:59-69 |
| PM4 | Yes, see PV9 | P:237-241, P:336-339 |
| PM5 | Yes | P:139-140, P:168-173 |
| PM6 | Yes, with PV3 | P:346-364 |
| PM7 | Yes | merged with PC2 |
| PM8 | Yes in shape; PV1 and PV2 break two of its mechanisms | P:249-264 |
| PM9 | Settled by ruling; see PV8 | P:33-35, P:334-335 |
| PD1 | Yes | letter headers; README |
| PD2 | Yes | letters; probe above |
| PD3 | Yes | P:87-93, P:325 |
| PD4 | Yes | P:184-187, P:295-296 |
| PD5 | Downgraded by ruling; the folded parts hold, except PV3 | P:51-55 |
| PD6 | Yes | P:70-77, P:281-282, P:314 |

## Question 2: contradictions and order

The main order holds. Task 1 produces every script interface the later tasks call; Task 2's
page reads only documents Task 6 writes; round close runs closed, then read and stage, then
rule, then commit. The breaks are PV3 (the close agent's in-place meta writes contradict its own
"never into `photos/`"), PV4 (`duplicate` drops against integrity's no-repeat rule), and PV7
(Task 5 samples a `sheet` output that no step before it produces).

## Question 3: mechanisms stated from memory

Backed: the `node --test` file form, `flock`, the `magick -limit` units, curl's retry and
Retry-After behavior, the `cairn-run-gate` header, Playwright 1.63.0, `chromium-shot --mobile`,
the 25-file upload call, the 50-write batch, and the `doc_id` pattern. Misquoted or unbacked:
the headless fallback trigger (PV1), the live-page keep and pass (PV2), the approval count
(PV8), and the asset pointer (PV9).

## Question 4: self-containment

Task 1's dispatch is self-contained except for its commit form (PV12). Task 4's brief list is
complete apart from the candidate directory layout (PV5) and the research directory path (PV16).
The Task 6 round-open and close agents cannot be zero-context today: nothing records the
database layout Task 2 invents (PV6).

## Question 5: proportionality

Proportionate. Geoff's attended time is one sitting, one approval burst, and one letter read,
which fits an easy curation workflow. The machinery (four script modes, seven search agents, a
spot-check reader, round-open and close agents) sits where the licensing risk is, and each piece
replaces main-loop grinding rather than adding a gate. The one place the plan doubles a proof is
Task 2's headless render plus a live pass, and both are needed: only the headless render reaches
150 thumbnails at 390px, and only the live pass reaches the real store. No finding below asks
for more machinery; most remove an ambiguity.

## Findings

### Majors

**PV1 (major), P:249-251.** The fallback trigger misreads the runtime contract, and the fallback
can mask a live-store failure.
- Defect: the plan says `claude.use("db")` resolves `null` "after 10 s when no viewer answers"
  outside claude.ai. The contract (`claude.d.ts:19-30`) says a page served top-level by the
  platform gets `null` from every `use()`. It also says "any other top-level copy of the page (a
  saved file, another host) has no `window.claude` at all". The 10 s `null` applies only when
  the page is "Framed by a host that never answers". The render agent's locally served page is
  the no-`window.claude` case. A page written to the plan's trigger throws a `ReferenceError`
  there, and Task 2's render acceptance fails on its first run. Second, a published page that
  meets `null` (the platform-served top-level view, or a failed module load) would fall back to
  an in-memory store. A sitting could then run with choices saved nowhere, and nothing would say
  so.
- Fold: fall back to the fixture store only when `window.claude` is absent and the page is
  served from `localhost`. On a `null` `db` anywhere else, show a plain "cannot save choices
  here" state with no grid. One line in Task 2's outcome and one render case.

**PV2 (major), P:261-264.** The live functional pass has no executor for its keep and pass.
- Defect: "make a keep and a pass" on the published page needs a browser signed in to claude.ai
  as Geoff, since the artifact is private. `ArtifactData` writes documents directly and does not
  exercise the page's own writes. The plan names no tool, and the conductor has no stated way to
  press `Y` on the live page. Yet this pass is the plan's only proof that a choice made on the
  page reaches the store, and every sitting rests on that.
- Fold: pick one mechanism and name it. Recommended: the conductor seeds the `test` round, then
  Geoff makes one keep and one pass on it as the first step of the sitting notice (under a
  minute, and he has the page open anyway). The conductor confirms both with an `ArtifactData`
  `list`, sets `closed`, has Geoff press `Y` once to see the refusal, then deletes the `test`
  documents before seeding round one. If a claude-in-chrome session signed in to claude.ai
  exists, the conductor may drive it instead; verify that once before relying on it.

**PV3 (major), P:348 against P:359-361 and P:365-367.** The close agent writes tracked meta
files in place.
- Defect: step 2 says the close agent works "into `photos/.candidates/_close/`, never into
  `photos/`". The same step then merges into `rejected.json` and `pack.json` and writes waivers,
  "tracked meta files, edited in place", which is `photos/meta/`. Those edits then sit
  uncommitted in the shared tree while the conductor rules on returns, which is the exposure
  PD5's folded parts removed. Step 3's failure path "resets the new paths and moves the pairs
  back" but never restores the meta files. A round the conductor sends back to Geoff also leaves
  a `rejected.json` that already names his passes.
- Fold: the close agent writes the updated `rejected.json`, `pack.json`, and `waivers.json`
  into `_close/meta/`. Step 3's single Bash call copies them over `photos/meta/` with the pairs,
  and its failure path restores them with `git checkout -- photos/meta`.

**PV4 (major), P:168-171, P:188-189, P:359-362, against S:219-224.** `duplicate` drops break
integrity at close, and the losing copies never leave.
- Defect: integrity fails when any `source_id`, `origin_id`, or hash "repeats across picked,
  rejected, and pack records". A `duplicate` drop exists precisely because its keys match
  another record. `sheet`'s losers share keys with their winners, and an agent's `duplicate`
  drop shares keys with an existing record. If close merges either into `rejected.json`, the
  pool fails integrity as soon as the winner is kept, or on the second copy of any passed image.
  Round one hits this on its first cross-agent overlap: PM5 names
  `File:Corner_Coxsone_Dodd_Kingston_Jamaica.jpg`, found in both research-a's and research-b's
  files. If close does not merge them, the plan has no rule for the losing files: they are
  "unsorted", so they stay in `.candidates/` and resurface every round.
- Fold: `duplicate` drops never enter `rejected.json`. The round log lists them, and close
  deletes the losing candidate files, since the winner's record already carries the keys a
  later round checks against. Add one Task 1 fixture: a rejected entry that repeats a picked
  key fails integrity, which pins the rule the fold relies on.

**PV5 (major), P:70-72, P:168, P:311-312, P:348.** Non-candidate JSON shares the candidate glob.
- Defect: `sheet` reads every `.candidates/*/*.json`, and `candidate <dir>` validates every
  record in the directory. The same directories also hold `drops.json` (P:312), raw evidence
  saved beside the records (P:70-72; an `api.php` `imageinfo` response is JSON), and the close
  agent's `_close/`. `sheet --out` names no location, and output inside `.candidates/` would be
  read back as candidates. Every agent's self-check then fails on its own `drops.json`, and
  `sheet` ingests raw API responses and staged picked records.
- Fold: a candidate record is `<source>-<key>.json` in the spec's filename form (S:204-205,
  `commons-12345.json`). `candidate` and `sheet` read only that pattern and skip directories
  whose names start with `_`. Raw evidence goes under `<agent>/raw/`. `sheet --out` writes to
  `photos/.candidates/_sheet/`. State this in Task 1's interfaces and in Task 4's brief.

**PV6 (major), P:229-231, P:237-242, P:336-339, P:346-348.** The Task 6 agents cannot learn the
database layout.
- Defect: Task 2's implementer chooses the collection paths, document ids, and field names for
  round, candidate, and choice documents, and the plan pins none of them. The round-open agent
  seeds those documents, and the close agent reads choices by round. Both start with zero
  context, and nothing tells them the layout. Guessing wrong seeds documents the page never
  reads, which surfaces only when Geoff opens an empty grid.
- Fold: Task 2's acceptance adds that the page's layout is written once, as the collection
  paths, the document id rule, and each document's fields, in a comment block at the top of
  `tools/photos/sorter.html` and in the implementer's report. The Task 6 dispatches quote that
  block. The artifact URL goes into the round log at Task 2's close, and every Task 6 dispatch
  names it.

### Minors

**PV7 (minor), P:321-324.** Task 5 draws its sample "over the `sheet` output", but `sheet`
first runs at Task 6's round open. Fold: the Task 5 reader runs `sheet --out` itself first, and
the round-open agent reuses that output unless a Task 5 re-verification changed a record.

**PV8 (minor), P:334-335.** "About 12 approval prompts" counts only uploads. Seeding about 150
candidate documents plus the round document takes four `ArtifactData` batches (50 writes per
batch, one approval each), and setting `closed` takes one more. Fold: announce about 16.

**PV9 (minor), P:240, P:338-339.** The contract says to store the asset `id` in `db` as "the
durable pointer" and build the display URL as `"/_blob/" + id` (`assets.d.ts:30-43`). The
upload's `url` is "already correct for the current view". Fold: candidate documents carry both
the `id` and the returned `url`, and the page renders `/_blob/<id>`.

**PV10 (minor), P:336-339.** The Artifact tool asks Claude to read a file it did not write
before publishing it. The round-open agent uploads about 300 images that the search agents
wrote. Reading them all would cost several hundred thousand tokens, and skipping the read
silently breaks the rule. Fold: the round-open agent re-encodes each thumbnail and preview from
the source with the pinned command, so it wrote what it uploads, and it views a handful as a
sample. The conductor's one-screen read covers the rest.

**PV11 (minor), P:227-231.** The Artifact tool's contract binds the agent that writes and
publishes: that agent loads `artifact-design`, and it loads `artifact-capabilities` before it
passes `capabilities`. The conductor loading both skills and quoting them thickens the main loop
and still leaves the publishing agent outside the contract. Fold: the implementer loads both
skills itself, checks the roster for `db` and `assets`, and stops with a report if either is
missing.

**PV12 (minor), P:142-150, P:45-50.** Task 1's dispatch text does not carry the Global
constraints' commit form. `dubplate-implementer` step 6 commits "only the task's files" in its
own form. In a shared index, a plain `git commit` after `git add` would sweep in anything lane 1
has staged. Fold: add one sentence to the verbatim dispatch: commit with `git add -N --` and
`git commit -- <paths>` in one call.

**PV13 (minor), F:45 against P:238-242.** The fold record says PC4's reopen path landed in
Task 2 ("reopening clears it"), but the plan has no reopen text. Fold: one clause in Task 2's
outcome: the conductor clears `closed` to reopen, and the page accepts changes again.

**PV14 (minor), P:351-352.** The close agent "encodes" each keep again, but P:245-246 promises
that the preview is "the exact encode round close commits". Fold: close copies
`<id>.preview.webp` to the slug's `.webp` and hashes it, so the file Geoff judged is the file
that ships.

**PV15 (minor), P:59-66.** `flock` caps Commons concurrency at 1, but serial requests at LAN
latency can pass the robot policy's "below 5 requests per second". Fold: the locked command
ends with `sleep 0.25` inside the lock.

**PV16 (minor), P:271-276, P:284-291.** Task 3 says `outreach/README.md`, and Task 4 lists
research files by bare name. Neither gives the directory,
`docs/research/2026-09-24-identity-photos/`. An archive inventory number with a
space has no stated normalization into the `source_id` pattern. Fold: name the directory in
both tasks, and say that spaces and slashes in an archive key become `_`.

## Counts

0 blockers, 6 majors (PV1 to PV6), 10 minors (PV7 to PV16).
Of the fold's 1 blocker and 25 majors, the blocker and 22 majors closed. PM9 is settled by
Geoff's ruling, PD5 stands as the conductor downgraded it, and PC4 closed in part (PV13).
