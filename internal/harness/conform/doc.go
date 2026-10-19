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
// (docs/plans/2026-09-04-rung-3a-scoring-core.md, Task 6).
// Nothing here writes corpus/ or any other harness package; it only
// reads corpus/analysis/components.json, corpus/manifest.json, and
// corpus/cases, all read-only.
//
// Rung 3, pass 3c adds a second tier: request.go grades request
// CONSTRUCTION, replaying every committed case's own request log
// through the real internal/engine/musicbrainz client against an
// internal/harness/stub server built from that same log. This is the
// first place conform drives a stub rather than only decoding one
// case's committed bytes by hand, because a request is bytes on a
// wire and a wire has two ends: the only way to grade what the client
// actually builds is to serve it a response and let it try. The
// charter above still holds, read-only and no harness package written;
// starting and closing a stub per case is driving, not writing.
//
// Rung 3, pass 3d adds tier 1: case.go replays every committed case
// through the matcher's own public entry point
// (internal/engine/tag.Album) against a stub built from that case's
// log, and grades the run against the case's derived record, field by
// field, under the grading policy the four ledgers in corpus/analysis
// give that case. It grades the request SET as well as the record: a
// run that leaves a logged request unissued took a different flow than
// pinned beets took, which no miss count can see.
//
// The runner reports a pending divergence and never writes one.
// Nothing here writes corpus/: a gate that edits the corpus is not a
// gate, because the expectation would become a function of the run.
// The failure message carries what a human needs to rule instead.
package conform
