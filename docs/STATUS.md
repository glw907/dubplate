# dubplate status

## Current state

Rungs 1b, 3, 4, 8, and rung 5's 5a and **5b-i** (the auth core) are closed; 5b-i merged at `1d4ed29`. R4's
exclusion denominator is UNREAD (box unreachable at 03:14 AKDT), owed before rung 15. The overnight window
(`plans/2026-09-12-overnight-window.md`) closed 8c and 5b-i at 9.1M of 10M.

## Next action

1. **5b-ii is executing** from `plans/2026-09-12-rung-5b-ii-auth-handlers.md` (revision 2.1, 7.5M ceiling,
   flag 6.0M). **Segment A closed at `aae338c`** (Tasks 1-3 accepted, about 1.7M): all five reproductions
   (B1, M1, M2, M3, M5) observed failing first; Task 1's extra CR/LF refusal in `splitSigned` is ruled kept
   (`Strict()` still ignores CR and LF); Task 2's second `fix` was ruled as one narrow round, the
   selected-source one-row test, accepted. **Segment B closed at `dfa7155`** (about 1.6M; pass total about
   3.3M, projected about 7M, under the 8.0M cut line, so no split). Rulings: `doctor` reports a missing
   roster as "missing"; `AddMember` refuses the admin flag on every channel, and 5c adds the CLI admin
   grant with its caller. Carried to 5c: rename `ErrRosterAdminChannelForbidden` when the CLI grant lands, and the last-admin
   lockout guard, atomic against concurrent self-demotion (configuration homes routes it to rung 5; no rung 5
   spec or plan carries it yet).
   Carried to 5d: the spec's "doctor exits 0" acceptance row (`spec:2332`) must run after the bootstrap
   verb. **Segment C closed** (Tasks 7-8 accepted first time; simplifier `105ad50`, its comment fix `95e7607`).
   **Segment D1 closed at `644dfba`**: Task 9 accepted after one comment round, **Task 10 accepted first
   time** (all twelve named mutations fired; test-first not kept, recorded). Pass spend about 5.6M of 7.5M.
   Carried to 5c from Task 10: the session-refresh call site (the seam's same-origin condition plus Task 2's
   clamped `RefreshSession`; no 5b-ii handler verifies a session) and the magic-link mail copy (a minimal
   placeholder today). **D2 reordered (conductor, 2026-09-23):** Task 11, then the simplifier over Tasks
   9-11, then the pass-end `web-auth-security-reviewer`, then Task 12's close, so the close records the
   security read and any fix it forces. Then the merge ritual (one lane, so gate, merge, gate).
   **Task 11 accepted first time (`c2613ba`); simplifier `08befea`. Pass-end security read at `46e37d1`
   (`research/2026-09-23-rung-5b-ii-security-review.md`): no blocker, five majors, fourteen minors; 5b-i's
   fixes hold except M4 (partial) and m10 (ASCII only).** Majors: cross-process roster writes lose and
   resurrect members; mailed links land outside the owned prefix (token to the fallthrough); a refused
   client's flood evicts a victim's per-address bucket; IPv6 per-client keys at /128; no trusted-proxy
   setting behind the tunnel (a 5d deploy gate). Spend about 6.3M, past the 6.0M flag: **Geoff ruled 2026-09-23**: majors 1-4 plus
   the relevant minors become 5c's security-fold Task 1; major 5 is a 5d deploy gate; 5b-ii closes on its
   ceiling. Task 12's close runs next, then the merge ritual.
2. **Critical path (conductor, 2026-09-23):** the web surface standard is ratified (Geoff, 2026-09-23, `94d8af3`), so 5c's plan
   is unblocked; configuration homes revision 3 is ratified too
   (2026-09-23; per-collection layout parity, conditional on a rung 12 Navidrome check) (gates rung 6's spec). Rung 6 runs as lane 2 beside 5c/5d once
   ruled. Rung 5 through 5d stays first, because the no-deploy carry blocks every box change.
3. Owed in `~/.dotfiles` (held: that tree had another session's uncommitted edits on 2026-09-23): the infra
   round's task 9, and a `pass-execute.js` stop-on-escalate option so a segment never builds on an
   unaccepted task (5b-ii's Task 5 was committed over Task 4's open escalation).


## Lane registry

| Lane | Tree | State |
|---|---|---|
| 1 | main tree (`/var/home/glw907/Projects/dubplate`, master) | ACTIVE. 5b-ii segment D2 (Task 11, then the close); D1 closed at `644dfba` |
| 2 | worktree removed, branch deleted | CLOSED. 5b-i merged at `1d4ed29`; reopens on rung 6 once its spec is ruled and its plan approved |
| doc | main tree | CLOSED. Rung 6 spec revision 2.2 committed after four lenses, two folds, a verification read, and a prose review; ratified by Geoff 2026-09-23 (all fourteen rulings yes), five passes, about 28M; errata owed |

## Open items (carried)

- **No deploy until 5d**: the env template selects `magic-link` without its nine required keys, so a render
  would refuse every server verb, shadow included (merged-range review, 2026-09-12). `newProductionRegistry`
  in `main.go` is the one registration line, a contested surface for the next two-lane run.
- The configuration homes spec (revision 3) routes tasks into rungs 6, 7, 9, 12, 14, 16 and the 2.0 list;
  the rung 5 spec's eighteen errata land with 5b-ii's plan prep (HISTORY). Seventy-three declared coverage
  gaps (rungs 3, 4a, 4b, 1b, 8a, 8b, 5a, 8c, 5b-i's six), none blocking. Fork B ruled FLAC-only for 1.0
  (ROADMAP `formats-1-1`); the group non-FLAC question is carried.
- Staged gate legs (warn today): 313 process citations in shipped Go comments (a comment-cleanup task flips
  `CHECK_CITATIONS=fail`); 134 exported-unused symbols (before `CHECK_EXPORTED_UNUSED=fail`; detail in
  HISTORY's 5b-i process-changes note).
- Leftovers: four-knob wall registration; `AdminEmail`'s contact row; five 3b items; `profile_hashes`;
  response-status extension; `Case.Extra`; `slices.Clone` (rung 9); the corrupted untracked 2026-09-04
  evidence doc; two review albums on the box; `FILEBROWSER_ADMIN_PASSWORD`; stale failed units;
  `DUBPLATE_SITE_NAME` re-mint.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend: about 61.6M plus the window's 9.1M; 5b-i's overage sits on R23.

