# Project site spec review: contract and criteria lens

Target: `docs/superpowers/specs/2026-09-25-project-site-design.md` at `e91f528` (revision 1).
Lens: is every promise testable, can any criterion pass vacuously, are Phase A's deliverables and
Phase B's preconditions stated so done is distinguishable from not-done, and are the ownership
boundaries enforceable.

Verdict: the spec's shape is sound and its ownership model is the right one. The gaps are all in
the same place: the sync mechanism's promises ("can't drift", "the editor can't change them",
"the gate fails") name a check but not the states that check must meet. Four findings are major.
None is a blocker, because every one can be folded into the spec text before the Phase A plan is
written. One finding carries an owner fork.

Facts checked against the tree at review time:

- `web/src/lib/themes/` holds `dubplate.css`, `fonts.css`, `inks.css`, `label.css`,
  `label-night.css`, `surface.css`, `system.css`. The fonts live in `web/static/fonts/` with
  `FONTS.md` and three OFL license files in `licenses/`, and `fonts.css` references them by
  absolute URL (`url('/fonts/...')`).
- `internal/reject` exists only under `_salvage/internal/reject/`. The live tree has no reject
  package yet.
- `internal/config/registry.go` holds the key registry the configuration reference would read.
- `docs/guide/` does not exist. The repo has no release tags except `pre-greenfield`, and no
  release checklist file.

## Findings, ranked by consequence

### C1 (major): the lock-file gate can pass vacuously

Location: spec lines 96-106.

The only stated gate is "fails when a vendored file's hash doesn't match the lock". As written,
that check passes when:

- the vendor directory is empty and the lock lists no files;
- a file listed in the lock is missing from the vendor directory (a hash loop over present files
  never visits it);
- a file sits in the vendor directory that the lock doesn't list;
- the lock file is absent;
- someone hand-edits a vendored file and then hand-edits the lock's SHA-256 to match.

The last case is inherent: the lock is a self-consistency check, and nothing in the site's build
can compare it to the tag because the dubplate repo is private. That limit is acceptable, but the
spec should say so rather than imply the lock proves provenance.

Proposed fold. State the gate's contract as these checks, each with a mutation proof in the Phase
A plan:

1. The lock exists and parses, and lists at least one file in each exported class (tokens, fonts,
   guides, generated pages). An empty or missing class fails.
2. The set of files in the vendor directory equals the set of files in the lock. A missing file
   and an unlisted file each fail with the path named.
3. Every file's SHA-256 matches its lock entry.
4. The sync script exports from the tag object (for example `git archive <tag>`), never from the
   checkout's working tree, and refuses a ref that isn't a tag or a tag that doesn't resolve to
   the commit it records.

Add one sentence naming the limit: the lock detects edits in the site repo, and provenance
against the tag is verified by re-running the sync script on the workstation (a `--verify` mode
that re-exports and diffs), not by the site's build.

### C2 (major, OWNER FORK on scope): "can't drift" covers the site-to-tag seam but not the guide-to-code seam

Location: spec lines 20-21 (ruling 2), 79-83, 88-106.

Ruling 2 promises the docs can't drift. The mechanism guarantees two things: the generated pages
match the code at the tag, and the site's copy matches the tag. The hand-written guides (Set up a
server, Run your server, For members) can still drift from the code inside dubplate: a guide can
name a renamed command, a removed flag, or a config key that no longer exists, and every check the
spec names passes. Those three pages are the ones the primary reader follows step by step.

The pin bump is also a manual checklist step in a checklist that doesn't exist yet (the repo has
no release tags and no release checklist). Nothing fails if the site's pin lags the newest
release.

Options:

- (a) Narrow the promise. The spec says the site can't drift from a tagged dubplate release, and
  hand-written guides are kept current by release review. Cheapest; leaves the operator path
  unchecked.
- (b) Add a name-existence leg in dubplate's `scripts/check.sh`: every command, flag, and config
  key named in `docs/guide/` in code formatting must exist in the cobra tree or the key registry.
  The generators already walk both, so the leg reuses them. Plus a site-side check at Phase B and
  at each release post that the lock's tag equals the newest dubplate release tag.
- (c) Option (b) plus executable guide steps (a smoke test that runs the setup guide's commands
  against a scratch box). Strongest and much more expensive; it overlaps `packaging`.

Recommendation: (b). It closes the most likely drift (a renamed command or key) at the cost of one
gate leg built on generators Phase A already writes, and it keeps the promise in ruling 2 true in
the sense a reader would take it.

Whatever the ruling, fold two more sentences: the Phase A dubplate lane creates the release
checklist file with the pin-bump step, and Phase B's precondition "the guides describe the tagged
0.8 release" is checked as "the lock's tag is the 0.8 release tag".

### C3 (major): the generated pages have no freshness criterion, and one source doesn't exist yet

Location: spec lines 58-59, 79-83, 96-99.

The sync script "exports the owned files from a dubplate release tag". For the three generated
pages to be exported, their output has to exist at the tag, or the sync script has to run the
generators against the tag's checkout. The spec doesn't say which. If the output is committed in
dubplate with no check that it is current, a tag can carry a stale reference and the lock will
faithfully pin the stale file.

Separately, the reject page "builds from `internal/reject`", which exists today only under
`_salvage/`. The spec treats it as present.

The reject page also promises "every reject reason". That is only checkable if the reasons form a
closed, enumerable set in one place.

Proposed fold:

- Generated output is committed in dubplate (for example `docs/guide/reference/`), and a
  `scripts/check.sh` leg regenerates it and fails on any diff. That makes a stale tag impossible
  and keeps the sync script free of a Go toolchain step.
- The reject page's precondition is that the greenfield tree carries the reject wording as a
  closed set (a table or enumerated type) the generator walks. Name the rung that lands it, or
  say the reject generator waits for it. Completeness then holds by construction, and the
  generator's test proves it can fail (a reason added without wording fails the build).
- The command and config generators state their source of truth the same way: the cobra tree
  from the root command, and the registry's declarations, including reserved and superseded keys
  or explicitly excluding them.

### C4 (major): Phase A has no done-definition

Location: spec lines 126-130, 120.

Phase A lists its parts but not what makes each one done. A reader can't tell from the spec
whether Phase A ends with eleven pages or one: it names "the landing page" but the site map has
eleven pages, and six of them are dubplate-owned or generated content that Phase A's own dubplate
lane writes. The guides also describe 0.8, which won't exist when Phase A runs, so Phase A's
guides are necessarily drafts.

Proposed fold. Add a Phase A acceptance list, each item with its check:

- The scaffold builds and deploys to the `workers.dev` address, and that address serves
  `noindex` (verified by fetching the page and reading the `X-Robots-Tag` header or robots meta,
  plus `robots.txt`).
- The sync script and gate exist and each C1 check has a mutation proof.
- The three generators exist, each with the freshness leg from C3.
- Every page in the site map exists as a route. Say which pages may carry draft or placeholder
  content at Phase A's end (recommended: the dubplate-owned guides are drafts against the planned
  0.8 surface; the generated pages are real output from whatever the tag carries).
- The trial record is committed with an entry for each admin-track page followed, including
  pages that produced no friction (so an empty record can't read as a clean trial), and its
  findings are filed through `engine-consult`.
- The is/isn't lines have moved: the principles spec points at the `docs/guide/` file and no
  second copy of the lines remains (a grep can check this). The trigger "after the principles
  spec's own rulings settle" (line 68) should name its observable signal, such as the principles
  spec reaching ratified status.

Add the mirror at Phase B: the production domain serves without `noindex`, and the `workers.dev`
address either keeps `noindex` or redirects.

### C5 (minor): "cairn's editor can't change them" names no mechanism

Location: spec lines 46-48.

The ownership boundary for vendored pages is asserted, not enforced. If the vendored guides sit
inside a cairn content collection, the editor may list them, and an editor publish commits
straight to the site's main branch. The C1 hash gate catches that edit only if the gate runs in
the path that deploys, not only in a local pre-commit check.

Proposed fold: state both halves. The vendor directory sits outside every collection cairn's
editor is configured to edit, and the hash gate runs in the deploy build, so a commit that touches
a vendored file fails to deploy. If cairn's site template gives no hook for a build-time check,
that is a trial finding for cairn, not a local patch.

### C6 (minor): the exported file set is declared only on the consumer side

Location: spec lines 90-99, 108-110.

The sync script in the site repo decides which dubplate files to copy. dubplate's owner has no
record that `inks.css` or `docs/guide/setup.md` is an export, so a rename in dubplate breaks the
site at the next sync with no warning in dubplate's own gate. The ownership boundary "dubplate owns
these" has no artifact in dubplate that says which files "these" are.

Proposed fold: dubplate carries a small export manifest (one file listing the exported paths),
the sync script reads it and fails on any listed path missing at the tag, and dubplate's gate
fails if a manifest path doesn't exist. This also settles the "which theme files travel" question
(lines 108-110) in a checkable place.

### C7 (minor): the font files' licenses and URL paths aren't in the travel list

Location: spec lines 90-94, 108-110.

The fonts are OFL-licensed, and the OFL requires the license to travel with the font files; the
license texts sit in `web/static/fonts/licenses/`. `fonts.css` also references the fonts at the
absolute path `/fonts/...`, so the "no translation step" claim holds only if the site serves them
at that same path.

Proposed fold: the travel list includes `web/static/fonts/licenses/` and `FONTS.md`, and the spec
states the site serves the vendored fonts at `/fonts/`. The About page's credits can then cite the
vendored license files.

### C8 (minor): the comparison page's re-verification trigger is undefined

Location: spec lines 72-74.

"Re-verified before each release that touches it" has no test for "touches it", so a claim's
checked date can age indefinitely.

Proposed fold: every claim carries a citation URL and a checked date in structured form, and the
check before each site release is that no checked date predates the previous site release (or a
fixed age the Phase A plan picks). A lint on the page's frontmatter or markup makes it mechanical.

## Not raised

- The two open rulings (domain, how a stranger gets the software) are correctly held as Phase B
  preconditions and aren't re-raised.
- "Every page names which of these readers it serves" is satisfied by the table's Reader column.
