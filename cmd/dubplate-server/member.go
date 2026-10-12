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
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/spf13/cobra"
)

// memberVerbWriteWait bounds how long a verb waits to open the store
// and, for add, to write the member: 30 seconds. It is a package
// variable rather than a constant so a test can shorten it, on the
// wait's own test seam; production never assigns it a second value.
var memberVerbWriteWait = 30 * time.Second

// osUserLookupID is os/user.LookupId, held as a package variable so a
// test can substitute a lookup that always fails, proving the
// uid:<n> fallback without depending on whether the test's passwd
// database happens to resolve the effective uid.
var osUserLookupID = user.LookupId

// newMemberCmd builds the member verb: add and list, the two
// subcommands that give an operator the roster internal/auth keeps in
// the store. Neither subcommand reads the environment itself; both
// resolve configuration through config.Load against reg, the same wall
// serve enforces.
func newMemberCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage the roster of group members",
	}
	cmd.AddCommand(newMemberAddCmd(environ, reg, stdout, stderr))
	cmd.AddCommand(newMemberListCmd(environ, reg, stdout, stderr))
	return cmd
}

// openMemberStore loads the wall and opens the store in write mode, the
// mode both verbs share: it refuses an absent store, a store at any
// version but the binary's, a restore in progress, and a leaf another
// user owns, each with the store's own message. Add writes through the
// handle, and list opens the same mode so one set of refusals covers
// both verbs. It also returns the state directory the store lives
// under. The caller closes the handle.
func openMemberStore(ctx context.Context, environ []string, reg *config.Registry) (*store.Handle, string, error) {
	cfg, err := config.Load(environ, reg)
	if err != nil {
		return nil, "", err
	}
	busyTimeout, err := storeBusyTimeout(environ)
	if err != nil {
		return nil, "", err
	}
	h, err := store.OpenWrite(ctx, cfg.StateDir, busyTimeout)
	return h, cfg.StateDir, err
}

// writingRoster returns a roster over h that may write: it opens the
// auth leaf under stateDir for the revocation epochs a removal and a
// re-add bump. A leaf auth.Open refuses is reported naming verb, and
// nothing is written. Only a writing verb calls it, so member list
// opens no auth leaf.
func writingRoster(h *store.Handle, stateDir, verb string) (*auth.Roster, error) {
	leaf, err := auth.Open(stateDir)
	if err != nil {
		return nil, fmt.Errorf("member %s: %w", verb, err)
	}
	return auth.NewRoster(h, auth.OpenEpochStore(leaf), time.Now), nil
}

// newMemberAddCmd builds "member add <address>": it opens the store in
// write mode and adds the member through the roster's one write path.
// --name and --username are required flags; an empty value still
// reaches the roster's own write-path validation, since cobra's
// required check proves only that the flag was set.
func newMemberAddCmd(environ []string, reg *config.Registry, _, stderr io.Writer) *cobra.Command {
	var name, username string
	var admin bool

	cmd := &cobra.Command{
		Use:           "add <address>",
		Short:         "Add a member to the roster",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, cancel := context.WithTimeout(cmd.Context(), memberVerbWriteWait)
			defer cancel()
			h, stateDir, err := openMemberStore(ctx, environ, reg)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, h.Close()) }()
			roster, err := writingRoster(h, stateDir, "add")
			if err != nil {
				return err
			}

			candidate := auth.Member{
				Address:     args[0],
				DisplayName: name,
				Username:    username,
				Admin:       admin,
			}
			logger := slog.New(slog.NewTextHandler(stderr, nil))
			return roster.AddMember(ctx, candidate, auth.OperatorActor(resolveActor()), logger)
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "the member's display name")
	cmd.Flags().StringVar(&username, "username", "", "the member's username")
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

// newMemberListCmd builds "member list": it opens the store and prints
// one line per live member, in the order they were added, to stdout,
// which is the subcommand's whole purpose. Nothing else this file
// writes carries an address.
func newMemberListCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List the roster's members",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			openCtx, cancel := context.WithTimeout(cmd.Context(), memberVerbWriteWait)
			defer cancel()
			h, _, err := openMemberStore(openCtx, environ, reg)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, h.Close()) }()

			members, err := auth.NewRoster(h, nil, nil).List(cmd.Context())
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
func printMemberList(w io.Writer, members []auth.Member) error {
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
// as the verb's audit actor (R-5c14): user.LookupId(effective uid),
// falling back to uid:<n> when the lookup fails, never $USER, which
// anyone who controls the environment a verb runs under can forge.
func resolveActor() string {
	euid := os.Geteuid()
	u, err := osUserLookupID(strconv.Itoa(euid))
	if err != nil {
		return fmt.Sprintf("uid:%d", euid)
	}
	return u.Username
}
