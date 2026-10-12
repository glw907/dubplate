package config

import (
	"fmt"
	"slices"
)

// Sensitivity is the mark every registered key carries: whether a
// change to it needs care. The zero value is no mark, so a declaration
// that forgets the field panics at registration rather than reading as
// not sensitive, on the precedent of ops.Privilege.
//
// On a store-homed [Setting] the mark gates the live write. On a key
// only the environment sets, which changes only at a restart, it marks
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
// for a mark no one named, the zero value included.
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
// [IdentitySource] names, with the description and mark every
// registered key carries.
type KeyDecl struct {
	Name        string
	Description string
	Mark        Sensitivity
}

// KeyNames returns the names of keys in order.
func KeyNames(keys []KeyDecl) []string {
	names := make([]string, len(keys))
	for i, k := range keys {
		names[i] = k.Name
	}
	return names
}

// WallRow is one row of the owned configuration wall: a key [Load]
// requires.
type WallRow struct {
	Name        string
	Description string
	Mark        Sensitivity
}

// ownedWall is the wall's required-now rows (spec, "The config wall,
// derived"). Rung 5's own owned-now row, DUBPLATE_IDENTITY_SOURCE, is
// registered from cmd/dubplate-server/auth.go rather than here, on the
// rung 1b registry contract: this package names no rung-specific key.
var ownedWall = []WallRow{
	{
		Name:        "DUBPLATE_LIBRARY_ROOT",
		Description: "the directory tree that holds the music library",
		Mark:        Sensitive,
	},
	{
		Name:        "DUBPLATE_STATE_DIR",
		Description: "the directory that holds the server's persistent state: the store, the auth records, and the snapshots",
		Mark:        Sensitive,
	},
	{
		Name:        "DUBPLATE_SITE_NAME",
		Description: "the name the server shows members and puts in the mail it sends",
		Mark:        NotSensitive,
	},
	{
		Name: "DUBPLATE_ADMIN_EMAIL",
		Description: "the operator's address: install seeds the first admin from it with member add, " +
			"it grants no authority at runtime, and it is the contact the music metadata lookups send",
		Mark: Sensitive,
	},
}

// OwnedWall returns the owned wall's rows in table order.
func OwnedWall() []WallRow {
	return slices.Clone(ownedWall)
}

// ownedNow reports whether name is an owned row in r, one [Load]
// requires.
//
// ruled, conductor 2026-09-07, R1
func (r *Registry) ownedNow(name string) bool {
	return slices.ContainsFunc(r.ownedWall, func(row WallRow) bool { return row.Name == name })
}

// ReservedClass is one named class of configuration the wall
// recognizes but does not yet accept: setting any of its Keys is a
// load error naming the class and Rung, distinct from an unrecognized
// key.
type ReservedClass struct {
	Name string
	Rung int
	Keys []string
}

// reservedClasses groups every DUBPLATE_* name that belongs to a
// capability not yet built, keyed to the rung that claims it. A name
// absent from every class here and from [supersededKeys] classifies as
// unknown.
var reservedClasses = []ReservedClass{
	{
		Name: "intake",
		Rung: 7,
		Keys: []string{
			"DUBPLATE_UPLOADS_DIR",
			"DUBPLATE_TRIGGER_DIR",
		},
	},
	{
		// Admission, tagging, layout, and collection routing (ladder
		// rung 9) own the directories a session moves through once
		// intake has accepted it.
		Name: "pipeline",
		Rung: 9,
		Keys: []string{
			"DUBPLATE_INBOX_DIR",
			"DUBPLATE_STAGING_DIR",
			"DUBPLATE_QUARANTINE_DIR",
		},
	},
	{
		Name: "monitoring",
		Rung: 9,
		Keys: []string{
			"DUBPLATE_PING_IMPORT_URL",
			"DUBPLATE_PING_BACKUP_URL",
			"DUBPLATE_PING_DISK_URL",
		},
	},
	{
		Name: "decision and review queues",
		Rung: 13,
		Keys: []string{
			"DUBPLATE_REVIEW_DIR",
		},
	},
}

// ReservedClasses returns the reserved classes in table order.
func ReservedClasses() []ReservedClass {
	return slices.Clone(reservedClasses)
}

// reservedClassOf reports the class name belongs to in r, if any.
//
// ruled, conductor 2026-09-07, R1
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
		"internal/reject (rung 7), not a config-supplied template",
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
//
// ruled, conductor 2026-09-07, R1
func (r *Registry) supersessionOf(name string) (string, bool) {
	msg, ok := r.superseded[name]
	return msg, ok
}

// IntegrationClass is one named integration the wall accepts only when
// its Flag is set: an adopter-optional capability rather than an owned
// one (build spec U4). Setting any of its keys with the flag unset is a
// load error naming the flag, so a half-configured integration is
// refused rather than silently ignored.
//
// Load requires every Required key when the flag is set.
type IntegrationClass struct {
	Name     string
	Rung     int
	Flag     KeyDecl
	Required []KeyDecl
}

// integrationClassOf reports the integration class name belongs to in
// r, by its Flag or a Required key, if any. It walks r's own
// declared classes rather than a package-level table, so a second
// integration class needs no second lookup site.
//
// ruled, conductor 2026-09-07, R5
func (r *Registry) integrationClassOf(name string) (IntegrationClass, bool) {
	for _, class := range r.integrationClasses {
		if name == class.Flag.Name || slices.ContainsFunc(class.Required, func(k KeyDecl) bool { return k.Name == name }) {
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
