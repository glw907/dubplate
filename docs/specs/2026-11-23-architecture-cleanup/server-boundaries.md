# Checklist: task 3, `server-boundaries`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [ ] Repo-wide, high: "Setting readers in four shapes, homed in auth" (`internal/auth/handlers.go:75`, `internal/intake/config.go:326`, `internal/ops/admin/admin.go:50`, `internal/ops/local/local.go:43`, `internal/ops/local/config.go:36`).
- [ ] Repo-wide, high: "auth imports serve for a port check" (`internal/auth/config.go:378`).
- [ ] Repo-wide, high: "auth holds both credentials and the member roster" (`internal/auth/reads.go:38`).
- [ ] Repo-wide, high: "store reports on tables other packages own (records 0011 and 0013)" (`internal/store/inspect.go:173`, `cmd/dubplate-server/doctorrows.go:1`, `internal/backup/backup.go:18`).
- [ ] Repo-wide, high: "Doctor rows in nine shapes, with no shared contract" (`internal/store/doctor.go:26`, `cmd/dubplate-server/doctor.go:30`, `cmd/dubplate-server/doctor.go:185`, `internal/backup/doctor.go:28`, `internal/serve/doctor.go:198`, `internal/serve/doctor.go:80`, `internal/store/doctor.go:53`, `internal/webapp/doctor.go:18`).
- [x] a2adc03 Repo-wide, medium: "The free-bytes probe written twice" (`internal/backup/restore.go:539`, `internal/intake/staging/staging_linux.go:53`).
- [ ] Repo-wide, medium: "cmd reads auth keys directly, and auth declares the SMTP keys" (`cmd/dubplate-server/serve.go:216`).
- [x] a806cf7 Repo-wide, medium: ""Rung" in operator output" (`cmd/dubplate-server/doctor.go:71`, `internal/config/keys.go:101`).
