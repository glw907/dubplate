# Rung 5, Pass 5c-i (The Security Fold, the `member` Verb, the CLI Admin Grant, and the Admin Roster API) Implementation Plan

Status: **revision 1, 2026-09-23, drafted just in time for plan review.**
Drafted from pass 5b-ii's close drafts
(`docs/superpowers/research/handoffs/draft-STATUS-5b-ii.md`,
`draft-HISTORY-5b-ii.md`, committed at `27cd7b5`), the 5b-ii pass-end
security review
(`docs/superpowers/research/2026-09-23-rung-5b-ii-security-review.md`),
and Geoff's 2026-09-23 routing ruling on it, against the rung 5
identity-and-auth spec at revision 2.1
(`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`), the
ratified rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`),
and the ratified web surface standard
(`docs/superpowers/specs/2026-09-12-web-surface-standard.md`).

**This plan proposes splitting 5c into two passes, and it is written as
the first of them in full with the second as a scoped outline.** The
spec's 5c row (`spec:2331`) carries fifteen deliverables. Geoff's
routing adds thirteen security findings to it, and `docs/STATUS.md`
carries five more items. Decomposed honestly, the whole is about twenty
tasks, which is past this build's twelve-task sizing line and past
the spec's own fifteen-task split trigger (`spec:2344-2346`, stated
for 5b and applied here by the same reasoning). The cut and its
reasons are in "The split" below.

**Ten folds are conductor-owed**, each marked at its own site as
`**Conductor-owed fold:**` with the question and the applied answer, so
the conductor ratifies or reverses each in one edit:

1. **The split.** Cut 5c into 5c-i (this plan) and 5c-ii (the outline at
   the end), or run one pass of about twenty tasks? Applied as the
   split. ("The split.")
2. **Geoff's "security-fold Task 1".** Realized as Tasks 1 through 5,
   every one landing before any caller builds on the roster, the link
   URL, or the limiters, or as one task? Applied as five tasks. (R-5c2.)
3. **`streamer_username` on the write path.** Required from 5c by the
   `member` verb and the roster API, or optional until rung 6's pass 6d?
   Applied as required on the write path, optional at load. (R-5c6.)
4. **`streamer_username`'s character set.** Applied as `[A-Za-z0-9._-]`
   with uniqueness compared case-folded, and the SQL `LIKE` underscore
   question carried to rung 12 as a constraint. (R-5c6.)
5. **The CLI admin grant's channel.** A new `identity.ChannelCommandLine`
   that `AddMember` admits the admin flag on, or a separate store
   method with no channel? Applied as the new channel. (R-5c7.)
6. **The last-admin guard's shape at rung 5**, where no demote or remove
   path exists. Applied as a write-path invariant with one reachable
   path today and no exported demote method. (R-5c8.)
7. **`doctor`'s failing-key report** (review finding 11) against 5b-ii's
   ruling that dropped per-key rows. Applied as naming each failing key
   on the source's own line, with no new rows. (R-5c10.)
8. **The spec's admin checkbox on the add-member form** (`spec:1417`)
   against fork 1 and 5b-ii's channel refusal. Applied as dropping the
   checkbox: the roster API reads no admin field. (R-5c13.)
9. **The limiter at capacity** (review finding 3's fix). Fail closed by
   refusing a new key when no bucket has refilled, or keep least-recently-used
   eviction? Applied as fail closed, with the lockout residual recorded.
   (R-5c3.)
10. **The ceiling.** Applied as 7.0M for 5c-i with the flag at 5.6M, and
    9.5M provisional for 5c-ii. ("Token ceiling.")

**What this pass is.** Tasks 1 through 5 fold the security majors and
minors Geoff routed here, so every later caller builds on corrected code.
Task 6 lands the CLI admin grant, renames the channel sentinel, and adds
the last-admin guard. Task 7 builds the `dubplate-server member` verb
group, `add` and `list`. Task 8 builds the admin gate and the roster API
the add-member page will call. Task 9 wires the gate and the API into the
serve verb and proves the exit gate end to end through the real shell.
Task 10 is the close. The pass is Go only.

**What this pass is not.** No `web/` tree, no page, no theme, no font, no
embed path, no served header set, no CSP, no HSTS, no gate leg, no
server-file check, no Node requirement, and no rejoin-sentence
correction: those are 5c-ii's. No systemd unit, no `deploy.sh` edit, no
`doctor` probe row, no verify endpoint, no env template row, no
trusted-proxy setting, and no epoch-bump verb: those are 5d's. **No
deploy of any kind.** A task that believes it needs any of these stops
and reports.

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process",
> the pass-execute chain per task, sequential, dispatched in the four
> segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and
> never reads a diff, a test log, or a gate transcript inline. Subagents
> never dispatch agents, so `code-simplifier` and the pass-end
> `web-auth-security-reviewer` are the conductor's dispatch. The
> conductor owns `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the
> lane registry, and the contested-surface rows. **This pass runs in lane
> 1, the main tree.** If rung 6 launches as lane 2 during this pass, it
> takes a worktree, and the live-executor check is the lane registry,
> never `pgrep -f Projects/dubplate`.

---

## The split

> **Conductor-owed fold:** split 5c into 5c-i and 5c-ii, or run one pass?
> Applied as the split, cut between the Go-side work and the `web/` tree.

**The count.** The spec's fifteen deliverables are about eleven tasks
once grouped by file: the web scaffold with the theme data and binding,
the fonts, the embed path with the R4.5 rewrite, the branding and
operator stylesheets, the served header set, three pages, the `member`
verb, the gate leg, and a render-and-read task. Geoff's routing adds four
majors and nine minors, about five tasks. `docs/STATUS.md`'s carried
items add the CLI grant and the last-admin guard (one task), the
session-refresh call site and the roster API the page needs (two tasks),
and the mail copy (folded into a fold task). With Task 0 and two closes,
that is about twenty-one tasks.

**The cut point: all Go-side auth work in 5c-i, the `web/` tree in
5c-ii.** Three reasons:

- **Geoff's ruling orders it.** The security fold is "fixed before any
  page builds on the roster, the link URL, or the limiters". A pass
  boundary makes that ordering structural: 5c-ii's plan is drafted
  against a tree where the fold has already merged and been reviewed.
- **The risk changes at the cut.** 5c-i is cookie-authenticated privilege
  and concurrency logic, the shape R24's second
  `web-auth-security-reviewer` read exists for (`spec:577-586`). 5c-ii
  is a new stack (SvelteKit static, DaisyUI 5, a Node build in the gate)
  whose reviewers are `svelte-reviewer` and `daisyui-a11y-reviewer`.
  Reviewing both in one pass would bury one surface under the other,
  which is the argument the spec makes for its own 5a/5b cut
  (`spec:2334-2339`).
- **5c-i is complete and gradable on its own.** Unlike 5b-ii's rejected
  cut, which would have shipped mechanisms with no caller, 5c-i ships
  its own callers: the `member` verb and the roster API are real entry
  points a security reviewer can attack end to end through the shell.
  The page is the roster API's second client, not its first.

**What moves out of 5c-i on purpose.** The SvelteKit server-file check
moves to 5c-ii (R-5c16): with no `web/` tree it would pass on an empty
iteration, which is a check that cannot fail. Review finding 18 (HSTS)
joins 5c-ii's served header set, per Geoff's routing. The admin gate's
document half (the add-member page's HTML route answering 404 to a
non-admin) is 5c-ii's, built on the gate this pass exports.

---

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/superpowers/research/2026-09-23-rung-5b-ii-security-review.md`,
   in full.** Majors 1 through 4 and minors 7, 8, 10, 11, 12, 13, 15, 17,
   and 19 are Tasks 1 through 5's whole content. **Its line numbers are
   unreliable** (it cites `internal/auth/epoch.go:416-441`, and that file
   is 288 lines at `27cd7b5`; the Go tree is unchanged since the reviewed
   `08befea`), so every task resolves a finding by its symbol. **Its
   "Attacked and found sound" list (`:283` onward) is this pass's
   do-not-re-litigate record**: the token's canonicality, the nonce
   lifetime, single use under concurrency, the seven-step consumption
   order, consumption logging, the generic answer, cookie attributes, the
   SMTP guard, the origin check's placement and behavior, the absent
   bearer bypass, key derivation, the admin flag's read, the startup
   refusals, and the body caps. A task that reopens one without new
   evidence is spending the pass's budget on a settled question.
2. **The rung 5 spec, revision 2.1.** The sections this pass implements:
   "What rung 5 builds" (`:93-147`), fork 1 (`:176-201`), R1, R2, R3, R37
   (`:280-332`, `:729-757`), "The roster" (`:1103-1163`), "Rate limits,
   abuse, and logging" (`:1165-1189`), "The verify-identity seam"
   (`:1217-1245`), the add-member page's paragraph (`:1415-1426`), the
   2.0 hooks (`:1841-1859`), the prior-art table (`:1861-1907`), the
   threat model (`:1909-1980`), every erratum from `:1982` onward, the
   pass split (`:2321-2346`), the contested surfaces (`:2348-2380`), the
   band and ceiling (`:2382-2441`), the staleness pre-task
   (`:2443-2507`), and risk 14 (`:2595-2601`).
3. **`docs/superpowers/research/handoffs/draft-HISTORY-5b-ii.md`**, its
   "The 5c handoff" (`:588`) and "Also carried to 5c" (`:622`) paragraphs,
   which are binding on this plan, and its declared coverage gaps. **The
   conductor may revise these drafts before this plan launches**; Task 0
   reads the latest committed copy and reports any change.
4. **The rung 6 spec**, ruling 8 (`:96-100`), "Identity in the store"
   (`:767-820`), and 6d's criteria (`:1047-1080`). They state what rung 6
   expects 5c to have built: the CLI admin grant, the last-admin guard
   inside the write path, and a roster whose `streamer_username` rung 6's
   import will require.
5. **`docs/superpowers/specs/2026-09-11-configuration-homes-design.md:668-669`**,
   which routes the last-admin guard to rung 5's plan.

**Every relied-on library or stdlib behavior below is quoted from `go
doc`, the Go source, or a man page read on 2026-09-23 against
`go1.27.1 linux/amd64`.** A task that relies on a behavior this plan
does not quote confirms it the same way and reports what it read.

---

## Goal

Every surface a page will build on is corrected, and the two roster
writers the spec names exist. After this pass, the roster survives two
writer processes and a hand edit. The loader refuses any file that is
not the schema. A mailed link lands inside the owned prefix. No
single refused client can reset a victim's bucket, and one IPv6 /64
counts as one client. A fresh install gets its first admin from
`dubplate-server member add --admin`, and no roster write can leave the
group with no admin. A signed-in admin can list and add members through
a JSON API that answers 404 to everyone else and refreshes the session
it verifies.

---

## Owner touchpoints: none expected

`spec:2646-2652` names rung 5's expected touchpoints, and none is 5c-i's.
Two folds change what an adopter sees, and each is forced by a
ratified document rather than chosen for taste, so each is a conductor
ruling:

- **Fold 3, `streamer_username` required on the write path.** Rung 6
  ruling 8, ratified by Geoff on 2026-09-23, refuses an import entry
  with no `streamer_username`. A member the verb or the API adds without
  one today is a member rung 6's import refuses tomorrow.
- **Fold 8, no admin checkbox on the add-member form.** Fork 1's own text
  (`spec:188-192`, `:122-124`) makes the CLI verb "the only path that
  grants the admin flag to an account that has never signed in". A new
  member added through the page has never signed in.

**If the conductor reads either as a product fork**, it goes to Geoff in
one combined question before Task 2 dispatches (fold 3) or before Task 8
dispatches (fold 8), and never mid-segment.

**What would become an owner touchpoint.** A secret minted (none this
pass). A change to what the box runs (**nothing deploys**; the env
template still selects `magic-link` without its nine required keys). A
third module dependency (R26 allows two, both taken; `encoding/json/v2`
and `syscall` are stdlib). Any threshold default changing on a task's
own judgment.

---

## Token ceiling: 7.0M for 5c-i (re-derived from 5b-ii's actual)

> **Conductor-owed fold:** rule 7.0M for 5c-i with the flag at 5.6M, and
> carry 9.5M as 5c-ii's provisional figure? Applied as stated.

`spec:2395-2397` grants the re-derivation: "Each ceiling is re-derived
from the previous pass's actual report before the next plan is drafted."
The previous pass is 5b-ii, which closed at about **6.4M** against a
7.5M ceiling for twelve tasks plus Task 0 (draft STATUS, item 1).

**The per-line basis, from 5b-ii's actual.** About **0.46M per task** on
the chain, fix rounds and simplifier rounds included; **0.95M** of plan
preparation; **0.2M** for Task 0; **0.6M** for the close. These are the
conductor's figures from 5b-ii's report, given to this plan's author
rather than re-derived from HISTORY.

**The arithmetic for 5c-i.**

1. Nine implementation tasks at 0.46M is **4.14M**. Four of the nine run
   at `opus` against four of twelve at 5b-ii, a heavier mix, so add
   **0.10M**: **4.24M**.
2. Task 0 at **0.20M**, the close at **0.60M**, and this plan's
   preparation with its review lenses and fold at **0.95M**: **1.75M**.
3. The pass-end `web-auth-security-reviewer` read at **0.25M** and its
   fold at **0.20M**; two simplifier rounds with their chunked reviews at
   **0.30M**; the conductor's turns and the merge ritual at **0.30M**:
   **1.05M**.
4. The total is **7.04M**. **Rule 7.0M with the flag at 5.6M.**

**5c-ii's provisional ceiling: 9.5M, and it is an assumption, not a
measurement.** No `web/` task has run in this repo. Assume **0.60M per
task**, which is 5b-ii's rate plus about 30% for three costs Go tasks do
not pay: an `npm` install and a bundle build inside every gate once the
leg exists, a render-and-read per page, and a new stack's first-time
friction. Ten tasks at 0.60M is 6.0M. Add 0.2M for Task 0, 1.0M for plan
preparation, 0.6M for the close, 0.9M for the three pass-end reviewers
with their folds, 0.3M for simplifier rounds, 0.3M for the conductor, and
0.1M for the string deck's register gate: **9.4M, ruled provisionally at
9.5M** and re-derived from 5c-i's actual at this pass's close.

**Recorded against the rung reservation.** R23 set rung 5 at 14.5M and
5a's close moved the live baseline to **15.75M** (`spec:2042-2050`).
Taking actuals for closed passes and ceilings for open ones: 3.9M (5a)
plus 5.5M (5b-i) plus 6.4M (5b-ii) plus 7.0M (5c-i) plus 9.5M (5c-ii)
plus 3.0M (5d, a stale figure from the original planning rate) is
**35.3M against 15.75M**. The **19.55M** difference is recorded against
the reservation rather than taken from another pass. The conductor
re-derives 5d's figure at 5c-ii's close.

**Re-derive at the segment A boundary** from Tasks 1 through 3's actual
reports, the three `opus`-heavy fold tasks. **If the projection passes
7.5M**, the conductor names the cut at the Task 6 / Task 7 boundary: the
fold and the grant stay 5c-i, and the verb, the API, and the wiring move
to the front of 5c-ii.

**Checkpoint: STATUS at every segment boundary**, after Task 3, after
Task 6, after Task 9, and at the close. Every task reports its spend,
its mutation outcomes, and its measured gate wall time.

**Deliverable count: thirty-six across Tasks 1 through 10, plus Task 0's report.** Tasks 1, 2, 4, 6, 7,
and 8 carry four each; Tasks 3 and 9 carry three; Task 5 carries two;
the close carries four. No task passes four.

---

## Global Constraints

- **Lane 1, the main tree.** The lane registry entry is written before
  the first dispatch and is the live-executor check. If rung 6's lane 2
  launches during this pass, it takes a worktree, and the rows under
  "Contested surfaces" govern the shared files.
- **No deploy until 5d.** No task runs `scripts/deploy.sh`, not even
  `--check`, and no task ssh's to the box.
- **Package territory, stated as a wall.** This pass writes
  `internal/auth` (`roster.go`, `limit.go`, `handlers.go`, `token.go`,
  `session.go`, `nonce.go` edited; `admin.go` born; their tests),
  `internal/identity/identity.go` and its test (the normalizer's
  character refusal and one new channel constant),
  `cmd/dubplate-server` (`auth.go`, `doctor.go`, `serve.go`, `main.go`
  edited; `member.go` born; their tests), and `config/roster.example.json`.
  **It writes nothing under `internal/config`**, which is rung 6 pass
  6c's condition for running in lane 2 (rung 6 spec `:1023`). It writes
  nothing under `internal/serve`, `internal/notify`, `internal/engine`,
  `internal/harness`, `internal/atomicfile`, `web/`, `systemd/`,
  `scripts/`, `env/`, `Makefile`, `CLAUDE.md`, `corpus/`, `tools/`, or
  `_salvage/`, and no `docs/` file outside Task 10's conductor-owned
  drafts. A task that believes it needs one stops and reports.
- **No new module dependency.** `go.mod` and `go.sum` stay byte-identical.
  `encoding/json/v2` is stdlib in go1.27.1 and builds with no
  `GOEXPERIMENT` (measured 2026-09-23: a scratch program importing it
  ran with `go env GOEXPERIMENT` empty). File locking uses `syscall.Flock`
  from the standard library. The salvage lock package imports
  `golang.org/x/sys/unix`, a third dependency R26 does not budget, so it
  is adopted in reasoning only.
- **`internal/auth` reads no environment and renders no HTML**
  (`spec:1471-1479`). The handlers and the gate it adds are
  `http.Handler` values that `cmd/dubplate-server` mounts.
- **Assertion discipline at the repo's bar.** No silent-success fakes, no
  vacuous assertions, no expectation derived from the code under test.
  Every hard-coded expectation cites the spec, an RFC, a quoted doc, or
  the security review.
- **A security property is proved by behavior, never by structure**
  (risk 3, `spec:2526-2531`).
- **Every reproduction in the security review becomes a gate-resident
  test that fails before the fix and passes after**, and the task
  reports it observed failing first. The reviewer's scratch tests were
  left at
  `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/611daa97-f61c-4f5b-953c-d1bc92fd16c2/scratchpad/tree/`
  (`internal/auth/zz_review*_test.go`, `internal/serve/zz_review_test.go`,
  `cmd/dubplate-server/zz_review_test.go`). That directory is under
  `/tmp` and may be gone; Task 0 reports whether it exists. If it is
  gone, each reproduction is rebuilt from the review's own text. Nothing
  is copied from it verbatim without meeting the repo's test bar.
- **Fixtures are obviously not credentials** (`spec:2378-2380`).
  `.gitleaksignore` stays untouched.
- **No token, secret, cookie value, or URL password reaches a log record
  or an error string.** The consumption route logs its path and never
  `RawQuery`.
- **Addresses reach exactly two outputs.** The roster audit record names
  the actor and the address added, because `spec:1139-1143` requires it.
  `member list` prints the roster to its own stdout, because listing the
  roster is the verb's purpose. The roster API's GET answer carries the
  roster to a verified admin. No other log record and no error string
  carries an address.
- **Universality.** No estate hostname, path, email, or personal name
  anywhere. Every tunable is a threshold with a default.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx`
  first where a call can block, which after Task 1 includes every roster
  write. Errors lowercase and context-first. Sentinels where a caller
  branches, asserted through `errors.Is`. **The em dash is banned in Go
  comments.** No process citation (a pass, a task, a ruling id, a plan
  line) in a shipped comment: the citation leg is staged and warns today.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  through `cairn-run-gate`. Anything reported as evidence runs with
  `-count=1` and reports wall time. Specific files only. Commit footer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an `opus`
  task and the dispatch's footer otherwise.

---

## Privileged fields and their channels

CLAUDE.md's planning-miss rule asks that every privileged field name
every channel that may set it. These are this pass's.

| Field | Channels that may set it | Channels that may not, and what refuses them |
| --- | --- | --- |
| A member's `admin` flag | `dubplate-server member add --admin`, which calls `AddMember` with `identity.ChannelCommandLine` (Task 6). An operator's hand edit of the roster file, read at the next reload (Task 1); it leaves no audit record, which is a recorded residual. Rung 6's import, later, once. | The roster API: it reads no admin field and refuses a request carrying one (Task 8), and the store refuses the flag on every channel but `ChannelCommandLine` (Task 6). The session cookie: the seam reads the flag from the roster record, never the request (5b-ii, found sound). Request-link and consume-link: neither writes the roster. |
| The audit record's actor | The seam's verified, normalized address for the API (Task 8). The operating-system user for the verb (Task 7, `spec:1141-1143`). | Any request field. `AddMember` refuses an empty or malformed actor (Task 2). |
| The audit record's channel | The caller's constant: `ChannelCookie` from the API, `ChannelCommandLine` from the verb. | An invalid channel value: `AddMember` refuses `!channel.Valid()` (Task 2). |
| `streamer_username` | The verb and the API, both through `AddMember` (Tasks 7, 8). A hand edit, validated at reload. | Nothing else writes the roster. |
| The session cookie's refreshed value | The admin gate, only for a request whose `Sec-Fetch-Site` is exactly `same-origin` (Task 8). | A cross-site or header-less request: the gate refreshes nothing. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution
decides writes nothing about the resolution in a shipped comment (the
citation rule); the close maps each resolution to its landing commit by
reading the task reports.

### R-5c1: the split

Stated under "The split". Conductor-owed fold 1.

### R-5c2: the security fold is Tasks 1 through 5

> **Conductor-owed fold:** Geoff's ruling names "5c's security-fold Task
> 1". Is that one task, or the fold as a block that lands first? Applied
> as a block of five tasks, all before Task 6.

Thirteen findings across seven files are too many for one task under
this plan's four-deliverable cap. The ruling's operative clause is
ordering, "fixed before any page builds on the roster, the link URL, or
the limiters", and Tasks 1 through 5 all land before the verb (Task 7),
the API (Task 8), or 5c-ii exist. 5b-ii took the same reading of its own
fold (R-5b17, two tasks).

### R-5c3: the limiter fails closed at capacity

> **Conductor-owed fold:** at capacity, refuse a new key when no bucket
> has refilled (the review's fix for finding 3), or keep evicting the
> least recently used bucket? Applied as fail closed.

The review's fix has a cost the review does not name. A map full of
unrefilled buckets refuses every new key until one refills, so an
attacker who fills the per-client map locks out every client not
already in it for up to an hour. After finding 4's /64 keying, filling
10,000 per-client keys takes 10,000 /64s. One /48 holds 65,536 of them.
The alternative hands a victim a fresh per-address budget on every
flood. Both outcomes are bounded by the global link-mail ceiling, which
review finding 6 already records as a group-wide lockout lever. Geoff
ruled the finding into 5c with its fix, so this plan applies the fix,
records the lockout as a residual, and points the pass-end security read
at it by name.

### R-5c4: the roster decodes through `encoding/json/v2`, strictly

Review finding 7 lists six loose variants. `encoding/json` v1 cannot
refuse two of them: `go doc encoding/json.Unmarshal` states it matches
keys "preferring an exact match but also accepting a case-insensitive
match", and `DisallowUnknownFields` rejects only keys that match no
field, so `"ADMIN"` still binds to `admin`. `go doc encoding/json/v2`
states "By default, unmarshaling uses case-sensitive matching" and "v2
rejects duplicate names", and `json.RejectUnknownMembers(true)` rejects
unknown members. Measured 2026-09-23 against go1.27.1:
`{"ADMIN":true}` fails with `unknown object member name "ADMIN"`, and
`{"admin":false,"admin":true}` fails with
`jsontext: duplicate object member name "admin"`.

v2 still decodes `null`, `{"schema_version":1}`, and
`{"schema_version":1,"members":null}` with a nil error and a nil
`Members` slice, while `"members":[]` yields a non-nil empty slice
(measured the same day). So the loader also requires `schema_version`
to equal 1 and `Members` to be non-nil after decoding. Together those
two checks refuse all three. The writer keeps encoding `"members": []`
for an empty roster, as `CreateRoster` does today.

### R-5c5: the store reloads on a changed file and fails closed on a broken one

`Lookup` and `List` stat the roster file on every call. When the device
and inode (`os.SameFile`, which `go doc os.SameFile` defines on Unix as
"the device and inode fields of the two underlying structures are
identical"), the size, or the modification time differ from the copy in
memory, they reload under the in-process mutex. Reads need no
cross-process lock: `atomicfile.WriteStream` replaces the file by rename,
so a reader opens either the whole old file or the whole new one.

A reload that finds the file absent, unreadable, unparseable, or invalid
returns that error. It does not keep serving the last good copy, because
the last good copy is exactly the member a hand-edit offboarding just
removed, and R2 already rules fail-closed. The request-link handler
tells a store failure apart from `ErrMemberNotFound`. Under the kind
answer, a store failure answers `failed` and never `unknown`, so a broken
roster never hands a known member the unknown-address copy. Under the
generic answer, the response bytes do not change and the log records the
failure.

### R-5c6: `streamer_username` is required on the write path and restricted in shape

> **Conductor-owed fold 3:** require `streamer_username` on the write
> path from 5c, or leave it optional until 6d? Applied as required on the
> write path and optional at load.
>
> **Conductor-owed fold 4:** the character set. Applied as
> `[A-Za-z0-9._-]`, with uniqueness compared case-folded.

Rung 6 ruling 8 (Geoff, 2026-09-23) refuses an import entry with no
`streamer_username`, and 6d criterion 6 has "the page and the verb
refuse a missing or duplicate `streamer_username` by name". A member
added without one at 5c is a member 6d refuses. The loader keeps the
field optional, per R3 and B9, so a hand-written roster that predates
this pass still loads. The close owes R3 an erratum.

Review finding 15 recommends `[a-z0-9._-]`, case-folded comparison, or
both. Lowercase-only would refuse a streamer account that already carries
an uppercase letter, and this plan cannot verify what the estate's
Navidrome accounts hold. So the character set admits both cases, and
uniqueness compares with `strings.EqualFold`, which alone closes the
`Alice`/`alice` collision. `%`, spaces, and every other character are
refused. `_` stays allowed. Whether Navidrome resolves a username with
SQL `LIKE`, where `_` is a wildcard, is the reviewer's "recollection" and
is carried to rung 12 as a constraint to verify, not settled here.

### R-5c7: the CLI admin grant is `identity.ChannelCommandLine`

> **Conductor-owed fold 5:** a new channel constant, or a separate store
> method that takes no channel? Applied as the new constant.

`AddMember` today refuses `Admin` on every channel, and its doc comment
says "The admin grant arrives together with that verb". The audit record
already takes a channel, and `spec:1141-1143` wants the CLI's record to
be "the same record". So the command line becomes one more
`identity.Channel`, `ChannelCommandLine`, whose `String` is
`command-line` and which `Valid` accepts. `AddMember` admits the admin
flag on that channel alone. `ErrRosterAdminChannelForbidden` is renamed
`ErrAdminGrantForbidden`, since after this change the refusal is about
who may grant, not about channels in general. No verification path
produces `ChannelCommandLine`: the seam sets `ChannelCookie`, and a test
proves it (Task 6). A later forward-auth front sets `ChannelHeader`.

### R-5c8: the last-admin guard is a write-path invariant

> **Conductor-owed fold 6:** rung 5 has no demote or remove path, so what
> does the guard guard? Applied as: every roster the store writes names at
> least one admin, checked under both locks against the file as re-read
> from disk, with no demote method exported ahead of a caller.

Configuration homes (`:668-669`) routes "a last-admin lockout guard,
atomic against concurrent self-demotion" to rung 5's plan, and rung 6's
spec (`:816-818`) expects "5c's last-admin guard" to run inside 6d's
write transaction. Geoff's 2026-09-12 refinement forbids exporting
anything ahead of its caller, so this pass builds no demote method.

The guard still has one reachable path today, and Task 6 proves it
there. A fresh roster from `CreateRoster` has no members, and a hand edit
can remove every admin. Adding a non-admin to either would write a
roster with no admin, which the guard refuses by its own sentinel. The
first member of a fresh install must therefore be an admin, which is the
runbook's ordering anyway. The invariant lives in the one internal
read-modify-write function every roster mutation routes through, so a
later demote or remove method inherits it by construction. `doctor`'s
roster row reports a roster that names no admin as failing, and names
the recovery (`member add --admin`).

### R-5c9: the public URL is a bare origin, and a link is origin plus prefix plus `/login/link`

Review finding 2's fix offers two refusals: a public URL carrying a path,
or one whose path is not the prefix. This plan takes the first.
`DUBPLATE_PUBLIC_URL`'s path must be empty or `/`, and `LinkURL` builds
`origin + prefix + "/login/link"`. One rule is simpler to state in an
error, and the prefix already expresses where dubplate lives. **5c-ii
places the consumption interstitial's route at `login/link` under the
prefix**, and its outline says so. The shell-level test this pass adds
proves a minted link's path is owned by the prefix and reaches a
registered handler rather than the fallthrough.

### R-5c10: `doctor` names each failing source key on the source's own line

> **Conductor-owed fold 7:** review finding 11 asks for one `<KEY>
> invalid` row per failing key; 5b-ii's fold 7 dropped per-key rows to
> keep one line per source. Applied as naming the keys on the one line.

The finding's substance is that no line names the offending key. The
source's own line can carry the names, for example `magic-link
unsatisfied: DUBPLATE_SESSION_SECRET_PREVIOUS invalid`, with key names
only and never a value. That keeps the line-count derivation 5b-ii's
tests assert (`TestDoctorCmd_LineCountFollowsRegistry`, `cmd/dubplate-server/doctor_test.go`). If the
conductor reverses to per-key rows, Task 5 extends those tests in the
same commit, and says so.

### R-5c11: session lifetime above the cap refuses at load and clamps at mint

Finding 12 offers "require lifetime <= cap, or clamp at mint". This plan
takes both. Load refuses a lifetime above the cap and names both keys,
because a configuration error belongs at configuration time (R38's
reasoning). `MintSession` clamps to the cap, which matches the clamp 5b-ii
already put on `RefreshSession`, so no cookie ever carries a `Max-Age`
past the cap. The nonce retention threshold gains a ceiling of **240h**,
ten times its 24h default, on the same ten-times rule the rate
thresholds' ceilings already follow (`cmd/dubplate-server/auth.go`,
the comment above `linkRequestsPerAddressCeiling`). `DUBPLATE_SMTP_PORT`
must parse to 1 through 65535.

### R-5c12: the admin gate answers 404 to everyone who is not a verified admin

The gate verifies the session through the seam, then reads `Admin` from
the verified identity (which the seam read from the roster). An absent
cookie, a cookie that fails verification, a revoked session, a member
removed from the roster, a roster store failure, and a verified
non-admin all answer the same 404 (`spec:1418-1420`, risk 14). The page's
existence is not a fact any of them learns. The reason reaches the
server log alone, and the log never carries the cookie value.

The gate is the session-refresh call site `docs/STATUS.md` carries. It
passes `sameOrigin` as true only when `Sec-Fetch-Site` is exactly
`same-origin`. The seam's own contract leaves that derivation to the
caller (`internal/auth/seam.go`, `Seam.VerifyIdentity`'s comment). The
header's other values (`same-site`, `cross-site`, `none`, per the MDN
reference `go doc net/http.CrossOriginProtection` links) and its absence
refresh nothing. The origin check that already wraps the shell does not
cover a GET: `go doc net/http.CrossOriginProtection` states "The GET,
HEAD, and OPTIONS methods are safe methods and are always allowed". So
the gate cannot infer same-origin from having been reached. On a
refresh, the gate sets the refreshed cookie through `SessionCookie`,
with the attributes 5b-ii's review found sound.

### R-5c13: the roster API reads no admin field

> **Conductor-owed fold 8:** `spec:1417` gives the add-member form "an
> admin checkbox"; fork 1 makes the CLI verb the only path to an admin
> for an account that has never signed in, and 5b-ii's write path refuses
> the flag on the cookie channel. Applied as no checkbox, and the API
> refusing any request that carries an `admin` field.

A checkbox the server always refuses is a control that cannot work. The
API refuses a request carrying `admin`, whatever its value, with its own
outcome and no write. The store's channel refusal stays as the second net.
The page lists who is an admin (`spec:1424`) and offers no grant. The
close owes `spec:1417` an erratum.

### R-5c14: the `member` verb resolves the roster through `config.Load` and refuses a foreign owner

The verb loads configuration the same way `serve` does, reads
`DUBPLATE_ROSTER_FILE` from the selected source, and refuses, naming the
selector, when the selected identity source requires no roster. The
runbook's ordering (R37: "install, render the environment, run the verb,
start the unit") means the environment is complete when the verb runs.

The verb refuses to write when its effective user id differs from the
owner of the roster file's containing directory, and names both ids. A
`member add --admin` run as root on the estate would otherwise create a
root-owned 0600 roster that the `music` unit cannot read, which makes
the server fail closed at its next start under R2. The owner comes from
`os.Stat(dir).Sys().(*syscall.Stat_t).Uid`; Task 7 quotes `go doc
os.FileInfo` for `Sys` before relying on it.

### R-5c15: the verb's audit record goes to its own stderr

`AddMember` logs through the logger it is handed. The verb hands it a
`slog` text handler on stderr, so the record reaches the invoking
terminal, or the journal when the verb runs under a unit. It does not
reach the serve unit's journal. That is a declared coverage gap, carried
to rung 6, whose store makes a member-change record durable or not
(rung 6 spec `:818-820` keeps the log record). The record's `client_ip`
attribute reads `local` for the command-line channel, so the record keeps
one shape.

### R-5c16: the SvelteKit server-file check moves to 5c-ii

A check over a `web/` tree that does not exist passes on an empty
iteration, and CLAUDE.md's rule is that a check either gates or does not
exist. It lands in 5c-ii's gate-leg task with the tree it checks.

### R-5c17: 5c-i leaves `internal/config` untouched

Rung 6's 6c may run in lane 2 only if 5c and 5d leave `internal/config`
untouched (rung 6 spec `:1023`). Every change this pass makes to
configuration goes through `cmd/dubplate-server/auth.go`'s own
validation, on rung 1b's registry contract.

---

## The threat model, as the in-pass reviewer's rubric

The spec's threat model (`spec:1909-1980`) governs. The claims this pass
must make true, each with the task that proves it:

- **Only a request carrying a session for a member whose roster record
  holds the admin flag can change the roster through the web** (Tasks 8,
  9). The flag is read from the roster, never the request. No web path
  sets it (Tasks 6, 8).
- **A roster change that fails validation leaves the file
  byte-identical, and every roster change that succeeds writes one audit
  record naming the actor and the channel** (Tasks 1, 2, 8, 9), across
  two writer processes (Task 1).
- **A member removed by hand stops resolving at the running server's
  next request** (Task 1), which closes the gap in claim 8 that finding 1
  opened.
- **A state-changing cookie-authenticated request without a same-origin
  signal is refused** (Task 9, through the real shell).
- **Neither a single address nor a single client can drive unbounded
  outbound mail, and one IPv6 /64 is one client** (Task 3).
- **A token never reaches a log record or a process outside dubplate**
  (Task 4: the link lands inside the owned prefix).
- **No operator secret reaches a log** (Task 4: the public URL's password
  is never echoed).

**The residuals this pass adds, stated rather than hidden:** a full
limiter map refuses new keys (R-5c3); a hand edit of the roster writes no
audit record; the CLI's audit record reaches the invoking terminal
rather than the service journal (R-5c15); an admin's shared device can
add a member, as the spec already records.

---

## Review focus

The five input classes most likely to bite a person that no task's
happy path exercises, each pinned by a test in the task that owns it.

1. **Two writers at once from two processes**, the web page and the CLI
   verb, while a third party hand-edits the file. Expected: nothing lost,
   nothing resurrected. Pinned in Task 1.
2. **A roster file an operator broke by hand while `serve` runs.**
   Expected: logins refuse with a named error rather than serving the
   stale copy or answering "unknown". Pinned in Task 1.
3. **An address or display name carrying an invisible character** (a
   zero-width space, a right-to-left override, a no-break space).
   Expected: refused at every entry point. Pinned in Task 2.
4. **A second `member add --admin` on an install that already has a
   roster.** Expected: the member is added and the file is never
   recreated. Pinned in Task 7.
5. **A signed-in member without the admin flag probing the roster API**
   with GET, POST, and a crafted `admin` field. Expected: 404 for all, no
   byte on disk changes. Pinned in Tasks 8 and 9.

---

## Task 0: The staleness pre-task

Model: **`opus`**, fresh context, dispatched by the conductor before any
implementer. Reads the spec, the security review, the 5b-ii close
drafts, and this plan against the tree as shipped. **Its output is
blockers with evidence, not fixes.** It writes no code and edits no
file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports that
HEAD.** This plan's pointers were resolved at **`27cd7b5`**. Grade a
moved pointer by content: a pointer off by fewer than about ten lines
whose content still matches is a note, and a pointer naming content that
no longer exists is a blocker.

**The first four items are blockers: a "no" stops the launch.**

1. **Is the lane registry consistent with a lane 1 launch in the main
   tree, and is the tree free?** Report `git status`, `git worktree
   list`, and the registry rows. Report any uncommitted change under
   `internal/`, `cmd/`, or `go.mod` that no conductor commit explains.
   *(Expected: 5b-ii merged, lane 1 reopening on 5c, lane 2 closed or
   in a worktree.)*
2. **Has 5b-ii merged, and are the close drafts folded or still
   drafts?** Report the merge commit, and the line counts and last
   commits of the two draft files and of `docs/HISTORY.md` and
   `docs/STATUS.md`. **Report every change to "The 5c handoff" or "Also
   carried to 5c" since `27cd7b5`.**
3. **Do the thirteen routed findings still reproduce at HEAD?** For each
   of majors 1 through 4 and minors 7, 8, 10, 11, 12, 13, 15, 17, and 19,
   report the symbol and the current line it lives at. *(Expected, all
   live: `(*Roster).AddMember` rewrites from `r.members` with no reload;
   `(*Roster).Lookup` never stats the file; `LinkURL` joins `linkPath`
   with no prefix; `(*Limiter).AllowLinkRequest` charges `perAddress`
   before the client check; `(*keyedLimiter).allow` evicts the least
   recently used bucket unconditionally; the per-client key is the raw
   `ClientIP` string; `parseRoster` uses `json.Unmarshal`;
   `identity.NormalizeAddress` refuses only `r <= ' ' || r == 0x7F`;
   `validatePublicURL` quotes the value with `%q`; `identitySourceStatus`
   names no key; `newMailer` parses the port with `strconv.Atoi` alone;
   `AddMember` accepts an empty actor and `Channel(0)`; the
   `streamer_username` uniqueness check compares with `==`; the handler
   normalizes before any limiter charge; the nonce sweep's mtime fallback
   covers unparseable records only.)* **A finding a later commit already
   fixed is a blocker for the task that was to fix it**, because that
   task's failing-first test would pass before the change.
4. **Is `internal/config` untouched since 5b-i's merge, and does rung 6's
   lane 2 have an executor in any tree?** Report the commit list for
   `internal/config` since `1d4ed29`, and the registry's lane 2 row.

**Notes (not blockers):**

5. Does the reviewer's scratch tree still exist at the path Global
   Constraints names? Report which files are present.
6. Confirm `encoding/json/v2` builds under the repo's own toolchain with
   no `GOEXPERIMENT`: report `go env GOEXPERIMENT` and whether a scratch
   test in `internal/auth` importing it compiles (and delete it).
7. Report `newServeShell`'s current mounts and the three route-name
   constants, the shape Task 9 extends.
8. Report `doctor_test.go`'s line-count test by name and the rows it
   derives from, the property R-5c10 preserves.
9. Report `newRootCmd`'s `AddCommand` block, which Task 7 extends.
10. Report every caller of `(*Roster).AddMember`, `(*Roster).List`, and
    `ErrRosterAdminChannelForbidden` in the module, since Tasks 1, 2, and
    6 change all three.
11. Report the spec lines this plan cites for fork 1's "only path" text
    (`:122-124`, `:188-192`) and the admin checkbox (`:1417`), the two
    passages R-5c13 turns on.

**Deliverable count: 1** (the report).

---

## Task 1: The roster across processes, and the strict loader

Model: **`opus`**. Depends on Task 0.

Review major 1 and minor 7. The roster's own correctness, fixed first,
because every later task reads or writes it.

**Files:**
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`
  (the request-link handler's store-failure branch only, R-5c5)
- Modify: `internal/auth/seam.go` only if `Lookup`'s changed error set
  needs a caller change (it should not: the seam already refuses any
  non-nil error)

**Interfaces:**
- Consumes: `internal/atomicfile.WriteStream`; `syscall.Flock`.
- Produces: `(*Roster).AddMember(ctx context.Context, candidate Member,
  actor string, channel identity.Channel, clientIP string, logger
  *slog.Logger) error`, `ctx` first because the write can now wait on
  another process. `(*Roster).List() ([]Member, error)`, which gains an
  error because a reload can fail. `(*Roster).Lookup` keeps its
  signature and gains the reload errors. A new sentinel `ErrRosterBusy`
  for a lock wait that outlives `ctx`. One unexported locked
  read-modify-write function that `AddMember` and `CreateRoster` route
  through, which Task 6's guard lives in.

**Outcomes:**

- **Every write takes an exclusive `flock` on the roster file's
  containing directory**, re-reads and validates the file from disk
  under that lock, merges the candidate into the disk copy rather than
  the memory copy, writes through `atomicfile.WriteStream`, re-reads, and
  only then releases. The in-process mutex still serializes writers
  inside one process. Why the directory: an atomic rename replaces the
  file's inode, so a lock on the file itself would not exclude a writer
  that opened the new inode. Locking the directory creates no file in
  the leaf. `man 2 flock` (read 2026-09-23): "If a process uses open(2)
  (or similar) to obtain more than one file descriptor for the same file,
  these file descriptors are treated independently by flock()", so two
  `Roster` values in one test process exclude each other exactly as two
  processes do. The same page: "A shared or exclusive lock can be placed
  on a file regardless of the mode in which the file was opened", so a
  read-only open of the directory suffices.
- **The lock wait is bounded by `ctx`.** The write polls `LOCK_EX|LOCK_NB`
  at a short fixed interval (the salvage lock package's shape, adopted in
  reasoning) and returns `ErrRosterBusy` when `ctx` ends first. It never
  blocks without bound: `flock(2)` states "A call to flock() may block if
  an incompatible lock is held by another process", and a blocking call
  cannot observe `ctx`.
- **`CreateRoster` runs its existence check and its write under the same
  directory lock.** Today it stats and then writes, so two concurrent
  bootstraps can both see "absent" and the second overwrites the first.
  Under the lock, the second sees the first's file and returns
  `ErrRosterExists`.
- **`Lookup` and `List` reload when the file changed** (R-5c5): device
  and inode through `os.SameFile`, size, and modification time. An
  unchanged file costs one `os.Stat`. After a successful write the stored
  stat is the written file's, so a writer never reloads its own write.
- **A reload that fails returns the failure** (absent, unreadable,
  unparseable, invalid), each wrapping the existing sentinel with the
  path. The request-link handler answers `failed` under the kind answer
  when `Lookup` returns anything but `ErrMemberNotFound`, and keeps its
  generic-answer bytes unchanged.
- **The loader is strict** (R-5c4): `encoding/json/v2` with
  `json.RejectUnknownMembers(true)`, case-sensitive by default and
  refusing duplicate names by default; `schema_version` must equal 1;
  `members` must be present and non-null.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Load | file absent | `ErrRosterMissing`, naming the path |
| Load | file unreadable | `ErrRosterUnreadable` |
| Load | zero-byte file, `null`, `{}`, `{"schema_version":1}`, `"members":null` | `ErrRosterUnparseable`, naming the path and which rule failed |
| Load | `schema_version` 0, 2, or 99 | `ErrRosterUnparseable`, naming the version rule |
| Load | an unknown member key, a case-variant key (`"ADMIN"`), a duplicate key | `ErrRosterUnparseable` |
| Load | `"members":[]` with version 1 | loads, empty |
| Reload | stat unchanged | cached copy, no read |
| Reload | changed, valid | reloaded; a hand-removed member stops resolving |
| Reload | changed, invalid or unparseable | the load error; `Lookup` refuses |
| Reload | file removed | `ErrRosterMissing` |
| Reload | replaced with identical bytes (new inode) | reloaded, same answers |
| Write lock | free | acquired |
| Write lock | held by another `Roster` value past `ctx` | `ErrRosterBusy`, file unchanged |
| Create | absent | created at 0600 |
| Create | present, or created concurrently by a second caller | `ErrRosterExists` for the loser, the winner's file intact |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The review's cross-process reproduction is a gate-resident test: two
  `Roster` values over one file, a CLI-side add, a hand removal by
  rewriting the file directly, then a serve-side add. The final file
  carries the CLI's member and the web's member and not the removed one,
  and the serve-side value stops resolving the removed member before its
  own add. **Fixture state:** both values opened before either write, so
  the serve-side value's memory copy is stale.
- A concurrent pair of `CreateRoster` calls, run with `-race`, leaves
  exactly one file whose contents are the winner's, and exactly one
  `ErrRosterExists`.
- A held-lock test holds the directory lock from a second `Roster`
  value's lock path (or a raw `flock` on the directory in the test) and
  asserts `ErrRosterBusy` under a short `ctx` and a byte-identical file.
- Each loose variant in finding 7 is a table row asserting refusal by
  sentinel, including `"ADMIN": true` and the duplicate `"admin"` key.
- The request-link test presents a known address against a roster file
  made unparseable after open, under the kind answer, and asserts
  `failed` and not `unknown`, with no admin contact in the body; under
  the generic answer it asserts the same bytes as a normal answer.
- No error string names an address. Doc comments carry no em dash.
  `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Merge from memory rather than from disk under the lock → the
   cross-process row resurrects the removed member.
2. Drop the reload from `Lookup` → the removed member still resolves.
3. Keep the last good copy on a failed reload → the broken-file row
   resolves the member.
4. Lock the roster file instead of its directory → the cross-process
   row loses a member (report whether it fires; if the test's timing
   cannot make it fire, say so and let the reviewer read it).
5. Stat-then-write in `CreateRoster` outside the lock → the concurrent
   create row can produce two successes (report the observed rate over
   repeated runs).
6. Decode with v1 `encoding/json` → the case-variant and duplicate-key
   rows load.
7. Drop the non-null `members` check → the `"members":null` row loads.
8. Answer `unknown` on any `Lookup` error → the store-failure row
   returns the admin contact.

**2.0 hook: extends hook 5's write half** (`spec:1848`): the one
read-modify-write function is where rung 6 mints the surrogate id inside
its transaction, and the stat-and-reload read side is what rung 6's store
replaces with a query.

**Prior art:** adopts the salvage lock package's poll-under-`ctx` shape
(`_salvage/internal/lock/lock.go`, `AcquireBlocking`) in reasoning, with
stdlib `syscall.Flock` in place of `golang.org/x/sys/unix` (R26). Adopts
`encoding/json/v2` (stdlib, Go team, measured above). Declines a file
watcher (`fsnotify`): a module dependency R26 does not budget, and a
stat per lookup is cheap at group scale, as the review itself notes.

**Independent: no.** Every later roster task builds on it.

**Deliverable count: 4** (the cross-process write with its lock, the
locked create, the reload with its store-failure answer, the strict
loader).

---

## Task 2: Validation tightening, and one sentinel per refusal

Model: `sonnet`. Depends on Task 1.

Review minors 8, 13, and 15, plus fold 3's write-path requirement, plus
the per-refusal sentinels the roster API maps to specific outcomes
(`spec:1421-1424`).

**Files:**
- Modify: `internal/identity/identity.go`, `internal/identity/identity_test.go`
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `config/roster.example.json` (its second member gains a
  `streamer_username`, so the reference file models the rule the write
  path enforces)

**Interfaces:**
- Consumes: Task 1's locked write function.
- Produces: exported sentinels, each wrapping `ErrRosterRecordInvalid`
  so every existing `errors.Is(err, ErrRosterRecordInvalid)` still
  holds: `ErrDuplicateAddress`, `ErrDuplicateStreamerUsername`,
  `ErrInvalidMemberAddress`, `ErrInvalidDisplayName`,
  `ErrInvalidStreamerUsername`, `ErrMissingStreamerUsername`, and
  `ErrInvalidAuditActor` (for the actor and channel refusal). Task 8
  maps each to an outcome.

**Outcomes:**

- **`identity.NormalizeAddress` refuses any rune for which
  `unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs,
  unicode.Zl, unicode.Zp)` holds** (the review's own fix for finding 8).
  `go doc unicode` lists `Cf` as "the set of Unicode characters in
  category Cf (Other, format)". The existing ASCII refusal stays as the
  special case it already is. The normalizer stays a fixed point.
- **Display names refuse the same set.** Today they refuse
  `unicode.IsControl` alone, which misses `Cf`.
- **`streamer_username` is restricted to `[A-Za-z0-9._-]` and unique
  under `strings.EqualFold`**, at load, in `doctor`'s collision check,
  and at write time inside the lock (R-5c6).
- **The write path requires `streamer_username`** (fold 3). The loader
  does not: a roster without one still loads.
- **`AddMember` refuses before any write** when `!channel.Valid()`, and
  when the actor does not meet its channel's rule. For a network channel
  the actor must normalize through `identity.NormalizeAddress`. For
  `ChannelCommandLine` (Task 6), the actor must be non-empty and carry no
  rune from the set above. Until Task 6 lands, only the network rule is
  reachable. Minor 13's fix as written ("refuse an actor that does not
  normalize") would refuse the operating-system user `spec:1141-1143`
  puts in the CLI's actor field, hence the per-channel rule.
- **Each refusal wraps its own sentinel**, and no error string names the
  refused value.

**Acceptance criteria:**

- The review's normalizer corpus is a table: U+0085, U+009B, U+202E,
  U+00A0, and U+200B inside an address each refuse. A zero-width
  variant of an existing member's address is refused, not added as a
  distinct member. **Fixture state:** a roster already holding the plain
  address.
- The Kelvin sign case the review mentions (a domain lowercasing to a
  different ASCII domain) is a table row whose expected outcome the
  implementer derives from `strings.ToLower`'s documented behavior and
  cites; if the normalizer accepts it, the task reports that as a
  finding rather than silently fixing or ignoring it.
- `Alice` against an existing `alice` is refused at write time and at
  load. `a%`, `a b`, and a non-ASCII name are refused. `a_b` and
  `A.b-c` are accepted.
- A write with no `streamer_username` is refused by
  `ErrMissingStreamerUsername`; a roster file with a member lacking one
  still loads.
- An empty actor, a malformed actor on the cookie channel, and
  `Channel(0)` are each refused by `ErrInvalidAuditActor` with the file
  byte-identical.
- Every sentinel satisfies `errors.Is(err, ErrRosterRecordInvalid)`.
- `config/roster.example.json` still loads through `OpenRoster` in the
  existing example-file test. `bash scripts/check.sh` green.

**Named mutations:**

1. Drop `unicode.Cf` from the refusal set → the U+202E and U+200B rows
   admit.
2. Compare `streamer_username` with `==` → the `Alice`/`alice` row
   admits.
3. Make the write path's requirement a load-time requirement → the
   legacy-roster load row fails.
4. Skip the channel validity check → the `Channel(0)` row writes.
5. Apply the address rule to the command-line actor → Task 6's
   operating-system-user row fails (report this one at Task 6 if it
   cannot be exercised here).
6. Wrap a new sentinel without `ErrRosterRecordInvalid` → the
   `errors.Is` row fails.

**2.0 hook: lands hook 5's "required" half early on the write path**
(S20 hook 5: "`streamer_username` becomes required, unique, and
stable"). Rung 6's 6d makes it required at load.

**Prior art:** none consulted beyond the stdlib `unicode` tables. The
refusal set is the review's own fix.

**Independent: yes, of Task 3 by files** (`identity.go`, `roster.go`,
the example file against `limit.go`, `handlers.go`). Both follow Task 1.

**Deliverable count: 4** (the invisible-character refusal across
addresses and display names, the principal's shape and case-folded
uniqueness, the write-path requirement, the actor and channel refusal
with the per-refusal sentinels).

---

## Task 3: The limiter fold

Model: **`opus`**. Depends on Task 1 (both touch `handlers.go`).

Review majors 3 and 4, and minor 17.

**Files:**
- Modify: `internal/auth/limit.go`, `internal/auth/limit_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`

**Interfaces:**
- Consumes: `net/netip`; `golang.org/x/time/rate` (already a dependency).
- Produces: `(*Limiter).AllowLinkRequest(now time.Time, address, client
  string) error` with a changed charging order (below). The handler calls
  it before normalizing.

**Outcomes:**

- **The per-client key is charged first, and the per-address key is
  charged only if the client was admitted** (finding 3's first fix). A
  client already refused never touches the per-address map.
- **An address that fails to normalize is charged against the client**
  (finding 17): the limiter charges the client, then normalizes, and an
  invalid address returns its `identity.ErrInvalidAddress`-wrapped error
  having spent the client's token. The handler's order becomes: cap and
  parse the body, charge through `AllowLinkRequest`, then look up the
  roster. The handler still answers `invalid_address` for a bad address
  and still logs no address.
- **At capacity, a keyed map evicts only a bucket that has fully
  refilled, and refuses the new key when none has** (finding 3's second
  fix, R-5c3). "Fully refilled" is `bucket.TokensAt(now) >=
  float64(bucket.Burst())`: `go doc golang.org/x/time/rate.Limiter.TokensAt`
  states it "returns the number of tokens available at time t", and
  `Burst` returns "the maximum burst size". The refusal wraps
  `ErrRateLimited` with its own reason text, so the handler answers 429.
  The scan for a refilled bucket starts at the least recently used end
  and stops at the first match. The doc comment states the scan's worst
  case against `maxLimiterKeys`.
- **Per-client keys are derived in the limiter, never in `ClientIP`**
  (finding 4's fix; `ClientIP` also feeds logs). `netip.ParseAddr`, then
  `Unmap`, then IPv4 keys on the address and IPv6 on `Prefix(64)`.
  `go doc net/netip.Addr.Unmap`: "if ip is an IPv6 address wrapping an
  IPv4 address, it returns the wrapped IPv4 address". `go doc
  net/netip.Addr.Prefix`: "Prefix keeps only the top b bits of IP". The
  Go source's `PrefixFrom` builds the prefix from `ip.withoutZone()`
  (`net/netip/netip.go`, read 2026-09-23), so a zoned link-local address
  keys without its zone. **A value that does not parse keys on one fixed
  shared bucket name**, so garbage cannot mint distinct keys. Today the
  peer address always parses; the case exists for 5d's trusted-proxy mode
  (finding 9, 5d's).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Keyed map | key present | charge it |
| Keyed map | below capacity, key absent | create it full, charge it |
| Keyed map | at capacity, a refilled bucket exists | evict the least recently used refilled bucket, create, charge |
| Keyed map | at capacity, none refilled | refuse, `ErrRateLimited` ("limiter full"), nothing evicted |
| Client key | IPv4 `192.0.2.7` | `192.0.2.7` |
| Client key | IPv4-mapped `::ffff:192.0.2.7` | `192.0.2.7` |
| Client key | IPv6 `2001:db8:1:2:3:4:5:6` | `2001:db8:1:2::/64` |
| Client key | `fe80::1%eth0` | `fe80::/64` |
| Client key | empty or unparseable | the shared invalid-key bucket |
| Link request | client refused | 429; the per-address map unchanged |
| Link request | client admitted, address invalid | `invalid_address`; the client charged |
| Link request | client admitted, address spent | 429 |

**Acceptance criteria:**

- The review's eviction reproduction is a test: exhaust a victim
  address at instant T, flood 10,000 distinct addresses from one client
  at the same T, and the victim's next request at T is still refused.
  **Fixture state:** the victim's bucket spent before the flood, same
  instant, so no refill can explain a pass.
- A second test fills a map with unrefilled buckets from distinct
  clients (not one), so the per-client order alone cannot explain the
  refusal, and asserts a new key is refused and the size stays at
  capacity.
- A refill test advances the clock past a full interval and asserts the
  oldest refilled bucket is the one evicted.
- The review's IPv6 reproduction is a test: 100 addresses inside one /64
  against a consumption limit of 1 admit exactly 1.
- The per-address map's size is asserted unchanged after a refused
  client's requests.
- `bash scripts/check.sh` green; the limiter tests run with `-race`.

**Named mutations:**

1. Charge the address before the client → the refused-client row grows
   the per-address map.
2. Evict the least recently used bucket unconditionally → the
   same-instant victim row admits.
3. Treat any bucket as refilled → the distinct-clients row admits.
4. Key IPv6 on the full address → the /64 row admits 100.
5. Skip `Unmap` → the IPv4-mapped row keys separately from its IPv4
   form.
6. Normalize before charging the client → the invalid-address row
   leaves the client uncharged.

**2.0 hook: neutral.** S20's `golang.org/x/time/rate` verdict is for the
same shape; nothing here is schema or driver.

**Prior art:** keeps `golang.org/x/time/rate` (adopted at 5b-ii, R26).
Declines an LRU library, as 5b-ii did.

**Independent: yes, of Task 2 by files.**

**Deliverable count: 3** (the charging order with the invalid-address
charge, the fail-closed eviction, the /64 keying).

---

## Task 4: The link inside the prefix, the load-time refusals, and the mail copy

Model: `sonnet`. Depends on Task 3 (`handlers.go`).

Review major 2 and minors 10 and 12, plus the magic-link mail copy
`docs/STATUS.md` carries.

**Files:**
- Modify: `internal/auth/token.go`, `internal/auth/token_test.go`
  (`LinkURL`)
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`
  (the link's prefix, the mail copy)
- Modify: `internal/auth/session.go`, `internal/auth/session_test.go`
  (the mint clamp)
- Modify: `cmd/dubplate-server/auth.go`, `cmd/dubplate-server/auth_test.go`
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/serve_test.go`
  (the prefix into `HandlersConfig`, the port bound in `newMailer`, and a
  shell-level routing test)

**Interfaces:**
- Consumes: `serve.Prefix`'s `String` and `Owns` (read only; this pass
  writes nothing under `internal/serve`).
- Produces: `LinkURL(publicOrigin, prefix, token string) (string,
  error)`, where `prefix` is the owned prefix's string form.
  `HandlersConfig` gains `Prefix string` and `SessionAbsoluteCap
  time.Duration` fields. `MintSession(secret []byte, address string, now
  time.Time, lifetime time.Duration)` (its signature at `27cd7b5`,
  `internal/auth/session.go:102`) gains a trailing `absoluteCap
  time.Duration` and clamps its expiry to the anchor plus the cap, the
  same clamp `RefreshSession` already applies.

**Outcomes:**

- **A mailed link is `origin + prefix + "/login/link"`** (R-5c9), with
  the token in the query exactly as today.
- **`DUBPLATE_PUBLIC_URL` refuses a path other than empty or `/`**,
  naming the key and the rule, never the value.
- **No refusal message quotes the public URL** (finding 10). Every
  message names the key and the defect. `url.Parse`'s own error is not
  wrapped verbatim either: the Go source formats it as
  `fmt.Sprintf("%s %q: %s", e.Op, e.URL, e.Err)` (`net/url/url.go`, read
  2026-09-23), which quotes the whole value, password included. The
  message reports `e.Err` alone or a fixed phrase.
- **Load refuses a session lifetime above the absolute cap**, naming
  both keys; **`MintSession` clamps** to the cap; **nonce retention
  gains a 240h ceiling**; **`DUBPLATE_SMTP_PORT` must parse to 1 through
  65535** (R-5c11). The port bound lands in the source's `Validate` hook
  so `doctor` and every verb see it, not only `newMailer`.
- **The mail copy is authored against the Microsoft editor standard**
  (`~/.claude/docs/voice/editor.md`), replacing the placeholder. It
  names no site, person, or deployment value, says what the link does,
  states its expiry in minutes computed from `LinkExpiry`, says it works
  once, and tells a reader who did not ask to ignore it. The subject
  stays short and generic. The copy lives beside the handlers as named
  constants or one small template. The conductor runs the register gate
  on it at the segment B boundary (see "The conductor's actions").

**States each check meets:**

| Check | State | Report |
| --- | --- | --- |
| Public URL path | empty, `/` | accepted |
| Public URL path | `/dubplate`, `/x/` | refused: "must carry no path; the serve prefix sets where dubplate lives" |
| Public URL | userinfo with a password | refused naming userinfo; the password absent from the message |
| Public URL | unparseable | refused naming the key; the value absent |
| Lifetime vs cap | lifetime <= cap | accepted |
| Lifetime vs cap | lifetime > cap | refused naming both keys |
| Retention | <= 240h | accepted |
| Retention | > 240h | refused naming the ceiling |
| SMTP port | `1`, `587`, `65535` | accepted |
| SMTP port | `0`, `65536`, `-1`, `abc` | refused naming the key |

**Acceptance criteria:**

- The review's routing reproduction becomes a shell-level test in
  `cmd/dubplate-server/serve_test.go`: a link minted through the real
  handler config for the default prefix has a path `serve.Prefix.Owns`
  reports true, and a request to that path through `shell.Handler()`
  with a stub subtree registered at `login` reaches the stub and never
  the fallthrough (a recording fallthrough asserts it saw nothing).
  **Fixture state:** a fallthrough that records `RawQuery`, set, so a
  pass cannot be explained by a 404 before the fallthrough.
- The review's userinfo reproduction is a test asserting the password
  literal is absent from the error text, for userinfo and for an
  unparseable value that carries a password.
- `doctor` with lifetime 87600h and cap 1h reports the source
  unsatisfied (proved through the existing doctor harness, the
  review's own repro).
- A mint with a lifetime above the cap produces a cookie whose
  `Max-Age` does not exceed the cap.
- The mail body test asserts the link, the expiry in minutes for a
  15-minute fixture, and the absence of any configured value other than
  the link.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Drop the prefix from `LinkURL` → the routing test reaches the
   fallthrough.
2. Accept any public-URL path → the `/dubplate` row loads.
3. Quote the value in the userinfo message → the password row fails.
4. Wrap `url.Parse`'s error with `%w` → the unparseable-password row
   fails.
5. Remove the mint clamp → the `Max-Age` row fails.
6. Parse the port with `Atoi` alone → the `0` row loads.

**2.0 hook: neutral**, and it protects hook 6's fallthrough half: a token
no longer reaches the slot rung 12's driver fills.

**Prior art:** none beyond the stdlib `net/url`, read in source.

**Independent: no.**

**Deliverable count: 4** (the prefixed link with the public-URL path
rule, the message redaction, the lifetime, retention, and port bounds
with the mint clamp, the mail copy).

---

## Task 5: `doctor` names the failing key, and an unreadable nonce record sweeps

Model: `sonnet`. Depends on Task 4 (`auth.go`'s refusal shapes feed
`doctor`).

Review minors 11 and 19.

**Files:**
- Modify: `cmd/dubplate-server/doctor.go`, `cmd/dubplate-server/doctor_test.go`
- Modify: `internal/auth/nonce.go`, `internal/auth/nonce_test.go`

**Outcomes:**

- **The selected identity source's one line names every failing key**
  (R-5c10), key names only, never a value, in the form `magic-link
  unsatisfied: <KEY> invalid[, <KEY> missing ...]`. The line count is
  unchanged, and the registry-derived line-count test still passes
  unmodified.
- **A nonce record the sweep cannot read falls back to its mtime**, the
  same fallback an unparseable record already takes, so it is removed
  once past retention. Its nonce still refuses a replay until then.

**Acceptance criteria:**

- The review's three cases each name their key on the source's line: a
  short `DUBPLATE_SESSION_SECRET_PREVIOUS`, a short
  `DUBPLATE_MAGIC_LINK_SECRET`, and a public URL with userinfo. The
  secret value is absent from the output.
- Two failing keys at once both appear on the one line. **Fixture
  state:** the source selected, so the line prints at all.
- A record made unreadable (mode 0000, in a test that skips under root
  and says so) is removed by a sweep past its mtime plus retention and
  retained before it, and its nonce refuses a replay while retained.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Print only the first failing key → the two-keys row fails.
2. Print the value beside the key → the secret-absent row fails.
3. Skip the mtime fallback on a read error → the unreadable row keeps
   the record forever.

**2.0 hook: neutral.**

**Prior art:** none.

**Independent: the nonce half is independent of every task by files;
the `doctor` half depends on Task 4.**

**Deliverable count: 2.**

---

## Task 6: The CLI admin grant, the renamed sentinel, and the last-admin guard

Model: **`opus`**. Depends on Tasks 1 and 2.

**Files:**
- Modify: `internal/identity/identity.go`, `internal/identity/identity_test.go`
  (one channel constant)
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `internal/auth/seam_test.go` (the no-command-line-channel proof)
- Modify: `cmd/dubplate-server/doctor.go`, `cmd/dubplate-server/doctor_test.go`
  (the no-admin roster row)

**Interfaces:**
- Produces: `identity.ChannelCommandLine`, `String` `command-line`,
  accepted by `Valid`. `auth.ErrAdminGrantForbidden`, replacing
  `ErrRosterAdminChannelForbidden` at every reference in the module.
  `auth.ErrNoAdmin`, wrapping `ErrRosterRecordInvalid`, returned by any
  write whose result would name no admin.

**Outcomes:**

- **`AddMember` admits `Admin: true` on `ChannelCommandLine` and refuses
  it on every other channel** with `ErrAdminGrantForbidden` (R-5c7). The
  refusal runs before any lock or write.
- **The channel's doc comment is corrected.** `identity.Channel` says
  "how a verified identity was asserted"; the command line asserts an
  operating-system user by process ownership, and the comment says so.
- **The seam never produces `ChannelCommandLine`**: a seam test asserts
  the channel of every verified identity is `ChannelCookie`.
- **Every roster the store writes names at least one admin** (R-5c8),
  checked inside Task 1's locked read-modify-write function against the
  disk copy, before the write. A violating write returns `ErrNoAdmin` and
  changes no byte. No demote or remove method is added.
- **`doctor`'s roster row fails for a roster that names no admin**,
  reading `no admin; run dubplate-server member add --admin`, within the
  existing row (no new row).

**States each check meets:**

| Write | Disk roster before | Candidate | Report |
| --- | --- | --- | --- |
| add | empty (fresh `CreateRoster`) | non-admin | `ErrNoAdmin`, file unchanged |
| add | empty | admin, command-line channel | written |
| add | empty | admin, cookie channel | `ErrAdminGrantForbidden`, file unchanged |
| add | members, none admin (hand edit) | non-admin | `ErrNoAdmin`, file unchanged |
| add | members, none admin | admin, command-line | written |
| add | one admin | non-admin, any valid channel | written |
| add | one admin, then a hand edit removes the admin before the lock | non-admin | `ErrNoAdmin` (the check reads the disk copy) |

**Acceptance criteria:**

- Every row above is a table case asserting the sentinel and the file's
  bytes before and after. **Fixture state for the last row:** the
  `Roster` value opened while the admin existed, so only a disk re-read
  explains the refusal.
- `grep -rn ErrRosterAdminChannelForbidden` over the module returns
  nothing; `errors.Is(err, ErrAdminGrantForbidden)` holds for each
  refused channel.
- An audit record for a command-line add carries `channel=command-line`
  and `client_ip=local`, and an operating-system-user actor passes
  Task 2's per-channel actor rule.
- `doctor` against an empty roster and against a hand-edited no-admin
  roster reports the roster row failing with the recovery text, and the
  line count is unchanged.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Admit the flag on `ChannelServiceToken` too → the cookie and
   service-token refusal rows write.
2. Check the admin invariant against the memory copy → the hand-edit
   race row writes.
3. Check the invariant after the write → the byte-identity rows fail.
4. Let the seam emit the command-line channel → the seam row fails.
5. Drop the `doctor` no-admin branch → the empty-roster row passes.

**2.0 hook: extends hook 5's write half.** Rung 6's 6d runs this guard
inside its write transaction (rung 6 spec `:816-818`) and criterion 6's
"a concurrent self-demotion of the last two admins leaves one" is this
invariant under the demote path 6d or rung 14 adds.

**Prior art:** none; this is the spec's own invariant.

**Independent: no.**

**Deliverable count: 4** (the channel with the grant, the rename, the
last-admin invariant, the `doctor` row).

---

## Task 7: The `dubplate-server member` verb group

Model: `sonnet`. Depends on Task 6.

Fork 1's first surface (`spec:122-124`, R37).

**Files:**
- Create: `cmd/dubplate-server/member.go`, `cmd/dubplate-server/member_test.go`
- Modify: `cmd/dubplate-server/main.go` (the `AddCommand` block)

**Interfaces:**
- Consumes: `config.Load`; `auth.OpenRoster`, `auth.CreateRoster`,
  `(*Roster).AddMember`, `(*Roster).List`; `identity.ChannelCommandLine`;
  `os/user.Current`.
- Produces: `newMemberCmd(environ []string, reg *config.Registry,
  stdout, stderr io.Writer) *cobra.Command`, the same constructor shape as
  its siblings, with `add` and `list` subcommands.

**Outcomes:**

- **`member add <address> --name <display name> --streamer-username
  <handle> [--admin]`** loads configuration (R-5c14), opens the roster,
  creating it through `CreateRoster` when it is absent, and adds the
  member through `AddMember` on `ChannelCommandLine`, with the
  operating-system user as the actor. On an existing roster it never
  creates: a second `add --admin` adds a second member to the existing
  file. `--name` and `--streamer-username` are required (fold 3).
- **The verb refuses when its effective uid differs from the roster
  directory's owner**, naming both numeric ids and the directory
  (R-5c14), before creating or writing anything.
- **The verb refuses, naming the selector, when the selected identity
  source requires no roster.**
- **`member list`** opens the roster read-only and prints one line per
  member in file order: address, display name, streamer username, and an
  admin marker, tab-separated, with a header line. It prints to stdout,
  which is the verb's purpose; nothing else it writes carries an address.
  Task 2 already refuses control and format characters in every field,
  so the output is terminal-safe.
- **The audit record goes to stderr** through a `slog` text handler
  (R-5c15).
- **Each refusal is one line on stderr and a nonzero exit**, in `main`'s
  existing error path. The line names the sentinel's meaning (for
  example "duplicate address", "the roster must name at least one
  admin") and never the refused value.

**States each check meets:**

| Invocation | Roster state | Report |
| --- | --- | --- |
| `add --admin a@example.test` | absent | created at 0600; one admin; audit record on stderr |
| `add --admin b@example.test` | exists with a | added; a still present; file not recreated |
| `add c@example.test` (no `--admin`) | absent | refused before `CreateRoster` runs ("the first member must be an admin"); nothing created; exit nonzero |
| `add d@example.test` (no `--admin`) | exists, empty (left by an earlier failed bootstrap) | `ErrNoAdmin` from the store's guard; file unchanged |
| `add` with a missing `--streamer-username` | any | refused before any open |
| `add` as a uid other than the directory owner | any | refused naming both ids; nothing created |
| `add` | roster lock held past the verb's deadline | "roster busy", exit nonzero |
| `list` | absent | the `ErrRosterMissing` line, exit nonzero |
| `list` | written by the API (Task 9 proves this end to end) | every member, admin marked |

> The verb's pre-check on an absent roster is the smaller surprise: a
> non-admin first add would otherwise create an empty file and then fail
> the store's guard. The store's guard (R-5c8) still covers the empty
> roster an interrupted bootstrap or a hand edit leaves, which the next
> row proves.

**Acceptance criteria:**

- Every row above is a test driving the cobra command with `SetArgs`
  against a temp-dir roster and a fixture environment that selects
  `magic-link` with synthetic secrets. **Fixture state:** the roster path
  absent for the create rows, present for the second-add row.
- The uid row runs by pointing the roster at a directory whose owner is
  not the test's uid when the test can arrange one (for example a
  root-owned system directory such as `/` when not running as root) and
  skips with a stated reason otherwise. It asserts nothing was created.
- The audit record is captured from the command's stderr and carries
  the actor, `command-line`, the address, `admin=true`, and
  `client_ip=local`.
- `member list` output is asserted exactly against a two-member roster.
- The `AddCommand` block lists `newMemberCmd`. The template-key
  classification test and the doctor line-count test pass unchanged.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Call `CreateRoster` unconditionally → the second-add row fails with
   `ErrRosterExists`.
2. Pass `ChannelCookie` → the `--admin` rows fail with
   `ErrAdminGrantForbidden`.
3. Drop the uid check → the uid row creates a file.
4. Print the refused address in the error → the refusal-line row fails.
5. Drop the absent-roster pre-check → the non-admin first-add row
   creates an empty file.

**2.0 hook: extends hook 5.** The verb is the write-side caller rung 6's
store keeps; 6d criterion 1 has "an imported admin can grant admin
through the CLI", which is this verb.

**Prior art:** `github.com/spf13/cobra`, already the repo's CLI library.
`os/user.Current` from the stdlib; the task quotes its `go doc` before
relying on it.

**Independent: yes, of Task 8 by files** (`cmd/dubplate-server/member.go`
and `main.go` against `internal/auth/admin.go`). Both follow Task 6.

**Deliverable count: 4** (`add` with the create-or-open rule, the uid
guard, `list`, the wiring).

---

## Task 8: The admin gate and the roster API

Model: **`opus`**. Depends on Tasks 2 and 6.

Fork 1's second surface, backend half. The page is 5c-ii's; this task
builds the handlers it calls and the gate its document route will
reuse.

**Files:**
- Create: `internal/auth/admin.go`, `internal/auth/admin_test.go`

**Interfaces:**
- Consumes: `Seam.VerifyIdentity`; `SessionCookie`; `(*Roster).List`,
  `(*Roster).AddMember`; Task 2's sentinels; `identity.ChannelCookie`.
- Produces: `auth.RequireAdmin(cfg AdminConfig, next http.Handler)
  http.Handler`, the gate, whose first caller is this task's API and whose
  second is 5c-ii's document route. `auth.NewRosterAPI(cfg AdminConfig)
  http.Handler`, one handler answering GET (list) and POST (add) and 405
  otherwise. `auth.AdminConfig{Seam Seam; Roster *Roster; ClientIP
  func(*http.Request) string; Now func() time.Time; Logger *slog.Logger;
  BodyCap int64}`. The verified identity reaches `next` through a request
  context value with an unexported key and an exported accessor, so the
  API reads the actor from the gate and never from the request.

**Outcomes:**

- **The gate** (R-5c12): verify through the seam with `sameOrigin` true
  only for `Sec-Fetch-Site: same-origin`; answer 404 for every refusal,
  including a verified non-admin; log the refusal reason and the client
  IP, never the cookie value and never an address; on success with a
  refreshed value, set the refreshed cookie; call `next` with the
  verified identity in the context.
- **GET** answers `no-store` JSON: a `members` array, each entry with
  `address`, `display_name`, `streamer_username`, and `admin`. A roster
  store failure answers 500 with an outcome, never a partial list.
- **POST** caps the body at `BodyCap` before parsing (`http.MaxBytesReader`,
  as the 5b-ii handlers do), reads `address`, `display_name`, and
  `streamer_username`, refuses any request carrying an `admin` field
  (R-5c13), and calls `AddMember` with the gate's actor,
  `ChannelCookie`, and `ClientIP`. It answers one outcome per result, in
  the same `{"outcome": ...}` shape as the 5b-ii handlers, always
  `no-store`.

**Outcomes table (the POST):**

| Result | Status | `outcome` |
| --- | --- | --- |
| added | 201 | `added` |
| `ErrDuplicateAddress` | 409 | `duplicate_address` |
| `ErrDuplicateStreamerUsername` | 409 | `duplicate_streamer_username` |
| `ErrInvalidMemberAddress` | 400 | `invalid_address` |
| `ErrInvalidDisplayName` | 400 | `invalid_display_name` |
| `ErrInvalidStreamerUsername` | 400 | `invalid_streamer_username` |
| `ErrMissingStreamerUsername` | 400 | `missing_streamer_username` |
| an `admin` field present | 400 | `admin_not_grantable` |
| body over the cap | 413 | `too_large` |
| body unparseable | 400 | `malformed` |
| `ErrRosterBusy` | 503 | `busy` |
| any other error | 500 | `failed` |

The outcomes are specific because the caller is already trusted
(`spec:1421-1424`); the copy a person reads is 5c-ii's string deck.

**Acceptance criteria:**

- The gate's refusal table is a test, each row asserting 404 and a
  byte-identical body across rows: no cookie, a forged cookie, an expired
  session, a session older than its epoch, a member removed from the
  roster after signing in, and a verified non-admin. **Fixture state:**
  a roster holding both an admin and a non-admin, and valid sessions
  minted for both, so the non-admin row's refusal comes from the flag and
  not a failed verification.
- A captured log for the forged-cookie row does not contain the cookie
  value, and no refusal log line contains an address.
- A same-origin GET with a valid admin session carries a `Set-Cookie`
  whose expiry moved forward and whose anchor did not; a
  `Sec-Fetch-Site: same-site` GET and a header-less GET carry none.
  **Fixture state:** a session old enough that refresh visibly moves its
  expiry.
- Every POST outcome row above is a test. Each refused row asserts the
  file's bytes unchanged. The `added` row asserts exactly one audit
  record naming the admin's address as actor and `cookie` as channel.
  Each refused row asserts no audit record.
- The `admin` field row holds for `admin=on`, `admin=false`, and
  `admin=` (present and empty).
- `bash scripts/check.sh` green.

**Named mutations:**

1. Answer 403 for a verified non-admin → the byte-identical 404 row
   fails.
2. Read the admin flag from a request field or header → the non-admin
   row reaches the API.
3. Refresh on any request → the header-less row carries a cookie.
4. Pass `admin` through to `AddMember` → the `admin` field row reaches
   the store (and the store's refusal answers `failed` rather than
   `admin_not_grantable`).
5. Take the actor from the form → a test posting a spoofed actor field
   finds it in the audit record.
6. Parse before capping → the over-cap row reads the whole body (assert
   through a reader that counts bytes read).

**2.0 hook: neutral to the schema, and it makes 2.0's web admin surface
cheap**: `RequireAdmin` is the gate every later admin page reuses.

**Prior art:** none external. The gate reuses the seam 5b-ii built and
the outcome shape its handlers use.

**Independent: yes, of Task 7 by files.**

**Deliverable count: 4** (the gate with its refresh, the GET, the POST
with its outcome table, the admin-field refusal).

---

## Task 9: The serve wiring, and the exit gate end to end

Model: `sonnet`. Depends on Tasks 4, 7, and 8.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/serve_test.go`

**Outcomes:**

- **`newServeShell` builds the seam it already constructs into the
  admin config** and mounts `RequireAdmin(cfg, NewRosterAPI(cfg))` under
  the owned prefix as the subtree `roster`, beside the three auth mounts.
  The comment that says "No handler here consumes the seam" is removed,
  since one now does. The body cap is the existing
  `DUBPLATE_AUTH_REQUEST_BODY_CAP` threshold.
- **The path is `<prefix>/roster`**, and 5c-ii's add-member page lives at
  a different route (its outline names it), so the document and the API
  never share a method.
- **The exit gate's cross-surface rows are proved through the real
  shell** (below), because a handler test cannot see the origin check or
  the prefix router.

**Acceptance criteria (every row through `shell.Handler()`):**

- A POST to `<prefix>/roster` with a valid admin cookie and neither
  `Origin` nor `Sec-Fetch-Site` is refused by the origin check and the
  file is unchanged.
- The same POST with `Sec-Fetch-Site: same-origin` adds the member and
  writes one audit record.
- A GET with a non-admin session answers 404; with an admin session it
  answers the list.
- **`member list` reads what the API wrote**: the test posts through the
  shell, then drives `newMemberCmd` with `list` against the same roster
  path and asserts the new member appears. **Fixture state:** the verb
  opened after the API's write, and the shell's roster opened before the
  verb's first write, so the second half proves the cross-process reload
  from the other side too: a `member add` through the verb is visible to
  a subsequent GET through the shell without a restart.
- A link minted by the real request-link handler, followed through the
  shell, reaches the owned `login` stub and not the fallthrough (Task
  4's test, re-run through the wired verb's own construction if Task 4
  built it against a hand-assembled shell).
- `bash scripts/check.sh` green.

**Named mutations:**

1. Mount the API outside the prefix → the non-admin row reaches the
   fallthrough's 404, which is indistinguishable; the implementer proves
   the mutation fires through the admin row instead and reports which
   row caught it.
2. Mount the API without `RequireAdmin` → the non-admin GET answers the
   list.
3. Construct a second roster value for the API → the verb-then-GET row
   still passes (both reload), so report that this mutation does not fire
   and why; the diff-reviewer confirms one roster value is shared.

**2.0 hook: neutral.**

**Prior art:** none.

**Independent: no.**

**Deliverable count: 3** (the mount, the path decision, the end-to-end
exit-gate rows).

---

## Task 10: The pass close, and the 5c-i to 5c-ii handoff

Model: `sonnet`. Depends on Tasks 0 through 9, the simplifier rounds,
and the pass-end security read with its fold.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-5c-i.md`
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-5c-i.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids,
  the gate evidence (`go test ./...` wall time and package-scoped times
  with `-count=1`), the resolutions R-5c1 through R-5c17 with where each
  landed, the ten conductor-owed folds with their rulings, the security
  read's disposition table, the declared coverage gaps (at minimum:
  R-5c15's CLI audit record; a hand edit leaves no audit record; the
  R-5c3 lockout residual; the Navidrome `LIKE` question carried to rung
  12), both budgets, and "what a later pass would be wrong to rediscover".
- **The spec corrections owed**, drafted paste-ready for the conductor:
  R3's write-path requirement (fold 3); `spec:1417`'s admin checkbox
  (fold 8); the 5c row's split into 5c-i and 5c-ii with both ceilings;
  R23's reservation against actuals; R24's 5c review shape across the
  split; review finding 11 against 5b-ii's fold 7 (R-5c10); the
  public-URL path rule (R-5c9) against `spec:1033`'s "Link hostnames
  derive from `DUBPLATE_PUBLIC_URL`".
- **The 5c-ii handoff**: the exported surface 5c-ii calls
  (`RequireAdmin`, the roster API's routes and outcome values, the
  request-link and consume-link routes and outcomes from 5b-ii, the
  `login/link` route the mailed link points at), the kind answer's
  admin contact as data, and every outline item below that this pass's
  results changed.
- **The STATUS draft** stays present tense and under 60 lines.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is
at most 60 lines. No gate change, so `bash scripts/check.sh` green as a
formality.

**Deliverable count: 4** (the HISTORY draft, the spec corrections, the
handoff, the STATUS draft).

---

## The exit gate

`bash scripts/check.sh` green, plus every assertion below. The first
five are the spec's 5c-row assertions (`spec:2331`) that fall in 5c-i's
half; the rest are the security fold's and this plan's.

1. A non-admin session receives 404 from the add-member route (the
   roster API) while an admin session receives the roster.
2. An add-member POST without a same-origin signal is denied, through
   the real shell.
3. An add-member POST that succeeds writes one audit record naming the
   actor and the channel, and one that fails validation writes none and
   leaves the file unchanged.
4. `member add --admin` against an absent roster creates it, and a
   second invocation adds to the existing file rather than creating a
   second.
5. `member list` reads what the roster API wrote.
6. Two writer processes over one roster lose no member and resurrect no
   hand-removed member, and the running server stops resolving a
   hand-removed member at its next lookup (finding 1).
7. A mailed link's path is owned by the prefix and never reaches the
   fallthrough (finding 2).
8. A victim's spent per-address bucket survives a same-instant flood of
   distinct addresses (finding 3).
9. One IPv6 /64 is one client (finding 4).
10. Every loose roster variant finding 7 lists fails to load.
11. No roster write can leave the roster with no admin, and no channel
    but the command line can set the admin flag.
12. No refusal message or log record carries a URL password, a cookie
    value, or a token.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches.

| Surface | Owner | 5c-i's position | Contention |
| --- | --- | --- | --- |
| `internal/auth` | rung 5 | Edited across Tasks 1-8; `admin.go` born | Rung 6 edits it only at 6d, after 5d |
| `internal/identity` | rung 5 | One constant, one refusal set | Rung 6's 6d adds `UserID`, after 5d |
| `cmd/dubplate-server/doctor.go` | rung 1's skeleton, rung 5's rows | Tasks 5, 6 | **Rung 6 lanes 6a and 6b edit it** (rung 6 spec `:1109-1118`); a concurrent lane 2 means a non-fast-forward rebase and the merge ritual's `diff-reviewer` |
| `cmd/dubplate-server/main.go` | rung 1 | Task 7's `AddCommand` line | **Rung 6 edits the registration block**; same rule |
| `cmd/dubplate-server/serve.go` | rung 5 | Tasks 4, 9 | Rung 6 edits it at 6d only |
| `internal/config` | rung 1b's contract | **Untouched** (R-5c17) | Rung 6's 6c runs in lane 2 on this condition |
| `config/roster.example.json` | rung 5 | Task 2's one field | none |
| `go.mod` | shared | **Untouched** | rung 6 adds its SQLite driver |
| `docs/` ledgers | conductor | Task 10 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. **Before Task 1:** rule Task 0's blockers and the ten folds; write
   the lane registry entry and the contested-surface rows; arm the
   runaway guard.
2. **At segment A's boundary:** STATUS; re-derive the ceiling from Tasks
   1-3; confirm every fold reproduction was observed failing first.
3. **At segment B's boundary:** STATUS; one `code-simplifier` round over
   Tasks 1-6's accepted Go, pasting `.claude/instructions/simplifier-brief.md`
   verbatim, its review chunked one `diff-reviewer` per package; **the
   register gate on Task 4's mail copy**, one `prose-voice-reviewer`
   dispatch against the Microsoft editor standard, folded by a small
   implementer dispatch if it returns findings.
4. **At segment C's boundary:** one `code-simplifier` round over Tasks
   7-9, chunked the same way; then **one fresh-context
   `web-auth-security-reviewer`** over this pass's diff (R24's 5c read,
   the roster write path and the admin surface it exists for). Its input
   is this plan's threat-model section, the spec's threat model, both
   prior reviews' "Attacked and found sound" lists as the
   do-not-re-litigate record, and the diff. Point it explicitly at
   R-5c3's lockout residual, the gate's 404 uniformity, and the
   cross-process lock. It does not receive the task decomposition. Fold
   or refuse each finding with the reason recorded; the record is 5c-ii's
   and 5d's do-not-re-litigate list.
5. **The merge ritual** (one lane in the main tree: gate, merge, gate;
   one `diff-reviewer` over the merged result only if a rebase was not a
   fast-forward), then **one `go-architecture-reader` per touched Go
   package** (`internal/auth`, `internal/identity`, `cmd/dubplate-server`),
   no plan in context.
6. **Re-derive 5c-ii's ceiling from this pass's actual** before 5c-ii's
   plan is drafted, and record the rung total against R23.
7. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, in
**four segments**. Every task runs `dubplate-implementer`, then
`diff-reviewer` (`claude-opus-5-5`), then the gate, with one re-dispatch
on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2, 3 | STATUS. Re-derive the ceiling. Confirm failing-first observations. |
| **B** | 4, 5, 6 | STATUS. Simplifier over Tasks 1-6. The mail copy's register gate. |
| **C** | 7, 8, 9 | STATUS. Simplifier over Tasks 7-9. The security read and its fold. |
| **D** | 10 | The merge ritual, the architecture reads, the ledger fold. |

Segments A through C carry three tasks each, every boundary on a task's
own green committed gate, so each is a resume point. D is the close.

**Models.** `sonnet` by default. **Four upshifts to `opus`, all
unconditional**, dispatched as `claude-opus-5-5`: Task 1 (a
cross-process concurrency fix whose broken variants pass every
single-process test), Task 3 (the limiter's eviction and keying, where a
wrong map passes every functional test), Task 6 (the privilege grant and
its invariant), and Task 8 (a cookie-authenticated privilege surface).
Task 0 runs at `opus` in fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Task 1 → Tasks 2, 3, 6, 8 (the locked write function and the reload).
- Task 1 → Task 3 also by file (`handlers.go`).
- Task 2 → Tasks 6, 8 (the sentinels and the actor rule).
- Task 3 → Task 4 (`handlers.go`).
- Task 4 → Task 5 (`auth.go`'s refusal shapes), Task 9 (the prefixed
  link).
- Task 6 → Tasks 7, 8 (the channel and the guard).
- Tasks 7, 8 → Task 9.
- Tasks 0-9 → Task 10.
- **Independent by files:** Tasks 2 and 3; Tasks 7 and 8; Task 5's nonce
  half against everything. They still run sequentially, because one tree
  runs one executor and the module gate runs over `./...`. The marks let
  the conductor reorder when a dependency stalls, never admit a second
  executor.

**Where this pass splits again, if it must:** at the Task 6 / Task 7
boundary, if segment A's re-derivation projects past 7.5M. Tasks 7-9
would open 5c-ii, ahead of the `web/` tasks. Nothing argues against
this cut the way 5b-ii's rejected cut was argued against, because
Tasks 1-6 are fixes and a grant whose callers already exist in tests.

**Budget:** 7.0M, flag 5.6M, STATUS at each boundary, spend scored at
close against the ceiling and against zero attended interactions. The
ten folds and seventeen resolutions are ruled at dispatch together and
do not count as an execution sitting.

---

## Pass 5c-ii: the `web/` tree, the pages, and the gate leg (scoped outline)

**This is an outline, not a plan.** 5c-ii's plan is authored just in time
after 5c-i merges, against the tree 5c-i leaves, with its ceiling
re-derived from 5c-i's actual. What follows fixes scope, order, and the
open questions its plan must close, so nothing here is rediscovered.

**Scope, from `spec:2331` less 5c-i's half.** `web/` born as a SvelteKit
static app under CLAUDE.md's web rule (Svelte 5 runes, DaisyUI 5,
`adapter-static`, client-side rendering only, no server files,
`tus-js-client` not needed until rung 7). The login page with its string
deck, the consumption interstitial, the add-member page. The one composed
theme as data under `web/src/lib/themes/` (R36) and the theme binding (DL
section 2). The embedded woff2 faces with licenses and manifest (R19, DL
R5.2). The `go:embed` path with a committed placeholder and `fs.Sub`
(`spec:1362-1390`). The R4.5 rewrite. The generated branding stylesheet
and the operator CSS file (R14's keys, declared in `cmd/dubplate-server`
through the registry, `internal/config` untouched). The served header
set: CSP, `nosniff`, `Referrer-Policy: no-referrer`, **and HSTS** (review
finding 18, Geoff's routing). The gate's `web/` leg with DL's G1 through
G10 **and the SvelteKit server-file check** (R-5c16), Node as a hard
requirement (R13), and the three rejoin sentences corrected
(`CLAUDE.md`, `scripts/check.sh:5`, and `Makefile:33`, whose wording is
"The web legs rejoin", plural). The web surface standard's section 19
rows for `CLAUDE.md` and `scripts/check.sh` land with the leg; its DL
R4.5 row (one rewrite per prerendered document, W3.7) governs the embed
task. Every page-as-built clause section 15 binds applies (W4.1's five
viewports, W4.10's touch floors, W5, W7.1 and W7.2's empty states on the
roster list, W8.4's plural count, W10.1-W10.7, W3.3, W3.7, W6.1-W6.6,
W6.11, W7.14, W7.15, W7.18, W8.1-W8.9, W9.1). No grammar layer, no
component set, no styleguide: rung 7's.

**Proposed task order (about ten tasks plus Task 0 and the close):**

1. The scaffold, the composed theme data, and the theme binding.
2. The fonts, their licenses, and the manifest.
3. The embed path, the placeholder, and the R4.5 rewrite per
   prerendered document, with R14's theme keys and their validation.
4. The branding stylesheet and the operator CSS file.
5. The served header set with HSTS, the cache rules (`no-cache`
   documents, `no-store` on the consumption route's document, immutable
   hashed assets), and the add-member document route behind 5c-i's
   `RequireAdmin`.
6. The login page and the string deck, with the deck's register gate.
7. The consumption interstitial at `login/link` under the prefix
   (R-5c9): reads the token from `window.location`, clears it from the
   address bar, and POSTs it in a form field.
8. The add-member page: the roster list with its empty state and admin
   marker, the form of address, display name, and streamer username with
   no admin control (R-5c13), and specific copy per 5c-i outcome value.
9. The gate leg: G1-G10, the server-file check, Node required, the
   rejoin sentences.
10. Render and read all three pages at W4.1's viewports before any
    attended check (`see-it-before-geoff-tests-it`).

**Open questions 5c-ii's plan must close with quoted evidence:**

- **The runtime prefix against SvelteKit's build-time base.**
  `DUBPLATE_SERVE_PREFIX` is a runtime threshold, and SvelteKit's
  `kit.paths.base` is a build-time value ("A root-relative path that must
  start, but not end with `/`", SvelteKit configuration docs, read
  2026-09-23). `kit.paths.relative` defaults to `true` ("`base` and
  `assets` imported from `$app/paths` will be replaced with relative asset
  paths during server-side rendering, resulting in more portable HTML").
  The plan must prove, by building and serving under a non-default
  prefix, whether relative paths plus relative API calls work under any
  prefix. If they do not, it must propose a conductor ruling (for
  example, the rewrite stamps the prefix, or a non-default prefix is
  refused until rung 7). **This is the pass's largest unknown.**
- **Where the CSP travels.** The same docs state "When pages are
  prerendered, the CSP header is added via a `<meta http-equiv>` tag (note
  that in this case, `frame-ancestors`, `report-uri` and `sandbox`
  directives will be ignored)". `spec:1348` requires `frame-ancestors
  'none'`, so the Go layer must serve the CSP as a header, with the hashes
  read from the build or computed from each embedded document at startup
  (`spec:1336-1339` leaves the choice open).
- **The `style-src 'unsafe-inline'` owner note** (`spec:1346-1352`)
  rides to that pass, per `spec:2646-2652`.
- **HSTS's `max-age` and its loopback carve-out** (R38 admits http on
  loopback, where HSTS must not be sent or is ignored), quoted from RFC
  6797.
- **Node and package pins** under the `dependency-upgrade` skill: the
  newest production SvelteKit, Svelte, Vite, DaisyUI, and Tailwind at
  plan time.

**Reviewers inside 5c-ii:** `svelte-reviewer` and
`daisyui-a11y-reviewer` with the web surface standard as their lens
(section 15), and a narrow `web-auth-security-reviewer` read over the
header set, the CSP, the embed path, and the add-member document route's
gate (R24's "returns for the admin page", page half). The AI-tells rule
runs inside; the string deck's register gate is separate (`spec:1432-1436`).

**2.0 hooks:** neutral except hook 5 through DL R4.2's per-member theme
row, which W5.10's signed-out cookie does not change.

**Provisional ceiling: 9.5M** ("Token ceiling" above), re-derived from
5c-i's actual.
