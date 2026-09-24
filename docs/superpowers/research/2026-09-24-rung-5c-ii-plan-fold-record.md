# Rung 5c-ii plan fold record

Target: `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`,
revision 1 (`3ebfe3c`) to revision 2 (folded, awaiting verification).
Reviews folded (`140e3cd`): the contract lens (C1 through C19), the
mechanics lens (M1 through M4, m1 through m11), and the risk lens (R1
through R13), all under `docs/superpowers/research/2026-09-24-rung-5c-ii-plan-review-*.md`.

Each finding was checked against the tree at `140e3cd` before it was
folded. The checks that decided a disposition:

- C1: `internal/config/registry.go:322-325` panics on an empty
  `KindString` default, and `config.go:182-187` reads an empty value as
  unset. Confirmed.
- C4: every validation refusal in `roster.go` (`checkAddition`, `addTo`)
  wraps `ErrRosterRecordInvalid` and the specific sentinel. Confirmed.
- R4: `linkTokenParam = "t"` (`internal/auth/token.go:64`); `LinkURL` is
  exported. Confirmed.
- m8 and R10: `actorRefusesRune`'s comment claims parity with
  `NormalizeAddress`, and `validateActor`'s comment makes the same claim
  (the fold covers both).
- m11: `newServeShell` calls `serve.NewShell(prefix, strip, nil)`
  (`cmd/dubplate-server/serve.go:215`). Confirmed.
- C10: kit emits the boot script as a bare `<script>` in hash mode
  (`render.js:586-588`, read from the scratchpad tarball). Confirmed.
- m3: the probe build's CSS carries `[data-theme=dubplate-gruvbox]`
  unquoted, and its stylesheet link reads `<link href="..." rel="stylesheet">`.
  Confirmed from the scratchpad build.
- C18: the magic-link source, the only identity source, lists
  `DUBPLATE_PUBLIC_URL` as Required (`auth.go:50`). Confirmed.
- Task 2's new mutation 7: 13 runes are both `Cf` and
  `Prepended_Concatenation_Mark`, measured, so dropping that term
  changes the derived set's size.

## Counts

47 finding IDs. **43 folded** (two of them with one part refused), **0
refused whole**, **4 owner forks** (Rulings 4 through 7). m4 is folded
into Ruling 2's recommended answer, not raised as its own fork.

**Why nothing is refused whole.** The risk and mechanics lenses each
cleared the core design ("attacked and found sound", "what the probes
confirmed"). Nearly every remaining finding is a missing named state or
a missing proof row inside a test file the task already writes. Each
such fix costs one table row, and each closes a defect that would
otherwise pass the gate. The two parts refused below are the only ones
whose cost exceeded the risk they remove.

## Convergent defects, folded once

| Root | IDs | Where it landed |
| --- | --- | --- |
| Content-Type compared through `mime.ParseMediaType`, with a charset row | C2, R2, M4 | R-5cii4 "The Content-Type check parses the header"; Task 1 outcome table, Content-Type rows, mutation 13 |
| Stylesheet anchor with zero or many kit links, insert after the last | C7, R6 | Task 6 "The tier links" and "Anchor states", mutation 6; Task 7 G7 row (no `_app/` sheet after `branding.css`) |
| Task 5's links pointed at routes that arrive in Task 6 | C6, M1 | Link insertion moved to Task 6 with the routes; Task 5 stamps the theme only; Task 6 extends the prefix proof; execution model's split note |
| Sub-path, method, and directory states, and the `..` 400 | C12, C14, R3, m5, m6 | New R-5cii21; Task 5 states; Task 6 route criterion; Task 2 `roster/x` row; Task 7 `members/` rows |
| n5's pre-existing rows and the false comment outside Task 2's files | C11, R10, m8 | R-5cii5 (regression rows, size 4174, comment); Task 2 Files and criteria; package-territory wall |
| Token-log test's parameter and capture channel | R4, C16, m11 | Task 7 "The minted link routes and never logs" and "The log channel"; Task 1 POST-refusal log row |
| The route's existence is public | m10, R13 | R-5cii19 "The gate hides the response, not the route"; Goal; exit gate item 1; Task 8 handoff |
| Theme selector form and composed-name matching | C8, m3 | R-5cii12 "The composed-name check"; Task 4 criterion; Task 5 palette states, mutation 5; Task 0 item 4 |

## Dispositions

### Contract lens

| ID | Disposition | Plan location |
| --- | --- | --- |
| C1 | Owner fork, **Ruling 4** (`none` sentinel defaults). The unforked part is folded: an empty match value resolves to the default and is tested, and the "empty included → refuses" row is gone. | Ruling 4; R-5cii12 "Why no default is empty", "An empty value resolves"; Task 5 states |
| C2 | Folded (convergent, Content-Type) | R-5cii4; Task 1 |
| C3 | Folded: the fields come from `r.PostForm`, and a query field key answers `malformed` | R-5cii4 "The three fields"; privileged-fields row; Task 1 query-field row, mutation 15 |
| C4 | Folded: the sentinel mapping is named, and two `failed` fixtures are built on the inner handler | R-5cii4 "The error mapping"; Task 1 table, "The `failed` rows", mutation 14 |
| C5 | Folded: GET, HEAD, and PUT criteria, plus the store-failure row on the inner handler | Task 1 acceptance |
| C6 | Folded (convergent, task order) | Tasks 5 and 6 |
| C7 | Folded (convergent, anchor) | Task 6 |
| C8 | Folded (convergent, selector) | R-5cii12; Tasks 4 and 5 |
| C9 | Folded: G6 and G7 run against the placeholder-only tree and first assert the three-document set; G3's built half is seeded | Task 7 acceptance; Task 3 G3 sentence and seeds |
| C10 | Folded, one part refused. Folded: the quote of kit's emission; the extractor refuses an unrecognized script opening; G6's scanner shares no function with the extractor. **Refused:** a G6 fixture document carrying attributed or uppercase tags. With startup refusing any script opening it does not recognize, such a tag cannot reach a served document, so the fixture would test the test for a state that cannot arrive. | R-5cii7; Task 7 G6 row, mutation 10 |
| C11 | Folded (convergent, n5) | R-5cii5; Task 2 |
| C12 | Folded (convergent, route states) | R-5cii21; Task 6 |
| C13 | Folded: startup hashes every inline script against the meta, so review focus 4's startup refusal is now real | R-5cii7; review focus 4 |
| C14 | Folded (convergent, route states) | R-5cii21; Task 5 |
| C15 | Folded: every state row is seeded; the drift fixture is a transitive package | Task 3 acceptance |
| C16 | Folded (convergent, token log): the token rides four request shapes, the out-of-prefix 404 joins the header list, and POST refusals are logged without addresses | Tasks 1 and 7 |
| C17 | Folded: the refresh precondition is asserted, the non-admin rows carry `admin` signals, and an inner-handler row with no identity is added | Task 1 acceptance |
| C18 | Folded: `DUBPLATE_PUBLIC_URL` is always set when serve runs, with the source line cited | Privileged fields, HSTS row |
| C19 | Folded: the expected contrast ratio comes from an independent, cited computation | Task 4 mutation 2 |

### Mechanics lens

| ID | Disposition | Plan location |
| --- | --- | --- |
| M1 | Folded (convergent, task order) | Tasks 5 and 6 |
| M2 | Owner fork, **Ruling 5** (`@types/node` 24.13.6). `skipLibCheck: true` is folded under every answer. | Ruling 5; R-5cii16; Task 3 scaffold, mutation 7 |
| M3 | Folded: rows for an absent optional entry and an absent non-optional entry; the version compare runs only on shared entries | R-5cii11 table; Task 3 seeds, mutation 6 |
| M4 | Folded (convergent, Content-Type) | R-5cii4; Task 1 |
| m1 | Folded: the allowlist drops `node_modules`, keeps `.svelte-kit`, and the criterion is re-aimed | R-5cii10; deviations line; Task 3 plumbing and gitleaks criterion |
| m2 | Folded: the DaisyUI plugin line is named | Task 3 scaffold |
| m3 | Folded (convergent, selector) | R-5cii12; Task 4 |
| m4 | Folded into **Ruling 2** as its recommended answer (the loopback exception) | Ruling 2; R-5cii6; Task 7 HSTS rows, mutation 3 |
| m5 | Folded (convergent, route states) | Task 6 route criterion |
| m6 | Folded (convergent, route states) | R-5cii21; Tasks 2, 5, 7 |
| m7 | Folded: the gate stays unexported until Task 7 adds its first outside caller | R-5cii18; Task 1 interfaces; Task 7 Files |
| m8 | Folded (convergent, n5) | Task 2 Files; package-territory wall |
| m9 | Folded: Task 3's implementer installs once to produce the lockfile | R-5cii11; Task 3 scaffold |
| m10 | Folded (convergent, route existence) | R-5cii19 |
| m11 | Folded (convergent, token log): `newServeShell` takes its logger | Task 7 "The log channel" |

### Risk lens

| ID | Disposition | Plan location |
| --- | --- | --- |
| R1 | Folded, one part refused. Folded: the path must end in `.css`, the content must be UTF-8 with no NUL byte, and a roster-fixture row is added. **Refused:** resolving the path with `EvalSymlinks` against the roster file and the state directory. The `.css` rule already excludes the roster and every secrets file by name. A `*.css` symlink to the roster takes a deliberate act, and guarding it would thread the state directory and the roster path into the file reader for that one case. | R-5cii13; privileged-fields row; Task 6 criteria, mutation 7 |
| R2 | Folded (convergent, Content-Type) | R-5cii4; Task 1 |
| R3 | Folded (convergent, route states) | R-5cii21 |
| R4 | Folded (convergent, token log) | Task 7 |
| R5 | Owner fork, **Ruling 6** (keep `no-referrer`) | Ruling 6; R-5cii6; Task 8 gap list |
| R6 | Folded (convergent, anchor) | Task 6; Task 7 G7 |
| R7 | Owner fork, **Ruling 7** (refuse startup at 5c-ii, carry to rung 12) | Ruling 7; Task 8 gap list |
| R8 | Folded: rung 12 re-validates every existing row's principal; the shared-origin residual is carried | Privileged fields, `streamer_username` row; Task 8 gap list |
| R9 | Folded: a privileged-fields row for reading the roster | Privileged fields |
| R10 | Folded (convergent, n5) | R-5cii5; Task 2 |
| R11 | Folded as declared deviation R-5cii20, with an erratum drafted at the close | Deviations list; R-5cii20; Task 8 errata |
| R12 | Folded as a carry to 5d: `deploy.sh` builds from a clean committed tree | Task 8 gap list |
| R13 | Folded (convergent, route existence) | R-5cii19; Goal; exit gate item 1; Task 8 handoff |

The risk lens's two notes are not findings. The `data:` URI note is
carried to 5c-iii through the Task 8 handoff. The gate log-volume note
gets no plan change, since journald bounds it and no task depends on it.

## The ceiling

The task count is unchanged at eight plus Task 0. The fold adds about
0.40M of named states and proof rows, and Task 6 takes a fourth
deliverable. The projection moves from 6.35M to 6.75M. The ceiling
moves from 8.8M to **9.4M** so the 80% flag (7.52M) stays one task above
the projection. Ruling 1 carries the new figure. 5c's three passes
still total about 20M.

## Errata owed to ratified documents

None of these edits a ratified document now. The close drafts each
erratum (Task 8).

- **Rung 5 spec, "The embed path" (`:1393-1394`):** "Every served
  response carries ... beside the CSP" narrows to every document for the
  CSP (R-5cii20). This is owed on approval, since the deviation is
  declared.
- **Rung 5 spec, "Presentation (R14)" ("Each is `KindString`"):** owed
  only if Ruling 4 is answered no, which makes the two optional keys
  integration-class rows.
- **Rung 5 spec, R32 (`Referrer-Policy: no-referrer`):** owed only if
  Ruling 6 is answered no.
- **No erratum** for Ruling 2's loopback exception, since no ratified
  document specifies HSTS. None for the operator file's `.css` rule
  either: DL R4.4 names "an optional file, named in config" and does not
  specify its class. None for the DaisyUI plugin line, which keeps
  W5.9's built-in themes enabled. CLAUDE.md is unchanged by the
  revision; Task 3 already edits its gate paragraph and web rule.

## Second fold

Revision 2 (`74066b1`) to revision 2.1 (second fold, awaiting prose
review). Input: the verification read at `cebcda2`
(`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-fold-verification.md`),
V1 through V9. Scope was limited to those nine. Each finding's evidence
was checked against the tree, `go doc`, or the kit 2.70.3 source in the
verifier's probe app before it was folded.

| ID | Severity | Disposition | Where |
| --- | --- | --- | --- |
| V1 | major | **Folded, with a different mechanism.** Confirmed: a child tsconfig's `include` replaces kit's list. The fold appends through kit's documented `kit.typescript.config` hook (`types/index.d.ts:851-861`; called at `write_tsconfig.js:131`) instead of restating kit's list, which would drift silently at a kit bump. Kit's generated list (`write_tsconfig.js:58-85`) is quoted in R-5cii16. `allowImportingTsExtensions` is set in the same hook. Measured on a scratch copy of the probe app: `src/` still checked, a `./lib.ts` script import accepted. A `src/` seeded-error row is added under every answer to Ruling 5, with mutation 8 (a plain `include` narrows the check). | Ruling 5 "What yes builds"; R-5cii16 "How the scripts join the set"; Task 3 scaffold, `src/` row, mutation 8 |
| V2 | major | **Folded.** Confirmed: `Makefile:52` is the `.PHONY` list and does not yet name `web`. Task 3 adds it; a row proves `make check` and `make web` fail naming `npm ci --prefix web` with `node_modules` deleted; mutation 9 drops the entry. | R-5cii10 "Order"; Task 3 leg, `make` rows, mutation 9 |
| V3 | minor | **Folded.** Mutation 15 now drops the query-key check; the `r.Form` read is stated as an equivalent mutant. Mutation 5's symptom corrected to `malformed`. The privileged-fields row adds `unknown_field`. The query-field row's fixture now names the `?admin=on` and `?foo=1` answers. | R-5cii4; privileged-fields table; Task 1 query-field row, mutations 5 and 15 |
| V4 | minor | **Folded.** Confirmed: `Register` mounts `pattern` and `pattern+"/"` (`serve.go:94-96`), and `terminated` dispatches through `s.mux.Handler(r)` (`serve.go:218`), leaving `r.Pattern` unset. Task 2's mount wraps the gated API in an exact-path check against `prefix.String() + "/roster"`, falling to `serve.Terminator()`. Rows for `roster/x`, `roster/`, and `roster/roster`; mutation 8 is the suffix test. Task 1's "any path under `roster/`" sentence moves to Task 2. | Task 1 GET outcome; Task 2 mount, rows, mutation 8; R-5cii21 roster row |
| V5 | minor | **Folded.** Fixture (i) keeps a second admin and first asserts `OpenRoster` succeeds on the edited file, so the row reaches `roster.go:798`, not `:460`. | Task 1 `failed` rows |
| V6 | minor | **Folded as a correctness fix, not an owner fork.** RFC 6797 section 8.3 matches by host alone, so the loopback exception is fixed under every answer. Ruling 2 now asks only the HSTS shape; "drop the exception" is removed from "What no builds". | Ruling 2; R-5cii6; Task 7 header row |
| V7 | minor | **Folded by withdrawal.** Task 5 mutation 2 is withdrawn: the "present more than once" state row refuses the planted token under the correct rewrite too, and that row is the guard. Redefining the token's legal site would add a parse-aware count for no added protection. | Task 5 mutation 2 |
| V8 | minor | **Folded.** Task 2's Files parenthetical names every n5 row in `roster_test.go`. | Task 2 Files |
| V9 | minor | **Folded.** The cache table's last row reads "every 404 and every non-JSON refusal", and names `writeAnswer`'s `no-store`. | R-5cii8 |

**Refused:** none. **Ceiling:** unchanged at 9.4M (flag 7.52M). The
second fold adds five rows and three mutations across Tasks 1 through 3
and withdraws one, well inside the 0.77M between the 6.75M projection
and the flag. The task count and deliverable count (31) are unchanged.

## Prose review fold

Revision 2.1 (`0bb3722`) to revision 2.2 (prose folded, awaiting owner
approval). Input: the prose review
(`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-prose-review.md`),
P1 through P15. Every changed pointer was checked against the tree
before it was written.

| ID | Severity | Disposition | Where |
| --- | --- | --- | --- |
| P1 | blocker | **Folded, re-derived.** The Reason now rests on 5c-ii's 6.75M projection and drops the bare 13 × 0.725M figure. The one-pass branch uses the plan's own adders: 13 × 0.725M (9.43M) plus 0.95M (5c-ii) and 0.60M (5c-iii) projects 10.98M, and 1.3 times that is a 14.3M ceiling with its 80% flag at 11.4M, above the projection. The review's 13.8M ceiling is not taken, since its flag would sit on the projection. | Ruling 1 |
| P2 | blocker | **Folded.** Each "no" builds one thing. Ruling 2: no HSTS header, finding 18 open. Ruling 5: `.js` gate scripts with JSDoc, the `ts-conventions` claim dropped. Each names the other option as an answer Geoff can give instead. Ruling 5's `skipLibCheck` sentence leaves the ruling; R-5cii16 already carries it. | Rulings 2 and 5 |
| P3 | warning | **Folded.** Ruling 3 asks one question: does Task 4 author the `dubplate` match file and the `gruvbox` palette file now. The stubs move into the Reason as fixed under every answer (R-5cii1). | Ruling 3 |
| P4 | warning | **Folded.** `thresholds.go:124-128` is now `internal/config/thresholds.go:123-126`. The `resolveIdentitySource` pointer now reads `internal/config/config.go:291-296`. Both checked against the tree. | Ruling 4; privileged-fields table, HSTS row |
| P5 | warning | **Folded.** The section opens with "Seven questions, each with a recommendation", the standing rulings, and a pointer to this record. | Rulings for Geoff |
| P6 | warning | **Folded.** "Fixed, not asked" is now two sentences pointing to R-5cii6. R-5cii6 takes the section 8.3 and 8.1.1 reasoning, the `isLoopbackHost` pointer (`auth.go:474-482`, checked), and the `validatePublicURL` pointer (`auth.go:465`, checked). | Ruling 2; R-5cii6 |
| P7 | warning | **Folded.** Ruling 7's Reason is split into separate sentences. | Ruling 7 |
| P8 | warning | **Folded** in the review's wording. | Ruling 4 |
| P9 | suggestion | **Folded.** The WebKit hedge is its own sentence. | Ruling 6 |
| P10 | suggestion | **Folded.** Ruling 1's "What yes builds" is four sentences. | Ruling 1 |
| P11 | suggestion | **Folded.** Ruling 5 now names kit's generated type files, the two vite configs, and `src/`, `test/`, and `tests/`, matching R-5cii16's list, and points to it. | Ruling 5 |
| P12 | suggestion | **Folded, with a figure change.** Decomposing the 2.8M showed it left out 5c-i's planning, which Ruling 6's 18M counted (the 5c plan fold record, `:511`). Planning is now 4.3M: 5c-i's 1.5M estimate, 5c-ii's 1.0M plus 1.0M for this split (`:503-505`), and 5c-iii's 0.8M. 5c then comes to about 22M, not 20M, and rung 5 to about 41M, not 39M. Ruling 1's "What yes builds" carries both. | Token ceiling, 5c-iii paragraph; Ruling 1 |
| P13 | suggestion | **Folded.** This record's second-fold headroom reads 0.77M. | This record, "Second fold" |
| P14 | suggestion | **Folded** at the R-5cii19 sentence the review names. The other two sites are left as the review directs. | R-5cii19 |
| P15 | suggestion | **Not folded.** No touched line fell in the named ranges. | None |

**Refused:** none. **Ceiling:** unchanged at 9.4M (flag 7.52M). The task
count and deliverable count (31) are unchanged.
