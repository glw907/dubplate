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

// operatorVerbs carries what the member and settings verbs take from the
// process, so a test builds them with values of its choosing. main fills
// it with [storeWait] and os/user.LookupId.
type operatorVerbs struct {
	// WriteWait bounds how long a verb waits to open the store and, for
	// member add, to write the member.
	WriteWait time.Duration

	// LookupID resolves the effective uid to the user an audit record
	// names; see [resolveActor].
	LookupID func(uid string) (*user.User, error)
}

// newMemberCmd builds the member verb: add, remove, demote, and list,
// the subcommands that give an operator the group's members, which
// internal/auth keeps in the store. None reads the environment itself; each resolves
// configuration through loadConfig against reg, the same wall serve
// enforces.
func (ov operatorVerbs) newMemberCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage the group's members",
	}
	cmd.AddCommand(ov.newMemberAddCmd(environ, reg, stderr))
	cmd.AddCommand(ov.newMemberRemoveCmd(environ, reg, stderr))
	cmd.AddCommand(ov.newMemberDemoteCmd(environ, reg, stderr))
	cmd.AddCommand(ov.newMemberListCmd(environ, reg, stdout, stderr))
	return cmd
}

// openOperatorStore loads the wall and opens the store in write mode, the
// mode both verbs share: it refuses an absent store, a store at any
// version but the binary's, a restore in progress, and a leaf another
// user owns, each with the store's own message. Add writes through the
// handle, and list opens the same mode so one set of refusals covers
// both verbs. It also returns the state directory the store lives
// under. The caller closes the handle.
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

// withStore runs fn over the store opened in write mode under the verb's
// write wait, and closes the handle after. fn's context carries that
// wait.
func (ov operatorVerbs) withStore(ctx context.Context, environ []string, reg *config.Registry, fn func(ctx context.Context, h *store.Handle, stateDir string) error) (err error) {
	ctx, cancel := context.WithTimeout(ctx, ov.WriteWait)
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
	return member.NewRoster(h, auth.OpenEpochStore(leaf), time.Now), nil
}

// newMemberAddCmd builds "member add <address>": it opens the store in
// write mode and adds the member through the roster's one write path.
// --name and --username are required flags; an empty value still
// reaches the roster's write-path validation, since cobra's required
// check proves only that the flag was set.
func (ov operatorVerbs) newMemberAddCmd(environ []string, reg *config.Registry, stderr io.Writer) *cobra.Command {
	var name, username string
	var admin bool

	cmd := &cobra.Command{
		Use:           "add <address>",
		Short:         "Add a member to the group",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ov.withStore(cmd.Context(), environ, reg, func(ctx context.Context, h *store.Handle, stateDir string) error {
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
				logger := slog.New(slog.NewTextHandler(stderr, nil))
				_, err = roster.AddMember(ctx, candidate, identity.LocalActor(resolveActor(ov.LookupID)), logger)
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
	cmd.SetErr(stderr)
	return cmd
}

// A rosterChange is a roster write on one member, named by id.
type rosterChange func(r *member.Roster, ctx context.Context, id identity.UserID, by identity.Actor, logger *slog.Logger) error

// newMemberRemoveCmd builds "member remove <address>": it tombstones
// the member, deletes their preferences, and revokes their sessions.
func (ov operatorVerbs) newMemberRemoveCmd(environ []string, reg *config.Registry, stderr io.Writer) *cobra.Command {
	return ov.newMemberChangeCmd("remove", "Remove a member from the group", environ, reg, stderr, (*member.Roster).RemoveMember)
}

// newMemberDemoteCmd builds "member demote <address>": it clears the
// member's admin flag.
func (ov operatorVerbs) newMemberDemoteCmd(environ []string, reg *config.Registry, stderr io.Writer) *cobra.Command {
	return ov.newMemberChangeCmd("demote", "Clear a member's admin flag", environ, reg, stderr, (*member.Roster).DemoteMember)
}

// newMemberChangeCmd builds the verb that resolves an address to the
// live member's id and applies one roster write to it. The write names
// the member by id, so one removed between the lookup and the write
// answers [member.ErrMemberNotFound], as an unknown address does. The
// audit record goes to stderr.
func (ov operatorVerbs) newMemberChangeCmd(verb, short string, environ []string, reg *config.Registry, stderr io.Writer, change rosterChange) *cobra.Command {
	cmd := &cobra.Command{
		Use:           verb + " <address>",
		Short:         short,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return ov.withStore(cmd.Context(), environ, reg, func(ctx context.Context, h *store.Handle, stateDir string) error {
				roster, err := writingRoster(h, stateDir, verb)
				if err != nil {
					return err
				}
				target, err := roster.Lookup(ctx, args[0])
				if err != nil {
					return err
				}
				logger := slog.New(slog.NewTextHandler(stderr, nil))
				return change(roster, ctx, target.ID, identity.LocalActor(resolveActor(ov.LookupID)), logger)
			})
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// newMemberListCmd builds "member list": it opens the store and prints
// one line per live member, in the order they were added, to stdout,
// which is the subcommand's whole purpose. Nothing else this file
// writes carries an address.
func (ov operatorVerbs) newMemberListCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List the group's members",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			openCtx, cancel := context.WithTimeout(cmd.Context(), ov.WriteWait)
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
			return printMemberList(stdout, members)
		},
	}
	cmd.SetErr(stderr)
	return cmd
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
// as the verb's audit actor: lookup(effective uid), falling back to
// uid:<n> when the lookup fails, never $USER, which anyone who controls
// the environment a verb runs under can forge.
func resolveActor(lookup func(uid string) (*user.User, error)) string {
	euid := os.Geteuid()
	u, err := lookup(strconv.Itoa(euid))
	if err != nil {
		return fmt.Sprintf("uid:%d", euid)
	}
	return u.Username
}
