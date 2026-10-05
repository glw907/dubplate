package matcher

// QuietImportTunables returns DefaultTunables with the four assignments
// an unattended quiet import makes, which is the configuration rung 4's
// shadow observes under and rung 9's pipeline will import under.
//
// It lives here rather than in the harness because internal/engine may
// import no harness package (TestEngineNeverImportsHarness,
// internal/harness/conform/arch_test.go:141-146) and conform's
// gradeTunables (defaults.go:72-79) is on the wrong side of that wall.
// It adds no new surface concept: Quiet, QuietFallback, NoneRecAction
// and Timid are all existing Tunables fields.
//
// Two of the four already ship at these values (Timid at tunables.go:312,
// QuietFallback at :316) and are assigned anyway, so the set reads as one
// configuration rather than as a diff a reader has to compute.
func QuietImportTunables() Tunables {
	tun := DefaultTunables()
	tun.Quiet = true           // the -q flag the incumbent passes beet import
	tun.QuietFallback = "skip" // corpus/profiles/estate.yaml import.quiet_fallback
	tun.NoneRecAction = "skip" // corpus/profiles/estate.yaml import.none_rec_action
	tun.Timid = false          // corpus/profiles/estate.yaml import.timid
	return tun
}
