package gate

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/glw907/dubplate/internal/reject"
)

const countLimit = 5

// manyEntries returns an archive of n empty tracks.
func manyEntries(t *testing.T, n int) []byte {
	t.Helper()
	f := newZIPFixture(t)
	for i := range n {
		f.file(fmt.Sprintf("%02d.flac", i), "")
	}
	return f.bytes()
}

// understated rewrites the end record to claim a single entry, as a hostile
// archive does to slip past a count read from it.
func understated(b []byte) []byte {
	c := bytes.Clone(b)
	end := len(c) - 22 // the writer emits no archive comment
	binary.LittleEndian.PutUint16(c[end+8:], 1)
	binary.LittleEndian.PutUint16(c[end+10:], 1)
	return c
}

// zip64Claiming appends a zip64 end record claiming one entry and points the
// end record at it, the archive the verification probe read 196,609 entries
// from.
func zip64Claiming(b []byte) []byte {
	end := len(b) - 22
	dirSize := binary.LittleEndian.Uint32(b[end+12:])
	dirOffset := binary.LittleEndian.Uint32(b[end+16:])

	rec := make([]byte, 56)
	binary.LittleEndian.PutUint32(rec[0:], 0x06064b50)
	binary.LittleEndian.PutUint64(rec[4:], 44)
	binary.LittleEndian.PutUint64(rec[24:], 1) // entries on this disk
	binary.LittleEndian.PutUint64(rec[32:], 1) // entries
	binary.LittleEndian.PutUint64(rec[40:], uint64(dirSize))
	binary.LittleEndian.PutUint64(rec[48:], uint64(dirOffset))

	loc := make([]byte, 20)
	binary.LittleEndian.PutUint32(loc[0:], 0x07064b50)
	binary.LittleEndian.PutUint64(loc[8:], uint64(end))
	binary.LittleEndian.PutUint32(loc[16:], 1)

	tail := bytes.Clone(b[end:])
	binary.LittleEndian.PutUint16(tail[8:], 0xffff)
	binary.LittleEndian.PutUint16(tail[10:], 0xffff)

	out := bytes.Clone(b[:end])
	out = append(out, rec...)
	out = append(out, loc...)
	return append(out, tail...)
}

func prefixed(b []byte) []byte {
	return append(bytes.Repeat([]byte("#!/bin/sh\n"), 10), b...)
}

func TestWalkZIPCountsCentralDirectoryHeaders(t *testing.T) {
	p := zipPolicy()
	p.MaxFiles = countLimit

	tests := []struct {
		name string
		zip  []byte
		want reject.Code // empty: the walk passes
	}{
		{"at the limit", manyEntries(t, countLimit), ""},
		{"one over the limit", manyEntries(t, countLimit+1), reject.TooManyFiles},
		{"an end record that understates the count", understated(manyEntries(t, countLimit+1)), reject.TooManyFiles},
		{"a zip64 end record that claims one entry", zip64Claiming(manyEntries(t, countLimit+1)), reject.TooManyFiles},
		{"a prefix before the archive", prefixed(manyEntries(t, countLimit+1)), reject.TooManyFiles},
		{"a prefixed archive at the limit", prefixed(manyEntries(t, countLimit)), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, w, err := walkStaged(t, tt.zip, p, false)
			if tt.want == "" {
				if err != nil {
					t.Fatalf("walk: %v", err)
				}
				if len(w.Members) != countLimit {
					t.Errorf("members = %d, want %d", len(w.Members), countLimit)
				}
				return
			}
			wantRefusal(t, err, tt.want, "")
		})
	}
}

// TestExceedsHeaders holds the count at its boundary, on an
// archive the standard reader would hold in memory header by header.
func TestExceedsHeaders(t *testing.T) {
	data := manyEntries(t, countLimit+1)
	r := &recordedReaderAt{r: bytes.NewReader(data)}
	start, err := directoryStart(r, int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	if !exceedsHeaders(r, int64(len(data)), start, countLimit) {
		t.Error("exceedsHeaders = false, want true for one header over the limit")
	}
	if exceedsHeaders(r, int64(len(data)), start, countLimit+1) {
		t.Error("exceedsHeaders = true at exactly the limit")
	}
	if zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data))); err != nil || len(zr.File) != countLimit+1 {
		t.Errorf("the standard reader reads %d entries, %v: the fixture is not what the test says", len(zr.File), err)
	}
}

func TestDirectoryStart(t *testing.T) {
	data := manyEntries(t, 2)
	end := len(data) - 22
	dirOffset := int64(binary.LittleEndian.Uint32(data[end+16:]))

	tests := []struct {
		name string
		data []byte
		want int64
		err  bool
	}{
		{"a plain archive", data, dirOffset, false},
		{"a prefixed archive", prefixed(data), dirOffset + 100, false},
		{"a zip64 end record", zip64Claiming(data), dirOffset, false},
		{"not an archive", []byte("plainly not a zip"), 0, true},
		{"an empty file", nil, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := directoryStart(bytes.NewReader(tt.data), int64(len(tt.data)))
			if (err != nil) != tt.err || got != tt.want {
				t.Errorf("directoryStart = %d, %v, want %d, error %t", got, err, tt.want, tt.err)
			}
		})
	}
}
