# Draft HISTORY entry: Rung 5, pass 5b-i (the auth core)

Draft ledger text for the conductor to fold into `docs/HISTORY.md` as a
new entry above "Configuration homes spec". Not itself a ledger; this
file is not read by the gate. Prose avoids spaced em dashes per the
ledger's register.

---

## Rung 5, pass 5b-i (the auth core: config contract, token, session, state records): closed 2026-09-12, lane 2

Executed against
`docs/superpowers/plans/2026-09-08-rung-5b-auth-core.md` (revision 2.1)
in the worktree at `.claude/worktrees/lane-2-5b`, branch `lane-2-5b`
from `7ba3cc1`. Task 0 (`opus`, fresh context) found no blockers beyond
the pointer drift the dispatch notes had already corrected. Tasks 1
through 9 ran across two windows: segment 2A (`wf_61f3cf53-665`, Tasks
1 through 3, closed `eddeadd` plus the `52b6a20` simplifier round) and
segment 2B (`wf_769302f4-cef`, Tasks 4 through 7, closed `ad6cb0c` plus
the `a97245e` simplifier round), then Tasks 8 and 9 landed at `1dd4c4b`,
`1857c10`, and `0f50558`. This entry is Task 10, the close.

### What landed

Task 1 (`615aa31`, `68860cd`) added `internal/identity`: a three-value
`Channel`, an `Identity` keeping `Principal` distinct from `Address`
(2.0 hook 5's first half), and `NormalizeAddress`, the one function
every keying, logging, and rate-limit lookup on a login address must
call.

Task 2 (`49d7859`, `opus`) changed `internal/config/config.go` and
`registry.go`: `Load` now requires every owned-now wall row the
registry carries rather than four names written by hand, closing risk
11 for every future rung by construction (A1). The identity-source
family is the wall's sixth classification: a rung declares a source by
its selector key, required keys, optional keys, and a validation hook
receiving the source's resolved values plus the `os.Environ()`
snapshot (A2, review B2, conductor-owed as applied), which is what
lets `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY`
validate their own relation without `internal/config` learning a rung
5 name.

Task 3 (`fdc5c7f`, fix `eddeadd`) added `cmd/dubplate-server/auth.go`:
`registerAuthConfig` declares `DUBPLATE_IDENTITY_SOURCE` as an
owned-now row and the `magic-link` source's nine required and two
optional names, retiring `internal/config/keys.go`'s pending "identity
source" row and the whole "auth and identity" reserved class in the
same commit (A3), the removal half of the rung 1b registry contract's
second exercise. Three supersession messages were rewritten:
`DUBPLATE_CONTRIBUTORS_FILE` now names `DUBPLATE_ROSTER_FILE`, and the
two Access rows describe the forward-auth seam generically. `main.go`
gained `newProductionRegistry`, the binary's one registry constructor.
The task's own handoff draft
(`research/handoffs/2026-09-12-rung-5b-task-3.md`) records its mutation
results in full and is the source for this entry's Task 3 row below.

Task 4 (`9a1a270`) added the four thresholds `DUBPLATE_MAGIC_LINK_EXPIRY`
(`15m`), `DUBPLATE_SESSION_LIFETIME` (`2160h`),
`DUBPLATE_SESSION_ABSOLUTE_CAP` (`4320h`), and
`DUBPLATE_NONCE_RETENTION` (`24h`), and `validateMagicLinkSource`: the
32-byte secret floor (R-5b8), the public URL's three conditions with a
whole-loopback-set carve-out (fold 8, ruled 2026-09-12) rather than the
literal name `localhost`, and the retention-vs-expiry relation (R31).

Task 5 (`b9c8ca5`, `opus`) added `internal/auth/token.go`: the
magic-link token, carrying the normalized address, expiry, mint time,
and a 128-bit `crypto/rand` nonce. Every field is base64url-encoded
before the length-prefixed MAC input's "." join (A4 fold 3, ruled
2026-09-12 as a correctness fix: raw joining breaks any address whose
local part carries a dot). Verification order is fixed: MAC through
`crypto/subtle.ConstantTimeCompare`, then expiry, then the nonce, so an
unauthenticated caller cannot fill the nonce store. The previous secret
verifies and never mints. The link is built from the configured public
origin, never a request value, with the token in the query string
(R-5b13).

Task 6 (`e294d47`, `opus`) added `internal/auth/session.go`: the
`__Host-dubplate_session` cookie (R-5b9), same wire shape as the token
under its own domain-separation label. Rolling refresh moves the issue
time and the expiry, never the anchor, so the absolute cap survives any
number of refreshes (A5). The serve verb now builds its shell through
`newServeShell`, passing the real cookie name to `NewStrip`, retiring
5a's declared coverage gap 4.

Task 7 (`ad6cb0c`) added `internal/auth/store.go`: the shared
substrate both state records sit on, one leaf (`auth`) with one
subdirectory per record kind, an opener creating both at 0700,
`RecordID` (full SHA-256 hex, untruncated, R-5b12), an atomic write at
0600, and `validID` refusing empty, `.`, `..`, and separator-bearing
ids. A confinement test writes through every entry point, including for
an address whose local part carries `..`, and walks the filesystem to
prove nothing escapes root and no raw address appears in a filename
(A6).

Task 8 (`1dd4c4b`, `opus`) added `internal/auth/nonce.go`: the
create-exclusive consumed-nonce store. The insert opens each record
with `O_CREATE|O_EXCL`, syncs the record and its directory, and reports
`ErrNonceReplayed` on collision (R11); `atomicfile`'s rename-over shape
cannot carry this write because it would silently replace the proof of
consumption. Each record carries consumed-at plus the resolved
`DUBPLATE_NONCE_RETENTION`, never the token's own expiry (conductor-owed
as applied), and a record the sweep cannot parse is retained and
counted until its mtime passes retention rather than deleted on sight
(conductor-owed as applied), because deleting a crash-truncated record
would make its token consumable a second time. The sweep runs at open
and is offered as a method; the interval driver is 5b-ii's (R-5b14).

Task 9 (`1857c10`, then `0f50558`) added `internal/auth/epoch.go`: one
record per normalized address holding the epoch a bump last set,
written through `atomicfile.WriteStream` so the overwrite fsyncs the
containing directory (R11). `CheckSession` refuses a session whose
**anchor** predates its address's epoch, and `CheckLink` refuses a link
whose mint time predates it (conductor-owed as applied: the anchor
rather than the issue time, because refresh moves the issue time and a
session refreshed once after a bump would otherwise be permitted
forever, R30's own failure mode reproduced on every request). `Read` is
the authority a subscriber that missed a notification can still
consult (A8). The `0f50558` follow-up fixed a real gap the task's own
review caught: both checks keyed the epoch lookup on the credential's
raw address rather than the normalized one, so a session or link whose
address differed only in case silently escaped revocation; both now run
through `identity.NormalizeAddress` before keying, matching `Bump` and
`Read`, with `TestCheckSessionNormalizesAddress` made table-driven over
both bump/check case-mismatch directions.

Three simplifier rounds ran across the segments rather than one per
task (`52b6a20`, `a97245e`, and the equivalent fold inside segment 2B):
`a97245e` folded the token's and session's near-identical wire-shape
helpers into three shared functions in `internal/auth/token.go`
(`encodeSigned`, `splitSigned`, `macMatches`), so the encoder, the
decoder, and the constant-time comparison cannot drift apart between
the two credentials.

### The exported surface

`Channel`, `Identity`, `NormalizeAddress` (`internal/identity`);
`Load`'s generic owned-row requirement, `IdentitySource`,
`Config`'s active-source accessor (`internal/config/config.go`,
`registry.go`); `registerAuthConfig`, `magicLinkSource`,
`validateMagicLinkSource`, the four thresholds
(`cmd/dubplate-server/auth.go`); `LinkToken`, `MintLink`, `VerifyLink`,
`LinkURL`, `encodeSigned`/`splitSigned`/`macMatches`
(`internal/auth/token.go`); `SessionCookieName`, `MintSession`,
`VerifySession`, `RefreshSession` (`internal/auth/session.go`);
`StoreDir`, `Open`, `RecordID`, `validID`, atomic write
(`internal/auth/store.go`); `NonceStore`, `OpenNonceStore`, `Insert`,
`Sweep`, `NonceSweep` (`internal/auth/nonce.go`); `EpochStore`,
`OpenEpochStore`, `Bump`, `Read`, `Subscribe`, `CheckSession`,
`CheckLink` (`internal/auth/epoch.go`).

### What the gate caught

Task 3 ran a fix round (`eddeadd`) closing a real gap its own review
found: the doctor-exit and selector-name assertions had run only
against scratch registries, so neither could fail if the real
`DUBPLATE_IDENTITY_SOURCE` name were declared wrongly; both moved onto
`productionRegistry()`. Task 9's `0f50558` follow-up closed the
case-normalization gap described above, found by the task's own review
rather than by a later pass. No task ran red and uncommitted this pass.

### The auth core's contract coverage, A1 through A8

| Contract | Go site | Proved by | Reach |
| --- | --- | --- | --- |
| A1 a registered owned row is a required row | `Load` (`internal/config/config.go:358-366`) | `TestLoad_RequiresEveryRegisteredOwnedNowRow` | Full |
| A2 one identity source selected by configuration, correlated set required by mechanism | `Load`'s `resolveIdentitySource` (`config.go:296-330`), `IdentitySource` (`registry.go`) | `TestLoad_IdentitySource`, `TestCollisionGate_IdentitySource*` (six rows) | Full |
| A3 rung 5's names declared in rung 5's own file, both retirements in one commit | `registerAuthConfig` (`cmd/dubplate-server/auth.go:39-101`) | `TestTemplateKeysCoverEveryOwnedNowRow`, `TestTemplateKeysClassifiedAgainstProductionRegistry` | Full |
| A4 the magic-link token cannot be forged, shifted, replayed, or outlived | `MintLink`, `VerifyLink` (`internal/auth/token.go`) | `TestLinkGoldenVectors`, `TestMACLengthPrefixesResistFieldShifting`, `TestVerifyLinkRefusals`, `TestVerifyLinkAcceptsPreviousSecret` | Full except the constant-time comparison's timing property (declared gap 1) |
| A5 a session cannot outlive its absolute cap, rolling refresh cannot extend it | `MintSession`, `VerifySession`, `RefreshSession` (`internal/auth/session.go`) | `TestRefreshHoldsTheAnchor`, `TestRefreshedSessionRefusedAtTheAbsoluteCap`, `TestSessionAndLinkCannotCross` | Full |
| A6 both state records named by one derivation, confined to one root | `RecordID`, `Open`, `validID` (`internal/auth/store.go`) | `TestRecordID_Vector`, `TestValidID`, `TestConfinement` | Full |
| A7 a consumed nonce cannot be consumed twice, across a restart or any accepted configuration | `Insert`, `Sweep` (`internal/auth/nonce.go`) | `TestInsertRefusesReplay`, `TestInsertRefusesReplayAcrossRestart`, `TestConcurrentInsertAdmitsOne`, `TestSweepPrunesOnCarriedExpiry`, `TestSweepRetainsMalformedRecord`, `TestSweepRetainsCrashTruncatedRecord` | Full except the directory-fsync's crash-consistency property (declared gap 2) |
| A8 the revocation seam exposes a read as well as a notification | `Bump`, `Read`, `Subscribe`, `CheckSession`, `CheckLink` (`internal/auth/epoch.go`) | `TestBumpReadDurable`, `TestCheckSession`, `TestCheckSessionRefreshedAfterBumpStillRefused`, `TestCheckLink`, `TestCheckSessionNormalizesAddress`, `TestReadAuthoritativeWithoutSubscriber` | Full |

### The resolutions, as ruled

Sixteen resolutions, R-5b1 through R-5b16, ruled by the conductor at
dispatch, 2026-09-08, stand as applied except where a later fold
corrected the reading (noted per row). Six carry a code site the marker
convention stamps; the other ten are process or plan decisions with no
single line to point at, and this table says so rather than inventing
one.

| Resolution | Ruling | Landed at |
| --- | --- | --- |
| R-5b1 | The pass splits; this plan is 5b-i | Plan only, no code site |
| R-5b2 | Auth keys declared in `cmd/dubplate-server/auth.go`, not `internal/auth` or `internal/config` | `cmd/dubplate-server/auth.go:39` |
| R-5b3 | All twelve names declared at 5b-i though six consumers arrive at 5b-ii | `cmd/dubplate-server/auth.go:39-58` |
| R-5b4 | `Load` requires every owned-now wall row generically, not by naming rung 5's selector | `internal/config/config.go:358-366` |
| R-5b5 | The per-source correlated set is a registry declaration family, not a widened integration class or a hard-coded switch | `internal/config/registry.go` `IdentitySource` type |
| R-5b6 | The two rotation-overlap names are the source's optional keys, never thresholds | `cmd/dubplate-server/auth.go:52-55` (`Optional`) |
| R-5b7 | `DUBPLATE_IDENTITY_SOURCE` lands as a literal row in the env template at the task that makes it owned-now | `env/dubplate-server.env.template:74` |
| R-5b8 | The secret length floor is 32 bytes | `cmd/dubplate-server/auth.go:68` (`secretLengthFloor`) |
| R-5b9 | The session cookie is named `__Host-dubplate_session` | `internal/auth/session.go:20` |
| R-5b10 | Four thresholds and their defaults (`15m`, `2160h`, `4320h`, `24h`) | `cmd/dubplate-server/auth.go:77-99` |
| R-5b11 | The state leaf is `auth`, one subdirectory per record | `internal/auth/store.go:18` (`leafDirName`) |
| R-5b12 | The record digest is full SHA-256 hex, untruncated | `internal/auth/store.go:94-99` (`RecordID`) |
| R-5b13 | The token rides in the link's query string, never a path segment | `internal/auth/token.go:63`, `:117` |
| R-5b14 | The nonce sweep's interval is a package constant; the driver arrives at 5b-ii | **As landed, partial**: the sweep runs at open and as a callable method (`internal/auth/nonce.go:80-90`, `:143`); no interval constant exists yet, because no driver calls it on a schedule at 5b-i. The plan's own text ("one hour, a package constant with its reason stated") describes 5b-ii's shape prematurely; the close corrects this as a plan-wording note rather than a gap this pass created, since R-5b14's operative clause ("the driver arrives at 5b-ii") is what landed. |
| R-5b15 | 5b-i writes no `internal/serve` file except the binary's construction call | Confirmed: `git diff` against this pass's base shows no `internal/serve/*.go` file touched; `cmd/dubplate-server/serve.go` is the one wiring change (Task 6) |
| R-5b16 | A fresh-context `web-auth-security-reviewer` runs at 5b-i's own pass end, and again at 5b-ii's | Owed to the conductor's post-Task-10 actions; not run inside this pass (see "What this pass owes" below) |

### The mutation table

Named mutations as the plan predicted them per task. Task 3's row is
quoted from that task's own persisted handoff
(`research/handoffs/2026-09-12-rung-5b-task-3.md`), the only task this
pass that wrote one. The other eight tasks' per-task pass-execute
transcripts (`wf_61f3cf53-665`, `wf_769302f4-cef`, and Tasks 8-9's own
runs) are session-bound and were not persisted as repo artifacts, so
this close compiles their rows by cross-referencing the plan's named
mutations against the landed tests and doc comments rather than by
re-running each one; every named test cited below exists at the stated
site and its assertion covers the mutation's stated failure mode.

**Task 1** (5 named): folding the principal into the address field,
adding a fourth channel value, admitting the zero value, lowercasing
without parsing, and comparing the cap against rune count instead of
byte length each map to a distinct row in `identity_test.go`
(`TestIdentity_PrincipalDistinctFromAddress`,
`TestChannel_StringAndValid`, `TestIdentity_Valid`,
`TestNormalizeAddress`'s unparseable-address and multi-byte-boundary
rows).

**Task 2** (7 named): dropping the registered-row requirement, requiring
a `Pending` row, requiring the selector with no source declared,
skipping the validation hook, dropping the environment-snapshot
parameter, misclassifying a source name as unknown, and printing a
source's resolved value in `doctor` each map to
`TestLoad_RequiresEveryRegisteredOwnedNowRow`,
`TestCollisionGate_IdentitySourceCarveOut`,
`TestCollisionGate_IdentitySource`, `TestLoad_IdentitySource`'s hook
row, and `internal/config`'s doctor-output tests. Mutation 5 (dropping
the `os.Environ()` parameter) is a compile-time failure by
construction: `validateMagicLinkSource`'s signature requires it.

**Task 3** (7 named, from the task's own handoff): declaring the twelve
names without deleting the reserved class fires a `Registry.claim`
panic at registration rather than a test failure. Deleting the reserved
class without declaring the source fails
`TestTemplateKeysClassifiedAgainstProductionRegistry`, naming the five
SMTP names unknown. Clearing `Pending` instead of deleting the row
fires louder than predicted: `AddWallRow`'s own switch panics
immediately, because the row's name was never `DUBPLATE_`-prefixed, so
the plan's "Load-time demand for `identity source`" wording is
corrected by this close. Registering the selector while skipping the
template row fails `TestTemplateKeysCoverEveryOwnedNowRow`, naming
`DUBPLATE_IDENTITY_SOURCE` (R-5b7's whole argument, fired). Rewriting
`DUBPLATE_CONTRIBUTORS_FILE`'s message by deletion fails the frozen
template's own classification. Two mutations are **confirmed unfired**:
`productionRegistry` restating the register calls instead of calling
the shared constructor (a maintenance defect, not behavioral, proven by
`go test ./cmd/dubplate-server/...` staying green), and
`TestCollisionGate_IdentitySourceCarveOut` reading the production table
instead of a fixture row (passes vacuously with no pending row left to
exercise, which this task's own fixture-row rewrite closed).

**Task 4** (7 named): comparing the secret's rune count instead of byte
length, moving the floor's exclusivity, admitting plain HTTP on any
host, comparing `url.Host` instead of `url.Hostname()`, dropping the
non-empty-host condition, reversing the retention/expiry inequality,
and including the offending value in an error each map to a row in
`cmd/dubplate-server`'s auth-config validation tests (the multi-byte
secret row, the at-the-floor row, the plain-HTTP row, the
`http://localhost:8080` port row, the `example.test` row, the
equal-and-below retention rows, and the no-secret-in-errors row).

**Task 5** (7 named, `opus`): concatenating fields without length
prefixes, joining without per-field encoding, dropping the
domain-separation label, checking expiry before the MAC, returning the
nonce on failed verification, and building the link from a request host
each map to `TestMACLengthPrefixesResistFieldShifting`,
`TestLinkGoldenVectors`'s dotted-address row, `TestVerifyLinkRefusals`'s
cross-purpose and both-broken rows, and `TestLinkURL`. Comparing MACs
with `==` or `bytes.Equal` is **reported unfired by design**: no
gate-resident test can prove a timing property; `token.go:264-266`
carries the reason `crypto/subtle.ConstantTimeCompare` is the call
(declared gap 1).

**Task 6** (6 named, `opus`): moving the anchor on refresh, dropping
`Secure` or setting `Domain`, reusing the link's domain-separation
label, checking the cap against issue time instead of anchor, moving
the anchor alongside the issue time, and leaving the serve verb passing
an empty cookie name each map to `TestRefreshHoldsTheAnchor`,
`TestSessionCookieAttributes`, `TestSessionAndLinkCannotCross`,
`TestRefreshedSessionRefusedAtTheAbsoluteCap`, and `serve_test.go`'s
cookie-name construction and fallthrough-removal rows.

**Task 7** (5 named): naming a record from its value instead of its
digest, truncating the digest, using rung 4's 0755/0644 permissions,
admitting a separator-bearing id, and joining a path in a caller
instead of through `StoreDir` each map to `TestConfinement`,
`TestRecordID_Vector`, `TestOpen_CreatesDirectoriesWithPermissions`, and
`TestValidID`. The caller-joins-its-own-path mutation is **reported
unfired by design**: nothing fails today because no second caller
exists yet to diverge; `store.go:36-37`'s doc comment states the reason
(the cost lands at 5d, when the unit's `ReadWritePaths` grant needs one
static path).

**Task 8** (5 named, `opus`): inserting through `atomicfile.WriteFile`,
checking existence before creating, sweeping on file mtime instead of
carried expiry, and deleting a malformed record on sight each map to
`TestInsertRefusesReplay`, `TestConcurrentInsertAdmitsOne`,
`TestSweepPrunesOnCarriedExpiry`, `TestRecordCarriesRetentionNotTokenExpiry`,
`TestSweepRetainsMalformedRecord`, and
`TestSweepRetainsCrashTruncatedRecord`. Dropping the directory `Sync`
call is **reported unfired by design**: crash consistency has no
gate-resident test; `nonce.go:234-260`'s `writeAndSync` doc comment
carries the argument (declared gap 2).

**Task 9** (6 named): checking only the session's issue time instead of
the mint time, keying the epoch on the raw address, comparing at the
epoch exclusively, and offering publication without a read each map to
`TestCheckLink`, `TestCheckSessionNormalizesAddress`,
`TestCheckSessionRefreshedAfterBumpStillRefused`, and
`TestReadAuthoritativeWithoutSubscriber`. Writing through
`atomicfile.WriteFile` instead of `WriteStream` is **reported unfired by
design**: the difference is the directory fsync, whose failure needs a
power cut; `epoch.go:53-58`'s doc comment carries the reason.

### The declared coverage gaps

**Six declared by the plan's own Task 10 template, corrected against
what execution found, against the running total of sixty-seven after
rung 8c, bringing the total to seventy-three.**

1. **A constant-time comparison has no gate-resident proof** (Task 5,
   mutation 4). `crypto/subtle.ConstantTimeCompare`'s use is provable
   only by reading the call; the criterion is a reviewer's read, not a
   test.
2. **Crash consistency rests on two fsync calls rather than on a test**
   (Tasks 8 and 9). A power cut is not a thing this gate reproduces;
   the doc comments at `internal/auth/nonce.go:234-260` and
   `internal/auth/epoch.go:53-58` carry the argument instead.
3. **No handler, no limiter, and no origin check exercises any of
   this** (R-5b1). 5b-ii is the first real caller of the token, the
   session, and both stores.
4. **The identity-source family has exactly one declared source**
   (`magic-link`, `cmd/dubplate-server/auth.go:40`). Its multi-source
   behavior (`TestIdentitySource_SharedKeyAcrossSources` and the
   collision-gate rows) is proved only against fixture registries; the
   second real source arrives with a forward-auth front nobody has
   built.
5. **Nothing validates a registered threshold's value except the two
   this pass's hook reaches** (Task 4, security review finding 24 from
   5a). Task 2's hook takes the environment snapshot, so
   `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY` validate
   at load and their relation with each other (R31). Every other
   registered threshold this pass declares (`DUBPLATE_SESSION_LIFETIME`,
   `DUBPLATE_SESSION_ABSOLUTE_CAP`) still accepts any parseable value,
   which is the rung 1b wall contract's own property rather than a gap
   this pass created.
6. **The sweep's interval driver does not exist** (R-5b14). The sweep
   is proved by direct call rather than by elapsed time; no goroutine
   calls it on a schedule until the serve verb's lifecycle wires one at
   5b-ii.

### What this pass owes

- **The fresh-context `web-auth-security-reviewer` read** (R-5b16) did
  not run this window, for budget reasons the overnight plan states;
  it is owed to the next attended session, before the merge ritual
  completes per CLAUDE.md's build-process section.
- **`code-simplifier` over the accepted diff**: three batched
  per-segment rounds ran (`52b6a20`, `a97245e`, and the equivalent fold
  inside segment 2B) rather than one dispatch per task, a deviation
  from CLAUDE.md's stated cadence, recorded here as the fold this
  window applied rather than left silent.

### The spec corrections

Drafted for the conductor to apply to
`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md` as dated
errata:

1. **R23's ceiling table and the pass split** (R-5b1): 5b becomes 5b-i
   and 5b-ii, both ceilings and the rung total tracked against R23's
   15.75M reservation (set at 5a's own close).
2. **R24's review shape** (R-5b16): a fresh-context
   `web-auth-security-reviewer` read at each half's pass end, not one
   read "inside 5b" as R24's original wording states.
3. **R29's wording**: `Load` requires every owned-now wall row rather
   than naming `DUBPLATE_IDENTITY_SOURCE` by hand, closing risk 11 for
   every rung rather than one key.
4. **The env template's row assignment** (R-5b7): one literal row
   (`DUBPLATE_IDENTITY_SOURCE=magic-link`) lands at 5b-i rather than at
   5d, because `TestTemplateKeysCoverEveryOwnedNowRow` is a hard gate.
   `spec:2205`'s "commented pending rows only" and `spec:2170`'s row
   assignment both take the correction; the other six non-secret names
   stay 5d's.
5. **Two of `spec:2168`'s exit-gate assertions belong elsewhere**: the
   roster-removal half of the third assertion is 5b-ii's, and the
   verify-subrequest half of the eighth is 5d's, since `spec:2170`
   gives the endpoint to 5d.
6. **The staleness pre-task's `serve` reserved-class bullet**
   (`spec:2297-2300`) is stale; 5a removed that class.
7. **The `keys.go` pointers at `spec:1506`** (`:177-178` and
   `:171-176`) resolve to `:165-166` and `:159-162`.
8. **The `atomicfile` pointer at `spec:434`** (`:45-56`) resolves to
   `:55-91`, with the directory fsync at `:82-89`.
9. **The epoch's session check reads the anchor, not the issue time**
   (review B1). `spec:1085-1086` says "a session whose issue time
   predates the epoch is refused"; as landed the check reads the
   anchor, the field refresh never moves.
10. **The token's field list gains the mint time.** `spec:1016-1017`
    enumerates three fields where R30 (`spec:640-642`) requires a
    fourth; the landed token carries four.
11. **The serialization claim covers the wire form.** `spec:1017-1019`
    states the claim; the landed token states the encoding
    (base64url-then-join), so no field value can carry the separator.
12. **The nonce record's carried value.** `spec:1664-1668` says the
    record carries "its expiry" without saying whose; the record
    carries consumed-at plus the resolved retention.
13. **R38's carve-out covers the loopback set** (`spec:758-767`), not
    the literal name `localhost` alone.
14. **The non-secret name count.** `spec:1640-1645` says "eight" and
    enumerates seven; `spec:2170` gives 5d "its eight literal rows".
    One (`DUBPLATE_IDENTITY_SOURCE`) lands at 5b-i, so six remain for
    5d.
15. **R33 and R24 belong to 5b.** `HISTORY:393-394` assigns R33 to this
    pass and R-5b16 corrects R24's one-read-inside-5b wording.
16. **R1's own `keys.go` pointer at `spec:284`** (`:176-177`) moves to
    `:165-166`, the same move as item 7 above.
17. **The ceiling erratum's own pointer.** The plan cited
    `spec:2078-2087` for it in revision 1; the correct pointer is
    `spec:2042-2050`, and the rung reservation's live baseline is the
    15.75M that erratum set.
18. **R-5b14's own wording** describes the one-hour interval constant
    as landing at 5b-i; as executed, only the sweep-at-open call and
    the callable method landed, and the interval driver (constant
    included) is 5b-ii's in full, matching the resolution's operative
    clause.

### The 5b-i to 5b-ii handoff

5b-ii inherits: the identity leaf and its address normalizer, which
every keying, logging, and rate-limit lookup runs through
(`spec:1161-1163`); the token's verifier and the nonce it returns, to
be inserted after verification and never before; the token's version
label as the only place a later field, a destination included, may be
added (R33, `spec:690-692`); the session's minter, verifier, and
refresher, with the rule that refresh never moves the anchor, that the
anchor is the session's original issue time, and that refresh counts
same-origin requests only (`spec:1081-1082`); the epoch's two refusal
checks, the session check reading the anchor and never the issue time,
applied to a session on every request and to a token at consumption,
plus the roster re-check that refuses a link for an address removed
from the roster (R30, `spec:2168`), which 5b-ii still owes; the seam's
`Read` as the authority a credential store stamps from; the
state-directory opener and its naming rule, which the roster's own leaf
sits beside rather than inside; `registerAuthConfig` and the declared
key set, which 5b-ii's remaining thresholds join rather than re-derive;
the constraint that `internal/auth` never reads the environment; 5a's
own carries unchanged (`ClientIP` never re-derived, the origin check
inserts after the strip and after normalization, `Origin` and
`Sec-Fetch-Site` survive ingress, `CrossOriginProtection.Check` returns
nil when both are absent, every URL builds from the configured public
origin and never `r.Host`); the merge carry on
`internal/serve/server.go:161`, forwarded to 5d; the 30-second read
timeout and the `http.MaxBytesReader` alternative (R-V), until
`spec:1590`'s cap lands; and the two module dependencies R26 budgets,
both of which 5b-ii is the first pass to take. **The nonce sweep's
interval driver, a package constant naming its one-hour period, is
5b-ii's to build alongside the goroutine that calls `Sweep` on it**
(R-5b14, corrected above). **The fresh-context security review is owed
before 5b-ii's own merge**, covering both halves together per R-5b16.

### Both budgets

Tokens: this window's execution against the 5.0M ceiling (4.0M flag)
ran over, landing at roughly **5.5M** for Tasks 1 through 10 combined
with their fix and simplifier rounds. **The overage is recorded against
R23's live 15.75M reservation rather than funded from another pass**
(ruled, conductor 2026-09-12, applying arithmetic 4.25 + 5.0 + 4.5 +
4.5 + 3.0 = 21.25M against the live 15.75M). Attended time: **zero
planning misses** (Task 0 found no blocker the dispatch notes had not
already corrected) and **zero execution sittings** in this window; the
rung 5 security cycle and the plan-prep tail for 5b-ii were both cut
from the window for budget reasons rather than surfaced mid-task, which
is why they appear above as owed rather than as a sitting.
