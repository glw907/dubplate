//go:build !linux

package serve

import "syscall"

// freebindControl does nothing off Linux, where syscall defines no
// IP_FREEBIND; it keeps a cross-compile of the tree building. An
// unconfigured address fails at bind there.
func freebindControl(_, _ string, _ syscall.RawConn) error {
	return nil
}
