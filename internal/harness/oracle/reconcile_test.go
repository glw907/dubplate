package oracle

import (
	"context"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// fakevenvOptions points opts.Venv at testdata/fakevenv, a stand-in
// interpreter that ignores Reconcile's -c invocation and prints a
// fixed, well-formed component/version map, so this test can drive
// Reconcile's real venvVersions probe without a uv venv (this file
// carries no oracle build tag, and the repository gate has none).
func fakevenvOptions() Options {
	return Options{Venv: "testdata/fakevenv"}
}

// TestReconcileIgnoresAProbeComponentPinnedNeverChecks proves the
// direction that is actually silent: Reconcile only ever compares the
// components named in its own pinned tuple. A venv probe reporting an
// extra component absent from pinned disagrees with the manifest in
// plain membership, and Reconcile still reports no error, because its
// loop iterates pinned, never installed. This is exactly the shape a
// versionProbe edit that named a component pinned forgot to check
// would take: silent, not loud. See
// TestReconcileNamesTheDisagreeingComponent's "wrong mediafile
// version" case for the loud counterpart, proving the sixth pin this
// package actually added is wired on both sides.
func TestReconcileIgnoresAProbeComponentPinnedNeverChecks(t *testing.T) {
	t.Setenv("FAKEVENV_EXTRA_COMPONENT", "somelib")

	m := &cassette.Manifest{
		Beets:     "2.13.1",
		Unidecode: "1.4.0",
		Jellyfish: "1.2.1",
		Lap:       "0.5.13",
		Numpy:     "2.5.2",
		MediaFile: "0.17.0",
	}

	if err := Reconcile(context.Background(), m, fakevenvOptions()); err != nil {
		t.Errorf("Reconcile with a probe-only extra component: got error %v, want nil (pinned never checks it)", err)
	}
}

// TestReconcileCatchesAMediaFileMismatchWithoutTheRealVenv is the
// loud counterpart, available in the untagged gate rather than only
// under make derive: a manifest whose mediafile pin disagrees with
// what the (fake) venv reports is caught and named, proving the sixth
// pin is actually wired into the pinned tuple Reconcile compares
// against, not merely into versionProbe's string.
func TestReconcileCatchesAMediaFileMismatchWithoutTheRealVenv(t *testing.T) {
	m := &cassette.Manifest{
		Beets:     "2.13.1",
		Unidecode: "1.4.0",
		Jellyfish: "1.2.1",
		Lap:       "0.5.13",
		Numpy:     "2.5.2",
		MediaFile: "0.0.1", // fakevenv always reports 0.17.0
	}
	err := Reconcile(context.Background(), m, fakevenvOptions())
	if err == nil {
		t.Fatal("Reconcile with a wrong mediafile pin: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "mediafile") {
		t.Errorf("error %q does not name mediafile", err)
	}
}
