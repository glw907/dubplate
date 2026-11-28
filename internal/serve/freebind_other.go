//go:build !linux

package serve

import "syscall"

// freebindControl does nothing off Linux. IP_FREEBIND is a Linux socket
// option, and syscall.IP_FREEBIND is defined on no other platform, so a
// single unconstrained definition breaks the cross-compile a
// contributor runs to check the tree builds (GOOS=darwin go build
// ./...). Every deployment target is Linux, so binding an address the
// host has not configured yet is a Linux property; elsewhere listen
// opens an ordinary listener and an unconfigured address fails at bind.
func freebindControl(_, _ string, _ syscall.RawConn) error {
	return nil
}
