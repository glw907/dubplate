package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// The two values a settings row's source column holds. A key whose
// value is the registry default has no row, so no source names it.
const (
	sourceEnvSeed = "env-seed"
	sourceAdmin   = "admin"
)

// selectRows reads every settings row, in key order so an evaluation's
// orphan list is stable.
const selectRows = `SELECT key, value, source, seed_value, actor, updated_at
	FROM settings ORDER BY key`

// A Row is one settings row as the store holds it.
type Row struct {
	Key   string
	Value string
	// Source is "env-seed" for a value the environment seeded and
	// "admin" for one an administrator set.
	Source string
	// SeedValue is the environment seed the row's value came from or,
	// on an admin row, replaced. It is empty when the row records
	// none; an environment seed is never empty, since an empty
	// environment value counts as unset.
	SeedValue string
	Actor     string
	UpdatedAt time.Time
}

// A Change is what the seed rule does to one setting's row.
type Change int

const (
	// Unchanged leaves the row, or its absence, as it is.
	Unchanged Change = iota
	// Seed writes a row for a key the store held none for.
	Seed
	// Reseed replaces an environment-seeded row's value.
	Reseed
	// Remove deletes an environment-seeded row, returning the key to
	// the registry default.
	Remove
)

// String returns c's name as a settings audit row's kind records it.
func (c Change) String() string {
	switch c {
	case Unchanged:
		return "unchanged"
	case Seed:
		return "seed"
	case Reseed:
		return "reseed"
	case Remove:
		return "remove"
	default:
		return fmt.Sprintf("Change(%d)", int(c))
	}
}

// An Outcome is the seed rule's verdict on one registered setting.
type Outcome struct {
	Key string
	// Stored is the key's row before the seed, or nil when the store
	// holds none.
	Stored *Row
	// Env is the environment's value for Key, as set, and EnvSet
	// reports whether it is set at all. A later entry wins over an
	// earlier one, and an empty value is unset.
	Env    string
	EnvSet bool
	// Value is the value in force once the seed applies, in string
	// form, or empty when Err is set.
	Value  string
	Change Change
	// Err is why the seed rule refuses Key, or nil when it does not.
	Err error
	// Report states the outcome in a sentence, a sensitive value
	// withheld.
	Report string

	setting config.Setting
}

// An Evaluation is the seed rule's verdict on a whole store.
type Evaluation struct {
	// Outcomes holds one outcome per registered setting, in
	// declaration order.
	Outcomes []Outcome
	// Orphans holds the rows no registered setting claims. The seed
	// rule never writes them.
	Orphans []Row
	// Err joins every refusal, each key's and each correlated rule's,
	// or is nil when the seed applies as a whole.
	Err error
}

// Evaluate runs the seed rule for reg's settings against the rows h
// holds and the environment in environ, as [os.Environ] lists it. It
// reads through h's read pool and writes nothing. The returned error
// reports a failed read; a refusal is the Evaluation's Err.
func Evaluate(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string) (Evaluation, error) {
	rows, err := readRows(ctx, h)
	if err != nil {
		return Evaluation{}, err
	}
	return evaluate(reg, environ, rows), nil
}

// EvaluateFacts runs the seed rule for reg's settings against facts, a
// [store.Facts] value [store.Inspect] already read, and the
// environment in environ. It opens no connection of its own: its
// caller has already inspected the store. When facts marks the
// settings table absent, a store below the version that adds it can
// hold no row for any key, so every outcome resolves to its default
// and reports that the store holds no settings yet, never comparing
// against the environment. check and doctor call it; backup calls
// [Evaluate] from its own write-mode handle instead (R-6c9).
func EvaluateFacts(reg *config.Registry, environ []string, facts store.Facts) Evaluation {
	if facts.SettingsAbsent {
		return evaluateTableAbsent(reg)
	}
	rows := make([]Row, len(facts.Settings))
	for i, r := range facts.Settings {
		rows[i] = Row{Key: r.Key, Value: r.Value, Source: r.Source, SeedValue: r.SeedValue, Actor: r.Actor, UpdatedAt: r.UpdatedAt}
	}
	return evaluate(reg, environ, rows)
}

// evaluateTableAbsent returns one outcome per registered setting
// resolving to its default, for a store whose version predates the
// settings table. It never reads environ: a store with pending
// migrations reads no setting from the environment (R-6c9's pre-seed
// read), so the report only ever says the store holds no settings yet.
func evaluateTableAbsent(reg *config.Registry) Evaluation {
	settings := reg.Settings()
	ev := Evaluation{Outcomes: make([]Outcome, len(settings))}
	for i, s := range settings {
		ev.Outcomes[i] = Outcome{
			Key:     s.Key,
			Value:   s.Default,
			Report:  "the store holds no settings yet",
			setting: s,
		}
	}
	return ev
}

// readRows reads every settings row through h's read pool.
func readRows(ctx context.Context, h *store.Handle) ([]Row, error) {
	rs, err := h.Query(ctx, selectRows)
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	defer func() { _ = rs.Close() }()
	rows, err := scanRows(rs)
	if err == nil {
		err = rs.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("read settings: %w", err)
	}
	return rows, nil
}

// rowScanner is the iteration both a read-pool query's [sql.Rows] and
// a write transaction's [store.Rows] offer.
type rowScanner interface {
	Next() bool
	Scan(dest ...any) error
}

// scanRows scans every settings row rs yields. The caller checks the
// iteration's own error.
func scanRows(rs rowScanner) ([]Row, error) {
	var rows []Row
	for rs.Next() {
		var (
			r    Row
			seed sql.NullString
			at   int64
		)
		if err := rs.Scan(&r.Key, &r.Value, &r.Source, &seed, &r.Actor, &at); err != nil {
			return nil, err
		}
		r.SeedValue = seed.String
		r.UpdatedAt = time.UnixMicro(at).UTC()
		rows = append(rows, r)
	}
	return rows, nil
}

// evaluate is the seed rule: each setting's outcome from its stored
// row and its environment value, then every correlated rule over the
// values in force after the whole seed, never over the stored rows or
// a partly applied set.
func evaluate(reg *config.Registry, environ []string, rows []Row) Evaluation {
	settings := reg.Settings()
	byKey := make(map[string]*Row, len(rows))
	var ev Evaluation
	for i := range rows {
		r := &rows[i]
		if slices.ContainsFunc(settings, func(s config.Setting) bool { return s.Key == r.Key }) {
			byKey[r.Key] = r
			continue
		}
		ev.Orphans = append(ev.Orphans, *r)
	}

	var refusals []error
	inForce := make(map[string]string, len(settings))
	outcomeIdx := make(map[string]int, len(settings))
	for _, s := range settings {
		o := outcome(s, environ, byKey[s.Key])
		outcomeIdx[s.Key] = len(ev.Outcomes)
		ev.Outcomes = append(ev.Outcomes, o)
		if o.Err != nil {
			refusals = append(refusals, o.Err)
			continue
		}
		inForce[s.Key] = o.Value
	}

	for _, rule := range reg.CorrelatedRules() {
		values := make(map[string]string, len(rule.Keys))
		for _, key := range rule.Keys {
			v, ok := inForce[key]
			if !ok {
				// The key refused on its own, so it has no value in
				// force to compare; its own refusal is reported.
				break
			}
			values[key] = v
		}
		if len(values) < len(rule.Keys) {
			continue
		}
		if err := rule.Check(values); err != nil {
			// Every key the rule names carries the refusal on its own
			// row (Task 5's doctor table): a reader of one key's row
			// alone still sees the pair the rule compared and why.
			wrapped := fmt.Errorf("correlated rule %q: %w", rule.Name, err)
			refusals = append(refusals, wrapped)
			for _, key := range rule.Keys {
				idx := outcomeIdx[key]
				ev.Outcomes[idx] = ev.Outcomes[idx].refuse(wrapped)
			}
		}
	}
	ev.Err = errors.Join(refusals...)
	return ev
}

// outcome decides s's verdict from its stored row, nil when the store
// holds none, and its environment value. Every comparison reads the
// row's value, the value in force, never its seed value. A bound
// applies to the value in force after the seed, so an out-of-bounds
// environment-seeded row that the environment re-seeds or removes is
// healed rather than refused.
func outcome(s config.Setting, environ []string, stored *Row) Outcome {
	o := Outcome{Key: s.Key, Stored: stored, setting: s}
	o.Env, o.EnvSet = lookupEnv(environ, s.Key)

	var errs []error
	if o.EnvSet {
		if err := s.Validate(o.Env); err != nil {
			errs = append(errs, fmt.Errorf("environment seed refused: %w", err))
		}
	}
	if stored != nil {
		if _, err := parse(s.Kind, stored.Value); err != nil {
			errs = append(errs, fmt.Errorf("stored %s value refused: %w", stored.Source, s.Validate(stored.Value)))
		}
		if stored.Source != sourceEnvSeed && stored.Source != sourceAdmin {
			errs = append(errs, fmt.Errorf("%s: stored row has unknown source %q", s.Key, stored.Source))
		}
	}
	if len(errs) > 0 {
		return o.refuse(errors.Join(errs...))
	}

	switch {
	case stored == nil && !o.EnvSet:
		o.Value = s.Default
		o.Report = fmt.Sprintf("no stored value; the default %s is in force", o.shown(s.Default))
	case stored == nil:
		o.Value, o.Change = o.Env, Seed
		o.Report = fmt.Sprintf("seeds %s from the environment", o.shown(o.Env))
	case stored.Source == sourceEnvSeed && !o.EnvSet:
		o.Value, o.Change = s.Default, Remove
		o.Report = fmt.Sprintf("the environment no longer sets it; returns from %s to the default %s",
			o.shown(stored.Value), o.shown(s.Default))
	case stored.Source == sourceEnvSeed && equal(s.Kind, o.Env, stored.Value):
		o.Value = stored.Value
		o.Report = fmt.Sprintf("seeded value %s is in force and the environment agrees", o.shown(stored.Value))
	case stored.Source == sourceEnvSeed:
		o.Value, o.Change = o.Env, Reseed
		o.Report = fmt.Sprintf("re-seeds from %s to %s from the environment", o.shown(stored.Value), o.shown(o.Env))
	case o.EnvSet && !equal(s.Kind, o.Env, stored.Value):
		errs = append(errs, fmt.Errorf("%s: the environment seeds %s but an administrator set %s; unset %s, or set it to %s",
			s.Key, o.shown(o.Env), o.shown(stored.Value), s.Key, o.shown(stored.Value)))
	default:
		o.Value = stored.Value
		o.Report = fmt.Sprintf("the administrator's value %s is in force", o.shown(stored.Value))
	}

	// A stored value stays in force when the seed leaves its row
	// unchanged, as it always leaves an administrator's, so its bounds
	// apply. The environment's value was validated above and the
	// default at registration.
	if stored != nil && o.Change == Unchanged {
		if err := s.Validate(stored.Value); err != nil {
			errs = append(errs, fmt.Errorf("stored %s value refused: %w", stored.Source, err))
		}
	}
	if len(errs) > 0 {
		return o.refuse(errors.Join(errs...))
	}
	return o
}

// refuse returns o refused for err.
func (o Outcome) refuse(err error) Outcome {
	o.Value, o.Change, o.Err = "", Unchanged, err
	o.Report = "refused: " + err.Error()
	return o
}

// shown returns v as a report prints it: quoted, or withheld for a
// sensitive setting.
func (o Outcome) shown(v string) string {
	if o.setting.Sensitive {
		return "(sensitive value withheld)"
	}
	return strconv.Quote(v)
}

// lookupEnv returns key's value in environ, a later entry winning over
// an earlier one, and reports whether it is set. An empty value is
// unset, the configuration wall's own rule.
func lookupEnv(environ []string, key string) (string, bool) {
	var value string
	for _, kv := range environ {
		if k, v, ok := strings.Cut(kv, "="); ok && k == key {
			value = v
		}
	}
	return value, value != ""
}

// parse reads s under kind, returning an int, a time.Duration, or a
// float64, so two values parsed under one kind compare with ==.
func parse(kind config.ThresholdKind, s string) (any, error) {
	switch kind {
	case config.KindInt:
		return strconv.Atoi(s)
	case config.KindDuration:
		return time.ParseDuration(s)
	case config.KindFloat:
		return strconv.ParseFloat(s, 64)
	default:
		return nil, fmt.Errorf("setting kind %v has no parser", kind)
	}
}

// equal reports whether a and b parse to the same value under kind,
// so "60m" equals "1h". A value that does not parse equals nothing.
func equal(kind config.ThresholdKind, a, b string) bool {
	va, errA := parse(kind, a)
	vb, errB := parse(kind, b)
	return errA == nil && errB == nil && va == vb
}
