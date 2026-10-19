# Draft HISTORY entry: rung 5, pass 5c-i (the security fold, the CLI admin grant, the `member` verb)

**Draft only. The conductor folds this into `docs/HISTORY.md` at the merge, path-limited.**

---

## Rung 5, pass 5c-i (the security fold, the CLI admin grant, the `member` verb): closed 2026-09-24, lane 1

Executed against `docs/plans/2026-09-23-rung-5c-pages-member-verb.md`
(revision 2.2, ratified by Geoff 2026-09-23 with all six rulings yes) in
the main tree, from `cb9ecc3` (5b-ii's merge). Task 0 (LAUNCH) is
`docs/research/2026-09-23-rung-5c-i-task-0-report.md`,
committed `978ef43`. The plan's full review history, fold rulings, and
finding dispositions live in
`docs/research/2026-09-23-rung-5c-plan-fold-record.md`.

Tasks 1 through 7 ran across two segments: A (Tasks 1-3, closed
`a8be067`) and B (Tasks 4-7, closed `f92ac3b`), followed by one
`code-simplifier` round (`e90f854`), the mail-copy register gate
(`d814726`, then `657f9ec` fix), the pass-end security read
(`ce8c0be`), its fold across two tasks (`95313d6`, then `963abfd` and
`3228a21` fix), and this close (Task 8).

### The gate evidence

`go test -count=1 ./...` at HEAD (`3228a21`): **2m10.739s** wall
(`internal/harness/conform` at 128.468s and `internal/harness` at
10.392s dominate; every other package under 6.8s). Package-scoped,
`-count=1`: `internal/auth` **6.073s**, `internal/identity` **0.008s**,
`internal/notify` **0.041s**, `internal/serve` **1.149s**,
`cmd/dubplate-server` **0.314s**.

`CAIRN_GATE_LANE=light cairn-run-gate 'bash scripts/check.sh'` at this
close's own commit: **exit 0**, 98-line log. The gitleaks leg scanned
about 121.96 MB in 59.1s and found no leaks. `go.mod`, `.gitleaksignore`,
and `env/` are byte-identical across `cb9ecc3..HEAD` (`git diff` empty
for all three; this pass added no dependency).

Two staged legs stayed at warning, unchanged in kind from 5b-ii: 325
process citations in Go comments (`CHECK_CITATIONS=warn`) and 128
exported-unused `internal/` identifiers (`CHECK_EXPORTED_UNUSED=warn`).

### What landed

Task 1 (`36f528d`) added cross-process roster sharing: a directory
`flock(2)` taken before the in-process mutex, one descriptor per
`Roster` value so two values in one process exclude each other as two
processes do, and a strict `encoding/json/v2` loader
(`RejectUnknownMembers`, `schema_version==1`, non-nil `members`). This
lands R-5c4 and R-5c5's design (fold: "convergent A/C" in the plan fold
record — the security review and the plan's own risk lens converged on
the same defect independently).

Task 2 (`428665d`, then `81cd77b` and `49363bf` fix rounds) tightened
roster validation: `streamer_username` becomes required on the write
path (Ruling 4) and restricted to `[A-Za-z0-9][A-Za-z0-9._@+-]*`
(Ruling 2, the conductor's read-only Navidrome query), display names
admit the zero-width joiner and non-joiner (Ruling 3) but the address
set still refuses them, and every refusal gets its own unexported
sentinel (R-5c20). The two fix rounds corrected fixture runes flagged
by `ST1018` to escape-concatenation form rather than literal invisible
bytes in the source.

Task 3 (`74919a6`) charges the per-client rate-limit key before the
per-address key, and the per-address key only for an address that
resolves against the roster (Ruling 1, R-5c3): a flood of addresses
that are not on the roster never touches the per-address map, so a
victim's spent bucket is never evicted by it, and
`evictRefilledOrAdmit` never resets an unrefilled bucket.

Task 4 (`16fd4c1`, then `af6c06f` and `55d0305` fix rounds) routes a
mailed link through the owned prefix (`JoinPath(prefix, linkPath)`, a
public URL carrying no path, R-5c9) and bounds session lifetime, nonce
retention, and the SMTP port at load. `af6c06f` dropped ruling and
finding citations from the shipped Go comments (the process-citation
rule the gate stages at warning today). `55d0305` added a test proving
`Prefix` and `SessionAbsoluteCap` wiring at `newServeShell`'s own call
site.

Task 5 (`f4fe4a3`, then `e8bd5f5` fix round, then `4c80b5d` as a
conductor-ruled test-only follow-up "5f") names each failing identity
source key on `doctor`'s own line through a keyed error type
(`errors.AsType`, R-5c10, fold 7) and sweeps an unreadable nonce record
by mtime rather than never. The fix round keyed every magic-link bound
refusal so `doctor` never reads an unsatisfied source as active. The
second fix round was a vacuous proof of named mutation 3 (the mutation
did not fire against the test as written); rather than park the lane,
the conductor ruled a narrow test-only fix at `4c80b5d`, which fires
the message-prefix-parse test in its natural form.

Task 6 (`730d3ce`) granted the admin flag from the command line alone:
a new `identity.ChannelCommandLine` that `AddMember` admits `Admin` on
and `Identity.Valid` refuses (R-5c7), `ChannelHeader` and
`ChannelServiceToken` dropped as unused (conductor-owed fold 12,
R-5c7), and the last-admin guard as a write-path invariant with no
exemption: `CreateRoster` writes the first admin in the same locked,
validated write, and no demote or remove method is exported (R-5c8).

Task 7 (`4d54a85`, then `537482e` fix) added the `dubplate-server
member` command group, `add` and `list`. The verb resolves the roster
through `config.Load` and refuses a foreign directory owner (R-5c14),
the actor comes from `user.LookupId(Geteuid)` with a `uid:<n>`
fallback, never `$USER` (R-5c14), and the audit record goes to the
verb's own stderr (R-5c15). The fix round covered `list`'s load-error
path and proved `list` reads what `add` wrote.

The simplifier round (`e90f854`, Opus 5.5) collapsed `Limiter`'s three
near-identical charge-and-refuse methods onto one
`keyedLimiter.chargeOrRefuse` helper and `MintSession`/`RefreshSession`'s
duplicated expiry clamp onto one `Session.clampExpiry` method, and
reworded `authEnviron`'s stale doc comment (`doctor` now also fails an
empty roster for naming no admin, not only a missing one). Its review
returned `fix` on the comment only, folded in Task S2.

The mail-copy register gate (`d814726`, then `657f9ec` fix) floors
`DUBPLATE_MAGIC_LINK_EXPIRY` at one minute at load (keyed through the
existing `sourceKeyError` shape), rounds any partial minute up in the
mailed copy so it never understates validity, and singularizes only at
exactly one minute ("works only once" replaces "works once"). The fix
round dropped the submitted value from the floor refusal's text (it
had surfaced in `doctor`'s per-key row) and added a test proving the
refusal's error string never contains the submitted duration.

### The pass-end security review

`docs/research/2026-09-24-rung-5c-i-security-review.md`
(`ce8c0be`): scope `git diff cb9ecc3..HEAD -- internal cmd config` at
`657f9ec`. **No critical, no major.** Majors 1-4 from 5b-ii's own review
are fixed and independently reproduced, the strongest check six real OS
processes adding 25 members each to one roster (151 of 151 landed, no
temp file left). Minors 7, 8, 11, 12, 13, 17, and 19 are fixed; minor
15 is fixed as specified but leaves gap m1; minor 10 is only partly
fixed (gap m2); the R-5c18 refusal of minor 6 is sound, with one
sentence added to the 5d runbook row.

| 5b-ii finding | Status |
| --- | --- |
| Major 1 (cross-process roster writes) | Fixed |
| Major 2 (mailed link outside the owned prefix) | Fixed |
| Major 3 (a refused client evicts a victim's per-address bucket) | Fixed (R-5c3 design) |
| Major 4 (IPv6 per-client key at /128) | Fixed |
| Minor 6 (kind-mode post-ceiling oracle) | Refusal sound (R-5c18); mail-ceiling lockout-lever sentence added to the 5d runbook |
| Minor 7 (loose roster loader) | Fixed |
| Minor 8 (non-ASCII control/format characters pass) | Fixed (Cf/Z*/Cc); residual n5 |
| Minor 10 (public URL password echoed) | Partial pre-fold (m2) |
| Minor 11 (`doctor` names no failing key) | Fixed |
| Minor 12 (`doctor` green on a bad port or lifetime) | Fixed |
| Minor 13 (empty actor / invalid channel) | Fixed |
| Minor 15 (`streamer_username` case and shape) | Fixed as specified; residual m1 |
| Minor 17 (no charge for an unnormalizable address) | Fixed; residual n1 |
| Minor 19 (unreadable nonce record never swept) | Fixed; residual m5 |

**Five new minors and five nits**, folded across two tasks:

- **Task S1** (`95313d6`, Opus 5.5): keyed roster uniqueness on the
  effective principal (`streamer_username`, else the normalized
  address) with `strings.EqualFold`, at load and on write (m1: a
  username could case-fold to another member's address, giving both
  members one Navidrome principal at rung 12); refused a roster file a
  group or other user may write, read from the descriptor already
  `fstat`'d (m3); logged the audit record at warning level when a write
  lands but the post-write re-read or `readRoster` fails, so a landed
  change is never silently unaudited (m4); refused a blank display name
  on the write path (n3); logged a cookie actor in its normalized form
  (n4).
- **Task S2** (`963abfd`, then `3228a21` fix): dropped `url.Parse`'s
  inner error from the public-URL refusal and checked the raw value for
  `@` before parsing, so a userinfo password containing `/`, `?`, or `#`
  can no longer leak a prefix into the config error (m2); swept an
  unreadable or unparseable nonce record at `mtime + max(retention, a
  link-expiry ceiling)` rather than `mtime + retention` alone, closing a
  replay window opened by lowering retention and link expiry together
  (m5); charged the per-client limit before reading the request body on
  `requestLink`, matching `consumeLink`'s order, so an oversized or
  unparseable body is charged like every other request (n1); refused an
  empty or `/` prefix in `LinkURL` (n2); corrected two stale test
  comments. The fix round renamed a test asserting the retention-vs-expiry
  relation rather than a floor it does not set, and corrected a comment
  naming the wrong failure mode for a charge-before-parse test.

**Refused: n5** (invisible code points outside Cf and Z* — Hangul
fillers, CGJ, VS16, braille blank — still pass the address normalizer).
Only a trusted admin can add such an address today, and refusing it
needs the derived Unicode Default_Ignorable_Code_Point table, which the
standard `unicode` package does not expose. Noted for 5c-ii's web write
path.

### The eight exit-gate assertions

Each names the test(s) proving it at HEAD (`3228a21`), run with
`go test -count=1 -run '<names>' ./...` and confirmed green.

| # | Assertion | Proving test(s) |
| --- | --- | --- |
| 1 | `member add --admin` against an absent roster creates it in one validated write; a second invocation adds to the existing file | `TestMemberAddCmd_CreatesRosterWithFirstAdmin`, `TestMemberAddCmd_AddsToExistingRosterWithAdmin` (`cmd/dubplate-server`) |
| 2 | Two independent roster values over one file lose no member, resurrect no hand-removed member, and the server stops resolving a hand-removed member at its next lookup | `TestRoster_CrossProcessWrites`, `TestRoster_ReloadsChangedFile`, `TestRoster_ModTimeTrustWindow`, `TestRoster_ConcurrentWritersAcrossValues`, `TestAddMember_BusyWhileLockHeld` (`internal/auth`) |
| 3 | A mailed link's path is owned by the prefix and never reaches the fallthrough | `TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree` (`cmd/dubplate-server`) |
| 4 | A victim's spent per-address bucket survives a same-instant flood of distinct addresses; the flood leaves the per-address map unchanged | `TestRequestLink_UnknownFloodLeavesSpentAddress`, `TestRequestLink_ChargingOrder`, `TestKeyedLimiterRefilledPolicy` (`internal/auth`) |
| 5 | One IPv6 /64 is one client | `TestClientKey`, `TestClientLimitsKeyIPv6OnItsSlash64` (`internal/auth`) |
| 6 | Every loose roster variant finding 7 lists fails to load | `TestOpenRoster_StrictLoader` (`internal/auth`) |
| 7 | No roster write can leave the roster with no admin, bootstrap included; the bootstrap's first member passes the same write-path validation as every later member; no channel but the command line can set the admin flag | `TestRosterWrites_AdminInvariantAndGrant`, `TestAddMember_ActorAndChannelRefusals` (`internal/auth`) |
| 8 | No refusal message or log record carries a URL password or a cookie value; 5b-ii's no-credential-in-consumption-log test still passes, re-run rather than rebuilt | `TestLogsCarryNoCredential` (`internal/auth`), `TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue` (`cmd/dubplate-server`) |

All eight passed at HEAD. `bash scripts/check.sh` also ran green
through `cairn-run-gate` (see "The gate evidence" above).

### The twenty resolutions (R-5c1 through R-5c20)

Each is stated in full in the plan
(`docs/plans/2026-09-23-rung-5c-pages-member-verb.md`) at
its own `### R-5cN` heading; this table names where each landed at
HEAD. R-5c12, R-5c13, R-5c16, and R-5c19's admin-gate/roster-API half
apply in 5c-ii and are recorded there rather than as gaps here.

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-5c1 | The split, 5c-i / 5c-ii, cut at Task 7 / Task 8 | This pass's own scope |
| R-5c2 | The security fold is Tasks 1 through 5, a block, not one task | Tasks 1-5 |
| R-5c3 | The per-address limit charges only roster addresses (Ruling 1) | `internal/auth/handlers.go`, `limit.go` (Task 3) |
| R-5c4 | The roster decodes through `encoding/json/v2`, strictly | `internal/auth/roster.go` (Task 1) |
| R-5c5 | The store reloads on a changed file, reads what it stats, fails closed on a broken one | `internal/auth/roster.go` (Task 1) |
| R-5c6 | `streamer_username` required on the write path, restricted shape (Rulings 2, 4) | `internal/auth/roster.go` (Task 2) |
| R-5c7 | CLI admin grant is `identity.ChannelCommandLine`; no verifier can emit it | `internal/identity`, `internal/auth/roster.go` (Task 6) |
| R-5c8 | Last-admin guard is a write-path invariant with no exemption | `internal/auth/roster.go` (Tasks 1, 6, 7) |
| R-5c9 | Public URL is a bare origin; a link is origin + prefix + `/login/link` | `internal/auth/token.go`, `cmd/dubplate-server/auth.go` (Task 4) |
| R-5c10 | `doctor` names each failing source key on the source's own line | `cmd/dubplate-server/doctor.go` (Task 5) |
| R-5c11 | Session lifetime above the cap refuses at load and clamps at mint | `cmd/dubplate-server/auth.go`, `internal/auth/session.go` (Task 4, simplifier) |
| R-5c12 | The admin gate answers 404 to a non-admin | 5c-ii (not this pass's Files) |
| R-5c13 | The roster API takes a form, reads no admin field, refuses one (Ruling 5) | 5c-ii, save Task 6's rule that only the command line can grant admin |
| R-5c14 | The `member` verb resolves the roster through `config.Load`, refuses a foreign owner | `cmd/dubplate-server/member.go` (Task 7) |
| R-5c15 | The verb's audit record goes to its own stderr | `cmd/dubplate-server/member.go` (Task 7) |
| R-5c16 | The SvelteKit server-file check moves to 5c-ii | Gate unchanged this pass; deferred |
| R-5c17 | 5c-i leaves `internal/config` untouched | Confirmed: no edit to `internal/config` across `cb9ecc3..HEAD` |
| R-5c18 | Review minor 6's code fix is refused; its runbook sentence goes to 5d | Security review disposition table, above; 5d runbook carry |
| R-5c19 | Lock waits carry named deadlines; readers never wait behind a writer's lock wait | `internal/auth/roster.go` (`rosterAPIWriteWait` 5s belongs to 5c-ii; `memberVerbWriteWait` 30s, Tasks 1, 7) |
| R-5c20 | Nothing is exported ahead of a cross-package caller | Applied throughout: `CreateRoster`, `ErrRosterExists`, `identity.ChannelCommandLine` each have their caller in `cmd/dubplate-server/member.go`; confirmed by the security review's own "exported surface" check |

### The six rulings and the eight conductor-owed folds

All six rulings (the table at the plan's head) were ruled yes by Geoff
on 2026-09-23 and are reflected in the resolutions above (Ruling 1 to
R-5c3, Ruling 2 to R-5c6, Ruling 3 to Task 2, Ruling 4 to R-5c6, Ruling
5 to R-5c13/Task 6, Ruling 6 to the token ceiling).

The eight conductor-owed folds (numbered 1, 2, 5, 6, 7, 10, 11, 12 in
the plan; folds 3, 4, 8, 9 belong to other passes' plans and are not
this plan's own) all applied as the plan states, confirmed by
execution:

1. **Fold 1, the split** — cut at the Task 7 / Task 8 boundary. Confirmed: 5c-i is Task 0, Tasks 1-7, and this close.
2. **Fold 2, the security fold as a block** — Tasks 1-5, all before Task 6 or Task 7. Confirmed by commit order.
5. **Fold 5, the CLI admin grant's channel** — `identity.ChannelCommandLine`. Confirmed at `730d3ce`.
6. **Fold 6, the last-admin guard's shape** — a write-path invariant, no exemption. Confirmed by `TestRosterWrites_AdminInvariantAndGrant`.
7. **Fold 7, `doctor`'s failing-key report** — a keyed error type, no new rows. Confirmed at `f4fe4a3`/`e8bd5f5`.
10. **Fold 10, the ceiling** — 6.5M execution ceiling, 5.2M flag. Confirmed under budget (see below).
11. **Fold 11, review minor 6** — code fix refused, runbook sentence carried to 5d. Confirmed in the security review disposition.
12. **Fold 12, the unused channels** — `ChannelHeader` and `ChannelServiceToken` dropped. Confirmed at `730d3ce`.

### The declared coverage gaps

**Seven**, all carried forward rather than closed by this pass (m1
through m5's roster-side residuals are fixed, not carried; these are
distinct):

1. **R-5c15's CLI audit record is stderr only.** No persistent audit
   log exists; a service restart loses the terminal history.
2. **A hand edit to the roster file leaves no audit record.** Only
   `AddMember`/`CreateRoster` log; a direct file edit bypasses both.
3. **A lockless hand edit inside a write's window is overwritten.**
   The writer re-reads and merges under the lock, but a hand edit that
   lands between the writer's read and its rename is lost (R9 residual,
   recorded in the plan's threat model).
4. **The per-address map's untracked admission past capacity.**
   `evictRefilledOrAdmit`'s no-refilled-bucket case admits the new key
   untracked rather than refusing it; unreachable below 10,000 roster
   members, per the security review.
5. **The rung 12 constraint that a principal must not map a non-admin
   member onto a streamer admin account** is not enforced anywhere in
   this pass (the roster has no streamer-mapping concept yet); it
   travels as a stated precondition to rung 12.
6. **Navidrome v0.63.2 matches usernames by case-insensitive equality**
   (`COLLATE NOCASE`, no `LIKE`), confirmed by the conductor's read-only
   query (Ruling 2) and cited as the reason R-5c6's shape widened; this
   is evidence carried forward, not a gap this pass closes.
7. **No promote path exists.** A non-admin member can be added, but no
   verb or API in this pass or 5c-ii's outline promotes one to admin
   short of a hand edit to the roster file (which R-5c5's fail-closed
   reload and R-5c8's guard both still validate on the next write).

### The spec corrections this pass owes

Drafted paste-ready for the conductor, per the plan's own list
(`plan:1988-1999`):

- The write-path `streamer_username` requirement (Ruling 4) against
  spec R3, R1's write path (`spec:291-292`), "The roster"
  (`spec:1106-1107`), and `spec:1417`'s "an optional streamer username":
  all four should read "required on the write path", since Task 2 and
  Task 6 make it so.
- `spec:1417`'s admin checkbox (Ruling 5): drop the checkbox from the
  add-member page's description. The command line is the only admin
  grant; 5c-ii's roster API refuses an `admin` key anywhere in the
  submitted form.
- `spec:1223`'s three channels become two: `ChannelHeader` and
  `ChannelServiceToken` are dropped (conductor-owed fold 12,
  "constants land with the source that asserts them").
- The spec's 5c row (`spec:2381`) splits into 5c-i and 5c-ii, each
  carrying its own ceiling: 5c-i closed under its 6.5M ceiling (see
  budgets below); 5c-ii's ceiling is its own plan's.
- R23's reservation against actuals (Ruling 6): rung 5's reservation
  moves from 15.75M to about 37M, 5c at about 18M across its two
  passes; the web surface standard's section 15 "4.5M ceiling is
  unchanged" line needs the same update.
- R24's 5c review shape across the split: the pass-end
  `web-auth-security-reviewer` ran once for 5c-i's own diff
  (`cb9ecc3..657f9ec`), not deferred to a single review spanning both
  halves; 5c-ii's plan should state its own review scope explicitly.
- Review finding 11 against 5b-ii's fold 7 (R-5c10): the 5b-ii close
  routed finding 11 ("doctor names no offending key") to 5c as a
  narrower residual than fold 7 anticipated; R-5c10 discharges it in
  full (`TestDoctorCmd_SourceLineNamesFailingKey`).
- The public-URL path rule (R-5c9) against `spec:1033`'s "Link
  hostnames derive from `DUBPLATE_PUBLIC_URL`": the spec line should
  add "the link's path is `<prefix>/login/link`; the public URL itself
  carries no path", since a path on the configured value now refuses
  at load.
- Review minor 6's refusal (R-5c18) against the 5b-ii close's routing:
  the 5b-ii close listed minor 6 as routed to 5c's security-fold Task 1
  for a first proposal; this pass's disposition is refusal of the code
  fix with a runbook sentence carried to 5d, not a code change, and the
  5b-ii close entry should note that outcome rather than leave the
  routing open.

### Both budgets

Tokens: about **5.3M in subagent tokens through the security fold**,
against a **6.5M execution ceiling** with the flag at **5.2M** (crossed
mostly by the security fold, per the conductor's supplied figures).
This close's own spend is not separately measured by a command and is
not included in the 5.3M figure. Attended time: **zero execution
sittings**; **zero planning misses** (Task 5's vacuous proof of named
mutation 3 was an execution miss the reviewer caught mid-pass, not a
plan ambiguity that a planning question would have surfaced). Gate:
`bash scripts/check.sh` measured exit 0 at this close's own re-run (see
"The gate evidence" above); no wall-clock total was requested for this
close.

### What a later pass would be wrong to rediscover

- **The roster's cross-process design is `flock(2)` on the directory
  inode plus a per-process mutex, in that lock order, verified by six
  real OS processes racing 25 writes each** (the security review's own
  probe). 5c-ii's roster API and any future writer must take the same
  order: `flock` before `r.mu`, `r.mu` held only for the swap, so
  readers never wait on another process's lock.
- **`streamer_username` uniqueness is keyed on the effective principal**
  (username when set, else the normalized address, `strings.EqualFold`),
  not on username-to-username comparison alone. A future write path
  (5c-ii's roster API, rung 12's importer) that adds its own uniqueness
  check without going through `validateRoster`/`addTo` reopens m1.
- **The admin flag is refused on every channel but the command line, in
  the write path itself**, not in any caller's own check. 5c-ii's roster
  API inherits this refusal by construction; it does not need, and must
  not add, its own admin-channel gate in front of `AddMember`.
- **`doctor`'s failing-key report reads through a keyed error type
  (`errors.AsType`), never a message-prefix convention.** A future
  source that returns a plain `fmt.Errorf` for a bad key will report
  with no key name; the keyed type is the only contract `doctor` reads.
- **The mail-ceiling ("kind" mode) enumeration bound is per-client
  only; the global outbound mail ceiling is a separate, group-wide
  lockout lever** with 11 or fewer roster members exempt from it at
  default rates. Both facts belong in the 5d runbook, not rediscovered
  from the limiter's source.

---

*End of draft. Corrections above are drafted for the conductor's fold;
none has been applied to `docs/HISTORY.md`, `docs/STATUS.md`, or the
spec.*
