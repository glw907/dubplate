package local

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

// TestNew_RecoversDeadOwners pins spec criterion 4. Owner A's lease
// file exists unlocked and names a queued and a running record; owner
// B has no lease file and names one running record. A terminal record
// naming A stays unchanged, which only a WHERE naming the two live
// states guarantees, since recovery marks A's records by owner.
func TestNew_RecoversDeadOwners(t *testing.T) {
	dir, h := newStore(t)
	a, b := newInstance(), newInstance()
	aLease := filepath.Join(leasesDir(dir), a)
	touch(t, aLease, 0o600)
	past := time.Now().Add(-time.Hour).UnixMicro()
	live := [][]byte{
		plant(t, h, record{state: "queued", owner: a, at: past}),
		plant(t, h, record{state: "running", owner: a, at: past, started: past}),
		plant(t, h, record{state: "running", owner: b, at: past, started: past}),
	}
	done := plant(t, h, record{state: "succeeded", owner: a, at: past, started: past, finished: past})
	doneBefore := mustRecord(t, h, done)

	newRunnerFor(t, h, nil, succeed)

	for _, id := range live {
		rec := mustRecord(t, h, id)
		if rec.state != "interrupted" || rec.finished == 0 || rec.updated == past {
			t.Errorf("record %x = %s, finished_at %d, updated_at %d: want interrupted with both times set now",
				id, rec.state, rec.finished, rec.updated)
		}
	}
	if doneAfter := mustRecord(t, h, done); !reflect.DeepEqual(doneAfter, doneBefore) {
		t.Errorf("terminal record changed: %+v, want %+v", doneAfter, doneBefore)
	}
	if _, err := os.Stat(aLease); !os.IsNotExist(err) {
		t.Errorf("the unlocked lease file survived recovery: stat = %v", err)
	}
}

// TestNew_LeavesHeldOwners pins the held row: a record whose owner's
// lease is held stays as it is, whatever its kind.
func TestNew_LeavesHeldOwners(t *testing.T) {
	_, h := newStore(t)
	held, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("TakeInstanceLease: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })
	ids := [][]byte{
		plant(t, h, record{state: "running", owner: held.Instance(), started: time.Now().UnixMicro()}),
		plant(t, h, record{kind: "other-kind", state: "queued", owner: held.Instance()}),
	}
	want := []string{"running", "queued"}

	newRunnerFor(t, h, nil, succeed)

	for i, id := range ids {
		if rec := mustRecord(t, h, id); rec.state != want[i] || rec.finished != 0 {
			t.Errorf("record %x = %s, finished_at %d; want %s, unfinished", id, rec.state, rec.finished, want[i])
		}
	}
}

// TestNew_UnreadableLease pins the unreadable row: a lease file the
// runner cannot open fails construction naming its path, and the
// failed runner leaves no lease of its own behind.
func TestNew_UnreadableLease(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("this test needs a non-root user: root opens a mode-0000 file")
	}
	dir, h := newStore(t)
	owner := newInstance()
	path := filepath.Join(leasesDir(dir), owner)
	touch(t, path, 0)
	id := plant(t, h, record{state: "running", owner: owner, started: time.Now().UnixMicro()})

	_, err := New(context.Background(), Config{Store: h, Logger: slog.New(slog.DiscardHandler)})
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("New() = %v, want a refusal naming %s", err, path)
	}
	if names := leaseNames(t, dir); !slices.Equal(names, []string{owner}) {
		t.Errorf("leases/ holds %v, want only the unreadable %s", names, owner)
	}
	if rec := mustRecord(t, h, id); rec.state != "running" {
		t.Errorf("record reads %s after a refused recovery, want running", rec.state)
	}
}

// touch creates an empty file at path with mode perm, as a dead
// owner's lease file is left.
func touch(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, perm)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}
