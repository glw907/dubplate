package auth

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/glw907/dubplate/internal/identity"
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
// is one key of at most [identity.MaxAddressLength] bytes, an 80-byte
// bucket, and two ring pointers, under half a kilobyte in all, so three
// full maps hold about fifteen megabytes however many distinct keys a
// flood presents. Ten thousand is two orders of magnitude above the
// addresses and client addresses a group's own members present in an
// hour, so honest traffic never reaches the bound and only a flood
// drives eviction.
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
// policy. Limiter derives no client address of its own.
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
		perAddress:  newKeyedLimiter(limits.LinkRequestsPerAddress, maxLimiterKeys),
		perClient:   newKeyedLimiter(limits.LinkRequestsPerClient, maxLimiterKeys),
		consumption: newKeyedLimiter(limits.ConsumptionAttemptsPerClient, maxLimiterKeys),
		linkMail:    newBucket(limits.LinkMailCeiling),
	}
}

// AllowLinkRequest charges one link request against address and against
// client, and returns an error wrapping [ErrRateLimited] when either
// limit is spent. Both are charged even when one refuses, so a client
// cycling through addresses spends its own budget on every attempt.
// address is normalized first, so two spellings of one address share a
// bucket; an address that does not normalize returns an error wrapping
// [identity.ErrInvalidAddress] and charges nothing.
func (l *Limiter) AllowLinkRequest(now time.Time, address, client string) error {
	key, err := identity.NormalizeAddress(address)
	if err != nil {
		// The normalizer's error can quote its input, and no error this
		// package returns names an address.
		return fmt.Errorf("rate limit key: %w", identity.ErrInvalidAddress)
	}
	addressOK := l.perAddress.allow(now, key)
	clientOK := l.perClient.allow(now, client)
	switch {
	case !addressOK:
		return fmt.Errorf("%w: link requests per address", ErrRateLimited)
	case !clientOK:
		return fmt.Errorf("%w: link requests per client", ErrRateLimited)
	}
	return nil
}

// AllowConsumption charges one link consumption attempt against client
// and returns an error wrapping [ErrRateLimited] when its limit is
// spent. The caller runs it before verifying the link, so a flood of
// forged tokens is bounded here rather than reaching the verifier.
func (l *Limiter) AllowConsumption(now time.Time, client string) error {
	if !l.consumption.allow(now, client) {
		return fmt.Errorf("%w: link consumption attempts per client", ErrRateLimited)
	}
	return nil
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

// keyedLimiter is a map from key to bucket holding at most capacity keys.
// Admitting a key past the bound evicts the least recently used one:
// a bucket nobody has charged recently is the one closest to full, and
// a bucket untouched for a whole [limitInterval] has refilled
// completely, so losing it and recreating it full later changes
// nothing. Evicting a recent key instead would hand whoever charged it
// a fresh bucket, which is exactly the reset a flood of distinct keys
// would be sent to buy.
//
// Recency is an intrusive doubly linked ring through the buckets
// themselves, typed so no lookup needs a type assertion.
type keyedLimiter struct {
	perInterval int
	capacity    int

	mu    sync.Mutex
	byKey map[string]*keyedBucket
	ring  keyedBucket // sentinel: ring.next is most recently used, ring.prev least
}

type keyedBucket struct {
	key        string
	bucket     *rate.Limiter
	prev, next *keyedBucket
}

func newKeyedLimiter(perInterval, capacity int) *keyedLimiter {
	k := &keyedLimiter{
		perInterval: perInterval,
		capacity:    capacity,
		byKey:       make(map[string]*keyedBucket),
	}
	k.ring.prev, k.ring.next = &k.ring, &k.ring
	return k
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
		oldest := k.ring.prev
		oldest.unlink()
		delete(k.byKey, oldest.key)
		fallthrough
	default:
		kb = &keyedBucket{key: key, bucket: newBucket(k.perInterval)}
		k.byKey[key] = kb
	}
	k.pushFront(kb)
	return kb.bucket.AllowN(now, 1)
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
