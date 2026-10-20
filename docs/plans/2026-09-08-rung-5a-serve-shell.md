# Rung 5, Pass 5a (The Serve Shell and Ingress) Implementation Plan

**Revision 2, 2026-09-08.** Revision 1 went through a two-lens fresh
context adversarial review, ingress security and plan integrity, folded
here. The record is
`docs/research/2026-09-08-rung-5a-plan-review.md`, whose
thirty-item "Verified sound" list is the do-not-re-litigate record for
this pass and whose five corrections to the reviews' own evidence carry
the measurements. Twenty-three findings folded: three blockers (the
strip is specified over `r.Header` alone so a trailer-borne identity
header survives ingress; R-D states rule 1 two ways, one vacuous and one
a blanket ban on percent-encoding; the ceiling anchors on a miscited row
and omits two costs this pass mandates), ten majors, and ten minors.
Nine folds are conductor-owed and carry an inline `ruled,
conductor 2026-09-08, R-x` marker: R-D, R-O, and the seven new
resolutions R-P through R-V. The ceiling and the ruling of all
twenty-two resolutions at dispatch are two more conductor decisions,
argued where each belongs. The task count does not change: Task 0 plus
eight. The ceiling moves from 4M to 4.25M, the 80% flag from 3.2M to
3.4M, and the declared coverage gaps from seven to six, because the
`IP_FREEBIND` boot property turned out to be measurable here. Fourteen
citation pointers were re-resolved and corrected. **The conductor ruled
on this revision 2026-09-08: all twenty-two resolutions R-A through R-V
stand as applied, R-D stands with a rung-12 revisit clause, the 4.25M
ceiling stands with its 3.4M flag routed to the conductor, and the
1.25M overage is recorded against R23's reservation.**

Drafted against the rung 5 identity-and-auth spec at **revision 2.1**
(`docs/specs/2026-09-08-rung-5-identity-auth.md`, landed at
`c5021ce`), whose six product forks Geoff ruled on 2026-09-08 and whose
thirty-eight rulings are applied. **No fork touches 5a's surface.** Fork
1 moves the roster, fork 2 the mail transport, fork 3 the thresholds,
fork 4 the enumeration copy, fork 5 the admin bootstrap, and fork 6 the
typed code, and every one of them lands at 5b, 5c, or 5d. No task in this
plan carries a fork marker.

**What this pass is.** 5a is the first of rung 5's four passes
(`spec:2031-2057`). It builds `internal/serve` and nothing else that
runs: the root composition order with its three-group ingress strip, the
six-rule path normalization, the case-sensitive segment-aware prefix
match with its terminator, the declared extension point, the empty
fallthrough slot, the `http.Server` with its deadline policy and the
`ResponseController` extension mechanism, the `IP_FREEBIND` listener set,
R27's two-address derivation, the `serve` reserved class retired into
three declared keys, and `newServeCmd` wired into the binary.

**5b builds the auth core, the identity leaf, and the mail seam. 5c
builds `web/`, the `member` verb, and the gate's web leg. 5d builds the
boundary endpoint, the deployment, and the rung close.** None is in scope
here, in any part. A task that believes it needs a token, a session
cookie, a roster, a rate limiter, an origin check, a page, a unit file,
or a `doctor` row stops and reports.

**Band: L.** `spec:2033` sets the rung at L band across four passes, and
5a inherits it. The band is what argues for a plan of this size against a
pass whose whole deliverable is one new package: three of rung 5's four
passes change a contract, and 5a's contract is the extension point every
later consumer builds against.

**Nine items, under no split trigger.** `spec:2049-2057` names a
fifteen-task split trigger for 5b alone and none for 5a. This plan lands
**Task 0 plus eight tasks**, so 5a stays one pass. The cut point is named
anyway in the execution model, because a re-derivation past the ceiling
at the Task 3 checkpoint is where the call would be taken.

**The composition order is the pass's own risk control.** Risk 1
(`spec:2221-2229`) is a defect no route table shows: a normalization
check registered as a route runs after `http.ServeMux` has already
cleaned the path and issued its redirect, and the mistake reads as
correct in review. Everything in this plan that asserts on the chain
rather than on a route exists for that risk. `spec:2043-2047` is the
argument for the whole pass boundary: the shell's composition order is
the artifact the entire trust boundary rests on, and reviewing it in the
same pass as the token and session mechanics would bury it.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain per task, sequential. The implementer
> is `cairn-implementer`, instructed to read and follow
> `.claude/agents/dubplate-implementer.md` before anything else. The
> conductor consumes reports and never reads diffs inline. Subagents
> never dispatch other agents; `code-simplifier` is the conductor's
> dispatch after a task's review accepts, and so is the pass-end
> `web-auth-security-reviewer`. The conductor owns the ledgers
> (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane registry,
> the contested-surface rows, and the batch commit of anything a task
> drafts into them. **This pass runs in LANE 2, a git worktree under
> `.claude/worktrees/`**, concurrently with pass 8b in the main tree. No
> task runs from the main tree, and no second executor enters the
> worktree.

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/specs/2026-09-08-rung-5-identity-auth.md`,
   revision 2.1.** The pass-split table's 5a row (`spec:2035`) is this
   plan's scope statement. The sections this pass implements are "What
   rung 5 builds" (`:87-141`), "The trust boundary" in full (`:765-904`,
   whose four subsections carry the strip, R27's two addresses, the
   forward-auth topology this pass does not build, and the one owned
   prefix), "The serve shell" in full (`:905-1005`, the composition
   order, the extension point, the deadlines, and the wiring), "Packages
   and their contracts" (`:1459-1491`), the `Serve` config table
   (`:1560-1578`), the 2.0 hook table (`:1835-1854`), the prior-art table
   (`:1855-1902`), the threat model (`:1903-1975`), the contested
   surfaces (`:2058-2091`), the band and ceiling (`:2092-2152`), the
   staleness pre-task (`:2153-2218`), the risks (`:2219-2313`), and the
   process section (`:2314-2363`). Rulings R6, R15, R20, R22, R26, R27,
   R32, and R33 are the ones 5a reads; the rest belong to later passes.
2. **`docs/research/2026-09-08-rung-5-spec-inventory.md`
   (INV), section 2 in full (`:229-387`), section 3 (`:388-522`), section
   9 (`:1050-1094`), and its thirteen surprises (`:1187-1256`).** Section
   2 is the serve shell as it exists in salvage beside what PA requires,
   subsection by subsection: 2.1 the composition order, 2.2 the strip
   set, 2.3 normalization, 2.4 the write timeouts, 2.5 the verb wiring,
   2.6 the unit this pass does not write. **Do not re-derive what INV
   settles; cite it, and verify a line you carry forward.**
3. **`docs/specs/2026-09-03-parallel-build-approach.md` (PA),
   its serve-seam paragraph (`PA:77-80`).** PA:79 is the requirement the
   spec discharges and this plan builds: the root-level composition
   order, the declared extension point, and which point the driver's
   proxy occupies. PA:80 is the scheduling rule Task 0 checks.
4. **`docs/specs/2026-09-02-greenfield-server-build.md` (BS),
   the auth birth requirements (`BS:393-523`).** Four clauses reach 5a:
   **B2** the forward-auth boundary and its `doctor` proof (`:409-425`),
   **B3** one owned prefix (`:427-432`), **B8** rate limits and abuse for
   its client-IP-mode half alone (`:478-490`), and **B12** proxy hygiene
   (`:512-523`), whose strip, normalization, and per-route write-timeout
   clauses are shell-level and therefore rung 5's, which is INV's
   surprise 5.
5. **`docs/research/2026-09-08-rung-5-auth-prior-art.md`
   (PRA)**, its lessons section (`:525-604`) for Authelia's
   forwarded-header rule and Miniflux's header-plus-trusted-network
   pairing. Its library table is 5b's; 5a adds no dependency.
6. **`_salvage/internal/serve/serve.go`, read-only, for four things
   only**: `mountBasePath` (`:82-107`), whose own doc comment records
   that `http.StripPrefix` does a plain string trim so `/uploadfoo` trims
   to `foo` and passes through; `securityHeaders` (`:110-128`), whose
   header set R32 adopts at 5c; `freebindControl` and `freebindListeners`
   (`:561-596`), the `IP_FREEBIND` mechanism R20 adopts; and the timeout
   constants and their single `http.Server` (`:598-633`), whose one
   global 30-second write timeout is the lesson B12 answers. **Nothing is
   copied. It is adoption material, never an import.**
7. **The tree, read-only:** `internal/config/registry.go` (the four
   declaration families, `requirePrefix`, the collision panic, and
   `validateThresholdDefault`); `internal/config/keys.go` (the `serve`
   reserved class at `:74-85`); `internal/config/config.go`'s `Load`
   (`:241-279`); `cmd/dubplate-server/main.go`'s `newRootCmd` (`:62-80`);
   `cmd/dubplate-server/shadow.go:19-60` as the declare-in-your-own-file
   precedent; `cmd/dubplate-server/doctor.go` and `doctor_test.go`'s two
   line-count assertions; `scripts/check.sh`.
8. **`docs/HISTORY.md`'s rung 1b entry** (`:485-500`), the registry
   contract in the words that bind: a rung declares its keys in its own
   file, no lane adds a name to `internal/config/keys.go`, and a lane
   removes only its own reserved class from that file in the same commit
   that declares the class in its own package. **Rung 5 is the first to
   exercise the removal half, and 5a is the pass that does it.**
9. **`CLAUDE.md`** in full, and `docs/STATUS.md`'s lane registry.
10. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** This
    pass writes Go only. No Python, no Svelte, no TypeScript. **No unit
    file and no deploy path exists in this pass** (R-K below), so
    `vps-conventions` is out of scope.

## Goal

Ship `internal/serve`: the HTTP shell every later rung mounts into, whose
root chain strips every inbound identity header before any route runs,
refuses a smuggled path before prefix routing decides anything, matches
its one owned prefix on segment boundaries and case-sensitive bytes,
terminates that prefix so no unknown owned path escapes, hands everything
else to a fallthrough slot that is empty at rung 5, and serves it all
from an `http.Server` whose deadlines protect an ordinary response
without killing a stream a later rung adds.

Land the three-group ingress strip with its underscore twins and its
hop-by-hop removal, the six-rule normalization, the segment-aware prefix
match and its terminator, the registration seam rungs 6, 7, 12, and 14
add a file at, R27's peer-address and reported-client-IP derivation, the
`ResponseController` write-deadline extension, the `IP_FREEBIND` listener
set, the three `serve` config keys declared in rung 5's own file with the
reserved class retired in the same commit, and the `serve` verb wired
into the binary.

Nothing here mints a token, sets a cookie, reads a roster, sends mail,
renders a page, limits a rate, checks an origin, writes a systemd unit,
touches the box, adds a module dependency, or reaches the live network.

## Owner touchpoints: none

5a raises no attended question. `spec:2356-2362` names the rung's three
expected touchpoints, and all three sit elsewhere: the six product forks
are ruled and closed, the `style-src 'unsafe-inline'` sitting rides with
5c's pages, and the box deploy is 5d's. **Anything else is a conductor
ruling** (`spec:2362`).

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **Any name added to `env/dubplate-server.env.template`.** That file is
  a secrets manifest `deploy.sh` renders from the age store, and it
  hard-fails the whole render on one unresolved name (CLAUDE.md, "Facts
  agents get wrong"). 5a's three keys are thresholds carrying defaults,
  so none needs a template row and none is minted. The template's rows
  are 5d's (`spec:1624-1651`). **A task that finds itself wanting a
  template row stops and reports.**
- **Any adopter-visible default.** The prefix default, the bind default,
  and the port default are what an adopter inherits when they configure
  nothing. R-A and R-B below rule all three as applied, conductor
  2026-09-08, and each states what would change under a different
  answer. A task changes none
  of them on its own judgment.
- **Any change to what the box runs.** 5a deploys nothing and enables
  nothing. The box's live surface is Navidrome, cloudflared, the guardian
  timers, and the shadow timers started 2026-09-08.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor, and the
80%-of-ceiling flag routes there too.

## Spec

`docs/specs/2026-09-08-rung-5-identity-auth.md` (revision
2.1), and behind it `specs/2026-09-02-greenfield-server-build.md`
(revision 2.1, ratified), whose auth birth requirements are quoted at
`BS:393-523`. 5a builds the third bullet of the spec's seven-part "What
rung 5 builds" list (`spec:101-108`), less the embedded-asset path and
the served-document rewrite, which DL R4.5 gives to 5c.

**Hook 6 is kept open rather than landed.** The spec's hook table
(`spec:1843`) records hook 6, the driver seam staying universal, as
"Neutral, and kept open. Rung 5 leaves the fallthrough empty and the
strip set as data." **Those two mechanisms are 5a's, and they are the
whole of 5a's position on the 2.0 hooks list.** Hook 5, the stable
identity, is landed by 5b and 5c and 5a does not reach it. Every other
hook is schema, tagging, layout, or migration work rung 5 does not touch.
Each task below states its own position, so the build-for-2.0 sweep reads
a decision at every task rather than an omission.

**The forward criterion this pass hands rung 12**, so rung 12 reads it
from the ledger rather than rediscovering it:

> Rung 12 contributes its driver's configured header name through the
> strip's data path and occupies the one root-level fallthrough slot. It
> writes no line of rung 5's strip code, adds no case to the composition
> chain, and inherits the write-deadline extension mechanism rather than
> changing the server's defaults. Rung 14 registers its API subtree at
> the same extension point and inherits the prefix terminator that keeps
> an unknown API path from reaching rung 12's fallthrough.
>
> Four properties of the ingress travel with the slot, so rung 12 builds
> against them rather than discovering them. The strip clears trailer
> fields as well as header fields, so a driver may not read identity from
> a trailer. Group 2 covers the whole `X-Forwarded-*` family, so a
> proxied request carries none of them unless the peer is an allowlisted
> proxy. No path reaching the fallthrough carries percent-encoding, a
> dot segment, a doubled separator, or a backslash, so a driver
> forwarding the raw path forwards a path already normalized to that
> shape. And an exempted identity header arrives collapsed to its last
> value.

## Token ceiling: 4.25M (re-derived from 8a's actual)

`spec:2101` proposes 3M for 5a and `spec:2149-2151` grants the
re-derivation in the same section: "Each ceiling is re-derived from the
previous pass's actual report before the next plan is drafted." The
previous pass's actual report is 8a's, and it does not support 3M.

**The basis.** INV's budget table (`INV:1101-1113`) reads 8a at about
3.6M against a 4.5M ceiling, for Task 0 plus nine chain tasks in one
tree, pure Go, no corpus and no network. That is 5a's shape almost
exactly. Two other patterns bind. Plan preparation is the largest single
line in every rung 3 pass (`INV:1117-1121`), and **ceilings are overrun
when the pass is a contract change rather than a build**: rung 1b ran
3.7M against 2.5M and crossed its 2.0M flag inside the chain with no
conductor action (`INV:1123-1128`). 5a is a contract change. It creates
the extension point every later consumer builds against, and it exercises
the registry's removal half for the first time.

**The arithmetic, anchored on 8a's actual rather than summed from
scratch.**

| Line | Basis | Figure |
| --- | --- | --- |
| 8a's measured total, Task 0 plus nine chain tasks | `INV:1113` | 3.60M |
| One fewer chain task (eight here, nine there) | 4a-i's measured chain rate, 2.14M over nine tasks, about 238K | −0.24M |
| Heavier prep | The rung 5 spec is 2,395 lines against rung 8's, and INV is 1,371; the close-read is larger even though the spec does the reasoning | +0.10M |
| The pass-end `web-auth-security-reviewer` 8a had no equivalent of | One reviewer pass, priced as `spec:2115` prices R24's second read at 5c | +0.25M |
| Per-task `code-simplifier` and its review, which 8a's actual does not carry | `HISTORY:367-368` records the simplifier as owed rather than run; `HISTORY:1719` prices 4a-i's at 0.15M plus a 0.05M review | +0.20M |
| The merge ritual, which 8a ran in the main tree and never paid | Rebase, in-tree gate, on-master gate, and the conditional post-rebase `diff-reviewer` | +0.05M |
| The revision 2 folds, which add rows to six tasks | Twenty-three findings, none adding a task | +0.10M |
| **Expected total** | | **4.06M** |

**Two of those lines are the review's blocker B3**, and they are why the
number moved. 8a is this plan's anchor and 8a paid neither cost: its
forward carry 1 records the simplifier as owed at 8b's opening, and its
own plan states that it ran in the main tree. A ceiling derived from 8a's
3.6M without adding them prices a pass this plan does not run.

**The ceiling is ruled at 4.25M, conductor 2026-09-08, as applied,
because 3M schedules its own stop rule to fire before the first task
lands.** Prep and Task 0 alone are about 1M. An 80% flag at 2.4M would
fire around Task 5 of eight, which is not a signal, it is the
arithmetic announcing itself. At 4.25M the flag lands at 3.4M, past the
chain and inside the close and the reviewer, where a stop is cheap and
the conductor's action is a real choice: run the reviewer in this pass
or hand the read to 5b.

**Recorded against the rung reservation.** `spec:2101` and R23
(`spec:563-570`) set rung 5 at 14.5M across four passes. At 4.25M for 5a
the rung total stands at **15.75M** until 5b closes. **This plan does not
take a giveback from another pass**, because 5b, 5c, and 5d each own
surfaces 5a cannot price, and the spec already schedules their
re-derivation from actuals. The conductor records the 1.25M overage
against the reservation and re-derives 5b from 5a's actual report.

**Flag at 80%, which is 3.4M.** Finish the task in flight, write STATUS,
and send one combined question to the conductor. **The conductor's action
when it fires is the choice this plan already names**: whether the
pass-end reviewer runs here or at 5b, and whether the close absorbs the
spec corrections or defers them. The arithmetic to check at that point is
3.4M against a 4.06M expectation, leaving 0.66M of forecast against 0.85M
of headroom.

**Re-derive the ceiling from Task 3's actual report.** Task 3 is the
normalization, the pass's one novel-mechanism task and its one `opus`
upshift. If the re-derivation passes 5M, the conductor checks the session
window cap before continuing and names the split at the Task 4 / Task 5
boundary.

**Checkpoint: STATUS after every four tasks**, so after Task 3 and after
Task 7, plus one at the close. Every task reports its own spend, its
mutation outcomes, and its measured gate wall time, and the conductor
sizes the next task against it. **Every task boundary is a resumable
stopping point** (`spec:2143-2147`), because pass 8a's chain met a
session window cap before its own ceiling would have bound.

## Global Constraints

- **Lane 2, a git worktree.** `spec:2035`'s tree column for 5a is
  "Worktree", and CLAUDE.md's Lanes section is the rule behind it: the
  gate runs the module leg over `./...` plus two whole-repo legs, so two
  code lanes need two trees. **Lane 1 runs pass 8b in the main tree,
  writing `internal/engine/tagio`, concurrently with this pass.** The
  lane registry entry is written before the first dispatch, and **the
  live-executor check is that registry, never `pgrep -f
  Projects/dubplate`**, which matches every lane once a worktree exists.
- **Package territory, stated as a wall.** This pass writes
  **`internal/serve`** (born this pass, including its tests), **one new
  file under `cmd/dubplate-server`** (the serve verb and its config
  declarations), **the `AddCommand` block and the two register calls
  above it in `cmd/dubplate-server/main.go`**, and **the deletion of the
  `serve` reserved class from `internal/config/keys.go`**. It writes
  **nothing** under `internal/engine`, `internal/harness`,
  `internal/atomicfile`, `corpus/`, `web/`, `systemd/`, `scripts/`,
  `env/`, `tools/`, `Makefile`, or `_salvage/`, and no `docs/` file
  outside Task 8's conductor-owned drafts. **`internal/config` is written
  in exactly one place**, the reserved-class deletion, and nowhere else:
  `config.go`'s `Load` is R29's edit and belongs to 5b (`spec:617-630`).
  A task that believes it needs one of those stops and reports.
- **`internal/serve` is closed to every other lane until rung 5 closes**
  (PA:80). Task 0 item 2 is the blocking check that no lane has written
  it. A violation invalidates 5a's whole design, which is why the spec
  makes it the pre-task's second question (`spec:2164-2166`).
- **No auth, no pages, no unit.** 5a's row says "No auth, no pages"
  (`spec:2035`) and its exit gate names no unit. The serve unit, its
  `ReadWritePaths` grant, the `deploy.sh` pre-creates, and the
  generalized state-directory check are 5d's whole first deliverable
  (`spec:2041`). **A task that reaches for `systemd/` stops and reports.**
- **The shell reads configuration through values passed in, never from
  the environment.** `spec:1470-1478` states the rule for
  `internal/auth`, and `spec:1474-1479` gives `internal/serve` the same
  shape by giving it the `http.Server`, the mux, and the strip while
  leaving resolution outside. `internal/serve` imports
  `internal/config` for nothing. Its constructor takes plain values.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects. R26 (`spec:585-594`) budgets rung 5 exactly two,
  `github.com/wneessen/go-mail` and `golang.org/x/time/rate`, and **both
  are 5b's**. 5a's every mechanism is stdlib: `net/http`,
  `net/textproto`, `net/url`, `net`, `syscall`. **One exception needs
  naming rather than assuming**: `IP_FREEBIND` is a Linux socket option,
  and the salvage code reaches it through `golang.org/x/sys/unix`
  (`_salvage/internal/serve/serve.go:570`). R-M below rules the stdlib
  `syscall` route instead, so the dependency count stays at zero.
- **Assertion discipline, at the repo's bar.** No silent-success fakes,
  no vacuous assertions, **no expectation derived from the code under
  test**. Every hard-coded expectation in this pass comes from the spec,
  from an RFC, or from documentation a task cites by URL and date. A test
  that asserts the strip set by asking the strip set for its own contents
  proves nothing; the enumeration is written out.
- **Every ordering property is proved by behavior, never by structure**
  (`spec:2224-2226`). A test that reads the chain's construction proves
  the code says what it says. A test that sends a request and reads what
  a terminal handler received proves the order. Risk 1 is the reason.
- **Universality.** No estate hostname, path, email, catalog value, or
  personal name anywhere (`spec:2364-2371`). Every fixture address is a
  documentation-reserved literal or an invented one. **Every tunable is a
  threshold with a default** (`spec:2380-2383`), and 5a's three are the
  prefix, the bind addresses, and the port.
- **Drivers own driver knowledge** (`spec:2373-2378`). The header name,
  the injection rules, the bypass paths, and the post-login destination
  live with the driver at rung 12. **5a ships the mechanism by which a
  driver contributes them and no driver's answer.** A task that hard-codes
  a Navidrome path, a Navidrome header, or `/rest` as anything but a
  comment stops and reports.
- **`go-conventions` before any Go.** The bar is the standing
  exemplary-stdlib-idiom bar (`idiomatic-beautiful-go`). Concrete types
  over interfaces; **no interface in `internal/serve` at all**, since the
  extension point takes `http.Handler`, which is the stdlib's own
  interface and not a new one. `any` not `interface{}`, no `init()`,
  guard clauses over nesting, `ctx` first, errors lowercase and
  context-first with `%w` only where a caller branches. **The em dash is
  banned in Go comments**, and vale scans a brand-new file before it is
  staged.
- **Sentinels where a caller branches.** The shell's refusals are HTTP
  statuses rather than Go errors, so `errors.Is` has little surface here.
  Where a constructor or a parse fails, the error wraps a package
  sentinel and the test asserts through `errors.Is` rather than on the
  message text.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  **in the worktree**. `make check` (vet, golangci-lint, test,
  tidy-check) over `./...`, shellcheck over `scripts/*`, ruff over every
  tracked `*.py`, gitleaks over the tracked worktree with `--no-git`,
  vale over Go comment prose, `docker compose config`, and a
  pattern-filtered `systemd-analyze verify`. **Never commit red. Specific
  files only.** Commit footer `Co-Authored-By: Claude
  <noreply@anthropic.com>`.
- **`.gitleaksignore` is the tracked empty file rung 1b left**, and this
  pass expects to append nothing: `internal/serve`'s fixtures are header
  names, paths, and loopback addresses, none of which is key-shaped. **If
  gitleaks flags anything, the fix is a fingerprint appended with its
  reason, and the task reports it** (RE cell (5,2), `spec:2081`). A
  concurrent append from lane 1 makes the rebase non-fast-forward and
  triggers the merge ritual's re-review.
- **Every task stamps its resolutions' markers.** A task landing the
  surface a resolution decides writes `// ruled, conductor 2026-09-08,
  R-x` on the symbol that carries the decision, in the repo's existing
  convention (`internal/config/registry.go:33`, `keys.go:35`,
  `cmd/dubplate-server/main.go:62`). Task 8's resolution table is then a
  grep rather than a reconstruction.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

## Contested surfaces

The spec's own table (`spec:2058-2091`) is the source. The conductor
writes these into the parallel-approach spec's ownership table before the
lane launches, per lane-launch checklist item 2.

| Surface | Owner | Consumers | 5a's position |
| --- | --- | --- | --- |
| `internal/serve` shell, mux, ingress strip | rung 5, created by it; rung 12 thereafter, **with no concurrent serve-writing lane** (PA:80) | 6, 7, 12, 14 add files at the declared extension point | **5a creates it, whole.** 5b mounts auth handlers into it; 5c adds the embedded assets, the document rewrite, and the served header set; 5d adds the verify endpoint. |
| `internal/config/keys.go` edits | rung 5 removes the `serve` and `auth and identity` reserved classes and the pending identity-source wall row, each in the commit that declares its replacement elsewhere | every later rung, by the same rule | **5a removes the `serve` class only** (Task 7). The `auth and identity` class, the pending wall row, and the three supersession-message rewrites are 5b's. |
| `internal/config/config.go`, `Load`'s owned-row resolution | rung 1's contract; rung 5 adds the identity-source require and the per-source correlated resolution (R29) | every rung that adds a required owned row | **Untouched by 5a.** That edit is 5b's, whole (`spec:617-630`). |
| `cmd/dubplate-server` root command and `doctor` rows | rung 1's skeleton; rung 5 adds `serve`, the `member` verb group, and the probe row kind | 6, 9, 13 | **5a adds `newServeCmd` and one `registerServeConfig` call** (Task 7). The `member` verb is 5c's; the probe rows are 5d's. |
| `.gitleaksignore` | rung 1b retired its one fingerprint; the lane that needs an allowed finding appends and names it in its pass report | every lane | **5a expects to append nothing** and reports it if it must. |
| `env/dubplate-server.env.template` | rung 4 by its own header; rung 5 adds commented pending rows only | rung 16 mints | **Untouched by 5a.** 5a's three keys are thresholds with defaults and need no row. |
| `systemd/`, `scripts/check.sh`, `scripts/deploy.sh` | rung 4 owns the units today; rung 5 adds the serve unit and the gate's web leg | every lane | **Untouched by 5a.** The unit and the state-dir generalization are 5d's; the `web/` leg is 5c's. |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. |

**Lane 1 runs pass 8b concurrently, in the main tree, writing
`internal/engine/tagio`.** Its territory and 5a's are disjoint by
package, and the two lanes share the tree only through
`.gitleaksignore`, `go.mod`, and `docs/`. **5a writes none of the three**,
so the expected rebase at the merge ritual is a fast-forward. Task 0 item
1 confirms the registry agrees.

**The merge ritual closes the lane**, stated here because the close task
executes it: the conductor takes the merge token, rebases the worktree
branch onto master, runs the full gate in-tree, merges `--ff-only`, runs
the full gate on master, and dispatches **one `diff-reviewer` over the
merged result whenever the rebase was not a fast-forward**. The second
code slot is released at this pass close rather than at the rung close
(`spec:2143-2145`), so the worktree is removed once master is green.

## The shell's contract, as outcomes

Every clause below is a property the pass ships. Each names the task that
lands it and the spec clause that requires it. A task builds to this
section; it does not re-derive it.

### C1. The root composition order

Four steps, in this order, each wrapping the next (`spec:907-944`, PA:79).

1. **The ingress strip.** It wraps everything, so no route is reachable
   with an identity header still attached. Task 2.
2. **Path normalization.** It refuses four shapes before prefix routing.
   It wraps the mux rather than registering as a route, because
   `http.ServeMux` performs its own path cleaning and redirect
   internally, so a check registered as a route would run after the
   decision it exists to guard. Task 3.
3. **Prefix routing.** The one owned prefix is matched segment-aware on
   case-sensitive bytes, never by a plain string trim. Task 4.
4. **The fallthrough.** Empty at rung 5. Rung 12's reverse proxy occupies
   it, which is the answer PA:79 asks for and PA:77 argues: by the
   one-owned-prefix rule the streamer owns everything else, so the driver
   is the root mux's fallthrough rather than a mounted subtree. Task 5.

**R27's derivation runs before step 1** and is not a step of its own,
because it decides nothing and refuses nothing. It computes two values
the strip's second group and every later rung's limiter read. Task 1.

**The order is invisible in a route table**, so three behavioral
assertions hold it (`spec:938-944`): a request bearing every spelling in
the strip set reaches a handler with none of them; a request whose raw
path and decoded path disagree is refused before any route runs; a path
that string-prefixes the owned prefix without matching a segment boundary
does not enter the owned surface.

### C2. The three-group ingress strip

The strip runs as a wrapper around the mux rather than inside any
handler, which is what makes "every path" true without enumerating paths
(`spec:767-771`). Three groups (`spec:772-812`).

**Group 1, identity spellings.** Two halves. The **driver's configured
header name** is the first, and at rung 5 no driver exists, so that half
is empty and **the mechanism by which a driver contributes its name is
what rung 5 ships**: a driver declares its header name through the driver
seam at rung 12, and the shell reads the declared name at construction.
**Rung 12 never edits rung 5's strip code.** The second half is a fixed
extras list, **each name in both its dash form and its underscore form**.

The underscore class is what SEC:42 names as the one that bites adopters,
and the Go mechanism is invisible, so it is asserted rather than
remembered: `textproto.CanonicalMIMEHeaderKey` uppercases the letter
after each dash and treats the underscore as an ordinary token character,
so `Remote_User` and `Remote-User` are two distinct map keys and a delete
of one leaves the other standing (`spec:779-787`). **The plan asserts that
behavior with a test rather than asserting it from memory.**

**The strip is written so a new spelling is one line of data**
(`spec:787-788`), and both spellings are generated from one source rather
than typed twice (`spec:2231-2236`).

**The strip clears trailer fields as well as header fields.** *(Ruled,
conductor 2026-09-08, R-P, as applied.)* Go's server moves every announced trailer key
out of `r.Header` into `r.Trailer` before the handler runs, and it removes
the `Trailer` announcement from the header map, so a strip that reads
`r.Header` alone cannot reach them. Measured on go 1.27.1: a chunked
request announcing `Trailer: Remote-User` and sending `Remote-User:
attacker` after its body arrives with `r.Header` clean and `r.Trailer`
carrying the value once the body is read. `httputil.ReverseProxy` forwards
trailers to the upstream, so rung 12 would hand the value on. **Every name
in the strip set is deleted from `r.Trailer` on the same pass as
`r.Header`**, and the pass's first claim names both maps.

**The exemption, which rung 5 ships as a mechanism and never exercises.**
Exactly one identity header may ever be exempted from the strip, it is
the one the active front's configuration names, and the exemption applies
only when the peer address matches that front's configured addresses.
Every other spelling is still stripped, the exempted header's underscore
twin included. **While magic-link is the active source the exemption is
impossible**, on BS:499-503's rule that the inactive source is never
consulted, and no front ships at 1.0 (`spec:849-862`). **A request
carrying the exempted name twice is collapsed to its last value** before
anything reads it. *(Ruled, conductor 2026-09-08, R-Q, as applied.)* A front
that appends rather than replaces produces exactly that shape, and
`Header.Get` returns the first value, so the rule decides which of the two
a downstream sees rather than leaving it to arrival order.

**Group 2, the forwarded-address group.** `X-Forwarded-For`, `X-Real-IP`,
`Forwarded`, and the `X-Forwarded-Host`, `X-Forwarded-Proto`, and
`X-Forwarded-Prefix` family. They are stripped on every path **unless
trusted-proxy mode is enabled and the peer address matches**, which is
R27's carve-out (`spec:789-794`). SEC's proxy-hygiene major asks for the
first three by name. **The `X-Forwarded-*` family joins them**
*(ruled, conductor 2026-09-08, R-R, as applied)*, because B12 asks for proxy
hygiene as a class and those three steer any absolute URL a downstream
builds: they reach the streamer through rung 12's fallthrough and 5b's
handlers through the owned prefix, both as client-asserted values. The
carve-out is the same one, since a trusted front sets them legitimately.

**Group 3, hop-by-hop hygiene.** Serve removes every header a request's
own `Connection` header lists, on ingress, before routing
(`spec:796-806`). Without it a client sending `Connection: Remote-User`
causes rung 12's freshly injected identity header to be deleted before it
leaves, **a downgrade any attacker can induce**. The injection-time
consequence stays rung 12's, inside `Rewrite`. R-O below rules what
happens to the fixed RFC hop-by-hop names, which the spec leaves open,
and rules the upgrade case the listed removal would otherwise break.

**Group 3 carries a keep-set, so the removal is not a client-controlled
deletion primitive.** *(Ruled, conductor 2026-09-08, R-S, as applied.)*
`Cookie`, `Origin`, `Sec-Fetch-Site`, and the shell's own session-cookie
name are never removed by the listed rule, whatever a `Connection` header
says. Each is an end-to-end header a client has no business declaring
hop-by-hop, and 5b's origin check reads two of them. **No live attack
rests on this**, and the rule is hygiene. The Fetch standard lists
`Connection` among the forbidden request-header names, so no
browser-driven cross-site request carries an attacker-chosen value, and a
client that deletes headers from its own request harms only itself. The
keep-set costs one line and closes the class before a later rung reads
one of those headers for something that matters.

**The fourth target, which is not a group.** Rung 5's own session cookie
is removed from any request before it is proxied to the streamer
(BS:441-443, `spec:808-813`). SEC:63 filed the salvage design's failure to
do this as a major, and **rung 5 ships the removal even though the proxy
that would need it arrives at rung 12, because the removal belongs to the
shell.** R-E below rules where in 5a it lands.

### C3. The normalization rules

Each is a **refusal with a status, and no partial routing**: the request
is answered and no route, no prefix decision, and no fallthrough sees it
(`spec:918-928`). **The status is `400 Bad Request` for every rule but the
length bound** *(ruled, conductor 2026-09-08, R-T, as applied)*, so a client
cannot tell which smuggling rule refused it. The length bound answers
`414` and is deliberately distinguishable, because an oversized path is
distinguishable from outside by its own size.

1. **A percent-encoded path.** The escaped form and the decoded form must
   be identical, which admits only a path carrying no percent-encoding at
   all. R-D rules the mechanism and states why the narrower reading was
   declined.
2. **Any `..` or `.` segment.** Dot-segment traversal, refused rather
   than cleaned. Cleaning is what `http.ServeMux` does, measured: it
   answers `/dubplate/./x` and `/dubplate/y/../x` with a 307 to
   `/dubplate/x`. The single dot is in the rule for the same reason as
   the double: the fallthrough hands `/rest/./x` to a streamer whose
   router may normalize differently.
3. **A doubled separator anywhere in the path.** The leading separator is
   not a segment, and **a single trailing separator is admitted**, so
   `/`, `/dubplate/`, and `/dubplate/x/` all pass and `//dubplate` and
   `/dubplate//x` are refused. Revision 1 stated this rule as "any empty
   segment", which refused the root path and the bare prefix with a
   separator, contradicting C4's own routing.
4. **Any backslash.** It matters on the fallthrough, where rung 12
   forwards a raw path to a streamer whose own router may read a
   backslash as a separator, **which is the smuggling shape one level
   down from the one C4 cites `mountBasePath` for**.
5. **A path past the stated byte bound**, answered `414`. R-H rules it as
   an open reading the spec does not carry.
6. **A request target whose path does not begin with a separator.**
   *(Ruled, conductor 2026-09-08, R-U, as applied.)* Measured on go 1.27.1, a
   `CONNECT` request arrives with an empty `URL.Path` and a populated
   `URL.Host`, so every other rule passes over it vacuously and it reaches
   the fallthrough. An absolute-form target is a different shape and is
   **admitted**: it arrives with its path intact and routes like any
   origin-form request, and refusing it would break an adopter's own
   proxy. Its residual is that `r.Host` is the authority the client chose,
   which the threat model records and 5b must not read as an origin.
   `OPTIONS *` never reaches this rule or any other: `net/http`'s server
   answers it from `globalOptionsHandler` unless
   `DisableGeneralOptionsHandler` is set, which 5a has no reason to set.

### C4. Prefix routing and the terminator

**dubplate's entire HTTP surface lives under exactly one configured
prefix, and the streamer owns everything else** (B3, `spec:884-904`). The
auth boundary is decided by configuration, never by a mux miss. MD's
routing rule 2, which hands `/api/*` to dubplate, was a blocker in both
adversarial reviews precisely because it collides with the streamer's own
API.

- **The match is segment-aware and case-sensitive.** A plain string trim
  is the defect class the salvage tree already proved one level down:
  `mountBasePath` documents that `http.StripPrefix` does a plain trim, so
  an unguarded `/uploadfoo` trims to `foo` and passes through. **A
  case-insensitive match would widen the owned surface past what an
  adopter configured**, so `/Prefix` reaches the fallthrough rather than
  the owned surface.
- **The terminator.** A path under the prefix that matches no registered
  route **returns 404 from dubplate and never reaches the fallthrough**
  (`spec:900-904`). Rung 5 builds it because rung 5 owns the mux, and
  rung 14's API lands under the same prefix and inherits it (RE cell
  (14,12)).

### C5. The extension point

The shape rung 5 ships is **a registration seam for the owned prefix plus
exactly one root-level fallthrough slot** (`spec:945-964`).

- **A consumer registers a subtree under the owned prefix by name at
  construction time**, in the binary's own wiring, and the shell owns the
  mux the registration lands in.
- **Two registrations of the same subtree are a programmer error and fail
  loudly at construction**, on the same precedent rung 1b's registry
  cites for its collision gate: `flag.(*FlagSet).Var`,
  `database/sql.Register`, and `expvar.Publish` all panic on a collision
  because the caller is a programmer who mis-declared, never an adopter
  who mis-configured (`internal/config/registry.go:23-30`).
- **The fallthrough slot takes at most one handler**, and a deployment
  with no streamer leaves it empty, **which is what makes rung 5 safe to
  ship without a driver**.
- **The static surface is two handlers rather than one**, because DL R4.5
  splits a document renderer for `index.html` from an immutable asset
  server for everything hashed. **5a builds neither**; it builds a seam
  that admits both, so rung 7 can swap the embedded bundle for a larger
  one without touching either.

### C6. The deadline policy

The salvage server sets four timeouts on a single `http.Server`, a
10-second read-header, a 30-second read, a 30-second write, and a
120-second idle (`_salvage/internal/serve/serve.go:599-602`, applied at
`:630-633`). **The write value is the one B12 answers**: 30 seconds kills
an SSE stream and any audio response longer than that, which is why B12
asks for per-route handling. **The read timeout is adopted rather than
rejected**, which revision 1 dropped by describing the salvage server as
carrying the write timeout alone. **Rung 5 sets the policy
even though rung 12 owns the routes that need it, because the policy is a
property of the `http.Server` the shell constructs and rung 12 cannot
change it without writing the shell** (`spec:965-981`).

- The mechanism is **`http.ResponseController.SetWriteDeadline`**, which
  lets a streaming handler extend or clear its own deadline while the
  server's default protects everything else.
- Rung 5 sets **a read-header timeout, a read timeout, an idle timeout,
  and a default write timeout**, and **exposes the extension mechanism to
  the fallthrough handler**.
- **The read timeout is what bounds a slow request body.**
  `ReadHeaderTimeout` stops covering the connection once the headers are
  read, and `IdleTimeout` applies between requests, so without it a body
  trickled a byte at a time holds a connection on a route that streams
  nothing. 5b's login and consumption POSTs take bodies. *(Ruled,
  conductor 2026-09-08, R-V, as applied: adopted at the salvage server's
  30 seconds. If a later rung wants a streaming upload, the 5b handoff
  carries the alternative,
  `http.MaxBytesReader` on the body plus
  `ResponseController.SetReadDeadline` per route.)*
- **Rung 5 has no streaming route of its own**, so the plan verifies the
  mechanism against a test handler that writes past the default rather
  than against the constant.

### C7. Two addresses, two jobs

R27 (`spec:595-607`, `spec:823-848`) is the spine of every trust decision
in the shell, which is why it is C7 rather than a paragraph inside the
limiter section it will eventually serve.

- **The peer address is `RemoteAddr`.** It is **the only input to any
  trust decision**. The trusted-proxy allowlist compares against it, the
  verify endpoint's permitted-addresses list compares against it, and **no
  configured header can change it**.
- **The reported client IP is derived, and only for rate limiting and
  logging.** When trusted-proxy mode is off, it is the peer address. When
  it is on, the shell reads the configured client-address header **if and
  only if the peer address matches one of the configured proxy
  addresses**, and otherwise falls back to the peer address.
- **A multi-hop value parses right to left**, taking the rightmost hop
  that is not itself a configured proxy address, **so an attacker cannot
  reach past the trusted chain by prepending entries**. Revision 1 stated
  the property as "an attacker prepending entries cannot choose the
  result", which its own acceptance fixture contradicts: with two
  configured proxies to the right, the answer is the leftmost entry, and
  that entry is the attacker's. **The residual ships with the property**:
  the derived value is attacker-controlled whenever the attacker is the
  nearest non-proxy hop, and the guarantee rests on every configured
  proxy appending the address it observed rather than forwarding what it
  was handed.

The distinction is what makes three guarantees true at once: a limit key
an attacker cannot pick, an audit trail that records who actually
connected, and a verify endpoint an internet client cannot reach by
naming the proxy's address in a header.

### C8. The bind surface

- **The bind default is loopback, and the reason ships with it**
  (`spec:1568-1578`). The login surface is anonymous-reachable and the box
  is production, so SEC's Access-retirement minor asks for the bridge
  gateway and loopback and **never `0.0.0.0`**. An adopter reaches the
  service through the tunnel or through their own proxy.
- **`IP_FREEBIND` lets the listener bind an address that is not yet
  present** (R20, `spec:542-549`). The salvage server does this for boot
  resilience when the configured address is not yet on any interface.
  Adopting the mechanism **keeps the unit off `network-online.target` and
  removes a boot-order failure mode**, and it costs one `Control`
  function. The address rung 12 will bind is not known at rung 5, which
  is the case the mechanism protects.
- **The residual ships with it.** `IP_FREEBIND` **suppresses the
  boot-time error a misconfigured address would otherwise raise**, so
  `doctor` carries a row reporting the addresses actually bound and the
  misconfiguration surfaces there instead. **That row is 5d's** (R-G).

## Resolutions this plan takes rather than discovers

Twenty-two, R-A through R-V, each stating what would change under the
other answer. A task builds to the resolution as written and reports if
the tree contradicts it.

**The conductor ruled all twenty-two at dispatch, 2026-09-08, each
standing as applied**, on 8a's precedent rather than at the close.
Revision 1 left them provisional through execution while asking the
close to list them "with their markers", and no task was told to write
a marker. R-A and R-B compile into a registry
declaration whose `validateThresholdDefault` refuses an empty
`KindString` default (`registry.go:320-322`), so a ruling that arrives
after Task 7 re-edits a landed contract file and its test literals.

**Each task stamps the marker on the symbol its resolution decides**, in
the repo's own convention, `// ruled, conductor 2026-09-08, R-x`, which
the tree already carries at `internal/config/registry.go:33`,
`keys.go:35`, `cmd/dubplate-server/main.go:62`, `shadow.go:27`, and
`doctor_test.go:150`. Task 8's table is then a search rather than a
reconstruction.

**R-P through R-V are revision 2's**, folded from the review. R-P clears
trailer fields, R-Q collapses a duplicate exempted header, R-R extends
group 2 to the `X-Forwarded-*` family, R-S bounds group 3 with a
keep-set, R-T names `400` as the shared refusal status, R-U refuses a
non-origin-form request target, and R-V adopts a read timeout. Each is
stated where the contract it changes is stated, in C2, C3, or C6, rather
than repeated here.

### R-A: the owned prefix defaults to `/dubplate`

`spec:1564` declares `DUBPLATE_SERVE_PREFIX` as "a threshold with a
default" and names no value. The registry refuses an empty `KindString`
default (`registry.go:320-322`), so a value is required at declaration
rather than deferrable.

**Ruled, conductor 2026-09-08, R-A, as applied: `/dubplate`.** Three constraints pick it. It must not collide
with the streamer's own surface, and Navidrome's is `/rest`, `/share`,
`/app`, and `/auth`. It must be one segment, so the segment-aware match
has a boundary to find. And it must carry no estate value, which the
product's own name satisfies for every adopter. **What would change under
another answer: one constant in the config declaration and the default
row in Task 7's table.** No shell code reads the value.

### R-B: the bind and port defaults, and the list kind

`spec:1562-1563` gives both keys defaults and names neither.

**Ruled, conductor 2026-09-08, R-B, as applied: `DUBPLATE_SERVE_BIND_ADDRS`
defaults to `127.0.0.1` and `DUBPLATE_SERVE_PORT` to `8080`.** The salvage configuration proves both
shapes: it read the bind list as comma-separated
(`env/dubplate.env.template:104-113`) and its tests carry `8080`
(`_salvage/internal/config/config_test.go:44-45`). Loopback is the
default `spec:1568-1571` argues for by name.

**The list has no registry kind, and 5a does not add one.** The registry's
four kinds are `KindInt`, `KindDuration`, `KindFloat`, and `KindString`
(`registry.go:296-327`). **A fifth kind is a contract change to rung 1's
config core**, which is 5b's contested surface and not 5a's. So
`DUBPLATE_SERVE_BIND_ADDRS` registers as `KindString` and
**`internal/serve` splits and validates it**, which also keeps the
address grammar in the package that binds. **What would change under
another answer: a `KindList` in `registry.go`, a `validateThresholdDefault`
case, and a doctor row shape, all of which is a rung 1b amendment this
pass has no mandate for.**

### R-C: the serve keys are declared in `cmd/dubplate-server/serve.go`

`spec:982-1005` names the wiring point and rung 4's precedent:
`registerShadowIntegration` and `registerShadowThresholds` live in
`cmd/dubplate-server/shadow.go` and are called above the `AddCommand`
block. **5a follows it exactly** (ruled, conductor 2026-09-08, R-C, as
applied): `registerServeConfig(reg)` and
`newServeCmd(...)` live in a new `cmd/dubplate-server/serve.go`.

**The alternative, declaring inside `internal/serve`, is declined**,
because it would make the shell import `internal/config`, which
`spec:1474-1479` gives it no reason to do and which R-M keeps out. **What
would change under the other answer: a `config` import in
`internal/serve` and a registration call that no longer sits beside its
sibling.**

**The reserved-class removal lands in the same commit as the
declaration**, which is the registry contract in the words rung 1b wrote
(`HISTORY:493-498`) and the first exercise of its removal half.

### R-D: the raw-versus-decoded comparison reads `EscapedPath` against `Path`

`spec:918-920` states the rule and not the mechanism, and the mechanism
has a trap. `net/url` populates `URL.RawPath` **only when the escaped form
differs from the default encoding of `URL.Path`**, so a request whose path
needed no escaping carries an empty `RawPath` and a comparison against it
would pass every unencoded path by accident and fail every ordinary one.

**Proposed: refuse when `r.URL.EscapedPath() != r.URL.Path`, a single
predicate and no second sentence.** `EscapedPath` returns `RawPath` when
it is a valid encoding of `Path` and otherwise re-derives the escaped
form, so it is defined for every request.

**Revision 1 stated the rule twice and one statement was vacuous**, which
the review filed as B2. The second statement, "a single decode of the
escaped form does not reproduce the decoded form byte for byte", refuses
nothing: measured on go 1.27.1 over thirteen paths,
`PathUnescape(EscapedPath()) == Path` holds for every one, `/a%2Fb`,
`/a%5Cb`, `/a%2E%2E/b`, and `/a%2520b` included, because `EscapedPath`
returns `RawPath` only when unescaping it reproduces `Path`. That
sentence is deleted here and from Task 3.

**The surviving predicate is a blanket ban on percent-encoding in a
path, and this plan takes it as the policy rather than as a side
effect.** *(Ruled, conductor 2026-09-08, R-D, as applied.)* Measured,
`EscapedPath() != Path` is true for `/a%20b`, `/%E2%98%83`, `/a%41b`,
`/%00`, and `/%0A` alongside every encoded separator, so a legitimately
escaped path is refused too. The reason to keep it: the shell owns no
path that needs an encoded byte, rung 12's driver declares its bypass
paths as literals, and a rule that admits some encodings has to decide
which, at the one place in the shell where a wrong decision is invisible.
It is also the only rule that reaches `%00` and `%0A`. **The strictness is
named rather than absorbed as a permanent default**: rung 12 revisits it
against the streamer's real client paths, share links, cover-art paths,
and any path a Subsonic client percent-encodes, with measured evidence
before relaxing it.

**The narrower mechanism was considered and declined.** `RawPath != ""`
is the idiomatic non-default-encoding test, and it is narrower than
either reading in a way the plan cannot use: measured, `/a%2Fb`,
`/a%2E%2E/b`, and `/a%41b` populate `RawPath` while `/a%5Cb` and `/a%20b`
do not, because the default encoder escapes a backslash and a space
itself. A rule keyed on it would admit `/a%20b` and refuse `/a%41b`,
which is a distinction no reader of the code could justify.

**This is the one mechanism in the pass the spec does not specify, and it
is why Task 3 carries the pass's `opus` upshift.** **What would change
under the other answer: rule 1 admits a path whose encoded bytes decode
to no separator, dot, or control character; Task 3's table gains the
admitted rows; and the forward criterion's sentence to rung 12 inverts.**

### R-E: the session-cookie removal lands at 5a, parameterized by the cookie name

`spec:808-813` says rung 5 ships the removal and gives no pass. **No pass
row in `spec:2035-2041` claims it**: 5a's row names the strip's three
groups, 5b's names the cookie's construction, and neither says who
removes it on the way out.

**Ruled, conductor 2026-09-08, R-E, as applied: 5a builds it, at the
fallthrough boundary, with the cookie name supplied at construction.**
The reason is the spec's own: **the
removal belongs to the shell**, and the shell is 5a's whole deliverable.
The cookie name does not exist until 5b, so the shell takes it as a
construction value and 5a's test supplies an arbitrary one. A shell
constructed with no cookie name removes nothing, which is the 5a wiring.

**What would change under the other answer: the removal moves to 5b,
where it becomes an edit to a file 5b otherwise only reads, and 5a ships a
fallthrough slot whose contract is silent on the one header the birth
requirement names.** That is the shape SEC:63 filed as a major.

### R-F: the origin check is 5b's, and 5a documents its insertion point

R15 (`spec:479-511`) makes the origin check a shell rule covering the
fallthrough, and `spec:2038` puts it in 5b's scope line. **5a builds
none** (ruled, conductor 2026-09-08, R-F, as applied). It records the
constraint the insertion must satisfy, so 5b adds a
wrapper rather than reordering the chain: **the check runs after the strip
and after normalization**, so a smuggled path never reaches it and no
inbound identity header is present when it decides.

**What would change under the other answer: 5a would need the session
cookie's name and the bearer-bypass pattern set, neither of which exists
until 5b.**

**Two facts travel with the insertion point, so 5b inherits them rather
than measuring them.** Go's `CrossOriginProtection.Check` returns nil when
neither `Sec-Fetch-Site` nor `Origin` is present, which is a documented
fail-open on non-browser requests and is not something the shell can
change. The shell keeps both headers out of group 3's listed removal
(R-S), so the check reads what the client actually sent.

**R33 is disposed of here, because it reads adjacent to Task 3 and is
not Task 3's.** `spec:674-686` governs the post-login destination: rung 5
accepts no caller-supplied destination, and the single-decode predicate
it states applies to that value. **It imposes nothing on 5a's path
rules**, which refuse rather than accept and which run on every request
rather than on one form field. 5b builds it.

### R-G: `doctor`'s client-IP-mode row is 5d's, and 5a exports what it reads

INV's surprise 2 is that **`doctor` is a reporter and the birth
requirements need a prover** (`INV:1194-1199`): it walks registry families
and prints `%-32s %s` lines, with no vocabulary for a check whose result
depends on evidence. R6 (`spec:348-378`) answers with a probe row kind,
and `spec:2041` puts the probe rows and the bound-address row in **5d's**
scope.

**Ruled, conductor 2026-09-08, R-G, as applied: 5a defers the reporting
and ships what the report reads.** The shell exposes the live client-IP mode as a value a caller can render,
and it exposes the addresses actually bound after `Listen` returns. **5d
renders both rather than re-deriving either**, which is what keeps B8's
"`doctor` reports which is live" honest: a `doctor` that recomputed the
mode from configuration would report what was configured, not what runs.

**What the descriptor proves, stated exactly so 5d does not over-read
it.** The mode descriptor is the shell's own constructed state, built
from the same arguments the mode is configured with. It proves that
configuration reached the shell, which is one hop better than re-reading
the environment and is **not** evidence that any request was classified.
The request-path evidence is 5d's probe row. The bound-address value is
different in kind: it comes back from `Listen` and is a measurement.

**What would change under the other answer: 5a would add the probe row
kind to `doctor`, which collides with two landed assertions**
(`doctor_test.go:129-131` and `TestDoctorCmd_LineCountFollowsRegistry`)
that R6 explicitly assigns to the pass that adds the rows, and which
`spec:369-378` says the plan must not discover at implementation time.
5d owns that collision.

### R-H: a path-length bound and an explicit `MaxHeaderBytes`

The threat model's reviewer rubric includes oversized paths
(`spec:1903-1975` read against the brief's attack list), and the spec
carries no bound. Go's default `MaxHeaderBytes` is 1 MB and applies to
the request line and headers together.

**Ruled, conductor 2026-09-08, R-H, as applied: `MaxHeaderBytes` is set
explicitly at the value the stdlib defaults to, so the number is a
stated decision rather than an inherited one, and normalization refuses
a path longer than a stated byte bound as rule 5.** The refusal is
`414 URI Too Long`, which is the status the
shape has. **The bound is a constant, not a threshold**, because
`spec:2380-2383`'s tunable list does not name it and a threshold nobody
tunes is a config surface for nothing.

**What would change under the other answer: C3 loses rule 5 and an
oversized path reaches the prefix match, where a very long segment costs
a comparison and nothing else.** The bound is cheap
insurance rather than a proven need, which is why it is stated as a
resolution rather than assumed.

### R-I: the trusted-proxy configuration is 5d's, and 5a's derivation takes values

R22 (`spec:556-562`) makes the trusted-proxy client-IP mode an
**integration class** with an enable flag and required-when-enabled keys,
and `spec:2041` puts that class in 5d's scope.

**Ruled, conductor 2026-09-08, R-I, as applied: 5a builds the
derivation and declares no key for it.** The shell's constructor takes
the proxy address set and the client-address
header name as plain values, both empty by default, and empty means mode
off, so `RemoteAddr` is the client address until an adopter says
otherwise (`spec:1607-1611`). **5d declares the class and passes the
resolved values in.**

**What would change under the other answer: 5a declares an integration
class whose enable flag nothing reads, which is exactly the defect risk
11 names** (`spec:2281-2286`): a key the registry claims and nothing
requires reads as done in every review and reports green in `doctor`.

### R-J: a fresh-context `web-auth-security-reviewer` runs at 5a's pass end

R24 (`spec:571-580`) names the reviewer inside 5b and again at 5c, and
names none at 5a.

**Ruled, conductor 2026-09-08, R-J, as applied: one runs here too, as a
conductor dispatch after the close task's gate and before the merge
ritual.** The argument is the spec's own
split-point paragraph (`spec:2043-2047`): the shell's composition order is
the artifact the whole trust boundary rests on, and the pass boundary
exists so that artifact gets read on its own rather than buried under
token mechanics. **A 5a with no security read defeats the reason 5a is a
separate pass.** The reviewer's rubric is the threat-model section below,
and it grades the implementation against the claims rather than against
taste (`spec:2340-2344`).

**What would change under the other answer: the read moves into 5b, where
it competes with the token, the session, the nonce store, the roster, and
the limiters for the same reviewer's attention.** The close drafts the
one-line spec correction to R24 either way, so the review shape records
what happened.

### R-K: 5a writes no systemd unit and no deploy path

`spec:2035`'s exit gate names no unit, and `spec:2041` gives 5d the serve
unit, its `ReadWritePaths` grant over both leaves, the two `deploy.sh`
pre-creates, and the generalized state-directory check. **INV section 2.6
(`INV:342-386`) is read as 5d's input rather than 5a's**, and this plan
carries no `vps-conventions` content for that reason. (Ruled, conductor
2026-09-08, R-K, as applied.)

**One consequence is stated so 5d inherits it rather than discovering
it**: `IP_FREEBIND` is what keeps the unit off `network-online.target`
(`INV:378-386`), so 5d's unit writes `After=` on that argument and cites
this pass's Task 6.

### R-L: the extension point registers by subtree name and takes `http.Handler`

`spec:949-957` states the seam's contract as behavior and names no
signature. **Ruled, conductor 2026-09-08, R-L, as applied: one method
on the shell's construction value taking a subtree name and an
`http.Handler`, called from the binary's own wiring before the shell is
built.** No new interface, because
`http.Handler` is the stdlib's own and the extension point's whole value
is that a consumer brings one.

**Two registrations of the same subtree panic at construction** (C5),
which is the rung 1b precedent and not a new rule. **What would change
under another answer: a variadic option list or a map passed to the
constructor, both of which lose the collision's name in the panic
message.**

### R-M: `IP_FREEBIND` through `syscall`, not `golang.org/x/sys/unix`

The salvage code sets the option through `golang.org/x/sys/unix`
(`_salvage/internal/serve/serve.go:570`). **R26 budgets rung 5 exactly two
module dependencies and neither is `x/sys`** (`spec:585-594`).

**Ruled, conductor 2026-09-08, R-M, as applied: `syscall.SetsockoptInt`
inside a `net.ListenConfig.Control` hook.** The constant `IP_FREEBIND`
is 15 on Linux and `syscall` carries
`IPPROTO_IP`. **The task verifies the constant against the kernel header
rather than carrying it from memory**, and names its source in the code
comment. **What would change under the other answer: `go.mod` gains
`golang.org/x/sys`, R26's count moves from two to three, and the spec's
prior-art table needs a row it does not have.**

### R-N: an empty fallthrough answers 404

`spec:930-937` says the fallthrough is empty at rung 5 and does not say
what an empty one returns.

**Ruled, conductor 2026-09-08, R-N, as applied: 404, the same status
the prefix terminator returns, with an empty body.** With no streamer
configured nothing owns that surface, and
404 is what "nothing is there" means. **501 is declined**: it advertises a
capability the deployment lacks, which tells an unauthenticated visitor
something about the deployment's shape for no gain.

**A consequence worth stating: the terminator and the empty fallthrough
are indistinguishable from outside**, which is a property rather than a
defect. The distinction lives in the log record. **What would change under
another answer: one constant and two test rows.**

### R-O: the fixed hop-by-hop names stay rung 12's

`spec:796-806` requires removal of every header the request's own
`Connection` header lists and says nothing about the fixed RFC 7230 set
(`Keep-Alive`, `Transfer-Encoding`, `TE`, `Trailer`, `Upgrade`,
`Proxy-Authorization`, `Proxy-Authenticate`).

**Ruled, conductor 2026-09-08, R-O, as applied: 5a removes what
`Connection` lists, with two carve-outs, and removes no fixed name of
its own.** `httputil.ReverseProxy` removes the
fixed set itself on the way out (`reverseproxy.go:469`), so a shell that
also removed them would duplicate work the stdlib already does correctly.
**The `Connection`-listed removal is the half the stdlib does not do at
ingress**, and it is the half that carries the attack: a client sending
`Connection: Remote-User` induces the downgrade.

**Carve-out 1, the upgrade case.** *(Ruled, conductor 2026-09-08, R-O,
as applied.)* Revision 1 declined the fixed set on the ground that removing it
would strip `Upgrade` from a websocket request, then adopted a rule that
strips `Upgrade` from exactly that request: a handshake sends
`Connection: Upgrade` alongside `Upgrade: websocket`, so the listed
removal deletes it. The review filed that as M-a. **A request whose
`Connection` names `Upgrade` keeps both `Upgrade` and `Connection`**, and
every other listed name is still removed. The precedent is the stdlib's
own: `ReverseProxy` reads `upgradeType` at `:464` before
`removeHopByHopHeaders` at `:469`, then re-sets `Connection: Upgrade` and
`Upgrade: <type>` at `:481-484`, under the comment "add back any
necessary for protocol upgrades, such as for websockets".

**Carve-out 2, the keep-set** (C2, R-S): `Cookie`, `Origin`,
`Sec-Fetch-Site`, and the session-cookie name are never removed by the
listed rule.

**What would change under the other answer: rung 12 inherits a shell that
has already stripped `Upgrade`, and the defect surfaces as a streaming
capability that silently does not work.**

## The threat model, as the pass-end reviewer's rubric

`spec:1903-1975` is the rung's threat model, and **the reviewer grades
against it: a finding that contradicts it is a finding against the spec
rather than against this implementation**. This section names the subset
5a is answerable for and the attacks the reviewer must actually try.

**What 5a protects.** The integrity of every trust decision made
downstream of the shell. 5a holds no secret and no roster; it holds the
guarantee that what reaches a handler is what a client actually sent,
minus everything a client must not be able to assert.

**Who the adversary is.** An unauthenticated visitor reaching serve from
the internet, since the login surface is public by construction. A
misconfigured or hostile proxy in front of serve. A mail gateway that
follows links is 5b's adversary and not 5a's.

**Entry points at 5a.** Every path, since the shell answers all of them:
the owned prefix with no routes registered under it yet, and the
fallthrough, which is empty.

**The claims 5a makes**, drawn from the spec's own claim list:

1. **No inbound identity field survives ingress on any path**, in either
   spelling, **in the header section or in a trailer**, including on a
   path that would later bypass authentication.
2. **A path cannot be smuggled through the prefix.** Percent-encoding, a
   `..` or `.` segment, a doubled separator, a backslash, an oversized
   path, and a target whose path does not begin with a separator are each
   refused before routing.
3. **No configured header can change a trust decision**, because every
   trust decision reads the peer address.
4. **An unknown path under the owned prefix returns 404 from dubplate and
   never reaches the fallthrough.**
5. **A response that outlives the default write deadline succeeds only
   when its handler extended the deadline**, and the default protects
   everything else.

**The attacks the reviewer must try**, named so the read is a search
rather than an impression:

- **Header spoofing before and after the strip.** Every spelling in the
  shipped set, in both dash and underscore form, plus a name the set does
  not carry, sent to a path under the prefix and to the fallthrough.
- **Trailer-borne identity.** A chunked request announcing a strip-set
  name in `Trailer` and sending it after the body, read at the handler
  both before and after the body is consumed.
- **Header deletion by the client.** `Connection` listing an identity
  header, a forwarded header, `Cookie`, `Origin`, `Sec-Fetch-Site`, and
  the session-cookie name, plus a websocket handshake's
  `Connection: Upgrade`.
- **The forwarded family.** `X-Forwarded-Host`, `X-Forwarded-Proto`, and
  `X-Forwarded-Prefix` from a non-allowlisted peer, and the same three
  from an allowlisted one.
- **Encoded-slash bypasses.** `%2F` and `%5C` in a segment, double-encoded
  forms, and a path whose escaped and decoded forms differ only in a
  segment the prefix match would read.
- **Dot-segment bypasses.** `..` in every position, encoded and plain,
  and a path that cleans to a legitimate one.
- **Case games.** The owned prefix in a different case, mixed case, and a
  case-varying path that string-prefixes the owned prefix without a
  segment boundary.
- **Oversized paths.** A path past the stated bound, and a request line
  past `MaxHeaderBytes`.
- **Slowloris against the timeouts.** A connection that sends headers a
  byte at a time, and one that opens and sends nothing, against the
  read-header and idle timeouts.
- **Host header and absolute-URI requests.** An absolute-form request
  target, a `Host` disagreeing with the request target's authority, and a
  request whose authority carries the owned prefix.
- **Method games on the fallthrough.** Every method against the empty
  fallthrough and against the terminator, `CONNECT` and `OPTIONS`
  included, plus a request with a body on a method that should carry
  none.
- **A slow body on a non-streaming route.** A POST that sends its headers
  promptly and then trickles its body, against the read timeout, beside
  the slowloris cases above.
- **Percent-encoded paths that should pass.** A benign encoded byte in a
  segment, to confirm rule 1's refusal policy is the one the plan states
  rather than an accident.

**The residuals 5a states rather than hides.** The forward-auth topology's
guarantees are inert here, because in that deployment serve is not in the
streamer's request path and the boundary is the adopter's proxy
(`spec:863-883`); 5a builds no verify endpoint and 5d writes that
documentation. `IP_FREEBIND` suppresses the boot-time error a misconfigured
bind address would raise (C8), and the row that surfaces it is 5d's. The
strip's extras list is only as complete as the documentation each name was
verified against, which is why Task 2 requires a citation per name and
reports any name it dropped. **`OPTIONS *` never reaches the chain**, so
the strip does not run on it: `net/http`'s server answers it from
`globalOptionsHandler` (`server.go:3409-3410`) unless
`DisableGeneralOptionsHandler` is set, and that handler writes only a
zero-length response. **An absolute-form target sets `r.Host` to the
authority the client chose** (C3 rule 6), so nothing downstream may derive
an origin, a base URL, or a scheme from it. **A derived client IP is
attacker-controlled whenever the attacker is the nearest non-proxy hop**
(C7), which bounds what the limiter key at 5b can promise.

**The reviewer receives** this section, the spec's own threat-model
section, and the diff. It does **not** receive this plan's task
decomposition, so its read is of the shipped shell rather than of the
plan's intent.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 5 spec and this plan
against the tree as shipped. **Its output is blockers with evidence, not
fixes**; it writes no code and edits no file. The conductor rules each
blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's `spec:`, `INV:`, `PA:`, `BS:`, and `PRA:` pointers
were resolved at `bc4016b` and re-resolved during the revision 2 fold,
which corrected fourteen of them. A citation that has moved since is a
finding for this task rather than a surprise for an implementer.

**Grade a moved pointer by whether its content moved.** A pointer off by
fewer than about ten lines whose cited content still resolves to the same
subject is a note in the report. A pointer landing on a different subject,
or naming content that no longer exists, is a blocker. Revision 1 made
every drift a blocker, which would have spent the task's whole report on
the plan's own arithmetic.

The checklist is the spec's own (`spec:2153-2218`), reduced to the
questions 5a's surface turns on and extended with the two this pass adds.
**The first three are blockers: a "no" stops the launch.**

1. **Is the lane registry consistent with where this pass intends to run,
   and is the target tree free?** `docs/STATUS.md`'s lane registry is the
   live-executor check, **never `pgrep -f Projects/dubplate`, which
   matches every lane once a worktree exists**. *(Expected at drafting:
   lane-1 IDLE then active on pass 8b in the main tree; lane-2 IDLE,
   about to take this pass in a worktree.)* **Report any live executor in
   the target worktree, any uncommitted change under `internal/serve` or
   `cmd/dubplate-server`, and confirm the worktree is at master's HEAD.**
2. **Has any other lane written `internal/serve`?** PA:80 forbids it until
   rung 5 closes, **and a violation invalidates 5a's whole design**
   (`spec:2164-2166`). *(Expected: the directory does not exist;
   `internal/` holds `atomicfile`, `config`, `engine`, and `harness`
   only.)* **Report the directory listing and any commit touching that
   path.**
3. **Has lane 1's pass 8b touched 5a's territory?** *(Expected: no. 8b
   writes `internal/engine/tagio` and its testdata.)* **Report any 8b
   commit touching `internal/config`, `cmd/dubplate-server`, `go.mod`, or
   `.gitleaksignore`**, since those are the four surfaces where the two
   lanes could collide and where the merge ritual's rebase would stop
   being a fast-forward.
4. **Does `internal/config/keys.go` still carry the `serve` reserved
   class at the shape this spec cites?** *(Expected: `:74-85`, name
   "serve", rung 5, two keys `DUBPLATE_SERVE_BIND_ADDRS` and
   `DUBPLATE_SERVE_PORT`, with the code comment arguing that a bind
   address is not an identity concern.)* **Task 7 removes it. Report the
   exact line range and the comment's presence**, because the comment's
   argument is what the new declaration's doc comment carries forward.
   **Also confirm the `auth and identity` class and the `{Name: "identity
   source", Pending: ...}` wall row are still present and untouched**;
   5a removes neither, and their absence means another lane has taken
   5b's work.
5. **Does the registry still expose `AddWallRow`, `AddReservedClass`,
   `AddIntegrationClass`, `AddThreshold`, and `KindString`?** *(Expected:
   `registry.go:138`, `:163`, `:211`, `:267`, and the kind constants
   around `:296-327`.)* **Report `validateThresholdDefault`'s
   `KindString` case**, which refuses an empty default, because R-A and
   R-B both depend on it.
6. **Does `AddThreshold` still skip the claim for a key that is also an
   integration class's `Optional` member?** *(Expected:
   `isOptionalIntegrationKey` at `:284`, called from `:273`.)* 5a
   registers no integration class, so the pairing does not fire here.
   **Report the shape anyway**, because 5d's trusted-proxy class is the
   first place outside rung 4's shadow class where it will.
7. **Is `cmd/dubplate-server/main.go`'s single `AddCommand` block still
   the wiring point?** *(Expected: `newRootCmd` at `:63`, `reg :=
   config.NewRegistry()` at `:64`, two `registerShadow*` calls at
   `:65-66`, and one `AddCommand` block at `:74-79` carrying four
   subcommands.)* **Report the doc comment's subcommand list**, which
   `spec:988-990` says already names "serve: run the HTTP server (rung
   5)". Task 7 adds to this block, and the merge ritual resolves any
   conflict with lane 1.
8. **Does `doctor` still derive every row from the registry?**
   *(Expected: `doctor.go:29-66`, four family loops, no per-rung
   knowledge.)* **5a adds no `doctor` code**, so this is a confirmation.
   **Report the two line-count assertions by name and line**
   (`doctor_test.go:129-131` and `TestDoctorCmd_LineCountFollowsRegistry`
   at `:151`), and **compute what 5a's declarations do to the count**:
   one reserved-class row removed, three threshold rows added, net plus
   two. If either assertion would fail on that arithmetic, **that is a
   blocker for Task 7 and the conductor rules it before dispatch.**
9. **Does `internal/config/config.go`'s `Load` still hard-code its four
   owned requires?** *(Expected: `:241-279`, four `w.require` calls plus
   the generic integration loop.)* **5a does not edit it**; R29 is 5b's.
   Report the shape so 5b inherits a verified reading.
10. **Does `_salvage/internal/serve/serve.go` still carry the four things
    this plan reasons from?** *(Expected: `mountBasePath` at `:82-107`
    with its plain-trim comment, `securityHeaders` at `:110-128`,
    `freebindControl` and `freebindListeners` at `:561-596`, and the
    timeout constants at `:598-612` applied to one `http.Server` at
    `:625-633`.)* **Report the four line ranges and the `x/sys/unix`
    import**, which R-M declines.
11. **Is `_salvage/internal/serve/serve.go`'s `requireSameOrigin` still
    present?** *(Expected: at `:152`.)* **5a builds no origin check**
    (R-F). Report its shape for 5b.
12. **Does `go.mod` still carry cobra and its two indirects only?**
    *(Expected: `github.com/spf13/cobra v1.10.2`, plus mousetrap and
    pflag indirect; `go 1.27.0`.)* **5a adds none. Report the Go version
    line**, because `http.ResponseController` needs 1.20 and
    `net/http.CrossOriginProtection` needs 1.25, and 5b depends on the
    second.
13. **Is `.gitleaksignore` still the tracked empty file rung 1b left**,
    and has any lane appended to it? **Report the line count without
    quoting any fixture literal**, because gitleaks scans the tracked
    worktree with `--no-git` and would flag the quotation.
14. **Has `scripts/check.sh` gained or lost a leg?** *(Expected: the
    module gate, shellcheck, ruff, gitleaks, vale, `docker compose
    config`, and the pattern-filtered `systemd-analyze verify`. The web
    leg is still absent; its rejoin is 5c's.)* **Report the leg list and
    the "rejoins at rung 7" sentence's line**, which 5c corrects, and
    **do not change it here**.
15. **Is the state-directory agreement check still hard-coded to
    `systemd/dubplate-shadow.service`?** *(Expected: `check.sh:153-168`.)*
    R12 generalizes it **at 5d**. Report the shape so 5d inherits it, and
    **confirm 5a's absence of a unit means the check's arithmetic does
    not move.**
16. **Does `systemd/dubplate-shadow.service` still carry the hardening
    template and the leaf `ReadWritePaths` grant 4b landed?** *(Expected:
    the hardening block carried forward from the incumbent minus
    `AF_NETLINK`, and one leaf grant.)* **5a writes no unit** (R-K).
    Report it as 5d's input.
17. **Verify `IP_FREEBIND`'s value, its `syscall` reachability, and the
    boot property itself on this toolchain.** R-M turns on the first two
    and Task 6's mutation 5 turns on the third. *(Measured during the
    revision 2 fold on go 1.27.1: `syscall.IP_FREEBIND = 0xf` at
    `$(go env GOROOT)/src/syscall/zerrors_linux_amd64.go:433`;
    `syscall.SetsockoptInt` and `GetsockoptInt` at `IPPROTO_IP` both
    succeed inside a `net.ListenConfig.Control` hook, on `AF_INET` and
    `AF_INET6` alike; and an unprivileged listen on `192.0.2.77:0` fails
    `bind: cannot assign requested address` without the hook and succeeds
    with it, that address being on no interface here.)* **Confirm all
    three against the worktree's own toolchain and report any
    divergence**, since Task 6's behavioral row rests on the third.
18. **Confirm `http.ResponseController` and `SetWriteDeadline` are
    present on the installed toolchain**, measured rather than
    remembered, alongside `net.ListenConfig.Control`. R15's own precedent
    is that the spec measures a stdlib surface's presence rather than
    assuming it (`spec:501-504`).
19. **Is `go test ./...`'s wall time still what this plan sizes gate runs
    against?** Run it once at `-count=1` in the worktree and report the
    total and `bash scripts/check.sh`'s total. *(4a measured `go test
    ./...` at about 2m06s and the gate at about 3.5 minutes.)*

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 19's timing measurement; the task writes nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

**Prior art:** none consulted. The task reads the tree.

**Independent: no.** Every task depends on it.

---

## Task 1: The two addresses

Model: `sonnet`. Depends on Task 0.

R27 is the spine of every trust decision the shell makes, so it is born
first and nothing downstream has a reason to invent a second reading of
`RemoteAddr`.

**Files:**
- Create: `internal/serve/clientaddr.go`
- Create: `internal/serve/clientaddr_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: a value carrying the trusted-proxy configuration (the proxy
  address set and the client-address header name, both empty for mode
  off), a function returning the **peer address** from an
  `*http.Request`, a function returning the **reported client IP**, and a
  value describing the **live mode** in a form 5d can render without
  re-deriving it (R-G). Tasks 2 and 5 consume the first three; 5d consumes
  the fourth.

**Outcomes (C7):**

- The peer address is `RemoteAddr`, port stripped, and nothing else can
  produce it. No configured header, no argument, and no mode changes
  what it returns. It is the only input to any trust decision in the
  shell.
- The reported client IP equals the peer address when trusted-proxy
  mode is off, which is the default, so `RemoteAddr` is the client
  address until an adopter says otherwise (`spec:1607-1611`).
- With mode on, the configured client-address header is read if and
  only if the peer address matches a configured proxy address, and
  otherwise the peer address is the answer. A request arriving from a
  non-allowlisted peer that carries the header gets the peer address, and
  the header's value never reaches a caller.
- A multi-hop value parses right to left, taking the rightmost hop
  that is not itself a configured proxy address. **An attacker cannot
  reach past the trusted chain by prepending entries**, which is the
  property the parse direction exists for, and it is narrower than
  "cannot choose the result": when the attacker is the nearest non-proxy
  hop, the answer is the attacker's own value. The doc comment states
  both halves.
- A malformed or empty header value falls back to the peer address
  rather than to an empty string, because an empty limiter key is one
  bucket every attacker shares.
- The mode descriptor states which of the two is live, so 5d's
  `doctor` row reports what runs rather than what was configured.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The package imports `net` and `net/http` and no `internal` package.
  `internal/config` does not appear.
- The peer-address function reads `r.RemoteAddr` and no header, provable
  by reading the function in full.
- One behavioral row holds threat-model claim 3 in the task that lands
  it: mode on, an allowlisted peer, and a client-address header naming
  a different address, asserting the peer-address function returns
  `RemoteAddr`'s host. Revision 1 held the pass's central trust claim by
  the structural criterion above, and admitted that its own mutation 4
  fired only in later tasks.
- The table test's proxy-mode rows carry a peer address that is not in
  the allowlist and a header that names an address that is, so a test
  cannot pass by conflating the two.
- The right-to-left parse is asserted with a value carrying three hops,
  two of them configured proxies and the leftmost attacker-supplied,
  and the expected answer is written out rather than computed from the
  input by the test.
- Every expected address in the table is a documentation-reserved or
  invented literal, never an estate address.
- IPv6 peers are covered, in bracketed `host:port` form, since
  `net.SplitHostPort` is the only correct reader of `RemoteAddr` and a
  naive split on `:` is the bug this row catches.
- The doc comments state the two-jobs rule and the reason for the parse
  direction, and state nothing the signature already says.
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Read the header without checking the peer address → the
   non-allowlisted-peer row fails, reporting the header's value where the
   peer address is wanted.
2. Parse the multi-hop value left to right → the three-hop row fails,
   reporting the attacker-supplied leftmost entry.
3. Return the header's value when the peer matches but the header is
   empty → the malformed-value row fails, reporting an empty key.
4. Make the peer-address function consult the header when mode is on →
   **the claim-3 row above fails in this task**, and every trust-decision
   test in Tasks 2 and 5 that uses it fails too. The task reports the
   mutation as one whose blast radius is the pass rather than the file.
5. Split `RemoteAddr` on the last `:` by hand instead of
   `net.SplitHostPort` → the IPv6 row fails.

**2.0 hook: neutral, and it keeps hook 6 open.** The derivation takes its
proxy set as values rather than reading a driver's configuration, so rung
12 contributes nothing here and the seam stays universal
(`spec:1843`).

**Independent: yes**, by files. It writes `clientaddr.go` only and
consumes nothing from any other task. It runs first because Tasks 2 and 5
consume it.

**Prior art:** adopts the field's normal answer rather than inventing
one. **Miniflux pairs `AUTH_PROXY_HEADER` with an explicit
`TRUSTED_REVERSE_PROXY_NETWORKS`** (PRA:558-567), which is the same shape
this derivation takes and is evidence that explicit configuration is the
field's answer. **Authelia requires a reverse proxy to remove and replace
`X-Forwarded-For` arriving from an untrusted client**, warning that a
rule keyed on it can otherwise be hijacked (PRA:579-591), which is the
argument behind Task 2's group 2. **Declines** any third-party real-IP
middleware: the rule is fifteen lines and the dependency budget is R26's
two, both 5b's.

**Deliverable count: 3** (the derivation, the mode descriptor, the
tests). Gate green; commit.

---

## Task 2: The three-group ingress strip

Model: `sonnet`. Depends on Tasks 0 and 1.

**Files:**
- Create: `internal/serve/strip.go`
- Create: `internal/serve/strip_test.go`

**Interfaces:**
- Consumes: Task 1's peer address and trusted-proxy configuration.
- Produces: the strip's data (the extras list, generated in both
  spellings from one source) and a wrapper taking an `http.Handler` and
  returning one, clearing header fields and trailer fields alike. Task 5
  composes it as the outermost step, and Task 5 owns every assertion that
  needs the owned prefix or the fallthrough to exist.

**Outcomes (C2):**

- The strip wraps the mux rather than living in any handler, which is
  what makes "every path" true without enumerating paths
  (`spec:767-771`).
- Group 1 removes every identity spelling in both dash and underscore
  form. The driver's configured header name is a construction value,
  empty at rung 5, and **the mechanism by which a driver contributes it
  is what this task ships**. Rung 12 adds a name, never a line of this
  file.
- Both spellings are generated from one source, so a new name is one
  line of data and cannot be added in one form only (`spec:2231-2236`).
- The underscore behavior is asserted, not remembered. A test proves
  that `textproto.CanonicalMIMEHeaderKey` leaves `Remote_User` and
  `Remote-User` as two distinct map keys and that deleting one leaves the
  other standing (`spec:779-787`). **That test asserts on the stdlib's
  behavior directly**, so it fails if the toolchain ever changes it.
- The strip clears `r.Trailer` of every name in the set, on the same
  pass as `r.Header` (R-P). Go moves announced trailer keys out of the
  header map before the handler runs, so a header-only strip leaves a
  trailer-borne identity value standing and `httputil.ReverseProxy`
  forwards it to the upstream.
- Group 2 removes `X-Forwarded-For`, `X-Real-IP`, `Forwarded`,
  `X-Forwarded-Host`, `X-Forwarded-Proto`, and `X-Forwarded-Prefix` on
  every path, unless trusted-proxy mode is on and the peer address
  matches (Task 1's decision, not a second reading of it). The family
  joins the first three under R-R, on B12's proxy-hygiene-as-a-class
  clause.
- Group 3 removes every header the request's own `Connection` header
  lists, before routing, **with two carve-outs**: a request whose
  `Connection` names `Upgrade` keeps `Upgrade` and `Connection` (R-O),
  and the keep-set `Cookie`, `Origin`, `Sec-Fetch-Site`, and the
  session-cookie name is never removed by the listed rule (R-S). The
  fixed RFC hop-by-hop names are not removed here (R-O).
- The exemption exists as a mechanism and is unreachable at rung 5.
  Exactly one identity header may be exempted, it is the one the active
  front's configuration names, the exemption applies only when the peer
  address matches that front's configured addresses, and **every other
  spelling is still stripped, the exempted header's underscore twin
  included** (`spec:849-862`). A shell constructed with no front exempts
  nothing. **A request carrying the exempted name twice is collapsed to
  its last value** (R-Q), because `Header.Get` returns the first and a
  front that appends rather than replaces produces that shape.

**The extras list is verified name by name, not carried from memory.**
`spec:774-778` requires it: "a fixed extras list this spec proposes and
the rung 5 plan verifies name by name against the documentation of each
reference implementation it names."

**The task's method, stated as a constraint rather than a list:** for
each candidate name, find the project's own documentation, record the URL
and the read date, and **admit the name only if that documentation states
it as a header the project sets or reads for identity**. Candidate
sources, each named because the spec's trust-boundary and prior-art
sections name the project: **Authelia's forwarded-headers documentation**
(PRA:579-591), **Miniflux's configuration documentation** for
`AUTH_PROXY_HEADER` (PRA:558-567), **oauth2-proxy's** upstream-header
documentation, **Cloudflare Access's** documented request headers, and
**Navidrome's** reverse-proxy authentication documentation, since it is
the streamer this build proxies. **A name the task cannot cite is dropped
and reported, never shipped on recollection.** The task reports the final
list with a citation per name and names every candidate it dropped.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The extras list is a single data declaration, and the underscore forms
  are derived from it rather than typed. A reviewer can add a name by
  adding one line.
- Every shipped name carries its documentation URL and read date in
  the declaration's own comment or in the task report, and the report
  names every dropped candidate with the reason.
- The test enumerates **the whole set** and asserts a recording handler
  received none of them, in both spellings. **The wrapper is the unit
  under test and two recording handlers are its terminals.** The
  under-the-prefix and on-the-fallthrough halves of `spec:938-941` are
  Task 5's, because the prefix is Task 4's artifact and the fallthrough
  is Task 5's; revision 1 asked this task for both.
- The test's expected set is written out rather than read from the
  strip's own data, so a strip that lost a name still fails.
- A chunked request announcing `Trailer: Remote-User` and sending the
  value after its body asserts the handler sees the name in neither
  `r.Header` nor `r.Trailer`, checked before and after the body is
  read, since Go populates the value only on the read (R-P).
- The `Connection`-listed removal is asserted with `Connection:
  Remote-User`, which is the induced-downgrade attack by name, and with a
  multi-valued `Connection` header.
- A `Connection: Upgrade, Remote-User` row asserts `Remote-User` gone
  and `Upgrade` and `Connection` intact (R-O), and **a
  `Connection: Origin` row asserts `Origin` standing** (R-S).
- The group 2 rows cover mode off, mode on with a matching peer, and mode
  on with a non-matching peer, and **the third asserts removal**. **Every
  member of the `X-Forwarded-*` family appears in the enumerated expected
  set**, so a missing name fails.
- The exemption's test constructs the strip with a front configured,
  proves the one named header survives for a matching peer, **proves its
  underscore twin does not**, and **proves a duplicate of the exempted
  name collapses to its last value** (R-Q).
- No handler in the package reads an identity header, provable by grep
  over the package.
- `go.mod` unchanged. Gate green.

**Named mutations this task reports:**

1. Drop the underscore derivation → the both-spellings rows fail, half
   the set surviving.
2. Delete after the wrapper returns instead of before calling the next
   handler → **every enumerated row fails**, the recording handler seeing
   the whole set. Revision 1 aimed this mutation at a fallthrough row
   this task does not own; the composed-chain version of it is Task 5's
   mutation list.
3. Skip group 2 when mode is off → the mode-off row fails, reporting
   `X-Forwarded-For` present.
4. Remove the `Connection`-listed removal → the induced-downgrade row
   fails.
5. Exempt the configured header for every peer rather than a matching one
   → the non-matching-peer exemption row fails.
6. Read the peer address from `X-Forwarded-For` inside the strip instead
   of from Task 1's function → the group 2 non-matching-peer row fails,
   and the task reports that this is the mutation Task 1's mutation 4
   predicted.
7. Drop the trailer clause → the trailer row fails, the handler reading
   the identity value out of `r.Trailer` after the body (R-P).
8. Remove every `Connection`-listed name unconditionally → the upgrade
   row fails, `Upgrade` gone from a handshake (R-O), and the
   `Connection: Origin` row fails with it (R-S).
9. Take the exempted header's first value instead of its last → the
   duplicate-exemption row fails (R-Q).
10. Drop the `X-Forwarded-*` family and keep the original three → the
    enumerated expected set fails on the missing names (R-R).

**2.0 hook: neutral, and it keeps hook 6 open.** `spec:1843` names the
strip set staying data as half of what keeps the driver seam universal.
This task lands that half. **The task states it explicitly in its report**,
so the build-for-2.0 sweep reads a decision.

**Prior art:** **adopts** stdlib `net/textproto`'s canonicalization as
the mechanism and asserts it. **Adopts** Authelia's forwarded-header rule
as the argument for group 2 (PRA:579-591), extended to the whole
`X-Forwarded-*` family Authelia's own documentation covers (R-R), and
Miniflux's explicit-pairing shape (PRA:558-567). **Adopts**
`httputil.ReverseProxy`'s upgrade carve-out as the precedent for R-O's:
it reads `upgradeType` at `reverseproxy.go:464`, removes hop-by-hop
headers at `:469`, and re-adds `Connection: Upgrade` and `Upgrade` at
`:481-484`. **Declines** any third-party header-scrubbing
middleware: the whole mechanism is a map delete over a generated list.
**Reads `_salvage/internal/serve/serve.go` for nothing here**, because the
salvage server has no strip at all, which is what SEC:42 filed.

**Independent: no.** Depends on Task 1.

**Deliverable count: 3** (the strip data with its citations, the wrapper
over headers and trailers, the tests). Gate green; commit.

---

## Task 3: Path normalization

Model: **`opus`**. Depends on Tasks 0 through 2. **The pass's one named
upshift**, argued in the execution model: the rules are specified and
**the mechanism is not**, and R-D's trap (an empty `RawPath` on a path
that needed no escaping) is the kind of defect that passes every ordinary
test and fails exactly where the guarantee matters.

**Files:**
- Create: `internal/serve/normalize.go`
- Create: `internal/serve/normalize_test.go`

**Interfaces:**
- Consumes: nothing from earlier tasks.
- Produces: a wrapper taking an `http.Handler` and returning one, which
  refuses a request outright or passes it through untouched. Task 5
  composes it as the second step.

**Outcomes (C3):**

- Each rule is a refusal with a status, and no partial routing
  happens. A refused request is answered by the wrapper; no route, no
  prefix decision, and no fallthrough sees it.
- Rule 1, no percent-encoding in the path. `EscapedPath()` and `Path`
  must be identical (R-D). This closes `%2F`, `%5C`, and `%2E%2E`, and it
  refuses a benign encoded byte too, which is the stated policy rather
  than a side effect.
- Rule 2, any `..` or `.` segment, refused rather than cleaned.
- Rule 3, a doubled separator anywhere in the path. The leading
  separator is not a segment and a single trailing separator is admitted.
- Rule 4, any backslash, which matters on the fallthrough where rung
  12 forwards a raw path to a streamer whose own router may read it as a
  separator (`spec:923-925`).
- Rule 5, a path past the stated byte bound (R-H), answered `414`.
- Rule 6, a request target whose path does not begin with a separator
  (R-U). An absolute-form target is admitted and routes on its path.
- The wrapper wraps the mux and cannot be registered as a route. The
  package exposes no way to register a check as a route
  (`spec:2224-2226`), because `http.ServeMux` performs its own path
  cleaning and redirect internally and a check registered as a route would
  run after the decision it exists to guard (`spec:925-928`).
- The refusal status is `400` and every rule but the length bound uses
  it (R-T), so a client cannot distinguish which smuggling rule refused
  it. The length bound answers `414`, and the shape it refuses is
  distinguishable from outside anyway.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The comparison is the single predicate `EscapedPath() != Path`, and
  **the doc comment states why `RawPath` alone is wrong**: it is
  populated only when the escaped form differs from the default encoding,
  so an unencoded path carries an empty value. **The doc comment also
  states that the rule refuses benign encoding**, which is R-D's policy
  and the thing a later reader will otherwise take for a bug.
- The test table carries, at minimum: a plain path that passes; `%2F` in
  a segment; `%5C`; a double-encoded slash; a benign encoded byte
  (`/%E2%98%83`); `..` plain; `..` percent-encoded; a single `.` segment;
  a doubled separator; `/`; the bare prefix with a trailing separator;
  `//` at the front; a backslash; a path at the bound; a path one byte
  past it; a `CONNECT` request target; and an absolute-form request
  target.
- Every expected status is a literal in the table, never read from the
  implementation's constant, **and every row carries one**. The passing
  rows (`/`, the trailing separator, absolute-form) carry the terminal
  handler's own assertion instead.
- Each refusal row asserts that the wrapped handler was never called,
  which is the no-partial-routing property, and the test's terminal
  handler records its own invocation rather than trusting the status.
- The passing rows assert the handler received **the path unchanged**,
  byte for byte, so normalization is a gate rather than a rewriter.
- No exported function in the package registers a check as a route, and
  the wrapper's only entry point takes and returns an `http.Handler`.
- `go.mod` unchanged. Gate green.

**Named mutations this task reports:**

1. Compare `RawPath` against `Path` instead of `EscapedPath()` → **the
   plain-path row fails**, because an unencoded path carries an empty
   `RawPath`. This is R-D's trap firing, and the task reports it as the
   mutation that proves the trap is real.
2. Clean `..` instead of refusing it → the `..` rows fail, reporting a
   200 from the terminal handler.
3. Drop the backslash rule → the backslash row fails.
4. Drop the doubled-separator rule → the `//` row fails.
5. Register the check as a route on a `ServeMux` instead of wrapping it →
   **the `..` row fails**, because the mux cleans and redirects first, and
   the task reports the observed redirect status as the evidence risk 1
   asks for.
6. Refuse with a different status per rule → the distinguishability the
   acceptance criteria forbid becomes visible, and **the status literals
   in the table catch it**, since every refusal row but the length bound
   carries `400` (R-T).
7. Refuse any empty segment rather than a doubled separator → **the `/`
   row and the trailing-separator row fail**, which is the contradiction
   revision 1 carried between C3 and Task 4's table.
8. Drop the origin-form rule → the `CONNECT` row fails, the terminal
   handler recording an invocation on a request with no path (R-U).
9. Admit a benign encoded byte → the `/%E2%98%83` row fails, and the task
   reports which of R-D's two policies the implementation took.
10. Drop the `.` segment from rule 2 → the single-dot row fails, and the
    task reports the `http.ServeMux` redirect it would have produced one
    level down.

**2.0 hook: neutral.** Normalization is a refusal, and 2.0 adds no path
shape it must admit. **The task states the decline** so the sweep reads
one. **One consequence is recorded for the forward criterion**: R-D's
blanket refusal of percent-encoding constrains every path rung 12
proxies and every path 2.0's OpenSubsonic surface serves.

**Prior art:** **adopts** stdlib `net/url`'s escaping model and asserts
its documented behavior rather than assuming it. **Reads
`_salvage/internal/serve/serve.go:82-107`** for the defect class it
proves: `mountBasePath`'s own comment records that a plain trim admits
`/uploadfoo`, which is the same bug shape one level down and is **evidence
the class is real in this codebase rather than hypothetical**
(`INV:290-301`). **Declines** every third-party path-sanitizing
middleware: a sanitizer rewrites, and every rule here refuses.

**Independent: yes**, by files. It writes `normalize.go` only and consumes
nothing from Tasks 1 or 2. It is ordered here because Task 5 composes it.

**Deliverable count: 2** (the wrapper over six rules, the tests). Gate
green; commit. **Checkpoint: write STATUS after this task.**

---

## Task 4: Prefix routing and the terminator

Model: `sonnet`. Depends on Tasks 0 and 3.

**Files:**
- Create: `internal/serve/prefix.go`
- Create: `internal/serve/prefix_test.go`

**Interfaces:**
- Consumes: nothing at runtime; it is ordered after normalization by Task
  5's composition.
- Produces: a value carrying the owned prefix, a segment-aware match, and
  the terminator handler. Task 5 composes it as the third step and
  registers subtrees under it.

**Outcomes (C4):**

- The match is segment-aware. A path enters the owned surface only
  when the prefix is followed by a separator or by nothing at all.
  `/dubplatefoo` does not enter, which is the `mountBasePath` defect
  refused rather than documented.
- The match is case-sensitive bytes. `/Dubplate` does not enter the
  owned surface and **reaches the fallthrough**, which is the exit gate's
  own assertion (`spec:2035`). A case-insensitive match would widen the
  owned surface past what an adopter configured.
- The terminator answers an unknown path under the prefix with 404 from
  dubplate, and **that request never reaches the fallthrough**
  (`spec:900-904`). Rung 14 inherits it so an unknown API path cannot be
  captured by rung 12's proxy (RE cell (14,12)).
- The bare prefix with no trailing separator is handled explicitly
  rather than left to `http.ServeMux`'s redirect, because the redirect is
  a decision the shell must own. The task states which of redirect or
  direct handling it takes and why, and asserts it.
- The prefix value is validated at construction: it begins with a
  separator, carries no trailing separator, and carries no character the
  normalization rules would refuse. **An invalid prefix is a construction
  error wrapping a package sentinel**, not a runtime 500.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The match is not `strings.HasPrefix` alone and is not
  `http.StripPrefix`. The doc comment cites the salvage comment's
  plain-trim finding as the reason.
- No call anywhere in the package lowercases or case-folds a path before
  matching.
- The test table carries: the bare prefix; the prefix with a separator;
  a path under it; `/dubplatefoo`; `/Dubplate`; `/dubplate` with a
  different case in a later segment (which **must** enter, since only the
  prefix is case-checked); the root path `/`; and a path that does not
  begin with the prefix.
- The bare prefix, the prefix with a trailing separator, and `/` each
  carry a written verdict: the first two enter the owned surface and
  reach the terminator when nothing is registered, and `/` reaches the
  fallthrough. C3 rule 3 admits all three, which is what revision 1's
  "any empty segment" contradicted.
- Each row asserts which of three terminals received the request, the
  registered route, the terminator, or the fallthrough, using three
  distinct recording handlers, so a row cannot pass by reaching the
  wrong one with the right status.
- The terminator row asserts **404 and that the fallthrough handler was
  never called**, which is the property rung 14 depends on.
- The construction-error rows assert through `errors.Is` on a package
  sentinel, never on message text.
- The prefix in every test is a literal supplied by the test, never the
  configured default, so the tests would pass under R-A's other answer.
- `go.mod` unchanged. Gate green.

**Named mutations this task reports:**

1. Replace the match with `strings.HasPrefix` → the `/dubplatefoo` row
   fails, entering the owned surface.
2. Lowercase the path before matching → the `/Dubplate` row fails,
   reaching the owned surface instead of the fallthrough.
3. Fall through instead of terminating on an unknown owned path → the
   terminator row fails, the fallthrough handler recording an
   invocation.
4. Terminate with 200 and an empty body → the status assertion fails.
5. Accept a prefix with a trailing separator at construction → the
   construction-error row fails, and the task reports which double
   separator that would have produced at match time.

**2.0 hook: neutral.** 2.0's OpenSubsonic surface lands under the same one
owned prefix (`spec:494-495` reading S20:284-285), so the terminator is
what keeps that surface from leaking into the fallthrough. **The task
records this as a hook the terminator protects rather than lands.**

**Prior art:** **adopts** the salvage tree's own lesson from
`mountBasePath` (`INV:290-301`) by refusing rather than documenting the
gap. **Declines `Mount(prefix, http.Handler)`**, already examined and
rejected at PA:77: by the one-owned-prefix rule the streamer owns
everything else, so the driver is the root mux's fallthrough rather than a
mounted subtree. **Declines** `http.StripPrefix` as the boundary
mechanism for the same reason its own salvage caller needed a guard.

**Independent: yes**, by files. It writes `prefix.go` only and consumes
nothing from Tasks 1, 2, or 3. It is ordered here because Task 5 composes
it.

**Deliverable count: 3** (the prefix value with its validation, the
terminator, the tests). Gate green; commit.

---

## Task 5: The extension point, the fallthrough, and the root composition

Model: `sonnet`, **with a conditional upshift to `opus` if Task 3's review
returned `fix`.** This is the pass's headline artifact and risk 1's
target.

**Files:**
- Create: `internal/serve/serve.go` (the shell's construction and the
  composed root handler)
- Create: `internal/serve/serve_test.go`
- Create: `internal/serve/doc.go` (the package doc comment stating the
  composition order and the extension point's contract)

**Interfaces:**
- Consumes: Tasks 1, 2, 3, and 4.
- Produces: the shell's construction value with **a subtree registration
  method** and **a fallthrough slot**, and a constructor returning the
  composed root `http.Handler`. Task 6 gives it an `http.Server` and
  listeners; Task 7 calls it from the binary; 5b, 5c, 5d, and rungs 6, 7,
  12, and 14 register against it.

**Outcomes (C1, C5, and the session-cookie removal):**

- The four steps compose in order: strip, normalization, prefix
  routing, fallthrough. The order is a property of the chain rather than a
  feature of the mux (`spec:907-909`).
- A consumer registers a subtree under the owned prefix by name at
  construction time, in the binary's own wiring, and **the shell owns
  the mux the registration lands in** (R-L).
- Two registrations of the same subtree panic at construction, on the
  precedent rung 1b's registry cites (`registry.go:24-31`). The panic
  names the subtree.
- The fallthrough slot takes at most one handler. A second is a
  construction error or a panic, and the task states which and why. **A
  deployment with no streamer leaves it empty, which is what makes rung 5
  safe to ship without a driver** (`spec:955-957`).
- An empty fallthrough answers 404 (R-N).
- The session cookie is removed from any request reaching the
  fallthrough (R-E), with the cookie name a construction value. A shell
  built with no cookie name removes nothing, which is 5a's own wiring.
- The static surface's two-handler shape is admitted, not built (C5).
  The seam must let rung 7 swap the embedded bundle for a larger one
  without touching either handler, and 5a proves that by registering two
  distinct subtrees in a test.
- The package doc comment states the composition order, the extension
  point's contract, and which point the driver's proxy occupies, which
  is PA:79's requirement discharged in the code rather than only in the
  spec.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Three order assertions, each behavioral (`spec:938-944`): a request
  bearing every spelling in the strip set reaches a registered handler
  with none of them; a request whose escaped and decoded paths disagree
  is refused and **no registered handler and no fallthrough handler is
  invoked**; a path that string-prefixes the owned prefix without a
  segment boundary reaches the fallthrough and not the owned surface.
- A fourth, which is the one the route table cannot show: a request
  carrying both an identity header and a `..` segment is refused **and
  the identity header was already gone when the refusal happened**. This
  asserts strip-before-normalization, an ordering no status code reveals.
  **The mechanism is a structured log record on the refusal path**, which
  names whether any strip-set spelling was present on the refused
  request, asserted through a test logger against the production root
  handler. Revision 1 asked for "a normalization step that records what
  it saw", which no test outside the chain can observe on a request that
  reaches no handler, and whose only implementations were a rebuilt chain
  (structural, forbidden by the Global Constraints) or a test-only seam.
  The same record is what R-N promised for telling the terminator from
  the empty fallthrough, so the two share one mechanism.
- The strip's under-the-prefix and on-the-fallthrough halves land
  here, moved from Task 2, which owns neither artifact: a request
  bearing every spelling in the set reaches a handler registered under
  the owned prefix with none of them, and the same request reaching the
  fallthrough handler carries none of them either (`spec:938-941`).
- The duplicate-subtree registration is asserted with `recover`, and the
  panic's message names the subtree.
- The empty-fallthrough row asserts 404 with no handler registered.
- The session-cookie row asserts the cookie is absent at the fallthrough
  handler **and present at a handler registered under the owned prefix**,
  since the removal is on the way out and not a second strip group.
- Two more cookie rows, because the stated one passes for an
  implementation that deletes the whole `Cookie` header: a request
  carrying three cookies including the named one asserts the other two
  arrive at the fallthrough unchanged, and a request carrying the named
  cookie twice asserts neither survives. Deleting the header outright
  would satisfy the row above while removing the streamer's own cookies
  from every proxied request, which surfaces at rung 12 as a broken
  session with a green 5a suite behind it.
- The two-subtree row registers two distinct names and proves each
  receives only its own paths.
- No test in this file reads the shell's construction to build its
  expectation. Every expected value is written out.
- The doc comment names all four steps in order and names the fallthrough
  as the driver's point.
- `go.mod` unchanged. Gate green.

**Named mutations this task reports:**

1. Swap the strip and normalization steps → **the fourth order assertion
   fails**, the normalization step recording an identity header it should
   never see. No other test in the pass fires, which is why that
   assertion exists.
2. Register normalization as a route on the mux instead of wrapping it →
   the `..` order assertion fails, reporting the mux's own redirect. This
   is risk 1 firing.
3. Put the prefix match before normalization → the raw-versus-decoded
   assertion fails, a smuggled path having already been routed.
4. Let the fallthrough see a request refused by normalization → the
   no-handler-invoked assertion fails.
5. Allow a duplicate subtree registration silently → the `recover` row
   fails.
6. Remove the session cookie on every path instead of at the fallthrough
   → the owned-surface half of the cookie row fails, and the task reports
   that this mutation would have broken 5b's every session read.
7. Delete the whole `Cookie` header at the fallthrough → the
   surviving-cookies row fails.
8. Strip after the prefix decision instead of before → the
   under-the-prefix strip row fails, the registered handler seeing the
   set.

**2.0 hook: neutral, and it lands the half of hook 6 that keeps the
driver seam universal.** `spec:1843`: "Rung 5 leaves the fallthrough
empty and the strip set as data." **This task lands the fallthrough half**
(Task 2 landed the strip half), and the two together are rung 5's whole
position on hook 6. The task states it in its report.

**Prior art:** **adopts** stdlib `net/http`'s handler-wrapping idiom, and
`http.ServeMux` for registered-subtree dispatch **inside** the owned
prefix only, never as the boundary mechanism. **Declines
`Mount(prefix, http.Handler)`** (PA:77) for the reason C1 step 4 states.
**Declines** every third-party router: the shell registers a handful of
subtrees, and a router's value is pattern matching this design refuses to
do at the boundary.

**Independent: no.** It composes Tasks 1 through 4.

**Deliverable count: 4** (the construction value with its registration
method, the fallthrough slot with its cookie removal, the composed root
handler, the tests). Gate green; commit.

---

## Task 6: The server, its deadlines, and the listener

Model: `sonnet`. Depends on Tasks 0 and 5.

**Files:**
- Create: `internal/serve/server.go`
- Create: `internal/serve/server_test.go`

**Interfaces:**
- Consumes: Task 5's composed root handler.
- Produces: the `http.Server` construction, the write-deadline extension
  mechanism exposed to the fallthrough handler, the `IP_FREEBIND`
  listener set, and `Run(ctx, ...)`. Task 7 calls `Run` from the verb.

**Outcomes (C6, C8):**

- `ctx` is the first parameter of `Run`, and `Run` serves until it is
  done, then shuts down with a drain bounded by a stated grace.
- Four timeouts are set: a read-header timeout, **a read timeout**
  (R-V), an idle timeout, and a default write timeout.
  **`MaxHeaderBytes` is set explicitly** (R-H), so the number is a
  decision rather than an inheritance. The read timeout is what bounds a
  body trickled a byte at a time, which the other three do not reach, and
  it is the constant the salvage server carries at
  `_salvage/internal/serve/serve.go:600` that revision 1 dropped by
  describing that server as setting a write timeout alone.
- The default write timeout does not become rung 12's problem. The
  extension mechanism is `http.ResponseController.SetWriteDeadline`, and
  **the shell exposes it to the fallthrough handler**
  (`spec:975-981`), so a streaming handler extends or clears its own
  deadline while the server's default protects everything else.
- Rung 5 has no streaming route of its own, so the mechanism is
  verified against a test handler that writes past the default rather
  than against the constant (`spec:978-981`). **This is the exit gate's
  own assertion** (`spec:2035`).
- The listener binds exactly the configured addresses, never
  `0.0.0.0` (C8). Binding exactly the list is what refuses a connection
  to any address not in it, since nothing listens there.
- `IP_FREEBIND` is set through `net.ListenConfig.Control` before
  bind, so the listener can bind an address the host has not configured
  yet (R20). **The stdlib `syscall` route is used** (R-M), and the
  constant carries its source in a comment. **The hook sets the option at
  `IPPROTO_IP` for both address families**, measured working on an
  `AF_INET6` listener, so no `IPV6_FREEBIND` constant is needed and none
  exists for this architecture.
- The boot property is asserted rather than argued. A listener set
  containing a documentation-reserved address that is on no interface
  opens successfully with the hook, and the same address fails to bind
  without it. Measured during the revision 2 fold: `192.0.2.77:0` returns
  `bind: cannot assign requested address` from a bare
  `net.ListenConfig`, and binds under the hook, unprivileged. Revision 1
  recorded the whole mechanism as unprovable here.
- A failure part way through opening the set closes every listener
  already opened before returning, so a partial bind leaks nothing.
- The addresses actually bound are readable after `Listen` returns
  (R-G), so 5d's `doctor` row reports what runs.
- `IP_FREEBIND`'s residual is stated in the doc comment: it suppresses
  the boot-time error a misconfigured address would raise, and the row
  that surfaces it instead is 5d's.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Every timeout is a named constant with a comment carrying its reason,
  and **the default write timeout's comment names the SSE and audio case
  it must not kill** and the mechanism that saves it.
- The extension test **writes past the default write deadline after
  extending it and succeeds, and a sibling row that does not extend
  fails**, so the test proves the deadline is real as well as
  extensible.
- Every timeout constant appears in the `http.Server` literal, **the read
  timeout among them**, and the read timeout's comment names the slow-body
  case and the two 5b POSTs that take bodies.
- The listener test binds an ephemeral port on loopback, asserts the
  bound address is readable, and asserts a second address in the same
  call also bound. **One row binds an IPv6 loopback address**, since the
  hook's socket level is the same for both families and the assertion is
  what proves it.
- The partial-failure row asserts no listener is left open, using an
  address that cannot bind alongside one that can. **That row's
  unbindable address is not the freebind row's address**, since freebind
  makes an absent address bindable.
- The `Control` hook is asserted behaviorally: a listen on a
  documentation-reserved address that is on no interface succeeds, and a
  sibling row without the hook fails with `EADDRNOTAVAIL`. The socket
  option is read back with `syscall.GetsockoptInt` in the same hook as a
  second, cheaper assertion.
- `Run` takes `ctx` first, and the shutdown path derives its timeout
  context with `context.WithoutCancel` so a cancelled parent does not
  cancel the drain.
- No `0.0.0.0` literal exists in the package, provable by grep.
- `go.mod` unchanged, **and `golang.org/x/sys` does not appear** (R-M).
  Gate green.

**Named mutations this task reports:**

1. Drop the write deadline entirely → **the sibling row that expects a
   failure without extension fails**, which is why that row exists.
2. Drop the `ResponseController` extension → the extending row fails.
3. Bind `0.0.0.0` instead of the list → the second-address row still
   passes and the grep criterion catches it. The task reports which of
   the two found it.
4. Return on the first bind error without closing the openers → the
   partial-failure row fails, a listener still holding its port.
5. Drop the `Control` hook → **the absent-address row fails**, the listen
   returning `bind: cannot assign requested address`. Revision 1
   pre-authorized this mutation as unfired on the belief that every test
   address is already present; a documentation-reserved address is not.
6. Take `ctx` off `Run` and read a package context → the compile fails,
   and the task reports it as a mutation the language catches.
7. Drop the read timeout → **no test fails**, because a slow-body test is
   a timing test this gate should not carry. The task reports it unfired
   and names the criterion that holds it, the constant's presence in the
   server literal with its reason comment (R-V).

**2.0 hook: this task keeps hook 6's streaming half open.** 2.0's own
streamer serves audio and SSE from the fallthrough's successor, and the
deadline extension is the mechanism that lets it. **The task records the
mechanism as the hook's enabler rather than as a hook it lands.**

**Prior art:** **adopts** `_salvage/internal/serve/serve.go:561-596`'s
`IP_FREEBIND` mechanism, per R20, **adopts its read timeout** (`:600`,
applied at `:631`), **and rejects its write-timeout policy**: the
30-second write timeout on one `http.Server` (`:601`, applied at `:632`)
is the lesson B12 answers, and INV surprise 10 records it
(`INV:1244-1248`).
**Adopts** stdlib `http.ResponseController`, which `spec:975-977` names
and which is the whole per-route mechanism. **Declines
`golang.org/x/sys/unix`** (R-M) and declines every graceful-shutdown
library: `http.Server.Shutdown` is the mechanism.

**Independent: no.** It serves Task 5's handler.

**Deliverable count: 4** (the server construction, the deadline extension
mechanism, the listener set, the tests). Gate green; commit.

---

## Task 7: The serve config class, the reserved-class retirement, and the verb

Model: `sonnet`. Depends on Tasks 0 through 6. **This task carries the
pass's one cross-cutting edit and its one contract change.**

**Files:**
- Create: `cmd/dubplate-server/serve.go`
- Create: `cmd/dubplate-server/serve_test.go`
- Modify: `internal/config/keys.go` (delete the `serve` reserved class
  only)
- Modify: `cmd/dubplate-server/main.go` (one `registerServeConfig(reg)`
  call above the block, one `newServeCmd(...)` line inside it)

**Interfaces:**
- Consumes: Task 6's `Run`, Task 5's construction value,
  `internal/config`'s registry.
- Produces: three registered thresholds, the `serve` verb, and the
  resolution that turns configuration into the values the shell takes.

**Outcomes:**

- Three keys are declared in rung 5's own file (R-C), following rung
  4's precedent exactly: `DUBPLATE_SERVE_BIND_ADDRS` (`KindString`,
  default `127.0.0.1`), `DUBPLATE_SERVE_PORT` (`KindInt`, default
  `8080`), and `DUBPLATE_SERVE_PREFIX` (`KindString`, default
  `/dubplate`) (R-A, R-B).
- The `serve` reserved class is deleted from `internal/config/keys.go`
  in the same commit, which is the removal half of the rung 1b contract
  **exercised here for the first time** (`spec:1494-1498`,
  `HISTORY:493-498`).
- The class's own argument survives the move. `keys.go:75-78` carries
  a comment saying the wall keeps the serve keys apart from auth because
  a bind address is not an identity concern. **The new declaration's doc
  comment carries that reasoning forward** rather than dropping it with
  the code.
- `DUBPLATE_SERVE_PREFIX` is new, and its declaration states that it
  is B3's one owned prefix rather than a routing convenience.
- The bind list is split and validated in `internal/serve`, not in
  the config declaration (R-B), so the address grammar lives with the
  code that binds. An unparseable address is an error naming the value.
- The verb wires into the single `AddCommand` block, one line, beside
  `newShadowCmd`. **The two register calls sit above the block** with
  rung 4's, in declaration order (`spec:982-1005`).
- `doctor` needs no change, because it derives every row from the
  registry (`spec:1000-1005`). The net row change is **minus one reserved
  class, plus three thresholds**, and Task 0 item 8 has already computed
  it against the two landed line-count assertions.
- `internal/serve` still imports no config package. The verb resolves
  values and passes them in.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The `keys.go` diff is a deletion and nothing else. **No name is added
  to that file**, which is the contract in the words rung 1b wrote.
  The `auth and identity` class, the pending identity-source wall row,
  and every supersession message are untouched.
- The three declarations sit in `cmd/dubplate-server/serve.go` beside the
  verb, and `registerServeConfig` is called from `newRootCmd` above the
  `AddCommand` block.
- A gate-resident test constructs the production registry exactly as
  the binary does and asserts the three keys are claimed and the `serve`
  reserved class is gone, on the precedent `registry.go:24-31` states.
- The default values are asserted against literals in the test, so R-A's
  or R-B's other answer changes one place.
- `doctor`'s output is asserted to carry the three new threshold rows
  and no `serve` reserved row, and the two landed line-count assertions
  still pass.
- The bind-list parse is tested in `internal/serve` with a single
  address, several addresses, whitespace around a separator, an empty
  entry, and an unparseable entry.
- The verb's own test asserts it resolves configuration and refuses to
  start on an invalid prefix or an unparseable bind list, **through
  `errors.Is` on Task 4's and this task's sentinels**.
- The verb does not start a listener in any test. Binding is Task 6's
  tested surface.
- `go.mod` unchanged. Gate green.

**Named mutations this task reports:**

1. Declare the keys without deleting the reserved class → **the registry
   collision panic fires**, because the class claims the same two names,
   and the gate-resident registry test reports it. This is the contract
   enforcing itself.
2. Delete the class without declaring the keys → the declaration test
   fails, the keys unclaimed, and `doctor` loses three rows.
3. Register `DUBPLATE_SERVE_PORT` as `KindString` → the default `8080`
   parses as a string and **no test fails on the kind alone**, so the
   task reports it unfired and names the acceptance criterion (the
   declaration's kind read by the reviewer) that holds it. If a row can
   be written that catches it, the task writes it and says so.
4. Give `DUBPLATE_SERVE_PREFIX` an empty default →
   `validateThresholdDefault` panics at registration
   (`registry.go:320-322`), and the registry test reports it.
5. Wire `newServeCmd` outside the `AddCommand` block → the verb is
   unreachable, and the verb's own invocation test fails.
6. Pass `*config.Config` into `internal/serve` → the import appears, and
   the reviewer's read of the package's import list catches it. The task
   reports which found it.

**2.0 hook: neutral.** The three keys are shell configuration, and 2.0's
schema, tagging, and layout hooks reach none of them. **The prefix key is
the one 2.0 touches indirectly**, since S20 puts the whole OpenSubsonic
surface under the same one owned prefix, and the declaration's comment
says so.

**Prior art:** **adopts** the repo's own rung 1b registry and rung 4's
`cmd/dubplate-server/shadow.go` declaration precedent
(`shadow.go:19-60`), which is the shape `spec:993-996` names. **Declines**
every configuration library: the registry is the repo's own contract and
a second mechanism beside it is exactly the collision rung 1b existed to
end.

**Independent: no.** It wires Tasks 5 and 6 into the binary.

**Checkpoint: write STATUS after this task.**

**Deliverable count: 4** (the three declarations with the class deleted,
the verb, the config resolution, the tests). Gate green; commit.

---

## Task 8: The pass close, and the 5a to 5b handoff

**Done, 2026-09-08.** Pass 5a closed and merged at `307001b`. The
HISTORY entry, the security-review record, the spec corrections, and
the 5b handoff this task specifies are landed; `docs/STATUS.md` carries
the pass's close and the next action.

Model: `sonnet`. Depends on Tasks 0 through 7.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit; **the implementer never commits a ledger**)
- Draft, handed to the conductor in the task return: the spec
  corrections, the coverage-gap list, and the 5b handoff

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate
  caught, and **what a later pass would be wrong to rediscover**.
- The shell contract's coverage table, C1 through C8, each with its Go
  site, what proves it, and, for a clause no test can reach, the stated
  reason.
- The twenty-two Resolutions are listed with their markers, so the
  fold that follows a conductor ruling is a search rather than a
  re-reading. Each row names the file and the symbol the marker sits on,
  and each states what execution confirmed or contradicted. The markers
  are `// ruled, conductor 2026-09-08, R-x`, stamped by the task that
  landed each resolution, so this table is a grep.
- The declared coverage gaps are recorded as gaps, not as passing
  checks, on one criterion: a surface 5a itself builds that no test in
  this pass can reach. The expected list, to be corrected against what
  execution found:
  1. **The strip's extras list is complete only to the documentation each
     name was verified against** (Task 2). A spelling no reference
     implementation documents is a spelling the set does not carry, and
     the mitigation is that a new one is one line of data.
  2. **The exemption mechanism is unreachable at rung 5** and is tested
     against a synthetic front. No real forward-auth front exists to grade
     it until one is built.
  3. **No `doctor` row reports the client-IP mode or the bound
     addresses** (R-G). 5a exports both; 5d renders them, and the mode
     descriptor is the shell's constructed state rather than
     request-path evidence.
  4. **The session-cookie removal is tested with an arbitrary cookie
     name** (R-E), because the real name arrives at 5b. 5b's own tests
     close it.
  5. **Slowloris resistance and slow-body resistance rest on the timeout
     constants rather than on a test**, since a byte-at-a-time client is
     a timing test this gate should not carry. The pass-end reviewer's
     rubric names both attacks, and the read-header, read, and idle
     constants are what answer them.
  6. **No page, no auth handler, and no real consumer has registered at
     the extension point.** The seam is proved with test handlers, and
     5b is the first real registration.
  **Six, not the seven revision 1 expected.** `IP_FREEBIND`'s boot
  property left the list, because it turned out to be measurable here
  (Task 6). **The count is stated and joined to the running total**,
  which moves from forty-three to forty-nine (`docs/STATUS.md`'s carried
  list), so rung 15 does not re-count.
- The spec corrections the pass owes, drafted for the conductor:
  - **R24's review shape gains 5a's reviewer** (R-J), or records that the
    conductor ruled the other way and the read moved to 5b.
  - **The session-cookie removal's pass assignment** (R-E), which no pass
    row in `spec:2035-2041` currently names.
  - **The hop-by-hop scope** (R-O), which `spec:796-806` leaves open
    between the `Connection`-listed set and the fixed RFC set, **and the
    upgrade carve-out and keep-set the listed removal needs** (R-O, R-S).
  - **The strip's coverage of trailer fields** (R-P), which the spec's
    strip section does not mention and which its claim list states as a
    header-only guarantee.
  - **Group 2's extension to the `X-Forwarded-*` family** (R-R), against
    `spec:795-800`, which names three headers.
  - **Rule 1's refusal policy and rule 3's restatement** (R-D, C3), since
    `spec:918-920` states the rule and not the mechanism and revision 2
    of this plan is where the mechanism is decided.
  - **The read timeout** (R-V), against `spec:965-981`, which names three
    timeouts.
  - **Any correction Task 2's citation work forced** on the spec's claim
    that the extras list covers "the trusted-header spellings the
    forward-auth field uses", if a named project's documentation did not
    support a name the spec assumed.
  - **Any correction Task 0 found** in this plan's own pointers.
- The 5b handoff, naming what 5b inherits rather than rebuilds: the
  composed root handler it mounts auth handlers into; the extension
  point's registration method and its collision panic; Task 1's client-IP
  derivation, which the three limiter keys read and which 5b must not
  re-derive; the constraint that the origin check inserts after the strip
  and after normalization (R-F); the session-cookie name the shell already
  takes at construction (R-E); the keep-set that guarantees `Origin` and
  `Sec-Fetch-Site` survive ingress (R-S), alongside the fact that Go's
  `CrossOriginProtection.Check` returns nil when both are absent; the read
  timeout's value and the `MaxBytesReader` alternative if the conductor
  declined it (R-V); the disposal of R33 as 5b's post-login destination
  rule (R-F); rule 1's blanket ban on percent-encoding (R-D) as a named
  strictness rather than a permanent default, which rung 12 revisits
  against the streamer's real client paths, share links, cover-art
  paths, and any path a Subsonic client percent-encodes, with measured
  evidence before relaxing it; and the fact that `internal/config`'s
  `keys.go` still carries the `auth and identity` class and the pending
  wall row, both untouched and both 5b's to retire.
- The evidence, not a bare "green": the mutation table across Tasks 1
  through 7 with every outcome, including the deliberately unfired ones;
  the strip set's final name list with a citation per name and every
  dropped candidate; the measured wall times for `internal/serve`, `go
  test ./...`, and `bash scripts/check.sh`; and confirmation that no
  existing suite changed, since this pass touches no landed package but
  `internal/config`'s one deletion and `main.go`'s two lines.
- Both budgets scored: tokens against the 4.25M ceiling, and attended
  time as two counts. **The target is zero planning misses and zero
  execution sittings.**

**Acceptance criteria:**

- Every gap in the list is a gap this pass created or inherited, with the
  task and the mutation that names it.
- The mutation table has a row per named mutation across every task, and
  **every unfired mutation carries its reason**.
- `docs/STATUS.md` stays present tense and under 60 lines after the edit,
  with the pass ledger moved to HISTORY.
- No ledger is committed by the implementer, and the lane registry
  entry is the conductor's.

**2.0 hook: neutral.** The close records hook 6 as kept open by Tasks 2
and 5 together, and hook 5 as 5b's and 5c's.

**Prior art:** none consulted. The close reads the pass.

**Independent: no.**

**Deliverable count: 4** (the HISTORY entry, the STATUS edit, the drafted
corrections, the 5b handoff).

---

## The conductor's actions at the pass end

Stated here because they are not tasks and no implementer performs them.

1. **`code-simplifier` over each accepted task's changed Go**, dispatched
   after that task's review accepts and before the close. Subagents cannot
   dispatch agents, so this is the conductor's dispatch every time
   (`spec:2321-2323`).
2. **One fresh-context `web-auth-security-reviewer` over the whole
   shell** (R-J), dispatched after Task 8's gate and before the merge
   ritual. **Its input is the threat-model section of this plan, the
   spec's own threat-model section (`spec:1903-1975`), and the diff.**
   It does not receive this plan's task decomposition. Its findings are
   folded or refused with the reason recorded, and **that record is the
   do-not-re-litigate list for 5b, 5c, and 5d** (`spec:2345-2350`).
3. **The merge ritual, which closes the lane.** Merge token, rebase onto
   master, full gate in the worktree, `--ff-only` merge, full gate on
   master, and **one `diff-reviewer` over the merged result whenever the
   rebase was not a fast-forward**. The four surfaces where lane 1's 8b
   could make it non-fast-forward are `internal/config`,
   `cmd/dubplate-server`, `go.mod`, and `.gitleaksignore`, and Task 0
   item 3 reports on all four before the pass starts.
4. **Release the second code slot.** The worktree is removed once master
   is green, because the slot passes at a pass close rather than a rung
   close (`spec:2143-2145`).
5. **Re-derive 5b's ceiling from 5a's actual report** before 5b's plan is
   drafted, and record the rung total against R23's 14.5M reservation.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly sequential**,
with `repo` pointed at the lane-2 worktree. Nothing here runs in parallel,
and **no second executor enters the worktree**. Every task runs
`cairn-implementer`, **instructed in its dispatch to read and follow
`.claude/agents/dubplate-implementer.md`**, then `diff-reviewer`
(`claude-opus-5`), then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. **The runaway guard is armed at
launch. Subagents never dispatch agents.**

**Models.** `sonnet` by default. **One named upshift to `opus`:**

- **Task 3, path normalization.** The brief names the normalization and
  the strip as the two candidates and asks for a decision with its
  reason. **The strip stays `sonnet`**: its mechanism is a map delete over
  a generated list, `textproto.CanonicalMIMEHeaderKey`'s behavior is
  stated exactly in the spec and asserted by a test rather than
  remembered, and the extras list is verified against documentation
  rather than reasoned about. Nothing there is under-specified. **The
  normalization takes the upshift**: the rules are specified and the
  mechanism is not, R-D's `RawPath` trap is a defect that passes every
  ordinary test and fails exactly where the guarantee matters, and the
  rule set is the one place in the pass where a subtly wrong reading of
  the stdlib produces a shell that looks correct and is not. That is
  "novel correctness-critical logic the plan cannot fully specify" in
  CLAUDE.md's own words.

**Task 0 also runs at `opus`**, fresh context, as the brief requires: a
staleness check that misses a moved citation costs more than the model
difference, and plan staleness is the build's named dominant defect class.

**One named runner-up**, upshifted on a condition rather than by policy.
**Task 5, the composition**, if Task 3's review returns `fix`. Task 5 is
risk 1's target and the pass's headline artifact, and its correctness
rests on the normalization Task 3 landed. The composition itself is fully
specified by C1 and by the spec's own three assertions, so Task 5 stays
`sonnet` unless that evidence weakens.

**Review.** Rung 5 keeps the **full fresh-context adversarial cycle**
(CLAUDE.md's build-process section, `spec:2326-2351`), and 5a's share of
it is the in-chain `diff-reviewer` per task plus **one fresh-context
`web-auth-security-reviewer` at the pass end** (R-J), dispatched by the
conductor with the threat model as its rubric. The two-lens cycle over
the whole subsystem is 5d's.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- Task 1 → Task 2 → Task 5. The strip's group 2 reads the peer address,
  and the composition wraps the strip.
- Task 3 → Task 5. The composition wraps normalization.
- Task 4 → Task 5. The composition wraps prefix routing.
- Task 5 → Task 6 → Task 7.
- Tasks 0 through 7 → Task 8.
- **Tasks 1, 3, and 4 are mutually independent by files**, writing
  `clientaddr.go`, `normalize.go`, and `prefix.go` and consuming nothing
  from each other. **They still run sequentially**, because one tree runs
  one executor and the module gate runs over `./...` regardless. The
  marks exist so a conductor can reorder them when a dependency stalls,
  never so a second executor enters the worktree.

**Where this pass splits, if it splits.** Task 0 plus eight tasks is
under no split trigger the spec names for 5a, so **the recommendation is
to keep one pass.** The cut point, if the Task 3 re-derivation puts the
ceiling past 5M, is the **Task 4 / Task 5 boundary**, which is where the
work changes from four independent guards to the composition that
sequences them. Tasks 0 through 4 would be 5a (the two addresses, the
strip, normalization, prefix routing), and Tasks 5 through 8 would be
5a-ii (the composition, the server, the config and verb, the close).
**Two reasons argue against taking it.** The first half alone ships four
wrappers nothing composes, which is a surface no consumer can use and no
reviewer can grade for the property that matters, since **every ordering
guarantee lives in the composition**. And the pass-end security read
would land on half a shell, which defeats the reason `spec:2043-2047`
made 5a a pass of its own.

**Budget:** 4.25M ceiling, re-derived here from 8a's actual and ruled
by the conductor 2026-09-08, against `spec:2101`'s 3M. Rung 5's combined reservation
stands at 15.75M against R23's 14.5M, the difference recorded rather than
taken from another pass. **Flag at 80%, which is 3.4M**, routed to the
conductor; re-derive after Task 3; STATUS after Task 3, after Task 7, and
at the close. Spend is scored at close against the ceiling and against
**zero attended interactions**, which is this pass's target. The
twenty-two Resolutions are ruled at dispatch rather than at the close, on
8a's precedent; **none counts as an execution sitting**, and each carries
its marker in the code so a later ruling is a grep.
