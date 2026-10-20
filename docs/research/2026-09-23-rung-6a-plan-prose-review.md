# Prose and fact review: rung 6a plan, revision 2.1

The reviewer was `prose-voice-reviewer` (Opus), writing on 2026-09-23. It was read-only, so the conductor saved this condensed record of its report.
`tellgrader --register docs` returned 0 findings. The review confirmed the ceiling arithmetic, the commit ids, the spec line
pointers, and the driver, header, and stdlib quotes.

## Blockers

- **B1** (`plan:75-85`). Ruling 2's question ("exit nonzero only for a state `serve` cannot start
  through") contradicts its "yes" list, which exits nonzero on an interrupted create and on a held
  lock. Rewrite the question as "exit 0 when the store is absent, older, or current, and nonzero for every other
  state". Add a sentence saying why an interrupted create exits nonzero.
- **B2** (`plan:530-531`, `:912-913`, `:1701-1708`). The first-check order contradicts itself.
  R-6a7 puts the restore-intent check first, Task 1 puts the owner check first, and Task 6 puts the
  intent check fifth, after serve's leaf create. The spec (`:235`) puts the intent check first for
  every writing open. Fix: stat, then the owner check, then the intent check, and only then create,
  probe, lock, or open. State the order once, and point R-6a7 and Task 1 at it.

## Warnings: facts

- **W1** (`:377` against `:2030-2031`). The write mode's exporting caller is stated two ways. Fix:
  6b-i exports the write method with `ops/local` and the write mode with `backup`.
- **W2** (`:131-150`). The spec-deviations list omits E4, E10, and E11, which follow the rulings. E8
  drops "and logs the holders".
- **W3** (`:1531-1532`, `:1666-1667`). The hook 15 claim is not in the spec's hook ledger, which
  lands hook 15 at 6c. Drop the claim or cite a source for it.
- **W4** (`:665-681`). R-6a14's heading says "two named constants", but the section defines three.
- **W5** (`:2097`, `:359-361`). The contested-surfaces row omits that 5c-i Task 4 also edits
  `doctor_test.go`.
- **W6** (`:598-601`, `:1206-1208`). R-6a10's emphasis misreads the spec at `:327-329`. Quote
  `lang_transaction.html` section 3 and say the spec names the four codes.
- **W7** (`:505-506`, `:1786-1787`). The plan says "two probes" and then names three. The probe name
  "D2" collides with the risk lens's finding ID D2.

## Warnings: register

- **W8** (`:89-112`). Ruling 3 is one 275-word paragraph. Split it on Reason, What yes builds, and What no
  builds. Also split Ruling 2's build sentence.
- **W9** (`:1701-1708`, `:1734-1742`, `:2029-2044`). The mode order, serve's procedure, and the 6b-i
  handoff are run-on sentences. Make each one a numbered list or bullets.
- **W10**. About ten sections turn on "So". Lead with the decision in most of them.

## Suggestions

- **S1**: Say 5.45M is "just past" the 5.44M flag.
- **S2**: The plan says "three more multi-host types" and then lists four constants.
- **S3**: Add a line that restore-in-progress and busy come before classification.
- **S4**: Define CH:165-168.
- **S5**: "Halves" becomes "parts".
- **S6**: Say "nine fixtures across seven rules".
- **S7**: Spell out the 4.55M remainder's arithmetic.
- **S8**: Put the headings in sentence case.
- **S9**: Replace the three fragment topic sentences.
- **S10**: Drop the reflexive bold lead phrases from the Outcomes bullets.
- **S11**: Rename R-6a7 to cover staging cleanup.
