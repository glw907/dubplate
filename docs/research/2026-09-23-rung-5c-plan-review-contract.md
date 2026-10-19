# Rung 5c plan review, lens: contract and criteria (fresh context, 2026-09-23)

Target: `docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md` at `dba1679` (5c-i in full, 5c-ii outlined). I read it against these sources:

- the rung 5 spec's 5c row (`spec:2331`) and its errata;
- the 5b-ii security review;
- `docs/STATUS.md` and the 5b-ii close drafts (`draft-HISTORY-5b-ii.md` "The 5c handoff" and "Also carried to 5c");
- section 15 of the web surface standard;
- the tree at `dba1679` (`roster.go`, `seam.go`, `handlers.go`, `limit.go`, `identity.go`, `doctor.go`, `auth.go`, `serve.go`, `prefix.go`, `atomicfile.go`).

## Verdict

The plan is sound in structure and mostly sound in its criteria. The split is well argued, and both halves can be graded. Almost every criterion names its fixture and the mutation that would break it. STATUS's four carried items and 5b-ii's handoff each land in exactly one place. Six defects would cost a fix round or ship a wrong property:

- The last-admin invariant as written refuses `CreateRoster`'s own empty write, so bootstrap cannot work (fold 6).
- Fold 9 applies fail-closed eviction to every keyed map. That hands anyone holding one free /48 a group-wide sign-in lockout, and the plan's claim that the global ceiling bounds it is wrong.
- Review minor 6 was routed to 5c's security fold, and it lands nowhere: no task, no refusal, no record.
- The 404-uniformity criterion compares refusal rows only with each other. It never compares them with an unregistered path under the prefix, which is the comparison risk 14 is about, and it checks the body but not the headers.
- Task 7's uid row cannot fail, because `/` refuses the write with EACCES whether the guard exists or not.
- The privileged-fields table misses that `streamer_username` becomes a privilege channel at rung 12.

None of these is a blocker. Each fold is small. Eight of the ten conductor-owed folds are sound as applied, and folds 6 and 9 are defects.

## Findings

### 1. MAJOR: the last-admin invariant refuses `CreateRoster`'s empty write, so bootstrap is impossible as specified

**Location:** plan:807-809 (Task 1's Produces entry), :514-536 (R-5c8), :1358-1361 and :1370 (Task 6), :1466-1471 (Task 7).

**Defect:** Task 1 routes both `AddMember` and `CreateRoster` through "one unexported locked read-modify-write function ... which Task 6's guard lives in". Task 6 then states that "every roster the store writes names at least one admin", checked inside that function. `CreateRoster` writes `{"schema_version":1,"members":[]}` (`roster.go:148-160`), which names no admin. Taken literally, the guard refuses the create. The result is that Task 6's row 1 ("empty (fresh `CreateRoster`)") and Task 7's first row ("created at 0600; one admin") both become unreachable. An implementer will either special-case the create silently or escalate. That is a planning miss either way.

**Evidence:** `CreateRoster` at `internal/auth/roster.go:141-162`. R-5c8's own text says "A fresh roster from `CreateRoster` has no members".

**Proposed fold:** make the bootstrap one write. `CreateRoster` takes the first member (with `Admin` required, channel `ChannelCommandLine`, actor, and logger) and writes it under the directory lock in one create. The invariant then holds for every write with no exemption. An empty roster becomes reachable only by hand edit, and the guard still covers that state. Task 7's absent-roster pre-check and its "interrupted bootstrap" row collapse into the store's own refusal. The exported signature changes, but `CreateRoster` has no production caller yet: Task 7 is its first. If the conductor prefers to keep today's signature, the alternative is an explicit exemption for the create in R-5c8 and Task 6, stated and tested.

### 2. MAJOR: fold 9's fail-closed eviction reaches every keyed map, which turns the client-keyed maps into a cheap group-wide lockout (possible OWNER FORK)

**Location:** plan:411-427 (R-5c3), :1072-1081 (Task 3's outcome), :1101-1102.

**Defect:**

- **The rule reaches all three maps.** Task 3 states the rule for "a keyed map", which covers the per-address, per-client, and consumption maps alike (`limit.go:62-67`). All three share `newKeyedLimiter`.
- **What an attacker gets.** With /64 keying, 10,000 consumption POSTs from 10,000 /64s fill the consumption map. That refuses every member's link consumption for up to an interval, and the attack repeats hourly. The per-client link-request map fails the same way.
- **The cost of the attack.** A /48 holds 65,536 /64s, and tunnel brokers hand out /48s free.
- **The bound the plan claims does not exist.** R-5c3 says "both outcomes are bounded by the global link-mail ceiling". The lockout sends no mail, so the ceiling does not bound it.
- **Evicting a client key costs nothing.** An attacker gains nothing from evicting a client-keyed bucket, because it already holds every /64 whose bucket it could reset. Finding 3's harm, the reset of a victim's protection, exists only in the per-address map.

**Evidence:** the review's own fix text is scoped to "a victim's spent per-address bucket" (review finding 3). `NewLimiter` builds all three maps from one constructor.

**Proposed fold (options; recommend A):**

- **A (recommended):** scope the refilled-only eviction to the per-address map. At capacity with no refilled bucket, admit the new key untracked: never evict an unrefilled bucket, and still charge the per-client key and, for a known address, the global mail ceiling. Keep LRU eviction on the two client-keyed maps. The victim's spent bucket survives, which satisfies finding 3's reproduction and exit gate 8, and no lockout lever is created. The distinct-clients test at plan:1119-1122 becomes "a new key is admitted untracked and no unrefilled bucket is evicted".
- **B:** the plan as written (fail closed everywhere), with R-5c3's bound claim corrected and the consumption-map lockout named as a residual.
- **C:** fail closed on the per-address map only. That is still a lockout: about 500 /64s at 20 requests each fill it.

A and C are conductor-rulable within "Geoff ruled the finding into 5c with its fix". B knowingly ships a group-wide availability lever. If the conductor leans toward B, put it to Geoff.

### 3. MAJOR: review minor 6 is routed to 5c's security fold and lands in no task

**Location:** plan:144-146 (the finding list omits 6), :739-756 (Task 0 item 3 lists thirteen, 6 absent), :423-424.

**Defect:** the 5b-ii close's proposed routing, which the plan names as binding (plan:168-171), puts minor 6 (the kind-mode global ceiling bounds no enumeration, and a spent ceiling is a membership oracle) into "5c security-fold Task 1". The draft calls minor 6 out by name as "proposed here for the first time" (`draft-HISTORY-5b-ii.md`, disposition table, and the 5c handoff paragraph at :612-614). The plan cites finding 6 only as an existing residual (R-5c3) and neither folds nor refuses it. Task 0 would not catch the gap, because its list omits 6 too. Minor 18 is correctly routed to 5c-ii.

**Evidence:** `draft-HISTORY-5b-ii.md`, the row "Minor 6 ... 5c security-fold Task 1", and the handoff's list "minors 6, 7, 8, 10, 11, 12, 13, 15, 17, 19".

**Proposed fold:** an eleventh conductor-owed fold that disposes of minor 6 explicitly. Kind mode answers `unknown` against `sent` by design, so the post-ceiling oracle adds little. The review's fix, a global request ceiling that answers every request 429, is itself a cheaper group-wide lockout than the one it closes. Recommend refusing the code fix with that one-line reason, and carrying the review's second half to 5d's runbook ("kind mode's enumeration bound is the per-client key alone"). Either way, add 6 to Task 0 item 3's list, or record it there as refused, so the scope ledger is complete.

### 4. MAJOR: the 404-uniformity criteria never compare against an unregistered path, and they check bodies but not headers

**Location:** plan:577-585 (R-5c12), :1594-1600 (Task 8's gate criterion), :1673-1674 (Task 9), :1761-1762 (exit gate 1), :1940-1942 (5c-ii item 5).

**Defect:**

- **The wrong comparison.** Risk 14 and `spec:1418-1420` require that the route "answers 404 rather than 403, so the page's existence is not a fact a non-admin learns". Task 8 asserts only "404 and a byte-identical body across rows", meaning refusal rows against each other. A gate that answers `{"outcome":"not_found"}` as `no-store` JSON passes every row, while an unregistered path under the prefix answers `http.NotFound`'s `text/plain` "404 page not found" (`internal/serve/prefix.go:133`). A non-admin can then tell that `<prefix>/roster` exists.
- **Headers go unchecked.** The criterion checks the body only. A `Set-Cookie` refresh on a non-admin refusal, or a different `Content-Type` or `Cache-Control`, leaks the same fact.
- **A refusal row is missing.** R-5c12 lists "a roster store failure" as a refusal, and Task 8's refusal table omits it.
- **5c-ii inherits the gap.** 5c-ii reuses `RequireAdmin` for the HTML document route, where W3.7's 404 split applies. The plan does not say whose 404 the gate emits there.

**Evidence:** `spec:1418-1420`, risk 14 at `spec:2595-2601`, and `prefix.go:123-133` (the terminator uses `http.NotFound`).

**Proposed fold:** Task 9, which runs through the real shell, asserts that a non-admin's GET and POST to `<prefix>/roster` match an unregistered `<prefix>/<unknown>` in status, body, and every header except `Date`, with no `Set-Cookie`. Task 8 adds the store-failure row and asserts no `Set-Cookie` on any refusal. State in `RequireAdmin`'s contract that its refusal is the shell's not-found answer. Carry to the 5c-ii outline that the refusal must track the served 404 document once W3.7's 404 lands.

### 5. MAJOR: Task 7's uid row cannot fail, and named mutation 3's claimed catch is false

**Location:** plan:1490-1493, :1508.

**Defect:** the fixture points the roster at a directory the test's uid does not own, "such as `/` when not running as root", and asserts that nothing was created. With the uid check deleted, the verb still creates nothing, because `CreateRoster` against `/` fails with EACCES. The assertion passes either way. Mutation 3 claims "Drop the uid check → the uid row creates a file", which cannot happen. This is the "fixture state so the test cannot pass vacuously" planning-miss item.

**Evidence:** the review of R-5c14 at plan:622-628. A non-root process cannot create a file in `/`.

**Proposed fold:** assert the refusal itself. The error names both numeric ids and the directory, it wraps a named sentinel, and it arrives before any open. Mutation 3's catch becomes the absent refusal message: the error is EACCES from the create instead of the uid refusal. Optionally add a positive row asserting that a temp directory the test owns passes the guard.

### 6. MAJOR: the privileged-fields table misses that `streamer_username` becomes a privilege channel at rung 12

**Location:** plan:372-383 (the table's `streamer_username` row).

**Defect:** the seam's `Principal` is `StreamerUsername` when set (`seam.go:125-128`). Rung 12 injects that principal into the streamer's proxy header (`spec:1243-1245`), and rung 6's spec quotes Navidrome as auto-creating an unknown proxied user (rung 6 spec, "Identity in the store"). So whoever sets `streamer_username` decides which streamer account a member becomes. A cookie-channel admin who adds a member with the handle of the streamer's own admin account grants streamer admin through the web. Fork 1 and fold 8 exist to prevent exactly that. The table lists the field's writers but not this consequence, and nothing carries it forward.

**Evidence:** `seam.go:125-128`, `spec:1243-1245`, rung 6 spec :805-812.

**Proposed fold:** add the consequence to the table's `streamer_username` row. Carry a constraint to rung 12 beside the `LIKE` question, and add it to Task 10's declared gaps: a principal must not map a non-admin member onto a streamer admin account, or rung 12 must verify the mapping. No 5c-i code change.

### 7. MINOR: Task 7's state table omits the malformed, missing-directory, create-race, and user-lookup states

**Location:** plan:1466-1476.

**Defect:** the verb is the bootstrap path and the second writer, yet its table has no row for any of these states:

- the roster present but unparseable or invalid, for both `add` and `list`;
- the roster's directory absent (the uid check's `os.Stat(dir)` fails);
- a concurrent `member add --admin` pair on an absent roster, where the loser's `CreateRoster` returns `ErrRosterExists` (does the verb reopen and add, or exit nonzero?);
- `os/user.Current` failing (what is the actor then?).

**Proposed fold:** add the four rows with their reports. Finding 1's single-write bootstrap resolves the race row.

### 8. MINOR: Task 2's Kelvin-sign criterion passes whatever the outcome

**Location:** plan:994-998.

**Defect:** "The implementer derives [the expected outcome] ... if the normalizer accepts it, the task reports that as a finding". No outcome can fail this criterion.

**Proposed fold:** pin the outcome now. `strings.ToLower` maps U+212A to `k`, and the stored and mailed form is the ASCII owner's address, so accepting it is harmless. Assert acceptance with the ASCII result, citing the `unicode` case tables. The alternative is to refuse any address whose lowercased form differs from a case-mapped ASCII form. Rule one.

### 9. MINOR: two exports have no caller outside tests or their own package

**Location:** plan:1549-1551 (the context accessor), :957-958 (`ErrInvalidAuditActor`).

**Defect:** the verified-identity accessor is "exported" while its only reader, the roster API, sits in the same package. `ErrInvalidAuditActor` has no caller that branches on it: the API maps an unreachable actor failure to `failed`, and the verb prints the message. Both break Geoff's 2026-09-12 refinement ("nothing is exported ahead of a caller") and would trip the staged exported-unused leg.

**Proposed fold:** keep the accessor unexported until a cross-package caller exists. Either unexport `ErrInvalidAuditActor` or name the caller that branches on it.

### 10. MINOR: Task 3's distinct-clients test does not name its unit, and through `AllowLinkRequest` it tests the wrong map

**Location:** plan:1119-1122.

**Defect:** filling the per-address map "from distinct clients" through `AllowLinkRequest` also fills the per-client map to capacity. The next request is then refused by the per-client map first (charged first, by Task 3's own order), and the per-address refusal is never exercised.

**Proposed fold:** name the unit, `keyedLimiter.allow` directly or the per-address map in isolation, and say which map's size is asserted.

### 11. MINOR: Task 6's "line count unchanged" contradicts a row that prints only on failure

**Location:** plan:1390-1392.

**Defect:** the roster row prints only when the roster fails (`doctor.go:83-97`). Today an empty roster opens cleanly and prints no row. After Task 6 it prints one, so the count rises by one relative to today in that state.

**Proposed fold:** "the registry-derived line-count test passes unmodified; a no-admin roster prints the roster row exactly once, as any roster failure does".

### 12. MINOR: R-5c10 does not say how `doctor` gets key names out of `Validate`

**Location:** plan:550-562, :1289-1293.

**Defect:** `validateMagicLinkSource` returns an `errors.Join` of messages that each begin with a key (`auth.go:234-283`). Relational refusals, such as retention against expiry and the new lifetime-against-cap check, name threshold keys that already have their own rows. The plan does not say whether `doctor` parses message prefixes, which breaks silently when a message is reworded, or uses a keyed error type. It also does not say whether threshold keys appear on the source's line.

**Proposed fold:** a keyed error type defined in `cmd/dubplate-server`, so `internal/config` stays untouched, and one sentence on relational refusals.

### 13. MINOR: fold 3's errata list is incomplete

**Location:** plan:1730-1736.

**Defect:** besides R3, `spec:1417` gives the form "an optional streamer username", and R1 (`spec:291-292`) and "The roster" (`spec:1106-1107`) call the field optional on the write path. The close drafts only R3's erratum.

**Proposed fold:** name all three passages in Task 10's corrections.

### 14. MINOR: gaps in the 5c-ii outline against section 15 and the handoff

**Location:** plan:1907-1953, :1659-1661.

**Defect:** the outline misses several items its sources bind:

- Section 15 also binds W1.2 and W1.4's register, W3.6, and W4.4 and W4.6's stacking, and the outline omits all of them.
- The handoff's "logout copy's bound" (a stateless session survives sign-out) has no outline item.
- Task 9 says the add-member page's route is one "its outline names", and the outline names none.
- The spec row's "no log record emitted on the consumption route containing the token" is assigned to neither half. Exit gate 12 claims "token", but no 5c-i criterion proves it.

**Proposed fold:** add the four to the outline, name the page route, and assign the token-log assertion, most cheaply as a re-run of 5b-ii's test cited in 5c-i's exit gate.

### 15. MINOR (OWNER FORK flag): the rung reservation more than doubles against a ratified statement that 5c's ceiling is unchanged

**Location:** plan:232-279.

**Defect:** the arithmetic checks out (7.04M, 9.4M, 35.3M against 15.75M). But section 15 of the web surface standard, ratified by Geoff, says 5c's "4.5M ceiling is unchanged by this standard". 5c is now 16.5M across two passes, and the rung sits 19.55M over its reservation. The plan records this as a conductor fold "against the reservation".

**Proposed fold:** state the rung overrun in the plan header, where Geoff reads it at the approval gate. Add the section 15 figure to Task 10's corrections. Options: A (recommended), surface it at plan approval with no separate question; B, rule it as a conductor fold as drafted.

### 16. MINOR: "two writer processes" is proven with two `Roster` values in one process

**Location:** plan:1772-1774 (exit gate 6), :875-881.

**Defect:** the in-process proof rests correctly on the quoted `flock(2)` text about independent open file descriptions. The exit gate still claims processes.

**Proposed fold:** reword the gate to "two independent store values, which `flock(2)` treats as two processes (quoted)". The other option is to add one re-exec subprocess test.

### 17. MINOR: R-5c14 assumes the service environment exists in an operator's shell

**Location:** plan:614-620.

**Defect:** the verb runs `config.Load`, which requires every owned wall row, the source's secrets and SMTP keys included. On a box, those live in the unit's environment file, not in an operator's shell. The uid guard also requires running as the service user. The plan cites the runbook's ordering but never states how a stranger's `member add` receives the environment. Ease of setup by a stranger is a primary product goal.

**Proposed fold:** carry a 5d runbook row naming the invocation (for example `systemd-run` with the unit's environment file and user). Alternatively, narrow the verb's load to the keys it reads, which is a design change for the conductor to rule.

### 18. MINOR: Task 1's "stat unchanged, no read" row names no fixture that could show a read

**Location:** plan:863.

**Defect:** no seam observes a read, so "no read" is unprovable as written.

**Proposed fold:** name the fixture. Rewrite the file in place with the same inode and size, restore the modification time with `os.Chtimes`, and assert the old answer persists. That shows the cache served the lookup.

## The ten conductor-owed folds

| Fold | Judgment |
| --- | --- |
| 1. The split | Sound. Both halves are gradable, and 5c-i's exports have callers in 5c-i (Task 9 mounts `RequireAdmin` and `NewRosterAPI`; Task 7 calls `ChannelCommandLine`), apart from finding 9's two exports. |
| 2. Five fold tasks | Sound. The ruling's operative clause is ordering, and Tasks 1 through 5 precede every caller. The block is incomplete by one finding (finding 3). |
| 3. `streamer_username` required on write | Sound. Rung 6's ruling 8 and 6d criterion 6 force it. The errata list is incomplete (finding 13). |
| 4. The character set | Sound. `EqualFold` over an ASCII-only set closes the case collision, and the `LIKE` question is correctly carried. |
| 5. `ChannelCommandLine` | Sound. Rung 6 spec :815 ("except 5c's CLI grant") agrees. |
| 6. The last-admin guard | **Defect as applied**: it refuses `CreateRoster`'s own write (finding 1). The shape (an invariant, no demote method) is right. |
| 7. `doctor`'s key names on one line | Sound. The extraction mechanism is unspecified (finding 12). |
| 8. No admin checkbox | Sound. Fork 1 (`spec:188-192`) and the Geoff-ratified rung 6 spec (:815-816) both make the CLI grant and the import the only admin channels. |
| 9. Fail closed at capacity | **Defect as applied**: it reaches the client-keyed maps and creates a cheap group-wide lockout, and its bound claim is wrong (finding 2). |
| 10. The ceiling | The arithmetic is sound. Its visibility against section 15's ratified 4.5M line should be surfaced (finding 15). |

## Checked and sound

- Every relied-on behavior at the plan's load-bearing sites is quoted, with dates: `flock(2)` twice, `encoding/json/v2`, `os.SameFile`, `rate.Limiter.TokensAt`, `netip`, `url.Parse`'s error format, and `CrossOriginProtection`'s safe methods.
- Task 1's and Task 6's fixture states are named so that only a disk re-read explains each result.
- The four STATUS carries each land once: the rename and the guard in Task 6, the refresh call site in Task 8, and the mail copy in Task 4.
- The 5b-ii handoff's inherited contracts are each consumed or passed on: the interstitial's contract, `no-store`, `Referrer-Policy`, HSTS, and the three rejoin sentences.
- Each of the spec 5c row's exit-gate assertions lands in exactly one half, except the token-log assertion (finding 14).
- The security review's do-not-re-litigate list is carried into Task 0's inputs and into the pass-end security read's inputs.
