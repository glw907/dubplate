// Package config is the portability wall: the only place dubplate reads
// the process environment. A misconfigured adopter install fails at
// [Load], never silently borrowing this estate's own values.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ErrMissing marks a required setting that Load found unset. Load wraps
// it once per missing field and joins every wrap with [errors.Join], so a
// caller can print the complete list rather than stopping at the first
// miss.
var ErrMissing = errors.New("required setting unset")

// Config is dubplate's full, validated configuration. Load is the only
// constructor; every field it does not default is required, so a value
// present here is a value [Load] confirmed was set.
type Config struct {
	AdminEmail          string
	SMTP                SMTPConfig
	R2                  R2Config
	Navidrome           NavidromeConfig
	Contributors        string
	RejectEmailTemplate string
	BeetsConfig         string
	BeetsDB             string
	NavidromeDB         string
	Paths               Paths
	Thresholds          Thresholds
	// PingURLs maps a lowercased check slug to its ping URL, populated by
	// scanning the environment for DUBPLATE_PING_<SLUG>_URL. An empty map
	// is valid: no checks configured, no pings sent.
	PingURLs map[string]string
	// Collections lists the beets flexible-attribute collection names, in
	// query-prefixed path order; Collections[0] is the default collection,
	// which keeps today's unprefixed layout.
	Collections []string
}

// SMTPConfig holds the outbound mail settings notify needs for implicit
// TLS delivery.
type SMTPConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	From     string
}

// R2Config holds the credentials and location backup and restore need to
// reach the Cloudflare R2 bucket via rclone.
type R2Config struct {
	AccountID       string
	Bucket          string
	Remote          string
	AccessKeyID     string
	SecretAccessKey string
}

// NavidromeConfig holds the URL and admin credentials the scan trigger
// and doctor probes use against Navidrome's native API.
type NavidromeConfig struct {
	URL           string
	AdminUser     string
	AdminPassword string
}

// Paths holds every directory dubplate reads or writes on the host.
type Paths struct {
	LibraryDir    string
	InboxDir      string
	StagingDir    string
	StateDir      string
	RunDir        string
	ReviewDir     string
	QuarantineDir string
	BackupStaging string
}

// Thresholds holds the behavior tunables that are the only fields Load
// defaults; every other field is required with no default.
type Thresholds struct {
	QuiescentMin  int
	FreeFloorGB   int
	ArchiveCapGB  int
	RetentionDays int
	DiskPct       int
}

// Values carried forward unchanged from the retired shell layer's
// music-import and musicbox-disk-check scripts, whose env vars were named
// MUSICBOX_QUIESCENT_MINUTES=10, MUSICBOX_MIN_FREE_GB=5,
// MUSICBOX_MAX_ARCHIVE_BYTES=2147483648 (2 GiB, per docs/HISTORY.md),
// MUSICBOX_REVIEW_RETENTION_DAYS=60, MUSICBOX_DISK_THRESHOLD_PCT=90.
const (
	defaultQuiescentMin  = 10
	defaultFreeFloorGB   = 5
	defaultArchiveCapGB  = 2
	defaultRetentionDays = 60
	defaultDiskPct       = 90
)

// get reads a single environment variable. Load and its helpers read
// every setting through this one choke point, so the packaging
// initiative can later overlay a non-secret config file by changing get
// alone, with no call site touched (the spec's composable-source seam;
// only env is wired now).
func get(name string) (string, bool) {
	return os.LookupEnv(name)
}

// Load reads Config from the process environment. It returns every
// required field that is unset, wrapped in [ErrMissing] and joined via
// [errors.Join], rather than stopping at the first miss, so a caller such
// as the doctor subcommand can report the complete list in one pass.
func Load() (*Config, error) {
	var errs []error

	require := func(name string) string {
		v, ok := get(name)
		if !ok || v == "" {
			errs = append(errs, fmt.Errorf("%s: %w", name, ErrMissing))
		}
		return v
	}

	requirePort := func(name string) int {
		v := require(name)
		if v == "" {
			return 0
		}
		port, err := strconv.Atoi(v)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: invalid port %q", name, v))
			return 0
		}
		return port
	}

	cfg := &Config{
		AdminEmail: require("DUBPLATE_ADMIN_EMAIL"),
		SMTP: SMTPConfig{
			Host:     require("DUBPLATE_SMTP_HOST"),
			Port:     requirePort("DUBPLATE_SMTP_PORT"),
			User:     require("DUBPLATE_SMTP_USER"),
			Password: require("DUBPLATE_SMTP_PASSWORD"),
			From:     require("DUBPLATE_SMTP_FROM"),
		},
		R2: R2Config{
			AccountID:       require("DUBPLATE_R2_ACCOUNT_ID"),
			Bucket:          require("DUBPLATE_R2_BUCKET"),
			Remote:          require("DUBPLATE_R2_REMOTE"),
			AccessKeyID:     require("DUBPLATE_R2_ACCESS_KEY_ID"),
			SecretAccessKey: require("DUBPLATE_R2_SECRET_ACCESS_KEY"),
		},
		Navidrome: NavidromeConfig{
			URL:           require("DUBPLATE_NAVIDROME_URL"),
			AdminUser:     require("DUBPLATE_NAVIDROME_ADMIN_USER"),
			AdminPassword: require("DUBPLATE_NAVIDROME_ADMIN_PASSWORD"),
		},
		Contributors:        require("DUBPLATE_CONTRIBUTORS_FILE"),
		RejectEmailTemplate: require("DUBPLATE_REJECT_EMAIL_TEMPLATE"),
		BeetsConfig:         require("DUBPLATE_BEETS_CONFIG"),
		BeetsDB:             require("DUBPLATE_BEETS_DB"),
		NavidromeDB:         require("DUBPLATE_NAVIDROME_DB"),
		Paths: Paths{
			LibraryDir:    require("DUBPLATE_LIBRARY_DIR"),
			InboxDir:      require("DUBPLATE_INBOX_DIR"),
			StagingDir:    require("DUBPLATE_STAGING_DIR"),
			StateDir:      require("DUBPLATE_STATE_DIR"),
			RunDir:        require("DUBPLATE_RUN_DIR"),
			ReviewDir:     require("DUBPLATE_REVIEW_DIR"),
			QuarantineDir: require("DUBPLATE_QUARANTINE_DIR"),
			BackupStaging: require("DUBPLATE_BACKUP_STAGING"),
		},
		Thresholds:  loadThresholds(&errs),
		PingURLs:    pingURLsFromEnv(),
		Collections: loadCollections(),
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return cfg, nil
}

// loadThresholds reads the behavior tunables, defaulting any that are
// unset and recording a parse error on any that are set but not an
// integer.
func loadThresholds(errs *[]error) Thresholds {
	optInt := func(name string, def int) int {
		v, ok := get(name)
		if !ok || v == "" {
			return def
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			*errs = append(*errs, fmt.Errorf("%s: invalid integer %q", name, v))
			return def
		}
		return n
	}

	return Thresholds{
		QuiescentMin:  optInt("DUBPLATE_QUIESCENT_MINUTES", defaultQuiescentMin),
		FreeFloorGB:   optInt("DUBPLATE_FREE_FLOOR_GB", defaultFreeFloorGB),
		ArchiveCapGB:  optInt("DUBPLATE_ARCHIVE_CAP_GB", defaultArchiveCapGB),
		RetentionDays: optInt("DUBPLATE_RETENTION_DAYS", defaultRetentionDays),
		DiskPct:       optInt("DUBPLATE_DISK_PCT", defaultDiskPct),
	}
}

// loadCollections reads the comma-separated DUBPLATE_COLLECTIONS list,
// defaulting to a single unnamed collection (today's layout) when unset
// or when every entry is blank.
func loadCollections() []string {
	v, ok := get("DUBPLATE_COLLECTIONS")
	if !ok || v == "" {
		return []string{""}
	}

	var cols []string
	for c := range strings.SplitSeq(v, ",") {
		if c = strings.TrimSpace(c); c != "" {
			cols = append(cols, c)
		}
	}
	if len(cols) == 0 {
		return []string{""}
	}
	return cols
}

// pingURLsFromEnv scans the process environment for
// DUBPLATE_PING_<SLUG>_URL variables. Ping URLs are inherently
// environment-shaped: an adopter adds a new check by setting a new
// variable, not by touching config code, so this scans [os.Environ]
// directly rather than going through get.
func pingURLsFromEnv() map[string]string {
	const prefix, suffix = "DUBPLATE_PING_", "_URL"

	urls := make(map[string]string)
	for _, kv := range os.Environ() {
		name, value, ok := strings.Cut(kv, "=")
		if !ok || len(name) <= len(prefix)+len(suffix) ||
			!strings.HasPrefix(name, prefix) || !strings.HasSuffix(name, suffix) {
			continue
		}
		slug := strings.ToLower(name[len(prefix) : len(name)-len(suffix)])
		if slug == "" {
			continue
		}
		urls[slug] = value
	}
	return urls
}
