# STATUS draft: the transliteration pass close (patch to docs/STATUS.md)

Patch form: each block replaces or extends the named section of the current
`docs/STATUS.md`, so no carried item is dropped. HISTORY's top entry holds
the detail.

## Current state (add to the closed list)

**The transliteration pass** closed 2026-10-03 at `5530cbb`. No binary carries
Unidecode data. A scoring verb refuses to start until `dubplate-server
unidecode install` has built the table into `$DUBPLATE_STATE_DIR/unidecode/`,
and `doctor` reports it.

## Next action (replace the overnight paragraph and the Resume block)

Fold `research/handoffs/draft-HISTORY-translit-table.md` and this draft (must-keep
list), run the exit-gate checks, dispatch one `diff-reviewer` on the fold, and
push (gitleaks first). Then the photo integrity gate line, then 5d-i
(approved), under a new window grant. Box read-only; nothing deploys until 5d.

## Lane registry (replace lane 1's row)

| 1 | main tree (master) | transliteration CLOSED at `5530cbb`: segment A `wf_581a3dd5-5c5` and `wf_45c82967-ff9`, simplifier `aeea780`, Task 3c `wf_560e7982-63a` (after `wf_ca338e5b-9ed` stopped with no edit); about 3.66M against the 3.3M ceiling (flag 2.64M crossed under the overnight grant); lane free |

## Open items (add)

- **To 5d's runbook (staged deploy and table).** Run `deploy.sh --check` first
  and read the staged install's `would run: systemd-run ... unidecode install`
  line. The deploy stages the binary at `/usr/local/bin/dubplate-server.new`,
  installs the table under `systemd-run`, and only then swaps. A failed
  install leaves the old binary live and scoring, and leaves `.new` for the
  next deploy to overwrite. Rollback is redeploying the previous commit. The
  box's systemd 257 has not yet accepted the property list (the man pages and
  a workstation run on 259 say yes); the first deploy proves it. The
  workstation's verified wheel cache (`make unidecode-wheel`'s output) is the
  estate's offline copy and the `--from` source if PyPI ever deletes the 1.4.0
  release.
- **Workstation prerequisite:** `make unidecode-wheel`, once per machine,
  before `make test`. A missing wheel fails the tests naming that command; the
  gate never fetches. Go's test cache does not track the wheel, so prove
  "wheel absent" with `-count=1`.
- **To `give-back` and `packaging`:** the git-history question for the deleted
  `unidecode-1.4.0.tbl` (in every commit before `e8e158d`), the process-wide
  holder's tension for a published library, and the Unidecode-derived items
  the repo still carries (HISTORY, "Routed, not decided here"), all for the
  lawyer's read before the first stranger-installable release.
- **To the architecture-cleanup pass:** the ROUTE items in
  `research/2026-09-27-translit-architecture-reads.md`, including the four
  `TestMain` shim copies and the four copies of the owner and leaf checks.
- **Gate note:** a gate run once failed on a `/tmp` disk-quota error and the
  identical rerun passed. Treat it as a lost run, free the space, re-issue.
- **Staged legs (replace the counts):** 323 process citations (was 325) and
  129 exported-unused (was 128; the added entry is the test helper
  `unidecodetest.LoadCachedTable`, expected to stay on the list).
- **Pinned (add):** the Unidecode 1.4.0 wheel and table sha256 and the 190
  data modules live in `internal/engine/matcher/unidecode/pins.go`.
