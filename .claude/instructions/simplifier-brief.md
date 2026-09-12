# Simplifier brief (dubplate)

The conductor pastes this file into every `code-simplifier` dispatch for this repo. It
replaces a clarity brief, because clarity is what the comments this repo over-writes already
claim to serve.

## Hunt these four patterns

All four are decidable by reading. Report each hit with `file:line` and the collapsed form.

1. Two function bodies in one package that differ by one statement. Keep one
   implementation; the other calls it with the statement as its argument.
2. A free function whose first parameter is the package's own type. Make it a method on
   that type.
3. A function returning three or more results of the same type. Return one struct whose
   field names say what each value is.
4. A struct field, method, or parameter whose only non-test caller is the package that
   declares it. Unexport it, or delete the seam and let the rung that needs it add it with
   its consumer in the same commit.

Grep the whole module before claiming a symbol has no caller outside its own file. A
same-package test still reaches an unexported symbol, so unexporting breaks nothing.

## Standing rules

- Preserve every assertion. A test that changes shape must still fail for the same reason.
- No em dashes anywhere.
- Commit only if `bash scripts/check.sh` is green, run through `cairn-run-gate`.
- Never touch `docs/STATUS.md`, `docs/HISTORY.md`, or `ROADMAP.md`.
