package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/spf13/cobra"
)

// memberVerbWriteWait bounds how long a write waits for the roster's
// directory lock (R-5c19): 30 seconds. It is a package variable rather
// than a constant so a test can shorten it, on the lock row's own test
// seam; production never assigns it a second value.
var memberVerbWriteWait = 30 * time.Second

// osUserLookupID is os/user.LookupId, held as a package variable so a
// test can substitute a lookup that always fails, proving the
// uid:<n> fallback without depending on whether the test's passwd
// database happens to resolve the effective uid.
var osUserLookupID = user.LookupId

// errForeignRosterOwner reports that the verb's effective uid does not
// own the roster file's containing directory (R-5c14). Writing there
// would create a roster the serving unit, running as the directory's
// owner, cannot read, which fails the server closed at its next start
// under R2.
var errForeignRosterOwner = errors.New("member: roster directory owned by a different user")

// newMemberCmd builds the member verb: add and list, the two
// subcommands rung 5c-i gives an operator for the roster
// internal/auth keeps (R37). Neither subcommand reads the environment
// itself; both resolve configuration through config.Load against reg,
// the same wall serve enforces.
func newMemberCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "member",
		Short: "Manage the roster of group members",
	}
	cmd.AddCommand(newMemberAddCmd(environ, reg, stdout, stderr))
	cmd.AddCommand(newMemberListCmd(environ, reg, stdout, stderr))
	return cmd
}

// newMemberAddCmd builds "member add <address>": it loads
// configuration, resolves the roster path the selected identity
// source declares, refuses a foreign directory owner, then creates the
// roster when it is absent or adds to it when it exists (R-5c8,
// R-5c14). --name and --username are required flags; an
// empty value still reaches the roster's own write-path validation,
// since cobra's required check proves only that the flag was set.
func newMemberAddCmd(environ []string, reg *config.Registry, _, stderr io.Writer) *cobra.Command {
	var name, username string
	var admin bool

	cmd := &cobra.Command{
		Use:           "add <address>",
		Short:         "Add a member to the roster",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			rosterPath, ok := selectedRosterPath(environ, reg)
			if !ok {
				return fmt.Errorf("member: identity source %q requires no roster", cfg.IdentitySource)
			}
			if err := checkRosterDirOwner(filepath.Dir(rosterPath)); err != nil {
				return err
			}

			candidate := auth.Member{
				Address:     args[0],
				DisplayName: name,
				Username:    username,
				Admin:       admin,
			}
			actor := resolveActor()
			logger := slog.New(slog.NewTextHandler(stderr, nil))

			ctx, cancel := context.WithTimeout(cmd.Context(), memberVerbWriteWait)
			defer cancel()
			return addMember(ctx, rosterPath, candidate, actor, logger)
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

// addMember opens the roster at path and adds candidate under actor,
// creating the roster when it is absent (R-5c8). When CreateRoster
// loses a concurrent bootstrap race, it reports ErrRosterExists; this
// opens the winner's roster and adds through AddMember once, so two
// concurrent first adds both land and the verb never overwrites a
// roster.
func addMember(ctx context.Context, path string, candidate auth.Member, actor string, logger *slog.Logger) error {
	roster, err := auth.OpenRoster(path)
	if errors.Is(err, auth.ErrRosterMissing) {
		_, createErr := auth.CreateRoster(ctx, path, candidate, actor, logger)
		if errors.Is(createErr, auth.ErrRosterExists) {
			roster, err = auth.OpenRoster(path)
			if err != nil {
				return err
			}
			return roster.AddMember(ctx, candidate, actor, identity.ChannelCommandLine, "", logger)
		}
		return createErr
	}
	if err != nil {
		return err
	}
	return roster.AddMember(ctx, candidate, actor, identity.ChannelCommandLine, "", logger)
}

// newMemberListCmd builds "member list": it loads configuration,
// opens the roster read-only, and prints one line per member in file
// order to stdout, which is the subcommand's whole purpose. Nothing
// else this file writes carries an address.
func newMemberListCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "list",
		Short:         "List the roster's members",
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			rosterPath, ok := selectedRosterPath(environ, reg)
			if !ok {
				return fmt.Errorf("member: identity source %q requires no roster", cfg.IdentitySource)
			}
			roster, err := auth.OpenRoster(rosterPath)
			if err != nil {
				return err
			}
			members, err := roster.List()
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
// per member in file order, address, display name, username,
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

// checkRosterDirOwner refuses when the effective uid does not own dir,
// naming both ids and dir (R-5c14). io/fs.FileInfo's doc comment
// states "Sys() any // underlying data source (can return nil)"; on
// this platform, the value os.Stat returns carries a *syscall.Stat_t.
func checkRosterDirOwner(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		return fmt.Errorf("member: roster directory: %w", err)
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return fmt.Errorf("member: roster directory: cannot determine its owner")
	}
	euid := os.Geteuid()
	if stat.Uid != uint32(euid) {
		return fmt.Errorf("%w: effective uid %d, directory %s owned by uid %d", errForeignRosterOwner, euid, dir, stat.Uid)
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
