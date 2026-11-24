# Checklist: task 1, `vocabulary`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [x] 69028bb Repo-wide, high: "One actor, three types" (`internal/ops/ops.go:24`, `internal/settings/write.go:51`).
- [x] 6781afd Repo-wide, high: "Four sentinels for "no valid session"" (`internal/ops/opshttp/mount.go:16`, `cmd/dubplate-server/intake.go:73`, `internal/ops/errors.go:132`).
- [x] 507e342 Repo-wide, high: ""Seam" names interfaces and structs" (`internal/auth/seam.go:21`, `internal/auth/seam.go:38`, `internal/ops/admin.go:126`, `internal/ops/admin/admin.go:64`, `internal/ops/opshttp/authorize.go:29`).
- [x] 69028bb Repo-wide, low: "ClientIP and ClientAddress" (`internal/ops/ops.go:88`).
- [x] 2293db1 Repo-wide, medium: ""Store" in two senses" (`internal/auth/doctor.go:42`).
- [x] 2293db1 Repo-wide, medium: "Resolve, Check, and Validate with several meanings" (`internal/serve/config.go:55`, `internal/intake/config.go:401`).
- [x] edd31e3 Repo-wide, medium: "Two address normalizers" (`internal/config/config.go:260`, `internal/auth/roster.go:239`).
- [x] c21621d Repo-wide, medium: "Package name `local`" (`internal/ops/local:1`).
- [x] 288b1f3 Repo-wide, low: "Two sentinel-text conventions" (`internal/auth/request.go:15`).
- [x] 2293db1 Repo-wide, medium: "AdminConfig configures every member's session check" (`internal/auth/admin.go:12`).
