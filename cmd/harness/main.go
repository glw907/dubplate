// Command harness is dubplate's workstation-only dev tool for the
// conformance corpus's capture-derive loop (rung 2's cassette
// substrate). It is never shipped, and the repository gate never runs
// it against the live network: `go test` exercises only its request
// planning and file I/O against fakes.
//
//   - capture --mbid <release> --out <dir>: seeds a case by fetching
//     one release lookup, and, for a Pseudo-Release, its linked
//     actual release too, at beets' own maximal inc= (U6).
//   - capture --case <dir>: fetches exactly the requests a
//     derivation's stub could not answer, appending each to the case.
//   - freeze --case <dir>: derives a case and writes its expected
//     record back to case.json, refusing to do so when the
//     derivation leaves the case half-derived.
//
// Every live request carries the identified User-Agent
// corpus/README.md documents; the tool refuses to run at all when
// DUBPLATE_HARNESS_CONTACT is unset.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at release build
// time; it stays "dev" for workstation invocations, which is every
// invocation this tool ever sees.
var version = "dev"

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "harness",
		Short:         "Workstation dev tool for the conformance corpus's capture-derive loop",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(newCaptureCmd(os.Environ(), version))
	cmd.AddCommand(newFreezeCmd())
	return cmd
}
