package tagio

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// TestSniff covers Sniff's table: the golden FLAC file, one blob per
// container family tagio does not yet recognize, a three-byte
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

// TestSupportedFormats pins the supported set to FLAC alone.
func TestSupportedFormats(t *testing.T) {
	if got, want := SupportedFormats(), []Format{FormatFLAC}; !slices.Equal(got, want) {
		t.Fatalf("SupportedFormats = %v, want %v", got, want)
	}
}

// TestReadAndWriteRefuseANonFLACFile grades the sniff in front of Read
// and Write: a file that is not FLAC answers Sniff's
// ErrUnsupportedContainer, worded as Sniff words it rather than as
// Parse's signature refusal, and Write leaves its bytes untouched.
// The golden still reads, writes and parses, so the refusal is the
// sniff's and not a broken entry point.
func TestReadAndWriteRefuseANonFLACFile(t *testing.T) {
	data := []byte("OggS\x00\x02\x00\x00\x00\x00\x00\x00")
	path := writeFixture(t, data)
	want := "tagio: " + path + ": tagio: unsupported container"

	if _, err := Read(path); !errors.Is(err, ErrUnsupportedContainer) || err.Error() != want {
		t.Errorf("Read error = %v, want %q", err, want)
	}
	if err := Write(path, TagSet{}, nil, 0o644); !errors.Is(err, ErrUnsupportedContainer) || err.Error() != want {
		t.Errorf("Write error = %v, want %q", err, want)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, data) {
		t.Errorf("Write changed a refused file: got %q, want %q", got, data)
	}
	wantParse := want + `: signature "OggS"`
	if _, err := Parse(path); err == nil || err.Error() != wantParse {
		t.Errorf("Parse error = %v, want %q", err, wantParse)
	}

	golden := copyFixture(t, filepath.Join("testdata", "golden.flac"))
	if _, err := Read(golden); err != nil {
		t.Errorf("Read golden: %v", err)
	}
	if err := Write(golden, TagSet{}, nil, 0o644); err != nil {
		t.Errorf("Write golden: %v", err)
	}
	if _, err := Parse(golden); err != nil {
		t.Errorf("Parse golden: %v", err)
	}
}
