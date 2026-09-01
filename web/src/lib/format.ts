/**
 * Byte-count formatting shared by every U10 component that renders a
 * server- or client-observed size (SpecCard's `{max_album_size}` token,
 * FileRow's progress readout) -- one implementation so the two never
 * drift apart in rounding or unit choice.
 */

const gibibyte = 1024 * 1024 * 1024;
const mebibyte = 1024 * 1024;

/**
 * Renders bytes as a one-decimal MiB or GiB figure ("512.0 MB" below
 * 1 GiB, "2.0 GB" at or above it), matching `internal/serve`'s own
 * `humanGiB` convention for the values that originate server-side.
 */
export function formatBytes(bytes: number): string {
	if (bytes >= gibibyte) {
		return `${(bytes / gibibyte).toFixed(1)} GB`;
	}
	return `${(bytes / mebibyte).toFixed(1)} MB`;
}
