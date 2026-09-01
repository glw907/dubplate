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
}

interface configWire {
	site_name: string;
	collections: string[];
	chunk_mb: number;
	max_album_bytes: number;
	max_sessions_per_user: number;
	review_clock_days: number;
	format_spec: string;
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
	};
}

// --- the reject wire format (internal/reject.Reason's MarshalJSON) -------

/** internal/reject.Reason's wire shape: verbatim server copy, no client lookup table. */
export interface RejectReason {
	code: string;
	friendly: string;
	next_step: string;
	label: string;
}

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

	private readonly driver: TusDriver;
	private readonly storage: KeyValueStorage | null;
	private readonly handles = new Map<string, TusUploadHandle>();
	private readonly fileObjects = new Map<string, File>();

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
				onRejected: () => {
					entry.status = 'failed';
					entry.failureKind = 'rejected';
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
