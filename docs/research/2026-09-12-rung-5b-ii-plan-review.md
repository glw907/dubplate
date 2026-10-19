# Rung 5b-ii plan review: the merged record

Three fresh-context adversarial reviews of
`docs/plans/2026-09-12-rung-5b-ii-auth-handlers.md` revision
1 (2712 lines), each read in full against HEAD `b91bae1` on 2026-09-12,
merged here with every finding's disposition. The three lenses:

- **Contract and code fidelity.** Every interface, type, function,
  constant, key name, file path, and behavior the plan attributes to
  landed code or to the rung 5 spec, verified by reading the code and the
  spec.
- **Security and threat model.** The handlers, the limiters, the origin
  check, the roster write path, the notify seam, and the verify seam as
  designed, graded against `spec:1907-1980`, `spec:1036-1314`, the ruled
  product forks, and the 5b-i security read.
- **Execution mechanics and budget.** Segments and boundary work, Files
  disjointness, contested surfaces, mutation detectability, the ceiling
  derivation, Task 0's grep-ability, ledger hygiene, per-task models, and
  the prior-art adoption criteria.

The plan is at **revision 2** with every finding folded. The three
per-lens files are superseded by this record and deleted.

`plan:` line numbers below are revision 1's, as each reviewer cited them;
`spec:` is `docs/specs/2026-09-08-rung-5-identity-auth.md`.

---

## Counts

| Severity | Contract | Security | Mechanics | Total |
| --- | --- | --- | --- | --- |
| Blocker | 1 | 1 | 0 | **2** |
| Major | 4 | 5 | 6 | **15** |
| Minor | 4 | 1 | 3 | **8** |
| **Total** | **9** | **7** | **9** | **25** |

All twenty-five are CONFIRMED. None is PLAUSIBLE-only.

| Disposition | Blocker | Major | Minor | Total |
| --- | --- | --- | --- | --- |
| Folded as recommended | 2 | 9 | 5 | **16** |
| Folded with adaptation | 0 | 6 | 3 | **9** |
| Rejected | 0 | 0 | 0 | **0** |

Folded as recommended: CONTRACT-1, CONTRACT-2, CONTRACT-4, CONTRACT-6,
CONTRACT-7, CONTRACT-9, SECURITY-1, SECURITY-2, SECURITY-4, SECURITY-5,
SECURITY-6, MECHANICS-2, MECHANICS-4, MECHANICS-5, MECHANICS-7,
MECHANICS-9. Folded with adaptation, each adaptation named in its own
entry: CONTRACT-3, CONTRACT-5, CONTRACT-8, SECURITY-3, SECURITY-7,
MECHANICS-1, MECHANICS-3, MECHANICS-6, MECHANICS-8.

**Three pairs are one defect seen through two lenses**, folded once
each: CONTRACT-4 with SECURITY-6 (the `SMTP_FROM` load-time validation),
CONTRACT-5 with MECHANICS-1 (the ceiling derivation), and CONTRACT-8 with
MECHANICS-8 (`config/roster.example.json`). **No two reviews contradict
each other.** The one place two remedies differed in force, the ceiling,
is recorded under CONTRACT-5 and resolved on the stronger evidence.

**Task count is unchanged**: Task 0 plus twelve tasks. No task was proved
unbuildable and nothing was split. **The ceiling moved from 5.5M to
7.5M**, flag from 4.4M to 6.0M, and the rung total from 22.4M to 24.4M
against R23's live 15.75M.

---

## The seven conductor-owed folds

Each is a decision the spec does not settle. Revision 2 applies each as
its reviewer recommended and marks it at its own site as
`**Conductor-owed fold:**` with the question and the applied answer, so
the conductor ratifies or reverses in one edit. The plan header lists all
seven.

| # | Question | Applied as | Site | From |
| --- | --- | --- | --- | --- |
| 1 | Rule the ceiling at 7.5M with the flag at 6.0M, or take the Task 8 / Task 9 cut now at drafting time? | The 7.5M ceiling, one pass | Token ceiling section | CONTRACT-5, MECHANICS-1 |
| 2 | Does R3's `doctor` roster collision row land in Task 4, widening its Files into `doctor.go` and taking it to five deliverables against the plan's four-deliverable cap, or route forward as a declared gap? | Task 4's fifth deliverable | Task 4 head | CONTRACT-3 |
| 3 | Does R9's "a send failure is a failed login" govern a `generic` deployment, or does `generic` answer before the send? | The narrow reading: R9 governs `kind` alone | R-5b28, Task 10 | SECURITY-3 |
| 4 | Is the admin-flag invariant enforced in the store's write path, or left to 5c's caller with a gap record? | Enforcement in the write path | Task 5 | SECURITY-5 |
| 5 | Does Task 8 run at `opus` unconditionally, get its own segment boundary, or run plainly at `sonnet`? | `opus` unconditionally | Task 8 head, Execution model | MECHANICS-3 |
| 6 | May M5 ship without a failing-first test if the allocation benchmark cannot be made to fail, recorded as a declared gap? | The named fallback the task reports | Task 1, outcome M5 | MECHANICS-5 |
| 7 | Does M4 drop the per-key rows, preserving R10, or does R10 get amended to admit per-key rows for the selected source? | Dropping the per-key rows | Task 2, outcome M4 | MECHANICS-6 |

---

## Findings, by id

### CONTRACT-1 (blocker): the pass declares seven remaining thresholds where the spec leaves six, and invents the seventh

**Claim.** Task 3 declares "seven thresholds ... fork 3's five rate
values" and names a seventh as "the consumption-attempt key's own
companion" (`plan:1437-1450`), with the count repeated at `plan:48`,
`:221`, `:353`, `:471`, `:1083`, `:1414`.

**Evidence, re-verified.** `spec:1587-1593` declares ten thresholds and
says "Product fork 3 sets the first seven values". Fork 3's seven are
enumerated at `spec:216-221` and ruled at `spec:228-229`. 5b-i landed
four of the ten at `cmd/dubplate-server/auth.go:75-100`, registered at
`:108-115`: `DUBPLATE_MAGIC_LINK_EXPIRY`, `DUBPLATE_SESSION_LIFETIME`,
`DUBPLATE_SESSION_ABSOLUTE_CAP`, `DUBPLATE_NONCE_RETENTION`, three of
them fork 3 durations. Ten minus four is **six**, which is what the
plan's own enumeration lists, and four of fork 3's seven remain rather
than five. No "consumption-attempt companion" exists in
`spec:1587-1593`.

**Disposition: folded as recommended.** Every count is six. "fork 3's
five rate values" reads "fork 3's four remaining rate values". The
phantom sentence is deleted and replaced by the arithmetic written out,
so the task cannot invent a name to reach a count. Task 0 item 11 reports
a net of plus six.

### CONTRACT-2 (major): the origin check's tightening and its `AddInsecureBypassPattern` exemption cannot both hold

**Claim.** The wrapper denies a protected request carrying neither
`Origin` nor `Sec-Fetch-Site` (`plan:1957-1960`), a bearer request is
exempt "through `AddInsecureBypassPattern`" (`plan:1968-1972`), and the
gate asserts a bypassed pattern's POST without either header is allowed
(`plan:2010-2012`).

**Evidence, re-verified on go1.27.1.** `Check` returns nil for the
neither-header case at the toolchain's `src/net/http/csrf.go:155-158`
("Neither Sec-Fetch-Site nor Origin headers are present") **before** any
bypass consultation. `isRequestExempt` (`csrf.go:185-198`) is unexported
over a private `c.bypass` field, so a wrapper cannot ask the stdlib value
what matched. A wrapper denying whenever `Check` returns nil and both
headers are absent therefore denies a bypassed path too, and the
contradiction would ship unexercised into rung 12's lap, because the plan
concedes no test in this pass can reach a bypassed path.

**Disposition: folded as recommended.** Task 8 owns its own bypass
pattern set in its own `http.ServeMux` with a sentinel handler, in the
shape `csrf.go:185-190` uses, consulted **before** the neither-header
denial, and its one registration method also calls
`AddInsecureBypassPattern` so the stdlib's cross-origin branch honors the
same pattern. That method is Task 8's produced surface. D10 carries the
evidence, R-5b27 carries the consequence, and two mutations were added
(consult the set after the denial; register through the stdlib alone).

### CONTRACT-3 (major): R3's `doctor` collision check is dropped, and no task, gap, or handoff carries it

**Claim.** D6 states the roster's validation with startup and write-time
checks (`plan:670-676`) and the plan bars `doctor` work beyond the M4
rendering fix (`plan:473-474`, `:591`).

**Evidence, re-verified.** R3 requires three checks:
`spec:325-327`, "collision-checked against every roster name at startup,
**in `doctor`**, and now at write time", repeated at `spec:1108-1110`.
No task in revision 1 built it, Task 12's gap list did not carry it, and
`spec:2332`'s 5d row gives 5d "`doctor`'s probe rows and its
bound-address row", which a roster validation row is not.

**Disposition: folded with adaptation.** The reviewer's primary option
was taken, the row landing in Task 4 with its Files widened to
`cmd/dubplate-server/doctor.go` and `doctor_test.go` for that row alone.
**The adaptation is the accounting**: Task 4 now carries five
deliverables, one above the cap the plan otherwise holds, and rather than
breach it silently the fold is marked conductor-owed at Task 4's head
with the alternative named (route it forward as a declared gap). The row
names no address and no principal value, on `identitySourceStatus`'s own
precedent at `cmd/dubplate-server/doctor.go:161-166`, and it is one row,
so R10's line-count arithmetic is untouched.

### CONTRACT-4 (major) and SECURITY-6 (major): the From value's load-time validation has no owner

**One defect, found by two lenses.** Task 6 carried the clause with a
conditional, "this task states the division and adds the load-side check
if Task 0 reports it absent" (`plan:1765-1769`, `plan:1749-1752`).

**Evidence, re-verified.** The rule is `spec:1289`, "The From value is
validated at config load rather than at send". The check is absent:
`validateMagicLinkSource` (`cmd/dubplate-server/auth.go:124-156`) checks
four secret lengths, calls `validatePublicURL`, and compares the nonce
retention against the link expiry, and never reads `DUBPLATE_SMTP_FROM`,
which is a Required key at `auth.go:52`. Task 6 could not have landed it:
its Files are `internal/notify/*`, `go.mod`, `go.sum`
(`plan:1731-1735`), while the validation lives in a file the package wall
assigns to Task 3 (`plan:470-471`). And no Task 0 item asked, so the
conditional's trigger could never fire.

**Disposition: folded as recommended, once.** The check is a Task 3
outcome: `DUBPLATE_SMTP_FROM` parses as a single address through
`net/mail.ParseAddress` and carries no CR or LF, checked inside
`validateMagicLinkSource`, reported by key name and never by value. Task
6 states the division and builds nothing there, with a grep criterion
proving it. **Task 0 gains item 23**, which reads the function key by key
and reports what it validates. SECURITY-6's operational argument is
carried into Task 3's outcome text: an unparseable From makes every send
fail, R9 turns every failed send into a failed login, so a typo in one
environment row would otherwise produce a deployment that starts clean,
passes `doctor`, and signs nobody in. Task 3's deliverable count goes
from three to four.

### CONTRACT-5 (major) and MECHANICS-1 (major): the ceiling's basis inverts the line it cites

**One subject, two lenses, no contradiction.** Both found revision 1's
derivation resting on figures its cited source does not carry. MECHANICS-1
carries the stronger evidence and its remedy is the one applied.

**Evidence, re-verified.** `docs/HISTORY.md:533-535` is the whole of
5b-i's token line: "this window's execution against the 5.0M ceiling
(4.0M flag) ran over, landing at roughly **5.5M** for Tasks 1 through 10
combined with their fix and simplifier rounds." `docs/HISTORY.md:540-543`
says the security cycle and 5b-ii's plan-prep tail "were both cut from
the window", so they sit outside that 5.5M. The 5.5M is therefore the
chain with its fix and simplifier rounds, not the all-in close figure
revision 1 read it as. The "3.4M chain" has no source: the only 3.4M in
the repo is 5a's 80% flag against a 4.25M ceiling
(`docs/HISTORY.md:1692`), and 5a's own chain was "1.68M across 18 agents"
(`docs/HISTORY.md:1689`). The 0.19M per task is the workstation-wide rate
at `docs/plans/2026-09-12-overnight-window.md:126-130`, whose
own next sentence reads "Close cost is a separate line at 0.9M-1.4M per
pass", so the 0.309M derived from it is not an all-in rate either. 5b-i
was budgeted "2.1M chain + 1.4M close" at that plan's `:331` and landed
at 5.5M for the chain half alone. 5a's itemization at
`docs/HISTORY.md:1688-1692` shows what an all-in figure looks like in
this ledger, and 5b-i's 5.5M carries none of those tail lines.

**Disposition: folded with adaptation.** MECHANICS-1's option (a) is
taken, the ceiling ruled at 7.5M with the flag at 6.0M and one pass kept.
**The adaptation is the per-unit arithmetic.** Both reviewers divided
5.5M by eleven chain units for 0.50M; HISTORY attributes the 5.5M to
"Tasks 1 through 10", which excludes Task 0, so revision 2 divides by ten
for **0.55M per task** and prices Task 0 separately at the 0.13M 5a
measured (`HISTORY:1688`). Twelve tasks at 0.55M is 6.6M, plus 0.13M for
Task 0 is 6.73M of chain, plus 0.25M for the pass-end security read,
0.20M for its fold, and 0.30M for the conductor's turns and the merge
ritual, for **7.48M**. Ceiling 7.5M, flag 6.0M. The 3.4M and 0.309M
figures are withdrawn with the reason written into the plan. The rung
total is restated as 3.90M + 5.5M + 7.5M + 4.5M + 3.0M = **24.4M against
the live 15.75M**, an **8.65M** overage recorded against R23's
reservation, with a note that 5c's and 5d's own figures are carried from
the same planning rate 5b-i overran and are not measurements. The
segment-B re-derivation trigger moves from 6.0M to **8.0M**, since the
old trigger was met at drafting time under the corrected basis. Task 12's
drafted spec corrections and the conductor's pass-end actions carry the
new figures, and the conductor now re-derives both 5c's and 5d's ceilings
at the close.

### CONTRACT-6 (minor): one of the two cited marker-convention exemplars does not exist

**Evidence, re-verified.** `internal/config/registry.go:375` is `// ruled,
conductor 2026-09-12, B2`, the convention.
`grep -n 'ruled' cmd/dubplate-server/doctor.go` returns nothing; line 164
sits inside `identitySourceStatus`'s doc comment, whose citation form is
`// (Resolution R-5b6).` at `:166`, which Task 12's grep would miss.

**Disposition: folded as recommended.** The `doctor.go:164` pointer is
replaced by `internal/auth/nonce.go:103` (`// ruled, conductor
2026-09-12, A7 fold 4`) and `internal/auth/epoch.go:58`, both verified,
and the grep Task 12 runs is written out as
`grep -rn 'ruled, conductor 2026-09-12, R-5b'`. The dead pointer is named
so a reader does not go looking for it.

### CONTRACT-7 (minor): three spec pointers land off their subject, one on a section break

**Evidence, re-verified.** (1) `spec:2179-2185` is the tail of the R24
review-shape erratum and the head of the R29 wording erratum; the split's
pre-authorization is the R23 erratum at `spec:2163-2172`. (2)
`spec:2506-2508` is the `.gitleaksignore` staleness bullet's tail, a blank
line, and a horizontal rule; the universality rule R-5b23 rests on is
`spec:1592-1593`, with the Universality section at `spec:2654` onward.
(3) The From rule is `spec:1289`; lines 1294-1295 are the salvage
package's two test-only fields.

**Disposition: folded as recommended.** All three repointed. The
five-line drift through the mail-seam citations is left to Task 0 to
report as notes under the plan's own grading rule, except
`spec:1310-1312`, which revision 2 corrected to `spec:1305-1306` in
passing while editing that outcome. Task 12's spec-correction list
records all three repointings.

### CONTRACT-8 (minor) and MECHANICS-8 (minor): `config/roster.example.json` is taken from 5d and lands in a directory with no ownership row

**One file, two gaps.** CONTRACT-8: `spec:2332`'s 5d row names "the
example roster file (R37)" and the plan takes it at Task 4 with the take
recorded nowhere. MECHANICS-8: `config/` has no row in the plan's
contested-surface table and none in `PA:175-210`, while `deploy.sh`
treats it as live surface (`:251`, `:284`, `:426`, `:537`).

**Evidence, re-verified.** R37 at `spec:751-752` assigns no pass and says
the repo ships the file "never as a file `deploy.sh` copies".
`deploy.sh` names each file it delivers rather than syncing the
directory, so nothing new reaches the box.

**Disposition: folded with adaptation.** Both remedies taken: the take is
recorded as a clause under R-5b20, Task 12 drafts the correction to
`spec:2332`'s 5d row, a `config/` row joins the plan's contested-surface
table, and the conductor writes the same row into `PA:175-210` at the
lane launch (pass-end action 7). Task 4 states that the example is not
deployed, citing `spec:751-752`. **The adaptation corrects the
reviewer**: MECHANICS-8 gave `config/contributors.yaml`'s retirement to
5d, and `spec:752-754` gives it to **rung 16** ("stays where it is until
rung 16 retires the incumbent importer, and rung 5 neither reads nor
moves it"). Revision 2 says rung 16.

### CONTRACT-9 (minor): Task 2's refresh clamp changes an exported signature the plan does not name as one

**Evidence, re-verified.** `internal/auth/session.go:130` is
`func RefreshSession(secret []byte, s Session, now time.Time, lifetime
time.Duration) (string, Session)`, setting `s.ExpiresAt =
s.IssuedAt.Add(lifetime)` at `:132` with no clamp and no cap parameter;
the cap lives on `VerifySession`'s list at `:167`. Implementing m2 changes
an exported signature Tasks 9 and 10 call, which is the same class of
change the plan orders the whole pass around for B1's `Insert`.

**Disposition: folded as recommended.** Task 2's Produces line names the
signature change in B1's own shape, with the current signature quoted and
both call sites named, so the reviewer is not handed a Produces line that
omits the change. The ordering already protected it; the cost was to the
reviewer and the cost is paid.

### SECURITY-1 (blocker): the pass forbids logging a roster address, and then requires it

**Claim.** Global Constraints read "No token, no secret, no cookie value,
and no roster address ever reaches a log record or an error string" and
close with "This is a criterion on every task from Task 4 onward"
(`plan:526-531`). D7 and Task 5 require an audit record naming the address
added (`plan:686-688`, `plan:1682-1684`), and Task 4's own criterion pulls
the other way (`plan:1582`).

**Evidence, re-verified.** `spec:1184` is "Tokens and secrets never enter
a log record"; `spec:1946` is "A token never reaches a log record".
Neither names addresses. `spec:1139-1141` requires "Every roster change
writes one audit record: who made it, through which assertion channel,
which address was added ...". The plan's constraint was stricter than both
clauses it cited, and the extra strictness is what collided.

**Disposition: folded as recommended.** The constraint is restated at the
spec's own width, tokens, secrets and cookie values, with both source
clauses quoted. A second bullet names the single exception, the roster
audit record Task 5 writes, and states that no other record in the pass
carries an address and no error string carries one anywhere. Task 4's
criterion now reads "outside the roster audit record Task 5 writes". Task
5's audit criterion asserts the address present. **Task 6 gains the
criterion the reviewer asked for**: the salvage package "logs one record
on success" (`spec:1296-1300`) and the natural attribute is the
recipient, so the seam's success record names the outcome and nothing
that identifies a person, its doc comment says why, and a test scans the
captured log for the recipient's own literal.

### SECURITY-2 (major): the consumption limiter has no place in the order the plan fixes

**Claim.** The plan fixes six steps in its strongest terms, "in this order
and no other" (`plan:2191-2193`), then adds "The per-client consumption
limiter gates the attempt" as an unordered sentence four lines later
(`plan:2197`). The request-link handler places its limiters explicitly ten
lines earlier (`plan:2181-2186`).

**Evidence, re-verified.** `spec:1178-1179`, "Consumption attempts keep
their single per-client-IP key." A brute-forcer fails at the MAC on every
attempt, so a limiter anywhere after the MAC is never reached by the flood
it exists to bound; a limiter after `Insert` burns a legitimate member's
nonce and then refuses the request. Task 10's criteria assert the four
order rows by nonce-directory count and none exercises the limiter's
position; no mutation moves it.

**Disposition: folded as recommended.** The limiter is the order's first
step, keyed on the reported client IP the caller supplies, stated in D12,
in Task 10's outcome, and in the Inheritance section's mapping sentence
(with a note that the quoted constraint is the security review's own words
and that the limiter is prepended to them rather than substituted for
them). One criterion added: N+1 attempts from one client address, every
one carrying a token whose MAC fails, are refused by the limiter rather
than by the verifier, asserted against the server log's outcome. One
mutation added: move the limiter after the MAC and that row stops
tripping. Exit-gate assertion 17 carries it.

### SECURITY-3 (major): the `generic` unknown-address answer is not required to be indistinguishable

**Claim.** The plan describes the two modes and requires nothing of the
generic one (`plan:2187-2190`), while R-5b22 rules the key's shape and
says nothing about what the generic answer must not reveal
(`plan:828-839`).

**Evidence, re-verified.** Fork 4 as ruled at `spec:244-247`: "The kind
answer names the configured admin contact, and a threshold flag switches
the whole deployment to the generic outcome." Two channels stayed open.
**Timing**: a known address runs a mint, a URL construction and an SMTP
dial before answering, an unknown one answers at once. **The response**:
the plan carries R9 verbatim, "a send failure is a failed login", so under
`generic` an unknown address returns the generic outcome while a known
address whose send fails returns a failure, and the difference is
membership. The plan's rubric carried no generic-mode claim, its attack
list answered enumeration with the rate bound rather than the oracle, and
Task 10's byte-identity criterion was scoped to consumption alone.

**Disposition: folded with adaptation.** The reviewer's outcome,
criterion and mutation are all added to Task 10, and the reviewer's
narrow reading is taken: R9 governs the `kind` deployment, `generic`
answers before the send, and a send failure under `generic` is logged
rather than rendered. **The adaptation is that the conflict is resolved as
a numbered resolution rather than a task note**, which is what the
reviewer asked for: **R-5b28** is new, stating the collision between two
ruled clauses, the reading that keeps both, and what changes under the
other answer, and it is marked conductor-owed. Two criteria were added
(the pair compared to each other, and the pair compared again when the
known address's send fails) and two mutations (surface the failure;
answer only after the send completes, caught by a blocking seam fake).
Rubric claim 12 and exit-gate assertion 15 carry it, and the attack list
names the oracle separately from the rate bound.

### SECURITY-4 (major): the fail-closed loader and the bootstrap that creates the file have no stated seam

**Claim.** R2 makes a missing roster file a load error and Task 4 says so
(`plan:1540-1542`); R37 makes `member add --admin` the bootstrap that
creates it and R-5b20 gives the creation to the store (`plan:808-810`).
Task 4's only constructor refuses a missing file, its Produces line names
no write-mode open (`plan:1526-1529`), Task 5 adds no constructor, and the
contested-surface row closes the package to 5c (`plan:586`).

**Evidence, re-verified.** `spec:749-752`: "A fresh install therefore
starts with no roster, which R2 makes a load error, and `dubplate-server
member add --admin` is what clears it: the verb creates the file when it
is absent and refuses to create a second one." Task 5's own criterion
asserts mode 0600 "after a create", proving a create path was expected
without naming who opens it. The cheap wrong resolution under time
pressure is to soften the loader into the fail-open shape R2 supersedes
(`_salvage/internal/config/contributors.go:33-39`), which refuses every
login as unknown and then, under fork 4's kind answer, hands every visitor
the admin contact.

**Disposition: folded as recommended.** D6 and Task 4's Produces line name
**two entry points**: a read-open that fails closed on an absent file, and
a write-open the CLI bootstrap calls that creates the file at 0600 and
refuses an existing one. R-5b20 restates both. Task 4 gains criteria for
all three behaviors and two mutations (write-open succeeding against an
existing file; read-open tolerating absence). Task 5 gains the criterion
that the read-open still refuses a missing file after the write-open
exists, plus the matching mutation. The 5c handoff names the write-open
constructor so 5c inherits a call rather than reopening the package.

### SECURITY-5 (major): the admin-flag invariant ships with no enforcement, no test, and no gap record

**Claim.** Task 5 deferred it: "the rule that only the CLI verb may set it
on an account that has never signed in is 5c's to enforce at its caller,
and this task's doc comment states where that rule lives"
(`plan:1665-1670`). The rubric's ten claims omitted it and Task 12's
six-item gap list omitted it, and the pass-end reviewer receives the
rubric rather than the task decomposition (`plan:903-906`).

**Evidence, re-verified.** The invariant appears three times:
`spec:1949` ("no web path sets it on a first login"), fork 1 at
`spec:191-192` ("the only path to the first admin is the CLI verb"), and
`spec:1236` ("granting the admin flag to a first login stays
impossible"). This pass ships the roster's only write path, and that path
already receives the assertion channel as an argument for the audit
record, so the store can already tell a cookie-asserted web add from the
CLI verb's operating-system actor.

**Disposition: folded as recommended.** The store refuses a write that
sets the admin flag on any assertion channel but the command-line actor,
with its own sentinel, because rung 5 has no promote path and the
add-member page is rung 5's only web writer. D7 states it, Task 5 carries
the outcome, the criterion (a cookie-channel add with `admin: true` is
refused and changes no byte; a command-line add with the same flag
succeeds) and the mutation. Rubric claim 11 and exit-gate assertion 16
carry it, so the pass-end reviewer grades an enforced invariant rather
than a prose claim. The 5c handoff says the add-member handler inherits a
refusal rather than a rule to remember. Task 5's deliverable count goes
from three to four.

### SECURITY-7 (minor): two of the three carried constraints are mapped wider than the tasks deliver

**Claim.** "Constraint 2 is a criterion on Tasks 9, 10 and 11"
(`plan:332-334`) while Task 11 carries no such criterion and no mutation
(`plan:2341-2357`). Constraint 3 is mapped to "Task 10's
`MaxBytesReader` call" while its own words are "**Every** auth endpoint
caps its body" (`plan:325-328`, `spec:1181`) and Task 10 builds three
endpoints, of which the sign-out handler's outcome named no cap.

**Disposition: folded with adaptation.** Both halves are taken, in
opposite directions, which is what the reviewer's "narrow the mapping, or
add the criterion" offered. **Constraint 2 is narrowed** to Tasks 9 and
10, with the reason stated (Task 11's error handling is fail-closed by
construction, so naming it would put a constraint on a task that does not
carry it). **Constraint 3 is widened** to all three endpoints: the
sign-out handler's outcome now caps its body, and the criterion asserts an
over-cap body refused without the parser running on all three.

### MECHANICS-2 (major): segment D runs Tasks 11 and 12 behind Task 10 with no place to rule a second `fix`

**Claim.** Segment D was Tasks 9, 10, 11 and 12 in one launch
(`plan:2648-2654`), while the plan itself states that pass-execute has no
mid-run hook (`plan:60-64`, `:2639-2643`) and that a second `fix` is the
conductor's decision (`:2635-2637`). Task 10 is where that decision is
expensive, and Tasks 11 and 12 sit behind it: Task 11 mounts Task 10's
handlers and Task 12 drafts the close from Task 10's outcome.

**Evidence, re-verified.**
`docs/plans/2026-09-12-overnight-window.md:134-139`: pass-execute
"cannot honor either plan's mid-pass conductor steps, nor the window's own
checkpoint cadence, **nor a decision on a second `fix` verdict before
every later task has already run**". Segment C's boundary guarded Task
10's inputs and nothing guarded its output.

**Disposition: folded as recommended.** Segment D splits at the Task 10 /
Task 11 boundary: **D1 is Tasks 9 and 10, D2 is Tasks 11 and 12.** The new
boundary's work is one line, accept Task 10 or rule the second `fix`, and
it is the cheapest boundary in the pass because nothing is re-run to reach
it. The segment table, the inline boundary blocks, the checkpoint
sentence, and segment C's own boundary text all carry the new names.

### MECHANICS-3 (major): Task 8's conditional upshift cannot be executed

**Claim.** Task 8 read "Upshift to `opus` if Task 7's review returns
`fix`" (`plan:1938-1939`) while the segment table puts Tasks 7 and 8 in
segment C (`plan:2653`). A pass-execute launch fixes each task's model in
its args at dispatch and the conductor sees nothing until the launch
returns, so the condition can never fire and the "named runner-up"
mechanism is dead text.

**Disposition: folded with adaptation.** The reviewer offered three
options without ranking them; **Task 8 runs at `opus` unconditionally**,
which is the option the plan's own argument already supports (its
correctness rests on a placement argument rather than code volume and its
blast radius is the whole served surface), and revision 2 adds a second
reason the reviewer could not have known: CONTRACT-2 found Task 8
carrying a bypass mechanism that cannot work as revision 1 described it.
The fold is marked conductor-owed with the two alternatives named. The
Execution model now lists four unconditional upshifts and states plainly
that no conditional upshift remains and why.

### MECHANICS-4 (major): exit-gate assertion 3 is owned by no task

**Claim.** "A request from a non-allowlisted peer carrying a forged
client-address header is limited and logged under its peer address"
(`plan:2535-2536`) is owned by nobody. Grepping the plan for `forged`
returned exactly one hit. Task 7 disclaims the surface ("this package does
not derive one", `plan:1869-1872`) yet carries mutation 4 for a defect
that by its own contract cannot live in `internal/auth/limit.go`. Task 10
takes `ClientIP` as a value; Task 11's criteria have no such row.

**Evidence, re-verified.** `ClientIP`'s own header handling is covered by
5a at `internal/serve/clientaddr_test.go:62` and `:206`; the "is limited
and logged" half needs a limiter and a handler, neither of which existed
at 5a.

**Disposition: folded as recommended.** The row joins Task 10's
acceptance criteria, where both the limiter and the log record are
reachable, with the 5a coverage cited so the task knows which half is
already proved. Task 7's mutation 4 is restated in place as a call-site
defect that belongs to Task 10, and it is added to Task 10's mutation list
as number 12. Exit-gate assertion 3 names its owner.

### MECHANICS-5 (major): M5's mutation is not detectable by the test the plan names

**Claim.** The plan requires every security-review reproduction to be a
gate-resident test that fails against the code as it stands today
(`plan:511-515`), and M5's named acceptance row, "a value of separators
returning the malformed sentinel with a bounded allocation"
(`plan:1232-1233`), passes before any change.

**Evidence, re-verified.** `internal/auth/token.go:244-248`: a separator
flood returns a huge `len(parts)` and fails `if len(parts) != count+1`;
a bounded `SplitN` with `count+2` returns `count+2` parts and fails the
same check at the same line with the same sentinel. The decode loop at
`:250-256` is never reached either way, so M3's strict decoder does not
separate them. `testing.AllocsPerRun` counts allocations rather than
bytes, and `strings.Split` allocates one slice plus one backing array
regardless of element count, so an allocation count is flat across the
mutation.

**Disposition: folded as recommended, with the reviewer's own fallback
made explicit.** The measurement is named:
`testing.Benchmark(...).AllocedBytesPerOp` against a threshold
proportional to the input length, with the threshold written out and its
reason stated. The acceptance row is rewritten to assert on bytes rather
than on the sentinel, with the reason written into the row. Mutation 7
reads on allocated bytes and adds the honesty clause: if the mutation does
not fire, the measurement is wrong rather than the mutation, and the task
reports that. **The fallback is marked conductor-owed**, because waiving a
failing-first test waives a Global Constraint: if the assertion cannot be
made to fail, the task reports M5 as fixed without one, with the
measurement it tried, and Task 12 records it as declared coverage gap 7.

### MECHANICS-6 (major): Task 2's M4 revokes a ruled, gate-pinned invariant (R10)

**Claim.** Task 2 was told to render "one row per unsatisfied
identity-source key" (`plan:1355-1357`), which contradicts R10.

**Evidence, re-verified.** `cmd/dubplate-server/doctor_test.go:146-152`
states the rule, "one per declared identity source regardless of how many
keys that source requires ... ruled, conductor 2026-09-07, R10", and
`:232-242` asserts it with the failure message "one line per source, not
per key". Nothing in the plan reconciled the two: Task 2's outcome never
cited R10, and Task 0 item 11's blast-radius arithmetic covered Task 3's
thresholds alone. Worse, the change might have passed R10's test by
accident, because that test's scratch source is declared but never
selected, so a per-key expansion scoped to the selected source adds no
line there.

**Disposition: folded with adaptation.** The reviewer's second option is
taken: **M4 holds to flattening the joined error and consulting
`Optional`, and the per-key rows are dropped**, because D5's load-bearing
property ("no `doctor` run reports a non-failing table while `Load`
returns an error") does not need them and R10 is ruled and pinned. The
fold is marked conductor-owed with the alternative named (amend R10 for
the selected source, updating the test's doc comment and failure message
in Task 2's own diff). **The adaptation**: the operator's need for a name
rather than a verdict is met inside the source's single row, whose text
may name the unsatisfied keys without adding a line, and the task reports
which form it chose. D5 is restated. Task 2 gains a criterion that
`TestDoctorCmd_LineCountFollowsRegistry` runs unmodified and the line count
is unchanged, and **Task 0 item 11 gains the reconciliation** so the
collision is graded before dispatch rather than at the gate.

### MECHANICS-7 (minor): two task headers declare a dependency set narrower than the task's own Consumes line

**Claim.** Task 10's header omitted Task 2 while its sign-out handler
calls Task 2's cookie-clearing helper (`plan:2185-2187`, `plan:1303-1305`).
Task 11's header read "Tasks 0, 3, 8 and 10" while its Consumes line named
Tasks 4, 5, 6, 7 and 9 (`plan:2296-2299`). The ordered list at
`plan:2683-2707` carried neither edge, and the plan states the marks exist
"so a conductor can reorder them when a dependency stalls".

**Disposition: folded as recommended.** Task 10's header reads "Tasks 0,
1, 2, 3, 4, 6, 7 and 9" with the reason stated, and revision 2 adds the
second half of that edge too, since Task 10's session path calls Task 2's
clamped `RefreshSession` (CONTRACT-9). Task 11's header names its whole
Consumes set. Both edges join the ordered dependency list with the reason
each was missed.

### MECHANICS-8 (minor)

Merged with CONTRACT-8 above. **Folded with adaptation.**

### MECHANICS-9 (minor): the `_salvage/internal/notify` adoption records provenance, not the four criteria

**Claim.** CLAUDE.md requires an adoption judged on license, code quality
against the house bar, maintenance, and adoption, "with the evidence
recorded in the plan". The `go-mail` adoption at `plan:1812-1817` meets it
and honestly flags adoption as the weak cell; the salvage adoption above
it argued provenance plus a deferral ("under fresh review at the
exemplary-Go bar"), with quality promised to execution rather than
recorded and no criterion disposed of as degenerate. Task 5's parallel
salvage adoption is the counter-example that shows the shape is available
(`plan:1689-1692`, citing `atomicfile.go:82-89` with the property being
adopted).

**Disposition: folded as recommended.** Task 6's prior-art block records
the evidence Task 0 item 14 collects: 96 lines, exported surface
`Mailer`, `New`, `Send`, `implicitTLSPort = 465` at
`_salvage/internal/notify/notify.go:24`, two test-only fields each
carrying a comment saying so, `notify_test.go` beside it, and the gate it
passes. License and adoption are stated as degenerate for in-repo code
with the reason. Maintenance is answered as this repo's own, since there
is no upstream to track. The fresh review at the exemplary-Go bar is
restated as a criterion the task is graded on, reporting what it changed
from the salvage shape and why, line by line, rather than a promise.

---

## What all three reviews checked and found sound

Recorded because it is this pass's do-not-re-litigate list, alongside the
5b-i security read's own "Attacked and found sound" section.

**Code fidelity.** Every plan pointer into the tree resolves at `b91bae1`
with the single exception of CONTRACT-6: the five security-review
reproductions, every `cmd/dubplate-server/auth.go` pointer, the serve
composition Task 8 inserts into line for line, the `internal/config`
family Task 3 declares through (so the remaining thresholds can be
declared and resolved with `internal/config` closed), `internal/atomicfile`
(where the spec's own pointer is stale and the plan's is right), the
stdlib claims in Task 8 measured on go1.27.1, `Store.Write` having no
caller outside its own file, R-5b26's premise (no redirect validator
exists; `MintLink` takes no destination argument), the salvage inputs,
every HISTORY range, and every spec section range in Primary inputs.

**Scope.** Every 5b item 5b-i did not ship has a task here, and the two
exit-gate assertions handed elsewhere are correctly placed. No 5c item is
taken: no task writes `web/`, the `member` verb, a theme, or the gate's
`web/` leg. CONTRACT-3 was the one gap in that mapping and CONTRACT-8 the
one unrecorded take.

**Security argument.** Constraint 1 is carried whole, with B1 as Task 1's
first deliverable and the `max` in one place inside the store. Replay is
attacked behaviorally rather than structurally. Enumeration keying is
right where the plan places it, both limiters ahead of the roster lookup
so a hit and a miss reach the same limiter. R27's derivation is neither
re-litigated nor re-derived. The limiter map's bound is proved the way R16
asks, sampled during the flood. The origin check's placement is argued
rather than assumed, and the coverage criterion runs against the shell as
the serve verb composes it. The post-login destination is disposed
correctly and the plan catches the `b91bae1` erratum's false claim rather
than repeating it. CSRF on the consumption POST is covered by
construction. Sign-out clears the cookie correctly and claims nothing
more. The seam reads the admin flag from the roster, with the provenance
row and its mutation. The notify seam's guard covers the subject. The
end-of-pass security read is required and scoped, with the task
decomposition withheld. The deferral of `Referrer-Policy: no-referrer`,
`nosniff`, and the consumption route's `no-store` to 5c is coherent.

**Mechanics.** The four inline boundary blocks, the checkpoint sentence,
and the segment table all named the same boundaries with the same tasks,
and each boundary carried work the next segment depends on. No task edits
`docs/STATUS.md` or `docs/HISTORY.md`, matching `PA:196`. Every task names
a model. Task 0's checklist is grep-able and its pointers resolve, with
the grading rule that makes line numbers safe stated up front; four of
item 4's six expected findings were spot-checked and all four land, and
item 3's "493 lines, untracked" is exact. Every contested surface the pass
writes has an owner row and rung 5 owns every one, with `config/` the only
gap. Files per task are disjoint enough to review and every overlap is
sequentially ordered by a declared dependency. The internal arithmetic of
the ceiling section is self-consistent once its inputs are granted, which
is why MECHANICS-1 cited the source lines rather than the totals. Most
named mutations are detectable by their named tests, and the two that are
not are declared as such rather than hidden.

**Task 2 is the widest diff** (twelve findings across six file pairs), and
the mechanics reviewer recorded that the batching is a judgment it would
not overturn while naming it as the place a reviewer's attention will be
thinnest. Revision 2 leaves the batching alone.
