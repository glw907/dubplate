package main

import (
	"errors"
	"io"
	"path/filepath"
	"slices"
	"strings"
	"testing"

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
		want    []collection
		wantErr string
	}{
		{name: "none", values: nil, want: nil},
		{
			name:   "two in order",
			values: []string{"first=a.json", "second=dir/b.json"},
			want:   []collection{{"first", "a.json"}, {"second", "dir/b.json"}},
		},
		{name: "splits at the first equals", values: []string{"a=x=y.json"}, want: []collection{{"a", "x=y.json"}}},
		{name: "no equals", values: []string{"a.json"}, wantErr: `"a.json": want name=manifest`},
		{name: "empty manifest", values: []string{"a="}, wantErr: `"a=": want name=manifest`},
		{name: "empty name is the stack's to refuse", values: []string{"=a.json"}, want: []collection{{"", "a.json"}}},
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

func TestBuildCollections(t *testing.T) {
	const seedDigest = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	one, oneDigest, err := buildCollections(t.TempDir(), []collection{{"music", fixtureManifest("manifest.json")}})
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 1 || one[0].Tracks != 29 || oneDigest != seedDigest {
		t.Errorf("one collection = %+v, digest %s; want 29 tracks and the seed digest", one, oneDigest)
	}

	root := t.TempDir()
	cols := []collection{
		{"first", fixtureManifest("collection-a.json")},
		{"second", fixtureManifest("collection-b.json")},
	}
	two, twoDigest, err := buildCollections(root, cols)
	if err != nil {
		t.Fatal(err)
	}
	if len(two) != 2 || two[0].Name != "first" || two[1].Name != "second" || two[0].Dir == two[1].Dir {
		t.Fatalf("collections = %+v, want first and second in distinct directories", two)
	}
	for _, c := range two {
		if !strings.HasPrefix(c.Dir, root) {
			t.Errorf("%s built outside the root: %s", c.Name, c.Dir)
		}
	}
	// The same libraries under swapped names are a different list.
	_, swapped, err := buildCollections(t.TempDir(), []collection{{"second", cols[0].manifest}, {"first", cols[1].manifest}})
	if err != nil {
		t.Fatal(err)
	}
	if swapped == twoDigest || twoDigest == seedDigest {
		t.Errorf("digests: two %s, swapped %s, seed %s; want all distinct", twoDigest, swapped, seedDigest)
	}

	_, _, err = buildCollections(t.TempDir(), []collection{{"gone", filepath.Join(root, "absent.json")}})
	if err == nil || !strings.Contains(err.Error(), `collection "gone"`) {
		t.Errorf("missing manifest error = %v, want one naming the collection", err)
	}
}
