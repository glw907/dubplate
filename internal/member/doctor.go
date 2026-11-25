package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/store"
)

// usersVersion is the lowest store version holding the users table
// (0003_catalog.sql).
const usersVersion = 3

// adminEmailKey is the owned wall row naming the operator's address.
// The environment only seeds the first admin: the roster alone decides
// who is one, so the admin email row warns when the address no longer
// names a live admin.
const adminEmailKey = "DUBPLATE_ADMIN_EMAIL"

// The two warnings the admin email row can carry. A removed member and
// an address no member holds share the first, since one action fixes
// both: member add with --admin reactivates a removed row, and the
// variable may instead name a live admin. A live non-admin has no verb
// to promote it, since member add refuses a live address, so the only
// action is to name a live admin.
const (
	adminEmailNotMember = "warning: this address is not a live member; run dubplate-server member add <address> --admin, or set " + adminEmailKey + " to a live admin"
	adminEmailNotAdmin  = "warning: this address is a live member without the admin flag; set " + adminEmailKey + " to a live admin"
)

// standing is where one address stands in the roster.
type standing int

const (
	// standingNotRead is the zero value: no address was named, or the
	// store holds no users table to look it up in.
	standingNotRead standing = iota

	// standingNone is an address no users row holds.
	standingNone

	// standingRemoved is an address whose row is tombstoned, whatever
	// its admin flag.
	standingRemoved

	// standingMember is an address held by a live member without the
	// admin flag.
	standingMember

	// standingAdmin is an address held by a live admin.
	standingAdmin
)

// Facts are the roster facts a store inspection reads for the doctor
// command: the live member and admin counts, and where one address
// stands.
type Facts struct {
	members, admins int
	standing        standing
}

// Reader returns the [store.Reader] that fills f from the users table,
// looking up the standing of the address environ names as the admin
// email. An unset or unparseable address names no one, so the reader
// looks none up.
func (f *Facts) Reader(environ []string) store.Reader {
	value, _ := config.Lookup(environ, adminEmailKey)
	address, _ := identity.NormalizeAddress(value)
	return func(ctx context.Context, q store.Querier, version int) error {
		if version < usersVersion {
			return nil
		}
		err := q.QueryRowContext(ctx, `SELECT count(*), coalesce(sum(admin), 0) FROM users WHERE tombstoned_at IS NULL`).
			Scan(&f.members, &f.admins)
		if err != nil {
			return fmt.Errorf("count members: %w", err)
		}
		if address == "" {
			return nil
		}
		var admin, removed bool
		err = q.QueryRowContext(ctx, `SELECT admin, tombstoned_at IS NOT NULL FROM users WHERE address = ?`, address).
			Scan(&admin, &removed)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			f.standing = standingNone
		case err != nil:
			return fmt.Errorf("read the admin address's standing: %w", err)
		case removed:
			f.standing = standingRemoved
		case admin:
			f.standing = standingAdmin
		default:
			f.standing = standingMember
		}
		return nil
	}
}

// membersRowName is the name of the row reporting the members the
// store holds.
const membersRowName = "members"

// DoctorRows renders the members row from report and the roster facts
// f its inspection read. The row checks the report's StoreVersion
// before Facts: a version 1 store carries no facts, while a version 2
// store carries facts without the users table. A store the store row cannot vouch for reads "not
// read"; the row fails only for a store that holds the users table and
// no live admin, since nobody could then add a member through the
// page.
func DoctorRows(report store.Report, f Facts) []doctor.Row {
	row := doctor.Row{Name: membersRowName}
	row.Status, row.Failed = membersStatus(report, f)
	return []doctor.Row{row}
}

func membersStatus(r store.Report, f Facts) (status string, failed bool) {
	switch {
	case r.State == store.StateAbsent:
		return "no store yet; start serve, then member add --admin", false
	case !r.State.Vouches():
		return "not read", false
	case r.StoreVersion < usersVersion:
		return fmt.Sprintf("members not recorded before store version %d", usersVersion), false
	case r.Facts == nil:
		return "not read", false
	case f.members == 0:
		return "no members; run dubplate-server member add --admin", true
	case f.admins == 0:
		return fmt.Sprintf("members: %d, admins: 0; run dubplate-server member add --admin", f.members), true
	}
	return fmt.Sprintf("members: %d, admins: %d", f.members, f.admins), false
}

// RefineAdminEmail refines row, one of the config wall's rows, from
// report and the roster facts f its inspection read, and returns any
// other row unchanged. The admin email row's status stands unless it
// is "ok". An "ok" stays when the store cannot speak for the roster
// (not vouched for, absent, older than the users table)
// and when the roster holds no live member, which the members row
// already fails. Otherwise the address's standing decides: a live
// admin is "ok", and any other standing is a warning in the row's
// status, which never fails doctor.
func RefineAdminEmail(row doctor.Row, report store.Report, f Facts) doctor.Row {
	if row.Name == adminEmailKey {
		row.Status = adminEmailStatus(row.Status, report, f)
	}
	return row
}

func adminEmailStatus(wall string, report store.Report, f Facts) string {
	switch {
	case wall != "ok", !report.State.Vouches():
		return wall
	case report.Facts == nil, report.StoreVersion < usersVersion, f.members == 0:
		return wall
	}
	switch f.standing {
	case standingNone, standingRemoved:
		return adminEmailNotMember
	case standingMember:
		return adminEmailNotAdmin
	}
	return wall
}
