# Rung 6d plan prose review (revision 2.1)

One `prose-voice-reviewer` (Opus) graded the plan as developer docs under Google style.
`tellgrader --register docs` found 0 tells over 14,805 words, with cadence CV 1.39. The reviewer
spot-checked about 60 citations and found most exact. Its verdict: the voice is sound and the
risk is factual. Line numbers refer to the plan at `985ff3a`.

## Blockers

- **B1. Ruling 3's yes/no is inverted at `:17`, `:71`, and `:134`.** Those lines say a "yes" on
  Ruling 3 adds a task and splits the pass. Ruling 3 itself (`:86-95`), `:57`, 6d-R8 (`:623`),
  and Task 9 (`:1306`) say that "no" builds Task 8b. Rewrite as "a 'no' on Ruling 2 or on
  Ruling 3". The error dates from revision 1.
- **B2. 6d-R11 (`:626`) gives a false reason.** It says 5d never starts `serve`, but 5d's step 14
  starts it once, supervised. Rewrite the cell: "Not built: no pre-6d store or snapshot can
  exist. 5d's one supervised start runs the pre-6d binary, which never opens a store, and
  `OpenWrite` refuses an absent one."
- **B3. Two spec citations in the deviations table are wrong, and Task 9 would copy them into the
  spec.** 6d-R2 (`:616`) cites `:853-855`, but the quote is at spec `:839-840`. 6d-R11 cites
  `:715-716`, but the quote is at `:713-714`. Task 0 does not re-resolve spec lines.

## Warnings

- **W1. Four more wrong citations:**
  - `:408` cites CH `:66`; the right lines are `:83-84`.
  - `:1363` cites spec `:855-860`; the right lines are `:843-846`.
  - `:238` cites spec `:361-365`; the right lines are `:375-376`.
  - `:137` should cite both `:1128-1130` and `:1141`.
- **W2. Citation bases are mixed.** The header says `20f4a7c`, but the members-page lines
  (`:223-224`, `:564`) and the `run.ts` lines (`:567`, `:982`) are at a later HEAD. Label each
  base, and add a line to "Not verified" saying the `run.ts` and members-page lines are at
  `985ff3a` and that R2 at `92147bc` moved the members lines.
- **W3. Task 4's fixture-state claim pairs the wrong mutations (`:971-973`).** Rewrite: "The
  add-member row has added a member before the first reset. A reset whose baseline is only the
  two seeded rows then answers 2 where the correct reset answers 3. A reset that deletes nothing
  answers the right count but leaves the rendered list longer than it."
- **W4. The ceiling variants omit the split cost (`:134-135`), and `:72` points to a "table"
  that is a numbered list.** Recompute each variant with about 0.5M added for the split, and say
  "the 'Token ceiling' arithmetic".
- **W5. Tasks 6 and 7 overlap on `mintLinkFor` (`:1103-1105`, `:1163-1164`, `:284-285`).** Say in
  Task 6 that `mintLinkFor` stops calling `resolveAuthConfig`. In Task 7, confirm that no caller
  remains.
- **W6. Owner touchpoints (`:110-115`) contradicts itself.** Rewrite it as three numbered upgrade
  steps: deploy with `try-restart`, which leaves the unit stopped; one supervised start and stop,
  which creates the store; then `member import --dry-run` and `member import` through
  `systemd-run`. Close with "Apart from the supervised start in step 2, nothing in the upgrade
  starts, restarts, or enables `serve`."
- **W7. Ruling labels collide with other documents' labels, and review labels are undefined.**
  - `:78` should say "spec ruling 9".
  - `:81` should say "rung 5 review's minor finding 14, and the spec's erratum row 10".
  - `:671`, `:676`, `:1132`, and `:1232` say "6c's Ruling 2" next to this plan's Ruling 2.
  - Add one Sources line that defines B1, M1, M2, C-M10, R-M3, D10, S20, and R5S.
- **W8. Rulings 2 and 3 are hard to scan.** Define the dual-key epoch and the store identity in
  one sentence each. Split the long reasons into short sentences, and keep each recommendation
  and its "what each answer builds".
- **W9. Two procedures are run-on prose.** Write Task 8's runbook outcome (`:1237-1245`) and
  R-6d19's order (`:597-600`) as numbered steps.
- **W10. The execution model (`:1355`) contradicts Task 8's header.** Rewrite: "No task runs in
  parallel. Every task from Task 3 to Task 7 reads the one before, and Task 8 runs last to keep
  one gate lane."
- **W11. A stale parenthetical (`:1220`).** Rewrite as "`docs/server-runbook.md` (the template
  edit moved to Task 5)."

## Suggestions

- **Citations.**
  - `:37-38` and `:383` should cite `store.go:286-298`.
  - `:472` should say "finds only comments and tests".
- **Wording.**
  - `:7-8` should read "will first rewrite most of the files this plan names".
  - `:12` should read "planning spend is recorded against rung 6, outside this ceiling".
  - Split the status line at `:3-4`.
- **Task 3's Files.** Add `identity.go`'s `Principal` doc comment, which still promises the id
  substitution that R-6d10 rules out.
