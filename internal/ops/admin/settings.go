package admin

import (
	"context"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
)

// The homes a [ops.SettingEntry] names: the store's settings table,
// which an admin edits, or the server's environment, which only an
// operator with box access changes.
const (
	homeStore       = "store"
	homeEnvironment = "environment"
)

// applyNames and markNames are the wire forms of the registry's apply
// modes and marks.
var (
	applyNames = map[config.ApplyMode]string{
		config.ApplyImmediately:   "immediately",
		config.ApplyNextOperation: "next_operation",
		config.ApplyRestart:       "restart",
	}
	markNames = map[config.Sensitivity]string{
		config.NotSensitive: "not_sensitive",
		config.Sensitive:    "sensitive",
	}
)

// ListSettings answers every registered key: the store-homed settings
// first, each with the value in force, its source, and the seed it
// replaced, then every key only the environment sets, with no value at
// all. Those hold site paths, addresses, and credentials, so their
// values never cross the wire; each applies at the next start. A
// store-homed key the serving environment sets answers the
// environment's value and source, which the store holds in step, so the
// page can show it as set in the environment file.
func (s *Seam) ListSettings(ctx context.Context, _ ops.Actor, _ ops.ListSettingsRequest) (ops.ListSettingsResponse, error) {
	reg := s.cfg.Registry
	eval, err := settings.Evaluate(ctx, s.cfg.Store, reg, s.cfg.Environ)
	if err != nil {
		return ops.ListSettingsResponse{}, err
	}
	var entries []ops.SettingEntry
	for i, setting := range reg.Settings() {
		entry := ops.SettingEntry{
			Name: setting.Key, Home: homeStore, Description: setting.Description,
			Apply: applyNames[setting.Apply], Mark: markNames[setting.Mark],
			Value: &setting.Default, Source: new(settings.SourceDefault),
		}
		outcome := eval.Outcomes[i]
		if row := outcome.Stored; row != nil {
			entry.Value, entry.Source = &row.Value, &row.Source
			if row.SeedValue != "" {
				entry.SeedValue = &row.SeedValue
			}
		}
		if outcome.EnvSet {
			entry.Value, entry.Source = &outcome.Env, new(settings.SourceEnvironment)
		}
		entries = append(entries, entry)
	}

	seen := make(map[string]bool)
	environment := func(name, description string, mark config.Sensitivity) {
		if seen[name] {
			return
		}
		seen[name] = true
		entries = append(entries, ops.SettingEntry{
			Name: name, Home: homeEnvironment, Description: description,
			Apply: applyNames[config.ApplyRestart], Mark: markNames[mark],
		})
	}
	for _, row := range reg.OwnedWall() {
		environment(row.Name, row.Description, row.Mark)
	}
	for _, src := range reg.IdentitySources() {
		for _, k := range slices.Concat(src.Required, src.Optional) {
			environment(k.Name, k.Description, k.Mark)
		}
	}
	for _, class := range reg.IntegrationClasses() {
		for _, k := range slices.Concat([]config.KeyDecl{class.Flag}, class.Required) {
			environment(k.Name, k.Description, k.Mark)
		}
	}
	for _, th := range reg.Thresholds() {
		environment(th.Key, th.Description, th.Mark)
	}
	return ops.ListSettingsResponse{Settings: entries}, nil
}

// SetSetting writes the admin's value for a store-homed key. A
// sensitive key is gated and, once written, noticed.
func (s *Seam) SetSetting(ctx context.Context, actor ops.Actor, req ops.SetSettingRequest) (ops.SetSettingResponse, error) {
	res, err := s.writeSetting(ctx, actor, req.Key, func(ctx context.Context, by settings.Author, guard settings.Guard, now time.Time) (settings.Result, error) {
		return settings.Set(ctx, s.cfg.Store, s.cfg.Registry, s.cfg.Environ, by, guard, req.Key, req.Value, now)
	})
	if err != nil {
		return ops.SetSettingResponse{}, err
	}
	return ops.SetSettingResponse{Key: req.Key, Apply: applyNames[res.Apply], Source: res.Source}, nil
}

// ResetSetting removes the admin's value for a store-homed key. A
// sensitive key is gated and, once reset, noticed.
func (s *Seam) ResetSetting(ctx context.Context, actor ops.Actor, req ops.ResetSettingRequest) (ops.ResetSettingResponse, error) {
	res, err := s.writeSetting(ctx, actor, req.Key, func(ctx context.Context, by settings.Author, guard settings.Guard, now time.Time) (settings.Result, error) {
		return settings.Reset(ctx, s.cfg.Store, s.cfg.Registry, s.cfg.Environ, by, guard, req.Key, now)
	})
	if err != nil {
		return ops.ResetSettingResponse{}, err
	}
	return ops.ResetSettingResponse{Key: req.Key, Apply: applyNames[res.Apply], Source: res.Source}, nil
}

// writeSetting runs one settings write for actor on key: the gate when
// key is sensitive, then write under writeWait with actor as the author
// and actor's in-write admin check as the guard, then the notice once
// a sensitive write has committed. A key the registry does not hold
// carries no mark, so it reaches write ungated and is refused there.
func (s *Seam) writeSetting(ctx context.Context, actor ops.Actor, key string,
	write func(context.Context, settings.Author, settings.Guard, time.Time) (settings.Result, error),
) (settings.Result, error) {
	if s.settings[key].Mark == config.Sensitive {
		if err := s.requireFresh(ctx, actor); err != nil {
			return settings.Result{}, err
		}
	}
	by := settings.Author{Actor: actor.String(), Channel: actor.Channel().String()}
	guard := func(ctx context.Context, tx *store.Tx) error {
		return auth.CheckAdmin(ctx, tx, actor.UserID())
	}
	writeCtx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	res, err := write(writeCtx, by, guard, s.cfg.Now())
	if err != nil {
		return settings.Result{}, refusal(err)
	}
	if res.Written && res.Sensitive {
		s.notice(ctx, actor, key, res)
	}
	return res, nil
}
