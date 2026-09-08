---
name: commit-only-named-paths
description: In this repo, commit with `git commit --only <paths>`; a plain commit sweeps in pre-staged unrelated changes
metadata:
  type: feedback
---

Stage and commit with `git commit --only <paths>`, never a bare `git commit` after
`git add`.

**Why:** dubplate's working tree routinely carries unrelated in-flight edits from
another pass, and some of them are already *staged* in the index (a `D ` deletion,
for instance). `git add <my files>` then `git commit` commits the whole index, so
the unrelated staged change rides along silently. This happened on 2026-09-01 with a
staged `config/filebrowser.yaml` deletion; the fix was `git reset --soft HEAD~1`
followed by `git commit --only <my files>`, which left the foreign staged deletion in
the index untouched.

**How to apply:** any commit in this repo. Check `git status --short` for a staged
column entry (first character non-space) on a file you did not touch before
committing, and verify with `git show --stat HEAD` afterwards.

**Recurrence (2026-09-01, same day, F2 task):** read this memory at task start, then
still ran plain `git add <my files>` + `git commit -m ...` with no `--only` and no `--`
pathspec, and the same `config/filebrowser.yaml` staged deletion rode along again.
Reading a memory is not the same as checking its condition against the live `git
status` immediately before the commit call — do the `git status --short` check as a
literal step right before typing the commit command, not from memory of having read
the rule. If caught immediately after a bad commit (before reporting done and before
any further commits), `git reset --soft HEAD~1` undoes it with the correct files still
staged, then `git commit --only <paths>` redoes it clean. That is cheaper and leaves
less history noise than a second corrective commit restoring the swept-in file.
