# Rungs 12 and 16 re-plan: fold record

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 1 at `42c06a1`, revised in place
to revision 2. Reviews folded: `2026-10-05-rungs-12-16-replan-review-{contract,mechanics,data,consistency}.md` (ids
C, M, D, K; over-ceremony ids are prefixed by lens: cO, mO, dO).

Method: each finding was checked against the tree before acting. Spot checks run for this fold:

- `roster.go:340` decodes with `json.RejectUnknownMembers(true)`.
- `removeCookie` and `withoutCookie` read `Strip.sessionCookie` only (`serve.go:298,320,360`).
- The only production `NewStrip` call passes `""`, `auth.SessionCookieName`, and an empty `FrontExemption`.
- `SetFallthrough` and `addBearerBypass` have no production caller.
- The configuration homes spec's bearer admin mount (`:123-126`) and its private gate on a capability report
  (`:221-224`) exist as quoted.
- The build spec's ruling 1 adopter clause (`:976-988`) and the "first-class leaf" sentence (`:459-461`) exist as
  quoted.
- `internal/engine/tagio`'s only importer is `internal/engine/shadow`.
- The serve unit has empty capability lines and two `ReadWritePaths`, and `scripts/firewall-rules.json` opens tcp/22
  alone.
- The env template carries `DUBPLATE_PUBLIC_URL=https://music.907.life`.
- The 6d plan is at revision 2.2, unruled, and its `member import` deploy step reads the roster.
- The auth secrets are required environment rows (`cmd/dubplate-server/auth.go:48-49`).

The owner's standing bars were applied as the deciding test throughout: (a) architecture, (b) install ease, and
(c) the seven rulings stay closed.

## Convergent defects, fixed at the root

| Root | Ids | Disposition |
|---|---|---|
| The roster row has no sanctioned migration path | C4, M3, D2, K12; mO1, dO1 | Folded. "The username rename" states why no verb can read the old row. The first roster-reading deploy (6d's, or rung 16) carries a named runbook step: move aside, `member add --admin`, doctor check. The removal pass does not deploy ("The removal pass" intro). |
| The archive is one unverified copy, and the import's order is unstated | D1, K13 | Folded. New section "Preserving the archive (now)". Rung 16 step 4 reads the verified copy and runs before the re-layout. Copy location decided (below), not a fork. |
| The bearer bypass is misfiled as Navidrome-facing | M2, D7 | Folded as M2's option (a). The bypass stays as generic hardening, with its consumers named (rung 14's bearer admin mount; API non-GETs from slice 5). Its two contract sentences stay in code, and streamer rev 3's slice 6 records them. The `netip.Addr` sentence is deleted, since the bypass is path-keyed. M14 corrected the count to "three pieces". |
| The rename list is incomplete | C7, M10, D6, K16 (and M6's roster example) | Folded. A grep acceptance line, the wire codes, the flag, the column head, `run.ts`, the example roster, and the runbook are all named. |
| Shadow residue is everywhere | C8, M5, K15, D8 (Makefile part) | Folded. Enumerated orphan dispositions, a `git grep` line, and a `go list -deps` line, all under "The shadow program", conditional on Owner ruling 1. |
| Front, driver, and fallthrough prose survives outside `internal/serve` | C9, M13 | Folded. The prose criterion is extended to `cmd/dubplate-server/serve*.go`, `internal/auth`, and `internal/config`. Config row dispositions are named. The test rename is named. |
| The in-place `0003` edit trips the hash pin and the append-only rule | M9, D5, K11 | Folded. Re-pin in the same commit from an independent `sha256sum`. The precondition is doctor's `store` row reading `absent`, plus no remote snapshot. A rung 6 spec erratum is listed in supersedes and in doc change 5. |
| The root URL and the owned-prefix rule against `/rest` | M15, K5 | Folded as a routing decision, not an owner fork. Slice 1 measures clients against a sub-path base URL, and slice 3 decides (table row). The `;`, case, and percent-encoding carries go to slice 3. The `servePrefixThreshold` comment is rewritten. Client compatibility is an evidence question, and the slice 3 plan returns it to Geoff's approval gate. |
| Rung 16 rollback and backup | C10 (rollback part), D3 | Folded. Rung 16 order, steps 1 to 3, and the two-form rollback. Table row for backup provisioning. |
| Import member mapping and reruns | C10 (import part), D4 | Folded. The "case-insensitive matching is moot" row is replaced by a slice 8 row. Streamer rev 3 adds four slice 8 criteria (dry run, address-then-NOCASE match with refusal on disagreement, transactional or idempotent, extracted copy). |
| Two web reviewers on a rename | cO1, mO2 | mO2 folded: one shared brief scoped to the rename diff (Close). cO1 refused: `CLAUDE.md` fans `web/` out to both reviewers by surface. |

## Remaining dispositions

### Contract lens

- **C1, folded.** The rung 12 "streamer's configuration" bullets: always on, tunables as defaulted thresholds,
  secrets on the auth-secret pattern generated by `init`, and a per-slice delta to the operator-supplied set. Doc
  change 2 rewrites "Cutover and operations". The reviewer's "minted at first start" form was refused as new
  machinery against the repo's own secret pattern; the conventional fix was taken.
- **C2, folded.** Table rows: firewall and DNS to rung 16; `CAP_NET_BIND_SERVICE` and the certificate leaf to slice
  8, also in doc change 2.
- **C3, owner fork.** Became Owner ruling 3. Its "defects have no disposition" half is answered in both of the
  ruling's branches.
- **C5, folded in part.** The checkable criteria are in: the `go doc` grep, the exported-unused leg, the
  architecture-reader verdict, and the named doctor tests. Two parts are refused:
  - A 404 test "failing against today's wiring" is refused, because production never set a fallthrough and the
    outward behavior is unchanged (verified against mechanics' "What holds"). `TestEmptyFallthroughAnswers404` is
    re-scoped instead.
  - The bearer-refusal test is refused because the bypass stays.
- **C6, folded in part.** Ruling 7's second bullet is corrected as a factual error; the ruling itself is unchanged.
  The count and the per-slice delta are added. The golden test over required names is refused, because the
  per-slice delta line and `keys.go` diffs already make growth visible. The identity-source default joins the
  cleanup inventory under Owner ruling 2's yes branch.
- **C11, folded.** Table row: command naming to slice 8.
- **C12, folded.** Rung 12 paragraph: no public port before slice 8, and attended checks on the workstation or through
  `ssh -L`.
- **cO2, refused.** The document set grew to nine items (K3, K4, K6, K7, K21). Document lanes share the tree under
  the standing grant, so the lane costs one registry entry.

### Mechanics lens

- **M1, folded.** The full cookie set is named, and `NewStrip(proxy TrustedProxyConfig) Strip` is stated.
- **M4, folded.** Acceptance: re-point probes at a spy subtree or a non-404 status, delete only single-subject
  tests, and list a disposition per test.
- **M6, folded.** Doc change 8 covers the runbook, and the rename covers the example roster.
- **M7, folded.** `web-auth-security-reviewer` is added to Close.
- **M8, folded.** `Prefix.Route` stays with `http.NotFoundHandler()`, and later root routes are named routes.
- **M11, folded.** `statusCapture`'s methods stay, with the rationale restated.
- **M12, folded.** The exemption list is corrected, and `collapseToLast` is deleted outright.
- **M14, folded.** "Three for good", and the bypass is kept.
- **mO3, folded.** 6d moves to revision 2.3 after the removal pass ("The removal pass" intro, with K10).

### Data lens

- **D8, folded.** Rung 17's disposal covers every archive copy and the shadow directories.
- **dO2, folded.** The roster stays at `schema_version` 1, with no version gate (rename section).

### Consistency lens

- **K1, folded.** "Every build-spec rung 16 item not marked moot still stands", plus rows for backup keys, the
  single-link migration, relocation (now unconditional), store identity, and the beets database (moot). Moot rows
  are added for ExtAuth ordering, `UIWelcomeMessage`, and compose.
- **K2, owner fork.** Became Owner ruling 1. It reverses a ratified owner ruling, so the spec cannot take it.
- **K3, folded.** Rung 14's private gate is re-keyed to slice 3's projector, and private ships at first launch,
  per the existing 2.0 ruling. The configuration homes spec is added to supersedes and to doc change 3, and the
  ladder line is corrected.
- **K4, folded.** Doc change 1 adds a header-level erratum listing the sections.
- **K6, folded.** Doc change 4 adds a principles-spec erratum.
- **K7, folded.** Doc change 2 covers "Relation to the current ladder", the header, cutover, the arm64 walk,
  hook 19 to slice 8, hook 15's clause, and the study-timing open item.
- **K8, folded.** A ladder paragraph lists the hooks per rung, and doc change 6 rewords `CLAUDE.md`.
- **K9, folded.** Table rows: re-validation and m1 go to 6d and then slice 6; offboarding goes to slice 6; R33 and
  the header-set reconcile are moot.
- **K10, folded.** 6d's approval gate is on revision 2.3, and the four contents are listed. This is a conductor
  sequencing call, not an owner fork.
- **K14, owner fork.** Became Owner ruling 2. A ROADMAP entry leaving for `IDEAS.md` is scope, which ruling 4 does
  not cover.
- **K17, folded.** The moot reason is reworded.
- **K18, folded in part.** UNI-H2 is re-routed to slice 3 with the root-route set, because TLS was never its reason.
  Landing it in the removal pass is refused as accretion by adjacency in a removal-only pass.
- **K19, folded conditionally.** The deck records the departure if Microsoft's A-Z spells "user name". This fold
  did not verify the guide entry.
- **K20, folded.** The roster walk is conductor-run on the workstation. `CLAUDE.md`'s touchpoint list says "only
  these", so the default is unattended, and an undrivable client falls to rung 16's attended gate.
- **K21, folded.** Doc change 7 covers ROADMAP's greenfield and `packaging` text. Doc change 9 has the conductor
  point the two memory files.

## Candidate forks judged

| Candidate | Verdict | Reason |
|---|---|---|
| K2, shadow deletion | Owner ruling 1 | Reverses a ratified owner ruling's adopter clause. |
| K14, forward-auth front | Owner ruling 2 | Product scope beyond ruling 4's text. |
| C3, fresh-host walk | Owner ruling 3 | Changes the place ruling 7 names, and spends a throwaway VPS. |
| D1, archive copy location | Spec decision | The workstation age store is the only encrypted store before rung 16, and rung 16's remote adds the second copy. The 3-2-1 practice settles "off-box, encrypted". |
| M15, `/rest` at root | Spec decision (routed) | Client compatibility with a sub-path base URL is measurable at slice 1, and slice 3's plan brings the result to Geoff's gate. |

## Errata owed

The revision changes the meaning of these ratified documents. None was edited by this fold, and the spec's
"Document changes" lists each one:

1. Build spec: a header-level erratum, errata at rungs 12, 15, 16, and 17 and at launch bar part 3, and the shadow
   section's retirement. Ruling 1's adopter clause and the "first-class leaf" sentence are added only if Owner
   ruling 1 is yes.
2. Streamer spec revision 3: the expanded list in doc change 2.
3. Configuration homes spec: the rung 12 verifications and the shadow-pair erratum become moot, and the private
   gate is re-keyed.
4. Group-scale principles spec: "Before 0.8", A13, the delegated route row, and versions.
5. Rung 6 spec: the one-time append-only exception for `0003`.
6. `CLAUDE.md`: versions, build-for-0.8, and the touchpoint rewording.
7. Not ratified, but owed: the 6d plan's revision 2.3, `docs/server-runbook.md`, and the two memory files.

## Measures

**Disposition counts:** 63 finding ids.

| Disposition | Count | Ids |
|---|---|---|
| Folded, fully or in part | 58 | Partial folds are C5, C6, K18, and K19 (conditional). |
| Refused outright | 2 | cO1, cO2 |
| Owner fork | 3 | K2, K14, C3 |
| Partial refusals inside folds | 4 parts | C1's first-start minting, C5's two tests, C6's golden test, K18's in-pass landing |

**New-mechanism findings:**

- Folded:
  - D1, the off-box encrypted copy. Source: the 3-2-1 backup practice (US-CERT, "Data Backup Options").
    Measured defect: one copy, with no sha256, on a host rung 16 may reprovision.
  - D4, the import's dry run and rerun safety. Source: the repo's own 6d `member import --dry-run`. Measured
    defect: a rerun after a partial run doubles play counts and playlists.
  - D3, the backup and restore proof ordered before irreversible steps. Source: the runbook's existing backup verb.
    Measured defect: backup is off on the box, and `.prev` is gone.
- Refused:
  - C1's secret minted at first start. The repo's auth-secret pattern plus `init` was taken.
  - C6's required-names golden test.
- Owner forks: C3's fresh-host walk (ruling 3).

**Spec line count:** 223 before (`42c06a1`), 443 after. The growth is tables and enumerated acceptance
replacing adjectives, plus the three owner rulings, the archive chore, and rung 16's ordered steps.

**Ceiling:** the spec carries no token ceiling, and none changed. The removal pass plan sets its own.

## Second fold

Input: `2026-10-05-rungs-12-16-replan-fold-verification.md` (ids V). The spec moves from revision 2 to revision
2.1 in place. Scope was the verification findings only, each checked against the tree before editing. Owner rulings
were not touched.

| Id | Disposition |
|---|---|
| V1 | Folded. Tree: `cmd/dubplate-server/doctor.go:108-115` prints the `roster` row only on failure; `authStoreDoctorRow` reports the auth leaf. The roster step's check is now "`member list` shows the one admin, and doctor prints no `roster` row". |
| V2 | Folded. The seven numbered steps are replaced by six ordering constraints (`.prev` before the first `--apply`; backup and a proven restore before the import and re-layout; a store snapshot immediately before the import; import before re-layout; the no-move check after re-layout; the attended gate last) and a list of items the rung 16 plan must include (the 0.8 deploy, the roster step, `init` and TLS, firewall and DNS, enabling `serve`, the import, the re-layout). The sequence is the plan's. The carried-items row "Rung 16, step 1" now reads "Rung 16". |
| V3 | Folded, re-keyed by an owner grant. Tree: streamer spec `:660-664` gives slice 2 the protocol client, checker, differ, and `kin-openapi` in `go.mod`, so slice 2 is a code lane. Grant (Geoff, 2026-10-05, relayed by the conductor mid-fold): three code lanes when useful and efficient. Ruling 5 and the ladder text now say slice 1 opens now as a document lane and slice 2 opens now as a third code lane in its own git worktree. `CLAUDE.md`'s "harness work never moves to a worktree" covers the matcher's corpus and venv harness, not slice 2's; if the slice 2 plan finds its harness needs the main tree, it says so and takes the next free slot. The `go.mod` collision is handled by order: `kin-openapi` rebases onto the removal pass's dependency sweep, under the merge ritual. `CLAUDE.md`'s "up to two code lanes" line is owed an update in the document pass (added to doc change 6). |
| V4 | Folded. Tree: `go doc -all ./internal/serve` carries generic "in front of" prose (`doc.go`, HSTS) and a "same-host front" in the trusted-proxy doc; `rendered_test.go:135-140` carries `component/surface/shadow/...` ids. Both greps now name the removed identifiers (`front ?exemption`, `forward-auth`, `driver`, `fallthrough`, `exemption`; `engine/shadow`, `DUBPLATE_SHADOW`, `registerShadow`, `newShadowCmd`, `shadow (observe|program|runner|store|integration|class|leaf)`) and state which legitimate hits stay. |
| V5 | Folded. `tagio` stays intact with rung 9 as its consumer; the `Tags` unexport is dropped, and HISTORY records that the staged exported-unused leg warns on `tagio` until rung 9. |
| V6 | Folded. Doc change 1 now supersedes the estate shadow unconditionally and the leaf sentence, ruling 1's adopter clause, and the oracle section's adopter text (with its erratum block) only if Owner ruling 1 is yes. |
| V7 | Folded. The ladder line reads "unchanged in scope", with 6d's move to revision 2.3 and rung 9's inherited consumers named. |
| V8 | Folded by assigning one writer per item: the document pass marks only the moot or answered items (shadow length, Navidrome retention, the study rung beside the ladder); slice 1 answers every item owed to the study rung. |
| V9 | Refused. Tree: `internal/engine/store/testdata/fixtures/0003_catalog.sql:5` inserts into `users (..., streamer_username, ...)`, as do the `0004` and `0005` fixtures. The spec's "`0003` to `0005`" is correct. |
| Q4.2 | Folded. The `go-architecture-reader` verdict bullet is dropped (the reader still runs at close per `CLAUDE.md`), and the doctor web-row and probe-route rename bullet is dropped; neither answers a measured defect. |
| Q4.4 | Kept. Task 0's count of the operator-supplied set is a cheap baseline the slice plans measure against. |

**New mechanisms:** none.

**Spec line count:** 443 before, 457 after. V2's constraints plus the must-include list run longer than the seven
steps, and V3's grant adds lines to ruling 5 and doc change 6.
