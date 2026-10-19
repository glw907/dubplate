# Rung 4, Pass 4a Planning Inventory

Research report, 2026-09-05, prepared for pass 4a's plan authoring,
against HEAD `54e534b0e83a305fe02c58cc41dd05ec7ee573a0`. Scope, quoted
from the rung 4 spec's pass-split table
(`docs/superpowers/specs/2026-09-05-rung-4-shadow.md:1616`):

> **4a** The runner | `internal/atomicfile` re-landed;
> `matcher.QuietImportTunables`; the config integration class, the
> duration threshold, the six rows and their `Register` call site; the
> FLAC reader and its mapping; the capture with its grouping and
> partial-read rules; the reconcile step's basename join and
> consumed-report ledger; the match step; the store, including
> `passes/`; the recording transport and the `DefaultCompat` re-score;
> the divergence classification with its four deferring branches;
> `analysis.AdjudicationEntry`'s two fields and their checks; the
> marshalling in `cmd/dubplate-server`; the `shadow` subcommand's three
> verbs; the report; `reevaluate` | `bash scripts/check.sh` green, plus
> one end-to-end `observe` run against a hand-authored inbox and a
> hand-authored run report in the pass's own `testdata`, producing one
> agreement, one divergence of each class, one excluded multi-album
> directory, and one deferred observation of each error branch | Main
> tree

Three sections follow: (a) the spec as the plan's contract, (b) the
engine, harness and config seams as landed, (c) the estate inputs and
the process record. A fourth section, (d), reconciles all three and
lists the open questions each raised.

---

## (a) The spec as the plan's contract

Read against the tree at HEAD `54e534b0`. Every spec citation below is
written `spec:NNN`; every tree citation is `path:line` and was opened
and checked, not carried over from another document's paraphrase.

---

### A.1 The deliverable inventory, section by section

#### A.1.1 `internal/atomicfile`, re-landed

| Item | Spec | Tree |
| --- | --- | --- |
| Package born at `internal/atomicfile`, re-landed from salvage on the merit-admission rule, never imported | `spec:759-765` | Absent today: `internal/` holds `config`, `engine`, `harness` only |
| Source material: `WriteFile(path string, data []byte, perm os.FileMode) error`, temp-in-target-directory, write, sync, close, chmod, rename | — | `_salvage/internal/atomicfile/atomicfile.go:15-41`, 42 lines total, package doc at `:1-3` |
| Every store file is written through it | `spec:759-760` | — |
| Consumers named: rungs 6, 9, 13 | `spec:764-765` | Contested-surface row written at `docs/superpowers/specs/2026-09-03-parallel-build-approach.md:181` |

The salvage version is compatible with a single-directory `ReadWritePaths` grant because the temp file is created in the target's own directory (`atomicfile.go:16-17`).

#### A.1.2 `matcher.QuietImportTunables`

| Item | Spec | Tree |
| --- | --- | --- |
| Exported named constructor in `internal/engine/matcher`, returning `DefaultTunables()` with four assignments applied | `spec:955-962` | `matcher.DefaultTunables` at `internal/engine/matcher/tunables.go:266` |
| `QuietFallback = "skip"` | `spec:934` | Already the shipped default, `tunables.go:316` |
| `NoneRecAction = "skip"` | `spec:934` | Ships `"ask"`, `tunables.go:317`. One of the two that move |
| `Timid = false` | `spec:934` | Already the shipped default, `tunables.go:312` |
| `Quiet = true` | `spec:934` | Ships `false`, `tunables.go:315`. The other one that moves |
| Reason it cannot come from the harness | `spec:950-955` | `gradeTunables` lives at `internal/harness/conform/defaults.go:72-79`, on the far side of `TestEngineNeverImportsHarness` (`internal/harness/conform/arch_test.go:141-146`) |
| Adds no new surface concept | `spec:959-962` | All four are existing `Tunables` fields |

The four assignments correspond to the box's live beets config at
`config/beets-config.yaml:16-18` plus the `-q` flag the incumbent
passes (`_salvage/internal/beets/beets.go:83-96`). The estate's beets
config sets no `match:` block and no `musicbrainz:` block, so the
whole threshold and distance surface stays at beets' shipped defaults;
the harness states the same consequence at
`internal/harness/conform/defaults.go:16-18`.

#### A.1.3 The config integration class, the duration threshold, the six rows, the `Register` call site

**The class (ruling 1, option (a), provisional, owner ruling open).**

| Item | Spec | Tree |
| --- | --- | --- |
| `IntegrationClass{Name, Rung, Flag, Keys}` beside `reservedClasses` | `spec:139-142` | `reservedClasses` at `internal/config/keys.go:57-142`, `ReservedClass` type at `:47-51` |
| `rejectNamespace` accepts a class's key when the flag is set, errors naming the flag when it is not | `spec:143-144` | `rejectNamespace` at `internal/config/config.go:133-152`; today every non-owned `DUBPLATE_*` errors as reserved, superseded, or unknown |
| `Load` requires every key of every enabled class | `spec:144-145` | `Load` at `internal/config/config.go:159-175` |
| `doctor` reports each class with its flag state | `spec:145-146` | Doctor derives its rows from `config.OwnedWall()`, `config.ReservedClasses()`, `config.Thresholds()` at `cmd/dubplate-server/doctor.go:33-51` |
| Rung 1b will absorb the table as data | `spec:1429-1432`, risk 6 at `spec:1765-1768` | Rung 1b has not landed (`docs/STATUS.md:29`) |

**The six rows** (`spec:1436-1443`):

| Key | Kind | Default |
| --- | --- | --- |
| `DUBPLATE_SHADOW_ENABLED` | integration flag | off |
| `DUBPLATE_SHADOW_OBSERVE_DIR` | required-when-enabled | none |
| `DUBPLATE_SHADOW_ORACLE_STATE_DIR` | required-when-enabled | none |
| `DUBPLATE_SHADOW_REVIEW_DIR` | required-when-enabled | none |
| `DUBPLATE_SHADOW_FLOOR` | threshold, duration | `720h` |
| `DUBPLATE_SHADOW_BUDGET` | threshold, int | `60` |

`DUBPLATE_LIBRARY_ROOT` is reused, not re-minted (`spec:1445-1447`); it
is an owned row at `internal/config/keys.go:20`. `DUBPLATE_BEETS_DB`
stays retired and unamended (`spec:1447-1448`; the retirement message
is at `keys.go:178`). Note that `DUBPLATE_REVIEW_DIR` is already
reserved for rung 13 at `keys.go:130`, so the shadow's
`DUBPLATE_SHADOW_REVIEW_DIR` is a distinct name and both will appear in
one wall.

**The threshold-kind work** (`spec:1450-1459`):

| Item | Tree |
| --- | --- |
| `Threshold` is `{Key, Default, Doc}`, `ParseInt` the only parser | `internal/config/thresholds.go:13-17`, `:43-59` |
| `thresholdStatus` calls `ParseInt` unconditionally, so a duration renders `invalid:` | `cmd/dubplate-server/doctor.go:97-107` |
| 4a adds a `Kind` field, `config.ParseDuration`, and a switch in `thresholdStatus` | — |
| `KindFloat` stays rung 3's follow-on | `docs/HISTORY.md:678-685` precedent |

`Threshold` is compared with `!=` in
`internal/config/thresholds_test.go:35` and `:45`, so `Kind` must be a
comparable type for those tests to keep compiling.

**The `Register` call site** (`spec:1078-1090`): registration happens
in `cmd/dubplate-server`, once, from the root command's construction,
because `config.Thresholds()` returns only what someone passed to
`config.Register` and rung 1 registers none
(`internal/config/thresholds.go:19-30`). `RE:101`'s "each rung adds its
own file and registers through it" is the governing award, and INV
O14's registration-seam question is answered by declining to build a
seam.

**The `.gitleaksignore` rider** (`spec:1461-1470`). Verified today:
`.gitleaksignore` is a single line pinning a `generic-api-key` finding
to `internal/config/keys_test.go` line 66, and line 66 is the
backup-targets fixture row inside `TestReservedClassOf`'s table. The
mechanism is line-number pinning, so any insertion above line 66 in
that file moves the finding and reds the gitleaks leg of
`scripts/check.sh` for every lane at once. The safe insertion point is
below line 66; otherwise re-fingerprint `.gitleaksignore` in the same
commit. The pre-task check (`spec:1680-1682`) passes as of HEAD.

#### A.1.4 The FLAC reader and its mapping

| Item | Spec | Tree |
| --- | --- | --- |
| Option (a): a minimal read-only FLAC reader inside `internal/engine/shadow`, unexported, pure stdlib, roughly 150 lines | `spec:995-1005` | — |
| Reads two block types: `STREAMINFO` (0, for track length) and `VORBIS_COMMENT` (4) | `spec:1003-1005` | — |
| No salvage adoption exists | `spec:983-989` | The only FLAC contact is `sampleInfo`, shelling `metaflac --show-sample-rate --show-bps` at `_salvage/internal/gate/gate.go:122-140`; it reads no Vorbis comment. Package doc delegating tags to `beet` at `_salvage/internal/beets/beets.go:1-3` |
| Named provisional in its own doc comment; rung 8 absorbs it into `internal/engine/tagio` | `spec:1007-1012`, `spec:1476-1478` | `tagio` named as planned at `internal/engine/tag/doc.go:8-11`, does not exist |
| Typed-default rule governs an absent or unparseable tag | `spec:1020-1022` | `ProjectItemFields`' rule at `internal/engine/matcher/info.go:178-188` |
| Fixture constructed in Go inside the test, no vendored binary | `spec:1024-1030` | `internal/harness/oracle/testdata/template.flac` stays harness territory |

**The mapping table, derived from pinned mediafile 0.17.0**
(`tools/venv/.venv/lib/python3.13/site-packages/mediafile/__init__.py`).
`matcher.ItemFields` is nineteen fields at
`internal/engine/matcher/info.go:124-152`, with the beets field names
at `:155-174`.

| `ItemFields` member | Vorbis key(s), in mediafile's own declaration order | mediafile line |
| --- | --- | --- |
| `Title` | `TITLE` | `:392` |
| `Artist` | `ARTIST` | `:398` |
| `Album` | `ALBUM` | `:410` |
| `AlbumArtist` | `ALBUM ARTIST`, then `ALBUM_ARTIST`, then `ALBUMARTIST` | `:542-544` |
| `Track` | `TRACK`, then `TRACKNUMBER` | `:463-464` |
| `Disc` | `DISC`, then `DISCNUMBER` | `:480-481` |
| `DiscTotal` | `DISCTOTAL`, then `DISCC`, then `TOTALDISCS` | `:488-490` |
| `Comp` | `COMPILATION` | `:535` |
| `Label` | `LABEL`, then `PUBLISHER` | `:574-575` |
| `Barcode` | `BARCODE`, then four read-only aliases (`UPC`, `EAN/UPN`, `EAN`, `UPN`) | `:616-620` |
| `CatalogNum` | `CATALOGNUMBER`, then read-only `CATALOGID`, `DISCOGS_CATALOG` | `:604-606` |
| `Country` | `RELEASECOUNTRY` | `:659` |
| `Media` | `MEDIA` | `:673` |
| `AlbumDisambig` | `MUSICBRAINZ_ALBUMCOMMENT` | `:680` |
| `MBTrackID` | `MUSICBRAINZ_TRACKID` | `:764` |
| `MBAlbumID` | `MUSICBRAINZ_ALBUMID` | `:782` |
| `Year` | derived from `DATE`, with a `YEAR` fallback, through `date.year_field()` | `:685-693` |
| `Length` | not a Vorbis comment. `STREAMINFO`'s total-sample count over the sample rate | `spec:1004` |
| `DataSource` | **no mediafile field exists.** A grep of pinned mediafile for `data_source` or `DATA_SOURCE` returns nothing | see A.4.5 |

`MUSICBRAINZ_ALBUMID` at `:782` and `CATALOGNUMBER` at `:604` are the
two rows the spec cites by name, and both are exact.

#### A.1.5 The capture

| Item | Spec |
| --- | --- |
| Walk `<observe dir>/<uploader>/<album>/` per pass; record per album directory a fixture: `matcher.ItemFields` per track, the observation's evaluation date, and a digest over the file set (names, sizes, modification times) | `spec:405-409` |
| The digest is the identity: an unmoved digest is not re-captured, so the capture is idempotent and a mid-write capture is superseded by the next one | `spec:409-411` |
| **Partial-read rule:** a capture that cannot read every entry it listed is discarded. Never stored, never scored, never counted. The pass logs it and moves on | `spec:413-423` |
| Reason the supersede rule does not cover it: once the importer moves an album out of the inbox there is no next pass. `review.Move` tries `os.Rename` first with a copy-then-remove EXDEV fallback (`_salvage/internal/review/move.go:24-40`), so a walk can read three tracks and take ENOENT on the fourth | `spec:415-423` |
| **Grouping rule:** after reading a directory's items, group them by `albumartist` and `album`. One group is observed; more than one is **excluded**, with its group count and basename recorded | `spec:441-447` |
| The report prints the excluded-multi-album count beside the gate-rejected one | `spec:447-450` |
| Porting beets' own `group_albums` pre-step is declined, named as the evidence-based follow-on if the excluded count is material | `spec:450-455` |
| **Quiescence rule:** the poll interval must be strictly less than the incumbent's quiescence window; the timer is `*:0/5` for that reason. The shadow reads no incumbent config; the constraint lives in the unit's comment | `spec:457-470` |

Tree support for the grouping rule: `config/beets-config.yaml:23` sets
`group_albums: yes`, with the estate's own reason at `:19-22`; the
incumbent runs one `beet import` over the whole staging tree at
`_salvage/internal/importer/importer.go:1100`; `tag.Album` has no
grouping and calls `matcher.MostCommonTags(items)` over every item it
is handed (`internal/engine/tag/tag.go:75`).

#### A.1.6 The reconcile step

| Item | Spec | Tree |
| --- | --- | --- |
| Read the incumbent's run reports, resolve each `AlbumOutcome` to a captured fixture, record the oracle signal | `spec:472-474` | `RunReport` at `_salvage/internal/importer/importer.go:152-172`, `AlbumOutcome` at `:134-150` |
| **The join key is the album basename, never `SourceDir`** | `spec:476-491` | `sourceDir` is documented "the uploader's inbox directory" at `importer.go:212`; all three assignments are uploader-level (`:766`, `:908`, `:1066`); `toReport` copies no per-album path (`:231-240`) |
| `SessionID` does not rescue it | `spec:487-490` | Empty for a pre-uploader leftover or a synthetic orphan, `importer.go:139-142` |
| `imported` is an **apply**; the oracle's release MBID is read from `mb_albumid` on the first file `library_paths` names, using the shadow's own FLAC reader | `spec:493-498` | Beets wrote it there under `write: yes` (`config/beets-config.yaml:15`); `MUSICBRAINZ_ALBUMID` is the Vorbis style (`mediafile/__init__.py:782`) |
| Applied files are matched back to the fixture by track titles and lengths as well as by `mb_albumid`, because `library_paths` is itself a positional correlation | `spec:499-505` | The positional match and its "one shared beet invocation carries no per-uploader join key" reason at `importer.go:1116-1136` |
| `review` is a **skip**, resolved by the album's basename in a tag log's `skip ` lines or under `<review dir>/<date>/<uploader>/<basename>` | `spec:506-511` | `parseSkips` at `_salvage/internal/beets/beets.go:115-135`; tag logs are written atomically at `:104` and never swept |
| `quarantined` and `rejected` are **not oracle verdicts**; the observation is discarded and the excluded count printed with its reason | `spec:512-516` | Gate-rejected material never reaches beets, `importer.go:945`, `:499-505` |
| Unresolvable observations are `unreconciled` with a reason, counted separately, under `CaseGradeSummary` discipline | `spec:518-521` | `internal/harness/conform/case.go:109-110` |
| **Three named `unreconciled` reasons:** an ambiguous basename; an `imported` outcome with empty `library_paths`; a fixture whose run report was already swept | `spec:522-528` | The empty-`library_paths` case is `AlbumOutcome.Error`'s "a beets-database match that never turned up" at `importer.go:145-149` |
| **The consumed-report ledger:** `consumed-reports.json` records which report files a pass has already reconciled, so a pass never re-reads one and a gap is a named, reported condition | `spec:538-544` | — |
| The run reports are a rolling window of roughly seven and a half hours | `spec:530-537` | `sweepReports` called at `importer.go:389`, defined at `:533-548`; `defaultReportRetention = 30` at `_salvage/internal/config/config.go:172`; timer `*:0/15` at `systemd/dubplate-import.timer:10` |
| Backstop H is swept at `RetentionDays`, default 60 | `spec:538-540` | `_salvage/internal/config/config.go:155`, sweep at `importer.go:386` |
| **Why not the beets database (mechanism C):** reading `mb_albumid` off the applied library file dissolves the config-wall problem, adds no sqlite driver, leaves the retirement message standing. "O5 is closed by not asking it" | `spec:548-557` | `DUBPLATE_BEETS_DB` retired at `internal/config/keys.go:178` |

#### A.1.7 The match step

For each reconciled observation: drive `tag.Album` against the
captured `ItemFields` under the estate tunables and `FullCompat()`,
spending `ClassBackground` budget, through a recording transport. Then
re-score that run's own candidate set under `DefaultCompat()`, driving
no second `tag.Album` and spending no second token. Classify, record,
and stage a pending row where the classification is news
(`spec:559-565`).

`tag.Album`'s signature is fixed and verified at
`internal/engine/tag/tag.go:68-69`, matching the 3d handoff's item 1
(`docs/HISTORY.md:273-279`). The construction inputs (`spec:928-948`):

- `src`: `musicbrainz.NewClient` refuses a missing `UserAgent` or
  `Budget` rather than defaulting (`internal/engine/musicbrainz/client.go:124-130`).
- `items`: the FLAC reader's output.
- `tun`: `matcher.QuietImportTunables()`.
- `compat`: both modes, per ruling 3.
- `evalDate`: the observation's capture date, stored so re-evaluation
  reproduces it.
- `opts`: zero. `SearchIDs` selects the explicit-ID branch
  (`tag.go:78-85`) and the shadow has none.

**The shadow reads no clock.** `shadow.Run` takes `now time.Time`,
supplied by `cmd/dubplate-server`, and every timestamp the pass writes
derives from that one value (`spec:964-970`).
`TestEngineNeverReadsTheClock` walks the whole `internal/engine` tree
and descends into nested packages
(`internal/harness/conform/arch_test.go:145-154`, walk at `:68-119`,
nested-descent proof at `:191-205`). Note that the check fires on
`time.Now` calls specifically (`arch_test.go:106-112`), so it
constrains the call and not every clock-shaped read.

#### A.1.8 The store

Layout, verbatim (`spec:736-747`):

```
<state dir>/shadow/
    started-at                      the clock start, written once, never rewritten
    passes/<pass-id>.json           one record per observe pass: start, end, counts, budget name and spend
    observations/<obs-id>/
        observation.json            the divergence record, or the agreement record
        case.json                   the cassette-shaped fixture, Expected nil
        responses/NNN.json          the captured response bodies
    pending/<obs-id>.json           the AdjudicationEntry a checkpoint batch presents
    resolved/<obs-id>.json          a divergence reevaluate retired, and the commit that retired it
    consumed-reports.json           the incumbent run reports already reconciled, so a gap is detectable
```

One directory per observation so two writers never contend for one
file (`spec:733-735`). Durable in the state directory per `BS:573-575`;
rung 6's store supersedes none of it.

**`passes/` exists because a `Budget` is in-process state**
(`spec:749-757`). `spent` is an int guarded by a mutex on the struct
and created per call to `Limiter.Budget` (`internal/engine/musicbrainz/limiter.go:139`,
accessors at `:400-408`); nothing persists it. `observe` and `report`
are separate process invocations, so each pass writes its own
`Budget.Name()` and `Budget.Spent()` and `report` sums them, printing
per-run and cumulative separately.

**No lock package.** `internal/lock` stays unadopted at rung 4: the
timer path is a systemd singleton, `reevaluate` writes only under
`resolved/`, `report` opens nothing for writing, and `observe` only
ever creates. The residual hand-run race costs at worst a re-captured
fixture (`spec:767-773`).

The `observations/<obs-id>/` shape matches the corpus case-directory
convention exactly, where `Request.Response` names a file relative to
the case directory (`internal/harness/cassette/cassette.go:72`,
readability check at `:225`, `corpus/README.md:24-25`).

#### A.1.9 The recording transport and the `DefaultCompat` re-score

**Ruling 3, option (c)** (provisional, owner ruling open), with the
mechanics corrected by fold B3 (`spec:206-259`).

| Item | Spec | Tree |
| --- | --- | --- |
| Mode one drives `tag.Album` under `FullCompat()` through a recording transport | `spec:559-563` | `FullCompat` at `internal/engine/matcher/compat.go:31-33`; the transport seam is `ClientConfig.HTTP` at `internal/engine/musicbrainz/client.go:98` |
| Mode two **re-scores the recorded candidate set**: translate the recorded responses, re-run the assignment solve, re-score the album distance, re-sort, recommend, all under `DefaultCompat().Apply(tun)`, issuing no search and driving no second `tag.Album` | `spec:228-231` | `DefaultCompat` at `compat.go:41-46`; `Apply` swaps `StringDistPatterns` at `:51-62` |
| The precedent it now correctly mirrors | `spec:210-225` | `GradeMode`'s doc at `internal/harness/conform/case.go:141-153` ("ModeDefault does not drive tag.Album"); `gradeCaseDefault` at `:517-526`, replaying only the lookups the committed candidates name at `:591-604` |
| The flow-change mechanism: the ID-consensus early return fires only at `RecStrong` | `spec:219-225` | `internal/engine/tag/tag.go:96-118`, condition at `:106`; fall-through to `src.SearchReleases` at `:125-126` |
| **The derived figure's limit is stated, not absorbed:** a re-scored `DefaultCompat` verdict is not the product's true one wherever the real flow would have widened the candidate set, so the compat-class annotation is a hypothesis and the divergence row says so | `spec:230-237` | `conform`'s own exactness argument covers ten pinned cases (`case.go:149-152`) and does not generalize |
| **The budget correction:** any replay client the re-scoring or `reevaluate` builds takes `musicbrainz.Unmetered()` | `spec:239-247` | `doWithRetry` calls `c.budget.Wait(ctx)` before every attempt (`client.go:346`) whatever `ClientConfig.HTTP` holds (`:98`); `Unmetered` at `limiter.go:354`; `gradeClient`'s precedent at `case.go:327-336` |
| **The conductor ruling that rides it:** `conform`'s `GradeMode` doc is itself marked provisional, and 4a retires that mark | `spec:255-259`, conductor fold at `spec:59` | The mark is at `internal/harness/conform/case.go:152-153` |

#### A.1.10 The divergence classification

**Four outcomes, three of them divergences** (`spec:636-649`):

| Oracle | Go | Class |
| --- | --- | --- |
| apply, release R | apply, release R | `agreement`, no row |
| skip | skip (or `tag.ErrNoCandidates`) | `agreement`, no row |
| skip | apply, release G | **`go-applies-oracle-skipped`** |
| apply, release R | skip (or `tag.ErrNoCandidates`) | **`go-skips-oracle-applied`** |
| apply, release R | apply, release G, `R != G` | **`go-applies-different-release`** |

`go-applies-oracle-skipped` is the class a false confident match hides
in and the launch bar's own clause (`BS:69-71`). The spec names it
deliberately rather than discovering it as an unclassifiable
observation (`spec:646-649`).

**The four deferring branches, an enumeration closed deliberately**
(`spec:651-668`):

1. `musicbrainz.ErrUnavailable` defers the observation for a later
   pass (`internal/engine/musicbrainz/client.go:74-84`).
2. `musicbrainz.ErrBudgetExhausted` defers the same way (`limiter.go:45-48`;
   returned from `Budget.Wait` at `:387`).
3. `musicbrainz.ErrLimiterClosed` defers the same way (`limiter.go:50-53`;
   returned at `:300`, `:319`).
4. `tag.ErrNoCandidates` records "no match", which agrees with an
   oracle skip and diverges from an oracle apply (declared
   `internal/engine/tag/tag.go:18`, returned `:139`).

**No error other than `tag.ErrNoCandidates` may ever be recorded as a
Go skip** (`spec:662-668`). A truncated album written as
`go-skips-oracle-applied` is a fabricated divergence in the one
artifact whose value is honesty, and it consumes exactly the owner
adjudication ruling B's stamp exists to save. The report's `deferred`
row counts by reason so a chronically undersized cap is visible.

The oracle's own signal is binary and is the only comparable quantity:
`_summary_judgment` returns `APPLY` when and only when the
recommendation is strong
(`tools/venv/.venv/lib/python3.13/site-packages/beets/ui/commands/import_/session.py:335-341`),
mirrored by `matcher.ActionFor` (`internal/engine/matcher/verdict.go:179-192`).
`RecMedium`, `RecLow` and `RecNone` are indistinguishable to the
oracle (`spec:610-623`).

#### A.1.11 The two rows

**The observation record, `observation.json`, rung 4's own type**
(`spec:684-693`):

| Field | Source |
| --- | --- |
| `id` | `"shadow-" + <observation id>`, deterministic from the capture |
| `album` | the captured `album` field, or the source directory's basename when empty |
| `matcher_commit` | ruling B's stamp (`R3:93-95`) |
| `matcher_config_hash` | a hash over the `Tunables` and `Compat` pair the run used |
| `eval_date` | the capture date |
| `oracle` | `signal` (apply/skip), `release_mbid` where one exists, `uploader`, `album_basename`, and the artifact the signal was read from |
| `go` | per mode: `recommendation`, `action`, the best candidate's album id and distance, and `Dist.NonzeroKeys()` |
| `class` | one of the three divergence classes, or `agreement` |
| `budget` | the pass's `Budget.Name()` and `Budget.Spent()` |

Tree support for the `go` block: `matcher.Candidate` is `{Info
*AlbumInfo, Dist *Distance, Assignment}` at
`internal/engine/matcher/candidate.go:12-16`; `AlbumInfo.AlbumID` at
`internal/engine/matcher/info.go:57`; `(*Distance).Distance()` at
`internal/engine/matcher/penalty.go:96`; `(*Distance).NonzeroKeys()`
at `penalty.go:166`.

**The staged pending row, `pending/<obs-id>.json`, an
`analysis.AdjudicationEntry` and nothing more, written only when the
class is news** (`spec:697-708`):

| Field | Source |
| --- | --- |
| `id` | the observation id |
| `subject.album` | the observation's `album` |
| `subject.release_mbid` | the release under dispute (ruling 2) |
| `matcher_commit`, `matcher_config_hash` | the two schema additions |
| `disposition` | `analysis.Pending` (`internal/harness/analysis/adjudication.go:18`) |
| `adjudicator` | `analysis.AgentAdjudicator` (`adjudication.go:30`) |
| `suspected`, `release_truth`, `reason`, `date` | four of the five fields the checker requires non-empty (`adjudication.go:176-190`), rendered from the observation rather than emitted as `TODO` |
| `citation` | the observation id and its class. A convention, never a check |

The five checked prose fields are `suspected` (`:176`),
`release_truth` (`:179`), `reason` (`:182`), `adjudicator` (`:185`),
and `date` (`:188`). `citation` is not among them, which is the
correction rung 4 does not inherit from the 3d handoff's wording at
`docs/HISTORY.md:295-302` (`spec:1522-1529`).

**`matcher_config_hash` is deliberately not `config_hash`**
(`spec:718-723`): `cassette.Expected.ConfigHash` is a hash of the
Python shim's effective config (`internal/harness/cassette/cassette.go:80`),
and conflating the two would make a rung 15 comparison silently wrong.

**`disposition` stays owner-only** (`spec:725-728`): a non-`Pending`
disposition set by `AgentAdjudicator` is a gate error
(`adjudication.go:198-200`).

**Ruling 2, option (a)** (provisional, owner ruling open):
`ReleaseMBID` is the release under dispute, filled by whichever side
proposed one, and the field's doc comment is restated to say so
(`spec:176-190`). The pair rule fires at `adjudication.go:168-171`; the
existing doc that already carries this reading is at
`adjudication.go:32-34`.

#### A.1.12 `analysis.AdjudicationEntry`'s two fields and their checks

```go
MatcherCommit     string `json:"matcher_commit,omitempty"`
MatcherConfigHash string `json:"matcher_config_hash,omitempty"`
```

Landing beside the existing fields at
`internal/harness/analysis/adjudication.go:46-67` (`spec:834-840`).

The matching rules in `checkAdjudicationEntry` (`adjudication.go:158`):
a shadow-subject row (`Album` and `ReleaseMBID` both set, the branch at
`:168`) requires `MatcherCommit`; a case-subject row is refused if it
carries either field. Both fields are `omitempty`, so the committed
corpus rows validate byte for byte unchanged. **Each rule ships a
hand-authored fixture in the package's own `testdata` proving it
fires** (`spec:842-848`).

#### A.1.13 The marshalling in `cmd/dubplate-server`

**`internal/engine/shadow` may import no harness package, so it
marshals neither artifact** (`spec:777-801`). `TestEngineNeverImportsHarness`
walks the whole `internal/engine` tree and fails any file importing
anything under `internal/harness` (`internal/harness/conform/arch_test.go:141-146`,
`checkNoHarnessImports` at `:31-64`, `testdata` skipped at `:38-40`).
So the package cannot reach `cassette.Case` or `analysis.AdjudicationEntry`.

The seam: `internal/engine/shadow` returns typed engine-side results
(observations, verdict pairs, classes, recorded responses);
`cmd/dubplate-server` imports both trees and writes `case.json` in
`cassette.Case`'s shape and `pending/<obs-id>.json` in
`analysis.AdjudicationEntry`'s. `observation.json` is the engine's own
type and needs no harness import (`spec:794-801`).

**Every staged row is validated before it is written** (`spec:803-811`):
the subcommand runs `analysis.CheckAdjudications` (exported at
`adjudication.go:138`) over each row before staging it, so a malformed
row fails on the box within one pass. `checkAdjudicationEntry` itself
is unexported (`:158`).

**The shadow writes into `corpus/analysis/adjudications.json` never**
(`spec:825-830`). The box stages rows under `pending/`; the conductor
fetches them read-only at a checkpoint; the owner rules and commits
them on the workstation. That preserves the 3d handoff's
runner-prints-and-writes-nothing rule (`docs/HISTORY.md:295-302`).

The shape is shared with the corpus runner's printed row, constructed
at `internal/harness/conform/case.go:467-477` and marshalled with
`json.MarshalIndent(row, "", "  ")` at `:478` (`spec:813-823`).

#### A.1.14 The subcommand's three verbs

`cmd/dubplate-server/shadow.go`, one file, matching the shape
`newCheckCmd` and `newDoctorCmd` set: a constructor taking `(environ
[]string, stdout, stderr io.Writer)` returning `*cobra.Command`, with
`SilenceUsage` and `SilenceErrors` both true and main as the sole
stderr writer (`spec:1062-1067`; `cmd/dubplate-server/check.go:16`,
`doctor.go:23`, `main.go:33-47`).

| Verb | Contract | Spec |
| --- | --- | --- |
| `shadow observe` | one capture, reconcile, and match pass. The timer's `ExecStart` | `spec:1069-1070` |
| `shadow report` | prints the launch-gate report | `spec:1071` |
| `shadow reevaluate` | runs the mechanical re-evaluation | `spec:1072` |

Registration edits `newRootCmd`'s single `cmd.AddCommand(...)` block
(`cmd/dubplate-server/main.go:48-52`). The doctor's integration rows
come from `internal/config`'s own tables rather than a hard-coded
list, matching how `doctor.go:33-51` already derives its wall and
reserved-class rows (`spec:1074-1077`).

**No flags are specified for any verb.** See A.4.1.

**The stamp refusal** (`spec:886-892`): `shadow observe` errors rather
than observing when the matcher stamp is empty or `dev`, because an
unstamped divergence cannot be mechanically retired and is worse than
no observation. Today `main.version` defaults to `"dev"`
(`cmd/dubplate-server/main.go:31`), `build-linux` carries no
`-ldflags` at all (`Makefile:6-7`), `main.matcherCommit` does not
exist anywhere in `cmd/dubplate-server`, and `newVersionCmd` prints
one line and exits zero (`main.go:56-64`).

#### A.1.15 The report

Printed in the pinned `"%-32s %s"` row form (`cmd/dubplate-server/doctor.go:63-67`),
rows at `spec:1097-1113`:

| Row | Source |
| --- | --- |
| clock start | `started-at`, written on the first successful pass |
| floor | the configured duration, default `720h` |
| elapsed against floor | `now` minus the clock start |
| observed albums | split applied / skipped |
| excluded, gate-rejected | never reached beets |
| excluded, multi-album directory | grouped to more than one album |
| unreconciled | by reason |
| report gaps | incumbent reports swept before the shadow consumed them |
| deferred | counted separately for `ErrUnavailable`, `ErrBudgetExhausted`, `ErrLimiterClosed` |
| divergences | by class, and by disposition over the four values at `adjudication.go:17-22` |
| false confident matches | `go-applies-oracle-skipped` rows whose disposition is anything but `oracle-wrong`. **Zero at the launch gate** |
| pending | staged rows awaiting the owner. **Zero at the launch gate** |
| matcher commits | each commit observed under, with its divergence count |
| budget | summed over `passes/`, per-run and cumulative |
| store size | the state directory's footprint |

**Two clauses gate, not one** (`spec:1118-1132`). Launch bar part 3 has
two halves (`BS:69-71`): the pending count, because `Pending` "is
legal at every gate this package runs; the launch gate, a different
gate at a different time, refuses it" (`adjudication.go:10-12`), and
the false-confident-match row. The observed album count is reported
and never gated (`BS:249-253`, `BS:812-815`).

#### A.1.16 `reevaluate`

`dubplate-server shadow reevaluate` walks every observation whose
`matcher_commit` differs from the running build's, replays that
observation's stored responses through the shadow's replay transport,
re-runs `tag.Album` at the observation's stored `eval_date`, and
(`spec:852-873`):

- **retires** a divergence that no longer reproduces, writing
  `resolved/<obs-id>.json` with the retiring commit and both verdicts,
  and removing the staged pending row;
- **restamps** a divergence that still reproduces with the current
  commit;
- **records a new divergence** on an observation that previously
  agreed, which a retire-only design would silently lose;
- **records un-re-evaluable** on an observation whose replay asks for
  a request the recording does not carry, leaving the row stamped and
  staged. The replay transport must answer a miss with a refusal
  rather than a 404 the flow would absorb, because a newer matcher
  revision can take a different branch for the same input exactly as
  a different `Compat` can.

The replay client takes `musicbrainz.Unmetered()` (`limiter.go:354`),
matching `gradeClient` (`case.go:330-336`). It spends no budget, reads
no network, and reads no clock (`spec:879`). This is the mechanism
ruling B's adjudication-load reduction rests on (`R3:93`, `DA:534-535`),
which INV O8 reports as unbuilt and unassigned anywhere.

#### A.1.17 The exit gate

`bash scripts/check.sh` green, plus one end-to-end `observe` run
against a hand-authored inbox and a hand-authored run report in the
pass's own `testdata`, producing (`spec:1616`):

- one agreement,
- one divergence of each class (three),
- one excluded multi-album directory,
- one deferred observation of each error branch (three).

Plus, from the review-scaling section (`spec:1816-1822`), **one
mutation proof per checker** on the standing assertion discipline.

#### A.1.18 What 4a explicitly does not build

`internal/engine/tagio` (rung 8), `internal/engine/store` and the ops
seam (rung 6), a beets database reader, a sqlite driver, applied
fields, monitoring, the pipeline wiring, corpus writes, adopter
deployment artifacts, and `internal/lock` (`spec:1472-1494`).
Everything under 4b's row is out of 4a: the two unit files,
`env/dubplate-server.env.template`, the six `deploy.sh` edits,
`dubplate-server version` printing both stamps, the fresh-context unit
read, the deploy sitting, the SELinux verifications, and the recorded
deploy date (`spec:1617`).

---

### A.2 Figures the spec relies on, re-verified against the tree

| Figure | Spec's use | Tree, verified |
| --- | --- | --- |
| Quiescence window | The `*:0/5` shadow interval must be strictly under it | `defaultQuiescentMin = 10` at `_salvage/internal/config/config.go:152`, applied at `_salvage/internal/importer/importer.go:790-798` |
| Import timer cadence | Shares the per-IP MusicBrainz budget; sets the report-window arithmetic | `OnCalendar=*:0/15` at `systemd/dubplate-import.timer:10`, `RandomizedDelaySec=300` at `:11`, `Persistent=true` at `:12` |
| Run-report retention | The rolling seven-and-a-half-hour window, `Persistent=true` on the shadow timer, the gap-detection requirement | `defaultReportRetention = 30` at `_salvage/internal/config/config.go:172`; sweep call at `importer.go:389`; `sweepReports` at `:533-548` |
| Review-tree retention | Backstop H is swept too | `defaultRetentionDays = 60` at `_salvage/internal/config/config.go:155`, sweep at `importer.go:386` |
| Run-report shape | The reconcile step's join and its `unreconciled` reasons | `RunReport` at `importer.go:152-172` (`Version`, `RunID`, `StartedAt`, `EndedAt`, `LockHeld`, `Error`, `Albums`); `AlbumOutcome` at `:134-150` (`Uploader`, `SourceDir`, `Outcome`, `SessionID`, `LibraryPaths`, `Collection`, `Error`); outcome constants at `:112-125` |
| Incumbent lock semantics | The trap the shadow's code must close on its own | `acquireOrRecordHeld` at `importer.go:397-413` writes a `LockHeld` report, pings success, and returns nil |
| Tag-log durability | Backstop D | Written through `atomicfile.WriteFile` at `_salvage/internal/beets/beets.go:104`, re-read from the persisted copy at `:108-113`; never swept |
| `library_paths` positional correlation | Why the applied half is matched back by titles and lengths as well | `importer.go:1116-1136`, with the "one shared beet invocation carries no per-uploader join key" comment at `:1123-1127` |
| Gate rejection never reaches beets | The excluded-count discipline | `importer.go:945`, `:499-505` |
| `.gitleaksignore` pin | Risk 7 and the pre-task | `.gitleaksignore` line 1 pins line 66 of `internal/config/keys_test.go`; line 66 is still the backup-targets fixture row |
| `check.sh` covers new units with no new wiring | Deploy edit 1's gate story | glob at `scripts/check.sh:137`, missing-executable filter at `:139` |
| `deploy.sh` line landings | Every one of the six edits | `UNIT_FILES` at `:56-64`, hand-enable precedent at `:49-55` and `:398-408`, `TIMER_UNITS` at `:407`, pre-create shape at `:357-359`, render loop at `:175-194` with resolution at `:181-182` and abort at `:190-194`, install block at `:291-316`, `REMOTE_ENV` at `:28`, binary step and rung-4 boundary comment at `:251-289`, smoke test at `:288-289`. File is 443 lines |
| Env-file collision | Why a second template is needed | `env/dubplate.env.template` carries exactly 39 `DUBPLATE_*` names; the two owned rows are `DUBPLATE_ADMIN_EMAIL` at `:25` and `DUBPLATE_STATE_DIR` at `:74`; `DUBPLATE_LIBRARY_DIR` at `:71` and `DUBPLATE_SERVE_SITE_NAME` at `:105` are the pre-rename names (superseded at `internal/config/keys.go:165-166`); the header anticipating rung 4 is at `:1-3` |
| `ping-fail` no-op | The `OnFailure=` line is wired and inert | `_salvage/cmd/dubplate/pingfail.go:36-42`; helper reads `/etc/dubplate/env` at `systemd/dubplate-ping-fail@.service:11`; the wire-ahead-of-mint precedent at `systemd/dubplate-serve.service:29-35` |
| Root-filesystem sizing | Risk 5 | `VOLUME_NAME="musicbox-data"` / `VOLUME_SIZE_GB=50` at `scripts/provision.sh:16-17` |
| The contested-surface rows | The lane-launch checklist item the conductor owed | All nine written at `docs/superpowers/specs/2026-09-03-parallel-build-approach.md:168-184`, each carrying "conductor ruling, 2026-09-05" |

---

### A.3 The do-not-re-litigate record

The spec review's "Verified sound" list has 28 items
(`docs/superpowers/research/2026-09-05-rung-4-spec-review.md:487-676`).
The plan must not reopen any of them. In brief: the ladder-entry and
build-spec quotations (1); the move-and-write hazard that decides the
mechanism (2); the oracle's binary signal (3); declining the beets
database (4); declining `ExecStartPost=` (5); the hardening set and
the `AF_NETLINK` reasoning (6); every `deploy.sh` edit landing on a
real line (7); the `vps-conventions` citations (8); the gate already
covering both new unit files (9); the environment-file diagnosis (10);
the `import.lock` trap (11); the quiescence arithmetic (12); the
`analysis` schema work and both fields as `omitempty` (13); the two
architectural tests requiring `now` as a parameter (14); the four
estate assignments (15); every `musicbrainz` citation (16); the
pseudo-release identity not manufacturing a false
`go-applies-different-release` (17); the 3d handoff read item for item
(18); the `.gitleaksignore` fragility (19); the threshold-registry gap
and leaving `KindFloat` to rung 3's follow-on (20); the subcommand
shape (21); `atomicfile` as a sound re-landing candidate (22); no
salvage FLAC reader existing (23); the gate-rejected exclusion (24);
the shared adjudication shape and the config-hash name split (25); the
pass split, tree allocation and review scaling (26); the staleness
pre-task's three blocking questions (27); the launch-gate and clock
arithmetic (28).

**One half of one finding is rejected and must not be re-raised**
(`:17-33`): review 1's "two templates cannot render" framing of B5's
first half. The renderer half is correct, but deploy edit 5's literal
column is the mechanism that resolves it, so the design can render.
What was actually wrong is that revision 1 never said which greenfield
rows are literals, and a bare `DUBPLATE_STATE_DIR=` line would
silently inherit the incumbent's exported value. B5 folds that half
as a required explicit statement.

---

### A.4 Where the spec leaves a plan-level choice open

These are places an implementer would have to invent an answer. Each
is a decision the plan owes, ordered by how much of 4a is blocked on
it.

#### A.4.1 The 4a/4b seam on the matcher stamp is unresolved

The spec assigns `shadow observe`'s refusal on an empty or `dev` stamp
to 4a (`spec:886-892`, and the three verbs are in 4a's scope row at
`spec:1616`), and assigns the `-ldflags` change, the second variable,
and `version` printing both rows to 4b (`spec:1336-1348`, `spec:1617`).
`main.matcherCommit` does not exist in `cmd/dubplate-server` today,
and `Makefile:6-7` sets no `-ldflags`. So 4a must land a stamp
variable that no build path sets, and 4a's own end-to-end `observe`
exit-gate run has to reach the observation code with a stamp that is
neither empty nor `dev`. The spec does not say whether the stamp is
injected as a `shadow.Run` parameter, an overridable package variable
the test writes, or a flag. **This is the sharpest under-specification
in the pass and the plan must settle it in task order.**

#### A.4.2 How the runner reads its configuration values

Ruling 1 option (a) fully specifies the wall's *acceptance* behavior
and the doctor's *reporting* behavior. It never specifies how the six
values reach the runner. `config.Config` is four fields
(`internal/config/config.go:56-65`) with no home for the observe,
oracle-state, or review paths. The choices are new `Config` fields, a
per-class accessor returning a name-to-value map, or a direct lookup
in the subcommand. The last would reintroduce a second environment
reader beside the package the wall's doc calls "the only place the
process environment is read" (`config.go:1-2`).

#### A.4.3 The two threshold rows sit in two mechanisms at once

The key table (`spec:1436-1443`) lists `DUBPLATE_SHADOW_FLOOR` and
`DUBPLATE_SHADOW_BUDGET` under "the rows, per ruling 1's option (a)",
and the subcommand section registers them as thresholds through
`config.Register` in `cmd/dubplate-server` (`spec:1078-1090`). A
registered threshold is not in `keys.go`'s tables, so `rejectNamespace`
(`config.go:133-152`) would reject either name as an unknown key the
moment the environment file sets it. Two readings are available and
the spec picks neither: the thresholds also join the integration
class's `Keys` (in which case "Load requires every key of every
enabled class" contradicts their having defaults), or the wall gains a
fourth state for a name that is accepted and never required. This is
close to an internal inconsistency rather than a mere gap, because the
greenfield environment template is specified to carry the floor as a
literal row (`spec:1267`), and that row would abort `Load` today.

#### A.4.4 A partial skip is a reconcile outcome the spec does not name

`resolveImported` marks a unit `OutcomeReview` with `Error: "left in
staging after import"` when beets took some but not all of its files,
and the function's own comment calls it "a partial skip that took some
but not all of a unit's files, the one shape a beets tag log can
produce that this album-as-a-unit model does not cleanly resolve"
(`_salvage/internal/importer/importer.go:1179-1199`). The spec's
reconcile step maps `review` to a clean oracle skip (`spec:506-511`).
Recording a partly-imported album as an oracle skip manufactures a
divergence whenever Go applies, which is the fabricated-divergence
failure the spec forbids in the next section (`spec:662-668`). The
grouping rule excludes multi-group directories, so this is narrower
than it first looks, but it is reachable for a single-group directory
with one unreadable track. The likely resolution is a fourth
`unreconciled` reason keyed on a non-empty `AlbumOutcome.Error` for a
`review` row.

#### A.4.5 The mapping table cannot cover nineteen fields as stated

The spec says "The Vorbis-comment key for each of the nineteen fields
is derived from pinned mediafile 0.17.0's own storage styles"
(`spec:1014-1018`). Two of the nineteen have no such key. `Length`
comes from `STREAMINFO`, which the spec acknowledges elsewhere
(`spec:1004`). `DataSource` has no mediafile field at all: pinned
mediafile 0.17.0 contains no `data_source` or `DATA_SOURCE` storage
style, because beets keeps it as a flexible attribute in its database
rather than in the file. The reader will always produce `DataSource ==
""` for an inbox file, which `ProjectItemFields`' typed-default rule
makes correct behavior (`internal/engine/matcher/info.go:115-123`
records that this field's true beets default is `None` and models it
as `""` anyway). The plan should state the table as seventeen
Vorbis-key rows plus one `STREAMINFO` derivation plus one field with
no file source, with the reason.

#### A.4.6 The mapping owes a precedence rule, not only keys

Seven of the seventeen fields have more than one Vorbis style, and
mediafile takes the first declaring style that yields a value:
`AlbumArtist` has three (`mediafile/__init__.py:542-544`), `DiscTotal`
three (`:488-490`), `CatalogNum` three with two read-only
(`:604-606`), `Barcode` five with four read-only (`:616-620`), `Track`
two (`:463-464`), `Disc` two (`:480-481`), `Label` two (`:574-575`).
`Year` has no key of its own and is derived from `DATE` with a `YEAR`
fallback (`:685-693`). `Comp` is a boolean parse of `COMPILATION`
(`:535`). The plan's per-row citation requirement (`spec:1017-1019`)
should carry the ordering and the read-only exclusions, not a single
key per field.

#### A.4.7 The observation id is undefined

The record's `id` is `"shadow-" + <observation id>`, "deterministic
from the capture" (`spec:686`), and the capture's identity is a digest
over names, sizes, and modification times (`spec:405-411`). Whether
the observation id *is* that digest is never said, and the choice has
consequences: the id names the store directory, the pending row, and
the `resolved/` row, so an id that moves when a file's mtime moves
would fragment an observation across passes, while an id that does not
incorporate the digest cannot distinguish a re-uploaded album from its
predecessor.

#### A.4.8 The report-gap mechanism has no sequence to work against

`consumed-reports.json` exists so "a gap in the report sequence is a
named, reported condition rather than silence" (`spec:541-544`). Run
report files are named `report-<runid>.json` and `RunID` is
explicitly "not derived from `StartedAt`, since several runs sharing a
test's frozen clock must not collide on one path"
(`_salvage/internal/importer/importer.go:139-145`). So the file names
carry no order. A gap can be detected from mtime ordering or by
reading `StartedAt`/`EndedAt` out of each report, and the spec picks
neither.

#### A.4.9 The tag-log basename resolution is under-specified

Backstop D resolves a skip "by the album's basename in a tag log's
`skip ` lines" (`spec:506-508`). The paths beets logs are staging
paths, and under `group_albums: yes` over a flat contributor directory
a skip line can name item files rather than an album directory. The
incumbent's own resolution is `findUnit(pending, path)` with a
`deriveOrphanUnit` fallback (`importer.go:1147-1161`), and
`parseSkips` returns whatever beets wrote
(`_salvage/internal/beets/beets.go:120-135`). The plan should either
mirror that resolution or state the narrowing that makes a plain
basename comparison sound.

#### A.4.10 The `case.json` fixture's non-mechanical fields

`cassette.Case` requires `Profile`, a profile name under
`corpus/profiles` (`internal/harness/cassette/cassette.go:44`), and
`EvalDate` as a string (`:45`). The spec specifies only "the
cassette-shaped fixture, `Expected` nil" (`spec:742`). The plan owes
the profile value (`"estate"` being the obvious candidate, since the
estate profile at `corpus/profiles/estate.yaml:16-19` is what the four
tunable assignments reduce to) and the date format.

#### A.4.11 The exit gate's error branches are not producible from an inbox alone

The gate asks for "one deferred observation of each error branch" from
one end-to-end `observe` run against hand-authored `testdata`
(`spec:1616`). `ErrLimiterClosed` is returned only when the limiter
shuts down (`internal/engine/musicbrainz/limiter.go:300`, `:319`) and
`ErrBudgetExhausted` only when the cap is reached (`:387`), so both
require either an injected transport and budget or a deliberately tiny
cap. `ErrUnavailable` requires every retry to be spent
(`internal/engine/musicbrainz/client.go:392`, `maxAttempts = 7` at
`:42`). The plan owes the seam: what `shadow.Run` accepts so a test can
drive all three, or a split of the exit gate into an end-to-end run
plus branch-level tests.

#### A.4.12 `shadow.Run`'s seam is named only by its constraints

The spec fixes `now time.Time` as a parameter (`spec:964-970`) and the
marshalling boundary (`spec:794-801`), and correctly leaves the rest to
the plan. But the two architectural tests and the recording and
replay transports together constrain the signature enough that the
plan should state it: no `internal/harness` import, no `time.Now`
call, and the HTTP transport supplied by the caller through
`ClientConfig.HTTP` (`internal/engine/musicbrainz/client.go:98`)
rather than built inside the engine.

#### A.4.13 `Register` from `newRootCmd` double-registers under test

`config.Register` appends unconditionally
(`internal/config/thresholds.go:28-30`) and the registry is a package
global (`:22`). The spec puts the call "from the root command's own
construction" (`spec:1080-1082`), and `newRootCmd()` is called from
`main_test.go:14` as well as from `main()`. A test binary that
constructs the root command more than once would register the shadow
thresholds more than once and print duplicate doctor rows.
`cmd/dubplate-server/doctor_test.go:98-108` iterates the same registry
it compares against, so this does not red the gate today, and it is
still a latent defect the plan should close by registering once.

#### A.4.14 "The first successful pass" is undefined

`started-at` is "written on the first successful pass"
(`spec:1099`, `spec:1586-1590`), and the launch date is that date plus
the floor (`DA:525`). A pass that captures nothing, reconciles
nothing, and errors on nothing is not obviously either successful or
not. The definition is load-bearing for a calendar figure the close is
required to record as first class.

#### A.4.15 Smaller open choices

- The `<pass-id>` scheme for `passes/<pass-id>.json` (`spec:739`).
- The on-disk format of `started-at` (`spec:738`).
- Whether the wire interval is recorded in the pass record. Only
  `Budget.Name()` and `Budget.Spent()` are (`spec:740`, `spec:755-757`),
  so a later reader of the report cannot reconstruct the pace a run
  used.
- What `analysis.CheckAdjudications`' `caseIDs` argument receives on
  the box. Passing `nil` is sound, because a shadow subject never
  reaches the case-resolution branch
  (`internal/harness/analysis/adjudication.go:162-171`), and the plan
  should say so, to keep an implementer from building a corpus reader
  into the box binary.
- The `doctor.go` edit for the duration kind. The spec folds "a switch
  in `thresholdStatus`" into the threshold work (`spec:1456`); the
  contested-surface row covers `doctor.go`'s rows, so it is authorized,
  and the plan should name the file explicitly rather than leaving it
  implicit inside a `config` task.

---

### A.5 Internal inconsistencies and stale text

#### A.5.1 The pass-split section cites a ruling that no longer exists

`spec:1608-1609` reads "4b cannot close without an owner sitting
inside it (ruling 7)". The deploy sitting is **ruling 6**
(`spec:342-368`). Revision 1's seventh ruling on response storage was
retired rather than answered (`spec:110-123`), and this cross-reference
was not renumbered with it.

#### A.5.2 Ruling 5 carries both its pre-fold and post-fold arithmetic

`spec:299-301` concludes "at one request per second the run's wire
time stays inside sixty seconds", and the M6 fold twenty-five lines
later re-derives the same run at two seconds a request as "under four
minutes" (`spec:328-330`). Under the conductor's ruling that the pace
is `2 * musicbrainz.DefaultInterval` (`spec:59`), the first sentence is
stale. The plan should quote only the second. Both figures are loose
in the safe direction: 60 tokens at two seconds a token bounds the run
at two minutes, not four.

#### A.5.3 The owner-touchpoint sentence contradicts the process the plan will run

`spec:1824-1825` says the six rulings are "batched at one checkpoint
**before 4a's plan is drafted**". The conductor's fold paragraph
(`spec:59`) and `docs/STATUS.md:24-26` both record the rulings as
still open with 4a proceeding provisionally. The plan should not read
`spec:1824` as a gate; the operative instruction is the header's, that
each recommendation is applied provisionally and marked inline.

#### A.5.4 "Nineteen fields" against the mapping's real reach

Covered at A.4.5. It reads as an inconsistency rather than a gap
because `spec:1011-1012` restates it as "rung 4 owns nineteen fields
read one way".

#### A.5.5 `Dist.NonzeroKeys()` names a field, not the type

`spec:691` writes `Dist.NonzeroKeys()`. `Candidate.Dist` is a
`*matcher.Distance` (`internal/engine/matcher/candidate.go:14`) and
`NonzeroKeys` is a method on `*Distance` (`internal/engine/matcher/penalty.go:166`).
The expression is reachable; the phrasing just names the field where a
reader might look for a type.

---

### A.6 Citation drift, non-blocking

Every load-bearing citation I checked is exact, including all of
`deploy.sh`, `check.sh:137`/`:139`, `config/beets-config.yaml`, the
salvage importer and its config defaults, `.gitleaksignore` and
`keys_test.go:66`, both mediafile rows, the whole `musicbrainz` set,
the `conform` set, and the 3d handoff at `docs/HISTORY.md:271-303`.
Four citations are off by a line or two and none changes a claim:

- `_salvage/internal/atomicfile/atomicfile.go:14` (`spec:763`): line
  14 is inside the doc comment; `WriteFile` begins at `:15`.
- `internal/harness/cassette/cassette.go:45-56` (`spec:797`): `Case`
  runs `:41-57`.
- `systemd/dubplate-serve.service:96-125` (`spec:1156`): the hardening
  block runs `:98-126`.
- `_salvage/internal/beets/beets.go:116-135` (`spec:507`): `parseSkips`'
  doc begins at `:115`.

---

### A.7 What the plan may treat as settled

Landed and verified, so no task needs to re-derive them: `tag.Album`'s
signature and `Proposal`'s five fields
(`internal/engine/tag/tag.go:23-29`, `:68-69`); the shadow subject
shape already modeled with zero committed rows
(`internal/harness/analysis/adjudication.go:32-39`); `ClassBackground`
naming this caller (`internal/engine/musicbrainz/limiter.go:16-18`,
`:25`); `NewLimiter`, `Budget`, `Spent`, `Name`, `Unmetered`,
`DefaultInterval` and the three sentinels, all at the lines the spec
cites; the subcommand skeleton's constructor shape and single
`AddCommand` block; the `shadow` verb already reserved in `main.go:11`;
the doctor's pinned row form; both architectural tests and their own
fires-on-its-own-fixture proofs (`internal/harness/conform/arch_test.go:158-205`);
and all nine contested-surface rows, written into the parallel-approach
spec at `:168-184`.

---

## (b) The engine, harness and config seams as landed

Read against the working tree at `HEAD` (`54e534b0`). Every figure
below is checked against the file, never against the spec's paraphrase.
Where the spec's own citation was verified I say so; where it is off I
say that too.

### 1. `internal/engine/tag`: the entry point the shadow drives

The signature is exactly as the 3d handoff records it
(`internal/engine/tag/tag.go:68-69`):

```go
func Album(ctx context.Context, src *musicbrainz.Client, items []matcher.ItemFields,
    tun matcher.Tunables, compat matcher.Compat, evalDate time.Time, opts Options) (*Proposal, error)
```

`Proposal` is five fields (`tag.go:23-29`): `CurArtist`, `CurAlbum`,
`Candidates []matcher.Candidate` (documented ascending distance,
stable), `Recommendation matcher.Recommendation`, `Action
matcher.Action`. `Options` is one field, `SearchIDs []string`
(`tag.go:39-41`); the shadow passes the zero value, taking the
ID-consensus and search path (`tag.go:78`).

Flow facts the classification rests on:

- The ID-consensus early return fires only at `rec ==
  matcher.RecStrong` with `!tun.Timid` and `cands.Len() > 0`
  (`tag.go:106`), shipping the unsorted insertion-order slice, which
  holds at most one candidate (`tag.go:111-117`). So `Candidates[0]`
  is well defined on both paths whenever the call returns without
  error.
- The search path calls `src.SearchReleases` at `tag.go:126` and fans
  out one `lookupAndAdd` per returned id (`tag.go:130-134`).
- `ErrNoCandidates` is declared at `tag.go:18` and returned at
  `tag.go:139`. It is the only "no match" signal.
- A 404 on a lookup is swallowed inside `lookupAndAdd`
  (`tag.go:167-169`), so `musicbrainz.ErrNotFound` never reaches the
  shadow.
- Every other client error is wrapped in a message, `"tag: looking up
  %s: %w"` (`tag.go:170`) or `"tag: searching: %w"` (`tag.go:128`).
  **The shadow must classify with `errors.Is`, never equality**, for
  `ErrUnavailable`, `ErrBudgetExhausted` and `ErrLimiterClosed` alike.
- `Album` applies compat and validates before anything else
  (`tag.go:70-73`), so a bad `Tunables` fails before a request is
  spent.

The package doc states the no-harness and no-clock rules for itself
(`internal/engine/tag/doc.go:13-14`) and names the `tagio` pairing at
`doc.go:8-11`.

Test-shape precedent worth adopting: `tag_test.go:25-50` serves
fixtures from an `httptest.Server` recording every
`r.URL.RequestURI()`, and `tag_test.go:56-60` sets `Retry-After: 0` so
a retryable status collapses its backoff instead of spending 31 real
seconds (the fold-M-h note at `tag_test.go:52-55`). The shadow's
`ErrUnavailable` branch test needs the same trick.

### 2. `internal/engine/matcher`

**`ItemFields`** is nineteen fields (`info.go:124-152`), eighteen
scalars plus `Comp bool`. Field names and Go types, in declaration
order: `Artist`, `Album`, `AlbumArtist` (string), `Year`, `DiscTotal`
(int), `MBAlbumID`, `Label`, `Barcode`, `CatalogNum`, `Country`,
`Media`, `AlbumDisambig`, `DataSource` (string), `Length` (float64),
`Title` (string), `Track` (int), `MBTrackID` (string), `Disc` (int),
`Comp` (bool). The beets field-name constants are `info.go:156-176`.

**`ProjectItemFields(raw map[string]any) ItemFields`** (`info.go:189-211`)
is the only landed producer, and it reads a decoded JSON bag rather
than a file. Its typed-default rule is the one the FLAC reader must
reproduce (`info.go:178-188`, helpers at `:215-259`): an absent or
wrongly typed value takes the Go zero value. `info.go:148-150` carries
the standing warning the spec leans on, that two decoders over one
field bag drift silently.

**`Tunables`** is `tunables.go:61-245`; `DefaultTunables()` is
`tunables.go:266-331`, with `SearchLimit: 5` at `tunables.go:309`,
`Timid: false` at `:312`, `Quiet: false` at `:315`, `QuietFallback:
"skip"` at `:316`, `NoneRecAction: "ask"` at `:317`. `Validate()` is
`tunables.go:455-498`; it resolves `Required`, both `Preferred`
pattern lists, every `MaxRec` level in sorted key order, and the two
action keys against their separate legal sets, joining every problem
rather than returning the first.

**`Compat`** is two bools (`compat.go:15-22`); `FullCompat()`
`compat.go:31-33`, `DefaultCompat()` `compat.go:41-46`, `Apply`
`compat.go:51-63`. `Apply` swaps `StringDistPatterns` and nothing
else, which is why a mode change moves distance but never translation.

**Verdicts.** `Recommendation` is an ordered int enum (`verdict.go:12-19`)
with `String()` at `:23-36` and `ParseRecommendation` at `:41-54`.
`Action` is a string type with four constants (`verdict.go:60-67`).
`ActionFor(tun, rec)` is `verdict.go:179-210`; under `Quiet` it
returns `ActionApply` when and only when `rec == RecStrong`, else the
parsed `QuietFallback`. With the estate's `QuietFallback = "skip"` the
projection is exactly `apply`/`skip` and can never be `asis` or `ask`,
which is what makes the oracle comparison binary. `Recommend` is
`verdict.go:107-130`, `downgrade` `verdict.go:146-169`.

**`Candidate`** is `{Info *AlbumInfo; Dist *Distance; Assignment
Assignment}` (`candidate.go:12-16`). The chosen release id for a
divergence row is `Proposal.Candidates[0].Info.AlbumID`
(`info.go:57`). The row's `Dist.NonzeroKeys()` is `penalty.go:166`;
`Dist.Distance()` is `penalty.go:96`.

**Where `QuietImportTunables` sits.** `gradeTunables` is
`internal/harness/conform/defaults.go:72-79`, and its overlay is one
line, `defaults.go:77` (`tun.Quiet = true`). The other three
assignments are not there. They live in `withImportKeys`
(`internal/harness/conform/translate.go:221-226`), reached through
`mergeProfile` (`translate.go:203-217`) for the `estate` and `locale`
profiles, and they set `QuietFallback = "skip"`, `NoneRecAction =
"skip"`, `Timid = false`. The estate profile file agrees
(`corpus/profiles/estate.yaml:16-19`), as does the box config
(`config/beets-config.yaml:16-18`).

So `matcher.QuietImportTunables()` is exactly `DefaultTunables()` plus
those four assignments, and it equals `gradeTunables("estate")` value
for value today. `conform` cannot consume it in place of
`gradeTunables`, because `gradeTunables` must stay profile-driven
(`stock` overlays nothing, `translate.go:205-206`). The two
definitions therefore coexist, which is a live drift risk at the seam
`info.go:148-150` warns about. A one-line `reflect.DeepEqual`
assertion in `conform` (harness reads engine, the permitted direction)
closes it cheaply.

### 3. `internal/engine/musicbrainz`

**`ClientConfig`** is `client.go:93-99`: `BaseURL`, `UserAgent`
(required), `Budget *Budget` (required), `Timeout`, and `HTTP
*http.Client`. Its doc comment (`client.go:86-92`) specifies the
User-Agent composition verbatim, `dubplate-server/<version> ( <contact>
)`, built in `cmd/dubplate-server` from `Config.AdminEmail`.
`NewClient` refuses a missing `UserAgent` or `Budget` (`client.go:124-130`).

**The transport seam a recording transport hooks is
`ClientConfig.HTTP`** (`client.go:98`, consumed at `client.go:137-144`).
A recording or replay `http.RoundTripper` is installed as `HTTP:
&http.Client{Transport: rt, Timeout: ...}`. One trap: when `HTTP` is
non-nil the client ignores `Timeout` entirely (`client.go:137-144`),
so a recording client must set its own timeout or inherit
`http.Client`'s zero value, which is no timeout at all.

**Requests.** `releaseRequestURL` (`client.go:183-189`) and
`searchRequestURL` (`client.go:193-200`) both build the query through
`url.Values.Encode()`, which sorts by key. The recorded `path_query`
is therefore already in `cassette`'s canonical form by construction
(see section 5).

**Retry.** `doWithRetry` is `client.go:343-393`. It calls
`c.budget.Wait(ctx)` before **every** attempt including retries
(`client.go:346`), which is R13 divergence 2 and the reason a retried
request spends a second token. `maxAttempts = 7` (`client.go:42`),
retryable statuses at `client.go:50-56`. `retryWait` (`client.go:403-412`)
honors a delta-seconds `Retry-After` and falls through to the computed
backoff on an HTTP-date form, because reducing a date needs a clock
read the arch test forbids. `backoff` is `client.go:418-423`, `500ms *
2^(n-1)` capped at 120s, zero on the first attempt.

**`AlbumForID`** is `client.go:296-334`. The pseudo-release second
fetch is at `client.go:325`, guarded by `PseudoReleaseTarget` at
`:320`, and it is **per candidate**, which is the basis of the
eleven-request worst case.

**Limiter.** `Class` at `limiter.go:20-26` with `ClassBackground` at
`limiter.go:25` and its comment naming rung 4's shadow at
`limiter.go:16-18`. `reservationInterval = 4` (`limiter.go:43`).
`ErrBudgetExhausted` (`limiter.go:48`, doc `:45-48`), `ErrLimiterClosed`
(`limiter.go:53`, doc `:50-53`). `DefaultInterval = time.Second`
(`limiter.go:62`). `NewLimiter(interval time.Duration) *Limiter`
(`limiter.go:110-115`); the interval is the caller's, and
`reservationInterval` and the class policy are untouched by changing
it, exactly as the spec claims. `Limiter.Budget(name string, class
Class, limit int) *Budget` (`limiter.go:139-145`); **a limit of zero
or less means no cap** (`limiter.go:130-133`). `Unmetered()`
(`limiter.go:354-359`) never waits and never exhausts. `Budget.Wait`
(`limiter.go:372-381`) reserves before waiting and releases on any
error out of the wait. `Spent()` `limiter.go:400-404`, `Name()`
`limiter.go:408`.

Two consequences for 4a's exit gate. `Unmetered()` can never produce
`ErrBudgetExhausted`, so the exhaustion fixture needs a real `Limiter`
with a positive limit; a `NewLimiter` at a sub-millisecond interval
keeps the test off the clock. `ErrLimiterClosed` is reachable through
`Limiter.Close()` (`limiter.go:149`).

### 4. `internal/harness/conform`: the re-score the second mode mirrors

`gradeCaseDefault` is `case.go:539-659`, and its five steps are the
shape rung 4's second mode copies:

1. `compatFor(ModeDefault, policy).Apply(tun)` then `Validate`
   (`case.go:568-572`).
2. For each committed candidate, `client.AlbumForID(ctx, w.AlbumID,
   tunDefault)` against the case's own stub, **issuing no search**
   (`case.go:591-604`).
3. `cands.Add(tunDefault, items, info, evalDate)` per candidate, which
   re-solves the assignment and re-scores (`case.go:605-608`).
4. `cands.Sorted()` then `matcher.Recommend` (`case.go:612-613`).
5. `matcher.ActionFor` (`case.go:618`).

`gradeClient` is `case.go:330-336` and takes `musicbrainz.Unmetered()`,
the precedent the spec cites for both the re-score and `reevaluate`.
Note the re-score **re-translates from the response bytes** rather
than reusing the `ModeCompat` `AlbumInfo`. Translation output is
identical between modes, because `Compat.Apply` touches only
`StringDistPatterns` (`compat.go:51-63`), so the re-fetch is a shape
choice rather than a necessity. Mirroring it is still the right call,
since it makes the stored responses load-bearing and the two paths
identical.

**The provisional mark 4a retires** is the last sentence of
`GradeMode`'s doc comment, "Revision 2, fold B2; provisional, conductor
ruling owed", at `case.go:152-153`. The paragraph it closes is
`case.go:141-153` and states the flow-change argument the spec quotes.
The conductor's ruling (spec `:59`) settles it, so the edit is
deleting that clause and naming the ruling.

Other cited anchors, all verified: `CaseGradeSummary`'s "what it
compared, never what it visited" doc at `case.go:109-110`; the pending
row constructed at `case.go:467-477` and marshalled with
`json.MarshalIndent(row, "", "  ")` at `case.go:478`; `res.graded =
true` on the `ModeDefault` path at `case.go:610`, the dead-but-harmless
field 3d recorded; `CheckDefaultCompatMatchesTheLedger` at `case.go:668`.

The package charter is `internal/harness/conform/doc.go:1-48`,
including "Nothing here writes `corpus/`" at `doc.go:44-47`.

### 5. `internal/harness/analysis` and `internal/harness/cassette`

`AdjudicationSubject` is three fields (`adjudication.go:35-39`),
`AdjudicationEntry` is ten plus `DefectFree` (`adjudication.go:46-67`).
Dispositions at `adjudication.go:17-22`, `Pending` at `:18`;
`AgentAdjudicator = "agent"` at `adjudication.go:30`.
`LoadAdjudications` `adjudication.go:125-131`;
`CheckAdjudications(caseIDs []string, l *AdjudicationLedger) error`
`adjudication.go:138-156`; `checkAdjudicationEntry` unexported at
`adjudication.go:158`.

The pair rule is the switch at `adjudication.go:161-174`. The shadow
shape is accepted at `:168-169` when `Album` and `ReleaseMBID` are
both set, and `:170-171` errors when one is set without the other.
The five non-empty prose checks are `:176-190` (`Suspected`,
`ReleaseTruth`, `Reason`, `Adjudicator`, `Date`); `Citation` is never
checked, confirming 3d's own correction (`docs/HISTORY.md:1524-1529`).
The agent rule is `:198-200`. The `defect_free` block's own rules are
`:205-221`.

Two facts that govern where the new fields land:

- `readJSON` sets `DisallowUnknownFields` (`ledger.go:19-24`). Adding
  `MatcherCommit` and `MatcherConfigHash` with `omitempty` is
  therefore safe in both directions: the 21 committed rows re-marshal
  byte-identically, and a shadow-written row carrying only known keys
  loads.
- `CheckAdjudications` is the exported entry the subcommand can call
  over a single-row ledger; its first argument is the known case ids,
  which a shadow-subject row never consults (`adjudication.go:164-167`
  is the only reader).

Per-row rules in this package are proven by **inline Go literal
variants** in `TestCheckAdjudications` (`adjudication_test.go:10-113`,
base entry at `:11-21`, shadow subject at `:23-24`, half-shadow at
`:29-30`). `testdata/` holds cross-ledger fixtures only
(`testdata/defect_free`, `testdata/oracle_bugs_probe_rule`, loaded at
`adjudication_test.go:203-219`). The spec asks for testdata fixtures
for the two new rules; the package's own precedent for a per-row rule
is the inline table.

`cassette.Case` is `cassette.go:41-56`, `Request` `:69-73`, `Expected`
`:76-88`, `Candidate` `:103-109`. `LoadCase` validates that every
`path_query` is already canonical (`cassette.go:168-170`, `:189-191`),
and `canonicalizePathQuery` (`cassette.go:260-268`) is **unexported**,
so `cmd/dubplate-server` cannot call it. The client's
`url.Values.Encode()` already sorts, so a recorded request is
canonical by construction, and a hand-authored 4a fixture must be
written that way by hand. `Case.Write` (`cassette.go:237-252`) uses
`json.MarshalIndent(c, "", " ")`, a single-space indent, plus a
trailing newline and `os.WriteFile` at `0o644`, which is not an atomic
write.

### 6. `internal/harness/conform/arch_test.go`, every rule

Two invariants, four proofs, read in full.

- **`TestEngineNeverImportsHarness`** (`arch_test.go:141-146`) walks
  `internal/engine` with `checkNoHarnessImports` (`:30-65`). It fails
  any `.go` file importing `github.com/glw907/dubplate/internal/harness`
  or anything under it (`:21`, prefix test at `:55`). The walk uses
  `parser.ImportsOnly` and **skips any directory named `testdata`
  entirely** (`:36-40`). It descends into nested packages, so
  `internal/engine/shadow` is covered the day it exists.
- **`TestEngineNeverReadsTheClock`** (`arch_test.go:152-157`) walks
  the same root with `checkNoTimeNow` (`:75-120`). It resolves the
  local name of the `time` import (`importedName`, `:124-136`) and
  flags any `time.Now` call. It parses with `parser.AllErrors`, walks
  **`_test.go` files too**, and skips only `testdata`. So the shadow's
  own tests may not call `time.Now` either; `tag_test.go:20` is the
  pattern to copy, a package-level fixed `evalDate`. The check catches
  `time.Now` alone, so `time.Since` and friends are not mechanically
  forbidden, which makes the code rule the real guard.
- The two fire-proofs are `TestHarnessImportCheckFiresOnItsOwnFixture`
  (`:163-172`) and `TestTimeNowCheckFiresOnItsOwnFixture` (`:178-187`),
  each against a fixture tree under `testdata/arch/`.
- `TestTimeNowCheckDescendsIntoNestedPackages` (`:194-206`) pins the
  descent behavior against a nested fixture, which is the guarantee
  that a new `internal/engine/shadow` cannot escape the walk.

Nothing else in the tree is an architectural check; the only other
`Test*Never*` names are behavioral.

No mechanical rule stops `internal/engine/shadow` importing
`internal/config`. The convention is written prose, at
`tunables.go:50-54` ("This pass names them and defers registration; it
imports nothing from `internal/config`") and `client.go:86-92`
(identity belongs to config and reaches the engine through `cmd`). The
shadow should take its paths and knobs as parameters on the same
grounds.

### 7. `internal/config`, read in full

**`keys.go`.** `WallRow{Name, Pending}` at `:8-14`; `ownedWall` five
rows at `:19-25` (`DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`,
`DUBPLATE_SITE_NAME`, `DUBPLATE_ADMIN_EMAIL`, and the pending "identity
source" row at `:24`). `OwnedWall()` returns a clone (`:28-30`);
`ownedNow` at `:34-41`. `ReservedClass{Name, Rung, Keys}` at `:47-51`;
`reservedClasses` is eight classes at `:57-142`, the shape an
`IntegrationClass` table sits beside. `ReservedClasses()` `:145-147`,
`reservedClassOf` `:150-157`. `supersededKeys` `:164-183`, with
`DUBPLATE_LIBRARY_DIR` at `:165`, `DUBPLATE_SERVE_SITE_NAME` at `:166`,
and both beets retirements at `:177-178`. `supersessionOf` `:187-190`.

**`config.go`.** `Config` is four fields (`:56-65`), `AdminEmail`
normalized lowercase at `:60-64` and parsed at `:114-125`. `wall` and
`newWall` at `:72-89`; `get` treats an empty value as unset (`:92-98`).
`rejectNamespace` is `:133-152`, and `Load` is `:159-175`.

**A gap the plan must close explicitly.** `rejectNamespace` exempts
only `ownedNow(name)` (`config.go:135`). It knows nothing about
`registeredThresholds`. Today a set `DUBPLATE_SHADOW_FLOOR` or
`DUBPLATE_SHADOW_BUDGET` would fail `Load` with "unknown configuration
key" (`config.go:150`), even after `config.Register`. So the
integration class, or some other acceptance path, has to cover the two
threshold names as well as the four required-when-enabled ones, or the
rendered `/etc/dubplate/server.env` fails at load on the box, which is
the exact failure mode ruling 1 exists to prevent.

**`thresholds.go`.** `Threshold{Key, Default, Doc}` at `:13-17`, all
strings. `registeredThresholds` is a package-level slice at `:22`, and
`Register` is a bare append with no dedup (`:28-30`), called "from an
explicit setup call rather than an init function" at `:24-27`.
`Thresholds()` clones (`:33-35`). `ParseInt(environ []string, t
Threshold) (int, error)` is `:43-60` and is the only parser; it parses
`t.Default` when the key is unset and reports a bad default as a
registration bug (`:48-51`). **A duration threshold needs three
edits**: a kind discriminator on `Threshold`, a `ParseDuration` beside
`ParseInt`, and a switch in `thresholdStatus`, which today calls
`ParseInt` unconditionally (`cmd/dubplate-server/doctor.go:100-107`)
and would render a duration as `invalid:`.

The append-only global has a second consequence. `doctor_test.go`
builds its expectation from `config.Thresholds()` itself
(`doctor_test.go:98-108`) and then asserts there are no extra lines
(`:110-112`), so it tolerates new rows. But if registration happens
inside `newRootCmd`, every test that constructs a root command appends
again (`main_test.go:14` calls it once already), and nothing resets
the registry outside `internal/config`'s own test helper
(`thresholds_test.go:13-15`).

**`keys_test.go` and `.gitleaksignore`.** `.gitleaksignore` holds one
line, a `file:rule:startline` fingerprint pinning
`internal/config/keys_test.go` at line 66. Line 66 is the first row of
`TestReservedClassOf`'s table literal, whose rows run `:66-72` inside
a table declared at `:60-73`; the file is 123 lines. The fingerprint
is line-number keyed and content-independent, so **any insertion above
line 66 in that file moves the finding and reds the gitleaks leg**
(`scripts/check.sh:107`), while an insertion below it is inert.
Appending a new test function at the end of the file is the safe
form. The alternative is re-fingerprinting `.gitleaksignore` in the
same commit. Adding rows to `keys.go` shifts nothing in `keys_test.go`.

### 8. `cmd/dubplate-server`

`main.go` is 65 lines. The doc comment carries the verb list, with
`shadow` already reserved for rung 4 at `main.go:11`; a rung claiming a
verb amends that list (`main.go:5-6`). `version` is a package var
defaulting to `"dev"` with the `-ldflags` claim at `main.go:29-31`.
`newRootCmd` is `:41-54` and the single `AddCommand` block is
`:48-52`, three constructors today. `newVersionCmd` prints one line
and returns (`:56-65`), confirming a smoke test cannot assert a
matcher stamp today.

The subcommand shape a new file copies is `newCheckCmd(environ
[]string, stdout, stderr io.Writer) *cobra.Command` (`check.go:16-32`)
and `newDoctorCmd` with the same signature (`doctor.go:23-61`). Both
set `SilenceUsage` and `SilenceErrors` true and call
`cmd.SetErr(stderr)`, leaving `main` the sole stderr writer
(`check.go:13-15`, `main.go:35-38`).

`doctor.go` prints three row groups in order, owned wall (`:33-38`),
reserved classes (`:40-45`), thresholds (`:47-51`), all derived from
`internal/config`'s tables. `printRow` pins `"%-32s %s"` at
`doctor.go:64-66`, the form the launch-gate report and the two version
stamps reuse. `rowStatuses` `:73-95`, `thresholdStatus` `:100-107`,
`splitKeyReason` `:113-119`.

`go.mod` requires cobra alone (`go.mod:5`), with two indirect pflag and
mousetrap entries. Go 1.27.0.

### 9. `_salvage`: `atomicfile`, and the FLAC question

`_salvage/internal/atomicfile/atomicfile.go` is 42 lines, one exported
function, `WriteFile(path string, data []byte, perm os.FileMode)
error` (`:15-41`): temp file in the target's own directory, write,
`Sync`, `Close`, `Chmod`, `Rename`, with a `defer os.Remove(tmpPath)`
cleanup at `:22`. Re-landing means the file plus its tests. The test
file is 95 lines and sets the bar: a four-case table over create,
overwrite-shorter, permissions, and empty content (`atomicfile_test.go:9-70`),
a leftover-temp-file assertion (`:72-87`), and a missing-directory
failure case (`:89-95`). That is a clean merit admission with nothing
to rewrite; a modern-idiom pass would touch nothing beyond confirming
`os.CreateTemp` usage.

**No FLAC reading code is adoptable.** The only FLAC contact in the
salvage tree is `_salvage/internal/gate/gate.go`, which shells out:
`decodes` runs `flac -t` (`gate.go:116-118`) and `sampleInfo` runs
`metaflac --show-sample-rate --show-bps` and parses two lines
(`gate.go:122-141`). Neither reads a Vorbis comment. The extension
allowlist at `gate.go:26` is the only other mention. So INV b.2's
finding holds and the spec's pure-stdlib reader stands; the one usable
idea is the `exec.CommandContext` shape, which option (a) does not
need.

**The container layout has a committed precedent for the test
fixture.** `internal/harness/oracle/testdata/gen_template_flac.py`
authors a 42-byte valid FLAC, and its `main` states the exact framing
(`gen_template_flac.py:64-75`): four-byte `fLaC` signature, a one-byte
header of last-block flag `0x80` OR'd with the block type, a
three-byte big-endian payload length, then a 34-byte STREAMINFO
payload. The committed `template.flac` is 42 bytes and matches. A Go
fixture builder adds a second block of type 4 and moves the
last-block flag onto it. That file stays harness territory and unread
by the engine.

**The mapping's evidence, and its two traps.** The Vorbis key for each
field is the bare `StorageStyle("KEY")` entry in pinned mediafile
0.17.0. Spot-verified: `MUSICBRAINZ_ALBUMID` at
`mediafile/__init__.py:782`, `CATALOGNUMBER` at `:604`,
`RELEASECOUNTRY` at `:658`, `MEDIA` at `:672`,
`MUSICBRAINZ_ALBUMCOMMENT` at `:678`, `MUSICBRAINZ_TRACKID` at `:763`,
`TITLE` at `:392`, `ARTIST` at `:398`, `ALBUM` at `:410`. Two traps the
plan's table must record:

1. **Several fields carry more than one Vorbis key, in a read order
   that matters.** `albumartist` has three, `ALBUM ARTIST` then
   `ALBUM_ARTIST` then `ALBUMARTIST` (`:542-544`); `track` has `TRACK`
   then `TRACKNUMBER` (`:463-464`); `disc` and `disctotal` likewise
   (`:480-481`, `:488-490`); `label` has `LABEL` then `PUBLISHER`
   (`:574-575`); `barcode` and `catalognum` each carry read-only
   aliases. `MediaField.__get__` walks the styles in declaration order
   and takes the **first truthy** value (`mediafile/fields.py:48-54`),
   so an empty or zero value falls through to the next key.
   `StorageStyle.fetch` takes `mutagen_file[key][0]`, the first value
   of a multi-valued comment (`mediafile/storage/base.py:77-84`), and
   mutagen's Vorbis lookup is case-insensitive.
2. **Two of the nineteen fields are not tags at all.** `length` is a
   property off the container's STREAMINFO
   (`mediafile/__init__.py:892-895`), which is why the reader parses
   block type 0. `data_source` has no `MediaField` anywhere in
   mediafile, because it is a beets flexible attribute; on an inbox
   file it is absent and takes the Go zero value under
   `ProjectItemFields`' own rule. `year` is not a plain field either:
   it comes from a `DateField` over `DATE` with a `YEAR` fallback
   (`:685-693`).

### 10. The gate a new package must pass

`scripts/check.sh` runs, in order: `make check` (`:41`), shellcheck
over `scripts/*` (`:70-77`), ruff `check` and `format --check` over
tracked `*.py` (`:85-96`), `gitleaks detect --no-git -s . --redact`
(`:107`), `scripts/vale-comments.sh` (`:112`), `docker compose config`
when Docker is present (`:117-123`), and the pattern-filtered
`systemd-analyze verify` over `systemd/*.service systemd/*.timer
systemd/*.path` (`:136-148`). The missing-executable filter matches
`/usr/local/bin/[A-Za-z0-9_.-]+` (`:139`), which covers
`dubplate-server`, so a new unit needs no wiring. Every tool is a hard
requirement; a missing one exits 1 (`:19-38`).

`make check` is `vet lint test tidy-check` (`Makefile:27`), so `go vet
./...`, `golangci-lint run ./...`, `go test ./...`, and a `go mod
tidy` diff check (`Makefile:12-23`). `build-linux` carries no
`-ldflags` today (`Makefile:6-7`), confirming the spec's B6 finding.
`derive` is oracle-tagged and outside the gate (`Makefile:38-39`).

`.golangci.yml` enables errcheck (with `check-type-assertions: true`),
govet, ineffassign, staticcheck, unused, modernize, misspell, unparam,
plus gofmt and goimports formatters. `unparam` and `unused` matter for
a new package: an unexported helper nothing calls, or a parameter
nothing reads, reds the gate.

Vale lints Go **comment prose** as Markdown through the vendored
`glw907` overlay (`.vale.ini:5-11`), at `--minAlertLevel=error`
(`vale-comments.sh:59`), over tracked and untracked `.go` files with
`_salvage/` excluded (`vale-comments.sh:53-56`). The em dash is banned
with no opt-out; `glw907.Judgment` is downgraded to warning
(`.vale.ini:16`). A brand-new file is scanned before it is staged, so
a new package's doc comments are gated from the first run.

### 11. Figures the plan's floors and mutations rest on

| Figure | Value | Site |
| --- | --- | --- |
| `maxAttempts` | 7 | `client.go:42` |
| `SearchLimit` default | 5 | `tunables.go:309` |
| pseudo-release second fetch | per candidate | `client.go:320-333` |
| token spend | one per attempt, retries included | `client.go:346` |
| `DefaultInterval` | `time.Second` | `limiter.go:62` |
| `reservationInterval` | 4 | `limiter.go:43` |
| `Budget` limit semantics | `<= 0` means no cap | `limiter.go:130-133`, `:386` |
| `defaultTimeout` | 30s, ignored when `HTTP` is set | `client.go:38`, `:137-144` |
| `ItemFields` field count | 19 | `info.go:124-152` |
| `AdjudicationEntry` fields | 10 plus `DefectFree` | `adjudication.go:46-67` |
| non-empty prose checks | 5, `citation` excluded | `adjudication.go:176-190` |
| committed shadow-subject rows | 0 (rung 4 is the first producer) | `docs/HISTORY.md:1510` |
| `.gitleaksignore` pinned line | 66 of a 123-line file | `.gitleaksignore:1` |
| owned wall rows | 4 required plus 1 pending | `keys.go:19-25` |
| reserved classes | 8 | `keys.go:57-142` |
| registered thresholds | 0 | `thresholds.go:22`, no non-test caller of `Register` |
| `AddCommand` entries | 3 | `main.go:48-52` |
| doctor row form | `"%-32s %s"` | `doctor.go:65` |
| `atomicfile.go` | 42 lines, one function | `_salvage/internal/atomicfile/atomicfile.go` |
| FLAC template | 42 bytes, `fLaC` + 1-byte header + 3-byte length + 34-byte STREAMINFO | `gen_template_flac.py:64-75` |
| module dependencies | cobra only | `go.mod:5` |

Two spec citations I could not confirm as written. The env-template
table calls the greenfield shadow rows "the four `DUBPLATE_SHADOW_*`
rows" and then enumerates five values, while the key table lists six
rows including the enable flag (spec `:1267`, `:1436-1443`). And
`RunReport`/`AlbumOutcome` live at
`_salvage/internal/importer/importer.go:156-172` and `:134-150`, close
to the spec's `:126-172` but worth pinning exactly, since
`AlbumOutcome` carrying no per-album path is the whole basis of the
basename join (`:134-150`, `SourceDir` documented at `:136`).

---

## (c) The estate inputs and the process record

### 1. The incumbent's run report: the shape 4a's testdata must imitate

The report is the reconcile step's primary artifact (mechanism E), and
it is a versioned Go struct, not a loose bag.

**`RunReport`** at `_salvage/internal/importer/importer.go:155-172`:

| Field | JSON | Notes |
| --- | --- | --- |
| `Version` | `version` | `ReportVersion = 2` (`importer.go:103`) |
| `RunID` | `run_id` | names the file; 8 random bytes hex-encoded (`importer.go:456-462`), deliberately not time-derived |
| `StartedAt` / `EndedAt` | `started_at` / `ended_at` | RFC 3339 through `time.Time` |
| `LockHeld` | `lock_held` | true when the run exited on a held lock; every field below is zero (`:162-164`) |
| `Error` | `error,omitempty` | per-run failure; empty on a clean run, a held lock included |
| `Albums` | `albums` | `[]AlbumOutcome` |

**`AlbumOutcome`** at `importer.go:134-150`: `uploader`, `source_dir`,
`outcome`, `session_id,omitempty`, `library_paths,omitempty`,
`collection,omitempty`, `error,omitempty`.

**The outcome vocabulary**, four values, `importer.go:107-125`:

- `imported` (`OutcomeImported`, `:109`) is an apply.
- `review` (`OutcomeReview`, `:112`) is a skip.
- `quarantined` (`OutcomeQuarantined`, `:115`) and `rejected`
  (`OutcomeRejected`, `:123`) are gate outcomes, never oracle verdicts.

**The two semantics the spec review corrected, verified against the
tree.**

`SourceDir` is set at the uploader level by all three assignment
sites: `importer.go:769` (`filepath.Dir(srcDir)` in the
session-staging path), `importer.go:905-907` (`collectUnits`,
`sourceDir: inboxUserDir`), and `importer.go:1066-1068`
(`deriveOrphanUnit`, `filepath.Join(im.Cfg.Paths.InboxDir, uploader)`).
`albumUnit.stagedDir` is the only per-album path (`importer.go:213`)
and `toReport` does not copy it (`importer.go:231-240`).
`stageAllContributors` stages every quiescent album of every
contributor in one run (`importer.go:782-800`), so several rows
routinely share one `SourceDir`. The basename join is the only
resolution available, and 4a's fixtures must be authored so that a
`source_dir` join would be ambiguous, or the join's correctness is
untested.

`LibraryPaths` is populated only for `imported` and is a **positional**
correlation, not a join: `beetImportAndFileSkips` matches staging
order against the beets database's album order, and a count mismatch
leaves the extra units' `LibraryPaths` empty rather than guessing
(`importer.go:1119-1127`, with the reason in the comment at
`:1119-1123`). `AlbumOutcome`'s own doc says the same
(`importer.go:129-133`). The `Error` field's documented example is "a
beets-database match that never turned up" on an outcome that still
reads `imported` (`importer.go:145-149`), which is the `unreconciled`
reason the spec names.

**The file's path and naming.** `report-<runid>.json` under the state
directory: `reportFilePrefix = "report-"` and `reportFileSuffix =
".json"` (`importer.go:76-77`), joined in `reportPath`
(`importer.go:508-510`), written atomically through
`atomicfile.WriteFile` (`importer.go:527`) after `MkdirAll` on the
state directory (`importer.go:520`). On the box that directory is
`/var/lib/musicbox` (`env/dubplate.env.template:59-70`,
`scripts/deploy.sh:340-357`).

**The rolling window is real.** `sweepReports` keeps at most
`retention` of the newest, deleting by ascending mtime
(`importer.go:533-548`), called on every pass at `importer.go:389`,
with `defaultReportRetention = 30`
(`_salvage/internal/config/config.go:172`, wired at `:368`). The
import timer fires `*:0/15` with `RandomizedDelaySec=300`
(`systemd/dubplate-import.timer:10-11`), so thirty reports is roughly
seven and a half hours.

**The quiescence window**, at its cited lines: `cutoff :=
im.Now().Add(-time.Duration(im.Cfg.Thresholds.QuiescentMin) *
time.Minute)` then `quiescentSince(inboxUserDir, cutoff)`
(`importer.go:792-793`), with `defaultQuiescentMin = 10`
(`_salvage/internal/config/config.go:152`, wired at `:362`).
`quiescentSince` walks files rather than directories, and the comment
explains why (`importer.go:1273-1298`): appending two levels down
updates no parent's mtime. The window is per **contributor**, not per
album, so one still-changing file anywhere under
`<inbox>/<uploader>/` holds that uploader's whole tree back for the
run.

**The lock file and its flock semantics.** `lock.Path(stateDir)` joins
`import.lock` (`_salvage/internal/lock/lock.go:20-30`), and its doc
states the persistent-directory rule outright. `Acquire` opens
`O_CREATE|O_RDWR` and takes `LOCK_EX|LOCK_NB`, mapping `EWOULDBLOCK`
to `ErrHeld` (`lock.go:43-68`); it then clears `FD_CLOEXEC` because
`dubplate beet` holds the lock through an exec (`lock.go:59-66`).
`acquireOrRecordHeld` treats `ErrHeld` as healthy: it writes a
minimal `LockHeld` report, pings **success**, and returns nil
(`importer.go:405-412`). That is risk 3's whole mechanism, and it is
why the shadow must never open that path.

### 2. The tag log

The retired shell layer's Go successor is
`_salvage/internal/beets/beets.go`. `Runner.Import` runs `beet import
-q -l <scratch>` (`beets.go:83-88`), reads the scratch log, and
persists it atomically to `<StateDir>/import-<timestamp>.log` where
the timestamp is `time.Now().Format("20060102-150405.000000000")`
(`beets.go:103-107`). Nothing sweeps it, which is why the spec takes
it as the durability backstop.

What it records is skips only. `parseSkips` (`beets.go:116-135`) keeps
lines whose prefix is `skip ` (`:120`), splits the remainder on `;`,
trims at most one leading space per part, and drops empties. The paths
beets logs are **staging** paths (the call site passes
`im.Cfg.Paths.StagingDir`, `importer.go:1096`), so a skip line reads
`skip /srv/music/staging/<uploader>/<album basename>`. The basename
survives the inbox-to-staging move for a directory upload: `stageUser`
moves each inbox entry to `filepath.Join(destDir, name)` with `name`
the inbox entry's own name (`importer.go:844-856`).

**One estate-shaped hole the plan should name.** An **archive** upload
does not preserve the basename. `stageArchive` extracts into a
scratch directory and moves each extracted top-level entry under a
collision-safe name (`importer.go:861-888`), then removes the archive
from the inbox (`:887`). The shadow's capture walks album directories,
so a `.zip` in the inbox is never captured at all, and the album that
eventually appears in the run report has a basename the shadow never
saw. That is a coverage gap in the capture step, not a defect in the
join, and it belongs in 4a's declared-gap list.

### 3. The review tree layout

`filerFor(dateStr, uploader)` builds
`filepath.Join(im.Cfg.Paths.ReviewDir, dateStr, uploader)`
(`importer.go:499-505`), and `dateStr` is the **run's** `StartedAt`
formatted `2006-01-02` (`importer.go:271`), not the album's date. So
the layout is `<review dir>/<YYYY-MM-DD>/<uploader>/<basename>`,
matching `docs/STATUS.md:46-47`'s live example, and the estate's
`ReviewDir` is `/srv/music/review` (`cloud-init/user-data.yaml:118`).

The basename is **not** guaranteed equal to the album basename.
`moveInto` uses `uniqueDestPath`
(`_salvage/internal/review/review.go:64-73`), which appends an
`-HHMMSS` suffix on a same-day collision and then `.2`, `.3`, …
(`review.go:83-98`). The spec's backstop H therefore resolves on a
prefix rather than an equality, and 4a's fixtures should include one
suffixed directory to prove the matcher does not require exactness.
`Sweep` prunes top-level dated subdirectories at `RetentionDays`,
defaulting to 60 (`review.go:117-125`, `importer.go:385-387`,
`_salvage/internal/config/config.go:155`).

### 4. The inbox layout the import timer consumes

`listContributorDirs` reads the inbox's top-level **directories**
only, sorted (`importer.go:1238-1254`), and `stageUser` reads one
level below that (`importer.go:839-857`). So the consumed layout is
`<inbox dir>/<uploader>/<top-level entry>`, where the entry is an
album directory or an archive file. The estate's inbox is
`/srv/music/inbox`, created `chmod 2775` by cloud-init
(`cloud-init/user-data.yaml:118`, `:121`) and granted to the serve
unit (`systemd/dubplate-serve.service:94`). The shadow's own
`DUBPLATE_SHADOW_OBSERVE_DIR` points at that same path (spec's config
table), and its capture walk is `<observe dir>/<uploader>/<album>/`.

### 5. `config/beets-config.yaml`, read in full

56 lines. What matters to 4a, with line numbers:

- `directory: /srv/music/library` (`:10`) and `library:
  /srv/music/state/library.db` (`:11`). The second is the file the
  spec's mechanism C declines to open.
- `import.move: yes` (`:14`) and `import.write: yes` (`:15`). Together
  these are the capture-before-import hazard: files are renamed into
  the library and their tags rewritten with beets' answer.
- `quiet_fallback: skip` (`:16`), `none_rec_action: skip` (`:17`),
  `timid: no` (`:18`). These are three of the four assignments
  `matcher.QuietImportTunables` must apply; the fourth, `Quiet: true`,
  comes from the `-q` flag at `_salvage/internal/beets/beets.go:87`
  rather than from the file.
- `group_albums: yes` (`:23`), with the estate's own reason recorded
  at `:19-22`, including the 2026-08-31 nine-albums-as-one-135-track
  candidate incident. Still set at HEAD; the staleness pre-task's
  check passes today.
- `original_date: yes` (`:27`) and `per_disc_numbering: yes` (`:29`).
  Both change the **applied** year and track numbers, so a library
  file's tags are not a faithful echo of the pre-import inbox tags
  even for fields the matcher reads.
- `paths` (`:37-41`), the query-prefixed collection routing at `:38`
  and the unprefixed default at `:39`, with `singleton` and `comp` at
  `:40-41`.
- `plugins: musicbrainz fetchart embedart replaygain` (`:47`), with the
  explicit-musicbrainz reason at `:43-46`. `fetchart.auto`,
  `embedart.auto` and `replaygain` (`:49-56`) all write into the
  library file after the import.

**The applied field the reconcile step reads.** `mb_albumid` reaches
the FLAC through mediafile's Vorbis storage style
`MUSICBRAINZ_ALBUMID`
(`tools/venv/.venv/lib/python3.13/site-packages/mediafile/__init__.py:782`);
the same file carries `CATALOGNUMBER` at `:604` as the second key the
spec cites. `corpus/profiles/estate.yaml:16-19` is this file reduced
to the subset that can move a verdict, and it confirms that no
`match:` and no `musicbrainz:` block exists, so the whole distance and
threshold surface runs at beets' shipped defaults.

### 6. What the hand-authored testdata must contain

#### 6a. The transport: `stub.Server` cannot serve this run

Three independent reasons, each decisive on its own.

1. **The arch wall forbids it inside the engine.**
   `checkNoHarnessImports` walks every `.go` file under the root and
   skips only directories named `testdata`
   (`internal/harness/conform/arch_test.go:30-44`), so `_test.go`
   files are in scope, and `TestEngineNeverImportsHarness` points it
   at `internal/engine` (`arch_test.go:141-146`). Any test file in
   `internal/engine/shadow` that imports `internal/harness/stub` or
   `internal/harness/cassette` reds the gate.
2. **A stub miss is silent on the path that matters.** `ServeHTTP`
   records a `Miss` and answers 404 (`internal/harness/stub/stub.go:147-148`).
   In `tag.Album`'s flow a 404 on a lookup is absorbed and the
   candidate dropped (`internal/engine/tag/tag.go:166-170`, doc at
   `:63-66`), so a fixture authoring error becomes a changed verdict
   rather than a failure. Only a search 404 surfaces (`tag.go:128`,
   via `SearchReleases`' documented `ErrNotFound` at `client.go:250-252`).
3. **The stub cannot express a non-200 status at all.**
   `writeResponse` always writes 200 (`stub.go:151-154`), the only
   other answer is the 404, and `cassette.Request` carries no status
   field (`internal/harness/cassette/cassette.go:69-73`; the
   response-status extension is declined for the fourth pass running,
   `docs/HISTORY.md:326-329`). The `ErrUnavailable` branch needs a
   retryable status, so it is unreachable through the stub.

**The precedent to follow is `internal/engine/tag/tag_test.go`.**
`taggingServer` is a plain `httptest.NewServer` routing
`/ws/2/release` to `search.json` and `/ws/2/release/<id>` to
`release-<id>.json` under `testdata/<name>`, recording every request
URI in flow order (`tag_test.go:25-51`); `statusServer` answers a
fixed status with `Retry-After: 0` so a retryable status collapses its
backoff to no wall time (`tag_test.go:56-66`, reason at `:53-55`);
`clientFor` builds a `musicbrainz.Client` with `BaseURL` pointed at the
server and `Budget: musicbrainz.Unmetered()` (`tag_test.go:69-76`).
The committed fixture directories are `early-return` (one release plus
a search body) and `search-fallback` (two releases plus a search
body).

**The cassette-shaped fixtures are outputs, not inputs.** The spec's
`observations/<obs-id>/case.json` is written in `cassette.Case`'s
shape with `Expected` nil, marshalled in `cmd/dubplate-server` (spec's
B4 fold). The test asserts on that written file. A test in
`cmd/dubplate-server` may legally import `internal/harness/cassette`
to decode and assert it, and that assertion is worth making:
`LoadCase`'s canonicalization rule (`cassette.go:15-25`) is exactly
what a rung 15 consumer will apply to the shadow's output, so proving
the written case round-trips through `cassette.LoadCase` is the
honest check.

#### 6b. The eight album directories the exit gate needs

The exit gate names one agreement, one divergence of each of three
classes, one excluded multi-album directory, and one deferred
observation of each of three error branches. That is eight inbox
directories under one uploader, plus the artifacts that resolve them.

| # | Basename | Inbox FLAC tags | Response fixture | Run report row | Backstop | Expected outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | `agree-apply` | unanimous `mb_albumid = R1`; title/length/artist/album matching R1 closely | `release-R1.json`, scoring `RecStrong` | `imported`, `library_paths` → a library FLAC tagged `MUSICBRAINZ_ALBUMID=R1` | none | agreement, no row. One request (the ID-consensus early return, `tag.go:96-118`, condition at `:106`) |
| 2 | `agree-skip` | no `mb_albumid` | `search.json` with an empty `releases` array | `review` | tag log `skip ` line, and `<review>/<date>/<uploader>/agree-skip` | agreement. Go returns `ErrNoCandidates` (`tag.go:139`) |
| 3 | `go-skips` | no `mb_albumid` | `search.json` with empty `releases` | `imported`, `library_paths` → FLAC tagged `MUSICBRAINZ_ALBUMID=R2` | none | `go-skips-oracle-applied` |
| 4 | `go-applies` | unanimous `mb_albumid = G1` | `release-G1.json`, `RecStrong` | `review` | tag log skip line plus a review directory carrying a `-HHMMSS` suffix, per `review.go:83-98` | `go-applies-oracle-skipped`. Ruling 2 fills `subject.release_mbid` from `Candidates[0].Info.AlbumID` |
| 5 | `go-different` | unanimous `mb_albumid = G2` | `release-G2.json`, `RecStrong` | `imported`, `library_paths` → FLAC tagged `MUSICBRAINZ_ALBUMID=R3` | none | `go-applies-different-release`. `R3` needs no response body, only the library tag |
| 6 | `multi-album` | four FLACs carrying two distinct `(albumartist, album)` pairs | none | none needed | none | excluded, multi-album directory. Never reaches the match step |
| 7 | `defer-unavailable` | unanimous `mb_albumid = U1` | the handler answers 503 with `Retry-After: 0` for `/ws/2/release/U1` | any row, or none | none | deferred, `ErrUnavailable` after `maxAttempts = 7` (`client.go:42`, `:385-392`) |
| 8 | `defer-budget` | unanimous `mb_albumid = B1` | `release-B1.json` (never served) | any row, or none | none | deferred, `ErrBudgetExhausted` from `Budget.reserve` (`limiter.go:378-390`) |

Sorted-basename ordering puts `defer-budget` before the others
alphabetically, so the pass's iteration order must be stated in the
plan and the fixture names chosen against it, or the budget cap will
exhaust the wrong album. A lexical walk over `os.ReadDir` is the
obvious rule and is what `listContributorDirs` already does
(`importer.go:1252`).

**The gate-rejected exclusion**, though the exit gate does not require
it, costs one more directory plus a `rejected` or `quarantined` row
and is worth adding: the report prints that count beside the
multi-album one, and a report row that reads zero forever is
indistinguishable from a broken counter.

**The `unreconciled` reasons** are similarly cheap and cover three
named branches (`importer.go:127-133`, `:145-149`): an ambiguous
basename (two uploaders with an identically named album in one
report), an `imported` row with empty `library_paths`, and a fixture
whose report was already swept (a `consumed-reports.json` naming a
report file that no longer exists).

#### 6c. `ErrLimiterClosed` cannot share a pass with `ErrBudgetExhausted`

This is a finding, not a preference. `Budget.Wait` reserves before it
waits (`limiter.go:369-377`, with the lost-update reason documented at
`:362-368`), so `reserve` returns `ErrBudgetExhausted`
(`limiter.go:385-388`) before the injected `wait` function is ever
called. Once the cap is spent, every later album gets
`ErrBudgetExhausted` and the limiter is never consulted. In the other
order, `Limiter.Close` fails every pending **and future** waiter
(`limiter.go:148-149`, returns at `:300` and `:319`), so once closed,
every later album gets `ErrLimiterClosed` and the cap is never
reached. And `Budget.wait` is an unexported field (`limiter.go:340`),
so no test outside `internal/engine/musicbrainz` can inject a
per-album failure.

A single `observe` pass therefore cannot produce all three deferral
branches while the pass holds one `Budget`, which the spec's `passes/`
record requires it to. Recommendation for the plan: the end-to-end
run covers `ErrUnavailable` and `ErrBudgetExhausted`, and
`ErrLimiterClosed` is proven by a second focused test that runs one
pass against a pre-closed `musicbrainz.NewLimiter` and asserts every
observation defers with that reason.

#### 6d. The FLAC fixtures, and where the builder lives

`matcher.ItemFields` is nineteen typed fields plus `Comp`
(`internal/engine/matcher/info.go:124-152`), with the beets field
names the reader's mapping targets listed at `info.go:154-176` and the
typed-default rule at `ProjectItemFields` (`info.go:178-188`). The
end-to-end inbox needs, per track, at minimum `ALBUMARTIST`, `ALBUM`,
`TITLE`, `TRACKNUMBER`, and `MUSICBRAINZ_ALBUMID` where the fixture
calls for a consensus, plus a `STREAMINFO` block carrying the length.

The spec builds the reader's own fixture in Go inside the test. The
end-to-end run needs the same builder from a different place. If the
observe test lives in `cmd/dubplate-server` (where the marshalling
is), an unexported builder in `internal/engine/shadow`'s test file is
unreachable. Three ways out, and the plan should pick one at
task-authoring time: put the end-to-end test in
`internal/engine/shadow` against an `httptest` server and assert the
marshalled artifacts in a separate, smaller `cmd` test; duplicate a
ten-line FLAC builder in the `cmd` test; or commit the eight albums'
FLACs as binary testdata. The third contradicts the spec's "no binary
fixture is vendored twice" only if the reader's unit test also
vendors one, so it is available, and gitleaks runs `--no-git` over the
tracked worktree with the corpus already in scope (`CLAUDE.md`, Gate),
so binary testdata adds no new gate exposure.

### 7. The process record

#### 7a. The lane-launch checklist (`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:196-204`)

Eight items, each present because something went wrong without it
(`:196`):

1. Reverse edges cleared for any rung run out of ladder order.
2. **Contested surfaces written** for every surface the lane touches.
3. Runaway guard armed, one per lane or one polling every lane's
   transcript directory, mandatory past roughly 30 minutes.
4. **Lane registry entry written**, which replaces `pgrep -f
   Projects/dubplate` as the live-executor check, because worktrees
   live under `.claude/worktrees/` inside the repo and that pattern
   matches every lane by design once a second exists.
5. Venv provisioned if needed (`uv sync` against
   `tools/venv/requirements.txt`), with the standing policy that
   **corpus and harness work never moves to a worktree**.
6. Sub-ceiling assigned from the phase ceiling.
7. AC power for a multi-lane session, gate lock in force.
8. Throughput baseline recorded.

The conductor cap is two code lanes plus any number of document lanes
(`PA:194`), the conductor stays thin with no source file, diff, or
gate transcript (`PA:200`), and each lane's plan header carries a
sub-ceiling summing to the phase ceiling so the 80% stop rule has a
yardstick inside a lane that cannot see the others (`PA:199`).

#### 7b. The rung 4 rows already written into `PA`'s ownership table

All nine the spec's contested-surfaces section owes are present at
HEAD, each stamped `conductor ruling, 2026-09-05`:

| Line | Surface |
| --- | --- |
| `PA:178` | the state directory layout, rung 4, first to deploy one |
| `PA:179` | `corpus/analysis/adjudications.json` and the divergence schema |
| `PA:180` | `internal/engine/shadow` |
| `PA:181` | `internal/atomicfile`, born in pass 4a, re-landed from salvage on merit |
| `PA:182` | `cmd/dubplate-server` |
| `PA:183` | `internal/config` (`keys.go`, `thresholds.go`), rung 4 as the only live code lane, **re-fingerprinting `.gitleaksignore` on any insertion above its pinned line** |
| `PA:184` | `systemd/`, `scripts/deploy.sh`, `env/`, pass 4b's half |
| `PA:169` | `internal/harness/{cassette,stub,analysis}`, amended for rung 4's `MatcherCommit` and `MatcherConfigHash` |
| `PA:171` | `internal/engine/{matcher,musicbrainz}`, amended for `QuietImportTunables` **and nothing else** |

Lane-launch item 2 is therefore satisfied before 4a's launch, and the
plan can cite these rows rather than re-deriving them.

**The `.gitleaksignore` mechanism, verified.** `.gitleaksignore` holds
one line, `internal/config/keys_test.go:generic-api-key:66`. Line 66
of `internal/config/keys_test.go` is the first table row of the
class-classification test, the `DUBPLATE_R2_BUCKET` row. The
fingerprint is by **line number**, so any insertion above line 66 in
that file moves the finding and reds the gitleaks leg for every lane
at once. The safe insertion point for rung 4's new key rows is
**below** line 66; the alternative is re-fingerprinting
`.gitleaksignore` in the same commit. Neither the plan nor any commit
message quotes the fixture literal, because gitleaks scans the tracked
worktree with `--no-git` and would flag the quotation itself.

**The registration surfaces, verified.** `config.Threshold` is
`{Key, Default, Doc}` with `ParseInt` as the only parser
(`internal/config/thresholds.go:13-59`); `Register` is called "from an
explicit setup call rather than an init function" (`thresholds.go:24-30`)
and rung 1 registers none (`:19-22`). `ownedWall` holds the four owned
rows plus the pending identity-source row
(`internal/config/keys.go:19-25`), and `reservedClasses` is the table
shape the integration class sits beside (`keys.go:44-142`).

#### 7c. The 3d plan as the process exemplar

**Execution model**
(`docs/superpowers/plans/2026-09-05-rung-3d-verdicts-gate.md:5197-5273`).
Workflow mode per the standing grant, the pass-execute chain per task,
**strictly sequential**, one code lane in this tree. Each task runs
`cairn-implementer` (which commits its own task's specific files once
`bash scripts/check.sh` is green), then `diff-reviewer`, then the
gate, with one re-dispatch on `fix` and a second `fix` as the
conductor's decision. The conductor dispatches `code-simplifier` over
each accepted task's changed Go after review and before the close, and
owns every ledger edit, the lane registry, the contested-surface rows,
and the batch commit. Runaway guard armed at launch. Subagents never
dispatch agents.

Models: `sonnet` by default, with **named** upshifts to `opus` and the
reason spelled out per task (`:5215-5233`) rather than a blanket
policy, plus a named runner-up the conductor upshifts on a `fix`
verdict (`:5235-5238`). Review scales to rung risk, with one named
exception this pass took and 3c declined (`:5240-5248`). Dependencies
and order are stated exactly, including which tasks are independent
and the statement that **they are still run sequentially** because one
tree runs one executor and the module gate runs `./...` regardless
(`:5250-5266`). A split point is named rather than the scope packed,
with the recommendation against taking it and a numeric trigger for
the conductor to take it anyway (`:5268-5285`).

**Budget** (`:363-421`). A ceiling stated as a **starting** figure with
a named re-derivation point, a basis read off HISTORY per prior pass
rather than asserted (3a ≈2.84M, 3b ≈4.82M with a 270K-per-task rate,
3c ≈4.90M broken into seven line items), an itemized forecast, and the
honest sentence when the arithmetic exceeds the start: "the ceiling is
the stop rule's yardstick, not a forecast" (`:390-397`). Flag at 80%
of the live ceiling, routed to the conductor with the window cap named
rather than half-quoted (`:399-408`). Checkpoint: STATUS after every
task. Phase ceiling carried as a reservation, explicitly "not a figure
any spend is currently measured against" (`:412-421`).

4a's own figures, from the spec's Band and ceiling section: **5M
proposed for 4a**, 1.5M for 4b, each re-derived from 4a's actual
report, on the 2c precedent. Scoring at close is tokens against the
ceiling plus attended time as two counts, with rung 4's expected owner
touchpoints being **none** inside 4a.

#### 7d. `CLAUDE.md`'s three governing sections

**Build process.** Rung 4 sits in the evidence phase (rungs 2, 3, 4).
Inside a phase there are no clears and no re-brainstorm. The
conducting loop per lane: execute via the pass-execute chain with the
guard armed, prep the next rung's plan against a tree no other
executor is writing, land queued repo edits in one batch after the
executor finishes, plans just-in-time. The lane closes through the
merge ritual. **Subagents cannot dispatch agents**, so the conductor
owns the `code-simplifier` dispatch.

**Gate.** `bash scripts/check.sh` before every commit: `make check`
(vet, golangci-lint, test, tidy-check) over `./...`, shellcheck over
`scripts/*`, ruff `check` and `format --check` over every tracked
`*.py`, gitleaks `--no-git` over the tracked worktree, vale over Go
comment prose with `_salvage/` excluded, `docker compose config`, and
a pattern-filtered `systemd-analyze verify`. A check either gates or
does not exist. Never commit red. The `web/` leg rejoins at rung 7, so
4a adds nothing there.

**Stack rules.** `go-conventions` before any Go. Module
`github.com/glw907/dubplate`. Wrapped errors with sentinels via
`errors.Is`, `ctx` first parameter, every subprocess under
`exec.CommandContext` (4a runs none), atomic writes via
`internal/atomicfile`, table-driven tests under the
assertion-discipline rules. `python-conventions` is not in scope for
4a; `vps-conventions` is 4b's except for the state-directory placement
rule below.

#### 7e. `go-conventions`, the parts 4a leans on

- **File writes**: all mutations atomic (temp, sync, rename),
  `internal/atomicfile.WriteFile(path, data, perm)` named as the
  canonical signature (`~/.claude/skills/go-conventions/SKILL.md:298-310`).
  Directories 0755, plain files 0644. The salvage implementation is 42
  lines and already matches
  (`_salvage/internal/atomicfile/atomicfile.go:15-42`), including the
  chmod-after-close step that survives umask.
- **Assertion discipline** (`SKILL.md:323-367`): for every test, "what
  would have to be wrong about the code under test for this test to
  fail?" Named anti-patterns that bite this pass are silent-success
  fakes, self-derived expectations, and **goldens as last resort**.
  The last one matters, because the natural way to write the
  end-to-end observe assertion is a golden `observation.json`; the
  rule says assert on parsed structure instead, reserving goldens for
  surfaces whose value is the literal bytes.
- **Anti-patterns** (`SKILL.md:42-80`): no interface until a second
  consumer exists (the FLAC reader is one concrete type), no
  functional options, no `context.Context` in a struct, no
  package-level `init()`, no vague package names, guard clauses over
  nesting, `var t []T` over `[]T{}` unless the JSON boundary needs the
  distinction, which for the report's per-reason counts it may.
- **Unit tests** (`SKILL.md:314-321`): same package as source,
  table-driven, sub-tests, no third-party assertion libraries, error
  paths as first-class tests.

#### 7f. `vps-conventions`, the state-directory placement rule only

Two clauses, both at
`~/.claude/skills/vps-conventions/SKILL.md:39-45`. A `ReadWritePaths=`
entry **must exist on the host before the unit starts** or startup
fails with "Failed to set up mount namespacing"; the unit's own
script cannot create it, so the deploy step pre-creates it and the
deploy comment names every unit that depends on it. New directories
under `/var/lib` on an SELinux-enforcing box inherit `var_lib_t`, fine
for unconfined services, to be **verified once on the first real
deploy rather than assumed**. Both are 4b's to execute. What 4a owes
is the shape they constrain: the store's root must be a single
directory the unit can be granted whole (`<state dir>/shadow/` with
`passes/`, `observations/<obs-id>/`, `pending/`, `resolved/`,
`started-at`, `consumed-reports.json`), with no write anywhere else,
so that 4b's `ReadWritePaths=` is one path and the argument is static
rather than assumed (`SKILL.md`, Gate).

---

## (d) Cross-section consistency and open questions

### Deliverable-to-seam mapping

Section (a)'s A.1.1-A.1.17 name seventeen deliverables (A.1.18 is a
scope boundary, not a deliverable). Every one maps to a landed seam
section (b) states, or is flagged below as unmapped.

| Deliverable | Seam (b) confirms |
| --- | --- |
| A.1.1 `internal/atomicfile` | (b)9: exact file, line count, test shape |
| A.1.2 `QuietImportTunables` | (b)2: `gradeTunables`/`withImportKeys` split, value equality today, the drift-guard recommendation |
| A.1.3 config integration class, thresholds, `.gitleaksignore` | (b)7 confirms the mechanism **and surfaces a gap the spec does not close**: `rejectNamespace` exempts only `ownedNow`, so the two threshold rows are unaccepted under any reading the spec states (open question 1, below) |
| A.1.4 FLAC reader and mapping | (b)9 confirms no salvage FLAC code exists and gives the mapping's evidence; (c)6d supplies the fixture-builder placement question |
| A.1.5 The capture | (c)4-5 confirm the inbox layout and `group_albums: yes`'s estate reason |
| A.1.6 The reconcile step | (c)1-3 confirm `RunReport`/`AlbumOutcome` shape, the tag log, and the review tree exactly |
| A.1.7 The match step | (b)1,3: `tag.Album`'s signature and construction inputs, `ClientConfig.HTTP` as the transport seam |
| A.1.8 The store | (b)5: the `observations/` directory shape matches the corpus case-directory convention; no direct (b) confirmation of `passes/`'s own shape beyond the `Budget` accessors |
| A.1.9 Recording transport, `DefaultCompat` re-score | (b)4: `gradeCaseDefault`'s five steps, verified line for line |
| A.1.10 Divergence classification | (b)1,3: `errors.Is` requirement, the three sentinels, `ErrNoCandidates` |
| A.1.11 The two rows | (b)5: `AdjudicationSubject`/`AdjudicationEntry` field counts and the pair rule |
| A.1.12 The two schema additions | (b)5: `DisallowUnknownFields`, `omitempty` safety, **and a fixture-shape inconsistency** (open question 14, below) |
| A.1.13 The marshalling boundary | (b)6: `TestEngineNeverImportsHarness`, confirmed to descend into nested packages |
| A.1.14 The subcommand's three verbs | (b)8: the shape precedent and figures; A.4.1's stamp-seam gap is unresolved by either section |
| A.1.15 The report | (b)8: `printRow`'s pinned form is the only direct seam; the row list itself has no (b) counterpart, since nothing landed computes it yet |
| A.1.16 `reevaluate` | (b)3,4: `Unmetered()`, `gradeClient`'s precedent |
| A.1.17 The exit gate | (c)6a-6c: the stub's three disqualifying reasons and the `ErrBudgetExhausted`/`ErrLimiterClosed` mutual exclusion, neither of which section (a) states as bluntly |

Two deliverables surface a genuine seam gap rather than a mere
absence of prior art: A.1.3 (the threshold-acceptance path) and A.1.15
(the report has no landed computation to seam against, which is
expected of a rung that builds it for the first time and not a
finding).

### Testdata-to-deliverable mapping

Section (c)'s testdata requirements each serve one or more of section
(a)'s deliverables:

| Testdata requirement (c) | Deliverable served |
| --- | --- |
| 6a, the transport (`stub.Server` unusable, `httptest` precedent) | A.1.9 (recording transport), A.1.17 (exit gate) |
| 6b, the eight album directories | A.1.5 (capture), A.1.6 (reconcile), A.1.10 (classification), A.1.17 (exit gate) |
| 6c, the `ErrLimiterClosed`/`ErrBudgetExhausted` exclusivity finding | A.1.10 (classification), A.1.17 (exit gate) |
| 6d, the FLAC fixture builder's placement | A.1.4 (FLAC reader), A.1.17 (exit gate) |

No testdata requirement in (c) lacks a deliverable to serve, and no
deliverable that needs testdata (A.1.4, A.1.5, A.1.6, A.1.9, A.1.10,
A.1.17) lacks a (c) section addressing it.

### Figures restated once

Every figure appearing in more than one section agrees across all of
them; none is restated below more than the single time needed to
settle it. `maxAttempts = 7` (`client.go:42`, cited a-A.1.10, a-A.2,
b-3, b-11). `SearchLimit` default 5 (`tunables.go:309`, a-A.1.2, b-2,
b-11). `DefaultInterval = time.Second` and `reservationInterval = 4`
(`limiter.go:62`, `:43`, a-Ruling 5, b-3, b-11). The quiescence window,
ten minutes (`_salvage/internal/config/config.go:152`, a-A.2, c-1).
Run-report retention, 30 files at `*:0/15` with 300s jitter, roughly
seven and a half hours (`config.go:172`, `importer.go:389`,
`dubplate-import.timer:10-11`, a-A.1.6, a-A.2, c-1). Review-tree
retention, 60 days (`config.go:155`, a-A.1.6, a-A.2, c-3). `atomicfile`,
42 lines, one function (`_salvage/internal/atomicfile/atomicfile.go`,
a-A.1.1, a-A.7, b-9, c-7e). The FLAC template, 42 bytes with its exact
framing (`gen_template_flac.py:64-75`, a-A.1.4, b-9, b-11). `ItemFields`,
19 fields (`info.go:124-152`, a-A.1.4, b-2, b-11). `AdjudicationEntry`,
10 fields plus `DefectFree`, 5 non-empty prose checks with `citation`
excluded (`adjudication.go:46-67`, `:176-190`, a-A.1.11, a-A.1.12, b-5,
b-11). `.gitleaksignore`'s pinned line, 66 of 123
(`.gitleaksignore:1`, a-A.1.3, a-A.2, b-7, b-11, c-7b). The doctor's
`"%-32s %s"` row form (`doctor.go:64-67`, a-A.1.15, b-8, b-11). `go.mod`
requiring cobra alone (`go.mod:5`, b-8, b-11). 4a's 5M ceiling against
4b's 1.5M (a-A.1.17 pass-split table, c-7c).

### Combined open-question list, deduplicated

Twenty-five raw questions across the three sections collapse to twenty
after merging exact duplicates. Each item below carries every section
that raised it.

1. **[(a) open-question 2; (b) open-question 1]** Do
   `DUBPLATE_SHADOW_FLOOR` and `DUBPLATE_SHADOW_BUDGET` join the
   integration class's `Keys`, or take a separate acceptance path,
   given `rejectNamespace` exempts only `ownedNow` and knows nothing
   of `registeredThresholds`?
2. **[(a) open-question 3]** How do the six shadow config values
   (paths, floor, budget) reach the runner, given `config.Config` has
   no home for them and the package doc calls itself the sole
   environment reader?
3. **[(a) open-question 1]** What construct carries the matcher stamp
   into `shadow observe`'s refusal check for 4a's own exit gate, given
   4b owns the `-ldflags` change that would set it for real?
4. **[(a) open-question 4]** Is a partial skip (`OutcomeReview` with
   "left in staging after import") a fourth named `unreconciled`
   reason, since mapping it to a clean oracle skip would fabricate a
   divergence?
5. **[(a) open-question 5]** Is the observation `id` the capture's
   file-set digest itself, or a separately generated identifier?
6. **[(a) open-question 6]** Does `consumed-reports.json` detect a gap
   by mtime ordering or by reading `StartedAt`/`EndedAt` out of each
   report, since `RunID` carries no order?
7. **[(a) open-question 7; (c) open-question 1]** Can the exit gate's
   three deferring error branches come from one end-to-end `observe`
   run, given `ErrBudgetExhausted` and `ErrLimiterClosed` are mutually
   exclusive within one pass's `Budget`?
8. **[(a) open-question 8; (b) open-question 5]** What does the
   shadow's `case.json` carry for `Profile` and `EvalDate`, and must
   the written file be byte-identical to what `cassette.Case.Write`
   produces (which is not atomic) or only `LoadCase`-valid?
9. **[(a) open-question 9; (c) open-question 5]** How is a basename
   resolved against beets' own renaming and collision behavior: the
   tag log's staging-path skip lines, and the review tree's
   `-HHMMSS`/`.2`-suffixed directories?
10. **[(a) open-question 10]** Is `spec:1608-1609`'s "(ruling 7)" a
    stale cross-reference that means ruling 6, the deploy sitting?
11. **[(a) open-question 11]** What defines "the first successful
    pass" that writes `started-at`, given a pass that captures and
    errors on nothing is not obviously either?
12. **[(a) open-question 12; (b) open-question 3]** Should
    `config.Register`'s call site guard against double registration
    (for example `sync.OnceFunc`), given `newRootCmd()` is already
    called from `main_test.go` as well as from `main()`?
13. **[(b) open-question 2]** Which rows land in
    `env/dubplate-server.env.template` and what literal value enables
    the shadow, given the spec's own table (four rows) and key list
    (six rows) disagree and the template table omits
    `DUBPLATE_SHADOW_ENABLED` entirely?
14. **[(b) open-question 4]** Do the two new `AdjudicationEntry`
    checks (`MatcherCommit`, `MatcherConfigHash`) get hand-authored
    `testdata` fixtures as the spec asks, or inline Go literal
    variants, which is the package's own landed precedent for a
    per-row rule?
15. **[(b) open-question 6]** Does the incumbent's `report-*.json`
    decode struct sit engine-side with the reconcile logic, or in
    `cmd/dubplate-server` beside the other marshalling, given
    `internal/engine/shadow` may not import the salvage-tier
    `RunReport`/`AlbumOutcome` types and the incumbent retires at rung
    16?
16. **[(b) open-question 7]** Is it settled, in the plan, that
    `internal/engine/shadow` takes its paths, floor, cap, and
    User-Agent as parameters rather than importing `internal/config`,
    given no arch test forbids the import and the rule is presently
    only prose convention?
17. **[(c) open-question 2]** Where does the end-to-end `observe`
    test live, and how does it get FLAC bytes, given
    `internal/engine/shadow`'s test files may not import
    `internal/harness` and the marshalling the exit gate asserts on
    lives in `cmd/dubplate-server`?
18. **[(c) open-question 3]** Does the plan state the capture walk's
    iteration order (a lexical `os.ReadDir` walk, matching the
    incumbent's own `listContributorDirs`), since the
    budget-exhaustion fixture's correctness depends on which album the
    pass reaches first?
19. **[(c) open-question 4]** Is an archive upload's invisibility to
    the capture step (a `.zip` never seen, so its eventual run-report
    basename is never captured) a declared coverage gap for 4a's
    close, or does the capture step gain an archive branch?
20. **[(c) open-question 6]** Should the 4a exit gate also require one
    gate-rejected exclusion and one `unreconciled` row of each named
    reason, since both counters print in the report and a counter
    reading zero forever is indistinguishable from a broken one?

This inventory does not answer any of the twenty; each is the plan's
to settle.
