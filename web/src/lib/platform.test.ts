import { describe, expect, it, vi } from 'vitest';
import { isIOSSafari, supportsDirectorySelect } from './platform';

describe('supportsDirectorySelect', () => {
	// jsdom implements neither Chromium's nor WebKit's `webkitdirectory`
	// extension, so the real environment already exercises the "absent"
	// branch every other test in this file runs under; only the "present"
	// branch needs a mocked input to prove the feature-detection itself
	// is correct, not just its default answer here.
	it('reports false when the created input has no webkitdirectory property (jsdom, real WebKit/iOS)', () => {
		expect(supportsDirectorySelect()).toBe(false);
	});

	it('reports true when the created input carries webkitdirectory', () => {
		const original = document.createElement.bind(document);
		vi.spyOn(document, 'createElement').mockImplementation((tag: string) => {
			const el = original(tag);
			if (tag === 'input') {
				Object.defineProperty(el, 'webkitdirectory', { value: false, configurable: true });
			}
			return el;
		});
		expect(supportsDirectorySelect()).toBe(true);
		vi.restoreAllMocks();
	});
});

describe('isIOSSafari', () => {
	const iphoneSafariUA =
		'Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Mobile/15E148 Safari/604.1';
	const ipadDesktopUA =
		'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_6) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3 Safari/605.1.15';
	const androidChromeUA =
		'Mozilla/5.0 (Linux; Android 14) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/128.0 Mobile Safari/537.36';
	const iosChromeUA =
		'Mozilla/5.0 (iPhone; CPU iPhone OS 18_3_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) CriOS/128.0 Mobile/15E148 Safari/604.1';

	function withNavigator<T>(userAgent: string, maxTouchPoints: number, fn: () => T): T {
		const original = { userAgent: navigator.userAgent, maxTouchPoints: navigator.maxTouchPoints };
		Object.defineProperty(navigator, 'userAgent', { value: userAgent, configurable: true });
		Object.defineProperty(navigator, 'maxTouchPoints', { value: maxTouchPoints, configurable: true });
		try {
			return fn();
		} finally {
			Object.defineProperty(navigator, 'userAgent', { value: original.userAgent, configurable: true });
			Object.defineProperty(navigator, 'maxTouchPoints', { value: original.maxTouchPoints, configurable: true });
		}
	}

	it('is true for iPhone Safari', () => {
		expect(withNavigator(iphoneSafariUA, 5, () => isIOSSafari())).toBe(true);
	});

	it('is true for iPadOS Safari, whose UA claims Macintosh (touch-point disambiguation)', () => {
		expect(withNavigator(ipadDesktopUA, 5, () => isIOSSafari())).toBe(true);
	});

	it('is false for a real Mac with no touch points', () => {
		expect(withNavigator(ipadDesktopUA, 0, () => isIOSSafari())).toBe(false);
	});

	it('is false for Android Chrome', () => {
		expect(withNavigator(androidChromeUA, 5, () => isIOSSafari())).toBe(false);
	});

	it('is false for Chrome on iOS (still WebKit under the hood, but not Safari)', () => {
		expect(withNavigator(iosChromeUA, 5, () => isIOSSafari())).toBe(false);
	});
});
