package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/store"
	"github.com/spf13/cobra"
)

// lookupUser resolves the effective uid to the user an audit record
// names; see [resolveActor].
var lookupUser = user.LookupId

// newMemberCmd builds the member verb: add, remove, demote, and list,
// the subcommands that give an operator the group's members, which
// internal/member keeps in the store. None reads the environment itself; each resolves
// configuration through loadConfig against reg, the same wall serve
// enforces.
func newMemberCmd(environ []string, reg *config.Registry) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage the group's members",
	}
	cmd.AddCommand(
		newMemberAddCmd(environ, reg),
		newMemberRemoveCmd(environ, reg),
		newMemberDemoteCmd(environ, reg),
		newMemberListCmd(environ, reg),
	)
	return cmd
}

// openOperatorStore loads the wall and opens the store in write mode, the
// mode every operator verb shares: it refuses an absent store, a store at
// any version but the binary's, a restore in progress, and a leaf another
// user owns, each with the store's own message. A verb that only reads
// opens the same mode, so one set of refusals covers them all. It also
// returns the state directory the store lives under. The caller closes
// the handle.
func openOperatorStore(ctx context.Context, environ []string, reg *config.Registry) (*store.Handle, string, error) {
	cfg, err := loadConfig(environ, reg)
	if err != nil {
		return nil, "", err
	}
	busyTimeout, err := store.BusyTimeout(environ)
	if err != nil {
		return nil, "", err
	}
	h, err := store.OpenWrite(ctx, cfg.StateDir, busyTimeout)
	return h, cfg.StateDir, err
}

// withStore runs fn over the store opened in write mode under
// [storeWait], and closes the handle after. fn's context carries that
// wait.
func withStore(ctx context.Context, environ []string, reg *config.Registry, fn func(ctx context.Context, h *store.Handle, stateDir string) error) (err error) {
	ctx, cancel := context.WithTimeout(ctx, storeWait)
	defer cancel()
	h, stateDir, err := openOperatorStore(ctx, environ, reg)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, h.Close()) }()
	return fn(ctx, h, stateDir)
}

// writingRoster returns a roster over h that may write: it opens the
// auth leaf under stateDir for the revocation epochs a removal and a
// re-add bump. A leaf auth.Open refuses is reported naming verb, and
// nothing is written. Only a writing verb calls it, so member list
// opens no auth leaf.
func writingRoster(h *store.Handle, stateDir, verb string) (*member.Roster, error) {
	leaf, err := auth.Open(stateDir)
	if err != nil {
		return nil, fmt.Errorf("member %s: %w", verb, err)
	}
	return member.NewRoster(h, auth.NewEpochStore(leaf), time.Now), nil
}

// newMemberAddCmd builds "member add <address>": it opens the store in
// write mode and adds the member through the roster's one write path.
// --name and --username are required flags; an empty value still
// reaches the roster's write-path validation, since cobra's required
// check proves only that the flag was set.
func newMemberAddCmd(environ []string, reg *config.Registry) *cobra.Command {
	var name, username string
	var admin bool

	cmd := &cobra.Command{
		Use:   "add <address>",
		Short: "Add a member to the group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(cmd.Context(), environ, reg, func(ctx context.Context, h *store.Handle, stateDir string) error {
				roster, err := writingRoster(h, stateDir, "add")
				if err != nil {
					return err
				}
				candidate := member.Member{
					Address:     args[0],
					DisplayName: name,
					Username:    username,
					Admin:       admin,
				}
				logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))
				_, err = roster.AddMember(ctx, candidate, identity.LocalActor(resolveActor()), logger)
				return err
			})
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "the member's display name (required)")
	cmd.Flags().StringVar(&username, "username", "", "the member's username (required)")
	cmd.Flags().BoolVar(&admin, "admin", false, "grant the member the admin flag")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("username"); err != nil {
		panic(err)
	}
	return cmd
}

// A rosterChange is a roster write on one member, named by id.
type rosterChange func(r *member.Roster, ctx context.Context, id identity.UserID, by identity.Actor, logger *slog.Logger) error

// newMemberRemoveCmd builds "member remove <address>": it tombstones
// the member, deletes their preferences, and revokes their sessions.
func newMemberRemoveCmd(environ []string, reg *config.Registry) *cobra.Command {
	return newMemberChangeCmd("remove", "Remove a member from the group", environ, reg, (*member.Roster).RemoveMember)
}

// newMemberDemoteCmd builds "member demote <address>": it clears the
// member's admin flag.
func newMemberDemoteCmd(environ []string, reg *config.Registry) *cobra.Command {
	return newMemberChangeCmd("demote", "Clear a member's admin flag", environ, reg, (*member.Roster).DemoteMember)
}

// newMemberChangeCmd builds the verb that resolves an address to the
// live member's id and applies one roster write to it. The write names
// the member by id, so one removed between the lookup and the write
// answers [member.ErrMemberNotFound], as an unknown address does. The
// audit record goes to stderr.
func newMemberChangeCmd(verb, short string, environ []string, reg *config.Registry, change rosterChange) *cobra.Command {
	return &cobra.Command{
		Use:   verb + " <address>",
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(cmd.Context(), environ, reg, func(ctx context.Context, h *store.Handle, stateDir string) error {
				roster, err := writingRoster(h, stateDir, verb)
				if err != nil {
					return err
				}
				target, err := roster.Lookup(ctx, args[0])
				if err != nil {
					return err
				}
				logger := slog.New(slog.NewTextHandler(cmd.ErrOrStderr(), nil))
				return change(roster, ctx, target.ID, identity.LocalActor(resolveActor()), logger)
			})
		},
	}
}

// newMemberListCmd builds "member list": it opens the store and prints
// one line per live member, in the order they were added, to stdout,
// which is the subcommand's whole purpose. Nothing else this file
// writes carries an address.
func newMemberListCmd(environ []string, reg *config.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List the group's members",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			openCtx, cancel := context.WithTimeout(cmd.Context(), storeWait)
			defer cancel()
			h, _, err := openOperatorStore(openCtx, environ, reg)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, h.Close()) }()

			members, err := member.NewRoster(h, nil, nil).List(cmd.Context())
			if err != nil {
				return err
			}
			return printMemberList(cmd.OutOrStdout(), members)
		},
	}
}

// printMemberList writes members to w: a header line, then one line
// per member in the order given, address, display name, username,
// and an admin marker, tab-separated.
func printMemberList(w io.Writer, members []member.Member) error {
	if _, err := fmt.Fprint(w, "ADDRESS\tDISPLAY NAME\tUSERNAME\tADMIN\n"); err != nil {
		return err
	}
	for _, m := range members {
		marker := ""
		if m.Admin {
			marker = "yes"
		}
		if _, err := fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", m.Address, m.DisplayName, m.Username, marker); err != nil {
			return err
		}
	}
	return nil
}

// resolveActor returns the operating-system user of the effective uid
// as the verb's audit actor, falling back to uid:<n> when the lookup
// fails, never $USER, which anyone who controls the environment a verb
// runs under can forge.
func resolveActor() string {
	euid := os.Geteuid()
	u, err := lookupUser(strconv.Itoa(euid))
	if err != nil {
		return fmt.Sprintf("uid:%d", euid)
	}
	return u.Username
}
