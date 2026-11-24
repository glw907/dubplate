package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// tusRouteName is the tus endpoint's subtree name. It cannot live under
// api, whose mount answers every path beneath it.
const tusRouteName = "tus"

// newIntake builds the server's intake over h for the staging directory
// cfg names: its limits and windows read through acc at each use, its
// sweep interval and worker count read once here, and its stall mail
// sent through send to the live admins roster lists.
func newIntake(ctx context.Context, h *store.Handle, cfg *config.Config, acc *settings.Accessor, allowlist []string,
	roster *auth.Roster, send func(ctx context.Context, to, subject, body string) error, logger *slog.Logger) (*intake.Intake, error) {
	interval, err := intake.SweepInterval(ctx, acc)
	if err != nil {
		return nil, err
	}
	workers, err := intake.CheckWorkers(ctx, acc)
	if err != nil {
		return nil, err
	}
	return intake.New(intake.Config{
		Store: h, Log: logger, Staging: cfg.WallValue(intake.StagingKey), Now: time.Now,
		EarliestAdmin: auth.EarliestAdmin,
		Limits:        intake.LimitsFrom(acc, allowlist),
		Windows:       intake.WindowsFrom(acc),
		SweepInterval: interval,
		CheckWorkers:  workers,
		Admins:        liveAdminAddresses(roster),
		Send:          send,
	})
}

// liveAdminAddresses reads the stall mail's recipients from roster at
// each call: every live member holding the admin flag, as the
// sensitive-change notice reads them, so a demoted or removed admin
// hears nothing.
func liveAdminAddresses(roster *auth.Roster) func(context.Context) ([]string, error) {
	return func(ctx context.Context) ([]string, error) {
		members, err := roster.List(ctx)
		if err != nil {
			return nil, err
		}
		var admins []string
		for _, m := range members {
			if m.Admin {
				admins = append(admins, m.Address)
			}
		}
		return admins, nil
	}
}

// tusVerifier admits a tus request whose session verifies as any
// member, through the same check the API's verifier makes. A request
// signed out wraps [ops.ErrUnauthenticated], which the mount answers
// 401; any other failure is logged here and answers 500.
func tusVerifier(cfg auth.AdminConfig) intake.Verifier {
	return func(w http.ResponseWriter, r *http.Request) (identity.UserID, error) {
		v, err := auth.CheckRequest(cfg, w, r)
		switch {
		case errors.Is(err, ops.ErrUnauthenticated):
			return identity.UserID{}, err
		case err != nil:
			cfg.Logger.Error("tus request could not verify its session", "error", err)
			return identity.UserID{}, err
		}
		return v.Identity.UserID, nil
	}
}
