package main

import (
	"bytes"
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// storeRowLines is the store row's fixed contribution to doctor's
// line-count assertions: it prints exactly one line regardless of
// state, unlike reg's own declared families, so a count built from reg
// alone must add it explicitly.
const storeRowLines = 1

// membersRowLines is the members row's fixed contribution to doctor's
// line-count assertions, storeRowLines's own precedent: it prints
// exactly one line regardless of state, so a count built from reg alone
// must add it explicitly. It is a literal, so a row dropped from doctor
// changes the count a test expects.
const membersRowLines = 1

// backupDoctorRowsOff and backupDoctorRowsOn are Task 7's backup
// family's fixed contribution to doctor's line-count assertions: the
// four facts rows (backupLastSuccessRow, backupWALSizeRow,
// backupFreeSpaceRow, backupLocalCopiesRow) plus the one "backup
// remote" row, all of which print unconditionally regardless of reg's
// own declared families and regardless of DUBPLATE_BACKUP_ENABLED,
// storeRowLines's own precedent: the backup remote row reads
// "unavailable" rather than disappear while the flag is off, so both
// constants carry the same total. Named separately, one per flag
// state, because [backupRemoteRowStatus]'s content, not its presence,
// is what the flag controls, and a caller asserting a flag-off total
// should never have to know that fact to pick the right constant.
const (
	backupDoctorRowsOff = 5
	backupDoctorRowsOn  = 5
)

// unidecodeRowLines is the unidecode row's fixed contribution to
// doctor's line-count assertions, storeRowLines's own precedent: it
// prints exactly one line regardless of state.
const unidecodeRowLines = 1

// authStoreRowLines is the auth store row's fixed contribution to
// doctor's line-count assertions, storeRowLines's own precedent: it
// prints exactly one line regardless of state.
const authStoreRowLines = 1

// probeRowLines is the four boundary rows' fixed contribution to
// doctor's line-count assertions (client ip, web, bind addresses,
// boundary probe), each printing exactly one line whatever its state.
// It is a literal, never taken from the rows' own list, so a row
// dropped from doctor changes the count a test expects.
const probeRowLines = 4

// musicVolumeRowLines is the music volume row's fixed contribution to
// doctor's line-count assertions, storeRowLines's own precedent.
const musicVolumeRowLines = 1

func TestDoctorCmd(t *testing.T) {
	tests := []struct {
		name         string
		environ      []string
		wantErr      bool
		rowOverrides map[string]string // owned row name -> expected status, overriding "ok"
	}{
		{
			name:    "full valid wall",
			environ: baseEnviron(),
			wantErr: false,
		},
		{
			name:    "missing two keys",
			environ: dropKey(dropKey(baseEnviron(), "DUBPLATE_LIBRARY_ROOT"), "DUBPLATE_SITE_NAME"),
			wantErr: true,
			rowOverrides: map[string]string{
				"DUBPLATE_LIBRARY_ROOT": "missing",
				"DUBPLATE_SITE_NAME":    "missing",
			},
		},
		{
			name: "bad ADMIN_EMAIL",
			environ: append(dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
				"DUBPLATE_ADMIN_EMAIL=not-an-address"),
			wantErr: true,
			rowOverrides: map[string]string{
				"DUBPLATE_ADMIN_EMAIL": "invalid: ",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The store row doctor now always prints reads
			// DUBPLATE_STATE_DIR from environ itself, so every case
			// here points it at a fresh t.TempDir() rather than
			// baseEnviron's literal host path. The unidecode row now
			// prints unconditionally too, so every case installs a
			// table into that same directory.
			stateDir := t.TempDir()
			installTestUnidecodeTable(t, stateDir)
			environ := replaceKey(tt.environ, "DUBPLATE_STATE_DIR", stateDir)

			reg := freshRegistry()

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
			cmd.SetOut(&stdout)
			err := cmd.Execute()

			// The same statuses map doctor's own RunE built from the
			// same Load call, so the threshold expectation below
			// reflects a relational refusal exactly as the row under
			// test would.
			_, loadErr := config.Load(environ, reg)
			statuses := rowStatuses(loadErr)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Execute() error = nil, want error")
				}
				if got := err.Error(); got != "configuration is incomplete" {
					t.Errorf("Execute() error = %q, want %q", got, "configuration is incomplete")
				}
			} else if err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}
			if got := stderr.String(); got != "" {
				t.Errorf("stderr = %q, want empty: doctor's table is stdout-only, main prints the returned error", got)
			}

			lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
			li := 0

			for _, row := range reg.OwnedWall() {
				want := cmp.Or(tt.rowOverrides[row.Name], "ok")
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before owned row %s", row.Name)
				}
				gotLine := lines[li]
				wantPrefix := fmt.Sprintf("%-32s %s", row.Name, want)
				if !strings.HasPrefix(gotLine, wantPrefix) {
					t.Errorf("owned row %d = %q, want prefix %q", li, gotLine, wantPrefix)
				}
				li++
			}

			for _, class := range reg.ReservedClasses() {
				wantLine := fmt.Sprintf("%-32s disabled (arrives at rung %d)", class.Name, class.Rung)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before reserved class %s", class.Name)
				}
				if lines[li] != wantLine {
					t.Errorf("reserved row %d = %q, want %q", li, lines[li], wantLine)
				}
				li++
			}

			for _, class := range reg.IntegrationClasses() {
				status := "disabled"
				if enabled, err := class.Enabled(environ); err == nil && enabled {
					status = "enabled"
				}
				wantLine := fmt.Sprintf("%-32s %s", class.Name, status)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before integration class %s", class.Name)
				}
				if lines[li] != wantLine {
					t.Errorf("integration row %d = %q, want %q", li, lines[li], wantLine)
				}
				li++
			}

			for _, th := range reg.Thresholds() {
				want := thresholdStatus(environ, statuses, th)
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before threshold row %s", th.Key)
				}
				wantPrefix := fmt.Sprintf("%-32s %s", th.Key, want)
				if !strings.HasPrefix(lines[li], wantPrefix) {
					t.Errorf("threshold row %d = %q, want prefix %q", li, lines[li], wantPrefix)
				}
				li++
			}

			// The store row prints unconditionally, for every state
			// row.
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the store row")
			}
			if wantPrefix := fmt.Sprintf("%-32s", "store"); !strings.HasPrefix(lines[li], wantPrefix) {
				t.Errorf("store row %d = %q, want prefix %q", li, lines[li], wantPrefix)
			}
			li++

			// The members row follows the store row, also
			// unconditionally; these cases have no store yet.
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the members row")
			}
			if want := fmt.Sprintf("%-32s %s", membersDoctorRow, "no store yet; start serve, then member add --admin"); lines[li] != want {
				t.Errorf("members row %d = %q, want %q", li, lines[li], want)
			}
			li++

			// The auth store row follows the members row, also
			// unconditionally.
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the auth store row")
			}
			if wantPrefix := fmt.Sprintf("%-32s", authStoreDoctorRow); !strings.HasPrefix(lines[li], wantPrefix) {
				t.Errorf("auth store row %d = %q, want prefix %q", li, lines[li], wantPrefix)
			}
			li++

			// The backup rows follow the auth store row: none of
			// these cases sets DUBPLATE_BACKUP_ENABLED, and the fresh
			// state directory holds no store, so every one reads
			// "not checked" or "unavailable".
			for _, key := range []string{backupLastSuccessRow, backupWALSizeRow, backupFreeSpaceRow, backupLocalCopiesRow, backupRemoteRow} {
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before backup row %s", key)
				}
				if wantPrefix := fmt.Sprintf("%-32s", key); !strings.HasPrefix(lines[li], wantPrefix) {
					t.Errorf("backup row %d = %q, want prefix %q", li, lines[li], wantPrefix)
				}
				li++
			}

			// The unidecode row follows the backup rows, always
			// printing exactly one line; every case here installed a
			// table into the state directory above, so it reads "ok".
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the unidecode row")
			}
			if wantPrefix := fmt.Sprintf("%-32s", unidecodeDoctorRow); !strings.HasPrefix(lines[li], wantPrefix) {
				t.Errorf("unidecode row %d = %q, want prefix %q", li, lines[li], wantPrefix)
			}
			li++

			// The four boundary rows follow the unidecode row.
			for _, key := range []string{clientIPDoctorRow, webDoctorRow, bindAddrsDoctorRow, boundaryProbeDoctorRow} {
				if li >= len(lines) {
					t.Fatalf("stdout ran out of lines before boundary row %s", key)
				}
				if wantPrefix := fmt.Sprintf("%-32s", key); !strings.HasPrefix(lines[li], wantPrefix) {
					t.Errorf("boundary row %d = %q, want prefix %q", li, lines[li], wantPrefix)
				}
				li++
			}

			// The music volume row closes the fixed-count rows.
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the music volume row")
			}
			if wantPrefix := fmt.Sprintf("%-32s", musicVolumeDoctorRow); !strings.HasPrefix(lines[li], wantPrefix) {
				t.Errorf("music volume row %d = %q, want prefix %q", li, lines[li], wantPrefix)
			}
			li++

			if li != len(lines) {
				t.Errorf("stdout has %d extra line(s) beyond owned+reserved+threshold+store+auth store+backup rows: %v", len(lines)-li, lines[li:])
			}
		})
	}
}

// TestDoctorCmd_LineCountFollowsRegistry proves doctor's derivation in
// the direction TestDoctorCmd's own line-count assertion (preserved
// unchanged above, at its "stdout has %d extra line(s)" check) cannot
// reach. TestDoctorCmd re-derives its expectation from the same
// registry it hands to newDoctorCmd, so it proves the two agree, not
// that adding a declaration moves the count: a family rung 1 never
// populates (rung 1 registers no threshold and no integration class of
// its own) could never fail that assertion. This test starts from a
// fresh registry and adds one declaration to each of the four families
// in turn, on scratch DUBPLATE_TEST_SCRATCH_* names, proving each
// addition costs exactly one doctor line, printed in that family's own
// block: one per owned wall row, one per reserved class regardless of
// its key count, one per integration class, one per threshold, and one
// per declared identity source regardless of how many keys that source
// requires.
//
// ruled, conductor 2026-09-07, R10
func TestDoctorCmd_LineCountFollowsRegistry(t *testing.T) {
	reg := freshRegistry()
	lines := doctorLines(t, reg)

	ownedCount := len(reg.OwnedWall())
	reservedCount := len(reg.ReservedClasses())
	integrationCount := len(reg.IntegrationClasses())
	thresholdCount := len(reg.Thresholds())
	sourceCount := len(reg.IdentitySources())
	// storeRowLines, membersRowLines, authStoreRowLines, and
	// backupDoctorRowsOff are the store row, the members row, the auth
	// store row, and the backup rows: each
	// prints unconditionally, so none is one of reg's own declared
	// families. doctorLines's own environ
	// (baseEnviron) never sets DUBPLATE_BACKUP_ENABLED.
	wantTotal := ownedCount + reservedCount + integrationCount + thresholdCount + sourceCount + storeRowLines + membersRowLines + authStoreRowLines + backupDoctorRowsOff + unidecodeRowLines + probeRowLines + musicVolumeRowLines
	if len(lines) != wantTotal {
		t.Fatalf("doctor printed %d lines against a fresh registry, want %d (owned %d + reserved %d + integration %d + threshold %d + source %d + store %d + members %d + auth store %d + backup %d + unidecode %d + boundary %d + music volume %d)",
			len(lines), wantTotal, ownedCount, reservedCount, integrationCount, thresholdCount, sourceCount,
			storeRowLines, membersRowLines, authStoreRowLines, backupDoctorRowsOff, unidecodeRowLines, probeRowLines, musicVolumeRowLines)
	}

	// Owned wall: one line per declared row.
	reg.AddWallRow(config.WallRow{Name: "DUBPLATE_TEST_SCRATCH_OWNED", Description: testDescription, Mark: config.NotSensitive})
	lines = doctorLines(t, reg)
	ownedCount++
	wantTotal++
	if len(lines) != wantTotal {
		t.Fatalf("adding one owned wall row: doctor printed %d lines, want %d", len(lines), wantTotal)
	}
	if got, wantPrefix := lines[ownedCount-1], "DUBPLATE_TEST_SCRATCH_OWNED"; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("owned block's last line = %q, want it to start with %q (own family's block)", got, wantPrefix)
	}

	// Reserved: one line per class, regardless of how many keys the
	// class carries.
	reg.AddReservedClass(config.ReservedClass{
		Name: "scratch reserved",
		Rung: 99,
		Keys: []string{
			"DUBPLATE_TEST_SCRATCH_RESERVED_A",
			"DUBPLATE_TEST_SCRATCH_RESERVED_B",
			"DUBPLATE_TEST_SCRATCH_RESERVED_C",
		},
	})
	lines = doctorLines(t, reg)
	reservedCount++
	wantTotal++
	if len(lines) != wantTotal {
		t.Fatalf("adding one reserved class of three keys: doctor printed %d lines, want %d (one line per class, not per key)", len(lines), wantTotal)
	}
	if got, wantPrefix := lines[ownedCount+reservedCount-1], "scratch reserved"; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("reserved block's last line = %q, want it to start with %q (own family's block)", got, wantPrefix)
	}

	// Integration: one line per class.
	reg.AddIntegrationClass(config.IntegrationClass{
		Name:     "scratch integration",
		Flag:     testKey("DUBPLATE_TEST_SCRATCH_FLAG"),
		Required: []config.KeyDecl{testKey("DUBPLATE_TEST_SCRATCH_REQUIRED")},
	})
	lines = doctorLines(t, reg)
	integrationCount++
	wantTotal++
	if len(lines) != wantTotal {
		t.Fatalf("adding one integration class: doctor printed %d lines, want %d", len(lines), wantTotal)
	}
	if got, wantPrefix := lines[ownedCount+reservedCount+integrationCount-1], "scratch integration"; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("integration block's last line = %q, want it to start with %q (own family's block)", got, wantPrefix)
	}

	// Thresholds: one line per threshold.
	reg.AddThreshold(config.Threshold{Key: "DUBPLATE_TEST_SCRATCH_THRESH", Description: testDescription, Mark: config.NotSensitive, Default: "1"})
	lines = doctorLines(t, reg)
	thresholdCount++
	wantTotal++
	if len(lines) != wantTotal {
		t.Fatalf("adding one threshold: doctor printed %d lines, want %d", len(lines), wantTotal)
	}
	if got, wantPrefix := lines[ownedCount+reservedCount+integrationCount+thresholdCount-1], "DUBPLATE_TEST_SCRATCH_THRESH"; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("threshold block's last line = %q, want it to start with %q (own family's block)", got, wantPrefix)
	}

	// Identity sources: one line per declared source, whatever its key
	// count. Without this block the family passes the two assertions
	// above vacuously, because a fresh registry declares no source.
	reg.AddIdentitySource(config.IdentitySource{
		Name:     "scratch source",
		Selector: "DUBPLATE_TEST_SCRATCH_SELECTOR",
		Required: []config.KeyDecl{testKey("DUBPLATE_TEST_SCRATCH_SOURCE_A"), testKey("DUBPLATE_TEST_SCRATCH_SOURCE_B")},
	})
	lines = doctorLines(t, reg)
	sourceCount++
	wantTotal++
	if len(lines) != wantTotal {
		t.Fatalf("adding one identity source of two keys: doctor printed %d lines, want %d (one line per source, not per key)", len(lines), wantTotal)
	}
	if got, wantPrefix := lines[ownedCount+reservedCount+integrationCount+thresholdCount+sourceCount-1], "scratch source"; !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("identity source block's last line = %q, want it to start with %q (own family's block)", got, wantPrefix)
	}
}

// TestDoctorCmd_BackupRowsWithFlagOn proves backupDoctorRowsOn's own
// total against a full doctor run with DUBPLATE_BACKUP_ENABLED set and
// a valid, fully-credentialed remote (through the rclone stand-in, so
// the features probe never needs a real backend): the same five backup
// rows print as with the flag off, and the "backup remote" row's own
// content changes from "unavailable" to the probe's own answer.
func TestDoctorCmd_BackupRowsWithFlagOn(t *testing.T) {
	useBackupStandin(t)
	t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":true}}`)

	environ := append(doctorEnviron(t),
		"DUBPLATE_BACKUP_ENABLED=true",
		"DUBPLATE_BACKUP_REMOTE=name:"+t.TempDir(),
		"RCLONE_CONFIG_NAME_TYPE=local",
	)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	wantTotal := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + membersRowLines + authStoreRowLines + backupDoctorRowsOn + unidecodeRowLines + probeRowLines + musicVolumeRowLines + len(reg.Settings())
	if len(lines) != wantTotal {
		t.Errorf("stdout has %d line(s), want %d", len(lines), wantTotal)
	}

	wantPrefix := fmt.Sprintf("%-32s", backupRemoteRow)
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, wantPrefix) {
			found = true
			if strings.Contains(line, "unavailable") {
				t.Errorf("backup remote row = %q, want it to run the checks now the flag is on, not report unavailable", line)
			}
		}
	}
	if !found {
		t.Fatalf("stdout has no %q row: %q", backupRemoteRow, stdout.String())
	}
}

// TestDoctorCmd_IdentitySourceRows proves the per-source row's three
// states and its silence about values: the selected source reports
// active when its required rows resolve and "unsatisfied: " plus the
// failing key when one does not, every other declared source reports
// inactive, and no row carries a resolved value, because a source's
// required set carries secrets (Resolution R-5b6).
func TestDoctorCmd_IdentitySourceRows(t *testing.T) {
	const (
		selector  = "DUBPLATE_TEST_SCRATCH_SELECTOR"
		requiredA = "DUBPLATE_TEST_SCRATCH_A_REQUIRED"
		requiredB = "DUBPLATE_TEST_SCRATCH_B_REQUIRED"
		valueA    = "scratch-fixture-value"
	)

	registry := func() *config.Registry {
		reg := freshRegistry()
		reg.AddIdentitySource(config.IdentitySource{Name: "scratch-a", Selector: selector, Required: []config.KeyDecl{testKey(requiredA)}})
		reg.AddIdentitySource(config.IdentitySource{Name: "scratch-b", Selector: selector, Required: []config.KeyDecl{testKey(requiredB)}})
		reg.AddWallRow(config.WallRow{Name: selector, Description: testDescription, Mark: config.NotSensitive})
		return reg
	}

	tests := []struct {
		name      string
		extra     []string
		wantLines []string
	}{
		{
			name:  "the selected source with every required row set is active",
			extra: []string{selector + "=scratch-a", requiredA + "=" + valueA},
			wantLines: []string{
				fmt.Sprintf("%-32s %s", "scratch-a", "active"),
				fmt.Sprintf("%-32s %s", "scratch-b", "inactive"),
			},
		},
		{
			name:  "the selected source with a required row missing is unsatisfied and names the key",
			extra: []string{selector + "=scratch-a"},
			wantLines: []string{
				fmt.Sprintf("%-32s %s", "scratch-a", "unsatisfied: "+requiredA+" missing"),
				fmt.Sprintf("%-32s %s", "scratch-b", "inactive"),
			},
		},
		{
			name:  "no source is selected: every source is inactive",
			extra: []string{selector + "=neither"},
			wantLines: []string{
				fmt.Sprintf("%-32s %s", "scratch-a", "inactive"),
				fmt.Sprintf("%-32s %s", "scratch-b", "inactive"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())
			environ = append(environ, tt.extra...)

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, registry(), &stdout, &stderr)
			_ = cmd.Execute()

			out := stdout.String()
			for _, want := range tt.wantLines {
				if !strings.Contains(out, want+"\n") {
					t.Errorf("stdout = %q, want it to contain the line %q", out, want)
				}
			}
			if strings.Contains(out, valueA) {
				t.Errorf("stdout = %q, want no resolved value: a source's required set carries secrets", out)
			}
		})
	}
}

// doctorKeyStatus returns the status doctor's output out prints for
// the row keyed key, whose name may run past the 32-column pad, and
// reports whether exactly one such row exists.
func doctorKeyStatus(out, key string) (string, bool) {
	var found []string
	for line := range strings.SplitSeq(out, "\n") {
		if rest, ok := strings.CutPrefix(line, key+" "); ok {
			found = append(found, strings.TrimLeft(rest, " "))
		}
	}
	if len(found) != 1 {
		return "", false
	}
	return found[0], true
}

// TestDoctorCmd_AuthSettingRefusals proves an auth setting serve's
// seed would refuse fails doctor on that setting's own row, with no
// store yet, and leaves the magic-link source row active, since the
// bounds and both relations are the seed rule's, not the source's
// hook. A relation's refusal reaches each key it names. The last case
// is review finding 12's repro, a ten-year lifetime beside a one-hour
// cap, which the lifetime's own ceiling now refuses first.
func TestDoctorCmd_AuthSettingRefusals(t *testing.T) {
	tests := []struct {
		name string
		env  []string
		// want maps each refusing key to every part its row must name.
		want map[string][]string
	}{
		{
			name: "retention below the expiry",
			env:  []string{"DUBPLATE_MAGIC_LINK_EXPIRY=45m", "DUBPLATE_NONCE_RETENTION=30m"},
			want: map[string][]string{
				"DUBPLATE_MAGIC_LINK_EXPIRY": {"refused:", "DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"},
				"DUBPLATE_NONCE_RETENTION":   {"refused:", "DUBPLATE_NONCE_RETENTION", "DUBPLATE_MAGIC_LINK_EXPIRY"},
			},
		},
		{
			name: "lifetime above the cap",
			env:  []string{"DUBPLATE_SESSION_LIFETIME=2h", "DUBPLATE_SESSION_ABSOLUTE_CAP=1h", "DUBPLATE_REAUTH_WINDOW=30m"},
			want: map[string][]string{
				"DUBPLATE_SESSION_LIFETIME":     {"refused:", "DUBPLATE_SESSION_LIFETIME", "DUBPLATE_SESSION_ABSOLUTE_CAP"},
				"DUBPLATE_SESSION_ABSOLUTE_CAP": {"refused:", "DUBPLATE_SESSION_LIFETIME", "DUBPLATE_SESSION_ABSOLUTE_CAP"},
			},
		},
		{
			name: "a ten-year lifetime beside a one-hour cap",
			env:  []string{"DUBPLATE_SESSION_LIFETIME=87600h", "DUBPLATE_SESSION_ABSOLUTE_CAP=1h", "DUBPLATE_REAUTH_WINDOW=30m"},
			want: map[string][]string{
				"DUBPLATE_SESSION_LIFETIME": {"refused:", "DUBPLATE_SESSION_LIFETIME", "above the ceiling 17520h"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(doctorEnviron(t), tt.env...)
			var stdout, stderr bytes.Buffer
			if err := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr).Execute(); err == nil {
				t.Fatal("Execute() error = nil, want a refusal")
			}
			out := stdout.String()
			for key, parts := range tt.want {
				status, ok := doctorKeyStatus(out, key)
				if !ok {
					t.Fatalf("stdout = %q, want exactly one row keyed %s", out, key)
				}
				for _, part := range parts {
					if !strings.Contains(status, part) {
						t.Errorf("%s row = %q, want it to contain %q", key, status, part)
					}
				}
			}
			if want := fmt.Sprintf("%-32s %s\n", "magic-link", "active"); !strings.Contains(out, want) {
				t.Errorf("stdout = %q, want the source row %q", out, want)
			}
		})
	}
}

// TestRowStatuses_FlattensNestedJoin proves the review M4 flatten fix
// against a synthetic tree rowStatuses's own package tests never
// otherwise reach: an outer errors.Join wrapping a single inner
// errors.Join, mixing a config.ErrMissing leaf with an ordinary
// invalid leaf. errors.Is walks the whole tree regardless of nesting
// depth, so a rowStatuses that unwraps only one level classifies the
// entire inner join as "missing" and B_KEY's own "invalid: bad" reason
// never surfaces (mutation 1).
func TestRowStatuses_FlattensNestedJoin(t *testing.T) {
	err := errors.Join(errors.Join(
		fmt.Errorf("A_KEY: %w", config.ErrMissing),
		errors.New("B_KEY: bad"),
	))

	statuses := rowStatuses(err)

	if got := statuses["A_KEY"]; got != "missing" {
		t.Errorf(`statuses["A_KEY"] = %q, want "missing"`, got)
	}
	if got := statuses["B_KEY"]; got != "invalid: bad" {
		t.Errorf(`statuses["B_KEY"] = %q, want "invalid: bad"`, got)
	}
}

// TestDoctorCmd_RowStatusesSurviveSimultaneousFailures proves the
// review M4 flatten fix directly: two of the identity source's own
// Required keys failing at once must each surface as their own row
// status, not one entry naming the first and swallowing the second's
// reason inside a stray newline.
func TestDoctorCmd_RowStatusesSurviveSimultaneousFailures(t *testing.T) {
	environ := append(doctorEnviron(t),
		"DUBPLATE_MAGIC_LINK_SECRET="+fixtureSecretBelowFloor,
		"DUBPLATE_SESSION_SECRET="+fixtureSecretBelowFloor,
	)

	_, loadErr := config.Load(environ, productionRegistry())
	if loadErr == nil {
		t.Fatal("Load() error = nil, want a refusal: both secrets are below the floor")
	}
	statuses := rowStatuses(loadErr)

	for _, key := range []string{"DUBPLATE_MAGIC_LINK_SECRET", "DUBPLATE_SESSION_SECRET"} {
		got, ok := statuses[key]
		if !ok {
			t.Errorf("statuses[%s] missing, want an entry naming its own failure", key)
			continue
		}
		if !strings.HasPrefix(got, "invalid: too short") {
			t.Errorf("statuses[%s] = %q, want it to start with %q", key, got, "invalid: too short")
		}
		if strings.Contains(got, "\n") {
			t.Errorf("statuses[%s] = %q, carries another row's line inside its own reason", key, got)
		}
	}
}

// TestDoctorCmd_ShortPreviousSecretMakesSourceUnsatisfied proves the
// review M4 Optional-consultation fix: a short rotation-overlap secret
// is one of magic-link's Optional keys, and with every Required key
// satisfied it is the only thing wrong, yet the source's own row must
// still read unsatisfied rather than active.
func TestDoctorCmd_ShortPreviousSecretMakesSourceUnsatisfied(t *testing.T) {
	environ := append(doctorEnviron(t), "DUBPLATE_SESSION_SECRET_PREVIOUS="+fixtureSecretBelowFloor)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want a refusal: the previous secret is below the floor")
	}

	want := fmt.Sprintf("%-32s %s", "magic-link", "unsatisfied")
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want a line with prefix %q", stdout.String(), want)
	}
}

// TestDoctorCmd_SelectedSourcePrintsOneLine proves doctor prints
// exactly one line for the selected identity source even when two of
// its Required keys fail at once: the existing selected-source tests
// use strings.Contains against the magic-link row, so they pass just
// as well against a doctor that printed one extra line per failing
// key rather than one line per source (review finding,
// doctor_test.go:433). Both DUBPLATE_MAGIC_LINK_SECRET and
// DUBPLATE_SESSION_SECRET are magic-link's own Required keys
// (auth.go), so this reuses the exact simultaneous-failure fixture
// TestDoctorCmd_RowStatusesSurviveSimultaneousFailures already proves
// rowStatuses keeps as two separate entries.
func TestDoctorCmd_SelectedSourcePrintsOneLine(t *testing.T) {
	environ := append(doctorEnviron(t),
		"DUBPLATE_MAGIC_LINK_SECRET="+fixtureSecretBelowFloor,
		"DUBPLATE_SESSION_SECRET="+fixtureSecretBelowFloor,
	)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want a refusal: both secrets are below the floor")
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")

	matches := 0
	for _, line := range lines {
		if strings.HasPrefix(line, "magic-link") {
			matches++
		}
	}
	if matches != 1 {
		t.Errorf("stdout has %d line(s) starting with %q, want exactly 1: one line per source, not per key", matches, "magic-link")
	}

	// storeRowLines and backupDoctorRowsOff are the store row and the
	// backup rows: doctorEnviron loads cleanly, so no unclaimed-error
	// line prints, and it never sets DUBPLATE_BACKUP_ENABLED.
	wantTotal := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + membersRowLines + authStoreRowLines + backupDoctorRowsOff + unidecodeRowLines + probeRowLines + musicVolumeRowLines + len(reg.Settings())
	if len(lines) != wantTotal {
		t.Errorf("stdout has %d line(s), want %d: one line per source, not per key", len(lines), wantTotal)
	}
}

// TestDoctorCmd_SourceLineNamesFailingKey proves R-5c10: the selected
// source's own line names the offending key, key name only and never a
// value, for each of the review's three cases: a short rotation-overlap
// secret, a short primary secret, and a public URL carrying userinfo.
func TestDoctorCmd_SourceLineNamesFailingKey(t *testing.T) {
	tests := []struct {
		name    string
		extra   []string
		wantKey string
	}{
		{
			name:    "short rotation-overlap secret",
			extra:   []string{"DUBPLATE_SESSION_SECRET_PREVIOUS=" + fixtureSecretBelowFloor},
			wantKey: "DUBPLATE_SESSION_SECRET_PREVIOUS",
		},
		{
			name:    "short primary secret",
			extra:   []string{"DUBPLATE_MAGIC_LINK_SECRET=" + fixtureSecretBelowFloor},
			wantKey: "DUBPLATE_MAGIC_LINK_SECRET",
		},
		{
			name:    "public URL carrying userinfo",
			extra:   []string{"DUBPLATE_PUBLIC_URL=https://operator:fixture-pass-not-real@dubplate.example.test"},
			wantKey: "DUBPLATE_PUBLIC_URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(doctorEnviron(t), tt.extra...)
			reg := productionRegistry()

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
			if err := cmd.Execute(); err == nil {
				t.Fatal("Execute() error = nil, want a refusal")
			}

			want := fmt.Sprintf("%-32s unsatisfied: %s invalid", "magic-link", tt.wantKey)
			out := stdout.String()
			if !strings.Contains(out, want) {
				t.Errorf("stdout = %q, want it to contain %q", out, want)
			}
			for _, secret := range []string{fixtureSecretBelowFloor, "fixture-pass-not-real"} {
				if strings.Contains(out, secret) {
					t.Errorf("stdout = %q, want no secret value present", out)
				}
			}
		})
	}
}

// TestDoctorCmd_SourceLineNamesTwoFailingKeysAtOnce proves R-5c10's
// acceptance criterion that two failing keys at once both appear on the
// one line: both magic-link secrets below the floor, on the fixture
// TestDoctorCmd_RowStatusesSurviveSimultaneousFailures already proves
// rowStatuses keeps as two separate entries. Fixture state: the source
// is selected (doctorEnviron), so the line prints at all.
func TestDoctorCmd_SourceLineNamesTwoFailingKeysAtOnce(t *testing.T) {
	environ := append(doctorEnviron(t),
		"DUBPLATE_MAGIC_LINK_SECRET="+fixtureSecretBelowFloor,
		"DUBPLATE_SESSION_SECRET="+fixtureSecretBelowFloor,
	)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want a refusal: both secrets are below the floor")
	}

	out := stdout.String()
	for _, key := range []string{"DUBPLATE_MAGIC_LINK_SECRET", "DUBPLATE_SESSION_SECRET"} {
		if !strings.Contains(out, key+" invalid") {
			t.Errorf("stdout = %q, want it to name %s", out, key)
		}
	}

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	found := false
	for _, line := range lines {
		if strings.HasPrefix(line, "magic-link") {
			found = true
			if !strings.Contains(line, "DUBPLATE_MAGIC_LINK_SECRET invalid") || !strings.Contains(line, "DUBPLATE_SESSION_SECRET invalid") {
				t.Errorf("magic-link line = %q, want both failing keys on the one line", line)
			}
		}
	}
	if !found {
		t.Fatal("no line starts with magic-link")
	}
}

// TestIdentitySourceStatus_KeyFromKeyedTypeNotMessagePrefix proves
// R-5c10's mechanism: doctor reads a Validate refusal's failing key from
// [sourceKeyError]'s own field via errors.As, never by parsing the
// rendered message's prefix. The fixture wraps [sourceKeyError] behind
// an outer fmt.Errorf, so the leaf error's rendered text carries a
// colon-bearing prefix before the real key's own colon: a doctor that
// cut the message at its first or its last ": " would read that outer
// prefix, not the real key.
func TestIdentitySourceStatus_KeyFromKeyedTypeNotMessagePrefix(t *testing.T) {
	const (
		selector = "DUBPLATE_TEST_SCRATCH_SELECTOR"
		key      = "DUBPLATE_TEST_SCRATCH_KEY"
	)
	src := config.IdentitySource{
		Name:     "scratch",
		Selector: selector,
		Validate: func(map[string]string, []string) error {
			return fmt.Errorf("scratch refusal: %w", &sourceKeyError{key: key, reason: "not really used"})
		},
	}
	environ := []string{selector + "=scratch"}

	got := identitySourceStatus(environ, src, map[string]string{})
	want := "unsatisfied: " + key + " invalid"
	if got != want {
		t.Errorf("identitySourceStatus() = %q, want %q", got, want)
	}
}

// TestDoctorCmd_MembersRow is the members row's golden over every
// state it renders, through the real inspect mode and doctor's own exit
// status. Fixture state: the no-admin store holds a tombstoned admin
// beside a live non-admin, so a count that includes tombstoned rows
// reads one admin and passes; the passing store holds a tombstoned
// member and a tombstoned admin beside its live rows, so the counts
// show only the live ones; the two older stores come from the helper
// at version 2 and at version 1, the second of which carries no facts,
// so a rule keyed on Facts rather than on the report's StoreVersion
// reads it as "not read" and fails the version 1 row.
func TestDoctorCmd_MembersRow(t *testing.T) {
	head := store.Inspect(t.Context(), store.InspectOptions{StateDir: t.TempDir(), BusyTimeout: testBackupBusy}).BinaryVersion
	reconcilePending := `{"format": 1, "state": "reconcile-pending", "stamp": "20260925T120000.000000Z",
"source": {"remote": "remote:dubplate", "snapshot": "20260924T030000.000000Z"},
"incoming": "restore-20260925T120000.000000Z.db", "pre_restore": "pre-restore-20260925T120000.000000Z.db",
"discard_current": false, "library": null, "version": 2}`
	live := func(address, username string, admin bool) seedMember {
		return seedMember{address: address, displayName: "Name", username: username, admin: admin}
	}
	gone := func(address, username string, admin bool) seedMember {
		return seedMember{address: address, displayName: "Name", username: username, admin: admin, tombstoned: true}
	}
	tests := []struct {
		name     string
		setup    func(t *testing.T, stateDir string)
		want     string
		wantFail bool
	}{
		{
			name:  "no store yet",
			setup: func(*testing.T, string) {},
			want:  "no store yet; start serve, then member add --admin",
		},
		{
			name:  "store version 1, which carries no facts",
			setup: func(t *testing.T, stateDir string) { olderStore(t, stateDir, 1) },
			want:  "members not recorded before store version 3",
		},
		{
			name:  "store version 2, which carries facts without the users table",
			setup: func(t *testing.T, stateDir string) { olderStore(t, stateDir, 2) },
			want:  "members not recorded before store version 3",
		},
		{
			name:     "current, no live member",
			setup:    func(t *testing.T, stateDir string) { seedBackupStore(t, stateDir) },
			want:     "no members; run dubplate-server member add --admin",
			wantFail: true,
		},
		{
			name: "live members and only a tombstoned admin",
			setup: func(t *testing.T, stateDir string) {
				seedBackupStore(t, stateDir)
				seedMembers(t, stateDir, gone("old@example.test", "old", true), live("plain@example.test", "plain", false))
			},
			want:     "members: 1, admins: 0; run dubplate-server member add --admin",
			wantFail: true,
		},
		{
			name: "live members with a live admin",
			setup: func(t *testing.T, stateDir string) {
				seedBackupStore(t, stateDir)
				seedMembers(t, stateDir,
					live("boss@example.test", "boss", true), live("plain@example.test", "plain", false),
					gone("old@example.test", "old", true), gone("former@example.test", "former", false))
			},
			want: "members: 2, admins: 1",
		},
		{
			name: "older at version 3 holding the users table",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, 3)
				execStoreSQL(t, path, `INSERT INTO users (id, address, display_name, username, admin, created_at, updated_at)
					VALUES (x'f001', 'boss@example.test', 'Name', 'boss', 1, 1, 1)`)
			},
			want: "members: 1, admins: 1",
		},
		{
			name: "newer than the binary",
			setup: func(t *testing.T, stateDir string) {
				execStoreSQL(t, olderStore(t, stateDir, head), fmt.Sprintf("PRAGMA user_version = %d", head+1))
			},
			want:     "not read",
			wantFail: true,
		},
		{
			name: "a restore-intent record",
			setup: func(t *testing.T, stateDir string) {
				seedBackupStore(t, stateDir)
				if err := os.WriteFile(filepath.Join(stateDir, "store", "restore-intent"), []byte(reconcilePending), 0o600); err != nil {
					t.Fatalf("write the restore-intent record: %v", err)
				}
			},
			want:     "not read",
			wantFail: true,
		},
		{
			name:     "not a store",
			setup:    buildZeroByteStoreLeaf,
			want:     "not read",
			wantFail: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := doctorEnviron(t)
			tt.setup(t, environValue(environ, "DUBPLATE_STATE_DIR"))

			rows, _, err := doctorRows(t, environ)
			if got := rows["members"]; got != tt.want {
				t.Errorf("members row = %q, want %q", got, tt.want)
			}
			if tt.wantFail != (err != nil) {
				t.Errorf("Execute() error = %v, want a failure = %v", err, tt.wantFail)
			}
		})
	}

	t.Run("a busy store", func(t *testing.T) {
		environ := doctorEnviron(t)
		stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
		seedBackupStore(t, stateDir)
		lock, err := os.Open(filepath.Join(stateDir, "store", "store.lock"))
		if err != nil {
			t.Fatalf("open store.lock: %v", err)
		}
		defer func() { _ = lock.Close() }()
		if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
			t.Fatalf("flock store.lock: %v", err)
		}

		rows, _, err := doctorRows(t, environ)
		if got := rows["members"]; got != "not read" {
			t.Errorf("members row = %q, want %q", got, "not read")
		}
		if err == nil {
			t.Error("Execute() error = nil, want the store row's failure")
		}
	})
}

// doctorEnviron returns authEnviron with DUBPLATE_STATE_DIR pointed at
// a fresh t.TempDir() of its own: doctor's store row now inspects that
// directory, and authEnviron's own fixed path is not one a test process
// may create the store leaf's directories under. The directory holds no
// store, so the members row reads "no store yet" and leaves the exit
// alone; a test that creates a store there seeds an admin with
// [seedDoctorAdmin], since doctor fails a store holding no live admin.
func doctorEnviron(t *testing.T) []string {
	t.Helper()
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)
	environ := replaceKey(authEnviron(t), "DUBPLATE_STATE_DIR", stateDir)
	return append(environ, "DUBPLATE_SERVE_PORT="+strconv.Itoa(reservedPort()))
}

// seedDoctorAdmin gives the store under environ's state directory one
// live admin, the state a deployed store is in once the operator has
// run member add --admin.
func seedDoctorAdmin(t *testing.T, environ []string) {
	t.Helper()
	seedMembers(t, environValue(environ, "DUBPLATE_STATE_DIR"),
		seedMember{address: "doctor-admin@example.test", displayName: "Admin", username: "doctoradmin", admin: true})
}

// doctorLines runs newDoctorCmd against reg with a satisfied base
// environ and returns its stdout split into lines. It ignores
// Execute's own error: a scratch owned row this test never sets a
// value for makes the wall incomplete, which is expected and beside
// the point of a line-count assertion.
func doctorLines(t *testing.T, reg *config.Registry) []string {
	t.Helper()
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())
	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	_ = cmd.Execute()
	return strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
}

// TestDoctorCmd_LiteralSpotCheck hard-codes two known lines so the
// table-driven test above cannot go vacuous if OwnedWall ever returned
// an empty table.
func TestDoctorCmd_LiteralSpotCheck(t *testing.T) {
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir)

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, freshRegistry(), &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	want := fmt.Sprintf("%-32s %s\n", "DUBPLATE_ADMIN_EMAIL", "ok")
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
	}

	// The pending "identity source" row this spot check used to pin
	// retired with rung 5 (keys.go's ownedWall no longer carries it);
	// its owned-now replacement, the selector row, is the second
	// literal now. freshRegistry() alone declares no identity
	// source, so this half runs against the production registry and a
	// fully satisfied environment instead.
	var authStdout, authStderr bytes.Buffer
	authCmd := newDoctorCmd(doctorEnviron(t), productionRegistry(), &authStdout, &authStderr)
	if err := authCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	wantIdentity := fmt.Sprintf("%-32s %s\n", "DUBPLATE_IDENTITY_SOURCE", "ok")
	if !strings.Contains(authStdout.String(), wantIdentity) {
		t.Errorf("stdout = %q, want it to contain %q", authStdout.String(), wantIdentity)
	}

	// Reserved classes print in table order, and the table lists
	// pipeline before monitoring, two classes at the same rung, so the
	// table is the only order that separates them: the surviving rungs
	// already ascend, so a rung sort cannot be told apart from the
	// table. This pins the table against a name sort, which would print
	// monitoring first. TestDoctorCmd itself cannot pin this: it builds
	// its expected lines by iterating config.ReservedClasses() itself,
	// so a registry that sorted both the production output and the
	// test's own expectations would still agree. This assertion pins
	// the order against the raw stdout instead.
	out := stdout.String()
	pipelineIdx := strings.Index(out, "pipeline")
	monitoringIdx := strings.Index(out, "monitoring")
	if pipelineIdx < 0 || monitoringIdx < 0 {
		t.Fatalf("stdout missing a reserved class row: pipeline idx %d, monitoring idx %d", pipelineIdx, monitoringIdx)
	}
	if pipelineIdx >= monitoringIdx {
		t.Errorf("pipeline printed at %d, monitoring at %d; want pipeline before monitoring", pipelineIdx, monitoringIdx)
	}
}

// TestDoctorCmd_IntegrationRow proves an integration class's row
// reports its enable flag: disabled until the flag is set, enabled
// once it is. The duration-threshold rendering that rides alongside it
// is proven by TestDoctorCmd_StoreBusyTimeoutDefault.
func TestDoctorCmd_IntegrationRow(t *testing.T) {
	useBackupStandin(t)
	t.Setenv(backupStandinStdoutEnv, `{"Features":{"Copy":false,"Move":true}}`)

	reg := freshRegistry()
	registerBackupConfig(reg)
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())

	tests := []struct {
		name  string
		extra []string
		want  string
	}{
		{name: "flag unset", want: "disabled"},
		{name: "flag set", extra: []string{
			"DUBPLATE_BACKUP_ENABLED=true",
			"DUBPLATE_BACKUP_REMOTE=name:" + t.TempDir(),
			"RCLONE_CONFIG_NAME_TYPE=local",
		}, want: "enabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(append(slices.Clone(environ), tt.extra...), reg, &stdout, &stderr)
			_ = cmd.Execute()

			want := fmt.Sprintf("%-32s %s\n", "backup", tt.want)
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
			}
		})
	}
}

// TestDoctorCmd_FloatThreshold proves a registered KindFloat threshold
// prints its effective value in the thresholds block, in the same
// pinned "%-32s %s" form every other row uses. wantLine is a literal
// ("0.04"), not derived by calling config.ParseFloat itself.
func TestDoctorCmd_FloatThreshold(t *testing.T) {
	reg := freshRegistry()
	reg.AddThreshold(config.Threshold{
		Key:     "DUBPLATE_TEST_STRONG_REC_THRESH",
		Kind:    config.KindFloat,
		Default: "0.04",
		Doc:     "a float between 0 and 1",

		Description: testDescription,
		Mark:        config.NotSensitive,
	})

	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir)

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	wantLine := fmt.Sprintf("%-32s %s\n", "DUBPLATE_TEST_STRONG_REC_THRESH", "0.04")
	if !strings.Contains(stdout.String(), wantLine) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantLine)
	}
}

// TestDoctorCmd_RefusedSettingPrintsOnce proves a setting the seed
// rule refuses reaches the operator on its own setting row alone: the
// unclaimed-load-error list must not repeat a key a row already owns.
func TestDoctorCmd_RefusedSettingPrintsOnce(t *testing.T) {
	const key = "DUBPLATE_SESSION_LIFETIME"
	environ := append(doctorEnviron(t),
		key+"=87600h",
		"DUBPLATE_SESSION_ABSOLUTE_CAP=1h",
	)

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
	if err := cmd.Execute(); err == nil {
		t.Fatal("Execute() error = nil, want a refusal: the lifetime is above its ceiling")
	}

	var named []string
	for line := range strings.SplitSeq(stdout.String(), "\n") {
		if strings.HasPrefix(line, key+" ") {
			named = append(named, line)
		}
	}
	if len(named) != 1 {
		t.Fatalf("rows keyed %s = %q, want exactly one (the setting row)", key, named)
	}
	if !strings.HasPrefix(named[0], fmt.Sprintf("%-32s refused:", key)) {
		t.Errorf("the one row keyed %s = %q, want the setting row reading refused", key, named[0])
	}
}

// TestUnclaimedLoadErrors proves the list holds only DUBPLATE_ keys no
// row owns, in sorted order.
func TestUnclaimedLoadErrors(t *testing.T) {
	reg := freshRegistry()
	wall := reg.OwnedWall()
	if len(wall) == 0 {
		t.Fatal("registry has no owned wall rows to test with")
	}
	owned := wall[0].Name

	statuses := map[string]string{
		"OTHER_PREFIX_KEY":  "refused",
		"DUBPLATE_ZZ_LATE":  "refused",
		"DUBPLATE_AA_EARLY": "refused",
		owned:               "refused",
	}
	got := unclaimedLoadErrors(reg, statuses)
	want := []string{"DUBPLATE_AA_EARLY", "DUBPLATE_ZZ_LATE"}
	if !slices.Equal(got, want) {
		t.Errorf("unclaimedLoadErrors() = %q, want %q", got, want)
	}
}

// TestDoctorCmd_StoreBusyTimeoutDefault proves the threshold row shows
// 15s by default and the set value when set, in the same "%-32s %s"
// form every duration threshold uses.
func TestDoctorCmd_StoreBusyTimeoutDefault(t *testing.T) {
	reg := freshRegistry()
	registerStoreConfig(reg)

	tests := []struct {
		name  string
		extra []string
		want  string
	}{
		{name: "default", want: "15s"},
		{name: "set", extra: []string{"DUBPLATE_STORE_BUSY_TIMEOUT=30s"}, want: "30s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir()), tt.extra...)

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
			_ = cmd.Execute()

			want := fmt.Sprintf("%-32s %s\n", "DUBPLATE_STORE_BUSY_TIMEOUT", tt.want)
			if !strings.Contains(stdout.String(), want) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), want)
			}
		})
	}
}

// TestDoctorCmd_StoreBusyTimeoutOutOfRange proves the states table's
// busy-timeout rows: 0s, -1s, 500us, and 1000h (3600000000ms, the risk
// review's measured over-range value) each leave the store row reading
// "not checked" and naming the key and the bound it crossed, while the
// threshold row still shows the parsed value unchanged, and doctor
// exits 1. 2147483647ms passes both rows.
func TestDoctorCmd_StoreBusyTimeoutOutOfRange(t *testing.T) {
	reg := freshRegistry()
	registerStoreConfig(reg)

	tests := []struct {
		name       string
		value      string
		wantErr    bool
		wantBound  string
		wantValStr string
	}{
		{name: "zero", value: "0s", wantErr: true, wantBound: "at least 1ms", wantValStr: "0s"},
		{name: "negative", value: "-1s", wantErr: true, wantBound: "at least 1ms", wantValStr: "-1s"},
		{name: "below one millisecond", value: "500us", wantErr: true, wantBound: "at least 1ms", wantValStr: "500µs"},
		{name: "far above the ceiling", value: "1000h", wantErr: true, wantBound: "at most 2147483647ms", wantValStr: "1000h0m0s"},
		{name: "at the ceiling", value: "2147483647ms", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			installTestUnidecodeTable(t, stateDir)
			environ := append(replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir),
				"DUBPLATE_STORE_BUSY_TIMEOUT="+tt.value)

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
			err := cmd.Execute()

			if !tt.wantErr {
				if err != nil {
					t.Fatalf("Execute() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("Execute() error = nil, want a refusal: the busy timeout is out of range")
			}

			out := stdout.String()
			wantThreshold := fmt.Sprintf("%-32s %s\n", "DUBPLATE_STORE_BUSY_TIMEOUT", tt.wantValStr)
			if !strings.Contains(out, wantThreshold) {
				t.Errorf("stdout = %q, want it to contain %q (the threshold row shows the parsed value unchanged)", out, wantThreshold)
			}

			wantStore := fmt.Sprintf("%-32s not checked: DUBPLATE_STORE_BUSY_TIMEOUT must be %s\n", "store", tt.wantBound)
			if !strings.Contains(out, wantStore) {
				t.Errorf("stdout = %q, want it to contain %q", out, wantStore)
			}
		})
	}
}

// TestDoctorCmd_AbsentStoreCreatesNothing proves a doctor run against
// an absent store creates nothing under the state directory: the
// directory listing is identical before and after.
func TestDoctorCmd_AbsentStoreCreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir)
	reg := freshRegistry()
	registerStoreConfig(reg)

	before, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir(%s) before: %v", stateDir, err)
	}

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	_ = cmd.Execute()

	after, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir(%s) after: %v", stateDir, err)
	}
	if len(before) != 0 || len(after) != 0 {
		t.Fatalf("state directory listing before = %v, after = %v, want both empty", before, after)
	}
}

// TestDoctorCmd_SettingsRow_ExitCode is Ruling 2's exit-code half, end
// to end: doctor exits nonzero on an admin-conflict row and exits zero
// on a pending re-seed.
func TestDoctorCmd_SettingsRow_ExitCode(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		stored   string
		env      string
		wantFail bool
	}{
		{name: "admin conflict fails", source: "admin", stored: "0.9", env: "0.3", wantFail: true},
		{name: "pending re-seed does not fail", source: "env-seed", stored: "0.6", env: "0.3", wantFail: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := doctorEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			seedBackupStore(t, stateDir)
			seedDoctorAdmin(t, environ)

			h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
			if err != nil {
				t.Fatalf("OpenWrite: %v", err)
			}
			seedSettingRow(t, h, "DUBPLATE_BACKUP_GUARD_FRACTION", tt.stored, tt.source)
			if err := h.Close(); err != nil {
				t.Fatalf("close: %v", err)
			}
			environ = append(environ, "DUBPLATE_BACKUP_GUARD_FRACTION="+tt.env)

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
			err = cmd.Execute()
			if tt.wantFail && err == nil {
				t.Error("Execute() error = nil, want a refusal")
			}
			if !tt.wantFail && err != nil {
				t.Errorf("Execute() error = %v, want nil", err)
			}
		})
	}
}

// TestDoctorCmd_AuthStoreRow proves the auth store row reports, before
// serve starts, each state of the auth leaf serve would refuse or
// could not write: a leaf wider than 0700, in the words auth.Open
// refuses it with; a leaf another uid owns; a record subdirectory the
// process may not write; and an absent leaf whose state directory is
// not writable either. An absent leaf under a writable state directory
// and a well-formed leaf both pass, so the row cannot fail every case.
func TestDoctorCmd_AuthStoreRow(t *testing.T) {
	tests := []struct {
		name string
		// setup shapes the auth leaf under stateDir and returns the
		// status text the row must carry, with the leaf's path
		// substituted where it is named.
		setup      func(t *testing.T, stateDir string) string
		wantFailed bool
	}{
		{
			name:  "an absent leaf under a writable state directory",
			setup: func(*testing.T, string) string { return "absent; serve creates it" },
		},
		{
			name: "a leaf serve built",
			setup: func(t *testing.T, stateDir string) string {
				if _, err := auth.Open(stateDir); err != nil {
					t.Fatalf("fixture state: auth.Open: %v", err)
				}
				return "ok"
			},
		},
		{
			name: "a leaf at mode 0750",
			setup: func(t *testing.T, stateDir string) string {
				leaf := makeAuthLeaf(t, stateDir)
				chmodRestoring(t, leaf, 0o750)
				return fmt.Sprintf("leaf %s has mode 750, want at most 700", leaf)
			},
			wantFailed: true,
		},
		{
			name: "a leaf another uid owns",
			setup: func(t *testing.T, stateDir string) string {
				leaf := makeAuthLeaf(t, stateDir)
				orig := geteuid
				geteuid = func() int { return os.Geteuid() + 1 }
				t.Cleanup(func() { geteuid = orig })
				return fmt.Sprintf("leaf %s is owned by uid %d, want uid %d", leaf, os.Geteuid(), os.Geteuid()+1)
			},
			wantFailed: true,
		},
		{
			name: "a nonce subdirectory the process may not write",
			setup: func(t *testing.T, stateDir string) string {
				leaf := makeAuthLeaf(t, stateDir)
				nonce := filepath.Join(leaf, "nonce")
				chmodRestoring(t, nonce, 0o500)
				return nonce + " is not writable: permission denied"
			},
			wantFailed: true,
		},
		{
			name: "an absent epoch subdirectory under a leaf the process may not write",
			setup: func(t *testing.T, stateDir string) string {
				leaf := makeAuthLeaf(t, stateDir)
				if err := os.Remove(filepath.Join(leaf, "epoch")); err != nil {
					t.Fatalf("fixture state: remove epoch: %v", err)
				}
				chmodRestoring(t, leaf, 0o500)
				return filepath.Join(leaf, "epoch") + " is absent, and " + leaf + " is not writable: permission denied"
			},
			wantFailed: true,
		},
		{
			name: "an absent leaf under a state directory the process may not write",
			setup: func(t *testing.T, stateDir string) string {
				chmodRestoring(t, stateDir, 0o500)
				return "absent, and serve cannot create it: " + stateDir + " is not writable: permission denied"
			},
			wantFailed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := doctorEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			want := tt.setup(t, stateDir)

			rows, _, err := doctorRows(t, environ)
			if got := rows[authStoreDoctorRow]; got != want {
				t.Errorf("auth store row = %q, want %q", got, want)
			}
			if tt.wantFailed && err == nil {
				t.Error("Execute() error = nil, want a refusal")
			}
			if !tt.wantFailed && err != nil {
				t.Errorf("Execute() error = %v, want nil", err)
			}
		})
	}
}

// TestDoctorCmd_AuthStoreRowSpeaksServesRefusal proves the mode
// failure reads in serve's own words: the row's text is the tail of
// the error auth.Open returns for the same leaf, so an operator who
// sees one recognizes the other.
func TestDoctorCmd_AuthStoreRowSpeaksServesRefusal(t *testing.T) {
	environ := doctorEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	leaf := makeAuthLeaf(t, stateDir)
	chmodRestoring(t, leaf, 0o750)

	_, openErr := auth.Open(stateDir)
	if openErr == nil {
		t.Fatal("fixture state: auth.Open accepted a 0750 leaf")
	}
	rows, _, _ := doctorRows(t, environ)
	row := rows[authStoreDoctorRow]
	if row == "" || !strings.HasSuffix(openErr.Error(), row) {
		t.Errorf("auth store row = %q, want the tail of serve's refusal %q", row, openErr)
	}
}

// makeAuthLeaf builds the auth leaf and its two record subdirectories
// under stateDir through auth.Open, the way serve builds them, and
// returns the leaf's path.
func makeAuthLeaf(t *testing.T, stateDir string) string {
	t.Helper()
	if _, err := auth.Open(stateDir); err != nil {
		t.Fatalf("fixture state: auth.Open: %v", err)
	}
	return auth.StoreDir(stateDir)
}

// chmodRestoring sets path's mode and restores 0700 at cleanup, so
// t.TempDir's own removal can still descend into it.
func chmodRestoring(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("fixture state: Chmod %s: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
}

// TestDoctorCmd_ServePortOutOfBound proves doctor fails a serve port
// serve could not bind under its unit, in the threshold's own row,
// naming the capability: before the bound, port 80 read as a plain
// value and the probe as "no listener", and doctor exited 0.
func TestDoctorCmd_ServePortOutOfBound(t *testing.T) {
	for _, port := range []string{"0", "80"} {
		t.Run(port, func(t *testing.T) {
			rows, _, err := doctorRows(t, withEnv(doctorEnviron(t), "DUBPLATE_SERVE_PORT="+port))
			if err == nil {
				t.Error("Execute() error = nil, want a refusal")
			}
			got := rows["DUBPLATE_SERVE_PORT"]
			want := "invalid: " + port + " is below 1024: the serve unit withholds CAP_NET_BIND_SERVICE"
			if !strings.HasPrefix(got, want) {
				t.Errorf("DUBPLATE_SERVE_PORT row = %q, want prefix %q", got, want)
			}
		})
	}
}

// TestDoctorCmd_SourceLinePrintsReasonCarryingNoValue proves the
// selected source's line prints a refusal's reason beside a key that
// has no row of its own, the reason check prints, whenever that reason
// does not quote the configured value: a public URL with a path and a
// short secret each print theirs, while an SMTP port whose reason
// quotes the port prints the key alone.
func TestDoctorCmd_SourceLinePrintsReasonCarryingNoValue(t *testing.T) {
	tests := []struct {
		name     string
		entry    string
		want     string
		wantNone string
	}{
		{
			name:  "a public URL carrying a path",
			entry: "DUBPLATE_PUBLIC_URL=https://dubplate.example.test/music",
			want:  "unsatisfied: DUBPLATE_PUBLIC_URL invalid (must carry no path; the serve prefix sets where dubplate lives)",
		},
		{
			name:     "a short session secret",
			entry:    "DUBPLATE_SESSION_SECRET=" + fixtureSecretBelowFloor,
			want:     "unsatisfied: DUBPLATE_SESSION_SECRET invalid (too short, want at least 32 bytes)",
			wantNone: fixtureSecretBelowFloor,
		},
		{
			name:     "an SMTP port whose reason quotes it",
			entry:    "DUBPLATE_SMTP_PORT=70000",
			want:     "unsatisfied: DUBPLATE_SMTP_PORT invalid",
			wantNone: "70000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, _, err := doctorRows(t, withEnv(doctorEnviron(t), tt.entry))
			if err == nil {
				t.Error("Execute() error = nil, want a refusal")
			}
			got := rows["magic-link"]
			if got != tt.want {
				t.Errorf("magic-link row = %q, want %q", got, tt.want)
			}
			if tt.wantNone != "" && strings.Contains(got, tt.wantNone) {
				t.Errorf("magic-link row = %q, want it to never quote %q", got, tt.wantNone)
			}
		})
	}
}

// TestDoctorCmd_AuthSettingsGolden is spec criterion 10's doctor half:
// after a start seeds every auth key from the environment, each row
// prints the value in force, its env-seed source, its apply mode, and
// the sensitive mark on the link expiry, the session lifetime, and the
// absolute cap, whose values print like any other. Each value differs
// from its default, so a row that read the default fails.
func TestDoctorCmd_AuthSettingsGolden(t *testing.T) {
	want := []struct{ key, value, status string }{
		{"DUBPLATE_MAGIC_LINK_EXPIRY", "20m", `"20m" (source env-seed, next operation, sensitive)`},
		{"DUBPLATE_SESSION_LIFETIME", "1h", `"1h" (source env-seed, next operation, sensitive)`},
		{"DUBPLATE_SESSION_ABSOLUTE_CAP", "4000h", `"4000h" (source env-seed, immediately, sensitive)`},
		{"DUBPLATE_NONCE_RETENTION", "2h", `"2h" (source env-seed, next operation)`},
		{"DUBPLATE_LINK_REQUESTS_PER_ADDRESS_PER_HOUR", "7", `"7" (source env-seed, restart)`},
		{"DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "22", `"22" (source env-seed, restart)`},
		{"DUBPLATE_LINK_CONSUMPTION_ATTEMPTS_PER_CLIENT_PER_HOUR", "33", `"33" (source env-seed, restart)`},
		{"DUBPLATE_LINK_MAIL_CEILING_PER_HOUR", "66", `"66" (source env-seed, restart)`},
		{"DUBPLATE_AUTH_REQUEST_BODY_CAP", "8192", `"8192" (source env-seed, next operation)`},
		{"DUBPLATE_UNKNOWN_ADDRESS_ANSWER", "generic", `"generic" (source env-seed, next operation, sensitive)`},
	}
	environ := withMusicVolume(t, doctorEnviron(t))
	for _, w := range want {
		environ = append(environ, w.key+"="+w.value)
	}
	if _, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil {
		t.Fatalf("prepare: %v", err)
	}
	seedDoctorAdmin(t, environ)

	var stdout, stderr bytes.Buffer
	if err := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr).Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil; stdout %q", err, stdout.String())
	}
	for _, w := range want {
		status, ok := doctorKeyStatus(stdout.String(), w.key)
		if !ok {
			t.Errorf("stdout = %q, want exactly one row keyed %s", stdout.String(), w.key)
			continue
		}
		if status != w.status {
			t.Errorf("%s row = %q, want %q", w.key, status, w.status)
		}
	}
}

// TestDoctorCmd_AdminEmailRow is the DUBPLATE_ADMIN_EMAIL row's golden
// over every roster standing the store can report: ok when the address
// names a live admin (a mixed-case value included, which the wall
// lowercases), a warning that never fails doctor when it names a live
// non-admin, a removed member, or no member, and the plain "ok" where
// the roster cannot speak (an empty roster, which the members row
// already fails; an absent, unvouched, or older store). Fixture state:
// every roster-bearing store holds a live admin other than the named
// address except the empty one, so exit status 0 is reachable and a
// warning that failed doctor would show; the removed address belongs
// to a tombstoned admin, so a lookup that ignores tombstones reads it
// as an admin and reports ok.
func TestDoctorCmd_AdminEmailRow(t *testing.T) {
	const (
		removedOrNone = "warning: this address is not a live member; run dubplate-server member add <address> --admin, or set DUBPLATE_ADMIN_EMAIL to a live admin"
		nonAdmin      = "warning: this address is a live member without the admin flag; set DUBPLATE_ADMIN_EMAIL to a live admin"
	)
	roster := func(t *testing.T, stateDir string) {
		seedBackupStore(t, stateDir)
		seedMembers(t, stateDir,
			seedMember{address: "boss@example.test", displayName: "Boss", username: "boss", admin: true},
			seedMember{address: "plain@example.test", displayName: "Plain", username: "plain"},
			seedMember{address: "gone@example.test", displayName: "Gone", username: "gone", admin: true, tombstoned: true})
	}
	tests := []struct {
		name     string
		address  string
		setup    func(t *testing.T, stateDir string)
		want     string
		wantFail bool
	}{
		{name: "a live admin", address: "boss@example.test", setup: roster, want: "ok"},
		{name: "a mixed-case address naming a live admin", address: "Boss@Example.TEST", setup: roster, want: "ok"},
		{name: "a live non-admin", address: "plain@example.test", setup: roster, want: nonAdmin},
		{name: "a tombstoned admin", address: "gone@example.test", setup: roster, want: removedOrNone},
		{name: "an address no member holds", address: "nobody@example.test", setup: roster, want: removedOrNone},
		{
			name:     "an empty roster, which the members row fails",
			address:  "boss@example.test",
			setup:    func(t *testing.T, stateDir string) { seedBackupStore(t, stateDir) },
			want:     "ok",
			wantFail: true,
		},
		{name: "no store yet", address: "boss@example.test", setup: func(*testing.T, string) {}, want: "ok"},
		{
			name:     "a store the inspection cannot vouch for",
			address:  "boss@example.test",
			setup:    buildZeroByteStoreLeaf,
			want:     "ok",
			wantFail: true,
		},
		{
			name:    "a store older than the users table",
			address: "boss@example.test",
			setup:   func(t *testing.T, stateDir string) { olderStore(t, stateDir, 2) },
			want:    "ok",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := withEnv(doctorEnviron(t), "DUBPLATE_ADMIN_EMAIL="+tt.address)
			tt.setup(t, environValue(environ, "DUBPLATE_STATE_DIR"))

			rows, _, err := doctorRows(t, environ)
			if got := rows["DUBPLATE_ADMIN_EMAIL"]; got != tt.want {
				t.Errorf("DUBPLATE_ADMIN_EMAIL row = %q, want %q", got, tt.want)
			}
			if tt.wantFail != (err != nil) {
				t.Errorf("Execute() error = %v, want a failure = %v", err, tt.wantFail)
			}
		})
	}

	t.Run("the row keeps its place before the store row", func(t *testing.T) {
		environ := withEnv(doctorEnviron(t), "DUBPLATE_ADMIN_EMAIL=plain@example.test")
		roster(t, environValue(environ, "DUBPLATE_STATE_DIR"))
		var stdout, stderr bytes.Buffer
		cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("Execute() error = %v, want nil: a warning never fails doctor", err)
		}
		out := stdout.String()
		email, store := strings.Index(out, "DUBPLATE_ADMIN_EMAIL "), strings.Index(out, "\nstore ")
		if email < 0 || store < 0 || email > store {
			t.Errorf("DUBPLATE_ADMIN_EMAIL at %d, store at %d, want the wall row first", email, store)
		}
	})

	t.Run("an invalid address keeps the wall's refusal", func(t *testing.T) {
		environ := withEnv(doctorEnviron(t), "DUBPLATE_ADMIN_EMAIL=not-an-address")
		roster(t, environValue(environ, "DUBPLATE_STATE_DIR"))
		rows, _, err := doctorRows(t, environ)
		if got := rows["DUBPLATE_ADMIN_EMAIL"]; !strings.HasPrefix(got, "invalid") {
			t.Errorf("DUBPLATE_ADMIN_EMAIL row = %q, want the wall's invalid status", got)
		}
		if err == nil {
			t.Error("Execute() error = nil, want the wall's failure")
		}
	})
}

// TestDoctor_MusicVolumeRow pins the device row: on a fresh box whose
// staging and library root do not exist yet it names both and leaves
// doctor's exit alone; over sibling directories it reads ok, and a
// route collection's root reads "not created" until it exists, then
// reports its mount.
func TestDoctor_MusicVolumeRow(t *testing.T) {
	rows, _, err := doctorRows(t, doctorEnviron(t))
	if err != nil {
		t.Fatalf("doctor on a fresh box: %v", err)
	}
	want := "staging directory /srv/music/staging does not exist; library root /srv/music/library does not exist"
	if got := rows[musicVolumeDoctorRow]; got != want {
		t.Errorf("fresh box: row = %q, want %q", got, want)
	}

	environ := withMusicVolume(t, doctorEnviron(t))
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	seedBackupStore(t, stateDir)
	seedDoctorAdmin(t, environ)
	h, err := store.OpenWrite(t.Context(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	err = h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, owner, visibility, writer, intake_route)
			SELECT ?, 'Library', ?, id, 'group', 'server', 'default' FROM users LIMIT 1`,
			bytes.Repeat([]byte{0xab}, 16), []byte("abababababababababababababababab"))
		return err
	})
	if cerr := h.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		t.Fatalf("seed a route collection: %v", err)
	}
	root := filepath.Join(environValue(environ, "DUBPLATE_LIBRARY_ROOT"), "abababababababababababababababab")

	rows, _, err = doctorRows(t, environ)
	if err != nil {
		t.Fatalf("doctor over sibling directories: %v", err)
	}
	if want := "ok: staging and the library root share one mount; collection root " + root + " not created"; rows[musicVolumeDoctorRow] != want {
		t.Errorf("row = %q, want %q", rows[musicVolumeDoctorRow], want)
	}

	if err := os.Mkdir(root, 0o755); err != nil {
		t.Fatal(err)
	}
	rows, _, err = doctorRows(t, environ)
	if err != nil {
		t.Fatalf("doctor with the collection root created: %v", err)
	}
	if want := "ok: staging and the library root share one mount; collection root " + root + " on the library root's mount"; rows[musicVolumeDoctorRow] != want {
		t.Errorf("row = %q, want %q", rows[musicVolumeDoctorRow], want)
	}
}
