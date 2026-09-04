// Package conform is rung 3's tier 0: the typed decoder for
// corpus/analysis/components.json and the runner that replays every
// committed row against the Go matcher's own primitives
// (internal/engine/matcher).
//
// It lives on the harness side of the wall, not the engine's, because
// what it does here is read a harness artifact (a corpus table) and
// drive engine primitives to grade them against it. That is the
// harness -> engine direction the pass's first architectural invariant
// requires: internal/engine never imports internal/harness (see
// TestEngineNeverImportsHarness in arch_test.go, and
// internal/engine/matcher's own doc comment). Putting a conformance
// runner inside the package it grades would invert that dependency for
// a test-only convenience, and the runner would still need to import
// analysis and cassette types the engine has no reason to know about.
//
// This package births internal/harness/conform: rung 3, pass 3a's own
// new territory
// (docs/superpowers/plans/2026-09-04-rung-3a-scoring-core.md, Task 6).
// Nothing here writes corpus/ or any other harness package; it only
// reads corpus/analysis/components.json, corpus/manifest.json, and
// corpus/cases, all read-only.
package conform
