package serve

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// ErrInvalidBindAddr is the sentinel [SplitBindAddrs] wraps when one
// entry of a comma-separated bind-address list fails to parse as an IP
// address.
var ErrInvalidBindAddr = errors.New("invalid bind address")

// SplitBindAddrs splits raw, DUBPLATE_SERVE_BIND_ADDRS' own
// comma-separated grammar, into the individual addresses [Listen]
// binds. Each entry is trimmed of surrounding ASCII space before it is
// checked, so "127.0.0.1, ::1" and "127.0.0.1,::1" split the same way.
// An empty entry, from a doubled or trailing comma, and an entry that
// does not parse as an IP address are both refused, wrapping
// [ErrInvalidBindAddr] and naming the offending value, rather than
// silently dropped: a shell that silently dropped a malformed address
// would bind fewer interfaces than the operator configured with no
// signal that it happened.
//
// The grammar lives here rather than in the config declaration (R-B):
// DUBPLATE_SERVE_BIND_ADDRS registers as a string threshold because
// the registry has no list kind, and the address grammar belongs with
// the code that binds rather than with a generic string threshold.
func SplitBindAddrs(raw string) ([]string, error) {
	parts := strings.Split(raw, ",")
	addrs := make([]string, 0, len(parts))
	for _, part := range parts {
		addr := strings.TrimSpace(part)
		if addr == "" {
			return nil, fmt.Errorf("%q: empty entry: %w", raw, ErrInvalidBindAddr)
		}
		if net.ParseIP(addr) == nil {
			return nil, fmt.Errorf("%q: %w", addr, ErrInvalidBindAddr)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}
