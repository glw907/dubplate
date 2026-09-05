package musicbrainz

import (
	"context"
	"errors"
	"slices"
	"sync"
	"time"
)

// Class ranks one caller's claim on the shared MusicBrainz wire budget.
// Three classes are enough for every consumer the specs name: an
// uploader probe holds a contributor on the page through its lookups
// and must outrank a batch import (organizer design, "a user-present
// probe outranks a batch import"); interactive work must not starve
// timer imports (build spec, the same section); and rung 4's shadow
// spends deferred background budget through this limiter and never
// inline with the probe. Rung 9 re-verifies the policy against its real
// callers.
type Class int

const (
	ClassUserPresent Class = iota // a person is waiting on this request
	ClassBatch                    // an operator- or timer-triggered import
	ClassBackground               // deferred work: rung 4's shadow, and later
)

const numClasses = int(ClassBackground) + 1

// reservationInterval is the anti-starvation floor: every
// reservationInterval-th served token is reserved for the non-empty
// classes BELOW the one strict priority would have served, taken in
// round-robin so consecutive reservations rotate through them. Strict
// priority alone satisfies the precedence requirement and violates the
// anti-starvation one; sending every reservation to the LOWEST non-empty
// class instead satisfies neither, because a saturated ClassUserPresent
// queue plus a queued ClassBackground waiter starves ClassBatch
// outright, which is the timer import the build spec's constraint
// protects; and a weighted fair queue makes a probe wait its weighted
// turn behind a batch, which is the latency the uploader journey cannot
// absorb. This one integer is the whole difference, and it is what rung
// 9 re-verifies rather than a whole policy.
const reservationInterval = 4

// ErrBudgetExhausted is returned by [Budget.Wait] when the budget's own
// limit is spent. A caller branches on it to stop asking rather than to
// retry, which is why it is a sentinel.
var ErrBudgetExhausted = errors.New("caller budget exhausted")

// ErrLimiterClosed is returned to every waiter when the limiter shuts
// down. A caller branches on it to distinguish process shutdown from
// its own budget running out.
var ErrLimiterClosed = errors.New("limiter closed")

// DefaultInterval is the minimum spacing between two MusicBrainz
// requests, one second. beets' musicbrainz.org branch pins rate_limit =
// 1.0 (beetsplug/_utils/musicbrainz.py:597-599) and builds its session
// as LimiterTimeoutSession(per_second=1.0) with burst 1, which binds
// before the 0.25s RateLimitAdapter floor (contract C20). The repo's
// own live spender holds the same line
// (cmd/harness/capture.go:27-31).
const DefaultInterval = time.Second

// Limiter paces every MusicBrainz request one process issues, across
// every caller. It owns a goroutine, which is the one place in this
// package that is justified: the limiter's entire job is to make
// concurrent callers wait in a defined order, and rung 7's probe,
// rung 9's import and rung 4's shadow are three real concurrent
// callers.
//
// It reads no clock. Pacing comes from a ticker it owns, and a token is
// "the next tick", never "now minus the last request". That keeps
// internal/harness/conform's no-time.Now invariant literal rather than
// exempted, and it makes the tests deterministic: they drive the tick
// channel by hand.
type Limiter struct {
	done chan struct{}
	stop func() // stops the ticker NewLimiter owns; nil for an injected tick channel

	mu      sync.Mutex
	ready   *sync.Cond // wakes awaitQueued when a waiter registers
	closed  bool
	queues  [numClasses][]*waiter
	spare   bool   // one unclaimed token, held so an idle limiter delays nobody
	served  int    // tokens handed to a waiter, which is what a reservation counts
	cursor  Class  // where the next reserved token resumes its round-robin
	queued  int    // cumulative registrations, for awaitQueued
	handled int    // cumulative ticks the dispatcher has finished, for awaitTicks
	hold    func() // the offer seam holdBeforeOffer installs; nil in production
}

// waiter is one blocked [Budget.Wait]. The dispatcher offers a token by
// sending on grant, which is unbuffered so an offer nobody receives never
// completes, and abandons the offer when the departing waiter closes
// gone.
type waiter struct {
	grant chan struct{}
	gone  chan struct{}
}

// NewLimiter returns a limiter pacing the wire at one request per
// interval and starts its dispatcher. Close it when the process shuts
// down. interval must be positive; DefaultInterval is the MusicBrainz
// value.
//
// It takes an interval rather than a requests-per-second integer
// because beets' own non-musicbrainz.org branch derives a float rate
// (ratelimit / ratelimit_interval, :600-606) and an integer forbids a
// deliberately conservative sub-1/s pace.
func NewLimiter(interval time.Duration) *Limiter {
	ticker := time.NewTicker(interval)
	l := newLimiter(ticker.C)
	l.stop = ticker.Stop
	return l
}

// newLimiter is the first of the four unexported test seams, which
// together let a test drive the dispatcher without reading a clock:
// newLimiter injects the tick channel, [Limiter.awaitQueued] observes a
// waiter having registered, [Limiter.awaitTicks] observes a tick having
// been handled, and [Limiter.holdBeforeOffer] stops the dispatcher
// between picking a waiter and offering it the token.
func newLimiter(ticks <-chan time.Time) *Limiter {
	l := &Limiter{done: make(chan struct{})}
	l.ready = sync.NewCond(&l.mu)
	go l.dispatch(ticks)
	return l
}

// Budget returns a named per-caller budget in class, spending at most
// limit requests; a limit of zero or less means no cap, and the budget
// is still paced and still counts what it spends. The name is what makes
// spend attributable in a log and is not otherwise interpreted.
//
// This is the "per-caller budget interface" the rung 3 spec names: a
// seam, not a Go interface. go-conventions holds an interface back until
// a second implementation exists, and one concrete type with an
// injected wait function serves both the metered and the unmetered case.
func (l *Limiter) Budget(name string, class Class, limit int) *Budget {
	return &Budget{
		name:  name,
		limit: limit,
		wait:  func(ctx context.Context) error { return l.acquire(ctx, class) },
	}
}

// Close stops the dispatcher and fails every pending and future waiter
// with [ErrLimiterClosed]. It is idempotent.
func (l *Limiter) Close() error {
	l.mu.Lock()
	if l.closed {
		l.mu.Unlock()
		return nil
	}
	l.closed = true
	l.mu.Unlock()

	close(l.done)
	if l.stop != nil {
		l.stop()
	}
	return nil
}

// awaitQueued blocks until n waiters have registered since the limiter
// was built. A test that fed a tick before its waiter registered would
// lose that tick to the spare slot, so registration is observed on the
// dispatcher's own state rather than on a signal a waiter sends before
// it enqueues.
func (l *Limiter) awaitQueued(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.queued < n {
		l.ready.Wait()
	}
}

// awaitTicks blocks until the dispatcher has finished handling n ticks.
// A tick send returns as soon as the dispatcher receives it, which is
// before the token has become a grant or a spare, so a test that fed a
// tick and then acted on its effect would be racing the dispatcher.
func (l *Limiter) awaitTicks(n int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for l.handled < n {
		l.ready.Wait()
	}
}

// holdBeforeOffer installs the offer seam: f runs after serve has picked
// a waiter and before it offers that waiter the token. That window is the
// only place a test can cancel a waiter whose token is already in flight,
// which is the refusal the dispatcher's own recovery path exists for; a
// waiter cancelled before its tick has already left its queue and is
// never picked. Production leaves it nil and serve pays one nil compare.
func (l *Limiter) holdBeforeOffer(f func()) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.hold = f
}

func (l *Limiter) dispatch(ticks <-chan time.Time) {
	for {
		select {
		case <-l.done:
			return
		case <-ticks:
			l.serve()
			l.mu.Lock()
			l.handled++
			l.ready.Broadcast()
			l.mu.Unlock()
		}
	}
}

// serve hands one token to one waiter. A waiter that departs between
// being picked and being offered the token does not consume it: the
// offer is abandoned and the next waiter is tried, and only a token a
// waiter actually took advances the reservation counter.
func (l *Limiter) serve() {
	for {
		l.mu.Lock()
		w, class, reserved := l.pick()
		if w == nil {
			l.spare = true
			l.mu.Unlock()
			return
		}
		hold := l.hold
		l.mu.Unlock()

		if hold != nil {
			hold()
		}

		select {
		case w.grant <- struct{}{}:
			l.mu.Lock()
			l.served++
			if reserved {
				l.cursor = (class + 1) % Class(numClasses)
			}
			l.mu.Unlock()
			return
		case <-w.gone:
		case <-l.done:
			return
		}
	}
}

// pick removes the waiter this token belongs to and reports its class
// and whether the reservation floor chose it. It reports a nil waiter
// when every queue is empty. l.mu must be held.
func (l *Limiter) pick() (w *waiter, class Class, reserved bool) {
	top, ok := l.highest()
	if !ok {
		return nil, 0, false
	}
	class = top
	if (l.served+1)%reservationInterval == 0 {
		if below, ok := l.nextBelow(top); ok {
			class, reserved = below, true
		}
	}
	w = l.queues[class][0]
	l.queues[class] = l.queues[class][1:]
	return w, class, reserved
}

func (l *Limiter) highest() (Class, bool) {
	for c := range numClasses {
		if len(l.queues[c]) > 0 {
			return Class(c), true
		}
	}
	return 0, false
}

// nextBelow returns the non-empty class the reservation cursor reaches
// first among the classes ranked strictly below top.
func (l *Limiter) nextBelow(top Class) (Class, bool) {
	for i := range numClasses {
		c := Class((int(l.cursor) + i) % numClasses)
		if c > top && len(l.queues[c]) > 0 {
			return c, true
		}
	}
	return 0, false
}

func (l *Limiter) acquire(ctx context.Context, class Class) error {
	w := &waiter{grant: make(chan struct{}), gone: make(chan struct{})}

	l.mu.Lock()
	switch {
	case l.closed:
		l.mu.Unlock()
		return ErrLimiterClosed
	case l.spare:
		l.spare = false
		l.mu.Unlock()
		return nil
	}
	l.queues[class] = append(l.queues[class], w)
	l.queued++
	l.ready.Broadcast()
	l.mu.Unlock()

	select {
	case <-w.grant:
		return nil
	case <-ctx.Done():
		l.depart(w, class)
		return ctx.Err()
	case <-l.done:
		l.depart(w, class)
		return ErrLimiterClosed
	}
}

// depart tells a dispatcher mid-offer that the token was not taken and
// drops the waiter from its queue if it was never picked.
func (l *Limiter) depart(w *waiter, class Class) {
	close(w.gone)

	l.mu.Lock()
	defer l.mu.Unlock()
	if i := slices.Index(l.queues[class], w); i >= 0 {
		l.queues[class] = slices.Delete(l.queues[class], i, i+1)
	}
}

// Budget is one named caller's claim on a [Limiter], or the unmetered
// stand-in [Unmetered] returns.
type Budget struct {
	name  string
	limit int
	wait  func(context.Context) error

	mu    sync.Mutex
	spent int
}

// Unmetered returns a budget that never waits and never exhausts, for
// offline runs. internal/harness/conform uses it: the corpus holds 351
// committed requests and pacing them at one per second would add roughly
// six minutes to a gate that runs in 58 to 62 seconds today. It is a
// named constructor rather than a nil-limiter check, because a nil check
// at every call site is the defensive-nil pattern go-conventions
// rejects; the injected wait function is what makes the two cases one
// type with no branch.
func Unmetered() *Budget {
	return &Budget{
		name: "unmetered",
		wait: func(ctx context.Context) error { return ctx.Err() },
	}
}

// Wait blocks until the budget may issue one request, returning
// ctx.Err() on cancellation, [ErrBudgetExhausted] when the limit is
// spent, and [ErrLimiterClosed] when the limiter shut down.
//
// It reserves its spend before it waits and releases the reservation on
// any error path out of the wait, so the limit caps concurrent callers
// as well as sequential ones. Checking spent against limit and only
// then waiting is a lost update: two goroutines sharing a budget with
// limit 1 both read spent == 0, both queue, and both are granted, and
// no -race run reports it because there is no data race, only a wrong
// decision.
func (b *Budget) Wait(ctx context.Context) error {
	if err := b.reserve(); err != nil {
		return err
	}
	if err := b.wait(ctx); err != nil {
		b.release()
		return err
	}
	return nil
}

func (b *Budget) reserve() error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.limit > 0 && b.spent+1 > b.limit {
		return ErrBudgetExhausted
	}
	b.spent++
	return nil
}

func (b *Budget) release() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spent--
}

// Spent reports how many requests this budget has been granted.
func (b *Budget) Spent() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.spent
}

// Name reports the budget's name, for the caller that logs its own
// spend.
func (b *Budget) Name() string { return b.name }
