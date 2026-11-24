# Checklist: task 2, `wire-contract`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [~] e946adb, held for ops, admin, and the wire; not held for intake's in-process `[]byte` ids, which task 5's `server-data` checklist carries as its own finding (`server-data.md:87`): Repo-wide, high: "Ids outside `internal/id` (record 0012)" (`internal/intake/submission.go:112`, `internal/id/id.go:47`, `internal/identity/identity.go:202`, `internal/intake/seam.go:350`, `internal/intake/worker.go:73`, `internal/ops/ops.go:104`, `internal/ops/local/jobs.go:99`).
- [x] b51632f: Repo-wide, high: "A setting is "name" in one message and "key" in the next" (`internal/ops/admin.go:67`).
- [x] b51632f: Repo-wide, medium: ""Mark" for a field of type Sensitivity" (`internal/ops/admin.go:71`, `internal/ops/admin/settings.go:16`).
- [x] 2edc6a2: Repo-wide, medium: "Wire enums in kebab-case and snake_case" (`internal/settings/evaluate.go:20`).
- [x] 5dddd13: Repo-wide, medium: "Reason, verdict, and refusal for one value" (`internal/ops/submissions.go:87`).
- [x] 5dddd13: Repo-wide, medium: "TS copy constants named for a narrower scope" (`web/src/lib/api/v1.ts:77`).
- [x] 5dddd13, 2edc6a2: Repo-wide, low: "Members, roster, and users" (`cmd/dubplate-server/member.go:42`).
- [x] e946adb: Repo-wide, low: "Job types shaped like wire operations (record 0003)" (`internal/ops/ops.go:135`).
