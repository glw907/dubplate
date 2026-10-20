# Pass 5c-i security review (rung 5, identity and auth), 2026-09-24

**Scope.** `git diff cb9ecc3..HEAD -- internal cmd config` at HEAD 657f9ec. I graded it against the plan's rubric (R-5c3 to R-5c20, the threat-model section, Review focus), spec :1940-2009, and the do-not-relitigate lists from 5b-i and 5b-ii.

**How it was tested.** Every probe ran against a `git archive HEAD` copy in `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/82a7130f-dc63-4c92-b308-9af3d3bb0dde/scratchpad/5c-i-secreview/`, using scratch `zz_review*_test.go` files and one helper binary, `xproc`. `go test -race -count=1 ./internal/auth ./internal/identity ./cmd/dubplate-server` passed on the copy, scratch probes included. The repo tree is unchanged: `git status` shows only the two untracked research files that were there before.

## Verdict

- **No critical and no major findings.**
- **Majors 1-4 from 5b-ii are fixed**, and I reproduced each fix independently. The strongest check was six real OS processes adding 25 members each to one roster at the same time: 151 of 151 members landed and no temp file was left behind.
- **Minors 7, 8, 11, 12, 13, 17 and 19 are fixed.** Minor 15 is fixed as specified but leaves one gap (m1). Minor 10 is only partly fixed (m2).
- **The R-5c18 refusal of minor 6 is sound**, with one sentence to add to the 5d runbook row.

## Findings, ranked by consequence

### m1. MINOR (becomes a rung 12 blocker): a streamer_username can equal another member's address, and the seam then gives both members one principal
**Location.**
- `internal/auth/seam.go:119-122`: the principal falls back to the address when no username is set.
- `internal/auth/roster.go:495-502`: at load, uniqueness compares usernames only with usernames.
- `internal/auth/roster.go:743-747`: the same at write time.

**Defect.** R-5c6 allows `@` and `.` in the write-path pattern, and the loader keeps `streamer_username` optional. Either path therefore accepts a member whose username case-folds to another member's address when that other member has no username.
- The seam then gives both members the same `Principal`.
- R-5c6 cites Navidrome's `FindByUsername` as case-insensitive equality, so at rung 12 both members become one streamer account.
- R-5c6's own rationale ("a case collision is one principal shared by two people whatever path wrote it") is exactly this case, and the roster can see it. It is not the streamer-side mapping the plan carries to rung 12.

**Evidence.** `TestReviewPrincipalCollision` (scratch):
- `AddMember(mallory, streamer_username="Alice@Example.test")` on `ChannelCommandLine` returned `err=<nil>`. Mallory's principal is then `"Alice@Example.test"` and alice's is `"alice@example.test"`.
- A hand-written roster with the same collision loads with `err=<nil>`.

The web add path in 5c-ii will be a second writer that accepts this.

**Fix.** Make uniqueness key on the *effective principal*: the username when set, otherwise the normalized address, compared with `strings.EqualFold`. Apply it in `validateRoster` and in `addTo`. Land it before 5c-ii's roster API, or record it as a hard rung 12 precondition.

### m2. MINOR: minor 10 is only partly fixed; the public-URL refusal still quotes part of a userinfo password
**Location.** `cmd/dubplate-server/auth.go:439-445`. `validatePublicURL` still formats `uerr.Err`, the inner error from `url.Parse`.

**Defect.** When the password contains `/`, `?` or `#`, `url.Parse` cuts the authority at that character. It then reads `user:pass` as `host:port` and fails with `invalid port ":<password prefix>" after host`, quoting everything up to the cut. Base64-style passwords often contain `/`. The line reaches stderr and the journal through `serve`, `check` and `member`. `doctor` is not affected, because it prints key names only.

**Evidence.** `TestReviewPublicURLLeak` (scratch):
```
"https://operator:fixturepass/notreal@dubplate.example.test" -> DUBPLATE_PUBLIC_URL: invalid URL: invalid port ":fixturepass" after host
```
The `?` and `#` variants give the same output. The repo's `TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue` misses this because its unparseable case (`@host:notaport`) quotes only the port.

**Fix.** Drop the inner reason and return a fixed `"invalid URL"`. Alternatively, when the raw value contains `@`, answer `"carries userinfo"` before parsing. Add the `/`, `?` and `#` rows to the test.

### m3. MINOR: the roster file's mode is never checked, and hot reload now applies a hand edit on the next request
**Location.** `internal/auth/roster.go:266-293` (`readRoster`, no mode check) and `:433-444` (`current`, reload on every call).

**Defect.**
- The roster now carries the admin flag, and since this pass a changed file takes effect at the next request with no restart.
- A roster file (or its directory) that is group- or world-writable lets any local user who can write it make themselves admin. No audit record is written.
- The store writes 0600, but a hand-created or restored file keeps whatever mode it was given.
- The nonce store already refuses a leaf wider than 0700 (`internal/auth/store.go:70`, the 5b-i m3 precedent). The roster, now the more privileged store, has no equivalent check.
- The spec keeps local users out of the threat model, so this is defense in depth and not an exploit path.

**Fix.** In `readRoster`, refuse a file whose `Mode().Perm()&0o022 != 0`, reusing the descriptor it already `fstat`s. Name the path, not the content. Optionally, `OpenRoster` can also refuse a group- or world-writable directory. `doctor` then shows this through its existing roster row.

### m4. MINOR: a write that lands but reports an error writes no audit record
**Location.** `internal/auth/roster.go:408-411` (`writeRoster`, then the re-read `readRoster`); `:698-705` and `:250-256` (the audit record is logged only on a nil return); `internal/atomicfile/atomicfile.go:78-89` (rename, then directory fsync).

**Defect.** Suppose the rename succeeds, but then the directory fsync or the post-write `readRoster` fails (EIO, or a transient read error).
- `AddMember` or `CreateRoster` returns an error and logs nothing, yet the file on disk now holds the new member.
- That breaks the threat-model claim "every roster change that succeeds writes one audit record naming the actor and the channel".
- The operator sees a failure. A retry then reports a duplicate address, so nothing on the terminal says that a member was added.

**Fix.** Split the result of `writeLocked`: once `writeRoster` has passed its rename, the change has happened. Log the audit record from that point, or log a warning-level record with the same fields ("roster write landed, re-read failed"), whatever follows. The narrowest version: have `writeRoster` report whether the rename happened, and let the caller log on that.

### m5. MINOR (nit): the unreadable-record sweep reads the current retention, not the retention stored in the record
**Location.** `internal/auth/nonce.go:265-270` (`mtimeFallback`), now also reached for records that cannot be read.

**Defect.** Minor 19's fix sweeps an unreadable record at `mtime + current retention`. A readable record carries its own `max(consumedAt+retention, tok.ExpiresAt)` (the B1 property).
- **Replay path.** Suppose an operator lowers both `DUBPLATE_NONCE_RETENTION` and the link expiry below a link still outstanding under the old config (at most the 1h ceiling). An unreadable record for that link can then be swept before the token expires, and the link replays.
- **Required conditions.** The replay needs a retention below 1h (the default is 24h), a config change, and a record that is unreadable or unparseable.
- **Same exposure before.** Unparseable records already had this behavior.

**Fix.** Use `mtime + max(retention, linkExpiryCeiling)` as the fallback, with the 1h ceiling passed into `OpenNonceStore`. That is cheap and closes the case by construction.

### Nits (polish-level, listed for coverage)
- **n1. Unmetered bad bodies.** `internal/auth/handlers.go:192-195`: request-link charges the per-client key after `readForm`, so oversized or unparseable bodies are never charged. Consumption charges before parsing (`:337-344`). The cost is CPU only, bounded by BodyCap. Moving the charge above `readForm` would make the two endpoints consistent.
- **n2. Empty prefix.** `internal/auth/token.go:164` and `handlers.go:108`: `HandlersConfig.Prefix` is a plain string that neither `NewHandlers` nor `LinkURL` validates. A future caller (5c-ii's rewiring) that leaves it empty silently mints `/login/link`, which is major 2 again. `authHandlersConfig`'s test pins today's wiring. A one-line refusal of an empty or `"/"` prefix in `LinkURL` would stop a regression.
- **n3. Empty display name.** `internal/auth/roster.go:527`: an empty or all-space display name passes the write path. `member add --admin --name ""` bootstraps an admin with `display_name:""` (scratch `TestReviewEmptyName`), and the kind answer then hands strangers an admin contact with no name. This is the same "flag present but empty" class as Review focus 5.
- **n4. Actor not normalized in the audit record.** `internal/auth/roster.go:756-762`: the record logs the actor as supplied. `"B@Example.test"` passes the re-check (it normalizes) but is logged un-normalized. Log the normalized form on the cookie channel.
- **n5. Invisible characters outside Cf and Z\*.** `internal/identity/identity.go:64-72`: invisible code points in other categories still pass the normalizer: U+3164 and U+115F (Hangul fillers, Lo), U+034F (CGJ, Mn), U+FE0F (VS16, Mn) and U+2800 (braille blank, So). An address carrying one looks identical to one without it and still normalizes to itself (scratch `TestReviewNormalizeResidual`). Only a trusted admin can add such an address. Refusing Unicode Default_Ignorable_Code_Point plus these fillers would close it.

## Verify-fixed table (5b-ii findings this pass folds)

| 5b-ii finding | Status | Evidence |
| --- | --- | --- |
| Major 1: roster across processes, hand-removal resurrection, no reload | **Fixed** | `roster.go:345-367` (directory `flock`, polled under ctx), `:379-412` (re-read and validate from disk under the lock, merge from disk), `:433-461` (reload per call, `fstat` on the read descriptor, 1s mtime trust window). Tests: `TestRoster_CrossProcessWrites`, `TestRoster_ReloadsChangedFile`, `TestRoster_ModTimeTrustWindow`, `TestRoster_ConcurrentWritersAcrossValues`, `TestAddMember_BusyWhileLockHeld`. Probe: 6 real OS processes × 25 `AddMember` + `Lookup` each gave 151 of 151 members, no temp files; `-race` clean. |
| Major 2: link outside the owned prefix | **Fixed** | `token.go:164-172` (`JoinPath(prefix, linkPath)`), `auth.go:455-457` (public URL carries no path), `serve.go` `authHandlersConfig` passes `prefix.String()`. Test: `TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree` (both origin forms reach the `login` subtree, and the fallthrough never runs). Probe: `LinkURL` table over 4 origins × 6 prefixes. No shell code logs `RawQuery`. Residual: n2. |
| Major 3: a refused client's flood evicts a victim's spent per-address bucket | **Fixed (R-5c3 design)** | `handlers.go:195` (per-client first), `:223-233` (per-address charged only for a known address), `limit.go:160-168,240-252` (`evictRefilledOrAdmit`). Tests: `TestRequestLink_UnknownFloodLeavesSpentAddress` (10,000 addresses, map size unchanged, victim still 429), `TestRequestLink_ChargingOrder`, `TestKeyedLimiterRefilledPolicy`, `TestRequestLink_GenericOverLimitAnswersAsUnknown`. |
| Major 4: per-client key is a full IPv6 /128 | **Fixed** | `limit.go:130-143`. Tests: `TestClientKey`, `TestClientLimitsKeyIPv6OnItsSlash64`. Probe: `::ffff:1.2.3.4`→`1.2.3.4`; two hosts in one /64 → `2001:db8:1:2::/64`; the zone is dropped; `""` and garbage → one shared `unparsed` bucket. |
| Minor 7: loose roster loader | **Fixed** | `roster.go:301-317` (json/v2, `RejectUnknownMembers`, schema_version==1, non-nil members). Test: `TestOpenRoster_StrictLoader`. Probe: `null`, `{}`, version 99, `"member"`, `"ADMIN"`, duplicate `admin`, `members:null`, trailing data, `1.0`, `Schema_Version`, a BOM prefix: all refused. |
| Minor 8: non-ASCII control and format characters pass | **Fixed (Cf/Z\*/Cc)** | `identity.go:64-72`, `roster.go:609-617` (display name admits only U+0020, ZWJ and ZWNJ). Tests: `TestNormalizeAddress`, `TestDisplayNameCarriesRefusedRune`, `TestAddMember_ZeroWidthAddressVariantRefused`. Probe: U+200B, U+202E, U+00A0, U+0085 and U+180E refused. Residual: n5. |
| Minor 10: public URL password echoed | **Partial** | The whole value is never quoted (`auth.go:446-464`; `TestValidateMagicLinkSource_PublicURLNeverEchoesTheValue`), but the inner error from `url.Parse` still quotes a password prefix. See m2. |
| Minor 11: doctor names no failing key | **Fixed** | `doctor.go` `identitySourceStatus` walks the join and reads `sourceKeyError` through `errors.AsType`; `auth.go:246-253`. Tests: `TestDoctorCmd_SourceLineNamesFailingKey` (short previous secret, short primary secret, userinfo URL, expiry over the ceiling), `TestDoctorCmd_SourceLineNamesTwoFailingKeysAtOnce`, `TestIdentitySourceStatus_KeyFromKeyedTypeNotMessagePrefix`. Keys only, never values. |
| Minor 12: doctor green on a bad port or lifetime > cap | **Fixed** | `auth.go:344-353` (port 1-65535), `:314-319` (lifetime > cap refused, both keys named), `:105` (retention ceiling 240h), `session.go` `clampExpiry` at mint. Tests: `TestValidateMagicLinkSource_SMTPPort`, `TestValidateMagicLinkSource_SessionLifetimeAboveCap`, `TestDoctorCmd_SessionLifetimeAboveCapReportsUnsatisfied`, `TestMintSessionClampsAtTheAbsoluteCap`. |
| Minor 13: empty actor and invalid channel | **Fixed** | `roster.go:570-585,711-723`. Test: `TestAddMember_ActorAndChannelRefusals`. Probe: `Channel(0)`, `Channel(7)`, an empty CLI actor and a zero-width cookie actor all refused before the lock. |
| Minor 15: streamer_username case and shape | **Fixed as specified** | `roster.go:51` (pattern), `:495-502` (load, case-folded), `:743-747` (write, `EqualFold`). Tests: `TestAddMember_StreamerUsernameCaseFoldDuplicate`, `TestOpenRoster_StreamerUsernameCaseFoldDuplicate`, `TestValidateWriteCandidate_StreamerUsernameShape`/`Required`. Residual: m1 (collision across the username and address namespaces). |
| Minor 17: no charge for an address that fails to normalize | **Fixed** | `handlers.go:195-207` (per-client charged before normalizing). Test: `TestRequestLink_ChargingOrder`. Residual: n1 (bodies that fail to parse are still uncharged). |
| Minor 19: unreadable nonce record never swept | **Fixed** | `nonce.go:245-271`. Test: `TestSweepRetainsUnreadableRecord` (retained while inside the window, replay refused through `O_EXCL`, removed after mtime+retention, startup unaffected). Residual: m5. |
| Minor 6: kind-mode post-ceiling oracle (refused, R-5c18) | **Refusal sound** | The review's fix (a global request ceiling answering 429 to everyone) is a cheaper group-wide lockout than the oracle it closes, and fork 4 already discloses membership. The refusal covers only the oracle half, though. Minor 6 also named the mail ceiling as a **group-wide lockout lever**: with 12 or more known addresses at 5/h each, 60/h empties the ceiling from three /64s. Add that to the 5d runbook row beside "kind mode's enumeration bound is the per-client key alone". A group of 11 or fewer members cannot reach the default ceiling through per-address charges. |

## Attacked and found sound (the do-not-relitigate record for 5c-ii and 5d)

**Roster locking and reload**
- **Cross-process lock design.** The `flock` sits on the directory inode, so it holds under any spelling of the path. Each `Roster` opens its own descriptor, so two values in one process exclude each other as two processes do. The writer takes the `flock` before `r.mu` and holds `r.mu` only for the swap, so `Lookup` and `List` never wait on another process's lock. If a swap lands out of order, the older snapshot fails `os.SameFile` against the file renamed in later, and the next call reloads. The 30s `memberVerbWriteWait` ends with `errRosterBusy`.
- **Stat-and-read order.** `fstat` runs on the same descriptor the content is read from, and `readAt` is taken before the open. I checked the 1s trust window against 1s filesystem timestamps and a coarse kernel clock: an edit after `readAt` always moves mtime past the cached value. Inode reuse by atomicfile's temp files cannot fake a match, because mtime moves forward. A future mtime only forces a re-read, which costs time and is not a security gap.
- **Fail-closed reload.** A broken file drops the cache.
  - Request-link: under the kind answer it answers 500 `failed`, never `unknown`, and `adminContact` is never read from the failing store. Under the generic answer the bytes stay `requested`.
  - Consumption and the seam both refuse.
  - The writer re-reads and validates under the lock, so it never overwrites a broken file.

**Loader, bootstrap and write authorization**
- **Strict loader error text.** json/v2 errors quote member *names* and JSON pointers, never values: type mismatch, trailing data, invalid UTF-8 and a duplicate `address` were all probed. The only way an address appears is when it is used as a JSON member name, which is contrived.
- **Validated single-write bootstrap.** `CreateRoster` runs `checkAddition` (every write-path rule plus the command-line actor) and requires an admin through the `writeLocked` guard. It re-checks absence with a stat under the lock and returns `ErrRosterExists`. Two concurrent bootstraps both land (`TestCreateRoster_ConcurrentCreate`, `TestMemberAddCmd_ConcurrentBootstrapBothLand`). An existing file that is malformed or empty is never recreated.
- **Admin grant and actor re-check.** The flag is admitted only on `ChannelCommandLine`, and that is checked before the lock. A cookie actor's admin flag is re-read from the disk copy under the lock: an in-place hand demotion refuses the write, and a case-variant actor normalizes and passes. `Identity.Valid` refuses `ChannelCommandLine`, and the seam emits only `ChannelCookie` (`TestMagicLinkSeam_VerifyIdentity_NeverCommandLine`). Channel ints are never persisted, so renumbering after dropping `ChannelHeader` is safe.
- **Last-admin guard.** It lives in the one `writeLocked` path every mutation takes, bootstrap included, with no exemption. No demote or remove method is exported.

**The `member` verb**
- It loads through `config.Load` and refuses when the selected source requires no roster. It refuses a foreign directory owner, comparing the effective uid.
- The actor comes from `user.LookupId(euid)` with a `uid:<n>` fallback, never `$USER`.
- The audit record goes to stderr with `client_ip=local`.
- `member list` cannot emit terminal control sequences: the loader refuses Cc/Cf/Z\* in addresses and display names (Cf apart from ZWJ and ZWNJ), and usernames must be printable ASCII.

**Limiter**
- The per-client key is charged first. The per-address key is charged only for roster addresses, so the map is bounded by the roster.
- LRU eviction on the two client-keyed maps hands a flooder nothing.
- `evictRefilledOrAdmit` never resets an unrefilled bucket, and it is unreachable below 10,000 members.
- A targeted per-address lockout is soft: each request the attacker spends mails a working link to the victim.

**Generic answer timing.** The only new difference between known and unknown is one mutex and map operation on the known path. That path already started a goroutine that the unknown path does not, and 5b-ii measured the pair as indistinguishable.

**Link, session and mail copy**
- **Link inside the prefix.** `JoinPath` cleans both `https://h` and `https://h/`. `serve.NewPrefix` forbids a root prefix, a trailing slash, percent-encoding and dot segments. Nothing in `internal/serve` or `cmd/dubplate-server` logs `RawQuery`.
- **Session clamp.** Load refuses a lifetime above the cap, and `MintSession` clamps at mint. A zero cap from a mis-wired caller fails closed (instant expiry).
- **Link expiry floor and mail copy.** The floor is 1 minute and the ceiling 1 hour. The copy rounds up and only then applies the singular form, so it never understates the expiry.

**Normalization.** The Kelvin sign and the dotted capital I normalize to an ASCII roster address, which is harmless: mail always goes to the roster's stored `member.Address`, never to the typed string.

**Error text.** Secrets are never echoed. Values of thresholds, the SMTP port and `unknown_address_answer` are printed, but none is a secret. Roster errors name the path and the record position, never an address, apart from the audit record and `member list`, which R-5c14 and R-5c15 allow.

**Exported surface (R-5c20).** The new exports are `CreateRoster`, `ErrRosterExists` and `identity.ChannelCommandLine`, and each has a caller in `cmd/dubplate-server/member.go`. `List` and `AddMember` changed signature and their callers are in cmd. `rosterAPIWriteWait` and the roster API sentinels have not been built ahead of 5c-ii.

## Scratch artifacts (outside the repo)

All in `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/82a7130f-dc63-4c92-b308-9af3d3bb0dde/scratchpad/5c-i-secreview/`:
- `cmd/dubplate-server/zz_review_test.go` (m2)
- `internal/auth/zz_review_principal_test.go` (m1)
- `internal/auth/zz_review_json_test.go`
- `internal/auth/zz_review_loose_test.go`
- `internal/auth/zz_review_link_test.go`
- `internal/auth/zz_review_actor_test.go`
- `internal/auth/zz_review_empty_test.go` (n3)
- `internal/identity/zz_review_test.go` (n5)
- `internal/auth/zzxproc/main.go`, built as `xproc` (the cross-process probe)

Counts: 0 critical, 0 major, 5 minor, 5 nits.

**Verdict:** 5c-i fixes majors 1-4 and the routed minors, one of them only partly. Five minors remain and none blocks 5c-ii. Fix m1 (username-to-address principal collision) and m2 (partial password in the URL error) before 5c-ii opens the web write path; m3 and m4 are cheap enough to fold with them.

## Conductor dispositions (2026-09-24)

Folded: m1, m2, m3, m4, m5, n1, n2, n3, and n4, through tasks S1 and S2 of the security fold.

Refused: n5. Only an admin can add such an address, and refusing it needs the derived Default_Ignorable_Code_Point table, which the standard `unicode` package does not provide. It is noted for 5c-ii's web write path.

Carried to 5d: minor 6's mail-ceiling lockout-lever sentence joins the runbook row.
