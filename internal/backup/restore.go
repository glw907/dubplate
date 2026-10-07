package backup

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
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
	// unopened, where a restore would otherwise refuse it. It never
	// combines with Resume.
	DiscardCurrent bool

	// Remote is the backup remote, as [ParseRemote] reads it, that a
	// Snapshot restore reads.
	Remote string

	// Library also restores the library tree, from Remote's library/
	// into LibraryRoot, which must exist and, unless IntoNonempty is
	// set, be empty but for a lost+found directory. It needs Snapshot.
	Library bool

	// LibraryRoot is the library root. Without Library, the
	// reconciliation report still reads the tree under it.
	LibraryRoot string

	// IntoNonempty permits a Library restore into a root that holds
	// files. A local file the copy replaces is moved into the sibling
	// directory <root>.restore-displaced-<stamp>, never deleted. A
	// resumed copy displaces the same way whether or not it is set.
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
	case c.DiscardCurrent && c.Resume:
		return errors.New("restore: --discard-current never combines with --resume")
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
		// The record takes only an absolute root, and refusing one there
		// would come after the fetch and the quiesce.
		if !filepath.IsAbs(root) {
			return fmt.Errorf("restore: library root %s is not an absolute path", root)
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

	r := restorer{sess: sess, out: cfg.Out, resumed: cfg.Resume, root: cfg.LibraryRoot}
	if cfg.Resume {
		return r.resume(ctx)
	}
	if cfg.Library {
		if err := checkRoot(root, displacedPath(root, sess.Stamp()), cfg.IntoNonempty); err != nil {
			return err
		}
	}
	in, err := r.stage(ctx, cfg, remote, root)
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

	// resumed is set when the restore takes up a record an earlier run
	// wrote.
	resumed bool

	// root is the configured library root, which the report reads when
	// the restore owes no library half.
	root string
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

// stage chooses the snapshot, checks the remote library against root
// when one is owed, fetches the snapshot or copies the local file into
// staging/, and verifies the stage. It returns the source the record
// will name.
func (r restorer) stage(ctx context.Context, cfg RestoreConfig, remote Remote, root string) (store.RestoreIntent, error) {
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
		if cfg.Library {
			if err := src.checkLibrary(ctx, root, r.out); err != nil {
				return in, err
			}
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

// record settles the current store, writes the restore-intent record,
// and goes on to the swap.
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

// swap sets the current store aside and then moves the incoming store
// into db/, rewrites the backup reference for a remote snapshot, and
// advances the record. A resume runs both halves, since each does only
// what remains of it.
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
		if err := r.sayKept(); err != nil {
			return err
		}
		return r.say("the restored store is older than this binary: start serve to finish the restore")
	}
	stepHook(stepReconcileOwed)
	return r.report(ctx, in)
}

// library copies the remote's library/ into the library root the
// record names, and advances the record. It re-resolves the
// root, refusing one that is gone or now resolves elsewhere, but never
// re-checks that the root is empty: the record proves the check passed
// before any copy, and a copy this restore began and was stopped fills
// the root itself. A resumed copy moves each local file it replaces
// into the displaced directory, with or without --into-nonempty.
func (r restorer) library(ctx context.Context, in store.RestoreIntent) error {
	remote, err := ParseRemote(in.Remote)
	if err != nil {
		return fmt.Errorf("restore the library: %w", err)
	}
	root, err := resolveRoot(in.LibraryRoot)
	if err != nil {
		return fmt.Errorf("restore the library: the recorded library root: %w", err)
	}
	if root != in.LibraryRoot {
		return fmt.Errorf("restore the library: the recorded library root %s now resolves to %s", in.LibraryRoot, root)
	}
	t, err := r.transport()
	if err != nil {
		return err
	}

	args := []string{"copy", remote.join("library"), root}
	var displaced string
	// A resume displaces whatever the first run's flags were: a person
	// may have placed a file in the root between the runs. It moves none
	// of the files the interrupted copy wrote, since rclone's copy skips
	// a file that is identical at the destination and --backup-dir takes
	// only a file the copy overwrites.
	if in.IntoNonempty || r.resumed {
		// A resume never ran checkRoot, and a link may have been made
		// since any run's check, while the fetch, quiesce, and swap ran.
		if err := refuseSymlinks(root); err != nil {
			return err
		}
		displaced = displacedPath(root, r.sess.Stamp())
		if err := r.makeDisplaced(displaced); err != nil {
			return err
		}
		args = append(args, "--backup-dir", displaced)
	}
	if err := t.run(ctx, nil, args...); err != nil {
		if f, ok := errors.AsType[*failure](err); ok {
			err = fmt.Errorf("%s: %w", f.code, err)
		}
		return fmt.Errorf("restore the library into %s: %w", root, err)
	}
	if displaced != "" {
		if err := removeIfEmpty(displaced); err != nil {
			return err
		}
	}
	stepHook(stepCopied)
	return r.advance(ctx, in)
}

// displacedPath is where a restore stamped stamp moves each local file
// under root that its library copy replaces: a sibling of root, since
// rclone refuses a backup directory inside the copy's destination.
func displacedPath(root, stamp string) string {
	return root + ".restore-displaced-" + stamp
}

// checkRoot checks the resolved library root before any rclone run. It
// refuses a root that holds anything but a lost+found directory, which
// a freshly made ext file system holds, unless intoNonempty is set;
// with it, it refuses a root holding a symlink, as [refuseSymlinks]
// does, and a displaced path that already exists, which a restore
// never merges into.
func checkRoot(root, displaced string, intoNonempty bool) error {
	des, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("restore the library: %w", err)
	}
	des = slices.DeleteFunc(des, func(de fs.DirEntry) bool { return de.Name() == "lost+found" && de.IsDir() })
	if !intoNonempty {
		if len(des) > 0 {
			return fmt.Errorf("restore the library: library root %s is not empty: restore into it with --into-nonempty", root)
		}
		return nil
	}
	if err := refuseSymlinks(root); err != nil {
		return err
	}
	switch _, err := os.Lstat(displaced); {
	case err == nil:
		return fmt.Errorf("restore the library: %s already exists: move it away before a restore with --into-nonempty", displaced)
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("restore the library: %w", err)
	}
	return nil
}

// refuseSymlinks refuses root when a symlink sits anywhere under it,
// naming the first the walk meets. rclone v1.74.3 lists a symlinked
// directory as absent, so a copy writes the remote's file through the
// link and never moves the local one into the backup directory, and it
// replaces a symlinked file with a regular one. Either way a local file
// is lost with nothing displaced to name it.
//
// The walk passes over root's lost+found, which a fresh ext file system
// makes owned by root at mode 0700 and [checkRoot] allows. It refuses
// any other directory it cannot read, naming it, since a link inside
// could go unseen: rclone's copy leaves such a directory unread when
// the remote holds no directory of that name, and fails on it when the
// remote does.
func refuseSymlinks(root string) error {
	lostFound := filepath.Join(root, "lost+found")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if errors.Is(err, fs.ErrPermission) {
			return fmt.Errorf("%s cannot be read, so the restore cannot check it for symlinks: make it readable by the user the restore runs as, or move it out of the library root", quotePath(p))
		}
		if err != nil {
			return err
		}
		if p == lostFound && d.IsDir() {
			return fs.SkipDir
		}
		if d.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("%s is a symlink: a library copy writes through a symlink without displacing the local file; replace the link with what it names, or remove it", quotePath(p))
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("restore the library: %w", err)
	}
	return nil
}

// checkLibrary checks the remote library and the library root before
// the fetch. It refuses a remote with no library/, and a library root
// whose file system lacks room for the remote tree's bytes, a floor
// since a backend may report no size for a file. It prints the root,
// whether it is a mount point, and the free space it read, since an
// unmounted mount point is an empty directory no check can tell from a
// root on its parent's file system.
func (s source) checkLibrary(ctx context.Context, root string, out io.Writer) error {
	dirs, err := s.t.list(ctx, s.remote.join(""), true)
	if err != nil {
		return err
	}
	if !slices.ContainsFunc(dirs, func(e remoteEntry) bool { return e.IsDir && e.Name == "library" }) {
		return fmt.Errorf("restore the library: the remote holds no %s", s.remote.join("library"))
	}
	size, err := s.librarySize(ctx)
	if err != nil {
		return err
	}
	free, mount, err := rootSpace(root)
	if err != nil {
		return err
	}
	kind := "not a mount point"
	if mount {
		kind = "a mount point"
	}
	if _, err := fmt.Fprintf(out, "library root %s is %s: %d bytes free; the remote's library/ holds at least %d bytes in %d files, %d of them without a size\n",
		root, kind, free, size.Bytes, size.Count, size.Sizeless); err != nil {
		return fmt.Errorf("restore: write output: %w", err)
	}
	if free < uint64(max(size.Bytes, 0)) {
		return fmt.Errorf("restore the library: free space on %s: %d bytes free, need %d", root, free, size.Bytes)
	}
	return nil
}

// rootSpace reports the free bytes on root's file system, and whether
// root is a mount point: whether its device differs from its parent
// directory's.
func rootSpace(root string) (free uint64, mount bool, err error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(root, &st); err != nil {
		return 0, false, fmt.Errorf("restore the library: statfs %s: %w", root, err)
	}
	dev, err := device(root)
	if err != nil {
		return 0, false, err
	}
	parent, err := device(filepath.Dir(root))
	if err != nil {
		return 0, false, err
	}
	return st.Bavail * uint64(st.Frsize), dev != parent, nil
}

func device(p string) (uint64, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, fmt.Errorf("restore the library: %w", err)
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, fmt.Errorf("restore the library: %s: no device in its file info", p)
	}
	return st.Dev, nil
}

// makeDisplaced creates the displaced directory at mode 0700, since
// rclone would create a missing one at 0755 under umask 022. A first
// run refuses a directory that exists; a resume accepts the one its
// interrupted run created, after [checkDisplaced] passes it.
func (r restorer) makeDisplaced(p string) error {
	err := os.Mkdir(p, 0o700)
	if r.resumed && errors.Is(err, fs.ErrExist) {
		return checkDisplaced(p, os.Getuid())
	}
	if err != nil {
		return fmt.Errorf("restore the library: create the displaced directory: %w", err)
	}
	return nil
}

// checkDisplaced refuses p unless it is a directory, not a symlink,
// owned by uid at mode 0700, as [restorer.makeDisplaced] creates it.
func checkDisplaced(p string, uid int) error {
	info, err := os.Lstat(p)
	if err != nil {
		return fmt.Errorf("restore the library: %w", err)
	}
	var found string
	st, ok := info.Sys().(*syscall.Stat_t)
	switch {
	case info.Mode()&fs.ModeSymlink != 0:
		found = "a symlink"
	case !info.IsDir():
		found = "not a directory"
	case !ok || int(st.Uid) != uid:
		found = "not owned by uid " + strconv.Itoa(uid)
	case info.Mode().Perm() != 0o700:
		found = fmt.Sprintf("at mode %#o", info.Mode().Perm())
	default:
		return nil
	}
	return fmt.Errorf("restore the library: displaced directory %s is %s: want a directory at mode 0700 owned by uid %d", p, found, uid)
}

// removeIfEmpty removes the displaced directory p when the copy moved
// nothing into it. One that holds anything is kept, whichever run
// moved the files, and never removed recursively.
func removeIfEmpty(p string) error {
	des, err := os.ReadDir(p)
	if err != nil {
		return fmt.Errorf("restore the library: read the displaced directory: %w", err)
	}
	if len(des) > 0 {
		return nil
	}
	if err := os.Remove(p); err != nil {
		return fmt.Errorf("restore the library: remove the empty displaced directory: %w", err)
	}
	return nil
}

// report reconciles the restored catalog against the library root,
// writes the report, and removes the record, which ends the restore.
func (r restorer) report(ctx context.Context, in store.RestoreIntent) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("restore: %w", err)
	}
	root := cmp.Or(in.LibraryRoot, r.root)
	rep, err := r.reconcile(ctx, root)
	if err != nil {
		return err
	}
	var line string
	if in.LibraryRoot != "" {
		if line, err = displacedLine(displacedPath(root, r.sess.Stamp())); err != nil {
			return err
		}
	}
	head := reportHead(r.sess.Stamp(), in, line)
	path, err := r.sess.WriteReport(append([]byte(head), rep.render(math.MaxInt)...))
	if err != nil {
		return err
	}
	stepHook(stepReported)
	if err := r.sayKept(); err != nil {
		return err
	}
	if err := r.sess.Complete(); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(r.out, "%s%s", head, rep.render(summaryLimit)); err != nil {
		return fmt.Errorf("restore: write output: %w", err)
	}
	return r.say("restored; the reconciliation report is %s", path)
}

// reconcile reconciles the restored store against root on a read
// handle the session's lock covers.
func (r restorer) reconcile(ctx context.Context, root string) (_ *report, err error) {
	h, err := r.sess.OpenRead(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, h.Close()) }()
	return reconcile(ctx, h, root)
}

// reportHead is the head every restore's report carries: the stamp, the
// source, and displaced, the displaced-files line or "".
func reportHead(stamp string, in store.RestoreIntent, displaced string) string {
	return "restore " + stamp + "\nsource: " + sourceLine(in) + "\n" + displaced
}

// sourceLine names the source in names.
func sourceLine(in store.RestoreIntent) string {
	if in.Snapshot != "" {
		return "snapshot " + in.Snapshot + " on " + in.Remote
	}
	return "file " + in.File
}

// displacedLine names the displaced directory p and how many files it
// holds, or is "" when the copy displaced nothing and p is gone.
func displacedLine(p string) (string, error) {
	if _, err := os.Lstat(p); errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	n, err := countFiles(p)
	if err != nil {
		return "", fmt.Errorf("restore: count the displaced files: %w", err)
	}
	return fmt.Sprintf("displaced local files: %s holds %d\n", p, n), nil
}

// sayKept tells the operator where the pre-restore copy is, when the
// record owes one. No prune removes the copy, since no backup holds the
// writes the restore rolled away.
func (r restorer) sayKept() error {
	if p := r.sess.PreRestorePath(); p != "" {
		return r.say("the replaced store is kept at %s until you remove it by hand: no backup holds the writes the restore rolled away", p)
	}
	return nil
}

// say writes one line to the operator.
func (r restorer) say(format string, args ...any) error {
	if _, err := fmt.Fprintf(r.out, format+"\n", args...); err != nil {
		return fmt.Errorf("restore: write output: %w", err)
	}
	return nil
}
