---
name: admin-identity-in-copy
description: dubplate has no admin display-name field; user-facing "ask the admin" copy uses Config.AdminEmail directly, never a name
metadata:
  type: project
---

dubplate's `internal/config.Config.AdminEmail` is the only admin identity the
codebase carries. `internal/serve/serve.go`'s `rosterAbsentFailure` and
`warmRosterAbsentPage` already render "the admin (<email>)" straight from
`AdminEmail`, no separate display-name field. `config/contributors.yaml` has
no `role: admin` entries (checked live 2026-09-01), so there is no roster
fallback to derive a name from either.

**Why:** a U13 fold task (F3.4) asked to replace a hardcoded "Geoff" in
`config/reject-email.txt` with a config-driven value and floated adding a
display-name field or a contributors.yaml roster lookup if nothing usable
existed in config. Neither was needed: `AdminEmail` alone, rendered as "the
admin (email@example.com)", matches the pattern serve.go already ships and
satisfies the adopter goal (no site-identity default outside
`internal/config`).

**How to apply:** if a future task wants a *name* rather than an email for
the admin, that is new scope (a real field addition, a template change on
`env/dubplate.env.template`, and a roster convention for `role: admin`) --
flag it rather than improvising, since none of that scaffolding exists yet.
