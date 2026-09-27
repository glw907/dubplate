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
	t.digest = sweepDigest(t)

	encoded := encode(t)
	sum = sha256.Sum256(encoded)
	got = hex.EncodeToString(sum[:])
	if got != tableSHA256 {
		return nil, fmt.Errorf("%w: got %s, want %s", ErrTableChecksum, got, tableSHA256)
	}
	return encoded, nil
}

// extract reads every unidecode/x???.py data module out of a wheel's
// zip archive and lexes it, refusing a module whose name looks like a
// data module but fails to parse as one hex section id, and refusing
// any count other than dataModuleCount: a pinned wheel cannot ship an
// anomalous name or a different module count, so either is a defect in
// this extraction, not a file to skip.
func extract(wheel []byte) (map[uint16][]*string, error) {
	r, err := zip.NewReader(bytes.NewReader(wheel), int64(len(wheel)))
	if err != nil {
		return nil, fmt.Errorf("unidecode: read wheel zip: %w", err)
	}

	blocks := make(map[uint16][]*string)
	for _, f := range r.File {
		base, isModule := dataModuleName(f.Name)
		if !isModule {
			continue
		}
		section, err := strconv.ParseUint(base[1:], 16, 16)
		if err != nil {
			return nil, fmt.Errorf("unidecode: %s: not a hex section id: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("unidecode: open %s: %w", f.Name, err)
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
		blocks[uint16(section)] = entries
	}
	if len(blocks) != dataModuleCount {
		return nil, fmt.Errorf("unidecode: wheel carries %d data modules, want %d", len(blocks), dataModuleCount)
	}
	return blocks, nil
}

// dataModuleName reports whether name is a data module under
// unidecode/ (unidecode/x000.py through unidecode/x1f6.py), and if so
// its base name without the .py suffix. The Go form of
// tools/unidecode-table/generate.py's `len(name) == 4 and name[0] ==
// "x"` module filter, applied to a zip entry's path instead of an
// imported module's __name__.
func dataModuleName(entryName string) (base string, ok bool) {
	const prefix = "unidecode/"
	const suffix = ".py"
	if !strings.HasPrefix(entryName, prefix) || !strings.HasSuffix(entryName, suffix) {
		return "", false
	}
	base = strings.TrimSuffix(strings.TrimPrefix(entryName, prefix), suffix)
	if len(base) != 4 || base[0] != 'x' {
		return "", false
	}
	return base, true
}

// sweepDigest hashes t's own [Table.Lookup] over every codepoint in
// [sweepLo, sweepHi], in tools/unidecode-table/generate.py's
// sweep_digest encoding. It runs before the table's bytes exist, so it
// reads t's in-memory blocks directly rather than through Parse.
func sweepDigest(t *Table) [32]byte {
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
