package intake

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/tus/tusd/v2/pkg/memorylocker"
)

// leaseKind names the kind lease that makes one process staging's owner.
const leaseKind = "intake"

// Config is what [New] needs from the binary's wiring.
type Config struct {
	Store *store.Handle
	Log   *slog.Logger
	// Staging is the staging directory: tusd's uploads under tus/, and
	// each archive's extracted members under submissions/<id>/x/.
	Staging string
	Now     func() time.Time
	// Limits reads the intake limits in force. It runs before each
	// create and each check, so a changed setting reaches the next one.
	Limits func(ctx context.Context) (Limits, error)
	// Windows reads the sweep's windows in force, once per sweep.
	Windows func(ctx context.Context) (Windows, error)
	// SweepInterval is how often the sweep runs, and CheckWorkers how
	// many submissions are checked at once. [Intake.Start] reads both
	// once.
	SweepInterval time.Duration
	CheckWorkers  int
	// Admins reads the addresses of the live admins a stall mail goes
	// to, at each mail, so an admin removed meanwhile hears nothing.
	Admins func(ctx context.Context) ([]string, error)
	// Send mails one address.
	Send func(ctx context.Context, to, subject, body string) error
}

// Windows are the sweep's windows.
type Windows struct {
	// ResumeWindow is how long an upload may sit idle before it
	// expires. A tus upload no row names waits as long before it goes.
	ResumeWindow time.Duration
	// Hold is how long a check that could not run waits for a retry.
	Hold time.Duration
	// StallAfter is how long a check may run before the admins hear.
	StallAfter time.Duration
}

// An Intake is the server's one intake. The tus mount, the seam, the
// checking workers, and the sweep share its writer, so a receipt or a
// retry wakes the worker that checks it, and its tus locker, so
// cleanup takes the lock tusd holds an upload under before removing it.
type Intake struct {
	cfg    Config
	subs   *submissions
	root   *os.Root
	locker *memorylocker.MemoryLocker

	lease  *store.KindLease
	cancel context.CancelFunc
	done   sync.WaitGroup
}

// New returns the intake over cfg, holding the staging directory open
// until [Intake.Close]. It starts nothing.
func New(cfg Config) (*Intake, error) {
	root, err := os.OpenRoot(cfg.Staging)
	if err != nil {
		return nil, fmt.Errorf("open staging: %w", err)
	}
	return &Intake{
		cfg:    cfg,
		subs:   newSubmissions(cfg.Store, cfg.Log, cfg.Staging, cfg.Now, staging.FreeBytes),
		root:   root,
		locker: memorylocker.New(),
	}, nil
}

// Seam returns the member-facing submission operations.
func (in *Intake) Seam() *Seam {
	return &Seam{subs: in.subs, limits: in.cfg.Limits}
}

// TusMount returns the handler for the shell's subtree tus, which takes
// uploads into staging's tus directory for the submissions the store
// holds. The shell hands it the full path, prefix included.
func (in *Intake) TusMount(cfg TusConfig) (http.Handler, error) {
	return newTusMount(in.subs, in.locker, cfg)
}

// Start takes the intake lease, runs the startup pass under it, and
// starts the checking workers and the sweep, which run until ctx ends
// or [Intake.Close]. A lease another process holds refuses the start
// with an error wrapping [store.ErrLeaseBusy] that names the holder,
// before any recovery runs. Start holds the lease until Close.
func (in *Intake) Start(ctx context.Context) error {
	lease, err := in.cfg.Store.TakeKindLease(leaseKind)
	if err != nil {
		return fmt.Errorf("intake: %w", err)
	}
	if err := in.startup(ctx); err != nil {
		return errors.Join(fmt.Errorf("intake startup pass: %w", err), lease.Release())
	}
	in.lease = lease
	ctx, in.cancel = context.WithCancel(ctx)
	in.done.Go(func() { in.work(ctx) })
	in.done.Go(func() { in.sweepEvery(ctx) })
	return nil
}

// Close stops the workers and the sweep and waits for both to return,
// so no check commits and no cleanup removes after it, then releases
// the lease and the staging directory. The caller closes the store
// after it.
func (in *Intake) Close() error {
	var err error
	if in.cancel != nil {
		in.cancel()
		in.done.Wait()
		err = in.lease.Release()
	}
	return errors.Join(err, in.root.Close())
}
