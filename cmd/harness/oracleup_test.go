package main

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/stack"
	"github.com/spf13/cobra"
)

func TestOracleUpNamesMissingPodman(t *testing.T) {
	t.Setenv("PATH", t.TempDir())
	cmd := newOracleUpCmd()
	cmd.SetArgs(nil)
	err := cmd.ExecuteContext(t.Context())
	if !errors.Is(err, stack.ErrNoPodman) {
		t.Fatalf("error = %v, want ErrNoPodman", err)
	}
	if !strings.Contains(err.Error(), "podman") {
		t.Errorf("error %q does not name podman", err)
	}
}

func TestOracleUpRegistered(t *testing.T) {
	cmd, _, err := newRootCmd().Find([]string{"oracle-up"})
	if err != nil || cmd.Name() != "oracle-up" {
		t.Fatalf("Find(oracle-up) = %v, %v", cmd, err)
	}
}

func TestParseCollections(t *testing.T) {
	tests := []struct {
		name    string
		values  []string
		want    []fixture.Collection
		wantErr string
	}{
		{name: "none", values: nil, want: nil},
		{
			name:   "two in order",
			values: []string{"first=a.json", "second=dir/b.json"},
			want:   []fixture.Collection{{Name: "first", Manifest: "a.json"}, {Name: "second", Manifest: "dir/b.json"}},
		},
		{name: "splits at the first equals", values: []string{"a=x=y.json"}, want: []fixture.Collection{{Name: "a", Manifest: "x=y.json"}}},
		{name: "no equals", values: []string{"a.json"}, wantErr: `"a.json": want name=manifest`},
		{name: "empty manifest", values: []string{"a="}, wantErr: `"a=": want name=manifest`},
		{name: "empty name is the stack's to refuse", values: []string{"=a.json"}, want: []fixture.Collection{{Name: "", Manifest: "a.json"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCollections(tt.values)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error = %v, want one containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("collections = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCollectionAndManifestFlagsExclude(t *testing.T) {
	for _, newCmd := range []func() *cobra.Command{newOracleUpCmd, newRecordOracleCmd} {
		cmd := newCmd()
		cmd.SetArgs([]string{"--manifest", "m.json", "--collection", "a=m.json"})
		cmd.SetErr(io.Discard)
		err := cmd.ExecuteContext(t.Context())
		if err == nil || !strings.Contains(err.Error(), "none of the others can be") {
			t.Errorf("%s with both flags: error = %v, want a mutual-exclusion refusal", cmd.Name(), err)
		}
	}
}
