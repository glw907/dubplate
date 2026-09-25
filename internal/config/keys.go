package config

import "slices"

// WallRow is one row of the owned configuration wall: a name and,
// for a row whose requirement has not arrived yet, the status
// [Load] and the doctor subcommand report in its place.
type WallRow struct {
	Name string
	// Pending is empty for a row [Load] requires now. A non-empty
	// value is the status doctor reports instead, and Load neither
	// requires nor reads the row while it is set.
	Pending string
}

// ownedWall is the wall's required-now rows (spec, "The config wall,
// derived"). Rung 5's own owned-now row, DUBPLATE_IDENTITY_SOURCE, is
// registered from cmd/dubplate-server/auth.go rather than here, on the
// rung 1b registry contract: this package names no rung-specific key.
var ownedWall = []WallRow{
	{Name: "DUBPLATE_LIBRARY_ROOT"},
	{Name: "DUBPLATE_STATE_DIR"},
	{Name: "DUBPLATE_SITE_NAME"},
	{Name: "DUBPLATE_ADMIN_EMAIL"},
}

// OwnedWall returns the owned wall's rows in table order.
func OwnedWall() []WallRow {
	return slices.Clone(ownedWall)
}

// ownedNow reports whether name is an owned row in reg that [Load]
// requires now, as opposed to a row still carrying a Pending status.
//
// ruled, conductor 2026-09-07, R1
func ownedNow(reg *Registry, name string) bool {
	for _, row := range reg.ownedWall {
		if row.Name == name {
			return row.Pending == ""
		}
	}
	return false
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
			"DUBPLATE_PING_NAVIDROME_URL",
		},
	},
	{
		Name: "decision and review queues",
		Rung: 13,
		Keys: []string{
			"DUBPLATE_REVIEW_DIR",
		},
	},
	{
		Name: "streamer",
		Rung: 12,
		Keys: []string{
			"DUBPLATE_NAVIDROME_URL",
			"DUBPLATE_NAVIDROME_ADMIN_USER",
			"DUBPLATE_NAVIDROME_ADMIN_PASSWORD",
		},
	},
}

// ReservedClasses returns the reserved classes in table order.
func ReservedClasses() []ReservedClass {
	return slices.Clone(reservedClasses)
}

// reservedClassOf reports the class name belongs to in reg, if any.
//
// ruled, conductor 2026-09-07, R1
func reservedClassOf(reg *Registry, name string) (ReservedClass, bool) {
	for _, class := range reg.reservedClasses {
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
	"DUBPLATE_LIBRARY_DIR":     "renamed to DUBPLATE_LIBRARY_ROOT",
	"DUBPLATE_SERVE_SITE_NAME": "renamed to DUBPLATE_SITE_NAME",
	"DUBPLATE_ACCESS_TEAM_DOMAIN": "moved to the forward-auth front's own config; " +
		"the seam admits one, and its configuration arrives with the rung that builds it",
	"DUBPLATE_ACCESS_AUD": "moved to the forward-auth front's own config; " +
		"the seam admits one, and its configuration arrives with the rung that builds it",
	"DUBPLATE_PUBLIC_NAVIDROME_URL": "moved to the streamer driver's own config, " +
		"arriving at rung 12",
	"DUBPLATE_CONTRIBUTORS_FILE": "renamed to DUBPLATE_ROSTER_FILE, the magic-link " +
		"identity source's own roster location key (rung 5)",
	"DUBPLATE_REJECT_EMAIL_TEMPLATE": "retired: reject copy is a code contract in " +
		"internal/reject (rung 7), not a config-supplied template",
	"DUBPLATE_BEETS_CONFIG": "retired: the greenfield server never embeds beets",
	"DUBPLATE_BEETS_DB":     "retired: the greenfield server never embeds beets",
	"DUBPLATE_NAVIDROME_DB": "retired: the streamer driver reaches Navidrome " +
		"through its HTTP API only, never its database file",
	"DUBPLATE_TUNNEL_TOKEN": "retired: compose.yaml's cloudflared service reads " +
		"this directly, dubplate-server never does",
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

// supersessionOf reports the disposition message reg carries for a
// superseded name, if any.
//
// ruled, conductor 2026-09-07, R1
func supersessionOf(reg *Registry, name string) (string, bool) {
	msg, ok := reg.superseded[name]
	return msg, ok
}

// IntegrationClass is one named integration the wall accepts only when
// its Flag is set: an adopter-optional capability rather than an owned
// one (build spec U4). Setting any of its keys with the flag unset is a
// load error naming the flag, which is what keeps a half-configured
// integration from running unattended and observing nothing.
//
// Required and Optional differ in what Load does with them, not in what
// rejectNamespace accepts. Load requires every Required key when the
// flag is set. An Optional key is accepted and never required, because
// it is a registered threshold whose Default stands when it is unset
// (thresholds.go:60-79, parseThreshold, shared by ParseInt and
// ParseDuration).
type IntegrationClass struct {
	Name     string
	Rung     int
	Flag     string
	Required []string
	Optional []string
}

// integrationClassOf reports the integration class name belongs to in
// reg, by its Flag, a Required key, or an Optional key, if any. It
// walks reg's own declared classes rather than a package-level table,
// so a second integration class needs no second lookup site (rung 4's
// class registers itself as an example, from
// cmd/dubplate-server/shadow.go).
//
// ruled, conductor 2026-09-07, R5
func integrationClassOf(reg *Registry, name string) (IntegrationClass, bool) {
	for _, class := range reg.integrationClasses {
		if name == class.Flag || slices.Contains(class.Required, name) || slices.Contains(class.Optional, name) {
			return class, true
		}
	}
	return IntegrationClass{}, false
}

// Enabled reports whether c's flag resolves true in environ. An unset or
// empty flag is off. A value strconv.ParseBool refuses is reported as an
// error rather than silently read as off, because a typo that disables
// an unattended observer is the failure this class exists to prevent.
func (c IntegrationClass) Enabled(environ []string) (bool, error) {
	w := newWall(environ)
	return w.integrationEnabled(c)
}
