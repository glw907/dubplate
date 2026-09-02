import { fireEvent, waitFor } from '@testing-library/dom';
import { render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import App from './App.svelte';
import { uploadStore } from './lib/store/upload.svelte';

/**
 * App.svelte reads the module-singleton {@link uploadStore} directly
 * (unlike History/Journey, it has no `store` prop to inject a fake
 * one), so these tests drive it through the real singleton and reset
 * every field it owns before each test -- vitest gives this file its
 * own module registry, so nothing here leaks into another test file.
 */
function resetUploadStore(): void {
	uploadStore.config = null;
	uploadStore.signedOut = false;
	uploadStore.files = [];
	uploadStore.journey = null;
	uploadStore.journeyOutcome = null;
	uploadStore.history = [];
	uploadStore.resetSession();
}

function jsonResponse(body: unknown, status = 200): Response {
	return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': 'application/json' } });
}

const configWire = {
	site_name: 'Wright Family Music Server',
	collections: [''],
	chunk_mb: 8,
	max_album_bytes: 2 * 1024 * 1024 * 1024,
	max_sessions_per_user: 3,
	review_clock_days: 60,
	format_spec: 'FLAC only, 44100 Hz / 16-bit or better, one album per upload, up to 2.0 GiB total.',
	private_collection: 'Private Recordings',
};

/** internal/reject.ReasonTooManySessions's wire shape verbatim, per internal/reject/reject.go. */
const tooManySessionsReason = {
	code: 'too-many-sessions',
	friendly: 'You already have the maximum number of uploads in progress.',
	next_step: 'Wait for one of your other uploads to finish, then try again.',
	label: 'too many uploads',
};

beforeEach(() => {
	vi.unstubAllGlobals();
	localStorage.clear();
	resetUploadStore();
});

afterEach(() => {
	vi.unstubAllGlobals();
	localStorage.clear();
	resetUploadStore();
});

describe('App: config load failure', () => {
	it("shows a recoverable error, not a blank page, when /api/config fails (U13 fold F5.8)", async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 500 })));

		render(App);

		await waitFor(() => expect(screen.getByRole('heading', { name: "Couldn't load the uploader" })).toBeInTheDocument());
		expect(screen.getByRole('button', { name: 'Reload' })).toBeInTheDocument();
		// Never both: the loading window and the failed state read as the
		// same blank page unless they are mutually exclusive.
		expect(screen.queryByLabelText('Choose files', { selector: 'input' })).not.toBeInTheDocument();
	});
});

describe('App: session creation failure', () => {
	function stubFetch(): void {
		vi.stubGlobal(
			'fetch',
			vi.fn((input: RequestInfo | URL) => {
				const url = String(input);
				if (url.includes('/api/config')) {
					return Promise.resolve(jsonResponse(configWire));
				}
				if (url.includes('/api/session')) {
					return Promise.resolve(jsonResponse(tooManySessionsReason, 429));
				}
				return Promise.resolve(jsonResponse([]));
			}),
		);
	}

	it("renders the section-4 rejected-album copy, never a permanent 'waiting' row, when POST /api/session fails (U13 fold F5.6)", async () => {
		stubFetch();
		render(App);

		const input = await waitFor(
			() => screen.getByLabelText('Choose files', { selector: 'input' }) as HTMLInputElement,
		);
		const files = [new File([], 'track01.flac'), new File([], 'track02.flac')];
		Object.defineProperty(input, 'files', { value: files, configurable: true });
		await fireEvent.change(input);

		await waitFor(() => expect(screen.getByRole('heading', { name: "Couldn't add this album" })).toBeInTheDocument());
		expect(screen.getByText(new RegExp(tooManySessionsReason.friendly))).toBeInTheDocument();
		expect(screen.getByText(new RegExp(tooManySessionsReason.next_step))).toBeInTheDocument();
		expect(screen.getByRole('button', { name: 'Try another album' })).toBeInTheDocument();
		// The rows this session's files were added as never get a chance
		// to render at all in this branch, so "waiting" cannot appear.
		expect(screen.queryByText('waiting')).not.toBeInTheDocument();
	});
});

describe('App: History\'s "Upload to Private Recordings" from a stale session', () => {
	it('guards against an empty tag form rather than opening one with zero tracks (U13 fold F5.13)', async () => {
		const sessionA = 'a'.repeat(26);
		localStorage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Ghost Album', kind: 'pushed-back' }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn((input: RequestInfo | URL) => {
				const url = String(input);
				if (url.includes('/api/config')) {
					return Promise.resolve(jsonResponse(configWire));
				}
				return Promise.resolve(jsonResponse([]));
			}),
		);

		render(App);

		await waitFor(() => expect(screen.getByLabelText('Choose files', { selector: 'input' })).toBeInTheDocument());
		await fireEvent.click(screen.getByRole('button', { name: 'Your uploads' }));

		await waitFor(() => expect(screen.getByText('Ghost Album')).toBeInTheDocument());
		await fireEvent.click(screen.getByText('Ghost Album'));
		await fireEvent.click(screen.getByRole('button', { name: 'Upload to Private Recordings' }));

		// pendingFiles was never populated in this browser tab (a History
		// row is by definition a past session), so the tag form must never
		// mount with zero tracks -- the user lands back on the idle picker.
		expect(screen.queryByRole('heading', { name: 'Tag this album' })).not.toBeInTheDocument();
		await waitFor(() => expect(screen.getByLabelText('Choose files', { selector: 'input' })).toBeInTheDocument());
	});
});

describe('App: idle layout (U13 fold F5.1/F5.2)', () => {
	it('renders the content column title, the always-visible drop panel, and the centered hint beneath the primary picker', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(configWire)));

		render(App);

		await waitFor(() => expect(screen.getByLabelText('Choose files', { selector: 'input' })).toBeInTheDocument());
		expect(screen.getByRole('heading', { name: 'Wright Family Music Server', level: 1 })).toBeInTheDocument();
		expect(screen.getByText('add music to the library')).toBeInTheDocument();
		expect(screen.getByText('Drop an album anywhere on this page')).toBeInTheDocument();
		expect(screen.getByText('One album at a time. Albums go in whole.')).toBeInTheDocument();
	});
});
