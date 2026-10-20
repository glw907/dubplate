# Adversarial review: the rung 5 pass 5b-i plan

Fresh-context review (2026-09-08) of
`plans/2026-09-08-rung-5b-auth-core.md` revision 1, in two independent
lenses: auth security, reading the plan as an attacker against the token,
the session, and the two state records it specifies, testing what a
forged field, a replayed nonce, a refreshed cookie, and an offboarded
member can each reach, and grading every claim against the rung 5 spec's
threat model and 5a's own security-review record; and plan integrity,
checked clause by clause against the rung 5 spec revision 2.1, the spec
inventory, the auth prior-art record, HISTORY's 5a and rung 1b entries,
the 2.0 spec's forward hooks, the parallel-build approach, the 5a
exemplar plan and its review, and the tree at `6a9570d`. Folded as
revision 2 the same day; the merged 31-item "Verified sound" list below
is the do-not-re-litigate record for the pass, alongside the rung 5
spec's own records.

The two reviews are merged here: findings naming the same defect are
deduplicated, the stronger evidence kept and both lenses noted, and the
result renumbered B1 through B3, M-a through M-i, m1 through m16.
Thirty-seven raised findings became twenty-eight after nine merges. One
merge promoted a minor to a major, because the two lenses had found
different halves of the same dropped checklist bullet and only one half
carried a load-bearing consequence. Every finding was re-verified by the
fold editor before folding. Every mechanism claim about `net/url`,
`net/http` cookies, `crypto/subtle`, `os` file semantics, and the gate's
own gitleaks was measured on the installed toolchain (go 1.27.1,
gitleaks at `/home/linuxbrew/.linuxbrew/bin/gitleaks`) rather than
carried. Eight reviewer measurements did not survive that check in the
form they were written and are corrected below. None of the eight
retires a finding; two of them make a fold cheaper and more precise than
the review proposed. No finding was rejected outright. Nine items are
conductor-owed rather than folded silently, and each is applied
provisionally in the plan with an inline marker.

Every plan line cited below is revision 1's, measured at `6a9570d`. The
folds move many of them, so a reader chasing a pointer into revision 2
should chase the symbol rather than the line.

Both lenses agree the plan is well grounded where it reasons from the
spec, the inventory, and the tree. The pass split is taken at the spec's
own pre-authorized cut with all three of its instructions followed. The
5b-i and 5b-ii division of `spec:2168`'s scope is complete and
non-overlapping. Every exclusion ends in a stop condition. The registry
contract is followed in the words rung 1b wrote, and R-5b4's generic
form is a genuine improvement on R29's literal wording that the plan
flags as a correction rather than absorbing. The prior-art record is
cited accurately down to the version, the date, and the importer count.
The 2.0 hook accounting is right and no task builds a 2.0 feature early.
The prose carries no em dash and no en dash in 2,342 lines.

The defects concentrate in four places. The revocation epoch is checked
against the one session field rolling refresh moves, so a refresh after
an epoch bump un-revokes the session the epoch exists to kill. Two
mechanisms the pass needs have no landing site: the correlated
threshold validation cannot run inside the hook the plan contracts, and
address normalization is assumed by four tasks and owned by none. The
task that retires two `keys.go` literals breaks five gate-resident test
sites, two of them outside any task's declared territory, and the
pre-task predicts the opposite. And the token's wire encoding is
unspecified in a plan that mandates a checked-in literal of it.

---

## Corrections to the reviews' own evidence

**1. The auth lens's gitleaks fold is both wider and narrower than the
measurement supports.** It proposes banning `token`, `key`, `secret` and
`apikey` from any identifier sitting beside a high-entropy golden
literal. Measured on the gate's own binary with the gate's own
invocation (`scripts/check.sh:110`, `gitleaks detect --no-git -s .
--redact`, no config file, `.gitleaksignore` the tracked empty file), the
rule needs both halves, and either half alone is enough to stay clean.
`const testSecret = "dubplate-test-magic-link-secret-not-a-real-key-0123456789"`
reports "no leaks found", because the value's entropy does not clear the
threshold. `const linkSecret = "<48 hex digits>"` (the literal is not reproduced here because quoting it trips the gate)
fires generic-api-key. `const goldenVector = "<180-char base64>"` is
clean and the same literal named `goldenToken` fires at entropy 5.17;
`wantEncoded`, `nonceDigest` and `goldenSessionCookie` are clean, while
`goldenKey`, `magicLinkSecret` and `wantToken` fire. So the plan's
existing "visibly synthetic literal" rule (`plan:496-501`) already
covers every fixture secret with no rename at all, and the constraint
the pass actually needs applies only to the golden literals that must
carry high entropy because they carry a MAC. Folded as a conjunction
rather than an identifier ban, which is `check.sh:104-109`'s own reading
of the rule and `internal/config/keys_test.go:59-65`'s own precedent in
the tree.

**2. The integrity lens says the ceiling's 6.75M overage against R23's
14.5M "is not an error", on the ground that the spec's own erratum does
the arithmetic the same way.** The two operations are different.
`spec:2042-2050` moved the reservation from 14.5M to **15.75M** and
recorded the 1.25M difference at 5a's close. Recomputing the whole rung
against 14.5M now re-records that same 1.25M inside a 6.75M total. The
live baseline is 15.75M, the increment the split creates is **5.5M**, and
21.25M is where the reservation lands. The auth lens's m2 stands and this
Verified sound item does not; the item is dropped from the merged list
and the figure folds as m3.

**3. The integrity lens says `keys_test.go` and `doctor_test.go` "appear
in no task's Files list".** `cmd/dubplate-server/doctor_test.go` is in
**Task 2's** list (`plan:1237`), and `internal/config/registry_test.go`
is too (`plan:1235`). `internal/config/keys_test.go` appears nowhere in
the plan, confirmed by grep. The finding stands and the fold changes
shape: the two files Task 3's deletions break belong in **Task 3's**
list, because a file listed under the task before the one that breaks it
gives the breaking task no territory.

**4. The auth lens's Task 7 pointers land on the wrong clauses.** Task
7's Interfaces line is `plan:1817-1818`, cited as `:1815-1817`, and its
independence mark is `plan:1893-1894`, cited as `:1875-1877`, which is
mutations 4 and 5. The header the finding turns on is `plan:1806`
exactly as both lenses cite it, and the finding stands unchanged.

**5. Both lenses' pointers into the ordering section drift.** The
dependency list is `plan:2302-2311` and the mutual-independence paragraph
is `plan:2312-2318`. The auth lens cites `:2295-2301` and `:2298-2301`;
the integrity lens cites `:2305-2311` and `:2312-2315`. Only the
integrity lens's dependency-list pointer resolves, and the ordering
finding stands on the paragraph at its real lines.

**6. The two lenses disagree on how many times the `NewStrip` pointer is
wrong, and the integrity lens is right.** Four sites carry `:137`:
`plan:170-171` splits the pointer across a line break, and `plan:248`,
`plan:1077` and `plan:1731` carry it whole. The call is
`cmd/dubplate-server/serve.go:136`; line 137 is the `NewShell` call. All
four fold.

**7. The auth lens says Task 0's item 20 "partially absorbs"
`spec:2334-2337`.** It does not. Item 20's parenthetical enumerates
`spec:2315-2317`'s list exactly (R2.1's names, R4.5's rewrite, R5.6's
directive set, the ten gates) and reaches neither of `spec:2334-2337`'s
two questions, which ask whether the standard still ships no default
theme block and whether DL section 7 still gives the roster to rung 7.
The bullet is dropped outright. It is 5c's substance either way, so it
folds as a noted deferral rather than as a restored item, and only
`spec:2325-2327` folds as a restored item.

**8. The auth lens grades Task 6's construction-only closure of 5a's gap
4 against the plan's "a cryptographic property is proved by behavior"
constraint (`plan:491-495`).** That constraint reaches cryptographic
properties, and a construction value passed to a constructor is not one,
so the citation does not carry the finding. The finding stands on a
different and simpler ground: the property behind gap 4
(`HISTORY:340-341`) is behavioural, `internal/serve/strip.go:293-296`
records that `Shell` reads the name to remove that one cookie on the way
out, and 5a's existing test proves the removal for an arbitrary name
rather than for the real one. Retired to **m8** and folded as one
behavioural row.

---

## BLOCKERS

### B1. Rolling refresh advances the session field the epoch check reads, so a refresh after an epoch bump un-revokes the session permanently

**Lens.** Auth security.

**Plan lines.** A5 (`plan:648-658`), Task 6's refresh outcome
(`plan:1723-1725`) and its mutation 4 (`plan:1772-1773`), A8
(`plan:684-696`), Task 9's outcome (`plan:2027-2031`) and its criteria
(`plan:2046-2051`), the 5b-ii handoff clause (`plan:2172-2173`), and
threat-model claim 4 (`plan:920-921`).

**Evidence.** The session carries four fields, "the normalized address,
the issue time, an anchor time, and an expiry" (`plan:1717-1718`,
`spec:1071-1073`). Refresh "moves the expiry and never the anchor"
(`plan:1723`), and nothing says what it does to the issue time. Task 6's
mutation 4 settles it: "Check the cap against the issue time rather than
the anchor -> the refreshed-past-the-cap row fails" (`plan:1772-1773`)
can only fire if refresh moves the issue time, because otherwise the two
fields hold one value and the mutation is inert. Four fields where two
always agree would be three fields, which is the same reading from the
other side. Task 9's refusal check then reads exactly the field refresh
moves: "A session whose issue time predates its address's epoch is
refused" (`plan:2027`, A8 at `plan:691-692`, `spec:1085-1086`).

**Consequence, built as written.** A live session refreshed once after an
epoch bump carries an issue time later than the epoch, and every epoch
check from then on permits it. Correct ordering inside one request
(check, then refresh) saves the current request and not the next one,
because the cookie the response sets already carries the post-bump issue
time. Refresh runs on every request with no roster check and no token
check in front of it, so the window is one request wide and the effect
is permanent. R30 states the identical failure mode for a freshly minted
session in the same words: an outstanding link "mints a session whose
issue time postdates the epoch, which the epoch check then permits for
the full session lifetime" (`spec:643-645`). The plan spends Task 9's
mint-time check to close that door and leaves the refresh door open
beside it. `spec:1942` claims "an epoch bump refuses every session older
than it", and `plan:920-921` repeats the claim, so the pass's
`web-auth-security-reviewer` grades against a claim the code does not
hold. Nothing in Task 6, Task 9, or the handoff (`plan:2172-2173`, which
carries only the anchor rule) forbids the advance or fixes an ordering.

**Fix folded, conductor-owed.** The epoch's session check reads the
**anchor**, which refresh never moves and which is the session's original
issue time. A8 (`plan:684-696`) and Task 9's outcome (`plan:2027`) say
so, and Task 6's outcomes name the anchor as the original issue time so
the field's meaning travels with it. Task 6 gains an acceptance row: a
session refreshed after an epoch bump is still refused, with the refresh
performed before the check. Task 9 gains the same row against the epoch
store, and a mutation "check the epoch against the issue time rather
than the anchor", which fires on it. The rule goes into the 5b-ii
handoff verbatim, so the middleware inherits it. The close drafts the
spec correction against `spec:1085-1086`, alongside the R29 and R-5b7
corrections already listed.

### B2. Task 4's R31 correlated check cannot run inside the hook Task 2 contracts, and Task 4's territory forbids widening it

**Lens.** Auth security.

**Plan lines.** Task 2's hook contract (`plan:1263-1264`) and A2
(`plan:601-608`); R-5b10's threshold declarations (`plan:818-821`); Task
4's outcome (`plan:1509-1510`) and its Files list (`plan:1489-1490`); the
territory wall (`plan:451-466`); 5b-i's exit gate (`plan:340`); declared
coverage gap 5 (`plan:2136-2139`).

**Evidence.** The hook runs "over the resolved values" of a declared
source's required and optional keys (`plan:1263-1264`, A2 at
`plan:601-608`). R-5b10 declares `DUBPLATE_MAGIC_LINK_EXPIRY` and
`DUBPLATE_NONCE_RETENTION` as registry **thresholds** (`plan:818-821`),
not as source keys, and `Registry.claim`
(`internal/config/registry.go:119-124`) panics through `panicDuplicate`
(`registry.go:95-97`) when one key is declared in two families, so they
cannot be both. Measured against the tree, the hook as contracted never
sees either value: a threshold resolves at its point of use through
`config.ParseDuration(environ, t)`
(`internal/config/thresholds.go:91-93`), and `Load` reads no threshold
value at all (`internal/config/config.go:248-287`, recorded as 5a
security-review finding 24). Task 4 nonetheless requires the hook to
refuse "a `DUBPLATE_NONCE_RETENTION` below `DUBPLATE_MAGIC_LINK_EXPIRY`
(R31)" (`plan:1509-1510`) and requires that "Every refusal joins the
rest of `Load`'s errors" (`plan:1512-1514`), which only holds inside
`Load`. Task 4's Files list is `cmd/dubplate-server/auth.go` and
`auth_test.go` alone (`plan:1489-1490`), and the wall makes an
`internal/config` edit outside its task a stop-and-report
(`plan:464-466`), so the implementer cannot widen the signature where
the requirement lands.

**Consequence, built as written.** One of 5b-i's own eight exit-gate
assertions, "a configuration whose nonce retention floor is below the
link expiry fails to load" (`plan:340`), has no landing site. R31
(`spec:650-660`) is the ruling behind it and it is the only thing
standing between an adopter who raises the link expiry and a consumed
nonce pruned while its token still verifies, which is the property
`spec:1938-1940` claims and which R11 spends a create-exclusive insert
and two fsyncs to buy. The implementer's available shapes are all wrong:
validate at first use rather than at load, redeclare the two durations as
source keys and lose `doctor`'s effective-value row (R-5b6's reason for
the threshold family), or edit `internal/config` outside its task.
Declared coverage gap 5 (`plan:2136-2139`) asserts that "nothing
validates a threshold value", which contradicts Task 4's own outcome and
confirms that the two tasks were written against different hook
contracts.

**Fix folded, conductor-owed.** Task 2's Produces and A2 state the
hook's signature: the hook receives the source's resolved values **and
the environment snapshot** in `os.Environ()` form, so a rung's own file
can validate a registered threshold's value, and a relation between two
of them, without putting a rung 5 name in `internal/config`. Task 2
gains an acceptance row proving it against a fixture: a threshold
registered on a `DUBPLATE_TEST_SCRATCH_*` name, a fixture hook that
refuses one value of it, and an assertion that the refusal joins `Load`'s
other errors. Task 2 gains a mutation "pass the resolved values alone",
which fires on that row. Task 4's Consumes names the environ-bearing hook
explicitly. Declared coverage gap 5 is corrected to say that finding
24's residual is closed for the two correlated thresholds this pass
validates and stands for every other registered threshold.

### B3. Task 3's two `keys.go` deletions break five gate-resident test sites, two of them outside its declared territory, and Task 0 item 10 predicts the opposite

**Lens.** Plan integrity.

**Plan lines.** Task 0 item 10 (`plan:1042-1052`), Task 3's Files list
(`plan:1367-1379`) and its gate criterion (`plan:1444`), Task 2's Files
list (`plan:1229-1237`), A1 (`plan:590-597`), the no-vacuous-assertions
constraint (`plan:486-490`).

**Evidence.** Item 10's expectation is that
"`internal/config/config_test.go`,
`cmd/dubplate-server/check_test.go`, `doctor_test.go`, and
`shadow_test.go` all build registries through `config.NewRegistry()`,
which registers no source, so none changes". Three of the four change,
and a fifth file item 10 never names changes too. Measured against the
tree at `6a9570d`:

1. `internal/config/keys_test.go:5-21`, `TestOwnedWall_IdentitySourceRow`,
   reads the package-level `OwnedWall()` and calls
   `t.Fatal("OwnedWall() has no identity source row")` when the row is
   absent. Task 3 deletes `keys.go:24`, so it fails outright. It builds
   no registry, which is why the registry-constructor survey never sees
   it.
2. `internal/config/keys_test.go:75`, `TestReservedClassOf`'s row
   `{input: "DUBPLATE_SMTP_HOST", wantClass: "auth and identity",
   wantRung: 5, wantFound: true}`, fails once Task 3 deletes
   `keys.go:59-73`.
3. `cmd/dubplate-server/doctor_test.go:243-258`,
   `TestDoctorCmd_LiteralSpotCheck`, builds through
   `config.NewRegistry()` at `:245` and asserts at `:255-258` that
   stdout contains the literal line `identity source
   pending (arrives at rung 5)`. The row it spot-checks is the row
   `NewRegistry` seeds from `keys.go:24`, so building through the fresh
   registry is exactly why it breaks.
4. `internal/config/config_test.go:303-316`,
   `TestLoad_RegisteredThresholdIsNotUnknown`, calls
   `Load(append(baseEnviron(), "DUBPLATE_CLASSIFY_THRESHOLD=7"), reg)`
   against `classificationRegistry()`, which registers
   `DUBPLATE_CLASSIFY_OWNED` as an owned-now wall row
   (`config_test.go:195`), and asserts `err == nil`. A1 makes that unset
   row a load error, so it fails at **Task 2**, one task before the
   deletions.
5. `internal/config/registry_test.go:302-307`,
   `TestCollisionGate_IdentitySourceCarveOut`, asserts that
   `r.claims["identity source"]` is absent. It goes vacuous once no
   status row exists, against `plan:486-490`.

Of these, `keys_test.go` is in no task's Files list, and
`doctor_test.go` and `registry_test.go` are in **Task 2's**
(`plan:1235-1237`) while Task 3 is the task whose deletions break them
(correction 3). Task 3's criterion is "`bash scripts/check.sh` green in
the worktree" (`plan:1444`), which it cannot reach inside its declared
territory.

**Consequence, built as written.** Task 3 hits a red gate on files it is
not authorized to open, and the wall's own stop condition
(`plan:464-466`) fires on a task doing exactly what its plan told it to
do. This is the same defect class 5a recorded as its one planning miss:
"the template-classification consequence of rung 1b's contract, which a
planning-time probe of `config_test.go` would have caught"
(`HISTORY:526-529`). Two of the five sites also need a decision rather
than a repair, because the row they assert is gone: the pass owes a
statement of what replaces `TestOwnedWall_IdentitySourceRow` and
`TestDoctorCmd_LiteralSpotCheck`'s identity-source line, and an
implementer left to choose will weaken an assertion rather than move it.

**Fix folded.** Task 3's Files list gains `internal/config/keys_test.go`
and `cmd/dubplate-server/doctor_test.go`, and the territory wall gains
both (M-h). Task 0 item 10 is rewritten to survey by grep for the two
deleted literals rather than by registry constructor, naming these five
sites as its prediction and asking the pre-task to confirm or correct
it, with a mismatch a blocker for Task 3 rather than a note. Task 3
gains three acceptance criteria: `TestOwnedWall_IdentitySourceRow` is
deleted rather than weakened, because its subject no longer exists;
`TestDoctorCmd_LiteralSpotCheck`'s identity-source line is replaced by
the owned-now selector row printed against the production registry, so
the spot check keeps a literal to check; and `TestReservedClassOf`'s
`DUBPLATE_SMTP_HOST` row moves to the magic-link source's required set
with the transition named. Task 3 gains a named mutation for
`registry_test.go:302-307` going vacuous, restated so it asserts the
carve-out against a fixture status row the test registers itself. Task 2
gains an acceptance criterion that `classificationRegistry`'s
environment satisfies its own owned-now row, which is the fourth site.

---

## MAJORS

### M-a. No task owns address normalization, and four tasks take "the normalized login address" as given

**Lens.** Auth security.

**Plan lines.** Task 1's Produces (`plan:1151-1156`), A4 (`plan:635-637`),
A6 (`plan:662-665`), Task 5 (`plan:1587-1588`), Task 6 (`plan:1717-1718`),
Task 9's outcome (`plan:2017`) and its normalization criterion
(`plan:2052-2054`).

**Evidence.** `spec:1161-1163` states the property and its scope: "Email
matching is case-insensitive on a normalized, length-capped, lowercased
address, and the normalization happens before any keying, any logging,
and any rate-limit lookup." The plan uses "the normalized login address"
as a given at `plan:1151`, `:1587`, `:1717`, `:663`, `:686` and `:2017`,
and no task's Outcomes or Acceptance criteria land a normalizer. No task
states the length cap as a literal anywhere. Task 9 alone asserts
normalization behaviour, and it asserts it of the epoch store
(`plan:2052-2054`).

**Consequence, built as written.** The epoch store normalizes internally
while the token minter and the session minter accept whatever a caller
passes. The epoch record's name is the digest of the normalized address
(A6, `plan:663-665`), so a divergence between what Task 5 embeds in a
MAC and what Task 9 digests produces a silent revocation miss rather
than an error. That is exactly the failure Task 9's own criterion names,
"A non-normalized key silently fails to revoke" (`plan:2053-2054`), and
the criterion closes it for one of the three call sites. The
`internal/config` wall already normalizes an address in the repo's own
idiom (`internal/config/config.go:162-173`, `mail.ParseAddress` then
`strings.ToLower`), so the vocabulary exists and has no home in this
pass. The length cap is the other half: without it, a caller-supplied
address bounds nothing, and `spec:1150-1151` requires the principal to be
"ASCII-printable, unique, and bounded".

**Fix folded.** The normalizer is assigned to **Task 1**.
`internal/identity` is the dependency-free vocabulary leaf every
consumer already imports, `net/mail` is stdlib and already on the pass's
allowed list (`plan:481-484`), and the package's no-internal-import rule
(`spec:1467-1469`) is unaffected. Task 1 gains an outcome for a
normalize function that parses as a single RFC 5322 address, lowercases,
and refuses above a stated length cap written out as a literal with its
reason. Task 1 gains acceptance rows for a case difference, the cap
boundary at and one over, and an unparseable value, plus a criterion
that the returned value is what the identity's address field carries.
Tasks 5, 6 and 9 call it rather than assume a caller did, stated in each
task's Consumes line. The rule joins the 5b-ii handoff in
`spec:1161-1163`'s own words, so the limiters and the roster inherit it.
Task 1's deliverable count moves from 2 to 3.

### M-b. The token's wire encoding is unspecified, and the plan narrows the spec's serialization claim to the MAC input alone

**Lens.** Auth security.

**Plan lines.** A4 (`plan:636-638`), Task 5's first outcome
(`plan:1589-1591`), its golden-vector criterion (`plan:1619-1623`), its
field-shifting criterion (`plan:1624-1626`), the threat-model rubric
(`plan:931-932`).

**Evidence.** `spec:1017-1019` makes the claim about the serialization:
"The serialization is unambiguous: every field is length-prefixed, so no
value can be shifted across a boundary by a separator inside it." The
plan narrows it to the MAC input: "Every field is length-prefixed **in
the MAC input**" (`plan:1589`), and its only grading row is a MAC-input
row, "two different field triples that would serialize identically
without length prefixes are asserted to produce different MACs"
(`plan:1624-1626`). Nothing in Task 5 states the encoded form Task 5 also
mandates as a checked-in literal (`plan:1619-1623`), and nothing requires
an address containing the encoding's separator to round-trip.

**Consequence, built as written.** The implementer picks the wire format,
and the common shape joins fields with a separator and appends the MAC.
If the separator is `.`, `first.last@example.test` produces too many
parts, the parse is malformed, and every dotted address silently cannot
log in. Measured, that is the majority shape of a real address. The
failure is closed rather than open, because a shifted parse produces a
different length-prefixed MAC input and the MAC fails, so no forgery
follows. The cost is a working login for one adopter and a broken one for
the next, decided by a choice no acceptance criterion grades: the golden
vector pins one address and the field-shifting row pins the MAC input,
and both pass with the hole open.

**Fix folded, conductor-owed on the encoding.** Task 5 gains an outcome
stating the wire encoding, applied provisionally as base64url per field
before any separator is applied, so no field value can carry the
separator, with the reason documented beside the MAC-input reason. Task 5
gains two acceptance rows, a golden vector for an address carrying the
separator character and one for an address carrying `+` and `=`, each
asserted against a written-out literal. A4 (`plan:636-638`) is amended so
the length-prefix claim covers the serialization as `spec:1017-1019`
words it rather than the MAC input alone. Measured for the link builder
the same rows travel through: `url.JoinPath` plus `Query().Encode()`
percent-encodes `+`, `/` and `=` into
`?t=AAAA%2B%2F%3Dabc`, so base64 padding and slashes survive the query
whatever the body encoding is.

### M-c. The nonce record's carried expiry is unspecified, and Task 8's outcome and its own acceptance criterion take opposite readings

**Lens.** Auth security.

**Plan lines.** Task 8's outcome (`plan:1926-1928`), its Consumes line
(`plan:1910-1911`), its R31 criterion (`plan:1953-1956`), A7
(`plan:677-682`), R-5b10 (`plan:821-829`).

**Evidence.** Task 8's outcome says the sweep prunes on "the record's own
carried expiry" (`plan:1926-1928`) and its Consumes line takes "Task 5's
nonce and expiry" (`plan:1910-1911`), which is the token's expiry. Under
that reading the record dies exactly when the token dies,
`DUBPLATE_NONCE_RETENTION` has no runtime consumer at all, and R31's
load-time floor guards a mechanism the design does not have. Task 8's own
acceptance criterion asserts the opposite reading: "with a retention
below the link expiry, a still-valid token's nonce would be swept"
(`plan:1953-1955`) can only happen if the retention value governs the
record's lifetime. R-5b10's justification is written in the second
reading throughout ("a day of retention survives a clock adjustment or a
paused process", "under a longer one, nothing breaks and the directory
grows", `plan:824-829`). The spec is no clearer: `spec:1664-1668` says
the record carries its expiry, the sweep prunes on it, and the retention
threshold "is a floor config load holds at or above the link expiry".

**Consequence, built as written.** Under the first reading, an exit-gate
assertion of this pass proves nothing, R31's whole ruling is inert, and
the pass ships an adopter-visible duration knob with no consumer.
Whichever reading an implementer takes, the reviewer grading `spec:1940`
("a nonce cannot be consumed twice ... across any accepted retention
configuration") cannot tell from the code which mechanism was intended.

**Fix folded, conductor-owed.** Task 8's outcome states the record's
carried value: consumed-at plus the resolved `DUBPLATE_NONCE_RETENTION`,
never the token's expiry, which R31's floor then holds at or above the
link expiry so the record always outlives every token that could bear
its nonce. Task 8's Consumes line takes the resolved retention from Task
4's resolver rather than Task 5's expiry. Task 8 gains an acceptance row
that a record survives past its token's expiry when retention exceeds
it. A7 (`plan:677-679`) says the same thing in one clause. The close
drafts the spec correction against `spec:1664-1668`.

### M-d. The sweep deletes a malformed nonce record, which fails open on a record a crash truncated inside the link expiry

**Lens.** Auth security.

**Plan lines.** Task 8's outcome (`plan:1931-1933`), its acceptance row
(`plan:1957`), its mutation 5 (`plan:1977-1978`).

**Evidence.** Task 8 mandates deletion: "A sweep that meets a malformed
record removes it and reports the count, rather than aborting the pass,
because a record it cannot parse is a record it cannot honor"
(`plan:1931-1933`), asserted at `plan:1957` and protected by mutation 5.
The insert creates the record exclusively and then writes it
(`plan:1918-1921`). Measured on go 1.27.1: `os.OpenFile` with
`O_CREATE|O_EXCL|O_WRONLY` and mode `0600` succeeds once, the created
file reports `-rw-------`, the second open returns an error satisfying
`errors.Is(err, os.ErrExist)`, and a file created and closed with nothing
written reports size 0. So a crash between the create and the write
leaves a zero-length record.

**Consequence, built as written.** While the zero-length record stands, a
replay is refused, because `O_EXCL` does not care what the file
contains. The sweep then removes it as malformed, and the token that
produced it is consumable again for as long as its expiry allows. Task
8's stated reason points the wrong way here: the record whose expiry is
unknown is precisely the one whose token may still verify, so the safe
reading of an unparseable record is to keep it. `spec:1938-1940` claims a
nonce cannot be consumed twice "across a restart", and a crash is the
restart case this store's two fsyncs exist for.

**Fix folded, conductor-owed.** Task 8's malformed-record rule becomes
retain-and-report: a record whose payload does not parse is retained
until its file mtime is older than the resolved retention, then removed
and counted, so an unknown expiry never shortens single use. Mutation 5
is restated as "delete a malformed record on sight -> the
crash-truncated-record row fails". Task 8 gains an acceptance row that
writes a zero-length record by hand, runs the sweep, and asserts a
second insert of that nonce still returns the replay sentinel. The
mtime fallback is the one place the store reads a filesystem timestamp,
and the doc comment says why.

### M-e. gitleaks fires on the golden token Task 5 mandates, and the plan predicts no `.gitleaksignore` append

**Lenses.** Auth security, with the mechanism re-measured by the fold
editor (correction 1).

**Plan lines.** Task 5's golden-vector criterion (`plan:1619-1623`), Task
6's cookie criterion (`plan:1745-1747`), the fixture constraint
(`plan:496-501`), the `.gitleaksignore` contested-surface row
(`plan:558`), the gate criterion on every task (`plan:526-527`).

**Evidence.** Measured on the gate's own binary with the gate's own
invocation (`scripts/check.sh:110`), `.gitleaksignore` the tracked empty
file and no config: a 180-character base64 literal on a line whose
identifier is `goldenToken` fires generic-api-key at entropy 5.17, and
so do `goldenKey`, `magicLinkSecret` and `wantToken`. The same literal
named `goldenVector`, `wantEncoded`, `nonceDigest` or
`goldenSessionCookie` reports "no leaks found". A low-entropy value stays
clean under any name: `const testSecret =
"dubplate-test-magic-link-secret-not-a-real-key-0123456789"` is clean,
while `const linkSecret = "<48 hex digits>"` fires.
`check.sh:104-109` records the mechanism and the precedent in the gate
itself: "The generic-api-key rule fires only when a key-like field name
sits on the same line as a value whose entropy clears its threshold ...
and it is renamed rather than suppressed, so no `.gitleaksignore` entry
remains." `internal/config/keys_test.go:59-65` carries the same rule as
an in-tree comment.

**Consequence, built as written.** Task 5 mandates a checked-in encoded
token as a literal and Task 6 mandates the same shape for the session
cookie. Both must carry high entropy, because both carry a MAC, so the
low-entropy escape the plan's fixture rule offers does not reach them.
The plan states that "5b-i expects to append nothing" to
`.gitleaksignore` (`plan:558`) while requiring the gate green at every
task end (`plan:526-527`), and no task carries the constraint that
reconciles them. The pass's highest-risk task is set up to commit red on
its central deliverable, and the cheap wrong fix is an allowlist entry on
a contested shared surface.

**Fix folded.** The Global Constraints fixture paragraph
(`plan:496-501`) gains the measured rule as a conjunction: gitleaks'
generic-api-key rule needs a key-like identifier **and** a
high-entropy value on the same line, so a fixture secret stays clean
through its visibly synthetic value, and a golden literal that must
carry entropy stays clean through an identifier that does not read as a
key name (no `token`, `key`, `secret` or `apikey`). Both halves cite
`check.sh:104-109`'s rename-rather-than-suppress precedent and
`keys_test.go:59-65`. The constraint is an acceptance criterion on Tasks
5, 6, 8 and 9, and `plan:558`'s "expects to append nothing" becomes the
assertion it then earns.

### M-f. `spec:1081`'s "Refresh counts same-origin requests only" appears nowhere in the plan or in the handoff

**Lens.** Auth security.

**Plan lines.** A5 (`plan:648-658`), Task 6's refresh outcome
(`plan:1723-1725`), the 5b-ii handoff clause (`plan:2172-2173`).

**Evidence.** `spec:1080-1082` states three properties of rolling refresh
in one paragraph: "Rolling refresh moves the expiry and never the anchor,
so the absolute cap survives refresh with no server-side state. Refresh
counts same-origin requests only." The plan carries the first and drops
the third. Grep for "same-origin" across the plan returns three hits:
5b-ii's exit-gate cell (`plan:341`), the SSO precondition sentence
(`plan:1721`), and nothing else. 5b-i builds the refresher
(`plan:1709-1712`) and 5b-ii builds the caller and the origin check, so
the constraint sits in the crack between the two halves, and the handoff
clause that would carry it names only the anchor.

**Consequence, built as written.** A cross-origin subresource load keeps
a session alive up to the absolute cap on a machine its owner has walked
away from. That is the shared-device adversary the threat model names
(`spec:1917-1918`, `plan:906-908`), and the rolling lifetime is 90 days
under R-5b10, so the window is the whole cap. 5b-ii's middleware is where
the rule is applied and this pass's handoff is where 5b-ii learns it.

**Fix folded.** A5 (`plan:648-658`) gains the clause, and Task 6's
Outcomes carry it as a stated precondition of the refresher's contract,
documented on the refresher itself so a caller reading the function
learns it. The clause goes verbatim into the 5b-ii handoff beside the
anchor rule and B1's epoch rule, so the middleware inherits all three
rather than rediscovering them.

### M-g. R33 is reassigned away from 5b against a handoff the plan calls binding, R-V is dropped, and R24 is misrouted the same way

**Lenses.** Both, independently, on R33; the R-V half is the integrity
lens's; the R24 half is the fold editor's, found while re-resolving the
routing line.

**Plan lines.** The ruling routing (`plan:84-87`), the handoff declared
binding (`plan:92-94`), the Inheritance section (`plan:233-262`), Task
10's handoff enumeration (`plan:2169-2187`), R-5b16 (`plan:929-941`).

**Evidence.** The plan declares `HISTORY:379-408` binding on it
(`plan:92-94`) and then assigns R33 to "5c and 5d" (`plan:87`).
`HISTORY:393-394` assigns it to this pass: "R33's disposal as 5b's
post-login destination rule (`spec:674-686`, not Task 3's)". The
Inheritance section walks the handoff's clauses and never dispositions
it, so the reassignment carries no argument and no stop condition.
`HISTORY:391-392` carries a second clause the plan drops entirely, "the
30-second read timeout and the `http.MaxBytesReader` alternative if a
later rung wants a streaming upload (R-V)", which bears on
`spec:1590`'s auth request body cap that R-5b10 defers to 5b-ii
(`plan:826-827`). Task 10's handoff enumeration (`plan:2179-2187`) lists
5a's carries as `ClientIP`, the origin check's insertion point, the
`Origin`/`Sec-Fetch-Site` keep-set, and the public-origin URL rule, and
names neither. The same routing line lists **R24** as 5c's and 5d's
(`plan:87`) while R-5b16 (`plan:929-941`) reads R24, argues against its
one-read-inside-5b wording, and drafts the correction, so the misrouting
is the same defect twice.

**Consequence, built as written.** R33's half that lands in 5b-i is
concrete: `spec:690-692` says that if a destination parameter is ever
introduced, "it carries it inside the MAC-covered token rather than as a
free query parameter on the consumption POST", which constrains the token
format Task 5 fixes and the extensibility its version label buys. Task 5
states neither that the token carries no destination field at rung 5 nor
that the version label admits one later, so 5c inherits a token whose
extension point nobody named and a rule nobody wrote down. R-V's read
timeout is the fact 5b-ii needs before it caps a request body, and a
clause dropped from a binding handoff has no carrier at all.

**Fix folded, conductor-owed.** The routing line (`plan:84-87`) moves R33
and R24 into 5b-i's read list. The Inheritance section gains two rows:
R33's disposition, stating that rung 5 accepts no caller-supplied
destination, that the token carries no destination field at 5b-i, and
that any later destination rides inside the MAC-covered token; and R-V's
read timeout with the `MaxBytesReader` alternative, carried to 5b-ii for
the body cap. Task 5 gains one outcome recording that the version label
exists so a later field can be added without changing the meaning of an
existing one, and one acceptance criterion that the minter's signature
offers no destination. Both clauses join Task 10's handoff enumeration
with the `HISTORY` line that carries each.

### M-h. The package-territory wall omits five files the plan's own tasks must write

**Lens.** Plan integrity.

**Plan lines.** The wall (`plan:451-466`), against Task 2's Files list
(`plan:1229-1237`) and Task 3's (`plan:1367-1379`).

**Evidence.** The wall enumerates the writable set as
`internal/identity`, `internal/auth`, `internal/config` ("`config.go`,
`registry.go`, `thresholds.go`, and `keys.go`"), one new file under
`cmd/dubplate-server`, `main.go`'s register calls, `serve.go`'s
cookie-name wiring, `serve_test.go`'s `productionRegistry`,
`templatekeys_test.go`'s `classifyKey`, and one env template row. Task 2's
own Files list modifies `cmd/dubplate-server/doctor.go` and
`doctor_test.go` (`plan:1236-1237`), neither of which the wall names, and
`internal/config/config_test.go` and `registry_test.go`
(`plan:1234-1235`), which the four-file enumeration excludes. The wall
names `internal/identity` "including its tests" and does not extend the
same clause to `internal/config`. `doctor.go` is absent from the
forbidden list too (`plan:461-465`), so a `diff-reviewer` grading Task
2's diff against the wall has no ruling to apply, and the wall's closing
stop condition would fire on a task doing what its plan says. B3 adds two
more files the wall must name.

**Consequence, built as written.** The wall is the artifact the merge
ritual's rebase and the pass-end reviewer both read a diff against, and a
wall that contradicts three task Files lists is worse than no wall,
because a reviewer cannot tell an authorized edit from a scope leak. The
5a plan review's Verified sound item 29
(`research/2026-09-08-rung-5a-plan-review.md:849-854`) records that this
plan family states its file sets exactly, and this one does not.

**Fix folded.** The wall's `cmd/dubplate-server` clause names
`doctor.go`'s per-source row loop, `doctor_test.go`'s fifth-family
assertion, and `doctor_test.go`'s literal spot check. The
`internal/config` clause changes from a four-file enumeration to
"`internal/config`, each file for the change its task names, its tests
included", which reaches `config_test.go`, `registry_test.go` and
`keys_test.go`. The forbidden list is unchanged, and `internal/serve`
stays on it under R-5b15.

### M-i. Task 0 claims the spec's staleness checklist whole and drops two bullets, one of them R-5b7's entire evidence base

**Lenses.** Both. The integrity lens graded it major on R-5b7's
dependency; the auth lens graded it minor on the count. Merged and
graded major, because the dropped bullet is the one an argument rests on.

**Plan lines.** Task 0's preamble (`plan:970-972`), its items 11 and 20
(`plan:1067-1071`, `:1114-1116`), the ceiling table's staleness line
(`plan:388`), R-5b7 (`plan:780-793`).

**Evidence.** The plan states "The checklist is the spec's own
(`spec:2281-2345`), every bullet, reduced to nothing and extended with
the six this pass adds." The spec carries seventeen bullets
(`spec:2291`, `2295`, `2297`, `2301`, `2303`, `2307`, `2310`, `2313`,
`2315`, `2318`, `2322`, `2325`, `2328`, `2331`, `2334`, `2338`, `2342`).
Mapped item by item against Task 0's twenty-four, fifteen are carried,
nine items are this pass's own additions, and two are dropped:
`spec:2325-2327` ("Does `env/dubplate-server.env.template` still carry
the three row states in its header, and does `deploy.sh` still abort the
render on one unresolved name?") and `spec:2334-2337` (the
default-theme-block pair). Item 11 reads what the templates list, not
the header's row states (correction 7 settles that item 20 does not
absorb `spec:2334` either). R-5b7 rests its whole argument on exactly
the two facts `spec:2325` asks about: that the template carries a
literal second row state
(`env/dubplate-server.env.template:8-16`, verified today, with the
literal state at `:12-13`) and that a bare row would abort `deploy.sh`'s
render. The count is also arithmetically inconsistent with the
twenty-four-item list `plan:388` prices, which says "seventeen bullets
plus the six this pass adds".

**Consequence, built as written.** The pre-task no longer checks the
facts R-5b7 turns on, so the one resolution whose argument reaches a file
outside `internal/config` is unverified at dispatch. If the greenfield
template's header changed, Task 3 writes a row in a state the template no
longer documents and `deploy.sh` decides at the next render what the plan
assumed at authoring.

**Fix folded.** `spec:2325-2327` is restored as a Task 0 item, with the
expectation stated as the three row states at
`env/dubplate-server.env.template:8-16` and `deploy.sh`'s
abort-on-unresolved-name path by line, since R-5b7 turns on both.
`spec:2334-2337` is named and deferred to 5c in one clause on item 20,
as item 20 already does for `spec:2315`. The preamble becomes "sixteen of
the spec's seventeen bullets, the design-language default-theme bullet
named and deferred to 5c, extended with the nine this pass adds", and
`plan:388` is restated to match. Task 0's item count moves from
twenty-four to twenty-five.

---

## MINORS

**m1. The `NewStrip` call is at `serve.go:136`, cited as `:137` at four
sites.** Raised by both lenses, with the integrity lens's count of four
correct (correction 6).
`strip := serve.NewStrip("", "", serve.TrustedProxyConfig{},
serve.FrontExemption{})` is
`cmd/dubplate-server/serve.go:136`; `:137` is `shell := serve.NewShell(...)`,
a different call. The four sites are `plan:170-171` (split across a line
break), `plan:248`, `plan:1077` and `plan:1731`. By Task 0's own grading
rule (`plan:965-968`) a pointer landing on a different subject is a
blocker; it grades minor here because the symbol name travels with every
citation, so an implementer self-corrects. The adjacent pointers all
resolve: `NewStrip`'s signature is `internal/serve/strip.go:300` exactly
as item 14 states it, and `strip.go:293-296` records that the second
value is the shell's session-cookie name. Folded at all four sites.

**m2. The 15.75M ceiling erratum is at `spec:2042-2050`, not
`spec:2078-2087`.** Raised by both lenses. `spec:2078-2087` is the tail
of the hop-by-hop-scope erratum and the head of the trailer-fields
erratum, a different subject 36 lines away. The ceiling erratum is
`spec:2042-2050`: "R23's ceiling table (`spec:563-567`) is corrected ...
Rung 5's combined reservation moves from 14.5M to **15.75M**, the 1.25M
difference recorded against the reservation ... about 3.9M against the
4.25M ceiling." The substance the plan states is right at the right
lines. Folded, and added to Task 0's expected-findings set so the
pre-task confirms rather than discovers it.

**m3. The 6.75M overage is computed against the superseded 14.5M
reservation the plan itself says was already moved.** `plan:417-419`
states the reservation stands at 15.75M after 5a's close, and
`plan:421` then computes "21.25M against R23's 14.5M" and calls the
difference 6.75M, repeated at `plan:2246-2247` as a conductor-owed
ruling. Against the live 15.75M the difference is **5.5M**, and the
1.25M inside the plan's 6.75M was recorded at 5a's close already
(correction 2). The row label "5a 4.25M actual-against-ceiling"
(`plan:419-420`) also uses a ceiling in a sum the sentence calls an
actual; 5a's actual was about 3.90M (`HISTORY:521`). Folded, conductor-owed:
the overage is restated as 21.25M against 15.75M, a 5.5M difference
recorded rather than taken from another pass, the 5a row is relabelled
"4.25M ceiling, about 3.90M actual", and the figure carries into the
conductor's action list.

**m4. R-5b7's "other eight non-secret names" double-counts, and the
spec's own "eight" is itself wrong.** Raised by both lenses.
`spec:1640-1645` reads "Rung 5's eight required non-secret names get a
row state too" and then enumerates seven: `DUBPLATE_IDENTITY_SOURCE`,
`DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL`, and
`DUBPLATE_SMTP_HOST/PORT/USER/FROM`. Counted independently against the
required-when-magic-link table (`spec:1524-1536`), the non-secret names
there are six, and the selector is the seventh. The plan takes the
selector at 5b-i and says "The other eight non-secret names stay 5d's"
(`plan:791`), which leaves six and implies a nine-name total the spec
nowhere states. `spec:2170` also gives 5d "its eight literal rows". The
substantive half of R-5b7 checks out: `TestTemplateKeysCoverEveryOwnedNowRow`
(`cmd/dubplate-server/templatekeys_test.go:74-89`) reaches only
owned-now wall rows, so none of the six trips a gate before 5d. Folded
as "the other six non-secret names", with the count correction drafted
against both `spec:1640` and `spec:2170`.

**m5. Task 7's declared dependency on Task 4 contradicts its own
Interfaces line, its independence mark, and the plan's dependency
list.** Raised by both lenses. Task 7's header reads "Depends on Tasks 0
and 4" (`plan:1806`), its Interfaces line says "Consumes: nothing from
this pass beyond the state directory as a plain string"
(`plan:1817-1818`), and its independence mark says "Independent: yes, by
files. It writes `store.go` only and consumes nothing from Tasks 1, 5,
or 6" (`plan:1893-1894`, correction 4). The dependency list names "Task
7 -> Task 8" and never "Task 4 -> Task 7" (`plan:2302-2311`), and the
independence paragraph calls Tasks 1, 2 and 7 mutually reorderable
(`plan:2312-2318`). The state directory is rung 1's `Config.StateDir`
(`internal/config/config.go:75`), not one of Task 4's four thresholds, so
nothing Task 4 produces reaches Task 7. Folded: Task 7's header becomes
"Depends on Task 0". The same paragraph describes Task 2's territory as
"`internal/config`'s four files" where its Files list also carries two
`cmd/dubplate-server` files, so the paragraph names them too and the
reorderability claim becomes checkable as written.

**m6. R38's localhost carve-out refuses `http://127.0.0.1` and
`http://[::1]`, which browsers treat as secure contexts.** R38's own
reasoning is the browser rule: "a browser refuses a `Secure` cookie over
plain HTTP on any origin but localhost" (`spec:760-761`). The W3C Secure
Contexts definition of a potentially trustworthy origin covers
`127.0.0.0/8` and `::1` as well as the name `localhost`, so the check as
the plan words it, "whose scheme is not `https` unless its host is
`localhost`" (`plan:1509-1510`), refuses a loopback-IP adopter for a
reason that does not apply to them. Measured on go 1.27.1:
`url.Parse("http://127.0.0.1:8080").Hostname()` is `127.0.0.1` and
`netip.ParseAddr` reports it loopback; `http://[::1]:8080` gives
hostname `::1`, also loopback; `http://192.168.1.10:8080` gives
`192.168.1.10`, not loopback. The plan's acceptance rows
(`plan:1523-1525`) name only `http://localhost:8080`. Folded,
conductor-owed because it widens a ruled clause: the carve-out becomes
scheme https, or a host whose parsed value is `localhost` or whose parsed
address reports loopback, with the W3C definition cited by URL and read
date. Acceptance rows are added for `http://127.0.0.1:8080` passing and
`http://192.168.1.10:8080` failing, and the widening is drafted as a spec
correction against `spec:758-767`.

**m7. Task 4's URL check is stated as two conditions where it needs
three, and its parse-failure row is nearly unfireable.** Raised by both
lenses, one on the missing host condition and one on the fixture.
Measured on go 1.27.1: `url.Parse("example.test")` and
`url.Parse("not a url")` both return `err == nil` with `Scheme` and
`Host` empty and the whole value in `Path`; only a value such as
`"://bad"` errors, with "missing protocol scheme". So the acceptance row
"a value that does not parse as a URL fails" (`plan:1525`) needs a
deliberately chosen fixture to fire at all, and an implementer picking
the obvious one gets a row that passes through the scheme branch while
proving nothing, which `plan:486-490` forbids. A bare hostname passes the
parse and is caught only incidentally by the https test, and a URL with
an empty host produces a path-only link, which defeats security-review
finding 7's whole point by producing no origin. The passing localhost row
also turns on a mechanism the plan does not name:
`url.Parse("http://localhost:8080").Host` is `localhost:8080` and only
`Hostname()` is `localhost`, so a check written against `Host` fails the
row the plan requires to pass. Folded: Task 4's outcome states three
conditions, parses, carries a non-empty host, and carries the https
scheme unless the host is loopback (m6), with the comparison made against
the parsed hostname. The "does not parse" row splits into two, a bare
hostname (`example.test`) refused for an empty host and `://bad` refused
for the parse, each with the key named.

**m8. Task 6 closes 5a's coverage gap 4 with a construction assertion
where the gap's property is behavioural.** Task 6's only criterion for
the gap it closes is "`serve_test.go` asserts the shell is constructed
with the cookie name rather than the empty string" (`plan:1758-1759`).
The property behind gap 4 (`HISTORY:340-341`, "The session-cookie
removal is tested with an arbitrary cookie name (R-E), because the real
name arrives at 5b") is behavioural, and
`internal/serve/strip.go:293-296` records that `Shell` reads this name to
remove that one cookie on the way out to the fallthrough, which is the
surface rung 12 proxies to a streamer. The auth lens graded this against
the plan's cryptographic-property constraint, which does not reach it
(correction 8). Folded as one behavioural row: a request carrying
`__Host-dubplate_session` alongside another cookie reaches the
fallthrough with the session cookie absent and the other cookie
readable, run against the shell as the serve verb constructs it, so the
gap closes on behaviour and the construction row stays as the wiring
check it is.

**m9. The roster-removal exit-gate assertion is assigned to 5b-ii in
prose and omitted from 5b-ii's own exit-gate row.** `plan:343-346` rules
that "A link for an address removed from the roster" is refused by the
consumption-time roster re-check, "which is 5b-ii's". The 5b-ii
exit-gate cell (`plan:341`) enumerates six assertions and does not carry
it, and 5b-ii's plan is drafted just in time from this pass's handoff
(`plan:18-19`), so the assertion has no carrier at all. It is one of
`spec:2168`'s own assertions and R30 (`spec:637-648`) is the ruling
behind it. Folded into 5b-ii's exit-gate cell and named in the 5b-ii
handoff beside the epoch's refusal checks.

**m10. The token gains a mint-time field the spec's token section does
not carry, and the close drafts no correction for it.** Raised by both
lenses. `spec:1016-1017` lists the token's fields as "a normalized login
address, an expiry, and a 128-bit random nonce from `crypto/rand`", three
fields. R30 then requires the consumption POST to compare "the token's
own mint time" against the address's epoch (`spec:640-642`), and Task 9's
criterion asserts the link-before-the-bump row through it
(`plan:2049-2051`). So the plan is right to land four fields and the
token would be unbuildable at three; the defect is that the spec's own
enumeration contradicts R30 and the close's correction list
(`plan:2145-2168`) does not draft it. Folded as one line in the drafted
corrections: `spec:1016-1017`'s field list gains the mint time, on R30's
own requirement.

**m11. The deploy consequence the plan says needs stating rather than
discovering is itself wrong about what unblocks a deploy.**
`plan:301-309` says "the next deploy of a 5b-i-or-later binary fails at
`Load` until `/etc/dubplate/server.env` carries the row". R-5b7 puts
`DUBPLATE_IDENTITY_SOURCE=magic-link` into
`env/dubplate-server.env.template` as a literal at Task 3, and
`deploy.sh` renders `/etc/dubplate/server.env` from that template, so the
rendered file carries the row from Task 3 onward. Verified against the
template as it stands: it carries no rung 5 name, so `Load` then fails on
the nine required-when-selected keys (A2 at `plan:605-609`,
`spec:1522-1536`), three of them secrets rung 16 mints
(`spec:1632-1638`). The paragraph's conclusion is unaffected, since no
deploy runs during rung 5's open passes and no gate leg reads a rendered
environment. Folded by restating the mechanism, so 5d inherits the right
list of what its exit gate must satisfy (`spec:2170`).

**m12. The seventeen-item decomposition that triggers the split prints as
sixteen; the revocation seam is missing by name.** `plan:321-332`
enumerates `spec:2168`'s 5b row as sixteen semicolon-separated items and
then says "that is seventeen items". Counted independently:
identity leaf, token serialization, session cookie, nonce store, epoch
store, roster store's five parts, consumption re-checks, limiters,
origin check, verify seam, `internal/notify`, reserved-class removal,
wall-row retirement, the `Load` edit, correlated validations,
supersession rewrites. `spec:2168` names "the revocation seam as a read
plus a notification" as an item distinct from the epoch store, which is
where the seventeenth went. The split still triggers either way, since
`spec:2179`'s rule is "exceeds fifteen tasks" and the plan's own 5b-i /
5b-ii division sums to eleven plus six. Folded by adding the seam to the
enumeration between the epoch store and the roster store.

**m13. `docs/STATUS.md` carries no section named "lane registry", yet the
plan makes it the blocking live-executor check.** The plan names the lane
registry in `docs/STATUS.md` at `plan:61-63`, `:192-193`, `:448-450` and
Task 0 item 1 (`plan:974-983`), where a "no" stops the launch. Verified:
`docs/STATUS.md` at `6a9570d` has four sections, Current state, Next
action, Open items (carried), and Pinned versions. Lane state lives as
prose inside Next action at `STATUS:24-32`, which reads "Lane-1 (main
tree): IDLE ... Lane-2 (worktree `lane-2-5a`): IDLE, torn down at 5a's
merge ... No lane active; next session picks one before dispatch."
Item 1's expectation, "lane-1 taking pass 8c in the main tree", is not
what STATUS says. `git worktree list` shows one tree at `6a9570d
[master]`, so the tree half of the blocker passes on the facts. STATUS is
also at exactly 60 lines, so Task 10's "under 60 lines after the edit"
requires moving rather than trimming. Folded, conductor-owed: the
conductor adds a labeled "Lane registry" block to `docs/STATUS.md`
before the first dispatch, or the plan's four references name Next
action as where lane state lives. Item 1's expectation is restated as
"lane-1 IDLE with pass 8c's plan as its next action", and Task 10's
STATUS criterion says the pass ledger moves to HISTORY rather than being
summarized.

**m14. Task 3 calls the same addition both the fourth and the fifth
register call.** `cmd/dubplate-server/main.go:84-86` carries three
register calls (`registerShadowIntegration`, `registerShadowThresholds`,
`registerServeConfig`). Task 3's Files line (`plan:1374-1375`) and the
merge-carry table (`plan:269`) both say fourth; the outcome at
`plan:1405` says "so a fifth register call is added once". The
contested-surfaces row (`plan:555`) says "one register call", which is
right either way. Folded to "fourth".

**m15. Task 2's "six classes" instruction conflates two countings the
tree keeps apart.** A2 calls a declared source's key "the wall's sixth
classification" (`plan:609-610`), which is right in one sense:
`internal/config/thresholds.go:104-111` says `thresholdOf` "is the fifth
classification the wall recognizes, alongside `ownedNow`,
`reservedClassOf`, `supersessionOf`, and `integrationClassOf`", so a
sixth lookup joins five. But `config.go:6-42`'s package doc numbers a
different set of five (owned-required-now, reserved, adopter-optional
integration, superseded, unknown) that does not carry thresholds as a
numbered class at all; thresholds appear only inside class 3's Optional
sentence and in the prose at `config.go:44-49`. Extending the numbered
list to six for sources leaves thresholds unnumbered while a comment
three files away calls them the fifth. Folded: Task 2's outcome states
which counting the doc edit takes, applied as numbering the source class
sixth and adding one sentence reconciling thresholds' unnumbered status,
with the reconciliation named as part of the deliverable.

**m16. Four supporting citations resolve to adjacent or wrong lines.**
Raised by the integrity lens as two findings, merged here. (1)
`plan:280` and `plan:313` cite `spec:2489-2490` for "Anything else is a
conductor ruling"; that sentence is `spec:2488`, and `2489-2490` are a
blank line and a horizontal rule. (2) `plan:54` and `plan:2228` cite
`spec:2455-2457` for `code-simplifier` running "after each accepted
task"; `spec:2455-2457` says "Subagents cannot dispatch agents, so the
conductor owns the `code-simplifier` dispatch at each pass **close**".
The per-task cadence is CLAUDE.md's git-conventions rule, which every
task's own commit invokes, so the practice is right and the citation does
not carry it. (3) `plan:822` cites `spec:210-227` for fork 3's ruled
values; fork 3 runs `spec:216-229`, so the range opens inside fork 2's
mail-transport paragraph and closes before the ruling sentence at
`spec:228-229`. The three durations R-5b10 derives are correct: fork 3
rules 15 minutes, 90 days and 180 days, and 2160h and 4320h are those in
hours. (4) Task 10's `keys.go` pointer correction (`plan:2164-2165`)
names `spec:1506` alone and misses R1's own copy of the same stale
pointer at `spec:284`, "(`keys.go:176-177`)"; the real lines are
`keys.go:165-166` for `DUBPLATE_CONTRIBUTORS_FILE` and `:159-162` for the
two Access rows. Folded: all four corrected, with `spec:284` added to
Task 0 item 5's report and to Task 10's drafted correction.

---

## Rejected findings

None retired. Every finding in both reviews survived re-verification.
One minor was promoted to a major by its merge (M-i), because the two
lenses had found different halves of the same dropped checklist bullet
and only the integrity lens's half named the argument that rests on it.
One major was re-grounded on different evidence and re-graded to a minor
(m8, correction 8), because the constraint the auth lens cited does not
reach a construction value. Six reviewer measurements were corrected
without moving their findings (corrections 2 through 7), and one
correction made a fold both cheaper and more precise than the review
proposed (correction 1).

---

## Verified sound

Do not re-litigate these. Each was checked by the fold editor against the
named artifact, and every mechanism claim was measured on go 1.27.1 with
the gate's own gitleaks binary.

1. The pass split is authorized and taken at the spec's own cut.
   `spec:2179-2184` reads "If 5b's plan exceeds fifteen tasks, split the
   pass rather than the tasks", with the cut point "between the token,
   session, and state records on one side and the roster, limiters, and
   seam on the other, which is where the risk changes from cryptographic
   construction to input validation", and instructs the conductor to name
   the follow-on 5b-ii and re-derive both ceilings. The plan does all
   three. The split rule is about tasks, and the plan's decomposition
   argument, that a task ends in an independently testable deliverable a
   reviewer could reject while accepting its neighbor, is the right
   reading. See m12 for the enumeration's own count.
2. The 5b-i and 5b-ii division of `spec:2168`'s seventeen items is
   complete and non-overlapping. 5b-i takes the identity leaf, token,
   session, nonce store, epoch store, revocation seam, reserved-class
   removal, wall-row retirement, the `Load` edit, the correlated
   validations, and the three supersession rewrites, eleven items. 5b-ii
   takes the roster store, consumption re-checks, limiters, origin check,
   verify seam, and `internal/notify`, six. Nothing in the spec's row is
   unassigned.
3. Every one of 5b-i's eight stated exit-gate assertions (`plan:340`)
   appears as an acceptance criterion in a named task: the replayed nonce
   across a restart and the MAC-failure-never-reaches-the-store row at
   Task 8 (`plan:1940-1947`); the session older than the epoch and the
   link minted before the bump at Task 9 (`plan:2046-2051`); refresh
   never moving the anchor and a session past the cap at Task 6
   (`plan:1743-1748`); the retention-below-expiry load refusal at Task 4
   (`plan:1526-1527`, whose landing site B2 fixes); `doctor` nonzero
   naming the selector at Task 3 (`plan:1428-1431`).
4. The two assertions of `spec:2168`'s 5b exit gate that belong to
   neither half are correctly re-routed with the reason for each: the
   roster-removal clause needs the consumption-time roster re-check,
   which R30 (`spec:637-648`) gives to 5b-ii, and the verify-subrequest
   clause grades the forward-auth verify endpoint, which `spec:2170`
   gives to 5d. Both are drafted as spec corrections (`plan:343-347`,
   `:2158-2161`). See m9 for the carrier the first one lacked.
5. Every exclusion carries a stop condition ending in stop-and-report:
   the pass scope statement (`plan:39-45`), the territory wall
   (`plan:451-466`), the no-roster/limiter/mail/origin/handler/page/unit
   rule (`plan:470-472`), the no-client-address rule (`plan:241-243`), the
   no-module-dependency rule (`plan:478-484`), and R-5b15's
   no-`internal/serve` rule (`plan:871-879`).
6. The ceiling's task arithmetic is internally correct. 5a's spend at
   `HISTORY:521-525` reads exactly as the plan's basis table states it
   and its nine line items sum to 3.90M. The chain rate 1.68M over eight
   tasks is 0.21M, and two more tasks at that rate is the +0.42M line.
   3.90 + 0.42 + 0.10 + 0.02 + 0.05 = 4.49M. Eighty percent of 5.0M is
   4.0M and of 4.5M is 3.6M, and the flag-inside-the-chain argument
   against 4.5M is right, with rung 1b's overrun (`spec:2225-2227`) the
   recorded case behind it. Only the reservation baseline is wrong (m3).
7. The rung 1b registry contract is quoted faithfully and applied
   correctly. `HISTORY:1496-1507` says the wall becomes registry
   contents, `Load` takes the registry it enforces against, a rung
   declares its keys in its own file inside the package that owns them,
   no lane adds a name to `internal/config/keys.go`, and a lane removes
   only its own reserved class in the same commit that declares the class
   in its own package. The plan's paraphrase at `:96-104` is accurate, and
   its claim that 5b-i exercises the removal half for a pending wall row
   for the first time is right: the contract as written covers reserved
   classes alone, R5 (`spec:345-347`) authorizes the wall-row extension
   explicitly, and `spec:1506-1514` authorizes the supersession-message
   extension.
8. R-5b2's argument holds in both directions. Declaring inside
   `internal/auth` would violate `spec:1471-1479`, declaring inside
   `internal/config` would violate `HISTORY:1501-1505`, and
   `cmd/dubplate-server/serve.go:11-25` is the in-tree precedent the
   resolution cites.
9. R-5b4's generic form is a genuine improvement over R29's literal
   wording, and the plan flags it as a correction rather than absorbing
   it. R29 (`spec:630-631`) does ask for "a `w.require` call for
   `DUBPLATE_IDENTITY_SOURCE`", and "require every owned-now wall row"
   discharges the same requirement while closing risk 11
   (`spec:2409-2414`) for every later rung and keeping rung 5's
   vocabulary out of the config core. Rung 1's four rows are unaffected
   because they are the rows `NewRegistry` seeds. The plan also
   anticipates the double-report hazard with a criterion that a missing
   registered row is named exactly once.
10. R29's premises hold against the tree as the spec states them. `Load`
    hard-codes four owned requires plus the generic integration loop
    (`internal/config/config.go:248-287`), `Config` carries no generic
    wall accessor (`config.go:66-118`), and `doctor` prints "ok" for any
    row `Load` never mentions (`doctor.go:34`).
11. R-5b6's threshold argument is verified in both halves.
    `cmd/dubplate-server/doctor.go:142-148` prints a `KindString`
    threshold's effective value through `config.ParseString`, which would
    print a rotation secret on an operator's terminal, and
    `internal/config/registry.go:320-323` panics on an empty
    `KindString` default, which an unset-by-default rotation key needs.
    The source's optional set is the right home on both counts.
12. R-5b7's template mechanism is correct. `templateKeys`
    (`templatekeys_test.go:124-149`) skips any line not beginning with
    `DUBPLATE_` after trimming, so a commented pending row would not
    satisfy `TestTemplateKeysCoverEveryOwnedNowRow`
    (`templatekeys_test.go:74-89`), and a bare `NAME=` row would abort
    `deploy.sh`'s render. A literal is the only row state satisfying
    both, and `env/dubplate-server.env.template:8-16` carries the three
    row states with the literal state at `:12-13`. The contradiction with
    `spec:2205` and `spec:2170` is real and correctly routed to the close.
    See m4 for the count inside the argument and M-i for the pre-task
    bullet that verifies it.
13. R-5b8's floor is measured correctly against RFC 2104 section 3.
    `hmac.New(sha256.New, key)` takes the raw configured string's bytes
    as the key and produces a 32-byte tag, measured with a 41-byte raw
    string key, so counting bytes of the untransformed value is the right
    measurement and a base64 rendering of 32 random bytes clears it with
    room to spare.
14. R-5b9's cookie name round-trips through `net/http` unchanged, and
    Task 6 is right to assert each attribute individually rather than
    lean on the stdlib. Measured, `http.SetCookie` emits
    `__Host-dubplate_session=v1.abc; Path=/; HttpOnly; Secure;
    SameSite=Lax` and `req.Cookie` reads it back, while
    `http.Cookie.Valid()` returns nil for a `__Host-` cookie with
    `Secure` false, with a `Domain` set, and with `Path` `/sub`, and
    `String()` emits the `Domain` regardless. Leaning on `Valid()` would
    not have fired Task 6's mutation 2.
15. R-5b10's duration arithmetic is right: 2160h is 90 days and 4320h is
    180 days, matching product fork 3's ruled values. See m16 for the
    citation.
16. R-5b11's single `auth` leaf with `nonce/` and `epoch/` beneath it and
    R-5b12's untruncated 64-character hex digest match R34
    (`spec:694-704`) and `spec:1691-1697`'s single-static-path argument,
    and `internal/engine/shadow/store.go` carries the shape the plan says
    it adopts (`storeDirName:41`, `StoreDir:46`, `OpenStore:53`,
    `validID:66`).
17. R-5b13's argument holds: R32 bounds the token's exposure with a
    consumption route that logs its path and never `RawQuery`, and
    `internal/serve/serve.go:446`'s `boundedPath` is the refusal record a
    path-borne token would reach.
18. The three toolchain facts Tasks 7 and 8 rest on are measured true on
    go 1.27.1. `os.OpenFile` with `O_CREATE|O_EXCL` returns an error
    satisfying `errors.Is(err, os.ErrExist)` for an existing file;
    `(*os.File).Sync` on a directory handle returns nil on this
    filesystem; `crypto/subtle.ConstantTimeCompare` is present, returns 1
    on equal inputs and 0 on differing lengths. `os.MkdirAll(dir, 0700)`
    yields `drwx------` and a file created 0600 reports `-rw-------`, so
    Task 7's permission assertions are writable as stated. `go.mod:3`
    carries `go 1.27.0` and `go.mod:5-10` carries cobra plus two
    indirects, so the no-new-dependency constraint is checkable.
19. R11's `atomicfile` argument holds and the plan's re-resolution of the
    spec's stale pointer is right. `WriteFile` is at
    `internal/atomicfile/atomicfile.go:16`, `WriteStream` at `:55-91`,
    the directory open at `:82` and its `Sync` at `:87-89`, so
    `WriteStream` is the only entry point that fsyncs the containing
    directory after the rename.
20. The link builder's stdlib shape works as Task 5 assumes. Measured,
    `url.JoinPath` plus `Query().Encode()` builds
    `https://example.test/auth/consume?t=AAAA%2B%2F%3Dabc` from a
    configured origin and preserves a non-root origin path
    (`https://example.test/dubplate/auth/consume?t=xyz`).
21. Task 3's mutation 1 is sound. `Registry.claim`
    (`registry.go:119-124`) panics across families through
    `panicDuplicate` (`registry.go:95-97`), so declaring the twelve names
    without deleting the reserved class panics at registration on the
    eight colliding keys.
22. The registry surface is exactly as Task 0 item 6 expects:
    `AddWallRow` at `registry.go:138` with the status-row carve-out
    documented at `:131-137`, `AddReservedClass` at `:163`,
    `AddIntegrationClass` at `:211`, `AddThreshold` at `:267`,
    `requirePrefix` at `:106`, `claim` at `:119`, `panicDuplicate` at
    `:95`, `validateThresholdDefault` at `:306`; `thresholds.go`'s kind
    constants at `:16-24`, `Threshold` at `:45`, `thresholdOf` at `:112`,
    `ParseString` at `:128`; `IntegrationClass.Enabled` at `keys.go:227`.
23. The `keys.go` facts the pass turns on are as cited: the pending row at
    `keys.go:24`, the `auth and identity` reserved class at
    `keys.go:59-73` with eight keys and rung 5, the two Access
    supersession messages at `keys.go:159-162`, and
    `DUBPLATE_CONTRIBUTORS_FILE`'s at `keys.go:165-166`. The plan's
    correction of `spec:1506`'s stale pointers is right. See m16 for the
    third copy at `spec:284`.
24. The two merge carries are real and correctly located.
    `cmd/dubplate-server/serve_test.go:19-25` restates `newRootCmd`'s
    register calls by hand, and `templatekeys_test.go:95-100` returns
    "owned wall" for any wall row including a `Pending` one, where
    `ownedNow` (`keys.go:36-43`) would refuse it. The ruling on each is
    argued rather than asserted, and the third carry's forwarding to 5d
    carries R-G's reason to be in the file.
25. The lane, worktree, implementer and merge-ritual rules are right and
    agree with the concurrent plan. `spec:2168`'s tree column for 5b is
    "Worktree"; `.claude/agents/dubplate-implementer.md` exists; the rung
    8c plan independently names the same worktree path and lane
    numbering; and the two territories are disjoint, since 8c's own wall
    excludes `internal/config`, `cmd/`, `env/`, `go.mod` and
    `.gitleaksignore`. The rules are stated in three places
    (`plan:57-63`, `:443-450`, `:570-578`) and match CLAUDE.md's Lanes
    section and `spec:2273-2275`'s second-code-slot rule. See m13 for
    where the lane registry actually lives.
26. The prior-art record is cited accurately and every pointer resolves:
    `PRA:23-44` the verdict table, `:46-54` the stateless-cookie
    argument, `:55-73` gorilla/sessions, `:75-93` alexedwards/scs,
    `:105-113` the construction-is-specified argument, `:114-140`
    golang-jwt/jwt/v5, `:141-155` o1egl/paseto, `:156-175`
    aidanwoods.dev/go-paseto, `:568-578` Gitea as the road not taken,
    `:655-677` the provisional rulings, `:678-717` the evidence notes.
    Every version, date and importer figure the plan repeats matches
    `spec:1870-1876`'s own table. The adoptions of
    `internal/engine/shadow/store.go`'s shape at Task 7 and of
    `atomicfile.WriteStream` at Task 9, and the decline of both
    `atomicfile` entry points for the nonce record, each carry the reason
    R11 states.
27. The 2.0 hook accounting is right and no task builds a 2.0 feature
    early. `S20:367-369` is hook 5 exactly and `S20:245-246` is the
    valid-after-epoch requirement exactly. Task 1 lands hook 5's
    separate-principal half, Task 7 lands the digest derivation that
    makes rung 6's re-key one function (R34), and Task 9 lands the
    read-plus-notification seam R35 requires. Hook 6 stays open by
    naming no driver. Nothing mints a surrogate id, no credential store
    is built, and no app-password surface appears.
28. The threat model given to the in-pass reviewer is faithful to
    `spec:1936-1952` and correctly scoped, and withholding the task
    decomposition matches `spec:2471-2473` and the 5a plan review's
    Verified sound item 30. Telling the reviewer that limiting,
    enumeration, origin checking, mail injection and roster validation
    are 5b-ii's is right. See B1 for the one claim in the list the code
    would not have held.
29. The 5a records are read correctly and none is reopened. The 5a plan
    review's Verified sound list is at `:705-860` with thirty items, and
    the items the plan names say what it says they say, including the
    re-resolution of item 11's `HISTORY` pointer to `:1491-1507`. The 5a
    security review's dispositions are read correctly: finding 7 routed
    to this pass (`plan:124-127`, matching
    `research/2026-09-08-rung-5a-security-review.md:82-88`), findings 16,
    22, 23 and 24 left to 5c, 5d and the wall contract. The running
    coverage-gap total of sixty-three after 5a matches `HISTORY:325-329`
    and `docs/STATUS.md:36-38`, whose own breakdown sums to sixty-three.
30. The salvage claims are true. There is no `_salvage/internal/auth`;
    `_salvage/internal/access/access.go` is the Cloudflare Access JWT
    verifier the plan describes; `_salvage/internal/config/contributors.go`
    is the fail-open roster loader R2 supersedes, correctly routed to
    5b-ii. Flagging the phantom-adoption risk in Task 0 item 19 is the
    right instinct.
31. Deliverable counts are stated per task and none exceeds four (Task 0:
    1, T1: 2, T2: 4, T3: 4, T4: 3, T5: 3, T6: 3, T7: 3, T8: 3, T9: 3,
    T10: 4), with Tasks 2 and 3 named as the two to watch, matching
    CLAUDE.md's pass-sizing rule. The plan carries no em dash and no en
    dash in 2,342 lines, and no calendar date, launch date, pace or weeks
    estimate appears. The pass names no secret value, mints nothing, and
    adds exactly one literal template row.
