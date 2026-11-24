# Checklist: task 2, `wire-contract`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [ ] Repo-wide, high: "Ids outside `internal/id` (record 0012)" (`internal/intake/submission.go:112`, `internal/id/id.go:47`, `internal/identity/identity.go:202`, `internal/intake/seam.go:350`, `internal/intake/worker.go:73`, `internal/ops/ops.go:104`, `internal/ops/local/jobs.go:99`).
- [ ] Repo-wide, high: "A setting is "name" in one message and "key" in the next" (`internal/ops/admin.go:67`).
- [ ] Repo-wide, medium: ""Mark" for a field of type Sensitivity" (`internal/ops/admin.go:71`, `internal/ops/admin/settings.go:16`).
- [ ] Repo-wide, medium: "Wire enums in kebab-case and snake_case" (`internal/settings/evaluate.go:20`).
- [ ] Repo-wide, medium: "Reason, verdict, and refusal for one value" (`internal/ops/submissions.go:87`).
- [ ] Repo-wide, medium: "TS copy constants named for a narrower scope" (`web/src/lib/api/v1.ts:77`).
- [ ] Repo-wide, low: "Members, roster, and users" (`cmd/dubplate-server/member.go:42`).
- [ ] Repo-wide, low: "Job types shaped like wire operations (record 0003)" (`internal/ops/ops.go:135`).
