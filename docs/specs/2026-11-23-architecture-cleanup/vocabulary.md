# Checklist: task 1, `vocabulary`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [ ] Repo-wide, high: "One actor, three types" (`internal/ops/ops.go:24`, `internal/settings/write.go:51`).
- [ ] Repo-wide, high: "Four sentinels for "no valid session"" (`internal/ops/opshttp/mount.go:16`, `cmd/dubplate-server/intake.go:73`, `internal/ops/errors.go:132`).
- [ ] Repo-wide, high: ""Seam" names interfaces and structs" (`internal/auth/seam.go:21`, `internal/auth/seam.go:38`, `internal/ops/admin.go:126`, `internal/ops/admin/admin.go:64`, `internal/ops/opshttp/authorize.go:29`).
- [ ] Repo-wide, low: "ClientIP and ClientAddress" (`internal/ops/ops.go:88`).
- [ ] Repo-wide, medium: ""Store" in two senses" (`internal/auth/doctor.go:42`).
- [ ] Repo-wide, medium: "Resolve, Check, and Validate with several meanings" (`internal/serve/config.go:55`, `internal/intake/config.go:401`).
- [ ] Repo-wide, medium: "Two address normalizers" (`internal/config/config.go:260`, `internal/auth/roster.go:239`).
- [ ] Repo-wide, medium: "Package name `local`" (`internal/ops/local:1`).
- [ ] Repo-wide, low: "Two sentinel-text conventions" (`internal/auth/request.go:15`).
- [ ] Repo-wide, medium: "AdminConfig configures every member's session check" (`internal/auth/admin.go:12`).
