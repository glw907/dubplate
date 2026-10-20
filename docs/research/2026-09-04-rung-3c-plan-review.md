# Adversarial review: the rung 3 pass 3c plan

Fresh-context review (2026-09-04) of
`plans/2026-09-04-rung-3c-client-limiter.md` revision 1, in two
independent lenses: beets and wire fidelity, verified contract by
contract against pinned beets 2.13.1 in `tools/venv/.venv`, and harness,
process, and plan integrity, verified against the tree at `f2bc19e` (the
plan says `d518e57`; the two commits since are STATUS-only). Folded as
revision 2 the same day; the merged 41-item "Verified sound" list below
is the do-not-re-litigate record for the pass, alongside the rung 3 spec
review and 3a's and 3b's own review records.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B3, M-a through M-h, m1 through m26. Every
finding was re-verified by the fold editor against the tree or the venv
before folding. Four citation claims did not survive that check and are
rejected below with the evidence. Five items are conductor-owed rather
than folded silently, and each is applied provisionally in the plan with
an inline marker.

Both lenses agree the plan's query and request contracts are unusually
well grounded. One lens reproduced all 48 committed search `path_query`
strings and all 303 committed lookup `path_query` strings from
`input.items` alone using C5 through C11 and C18, byte for byte, with
zero mismatches; the other reproduced the same 48 through an independent
Python reimplementation. The findings concentrate in three places the
corpus cannot reach: the limiter's concurrency, the client's retry seam,
and the two doc surfaces Task 3 falsifies.

---

## Rejected, with the evidence

**1. `beets/util/id_extractors.py:20` is cited correctly (review 2's m2,
its `id_extractors` clause only).** Rejected. Line 20 is
`"musicbrainz": re.compile(r"(\w{8}(?:-\w{4}){3}-\w{12})")`; `:18` is
the `deezer` pattern. The plan's `:20` points at the pattern it names,
at C13 and at Task 6's `Release` doc comment. Review 1's Verified-sound
item agrees.

**2. `internal/harness/synth/release.go:128` is cited correctly (review
2's m2, its `synth` clause only).** Rejected. Line 128 is `Position
string \`json:"position"\``; `:127` is the `ID` field. The plan's `:128`
points at the string-typed `position` it names.

**3. `internal/engine/matcher/track_artist.go:33` is cited correctly
(review 2's m2, its `track_artist` clause only).** Rejected. Line 32 is
the `addTrackArtist` signature and line 33 is the `if` carrying the
inline `slices.Contains(vaArtists, strings.ToLower(item.Artist))`. The
plan's `:33` points at the expression Task 3 replaces.

**4. `internal/config/keys.go:19-25` is cited correctly (review 2's m2,
its `keys.go` clause only).** Rejected. Line 19 opens `var ownedWall =
[]WallRow{` and line 25 closes it; `:16-18` is the doc comment above it.
The plan cites the declaration deliberately, because the claim it
supports is that `ownedWall` is still a slice literal rather than a
registry.

Two further clauses of review 2's m2 survive and are folded inside m12:
`corpus_test.go:133-135` is really `:134-136`, and
`conform/translate.go:306-308` is really `:301`.

---

## BLOCKERS

### B1. The reserved token starves `ClassBatch`, the exact class the constraint the plan quotes exists to protect

Both lenses found this independently and converged on the same fix.

The plan's `Class` doc comment states the requirement it designs against:
"interactive work must not starve timer imports (build spec, the same
section)", a faithful quotation of
`docs/specs/2026-09-02-greenfield-server-build.md:693-694`.
`ClassBatch` is the plan's own name for "an operator- or timer-triggered
import".

The dispatcher rule is "on each tick it serves the highest-priority
non-empty queue, except that every `reservationInterval`-th served token
goes to the lowest-priority non-empty queue", with `reservationInterval
= 4`. With `ClassUserPresent` and `ClassBackground` both non-empty, which
is rung 7's probe plus rung 4's shadow, the two callers the plan names by
rung, `ClassBatch` receives nothing. It never wins the normal token,
because `ClassUserPresent` outranks it. It never wins the reserved token,
because `ClassBackground` is the lowest non-empty class. Rung 9's timer
import is the starved party, which is the case the comment promises is
covered.

The plan's proof cannot see it.
`TestDispatcherPrefersUserPresentAndStillFeedsBackground` queues one
`ClassBackground` waiter and eight `ClassUserPresent` waiters and never
queues a `ClassBatch` waiter at all. Task 5's seven named mutations never
queue all three classes at once either. An implementer builds this
exactly as written, every proof goes green, and the property the build
spec names is absent.

**Fix, as both lenses recommend it.** The reserved token goes to the
non-empty classes *below* the class strict priority would have served, in
round-robin, so a saturated probe stream alternates the reservation
`ClassBatch`, `ClassBackground`, `ClassBatch`. The doc comment carries
that sentence. Task 5 grows a three-class row: one `ClassBatch` waiter
and one `ClassBackground` waiter queued behind a saturated probe stream,
asserting both are served within eight tokens and naming which token each
got, plus the matching mutation ("the reserved token always goes to the
lowest class") so the middle class has a firing proof of its own.

### B2. Task 5's limiter test has no waiter-registration handshake, and the plan codifies the false rationale as a Global Constraint

`TestDispatcherPrefersUserPresentAndStillFeedsBackground` is given
verbatim with this comment on its mechanism: "the `queued` channel is how
the test knows a waiter is in its queue without sleeping". Task 5's
behavior contract raises it to a rule for the whole task: "a test that
needs to observe a waiter having queued does so by receiving a signal the
waiter sends before calling `Wait`, never by sleeping."

The signal is sent **before** `Wait`, so it proves nothing about
registration:

```go
go func() {
    queued <- struct{}{}
    if err := b.Wait(context.Background()); err != nil { ... }
}()
```

Between the send on `queued` and the enqueue inside `Wait` there is an
unbounded scheduling window. The main goroutine then does `ticks <-
time.Time{}` on an unbuffered channel, which completes as soon as the
*dispatcher* receives it, not as soon as a waiter is queued. Under the
plan's own dispatcher rules ("a tick with no waiter is kept as one spare
token", and a waiter arriving while the spare is held "proceeds
immediately"), the first tick becomes the spare and the first waiter to
actually register consumes it. That waiter can be the `shadow`, because
`start(shadow, "shadow")` runs first. `order[0]` is then `"shadow"` and
the precedence assertion fails.

The race is not theoretical. Nine goroutines are launched and the main
goroutine only ever waits on a buffered channel (`queued` is
`make(chan struct{}, 16)`), so the sends never block and the main
goroutine can reach the first tick with zero waiters registered. The
consequence is a flaky test inside `go test ./...`, which is the repo's
gate leg, on the pass's only concurrent code, under a plan that tells the
implementer the pattern is sound.

**Fix.** The limiter needs a registration observation point the test can
block on, and the Global Constraint must be rewritten around it. Have
`Wait` enqueue by sending its request to the dispatcher on an unexported
channel, and add an unexported test seam beside `newLimiter`, for example
`func (l *Limiter) awaitQueued(n int)`, that blocks until the dispatcher
has registered `n` waiters (a counter under the existing mutex plus a
`sync.Cond`, or a dispatcher-side `registered chan<- struct{}` the test
drains). The test then reads `start(shadow, ...); l.awaitQueued(1); for
range 8 { start(probe, ...) }; l.awaitQueued(9)` before the first tick.
Both the precedence row and the `order[reservationInterval-1] ==
"shadow"` row become deterministic. The plan's constraint sentence says
that a pre-`Wait` signal is not a registration proof, so the next task
does not repeat it.

### B3. Task 6's retry test feeds three ticks from an unsynchronized goroutine against a spare capped at one, so tokens are dropped and the test deadlocks

Task 6's `TestReleaseRetriesOnTransientStatusAndMetersEveryAttempt` opens:

```go
ticks := make(chan time.Time)
l := newLimiter(ticks)
defer l.Close()
budget := l.Budget("test", ClassBatch, 0)
go func() {
    for range 3 {
        ticks <- time.Time{}
    }
}()
```

The client then makes three sequential attempts (503, 503, 200), each
acquiring a token under R13's divergence 2. The feeder goroutine is not
synchronized with those attempts, and the plan's dispatcher spec caps the
spare at one, a cap Task 5's mutation 4 makes a named acceptance
criterion.

So the dispatcher receives tick 1 with no waiter queued and holds it as
the spare. The feeder immediately sends tick 2; the dispatcher receives
it, the spare is already held, and that token is dropped. Attempt 1 takes
the spare, attempt 2 blocks, and tick 3 may or may not arrive before
attempt 2 registers. When it does not, the test hangs until `go test`'s
panic timeout. The window is wide: between the `go func()` and the first
`budget.Wait` the main goroutine runs `NewClient` and its validation,
assigns `c.sleep`, builds the request URL, and enters `doWithRetry`. The
feeder needs only two unbuffered sends to win.

**Fix.** Do not feed ticks ahead of demand. Drive the tick from inside
the scripted `RoundTripper`, one tick per observed attempt, which is
exactly one token per attempt and is what the test asserts anyway; or
pair each tick with B1's handshake, `l.awaitQueued(1); ticks <-
time.Time{}` once per attempt. Whichever shape is chosen, the plan states
that a hand-driven tick channel produces a deterministic token count only
when each tick is delivered after its waiter has registered, because the
spare is capped at one and an early tick is discarded.

---

## MAJORS

### M-a. Task 3 mandates a doc comment describing a relationship the code does not have, leaves an exported tunable with no reader, and omits the two doc comments its change falsifies

Both lenses found this; the two falsified comments are review 1's find
and the readerless-field measurement is review 2's.

R15 and Task 3 close 3a's carry by having `dataSourcePenalty` return a
package constant instead of `d.tun.DataSourceMismatchPenalty`. The
carry is real and the fix is right: the Go fallback is at
`internal/engine/matcher/penalty.go:404-409`, returning
`d.tun.DataSourceMismatchPenalty` at `:408`, while beets falls back to
the class literal `MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY
= 0.5` (`$SP/beets/metadata_plugins.py:160-169`, `:180`). Three things
around it are wrong.

**Two doc comments become false and neither is in Task 3's Files list.**

- `internal/engine/matcher/penalty.go:334-337`, on `AddDataSource`
  itself: "The penalty is the loaded source's own, taken from
  Tunables.DataSourcePenalties, and Tunables.DataSourceMismatchPenalty
  when no loaded source is named after". After the change it is the
  constant.
- `internal/engine/matcher/tunables.go:138-140`, on `DataSourcePenalties`:
  "the loaded metadata source set AddDataSource's get_penalty(after)
  scans before falling back to DataSourceMismatchPenalty". Same.

Neither vale nor golangci-lint sees a stale prose claim, so both would
ship as committed contradictions of the code, on a rung whose whole
premise is that a divergence between the record and the port is the
defect class.

**The field's stated "real job" does not exist.** The plan instructs the
implementer to rewrite the field's doc comment "to say what it is now
for: it populates `DataSourcePenalties["MusicBrainz"]` in
`DefaultTunables`". It does not.
`internal/engine/matcher/tunables.go:282-285` writes two independent
literals:

```go
DataSourceMismatchPenalty: 0.5,
DataSourcePenalties: map[string]float64{
    "MusicBrainz": 0.5,
},
```

No expression connects them. Measured over the tree with `_salvage`
excluded, the field's readers today are exactly `penalty.go:408` (the
fallback Task 3 replaces) and `internal/harness/conform/component.go:673`
and `:721` (the two compensations Task 3 removes), plus the
`DefaultTunables` literal and one `tunables_test.go` assertion. After
Task 3 the field has **zero production readers**, and `golangci-lint`'s
`unused` does not report an exported field of an exported struct, so the
gate stays green over a tunable that configures nothing.

**Fix.** Add `penalty.go:334-337` and `tunables.go:138-140` to Task 3's
Files list with their exact rewording. Then pick one honest ending for
the field and say which in the plan: wire it, by building
`DataSourcePenalties` in `DefaultTunables` from the field so the doc
comment becomes true and the field stays load-bearing; or keep it as the
config-load seam for `musicbrainz.data_source_mismatch_penalty` /
`source_weight` whose only consumer today is `DefaultTunables`'s own
map, said plainly in the comment and carried forward to the rung-1b
registration task. **Conductor-owed** (see the fold's ruling list); the
fold applies the wiring option provisionally, because in beets the map's
MusicBrainz entry *is* the plugin's own resolved
`data_source_mismatch_penalty` (`$SP/beets/metadata_plugins.py:182-195`),
so wiring is the faithful shape and is behavior-neutral at 0.5.

### M-b. The client's production wire pace is never stated, and C12 foregrounds the wrong number for musicbrainz.org

C12 gives `RateLimitAdapter(rate_limit=0.25, ...)` "enforcing a 0.25s
process-wide floor between any two requests", and R2 cites that same
adapter as what the spare token reproduces. That is the
four-requests-per-second floor. It is not the rate beets paces
MusicBrainz at.

`$SP/beetsplug/_utils/musicbrainz.py:597-599` sets `self.api_host,
self.rate_limit = "https://musicbrainz.org", 1.0` for the
`musicbrainz.org` host, and `:612-613` builds the session as
`LimiterTimeoutSession(per_second=self.rate_limit)`, whose
`requests_ratelimiter` mixin defaults `burst: float = 1`
(`$SP/requests_ratelimiter/requests_ratelimiter.py:35`) and constructs
the bucket as `per_second * burst` over `Duration.SECOND * burst`
(`:47-51`). The effective pace against the live API is the tighter of the
two, **1 request per second**. The repo's own live spender agrees:
`cmd/harness/capture.go:27-31`, `requestInterval = time.Second`, "at most
one request per second against the live API, regardless of MusicBrainz's
own configured rate limit".

The plan's `NewLimiter(perSecond int)` never names the MusicBrainz value,
`ClientConfig` carries no rate at all, and the only place `1/s` appears
is incidental, inside `Unmetered`'s rationale. An implementer reading
C12's prominent 0.25s can reasonably ship `NewLimiter(4)` and put rung
4's shadow four times over MusicBrainz's published policy.

Review 2 found the same gap from the signature side: `NewLimiter(perSecond
int)` cannot express beets' non-`musicbrainz.org` branch, which derives a
float (`ratelimit / ratelimit_interval`,
`$SP/beetsplug/_utils/musicbrainz.py:600-606`), nor a deliberately
conservative sub-1/s setting; the organizer spec records the anonymous
limit as one request per second with more for identified clients
(`docs/specs/2026-09-01-dubplate-organizer-design.md:531-533`).

**Fix.** State the pace as a contract citing `_utils/musicbrainz.py:597-599`
and `:612-613` and the `requests_ratelimiter` burst; reword C12 so the
0.25s adapter is described as the secondary floor it is and say plainly
that for `musicbrainz.org` the 1/s session limiter binds first; and take
an interval rather than an integer rate, `NewLimiter(interval
time.Duration)` with a named `DefaultInterval = time.Second` carrying the
citation. **Conductor-owed**, because shipping a default wire pace is a
choice with a material consequence for rung 4's shadow.

### M-c. `Retry-After`'s HTTP-date form needs a clock read, so R13's "exactly two named deliberate divergences" is incomplete and the arch check forbids the faithful port

Review 2's find; review 1 saw the same gap from the fixture-table side
and is folded here.

The plan says "`Retry-After` honored over the computed backoff when the
response carries one", with R13 declaring exactly two deliberate
divergences from beets, and Task 6's mutation 6 names only the seconds
form.

beets' behavior is urllib3's, and urllib3 accepts both forms.
`parse_retry_after` (`$SP/urllib3/util/retry.py:328-347`) parses a bare
integer as seconds and otherwise parses an HTTP-date and computes
`retry_date - time.time()`; `get_retry_after` (`:349-357`) feeds both.
Reproducing that in Go requires reading the current time.
`internal/engine/musicbrainz/client.go` is inside
`TestEngineNeverReadsTheClock`'s walk
(`internal/harness/conform/arch_test.go:152-157`, rooted at
`internal/engine`), so `time.Now` reds the gate, and `time.Since` passes
the checker while violating the sentence R2 wrote about it ("a green walk
is not proof a file is clock-free"). R2 resolves the clock question for
`limiter.go` alone and never notices the client has the same problem.

An implementer who narrows `Retry-After` to delta-seconds ships an
unrecorded third divergence; one who ports it faithfully reds the gate or
reaches for an exemption the plan forbids.

**Fix.** Name the narrowing in R13 as divergence 3: delta-seconds only,
an HTTP-date `Retry-After` falls through to the computed backoff, because
honoring it would require a clock read
`internal/harness/conform/arch_test.go` forbids. Put the reason on the
code, add a table row for the date form to Task 6's fixture table, and
carry it in Task 8's forward carries beside the other two.

### M-d. `Budget`'s cap is a check-then-act race, and the type exists only for concurrent callers

The plan says "the budget's limit is checked **before** the wait, not
after, so an exhausted budget never occupies a queue slot; `Spent`
increments only on a granted token." `Budget.Wait`'s contract promises
`ErrBudgetExhausted` "when the limit is spent", and `Budget` is given a
`sync.Mutex`, so concurrent use is plainly intended; the `Class` comment
names three concurrent callers by rung.

Two goroutines sharing one `Budget` with `limit == 1` and `spent == 0`
both pass the pre-wait check, both queue, both are granted, and `spent`
ends at 2. The cap is advisory under exactly the conditions the type was
built for, and no `-race` run reports it, because there is no data race,
only a lost update in the decision. Task 5's mutation 6 tests the cap
sequentially, so it passes.

**Fix.** Reserve before waiting and release on failure: increment `spent`
under `b.mu` and return `ErrBudgetExhausted` when the incremented value
exceeds `limit`, then decrement on any error path out of the wait
(cancellation, limiter close). State that shape in the behavior contract
in place of "checked before the wait", and add a test row: `limit: 2`,
four concurrent waiters, assert exactly two succeed and `Spent() == 2`.

### M-e. `SearchReleases` is specified without its wire shape, and the search response is not a bare array

Every other wire shape in this plan is nailed down and this one is not.
`SearchReleases` is defined as returning "the release ids in response
order", citing `search` and `candidates`. C4 quotes only `search`'s
parameter handling. Neither states the key the ids come from.

`$SP/beetsplug/_utils/musicbrainz.py:665-685` ends with

```python
normalised_entity = entity.replace("-", "_")
return self._get_resource(entity, **kwargs)[f"{normalised_entity}s"]
```

so the body is an object and the ids live under `"releases"`. Measured
mechanically over the 48 committed search responses (top-level key names
only, no body read into context): every one carries exactly `("count",
"created", "offset", "releases")`. An implementer decoding into
`[]struct{ ID string }` fails all 48 searches in Task 7 and has to
rediscover the shape.

**Fix.** Add a contract, C19: a search response is an object whose
`releases` array carries the ordered hits, each with an `id`; the client
decodes `struct { Releases []struct{ ID string \`json:"id"\` }
\`json:"releases"\` }`, and `count`, `created` and `offset` are ignored at
rung 3. Name the 48-of-48 measurement as the evidence.

### M-f. The escaped-hyphen mutation's measured miss count is 6, not 5, in both tasks that state it

Task 7's mutation 2 says "Remove `-` from the Lucene escape set → **the 5
`arid` searches miss**", and Task 4's mutation 2 repeats the count.

Re-measured by the fold editor over every committed search row: six of
the 48 carry an escaped hyphen, across six distinct cases and five
distinct query strings. The five `arid` rows carry four escaped hyphens
each inside the various-artists MBID, and one non-`arid` row carries one:

```
release%3A%28bob+dylan+%5C-+the+bootleg%29+artist%3A%28bob+dylan%29
```

The six cases are `nonva-va-artists-misflag`, `real-noise-dylan-bootleg`,
`real-noise-zappa-cucamonga`, `real-nonlatin-utada-first-love-locale`,
`real-nonlatin-utada-first-love`, and `va-track-mismatch-all`; `grep -c
'%5C-'` over the 48 committed search `path_query` values returns 6 and
over the 46 distinct values returns 5.

The plan tells the reviewer to check reported miss counts against the
plan's own figures, so a wrong figure either gets fudged into agreement
or costs a re-dispatch. It is the same shape as 3b's M-c.

**Fix.** State 6 searches over 6 cases in both places, and name the two
contributors (five `arid` rows through the MBID, one `release:` clause
through the dylan title) so the number is checkable rather than
memorised.

### M-g. Task 2 changes `stub.New`'s behavior for `oracle.Derive`, and the plan checks only that the derivation path still compiles

The plan says "**`New`'s signature does not change**, so the four
existing consumers keep compiling untouched: `internal/harness/oracle/
oracle.go:163`, `ties.go:65`, `album_info.go:57`, and
`cmd/harness/capture.go`'s use of `stub.Miss`. The task verifies that by
building, not by assertion."

Compilation is the wrong check for three of them. `oracle.Derive` is the
derivation path, and `stub.New` is what pinned beets talks to while a
committed record is produced. Adding a second replay map changes what
that server answers: a request that previously drew a 404 and a recorded
`Miss` now draws a 200 with a body, which is a change to the oracle's own
input and therefore to `make derive`'s byte-identity invariant, the
standing invariant the plan restates in its Global Constraints.

The change is inert today (zero committed cases carry an `extra` key,
measured over all 215) and stays inert for as long as an `Extra` entry is
by construction a request the current flow does not issue. Both of those
are arguments the plan never makes, on a shared instrument, in the one
pass that reshapes it.

The consumer list is also one short. `stub.New` has a fifth caller,
`cmd/harness/capture_test.go:544`, which the plan's "four existing
consumers" omits.

**Fix.** State the derivation rule in Task 2's behavior contract and in
`stub`'s package doc: an `Extra` entry is a response the flow does not
request, so the extra map cannot change a derivation, and if a future
case makes that false the entry belongs in `Requests`. Add the standing
assertion that zero committed cases carry `Extra` at this pass, so the
day E1 populates the channel the derivation question is raised rather
than discovered in a byte diff. Correct the consumer count to five.

### M-h. Task 4's mutation 8 asks for a determinism Go's map iteration cannot give

Both lenses found this.

Task 4's mutation 8 says "Iterate `FIELDS_TO_MB_KEYS` instead of
`tun.ExtraTags` → the two-tag order row fails, and it must fail on
**every** run rather than flakily; if it passes on some runs the fixture
has too few tags and the task enlarges it until the mutation is
deterministic, reporting the size that was needed."

Go randomises map iteration order per range statement. A mutant ranging
over an `n`-entry map produces the fixture's expected order with
probability `1/n!`, which falls fast and never reaches zero. The
instruction cannot be satisfied, and the likely outcome is an implementer
running the mutant a few times, seeing red each time, and reporting
determinism it did not establish. That is the false-proof shape the
assertion discipline exists to prevent, in a task whose whole
justification is fixture quality.

**Fix.** Replace the mutant with a deterministic one: iterate the
`FIELDS_TO_MB_KEYS` map in sorted key order
(`slices.Sorted(maps.Keys(...))`), and choose the fixture's `ExtraTags`
order so it differs from sorted MB-key order. `["media", "label"]`
renders `format:(...) label:(...)` while sorted MB keys give `catno`,
`country`, `date`, `format`, and so on, so the mutation fails on every
run for a stated reason rather than a probabilistic one.

---

## MINORS

**m1.** The plan cites `util.plurality` at
`$SP/beetsplug/musicbrainz.py:731`. The call is at `:727`; `:729-730` is
the `catalognum` space-strip and `:732` is `criteria[mb_field] = value`.
C8's own range `:721-732` is correct.

**m2.** R11 names the rung 3 spec `:226` as one of three "losing"
documents on the `search_limit` legacy spelling, and Task 8 drafts a
correction for it. The spec line reads "`musicbrainz.search_limit` with
its legacy `match.searchlimit` spelling", immediately followed by
"`musicbrainz.data_source_mismatch_penalty` with its legacy
`source_weight` spelling". The second item's legacy path is
`musicbrainz.source_weight`, so the line's convention is legacy spelled
relative to the plugin namespace, under which `match.searchlimit` is
correct. `internal/engine/matcher/tunables.go:122` ("musicbrainz.searchlimit")
and `docs/HISTORY.md:600` ("the `SearchLimit` legacy spelling is
`musicbrainz.searchlimit`, not a top-level `match:` key") are genuinely
wrong and stay corrections. Change the spec item to an optional
clarification, spelling it `musicbrainz.match.searchlimit`, and drop
"losing document" for that one target; a bare "correction" there would
make the two adjacent items inconsistent.

**m3.** C12 and R13 both say "the same five retryable statuses (429, 500,
502, 503, 504)". `urllib3`'s `Retry.is_retry`
(`$SP/urllib3/util/retry.py:420-431`) also retries a status in
`RETRY_AFTER_STATUS_CODES = frozenset([413, 429, 503])` (`:202`) whenever
the response carries a `Retry-After` header, so 413 plus `Retry-After` is
a sixth retryable case. It is unreachable from MusicBrainz, and the
contract is stated as exhaustive with a Task 6 sentinel test against it.
Record it as a deliberate non-port with that reason.

**m4.** Task 1's round-trip failure message reports `len(want)-len(got)`
as "changed %d bytes", which prints `0` for the common case of an
equal-length divergence. Report the first differing offset and a short
window around it instead.

**m5.** Task 2's post-loop identity guard is `if got := srv.Consumed();
len(got) == 1 && got[0] != c.Extra[0]`, which passes silently when
`Consumed()` is empty or holds two entries. Invert it: `if got :=
srv.Consumed(); len(got) != 1 || got[0] != c.Extra[0]`.

**m6.** `boolField` accepts only a JSON bool. beets stores `comp` as a
`types.Boolean` integer flag, so a case carrying `"comp": 1` would
project `false` here and `True` in the shim. Measured: 15 of 215 cases
carry `comp`, all 15 as JSON `true`, so nothing moves today. Note it in
the field's comment the way `intField`'s own comment notes the float64
arrival (`internal/engine/matcher/info.go:209-211`).

**m7.** Task 3's "named mutation" 5 is a restoration that must leave
72/72 green, which is a no-op. It is useful evidence and it is not a
firing proof, and listing it under "Named mutations this task reports"
muddies the assertion-discipline count. Move it to the acceptance
evidence paragraph.

**m8.** Whether `Client.Release` calls `(*Release).Validate` is left
open: `AlbumForID`'s contract says "fetch, validate, translate" while
`Release`'s says only that it errors "when the body does not decode", and
Task 7 calls `Release`. Meanwhile `internal/engine/musicbrainz/translate.go:46-48`
already calls `rel.Validate()` as `TranslateAlbum`'s first act, so
"fetch, validate, translate" validates twice. Measured, both readings are
safe: all 303 committed lookup responses decode and validate cleanly.
Settle it anyway, because rung 4's shadow branches on the sentinel:
`TranslateAlbum` owns validation, `Release` decodes and does not
validate, and `AlbumForID` adds no explicit call.

**m9.** Task 7's wall-time basis says the sweep's decode "is the same
decode 3b's browse-fallback guard does over the same files". It is not:
that guard decodes into a two-field anonymous struct
(`internal/harness/conform/translate.go:423-428`,
`struct { ID string; Media []struct{ Tracks []json.RawMessage } }`),
while the sweep decodes the full `musicbrainz.Release`. The estimate
survives anyway; probed, all 303 committed lookup responses decode into
`musicbrainz.Release` and pass `Validate` in 1.89s with zero decode and
zero validate errors. Correct the basis sentence and keep the 3-to-8s
expectation.

**m10.** `checkNoTimeNow` walks **every** `.go` file under
`internal/engine`, `_test.go` included
(`internal/harness/conform/arch_test.go:75-89` skips only a `testdata`
directory). The plan warns only that `limiter.go` "calls `time.Now`
nowhere". Say the same of `limiter_test.go` and `client_test.go`, since a
backoff-schedule or elapsed-time assertion is the natural thing to reach
for and it reds the gate for every lane.

**m11.** R14 frames all 56 case-mapping divergences as "Python `re` and
full-case semantics against Go RE2 and simple case mapping". The sweep
file's own header says otherwise
(`internal/engine/matcher/testdata/case-mapping-sweep.tsv:7-10`): "One of
them, U+0130, is the full-versus-simple case mapping difference the sweep
went looking for. The other 55 are a Unicode version skew: Go 1.27
carries Unicode 17.0.0 case tables and CPython 3.13 carries 15.1.0." That
matters for the query path, because the version-skew half moves the day
either toolchain moves and the plan's single asserted row (`İ`) is the one
member that will not. Say so where the divergence is declared.

**m12.** Citation drift, none load-bearing alone, all worth fixing before
an implementer follows one into code. `conform/translate.go:306-308` for
"the stub is 3c's business" is `:301`. `internal/harness/corpus_test.go:133-135`
for the `c.Extra` iteration is `:134-136`. `info.go:204-210` for
`stringField`'s shape is `:202-207`; `:209-211` is `intField`'s doc.
`_utils/requests.py:120-136` for the adapter's wait is `:128-135`.
`match.py:443-446` for the search-term default is `:445-447`.
`metadata_plugins.py:419-427` for the search's bare `except Exception`
omits that the handler re-raises first when `config["raise_on_error"]` is
set (`:422-423`). And "verified at HEAD `d518e57`" is two STATUS-only
commits behind `f2bc19e`; Task 0 re-verifies against the real HEAD.

**m13.** The `HISTORY` citation points one line short of the text to
correct. `docs/HISTORY.md:597-599` is the guard's heading and the
`DataSourceMismatchPenalty` citation; the wrong spelling is on `:600`.
Cite `:597-601` or `:600`. Task 0 item 11 also says "the two documents
R11 names" and then names three targets; say "the one code site and the
two documents".

**m14.** The `Case` struct snippet silently deletes four landed comments.
The plan shows the reshaped struct without the trailing comments
`internal/harness/cassette/cassette.go:43-48` carries on `Input`,
`Profile`, `Requests` and `Expected` ("items in pinned order; no file
references at 2a", "the committed request log: an expectation, not
scaffolding", and so on). An implementer pasting the block drops them.
Show the block with the existing comments intact.

**m15.** The three-row table and the test that implements it disagree on
row 3. The prose table gives row 3 `Consumed()` "empty" while the test
gives the same row `wantConsumed: 1`. Both readings are defensible (per
event versus cumulative) and the plan explains the cumulative choice
after the table. Make the prose table cumulative so it matches the
accessor it names.

**m16.** Nothing forbids two `Extra` entries sharing a key or a response
file. Task 1 adds the `Extra`-versus-`Requests` key rule, and
`TestEveryLoggedRequestResolvesToExactlyOneResponse`
(`internal/harness/corpus_test.go:105-119`) covers `c.Requests` only. Two
`Extra` rows with the same `(Method, PathQuery)` load clean and the
stub's extra map silently shadows one, which is the defect the existing
test exists to prevent one map earlier. Extend the new rule to
`Extra`-versus-`Extra`.

**m17.** The extra-tag walk loses beets' de-duplication. Task 4 iterates
`tun.ExtraTags` and looks up the MB key; beets builds a dict keyed by tag
(`$SP/beetsplug/musicbrainz.py:704-708`), so `extra_tags: ["tracks",
"tracks"]` yields one criterion there and two in a slice walk. Skip a tag
already emitted, and add the duplicate row to the extra-tag fixture
table.

**m18.** `SearchReleases`' error contract is unspecified. Task 6 states
the error map for lookups only. C15 records that beets swallows a search
failure into an empty result, which is a real fork an implementer must be
told to take or refuse. State which sentinel a search returns on 404, on
a body that does not decode, and on exhausted retries.

**m19.** `VALikely` is exported ahead of any out-of-package consumer.
Task 4 exports it while `AlbumSearchFor(items, tun)` computes the
likelies internally, so the only caller in this pass is inside the
package; R1's seam sentence names `musicbrainz.VALikely` as one of the
three things `conform` builds the query through, which Task 7's own
Consumes list contradicts. R6 rejects `Unused()` on precisely this ground
("Shipping an exported accessor with no consumer is dead surface under
go-conventions"). Keep it unexported until 3d's flow needs it, and
correct R1's sentence to name `AlbumSearchFor` alone.

**m20.** The User-Agent contact repurposes a field documented as never
displayed, and departs from the precedent it cites. R11 proposes taking
the contact from `Config.AdminEmail`, whose own doc comment says
"adopters compare it, never display it"
(`internal/config/config.go:60-64`), and cites
`cmd/harness/capture.go:130-134` as precedent. That precedent takes its
contact from a dedicated input, `DUBPLATE_HARNESS_CONTACT`, whose comment
says the tool "never invents one (never a literal, per CLAUDE.md's
universality rule)" (`cmd/harness/capture.go:21-26`) and refuses to run
unidentified. Record the divergence, and name a dedicated contact row as
the post-1b follow-on rather than implying `AdminEmail` is the settled
answer.

**m21.** Task 6's deliverable count of four omits two named deliverables.
The task's own file list also modifies `release.go` (the `wireInt`
sentinel wrap, a behavior change with its own mutation) and `doc.go`.
CLAUDE.md requires saying plainly when a task passes roughly four. Count
six and say so.

**m22.** `corpus/README.md` is edited by this pass and appears in no
surface list. The contested-surface section says "3c touches seven" and
Task 8 drafts four `corpus/README.md` corrections for the conductor. The
`corpus/analysis/*` and `corpus/cases/*` row does not cover the README
and the `docs/` ledgers row does not either. Name it as an eighth
surface, conductor-owned by the same standing rule, in the same table
that asks for the `corpus/manifest.json` row. **Conductor-owed.**

**m23.** The 80% flag routes to Geoff during a window STATUS declares
unattended. The plan says "finish the task in flight, write STATUS, and
one combined question goes to Geoff before continuing", while
`docs/STATUS.md:29-31` records the standing grant to "run unattended
through 3c execution, 3d prep, and 3d execution as the clock allows" and
CLAUDE.md's grants say to continue unblocked work when Geoff is
unavailable. Route the flag to the conductor, with STATUS as the record
and Geoff's reading of it as the touchpoint. **Conductor-owed**, because
it reads against CLAUDE.md's literal rule.

**m24.** Task 7's case-accounting identity assumes at most one search per
case without saying so. `Cases == Searches + CasesNoSearch` holds today
(measured: 164 cases at 0 searches and 1 lookup, 3 at 0/2, 26 at 1/1, 4
at 1/2, 9 at 1/5, 9 at 1/6) and follows from `tag_album` issuing one
search, and a future two-search case would red the gate for a reason the
message does not name. Say in the message that a case carrying two
searches is itself the finding.

**m25.** `albumCriteria`'s signature passes what its first parameter
already carries. It takes `likelies`, `itemCount`, `artist`, `album`,
`vaLikely` and `tun`, where `artist` and `album` are `likelies.Artist`
and `likelies.Album` at every call site this pass builds
(`$SP/beets/autotag/match.py:445-447`). Six parameters with two redundant
is the shape go-conventions asks to collapse. Drop `artist` and `album`
until 3d's explicit-search-term surface needs them, and say so where the
plan already explains why the override is not modelled.

**m26.** `Wait`'s cancellation property is stated with no mechanism. The
plan requires that a waiter losing the race to `ctx.Done()` not consume a
token, and that the next waiter still get the token that was in flight.
With a dispatcher that has already sent on a grant channel, that needs a
designed return path. The task is upshifted to `opus`, so the design is
in the right hands; the plan should still name the mechanism it expects
(an unbuffered grant handshake, or a token returned to the spare slot) so
the reviewer has something to check the code against.

---

## VERIFIED SOUND, the do-not-re-litigate record

Merged from both lenses and re-checked by the fold editor against the
tree at `f2bc19e` or the pinned venv. A later round should not reopen
these.

1. **C1, `fmt=json` forced with no exception.** `MusicBrainzAPI.request`
   at `$SP/beetsplug/_utils/musicbrainz.py:615-619`,
   `kwargs["params"]["fmt"] = "json"` at `:618`. All 351 committed
   requests carry it.
2. **C2, the API root and the ignored keys.** `api_root` at `:608-610`;
   the `musicbrainz.org` branch pinning `api_host` and `rate_limit` at
   `:597-599`; the other-host branch at `:600-606`. The corpus's own
   adapter appends `host` / `https: no` / `ratelimit: 100` at
   `internal/harness/oracle/oracle.go:247`.
3. **C3, the sixteen-element `RELEASE_INCLUDES`.**
   `$SP/beetsplug/_utils/musicbrainz.py:50-67`, sixteen entries in
   exactly the order the plan's Go literal carries; `get_release`'s
   default at `:687-690`; `kwargs["inc"] = "+".join(...)` at `:635`.
   Measured: one distinct `inc=` value across all 303 committed lookups,
   byte-identical to `internal/harness/mbrequest/mbrequest.go:22-24`.
4. **C4, a search carries no `inc`.** `_get_resource` sets `inc` only
   when `includes` is truthy (`:634-635`) and `search` never passes one
   (`:665-685`). Measured: `inc` on 303 of 351, on none of the 48
   searches.
5. **C5, the nineteen-character Lucene escape set.**
   `LUCENE_SPECIAL_CHAR_PAT` at `$SP/beetsplug/_utils/musicbrainz.py:48`,
   `([-+&|!(){}[\]^"~*?:\\/])`. Measured over the committed logs: `%27`
   twice and `%2C` once survive unescaped, `%5C` appears 24 times.
6. **C6, lower then strip then drop then escape then wrap.**
   `format_search_term` at `:651-663`: `if not (term :=
   term.lower().strip()): return ""`, then the `sub`, then the
   `f"{field}:({term})"` wrapper at `:660-661`. The committed
   `release%3A%28frank+zappa%5C%3A+cucamonga%29` proves the wrapper's own
   colon and parens are unescaped while the term's colon is escaped.
7. **C7, single-space join with empties filtered.** `:679-681`, no
   literal `AND`.
8. **C8, the criteria order and the eight-entry field map.**
   `get_album_criteria` at `$SP/beetsplug/musicbrainz.py:714-734`,
   `{"release": album} | ({"arid": VARIOUS_ARTISTS_ID} if va_likely else
   {"artist": artist})` at `:717-719`, the extra-tag loop at `:721-732`
   with its three value rules. `VARIOUS_ARTISTS_ID` at `:43`, matching
   `internal/engine/musicbrainz/credit.go:11`. `FIELDS_TO_MB_KEYS` is
   **eight** entries at `:47-56`, exactly the mapping the plan's header
   correction states. `extra_mb_field_by_tag` at `:697-712` is a dict
   comprehension over `self.config["extra_tags"].as_str_seq()`, so
   insertion order is the operator's order.
9. **C9, the pre-filter's second lower-and-strip.**
   `get_search_query_with_filters` at `:736-753`, returning `("", {k: _v
   for k, v in criteria.items() if (_v := v.lower().strip())})` at
   `:750-752`, with the query text half always empty for MusicBrainz.
10. **C10, `va_likely`'s three disjuncts and the consensus asymmetry.**
    `$SP/beets/autotag/match.py:451-455`; `VA_ARTISTS` at
    `$SP/beets/autotag/distance.py:29`; `likelies, consensus =
    get_most_common_tags(items)` at `match.py:407`; the albumartist
    override at `$SP/beets/util/__init__.py:842-844`, which rewrites
    `likelies["artist"]` and leaves `consensus["artist"]` alone.
    `internal/engine/matcher/likelies.go:47-83` reproduces both halves,
    override at `:78-80`.
11. **The six plurality extra tags are all available from `Likelies`.**
    `get_most_common_tags`'s field list
    (`$SP/beets/util/__init__.py:822-836`) contains `year`, `label`,
    `barcode`, `catalognum`, `country`, `media`, and `matcher.Likelies`
    carries all six (`internal/engine/matcher/likelies.go:11-25`). The
    counting is identical: `Counter.most_common(1)` is first-seen on ties
    and `matcher.plurality` reproduces that explicitly
    (`likelies.go:95-121`), over the same item order.
12. **The `date:(0)` trap is real and the plan's reading is right.**
    `value = str(most_common)` at `musicbrainz.py:728` over
    `i.get("year")`, which on a beets `Item` is the Integer field's own
    zero, so an unset year ships as `date:(0)`. `strconv.Itoa(likelies.Year)`
    reproduces it.
13. **C11, `search_limit` 5 and the legacy read.**
    `MetadataSourcePlugin.__init__` adding `"search_limit": 5` at
    `$SP/beets/metadata_plugins.py:201`; `_search_api` reading it at
    `:415` into `SearchParams` at `:416`; the legacy read
    `self.config["match"]["searchlimit"]` at
    `$SP/beetsplug/musicbrainz.py:292-294`, with `self.config` being
    `beets.config[self.name]` (`$SP/beets/plugins.py:200`), so the path
    is `musicbrainz.match.searchlimit`; beets' own deprecation string
    spelling it `'musicbrainz.searchlimit'` at `:297`, which is where the
    wrong copies came from. Measured: all 48 committed searches carry
    `limit=5`. `internal/engine/matcher/tunables.go:281` ships 5.
14. **C12's transport numbers.** User-Agent set once on the session at
    `$SP/beetsplug/_utils/requests.py:76`; `Retry(total=6,
    backoff_factor=0.5, status_forcelist=[500, 502, 503, 504, 429])` at
    `:78-88`; `RateLimitAdapter(rate_limit=0.25, max_retries=retry)` at
    `:89`; `kwargs.setdefault("timeout", 10)` then `raise_for_status()`
    at `:99-101`; the process-wide floor under a lock at `:128-135`.
    Backoff: `$SP/urllib3/util/retry.py:309-326`, zero when the
    consecutive-error run is one or less, then `0.5 * 2**(n-1)`, capped
    at `DEFAULT_BACKOFF_MAX = 120` (`:210`) with `backoff_jitter: float =
    0.0` (`:238`), giving 0, 1, 2, 4, 8, 16.
    `respect_retry_after_header: bool = True` (`:234`) and `Retry.sleep`
    (`:373-387`) takes `Retry-After` over the computed backoff. Retries
    do sit inside the adapter's own `urlopen` loop, so only the outermost
    attempt is metered. `requests_ratelimiter`'s `burst` default is 1
    (`requests_ratelimiter.py:35`), so the spare-token restatement is
    faithful. (M-b corrects only which floor binds for `musicbrainz.org`,
    never these numbers.)
15. **C13, `album_for_id`'s shape.** `$SP/beetsplug/musicbrainz.py:765-800`:
    `_extract_id` first with `return None` and no request at `:771-773`;
    `get_release` inside `try/except HTTPNotFoundError` returning `None`
    at `:776-781`; the pseudo branch at `:785-798` with the second
    `get_release` **outside** the try. `_extract_id` delegates to
    `extract_release_id` at `$SP/beets/metadata_plugins.py:272-278`, and
    the musicbrainz pattern is `(\w{8}(?:-\w{4}){3}-\w{12})` at
    `$SP/beets/util/id_extractors.py:20`. Measured: all 303 committed
    lookup ids match that pattern.
16. **C14, the per-id fan-out order.** `candidates` at
    `$SP/beets/metadata_plugins.py:442-446`; `albums_for_ids` at
    `:250-259`; the `as_completed` loop at `:91-108`. All three committed
    profiles pin `plugins: musicbrainz`, so the corpus is single-source
    and ordering is the response's.
17. **C15's failure map.** 404 on a lookup caught at
    `$SP/beetsplug/musicbrainz.py:777-781`; 404 on a search reaching
    `_search_api`'s handler at `$SP/beets/metadata_plugins.py:419-427`;
    `HTTPNotFoundError` registered in `explicit_http_errors`
    (`$SP/beetsplug/_utils/requests.py:161-163`) and mapped in
    `handle_http_error` (`:188-202`), which only ever sees
    `requests.exceptions.HTTPError`, so a `RetryError` is never mapped.
18. **C16, `search_query_ascii` is structurally dead for MusicBrainz.**
    `_search_api` unidecodes `query` at
    `$SP/beets/metadata_plugins.py:412-413`; the plugin returns `query =
    ""` at `$SP/beetsplug/musicbrainz.py:750`; `get_search_response`
    reads only `params.filters` and `params.limit` at `:755-762`.
19. **C17, the whole request census, measured off the tree by both
    lenses independently.** 215 cases; 351 logged requests, all `GET`;
    303 lookups; 48 searches; 167 cases with no search; one distinct
    `inc=`; one distinct `limit=` (`5`); `fmt=json` on all 351; 46
    distinct search `path_query` values; 43 `artist%3A%28` and 5
    `arid%3A%28`; zero cases carrying an `extra` key; per-case shapes
    exactly `(0,1)×164, (0,2)×3, (1,1)×26, (1,2)×4, (1,5)×9, (1,6)×9`.
    Also: no case logs the same release MBID twice, so "distinct MBIDs"
    and "logged lookups" are both 303, which is what Task 7's per-case
    issued-count assertion needs; the 25 multi-lookup cases the plan
    names are `3+4+9+9`.
20. **C18, Go's `net/url` reproduces the committed canonical form.**
    Probed offline against the maximal-`inc` lookup, the escaped-colon
    `arid` search, and the escaped-hyphen `artist` search:

    ```
    /ws/2/release/x?fmt=json&inc=artists%2Bmedia%2B...%2Bgenres%2Btags
    /ws/2/release?fmt=json&limit=5&query=release%3A%28frank+zappa%5C%3A+cucamonga%29+arid%3A%2889ad4ac3%5C-39f7%5C-470e%5C-963a%5C-56509c546377%29
    /ws/2/release?fmt=json&limit=5&query=release%3A%28bob+dylan+%5C-+the+bootleg%29+artist%3A%28bob+dylan%29
    ```

    each byte-identical to a committed `path_query`. `url.Values.Encode`
    sorts by key while `cassette.canonicalizePathQuery`
    (`internal/harness/cassette/cassette.go:210-228`) sorts whole pairs;
    for `fmt`/`inc` and `fmt`/`limit`/`query` the two orders coincide, as
    the plan says, and the coincidence is correctly recorded rather than
    relied on silently.
21. **The whole query contract set reproduces the corpus, twice.**
    Lens 1 replayed `get_most_common_tags` + `va_likely` +
    `get_album_criteria` + `get_search_query_with_filters` +
    `format_search_term` + `urlencode` + the cassette pair sort against
    `input.items` for every searching case: `search cases: 48
    mismatches: 0`. Lens 2 reimplemented C6 through C10 independently in
    Python over the same 48 and also got 0 mismatches. Task 7's search
    half is achievable exactly as specified.
22. **All 303 committed lookup responses decode, validate, and
    translate.** Full `musicbrainz.Release` decode plus
    `(*Release).Validate` over every one: `decode errors: 0 validate
    errors: 0`, and `maxTracks: 470`, which is Task 0 item 9's expected
    figure against `browseMaxTracks` 500
    (`internal/harness/conform/translate.go:22-29`). Lens 2 additionally
    ran all 303 through `TranslateAlbum` under `DefaultTunables` with 0
    failures. Nothing in the tree had decoded all 303 through the real
    type before.
23. **R4 and R5, the empty migration and its offline proof.** Probed by
    loading each of the 215 committed cases through `cassette.LoadCase`
    and re-writing it through `(*Case).Write` into a temp dir: `cases:
    215 bad: 0`, byte-identical every time. So the proposed
    `TestEveryCommittedCaseRoundTripsThroughTheCurrentSchema` passes as
    written today, and `Extra`'s `omitempty` nil value means the reshape
    moves no committed byte. Every committed case is written through
    `Case.Write` (`cmd/harness/{capture,freeze,seeditems,synth}.go`),
    which is why.
24. **The schema constant is genuinely single-sourced.** The only pins
    are `internal/harness/corpus_test.go:160` and `corpus/manifest.json:2`,
    plus `cassette.Manifest`'s field; no Python file reads
    `schema_version` and `oracle.Reconcile` checks dependency pins rather
    than the schema version, so `make derive` is unaffected and "move it
    rather than duplicate it" is achievable as written.
25. **R7, `Case.Extra` has no producer.** The only writes in the tree are
    `internal/harness/cassette/cassette_test.go:67` and `:238`. `synth`'s
    `OpExtra` (`internal/harness/synth/synth.go:143`) is an extra *item*
    override and is unrelated. `internal/harness/analysis`'s
    `releaseLookups` reads `c.Requests` only (`album_info.go:248-264`),
    so the reshape needs no `analysis` edit.
26. **R2's premise about `checkNoTimeNow`.** It flags only a selector
    call named `Now` on the `time` package
    (`internal/harness/conform/arch_test.go:99-113`) and roots at
    `internal/engine` (`:152-157`), so `time.NewTicker` passes as written
    and the invariant needs no amendment. (m10 adds only that the walk
    also covers `_test.go`; M-c adds only that the client is inside the
    same walk.)
27. **R9's independence argument holds structurally.** `mbrequest` is
    what `cmd/harness/capture.go:279` and
    `internal/harness/synth/synth.go:41` use to *produce* a committed
    request, both through `mbrequest.CanonicalizeQuery` with
    `mbrequest.MaximalReleaseInclude`; the engine importing that constant
    would make Task 7's grading self-derived. 3b's "worth revisiting if a
    fourth appears" trigger is at `docs/HISTORY.md:352-359`, the trigger
    is met, and the inline `"/ws/2/release/"` literals at
    `capture.go:279` and `synth.go:41` are real and correctly filed as
    not-3c's.
28. **R12's tenth declared gap is real.** No committed profile writes a
    `musicbrainz:` block; `corpus/profiles/estate.yaml:10` says so in as
    many words, and the only `musicbrainz:` stanza the derivation ever
    sees is the adapter's `host` / `https` / `ratelimit` append at
    `oracle.go:247`. `RECORDING_INCLUDES` and `get_recording` exist
    (`$SP/beetsplug/_utils/musicbrainz.py:68-74`, `:692-697`) and no
    committed case issues one.
29. **R14's strip divergence is exactly the four code points, and only
    those.** Swept all valid code points in both runtimes: Python's
    `str.isspace()` set is Go's `unicode.IsSpace` set plus U+001C,
    U+001D, U+001E, U+001F, with nothing Go-only. So `unicode.IsSpace(r)
    || (r >= 0x1c && r <= 0x1f)` is the exact predicate. And `İ`
    (U+0130): Go `strings.ToLower` yields `"i"` (U+0069), CPython yields
    `"i" + U+0307`, so the asserted divergence row is correct. The sweep
    file carries 56 data rows. (m11 corrects only how the other 55 are
    characterized.)
30. **R15's beets side.** `get_penalty`'s class-literal fallback at
    `$SP/beets/metadata_plugins.py:160-169` and `:180`; the
    `source_weight`-first resolution at `:190-195`; the init wiring at
    `:199-203`. Both `conform` compensations exist as described,
    `internal/harness/conform/component.go:673` and `:721`, and both are
    behavior-neutral to remove, since `tunablesWithOverrides` starts from
    `DefaultTunables()` whose `DataSourceMismatchPenalty` is already 0.5.
31. **R16's premise.** `internal/config/keys.go:19-25` still carries
    `ownedWall` as a slice literal, `Threshold.Default` is a string with
    `ParseInt` the only parser
    (`internal/config/thresholds.go:43-60`), `Config.AdminEmail` is a
    required wall row parsed as an RFC 5322 address
    (`internal/config/config.go:60-64`, `keys.go:23`), and
    `.gitleaksignore` pins `internal/config/keys_test.go:generic-api-key:66`.
    So the registration deferral and the "no new wall row, no new secret"
    claim both hold.
32. **`Release` needs no `_normalize_data` pass.** 3b's decoder names the
    wire keys directly and says so
    (`internal/engine/musicbrainz/release.go:15-33`, `ArtistCredit
    []CreditPart \`json:"artist-credit"\`` at `:44`), with the relations
    regrouping half in `GroupRelations` (`relations.go:28-40`) that
    `PseudoReleaseTarget` calls (`pseudo.go:25`). The plan is right to
    specify no normalization step in the client.
33. **`errors.Is` survives `json.Unmarshal` through a custom
    unmarshaler.** Probed on the real toolchain (go1.27.0), with the
    unmarshaler nested two levels inside slices: `err=position true:
    malformed errors.Is=true`. Task 6's `wireInt` wrapping and its named
    mutation 8 work as written.
34. **The `go.mod` claim.** cobra `v1.10.2` plus `mousetrap` and `pflag`
    indirects at `go 1.27.0`, nothing else. No `golang.org/x/time` and no
    `golang.org/x/text`, so the no-new-dependency constraint is
    checkable, and every construct the plan's snippets use (`for range
    n`, `slices`, `errors.Join`, `cmp.Compare`) compiles.
35. **The corpus size figures.** 351 response files totalling 77.6 MB,
    largest single body 8,893.93 KB (9,107,383 bytes), both exactly as
    the plan states.
36. **The anchor citations in the harness tree.** `Case.Extra` at
    `cassette.go:47`; the `Requests` validation loop at `:152-167` and
    the `Extra` loop at `:169-176`; `LoadCase`'s doc at `:117-121`;
    `TestEveryResponseFileIsConsumed` at `corpus_test.go:127-152`;
    `wantManifestSchemaVersion` at `:154-160`;
    `TestManifestParsesAndMatchesSchemaVersion` at `:162-174`;
    `stub.Server` at `stub.go:40-49`, `New` at `:51-77`, `ServeHTTP` at
    `:79-91`, the accessors at `:93-105`; `mergeProfile` at
    `internal/harness/conform/translate.go:193-204` and
    `releaseLookupPrefix` at `:20`.
37. **The committed search response shape.** All 48 carry exactly the
    top-level key set `count`, `created`, `offset`, `releases` (key names
    read mechanically; no body entered context). This is the measurement
    M-e asks the plan to record.
38. **The header's two corrections of the planning inventory.** Measured:
    15 of 215 cases carry `comp` on an item, and the same 15 carry `comp:
    true`. `FIELDS_TO_MB_KEYS` is eight entries.
39. **The six contested-surface rows the plan calls "written" exist
    verbatim** in
    `docs/specs/2026-09-03-parallel-build-approach.md:160-175`,
    including the `{cassette,stub,analysis}` row that names 3c's change
    by name, and `corpus/manifest.json` genuinely has no row. (m22 adds
    only `corpus/README.md` to the same owed list.)
40. **Task 0's expected answers match the live state.**
    `docs/STATUS.md:25-33` reads "lane-1 (code, main tree): PREP for
    **pass 3c**", names `wf_ede6e00d-bee`, records no pid, and has lane-2
    BLOCKED with the doc lane IDLE. Items 1 and 2's expected answers
    match too.
41. **The ceiling arithmetic and the pass shape.** 2.7M over ten executed
    units is 270K each, nine units is 2.4M, plus one unit for the two
    `opus` tasks gives roughly 3.0M, plus a 2.1M prep bound gives 5.1M
    against 6M, and the 4.8M flag is 80%. The 12M phase-1 reservation
    restates 3a's and 3b's figure unchanged
    (`plans/2026-09-04-rung-3a-scoring-core.md:221`,
    `plans/2026-09-04-rung-3b-translation.md:275`). Eight tasks plus a
    pre-task takes workflow mode, the split point is named at the Task 2
    / Task 3 boundary with the number that would trigger it, and every
    deliverable count is at or under four (m21's Task 6 excepted).

---

## Verdict

Three blockers, eight majors, twenty-six minors, four rejected citation
claims. The blockers are the two hand-driven concurrency tests that
cannot be deterministic as written and the dispatcher rule that provably
violates the constraint the plan quotes as its own justification. All
three sit in the pass's two `opus` tasks and none is visible to the
proofs the plan ships, which is the shape that has to stop a dispatch.

Everything the plan says about pinned beets, about the corpus, and about
the harness surfaces it will touch checked out. Both lenses ran probes
designed to break the plan's load-bearing premises (the empty migration,
the canonical encoding, the search construction, the decodability of all
303 committed lookups) and every one came back clean, which is a better
foundation than either 3a's or 3b's plan had at review time. Fold the
findings and the pass is ready.
