---
paths:
  - "internal/harness/oracle/*.py"
  - "internal/engine/matcher/testdata/gen_featuring_golden.py"
  - "tools/unidecode-table/generate.py"
---

# Python

**Python** (`internal/harness/oracle/*.py`, `internal/engine/matcher/testdata/gen_featuring_golden.py`,
`tools/unidecode-table/generate.py`):
invoke the `python-conventions` skill BEFORE writing any Python. Python
stays confined to workstation-only derivation and generation tooling;
`generate.py` only re-derives the table pin (its output is never
committed; the binary builds the table at setup with
`dubplate-server unidecode install`);
nothing Python ships to the box. The oracle shim drives pinned beets
under a uv-managed venv (`tools/venv/`) that the gate never requires and
never runs (derivation is `make derive`, tagged `oracle`, workstation
only). ruff (`ruff check` and `ruff format --check`) joins `scripts/check.sh`
for every tracked `*.py` file, so the shim's bar is enforced, not
aspirational, even though the gate can't execute it.
