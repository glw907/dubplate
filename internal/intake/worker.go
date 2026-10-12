package intake

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"path"
	"runtime/debug"
	"slices"
	"sync"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/reject"
)

// work runs CheckWorkers checks at a time until ctx ends. The store is
// the queue: each wake lists the checking submissions and hands each one
// no worker holds to a worker. A check that commits nothing, on a store
// fault, waits for the next wake, which the sweep sends every interval.
func (in *Intake) work(ctx context.Context) {
	jobs := make(chan []byte)
	var mu sync.Mutex
	held := make(map[string]bool)
	var workers sync.WaitGroup
	for range in.cfg.CheckWorkers {
		workers.Go(func() {
			for id := range jobs {
				in.check(ctx, id)
				mu.Lock()
				delete(held, string(id))
				mu.Unlock()
			}
		})
	}
	defer workers.Wait()
	defer close(jobs)

	for {
		ids, err := in.queue(ctx)
		if err != nil && ctx.Err() == nil {
			in.cfg.Log.Error("intake: read the check queue", "error", err)
		}
		for _, id := range ids {
			mu.Lock()
			busy := held[string(id)]
			held[string(id)] = true
			mu.Unlock()
			if busy {
				continue
			}
			select {
			case jobs <- id:
			case <-ctx.Done():
				return
			}
		}
		select {
		case <-in.subs.wake:
		case <-ctx.Done():
			return
		}
	}
}

// queue lists the checking submissions, longest waiting first.
func (in *Intake) queue(ctx context.Context) ([][]byte, error) {
	return queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) ([]byte, error) {
		var id []byte
		return id, rows.Scan(&id)
	}, `SELECT id FROM submissions WHERE state = ? ORDER BY updated_at, id`, stateChecking)
}

// check runs the gate over checking submission id and commits its
// verdict, then removes the bytes the verdict leaves outside the staged
// set. A check that could not run commits check-unfinished, keeping the
// staged bytes and the reservation. A cancelled ctx commits nothing:
// the submission stays checking with its bytes and reservation for the
// next start, so a deploy never asks a member to retry. A store fault
// commits nothing either, and the next wake tries again.
func (in *Intake) check(ctx context.Context, id []byte) {
	log := in.cfg.Log.With("submission", wireID(id))
	v, err := in.judgeContained(ctx, id, log)
	if ctx.Err() != nil {
		return
	}
	switch {
	case errors.Is(err, errWrongState):
		// Settled already, by the reservation's own refusal or another
		// writer; the bytes it left are cleanup's.
	case errors.Is(err, gate.ErrCouldNotRun):
		log.Warn("intake: a check could not run", "error", err)
		if err := in.unfinish(ctx, id); err != nil {
			log.Error("intake: record a check that could not run", "error", err)
		}
		return
	case err != nil:
		log.Error("intake: a check could not read the store", "error", err)
		return
	default:
		if err := in.subs.commitVerdict(ctx, id, v); err != nil && !errors.Is(err, errWrongState) {
			log.Error("intake: commit a verdict", "error", err)
			return
		}
	}
	if err := in.cleanSubmission(ctx, id); err != nil {
		log.Error("intake: clean up after a verdict", "error", err)
	}
}

// judgeContained is judge with a panic in a decoder or the ZIP reader, on a
// member's bytes, turned into [gate.ErrCouldNotRun]. The panic would
// otherwise end the process, and the startup pass would queue the same
// submission to end it again.
func (in *Intake) judgeContained(ctx context.Context, id []byte, log *slog.Logger) (v verdict, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("intake: a check panicked", "panic", r, "stack", string(debug.Stack()))
			v, err = verdict{}, fmt.Errorf("%w: panic: %v", gate.ErrCouldNotRun, r)
		}
	}()
	return in.judge(ctx, id)
}

// unfinish moves checking submission id to check-unfinished, keeping
// its staged bytes and its reservation. An archive's extracted members
// are scratch until a verdict, so they go first: the retry extracts
// again.
func (in *Intake) unfinish(ctx context.Context, id []byte) error {
	if err := in.root.RemoveAll(extractDir(id)); err != nil {
		return fmt.Errorf("remove extracted members: %w", err)
	}
	return in.subs.commitVerdict(ctx, id, verdict{to: stateCheckUnfinished, reason: string(reject.CouldNotCheck)})
}

// extractDir is where submission id's archive extracts, under staging.
func extractDir(id []byte) string {
	return path.Join("submissions", hex.EncodeToString(id), "x")
}

// uploadName is a tus upload's data file under staging.
func uploadName(file []byte) string {
	return path.Join("tus", hex.EncodeToString(file))
}

// checked is a checking submission as its check reads it.
type checked struct {
	id        []byte
	unit      gate.Unit
	container gate.Container
	// allowlist is the collection's admission_format_allowlist.
	allowlist sql.NullString
	files     []checkedFile
}

type checkedFile struct {
	id, path []byte
	role     string
}

// judge reads checking submission id and runs the gate over its staged
// bytes. An error wrapping [gate.ErrCouldNotRun] is the server failing,
// errWrongState says the submission is not checking, and any other
// error is a store fault.
func (in *Intake) judge(ctx context.Context, id []byte) (verdict, error) {
	sub, err := in.readChecking(ctx, id)
	if err != nil {
		return verdict{}, err
	}
	lim, err := in.cfg.Limits(ctx)
	if err != nil {
		return verdict{}, fmt.Errorf("read intake limits: %w", err)
	}
	lim.Policy.Allowlist = narrow(sub.allowlist, lim.Policy.Allowlist)
	if sub.container == gate.ContainerZip {
		return in.judgeZIP(ctx, sub, lim)
	}
	return in.judgeFiles(ctx, sub, lim.Policy)
}

// readChecking reads submission id and its files, in path order, while
// it is checking, and returns errWrongState when it is not.
func (in *Intake) readChecking(ctx context.Context, id []byte) (checked, error) {
	sub := checked{id: id}
	err := in.cfg.Store.QueryRow(ctx, `SELECT s.verdict_unit, s.container, c.admission_format_allowlist
		FROM submissions s JOIN collections c ON c.id = s.collection_id WHERE s.id = ? AND s.state = ?`,
		id, stateChecking).Scan(&sub.unit, &sub.container, &sub.allowlist)
	if errors.Is(err, sql.ErrNoRows) {
		return checked{}, fmt.Errorf("%w: %x is not checking", errWrongState, id)
	}
	if err != nil {
		return checked{}, fmt.Errorf("read submission: %w", err)
	}
	sub.files, err = queryAll(ctx, in.cfg.Store, func(rows *sql.Rows) (checkedFile, error) {
		var f checkedFile
		return f, rows.Scan(&f.id, &f.path, &f.role)
	}, `SELECT id, rel_path, role FROM submission_files WHERE submission_id = ? ORDER BY rel_path`, id)
	if err != nil {
		return checked{}, fmt.Errorf("read submission files: %w", err)
	}
	return sub, nil
}

// judgeFiles checks each loose audio file where tusd left it. The
// extras ride along unchecked.
func (in *Intake) judgeFiles(ctx context.Context, sub checked, p gate.Policy) (verdict, error) {
	var v verdict
	var results []gate.FileResult
	var extras []string
	for _, f := range sub.files {
		if f.role != "audio" {
			extras = append(extras, string(f.path))
			continue
		}
		res, err := gate.CheckFile(ctx, gate.StagedFile{Root: in.root, Name: uploadName(f.id), Path: string(f.path)}, p, sub.unit)
		if err != nil {
			return verdict{}, err
		}
		v.files = append(v.files, fileVerdict{id: f.id, result: res})
		results = append(results, res)
	}
	v.conclude(sub.unit, results, extras)
	return v, nil
}

// judgeZIP walks the staged archive, reserves its extracted total,
// extracts it under names the server mints, and checks each audio
// member. A reservation the space limits refuse rejects the submission
// in its own write, which judgeZIP reports as errWrongState.
func (in *Intake) judgeZIP(ctx context.Context, sub checked, lim Limits) (verdict, error) {
	archive := sub.files[0].id
	walk, err := gate.WalkZIP(in.root, uploadName(archive), lim.Policy, sub.unit == gate.UnitTrack)
	if err != nil {
		return zipFault(archive, err)
	}
	defer func() { _ = walk.Close() }()

	refused, err := in.subs.reserveExtracted(ctx, sub.id, walk.Total, lim.limits())
	switch {
	case err != nil:
		return verdict{}, err
	case refused != "":
		return verdict{}, fmt.Errorf("%w: refused %s at extraction", errWrongState, refused)
	}

	dir := extractDir(sub.id)
	ids := make([][]byte, len(walk.Members))
	names := make([]string, len(walk.Members))
	for i := range walk.Members {
		ids[i] = store.NewID()
		names[i] = hex.EncodeToString(ids[i])
	}
	if err := walk.Extract(ctx, in.root, dir, names); err != nil {
		return zipFault(archive, err)
	}

	var v verdict
	var results []gate.FileResult
	var extras []string
	for i, m := range walk.Members {
		member := memberVerdict{id: ids[i], path: m.Path, size: m.Size}
		if gate.IsAudio(string(m.Path)) {
			res, err := gate.CheckFile(ctx, gate.StagedFile{Root: in.root, Name: path.Join(dir, names[i]), Path: string(m.Path)},
				lim.Policy, sub.unit)
			if err != nil {
				return verdict{}, err
			}
			member.result = res
			results = append(results, res)
		} else {
			extras = append(extras, string(m.Path))
		}
		v.members = append(v.members, member)
	}
	for _, name := range walk.Extras {
		extras = append(extras, string(name))
	}
	v.conclude(sub.unit, results, extras)
	return v, nil
}

// zipFault turns an error from the archive's walk or extraction into a
// verdict when it is the archive's fault, recorded on the submission and
// on its archive, and otherwise into the server's failure to run.
func zipFault(archive []byte, err error) (verdict, error) {
	var refusal *gate.ZIPRefusal
	if !errors.As(err, &refusal) {
		return verdict{}, fmt.Errorf("%w: %w", gate.ErrCouldNotRun, err)
	}
	code := refusal.Reason.Code
	return verdict{to: stateRejected, reason: string(code),
		files: []fileVerdict{{id: archive, result: gate.FileResult{Verdict: code}}}}, nil
}

// conclude sets v's move from results, the gate's result for each audio
// file, and extras, the paths of the rest. An album with a failing file
// is rejected and its per-file verdicts say why; one whose files all
// pass meets the album checks, and passing them waits in matching with
// its key and warnings. Separate tracks are accepted while one passes,
// and the extras a separate-tracks archive dropped are named in a
// warning each.
func (v *verdict) conclude(unit gate.Unit, results []gate.FileResult, extras []string) {
	if unit == gate.UnitTrack {
		v.to = stateRejected
		if slices.ContainsFunc(results, func(r gate.FileResult) bool { return r.Verdict == "" }) {
			v.to = stateAccepted
		}
		for _, name := range extras {
			v.warnings = append(v.warnings, gate.FileReason{Path: []byte(name), Reason: reject.Of(reject.ExtrasDropped)})
		}
		return
	}
	if slices.ContainsFunc(results, func(r gate.FileResult) bool { return r.Verdict != "" }) {
		v.to = stateRejected
		return
	}
	album := gate.CheckAlbum(results, extras)
	if album.Verdict != "" {
		v.to, v.reason = stateRejected, string(album.Verdict)
		return
	}
	v.to, v.albumKey = stateMatching, album.Key
	for _, code := range album.Warnings {
		v.warnings = append(v.warnings, gate.FileReason{Reason: reject.Of(code)})
	}
}
