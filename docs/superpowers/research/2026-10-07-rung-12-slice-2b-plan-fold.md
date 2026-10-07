# Rung 12 slice 2b plan fold (revision 1 to 1.1)

Target: `docs/superpowers/plans/2026-10-07-rung-12-slice-2b-oracle-harness.md`, revision 1 at `0ae2b54`, folded to
revision 1.1 (uncommitted at writing; the conductor commits). Reviews folded:

- contract lens, `2026-10-07-rung-12-slice-2b-plan-review-contract.md` (C1 to C15, O1);
- mechanics lens, `2026-10-07-rung-12-slice-2b-plan-review-mechanics.md` (M1 to M18);
- risk lens, `2026-10-07-rung-12-slice-2b-plan-review-risk.md` (DR1 to DR14).

Each finding was checked before it was acted on. Checked directly for this fold, on 2026-10-07:

- the pinned XSD, fetched to `/tmp/claude-1000/fold2b/` (sha256 `9baee72f...0387`): the `Response` choice at `:12`
  is `minOccurs="0" maxOccurs="1"` (DR1); `track` on `Child` is `xs:int` (`:232`); `Captions`, `AudioTrack`,
  `VideoConversion`, `ArtistInfoBase` exist and have no same-name OpenAPI schema (C12);
- `xmllint --version`: the version line goes to stderr, stdout is empty (C11, M8);
- libxml2 `v2.15.4` `doc/xmllint.xml` DIAGNOSTICS (exit codes 0 to 10) and its `--nonet` entry ("Has no effect with
  libxml2 2.15 or later"), and the 2.15.0 NEWS ("The built-in HTTP client ... were removed") (DR8);
- `go doc os/exec.ExitError` (M1), `go doc net/url.ParseQuery` and `URL.Query` (C10), and the go1.27.1 encoder source
  `marshal.go:224-225`, `:744-748`, `xml.go:1922` (M5, DR7);
- the uupd journal (`"msg":"Brew module" ... "Enabled":true`, `Updating Homebrew` on 2026-10-07), `systemctl
  is-enabled` (`uupd.timer` enabled, `brew-upgrade.timer` disabled), `ublue-os/uupd` `drv/brew/brew.go:50`, `:63`
  (`brew update`, then `brew upgrade -y`), `brew pin --help`, `brew info libxml2` ("Installed (as dependency)"), and
  the Brewfile (no `libxml2` line) (M2, DR9);
- the Navidrome multi-library page at `b0fe572`: `:30` "Admin users automatically have access to all libraries",
  `:150` (M6, C6);
- the study record's sections 1, 3, and 4, the divergence ledger's "Test owed" column (all 66 cells), and the spec at
  `:340-345`, `:809-818`, `:871-926` (C1, C2, M4, DR3, DR6, M18);
- the pinned OpenAPI bundle: 87 paths, 33 of them in neither section 3 nor the empty-success set (computed with `jq`
  and `comm`); `GetAlbumResponse`'s `oneOf`/`allOf` shape and `required: ["album"]`; `Indexes` requires
  `ignoredArticles` and `lastModified` and leaves `additionalProperties` unset (C1, DR1, C5);
- the 2a code and data: `grade/testdata/seed-ledger.json` (27 active rows, 20 on `serverVersion`/`type` paths),
  `grade_test.go:463-470`, `ledger.go:305-308` (`applies`), `grade.go:263-298`, `conformance.go:92-117`,
  `fixture/manifest.go:67-83`, `tagio/encode.go:123`, the seed manifest (no `genres`), the seed session's metadata
  (62 XML responses, all `application/xml`), gonic `015` (`replayGain: null`) and `008` (`releaseTypes: [""]`), and
  the seed golden (no JSON `contract` finding on `getIndexes`) (C5, C7, M3, M7, M10, M13);
- `~/.claude/workflows/pass-execute.js:240-268`, `:359` (`PASS_CLASSES.tool.gateLane` is `"light"`; the `docs` bar)
  (M17, C14).

The mechanics lens's live measurements (the 1,000-error stderr cut, gonic's 41-line projection, the light lane's
605 MB peak, `xmllint` inside the `systemd-run` scope) are taken as that lens recorded them, as 2a's fold took its
probes.

## Two convergent defects, fixed at the root

**The assignment and the ledger's slices had no named source (C1, C2, M4, DR3).** Section 3 names 51 of 87
endpoints, splits six verdicts one field cannot hold, gives no slice, and no test checked a value; R13's slice rule
left rows undecided. Revision 1.1 fixes every value in the plan. R27 names the three sources (section 3 under the
rulings, the empty-success additions, spec `:809`'s error 0 for the 35 others), takes a split verdict's narrow value
with the wider grant kept in its reason (C1's rule), keeps an oracle whose exception a ledger row already names, and
fixes each served endpoint's slice from the ladder text. R13 becomes one mechanical rule plus a list of every row
without a slice token. Task 5's values test pins all of it and refuses a `contract-only` value not on a named list
(DR3's direction). Task 6 recomputes each slice. C1's narrow rule and DR3's "contract-only is the silent direction"
pull opposite ways on `getOpenSubsonicExtensions`, `getPlayQueue`, and `savePlayQueue`; the fold takes the narrow value
and puts each on the test's named list, so the loss is explicit and a new one needs a test edit.

**The XML tier could pass a body for the wrong reason (DR1, M13, C4, M12, M5, DR7).** The XSD's choice is
`minOccurs="0"`, so xmllint accepts any envelope for any endpoint; a residual left in the projection hides everything
after it from xmllint; and the line-to-ordinal table assumed one line per element. Revision 1.1 binds each body to its
endpoint, status, and content type from the pinned OpenAPI document, drops each residual from the projected document
after reporting it (which also retires the dedupe rule), and builds the line table from the newlines actually written,
with proofs on an indented gonic body and a multi-line text node.

## Dispositions

| IDs | Disposition | Where |
|---|---|---|
| C1, M4 (assignment half), DR3 | Folded (convergent; above). C1's loader rule (`not-served-0.8` implies `contract-only`) and value test, DR3's direction test, and M4's representation choice (one value plus the reason) land together. | R27; Task 5 outcomes, states, test mandate, mutations; "What slice 3 consumes" |
| C2, M4 (slice half) | Folded. R13's rule and list; L-34, L-37, L-06, and the agreement rows are listed; Task 6 recomputes `slice` and `TestCommittedLedger`'s owed-by-slice map takes the counts. One correction to the lenses: L-06's cell carries "slice 4" inside its citation, so it takes its token. | R13; Task 6 |
| C3, M9 | Folded as M9's conventional fix: the "`format: any` covers both" claim is withdrawn, R20 states that the paths agree only for lists of two or more and how a singleton `contract` row is written, and Task 4 gains a singleton proof pinning the unindexed path. C3's mechanism (a `maxOccurs` flag per ordinal and indexed XML singletons) refused: the miss is loud (an unledgered finding, never a hidden one), and it would give the contract tier's XML paths a grammar the oracle tier's XML paths do not share, splitting R20's one grammar. | R20; Task 4 |
| C4, M12 | Folded. Residuals are reported, then dropped from the projected document; the proof puts a residual element first, then an xmllint-only error, then `titel`, and asserts each origin. The dedupe rule is gone. | Task 3 projection, proofs, mutations |
| C5, DR5 | Folded. Task 5's three-source fixture state is named (dubplate's bodies differ wherever a comparison is claimed; an unmapped request; non-zero counts asserted). C5's "one unknown attribute" perturbation replaced by removing a required member: `Indexes` leaves `additionalProperties` unset, so an unknown attribute would not violate. Task 4's twin uses gonic's `getAlbum` (it carries residuals) and asserts a literal, non-empty list. | Task 4, Task 5 |
| C6, DR13 (third item) | Folded. Task 10's gonic-only form is written, with L-19's Navidrome half kept as slice 1's observation and an erratum slot. C6's "grant the admin the new library" refused: the multi-library page `:30` says admins see every library. | Unattended-run rulings; Task 10; Errata section |
| C7, M10 | Folded. 20 rows retire, not 22 (C7 counted every `volatile` row; two are `song/created` rows that stay). The stale-row test's live row moves to `song/created`. | Task 7 files and outcomes |
| C8, M16, DR13 (second item) | Folded. The test pins each prediction's recorded state. | Task 8 |
| C9, M11, DR13 (first item) | Folded with [contract] E10's actual pair: the escape `&` against a literal `&` in JSON. The E6 fixtures are named (gonic `015`, `008`, and a perturbation). | Task 7 |
| C10 | Folded through the conventional fix: the scan reads `ParseQuery`'s documented partial map and treats a parse error as a finding (fail closed); both docs quoted; a malformed-query row added. C10's raw-token regex fallback refused: `ParseQuery` already returns every valid parameter beside its error, so a second parser adds nothing. | Quotes table; Task 7 |
| C11, M8 | Folded. The version is read from stderr and the fake writes there; only `Schemas validity error` lines count, the trailer is expected, and any other line under exit 3 is a tool failure. C11's newline item folds under M5. | Quotes table; Task 3 runner, proofs, mutations |
| C12 | Folded, all four items: the per-type additions golden, the `isCompilation` residual case (it kills the union mutation), the no-counterpart types, and the twin's non-empty list (through DR5's gonic body). | Task 3, Task 4 |
| C13, M18 (second item), DR6 | Folded. L-51's `test` takes ruled text and joins the fidelity exemptions; the close amends spec `:342-345`; HISTORY routes to slice 3 how a row on a `contract-only` endpoint closes. | Task 6; Task 12 step 6; errata below |
| C14 | Folded in part: every pass claim cites a screenshot or a DOM capture, and the diff-reviewer checks each. The relabel refused: `pass-execute`'s `docs` bar already reads "Prose register belongs to the register chain, not this review", so no register chain runs in-chain; the plan now says so. | Task 11 |
| C15, M14 | Folded. The rebase replays the plan commits and is no stop signal; `npm ci --prefix web` after a `web/` bump. | Task 0 |
| O1, DR14 | Folded. Each client walks the whole script against Navidrome and the folder-view step against gonic. | Task 11; ceiling note |
| M1 | Folded. `Cmd.Stderr` to the runner's own buffer; the `ExitError` doc quoted; a fake emitting more than 64 KiB proves every line survives. | Quotes table; Task 3 |
| M2, DR9 | Folded as M2 states it. DR9's premise is wrong: `brew-upgrade.timer` is disabled, but `uupd.timer` is enabled and its Brew module runs `brew upgrade -y` daily. The lane launch adds the Brewfile line and `brew pin libxml2`; the refusal names the bump path; the close's CLAUDE.md batch records it. | Quotes table; R18; Lane launch 5; Task 12 |
| M3 | Folded through its first option: `divergence` and `test` are required on `owed` rows by the loader and on every committed row by a committed-ledger test, so the 27 test-only active rows stay. The three test files that build `owed` rows with a class are named. | Task 6 |
| M5, DR7 | Folded. Whitespace-only text is not copied; the line table counts written newlines; proofs on a real indented gonic body and a multi-line text node; a prefixed body is one "unsupported" violation, never a pass. M5's "own writer" option not taken: no oracle or Go encoder emits a prefixed body, so supporting one buys nothing at 0.8. | Quotes table; R20; Task 3 |
| M6 | Folded with the conductor as the observer: implementers carry no browser tool, and the conductor drives Playwright MCP against the container `harness oracle-up` starts, recording the login exchange and the creation request in a committed note before segment C. M6's implementer-run Playwright script refused for that reason. Disjoint container paths and the fake's two routes folded. | Unattended-run rulings; Task 9 |
| M7 | Folded, all four items: manifests beside `manifest.json`; a shared `genres` value; `albumartist` as the key's set, with `Loose files` excluded; the grader builds each collection to digest it. | Quotes table; Task 9; Task 10 |
| M13 | Folded inside DR1's binding (status and content type at the envelope). | Task 3, Task 4 |
| M15 | Folded by removal: Task 6's report counts are dropped. `TestCommittedLedger` already pins owed counts by slice, and no grade report reader needs them, so the seed golden stays untouched by Task 6. | Task 6 |
| M17 | Folded. Task 1's entry carries `gate` and `gateLane: "heavy"`. | Header, "Per-task gate" |
| M18 (first item), DR10 | Folded. One scope row names the known-deviation list's mechanism and owners. | Scope table; Task 8 |
| DR1 | Folded (convergent; above). | Quotes table; Task 3 binding, proofs, mutations; Task 4 |
| DR2 | Folded. Accounting runs one way; fidelity binds `owed` rows; an active row derived from a record row names a pair from its sources; R28 tells slice 3 how a `Both` row activates; two activation proofs. | R28; Task 6 |
| DR4 | Folded. The body scan lives in `session` beside 2a's request rule, run by the writer and the loader; `grade` and `record-oracle` reach it from there. | Global constraints; Task 7; Task 12 |
| DR8 | Refused. libxml2 2.15.0 removed the built-in HTTP client, and `doc/xmllint.xml` says `--nonet` "Has no effect with libxml2 2.15 or later"; R18 pins 2.15.4, so the pinned build cannot open a connection. The quote now sits in the quotes table. | Quotes table |
| DR11 | Folded, narrowed: test schemas use invented type names, and the close checks that no `.xsd` file under `xmlrule/testdata/` names a real XSD type. The grep is limited to `.xsd` files, since C12's additions golden names XSD types legitimately. | Task 3; Task 12 acceptance |
| DR12 | Folded. Every tree [base-url] hashed is hashed; a changed hash stops the run; committed screenshots are cropped and read for host paths. | Unattended-run stops; Task 11 |

## Rulings for Geoff

One, carried from the conductor's relay of the plan author's budget fork; revision 1 at `0ae2b54` did not carry its
text, so revision 1.1 adds it as Ruling 1: run 2b as one eleven-task pass at 9.0M, or split Tasks 9 to 11 into a pass
2c. Recommendation: one pass. The folds gave Tasks 10 and 11 written fallbacks, and a split costs about 0.8M more for
a second close and merge. No finding became a fork: every question the lenses raised has an answer in the records,
the code, or a quoted tool behavior.

## Errata owed to ratified documents (not edited here)

- **Spec `:342-345`** makes gonic authoritative for `getIndexes` and `getMusicDirectory`. Under ruling C both are
  contract only. Task 12 step 6 carries the amendment with the forks-ruled note.
- **Study record section 4**, "Routed to slice 3: ... the two-collection fixture that proves the filter", contradicts
  its own "Routed" list and spec ladder entry 2, which route the fixture to slice 2. The plan follows the spec (R22).
  Owed: section 4's line names slice 2 for the fixture and slice 3 for the private-collection row.
- **Study record section 3**, cross-cutting row "The unimplemented answer and the empty-success set: Contract only",
  overlaps the info family row ("Both"), whose five members are in the empty-success set. R27 reads the specific row as
  governing. Owed: one clause in the cross-cutting row excepting the info family.
- **Conditional:** if Task 10 takes its gonic-only form, L-19's Navidrome half rests on slice 1's observation, and the
  plan's errata section records it.

The ledger record's L-51 and L-52 cells keep their pre-ruling text as dated evidence. The committed ledger carries the
ruled texts, and the transcription test names each exemption with its ruling, so the change is explicit.

## Measures

- **Dispositions:** 48 finding IDs in 32 rows. 47 folded, 6 of them in part with a sub-proposal refused (C3's
  `maxOccurs` paths, C6's grant step, C10's regex fallback, C14's relabel, M5's own writer, M6's implementer script).
  One refused whole (DR8). No new owner fork; one carried budget fork (Ruling 1).
- **New-mechanism findings folded, each with its source and measured defect:**
  - the endpoint binding (DR1, M13): source, 2a's own conformance checker, which resolves route, status, and content
    type before validating (`conformance.go:92-117`), as kin-openapi's response validation does; defect, the pinned
    XSD's `minOccurs="0"` choice, read at `:12`;
  - dropping residuals before xmllint (M12, C4): source, slice 1's projection method and Ampache's conformance test,
    both of which remove what the authority does not cover before validating; defect, the mechanics probe on gonic's
    `getArtist`, where xmllint reported nothing after `roles`;
  - the assignment values test (C1, DR3): source, the record-against-data transcription test Task 6 already runs,
    itself 2a's Task 11 carry; defect, 36 of 87 endpoints without a section 3 row and no test reading any value;
  - the slice recomputation (C2): same source; defect, revision 1's rule decided no agreement row and left L-06,
    L-34, and L-37 outside its cases;
  - the per-type additions golden (C12): source, the Go `-update` golden convention 2a's goldens use; defect, a count
    proof that asserted nothing.
  Five folded.
- **New-mechanism findings refused:** C3's `maxOccurs` path rule (loud failure mode; splits the path grammar); C10's
  raw-token fallback (`ParseQuery`'s documented partial map answers it); DR8's `--nonet` (no effect on the pinned
  build); M5's own writer (no prefixed body exists to support). Four refused.
- **Plan lines:** 827 before, 1023 after.
- **Ceiling:** unchanged at 9.0M (flag 7.2M). The folds add proofs to Tasks 3, 5, and 6 and one conductor observation;
  they remove Task 3's dedupe rule and half of Task 11's GUI walks. Tasks unchanged at eleven build tasks plus Task 0,
  the conditional Task R, and the close; segments unchanged.

## Second fold (revision 1.1 to 1.2)

A narrow fold of the verification read (`2026-10-07-rung-12-slice-2b-plan-fold-verification.md`, V1 major, V2 to V9
minor), scoped by the conductor to V1, V2, and V5, plus any other finding that one line of the plan resolves with no
new mechanism. Each finding was checked against the plan and the code before it was acted on. No probe was needed:
every behavior stated below is read from the plan at revision 1.1 or the 2a code at `c4b7afd`.

Checked directly:

- **V1.** The plan's conductor step (`:836-844` at 1.1) observes and records, and names no replay. Task 9's readiness
  (`:865-866`) assumes Navidrome's `getScanStatus` `count` is the total across libraries, which nothing quoted
  covers. Task 10's fallback fired only "before Task 9" (`:913`). 2a's stack gives Navidrome one mount, `/music`
  (`stack/stack.go:136-143`), so a library the conductor creates by hand holds no tracks. A replay therefore settles
  the request and the folder listing, but not the count rule. The fold says so rather than claim the step fixes the
  rule, and moves the count question to Task 9's live evidence under the new "during Task 9" trigger.
- **V2.** `Session.validate` refuses a body whose envelope `serverVersion` differs from its source's metadata
  (`session/session.go:369-377`), and both `Write` (`:195`) and `Load` (`:173`) run it. `Grade` reads every
  session through `Load` (`grade/grade.go:101`). 2a's `handSessionAt` helper writes through `Write` with Navidrome's
  version on both sources (`grade/grade_test.go:122-145`). Task 5's `dubplate` source mixes gonic and Navidrome
  bodies, so it cannot load as written. The same check reaches Task 4's twin, whose two sources hold gonic's body.
- **V5.** The ceiling's own arithmetic (`:125-132`) gives a build-task rate of about 0.58M. The split is 2b at about
  6.5M (0.25M + 8 x 0.58M + 0.3M + 1.3M) plus 2c at about 3.3M (3 x 0.58M + Task 11's 0.3M premium + 1.3M). That
  sums to about 9.8M against the one-pass estimate of 8.55M, a delta of about 1.3M, the second close. Revision 1.1
  set 9.83M against the 9.0M ceiling, which mixes an estimate with a ceiling. Its 2c line also named "the conductor's
  observation" where the 0.3M is Task 11's premium; the ceiling note prices the observation as offset by the removed
  walks, in either shape.

| ID | Disposition | Where (revision 1.2) |
|---|---|---|
| V1 | Folded. The conductor's step replays the recorded login and creation with `curl` against a fresh `harness oracle-up` and quotes the creation status, `getMusicFolders`, and `getScanStatus`. The record states the replay cannot settle the count rule over two populated libraries. Task 9 gains a degraded form (Navidrome mounts the first collection only; readiness checks each oracle against its own collections), taken before Task 9 or during it. During Task 9 the implementer reports the live contradiction, and the conductor re-dispatches in the degraded form. The trigger reads "before or during Task 9" in the rulings list, Task 10, "Not verified", and the errata line. Task 9's live evidence also quotes `getScanStatus`. | Unattended-run rulings; Task 9 conductor step, live evidence, degraded form; Task 10; "Not verified at planning"; Errata |
| V2 | Folded. Task 5's proof rewrites every `dubplate` body's `serverVersion` to its source metadata's one value, citing the loader check. It asserts its non-zero oracle counts below the envelope, since before Task 7's E1 rule the rewritten value differs from both oracles' and would make a bare count vacuous. Task 4's twin names gonic's `serverVersion` in both sources' metadata, the same check on the same helper. | Task 5 test mandate; Task 4 test mandate |
| V5 | Folded. Ruling 1 states the split's estimates (about 6.5M and 3.3M, about 9.8M together) against this pass's 8.55M estimate: about 1.3M more, the second close, with each split pass carrying its own fix-round headroom besides. The fallback claim now names Task 9's and Task 10's degraded forms and drops Task 11's walk, which was the ordinary walk. The recommendation and the fork are unchanged, and Ruling 1 stays Geoff's. | Rulings for Geoff, item 1 |
| V6 | Taken, one sentence inside V1's rewrite of the same paragraph. The step's stated reason is now the real trade-off: a one-time observation with no script to maintain, at the cost of GUI work in the conductor and a re-observation owed by R24's next Navidrome bump. M6's row above keeps its original reason as the first fold's dated text, and this row corrects it. | Task 9 conductor step |
| V7 | Taken, one Files edit: Task 6 names Task 5's assignment loader and its tests (the agreements' validation). | Task 6 Files |
| V3 | Owed, before segment C (Task 7): drop `grade`'s L-07 finding, since the session loader's refusal makes it unreachable through `Grade`; the writer's and loader's refusal is the L-07 check, and the credential proof calls the `session` scan directly. | Task 7 outcomes, proof |
| V4 | Owed, before the lane launch: quote `brew pin --help`'s note on dependents, name the held dependents (chafa, ffmpeg, gdk-pixbuf, imagemagick, librsvg, libxkbcommon, shared-mime-info, xdotool) in the lane launch's Brewfile line and the close's CLAUDE.md batch, and end R18's bump path with a `brew upgrade` of them. | Quotes table; R18; Lane launch 5; Task 12 step 6 |
| V8 | Owed, before segment C (Task 7): say how the scan separates a leaf's query from its URL without `url.Parse` (which errs on a bad path escape before the query is read), and add one states row with a bad path escape. | Task 7 L-07 outcome; states table |
| V9 | Owed, before segment A (Task 3): under exit 3, an unmatched line that directly follows a validity-error line continues that error's message; only an unmatched line with no error line before it is a tool failure. | Task 3 runner; states table |

**Measures.**

- **Dispositions:** nine findings. Five folded or taken (V1, V2, V5 in scope; V6 and V7 as one-line edits). Four owed
  (V3, V4, V8, V9), each with the boundary it must clear before. None refused.
- **New mechanism:** none. The replay is the "replayed over HTTP" criterion revision 1.1 already named, made a step.
  The degraded form is Task 9's own one-collection path, which already exists, applied to Navidrome alone.
- **Plan lines:** 1023 before, 1054 after.
- **Ceiling:** unchanged at 9.0M (flag 7.2M). The replay adds a few HTTP calls to a conductor step that already runs
  the container. The degraded form replaces Task 9's two-library Navidrome work rather than adding to it, and its
  re-dispatch is the one fix round the 0.45M above the 8.55M estimate already prices.
- **Ruling 1 correction for the summary above:** the "about 0.8M more" in "Rulings for Geoff" reads about 1.3M more,
  estimate against estimate, and the fallbacks it names are Task 9's and Task 10's degraded forms.

**Owed items landed (revision 1.3).** The four owed findings are landed before execution. V4 is in the quotes table
(`brew pin --help`'s dependents note, quoted), R18 (the bump path ends with a `brew upgrade` of the eight held
dependents), Lane launch 5 (the pin recorded as done at launch on 2026-10-07, dotfiles `c25573b`), and Task 12 step 6
(the CLAUDE.md batch names them). V9 is in Task 3's runner and its states table (an unmatched line directly after a
validity-error line continues that error's message). V3 is in Task 7's outcomes and proof (`grade` has no L-07
finding; the credential proof calls the `session` scan directly). V8 is in Task 7's L-07 outcome, its states table,
and its mutations (the scan cuts at the first `?` and parses only the remainder; a bad-path-escape row is added).
