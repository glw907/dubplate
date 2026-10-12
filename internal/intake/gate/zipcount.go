// The directory-offset arithmetic in directoryStart is adapted from
// readDirectoryEnd in the Go standard library's archive/zip/reader.go.
//
// Copyright 2009 The Go Authors. All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are
// met:
//
//    * Redistributions of source code must retain the above copyright
// notice, this list of conditions and the following disclaimer.
//    * Redistributions in binary form must reproduce the above
// copyright notice, this list of conditions and the following disclaimer
// in the documentation and/or other materials provided with the
// distribution.
//    * Neither the name of Google LLC nor the names of its
// contributors may be used to endorse or promote products derived from
// this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT
// OWNER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL,
// SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
// LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE,
// DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY
// THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
// (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
// OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package gate

import (
	"archive/zip"
	"bufio"
	"encoding/binary"
	"io"
)

// Record layouts from the ZIP application note, as archive/zip names them.
const (
	directoryHeaderSignature = 0x02014b50
	directoryHeaderLen       = 46
	directoryEndSignature    = 0x06054b50
	directoryEndLen          = 22
	directory64LocSignature  = 0x07064b50
	directory64LocLen        = 20
	directory64EndSignature  = 0x06064b50
	directory64EndLen        = 56
)

// exceedsHeaders reports whether the central directory at start holds more
// than limit headers, reading them in constant memory. It stops where
// [zip.NewReader] stops: at the first record that is not a complete header.
// The end record's entry count plays no part, because archive/zip compares
// it only modulo 65536 and a zip64 record can claim any number.
//
// A read error ends the count early like a short directory; the caller
// consults its recorded reader to tell the two apart.
func exceedsHeaders(r io.ReaderAt, size, start int64, limit int) bool {
	br := bufio.NewReader(io.NewSectionReader(r, start, size-start))
	for headers := 1; skipHeader(br) == nil; headers++ {
		if headers > limit {
			return true
		}
	}
	return false
}

// skipHeader consumes one central-directory header and the name, extra
// field, and comment it carries. It fails as readDirectoryHeader does: with
// [zip.ErrFormat] on a wrong signature and a read error on a short record.
func skipHeader(br *bufio.Reader) error {
	var h [directoryHeaderLen]byte
	if _, err := io.ReadFull(br, h[:]); err != nil {
		return err
	}
	if binary.LittleEndian.Uint32(h[:]) != directoryHeaderSignature {
		return zip.ErrFormat
	}
	n := int(binary.LittleEndian.Uint16(h[28:])) +
		int(binary.LittleEndian.Uint16(h[30:])) +
		int(binary.LittleEndian.Uint16(h[32:]))
	_, err := br.Discard(n)
	return err
}

// directoryStarts returns the offsets where [zip.NewReader] may read the
// first central-directory header, base offset included. There is one unless
// the end record claims a prefix and a header-shaped record sits at the
// unprefixed offset: archive/zip then drops the prefix only when that record
// also passes readDirectoryHeader's checks, which the count does not repeat,
// so both offsets are returned and the caller counts from each.
func directoryStarts(r io.ReaderAt, size int64) ([]int64, error) {
	var end []byte
	var endOffset int64
	for i, n := range []int64{1024, 65 * 1024} {
		n = min(n, size)
		buf := make([]byte, n)
		if _, err := r.ReadAt(buf, size-n); err != nil && err != io.EOF {
			return nil, err
		}
		if p := findEndSignature(buf); p >= 0 {
			end = buf[p:]
			endOffset = size - n + int64(p)
			break
		}
		if i == 1 || n == size {
			return nil, zip.ErrFormat
		}
	}

	records := uint64(binary.LittleEndian.Uint16(end[10:]))
	dirSize := uint64(binary.LittleEndian.Uint32(end[12:]))
	dirOffset := uint64(binary.LittleEndian.Uint32(end[16:]))

	// These values mean the archive can be zip64.
	if records == 0xffff || dirSize == 0xffffffff || dirOffset == 0xffffffff {
		p, err := findDirectory64End(r, endOffset)
		if err == nil && p >= 0 {
			endOffset = p
			dirSize, dirOffset, err = readDirectory64End(r, p, dirSize, dirOffset)
		}
		if err != nil {
			return nil, err
		}
	}

	const maxInt64 = 1<<63 - 1
	if dirSize > maxInt64 || dirOffset > maxInt64 {
		return nil, zip.ErrFormat
	}
	base := endOffset - int64(dirSize) - int64(dirOffset)
	start := base + int64(dirOffset)
	if start < 0 || start >= size {
		return nil, zip.ErrFormat
	}

	// An end record that claims a prefix the directory does not carry: when
	// a header sits at the unprefixed offset, the prefix may be zero.
	if base > 0 {
		off := int64(dirOffset)
		br := bufio.NewReader(io.NewSectionReader(r, off, size-off))
		if skipHeader(br) == nil {
			return []int64{start, off}, nil
		}
	}
	return []int64{start}, nil
}

// findDirectory64End reads the zip64 locator just before the end record and
// returns the offset of the zip64 end record, or -1 when there is none.
func findDirectory64End(r io.ReaderAt, endOffset int64) (int64, error) {
	locOffset := endOffset - directory64LocLen
	if locOffset < 0 {
		return -1, nil
	}
	buf := make([]byte, directory64LocLen)
	if _, err := r.ReadAt(buf, locOffset); err != nil {
		return -1, err
	}
	if binary.LittleEndian.Uint32(buf) != directory64LocSignature ||
		binary.LittleEndian.Uint32(buf[4:]) != 0 || // the disk holding the zip64 end record
		binary.LittleEndian.Uint32(buf[16:]) != 1 { // the number of disks
		return -1, nil
	}
	return int64(binary.LittleEndian.Uint64(buf[8:])), nil
}

// readDirectory64End returns the directory size and offset the zip64 end
// record at offset holds.
func readDirectory64End(r io.ReaderAt, offset int64, dirSize, dirOffset uint64) (uint64, uint64, error) {
	buf := make([]byte, directory64EndLen)
	if _, err := r.ReadAt(buf, offset); err != nil {
		return dirSize, dirOffset, err
	}
	if binary.LittleEndian.Uint32(buf) != directory64EndSignature {
		return dirSize, dirOffset, zip.ErrFormat
	}
	return binary.LittleEndian.Uint64(buf[40:]), binary.LittleEndian.Uint64(buf[48:]), nil
}

// findEndSignature returns the offset of the last end record in b whose
// comment fits, or -1.
func findEndSignature(b []byte) int {
	for i := len(b) - directoryEndLen; i >= 0; i-- {
		if binary.LittleEndian.Uint32(b[i:]) != directoryEndSignature {
			continue
		}
		commentLen := int(binary.LittleEndian.Uint16(b[i+directoryEndLen-2:]))
		if i+directoryEndLen+commentLen > len(b) {
			return -1
		}
		return i
	}
	return -1
}
