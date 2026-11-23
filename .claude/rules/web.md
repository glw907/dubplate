---
paths:
  - "web/**"
---

# Web

**Web** (`web/`): SvelteKit in static mode (Geoff, 2026-09-07;
supersedes the standalone rule): Svelte 5 runes, DaisyUI 5,
`adapter-static` with client-side rendering only, built to static
assets embedded via `go:embed` and served by the Go binary, the only
server. Server-only SvelteKit files (`+page.server.*`, `+server.*`,
`hooks.server.*`, form actions) are forbidden and the gate checks that
none exists. Clean, idiomatic SvelteKit is the bar, as in the cairn
family. Invoke `svelte-conventions` and `ts-conventions` before writing components or TS. `tus-js-client` pinned 4.3.x, from `intake`. Node is workstation
build-time only; nothing Node ships to the box. The login page was
the first page built under this rule. Push hard toward the stack's own idiom
(Geoff, 2026-09-01): reach for the DaisyUI component class and its documented
modifiers before a hand-rolled utility stack, never fight a component's layout
model, take colors only from theme tokens, prefer `$derived` to an effect that
mirrors state, and keep `$effect` for real side effects with cleanup. The idiom
wins over a finding's literal fix text, un-idiomatic code adjacent to a change is
in scope for it, and the exemplar is `~/Projects/cairn-cms/src/lib/admin-toolkit/`:
imitating it beats interpreting the rule. The look is the visual identity
spec (`docs/specs/2026-09-24-visual-identity-design.md`,
Geoff, 2026-09-25): every `web/` task reads it, and it outranks any
styling habit.
