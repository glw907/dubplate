import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import type { TusDriver, TusUploadConfig, TusUploadHandle, TusUploadHandlers } from '../tus';
import { UploadStore, sessionIdFromUploadURL, type KeyValueStorage } from './upload.svelte';

/** An in-memory {@link KeyValueStorage}, so resume-anchor tests never touch jsdom's real localStorage. */
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

/**
 * A {@link TusDriver} stub standing in for the "stubbed endpoint" the
 * plan asks the store's tests to run against: `start` never touches the
 * network, only records its call and hands the test the handlers so it
 * can play a server's callbacks back by hand.
 */
class FakeTusDriver implements TusDriver {
	calls: { file: File; config: TusUploadConfig; handlers: TusUploadHandlers }[] = [];
	aborts: string[] = [];

	start(file: File, config: TusUploadConfig, handlers: TusUploadHandlers): TusUploadHandle {
		this.calls.push({ file, config, handlers });
		return {
			abort: () => {
				this.aborts.push(file.name);
			},
		};
	}

	lastHandlers(): TusUploadHandlers {
		const call = this.calls.at(-1);
		if (!call) {
			throw new Error('FakeTusDriver: start() was never called');
		}
		return call.handlers;
	}
}

function testFile(name = 'track01.flac', size = 30 * 1024 * 1024): File {
	// A real Blob sized to `size` would just churn memory for a value the
	// test only ever reads back off `.size`; overriding the property
	// directly gets the same store-visible shape for free.
	const file = new File([], name, { type: 'audio/flac', lastModified: 0 });
	Object.defineProperty(file, 'size', { value: size });
	return file;
}

function jsonResponse(body: unknown, status = 200, contentType = 'application/json'): Response {
	return new Response(JSON.stringify(body), { status, headers: { 'Content-Type': contentType } });
}

const config = {
	site_name: 'dubplate',
	collections: [''],
	chunk_mb: 8,
	max_album_bytes: 2 * 1024 * 1024 * 1024,
	max_sessions_per_user: 3,
	review_clock_days: 60,
	format_spec: 'FLAC only, 44100 Hz / 16-bit or better, one album per upload, up to 2.0 GiB total.',
	private_collection: 'Private Recordings',
};

beforeEach(() => {
	vi.unstubAllGlobals();
});

describe('loadConfig', () => {
	it('fetches /api/config and converts every field from its wire name', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(config)));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.loadConfig();

		expect(store.config).toEqual({
			siteName: 'dubplate',
			collections: [''],
			chunkMB: 8,
			maxAlbumBytes: 2 * 1024 * 1024 * 1024,
			maxSessionsPerUser: 3,
			reviewClockDays: 60,
			formatSpec: config.format_spec,
			privateCollection: 'Private Recordings',
		});
	});

	it('sets signedOut on a 401 rather than throwing into the caller', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.loadConfig();

		expect(store.signedOut).toBe(true);
		expect(store.config).toBeNull();
	});

	it('sets signedOut on an HTML body (an Access reauth page riding a 200)', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(new Response('<html>sign in</html>', { status: 200, headers: { 'Content-Type': 'text/html' } })),
		);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.loadConfig();

		expect(store.signedOut).toBe(true);
	});
});

describe('addFile and startFile', () => {
	async function readyStore(): Promise<{ store: UploadStore; driver: FakeTusDriver }> {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(config)));
		const driver = new FakeTusDriver();
		const store = new UploadStore(driver, memoryStorage());
		await store.loadConfig();
		return { store, driver };
	}

	it('adds a file as pending without starting it', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile());

		expect(store.files).toHaveLength(1);
		expect(store.files[0]).toMatchObject({ id, status: 'pending', progressBytes: 0, ackedBytes: 0 });
		expect(driver.calls).toHaveLength(0);
	});

	it('starts an upload using the chunk size from /api/config, never a hardcoded default', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile());

		store.startFile(id, { session_id: 'sess-1' });

		expect(driver.calls).toHaveLength(1);
		expect(driver.calls[0].config.chunkSizeBytes).toBe(8 * 1024 * 1024);
		expect(driver.calls[0].config.metadata).toMatchObject({ filename: 'track01.flac', session_id: 'sess-1' });
		expect(store.files[0].status).toBe('uploading');
	});

	it('keeps progressBytes (transfer) and ackedBytes (server-confirmed) distinct, acked lagging transfer', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile('track01.flac', 1000));
		store.startFile(id, {});

		driver.lastHandlers().onProgress(400, 1000);
		driver.lastHandlers().onAcked(256);

		const entry = store.files[0];
		expect(entry.progressBytes).toBe(400);
		expect(entry.ackedBytes).toBe(256);
	});

	it('marks a file ok and sets both progress fields to its full size on success', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile('track01.flac', 1000));
		store.startFile(id, {});

		driver.lastHandlers().onSuccess('https://example/files/abc');

		const entry = store.files[0];
		expect(entry.status).toBe('ok');
		expect(entry.progressBytes).toBe(1000);
		expect(entry.ackedBytes).toBe(1000);
		expect(entry.tusUploadURL).toBe('https://example/files/abc');
	});

	it('renders the honest stalled state on retry-budget exhaustion, distinct from a server rejection', async () => {
		const { store, driver } = await readyStore();
		const stalledId = store.addFile(testFile('a.flac'));
		store.startFile(stalledId, {});
		driver.lastHandlers().onStalled();
		expect(store.files[0]).toMatchObject({ status: 'failed', failureKind: 'connection' });

		const rejectedId = store.addFile(testFile('b.flac'));
		store.startFile(rejectedId, {});
		driver.lastHandlers().onRejected(null);
		expect(store.files[1]).toMatchObject({ status: 'failed', failureKind: 'rejected', verdict: null });
	});

	it('stores the rejection wire payload as the file\'s verdict, for the row to render the real reason', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile('b.flac'));
		store.startFile(id, {});

		driver.lastHandlers().onRejected({
			code: 'no-space',
			friendly: "The library doesn't have enough free space to accept this upload right now.",
			next_step: 'Try again in a little while, or let the admin know if this keeps happening.',
			label: 'library is full',
		});

		expect(store.files[0].verdict).toEqual({
			Name: 'b.flac',
			OK: false,
			Reason: {
				code: 'no-space',
				friendly: "The library doesn't have enough free space to accept this upload right now.",
				next_step: 'Try again in a little while, or let the admin know if this keeps happening.',
				label: 'library is full',
			},
		});
	});

	it('sets the store-level signedOut flag on a 401/HTML failure without marking the file failed', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile());
		store.startFile(id, {});

		driver.lastHandlers().onSignedOut();

		expect(store.signedOut).toBe(true);
		// The 401 path costs no retry budget and is not a per-file verdict:
		// the file's own status is left alone, since the whole page (not
		// this one row) renders the reload-to-sign-in state.
		expect(store.files[0].status).toBe('uploading');
	});

	it('records a downshifted chunk size and uses it on the next startFile call', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile());
		store.startFile(id, {});

		driver.lastHandlers().onDownshift(4 * 1024 * 1024);
		expect(store.files[0].downshiftedChunkBytes).toBe(4 * 1024 * 1024);

		driver.lastHandlers().onStalled();
		store.retryFile(id, {});

		expect(driver.calls.at(-1)?.config.chunkSizeBytes).toBe(4 * 1024 * 1024);
	});

	it('retryFile refuses a file that is not in the failed state', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile());
		store.startFile(id, {});

		store.retryFile(id, {});

		expect(driver.calls).toHaveLength(1);
	});

	it('abortFile calls the handle without touching file state or the resume anchor', async () => {
		const { store, driver } = await readyStore();
		const id = store.addFile(testFile('track01.flac'));
		store.startFile(id, {});
		driver.lastHandlers().onUploadURLAvailable('https://example/files/partial');

		store.abortFile(id);

		expect(driver.aborts).toEqual(['track01.flac']);
		expect(store.resumableAnchors).toHaveLength(1);
	});
});

describe('resume anchoring', () => {
	it('persists only {fileId, name, size, tusUploadURL} -- never file bytes -- once the URL is known', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(config)));
		const driver = new FakeTusDriver();
		const storage = memoryStorage();
		const store = new UploadStore(driver, storage);
		await store.loadConfig();
		const id = store.addFile(testFile('track01.flac', 500));
		store.startFile(id, {});

		driver.lastHandlers().onUploadURLAvailable('https://example/files/xyz');

		expect(store.resumableAnchors).toEqual([{ fileId: id, name: 'track01.flac', size: 500, tusUploadURL: 'https://example/files/xyz' }]);
		const raw = storage.getItem('dubplate:upload:resume-anchors');
		expect(raw).not.toBeNull();
		expect(JSON.parse(raw as string)).toEqual(store.resumableAnchors);
	});

	it('forgets the anchor once the upload succeeds', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(config)));
		const driver = new FakeTusDriver();
		const store = new UploadStore(driver, memoryStorage());
		await store.loadConfig();
		const id = store.addFile(testFile());
		store.startFile(id, {});
		driver.lastHandlers().onUploadURLAvailable('https://example/files/xyz');

		driver.lastHandlers().onSuccess('https://example/files/xyz');

		expect(store.resumableAnchors).toHaveLength(0);
	});

	it('hydrates a persisted anchor as a needsReselect row on construction, with no File attached', () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f-1', name: 'track01.flac', size: 500, tusUploadURL: 'https://example/files/xyz' }]),
		);

		const store = new UploadStore(new FakeTusDriver(), storage);

		expect(store.files).toEqual([
			{
				id: 'f-1',
				name: 'track01.flac',
				size: 500,
				status: 'pending',
				progressBytes: 0,
				ackedBytes: 0,
				downshiftedChunkBytes: null,
				failureKind: null,
				verdict: null,
				tusUploadURL: 'https://example/files/xyz',
				needsReselect: true,
			},
		]);
	});

	it('startFile refuses a needsReselect row until reselectFile clears the flag', () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f-1', name: 'track01.flac', size: 500, tusUploadURL: 'https://example/files/xyz' }]),
		);
		const driver = new FakeTusDriver();
		const store = new UploadStore(driver, storage);

		store.startFile('f-1', {});

		expect(driver.calls).toHaveLength(0);
	});

	it('reselectFile resumes the same row on a name+size match, without duplicating it', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f-1', name: 'track01.flac', size: 500, tusUploadURL: 'https://example/files/xyz' }]),
		);
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse(config)));
		const driver = new FakeTusDriver();
		const store = new UploadStore(driver, storage);
		await store.loadConfig();
		const anchor = store.resumableAnchors[0];

		const id = store.reselectFile(anchor, testFile('track01.flac', 500));

		expect(id).toBe('f-1');
		expect(store.files).toHaveLength(1);
		expect(store.files[0].needsReselect).toBe(false);

		store.startFile('f-1', {});
		expect(driver.calls).toHaveLength(1);
		expect(driver.calls[0].config.resumeURL).toBe('https://example/files/xyz');
	});

	it('reselectFile refuses a file whose name or size does not match the anchor', () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f-1', name: 'track01.flac', size: 500, tusUploadURL: 'https://example/files/xyz' }]),
		);
		const store = new UploadStore(new FakeTusDriver(), storage);
		const anchor = store.resumableAnchors[0];

		expect(store.reselectFile(anchor, testFile('track01.flac', 999))).toBeNull();
		expect(store.reselectFile(anchor, testFile('wrong-name.flac', 500))).toBeNull();
		expect(store.files[0].needsReselect).toBe(true);
	});
});

describe('createSession', () => {
	it('POSTs album/collection/expected_files and adopts the returned session id', async () => {
		const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ session_id: 'sess-123' }, 201));
		vi.stubGlobal('fetch', fetchMock);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		const id = await store.createSession('Album Title', 'family', 3);

		expect(id).toBe('sess-123');
		expect(store.session).toMatchObject({ sessionId: 'sess-123', album: 'Album Title', collection: 'family', state: 'uploading' });
		const [, init] = fetchMock.mock.calls[0] as [string, RequestInit];
		expect(JSON.parse(init.body as string)).toEqual({ album: 'Album Title', collection: 'family', expected_files: 3 });
	});

	it('sets signedOut and rethrows on a 401', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await expect(store.createSession('Album', '', 1)).rejects.toThrow();
		expect(store.signedOut).toBe(true);
	});
});

describe('resetSession', () => {
	it('clears session/journey/journeyOutcome and stops any running journey poll (U13 fold F5.7)', async () => {
		vi.useFakeTimers();
		const fetchMock = vi.fn().mockResolvedValue(jsonResponse({ session_id: sessionA, state: 'uploaded', files: [] }));
		vi.stubGlobal('fetch', fetchMock);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());
		store.session = { sessionId: sessionA, album: 'Album A', collection: 'family', state: 'uploading', detail: '' };
		store.journeyOutcome = {
			promoted: true,
			pushedBack: false,
			probeUnavailable: false,
			detail: '',
			matchedArtist: '',
			matchedAlbum: '',
			matchedYear: '',
			files: [],
		};
		store.startJourneyPolling(sessionA, 1000);
		await vi.advanceTimersByTimeAsync(0);
		expect(store.journey).not.toBeNull();

		store.resetSession();

		expect(store.session).toEqual({ sessionId: null, album: '', collection: '', state: 'uploading', detail: '' });
		expect(store.journey).toBeNull();
		expect(store.journeyOutcome).toBeNull();

		// A second album's poll must never land against the first
		// album's interval: if resetSession left the timer running, this
		// would still tick and repopulate store.journey.
		fetchMock.mockClear();
		await vi.advanceTimersByTimeAsync(5000);
		expect(fetchMock).not.toHaveBeenCalled();
		expect(store.journey).toBeNull();
		vi.useRealTimers();
	});
});

const sessionA = 'a2b2c2d2e2f2g2h2i2j2k2l2m2'; // 26 lowercase base32 characters
const sessionB = 'n2o2p2q2r2s2t2u2v2w2x2y2z2'; // a second, distinct 26-character id

describe('sessionIdFromUploadURL', () => {
	it('recovers the 26-character session id prefix from a completed upload URL', () => {
		expect(sessionIdFromUploadURL(`/upload/files/${sessionA}-0`)).toBe(sessionA);
		expect(sessionIdFromUploadURL(`https://music.example.com/upload/files/${sessionA}-2`)).toBe(sessionA);
	});

	it('returns null for a URL whose last segment is not a session-id-shaped prefix', () => {
		expect(sessionIdFromUploadURL('/upload/files/not-a-session-id')).toBeNull();
		expect(sessionIdFromUploadURL('')).toBeNull();
	});
});

describe('finishSession', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('POSTs to the session finish endpoint and converts every wire field', async () => {
		const fetchMock = vi.fn().mockResolvedValue(
			jsonResponse({
				promoted: true,
				matched_artist: 'My Bloody Valentine',
				matched_album: 'Loveless',
				matched_year: '1991',
				files: [],
			}),
		);
		vi.stubGlobal('fetch', fetchMock);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		const outcome = await store.finishSession(sessionA, 'Loveless');

		expect(outcome).toEqual({
			promoted: true,
			pushedBack: false,
			probeUnavailable: false,
			detail: '',
			matchedArtist: 'My Bloody Valentine',
			matchedAlbum: 'Loveless',
			matchedYear: '1991',
			files: [],
		});
		expect(store.journeyOutcome).toEqual(outcome);
		const [url, init] = fetchMock.mock.calls[0] as [string, RequestInit];
		expect(url).toBe(`/upload/api/session/${sessionA}/finish`);
		expect(init.method).toBe('POST');
	});

	it('remembers a pushed-back outcome in storage for History to recover after a reload', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse({ promoted: false, pushed_back: true, detail: 'no match', files: [] })),
		);
		const storage = memoryStorage();
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.finishSession(sessionA, 'Unknown Album');

		const raw = storage.getItem('dubplate:upload:pushed-back');
		expect(JSON.parse(raw ?? '[]')).toEqual([{ sessionId: sessionA, album: 'Unknown Album', kind: 'pushed-back' }]);
	});

	it('remembers a probe-unavailable outcome distinctly from a pushed-back one', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse({ promoted: false, probe_unavailable: true, files: [] })),
		);
		const storage = memoryStorage();
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.finishSession(sessionA, 'Some Album');

		const raw = storage.getItem('dubplate:upload:pushed-back');
		expect(JSON.parse(raw ?? '[]')).toEqual([{ sessionId: sessionA, album: 'Some Album', kind: 'probe-unavailable' }]);
	});

	it('does not remember anything for a straightforwardly promoted session', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse({ promoted: true, files: [] })));
		const storage = memoryStorage();
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.finishSession(sessionA, 'Loveless');

		expect(storage.getItem('dubplate:upload:pushed-back')).toBeNull();
	});

	it('sets signedOut and rethrows on a 401', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await expect(store.finishSession(sessionA, 'Album')).rejects.toThrow();
		expect(store.signedOut).toBe(true);
	});
});

describe('journey polling', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
		vi.useRealTimers();
	});

	it('fetches once immediately and converts the wire shape', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse({ session_id: sessionA, state: 'tagging', album: 'Loveless', files: [] }),
			),
		);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.pollJourneyOnce(sessionA);

		expect(store.journey).toEqual({
			sessionId: sessionA,
			state: 'tagging',
			detail: '',
			album: 'Loveless',
			libraryLink: '',
			files: [],
		});
	});

	it('polls again after the interval elapses, and stops once told to', async () => {
		vi.useFakeTimers();
		const fetchMock = vi
			.fn()
			.mockResolvedValue(jsonResponse({ session_id: sessionA, state: 'uploaded', files: [] }));
		vi.stubGlobal('fetch', fetchMock);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		const stop = store.startJourneyPolling(sessionA, 1000);
		await vi.advanceTimersByTimeAsync(0);
		expect(fetchMock).toHaveBeenCalledTimes(1);

		await vi.advanceTimersByTimeAsync(1000);
		expect(fetchMock).toHaveBeenCalledTimes(2);

		stop();
		await vi.advanceTimersByTimeAsync(3000);
		expect(fetchMock).toHaveBeenCalledTimes(2);
	});

	it('sets signedOut on a poll that reports 401, without throwing', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.pollJourneyOnce(sessionA);

		expect(store.signedOut).toBe(true);
	});
});

describe('loadHistory', () => {
	afterEach(() => {
		vi.unstubAllGlobals();
	});

	it('fetches /api/history and converts every row', async () => {
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse([{ session_id: sessionA, state: 'in-library', album: 'Loveless' }])),
		);
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await store.loadHistory();

		expect(store.history).toEqual([
			{
				sessionId: sessionA,
				album: 'Loveless',
				state: 'in-library',
				detail: '',
				libraryLink: '',
				files: [],
				resumable: false,
			},
		]);
	});

	it('marks a row resumable when a resume anchor derives the same session id', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f1', name: 'track01.flac', size: 10, tusUploadURL: `/upload/files/${sessionA}-0` }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse([{ session_id: sessionA, state: 'uploading' }])),
		);
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.loadHistory();

		expect(store.history[0].resumable).toBe(true);
	});

	it('sorts resumable rows before non-resumable ones', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:resume-anchors',
			JSON.stringify([{ fileId: 'f1', name: 'track01.flac', size: 10, tusUploadURL: `/upload/files/${sessionB}-0` }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(
				jsonResponse([
					{ session_id: sessionA, state: 'in-library', album: 'Not Resumable' },
					{ session_id: sessionB, state: 'uploading', album: 'Resumable' },
				]),
			),
		);
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.loadHistory();

		expect(store.history.map((r) => r.sessionId)).toEqual([sessionB, sessionA]);
	});

	it('folds in a locally-remembered pushed-back session absent from the ledger', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Mystery Album', kind: 'pushed-back' }]),
		);
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(jsonResponse([])));
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.loadHistory();

		expect(store.history).toEqual([
			{ sessionId: sessionA, album: 'Mystery Album', state: 'pushed-back', detail: '', libraryLink: '', files: [], resumable: false },
		]);
	});

	it('drops a locally-remembered pushed-back record once the ledger reports a newer event for the same session', async () => {
		const storage = memoryStorage();
		storage.setItem(
			'dubplate:upload:pushed-back',
			JSON.stringify([{ sessionId: sessionA, album: 'Now Promoted', kind: 'pushed-back' }]),
		);
		vi.stubGlobal(
			'fetch',
			vi.fn().mockResolvedValue(jsonResponse([{ session_id: sessionA, state: 'in-library', album: 'Now Promoted' }])),
		);
		const store = new UploadStore(new FakeTusDriver(), storage);

		await store.loadHistory();

		expect(store.history).toHaveLength(1);
		expect(store.history[0].state).toBe('in-library');
	});

	it('sets signedOut without throwing on a 401', async () => {
		vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('', { status: 401 })));
		const store = new UploadStore(new FakeTusDriver(), memoryStorage());

		await expect(store.loadHistory()).resolves.toBeUndefined();
		expect(store.signedOut).toBe(true);
	});
});
