# Pass 5b-ii security review (rung 5, identity and auth), fresh context, 2026-09-23

Scope: `git diff 1d4ed29..HEAD -- '*.go' go.mod go.sum config/` at HEAD 08befea. I graded it against the plan's threat-model section, spec lines 1909-1980, and the 5b-i do-not-relitigate list. Every reproduction ran against a `git archive HEAD` copy in the session scratchpad (`.../scratchpad/tree`), using scratch `zz_review*_test.go` files and a built `dps` binary. The repo was not modified; the only writes outside the scratchpad were the Go build cache.

## Verdict

I found no blocker. The 5b-i fixes hold with two exceptions: M4 is only partly fixed, and m10 covers ASCII control characters but not non-ASCII ones.

- **Verified fixed:** B1 (the retention tightening no longer opens a replay window), M1, M2 within one process, M3, M5, m1, m2 on refresh, m3, m4, m5, m6, m7, m8, m9, m12 and m14.
- **The seven-step consumption order, single use under concurrency, the generic answer, the SMTP guard and the origin check are sound.** The generic answer is identical in bytes and in measured timing.

There are five majors:
1. The roster write path only protects writers inside one process. The spec's own CLI verb is a second writer, so members get lost or come back after removal.
2. With the documented public-URL shape, every mailed link lands outside the owned prefix, in the fallthrough.
3. A flood of distinct addresses, even from one client that is already refused, resets any spent per-address bucket.
4. The per-client limit key is a full IPv6 /128.
5. `serve` has no trusted-proxy setting at all. Behind the estate's cloudflared tunnel, every client shares one loopback bucket. This is a 5d deploy gate.

I recommend fixing majors 1-4 before 5c builds the add-member page and the link page on top of them.

## Findings

### 1. MAJOR, CONFIRMED: the roster write path loses members written by another process and brings back members removed by hand
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/roster.go:338-374` (AddMember re-checks and rewrites from the in-memory `r.members`); `:286-301` (Lookup never reloads).

**Attack / failure.** The spec names two writers: the `member` verb (a separate CLI process) and the web add-member page inside `serve`. Hand edits are a third (spec lines 179, 297, risk 13). `r.mu` serializes writers inside one process only.
- When serve adds a member, it rewrites the file from its stale in-memory copy. That silently drops any member the CLI added since serve started.
- It also brings back any member an operator removed by hand to offboard them.
- Serve keeps resolving a removed member until it restarts. Claim 8's "a link for an address removed from the roster is refused at consumption" therefore holds only for removals serve itself makes, and serve has no removal path.

**Repro:** `go test ./internal/auth -run TestReviewRosterCrossProcess -v`. Two `Roster` instances share one file; the CLI instance adds `cli-added`; `leaver` is removed by editing the file; the serve instance then adds `web-added`.
```
file after CLI add + hand removal: {...admin..., cli-added...}
serve still resolves leaver@example.com after removal (no reload)
file after serve's web add:  {...admin..., leaver@example.com..., web-added...}
```

**Fix:**
- Take a cross-process advisory lock (`syscall.Flock` on a sibling lock file) around re-read, validate, write and re-read.
- Re-read the file from disk under that lock before the collision check. Merge from disk, not from memory.
- Have `Lookup` reload when the file's mtime, size or inode changes (a stat per call is cheap at group scale).
- Add a gate test with two `Roster` instances over one file.

### 2. MAJOR, CONFIRMED (routing): mailed links point outside the owned prefix, so the token reaches the fallthrough
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/token.go:189,295-303`. LinkURL joins the fixed `/login/link` onto `DUBPLATE_PUBLIC_URL`. Separately, `cmd/dubplate-server/auth.go:358-380` never ties the URL's path to `DUBPLATE_SERVE_PREFIX`.

**Attack / failure.**
- The spec says link hostnames derive from `DUBPLATE_PUBLIC_URL` (spec line 1033), and the handler fixture uses a bare origin (`https://music.example.com`). Under that shape every link is `https://host/login/link?t=...`.
- With the default prefix `/dubplate`, that path is outside the owned surface. `Prefix.Route` hands it to the fallthrough, whose cookie strip removes the session cookie but not the query.
- At rung 5 that means a 404, so no mailed link reaches a page. At rung 12 the fallthrough is the streamer's reverse proxy. Every magic-link token would then go, in its query string, to a third-party process and its access logs. That defeats R32's "the consumption route logs the path and never RawQuery".

**Repro:** `go test ./internal/serve -run TestReviewLinkLandsOnFallthrough -v`
```
GET /login/link -> status 298; fallthrough saw RawQuery="t=TOKENVALUE"
```

**Fix:** have LinkURL take the owned prefix and build `origin + prefix + /login/link`. Also refuse a `DUBPLATE_PUBLIC_URL` that carries a path, or one whose path is not the prefix. Assert in a shell-level test that a minted link routes to an owned handler.

### 3. MAJOR, CONFIRMED: a flood from one refused client evicts a victim's spent per-address bucket
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/limit.go:85-86` (the per-address key is charged even when the client is refused); `:163-181` (LRU eviction at capacity).

**Attack.**
- The comment's premise is that the least-recently-used bucket has refilled. Under a flood of distinct keys, though, a bucket charged seconds ago becomes the LRU entry.
- Because `perAddress.allow` runs before the client check, and runs even when the client is refused, one client that is already rate-limited can push 10,000 new address keys and evict the victim's bucket.
- The victim then gets a fresh five-per-hour budget, over and over. Combined with finding 4, or any multi-client attacker, per-address mail to a victim rises to the global ceiling (60/h by default, 600 at most). Reaching that ceiling also locks every other member out of sign-in mail.

**Repro:** `go test ./internal/auth -run TestReviewPerAddressEvictionFromRefusedClient -v`. The victim is exhausted at instant T, then flooded from one client at the same T.
```
flood from ONE client: 9980 of 10000 requests refused (per-client), per-address map size 10000
victim request after flood, same instant: err=<nil>
```

**Fix:**
- Charge the per-address key only after the per-client check admits the request.
- At capacity, evict only a bucket that has fully refilled (`Tokens() >= burst`). If no entry qualifies, refuse the new key: fail closed rather than reset.
- Test the "exhausted, then flood, same instant, still refused" case.

### 4. MAJOR, CONFIRMED: per-client keys are full IPv6 addresses, so one /64 counts as 2^64 clients
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/limit.go:86,101` keys on the raw `ClientIP` string from `/var/home/glw907/Projects/dubplate/internal/serve/clientaddr.go:68-114`.

**Attack.** Any IPv6 host holds at least a /64. The per-client link-request limit and the consumption limit disappear, and so does the per-client key that fork 4's kind answer rests on for enumeration.

**Repro:** `go test ./internal/auth -run TestReviewIPv6PerClientKey -v`
```
consumption limit 1/h per client; 100 addresses inside one /64 admitted: 100
```

**Fix:** in the limiter's key derivation (not in `ClientIP`, which also feeds logs), parse with `netip.ParseAddr`. Key IPv6 on its /64 prefix (`Prefix(64)`); key IPv4, and IPv4-mapped addresses unmapped, on /32.

### 5. MAJOR (5d deploy gate), CONFIRMED mechanism: behind the tunnel, every client shares one loopback bucket
**Location:** `/var/home/glw907/Projects/dubplate/cmd/dubplate-server/serve.go:213,252`. `proxyConfig` is hardwired to its zero value, and no threshold can set it.

**Attack.** The estate fronts serve with cloudflared on the same host, so every peer is 127.0.0.1 or ::1.
- One visitor spending 20 link requests per hour (or 30 consumption attempts) locks sign-in out for the whole group.
- The per-client log field names loopback for everyone.

**Repro:** `go test ./internal/auth -run TestReviewLoopbackPeerSharedBucket -v`. It sends 20 attacker probes, then one real member, all from peer 127.0.0.1 with distinct `CF-Connecting-IP` values.
```
real member after attacker's 20 probes, both via loopback peer: 429 {"outcome":"rate_limited"}
```

**Fix (5d):** register trusted-proxy thresholds: a proxy address list and a header, such as `CF-Connecting-IP` for the tunnel. Have doctor print the mode. Refuse to start in direct mode when the bind address is loopback and the public URL is not.

### 6. MINOR, CONFIRMED: in kind mode the global ceiling bounds no enumeration, and a spent ceiling is itself a membership signal
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/handlers.go:192-207`.

**Attack.**
- Unknown-address answers never touch the ceiling, so only the per-client key bounds enumeration, and findings 4 and 5 defeat that key.
- Once the ceiling is spent, a known address answers 429 while an unknown one answers 200, so the oracle survives.
- The ceiling is also a group-wide lockout lever. With 12 known addresses at 5/h, three clients spend 60/h. The kind answer itself hands out the admin's address.

**Repro:** `go test ./internal/auth -run TestReviewKindOracleAfterCeiling -v`
```
known after ceiling:   429 {"outcome":"rate_limited"}
unknown after ceiling: 200 {"outcome":"unknown","admin_contact":{...}}
```

**Fix:** add a global link-request ceiling across all addresses that answers every request 429 once spent. State in the runbook that kind mode's enumeration bound is the per-client key alone.

### 7. MINOR, CONFIRMED: the roster loader accepts files that are not the schema, so a malformed roster starts the server empty
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/roster.go:168-178`.

**Failure.** The loader does not check `schema_version`, does not use `DisallowUnknownFields`, and accepts `null`.
- `null` and `{}` load as empty rosters.
- `schema_version: 99` is read with version 1's meanings.
- A typo such as `"member"` loads an empty roster.
- `encoding/json` matches `"ADMIN": true` case-insensitively, and the last of two duplicate `"admin"` keys wins.

This contradicts claim 7: the server starts in a degraded state rather than refusing.

**Repro:** `go test ./internal/auth -run TestReviewRosterLoose -v` shows all six variants loaded. `null` loads with members=[], and the case variant loads with Admin:true.

The built binary shows the same with `DUBPLATE_ROSTER_FILE=null.json`:
- `dps serve` runs until the timeout (exit 124).
- `dps doctor` prints `magic-link active` and exits 0.

**Fix:** decode with `json.Decoder.DisallowUnknownFields()`. Require `schema_version == 1` and a non-null `members` key. Refuse `null`.

### 8. MINOR, CONFIRMED: non-ASCII control and format characters pass NormalizeAddress (m10 fixed for ASCII only)
**Location:** `/var/home/glw907/Projects/dubplate/internal/identity/identity.go:50`. The check is `r <= ' ' || r == 0x7F`. The display-name check at `roster.go:265-267` uses `unicode.IsControl`, which does not cover the Cf category.

**Attack.** These all normalize as fixed points: U+0085 (NEL), U+009B (CSI, the character roster.go's own comment calls terminal-dangerous), U+202E (right-to-left override), U+00A0 and U+200B.
- They enable lookalike addresses in the admin UI, the audit record and a future `member list`.
- A zero-width variant of an existing address counts as a distinct member.
- Separately, the Kelvin sign in a domain lowercases to a different ASCII domain.

**Repro:** `go test ./internal/auth -run TestReviewNormalizeCorpus -v` shows `"a\u009bb@example.com" -> "a\u009bb@example.com" fixed=true`, and the same for the other characters listed.

**Fix:** refuse any rune with `unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs, unicode.Zl, unicode.Zp)`. Apply the same Cf and Z* refusal to display names.

### 9. MINOR, CONFIRMED (reachable once a trusted proxy is configured): the forwarded hop is not validated as an IP
**Location:** `/var/home/glw907/Projects/dubplate/internal/serve/clientaddr.go:121-135`.

**Attack.** The threat model names a misconfigured proxy as an adversary. One that passes the client's `X-Forwarded-For` through unchanged lets the attacker choose any string, up to `MaxHeaderBytes` (1 MiB), as the limiter key and as the `client_ip` in every auth log line. The limiter's "about 15 MB" bound becomes about 10 GB (10,000 keys at up to 1 MiB each), and each log line grows the same way.

**Repro:** `go test ./internal/serve -run TestReviewForwardedHopUnvalidated -v`
```
ClientIP from a passthrough proxy: 200000 bytes ...
ClientIP: "not an ip; drop table"
```

**Fix:** `netip.ParseAddr` each hop, and treat an unparseable hop as a malformed header, falling back to the peer address.

### 10. MINOR, CONFIRMED: the refusal message echoes a public URL's userinfo password
**Location:** `/var/home/glw907/Projects/dubplate/cmd/dubplate-server/auth.go:361,364,367,370,374,377`. Each message formats `%q` of the whole value.

**Repro:** run `dps serve` with `DUBPLATE_PUBLIC_URL=https://operator:fixture-pass-not-real@dubplate.example.test`. The output line quotes the full value including the password:
```
DUBPLATE_PUBLIC_URL: "https://operator:fixture-pass-not-real@dubplate.example.test" carries userinfo, ...
```
That line reaches stderr and the journal.

**Fix:** never quote the value. Name the defect only, or redact `u.User` before formatting.

### 11. MINOR, CONFIRMED: M4 partly fixed; doctor still names no key for a bad identity-source value
**Location:** `/var/home/glw907/Projects/dubplate/cmd/dubplate-server/doctor.go:268-288`. Source keys are never printed as rows.

The verdict is now right: `magic-link unsatisfied`, and the retention-below-expiry case names its threshold row. But no line names the offending key for:
- a short `DUBPLATE_SESSION_SECRET_PREVIOUS`,
- a short `DUBPLATE_MAGIC_LINK_SECRET`,
- a public URL with userinfo.

**Repro:** `dps doctor` with `DUBPLATE_SESSION_SECRET_PREVIOUS=short` greps to only `magic-link unsatisfied` and `configuration is incomplete`. `dps serve` with the same environment names the key.

**Fix:** print one `<KEY> invalid` or `<KEY> missing` row per failing source key, as M4 asked. Print the key name only, never the value.

### 12. MINOR, CONFIRMED: doctor is green for configurations that break sign-in or mislabel the cookie
**Location:** `/var/home/glw907/Projects/dubplate/cmd/dubplate-server/serve.go:299-311` (port parsed only with Atoi); `/var/home/glw907/Projects/dubplate/cmd/dubplate-server/auth.go:263` (no lifetime ceiling and no lifetime-vs-cap check); `/var/home/glw907/Projects/dubplate/internal/auth/session.go:102-111` (MintSession does not clamp to the cap).

**Failure.**
- `DUBPLATE_SMTP_PORT=0` loads, and every send then fails at runtime. validateSMTPFrom's own comment says this class must be caught at load.
- `SESSION_LIFETIME=87600h` with `ABSOLUTE_CAP=1h` loads. The first cookie then carries a Max-Age of 10 years for a session the server refuses after an hour; m2 is fixed on refresh but not at mint.

**Repro:** `dps doctor` with those values (plus the `null` roster) prints `magic-link active` and exits 0.

**Fix:**
- Bound the port to 1-65535.
- Require lifetime <= cap, or clamp at mint.
- Give nonce retention a sane ceiling.

### 13. MINOR, CONFIRMED: AddMember accepts an empty actor and an invalid channel
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/roster.go:328-383`.

**Failure.** Claim 6 requires the audit record to name the actor and the channel. A caller passing `""` and `Channel(0)` writes the roster and logs `actor="" channel=unknown`.

**Repro:** `go test ./internal/auth -run TestReviewAddMemberAuditFields -v` prints `err=<nil> audit=... actor="" channel=unknown`.

**Fix:** before any write, refuse `!channel.Valid()` and refuse an actor that does not normalize.

### 14. MINOR, PLAUSIBLE: epoch monotonicity holds only within one process
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/epoch.go:416-441`. The read-modify-write is guarded by `e.mu` only.

**Failure.** Bump has no production caller yet. Whichever offboarding surface 5c or 5d adds will run as a separate process from serve. Two near-simultaneous bumps from different processes can store the earlier time. Sessions minted in the gap between them would survive, which is a window of seconds.

**Fix:** use the same flock approach as finding 1 around `advance`.

**Also note:** nothing in 5b-ii, or in 5c's listed deliverables (`member add` and `list`), gives an operator a way to bump an epoch. Offboarding currently means a hand edit plus a restart, and finding 1 makes that hand edit fragile.

### 15. MINOR, PLAUSIBLE: streamer_username uniqueness is case-sensitive and admits LIKE metacharacters
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/roster.go:211-216,246-253,344-349`.

**Failure.** `Alice` and `alice` both load, as do `a%`, `a_` and names with spaces. If rung 12's streamer resolves the principal case-insensitively or with SQL `LIKE` (my recollection of Navidrome's username lookup, which rung 12 must verify), two members can map to one streamer account.

**Fix:** restrict to `[a-z0-9._-]`, compare case-folded, or both.

### 16. MINOR, PLAUSIBLE (residual): under the generic answer, the background send is observable as load
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/handlers.go:177-188`.

**Measured:** 3,000 alternating requests over a real listener give identical bytes, status and headers.
```
known   p10=173µs p50=233µs p90=360µs
unknown p10=181µs p50=239µs p90=346µs
```
The response itself meets R-5b28. The residual is that a known address starts a TLS and SMTP exchange after the response. On a small VPS, a probe running at the same moment could see the extra CPU or connection load. That is noise-dominated, but it is not zero.

**Fix:** none required; record it as a residual. If it matters, queue sends through one worker with jittered start times.

### 17. MINOR, CONFIRMED (nit): the request-link endpoint charges no limiter for an address that fails to normalize
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/handlers.go:160-170`.

**Failure.** Invalid-address posts are unlimited body parses. No mail is sent, so the cost is CPU only.

**Fix:** charge the per-client key before normalizing.

### 18. MINOR, CONFIRMED (nit): the origin check admits an `http://` Origin on the same host when Sec-Fetch-Site is absent
**Location:** `/var/home/glw907/Projects/dubplate/internal/serve/origin.go:273-275`. This is the stdlib's host-only comparison.

**Repro:** `go test ./internal/serve -run TestReviewOriginEdges -v` shows `origin http downgrade same host -> 200`.

Only browsers too old to send Sec-Fetch-Site are exposed, and only to a network attacker. 5c's header set must include HSTS.

### 19. MINOR, CONFIRMED (nit): an unreadable nonce record is never swept
**Location:** `/var/home/glw907/Projects/dubplate/internal/auth/nonce.go:213-221,241-245`. The mtime fallback applies only to unparseable records, not to unreadable ones.

**Repro:** `TestReviewStoreModes` gives `sweep={Removed:0 Malformed:1}`, and the record's nonce still refuses a replay, which is correct. The record stays on disk forever.

**Fix:** stat for mtime on a read failure too.

## Status of each 5b-i finding

| Finding | Status | Evidence |
| --- | --- | --- |
| B1 | Fixed | `TestReviewNonceB1`: 1s retention, 48h token. At +25h: Removed 0 and the replay is refused. The record is retained at expiry and removed at expiry+1s. A sub-second expiry is still retained. |
| M1 | Fixed | A quoted space or comma is refused. Every accepted corpus entry is a fixed point. |
| M2 | Fixed within one process | A backward bump returns `ErrEpochBackward` and the session stays revoked. Across processes, see finding 14. |
| M3 | Fixed | 0 of 64x6 final-character variants verify. CR, LF, a leading space, a trailing `=`, and doubled or trailing separators are all refused. |
| M4 | Partial | See finding 11. |
| M5 | Fixed | 3 allocations for 64 KiB of separators. |
| m1 | Fixed | The rendered header carries `Max-Age=2592000; HttpOnly; Secure; SameSite=Lax; Path=/`. |
| m2 | Fixed on refresh, not at mint | See finding 12. |
| m3 | Fixed | A 0755 leaf is refused. |
| m4 | Fixed | Userinfo, a query and a fragment are refused. See finding 10 for the echo in the message. |
| m5 | Fixed | Link expiry ceiling of 1h, cap ceiling of 2 years, floors of 1s. |
| m6 | Fixed | `Store.Write` is gone. |
| m7 | Fixed | `idRefusalReason` never prints the id. |
| m8 | Fixed | One unreadable record no longer blocks startup, and its nonce still refuses a replay. |
| m9 | Fixed | The clear cookie renders `Max-Age=0` with matching attributes. The handler doc does not claim revocation. |
| m10 | ASCII only | See finding 8. |
| m12 | Fixed | `RecordID` hashes a length-prefixed kind and value. |
| m14 | Fixed | A panicking subscriber leaves `err=<nil>`, and the next subscriber is still called. |
| m11, m13 | Carried to 5d | As the plan states. |

## Attacked and found sound (the do-not-relitigate record for 5c and 5d)

- **Token canonicality and CR/LF refusal.** `base64.RawURLEncoding.Strict()` plus an explicit CR/LF check give one string per credential. SplitN bounds allocation. Both the link and the session share `splitSigned`.
- **Nonce lifetime under any retention and any clock step.** The record's sweep time is `max(consumedAt+retention, tok.ExpiresAt)` in absolute time. A forward step past the record's expiry is also past the token's. A crash between create and write means `Insert` never returned nil, so no session was minted.
- **Single use under concurrency.** 64 concurrent consumptions of one link produce exactly 1 success, and `-race` is clean.
- **The seven-step consumption order.**
  - The limit is charged before the body is parsed.
  - A removed-address link and a revoked link are both refused with byte-identical `{"outcome":"refused"}`.
  - Neither consumes a nonce: exactly one record is on disk after one success and three refusals.
  - A replay is refused and sets no cookie.
  - A token supplied only in the query string is never read, because only PostForm is.
  - Every error denies.
- **What consumption logs carry.** The reason is a sentinel string, with no token, nonce or address. The path is logged and RawQuery never is.
- **The generic unknown-address answer.** Status, body, Content-Type and `no-store` are identical, and measured latency is indistinguishable. The send runs under `context.WithoutCancel`, bounded by go-mail's own timeout.
- **Cookie attributes and sign-out.** The cookie carries the `__Host-` prefix with Secure, HttpOnly, SameSite=Lax, Path=/, no Domain, and Max-Age plus Expires. Sign-out sends a matching clear cookie. Sessions stay stateless, as documented.
- **The SMTP header-injection guard.** CR or LF in the recipient, subject or From is refused before dialing. A comma-joined multi-recipient value is refused by go-mail's parse. U+2028 is not a header break. Send errors print only `SendError.Reason` and the code, never go-mail's "affected recipient(s)" list.
- **Origin check placement.** The chain runs strip, then normalize, then the TRACE/TRACK 405, then the origin check, then routing. Every subtree and the fallthrough sit behind it.
- **Origin check behavior.** Allowed: Sec-Fetch-Site `same-origin` or `none`, and a matching Origin. Refused: a POST with no headers, Sec-Fetch-Site `same-site`, `cross-site` or a bogus value, `Origin: null`, an empty Origin, an empty Sec-Fetch-Site paired with an evil Origin, a lowercase `post`, and PROPFIND.
- **The bearer bypass.** No bypass is registered at rung 5, and a mux redirect never counts as a bypass match. Go 1.27.1 carries the CVE-2025-47910 fix.
- **Rate-limit key derivation.** No key uses a token or cookie string. Keys are the normalized address and ClientIP, and the three key spaces are separate. In direct mode, forged forwarding headers from any peer are ignored.
- **Admin flag.** AddMember refuses `Admin` on every channel, including an invalid one. The collision re-check runs under the lock. A failed validation writes no byte. The seam reads `Admin` from the roster, never from the request.
- **Startup fails closed on each of these:** a missing roster, a leaf wider than 0700, an unreadable nonce directory, a short current or previous secret, a bad or userinfo public URL, an unknown identity source, and an unparseable SMTP port string.
- **Epoch.** It moves forward only within a process. A panicking subscriber is contained. Both refusal checks compare at equality against the anchor or the mint time.
- **Body caps.** `MaxBytesReader` is applied before `ParseForm` on all three endpoints, with the cap bounded to 512 B-64 KiB. Every response is `no-store` JSON.
- **The normalizer's error text.** It quotes its input, but every handler and limiter caller drops it before logging.

## Scratch artifacts

These are the scratch reproduction files; none were written into the repo:
- `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/611daa97-f61c-4f5b-953c-d1bc92fd16c2/scratchpad/tree/internal/auth/zz_review*_test.go` (seven files)
- `.../scratchpad/tree/internal/serve/zz_review_test.go`
- `.../scratchpad/tree/internal/notify/zz_review_test.go`
- `.../scratchpad/tree/cmd/dubplate-server/zz_review_test.go`
- the built binary at `.../scratchpad/dps`

Counts: 0 blockers, 5 majors, 14 minors.
