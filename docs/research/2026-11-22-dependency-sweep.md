# Dependency sweep, 2026-11-22

Scope: two pins of the `dev-friction` task of `2026-11-22-overnight-design.md`, taken one package at a time: the store's
SQLite driver (`modernc.org/sqlite` with the `modernc.org/libc` its `go.mod` names) and `certifi` in the oracle venv. No
sweep of every outdated module ran, so the next full sweep starts from `2026-10-12-dependency-sweep.md` plus this
record's "after" column. Backlog #129 and #175 close here.

## Before and after

| Package | Before | After | Kind |
|---|---|---|---|
| modernc.org/sqlite | v1.60.1 | v1.63.0 | Go module, minor; the version probed and verified on 2026-11-22, which is also the day it was released (the backlog's `v1.62.1` does not exist) |
| modernc.org/libc | v1.77.1 | v1.80.1 | Go module, indirect; the pin v1.63.0's `go.mod` names, and the driver's changelog requires a downstream `go.mod` to match it |
| certifi | 2026.7.22 | 2026.11.20 | PyPI, via `uv pip compile --upgrade-package certifi`; line 5 of `tools/venv/requirements.txt` alone changed |

`go mod tidy` also moved other lines of `go.sum` (the driver's own transitive requirements) and no other line of
`go.mod`.

## Survey

### modernc.org/sqlite v1.60.1 to v1.63.0

Read from `CHANGELOG.md` in the module cache at v1.63.0, against the eight files that import the driver, all in
`internal/store`: `conn.go`, `driver_test.go`, `migrate.go`, `migrate_test.go`, `open.go`, `conn_test.go`,
`write.go`, and `write_test.go`.

Features to leverage: none taken. `Fdatasync` (v1.62.0, opt-in, off by default), the SQLite authorizer (v1.61.0), `vfs.Register`
(v1.63.0), and `_time_format=string_no_monotonic` (v1.61.0) have no caller in this repo, and nothing here hand-rolls
them. They stay with the first spec that wants one.

Practices to change:

- **A statement waiting under `busy_timeout` stops once its context is done** (v1.61.0), and `Tx.Commit` does the same
  for the context given to `BeginTx` (v1.61.1). `Handle.Update` passes its deadline context to `BeginTx`, so a write that
  waits on another process's lock now returns the context's error when the deadline passes first, where it used to wait
  the full busy timeout and return `ErrBusy`. `internal/store/write.go` kept its documented contract by rewording the
  comment, and `TestUpdateDeadline` states the new behavior. `classifyErr` (`conn.go:223`) passes an error that carries
  no `*sqlite.Error` through unchanged, so the context's error reaches the caller whole. The callers that print it are
  covered in the pull request: `internal/intake/seam.go:130` and `internal/ops/admin/admin.go:121` answer a busy store
  and a passed deadline with the same words, and the operator verbs in `cmd/dubplate-server` print the raw error.
- **A zero-length BLOB reads as an empty `[]byte`, not nil** (v1.63.0). The one `[]byte` scan in the package is
  `readCollectionRoots` (`inspect.go:431`), and `root_path` is never empty, so nothing compares a scanned blob with
  nil.
- **Connections open with `SQLITE_OPEN_NOMUTEX`** (v1.63.0). It matters only to code that takes the SQLite handle from
  unexported fields. The store reaches the driver through `database/sql` and its `*sqlite.Error` alone.
- **A TEXT argument with a NUL reaches Go functions and virtual tables whole** (v1.63.0). The store registers neither.
- **`BeginTx` and `PrepareContext` with a finished context return its error** (v1.61.0), and a shared-cache `BeginTx`
  waits for a lock (v1.61.0). The store opens no shared-cache connection.

Gate risks:

- `internal/store/write_test.go`, `TestUpdateDeadline`: the case that assumed the old wait failed on the new driver at
  `write_test.go:601` and `:605`, and is rewritten. Probed on 2026-11-22 in two fresh copies: the whole Go suite's
  failure list differed between v1.60.1 and v1.63.0 by that case alone.
- `commitMigration` (`migrate.go:377`) matches `no transaction is active` on `Tx.Commit`'s error. The v1.61.1 change
  stops only the wait for a lock, so that text is unchanged, and `migrate_test.go` covers it.
- SQLite moves from 3.53.4 to 3.54.1 with v1.61.0, and the vendored `lib/` moves again in v1.62.0 with SQLite staying at
  3.54.1. The store's migrations run under the gate and `invariant_test.go` and `schema_test.go` read their results.

### certifi 2026.7.22 to 2026.11.20

Features to leverage: none. Practices to change: none. The package is a TLS root bundle that `requests` loads for
beets' network calls, and nothing in this repo reads it. Gate risk: none, since the gate never builds the venv and no
test reads the bundle. Measured in the rebuilt venv: the bundle held 121 certificates and 12 lines naming `emSign` at
2026.7.22, and 113 certificates and 3 such lines at 2026.11.20, which is the removal Dependabot alert #1 asks for
(the emSign roots whose intermediate keys are compromised). The bundle holds TLS roots only, so `make derive` is not
needed and the oracle corpus's `corpus/manifest.json` records no certifi version.

## Held majors

None. Neither bump crosses a major.
