# Rung 12 slice 2a plan review: domain risk

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md` at `f62a151` (draft revision 1,
pass class `tool`). Lens: domain risk (oracle grading wrong, GPL contamination, lane contention, export ahead of a
caller, drift, the production box), plus proportionality and the four planning-miss items. Reviewed 2026-10-06.
Every claim below was checked against the tree at `87b6380` (code unchanged since `0cb33ae`) or against
kin-openapi v0.149.0 in the module cache, and the bundled OpenAPI file was scanned from the conductor's scratchpad
copy (sha256 `cb54c03c…ed46bfb6`, 453,720 bytes, matching the plan).

Counts: 2 High, 3 Medium, 7 Low correctness findings; 2 over-ceremony items; 1 OWNER FORK.

## What holds

- **The production box.** Nothing in the plan contacts it. `cmd/harness` is not built by `scripts/deploy.sh`, which
  builds only `dubplate-server`, and every container binds loopback on the workstation.
- **GPL contamination, at the code level.** The prior-art table declines both GPL Go clients, implementers read
  only Navidrome's Apache-2.0 docs site and gonic's README, and kin-openapi's six transitive modules carry
  permissive licenses. Recorded responses are server output over a fixture dubplate authored, not GPL code. The
  bundled OpenAPI file holds no auth query string and gitleaks finds nothing in it, so committing it under R3 does
  not trip the gate.
- **R1 (worktree).** The rule's stated reason is the untracked beets venv (parallel-build approach `:255`), which
  this harness never uses. The reading is sound.
- **R6 (boundary as a Go test)** is the existing `arch_test.go` invariant pattern, including the fire-on-fixture
  proof.
- **The light lane.** `--scope-changed` derives the change set from `git merge-base master HEAD`
  (`scripts/check.sh:178-184`) and runs the rendered leg only for `web/`, `internal/serve/`, `internal/auth/`,
  `internal/identity/`, or the serve shell's files (`:233-240`). No 2a task touches one.

## Correctness findings, ranked by consequence

### R1. High. Ledger rows are not scoped to the comparison pair, and the oracle-against-oracle proof forces broad suppression rows

`plan:471-487`, `plan:551-554`, `plan:16-17`.

**Defect.** An `active` row matches on "format and a path pattern into the response" plus one named oracle
(`:471-473`). Nothing keys the match to the pair of sources being compared. The 2a proof grades Navidrome against
gonic and demands zero unledgered findings (`:551-553`). Those two servers disagree on far more than the folder
view: field sets per song and album, the `path` value, id shapes, cover-art id forms, and list order all differ. The
spec never asks for an oracle-against-oracle grade. Its oracle tier "diffs dubplate's responses against them" (spec
`:274-278`). The four classes (`:476-478`) have no home for "the two oracles disagree". A `volatile` row still
compares presence (`:478`), so it cannot quiet a field one oracle emits and the other omits. The Task 8 implementer
must then either fail the proof or stretch `by-construction` or `volatile` over ordinary oracle disagreements. Either
stretched row names Navidrome and a path. In slice 3, when dubplate is graded against Navidrome, that row matches the
same path and suppresses a real dubplate divergence. Staleness (`:483-484`) catches only rows that match nothing.
It cannot catch a row that matches too much. This is the "passes when dubplate diverges" failure, seeded by the
harness's own proof.

**Proposed fold.**
1. Every row's match key includes the source pair (or the session's comparison kind). A row adjudicated on an
   oracle-against-oracle grade can never match a comparison with dubplate on one side.
2. Add a fifth class, `oracle-split` (the two oracles disagree; 2b's per-endpoint assignment says which one
   governs), legal only on an oracle-against-oracle pair.
3. A row's path pattern names one field path, with an array-index wildcard as the only wildcard. The report prints
   each row's match count, so an over-broad row shows on review.
4. Restate the proof's assertion as "exactly the declared set": every finding is ledgered, and every
   `by-construction` row for the folder view fires. The task report lists every row the proof added, with its class
   and match count.

### R2. High. Positional id pairing breaks on the fixture the plan itself builds

`plan:545-546`, `plan:560-563`, `plan:317-322`.

**Defect.** "The grader pairs ids by the position they occupy in corresponding responses" (`:545`). Task 1's
fixture deliberately includes diacritics, non-Latin script, sort tags that differ from display names, a compilation,
and a same-title pair (`:317-322`). Those are exactly the inputs on which two servers sort `getArtists`, index
buckets, and directory children differently. One reordered list then pairs every id in it wrongly, reports id
conflicts wherever the id recurs (`parent`, `albumId`, `artistId`), and cascades into value findings on every paired
element. That noise feeds R1's ledger inflation. The test mandate also contradicts itself. A "reordered
order-significant array" perturbation must yield "exactly one finding" (`:562-563`), but under positional pairing a
reorder of id-bearing elements also yields id-pair conflicts. Finally, which arrays are order-significant has no
named home (code table, ledger, or spec), yet "treat arrays as unordered everywhere" is a mutation (`:564`).

**Proposed fold.** Pair entities by fixture identity, not position. The recorder already resolves symbolic keys
against the manifest (`:441-444`), so the grader maps each response entity to its manifest key (album artist and
title, disc and track) and pairs ids through that map. Order then becomes its own assertion, applied only on a
named list of order-significant paths (song order within an album, index-bucket order), kept as a table in the grade
package with its source cited. The reorder perturbation then yields one order finding and no id finding.

### R3. Medium. The conformance package's tests have no self-contained input, and the boundary test cannot see that

`plan:526-530`, `plan:352-355`, `plan:176-182`.

**Defect.** Task 7's proofs need an OpenAPI document and are "hand-built from the seed session's real `getAlbum`
body" (`:526`). Both live under `internal/harness/subsonic/data/` (`:180`, `:436`). A test under
`internal/opensubsonic/conformance` can reach them only by a relative file read across the subtree
(`../../harness/...`). Task 2's boundary leg lists Go imports (`:352`), so a file read passes it. The result is a
package that the give-back intent publishes later whose tests break on extraction. A 453 KB Apache-2.0 file copied
into its testdata would also need its license beside it.

**Proposed fold.** The conformance package's tests use a small OpenAPI fragment written in-repo for them (one path,
one `200` schema, one path without a schema, one undeclared status) plus a minimal `getAlbum` body. All of it sits in
the package's own `testdata/`, with no oracle output. The full pinned file and the real seed body are exercised from
the harness side (Task 8's gate test). Add to Task 7's acceptance: no file under `internal/opensubsonic/` reads a path
outside that subtree (`git grep -n '\.\./' -- internal/opensubsonic` prints nothing).

### R4. Medium. Task 0's main-tree sweep can collide with 6d's executor, and the merge window is narrower than R9 reads

`plan:207-213`, `plan:286-289`, `plan:579-580`; 6d plan `:412`, `:1536-1539`.

**Defect.** 6d runs in the main tree on master (6d `:412`). R9 lets 2a's Task 0 sweep commit on master in the main
tree "only while no 6d sweep is in flight" (`:210`). If 6d is executing any task, that is a second executor in the
same tree. `CLAUDE.md`'s lane rule forbids it ("Never launch a second executor into the SAME tree"). The merge has
the same shape. `git merge --ff-only` and the full gate on master happen in 6d's tree. 6d's own text allows the
`go.mod` addition "before Task 0 or after 6d's merge" (6d `:1538-1539`). kin-openapi cannot land before, since
`tidy-check` drops an unimported requirement (`:212-213`, correct). So 2a merges only after 6d closes whenever 6d
has launched first, and the lane holds its worktree until then.

**Proposed fold.** R9 reads: Task 0's sweep runs in the main tree only when the lane registry shows no lane 1
executor. Otherwise 2a adopts the most recent sweep on master as its own in one line, and the worktree is cut from
master's tip (`git worktree add` changes no main-tree file). The merge step reads: "after lane 1's close, or before
lane 1 launches", checked against the registry. See the OWNER FORK below for the ordering.

### R5. Medium. The privileged version and digest fields have unnamed absent states and an ambiguous source

`plan:445-446`, `plan:485-488`, `plan:377-378`.

**Defect.** The ledger refuses a row whose version differs from the session's `serverVersion` (`:485-487`). Task 5
takes `serverVersion` "from the oracle's own envelope" (`:446`) but names no report when the envelope omits it or
carries an empty string. In that case an `active` row with an empty version equals an empty session version, so the
privileged check passes vacuously (planning-miss item 2). The digest has a similar gap. The pin file holds a
registry digest (`:377`). The session records "`podman image inspect` on the running images" (`:445`) without saying
which field. For a multi-arch image, the manifest-list digest and the platform image digest differ, so the recorded
digest may never equal the pin, and no check says whether it must (planning-miss item 1: the podman behavior is not
quoted).

**Proposed fold.** Task 5's states table gains: `serverVersion` absent or empty refuses the recording, naming the
oracle. Task 6 refuses an `active` row with an empty version. Task 3 or 5 quotes `podman image inspect`'s `Digest`
and `RepoDigests` semantics from `podman-image-inspect(1)` and records the `RepoDigests` entry that equals the
pinned digest. The recorder refuses when none does, which proves the running image is the pinned one.

### R6. Low. Teardown on cancellation needs a fresh context, and podman leaves anonymous volumes

`plan:383-385`, `plan:398-404`.

**Defect.** "Teardown removes both containers on every path: ... and context cancellation" (`:384-385`). A teardown
issued through `exec.CommandContext` on the already-cancelled context never starts. `Cmd.Start` checks
`c.ctx.Done()` before it forks and returns `c.ctx.Err()` (`$GOROOT/src/os/exec/exec.go:713-718`). The plan's fake-runtime proof would catch this, but only after a fix round, because the plan does
not quote the behavior (planning-miss item 1). Both images also declare `VOLUME`s. A `podman rm -f` without `-v`
leaks one anonymous volume per run, and every later slice and the 2b roster walks run the stack.

**Proposed fold.** Quote the `exec` behavior. Teardown runs under `context.WithoutCancel(ctx)` with its own timeout.
The run uses `--rm`, or teardown uses `podman rm -f -v`, and the fake asserts the volume flag. Containers carry a
per-run `--label`, so a later run can name orphans left by a killed process.

### R7. Low. Port choice can race slice 1's probe oracles

`plan:374-376`, `plan:241-242`; slice 1 plan `:252-254`.

**Defect.** "Each run picks its loopback ports at start." If that means binding a free port, closing it, and then
handing it to podman, slice 1's probe instances (started on their own ports in parallel) can take it in between.

**Proposed fold.** Publish with no host port (`-p 127.0.0.1::4533`) and read the assigned one with `podman port`. The
fake runtime then proves the argv carries no fixed host port.

### R8. Low. A seed without a Vorbis comment makes the fixture library depend on the building binary

`plan:323-327`; `internal/engine/tagio/write.go:122`, `:239-245`.

**Defect.** `tagio.Write` keeps a seed's existing vendor string. When the seed has no `VORBIS_COMMENT` block, it
stamps a fresh one from `debug.ReadBuildInfo`: `dubplate development` in a test binary, or the VCS pseudo-version in
a built `harness`. Two builds in one process stay identical, so Task 1's determinism proof passes. A seed session
recorded through `harness record-oracle` then reports song `size` values that differ from a library a later slice's
test binary builds. That is oracle noise on every song.

**Proposed fold.** The builder refuses a seed with no `VORBIS_COMMENT` block (the `flac` encoder writes one by
default). The states table gains that row, and a mutation proves the refusal.

### R9. Low. Image seeds need a stated generated origin

`plan:315`, `plan:324`.

**Defect.** The source is planned to go public at 0.8. Audio seeds are encoded short FLAC, but the cover and embedded
images have no stated origin.

**Proposed fold.** One sentence: every image seed is generated (for example by `ffmpeg -f lavfi testsrc`), never
downloaded, and its README line records the command.

### R10. Low. Three surfaces are built ahead of their first consumer

`plan:497-498`, `plan:547-548`, `plan:348-349`.

**Defect.** The repo rule is that a hook or export is built by its first consumer. (a) Task 6's "transcription
proof" converts slice 1's Markdown table to rows and back. Its only consumer is 2b's Task 11 (`:605`), so in 2a it is
test-only code. (b) Search's set-plus-top-N comparison has no search endpoint in any session (`:547-548`). Its first
consumer is slice 4 (art and search). (c) The client's envelope exposes `type` and `openSubsonic` (`:349`), and no
2a caller reads them. The recorder reads only `serverVersion`.

**Proposed fold.** Move (a) to 2b Task 11 and (b) to slice 4's grader work, keeping the grader's per-path comparison
hook that slice 4 extends. For (c), either have the recorder store `type` and `openSubsonic` in session metadata,
which is cheap and useful for drift, or leave them out of the envelope until a reader exists.

### R11. Low. Task 9's credential grep is vacuous on the session format

`plan:595`, `plan:416-419`.

**Defect.** Sessions store parameters as structured metadata, not query strings, so `[?&](t|s|p|apiKey)=` cannot
match a leaked parameter. It also omits `u`.

**Proposed fold.** Grep the metadata's parameter keys (`"(u|p|t|s|apiKey)"\s*:` under `data/sessions/`), or drop the
line and rely on Task 4's writer refusal and its mutation.

### R12. Low. kin-openapi has more silent-nil paths than the one quoted

`plan:113-115`, `plan:510-513`; `openapi3filter/validate_response.go` at v0.149.0.

**Defect.** The quotes table cites the empty-content nil (`:96-112`). `ValidateResponse` also returns nil when the
operation has no responses (`responses.Len() == 0`), for an undocumented status unless `IncludeResponseStatus` is
set, when `contentType.Schema == nil`, for 304, 301, 307, and 308, and for `HEAD`. Task 7's "no schema" proof uses an
undeclared status (`:528-529`). Without `IncludeResponseStatus`, that path returns nil, which is indistinguishable
from valid.

**Proposed fold.** Quote these paths in the table. State that the checker classifies route, status, content type,
and schema presence itself before calling `ValidateResponse`, and that it sets `IncludeResponseStatus` and
`MultiError`.

### R13. Low (not verified). The read-only bind mount may need an SELinux label on this workstation

`plan:374-376`, `plan:616-625`.

**Defect.** Bluefin runs SELinux enforcing, and rootless podman bind mounts commonly need `:z` or `:Z` before the
container can read them. If they do, Task 3's argv proof, written before the live run, pins the wrong argv.

**Proposed fold.** Add to "Not verified at planning". Task 3 runs the live start before freezing the argv assertion,
and records whether the label was needed.

## Over-ceremony, ranked by token and clock cost

### O1. The vacuity read duplicates the per-task mutation mandate

`plan:576-578`. Every task already names its mutations, the implementer reports each one failing, and an Opus
`diff-reviewer` reads each diff. A seventh fresh-context Opus read over six packages costs roughly 0.3M to 0.5M and
one fold task. **Fold:** scope it to Task 8's grader alone, the one task the plan specifies only by outcome, or drop
it.

### O2. Ahead-of-consumer builds

R10 (a) and (b) cost implementation, a review read, and architecture-read surface in 2a for no 2a caller. Moving them
saves roughly 0.3M and keeps the grade package smaller at its first architecture read.

## OWNER FORK

**F1. Which code lane goes first: 6d or 2a's merge?** (R4.) 2a cannot merge while 6d runs, and its kin-openapi
cannot land before 6d's Task 0.
- (a) Launch 6d first. 2a executes in parallel in its worktree, holds the merge-ready branch until 6d closes, then
  rebases and merges. Cost: the worktree and lane row stay open, and the rebase crosses all of 6d.
- (b) Launch 2a first and delay 6d's Task 0 until 2a merges. Cost: the spine rung waits on a harness lane.
- (c) Run both, and 6d's Task 0 sweep adopts kin-openapi by importing it early. This is refused by the
  export-ahead-of-a-caller rule.

**Recommendation: (a).** 6d is on the spine (6, 14, 7, 9) and the harness's first grading consumer is slice 3.
2a's branch touches only new paths plus `go.mod`, `cmd/harness`, and one test file, so the later rebase is cheap.
