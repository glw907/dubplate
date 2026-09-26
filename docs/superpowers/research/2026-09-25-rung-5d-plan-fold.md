# Rung 5d plan fold: revision 1 to revision 2

Folded: `docs/superpowers/plans/2026-09-25-rung-5d-boundary-deployment-close.md`, revision 1
(`af7d941`) to revision 2, against the three lens reviews committed at `70ce7be`:
`2026-09-25-rung-5d-plan-review-contract.md` (C: 0 blockers, 7 majors, 16 minors),
`-mechanics.md` (K: 0 blockers, 6 majors, 12 minors), and `-risk.md` (R: 1 blocker, 4 majors, 12
minors). 58 findings in all.

Every finding was checked against the tree at `70ce7be` or the box before it was folded. The box was
read once, read-only (`ssh -o BatchMode=yes`), on 2026-09-25. Sources of authority: the rung 5 spec, the
rung 6 spec's 6d section, the repo CLAUDE.md, and the global secrets rule.

## Counts

| Disposition | Count |
| --- | --- |
| Folded | 58 (3 with a named part refused, below) |
| Refused outright | 0 |
| Owner fork | 0 |

Duplicates fold once, under the first ID in each row below; the other IDs point to it.

**Why nothing was refused outright.** Anthropic's guidance warns that a reviewer asked to find gaps
usually finds some. Each finding here was checked against the tree or the box, and each major held.
The blocker (R-B1) and the three convergent majors (the whole-estate apply, "local only", and the
header counting) rest on facts the fold confirmed: the shadow timer's `*:0/5`, the apply's
binary-before-env order, the `methodOnly` 405, and the canonicalizing parser. Most minors cost a
sentence, a table row, or a mutation. Where a proposed fix cost more than the risk it removed, that
part alone was refused and the cheaper half was folded:

1. **C-M3's "retire the shadow line into the loop"**: refused. `install -d -m 0700` re-modes an
   existing directory (measured below), and the live shadow leaf is 0755 on the box. The loop skips
   `shadow` by name instead, which is K-M4's proposal.
2. **C-M7 and R-m12's "drive the real `render_template` from the Go test"**: refused. `deploy.sh`
   runs top-level code (it sources secrets and resolves the host), so it would need restructuring to
   be sourceable. The parity rows and Task 10's real `--check` cover the drift at far less cost.
3. **R-B1's "stop `dubplate-import.timer` across the apply"**: refused on verified mechanics. The
   apply runs `systemctl enable --now $TIMER_UNITS` (`deploy.sh:526-527`), which restarts the import
   timer mid-apply. That timer is `Persistent=true`, and `man systemd.timer` says so: "the service unit
   is triggered immediately if it would have been triggered at least once during the time when the
   timer was inactive". A stop across a quarter-hour would fire an import right beside the `uv tool
   install --force beets` step, which is the opposite of the intent. The apply is timed against the
   import's calendar instead.

## New facts the fold established

- **The box's beets is 2.13.1**, and `deploy.sh` pins 2.14.1 (`:32`, adopted at `367564c`). Any apply
  upgrades the live importer. The fold treats the upgrade as the intended state: the matcher was
  re-derived against the 2.14.1 oracle, and STATUS records beets 2.14.1 as adopted. The fold names the
  upgrade in Task 10's expected set and to Geoff as a shadow-evidence event. It is not a ruling,
  because deploying an adopted pin has one correct answer.
- `config/contributors.yaml` and `config/reject-email.txt` differ on the box (md5 compared).
  `compose.yaml` matches. All five timers are enabled. `/var/lib/dubplate` and `shadow` are `music`
  0755.
- `dubplate-shadow.timer` is not in `TIMER_UNITS` (`deploy.sh:526`), so a stop survives the apply.
- `install -d` (GNU coreutils 9.10, measured in the scratchpad): on an existing 0755 directory `-m
  0700` leaves it 0700. On a missing parent it creates the parent with the default mode and applies
  `-m` only to the leaf.
- The shadow's floor: a short attended pause costs nothing. The window is counted from the first
  successful pass (`internal/engine/shadow/run.go:89-92`), and a gap is counted only when a report is
  swept before it is consumed (`reconcile.go:49,78-81`). This closes R-B1's open item.
- `secret-set.sh --stdin` reads with `IFS= read -rsp` (`:120`), so a piped value works and stays out of
  argv.
- `doctorEnviron` (`doctor_test.go:908-916`) already exists as a green doctor fixture, and no test
  fixture sets `DUBPLATE_SERVE_PORT`.
- The ping helper maps a dash in a slug to an underscore (`_salvage/cmd/dubplate/pingfail.go:64`), so
  the `server-serve` slug resolves to `DUBPLATE_PING_SERVER_SERVE_URL`.

## Dispositions

### Convergent roots (fixed once)

| IDs | Root | Disposition |
| --- | --- | --- |
| **R-B1**, C-M1 | Task 10's sitting. The quiet-box check cannot pass against a 5-minute timer; there is no rollback; the binary swaps before `server.env`. | **Folded**, Task 10 rewritten. It adds a "Why the order matters" preamble. The quiet check drops the ten-minute clause and gates on the next quarter-hour boundary (step 5). The window stops the shadow timer and takes `.prev` copies of the binary and `server.env` (step 7). One manual shadow pass follows the apply (step 9), then the timer restarts (step 11). A written rollback sits in "If a step fails". The import-timer stop is refused, as the list above explains. |
| **R-M1**, C-M2, C-m9, K-m6 | The whole-estate apply. `--delete` pushes two changed files; compose, beets, and the timers run; the builds and refusals come after remote writes; no clean tree is required. | **Folded**. Task 0 item 10 adds the box reads (beets, config and compose md5, timer states, a names-only diff of the incumbent env). Task 10 step 6 lists the expected `--check` result for every step, and any other line is a stop. R-5d13, moved to Task 4D, puts the clean tracked-tree refusal (`--untracked-files=no`), the `web/` refusal, `npm ci`, the web build, and `make build-linux` above the first remote write. Task 10 step 1 requires a clean tree. |
| **R-M2**, K-M3, C-m1, K-m12 | "Local caller only" fails under a same-host proxy, and the compared address is unnamed. | **Folded**. R-5d8's peer rule is `PeerAddress`, unmapped `netip.Addr`, loopback or a bind address, and not in the trusted-proxy set; `ClientIP` is never consulted. The privileged-fields row states what holds and records the residual (an unconfigured same-host proxy). The residual also appears in Task 7 and in Task 8's residue list. R-5d9 adds `not probed: every local address is a trusted proxy`. Task 3 adds the trusted-loopback and `192.0.2.10`-naming-`127.0.0.1` rows, with mutations 4 (compare through `ClientIP`) and 5 (drop the exclusion). |
| **C-M6**, R-M3, K-m1 | The probe's header counting rests on recalled `textproto` behavior; mutation 4 is vacuous. | **Folded**. R-5d8 quotes `Header.Values` and `ReadMIMEHeader` from `go doc` and builds the request with `Header.Add` through one shared builder. Task 3's no-strip proof runs over `httptest.NewServer` and asserts the identity count equals the number of distinct canonical names. Mutation 4 is replaced by mutation 6 (count by map index) and mutation 7 (a strip that deletes only dash forms). |
| **C-M3**, K-M4 | The state-dir check proves array membership, not the pre-create. | **Folded** (one part refused, above). R-5d11 uses one loop over `"${DUBPLATE_SERVER_LEAVES[@]}"` and skips `shadow` by name. The check asserts the loop reads the array, runs `install -d`, and keeps the shadow line. Task 5 adds seeds 7 (a literal list) and 8 (the shadow line deleted). |
| **C-M5**, K-M5 | Doctor tests dial the host's real `127.0.0.1:8080`. | **Folded**. In Task 2's green baseline, every doctor test environment carries a reserved, closed `DUBPLATE_SERVE_PORT`. Task 6 uses the same port. Task 10 step 3 substitutes it. K-M5's other half (the parser appends, so deleting a row cannot fail) is folded in Task 6: fixture values replace existing rows and never append, and the named mutation deletes the roster row. |
| **C-M7**, K-M6, R-m1 | `net.InterfaceAddrs` under `RestrictAddressFamilies`. | **Folded**. R-5d7 quotes the netlink call from `net/interface_linux.go` and adds a fifth state, `not checked: <error>`, which fails and names the likely cause. The status function takes `InterfaceAddrs`'s result as arguments. Task 7 runs doctor without the family restriction and says why, as does Task 10 step 13. Task 2 adds mutation 4. |
| **K-M2** | The probe's method refusal is a 405 with `Allow`. | **Folded**. R-5d8 checks the peer, then the method, and both refusals answer `serve.Terminator()`. `methodOnly` is not used (verified at `serve.go:380-389`). Task 3 asserts non-local `GET`, non-local `POST`, and local `POST` each get a 404 with no `Allow` header. |
| **K-M1** | Namespace proof placement. | **Folded**. Task 5's evidence uses `/dev/shm/<unique>` for the leaves and a copy of the binary, drops `User=`, runs the positive half first in the same place, and quotes `man systemd.exec` for why `/tmp`, `/var/tmp`, and the home directory are excluded. |
| **R-M4** | The no-deploy rule rests on instructions alone. | **Folded, option 1** (R-5d16, Task 4D). No flag means check, only `--apply` writes, and both flags together are refused. Proved against `192.0.2.1`, where no mode can write. **Why option 1:** it makes the named failure, a dropped flag, safe in the tool, and it protects every later lane too. Option 2, conductor-only `--check`, is not taken: each task's `--check` evidence is that task's own acceptance, and the conductor never reads transcripts. Residual recorded: a deliberate `--apply` is stopped only by the dispatch prompt. |

### Contract lens (C)

| ID | Disposition |
| --- | --- |
| C-M1 | See R-B1. |
| C-M2 | See R-M1. |
| C-M3 | See the convergent row. |
| C-M4 | **Folded**, Global constraints and Task 2. `doctorEnviron` is the green baseline, and one test asserts it exits 0 with every row stated. Every exit-1 row is that baseline plus one perturbation. The existing fixture is reused, not rebuilt. |
| C-M5 | See the convergent row. |
| C-M6 | See the convergent row. |
| C-M7 | See the convergent row. |
| C-m1 | See R-M2. |
| C-m2 | **Folded**, R-5d9 and Task 3. Duplicate: K-m2. An absent bind address reads `not probed (not on any interface)`. `no listener` covers only a refused connection or a dial timeout. A connection that accepts and then hangs reads `unexpected answer: timeout` and fails. The dial and the response are bounded separately. |
| C-m3 | **Folded**, Tasks 5, 6, and 9. Duplicates: K-m5, R-m4. The false render-abort sentence is gone from Task 5. Task 6 states that `--check` stops at the render until the mint. Task 9 names how a `deploy.sh` fold is proved in the meantime. |
| C-m4 | **Folded**, Task 1 mutation 2 restated: skipping `Unmap` fails the resolver row and a limiter row keyed from the unmapped peer. |
| C-m5 | **Folded**, R-5d12 and Task 5. Duplicate: K-m11. `systemd-analyze` joins the hard-required list at the top of `check.sh`, so a re-added guard in the leg has no job. The "mutation" becomes a proof, with a named shim-`PATH` method to hide the tool. |
| C-m6 | **Folded**, Ruling 3's no and the exit gate's "Deferred, by name". |
| C-m7 | **Folded**, Task 6 (one part refused, above). Duplicate: R-m12. Parity table rows follow `deploy.sh:183-205`'s documented states. The self-mutation is replaced by a product mutation (delete the roster row). The test's doc comment names Task 10's `--check` as the render's only real proof. |
| C-m8 | **Folded**, Task 6. The task quotes which rows `Load` and `doctor` stat or open and overrides exactly those. "Not verified" carries the list until Task 6 quotes it. |
| C-m9 | See R-M1. |
| C-m10 | **Folded**, Task 11 errata. Duplicate: R-m3. Covers the route list, the R6 dial wording, and the `deploy.sh` closed list (R-5d13 and R-5d16). The erratum list below records each. |
| C-m11 | **Folded**, R-5d10. Duplicate: K-m9. The slug is `server-serve`, which maps to `DUBPLATE_PING_SERVER_SERVE_URL` (verified). |
| C-m12 | **Folded**, R-5d4 and Task 1. Duplicates: K-m8, R-m9. A hop that fails `ParseAddr` is parsed with `netip.ParseAddrPort` (quoted) and its address kept. A bracketed address with no port stays malformed. Adds Task 1 rows and mutation 5. |
| C-m13 | **Folded**, R-5d11. `install -d`'s behavior is measured, not recalled, and the loop's order after the shadow line is argued. "Not verified" carries the box's coreutils, and Task 10 step 10 reads the modes. |
| C-m14 | **Folded**, R-5d8 and Task 3. The server's handling of `Connection`-listed headers is cited from the mechanics lens's measurement. Task 3's wire-level no-strip row asserts sentinel count 1, which re-proves it in the gate. |
| C-m15 | **Folded**, Task 8. A by-design row cites the ratifying clause. UNI's High on the forward-auth shape is named as carried to rung 12. |
| C-m16 | **Folded**. Adds a `DUBPLATE_PUBLIC_URL` privileged-fields row. "SMTP transport not exercised on the box" is a declared gap carried to rung 16 (Task 8 residue, Task 11). No mail step is added: a live send needs the serve unit running and a forged-origin request, and it sends a real message during the sitting, which costs more than the gap. |

### Mechanics lens (K)

| ID | Disposition |
| --- | --- |
| K-M1 | See the convergent row. |
| K-M2 | See the convergent row. |
| K-M3 | See R-M2. |
| K-M4 | See C-M3. |
| K-M5 | See C-M5 (both halves). |
| K-M6 | See C-M7. |
| K-m1 | See C-M6. |
| K-m2 | See C-m2. |
| K-m3 | **Folded**. R-5d5: `client ip` reads `not checked` when `Load` failed. R-5d2 names flag-false-with-keys-set (`config.go:301`, "set without") and a non-boolean flag. Tasks 1 and 2 add the rows. |
| K-m4 | **Folded**, R-5d10. `TimeoutStopSec=30` covers the 25-second `shutdownGrace` (quoted). |
| K-m5 | See C-m3. |
| K-m6 | See R-M1. |
| K-m7 | **Folded**, Task 1 fixture. There is no stub mailer. An unparseable address consumes the bucket, because `allowLinkClient` runs before parse and send (`handlers.go:189-197`, verified). The fixture names the `Origin` and `Sec-Fetch-Site` values, and a spy route registered before `Handler()` (`serve.go:84-90`). |
| K-m8 | See C-m12. |
| K-m9 | See C-m11. |
| K-m10 | **Folded**, R-5d6 and Task 2. Doctor passes a discarding logger to `newWebRoutes`, and a neutral-branding row holds the line count. |
| K-m11 | See C-m5. |
| K-m12 | See R-M2. |

### Risk lens (R)

| ID | Disposition |
| --- | --- |
| R-B1 | See the convergent row. |
| R-M1 | See the convergent row. |
| R-M2 | See the convergent row. |
| R-M3 | See C-M6. |
| R-M4 | See the convergent row. |
| R-m1 | See C-M7. |
| R-m2 | **Folded**, Ruling 2's text, Task 8, Task 11, and the STATUS carry. Adds the BS ladder entry 5 and BS:510-513 errata. R21-only findings are marked carried, not absorbed. "To rung 12" names the doctor proof as well as the endpoint. |
| R-m3 | See C-m10; R-5d16's default flip joins the same erratum. |
| R-m4 | See C-m3. |
| R-m5 | **Folded**, Task 10 step 3. The local doctor runs under `env -i` over the rendered rows only, with scratch values substituted. |
| R-m6 | **Folded**, Task 10 step 2. Uses `openssl rand -base64 48` piped to `--stdin`, keeping the value out of argv. The registry note says why the age store applies (no Worker; the worker-only exception covers cairn per-site keys) and keeps both names out of `WORKER_SECRETS`. |
| R-m7 | **Folded** as a recorded residual. It appears in the privileged-fields row and Task 8's residue list, and Task 11 carries it to rung 16 (`LoadCredential=` or a dedicated uid). |
| R-m8 | **Folded**. Task 7 states the proxy's duty to overwrite or append the header, and that group 2 survives from a trusted peer. Task 8 and Task 11 carry "strip every group 2 name but the configured header from a trusted peer" to rung 12. The strip at `strip.go:455-466` was verified. |
| R-m9 | See C-m12. |
| R-m10 | **Folded**, Task 10's preamble and steps 12 and 15. The record keeps statuses, exit codes, and counts, never a member field. |
| R-m11 | **Folded**, Task 10's rollback: `systemctl stop` and `reset-failed` on the new unit first. |
| R-m12 | See C-m7. |
| R's note (not a finding): matcher changes ship into the shadow | **Folded** into Task 10 step 6's expected set as a named shadow-evidence event. |

## Rulings after the fold

1. **Two passes.** 5d-i (Task 0 and Tasks 1 to 6) runs at a 7.0M ceiling (flag 5.6M); 5d-ii (Tasks 7
   to 11) runs at 3.5M (flag 2.8M). Recommendation: yes. The split follows from Task 4D, which makes
   eleven tasks.
2. **The verify endpoint moves to rung 12.** Recommendation: yes. The substance is unchanged; the
   errata list is now complete.
3. **The attended first-deploy sitting**, now with the timer window, the `.prev` pair, and a written
   rollback. Recommendation: yes.

### Ceiling re-derivation, from "Token ceiling"

| Half | Base | Adders | Projection | Ceiling | Flag |
| --- | --- | --- | --- | --- | --- |
| 5d-i | 6 tasks × 0.76 = 4.56M | +0.30 heavy lane, +0.25 fold criteria, +0.25 5d-i close | 5.36M | 5.36 / 0.8 = 6.70 → **7.0M** | **5.6M** |
| 5d-ii | 2 tasks × 0.76 = 1.52M | +0.35 second lens, +0.25 whole-subsystem read, +0.50 close and sitting, +0.10 sitting steps | 2.72M | 2.72 / 0.8 = 3.40 → **3.5M** | **2.8M** |

The two projections total 8.08M, against revision 1's 6.72M. The ceilings total 10.5M, against
8.5M. Rung 5's reservation moves to about 48.5M. Under Ruling 2's no, 5d-i becomes 7 × 0.76 + 0.80 =
6.12M, a ceiling of **8.0M** with a flag of **6.4M**.

**Task count.** Under Ruling 2's yes, eleven tasks: 1, 2, 3, 4D, 5, 6, 7, 8, 9, 10, and 11. Under its
no, twelve. Either count passes ten, so the pass splits at the Task 6 seam, and 5d-ii follows as a
fresh session from the 5d-i close's resume prompt.

## Errata owed, recorded here and never edited into a ratified document

Task 11 pastes each of these into the spec's "Errata the close applies" section, or into the build
spec's errata home, dated:

1. **The rung 5 spec's 5d row (`:2519`).** Under Ruling 2's yes, the verify endpoint and its class
   move to rung 12. The pass runs as two passes in lane 1, not a worktree released at the close. The
   example-roster clause has already landed. "The env template's commented pending rows and its eight
   literal rows" becomes R-5d14's rows.
2. **5b-i's moved assertion**, "a verify subrequest from such a peer is refused", moves to rung 12.
3. **BS ladder entry 5 (`BS:707-714`)**: "the forward-auth verify endpoint with its own trust-boundary
   documentation" moves to rung 12. This is Ruling 2's erratum on the build spec.
4. **BS:510-513**: "`dubplate doctor` proves the boundary by verifying that a spoofed identity header
   from a non-trusted source is rejected" is carried to rung 12 for the forward-auth topology. 5d's
   probe proves serve's own strip.
5. **The owned-prefix route list (spec `:915-918`)** gains the `boundary-probe` route.
6. **R6's wording (spec `:366-367`)**: "dials the service's own configured bind address over
   loopback" becomes "dials each configured bind address from that address".
7. **"Deployment" (spec `:1765`)**: "`deploy.sh` gains the unit and the two pre-creates, and nothing
   else" gains R-5d13 (the clean-tree refusal and the local build before the first remote write) and
   R-5d16 (check by default, `--apply` to write).
8. **The environment template section (spec `:1657-1667`)**: the four SMTP rows are bare, not literal,
   and "none is ever minted" is stale. Under Ruling 3's yes, "rung 16 mints all three" becomes "5d mints
   the two HMAC keys; the SMTP password was minted 2026-08-31".
9. **R23 (spec `:2799-2837`, "Process")**: 5d's 3M becomes 7.0M plus 3.5M, and the rung reservation
   becomes about 48.5M.
10. **The 5d exit gate's namespace clause**: it is recorded as Task 5's `/dev/shm` workstation
    evidence plus Task 10's box start.
11. **The spec's existing paste-ready blocks**: ladder entry 5 (page stack), ladder entry 7, B12, B4,
    and every other block in "Errata the close applies" not yet applied. These are carried from
    revision 1.

**Cross-lane notes.** These are not errata, because the translit plan is an approved plan, not a
ratified document; the conductor writes them on lane 2's registry row.

- Once Task 4D lands, translit's apply command is `deploy.sh --apply`.
- The translit plan's rollback ("redeploying the previous commit", its `:1271`) fails for every commit
  from 5b-i on, because each one renders `DUBPLATE_IDENTITY_SOURCE` without secrets. After 5d's
  sitting, a rollback restores a saved binary and `server.env` pair, as Task 10's does.
