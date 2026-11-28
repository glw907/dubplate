package ops

// A Detail is an enumerated name finer than an [Error]'s Code, which a
// client may branch on where Reason is free text. A client treats a
// value it does not know as opaque, since the set grows additively
// within a version.
type Detail string

// The values an [Error]'s Detail takes.
const (
	DetailDuplicateAddress   Detail = "duplicate_address"
	DetailDuplicateUsername  Detail = "duplicate_username"
	DetailInvalidAddress     Detail = "invalid_address"
	DetailInvalidDisplayName Detail = "invalid_display_name"
	DetailBlankDisplayName   Detail = "blank_display_name"
	DetailInvalidUsername    Detail = "invalid_username"
	DetailMissingUsername    Detail = "missing_username"
	DetailLastAdmin          Detail = "last_admin"
	DetailEnvironmentSets    Detail = "environment_sets"
	DetailBusy               Detail = "busy"
	DetailStoreURL           Detail = "store_url"
	DetailExtra              Detail = "extra"
	DetailManifest           Detail = "manifest"
)
