# admin-api-1 plan review: contract and criteria

Target: `docs/superpowers/plans/2026-10-11-admin-api-1.md` at `9c58bcc`. Lens: is every promise testable, does each
criterion name its fixture and the reason it would fail, can any pass vacuously, and does every design criterion
(AC1 to AC22) and owed fold item (V4 to V7, over-ceremony 2) land in a task with a proof. Code was read in this
worktree (unchanged from `551bf4f` for every pointer below). Plan lines are `plan:N`.

## Coverage verdict

Every AC lands in at least one task, and every owed fold item lands: V4 in T6 (`plan:503-505`), V5 in P7 and T5
(`plan:200-205`, `:453-454`, `:480`), V6 and V7 in T3 (`plan:380-385`), over-ceremony 2 in P8 and T8's third mutation
(`plan:206-212`, `:585`). The member-cap removal (P12), the moot contested-surfaces table, the unplanned frozen
errata, and the CH:171-172 edit in T5 are consistent with the spec and the adapter's errata rule; CH:171-172 still
reads "a sensitive change notifies the admin address", so T5's grep check fails at the start commit as it should.

| AC | Task(s) | Named proof | Gap |
|---|---|---|---|
| 1 | T6 | parity, keyword walk | code and detail enumeration parity has no proof (F4) |
| 2 | T6 | three named proofs | one mutation names the wrong proof (F9) |
| 3 | T6 | none | F12 |
| 4 | T6 | api-path states, plain-text mutation | none |
| 5 | T1, T6 | authorizer fakes, 401 fixtures | bearer, actor-field, log, and channel clauses unfixtured (F7) |
| 6 | T1, T5, T6 | principals differ, rename fixture | log clause unfixtured (F7) |
| 7 | T1, T6 | V4 marshal and property-set test | `start_backup` conflict answer unchecked (F12) |
| 8 | T2 | per-family panics, sensitive pin | none |
| 9 | T5, T6 | sentinel scan | sentinel set infeasible as worded (F11) |
| 10, 11 | T4 | five fixtures | wall key, unregistered key, env-equal no-op, env-set reset unfixtured (F6) |
| 12 | T5, T6 | anchor fixtures, refresh fixture | the five gated calls and two ungated counter-cases not enumerated (F6) |
| 13 | T4, T5 | three-admin mailer, failing and blocking mailers | CLI and seed-rule "by construction", non-sensitive write unfixtured (F10) |
| 14 | T7 | four fixtures | `serve`'s older-store message unfixtured (F6) |
| 15 | T3, T6, T7 | AC15 fixtures | the removal's own bump has no killing proof (F2) |
| 16 | T3, T5, T7 | held-lock concurrency | the in-flight actor re-read has no killing proof (F1) |
| 17 | T5, T6, T11 | admin-field body, `roster` 404s | none |
| 18 | T5, T6 by reference | none enumerated | F6 |
| 19 | T9 | rendered rows | rows do not say what they intercept (F13); deck gate vacuity (F5) |
| 20 | T10 | rendered rows | "every key" row and the restart-key line unproved (F13) |
| 21 | T6, T8 | standings fixtures, forbidden fixture | runbook clauses unchecked, normalization state (F14) |
| 22 | T4, T7, T11, Close | the riders' proofs | #25 fixture underspecified (F8) |

## Findings, ranked by consequence

### F1. major, correctness. The in-flight demotion mutant survives every named fixture

- **Location:** `plan:405` (T3 fixture "An actor demoted before a removal, a demotion, or a cookie-channel add is
  refused"), `plan:410-411` (mutation "read the actor's flag by address or outside the `Update`"), `plan:466` (T5's
  "AC16's settings-write actor clause").
- **Defect:** the spec's reason for the in-write re-read is "so a demotion is final for writes already in flight"
  (spec:252-253). Every named fixture demotes the actor *before* the call, which a read made before `Update` also
  sees, so the "outside the `Update`" mutant passes. The plan's own convention then makes this a finding against the
  test mid-task (`plan:102-104`), which is a fix the plan can name now. The "by address" half has no killing fixture
  either: `ops.Actor` carries no address (`plan:318-320`), so the address never differs from the id's row in any
  fixture.
- **Fold:** name the in-flight fixture on the AC16 pattern already in the plan: a second connection holds `BEGIN
  IMMEDIATE` with the actor's demotion uncommitted, the actor's removal (T3) and a settings write (T5) start, the test
  commits the demotion after a fixed hold well inside the busy timeout, and each write answers `forbidden` with
  nothing written. Drop "by address" from the mutation, since the type makes it inexpressible, or name a fixture that
  moves the actor's address by direct store write.

### F2. major, correctness. Nothing proves the removal's own epoch bump

- **Location:** `plan:397-400` (AC15 fixtures), `plan:408-411` (T3 mutations).
- **Defect:** the tombstone already refuses a removed member's session and a link consumption (lookup at
  `roster.go:257`; `handlers.go:420` checks roster and epoch before the insert), and the re-add bump revokes every
  credential anchored before it. So a mutant that skips the removal's bump passes every named fixture, including
  "stays refused after the re-add". The spec keeps the removal bump on purpose (spec:340-342) and makes removal
  `Bump`'s first production caller (spec:335). The mutation list has "skip the re-add bump" but no "skip the removal
  bump".
- **Fold:** after the removal commits, the AC15 fixture reads the address's epoch through `EpochStore.Read` and
  asserts it equals the injected removal time; add "skip the removal bump" to T3's mutations.

### F3. major, correctness. The `busy` fixture reaches only one of the two mapped errors

- **Location:** `plan:453-454`, `plan:480` ("a held write lock under a short deadline"), `plan:484` (mutation "map
  `ErrBusy` to `internal`"), `plan:285` (states row).
- **Defect:** the two errors come from different holders. `internal/engine/store/write_test.go:573-585` gets
  `context.DeadlineExceeded` only when the *in-process* write connection is held; `:588-606` records, measured, that
  "SQLite's busy handler does not stop for the driver's interrupt, so a deadline shorter than the busy timeout still
  waits the full busy timeout on another process's write lock", which yields `ErrBusy`. "A held write lock under a
  short deadline" names neither holder. Whichever the implementer builds, the other branch has no fixture, and a
  mutant that drops it survives. The unfixtured branch may be the common one: two concurrent admin writes inside
  `serve` contend for the in-process connection and answer `DeadlineExceeded`.
- **Fold:** quote `write_test.go:588-590` under "Behaviors this plan relies on", and split the fixture: (a) a second
  handle holding a transaction with a short busy timeout, which answers `ErrBusy`; (b) the same handle's write
  connection held, under a short deadline, which answers `DeadlineExceeded`. Each runs against one settings write and
  one member write. Add the mutation "map `DeadlineExceeded` to `internal`".

### F4. major, correctness. P7's detail list is cited from a list that is wrong for it, and no test holds the document's enumerations to Go

- **Location:** `plan:200-205` (P7), `plan:496-502` (T6 contract tests), AC1.
- **Defect:** P7 defines the detail enumeration as "every roster-write outcome the members page keys on
  (`+page.svelte:46-54`)". That map holds `duplicate_principal`, which no Go code emits and which T9 removes under #53,
  and lacks `invalid_address` (`handlers.go:57`, mapped for adds at `admin.go:324-325`), which T9 needs for #17's focus
  row (`plan:593-594`). An implementer following P7 literally ships a dead detail and no `invalid_address`. Separately,
  AC1 says the document describes "every wire code and detail it can answer", but the only parity test is on
  operations; a constant added in Go and missing from the document's enum passes unless some test happens to produce
  that refusal.
- **Fold:** P7 lists the values outright: `duplicate_address`, `duplicate_username`, `invalid_address`,
  `invalid_display_name`, `blank_display_name`, `invalid_username`, `missing_username`, `last_admin`,
  `environment_sets`, `busy`. T6 adds one parity test: the document's `code` enum equals `ops.Codes()` and its
  `detail` enum equals the `Detail*` constant set, both ways. Mutation: add a constant the document omits.

### F5. major, correctness. The deck-coverage gate loses its vacuity guard, and the new parser never gets one

- **Location:** `plan:598-600` (T9), `plan:657-658` (T11).
- **Defect:** `deck-coverage.ts:155-167` guards the outcome parser with a must-have list and a floor of 21, so a
  broken regex fails rather than requiring nothing. T9 adds a second regex for `Detail<Name>` with no guard, and T11
  deletes the must-have list and the floor outright. After T11, neither the remaining `handlers.go` parse nor the new
  details parse can fail on a zero-match pattern, so the gate passes vacuously.
- **Fold:** T9 gives the details parse its own must-haves (`last_admin`, `busy`, `invalid_address`) and an exact
  count of P7's values. T11 narrows the existing guard to `handlers.go` (`signed_out` and its own count) rather than
  deleting it. Mutation in each: break the regex and see the gate fail.

### F6. major, correctness. No ledger maps the states rows and AC clauses to fixtures, and several have none

- **Location:** `plan:268-291` (States), `plan:466-467` (T5: "every states-table row reachable at the seam"),
  `plan:519-520` (T6: "AC18 for every HTTP-reachable state").
- **Defect:** the adapter requires that "each proof names the fixture state it needs". AC18 is assigned by
  reachability, not by row, so a row both tasks consider the other's goes unproved with every acceptance line still
  true. Rows and clauses with no named fixture in any task:
  - remove or demote of an unknown id or a removed member answers `not_found` (spec:491);
  - demoting a non-admin writes nothing and answers success (spec:495);
  - each add refusal's `detail` under `conflict` or `invalid_request` (spec:498);
  - `set_setting` on a wall key and on an unregistered key (spec:499; AC10);
  - `set_setting` equal to the environment's value writes nothing (spec:501; AC10);
  - `reset_setting` of an `admin` row the environment sets answers `environment_sets` (spec:502; AC11);
  - the five gated calls and the two stale-but-ungated counter-cases (AC12; T5 names only "inside and outside the
    window" and one mutation, for `add_member`);
  - a roster-lookup failure during verification answers 500 (`plan:283`; only the setting-read failure is named);
  - `auth.Open` refusing `member add`, `remove`, and `demote`, and `member list` opening no auth leaf (`plan:288`;
    neither T3 nor T7 names it);
  - `serve`'s pre-open refusal of an older store holding an `admin` row names the recovery (AC14; T7's older-store
    fixture is the verb's, not `serve`'s);
  - `doctor` over a store older than the users table (`plan:291`; T8 names absent and unvouched only);
  - a target removed between the CLI's read and write (`plan:540-541`).
- **Fold:** one table in the plan, one row per states row (the spec's 21 and the plan's 10) and per AC clause not
  already fixtured, naming its task, fixture, and the mutation that kills it. Most rows are one table-test case in a
  test the task already writes.

### F7. minor, correctness. AC5's and AC6's mount clauses have no failing test

- **Location:** `plan:494-495`, `plan:518-519`, `plan:530-533`.
- **Defect:** "read an `Authorization` header" is listed as a mutation, but no fixture sends one, so a mount that
  reads it and ignores it, or admits on it, is untested. "A body carrying an actor field is ignored", the log record's
  actor, principal, channel, and `client_ip` fields, and the mount's channel being `ChannelCookie` (AC5) are named in
  the outcome and acceptance but have no fixture. The old plan's "a test reads the record's principal field" was
  dropped.
- **Fold:** T6 fixtures: a request with a well-formed `Authorization: Bearer` header and no cookie answers 401, and
  one with a valid cookie and a bogus bearer header is admitted as the cookie's member; a `set_setting` body carrying
  `"actor":"member:<other id>"` records the session's actor and channel `cookie` in `settings_audit`; a captured
  `slog` handler asserts one record per admitted operation with the four fields, the principal under its own key.

### F8. minor, correctness. The #25 fixture cannot be written as worded for most kinds

- **Location:** `plan:433-434`.
- **Defect:** "an `admin` row whose value does not parse, whose refusal wraps the parse error (`errors.Is`)" names
  no kind and no target. `time.ParseDuration` returns a fresh `errors.New` value, so a duration key gives `errors.Is`
  nothing to match. `strconv.Atoi` and `ParseFloat` return a `*NumError` whose `Err` is `strconv.ErrSyntax`. Today's
  code wraps `s.Validate(stored.Value)` (`evaluate.go:265`), and `boundedValidator`'s message wraps no parse error
  (`internal/config/settings.go:182-186`), so `errors.Is(err, strconv.ErrSyntax)` fails at the start commit and passes
  after the fix.
- **Fold:** name an int-kind key and `strconv.ErrSyntax`, and quote `go doc strconv.NumError` under the behaviors.

### F9. minor, correctness. One T6 mutation names the wrong proof

- **Location:** `plan:532` ("leave a response object open in the loaded copy (the rename proof passes, so the test
  fails)").
- **Defect:** the rename proof uses a `required` field (AC2), which fails against an open schema too: the renamed
  field is missing. The proof that an open object silences is "a Go field the document omits" (spec:71-74).
- **Fold:** the annotation reads "the omitted-Go-field proof passes, so the test fails".

### F10. minor, correctness. Two T4 mutations and AC13's negative clauses have no killing fixture

- **Location:** `plan:420-421` ("AC13's command-line and seed-rule clauses hold by construction"), `plan:437`
  (mutation "write the audit row after the commit"), `plan:476`.
- **Defect:** an audit row written in a second `Update` still exists in every named fixture, so that mutant passes.
  "By construction" is a true argument but not a test; nothing stops a later edit from threading a mailer into
  `internal/settings`. No T5 fixture runs a non-sensitive write against the recording mailer.
- **Fold:** an audit insert made to fail (a `settings_audit` trigger in the test store, or a dropped table) rolls the
  setting back; the layering rule (`plan:109-113`) gains "`internal/settings` imports no `internal/notify`", checked
  where the other import rules are; T5's three-admin fixture adds a non-sensitive write and a non-sensitive reset,
  each with zero sends.

### F11. minor, correctness. AC9's sentinel fixture is infeasible for some rows and collides with AC21's

- **Location:** `plan:522`.
- **Defect:** "sentinel values on every wall and integration row" cannot hold for rows whose value set is closed:
  `DUBPLATE_IDENTITY_SOURCE` (`auth.go:307`) and the integration flags. `DUBPLATE_ADMIN_EMAIL`'s sentinel shows up in
  `list_members` whenever it names a member, which T6's own AC21 fixture (`plan:528`) makes it do. An implementer
  narrows the set silently, which is the vacuity this lens looks for.
- **Fold:** name the sentinel per row (a unique path, site name, address that no member holds, port, and host), list
  the closed-set rows excluded with the reason, and keep the AC9 server separate from the AC21 one.

### F12. minor, correctness. Two document-content clauses have no check

- **Location:** `plan:496-498`; AC3, AC7.
- **Defect:** the version policy text (AC3) and `start_backup`'s `conflict` response (AC7) are document facts that no
  contract test reads; parity and the keyword walk pass without either.
- **Fold:** the contract test asserts `start_backup` declares a 409 `conflict` response, and the `info.description`
  names the additive policy's four enumerations (a substring check is enough).

### F13. minor, correctness. Several rendered rows do not say what they intercept or why they cannot pass empty

- **Location:** `plan:608-619` (T9), `plan:637-647` (T10), `plan:626-628`.
- **Defect:** the plan's own convention says each row "says what it intercepts and what it reaches for real"
  (`plan:105-108`). "Each refusal's copy", "each add refusal moves focus", and "the confirm step blocks the call" do
  not, and the last needs a request count of zero before confirming to be non-vacuous. "Every key with its description
  and mark" has no expected list, so a page that renders the API's answer passes whatever the API omits. T10's outcome
  promises the `restart` key's next-start line (spec:211-212) with no row, and P5's keep-the-pending-change and new-tab
  behavior on the members page (`plan:191-193`) has no T9 row.
- **Fold:** each row names its interception; the confirm row counts requests; the key row compares against a list the
  Go host passes from the production registry; add a `restart`-key row and a T9 re-authentication row asserting the
  form still holds the entry and the link has `target="_blank"`.

### F14. minor, correctness. T8 leaves the runbook unchecked and the address-variant state unnamed

- **Location:** `plan:572-585`.
- **Defect:** T5's CH edit carries a grep check; T8's three runbook edits carry none. `Inspect` compares the
  configured address against stored rows, which `identity.NormalizeAddress` wrote (`roster.go:147`); `requireEmail`
  lowercases (`config.go:216`), but no fixture proves a mixed-case `DUBPLATE_ADMIN_EMAIL` naming a live admin reports
  ok rather than "no member".
- **Fold:** joined-line greps on `docs/server-runbook.md` for `--admin` in step 10 and the restore line, as T5 does;
  one fixture with a mixed-case address naming a live admin.

### F15. minor, correctness. The refresh's counter-case and the invalid-identity state are unnamed

- **Location:** `plan:177-179`, `plan:181-184`, `plan:284`.
- **Defect:** `RequireAdmin` refreshes only when `Sec-Fetch-Site` is `same-origin` (`admin.go:113`). The plan names
  the same-origin case only, so a verifier that always refreshes passes. The `!v.Identity.Valid()` branch
  (`admin.go:120-121`) is a refusal that P3's list of `gateRefusals` errors does not cover, so its answer (401 under
  the spec's "no valid session") is unstated.
- **Fold:** quote `admin.go:113` under the behaviors; add a states row and fixture for a request past the refresh
  point without `Sec-Fetch-Site: same-origin` (admitted, no `Set-Cookie`) and one for a verified but invalid identity
  (401).

### F16. minor, correctness. T5 promises notice content the spec does not set, and nothing asserts it

- **Location:** `plan:450-452`.
- **Defect:** "naming the key, the actor's principal, and the old and new values" comes from the old plan
  (`2026-10-10-rung-14a-admin-api.md:574`), not the spec (spec:280-285). No fixture reads the message.
- **Fold:** keep it, since it is what a notice is for, and assert the four fields in the recording mailer's captured
  message; or drop the clause.

### F17. minor, correctness. The plan defers the quoted-behavior gap to execution

- **Location:** `plan:265-266`.
- **Defect:** "A task that relies on another library or stdlib behavior quotes it ... in its PR notes before the code
  lands" moves the adapter's first gap from plan approval to execution. F3, F8, and F15 are three behaviors the
  fixtures already depend on and the section lacks.
- **Fold:** add the three quotes (F3, F8, F15) and keep the catch-all line for behaviors discovered mid-task.

### F18. minor, optional. Over-ceremony: the close render matrix

- **Location:** `plan:39-40`, `plan:680-681`.
- **Defect:** "in both palettes at phone and desktop widths, in each refusal state" is two pages, two palettes, two
  widths, and about a dozen refusal states (five shared, seven add refusals, the settings refusals), so 60 or more
  screenshots read in the main loop. At roughly 1.5k tokens an image plus the read, that is on the order of 100k
  tokens and 20 to 30 minutes of the Close's 100. Refusal copy renders in one alert component, so the base states
  already prove palette and width.
- **Fold:** every base state (list, empty, form, settings list, environment-set key) in the full matrix; each
  refusal state once, at phone width in one palette, plus the re-authentication flow in both palettes.

### F19. minor, optional. Plan departs from the adapter's fast-lane text

- **Location:** `plan:87-97` against the adapter's "Fast lane: ... Pass no `--base`".
- **Defect:** the departure is reasoned (T0's machinery files promote every later scoped run), but the adapter wins
  where it disagrees with process (pass-core), and the plan does not record the departure as such. This lens defers
  the mechanics to the mechanics review.
- **Fold:** one sentence naming the departure and its reason, plus a "Found, not done" item for the adapter text.

## Scope check: does any task promise beyond the design?

- T5's notice content (F16).
- T8's added runbook line naming the three new verbs' `systemd-run` form (`plan:574-576`): one line, needed for the
  verbs to run under the unit's grants, within the spec's runbook scope.
- T9's #17 (`plan:593-594`): AC19's "each add refusal still moves focus" covers `invalid_address`, so it is required,
  not widened.
- T6's route inventory (`plan:506-508`) comes from the contested-surfaces table the plan calls moot (spec:530); its
  obligation also stands on PS A2, which the plan cites, so the scope is right.

No owner forks: every fold above is a method or proof choice.

## Counts

0 blockers, 6 major (F1 to F6), 13 minor (F7 to F19), of which 17 are `correctness` and 2 `optional` (F18, F19). One
over-ceremony item (F18). 0 owner forks.
