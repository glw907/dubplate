import { describe, expect, it } from 'vitest';
import { LiveRegion, progressAnnouncementText, shouldAnnounceProgress } from './announce';

describe('shouldAnnounceProgress', () => {
	it('announces the first call regardless of percent', () => {
		expect(shouldAnnounceProgress(null, 3)).toBe(true);
	});

	it('does not announce a delta below the threshold', () => {
		expect(shouldAnnounceProgress(10, 30)).toBe(false);
	});

	it('announces once the delta reaches the threshold', () => {
		expect(shouldAnnounceProgress(10, 35)).toBe(true);
	});

	it('always announces reaching 100%, even on a small delta', () => {
		expect(shouldAnnounceProgress(90, 100)).toBe(true);
	});
});

describe('progressAnnouncementText', () => {
	it('renders a rounded percent', () => {
		expect(progressAnnouncementText(42.6)).toBe('43 percent confirmed');
	});

	it('clamps out-of-range input', () => {
		expect(progressAnnouncementText(-5)).toBe('0 percent confirmed');
		expect(progressAnnouncementText(150)).toBe('100 percent confirmed');
	});
});

function fakeTarget(): { textContent: string } {
	return { textContent: '' };
}

describe('LiveRegion', () => {
	it('writes nothing until attached', () => {
		const region = new LiveRegion();
		region.announceState('hello');
		// No target attached: nothing to assert against but that this does not throw.
	});

	it('writes a progress announcement only once the delta clears the threshold', () => {
		const target = fakeTarget();
		const region = new LiveRegion();
		region.attach(target);

		region.announceProgress(10);
		expect(target.textContent).toBe('10 percent confirmed');

		region.announceProgress(20);
		expect(target.textContent).toBe('10 percent confirmed'); // below threshold: unchanged

		region.announceProgress(40);
		expect(target.textContent).toBe('40 percent confirmed');
	});

	it('resets the progress baseline on a state announcement', () => {
		const target = fakeTarget();
		const region = new LiveRegion();
		region.attach(target);

		region.announceProgress(50);
		region.announceState('Checking your upload.');
		expect(target.textContent).toBe('Checking your upload.');

		// After a state change, the very next progress call announces
		// again even though the delta from 50 is small.
		region.announceProgress(55);
		expect(target.textContent).toBe('55 percent confirmed');
	});

	it('always writes a completion announcement', () => {
		const target = fakeTarget();
		const region = new LiveRegion();
		region.attach(target);

		region.announceCompletion('Loveless is in your library.');
		expect(target.textContent).toBe('Loveless is in your library.');
	});

	it('stops writing once detached', () => {
		const target = fakeTarget();
		const region = new LiveRegion();
		region.attach(target);
		region.detach();

		region.announceState('should not land');
		expect(target.textContent).toBe('');
	});
});
