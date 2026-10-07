# Rung 6d architecture reads: the findings left for later (2026-10-07)

Seven `go-architecture-reader` dispatches graded the packages rung 6d touched, on master at `6db67b7`:
`cmd/dubplate-server` (workmanlike) and `internal/auth`, `internal/identity`, `internal/engine/store`,
`internal/backup`, `internal/settings`, and `internal/config` (each sound with nits). They returned 9 correctness,
32 architecture, and 33 polish findings.

Two fold tasks landed what 6d owns. `5079b38` folded all nine correctness findings, including lines 6d did not
write: one hardened invisible-rune predicate shared by `internal/identity` and `internal/auth`, no raw address in a
`net/mail` error, a relative library root refused at backup as at restore, reconcile judging files from the one
walk that follows no symlink, non-finite floats refused for thresholds, the unused `WallRow.Pending` seam removed,
one constant block for the secret and SMTP key names, and a nonce-retention test that asserts the reader serve
actually wires. `4de73e7` folded the architecture findings on lines 6d changed: `storeBusyTimeout` returns
`(time.Duration, error)`, doctor's rows print from one slice, `backupFactRow` became `doctorRow`, the proxy
re-parse is gone, and the reconciler shares restore's report head.

## Refused, with the reason

- `internal/auth` `VerifySession`'s shape: about 36 test references, two in `cmd`; retargeting is not mechanical.
- `internal/auth` `RunSweep`'s `now` parameter: it is the clock seam the sweep tests drive.
- `internal/settings`' kind-to-parser triplication: 6d only added `KindString`; the fix spans three packages.
- `internal/auth` `normalizeMemberAddress`: a roster test pins the call-site count by grep.

## To file (on lines 6d did not touch)

- **cmd:** the SMTP port and From refusals are plain errors, not `*sourceKeyError`, so the comment at
  `doctor.go:524-527` ("every refusal is keyed") is false (`auth.go:381,384,399,402`).
- **cmd:** the serve-port bound rides on the magic-link source's `Validate` hook, and threshold bounds are three
  hand-written checks; `config.Threshold` has no Floor or Ceiling (`serve.go:117-135`, `store.go:30-49`).
- **cmd:** doctor's store rendering belongs in a doctor-side file, and the route-name constants are split
  (`store.go:219-558`, `web.go:29-39`).
- **cmd:** the uid-to-username lookup and the `Stat_t` owner check are written twice, and `resolveActor` bypasses
  the `geteuid` seam (`member.go:166-177`, `unidecode.go:260-296`, `doctor.go:301-305`).
- **cmd:** `authHandlersConfig` takes ten positional parameters, and there are four shapes of one reader over the
  accessor (`serve.go:477-506`, `backup.go:108-121`, `store.go:113-121`, `auth.go:483-499`).
- **cmd (polish):** environment key literals repeated (`backup.go`, `restore.go`, `store.go`, `unidecode.go`); a
  duplicated prefix parse in `webRowStatus` (`doctorprobe.go:95-102`); exit statuses 75 and 130 declared twice.
- **auth:** `Bump` and `Subscribe` have no caller outside tests (`epoch.go:72,103,122-138,207`); `Seam.Read` has
  none (`seam.go:35-38,156`); a large exported surface has no consumer outside the package; `rosterAPI.writeWait`
  is a test-only field (`admin.go:79,169-171`); three `syncDir` copies (`nonce.go:359-372`, the store's
  `snapshot.go:206`, `atomicfile`) want one `atomicfile.SyncDir`.
- **identity:** `MaxAddressLength` and `ErrInvalidAddress` have no production caller; `DisplayName` and `Source` are
  unread, and `Valid`'s doc still says "non-empty source" though the code no longer checks it
  (`identity.go:19-23,29,203-207`).
- **store:** `waitLock` duplicates the `acquireBounded` loop and `acquireExclusive` has no production caller
  (`lock.go:57`, `open.go:386`); `Facts.SettingsGeneration` is unread, and `SettingsRow` duplicates
  `settings.Row` and its scan (`inspect.go:163,186,369-389,395-398`).
- **backup:** `ParseRemote` has no accessor and a flat error, with its alphabet duplicated (`remote.go:11-51`);
  backup rebuilds the store leaf's snapshot directory and prunes by name, which belongs on the store
  (`backup.go:209`, `prune.go:159-186`).
- **settings:** `Row` duplicates `store.SettingsRow` and the SELECT and scan are written twice
  (`evaluate.go:26-30,136-138,168-183`); the accessor's mutex is held across the database round trip
  (`accessor.go:88-112`), which the segment C security read's Low 1 fold made unreachable from a bad cookie but
  which still serializes every valid read; a second `Validate` call wraps with `%w` and can render `%!w(<nil>)`
  (`evaluate.go:264-266`).
- **config:** `OwnedWall`, `ReservedClasses`, and `Registry.Superseded` have no production caller, and four owned
  key literals are duplicated (`keys.go:28-30,98-101`, `registry.go:514-516`, `config.go:399-402`).

The citation and comment-density polish across these packages belongs to the staged citation leg.
