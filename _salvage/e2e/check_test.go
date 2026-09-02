package e2e

import "testing"

// --- dubplate check ------------------------------------------------------

// TestCheckBareArgExitsNonzero proves `dubplate check` with no name
// exits nonzero rather than silently no-op'ing: a typoed or missing
// argument in a systemd ExecStart line must not look healthy.
func TestCheckBareArgExitsNonzero(t *testing.T) {
	e := newTestEnv(t)

	output, code := e.run("check")
	if code == 0 {
		t.Fatalf("check exited 0, want nonzero with no argument: %s", output)
	}
}

// TestCheckUnknownNameExitsNonzero proves `dubplate check bogus` exits
// nonzero: an unknown check name is a typoed ExecStart, not a silent
// no-op that would report healthy.
func TestCheckUnknownNameExitsNonzero(t *testing.T) {
	e := newTestEnv(t)

	output, code := e.run("check", "bogus")
	if code == 0 {
		t.Fatalf("check bogus exited 0, want nonzero for an unknown check name: %s", output)
	}
}

// --- dubplate doctor -------------------------------------------------------

// TestDoctorMissingRequiredEnvExitsNonzero proves `dubplate doctor` exits
// nonzero when config does not resolve, the same as every other config
// consumer, rather than reporting success because the probes were merely
// skipped.
func TestDoctorMissingRequiredEnvExitsNonzero(t *testing.T) {
	e := newTestEnv(t)
	delete(e.env, "DUBPLATE_SMTP_HOST")

	output, code := e.run("doctor")
	if code == 0 {
		t.Fatalf("doctor exited 0, want nonzero with DUBPLATE_SMTP_HOST unset: %s", output)
	}
}
