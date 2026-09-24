package auth

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// registeredDefaults restates the defaults cmd/dubplate-server registers
// for DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR (5),
// DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR (20),
// DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR (30), and
// DUBPLATE_LINK_MAIL_CEILING_PER_HOUR (60). This package cannot import
// package main, so the values are restated once here, and every test
// below reads its expectation through a field of this value rather
// than through a literal of its own.
var registeredDefaults = Limits{
	LinkRequestsPerAddress:       5,
	LinkRequestsPerClient:        20,
	ConsumptionAttemptsPerClient: 30,
	LinkMailCeiling:              60,
}

var limitEpoch = time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)

func address(i int) string { return fmt.Sprintf("member%d@example.com", i) }

// client returns the i'th client address, each in a /64 of its own, so
// every i is a distinct client key.
func client(i int) string { return fmt.Sprintf("2001:db8:%x:%x::1", i>>16, i&0xffff) }

// TestLinkRequestLimitsTrip proves each link-request limit refuses on
// its own count and names itself, and that a fresh key is still
// admitted once it has.
func TestLinkRequestLimitsTrip(t *testing.T) {
	tests := []struct {
		name     string
		admitted int
		allow    func(l *Limiter) error
		fresh    func(l *Limiter) error
		wantText string
	}{
		{
			name:     "one address trips the per-address limit",
			admitted: registeredDefaults.LinkRequestsPerAddress,
			allow:    func(l *Limiter) error { return l.allowLinkAddress(limitEpoch, "member@example.com") },
			fresh:    func(l *Limiter) error { return l.allowLinkAddress(limitEpoch, "other@example.com") },
			wantText: "link requests per address",
		},
		{
			name:     "one client trips the per-client limit",
			admitted: registeredDefaults.LinkRequestsPerClient,
			allow:    func(l *Limiter) error { return l.allowLinkClient(limitEpoch, "192.0.2.7") },
			fresh:    func(l *Limiter) error { return l.allowLinkClient(limitEpoch, "203.0.113.9") },
			wantText: "link requests per client",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			for i := range tt.admitted {
				if err := tt.allow(l); err != nil {
					t.Fatalf("request %d of %d refused: %v", i+1, tt.admitted, err)
				}
			}
			err := tt.allow(l)
			if !errors.Is(err, ErrRateLimited) {
				t.Fatalf("request %d error = %v, want wrapping ErrRateLimited", tt.admitted+1, err)
			}
			if got := err.Error(); got != ErrRateLimited.Error()+": "+tt.wantText {
				t.Errorf("error = %q, want the %q limit named", got, tt.wantText)
			}
			if err := tt.fresh(l); err != nil {
				t.Errorf("fresh key refused: %v", err)
			}
		})
	}
}

// TestLimitsRefillOverTheInterval proves each count is per
// [limitInterval]: a spent key is admitted its full count again one
// interval later, with no real time passing.
func TestLimitsRefillOverTheInterval(t *testing.T) {
	tests := []struct {
		name  string
		count int
		allow func(l *Limiter, now time.Time) error
	}{
		{"consumption attempts per client", registeredDefaults.ConsumptionAttemptsPerClient, func(l *Limiter, now time.Time) error {
			return l.AllowConsumption(now, "192.0.2.1")
		}},
		{"link mail ceiling", registeredDefaults.LinkMailCeiling, func(l *Limiter, now time.Time) error {
			return l.AllowLinkMail(now)
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			for _, now := range []time.Time{limitEpoch, limitEpoch.Add(limitInterval)} {
				for i := range tt.count {
					if err := tt.allow(l, now); err != nil {
						t.Fatalf("at %v: attempt %d of %d refused: %v", now, i+1, tt.count, err)
					}
				}
				if err := tt.allow(l, now); !errors.Is(err, ErrRateLimited) {
					t.Fatalf("at %v: attempt %d error = %v, want wrapping ErrRateLimited", now, tt.count+1, err)
				}
			}
		})
	}
}

// TestLinkMailCeilingIgnoresKeys proves the global ceiling stops mail
// after its count even when every request carries a fresh address from
// a fresh client, so neither keyed limit ever trips.
func TestLinkMailCeilingIgnoresKeys(t *testing.T) {
	l := NewLimiter(registeredDefaults)
	for i := range registeredDefaults.LinkMailCeiling + 1 {
		if err := l.allowLinkClient(limitEpoch, client(i)); err != nil {
			t.Fatalf("request %d from a fresh client refused: %v", i+1, err)
		}
		if err := l.allowLinkAddress(limitEpoch, address(i)); err != nil {
			t.Fatalf("request %d for a fresh address refused: %v", i+1, err)
		}
		err := l.AllowLinkMail(limitEpoch)
		if i < registeredDefaults.LinkMailCeiling {
			if err != nil {
				t.Fatalf("mail %d of %d refused: %v", i+1, registeredDefaults.LinkMailCeiling, err)
			}
			continue
		}
		if !errors.Is(err, ErrRateLimited) {
			t.Errorf("mail %d error = %v, want wrapping ErrRateLimited", i+1, err)
		}
	}
}

// TestKeyedLimiterBoundHoldsUnderFlood drives strictly more distinct
// keys than [maxLimiterKeys] through each key space and samples the
// map's length after every key, so a bound that is only restored at the
// end still fails. The per-address row stops a hundred keys past the
// bound: at one instant no bucket has refilled, so every key past it
// scans the whole map, and a second full map's worth would spend the
// run on that documented worst case rather than on the bound.
func TestKeyedLimiterBoundHoldsUnderFlood(t *testing.T) {
	tests := []struct {
		name  string
		keys  int
		drive func(l *Limiter, i int)
		space func(l *Limiter) *keyedLimiter
	}{
		{"link requests per address", maxLimiterKeys + 100, func(l *Limiter, i int) {
			_ = l.allowLinkAddress(limitEpoch, address(i))
		}, func(l *Limiter) *keyedLimiter { return l.perAddress }},
		{"link requests per client", 2*maxLimiterKeys + 1, func(l *Limiter, i int) {
			_ = l.allowLinkClient(limitEpoch, client(i))
		}, func(l *Limiter) *keyedLimiter { return l.perClient }},
		{"consumption attempts per client", 2*maxLimiterKeys + 1, func(l *Limiter, i int) {
			_ = l.AllowConsumption(limitEpoch, client(i))
		}, func(l *Limiter) *keyedLimiter { return l.consumption }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			space := tt.space(l)
			for i := range tt.keys {
				tt.drive(l, i)
				if n := space.size(); n > maxLimiterKeys {
					t.Fatalf("after %d distinct keys the map holds %d, bound is %d", i+1, n, maxLimiterKeys)
				}
			}
			if n := space.size(); n != maxLimiterKeys {
				t.Errorf("after the flood the map holds %d, want exactly %d", n, maxLimiterKeys)
			}
		})
	}
}

// TestKeyedLimiterEvictsLeastRecentlyUsed fills a map to its bound,
// touches a spent key, and admits one more: the least recently used key
// goes, and the spent key keeps its spent bucket.
func TestKeyedLimiterEvictsLeastRecentlyUsed(t *testing.T) {
	k := newKeyedLimiter(1, 3, evictLeastRecent)
	k.allow(limitEpoch, "victim")
	k.allow(limitEpoch, "oldest")
	k.allow(limitEpoch, "middle")
	if k.allow(limitEpoch, "victim") {
		t.Fatal("victim admitted twice on a count of one")
	}
	k.allow(limitEpoch, "newest")

	if _, ok := k.byKey["oldest"]; ok {
		t.Error("least recently used key survived eviction")
	}
	for _, key := range []string{"victim", "middle", "newest"} {
		if _, ok := k.byKey[key]; !ok {
			t.Errorf("key %q evicted, want it kept", key)
		}
	}
	if k.allow(limitEpoch, "victim") {
		t.Error("victim admitted after eviction ran: its spent bucket was reset")
	}
}

// TestKeySpacesDoNotShareBuckets proves the three key spaces are
// separate maps: one client address spent in one space is untouched in
// another, and a flood that evicts every key in one space leaves a
// spent key in another spent.
func TestKeySpacesDoNotShareBuckets(t *testing.T) {
	const spent = "192.0.2.1"
	tests := []struct {
		name  string
		count int
		spend func(l *Limiter) error
		flood func(l *Limiter, i int)
	}{
		{
			name:  "consumption survives a link request flood",
			count: registeredDefaults.ConsumptionAttemptsPerClient,
			spend: func(l *Limiter) error { return l.AllowConsumption(limitEpoch, spent) },
			flood: func(l *Limiter, i int) {
				_ = l.allowLinkClient(limitEpoch, client(i))
				_ = l.allowLinkAddress(limitEpoch, address(i))
			},
		},
		{
			name:  "link requests per client survive a consumption flood",
			count: registeredDefaults.LinkRequestsPerClient,
			spend: func(l *Limiter) error { return l.allowLinkClient(limitEpoch, spent) },
			flood: func(l *Limiter, i int) { _ = l.AllowConsumption(limitEpoch, client(i)) },
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			for i := range tt.count {
				if err := tt.spend(l); err != nil {
					t.Fatalf("attempt %d of %d refused: %v", i+1, tt.count, err)
				}
			}
			if err := l.AllowConsumption(limitEpoch, "203.0.113.9"); err != nil {
				t.Fatalf("an unrelated client's consumption refused: %v", err)
			}
			for i := range maxLimiterKeys + 1 {
				tt.flood(l, i)
			}
			if err := tt.spend(l); !errors.Is(err, ErrRateLimited) {
				t.Errorf("spent key after the other space's flood: error = %v, want wrapping ErrRateLimited", err)
			}
		})
	}
}

// TestKeyedLimiterConcurrent charges one shared key and many distinct
// keys from many goroutines, under each at-capacity policy. Run under
// -race it proves the map is guarded; under any run it proves the
// shared key admits exactly its count and the bound holds. Each
// goroutine inserts one distinct key between its touches of the shared
// key, so at most goroutines keys arrive between two touches and a
// capacity above that never evicts the shared key.
func TestKeyedLimiterConcurrent(t *testing.T) {
	const (
		count      = 7
		capacity   = 16
		goroutines = 8
		perWorker  = 200
	)
	for _, policy := range []atCapacity{evictLeastRecent, evictRefilledOrAdmit} {
		t.Run(fmt.Sprint(policy), func(t *testing.T) {
			k := newKeyedLimiter(count, capacity, policy)
			var (
				wg       sync.WaitGroup
				mu       sync.Mutex
				admitted int
				largest  int
			)
			for g := range goroutines {
				wg.Go(func() {
					for i := range perWorker {
						ok := k.allow(limitEpoch, "shared")
						k.allow(limitEpoch, client(g*perWorker+i))
						n := k.size()
						mu.Lock()
						if ok {
							admitted++
						}
						largest = max(largest, n)
						mu.Unlock()
					}
				})
			}
			wg.Wait()
			if largest > capacity {
				t.Errorf("map held %d keys at its largest, bound is %d", largest, capacity)
			}
			if admitted != count {
				t.Errorf("shared key admitted %d times, want exactly %d", admitted, count)
			}
		})
	}
}

// TestKeyedLimiterRefilledPolicy fills a per-address-policy map to its
// bound at one instant with a count of two: key a spent twice, then b
// and c once each, so a is the least recently used key and the only
// one with no token left. A new key then arrives at each offset. No
// bucket has refilled at the same instant, so the new key is admitted
// untracked and nothing is evicted. Half an interval later b and c
// have refilled and a has not, so the least recently used refilled
// bucket, b, goes and a keeps its spent bucket. A whole interval later
// every bucket has refilled and the least recently used, a, goes.
func TestKeyedLimiterRefilledPolicy(t *testing.T) {
	tests := []struct {
		name     string
		after    time.Duration
		wantGone string
		wantKept []string
		tracked  bool
	}{
		{
			name:     "no refilled bucket admits the new key untracked",
			after:    0,
			wantKept: []string{"a", "b", "c"},
		},
		{
			name:     "the least recently used refilled bucket goes",
			after:    limitInterval / 2,
			wantGone: "b",
			wantKept: []string{"a", "c", "d"},
			tracked:  true,
		},
		{
			name:     "once all have refilled the least recently used goes",
			after:    limitInterval,
			wantGone: "a",
			wantKept: []string{"b", "c", "d"},
			tracked:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := newKeyedLimiter(2, 3, evictRefilledOrAdmit)
			for _, key := range []string{"a", "a", "b", "c"} {
				if !k.allow(limitEpoch, key) {
					t.Fatalf("filling charge on %q refused", key)
				}
			}
			now := limitEpoch.Add(tt.after)
			if !k.allow(now, "d") {
				t.Fatal("new key at capacity refused, want admitted")
			}
			if n := k.size(); n != 3 {
				t.Errorf("map holds %d keys, want its capacity 3", n)
			}
			if _, ok := k.byKey["d"]; ok != tt.tracked {
				t.Errorf("new key tracked = %v, want %v", ok, tt.tracked)
			}
			if _, ok := k.byKey[tt.wantGone]; tt.wantGone != "" && ok {
				t.Errorf("key %q kept, want it evicted", tt.wantGone)
			}
			for _, key := range tt.wantKept {
				if _, ok := k.byKey[key]; !ok {
					t.Errorf("key %q evicted, want it kept", key)
				}
			}
		})
	}
}

// TestClientKey pins the key each shape of client address is limited
// under. The IPv6 rows key on the /64 (RFC 4291's interface identifier
// is the low 64 bits), and a zone names a local interface, not a
// network.
func TestClientKey(t *testing.T) {
	tests := []struct {
		name   string
		client string
		want   string
	}{
		{"IPv4", "192.0.2.7", "192.0.2.7"},
		{"IPv4-mapped IPv6", "::ffff:192.0.2.7", "192.0.2.7"},
		{"IPv6", "2001:db8:1:2:3:4:5:6", "2001:db8:1:2::/64"},
		{"IPv6 link-local with a zone", "fe80::1%eth0", "fe80::/64"},
		{"empty", "", unparsedClientKey},
		{"not an address", "not an address", unparsedClientKey},
		{"address with a port", "192.0.2.7:40000", unparsedClientKey},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clientKey(tt.client); got != tt.want {
				t.Errorf("clientKey(%q) = %q, want %q", tt.client, got, tt.want)
			}
		})
	}
}

// TestClientLimitsKeyIPv6OnItsSlash64 walks 100 addresses inside one
// /64 against each per-client limit set to one. Any IPv6 host holds at
// least a /64, so keying on the full address would hand it 2^64
// budgets. The IPv4-mapped row presents one address in its two
// spellings, which must share a bucket.
func TestClientLimitsKeyIPv6OnItsSlash64(t *testing.T) {
	one := Limits{LinkRequestsPerAddress: 1, LinkRequestsPerClient: 1, ConsumptionAttemptsPerClient: 1, LinkMailCeiling: 1}
	inOne64 := func(i int) string { return fmt.Sprintf("2001:db8:1:2::%x", i+1) }
	mapped := func(i int) string { return []string{"192.0.2.7", "::ffff:192.0.2.7"}[i%2] }
	tests := []struct {
		name   string
		client func(i int) string
		allow  func(l *Limiter, client string) error
	}{
		{"consumption, one /64", inOne64, func(l *Limiter, c string) error { return l.AllowConsumption(limitEpoch, c) }},
		{"link requests, one /64", inOne64, func(l *Limiter, c string) error { return l.allowLinkClient(limitEpoch, c) }},
		{"consumption, IPv4 and its mapped form", mapped, func(l *Limiter, c string) error { return l.AllowConsumption(limitEpoch, c) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(one)
			var admitted int
			for i := range 100 {
				if tt.allow(l, tt.client(i)) == nil {
					admitted++
				}
			}
			if admitted != 1 {
				t.Errorf("100 requests from one client key admitted %d times, want 1", admitted)
			}
		})
	}
}
