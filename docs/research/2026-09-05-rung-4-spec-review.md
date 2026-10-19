# Adversarial review: the rung 4 shadow spec

Two fresh-context reviews (2026-09-05) of
`specs/2026-09-05-rung-4-shadow.md` at revision 1, one through the
estate-safety-and-operations lens and one through the
evidence-integrity-and-engine-seam lens, merged here and folded as
revision 2 of the spec the same day. The merged "Verified sound" list
below is the do-not-re-litigate record for every rung 4 pass plan.

---

Findings are renumbered across both reviews. A finding both lenses
raised carries the stronger evidence and names both readings. Every
finding was verified against the tree at `440131e3` before folding.
One half of one finding is rejected; its evidence is directly below.

## Rejected, with the evidence

**The "two templates cannot render" framing of B5's first half.**
Review 1 reads `DUBPLATE_STATE_DIR` as needing two values from one flat
shell namespace and concludes the render aborts or silently grants the
shadow write over `/var/lib/musicbox`. The renderer half is right
(`scripts/deploy.sh:181-182` resolves `${!name}` in one process-wide
namespace, `:190-194` exits 1 on any miss), but the spec's own deploy
edit 5 is the mechanism that resolves it: "a template line carrying a
value on its right-hand side is a literal", and the spec already places
`DUBPLATE_STATE_DIR` among the non-secret paths that need no mint. So
the design can render. What is actually wrong is that the spec never
says which greenfield rows are literals, and the default reading of a
bare `DUBPLATE_STATE_DIR=` line is the incumbent's exported
`/var/lib/musicbox`. The failure is silent rather than loud. B5 folds
that half as a required explicit statement, not as an unrenderable
design.

---

## BLOCKERS

**B1. The shadow's album unit is not the oracle's, because the box runs
`group_albums: yes`.** (Both lenses.)
The capture walks `<observe dir>/<uploader>/<album>/` and compares one
directory to one oracle verdict. The box's beets sets `group_albums:
yes` (`config/beets-config.yaml:23`), and its own comment records why,
from this estate's real material: contributor ZIPs extract flat, "so
one staging dir holds many albums. Without this, beets saw the first
seed's nine albums as one 135-track candidate and skipped them all
(2026-08-31)" (`config/beets-config.yaml:19-22`). The incumbent runs
one `beet import` over the whole staging tree
(`_salvage/internal/importer/importer.go:1100`), so beets groups by
`(albumartist, album)` across it and is free to split one captured
directory into several verdicts or to span two. The corpus deliberately
sits below this step and says so (`corpus/README.md`,
`import.group_albums` dropped because the corpus tests the matcher from
the items-and-responses boundary, below the importer's grouping step),
so `tag.Album` inherited no grouping: it takes `items
[]matcher.ItemFields` and calls `matcher.MostCommonTags(items)` over
all of them (`internal/engine/tag/tag.go:68`, `:75`). An implementer
building this as written drives one proposal against N oracle outcomes
and every mismatch lands in the launch-bar record as a divergence
(`BS:69-71`). Risk 1's own mitigation, a deliberate backlog feed, makes
multi-album directories more likely rather than less.
*Fix:* state the grouping rule as a first-class part of the mechanism.
Either reproduce beets' grouping in the capture, naming the source it
is ported from, or observe only directories that group to exactly one
album and count and name every excluded multi-album directory in the
report, on the gate-rejected exclusion's own precedent. Add
`group_albums` to the staleness pre-task.

**B2. `AlbumOutcome.SourceDir` cannot identify an album, so the
reconcile join is ambiguous by construction.** (Both lenses.)
The reconcile step joins "each `AlbumOutcome` whose `SourceDir` matches
a captured fixture". The spec reads `importer.go:1066` correctly and
then joins on a key that cannot join. `albumUnit.sourceDir` is
documented as "the uploader's inbox directory"
(`_salvage/internal/importer/importer.go:212`) and all three
assignments set it at the uploader level, never the album: `:766`,
`:908`, `:1066`. `AlbumOutcome` carries no per-album path at all
(`:134-150`); the album name lives only in `albumUnit.stagedDir`
(`:213`), which `toReport` does not copy (`:231-240`).
`stageAllContributors` stages every quiescent album of every
contributor in one run (`:782-800`), so a run report routinely carries
several rows with identical `SourceDir`, and the join maps every
fixture to every outcome. `SessionID` does not rescue it: it is empty
for a leftover staged before the uploader shipped or a synthetic orphan
(`:139-142`), and the wind-down retired every intake path, so this
material carries no sidecar. The apply half is worse than ambiguous:
`LibraryPaths` is filled by a **positional** match, staging order
against the beets database's album order, because "one shared beet
invocation carries no per-uploader join key" (`:1119-1127`), and a
count mismatch leaves the extra units empty. For the skip half, the
half ruling 3 says most needs adjudication, the report gives no album
identity whatsoever.
*Fix:* the join key must be the album basename, with `SourceDir` as an
uploader-level prefilter. The material exists: the tag log's `skip `
lines carry staged paths (`_salvage/internal/beets/beets.go:116-135`),
the review tree files under `<review>/<date>/<uploader>/<basename>`,
and `LibraryPaths` names the applied files, readable back through the
shadow's own FLAC reader. State the tie-break, and record every album
the join cannot resolve as `unreconciled`.

**B3. Ruling 4's option (c) cites the passage that forbids it, and its
second mode cannot run.** (Evidence lens.)
The spec says the two-mode scoring "mirrors `conform`'s own two-mode
sweep (`internal/harness/conform/case.go:141-160`)" and that the second
mode replays stored responses to drive a second scoring pass. That
passage states the opposite: "ModeDefault does not drive tag.Album. It
re-scores each case's own committed candidate set under the corrected
table, and the reason is that a corrected verdict changes the FLOW"
(`case.go:141-153`), and `gradeCaseDefault` replays only the lookups
the committed candidates name and issues no search (`case.go:517-526`,
`:591-593`). The mechanism is the ID-consensus early return, which
fires only at `rec == matcher.RecStrong` (`tag.go:96-118`, condition at
`:106`); `matcher.Compat.Apply` swaps `StringDistPatterns`
(`internal/engine/matcher/compat.go:51-60`), which moves the distance
and therefore the recommendation, so an album that takes the one-lookup
early return under `FullCompat` falls through to `src.SearchReleases`
(`tag.go:125-126`) under `DefaultCompat`, a request no recording ever
captured. The budget claim is independently wrong: the budget sits
above the transport, not below it. `doWithRetry` calls
`c.budget.Wait(ctx)` before every attempt (`client.go:346`) whatever
`ClientConfig.HTTP` holds (`client.go:98`), so reusing the client for a
replay doubles the `ClassBackground` spend and doubles `Budget.Spent()`
(`limiter.go:400-404`), the figure the report prints as honest.
*Fix:* score the second mode the way `gradeCaseDefault` does, re-scoring
the FullCompat run's own recorded candidate set under
`DefaultCompat().Apply(tun)` with no search and no second `tag.Album`;
give any replay client `musicbrainz.Unmetered()` (`limiter.go:354`), as
`gradeClient` does (`case.go:330-336`); and state the limit of the
derived figure, since a re-scored verdict is a hypothesis wherever the
real flow would have widened the candidate set.

**B4. Three deliverables put `internal/engine/shadow` on the wrong side
of the harness-import wall.** (Evidence lens.)
Deliverables 3 and 4 name a "case-shaped fixture" and "pending rows in
`analysis.AdjudicationEntry`'s exact shape"; the store layout writes
`case.json` and `pending/<obs-id>.json`.
`TestEngineNeverImportsHarness` walks the whole `internal/engine` tree
and fails any file importing anything under `internal/harness`
(`internal/harness/conform/arch_test.go:141-146`,
`checkNoHarnessImports` at `:55-65`), so a package at
`internal/engine/shadow` may import neither `internal/harness/cassette`
nor `internal/harness/analysis`. The spec proves it knows the wall
exists, invoking this same test to justify landing
`matcher.QuietImportTunables` on the engine side, and then does not
apply it to the two artifacts the whole rung produces. An implementer
either reds the gate or hand-rolls duplicate structs, reintroducing the
silent-drift failure this repo has already named
(`internal/engine/matcher/info.go:148-150`) at the seam where drift
makes a rung 15 comparison wrong. A third consequence:
`checkAdjudicationEntry` is unexported (`adjudication.go:158`), so the
shadow cannot validate a row before staging it and a malformed row is
discovered weeks later.
*Fix:* keep the artifact writing where both sides are reachable.
`cmd/dubplate-server` imports both freely, and the spec already routes
`now` and the User-Agent through it. `internal/engine/shadow` returns
typed engine-side results; the subcommand marshals them into
`cassette.Case` and `analysis.AdjudicationEntry` and runs the exported
`analysis.CheckAdjudications` (`adjudication.go:138`) over each staged
row before it is written.

**B5. The greenfield template's literal rows are unstated, and two rows
it needs are not in the age store.** (Estate lens.)
`scripts/deploy.sh` resolves every template line through one flat shell
namespace (`:181-182`) and exits 1 on any miss before any remote step
(`:190-194`). Two consequences:

- **The greenfield `DUBPLATE_STATE_DIR` must be a stated literal.** The
  incumbent's template carries a bare `DUBPLATE_STATE_DIR=`
  (`env/dubplate.env.template:74`) whose value is deliberately
  `/var/lib/musicbox` (`:59`). If the greenfield file inherits that
  namespace, the shadow unit's `ReadWritePaths` either grants write over
  the incumbent's lock, tag logs, and run reports, the exact grant the
  read-only guarantee exists to withhold, or the unit fails to start.
  Deploy edit 5 is the mechanism, but the spec never says which rows use
  it.
- **`DUBPLATE_SITE_NAME` and `DUBPLATE_LIBRARY_ROOT` are not minted.**
  The spec states "**Secrets: none** … `DUBPLATE_SITE_NAME` stays minted
  in the age store from the wind-down." A name-only check says
  otherwise: the store carries `DUBPLATE_SERVE_SITE_NAME` and
  `DUBPLATE_LIBRARY_DIR`, the pre-rename names, and
  `~/.dotfiles/secrets/registry.md` names only `DUBPLATE_SERVE_SITE_NAME`
  (minted 2026-09-01, recorded as non-secret masthead text).
  `internal/config/keys.go:165-166` renamed both. A greenfield template
  carrying `DUBPLATE_SITE_NAME=` and `DUBPLATE_LIBRARY_ROOT=` aborts
  4b's deploy at the render, before ssh.

*Fix:* name the literal rows explicitly in the greenfield template,
`DUBPLATE_STATE_DIR=/var/lib/dubplate` and
`DUBPLATE_LIBRARY_ROOT=/srv/music/library` among them; decide
`DUBPLATE_SITE_NAME` in the open, literal or minted, and say which; and
delete the "the registry gains no row" sentence or restate it against
the store.

**B6. The stamp guard's only surfacing mechanism does not exist.**
(Both lenses; blocker on the estate lens, major on the evidence lens.)
The spec pairs a hard refusal, "`shadow observe` errors rather than
observing when the stamp is empty or `dev`", with a claimed early
warning: "`deploy.sh` sets the stamp on the cross-compile, and the
existing `dubplate-server version` smoke test (`scripts/deploy.sh:288-289`)
is where a missing stamp surfaces before any timer fires." Neither half
holds. `build-linux` carries no `-ldflags` at all (`Makefile:6-7`), so
`main.version` is `"dev"` on the box today and there is nothing for the
matcher stamp to join. `newVersionCmd` prints `version` and nothing
else (`cmd/dubplate-server/main.go:56-64`), so the smoke test cannot
assert on a stamp under any implementation. The combined failure is the
one this rung names as its worst: an unattended shadow observing
nothing, firing `OnFailure=dubplate-ping-fail@shadow.service` into a
slug whose URL is unset and therefore no-ops
(`_salvage/cmd/dubplate/pingfail.go:41`), so there is silence on both channels for
thirty days, discovered at rung 15.
*Fix:* make deploy edit 6 an actual change: `build-linux` gains
`-ldflags "-X main.version=… -X main.matcherCommit=…"` with the source
of each named, `dubplate-server version` prints both, and the smoke
test asserts a non-`dev` matcher stamp. Add it to the staleness
pre-task.

## MAJORS

**M1. `ErrBudgetExhausted` is unclassified, and the enumeration reads as
complete.** (Both lenses.) The divergence table enumerates four
outcomes and then closes the space with "Two branches are neither",
naming `ErrUnavailable` and `ErrNoCandidates`.
`musicbrainz.ErrBudgetExhausted` is a third sentinel and exists exactly
so a caller branches: "A caller branches on it to stop asking rather
than to retry, which is why it is a sentinel" (`limiter.go:45-48`). It
surfaces through `budget.Wait` (`client.go:346`) and returns wrapped
out of `tag.Album`. At ruling 6's own arithmetic sixty tokens drain at
eight or nine albums, so exhaustion mid-run is the ordinary path on any
backlog, including the first pass after deploy. With no branch, the
leftover error falls to the table's only "Go produced nothing" row and
writes a fabricated `go-skips-oracle-applied` divergence against an
oracle apply, consuming the owner adjudication ruling B's stamp exists
to save. `ErrLimiterClosed` (`limiter.go:50-53`) has the same exposure
at shutdown.
*Fix:* add both as deferring branches beside `ErrUnavailable`, have the
report's `deferred` row count by reason, and state that no error other
than `ErrNoCandidates` may ever be recorded as a Go skip.

**M2. The unit names no `[Unit]` ordering and no `EnvironmentFile=`.**
(Estate lens.) The unit section specifies `Type`, `User`, `ExecStart`,
the hardening set, `RestrictAddressFamilies`, `ReadWritePaths`, and
`OnFailure`, and stops. Every network-touching oneshot on this box
pairs `After=network-online.target` with `Wants=network-online.target`
(`systemd/dubplate-import.service:9-10`,
`systemd/dubplate-backup.service:13-14`), and the serve unit documents
at length why it deliberately does not
(`systemd/dubplate-serve.service:12-27`): the estate treats this as a
decision, never an omission. A `*:0/5` timer fires before the network
is up after a reboot. Separately, every sibling unit carries
`EnvironmentFile=/etc/dubplate/env` (`dubplate-import.service:14`), and
the shadow's whole environment-file section exists to feed a line the
unit never names.
*Fix:* specify the full head, `After=`/`Wants=network-online.target`
and `EnvironmentFile=/etc/dubplate/server.env`, with the network
pairing's reason in the unit's own comment.

**M3. The reconcile step's primary artifact is pruned on a rolling
window the spec never reads.** (Estate lens.) Mechanism E reads the
incumbent's run reports. Those are swept on every import pass,
`im.sweepReports(im.Cfg.Thresholds.ReportRetention)`
(`_salvage/internal/importer/importer.go:389`), which "keeps at most
retention of the newest `report-*.json`" (`:533-548`), with
`defaultReportRetention = 30`
(`_salvage/internal/config/config.go:172`). A `RunReport` is written on
every exit path including a held lock (`importer.go:152-156`) and the
timer fires `*:0/15`, so thirty reports is roughly seven and a half
hours of history. Backstop H is swept too, at `RetentionDays`
(`importer.go:386`), defaulting to 60 (`config.go:155`). Only D, the
tag logs, is never swept, and it records skips only. Against that the
spec sets `Persistent=false` with "a missed pass costs at most a later
capture". A shadow outage past seven and a half hours costs the oracle
signal permanently for every album imported inside it, during the
unattended window whose whole product is that record.
*Fix:* state the retention facts with their citations, and either
persist consumed-report state and name a report-generation gap as a
reported condition or set `Persistent=true` and rewrite the
justification. Add both retention defaults to the staleness pre-task.

**M4. The capture can read an album mid-move, and the supersede rule
does not cover it.** (Estate lens.) "A capture that lands mid-write
produces a fixture with a digest the next pass supersedes" holds at the
upload boundary and not at the other one: once the importer moves the
album out of the inbox there is no next pass to supersede anything.
`review.Move` tries `os.Rename` first (`_salvage/internal/review/move.go:24-40`),
atomic per top-level entry on one filesystem, and falls back to a
recursive copy-then-remove across an EXDEV boundary. Either way, a
shadow walk that lists a directory and then opens its files one at a
time can read three tracks and take ENOENT on the fourth. The spec
names no rule, so the stored fixture is a truncated album scored as
complete.
*Fix:* one rule, stated: a capture that cannot read every entry it
listed is discarded, never stored, and the pass moves on. That also
makes the supersede claim true as written.

**M5. The read-only guarantee is unit-scoped; two of the three verbs are
designed to be hand-run.** (Estate lens.) The namespace argument is
true of `ExecStart` and only of `ExecStart`. `shadow report` and
`shadow reevaluate` have no timer, risk 3 contemplates a hand-run
`observe`, and the deploy sitting is a hand-run session over ssh where
`deploy.sh` connects as `root@$HOST` (`scripts/deploy.sh:110`). A verb
run by hand carries no `ProtectSystem`, no `ReadWritePaths`, and full
write access to the library and the incumbent's state.
*Fix:* scope the claim to the timer path and add the operator rule
beside it: a hand-run verb goes through `systemctl start
dubplate-shadow.service` or `systemd-run` with the same properties,
never the bare binary as root. Put it in the quiet-box procedure and in
the unit's comment.

**M6. The MusicBrainz rate limit is per-IP and shared with the box's own
beets imports.** (Estate lens.) Ruling 6 reasons entirely inside the
shadow's own limiter, which paces "every MusicBrainz request one
process issues" (`limiter.go:64-65`). The box's other MusicBrainz
client is `beet import`, which pins its own `rate_limit = 1.0`
independently (recorded at `limiter.go:55-61`), fires `*:0/15`
(`systemd/dubplate-import.timer:10`), and shares the box's single
egress IP. A `*:0/5` shadow will regularly run concurrently with an
import, and the two together exceed one request per second, which would
degrade the production path this rung promises to leave alone.
*Fix:* name the shared per-IP budget as a constraint and choose a
defensible mitigation: skip a pass while an import is plausibly in
flight, readable without touching the lock, or halve the shadow's pace
and re-derive the cap.

**M7. The rollback offers an action that resets the launch clock.**
(Estate lens.) "`systemctl disable --now …`, and optionally remove
`/var/lib/dubplate`." The store holds `started-at`, written once and
never rewritten, and launch date equals rung 4's deploy date plus the
floor (`DA:524-527`). Removing the directory discards every
observation, every stored response the re-evaluator depends on, and the
clock start, offered as an optional tidy-up. The two halves are not
symmetric: disabling the timer is free and reversible; removing the
store is neither.
*Fix:* split the rollback into "stop observing" and "discard the
evidence", and say in those words that the second restarts the floor
and needs the owner's explicit decision.

**M8. The report's budget row cannot be printed by `shadow report`.**
(Evidence lens.) A `Budget` is in-process state: `spent` is an int
guarded by a mutex on the struct (`limiter.go:393-404`), created per
call to `Limiter.Budget` (`limiter.go:139-145`), and nothing persists
it. `observe` and `report` are separate process invocations, so a
`report` run builds no limiter and prints a zero that reads as a
measurement of an unattended thirty-day window.
*Fix:* each `observe` pass writes its `Budget.Name()` and
`Budget.Spent()` into its own pass record under the store, and `report`
sums them, per-run and cumulative separately. The store layout gains
that record.

**M9. The report cannot evaluate half of launch bar part 3.** (Evidence
lens.) The gating clause is named as the pending count alone. Launch
bar part 3 has two clauses: "zero unadjudicated divergences and zero
false confident matches against live beets decisions" (`BS:69-71`). The
spec names `go-applies-oracle-skipped` as "the class a false confident
match hides in" and never reports the count. It is derivable from the
disposition breakdown, since such a row whose disposition lands at
`OracleCorrect` (`adjudication.go:20`) is a false confident match and
one at `OracleWrong` is not, but nothing says so, and ruling 2's
tier-one refinement, which makes this a release-truth question rather
than an oracle question (`BS:64-68`), is never cited.
*Fix:* add the row with its definition, state that it must be zero at
the launch gate beside the pending count, and cite `BS:64-71`.

**M10. The row table's `oracle`, `go`, and `class` fields do not exist
and are not added.** (Evidence lens.) The table promises one shape
"written into the observation record and lifted verbatim into the
pending adjudication artifact", and the schema section adds only
`MatcherCommit` and `MatcherConfigHash`.
`analysis.AdjudicationEntry` has no `oracle`, `go`, or `class` field
(`adjudication.go:46-67`), and `AdjudicationSubject` has three
(`:32-39`). "Lifted verbatim" therefore contradicts the two-field
addition: an implementer drops the three fields, losing every
mechanical fact a rung 15 re-evaluation and M9's count need, or adds
three unspecified fields to the shared corpus type with no rules
written for them.
*Fix:* split the table in two, the observation record's own shape
carrying `oracle`/`go`/`class` and the `AdjudicationEntry` subset
staged under `pending/`, and state how the observation's facts render
into the prose fields the checker requires. If rung 15 must read
`class` mechanically off a committed row, add it as an `omitempty`
block with its own check on the `DefectFree` precedent
(`adjudication.go:58-67`).

**M11. Ruling 2's option (b) is not open; ratified ruling B forecloses
it.** (Evidence lens.) Option (b), storing the verdict and item fields
and discarding the responses, is presented as a live option and treated
as a live branch in risk 5. Ruling B is ratified as "rung 4 deploys at
functional completeness, with every divergence stamped with the matcher
commit so divergences from superseded revisions are re-evaluated
mechanically instead of consuming owner adjudication"
(`specs/2026-09-03-rung-3-go-matcher.md:93`, recommendation at `:95`,
carried into the handoff at `docs/HISTORY.md:280-286`). The spec's own
analysis establishes that mechanical re-evaluation of a live
observation is only possible against the responses the observation saw.
So (b) retires a ratified requirement, silently, in a list whose stated
rule is "Nothing the record already settles appears here".
*Fix:* retire ruling 2, state the response store as a consequence of
ruling B, and keep the disk sizing as risk 5.

## MINORS

- **m1.** The identity-source pending row is `internal/config/keys.go:24`,
  not `:23` (line 23 is `DUBPLATE_ADMIN_EMAIL`). `reservedClasses` is
  `:57-142`, not `:57-140`.
- **m2.** The `internal/engine/tag/tag.go` citation cluster is wrong
  throughout, though every claim it carries is true. Correct anchors:
  `Album`'s signature `:68-69` (cited `:83-84`, twice); `Proposal`
  `:23-29` (cited `:38-44`); the `opts.SearchIDs` branch `:78-85`
  (cited `:93-101`); the ID-consensus early return `:96-118` with its
  condition at `:106` (cited `:117-133`); the search path `:121-134`
  (cited `:140-149`); `ErrNoCandidates` declared `:18` and returned
  `:139` (cited `:78-82`).
- **m3.** "the five prose fields `checkAdjudicationEntry` requires
  non-empty" names `citation`, which is not checked, and omits
  `adjudicator`, which is: `adjudication.go:176-189` checks
  `Suspected`, `ReleaseTruth`, `Reason`, `Adjudicator`, `Date`.
  `docs/HISTORY.md:295-302` carries the same slip and the spec inherits
  it.
- **m4.** The "keep the data path name" ruling is cited as
  `docs/HISTORY.md:2211-2213`, which is the compose project-name pin.
  The kept-identifiers record is `docs/HISTORY.md:2218-2220`. The
  argument survives; the citation does not.
- **m5.** Risk 5 sizes the store "against a 50 GB volume
  (`scripts/provision.sh:14-24`)". That volume is `musicbox-data`
  (`provision.sh:16-17`), the `/srv/music` library volume.
  `/var/lib/dubplate` sits on the root filesystem, as
  `scripts/deploy.sh:367-369` states outright.
- **m6.** Ruling 7's "It does not add a second touchpoint" contradicts
  the accounting it cites: `BS:894-897` prices "one attended gate (rung
  16's single-sitting verification)" plus one visual check and the
  adjudication stream. A deploy sitting is a second attended gate. The
  recommendation is right; the justification should say so and argue
  it is worth it, which the preceding sentence already does.
- **m7.** Ruling 6's worst case is understated. `MergePseudoRelease`'s
  second fetch is per candidate
  (`internal/engine/musicbrainz/client.go:320-333`), so a search path
  returning `SearchLimit` = 5 ids
  (`internal/engine/matcher/tunables.go:309`) can cost 1 + 5 + 5 = 11
  requests, and `maxAttempts = 7` (`client.go:42`) multiplies a
  degraded run. Sixty tokens still covers five albums; state the real
  bound.
- **m8.** The apply half assumes `library_paths` is populated.
  `AlbumOutcome.LibraryPaths` is filled only for `OutcomeImported`
  (`_salvage/internal/importer/importer.go:127-133`) by a positional
  match that leaves extras empty (`:1119-1127`), and `Error` explicitly
  carries "a beets-database match that never turned up" (`:145-149`) on
  an outcome that still reads `imported`. Name it as an `unreconciled`
  reason.
- **m9.** The ping no-op claim cites
  `systemd/dubplate-ping-fail@.service:15-19`, which documents the
  config-load bypass. The no-op is `_salvage/cmd/dubplate/pingfail.go:41`,
  with the estate precedent at `systemd/dubplate-serve.service:29-34`.
  Note also that the helper reads `EnvironmentFile=/etc/dubplate/env`
  (`:11`), the incumbent's file, so the `OnFailure=` line is wired and
  inert for the whole shadow window.
- **m10.** The `RE:` document key is used at four load-bearing claims
  (`:734`, `:1023`, `:1105`, `:1160`) and never defined in the header's
  key list. It resolves to
  `research/2026-09-04-reverse-edge-completion.md`, where all four
  citations are correct.
- **m11.** "Rung 4's deploy is merge point M1" cites `DA:328-331`,
  which are the Wave-A lane table rows. The M1 heading is `DA:335`.
- **m12.** "the estate's beets config sets no `match:` and no
  `musicbrainz:` block at all" cites `BS:240-241`, which says nothing
  of the kind. The evidence is `config/beets-config.yaml` itself and
  `internal/harness/conform/defaults.go:16-18`.
- **m13.** The doctor's rows "come from `internal/config`'s own tables"
  is true for the wall and the reserved classes, which are
  package-level vars, and not for thresholds: `config.Thresholds()`
  returns only what someone passed to `config.Register`, called "from
  an explicit setup call rather than an init function" by design
  (`internal/config/thresholds.go:19-30`). Rung 4's two thresholds
  appear in `doctor` only if the spec names a call site.
- **m14.** The contested-surfaces table omits the state-directory
  layout, which `RE:102` assigns to rung 4 by name, and the
  `corpus/analysis/adjudications.json` divergence-schema row, which
  `RE:198` assigns to rung 4 as well.
- **m15.** Deploy edit 5's literal column extends a row-state
  vocabulary `RE:100` already reasoned through for this file: "a name
  whose value arrives later lands as a commented pending row on the
  `config.WallRow.Pending` precedent, never as a live `NAME=` line".
  The literal is a third state covering a different case, and it is
  backward compatible (no tracked template line carries a right-hand
  value, and the loop reads only the name, `deploy.sh:181-182`). Say
  that it extends the convention and why non-secret paths warrant a
  third state.

## Verified sound, do not re-litigate

Merged from both reviews; every item was checked against the tree.

1. **The ladder entry is quoted verbatim and correctly**
   (`BS:572-580`), as are the shadow constraints (`:262-268`), the
   coarse-oracle framing (`:255-260`), ruling 1's floor (`:804-815`),
   the observed-count-reported-not-gated rule (`:249-253`, `:812-815`),
   launch-bar part 3 (`:69-71`), box safety (`:864-874`), the
   estate-operational-record boundary (`:283-288`), the packaging
   boundary (`:882-884`), universality's four tests (`:116-120`), U4's
   required-when-enabled shape (`:150-158`), and the
   shadow-as-first-class-leaf argument (`:335-337`).
2. **The hazard that decides the mechanism is exactly right.**
   `config/beets-config.yaml:14-15` sets `move: yes` and `write: yes`,
   so a post-import read recovers the oracle's answer and not the
   question. This is the spec's central insight and it holds.
3. **The oracle's binary signal is verified in pinned beets.**
   `_summary_judgment` returns `APPLY` when and only when the
   recommendation is strong, otherwise resolving `quiet_fallback`
   (`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:335-341`),
   mirrored by `matcher.ActionFor`
   (`internal/engine/matcher/verdict.go:179-192`). `RecMedium`,
   `RecLow` and `RecNone` are genuinely indistinguishable to the
   oracle, so the binary projection of `Proposal.Action` is the only
   comparable quantity.
4. **Declining mechanism C (the beets database) is correct and
   well-argued.** `DUBPLATE_BEETS_DB` is retired with "the greenfield
   server never embeds beets" (`internal/config/keys.go:179`), and
   `mb_albumid` is reachable off the applied file:
   `MUSICBRAINZ_ALBUMID` is mediafile's Vorbis style
   (`mediafile/__init__.py:782`, `CATALOGNUMBER` at `:604`), and
   `MBAlbumID` is one of the nineteen `ItemFields`
   (`internal/engine/matcher/info.go:130`). "O5 is closed by not
   asking it" stands.
5. **Mechanism G is rightly declined.** `dubplate-import.service:54`
   grants `ReadWritePaths=/srv/music /var/lib/musicbox`, so an
   `ExecStartPost=` would inherit the incumbent's write grants and
   would edit a frozen unit against the additive-deploy boundary.
6. **The hardening set is carried correctly.** Every directive named is
   present at `systemd/dubplate-serve.service:99-115`, and the
   `AF_NETLINK` reasoning is right: the family joined for a CPython
   subprocess's `getaddrinfo`, and "The Go process itself never needed
   it" (`:116-125`). `CGO_ENABLED=0` in `build-linux` (`Makefile:7`)
   forces the pure resolver, so denying the family is safe.
7. **Every `deploy.sh` edit lands on a real line.** `UNIT_FILES`
   (`:56-64`), `TIMER_UNITS` and the hand-enable precedent (`:407`,
   `:49-55`, `:398-405`), the pre-create shape (`:357-359`) and its
   namespacing rationale (`:350-356`), the render loop (`:175-194`) and
   install block (`:291-316`) both hard-wired to `REMOTE_ENV` (`:28`),
   and the binary step with its rung-4 boundary comment (`:251-260`,
   `:273-289`). "The path already exists" is correct.
8. **The `vps-conventions` citations are all exact:** ReadWritePaths
   pre-creation (`SKILL.md:39-42`), `var_lib_t` verification (`:43-45`),
   path-unit coalescing and the timer backstop (`:47-55`), rsync exit 23
   forcing manifest edits into the same commit (`:70-74`), and the Gate
   section's "argued statically is acceptable, assumed is not"
   (`:76-80`), which is why the fresh-context unit read is the right
   narrow exception.
9. **The gate already covers both new files.** `scripts/check.sh:137`
   globs `systemd/*.service systemd/*.timer systemd/*.path` and the
   filter at `:139` matches `/usr/local/bin/[A-Za-z0-9_.-]+`. No new
   wiring needed.
10. **The environment-file diagnosis is right.** The template carries
    exactly 39 `DUBPLATE_*` names; `DUBPLATE_ADMIN_EMAIL` (`:25`) and
    `DUBPLATE_STATE_DIR` (`:74`) are the two owned rows; every other one
    hits `rejectNamespace` as reserved, superseded, or unknown
    (`internal/config/config.go:133-152`), so a greenfield unit reading
    `/etc/dubplate/env` fails at load every time. The header at `:1-3`
    anticipates this rung by name.
11. **The `import.lock` trap is real and correctly diagnosed.**
    `acquireOrRecordHeld` treats `lock.ErrHeld` as healthy: it writes a
    `LockHeld` report, pings success, and returns nil
    (`_salvage/internal/importer/importer.go:397-413`). A shadow holding
    that lock would stall imports while every monitor read green.
    Naming it a code rule rather than a sandbox grant is right.
12. **The quiescence constraint is sound and the arithmetic works.**
    Ten minutes by default (`_salvage/internal/config/config.go:152`,
    applied at `importer.go:790-798`), import timer `*:0/15` with 300s
    jitter (`systemd/dubplate-import.timer:10-11`), against a
    six-minute worst-case shadow gap. At least one quiescent capture
    always falls between the last write and the earliest possible
    import. The `RunSession` path that bypasses quiescence
    (`importer.go:281-284`) is covered by the pre-task's disabled-nudge
    check.
13. **The `analysis` schema work is correct.** `AdjudicationEntry`
    (`adjudication.go:46-67`), the shadow subject shape modeled and
    carrying zero committed rows (verified:
    `corpus/analysis/adjudications.json` holds 21 entries, 0 with
    `subject.album`, 0 pending), the pair rule (`:168-171`), owner-only
    disposition via the `AgentAdjudicator` sentinel (`:198-200`), and
    `Pending` legal at `check.sh` but refused at the launch gate
    (`:10-12`). Both new fields as `omitempty` leave the 21 rows
    byte-identical. `BySubjectCase` already omits shadow rows
    (`:107-122`).
14. **The runner's construction constraints check out.**
    `TestEngineNeverImportsHarness` (`arch_test.go:141`) and
    `TestEngineNeverReadsTheClock` (`:152`) both walk the whole
    `internal/engine` tree and descend into nested packages
    (`:189-205`), so `shadow.Run` taking `now` is required, not
    stylistic; `gradeTunables` really is on the harness side
    (`defaults.go:72-79`); and `matcher.QuietImportTunables` adds no new
    concept.
15. **The four estate assignments are right.** `DefaultTunables()` ships
    `QuietFallback: "skip"`, `NoneRecAction: "ask"`, `Timid: false`,
    `Quiet: false` (`tunables.go:312-317`), so `Quiet = true` and
    `NoneRecAction = "skip"` are the two that move, and the four-assignment
    set is equivalent to `gradeTunables("estate")`. `original_date` and
    `per_disc_numbering` are provably verdict-inert
    (`internal/harness/conform/translate.go:239-240`).
16. **Every `musicbrainz` citation is exact:** `ErrUnavailable` and its
    rung-4 branch (`client.go:74-84`), the `UserAgent` doc quoted
    verbatim (`:86-92`), the `HTTP` seam (`:98`), `NewClient` refusing
    rather than defaulting (`:121-130`), the pseudo-release second fetch
    (`:320-333`), retries spending a token by design (`:343-347`), and
    on the limiter: `ClassBackground` naming this caller
    (`limiter.go:16-18`, `:25`), `reservationInterval` as the one knob
    (`:43`), `DefaultInterval` (`:62`), `NewLimiter` (`:110`), `Budget`
    (`:139`), `Spent`/`Name` (`:400-408`).
17. **The pseudo-release identity holds.** `MergePseudoRelease` keeps the
    pseudo release's own `AlbumID`
    (`internal/engine/musicbrainz/pseudo.go:49-58`), matching what beets
    writes to `mb_albumid`, so a pseudo release does not manufacture a
    `go-applies-different-release` divergence.
18. **The 3d handoff is read correctly, item for item**
    (`docs/HISTORY.md:271-303`), including the unmoved `AdminEmail`
    User-Agent carry (`:308-310`), the float-threshold gap precedent
    (`:679-685`), and the round-robin that replaced the rule which
    "starved `ClassBatch` outright" (`:742-748`).
19. **The `.gitleaksignore` fragility is exactly stated.**
    `.gitleaksignore:1` is `internal/config/keys_test.go:generic-api-key:66`
    and line 66 is the `DUBPLATE_R2_BUCKET` row. The
    add-below-66-or-re-fingerprint rule is right and belongs in the
    pre-task.
20. **The threshold registry gap is real.** `Threshold` is
    `{Key, Default, Doc}` with `ParseInt` the only parser
    (`internal/config/thresholds.go:13-59`), and `thresholdStatus` calls
    it unconditionally (`cmd/dubplate-server/doctor.go:97-107`), so a
    duration would render `invalid:` today. Adding `Kind` plus
    `ParseDuration` is the minimal fix, and leaving `KindFloat` to rung
    3's follow-on is right.
21. **The subcommand shape matches the landed skeleton:** the
    `(environ, stdout, stderr)` constructor (`check.go:16`,
    `doctor.go:23`), `SilenceUsage`/`SilenceErrors` with main as sole
    stderr writer (`main.go:33-47`), the single `AddCommand` block
    (`:48-52`), the `shadow` verb reserved for rung 4 (`main.go:11`),
    and doctor's `"%-32s %s"` row form (`doctor.go:62-67`).
22. **`atomicfile` is a sound re-landing candidate:** 42 lines,
    temp-in-target-directory then rename
    (`_salvage/internal/atomicfile/atomicfile.go:11-20`), already the
    incumbent's own tag-log mechanism
    (`_salvage/internal/beets/beets.go:104`), and compatible with a
    single-directory `ReadWritePaths` grant.
23. **"No salvage-tier FLAC reader exists" is verified.** The only FLAC
    contact is `sampleInfo` shelling `metaflac`
    (`_salvage/internal/gate/gate.go:120-140`), which reads no Vorbis
    comment, and the pre-pivot server delegated every tag question to
    `beet` (`_salvage/internal/beets/beets.go:1-3`). Option (a) is the
    right call, and naming rung 8's absorption task is the right
    containment.
24. **The gate-rejected exclusion is correct and honest.**
    `quarantined` and `rejected` are gate verdicts that never reach
    beets, so counting them as observations would inflate the record;
    printing the excluded count with its reason is the
    `CaseGradeSummary` discipline (`internal/harness/conform/case.go:109-110`)
    applied correctly.
25. **The adjudication artifact genuinely shares the runner's shape:**
    the same `analysis.AdjudicationEntry` marshalled the same way
    (`case.go:467-478`), and the `matcher_config_hash` /
    `cassette.Expected.ConfigHash` name split (`cassette.go:80`) is a
    real conflation risk correctly avoided.
26. **The pass split, the tree allocation, and the review scaling** are
    consistent with the model: rung 4 writes `internal/harness/analysis`
    and harness work never moves to a worktree, rung 4 holds the primary
    tree for both passes, rung 4 is not one of the three
    full-adversarial-cycle rungs, and `BS:864-874` names rung 4 as the
    one pre-cutover rung that writes to the box.
27. **The staleness pre-task is the right shape**, and its three
    blocking questions are the right three. Ruling B's re-evaluator
    being owned here rather than deferred is correct: `DA:507-509` and
    `DA:531-535` make the stamp worthless without it.
28. **The launch-gate and clock arithmetic.** `4 → 15` and `4 → 16` as
    both calendar and artifact gates (`DA:113-115`), `4 → 9` genuinely
    none (`DA:48`), launch date equal to deploy plus floor
    (`DA:524-527`), the floor's 30-day default as adopter-facing with
    the estate's own value a separate choice (`DA:511-520`), and the
    cell (8,4) `SOFT-B` sentence paraphrased from `RE:190` and `RE:202`
    verbatim.

## Verdict

**A fold plus one rewrite of the reconcile step; the architecture
survives.** The hard part is right: the pre-import capture is the only
mechanism that sees the question, declining the beets database
dissolves a real config-wall problem, the hardening is faithfully
carried, and the read-only-by-namespace argument is the strongest one
available on this box. What fails is the layer where the design meets
the incumbent's own bookkeeping and the harness wall. Two blockers say
the same thing from different sides, that one inbox directory is not
one album and that the run report cannot say which album it is talking
about; one says the two-mode scoring cannot run the way the passage it
cites actually works; one puts the rung's two artifacts on the wrong
side of an architectural test the spec itself invokes elsewhere; one
says the deploy would abort at the render on two renamed names; and one
says the unstamped-build guard has no way to announce itself in a rung
whose named worst outcome is thirty unattended days of nothing. None of
it touches the box and none of it re-decides anything Geoff has ruled.
