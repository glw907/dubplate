package auth

import (
	"errors"
	"fmt"
	"net/netip"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// ErrRateLimited reports that a login request or a link consumption was
// refused because one of [Limiter]'s limits is spent. The wrapped text
// names which limit tripped, never the key it tripped on; the caller
// logs the refusal with the client address it already holds.
var ErrRateLimited = errors.New("auth: rate limit tripped")

// limitInterval is the window every [Limits] count is stated over: a
// count of n admits n events per interval, refilling one event every
// interval/n.
const limitInterval = time.Hour

// maxLimiterKeys bounds each of [Limiter]'s three keyed maps. An entry
// is one key of at most 254 bytes (a normalized address; a client key
// is shorter), an 80-byte bucket, and two ring pointers, under half a
// kilobyte in all, so three full maps hold about fifteen megabytes
// however many distinct keys a flood presents. Ten thousand is two
// orders of magnitude above the addresses and client networks a
// group's own members present in an hour, so honest traffic never
// reaches the bound. The per-address map holds only roster addresses,
// so it reaches the bound only on a roster that large.
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
// submitted address, link requests per client address, and link
// consumption attempts per client address, each over its own key space,
// plus one global ceiling on outbound link mail. A Limiter is safe for
// concurrent use and logs nothing.
//
// Every method takes the caller's clock reading as now, and a client
// address the caller has already resolved through its trusted-proxy
// policy. Limiter derives no client address of its own, but it keys a
// client on its network rather than its address: an IPv4 address,
// or an IPv4-mapped IPv6 one, keys on itself, and an IPv6 address keys
// on its /64, since any IPv6 host holds at least a /64 and would
// otherwise present 2^64 clients. A value that does not parse as an
// address shares one bucket with every other such value.
type Limiter struct {
	perAddress  *keyedLimiter
	perClient   *keyedLimiter
	consumption *keyedLimiter
	linkMail    *rate.Limiter
}

// NewLimiter returns a Limiter enforcing limits, each bucket starting
// full.
func NewLimiter(limits Limits) *Limiter {
	return &Limiter{
		perAddress:  newKeyedLimiter(limits.LinkRequestsPerAddress, maxLimiterKeys, evictRefilledOrAdmit),
		perClient:   newKeyedLimiter(limits.LinkRequestsPerClient, maxLimiterKeys, evictLeastRecent),
		consumption: newKeyedLimiter(limits.ConsumptionAttemptsPerClient, maxLimiterKeys, evictLeastRecent),
		linkMail:    newBucket(limits.LinkMailCeiling),
	}
}

// allowLinkClient charges one link request against client, and returns
// an error wrapping [ErrRateLimited] when its limit is spent. The
// caller charges it first, for every request, before it reads the
// address: an address that fails to normalize still spends the
// client's budget, and a refused client never reaches the per-address
// map.
func (l *Limiter) allowLinkClient(now time.Time, client string) error {
	return l.perClient.chargeOrRefuse(now, clientKey(client), "link requests per client")
}

// allowLinkAddress charges one link request against address, which the
// caller has normalized and found on the roster, and returns an error
// wrapping [ErrRateLimited] when its limit is spent. The caller never
// charges an address the roster does not hold, so the per-address map
// holds no more keys than the roster has members and a flood of
// made-up addresses never reaches it.
func (l *Limiter) allowLinkAddress(now time.Time, address string) error {
	return l.perAddress.chargeOrRefuse(now, address, "link requests per address")
}

// AllowConsumption charges one link consumption attempt against client
// and returns an error wrapping [ErrRateLimited] when its limit is
// spent. The caller runs it before verifying the link, so a flood of
// forged tokens is bounded here rather than reaching the verifier.
func (l *Limiter) AllowConsumption(now time.Time, client string) error {
	return l.consumption.chargeOrRefuse(now, clientKey(client), "link consumption attempts per client")
}

// AllowLinkMail charges one outbound link mail against the global
// ceiling and returns an error wrapping [ErrRateLimited] when it is
// spent. The ceiling is shared by every key, so an enumerator walking
// distinct addresses from distinct clients still cannot send more mail
// than it admits.
func (l *Limiter) AllowLinkMail(now time.Time) error {
	if !l.linkMail.AllowN(now, 1) {
		return fmt.Errorf("%w: outbound link mail ceiling", ErrRateLimited)
	}
	return nil
}

// newBucket returns a bucket admitting perInterval events per
// [limitInterval], starting full.
func newBucket(perInterval int) *rate.Limiter {
	return rate.NewLimiter(rate.Every(limitInterval/time.Duration(perInterval)), perInterval)
}

// unparsedClientKey is the one key every client value that does not
// parse as an address shares, so malformed input cannot mint distinct
// keys. No parsed address or prefix formats as this string.
const unparsedClientKey = "unparsed"

// clientKey returns the key client is limited under: an IPv4 address,
// unmapped from IPv6 when it arrives wrapped, keys on itself, and an
// IPv6 address keys on its /64, zone dropped.
func clientKey(client string) string {
	addr, err := netip.ParseAddr(client)
	if err != nil {
		return unparsedClientKey
	}
	addr = addr.Unmap()
	if addr.Is4() {
		return addr.String()
	}
	// Prefix refuses only a length outside 0 through 128, and 64 is
	// inside it for every IPv6 address.
	prefix, _ := addr.Prefix(64)
	return prefix.String()
}

// atCapacity is what a keyedLimiter does when a new key arrives and it
// already holds its capacity.
type atCapacity int

const (
	// evictLeastRecent evicts the least recently used bucket. A bucket
	// nobody has charged recently is usually the one closest to full,
	// and evicting a recent key instead would hand whoever charged it
	// a fresh bucket, which is exactly the reset a flood of distinct
	// keys would be sent to buy. Under a flood, though, a bucket
	// charged seconds ago can be the least recently used one, so this
	// policy suits only a key space where a reset hands the flooder
	// nothing it does not already hold, such as its own client key.
	evictLeastRecent atCapacity = iota

	// evictRefilledOrAdmit evicts the least recently used bucket that
	// has fully refilled, and when none has, admits the event without
	// tracking the new key, so no spent bucket is ever reset. The scan
	// starts at the least recently used end and stops at the first
	// refilled bucket; it cannot stop at the least recently used bucket
	// alone, because refill time depends on the tokens spent, not on
	// recency. At worst it reads every bucket, [maxLimiterKeys] of
	// them, under the map's lock.
	evictRefilledOrAdmit
)

// keyedLimiter is a map from key to bucket holding at most capacity
// keys, with policy deciding what a new key does once it is full.
//
// Recency is an intrusive doubly linked ring through the buckets
// themselves, typed so no lookup needs a type assertion.
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

// chargeOrRefuse charges one event against key at now and returns an
// error wrapping [ErrRateLimited] naming reason when the limit is
// spent.
func (k *keyedLimiter) chargeOrRefuse(now time.Time, key, reason string) error {
	if !k.allow(now, key) {
		return fmt.Errorf("%w: %s", ErrRateLimited, reason)
	}
	return nil
}

// allow charges one event against key's bucket at now, creating the
// bucket full if key is absent, and reports whether it was admitted.
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

// evictable returns the bucket policy evicts at now, or nil when it
// evicts none.
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

// size returns how many keys the map holds.
func (k *keyedLimiter) size() int {
	k.mu.Lock()
	defer k.mu.Unlock()
	return len(k.byKey)
}
