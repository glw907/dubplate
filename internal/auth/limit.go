package auth

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// errRateLimited is wrapped by every [Limiter] refusal. The wrapped text
// names which limit tripped, never the key it tripped on.
var errRateLimited = errors.New("rate limit tripped")

// limitInterval is the window every [Limits] count is stated over: a
// count of n admits n events per interval, refilling one every interval/n.
const limitInterval = time.Hour

// maxLimiterKeys bounds each keyed map. An entry is under half a
// kilobyte, so three full maps hold about fifteen megabytes however many
// keys a flood presents, and the bound sits two orders of magnitude
// above what a group's own members present in an hour.
const maxLimiterKeys = 10_000

// Limits is the four per-hour counts a [Limiter] enforces, resolved
// from configuration by the caller. Each must be positive.
type Limits struct {
	LinkRequestsPerAddress       int
	LinkRequestsPerClient        int
	ConsumptionAttemptsPerClient int
	LinkMailCeiling              int
}

// A Limiter holds the auth surface's rate limits: link requests per
// submitted address, link requests per client, and link consumption
// attempts per client, plus one global ceiling on outbound link mail. A
// Limiter is safe for concurrent use and logs nothing. Callers pass a
// client address already resolved through the trusted-proxy policy.
type Limiter struct {
	perAddress  *keyedLimiter
	perClient   *keyedLimiter
	consumption *keyedLimiter
	linkMail    *rate.Limiter
}

// NewLimiter returns a Limiter enforcing limits, each bucket full.
func NewLimiter(limits Limits) *Limiter {
	return &Limiter{
		perAddress:  newKeyedLimiter(limits.LinkRequestsPerAddress, maxLimiterKeys, evictRefilledOrAdmit),
		perClient:   newKeyedLimiter(limits.LinkRequestsPerClient, maxLimiterKeys, evictLeastRecent),
		consumption: newKeyedLimiter(limits.ConsumptionAttemptsPerClient, maxLimiterKeys, evictLeastRecent),
		linkMail:    newBucket(limits.LinkMailCeiling),
	}
}

// allowLinkClient is charged first, before the address is read, so an
// address that fails to normalize still spends the client's budget.
func (l *Limiter) allowLinkClient(now time.Time, client string) error {
	return l.perClient.chargeOrRefuse(now, clientKey(client), "link requests per client")
}

// allowLinkAddress is charged only for a roster address, so a flood of
// made-up addresses never reaches the per-address map.
func (l *Limiter) allowLinkAddress(now time.Time, address string) error {
	return l.perAddress.chargeOrRefuse(now, address, "link requests per address")
}

// allowConsumption runs before the link is verified, so forged tokens
// are bounded here rather than reaching the verifier.
func (l *Limiter) allowConsumption(now time.Time, client string) error {
	return l.consumption.chargeOrRefuse(now, clientKey(client), "link consumption attempts per client")
}

// allowLinkMail charges the global ceiling, so an enumerator walking
// distinct addresses from distinct clients still cannot exceed it.
func (l *Limiter) allowLinkMail(now time.Time) error {
	if !l.linkMail.AllowN(now, 1) {
		return fmt.Errorf("%w: outbound link mail ceiling", errRateLimited)
	}
	return nil
}

func newBucket(perInterval int) *rate.Limiter {
	return rate.NewLimiter(rate.Every(limitInterval/time.Duration(perInterval)), perInterval)
}

// unparsedClientKey is shared by every client value that does not parse,
// so malformed input cannot mint keys. No address or prefix formats as it.
const unparsedClientKey = "unparsed"

// clientKey returns the key client is limited under: an IPv4 address,
// unmapped when IPv6-wrapped, keys on itself, and an IPv6 address on its
// /64, since any IPv6 host holds at least a /64 and would otherwise
// present 2^64 clients.
func clientKey(client string) string {
	addr, err := netip.ParseAddr(client)
	if err != nil {
		return unparsedClientKey
	}
	addr = addr.Unmap()
	if addr.Is4() {
		return addr.String()
	}
	// Prefix fails only for a length outside 0 through 128.
	prefix, _ := addr.Prefix(64)
	return prefix.String()
}

// atCapacity is what a full keyedLimiter does with a new key.
type atCapacity int

const (
	// evictLeastRecent evicts the least recently used bucket. Under a
	// flood that bucket may be seconds old, so it suits only a key space
	// where a reset hands the flooder nothing it does not already hold,
	// such as its own client key.
	evictLeastRecent atCapacity = iota

	// evictRefilledOrAdmit evicts the least recently used bucket that has
	// fully refilled and, when none has, admits the event untracked, so
	// no spent bucket is reset. Refill depends on tokens spent, not
	// recency, so the scan may read every bucket under the lock.
	evictRefilledOrAdmit
)

// keyedLimiter maps key to bucket, holding at most capacity keys.
// Recency is an intrusive ring through the buckets.
type keyedLimiter struct {
	perInterval int
	capacity    int
	policy      atCapacity

	mu    sync.Mutex
	byKey map[string]*keyedBucket
	ring  keyedBucket // sentinel: ring.next is most recently used, ring.prev least
}

type keyedBucket struct {
	key        string
	bucket     *rate.Limiter
	prev, next *keyedBucket
}

func newKeyedLimiter(perInterval, capacity int, policy atCapacity) *keyedLimiter {
	k := &keyedLimiter{
		perInterval: perInterval,
		capacity:    capacity,
		policy:      policy,
		byKey:       make(map[string]*keyedBucket),
	}
	k.ring.prev, k.ring.next = &k.ring, &k.ring
	return k
}

func (k *keyedLimiter) chargeOrRefuse(now time.Time, key, reason string) error {
	if !k.allow(now, key) {
		return fmt.Errorf("%w: %s", errRateLimited, reason)
	}
	return nil
}

func (k *keyedLimiter) allow(now time.Time, key string) bool {
	k.mu.Lock()
	defer k.mu.Unlock()
	kb, ok := k.byKey[key]
	switch {
	case ok:
		kb.unlink()
	case len(k.byKey) >= k.capacity:
		victim := k.evictable(now)
		if victim == nil {
			// A new bucket starts full, so tracking the key would
			// have admitted this event too.
			return true
		}
		victim.unlink()
		delete(k.byKey, victim.key)
		fallthrough
	default:
		kb = &keyedBucket{key: key, bucket: newBucket(k.perInterval)}
		k.byKey[key] = kb
	}
	k.pushFront(kb)
	return kb.bucket.AllowN(now, 1)
}

// evictable returns nil when policy evicts no bucket.
func (k *keyedLimiter) evictable(now time.Time) *keyedBucket {
	if k.policy == evictLeastRecent {
		return k.ring.prev
	}
	for kb := k.ring.prev; kb != &k.ring; kb = kb.prev {
		if kb.bucket.TokensAt(now) >= float64(kb.bucket.Burst()) {
			return kb
		}
	}
	return nil
}

func (kb *keyedBucket) unlink() {
	kb.prev.next, kb.next.prev = kb.next, kb.prev
}

func (k *keyedLimiter) pushFront(kb *keyedBucket) {
	kb.prev, kb.next = &k.ring, k.ring.next
	k.ring.next.prev = kb
	k.ring.next = kb
}
