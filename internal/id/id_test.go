package id

import (
	"bytes"
	"testing"
)

func TestID(t *testing.T) {
	a, b := New(), New()
	if len(a) != 16 {
		t.Fatalf("len(New()) = %d, want 16", len(a))
	}
	if a == b {
		t.Fatalf("two New() calls both returned %x", a[:])
	}
	if a.IsZero() {
		t.Fatalf("New() returned the zero id")
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want ID
		ok   bool
	}{
		{"lowercase", "0123456789abcdef0123456789abcdef", ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}, true},
		{"all zeros", "00000000000000000000000000000000", ID{}, true},
		{"31 characters", "0123456789abcdef0123456789abcde", ID{}, false},
		{"33 characters", "0123456789abcdef0123456789abcdef0", ID{}, false},
		{"uppercase", "0123456789ABCDEF0123456789ABCDEF", ID{}, false},
		{"not hex", "0123456789abcdef0123456789abcdeg", ID{}, false},
		{"empty", "", ID{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.in)
			if (err == nil) != tt.ok {
				t.Fatalf("Parse(%q) error = %v, want success %v", tt.in, err, tt.ok)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %x, want %x", tt.in, got[:], tt.want[:])
			}
		})
	}
}

func TestIsZero(t *testing.T) {
	if !(ID{}).IsZero() {
		t.Error("ID{}.IsZero() = false, want true")
	}
	if (ID{15: 1}).IsZero() {
		t.Error("an id with its last byte set reports IsZero")
	}
}

func TestString(t *testing.T) {
	id := ID{0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef, 0x01, 0x23, 0x45, 0x67, 0x89, 0xab, 0xcd, 0xef}
	const want = "0123456789abcdef0123456789abcdef"
	if got := id.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestValue(t *testing.T) {
	id := ID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	v, err := id.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	got, ok := v.([]byte)
	if !ok || !bytes.Equal(got, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}) {
		t.Errorf("Value() = %#v, want the 16 bytes as a []byte", v)
	}
}
