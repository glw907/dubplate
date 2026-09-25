# Draft: `docs/STATUS.md` after 6a merges

Line count of the body below (the part that replaces `docs/STATUS.md`,
excluding this draft's own header and this note): produced by
`sed -n '/^# dubplate status/,$p' draft-STATUS-6a.md | wc -l` once
pasted standalone. Drafted assuming 6a's merge ritual (rebase,
in-worktree gate, `--ff-only`, master gate, architecture reads) is
clean; the conductor updates commit ids if the merge changes them.

---

# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a, 5b-i, 5b-ii, and 5c-i are closed. **Rung 6's pass 6a (the
store foundation) is closed**, merged from lane 2 (`aa3b13e`..master). `internal/engine/store`
lands: the leaf and its lock, three SQLite connection classes plus the inspect class, one write
method with its recording transaction, the migration runner with pre-migration snapshots, the
first three migrations with the schema-invariant test, and the three open modes. `doctor` and
`check` are the first production callers; nothing in production creates or migrates a store yet.
Spend about 3.5M plus close/merge overhead against the 6.8M ceiling (5.44M flag); zero attended
interactions after approval. Eleven owed errata (E1-E11) are drafted, not yet applied to the spec.

## Rulings taken (Geoff, 2026-09-24)

- **The 6a plan is approved**, all four rulings yes. Its eleven errata E1-E11 are applied.
- **5c-i's flag crossing is accepted**; the stricter shared-namespace rule is declined.
- **Follow beets on the "ft" fix**: the oracle pin moves to the 2.14 line (`6d4dd05`).
- **The layout carries disc numbers**, a `1-01` prefix on multi-disc albums only.

## Next action

1. **6b-i's plan is in drafting** (`ops/local`, the write method's export, `backup`, `leases/`,
   the staging-stager rule; see the 6a close's handoff). Just-in-time, not front-loaded.
2. **5c-ii runs in lane 1** (approved 2026-09-24, all seven rulings yes; 9.4M ceiling). Segment A
   in flight. Segment B is held for the visual identity tokens (spec pending, session dubplate-04).
3. The architecture cleanup pass gains 6a's reads, joining the staged comment-cleanup task.

## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | OPEN on 5c-ii since 2026-09-24 |
| 2 | worktree `.claude/worktrees/lane-2-6a`, branch `lane-2-6a` | CLOSED; 6a merged; removable once 6b-i's lane is chosen |

## Open items (carried)

- No deploy until 5d: the env template selects `magic-link` without its nine keys.
- To 5c-ii: the admin gate, the roster API, serve wiring, `web/`, security n5 on the web write path.
- To 5d: the "`doctor` exits 0" row, `ReadWritePaths`, the nine keys, the trusted proxy, the runbook rows.
- To 6b-i: the seven handoff items (surface, intent-record rule, staging rule, checkpoint busy
  row, VACUUM INTO's WAL set, discarded-connection rule, runbook lines). To 6b-ii: the intent
  record's format. To 6d: the migration-wait key, serve's caller, the settle-loop bound.
- To rung 12: no principal maps a non-admin onto a streamer admin. To rung 7: the path validator
  (R-6a9), archive uploads.
- 94 declared coverage gaps (86 carried plus 6a's 8), none blocking. Staged legs warn: 325
  process citations, 128 exported-unused. Fork B: 1.0 is FLAC-only, group non-FLAC carried.
- Pinned: Navidrome `0.63.2`, cloudflared `2026.8.2`, beets `2.13.1`, mediafile `0.17.0`,
  unidecode `1.4.0`, lap `0.5.13`; SQLite driver `modernc.org/sqlite v1.59.0` (6a). Spend: about
  61.6M plus 9.1M, 6.4M, about 8.3M, and about 3.5M+ (6a) across the windows above.
