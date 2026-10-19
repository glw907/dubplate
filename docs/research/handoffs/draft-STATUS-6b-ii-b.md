<!--
Draft STATUS.md delta, modeled on draft-STATUS-6b-ii-a.md's shape, as it
would read once 6b-ii-b merges and lane 2 closes. The conductor folds
this against the real docs/STATUS.md, which has moved since this draft
was written (5c-iii and other passes closed); this draft states only
6b-ii-b's own delta, not a full rewrite. The conductor fills the
bracketed fields (merge SHA, actual spend, architecture-read routing)
during the fold. Line count below ("## Current state" through EOF) is
at most 60, per the plan's Task 8 acceptance criterion.
-->

# dubplate status (6b-ii-b delta)

## Current state

**6b-ii-b** (the offline `restore` verb, `<merge-sha>`, `<actual>` against
7.53M) is closed; HISTORY, top entry alongside 6b-ii-a. Rung 6 (`6a`,
`6b-i`, `6b-ii-a`, `6b-ii-b`) is now fully closed.

## Next action

1. **6c plans and runs next** in lane 2 (approved at revision 2.2,
   `50a1cf8`, 5.4M ceiling, flag 4.32M), then the transliteration pass
   (approved, 3.3M).
2. **To 6d** (not yet planned): `ServeOptions.Reconcile` and
   `backup.Reconcile` as serve's own wiring (`OpenServe`'s `Reconcile`
   parameter has no production caller yet); a store identity so restore
   can refuse another store's snapshot; the network actor's
   `member:<principal>` form, if 6d's wiring needs restore to carry one.
3. **To rung 16** (the runbook): the lines in HISTORY's 6b-ii-b entry —
   a kept pre-restore copy is never auto-pruned and the operator removes
   it by hand; a mount-point library root's displaced directory sits
   uncounted on the parent filesystem; an unreadable non-`lost+found`
   directory under the root refuses a displacing copy; the manual
   procedure for a resume a record refuses.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 2 | worktree `lane-2-6b-ii-b` (branch `lane-2-6b-ii-b`) | CLOSED; merged at `<merge-sha>` |

## Open items (carried)

- No deploy until 5d (unchanged by this pass).
- Errata owed (this pass): spec criterion 16 and line 712 (a pre-restore
  copy is not auto-pruned, unlike a pre-migration copy); the plan's
  privileged-fields row for a pre-restore copy; R-6bii11's displaced-
  directory paragraph and Task 6's summary (`--backup-dir` applies on
  every resume, not only a first run under `--into-nonempty`).
- Carried non-blocking (not fixed this pass): the reconcile walk
  descends into `lost+found`; `Reconcile` does not check the root is a
  directory; `makeDisplaced` creates its directory at step 6;
  `sayKept` prints before `Complete`; a first-run symlink race inside
  rclone's own directory read; `fetch_test.go`'s duplicated stand-in
  names; `cmd/dubplate-server/backup.go`'s second remote-name validator.
- To rung 12: the streamer database's place in a restore (0.8 forward
  hook), unchanged.
- 99 declared coverage gaps, none blocking. Staged legs warn: 326
  process citations (unchanged), 126 exported-unused (up one from
  6b-ii-a's 125). Fork B: 0.8/0.9 FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`,
  mediafile `0.17.0`, unidecode `1.4.0`, lap `0.5.13`,
  `modernc.org/sqlite v1.59.0`, rclone v1.74.3. Spend: HISTORY.
