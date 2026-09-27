# Transliteration pass, Task 0 report (2026-09-26)

The read-only staleness pre-task ran at HEAD `4f4930b`, against the plan's base `1c012a8`, on
Go 1.27.1. **Verdict: launch.** All five blockers read "no".

1. **Lanes.** Lane 1 closed at `1a08eae`, lane 2 has none, and the photo lane is closed. No
   warm changes sit in plan territory.
2. **Pointers.** Every cited pointer still resolves. Nothing touches `internal/engine/matcher`,
   `internal/harness/conform`, `cmd/harness`, `scripts/deploy.sh`, or `tools/unidecode-table`.
   Some lines have drifted:
   - `main.go:127` is unchanged, with one `newRestoreCmd` line added.
   - `newDoctorCmd` still starts at `doctor.go:29`, but the store row's precedent is now at
     `:106-110`. Rung 6c added a settings-rows block and a `settingsFailed` exit term.
   - `lookupEnviron` is now at `doctor.go:412`.
   - `storeRow`'s not-checked path is now at `store.go:134-149`.
   - `shadow.go` is byte-identical: `shadow.Run` is at `:284` and `shadow.Reevaluate` at `:375`.
3. **Table and pin.** The `.tbl` sha256 matches, and the manifest still pins unidecode 1.4.0.
4. **PyPI.** Unidecode 1.4.0's wheel sha256 is `c3c7606c...0021`, and the release is not
   yanked.
5. **Overlay probe.** Exactly four packages reach `Transliterate`: matcher, tag, shadow, and
   conform.

## Notes for Task 3

- Only `shadow observe` and `shadow reevaluate` reach scoring. `shadow report` does not.
- No existing test runs `observe` or `reevaluate` past their refusals, so Task 3's
  installed-table tests are new coverage.
- The doctor row count is now computed as
  `... + storeRowLines + backupDoctorRowsOff/On + len(reg.Settings())` (`doctor_test.go:232`,
  `:342`, `:624`). The new unidecode row needs its own named constant in that sum.
- The plan's `DUBPLATE_STATE_DIR` test lines have each shifted by 2: `:417`, `:917`, `:928`,
  `:939`, `:994`, `:1047`, `:1079`, and `:1119`. Two new sites, `:1155` and `:1194` (6c's
  seed-rule tests), need a check for any scoring path that needs an installed table.
- The "full valid wall" case now starts near line 39.
