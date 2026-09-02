import { render, screen } from '@testing-library/svelte';
import { afterEach, describe, expect, it } from 'vitest';
import WakeLockBar from './WakeLockBar.svelte';
import { WakeLock } from '../wakeLock.svelte';

/** `WakeLock.supported` reads `navigator.wakeLock` at construction time, so this must run before `new WakeLock()`. */
function stubWakeLockSupport(): void {
	(navigator as unknown as { wakeLock?: unknown }).wakeLock = {
		request: async () => ({ release: async () => {}, addEventListener: () => {} }),
	};
}

afterEach(() => {
	delete (navigator as unknown as { wakeLock?: unknown }).wakeLock;
});

describe('WakeLockBar', () => {
	it('shows the toggle filled when the lock is held (U13 fold F6.3)', () => {
		stubWakeLockSupport();
		const wakeLock = new WakeLock();
		wakeLock.on = true;

		render(WakeLockBar, { wakeLock });

		const button = screen.getByRole('button', { name: 'Keep my screen awake' });
		expect(button).toHaveAttribute('aria-pressed', 'true');
		expect(button).toHaveClass('btn-primary');
		expect(button).not.toHaveClass('btn-outline');
	});

	it('shows the toggle outlined when the lock is off', () => {
		stubWakeLockSupport();
		const wakeLock = new WakeLock();

		render(WakeLockBar, { wakeLock });

		const button = screen.getByRole('button', { name: 'Keep my screen awake' });
		expect(button).toHaveAttribute('aria-pressed', 'false');
		expect(button).toHaveClass('btn-outline');
		expect(button).not.toHaveClass('btn-primary');
	});

	it('hides the toggle entirely on a browser with no Wake Lock API', () => {
		const wakeLock = new WakeLock();
		render(WakeLockBar, { wakeLock });
		expect(screen.queryByRole('button', { name: 'Keep my screen awake' })).not.toBeInTheDocument();
	});
});
