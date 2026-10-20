# Rung 12 slice 2a plan fold verification (revision 1.1)

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-2-oracle-harness.md`, revision 1.1 at `1d6e40d`. Fold
record: `docs/research/2026-10-06-rung-12-slice-2-plan-fold.md`. I had no part in the fold. Findings are
correctness gaps ranked by consequence. Polish is out of scope.

Checked directly for this read:

- `scripts/check.sh:178-186`, `:265-280`, and `:360-370`, plus `Makefile:134`.
- `internal/engine/tagio/write.go:122-128` and `:239-245`.
- The parallel-build approach `:250-258`, `CLAUDE.md:99-110`, the re-plan `:46-56`, and the spec `:256-330` and
  `:720-770`.
- In kin-openapi v0.149.0: `openapi3filter/validate_response.go:20-130`, `openapi3/response.go:55-80`,
  `openapi3/schema.go:2270-2330`, and `openapi3/errors.go:29-70`.
- The pinned OpenAPI file in the session scratchpad, walked over all 87 paths.
- Probes:
  - `go vet -tags oracle ./cmd/harness/... ./internal/harness/...` on master: rc 0, which confirms the close step
    needs no venv.
  - A live Navidrome container from the pinned digest. `-p 127.0.0.1::4533` gave `127.0.0.1:42043` through
    `podman port`, `--label` filtered it, and `podman rm -f -v` returned the volume count from 10 to 8.
  - `podman image inspect` on both pinned digests: each digest appears in `Digest` and in `RepoDigests`.

Counts: 0 blockers, 3 major, 9 minor. No over-ceremony finding costs more than the risk it removes.

## Question 1: the convergent blocker

It closed at its location. The committed ledger takes only rows that name `dubplate` (`plan:298-304`, `:617-618`),
and a gate test enforces that rule. The oracle pair is graded under `grade/testdata/seed-ledger.json` and pinned as a
golden (`:739-747`). The differ's proof moved to a same-source perturbation table (`:753-758`). The contract tier is
proven on a hand-authored document plus exact real-body violation lists (`:693-705`), and both measured lists match
the probe output in the scratchpad (`kin/out.txt`).

Most proofs can fail. Three gaps remain, and each one touches what slice 3 consumes: V1 (the library digest), V2 (a
contract row hiding a dubplate violation), and V3 (the perturbation baseline cannot be zero).

## Correctness findings

### V1 (major): the library digest is per session, so the mixed-digest refusal is vacuous and slice 3's join is unstated

`plan:533`, `:718-719`, `:760`, `:35`.

**Defect.** Task 4 records one library digest per session (`:533`). The grader takes one session (`:718`) and
"refuses to compare sources recorded over different library digests" (`:719`). Inside one session every source
shares the one digest. No fixture can carry two digests, so the mutation "drop the library-digest check" (`:760`)
fails no test. C5's purpose was slice 3's case (`:35`): an in-gate dubplate response compared with the committed
oracle session over a library the slice 3 test rebuilt. The plan never says how a dubplate source enters a graded
session. It could be a second session handed to the grader, or a source added through Task 4's writer. Either path
needs a per-source digest.

**Narrow fold.**
- Task 4 records the library digest per source.
- Task 4's writer can add a source to an existing session, which is how slice 3 joins.
- Task 8 compares the two sources' digests.
- The mutation's fixture is a hand-built session whose two sources carry different digests.

### V2 (major): a `contract` row is not keyed to the source whose body violates

`plan:603-605`, `:617-623`, `:720`.

**Defect.** The contract tier checks "every JSON response of both sources" (`:720`). A `contract` row suppresses
"the `contract` kind at its path" and matches on the pair (`:617`, `:621-622`). Neither the `Finding` nor the row
names which source's body broke the schema. From slice 3 on, the fold record expects oracle-side contract rows such as
Navidrome's `/album/userRating`. Under the pair (dubplate, Navidrome), that row also suppresses a dubplate body with
the same violation. The spec builds this tier for exactly that case: "a response can match an oracle and still violate
the schema, and only this tier catches that" (spec `:259-261`). This is R1's defect class, which hides a real
divergence, moved from the oracle tier to the contract tier.

**Narrow fold.**
- A `contract` finding carries its one source.
- A `contract` row names the source it excuses, and a row matches only a finding from that source.
- Task 6's pair proof gains one case: an oracle-source `contract` row leaves the same-path dubplate `contract`
  finding unledgered.

### V3 (major): the perturbation baseline cannot be zero if it runs the full grader, and one mutation then has no killer

`plan:753-760`, `:702-704`, `:733-734`.

**Defect.** The baseline is "one real seed-session body … compared against itself", and it must yield "exactly
zero findings" (`:753-755`). The reorder row needs an album's song array, the order table's only entry (`:733-734`),
so the body is a `getAlbum`. Task 7 asserts that both oracles' real `getAlbum` JSON bodies violate the schema
(`:702-704`). The grader runs the contract tier on every JSON response (`:720`), so the baseline carries contract
findings on both sides. If the implementer keeps the baseline at zero by driving the differ alone, then nothing kills
"skip the contract tier when the oracle tier is clean" (`:759`). The seed grade cannot kill it either, because the
oracle pair's oracle tier is never clean.

**Narrow fold.**
- The perturbation table drives the oracle-tier differ directly.
- Add one grader-level case: the real Navidrome `getAlbum` JSON against itself yields zero oracle-tier findings and
  exactly the `contract` findings Task 7 recorded. That case kills the skip mutation.

### V4 (minor): the branch filter's mechanism is stated, not quoted, and `errors.As` alone loses branch membership

`plan:672-676`.

**Defect.** A flattening `errors.As` walk drops which `oneOf` alternative a leaf came from. The mechanics probe ran
that kind of walk (`p2/main.go`'s `leaves`). The source does keep the grouping:
- The `oneOf` `SchemaError` sets `Origin = fmt.Errorf("doesn't match schema due to: %w", validationErrors)`
  (`schema.go:2326`).
- `validationErrors` is a `multiErrorForOneOf` holding one error per failing branch, in `OneOf` order
  (`schema.go:2284`, `:2301-2304`), and it unwraps to `MultiError` (`errors.go:57-58`).
- When no branch matched, entry `i` is `OneOf[i]`.

**Narrow fold.**
- Quote those lines in the quotes table.
- State the rule: descend each `SchemaField == "oneOf"` error's `Origin`, then keep the entry whose `OneOf` schema's
  `status` enum holds the body's `status`.

### V5 (minor): Task 6's path grammar makes the attribute-to-element perturbation a non-finding

`plan:606-607` against `:757-758`.

**Defect.** An XML attribute and a child element map to the same segment name (`:606-607`). Under that rule, the row
"an XML attribute moved to a child element" (`:757-758`) yields an identical tree, which means zero findings where the
row expects one. No kind is named for the difference.

**Narrow fold.** Each tree node keeps its XML form, attribute or element. A form mismatch at one path is one `type`
finding.

### V6 (minor): identity pairing needs the manifest, and the grader is not given one

`plan:718`, `:726-731`, `:416-421`.

**Defect.** Pairing runs "through its manifest identity", and an entity that "matches no manifest identity" is a
finding. The grader's inputs are a session, a ledger, the document, and a pair, with no manifest. The fixture's track
with every optional tag missing (`:420-421`) may lack the disc and track numbers that a song's identity uses.

**Narrow fold.**
- Name the manifest, the committed one or a digest-checked path, as a grader input.
- Give a song's identity a fallback when its track number is absent, such as the title.

### V7 (minor): the privileged-version rule is singular, and the seed grade's test rows name two oracles

`plan:632-635`, `:739-741`.

**Defect.** An `active` row's version must equal "the session's recorded `serverVersion` for that oracle". A
Navidrome-against-gonic test row has two oracle sources and one version column.

**Narrow fold.** A row carries one version per non-dubplate source.

### V8 (minor): the pinned file's shape is overstated

`plan:165`, `:672`.

**Defect.** Measured over the 87 paths:
- 14 operations' `200` responses carry non-JSON content, among them `stream`, `download`, `getCoverArt`, and
  `hls.m3u8`.
- `getTranscodeDecision` (POST) is not a success-or-failure union.

The claim holds for all ten seed endpoints, so 2a is unaffected. Slices 4 and 5 would inherit a false premise.

**Narrow fold.** Scope both sentences to the JSON `subsonic-response` operations, and keep the "no readable `status`"
fallback.

### V9 (minor): the ceiling arithmetic prices seven build tasks, not eight

`plan:65-71`.

**Defect.** The pricing names "six Sonnet build tasks at 0.5M" plus the Opus Task 8. Tasks 1 to 7 are seven Sonnet
tasks (`:81`), so one task, about 0.5M, is unpriced.

**Narrow fold.** Price seven, or state which task is absorbed and why.

### V10 (minor): the 6d pointers drifted with 6d revision 2.5 (`28466d4`)

`plan:134-136`, `:279`, `:327-328`.

**Defect.** 6d is now revision 2.5. "Where it runs" moved to `:432`, the sweep to `:950-957`, and the "before Task 0
or after 6d's merge" sentence to `:1708`. Task 0 re-resolves pointers, so the cost is low.

**Narrow fold.** Re-cite the pointers at 6d revision 2.5.

### V11 (minor): five podman behaviors rely on recall, not a quote

`plan:177-178`, `:479-484`, `:491`, `:497`.

**Defect.** These five behaviors appear in no quote and in no Task 0 quote item:
- the `-p 127.0.0.1::<port>` publish;
- `podman port`;
- `--label` filtering;
- `--tmpfs`;
- `rm -f -v` removing anonymous volumes.

This read proved all but `--tmpfs` live (mechanics M6 proved that one), so the mechanisms hold.

**Narrow fold.** Add `podman-run(1)`'s `-p` and `--tmpfs`, `podman-port(1)`, and `podman-rm(1)`'s `-v` to Task 0's
quote list.

### V12 (minor): the `HEAD` proof contradicts the stated classification order

`plan:666-670`, `:700`.

**Defect.** The checker resolves the route first. The test document declares no `HEAD` operation, so a `HEAD`
request reports "no route" before it reaches "no schema", while the proof expects "no schema". kin itself checks
`HEAD` before it routes (`validate_response.go:21-23`).

**Narrow fold.** The checker tests for `HEAD` and the four 3xx statuses before it routes.

## Question 2: the other High and major findings

| ID | Closed at its location? |
|---|---|
| C2, R2 | Yes (`:533-537`, `:568-572`, `:726-731`). V6 is a residue. |
| C3 | Yes. The pre-resolution matches kin's own order (`response.go:58-80`, `validate_response.go:44-57`). V12 is a residue. |
| C4, M4 | Yes. The two-sibling body matches the probe table. V4 is a residue. |
| C5, R8 | Partly. The seed vendor rule is verified at `write.go:122-128`. The digest half is V1. |
| C6, C7 | Yes. V5 and V7 are residues. |
| M1 | Yes. All four quotes are verbatim at HEAD. |

## Question 3: section contradictions and task order

Numbering holds: "Tasks 0 to 9" makes ten, segments A (1 to 3), B (4 to 6), and C (7 and 8), and 2b is Tasks 10 to
13. Task 9's six new packages and seven architecture reads also match. The build order is sound:

- Task 3 needs Tasks 1 and 2.
- Task 5 needs Tasks 1, 3, and 4.
- Task 7's loader lands in the `grade` package that Task 6 creates.
- Task 8 needs Tasks 5, 6, and 7.

The contradictions are V3, V5, and V12.

## Question 4: mechanisms stated from memory

These were proven or quoted for this read:
- the kin silent-nil paths;
- `Responses.Status` and `Default`;
- `tagio`'s vendor handling;
- the oracle-tag vet;
- the podman publish, port, label, and volume-removal behavior;
- the image digest fields.

Two remain unquoted: V4 (the branch grouping) and V11 (the podman man pages).

## Question 5: over-ceremony

The fold grew the plan from 625 to 826 lines. Most of the growth is quotes and states that close measured defects.
None of it costs more than the risk it removes.

One optional cost note: the full-report golden of oracle-pair findings will churn on every grader change through
slice 8. Regenerating it is cheap, and the named assertions carry the correctness load.

## Question 6: the worktree citation

Accurate. `CLAUDE.md:103-104` states the rule, and the parallel-build approach `:255` (item 5) gives the venv reason
quoted in R1. Re-plan ruling 5 (`:48-54`) places slice 2 in its own worktree. The `go vet -tags oracle` residue is
answered and was proven at rc 0.

**Verdict: needs second fold.** The fold is narrow. Each of V1 to V3 is a few sentences in Tasks 4, 6, and 8, and the
minors ride along.
