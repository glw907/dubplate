import { waitFor } from '@testing-library/dom';
import { render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Journey from './Journey.svelte';
import { UploadStore, type KeyValueStorage } from '../store/upload.svelte';

function memoryStorage(): KeyValueStorage {
	const data = new Map<string, string>();
	return {
		getItem: (key) => data.get(key) ?? null,
		setItem: (key, value) => {
			data.set(key, value);
		},
		removeItem: (key) => {
			data.delete(key);
		},
	};
}

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

/** Never touches the network; {@link UploadStore}'s constructor requires one even though this file never uploads a file. */
const noopDriver = { start: () => ({ abort: () => {} }) };

function newStore(): UploadStore {
	return new UploadStore(noopDriver, memoryStorage());
}

beforeEach(() => {
	vi.unstubAllGlobals();
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('Journey', () => {
	it('shows step 3 current while finishSession is still resolving', async () => {
		const deferred: { resolve: (res: Response) => void } = { resolve: () => {} };
		vi.stubGlobal(
			'fetch',
			vi.fn(() => new Promise<Response>((resolve) => (deferred.resolve = resolve))),
		);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store });

		expect(screen.getByText('Matching')).toBeInTheDocument();
		const matching = screen.getByText('Matching').closest('li');
		expect(matching).toHaveAttribute('aria-current', 'step');

		// Resolve the pending finish call so the test does not leak a
		// dangling fetch/interval into the next test.
		deferred.resolve(jsonResponse({ promoted: true, files: [] }));
		await waitFor(() => expect(screen.queryByText('adding it now')).toBeInTheDocument());
	});

	it('freezes the spine and shows the no-match alert on a pushed-back outcome', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse({
					promoted: false,
					pushed_back: true,
					detail: 'no confident match',
					files: [],
				}),
			),
		);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Unknown Album', totalFiles: 3, store });

		await waitFor(() => expect(screen.getByText('No MusicBrainz match')).toBeInTheDocument());
		const matching = screen.getByText('Matching').closest('li');
		expect(matching).toHaveClass('step-error');
	});

	it('shows the matched line and starts journey polling once promoted with matched metadata', async () => {
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return Promise.resolve(
					jsonResponse({
						promoted: true,
						matched_artist: 'My Bloody Valentine',
						matched_album: 'Loveless',
						matched_year: '1991',
						files: [],
					}),
				);
			}
			return Promise.resolve(jsonResponse({ session_id: 'a'.repeat(26), state: 'tagging', files: [] }));
		});
		vi.stubGlobal('fetch', fetchMock);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store });

		await waitFor(() =>
			expect(
				screen.getByText('Matched: My Bloody Valentine - Loveless (1991). You can close this page now.', {
					selector: 'p',
				}),
			).toBeInTheDocument(),
		);
		await waitFor(() => expect(store.journey?.state).toBe('tagging'));
	});

	it('live-transitions to the in-library card once the poll reports in-library', async () => {
		let journeyCalls = 0;
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return Promise.resolve(jsonResponse({ promoted: true, files: [] }));
			}
			journeyCalls += 1;
			return Promise.resolve(
				jsonResponse({
					session_id: 'a'.repeat(26),
					state: 'in-library',
					album: 'Loveless',
					library_link: 'https://music.example.com/app/#/album?filter=x',
					files: [],
				}),
			);
		});
		vi.stubGlobal('fetch', fetchMock);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store });

		await waitFor(() => expect(screen.getByText('Loveless is ready.')).toBeInTheDocument());
		expect(screen.getByRole('heading', { name: 'In your library' })).toBeInTheDocument();
		expect(screen.getByRole('link', { name: 'Open in Navidrome' })).toHaveAttribute(
			'href',
			'https://music.example.com/app/#/album?filter=x',
		);
		expect(journeyCalls).toBeGreaterThan(0);
	});
});
