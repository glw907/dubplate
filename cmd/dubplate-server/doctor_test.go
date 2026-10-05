package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// storeRowLines is the store row's fixed contribution to doctor's
// line-count assertions: it prints exactly one line regardless of
// state, unlike reg's own declared families, so a count built from reg
// alone must add it explicitly.
const storeRowLines = 1

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

// probeRowLines is the four boundary rows' fixed contribution to
// doctor's line-count assertions (client ip, web, bind addresses,
// boundary probe), each printing exactly one line whatever its state.
// It is a literal, never taken from the rows' own list, so a row
// dropped from doctor changes the count a test expects.
const probeRowLines = 4

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

			reg := config.NewRegistry()
			registerShadowIntegration(reg)
			registerShadowThresholds(reg)

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
				want := "ok"
				switch {
				case row.Pending != "":
					want = row.Pending
				case tt.rowOverrides[row.Name] != "":
					want = tt.rowOverrides[row.Name]
				}
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
			// row: this reg declares no identity source, so no
			// roster row follows the thresholds.
			if li >= len(lines) {
				t.Fatalf("stdout ran out of lines before the store row")
			}
			if wantPrefix := fmt.Sprintf("%-32s", "store"); !strings.HasPrefix(lines[li], wantPrefix) {
				t.Errorf("store row %d = %q, want prefix %q", li, lines[li], wantPrefix)
			}
			li++

			// The backup rows follow the store row (Task 7): none of
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

			if li != len(lines) {
				t.Errorf("stdout has %d extra line(s) beyond owned+reserved+threshold+store+backup rows: %v", len(lines)-li, lines[li:])
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
	reg := config.NewRegistry()
	lines := doctorLines(t, reg)

	ownedCount := len(reg.OwnedWall())
	reservedCount := len(reg.ReservedClasses())
	integrationCount := len(reg.IntegrationClasses())
	thresholdCount := len(reg.Thresholds())
	sourceCount := len(reg.IdentitySources())
	// storeRowLines and backupDoctorRowsOff are the store row and
	// Task 7's backup rows: both print unconditionally, so neither is
	// one of reg's own declared families. doctorLines's own environ
	// (baseEnviron) never sets DUBPLATE_BACKUP_ENABLED.
	wantTotal := ownedCount + reservedCount + integrationCount + thresholdCount + sourceCount + storeRowLines + backupDoctorRowsOff + unidecodeRowLines + probeRowLines
	if len(lines) != wantTotal {
		t.Fatalf("doctor printed %d lines against a fresh registry, want %d (owned %d + reserved %d + integration %d + threshold %d + source %d + store %d + backup %d + unidecode %d + boundary %d)",
			len(lines), wantTotal, ownedCount, reservedCount, integrationCount, thresholdCount, sourceCount,
			storeRowLines, backupDoctorRowsOff, unidecodeRowLines, probeRowLines)
	}

	// Owned wall: one line per declared row.
	reg.AddWallRow(config.WallRow{Name: "DUBPLATE_TEST_SCRATCH_OWNED"})
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
		Flag:     "DUBPLATE_TEST_SCRATCH_FLAG",
		Required: []string{"DUBPLATE_TEST_SCRATCH_REQUIRED"},
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
	reg.AddThreshold(config.Threshold{Key: "DUBPLATE_TEST_SCRATCH_THRESH", Default: "1"})
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
		Required: []string{"DUBPLATE_TEST_SCRATCH_SOURCE_A", "DUBPLATE_TEST_SCRATCH_SOURCE_B"},
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
	wantTotal := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + backupDoctorRowsOn + unidecodeRowLines + probeRowLines + len(reg.Settings())
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
		reg := config.NewRegistry()
		reg.AddIdentitySource(config.IdentitySource{Name: "scratch-a", Selector: selector, Required: []string{requiredA}})
		reg.AddIdentitySource(config.IdentitySource{Name: "scratch-b", Selector: selector, Required: []string{requiredB}})
		reg.AddWallRow(config.WallRow{Name: selector})
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

// TestDoctorCmd_NoRunReportsAllGreenOnLoadError is the M4 gate test
// (D5's own load-bearing acceptance criterion): with the magic-link
// source selected and every other row satisfied, and only
// DUBPLATE_NONCE_RETENTION below the link expiry, doctor's exit status
// is 1 and at least one printed row names a failure. The property is
// asserted generally, not against one row's literal text, so it also
// covers whatever other row a future rung's own hook fails against.
func TestDoctorCmd_NoRunReportsAllGreenOnLoadError(t *testing.T) {
	environ := append(doctorEnviron(t),
		"DUBPLATE_MAGIC_LINK_EXPIRY=45m",
		"DUBPLATE_NONCE_RETENTION=30m",
	)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want a refusal: the retention is below the link expiry")
	}

	lines := strings.Split(strings.TrimRight(stdout.String(), "\n"), "\n")
	failing := 0
	for _, line := range lines {
		if strings.Contains(line, "invalid:") || strings.Contains(line, "missing") || strings.Contains(line, "unsatisfied") {
			failing++
		}
	}
	if failing == 0 {
		t.Errorf("stdout = %q, want at least one row naming a failure: doctor must never print an all-satisfied table while Load refuses", stdout.String())
	}

	// The source row itself must report the failure, not just some
	// threshold row (review M4): DUBPLATE_NONCE_RETENTION is a
	// threshold key, not one of magic-link's Required or Optional
	// names, so a doctor that only swept those two lists would print
	// this row "active" while the source's own Validate hook is the
	// refusal Load returned.
	want := fmt.Sprintf("%-32s %s", "magic-link", "unsatisfied")
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want a line with prefix %q", stdout.String(), want)
	}
}

// TestDoctorCmd_SessionLifetimeAboveCapReportsUnsatisfied is review
// finding 12's own repro (a 10-year session lifetime paired with a
// one-hour absolute cap): doctor's exit status is 1 and the magic-link
// row itself reports unsatisfied, rather than a green table that hides
// a configuration that mints a cookie its own server refuses within the
// hour.
func TestDoctorCmd_SessionLifetimeAboveCapReportsUnsatisfied(t *testing.T) {
	environ := append(doctorEnviron(t),
		"DUBPLATE_SESSION_LIFETIME=87600h",
		"DUBPLATE_SESSION_ABSOLUTE_CAP=1h",
	)
	reg := productionRegistry()

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute() error = nil, want a refusal: the lifetime is above the absolute cap")
	}

	want := fmt.Sprintf("%-32s %s", "magic-link", "unsatisfied")
	if !strings.Contains(stdout.String(), want) {
		t.Errorf("stdout = %q, want a line with prefix %q", stdout.String(), want)
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

	// storeRowLines and backupDoctorRowsOff are the store row and
	// Task 7's backup rows: doctorEnviron's roster names an admin, so
	// no roster row adds a second, and it never sets
	// DUBPLATE_BACKUP_ENABLED.
	wantTotal := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + backupDoctorRowsOff + unidecodeRowLines + probeRowLines + len(reg.Settings())
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
		{
			// DUBPLATE_MAGIC_LINK_EXPIRY is a registered threshold,
			// not one of magic-link's Required or Optional keys, so
			// its bound refusal reaches this row only through
			// validateMagicLinkSource's own [sourceKeyError], never
			// through the per-key Required/Optional sweep above.
			name:    "link expiry above its ceiling",
			extra:   []string{"DUBPLATE_MAGIC_LINK_EXPIRY=2h"},
			wantKey: "DUBPLATE_MAGIC_LINK_EXPIRY",
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

// TestDoctorCmd_RosterRow proves the R3 roster row: a valid roster
// naming an admin reports no extra row and no failure, while a missing
// roster, a roster carrying a duplicate streamer_username, a roster
// carrying two bad records, and a hand-edited roster that loads but
// names no admin, with members or without, and a roster a group may
// write, each render one roster row
// and take doctor's own exit status nonzero alongside it. A missing
// roster fails because R2 makes it a startup refusal. Named mutation 8 (drop the doctor roster row)
// makes every failing case report a clean table for a roster that
// refuses to load; the two-record case pins the one-row rule against
// an errors.Join newline leaking into the table.
func TestDoctorCmd_RosterRow(t *testing.T) {
	const fixtureAddressA = "roster-row-a@example.test"
	const fixtureAddressB = "roster-row-b@example.test"
	const fixtureAddressC = "roster-row-c@example.test"
	const fixturePrincipal = "roster-row-principal"

	tests := []struct {
		name          string
		rosterContent string      // empty means the roster file is never created
		rosterMode    os.FileMode // the file's mode once written; zero means 0600
		wantStatus    string      // the roster row's status prefix; empty means no roster row
	}{
		{
			name:          "a valid roster",
			rosterContent: fmt.Sprintf(`{"schema_version":1,"members":[{"address":%q,"display_name":"A","admin":true}]}`, fixtureAddressA),
			wantStatus:    "",
		},
		{
			name:          "a hand-edited roster with no members",
			rosterContent: `{"schema_version":1,"members":[]}`,
			wantStatus:    "no admin; run dubplate-server member add --admin",
		},
		{
			name:          "a hand-edited roster naming no admin",
			rosterContent: fmt.Sprintf(`{"schema_version":1,"members":[{"address":%q,"display_name":"A"}]}`, fixtureAddressA),
			wantStatus:    "no admin; run dubplate-server member add --admin",
		},
		{
			name:          "a roster not yet bootstrapped",
			rosterContent: "",
			wantStatus:    "missing",
		},
		{
			name: "a roster with a duplicate streamer_username",
			rosterContent: fmt.Sprintf(
				`{"schema_version":1,"members":[`+
					`{"address":%q,"display_name":"A","streamer_username":%q},`+
					`{"address":%q,"display_name":"B","streamer_username":%q}`+
					`]}`,
				fixtureAddressA, fixturePrincipal, fixtureAddressB, fixturePrincipal),
			wantStatus: "invalid: ",
		},
		{
			name:          "a group-writable roster",
			rosterContent: fmt.Sprintf(`{"schema_version":1,"members":[{"address":%q,"display_name":"A","admin":true}]}`, fixtureAddressA),
			rosterMode:    0o620,
			wantStatus:    "invalid: auth: roster file unreadable: ",
		},
		{
			name: "a roster with two bad records",
			rosterContent: fmt.Sprintf(
				`{"schema_version":1,"members":[`+
					`{"address":%q,"display_name":"A","streamer_username":%q},`+
					`{"address":%q,"display_name":"B","streamer_username":%q},`+
					`{"address":%q,"display_name":"C","streamer_username":%q}`+
					`]}`,
				fixtureAddressA, fixturePrincipal, fixtureAddressB, fixturePrincipal, fixtureAddressC, fixturePrincipal),
			wantStatus: "invalid: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rosterPath := filepath.Join(t.TempDir(), "roster.json")
			if tt.rosterContent != "" {
				if err := os.WriteFile(rosterPath, []byte(tt.rosterContent), 0o600); err != nil {
					t.Fatalf("write fixture roster: %v", err)
				}
				if tt.rosterMode != 0 {
					if err := os.Chmod(rosterPath, tt.rosterMode); err != nil {
						t.Fatalf("Chmod fixture roster: %v", err)
					}
				}
			}

			environ := append(dropKey(doctorEnviron(t), "DUBPLATE_ROSTER_FILE"), "DUBPLATE_ROSTER_FILE="+rosterPath)
			reg := productionRegistry()

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
			err := cmd.Execute()

			wantFailure := tt.wantStatus != ""
			if wantFailure && err == nil {
				t.Fatal("Execute() error = nil, want a refusal")
			}
			if !wantFailure && err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}

			out := stdout.String()
			lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
			var rosterRows []string
			for _, line := range lines {
				if strings.HasPrefix(line, "roster ") {
					rosterRows = append(rosterRows, line)
				}
			}
			if !wantFailure {
				if len(rosterRows) != 0 {
					t.Errorf("stdout has roster row(s) %q, want none", rosterRows)
				}
				return
			}

			if len(rosterRows) != 1 {
				t.Fatalf("stdout has %d roster row(s), want exactly 1; stdout = %q", len(rosterRows), out)
			}
			wantPrefix := fmt.Sprintf("%-32s %s", "roster", tt.wantStatus)
			if !strings.HasPrefix(rosterRows[0], wantPrefix) {
				t.Errorf("roster row = %q, want prefix %q", rosterRows[0], wantPrefix)
			}

			for _, value := range []string{fixtureAddressA, fixtureAddressB, fixtureAddressC, fixturePrincipal} {
				if strings.Contains(out, value) {
					t.Errorf("stdout = %q, want it to not contain fixture value %q", out, value)
				}
			}

			// +1 for the roster row (this failing case), +1 for the
			// store row, which always prints, plus Task 7's backup
			// rows (doctorEnviron never sets DUBPLATE_BACKUP_ENABLED)
			// and +1 for the unidecode row, doctorEnviron's installed
			// table.
			wantTotal := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) + len(reg.IdentitySources()) + 2 + backupDoctorRowsOff + unidecodeRowLines + probeRowLines + len(reg.Settings())
			if len(lines) != wantTotal {
				t.Errorf("stdout has %d line(s), want %d (the registry's own rows plus one roster row plus one store row); stdout = %q", len(lines), wantTotal, out)
			}
		})
	}
}

// doctorEnviron returns authEnviron with DUBPLATE_ROSTER_FILE naming a
// roster of one admin, written into t.TempDir(), and DUBPLATE_STATE_DIR
// pointed at a fresh t.TempDir() of its own: doctor's store row now
// inspects that directory, and authEnviron's own fixed path is not one
// a test process may create the store leaf's directories under. doctor
// fails a roster that names no admin, so a test expecting the roster
// to pass needs one that does rather than authEnviron's empty roster.
func doctorEnviron(t *testing.T) []string {
	t.Helper()
	rosterPath := filepath.Join(t.TempDir(), "roster.json")
	const admin = `{"schema_version":1,"members":[{"address":"doctor-admin@example.test","display_name":"Admin","admin":true}]}`
	if err := os.WriteFile(rosterPath, []byte(admin), 0o600); err != nil {
		t.Fatalf("write fixture roster: %v", err)
	}
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)
	environ := replaceKey(authEnviron(t), "DUBPLATE_STATE_DIR", stateDir)
	return append(dropKey(environ, "DUBPLATE_ROSTER_FILE"),
		"DUBPLATE_ROSTER_FILE="+rosterPath,
		"DUBPLATE_SERVE_PORT="+strconv.Itoa(reservedPort()))
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
	cmd := newDoctorCmd(environ, config.NewRegistry(), &stdout, &stderr)
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
	// literal now. config.NewRegistry() alone declares no identity
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

	// Resolution R15: reserved classes print in table order, not rung
	// order, and rung 13 (decision and review queues) precedes rung 12
	// (streamer) in that table. TestDoctorCmd itself cannot pin this: it
	// builds its expected lines by iterating config.ReservedClasses()
	// itself, so a registry that sorted both the production output and
	// the test's own expectations would still agree. This assertion pins
	// the order against the raw stdout instead.
	out := stdout.String()
	queuesIdx := strings.Index(out, "decision and review queues")
	streamerIdx := strings.Index(out, "streamer")
	if queuesIdx < 0 || streamerIdx < 0 {
		t.Fatalf("stdout missing a reserved class row: queues idx %d, streamer idx %d", queuesIdx, streamerIdx)
	}
	if queuesIdx >= streamerIdx {
		t.Errorf("decision and review queues printed at %d, streamer at %d; want queues before streamer (R15)", queuesIdx, streamerIdx)
	}
}

// TestDoctorCmd_ShadowRows proves the duration threshold renders as a
// duration rather than "invalid:", and the shadow integration row
// reflects its flag state.
func TestDoctorCmd_ShadowRows(t *testing.T) {
	reg := config.NewRegistry()
	registerShadowIntegration(reg)
	registerShadowThresholds(reg)
	stateDir := t.TempDir()
	installTestUnidecodeTable(t, stateDir)
	environ := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", stateDir)

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	wantFloor := fmt.Sprintf("%-32s %s\n", "DUBPLATE_SHADOW_FLOOR", "720h0m0s")
	if !strings.Contains(stdout.String(), wantFloor) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantFloor)
	}

	wantBudget := fmt.Sprintf("%-32s %s\n", "DUBPLATE_SHADOW_BUDGET", "60")
	if !strings.Contains(stdout.String(), wantBudget) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantBudget)
	}

	wantDisabled := fmt.Sprintf("%-32s %s\n", "shadow", "disabled")
	if !strings.Contains(stdout.String(), wantDisabled) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantDisabled)
	}

	enabledEnviron := append(environ,
		"DUBPLATE_SHADOW_ENABLED=true",
		"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=/var/lib/incumbent",
		"DUBPLATE_SHADOW_REVIEW_DIR=/srv/review",
	)
	stdout.Reset()
	cmd = newDoctorCmd(enabledEnviron, reg, &stdout, &stderr)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	wantEnabled := fmt.Sprintf("%-32s %s\n", "shadow", "enabled")
	if !strings.Contains(stdout.String(), wantEnabled) {
		t.Errorf("stdout = %q, want it to contain %q", stdout.String(), wantEnabled)
	}
}

// TestDoctorCmd_FloatThreshold proves a registered KindFloat threshold
// prints its effective value in the thresholds block, in the same
// pinned "%-32s %s" form every other row uses. wantLine is a literal
// ("0.04"), not derived by calling config.ParseFloat itself.
func TestDoctorCmd_FloatThreshold(t *testing.T) {
	reg := config.NewRegistry()
	reg.AddThreshold(config.Threshold{
		Key:     "DUBPLATE_TEST_STRONG_REC_THRESH",
		Kind:    config.KindFloat,
		Default: "0.04",
		Doc:     "a float between 0 and 1",
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

// TestDoctorCmd_StoreBusyTimeoutDefault proves the threshold row shows
// 15s by default and the set value when set, in the same "%-32s %s"
// form every duration threshold uses.
func TestDoctorCmd_StoreBusyTimeoutDefault(t *testing.T) {
	reg := config.NewRegistry()
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
	reg := config.NewRegistry()
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
	reg := config.NewRegistry()
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
