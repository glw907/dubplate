# Checklist: task 8, `harness-shape`

Each line is one finding from `docs/research/2026-11-23-architecture-read.md`, quoted to its first
sentence; the report holds the full finding and its fix. The task ticks each box as the spec's
"Done items every task carries" section says.

- [ ] Repo-wide, medium: "Two Violation vocabularies for one protocol" (`internal/opensubsonic/conformance:1`, `internal/harness/subsonic/xmlrule/check.go:16`).
- [ ] Repo-wide, low: "Analysis loaders named two ways" (`internal/harness/analysis:1`).
- [ ] Repo-wide, high: "cmd/harness hand-copies the harness case layout" (`cmd/harness/capture.go:42`, `cmd/harness/freeze.go:104`, `internal/harness/oracle/oracle.go:40`, `internal/harness/stub/stub.go:34`).
- [ ] Repo-wide, high: "opensubsonic/conformance sits apart from its XML twin" (`internal/opensubsonic/conformance/conformance.go:1`).
- [ ] Repo-wide, high: "Import rules enforced six ways" (`internal/harness/conform/arch_test.go:1`).
- [ ] Repo-wide, medium: "A flat harness command tree, and two meanings of "oracle"" (`cmd/harness/main.go:66`).
- [ ] Repo-wide, medium: "Recorder domain logic in package main" (`cmd/harness/recordoracle.go:225`).
- [ ] Repo-wide, medium: "repoRoot written seven times" (`internal/harness/oracle/oracle.go:81`, `cmd/harness/freeze.go:104`, `internal/harness/corpus_test.go:24`).
