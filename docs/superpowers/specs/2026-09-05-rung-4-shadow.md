# Rung 4: the shadow

Status: **DRAFT, rulings ratified.** Revision 2, 2026-09-05. All six open rulings below were accepted as recommended by Geoff, 2026-09-07: (1) the integration class in `internal/config`, option (a); (2) `ReleaseMBID` is the release under dispute, option (a); (3) both compat modes with the second a re-score, option (c); (4) the shipped 720h floor; (5) 60 requests per run; (6) the first deploy is a short owner sitting, option (a). Every "provisional, owner ruling N open" mark in the 4a plan reads as ruled.

Erratum, 2026-09-07 (Geoff): the build carries no calendar schedule; the shadow floor is an evidence minimum, never a launch date, and the date arithmetic below was removed.

The six open rulings in the next section go to Geoff. Under the
2026-09-05 unattended grant, the plan that follows this spec applies
each recommendation **provisionally** and marks every provisional
application in its own header, so a ruling that comes back differently
is a named edit rather than a rediscovery. One thing the grant does not
cover: **the box deployment is a box write**, and it waits for the
quiet-box check. Being the first production deploy of any greenfield
code, it is named here as an owner touchpoint rather than assumed
(ruling 6).

Written against the ratified build spec
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
revision 2.1, hereafter BS), the rulings
(`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`),
the dependency analysis
(`docs/superpowers/research/2026-09-03-rung-dependency-analysis.md`,
DA), the execution model
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`,
revision 2, PA), the rung 3 spec
(`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`, R3), the
reverse-edge completion
(`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`,
RE), and the rung 4 spec inventory
(`docs/superpowers/research/2026-09-05-rung-4-spec-inventory.md`,
INV), which is this rung's survey of the estate and the seam. A
reference to "the build spec's ruling 1" means the shadow-floor ruling
at `BS:804-815`; a bare "ruling N" is always one of this document's
own six.

Nothing here amends the build spec, the launch bar, or the ladder. Every
factual claim cites `path:line` against the tree at
`440131e3569dd5237432c60c6708d1cf29b59056`; the inventory's own
findings are cited rather than re-derived.

Revision 2, 2026-09-05: folds two fresh-context adversarial reviews,
merged at `docs/superpowers/research/2026-09-05-rung-4-spec-review.md`,
blockers B1-B6, majors M1-M11, minors m1-m15. All thirty-two findings
are folded; three take an adaptation rather than the literal fix, each
named inline where it lands, and one half of one finding is rejected
with its evidence in the review's own header. The load-bearing changes:
the observation mechanism now carries a grouping rule and an
album-basename join, because the box runs `group_albums: yes` and
`AlbumOutcome.SourceDir` names the uploader's directory rather than the
album (B1, B2); the two-mode scoring re-scores the recorded candidate
set instead of re-driving `tag.Album` through a replay transport, which
is what `conform`'s own `ModeDefault` does and why (B3); the two
artifacts are marshalled in `cmd/dubplate-server` rather than in
`internal/engine/shadow`, which may import no harness package (B4); the
greenfield template's literal rows are named and the two renamed wall
names are settled against the age store (B5); deploy edit 6 becomes a
real `-ldflags` change with the smoke test asserting the stamp (B6);
and the response-store question leaves the open list entirely, because
ratified ruling B already settles it.

**Conductor rulings on the fold (2026-09-05), the four items short of an owner ruling:** the shadow's wire pace is `2 * musicbrainz.DefaultInterval`, amending the 3d handoff's recorded `DefaultInterval`, because the box's own `beet import` shares the per-IP MusicBrainz budget and the 60-token cap is re-derived against the slower pace; the two contested-surface rows (the state directory layout; `corpus/analysis/adjudications.json` and the divergence schema) are written into the parallel-approach spec's table; rung 4's second scoring mode tracks `conform`'s `ModeDefault` re-score, which the conductor ruled at 3d's plan fold (revision 2.1, B2 option a) and whose in-code "provisional" mark pass 4a retires; `DUBPLATE_SITE_NAME` stays a template literal and the age store is not re-keyed. The six owner rulings stay open; pass 4a's plan applies each recommendation provisionally and marks it, and pass 4b's deploy waits for the owner sitting ruling 6 recommends.

## What rung 4 builds

The ladder entry, verbatim (`BS:572-580`):

> **Shadow** (M to build, continuous to launch) [E]. The shadow runner
> under `engine/shadow` and its subcommand; the divergence log (durable
> in the state directory; the store arrives later) and the adjudication
> artifact; and the deployment the band now prices: a hardened systemd
> unit per vps-conventions with pre-created `ReadWritePaths`, its
> environment file, and the deploy.sh path rung 1 repointed. The
> observation mechanism is designed in this rung's plan. From here the
> evidence clock runs unattended while the spine builds.

Six deliverables:

1. **The observation mechanism**, chosen below and designed in full:
   capture the album's item fields while it still sits in the inbox,
   reconcile it against the incumbent's own run record after the
   import, and compare.
2. **The shadow runner** under `internal/engine/shadow`, and its
   `shadow` subcommand on `cmd/dubplate-server` (a verb `main.go:11`
   already reserves for this rung).
3. **The observation store**: durable, in the state directory, one
   directory per observed album, carrying the divergence record, the
   case-shaped fixture, and the captured MusicBrainz responses, plus a
   per-pass record.
4. **The adjudication artifact**: pending rows in
   `analysis.AdjudicationEntry`'s exact shape, staged on the box and
   presented in checkpoint batches, never written into the corpus by
   the box. Both this and the case-shaped fixture are marshalled in
   `cmd/dubplate-server`, because `internal/engine/shadow` may import
   no harness package.
5. **The launch-gate report** and the mechanical re-evaluation that
   ruling B's adjudication-load reduction depends on.
6. **The deployment**: a hardened unit per vps-conventions with
   pre-created `ReadWritePaths`, a greenfield environment file, and the
   `deploy.sh` changes both need.

Rung 4's deploy is merge point M1 and the evidence phase's boundary
(`DA:335`, `PA:75`, `BS:910-912`). From that deploy the evidence
clock runs unattended while the spine builds.

## Open rulings for Geoff

Each question below is one the written record does not settle. Each
carries its options and a recommendation. Nothing the record already
settles appears here; the twenty-nine settled decisions are INV's
section (c.1) and are not reopened.

Revision 1 carried a seventh, on whether the shadow stores request and
response fixtures on the box. **It is retired, not answered** (fold,
M11). Ratified ruling B is "rung 4 deploys at functional completeness,
with every divergence stamped with the matcher commit so divergences
from superseded revisions are re-evaluated mechanically instead of
consuming owner adjudication" (`R3:93`, recommendation at `:95`,
carried into the handoff at `docs/HISTORY.md:280-286`). Mechanical
re-evaluation of a live observation is possible only offline, against
the responses that observation saw, so storing them is a consequence of
a ratified ruling rather than a preference between two legal designs.
Declining it would retire ruling B silently. The store therefore keeps
the captured item fields, the request set, and every response body per
observed album; the disk cost is priced in risk 5, and the trigger to
build a retention knob is named there rather than the knob built now.

### Ruling 1: where the shadow's configuration lives

`internal/config.Load` errors on every set `DUBPLATE_*` name that is
not one of the four owned rows (`internal/config/config.go:133-152`,
rows at `internal/config/keys.go:19-25`). The wall has exactly four
states today: owned-now, reserved, superseded, unknown. The shadow
needs a fifth, because it is an integration rather than an owned
capability: U4 requires an explicit enable flag whose correlated config
becomes required-when-enabled (`BS:150-158`), and the shadow is
explicitly not in the owned set (`BS:141-149`).

**Options.**

- **(a) An integration class in `internal/config`.** An
  `IntegrationClass{Name, Rung, Flag, Keys}` table beside
  `reservedClasses` (`keys.go:57-142`). `rejectNamespace` accepts a
  class's key when the class's flag is set and errors naming the flag
  when it is not; `Load` then requires every key of every enabled
  class. `doctor` reports each class with its flag state.
- **(b) An accepted-optional key set.** A flat list of names
  `rejectNamespace` tolerates and `Load` never requires, read by the
  subcommand through its own lookup. Roughly twenty lines, no
  required-when-enabled semantics.
- **(c) Subcommand flags outside the wall entirely.** The shadow's
  paths and knobs arrive as cobra flags on `dubplate-server shadow`,
  and `internal/config` is untouched.

**Recommendation: (a).** U4 mandates the shape, and rung 5 needs the
identical mechanism for the identity source, which already sits in the
wall as a pending row (`keys.go:24`). Building it once here spares rung
5 the same design and gives rung 1b's registry a class to absorb rather
than a shape to invent. The cost of (b) is the failure mode that
matters most in this whole rung: a half-configured shadow that runs
unattended for thirty days observing nothing, which is exactly what
required-when-enabled prevents. The cost of (c) is that the estate's
inbox and review paths would move into a committed unit file's
`ExecStart` line, where an adopter has to edit them, instead of into a
rendered environment file where U1 wants them.

### Ruling 2: how a skipped album's `Subject.ReleaseMBID` is filled

`checkAdjudicationEntry` accepts the shadow subject shape only when
`Album` and `ReleaseMBID` are both set, and errors when one is set
without the other
(`internal/harness/analysis/adjudication.go:168-171`). Under
`quiet_fallback: skip` the oracle writes nothing for a skipped album
(`config/beets-config.yaml:16-17`, verified in pinned beets at
`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:335-341`),
so the oracle supplies no MBID for exactly the half that most needs
adjudication.

**Options.** (a) `ReleaseMBID` is the release under dispute, filled by
whichever side proposed one: the oracle's when the oracle applied, the
Go matcher's `Candidates[0].Info.AlbumID` when the oracle skipped and
Go applied. (b) Relax the pair rule for shadow subjects, so `Album`
alone is legal. (c) A third subject shape.

**Recommendation: (a), with the field's doc comment restated to say
so.** Every divergence has exactly one release under dispute, and the
row's own body records which side proposed it, so nothing is lost and
the pair rule stays a check that can fire. (b) turns a firing check
into a permissive one for the shadow rows specifically, which is the
assertion-discipline failure this repo has already paid for twice.
`adjudication.go:32-34` already reads "an album plus release MBID for a
live shadow divergence rung 4 will write", so (a) is the reading the
field was written for.

### Ruling 3: which `matcher.Compat` the shadow runs

`FullCompat()` reproduces every ledgered defect and is what `conform`'s
`ModeCompat` runs; `DefaultCompat()` turns both switches off, following
the ledger's own oracle-wrong dispositions
(`internal/engine/matcher/compat.go:31-46`).

**Options.** (a) `FullCompat()` alone: the shadow matches the oracle,
so a compat-class divergence never appears and the shadow measures the
port rather than the product. (b) `DefaultCompat()` alone: the shadow
measures the product, and every ledgered defect the oracle still
carries arrives as a divergence the owner has already ruled on. (c)
Both: run the flow once under `FullCompat`, then **re-score that run's
own candidate set** under `DefaultCompat` and record both verdicts.

**Recommendation: (c), with the second mode's mechanics corrected
(fold, B3).** Revision 1 described the second mode as a replay of the
stored responses through a second `tag.Album` call, and cited
`conform`'s two-mode sweep as the precedent. That citation says the
opposite, and it is the reason the corrected mechanics matter:
"ModeDefault does not drive tag.Album. It re-scores each case's own
committed candidate set under the corrected table, and the reason is
that a corrected verdict changes the FLOW"
(`internal/harness/conform/case.go:141-153`), and `gradeCaseDefault`
replays only the lookups the committed candidates name, issuing no
search (`case.go:517-526`, `:591-593`). The mechanism is the
ID-consensus early return, which fires only at `rec ==
matcher.RecStrong` (`tag.go:96-118`, condition at `:106`);
`matcher.Compat.Apply` swaps `StringDistPatterns`
(`internal/engine/matcher/compat.go:51-60`), which moves the distance
and therefore the recommendation, so an album that takes the one-lookup
early return under `FullCompat` falls through to `src.SearchReleases`
(`tag.go:125-126`) under `DefaultCompat` -- a request the recording
never captured.

So the second mode re-scores: translate the recorded responses, re-run
the assignment solve, re-score the album distance, re-sort, and
recommend, all under `DefaultCompat().Apply(tun)`, issuing no search
and driving no second `tag.Album`. **The derived figure's limit is
stated rather than absorbed**: a re-scored `DefaultCompat` verdict is
not the product's true `DefaultCompat` verdict wherever the real flow
would have widened the candidate set, so the compat-class annotation is
a hypothesis and the divergence row says so. `conform`'s own argument
for why re-scoring is exact (`case.go:149-152`, medium decided by
minimum distance alone) covers ten pinned corpus cases and does not
generalize to arbitrary live albums.

The budget is the second correction. The budget sits above the
transport, not below it: `doWithRetry` calls `c.budget.Wait(ctx)`
before every attempt (`client.go:346`) whatever `ClientConfig.HTTP`
holds (`client.go:98`), so a replay through the metered client would
double the `ClassBackground` spend and double `Budget.Spent()`
(`limiter.go:400-404`), the figure the report prints as honest. Any
replay client the re-scoring or `reevaluate` builds takes
`musicbrainz.Unmetered()` (`limiter.go:354`), exactly as `gradeClient`
does (`case.go:330-336`).

What (c) buys is unchanged: a divergence where `DefaultCompat`
disagrees with the oracle while `FullCompat` agrees is a known
compat-class divergence, annotated with its hypothesis and never
presented as news; a divergence where both modes disagree is news and
gets a pending row.

**One conductor ruling rides this.** `conform`'s `GradeMode` doc
carries "Revision 2, fold B2; provisional, conductor ruling owed"
(`case.go:152-153`), so the shape rung 4 now mirrors is itself
provisional. If that ruling moves `ModeDefault`, rung 4's second mode
moves with it.

### Ruling 4: the estate's own shadow floor

The build spec's ruling 1 makes the floor a configurable duration with
a 30-day default (`BS:804-815`). The estate's own value is a separate choice and is the
figure the launch gate's floor check reads (`DA:511-520`).

**Options.** Any duration. The shipped default is `720h`.

**Recommendation: take the shipped 30-day default.** The clock starts
at 4b's deploy (`DA:315`, `DA:525`), and the spine (rungs 6, 14, 7, 9)
plus the surfaces phase (10 through 13) will not complete inside thirty
days, so any plausible floor is satisfied before rung 15 evaluates it
(`PA:319`, `PA:327-333`). Taking the default costs nothing and keeps
the estate's setting equal to the adopter-facing one, which is the
honest position for a value the build spec's ruling 1 justified
through the universal-adopter lens.

### Ruling 5: the shadow's MusicBrainz budget cap

The class and the pace are fixed: `ClassBackground`
(`internal/engine/musicbrainz/limiter.go:25`) at `DefaultInterval =
time.Second` (`limiter.go:62`). The cap's value is nowhere on the
record (INV O12).

**Options.** Any integer, spent per run through
`Limiter.Budget(name, class, limit)` (`limiter.go:139`).

**Recommendation: 60 requests per run, registered as a threshold with
that default.** An album costs one lookup on the ID-consensus early
return (`tag.go:96-118`) and, on the search path (`tag.go:121-134`),
one search plus `SearchLimit` lookups, with `SearchLimit` shipping at 5
(`internal/engine/matcher/tunables.go:309`). **The real bound is
eleven, not seven** (fold, m7): `MergePseudoRelease`'s second fetch is
per candidate (`client.go:320-333`), so a search returning five ids can
cost 1 + 5 + 5, and `maxAttempts = 7` (`client.go:42`) multiplies a
degraded run because a retried request spends a second token by design
(`client.go:343-347`). Sixty tokens therefore covers five albums in the
worst case and eight or nine in the ordinary one, which outpaces any
plausible arrival rate on owner-curated material (`BS:243-249`), and at
two seconds a request the run's wire time stays inside two minutes,
comfortably under the five-minute timer chosen below (corrected
2026-09-07 at the 4a-ii close: the pre-fold sentence read "at one
request per second ... inside sixty seconds", which the M6 fold and the
conductor's pace ruling supersede). `Budget.Spent()`
is the honest count the report prints (`limiter.go:400`), persisted per
pass so `shadow report` can sum it.

**Exhaustion is a branch, not a verdict** (fold, M1). At this cap a run
with a backlog reaches `musicbrainz.ErrBudgetExhausted` mid-pass as the
ordinary path, including the first pass after deploy, which sees
whatever the inbox already holds. That sentinel exists so a caller
branches, "to stop asking rather than to retry"
(`limiter.go:45-48`); the divergence classification below defers on it
rather than recording a Go skip.

**The wire budget is shared with the box's own beets, and the limiter
cannot see it** (fold, M6). `Limiter` paces "every MusicBrainz request
one process issues" (`limiter.go:64-65`). The box's other MusicBrainz
client is `beet import`, which pins its own `rate_limit = 1.0`
independently (recorded at `limiter.go:55-61`), fires `*:0/15`
(`systemd/dubplate-import.timer:10`), and shares the box's single
egress IP, where MusicBrainz's own limit actually applies. A `*:0/5`
shadow will regularly run alongside an import, and two processes each
pacing at one per second exceed the shared quota, which would throttle
the production path this rung promises to leave alone. Two mitigations
are available: skip a pass while an import is plausibly in flight,
readable from the newest `report-*.json`'s `ended_at` without touching
the lock, or build the shadow's limiter at `2 *
musicbrainz.DefaultInterval` and re-derive the cap. **The second is the
recommendation**, because it needs no inference about another process's
schedule and the sixty-token cap survives it: five albums at eleven
requests and two seconds a request is under four minutes, still inside
the five-minute timer. `reservationInterval` and the class policy are
untouched, since this is the interval the shadow hands `NewLimiter`
(`limiter.go:110`) rather than the limiter's own rule.

**This is a conductor ruling, not an owner one, and the plan does not
assume it.** The 3d handoff records the shadow's wire pace as
`DefaultInterval` (`docs/HISTORY.md:287-293`, 3c contract C20), so
halving it departs from a written handoff item. The conductor rules it
before 4a's plan; if the ruling keeps `DefaultInterval`, the first
mitigation lands instead and the pre-task's import-timer question
becomes load-bearing rather than informational.

### Ruling 6: whether the first deploy is a conductor action or an owner sitting

Rung 4 is one of three rungs that touch the box (`BS:864-874`), the
write is additive and read-only with respect to the library
(`BS:262-263`), and it goes through `scripts/deploy.sh` with `--check`
first. Whether Geoff sanctions it explicitly is written nowhere. The
nearest precedent is `scripts/deploy.sh:49-55`, where the uploader's
serve unit is deliberately excluded from the timer-enable list and
started "by hand, under Geoff's sanction, never by this script".

**Options.** (a) An owner sitting: `--check` output read together, the
timer enabled by hand, the SELinux verification closed, the deploy date
recorded. (b) A conductor action under the standing grant, reported
afterwards.

**Recommendation: (a), and it is short.** This is the first production
deploy of any greenfield code, it is the act that starts the launch
clock, and the build spec's owner-touchpoint accounting
(`BS:894-903`) counted the shadow's adjudication stream without
counting its deploy. One sitting is cheaper than discovering the
omission at rung 15. **It does add a second attended gate to a record
that priced one** (fold, m6): `BS:894-897` prices "one attended gate
(rung 16's single-sitting verification)", one attended visual check,
and the standing adjudication stream, and a deploy sitting is not one
of them. The argument stands on the first-production-deploy reason
alone, which the preceding sentences already make; the accounting is
stated honestly rather than argued away.

## The observation mechanism

This is the spec's central design decision. The record deliberately
left it here (`BS:266-268`, restated at `BS:576-578`).

### The hazard that decides it

The box's beets runs `move: yes` and `write: yes`
(`config/beets-config.yaml:14-15`). An applied album's files are
renamed into the library and their tags rewritten with beets' own
answer. After the import, the input the matcher would need is gone, and
reading the library files back hands `tag.Album` a pre-tagged input and
a trivially confirming verdict, which measures nothing. **Any
observation mechanism that starts after the import recovers the
oracle's answer and not the question** (INV a.3). Only a capture before
the import preserves both.

### The candidates

INV a.4 enumerates eight. Summarized, with the reason each is taken or
declined:

| # | Mechanism | Taken |
| --- | --- | --- |
| A | Poll the inbox before the importer stages it | **Yes**, as the capture step. It is the only one that sees the question |
| B | Poll the staging directory between staging and `beet import` | No. The window is inside one `runPipeline` call (`_salvage/internal/importer/importer.go:273` to `:1100`), typically seconds |
| C | Read `/srv/music/state/library.db` | **No**, and this is a departure from INV's own reading. See below |
| D | Read `/var/lib/musicbox/import-*.log` | **Yes**, as a durability backstop for the skip half. Never swept, written atomically (`_salvage/internal/beets/beets.go:104`) |
| E | Read `/var/lib/musicbox/report-*.json` | **Yes**, as the reconcile step. The single artifact that names the applied/skipped split per album (`importer.go:126-172`) |
| F | A `.path` unit watching the state directory | No. Noisy, coalescing, lossy in the re-arm gap, and it requires a timer backstop anyway (`~/.claude/skills/vps-conventions/SKILL.md:47-55`) |
| G | `ExecStartPost=` on `dubplate-import.service` | No. It runs inside the incumbent's sandbox with `ReadWritePaths=/srv/music /var/lib/musicbox` (`systemd/dubplate-import.service:54`), and it edits a frozen unit, breaking the additive-deploy boundary rung 1 established (`docs/HISTORY.md:2097-2100`) |
| H | Poll `/srv/music/review/<date>/<user>/` | **Yes**, as the skip half's second corroboration. Files are at rest once filed |

### The mechanism: capture, reconcile, match

**Capture (A).** On each pass, walk `<observe dir>/<uploader>/<album>/`
and record, per album directory, a fixture: the `matcher.ItemFields`
read from each track, the observation's evaluation date, and a digest
over the file set (names, sizes, modification times). The digest is the
identity: a directory whose digest has not moved is not re-captured, so
the capture is idempotent and a mid-write capture is superseded by the
next one.

**A capture that cannot read every entry it listed is discarded** (fold,
M4). The supersede rule above holds at the upload boundary and not at
the other one: once the importer moves an album out of the inbox there
is no next pass to supersede anything. `review.Move` tries `os.Rename`
first (`_salvage/internal/review/move.go:24-40`), atomic per top-level
entry on one filesystem, with a recursive copy-then-remove fallback
across an EXDEV boundary; either way a walk that lists a directory and
then opens its files one at a time can read three tracks and take
ENOENT on the fourth. A partial read is never stored, never scored, and
never counted as an observation: the pass logs it and moves on, and the
album is captured whole on a later pass or not at all.

**The grouping rule, because the box's album unit is not the
directory** (fold, B1). The box's beets sets `group_albums: yes`
(`config/beets-config.yaml:23`), and its own comment records why, from
this estate's real material: contributor ZIPs extract flat, "so one
staging dir holds many albums. Without this, beets saw the first seed's
nine albums as one 135-track candidate and skipped them all
(2026-08-31)" (`config/beets-config.yaml:19-22`). The incumbent runs
one `beet import` over the whole staging tree
(`_salvage/internal/importer/importer.go:1100`), so beets groups by
album tags across it and is free to split one captured directory into
several verdicts. The corpus deliberately sits below this step
(`corpus/README.md`, `import.group_albums` dropped because the corpus
tests the matcher from the items-and-responses boundary, below the
importer's grouping step), so rung 3 inherited no grouping and
`tag.Album` has none: it takes `items []matcher.ItemFields` and calls
`matcher.MostCommonTags(items)` over all of them (`tag.go:68`, `:75`).

**Rung 4 observes only what groups to exactly one album, and counts the
rest.** After reading a directory's items, group them by their
`albumartist`-or-`artist` and `album` fields, beets' own key (`cmp.Or(AlbumArtist, Artist)` and `Album`, `beets/importer/stages.py:102-103`; conductor ruling C1 at 4a's plan fold, 2026-09-05). A directory yielding one group is
observed. A directory yielding more than one is **excluded**, with its
group count and its basename recorded, and the report prints the
excluded-multi-album count beside the gate-rejected one, on the same
"report what was actually compared" discipline. Porting beets' own
`group_albums` pre-step is the alternative and is declined here for the
same reason mechanism B is: it would put a second, unledgered port of
importer behavior inside a rung whose whole product is the honesty of a
comparison. If the record proves thin at a checkpoint and the excluded
count is a material share of it, porting the pre-step is the named
follow-on, and the count is what makes that decision evidence-based
rather than speculative.

**The quiescence rule, stated as a constraint rather than a copy.** The
incumbent imports an inbox entry only once every file's modification
time is at or before `now - QuiescentMin`, ten minutes by default
(`_salvage/internal/importer/importer.go:790-798`,
`_salvage/internal/config/config.go:152`), and its timer fires at
`*:0/15` with a 300-second jitter
(`systemd/dubplate-import.timer:9-11`). So **the shadow's own poll
interval must be strictly less than that quiescence window**, and the
timer below is set at `*:0/5` for exactly this reason. Under that rule
the last capture taken before an album leaves the inbox is always a
quiescent one, which is the fixture the reconcile step needs. The
shadow reads no incumbent config to know this; the interval is the
shadow's own configuration and the constraint is recorded in the unit's
own comment.

**Reconcile (E, backstopped by D and H).** Read the incumbent's run
reports and resolve each `AlbumOutcome` to a captured fixture, then
record the oracle signal.

**The join key is the album basename, never `SourceDir`** (fold, B2).
Revision 1 joined on `AlbumOutcome.SourceDir`, and that field cannot
identify an album. It is documented as "the uploader's inbox directory"
(`_salvage/internal/importer/importer.go:212`) and all three
assignments set it at the uploader level (`:766`, `:908`, `:1066`);
`AlbumOutcome` carries no per-album path at all (`:134-150`), because
the album name lives only in `albumUnit.stagedDir` (`:213`), which
`toReport` does not copy (`:231-240`). `stageAllContributors` stages
every quiescent album of every contributor in one run (`:782-800`), so
a run report routinely carries several rows with identical `SourceDir`,
and a `SourceDir` join maps every fixture to every outcome.
`SessionID` does not rescue it either: it is empty for a leftover
staged before the uploader shipped or a synthetic orphan (`:139-142`),
and the wind-down retired every intake path, so this material carries
no sidecar. So `SourceDir` narrows to the uploader and nothing more,
and the album basename resolves the row:

- `imported` is an **apply**. The oracle's chosen release MBID is read
  from `mb_albumid` on the first file `library_paths` names, which the
  shadow's own FLAC reader already does. Beets wrote it there
  (`config/beets-config.yaml:15`), and `MUSICBRAINZ_ALBUMID` is
  mediafile's Vorbis key
  (`tools/venv/.venv/lib/python3.13/site-packages/mediafile/__init__.py:782`).
  The applied files are matched back to the fixture by their track
  titles and lengths as well as by `mb_albumid`, because
  `library_paths` is itself a positional correlation: the importer
  fills it by matching staging order against the beets database's album
  order, since "one shared beet invocation carries no per-uploader join
  key", and a count mismatch leaves the extra units empty
  (`importer.go:1119-1127`).
- `review` is a **skip**, resolved by the album's basename in a tag
  log's `skip ` lines (`_salvage/internal/beets/beets.go:116-135`,
  never swept) or under `<review dir>/<date>/<uploader>/<basename>`.
  This is the half with no oracle MBID and the half ruling 2 fills, so
  its identity comes from the two backstops rather than from the run
  report.
- `quarantined` and `rejected` are **not oracle verdicts at all**.
  Gate-rejected material never reaches beets (`importer.go:945`,
  `:499-505`), so the observation is discarded rather than counted. The
  report prints the excluded count with its reason; an observed-album
  count that included them would lie.

An observation the reconcile step cannot resolve is recorded as
`unreconciled` with the reason and is counted separately. That is the
`CaseGradeSummary` discipline applied here: report what was actually
compared, never what was visited
(`internal/harness/conform/case.go:109-110`). Three named
`unreconciled` reasons, rather than one bucket: an ambiguous basename;
an `imported` outcome whose `library_paths` is empty, which the
importer's own `Error` field describes as "a beets-database match that
never turned up" on an outcome that still reads `imported`
(`importer.go:127-133`, `:145-149`); and a fixture whose run report was
already swept.

**The run reports are a rolling seven-and-a-half-hour window, and the
shadow is written against that fact** (fold, M3). `sweepReports` runs
on every import pass (`importer.go:389`) and "keeps at most retention
of the newest `report-*.json`" (`:533-548`), with
`defaultReportRetention = 30` (`_salvage/internal/config/config.go:172`).
A `RunReport` is written on every exit path including a held lock
(`importer.go:152-156`) and the timer fires `*:0/15`, so thirty reports
is roughly seven and a half hours of history, not a durable ledger.
Backstop H is swept too, at `RetentionDays`, defaulting to 60
(`importer.go:386`, `config.go:155`); only D, the tag logs, is never
swept, and it records skips only. Two consequences the design takes
rather than absorbs. The shadow records which report files it has
already consumed, so a pass never re-reads one and a **gap** in the
report sequence is a named, reported condition rather than silence. And
the timer below sets `Persistent=true`, because a shadow outage longer
than that window costs the oracle signal permanently for every album
imported inside it.

**Why not the beets database (C).** INV a.4 and b.2 both treat reading
`/srv/music/state/library.db` as the natural source of the oracle's
applied release MBID, and INV's O5 then spends three sections on the
config-wall problem it creates, because `DUBPLATE_BEETS_DB` is retired
outright with "retired: the greenfield server never embeds beets"
(`internal/config/keys.go:179`). Reading `mb_albumid` off the applied
library file dissolves the whole question. It needs the FLAC reader the
capture step already builds, it adds no sqlite driver to a module whose
only requirement today is cobra, and it leaves the retirement message
standing unamended. **O5 is closed by not asking it.**

**Match.** For each reconciled observation, drive `tag.Album` against
the captured `ItemFields`, under the estate tunables and
`FullCompat()`, spending `ClassBackground` budget, through a recording
transport. Then re-score that run's own candidate set under
`DefaultCompat()` (ruling 3), which drives no second `tag.Album` and
spends no second token. Classify, record, and stage a pending row where
the classification is news.

### The read-only guarantee

The unit runs `ProtectSystem=strict` with `ReadWritePaths=` naming the
shadow's own state directory and nothing else. The kernel mounts every
other path read-only inside the unit's namespace, so the library, the
inbox, the review tree, the beets database, and the incumbent's own
state directory are unwritable to the shadow by construction, not by
convention. That is the strongest form of the "additive and read-only
with respect to the library" constraint (`BS:262-263`) available on
this box.

**The guarantee is the timer path's, and only the timer path's** (fold,
M5). `ProtectSystem` and `ReadWritePaths` bind `ExecStart` and nothing
else. `shadow report` and `shadow reevaluate` have no timer, risk 3
contemplates a hand-run `observe`, and the deploy sitting is a hand-run
session over ssh where `deploy.sh` connects as `root@$HOST`
(`scripts/deploy.sh:110`) -- a verb run that way carries no sandbox at
all and has full write access to the library and to the incumbent's
state. So the guarantee comes with an operator rule, stated here, in
the quiet-box procedure, and in the unit's own comment: **a hand-run
shadow verb goes through `systemctl start dubplate-shadow.service` or
`systemd-run` with the same properties, never the bare binary as
root.**

**One trap the code must close on its own,** because the sandbox does
not: the shadow must never open `<incumbent state dir>/import.lock`.
The incumbent treats a held lock as healthy and returns nil without
importing (`importer.go:397-410`), so a shadow holding it would stall
the box's own import while reporting green. `flock` on a read-only
descriptor succeeds on Linux, so this is a code rule with its reason in
the code, not a grant the unit can withhold.

### The race with beets

The capture reads files the incumbent is about to `rename(2)` out from
under it. A rename out from under an open descriptor is safe on Linux,
so an in-flight read completes against the original inode. The
correctness question is not the read but the ordering, and the
quiescence rule above answers it: the poll interval is strictly shorter
than the incumbent's quiescence window, so a quiescent capture always
precedes the import. A capture that lands mid-write produces a fixture
with a digest the next pass supersedes.

### The oracle signal it recovers

The oracle is binary. In quiet mode `_summary_judgment` returns `APPLY`
when and only when the recommendation is `strong`, and otherwise
resolves `quiet_fallback` to `SKIP`
(`.../beets/ui/commands/import_/session.py:335-341`), mirrored exactly
by `matcher.ActionFor`
(`internal/engine/matcher/verdict.go:179-192`). So the comparable
quantity is a binary projection of `Proposal.Action`, `ActionApply`
against `ActionSkip`, and nothing finer. `RecStrong` is exactly
attested; `RecMedium`, `RecLow`, and `RecNone` are indistinguishable to
the oracle. Distances, per-key penalties, track vectors, and
assignments are logged as evidence and never graded, which is
`BS:257-260` read against the landed types.

### The box-write footprint

One directory: `<state dir>` and everything beneath it. Nothing else on
the box is written by the shadow at any point. The deploy itself writes
the unit files, the environment file, and the binary, which is
`deploy.sh`'s existing footprint plus two units and one file.

## The divergence log and the adjudication artifact

### The divergence classes

Four outcomes, three of them divergences:

| Oracle | Go | Class |
| --- | --- | --- |
| apply, release R | apply, release R | agreement, no row |
| skip | skip (or `tag.ErrNoCandidates`) | agreement, no row |
| skip | apply, release G | **`go-applies-oracle-skipped`**. The class a false confident match hides in, and the launch bar's own clause (`BS:69-71`) |
| apply, release R | skip (or `tag.ErrNoCandidates`) | **`go-skips-oracle-applied`**. The tractable class: the oracle names a release and the fixture is honest |
| apply, release R | apply, release G, `R != G` | **`go-applies-different-release`** |

The third row is a class INV does not name and the record does not
anticipate. It is where "zero false confident matches against live
beets decisions" (`BS:69-71`) is most directly measured, so it is named
here rather than discovered as an unclassifiable observation.

Four branches are neither, and the enumeration is closed deliberately
(fold, M1). `musicbrainz.ErrUnavailable` defers the observation for a
later pass rather than recording a divergence (`client.go:74-84`).
`musicbrainz.ErrBudgetExhausted` (`limiter.go:45-48`) and
`musicbrainz.ErrLimiterClosed` (`limiter.go:50-53`) defer it the same
way: both are sentinels precisely so a caller branches, "to stop asking
rather than to retry", and at ruling 5's cap a backlogged run reaches
exhaustion as the ordinary path. `tag.ErrNoCandidates` (declared
`tag.go:18`, returned `:139`) records "no match", which agrees with an
oracle skip and diverges from an oracle apply.

**No error other than `tag.ErrNoCandidates` may ever be recorded as a
Go skip.** A truncated album written as `go-skips-oracle-applied` is a
fabricated divergence in the one artifact whose whole value is that it
is honest, and it consumes exactly the owner adjudication ruling B's
stamp exists to save. The report's `deferred` row counts by reason, so
a chronically undersized cap is visible rather than silent.

### The two rows

Revision 1 described one shape "lifted verbatim" into the pending
adjudication artifact. That is not available: `analysis.AdjudicationEntry`
has no `oracle`, `go`, or `class` field (`adjudication.go:46-67`) and
`AdjudicationSubject` has exactly three (`:32-39`), so an implementer
following it either drops the mechanical facts a rung 15 re-evaluation
needs or adds three unspecified fields to the shared corpus type. **The
shape is two rows, one of which is a projection of the other** (fold,
M10).

**The observation record**, `observation.json`, rung 4's own type,
carrying every mechanical fact:

| Field | Source |
| --- | --- |
| `id` | `"shadow-" + <observation id>`, deterministic from the capture |
| `album` | the captured album's `album` field, or the source directory's basename when it is empty |
| `matcher_commit` | the stamp ruling B requires (`R3:93-95`) |
| `matcher_config_hash` | a hash over the `Tunables` and `Compat` pair the run used |
| `eval_date` | the capture date, so re-evaluation reproduces it exactly |
| `oracle` | `signal` (`apply` / `skip`), `release_mbid` where one exists, `uploader`, `album_basename`, and the artifact the signal was read from |
| `go` | per mode: `recommendation`, `action`, the best candidate's album id and distance, and `Dist.NonzeroKeys()` |
| `class` | one of the three divergence classes above, or `agreement` |
| `budget` | the pass's `Budget.Name()` and `Budget.Spent()` |

**The staged pending row**, `pending/<obs-id>.json`, an
`analysis.AdjudicationEntry` and nothing more, written only when the
class is news:

| Field | Source |
| --- | --- |
| `id` | the observation id |
| `subject.album` | the observation's `album` |
| `subject.release_mbid` | the release under dispute (ruling 2) |
| `matcher_commit`, `matcher_config_hash` | the two schema additions below |
| `disposition` | `pending-adjudication` (`analysis.Pending`, `adjudication.go:18`) |
| `adjudicator` | `analysis.AgentAdjudicator` (`adjudication.go:30`) |
| `suspected`, `release_truth`, `reason`, `date` | four of the five fields `checkAdjudicationEntry` requires non-empty (`adjudication.go:176-189`), the fifth being `adjudicator` above; the runner renders them from the observation rather than emitting `TODO` |
| `citation` | the observation id and its class. `citation` is a convention rather than a check: `checkAdjudicationEntry` never tests it, though all 21 committed rows carry one |

The rendering is one-way and lossless in the direction that matters: a
pending row names the observation whose record carries the class, the
two verdicts, and the responses, so a rung 15 re-evaluation reads the
mechanical facts off `observation.json` and the owner's ruling off the
committed corpus row. Nothing mechanical is duplicated into the shared
corpus type, which is what keeps the two decoders from drifting
(`internal/engine/matcher/info.go:148-150`).

**`matcher_config_hash` is deliberately not named `config_hash`.**
`cassette.Expected.ConfigHash` is a hash of the Python shim's effective
config (`internal/harness/cassette/cassette.go:80`). The shadow's is a
hash over the Go matcher's `Tunables` and `Compat`. Conflating them
would make a rung 15 comparison silently wrong, so the names differ at
the JSON boundary (INV b.9).

**`disposition` stays owner-only.** A non-`Pending` disposition set by
`AgentAdjudicator` is a gate error (`adjudication.go:198-200`), so
every row the shadow writes is `pending-adjudication` with
`adjudicator: agent`, and the owner's ruling is what moves it.

### Where it lives

Durable in the state directory, per `BS:573-575`; the store arrives at
rung 6 and supersedes none of this. One directory per observation, so
two writers never contend for one file:

```
<state dir>/shadow/
    started-at                      the clock start, written once, never rewritten
    passes/<pass-id>.json           one record per observe pass: start, end, counts, budget name and spend
    observations/<obs-id>/
        observation.json            the divergence record above, or the agreement record
        case.json                   the cassette-shaped fixture, Expected nil
        responses/NNN.json          the captured response bodies
    pending/<obs-id>.json           the AdjudicationEntry a checkpoint batch presents
    resolved/<obs-id>.json          a divergence reevaluate retired, and the commit that retired it
    consumed-reports.json           the incumbent run reports already reconciled, so a gap is detectable
```

**`passes/` exists because a `Budget` is in-process state** (fold, M8).
`spent` is an int guarded by a mutex on the struct
(`limiter.go:393-404`), created per call to `Limiter.Budget`
(`limiter.go:139-145`), and nothing persists it. `observe` and `report`
are separate process invocations, so a `report` that built its own
limiter would print a zero that reads as a measurement of an unattended
thirty-day window. Each pass writes its own `Budget.Name()` and
`Budget.Spent()`, and `report` sums them, printing per-run and
cumulative separately.

Every file is written through `atomicfile.WriteFile`, so a reader never
sees a partial write. That package does not exist in the greenfield
tree (`internal/` holds `config`, `engine`, `harness` only); it is
salvage tier at `_salvage/internal/atomicfile/atomicfile.go:14`, 42
lines, and rung 4 re-lands it as `internal/atomicfile` on the merit
admission rule rather than importing it. Rungs 6, 9, and 13 all want
it, so it is born once here.

**No lock package.** `internal/lock` stays unadopted at rung 4. The
timer path is a systemd singleton, `reevaluate` writes only under
`resolved/`, `report` opens nothing for writing, and the store's
one-file-per-observation layout means `observe` only ever creates. The
residual is a hand-run `observe` racing the timer, which is named in
risk 3 and costs at worst a re-captured fixture. `internal/lock` is
whichever later rung first has two real writers.

### Which package writes these artifacts

**`internal/engine/shadow` may import no harness package, so it
marshals neither artifact** (fold, B4). `TestEngineNeverImportsHarness`
walks the whole `internal/engine` tree and fails any file importing
anything under `internal/harness`
(`internal/harness/conform/arch_test.go:141-146`,
`checkNoHarnessImports` at `:55-65`), and the walk descends into nested
packages. So the package cannot reach `cassette.Case`
(`internal/harness/cassette/cassette.go:45-56`) or
`analysis.AdjudicationEntry`, which is precisely the shape deliverables
3 and 4 name. Revision 1 invoked this same test to justify landing
`matcher.QuietImportTunables` on the engine side and then did not apply
it to the two artifacts the rung produces; an implementer following it
either reds the gate or hand-rolls duplicate structs, which reintroduces
the silent-drift failure this repo has already named
(`internal/engine/matcher/info.go:148-150`) at the seam where drift
would make a rung 15 comparison wrong.

The seam is the one the spec already uses for `now` and the User-Agent.
`internal/engine/shadow` returns typed engine-side results:
observations, verdict pairs, classes, and the recorded responses.
`cmd/dubplate-server` imports both trees freely and does the
marshalling, writing `case.json` in `cassette.Case`'s shape and
`pending/<obs-id>.json` in `analysis.AdjudicationEntry`'s.
`observation.json` is the engine's own type and needs no harness import
at all.

**Every staged row is validated before it is written.**
`checkAdjudicationEntry` is unexported (`adjudication.go:158`), but
`analysis.CheckAdjudications` is exported (`adjudication.go:138`) and
takes the known case ids as its first argument, which a shadow row does
not need. The subcommand runs it over each row before staging it, so a
malformed row fails on the box within one pass rather than weeks later
when the owner tries to commit it. That is the same
prove-the-instrument-fires discipline the process section asks of the
checkers, applied to the artifact.

### How it shares shape with the corpus's ledger and the runner's printed row

`gradeCaseVerdict` refuses an unadjudicated divergence and prints the
exact row a human pastes, constructed at
`internal/harness/conform/case.go:467-477` and marshalled with
`json.MarshalIndent(row, "", "  ")` (`case.go:478`). The shadow's
pending row is that same `analysis.AdjudicationEntry`, marshalled the
same way, differing only in which subject shape is populated. So the
adjudication stream reads one artifact whether a row came from the
corpus runner or from the box, and rung 15 reads one artifact when it
evaluates the launch bar.

The shadow writes into `corpus/analysis/adjudications.json` never. The
box stages rows under `pending/`; the conductor fetches them read-only
at a checkpoint; the owner rules them and commits them into the corpus
on the workstation. That preserves the 3d handoff's rule that the
runner prints and writes nothing (`docs/HISTORY.md:295-302`) while
giving an unattended box a durable place to leave a row.

### The two schema additions

`analysis.AdjudicationEntry` gains two fields
(`internal/harness/analysis/adjudication.go:46-67`):

```go
MatcherCommit     string `json:"matcher_commit,omitempty"`
MatcherConfigHash string `json:"matcher_config_hash,omitempty"`
```

`checkAdjudicationEntry` (`adjudication.go:158`) gains the matching
rules: a shadow-subject row (`Album` and `ReleaseMBID` both set,
`adjudication.go:168`) requires `MatcherCommit`; a case-subject row is
refused if it carries either field. Both are `omitempty`, so the 21
committed corpus rows round-trip through a decode and re-encode with no
key gained or lost (corrected 2026-09-07 at the 4a-ii close: byte
identity was never available, because the committed ledger orders
`corrected_verdict` after `date` while `AdjudicationEntry` declares it
before `Reason`, and the file carries a literal `<` and three `>` that
Go's encoder escapes). Each rule ships a hand-authored fixture as an
inline Go literal variant in the package's own table test, which is the
package's landed precedent for a per-row rule; `testdata/` holds
cross-ledger fixtures only (corrected 2026-09-07 at the 4a-ii close,
Resolution R14).

### How a superseded-commit divergence is re-evaluated mechanically

`dubplate-server shadow reevaluate` walks every observation whose
`matcher_commit` differs from the running build's, replays that
observation's stored responses through the shadow's own replay
transport, re-runs `tag.Album` at the observation's own stored
`eval_date`, and:

- retires a divergence that no longer reproduces, writing
  `resolved/<obs-id>.json` with the retiring commit and both verdicts,
  and removing the staged pending row;
- leaves a divergence that still reproduces, restamping it with the
  current commit;
- records a **new** divergence on an observation that previously
  agreed, which is evidence the launch bar wants and which a
  retire-only design would silently lose;
- and records **un-re-evaluable** on an observation whose replay asks
  for a request the recording does not carry, leaving the row stamped
  and staged rather than retiring it. A newer matcher revision can take
  a different branch through `tag.Album` for the same input, exactly as
  a different `Compat` can (ruling 3's evidence), so the replay
  transport must answer a miss with a refusal rather than a 404 the
  flow would absorb. A silently retired divergence is the one outcome
  worse than an un-retired one.

The replay client takes `musicbrainz.Unmetered()` (`limiter.go:354`),
since a replay spends no real quota, which is the same choice
`gradeClient` makes for the corpus sweep (`case.go:330-336`).

It spends no MusicBrainz budget, reads no network, and reads no clock.
This is the mechanism ruling B's adjudication-load reduction rests on
(`R3:93`, `DA:534-535`), which INV O8 correctly reports as unbuilt and
unassigned anywhere. Rung 4 owns it, because a stamp with no
re-evaluator hands the owner exactly the load the ruling promised to
avoid.

**The shadow refuses to run on the box with an unstamped build, and the
refusal has to be able to announce itself** (fold, B6). `shadow
observe` errors rather than observing when the stamp is empty or `dev`,
because an unstamped divergence cannot be mechanically retired and is
therefore worse than no observation. Revision 1 paired that refusal
with a claimed early warning that does not exist. `main.version`'s doc
comment says it "is set via `-ldflags` at release build time"
(`cmd/dubplate-server/main.go:29-31`), but **no build path sets it**:
`build-linux` carries no `-ldflags` at all (`Makefile:6-7`), so the box
binary reports `dev` today and there is nothing for a matcher stamp to
join. And `newVersionCmd` prints `version` and exits zero
(`main.go:56-64`), so the smoke test at `scripts/deploy.sh:288-289`
cannot assert on a stamp under any implementation. Left as written, an
unstamped binary deploys clean, every timer pass refuses, and the
`OnFailure=` chain fires into a slug whose URL is unset and therefore
no-ops (`_salvage/cmd/dubplate/pingfail.go:41`) -- silence on both
channels for thirty days, discovered at rung 15, which is this rung's
own named worst outcome.

So three things move together, and deploy edit 6 below carries them:
`build-linux` gains `-ldflags "-X main.version=<git describe --tags
--always --dirty> -X main.matcherCommit=<git rev-parse HEAD>"`, both
variables, since neither is stamped today; `dubplate-server version`
prints both, one per line, in the pinned `"%-32s %s"` row form the
doctor already uses; and the deploy's smoke step asserts the matcher
stamp is neither empty nor `dev` and aborts the deploy when it is,
which `set -euo pipefail` already turns into a whole-deploy abort. The
staleness pre-task asks whether `build-linux` still sets both.

## The runner and its subcommand

### What it constructs

`tag.Album`'s signature is fixed by the 3d handoff
(`docs/HISTORY.md:273-279`), verified at
`internal/engine/tag/tag.go:68-69`:

```go
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
    tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error)
```

- **`src`.** `musicbrainz.NewClient` refuses a missing `UserAgent` or
  `Budget` rather than defaulting (`client.go:124-130`).
- **`items`.** The gap. See the next section.
- **`tun`.** `matcher.DefaultTunables()`
  (`internal/engine/matcher/tunables.go:266`) with the estate's four
  import assignments: `QuietFallback = "skip"`, `NoneRecAction =
  "skip"`, `Timid = false`, `Quiet = true`. The box's live config
  reduces to exactly those (`config/beets-config.yaml:16-23`, reduced
  in `corpus/profiles/estate.yaml:16-19`), and the estate's beets
  config sets no `match:` and no `musicbrainz:` block at all, so the
  whole distance and threshold surface runs at beets' shipped defaults
  (`config/beets-config.yaml` itself, and the harness's own statement
  of the consequence: "no committed profile sets a match: key, so every
  one of the 215 committed cases echoes the same twenty shipped
  defaults", `internal/harness/conform/defaults.go:16-18`; fold, m12).
- **`compat`.** Both, per ruling 3, with the second mode a re-score
  rather than a second flow.
- **`evalDate`.** The observation's capture date, stored in the fixture
  so re-evaluation reproduces it exactly.
- **`opts`.** Zero. `SearchIDs` selects the explicit-ID branch
  (`tag.go:78-85`) and the shadow has none.

**The construction cannot come from the harness.**
`TestEngineNeverImportsHarness`
(`internal/harness/conform/arch_test.go:141`) fails any
`internal/engine` file importing `internal/harness`, and
`gradeTunables` (`internal/harness/conform/defaults.go:72-79`) is on
the wrong side of that wall. Rung 4 therefore lands
`matcher.QuietImportTunables()`, an exported named constructor in
`internal/engine/matcher` returning `DefaultTunables()` with those four
assignments applied, and `conform`'s `gradeTunables` keeps its own
profile-merge path unchanged. This adds no new surface concept:
`Quiet`, `QuietFallback`, `NoneRecAction`, and `Timid` are already
`Tunables` fields (`tunables.go:187`, `:205`, `:210`, `:215`). One
definition, on the engine side, where both callers can reach it.

**The shadow reads no clock.** `TestEngineNeverReadsTheClock`
(`arch_test.go:152`) walks the whole `internal/engine` tree, `shadow`
included, and the walk descends into nested packages
(`arch_test.go:189-205`). So `shadow.Run` takes `now time.Time` as a
parameter, supplied by `cmd/dubplate-server`, exactly as `evalDate` is.
Every timestamp the pass writes derives from that one value, which also
makes a whole pass deterministic under test.

### The `ItemFields` producer, and the rung 4 answer

`matcher.ItemFields` is nineteen typed fields off one user item
(`internal/engine/matcher/info.go:124-152`). Two producers exist and
neither reads a file: `ProjectItemFields` projects a corpus case's
decoded JSON bag (`info.go:189`) and `synth.ProjectItems` derives from
a captured release document
(`internal/harness/synth/seeditems.go:28`). Rung 8's
`internal/engine/tagio` is named as planned
(`internal/engine/tag/doc.go:8-11`) and does not exist.

**There is no salvage-tier adoption available.** INV b.2 searched
`_salvage/` for FLAC metadata reading. The only FLAC contact is
`_salvage/internal/gate/gate.go:120-138`, which shells `metaflac
--show-sample-rate --show-bps` for the CD-quality check and reads no
Vorbis comment. The pre-pivot server delegated every tag question to
`beet` (`_salvage/internal/beets/beets.go:1-3`). So the salvage tier
offers one usable idea, the `exec.CommandContext` shape, and no code.

**Options.** (a) A minimal read-only FLAC reader inside
`internal/engine/shadow`, unexported. (b) Shell to `metaflac
--export-tags-to=-` per file. (c) Depend on rung 8's `tagio`.

**The rung 4 answer is (a), and it is a decision this spec takes rather
than a ruling.** (c) blocks rung 4 on lane T, and rung 4's deploy date
starts the shadow's evidence clock (`DA:525`), so waiting on another
lane delays that clock for no reason. (b) adds a subprocess and a
binary dependency to a hardened unit, costs a fork and exec per file,
and still leaves rung 4 owning the
Vorbis-key to beets-field mapping, which is the actual work. (a) pays
that same mapping cost and nothing else: a FLAC file is a `fLaC`
marker, a run of metadata blocks, and the two the shadow reads are
`STREAMINFO` (block type 0, for track length) and `VORBIS_COMMENT`
(block type 4). Pure stdlib, no dependency, roughly 150 lines.

**Its cost, stated rather than absorbed.** The reader is
shadow-internal and unexported, it is named provisional in its own doc
comment, and rung 8's plan carries the task of absorbing it into
`internal/engine/tagio` and repointing the shadow. Rung 8 owns the
writer and the full field surface; rung 4 owns nineteen fields read
one way.

**The mapping's evidence.** The Vorbis-comment key for each of the
nineteen fields is derived from pinned mediafile 0.17.0's own storage
styles (`tools/venv/.venv/lib/python3.13/site-packages/mediafile/__init__.py`,
for example `CATALOGNUMBER` at `:604` and `MUSICBRAINZ_ALBUMID` at
`:782`), the version `corpus/manifest.json` pins. The plan records the
derived table with its citation per row, which is U3's rule applied to
a mapping rather than a default. `ProjectItemFields`' typed-default
rule (`info.go:178-188`) governs an absent or unparseable tag: the
field takes its Go zero value, exactly as beets' `Item.get` does.

**Its test fixture is built in the test, not vendored.** A minimal FLAC
carrying a `STREAMINFO` and a `VORBIS_COMMENT` block is about a hundred
bytes and is constructed in Go inside the test, so no binary fixture is
vendored twice and the mutation that proves the reader fires is a byte
edit rather than a re-derivation.
`internal/harness/oracle/testdata/template.flac` stays where it is,
harness territory, unread by the engine.

### The budget, the pace, and the contact

`limiter.NewLimiter(<interval>)` (`limiter.go:110`), and
`Limiter.Budget("shadow", musicbrainz.ClassBackground, <cap>)`
(`limiter.go:139`). `ClassBackground`'s own comment names this caller:
"rung 4's shadow spends deferred background budget through this limiter
and never inline with the probe" (`limiter.go:16-18`). The cap is
ruling 5, and so is the interval: `musicbrainz.DefaultInterval`
(`limiter.go:62`) is what the 3d handoff records, and ruling 5 proposes
`2 * DefaultInterval` against the per-IP budget the box shares with its
own beets, which is the one conductor ruling this rung owes before 4a's
plan.

The User-Agent is composed in `cmd/dubplate-server`, not in the engine.
`ClientConfig`'s doc comment specifies it verbatim: "the identity is
site-specific, so it belongs to internal/config and reaches here
through cmd/dubplate-server's composition (`dubplate-server/<version> (
<contact> )`, on cmd/harness/capture.go's own precedent, with the
contact from Config.AdminEmail)" (`client.go:86-92`). `Config.AdminEmail`
exists and is normalized (`internal/config/config.go:60-64`). The
precedent refuses to run unidentified
(`cmd/harness/capture.go:116-120`), and the shadow does the same. Rung
4 is the first rung to talk to the live API from the server binary, so
rung 4 builds that function; INV O11 and `docs/HISTORY.md:308-310`
record it as an unmoved carry with no other owner. No new key: the
dedicated contact row stays the post-1b follow-on beside the four-knob
registration.

### The subcommand

`cmd/dubplate-server/shadow.go`, one file, matching the shape
`newCheckCmd` and `newDoctorCmd` already set: a constructor taking
`(environ []string, matcherCommit string, stdout, stderr io.Writer)`
returning (corrected 2026-09-07 at the 4a-ii close, conductor ruling B7:
the matcher stamp is a fourth parameter, because Resolution R3 rejects
the package variable a stamp table would otherwise have to write)
`*cobra.Command` (`cmd/dubplate-server/check.go:16`, `doctor.go:23`),
with `SilenceUsage` and `SilenceErrors` both true and main as the sole
stderr writer (`main.go:35-38`). Three verbs:

- `shadow observe` runs one capture, reconcile, and match pass. The
  timer's `ExecStart`.
- `shadow report` prints the launch-gate report.
- `shadow reevaluate` runs the mechanical re-evaluation above.

Registration edits `newRootCmd`'s single `cmd.AddCommand(...)` block
(`main.go:48-52`), and the doctor's integration rows come from
`internal/config`'s own tables rather than a hard-coded list, which is
how `doctor.go:23-59` already derives its wall and reserved-class rows.
**The threshold rows need a named call site** (fold, m13):
`config.Thresholds()` returns only what someone passed to
`config.Register`, called "from an explicit setup call rather than an
init function" by design (`internal/config/thresholds.go:19-30`), and
rung 1 registers none. Rung 4's two thresholds are registered in
`cmd/dubplate-server`, once, from the root command's own construction,
so registration order stays visible in the caller's code and both rows
appear in `doctor`. `RE:101` awards this
surface to rung 1's skeleton with "each rung adds its own file and
registers through it", so rung 4 adds its file and edits the one block,
and does not build a registration seam. INV O14's seam question is
answered by declining it: one line in one block, and rung 1b's registry
is where the wall half of that problem is already assigned.

### The launch-gate report

Printed in `doctor.go:63-67`'s pinned `"%-32s %s"` row form, so an
operator reads one table shape across the binary:

| Row | Source |
| --- | --- |
| clock start | `started-at`, written on the first successful pass |
| floor | the configured duration, default `720h` (ruling 4) |
| elapsed against floor | `now` minus the clock start |
| observed albums | split applied / skipped. Only the applied subset carries tier-one weight, because only there does the oracle name a release (`BS:257-260`) |
| excluded, gate-rejected | never reached beets, counted and named |
| excluded, multi-album directory | grouped to more than one album, so the album unit could not be resolved (the capture's grouping rule) |
| unreconciled | observations the reconcile step could not resolve, by reason |
| report gaps | incumbent run reports swept before the shadow consumed them |
| deferred | observations awaiting retry, counted separately for `ErrUnavailable`, `ErrBudgetExhausted`, and `ErrLimiterClosed` |
| divergences | by class, and by disposition over `adjudication.go:17-22`'s four values |
| false confident matches | `go-applies-oracle-skipped` rows whose owner disposition is anything but `oracle-wrong`. **Must be zero at the launch gate** |
| pending | staged rows awaiting the owner. **Must be zero at the launch gate** |
| matcher commits | each commit observed under, with its divergence count |
| budget | summed over `passes/`: each pass's `Budget.Name()` and `Budget.Spent()` (`limiter.go:400-408`), per-run and cumulative |
| store size | the state directory's own footprint (risk 5) |

The build spec's ruling 1 governs the reading: the observed album count
is reported and never gated (`BS:249-253`, `BS:812-815`).

**Two clauses gate, not one** (fold, M9). Launch bar part 3 reads
"zero unadjudicated divergences **and zero false confident matches
against live beets decisions**" (`BS:69-71`). The first clause is the
pending count, because `Pending` "is legal at every gate this package
runs; the launch gate, a different gate at a different time, refuses
it" (`adjudication.go:10-12`, `BS:900-902`). The second is the row
above: `go-applies-oracle-skipped` is the class a false confident match
hides in, and a row there that the owner disposed `oracle-wrong` is the
oracle's error rather than the matcher's, which is the build spec's
ruling 2, its tier-one refinement: "tier one is measured against
release truth rather than against the oracle where the oracle is
known-defective"
(`BS:64-68`). Every other disposition in that class counts against the
bar. The report prints the figure rather than leaving rung 15 to derive
it from a disposition breakdown.

## The deployment

### The unit

`systemd/dubplate-shadow.service`. The `[Unit]` section carries
`Description=`, `OnFailure=` below, and **the network-online pairing
every sibling oneshot on this box carries**,
`After=network-online.target` with `Wants=network-online.target`
(`systemd/dubplate-import.service:9-10`,
`systemd/dubplate-backup.service:13-14`; fold, M2). The estate treats
that pairing as a decision rather than a default, which is why the
serve unit spends fifteen lines saying why it does not carry it
(`systemd/dubplate-serve.service:12-27`); the shadow's own comment
carries its reason, that a `*:0/5` timer otherwise fires before the
network is up after a reboot and turns a boot into a failure ping.

`[Service]` is `Type=oneshot`, `User=music`,
`EnvironmentFile=/etc/dubplate/server.env` (the greenfield template's
render target below, and the line every sibling unit carries against
the incumbent's file, `dubplate-import.service:14`), and
`ExecStart=/usr/local/bin/dubplate-server shadow observe`. The
hardening template is `dubplate-serve.service`'s extended set
(`systemd/dubplate-serve.service:96-125`), not the oneshot baseline,
because the shadow makes outbound network requests. Carried verbatim:
`NoNewPrivileges`, `PrivateTmp`, `ProtectSystem=strict`, `ProtectHome`,
`ProtectKernelTunables`, `ProtectKernelModules`, `ProtectControlGroups`,
`RestrictNamespaces`, `RestrictRealtime`, `RestrictSUIDSGID`,
`LockPersonality`, `MemoryDenyWriteExecute`,
`SystemCallFilter=@system-service`, `SystemCallArchitectures=native`,
empty `AmbientCapabilities` and `CapabilityBoundingSet`, `UMask=0027`.

`RestrictAddressFamilies=AF_INET AF_INET6 AF_UNIX`, **without
`AF_NETLINK`**. The serve unit's own comment records why netlink joined
its set: glibc's `getaddrinfo` inside a CPython subprocess opens a
`NETLINK_ROUTE` socket, and "The Go process itself never needed it,
since Go's own resolver has its own fallback"
(`dubplate-serve.service:120-124`). The shadow runs no subprocess, so
the family stays denied.

`ReadWritePaths=` names the shadow's state directory and nothing else.
`OnFailure=dubplate-ping-fail@shadow.service`: the slug resolves
`DUBPLATE_PING_SHADOW_URL` straight from the process environment,
bypassing config load, and an unset URL no-ops rather than erroring
(`_salvage/cmd/dubplate/pingfail.go:41`, "ping-fail: no ping URL
configured; no-op"; the estate's own precedent for wiring a slug ahead
of its mint is `systemd/dubplate-serve.service:29-34`; fold, m9). So
the line is wired now and no key is minted, which matters because the
whole `DUBPLATE_PING_*` class is reserved for rung 9
(`internal/config/keys.go:118-124`). **Say plainly what that buys: no
monitoring at rung 4.** The helper reads
`EnvironmentFile=/etc/dubplate/env` (`systemd/dubplate-ping-fail@.service:11`),
the incumbent's file, and with no `DUBPLATE_PING_SHADOW_URL` minted the
`OnFailure=` line is wired and inert for the whole shadow window. It is
a placeholder for rung 9, not a channel; the stamp assertion at deploy
and the journal check at the sitting are what actually catch a shadow
that observes nothing. The helper execs the incumbent binary and
retires with it at rung 16.

`systemd/dubplate-shadow.timer`, `OnCalendar=*:0/5`,
`RandomizedDelaySec=60`, `Persistent=true`. The five-minute interval is
the quiescence constraint above, and the unit's comment states it with
its citation rather than leaving a reader to infer it. **`Persistent`
is `true`, against revision 1** (fold, M3): a missed pass does not cost
"at most a later capture", because the incumbent's run reports are
swept to the newest thirty (`importer.go:389`, `:533-548`,
`_salvage/internal/config/config.go:172`) and its timer fires `*:0/15`,
so the oracle signal for every album imported during an outage longer
than roughly seven and a half hours is gone permanently. A catch-up run
after a reboot is exactly the pass that recovers what is still
recoverable.

The gate already covers both files. `scripts/check.sh:137` globs
`systemd/*.service systemd/*.timer systemd/*.path` and the
missing-executable filter at `:139` matches
`/usr/local/bin/[A-Za-z0-9_.-]+`, which covers `dubplate-server`. A
malformed unit reds the gate with no new wiring.

### The state directory

**`DUBPLATE_STATE_DIR` on the box is `/var/lib/dubplate`, minted new,
not the incumbent's `/var/lib/musicbox`.** This spec takes the
decision rather than raising it, because the argument runs one way. A
shared directory means the unit's `ReadWritePaths` grants write over
the incumbent's lock, tag logs, and run reports, which is precisely the
grant the read-only guarantee exists to withhold. The standing "keep
the data path name" ruling (`docs/HISTORY.md:2218-2220`, the kept
identifiers bullet; `:2211-2213` is the compose project-name pin, fold
m4) was about
migrating a live directory, not about creating a new one, so nothing is
contradicted. The boundary retires cleanly with the incumbent at rung
16.

Cost: one pre-create step and one SELinux verification. A new directory
under `/var/lib` on an enforcing box inherits `var_lib_t`, fine for
these services but to be verified once on the first real deploy
(`~/.claude/skills/vps-conventions/SKILL.md:43-45`). The identical
check is already carried unclosed for `/var/lib/musicbox`
(`docs/HISTORY.md:2327-2328`), so the deploy sitting closes both.

### The environment file

`/etc/dubplate/env` carries 39 `DUBPLATE_*` names
(`env/dubplate.env.template`). Two are owned rows; 37 produce a load
error, and two rows the greenfield binary requires
(`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_SITE_NAME`) are absent, having been
renamed from `DUBPLATE_LIBRARY_DIR` and `DUBPLATE_SERVE_SITE_NAME`
(`internal/config/keys.go:165-166`). A shadow unit reading that file
fails at load, by design, every time.

So rung 4 adds a second template, `env/dubplate-server.env.template`,
rendered to `/etc/dubplate/server.env`, carrying the four owned rows
plus the shadow's integration rows. Both files coexist until rung 16
retires the incumbent. The existing template's header already
anticipates this: "this manifest describes the frozen pre-pivot
binary's wall ... this manifest evolves at rung 4"
(`env/dubplate.env.template:1-3`). Rung 4 amends that header to name
its sibling.

**Every greenfield row is named as a literal or as a store lookup, and
the rename is settled here** (fold, B5). The renderer resolves every
`NAME=` line through one flat shell namespace
(`scripts/deploy.sh:181-182`) and exits 1 listing every miss before any
remote step (`:190-194`), so a row that is neither is a deploy abort,
and a row that silently falls through to the incumbent's exported value
is worse than an abort. Revision 1 left both cases unstated and got one
fact wrong.

| Greenfield row | How it resolves |
| --- | --- |
| `DUBPLATE_STATE_DIR` | **literal** `/var/lib/dubplate`. Not a store lookup: the incumbent's own template carries a bare `DUBPLATE_STATE_DIR=` whose exported value is deliberately `/var/lib/musicbox` (`env/dubplate.env.template:59`, `:74`), and inheriting it would grant the shadow's unit write over the incumbent's lock, tag logs, and run reports |
| `DUBPLATE_LIBRARY_ROOT` | **literal** `/srv/music/library`. The age store carries `DUBPLATE_LIBRARY_DIR`, the pre-rename name (`internal/config/keys.go:165`), so a bare `NAME=` line aborts the render |
| `DUBPLATE_SITE_NAME` | **literal**, the estate's masthead text. The store carries `DUBPLATE_SERVE_SITE_NAME`, the pre-rename name (`keys.go:166`), and `~/.dotfiles/secrets/registry.md` records it as non-secret masthead text minted 2026-09-01. A rename in the age store would be a second place to keep one string in step; the literal keeps it in the estate's own operational record, where `BS:283-288` puts it |
| `DUBPLATE_ADMIN_EMAIL` | **store lookup**, unchanged. Already minted and already in the incumbent manifest (`env/dubplate.env.template:25`) |
| the six `DUBPLATE_SHADOW_*` rows (corrected 2026-09-07 at the 4a-ii close from "the four"; the key table carries six, the enable flag included) | **literals**: the estate's inbox, oracle state and review paths, the floor, and the budget. Putting `/srv/music/inbox` in the age store is a category error |

**Secrets: none, and now that is a checked statement rather than an
assumed one.** No row above is a secret, `secret-set.sh` is not run,
and `~/.dotfiles/secrets/registry.md` gains no row. Revision 1 said
`DUBPLATE_SITE_NAME` "stays minted in the age store from the
wind-down"; a name-only check of the store says otherwise, and that
sentence is deleted rather than repaired. The plan states the
no-mint conclusion explicitly, because CLAUDE.md's rule would otherwise
be read as requiring a mint for every new template name, and it states
the rename as the reason.

### The deploy.sh changes

**The binary path the ladder entry calls "the deploy.sh path rung 1
repointed" already exists.** `deploy.sh` cross-compiles through the
Makefile's `build-linux` target, rsyncs the result to
`/usr/local/bin/dubplate-server`, chmods it 0755, and smoke-tests
`dubplate-server version` with a failure aborting the whole deploy
(`scripts/deploy.sh:273-289`), and its boundary comment names this rung
by name: "Greenfield units start exec-ing /usr/local/bin/dubplate-server
from rung 4" (`:257-260`). Rung 4 consumes that path rather than
building it, and adds the first unit that execs it.

`scripts/deploy.sh` is 443 lines, nine idempotent steps, each with a
`--check` branch. Six edits and one new artifact:

1. **`UNIT_FILES`** gains `dubplate-shadow.service` and
   `dubplate-shadow.timer` (`:56-64`). The manifest edit lands in the
   same commit as the files, because rsync exits 23 on a missing source
   and aborts the whole deploy (`SKILL.md:70-74`).
2. **`TIMER_UNITS` is unchanged** (`:407`). The first `systemctl enable
   --now dubplate-shadow.timer` is a hand step at the deploy sitting,
   on the `dubplate-serve.service` precedent (`:49-55`, `:400-405`).
   The reason is not exposure, which the shadow has none of; it is that
   an adopter's deploy should not silently start observing, and that
   the enable is the act that starts the launch clock. After the first
   enable, systemd persists it and later deploys change nothing.
3. **A pre-create step for `/var/lib/dubplate`**, in the shape of
   `:357-359`: `run_remote "mkdir -p <dir> && chown music:music
   <dir>"`. Mandatory before the unit's first start, or systemd fails
   with "Failed to set up mount namespacing" (`SKILL.md:39-42`,
   restated at `scripts/deploy.sh:350-356`).
4. **A parameterized render-and-install helper.** The render loop
   (`:175-194`) and the install block (`:291-316`) are both written for
   one path, `REMOTE_ENV` at `:28`. Rung 4 lifts them into one helper
   taking a template and a remote path, and calls it twice. The helper
   keeps the in-memory-only property: no decrypted value ever touches
   local disk or a log.
5. **A literal column in the renderer.** The loop today requires every
   `NAME=` line to resolve to a non-empty exported value and fails
   listing every miss (`:190-194`), with no way to express a literal.
   The shadow's paths and the floor are non-secret, and putting
   `/srv/music/inbox` into the age store is a category error. So a
   template line carrying a value on its right-hand side is a literal,
   a bare `NAME=` must still resolve from the environment, and both
   template headers say so. Three lines in the loop, and backward
   compatible: no tracked template line carries a right-hand value
   today, and the loop reads only the name (`:181-182`).

   **This extends a row-state vocabulary `RE:100` already reasoned
   through for this file** (fold, m15): "a name whose value arrives
   later lands as a commented pending row on the `config.WallRow.Pending`
   precedent, never as a live `NAME=` line, because `deploy.sh` fails
   the whole render on one unresolved name." That covers a value that
   has not been minted yet. The literal is a third state, for a value
   that will never be minted because it is not a secret, and the
   greenfield template header names all three so a later rung does not
   re-derive them.
6. **The version and matcher-commit stamps, which do not exist today**
   (fold, B6). `build-linux` carries no `-ldflags` at all
   (`Makefile:6-7`), so this edit adds both variables rather than
   joining an existing flag: `-ldflags "-X main.version=... -X
   main.matcherCommit=..."`, the first from `git describe --tags
   --always --dirty` and the second from `git rev-parse HEAD`, computed
   in `deploy.sh` and passed through to `make build-linux`.
   `dubplate-server version` prints both rows
   (`cmd/dubplate-server/main.go:56-64` today prints one line), and the
   existing smoke step (`scripts/deploy.sh:288-289`) asserts the
   matcher stamp is neither empty nor `dev`, which `set -euo pipefail`
   turns into a whole-deploy abort. Without this edit the refusal in
   `shadow observe` is unobservable for thirty days.

Every new step carries its `--check` branch, per the standing rule.

### The quiet-box discipline

There is no mechanism in the repo that checks it. The rule is doctrine
(`CLAUDE.md` safety rails) and rung assignment (`BS:262-265`, `PA:299`),
and the closest thing to a mechanical check is `--check` mode, which
diffs every step and applies none (`scripts/deploy.sh:73-88`), plus the
`remote_reachable` probe that makes an unreachable box report honestly
(`:113-121`).

Rung 4 invokes it as a **human procedure**, and says so rather than
implying a check exists. The procedure, at the deploy sitting: confirm
no import or backup run is in flight and none is due inside the window
(the import timer fires `*:0/15` with 300 seconds of jitter,
`systemd/dubplate-import.timer:9-11`; backup fires daily at 11:30 UTC
with 600 seconds of jitter, `dubplate-backup.timer:9-10`); run
`deploy.sh --check` and read every step's diff; apply; enable the timer
by hand; verify the first pass in the journal; verify the SELinux
context on `/var/lib/dubplate` and on `/var/lib/musicbox`; record the
root filesystem's free space as the store's sizing baseline and whether
`dubplate-disk-check` watches it; record the deploy date. Read-only ssh
at every step until the apply.

**One operator rule holds after the sitting too:** a hand-run shadow
verb goes through `systemctl start dubplate-shadow.service` or
`systemd-run` carrying the unit's own properties, never the bare binary
as root, because the read-only guarantee is the unit's and not the
binary's.

### The rollback

**Two rollbacks, not one, because they are not symmetric** (fold, M7).

**Stop observing.** `systemctl disable --now dubplate-shadow.timer
dubplate-shadow.service`. Free, reversible, and the clock is preserved:
`started-at` stays on disk, so re-enabling the timer resumes the same
evidence window rather than starting a new one.

**Discard the evidence.** Removing `/var/lib/dubplate` deletes every
observation, every stored response the mechanical re-evaluator depends
on, and `started-at` itself. This restarts the floor from zero and
throws away up to thirty days of accumulated evidence toward the
launch bar's own material (`DA:524-527`). It is
**the owner's explicit decision, never a tidy-up step**, and the
procedure says so in those words.

Either way, **the library, the beets database, the inbox, the review
tree, and the incumbent's own state are untouched by construction**,
because the unit's only writable path is the shadow's own state
directory and the shadow's own code opens the incumbent's lock never.
There is no data migration to reverse and no unit to restore. The
frozen incumbent binary is never written by `deploy.sh` (`:251-260`),
so the four guardian units are unaffected by any rung 4 deploy or
rollback.

### Universality

The runner is a product capability: "the migration-shadowing capability
any beets-migrating adopter wants, not scaffolding" (`BS:335-337`,
`S4`). It is **adopter-optional, behind `DUBPLATE_SHADOW_ENABLED`**,
which is INV's O4 answered through U4's own rule (`BS:150-158`) and
through the build spec's ruling 1 and its adopter framing
(`BS:806-811`). A design that only worked against this box's directory
layout would contradict the ruling that set the floor.

The unit, the timer, and the environment template are **the estate's
operational record** (`BS:284-291`), never an adopter kit; adopter
deployment artifacts belong to the packaging initiative
(`BS:876-884`). Those two halves stay separable: every estate path
reaches the runner through `internal/config`, and the runner contains
no hostname, path, email, or personal name (`BS:116-120`).

## Config keys

**Rung 1b has not landed** (`docs/STATUS.md:29`), so
`internal/config/keys.go` is still the one file every rung's keys live
in, and rung 4 is the only live code lane. Rung 4 therefore registers
there, under the existing pattern, rather than through a registry that
does not exist. `PA:158` awards the wall to rung 1b with "no lane edits
`keys.go`"; that rule protects against two lanes colliding on one file,
and with one lane there is nothing to collide with. Rung 1b later
absorbs rung 4's table as data rather than re-deriving it.

The rows, per ruling 1's option (a):

| Key | Kind | Default | Rationale recorded at |
| --- | --- | --- | --- |
| `DUBPLATE_SHADOW_ENABLED` | integration flag | off | the class's own comment |
| `DUBPLATE_SHADOW_OBSERVE_DIR` | required-when-enabled | none | the inbox the incumbent consumes |
| `DUBPLATE_SHADOW_ORACLE_STATE_DIR` | required-when-enabled | none | where the incumbent's `report-*.json` and `import-*.log` live |
| `DUBPLATE_SHADOW_REVIEW_DIR` | required-when-enabled | none | where a skipped album is filed, the skip half's corroboration |
| `DUBPLATE_SHADOW_FLOOR` | threshold, duration | `720h` | the build spec's ruling 1, 30 days (`BS:804-815`) |
| `DUBPLATE_SHADOW_BUDGET` | threshold, int | `60` | ruling 5's arithmetic |

`DUBPLATE_LIBRARY_ROOT` is already an owned row (`keys.go:19-25`) and
is where an applied album's files land, so the shadow reuses it and
mints nothing. `DUBPLATE_BEETS_DB` stays retired and unamended, because
the shadow never opens the beets database.

**The threshold registry cannot express a duration.** `Threshold` is
`{Key, Default, Doc}` with `ParseInt` as the only parser
(`internal/config/thresholds.go:13-59`), and `thresholdStatus` calls
`ParseInt` unconditionally, so a duration registered today renders as
`invalid:` in the doctor table (`cmd/dubplate-server/doctor.go:100-108`).
Rung 4 adds a `Kind` field to `Threshold`, a `config.ParseDuration`
beside `ParseInt`, and a switch in `thresholdStatus`. `docs/HISTORY.md:678-685`
records the identical gap for rung 3's three float thresholds; a
`KindFloat` follows the same shape and stays that follow-on's, not
rung 4's.

**The `.gitleaksignore` fragility, restated because it reds the gate
for every lane at once.** The single allowed finding is pinned by line
number: `.gitleaksignore:1` reads
`internal/config/keys_test.go:generic-api-key:66`, and line 66 today is
the `DUBPLATE_R2_BUCKET` fixture row
(`internal/config/keys_test.go:66`). Any insertion above that line in
that file moves the finding and reds `scripts/check.sh`'s gitleaks leg.
The rule for every rung 4 task touching `keys_test.go`: add below line
66, or re-fingerprint `.gitleaksignore` in the same commit. Rung 1b
carries the same rider (`PA:54`).

## What rung 4 does not build

Stated because the boundary is easy to drift across.

- **Rung 8's `internal/engine/tagio`.** Rung 4 builds a shadow-local,
  unexported, read-only FLAC reader covering nineteen fields. It writes
  no tag, reads no format but FLAC, and is absorbed by rung 8.
- **The store.** `internal/engine/store` and the ops seam are rung 6's
  (`BS:573-575` puts the divergence log in the state directory
  precisely because the store arrives later). Rung 4 writes files.
- **A beets database reader.** The shadow never opens `library.db`.
- **A sqlite driver.** `go.mod` gains no dependency at rung 4.
- **Applied fields.** Rung 3 computes none, so the shadow's Go side has
  none and the comparison is verdict-level (`RE:186`, `R3:247-254`).
- **Monitoring.** The `DUBPLATE_PING_*` class is rung 9's
  (`keys.go:118-124`). Rung 4 wires the `OnFailure=` line and mints no
  key.
- **The pipeline wiring.** Rung 9's.
- **Corpus writes.** The shadow never writes `corpus/analysis/*`. Rows
  are staged on the box and committed by the owner on the workstation.
- **Adopter-facing deployment artifacts.** The packaging initiative's
  (`BS:876-884`).
- **`internal/lock`.** See the store section.

## What rung 4 needs from rung 3

All shipped, and all four items are the 3d handoff's own "rung 3 to
rung 4 handoff" section (`docs/HISTORY.md:271-303`), read against the
tree:

1. **The entry point**, `tag.Album` with `ctx` first, `evalDate` a
   parameter and never a clock read, `opts.SearchIDs` the only optional
   input, `Proposal` carrying `CurArtist`, `CurAlbum`, `Candidates`
   ascending and stable, `Recommendation`, and `Action`
   (`docs/HISTORY.md:273-279`, verified at `tag.go:23-29`, `:68-69`).
2. **The divergence stamp**, joining `analysis.AdjudicationEntry`
   beside the existing fields, with the shadow subject shape already
   modeled and carrying zero committed rows, so rung 4 is its first
   producer (`docs/HISTORY.md:280-286`, `adjudication.go:32-39`).
3. **The budget class**: `NewClient` requires a `Budget`, every request
   spends one token, the class is `ClassBackground` deferred behind
   rung 7's user-present probe, the pace is `DefaultInterval`, and rung
   4 branches on `ErrUnavailable` to defer and on `ErrNoCandidates` to
   record "no match" (`docs/HISTORY.md:287-293`). Rung 4 adds two
   branches the handoff does not name, `ErrBudgetExhausted` and
   `ErrLimiterClosed`, both deferring (fold, M1), and proposes halving
   the pace against the shared per-IP budget, which is the conductor
   ruling ruling 5 records.
4. **The pending-divergence file format**, `adjudicator: agent`, and
   the launch gate refusing any pending row
   (`docs/HISTORY.md:295-302`). One correction rung 4 does not inherit
   (fold, m3): the handoff's "all five prose fields non-empty" lists
   `citation` among them, and `checkAdjudicationEntry` does not check
   `citation`. The five it checks are `suspected`, `release_truth`,
   `reason`, `adjudicator`, and `date`
   (`internal/harness/analysis/adjudication.go:176-189`). Rung 4's rows
   carry a citation anyway, as all 21 committed rows do, but as a
   convention rather than a check.

Rung 3 also hands forward its **twenty-one declared corpus-coverage
gaps** (`docs/HISTORY.md`, the declared-gaps section) as a running
total rung 4 reads rather than re-counts, and the **rung 15 re-run
criterion**, which every rung 3 pass close carries forward and rung 4's
close carries too.

## Edges to later rungs

**Rung 9 consumes nothing from rung 4** (`DA:48` reads `4 -> 9` as
none). The traffic runs the other way: rung 9 re-verifies the limiter's
caller-priority policy against its real callers (`BS:698-700`,
`R3:243-245`), and the shadow is one of them. Rung 4's close reports
its measured `ClassBackground` behavior beside a live
`ClassUserPresent` and `ClassBatch` load, so rung 9 re-verifies against
a measurement rather than an assumption. `reservationInterval` is the
limiter's one knob (`limiter.go:43`), and the round-robin over classes
below the served one replaced a rule that "starved `ClassBatch`
outright whenever a probe stream and the shadow were both queued"
(`docs/HISTORY.md:742-748`).

**Rung 15 consumes the shadow record as a launch-bar part.** `4 -> 15`
is both an artifact gate and a calendar gate: the configurable floor
must have elapsed and every arrived divergence must be adjudicated
(`DA:113-115`). Launch bar part 3 is "a clean shadow record", zero
unadjudicated divergences and zero false confident matches against live
beets decisions (`BS:69-71`). The observed album count is reported and
not gated (`BS:251-253`). Rung 15 also reads the case-shaped fixtures
under `observations/` as candidate corpus material, with `Expected`
still nil.

**Rung 16 consumes the same record and the box.** `4 -> 16` is hard.
Rung 16 needs a quiet box and converges systemd, `deploy.sh`, compose,
and cloud-init (`BS:778-785`), so rung 4's unit must be stoppable and
removable inside that convergence, which the rollback above already
is. `/etc/dubplate/env` and the incumbent binary's deploy step retire
together at rung 16 (`scripts/deploy.sh:257-260`), and
`/etc/dubplate/server.env` becomes the only environment file.
`/var/lib/dubplate` survives; `/var/lib/musicbox` retires with the
incumbent.

**Rung 17 reconciles the runner's documentation.** `4 -> 17` is soft:
the shadow runner outlives its rung as the standing
migration-shadowing capability, and close-out reconciles its docs
(`DA:211-212`, `BS:794-798`).

**One inbound edge rung 4 must not break.** Cell (8,4) is `SOFT-B`:
rung 8's `tagio` may land unwired inside the deployed shadow binary
(`RE:190`, `RE:202`). Rung 4's plan states it in one sentence: **the
deployed binary carries landed-but-unwired packages from other lanes,
and none of them is reachable from the `shadow` subcommand's own call
graph.** The shadow's FLAC reader is its own, and repointing it at
`tagio` is rung 8's task, never a silent adoption.

## The evidence clock

**What starts it.** The first successful `shadow observe` pass, which
writes `started-at` once and never rewrites it. That is within minutes
of the deploy sitting's hand enable, so it agrees with "rung 4 starts
the clock at deploy" (`DA:315`) by construction rather than by
bookkeeping. **Rung 4's close records that date as a first-class figure
in `docs/HISTORY.md` and `docs/STATUS.md`**, alongside the running
observed-album count, because the launch gate's floor check reads it
(`DA:525`, `PA:303`) and reconstructing it later from journal entries
is worse than writing it down once.

**What the floor is.** A configurable duration with a 30-day default
(the build spec's ruling 1, `BS:804-815`), the estate's own value at
ruling 4.

**What is reported at the launch gate.** The report table above. Three
figures carry weight and only one gates: the observed album count is
reported (`BS:812-815`), the divergences are enumerated by disposition,
and the pending count must be zero (`adjudication.go:10-12`).

## The pass split

M band, **two passes**. The deployment half argues for the split, for
one reason: 4b cannot close without an owner sitting inside it (ruling
6, corrected 2026-09-07 at the 4a-ii close from "ruling 7"; the deploy
sitting is Resolution R10's ruling 6), and a pass that stalls waiting on
a sitting is a pass whose ceiling
and checkpoint arithmetic stop meaning anything. Splitting also puts
the merge point exactly at the phase boundary, since M1 is the deploy
(`DA:335`).

| Pass | Scope | Exit gate | Tree |
| --- | --- | --- | --- |
| **4a** The runner | `internal/atomicfile` re-landed; `matcher.QuietImportTunables`; the config integration class, the duration threshold, the six rows and their `Register` call site; the FLAC reader and its mapping; the capture with its grouping and partial-read rules; the reconcile step's basename join and consumed-report ledger; the match step; the store, including `passes/`; the recording transport and the `DefaultCompat` re-score; the divergence classification with its four deferring branches; `analysis.AdjudicationEntry`'s two fields and their checks; the marshalling in `cmd/dubplate-server`; the `shadow` subcommand's three verbs; the report; `reevaluate` | `bash scripts/check.sh` green, plus one end-to-end `observe` run against a hand-authored inbox and a hand-authored run report in the pass's own `testdata`, producing one agreement, one divergence of each class, one excluded multi-album directory, and one deferred observation of each error branch reachable in a single pass (corrected 2026-09-07 at the 4a-ii close: `Budget.wait` cannot be injected per album, so `ErrLimiterClosed` is proven in its own pass, Resolution R7) | Main tree |
| **4b** The deployment | `systemd/dubplate-shadow.{service,timer}`; `env/dubplate-server.env.template` with its literal rows and the existing template's header amendment; deploy.sh's six edits, the sixth being the two `-ldflags` stamps and the smoke-test assertion; `dubplate-server version` printing both stamps; the fresh-context unit read; the deploy sitting; the SELinux verifications; the recorded deploy date | `check.sh` green, `deploy.sh --check` clean, a non-`dev` matcher stamp asserted by the smoke test, the sitting completed, the first pass verified in the journal | Main tree |

Rung 4 writes `internal/harness/analysis`, and the standing policy is
that corpus and harness work never moves to a worktree (`PA:194-204`,
item 5). **Rung 4 holds the primary tree for both passes**, and lane A
or lane T takes the worktree, one at a time, exactly as during rung 3.

### Contested surfaces, and the rows the conductor owes

The conductor writes these into `PA`'s ownership table before the lane
launches, per lane-launch checklist item 2 (`PA:197`). Seven are new
and two amend an existing row.

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/shadow` (new) | rung 4, created by it; no other lane writes it | rung 15's re-run, rung 17's docs |
| `internal/atomicfile` (new, re-landed from salvage) | born by rung 4 pass 4a | rungs 6, 9, 13 |
| `cmd/dubplate-server` (`main.go`'s `AddCommand` block, `doctor.go`'s rows) | rung 1's skeleton; each rung adds its own file and registers through it (`RE:101`) | every rung adding a verb |
| `internal/config` (`keys.go`, `thresholds.go`) | rung 1b's registry when it lands; **until then rung 4 as the only live code lane**, by the grant this spec records | every rung declares its keys |
| `systemd/`, `scripts/deploy.sh`, `env/` | rung 1's; rung 4 adds two units, one template, and six deploy edits; rung 16 converges all of it | rungs 5, 7, 12, 16 |
| `internal/harness/analysis` (amends the existing row) | lane E1 as shipped; rung 3 passes 3b and 3d; **rung 4 adds `MatcherCommit`, `MatcherConfigHash`, and their checks** | every conformance consumer |
| `internal/engine/matcher` (amends the existing row) | rung 3 (lane E2); **rung 4 adds `QuietImportTunables` and nothing else** | rung 4's shadow, rung 9's pipeline, rung 15's re-run |
| the state directory layout (new; fold, m14) | **rung 4**, first to deploy one, and its unit's `ReadWritePaths` depends on it (`RE:102`) | rung 5's nonces and per-email epoch; rung 6's store supersedes neither |
| `corpus/analysis/adjudications.json` and the divergence schema (new; fold, m14) | lane E1 authored it; **rung 4 adds the shadow divergence rows and the matcher-commit stamp** as a schema addition, never by hand-editing a derived record (`RE:198`) | rungs 3, 8 and 15 read it; `disposition` stays owner-only |

The conductor also owes: the lane registry entry (`PA:194-204`, item
4), rung 4's sub-ceiling against the phase ceiling (item 6), the
runaway guard (item 3), and the `docs/STATUS.md` lane-registry update,
which is conductor-owned by the standing rule.

### Band and ceiling

4a is M, 4b is S. Rung 3's recent evidence: 3d closed at roughly 5.15M
(`docs/STATUS.md:26-28`). 4a has no corpus derivation, no venv, and no
tie-mirroring class of problem, and its hardest single piece is the
FLAC reader plus its mapping table. **Propose 5M for 4a and 1.5M for
4b**, each re-derived from 4a's actual report before 4b's plan is
drafted, on the 2c precedent of refining a ceiling from a measurement
rather than asserting one for the whole rung. Both sub-ceilings sum
into the evidence phase's combined ceiling, which the conductor states
at 4a's plan alongside every other live lane's.

## The staleness pre-task

Every rung 4 pass opens with a staleness pre-task before its first
implementer dispatch: one fresh-context agent reads this spec and the
pass plan against the tree as shipped, and reports blockers with
evidence rather than fixes. The conductor rules each one before the
first task dispatches. Plan staleness is the build's named dominant
defect class, and this rung's exposure is real: rung 1b may land
mid-rung and change the wall's shape, and the adjudication stream may
move dispositions.

The first three are blockers; a "no" stops the launch.

- **Has rung 1b landed?** If yes, the integration class and the
  duration threshold register through its registry and never by editing
  `keys.go`, and ruling 1's option (a) is re-expressed in that
  registry's own terms.
- **Is the lane registry consistent with where this pass intends to
  run, and is the primary tree free?** The live-executor check is the
  registry, never `pgrep -f Projects/dubplate`, which matches every
  lane once a worktree exists.
- **Is `.gitleaksignore:1` still
  `internal/config/keys_test.go:generic-api-key:66`, and is line 66
  still the row it names?**
- Does `config/beets-config.yaml` still set `group_albums: yes`? The
  capture's grouping rule and its excluded-directory count are written
  against it.
- Does `Makefile`'s `build-linux` set both `-ldflags` stamps, and does
  `dubplate-server version` print the matcher commit? A "no" at 4b
  means deploy edit 6 has not landed and the stamp refusal is
  unobservable.
- Are `defaultReportRetention` still 30 and `RetentionDays` still 60
  (`_salvage/internal/config/config.go:155`, `:172`)? The reconcile
  step's report-gap handling and the timer's `Persistent=true` are
  written against both.
- Is `tag.Album`'s signature unchanged, and are `Proposal`'s five
  fields unchanged?
- Are `analysis.AdjudicationEntry`, `AdjudicationSubject`, and
  `checkAdjudicationEntry` unchanged since 3d, and does
  `corpus/analysis/adjudications.json` still carry zero shadow-subject
  rows?
- Are `musicbrainz.ClassBackground`, `Limiter.Budget`,
  `DefaultInterval`, `ErrUnavailable`, and `ClientConfig` unchanged?
- Have `systemd/`, `scripts/deploy.sh`, `Makefile`, or
  `env/dubplate.env.template` moved since this spec was written?
- Does the incumbent's quiescence default still sit at ten minutes and
  its import timer at `*:0/15`? The five-minute shadow interval depends
  on both.
- Do `docs/HISTORY.md` and `docs/STATUS.md` still describe the box as
  this spec assumes: four enabled timers, `dubplate-serve.service` and
  the nudge units shipped and disabled, and the two albums awaiting
  review in `/srv/music/review/2026-08-31/geoff/`?
- Has `internal/config` changed shape?

## Risks

**1. A clean but empty record.** The wind-down retired every
contributor intake path, so the shadow observes what Geoff adds for his
own listening (`BS:243-249`, `DA:534-536`). Thirty days of zero albums
is a weak record even though the build spec's ruling 1 makes it
non-gating. Mitigation: the report separates observed from excluded and
prints both, that ruling already moved the evidentiary weight onto the
corpus's real-catalog and adversarial strata, and the deliberate
backlog feed `DA:535-536` suggests is named as an owner option at the
deploy sitting rather than discovered at rung 15. **A backlog feed also
raises the multi-album-directory count**, which the capture's grouping
rule excludes, so the deploy sitting names the excluded count as the
figure to watch at the first checkpoint.

**2. The capture races the incumbent's import.** Mitigation: the poll
interval is strictly shorter than the incumbent's quiescence window,
the capture is idempotent under a file-set digest, a capture that
cannot read every entry it listed is discarded rather than stored, and
the reconcile step's D and H backstops resolve the skip half without
the run report. An observation that still cannot be resolved is
recorded as `unreconciled` and counted, never silently dropped.

**3. The shadow stalls the box's own import.** A held
`/var/lib/musicbox/import.lock` makes the incumbent write a `LockHeld`
report, ping success, and return nil (`importer.go:397-410`), so a
shadow holding it would look green while the box imported nothing.
Mitigation: the shadow's code opens that path never, with the reason in
the code; its own store needs no lock; and the residual hand-run race
costs at most a re-captured fixture.

**4. The FLAC reader is rung 8's territory arriving early.**
Mitigation: it is unexported, shadow-local, named provisional in its
own doc comment, and rung 8's plan carries the absorption task. The
alternative, waiting on lane T, delays rung 4's deploy, which starts
the shadow's evidence clock.

**5. The store grows unboundedly.** Ratified ruling B keeps response
bodies (the retirement note in the rulings section). A MusicBrainz
release document with the shadow's `inc=` sets runs to a few hundred
kilobytes, and an album costs at most eleven of them, so a hundred
observed albums is comfortably under a gigabyte. **The sizing is
against the root filesystem, not the library volume** (fold, m5):
`/var/lib/dubplate` sits on root, and the 50 GB `musicbox-data` volume
(`scripts/provision.sh:16-17`) is `/srv/music`, which is why backup
staging was deliberately kept on the volume instead
(`scripts/deploy.sh:367-369`). Mitigation: the report prints the
store's size at every run; the deploy sitting records the root
filesystem's free space as the baseline and confirms whether
`dubplate-disk-check` watches it; and the trigger to build a retention
knob is named rather than the knob built now.

**6. The config integration class collides with rung 1b.** Mitigation:
it is a table in `keys.go` beside `reservedClasses`, in the same shape,
so 1b absorbs it as data. The staleness pre-task's first blocker
catches the case where 1b lands mid-rung.

**7. `.gitleaksignore`'s line pin reds the gate for every lane.**
Mitigation: the rule above, and the pre-task check.

**8. A superseded-commit divergence consumes owner adjudication
anyway.** This is what happens if `reevaluate` is cut for scope.
Mitigation: it is a named 4a deliverable, not a later rung's, and the
response store it depends on is a consequence of ratified ruling B
rather than an open question.

**9. The first production deploy of greenfield code.** Mitigation:
`deploy.sh --check` first and every step's diff read; a fresh-context
read of both unit files before any box write; the sanctioned hand
enable; the quiet-box procedure above; and a rollback that is two
`systemctl` verbs because the library is untouched by construction.

## Process

Each pass gets its own plan via writing-plans, drafted just in time in
the shadow of the previous pass's execution and never front-loaded.
Plans carry a token ceiling and a checkpoint interval. Execution runs
the implementer, `diff-reviewer`, gate chain per task, with workflow
mode at six or more tasks, which 4a will be. The gate is `bash
scripts/check.sh`, green at every task end, never commit red, specific
files only. Both budgets are scored at every pass close.

`go-conventions` before any Go. `vps-conventions` before the units,
`deploy.sh`, or anything that places a file or directory on the box.
Subagents cannot dispatch agents, so the conductor owns the
`code-simplifier` dispatch at each pass close.

**Review scales to rung risk.** Rung 4 is not one of the three
full-adversarial-cycle rungs; auth (5), streamer and SSO (12), and
cutover (16) keep that (`PA:120`, `PA:288-291`). Rung 4 rides the
in-chain `diff-reviewer` per task plus the pass-end gate. Two
exceptions, both narrow and both about the box:

- **The unit files and the deploy path get a fresh-context read before
  any box write.** One agent, zero context, reads
  `dubplate-shadow.service`, `dubplate-shadow.timer`, and every
  `deploy.sh` diff against `~/.claude/skills/vps-conventions/SKILL.md`
  and against `dubplate-serve.service` as the hardening template, and
  reports findings. `systemd-analyze verify` in the gate
  (`scripts/check.sh:125-148`) proves the units parse; it proves
  nothing about whether the `ReadWritePaths` grant is the minimum, and
  the gate rule is that a unit's write targets are argued statically
  rather than assumed (`SKILL.md`, Gate).
- **The observation mechanism's own honesty gets one mutation proof per
  checker**, on the standing assertion discipline. A shadow that
  silently observes nothing is this rung's equivalent of a conformance
  runner that grades nothing, and it would cost thirty days before
  anyone noticed. 4a's exit gate is written above as an end-to-end run
  producing one agreement and one divergence of each class, so the
  instrument is proven able to fire before it is deployed.

**Owner touchpoints rung 4 expects.** The six rulings above, batched
at one checkpoint before 4a's plan is drafted. The deploy sitting
(ruling 6), short. And the standing shadow adjudication stream from the
deploy onward, which `BS:896-903` already priced and which rung 4
starts rather than adds. One inherited ask rides that stream and is not
rung 4's to raise: 3d's request to re-disposition
`bug-assignment-rejection-gap` from `oracle-wrong` to
`no-correct-answer` (`docs/HISTORY.md`, Resolution R1).

---
