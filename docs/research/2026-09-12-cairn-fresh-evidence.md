# Fresh evidence 2026-09-12: cairn's runner numbers relayed by Geoff, and dubplate's counterpart (post-dates cairn's HISTORY Chassis-B2 entry)

Quoted from the cairn session's report to Geoff, verbatim:

"Yes, the runner change is measurably paying off. Before it, Task 4 ran 75 minutes with 381
shell calls at up to 427K tokens of context per call, and Task 7 spent 263 of its 562 calls
polling a log. After it, Task 8 made four runner calls and zero polls, and Tasks 10 through 12
each landed in roughly 20 to 30 minutes with no fix round. By workflow segment, the first
stretch (Tasks 4 to 7) cost 1.73M subagent tokens for 14 agents; the stretch covering Tasks 10
to 13 cost 0.75M for 15 agents, so per-agent cost roughly halved on comparable work. The
comment-only reduced gate and the blast-radius scoping also held: every fix round tonight ran
in under four minutes."

Reading: the blocking runner (cairn-run-gate, no log polling) is the single largest per-agent
saving cairn has measured (about 2x per agent on comparable work); the comment-only reduced gate
and blast-radius gate scoping bound fix rounds to under four minutes of clock.

## The dubplate-side counterpart, measured 2026-09-12

The runner change cairn is measuring is already half-landed here. Every chain agent in the
2026-09-12 two-lane window shows `cairn-run-gate` calls, so dubplate's implementers run the
gate through the blocking runner rather than backgrounding it. What remains is residual
polling around it: up to fourteen `tail`-or-`sleep` calls in a single agent, and up to 118
shell calls in the heaviest task (agent `a4c457b6` in run `wf_c808ec5e-1ba`). cairn's
pre-runner counts on comparable tasks were 381 and 562 shell calls, so dubplate's ceiling
today sits at roughly a fifth to a third of what cairn paid before its change, and the
remaining polling is a residue rather than the dominant shape.

Reading: dubplate cannot collect cairn's 2x per-agent saving, because it already took the
larger half of it. The residue is worth one edit to the implementer definition and nothing
more.
