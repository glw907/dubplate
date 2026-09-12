# dubplate overnight window 2026-09-12: where tokens and clock went

Read-only measurement. Sources: the five `pass-execute` journals and agent
transcripts under
`.../7968f95f-d580-4a67-909a-eac9091b4953/subagents/workflows/`, the
conductor-dispatched subagents under `.../subagents/`, the conductor's own
session transcript, `scratchpad/window-ledger.md`, and `docs/HISTORY.md`'s
two newest entries.

## Unit and one correction that matters for comparing with cairn

**Claude Code transcripts repeat the same `usage` object once per content
block of an assistant message.** A naive sum over `usage` objects
double-to-quintuple-counts every multi-block turn. Every number below
deduplicates on `message.id` first, then applies the calibrated unit
`(output_tokens + cache_creation_input_tokens) / 2`.

Scale check against the conductor's own ledger: deduplicated
`output + cache_creation` (that is, 2x the unit below) reproduces the
ledger's per-segment figures within about 15 percent (segment 2A: measured
0.77M, ledger 0.79M; segment 1A: 0.85M, ledger 0.90M; segment 2B: 0.85M,
ledger 0.74M). **If cairn's measurement did not dedupe, its numbers are
inflated 2-3x and the two are not comparable.**

## 1. Per-task table

`impl` / `rev` / `fix` are unit-tokens. `fix` is implementer plus reviewer
across all fix rounds. `extra` is out-of-chain work the conductor dispatched
for that task. `wall` is impl start to final review end. `gate` is measured
wall inside `scripts/check.sh` invocations; `runs` is how many times the full
gate ran.

### Pass 8c (lane 1, main tree) - rungs 8c absorb-and-repoint

| Task | model | impl | rev | fix | extra | runs | gate | wall | gate% | verdicts | kind |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 0 staleness | opus | 38,236 | - | - | - | 0 | 0s | 4.2m | 0% | n/a | read-only |
| 1 | sonnet | 74,302 | 31,083 | 0 | - | 1 | 122s | 14.9m | 14% | accept | code/test |
| 2 | sonnet | 108,021 | 42,640 | 53,948 | 27,524 | 5 | 533s | 43.5m | 20% | fix > escalate > conductor ruling > out-of-chain fix (no review) | code/test |
| 3 | sonnet | 51,896 | 26,254 | 38,606 | - | 2 | 245s | 22.0m | 19% | fix > fix > conductor accept with note | code/test |
| 4 CLOSE | sonnet | 86,260 | 38,808 | 66,690 | - | 3 | 264s | 29.9m | 15% | fix > fix > conductor applied six pointer fixes by hand | **docs-only** |

Pass 8c chain total: **591,268**. Plus simplifier 52,582, ledger fold 58,884.

### Pass 5b-i (lane 2, worktree `lane-2-5b`) - the auth core

| Task | model | impl | rev | fix | runs | gate | wall | gate% | verdicts | kind |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 0 staleness | opus | 41,976 | - | - | 3 | 205s | 10.6m | 32% | n/a | read-only |
| 1 | sonnet | 41,195 | 20,700 | 0 | 2 | 182s | 12.3m | 25% | accept | code/test |
| 2 | opus | 86,970 | 36,402 | 0 | 1 | 111s | 26.1m | 7% | accept | code/test |
| 3 | sonnet | 101,522 | 38,612 | 58,319 | 3 | 267s | 30.2m | 15% | fix > accept | code/test |
| 4 | sonnet | 65,079 | 27,070 | 0 | 1 | 76s | 12.1m | 10% | accept | code/test |
| 5 | opus | 126,222 | 29,946 | 0 | 3 | 295s | 29.1m | 17% | accept | code/test |
| 6 | opus | 58,480 | 30,407 | 0 | 2 | 160s | 14.4m | 19% | accept | code/test |
| 7 | sonnet | 67,513 | 20,680 | 0 | 1 | 68s | 12.2m | 9% | accept | code/test |
| 8 | opus | 60,736 | 34,647 | 0 | 2 | 166s | 15.4m | 18% | accept | code/test |
| 9 | sonnet | 50,780 | 28,278 | 47,684 | 2 | 223s | 16.6m | 22% | fix > accept | code/test |
| 10 CLOSE | sonnet | 68,154 | 42,836 | 0 | 2 | ~146s* | 21.1m | ~12% | **escalate** > conductor folded, no re-dispatch | **docs-only** |

\* Task 10 redirected the gate to a file, so the tool-call delta reads 2s; the
transcript's own `time` output records `real 1m13.5s` for one of the two runs,
so about 146s actual.

Pass 5b-i chain total: **1,050,485**. Plus simplifiers 159,257 (three rounds),
merged-range diff-review 41,936, ledger fold 79,684.

**Paint:** no task in either pass moved rendered paint. `web/` does not rejoin
the gate until rung 7, and neither pass touched it. The workstation ruling's
"slow suite only where it can catch something" therefore has no browser suite
to prune here - but it has an exact analogue, see section 3.

## 2. Where the window's tokens went

Window = 09:26-13:20 UTC (01:26-05:20 AKDT), the two-lane execution through
the 5b-i merge ritual. Excludes the pre-window plan prep (08:20-09:25 UTC) and
the 5b-ii prep phase (17:30-18:35 UTC), both broken out below.

| Bucket | unit-tokens | share |
| --- | --- | --- |
| In-chain implementers (12 code tasks) | 892,715 | 36.9% |
| In-chain reviewers (first round, 12 code tasks) | 366,721 | 15.2% |
| Fix rounds (impl + reviewer, 4 tasks) | 198,557 | 8.2% |
| Out-of-chain fix (8c Task 2 R6) | 27,524 | 1.1% |
| Task 0 staleness reads (2, opus) | 80,211 | 3.3% |
| **Close tasks 8c-4 and 5b-i-10, in-chain** | **302,749** | **12.5%** |
| Ledger folds (2, conductor-dispatched opus) | 138,568 | 5.7% |
| Simplifier rounds (4) | 211,840 | 8.8% |
| Merged-range diff-review | 41,936 | 1.7% |
| Conductor's own turns (105 turns) | 154,894 | 6.4% |
| **Window total** | **2,415,715** | 100% |

Roll-ups:
- Implementer work of any kind (including fix-round implementers and the out-of-chain fix): 1,275,906 (52.8%)
- Review of any kind (in-chain first round, fix-round reviews, merged-range): 634,507 (26.3%)
- **All close-and-ledger work (close tasks + their fix rounds + folds): 441,317 (18.3%)**
- Conductor overhead proper (own turns): 6.4%

Out of the window, for context:
- **Pre-window plan and standard prep** (08:20-09:25 UTC, 22 subagents plus
  conductor): about 2.0M. The single agent "Fold seven reviews into standard
  rev 2" alone cost **564,846 in 30 turns**.
- **5b-ii plan prep** (`wf_8fcd1614-ed3`, 7 agents, 17:36-18:35 UTC):
  **830,069**. Breakdown: draft 230,218; three parallel plan reviewers
  (mechanics 89,542, contract 101,412, security 95,422) = 286,376; fold
  152,846; 5b-i security read 82,851; spec errata 77,778.
  **That is one plan revision costing more than pass 8c's entire code
  execution (Tasks 1-3, implementers plus reviewers plus fixes: 454,274).**
- **Session-resume re-caching**: one conductor turn block at 17:33 UTC, after
  the roughly four-hour idle gap between the 5b-i close and the prep phase,
  carries `cache_creation = 574,539` in a single request - **303,912 unit in
  4 turns**, against 154,894 for the conductor's entire 105-turn execution
  window. The prompt cache expired over the gap and a 605K-token conversation
  was re-primed from scratch.

## 3. Clock: how much of each task was the gate

| | 8c | 5b-i |
| --- | --- | --- |
| Full `check.sh` invocations in-chain | 13 | 22 |
| Measured gate wall | 19.4m | 29.3m (+7.1m merge ritual) |
| Chain wall (first impl start to last review end) | 123.8m | 201.6m |
| Gate share of chain wall | **15.7%** | **14.5%** (18.0% with the merge ritual) |

Per-task gate share ranges 7% to 32%; the median code task pays about 18%.
Thirty-five full gate runs across fourteen tasks, 2.5 per task.

**What the gate actually spends its time on**, from the per-leg lines the
implementers reported:

- `gitleaks` over the tracked worktree: **56.7s to 2m13s**, scanning
  ~119.6 MB. It is the single dominant leg in most runs.
- `internal/harness/conform`: **125.6s** whenever the Go test cache misses it.
- Everything else (vet, lint, the touched package's tests, shellcheck, ruff,
  vale, compose config, systemd-analyze, state-dir agreement): seconds.

Both dominant legs are **blast-radius-insensitive**. The corpus under
`corpus/cases/` does not change, and `internal/harness/conform` is untouched
by anything either pass wrote - yet a docs-only close task pays both. 8c
Task 4, which committed two Markdown drafts and a spec erratum, ran the full
gate three times for **264s**. 5b-i Task 10, also Markdown-only, ran it twice
for about 146s. That is about **7 minutes of gate on two commits that touched
no code**, and 5b-i Task 0 - a read-only staleness probe - spent 205s more
running the gate plus `go test ./...` purely to record a baseline.

## 4. Fix rounds and escalations, classified

Six fix/escalate events across fourteen tasks. Only one was a genuine code
defect.

| Event | classification | cost | what it was |
| --- | --- | --- | --- |
| 5b-i Task 9, `fix` | **real defect** | 47,684 | `CheckSession`/`CheckLink` keyed the epoch lookup on the raw address, not the normalized one; a session differing only in case escaped revocation. Caught by the in-chain reviewer, fixed in one round, accepted. This is the gate earning its keep. |
| 5b-i Task 3, `fix` | **test evidence missing** (2 of 4 findings) + **comment-only** (2 of 4) | 58,319 | Acceptance criteria were asserted only against scratch registries, so a wrongly declared real key name could not fail the test; moved onto `productionRegistry()`. The other two findings were stale package-doc paragraphs. |
| 8c Task 2, `fix` then `escalate` | **comment-only** (2 of 3) + **plan-infeasible test outcome** (1 of 3) | 53,948 in-chain + 27,524 out-of-chain | Two findings were doc-comment restorations (per-row mediafile citations, naming ruling 3). The third asked for an invalid-UTF-8 projection row the implementer judged unbuildable through `tagio.Write`; it escalated, the conductor ruled a byte-splice route, and a separate implementer landed it **with no diff-reviewer at all** (recorded deviation). |
| 8c Task 3, `fix` then `fix` | **citation pointers**, entirely | 38,606 | Both rounds concerned the enumeration table **inside the task's return report** - no repo file changed in the fix round at all. The second verdict was one wrong subtest name (`TestParseRefusals` where the grading test lives in `picture_test.go`). Conductor accepted with a revision note. |
| 8c Task 4 CLOSE, `fix` then `fix` | **draft-content** + **citation pointers** | 66,690 | Round 1: rulings named by the wrong numbering, four wrong line pointers, ROADMAP text in the wrong draft. Round 2: six **off-by-one** line pointers (`spec:1035` where the paragraph starts at 1036, and four more of the same shape). Conductor applied all six by hand as a docs-only commit. |
| 5b-i Task 10 CLOSE, `escalate` | **draft-content**, entirely | 42,836 (review only) | Seven blockers: missing measured wall times, "(6 named)" enumerating five, two doc-comment claims that did not match the code, a phantom third simplifier round, a 65-line STATUS against a 60-line cap, a dropped "awaiting Geoff's read" block. Conductor corrected all seven while folding; no re-dispatch. |

Totals by cause:
- Real code defect: **47,684** (18% of fix-round spend)
- Vacuous/missing test evidence: about 29,000 (half of 5b-i Task 3)
- Comment-only, citation-pointer, and draft-content: **about 189,000**, that
  is roughly **71% of all fix-round tokens** plus the 42,836 escalate review
  on Task 10, none of which changed a line of shipped Go.

Off-by-one line pointers into a 5,000-line `docs/HISTORY.md` and a 1,300-line
spec are their own recurring sub-class: they caused a full reviewer round on
8c Task 3 and the second round on 8c Task 4.

## 5. What the close tasks cost against the code tasks

| | count | chain tokens | per task |
| --- | --- | --- | --- |
| Code tasks (8c 1-3, 5b-i 1-9) | 12 | 1,485,516 | **123,793** |
| Close tasks (8c Task 4, 5b-i Task 10) | 2 | 302,749 | **151,374** |

**A docs-only close task cost 22 percent more than the average code task**,
and that is before the ledger fold each one still needed: 58,884 for 8c and
79,684 for 5b-i, dispatched to opus after the close task escalated or
returned drafts the conductor had to correct anyway.

Escalation cost specifically:
- 8c Task 4's two fix rounds: **66,690**, and the second verdict was still six
  off-by-one line numbers that the conductor then fixed by hand. The whole
  round-trip bought a docs-only commit.
- 5b-i Task 10's escalate: **42,836** of reviewer time to produce seven
  findings the conductor then corrected itself while folding - the re-dispatch
  the chain would normally do was explicitly skipped as not worth it.

Full close-out cost across both passes: **441,317 (18.3% of the window)** for
work that produced Markdown. Neither close's drafts were accepted as written.

## 6. Overhead per launch and per conductor turn

**Per `pass-execute` launch (segmentation).** Five runs in the window
(1A, 1B, 2A, 2B, 2C), four segment boundaries. A boundary costs:

| item | typical |
| --- | --- |
| `code-simplifier` round over the segment | 38,364 - 63,160 (mean **52,960**) |
| Conductor turns to close, checkpoint STATUS, and relaunch | 10,000 - 31,000 |
| **per boundary** | **about 65,000-85,000** |

Four boundaries = **about 272,000, roughly 11% of the window**. Note the
simplifier rounds were already batched per segment rather than per task (the
correct economy); the marginal segmentation cost is the two extra rounds that
segmenting into five runs instead of two forced, about 100,000, plus the
conductor's relaunch turns.

There is **no** within-segment cache sharing to lose: `pass-execute` spawns a
fresh agent per task, so every implementer re-reads the plan regardless of
which run it sits in. Splitting a pass across runs costs the ritual, not the
reads.

**Per conductor turn.** 105 turns across the 4-hour execution window for
**154,894** unit, or **1,475 per turn**. The thin-conductor discipline held:
the conductor read no diffs and no gate logs, and its share of the window was
6.4 percent.

The contrast is the resume: **4 turns for 303,912** (76,000 per turn) when the
session was re-primed after the four-hour idle gap before the 5b-ii prep
phase. One idle gap cost twice what four hours of active conducting cost.

## 7. Model mix

| model | agents | tokens | share of agent spend |
| --- | --- | --- | --- |
| sonnet (implementers, incl. fix rounds and the out-of-chain fix) | 16 dispatches | 863,288 | 35.7% |
| opus (implementers, upshifted) | 4 | 332,408 | 13.8% |
| `claude-opus-5` (in-chain diff-reviewer) | 16 | 592,571 | 24.5% |
| opus (Task 0, simplifier, ledger folds, merged-range review) | 9 | 472,555 | 19.6% |

Reviewers are cheap per call (mean 37,036, range 20,680-42,836) and the
reviewer fleet as a whole is 26.3% of the window. The escalating reviews are
the expensive ones and they are all on prose.
