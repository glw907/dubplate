package gate

import (
	"os/exec"
	"testing"
)

// TestGateBuildsOnEveryPlatform is AC 34: the reason set, the gate core, and
// the staging leaf compile for every platform the project ships a client
// for, so a Linux-only call (a syscall.Statfs, say) cannot slip into a file
// without a build constraint.
//
// The build runs in a subprocess, which go test's cache cannot see, so a
// cached pass would replay after staging changed; the Makefile runs this
// test with -count=1 for that reason.
func TestGateBuildsOnEveryPlatform(t *testing.T) {
	for _, goos := range []string{"darwin", "windows"} {
		t.Run(goos, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), "go", "build", "./internal/reject", "./internal/intake/gate", "./internal/intake/staging")
			cmd.Dir = "../../.."
			cmd.Env = append(cmd.Environ(), "GOOS="+goos, "GOARCH=amd64", "CGO_ENABLED=0")
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Errorf("GOOS=%s go build: %v\n%s", goos, err, out)
			}
		})
	}
}
