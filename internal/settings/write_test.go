package settings

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// More test keys, beside helpers_test.go's.
const (
	keySensitive   = "DUBPLATE_TEST_SENSITIVE"
	keyRestart     = "DUBPLATE_TEST_RESTART"
	keyWall        = "DUBPLATE_TEST_WALL"
	keyThreshold   = "DUBPLATE_TEST_THRESHOLD"
	keyIntegration = "DUBPLATE_TEST_INTEGRATION"
	keyIntegExtra  = "DUBPLATE_TEST_INTEGRATION_EXTRA"
)

// testAuthor is the actor every write in these tests records.
var testAuthor = Author{Actor: "member:00112233445566778899aabbccddeeff", Channel: "cookie"}

// errGuardRefused is what a refusing test guard returns.
var errGuardRefused = errors.New("actor is not an admin")

// writeRegistry is newRegistry plus a sensitive immediate setting, a
// restart-applied setting, and one key of every family a write must
// refuse: a wall row, a threshold, and an integration class's flag and
// required key.
func writeRegistry() *config.Registry {
	reg := newRegistry()
	reg.AddSetting(config.Setting{
		Key: keySensitive, Kind: config.KindInt, Default: "5", Doc: "an integer",
		Apply: config.ApplyImmediately, Floor: "1", Ceiling: "10",
		Description: testDescription, Mark: config.Sensitive,
	})
	reg.AddSetting(config.Setting{
		Key: keyRestart, Kind: config.KindInt, Default: "5", Doc: "an integer",
		Apply: config.ApplyRestart, Floor: "1", Ceiling: "10",
		Description: testDescription, Mark: config.NotSensitive,
	})
	reg.AddWallRow(config.WallRow{Name: keyWall, Description: testDescription, Mark: config.Sensitive})
	reg.AddThreshold(config.Threshold{
		Key: keyThreshold, Kind: config.KindInt, Default: "3", Doc: "an integer",
		Description: testDescription, Mark: config.NotSensitive,
	})
	reg.AddIntegrationClass(config.IntegrationClass{
		Name: "test",
		Flag: config.KeyDecl{Name: keyIntegration, Description: testDescription, Mark: config.NotSensitive},
		Required: []config.KeyDecl{
			{Name: keyIntegExtra, Description: testDescription, Mark: config.Sensitive},
		},
	})
	return reg
}

// storeState is everything a write can change.
type storeState struct {
	rows  []storedRow
	audit []auditRow
	gen   int64
}

func snapshot(t *testing.T, h *store.Handle) storeState {
	t.Helper()
	return storeState{readSettings(t, h), readAudit(t, h), readGeneration(t, h)}
}

// wantUnchanged fails t unless h's store holds exactly before.
func wantUnchanged(t *testing.T, h *store.Handle, before storeState) {
	t.Helper()
	after := snapshot(t, h)
	if !slices.Equal(after.rows, before.rows) || !slices.Equal(after.audit, before.audit) || after.gen != before.gen {
		t.Errorf("store changed:\n rows  %v -> %v\n audit %v -> %v\n gen   %d -> %d",
			before.rows, after.rows, before.audit, after.audit, before.gen, after.gen)
	}
}

func adminRow(key, value, seed string) fixtureRow { return fixtureRow{key, value, "admin", seed} }

func seededRow(key, value string) fixtureRow { return fixtureRow{key, value, "env-seed", value} }

// setAudit and resetAudit are the audit rows a committed write leaves.
func setAudit(key, oldVal, newVal string) auditRow {
	return auditRow{at: testNow.UnixMicro(), kind: "set", actor: testAuthor.Actor, channel: "cookie", key: key, oldVal: oldVal, newVal: newVal}
}

func resetAudit(key, oldVal string) auditRow {
	return auditRow{at: testNow.UnixMicro(), kind: "reset", actor: testAuthor.Actor, channel: "cookie", key: key, oldVal: oldVal, newVal: "<NULL>"}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		fixture []fixtureRow
		environ []string
		key     string
		value   string
		// wantErr is nil for a write that answers, else the sentinel
		// the refusal wraps; the store must then be unchanged.
		wantErr   error
		wantRes   Result
		wantRows  []storedRow
		wantAudit []auditRow
		wantBump  bool
	}{
		{
			name: "no row: admin row, empty seed, audit, bump",
			key:  keyKeep, value: "30",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "30", Previous: "14", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "30", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keyKeep, "<NULL>", "30")},
			wantBump:  true,
		},
		{
			name:    "over an env-seed row: seed_value records the seed replaced",
			fixture: []fixtureRow{seededRow(keyKeep, "20")},
			key:     keyKeep, value: "30",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "30", Previous: "20", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "30", "admin", "20"}},
			wantAudit: []auditRow{setAudit(keyKeep, "20", "30")},
			wantBump:  true,
		},
		{
			name:    "over an admin row: seed_value keeps the earlier seed",
			fixture: []fixtureRow{adminRow(keyKeep, "25", "20")},
			key:     keyKeep, value: "30",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "30", Previous: "25", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "30", "admin", "20"}},
			wantAudit: []auditRow{setAudit(keyKeep, "25", "30")},
			wantBump:  true,
		},
		{
			name:    "over an admin row with no seed: seed_value stays empty",
			fixture: []fixtureRow{adminRow(keyKeep, "25", "")},
			key:     keyKeep, value: "30",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "30", Previous: "25", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "30", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keyKeep, "25", "30")},
			wantBump:  true,
		},
		{
			name: "a sensitive immediate key names both",
			key:  keySensitive, value: "7",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "7", Previous: "5", Apply: config.ApplyImmediately, Sensitive: true},
			wantRows:  []storedRow{{keySensitive, "7", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keySensitive, "<NULL>", "7")},
			wantBump:  true,
		},
		{
			name: "a restart key writes and names the restart",
			key:  keyRestart, value: "7",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "7", Previous: "5", Apply: config.ApplyRestart},
			wantRows:  []storedRow{{keyRestart, "7", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keyRestart, "<NULL>", "7")},
			wantBump:  true,
		},
		{
			name:    "a store already refusing one key takes a valid write to another",
			fixture: []fixtureRow{adminRow(keyKeep, "1", "")},
			key:     keyRetention, value: "48h",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "48h", Previous: "2160h", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "1", "admin", "<NULL>"}, {keyRetention, "48h", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keyRetention, "<NULL>", "48h")},
			wantBump:  true,
		},
		{
			name:    "a valid write to the refusing key clears it",
			fixture: []fixtureRow{adminRow(keyKeep, "1", "")},
			key:     keyKeep, value: "5",
			wantRes:   Result{Written: true, Source: SourceAdmin, Value: "5", Previous: "1", Apply: config.ApplyNextOperation},
			wantRows:  []storedRow{{keyKeep, "5", "admin", "<NULL>"}},
			wantAudit: []auditRow{setAudit(keyKeep, "1", "5")},
			wantBump:  true,
		},
		{
			name:    "an out-of-bounds write to the key already refusing is refused",
			fixture: []fixtureRow{adminRow(keyKeep, "1", "")},
			key:     keyKeep, value: "2",
			wantErr: ErrInvalidValue,
		},
		{name: "unparseable", key: keyKeep, value: "many", wantErr: ErrInvalidValue},
		{name: "below the floor", key: keyKeep, value: "2", wantErr: ErrInvalidValue},
		{name: "above the ceiling", key: keyKeep, value: "1001", wantErr: ErrInvalidValue},
		{name: "an unregistered key", key: "DUBPLATE_TEST_NOPE", value: "1", wantErr: ErrNotSetting},
		{name: "a wall key", key: keyWall, value: "1", wantErr: ErrNotSetting},
		{name: "a threshold", key: keyThreshold, value: "1", wantErr: ErrNotSetting},
		{name: "an integration flag", key: keyIntegration, value: "1", wantErr: ErrNotSetting},
		{name: "an integration required key", key: keyIntegExtra, value: "1", wantErr: ErrNotSetting},
		{
			name:    "the environment sets another value",
			fixture: []fixtureRow{seededRow(keyKeep, "20")},
			environ: []string{keyKeep + "=20"},
			key:     keyKeep, value: "30",
			wantErr: ErrEnvironmentSets,
		},
		{
			name:    "equal to the environment's value: writes nothing, names the environment",
			fixture: []fixtureRow{seededRow(keyKeep, "20")},
			environ: []string{keyKeep + "=20"},
			key:     keyKeep, value: "20",
			wantRes:  Result{Source: SourceEnvironment, Value: "20", Apply: config.ApplyNextOperation},
			wantRows: []storedRow{{keyKeep, "20", "env-seed", "20"}},
		},
		{
			name:    "equal to the environment's value in another spelling",
			fixture: []fixtureRow{seededRow(keyRetention, "720h")},
			environ: []string{keyRetention + "=720h"},
			key:     keyRetention, value: "43200m",
			wantRes:  Result{Source: SourceEnvironment, Value: "720h", Apply: config.ApplyNextOperation},
			wantRows: []storedRow{{keyRetention, "720h", "env-seed", "720h"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.fixture...)
			before := snapshot(t, h)

			got, err := Set(ctxFor(t), h, writeRegistry(), tt.environ, testAuthor, nil, tt.key, tt.value, testNow)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Set err = %v, want %v", err, tt.wantErr)
				}
				wantUnchanged(t, h, before)
				return
			}
			if err != nil {
				t.Fatalf("Set: %v", err)
			}
			if got != tt.wantRes {
				t.Errorf("Result = %+v, want %+v", got, tt.wantRes)
			}
			after := snapshot(t, h)
			if !slices.Equal(after.rows, tt.wantRows) {
				t.Errorf("settings = %v, want %v", after.rows, tt.wantRows)
			}
			if !slices.Equal(after.audit, tt.wantAudit) {
				t.Errorf("audit = %v, want %v", after.audit, tt.wantAudit)
			}
			wantGen := int64(fixtureGeneration)
			if tt.wantBump {
				wantGen++
			}
			if after.gen != wantGen {
				t.Errorf("generation = %d, want %d", after.gen, wantGen)
			}
		})
	}
}

func TestReset(t *testing.T) {
	tests := []struct {
		name      string
		fixture   []fixtureRow
		environ   []string
		key       string
		wantErr   error
		wantRes   Result
		wantRows  []storedRow
		wantAudit []auditRow
		wantBump  bool
	}{
		{
			name:      "an admin row: deleted, audited with an empty new value, bumped",
			fixture:   []fixtureRow{adminRow(keyKeep, "30", "20")},
			key:       keyKeep,
			wantRes:   Result{Written: true, Source: SourceDefault, Value: "14", Previous: "30", Apply: config.ApplyNextOperation},
			wantAudit: []auditRow{resetAudit(keyKeep, "30")},
			wantBump:  true,
		},
		{
			name:      "a sensitive admin row names the mark",
			fixture:   []fixtureRow{adminRow(keySensitive, "7", "")},
			key:       keySensitive,
			wantRes:   Result{Written: true, Source: SourceDefault, Value: "5", Previous: "7", Apply: config.ApplyImmediately, Sensitive: true},
			wantAudit: []auditRow{resetAudit(keySensitive, "7")},
			wantBump:  true,
		},
		{
			name:      "an unparseable admin row resets",
			fixture:   []fixtureRow{adminRow(keyKeep, "many", "")},
			key:       keyKeep,
			wantRes:   Result{Written: true, Source: SourceDefault, Value: "14", Previous: "many", Apply: config.ApplyNextOperation},
			wantAudit: []auditRow{resetAudit(keyKeep, "many")},
			wantBump:  true,
		},
		{
			name:    "no row writes nothing and names the default",
			key:     keyKeep,
			wantRes: Result{Source: SourceDefault, Value: "14", Apply: config.ApplyNextOperation},
		},
		{
			name:     "an env-seed row writes nothing and names the seed",
			fixture:  []fixtureRow{seededRow(keyKeep, "20")},
			environ:  []string{keyKeep + "=20"},
			key:      keyKeep,
			wantRes:  Result{Source: SourceEnvSeed, Value: "20", Apply: config.ApplyNextOperation},
			wantRows: []storedRow{{keyKeep, "20", "env-seed", "20"}},
		},
		{
			name:    "an admin row for a key the environment sets is refused",
			fixture: []fixtureRow{adminRow(keyKeep, "20", "20")},
			environ: []string{keyKeep + "=20"},
			key:     keyKeep,
			wantErr: ErrEnvironmentSets,
		},
		{name: "an unregistered key", key: "DUBPLATE_TEST_NOPE", wantErr: ErrNotSetting},
		{name: "a wall key", key: keyWall, wantErr: ErrNotSetting},
		{name: "a threshold", key: keyThreshold, wantErr: ErrNotSetting},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, tt.fixture...)
			before := snapshot(t, h)

			got, err := Reset(ctxFor(t), h, writeRegistry(), tt.environ, testAuthor, nil, tt.key, testNow)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("Reset err = %v, want %v", err, tt.wantErr)
				}
				wantUnchanged(t, h, before)
				return
			}
			if err != nil {
				t.Fatalf("Reset: %v", err)
			}
			if got != tt.wantRes {
				t.Errorf("Result = %+v, want %+v", got, tt.wantRes)
			}
			after := snapshot(t, h)
			if !slices.Equal(after.rows, tt.wantRows) {
				t.Errorf("settings = %v, want %v", after.rows, tt.wantRows)
			}
			if !slices.Equal(after.audit, tt.wantAudit) {
				t.Errorf("audit = %v, want %v", after.audit, tt.wantAudit)
			}
			wantGen := int64(fixtureGeneration)
			if tt.wantBump {
				wantGen++
			}
			if after.gen != wantGen {
				t.Errorf("generation = %d, want %d", after.gen, wantGen)
			}
		})
	}
}

// TestReset_EnvironmentSetsByChannel: an admin row for a key the
// environment sets is refused to a cookie reset and deleted by a
// command-line one.
func TestReset_EnvironmentSetsByChannel(t *testing.T) {
	environ := []string{keyKeep + "=20"}
	t.Run("the cookie channel is refused", func(t *testing.T) {
		h := newStore(t)
		writeFixture(t, h, adminRow(keyKeep, "20", "20"))
		before := snapshot(t, h)

		_, err := Reset(ctxFor(t), h, writeRegistry(), environ, testAuthor, nil, keyKeep, testNow)

		if !errors.Is(err, ErrEnvironmentSets) {
			t.Fatalf("Reset err = %v, want ErrEnvironmentSets", err)
		}
		wantUnchanged(t, h, before)
	})
	t.Run("the command line deletes the row", func(t *testing.T) {
		h := newStore(t)
		writeFixture(t, h, adminRow(keyKeep, "20", "20"))
		operator := Author{Actor: "operator:glw907", Channel: "command-line"}

		got, err := Reset(ctxFor(t), h, writeRegistry(), environ, operator, nil, keyKeep, testNow)

		if err != nil {
			t.Fatalf("Reset: %v", err)
		}
		if !got.Written || got.Previous != "20" {
			t.Errorf("Result = %+v, want a write that replaced 20", got)
		}
		if after := snapshot(t, h); len(after.rows) != 0 {
			t.Errorf("settings = %v, want the row deleted", after.rows)
		}
	})
}

// TestSet_CorrelatedRule: a write the rule refuses changes nothing,
// and a store already breaking the rule takes a write that adds no
// refusal, the per-key comparison covering both keys the rule names.
func TestSet_CorrelatedRule(t *testing.T) {
	t.Run("a write that breaks the rule is refused", func(t *testing.T) {
		h := newStore(t)
		writeFixture(t, h, adminRow(keyExpiry, "10m", ""), adminRow(keyNonce, "48h", ""))
		before := snapshot(t, h)

		_, err := Set(ctxFor(t), h, correlatedRegistry(), nil, testAuthor, nil, keyExpiry, "72h", testNow)

		if !errors.Is(err, ErrInvalidValue) {
			t.Fatalf("Set err = %v, want ErrInvalidValue", err)
		}
		wantErrText(t, err, "nonce retention covers link expiry")
		wantUnchanged(t, h, before)
	})

	t.Run("a store already breaking the rule takes a write that adds no refusal", func(t *testing.T) {
		h := newStore(t)
		writeFixture(t, h, adminRow(keyExpiry, "100h", ""), adminRow(keyNonce, "48h", ""))

		if _, err := Set(ctxFor(t), h, correlatedRegistry(), nil, testAuthor, nil, keyNonce, "60h", testNow); err != nil {
			t.Fatalf("Set while the rule already refuses: %v", err)
		}
		got := readSettings(t, h)
		want := []storedRow{{keyExpiry, "100h", "admin", "<NULL>"}, {keyNonce, "60h", "admin", "<NULL>"}}
		if !slices.Equal(got, want) {
			t.Errorf("settings = %v, want %v", got, want)
		}
	})
}

// TestSet_JointBreakCommitsOne: two writes, each valid alone, that
// jointly break a correlated rule are queued behind a second
// connection's open write transaction. Whichever commits first, the
// other reads the rows it left and is refused.
func TestSet_JointBreakCommitsOne(t *testing.T) {
	dir := newStateDir(t)
	h, other := openWrite(t, dir), openWrite(t, dir)
	reg := correlatedRegistry()
	// Expiry 10m, retention 48h. Expiry to 40h is valid beside 48h and
	// retention to 20h is valid beside 10m; together 20h < 40h.
	writeFixture(t, h, adminRow(keyExpiry, "10m", ""), adminRow(keyNonce, "48h", ""))

	held, release, holdDone := make(chan struct{}), make(chan struct{}), make(chan error, 1)
	holdCtx := ctxFor(t)
	go func() {
		holdDone <- other.Update(holdCtx, func(context.Context, *store.Tx) error {
			close(held)
			<-release
			return nil
		})
	}()
	<-held

	writes := []struct{ key, value string }{{keyExpiry, "40h"}, {keyNonce, "20h"}}
	errs := make([]error, len(writes))
	var wg sync.WaitGroup
	for i, w := range writes {
		ctx := ctxFor(t)
		wg.Go(func() {
			_, errs[i] = Set(ctx, h, reg, nil, testAuthor, nil, w.key, w.value, testNow)
		})
	}
	// Let both writes reach the write lock, and any pre-read a faulty
	// write makes happen, before the first commit.
	time.Sleep(300 * time.Millisecond)
	close(release)
	wg.Wait()
	if err := <-holdDone; err != nil {
		t.Fatalf("holding transaction: %v", err)
	}

	committed := 0
	for i, err := range errs {
		switch {
		case err == nil:
			committed++
		case !errors.Is(err, ErrInvalidValue):
			t.Errorf("write %v: %v, want nil or ErrInvalidValue", writes[i], err)
		}
	}
	if committed != 1 {
		t.Fatalf("%d writes committed, want exactly 1 (errors %v)", committed, errs)
	}
	ev, err := Evaluate(ctxFor(t), h, reg, nil)
	if err != nil || ev.Err != nil {
		t.Errorf("store after both writes: read %v, evaluation %v, want a store that starts", err, ev.Err)
	}
	if got := len(readAudit(t, h)); got != 1 {
		t.Errorf("%d audit rows, want 1", got)
	}
	if got := readGeneration(t, h); got != fixtureGeneration+1 {
		t.Errorf("generation = %d, want %d", got, fixtureGeneration+1)
	}
}

// TestReset_ClearsRefusingRowsOneAtATime: a store holding two
// out-of-bounds admin rows takes a reset of each in turn, and then
// evaluates clean.
func TestReset_ClearsRefusingRowsOneAtATime(t *testing.T) {
	h := newStore(t)
	reg := newRegistry()
	writeFixture(t, h, adminRow(keyKeep, "1", ""), adminRow(keyRetention, "1h", ""))

	for _, key := range []string{keyKeep, keyRetention} {
		if _, err := Reset(ctxFor(t), h, reg, nil, testAuthor, nil, key, testNow); err != nil {
			t.Fatalf("Reset %s: %v", key, err)
		}
	}
	if n := countAdmin(t, h); n != 0 {
		t.Errorf("%d admin rows remain, want 0", n)
	}
	ev, err := Evaluate(ctxFor(t), h, reg, nil)
	if err != nil || ev.Err != nil {
		t.Errorf("after both resets: read %v, evaluation %v, want clean", err, ev.Err)
	}
}

// TestReset_RuleRefusal: a reset whose post-reset rows break a
// correlated rule changes nothing.
func TestReset_RuleRefusal(t *testing.T) {
	h := newStore(t)
	// Expiry 10m resets to the 15m default, above the 12m retention.
	writeFixture(t, h, adminRow(keyExpiry, "10m", ""), adminRow(keyNonce, "12m", ""))
	before := snapshot(t, h)

	_, err := Reset(ctxFor(t), h, correlatedRegistry(), nil, testAuthor, nil, keyExpiry, testNow)

	if !errors.Is(err, ErrInvalidValue) {
		t.Fatalf("Reset err = %v, want ErrInvalidValue", err)
	}
	wantErrText(t, err, "nonce retention covers link expiry")
	wantUnchanged(t, h, before)
}

// TestWrites_AccessorSeesThem: an Accessor over a second handle
// reads a key, a reset commits, and the accessor then reads the
// default; a set then moves it again. Each needs the generation bump.
func TestWrites_AccessorSeesThem(t *testing.T) {
	dir := newStateDir(t)
	h, reader := openWrite(t, dir), openWrite(t, dir)
	reg := newRegistry()
	writeFixture(t, h, adminRow(keyKeep, "30", "20"))
	acc := NewAccessor(reader, reg)
	keep := func() int {
		t.Helper()
		v, err := acc.Int(ctxFor(t), keepSetting("14"))
		if err != nil {
			t.Fatalf("accessor: %v", err)
		}
		return v
	}
	if got := keep(); got != 30 {
		t.Fatalf("accessor before reset = %d, want 30", got)
	}

	if _, err := Reset(ctxFor(t), h, reg, nil, testAuthor, nil, keyKeep, testNow); err != nil {
		t.Fatalf("Reset: %v", err)
	}
	if got := keep(); got != 14 {
		t.Errorf("accessor after reset = %d, want the default 14", got)
	}

	if _, err := Set(ctxFor(t), h, reg, nil, testAuthor, nil, keyKeep, "40", testNow); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got := keep(); got != 40 {
		t.Errorf("accessor after set = %d, want 40", got)
	}
}

// TestWrites_AuditRidesTheTransaction: a store whose settings_audit
// insert a trigger aborts takes neither the set nor the reset.
func TestWrites_AuditRidesTheTransaction(t *testing.T) {
	h := newStore(t)
	writeFixture(t, h, adminRow(keyRetention, "48h", ""))
	err := h.Update(ctxFor(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, `CREATE TRIGGER refuse_audit BEFORE INSERT ON settings_audit
			BEGIN SELECT RAISE(ABORT, 'audit refused'); END`)
		return err
	})
	if err != nil {
		t.Fatalf("create trigger: %v", err)
	}
	before := snapshot(t, h)
	reg := newRegistry()

	if _, err := Set(ctxFor(t), h, reg, nil, testAuthor, nil, keyKeep, "30", testNow); err == nil {
		t.Error("Set committed past a refused audit insert")
	}
	wantUnchanged(t, h, before)

	if _, err := Reset(ctxFor(t), h, reg, nil, testAuthor, nil, keyRetention, testNow); err == nil {
		t.Error("Reset committed past a refused audit insert")
	}
	wantUnchanged(t, h, before)
}

// TestWrites_Guard: the guard runs inside the write transaction on
// every path, and its refusal changes nothing, the no-op answers
// included.
func TestWrites_Guard(t *testing.T) {
	refuse := func(context.Context, *store.Tx) error { return errGuardRefused }
	tests := []struct {
		name  string
		write func(h *store.Handle, guard Guard) error
	}{
		{"set", func(h *store.Handle, g Guard) error {
			_, err := Set(ctxFor(t), h, newRegistry(), nil, testAuthor, g, keyKeep, "30", testNow)
			return err
		}},
		{"set equal to the environment", func(h *store.Handle, g Guard) error {
			_, err := Set(ctxFor(t), h, newRegistry(), []string{keyKeep + "=30"}, testAuthor, g, keyKeep, "30", testNow)
			return err
		}},
		{"reset", func(h *store.Handle, g Guard) error {
			_, err := Reset(ctxFor(t), h, newRegistry(), nil, testAuthor, g, keyRetention, testNow)
			return err
		}},
		{"reset with no row", func(h *store.Handle, g Guard) error {
			_, err := Reset(ctxFor(t), h, newRegistry(), nil, testAuthor, g, keyKeep, testNow)
			return err
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := newStore(t)
			writeFixture(t, h, adminRow(keyRetention, "48h", ""))
			before := snapshot(t, h)

			if err := tt.write(h, refuse); !errors.Is(err, errGuardRefused) {
				t.Errorf("write err = %v, want the guard's refusal", err)
			}
			wantUnchanged(t, h, before)
		})
	}

	t.Run("a passing guard reads the transaction's rows", func(t *testing.T) {
		h := newStore(t)
		writeFixture(t, h, adminRow(keyRetention, "48h", ""))
		var calls, rows int
		guard := func(ctx context.Context, tx *store.Tx) error {
			calls++
			return tx.QueryRow(ctx, `SELECT count(*) FROM settings`).Scan(&rows)
		}

		if _, err := Set(ctxFor(t), h, newRegistry(), nil, testAuthor, guard, keyKeep, "30", testNow); err != nil {
			t.Fatalf("Set: %v", err)
		}
		if calls != 1 || rows != 1 {
			t.Errorf("guard called %d times seeing %d rows, want once seeing the fixture's 1", calls, rows)
		}
	})
}
