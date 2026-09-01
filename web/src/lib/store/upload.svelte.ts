/**
 * The upload store: every U10/U11 component's one source of state and
 * the only place that calls fetch or drives a tus upload -- components
 * are display-only over this file (the plan's U9 criteria). Session
 * state uses {@link LedgerState}, copied verbatim from
 * internal/ledger.State (the Go source of truth this file never
 * redefines independently): its landed vocabulary, not the plan prose's
 * earlier "matching"/"importing"/"pushed-back" draft, since that is
 * what `/api/journey` and `/api/history` actually serialize.
 */
import {
	chunkFloorBytes,
	createTusDriver,
	fetchJSON,
	SignedOutError,
	type RejectReason,
	type TusDriver,
	type TusUploadHandle,
} from '../tus';

// --- the ledger.State vocabulary, verbatim -------------------------------

/**
 * The one journey-state vocabulary the server emits, copied verbatim
 * from internal/ledger.State. Each state has exactly one sanctioned
 * appender server-side; this file only ever reads it off the wire, via
 * U11's journey polling.
 */
export type LedgerState =
	| 'uploading'
	| 'uploaded'
	| 'checking'
	| 'tagging'
	| 'in-library'
	| 'review'
	| 'rejected'
	| 'queued'
	| 'stalled';

// --- server config (/api/config) -----------------------------------------

/** GET /api/config's body, mirrored from internal/serve's configResponse. */
export interface ServerConfig {
	siteName: string;
	collections: string[];
	chunkMB: number;
	maxAlbumBytes: number;
	maxSessionsPerUser: number;
	reviewClockDays: number;
	formatSpec: string;
	/** Which {@link collections} entry is the private lane, per the operator's own naming. */
	privateCollection: string;
}

interface configWire {
	site_name: string;
	collections: string[];
	chunk_mb: number;
	max_album_bytes: number;
	max_sessions_per_user: number;
	review_clock_days: number;
	format_spec: string;
	private_collection: string;
}

function fromConfigWire(wire: configWire): ServerConfig {
	return {
		siteName: wire.site_name,
		collections: wire.collections,
		chunkMB: wire.chunk_mb,
		maxAlbumBytes: wire.max_album_bytes,
		maxSessionsPerUser: wire.max_sessions_per_user,
		reviewClockDays: wire.review_clock_days,
		formatSpec: wire.format_spec,
		privateCollection: wire.private_collection,
	};
}

// --- the reject wire format (internal/reject.Reason's MarshalJSON) -------

// internal/reject.Reason's wire shape is defined once in ../tus, the layer
// that first reads it off a rejected response; it is re-exported here so
// components keep importing the whole API vocabulary from one module.
export type { RejectReason };

/**
 * intake.FileVerdict as it actually serializes: the Go struct carries no
 * json tags, so the field names are its exported Go names verbatim
 * (Name/OK/Reason), not the snake_case the rest of this API uses.
 */
export interface FileVerdict {
	Name: string;
	OK: boolean;
	Reason: RejectReason;
}

// --- journey (GET /api/journey/<id>, GET /api/history) -------------------

/** internal/journey.Journey's wire shape, mirrored from internal/serve/api.go. */
interface journeyWire {
	session_id: string;
	state: LedgerState;
	detail?: string;
	album?: string;
	library_link?: string;
	files?: FileVerdict[];
}

/** One session's current journey, as U11's spine/history components read it. */
export interface Journey {
	sessionId: string;
	state: LedgerState;
	detail: string;
	album: string;
	libraryLink: string;
	files: FileVerdict[];
}

function fromJourneyWire(wire: journeyWire): Journey {
	return {
		sessionId: wire.session_id,
		state: wire.state,
		detail: wire.detail ?? '',
		album: wire.album ?? '',
		libraryLink: wire.library_link ?? '',
		files: wire.files ?? [],
	};
}

/** {@link Journey.state} widened with the two client-remembered outcomes {@link Journey} itself can never carry (see the pushed-back-memory note above). */
export type HistoryState = LedgerState | 'pushed-back' | 'probe-unavailable';

/** One row of `History.svelte`'s "Your uploads" list: a server-reported {@link Journey}, or a locally-remembered pushed-back/probe-unavailable session the ledger will never carry. */
export interface HistoryRow {
	sessionId: string;
	album: string;
	state: HistoryState;
	detail: string;
	libraryLink: string;
	files: FileVerdict[];
	/** True when a {@link ResumeAnchor} in this browser can resume this row's upload -- History's "resume" affordance. */
	resumable: boolean;
}

// --- session completion outcome (POST /api/session/<id>/finish) ----------

/**
 * internal/serve.finishResponse's wire shape: the same body both an
 * explicit POST .../finish and a lone ZIP's completing PATCH return
 * (internal/serve/upload.go's doc comment on finishResponse). The store
 * always reaches it through the explicit POST, since {@link TusDriver}
 * has no seam for reading a PATCH's response body.
 */
interface finishWire {
	promoted: boolean;
	pushed_back?: boolean;
	probe_unavailable?: boolean;
	detail?: string;
	matched_artist?: string;
	matched_album?: string;
	matched_year?: string;
	files: FileVerdict[];
}

/** A session's terminal completion outcome, per the custody-policy decision. */
export interface FinishOutcome {
	promoted: boolean;
	pushedBack: boolean;
	probeUnavailable: boolean;
	detail: string;
	matchedArtist: string;
	matchedAlbum: string;
	matchedYear: string;
	files: FileVerdict[];
}

function fromFinishWire(wire: finishWire): FinishOutcome {
	return {
		promoted: wire.promoted,
		pushedBack: wire.pushed_back ?? false,
		probeUnavailable: wire.probe_unavailable ?? false,
		detail: wire.detail ?? '',
		matchedArtist: wire.matched_artist ?? '',
		matchedAlbum: wire.matched_album ?? '',
		matchedYear: wire.matched_year ?? '',
		files: wire.files ?? [],
	};
}

/**
 * The fixed-width session-id token every completed upload's own URL is
 * prefixed with: `internal/serve/upload.go`'s `preCreate` names every
 * reserved slot's tus upload ID `"<session-id>-<slot>"`, for an explicit
 * multi-file session and an implicit lone-ZIP session alike, and the
 * session-identity plan-time decision fixes that ID at exactly 26
 * lowercase base32 characters (`^[a-z2-7]{26}$`). A lone ZIP never
 * calls `POST /api/session`, so its completed upload URL is the only
 * place the client can recover the session ID {@link Journey} polling
 * needs.
 */
const sessionIDPattern = /^[a-z2-7]{26}/;

/** Recovers a session ID from a completed upload's tus URL; null if url's last segment does not start with one. */
export function sessionIdFromUploadURL(url: string): string | null {
	const segment = url.split('/').filter(Boolean).pop() ?? '';
	const match = sessionIDPattern.exec(segment);
	return match ? match[0] : null;
}

// --- per-file client-local state -----------------------------------------

/** A file's local upload lifecycle, independent of the session's server-side {@link LedgerState}. */
export type FileStatus = 'pending' | 'uploading' | 'ok' | 'failed';

/** Why a `failed` file failed, for the copy the UI renders (passb-copy-final.md section 2). */
export type FailureKind = 'connection' | 'rejected' | null;

/** One file's full store-side state, exactly what a U10/U11 row component reads. */
export interface UploadFile {
	id: string;
	name: string;
	size: number;
	status: FileStatus;
	/** Bytes transferred over the wire so far (client-observed). */
	progressBytes: number;
	/** Bytes the server has actually confirmed via the tus offset -- distinct from progressBytes by design (M14/honest-progress). */
	ackedBytes: number;
	/** Non-null only once decideRetry has downshifted this file's chunk size below its starting value. */
	downshiftedChunkBytes: number | null;
	failureKind: FailureKind;
	/** Server verdicts for this file, populated once its session completes. */
	verdict: FileVerdict | null;
	/** The resume anchor: the tus upload URL alone, never a cached file blob. */
	tusUploadURL: string | null;
	/** True for an entry hydrated from a persisted anchor whose File object did not survive a full browser close. */
	needsReselect: boolean;
}

/** One upload session: the album/collection the files belong to and the server's journey state for it. */
export interface UploadSession {
	sessionId: string | null;
	album: string;
	collection: string;
	state: LedgerState;
	detail: string;
}

// --- resume anchor persistence (localStorage; URL only, no file bytes) ---

/** What {@link UploadStore} persists across a reload: enough to offer "resume", never enough to resume without the user. */
export interface ResumeAnchor {
	fileId: string;
	name: string;
	size: number;
	tusUploadURL: string;
}

/** The minimal storage surface the store needs, so a test can inject an in-memory fake instead of real `localStorage`. */
export type KeyValueStorage = Pick<Storage, 'getItem' | 'setItem' | 'removeItem'>;

const resumeStorageKey = 'dubplate:upload:resume-anchors';

function readAnchors(storage: KeyValueStorage | null): ResumeAnchor[] {
	if (!storage) {
		return [];
	}
	const raw = storage.getItem(resumeStorageKey);
	if (!raw) {
		return [];
	}
	try {
		const parsed = JSON.parse(raw) as unknown;
		return Array.isArray(parsed) ? (parsed as ResumeAnchor[]) : [];
	} catch {
		return [];
	}
}

function writeAnchors(storage: KeyValueStorage | null, anchors: ResumeAnchor[]): void {
	if (!storage) {
		return;
	}
	storage.setItem(resumeStorageKey, JSON.stringify(anchors));
}

// --- pushed-back/probe-unavailable memory (localStorage) ------------------
//
// internal/ledger's [LedgerState] vocabulary has no "pushed back" or
// "probe unavailable" member: per the custody-policy decision, a
// pushed-back session is never promoted, so it is never written to the
// ledger, and [journey.Resolve]'s live-session fallback has no case for
// [sessionOutcome.PushedBack]/[sessionOutcome.ProbeUnavailable] either
// (verified against internal/serve/api.go's liveSession switch) --
// this outcome exists ONLY in the synchronous POST .../finish response
// {@link finishSession} reads. Without a client-side memory, "Your
// uploads" (U11's History) could never show a pushed-back row at all,
// contradicting the plan's own criterion for it. This is that memory:
// the browser, not the server, is the only place that ever learned the
// fact, so it is the only place that can still show it after the page
// reloads -- the same reasoning [ResumeAnchor] already applies to a
// killed tab's in-flight upload.

/** One session {@link finishSession} learned was pushed back or unmatchable, remembered so {@link UploadStore.loadHistory} can still show it after a reload. */
export interface PushedBackRecord {
	sessionId: string;
	album: string;
	kind: 'pushed-back' | 'probe-unavailable';
}

const pushedBackStorageKey = 'dubplate:upload:pushed-back';

function readPushedBackRecords(storage: KeyValueStorage | null): PushedBackRecord[] {
	if (!storage) {
		return [];
	}
	const raw = storage.getItem(pushedBackStorageKey);
	if (!raw) {
		return [];
	}
	try {
		const parsed = JSON.parse(raw) as unknown;
		return Array.isArray(parsed) ? (parsed as PushedBackRecord[]) : [];
	} catch {
		return [];
	}
}

function writePushedBackRecord(storage: KeyValueStorage | null, record: PushedBackRecord): void {
	if (!storage) {
		return;
	}
	const records = readPushedBackRecords(storage).filter((r) => r.sessionId !== record.sessionId);
	records.push(record);
	storage.setItem(pushedBackStorageKey, JSON.stringify(records));
}

function defaultStorage(): KeyValueStorage | null {
	return typeof localStorage === 'undefined' ? null : localStorage;
}

function newFileId(): string {
	return typeof crypto !== 'undefined' && 'randomUUID' in crypto
		? crypto.randomUUID()
		: `f-${Math.random().toString(36).slice(2)}`;
}

// --- the store -------------------------------------------------------------

/**
 * The upload store. UI-free: every action here is either a pure state
 * mutation or a call through {@link TusDriver}/`fetch`, never a DOM
 * read. Construct with {@link createTusDriver} and `localStorage` in
 * the app (see {@link uploadStore}); a test constructs with a fake
 * driver and an in-memory {@link KeyValueStorage} instead.
 */
export class UploadStore {
	config = $state<ServerConfig | null>(null);
	session = $state<UploadSession>({ sessionId: null, album: '', collection: '', state: 'uploading', detail: '' });
	files = $state<UploadFile[]>([]);
	/** Set on a 401 or an HTML-where-JSON response from either a tus PATCH or a plain fetch; renders reload-to-sign-in. */
	signedOut = $state(false);
	/**
	 * The session's terminal completion outcome, once {@link finishSession}
	 * resolves: null while awaiting it, or for a lone-format-rejection
	 * (`!promoted && !pushedBack && !probeUnavailable`) U10's own
	 * rejected-alert copy already covers.
	 */
	journeyOutcome = $state<FinishOutcome | null>(null);
	/** The session's current polled journey (U11's steps spine); null until {@link pollJourneyOnce} first resolves. */
	journey = $state<Journey | null>(null);
	/** Every session GET /api/history returns for the signed-in contributor, plus this browser's own pushed-back memory, resumables first (see {@link loadHistory}). */
	history = $state<HistoryRow[]>([]);

	private readonly driver: TusDriver;
	private readonly storage: KeyValueStorage | null;
	private readonly handles = new Map<string, TusUploadHandle>();
	private readonly fileObjects = new Map<string, File>();
	private journeyPollHandle: ReturnType<typeof setInterval> | null = null;

	constructor(driver: TusDriver = createTusDriver(), storage: KeyValueStorage | null = defaultStorage()) {
		this.driver = driver;
		this.storage = storage;
		this.hydrateResumeAnchors();
	}

	/** Fetches /api/config; chunk size and every cap this store enforces come only from here. */
	async loadConfig(): Promise<void> {
		try {
			const wire = await fetchJSON<configWire>('/upload/api/config');
			this.config = fromConfigWire(wire);
		} catch (err) {
			if (err instanceof SignedOutError) {
				this.signedOut = true;
				return;
			}
			throw err;
		}
	}

	/**
	 * Creates a multi-file session via POST /api/session and adopts its
	 * ID; the implicit single-ZIP path never calls this (U5's pre-create
	 * mints a session ID itself when none is supplied).
	 */
	async createSession(album: string, collection: string, expectedFiles: number): Promise<string> {
		try {
			const res = await fetchJSON<{ session_id: string }>('/upload/api/session', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ album, collection, expected_files: expectedFiles }),
			});
			this.session = { sessionId: res.session_id, album, collection, state: 'uploading', detail: '' };
			return res.session_id;
		} catch (err) {
			if (err instanceof SignedOutError) {
				this.signedOut = true;
			}
			throw err;
		}
	}

	/** Adds file as a pending row and returns its local id; does not start the upload. */
	addFile(file: File): string {
		const id = newFileId();
		this.files.push({
			id,
			name: file.name,
			size: file.size,
			status: 'pending',
			progressBytes: 0,
			ackedBytes: 0,
			downshiftedChunkBytes: null,
			failureKind: null,
			verdict: null,
			tusUploadURL: null,
			needsReselect: false,
		});
		this.fileObjects.set(id, file);
		return id;
	}

	/**
	 * Starts (or resumes) id's upload. metadata carries whatever U5's
	 * pre-create hook needs for this file's path: `session_id` for a
	 * multi-file session member, or `album`/`collection` for an implicit
	 * single-ZIP session -- this store does not choose between them,
	 * since that choice belongs to U10's selection UI.
	 */
	startFile(id: string, metadata: Record<string, string>): void {
		const entry = this.find(id);
		const file = this.fileObjects.get(id);
		if (!entry || !file || !this.config) {
			return;
		}
		if (entry.needsReselect) {
			return;
		}

		entry.status = 'uploading';
		entry.failureKind = null;
		const chunkSizeBytes = entry.downshiftedChunkBytes ?? this.config.chunkMB * 1024 * 1024;

		const handle = this.driver.start(
			file,
			{
				endpoint: '/upload/files/',
				chunkSizeBytes,
				metadata: { filename: file.name, ...metadata },
				resumeURL: entry.tusUploadURL ?? undefined,
			},
			{
				onUploadURLAvailable: (uploadURL) => {
					entry.tusUploadURL = uploadURL;
					this.persistAnchor({ fileId: id, name: entry.name, size: entry.size, tusUploadURL: uploadURL });
				},
				onProgress: (progressBytes) => {
					entry.progressBytes = progressBytes;
				},
				onAcked: (ackedBytes) => {
					entry.ackedBytes = ackedBytes;
				},
				onSuccess: (uploadURL) => {
					entry.status = 'ok';
					entry.progressBytes = entry.size;
					entry.ackedBytes = entry.size;
					entry.tusUploadURL = uploadURL;
					this.forgetAnchor(id);
					this.handles.delete(id);
				},
				onSignedOut: () => {
					this.signedOut = true;
				},
				onStalled: () => {
					entry.status = 'failed';
					entry.failureKind = 'connection';
					this.handles.delete(id);
				},
				onRejected: (reason) => {
					entry.status = 'failed';
					entry.failureKind = 'rejected';
					entry.verdict = reason ? { Name: entry.name, OK: false, Reason: reason } : null;
					this.handles.delete(id);
				},
				onDownshift: (newChunkSizeBytes) => {
					entry.downshiftedChunkBytes = Math.max(chunkFloorBytes, newChunkSizeBytes);
				},
			},
		);
		this.handles.set(id, handle);
	}

	/** Retries a failed file from wherever the server last acknowledged it -- never re-sends confirmed bytes. */
	retryFile(id: string, metadata: Record<string, string>): void {
		const entry = this.find(id);
		if (!entry || entry.status !== 'failed') {
			return;
		}
		entry.status = 'pending';
		this.startFile(id, metadata);
	}

	/** Aborts an in-flight upload without discarding its server-side progress or its resume anchor. */
	abortFile(id: string): void {
		this.handles.get(id)?.abort();
		this.handles.delete(id);
	}

	// --- U11: journey, history, and session completion -----------------

	/**
	 * Runs sessionId's album verdict and (main-lane, non-private) tag
	 * check to a terminal outcome and returns it. The request is
	 * idempotent and blocks server-side until the real outcome is
	 * decided (internal/serve's `Session.waitOutcome`), so the caller
	 * never has to poll to find out whether it landed -- this is also
	 * why U11's Journey never starts {@link pollJourneyOnce} until this
	 * resolves: polling during the window this call is settling can
	 * observe the session's transient, not-yet-decided live state.
	 *
	 * album is remembered alongside a pushed-back or probe-unavailable
	 * outcome (see the pushed-back-memory note above) purely for
	 * History's display; it plays no role in the request itself.
	 */
	async finishSession(sessionId: string, album: string): Promise<FinishOutcome> {
		try {
			const wire = await fetchJSON<finishWire>(`/upload/api/session/${sessionId}/finish`, { method: 'POST' });
			const outcome = fromFinishWire(wire);
			this.journeyOutcome = outcome;
			if (outcome.pushedBack || outcome.probeUnavailable) {
				writePushedBackRecord(this.storage, {
					sessionId,
					album,
					kind: outcome.pushedBack ? 'pushed-back' : 'probe-unavailable',
				});
			}
			return outcome;
		} catch (err) {
			if (err instanceof SignedOutError) {
				this.signedOut = true;
			}
			throw err;
		}
	}

	/**
	 * Fetches sessionId's current journey once and stores it at
	 * {@link journey}. A transient failure (anything but signed-out) is
	 * swallowed rather than thrown: the 5s poll {@link startJourneyPolling}
	 * drives this with is itself the retry, and a page mid-poll should
	 * never crash on one dropped request.
	 */
	async pollJourneyOnce(sessionId: string): Promise<void> {
		try {
			const wire = await fetchJSON<journeyWire>(`/upload/api/journey/${sessionId}`);
			this.journey = fromJourneyWire(wire);
		} catch (err) {
			if (err instanceof SignedOutError) {
				this.signedOut = true;
			}
		}
	}

	/**
	 * Starts polling sessionId's journey every intervalMs (the plan's 5s
	 * default), firing once immediately rather than waiting out the
	 * first interval. Returns a stop function; {@link stopJourneyPolling}
	 * is the same operation by name, for a caller that would rather hold
	 * onto the store than the closure.
	 */
	startJourneyPolling(sessionId: string, intervalMs = 5000): () => void {
		this.stopJourneyPolling();
		void this.pollJourneyOnce(sessionId);
		this.journeyPollHandle = setInterval(() => void this.pollJourneyOnce(sessionId), intervalMs);
		return () => this.stopJourneyPolling();
	}

	/** Stops any journey poll {@link startJourneyPolling} started; a no-op if none is running. */
	stopJourneyPolling(): void {
		if (this.journeyPollHandle !== null) {
			clearInterval(this.journeyPollHandle);
			this.journeyPollHandle = null;
		}
	}

	/**
	 * Fetches GET /api/history and stores it at {@link history}: every
	 * session, in-flight and terminal, the caller's verified email has
	 * ever touched, plus this browser's own pushed-back memory for a
	 * session the ledger will never carry (the pushed-back-memory note
	 * above) -- filtered to drop a locally-remembered session the
	 * ledger has since superseded (a later ledger event for the same ID
	 * means the client's memory of it is stale). In-flight resumables
	 * sort first (U11 criterion): a row this browser can still resume
	 * via a {@link ResumeAnchor}.
	 */
	async loadHistory(): Promise<void> {
		try {
			const wire = await fetchJSON<journeyWire[]>('/upload/api/history');
			const resumableIDs = new Set(
				readAnchors(this.storage)
					.map((a) => sessionIdFromUploadURL(a.tusUploadURL))
					.filter((id): id is string => id !== null),
			);
			const ledgerRows: HistoryRow[] = wire.map((w) => {
				const j = fromJourneyWire(w);
				return {
					sessionId: j.sessionId,
					album: j.album,
					state: j.state,
					detail: j.detail,
					libraryLink: j.libraryLink,
					files: j.files,
					resumable: resumableIDs.has(j.sessionId),
				};
			});
			const knownIDs = new Set(ledgerRows.map((r) => r.sessionId));
			const localRows: HistoryRow[] = readPushedBackRecords(this.storage)
				.filter((r) => !knownIDs.has(r.sessionId))
				.map((r) => ({
					sessionId: r.sessionId,
					album: r.album,
					state: r.kind,
					detail: '',
					libraryLink: '',
					files: [],
					resumable: false,
				}));
			this.history = [...localRows, ...ledgerRows].sort((a, b) => Number(b.resumable) - Number(a.resumable));
		} catch (err) {
			if (err instanceof SignedOutError) {
				this.signedOut = true;
				return;
			}
			throw err;
		}
	}

	/**
	 * Every anchor persisted from a session this browser cannot currently
	 * resume without help: a page load after a full close has no File
	 * object for any of them, which is exactly the "re-select the file"
	 * store state the plan requires rather than a silent surprise.
	 */
	get resumableAnchors(): ResumeAnchor[] {
		return readAnchors(this.storage);
	}

	/**
	 * Re-attaches file to the hydrated {@link UploadFile} row a persisted
	 * anchor produced: only a name+size match resumes, since a
	 * differently named or sized pick is not the same upload. Returns
	 * anchor.fileId on success, null on a mismatch -- the row itself is
	 * never duplicated, only un-flagged.
	 */
	reselectFile(anchor: ResumeAnchor, file: File): string | null {
		if (file.name !== anchor.name || file.size !== anchor.size) {
			return null;
		}
		const entry = this.find(anchor.fileId);
		if (!entry) {
			return null;
		}
		entry.needsReselect = false;
		entry.tusUploadURL = anchor.tusUploadURL;
		this.fileObjects.set(anchor.fileId, file);
		return anchor.fileId;
	}

	// --- private ------------------------------------------------------

	private find(id: string): UploadFile | undefined {
		return this.files.find((f) => f.id === id);
	}

	private hydrateResumeAnchors(): void {
		for (const anchor of readAnchors(this.storage)) {
			this.files.push({
				id: anchor.fileId,
				name: anchor.name,
				size: anchor.size,
				status: 'pending',
				progressBytes: 0,
				ackedBytes: 0,
				downshiftedChunkBytes: null,
				failureKind: null,
				verdict: null,
				tusUploadURL: anchor.tusUploadURL,
				needsReselect: true,
			});
		}
	}

	private persistAnchor(anchor: ResumeAnchor): void {
		const anchors = readAnchors(this.storage).filter((a) => a.fileId !== anchor.fileId);
		anchors.push(anchor);
		writeAnchors(this.storage, anchors);
	}

	private forgetAnchor(fileId: string): void {
		writeAnchors(
			this.storage,
			readAnchors(this.storage).filter((a) => a.fileId !== fileId),
		);
	}
}

/** The app-wide store singleton; App.svelte and every U10/U11 component import this. */
export const uploadStore = new UploadStore();
