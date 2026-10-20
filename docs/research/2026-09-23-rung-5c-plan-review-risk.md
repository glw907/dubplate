# Rung 5c plan review: domain risk (authentication and authorization)

Target: `docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md` at `dba1679` (5c-i in full, 5c-ii outlined). Rubric: the rung 5 spec's threat model (`spec:1909-1980`) and fork 1 (`spec:178-195`). Do-not-relitigate: the "Attacked and found sound" list of `2026-09-23-rung-5b-ii-security-review.md`. I checked every claim against the tree at `dba1679` (`internal/auth/roster.go`, `handlers.go`, `limit.go`, `internal/identity/identity.go`, `cmd/dubplate-server/serve.go`, `auth.go`, `internal/serve/serve.go`, `prefix.go`).

## Verdict

5c-i is sound in shape, and I found no blocker. Two network paths could reach `ChannelCommandLine`, and neither does: the roster API hard-codes the cookie channel, and the seam only produces `ChannelCookie`. The admin flag is still read from the roster. The origin check covers the API's POST, the prefix fix sends every mailed link to an owned path, and moving the server-file check and the document route to 5c-ii leaves no gap in what 5c-i ships.

There are three majors:
1. **The fail-closed limiter fold (R-5c3) swaps finding 3 for something worse.** Finding 3 was a bounded mail nuisance. The fold makes it a cheap, sustained lockout: nobody in the group can sign in while the flood runs. The plan's cost math understates the cost of the attack by about an order of magnitude.
2. **The stat-and-reload design (R-5c5) does not fix the order of the stat and the read.** One ordering pins a stale roster until the file next changes, so a member removed by hand keeps signing in. Closing that gap is the whole point of the fold.
3. **The last-admin invariant contradicts `CreateRoster`.** `CreateRoster` writes an empty roster through the same guarded function. Task 6 will hit this and resolve it ad hoc, in the privilege code.

The rest are minors, several of them cheap now and expensive later.

## Findings

### 1. The fail-closed limiter lets a small IPv6 allocation lock the whole group out of sign-in (major, OWNER FORK)

**Location:** R-5c3 (plan:411-427); Task 3 outcome 3 (plan:1072-1081); the residuals list (plan:681-685).

**Scenario.** Unknown addresses get no mail (`handlers.go:192-196`), but every request still charges the per-address map (`limit.go:78`). The attack runs in three steps:
1. The attacker spreads requests for distinct, made-up addresses across a few thousand /64s. That is a trivial slice of one /48, which tunnel brokers hand out free.
2. A per-address bucket charged once refills in 12 minutes (5 per hour at burst 5). Keeping 10,000 of them unrefilled therefore costs about 14 requests per second. Each /64 is admitted 20 per hour, so that takes about 2,500 /64s.
3. Once the map is full and nothing has refilled, Task 3 refuses every new key. Every member who has not requested a link in the last few minutes gets 429 `limiter full`.

The lockout lasts as long as the flood, not "up to an hour". The consumption map is also per-client and fail-closed, so the same technique blocks a member holding a valid link: 10,000 active /64s, each re-touched every 2 minutes.

**Evidence.** The plan prices the attack at "10,000 /64s" and "up to an hour" (plan:419-422). That prices only the per-client map. The per-address map is the cheaper target, and the per-client-first order does not protect it: the attacker's clients are all admitted. The LRU eviction the fold replaces only reset a victim's bucket, bounded by the 60-per-hour global mail ceiling. For the per-client maps, fail-closed buys nothing: evicting an attacker-controlled client key hands the attacker nothing it does not already have, since it controls the distinct clients.

**Options.**
- **(a) Recommended.** Charge the per-address key only after the roster lookup, and only for a known address. The map is then bounded by roster size, cannot fill, and has nothing to evict. This fixes finding 3 at its root. Under the kind answer, known versus unknown is already disclosed, so a per-address 429 on a known address leaks nothing new. Under the generic answer, a known address over its limit answers `requested` and skips the send, so the response bytes stay uniform. The per-client maps keep LRU eviction.
- **(b)** Keep the plan's fail-closed fix, and record the lockout residual with the corrected cost and duration.
- **(c)** Fail-closed on the per-address map only, with LRU on the per-client and consumption maps. This halves the exposure but keeps the cheap per-address lockout.

Geoff routed finding 3 into 5c "with its fix", so changing the fix is his call.

**Proposed fold.** Put the corrected arithmetic in R-5c3, take option (a) if ruled, and point the pass-end reader at whichever option lands.

### 2. The reload does not fix the order of the stat and the read, so a stale roster can be pinned (major)

**Location:** R-5c5 (plan:451-469); Task 1 outcomes 1 and 4 (plan:813-817, 839-842).

**Scenario.** Suppose `Lookup` reads the file and then calls `os.Stat` on the path. A rename can land between the two: the read sees inode A, the stat records inode B. The cache then holds A's content under B's stat. Every later stat matches, so no reload fires until the file changes again, and a member removed by the rename keeps resolving. The writer has the same race: "re-reads, and only then releases", but a hand edit, which takes no lock, can land between its rename and its stat. The writer then caches its own content under the hand edit's stat, and the hand edit is ignored.

**Evidence.** The plan names the comparison fields but never says which file's stat is stored. The first ordering an implementer reaches for is `os.ReadFile` followed by `os.Stat`, which is the unsafe one. None of the Task 1 tests can catch it, because they rewrite the file between calls, never during one.

**Proposed fold.** Every read, whether an open, a reload, or a writer's re-read, does `os.Open`, then `f.Stat()`, then reads from the same descriptor, and stores that descriptor's stat. Add a named mutation, "stat the path after reading", with a rationale the reviewer reads if no test can fire it. The only remaining race is the harmless one: a new stat over old content just triggers one more reload.

### 3. `CreateRoster` writes an admin-less roster through the function the last-admin guard lives in (major)

**Location:** Task 1 Interfaces (plan:807-809); R-5c8 (plan:514-536); Task 6 outcome 4 (plan:1358-1361); Task 7 rows (plan:1470-1471).

**Scenario.** `CreateRoster` writes `"members": []` (`roster.go:176-186`). Task 1 routes it through "one unexported locked read-modify-write function ... which Task 6's guard lives in". Task 6's guard refuses every write that "would name no admin". Built as written, the guard refuses the bootstrap. The implementer then has to invent an exemption, such as an `allowEmpty` flag on the locked function, which a later demote method could pass by mistake. Or the implementer skips `CreateRoster` in the guard without stating it.

**Evidence.** Task 7's row "exists, empty (left by an earlier failed bootstrap)" assumes empty rosters are still written, so the plan has not settled which is the rule.

**Proposed fold, recommended.** Change `CreateRoster` to take the first member, which must be an admin, and write it under the directory lock in one step. The store then never writes an admin-less roster, and the guard has no exception. The "interrupted bootstrap" state disappears, and the verb's absent-roster pre-check (plan:1478-1482) becomes the store's own rule. Only a hand edit can then produce an empty or admin-less file, which the guard's remaining row covers. `CreateRoster` has one caller, the verb, so nothing is exported ahead of a caller.

**Alternative.** State the create path's exemption explicitly in R-5c8 and make it structural: a separate function, not a flag.

### 4. Authorization is checked at the gate but not under the write lock, and the ungated API is not fail-closed (minor)

**Location:** Task 8 Interfaces and outcomes (plan:1543-1570); Task 9 mutation 2 (plan:1694-1695).

**Scenario 1.** The gate reads `Admin` when the request arrives. `AddMember` re-reads the disk under the lock but never re-checks that the actor is still an admin. If a hand edit removes the actor's admin flag between the gate and the lock, the write proceeds on stale authorization. At rung 5 the window is milliseconds. Rung 6's criterion 6, where "a concurrent self-demotion ... leaves one", needs authorization atomic with the write.

**Scenario 2.** If `NewRosterAPI` is ever mounted without `RequireAdmin`, as Task 9's own mutation 2 does, its GET lists every address to anyone. Only a test catches that.

**Proposed fold.**
- On any channel but `ChannelCommandLine`, `AddMember` requires the actor to be an admin in the disk copy it re-read under the lock. That is one check next to the admin invariant, and it makes the threat-model claim "only a member whose roster record holds the admin flag can change the roster" a property of the store.
- `NewRosterAPI` answers the gate's 404 when the context carries no verified identity with `Admin` true. Alternatively, export only the gated composition and keep the raw API unexported.

### 5. The 404 uniformity rule covers the response body but not the headers or the terminator (minor)

**Location:** R-5c12 (plan:577-598); Task 8 outcome 1 and criteria (plan:1555-1560, 1594-1606).

**Scenario 1.** For a verified non-admin with a same-origin request, the seam returns a `RefreshedValue`. Suppose the gate sets the cookie "on success" meaning successful verification rather than a successful admin check. The non-admin's 404 then carries a `Set-Cookie`, which a nonexistent path never does.

**Scenario 2.** The gate's 404 body can differ from the owned terminator's (`http.NotFound`, `internal/serve/prefix.go:131-135`). Either difference tells a non-admin that `/roster` exists.

The route is public in open-source code, so this is minor. But risk 14 asks for it, and the fix is free.

**Proposed fold.**
- The gate refreshes only after the admin check passes.
- Task 8's refusal rows assert identical status, headers, and body, with no `Set-Cookie` on any row.
- Task 9 adds one shell row: a non-admin GET on `<prefix>/roster` is byte-identical to a GET on an unregistered owned path.

### 6. The cross-process lock wait can stall every authenticated request, and `busy` cannot reach a client (minor)

**Location:** Task 1 outcomes 1 and 2 (plan:813-833); R-5c5 "reload under the in-process mutex" (plan:457); Task 8 outcome table `ErrRosterBusy` (plan:1586).

**Scenario 1.** A serve-side writer holds `r.mu` while it polls `flock`. `Lookup` takes no `ctx` and reloads under that same mutex. So while any other process holds the directory lock, every seam verification blocks, and so does the login lookup. Examples: a verb suspended with Ctrl-Z mid-write, or an operator following a "wrap hand edits in `flock(1)`" habit. The stall lasts up to the writer's `ctx`.

**Scenario 2.** The API's `ctx` is the request's own. That context has no deadline shorter than the server's 30-second read and write timeouts (`internal/serve/server.go:37, 69`). So `ErrRosterBusy` fires only after the response can no longer be written.

**Proposed fold.**
- Acquire the directory lock before taking the in-process mutex that readers share. Alternatively, give reads their own lock, so readers never wait behind a writer that is waiting on another process.
- The roster API bounds `AddMember` with its own short deadline, for example 5 seconds, so the 503 `busy` row is reachable and testable.

### 7. `streamer_username`'s shape: leading characters, `..`, and `_` (minor, OWNER FORK on `_`)

**Location:** R-5c6 (plan:471-495); Task 2 outcome 3 (plan:971-973).

**Scenario.** `[A-Za-z0-9._-]` admits `..`, `.`, `-rf`, and `_`. The principal is the value a later rung puts on the wire, in the streamer's reverse-proxy user header at rung 12. It could also land in a path or an argv if any later tool leans on it.

Loosening the rule later is free. Tightening it strands records that rung 6's import then refuses. The plan already carries the SQL `LIKE` question to rung 12. A second question belongs beside it for rung 12 to verify: a `streamer_username` that case-insensitively equals an existing streamer admin account maps the member onto that account. `EqualFold` uniqueness inside the roster cannot see the streamer's own accounts.

**Options, for `_`.**
- **(a) Recommended.** Require an alphanumeric first character, `[A-Za-z0-9][A-Za-z0-9._-]*`, and keep `_`. This refuses `.`, `..`, and leading-dash values at no cost to any plausible account name.
- **(b)** Also refuse `_` until rung 12 settles the `LIKE` question. This is the safer direction, but it refuses a common character.
- **(c)** As planned.

**Proposed fold.** Take (a). Add the account-collision question to the rung 12 carry in Task 10's coverage gaps.

### 8. `ChannelCommandLine` becomes a valid `Identity` channel, and the channel doc comment is already wrong (minor)

**Location:** R-5c7 (plan:497-512); Task 6 outcomes 1-3 (plan:1350-1357).

**Scenario.** `Identity.Valid()` accepts any `Channel.Valid()` value (`identity.go:132-133`). Adding `ChannelCommandLine` to `Valid` makes a verified network `Identity` carrying the command-line channel pass validation. The admin grant then hinges on every future network caller passing a hard-coded constant rather than `verified.Channel`. Take a rung 12 verifier that builds an `Identity` from a header or token, plus a handler that forwards the identity's channel. Those two small mistakes compose into an admin grant. The one seam test proves that the seam never emits the channel. It cannot prove that no future verifier does.

Separately, the existing `Channel` doc comment (`identity.go:67-69`) says "Admin-only verbs reject a cookie-channel identity". That is the opposite of what `RequireAdmin` does. Task 6 only mentions rewording its first sentence.

**Proposed fold.**
- `Identity.Valid()` refuses `ChannelCommandLine`, since it is an audit channel, not an assertion a verifier can make. `RequireAdmin` refuses any identity whose channel is not a network channel.
- Mutation: "`Identity.Valid` accepts command-line" makes a new row fail.
- Task 6 rewrites the whole `Channel` comment, including the stale sentence.

### 9. The residuals list is missing two outcomes, and one review-focus claim overstates the guarantee (minor)

**Location:** residuals (plan:681-685); review focus 1 (plan:694-696); R-5c13 (plan:600-612).

The plan's list of residuals leaves out two outcomes:
- **A malformed hand edit takes every signed-in member offline until it is fixed.** R-5c5 fails closed, so the seam refuses all requests. This is defensible, but it is the most likely accident in the pass and should be stated, with `doctor` as the recovery.
- **No path, except a hand edit that leaves no audit record, promotes an existing member to admin.** `member add --admin` on an existing address is a duplicate, and fold 8 removes the web control. Whether to wait for rung 6's grant verb is an owner or rung 6 question. Record it rather than leave it to be found.

Review focus 1 expects "nothing lost" while a third party hand-edits the file. A hand edit takes no lock, so one that lands inside a writer's read-modify-write window is overwritten. The window is milliseconds, and the plan should say so rather than promise it.

**Proposed fold.** Add both residuals and correct the review-focus wording.

### 10. The pass-end security read is not given the findings it is meant to verify (minor)

**Location:** the conductor's action 4 (plan:1822-1832).

**Scenario.** The reader receives the threat models, both "found sound" lists, and the diff, but not the 5b-ii findings that Tasks 1-5 fold: majors 1-4 and minors 7, 8, 10-13, 15, 17, and 19. Without them, the only check that each fold actually closes its finding is the in-chain `diff-reviewer`, which grades against the plan's criteria rather than against the attack.

**Proposed fold.** Add the routed findings as a "verify fixed" input. Beyond R-5c3, the gate's 404, and the lock, point the reader at findings 1 and 2 of this review: the corrected lockout cost, and the order of the stat and the read. With that added, R24's read has everything in 5c-i to grade: the API, the gate, the verb, and the fold.

### 11. The `member` verb needs every magic-link secret in the operator's environment (minor, route to 5d's runbook)

**Location:** R-5c14 (plan:614-628).

**Scenario.** `config.Load` runs `validateMagicLinkSource`, which requires all nine keys, secrets included. The verb must run as the directory owner (`music`), and the env file is root-readable. The natural workaround is `sudo -u music env $(cat /etc/dubplate/...) dubplate-server member add ...`, which puts every secret into argv, readable by any local user through `/proc/*/cmdline`.

**Proposed fold.** Either resolve only `DUBPLATE_ROSTER_FILE` and the selector for `member`, or record in Task 10's 5d handoff that the runbook invokes the verb through `systemd-run --uid=music -p EnvironmentFile=...` or an equivalent path, never through argv.

### 12. What the split defers: the 5c-ii header set must cover the JSON routes, and the link tests should pin both public-URL forms (minor)

**Location:** 5c-ii outline item 5 (plan:1939-1942); Task 4 criteria (plan:1231-1238); Task 8 POST (plan:1564-1570).

What 5c-i ships has no hole: the mailed link reaches the owned terminator, which is silent, and nothing deploys until 5d.

**Two items belong in the outline.**
- The served header set (`nosniff`, `Referrer-Policy: no-referrer`) applies to every owned response, the roster API and the 5b-ii JSON handlers included, not only documents.
- 5c-ii's registration of the `login` subtree must be proved through the prefixed link test.

**Two small pins belong in 5c-i.**
- Task 4's routing test runs for both `https://h` and `https://h/`, which guards against a hand concatenation that replaces `JoinPath`'s cleaning.
- The roster API reads `r.PostForm` only, on 5b-ii's precedent: a token in the query string is never read. The admin-field refusal checks `PostForm`, and the plan states whether an `admin` in the query string is refused or ignored.

## Counts

Blockers 0, majors 3, minors 9.
