//go:build linux

package serve

import "syscall"

// freebindControl is a [net.ListenConfig] Control hook that sets
// IP_FREEBIND before bind(2), so listen can bind an address no
// interface carries yet. Linux honors the IPPROTO_IP option on an
// AF_INET6 socket too. The address is not made reachable: a listener
// on one nothing routes to accepts no traffic, which the doctor's bind
// row surfaces.
func freebindControl(_, _ string, c syscall.RawConn) error {
	var sockErr error
	if err := c.Control(func(fd uintptr) {
		sockErr = syscall.SetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_FREEBIND, 1)
	}); err != nil {
		return err
	}
	return sockErr
}
