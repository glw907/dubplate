package unidecode

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
)

// wantTableSHA256 is the sha256 tools/unidecode-table/generate.py
// reported for the table it wrote from the pinned Unidecode 1.4.0
// release, computed independently of this package's own tableSHA256
// pin: a build that happens to agree with a wrong pin still fails this
// test.
const wantTableSHA256 = "c954605b805e4a21d1d9e79b4d93bc6b89684de9996dd21720b256fc58a6ce85"

const (
	wantBlockCount = 190
	wantEntryCount = 48597
)

func TestBuildIsByteIdenticalToThePin(t *testing.T) {
	wheel := loadCachedWheel(t)

	built, err := Build(wheel)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	sum := sha256.Sum256(built)
	if got := hex.EncodeToString(sum[:]); got != wantTableSHA256 {
		t.Fatalf("Build sha256 = %s, want %s", got, wantTableSHA256)
	}

	table, err := decode(built)
	if err != nil {
		t.Fatalf("decode the built table: %v", err)
	}
	if len(table.blocks) != wantBlockCount {
		t.Fatalf("built table carries %d blocks, want %d", len(table.blocks), wantBlockCount)
	}
	var totalEntries int
	for _, entries := range table.blocks {
		totalEntries += len(entries)
	}
	if totalEntries != wantEntryCount {
		t.Fatalf("built table carries %d entries, want %d", totalEntries, wantEntryCount)
	}
}

func TestBuildRefusesWheelChecksumMismatch(t *testing.T) {
	wheel := loadCachedWheel(t)
	corrupted := append([]byte(nil), wheel...)
	corrupted[0] ^= 0xff

	_, err := Build(corrupted)
	if !errors.Is(err, ErrWheelChecksum) {
		t.Fatalf("Build(flipped byte) error = %v, want ErrWheelChecksum", err)
	}
}
