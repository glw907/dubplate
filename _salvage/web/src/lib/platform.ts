/**
 * Feature detection the selection UI (U10) needs to decide which
 * controls to show, plus the one iOS-Safari check the copy switch
 * needs (gate 3's rulings, docs/superpowers/plans/2026-08-31-dubplate-uploader.md).
 * Nothing here reads a file or its bytes -- both checks are static
 * environment probes.
 */

/**
 * Reports whether the running browser implements `webkitdirectory` on a
 * file input, feature-detected rather than user-agent-sniffed. Gate 3
 * confirms WebKit (bug 271705) has never implemented it, so this is
 * false on every iOS Safari and false everywhere `document` doesn't
 * exist (server-side rendering has no input to probe).
 */
export function supportsDirectorySelect(): boolean {
	if (typeof document === 'undefined') {
		return false;
	}
	const input = document.createElement('input');
	return 'webkitdirectory' in input;
}

/**
 * Reports whether the current browser is iOS Safari -- including
 * iPadOS's desktop-class user agent, which drops the "iPad" token and
 * is distinguished only by touch support on a Mac platform string. The
 * gate 3 rulings this file exists for (extension-first accept values,
 * the hidden multi-select gesture, no folder-select) apply to this
 * browser and no other.
 */
export function isIOSSafari(): boolean {
	if (typeof navigator === 'undefined') {
		return false;
	}
	const ua = navigator.userAgent;
	const isAppleTouch =
		/iP(hone|od|ad)/.test(ua) || (ua.includes('Macintosh') && navigator.maxTouchPoints > 1);
	const isWebKit = /WebKit/.test(ua) && !/CriOS|FxiOS|EdgiOS/.test(ua);
	return isAppleTouch && isWebKit;
}
