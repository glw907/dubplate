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
}

func TestExtractRefusesNotAZip(t *testing.T) {
	if _, err := extract([]byte("not a zip file")); err == nil {
		t.Fatal("extract(garbage) = nil error, want a zip read error")
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
}
