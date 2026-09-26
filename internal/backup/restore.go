package backup

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// The points in a restore at which stepHook runs, each after the step
// it names.
const (
	stepOpened        = "open"      // the lock taken and the snapshot chosen
	stepStaged        = "stage"     // the incoming store staged and verified
	stepQuiesced      = "quiesce"   // the current store settled
	stepRecorded      = "record"    // the restore-intent record written
	stepSetAside      = "swap-out"  // the current store set aside
	stepSwapped       = "swap"      // the incoming store in db/, every rename synced
	stepReferenced    = "reference" // the backup reference rewritten
	stepLibraryOwed   = "library"   // the record owing the library copy
	stepCopied        = "copy"      // the library copied
	stepReconcileOwed = "reconcile" // the record owing the report, or owing serve
	stepReported      = "report"    // the report written
)

// stepHook runs after each step of a restore, given the step's name. It
// does nothing in production; a test sets it to kill a re-executed
// test binary at one step, so every crash the tests prove is a real
// process death.
var stepHook = func(string) {}

// A RestoreConfig is one restore's inputs. Exactly one of Snapshot,
// FromFile, and Resume is set.
type RestoreConfig struct {
	// StateDir holds the store leaf.
	StateDir string

	// BusyTimeout bounds how long a statement waits on a SQLite lock.
	BusyTimeout time.Duration

	// Now stamps a new restore; a resume keeps its record's stamp.
	Now time.Time

	// Snapshot is the stamp of the snapshot to fetch from Remote, or
	// "latest" for the newest one there.
	Snapshot string

	// Older permits a Snapshot other than the latest. It never
	// combines with Library.
	Older bool

	// FromFile is a local store file to restore from.
	FromFile string

	// Resume continues the restore the leaf's restore-intent record
	// names, from the remote and library root the record names.
	Resume bool

	// DiscardCurrent moves a current store that cannot be opened aside,
	// unopened, where a restore would otherwise refuse it.
	DiscardCurrent bool

	// Remote is the backup remote, as [ParseRemote] reads it, that a
	// Snapshot restore reads.
	Remote string

	// Library also restores the library tree, from Remote's library/
	// into LibraryRoot. It needs Snapshot.
	Library      bool
	LibraryRoot  string
	IntoNonempty bool

	// Out receives what the restore reports to the operator.
	Out io.Writer

	// Logger receives the store's warnings.
	Logger *slog.Logger
}

// check refuses a combination of sources and flags a restore cannot
// run.
func (c RestoreConfig) check() error {
	n := 0
	for _, set := range []bool{c.Snapshot != "", c.FromFile != "", c.Resume} {
		if set {
			n++
		}
	}
	switch {
	case n != 1:
		return errors.New("restore: name exactly one of a snapshot, a file, or a resume")
	case c.Older && c.Snapshot == "":
		return errors.New("restore: --older applies only to a snapshot")
	case c.Older && c.Library:
		return errors.New("restore: --older never combines with --library: the library on the remote is the latest backup's")
	case c.Library && c.Snapshot == "":
		return errors.New("restore: --library needs a snapshot from the remote the library is on")
	case c.IntoNonempty && !c.Library:
		return errors.New("restore: --into-nonempty applies only to --library")
	}
	return nil
}

// Restore restores the store under cfg.StateDir, and with cfg.Library
// its library tree, from a remote snapshot or a local file, or resumes
// a restore a crash or a cancel interrupted.
//
// A new restore stages and verifies the incoming store, settles the
// current one, and writes a restore-intent record before it renames any
// file; from then on every writing open refuses the leaf until the
// restore finishes. It then swaps the incoming store into db/, keeping
// the current one as a pre-restore copy, rewrites the backup reference
// to name a restored remote snapshot, copies the library, and writes
// the reconciliation report. A restored store older than this binary is
// left for serve, which migrates it, to finish; Restore then reports
// that serve must start.
//
// A resume takes up from the step the record owes, from the source and
// library root the record names: it never fetches again once the record
// exists, and it judges every version against this binary rather than
// the record. A resume refuses a record that only serve can finish.
//
// A canceled ctx stops the restore at the step it reached, and no
// state is recorded after the cancel.
func Restore(ctx context.Context, cfg RestoreConfig) (err error) {
	if err := cfg.check(); err != nil {
		return err
	}
	var remote Remote
	if cfg.Snapshot != "" {
		if remote, err = ParseRemote(cfg.Remote); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
	}
	var root string
	if cfg.Library {
		if root, err = resolveRoot(cfg.LibraryRoot); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
	}

	sess, err := store.OpenRestore(cfg.StateDir, store.RestoreOptions{
		BusyTimeout: cfg.BusyTimeout,
		Now:         cfg.Now,
		Resume:      cfg.Resume,
		FromFile:    cfg.FromFile,
		Logger:      cfg.Logger,
	})
	if err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	defer func() { err = errors.Join(err, sess.Close()) }()

	r := restorer{sess: sess, out: cfg.Out}
	if cfg.Resume {
		return r.resume(ctx)
	}
	in, err := r.stage(ctx, cfg, remote)
	if err != nil {
		return err
	}
	in.LibraryRoot, in.IntoNonempty = root, cfg.IntoNonempty
	return r.record(ctx, cfg.DiscardCurrent, in)
}

// A restorer runs a restore's steps on its session.
type restorer struct {
	sess *store.RestoreSession
	out  io.Writer
}

// transport returns a transport whose rclone inherits the session's
// exclusive store lock, so a transfer that outlives the restore keeps
// every other writer out until it exits.
func (r restorer) transport() (transport, error) {
	bin, err := exec.LookPath("rclone")
	if err != nil {
		return transport{}, fmt.Errorf("restore: rclone is not on PATH: %w", err)
	}
	return transport{bin: bin, inherit: r.sess.DupFiles}, nil
}

// stage runs steps 1 and 2: it chooses the snapshot, fetches it or
// copies the local file into staging/, and verifies the stage. It
// returns the source the record will name.
func (r restorer) stage(ctx context.Context, cfg RestoreConfig, remote Remote) (store.RestoreIntent, error) {
	var in store.RestoreIntent
	if cfg.Snapshot == "" {
		stepHook(stepOpened)
		abs, err := filepath.Abs(cfg.FromFile)
		if err != nil {
			return in, fmt.Errorf("restore from %s: %w", cfg.FromFile, err)
		}
		if err := r.sess.StageFile(); err != nil {
			return in, err
		}
		in.File = abs
	} else {
		t, err := r.transport()
		if err != nil {
			return in, err
		}
		src := source{t: t, remote: remote}
		if err := r.warnRemote(remote); err != nil {
			return in, err
		}
		snap, err := src.resolve(ctx, cfg.Snapshot, cfg.Older)
		if err != nil {
			return in, err
		}
		stepHook(stepOpened)
		if err := src.fetch(ctx, r.sess, r.out, snap); err != nil {
			return in, err
		}
		in.Remote, in.Snapshot = cfg.Remote, snap.stamp
	}
	if _, err := r.sess.VerifyStage(ctx); err != nil {
		return in, err
	}
	stepHook(stepStaged)
	return in, nil
}

// warnRemote prints a warning when the remote a restore reads is not
// the one the leaf's backup reference names.
func (r restorer) warnRemote(remote Remote) error {
	referenced, err := r.sess.BackupRemote()
	if err != nil {
		return err
	}
	if w := remoteWarning(remote, referenced); w != "" {
		return r.say("%s", w)
	}
	return nil
}

// record runs steps 3 and 4, settling the current store and writing the
// restore-intent record, and goes on to the swap.
func (r restorer) record(ctx context.Context, discard bool, in store.RestoreIntent) error {
	if err := r.sess.Quiesce(ctx, discard); err != nil {
		return err
	}
	stepHook(stepQuiesced)
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	if err := r.sess.Record(in); err != nil {
		return err
	}
	stepHook(stepRecorded)
	return r.swap(ctx, in)
}

// resume takes up the restore from the step its record owes.
func (r restorer) resume(ctx context.Context) error {
	in := r.sess.Intent()
	switch r.sess.Owed() {
	case store.StepSwap:
		return r.swap(ctx, in)
	case store.StepLibrary:
		return r.library(ctx, in)
	case store.StepReconcile:
		return r.advance(ctx, in)
	default:
		return errors.New("restore --resume: the restored store is older than this binary: start serve to finish the restore")
	}
}

// swap runs step 5 in its two halves, then rewrites the backup
// reference for a remote snapshot, and advances the record. A resume
// runs it whole, since each half does only what remains of it.
func (r restorer) swap(ctx context.Context, in store.RestoreIntent) error {
	moved, err := r.sess.SwapOut(ctx)
	if err != nil {
		return err
	}
	for _, p := range moved {
		if err := r.say("moved the current store's file aside to %s", p); err != nil {
			return err
		}
	}
	stepHook(stepSetAside)
	if err := r.sess.SwapIn(); err != nil {
		return err
	}
	stepHook(stepSwapped)
	if in.Snapshot != "" {
		if err := r.sess.WriteBackupRef(in.Snapshot, in.Remote); err != nil {
			return err
		}
		stepHook(stepReferenced)
	}
	return r.advance(ctx, in)
}

// advance records the step just done and runs the one now owed: the
// library copy, the report, or nothing, for a store only serve can
// finish.
func (r restorer) advance(ctx context.Context, in store.RestoreIntent) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	step, err := r.sess.Advance(ctx)
	if err != nil {
		return err
	}
	switch step {
	case store.StepLibrary:
		stepHook(stepLibraryOwed)
		return r.library(ctx, in)
	case store.StepServe:
		stepHook(stepReconcileOwed)
		return r.say("the restored store is older than this binary: start serve to finish the restore")
	}
	stepHook(stepReconcileOwed)
	return r.report(ctx, in)
}

// library runs step 6, copying the remote's library/ into the library
// root the record names, and advances the record.
func (r restorer) library(ctx context.Context, in store.RestoreIntent) error {
	remote, err := ParseRemote(in.Remote)
	if err != nil {
		return fmt.Errorf("restore the library: %w", err)
	}
	t, err := r.transport()
	if err != nil {
		return err
	}
	if err := t.run(ctx, nil, "copy", remote.join("library"), in.LibraryRoot); err != nil {
		return fmt.Errorf("restore the library into %s: %w", in.LibraryRoot, err)
	}
	stepHook(stepCopied)
	return r.advance(ctx, in)
}

// report runs step 7: it writes the reconciliation report and removes
// the record, which ends the restore.
func (r restorer) report(ctx context.Context, in store.RestoreIntent) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	path, err := r.sess.WriteReport(reportText(r.sess.Stamp(), in))
	if err != nil {
		return err
	}
	stepHook(stepReported)
	if err := r.sess.Complete(); err != nil {
		return err
	}
	return r.say("restored; the reconciliation report is %s", path)
}

// reportText is the reconciliation report of the restore stamped stamp
// from the source in names.
func reportText(stamp string, in store.RestoreIntent) []byte {
	src := "file " + in.File
	if in.Snapshot != "" {
		src = "snapshot " + in.Snapshot + " on " + in.Remote
	}
	return []byte("restore " + stamp + "\nsource: " + src + "\n")
}

// say writes one line to the operator.
func (r restorer) say(format string, args ...any) error {
	if _, err := fmt.Fprintf(r.out, format+"\n", args...); err != nil {
		return fmt.Errorf("restore: write output: %w", err)
	}
	return nil
}
