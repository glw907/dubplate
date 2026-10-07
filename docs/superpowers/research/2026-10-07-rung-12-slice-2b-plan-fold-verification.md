# Rung 12 slice 2b plan fold: verification read

Target: `docs/superpowers/plans/2026-10-07-rung-12-slice-2b-oracle-harness.md`, revision 1.1 at `64d9d04` (revision 1
at `0ae2b54`), against the fold record `2026-10-07-rung-12-slice-2b-plan-fold.md` and the three lens reviews. A
fresh-context read: the reader had no part in the fold. Line numbers are the plan's at `64d9d04` unless a file is named.

**Verdict.** The fold is sound in its main lines and one gap from clean. Every blocker-free major closed at its cited
location except C6, whose fallback covers only a failure found before Task 9. Every new tool or library claim
reproduces. Ruling 1 can be ruled on the evidence given. The refusals hold, and one of them gives the wrong reason.

**Counts:** 0 blockers, 1 major, 8 minors. No owner fork.

## How it was checked

Probes ran in `/tmp/claude-1000/v2b/` only. No container started, and the box was never contacted.

- **Homebrew.** `brew pin --help`; `brew info libxml2` ("Installed (as dependency)"); `brew list --pinned` (empty);
  `brew uses --installed libxml2`. The local Homebrew source: `cmd/upgrade.rb:423-424`, `upgrade.rb:335-337`,
  `formula_installer.rb:485-509`, `dependency.rb:127-131`, and `install/check.rb:188-197`.
- **uupd.** `ublue-os/uupd` `drv/brew/brew.go` fetched: `:50` runs `brew update` and `:63` runs `brew upgrade -y`. The
  journal for 2026-10-07 shows `"msg":"Brew module" ... "Enabled":true`. `systemctl is-enabled` reports `uupd.timer`
  enabled and `brew-upgrade.timer` disabled.
- **libxml2.** `doc/xmllint.xml` and `NEWS` fetched at `v2.15.4`. `:357` (`--nonet`) reads "Has no effect with libxml2
  2.15 or later", `NEWS:229` reads "The built-in HTTP client and support for LZMA compression were removed", and
  `:1000-1002` gives exit 3 as "Validation error". A live probe ran the pinned `xmllint` on a body whose DOCTYPE and
  external entity name `http://127.0.0.1:18765/`, with a listener on that port. The listener received no connection,
  and xmllint printed `failed to load ... No such file or directory`. `xmllint --version` writes to stderr, and stdout
  stays empty.
- **Go 1.27.1.** The cited lines check out: `marshal.go:225` (`escapeText(p, t, false)`), `:744` (`if
  start.Name.Space != ""`), and `xml.go:1922`. `go doc` matches the quotes for `os/exec.ExitError`,
  `net/url.ParseQuery`, and `URL.Query`.
- **Navidrome.** The website at `b0fe572` matches the cited lines `:26`, `:30`, `:46-57`, `:62`, and `:150`.
- **pass-execute.** `pass-execute.js:256` (the `docs` bar), `:267` (`tool` defaults to `gateLane: "light"`), `:359`
  (the lane resolution), and the per-task `passClass` and `gate` at `:27-37`.
- **The pinned OpenAPI bundle.** It holds 87 paths. The `GetAlbumResponse` `oneOf`/`allOf` shape and `required:
  ["album"]` match. `Indexes` requires `ignoredArticles` and `lastModified` and leaves `additionalProperties` unset.
  The payload member and its required flag were computed for every path.
- **R27 and R13, recomputed from the records.** Section 3 names 51 endpoints, so 52 are served and 35 are not. The
  slice list covers all 52. Every L and A row whose "Test owed" cell carries no `Slice N` token appears in R13's list,
  and each listed slice agrees with R27. That holds for L-06, L-17, L-27, L-34, and A-04.
- **The 2a code.** `seed-ledger.json` has 27 rows: 20 `server-*` rows and 2 `song/created` rows. The plan also cites
  `ledger.go:141-230`, `:305-308`, `grade.go:101`, `:118`, `grade_test.go:116-145`, `session.go:173`, `:195`,
  `:331-394`, `stack.go:137-154`, and `oracleup.go`.

## Question 1: did each blocker and major close?

There were no blockers. The majors:

| Finding | Closed? | Where |
|---|---|---|
| C1, M4, DR3 (joint) | Yes | R27 `:353-373`; Task 5 values test and named contract-only list `:670-674`; mutations `:689-691` |
| C2 | Yes | R13's rule and list `:289-306`; Task 6 slice check `:736-737` |
| C3, M9 (joint) | Yes, as M9's fix | R20 `:333-336`; Task 4 `:611-613`, singleton proof `:630-631` |
| C4, M12 (joint) | Yes | Task 3 residual drop `:533-537`; behind-a-residual proof `:578-580`; mutation `:594` |
| C5 | Yes in text. The named fixture cannot load as written (V2) | Task 5 `:677-685`; Task 4 twin `:626-629` |
| C6 | **Partly** (V1) | Task 10 gonic-only form `:913-916`; nothing in Task 9 |
| M1 | Yes | Quotes `:248`; runner `:545`; 64 KiB proof `:589-590`; mutation `:596` |
| M2 | Yes, with a side effect left unquoted (V4) | Quotes `:240`; R18 `:320-325`; lane launch `:409-412`; Task 12 `:986-988`, `:1000-1001` |
| M3 | Yes | Task 6 `:695-707` |
| M5 | Yes | Quotes `:243-244`; R20 `:330-333`; Task 3 `:538`, `:549-555`; proofs `:581-583` |
| M6 | Yes, through the conductor | `:836-844`; fake routes `:879-881` |
| DR1 | Yes | Quotes `:236`; binding `:526-532`; proofs `:585-586`; mutation `:597` |
| DR2 | Yes | R28 `:374-377`; Task 6 accounting and fidelity `:730-735`; activation proofs `:743-745` |

## Question 2: contradictions and build order

The task order builds. No task consumes what a later task makes. Task 8 records through Task 7's writer scan and
grades XML through Task 4. Task 6's slice recomputation reads Task 5's committed assignment. The conductor's Playwright
step needs only 2a's `harness oracle-up`, which already prints the credentials the login needs (`oracleup.go`, its
doc comment). Two defects remain. The fold introduced one directly (V3). The fallback text leaves the other (V1).

## Question 3: new claims checked against the source

The pass-execute and xmllint-stderr claims were checked above. These five new claims also reproduce:

- the `brew pin` help sentence;
- `uupd` running `brew upgrade -y` daily, and that `brew upgrade` skips a pinned formula (`cmd/upgrade.rb:423-424`);
- `--nonet` being moot on 2.15.4, both from the docs and from the live probe;
- the encoder source lines behind the written-newline line table;
- `brew install` marking an installed dependency as installed on request (`install/check.rb:193-196`).

One quote is selective. The `brew pin --help` text the plan cites goes on: "Note: Other packages which depend on newer
versions of a pinned formula might not install or run correctly." V4 covers it.

## Question 4: can Geoff rule on Ruling 1?

Yes. The fork is clear and the recommendation follows from it. Two of its premises are off, and both lean toward the
recommendation, so neither changes the ruling (V5).

## Question 5: are the refusals sound?

- **DR8 (whole).** Sound. The docs and the live probe both show the pinned build opens no connection.
- **C3 (the `maxOccurs` paths).** Sound. A singleton miss surfaces as an unledgered finding, and a missing suppression
  always reads loud. A per-ordinal list flag would split R20's one path grammar between the two tiers.
- **C6 (the admin grant).** Sound. The multi-library page says at `:30` that admins see every library, and 2a's
  stack creates the Navidrome user as an admin (`ND_DEVAUTOCREATEADMINPASSWORD`, `stack.go:144`).
- **C10 (the regex fallback).** Sound for the query half. `ParseQuery` returns the valid parameters beside its error,
  and an error counts as a finding. The `url.Parse` half of C10 stays open (V8).
- **C14 (the relabel).** Sound. The `docs` bar at `pass-execute.js:256` says what the plan quotes.
- **M5 (its own writer).** Sound. No oracle body and no Go encoder writes a prefixed element, and a prefixed body is
  one unsupported violation, never a pass.
- **M6 (the implementer script).** The outcome is acceptable, but the stated reason does not answer M6 (V6).

## Findings

### V1 (major): the gonic-only fallback only fires before Task 9, and Task 9 has no degraded form

**Location.** `:183-185`, `:842-844`, `:858-866`, `:870-871`, and `:913-916`.

**Defect.** The unattended-run rulings let the conductor rule Navidrome's second library unmeasured if the request
"cannot be made or replayed" (`:183-185`, `:843-844`). The plan never says when a replay is tried. The conductor's
step only observes and records. The first replay therefore happens in Task 9's live evidence (`:870-871`). Task 10's
gonic-only form fires "only if Navidrome's second library was ruled unmeasured before Task 9" (`:913`). Task 9 states
only the two-library form: Navidrome creates each later collection by replay (`:862-864`), and readiness on both
oracles needs "one folder per collection" and a `getScanStatus` `count` equal to the total across collections
(`:865-866`).

The plan never states Navidrome's `getScanStatus` count across several libraries. Nothing quoted covers it, and
slice 1 created its second library nested inside the first (`[servers]` F2). So a replay that fails live, or a
readiness count that Navidrome reports per library, leaves Task 9's live evidence unmeetable. That leads to `fix`,
then a second `fix`, then `escalate`, and the 10-hour run stops in segment C. C6 named this defect. The fold closed
it for one trigger only. The same gap undercuts Ruling 1's premise that each live task now carries a written
fallback (`:89-91`).

**Proposed fold.**
- The conductor's step also replays what it recorded: a fresh `harness oracle-up` gets the login and the creation
  over HTTP with `curl`. The step then quotes `getMusicFolders` and `getScanStatus` from Navidrome with two libraries,
  and fixes the readiness rule from that quote.
- Task 9 gains a one-paragraph degraded form. Navidrome mounts the first collection only, and readiness checks each
  oracle against its own collection count.
- Task 10's trigger reads "ruled unmeasured before or during Task 9".

### V2 (minor): Task 5's three-source fixture cannot pass the session loader as written

**Location.** `:677-685`, against `session.go:173`, `:195`, `:370-375`, and `grade.go:101`.

**Defect.** `session.Write` and `session.Load` both run `validate`. It refuses any response whose body's
`serverVersion` differs from its source's single metadata `ServerVersion` ("body serverVersion %q, metadata %q").
The fixture state C5's fold names gives the `dubplate` source gonic's `getAlbum` and `getIndexes` bodies and
Navidrome's `getMusicFolders` body. Those bodies carry two different `serverVersion` values under one source, so the
proof's session refuses before any grade runs. The workaround is easy, but the plan's checklist asks that each proof
name a fixture state that can exist. As written, this one costs a fix round.

**Proposed fold.** The `dubplate` source's bodies carry one rewritten `serverVersion`, set in its source metadata,
and the proof says so. Every other difference the proof claims survives the rewrite.

### V3 (minor): the fold made `grade`'s L-07 finding unreachable

**Location.** `:769-777`, against `grade.go:101`.

**Defect.** DR4's fold has the session loader refuse any committed session that carries a credential URL
(`:775-776`). `Grade` reads every session through `session.Load` (`grade.go:101`). So the L-07 `contract` finding
"for that source at that leaf's path" (`:769-773`, "`grade`'s L-07 finding calls the same function", `:777`) can
never fire on a real grade. It runs only in tests. That leaves a test-only seam the close's architecture read will
flag, and an implementer builds a finding kind no session can produce. In revision 1, a hand-made or future
recorder's session could still reach it.

**Proposed fold.** Drop the grade-side finding. The writer's and loader's refusal is the L-07 check. The credential
proof then calls the `session` scan directly, and its mutations stay as listed.

### V4 (minor): pinning libxml2 stalls its dependents' upgrades, and the plan quotes only the pin's first sentence

**Location.** Quotes `:240`; R18 `:320-325`; lane launch `:409-412`.

**Defect.** The plan quotes only the first sentence of `brew pin --help`. The help text continues: "Note: Other
packages which depend on newer versions of a pinned formula might not install or run correctly."

The Homebrew source shows the mechanism. `formula_installer.rb:485-509` checks every recursive dependency. A pinned
dependency is unsatisfied when it is not the latest version and no minimum version is passed (`dependency.rb:82-84`).
The installer then raises "You must `brew unpin libxml2` as installing <formula> requires the latest version of pinned
dependencies."

Eight installed formulae depend on libxml2: chafa, ffmpeg, gdk-pixbuf, imagemagick, librsvg, libxkbcommon,
shared-mime-info, and xdotool. After the next libxml2 release, each of these that is itself outdated fails in
uupd's daily `brew upgrade -y` until a dubplate sweep unpins and bumps. The workstation's "take every minor and patch"
rule bends for the whole time. The dubplate gate itself stays correct.

**Proposed fold.**
- Quote the note in the quotes table.
- Name the held dependents in the lane launch's Brewfile line and in the close's CLAUDE.md batch.
- End R18's bump path with a `brew upgrade` of those dependents after the re-pin.

### V5 (minor): Ruling 1 understates the split's cost and overstates the fallbacks

**Location.** `:87-98`.

**Defect.** The "about 0.8M more" figure sets the split's summed estimates (6.49M plus 3.34M, which is 9.83M) against
the one-pass ceiling (9.0M). It should set them against the one-pass estimate (8.55M). Like for like, the delta is
about 1.3M, which is the second close.

The fallback claim is overstated too. "Task 11's single-oracle walk" is the plan's ordinary walk, not a fallback.
Task 9 has no fallback (V1).

Both errors favor the recommendation, so Geoff can rule as the text stands.

**Proposed fold.** State the delta as about 1.3M, estimate against estimate. Name Task 9's degraded form (V1) in
place of Task 11's walk.

### V6 (minor): the M6 refusal gives a reason that does not answer M6

**Location.** Fold record, the M6 row; plan `:836`.

**Defect.** M6 never asked for a browser tool. It proposed a headless Playwright script, run once through Bash, with
Playwright 1.63.0 and its Chromium already in `web/node_modules`. `dubplate-implementer` carries Bash and can run
that script. So "implementers carry no browser tool" does not refute the proposal.

The conductor's route is still acceptable for a one-time observation. Two of its costs go unweighed. The conductor
does GUI work inline, against "the conductor is thin." And nothing re-observes the request when R24's next deliberate
Navidrome bump changes the admin API.

**Proposed fold.** Correct the fold record's reason to the real trade-off: a one-time observation with no script to
maintain. Or dispatch the observation to a general-purpose agent that carries Playwright MCP. The plan's outcome
stands either way.

### V7 (minor): Task 6 changes the assignment loader but does not list it

**Location.** `:695-697`, `:663`, and `:721-723`.

**Defect.** Task 5's assignment loader refuses "an unknown member" (`:663`). Task 6 adds the `agreements` section and
its validation, which the states table names as "an agreement with an empty field" (`:723`). That change touches the
loader in `grade/`, but Task 6's Files names only `ledger.go` and the assignment's data file. A `diff-reviewer` may
then flag the necessary loader edit as out of scope. C7's fold fixed the same class of omission.

**Proposed fold.** Add the assignment loader and its tests to Task 6's Files.

### V8 (minor): the credential scan still fails open on a URL that `url.Parse` rejects

**Location.** `:769-773` and `:786-789`.

**Defect.** C10's fold reads the query with `ParseQuery` and counts a parse error as a finding. The plan never says
how a leaf's query is separated from the rest of the URL. The probe shows the gap. `url.Parse` returns `invalid URL
escape "%zz"` for `http://h/a%zz?u=admin&t=x&s=y` before any query is read. "A leaf with no query: no finding" then
reads that leaf as clean. C10 named this half, and the fold record answers only the semicolon half.

**Proposed fold.** For a leaf that begins with a scheme, take the text after the first `?` and pass it to
`ParseQuery` without `url.Parse`. Add one row with a bad path escape.

### V9 (minor): the "any other line is a tool failure" rule misreads one real xmllint output

**Location.** `:546-548` and `:565`.

**Defect.** The probe ran the pinned xmllint on a typed attribute whose value holds a newline. The encoder writes that
newline as `&#xA;`, and the parsed value keeps it. xmllint printed one error over two stderr lines: `t.xml:3: Schemas
validity error : ... '1` and then `2' is not a valid value of the atomic type 'xs:int'.` Under C11's fold rule, the
second line is "any other line" under exit 3, so a body violation becomes a tool failure that ends the grade. The
failure is loud, but it gets the wrong label, and a dubplate attribute can carry such a value.

**Proposed fold.** A line that matches no pattern and directly follows a validity-error line continues that error's
message. A tool failure is only an unmatched line with no error line before it.

## Over-ceremony

Nothing of consequence. The fold removed ceremony in two places: Task 3's dedupe rule and half of Task 11's walks. The
proofs it added are each tied to a measured defect.
