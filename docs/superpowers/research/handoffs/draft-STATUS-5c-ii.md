# Draft STATUS update: rung 5, pass 5c-ii close

Draft only. The conductor folds this into `docs/STATUS.md` (present tense, 60 lines or fewer)
and does not commit it as a standalone file.

## Current state

Rungs 1b, 3, 4, 8, rung 5's 5a through 5c-i, and rung 6's pass 6a are closed. **5c-ii is closed**
(the admin gate, the roster API, and the `web/` tree born gated; HISTORY, top entry): Tasks 1-7,
the simplifier, and the three pass-end reviews landed at `2424b9a` and `71de914`; the review fold
closed in two rounds (`d236d03`, `cc4649c`, `1306ae3`, `3acbdc0`); spend was about 8.7M against a
9.4M ceiling, past the 7.52M flag with Geoff's approval to continue. Beets 2.14.1 is adopted and
closed. The visual identity spec (revision 1.2, `782d864`) is complete and is 5c-iii's standing
input.

## Next action

**Resume prompt:** author the 5c-iii plan (fonts, the three real pages, the rendered gates, the
render-and-read, the identity spec's component layer) against 5c-ii's carried items below, then
run its full `spec-plan-review` cycle to "awaiting owner approval". 6b-ii and the matcher
word-boundary plan continue in parallel as document lanes.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | CLOSED on 5c-ii; free for the next code lane |
| 2 | worktree `.claude/worktrees/lane-2-6b-i`, branch `lane-2-6b-i` | see 6b-i's own STATUS lines |

## 5c-iii handoff

Full detail: `docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md`, "The 5c-iii handoff"
and "What stayed open" sections. In brief: 5c-iii builds on 5c-ii's routes, gate, roster API
outcome values, header step, and CSP builder; reads the visual identity spec 1.2 as standing
input; and carries the theme toggle, the `theme-color` metas, the mark and favicon, the ink
utilities, the rendered first-paint proof, the four ruled fonts, the named component CSS outside
DaisyUI's variables, and the branding-flattens-both-palettes rule. It also carries the a11y
reviewer's WCAG 1.4.11 question, Svelte S11 (reproducible builds) and S12 (absolute hrefs), and
security I2 (the login/link page's token-consumption discipline).

## Open items (carried)

- No deploy until 5d; `deploy.sh` must refuse a dirty `web/` and build the bundle from a clean
  tree before `go build` (5c-ii close).
- To 5d: the theme keys' `doctor` row, Ruling 6's old-browser `Referrer-Policy` residual, the
  branding-in-both-schemes runbook line, plus the prior carries (the "`doctor` exits 0" row,
  `ReadWritePaths`, the nine keys, the trusted proxy, `member add` argv, kind-mode enumeration,
  the mail-ceiling lever, and 6a's owed lines).
- To rung 12: security M2 (the roster API shares an origin with the Navidrome fallthrough),
  Ruling 7's availability question, re-validation of every existing roster row's
  `streamer_username` principal, and the header step's fallthrough reconciliation.
- Security L5 (no per-actor roster-POST throttle) and L6 (n5 whole-roster refusal doctor row)
  carry with no pass assigned yet.
- Eleven errata owed, drafted paste-ready in the HISTORY draft's "Errata owed" section: the rung 5
  spec's second-split erratum, the web surface standard's section 15 line, the "failing config
  load" row (R-5cii12), the embed-path CSP sentence (R-5cii20), R-5cii3's 404-tracking withdrawal,
  DL R3.6 and section 3's `gruvbox` naming (R-5cii22), W5.10's toggle split (R-5cii22), W5.9's
  clarifying erratum (A11), the staleness item (closed, not carried), the 5c-i HISTORY n5
  sentence (R-5cii5), and the section 19 row for the identity spec's withdrawal of W1.3.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.14.1`, mediafile `0.17.0`,
  unidecode `1.4.0`, lap `0.5.13`, `modernc.org/sqlite v1.59.0`. Spend to date: HISTORY, per pass.
