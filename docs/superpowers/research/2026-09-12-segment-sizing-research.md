# Sizing segments and checkpoints in a long-running coding-agent run

Research note, 2026-09-12. Every claim carries the URL it came from and that
source's date. Sources marked **(fetched)** were read in full; sources marked
**(search summary)** were read only through a search-result summary and should
be treated as weaker evidence.

## The short answer

There is no public consensus on a number. No vendor, no framework, and no paper
publishes "N tasks per segment" or "compact every K tokens" for a coding-agent
run. What exists is a stack of converging *principles* plus a handful of hard
numbers that bound the answer from different directions: per-step reliability
decay from the agent-reliability literature, context-fill degradation curves,
step and cost budgets from the agent frameworks, and a fifty-year-old
checkpoint-interval formula from HPC that turns out to apply almost unchanged
once the terms are named correctly.

Run that formula on this workstation's measured numbers and it returns three to
four tasks. The current practice is already at the optimum, and the optimum is
flat enough that the exact figure barely matters.

---

## 1. What the vendors say

### Anthropic

**Effective context engineering for AI agents** —
https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
**(fetched)**

- Names three techniques for long-horizon work: compaction, structured
  note-taking to files outside the context window, and multi-agent
  architectures. Compaction is the first lever; it "distills the contents of a
  context window in a high-fidelity manner".
- A subagent "returns only a condensed, distilled summary of its work (often
  1,000-2,000 tokens)" while exploring far more than that internally.
- Explicitly gives **no percentage threshold** for when to compact, and **no
  duration threshold** for when to segment. The tuning advice is to maximize
  recall in the compaction prompt first, then iterate on precision.

**Effective harnesses for long-running agents** (published 2025-11-26) —
https://anthropic.com/engineering/effective-harnesses-for-long-running-agents
**(fetched)**

This is the closest thing to a vendor answer on segment size, and its answer is
**one unit of verifiable work per fresh context**:

- Work is partitioned as a structured JSON feature list (200+ discrete features
  in their claude.ai-clone example), each marked `"passes": false` until done.
- Sessions process **one feature per session**. "Every subsequent session asks
  the model to make incremental progress, then leave structured updates."
- Each session starts by reading git log and a `claude-progress.txt` file,
  picking the highest-priority incomplete feature, booting the app via
  `init.sh`, and running an end-to-end check **before** starting new work.
- The handoff artifacts are the git history, the progress file, the JSON feature
  registry, and a working tree in mergeable condition. Not a conversation
  summary.
- The agent is forbidden from editing the feature list except to flip pass/fail,
  and forbidden from deleting tests.
- No token thresholds, no checkpoint-sizing numbers.

A secondary read of the same work (Addy Osmani, **Long-running Agents**,
2026-04-28 — https://addyosmani.com/blog/long-running-agents/ **(fetched)**)
reports the finding behind it: "summarization-as-compaction wasn't enough for
very long jobs; they had to do full context resets where the harness tears the
session down and rebuilds it from a structured handoff file." Osmani's own
checkpoint rule is deliberately unquantified: "write intermediate state to disk,
checkpoint every N units of work, recover from failures" — and explicitly "not
every step, not only the end." He also lists the long-run failure modes that
motivate an external checkpoint: agents declare a task complete when it isn't,
skew positive when grading their own work, delete failing tests, and drift in
alignment through repeated summarization.

**Best practices for Claude Code** —
https://code.claude.com/docs/en/best-practices **(fetched)**

The operational rules, all qualitative but all directly on point:

- "Most best practices are based on one constraint: Claude's context window
  fills up fast, and performance degrades as it fills."
- `/clear` between unrelated tasks. "Long sessions with irrelevant context can
  reduce performance."
- **A hard numeric rule exists for one case**: "If you've corrected Claude more
  than twice on the same issue in one session, the context is cluttered with
  failed approaches. Run `/clear` and start fresh." Two failed corrections is
  the published reset trigger.
- Commit is the natural checkpoint; the four-phase loop is Explore → Plan →
  Implement → Commit.
- On batch runs: "Refine your prompt based on what goes wrong with the first 2-3
  files, then run on the full set." A pilot of 2-3 before committing to the
  batch.
- `/batch` splits a change across **5 to 30 subagents**, each in its own
  worktree, each opening a PR.
- "The longer Claude works unattended, the more an independent check matters
  before you count the work as done" — and the reviewer must be a fresh context
  that "sees only the diff and the criteria you give it, not the reasoning that
  produced the change."
- A warning against over-checking: "A reviewer prompted to find gaps will
  usually report some, even when the work is sound."

**Checkpointing** — https://code.claude.com/docs/en/checkpointing **(fetched)**

- A checkpoint is created for **every prompt that starts a turn**; snapshots are
  kept for the **100 most recent** checkpoints in a session.
- Bash-driven file changes are **not** tracked; subagent edits are usually **not**
  restored by rewind. "Not a replacement for version control."
- Relevance here: the built-in checkpoint mechanism does not cover a chain that
  edits through subagents. Git commits, not `/rewind`, are the resume point for
  this workstation's shape.

**How we built our multi-agent research system** —
https://www.anthropic.com/engineering/multi-agent-research-system **(fetched)**

- Agents use ~4x the tokens of chat; multi-agent ~15x. Token usage alone
  explains **80% of performance variance** on their browsing eval.
- Decomposition granularity by query type: simple fact-finding 1 agent and 3-10
  tool calls; comparisons 2-4 subagents at 10-15 calls each; complex research
  10+ subagents. Lead agents typically spawn **3-5 subagents in parallel**.
- The lead agent "saves research plans to external memory before reaching the
  200,000-token truncation threshold," then spawns fresh subagents with clean
  contexts, "maintaining continuity through careful handoffs."
- Important caveat for a coding conductor: "most coding tasks involve fewer
  truly parallelizable tasks than research," so the multi-agent numbers do not
  transfer directly.

**When to use multi-agent systems (and when not to)**, 2026-01-23 —
https://claude.com/blog/building-multi-agent-systems-when-and-how-to-use-them
**(fetched)**

- Split when a subtask generates **>1000 tokens** of information irrelevant to
  subsequent work. That is the published decomposition threshold.
- Tool overload threshold: **15-20+ tools** suggests decomposition.
- Multi-agent costs **3-10x** the tokens of single-agent for equivalent work.
- **The strongest published argument against many boundaries**: avoid sequential
  phase splitting (planning → implementation → testing), because it creates
  "telephone game" coordination losses. Decompose by context isolation, not by
  phase.
- The verification subagent is named as the robust checkpoint mechanism: a
  dedicated agent validates output without needing full implementation context.

### OpenAI

**A practical guide to building agents** —
https://openai.com/business/guides-and-resources/a-practical-guide-to-building-ai-agents/
**(search summary; the PDF at cdn.openai.com would not render as text)**

- Two named triggers for handing control back to a human: **exceeding failure
  thresholds** (set limits on retries or actions, escalate when exceeded) and
  **high-risk actions** (sensitive, irreversible, high-stakes).
- Start with a single agent; add orchestration only when needed.
- Guardrails are layered, not a single gate.
- No task-count or token-count per segment.

**Guardrails and human review** —
https://developers.openai.com/api/docs/guides/agents/guardrails-approvals
**(search summary)**. Same structure: approvals attach to specific tool calls,
not to a clock or a step count.

### Google

**Build long-running AI agents that pause, resume, and never lose context with
ADK**, 2026-05-12 —
https://developers.googleblog.com/build-long-running-ai-agents-that-pause-resume-and-never-lose-context-with-adk/
**(fetched)**

- Checkpoint granularity is **the tool-invocation boundary**: "Every tool call
  creates an automatic checkpoint. If the container crashes immediately after
  `send_welcome_packet` runs, the state has already been written." Tool
  execution is the atomic unit of durability, so state persists *before* an
  irreversible effect is observable.
- Workflow-level segmentation is an explicit small state machine (their example:
  six named states), not a token budget.
- Pauses are placed at human-approval gates, and the state transition is applied
  atomically *before* the next inference call.
- No performance or overhead numbers.

Framework-level guidance in the same family (**search summary**, via
https://addyo.substack.com/p/long-running-agents and
https://fast.io/resources/ai-agent-checkpointing-resume/) names the two poles
directly: **step-based checkpointing** (after every thought or tool call, as
LangGraph does — minimal loss, maximal I/O) versus **milestone checkpointing**
(after major phases — cheaper, but risks repeating work). Choosing between them
is described as a granularity decision, with no formula offered.

---

## 2. Step and cost budgets in the agent frameworks

**(search summary)**, from
https://github.com/OpenHands/software-agent-sdk/issues/2406,
https://github.com/OpenHands/OpenHands/issues/6857, and
https://dev.to/truongpx396/swe-agent-deep-dive-build-your-own-guide-ade:

- **mini-SWE-agent**: default max step limit **250**; default
  `per_instance_cost_limit` **$3**.
- **SWE-agent v1.0** deliberately chose a **cost budget over a step budget**,
  because step counts vary too much across model families. This is the single
  most transferable design decision in this section: budget the resource that is
  comparable across models, not the one that isn't.
- **OpenHands**: default `MAX_ITERATIONS` ~**100**, enforced server-side as a
  hard ceiling on LLM calls per conversation. An open proposal sets subagent
  limits at **20-50 steps** (currently unlimited).
- OpenHands' RetryAgent re-allocates the unused budget from a cheap fast failure
  into the next attempt — budget is per-run, not per-attempt.

**Devin / Cognition (search summary)**, via
https://cognition.com/blog/how-cognition-uses-devin-to-build-devin **(fetched,
but the fetched copy carried no numbers)** and secondary reports at
https://vibeanswers.com/devin/session-stuck-in-a-loop/ and
https://frontierai.substack.com/p/one-month-of-using-devin:

- Rule of thumb: if a task would take a human **three hours or less**, Devin can
  probably do it.
- "Large-scale challenges should be broken into smaller, isolated tasks across
  separate sessions" (this line **is** in the fetched Cognition post).
- In-product warnings reportedly appear around **2.5 hours or ~10 ACUs**, with
  quality degradation in long sessions attributed to context limits. Treat the
  2.5h/10-ACU figure as secondary reporting, not a Cognition statement.

---

## 3. The decay evidence: the strongest argument for smaller segments

**How Fast Do Agents Rot? An Empirical Study of Long-Horizon Degradation in LLM
Agents for Production Decision-Making**, arXiv:2609.01660 —
https://arxiv.org/abs/2609.01660 **(fetched, abstract only)**

- 10,664 trajectories, models from 1.2B to 671B parameters.
- Task success follows a **geometric** pattern governed by a **single per-step
  reliability parameter**. Models fall "from near-perfect success to near zero
  within sixteen steps."
- **Degradation is driven by step count, not context length.** Constraining the
  context window *steepens* decay rather than easing it (logit slope -0.69 vs
  -0.44).
- Benchmark success drops from 0.42 at GAIA-length horizons to 0.24 at
  hundred-step production horizons.
- Recommends "horizon-aware evaluation and reliability budgeting in place of
  aggregate pass-rate metrics."

This is the load-bearing finding for segment sizing, and it cuts against the
naive reading of "context rot." If reliability is geometric in steps with
per-step rate p, then P(segment of S units all clean) = p^S. Segment size trades
linearly against boundary cost but **exponentially** against the probability of
a clean segment. That asymmetry is why the answer is always a small number.

**Is there a half-life for the success rates of AI agents?** (Toby Ord),
arXiv:2505.05115 — https://arxiv.org/abs/2505.05115 **(fetched, abstract)**

- Same shape from the other direction: a constant failure rate per minute of
  human-equivalent task duration produces exponentially declining success with
  task length. Each agent has a **half-life** — the task length at which success
  hits 50%.
- The authors caveat that the pattern may be specific to their
  research-engineering task suite.

**METR time horizons** — https://metr.org/blog/2026-1-29-time-horizon-1-1/
**(fetched)** and https://metr.org/time-horizons/ **(fetched)**

- Frontier 50% time horizon doubles roughly every **131 days** (TH1.1, models
  since 2023), revised down from 165 days.
- 50% horizons at TH1.1: Claude Opus 4.5 **320 min** [170, 729]; GPT-5 214 min;
  o3 121 min.
- The confidence interval is still 2.3x wide at the frontier.
- The original METR post (2025-03-19,
  https://metr.org/blog/2025-03-19-measuring-ai-ability-to-complete-long-tasks/
  **(fetched)**) gives the shape rather than the 50/80 ratio: near-100% success
  on tasks under ~4 minutes of human time, **under 10%** on tasks over ~4 hours.

The practical reading: the *50%* horizon is where a frontier agent coin-flips.
Anything you want to succeed unattended should sit well inside it, and the 80%
horizon (not published as a clean ratio in the sources read) is materially
shorter. A segment of work that a human would need several hours to do is at or
past the coin-flip line for the current frontier.

**Classifier Context Rot**, arXiv:2605.12366, 2026-05-12 —
https://arxiv.org/html/2605.12366v1 **(fetched)**

- Opus 4.6 with thinking: monitoring recall **98.6% → 88%** when 800K tokens of
  benign activity precede the dangerous action.
- Needle injection of obvious attacks: recall **99.7% at 100K → 69% at 800K**.
- All three frontier models tested (Opus 4.6, GPT 5.4 Thinking, Gemini 3.1 Pro)
  missed attacks at long context.
- Note this measures a *monitor*, not a coder, but it is the cleanest published
  curve of a review-shaped judgment degrading with context length — which is
  exactly the job a conductor does at a boundary.

**Context-fill thresholds (practitioner, weaker)** —
https://agentpatterns.ai/context-engineering/context-window-dumb-zone/ (reviewed
2026-07-22) **(fetched)**

- Explicitly **rejects** a universal percentage rule; recommends absolute token
  budgets per task type. Reasoning tasks: under 32K where feasible, since
  "effective window can be 10-20% of the advertised limit."
- Cites BABILong (models effectively use 10-20% of context for reasoning),
  NoLiMa (11 of 13 models below 50% accuracy at 32K without lexical cues),
  LongCodeBench (Claude 3.5 Sonnet bug-fixing 29% at 32K → 3% at 256K).
- Notes Claude Code auto-compacts near the top of the window and that teams
  commonly lower the trigger.
- Warns the other way: "compressing a long context into a shorter summary
  discards detail" and can lose intermediate state in multi-step tasks.

Other practitioner figures circulating (**search summary**, weak): quality
degrades at 70-80% fill (https://www.mindstudio.ai/blog/context-rot-ai-agents-auto-compact-fix),
"danger zone" above 60%, effective capacity 60-70% of advertised. These are
mutually inconsistent, which is itself the finding: **there is no agreed
context-fill number.** Treat fill percentage as a symptom to watch, not a
threshold to schedule against.

---

## 4. Batch size on the human-review side

https://codex.danielvaughan.com/2026/05/24/human-review-bottleneck-code-review-strategies-agent-output/
(updated 2026-09-10) **(fetched)**

- **250 changed lines maximum per PR**; larger PRs get slower, lower-quality
  review.
- Break agent output into small, dependent (stacked) PRs, not monolithic
  changesets.
- Review SLA targets: 4-hour first response, 24-hour resolution; a queue over
  two days is a structural bottleneck.
- A cited 25-40% band for agent-generated code share per team, above which
  review burden outweighs productivity gains.

**(search summary)**, from
https://www.futureproofing.dev/resources/ai-native-team/agentic-coding-workflow-2026
and related: teams on high-adoption multi-agent workflows report 98% more PRs
merged but 91% longer review times; AI-authored PRs wait 4.6x longer for a
reviewer to pick them up though review itself completes 2x faster. The
bottleneck migrates to review, which argues for boundaries that produce
review-ready units rather than boundaries that only save state.

---

## 5. The checkpoint-interval literature, and whether it transfers

**Checkpointing à la Young/Daly: An Overview** (Benoit et al.) —
https://dl.acm.org/doi/10.1145/3549206.3549328 and
https://icl.utk.edu/files/publications/2022/icl-utk-1569-2022.pdf **(search
summary; the PDF would not render as text)**. Survey follow-up: "Should we
always checkpoint à la Young/Daly?" —
https://www.sciencedirect.com/science/article/abs/pii/S0167739X24003777 **(search
summary)**.

- Young: `T_opt ≈ sqrt(2 · δ · MTTF)`, where δ is the cost of writing a
  checkpoint.
- Daly's refinement: `T_opt = sqrt(2δ(M+R)) − δ`, with M the mean time between
  failures and R the restart overhead.
- Assumptions: exponential (memoryless) failures, checkpoint cost independent of
  interval, preemptible tightly-coupled work.
- Known limits: designed for tightly-coupled preemptible applications; accuracy
  for dynamic task-based runtimes is an open question; with many concurrent
  tasks, checkpointing *more often* than Young/Daly can be better
  (http://www.ittc.ku.edu/~sun/publications/topc22.pdf **(search summary)**).

**No source found applies Young/Daly to LLM agent runs.** That appears to be an
unoccupied niche. It transfers better than one might expect, though, once the
terms are renamed — see below.

---

## 6. Recurring principles (what the sources actually agree on)

Ranked by how many independent sources assert them.

1. **Checkpoint at a natural, verifiable boundary — normally a commit.** Claude
   Code best practices (commit = checkpoint); Anthropic's harness post (git
   history as the handoff artifact); Google ADK (tool-invocation boundary);
   Devin (separate sessions per isolated task).
2. **Checkpoint before an irreversible or high-risk action.** OpenAI (explicit);
   Google ADK (state written before the effect); this is the one principle
   nobody hedges.
3. **The handoff must be structured state, not a conversation summary.**
   Anthropic's harness post (JSON feature list + progress file + git log);
   Anthropic's multi-agent post (plan saved to external memory before the
   truncation threshold); the multi-agent blog's "telephone game" warning
   against phase-shaped handoffs.
4. **The durable record lives in files, not in the transcript.** All of the
   above, plus the Claude Code docs.
5. **Reset on a fixed failure threshold, not only on a clock.** Claude Code:
   after two failed corrections. OpenAI: retry/action limits. OpenHands:
   `MAX_ITERATIONS`. SWE-agent: cost limit.
6. **Verification is a separate, fresh context.** Claude Code's adversarial
   review step; the multi-agent blog's verification subagent; Anthropic's
   harness evaluator; Osmani's "models skew positive when grading their own
   work."
7. **Pilot before committing to the full batch.** Claude Code's "refine your
   prompt based on what goes wrong with the first 2-3 files."
8. **Segment by context isolation, not by phase.** The multi-agent blog, stated
   as a warning against exactly the planning/implementation/testing split.
9. **Budget the comparable resource.** SWE-agent's move from step budgets to
   cost budgets.

Principles that are asserted but *not* agreed: any specific context-fill
percentage; any task count per segment; any wall-clock session length.

---

## 7. The two strongest arguments, stated fairly

**Strongest argument for smaller segments — reliability is geometric in steps.**
arXiv:2609.01660 finds success governed by a single per-step reliability
parameter, with models falling from near-perfect to near-zero **within sixteen
steps**, and finds the driver is step count rather than context length. Under
p^S, doubling segment size squares the failure probability while only halving
the amortized boundary cost. The boundary is also the only place a fault is
*detected*: undetected, a bad task poisons every task after it in the segment,
so expected wasted work grows with the square of segment size (S/2 wasted per
fault × S/M faults per segment). And a boundary does more than save state — the
fresh-context restart resets the very quantity that is decaying, so smaller
segments raise p as well as bounding the loss.

**Strongest argument for larger segments — every boundary is a lossy handoff,
and it costs the same as the fault it prevents.** Anthropic's own multi-agent
guidance (2026-01-23) warns that sequential phase splitting produces "telephone
game" coordination losses, and tells you to split only "when context can be
truly isolated." The agentpatterns survey warns that compressing context
"discards detail" and risks losing intermediate state. Osmani names alignment
drift through repeated summarization as a long-run failure mode. On this
workstation the boundary is not free — 65-85K tokens and a conductor round each
time — and the boundary is *itself* a judgment step performed by a model whose
accuracy at that job degrades measurably with context (Opus 4.6 monitoring
recall 98.6% → 88%). A boundary that drops a decision costs exactly what the
fault it was meant to catch would have cost.

The second argument has a clean mathematical consequence: **handoff loss and
checkpoint cost are the same term.** Both are a fixed per-boundary cost. So the
whole trade-off collapses into one classical form, which is why Young/Daly
transfers.

---

## 8. Recommendation for this workstation

### The model

Let S = tasks per segment. Measure the boundary cost δ in **task-equivalents**
(boundary tokens ÷ mean tokens per task through the implementer-reviewer-gate
chain), and let M = mean number of tasks between faults that need a redirect or
a rollback. Expected overhead per useful task is

    H(S) = δ_eff / S  +  S / (2M)

where δ_eff folds together the boundary's direct cost *and* its expected
handoff-loss cost (they are both per-boundary, so they add). The first term is
the checkpoint tax; the second is expected rework, since a fault lands on
average half a segment before it is detected. Minimizing gives Young's formula
unchanged:

    S* = sqrt(2 · M · δ_eff)

Being wrong by a factor of two in either direction costs exactly **25% more
overhead** (H(2S*) = H(S*/2) = 1.25 · H(S*), derived from the expression above).
That flatness is the single most useful property here: it means the tuning
question is not worth much attention once you are in the right neighborhood, and
it matches the HPC literature's practical experience that Young/Daly is robust
even where its assumptions are shaky.

### Running the numbers

A boundary costs 65-85K tokens (call it 75K). If a task through the chain costs
roughly 150-200K tokens, δ_eff ≈ 0.4-0.5 task-equivalents before handoff loss;
call it 0.5 with loss folded in.

| Fault rate M (tasks per redirect) | δ_eff = 0.25 | δ_eff = 0.5 | δ_eff = 1.0 |
|---|---|---|---|
| 6  | 1.7 | 2.4 | 3.5 |
| 8  | 2.0 | 2.8 | 4.0 |
| 12 | 2.4 | 3.5 | 4.9 |
| 20 | 3.2 | 4.5 | 6.3 |

Across every plausible cell, S* lands between 2 and 5, and for the central
estimates (M = 8-12, δ_eff = 0.5) it lands on **3 to 4**. The current practice is
already at the optimum. This is not a coincidence — it is the same arithmetic
that puts Anthropic's harness at one feature per session (their δ is near zero
because the handoff is a JSON file and a git log, so S* collapses toward 1) and
puts Devin at one isolated task per session.

Two effects push the answer *down* from this table and none push it up:
M is not independent of S (a fresh-context restart raises per-step reliability,
so M(S) declines as S grows), and the geometric decay in arXiv:2609.01660 means
the rework term is super-linear rather than linear in S.

### The rule

> **Segment a pass at three to four tasks, and place every boundary on a commit
> that the gate has already proved green. Override the count — never the
> placement — on any of four triggers: an irreversible or production-touching
> task (boundary immediately before it); a second `fix` verdict from the
> diff-reviewer on the same task (boundary immediately, per Claude Code's
> two-correction rule); a task whose Files or acceptance criteria are disjoint
> from the next task's (boundary at the seam, per the context-isolation
> principle); or the first segment of a plan whose shape is new (cut it to two
> tasks as a pilot, per the "first 2-3 files" rule). Carry the handoff as
> structured state in the repo — the STATUS ledger, the plan's task table, the
> git log — never as conversation the next segment has to re-read.**

Evidence for each clause:

- **Three to four**: Young/Daly with this workstation's measured δ and plausible
  M (section 8 table); convergent with the frontier practice of 1 unit per
  session where the handoff is nearly free (Anthropic harness post, Devin), and
  with the 25% flatness result that says nearby values are fine.
- **On a green commit**: Claude Code best practices (commit as checkpoint);
  Anthropic harness post (git log as handoff); Google ADK (state written before
  the effect completes); and the Claude Code checkpointing docs' own limitation
  that `/rewind` does not restore subagent edits, so git is the only real resume
  point for a chain like this one.
- **Before irreversible actions**: OpenAI's high-risk-action trigger; Google
  ADK's atomic pre-write.
- **Second `fix` verdict**: Claude Code's published two-correction reset rule,
  which is the only hard reset number any vendor publishes.
- **At a disjointness seam**: the 2026-01-23 multi-agent blog's context-isolation
  rule and its >1000-token-of-irrelevant-context threshold.
- **Two-task pilot on a new plan shape**: Claude Code's batch guidance.
- **Structured handoff, not conversation**: Anthropic's harness post; the
  telephone-game warning; Osmani's alignment-drift-through-summarization failure
  mode.

Deliberately *not* in the rule: a context-fill percentage. The published numbers
disagree (50%, 60%, 70-80%, "10-20% effective"), the best source on the subject
rejects a universal percentage outright, and arXiv:2609.01660 finds step count
rather than context length is the driver. Fill is worth a glance at each
boundary as a symptom; it is not a scheduling trigger.

### What to measure to tune it

Only three numbers matter, and two of them are already sitting in the pass
ledger unrecorded.

1. **δ_eff — boundary cost in task-equivalents.** Record, per boundary: tokens
   spent on the checkpoint round, and tokens spent on the tasks in the segment
   it closed. Divide. If δ_eff drifts above ~1.0 (a boundary costing as much as
   a task), the boundary ritual has grown and should be trimmed before the
   segment is lengthened — the formula's answer to an expensive checkpoint is
   longer segments, which is the wrong fix when the expense is self-inflicted.
2. **M — tasks per fault.** Count, per pass: `fix` verdicts from the
   diff-reviewer, gate failures that survived the implementer, and rollbacks or
   re-dispatches. M = tasks completed ÷ that count. Recompute S* each pass. This
   is the number that moves; δ_eff is nearly constant.
3. **Fault-detection latency in tasks.** For each fault, how many tasks had run
   between the defect being introduced and it being caught. If the mean
   approaches S/2 the model is behaving as assumed; if it is consistently near
   S, faults are being caught only at the boundary and not by the in-chain
   reviewer, which means the boundary is doing work the chain should do — and
   the cheaper fix is a better acceptance criterion, not a shorter segment.

One diagnostic worth adding, since it is nearly free: log whether each fault was
caught by the in-chain `diff-reviewer`, by the gate, or by the conductor at the
boundary. If the conductor's share is small, the boundaries are buying resume
points and fresh context but not fault detection, and the segment can safely go
to the top of the range. If the conductor's share is large, the reliability
argument dominates and the segment should sit at the bottom.

---

## Source list

Fetched and read in full:
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents
- https://anthropic.com/engineering/effective-harnesses-for-long-running-agents (2025-11-26)
- https://www.anthropic.com/engineering/multi-agent-research-system
- https://claude.com/blog/building-multi-agent-systems-when-and-how-to-use-them (2026-01-23)
- https://code.claude.com/docs/en/best-practices
- https://code.claude.com/docs/en/checkpointing
- https://www.anthropic.com/news/enabling-claude-code-to-work-more-autonomously (2025-09-29)
- https://developers.googleblog.com/build-long-running-ai-agents-that-pause-resume-and-never-lose-context-with-adk/ (2026-05-12)
- https://addyosmani.com/blog/long-running-agents/ (2026-04-28)
- https://cognition.com/blog/how-cognition-uses-devin-to-build-devin
- https://arxiv.org/abs/2609.01660
- https://arxiv.org/abs/2505.05115
- https://arxiv.org/html/2605.12366v1 (2026-05-12)
- https://metr.org/blog/2026-1-29-time-horizon-1-1/ (2026-01-29)
- https://metr.org/time-horizons/ (updated 2026-05-08)
- https://metr.org/blog/2025-03-19-measuring-ai-ability-to-complete-long-tasks/ (2025-03-19)
- https://agentpatterns.ai/context-engineering/context-window-dumb-zone/ (reviewed 2026-07-22)
- https://codex.danielvaughan.com/2026/05/24/human-review-bottleneck-code-review-strategies-agent-output/ (updated 2026-09-10)

Read only as search-result summaries (weaker evidence):
- https://openai.com/business/guides-and-resources/a-practical-guide-to-building-ai-agents/
- https://developers.openai.com/api/docs/guides/agents/guardrails-approvals
- https://github.com/OpenHands/OpenHands/issues/6857
- https://github.com/OpenHands/software-agent-sdk/issues/2406
- https://dev.to/truongpx396/swe-agent-deep-dive-build-your-own-guide-ade
- https://dl.acm.org/doi/10.1145/3549206.3549328
- https://icl.utk.edu/files/publications/2022/icl-utk-1569-2022.pdf
- https://www.sciencedirect.com/science/article/abs/pii/S0167739X24003777
- http://www.ittc.ku.edu/~sun/publications/topc22.pdf
- https://vibeanswers.com/devin/session-stuck-in-a-loop/
- https://frontierai.substack.com/p/one-month-of-using-devin
- https://www.futureproofing.dev/resources/ai-native-team/agentic-coding-workflow-2026
- https://www.mindstudio.ai/blog/context-rot-ai-agents-auto-compact-fix
- https://fast.io/resources/ai-agent-checkpointing-resume/
