# Prose review: rung 6b-ii plan (revision 3, `3df0fd0`)

Reviewer: `prose-voice-reviewer` (read-only; the conductor saved and condensed its report).
Register: Google developer docs, Go arm. tellgrader `--register docs`: 0 findings, 21,357 words,
943 sentences, cadence CV 1.44; the docs-register profile did not resolve, so no measures.

Checked and correct: every ceiling figure (one pass 7.76M projection, 8.24M flag, 10.30M
ceiling; 6b-ii-a 2.64M, 3.12M, 3.90M; 6b-ii-b 5.54M, 6.02M, 7.525M shown as 7.53M; split total
11.43M), spec `:1139`'s 5 tasks and 4.4M, the 38 deliverables, the chain 4C to 5A to 5B to 6 to
7, D1 to D10, the rclone v1.74.3 help quotes, the mechanics lens's numbers, most code pointers,
and all nine rulings' yes/no form.

## Blocker

1. **Ruling 4 omits `backup`'s exit 69** (`:119-128`). "An operator reads one table for both
   verbs" is false: `backup` exits 69 when backup integration is disabled
   (`cmd/dubplate-server/backup.go:72`, help at `:100`), and restore's R-6bii9 row exits 1 for
   the same condition. Rewrite the reason to state that `backup` also exits 69 (disabled) and
   75 (a backup already running, `backup.go:101`), that restore has no 69, and that the
   disabled-integration row differs between the verbs. Optionally offer: "If you want the
   tables identical, rule that row separately: restore then exits 69 on a disabled
   integration."

## Warning

2. **The discard rows may not cover Task 5B's crash** (`:791` against `:1598`). A crash
   between step 5's two calls with `--discard-current` leaves `db/` empty and the incoming file
   in `staging/`; "any of the three files still in `db/`" reads as "at least one", so the state
   falls to the refusing "any other combination" row. Rewrite `:791`: `| a discard | none, some,
   or all of the three files still in db/ | the rest moved | present | move each file still in
   db/, then the rename and the fsyncs |`. Rewrite `:1598`'s last cell: "the first discard row
   with every file already moved: the rename, then the fsyncs, then the discard's completion
   tuple."
3. **Stale lane and worktree premises.** 6b-i merged at `6f2c3e0`, its close is recorded at
   `f24c77b`, both lanes are closed, and the worktree is gone. Fix `:16` (cite HISTORY's 6b-i
   entry, not the worktree drafts), `:25-35` (pointers read against `33ceff5`, which master
   contains; Task 0 re-resolves on master; drop "5c-ii is committing"), `:299-300` (6b-i's
   close recorded about 0.39M; the plan keeps its higher 6a-based estimate), `:360-362` (5c-ii
   is on master, so a fresh worktree needs `npm ci --prefix web` before its first gate),
   `:1877-1881` and `:1963-1965` (lane 1 is free; whichever code lane opens next), and
   `:2030-2032` (both "Not verified" bullets settled at `f24c77b`).
4. **Exit gate item 9 says "R-6bii7's three tables"** (`:1848`); R-6bii7 has two (`:676`,
   `:704`), as Task 4A says. Rewrite: "R-6bii7's two tables."
5. **Spec rulings and plan Rulings share numbers** (`:429`, `:837-838`, `:944`, `:1735-1736`).
   Write "spec ruling 7", "spec ruling 11", "spec ruling 12" at each site.
6. **Ruling 9's question is hard to parse and calls `--abandon` a verb** (`:183-192`; Task 7
   `:1775` and the fold record call it a flag). Rewrite: "**9. When `restore --resume` refuses a
   record, is the way out a runbook procedure owed to rung 16, rather than a new
   `restore --abandon` flag?**" and "verb" becomes "flag".
7. **Ruling 6** (`:144-162`) offers two alternatives a yes/no cannot select, glosses an rclone
   quote with "whichever side is newer", and has a 70-word sentence. Rewrite the reason: plain
   `rclone copy` never deletes ("Doesn't delete files from the destination", `rclone copy
   --help`, v1.74.3) and transfers every non-identical file, "testing by size and modification
   time or MD5SUM", so it replaces a differing local file even when the local copy is newer.
   `rclone copy --help` does not list `--backup-dir`; `rclone help flags` lists it in the Sync
   group as "Make backups into hierarchy based in DIR." The fold probed it (fold record,
   "Probes"). Rewrite the close: "No runs plain `rclone copy`, and a differing local file is
   overwritten and lost. Instead of yes or no, you may name `--ignore-existing` (keeps every
   local file, never restores the backup's version) or `--update` (keeps only local files newer
   than the backup's). Either changes Task 6's differing-file row."
8. **Ruling 7 packs three questions** (`:164-174`). Rewrite: "**7. Should restore accept an
   empty library root on its parent's filesystem (possibly an unmounted mount point), print
   whether the root is a mount point, and leave the check to a runbook line owed to rung 16?**"
   Builds: "Yes prints the resolved root, whether it is a mount point, and the free space the
   library check read. No refuses an empty root on its parent's device unless
   `--root-not-a-mount` is given, a spec deviation that fires on every default layout's first
   restore." Drop the "strikes revision 1's claim" clause.

## Suggestion

9. Ruling 1's split count (`:76`): "The spec priced five tasks. The plan now carries ten build
   tasks (1, 2, 3, 4A, 4B, 4C, 5A, 5B, 6, and 7), five past the spec's count."
10. One range per claim: `write.go:51-52` (pool), `:52-54` (busy timeout), `prune.go:202-220`
    (`isLocalCopyName`, cited as `:204-222` at `:430` and `:763`).
11. Split R-6bii4's 26-line error paragraph (`:515-540`) into status per code, the six request
    refusals as a list, and the decoding rules; make the 110-word completion tuple
    (`:692-700`) a bullet list.
12. Rewrap the spliced lines at `:77`, `:159`, `:519-525`, `:590`, `:1953`.

Verdict: the register's plausible author, with no machine tells. Its defects are factual: fix
Ruling 4 before Geoff rules, and the discard row and stale premises before execution.
