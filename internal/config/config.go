// Package config is the portability wall: the only place dubplate reads
// the process environment. A misconfigured adopter install fails at
// [Load], never silently borrowing this estate's own values.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
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
	Serve               ServeConfig
	Upload              UploadConfig
	Contributors        string
	RejectEmailTemplate string
	BeetsConfig         string
	BeetsDB             string
	NavidromeDB         string
	Paths               Paths
	// LedgerPath is the contributions ledger's JSONL file, defaulted to
	// <StateDir>/contributions.jsonl when unset.
	LedgerPath string
	Thresholds Thresholds
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

// ServeConfig holds the settings `dubplate serve` needs to bind its
// listener and to verify Cloudflare Access assertions. PublicNavidromeURL
// is the externally reachable address the journey UI links into, distinct
// from [NavidromeConfig.URL]'s admin-API address.
type ServeConfig struct {
	BindAddrs          []string
	Port               int
	AccessTeamDomain   string
	AccessAUD          string
	PublicNavidromeURL string
	// SiteName is the uploader page's masthead text, joining /api/config
	// so the adopter goal holds: no site-identifying string is baked
	// into the web build.
	SiteName string
}

// UploadConfig holds the upload-path thresholds: chunk size, the
// per-album byte ceiling, how long an abandoned partial upload survives
// before a sweep reclaims it, and how many sessions one contributor may
// have in flight at once. Every field is defaulted.
type UploadConfig struct {
	ChunkMB            int
	MaxAlbumBytes      int64
	PartialTTL         time.Duration
	MaxSessionsPerUser int
}

// Paths holds every directory dubplate reads or writes on the host.
type Paths struct {
	LibraryDir string
	InboxDir   string
	StagingDir string
	StateDir   string
	// TriggerDir is the nudge trigger queue directory
	// (config/tmpfiles-dubplate.conf creates it at boot via
	// systemd-tmpfiles), backed by tmpfs under /run and so never a home
	// for durable state (that is StateDir's job). serve writes a
	// trigger file here per promoted session; nudge-consume reads and
	// deletes them.
	TriggerDir    string
	ReviewDir     string
	QuarantineDir string
	BackupStaging string
	// UploadsDir is the tus resumable-upload store; it must share a
	// filesystem with InboxDir since promotion into the inbox is an
	// atomic rename.
	UploadsDir string
}

// Thresholds holds the behavior tunables that are the only fields Load
// defaults; every other field is required with no default.
type Thresholds struct {
	QuiescentMin  int
	FreeFloorGB   int
	ArchiveCapGB  int
	RetentionDays int
	DiskPct       int
	// StallMinutes is how long a session may sit with no terminal ledger
	// event before it is reported stalled; see defaultStallMinutes for
	// the floor this default must clear.
	StallMinutes int
	// ReportRetention is how many run-report files survive before the
	// oldest are pruned.
	ReportRetention int
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

// importTimerIntervalMin mirrors systemd/dubplate-import.timer's
// OnCalendar=*:0/15: the worst-case wait for the backstop importer to
// even notice a finished, healthy upload.
const importTimerIntervalMin = 15

// defaultStallMinutes must exceed importTimerIntervalMin (the worst-case
// wait for the backstop timer to fire) plus defaultQuiescentMin (the
// importer's own wait before touching a session once triggered): 15+10 =
// 25. Below that floor, a perfectly healthy upload caught between timer
// ticks would be reported stalled before the backstop path ever gets to
// run it. 30 clears the floor with a five-minute margin.
const defaultStallMinutes = importTimerIntervalMin + defaultQuiescentMin + 5

const defaultReportRetention = 30

const (
	defaultUploadChunkMB = 8
	// defaultMaxAlbumBytes is 2 GiB, the same magnitude as
	// defaultArchiveCapGB: generous headroom for a multi-disc lossless
	// album while still bounding a single session's tus-store footprint.
	defaultMaxAlbumBytes      int64 = 2 * 1024 * 1024 * 1024
	defaultPartialTTL               = 24 * time.Hour
	defaultMaxSessionsPerUser       = 3
)

// get reads a single environment variable. Load and its helpers read
// every setting through this one choke point, so the packaging
// initiative can later overlay a non-secret config file by changing get
// alone, with no call site touched (the spec's composable-source seam;
// only env is wired now).
func get(name string) (string, bool) {
	return os.LookupEnv(name)
}

// optInt reads name as an integer, returning def when name is unset or
// empty and recording a parse error on errs when name is set but not a
// valid integer.
func optInt(name string, def int, errs *[]error) int {
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

// optInt64 is [optInt] for values too large for int on a 32-bit build,
// such as a byte ceiling.
func optInt64(name string, def int64, errs *[]error) int64 {
	v, ok := get(name)
	if !ok || v == "" {
		return def
	}
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: invalid integer %q", name, v))
		return def
	}
	return n
}

// optDuration reads name via [time.ParseDuration], returning def when
// name is unset or empty and recording a parse error on errs when name is
// set but not a valid duration.
func optDuration(name string, def time.Duration, errs *[]error) time.Duration {
	v, ok := get(name)
	if !ok || v == "" {
		return def
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: invalid duration %q", name, v))
		return def
	}
	return d
}

// optString reads name, returning def when name is unset or empty.
func optString(name, def string) string {
	if v, ok := get(name); ok && v != "" {
		return v
	}
	return def
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

	// requireList reads name as a comma-separated list, trimming and
	// dropping blank entries. A missing name, an empty value, or a value
	// with no non-blank entries all record one [ErrMissing].
	requireList := func(name string) []string {
		v, ok := get(name)
		if !ok || v == "" {
			errs = append(errs, fmt.Errorf("%s: %w", name, ErrMissing))
			return nil
		}
		var list []string
		for item := range strings.SplitSeq(v, ",") {
			if item = strings.TrimSpace(item); item != "" {
				list = append(list, item)
			}
		}
		if len(list) == 0 {
			errs = append(errs, fmt.Errorf("%s: %w", name, ErrMissing))
		}
		return list
	}

	paths := Paths{
		LibraryDir:    require("DUBPLATE_LIBRARY_DIR"),
		InboxDir:      require("DUBPLATE_INBOX_DIR"),
		StagingDir:    require("DUBPLATE_STAGING_DIR"),
		StateDir:      require("DUBPLATE_STATE_DIR"),
		TriggerDir:    require("DUBPLATE_TRIGGER_DIR"),
		ReviewDir:     require("DUBPLATE_REVIEW_DIR"),
		QuarantineDir: require("DUBPLATE_QUARANTINE_DIR"),
		BackupStaging: require("DUBPLATE_BACKUP_STAGING"),
		UploadsDir:    require("DUBPLATE_UPLOADS_DIR"),
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
		Serve: ServeConfig{
			BindAddrs:          requireList("DUBPLATE_SERVE_BIND_ADDRS"),
			Port:               requirePort("DUBPLATE_SERVE_PORT"),
			AccessTeamDomain:   require("DUBPLATE_ACCESS_TEAM_DOMAIN"),
			AccessAUD:          require("DUBPLATE_ACCESS_AUD"),
			PublicNavidromeURL: require("DUBPLATE_PUBLIC_NAVIDROME_URL"),
			SiteName:           require("DUBPLATE_SERVE_SITE_NAME"),
		},
		Upload:              loadUpload(&errs),
		Contributors:        require("DUBPLATE_CONTRIBUTORS_FILE"),
		RejectEmailTemplate: require("DUBPLATE_REJECT_EMAIL_TEMPLATE"),
		BeetsConfig:         require("DUBPLATE_BEETS_CONFIG"),
		BeetsDB:             require("DUBPLATE_BEETS_DB"),
		NavidromeDB:         require("DUBPLATE_NAVIDROME_DB"),
		Paths:               paths,
		LedgerPath:          optString("DUBPLATE_LEDGER_PATH", filepath.Join(paths.StateDir, "contributions.jsonl")),
		Thresholds:          loadThresholds(&errs),
		PingURLs:            pingURLsFromEnv(),
		Collections:         loadCollections(),
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
	return Thresholds{
		QuiescentMin:    optInt("DUBPLATE_QUIESCENT_MINUTES", defaultQuiescentMin, errs),
		FreeFloorGB:     optInt("DUBPLATE_FREE_FLOOR_GB", defaultFreeFloorGB, errs),
		ArchiveCapGB:    optInt("DUBPLATE_ARCHIVE_CAP_GB", defaultArchiveCapGB, errs),
		RetentionDays:   optInt("DUBPLATE_RETENTION_DAYS", defaultRetentionDays, errs),
		DiskPct:         optInt("DUBPLATE_DISK_PCT", defaultDiskPct, errs),
		StallMinutes:    optInt("DUBPLATE_STALL_MINUTES", defaultStallMinutes, errs),
		ReportRetention: optInt("DUBPLATE_REPORT_RETENTION", defaultReportRetention, errs),
	}
}

// loadUpload reads the upload-path thresholds, defaulting any that are
// unset and recording a parse error on any that are set but invalid.
func loadUpload(errs *[]error) UploadConfig {
	return UploadConfig{
		ChunkMB:            optInt("DUBPLATE_UPLOAD_CHUNK_MB", defaultUploadChunkMB, errs),
		MaxAlbumBytes:      optInt64("DUBPLATE_UPLOAD_MAX_ALBUM_BYTES", defaultMaxAlbumBytes, errs),
		PartialTTL:         optDuration("DUBPLATE_UPLOAD_PARTIAL_TTL", defaultPartialTTL, errs),
		MaxSessionsPerUser: optInt("DUBPLATE_UPLOAD_MAX_SESSIONS_PER_USER", defaultMaxSessionsPerUser, errs),
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
