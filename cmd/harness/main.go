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
//   - synth --spec <file> --out <dir>: builds a synthetic case's
//     case.json and release response from a compact release spec, the
//     input side of the authoring loop for every structural-stratum
//     case (corpus/README.md, "The synth tool").
//   - seed-items --case <dir>: projects a staged case's Input.Items
//     from its own captured response, the sanctioned way to author a
//     real-catalog case without an agent reading responses/*.json
//     directly (corpus/README.md, "The seed-items tool").
//   - fetch-unidecode: fills the workstation's cached copy of
//     Unidecode's pinned wheel, a no-op once the cache verifies. `make
//     unidecode-wheel` runs it; the gate never does.
//   - oracle-up: builds the streamer fixture library, runs the pinned
//     Navidrome and gonic oracles on it under podman, prints their
//     loopback URLs and run credentials, and removes them on interrupt.
//   - record-oracle: runs the same oracles and writes a session of the
//     answers each gives to a spec's requests, resolving symbolic
//     references to each oracle's own ids.
//
// Every live request carries the identified User-Agent
// corpus/README.md documents; the tool refuses to run at all when
// DUBPLATE_HARNESS_CONTACT is unset.
package main

import (
	"fmt"
	"net/http"
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
	cmd.AddCommand(newSynthCmd())
	cmd.AddCommand(newSeedItemsCmd())
	cmd.AddCommand(newFetchUnidecodeCmd(&http.Client{Timeout: fetchUnidecodeTimeout}))
	cmd.AddCommand(newOracleUpCmd())
	cmd.AddCommand(newRecordOracleCmd())
	return cmd
}
