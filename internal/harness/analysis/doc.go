// Package analysis holds the gate-resident checks over the conformance
// corpus's derived analyses: the threshold-proximity flag, the
// assignment-tie ledger, the per-case claims, and the component
// derivation table. Everything here is pure Go over committed files, so
// the repository gate runs all of it with no venv, no network, and no
// shim process.
//
// Two of the four analyses have a derivation half that beets alone can
// compute (internal/harness/oracle/derive_ties.py and
// derive_component.py, both behind the oracle build tag). This package
// never runs them: it reads what they committed and checks it against
// what the corpus's own records say, so a stale ledger fails the gate
// rather than aging quietly.
package analysis
