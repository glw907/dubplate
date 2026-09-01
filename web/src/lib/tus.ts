/**
 * The tus protocol boundary: the retry ladder and chunk-downshift policy
 * as pure, store-testable functions, a thin {@link TusDriver} interface
 * the upload store depends on instead of `tus-js-client` directly, and
 * the real driver built on top of it. The store's unit tests inject a
 * fake driver (docs/superpowers/plans/2026-08-31-dubplate-uploader.md's
 * "stubbed endpoint") that never touches the network; only this file's
 * {@link createTusDriver} talks to `tus-js-client` for real.
 */
import * as tus from 'tus-js-client';

/**
 * The retry ladder (plan wins over the stack brief's shorter default
 * here): five attempts, first immediate, the rest widening out to 20s,
 * so a multi-minute upload survives a transient drop without surfacing
 * as a hard failure.
 */
export const retryDelays: readonly number[] = [0, 3000, 5000, 10000, 20000];

/** The floor a stalling upload's chunk size halves down to and no further. */
export const chunkFloorBytes = 1 * 1024 * 1024;

/** The three edge-server statuses that trigger a same-offset chunk downshift. */
const downshiftStatuses = new Set([502, 504, 524]);

/**
 * How {@link decideRetry} classifies one failed request: `signed-out`
 * for a 401 or an HTML body where JSON was expected (the session
 * expired and the client is looking at Access's own login page, not the
 * API); `rejected` for any other 4xx, or a 507, which are server
 * verdicts no retry will change; `retryable` for every other 5xx.
 */
export type FailureKind = 'signed-out' | 'rejected' | 'retryable';

/**
 * Classifies status/contentType as {@link FailureKind}. An HTML
 * content-type on any status is treated as signed-out first: Access's
 * own reauth page can arrive on a 200, not only a 401. A missing or
 * sub-400 status is a network-level failure (offline, DNS, a dropped
 * connection, `status: 0` from {@link responseDetails}) rather than a
 * server verdict, so it is retryable the same as a 5xx -- otherwise the
 * most common transient failure the ladder exists for would look like a
 * terminal server rejection. 507 is carved out of the general 5xx range
 * ahead of that fallback: `internal/serve/upload.go` returns it only
 * for `reject.ReasonNoSpace`, a terminal admission-control verdict, not
 * a transient edge failure, so retrying it would spend budget on a
 * result that will not change.
 */
export function classifyFailure(status: number, contentType: string | null): FailureKind {
	if (contentType !== null && contentType.toLowerCase().startsWith('text/html')) {
		return 'signed-out';
	}
	if (status === 401) {
		return 'signed-out';
	}
	if (status === 507) {
		return 'rejected';
	}
	if (status < 400 || status >= 500) {
		return 'retryable';
	}
	return 'rejected';
}

/** One file's retry bookkeeping, threaded through successive {@link decideRetry} calls. */
export interface RetryState {
	/** How many entries of {@link retryDelays} this file has already used. */
	attempt: number;
	/** The tus offset the previous failure happened at, or null before any failure. */
	lastFailedOffset: number | null;
	/** The chunk size in effect now; only ever shrinks, down to {@link chunkFloorBytes}. */
	chunkSizeBytes: number;
}

/** One failed PATCH, as {@link decideRetry} needs to see it. */
export interface RetryFailure {
	status: number;
	contentType: string | null;
	offset: number;
}

/**
 * What the store does next after a failed PATCH, per {@link decideRetry}.
 * `stalled` is the retry ladder's own budget running out on a retryable
 * (5xx) failure -- connection trouble, the honest state with manual
 * retry (passb-copy-final.md section 2's "Upload stopped"). `rejected`
 * is a terminal 4xx other than 401, or a 507: a server verdict retrying
 * will not change, so it costs no retry budget at all rather than
 * waiting out the ladder first.
 */
export interface RetryDecision {
	action: 'retry' | 'sign-out' | 'stalled' | 'rejected';
	/** Nonzero only for `action: 'retry'`. */
	delayMs: number;
	/** The chunk size the next attempt (if any) should use. */
	chunkSizeBytes: number;
	/** True when this decision halved the chunk size, for the "slow connection" note. */
	downshifted: boolean;
}

/**
 * Decides what a file does after one failed PATCH: sign out (a 401 or
 * an HTML body -- costs no retry budget, per the plan's 401 criterion),
 * a server rejection (also no budget spent -- retrying a 4xx changes
 * nothing), or a retry, downshifting the chunk size toward
 * {@link chunkFloorBytes} when the same offset has now failed twice in
 * a row on one of the three edge-server statuses that downshift exists
 * for. Exhausting {@link retryDelays} on a retryable failure ends in
 * `stalled`, the honest state with manual retry.
 */
export function decideRetry(
	state: RetryState,
	failure: RetryFailure,
	delays: readonly number[] = retryDelays,
	floorBytes: number = chunkFloorBytes,
): RetryDecision {
	const kind = classifyFailure(failure.status, failure.contentType);
	if (kind === 'signed-out') {
		return { action: 'sign-out', delayMs: 0, chunkSizeBytes: state.chunkSizeBytes, downshifted: false };
	}
	if (kind === 'rejected') {
		return { action: 'rejected', delayMs: 0, chunkSizeBytes: state.chunkSizeBytes, downshifted: false };
	}
	if (state.attempt >= delays.length) {
		return { action: 'stalled', delayMs: 0, chunkSizeBytes: state.chunkSizeBytes, downshifted: false };
	}

	const sameOffsetAsLastFailure = state.lastFailedOffset === failure.offset;
	const downshift = sameOffsetAsLastFailure && downshiftStatuses.has(failure.status);
	const chunkSizeBytes = downshift
		? Math.max(floorBytes, Math.floor(state.chunkSizeBytes / 2))
		: state.chunkSizeBytes;

	return { action: 'retry', delayMs: delays[state.attempt], chunkSizeBytes, downshifted: downshift };
}

// --- the driver seam ----------------------------------------------------

/** Metadata tus-js-client sends with the creation POST. */
export interface TusUploadConfig {
	endpoint: string;
	chunkSizeBytes: number;
	metadata: Record<string, string>;
	/** A stored upload URL to resume from, from the store's resume anchor. */
	resumeURL?: string;
}

/** Callbacks {@link TusDriver.start} invokes as an upload progresses. */
export interface TusUploadHandlers {
	/** Fires once the creation POST returns a URL, well before completion -- the resume anchor to persist. */
	onUploadURLAvailable(uploadURL: string): void;
	onProgress(progressBytes: number, totalBytes: number): void;
	/**
	 * Fires once per completed chunk with the server-confirmed tus
	 * offset (`bytesAccepted`) -- the honest "acked" figure, distinct
	 * from `onProgress`'s client-observed transfer bytes and always
	 * lagging behind it (passb-copy-final.md section 2's "{acked_bytes}
	 * confirmed" readout).
	 */
	onAcked(bytesAccepted: number): void;
	onSuccess(uploadURL: string | null): void;
	/** A 401 or an HTML body: the session expired, no retry budget spent. */
	onSignedOut(): void;
	/** The retry ladder ran out on a retryable failure: connection trouble, manual retry available. */
	onStalled(): void;
	/**
	 * A terminal 4xx other than 401, or a 507: a server verdict, never
	 * retried. `reason` carries admission control's own rejection wire
	 * body when the response was JSON, so the store can render the real
	 * reason instead of a generic failure.
	 */
	onRejected(reason: RejectReason | null): void;
	onDownshift(chunkSizeBytes: number): void;
}

/** A running (or resumable) upload; `abort` stops it without discarding server-side progress. */
export interface TusUploadHandle {
	abort(): void;
}

/**
 * The seam the upload store depends on instead of `tus-js-client`
 * directly, so a unit test can inject a fake that never opens a socket.
 */
export interface TusDriver {
	start(file: File, config: TusUploadConfig, handlers: TusUploadHandlers): TusUploadHandle;
}

/**
 * `internal/reject.Reason`'s wire shape (`internal/reject.go`'s
 * `reasonWire`), read verbatim off a rejected request's JSON body: no
 * client-side copy of the friendly/next-step/label wording, ever. The
 * upload store re-exports this as its own `RejectReason` so the whole app
 * shares one definition of the shape.
 */
export interface RejectReason {
	code: string;
	friendly: string;
	next_step: string;
	label: string;
}

/** What one failed request tells {@link decideRetry} and the store about itself. */
interface ResponseDetails {
	status: number;
	contentType: string | null;
	reason: RejectReason | null;
}

/**
 * Reads the response status and content-type off a `tus-js-client`
 * {@link tus.DetailedError}'s attached response, when one is present. A
 * network-level failure (offline, DNS, a dropped connection) carries no
 * originalResponse and is treated as a retryable 0.
 */
function responseDetails(error: Error): ResponseDetails {
	if (!(error instanceof tus.DetailedError) || !error.originalResponse) {
		return { status: 0, contentType: null, reason: null };
	}
	const contentType = error.originalResponse.getHeader('Content-Type') ?? null;
	return {
		status: error.originalResponse.getStatus(),
		contentType,
		reason: parseRejectReason(contentType, error.originalResponse.getBody()),
	};
}

/**
 * Parses body as a {@link RejectReason} when contentType is JSON.
 * Malformed JSON, or any other content type, yields null rather than
 * throwing into the driver's `onError` handler -- a body the client
 * can't read is no worse than one that was never sent.
 */
export function parseRejectReason(contentType: string | null, body: string): RejectReason | null {
	if (contentType === null || !contentType.toLowerCase().startsWith('application/json')) {
		return null;
	}
	try {
		return JSON.parse(body) as RejectReason;
	} catch {
		return null;
	}
}

/**
 * Builds the real {@link TusDriver}: one `tus.Upload` per file, chunk
 * downshift implemented by aborting and recreating the Upload at a
 * smaller chunkSize against the same uploadUrl (tus-js-client has no
 * live-mutate option for a running upload's chunkSize; passing
 * `uploadUrl` and calling `start()` is itself the resume). retryDelays
 * is passed as `[]` to tus-js-client itself -- this driver owns the
 * retry ladder via {@link decideRetry} so the two policies can never
 * disagree about a retry's budget. `bytesSent` from the most recent
 * onProgress stands in for "the offset the failure happened at":
 * tus-js-client's onError carries no offset of its own.
 */
export function createTusDriver(): TusDriver {
	function start(file: File, config: TusUploadConfig, handlers: TusUploadHandlers): TusUploadHandle {
		const state: RetryState = { attempt: 0, lastFailedOffset: null, chunkSizeBytes: config.chunkSizeBytes };
		let lastProgressBytes = 0;
		let aborted = false;
		let current: tus.Upload | null = null;

		function launch(uploadURL: string | undefined, chunkSizeBytes: number): void {
			const upload: tus.Upload = new tus.Upload(file, {
				endpoint: config.endpoint,
				uploadUrl: uploadURL,
				chunkSize: chunkSizeBytes,
				retryDelays: [],
				storeFingerprintForResuming: true,
				metadata: config.metadata,
				onUploadUrlAvailable: () => {
					if (upload.url) {
						handlers.onUploadURLAvailable(upload.url);
					}
				},
				onProgress: (bytesSent, bytesTotal) => {
					lastProgressBytes = bytesSent;
					handlers.onProgress(bytesSent, bytesTotal);
				},
				onChunkComplete: (_chunkSize, bytesAccepted) => {
					handlers.onAcked(bytesAccepted);
				},
				onSuccess: () => {
					handlers.onSuccess(upload.url);
				},
				onError: (error) => {
					if (aborted) {
						return;
					}
					const { status, contentType, reason } = responseDetails(error);
					const decision = decideRetry(state, { status, contentType, offset: lastProgressBytes });
					state.lastFailedOffset = lastProgressBytes;

					if (decision.action === 'sign-out') {
						handlers.onSignedOut();
						return;
					}
					if (decision.action === 'stalled') {
						handlers.onStalled();
						return;
					}
					if (decision.action === 'rejected') {
						handlers.onRejected(reason);
						return;
					}

					state.attempt += 1;
					if (decision.downshifted) {
						state.chunkSizeBytes = decision.chunkSizeBytes;
						handlers.onDownshift(decision.chunkSizeBytes);
					}
					const resumeURL = upload.url ?? uploadURL;
					setTimeout(() => {
						if (!aborted) {
							launch(resumeURL, state.chunkSizeBytes);
						}
					}, decision.delayMs);
				},
			});
			current = upload;
			upload.start();
		}

		launch(config.resumeURL, config.chunkSizeBytes);

		return {
			abort() {
				aborted = true;
				void current?.abort();
			},
		};
	}

	return { start };
}

// --- JSON fetch with the same signed-out detection -----------------------

/** A JSON response, or the signed-out sentinel {@link fetchJSON} raises instead. */
export class SignedOutError extends Error {
	constructor() {
		super('signed out');
		this.name = 'SignedOutError';
	}
}

/**
 * Fetches path and decodes its JSON body, throwing {@link SignedOutError}
 * for a 401 or an HTML content-type -- the same classification
 * {@link decideRetry} applies to tus PATCHes, so `/api/config`,
 * `/api/session`, and friends render the identical reload-to-sign-in
 * state the upload path does.
 */
export async function fetchJSON<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await fetch(path, init);
	const contentType = res.headers.get('Content-Type');
	if (classifyFailure(res.status, contentType) === 'signed-out') {
		throw new SignedOutError();
	}
	if (!res.ok) {
		throw new Error(`fetch ${path}: ${res.status}`);
	}
	return (await res.json()) as T;
}
