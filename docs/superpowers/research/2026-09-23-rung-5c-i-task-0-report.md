# Pass 5c-i Task 0 staleness report

HEAD `cb9ecc3`. Verdict: **LAUNCH**. The conductor ruled it on 2026-09-23 with no blockers.

## Blockers

1. The lane registry and the tree are consistent with a lane 1 launch: yes. There is one worktree, and no
   tracked change under `internal/`, `cmd/`, or `go.mod`. The Go tree is unchanged since `08befea`.
   The only warm edits are the concurrent document lanes' files under `docs/`.
2. 5b-ii has merged and its close is folded: yes. The fold is at `d82e9cf`. Since the `27cd7b5` draft,
   "The 5c handoff" has changed in three ways. `AddMember` now refuses `Admin` on every channel,
   service token included (`roster.go:329-331`), which Task 6 and R-5c13 already assume. The
   security-fold routing now lists its findings explicitly (`HISTORY.md:633-636`). The rejoin pointer
   is re-keyed to `CLAUDE.md:140`.
3. All thirteen routed findings still reproduce at HEAD, and none is already fixed. Their locations
   are M1 `roster.go:328,351-354,286-301`, M2 `token.go:161-169`, M3 `limit.go:78-86,163-181`,
   M4 `handlers.go:152,166`, m7 `roster.go:170`, m8 `identity.go:50`, m10 `auth.go:361,364`,
   m11 `doctor.go:268-289`, m12 `serve.go:300`, m13 `identity.go:75`, m15 `roster.go:346,212`,
   m17 `handlers.go:160,166`, and m19 `nonce.go:208-221`.
4. `internal/config` is untouched since `1d4ed29`: yes. Lane 2 has no executor.

## Notes

- Minor 6 is still routed to 5c in HISTORY (`:491`, `:635`). R-5c18 refuses it, so the close records the
  refusal in the scope ledger.
- `encoding/json/v2` compiles under go1.27.1 with no GOEXPERIMENT.
- The anchors resolve as follows. `newServeShell` is at `serve.go:212`, and the route constants are
  at `serve.go:41-43`. `HandlersConfig` has no Prefix field yet. The line-count test is
  `TestDoctorCmd_LineCountFollowsRegistry` at `doctor_test.go:163`. `newRootCmd`'s `AddCommand`
  block is at `main.go:106-112`.
- `AddMember` has callers only in tests. `List` has one production caller, at `handlers.go:238`.
  `ErrRosterAdminChannelForbidden` is used at `roster.go:330` and in tests.
- The spec pointers are intact. "Only path to the first admin" is at `:190-192`, inside the cited
  `:188-192`.
- The exit gate's no-credential test is `TestLogsCarryNoCredential`, at `handlers_test.go:771`.
