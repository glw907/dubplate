import { fireEvent, waitFor } from '@testing-library/dom';
import { render, screen } from '@testing-library/svelte';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import History from './History.svelte';
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

function jsonResponse(body: unknown): Response {
	return new Response(JSON.stringify(body), { status: 200, headers: { 'Content-Type': 'application/json' } });
}

const noopDriver = { start: () => ({ abort: () => {} }) };

beforeEach(() => {
	vi.unstubAllGlobals();
});

afterEach(() => {
	vi.unstubAllGlobals();
});

const sessionA = 'a'.repeat(26);
const sessionB = 'b'.repeat(26);

describe('History', () => {
	it('shows the empty state when there is nothing yet', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse([])));
		render(History, { store: new UploadStore(noopDriver, memoryStorage()) });

		await waitFor(() => expect(screen.getByText('Nothing here yet.')).toBeInTheDocument());
	});

	it('renders a status chip per the copy contract for each landed state', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse([
					{ session_id: sessionA, state: 'in-library', album: 'Loveless' },
					{ session_id: sessionB, state: 'queued', album: 'Second Album' },
				]),
			),
		);
		render(History, { store: new UploadStore(noopDriver, memoryStorage()) });

		await waitFor(() => expect(screen.getByText('in library')).toBeInTheDocument());
		expect(screen.getByText('waiting its turn')).toBeInTheDocument();
	});

	it('sorts a resumable in-flight row before the rest', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f1', name: 'track01.flac', size: 10, tusUploadURL: `/upload/files/${sessionB}-0` }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse([
					{ session_id: sessionA, state: 'in-library', album: 'Finished Album' },
					{ session_id: sessionB, state: 'uploading', album: 'In Flight Album' },
				]),
			),
		);
		render(History, { store: new UploadStore(noopDriver, storage) });

		await waitFor(() => expect(screen.getByText('uploading · resume')).toBeInTheDocument());
		const rows = screen.getAllByRole('listitem');
		expect(rows[0]).toHaveTextContent('In Flight Album');
		expect(rows[0]).toHaveTextContent('Resume');
	});

	it('opens the no-match copy and its actions for a locally-remembered pushed-back session', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Mystery Album', kind: 'pushed-back' }]),
		);
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse([])));
		const onUploadPrivate = vi.fn();
		render(History, { store: new UploadStore(noopDriver, storage), onUploadPrivate });

		await waitFor(() => expect(screen.getByText('pushed back')).toBeInTheDocument());
		expect(screen.queryByText('No MusicBrainz match')).not.toBeInTheDocument();

		await fireEvent.click(screen.getByText('Mystery Album'));
		expect(screen.getByText('No MusicBrainz match')).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Upload to Private Recordings' }));
		expect(onUploadPrivate).toHaveBeenCalledOnce();
	});

	it("re-probes the session (not onFixNames) when a probe-unavailable row's Try again is clicked", async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Slow Service Album', kind: 'probe-unavailable' }]),
		);
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return Promise.resolve(jsonResponse({ promoted: true, files: [] }));
			}
			return Promise.resolve(jsonResponse([]));
		});
		vi.stubGlobal('fetch', fetchMock);
		const onFixNames = vi.fn();
		render(History, { store: new UploadStore(noopDriver, storage), onFixNames });

		await waitFor(() => expect(screen.getByText('pushed back')).toBeInTheDocument());
		await fireEvent.click(screen.getByText('Slow Service Album'));
		expect(screen.getByText("Can't check MusicBrainz right now")).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

		await waitFor(() =>
			expect(fetchMock).toHaveBeenCalledWith(expect.stringContaining(`/session/${sessionA}/finish`), expect.anything()),
		);
		expect(onFixNames).not.toHaveBeenCalled();
	});

	it('shows a recoverable error, never the lying "Nothing here yet.", when loadHistory fails (U13 fold F5.14)', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 500 })));
		render(History, { store: new UploadStore(noopDriver, memoryStorage()) });

		await waitFor(() => expect(screen.getByRole('heading', { name: "Couldn't load your uploads" })).toBeInTheDocument());
		expect(screen.getByRole('button', { name: 'Reload' })).toBeInTheDocument();
		expect(screen.queryByText('Nothing here yet.')).not.toBeInTheDocument();
	});

	it("shows a visible message, not a silent no-op, when a probe-unavailable row's Try again fails (U13 fold F5.14)", async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Slow Service Album', kind: 'probe-unavailable' }]),
		);
		const fetchMock = vi.fn((input: RequestInfo | URL) => {
			const url = String(input);
			if (url.includes('/finish')) {
				return Promise.resolve(new Response('', { status: 500 }));
			}
			return Promise.resolve(jsonResponse([]));
		});
		vi.stubGlobal('fetch', fetchMock);
		render(History, { store: new UploadStore(noopDriver, storage) });

		await waitFor(() => expect(screen.getByText('pushed back')).toBeInTheDocument());
		await fireEvent.click(screen.getByText('Slow Service Album'));
		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));

		await waitFor(() =>
			expect(screen.getByText("That didn't go through. Check your connection and try again.")).toBeInTheDocument(),
		);
	});

	it('drops a locally-remembered pushed-back record once the ledger reports a newer event for the same session', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Now In Library', kind: 'pushed-back' }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse([{ session_id: sessionA, state: 'in-library', album: 'Now In Library' }])),
		);
		render(History, { store: new UploadStore(noopDriver, storage) });

		await waitFor(() => expect(screen.getByText('in library')).toBeInTheDocument());
		expect(screen.queryByText('pushed back')).not.toBeInTheDocument();
	});
});
