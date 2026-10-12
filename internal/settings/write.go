package settings

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// The refusals [Set] and [Reset] return, each wrapped with the key and
// its cause.
var (
	// ErrNotSetting reports a key that is not a registered store-homed
	// setting: unregistered, or a wall, threshold, or integration key.
	ErrNotSetting = errors.New("not a store-homed setting")
	// ErrInvalidValue reports a write the validator refuses: a value
	// that does not parse or leaves its bounds, or a write that makes a
	// key or a correlated rule refuse where it did not before.
	ErrInvalidValue = errors.New("value refused")
	// ErrEnvironmentSets reports a write against a key the serving
	// environment sets.
	ErrEnvironmentSets = errors.New("the serving environment sets the key")
)

// The sources a [Result] names as in force.
const (
	// SourceAdmin is a value an administrator set.
	SourceAdmin = sourceAdmin
	// SourceEnvSeed is a value the seed rule wrote from the environment.
	SourceEnvSeed = sourceEnvSeed
	// SourceEnvironment is the serving environment's own value.
	SourceEnvironment = "environment"
	// SourceDefault is the registry default, no row being stored.
	SourceDefault = "default"
)

// A Guard runs inside a write's transaction, before anything is read
// or written, and refuses the write by returning an error. It lets the
// caller re-check, in the write's own transaction, a fact the write
// depends on, such as that the actor is still an administrator. A nil
// Guard checks nothing.
type Guard func(ctx context.Context, tx *store.Tx) error

// An Author is who a write is recorded as: the audit row's actor, in
// its wire form, and the channel it arrived over.
type Author struct {
	Actor   string
	Channel string
}

// A Result is what [Set] or [Reset] did, or why it did nothing.
type Result struct {
	// Written reports whether the store changed. A write that finds its
	// answer already in force changes nothing and leaves it false.
	Written bool
	// Source is where the value in force now comes from: one of the
	// Source constants.
	Source string
	// Value is the value in force now, in string form.
	Value string
	// Apply is when the setting's change takes effect.
	Apply config.ApplyMode
	// Sensitive reports whether the setting is marked sensitive.
	Sensitive bool
}

// A change is the one row edit a write decides on.
type change struct {
	kind string
	old  *Row // the stored row, or nil
	next *Row // the row after, or nil to delete it
}

// Set writes value as the administrator's value for the store-homed
// setting key, in one transaction: the row with source admin, the
// replaced seed in seed_value, one audit row, and a generation bump. A
// write over an env-seed row records the seed it replaced; over an
// admin row it keeps the earlier seed.
//
// The value must parse and sit within the key's bounds, and the
// serving environment, environ as [os.Environ] lists it, must not set
// the key to another value; a value equal to the environment's writes
// nothing and answers with [SourceEnvironment]. Then the seed rule runs
// over the rows before and after the write, and the write is refused
// when it makes any key refuse that did not refuse before, so a write
// cannot add a refusal the next start would raise.
//
// guard, when non-nil, runs first inside the transaction. Every read the
// decision rests on runs through that transaction, so two writes that
// are each valid alone cannot both commit. ctx must carry a deadline,
// as [store.Handle.Update] requires, and now stamps the rows written.
func Set(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string, by Author, guard Guard, key, value string, now time.Time) (Result, error) {
	s, err := lookup(reg, key)
	if err != nil {
		return Result{}, err
	}
	if err := s.Validate(value); err != nil {
		return Result{}, fmt.Errorf("%w: %w", ErrInvalidValue, err)
	}
	res := Result{Source: SourceAdmin, Value: value}
	err = write(ctx, h, reg, environ, by, guard, key, now, func(stored *Row) (*change, error) {
		if envValue, set := lookupEnv(environ, key); set {
			if !equal(s.Kind, envValue, value) {
				return nil, fmt.Errorf("%w: %s", ErrEnvironmentSets, key)
			}
			res.Source, res.Value = SourceEnvironment, envValue
			return nil, nil
		}
		var seed string
		if stored != nil {
			seed = stored.SeedValue
		}
		res.Written = true
		return &change{"set", stored, &Row{Key: key, Value: value, Source: sourceAdmin, SeedValue: seed}}, nil
	})
	return describe(res, s), err
}

// Reset deletes the administrator's row for the store-homed setting
// key, in one transaction with one audit row (its new value empty) and
// a generation bump, under the validator [Set] describes. Over no row
// or an env-seed row it writes nothing and answers with the source in
// force. Over an admin row for a key the serving environment sets it is
// refused with [ErrEnvironmentSets], since the key would read its
// default until the next start seeded the environment's value again.
//
// guard, ctx, and now are as for [Set].
func Reset(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string, by Author, guard Guard, key string, now time.Time) (Result, error) {
	s, err := lookup(reg, key)
	if err != nil {
		return Result{}, err
	}
	res := Result{Source: SourceDefault, Value: s.Default}
	err = write(ctx, h, reg, environ, by, guard, key, now, func(stored *Row) (*change, error) {
		switch {
		case stored == nil:
			return nil, nil
		case stored.Source == sourceEnvSeed:
			res.Source, res.Value = SourceEnvSeed, stored.Value
			return nil, nil
		case stored.Source != sourceAdmin:
			return nil, fmt.Errorf("%s: stored row has unknown source %q", key, stored.Source)
		}
		if _, set := lookupEnv(environ, key); set {
			return nil, fmt.Errorf("%w: %s", ErrEnvironmentSets, key)
		}
		res.Written = true
		return &change{"reset", stored, nil}, nil
	})
	return describe(res, s), err
}

// lookup returns the registered setting named key, or an error
// wrapping [ErrNotSetting].
func lookup(reg *config.Registry, key string) (config.Setting, error) {
	i := slices.IndexFunc(reg.Settings(), func(s config.Setting) bool { return s.Key == key })
	if i < 0 {
		return config.Setting{}, fmt.Errorf("%w: %s", ErrNotSetting, key)
	}
	return reg.Settings()[i], nil
}

// describe completes res with s's apply mode and mark.
func describe(res Result, s config.Setting) Result {
	res.Apply, res.Sensitive = s.Apply, s.Mark == config.Sensitive
	return res
}

// write runs one write's transaction: the guard, the rows read through
// the transaction, decide's verdict on key's stored row, the
// before-and-after evaluation, then the row edit, its audit row, and
// the generation bump. decide returns nil for a write that changes
// nothing.
func write(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string, by Author, guard Guard, key string, now time.Time, decide func(stored *Row) (*change, error)) error {
	return h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if guard != nil {
			if err := guard(ctx, tx); err != nil {
				return err
			}
		}
		rows, err := readTxRows(ctx, tx)
		if err != nil {
			return err
		}
		var stored *Row
		if i := slices.IndexFunc(rows, func(r Row) bool { return r.Key == key }); i >= 0 {
			stored = &rows[i]
		}
		c, err := decide(stored)
		if err != nil || c == nil {
			return err
		}
		if err := refuseNewRefusals(reg, environ, rows, key, c.next); err != nil {
			return err
		}
		return commit(ctx, tx, by, key, now, c)
	})
}

// refuseNewRefusals evaluates the seed rule over rows and over rows
// with key's row replaced by next (removed when next is nil), and
// returns an error wrapping [ErrInvalidValue] for each key the second
// evaluation refuses and the first did not. It compares each key's
// own [Outcome].Err: [Evaluation].Err joins every key's refusal, so a
// store that already refuses one key would refuse every write.
func refuseNewRefusals(reg *config.Registry, environ []string, rows []Row, key string, next *Row) error {
	after := slices.DeleteFunc(slices.Clone(rows), func(r Row) bool { return r.Key == key })
	if next != nil {
		after = append(after, *next)
	}
	before, afterOutcomes := evaluate(reg, environ, rows).Outcomes, evaluate(reg, environ, after).Outcomes

	var refusals []error
	for i, o := range afterOutcomes {
		// A correlated rule's refusal is one error shared by every key
		// the rule names, so it is reported once.
		if o.Err != nil && before[i].Err == nil && !slices.Contains(refusals, o.Err) {
			refusals = append(refusals, o.Err)
		}
	}
	if len(refusals) == 0 {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrInvalidValue, errors.Join(refusals...))
}

// commit writes c's row edit, its audit row, and the generation bump
// through tx.
func commit(ctx context.Context, tx *store.Tx, by Author, key string, now time.Time, c *change) error {
	at := now.UnixMicro()
	// oldValue and newValue stay nil, stored as NULL, for a set over no
	// row and a reset to none.
	var oldValue, newValue any
	if c.old != nil {
		oldValue = c.old.Value
	}
	var err error
	if c.next == nil {
		_, err = tx.Exec(ctx, `DELETE FROM settings WHERE key = ? AND source = 'admin'`, key)
	} else {
		newValue = c.next.Value
		var seed any
		if c.next.SeedValue != "" {
			seed = c.next.SeedValue
		}
		_, err = tx.Exec(ctx, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
			VALUES (?, ?, 'admin', ?, ?, ?)
			ON CONFLICT (key) DO UPDATE SET value = excluded.value, source = excluded.source,
				seed_value = excluded.seed_value, actor = excluded.actor, updated_at = excluded.updated_at`,
			key, c.next.Value, seed, by.Actor, at)
	}
	if err != nil {
		return fmt.Errorf("%s: write %s: %w", key, c.kind, err)
	}
	if err := insertAudit(ctx, tx, at, c.kind, by.Actor, by.Channel, key, oldValue, newValue); err != nil {
		return fmt.Errorf("%s: audit %s: %w", key, c.kind, err)
	}
	return bumpGeneration(ctx, tx)
}
