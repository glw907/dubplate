package serve

import (
	"errors"
	"slices"
	"testing"
)

func TestSplitBindAddrs(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "single address",
			raw:  "127.0.0.1",
			want: []string{"127.0.0.1"},
		},
		{
			name: "several addresses",
			raw:  "127.0.0.1,::1",
			want: []string{"127.0.0.1", "::1"},
		},
		{
			name: "whitespace around a separator",
			raw:  "127.0.0.1 , ::1",
			want: []string{"127.0.0.1", "::1"},
		},
		{
			name:    "empty entry",
			raw:     "127.0.0.1,,::1",
			wantErr: true,
		},
		{
			name:    "unparseable entry",
			raw:     "127.0.0.1,not-an-address",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SplitBindAddrs(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("SplitBindAddrs(%q) error = nil, want error", tt.raw)
				}
				if !errors.Is(err, ErrInvalidBindAddr) {
					t.Errorf("SplitBindAddrs(%q) error = %v, want it to wrap ErrInvalidBindAddr", tt.raw, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("SplitBindAddrs(%q) error = %v, want nil", tt.raw, err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("SplitBindAddrs(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}
