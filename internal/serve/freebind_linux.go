//go:build linux

package serve

import "syscall"

// freebindControl is a net.ListenConfig.Control hook that sets
// IP_FREEBIND on the listening socket before bind(2) runs, so listen can
// bind an address the host has not configured on any interface yet.
// The option is set at IPPROTO_IP for both address families: Linux
// honors it on an AF_INET6 socket, and syscall defines no
// IPV6_FREEBIND.
//
// IP_FREEBIND suppresses the boot-time error a misconfigured or
// not-yet-present address would otherwise raise. It does not make the
// address reachable, so a listener bound this way to an address nothing
// routes to accepts no traffic; the doctor's bind row surfaces that
// state.
func freebindControl(_, _ string, c syscall.RawConn) error {
	var sockErr error
	if err := c.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_FREEBIND, 1)
	}); err != nil {
		return err
	}
	return sockErr
}
