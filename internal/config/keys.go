package config

import (
	"fmt"
	"slices"
)

// Sensitivity is the class every registered key carries: whether a
// change to it needs care. The zero value is no class, so a declaration
// that forgets the field panics at registration rather than reading as
// not sensitive, on the precedent of ops.Privilege.
//
// On a store-homed [Setting] the sensitivity gates the live write. On a
// key only the environment sets, which changes only at a restart, it marks
// a path, an address, a credential, or a trust boundary. It never says
// whether a value is a secret.
type Sensitivity int

const (
	// NotSensitive marks a key whose change needs no particular care.
	NotSensitive Sensitivity = iota + 1
	// Sensitive marks a key whose change needs care.
	Sensitive
)

// String returns m's name as a report prints it, or "Sensitivity(N)"
// for a sensitivity no one named, the zero value included.
func (m Sensitivity) String() string {
	switch m {
	case NotSensitive:
		return "not sensitive"
	case Sensitive:
		return "sensitive"
	default:
		return fmt.Sprintf("Sensitivity(%d)", int(m))
	}
}

// KeyDecl declares one environment key an [IntegrationClass] or an
// [IdentitySource] names, with the description and sensitivity every
// registered key carries.
type KeyDecl struct {
	Key         string
	Description string
	Sensitivity Sensitivity
}

// KeyNames returns the names of keys in order.
func KeyNames(keys []KeyDecl) []string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.Key
	}
	return names
}

// WallRow is one row of the owned configuration wall: a key [Load]
// requires.
type WallRow struct {
	Key         string
	Description string
	Sensitivity Sensitivity
}

// StateDirKey is the owned wall row naming the state directory.
const StateDirKey = "DUBPLATE_STATE_DIR"

// ownedWall is the wall's required-now rows (spec, "The config wall,
// derived"). The identity selector's owned-now row,
// DUBPLATE_IDENTITY_SOURCE, is registered by internal/auth rather than
// here: this package names no key that belongs to another package.
var ownedWall = []WallRow{
	{
		Key:         "DUBPLATE_LIBRARY_ROOT",
		Description: "the directory tree that holds the music library",
		Sensitivity: Sensitive,
	},
	{
		Key:         StateDirKey,
		Description: "the directory that holds the server's persistent state: the store, the auth records, and the snapshots",
		Sensitivity: Sensitive,
	},
	{
		Key:         "DUBPLATE_SITE_NAME",
		Description: "the name the server shows members and puts in the mail it sends",
		Sensitivity: NotSensitive,
	},
	{
		Key: "DUBPLATE_ADMIN_EMAIL",
		Description: "the operator's address: install seeds the first admin from it with member add, " +
			"it grants no authority at runtime, and it is the contact the music metadata lookups send",
		Sensitivity: Sensitive,
	},
}

// ownedNow reports whether name is an owned row in r, one [Load]
// requires.
func (r *Registry) ownedNow(name string) bool {
	return slices.ContainsFunc(r.ownedWall, func(row WallRow) bool { return row.Key == name })
}

// ReservedClass is one named class of configuration the wall
// recognizes but does not yet accept: setting any of its Keys is a
// load error naming the class, distinct from an unrecognized key.
type ReservedClass struct {
	Name string
	Keys []string
}

// reservedClasses groups every DUBPLATE_* name that belongs to a
// capability not yet built. A name absent from every class here and
// from [supersededKeys] classifies as unknown.
var reservedClasses = []ReservedClass{
	{
		// Admission, tagging, layout, and collection routing own the
		// directories a session moves through once intake has accepted
		// it.
		Name: "pipeline",
		Keys: []string{
			"DUBPLATE_INBOX_DIR",
			"DUBPLATE_QUARANTINE_DIR",
		},
	},
	{
		Name: "monitoring",
		Keys: []string{
			"DUBPLATE_PING_IMPORT_URL",
			"DUBPLATE_PING_BACKUP_URL",
			"DUBPLATE_PING_DISK_URL",
		},
	},
	{
		Name: "decision and review queues",
		Keys: []string{
			"DUBPLATE_REVIEW_DIR",
		},
	},
}

// reservedClassOf reports the class name belongs to in r, if any.
func (r *Registry) reservedClassOf(name string) (ReservedClass, bool) {
	for _, class := range r.reservedClasses {
		if slices.Contains(class.Keys, name) {
			return class, true
		}
	}
	return ReservedClass{}, false
}

// supersededKeys maps a pre-pivot DUBPLATE_* name with no meaning in
// the greenfield wall to the message [wall.rejectNamespace] reports:
// the renamed key that replaces it, or, for a capability the
// greenfield server drops outright, that it retired with no
// successor.
var supersededKeys = map[string]string{
	"DUBPLATE_LIBRARY_DIR":          "renamed to DUBPLATE_LIBRARY_ROOT",
	"DUBPLATE_SERVE_SITE_NAME":      "renamed to DUBPLATE_SITE_NAME",
	"DUBPLATE_ACCESS_TEAM_DOMAIN":   "retired: dubplate delegates sign-in to no external service",
	"DUBPLATE_ACCESS_AUD":           "retired: dubplate delegates sign-in to no external service",
	"DUBPLATE_PUBLIC_NAVIDROME_URL": "retired: dubplate integrates with no external streamer",
	"DUBPLATE_CONTRIBUTORS_FILE": "retired: members live in the store; add the first " +
		"admin with dubplate-server member add ADDRESS --admin",
	"DUBPLATE_REJECT_EMAIL_TEMPLATE": "retired: reject copy is a code contract in " +
		"internal/reject, not a config-supplied template",
	"DUBPLATE_BEETS_CONFIG": "retired: the greenfield server never embeds beets",
	"DUBPLATE_BEETS_DB":     "retired: the greenfield server never embeds beets",
	"DUBPLATE_NAVIDROME_DB": "retired: dubplate integrates with no external streamer",
	"DUBPLATE_TUNNEL_TOKEN": "retired: only the incumbent stack's cloudflared " +
		"read this, and that stack is gone; dubplate-server never does",
	"DUBPLATE_R2_ACCOUNT_ID": "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote " +
		"in rclone's RCLONE_CONFIG_<NAME>_* environment form",
	"DUBPLATE_R2_BUCKET": "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote " +
		"in rclone's RCLONE_CONFIG_<NAME>_* environment form",
	"DUBPLATE_R2_REMOTE": "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote " +
		"in rclone's RCLONE_CONFIG_<NAME>_* environment form",
	"DUBPLATE_R2_ACCESS_KEY_ID": "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote " +
		"in rclone's RCLONE_CONFIG_<NAME>_* environment form",
	"DUBPLATE_R2_SECRET_ACCESS_KEY": "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote " +
		"in rclone's RCLONE_CONFIG_<NAME>_* environment form",
	"DUBPLATE_BACKUP_STAGING": "superseded: backup staging now lives in the store leaf " +
		"under DUBPLATE_STATE_DIR",
}

// supersessionOf reports the disposition message r carries for a
// superseded name, if any.
func (r *Registry) supersessionOf(name string) (string, bool) {
	msg, ok := r.superseded[name]
	return msg, ok
}

// IntegrationClass is one named integration the wall accepts only when
// its Flag is set: an adopter-optional capability rather than an owned
// one. Setting any of its keys with the flag unset is a
// load error naming the flag, so a half-configured integration is
// refused rather than silently ignored.
//
// Load requires every Required key when the flag is set.
type IntegrationClass struct {
	Name     string
	Flag     KeyDecl
	Required []KeyDecl
}

// integrationClassOf reports the integration class name belongs to in
// r, by its Flag or a Required key, if any. It walks r's own
// declared classes rather than a package-level table, so a second
// integration class needs no second lookup site.
func (r *Registry) integrationClassOf(name string) (IntegrationClass, bool) {
	for _, class := range r.integrationClasses {
		if name == class.Flag.Key || slices.ContainsFunc(class.Required, func(k KeyDecl) bool { return k.Key == name }) {
			return class, true
		}
	}
	return IntegrationClass{}, false
}

// Enabled reports whether c's flag resolves true in environ, with the
// semantics of integrationEnabled.
func (c IntegrationClass) Enabled(environ []string) (bool, error) {
	w := newWall(environ)
	return w.integrationEnabled(c)
}
