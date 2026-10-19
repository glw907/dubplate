# Draft STATUS update: rung 5, pass 5c-iii close

Draft only. The conductor folds this into `docs/STATUS.md` (present tense, 60 lines or fewer)
and does not commit it as a standalone file.

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-ii, and rung 6's 6a and 6b-i are closed. **5c-iii is
closed** (fonts, theme data, the rendered leg, the document head, the component layer; HISTORY,
top entry): Tasks 1-5, the simplifier, and the three pass-end reviews landed by `944e235`; the
review fold closed across `81988b1`, `2c880db`, `75fde16`, `8588dd0` (fold record `7ca3d53`),
with a `fix`-round diff-reviewer verdict resolved at `db1e111` (accepted by a second
`diff-reviewer`, which reproduced both red-first proofs); three merge-side architecture reads
ran at `db1e111`, all "workmanlike". Spend was about [SPEND-PLACEHOLDER, conductor fills]M
against a 10.8M ceiling. The visual identity spec is at revision 1.3 (`45b80cb`); three of this
pass's plan-named errata are already landed there.

## Next action

**Resume prompt:** run the just-in-time refresh read of this plan's Tasks 7 through 11 against
this close and its review fold (conductor action 5, per Ruling 1's declared just-in-time
deviation), fold it, then run 5c-iv's own Task 0 and Tasks 7-11 in lane 1. Plan
`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, ceiling 8.5M. The rendered leg's
script deadline is now 300s; a standalone rendered-test run measured 107.4s, the full heavy gate
about 3-4 minutes at about 560 MiB peak.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (master) | OPEN: 5c-iii closed at this entry; free for the refresh read and 5c-iv |
| 2 | worktree `lane-2-6b-ii-b` | see 6b-ii-b's own STATUS lines |

## 5c-iv handoff

Full detail: `docs/research/handoffs/draft-HISTORY-5c-iii.md`, "What stayed open"
and "5c-iv handoff" sections. In brief: 5c-iv builds on 5c-iii's rendered harness
(`cmd/dubplate-server/rendered_test.go`, `web/scripts/rendered/run.ts`), the `Makefile`'s `web`
target, and the component layer (`web/src/lib/styles/components.css`) — the four files crossing
Ruling 1's cut. It carries forward: the touch-target rule seeded at 20px (Ruling 8, full
render-and-read in Task 10); the `label-night`/`base-300` keyline's `contrast-exception` marker,
live until a page places a control there; svelte S12/a11y S9's stub-copy routing to Task 8; L2's
`{@html}` note to Task 7; and a11y S7's mono-utility and header-hover routing to Task 9 and rung
7. `rendered_test.go`'s `linkRequests`/`consumptions` fields are unset (`checkRateBudget` cannot
fail yet); Task 8 is the first task to spend a link, and the refresh read decides how it wires
these.

## Open items (carried)

- No deploy until 5d (unchanged from 5c-ii's carry).
- Nine errata owed, drafted paste-ready in the HISTORY draft's "Errata owed" section: the rung 5
  spec's third-split erratum (5c-iii into 5c-iii/5c-iv), the web surface standard's matching
  section 15 line, DL R5.5's family-count clause, W4.1's both-schemes line (now binds at rung 5,
  not rung 7), the rung 5 spec's admin-checkbox line, R-5cii8's members-row correction
  (`no-store`, not `private, no-cache`), R-5cii22's cookie `Max-Age`, and VI `:128-131`'s
  functional-branding-neutral carve-out (Ruling 10). Three plan-named errata (VI's clause
  number, Permanent Marker's license, the header halo) are already landed at `45b80cb` and need
  no further edit.
- a11y S2 (per-variant `color-scheme`) and a11y S4 (component-layer placement) are recorded, not
  folded: both are erratum-owed to VI and the web surface standard respectively, listed above
  with the nine.
- Latent, unfixed, and not yet gated: outline/dash button *text* still paints DaisyUI's raw
  `--btn-color` (label secondary/warning 1.83:1, label-night neutral 1.09:1, label-night primary
  3.18:1) — route to 5c-iv's refresh read as a precondition for Tasks 8 and 9, else rung 7; the
  ink border shows over DaisyUI's hover fill on outline buttons; a wrapped `<label class="input">`
  inside a status alert halos the inner input, not the wrapper; `settleFocusTransitions` awaits
  with no bound; root-absolute's `srcset` pattern false-positives on a base64 data URI starting
  with "/"; the fixture covers `.btn-outline` but not `.btn-dash`.
- To the architecture cleanup pass: this window's three reads (`internal/serve`,
  `internal/config`, `cmd/dubplate-server` at `db1e111`), leading the queue with
  `internal/serve`'s unused trusted-proxy machinery, `internal/config`'s stale package doc, and
  `cmd/dubplate-server`'s hand-agreed magic-link key names across four files.
- To 5d-i Task 1: `internal/serve`'s `NewTrustedProxyConfig` has no production caller; its Task
  0 must read the architecture-read finding before wiring trusted-proxy mode.
- To 6c Task 0/2: `internal/config`'s package doc calls backup an integration class, which is
  stale; do not build on it.
- To rung 7: svelte S13's input-inset-shadow and `label-night` keyline-vs-specimen observations,
  for the uploader's attended design review; the mono-utility surfaces (a11y S7).
- To rung 12: security M2 (CSP scope under a same-origin fallthrough streamer).
