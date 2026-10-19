# Soundness of the workstation's agentic development pattern, against public evidence

Research brief, 2026-09-12. Every external claim carries a URL that was opened and the source's date.
Claims marked SECONDARY were located through a search summary and not verified against the primary text.
Claims marked LOCAL come from this workstation's own record (`~/Projects/dubplate/docs/HISTORY.md`,
about 6,800 lines covering roughly eighteen closed passes) and are observational, not controlled.

## 1. What the local record already shows

The dubplate HISTORY ledger is the single most relevant dataset for this question, because it records
the pattern's own outcomes over eighteen passes at a consistent bar. Five facts stand out.

**LOCAL-1: the deterministic gate almost never catches anything the reviewer did not.** Across passes
2c, 2d, 3a, 3c and 4a the close entries say `scripts/check.sh` "stayed green at every task end and found
nothing novel" (`docs/HISTORY.md:4447`, `:4837`). A grep of the whole ledger for a gate-attributed catch
(gitleaks, vale, shellcheck, golangci blocking a commit) returns nothing. Every recorded defect was found
by the `diff-reviewer`, by a domain reviewer, or by the implementer's own test loop before commit. The
gate's value is therefore as the implementer's iteration signal, not as a second opinion after the fact.

**LOCAL-2: the reviewer's fix findings are mostly substantive, not cosmetic.** Sampling the four passes
whose close entries enumerate findings: real defects include a test suite that could not fail against
a doc-comment promise (`:4838`), a config field with zero production readers despite its comment
(`:4448`), a genuine `ENOENT` race with beets that aborted a whole capture pass (`:3705`), an
exclusion-stealing bug that would fabricate an oracle-skip signal (`:3712`), and named mutations that
did not fire (`:4450`). Comment-only or citation-only findings do occur (a stale citation, a doc-comment
clause, a missing printed row at `:3213`), but they are a minority of blocking findings, not the bulk.
The "is a full review dispatch justified by comment-only fix verdicts" question is therefore posed on a
false premise for this repo: comment-only findings are not what the reviewer mostly returns.

**LOCAL-3: second fix rounds and escalations have paid.** Pass 4a-i had five tasks take a fix round,
two of them twice, and the second rounds closed real defects (`:3676`). Two tasks escalated instead,
and both escalations corrected the *plan*, not the code: Task 1 proved the plan's own mutation narrative
was wrong (`:3686`), Task 8 forced a named constant with a pinning test (`:3695`). Escalation is doing
work that a mechanical second fix round would not.

**LOCAL-4: plan preparation, not execution, is the largest single line item.** Recorded budgets:
prep 0.83M of 3.9M (rung 5a, `:1688`), 0.87M of 2.9M (`:2173`), 0.87M of 3.6M (`:2580`), 1.07M of 3.7M
(`:2914`), 2.51M of 5.15M (rung 3d, `:4378`), 2.55M of about 5M (`:4788`), 2.106M of 4.82M (`:5165`).
Prep runs 22% to 49% of a pass, and in the two largest cases it exceeded the entire implementer chain.
The per-task chain, by contrast, is 1.5M to 2.3M across 16 to 24 agents.

**LOCAL-5: the simplifier is cheap and has never been recorded regressing behavior.** Costs 0.11M to
0.24M with a 0.05M to 0.07M review (`:2175`, `:3896`, `:5168`). Every recorded simplify review verdict is
"accept, no blockers" (`:3670`, `:3914`). One round "changed nothing" (`:2954`). Typical diffs are
9 to 18 files at roughly line-neutral to slightly negative (`:5191` 18 files +205/-223; `:4807` 15 files
+115/-157). Several passes also record deliberate "leftovers" the simplifier proposed and the reviewer
declined (`:5145`).

## 2. Element-by-element verdicts

### E1. Thin conductor: the main model never reads source, diffs, or gate transcripts during execution

**Verdict: supported, with one qualification.**

Anthropic's guidance is direct. "Since context is your fundamental constraint, use subagents to keep
research out of it... Subagents run in separate context windows and report back summaries"
(https://code.claude.com/docs/en/best-practices, living doc, read 2026-09-12). The costs page adds
"Delegate these to subagents so the verbose output stays in the subagent's context while only a summary
returns to your main conversation" (https://code.claude.com/docs/en/costs, read 2026-09-12). The context
engineering post quantifies the exchange: a subagent may burn "tens of thousands of tokens or more"
internally while returning "a condensed, distilled summary of its work (often 1,000-2,000 tokens)"
(https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents, 2025-09-29).

The independent support is the context-degradation literature. Chroma's context-rot study across 18
models reports needle-in-haystack accuracy drops of 20% to 50% between 10k and 100k+ tokens and finds
focused ~300-token prompts sharply outperform full ~113k-token prompts on LongMemEval
(https://www.trychroma.com/research/context-rot, July 2025). NoLiMa finds 10 of 13 models claiming 128K+
context fall below 50% of their short-context baseline at 32K tokens, GPT-4o from 99.3% to 69.7%
(https://arxiv.org/abs/2502.05167, ICML 2025). Cognition, arguing *for* one narrow multi-agent pattern,
names the same mechanism: "Context Rot is a well-documented phenomenon that is a result of models making
less intelligent decisions at longer and longer context lengths"
(https://cognition.com/blog/multi-agents-working, 2026-04-22).

**Qualification.** The single strongest counterweight in the literature is Cognition's original argument
that subagents lacking full context "make conflicting implicit decisions"
(https://cognition.com/blog/dont-build-multi-agents, 2025-06-12), and the equal-budget result that a
single agent with full context is information-theoretically never worse than message-passing between
agents at the same reasoning-token budget (Tran and Kiela, "Single-Agent LLMs Outperform Multi-Agent
Systems on Multi-Hop Reasoning Under Equal Thinking Token Budgets", https://arxiv.org/abs/2604.02460,
2026-04-02, revised 2026-04-11). Both bite on *decision* work, not on *reading*. The pattern here keeps
decisions in the conductor and moves only reading out, which is the split both sources endorse.

**Adjustment: none.** This is the best-supported element of the pattern.

### E2. Sonnet implementer, test-first, proving named mutations fail

**Verdict: supported, and the mutation-proof requirement is the strongest part.**

Test-first with an external checker is the one self-correction mechanism the literature endorses without
hedging. Huang et al. find intrinsic self-correction frequently fails to improve and sometimes degrades
performance, with prior positive results traced to oracle labels leaking into the loop
(https://arxiv.org/abs/2310.01798, ICLR 2024). CRITIC shows correction works once the model validates
against external tools (https://arxiv.org/abs/2305.11738, 2023). SWT-Bench reports that using generated
reproduction tests as a filter on proposed patches doubles SWE-Agent's precision
(https://arxiv.org/abs/2406.12952, NeurIPS 2024; SECONDARY).

The named-mutation requirement is the part with the sharpest support. MuTAP shows plain LLM-generated
tests are weak by mutation-testing standards, and that augmenting the prompt with surviving mutants
raises mutation score to 93.57% on synthetic buggy code against 66% for the coverage-based generator
Pynguin, detecting up to 28% more faulty human-written snippets
(https://arxiv.org/abs/2308.16557, 2023-08-31). Meta's production TestGen-LLM funnel is the scale
reference: 75% of generated test classes built, 57% passed reliably, 25% raised coverage
(https://arxiv.org/abs/2402.09171, 2024; SECONDARY). In other words, coverage flatters LLM-written tests
and only a kill-the-mutant check separates a test that can fail from one that cannot. LOCAL-2 confirms
the failure mode is live here: the reviewer repeatedly blocked on tests that could not fail.

**Adjustment: promote mutation proof from a convention to a gate artifact.** The plans already name
mutations per task, and LOCAL evidence shows three mutations in one pass recorded as unprovable rather
than papered over (`:3721`), which is the right behavior. The gap is that nothing mechanical checks the
claim. A per-task machine-readable mutation ledger (mutation, site, proving test, fired yes/no) that the
implementer emits and the reviewer grades would move this from prose the reviewer must reconstruct to a
table it can check, and would cut the fix rounds that currently spend a whole review cycle discovering
an unfired mutation.

### E3. The full repo gate runs inside the chain, per task

**Verdict: mixed. The gate is well justified as the implementer's feedback loop and poorly justified as a
per-task second opinion.**

Anthropic's framing is squarely about the iteration loop: "Give Claude a check it can run... Claude does
the work, runs the check, reads the result, and iterates until the check passes"
(https://code.claude.com/docs/en/best-practices). That is exactly what this pattern uses it for, and the
choice to run it inside the chain rather than the main loop is also directly supported (same page, on
keeping verbose output out of the conductor's context).

No published study compares per-task gates against batched gates for coding agents. That comparison is
**unstudied**. What LOCAL-1 supplies is the observation that in this repo the gate has not caught a
defect that the implementer's own loop and the reviewer did not, across at least five passes.

**Adjustment: extend the existing paint-moving carve-out to the whole-repo legs, on the same reasoning.**
The workstation ruling that browser suites join only for paint-moving tasks is the correct shape (see
E12). The same argument applies to gitleaks over the whole tracked worktree, vale over all Go comment
prose, `docker compose config`, `systemd-analyze verify`, and the state-dir agreement leg: none of them
can change verdict on a task that touched only `internal/auth`. Scope them to the blast radius per task
and run the full set once at chain end and on every push, which is what the repo already does for the
engine test suite. This is a clock-time and transcript-size adjustment, not a quality one, and it is
supported by absence of evidence rather than by a positive finding, so it should be reversible.

### E4. Fresh-context Opus diff-reviewer, criteria-anchored, returning accept/fix/escalate

**Verdict: supported on three of its four design choices, with one real risk the design does not yet
mitigate.**

*Fresh context: supported.* Anthropic: "A reviewer running in a fresh subagent context sees only the diff
and the criteria you give it, not the reasoning that produced the change, so it evaluates the result on
its own terms" (https://code.claude.com/docs/en/best-practices). Cognition's production numbers are the
best field evidence: Devin Review "catches an average of 2 bugs per PR, of which roughly 58% are severe
(logic errors, missing edge cases, security vulnerabilities)", and "the review agent having a completely
clean context also helps it go deeper into areas the original coding agent may not"
(https://cognition.com/blog/multi-agents-working, 2026-04-22). A direct measurement of the context-split
effect alone exists but is weak: cross-context review F1 28.6% against same-session self-review 24.6%
(https://arxiv.org/html/2603.12123; SECONDARY, not independently fetched).

*Criteria-anchored: strongly supported.* The single largest mitigation number in the LLM-as-judge
literature is reference-guided grading, which cuts the math-judging failure rate from 70% to 15%
(Zheng et al., https://arxiv.org/abs/2306.05685, 2023). Grading a diff against the task's written
acceptance criteria is the same mechanism. A 2026 cluster of rubric papers converges qualitatively on
decomposed, evidence-anchored rubrics beating holistic scores, though none yielded a clean delta I could
verify.

*A stronger model reviewing a weaker one: supported, and directionally important.* Xiang et al. ran 116
LCB hard and medium tasks in all four directions. Claude reviewing Codex raised pass rate from 71.6% to
89.7% (p=.001). Codex self-review reached only 84.5% (p=.022). Codex reviewing Claude *lowered* pass rate
from 91.4% to 82.8% (p=.046). Claude self-review left 91.4% unchanged. Conclusion verbatim: "the useful
pairing is asymmetric: use Claude to review Codex, not the other way around"
(https://arxiv.org/abs/2607.21656, 2026-07-22). Olausson et al. reached the same mechanism earlier:
self-repair is bottlenecked by the model's own critique ability, and using a stronger model for the
*feedback* (not the fix) produces substantially larger gains
(https://arxiv.org/abs/2306.09896, ICLR 2024). Opus reviewing Sonnet is on the right side of this arrow.

*The unmitigated risk: same-family bias.* Opus reviewing Sonnet is a same-family pairing, and this is
where the arrangement is weakest. Panickssery et al. measure self-preference scores above the 0.5
no-bias line at GPT-4 0.705, GPT-3.5 0.582, Llama 2 0.511, and show self-recognition ability correlates
linearly with self-preference strength (https://arxiv.org/abs/2404.13076, 2024). Wataoka et al. find the
bias tracks perplexity and familiarity rather than literal authorship, which is exactly the mechanism
that would carry across a model family (https://arxiv.org/abs/2410.21819, 2024). "Play Favorites"
documents same-family (not merely same-model) rating bias for GPT-4o and Claude 3.5 Sonnet
(https://arxiv.org/abs/2508.06709, 2025; no magnitude extractable). "Great Models Think Alike and this
Undermines AI Oversight" reports that error correlation across frontier models is rising over time and
argues this undermines oversight schemes generally (https://arxiv.org/abs/2502.04313, 2025-02;
abstract-level only). Note also that Zheng et al. measured Claude-v1 as the *worst* of the three judges
tested on position-bias consistency at 23.8%, and tied for worst on the verbosity attack at a 91.3%
failure rate. Those are 2023 model generations and should not be read forward, but they do mean "Claude
is a good judge" is not something the literature has established for the family as such.

**Adjustment: none to the reviewer's existence, one to its bias exposure.** Keep the reviewer. The single
cheapest hardening is to make the *high-risk* reviews cross-family rather than same-family. The pattern
already reserves a full fresh-context adversarial cycle for auth (rung 5), streamer/SSO (rung 12), and
cutover (rung 16). Those are the three places where a non-Anthropic reviewer would buy the most, and the
literature says the gain comes from decorrelated errors, not from the second model being better.

### E5. One fix round, then the conductor decides

**Verdict: supported, and the "conductor decides" half is the part doing the work.**

The convergence literature is consistent. Olausson et al. find self-repair gains modest and inconsistent,
sometimes absent once repair cost is counted, with i.i.d. resampling beating self-repair at small budgets
(https://arxiv.org/abs/2306.09896). Huang et al. find performance sometimes degrades after a
self-correction pass (https://arxiv.org/abs/2310.01798). The aggregated iterative-refinement picture
across Self-Refine, Reflexion and RCI-style work is that most gains land in rounds 1 to 2 and metrics
plateau by round 3 (SECONDARY, synthesized rather than pinned to one citation). Capping mechanical
rounds at one, with a human-equivalent decision after, is the shape that literature supports.

The over-correction finding sharpens it. Jin and Chen show LLM reviewers "frequently misclassify correct
code implementation as non-compliant or defective", and, counterintuitively, that prompts requesting
detailed explanations and corrections *increase* the error rate
(https://arxiv.org/abs/2603.00539, 2026-02-28; the specific 73.2%/87.9% false-negative figures reported
by search were not present in the abstract I fetched, so treat the direction as verified and the
magnitude as SECONDARY). Anthropic's own guidance says the same thing in plainer words: "A reviewer
prompted to find gaps will usually report some, even when the work is sound, because that is what it was
asked to do. Chasing every finding leads to over-engineering... Tell the reviewer to flag only gaps that
affect correctness or the stated requirements, and treat the rest as optional"
(https://code.claude.com/docs/en/best-practices).

**Does a second fix round ever pay?** LOCAL-3 says yes, but only when the conductor rules rather than
when the loop re-fires. In pass 4a-i the two twice-fixed tasks closed real defects, and the two
escalations both corrected the plan. The published evidence says a second *mechanical* round is where
returns flatten; it says nothing against a judged second round. The current rule already encodes exactly
this distinction and should stay.

**Adjustment: make the reviewer's severity label load-bearing.** Anthropic's advice is to have the
reviewer separate correctness gaps from optional findings. The CLAUDE.md gate-economy rule already has a
`commentOnly` marker for reduced fix-round gates. Extend that to a required per-finding severity
(blocking-correctness, blocking-contract, comment-only, optional) in the reviewer's fixed output shape,
and let a fix round whose findings are all comment-only or optional route to a doc-touch dispatch rather
than a full implementer re-dispatch. Caveat: no study measures LLM accuracy at severity labeling itself,
and the research lane found this to be a genuine evidence gap, so the severity should be advisory to the
conductor rather than an automatic bypass.

### E6. Batched `code-simplifier` pass per segment

**Verdict: contradicted in the general case, defensible as currently run, and the weakest-value element
in the pattern.**

The behavior-preservation risk is real and measured. Liu et al. tested ChatGPT and Gemini on 180 real
refactorings from 20 Java projects and found unsafe refactorings (functionality changed or syntax errors
introduced) at 13 of 176 solutions for ChatGPT and 9 of 137 for Gemini, roughly 7.4% and 6.6%. Their
proposed mitigation, RefactoringMirror, is a "detect-and-reapply tactic" that routes the change through
a tested refactoring engine rather than trusting the free-form LLM edit
(https://arxiv.org/abs/2411.04444, 2024-11-07). A larger 2026 study reports about 11.5% behavioral
inconsistency across 8,096 LLM-refactored segments (Empirical Software Engineering,
https://link.springer.com/article/10.1007/s10664-026-10911-6, 2026; paywalled, SECONDARY).

Against that, LOCAL-5 records zero behavior regressions across roughly a dozen simplifier rounds in this
repo. The difference is almost certainly the gate: every simplifier round here runs against a full test
suite plus a reviewer, which is the closest available approximation of RefactoringMirror's "reapply
through something tested" tactic. The papers measure unguarded refactoring.

The value side is where it is thin. Typical diffs are line-neutral (`:5191`, `:4807`), one round changed
nothing, and several passes record leftovers the reviewer declined. The strongest argument *for* keeping
it is SlopCodeBench, which measures long-horizon agentic degradation across 15 agents and finds
structural erosion rising in 77% of trajectories, verbosity in 75.5%, agent code 2.3x more verbose and
2.0x more eroded than 473 open-source Python repositories, and, critically, that "explicit quality
guidance reduces initial verbosity and erosion by up to a third, without affecting degradation rates"
(https://arxiv.org/abs/2603.24755, 2026-03-25, revised 2026-05-07). That last clause is the case for a
periodic external cleanup pass rather than better prompting: prompting moves the intercept, not the slope.

**Adjustment: keep it, batch it more coarsely, and keep the review.** The pass is cheap (0.11M to 0.24M),
never removable without losing the one mechanism that acts on SlopCodeBench's slope, and is currently run
with the guard the safety literature calls for. The only defensible change is frequency: per-segment
batching is more often than the evidence requires, and a per-rung or per-merge round would collect more
accumulated erosion per dispatch. Never drop the simplifier's own review; that review is what makes the
7.4% figure not apply here.

### E7. Plans drafted just in time by an Opus drafter

**Verdict: supported.**

Anthropic recommends separating planning from implementation for changes that are uncertain, multi-file,
or in unfamiliar code, and explicitly recommends against it when "you could describe the diff in one
sentence" (https://code.claude.com/docs/en/best-practices). The same page recommends interviewing the
owner and writing a spec, then "start a fresh session to execute it", and says "Time spent making the
spec precise pays off more than time spent watching the implementation". The just-in-time half is
supported by the local finding that the workstation already ratified: plan staleness is named in
CLAUDE.md as "the proven dominant defect class".

**Adjustment: none.**

### E8. Three to seven adversarial review lenses in parallel, folded by an editor

**Verdict: partly supported, and the most expensive element relative to its evidence.**

*For:* the ensembling result is real. SWR-Bench, built on 1,000 manually verified GitHub pull requests,
reports that a multi-review aggregation strategy increased F1 "by up to 43.67%" over single-pass review
(https://arxiv.org/abs/2509.01494, 2025-09-01, revised 2026-06-05). The panel-of-judges result points the
same way: a three-model PoLL panel beat a single GPT-4 judge on Cohen's kappa across KILT QA (0.763/0.906/
0.867 against 0.627/0.841/0.830) and on Chatbot Arena Hard correlation (Pearson 0.917 against 0.817), at
7 to 8 times lower cost, and "exhibits less intra-model bias due to its composition of disjoint model
families" (https://arxiv.org/abs/2404.18796, 2024).

*Against:* both results concern *diverse* panels. PoLL's stated bias advantage comes from disjoint model
families, which a fan-out of same-family lenses does not have. And the multi-agent literature is
consistently negative on adding parallel voices at equal compute: multi-agent debate does not reliably
beat self-consistency (https://arxiv.org/abs/2311.17371, 2023, rev. 2024); single agents match or beat
multi-agent systems at equal thinking-token budgets, grounded in the Data Processing Inequality
(https://arxiv.org/abs/2604.02460, 2026-04); and the MAST analysis of 1,600+ traces across 7 frameworks
found gains on popular benchmarks "often minimal" (https://arxiv.org/abs/2503.13657, 2025-03).

*The cost fact:* LOCAL-4. Plan prep is 22% to 49% of a pass and twice exceeded the whole implementer
chain. Nothing in the public record justifies seven lenses over three; SWR-Bench and PoLL both
demonstrate their effect with three.

**Adjustment: cap the default at three lenses, chosen for disjointness rather than count.** Three is what
the two supporting studies actually demonstrate. Choose the three for non-overlapping failure modes
(for example: a contract/criteria lens, a security or domain-risk lens, and a plan-staleness/consistency
lens) rather than adding a fourth of the same shape. Reserve five-plus for the three high-risk rungs the
pattern already singles out. On the current numbers this is the largest single token saving available
anywhere in the pattern, on the order of 0.5M to 1.5M per pass.

### E9. Two code lanes in parallel git worktrees, with a merge ritual

**Verdict: supported as an isolation mechanism, and the discipline around it matches the field reports.**

Anthropic lists worktrees first among parallelization mechanisms: "run separate CLI sessions in isolated
git checkouts so edits don't collide" (https://code.claude.com/docs/en/best-practices). Simon Willison's
report names the binding constraint correctly for this pattern's purposes: "the natural bottleneck on all
of this is how fast I can review the results", and parallelism pays when the work is well specified in
advance (https://simonwillison.net/2025/Oct/5/parallel-coding-agents/, 2025-10-05).

The strongest practitioner report is a measured incident log: seven concurrent sessions on one repo with
only file-level separation produced three "death spirals" ending in discarded work; worktrees alone were
insufficient at 13 concurrent agents, where git lock contention left 5 of 13 commits succeeding; and a
2GB repo grew to 9.82GB from per-worktree dependency installs
(https://readysolutions.ai/blog/2026-06-06-parallel-claude-code-sessions-one-repo/, 2026-06-06;
practitioner, and its cited 27.67% merge-conflict rate over 142,000+ parallel-agent PRs has unverified
provenance). The important finding there is that conflict probability tracked *task similarity*, not
session count. The workstation's contested-surface ownership table is precisely the control that finding
calls for, and a two-lane cap sits far below every threshold at which that report saw trouble.

Cognition's narrowed position is the one live tension: "Multi-agent systems work best today when writes
stay single-threaded and the additional agents contribute intelligence rather than actions"
(https://cognition.com/blog/multi-agents-working, 2026-04-22). Two code lanes are two write-threads. The
worktree plus ownership-table plus serialized merge ritual is the answer to that objection, and it is the
same three-part fix the incident report independently arrived at.

**Adjustment: none to the mechanism. One addition to the merge ritual.** Make the cross-lane
`diff-reviewer` dispatch conditional on semantic overlap rather than on whether the rebase was a
fast-forward. A clean fast-forward between lanes that both touched `internal/config` is riskier than a
conflicted rebase between lanes that touched disjoint packages, and the field evidence says similarity,
not merge mechanics, predicts the problem.

### E10. Token ceilings per pass with an 80% flag

**Verdict: unstudied as a mechanism, supported as an instinct.**

No published work evaluates per-task token budgets for agent workflows. Two adjacent results support the
instinct that more tokens stop helping and start hurting. "When More Thinking Hurts" finds marginal
utility turning negative beyond roughly 12K reasoning tokens for R1-32B, with both tested models peaking
near 10K and then declining (https://arxiv.org/html/2604.10739v1, 2026; SECONDARY). Inference-time
scaling work puts the optimal budget at about 1.0K tokens for easy problems and 7.5K for hard ones
(https://arxiv.org/pdf/2504.00294, 2025; SECONDARY). Against that, Anthropic's multi-agent post reports
"token usage by itself explains 80% of the variance" in their research eval
(https://www.anthropic.com/engineering/multi-agent-research-system, 2025-06-13), which cuts the other way
for breadth-first work.

LOCAL: the flag is crossed routinely and usually after every deliverable has landed (`:1688`, `:2173`,
`:2580`), and one entry records the conductor crossing the flag inside the chain and not acting on it
(`:2914`). The ceiling functions as a post-hoc accounting line more than a live control.

**Adjustment: re-key the flag to the point where it can still change a decision.** A flag that fires
during the final fold has no lever left to pull. Moving the check to the per-task checkpoint (which the
process already runs every four tasks) and comparing spend-to-date against tasks-remaining would make it
actionable. This is a housekeeping fix, not an evidence-driven one.

### E11. Long unattended runs with guards

**Verdict: supported, with one specific caution the local process does not yet carry.**

Anthropic's stop-gating ladder is the relevant guidance, including the detail that "Claude Code overrides
the hook and ends the turn after 8 consecutive blocks"
(https://code.claude.com/docs/en/best-practices). The caution is SlopCodeBench: quality erosion appears
in 77% of long-horizon trajectories and continues across iterations even when prompt guidance improves
the starting quality (https://arxiv.org/abs/2603.24755). Unattended length is itself a quality risk that
gates do not detect, because eroded code passes tests.

**Adjustment: treat run length as an input to the simplifier schedule.** If a pass ran unattended past
some threshold, the simplifier round for that segment should be non-optional and its reviewer should be
told to look for erosion and verbosity specifically, not only for reuse and altitude. This is the one
place where the erosion finding has an operational consequence.

### E12. Browser suites join the per-task gate only for paint-moving tasks

**Verdict: unstudied externally, and the local evidence is the best available.**

No study compares gate composition per task for coding agents. The CLAUDE.md rule already records the
evidence that produced it: across cairn chassis-B2's eight tasks, the per-task e2e suite "caught nothing
the checks, the diff reviewer, and CI did not, at a third of each task's hour." That is a defensible
local finding, it matches LOCAL-1's independent observation in a second repo, and it is the shape E3
recommends extending.

**Adjustment: none. Extend the reasoning rather than the rule (see E3).**

## 3. The three adjustments best supported by evidence

**A. Cut the default plan-review fan-out from three-to-seven lenses to three, chosen for disjoint failure
modes.** Both studies that support ensembling demonstrate the effect with three reviewers, and the panel
study attributes its bias advantage specifically to *disjoint model families* rather than to count
(https://arxiv.org/abs/2404.18796; https://arxiv.org/abs/2509.01494). The multi-agent literature is
consistently negative on additional same-shaped parallel voices at equal compute
(https://arxiv.org/abs/2604.02460; https://arxiv.org/abs/2311.17371; https://arxiv.org/abs/2503.13657).
Locally this is also the biggest line item in the pass: prep runs 22% to 49% of the budget and twice
exceeded the entire implementer chain. Highest confidence, largest saving, no quality argument against.

**B. Make the high-risk reviews cross-family, not just cross-context.** The asymmetric-review result is
the cleanest number in this whole brief: stronger-reviews-weaker moved pass rate 71.6% to 89.7%, while
the reverse direction *lowered* it 91.4% to 82.8% (https://arxiv.org/abs/2607.21656, 2026-07-22). Opus
reviewing Sonnet is on the right side of that arrow, so the arrangement is sound. The unaddressed risk is
same-family correlated blind spots, documented for self-preference
(https://arxiv.org/abs/2404.13076), for the familiarity mechanism that would carry across a family
(https://arxiv.org/abs/2410.21819), for same-family rating bias directly
(https://arxiv.org/abs/2508.06709), and as a rising trend across frontier models
(https://arxiv.org/abs/2502.04313). Apply it only where the pattern already spends a full adversarial
cycle: rungs 5, 12, and 16.

**C. Make the mutation ledger a graded artifact, and make finding severity part of the reviewer's fixed
output shape.** Two findings converge here. LLM-written tests are weak by mutation standards unless
something forces the kill-the-mutant check, and forcing it moves mutation score from roughly 66% to
93.57% (https://arxiv.org/abs/2308.16557). Reviewers over-report when asked to find gaps, and asking for
more explanation makes over-correction worse (https://arxiv.org/abs/2603.00539;
https://code.claude.com/docs/en/best-practices). A per-task mutation table the reviewer grades removes
the most common blocking finding class in this repo's own record, and a required severity label lets
comment-only rounds take a cheaper path. Caveat on the severity half: no study measures LLM accuracy at
severity labeling, so it should inform the conductor's decision rather than bypass it automatically.

## 4. The least-supported element

**The batched `code-simplifier` pass.** It is the only element in the pattern where the published
evidence points against the activity in its general form. Unguarded LLM refactoring changes behavior at
roughly 6.6% to 7.4% of attempts in a controlled study of 180 real refactorings
(https://arxiv.org/abs/2411.04444), and at about 11.5% across 8,096 segments in a larger 2026 study
(SECONDARY). Its measured benefit here is small: line-neutral diffs, one round that changed nothing, and
several proposals the reviewer declined.

Two things keep it defensible rather than cuttable. Its local record shows zero regressions, because it
always runs behind a full test suite and a review, which is the guard the refactoring papers themselves
recommend. And SlopCodeBench supplies the one argument that prompting cannot replace: quality guidance
"reduces initial verbosity and erosion by up to a third, without affecting degradation rates", so
something must act on accumulated erosion after the fact (https://arxiv.org/abs/2603.24755). Keep it,
batch it per rung rather than per segment, never run it without its review, and treat a long unattended
run as a trigger to run it rather than an occasion to skip it.

## 5. Evidence gaps worth naming

Four questions the pattern turns on have no usable public evidence.

- No study compares per-task against batched quality gates for coding agents.
- No study measures the correctness-versus-nit distribution of AI review output, or LLM accuracy at
  labeling severity. This is why adjustment C's severity half stays advisory.
- No controlled benchmark measures defects that AI review finds and human reviewers miss on a shared
  ground-truth corpus. Cognition's 2 bugs per PR at 58% severe is the closest field number and is
  vendor-reported (https://cognition.com/blog/multi-agents-working).
- No published data on AI-specific fix-induced regression rates, so "does a fix round introduce new
  defects" is unanswered externally. LOCAL-3 is the only evidence available and says the fixes held.

## 6. Sources

Anthropic and vendor:
- https://code.claude.com/docs/en/best-practices (living doc, read 2026-09-12)
- https://code.claude.com/docs/en/costs (read 2026-09-12)
- https://www.anthropic.com/engineering/multi-agent-research-system (2025-06-13)
- https://www.anthropic.com/engineering/building-effective-agents (2024-12-19)
- https://www.anthropic.com/engineering/effective-context-engineering-for-ai-agents (2025-09-29)
- https://www.anthropic.com/engineering/code-execution-with-mcp (2025-11-04)
- https://claude.com/blog/how-anthropic-teams-use-claude-code (2025-07-24)
- https://platform.claude.com/docs/en/build-with-claude/prompt-caching (read 2026-09-12)
- https://cognition.com/blog/dont-build-multi-agents (2025-06-12)
- https://cognition.com/blog/multi-agents-working (2026-04-22)
- https://manus.im/blog/Context-Engineering-for-AI-Agents-Lessons-from-Building-Manus (2025-07-18)
- https://www.trychroma.com/research/context-rot (2025-07)

Peer-reviewed and arXiv:
- https://arxiv.org/abs/2306.05685 (MT-Bench, LLM-as-judge, 2023)
- https://arxiv.org/abs/2404.13076 (self-recognition and self-preference, 2024)
- https://arxiv.org/abs/2410.21819 (self-preference tracks perplexity, 2024)
- https://arxiv.org/abs/2508.06709 (same-family self-bias, 2025)
- https://arxiv.org/abs/2502.04313 (correlated frontier-model errors, 2025-02)
- https://arxiv.org/abs/2404.18796 (Replacing Judges with Juries, 2024)
- https://arxiv.org/abs/2507.10535 (CodeJudgeBench, 2025-07-14, rev. 2025-08-14)
- https://arxiv.org/abs/2603.00539 (systematic overcorrection in code review, 2026-02-28)
- https://arxiv.org/abs/2607.21656 (cross-model code review asymmetry, 2026-07-22)
- https://arxiv.org/abs/2509.01494 (SWR-Bench, 2025-09-01, rev. 2026-06-05)
- https://arxiv.org/abs/2501.15134 (BitsAI-CR, 2025-01-25)
- https://arxiv.org/abs/2306.09896 (Is Self-Repair a Silver Bullet, ICLR 2024)
- https://arxiv.org/abs/2310.01798 (LLMs Cannot Self-Correct Reasoning Yet, ICLR 2024)
- https://arxiv.org/abs/2305.11738 (CRITIC, 2023)
- https://arxiv.org/abs/2308.16557 (MuTAP, 2023-08-31)
- https://arxiv.org/abs/2402.09171 (TestGen-LLM at Meta, 2024)
- https://arxiv.org/abs/2406.12952 (SWT-Bench, NeurIPS 2024)
- https://arxiv.org/abs/2411.04444 (LLMs in automated refactoring, 2024-11-07)
- https://arxiv.org/abs/2603.24755 (SlopCodeBench, 2026-03-25, rev. 2026-05-07)
- https://arxiv.org/abs/2604.02460 (single vs multi-agent at equal token budget, 2026-04-02)
- https://arxiv.org/abs/2503.13657 (MAST, why multi-agent systems fail, 2025-03)
- https://arxiv.org/abs/2311.17371 (Should we be going MAD, 2023, rev. 2024)
- https://arxiv.org/abs/2407.21787 (Large Language Monkeys, 2024-07-31)
- https://arxiv.org/abs/2307.03172 (Lost in the Middle, TACL 2023)
- https://arxiv.org/abs/2502.05167 (NoLiMa, ICML 2025)
- https://link.springer.com/article/10.1007/s10664-026-10911-6 (LLM refactoring consistency, 2026;
  paywalled, SECONDARY)

Practitioner:
- https://simonwillison.net/2025/Oct/5/parallel-coding-agents/ (2025-10-05)
- https://readysolutions.ai/blog/2026-06-06-parallel-claude-code-sessions-one-repo/ (2026-06-06)
- https://daniel.haxx.se/blog/2025/07/14/death-by-a-thousand-slops/ (2025-07-14)
- https://research.google/blog/resolving-code-review-comments-with-ml/ (2023-05-23)

Local:
- `~/Projects/dubplate/docs/HISTORY.md` (line references inline above)
- `~/.claude/CLAUDE.md`, `~/.claude/docs/model-economy.md`
- `~/Projects/dubplate/docs/superpowers/research/2026-09-04-agentic-coding-public-evidence.md`
