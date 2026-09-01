/**
 * The at-selection pre-flight (plan's three-layer format prevention,
 * layer 3): the instant files are chosen, before a single byte
 * uploads, a mixed batch is stopped cold rather than discovered at 80%
 * upload. This is deliberately narrow -- extension is the only thing
 * checkable without reading file bytes (the corrupt/below-CD-quality/
 * unsafe-path reasons all need a real decode, which is
 * `internal/gate`'s job once bytes are actually on the wire) -- so it
 * mirrors exactly one `internal/reject.Reason` (`ReasonBadExtension`),
 * verbatim from `internal/reject/reject.go`, rather than growing into
 * the client-side taxonomy table the plan bans. The server remains the
 * enforcement of record; this is a courtesy screen only.
 */
import type { RejectReason } from './store/upload.svelte';

/**
 * Verbatim mirror of `internal/reject.ReasonBadExtension`'s wire
 * values (`reject.go`'s `Friendly`/`NextStep`/`Label`/`code`) -- the
 * one reason this module is authorized to duplicate, since selection
 * time has no server round trip to source it from instead.
 */
export const badExtensionReason: RejectReason = {
	code: 'bad-extension',
	friendly: "This isn't a FLAC file, and the library keeps lossless FLAC.",
	next_step: 'Convert it to FLAC and upload it again.',
	label: 'not FLAC',
};

/**
 * Extensions the gate treats as an audio track that isn't FLAC --
 * flagged at pre-flight. Any other extension (cover art, a booklet
 * PDF, a `.cue` sheet) is an "extra" the gate never inspects
 * (passb-copy-final.md section 1's "The checks only look at audio
 * files. Booklets and other extras ride along.") and passes through
 * unflagged here too.
 */
const lossyAudioExtensions = new Set(['.mp3', '.m4a', '.aac', '.wav', '.ogg', '.wma', '.alac', '.ape']);

function extensionOf(filename: string): string {
	const dot = filename.lastIndexOf('.');
	return dot === -1 ? '' : filename.slice(dot).toLowerCase();
}

/** One screened file's pre-flight outcome. */
export interface PreflightVerdict {
	file: File;
	/** Null when the file passes the extension screen. */
	reason: RejectReason | null;
}

/**
 * Screens a ZIP selection: the archive's contents aren't inspectable
 * client-side (no full-file reads, per the research constraints), so a
 * lone `.zip` always passes pre-flight -- section 2's "we check the
 * tracks once the zip finishes uploading" is the honest statement that
 * this path's real check is server-side only.
 */
export function isZipSelection(files: File[]): boolean {
	return files.length === 1 && extensionOf(files[0].name) === '.zip';
}

/**
 * Screens a multi-file selection's extensions. A ZIP selection should
 * go through {@link isZipSelection} instead -- calling this on one
 * would flag the lone `.zip` itself as a bad extension.
 */
export function screenSelection(files: File[]): PreflightVerdict[] {
	return files.map((file) => ({
		file,
		reason: lossyAudioExtensions.has(extensionOf(file.name)) ? badExtensionReason : null,
	}));
}
