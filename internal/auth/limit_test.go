package auth

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
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
func client(i int) string  { return fmt.Sprintf("198.51.100.%d:%d", i%256, i) }

// TestLinkRequestKeysTripIndependently proves each of the two link keys
// refuses on its own: a request passes only when both admit it.
func TestLinkRequestKeysTripIndependently(t *testing.T) {
	tests := []struct {
		name     string
		admitted int
		request  func(i int) (address, client string)
		wantText string
	}{
		{
			name:     "one address from many clients trips the per-address limit",
			admitted: registeredDefaults.LinkRequestsPerAddress,
			request:  func(i int) (string, string) { return "member@example.com", client(i) },
			wantText: "link requests per address",
		},
		{
			name:     "many addresses from one client trips the per-client limit",
			admitted: registeredDefaults.LinkRequestsPerClient,
			request:  func(i int) (string, string) { return address(i), "192.0.2.7" },
			wantText: "link requests per client",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			for i := range tt.admitted {
				a, c := tt.request(i)
				if err := l.AllowLinkRequest(limitEpoch, a, c); err != nil {
					t.Fatalf("request %d of %d refused: %v", i+1, tt.admitted, err)
				}
			}
			a, c := tt.request(tt.admitted)
			err := l.AllowLinkRequest(limitEpoch, a, c)
			if !errors.Is(err, ErrRateLimited) {
				t.Fatalf("request %d error = %v, want wrapping ErrRateLimited", tt.admitted+1, err)
			}
			if got := err.Error(); got != ErrRateLimited.Error()+": "+tt.wantText {
				t.Errorf("error = %q, want the %q limit named", got, tt.wantText)
			}
			if err := l.AllowLinkRequest(limitEpoch, "stranger@example.com", "203.0.113.9"); err != nil {
				t.Errorf("fresh address from fresh client refused: %v", err)
			}
		})
	}
}

// TestLinkRequestNormalizesAddress proves the per-address key is the
// normalized address, and that an address the normalizer refuses is
// refused before any bucket is charged.
func TestLinkRequestNormalizesAddress(t *testing.T) {
	l := NewLimiter(registeredDefaults)
	spellings := []string{"Member@Example.COM", "member@example.com", "<MEMBER@example.com>"}
	for i := range registeredDefaults.LinkRequestsPerAddress {
		if err := l.AllowLinkRequest(limitEpoch, spellings[i%len(spellings)], client(i)); err != nil {
			t.Fatalf("request %d refused: %v", i+1, err)
		}
	}
	if err := l.AllowLinkRequest(limitEpoch, "MEMBER@EXAMPLE.COM", client(99)); !errors.Is(err, ErrRateLimited) {
		t.Errorf("another spelling of a spent address: error = %v, want wrapping ErrRateLimited", err)
	}

	l = NewLimiter(Limits{LinkRequestsPerAddress: 1, LinkRequestsPerClient: 1, ConsumptionAttemptsPerClient: 1, LinkMailCeiling: 1})
	err := l.AllowLinkRequest(limitEpoch, "not an address", "192.0.2.1")
	if !errors.Is(err, identity.ErrInvalidAddress) {
		t.Fatalf("unnormalizable address: error = %v, want wrapping identity.ErrInvalidAddress", err)
	}
	if err := l.AllowLinkRequest(limitEpoch, "member@example.com", "192.0.2.1"); err != nil {
		t.Errorf("refused address charged the client's only request: %v", err)
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
		if err := l.AllowLinkRequest(limitEpoch, address(i), client(i)); err != nil {
			t.Fatalf("request %d with fresh keys refused: %v", i+1, err)
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
// end still fails.
func TestKeyedLimiterBoundHoldsUnderFlood(t *testing.T) {
	tests := []struct {
		name  string
		drive func(l *Limiter, i int)
		space func(l *Limiter) *keyedLimiter
	}{
		{"link requests per address", func(l *Limiter, i int) {
			_ = l.AllowLinkRequest(limitEpoch, address(i), "192.0.2.1")
		}, func(l *Limiter) *keyedLimiter { return l.perAddress }},
		{"link requests per client", func(l *Limiter, i int) {
			_ = l.AllowLinkRequest(limitEpoch, "member@example.com", client(i))
		}, func(l *Limiter) *keyedLimiter { return l.perClient }},
		{"consumption attempts per client", func(l *Limiter, i int) {
			_ = l.AllowConsumption(limitEpoch, client(i))
		}, func(l *Limiter) *keyedLimiter { return l.consumption }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := NewLimiter(registeredDefaults)
			space := tt.space(l)
			for i := range 2*maxLimiterKeys + 1 {
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
	k := newKeyedLimiter(1, 3)
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
	var fresh int
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
			flood: func(l *Limiter, i int) { _ = l.AllowLinkRequest(limitEpoch, address(i), client(i)) },
		},
		{
			name:  "link requests per client survive a consumption flood",
			count: registeredDefaults.LinkRequestsPerClient,
			spend: func(l *Limiter) error {
				fresh++
				return l.AllowLinkRequest(limitEpoch, address(-fresh), spent)
			},
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
// keys from many goroutines. Run under -race it proves the map is
// guarded; under any run it proves the shared key admits exactly its
// count and the bound holds. Each goroutine inserts one distinct key
// between its touches of the shared key, so at most goroutines keys
// arrive between two touches and a capacity above that never evicts
// the shared key.
func TestKeyedLimiterConcurrent(t *testing.T) {
	const (
		count      = 7
		capacity   = 16
		goroutines = 8
		perWorker  = 200
	)
	k := newKeyedLimiter(count, capacity)
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
}
