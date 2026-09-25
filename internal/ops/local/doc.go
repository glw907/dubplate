// Package local runs the jobs [ops.Seam] names inside this process,
// against a store opened in write mode, and recovers the jobs of a
// process that died.
//
// A [Runner] holds an instance lease for its life and names that lease
// as the owner of every job record it inserts, so any process can tell
// a live owner's record from a dead one's by trying the owner's lease.
// A job of one kind runs at a time across every process, admitted by
// the store's kind lease.
//
// A job's state changes only by a statement that names the state it
// expects in its WHERE clause, so a terminal state never changes: a
// late transition into a record another process already finished
// matches no row and is refused. A cancel from another process reaches
// a running job through the record's cancel_requested column, which
// the owning runner reads at each progress report and on a fixed
// interval while the job runs.
package local
