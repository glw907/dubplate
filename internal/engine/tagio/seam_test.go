package tagio

import (
	"errors"
	"path/filepath"
	"slices"
	"testing"
)

// TestSniff covers Sniff's table: the golden FLAC file, one blob per
// container family the registry does not yet recognize, a three-byte
// file, an empty file, and a file whose first three bytes agree with
// FLAC's signature but whose fourth does not. Every fixture rides
// writeFixture, which always names its file "fixture.flac" regardless
// of the bytes it holds, so no row's success or failure ever turns on
// the filename's extension.
func TestSniff(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		golden     bool
		wantFormat Format
		wantErr    error
	}{
		{
			name:       "FLAC golden",
			golden:     true,
			wantFormat: FormatFLAC,
		},
		{
			name:    "ID3v2-prefixed blob",
			data:    []byte("ID3\x04\x00\x00\x00\x00\x00\x00rest of an MP3 file"),
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "MPEG frame-sync blob",
			data:    []byte{0xff, 0xfb, 0x90, 0x00, 0x00, 0x00, 0x00, 0x00},
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "OggS blob",
			data:    []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00"),
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "ftyp-boxed blob",
			data:    []byte{0x00, 0x00, 0x00, 0x20, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm'},
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "three-byte file",
			data:    []byte{'f', 'L', 'a'},
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "empty file",
			data:    []byte{},
			wantErr: ErrUnsupportedContainer,
		},
		{
			name:    "fLaX, fourth magic byte does not match",
			data:    []byte("fLaX the rest of a stream that is not FLAC"),
			wantErr: ErrUnsupportedContainer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", "golden.flac")
			if !tt.golden {
				path = writeFixture(t, tt.data)
			}

			format, err := Sniff(path)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Sniff error = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Sniff: %v", err)
			}
			if format != tt.wantFormat {
				t.Errorf("Sniff format = %v, want %v", format, tt.wantFormat)
			}
		})
	}
}

// TestSupportedFormatsReturnsCopy grades the supported set's copy
// contract on Mapping's own precedent: the registry carries exactly
// one entry, and mutating a returned slice must never touch the next
// call's result.
func TestSupportedFormatsReturnsCopy(t *testing.T) {
	first := SupportedFormats()
	if len(first) != 1 || first[0] != FormatFLAC {
		t.Fatalf("SupportedFormats = %v, want exactly [FormatFLAC]", first)
	}

	first[0] = Format(99)

	second := SupportedFormats()
	if len(second) != 1 || second[0] != FormatFLAC {
		t.Fatalf("mutating the first call's slice changed the second: %v", second)
	}
}

// TestSupportedFormatsOrderedByValue proves the order is fixed rather
// than a map's own iteration order, which Go deliberately randomizes:
// a synthetic three-row registry, inserted in descending Format
// value, must still come back ascending.
func TestSupportedFormatsOrderedByValue(t *testing.T) {
	saved := registry
	registry = map[Format]formatEntry{
		Format(5):  {},
		Format(2):  {},
		FormatFLAC: {magic: flacMagic, read: readFLAC, write: writeFLAC},
	}
	t.Cleanup(func() { registry = saved })

	got := SupportedFormats()
	want := []Format{FormatFLAC, Format(2), Format(5)}
	if !slices.Equal(got, want) {
		t.Fatalf("SupportedFormats = %v, want %v in ascending Format order", got, want)
	}
}

// TestEmptyRegistryBypassesReadAndWriteButNotParse asserts the
// dispatch is real rather than decorative: emptying the registry
// stops Read and Write from reaching the golden, each answering
// ErrUnsupportedContainer, while Parse keeps parsing it because Parse
// deliberately never consults the registry (ruled, m6). Restoring the
// registry restores all three.
func TestEmptyRegistryBypassesReadAndWriteButNotParse(t *testing.T) {
	path := copyFixture(t, filepath.Join("testdata", "golden.flac"))

	saved := registry
	registry = map[Format]formatEntry{}
	t.Cleanup(func() { registry = saved })

	if _, err := Read(path); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Read with emptied registry error = %v, want %v", err, ErrUnsupportedContainer)
	}
	if err := Write(path, TagSet{}, nil, 0o644); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Write with emptied registry error = %v, want %v", err, ErrUnsupportedContainer)
	}
	if _, err := Parse(path); err != nil {
		t.Errorf("Parse with emptied registry: %v, want the golden to still parse", err)
	}

	registry = saved

	if _, err := Read(path); err != nil {
		t.Errorf("Read after restoring registry: %v", err)
	}
	if err := Write(path, TagSet{}, nil, 0o644); err != nil {
		t.Errorf("Write after restoring registry: %v", err)
	}
	if _, err := Parse(path); err != nil {
		t.Errorf("Parse after restoring registry: %v, want the golden to still parse", err)
	}
}

// TestReadAndWriteRefuseAFuncLessRegistryRow proves Read and Write
// answer ErrUnsupportedContainer rather than panicking when the
// registry carries a row for the sniffed format whose func values are
// unset: Sniff succeeds against the row's magic, so the panic this
// guards against would otherwise happen inside Read and Write
// themselves, not inside Sniff.
func TestReadAndWriteRefuseAFuncLessRegistryRow(t *testing.T) {
	path := copyFixture(t, filepath.Join("testdata", "golden.flac"))

	saved := registry
	registry = map[Format]formatEntry{FormatFLAC: {magic: flacMagic}}
	t.Cleanup(func() { registry = saved })

	if _, err := Read(path); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Read against a func-less registry row error = %v, want %v", err, ErrUnsupportedContainer)
	}
	if err := Write(path, TagSet{}, nil, 0o644); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Write against a func-less registry row error = %v, want %v", err, ErrUnsupportedContainer)
	}
}

// TestReadAndWriteRouteThroughTheRegistry grades the dispatch's other
// direction: registering a second, deliberately wrong magic for
// FormatFLAC makes the golden unreadable and unwritable, proving Read
// and Write consult the registry's magic-to-implementation row rather
// than calling the FLAC implementation by name.
func TestReadAndWriteRouteThroughTheRegistry(t *testing.T) {
	path := copyFixture(t, filepath.Join("testdata", "golden.flac"))

	saved := registry
	registry = map[Format]formatEntry{
		FormatFLAC: {
			magic: [4]byte{'x', 'x', 'x', 'x'},
			read:  readFLAC,
			write: writeFLAC,
		},
	}
	t.Cleanup(func() { registry = saved })

	if _, err := Read(path); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Read against a mismatched magic error = %v, want %v", err, ErrUnsupportedContainer)
	}
	if err := Write(path, TagSet{}, nil, 0o644); !errors.Is(err, ErrUnsupportedContainer) {
		t.Errorf("Write against a mismatched magic error = %v, want %v", err, ErrUnsupportedContainer)
	}
}
