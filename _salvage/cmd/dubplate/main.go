// Command dubplate is the estate's single binary: import, backup, checks,
// and the doctor preflight for a self-hosted Navidrome family music server.
package main

import (
	"fmt"
	"os"
)

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
