package ops

// The thirteen values an [Error]'s Detail takes, each finer than its Code.
// A client treats a value it does not know as opaque, since the set
// grows additively within a version.
const (
	DetailDuplicateAddress   = "duplicate_address"
	DetailDuplicateUsername  = "duplicate_username"
	DetailInvalidAddress     = "invalid_address"
	DetailInvalidDisplayName = "invalid_display_name"
	DetailBlankDisplayName   = "blank_display_name"
	DetailInvalidUsername    = "invalid_username"
	DetailMissingUsername    = "missing_username"
	DetailLastAdmin          = "last_admin"
	DetailEnvironmentSets    = "environment_sets"
	DetailBusy               = "busy"
	DetailStoreURL           = "store_url"
	DetailExtra              = "extra"
	DetailManifest           = "manifest"
)
