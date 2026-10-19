# Rung 5c plan prose review

**Rung 5c plan review (revision 2.1)**

**The plan changed while I was reviewing it.** I read `675a1a0`. Partway through, `c73965b` landed. It records Geoff's approval ("all six rulings yes"), notes that the Navidrome query already ran, and widens the username pattern to `[A-Za-z0-9][A-Za-z0-9._@+-]*`. Line numbers below are at `675a1a0`. For anything past line 37, add 9 to get the line at HEAD. Findings 14 and 15 are about `c73965b` itself.

**Verdict.** This reads like a pass plan by the register's plausible author. It is specific and checkable, it states each trap with its mechanism, and it ties done to tests and mutations. The factual base is strong. Every Go stdlib, net/http, json/v2, flock, netip, unicode, os/user, x/time/rate and Navidrome claim I checked matches its source or my own measurement, with one net/http misstatement. The defects are left over from the two fold rounds:
- the split section's ruling box still describes revision 2's cut;
- the privileged-fields table points to a Task 10 that no longer exists;
- a few 5c-i sentences still assume the gate and web page are in this pass;
- the 5c-ii task count disagrees with the ceiling section;
- "assertions 1, 2, 3, and 5" of the spec's 5c row refers to a numbering that exists nowhere;
- every spec pointer from `spec:2316` on is 50 lines stale at HEAD, and Task 0's grading rule has no case for that.

After approval, the widened username pattern has no test row that would catch it being narrowed again. The main register problem is length. Revision history and lens credits recur throughout, and the rulings section is now a record.

**Scanner floor.** `tellgrader --register agent`: zero findings, 22,323 words, 1,165 sentences, cadence CV 1.07 (well varied), 16 tricolons (low for the length). No corpus manifest was named or found, so I graded against the agent-facing register alone.

**Measurements.** There is no `measures` object (the profile did not resolve), so the two share measures are unavailable.

| Measure | Value |
|---|---|
| sentences | 1,165 (scanner) |
| hinged_pair_share | unavailable |
| short_sentence_share | unavailable |
| average sentence length | about 19 words |
| longest sentence | Task 0 item 3's "Expected" semicolon series (about 130 words); Task 8's "The gaps are at minimum:" (about 105) |
| paragraph count | about 290 prose blocks (tables and headings excluded) |
| disproportionate paragraphs | Conductor's actions item 1 (about 210 words spanning three phases); the Global Constraints package-territory bullet |

**Claims I verified as correct:**
- **json:** `go doc encoding/json.Unmarshal` (case-insensitive match) and `encoding/json/v2` ("case-sensitive matching", "v2 rejects duplicate names"). I measured both error strings, plus the nil-`Members` cases, `null` members, invalid UTF-8 and trailing data. The `//go:build goexperiment.jsonv2` tag and `JSONv2: true` in the buildcfg baseline are both there.
- **os and errors:** `os.SameFile`, `errors.As` ("first error in err's tree") and `errors.Join` (`Unwrap() []error`).
- **flock(2):** all three quoted sentences.
- **netip:** `Unmap` and `Prefix` quotes; `PrefixFrom` uses `ip.withoutZone()`. I measured all four client-key rows.
- **net/url and os/user:** the `url.Error` format string (it does quote the password); the `lookup_stubs.go` fallback to `$USER` and `os.Getuid`.
- **net/http:** `CrossOriginProtection` safe-methods quote; the `Request.Form` quote.
- **rate and unicode:** `rate.Limiter.TokensAt` and `Burst`; `unicode.Cf`.
- **Kelvin sign:** `strings.ToLower` gives `a@kexample.com`, and `net/mail` accepts the form before lowering.
- **Navidrome v0.63.2:** `FindByUsername` builds `user.user_name = ? COLLATE NOCASE`, the interface comment matches, and `DbPath` defaults to `navidrome.db`. `compose.yaml` pins `0.63.2` with the `/data` volume.
- **Repo and tools:** `Makefile` uses `CGO_ENABLED=0`; the limiter defaults (5 and 20 per hour, 10,000 keys) and the ten-times ceiling comment match. The `sqlite3 -readonly` help text matches.
- **Arithmetic:** 4.72M, 5.2M, 37.1M, 21.4M, 9.25M, 18.3M and the 29 deliverables all check.

---

**Tier 1: invented or misquoted specifics**

1. **Tier 1, Warning. R-5c19, l.1026-1028.** Defect: "the request's own context has no deadline shorter than the server's 30-second timeouts" implies the server timeouts put a deadline on the request context. They don't. Evidence: `go doc net/http.Request.Context` says an incoming request's context "is canceled when the client's connection closes, the request is canceled (with HTTP/2), or when the ServeHTTP method returns." There is no deadline at all. The conclusion stands; the stated reason is wrong. Rewrite: "(an incoming request's context carries no deadline: `go doc net/http.Request.Context` says it is canceled only when the client's connection closes, an HTTP/2 request is canceled, or ServeHTTP returns, so a lock wait needs its own bound)".

2. **Tier 1, Suggestion (the query already ran). Ruling 2, l.70-73, and Conductor's actions item 1, l.2277-2289.** Defect: the plan says the query "writes nothing". Evidence: sqlite.org/wal.html says "The opening process must have write privileges for '-shm' wal-index shared memory file ... or else write access on the directory." A read-only open of a live WAL database still writes the wal-index. Rewrite: "which changes no data".

3. **Tier 1, Suggestion. R-5c3, l.566, and Ruling 1.** Defect: "tunnel brokers hand out /48s free" carries the lockout arithmetic, yet it is the one relied-on fact in the plan with no source. Rewrite: cite the broker checked (for example Hurricane Electric's tunnelbroker.net, after confirming), or soften it to "a /48 is cheap to obtain".

**Tier 2: contradictions left by the fold rounds**

4. **Tier 2, Blocker. The split's ruling box, l.211-212.** Defect: it says "cut between the Go-side work and the `web/` tree". That was revision 2's cut. Evidence: fold 1 (l.139-143) and the body (l.225-232) put the cut at the Task 7 / Task 8 boundary, with the Go admin gate and roster API in 5c-ii. The conductor ratifies this box. Rewrite: "Applied as the split, cut at the Task 7 / Task 8 boundary: the security fold, the CLI grant, and the `member` verb in 5c-i; the admin gate, the roster API, their wiring, and the `web/` tree in 5c-ii."

5. **Tier 2, Blocker. Privileged-fields table, `streamer_username` row, l.522.** Defect: "(Task 10 records it with the `LIKE` question)". There is no Task 10, and R-5c6 (l.689-696) settled the `LIKE` question. Evidence: Task 8 (l.2168-2172) records the rung 12 constraint. Rewrite: "(Task 8's HISTORY draft records it as a declared gap)".

6. **Tier 2, Warning. Global Constraints, l.465-466.** Defect: "The handlers and the gate it adds are `http.Handler` values". 5c-i adds no gate. Rewrite: "The handlers are `http.Handler` values that `cmd/dubplate-server` mounts, as 5c-ii's gate will be."

7. **Tier 2, Warning. Review focus item 1, l.1102.** Defect: "the web page and the CLI verb" are the two writers, but 5c-i has no web write path. Task 1's test uses two `Roster` values. Rewrite: "Two writers at once (two `Roster` values over one file, standing in for `serve` and the verb), with a hand edit landing between their writes."

8. **Tier 2, Warning. 5c-ii outline, l.2665 and l.2406-2407.** Defect: "the roster API stays at `<prefix>/roster` (5c-i)" and "The page is 5c-ii's; this task builds the handler" both speak as if the API were still in 5c-i. Rewrite: "(5c-ii Task 2)" and "The page is Task 10's."

9. **Tier 2, Warning. Task counts, l.2672-2673 against l.403.** Defect: the outline says "the two head tasks, about ten `web/` tasks, plus Task 0 and the close, thirteen in all". The numbered list runs 3-12, so that is 13 plus Task 0, or 14. The ceiling section says "thirteen tasks plus Task 0". Rewrite the outline: "thirteen tasks plus Task 0".

10. **Tier 2, Warning. Spec-row assertion numbering, l.251-253, l.2215-2218, Ruling 6 l.127, and l.2618-2628.** Defect: "assertions 1, 2, 3, and 5" of the spec's 5c row refers to a numbering the spec does not have. Evidence: the row is unnumbered prose with about 15 assertions. The 5c-ii head renumbers the same four as 1-4, with `member list` as item 4, not 5. Rewrite: name the four by content everywhere ("the non-admin 404, the same-origin denial, the web write's audit record, and `member list` reading the API's write") and drop the numbers.

11. **Tier 2, Warning. Task 3's interfaces (l.1544-1551) against R-5c20 (l.1039-1041) and Global Constraints (l.499).** Defect: Task 3 adds exported `AllowLinkClient` and `AllowLinkAddress`, yet the plan itself says "No other package calls the limiter". R-5c20 exports only for a caller outside the package. Evidence: `grep` shows the only callers are `internal/auth/handlers.go:166,213,265`. Rewrite: either name them `allowLinkClient` and `allowLinkAddress`, or add to R-5c20: "Methods on an already-exported type (`Limiter`) keep that type's surface; R-5c20 governs new names." Without one or the other, the architecture reader will flag it.

12. **Tier 2, Blocker. Spec pointers and Task 0's grading rule, l.1131-1135.** Defect: every spec pointer at or past `spec:2316` is 50 lines stale at HEAD. Evidence: `d82e9cf` (the 5b-ii close fold, committed before revision 2) added 50 lines of errata at 2316. The 5c row is now at 2381, not 2331. The re-derivation quote is at 2447, not 2395-2397. Risk 14 is at 2644 and the owner-touchpoints text at 2696. Task 0 grades a drift under about ten lines as a note and missing content as a blocker, so a 50-line drift with the content intact fits neither case. Rewrite: re-resolve the spec pointers at HEAD, and change the rule to "a pointer whose content survives at any offset is a note naming the new line; a pointer whose content is gone is a blocker."

13. **Tier 2, Warning. Task 7 mutation 3, l.2124-2125.** Defect: "Pass `ChannelCookie` → the `--admin` rows fail". The create row goes through `CreateRoster`, which takes no channel (Task 6's signature, l.1881-1883), so this mutation cannot touch it. Rewrite: "→ the second-add `--admin` row fails with the store's grant refusal."

14. **Tier 2, Blocker (`c73965b`). The widened username pattern is untested.** Defect: the pattern now admits `@` and `+`, but Task 2's rows (l.1474-1475) still accept only `a_b`, `A.b-c` and `x9`. A mutation that drops `@` passes every row, and the estate's one account name is email-shaped. Rewrite: add the row "`a@example.test`, `a+b@example.test` | streamer username, write | accepted (Ruling 2's estate check)" and the named mutation "Drop `@` from the class → the email-shaped row refuses."

    Three more spots went stale with the approval:
    - Ruling 2's "the one open fact is" is now settled.
    - Conductor's actions item 1 still says to run the query "Before Task 2". Mark it done, so a resumed conductor does not ssh to the box again.
    - Task 8's gap "if Ruling 2 was answered 'no'" can go.

15. **Tier 2, Suggestion (`c73965b`). Global Constraints, "Addresses reach exactly two outputs," l.488-493.** Defect: an email-shaped username is effectively an address, but the rule covers only the address field. Rewrite: add "A `streamer_username` counts as an address for this rule; the estate's is email-shaped."

16. **Tier 2, Suggestion. Ruling 6's "No" branch, l.127.** Defect: "about 2.3M" has no derivation. The ceiling section prices the two head tasks at 1.07M and one web task at 0.69M, which is 1.76M. Moot now; the fold record should either show where 2.3M comes from or correct it.

**Tier 3: owner-facing clarity of the rulings**

All six are ruled yes, so these findings matter for the record and for 5c-ii's bundle.

17. **Tier 3, Warning. Ruling 1, l.45-60.** Defects:
    - One yes/no carries two decisions.
    - It leans on terms the owner must look up in the review: "per-client", "least-recently-used", "/64", "/48", "bucket", "generic answer".
    - The "Under the generic answer" sentence covers only half the behavior; the kind answer's 429 appears only in R-5c3.

    Rewrite: "Should the sign-in limit on each address count only addresses on the roster? Recommendation: yes. The review's fix lets anyone with a few thousand free IPv6 networks fill the limiter and lock every member out of sign-in while the flood runs. Yes keeps made-up addresses out of the limiter entirely. A member over the limit sees 'too many requests' in kind mode, and the usual 'check your mail' in generic mode, with no mail sent. No builds the review's fix and records the lockout as a known risk."

18. **Tier 3, Warning. Ruling 5's "No" branch, l.108-114.** Defect: one sentence of about 80 words, with a colon and a list of five plan rows to rewrite. The owner needs the product outcome, not the plan's edit list. Rewrite: "No keeps the checkbox, so a signed-in admin can make someone an admin from the browser. That reverses part of fork 1, and the plan rewrites Task 6 and the roster API to allow it before either runs."

19. **Tier 3, Suggestion. Ruling 2, l.66-73.** Defect: the `COLLATE NOCASE` detail, the source path and the full command are author evidence, and the owner needs none of it to allow a read-only query. Rewrite: keep the question, the recommendation and both branches; leave the evidence in R-5c6.

**Tier 4: style and register**

20. **Tier 4, Warning. Process narration throughout.** Defect: the plan credits lenses 19 times and narrates revision history about 40 times. Examples: l.556-557 "revision 1 priced it wrongly", l.777-779, l.790-792, l.893-895, l.1443-1445, l.1491-1492, l.2101-2103, l.352-353, l.362-364. The register's rule is that every retained sentence must change what the agent does. Rewrite: keep the trap and drop the credit. For l.2101-2103: "Do not point the row at `/`: the create fails with `EACCES` whether or not the guard exists, so the row could not fail."

21. **Tier 4, Suggestion. Repeated sentence.** "Ruling 6 gates the launch; Rulings 2, 3, and 4 gate Task 2; ..." appears three times (l.42-43, l.326-327, l.2363-2365). Keep the Execution model copy.

22. **Tier 4, Suggestion. Bold lead-ins.** 94 of 196 bullets and 144 paragraphs open in bold. That dilutes the bold on the rules an agent must stop on, such as "No deploy of any kind" and "never `pgrep`". Reserve bold for those.

23. **Tier 4, Suggestion. Conductor's actions item 1, l.2271-2289.** Three phases (before Task 0, before Task 1, before Task 2) sit in one item of about 210 words. Split it into three items; the Task 2 part is now done.

**Counts:** Tier 1: 3. Tier 2: 13. Tier 3: 3. Tier 4: 4. Total 23: 4 Blockers (4, 5, 12, 14), 10 Warnings, 9 Suggestions.

**Sections that could move to `2026-09-23-rung-5c-plan-fold-record.md` without loss** (roughly 200-250 lines):
- The Status paragraph's revision history, l.3-26. Leave three lines and a pointer.
- The rulings section, l.38-129. Replace it with a six-row "ruled yes, 2026-09-23" table and move the full questions and "No" branches to the record.
- The dead "a 'no' ..." notes: R-5c3 l.552-554, R-5c6 l.668-673, R-5c13 l.919-921, and the "(written to its recommended answer; a 'no' ...)" clauses in the Task 2, 3 and 6 and 5c-ii Task 1 headers.
- The fold list's "Moved to Ruling N" placeholders, items 3, 4, 8 and 9 (l.147-148, 161-162).
- "The split"'s count paragraph (l.214-223) and its revision-2 history (l.227-232).
- The ceiling section's basis narrative (l.351-364), the 5c-ii provisional derivation, and the reservation arithmetic (l.395-418). Ruling 6 is ruled, and the conductor re-derives at close anyway. Keep the 5c-i arithmetic (l.373-393).
- Every "Revision 1/2 ..." and "the X lens's finding N" clause listed in finding 20.

The 5c-ii head (l.2394-2633) is 5c-ii's starting material, not fold history. It belongs in the plan or a 5c-ii seed file, not the fold record.

**Final verdict:** Yes, it reads as written by the register's plausible human author: precise, verified and mechanism-first. The seams from the two fold rounds and the length give away that it was machine-assembled; the individual sentences do not.

Files:
- `/var/home/glw907/Projects/dubplate/docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md`
- `/var/home/glw907/Projects/dubplate/docs/superpowers/research/2026-09-23-rung-5c-plan-fold-record.md`
- `/var/home/glw907/Projects/dubplate/docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`
- `/var/home/glw907/Projects/dubplate/internal/auth/limit.go`
- `/var/home/glw907/Projects/dubplate/internal/auth/handlers.go`
