import { fireEvent, waitFor } from '@testing-library/dom';
import { render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import Journey from './Journey.svelte';
import { LiveRegion } from '../announce';
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

	it("announces state changes through a caller-supplied announcer, never a region of its own", async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse({ promoted: false, pushed_back: true, detail: 'no confident match', files: [] }),
			),
		);
		const store = newStore();
		const target = { textContent: '' };
		const announcer = new LiveRegion();
		announcer.attach(target);

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Unknown Album', totalFiles: 3, store, announcer });

		await waitFor(() => expect(screen.getByText('No MusicBrainz match')).toBeInTheDocument());
		expect(target.textContent).toBe('No MusicBrainz match for this album.');
	});

	it('shows a retry control (never freezing on Matching forever) when finishSession itself fails (U13 fold F6.9)', async () => {
		const fetchMock = vi.fn((): Promise<Response> => Promise.reject(new TypeError('network drop')));
		vi.stubGlobal('fetch', fetchMock);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store });

		await waitFor(() => expect(screen.getByRole('button', { name: 'Try again' })).toBeInTheDocument());
		const matching = screen.getByText('Matching').closest('li');
		expect(matching).toHaveClass('step-error');

		fetchMock.mockResolvedValue(jsonResponse({ promoted: true, files: [] }));
		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
		await waitFor(() => expect(screen.queryByText('adding it now')).toBeInTheDocument());
	});

	it('stops journey polling once unmounted, even when finishSession resolves after unmount (U13 fold F6.10)', async () => {
		const deferred: { resolve: (res: Response) => void } = { resolve: () => {} };
		let journeyCalls = 0;
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return new Promise<Response>((resolve) => (deferred.resolve = resolve));
			}
			journeyCalls += 1;
			return Promise.resolve(jsonResponse({ session_id: 'a'.repeat(26), state: 'tagging', files: [] }));
		});
		vi.stubGlobal('fetch', fetchMock);
		const store = newStore();

		const { unmount } = render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store });
		unmount();
		deferred.resolve(jsonResponse({ promoted: true, files: [] }));
		await Promise.resolve();
		await Promise.resolve();

		expect(journeyCalls).toBe(0);
	});

	it('composes the section-4 rejection sentence and renders the per-file verdict list on a post-upload format rejection (U13 fold F6.2/F6.7)', async () => {
		const reason = {
			code: 'below-cd-quality',
			friendly: "This isn't CD quality, and the library keeps lossless FLAC.",
			next_step: 'Upload the CD-quality FLAC files instead.',
			label: 'below CD quality',
		};
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse({
					promoted: false,
					files: [
						{ Name: '01 Intro.flac', OK: true },
						{ Name: '02 Summer Song.mp3', OK: false, Reason: reason },
					],
				}),
			),
		);
		const store = newStore();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Some Album', totalFiles: 2, store });

		await waitFor(() => expect(screen.getByText("Couldn't add this album")).toBeInTheDocument());
		expect(screen.getByText('One track is below CD quality, and albums go in whole.')).toBeInTheDocument();
		expect(screen.getByText('01 Intro.flac')).toBeInTheDocument();
		expect(screen.getByText('fine')).toBeInTheDocument();
		expect(screen.getByText('below CD quality')).toBeInTheDocument();
		expect(screen.getByText(reason.friendly, { exact: false })).toBeInTheDocument();
	});

	it('offers a full-width "Try another album" retry on a post-upload format rejection, wired to the caller\'s own reset (U13 fold F8.2)', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse({
					promoted: false,
					files: [{ Name: '02 Summer Song.mp3', OK: false, Reason: { code: 'bad-extension', friendly: 'x', next_step: 'y', label: 'not FLAC' } }],
				}),
			),
		);
		const store = newStore();
		const onUploadAnother = vi.fn();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Some Album', totalFiles: 1, store, onUploadAnother });

		await waitFor(() => expect(screen.getByText("Couldn't add this album")).toBeInTheDocument());
		const retry = screen.getByRole('button', { name: 'Try another album' });
		expect(retry).toHaveClass('btn-primary', 'w-full');
		await fireEvent.click(retry);
		expect(onUploadAnother).toHaveBeenCalledOnce();
	});

	it('offers "Upload another album" on the in-library card, wired to the caller\'s own callback (U13 fold F6.11)', async () => {
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return Promise.resolve(jsonResponse({ promoted: true, files: [] }));
			}
			return Promise.resolve(
				jsonResponse({ session_id: 'a'.repeat(26), state: 'in-library', album: 'Loveless', files: [] }),
			);
		});
		vi.stubGlobal('fetch', fetchMock);
		const store = newStore();
		const onUploadAnother = vi.fn();

		render(Journey, { sessionId: 'a'.repeat(26), album: 'Loveless', totalFiles: 11, store, onUploadAnother });

		await waitFor(() => expect(screen.getByText('Loveless is ready.')).toBeInTheDocument());
		await fireEvent.click(screen.getByRole('button', { name: 'Upload another album' }));
		expect(onUploadAnother).toHaveBeenCalledOnce();
	});
});
