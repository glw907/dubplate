# Photo curation round one plan review: contract and criteria

Lens: contract and criteria. Plan: `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md`
at `1240bfa`. Spec: `docs/superpowers/specs/2026-09-25-photo-curation-design.md` rev 2.2.
Reviewer: Opus 5.5, 2026-09-25.

Verdict: the plan is sound in shape and the spec is almost fully covered. Its weak seam is the one
between Task 1's script, Task 2's page, and Task 6's close. Several things the page and the close
need are produced by no task, and a few criteria cannot pass as written. There are no blockers,
11 majors, and 8 minors. There are no owner forks: every finding has an architecturally clear
answer. Each fold below keeps the workflow as easy as the spec intends and adds no new owner
touchpoint.

Probes run for this review (scratchpad only, nothing written to the tree):

- `node --test tools/photos/` on Node v24.20.0 against a directory holding one passing test
  prints `Error: Cannot find module '.../tools/photos'` and reports `✖ tools/photos ... 'test
  failed'`. The command `node --test 'tools/photos/*.test.mjs'` passes, and it does not run
  helper `.mjs` files under `testdata/`.
- `curl` of the Flickr page `https://www.flickr.com/photos/dubdem/3251415411/` (200, about 520KB
  of raw HTML) carries a site-wide `"license-models":[...]` table whose first entry is
  `"name":"All Rights Reserved"`. The photo's own license appears only as a numeric
  `"license":4` in the photo model, beside CC BY 2.0 URLs.
- ImageMagick 7.1.2-31 and cwebp 1.6.0 are installed, matching the plan's tech stack.

## Majors

### PC1 (major): Task 1's gate command cannot pass on Node 24

Location: plan:94 and plan:113.

Defect: `node --test tools/photos/` fails on Node v24.20.0 with "Cannot find module" (proven
above). Node 22 and later treat positional arguments as file globs and do not accept a
directory. The Task 1 acceptance "`node --test tools/photos/` passes" is therefore unsatisfiable
as written. An implementer would improvise a different command, and the diff-reviewer would then
grade against a criterion the task never stated.

Fold: in both places, use `node --test 'tools/photos/*.test.mjs'` (proven passing). Name the
same string wherever Task 6 re-runs the tests.

### PC2 (major): the empty-pool criterion contradicts balance's exit rule

Location: plan:99-102 and plan:121-122.

Defect: balance "exits 1 on any unwaived breach", but an empty pool breaches every floor cell
and every surface floor. So "`integrity` and `balance` pass on the real, still-empty `photos/`"
cannot hold unless balance breaks its own rule. Share arithmetic on zero records also divides by
zero, and the plan names no expected output for that. Balance also has one exit code for two
different states: "breach" and "could not read the pool" (an unparsable record or a missing
`vocab.json`). The page's count bar is seeded from that output, so a crash and a real report look
the same to the conductor.

Fold:
- Balance exits 0 when there is no unwaived breach, 1 on a breach, and 2 on unreadable input,
  and it prints no JSON on exit 2.
- On an empty pool, balance exits 1, reports every share as 0 of 0 with the status `short` (never
  NaN), and lists every cell and surface as short.
- The acceptance becomes: `integrity` exits 0 on the empty pool, and `balance` exits 1 with that
  exact report. A fixture asserts each of the three exit codes.

### PC3 (major): the page's running counts and second-image flags have no producer

Location: plan:98-102, plan:137-142, and plan:214.

Defect: Task 6 step 1 runs "`check.mjs balance --json` over the pool plus candidates", but Task
1's interface takes one `--root` (the pool) and reads no candidates. Candidates live under
`.candidates/<agent>/`, carry no status, and would fail picked-record rules anyway.

The page needs three more things that no task produces:
- The derivation maps (scene to family, country code to region) to count Geoff's keeps live.
  The spec says family is "derived from `vocab.json`, never stored".
- The picked pool's artist and landmark keys, to flag a second image.
- A rule for two candidates in the same sitting that share an artist.

Candidates carry artist names only, because MusicBrainz IDs arrive at close (spec:114-118). A
flag computed on names can therefore miss a pair that later resolves to one MBID ("Lee Perry" and
"Lee 'Scratch' Perry"). Integrity then fails at close on a duplicate Geoff never saw flagged, and
he never waived it. The plan names no report for that state.

Fold:
- `balance --json` also emits the derivation maps and the picked pool's normalized artist and
  landmark names and IDs. It runs over the pool only.
- The conductor builds the round's candidate list and adds each candidate's derived family,
  region, and primary subject to it.
- The page adds keeps to the counts client-side. It flags a candidate whose normalized artist or
  landmark name matches the pool or another kept candidate.
- A duplicate that first appears after the MusicBrainz step goes back to Geoff unpicked, or goes
  into the round's one checkpoint question. The conductor never writes a waiver Geoff did not
  make.

### PC4 (major): the unchanging page has no way to find each round's files, and the round flag has no stated home

Location: plan:139-142 and plan:217-218.

Defect: Uploaded assets get server-assigned URLs that the page must reference "exactly as
given" (Artifact tool contract). A page that "never changes between rounds" therefore cannot know
round N's candidate list or thumbnail URLs unless a channel carries them. The plan names none.

The closed-round flag, a privileged field, has no named home, no named writer, and no reopen
path. The spec says "a later change of mind reopens the round explicitly". Choice documents are
keyed by `source_id` alone, so a closed round-one flag cannot be told apart from round two's open
state.

Fold: add one database document, `round`, written only by the conductor through `ArtifactData`.
- It carries the round number, `closed`, the balance JSON, and the candidate list inline, with
  each thumbnail and preview URL.
- Choice documents carry the round number, and only the page writes them. The conductor never
  writes a choice.
- `closed` is set and cleared only by the conductor, and the page has no control that writes it.
  Reopening means the conductor clears `closed`.

### PC5 (major): Task 2's acceptance misses the closed state, the review-focus load, and fixture isolation

Location: plan:144-149, against review focus 4 (plan:56-57).

Defect:
- The closed-round refusal is part of the outcome, but no acceptance criterion exercises it.
- Review focus 4 names 150 thumbnails on a phone, while the acceptance renders 20.
- "Save and reload correctly" can only run on the published artifact, where the database lives.
  The plan does not say that fixture choices must use fixture `source_id`s and be cleared before
  round one. If a fixture reuses a real ID, a test keep becomes Geoff's keep at close.

Fold: the acceptance renders the published page at desktop width and at 390px with a
150-candidate fixture whose IDs are namespaced (for example `fixture:1`...).
- A `Y`, an `N`, and a tap each save a choice, and all three survive a reload.
- With `closed` set through `ArtifactData`, `Y` and `N` are refused and the page shows the
  closed state.
- The fixture's documents and assets are deleted before round one's upload, and a `list` of the
  collection shows zero fixture documents.

### PC6 (major): no task supplies the full-size file that `Space` opens

Location: plan:137-141 and plan:214-215, against spec:321.

Defect: the spec's page opens the full-size file on `Space`, which is Geoff's only way to judge
focus and the tall crop. Task 6 uploads thumbnails and the candidate list only. Uploading raw
sources is not a safe substitute. Commons originals are often multi-megabyte JPEGs or TIFFs, and
150 of them press against the per-file and per-version asset limits.

Fold: at step 1, the conductor encodes each candidate with the spec's own 1600px command into a
preview WebP (about 164KB each, measured) and uploads it beside the thumbnail. `Space` opens the
preview. The preview is the exact file round close would commit, so Geoff judges what ships.

### PC7 (major): the candidate record has no defined schema, and `candidate` checks the JSON but not the files

Location: plan:103-104 and plan:194-196.

Defect: the spec's field table describes picked records. The plan says `candidate` checks
"required fields" without listing which. The implementer invents the candidate schema, and six
agents then conform to whatever it chose.

The size floor has no candidate-stage evidence. `width` and `height` are stored-file fields, and
a candidate has no source dimensions. So "`size_exception` where 1000 to 1199px" (Task 4) cannot
be checked, and a sub-1000px keep surfaces only at close, after Geoff spent a pick on it.

`candidate` also never hashes the source file against `source_sha256`. It never confirms that
the thumbnail exists, and never checks that an NC or permission-only license carries
`pack: true`. A wrong hash is caught only at close, where step 2 silently returns Geoff's keep.

Fold:
- List the candidate's required fields: the picked table minus `sha256`, `width`, `height`,
  `status`, and `withdrawn`, plus `screen`, plus `source_width` and `source_height`. The agent
  records the last two from `magick identify` and close strips them like `screen`.
- `candidate` hashes the sibling source file and requires the sibling thumbnail.
- `candidate` enforces the size floor and `size_exception` on the source long edge, capped at
  1600.
- `candidate` requires `pack: true` exactly when the license's class in `vocab.json` is `pack`.
- One fixture each covers a hash mismatch, a missing thumbnail, and an NC license without
  `pack`.

### PC8 (major): candidates that two agents both found are never deduplicated before the sitting

Location: plan:176-191 and plan:214.

Defect: territories split the research files, but one image can appear in several of them, and
the new searches can reach the same Commons file. Each agent checks only the rejected ledger. Two
parallel agents can therefore write the same `source_id` (or a Commons and Flickr pair sharing an
`origin_id`) into different subdirectories. Geoff sees the image twice and may keep both. The
duplicate is caught only by integrity at close, and no report is named for it.

Fold: Task 6 step 1 first deduplicates across all agent directories on `source_id`,
`origin_id`, and `source_sha256`, keeping the first found and logging the rest as `duplicate`
drops. The simplest home is `check.mjs candidate <dir>`, which accepts a directory and reports
cross-record repeats.

### PC9 (major): the Flickr raw-evidence rule is ambiguous in a way that can quote the wrong license

Location: plan:36-37, Task 4's brief, and Task 5.

Defect: "quotes the license link from the page HTML" (spec:259-260) is recalled, not proven.
The probe shows raw Flickr HTML with a site-wide `license-models` table that names every license,
"All Rights Reserved" first. The photo's own license is a numeric `"license":4` in the photo
model. An agent grepping for a license string can quote the table. The Task 5 reader uses "the
same raw channel" and can repeat the mistake, so the spot-check does not catch it.

Fold: Task 4's and Task 5's briefs name the Flickr evidence exactly:
- Quote the photo model's `"license":<n>` together with the license URL rendered for that
  photo.
- Map `<n>` through a table held in `vocab.json` that Task 1 writes from Flickr's published
  license list, with the list's URL recorded.
- Never quote from `license-models`.

For a Flickr-origin Commons file, the FlickreviewR template stays the evidence of record, as the
spec says.

### PC10 (major): "every rule is implemented" rests on nine fixtures, and the waiver path is unproven

Location: plan:110-119.

Defect: the fixtures cover nine rules. The Integrity and Balance bullets hold about twenty. None
of these is proven able to fail:
- The slug pattern.
- The `public_domain` required and forbidden rule, and "PD in the US".
- `artwork` required for a flyer.
- A `sha256` mismatch, and a picked record with no `.webp`.
- A withdrawn record that keeps its image.
- A `source_id` repeat against `rejected.json` or `pack.json`.
- A second landmark.
- The region cap, the posed ceiling, a share band, the landmark floor, and the surface floors.

Most consequentially, no fixture proves that a waiver suppresses exactly its breach. A waiver
match that is too loose (a wrong key still suppresses) or too strict (a right key does not)
passes the acceptance, and the waiver is the only owner-authority field in the script.

The planned fixtures also include no `photos/.candidates/` directory. An implementation that
recurses into it would flag every thumbnail as "a `.webp` without a record" and every draft
record as an unstatused JSON file, and that surfaces first at round close.

Fold: make the fixtures one table with one failing case per Integrity and Balance rule, which is
cheap as a table-driven `node:test`. Add these waiver cases:
- A waived breach passes.
- The same waiver with a wrong key fails.
- A waiver on an unknown `rule` fails integrity.

Add a fixture pool containing `.candidates/a/commons-1.webp` and `.json` that must pass
integrity.

### PC11 (major): the license list and other closed lists are unspecified, and an off-list term has no report

Location: plan:88-91 and plan:105-107.

Defect: Task 1 must write the license list, the public-domain bases, and the reason codes, but
neither the plan nor the spec enumerates them. The spec gives only examples. Separately, the
parenthetical "(empty arrays or objects)" reads as covering `vocab.json` too.

Ported CC licenses (`CC-BY-SA-3.0-NL` for Anefo, `-DE`, `-AT`) are common in exactly the
archives this round re-verifies. An agent whose license is off the list cannot add it (spec:76-78)
and has no named action. It either drops a good image or maps it to the unported ID. The second
choice is a wrong license ID committed as fact.

The starting reason codes also have no code for a vanished source page, which research-d's
unverified TSV will produce often.

Fold:
- Task 1 builds the license list from the licenses the research files actually cite, found by
  grep and recorded in the report, plus the spec's shipped classes. `vocab.json` is populated
  and never empty.
- Reason codes gain `unavailable` and `vocab`.
- An agent that meets an off-list term drops the candidate with `vocab` and names the proposed
  term in its report. The conductor extends `vocab.json` and re-queues the candidate next round.
  An agent never maps to a near term.

## Minors

### PC12 (minor): Task 5's sample is unspecified and its re-verify scope narrows the spec

Location: plan:203-209.

Defect: the plan does not say how the 20% is drawn or who computes the size. A reader could
pick easy records, or log fewer than the formula requires, and still meet "each sampled record
logged".

The plan also sends back "that agent's whole batch", while the spec says "a round with a wrong
quote is re-verified whole" (spec:316-317).

A Flickr license changed since the grant is logged, not counted as a mismatch (spec:261-262),
and Task 5 does not repeat that rule. A legitimate change would bounce a whole batch.

Fold:
- The conductor draws the sample with a recorded seed and passes the list to the reader.
- The log states the computed size and the pool it came from.
- Adopt the spec's whole-round scope.
- Add a third result, `changed-since-grant`.

### PC13 (minor): Task 6 leaves four close-time states without a named report

Location: plan:217-221.

Defect and fold, one state at a time:
- **Unsorted candidates** (Geoff neither kept nor passed them). They stay in `.candidates/` and
  on the page for the next round, and never go to `rejected.json`.
- **A hash mismatch or missing file.** It is listed in the round log and in the note to Geoff.
- **The slug "checks it is unused" rule** (spec:144). It is absent from the steps. Refuse a
  slug whose `.webp` or `.json` exists.
- **A MusicBrainz lookup with zero or several plausible matches.** Leave the ID absent, so the
  normalized-name key applies, and log the name. Never take the first hit.

### PC14 (minor): a transcluded Commons license template has no stated fetch step

Location: plan:49-51 and Task 4.

Defect: review focus 1 says the check "must still find a shipped license ID", but `check.mjs`
never reads wikitext, and the agent writes the ID. Raw wikitext for an Anefo or Bundesarchiv
file shows `{{Anefo}}`, not the license.

Fold: Task 4's brief has the agent also fetch `Template:<name>` with `action=raw` and quote the
license line inside it in `verified`.

### PC15 (minor): code-simplifier runs after Task 1's self-commit, with no re-test

Location: plan:94-95 and plan:222-224.

Defect: Task 1's implementer commits its own files, so the simplifier (Task 6 step 3) always
runs over committed code. Its edits then need a commit and a re-run of the tests, and neither
step is stated.

Fold: Task 1's chain runs the simplifier after the diff-reviewer accepts and before the
conductor commits (the implementer does not commit). Otherwise, Task 6 step 3 re-runs the Task 1
test command and commits with the Global constraints' form.

### PC16 (minor): Task 2 and Task 6 omit two spec page rules

Location: plan:137-142.

Defect: two spec rules have no task. "Pack candidates are not shown" (spec:324) and "passed
candidates are pruned from the file store after each round" (spec:332).

Fold: add the first rule to Task 2's outcome, and the second as a Task 6 close step.

### PC17 (minor): capability availability and JSON upload shape are assumed

Location: plan:131-133.

Defect: runtime capabilities depend on what is enabled for the person. The plan does not name
what happens if the database or asset-store capability is absent. Separately, a JSON file goes
in an upload call of its own (Artifact tool contract). That no longer matters once PC4 holds the
list in the `round` document.

Fold: the conductor-first step confirms both capabilities in the roster the skill serves, and
stops for Geoff if either is missing.

### PC18 (minor): "every committed image was kept by Geoff" has no stated proof

Location: plan:226-227.

Defect: the criterion is true only by the conductor's assertion. No step cross-checks it.

Fold: the round log lists the committed `source_id`s beside the `round` collection's keep
documents, and the two sets are equal. An empty keep set is valid and logged as zero.

## Spec coverage

Every spec requirement lands in some task except these:
- `Space` opening the full size (PC6).
- Hiding pack candidates and pruning the file store (PC16).
- The slug-unused check (PC13).
- Candidate dedupe across agents (PC8).

Deferred by design and correctly left to later rounds: waiver proposals after two empty rounds,
the done checkpoint, `outreach/grants/` records, the pre-release source-URL recheck, withdrawal,
and everything the spec assigns to rung 7.

Observation, not a finding: the working tree already carries Task 0's `.gitignore` line, the
`.gitleaks.toml` allowlist entry, and the STATUS lane row. Task 0's executor should verify them
rather than re-add them.
