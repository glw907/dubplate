# Rung 2a plan adversarial review, 2026-09-02

Plan: `docs/superpowers/plans/2026-09-02-rung-2a-cassette-substrate.md`
(committed at `fe53a1e`)
Spec: `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified, not under review), ladder entry 2 and
rulings 1-3
Governing research: the matcher mechanics study
(`research/2026-09-02-beets-matcher-mechanics.md`), the corpus report
(`research/2026-09-01-beets-conformance-corpus.md`), the evolution
report's U6/U7/U8 (`research/2026-09-02-beets-evolution-lessons.md`)
Lens: oracle integrity against the study's Q1 purity hazards and Q5
HTTP seam, determinism, the capture-derive loop's convergence, schema
completeness against what 2b and 2c will need, repo and gate fit
against the tree as it sits today, task sizing.

Every claim below was checked against a file in this worktree, a
citation in the mechanics study, or a command run here. Where the plan
is right about a mechanism that looks risky, that is recorded in
**Verified sound** at the end, because four of its load-bearing
assumptions hold and the implementer should not spend a turn
re-deriving them.

---

## Blockers

### [blocker 1] The applied-tag half of the oracle is not defined against any API the mechanics study read, and the "files or item fixtures" hedge quietly decides whether rung 8 can be proven at all

`Expected.AppliedTags map[string]any // the tag set beets wrote,
flattened` (Plan:128) is one of four fields the shim must emit
(Plan:193-195). The plan's stated guide is "the mechanics study's
pipeline map" (Plan:195-196).

The study has no pipeline map for this. Its Q2 pipeline map runs from
`_string_dist_basic` to `_recommendation` and stops; the only mention
of the apply side anywhere in the document is one parenthetical in Q4
naming `artist_credit`, `original_date`, `per_disc_numbering`, and
`overwrite_null` as "application-time only, hooks.py raw_data/item_data
paths" (study:464-465). The spec knows this: rung 8's entry says its
plan "opens with the write-side close read the mechanics study scoped
out (the apply-metadata field mapping and the mediafile layer)"
(Spec:598-600). So Task 2 asks an implementer to extract, on `opus`,
from a surface the research explicitly deferred to rung 8, with the
plan asserting research coverage that does not exist.

The consequence is not academic, because "the tag set beets wrote"
has two different referents and the plan picks neither:

- **`apply_metadata(match.info, match.mapping)` output**: beets field
  values set on in-memory `Item` objects. No files needed. Synthetic
  `Item(title=..., artist=..., length=...)` objects are enough, and
  `length` is settable directly, so the whole verdict-and-distance half
  of the oracle needs no audio at all.
- **What `item.write()` puts in the file**: mediafile's mapping from
  beets fields to format-specific tags (Vorbis comments for FLAC),
  which drops some fields, renames others, and is the only thing that
  answers rung 8's question.

Spec:593-594 makes rung 8 "Pure-Go FLAC tag writing for the server
scope, **proven against the corpus's expected applied tag sets from
rung 2**." That proof needs the second referent. Choosing item
fixtures (the cheaper branch of Plan:190-191) makes it structurally
impossible, and the failure surfaces six rungs later as "the corpus
does not contain what rung 8 needs," at which point every case must be
re-derived.

Three further things ride on the same undecided sentence and none is
named:

- Real files mean real audio. Nothing in any task produces it, nothing
  sizes it, and `corpus/` gains binary blobs no `.gitignore` line
  covers. (`ffmpeg` is present on this workstation, so generating
  silent FLACs of pinned durations is cheap; it is just unassigned.)
- The estate profile sets `import: {move: yes, write: yes}`
  (`config/beets-config.yaml:13-14`). A shim that drives a real import
  over real files **renames and moves them**, mutating the case
  directory it was invoked on. That is the repo's own recorded fact
  ("beets renames every file", CLAUDE.md).
- The estate profile also sets `original_date: yes` and
  `per_disc_numbering: yes` (`config/beets-config.yaml:29,31`), both on
  the study's application-time-only list. They move `AppliedTags` while
  moving no verdict, so the applied-tag record has determinants outside
  every surface the study enumerated.

**Fix:** decide the referent in the plan, not in the implementer's
head. The honest split is to derive verdict, action, and distances at
2a from synthetic items with no files, and to state explicitly that the
file-level applied tag set is deferred with a named home (2c or a rung
8 precondition), so rung 8 inherits a stated debt instead of a
surprise. If the file-level set is wanted now, Task 2 must additionally
own audio-fixture generation, a copy-into-tempdir discipline so the
import never touches `corpus/`, and a read-back step through mediafile.

### [blocker 2] `Case.Config` names beets config profiles that no task creates

`Config string // beets config profile name ("estate", "stock")`
(Plan:106). Task 2's adapter "renders a beets config from the case's
profile" (Plan:187-188). Task 1 creates the schema, `corpus/README.md`,
and `corpus/manifest.json` (Plan:90-95). Task 2 creates the stub, the
oracle, the shim, and the venv spec (Plan:165-174). Task 3 creates the
capture tool. Task 4 creates cases and the gate test.

Nothing creates the profiles. There is no file, no format, no location,
and no statement of what "estate" means.

That is not a naming nit, because the profile is verdict-moving in a way
the study is emphatic about. Study:470-474: "The estate config sets none
of the `match:` or `musicbrainz:` keys, so the whole surface runs at the
defaults; its verdict-moving lines are `plugins: musicbrainz ...`
(single source) and the quiet-mode trio." The candidate profile source,
`config/beets-config.yaml`, cannot be used literally:

- `directory: /srv/music/library` and `library: /srv/music/state/library.db`
  (lines 11-12) are box paths that do not exist on the workstation, and
  a literal box path in a corpus artifact is a universality violation
  against Plan:68-70.
- `plugins: musicbrainz fetchart embedart replaygain` (line 47) pulls
  in `replaygain: {backend: ffmpeg}` and `embedart: {auto: yes}` (lines
  51-56), which are inert for matching (study:620-628) but not inert
  for an import that writes.
- `group_albums: yes` (line 24) operates above the matcher; the corpus
  deliberately tests "from the (items, responses) boundary, below the
  importer's grouping logic" (study:612-614), so this key is dead in
  the profile and its presence will mislead.

Ruling 3 (Spec:713-716) makes the stock stratum tier-one gated, so the
two-profile mechanism is a launch-bar dependency, not a 2c
convenience, even though the stratum itself is correctly deferred
(Plan:25).

**Fix:** give Task 1 or Task 2 the profile artifact explicitly: where
the profiles live (`corpus/profiles/estate.yaml`, `stock.yaml` is the
obvious answer), what is derived from `config/beets-config.yaml` versus
overridden for the harness (paths, the stub host block, the plugin
list), and the rule that a change to the estate profile is a
corpus-revalidation event exactly like a beets bump.

### [blocker 3] beets caches config at first use per process, so a shim that derives more than one case, or more than one profile, silently returns the first case's answers

Study Q1.8 (study:126-131), with citations: `Distance._weights` is a
cached class property (distance.py:133-140), the track-length knobs
`track_length_grace` and `track_length_max` are `@cache`d
(distance.py:364-373), and `get_penalty` is `@cache`d
(metadata_plugins.py:159-169). The study's own closing clause is the
warning: "irrelevant across processes but **worth knowing when driving
beets in-process**."

The shim drives beets in-process, by design (Plan:193-194, "the autotag
API directly, not `beet import` CLI"). The plan never says one process
per case. Two shapes an implementer will reach for naturally both
break:

- Batch derivation (one shim invocation, N cases) is the obvious
  optimization once the corpus reaches 2c's 150-300 cases, and it is
  the shape the capture-derive loop invites, since the loop re-derives
  repeatedly.
- Any run mixing the "estate" and "stock" profiles in one process
  scores the second profile's cases under the first profile's weights
  and thresholds. Ruling 3 makes exactly that mix a gate input.

The failure is silent, deterministic, and reproducible. Two derivations
in the same batch order agree byte for byte, so the plan's determinism
proof (Plan:205-206) certifies the poisoned record as sound. This is
precisely the hazard Task 2's own preamble names: "an error here
poisons every expectation silently" (Plan:161-162).

**Fix:** state the invariant as a contract, not a convention. One shim
process per (case, profile), asserted by the adapter, with the shim
echoing the effective `match` weights and thresholds it actually ran
under so the adapter can compare them to the rendered profile and fail
loudly on drift.

### [blocker 4] The interface block omits the three types that carry the content, and the schema has no home for the request log two later tasks assert against

Plan:97 says "later tasks and passes build against these exactly."
Plan:99-134 then defines `Case`, `Response`, and `Expected` and leaves
undefined:

- **`Input`** (Plan:105). The type holding the items. Whether an item
  is a beets field bag, a file reference, or both is the same question
  blocker 1 leaves open, and it is Task 1's deliverable, not Task 2's.
- **`Candidate`** (Plan:127). Everything downstream reads it: the
  threshold-proximity flag needs the best and runner-up distances, U8's
  shared-candidate-subset grading needs a stable key (the study names
  the natural one, `(data_source, album_id)` at match.py:414), rung 3
  grades against distances at 1e-12 relative (Spec:535-536).
  "distances, in beets' final order" is a comment, not a type.
- **`Manifest`** (Plan:133 signature, Plan:136-139 prose). Prose lists
  the pins; no field set exists.

Task 1's stated deliverable is "schema types and IO" (Plan:154). The
plan hands the implementer a block that defines under half of it while
declaring it exact.

The sharper gap is the request log. Task 3's capture-derive loop is
driven by "the shim's request log from a dry derivation" (Plan:233-234).
Task 4's gate test asserts "zero misses against the recorded request
log" (Plan:256-257). `Case` has `Input`, `Config`, `EvalDate`,
`Responses`, `Expected`. There is no request log field, no statement of
whether the log is a case member or a derivation by-product, and no
statement of whether it is committed. Task 4's acceptance criterion
therefore references an artifact the schema Task 1 delivers cannot
hold.

**Fix:** define `Input`, `Candidate`, and `Manifest` in the block, and
decide the request log's home. Committing it as a case member is the
right call: it is the U8 conformance surface ("the MB query
construction is itself conformance surface... rung 3 must build the
same lookups pinned beets builds", lessons:420-423), which makes it an
expectation rung 3 is graded against, not scaffolding.

### [blocker 5] Task 2's determinism test puts Python and beets inside `go test ./...`, which the plan forbids and CLAUDE.md bans working around

Three statements collide.

1. The venv is "a workstation dev dependency, documented, **never part
   of the repo gate**" (Plan:48-50).
2. "**Determinism proof:** `Derive` twice on one case yields
   byte-identical output; **the test runs it and diffs**"
   (Plan:205-206), in `internal/harness/oracle/oracle_test.go`
   (Plan:169).
3. `Makefile:22` defines `check: vet lint test tidy-check`, and
   `test:` is `go test ./...`. Any `_test.go` under `internal/` is in
   the gate. `scripts/check.sh:29-32` runs `make check` as leg one.

So the determinism test is in the gate, and the gate then requires uv,
a provisioned venv, and beets 2.13.1 on every machine that runs
`scripts/check.sh`. The usual escape is a `command -v uv || t.Skip`
guard, and CLAUDE.md closes it explicitly: "A check either gates or does
not exist; never guard one with `command -v || skip`." A skip-guarded
oracle test is also the exact silent-decay shape
`scripts/vale-comments.sh:6-8` was written to reject.

The second half of the collision is `git status`. Plan:82-83 requires a
clean tree at every task end. A uv venv materializes a directory
(`tools/.venv` or `.venv`), and running the shim creates
`tools/__pycache__/`. `.gitignore` covers neither, and no task's file
list modifies `.gitignore` (Plan:90-95, 165-174, 216-219, 250-259).
This is the same shape as rung 1's blocker 1, one rung later.

**Fix:** pick the boundary and write it down. The clean version:
derivation tests live behind a build tag (`//go:build oracle`) with a
`make derive` target that runs them, `scripts/check.sh` never invokes
it, and the plan states that the corpus's frozen records, not the
derivation, are what the gate protects. That is also what Spec:495
means by "the frozen-verdict fast gate with no Python in it." Add the
venv and `__pycache__` lines to `.gitignore` in Task 2's file list.

---

## Majors

### [major 1] The determinism proof is Go-normalized, so it cannot fail for the reasons the study says derivation is nondeterministic

Plan:78-81 makes determinism "the product" and names the smoke tests'
double-derivation as the proof (Plan:205-206). Verified here:
`encoding/json` sorts map keys on marshal, so `AppliedTags` serializes
identically regardless of what order Python emitted it in, and float64
values re-marshal through Go's own formatter both times.

That is good hygiene and a bad test. Once the adapter decodes the
shim's JSON into Go types, the record is Go-normalized, and a
double-derivation diff can no longer observe: Python dict emission
order, Python float repr, or the one genuine set-order leak the study
found. Study Q1.6 (study:100-107): `extra_items` / `extra_tracks` come
from set differences sorted by `(disc, track, title)` and
`(index, title)` (match.py:77-80), and "two extras with identical sort
keys keep arbitrary relative order (PYTHONHASHSEED / address
dependent)." Two shim invocations are two processes with two hash
seeds. The plan never pins `PYTHONHASHSEED`.

Today extras are not in `Expected`, so nothing leaks; the moment 2b or
rung 3 needs them (and it will, since `missing_tracks` and
`unmatched_tracks` counts are distance inputs), the gate acquires an
intermittent failure whose cause is three documents away.

Relatedly, the same constraint bullet lists "multi-source interleaving"
as something the plan pins (Plan:80), and **the single-source
constraint is never stated anywhere in the plan.** The study is direct:
"The conformance corpus should stay single-source" (study:66-67), and
with two or more sources `as_completed` ordering makes candidate order
timing-dependent (metadata_plugins.py:88-110).

**Fix:** set `PYTHONHASHSEED=0` in the adapter's shim environment and
say so; add "every profile loads exactly one metadata source, asserted
by the adapter" to the global constraints; and make the determinism
test diff the **shim's raw stdout**, not the Go-normalized record, so
it can actually fail.

### [major 2] `faketime` is the plan's preferred date-pinning mechanism, it is not installed on this workstation, acquiring it is the last-resort software tier, and the acceptance test as written is circular

Plan:199-204: "prefer `faketime` around the shim; if unavailable on the
workstation, an equivalent mechanism chosen and documented."

Checked here: `command -v faketime` returns nothing and
`rpm -q libfaketime` reports "package libfaketime is not installed."
This is Bluefin DX on bootc/ostree, where an RPM that is not in Homebrew
means `rpm-ostree` layering, which CLAUDE.md ranks last resort and which
costs a reboot. The plan's preference orders the most expensive
available option first.

The mechanism question the brief asks resolves cleanly in faketime's
favor on one axis and against it on two others:

- **Call-time versus import-time**: the study puts
  `datetime.date.today().year` inside the year-penalty branches
  (distance.py:477, 487-489), so it is read per evaluation. LD_PRELOAD
  interception works for a CPython process. That half is sound.
- **Silence on failure**: if the preload does not take (a wrapper, a
  container, a statically linked interpreter), the process runs at real
  wall clock with no error. The acceptance criterion, "derives identical
  `Expected` under two different wall-clock dates" (Plan:202-204), does
  not catch it, because both "wall-clock dates" would be produced by
  the same faketime that failed to apply. The test is circular unless
  the two dates come from genuinely different real clocks, which nothing
  arranges.
- **Deadlock risk**: the study documents an unconditional 0.25s
  `RateLimitAdapter` floor between any two requests
  (requests.py:89, 106-135; study:537-542). A frozen-clock faketime
  invocation makes the elapsed-time check never satisfy. Offset-mode
  faketime is fine; freeze mode hangs the derivation.

A hermetic alternative exists and is strictly cheaper: the shim already
runs inside beets' process and already imports its internals, so
patching `datetime.date` for the derivation is three lines under the
shim's own control, with no external binary, no ostree layering, and no
preload silence.

**Fix:** invert the preference, and make the mechanism self-verifying
either way: the shim reports the effective date it observed in its JSON
output, and the adapter fails if it does not equal `Case.EvalDate`.
That single assertion turns a circular acceptance test into a real one.

### [major 3] The U7 boundary is drawn around the adapter, but the shim is the code that actually knows beets

Plan:169-170 names `internal/harness/oracle` "the ONLY code that knows
how to invoke beets — the U7 boundary." Plan:171-174 then puts
`tools/derive_oracle.py` outside it, and Plan:193-198 gives that file
the deepest beets knowledge in the repo: the autotag entry points, the
candidate structures, the recommendation enum, the apply path.

U7's falsifiable criterion is explicit: "when the differential is
pointed at a beets 3.x, **only the adapter package should need edits**"
(lessons:411-412). Under this layout a beets 3.0 bump edits `tools/`
first. Since beets 3.0 is a compatibility-shedding release whose
targets are exactly the plugin and metadata-source layers the shim
touches (lessons:396-410), this is the bump the lesson was written for.

The Go adapter's real job is small: render a config, start a stub,
launch a subprocess, parse JSON. That is not where a beets major
version lands.

**Fix:** move the shim under `internal/harness/oracle/` (Go tooling
happily carries a non-Go file beside its package; nothing in the gate
walks it) and restate the boundary as "the `oracle` package, shim
included." Then U7's criterion is true as written instead of true as
hoped.

### [major 4] `Expected.Action` is hand-transcribed from beets' source, which the plan's own constraint forbids

Plan:76-77: "Every expectation in the corpus is derived by executing
pinned beets; nothing is hand-transcribed from test assertions or
docs."

Plan:126: `Action string // apply | skip (quiet-mode mapping)`.
Plan:193-194: the shim "drives beets' autotag API directly (not
`beet import` CLI)."

Those cannot both hold. The quiet-mode mapping lives in
`_summary_judgment` (session.py:328-360) and the surrounding session
code (session.py:64-68, 121-125), which is the importer, not the
autotag API. A shim that stops at `tag_album` has no `Action` to
report, so the field can only be produced by the shim reimplementing
"strong applies `candidates[0]`, everything else takes
`quiet_fallback`" in Python. That is a hand transcription of the exact
kind the constraint bans, and it is transcribed from the layer the
study itself flags as config-dependent: `quiet_fallback`,
`none_rec_action`, and `import.timid` are all part of the
verdict-moving surface (study:456-459), plus
`import_task_before_choice` can preempt outright (study:336-337).

**Fix:** either import and call beets' own `_summary_judgment` from the
shim so the mapping is executed rather than restated (it is a module-
level function; U7 already prices the coupling), or drop `Action` from
2a's `Expected` and compute it in Go from `Verdict` plus the profile,
documented as a dubplate-side derivation rather than an oracle record.
Do not leave a hand-written Python branch labeled as derived.

### [major 5] Raw response bodies are not raw after a Go round-trip, which breaks Task 1's byte-identity contract and U6's untouched-capture requirement

Plan:120 stores bodies as `json.RawMessage`, and Plan:145-146 requires
"round-trip Write-then-LoadCase yields identical bytes." Plan:236-237
requires the capture tool to write "raw bodies untouched."

`encoding/json` compacts a `RawMessage` on marshal and HTML-escapes it
with the default encoder. Verified in a scratch module here:

```
body in:   {NL  "artist": "Simon & Garfunkel <x>"NL}
body out:  {"artist": "Simon & Garfunkel <x>"}
url in:    /ws/2/release/abc?fmt=json&inc=artists+media
url out:   /ws/2/release/abc?fmt=json&inc=artists+media
```

(`NL` marks a newline in the input.) The body is compacted, losing its
newlines and indentation, and then every `&`, `<`, and `>` is escaped,
in the body and in the canonical URL alike.

Ampersands are common in MusicBrainz artist and release titles, and the
study makes `&` semantically live: `SD_REPLACE` rewrites `&` to `and`
inside `string_dist` (distance.py:44, 87-89). The escaped form decodes
back to `&` on load, so no verdict moves; what breaks is the contract
Task 1 states, the U6 promise that "the fixture format stores the raw
response so fixture readers, never the fixtures, decide which fields
exist" (lessons:392-395), and reviewability of a corpus whose whole
value is being auditable by eye.

**Fix:** store each response body as its own file written with
`os.WriteFile`, byte for byte as the wire delivered it, with the case
JSON holding a filename. That satisfies byte-identity trivially,
keeps the corpus diffable, and removes the escaping question entirely.
If bodies must stay embedded, the writer needs
`json.Encoder.SetEscapeHTML(false)` and the plan should say so.

### [major 6] The stated matching rule and the `Response.URL` field disagree, and neither is specified against what beets actually sends

Plan:116-118 states the rule as "exact on method plus canonical path
plus sorted query string." Plan:119 then declares
`URL string // canonical: sorted query parameters`.

A URL includes scheme and host. The capture tool writes what it fetched
from `musicbrainz.org`; the stub serves on `127.0.0.1:<ephemeral port>`
chosen per run (Plan:187). If `URL` holds the full captured URL,
nothing ever matches; if it holds path plus query, the field name is
wrong. One of the two must go, and the plan does not say which.

Three specifics from Q5 that the canonical form must survive, none of
them named in the plan:

- **`fmt=json` is forced on every request** (musicbrainz.py:615-619,
  study:481-483). A capture tool that constructs its own URL without it
  writes a key that beets' request never produces.
- **`inc=` values are joined with `+`** (musicbrainz.py:634-637,
  study:490-491). In a query string `+` is also the space encoding, so
  any canonicalizer that decodes-sorts-reencodes turns
  `inc=artists+media` into `inc=artists media` and back into
  `inc=artists%20media`. Whether canonicalization is
  encoding-preserving or decode-and-reencode is exactly the detail that
  decides whether replay works.
- **Search queries are escaped Lucene** with backslash escapes and
  parenthesized field filters (musicbrainz.py:650-685, study:496-503).
  These round-trip through percent-encoding badly if the rule is loose.

`corpus/README.md` is assigned the request-matching rule (Plan:93-94),
which is the right home; the plan owes it the three constraints above
so the Task 1 implementer does not invent a canonical form the Task 3
capture tool then has to match by luck.

### [major 7] Task 4's corpus test is described as a replay with no client, and its "zero misses" criterion contradicts the schema's own recall-extension rule

Plan:254-257: the gate test "walks `corpus/`, validates every case
against the schema, verifies manifest pins, and replays each case's
responses through the stub asserting zero misses against the recorded
request log."

There is no client at rung 2. The Go matcher is rung 3 (Spec:527), and
the shim is excluded from the gate (Plan:48-50, and blocker 5). Nothing
issues requests, so "replays through the stub" reduces to feeding the
recorded request log into a stub built from the same case's recorded
responses and observing that they agree. That is a static consistency
check on two files written together, restated as a replay. It is
worth having, and it is not a replay, and describing it as one invites
a test that cannot fail. The repo's own rule is "no vacuous assertions
— prove tests can fail" (CLAUDE.md).

Worse, "zero misses" is only half a consistency check, and the other
half is contradicted by the schema. Plan:117-118 makes
"recall-extension responses (requests pinned beets never issued) legal
members," and Plan:141-143 keeps them from day one per U8. So an
unused response is legal, which means the test cannot simply assert
request-set equality, and the plan gives no way to tell a legitimate
recall-extension response from a stale one left over by the
capture-derive loop.

**Fix:** state the assertion in its true form. Every request in the
log resolves to exactly one response; every response is either
consumed by the log or explicitly marked as a recall extension in the
case; the manifest pins parse and agree with the schema version. Add
the negative proof the repo requires: a deliberately corrupted case
fixture that the test rejects.

### [major 8] `LoadManifest(path)` cannot perform the version reconciliation the plan assigns it, and putting that check in the gate pulls Python back into the gate

Plan:137-139: "Loading a corpus whose manifest disagrees with the
venv's installed versions is an error, not a warning." Plan:133:
`func LoadManifest(path string) (*Manifest, error)`. Plan:150-151 puts
"manifest version mismatch errors name the disagreeing component" in
Task 1's behavior contract, which makes it a Go test in `go test ./...`.

`LoadManifest` takes a path. It has no venv, no way to interrogate the
installed packages, and no business shelling out to `uv pip list` from
inside a gate leg that is supposed to be Python-free (Spec:495).

Two different checks are being conflated: parsing and validating the
manifest (pure, gate-safe, cheap) and reconciling it against a live
venv (derivation-time only, needs the venv, belongs beside `Derive`).

**Fix:** split them. `LoadManifest` parses; a separate
`(*Manifest).Reconcile(ctx, venv)` runs at derivation time and is the
one that errors naming the disagreeing component. Task 1's behavior
contract should test the first; Task 2 owns the second.

### [major 9] The schema carries none of what 2b and 2c need from the derivation, and the plan does not state the deferral

The brief's question, answered field by field against the plan's split
(Plan:15-28):

| 2b/2c needs | In `Expected`? | Notes |
|---|---|---|
| Per-component distance breakdown | No | `Distance._penalties` (distance.py:130) is what the component-test derivation tooling (Plan:16-17) turns into per-component expected values. |
| Per-key weighted distances | No | `dist[key]` (distance.py:232-238) is what rung 3's per-component conformance compares against. |
| Per-track distances | No | `dist.tracks` (distance.py:129-131, 527-530). |
| Assignment mapping | No | The tie detector (Plan:17-18, Spec:504-506) must "exclude or pin the expected mapping" (study:576-581). A mapping cannot be pinned in a schema with no field for it. |
| `extra_items` / `extra_tracks` | No | Study:100-107; also the extras counts drive `missing_tracks` / `unmatched_tracks`. |
| Threshold-proximity inputs | Partly | Best and runner-up distances suffice, and they exist only if `Candidate` carries a distance, which blocker 4 leaves undefined. |

Plan:97 says later passes "build against these exactly." Either that is
false and 2b changes the schema, or 2a should emit the full record now.
The plan says neither.

Emitting the whole thing at 2a is close to free: the shim already holds
the `AlbumMatch`, and `Distance` exposes `_penalties`, `tracks`, and
per-key weighted values without extra work. The alternative is building
the shim twice.

**Fix:** either widen `Expected` now (preferred, and it makes 2b's
tooling a consumer rather than a rebuild), or state plainly in the
split section that 2b extends the schema and that the "build against
these exactly" promise covers `Case` and `Response` only.

### [major 10] Task 2 is an L wearing an M's deliverable count

Plan:209-210 declares four deliverables: "stub with miss-recording;
adapter; shim + venv; date pinning." The actual list, reading the task
body:

1. The replay stub with miss recording and its test (Plan:165-168).
2. The adapter's config rendering from a profile (Plan:187-189) which,
   per blocker 2, includes inventing the profile artifact.
3. Item materialization, which per blocker 1 is an undecided design
   question, not a coding step (Plan:190-191).
4. Subprocess invocation, venv discovery, and JSON parsing
   (Plan:191-192).
5. The shim's verdict, recommendation, and ordered-candidate extraction
   (Plan:193-195).
6. The shim's applied-tag extraction, from a surface the study did not
   cover (blocker 1).
7. The shim's request log, which Task 3 consumes (Plan:233-234).
8. The uv environment spec plus version capture into the manifest
   (Plan:171-174).
9. Date pinning, including choosing a mechanism whose preferred option
   is not installed (major 2).
10. The determinism proof (Plan:205-206).

Ten items across two languages, containing the pass's entire
correctness-critical core by the plan's own admission (Plan:160-162),
and CLAUDE.md's rule is to "say plainly when it passes roughly four."

The natural cut is clean and follows the existing package boundaries:
**Task 2a** takes the stub, the profile artifact, and the adapter's Go
side against a hand-written fake shim (a script emitting a fixed JSON
document), which gets the Go contract and its tests landed and
reviewable on `sonnet`; **Task 2b** takes the real shim, the venv, date
pinning, and the determinism proof on `opus`, against an interface that
already exists. That also lets the shim's design questions be rejected
independently, which today they cannot be.

### [major 11] "Owner touchpoints: zero planned" is not true, because capturing against the real WS requires an identity only Geoff can supply

Plan:287-288 claims zero planned owner touchpoints, with the
MusicBrainz registration question batched to a checkpoint
(Plan:238-241). But Plan:70-73 and Plan:226-228 require the capture
tool to send an identified User-Agent,
`dubplate-harness/<version> ( <contact> )`, with the contact from a
flag or env value. Universality correctly keeps the value out of the
code. It does not conjure the value: MusicBrainz's terms require a
working contact, and there is no `DUBPLATE_HARNESS_CONTACT` in
`~/.local/secrets`, in `env/dubplate.env.template`, or anywhere else in
this repo.

Task 4 cannot produce a single smoke case without it, and a subagent
cannot invent one. Spec:519-521 also assigns "the registered
MusicBrainz application identity and User-Agent" to rung 2, which the
plan silently narrows to a maybe-later question.

**Fix:** name the contact as a pass input in the plan header alongside
the token ceiling, say where it lives (an env var sourced from
`~/.local/secrets` is the house pattern), and bundle it with the
registration question so one ask covers both. Then the touchpoint count
is one and honest, rather than zero and wrong.

---

## Minors

**[minor 1] The plan claims a raised ratelimit removes the stub-run
delay; the study says it does not.** Plan:188-190 renders the stub
config with "ratelimit raised — the mechanics study's documented seam."
The study is explicit that the floor is unconditional: "Rate limiting
is doubled up: a global `RateLimitAdapter` floor of 0.25s between any
two requests... **Against a local stub with a high configured ratelimit
the 0.25s floor still applies** — budget for it in differential-run wall
time" (study:537-542), and Spec:502-503 names that floor as a capture
tool fidelity requirement. Harmless at 2a's 5-10 cases; at 2c's 150-300
cases with 3-6 requests each it is minutes per full re-derivation, and
the plan should carry the number rather than the wrong intuition.

**[minor 2] Python enters the repo with no lint leg, and the plan's
"exemplary bar applies to both" has nothing enforcing it.** Plan:66-67
invokes `python-conventions` and claims the exemplary bar for the shim.
`scripts/check.sh` has legs for Go, shellcheck, Vale over Go comments,
compose, and systemd, and none for Python; its shellcheck loop iterates
`scripts/*` only, so `tools/*.py` is invisible to it. `.vale.ini`'s
`[formats]` maps `go = md` and its only section is `[*.go]`, so shim
docstrings face no prose gate while Go comments do. CLAUDE.md's "Stack
rules (mandatory skills by area)" names Go, shell, web, systemd, and
Cloudflare, and no task amends it for Python. Either add a `ruff` leg
and a CLAUDE.md stack-rule line, or state that the shim is reviewed by
`diff-reviewer` alone and drop the exemplary claim.

**[minor 3] Three names for two fields.** Plan:193-195 has the shim
emit "verdict, recommendation, per-candidate distances, and the applied
tag set"; the type has `Verdict` and `Action` (Plan:125-128). beets has
one concept here, `Recommendation` (match.py:269-323), plus the
importer's action. Pick two words and use them everywhere; rung 3 will
be reading this block as a contract.

**[minor 4] numpy is unpinned.** Plan:136-137 pins beets, unidecode,
jellyfish, and lap, matching the study's quadruple (study:132-135).
The study also names numpy as part of the assignment path
(match.py:12-13, pyproject.toml:30-33, study:26-28); `lap.lapjv` is
called on an `np.array` and `extend_cost=True` pads through numpy.
Recording numpy in the manifest costs one line and removes a question
at the first bump.

**[minor 5] The capture tool's CLI surface cannot express what its
prose requires.** Plan:222-223 gives one form,
`harness capture --mbid <release> --out corpus/cases/<id>`.
Plan:230-236 then requires the tool to capture search responses derived
from a case's items and to capture "any recorded misses" fed back from
a derivation. Both are case-directory-driven, not MBID-driven. Name the
second verb (`harness capture --case <dir>` consuming a miss list) so
Task 3's implementer does not bolt it onto the MBID form.

**[minor 6] The pseudo-release smoke case is the one case that cannot
be found by trying.** Plan:263-265 requires "one non-Latin release with
a pseudo-release follow-up," which needs a release whose status is
`Pseudo-Release` carrying a backward `transl-tracklisting` relation
(study:553-559). These are a small minority of MusicBrainz releases and
are not searchable by status through the plain search endpoint. Naming
one candidate MBID in the plan converts an open-ended hunt into a
fetch. Task 4's acceptance already depends on it ("the pseudo-release
case's directory visibly contains both releases' responses",
Plan:270-271).

**[minor 7] Task 3's tests cover request planning only, so nothing
verifies the capture tool's output until Task 4.** Plan:218-219 scopes
the test to "the request-planning logic (the network call itself is
exercised only manually)," which is right for the network. What is
missing is the cheap half: a test that a captured response set, written
by the tool's own writer, loads back through `cassette.LoadCase` and
matches the canonical form the stub expects. Without it, a
canonicalization disagreement between Task 1's matcher and Task 3's
writer surfaces as a mysterious miss during Task 4's captures, after
the network round trips have been spent.

**[minor 8] Two gate facts in the plan's neighborhood are already
wrong in the tree, and this pass is a good moment to note them.**
`scripts/check.sh:2-7` says `make check` covers "build, vet,
golangci-lint, test, tidy-check"; `Makefile:22` defines
`check: vet lint test tidy-check` with no build. And CLAUDE.md's safety
rails say "The gate runs gitleaks"; `scripts/check.sh` has no gitleaks
leg, though the binary is installed. Neither is this plan's defect, and
`corpus/` is the first large data directory the repo has carried, which
makes the second one newly relevant.

---

## Verified sound

Recorded so the implementer does not re-derive them.

**A test-only package at `internal/harness` is fine for this gate.**
Plan:254-255 puts `corpus_test.go` at `internal/harness/` while every
non-test file lives in subpackages (`cassette`, `stub`, `oracle`).
Verified in a scratch module on the repo's Go 1.27.0: `go build ./...`,
`go vet ./...`, and `go test ./...` all exit 0 with a directory holding
only a `_test.go` file. A second `main` package under `cmd/harness`
likewise joins `go list ./...` without incident, and `make build`
(`Makefile:5`) and `make install` (`Makefile:19`) both name
`./cmd/dubplate-server` explicitly, so the dev tool never ships.

**Nothing in the gate walks `corpus/` or `tools/` hostilely.**
`scripts/vale-comments.sh:50-56` selects `git ls-files '*.go'` minus
`_salvage/`, so JSON fixtures and Markdown in `corpus/` are invisible to
Vale; `scripts/check.sh:37-59` iterates `scripts/*` only, so a Python
file in `tools/` is never handed to shellcheck. The `.gitignore` gap in
blocker 5 is the only real interaction.

**The stub seam is exactly as the plan describes it, with no
monkeypatching.** Study Q5 (study:520-528): `musicbrainz.host` defaults
to `musicbrainz.org` with forced https and a hard 1 rps, and *any*
other host honors `musicbrainz.https` and `ratelimit`. So
`musicbrainz: {host: "127.0.0.1:PORT", https: no, ratelimit: 100}`
serving `/ws/2/...` is all that is needed. Plan:187-190 is right about
this.

**The verdict half of the oracle is cleanly extractable without an
import.** `tag_album` returns candidates already admitted through
`_add_candidate`, stably sorted by distance, plus the recommendation
(match.py:326-328, 383-468, 269-323). Verdict, per-candidate distances,
and final ordering need no importer, no library database, and no files.
Only the applied-tag half does (blocker 1). The plan's decision to
drive the autotag API rather than the CLI is correct for three of the
four fields.

**The capture-derive loop terminates, and the stub's miss recording is
what makes it safe.** The brief's hazard is real in general: a 404 on a
release lookup returns None and the candidate is silently skipped
(musicbrainz.py:777-781, study:548-550), which could let a derivation
take a short path and converge on a corpus that never exercises the
skipped lookup. Plan:167-168 defuses it precisely: the stub records the
unknown request as a miss, so a lookup beets swallowed is still visible
to the loop. The loop is monotone (each iteration captures at least one
new response and the reachable request set is finite), so it
terminates. Two supporting facts also check out: 404 is not in the
retry set (`Retry(total=6)` on 429/500/502/503/504, requests.py:78-91),
so one miss costs one request rather than seven; and a 404 on a *search*
raises rather than degrading (study:548-550), which fails loudly.

One caveat worth a sentence in `corpus/README.md`: capturing a response
can *retire* a request path rather than only adding one. If an ID
lookup succeeds and yields a strong recommendation with `timid` off,
`tag_album` returns immediately and the text search is never issued
(match.py:432-442, study:350-355), leaving earlier-captured search
responses orphaned. They are legal under the recall-extension rule, but
the README should describe the loop's output as a recorded fixed point,
not as a stable request set.

**Manifest pinning matches the study's requirement and the estate's
tooling.** Plan:136-139 pins the beets, unidecode, jellyfish, and lap
quadruple that study:132-135 names, and makes a mismatch an error
rather than a warning, which is the right stance for a corpus whose
transliteration tables sit inside the distance arithmetic
(distance.py:55-56). The uv venv choice also matches the box:
`docs/STATUS.md:57` records beets 2.13.1 installed via uv there, so the
oracle and the estate run the same install mechanism.

**The 2a/2b/2c split loses nothing from the spec's ladder entry 2.**
Checked deliverable by deliverable against Spec:489-526: the fast gate,
the version-bump differential, the tie detector, the proximity flag,
the oracle-bug ledger, the component-test tooling, the stock stratum,
the strata sizing, and the quarterly chore all have a stated home in 2b
or 2c (Plan:15-28); the U8 request-matching rule and the
recall-extension format land in 2a as the spec requires
(Spec:513-519, Plan:116-118, 141-143). The only item narrowed without
saying so is the MusicBrainz application identity (major 11).

**Go's JSON serialization is deterministic for this record shape.**
Verified here: `encoding/json` sorts map keys on marshal, so
`AppliedTags map[string]any` emits identically across runs, and float64
round-trips through Go's own shortest-representation formatter. This
holds only if the adapter decodes the shim's output into Go types
rather than passing its bytes through, which the plan should state
(and see major 1 for why the resulting test is weaker than it looks).

---

## The three I would fix first

1. **Decide what `AppliedTags` means, in the plan, before Task 2
   dispatches** (blocker 1, with blocker 2 riding along). It is the one
   finding with a consequence six rungs out: rung 8's FLAC tag writing
   is specified as proven against these records (Spec:593-594), and the
   cheap branch of Plan:190-191 makes that proof impossible while
   looking like a reasonable implementer choice. Deciding it also
   settles whether the corpus needs audio fixtures, whether derivation
   mutates case directories, and what the "estate" profile has to be,
   which is most of blocker 2 for free.

2. **State the process and gate boundaries as contracts** (blockers 3
   and 5, major 1). One shim process per case and profile, with the
   shim echoing the config it ran under; `PYTHONHASHSEED=0`;
   single-source asserted; derivation tests behind a build tag with
   `.gitignore` lines for the venv and `__pycache__`. These are all
   one-line commitments, and together they are the difference between a
   corpus whose determinism claim is proven and one whose determinism
   proof is a Go serializer certifying its own output.

3. **Split Task 2 and finish the interface block** (major 10,
   blocker 4). Ten deliverables across two languages carrying the
   pass's whole correctness-critical core is the shape the workstation
   rules exist to catch, and the split falls naturally at the Go/Python
   line: the stub, the profile, and the adapter against a fake shim on
   `sonnet`, then the real shim, venv, date pinning, and determinism
   proof on `opus`. Define `Input`, `Candidate`, and `Manifest` while
   splitting, since the second half cannot be dispatched against a
   contract that omits its inputs.

**Verdict:** the architecture is right and four of its riskiest bets
hold under inspection — the `musicbrainz.host` seam, the miss-recorded
capture-derive loop, the manifest quadruple, and a test-only harness
package in this gate — but the plan asserts research coverage for the
applied-tag extraction that the mechanics study explicitly deferred to
rung 8, names config profiles no task creates, leaves beets' per-process
config caching free to poison a multi-case derivation silently, puts
Python in a gate its own constraint keeps out, and declares four
deliverables for a task carrying ten; fix the five blockers, split
Task 2, and complete the interface block, and the substrate is worth
building on.
