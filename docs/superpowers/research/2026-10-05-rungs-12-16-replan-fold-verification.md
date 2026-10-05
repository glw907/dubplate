# Rungs 12 and 16 re-plan: fold verification

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2 at `33708a9`. Fold record:
`2026-10-05-rungs-12-16-replan-fold.md`. Read with fresh context against the tree, not against the fold's word.
Findings are limited to gaps that affect correctness, buildability, or the owner bars (clean architecture; install
ease), ranked by consequence.

**Counts:** 0 blocker, 3 major, 6 minor.

## Q1. Did the high and major findings close at the cited location?

Checked against the tree:

| Finding | Closed at | Tree evidence |
|---|---|---|
| D1, K13 (archive) | spec:172-186, :154-156 | Chore is read-only on the box; order against re-layout stated. |
| D2, C4, M3, K12 (roster row) | spec:315-323 | `roster.go:340` strict decode; `member.go:117-131` takes `CreateRoster` only on `ErrRosterMissing`; `--admin` exists (`member.go:102`). Closed, except the step 3 check (V1). |
| C1 (streamer config) | spec:124-131 | Holds. |
| C2 (TLS ingress) | table, spec:409-410 | Holds. |
| C3 (stranger walk) | Owner ruling 3 | Holds. |
| C5, M4 (vacuous criteria, probe tests) | spec:228-246 | `TestEmptyFallthroughAnswers404` (`internal/serve/serve_test.go:261`) and `TestNewServeShell_StripCarriesNoFrontExemption` (`cmd/dubplate-server/serve_test.go:1896`) exist. The `front` grep has false positives (V4). |
| M1 (cookie set) | spec:210-217 | `removeCookie`/`withoutCookie` read only `Strip.sessionCookie` (`serve.go:298,320,360`); sole production `NewStrip` call is `serve.go:324`. Holds. |
| M2 (bearer bypass) | spec:219-223 | `addBearerBypass` (`origin.go:88`) has no production caller; keeping it is consistent with the config homes spec. Holds. |
| M5, K15 (shadow orphans) | spec:250-273 | `cmd/dubplate-server` imports `internal/harness/{analysis,cassette}` only from `shadow.go`/`shadow_test.go`, so the `go list -deps` line is achievable. `tagio`'s only importer is `internal/engine/shadow`. The `git grep` line has false positives (V4); the `tagio` disposition is incoherent (V5). |
| M6, M7 | spec:293-295, :331, doc change 8 | Holds. |
| K1 (dropped rung 16 items) | spec:147, table | Holds; the ordered list it added is defective (V2). |
| K2, K14 | Owner rulings 1, 2 | Build spec `:459-461` and `:976-988` quote correctly. |
| K3 (private gate) | spec:139-141 | Holds. |
| K11, M9, D5 (`0003` hash) | spec:307-313 | `migrationHashes` (`internal/engine/store/schema_test.go:294`), enforced by `TestSchemaMigrationHashPins`; its doc already requires "an independent sha256sum run". Ledger check at `migrate.go:238` fails loudly. Doctor's `store` row prints `absent` for `StateAbsent` (`store.go:158-159`). Holds. |

## Findings

### V1 (major): the roster step's doctor check names the wrong row

**Where:** spec:323 ("Check that doctor's `auth store` row reads ok").

**Defect.** The `auth store` row reports the auth leaf, not the roster: `authStoreRowStatus` checks the leaf's mode,
owner, and the `nonce` and `epoch` record directories (`cmd/dubplate-server/doctor.go:210`, `:248-260`). The roster
has its own `roster` row, which prints only when it fails (`doctor.go:108-115`; runbook:229-231). A roster that is
still unparseable can sit beside an `auth store` row that reads ok, and on a box where serve never ran the row reads
`absent; serve creates it`, not `ok`. The claim was stated from memory: data review D2's fold text says "doctor's
`auth store` row reports it", and the fold copied it.

**Fold.** Step 3 becomes: "`member list` shows the one admin, and doctor prints no `roster` row." Or keep only
`member list`, which already proves the file loads.

### V2 (major): rung 16's numbered order omits the deploy and cannot run as written

**Where:** spec:146-163.

**Defect.** The seven steps leave out the 0.8 binary's deploy, the roster step (spec:318), `init`, the TLS listener,
firewall and DNS, and enabling serve. Step 1 provisions the backup secrets, which reach the box only through a
`deploy.sh --apply` render of `server.env`. The runbook's `.prev` set copies the binary, env file, and units
*before* a deploy (runbook:790-793). Taking it at step 2, after step 1's render, snapshots the new state, not the
rollback target. A rung 16 plan written against this list either reorders it or stalls.

**Fold.** Replace the numbered list with the ordering constraints that carry the safety argument, and leave the
sequence to the rung 16 plan:

- the `.prev` set precedes the first `--apply`;
- a backup and a proven restore precede the import and the re-layout;
- a store snapshot immediately precedes the import;
- the import precedes the re-layout (the audio-key fallback reason stays);
- the no-move assertion follows the re-layout;
- the attended gate is last.

This is also the largest piece of over-engineering the fold added (Q4), and the fold shortens the spec.

### V3 (major): slice 2 is a code lane, so the three concurrent lanes break the lane rules

**Where:** spec:47-49 (ruling 5), :119.

**Defect.** The spec calls slices 1 and 2 a "document-and-corpus lane". The streamer spec's own slice 2
(`2026-09-07-streamer-2-0-design.md:668-672`) builds Go code: the protocol module's client and conformance checker,
the differ, and `getkin/kin-openapi` as a module dependency, which is a `go.mod` change. That makes slice 2 a code
lane. Running it beside lane 1 (removal pass, then 6d) and rung 7 gives three code lanes against the standing grant
of two (`CLAUDE.md`, "Standing grants"). `CLAUDE.md` also says corpus and harness work never moves to a worktree, so
slice 2 would share lane 1's tree, which "Never launch a second executor into the SAME tree" forbids. Slice 2's
`go.mod` edit also collides with the removal pass's opening dependency sweep. The fold inherited this framing from
the streamer spec; it did not introduce it.

**Fold.** Slice 1 (study) opens now as a document lane. Slice 2 takes the next free code-lane slot, in the primary
tree, sequenced after the lane holding that tree reaches a merge point. This is a conductor sequencing call inside
ruling 5's intent; one line to Geoff in the approval note suffices.

### V4 (minor): two acceptance greps fail on legitimate prose

**Where:** spec:230, :271.

**Defect.** `grep -inE 'front|...'` over `go doc -all ./internal/serve` matches generic "in front of" text that
survives the removal: the package doc's "sit in front of the shell's handlers" and the HSTS doc's "the proxy or
tunnel in front of it". The trusted-proxy hardening the spec keeps is described exactly that way (spec:433-434).
The `shadow` grep also matches the rendered gate's theme ids `component/surface/shadow/...`
(`cmd/dubplate-server/rendered_test.go:135-140`) and `web.go:72` ("It cannot shadow a real"), neither of which is
CSS `box-shadow` or "silently shadow". Either criterion then forces contorted prose or fails.

**Fold.** Grep for the removed names (`FrontExemption`, `forward-auth`, `fallthrough`, `driver`, `exemption`;
`engine/shadow`, `DUBPLATE_SHADOW`, `shadow (program|runner|observe|store|class|leaf)`), or list the allowed hits.

### V5 (minor): the `tagio` disposition unexports the one function rung 9 is named to use

**Where:** spec:256-257.

**Defect.** After the shadow deletion `tagio` has no importer at all, so its whole exported surface is callerless,
not only `Tags`. The package doc names `Tags` as the single-parse entry point that rung 9 "may fold the two calls
into". Unexporting or removing only `Tags` is churn that rung 9 reverses, and it leaves `Read`, `Write`, and `Parse`
in the same state.

**Fold.** Keep `tagio` intact with rung 9 recorded as its consumer, and drop the `Tags` clause. Record in HISTORY
that the exported-unused leg will warn on `tagio` until rung 9.

### V6 (minor): doc change 1 retires shadow passages unconditionally

**Where:** spec:343-349 against spec:9-12 and Owner ruling 1's "No" branch (spec:77-79).

**Defect.** Doc change 1's header erratum covers "every passage on ... the shadow", and it lists "the tree shape's
... shadow leaf" and "Shadow: the standing oracle" with no condition. Under a "no" on Owner ruling 1, the shadow
stays as an adopter tool, and those passages partly stand.

**Fold.** Split doc change 1. The estate shadow (the floor, launch bar part 3, rung 4 "continuous to launch", the
box side) is unconditional. The leaf sentence, the adopter clause, and the oracle section's adopter text follow
Owner ruling 1.

### V7 (minor): "unchanged" rungs that this spec changes

**Where:** spec:102.

**Defect.** "Rungs 6d, 7, 9 ... are unchanged" contradicts 6d's move to revision 2.3 (spec:194-201) and the
consumers rung 9 inherits: the transliterator rebuild, `tagio`, `ClassBackground`, and the `readFLAC` salvage
(spec:254-264).

**Fold.** "unchanged in scope; 6d moves to revision 2.3, and rung 9's plan inherits the consumers named under 'The
shadow program'."

### V8 (minor): the document pass and the slice 1 lane both write the streamer spec's open items

**Where:** spec:372-374 (doc change 2) against slice 1's charter (streamer spec:653-657, "the answers to the items
'Open items' marks as owed to the study rung").

**Defect.** Two document lanes write the same section of one file with no order, which the lane checklist's
"contested surfaces written" step exists to catch.

**Fold.** Doc change 2 lands before slice 1 opens.

### V9 (minor): the `0003` fixture range is off by one

**Where:** spec:307-308 ("its fixtures (`0003` to `0005`)").

**Defect.** `testdata/fixtures/0003_catalog.sql` carries no `streamer_username`; only the `0004` and `0005` fixtures
do. The rename's `git grep` line catches the real sites, so this is only a wording fix.

**Fold.** "the `0004` and `0005` fixtures".

## Q2. Contradictions and buildability of the pass order

- The removal pass, then 6d revision 2.3, builds. The removal pass touches no box. The roster step correctly sits on
  the first deploy that reads the roster, and `deploy.sh` runs no doctor or check that would trip on the old-key file
  before the step runs.
- Slices 1 and 2: V3 (lane count, same tree, `go.mod`) and V8 (shared file).
- Rung 16: V2.
- Section contradictions: V6, V7.

## Q3. New mechanisms stated from memory

- **`member add --admin` with no roster file:** proven. The flag exists (`member.go:102`). `addMember` bootstraps
  through `CreateRoster` on `ErrRosterMissing` (`member.go:117-131`). The verb requires `--name` and today's
  `--streamer-username`, which the rename makes `--username`. `checkRosterDirOwner` requires the effective uid to own
  the directory, so `systemd-run` as `music` is required, as the spec says and the runbook already does
  (runbook:145-164).
- **Doctor check in the roster step:** wrong row (V1).
- **Hash re-pin:** correct. The variable is `migrationHashes`, the test is `TestSchemaMigrationHashPins`, and the
  "independent sha256sum" is the variable's own documented rule.
- **Precondition, store row `absent`:** correct (`store.go:158-159`); doctor writes nothing.
- **`go list -deps` line:** achievable; verified.
- **`git grep` and `go doc` greps:** false positives (V4).
- **Exported-unused leg:** exists, staged (`scripts/check.sh:78`, `:618-621`).
- **Archive chore:** steps 1-6 are plain tool use (`tar -t`, `sha256sum`, `scp`, `age`, `sqlite3 PRAGMA
  integrity_check` on an extracted copy). Nothing in them depends on repo behavior, and it is read-only on the box.

## Q4. Over-engineering the fold added, ranked by cost

1. **Rung 16's numbered seven-step order** (spec:149-160). It is a plan's job, it is incomplete, and it is
   misordered (V2). Replacing it with ordering constraints shortens the spec.
2. **Two `internal/serve` acceptance bullets with no measured defect** (spec:235-236, :246). "The architecture
   reader's verdict carries no unused-surface finding" turns a close-time reviewer into an acceptance criterion
   that Close (spec:330) already dispatches. "The doctor web-row and probe-route tests are named in the plan, and
   the diff touches them only to rename" guards against nothing a reviewer reported. Drop both. Keep the re-point
   and deletion rules: M4 measured the vacuity risk they close.
3. **The `Tags` unexport** (V5).
4. **Removal pass Task 0 counts the operator-supplied set** (spec:130-131). It is a cheap baseline for a delta the
   removal pass cannot move; optional, and low cost to keep.

Not over-engineering: the archive chore (a measured single-copy risk to members' data), the roster runbook step (no
verb path exists without it), the hash re-pin (the test forces it), and the carried-items table (routing that K1
and K9 found dropped).

## Q5. Can the owner rule the three rulings on this evidence?

- **Owner ruling 1:** yes. The citations hold, and both branches name their consequences. V6 should be folded so the
  "no" branch's document effect is stated.
- **Owner ruling 2:** yes. The three places the promise lives are named, and each branch states its edits.
- **Owner ruling 3:** yes. The cost (one throwaway VPS) and the box's age-store conflict are stated. It is a genuine
  fork, because it moves where Geoff's ruling 7 runs and spends money.

Forks judged: the fold decided nothing that belongs to Geoff. The in-place `0003` edit relaxes a spec design rule
(rung 6 spec:424), not an owner ruling, and its preconditions make it an architecture call. The archive location
and the `/rest` routing are correctly spec decisions. V3 is a sequencing correction inside ruling 5's intent. It
needs one line in the approval note, not a fourth ruling.
