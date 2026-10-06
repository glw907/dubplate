package musicbrainz

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"
)

// closeLimiter is every test's deferred shutdown. Close reports an
// error the tests do not swallow.
func closeLimiter(t *testing.T, l *Limiter) {
	t.Helper()
	if err := l.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
}

// startWaiter runs one Wait and reports the label it waited under, so a
// test reads the order the dispatcher served in rather than a timing.
func startWaiter(b *Budget, label string, served chan<- string) {
	go func() {
		if err := b.Wait(context.Background()); err != nil {
			served <- "error: " + err.Error()
			return
		}
		served <- label
	}()
}

func TestDispatcherPrefersUserPresentAndStillFeedsBackground(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	probe := l.Budget(ClassUserPresent, 0)
	background := l.Budget(ClassBackground, 0)

	// served records the order Wait returned in, which is the property
	// under test.
	served := make(chan string, 16)

	startWaiter(background, "background", served)
	l.awaitQueued(1)
	for range 8 {
		startWaiter(probe, "probe", served)
	}
	l.awaitQueued(9)

	var order []string
	for range 8 {
		ticks <- time.Time{}
		order = append(order, <-served)
	}

	// Precedence: the first token goes to a probe queued after the
	// background waiter, which was queued first.
	if order[0] != "probe" {
		t.Errorf("first token went to %q, want the user-present waiter", order[0])
	}
	// Anti-starvation: reservationInterval is 4, so the background waiter is
	// served on the fourth token and not left behind eight probes.
	if !slices.Contains(order, "background") {
		t.Errorf("background waiter never served across 8 tokens: %v", order)
	}
	if order[reservationInterval-1] != "background" {
		t.Errorf("token %d went to %q, want the reserved background waiter",
			reservationInterval, order[reservationInterval-1])
	}
}

// TestReservedTokensRotateThroughTheClassesBelow is the three-class row.
// Two classes cannot see the starvation this fixture catches: with a
// saturated background stream, a reservation sent to the lowest non-empty
// class serves the background class every time and the batch import
// never runs.
func TestReservedTokensRotateThroughTheClassesBelow(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	probe := l.Budget(ClassUserPresent, 0)
	batch := l.Budget(ClassBatch, 0)
	background := l.Budget(ClassBackground, 0)

	served := make(chan string, 32)

	startWaiter(batch, "batch", served)
	for range 3 {
		startWaiter(background, "background", served)
	}
	l.awaitQueued(4)
	const tokens = 3 * reservationInterval
	for range tokens {
		startWaiter(probe, "probe", served)
	}
	l.awaitQueued(4 + tokens)

	var order []string
	for range tokens {
		ticks <- time.Time{}
		order = append(order, <-served)
	}

	if order[0] != "probe" {
		t.Errorf("first token went to %q, want the user-present waiter", order[0])
	}
	for _, label := range []string{"probe", "batch", "background"} {
		if !slices.Contains(order, label) {
			t.Errorf("%s never served across %d tokens: %v", label, tokens, order)
		}
	}
	first, second := order[reservationInterval-1], order[2*reservationInterval-1]
	if first == "probe" || second == "probe" {
		t.Errorf("reserved tokens went to %q and %q, want two lower classes", first, second)
	}
	if first == second {
		t.Errorf("both reserved tokens went to %q, want a round-robin across the lower classes", first)
	}
}

func TestIdleTickIsHeldAsASpare(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	ticks <- time.Time{}
	l.awaitTicks(1)

	b := l.Budget(ClassUserPresent, 0)
	if err := b.Wait(context.Background()); err != nil {
		t.Fatalf("Wait after an idle tick: %v", err)
	}
	if b.Spent() != 1 {
		t.Errorf("Spent() = %d, want 1", b.Spent())
	}
}

func TestSpareIsCappedAtOne(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	ticks <- time.Time{}
	ticks <- time.Time{}
	l.awaitTicks(2)

	probe := l.Budget(ClassUserPresent, 0)
	served := make(chan string, 2)
	startWaiter(probe, "first", served)
	startWaiter(probe, "second", served)

	if got := <-served; got != "first" && got != "second" {
		t.Fatalf("first waiter reported %q", got)
	}
	// Opportunistic: the second waiter may not have reported yet, so this
	// probe catches a second grant only sometimes. awaitQueued below is
	// the deterministic assertion, and it is what hangs under a mutation
	// that hands out two spares.
	select {
	case extra := <-served:
		t.Errorf("a second waiter was served (%q) off two idle ticks, want one spare only", extra)
	default:
	}
	// The other waiter is still queued, which two spares would have
	// spared it from.
	l.awaitQueued(1)
}

func TestCancelledWaiterKeepsItsTokenForTheNextWaiter(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	leaving := l.Budget(ClassUserPresent, 0)
	next := l.Budget(ClassUserPresent, 0)

	ctx, cancel := context.WithCancel(context.Background())
	left := make(chan error, 1)
	go func() { left <- leaving.Wait(ctx) }()
	l.awaitQueued(1)

	served := make(chan error, 1)
	go func() { served <- next.Wait(context.Background()) }()
	l.awaitQueued(2)

	cancel()
	if err := <-left; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Wait returned %v, want context.Canceled", err)
	}
	if leaving.Spent() != 0 {
		t.Errorf("cancelled budget Spent() = %d, want 0", leaving.Spent())
	}

	ticks <- time.Time{}
	if err := <-served; err != nil {
		t.Fatalf("waiter behind the cancelled one: %v", err)
	}
	if next.Spent() != 1 {
		t.Errorf("next.Spent() = %d, want 1", next.Spent())
	}
}

// TestInFlightTokenPassesToTheNextWaiter proves the stronger half of the
// cancellation contract: a waiter that departs while its token is already
// in flight does not consume it, and the waiter behind it takes that same
// token. The holdBeforeOffer seam is what makes it deterministic; without
// it the cancelled waiter is gone from its queue before the tick arrives
// and the dispatcher never picks it at all.
func TestInFlightTokenPassesToTheNextWaiter(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	picked, release := make(chan struct{}), make(chan struct{})
	var once sync.Once
	l.holdBeforeOffer(func() {
		once.Do(func() {
			close(picked)
			<-release
		})
	})

	leaving := l.Budget(ClassUserPresent, 0)
	next := l.Budget(ClassUserPresent, 0)

	ctx, cancel := context.WithCancel(context.Background())
	left := make(chan error, 1)
	go func() { left <- leaving.Wait(ctx) }()
	l.awaitQueued(1)

	served := make(chan error, 1)
	go func() { served <- next.Wait(context.Background()) }()
	l.awaitQueued(2)

	ticks <- time.Time{}
	<-picked

	cancel()
	if err := <-left; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled Wait returned %v, want context.Canceled", err)
	}
	close(release)

	// No second tick is fed, so the only token that can reach the next
	// waiter is the one the cancelled waiter refused.
	if err := <-served; err != nil {
		t.Fatalf("waiter behind the refused offer: %v", err)
	}
	if leaving.Spent() != 0 {
		t.Errorf("cancelled budget Spent() = %d, want 0", leaving.Spent())
	}
	if next.Spent() != 1 {
		t.Errorf("next.Spent() = %d, want 1", next.Spent())
	}
}

func TestBudgetCapsSequentialSpend(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	b := l.Budget(ClassUserPresent, 2)
	for i := range 2 {
		res := make(chan error, 1)
		go func() { res <- b.Wait(context.Background()) }()
		l.awaitQueued(i + 1)
		ticks <- time.Time{}
		if err := <-res; err != nil {
			t.Fatalf("wait %d: %v", i+1, err)
		}
	}

	// A spare is armed before the third wait so that a budget that
	// failed to count its spend proceeds and returns nil here, rather
	// than blocking on a tick that never comes.
	ticks <- time.Time{}
	l.awaitTicks(3)

	if err := b.Wait(context.Background()); !errors.Is(err, ErrBudgetExhausted) {
		t.Errorf("third wait on a limit of 2 returned %v, want ErrBudgetExhausted", err)
	}
	if b.Spent() != 2 {
		t.Errorf("Spent() = %d, want 2", b.Spent())
	}
}

func TestBudgetCapsConcurrentSpend(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)
	defer closeLimiter(t, l)

	b := l.Budget(ClassUserPresent, 2)
	results := make(chan error, 4)
	for range 4 {
		go func() { results <- b.Wait(context.Background()) }()
	}
	// Only the two waiters that reserved a spend ever queue; the other
	// two are refused before they reach the dispatcher.
	l.awaitQueued(2)
	for range 4 {
		ticks <- time.Time{}
	}

	var granted, exhausted int
	for range 4 {
		switch err := <-results; {
		case err == nil:
			granted++
		case errors.Is(err, ErrBudgetExhausted):
			exhausted++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if granted != 2 || exhausted != 2 {
		t.Errorf("four concurrent waits on a limit of 2 gave %d granted and %d exhausted, want 2 and 2",
			granted, exhausted)
	}
	if b.Spent() != 2 {
		t.Errorf("Spent() = %d, want 2", b.Spent())
	}
}

func TestUnmeteredNeverWaitsAndStillCounts(t *testing.T) {
	b := Unmetered()
	for i := range 3 {
		if err := b.Wait(context.Background()); err != nil {
			t.Fatalf("unmetered wait %d: %v", i+1, err)
		}
	}
	if b.Spent() != 3 {
		t.Errorf("Spent() = %d, want 3", b.Spent())
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Wait(ctx); !errors.Is(err, context.Canceled) {
		t.Errorf("unmetered wait on a cancelled context returned %v, want context.Canceled", err)
	}
	if b.Spent() != 3 {
		t.Errorf("Spent() = %d after a cancelled wait, want 3", b.Spent())
	}
}

func TestCloseFailsPendingAndFutureWaiters(t *testing.T) {
	l := newLimiter(make(chan time.Time))
	b := l.Budget(ClassUserPresent, 0)

	pending := make(chan error, 1)
	go func() { pending <- b.Wait(context.Background()) }()
	l.awaitQueued(1)

	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := <-pending; !errors.Is(err, ErrLimiterClosed) {
		t.Errorf("pending waiter got %v, want ErrLimiterClosed", err)
	}
	if err := b.Wait(context.Background()); !errors.Is(err, ErrLimiterClosed) {
		t.Errorf("wait after Close got %v, want ErrLimiterClosed", err)
	}
	if b.Spent() != 0 {
		t.Errorf("Spent() = %d after two failed waits, want 0", b.Spent())
	}
	if err := l.Close(); err != nil {
		t.Errorf("second Close: %v", err)
	}
}

// TestCloseAbandonsATokenAlreadyOffered proves serve's shutdown arm,
// the one path no other test reaches: a token in flight to a waiter
// that is not receiving it, and a Close while it hangs there. The
// waiter is fabricated and enqueued directly, because a real
// [Budget.Wait] is always receiving on its own grant channel and so can
// never leave the dispatcher blocked mid-offer.
//
// awaitTicks is the assertion. The dispatcher counts a tick handled
// only after serve returns, so a serve that could not abandon its offer
// never reaches the count. Mutation: delete serve's `case <-l.done:`
// arm and this test hangs until the test binary's own timeout kills it.
func TestCloseAbandonsATokenAlreadyOffered(t *testing.T) {
	ticks := make(chan time.Time)
	l := newLimiter(ticks)

	w := &waiter{grant: make(chan struct{}), gone: make(chan struct{})}
	l.mu.Lock()
	l.queues[ClassUserPresent] = append(l.queues[ClassUserPresent], w)
	l.mu.Unlock()

	offering := make(chan struct{})
	var once sync.Once
	l.holdBeforeOffer(func() { once.Do(func() { close(offering) }) })

	ticks <- time.Time{}
	<-offering

	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	l.awaitTicks(1)

	select {
	case <-w.grant:
		t.Error("the abandoned waiter received its token, want the offer withdrawn at shutdown")
	default:
	}
}

func TestNewLimiterPacesFromItsOwnTicker(t *testing.T) {
	l := NewLimiter(time.Millisecond)
	defer closeLimiter(t, l)

	b := l.Budget(ClassUserPresent, 0)
	if err := b.Wait(context.Background()); err != nil {
		t.Fatalf("Wait against a real ticker: %v", err)
	}
	if b.Spent() != 1 {
		t.Errorf("Spent() = %d, want 1", b.Spent())
	}
}
