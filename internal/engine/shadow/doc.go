// Package shadow observes the incumbent importer's live decisions and
// compares them against the Go matcher's, so the launch bar's own
// question (does the Go matcher decide what beets decided, on this
// estate's own material) is answered by measurement rather than by
// argument. It reads the inbox the incumbent is about to import, scores
// each album through internal/engine/tag, reconciles the result against
// the incumbent's own run reports, and records every divergence with
// the evidence to re-evaluate it later. It applies nothing and writes
// no tag: the incumbent remains the only importer.
//
// It imports no internal/harness package, because the engine is the
// product and the harness grades it, and it imports internal/config
// never: every path, knob, budget, User-Agent and matcher stamp reaches
// this package as a parameter, on the same grounds
// internal/engine/matcher and internal/engine/musicbrainz state for
// themselves (tunables.go:50-54, client.go:86-92). It reads no clock
// either. Every evaluation date and pass time is a parameter, so a test
// pins them and two runs of one fixture agree.
package shadow
