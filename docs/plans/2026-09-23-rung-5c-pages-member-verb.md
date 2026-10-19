# Rung 5, Pass 5c-i (The Security Fold, the CLI Admin Grant, and the `member` Verb) Implementation Plan

Status: **revision 2.2, 2026-09-23, approved by Geoff with all six
rulings yes.** The review history, the full ruling questions with their
branches, and every finding's disposition live in
`docs/research/2026-09-23-rung-5c-plan-fold-record.md`.

Sources: pass 5b-ii's close (`docs/HISTORY.md`, top entry), the 5b-ii
pass-end security review
(`docs/research/2026-09-23-rung-5b-ii-security-review.md`)
and Geoff's 2026-09-23 routing of it, the rung 5 identity-and-auth spec
at revision 2.1 (`docs/specs/2026-09-08-rung-5-identity-auth.md`),
the ratified rung 6 spec
(`docs/specs/2026-09-23-rung-6-ops-seam-store.md`), and the
ratified web surface standard
(`docs/specs/2026-09-12-web-surface-standard.md`).

This plan splits 5c into two passes. It is 5c-i in full, with 5c-ii as
a scoped outline whose first two tasks are specified in full. The spec's
5c row (`spec:2381`) carries fifteen deliverables, Geoff's routing adds
thirteen security findings, and `docs/STATUS.md` carries five more items,
about twenty tasks in all: past this build's twelve-task line and the
spec's own fifteen-task split trigger (`spec:2391-2396`, stated for 5b
and applied here by the same reasoning).

## Rulings (ruled yes, Geoff, 2026-09-23)

| # | Ruling | Where it lands |
| --- | --- | --- |
| 1 | The per-address sign-in limit counts only addresses on the roster; the per-client and consumption limits keep least-recently-used eviction. | R-5c3, Task 3 |
| 2 | The conductor may run one read-only query of the box's Navidrome usernames. **Done 2026-09-23**: the estate holds one account, its name is email-shaped, and no case-insensitive duplicate exists, so the write-path shape widened to `[A-Za-z0-9][A-Za-z0-9._@+-]*`. The name is not recorded. | R-5c6, Task 2 |
| 3 | Display names allow the zero-width joiner and non-joiner (U+200D, U+200C); addresses keep refusing them. | Task 2 |
| 4 | The `member` verb and the add-member form require `streamer_username` from 5c; the loader keeps it optional. | R-5c6, Tasks 2, 6, 7 |
| 5 | The add-member page has no admin checkbox; the command line is the only admin grant. | R-5c13, Task 6, 5c-ii Task 1 |
| 6 | Rung 5's reservation moves from 15.75M to about 37M, with 5c at about 18M across its two passes. | Token ceiling |

---

**Eight conductor-owed folds**, each marked at its own site as
`**Conductor-owed fold:**` with the question and the applied answer, so
the conductor ratifies or reverses each in one edit. Their numbers are
kept for cross-reference.

- **Fold 1, the split.** Applied as two passes cut at the Task 7 / Task 8
  boundary, so every 5c-i export has its production caller inside 5c-i.
  ("The split.")
- **Fold 2, Geoff's "security-fold Task 1".** Applied as Tasks 1 through
  5, every one landing before any caller builds on the roster, the link
  URL, or the limiters. (R-5c2.)
- **Fold 5, the CLI admin grant's channel.** Applied as a new
  `identity.ChannelCommandLine` that `AddMember` admits the admin flag on
  and `identity.Identity.Valid` refuses. (R-5c7.)
- **Fold 6, the last-admin guard's shape at rung 5**, where no demote or
  remove path exists. Applied as a write-path invariant with no
  exemption: `CreateRoster` writes the first admin in the same locked,
  validated write, and no demote method is exported. (R-5c8.)
- **Fold 7, `doctor`'s failing-key report** (review finding 11). Applied
  as naming each failing key on the source's own line through a keyed
  error type, with no new rows. (R-5c10.)
- **Fold 10, the ceiling.** Applied as a 6.5M execution ceiling with the
  flag at 5.2M, planning counted as spent at launch. ("Token ceiling.")
- **Fold 11, review minor 6** (the kind answer's post-ceiling membership
  signal), routed to 5c by the 5b-ii close though absent from Geoff's
  list. Applied as refusing the review's code fix and carrying its
  runbook sentence to 5d. (R-5c18.)
- **Fold 12, the unused channels.** Applied as dropping `ChannelHeader`
  and `ChannelServiceToken`, which no production code asserts; each
  returns with the source that asserts it. (R-5c7, Task 6.)

**What this pass is.** Tasks 1 through 5 fold the security majors and
minors Geoff routed here, so every later caller builds on corrected code.
Task 6 lands the CLI admin grant, drops the two unused channels, renames
the grant sentinel, and makes the bootstrap one validated write under the
last-admin guard. Task 7 builds the `dubplate-server member` verb group,
`add` and `list`, the production caller of everything Task 6 exports.
Task 8 is the close. The pass is Go only.

**What this pass is not.** No admin gate, no roster API, and no serve
wiring for them: those open 5c-ii. No `web/` tree, no page, no theme, no
font, no embed path, no served header set, no CSP, no HSTS, no gate leg,
no server-file check, no Node requirement, and no rejoin-sentence
correction: those are 5c-ii's. No systemd unit, no `deploy.sh` edit, no
`doctor` probe row, no verify endpoint, no env template row, no
trusted-proxy setting, and no epoch-bump verb: those are 5d's. **No
deploy of any kind.** A task that believes it needs any of these stops
and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process",
> the pass-execute chain per task, sequential, dispatched in the three
> segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and
> never reads a diff, a test log, or a gate transcript inline. Subagents
> never dispatch agents, so `code-simplifier` and the pass-end
> `web-auth-security-reviewer` are the conductor's dispatch. The
> conductor owns `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the
> lane registry, and the contested-surface rows. This pass runs in lane
> 1, the main tree. If rung 6 launches as lane 2 during this pass, it
> takes a worktree, and **the live-executor check is the lane registry,
> never `pgrep -f Projects/dubplate`**.

---

## The split

> **Conductor-owed fold 1:** split 5c into 5c-i and 5c-ii, or run one
> pass? Applied as the split, cut at the Task 7 / Task 8 boundary: the
> security fold, the CLI grant, and the `member` verb in 5c-i; the admin
> gate, the roster API, their wiring, and the `web/` tree in 5c-ii.

Three reasons:

- Geoff's ruling orders it. The security fold is "fixed before any page
  builds on the roster, the link URL, or the limiters". A pass boundary
  makes that ordering structural: 5c-ii's plan is drafted against a tree
  where the fold has already merged and been reviewed.
- Every 5c-i export has its production caller in 5c-i (R-5c20). The verb
  calls `CreateRoster`, the command-line grant through `AddMember`, and
  `ErrRosterExists`, and passes `identity.ChannelCommandLine`. The admin
  gate's two callers, the roster API and the add-member document route,
  both land in 5c-ii, so the gate is born beside them. Do not cut inside
  Tasks 6 and 7: the grant and the bootstrap would ship with test-only
  callers.
- 5c-i is complete and gradable on its own. Its entry point is real: the
  `member` verb writes and reads the roster across processes, which a
  security reviewer can attack end to end. The cookie-authenticated admin
  surface is reviewed in 5c-ii with the page that uses it, where R24's
  "returns for the admin page" read belongs.

What moves to 5c-ii: the admin gate, the roster API, and their serve
wiring (5c-ii Tasks 1 and 2, specified in full at the head of its
outline), with the four assertions of the spec's 5c row that need them
(the non-admin 404, the same-origin denial, the web write's audit record,
and `member list` reading the API's write). The SvelteKit server-file
check moves too (R-5c16): with no `web/` tree it would pass on an empty
iteration. Review finding 18 (HSTS) joins 5c-ii's served header set, per
Geoff's routing.

---

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/research/2026-09-23-rung-5b-ii-security-review.md`,
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
   pass split (`:2371-2396`), the contested surfaces (`:2398-2430`), the
   band and ceiling (`:2432-2491`), the staleness pre-task
   (`:2493-2557`), and risk 14 (`:2644-2650`).
3. **`docs/HISTORY.md`'s top entry (rung 5, pass 5b-ii)**, its "The 5c
   handoff" and "Also carried to 5c" paragraphs, which are binding on this
   plan, and its declared coverage gaps. Task 0 resolves both paragraphs
   by heading and reports any drift from what this plan assumes.
4. **The rung 6 spec**, ruling 8 (`:96-100`), "Identity in the store"
   (`:767-820`), and 6d's criteria (`:1047-1080`). They state what rung 6
   expects 5c to have built: the CLI admin grant, the last-admin guard
   inside the write path, and a roster whose `streamer_username` rung 6's
   import will require.
5. **`docs/specs/2026-09-11-configuration-homes-design.md:668-669`**,
   which routes the last-admin guard to rung 5's plan.

**Every relied-on library or stdlib behavior below is quoted from `go
doc`, the Go source, or a man page read on 2026-09-23 against
`go1.27.1 linux/amd64`.** A task that relies on a behavior this plan
does not quote confirms it the same way and reports what it read.

---

## Goal

Every surface a page will build on is corrected, and the command-line
writer the spec names exists. After this pass, the roster survives two
writer processes and a hand edit. The loader refuses any file that is
not the schema. A mailed link lands inside the owned prefix. No
single refused client can reset a victim's bucket, and one IPv6 /64
counts as one client. A fresh install gets its first admin from
`dubplate-server member add --admin` in one validated write, and no
roster write can leave the group with no admin.

---

## Owner touchpoints: none remaining

`spec:2696-2700` names rung 5's expected touchpoints, and none is
5c-i's. The six rulings are answered. What would become an owner
touchpoint: a secret minted (none this pass); a change to what the box
runs (**nothing deploys**; the env template still selects `magic-link`
without its nine required keys); a third module dependency (R26 allows
two, both taken; `encoding/json/v2` and `syscall` are stdlib); any
threshold default changing on a task's own judgment.

---

## Token ceiling: 6.5M execution for 5c-i

> **Conductor-owed fold 10:** rule a 6.5M execution ceiling for 5c-i
> with the flag at 5.2M, planning counted as already spent at launch?
> Applied as stated.

`spec:2447-2448` grants the re-derivation: "Each ceiling is re-derived
from the previous pass's actual report before the next plan is drafted."
The basis is 5b-ii's recorded "about 6.4M" (`docs/HISTORY.md`, top
entry, "Both budgets") over the twelve tasks it bought: **0.533M per
task**, with Task 0, fix rounds, simplifiers, the security read, the
close, and the conductor's turns amortized in the rate. Planning is
spent before Task 0 dispatches, so the conductor records its actual
against the rung and outside this ceiling. The derivation of the basis,
5c-ii's provisional figure, and the rung reservation arithmetic are in
the fold record.

**The arithmetic for 5c-i.**

1. Eight tasks (Tasks 1 through 7 and the close, matching 5b-ii's
   twelve, which included its close) at 0.533M is **4.27M**. Task 0 is
   amortized in the rate, as it was at 5b-ii.
2. Three of the eight run at `opus` against four of twelve at 5b-ii:
   add **0.05M**. The two folds added scope inside existing tasks (the
   validated single-write bootstrap, the limiter's split charging, the
   keyed `doctor` errors, the channel drop, about twenty-five new table
   rows): add **0.25M**.
3. Three `go-architecture-reader` dispatches (`internal/auth`,
   `internal/identity`, `cmd/dubplate-server`): **0.15M**, an estimate,
   since 5b-ii's four reads carry no recorded figure.
4. **The projection is 4.72M.**

The flag sits a task's worth above the projection. 5b-ii crossed its
flag in its last segment, which produced its one execution sitting. A
ceiling of **6.5M** puts the 80% flag at **5.2M**, 0.48M above the
projection, about one task at the measured rate. The flag then marks a
real overrun rather than firing on plan. Re-derive at the segment A
boundary from Tasks 1 through 3's reports.

5c-ii's provisional figure is about 10.3M (9.25M execution and 1.0M
planning), re-derived from this pass's actual at its close. The rung
reservation stands at about 37.1M against 15.75M (Ruling 6).

Checkpoint: STATUS at every segment boundary, after Task 3, after
Task 7, and at the close. Every task reports its spend, its mutation
outcomes, and its measured gate wall time.

Deliverable count: twenty-nine across Tasks 1 through 8, plus Task 0's
report. Tasks 1, 2, 4, 6, 7, and 8 carry four each; Task 3 carries
three; Task 5 carries two. No task passes four.

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
  `session.go`, `nonce.go` edited, and `seam.go`'s one comment; their
  tests), `internal/identity/identity.go` and its test (the normalizer's
  character refusal and the channel set),
  `cmd/dubplate-server` (`auth.go`, `doctor.go`, `serve.go`, `main.go`
  edited; `member.go` born; their tests), and `config/roster.example.json`.
  **It writes nothing under `internal/config`**, which is rung 6 pass
  6c's condition for running in lane 2 (rung 6 spec `:1023`). It writes
  nothing under `internal/serve`, `internal/notify`, `internal/engine`,
  `internal/harness`, `internal/atomicfile`, `web/`, `systemd/`,
  `scripts/`, `env/`, `Makefile`, `CLAUDE.md`, `corpus/`, `tools/`, or
  `_salvage/`, and no `docs/` file outside Task 8's conductor-owned
  drafts. A task that believes it needs one stops and reports.
- **No new module dependency.** `go.mod` and `go.sum` stay byte-identical.
  `encoding/json/v2` ships in go1.27.1's standard library and builds with
  no `GOEXPERIMENT` set (measured 2026-09-23: a scratch program importing
  it ran with `go env GOEXPERIMENT` empty). It is still experiment-gated:
  every file carries `//go:build goexperiment.jsonv2`
  (`src/encoding/json/v2/arshal.go`), the experiment is on in go1.27's
  baseline (`src/internal/buildcfg/exp.go`), and `GOEXPERIMENT=nojsonv2`
  breaks the build. That is a stability class below GA stdlib, recorded
  here and in Task 1's prior-art line; the adoption stands. File locking uses `syscall.Flock`
  from the standard library. The salvage lock package imports
  `golang.org/x/sys/unix`, a third dependency R26 does not budget, so it
  is adopted in reasoning only.
- **`internal/auth` reads no environment and renders no HTML**
  (`spec:1471-1479`). The handlers are `http.Handler` values that
  `cmd/dubplate-server` mounts, as 5c-ii's gate will be.
- **Assertion discipline at the repo's bar.** No silent-success fakes, no
  vacuous assertions, no expectation derived from the code under test.
  Every hard-coded expectation cites the spec, an RFC, a quoted doc, or
  the security review.
- **A security property is proved by behavior, never by structure**
  (risk 3, `spec:2576-2580`).
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
- **Fixtures are obviously not credentials** (`spec:2428-2430`).
  `.gitleaksignore` stays untouched.
- **No token, secret, cookie value, or URL password reaches a log record
  or an error string.** The consumption route logs its path and never
  `RawQuery`.
- **Addresses reach exactly two outputs in 5c-i.** The roster audit record names
  the actor and the address added, because `spec:1139-1143` requires it.
  `member list` prints the roster to its own stdout, because listing the
  roster is the verb's purpose. (From 5c-ii, the roster API's GET answer
  carries the roster to a verified admin.) No other log record and no
  error string carries an address. A `streamer_username` counts as an
  address for this rule, since the estate's is email-shaped (Ruling 2).
- **Universality.** No estate hostname, path, email, or personal name
  anywhere. Every tunable is a threshold with a default.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx`
  first where a call can block, which after Task 1 includes every roster
  write. Errors lowercase and context-first. Exported only when a caller
  outside the package branches on it or calls it (R-5c20). Sentinels where a caller
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
| A member's `admin` flag | `dubplate-server member add --admin`, which calls `CreateRoster` or `AddMember` with `identity.ChannelCommandLine` (Tasks 6, 7). An operator's hand edit of the roster file, read at the next reload (Task 1); it leaves no audit record, a recorded residual. Rung 6's import, later, once. | The roster API, from 5c-ii: it refuses any request carrying an `admin` key (5c-ii Task 1), and the store refuses the flag on every channel but `ChannelCommandLine` (Task 6). The session cookie: the seam reads the flag from the roster record, never the request (5b-ii, found sound). An identity carrying `ChannelCommandLine` fails `identity.Identity.Valid` (Task 6), which 5c-ii's gate requires. Request-link and consume-link write no roster. |
| The authority to write the roster through the web | A cookie-channel actor whose record in the disk copy, re-read under the store's lock, holds the admin flag (Task 6). | A stale gate decision: `AddMember` re-checks the actor's admin flag under the lock, so a hand edit that demotes the actor between the gate and the write refuses the write. |
| The audit record's actor | The gate's verified, normalized address for the API (5c-ii Task 1), passed through an unexported context key. The operating-system user for the verb, looked up from the effective uid, falling back to `uid:<n>` (Task 7, `spec:1141-1143`). | Any request field. `$USER`. `AddMember` refuses an empty or malformed actor (Task 2). |
| The audit record's channel | The caller's constant: `ChannelCookie` from the API, `ChannelCommandLine` from the verb. | An invalid channel value: `AddMember` refuses `!channel.Valid()` (Task 2). |
| `streamer_username` | The verb (Task 7) and, from 5c-ii, the API, both through `CreateRoster` or `AddMember`. A hand edit, validated at reload. **It becomes a privilege channel at rung 12**: the seam's `Principal` is the username when set (`internal/auth/seam.go`, `VerifyIdentity`), rung 12 injects the principal into the streamer's proxy header (`spec:1243-1245`), and the rung 6 spec quotes Navidrome creating an unknown proxied user. So whoever sets the username decides which streamer account a member becomes, and an admin could map a member onto the streamer's own admin account. | Nothing else writes the roster. The roster's own case-insensitive uniqueness cannot see the streamer's accounts. **Carried to rung 12 as a constraint**: a principal must not map a non-admin member onto a streamer admin account, or rung 12 verifies the mapping (Task 8's HISTORY draft records it as a declared gap). |
| The session cookie's refreshed value | The admin gate, only after the admin check passes, and only for a request whose `Sec-Fetch-Site` is exactly `same-origin` (5c-ii Task 1). | A cross-site or header-less request, and any refused request including a verified non-admin: the gate sets no cookie. |
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
the limiters", and Tasks 1 through 5 all land before the verb (Task 7)
or 5c-ii exist. 5b-ii took the same reading of its own
fold (R-5b17, two tasks).

### R-5c3: the per-address limit charges only roster addresses (Ruling 1)

The review's fix for finding 3 ("at capacity, evict only a bucket that
has fully refilled; if no entry qualifies, refuse the new key") has a
cost the review does not name. Unknown
addresses get no mail (`internal/auth/handlers.go`, `requestLink`), but
today every request still charges the per-address map. A per-address
bucket charged once refills in 12 minutes (five per hour at burst five),
so holding 10,000 of them unrefilled costs about 14 requests a second.
At 20 admitted requests per /64 per hour, that is about 2,500 /64s, a
sliver of one /48, and a /48 is cheap to obtain. Once the map is
full, fail-closed refuses every new key, so every member who has not
asked for a link in the last few minutes gets 429 for as long as the
flood runs. The global mail ceiling does not bound this, since the flood
sends no mail. On the two client-keyed maps, fail-closed buys nothing:
evicting an attacker's client bucket hands it nothing it does not
already hold.

The design Ruling 1 chose:

- **The per-address key is charged after the roster lookup, and only for
  a known address.** The map then never holds more keys than the roster
  has members, so a flood of made-up addresses cannot touch it, and
  finding 3's reproduction (a victim's spent bucket, then 10,000 distinct
  addresses from one client) leaves the victim's bucket in place.
- **Under the kind answer**, a known address over its limit answers 429,
  which discloses nothing the kind answer does not already disclose.
  **Under the generic answer**, it answers `requested` and sends nothing,
  so the response bytes stay uniform (R-5b28 holds).
- **At capacity, the per-address map never evicts an unrefilled bucket.**
  It evicts the least recently used refilled bucket if one exists, and
  otherwise admits the request without tracking the new key. With
  known-only charging, capacity needs a roster of more than 10,000, so
  this is a stated backstop rather than a live path.
- **The per-client and consumption maps keep least-recently-used
  eviction**, unchanged.
- **The per-client key is still charged first, for every request**,
  before normalization (finding 17).

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
two checks refuse all three. Three more v2 behaviors, measured
2026-09-23, are harmless and stay: a member value of `null`
(`"admin":null`, `"streamer_username":null`) decodes as the zero value
and loads; invalid UTF-8 is refused; trailing data is refused. v2 is
experiment-gated (Global Constraints), which Task 1's prior-art line
records.

### R-5c5: the store reloads on a changed file, reads what it stats, and fails closed on a broken one

`Lookup` and `List` check the roster file on every call. When the device
and inode (`os.SameFile`, which `go doc os.SameFile` defines on Unix as
"the device and inode fields of the two underlying structures are
identical"), the size, or the modification time differ from the cached
copy, they reload. Reads take no cross-process lock:
`atomicfile.WriteStream` replaces the file by rename, so an open file is
either the whole old file or the whole new one.

**Every read, whether the first open, a reload, or a writer's re-read,
opens the file, takes `(*os.File).Stat` on that descriptor, reads from
that same descriptor, and caches that stat with that content.** `go doc
os.SameFile` says it "only applies to results returned by this package's
Stat", and `(*os.File).Stat` qualifies. The order matters: `os.ReadFile` followed by
`os.Stat(path)` can cache a renamed-in file's stat against the old
content, and every later check then matches, so a hand-removed member
keeps resolving until the file next changes. That is finding 1 again.
The fstat-first order leaves one race: an in-place write landing between
the fstat and the read, which caches an old stat over new content. That
direction is harmless, because the next check sees a changed stat and
reloads once. The change check itself may stat the path; what must never
happen is caching a path stat against content read through another
descriptor.

**A modification time too close to the read is not trusted.** An in-place
hand edit that keeps the inode and the size inside one timestamp tick
keeps the modification time too. So when the cached modification time is
within one second of the moment the cache was filled, the cache is
unverified and the next call re-reads regardless. The one-second window
is this plan's own rule, sized to cover a coarse filesystem timestamp; it
borrows the idea git's index uses for the same problem, and relies on no
git text. The workstation measured nanosecond-distinct modification
times; the box's filesystem is not measured.

A reload that finds the file absent, unreadable, unparseable, or invalid
returns that error. It does not keep serving the last good copy, because
the last good copy is exactly the member a hand-edit offboarding just
removed, and R2 already rules fail-closed. The request-link handler
tells a store failure apart from `ErrMemberNotFound`. Under the kind
answer, a store failure answers `failed` and never `unknown`, and the
admin contact (`adminContact`, which calls `List`) is never read from a
failing store. Under the generic answer, the response bytes do not change
and the log records the failure.

### R-5c6: `streamer_username` is required on the write path and restricted in shape (Rulings 2 and 4)

Rung 6 ruling 8 (Geoff, 2026-09-23) refuses an import entry with no
`streamer_username`, and 6d criterion 6 has "the page and the verb
refuse a missing or duplicate `streamer_username` by name". A member
added without one at 5c is a member 6d refuses. The loader keeps the
field optional, per R3 and B9, so a hand-written roster that predates
this pass still loads. The close owes three passages an erratum: R3,
R1's write path (`spec:291-292`), and "The roster" (`spec:1106-1107`),
plus `spec:1417`'s "an optional streamer username".

On the write path the shape is `[A-Za-z0-9][A-Za-z0-9._@+-]*`, up to the
existing 64-byte bound. The alphanumeric first character refuses `.`,
`..`, and a leading `-` at no cost to any plausible account name. Both
cases are admitted, because lowercase-only would refuse an existing
account that carries an uppercase letter. `_` stays allowed. Review
finding 15's worry, that Navidrome might resolve a username with SQL
`LIKE`, where `_` is a wildcard, is settled by its source: Navidrome
v0.63.2's `FindByUsername` builds `user.user_name = ? COLLATE NOCASE`,
under the interface comment "FindByUsername must be case-insensitive"
(`persistence/user_repository.go` and `model/user.go`, read 2026-09-23
through the GitHub API at tag `v0.63.2`). That is case-insensitive
equality with no wildcard, and it is exactly the collision
`strings.EqualFold` uniqueness prevents. The loader keeps today's ASCII-printable check. **Uniqueness compares with
`strings.EqualFold` everywhere**, at load, in `doctor`, and at write time
under the lock, because a case collision is one principal shared by two
people whatever path wrote it. So a file naming both `Alice` and `alice`
stops loading, and that is the refusal R3 wants at startup.

`@` and `+` are in the class because Ruling 2's query (done 2026-09-23,
"The conductor's actions" item 3) found the estate's one account name
email-shaped; an email-shaped name also falls under the address rule in
Global Constraints.

### R-5c7: the CLI admin grant is `identity.ChannelCommandLine`, and no verifier can emit it

> **Conductor-owed fold 5:** a new channel constant, or a separate store
> method that takes no channel? Applied as the new constant.

`AddMember` today refuses `Admin` on every channel, and its doc comment
says "The admin grant arrives together with that verb". The audit record
already takes a channel, and `spec:1141-1143` wants the CLI's record to
be "the same record". So the command line becomes one more
`identity.Channel`, `ChannelCommandLine`, whose `String` is
`command-line` and which `Channel.Valid` accepts, since the audit record
and `AddMember`'s channel check need it. The store admits the admin flag
on that channel alone. `ErrRosterAdminChannelForbidden` is renamed
`errAdminGrantForbidden`, unexported, since after this change the
refusal is about who may grant, and no caller outside `internal/auth`
branches on it (R-5c20).

> **Conductor-owed fold 12:** drop `ChannelHeader` and
> `ChannelServiceToken`? Applied as the drop. Neither has a producer:
> their only references are `roster_test.go` and the `String` and `Valid`
> switches, and the 5b-ii architecture read recommends dropping them
> until their sources land. Keeping them leaves four constants under a
> type documented as "how a verified identity was asserted": one a
> verifier emits, two nothing emits, and one that is not an assertion.
> After the drop the type holds `ChannelCookie`, the one assertion
> channel, and `ChannelCommandLine`, the audit channel, in a
> `channelNames` table that `String` and `Valid` both read. A forward-auth
> front or a service-token source adds its constant in the commit that
> asserts it. `spec:1223` lists the three channels descriptively and takes
> a one-line erratum ("constants land with the source that asserts
> them").

**An identity carrying `ChannelCommandLine` fails
`identity.Identity.Valid`, which 5c-ii's admin gate requires**. This
protects the web edge only where a caller runs `Valid`, and today no
production code does, so the gate's call and a mutation proving it
travel with the gate to 5c-ii. `Identity.Valid` accepts any valid channel today
(`internal/identity/identity.go`, `Identity.Valid`). If it accepted this
one, a later verifier that builds an `Identity` from a header or a token,
plus a handler that forwards the identity's channel into `AddMember`,
would compose into a web admin grant. The command-line channel is an
audit channel. It is never an assertion a verifier makes, so a verified
identity carrying it is invalid by type. Task 6 rewrites the whole
`Channel` doc comment, including its stale sentence "Admin-only verbs
reject a cookie-channel identity", which the add-member gate contradicts.
The same policy sentence in `internal/auth/seam.go`'s `VerifiedIdentity.Admin`
comment ("An admin verb rejects an identity whose Channel is
ChannelCookie; the one exception is the add-member handler") goes too:
after this pass no admin verb takes an `Identity`, so the comment states
only that the flag is read from the roster. `Identity.Valid`'s own
comment promises "a normalized address" while the code checks only for a
non-empty value; the comment is corrected to what the code checks.

### R-5c8: the last-admin guard is a write-path invariant with no exemption

> **Conductor-owed fold 6:** rung 5 has no demote or remove path, so what
> does the guard guard? Applied as: every roster the store writes names at
> least one admin, including the bootstrap, checked under the lock against
> the file as re-read from disk, with no demote method exported.

Configuration homes (`:668-669`) routes "a last-admin lockout guard,
atomic against concurrent self-demotion" to rung 5's plan, and rung 6's
spec (`:816-818`) expects "5c's last-admin guard" to run inside 6d's
write transaction. Geoff's 2026-09-12 refinement forbids exporting
anything ahead of its caller, so this pass builds no demote method.

**The bootstrap is one write.** An empty create routed through the
guarded function would be refused by the guard itself. `CreateRoster(ctx
context.Context, path string, first Member, actor string, logger
*slog.Logger) (*Roster, error)` now writes the first member in one create
under the directory lock. **It routes through the same locked
read-modify-write function as `AddMember`, in a mode where an absent file
is the starting state**, so the first member and the actor pass exactly
the write-path validation every later write passes: Task 2's address,
display-name, and `streamer_username` rules, the required username
(Ruling 4), and the command-line actor rule. It also requires
`first.Admin`, audits on `ChannelCommandLine`, re-checks the file's
absence under the lock, and returns `ErrRosterExists` when the file is
present. Without that validation, `member add --admin a@x --name A
--streamer-username ""` would write a bootstrap admin with no username.
Its one caller is
Task 7's verb. So the store never writes a roster without an admin, the
guard has no exemption, and exit gate 11 is literally true.

**The guard's reachable path today** is a roster a hand edit left with no
admin, or with no members at all. Adding a non-admin to it would write a
roster with no admin, and the guard refuses it. The invariant lives in
the one internal read-modify-write function every roster mutation routes
through, so a later demote or remove method inherits it by construction.
`doctor`'s roster row reports a roster that names no admin as failing,
and names the recovery (`member add --admin`).

Authorization is re-checked under the lock. On every channel but `ChannelCommandLine`, `AddMember` requires the
actor to hold the admin flag in the disk copy it re-read under the lock,
beside the admin invariant. That makes the threat-model claim "only a
member whose roster record holds the admin flag can change the roster" a
property of the store, not only of the gate. It is the atomicity rung 6's
"a concurrent self-demotion of the last two admins leaves one" needs.

### R-5c9: the public URL is a bare origin, and a link is origin plus prefix plus `/login/link`

Review finding 2's fix offers two refusals: a public URL carrying a path,
or one whose path is not the prefix. This plan takes the first.
`DUBPLATE_PUBLIC_URL`'s path must be empty or `/`, and `LinkURL` builds
`origin + prefix + "/login/link"`. One rule is simpler to state in an
error, and the prefix already expresses where dubplate lives. **5c-ii
places the consumption interstitial's route at `login/link` under the
prefix**, and its outline says so. The shell-level test this pass adds
proves a minted link's path is owned by the prefix and reaches a
registered handler rather than the fallthrough, for both `https://h` and
`https://h/`, so a hand concatenation that drops `JoinPath`'s cleaning
fails it. `Shell.Register` refuses a subtree name carrying `/`
(`internal/serve/serve.go`, `subtreeNameRefusal`), so the one handler
5c-ii registers at `login` serves both the login page and `login/link`.

### R-5c10: `doctor` names each failing source key on the source's own line

> **Conductor-owed fold 7:** review finding 11 asks for one `<KEY>
> invalid` row per failing key; 5b-ii's fold 7 dropped per-key rows to
> keep one line per source. Applied as naming the keys on the one line.

The finding's substance is that no line names the offending key. The
source's own line can carry the names, for example `magic-link
unsatisfied: DUBPLATE_SESSION_SECRET_PREVIOUS invalid`, with key names
only and never a value.

**How `doctor` gets the names.** Two routes feed the line. A missing or
invalid Required or Optional key already arrives in `doctor`'s per-key
statuses. A relational or format refusal arrives from the source's
`Validate` hook, which `identitySourceStatus` reruns directly and which
returns an `errors.Join` of messages that each begin with a key
(`cmd/dubplate-server/auth.go`, `validateMagicLinkSource`). Parsing that
prefix breaks silently the day a message is reworded. So Task 5 adds a
small keyed error type in `cmd/dubplate-server` (a key and a reason; its
`Error` renders `KEY: reason`), `validateMagicLinkSource` and
`validatePublicURL` return it, and `doctor` collects every key by
walking the join and calling `errors.As` on each member. One `errors.As`
over the whole join would not do: `go doc errors.As` states it finds "the
first error in err's tree", so a single call returns one key. The walk
reads the join's members through its `Unwrap() []error` method: `go doc
errors.Join` states "A non-nil error returned by Join implements the
Unwrap() []error method." A relational refusal names both keys it compares, threshold keys
included, because the relation is the source's. `internal/config` stays
untouched (R-5c17). That keeps the line-count derivation 5b-ii's
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

**Applies in 5c-ii Task 1**, which builds the gate; kept here so the
reasoning travels with the tasks that moved.

The gate verifies the session through the seam, then reads `Admin` from
the verified identity (which the seam read from the roster). An absent
cookie, a cookie that fails verification, a revoked session, a member
removed from the roster, a roster store failure, and a verified
non-admin all answer the same 404 (`spec:1418-1420`, risk 14). The page's
existence is not a fact any of them learns. The reason reaches the
server log alone, and the log never carries the cookie value.

**"The same 404" means the shell's own not-found answer.** The owned
terminator answers an unregistered path under the prefix with
`http.NotFound` (`internal/serve/prefix.go`, `terminated`). The gate's
refusal is that same call, so a refused request to `<prefix>/roster` and
a request to an unregistered `<prefix>/<anything>` match in status, in
every header except `Date`, and in body. No refusal carries `Set-Cookie`.
Do not compare refusal rows only with each other: a gate answering its
own JSON 404 passes that while still telling a non-admin the route
exists. When 5c-ii's served 404 document lands (W3.7), the
gate's refusal tracks it; the 5c-ii outline carries that.

The gate is the session-refresh call site `docs/STATUS.md` carries. It
passes `sameOrigin` as true only when `Sec-Fetch-Site` is exactly
`same-origin`. The seam's own contract leaves that derivation to the
caller (`internal/auth/seam.go`, `Seam.VerifyIdentity`'s comment). The
header's other values (`same-site`, `cross-site`, `none`, per the MDN
reference `go doc net/http.CrossOriginProtection` links) and its absence
refresh nothing. The origin check that already wraps the shell does not
cover a GET: `go doc net/http.CrossOriginProtection` states "The GET,
HEAD, and OPTIONS methods are safe methods and are always allowed". So
the gate cannot infer same-origin from having been reached. **The gate
sets the refreshed cookie only after the admin check passes**, through
`SessionCookie` with the attributes 5b-ii's review found sound. The seam
refreshes a verified non-admin's session too, and a gate that set the
cookie on successful verification would put a `Set-Cookie` on that
non-admin's 404, which an unregistered path never carries.

### R-5c13: the roster API takes a form, reads no admin field, and refuses one (Ruling 5)

**Applies in 5c-ii Task 1.** In 5c-i, Ruling 5 governs Task 6's rule that
only the command line sets the flag.

A checkbox the server always refuses is a control that cannot work.

**The POST body is form-encoded**, read through the same
`http.MaxBytesReader`-then-`ParseForm` path the 5b-ii handlers use
(`internal/auth/handlers.go`, `readForm`), and the responses are JSON.
The API reads its three fields from `r.PostForm` only, on 5b-ii's
precedent that nothing is read from a query string. **An `admin` key
present anywhere, in the body or the query string (`r.Form`, which
`go doc net/http.Request` describes as "the parsed form data, including
both the URL field's query parameters and the PATCH, POST, or PUT form
data"), whatever
its value, empty included, is refused** with its own outcome and no
write. Refusing the query-string form too costs nothing and leaves no
spelling of the field that reaches the store. The store's channel refusal stays as the second net.
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

**The actor is looked up from the effective uid.** The box's binary is
built with `CGO_ENABLED=0` (`Makefile`, `build-linux`), so `os/user` runs
its pure-Go path. There, `user.Current` falls back to `os.Getenv("USER")`
when the passwd lookup fails and builds from `os.Getuid()`, the real uid
(`src/os/user/lookup_stubs.go`, read 2026-09-23). The ownership check
uses the effective uid, so the actor must too: the verb calls
`user.LookupId(strconv.Itoa(os.Geteuid()))` and falls back to
`uid:<n>`, never to `$USER`.

**How an operator runs it is 5d's runbook.** `config.Load` requires every
key the source requires, secrets included, and on the box those live in
the unit's environment file, readable by root, while the verb must run as
the directory's owner. The natural workaround,
`sudo -u music env $(cat ...) dubplate-server member add ...`, puts every
secret into argv, where any local user reads it through
`/proc/*/cmdline`. Task 8 carries a 5d runbook row requiring an
invocation that keeps secrets out of argv. A transient unit (`systemd-run`
with the service's user and environment file) is the candidate, handed to
5d as a claim to prove rather than a recipe: whether a transient unit
accepts an environment-file property, and whether it needs flags to
return the verb's output to the terminal, is unverified here. Narrowing the verb's load to the keys it reads is
refused (fold record): the full load is what catches a misconfigured
environment before the first admin is written.

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

### R-5c18: review minor 6's code fix is refused, and its runbook sentence goes to 5d

> **Conductor-owed fold 11:** the 5b-ii close draft routes minor 6 to 5c
> (`draft-HISTORY-5b-ii.md`, the disposition row "Minor 6"), though
> Geoff's routing list does not name it. Applied as a refusal of the code
> fix, with the runbook half carried.

Minor 6 says that in kind mode the global ceiling bounds no enumeration,
and that once the ceiling is spent a known address answers 429 while an
unknown one answers `unknown`. The kind answer discloses known against
unknown by design (fork 4), so the post-ceiling signal adds little. The
review's fix, a global request ceiling answering every request 429 once
spent, is itself a cheaper group-wide lockout than the signal it closes.
So the code fix is refused. The review's second half, "kind mode's
enumeration bound is the per-client key alone", becomes a 5d runbook
sentence, carried in Task 8's handoff.

### R-5c19: the lock waits carry named deadlines, and readers never wait behind a writer's lock wait

Two constants bound the cross-process wait: **`memberVerbWriteWait`, 30
seconds**, in `cmd/dubplate-server`, which the verb applies (Task 7); and
**`rosterAPIWriteWait`, 5 seconds**, in `internal/auth`, which 5c-ii's
roster API applies to its own `context.WithTimeout` so its 503 `busy` row
is reachable (an incoming request's context carries no deadline: `go doc
net/http.Request.Context` says it is canceled only "when the client's
connection closes, the request is canceled (with HTTP/2), or when the
ServeHTTP method returns", so a lock wait needs its own bound). A writer takes the
directory `flock` **before** it takes the in-process mutex that `Lookup`
and `List` share, and holds the mutex only for the in-memory swap. So a
writer waiting on another process never blocks a login or a gate check in
its own. Two writers in one process exclude each other through `flock`
itself, since each opens its own descriptor (quoted in Task 1).

### R-5c20: nothing is exported ahead of a cross-package caller

Geoff's 2026-09-12 refinement, applied at the root rather than per
symbol. A new sentinel or
helper is exported only when a caller outside its package branches on it
or calls it. So the per-refusal errors Task 2 adds, the no-admin and
actor-not-admin errors, the busy error, and the renamed grant error
(`errAdminGrantForbidden`) are unexported: their only branching reader,
5c-ii's roster API, lives in `internal/auth`, and the verb prints an
error's text and branches on none. Exported in 5c-i: `CreateRoster` and
`identity.ChannelCommandLine`, both of which the verb calls or passes,
and `ErrRosterExists`, which it branches on. `RequireAdmin` and
`NewRosterAPI` are born exported in 5c-ii, where the document route and
the serve wiring in `cmd/dubplate-server` call them.

---

## The threat model, as the in-pass reviewer's rubric

The spec's threat model (`spec:1909-1980`) governs. The claims this pass
must make true, each with the task that proves it:

- **No channel but the command line can set the admin flag, and the
  store refuses a web write whose actor is not an admin in the disk
  copy** (Task 6). The web write path itself, the gate and the roster API,
  lands in 5c-ii and inherits both refusals.
- **A roster change that fails validation leaves the file
  byte-identical, and every roster change that succeeds writes one audit
  record naming the actor and the channel** (Tasks 1, 2, 6, 7), the
  bootstrap included, across two writer processes (Task 1).
- **A member removed by hand stops resolving at the running server's
  next request** (Task 1), which closes the gap in claim 8 that finding 1
  opened.
- **Neither a single address nor a single client can drive unbounded
  outbound mail, and one IPv6 /64 is one client** (Task 3).
- **A token never reaches a log record or a process outside dubplate**
  (Task 4: the link lands inside the owned prefix).
- **No operator secret reaches a log** (Task 4: the public URL's password
  is never echoed).

**The residuals this pass adds, stated rather than hidden:**

- A hand edit of the roster writes no audit record.
- A hand edit that takes no lock and lands inside a writer's
  milliseconds-long read-modify-write window is overwritten by that
  write. The lock excludes every writer that takes it; a text editor
  does not.
- A malformed hand edit takes every signed-in member offline until it is
  fixed, because the reload fails closed (R-5c5). `doctor` names the
  failure and is the recovery path.
- No path but a hand edit promotes an existing member to admin:
  `member add --admin` on an existing address is a duplicate, and the web
  never grants (Ruling 5). Whether a grant verb waits for rung 6 is rung
  6's question, recorded here so it is not rediscovered.
- The CLI's audit record reaches the invoking terminal rather than the
  service journal (R-5c15).
- The per-address map admits untracked keys past capacity, a path that
  needs a roster of more than 10,000 members (R-5c3).
- An admin's shared device can add a member, as the spec already records.

---

## Review focus

The five input classes most likely to bite a person that no task's
happy path exercises, each pinned by a test in the task that owns it.

1. **Two writers at once** (two `Roster` values over one file, standing in
   for `serve` and the verb), with a hand edit landing between their
   writes. Expected: nothing lost, nothing
   resurrected. A hand edit landing inside one write's window is the
   residual above, not a promise. Pinned in Task 1.
2. **A roster file an operator broke by hand while `serve` runs.**
   Expected: logins refuse with a named error rather than serving the
   stale copy or answering "unknown". Pinned in Task 1.
3. **An address or display name carrying an invisible character** (a
   zero-width space, a right-to-left override, a no-break space).
   Expected: refused at every entry point. Pinned in Task 2.
4. **A second `member add --admin` on an install that already has a
   roster.** Expected: the member is added and the file is never
   recreated. Pinned in Task 7.
5. **A `member add --admin` whose flags are present but empty**, for
   example `--streamer-username ""`. Expected: refused by the write-path
   validation with nothing created, the bootstrap included. Pinned in
   Task 6. (The non-admin probe of the roster API moves with that API to
   5c-ii.)

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor before any
implementer. Reads the spec, the security review, the 5b-ii close
drafts, and this plan against the tree as shipped. **Its output is
blockers with evidence, not fixes.** It writes no code and edits no
file. The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports that
HEAD.** This plan's code pointers were resolved at `27cd7b5` (the Go
tree is unchanged since `08befea`) and its spec pointers at `cd2b5ca`.
Grade a moved pointer by content: a pointer whose content survives at
any offset is a note naming the new line, and a pointer whose content is
gone is a blocker.

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
   report the symbol and the current line it lives at. Minor 6, the
   fourteenth finding the close draft routes to 5c, is refused by R-5c18;
   report whether the folded HISTORY still routes it to 5c, so the scope
   ledger stays complete. *(Expected, all
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
   constants, the shape 5c-ii Task 2 extends (this pass's Task 4 adds only
   the prefix to the handler config).
8. Report `doctor_test.go`'s line-count test by name and the rows it
   derives from, the property R-5c10 preserves.
9. Report `newRootCmd`'s `AddCommand` block, which Task 7 extends.
10. Report every caller of `(*Roster).AddMember`, `(*Roster).List`, and
    `ErrRosterAdminChannelForbidden` in the module, since Tasks 1, 2, and
    6 change all three.
11. Report the spec lines this plan cites for fork 1's "only path" text
    (`:122-124`, `:188-192`) and the admin checkbox (`:1417`), the two
    passages R-5c13 turns on.
12. Confirm the plan header records all six rulings as yes and Ruling
    2's query as done, and that R-5c6's shape admits `@` and `+`. Do not
    run the query again.
13. Name 5b-ii's test that proves no consumption log record carries the
    token (in `internal/auth/handlers_test.go`), which the exit gate
    re-runs.

**Deliverable count: 1** (the report).

---

## Task 1: The roster across processes, and the strict loader

Model: `opus`. Depends on Task 0.

Review major 1 and minor 7. The roster's own correctness, fixed first,
because every later task reads or writes it.

**Files:**
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`
  (the request-link handler's store-failure branch and `adminContact`,
  both callers of the changed `Lookup` and `List`, R-5c5)
- Modify: `internal/auth/seam.go` only if `Lookup`'s changed error set
  needs a caller change (it should not: the seam already refuses any
  non-nil error)

**Interfaces:**
- Consumes: `internal/atomicfile.WriteStream`; `syscall.Flock`.
- Produces: `(*Roster).AddMember(ctx context.Context, candidate Member,
  actor string, channel identity.Channel, clientIP string, logger
  *slog.Logger) error`, `ctx` first because the write can now wait on
  another process. `CreateRoster(ctx context.Context, path string)`
  gains `ctx` for the same reason; Task 6 then changes it to write the
  first admin (R-5c8). `(*Roster).List() ([]Member, error)`, which gains
  an error because a reload can fail. `(*Roster).Lookup` keeps its
  signature and gains the reload errors. An unexported busy error for a
  lock wait that outlives `ctx` (R-5c20). One unexported locked
  read-modify-write function that every roster mutation routes through,
  which Task 6's guard lives in.

**Outcomes:**

- **Every write takes an exclusive `flock` on the roster file's
  containing directory first**, then re-reads and validates the file from
  disk, merges the candidate into the disk copy rather than the memory
  copy, writes through `atomicfile.WriteStream`, re-reads, swaps the
  in-memory copy under the in-process mutex, and only then releases the
  `flock`. Why the directory: an atomic rename replaces the file's inode,
  so a lock on the file itself would not exclude a writer that opened the
  new inode. Locking the directory creates no file in the leaf. `man 2
  flock` (read 2026-09-23): "If a process uses open(2) (or similar) to
  obtain more than one file descriptor for the same file, these file
  descriptors are treated independently by flock()", so two `Roster`
  values in one test process exclude each other exactly as two processes
  do. The same page: "A shared or exclusive lock can be placed on a file
  regardless of the mode in which the file was opened", so a read-only
  open of the directory suffices. All three properties were measured
  2026-09-23 (a directory `flock` succeeds, a second descriptor gets
  `EWOULDBLOCK`, and the lock holds across a child's rename); opening the
  directory needs read permission on it.
- **The lock order keeps readers off a writer's wait** (R-5c19): the
  `flock` comes before the in-process mutex, and `Lookup` and `List` take
  only the mutex, so a writer waiting on another process never blocks a
  login in its own.
- **The lock wait is bounded by `ctx`.** The write polls `LOCK_EX|LOCK_NB`
  at a short fixed interval (the salvage lock package's shape, adopted in
  reasoning) and returns the busy error when `ctx` ends first. It never
  blocks without bound: `flock(2)` states "A call to flock() may block if
  an incompatible lock is held by another process", and a blocking call
  cannot observe `ctx`. The deadlines themselves are Task 7's and 5c-ii Task 1's
  named constants (R-5c19).
- **`CreateRoster` runs its existence check and its write under the same
  directory lock.** Today it stats and then writes, so two concurrent
  bootstraps can both see "absent" and the second overwrites the first.
  Under the lock, the second sees the first's file and returns
  `ErrRosterExists`.
- **`Lookup` and `List` reload when the file changed** (R-5c5): device
  and inode through `os.SameFile`, size, and modification time. **Every
  read opens the file, takes `(*os.File).Stat` on that descriptor, reads
  from that same descriptor, and caches that stat with that content**,
  the writer's post-write re-read included. A cache whose modification
  time is within one second of the moment it was filled is unverified,
  and the next call re-reads (R-5c5's one-second rule).
- **A reload that fails returns the failure** (absent, unreadable,
  unparseable, invalid), each wrapping the existing sentinel with the
  path. The request-link handler answers `failed` under the kind answer
  when `Lookup` returns anything but `ErrMemberNotFound`, and when
  `adminContact`'s `List` fails; it keeps its generic-answer bytes
  unchanged.
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
| Load | an unknown member key, a case-variant key (`"ADMIN"`), a duplicate key, invalid UTF-8, trailing data | `ErrRosterUnparseable` |
| Load | a member value of `null` (`"admin":null`, `"streamer_username":null`) | loads, as the zero value (measured; no refusal test is written for it) |
| Load | `"members":[]` with version 1 | loads, empty |
| Reload | stat unchanged, cache verified | cached copy, no read |
| Reload | stat unchanged, cache filled within one second of the file's modification time | re-read once |
| Reload | changed, valid | reloaded; a hand-removed member stops resolving |
| Reload | changed, invalid or unparseable | the load error; `Lookup` refuses |
| Reload | file removed | `ErrRosterMissing` |
| Reload | replaced with identical bytes (new inode) | reloaded, same answers |
| Write lock | free | acquired |
| Write lock | held by another `Roster` value past `ctx` | the busy error, file unchanged |
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
- **The cache-hit row is observable.** Age the file first: `os.Chtimes`
  it to T0, more than one second in the past, before `OpenRoster`, so the
  cache is verified from the start. Then rewrite it in place with the
  same inode and the same size, restore its modification time to T0, and
  assert the old answer persists. Only a served cache explains that. A
  second row skips the aging, so the cache fills within one second of the
  file's modification time, repeats the same in-place rewrite and
  restore, and asserts the new answer, which proves the one-second rule.
- **The stat-and-read order is a reviewer-verified criterion**, because
  no test can force the rename to land between a read and a stat. The
  diff-reviewer confirms the cached stat is always the fstat of the
  descriptor that was read. The change check may stat the path; no path
  reads through `os.ReadFile` or a second descriptor for content it
  caches.
- A concurrent pair of `CreateRoster` calls, run with `-race`, leaves
  exactly one file whose contents are the winner's, and exactly one
  `ErrRosterExists`.
- A held-lock test holds the directory lock through a raw `flock` on the
  directory in the test and asserts the busy error under a short `ctx`, a
  byte-identical file, and a `Lookup` on the same `Roster` value that
  returns while the writer is still waiting. **Fixture state:** the
  writer's goroutine started and blocked on the lock before the
  `Lookup`, so its return proves the lock order.
- Each loose variant in finding 7 is a table row asserting refusal by
  sentinel, including `"ADMIN": true` and the duplicate `"admin"` key.
- The request-link test presents a known address against a roster file
  made unparseable after open, under the kind answer, and asserts
  `failed` and not `unknown`, with no admin contact in the body; an
  unknown address against the same broken file also answers `failed`
  (the `adminContact` path); under the generic answer both assert the
  same bytes as a normal answer.
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
9. Stat the path after reading it with `os.ReadFile` → no test fires;
   report it as the reviewer-verified row and name the race it opens.
10. Drop the one-second rule → the inside-the-window row serves the
    old answer.
11. Take the in-process mutex before the `flock` → the held-lock row's
    `Lookup` blocks until the writer's `ctx` ends.

**2.0 hook: extends hook 5's write half** (`spec:1848`): the one
read-modify-write function is where rung 6 mints the surrogate id inside
its transaction, and the stat-and-reload read side is what rung 6's store
replaces with a query.

**Prior art:** adopts the salvage lock package's poll-under-`ctx` shape
(`_salvage/internal/lock/lock.go`, `AcquireBlocking`) in reasoning, with
stdlib `syscall.Flock` in place of `golang.org/x/sys/unix` (R26). Adopts
`encoding/json/v2`: Go team, shipped in go1.27.1's standard library, but
experiment-gated behind `goexperiment.jsonv2`, which is on in go1.27's
baseline (Global Constraints). The adoption is judged worth that
stability class, since v1 cannot refuse the case-variant and duplicate
keys finding 7 lists. The one-second modification-time window is this
plan's own rule (R-5c5). Declines a file watcher (`fsnotify`): a module
dependency R26 does not budget, and a stat per lookup is cheap at group
scale, as the review itself notes.

**Independent: no.** Every later roster task builds on it.

**Deliverable count: 4** (the cross-process write with its lock, the
locked create, the reload with its store-failure answer, the strict
loader).

---

## Task 2: Validation tightening, and one error per refusal

Model: `sonnet`. Depends on Task 1. Rulings 2, 3, and 4 shape it.

Review minors 8, 13, and 15, plus Ruling 4's write-path requirement,
plus the per-refusal errors the roster API maps to specific outcomes
(`spec:1421-1424`).

**Files:**
- Modify: `internal/identity/identity.go`, `internal/identity/identity_test.go`
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `config/roster.example.json` (its second member gains a
  `streamer_username`, so the reference file models the rule the write
  path enforces)

**Interfaces:**
- Consumes: Task 1's locked write function.
- Produces: unexported errors (R-5c20), each wrapping
  `ErrRosterRecordInvalid` so every existing `errors.Is(err,
  ErrRosterRecordInvalid)` still holds: one each for a duplicate address,
  a duplicate streamer username, an invalid member address, an invalid
  display name, an invalid streamer username, a missing streamer
  username, and an invalid audit actor or channel. 5c-ii Task 1, in the same
  package, maps each to an outcome.

**Outcomes:**

- **`identity.NormalizeAddress` refuses any rune for which
  `unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs,
  unicode.Zl, unicode.Zp)` holds** (the review's own fix for finding 8).
  `go doc unicode` lists `Cf` as "the set of Unicode characters in
  category Cf (Other, format)". The ordinary space is in `Zs` (measured:
  `unicode.In(' ', unicode.Zs)` is true), which changes nothing here,
  since the normalizer already refuses every rune at or below U+0020. The
  normalizer stays a fixed point.
- **Display names refuse a narrower set** (Ruling 3): `unicode.IsControl`,
  `Zl`, `Zp`, `Zs` **except U+0020**, and `Cf` **except U+200C and
  U+200D**. Do not apply the address set to display names: it refuses
  every name with a space in it, "Alice Example" included.
- **`streamer_username` on the write path matches
  `[A-Za-z0-9][A-Za-z0-9._@+-]*`** within the existing 64-byte bound
  (Ruling 2, R-5c6); the loader keeps today's ASCII-printable check.
  **Uniqueness compares with `strings.EqualFold`** at load, in `doctor`'s
  collision check, and at write time inside the lock.
- **The write path requires `streamer_username`** (Ruling 4). The loader
  does not: a roster without one still loads.
- **`AddMember` refuses before any write** when `!channel.Valid()`, and
  when the actor does not meet its channel's rule. For a network channel
  the actor must normalize through `identity.NormalizeAddress`. For
  `ChannelCommandLine` (Task 6), the actor must be non-empty and carry no
  rune from the address set, the ordinary space included. Until Task 6
  lands, only the network rule is reachable. Minor 13's fix as written
  ("refuse an actor that does not normalize") would refuse the
  operating-system user `spec:1141-1143` puts in the CLI's actor field,
  hence the per-channel rule.
- **Each refusal wraps its own error**, and no error string names the
  refused value.

**States each check meets:**

| Input | Field | Report |
| --- | --- | --- |
| U+0085, U+009B, U+202E, U+00A0, U+200B inside | address | refused (invalid address) |
| `a@\u212Aexample.com` (Kelvin sign) | address | accepted, normalized to `a@kexample.com` |
| `Alice Example` | display name | accepted |
| U+00A0, U+202E, U+200B inside | display name | refused (invalid display name) |
| U+200C or U+200D inside | display name | accepted (Ruling 3) |
| `a_b`, `A.b-c`, `x9` | streamer username, write | accepted |
| `a@example.test`, `a+b@example.test` | streamer username, write | accepted (Ruling 2's estate check found an email-shaped name) |
| `.x`, `..`, `-rf`, `_x`, `a%`, `a b`, non-ASCII | streamer username, write | refused (invalid streamer username) |
| `Alice` beside an existing `alice` | streamer username, write and load | refused (duplicate) |
| absent | streamer username, write | refused (missing) |
| absent | streamer username, load | loads |
| `""`, a malformed actor on the cookie channel, `Channel(0)` | audit actor, channel | refused (invalid audit actor), file unchanged |

**Acceptance criteria:**

- Every row above is a table case. A zero-width variant of an existing
  member's address is refused, not added as a distinct member. **Fixture
  state:** a roster already holding the plain address.
- **The Kelvin-sign row is pinned to acceptance.** `strings.ToLower` maps
  U+212A to `k` (measured 2026-09-23: `strings.ToLower("a@\u212Aexample.com")`
  is `"a@kexample.com"`, and `net/mail.ParseAddress` accepts the
  unlowered form). The stored and mailed form is the ASCII owner's
  address, so a lookalike typed at the login page mails the real owner.
- Every error satisfies `errors.Is(err, ErrRosterRecordInvalid)`.
- `config/roster.example.json` still loads through `OpenRoster` in the
  existing example-file test. `bash scripts/check.sh` green.

**Named mutations:**

1. Drop `unicode.Cf` from the address set → the U+202E and U+200B rows
   admit.
2. Refuse all of `Zs` in display names → the `Alice Example` row fails.
3. Refuse all of `Cf` in display names → the U+200D row fails.
4. Compare `streamer_username` with `==` → the `Alice`/`alice` row
   admits.
5. Drop the alphanumeric first character → the `.x` and `-rf` rows
   admit.
6. Drop `@` from the character class → the email-shaped row refuses.
7. Make the write path's requirement a load-time requirement → the
   absent-at-load row fails.
8. Skip the channel validity check → the `Channel(0)` row writes.
9. Wrap a new error without `ErrRosterRecordInvalid` → the `errors.Is`
   row fails.

**2.0 hook: lands hook 5's "required" half early on the write path**
(S20 hook 5: "`streamer_username` becomes required, unique, and
stable"). Rung 6's 6d makes it required at load.

**Prior art:** none consulted beyond the stdlib `unicode` tables. The
address set is the review's own fix.

**Independent: yes, of Task 3 by files** (`identity.go`, `roster.go`,
the example file against `limit.go`, `handlers.go`). Both follow Task 1.

**Deliverable count: 4** (the invisible-character refusal across
addresses and display names, the principal's shape and case-folded
uniqueness, the write-path requirement, the actor and channel refusal
with the per-refusal errors).

---

## Task 3: The limiter fold

Model: `opus`. Depends on Task 1 (both touch `handlers.go`). Ruling 1
shapes it.

Review majors 3 and 4, and minor 17, on the design R-5c3 states.

**Files:**
- Modify: `internal/auth/limit.go`, `internal/auth/limit_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`

**Interfaces:**
- Consumes: `net/netip`; `golang.org/x/time/rate` (already a dependency).
- Produces: `(*Limiter).AllowLinkRequest` splits into two unexported
  methods the handler calls at two points: `allowLinkClient(now
  time.Time, client string) error`, charged first for every request, and
  `allowLinkAddress(now time.Time, address string) error`, charged after
  the roster lookup for a known address only. Only
  `internal/auth/handlers.go` calls the limiter's methods (checked at
  `cd2b5ca`; `cmd/dubplate-server/serve.go` only constructs it), so the
  new names stay unexported (R-5c20). `AllowConsumption` and
  `AllowLinkMail` keep their signatures.

**Outcomes:**

- **The request-link handler's order becomes:** cap and parse the body;
  charge the client (`allowLinkClient`); normalize the address; look it
  up in the roster; for a known address only, charge the address
  (`allowLinkAddress`); then send, charging the global mail ceiling as
  today. A client already refused never reaches the address map, and an
  address that fails to normalize has already spent the client's token
  (finding 17). The handler still answers `invalid_address` for a bad
  address and still logs no address.
- **A known address over its per-address limit** answers 429 under the
  kind answer, and answers `requested` and sends nothing under the
  generic answer, so the generic answer's bytes stay uniform across
  known, unknown, and over-limit (R-5b28 holds). An unknown address never
  meets the per-address limit, so a flood of made-up addresses never
  touches the map (finding 3, R-5c3).
- **At capacity, the per-address map never evicts an unrefilled bucket.**
  It evicts the least recently used bucket that has fully refilled, and
  if none has, admits the request without tracking the new key.
  "Fully refilled" is `bucket.TokensAt(now) >= float64(bucket.Burst())`:
  `go doc golang.org/x/time/rate.Limiter.TokensAt` states it "returns the
  number of tokens available at time t", and `Burst` returns "the maximum
  burst size". The comparison reaches exact equality once a bucket
  refills (measured 2026-09-23). The scan starts at the least recently
  used end and stops at the first match. It cannot stop at the least
  recently used bucket alone, because refill time depends on tokens
  spent, not on recency; the doc comment says so and states the scan's
  worst case against `maxLimiterKeys`. With known-only charging this path
  needs a roster of more than 10,000 members.
- **The per-client and consumption maps keep least-recently-used
  eviction**, unchanged. The keyed map takes its at-capacity policy as a
  construction value, so one implementation serves both policies.
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
| Per-address map | key present | charge it |
| Per-address map | below capacity, key absent | create it full, charge it |
| Per-address map | at capacity, a refilled bucket exists | evict the least recently used refilled bucket, create, charge |
| Per-address map | at capacity, none refilled | admit untracked; nothing evicted |
| Client-keyed maps | at capacity, key absent | evict the least recently used bucket, create, charge |
| Client key | IPv4 `192.0.2.7` | `192.0.2.7` |
| Client key | IPv4-mapped `::ffff:192.0.2.7` | `192.0.2.7` |
| Client key | IPv6 `2001:db8:1:2:3:4:5:6` | `2001:db8:1:2::/64` |
| Client key | `fe80::1%eth0` | `fe80::/64` |
| Client key | empty or unparseable | the shared invalid-key bucket |
| Link request | client refused | 429; the address map unchanged |
| Link request | client admitted, address invalid | `invalid_address`; the client charged |
| Link request | client admitted, address unknown | kind: `unknown`; generic: `requested`; the address map unchanged |
| Link request | client admitted, known address spent | kind: 429; generic: `requested`, no send |

**Acceptance criteria:**

- The review's eviction reproduction is a handler-level test: exhaust a
  victim's known address at instant T, flood 10,000 distinct unknown
  addresses at the same T from as many distinct /64s (so the per-client
  order cannot explain the result), and the victim's next request at T
  is still refused, with the address map's size unchanged by the flood.
  **Fixture state:** the victim's bucket spent before the flood, at the
  same instant, so no refill can explain a pass.
- **The at-capacity policy is tested on the keyed map in isolation**, not
  through the handler, which would fill the per-client map first
  and exercise the wrong refusal: a
  per-address-policy map filled with unrefilled buckets admits a new key
  untracked, keeps its size at capacity, and evicts nothing; after a full
  interval the least recently used refilled bucket is the one evicted; a
  client-policy map at capacity evicts its least recently used bucket.
- Under the generic answer, a known address over its limit and an
  unknown address produce byte-identical responses, and no send is
  attempted for the over-limit known address (a recording `Send`
  asserts it).
- The review's IPv6 reproduction is a test: 100 addresses inside one /64
  against a consumption limit of 1 admit exactly 1.
- `bash scripts/check.sh` green; the limiter tests run with `-race`.

**Named mutations:**

1. Charge the address before the lookup, for every address → the
   unknown-flood row's map-size assertion fails.
2. Evict the least recently used bucket unconditionally on the
   per-address map → the isolated at-capacity row evicts a spent bucket.
3. Refuse the new key at capacity → the isolated
   row refuses instead of admitting.
4. Answer 429 under the generic answer for a known address over its limit
   → the generic byte-identity row fails.
5. Key IPv6 on the full address → the /64 row admits 100.
6. Skip `Unmap` → the IPv4-mapped row keys separately from its IPv4
   form.
7. Normalize before charging the client → the invalid-address row
   leaves the client uncharged.

**2.0 hook: neutral.** S20's `golang.org/x/time/rate` verdict is for the
same shape; nothing here is schema or driver.

**Prior art:** keeps `golang.org/x/time/rate` (adopted at 5b-ii, R26).
Declines an LRU library, as 5b-ii did.

**Independent: yes, of Task 2 by files.**

**Deliverable count: 3** (the split charging order with known-only
address charging, the per-policy eviction, the /64 keying).

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
- Modify: `cmd/dubplate-server/doctor_test.go` (the lifetime-above-cap
  row proved through the existing doctor harness)

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
  pass cannot be explained by a 404 before the fallthrough. The test runs
  for both public-URL forms, `https://h` and `https://h/`, and asserts
  the same link path for each.
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
- Modify: `cmd/dubplate-server/auth.go`, `cmd/dubplate-server/auth_test.go`
  (the keyed error type and its use in `validateMagicLinkSource` and
  `validatePublicURL`)
- Modify: `internal/auth/nonce.go`, `internal/auth/nonce_test.go`

**Outcomes:**

- **The selected identity source's one line names every failing key**
  (R-5c10), key names only, never a value, in the form `magic-link
  unsatisfied: <KEY> invalid[, <KEY> missing ...]`. Keys come from two
  routes: the per-key statuses for a missing or invalid Required or
  Optional key, and a keyed error type (a key and a reason, unexported,
  in `cmd/dubplate-server`) that the `Validate` hook returns and `doctor`
  collects by walking the join and calling `errors.As` on each member
  (R-5c10). A relational refusal names
  both keys it compares. The keyed type's `Error` still renders
  `KEY: reason`, so every verb's existing refusal text is unchanged.
  `TestDoctorCmd_LineCountFollowsRegistry` passes unmodified.
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
3. Parse the key from the message prefix instead of the keyed type → a
   test with a refusal whose reason begins with a colon-bearing phrase
   prints that phrase as a key.
4. Skip the mtime fallback on a read error → the unreadable row keeps
   the record forever.

**2.0 hook: neutral.**

**Prior art:** none.

**Independent: the nonce half is independent of every task by files;
the `doctor` half depends on Task 4.**

**Deliverable count: 2** (the keyed failing-key report, the
unreadable-record sweep).

---

## Task 6: The CLI admin grant, the channel set, the validated bootstrap, and the last-admin guard

Model: `opus`. Depends on Tasks 1 and 2. Ruling 5 shapes it.

**Files:**
- Modify: `internal/identity/identity.go`, `internal/identity/identity_test.go`
  (the channel set, `Identity.Valid`'s refusal of the command-line
  channel and its corrected comment, the rewritten `Channel` comment)
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Modify: `internal/auth/seam.go` (the `VerifiedIdentity.Admin` comment
  only) and `internal/auth/seam_test.go` (the no-command-line-channel
  proof)
- Modify: `cmd/dubplate-server/doctor.go`, `cmd/dubplate-server/doctor_test.go`
  (the no-admin roster row)

**Interfaces:**
- Produces: `identity.ChannelCommandLine`, `String` `command-line`,
  accepted by `Channel.Valid` and refused by `Identity.Valid`.
  `identity.ChannelHeader` and `identity.ChannelServiceToken` removed
  (fold 12). `errAdminGrantForbidden`, unexported, replacing
  `ErrRosterAdminChannelForbidden` at every reference in the module.
  `CreateRoster(ctx context.Context, path string, first Member, actor
  string, logger *slog.Logger) (*Roster, error)`, the validated
  single-write bootstrap (R-5c8). Two unexported errors, each wrapping
  `ErrRosterRecordInvalid` (R-5c20): the no-admin error ("the roster must
  name at least one admin") and a distinct actor-not-admin error ("the
  actor is not an admin").

**Outcomes:**

- **The channel set holds two constants** (fold 12): `ChannelCookie` and
  `ChannelCommandLine`, named through one `channelNames` table that
  `String` and `Valid` both read. The roster tests that used
  `ChannelHeader` or `ChannelServiceToken` move to `ChannelCookie`.
- **`AddMember` admits `Admin: true` on `ChannelCommandLine` and refuses
  it on `ChannelCookie`** with `errAdminGrantForbidden` (R-5c7). The
  refusal runs before any lock or write.
- **`identity.Identity.Valid` refuses `ChannelCommandLine`** (R-5c7), and
  its comment is corrected to what it checks (a non-empty address, not "a
  normalized address"). **The whole `Channel` doc comment is rewritten**:
  the command line asserts an operating-system user by process
  ownership, it is an audit channel no verifier emits, and the stale
  sentence "Admin-only verbs reject a cookie-channel identity" goes.
  **`seam.go`'s `VerifiedIdentity.Admin` comment** drops its policy
  sentence and states only that the flag is read from the roster.
- **The seam never produces `ChannelCommandLine`**: a seam test asserts
  the channel of every verified identity is `ChannelCookie`.
- **`CreateRoster` writes the first member in one create, through the
  locked read-modify-write function** with an absent file as its starting
  state (R-5c8). So `first` and `actor` pass exactly the validation
  `AddMember` runs (Task 2's address, display-name, and
  `streamer_username` rules, the required username, and the command-line
  actor rule), and then the create requires `first.Admin`, audits on
  `ChannelCommandLine` with `actor`, re-checks the file's absence under
  the lock, and returns `ErrRosterExists` when the file is present. It
  never writes an empty roster.
- **Every roster the store writes names at least one admin** (R-5c8),
  checked inside that same function against the disk copy (or the empty
  starting state), before the write, with no exemption. A violating write
  returns the no-admin error and changes no byte. No demote or remove
  method is added.
- **On `ChannelCookie`, `AddMember` requires the actor to hold the admin
  flag in the disk copy** it re-read under the lock (R-5c8), and refuses
  otherwise with the actor-not-admin error. 5c-ii's roster API is the
  first production caller of this branch; this task proves it in tests.
- **`doctor`'s roster row fails for a roster that names no admin**,
  reading `no admin; run dubplate-server member add --admin`. `doctor`
  counts admins through `List` itself; it needs no exported error.

**States each check meets:**

| Write | Disk roster before | Candidate, channel, actor | Report |
| --- | --- | --- | --- |
| create | absent | admin, command-line, valid | created at 0600 with one admin; one audit record |
| create | absent | non-admin | the no-admin error; nothing created |
| create | absent | admin with `streamer_username` absent or `""` | the missing-username error; nothing created |
| create | absent | admin with `streamer_username` `.x` | the invalid-username error; nothing created |
| create | absent | admin with a display name carrying U+202E | the invalid-display-name error; nothing created |
| create | absent | admin, actor `""` | the invalid-actor error; nothing created |
| create | present | any | `ErrRosterExists`; file unchanged |
| add | members, none admin (hand edit) | non-admin, command-line | the no-admin error, file unchanged |
| add | `{"schema_version":1,"members":[]}` (hand edit) | non-admin, command-line | the no-admin error, file unchanged |
| add | members, none admin | admin, command-line | written |
| add | one admin | admin, cookie, an admin actor | `errAdminGrantForbidden`, file unchanged |
| add | one admin | non-admin, cookie, the admin as actor | written |
| add | one admin | non-admin, cookie, a non-admin member as actor | the actor-not-admin error, file unchanged |
| add | one admin, then a hand edit removes the admin flag from the actor before the lock | non-admin, cookie, that actor | the actor-not-admin error, file unchanged (the check reads the disk copy) |
| add | one admin, then a hand edit removes the admin before the lock | non-admin, command-line | the no-admin error (the check reads the disk copy) |

**Acceptance criteria:**

- Every row above is a table case asserting the error by `errors.Is`
  (in-package, on the unexported errors) and the file's bytes before and
  after, or its absence. **Fixture state for the two hand-edit-race
  rows:** the `Roster` value opened while the admin existed, so only a
  disk re-read explains the refusal.
- `grep -rn 'ErrRosterAdminChannelForbidden\|ChannelHeader\|ChannelServiceToken'`
  over the module returns nothing.
- `identity.Identity{..., Channel: identity.ChannelCommandLine}.Valid()`
  is false with every other field valid; `ChannelCommandLine.Valid()` and
  `ChannelCookie.Valid()` are true; `Channel(0).Valid()` and `Channel(3)`
  are false.
- An audit record for a command-line create or add carries
  `channel=command-line` and `client_ip=local`, and an
  operating-system-user actor passes Task 2's per-channel actor rule.
- `doctor` against a hand-edited empty roster and against a hand-edited
  no-admin roster prints the roster row exactly once, failing with the
  recovery text, as any roster failure does, and
  `TestDoctorCmd_LineCountFollowsRegistry` passes unmodified.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Admit the flag on `ChannelCookie` too → the cookie refusal row writes.
2. Skip the candidate validation in `CreateRoster` → the four invalid
   create rows each create a file.
3. Let `CreateRoster` write an empty roster → the non-admin create row
   creates a file.
4. Check the admin invariant against the memory copy → the no-admin
   hand-edit race row writes.
5. Skip the actor's admin re-check under the lock → the demoted-actor
   race row writes.
6. Return the no-admin error for a non-admin actor → the actor-not-admin
   rows fail their `errors.Is` assertion.
7. Check the invariant after the write → the byte-identity rows fail.
8. Let `Identity.Valid` accept `ChannelCommandLine` → the validity row
   fails.
9. Let the seam emit the command-line channel → the seam row fails.
10. Drop the `doctor` no-admin branch → the empty-roster row passes.

**2.0 hook: extends hook 5's write half.** Rung 6's 6d runs this guard
inside its write transaction (rung 6 spec `:816-818`) and criterion 6's
"a concurrent self-demotion of the last two admins leaves one" is this
invariant, plus the actor's re-check, under the demote path 6d or rung
14 adds.

**Prior art:** none; this is the spec's own invariant. The
`channelNames` table form is the 5b-ii architecture read's.

**Independent: no.**

**Deliverable count: 4** (the channel set with the grant and its
validity refusal, the rename with the corrected comments, the validated
single-write bootstrap with the last-admin invariant and the actor
re-check, the `doctor` row).

---

## Task 7: The `dubplate-server member` verb group

Model: `sonnet`. Depends on Task 6.

Fork 1's first surface (`spec:122-124`, R37).

**Files:**
- Create: `cmd/dubplate-server/member.go`, `cmd/dubplate-server/member_test.go`
- Modify: `cmd/dubplate-server/main.go` (the `AddCommand` block)

**Interfaces:**
- Consumes: `config.Load`; `auth.OpenRoster`, `auth.CreateRoster`,
  `(*Roster).AddMember`, `(*Roster).List`, `auth.ErrRosterExists`;
  `identity.ChannelCommandLine`; `os/user.LookupId`; `os.Geteuid`.
- Produces: `newMemberCmd(environ []string, reg *config.Registry,
  stdout, stderr io.Writer) *cobra.Command`, the same constructor shape as
  its siblings, with `add` and `list` subcommands, and the constant
  `memberVerbWriteWait` (30 seconds, R-5c19).

**Outcomes:**

- **`member add <address> --name <display name> --streamer-username
  <handle> [--admin]`** loads configuration (R-5c14). When the roster is
  absent, it calls `CreateRoster` with the member as the first entry,
  which validates it like any write and refuses it unless `--admin` is
  set (R-5c8), so an invalid or non-admin first add creates nothing.
  Cobra's required-flag check proves presence only; an empty flag value
  reaches the store and is refused there. When the roster exists, it opens it and adds
  through `AddMember` on `ChannelCommandLine`. **When `CreateRoster`
  returns `ErrRosterExists`** because a concurrent bootstrap won, the
  verb opens the winner's roster and adds through `AddMember` once. It
  never overwrites a roster. `--name` and `--streamer-username` are
  required (Ruling 4). Every write runs under a context bounded by
  `memberVerbWriteWait`.
- **The actor is the operating-system user of the effective uid**:
  `user.LookupId(strconv.Itoa(os.Geteuid()))`, falling back to
  `uid:<n>` when the lookup fails, never to `$USER` (R-5c14). The task
  quotes `go doc os/user.LookupId` before relying on it.
- **The verb refuses when its effective uid differs from the roster
  directory's owner**, naming both numeric ids and the directory
  (R-5c14), before creating or writing anything, with its own unexported
  error. It quotes `go doc os.FileInfo` for `Sys` before relying on the
  `*syscall.Stat_t` assertion.
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
  existing error path. The line carries the error's own text (for
  example "duplicate address", "the roster must name at least one
  admin") and never the refused value.

**States each check meets:**

| Invocation | Roster state | Report |
| --- | --- | --- |
| `add --admin a@example.test` | absent | created at 0600 with one admin; audit record on stderr |
| `add --admin b@example.test` | exists with a | added; a still present; file not recreated |
| `add c@example.test` (no `--admin`) | absent | the store's no-admin refusal; nothing created; exit nonzero |
| `add d@example.test` (no `--admin`) | exists, no admin (hand edit) | the store's no-admin refusal; file unchanged |
| two `add --admin` at once, distinct addresses | absent | one creates, the other gets `ErrRosterExists`, reopens, and adds; the file holds both |
| `add` or `list` | exists, unparseable or invalid | the load error's line (path named, value never); exit nonzero |
| `add` | the roster's directory absent | refused naming the directory (5d's pre-create owns it); nothing created |
| `add` with a missing `--name` or `--streamer-username` | any | refused before any open |
| `add --admin a@example.test --name A --streamer-username ""` | absent | the store's missing-username refusal; nothing created |
| `add` as a uid other than the directory owner | any | refused naming both ids and the directory; nothing created |
| `add` | the euid has no passwd entry | actor `uid:<n>` in the audit record |
| `add` | roster lock held past `memberVerbWriteWait` | "roster busy", exit nonzero |
| `list` | absent | the `ErrRosterMissing` line, exit nonzero |
| `list` | written by an earlier `add` | every member, admin marked (5c-ii Task 2 proves it for a roster the API wrote) |

**Acceptance criteria:**

- Every row above is a test driving the cobra command with `SetArgs`
  against a temp-dir roster and a fixture environment that selects
  `magic-link` with synthetic secrets. **Fixture state:** the roster path
  absent for the create rows, present for the second-add row. The lock
  row shortens the deadline through a test seam on the constant's use
  (a package variable the test overrides, or a parameter the command's
  constructor takes), and says which.
- **The uid row asserts the refusal itself, and its fixture is a
  directory the test can write to but does not own**: a unique file
  under `/tmp`, which is root-owned and world-writable with the sticky
  bit, cleaned up by `t.Cleanup`. The row asserts the refusal's error
  (by `errors.Is` on the verb's unexported error), that its text names
  both ids and the directory, and that no file exists afterward. The row
  skips with a stated reason when the test runs as root or `/tmp` is
  owned by the test's uid. Do not point the row at `/`: the create fails
  with `EACCES` whether or not the guard exists, so the row could not
  fail. A positive row asserts
  a temp directory the test owns passes the guard.
- The race row runs the two invocations concurrently with `-race` and
  asserts both members present and one audit record per member.
- The no-passwd-entry row substitutes the lookup through a test seam and
  asserts the `uid:<n>` actor; it never reads `$USER`.
- The audit record is captured from the command's stderr and carries
  the actor, `command-line`, the address, `admin=true`, and
  `client_ip=local`.
- `member list` output is asserted exactly against a two-member roster.
- The `AddCommand` block lists `newMemberCmd`. The template-key
  classification test and `TestDoctorCmd_LineCountFollowsRegistry` pass
  unchanged.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Call `CreateRoster` unconditionally → the second-add row fails with
   `ErrRosterExists`.
2. Exit on `ErrRosterExists` instead of reopening → the race row holds
   one member.
3. Pass `ChannelCookie` to `AddMember` → the second-add `--admin` row
   fails with the store's grant refusal (asserted by its text, since the
   error is unexported). The create row cannot catch it, because
   `CreateRoster` takes no channel.
4. Drop the uid check → the uid row finds a roster created under `/tmp`
   and no refusal.
5. Build the actor from `user.Current` → the no-passwd-entry row reads
   the environment (report whether the test seam can make this fire; if
   not, the reviewer reads it).
6. Print the refused address in the error → the refusal-line row fails.

**2.0 hook: extends hook 5.** The verb is the write-side caller rung 6's
store keeps; 6d criterion 1 has "an imported admin can grant admin
through the CLI", which is this verb.

**Prior art:** `github.com/spf13/cobra`, already the repo's CLI library.
`os/user.LookupId` from the stdlib, chosen over `user.Current` because
the box's `CGO_ENABLED=0` build makes `Current` fall back to `$USER` and
the real uid (`src/os/user/lookup_stubs.go`).

**Independent: no.** It follows Task 6.

**Deliverable count: 4** (`add` with the create-or-open rule, the uid
guard, `list`, the wiring).

---

## Task 8: The pass close, and the 5c-i to 5c-ii handoff

Model: `sonnet`. Depends on Tasks 0 through 7, the simplifier round,
and the pass-end security read with its fold.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/research/handoffs/draft-STATUS-5c-i.md`
- Create: `docs/research/handoffs/draft-HISTORY-5c-i.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids,
  the gate evidence (`go test ./...` wall time and package-scoped times
  with `-count=1`), the resolutions R-5c1 through R-5c20 with where each
  landed, the six rulings and the eight conductor-owed folds with their
  outcomes, the security read's disposition table, the declared coverage
  gaps, both budgets, and "what a later pass would be wrong to
  rediscover". The gaps are at minimum: R-5c15's CLI audit record; a hand
  edit leaves no audit record; a lockless hand edit inside a write's
  window is overwritten; the per-address map's untracked admission past
  capacity; the constraint carried to rung 12 that a principal must not
  map a non-admin member onto a streamer admin account; the evidence
  that Navidrome v0.63.2 matches usernames by case-insensitive equality
  (`COLLATE NOCASE`, no `LIKE`); and the absent promote path.
- **The spec corrections owed**, drafted paste-ready for the conductor:
  the write-path `streamer_username` requirement (Ruling 4) against R3,
  R1's write path (`spec:291-292`), "The roster" (`spec:1106-1107`), and
  `spec:1417`'s "an optional streamer username"; `spec:1417`'s admin
  checkbox (Ruling 5); `spec:1223`'s three channels, which become two
  (fold 12: "constants land with the source that asserts them"); the 5c row's split into 5c-i and 5c-ii with both
  ceilings; R23's reservation against actuals (Ruling 6) and the web
  surface standard section 15's "4.5M ceiling is unchanged" line; R24's
  5c review shape across the split; review finding 11 against 5b-ii's
  fold 7 (R-5c10); the public-URL path rule (R-5c9) against `spec:1033`'s
  "Link hostnames derive from `DUBPLATE_PUBLIC_URL`"; review minor 6's
  refusal (R-5c18) against the 5b-ii close's routing.
- **The 5c-ii handoff**: the surface 5c-ii builds on (the roster's
  write path with the command-line grant, the actor re-check, and the
  per-refusal errors its roster API maps; `identity.ChannelCommandLine`
  and `Identity.Valid`'s refusal of it, which 5c-ii's gate requires; the
  request-link and consume-link routes and outcomes from 5b-ii; the
  `login/link` route the mailed link points at, served by the one
  handler registered at `login`), the kind answer's admin contact as
  data, and every outline item below that this pass's results changed,
  the two carried Go tasks first.
- **The 5d handoff**: the runbook row naming how an operator runs
  `member add` without putting secrets in argv (R-5c14), and the runbook
  sentence "kind mode's enumeration bound is the per-client key alone"
  (R-5c18).
- **The STATUS draft** stays present tense and under 60 lines.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is
at most 60 lines. No gate change, so `bash scripts/check.sh` green as a
formality.

**Deliverable count: 4** (the HISTORY draft, the spec corrections, the
two handoffs, the STATUS draft).

---

## The exit gate

`bash scripts/check.sh` green, plus every assertion below. The first is
the spec's 5c-row assertion (`spec:2381`) that falls in 5c-i's half. The
row's four assertions that need the roster API (the non-admin 404, the
same-origin denial, the web write's audit record, and `member list`
reading the API's write) move with it to 5c-ii. The rest are the security fold's and
this plan's.

1. `member add --admin` against an absent roster creates it in one
   validated write, and a second invocation adds to the existing file
   rather than creating a second.
2. Two independent roster values over one file, which `flock(2)` treats
   as two processes (quoted in Task 1), lose no member and resurrect no
   member removed by a hand edit between their writes, and the running
   server stops resolving a hand-removed member at its next lookup
   (finding 1).
3. A mailed link's path is owned by the prefix and never reaches the
   fallthrough (finding 2).
4. A victim's spent per-address bucket survives a same-instant flood of
   distinct addresses from distinct clients, and the flood leaves the
   per-address map unchanged (finding 3, Ruling 1).
5. One IPv6 /64 is one client (finding 4).
6. Every loose roster variant finding 7 lists fails to load.
7. No roster write can leave the roster with no admin, the bootstrap
   included; the bootstrap's first member passes the same write-path
   validation as every later member; and no channel but the command line
   can set the admin flag.
8. No refusal message or log record carries a URL password or a cookie
   value, and 5b-ii's test that no consumption log record carries the
   token (named by Task 0 item 13) still passes, re-run rather than
   rebuilt.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches.

| Surface | Owner | 5c-i's position | Contention |
| --- | --- | --- | --- |
| `internal/auth` | rung 5 | Edited across Tasks 1-6; `seam.go`'s comment in Task 6 | Rung 6 edits it only at 6d, after 5d |
| `internal/identity` | rung 5 | Two constants dropped, one added, `Identity.Valid`'s refusal of it, one refusal set | Rung 6's 6d adds `UserID`, after 5d |
| `cmd/dubplate-server/doctor.go` | rung 1's skeleton, rung 5's rows | Tasks 5, 6 | **Rung 6 lanes 6a and 6b edit it** (rung 6 spec `:1109-1118`); a concurrent lane 2 means a non-fast-forward rebase and the merge ritual's `diff-reviewer` |
| `cmd/dubplate-server/main.go` | rung 1 | Task 7's `AddCommand` line | **Rung 6 edits the registration block**; same rule |
| `cmd/dubplate-server/serve.go` | rung 5 | Task 4 (the prefix into the handler config) | Rung 6 edits it at 6d only |
| `cmd/dubplate-server/auth.go` | rung 5 | Tasks 4, 5 | Rung 6 edits the `Required` list at 6d only |
| `internal/config` | rung 1b's contract | **Untouched** (R-5c17) | Rung 6's 6c runs in lane 2 on this condition |
| `config/roster.example.json` | rung 5 | Task 2's one field | none |
| `go.mod` | shared | **Untouched** | rung 6 adds its SQLite driver |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record the planning spend's actual against the rung;
   it is outside the execution ceiling.
2. Before Task 1: rule Task 0's blockers and the eight folds; write the
   lane registry entry and the contested-surface rows; arm the runaway
   guard.
3. **Ruling 2's read-only Navidrome query is DONE (2026-09-23). Do not
   ssh to the box again for it.** It ran as `sudo sqlite3 -readonly
   /srv/music/state/navidrome/navidrome.db "SELECT user_name FROM user;"`,
   which changes no data. The estate holds one account, its name is
   email-shaped, and no case-insensitive duplicate exists; R-5c6's shape
   was widened to admit it before Task 2. The name is not recorded in any
   committed file.
4. At segment A's boundary: STATUS; re-derive the ceiling from Tasks
   1-3; confirm every fold reproduction was observed failing first.
5. At segment B's boundary: STATUS; one `code-simplifier` round over
   Tasks 1-7's accepted Go, pasting `.claude/instructions/simplifier-brief.md`
   verbatim, its review chunked one `diff-reviewer` per package; **the
   register gate on Task 4's mail copy**, one `prose-voice-reviewer`
   dispatch against the Microsoft editor standard, folded by a small
   implementer dispatch if it returns findings; then **one fresh-context
   `web-auth-security-reviewer`** over this pass's diff. Its input is
   this plan's threat-model section, the spec's threat model, both prior
   reviews' "Attacked and found sound" lists as the do-not-re-litigate
   record, **the 5b-ii findings Tasks 1 through 5 fold (majors 1-4 and
   minors 7, 8, 10, 11, 12, 13, 15, 17, 19) as a "verify fixed" list**,
   and the diff. Point it explicitly at the limiter design Ruling 1 chose
   and its corrected lockout arithmetic (R-5c3), the stat-and-read order
   (R-5c5), the lock order, the validated single-write bootstrap, and
   the actor's under-lock admin re-check. It does not receive the task
   decomposition. Fold or refuse each finding with the reason recorded;
   the record is 5c-ii's and 5d's do-not-re-litigate list. R24's read
   over the admin surface happens in 5c-ii, with the gate, the API, and
   the page.
6. The merge ritual (one lane in the main tree: gate, merge, gate;
   one `diff-reviewer` over the merged result only if a rebase was not a
   fast-forward), then **one `go-architecture-reader` per touched Go
   package** (`internal/auth`, `internal/identity`, `cmd/dubplate-server`),
   no plan in context.
7. Re-derive 5c-ii's figure from this pass's actual before 5c-ii's
   plan is drafted, and record the rung total against R23.
8. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, in
**three segments**. Every task runs `dubplate-implementer`, then
`diff-reviewer` (`claude-opus-5-5`), then the gate, with one re-dispatch
on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2, 3 | STATUS. Re-derive the ceiling. Confirm failing-first observations. |
| **B** | 4, 5, 6, 7 | STATUS. Simplifier over Tasks 1-7. The mail copy's register gate. The security read and its fold. |
| **C** | 8 | The close drafts, the merge ritual, the architecture reads, the ledger fold. |

Segment A carries three tasks and segment B four, each boundary on a
task's own green committed gate, so each is a resume point. C is the
close.

**Models.** `sonnet` by default. **Three upshifts to `opus`, all
unconditional**, dispatched as `claude-opus-5-5`: Task 1 (a
cross-process concurrency fix whose broken variants pass every
single-process test), Task 3 (the limiter's eviction and keying, where a
wrong map passes every functional test), and Task 6 (the privilege grant,
the validated bootstrap, and the invariant). Task 0 runs at `opus` in
fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Task 1 → Tasks 2, 3, 6 (the locked write function and the reload).
- Task 1 → Task 3 also by file (`handlers.go`).
- Task 2 → Task 6 (the per-refusal errors and the actor rule).
- Task 3 → Task 4 (`handlers.go`).
- Task 4 → Task 5 (`auth.go`'s refusal shapes).
- Task 6 → Task 7 (the channel, the bootstrap, and the guard).
- Tasks 0-7 → Task 8.
- **Independent by files:** Tasks 2 and 3; Task 5's nonce half against
  everything. They still run sequentially, because one tree runs one
  executor and the module gate runs over `./...`. The marks let the
  conductor reorder when a dependency stalls, never admit a second
  executor.

All six rulings are answered yes, so no task waits on one. For a resumed
conductor: Ruling 6 gated the launch; Rulings 2, 3, and 4 shape Task 2;
Ruling 1 shapes Task 3; and Ruling 5 shapes Task 6.

**Where this pass splits again, if it must:** at the Task 5 / Task 6
boundary, if segment A's re-derivation projects past the 6.5M ceiling.
Tasks 6 and 7 then move together to the head of 5c-ii, so the grant and
the bootstrap never ship without the verb that calls them (R-5c20). The
projection sits 1.78M under the ceiling, so this is a stated fallback,
not an expectation.

**Budget:** 6.5M execution, flag 5.2M, planning recorded separately,
STATUS at each boundary, spend scored at close against the ceiling and
against zero attended interactions. The eight folds and twenty
resolutions are ruled at dispatch together and do not count as an
execution sitting; the six rulings are the plan-approval gate's, not an
execution sitting.

---

## Pass 5c-ii: the `web/` tree, the pages, and the gate leg (scoped outline)

**This is an outline, not a plan, except for its head.** 5c-ii's plan is
authored just in time after 5c-i merges, against the tree 5c-i leaves,
with its figure re-derived from 5c-i's actual. Its first two tasks are
specified in full (R-5c1), so 5c-ii's plan adopts them rather than re-deriving them. The
rest fixes scope, order, and the open questions its plan must close.

### The head of 5c-ii: the admin gate, the roster API, and their wiring

These two tasks run first in 5c-ii, before any `web/` task, because the
add-member page is the roster API's second client and the document
route mounts the gate. Resolutions R-5c12, R-5c13, and R-5c19's
`rosterAPIWriteWait` govern them.

### 5c-ii Task 1: The admin gate and the roster API

Model: `opus`. Depends on 5c-i's merge (its Tasks 2 and 6). Ruling 5
shapes it.

Fork 1's second surface, backend half. The page is Task 10's; this task
builds the handler it calls and the gate its document route will reuse.

**Files:**
- Create: `internal/auth/admin.go`, `internal/auth/admin_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`
  (lift `readForm` from a method on `*handlers`, which reads
  `h.cfg.BodyCap`, into a package function taking the cap, so the 5b-ii
  handlers and the roster API share one body path rather than a
  duplicate the simplifier would undo)

**Interfaces:**
- Consumes: `Seam.VerifyIdentity`; `SessionCookie`; `(*Roster).List`,
  `(*Roster).AddMember`; 5c-i Task 2's and Task 6's unexported errors;
  `identity.ChannelCookie`.
- Produces: `auth.RequireAdmin(cfg AdminConfig, next http.Handler)
  http.Handler`, the gate, whose callers are this task's API and the
  add-member document route (5c-ii's served-header task), which mounts it
  from `cmd/dubplate-server`, so it is born exported with its
  cross-package caller in the same pass (R-5c20). `auth.NewRosterAPI(cfg AdminConfig) http.Handler`,
  which returns **the gated composition only**: `RequireAdmin` wrapping
  an unexported handler that answers GET (list) and POST (add) and 405
  otherwise. No ungated roster handler is exported, so no wiring mistake
  can mount one. `auth.AdminConfig{Seam Seam;
  Roster *Roster; ClientIP func(*http.Request) string; Now func()
  time.Time; Logger *slog.Logger; BodyCap int64}`. The verified identity
  reaches the inner handler through a request context value with an
  unexported key and an unexported accessor (R-5c20). The constant
  `rosterAPIWriteWait` (5 seconds, R-5c19).

**Outcomes:**

- **The gate** (R-5c12): verify through the seam with `sameOrigin` true
  only for `Sec-Fetch-Site: same-origin`; require `Identity.Valid()` and
  `Admin`; answer every refusal with `http.NotFound`, the shell
  terminator's own answer, including a verified non-admin and a roster
  store failure; set no cookie on any refusal; log the refusal reason and
  the client IP, never the cookie value and never an address. **Only
  after the admin check passes**, set the refreshed cookie when the seam
  returned one, then call the inner handler with the verified identity in
  the context.
- **The inner handler refuses on its own** with the same
  `http.NotFound` when the context carries no valid admin identity, so
  it fails closed even if a later edit bypassed the gate.
- **GET** answers `no-store` JSON: a `members` array, each entry with
  `address`, `display_name`, `streamer_username`, and `admin`. A roster
  store failure answers 500 with an outcome, never a partial list.
- **POST** takes a form-encoded body (R-5c13): cap at `BodyCap` before
  parsing (`http.MaxBytesReader`, then `ParseForm`, as the 5b-ii handlers'
  `readForm` does, after the lift above); read `address`, `display_name`, and
  `streamer_username` from `r.PostForm` only; refuse when an `admin` key
  appears in `r.Form`, body or query string, with any value, empty
  included; then call `AddMember` under a context bounded by
  `rosterAPIWriteWait`, with the gate's actor, `ChannelCookie`, and
  `ClientIP`. It answers one outcome per result, in the same `{"outcome":
  ...}` JSON shape as the 5b-ii handlers, always `no-store`.

**Outcomes table (the POST):**

| Result | Status | `outcome` |
| --- | --- | --- |
| added | 201 | `added` |
| duplicate address | 409 | `duplicate_address` |
| duplicate streamer username | 409 | `duplicate_streamer_username` |
| invalid member address | 400 | `invalid_address` |
| invalid display name | 400 | `invalid_display_name` |
| invalid streamer username | 400 | `invalid_streamer_username` |
| missing streamer username | 400 | `missing_streamer_username` |
| an `admin` key in the body or query | 400 | `admin_not_grantable` |
| body over the cap | 413 | `too_large` |
| body unparseable | 400 | `malformed` |
| the lock wait outlived `rosterAPIWriteWait` | 503 | `busy` |
| any other error, the actor-not-admin refusal included | 500 | `failed` |

The outcomes are specific because the caller is already trusted
(`spec:1421-1424`); the copy a person reads is 5c-ii's string deck.

**Acceptance criteria:**

- **The gate's refusal table** is a test: no cookie, a forged cookie, an
  expired session, a session older than its epoch, a member removed from
  the roster after signing in, a roster store failure, and a verified
  non-admin. Every row answers the same status, the same headers (all
  but `Date`), and the same body as `http.NotFound` does, and carries no
  `Set-Cookie`. **Fixture state:** a roster holding both an admin and a
  non-admin, and valid sessions minted for both, so the non-admin row's
  refusal comes from the flag and not a failed verification; the
  non-admin's request is same-origin and old enough that the seam
  returns a refreshed value, so a gate that set the cookie before the
  admin check would fail the no-`Set-Cookie` assertion.
- A captured log for the forged-cookie row does not contain the cookie
  value, and no refusal log line contains an address.
- A same-origin GET with a valid admin session carries a `Set-Cookie`
  whose expiry moved forward and whose anchor did not; a
  `Sec-Fetch-Site: same-site` GET and a header-less GET carry none.
  **Fixture state:** a session old enough that refresh visibly moves its
  expiry.
- Every POST outcome row above is a test. Each refused row asserts the
  file's bytes unchanged and no audit record. The `added` row asserts
  exactly one audit record naming the admin's address as actor and
  `cookie` as channel.
- The `admin` row holds for `admin=on`, `admin=false`, and `admin=` in
  the body, and for `?admin=` in the query string.
- The `busy` row holds the directory lock from the test and asserts 503
  within the shortened deadline (the constant's use takes a test seam,
  and the task says which).
- `bash scripts/check.sh` green.

**Named mutations:**

1. Answer a JSON 404 or a 403 for a verified non-admin → the
   `http.NotFound` equality row fails.
2. Set the refreshed cookie before the admin check → the non-admin row
   carries `Set-Cookie`.
3. Read the admin flag from a request field or header → the non-admin
   row reaches the inner handler.
4. Refresh on any request → the header-less row carries a cookie.
5. Check `admin` in `r.PostForm` only → the query-string row writes or
   reaches the store.
6. Take the actor from the form → a test posting a spoofed actor field
   finds it in the audit record.
7. Parse before capping → the over-cap row reads the whole body (assert
   through a reader that counts bytes read).
8. Return the inner handler ungated from `NewRosterAPI` → a non-admin
   GET answers the list.
9. Skip `Identity.Valid` in the gate → a test identity on
   `ChannelCommandLine` with `Admin` true, injected through a stub `Seam`,
   reaches the inner handler. This is the one production call that makes
   5c-i's `Identity.Valid` refusal protect anything.

**2.0 hook: neutral to the schema, and it makes 2.0's web admin surface
cheap**: `RequireAdmin` is the gate every later admin page reuses.

**Prior art:** none external. The gate reuses the seam 5b-ii built and
the outcome shape and body handling its handlers use.

**Independent: no.** It opens 5c-ii.

**Deliverable count: 4** (the gate with its refresh after the admin
check, the GET, the POST with its outcome table and deadline, the
admin-key refusal).

### 5c-ii Task 2: The serve wiring, and the exit gate end to end

Model: `sonnet`. Depends on 5c-i's merge (its Tasks 4 and 7) and 5c-ii
Task 1.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/serve_test.go`

**Outcomes:**

- **`newServeShell` builds the seam it already constructs into the
  admin config** and mounts `NewRosterAPI(cfg)`, which is already gated,
  under the owned prefix as the subtree `roster`, beside the three auth
  mounts. The comment that says "No handler here consumes the seam" is
  removed, since one now does. The body cap is the existing
  `DUBPLATE_AUTH_REQUEST_BODY_CAP` threshold.
- **The path is `<prefix>/roster`**, and 5c-ii's add-member page lives at
  `<prefix>/members`, so the document and the API never share a route.
- **The exit gate's cross-surface rows are proved through the real
  shell** (below), because a handler test cannot see the origin check or
  the prefix router.

**Acceptance criteria (every row through `shell.Handler()`):**

- A POST to `<prefix>/roster` with a valid admin cookie and neither
  `Origin` nor `Sec-Fetch-Site` is refused by the origin check and the
  file is unchanged.
- The same POST with `Sec-Fetch-Site: same-origin` adds the member and
  writes one audit record.
- **A non-admin's GET and POST to `<prefix>/roster` are
  indistinguishable from a GET and POST to an unregistered
  `<prefix>/<unknown>`**: same status, same body, every header equal but
  `Date`, and no `Set-Cookie` (risk 14, `spec:1418-1420`). The POSTs carry
  `Sec-Fetch-Site: same-origin`, so the origin check admits both and the
  comparison reaches the gate and the terminator. With an admin session,
  the GET answers the list.
- **`member list` reads what the API wrote**: the test posts through the
  shell, then drives `newMemberCmd` with `list` against the same roster
  path and asserts the new member appears. **Fixture state:** the verb
  opened after the API's write, and the shell's roster opened before the
  verb's first write, so the second half proves the reload from the other
  side too: a `member add` through the verb is visible to a subsequent
  GET through the shell without a restart.
- A link minted by the real request-link handler, followed through the
  shell, reaches the owned `login` stub and not the fallthrough (5c-i
  Task 4's test, re-run through the wired verb's own construction if
  Task 4 built it against a hand-assembled shell).
- `bash scripts/check.sh` green.

**Named mutations:**

1. Mount the API outside the prefix → the admin GET reaches the
   fallthrough and answers 404 instead of the list.
2. Have `newServeShell` answer the gate's refusal with anything but the
   terminator's `http.NotFound` (for example a wrapping handler that adds
   a header) → the indistinguishability row fails.
3. Construct a second roster value for the API → the verb-then-GET row
   still passes (both reload), so report that this mutation does not fire
   and why; the diff-reviewer confirms one roster value is shared.

**2.0 hook: neutral.**

**Prior art:** none.

**Independent: no.**

**Deliverable count: 3** (the mount, the path decision, the end-to-end
exit-gate rows).

**The spec 5c row's exit-gate assertions that move with them**
(`spec:2381`), each proved by 5c-ii Task 2 through the real shell:

1. A non-admin session receives 404 from the add-member route (the
   roster API, and later the add-member document) while an admin session
   receives the roster.
2. An add-member POST without a same-origin signal is denied.
3. An add-member POST that succeeds writes one audit record naming the
   actor and the channel, and one that fails validation writes none and
   leaves the file unchanged.
4. `member list` reads what the roster API wrote.

**The review focus line that moves with them:** a signed-in member
without the admin flag probing the roster API with GET, POST, and a
crafted `admin` field. Expected: the terminator's 404 for all, and no
byte on disk changes.

### The rest of 5c-ii

**Scope, from `spec:2381` less 5c-i's half.** `web/` born as a SvelteKit
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
W6.11, W7.14, W7.15, W7.18, W8.1-W8.9, W9.1, and W1.2 and W1.4's
register, W3.6, and W4.4 and W4.6's stacking). No grammar layer, no
component set, no styleguide: rung 7's.

**Routes, fixed now.** The login page at `<prefix>/login`, the
interstitial at `<prefix>/login/link` (R-5c9), and the add-member page at
`<prefix>/members`; the roster API sits at `<prefix>/roster` (5c-ii Task 2).
`Shell.Register` refuses a subtree name carrying `/`
(`internal/serve/serve.go`, `subtreeNameRefusal`), so one handler
registered at `login` serves both the login page and `login/link`, and
SvelteKit's `_app` directory needs its own subtree. The registration of
`login` is proved through 5c-i's prefixed-link test.

**Proposed task order: the two head tasks above, ten `web/` tasks, and
the close, thirteen tasks plus Task 0.** Thirteen passes
this build's twelve-task line, so 5c-ii's plan should propose its own
split, most naturally after the served-header task (the Go-side and
delivery work on one side, the three pages, the gate leg, and the
render-and-read on the other). The `web/` tasks, numbered from 3:

3. The scaffold, the composed theme data, and the theme binding.
4. The fonts, their licenses, and the manifest.
5. The embed path, the placeholder, and the R4.5 rewrite per
   prerendered document, with R14's theme keys and their validation.
6. The branding stylesheet and the operator CSS file.
7. The served header set (CSP on documents; `nosniff`,
   `Referrer-Policy: no-referrer`, and HSTS on **every owned response**,
   the roster API's and the 5b-ii JSON handlers' included, not only
   documents); the cache rules (`no-cache` documents, `no-store` on the
   consumption route's document, immutable hashed assets); and the
   add-member document route behind 5c-ii Task 1's `RequireAdmin`, whose refusal
   must track the served 404 document once W3.7's 404 lands rather than
   `http.NotFound`'s plain text.
8. The login page and the string deck, with the deck's register gate.
   The deck includes the sign-out copy, bounded by the 5b-ii handoff: a
   stateless session survives sign-out, so the copy never claims
   revocation.
9. The consumption interstitial at `login/link` under the prefix
   (R-5c9): reads the token from `window.location`, clears it from the
   address bar, and POSTs it in a form field.
10. The add-member page: the roster list with its empty state and admin
   marker, the form of address, display name, and streamer username with
   no admin control (R-5c13), and specific copy per 5c-i outcome value.
11. The gate leg: G1-G10, the server-file check, Node required, the
   rejoin sentences.
12. Render and read all three pages at W4.1's viewports before any
    attended check (`see-it-before-geoff-tests-it`).

**Open questions 5c-ii's plan must close with quoted evidence:**

- **The runtime prefix against SvelteKit's build-time base.**
  `DUBPLATE_SERVE_PREFIX` is a runtime threshold, and SvelteKit's
  `kit.paths.base` is a build-time value ("A root-relative path that must
  start, but not end with `/`", SvelteKit configuration docs, read
  2026-09-23). `kit.paths.relative` defaults to `true` ("`base` and
  `assets` imported from `$app/paths` will be replaced with relative asset
  paths during server-side rendering, resulting in more portable HTML").
  The kit source answers half the question already
  (`@sveltejs/kit/src/runtime/client/client.js`): on hydration
  the client resolves the route from `location.pathname` with the
  build-time `base` sliced off, so under a prefix the build never saw the
  route resolves to nothing. The same docs page says single-page-app
  fallback pages always use absolute paths. **The first candidate to
  prove is SvelteKit's universal `reroute` hook**, which the router
  docs say "allows you to modify the URL before it is used to determine
  which route to render" and which lives in `src/hooks.js`, not a server
  file, so the gate's server-file check allows it, paired with the Go
  rewrite stamping the runtime prefix into each document for the hook to
  read. `router.type: 'hash'` is documented too but disables
  prerendering. The plan must prove its choice by building and serving
  under a non-default prefix. If nothing works, it proposes a conductor
  ruling (for example, a non-default prefix refused until rung 7).
  **This is the pass's largest unknown.**
- **Where the CSP travels.** The same docs state "When pages are
  prerendered, the CSP header is added via a `<meta http-equiv>` tag (note
  that in this case, `frame-ancestors`, `report-uri` and `sandbox`
  directives will be ignored)". `spec:1348` requires `frame-ancestors
  'none'`, so the Go layer must serve the CSP as a header, with the hashes
  read from the build or computed from each embedded document at startup
  (`spec:1336-1339` leaves the choice open).
- **The `style-src 'unsafe-inline'` owner note** (`spec:1346-1352`)
  rides to that pass, per `spec:2696-2700`.
- **HSTS's `max-age` and its loopback carve-out** (R38 admits http on
  loopback, where HSTS must not be sent or is ignored), quoted from RFC
  6797.
- **Node and package pins** under the `dependency-upgrade` skill: the
  newest production SvelteKit, Svelte, Vite, DaisyUI, and Tailwind at
  plan time.

**Reviewers inside 5c-ii:** `svelte-reviewer` and
`daisyui-a11y-reviewer` with the web surface standard as their lens
(section 15), and **R24's full `web-auth-security-reviewer` read** over
the admin gate, the roster API, their wiring, the header set, the CSP,
the embed path, and the add-member document route: the
cookie-authenticated, privilege-granting write R24 exists for, reviewed
with the page that uses it. Its inputs add 5c-i's security-read record
as do-not-re-litigate. The AI-tells rule
runs inside; the string deck's register gate is separate (`spec:1432-1436`).

**2.0 hooks:** neutral except hook 5 through DL R4.2's per-member theme
row, which W5.10's signed-out cookie does not change.

**Provisional figure: about 10.3M** (about 9.25M execution and 1.0M
planning; "Token ceiling" above, per Ruling 6), re-derived from
5c-i's actual, plus about 1.0M more planning if its plan splits.
