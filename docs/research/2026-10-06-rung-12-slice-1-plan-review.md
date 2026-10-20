# Rung 12 slice 1 plan review: contract, mechanics, and proportionality

Target: `docs/superpowers/plans/2026-10-06-rung-12-slice-1-study.md`, draft revision 1 at `981876e`. One lens, pass
class `docs`, folded by the reviewer into revision 1.1. Findings are ranked by consequence: correctness and stated
requirements only, with over-ceremony ranked by its token and clock cost.

## What was verified

- Every cited streamer-spec line was spot-checked at `981876e`. The spec is unchanged since `bdd1472`, so the plan's
  "cited at `0cb33ae`" holds. The checked citations are `:43-50`, `:266-271`, `:276-280`, `:285-287`, `:316-318`,
  `:327-328`, `:341-347`, `:374-376`, `:412-447`, `:563-569`, `:650-673`, `:724-736`, `:856-891`, and `:917-955`. All
  match except F11.
- The re-plan citations `:48-54`, `:113-141`, `:157-159`, `:180-194`, `:384-386`, and `:427` match. The survey
  citations `:131-151`, `:161-172`, `:261`, and `:267-268` match. The fold record's `:38` and `:127` match.
- Licenses, as the GitHub API reports them on 2026-10-06: Navidrome, gonic, Feishin, Rufin, Amperfy (`BLeeEZ/amperfy`),
  and Tempus are GPL-3.0. Melodee is MIT. `opensubsonic/open-subsonic-api` and `navidrome/website` are Apache-2.0. The
  clean-room rules' license premises hold.
- Workstation state: Feishin 1.17.0 and Rufin 0.16.6 are installed as system Flatpaks and listed in
  `~/.dotfiles/bluefin/flatpaks.txt:17,19`. `~/Android/` does not exist. `podman` and `ffmpeg` are on the path. Rufin's
  `Cargo.toml` depends on `oo7` 0.6.0 and `keyring-core`.
- `~/.claude/workflows/pass-execute.js` runs an implementer that commits under a gate per task, sequential by default,
  with an optional `parallel` mode in one tree. `model-economy.md` seats `diff-reviewer` at `claude-opus-5-5`, effort
  `medium`.
- The slice 2 plan does not exist yet on master, so the version hand-off was read from this plan alone.

## Findings

| ID | Severity | Location (rev 1) | Defect | Disposition |
|---|---|---|---|---|
| F1 | Major | "Rulings", Ruling 1 | Not a genuine fork. The approved re-plan already rules that "the per-slice roster walk is conductor-run on the workstation (desktop clients and the Android emulator). It is never an owner touchpoint" (`:140-141`). The plan's own recommendation agrees. The yes form also says it "installs Feishin", which is already installed. | Folded as settled item 9. It cites the re-plan, corrects the install list (only the Android SDK emulator and the Tempus APK are new), and states that slice 2 reuses the install. "Rulings" now reads none at approval. Ruling 1 references were removed from Tasks 0, 3, and 9. |
| F2 | Major | Task 3 acceptance; Ruling 1 | "No setting in Geoff's own client profiles changed" has no proof and cannot fail. Isolation was scoped to Rufin's configuration home only, but Rufin keeps credentials through `oo7`, so a config-only isolation can still write Geoff's credential store. Feishin may hold his profile too. | Folded. Isolation now covers the credential store and applies to both desktop clients (settled item 9). Task 3's acceptance requires a before-and-after hash of each client's `~/.var/app/<id>/` configuration and data files, caches excluded, with every changed file listed by cause. It also records the client version walked. |
| F3 | Major | Task 0 outcome and acceptance; Tasks 1, 3, 4 | One shared probe serves four parallel tasks. Navidrome's base path is one setting per instance, and its transcode cache is shared, so Task 4's filling-versus-complete observation (planning-miss check 4) cannot be proven while other agents stream. The acceptance "answered `ping` at their own root and Navidrome at `/dubplate`" is ambiguous about how many instances exist. A short synthetic fixture can make the filling state vacuous. | Folded. Task 0 proves a parameterized recipe on three named instances (gonic root, Navidrome root, Navidrome `/dubplate`). Each observing task starts and stops its own instances. The fixture carries one track long enough to observe a filling entry. The contested-surfaces row was updated to match. |
| F4 | Major | Task 9 acceptance | The `grep "owed to the study rung"` check can pass with owed items unanswered. The phrase wraps at `:418-419` (transcoding) and `:654-655` (CORS), so grep misses both. The owed sites at `:375`, `:671-672`, `:926`, `:934`, and `:954-955` never carry the phrase. `:118` is a revision 2.5 changelog line. | Folded. The acceptance now names each spec site by line and states why a phrase search is not the check. It keeps `grep -c "from study rung"` returning 0 and adds Melodee's quality cell (`:870`). |
| F5 | Minor | Header, "Review" and "Executor" | `diff-reviewer` is set to Sonnet. That departs from the seat table (`claude-opus-5-5`, `medium`) with no reason given, on the read that gates clean-room breaches. Several dispatches name no effort, which the workstation rule requires. | Folded. Reviewers, the fact read, and `prose-voice-reviewer` run at `claude-opus-5-5`, effort `medium`. The evidence tasks run at Sonnet `high` and the landing agents at Sonnet `medium`. The cost is about seven Opus reviewer reads instead of Sonnet ones, absorbed by F6 and the ceiling. |
| F6 | Minor (over-ceremony, low cost) | Task 9, "Each landed file gets its `diff-reviewer` read" | Per-file reviews of Task 7's two files repeat the Task 8 fact read over the same files, and the STATUS and HISTORY edits get separate dispatches. | Folded. B-land takes one `diff-reviewer` over its whole diff, which also covers the post-fact-read fold. |
| F7 | Minor | Header, "Execution mode" | The departure from `pass-execute` was left for the lens to grade. Graded: it is justified. The runner's chain commits each task under the full gate, so six heavy gates (Chromium, serialized lock) would land six Markdown files. Its `parallel` mode would put six committing executors in lane 1's tree, which the build rule forbids. A-land keeps the review half of the chain. | Folded. The plan now states both reasons and drops the invitation to grade. No restructure. |
| F8 | Minor | Settled item 5; Task 0 | "Whichever lands first wins" names no channel. Task 0's pin file reaches master only at A-land, after Tasks 1 to 6, so slice 2 cannot see it in time. Task 0 is not told where to look for slice 2's pins. | Folded into Task 0's constraints. Task 0 reads slice 2's plan and worktree branch before pinning. The conductor hands Task 0's pin list to the slice 2 lane when Task 0 reports. |
| F9 | Minor | Task 9, STATUS bullet | "The 'Slice 1' and 'Owed' carries are removed" would delete the Owed line's unrelated items (`docs/STATUS.md:46-48`: the site spec's `init`/`upgrade` naming, the hook 18 consumer, the wire-types reading). | Folded. Only the slice 1 clauses are removed. |
| F10 | Minor | Task 2, AAC bullet | Cites `:945-946`. `:945` is the CORS item. | Folded to `:946`. |

Refused:

- **Lowering the 6.0M ceiling.** A ceiling caps spend; it does not forecast it. The re-plan document pass spent about
  1.6M of 3.5M, but this lane reads source across seven repositories at pinned commits, runs containers, installs an
  emulator, and walks three clients live. Rung 5d-ii overran its docs-heavy ceiling. F5's Opus reviewers add spend,
  and F6 removes a little. The evidence does not support a different number, so the ceiling stays.
- **Splitting segment A to the three-to-four default.** The plan already records the disjoint-Files override. Every
  evidence task writes one new file outside the tree, and A-land is the one green commit.
- **Moving Task 6 to Haiku.** Its license lookups are mechanical, but it also flags rows whose verdict the new
  evidence would change, which is judgment. The saving is small.

Clean-room rules: sound and checkable as written. Each rule names a reviewer-checkable property: no fenced copyleft
text, a `repo@commit:path:line` citation, a channel per claim, and an unconfirmed marker. The A-land reviewer and the
Task 8 fact read both check them. The license premises were verified above. No finding.

Contract: every task names the later slice or spec item it settles, and every acceptance criterion can fail once
F2, F3, and F4 are folded.

## Owner rulings

None remain at approval. Ruling 1 had one clearly correct answer, already ratified in the re-plan (`:140-141`), and
is settled item 9. The three forks the evidence will raise (the `transcoding` extension, folder browsing at launch,
and the module license and NOTICE) stay at the close, as the plan says.

## Measures

- Plan lines: 484 before, 518 after.
- Ceiling: 6.0M, flag 4.8M, unchanged.
- Tasks: ten, unchanged. Segments: two, unchanged.
- Rulings at approval: one before, none after.
- Verification read: not needed. No finding was a blocker.
