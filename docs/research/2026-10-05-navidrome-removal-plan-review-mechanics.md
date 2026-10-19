# Navidrome removal plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-10-05-navidrome-removal.md` at `7036ebc` (revision 1). Spec:
`docs/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1. Lens: does every task build in the
stated order and leave master green; are the orphan lists, test dispositions, migration re-pin, and rename reach
complete; is the end state idiomatic; is the ceiling plausible.

## Method

A throwaway detached worktree at `7036ebc` under the scratchpad (removed before return). In it:

- Task 2a's core deletion, done for real: `git rm` of `cmd/dubplate-server/shadow.go`, `shadow_test.go`, and
  `internal/engine/shadow/`, then the minimal edits to `main.go`, `doctor_test.go`, `serve_test.go`, and
  `unidecode.go` (`setUnidecodeTransliterator` deleted). Then `go build ./...`, `go vet ./...`,
  `go vet -tags oracle,rendered ./...`, `golangci-lint run ./...`, and `go list -deps ./cmd/dubplate-server`.
  `go test ./internal/config` passed. `go test ./cmd/dubplate-server` is not evidence either way: its doctor
  tests fail with "configuration is incomplete" in a fresh worktree at `7036ebc` as well, before any edit,
  because a fresh tree lacks the workstation prerequisites the gate assumes. Runtime test fallout from the
  deletion was therefore read from source, not run.
- The exported-unused leg's body (`scripts/check.sh:566-611`) run on its own, with the twenty-line cap removed, at
  base and after the scratch deletion, then diffed.
- Source reads and `git grep` for everything else, quoted below.

Results that hold: the shadow deletion builds and vets clean once the plan's named files are edited, under every
build tag (`oracle`, `rendered`). `golangci-lint run ./...` reports `0 issues`. `go list -deps ./cmd/dubplate-server`
then carries no `internal/harness`, `internal/engine/shadow`, or `internal/engine/matcher` package. No env template,
unit, or script names a `DUBPLATE_SHADOW_*`, `DUBPLATE_NAVIDROME_*`, or `DUBPLATE_ACCESS_*` key. Reading 3 holds:
only `backup.go:32` and `proxy.go:25` declare a production `IntegrationClass`, neither with `Optional`. The `0003`
hash appears in one place (`schema_test.go:297`); no committed database fixture carries the old ledger hash. The
username rename's file list matches every `git grep -iE 'streamer[_ -]?user(name)?|streamerUser|streamer-hint'`
hit in code, SQL, web, and config. The deck-coverage gate reads Go outcome values by regex
(`deck-coverage.ts:29`), so renaming the constants and deck keys in one commit (Task 4b) keeps it green.

## Correctness gaps, ranked by consequence

### PM1. Medium. Task 1's prose acceptance cannot see most of the stale comments it exists to catch

`plan:256-258`. The criterion is `go doc -all ./internal/serve | grep -inE '...'`. `go doc -all` prints exported
declarations only. At base it reports about 30 lines; a source grep over the package's non-test files reports 78.
Several stale unexported comments are named neither in the outcome's list (`plan:207-213`) nor reachable by the
criterion:

```
internal/serve/origin.go:9:   prefix routing, so every registered subtree and the fallthrough sit
internal/serve/serve.go:246:  subtree or the fallthrough. TRACE echoes the request back to the
internal/serve/serve.go:255:  the fallthrough accepts, and an Allow header naming a guessed set
internal/serve/serve.go:275:  rather than reach the proxy at the fallthrough.
internal/serve/serve.go:389:  to, registered subtrees and the fallthrough alike; doc.go reserves
internal/serve/serve.go:435:  would silently downgrade every proxied stream rung 12 serves to a
internal/serve/serve.go:474:  terminator or the empty fallthrough is likewise no refusal, and
internal/serve/prefix.go:47:  from the fallthrough.
internal/serve/probe.go:33:   identity spelling the strip removes while no driver header is
```

The blanket outcome ("Every comment in `internal/serve` ... describes the package as it stands") covers them in
intent, but the acceptance a diff-reviewer grades against passes with all of them standing. Task 3's identical
criterion over `internal/auth` and `internal/config` has the same blind spot (`config.go:121`, `:424`,
`registry.go:289` are unexported).

**Fold:** add a source criterion beside the `go doc` one:
`git grep -niE 'front ?exemption|forward-auth|driver|fallthrough|navidrome|streamer' -- 'internal/serve/*.go' ':!*_test.go'`
prints only the named survivors: `strip.go`'s authentik "forward-auth configuration" citation (a source
title), and `origin.go:42-54`'s bearer-bypass "exemption" wording, if kept. Mirror it for Task 3 over
`internal/auth` and `internal/config` (allowing `limit.go:231`'s `fallthrough` keyword and the two retired keys).

### PM2. Medium. Two re-pointed tests can no longer fail against the regression they are named for

`plan:224`, `:229`. After Task 1, both arms of `Prefix.Route` answer the same 404: `Terminator()` is
`http.NotFound` (`prefix.go:131-135`), and `next` is `http.NotFoundHandler()`. The registered patterns are all
`prefix+"/"+name` (`serve.go:104-106`), so a path outside the prefix never matches the mux either.

- `TestOrderPrefixRequiresSegmentBoundary` re-pointed to "owned spy never invoked, status 404". Mutate `Owns` to a
  plain `strings.HasPrefix(path, p.value)`: `/anchorfoo` then routes to `terminated()`, the mux finds no
  `/anchor/foo` pattern, and the terminator answers 404 with the spy untouched. The test still passes.
- `TestUnclaimedOwnedPathReachesTheTerminator` re-pointed to "404, owned spy not invoked". Mutate `Route` to send
  every request to `next`: the answer is still 404 with the spy untouched. The test still passes.

Both break the plan's own fixture rule (`plan:268-269`: "each re-pointed probe asserts its spy was reached, or
asserts a status a 404 cannot produce"). The spec already concedes this shape for `TestEmptyFallthroughAnswers404`
alone (spec:249-250).

**Fold:** delete both shell-level tests, with the reason that the owned-versus-unowned decision is no longer
observable through the composed shell and `prefix_test.go`'s `TestRoute` table proves it with distinct `owned`
and `next` spies. The alternative, keeping them under a name that claims only "a 404", proves nothing a reviewer
should credit. Either way, the dispositions table says which.

### PM3. Medium. The shadow orphan list misses most of what the deletion orphans, and misstates `tagio`

`plan:333-337`, `:529-531`. The exported-unused leg, run uncapped, at base and after the scratch shadow deletion
(127 names, then 137):

```
> internal/engine/matcher/quiet.go        QuietImportTunables
> internal/engine/matcher/translit.go     SetTransliterator
> internal/engine/matcher/verdict.go      ActionApply
> internal/engine/musicbrainz/client.go   DefaultBaseURL
> internal/engine/musicbrainz/limiter.go  Class
> internal/engine/musicbrainz/limiter.go  ClassBackground
> internal/engine/musicbrainz/limiter.go  DefaultInterval
> internal/engine/musicbrainz/limiter.go  ErrBudgetExhausted
> internal/engine/musicbrainz/limiter.go  ErrLimiterClosed
> internal/engine/musicbrainz/limiter.go  Spent
> internal/engine/tag/tag.go              Proposal
> internal/serve/serve.go                 ReadFrom
< internal/engine/shadow/reevaluate.go     ReevaluateResult
< internal/engine/shadow/transport.go      RoundTrip
```

Three consequences:

1. `tagio` is not newly warned. It carries nine names at base and the same nine after; no package imports it today
   (`git grep -l 'engine/tagio"'` returns only the shadow package). The plan's "Its exported names now have no
   importer, so the exported-unused leg warns on them" (`plan:336-337`) and the HISTORY line built on it are
   wrong about cause and timing.
2. `internal/serve` gains `ReadFrom` at Task 2a (the leg is name-based, and `ReadFrom` was matched by a reference in
   the shadow package). Task 1's criterion is checked at Task 1's commit and holds there, but the pass ends with
   the spec's "no hit in `internal/serve`" further from true than Reading 1 records. Reading 1 should name it.
3. `matcher.SetTransliterator` loses its one production caller with `setUnidecodeTransliterator`, and
   `QuietImportTunables` loses its only caller. The plan keeps both silently while pruning `ClassBackground` and
   `ErrUnavailable` under Reading 2's "a consumer re-adds it" logic. That is the inconsistency ruling 6 ("a
   removal reshapes what remains") asks the plan to settle explicitly.

Some of the twelve are consumed by type inference rather than by name (`tag.Proposal` is `tag.Album`'s return,
used by `internal/harness/conform/case.go:401`), so the leg over-reports them, which is why each needs a
disposition rather than a blanket prune.

**Fold:** Task 2b gets a table of the twelve with a disposition each (keep, naming the consumer or the inference;
or prune), and Task 5's HISTORY records the real delta, replacing the `tagio` sentence. Reading 1 adds `ReadFrom`.

### PM4. Medium. Pruning `ClassBackground` leaves the limiter's round-robin machinery hollow

`plan:90-92`, `:333-335`, `:358-359`. The reservation picks "the non-empty classes BELOW the one strict priority
would have served, taken in round-robin" (`limiter.go:30-33`) through a `cursor` field and a modulo walk
(`limiter.go:283-291`). With two classes, the only class below `ClassUserPresent` is `ClassBatch`, so the rotation
rotates through one element: `cursor`, the modulo walk, and `TestReservedTokensRotateThroughTheClassesBelow`
(`limiter_test.go:78`) become structure with nothing to do. The anti-starvation rationale is also argued in
three-class terms ("a saturated ClassUserPresent queue plus a queued ClassBackground waiter starves ClassBatch
outright", `limiter.go:35-37`). The plan says only "restated for two classes", which leaves the implementer either
to keep a hollow cursor or to redesign concurrency code without a named outcome.

The prune is also uneven: `ClassUserPresent` and `ClassBatch` were already consumerless at base (both in the base
leg list), and after the shadow goes the whole `Limiter` has no non-test caller. Pruning one class of three, when
none of the three has a caller, buys no consumer-honesty.

The spec sentence is ambiguous on this: "`musicbrainz.ClassBackground` and `ErrUnavailable`'s shadow
justification stay only if rung 9's spec names a consumer. Otherwise they are pruned" (spec:265-266) reads either as
pruning the symbols or as pruning their shadow justification.

**Fold (recommended):** read the sentence as pruning the justification. Keep `ClassBackground` and
`ErrUnavailable`, and rewrite their comments to state the class and the sentinel without the shadow ("deferred
work", "a caller that defers rather than fails"). Drop the limiter mutation from Task 2b's acceptance. That
removes a concurrency rewrite from a removal pass. **If the symbols are pruned anyway:** Task 2b's outcome must
name collapsing `cursor` and `nextBelow` to "the next class below", deleting or re-scoping the rotation test, and
a named mutation for the two-class reservation.

### PM5. Low-medium. Task 1's exported-unused criterion cannot be read off the gate

`plan:262`, `scripts/check.sh:617-618`. The leg prints `head -n 20` of a sorted list. At base the list holds 127
names and the `internal/serve` entries sit at lines 119 to 127, so the gate never shows them. Reading 1's numbers
came from an isolated run the task does not tell the implementer how to reproduce.

**Fold:** name the command in the acceptance, for example run the leg body with the cap lifted, or add an
`internal/serve`-scoped `grep` of the leg's work file. Cite the expected list literally (the eight survivors in
Reading 1).

### PM6. Low. Test-helper fallout is missing from the dispositions table, so the first gate run fails lint

`plan:232`. `.golangci.yml` enables `unused` and `unparam`, and golangci-lint lints test files by default.

- `cookieValue` (`serve_test.go:289`) is used only by `TestSessionCookieOtherCookiesSurviveAtFallthrough`;
  `cookieLines` (`:475`) only by deleted cookie-removal tests. Both become unused.
- `newTestShell(t, sessionCookie string)` (`serve_test.go:22`) passes `sessionCookie` to `NewStrip`. Once
  `NewStrip` drops the parameter, the helper's parameter is dead, and the table says "Signature only", which
  reads as keeping it.
- `flushingHandler` stays used (the owned half keeps it).

The gate catches all of this, so the cost is one wasted gate cycle, not a wrong commit. **Fold:** the table gains
"`cookieValue`, `cookieLines`: delete" and "`newTestShell`: drop the `sessionCookie` parameter and the driver-header
paragraph of its doc".

### PM7. Low. Reading 7 misses the test file, so Task 1's `cmd` prose criterion fails as written

`plan:106-108`, `:257-258`. The criterion allows only "rung and history citations in tests". `serve_test.go` carries
"driver" in a different sense: `:73` ("the sweep driver's context"), `:562`, `:588-605`, including the test name
`TestNewServeShell_SweepDriverStopsWithTheGivenContext`. None is a rung or history citation.

**Fold:** extend Reading 7 to `serve_test.go` (rename the test and its comments in the same rephrase), or scope
the `cmd` criterion to non-test files.

### PM8. Low. Task 2a's unknown-key test can be written vacuously

`plan:309-310`. `internal/config`'s own registry never knew the shadow keys (the class was registered from
`cmd/dubplate-server/shadow.go`). A test placed in `internal/config` against `fixtureRegistry()` passes at the parent
commit. **Fold:** name the location (`cmd/dubplate-server`, against `productionRegistry()`), and require the report
to show it failing at the parent commit, where the flag is accepted and the error names a missing Required key
instead.

### PM9. Low. Two `Optional` sites fall outside Task 2a's cited lines

`plan:281-283`, `:293-295`. `config.go:21-28`'s package doc still describes integration "Optional rows" and cites
"the shadow and backup classes, for example". Task 2a removes `Optional` without listing `config.go`; Task 3 edits
`config.go:15-30` for other reasons and does not name this. The registry's pairing skip has a second half in
`AddThreshold` (`registry.go:285-300`, "that pairing is the one exemption AddThreshold makes"), outside the cited
`:197-230`. The haiku pre-flight may catch the line drift, but not an unnamed outcome. **Fold:** Task 2a names both.

### PM10. Low. Task 4b's acceptance grep depends on Task 4a, which the plan calls independent

`plan:52-53`, `:488-489`. The grep runs over `internal`, which includes `internal/engine/store`'s seven
`schema_test.go` hits and the four SQL files until 4a lands. The independence note lets the conductor "reorder 4a
around a stalled task"; 4a's precondition waits on the document pass's archive chore, so a stall is the expected
case. **Fold:** scope 4b's grep with `':!internal/engine/store'`, or mark 4b as after 4a.

### PM11. Low. `BACKLOG.md:43` names the retired flag

"`--streamer-username` as required, though `MarkFlagRequired` enforces both". **Fold:** Task 4b or Task 5 updates
the entry.

### PM12. Low. The identity set should leave `Strip`

`plan:197-200`. With the driver header gone, `identityKeys` has no input, so `Strip.identity` (`strip.go:276-278`)
becomes a per-instance copy of a package constant, and `deleteIdentity`, `scrubTrailerOnRead`, and
`trailerScrubbingBody.keys` thread a map that never varies. "Reshape to a set derived once" admits both shapes.
**Fold:** state the end shape: one package-level set derived at init (a `var` initialized from `identityExtras`),
read directly; `Strip.identity` and the `keys` parameters and field go. That is the stdlib shape for a fixed
lookup table and removes four now-dead parameters.

## Over-ceremony, ranked by cost

### PO1. Four architecture readers grade packages whose non-test change is one comment

Reading 6, `plan:104-105`, `:513-516`. `internal/engine/tag`, `tagio`, `matcher`, and `internal/engine/store`
(non-test change: `leaf.go:14`'s comment; the `0003` edit is SQL) each take a `go-architecture-reader`. A reader
grades the whole package, unchanged by a comment, so each returns the package's standing debt, which the plan
already routes to the cleanup inventory. At a typical reader dispatch that is several hundred thousand tokens for
no removal-attributable finding. **Fold:** readers for packages with a non-comment non-test change only:
`internal/serve`, `cmd/dubplate-server`, `internal/config`, `internal/auth`, `internal/harness/analysis`, and
`internal/engine/musicbrainz` if PM4's prune stands.

### PO2. The limiter rewrite itself

If PM4's recommended reading is taken, Task 2b loses its one piece of novel concurrency work and its mutation, and
becomes a comment and `analysis` task. That is the larger saving.

## Ceiling

8.0M for seven implementer tasks, seven Opus diff reviews, about seven full heavy-lane gates plus re-issues, a
simplifier with its review, the security, Svelte, and a11y reviewers, ten architecture readers, and a fold. Task 1
is the risk: about 30 test dispositions, comment rewrites across eight files, and the vacuity traps above; one
`fix` round there is likely. The estimate is plausible but tight with ten readers. PO1 and PO2 together bring it
back to comfortable headroom without cutting a correctness check.

## What holds

- Ordering: Task 1 before 2a (`cmd/dubplate-server/serve.go`, `serve_test.go`), 2a before 2b and 3
  (`doctor_test.go`), and 4b after 1 and 3 (`internal/auth`) are the real file dependencies; each commit can be
  green in that order.
- `http.ServeContent` reaches `io.CopyN(w, sendContent, sendSize)` (`net/http/fs.go:430`, Go 1.27.1), and
  `*io.LimitedReader` defines `Read` only (`io/io.go:472`), so `statusCapture.ReadFrom`'s restated rationale is
  sound, and deleting it fails the re-pointed `TestReadFromReachesRealWriter` (`plainReader` hides `WriteTo`).
- The `0003` edit: `migrate.go:237-238` refuses any ledger row whose sum differs, and `TestSchemaMigrationHashPins`
  is the only pin. The case-folded duplicate row's mutation (drop `COLLATE NOCASE`) fails both it and the hash pin.
- `corpus/analysis/adjudications.json` carries no `album`, `release_mbid`, `matcher_commit`, or
  `matcher_config_hash` key (each `grep -c` prints `0`), so the case-only reshape loads it unchanged, and
  `readJSON`'s `DisallowUnknownFields` (`ledger.go:19-20`) makes the refusal row real.
- The roster's `RejectUnknownMembers(true)` (`roster.go:340`) and the API's closed `rosterFields`
  (`admin.go:31`, `:287`) make Task 4b's old-key rows and their mutations real.
