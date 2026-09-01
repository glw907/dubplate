/**
 * The Screen Wake Lock toggle's state, held here rather than inside the
 * component that draws it (`WakeLockBar.svelte`): App.svelte builds one
 * {@link WakeLock} and hands it to both places the toggle appears, its own
 * in-flight branch and `Journey.svelte`, so a lock requested mid-transfer
 * survives the hand-off instead of dropping and re-arming. It is the same
 * shape the shared aria-live region already uses.
 */

/**
 * A structural fallback for the Screen Wake Lock API: some
 * TypeScript lib.dom.d.ts versions ship it, some do not, and this file
 * must build cleanly either way (U9's "no Node shims" production-build
 * criterion extends to never depending on a DOM lib version this
 * workstation happens to have).
 */
interface WakeLockSentinelLike {
	release(): Promise<void>;
	addEventListener(type: 'release', listener: () => void): void;
}
interface NavigatorWithWakeLock {
	wakeLock?: { request(type: 'screen'): Promise<WakeLockSentinelLike> };
}

/** One Screen Wake Lock toggle's state and its request/release logic. */
export class WakeLock {
	/** Whether a lock is currently held, which the browser can also drop on its own. */
	on = $state(false);
	/** A browser capability, fixed at construction, so not reactive state. */
	readonly supported = typeof navigator !== 'undefined' && 'wakeLock' in navigator;

	private sentinel: WakeLockSentinelLike | null = null;

	/** Requests the lock when off, releases it when on. */
	async toggle(): Promise<void> {
		if (this.on) {
			await this.release();
			return;
		}
		const nav = navigator as Navigator & NavigatorWithWakeLock;
		if (!nav.wakeLock) {
			return;
		}
		try {
			this.sentinel = await nav.wakeLock.request('screen');
			this.on = true;
			this.sentinel.addEventListener('release', () => {
				this.on = false;
			});
		} catch {
			this.on = false;
		}
	}

	/** Releases the lock if held; a no-op otherwise, safe to call from `onDestroy` unconditionally. */
	async release(): Promise<void> {
		if (this.sentinel) {
			await this.sentinel.release();
			this.sentinel = null;
		}
		this.on = false;
	}
}
