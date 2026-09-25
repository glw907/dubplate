# Photo curation round one plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-25-photo-curation-round-one.md` at `1240bfa` (cited as
P), against the spec at revision 2.2 (cited as S). Lens: can each task be carried out as written,
by the named agent type with its tools. Every claim comes from a probe run on 2026-09-25 in the
session scratchpad outside the repo, or from a quoted primary source. No probe sent Geoff's
contact address anywhere; the Wikimedia and Flickr probes used a contact-free User-Agent.

Counts: 1 blocker, 8 majors, 8 minors. One owner fork, inside PM9.

## What holds

- Tool versions match the pins: ImageMagick `7.1.2-31 Q16-HDRI`, cwebp `1.6.0` (libsharpyuv
  0.4.2), Node `v24.20.0`, curl `8.22.0`, git `2.55.0`, gitleaks `8.30.1`. A general-purpose
  subagent runs all of them: this review is one, and every probe below ran from its Bash tool.
- `DUBPLATE_HARNESS_CONTACT` is set in this subagent's shell (it comes from `~/.bashrc`), so
  search agents can build the User-Agent without being handed the value.
- The encode pipeline runs clean on an 8000x6000 JPEG, a CMYK JPEG with an embedded CMYK profile,
  and a 16-bit TIFF. All three come out as 8-bit sRGB WebP at 1600px on the long edge.
- A Flickr photo page returns 200 to curl. It carries the license link
  (`creativecommons.org/licenses/by/2.0`) and a sizes JSON with the original's dimensions (the
  probed TSV row's original is 614x561, a `too-small` drop). No Flickr API key is needed.
- gitleaks with the repo's `.gitleaks.toml` finds nothing in a record-shaped JSON file (64-hex
  hashes, Wayback and Flickr URLs, a FlickreviewR quote) placed in `photos/` or in
  `tools/photos/testdata/`, so photo files do not redden lane 1's gitleaks leg.
- Lane 1's `make tidy-check` diffs only `-- go.mod go.sum` (Makefile:31), so photo-lane files
  cannot fail it.
- Every research input Task 4 names exists. Task 0 has already landed at `fae78c3`: the
  `.gitignore` and `.gitleaks.toml` entries and the STATUS lane row are present.

## Blocker

### PM1 (blocker): `node --test tools/photos/` fails on Node 24, so Task 1's gate and acceptance cannot pass as written

- Location: P:94, P:113.
- Evidence: a probe tree held `tools/photos/check.test.mjs` with one passing test. On Node
  24.20.0, `node --test tools/photos/` and `node --test tools/photos` both exit 1 with
  `Error: Cannot find module '.../tools/photos'`, so the runner treats a directory argument as a
  module path. `node --test 'tools/photos/*.test.mjs'` passes. The Node v24.x test runner doc
  (`doc/api/test.md`, "Running tests from the command line") says: "By default, Node.js will run
  all files matching these patterns: `**/*.test.{cjs,mjs,js}` ..." It also says "Alternatively,
  one or more glob patterns can be provided as the final argument(s)". A glob is accepted. A
  directory is not.
- Bare `node --test` from the repo root is no fix. With type stripping on, the default patterns
  also match `**/*.test.{cts,mts,ts}`, so the run would sweep `_salvage/web/src/*.test.ts`.
- Fold: replace both occurrences with `node --test tools/photos/check.test.mjs`, or with the
  quoted glob `node --test "tools/photos/*.test.mjs"`.

## Majors

### PM2 (major): `dubplate-implementer` is wired to `scripts/check.sh`, so the Task 1 dispatch must override its gate and its gate lane in words

- Location: P:92-95. Agent definition `.claude/agents/dubplate-implementer.md:3`, `:73-81`.
- Evidence: the description says the agent "clears the gate string the dispatch names (today
  `bash scripts/check.sh`)". Workflow step 5 then fixes the form: "in the `cd
  /var/home/glw907/Projects/dubplate && bash scripts/check.sh` form (today's dispatch gate)". The
  agent can take another gate, but only if the dispatch says so explicitly. Without that, it runs
  `check.sh`. That breaks the plan's "Never run `scripts/check.sh`" rule and makes the agent wait
  on the per-tree `check.sh.lock` (check.sh:116-121) that lane 1 holds.
- `cairn-run-gate` defaults to the heavy lane, which is one machine-wide lock. STATUS:12 says
  every full lane 1 gate from Task 3 on launches Chromium. A heavy-lane Node test would queue
  behind those gates. The script header says `CAIRN_GATE_LANE=light` "queues on its own lock, so
  it never waits behind a browser gate".
- The per-task `diff-reviewer` runs `git status --short`, `git diff`, and `git diff --staged`
  (diff-reviewer.md:16). In a shared tree, after the implementer has committed, those commands
  show lane 1's in-progress edits and none of Task 1.
- Fold: the dispatch carries these lines verbatim.
  - "Your gate string is `cd /var/home/glw907/Projects/dubplate && node --test
    tools/photos/check.test.mjs && node tools/photos/check.mjs integrity`. Run it as
    `CAIRN_GATE_LANE=light cairn-run-gate '<that string>'`. This replaces step 5's `check.sh`
    form. Never run `scripts/check.sh`."
  - "Read `~/.claude/skills/ts-conventions/SKILL.md` for comment rules."
  - The reviewer dispatch names the commit and scopes the diff: "Review `git show <sha> --
    tools/photos photos/meta`. Uncommitted changes elsewhere belong to lane 1 and are not this
    task."

### PM3 (major): six parallel search agents exceed Wikimedia's documented concurrency limits

- Location: P:172-191, P:34-37.
- Evidence (fetched raw on 2026-09-25):
  - The Wikitech Robot policy, under "Action API rules": "If unauthenticated, keep the
    concurrency of your requests to 1 at a time, and below 5 requests per second overall."
  - The same page, under "Media API rules" (`upload.wikimedia.org`): "Always keep a total
    concurrency of at most 2, and limit your total download speed to 25 Mbps". It adds "Prefer
    thumbnails to downloading of originals if possible."
  - MediaWiki "Wikimedia APIs/Rate limits": "limit the number of concurrent requests to 3 or
    fewer", and a compliant User-Agent gets "200" requests per minute.
  - The User-Agent policy, as quoted in the spec review, says scripts "may be blocked without
    notice".
- Six agents on one IP, each running curl, reach six-way concurrency on the Action API and on the
  media host. The consequence is 429s that cost agent turns, and at worst a block on the
  workstation IP that stalls every later round.
- Fold: add to the Global constraints (the brief carries it to each agent):
  - Every Commons API or `index.php` request runs as `flock
    /tmp/dubplate-photos-wm-api.lock curl --retry 5 ...`. curl's manual says `--retry` covers
    HTTP 429 and 503 and that "curl complies with the Retry-After: response header".
  - Every `upload.wikimedia.org` download runs under a second lock (`...-wm-media.lock`) with
    `--limit-rate 2M`.
  - Size-filter in batches before downloading: one `api.php?action=query&prop=imageinfo&iiprop=size|url`
    request takes up to 50 titles ("Where supported, use batch requests"). By the spec's own
    table, about 30 of 113 known files are under 1000px and need no download at all.

### PM4 (major): the sorting page cannot find a candidate list uploaded to the asset store, and nothing is uploaded for `Space`

- Location: P:137-142, P:214-216. S:321.
- Evidence: the runtime contract (`assets.d.ts`, contract 0.2.59) says "a listing carries no
  names or meaning of its own, so `list()` is for housekeeping ..., not a gallery source". It also
  says "The `db` document an id is stored in remains the asset's index". A page that "never
  changes between rounds" therefore has no way to learn the id of this round's list file.
- Task 6 uploads "thumbnails and the candidate list" only, yet `Space` opens the full-size file.
  The originals cannot simply be uploaded either. The asset store accepts png, jpeg, gif, and
  webp but not TIFF, and it caps each file at 20 MiB. Some archive originals are TIFF or larger
  than that.
- The mechanism that works is `db`. `ArtifactData` `batch` takes "up to 50 set, update or delete
  writes", each from a local JSON `file_path`. A document may be up to 256 KiB, and a query may
  return up to 1000 documents.
- Fold, in Task 2's outcome and in Task 6 step 1:
  - The page reads candidates from a db collection such as `rounds/<n>/candidates`, one document
    per candidate. Each document carries the caption fields, `thumb` and `preview` asset ids, and
    the flags.
  - The balance report is one db document per round.
  - The conductor uploads the 600px thumbnail and a 1600px preview WebP (the pinned encode), then
    seeds the documents with the ids the upload result returns.
  - An alternative is a `candidates.json` published through `files` on a republish, which the
    page fetches by relative path. The id join is still needed.

### PM5 (major): nothing in the plan produces the candidate list, the "pool plus candidates" balance, or the second-image flag

- Location: P:98-104, P:137-141, P:214.
- Evidence: `balance [--root <dir>]` counts picked records under one root. Candidates have no
  `status` and sit in `.candidates/<agent>/`, so "balance --json over the pool plus candidates"
  has no interface. The page must flag "a second image of a picked artist or landmark" (S:327),
  which needs the picked records' dedupe keys, and the balance JSON carries only "every count,
  band, and breach".
- No task merges six agents' records into one list. That merge has to drop pack records, dedupe
  across agents on `source_id` and `origin_id` (at least one known overlap:
  `File:Corner_Coxsone_Dodd_Kingston_Jamaica.jpg` is in both research-a's and research-b's
  files), sort by score, and join the asset ids from PM4. Done by hand in the main loop, that is
  grinding over about 150 records.
- Fold: give Task 1 a fourth mode, `check.mjs sheet [--root]`. It reads every
  `.candidates/*/*.json`, drops pack records, dedupes, sets the second-image flag against the
  picked keys, orders by `screen` score, and writes one JSON file per candidate, ready for an
  `ArtifactData` batch. `balance --json` also emits the picked artist and landmark keys. This
  raises Task 1 from three modes to four, and the conductor should state that at dispatch.

### PM6 (major): round close is written as inline conductor work at a scale the main loop should not grind

- Location: P:217-221.
- Evidence: step 2 covers every keep, likely 50 to 150 in round one. Each keep needs a hash
  check, a slug, an encode, and a record edit, and then come the MusicBrainz loop, the waivers,
  the rejected merge, and the pack moves. In the Opus main loop that is hundreds of tool calls
  against a 3.5M ceiling. It also breaks the conductor-is-thin rule. The MusicBrainz step is a
  name search, and the plan never says what to do with an ambiguous or missing match.
- A subagent can do this work: `ArtifactData` is loadable in this general-purpose subagent
  session, so the agent can read the closed round's choices itself (`list` with `out_dir`).
- Fold: step 2 dispatches a `general-purpose` Sonnet agent at `high` for close steps 2 to 4. It
  returns a structured report (kept, returned unpicked with the reason, waivers written, the MBID
  table, integrity output). The conductor keeps step 1 (closing the round in the db) and step 5
  (the commit). The MBID rule: record an MBID only on a single exact-name match; otherwise leave
  it empty and let the normalized-name key apply.

### PM7 (major): the empty-pool acceptance contradicts balance's exit rule

- Location: P:100-102 against P:121-122. S:47.
- Evidence: balance "exits 1 on any unwaived breach". An empty pool breaches all 15 floor cells
  (each needs at least two images), the landmark floors, and the surface floors. The same task
  also requires that balance "pass on the real, still-empty `photos/`". Shares on an empty pool
  are 0 of 0. This is the "each state a check can meet names its report" class that the repo's
  planning-miss rule lists.
- Fold: state the empty case exactly. Balance on zero picked records exits 1 and lists every
  floor, landmark floor, and surface floor as short. Shares print as `0 of 0` with no share
  breach. `--json` is well-formed. The acceptance becomes "runs without an exception, emits that
  report, exits 1". Integrity alone must exit 0.

### PM8 (major): Task 2's save-and-reload acceptance cannot run where the plan puts it

- Location: P:144-146.
- Evidence: the runtime resolves `claude.use("db")` to `null` "after 10 s when no viewer
  answers" outside claude.ai. A local render therefore cannot prove saves. CLAUDE.md: "an MCP
  server reaches only the main loop", so a subagent cannot drive Playwright MCP, and the rule
  keeps the conductor from grinding a render loop. Running the 20-candidate fixture on the real
  artifact would leave fixture documents in the production db.
- Fold:
  - The page ships an in-memory store it uses when `use("db")` is `null`, seeded from a local
    fixture file. A subagent renders it with headless `chromium-browser` at 1280 and 390 widths
    and drives the keyboard and tap paths.
  - On the published page, run the skill's one functional pass: fixture documents under
    `rounds/test`, an `ArtifactData list` to confirm the page's writes, then `delete` of those
    documents.
  - The conductor reads the screenshots, per `see-it-before-geoff-tests-it`.

### PM9 (major, with one OWNER FORK): asset uploads and pruning cost Geoff's attended time as written

- Location: P:214-216. S:330-332.
- Evidence: the Artifact tool uploads "up to 25 ... files in one call under one approval". About
  150 thumbnails plus 150 previews is 12 calls, and each JSON file takes "a call of its own". Two
  things about deletes: the tool says "the person confirms every delete", and it deletes one
  asset per call. The spec's pruning of passed candidates would therefore cost about 100
  confirmations. The plan's Task 6 also leaves out pruning, which the spec requires.
- Fold:
  - The page itself prunes. After close, one "Remove passed images" button calls
    `assets.delete` for each passed id. The contract allows that on "a deliberate user action".
  - Or skip pruning until `assets.list().usage` nears `maxBytes`, which the page can display.
    The per-artifact quota is published only at runtime.
- OWNER FORK: Geoff either allowlists the Artifact asset uploads for this lane or accepts about
  12 approvals per round, batched into one announced moment before the sitting.

## Minors

### PM10 (minor): `source_id` is the db document key, but archive IDs break the doc-id pattern

- Evidence: the `ArtifactData` `doc_id` pattern is `[A-Za-z0-9_\-.~:@+]{1,200}`, which allows no
  space and no slash. An archive inventory number such as `Bild 183-...`, or an archive name with
  a space, fails the write.
- Fold: `check.mjs candidate` enforces `source_id` against
  `^(commons|flickr|archive):[A-Za-z0-9_.~:@+-]{1,190}$`.

### PM11 (minor): integrity's scope and candidate mode's key check are unstated

- Evidence: the rules "no `.webp` lacks a picked record" and "no JSON file fails to parse" would
  trip on every candidate thumbnail and draft record if integrity walks into `.candidates/`.
- Fold: integrity reads top-level `photos/*.{json,webp}` and `photos/meta/*.json` only, and
  ignores `.candidates/`. `candidate` also reports a key collision with picked, rejected, or pack
  records, which is the check S:295 asks agents to make.

### PM12 (minor): the commit mechanic leaves stray intent-to-add entries if a commit fails

- Evidence (git 2.55.0 probes):
  - Another lane's `git commit -a` sweeps in an intent-to-add path with full content.
  - `git diff` lists it as a new file.
  - `git stash` refuses with "Entry 'q.json' not uptodate. Cannot merge." while it is present.
  - Committing an untracked path without `-N` fails with "pathspec ... did not match".
  - `git reset -q -- <path>` returns an intent-to-add path to untracked and keeps the file.
  - A concurrent commit can also fail with `cannot lock ref 'HEAD'` rather than `index.lock`.
- Fold: run `git add -N` and `git commit --` in one Bash call. On failure, run `git reset -q --
  <new paths>` before retrying. Retry on either lock message.

### PM13 (minor): parallel ImageMagick decodes can push memory next to lane 1's heavy gate

- Evidence: a 48MP JPEG peaks at 718536 KB RSS in `magick ... -resize`. The machine has 15 GB
  with about 6 GB available, and lane 1's heavy gate is capped at 8G.
- Fold: add `-limit memory 512MiB -limit map 1GiB` to the pinned `magick` call. Limits change
  memory use, not output. Or fetch Commons' standard 3840px thumbnail when the original is larger.
  Every agent's temporary PNG goes inside its own candidate subdirectory.

### PM14 (minor): the code-simplifier step never fires as timed

- Location: P:222-224.
- Evidence: Task 1's implementer commits its own files, and round close does not normally touch
  `tools/photos/`. So "before the round's first commit touching those files" has no trigger.
- Fold: the conductor dispatches `code-simplifier`, with the brief pasted verbatim, after Task
  1's reviewer accepts. It then re-runs the PM1 gate and commits the refinements.

### PM15 (minor): two actions are assigned to agents that cannot take them

- Task 3's `prose-voice-reviewer` read (P:170): CLAUDE.md says "Subagents cannot dispatch
  agents", so the conductor dispatches it after the drafts land.
- Publishing a page Claude did not write (P:135): the Artifact tool says "Claude reads the whole
  file before publishing it", so a conductor publish means a source read. Two options:
  - The Task 2 agent publishes. A general-purpose agent has the Artifact tool (this session does;
    the Explore and Plan agent types list it as excluded).
  - The conductor accepts the one read.
- Either way, every later upload is made by whoever published or first ran `action: read` on the
  artifact. The tool says "A publish to an artifact this conversation has not read or published
  is refused".

### PM16 (minor): the contact value could leak into files

- Evidence: the spec keeps it out of requests other than User-Agents, but no rule keeps agents
  from writing it into candidate records, `drops.json`, or the round log.
- Fold: one Global constraint. Reference `$DUBPLATE_HARNESS_CONTACT` only inside the curl command
  line, and never write or echo its value.

### PM17 (minor): the per-agent turn count threatens the 3.5M ceiling

- Evidence: research-d's TSV has 139 rows, and the research files name 141 distinct Commons
  files. About six tool turns per candidate (page, download, identify, encode twice, record)
  comes to several hundred turns for research-c and research-d each. This is an estimate, not a
  probe.
- Fold: each agent writes one small fetch-and-encode script, serialized under PM3's locks, and
  spends its turns on judgment. Split research-d into two agents of about 70 rows.
