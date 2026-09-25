package backup

import (
	"strings"
	"testing"
)

// TestPruneCheckSafety pins the safety rule's own-snapshot rows,
// direct against checkSafety so the boundary between "equal" and
// "later" needs no real remote: R-6bi15 skips every prune step unless
// the listing holds the run's own snapshot with its stamp strictly
// greater than every other parsed stamp.
func TestPruneCheckSafety(t *testing.T) {
	own := "20260101T000000.000000Z"
	before := "20251231T000000.000000Z"
	after := "20260102T000000.000000Z"

	tests := []struct {
		name    string
		entries []remoteEntry
		wantOK  bool
		wantSub string
	}{
		{
			// The states table's "a remote stamp equal to this run's
			// own" row: the only entry's stamp equals own's. It must
			// read as the run's own snapshot, not as a foreign stamp
			// no earlier than own's (which would wrongly skip every
			// prune step on every run, since own's entry always
			// equals own's stamp).
			name:    "the listing holds only own's own snapshot",
			entries: []remoteEntry{{Name: own + ".db"}},
			wantOK:  true,
		},
		{
			// The states table's "own snapshot missing" row: an empty
			// listing (or, equivalently, an absent snapshots/) never
			// finds own's stamp.
			name:    "the listing is empty",
			entries: nil,
			wantOK:  false,
			wantSub: "is missing from the remote listing",
		},
		{
			name:    "own missing, an older foreign stamp present",
			entries: []remoteEntry{{Name: before + ".db"}},
			wantOK:  false,
			wantSub: "is missing from the remote listing",
		},
		{
			name:    "a later foreign stamp beside own",
			entries: []remoteEntry{{Name: own + ".db"}, {Name: after + ".db"}},
			wantOK:  false,
			wantSub: "is not before this run's own",
		},
		{
			name:    "a name that does not parse as a stamp is ignored",
			entries: []remoteEntry{{Name: own + ".db"}, {Name: "notes.db"}},
			wantOK:  true,
		},
		{
			name:    "a directory entry is ignored",
			entries: []remoteEntry{{Name: own + ".db"}, {Name: after, IsDir: true}},
			wantOK:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := prune{stamp: own}
			reason, ok := p.checkSafety(tt.entries)
			if ok != tt.wantOK {
				t.Fatalf("checkSafety(%v) ok = %v, reason %q; want ok = %v", tt.entries, ok, reason, tt.wantOK)
			}
			if tt.wantSub != "" && !strings.Contains(reason, tt.wantSub) {
				t.Errorf("reason %q does not contain %q", reason, tt.wantSub)
			}
		})
	}
}
