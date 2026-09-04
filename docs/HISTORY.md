# dubplate history (born musicbox; entries keep the names of their day)

Per-pass ledger, newest first.

## Rung 2, pass 2c (Real-catalog and adversarial strata): 147 new cases, six oracle-bug hypotheses, first adjudication batch (2026-09-03)

Twenty plan tasks executed across a single sequential pass-execute
chain against
`docs/superpowers/plans/2026-09-03-rung-2c-catalog-adversarial.md`
(revision 2). Commits `3bbdfbd` → `890bb70` (Task 20 close at `f5bb1bd`; the
locale re-base and 19b prose fix at `890bb70`).

What landed: the adjudication and known-oracle-bug ledger formats
(Task 1); the `harness seed-items` projection and `--facts` tools
(Task 2); the real-catalog selection dossier, 96 primary rows plus 12
reserve rows across seven axes (Task 3); twelve capture batches (Tasks
4-15) landing 94 real-catalog primary cases against real MusicBrainz
document shapes (two two-release sub-axis rows lost to the mechanism's
own escape hatch, real captures over MB-noise, non-Latin script,
classical/compilation, genre/era spread, reissue/remaster, and
VA/box-set releases, one `ignore_video_tracks` finding); the
adversarial stratum (Tasks 16-18), synthetic cases reproducing two
known string-heuristic beets defects (the `"ft"` word-boundary false
positive and the parenthetical/remix near-duplicate discount, ten
cases, both moving strong/apply to medium/skip) plus six
per-pair track-assignment rejection-gap cases (the album verdict never
moves even while `assignment`/`applied_fields.title` are provably
wrong) plus the alias-mis-selection hypothesis probed against two of
Task 4's own real-nonlatin cases, plus Task 18's `locale` profile and
its A/B regression-check case pair for `#6358`; the stock-config
stratum (Tasks 19a/19b), 36 siblings (28 mandatory + 8 supplementary)
proving zero verdict-level divergence between `estate` and `stock`
across every real-catalog axis, with the divergence landing only in
`applied_fields` (`year` for reissues, `tracktotal` for box sets).
Corpus total: **215 cases** (7 smoke + 61 structural + 94 real-catalog
primary + 17 adversarial + 36 stock-config), read off the tree
directly. Determinism: two independent uncached (`-count=1`) full
`make derive` runs against the whole `internal/harness/...` and
`cmd/harness/...` tree, 215 and 215 cases actually re-derived of 215
committed (`TestCorpusReDerivesToItsCommittedRecords`'s own per-case
subtests, the count that proves a case was actually re-derived rather
than skipped for want of a committed record), both runs exiting 0 with
zero failures across every package, both byte-identical to the
committed records. Wall time: run 1 1729.58s (28m50s) total (corpus
derive 927.91s, ties derive 786.04s, component derive 1.22s); run 2
1582.53s (26m23s) total (corpus derive 812.66s, ties derive 754.48s,
component derive 1.27s) — this is rung 3's budget line, exactly as
2b's was 2c's own (2b's own pair ran ~2m13s each on 68 cases; 2c's
real-catalog and box-set growth moved that budget line by roughly
13x). The heaviest single cases in either derive phase are the
box-set axis's largest releases: ties derive's own re-solve-per-pair
algorithm (`internal/harness/oracle/derive_ties.py`) pays for
`real-boxset-rachmaninov-complete`'s 470 tracks and its stock twin
individually, each taking on the order of a minute alone, the same
box sets that dominate the byte-growth figure below. Corpus byte
growth (m7): the committed corpus held 4.0 MB across 114 response
files before this pass; it now holds **81.3 MB across 351 response
files**, well past the plan's projected 10-15 MB
landing point. The overshoot is a response-*size* story, not a
response-*count* one (351 files is close to what the plan's own count
math expected): a handful of the VA/box-set axis's largest releases
serialize their entire tracklist inline in one WS/2 document
(`real-boxset-rachmaninov-complete`, 470 tracks, 9.1 MB alone;
`real-boxset-scriabin-complete`, 321 tracks, 3.1 MB;
`real-boxset-mahler-complete`, 133 tracks, 1.7 MB), and Task 19a's
mandatory-sibling `cp -r` rule doubles every one of them for its stock
twin.

What the gate caught: nothing new at Task 20 itself (README-only and
analysis-doc changes, no code); across the pass, the recurring catch
was a case's own `claims.json` failing to notice a degraded record
(the claims-checker's whole design point) and, once, a stratification
undercount in Task 19b's own README prose (fixed same-pass, commit
`9915907`).

What a later pass would be wrong to rediscover:

- **File-level derivation is 2d's Task 13, unchanged in scope**:
  `internal/harness/oracle/derive_file_tags.py` and its Go driver,
  reading back independently through `mediafile.MediaFile(path)` after
  `item.write()`, never trusting the write call's own return value,
  over 12 subject cases (5 from the committed corpus, 7 flagged by
  this pass's dossier).
- **The audio-fixture route is ruled** (M4): vendor one small static
  template FLAC blob under `internal/harness/oracle/testdata/`,
  committed, with its provenance and byte size recorded in the README.
  `make derive` gains no binary dependency; `ffmpeg`/`flac` are not
  adopted as prerequisites.
- **`derive_file_tags.py` is a second, independent entry point** (M5),
  shaped like `derive_component.py`: it reads a committed record's
  `AppliedFields`, builds beets items from those fields, allocates its
  own scratch directory, writes and reads back there, and is never
  called from `oracle.Derive`. The invariant: no item in the corpus
  derivation path ever gains a non-empty `path`, proved by `make
  derive` re-deriving all committed records byte-identically after 2d's
  change.
- **`file_tags.json` needs an oracle-tagged re-derive-and-diff test**
  (M3), `TestFileTagsLedgerMatchesTheCorpus`, matching
  `TestTieLedgerMatchesTheCorpus`/`TestComponentTableReDerivesToItsCommittedTable`'s
  precedent, plus the `-update-ledgers` regeneration path documented in
  `corpus/README.md`. A gate-resident structural check alone (case ids
  resolve, pointers resolve) is not enough; nothing else would catch a
  `mediafile` version bump moving the tag shapes.
- **`Reconcile` lives in `internal/harness/oracle/reconcile.go`, not
  `oracle.go`, and its `versionProbe` tuple names five components
  today** (m2: beets, unidecode, jellyfish, lap, numpy), not four. 2d's
  Task 13 must edit `reconcile.go` and `versionProbe` explicitly when
  it adds `mediafile` to `Manifest`, or a reconcile silently compares
  against a missing key.
- **The version-bump differential and quarterly re-capture chore docs
  are unwritten**, deferred with Task 13. The differential doc's
  wall-time section must use the M8 inter-request-floor formula
  (`1.04s + 0.25s × (requests − 1)`) at the corpus's real size once 2d
  closes, not a flat per-case figure.
- **The whole-corpus README rewrite happens again at 2d's close**,
  once, to add the file-tags section; this pass's rewrite is not
  wasted work, since 2d's version extends it rather than replacing it.
- **The response-status extension was deliberately not taken this
  pass**: the stub has no third outcome and the schema carries no
  response status, because all four string/assignment/alias
  known-oracle-bug classes this pass probed are distance or selection
  defects, not transport defects — none needs a 404 or 503 to
  reproduce. It carries forward as a named non-adoption, available to
  whichever later rung first has a case that genuinely needs it.
- **The `profile_hashes` gap (m1)**, distinct from the 2d file-tags
  items above: `corpus/manifest.json` gained a `locale` entry under
  `profile_hashes` at Task 18, but nothing in the gate or the
  derivation path computes or checks it yet for *any* profile. A
  profile edit does not fail loudly the way a beets version pin
  already does; this is scoped follow-on work, not something the
  table's presence implies is built.
- **Tier C exclusions and profile gaps 2b carried that 2c did not
  close**: multi-source ordering and most of the `data_source`
  parameter space stay blocked by the single-source ruling (unchanged
  from 2b); `match.required`/`match.ignored` candidate-dropping stays
  unprobed (same shape, still open); `preferred.original_year: yes`
  stays unreachable under all three profiles now committed (`estate`,
  `stock`, and Task 18's own `locale`, which only toggles
  `import.languages`/`import.ignored_alias_types`) — a fourth profile
  or an extension to `locale.yaml` is still owed if a later rung wants
  this branch probed.
- **The alias-duplicate incident (Task 17) and its standing rule**: an
  earlier pass of Task 17 captured four byte-identical `adv-alias-*`
  cases duplicating releases the corpus had already committed under
  Task 4's `real-nonlatin-*` rows, because case selection never
  checked the committed tree first. The conductor's escalation ruling
  folded the four duplicates back into their existing twins (commits
  `e805540`, `c685894`); nothing was lost, but the rework was pure
  waste. **The rule a later pass must inherit: check `corpus/cases/`
  before any capture**, real or synthetic, that might already exist
  under a different id.
- **The `_preferred_alias` unreachability finding and Task 18's
  resolution of it**: Task 17's alias-mis-selection probe found
  `_preferred_alias` (`beetsplug/musicbrainz.py:128-155`) structurally
  unreachable under both `estate` and `stock`, since both leave
  `import.languages` empty, so its generator yields nothing regardless
  of what a release document carries. This was left as a scoping
  finding, not a reproduction, and deferred to Task 18, which added the
  `locale` profile and an A/B case pair
  (`real-nonlatin-utada-first-love` / `-locale`) to probe it directly.
  Task 18's first cut based `locale.yaml` on `stock.yaml` (as the plan
  itself prescribed), which the review caught as a confound: the pair
  carried the whole estate/stock delta besides the locale keys. The
  conductor ruled the clean experiment and the profile was re-based
  same pass (commit `890bb70`): `locale.yaml` is now `estate.yaml`
  plus exactly `import.languages: [ja]` and
  `import.ignored_alias_types: []`, so every divergence in the frozen
  pair is attributable by construction. The result: the mechanism *is*
  reachable and *does* fire once `import.languages` names a matching
  locale, moving the four sort-name fields it owns (`artist_sort`,
  `albumartist_sort`, `artists_sort`, `albumartists_sort`) on all 12
  items, plus one genuine bonus finding reported verbatim: the
  MusicBrainz "Various Artists" *entity* itself carries a primary
  `locale=ja` alias, so `_preferred_alias` also fires for two
  VA-credited losing candidates and moves their scoring inputs
  slightly (no verdict effect). Everything else in the pair is
  byte-identical — evidence against `#6358`'s report, pending Geoff's
  own ruling (`bug-locale-alias-ignored`, adjudication batch 1).

Carried:

- **The pending adjudication batch is the launch gate's open item.**
  Nineteen rows across five hypotheses sit at
  `disposition: "pending-adjudication"` in
  `corpus/analysis/adjudications.json`
  (`docs/superpowers/research/2026-09-03-adjudication-batch-1.md` is
  the artifact prepared for Geoff's ruling). This pass's own gate
  accepts them indefinitely; it is the **launch gate**, a different
  gate at a later time, that refuses an unadjudicated divergence.
  `bug-threshold-boundary` (`#5458`) is the corpus's one hypothesis
  already ruled (`oracle-correct`, Geoff, 2026-09-03) and is not part
  of the pending batch.
- `corpus/analysis/divergences.md` stays empty at this pass's close,
  reviewed and carried forward: no case across the real-catalog,
  adversarial, or stock-config strata contradicted a specific beets
  test assertion it was modeled on (findings of that shape instead
  became adjudication-batch rows or plain README prose).
- The two-release sub-axis closed at 4-of-6 primary rows delivered (2
  discarded via the mechanism's own escape hatch, the sole reserve
  spent by Task 4); a later two-release row anywhere in the corpus
  should budget for the same WS/2 search-ranking risk with no reserve
  behind it.
- Every item from 2b's own carry-forward that 2c did not separately
  close above (Tier C, `match.required`/`match.ignored`,
  `preferred.original_year`) stays open into 2d and beyond.

Budgets: ≈8.2M subagent tokens total (≈6.5M through Task 17's fold,
1.4M for Tasks 18-20's chain, 0.27M for the close fixes) against the
8M post-Task-3 re-derived ceiling (12M was the plan's pre-re-derivation
starting figure; finishing past the 80% flag approved by Geoff at the
flag, 2026-09-03). Attended interaction points: one mid-pass (the
combined 80%-flag question); the adjudication batch is delivered, not
yet spent.

## Rung 2, pass 2b (Structural stratum): analyses, synth, 61 synthetic cases (2026-09-02/03)

Fourteen plan tasks plus one staleness pre-task, executed overnight
across seven pass-execute workflows against
`docs/superpowers/plans/2026-09-02-rung-2b-structural-stratum.md`
(revision 2). Commits cec3c12 → 2219b19. All tasks accepted; five
needed one fix round; four escalations, each ruled by the conductor
and recorded in
`research/2026-09-02-rung-2b-staleness-corrections.md`.

What landed: the machinery half (threshold/weight echo with schema v2
re-freeze; `harness freeze` with `Options.Dir`; `harness synth` and
the authoring loop; the proximity flag, tie detector, per-case claims
checker, and 72-row component derivation table; the shared
`mbrequest` package), then the 61-case structural stratum across all
15 shape families, every record derived through real pinned beets,
plus the claims schema's zero-candidate mode and `penalty_keys_absent`
field. Corpus total: 68 cases, all re-deriving byte-exact across two
independent uncached runs (~2m13s each: corpus 70.5s, ties 54s,
component 0.7s; this is 2c's budget line).

What a later pass would be wrong to rediscover: the staleness review
pattern paid for itself (a fresh-context read of the case tasks
against the machinery as shipped found five blockers before any
dispatch; plan staleness remains the dominant defect class); the
"combined" completeness shape is unexpressible (`lapjv`'s leftover
set is one-sided); the VA-misflag has NO record-level observable
under the empty-search convention, so its witness is the frozen
request log's `arid:` query, which `make derive` mechanically
enforces; the response hand-edit escape hatch (edit generated
responses BEFORE first freeze, named in the README row) is the
sanctioned route for shapes `ReleaseSpec` cannot express; a
sub-second `make derive` is a cache hit and proves nothing, so force
`-count=1` for evidence.

Carried: two smoke cases tied (pinned; rung 3's solver must mirror
lapjv tie-breaking); family 15 parks cases on thresholds with
acknowledged proximity rows; the candidate-absent country direction
documented as declined; claims formatting drift cosmetic. 2c's
inventory is unchanged from the 2a plan's split section, plus the
determinism budget line above.

Budgets: ≈5.8M subagent tokens against the 6M ceiling (Geoff's 12M
overnight grant untouched); human interaction: two batched bedtime
questions (ceiling, 2c scope), zero mid-execution.

## Rung 2, pass 2a (Cassette substrate and capture): schema, stub, shim, capture tool, smoke corpus (2026-09-02)

Executed under two pass-execute workflows (wf_5ed5b773, wf_fe30d704)
plus one escalation follow-up, against
`docs/superpowers/plans/2026-09-02-rung-2a-cassette-substrate.md`
(revision 2). Commits 79420d3 → 69a1804. Five plan tasks, all accepted;
one fix round (task 5); one escalation, ruled in-scope and closed by
follow-up commit 6d45c28.

What landed: `internal/harness/cassette` (the frozen schema, verbatim
response-file store, byte-wise canonicalization rule);
`internal/harness/stub` (replay server with a miss list) and
`internal/harness/oracle` (adapter proven against a fake shim, then the
real `derive_oracle.py` driving pinned beets 2.13.1 in a uv venv, with
internal self-verifying date pinning and a raw-stdout determinism
proof); the two config profiles with their documented transform;
`cmd/harness` (capture tool: `--mbid` seeding with maximal `inc=` and
the pseudo-release double fetch, `--case` miss consumption, 1 req/s
identified wire discipline, contact refusal); and a seven-case live
smoke corpus (clean-match, should-ask, should-reject, various-artists,
multi-disc, non-latin pseudo-release with its transl-tracklisting
follow-up, year-penalty), every record derived through the real shim,
all seven re-derivable from the committed tree
(`TestCorpusReDerivesToItsCommittedRecords`) with the date-pinning
proof riding the real year-penalty case. Gate grew ruff and gitleaks
legs; `make derive` runs the oracle-tagged tests.

What a later pass would be wrong to rediscover: the `inc=` join is
`%2B` on the wire, never a literal `+` (requests percent-escapes it;
every live case missed its primary lookup until the capture tool's
canonicalization matched, and the original cassette doc's "literal +"
wire fact was corrected in 6d45c28; do not "fix" it back); the stub
must key on `r.URL.EscapedPath()`, since `r.URL.Path` arrives
percent-decoded; the miss loop's contract is `<case dir>/misses.json`,
written by `Derive`, consumed by `harness capture --case`; a
strong ID match legally orphans search captures into `Extra` (recorded
fixed point, not a stable request set).

Carried to 2b/2c: three non-blocking hardenings on
`corpus_derive_test.go` (skip-all-green residual hole, unguarded
`Candidates[0]`, redundant equality branch); the line-pinned
`.gitleaksignore` fingerprint; git-history gitleaks scanning (deferred
as a heavier job); response-number gaps in two cases (documented,
harmless).

Budgets: ≈1.75M subagent tokens against the 5M ceiling, plus conductor
context. Human interaction: one batched ask (the MusicBrainz contact,
minted as `DUBPLATE_HARNESS_CONTACT` in `~/.bashrc` via dotfiles
3a1c551; registration declined, anonymous 1 req/s stands).

## Rung 1 (Foundation): the repo pivot, config core, check/doctor (2026-09-02)

The greenfield build's first rung, run under the pass-execute workflow
(implementer/diff-reviewer/gate chain per task, owner opt-in). Landed
against `docs/superpowers/plans/2026-09-02-rung-1-foundation.md` and the
ratified build spec (`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
revision 2.1): the three Open rulings that spec's review round raised
(shadow floor, oracle-bug policy, stock-config gating) were all resolved
by Geoff the same day, closing the item STATUS carried as "Open rulings
1-3 await Geoff."

Five tasks: the pivot commit tagged `pre-greenfield` and moved `cmd/`,
`internal/`, `web/`, and `e2e/` under `_salvage/`, freezing them
read-only, and stood up `cmd/dubplate-server` as the module's only Go
package; the ops audit removed `config/filebrowser.yaml` as an orphan
and made `scripts/deploy.sh`'s binary step additive (rsyncs to the new
`/usr/local/bin/dubplate-server` path, never touching the incumbent
`/usr/local/bin/dubplate` the four guardian units still exec); the
config core (`internal/config`) implemented the derived wall as code:
owned, reserved-for-a-rung, and superseded key classes, plus a
threshold registry, all contract-tested from a stub before
implementation; `check`
and `doctor` scaffolds gave the wall an observable surface with a pinned
line format; this ledger close reconciled ROADMAP, CLAUDE.md, and
`.claude/settings.json` against the pivoted tree.

What the gates caught and a later rung must not rediscover:
`scripts/vale-comments.sh`'s file-list union needed an explicit
`_salvage/` exclusion, because git's `*.go` pathspec matches at any
depth and would otherwise re-admit all 76 salvage Go files into the
prose-comment scan; `go mod tidy` pruned the module to cobra's set once
the salvage tree's other dependencies left the build; the additive
deploy boundary (new path beside the untouched incumbent binary) is the
mechanism that lets rung 1 land without any box write, and every rung
before the launch cutover (rung 16) must preserve it. No box access
occurred this rung, per the spec's box-safety section.

Post-close fix trail: task 4's review escalated a silent-error defect
(SilenceErrors set with no stderr print in main). Three commits closed
it under review—a531d69 (main prints the error once; check stops
double-printing), 1f3ef3f (doctor returns "configuration is
incomplete" instead of writing the raw load error), baac9f4 (the
no-usage-block contract gains a live test, mutation-proved)—and
db3d11b landed the conductor-dispatched code-simplifier pass. Two more
lessons for later rungs: cobra's usage writer falls back to os.Stderr
when only SetErr is wired, so command tests must wire SetOut to
observe a leaked usage block; and doctor's row expectations must
derive from config's own tables (owned, reserved, Thresholds()),
never a literal count. Budget: about 1.0M subagent tokens for the
workflow (12 agents, five tasks) plus about 0.3M across the fix and
verify dispatches, against the 5M ceiling; zero attended interactions
during execution. Verdict trail: T1-T3 accept first-pass; T5 accept
after one fix round; T4 escalate plus two fix rounds.

## Uploader Pass B close + the greenfield pivot (2026-09-01/02)

Pass B built and hardened the uploader web app end to end, then Geoff
superseded its launch with the greenfield dubplate-server ruling. What
landed in code (all on master, salvage tier): U9-U11 web app; the U13
review workflow (57 agents, 45 confirmed serious findings from 5
reviewers with adversarial verification) and its complete fold F1-F8
(TagProbe built and wired -- the probe never existed and no main-lane
upload could reach the library; serve security majors; reject-copy
contract; semantic theme tokens; App.svelte structure; component fixes;
saturated palette + soft surfaces; state-text lightness + rejected
CTA); a three-round visual-fidelity chain ending PASS; code-simplifier
consolidation (AlertCard, Store access discipline). Deployed to the box
2026-09-02 with "907 Music" minted, then wound down the same day.

What the gates caught that later work must not rediscover: DaisyUI 5
theme blocks define NO semantic color defaults for custom themes
(navidrome-* themes shipped without --color-error et al. and every
status surface rendered gray); DaisyUI 5 .alert is display:grid (flex
utilities are inert; single wrapper child required); -soft variants
color text with the raw token, so dark themes need Material-400-class
state tokens; the beets probe cannot use --pretend (it skips MusicBrainz
entirely) -- the real dry-run shape is in internal/beets/probe.go; the
box serve unit needed AF_NETLINK debate but not the family (glibc falls
back); find -newermt with a bare relative string fails silently in
guard scripts (use -mmin).

The pivot: Geoff's device check hit the shared-Access-org ASC branding
on the login page (a page never rendered before asking him -- now the
see-it-before-geoff-tests-it memory). The fix path escalated by
ruling into: magic-link login + Navidrome SSO (spec
2026-09-02-magic-link-login-design.md; two adversarial reviews in
docs/superpowers/research/2026-09-02-login-spec-reviews/ found 6
security blockers + 3 universality criticals -- all now birth
requirements), then the full greenfield ruling set
(docs/superpowers/research/2026-09-02-greenfield-server-rulings.md).
Interim estate wound down 2026-09-02: FileBrowser gone (container,
compose, env, DNS, Access app), uploader serve + nudge units disabled,
uploader tunnel rule + both Access apps deleted; box runs Navidrome +
cloudflared + guardian timers only. The 37 deferred U13 minors live in
docs/superpowers/research/2026-09-01-u13-findings.json and matter only
to salvaged components. The dubplate-organizer initiative spec was
brainstormed and four-review hardened the same nights (80f0f5f + later
amendments); the organizer spec + rulings doc are the greenfield
build's inputs.

## Uploader Pass A close: U12a live, gates 1–2 passed (2026-09-01)

The uploader went live at https://music.907.life/upload behind a
path-scoped Access app (Geoff relocated it mid-sitting from its own
hostname; the AUD survived the domain edit, so the minted secret held).
serve mounts its whole tree under /upload via StripPrefix with tusd's
BasePath threaded through Options.TusBasePath, because tusd bakes the
configured base verbatim into the absolute Location clients resume
against. Gates 1–2 were measured through the full production path by
logging in through Access with an emailed code read over JMAP and
driving fetch from the authenticated page: SSE heartbeats at exact 15s
cadence with no edge buffering, Last-Event-ID resume in 120ms, tus
offset persistence through an aborted PATCH. The gate found one defect
(tusd stamping http:// behind the tunnel; fixed same hour with
RespectForwardedHeaders).

What a later pass would be wrong to rediscover: Docker's host-gateway
magic resolves to docker0's gateway, never a named bridge's, so the
compose alias pins 172.30.0.1 literally; a compose network recreate
reattaches live containers WITHOUT service aliases (502'd Navidrome
until force-recreate; now a vps-conventions row); Access service tokens
pass Access but die at serve's roster-gated JWT check by design, so
machine verification goes through the email-code flow; the Access login
page wears the ASC org's default branding (Pass B fixes the app's own
name/logo). Pass B standing inputs live in
docs/superpowers/research/passb-{stack-brief,copy-deck}.md; the copy
deck carries seven pre-U9 revision items.

## Go rewrite close: box cutover and estate rename, T12–T14 (2026-08-31)

Geoff's no-users ruling collapsed the plan's staged migration (additive
env window, per-subcommand flips, soak-gated T13) into one evening: the
estate has a library worth protecting (backed up, verified) but no users
a botched cutover could hurt, so the windows bought safety nobody
needed. One continuous run: T12–T14 repo work through the chain (two
review rounds each; the reviewers caught a ReadWritePaths grant that
would have failed every backup — fixed by clearing staging CONTENTS so
the leaf grant holds — and a compose project-name float that would have
orphaned every container, fixed by pinning `name: musicbox`); one
deploy shipping the binary, nine units, and the 36-name
single-generation env at /etc/dubplate; doctor fully green first try;
all four jobs green on supervised first runs; old units/scripts/paths
removed from the box; pings checks re-registered as dubplate-* inside
the grace window (no false alert); `~/Projects/musicbox` →
`~/Projects/dubplate`. The soak runs post-hoc: overnight timers prove
unattended operation with the dead-man checks as witness.

What a later pass would be wrong to rediscover:

- Compose derives its project name from --project-directory's basename;
  moving the directory without pinning `name:` makes `up -d` orphan
  every running container and start duplicates. The pin carries a
  never-remove-without-migration comment in compose.yaml.
- A ReadWritePaths LEAF grant cannot survive os.RemoveAll on the
  granted directory itself (needs parent write). dubplate-backup clears
  contents instead; the unit comment carries the contract.
- Kept identifiers (Hetzner names, tunnel, compose project,
  /var/lib/musicbox, /srv/music) are documented in README; the estate
  rename is complete EXCEPT those, deliberately.
- `Paths.RunDir`/`DUBPLATE_RUN_DIR` is required by config.Load but read
  by nothing — dead config, escalated to the uploader pass's U1 rather
  than patched around the frozen inventory.
- Slug swap ordering that avoids false dead-man alerts: add new checks,
  re-mint URLs, redeploy env, prove one ping, THEN remove old slugs —
  all inside the shortest check's period+grace.

## Go rewrite, repo phase T1–T11 (2026-08-31)

The shell layer's replacement landed as one `dubplate` binary: sixteen
internal packages plus cmd and an e2e harness with stub beet/rclone/
sqlite3, all through the implementer/opus-diff-reviewer/gate chain via
two pass-execute workflows (wf_edc843fb, wf_048446d1) plus follow-up
dispatches. All eleven tasks accepted; consolidation added the
code-simplifier sweep, poplar's vale comment-prose gate (vendored
overlay, hard-fail), and tidy-check. Commits 6ec7aab → 4489c99. Spend
≈4.3M of the raised 5.5M ceiling (original 4M overrun approved by Geoff
at the 3.2M checkpoint); human interaction: plan approval (pre-session),
one combined overage/SMTP question, one per-step-sanction standing
grant. The chain caught real defects at every level: a config env-scan
panic, a vacuous TLS assertion, ping-fail exit semantics, two importer
shell-parity divergences, and a contract-wording regression (T5).

What a later pass would be wrong to rediscover:

- An implementer agent can DIE mid-task (API drop) leaving compiling but
  untested, unreviewed code; the workflow journal records nothing for it
  and neighbor tasks inherit red gates that look like their own
  failures. Signature: a task with no journal record + innocent
  escalations citing out-of-scope lint. Recovery: complete + review the
  orphan task explicitly; re-review the neighbors after.
- Dispatches carrying a reviewer's verbatim LOCATION/FINDING/FIX text as
  acceptance criteria went through at zero fix rounds; paraphrased
  criteria averaged one. Pre-extract findings, never summarize them.
- Rejection wording is per-extension via a path-aware builder inside
  RenderEmail (Friendly() cannot carry the extension; its signature is
  contract). The shell's ungrammatical "an WAV" is pinned deliberately.
- `internal/beets` carries `subsonicClient = "musicbox"` matching the
  live shell; it flips in T14's rename sweep, never earlier.
- review.Filer's dirs are direct destinations, not tree roots
  (documented at review.go:22): importer builds two differently
  configured Filers, one for filing, one for sweeping.
- The e2e ENOTDIR blocked-path fixture (a plain file where a dir is
  expected) is deliberate: chmod fixtures pass trivially under root.

A separate agent executed the bring-up runbook against the live box; the
close-out session verified steps 5–7 read-only from the box and committed
the outcomes. Seed result: nine ZIPs in, seven albums imported (Angel Du$t,
Bully, Cheekface, De La Soul, Nightmares on Wax, TV Girl, Wax Tailor — 7
albums / 105 items in the beets DB), two review-queued with reasons
(Aim "Hinterland (Remastered)", illuminati hotties "POWER") plus a TV Girl
zine PDF routed to review. Post-seed backup succeeded 18:06 UTC:
`library/` in R2 at 112 objects / 2.794 GiB, `state/` holding both DB
snapshots. Monitoring truth: import and backup success pings green; the
failure-ping path is confirmed dead on the box (curl 404 in both
`musicbox-ping-failure@` journals — the known pre-wave GET-/fail bug) and
the disk/navidrome check timers were never enabled by the pre-wave deploy;
both fixed by the interim deploy this entry accompanies.

What a later pass would be wrong to rediscover:

- Contributor ZIPs extract FLAT (loose `Artist - Album - NN Track.flac`),
  so beets needs `group_albums: yes`; without it the whole seed read as
  one 135-track candidate and skipped.
- Since beets 2.4, MusicBrainz is a metadata-source plugin: an explicit
  `plugins:` line without `musicbrainz` leaves the autotagger with zero
  candidate sources and every album skips.
- Hetzner Cloud blocks outbound 25 and 465 by default on new accounts;
  Fastmail submission works on 587 STARTTLS from the box.
- `safe_extract` needs bsdtar on the host; it was missing from cloud-init
  packages and broke the first seeded import (hand-installed on the box,
  cloud-init now carries it).

## Architecture refinement pass (2026-08-30/31)

Adversarial review of the estate as built, then a fix wave. Phase A: nine
finder lenses plus an opus verify wave over both repos, the spec chain, and
live Cloudflare state — 72 findings, 40 confirmed, 12 endorsed, 0 refuted
(ledger: ~/.dotfiles/docs/superpowers/plans/2026-08-30-musicbox-arch-refinement-findings.json).
Fix wave: 11 tasks + lock follow-ups through implementer/diff-reviewer/gate
chains, all accepted (commit 087df6c); pings hardened in its own repo
(40fdfa7). Spend ~6.2M tokens against a 6M ceiling (3% over, driven by an
uploader design brainstorm folded into the same session); human
interaction: one combined triage question, approvals otherwise.

What a later pass would be wrong to rediscover:

- A lock file in a systemd RuntimeDirectory is a lifecycle bug (deleted at
  unit stop, unlinked-inode races across units). Locks live in
  /var/lib/musicbox; per-run scratch keeps /run. Doctrine now in the
  vps-conventions skill — invoke it before touching units.
- ping_hc fail is POST; the Worker 404s GET /fail. Every failure ping ever
  sent before this pass was dropped.
- bash cannot re-arm errexit inside a command-substitution subshell; the
  M3 first attempt guarded notify() with a subshell set -e that did
  nothing. Best-effort is enforced by if-guarding at the definition.
- The beets database is the only import success record (tag log = skips
  only; beets renames everything). The uploader design builds on this.
- Review standard that caught real bugs: verify test non-vacuity by
  mutation (M10's reviewer), and re-run implementer claims yourself
  (M2b's reviewer caught a misdescribed unit).
- The estate renamed musicbox → dubplate (2026-08-31); repo/paths converge
  at rewrite T14. The rewrite, uploader, and packaging specs were authored
  and adversarially reviewed this same session.

Carried out of the pass: SELinux one-time check on /var/lib/musicbox at
first deploy; provision.sh's ssh_cidrs_json exits only its substitution
subshell on malformed input (diagnostic degrades to a jq error); pings
remove-check has no --force for a permanently orphaned route record.

## Spec gaps closed: rclone/msmtp provisioning, Subsonic scan trigger (2026-08-31)

Closed the two real spec gaps the first live deploy exposed (previous entry below),
without touching the live box.

- **`rclone`/`msmtp` were never in `cloud-init/user-data.yaml` at all.** Confirmed both
  are in EPEL for AlmaLinux 10 via a container check
  (`podman run --rm almalinux:10 ...dnf -q list rclone msmtp`) before adding them,
  rather than guessing an unattended `set -e` first-boot script into a package that
  doesn't exist: `rclone.x86_64 1.74.3-1.el10_2` and `msmtp.x86_64 1.8.25-6.el10_1`,
  both from the `epel` repo alone, no RPM Fusion or CRB dependency. Installed
  immediately after `epel-release` in the existing EPEL/RPM-Fusion runcmd block,
  ahead of the RPM Fusion key import flac/ffmpeg need, so a failed RPM Fusion step
  can't take them down too.
- **`trigger_navidrome_scan()` rewritten from the dead native `/api/scanner/scan`
  endpoint to the Subsonic API's `GET /rest/startScan`** (query-param auth:
  `u`/`t`/`s`/`v`/`c`, `t=md5(password+salt)`, a fresh random `s` per call). The
  credential-bearing params never touch argv: the full URL is written to a curl
  `--config` file (mode 0600) in a new `RUNTIME_DIR` (same directory as the import
  lock file -- `/run/musicbox` on the box, the fixture's `WORK_DIR` in tests), and
  the config file is removed via a function-scoped `RETURN` trap. Salt is generated
  via `od -An -N8 -tx1 /dev/urandom | tr -d ' \n'` rather than a `tr | head`
  pipeline, which would risk a SIGPIPE-driven pipefail failure. The existing
  failure-tolerant shape is unchanged: a failed request logs to stderr and returns
  0, never killing the run (the daily Navidrome cron scan is still the backstop).
  `tests/music-import.bats`'s shared curl stub now also captures `--config` file
  contents to a separate log (the real file is gone by the time a test could
  inspect it otherwise), plus a `CURL_FAIL_STARTSCAN` switch to exercise the
  failure path; two new tests cover the call shape and the failure tolerance.

Neither fix has reached the already-provisioned live box yet (cloud-init is
first-boot-only); see STATUS's "reprovision pending" open item.

## Task 4/6: first live deploy (2026-08-31)

`scripts/deploy.sh --check` then a real run against the provisioned box. Two env-template
vars didn't resolve as assumed going in: `NAVIDROME_ADMIN_PASSWORD` was stored under the
older name `NAVIDROME_ADMIN_PASS` (renamed in the age store to match, same value, no
rotation), and `MUSICBOX_HC_NAVIDROME_URL` had never been minted at all. Minted it via
`~/Projects/pings/scripts/add-check musicbox-navidrome --period 60 --grace 15` (mirroring
`musicbox-disk`'s already-accepted pre-minted-but-unwired state) so the render gate could
pass; it sits armed and unpinged until Task 6 builds the Navidrome-heartbeat script.

The deploy itself succeeded, but the acceptance sweep surfaced real bugs a fixture-backed
test suite can't see — **wrong to rediscover**, each one below:

1. **FileBrowser crash-looped**: `cacheDir` defaults to a `tmp` folder relative to the
   process's working directory, which isn't part of any bind mount and isn't writable by
   UID 2000. Fixed by pointing `server.cacheDir` at the already-writable, already-persistent
   `/home/filebrowser/data/cache` (`config/filebrowser.yaml`).
2. **FileBrowser's own first-run admin bootstrap then failed**: its built-in default admin
   password is shorter than our own `auth.methods.password.minLength: 8` policy, so account
   creation errored and the container kept crash-looping. Minted a new secret
   `FILEBROWSER_ADMIN_PASSWORD` (FileBrowser's native env var name) and wired it into
   `compose.yaml`'s `filebrowser` service; Task 8 replaces this bootstrap account with
   per-contributor ones.
3. **music-import.service failed on `/run/lock/musicbox-import.lock: Permission denied`**:
   `/run/lock` is root:root `0755` on Alma (unlike Debian's world-writable `1777`), so the
   `User=music` service couldn't create its flock file there. Added
   `RuntimeDirectory=musicbox` to the unit (systemd creates `/run/musicbox` owned by
   `music:music` every boot) and pointed the script's lock-file default there.
4. **music-import then failed on `curl: (7) ... Failed to connect to localhost port 4533`**:
   `trigger_navidrome_scan()` runs on the host as `music`, not inside the compose network,
   but `compose.yaml`'s `navidrome` service only used `expose` (container-network-only).
   Switched to `ports: - "127.0.0.1:4533:4533"` — loopback-only, so Navidrome stays exactly
   as unreachable from the internet as before (still only `cloudflared` reaches it
   externally), while becoming reachable from the box itself.
5. **Same call's error handling was also incomplete**: the login `curl` had no failure
   guard, so `set -e`/`pipefail` killed the *entire* import run on any Navidrome-unreachable
   condition, not just the graceful "skip the scan trigger" path the function's comments
   already described for a failed login. Added `|| true` to match.
6. **The native admin scan-trigger endpoint doesn't exist in the pinned Navidrome
   `0.63.2`**: confirmed against upstream source (`server/subsonic/library_scanning.go`,
   `server/subsonic/api.go` line ~181) — `StartScan` is registered as the *Subsonic* API's
   admin-only `startScan` action (query-param auth: `u`/`p`/`v`/`c`), not a native
   `/api/scanner/scan` JWT-bearer endpoint. Every `music-import` run now logs a 404 there
   after a real login success. **Not fixed this pass** — left as a known, non-fatal bug
   (Navidrome's own `ND_SCANNER_SCHEDULE: "0 3 * * *"` cron backstop still keeps the library
   current) rather than rush an unverified Subsonic-auth rewrite. Follow-up: rewrite
   `trigger_navidrome_scan()` to call `GET /rest/startScan` with Subsonic auth params.

**Two gaps could not be closed this pass, both blocked by the Claude Code auto-mode
classifier refusing an on-box `dnf install` from a subagent** (same shape as Task 5's
DELETE/create-tunnel refusal above — mutating infra calls need Geoff present, not a
subagent acting alone):

- **`ffmpeg`/`flac` are still absent on this box.** The EPEL + RPM Fusion `runcmd` block
  that installs them (`cloud-init/user-data.yaml`) was added to the repo *after* this box
  was provisioned (Task 4a), so it never ran here — a cloud-init-only fix never converges
  an already-running box. beets' `replaygain` plugin fails to load on every `music-import`
  invocation as a result (logged, non-fatal against an empty inbox; will break real
  imports needing the quality gate's `flac -t`/`metaflac` and replaygain analysis).
- **`music-backup.service` cannot run at all**: `sqlite3` and `rclone` are both missing on
  the box, and neither was ever listed in `cloud-init/user-data.yaml` in the first place (a
  real spec gap, not just an unapplied one). Added `sqlite` to the `packages:` list (base
  AlmaLinux repo, confirmed by lookup) for future re-provisions; `rclone` and `msmtp` (also
  missing, also never listed — `lib.sh`'s `notify()` needs it, untested this pass since an
  empty-inbox run never reaches it) are left as an open item rather than guessed into an
  unattended `set -e` first-boot script without a confirmed AlmaLinux 10 install path.
  Confirmed via a read-only `rclone lsjson` against `music-library` that `state/` is empty,
  consistent with no backup ever having run.

Acceptance sweep otherwise green: all three containers running with zero restarts after a
60s settle; the tunnel reads `status: healthy` with active connections; `music.907.life`
serves Navidrome's web UI with no Access interstitial and answers `/rest/ping`;
`inbox.907.life` presents the Cloudflare Access challenge before any FileBrowser content;
both timers active; the Navidrome admin account was created via `POST /auth/createAdmin`
and its login verified via `POST /auth/login` (API only, no secret values ever printed —
checked by name, length, and boolean fields throughout this whole pass).

## Task 5: Cloudflare tunnel wiring, plus a same-day secret rotation (2026-08-31)

Geoff added Cloudflare Tunnel:Edit to `CLOUDFLARE_API_TOKEN` 2026-08-30, unblocking Task 5.
An agent created tunnel `musicbox` (`POST .../cfd_tunnel`), then displayed that call's full
JSON response in a transcript while reporting progress. **Wrong to rediscover**: a
`cfd_tunnel` CREATE response body carries the tunnel's `TunnelSecret` and its connector
`token` alongside the id — displaying or logging the raw body exposes both, not just an
identifier. The fix at the call site is mechanical: always pipe the response through
`jq -r '.result.id'` (or whichever single field is actually needed) and never let the full
body reach a transcript, a log, or a report.

Once flagged, the tunnel was treated as compromised under standing incident doctrine (a
secret printed into a transcript is compromised, full stop — no risk-assessment exception).
Remediation was delete-and-recreate, the cheap definitive path since `cloudflared` was not
yet deployed anywhere against the exposed token:

1. Deleted tunnel `66fa450c-5abb-459d-a1f8-b80d8d19e07c` (`DELETE .../cfd_tunnel/{id}`).
2. Created a new tunnel under the same name `musicbox`, this time piping the response
   through `jq -r '.result.id'` only: id `b6781175-000b-40c3-a54a-05906ab25037`.
3. Fetched its token (`GET .../cfd_tunnel/{id}/token`) and piped it directly into
   `secret-set.sh MUSICBOX_TUNNEL_TOKEN --stdin`, overwriting the old value. Verified
   name-only (`grep -c` on `~/.local/secrets`, count 1) and by length only (240 chars) —
   never printed.
4. Re-PUT the ingress from `config/cloudflared-tunnel-config.json` to the new tunnel id and
   verified with GET: `music.907.life` → `http://navidrome:4533`, `inbox.907.life` →
   `http://filebrowser:80`, 404 catch-all.
5. Repointed both DNS CNAMEs (`music.907.life`, `inbox.907.life`, zone `907.life`) to
   `b6781175-000b-40c3-a54a-05906ab25037.cfargotunnel.com`, proxied.

**Process note**: the DELETE and the tunnel-recreate POST were both refused by the Claude
Code auto-mode classifier when a subagent attempted them (destructive/mutating infra calls
without the user present); read-only GETs against the same account worked fine throughout.
The rotation completed once the conductor ran those two calls from the main session with
Geoff present, then handed the new tunnel id back for the remaining wiring and docs.

Cloudflare Access was untouched throughout (confirmed read-only via `CF_ZT_TOKEN`:
`inbox.907.life` is the only Access app on the musicbox surface; `music.907.life` stays
Access-free). `cloudflared` is not deployed on the box yet, so both hostnames 530 until
Task 4's first real `deploy.sh` — expected, not chased.

## Task 2 follow-up: quality gate, ping hardening, config sync (2026-08-30)

TDD, red-first for every case, against the spec's freshly amended import-pipeline
section (step 4). Four items, one commit each:

- **`quality_gate <dir>`** (`scripts/lib.sh`): walks every audio file under dir
  (case-insensitive flac/mp3/m4a/ogg/opus/wav/aiff extensions; anything else is out
  of scope). A non-FLAC extension fails outright; a FLAC must pass `flac -t`'s full
  decode test and carry >=44.1kHz/16-bit (`metaflac --show-sample-rate
  --show-bps`). Prints "path\treason" per failure, exit 0 only if everything
  passed. Wired into `music-import` as `quarantine_quality_failures`, between the
  free-space abort and `beet import`: any failing file takes its **whole**
  containing album directory to `review/<date>/<user>/` (a partial import of just
  the good tracks never happens), never just the offending file.
- **Uploader copy is templated, not generated**: `config/reject-email.txt` is
  Geoff's verbatim consumer-facing copy with three literal placeholders (`{first
  name}`, `{album or file name}`, `{one reason line per file, from the mapping
  below}`), filled by plain bash substring replacement in
  `render_reject_email()` — no sed, no metacharacter risk. `friendly_reason()`
  translates `quality_gate`'s raw technical reason into one of three fixed human
  sentences (non-FLAC extension, failed decode, below CD quality); Geoff's own
  copy of the notice stays technical (raw path + reason), same asymmetry as the
  existing beets-skip notice. First name comes from `uploader_name()`, the first
  word of the matched `contributors.yaml` `name` field (mirrors `uploader_email`'s
  lookup). Default template path is `/opt/musicbox/config/reject-email.txt` —
  `config/` is already rsynced wholesale to that tree by `deploy.sh`, so this
  needed no new deploy step; only `contributors.yaml` and `beets-config.yaml` get
  dedicated ones, for permissions/install-location reasons this file doesn't
  share.
- **`ping_hc` hardened**: `--retry 3 --max-time 10` added to its `curl` call, so a
  slow or flaky healthchecks.io can never be the reason an otherwise-successful
  run hangs or gets misreported. The empty-vs-unset URL guard already existed;
  only a new bats case (env var set but empty) was added to pin it down.
- **Flock-held early exit now pings success before `exit 0`** (spec decision):
  a run that finds the lock already held (another scheduled run, or an
  interactive `music-beet` session) is the healthy case this check exists for,
  not a failure — healthchecks.io has no third state between success and
  failure. Accepted residual: a genuinely wedged import would also report
  healthy via this path until someone notices stale music in the library.
- **Beets config synced to the workstation's** (`~/.dotfiles/beets/.config/beets/config.yaml`):
  `original_date: yes`, `per_disc_numbering: yes`, and the `replaygain` plugin
  (backend `ffmpeg`), added to `config/beets-config.yaml` alongside the existing
  quiet-mode section. flac (the quality gate) and ffmpeg (replaygain's analysis
  backend) are installed on the host via a dedicated `cloud-init/user-data.yaml`
  runcmd step, not the top-level `packages:` list (that module runs before
  runcmd, too early to enable a repo first): `dnf -y install epel-release`,
  `dnf config-manager --set-enabled crb` (both EPEL and RPM Fusion need it),
  then the RPM Fusion free EL10 GPG key is fetched and `rpm --import`ed and
  its release package installed against it (gpgcheck on throughout, no
  `--nogpgcheck`), then `dnf install -y flac ffmpeg`. Ruling: both binaries
  parse untrusted contributor uploads, so they stay on dnf-automatic's
  security-update path rather than a pinned static build, which would freeze
  a codec-parsing attack surface instead of patching it.
- **Known limitation, accepted**: `quality_gate`'s output channel is
  tab/newline-delimited ("path\treason" per line), so a filename containing a
  literal tab or newline would misparse downstream in
  `quarantine_quality_failures`. The `find -type f -print0` walk itself is
  correct (NUL-delimited, handles any filename); the limitation is only in
  what quality_gate prints afterward. Not expected in practice for
  contributor uploads; revisit if it ever is.

Fixed a docs nit found while touching STATUS.md: Task 4a's on-box checks table
(above, Task 4a's own section) lists nine checks but was labeled "8 of 8"; now "9
of 9" in both STATUS.md and here.

Full test suite: `tests/lib.bats` (4 new `quality_gate` cases + 1 `ping_hc` case)
and `tests/music-import.bats` (1 new flock-ping case, 1 new quality-gate
end-to-end case; three existing fixtures switched from placeholder `.flac` text to
real ffmpeg+flac-encoded audio, since the gate now actually decodes them).
`bash scripts/check.sh` exits 0.

## Task 2 follow-up, fix round: Opus review of 5da3844..d43972b (2026-08-30)

Six items landed, TDD where a test was named:

- **Blocking: unquoted `&` in `render_reject_email`'s `${var//pat/repl}`
  substitutions.** Confirmed on this box's bash (5.3.9): an unquoted
  replacement containing a literal `&` gets expanded to the matched pattern
  text, exactly like sed — an album named "Sturm & Drang" rendered as "Sturm
  {album or file name} Drang" instead of the real name. Fixed by quoting
  every replacement value; bats case covers the ampersand literally.
- **Blocking: an `mv` into an already-occupied `review/<date>/<user>/<album>`
  aborted the whole run under `set -e`** (a same-day repeat quarantine of an
  identically named album is the common case, not an edge case). New
  `lib.sh` helpers: `unique_dest_path` (appends a `-HHMMSS` suffix on
  collision, then `.2`/`.3`/... if even that collides) and
  `safe_mv_to_review` (uses it, and warns-and-returns-nonzero instead of
  letting a failed move take the run down). Applied to both
  `quarantine_quality_failures` and `move_skips_and_notify`, which shared the
  same unguarded-`mv` shape. bats case: two identically named "Bad Album"
  quarantines in one test run both survive and beet import still runs both
  times.
- **A missing `reject-email.txt` template previously sent the uploader an
  empty-bodied email** (`render_reject_email`'s failure was silently
  swallowed by command substitution). Now an explicit `[ -f ... ]` check
  before the `notify` call skips the uploader-facing email and logs why;
  Geoff's technical summary still covers the album either way.
- **Each quality-gate reason line is now prefixed with its file's
  basename** in the uploader-facing email, so an album with more than one
  failing file reads as N per-file lines instead of running together.
- **Added the 8-bit-FLAC branch's missing test coverage** (`ffmpeg -acodec
  pcm_u8` piped into `flac`), the one `quality_gate` branch that had no
  case before this round.
- **CRB enablement** added to cloud-init's EPEL/RPM Fusion runcmd block
  (`dnf config-manager --set-enabled crb`, right after `epel-release`):
  both repos need packages CRB carries, and ffmpeg's dependency chain fails
  without it.

Follow-up verification round found one more leak in the `mv`-collision fix:
`safe_mv_to_review` failing for a **non-collision** reason (permissions, a
bad mount) still `continue`d out of the loop with no Geoff notice and the
album still sitting in staging — exactly what `beet import`'s unconditional
recursive scan of `$STAGING_DIR` would then ingest, silently defeating the
whole quality gate. Fixed with a second `safe_mv_to_review` attempt to
`$MUSICBOX_ROOT/quarantine-failed/<date>/<user>`, a holding dir outside both
staging and review's own tree (so a problem isolated to `review_root`
doesn't also block the fallback); Geoff's summary gets a `could not move;
excluded from import` marker either way, and a distinct all-caps marker for
the (untested, believed-unreachable without a fully read-only filesystem)
case where even the fallback fails. The holding dir's root is a top-level
constant (`QUARANTINE_FAILED_DIR`) beside `REVIEW_DIR` and the rest of the
tree, deliberately not created up front: a healthy box never has one.

Also fixed in `safe_mv_to_review` itself: its `mkdir -p "$dest_dir"` call
was unguarded, so a destination that couldn't even be created would have
aborted the whole run under `set -e` — the exact failure mode the function
exists to prevent. Added a matching
`[ -n "$geoff_summary" ] || return 0` guard before the final Geoff `notify`,
defense-in-depth alongside the pre-loop `album_order` check (every branch of
the loop already appends something, so this is currently unreachable, but
decouples "should Geoff be emailed" from "was there at least one album_key"
for whatever a future change to the loop might do). TDD: bats case forces
the primary move to fail via `chmod 0555` on the review dir and asserts the
album lands in `quarantine-failed`, never in staging, with the marker in
Geoff's email and no uploader email.

Also fixed, while adding coverage for the above: three existing
`tests/music-import.bats` assertions used `! grep ...` as a non-final line —
under bats' `set -e`, a command negated with a leading `!` never triggers
errexit regardless of its actual result, so a failing negated assertion that
isn't the test's last line silently does nothing. Converted to bats' `run` +
`[ "$status" -ne 0 ]` idiom throughout the file for safety, whether or not a
given instance was actually latent.

## Task 5/6: Cloudflare Access + Navidrome secrets, tunnel blocked (2026-08-30)

Control-plane pass, no box changes. Findings and what landed:

- **Tunnel creation blocked, not a code problem**: `POST
  /accounts/{a}/cfd_tunnel` 403s `{"code":10000,"message":"Authentication
  error"}` on both `CLOUDFLARE_API_TOKEN` and `CF_ZT_TOKEN` — neither carries the
  Cloudflare Tunnel:Edit permission group (confirmed: both tokens list/read
  tunnels fine, i.e. GET `.../cfd_tunnel` is 200; only the POST 403s). This
  blocks Task 5's steps 1-4 (create tunnel, fetch token, PUT ingress config, DNS
  CNAMEs to `<tunnel-id>.cfargotunnel.com`) and, downstream, Task 4's first real
  `deploy.sh` run (cloudflared needs `MUSICBOX_TUNNEL_TOKEN` to come up).
  Resolution needs Geoff to add Cloudflare Tunnel:Edit to one token via the
  dashboard (API self-management of token scopes is deliberately locked out).
  Detail in registry.md's `CF_ZT_TOKEN` entry.
- **Cloudflare Access, done**: confirmed no Access application exists on any
  `907.life` hostname (checked the full account app list — zero matches), so
  `music.907.life` correctly has none. Created a self-hosted Access application
  for `inbox.907.life` (`CF_ZT_TOKEN`, which — unlike `CLOUDFLARE_API_TOKEN` —
  can read Access identity providers) restricted to the account's existing
  `onetimepin` identity provider (`e33c7c2d-facd-446c-96de-8215f64cde76`,
  `auto_redirect_to_identity: true` since it's the only allowed IdP), with an
  allow policy for `geoff-login@907.life` (policy
  `92425260-87f6-4a3a-b66e-d71bb8d9d6a8`). Family emails join this policy at
  Task 8 per plan. **Not yet verified end to end**: the "Access challenge
  appears before FileBrowser's login" acceptance check needs the tunnel serving
  traffic first, so it waits on the tunnel-permission fix above.
- **Task 6 Navidrome secrets, done**: generated and stored
  `ND_PASSWORDENCRYPTIONKEY` (openssl rand -hex 32), `NAVIDROME_ADMIN_PASS`
  (openssl rand -hex 16), `NAVIDROME_ADMIN_USER` (`admin`) via `secret-set.sh`.
  No value printed; lengths only (64/32/5 chars) confirmed against
  `~/.local/secrets`.
- **registry.md, done**: added rows + per-secret sections for
  `MUSICBOX_TUNNEL_TOKEN` (marked not-yet-minted with the blocker noted),
  `ND_PASSWORDENCRYPTIONKEY`, `NAVIDROME_ADMIN_USER/PASS`, and the four R2 pairs
  already in the store (`MUSICBOX_R2_*` — musicbox-vps token, Object Read &
  Write, bucket `music-library`; `MUSIC_R2_RO_*` — claude-code-thinkpad-x1
  token, Object Read only, same bucket); noted `HCLOUD_TOKEN`'s 2026-08-30
  rotation. Committed to `~/.dotfiles` as two commits (registry.md text, then
  `values.age` separately) — a third file in the working tree,
  `bin/.local/bin/secret-receive`, was mid-edit from an unrelated session and
  deliberately left untouched/uncommitted.

Decision a later pass would be wrong to rediscover from scratch: **use
`CF_ZT_TOKEN`, not `CLOUDFLARE_API_TOKEN`, for anything touching Access
identity providers** — the general-purpose token gets `auth.forbidden` reading
`.../access/identity_providers`, even though it can read/write Access apps and
policies fine.

## Task 4a: AlmaLinux 10 image-defect fix round, rebuild to 8/8 green (2026-08-31)

Two deterministic image defects blocked the recreated box at 5/8 green (see the prior
round's entry below): Docker's missing `xt_addrtype` kernel module
([docker/docker-ce-packaging#1340](https://github.com/docker/docker-ce-packaging/issues/1340))
and the Hetzner `alma-10` image shipping `SELINUX=permissive` instead of the distro
default of enforcing. Both fixes landed in `cloud-init/user-data.yaml`'s `runcmd` plus a
new top-level `power_state` key:

- **Docker/kernel-module fix**: a new `runcmd` item installs
  `kernel-modules-extra-$(uname -r)` — the exact NVR paired to the boot's running
  kernel — before the Docker CE install, falling back to `dnf update kernel &&  dnf
  install kernel-modules-extra` only if the mirrors have moved past this image's shipped
  kernel (that fallback needs the reboot below to take effect). `systemctl enable --now
  docker` became `systemctl enable docker` — starting it before the reboot would still run
  under the old kernel and fail identically.
- **SELinux fix**: `sed` flips `/etc/selinux/config` to `SELINUX=enforcing` and `touch
  /.autorelabel` schedules a full relabel on next boot (the safe path — `setenforce 1`
  alone would trust every file on a vendor image already carries correct labels, which
  permissive mode never enforced).
- **Reboot ordering**: a single `power_state: {mode: reboot}` module — not a `reboot` call
  inside `runcmd` — runs after every other cloud-init module completes and has its
  once-per-instance completion recorded, covering both fixes in one reboot. The
  autorelabel triggers its own second, self-contained reboot; the acceptance run must
  tolerate SSH dropping twice.

First attempt at this round left the fix uncommitted while re-verifying against the live
box — caught by a review pass that diffed the deployed `/var/lib/cloud/instance/user-data.txt`
sha256 against `HEAD`'s copy of `cloud-init/user-data.yaml` and found them different
(`196e76f...` deployed vs `db0b633...` at HEAD). Lesson for any future fix-round: verifying
green against a live box proves nothing about the repo until the fix is committed and that
commit's own blob hash is re-checked against the deployed file — a worktree edit and an
unpushed commit look identical to `ssh`.

Re-verified after commit `3fa905d`: `sha256sum cloud-init/user-data.yaml` at HEAD equals
the deployed `/var/lib/cloud/instance/user-data.txt` hash
(`196e76fbc51d5a5828c6a378761d480b790160cc9fd07fa7a1f41ab1557dfa5a`). `bash
scripts/check.sh` exits 0.

**9 of 9 on-box checks green, re-verified 2026-08-31 against `root@5.78.148.42`:**

| Check | Command | Result |
|---|---|---|
| cloud-init | `cloud-init status --wait` | `status: done` — PASS |
| volume mount | `mountpoint -q /srv/music` | mounted — PASS |
| tree ownership | `ls -lnd /srv/music/{library,inbox,review,state,staging}` | all `2000 2000`; `inbox` is `drwxrwsr-x` (2775, setgid) — PASS |
| music user | `id music` | `uid=2000(music) gid=2000(music) groups=2000(music)` — PASS |
| docker | `systemctl is-active docker` | `active` — PASS |
| SELinux | `getenforce` | `Enforcing` — PASS |
| sshd password auth | `sshd -T \| grep -i '^passwordauthentication'` | `passwordauthentication no` — PASS |
| yq / jq | `/usr/local/bin/yq --version` / `jq --version` | `v4.53.6` / present (`jq-` is a cosmetic upstream version-string quirk on this AlmaLinux 10 build, not a functional defect) — PASS |
| fstab | `grep srv/music /etc/fstab` | exactly one line: `UUID=6614e3dd-9014-4949-9f4c-effcd317bd3a /srv/music ext4 defaults,nofail,x-systemd.device-timeout=30s 0 2` — PASS |

## Task 4a: real provisioning, forced-root-password fix round (2026-08-30/31)

`scripts/provision.sh --check` matched expectations (alma-10 available, no fallback
needed), then `scripts/provision.sh` created firewall `musicbox-fw` (tcp/22 only, both
address families), volume `musicbox-data` (50GB, `hil`), and server `musicbox` (CPX21,
`alma-10`, `hil`). Server IPv4 `5.78.148.42`, IPv6 `2a01:4ff:1f0:97c6::1`. CPX21 traffic
allowance at `hil`: 2199023255552 bytes = 2 TiB/month (`hcloud server-type describe cpx21
-o json`), overage $1.20/TB — generous for a family Navidrome, no need to default clients
to transcoded streams.

A fix round closed three blocking findings from a forced-root-password blocker, and the
server was recreated to pick them up:

1. `scripts/provision.sh` registers an `hcloud ssh-key` named `musicbox-workstation`
   idempotently and passes `--ssh-key musicbox-workstation` to `hcloud server create`.
   Confirmed on the recreate: no "Root password: ..." line appeared in the create output.
2. `cloud-init/user-data.yaml`'s `runcmd` carries a defense-in-depth guard — `chage -M -1
   -I -1 -E -1 -d "$(date +%Y-%m-%d)" root` — as the first item, so any future path that
   creates the server without a key can't reintroduce this blocker. Confirmed on-box:
   `chage -l root` shows `Password expires: never`.
3. The prior round's one-time-manual-SSH-unblock instruction was removed; nothing was
   hand-applied to a live box — the box was deleted and recreated from the fixed repo
   instead (volume `musicbox-data` survived and reattached, `blkid` guard prevented
   reformat, box got IPv4 `5.78.148.42` again).

Re-verification after that round found a new, unrelated blocker: `cloud-init status
--wait` returned `status: error` because Docker failed to start
(`docker/docker-ce-packaging#1340`, see the entry above), which also aborted the
downstream yq-install `runcmd` item. `getenforce` also reported `Permissive`, traced to
the Hetzner `alma-10` image's own shipped default, not to anything cloud-init touched.
Both became the next fix round, entry above.

## Task 3: service configuration — compose, systemd units, deploy (2026-08-30)

Landed `compose.yaml` (Navidrome, FileBrowser Quantum, cloudflared — no service
publishes a host port; only cloudflared is reachable from outside), `config/
navidrome.toml`, `config/filebrowser.yaml`, `config/msmtprc`, `config/beets-config.yaml`,
`config/cloudflared-tunnel-config.json`, the four `systemd/music-{import,backup}.
{service,timer}` units plus a shared `systemd/musicbox-ping-failure@.service` template,
`env/musicbox.env.template`, and `scripts/deploy.sh`. Extended `scripts/check.sh` with
`docker compose config -q` and a pattern-filtered `systemd-analyze verify`. All pinned
tags looked up and dated in `docs/STATUS.md`.

Two Task 2/Task 3 interface mismatches found and fixed (Task 3's own acceptance criteria
named the authoritative env var scheme, so Task 2's code was the thing that moved):
- **healthchecks.io slugs**: `music-import`/`music-backup` called `ping_hc
  musicbox-import` / `ping_hc musicbox-backup`, which resolves (per `lib.sh`'s own
  uppercase-and-dash-to-underscore rule) to `MUSICBOX_HC_MUSICBOX_IMPORT_URL` /
  `MUSICBOX_HC_MUSICBOX_BACKUP_URL` — doubled "MUSICBOX", not the
  `MUSICBOX_HC_IMPORT_URL` / `MUSICBOX_HC_BACKUP_URL` this task's env template requires.
  Fixed both scripts to call `ping_hc import` / `ping_hc backup`; updated
  `tests/lib.bats` and `tests/music-import.bats` to match (still 16 green). Monitoring
  would otherwise have silently no-op'd forever — `ping_hc` treats an unset URL var as a
  no-op, not an error, precisely so a monitoring bug never takes down the pipeline it
  monitors, which also means a naming bug like this one would never have surfaced as a
  failure anywhere.
- **Navidrome scan-trigger credentials**: `music-import`'s `trigger_navidrome_scan()`
  read `NAVIDROME_USER`/`NAVIDROME_PASSWORD`; this task's env template is required to
  name `NAVIDROME_ADMIN_USER`/`NAVIDROME_ADMIN_PASSWORD`. Renamed in the script and in
  `tests/music-import.bats`.

Two review findings fixed after the first pass (both would have broken the first real
deploy, not just this repo's own gate):
- **`/etc/musicbox/env`'s parent directory didn't exist.** Nothing in `provision.sh` or
  `cloud-init/user-data.yaml` creates `/etc/musicbox` (only `/srv/music` is built at
  provision time), so `install -m 600 ... /etc/musicbox/env` died with "cannot create
  regular file" on a fresh box, after the scripts and units had already landed. Fixed by
  adding `-D` to that `install` call, which creates the missing parent and is a no-op on
  every later re-deploy.
- **`config/msmtprc` was never actually installed to `/etc/msmtprc`.** The file's own
  header already claimed deploy.sh put it there, but deploy.sh only rsynced `config/` to
  `/opt/musicbox/config` for the record — `lib.sh`'s `notify()` calls plain `msmtp`, which
  resolves its account from `/etc/msmtprc` and would have found nothing there, aborting
  `music-import` under `set -euo pipefail` on the first skipped-track notification. Added
  an explicit install step (rsync to a temp path, then `install -m 0644` to
  `/etc/msmtprc`), with a matching diff-and-report line in `--check`.

Decisions a later pass would be wrong to rediscover from scratch:
- **Navidrome config split**: non-secret baseline settings (`MusicFolder`, `DataFolder`,
  `LogLevel`) live in `config/navidrome.toml`, bind-mounted read-only at
  `/data/navidrome.toml` alongside the read-write `/data` mount for the DB — Docker lets
  a single-file bind mount overlay a path inside an already-mounted directory. Scanner
  settings and the encryption key live in `compose.yaml`'s `environment:` instead (ND_
  env vars override the toml), specifically so the two option names the task called out
  by name (`ND_SCANNER_SCHEDULE`, `ND_SCANNER_WATCHERWAIT`) are literally grep-able in
  compose.yaml. `ND_CONFIGFILE=/data/navidrome.toml` is set explicitly rather than relied
  on implicitly, per navidrome.org's own Docker install guidance.
- **FileBrowser Quantum 1.5.x has no static per-user YAML config** — verified against the
  actual `v1.5.5-stable` tag's Go source (`backend/common/settings/structs.go`), not just
  the docs site (which mixes in v2.0 beta content). Users and their scopes live in its own
  SQLite DB, created via CLI/API, not this repo. `config/filebrowser.yaml` instead
  configures one source (`/srv/music/inbox`) with `createUserDir: true` +
  `defaultUserScope: "/"`, so creating an account named `alice` automatically scopes her
  to exactly `/srv/music/inbox/alice` — no per-user YAML needed, and the username IS the
  inbox directory name music-import already matches against `contributors.yaml`. Task 8
  creates the real accounts (one per `contributor`-role entry) once Task 0 lands real
  names.
- **cloudflared runs `tunnel --no-autoupdate run` with no local `--config`.** A
  token-driven tunnel is remotely managed; a local ingress file is silently ignored
  (verified against upstream `cloudflared` issues, not assumed). The actual ingress
  mapping is `config/cloudflared-tunnel-config.json`, a ready-to-PUT API body for Task 5
  to apply once the named tunnel exists — see STATUS.md's Open items, this is a real gap
  Task 5 must not skip.
- **msmtp's `passwordeval` reads `$MUSICBOX_SMTP_PASSWORD` from its own inherited process
  environment** (`printenv MUSICBOX_SMTP_PASSWORD`), not by re-reading
  `/etc/musicbox/env` from disk. That variable is already present in music-import's
  environment via its unit's `EnvironmentFile=`; having msmtp open the file a second time
  would require it group-readable by the unprivileged `music` user, widening every OTHER
  secret in that one file to the same process for no actual reduction in blast radius
  (that process already receives all of them via `EnvironmentFile=`). `config/msmtprc`
  itself holds no secret and installs world-readable at `/etc/msmtprc`.
- **Deploy target directory**: `deploy.sh` rsyncs `compose.yaml` + `config/` to
  `/opt/musicbox` on the box (not `/srv/music`, which is data, restored from R2, and
  never holds repo-sourced config). `docker compose` runs with `--project-directory
  /opt/musicbox` so the compose file's relative `./config/...` volume paths resolve
  regardless of the caller's cwd.
- **`deploy.sh --check` never prints secret contents**, even when diffing the rendered
  `/etc/musicbox/env` against what's already on the box: it reports only whether the
  content differs (`diff -q` via process substitution, nothing touches local disk).
  Exercised in this pass against an intentionally-unreachable host and against
  `127.0.0.1` (connection refused) — both paths report cleanly and exit 0 without a real
  box to deploy to.
- **beets on the box**: installed via `uv tool install --force beets==2.13.1` as the
  `music` user (`HOME=/srv/music`), then symlinked `/usr/local/bin/beet ->
  /srv/music/.local/bin/beet` — `uv tool install`'s default bin dir is not on systemd's
  default `PATH`, so without the symlink `music-import`'s `beet import` call would fail
  with "command not found" the first time a real timer fires. `config/beets-config.yaml`
  mirrors the workstation's `~/.dotfiles/beets` layout and plugins (fetchart, embedart),
  repointed at `/srv/music`, plus a quiet-mode section (`quiet_fallback: skip`,
  `none_rec_action: skip`) so `beet import -q` sends anything it can't confidently tag to
  review instead of guessing or blocking on input that will never arrive.

## Task 2: import, backup, and wrapper scripts (2026-08-30)

Landed `scripts/lib.sh` (`quiescent`, `safe_extract`, `free_space_ok`, `notify`,
`ping_hc`), TDD with `tests/lib.bats` written and confirmed red before any
implementation existed. Landed `music-import`, `music-backup`, `music-beet`, and
`scripts/check.sh` as the repo gate (shellcheck over every `scripts/*` bash file,
then the full bats suite). `tests/music-import.bats` runs the importer end to end
against a fixture `/srv/music` tree with `beet`, `curl`, and `msmtp` stubbed on
PATH: happy path, a beets skip filed to review with both notifications, a
non-quiescent inbox left untouched, and a free-space abort that fires the fail
ping before beet ever runs.

Two one-line review cleanups from Task 1 landed alongside: the volume glob guard
in `cloud-init/user-data.yaml` now also rejects a non-block-device match (a
zero-match glob previously fell through as a false positive on some shells), and
`provision.sh`'s firewall-rule check now fails with its own message when the
`hcloud firewall describe` call itself errors, instead of reporting a network
blip as a missing rule.

Decisions a later pass would be wrong to rediscover from scratch:
- **Contributor matching**: `music-import` matches an inbox directory name
  (`/srv/music/inbox/<user>`) against contributors.yaml's `name` field verbatim;
  there is no separate username/slug field. If Task 3/8's FileBrowser account
  provisioning wants a different inbox directory naming scheme, `contributors.yaml`
  needs a dedicated field and `uploader_email()` needs updating to match on it.
- **Navidrome scan trigger**: implemented against the native admin API (`POST
  /auth/login` for a JWT, then `POST /api/scanner/scan` with `x-nd-authorization:
  Bearer <token>`), not the Subsonic API, because Subsonic has no scan-trigger
  endpoint. This is the UI's internal API rather than a documented stable
  contract; verify it against the pinned Navidrome image tag once Task 3 chooses
  one, since a trigger failure is only best-effort (logged, not fatal) precisely
  because this endpoint is not guaranteed stable.
- **State DB paths default to** `$MUSICBOX_ROOT/state/library.db` (beets) and
  `$MUSICBOX_ROOT/state/navidrome/navidrome.db` (Navidrome), both overridable via
  env var. Task 3's `compose.yaml` must either mount Navidrome's data dir there or
  the env var overridden in `/etc/musicbox/env`.
- **Archive cap default**: `MUSICBOX_MAX_ARCHIVE_BYTES` defaults to 2 GiB in
  `music-import`; revisit if a real Bandcamp download exceeds it.
- New host dependency: `yq` (mikefarah/go, v4) for parsing `contributors.yaml`,
  needed on the box alongside `beet`/`rclone`/`msmtp`/`bsdtar`. Add it to Task 3's
  `deploy.sh` package list.
