# The oracle venv

`internal/harness/oracle/derive_oracle.py` drives real, pinned beets to
derive the conformance corpus's expected records. This directory pins
the environment it runs under. Nothing here ships to the box, and the
repo gate (`bash scripts/check.sh`) never needs it: derivation tests
carry the `oracle` build tag and run only through `make derive`.

## Build it

```sh
uv venv --python 3.13 tools/venv/.venv
uv pip install --python tools/venv/.venv/bin/python -r tools/venv/requirements.txt
```

The venv itself is gitignored. `requirements.in` is the intent (one
line: the beets pin) and `requirements.txt` is the resolved lock that
the install reads.

`make derive` looks for the venv at `tools/venv/.venv`. Set
`DUBPLATE_ORACLE_VENV` to point the derivation tests somewhere else.

## What the pins mean

A derived record is only meaningful against the exact versions that
produced it. unidecode's transliteration tables, jellyfish's edit
distance, and lap's assignment solver all sit directly in beets'
distance path, so a silent bump in any of them can move a verdict with
nothing in the corpus to show for it.

`corpus/manifest.json` records the versions the corpus was derived
under, and `oracle.Reconcile` compares that record against the live
venv before a derivation is believed, naming any component that
disagrees. A version change is a corpus-revalidation event, the same as
a config profile edit: re-derive, adjudicate what moved, then update
the manifest.

The interpreter is CPython 3.13. The shim serializes its own floats
with `%.17g` rather than leaning on the interpreter's `repr`, so a
patch-level CPython difference cannot silently reshape a record, but
the derivation environment is still worth keeping on one Python minor
version.

## Re-pin

```sh
uv pip compile --python tools/venv/.venv/bin/python \
  tools/venv/requirements.in -o tools/venv/requirements.txt
```

Then rebuild the venv, run `make derive`, and update
`corpus/manifest.json` with whatever `Reconcile` names.
