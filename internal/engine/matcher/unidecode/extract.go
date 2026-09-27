package unidecode

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Build turns a Unidecode wheel into this package's table format,
// refusing at the first stage that disagrees with the pin: the
// wheel's own checksum, the zip structure, the set of data modules,
// each module's grammar, and finally the built table's own checksum.
// Nothing after a refused stage runs, so a caller never scores against
// a table this extraction only partly trusts.
func Build(wheel []byte) ([]byte, error) {
	sum := sha256.Sum256(wheel)
	got := hex.EncodeToString(sum[:])
	if got != WheelSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrWheelChecksum, got, WheelSHA256)
	}

	blocks, err := extract(wheel)
	if err != nil {
		return nil, err
	}

	t := &Table{version: Version, sweepLo: sweepLo, sweepHi: sweepHi, blocks: blocks}
	t.digest = t.sweepDigest()

	encoded := t.encode()
	sum = sha256.Sum256(encoded)
	got = hex.EncodeToString(sum[:])
	if got != tableSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrTableChecksum, got, tableSHA256)
	}
	return encoded, nil
}

// extract reads every entry of a wheel's zip archive to EOF, refusing
// at the first entry whose CRC or declared size disagrees with its
// bytes, so a defect anywhere in the archive is caught rather than
// skipped alongside the entries this package has no use for. Every
// unidecode/x???.py data module is lexed; a name that looks like a
// data module but does not match exactly x plus three lowercase hex
// digits and a .py suffix, or a section seen twice, is refused. So is
// any count other than dataModuleCount: a pinned wheel cannot ship an
// anomalous name, a duplicate section, or a different module count, so
// any of those is a defect in this extraction, not a file to skip.
func extract(wheel []byte) (map[uint16][]*string, error) {
	r, err := zip.NewReader(bytes.NewReader(wheel), int64(len(wheel)))
	if err != nil {
		return nil, fmt.Errorf("unidecode: read wheel zip: %w", err)
	}

	blocks := make(map[uint16][]*string)
	matched := 0
	for _, f := range r.File {
		section, isModule, nameErr := dataModuleName(f.Name)

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("unidecode: open %s: %w", f.Name, err)
		}

		if !isModule {
			_, readErr := io.Copy(io.Discard, rc)
			closeErr := rc.Close()
			if nameErr != nil {
				return nil, nameErr
			}
			if readErr != nil {
				return nil, fmt.Errorf("unidecode: read %s: %w", f.Name, readErr)
			}
			if closeErr != nil {
				return nil, fmt.Errorf("unidecode: close %s: %w", f.Name, closeErr)
			}
			continue
		}

		content, err := io.ReadAll(rc)
		closeErr := rc.Close()
		if err != nil {
			return nil, fmt.Errorf("unidecode: read %s: %w", f.Name, err)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("unidecode: close %s: %w", f.Name, closeErr)
		}

		entries, err := lexModule(f.Name, content)
		if err != nil {
			return nil, err
		}
		if _, dup := blocks[section]; dup {
			return nil, fmt.Errorf("unidecode: %s: duplicate data module for section %#04x", f.Name, section)
		}
		blocks[section] = entries
		matched++
	}
	if matched != dataModuleCount {
		return nil, fmt.Errorf("unidecode: wheel carries %d data modules, want %d", matched, dataModuleCount)
	}
	return blocks, nil
}

// dataModuleName reports whether entryName is a data module under
// unidecode/ (unidecode/x000.py through unidecode/x1f6.py), and if so
// its section id. The Go form of tools/unidecode-table/generate.py's
// `len(name) == 4 and name[0] == "x"` module filter, applied to a zip
// entry's path instead of an imported module's __name__. Only the
// `unidecode/x` prefix gates a skip: any entry under it that does not
// match exactly x plus three lowercase hex digits and a .py suffix
// (the wrong extension, wrong length, or uppercase hex) is refused by
// name rather than silently skipped. A pinned wheel cannot ship one,
// so an entry shaped like a near-miss module name is a defect in the
// extraction, not a file to pass over.
func dataModuleName(entryName string) (section uint16, ok bool, err error) {
	const prefix = "unidecode/x"
	if !strings.HasPrefix(entryName, prefix) {
		return 0, false, nil
	}
	const suffix = ".py"
	hex := strings.TrimPrefix(entryName, prefix)
	if !strings.HasSuffix(hex, suffix) {
		return 0, false, fmt.Errorf("unidecode: %s: refused as a data module name, want unidecode/x[0-9a-f]{3}.py", entryName)
	}
	hex = strings.TrimSuffix(hex, suffix)
	if len(hex) != 3 || !isLowerHex3(hex) {
		return 0, false, fmt.Errorf("unidecode: %s: refused as a data module name, want unidecode/x[0-9a-f]{3}.py", entryName)
	}
	id, err := strconv.ParseUint(hex, 16, 16)
	if err != nil {
		return 0, false, fmt.Errorf("unidecode: %s: not a hex section id: %w", entryName, err)
	}
	return uint16(id), true, nil
}

// isLowerHex3 reports whether s is exactly three lowercase hex digits.
func isLowerHex3(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, b := range []byte(s) {
		if (b < '0' || b > '9') && (b < 'a' || b > 'f') {
			return false
		}
	}
	return true
}

// sweepDigest hashes t's own [Table.Lookup] over every codepoint in
// [sweepLo, sweepHi], in tools/unidecode-table/generate.py's
// sweep_digest encoding. It runs before the table's bytes exist, so it
// reads t's in-memory blocks directly rather than through Parse.
func (t *Table) sweepDigest() [32]byte {
	h := sha256.New()
	var cp [4]byte
	var length [4]byte
	for codepoint := uint32(sweepLo); codepoint <= sweepHi; codepoint++ {
		binary.BigEndian.PutUint32(cp[:], codepoint)
		h.Write(cp[:])
		repl, ok := t.Lookup(rune(codepoint))
		if !ok {
			h.Write([]byte{0x00})
			continue
		}
		h.Write([]byte{0x01})
		binary.BigEndian.PutUint32(length[:], uint32(len(repl)))
		h.Write(length[:])
		h.Write([]byte(repl))
	}
	var digest [32]byte
	copy(digest[:], h.Sum(nil))
	return digest
}
