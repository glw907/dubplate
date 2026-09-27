package unidecode

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"strings"
	"testing"
)

// buildZip runs build against a fresh zip.Writer and returns the
// closed archive's bytes.
func buildZip(t *testing.T, build func(*zip.Writer)) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	build(zw)
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	return buf.Bytes()
}

func addFile(t *testing.T, zw *zip.Writer, name, content string) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatalf("create %s: %v", name, err)
	}
	if _, err := w.Write([]byte(content)); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
}

// addRawFile adds an entry via CreateRaw, whose stored CRC32 and
// UncompressedSize64 are exactly what the caller sets: CreateRaw never
// overwrites them from the written bytes, unlike Create.
func addRawFile(t *testing.T, zw *zip.Writer, name string, content []byte, crc uint32, uncompressedSize uint64) {
	t.Helper()
	fh := &zip.FileHeader{
		Name:               name,
		Method:             zip.Store,
		CRC32:              crc,
		UncompressedSize64: uncompressedSize,
		CompressedSize64:   uint64(len(content)),
	}
	w, err := zw.CreateRaw(fh)
	if err != nil {
		t.Fatalf("create raw %s: %v", name, err)
	}
	if _, err := w.Write(content); err != nil {
		t.Fatalf("write raw %s: %v", name, err)
	}
}

// validModules builds n synthetic data modules, one distinct string
// entry each, plus a few non-data files a real wheel also carries, so
// extract's own skip logic is exercised alongside its data path.
func validModulesZip(t *testing.T, n int) []byte {
	t.Helper()
	return buildZip(t, func(zw *zip.Writer) {
		for i := range n {
			name := fmt.Sprintf("unidecode/x%03x.py", i)
			addFile(t, zw, name, fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
		addFile(t, zw, "unidecode/__init__.py", "# not a data module\n")
		addFile(t, zw, "unidecode/py.typed", "")
		addFile(t, zw, "Unidecode-1.4.0.dist-info/METADATA", "Version: 1.4.0\n")
	})
}

func TestExtractValidWheel(t *testing.T) {
	blocks, err := extract(validModulesZip(t, dataModuleCount))
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(blocks) != dataModuleCount {
		t.Fatalf("extract returned %d blocks, want %d", len(blocks), dataModuleCount)
	}
	entries, ok := blocks[0x07]
	if !ok || len(entries) != 1 || entries[0] == nil || *entries[0] != "v007" {
		t.Fatalf("blocks[0x07] = %v, want one entry \"v007\"", entries)
	}
}

func TestExtractRefusesWrongCount(t *testing.T) {
	_, err := extract(validModulesZip(t, dataModuleCount-1))
	if err == nil {
		t.Fatal("extract(189 modules) = nil error, want a count refusal")
	}
	if !strings.Contains(err.Error(), "189") {
		t.Fatalf("extract(189 modules) error = %q, want it to name the count", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a module-count refusal must not read as a wheel checksum error")
	}
}

func TestExtractRefusesTooManyModules(t *testing.T) {
	_, err := extract(validModulesZip(t, dataModuleCount+1))
	if err == nil {
		t.Fatal("extract(191 modules) = nil error, want a count refusal")
	}
	if !strings.Contains(err.Error(), "191") {
		t.Fatalf("extract(191 modules) error = %q, want it to name the count", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a module-count refusal must not read as a wheel checksum error")
	}
}

func TestExtractRefusesMalformedDataName(t *testing.T) {
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		for i := range dataModuleCount {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
		addFile(t, zw, "unidecode/xzzz.py", "data = (\n'bad',\n)\n")
	})
	_, err := extract(zipBytes)
	if err == nil {
		t.Fatal("extract with a malformed data-module name = nil error, want a refusal")
	}
	if !strings.Contains(err.Error(), "xzzz.py") {
		t.Fatalf("extract error = %q, want it to name unidecode/xzzz.py", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a malformed-name refusal must not read as a wheel checksum error")
	}
}

// TestExtractRefusesNearMissDataName covers names that carry the
// unidecode/x prefix dataModuleName's own prefix check alone would
// accept, but whose hex portion is the wrong length, uses uppercase
// digits, or whose extension is not .py: strconv.ParseUint accepts
// A-F, so the length and case rules must be enforced separately from
// the parse.
func TestExtractRefusesNearMissDataName(t *testing.T) {
	tests := []struct {
		name  string
		entry string
	}{
		{name: "wrong length, too short", entry: "unidecode/x12.py"},
		{name: "wrong length, too long", entry: "unidecode/x0000.py"},
		{name: "uppercase hex", entry: "unidecode/xABC.py"},
		{name: "wrong extension", entry: "unidecode/x000.pyc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			zipBytes := buildZip(t, func(zw *zip.Writer) {
				for i := range dataModuleCount {
					addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
				}
				addFile(t, zw, tt.entry, "data = (\n'bad',\n)\n")
			})
			_, err := extract(zipBytes)
			if err == nil {
				t.Fatalf("extract with entry %s = nil error, want a refusal", tt.entry)
			}
			if !strings.Contains(err.Error(), tt.entry) {
				t.Fatalf("extract error = %q, want it to name %s", err, tt.entry)
			}
			if errors.Is(err, ErrWheelChecksum) {
				t.Fatal("a near-miss name refusal must not read as a wheel checksum error")
			}
		})
	}
}

func TestExtractRefusesDuplicateSection(t *testing.T) {
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		for i := range dataModuleCount {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
		addFile(t, zw, "unidecode/x000.py", "data = (\n'dup',\n)\n")
	})
	_, err := extract(zipBytes)
	if err == nil {
		t.Fatal("extract with a duplicated section = nil error, want a refusal")
	}
	if !strings.Contains(err.Error(), "unidecode/x000.py") {
		t.Fatalf("extract error = %q, want it to name the duplicate entry", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a duplicate-section refusal must not read as a wheel checksum error")
	}
}

func TestExtractRefusesNotAZip(t *testing.T) {
	_, err := extract([]byte("not a zip file"))
	if err == nil {
		t.Fatal("extract(garbage) = nil error, want a zip read error")
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a zip-read refusal must not read as a wheel checksum error")
	}
}

func TestExtractRefusesCRCMismatch(t *testing.T) {
	content := []byte("data = (\n'v000',\n)\n")
	wrongCRC := crc32.ChecksumIEEE(content) + 1
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		addRawFile(t, zw, "unidecode/x000.py", content, wrongCRC, uint64(len(content)))
		for i := 1; i < dataModuleCount; i++ {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
	})
	_, err := extract(zipBytes)
	if !errors.Is(err, zip.ErrChecksum) {
		t.Fatalf("extract with a bad CRC error = %v, want zip.ErrChecksum", err)
	}
	if !strings.Contains(err.Error(), "unidecode/x000.py") {
		t.Fatalf("extract error = %q, want it to name the entry", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a zip-level CRC error must not read as a wheel checksum error")
	}
}

// TestExtractRefusesNonModuleCRCMismatch proves that extract's drain of
// a non-module entry (the io.Copy to io.Discard dataModuleName's
// isModule=false path takes) still reads the entry through the zip
// reader's own CRC check, so a corrupted entry this package has no use
// for is refused rather than skipped unread.
func TestExtractRefusesNonModuleCRCMismatch(t *testing.T) {
	content := []byte("Version: 1.4.0\n")
	wrongCRC := crc32.ChecksumIEEE(content) + 1
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		for i := range dataModuleCount {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
		addRawFile(t, zw, "Unidecode-1.4.0.dist-info/METADATA", content, wrongCRC, uint64(len(content)))
	})
	_, err := extract(zipBytes)
	if !errors.Is(err, zip.ErrChecksum) {
		t.Fatalf("extract with a bad CRC on a non-module entry error = %v, want zip.ErrChecksum", err)
	}
	if !strings.Contains(err.Error(), "dist-info/METADATA") {
		t.Fatalf("extract error = %q, want it to name the entry", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a non-module CRC error must not read as a wheel checksum error")
	}
}

// TestExtractRefusesNonModuleSizeMismatch is TestExtractRefusesNonModuleCRCMismatch's
// size-declaration counterpart: the drain must read to EOF against the
// declared UncompressedSize64, not merely open and discard the entry.
func TestExtractRefusesNonModuleSizeMismatch(t *testing.T) {
	content := []byte("Version: 1.4.0\n")
	crc := crc32.ChecksumIEEE(content)
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		for i := range dataModuleCount {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
		addRawFile(t, zw, "Unidecode-1.4.0.dist-info/METADATA", content, crc, uint64(len(content))+5)
	})
	_, err := extract(zipBytes)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("extract with an oversized declared length on a non-module entry error = %v, want io.ErrUnexpectedEOF", err)
	}
	if !strings.Contains(err.Error(), "dist-info/METADATA") {
		t.Fatalf("extract error = %q, want it to name the entry", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a non-module size error must not read as a wheel checksum error")
	}
}

func TestExtractRefusesSizeMismatch(t *testing.T) {
	content := []byte("data = (\n'v000',\n)\n")
	crc := crc32.ChecksumIEEE(content)
	zipBytes := buildZip(t, func(zw *zip.Writer) {
		addRawFile(t, zw, "unidecode/x000.py", content, crc, uint64(len(content))+5)
		for i := 1; i < dataModuleCount; i++ {
			addFile(t, zw, fmt.Sprintf("unidecode/x%03x.py", i), fmt.Sprintf("data = (\n'v%03d',\n)\n", i))
		}
	})
	_, err := extract(zipBytes)
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("extract with an oversized declared length error = %v, want io.ErrUnexpectedEOF", err)
	}
	if !strings.Contains(err.Error(), "unidecode/x000.py") {
		t.Fatalf("extract error = %q, want it to name unidecode/x000.py", err)
	}
	if errors.Is(err, ErrWheelChecksum) {
		t.Fatal("a size-mismatch refusal must not read as a wheel checksum error")
	}
}
