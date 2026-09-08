//go:build linux

package serve

import "syscall"

// freebindControl is a net.ListenConfig.Control hook that sets
// IP_FREEBIND on the listening socket before bind(2) runs, so Listen can
// bind an address the host has not configured on any interface yet
// (R20). syscall.IP_FREEBIND is the stdlib's own constant (R-M): its
// source is Linux's <linux/in.h>, and on this toolchain (go1.27.1,
// linux/amd64) it measures 15 (0xf), which Task 0 confirmed against a
// live bind before this task built on it. The option is set at
// IPPROTO_IP for both address families rather than at IPV6_FREEBIND for
// AF_INET6, measured working on an AF_INET6 listener; no IPV6_FREEBIND
// constant exists for this architecture, so none is referenced.
//
// IP_FREEBIND's residual: it suppresses the boot-time error a
// misconfigured or not-yet-present address would otherwise raise. It
// does not make the address reachable, so a listener bound this way to
// an address nothing routes to accepts no traffic; surfacing that state
// is 5d's doctor row, not this package's concern.
func freebindControl(_, _ string, c syscall.RawConn) error {
	var sockErr error
	if err := c.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_FREEBIND, 1)
	}); err != nil {
		return err
	}
	return sockErr
}
