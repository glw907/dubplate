<!--
  @component
  The uploader's root shell and U10/U11's composition root: loads
  `/api/config`, then walks the selection -> pre-flight -> (private
  lane tag form) -> upload -> journey flow through U10/U11's
  display-only components, translating their callback props into the
  store's own methods (the only place in this file that touches
  `uploadStore` for anything beyond reading state). Once every file in
  the current selection reaches `'ok'`, this shell hands off to U11's
  `Journey` for the post-upload steps spine, outcomes, and completion;
  a "Your uploads" link swaps the whole main region for U11's `History`
  independently of upload phase. This is also where U11's one shared
  {@link LiveRegion} lives: the `aria-live` element renders unconditionally
  in `<main>` from first paint (not only once `Journey` mounts), and an
  effect announces the current selection's aggregate server-acknowledged
  progress through it while `phase` is `'uploading'` -- a screen-reader
  user hears progress during the transfer itself, not only after every
  file lands.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import AlbumCard, { deriveTitleFromFilename } from './lib/components/AlbumCard.svelte';
	import AlertCard from './lib/components/AlertCard.svelte';
	import CollectionPicker, { isPrivateCollection } from './lib/components/CollectionPicker.svelte';
	import DropZone from './lib/components/DropZone.svelte';
	import FileRow from './lib/components/FileRow.svelte';
	import History from './lib/components/History.svelte';
	import Journey from './lib/components/Journey.svelte';
	import PickerButton from './lib/components/PickerButton.svelte';
	import SpecCard from './lib/components/SpecCard.svelte';
	import WakeLockBar from './lib/components/WakeLockBar.svelte';
	import { LiveRegion } from './lib/announce';
	import { isIOSSafari, supportsDirectorySelect } from './lib/platform';
	import { isZipSelection, screenSelection, type PreflightVerdict } from './lib/preflight';
	import { FetchRejectedError, SignedOutError } from './lib/tus';
	import {
		sessionIdFromUploadURL,
		uploadStore,
		type FileVerdict,
		type HistoryRow,
		type RejectReason,
	} from './lib/store/upload.svelte';
	import { WakeLock } from './lib/wakeLock.svelte';

	/** The picker/drop accept list: extension-first per gate 3 (audio/* greys out FLAC on iOS Safari). */
	const acceptExtensions = '.zip,.flac';

	// Capabilities, not reactive state -- computed once rather than
	// re-probed (a throwaway <input> element, `navigator.userAgent`) on
	// every render. isIOS drives the copy switch off the real platform
	// (gate 3's rulings) rather than a CSS breakpoint an iPad's wide
	// viewport can slip past (Pass B review finding).
	const supportsFolderSelect = supportsDirectorySelect();
	const isIOS = isIOSSafari();

	type Phase = 'idle' | 'preflight-rejected' | 'tagging' | 'uploading';
	type View = 'upload' | 'history';

	let phase = $state<Phase>('idle');
	let view = $state<View>('upload');
	let selectedCollection = $state('');
	let pendingFiles = $state<File[]>([]);
	let preflightRejections = $state<PreflightVerdict[]>([]);
	let uploadFileIds = $state<string[]>([]);
	let tagArtist = $state('');
	let tagAlbum = $state('');
	let tagYear = $state('');
	let tagTrackTitles = $state<string[]>([]);

	// Whether /api/config ever resolved: distinguishes the transient
	// first-paint window (nothing to show yet) from a genuine failure
	// (something to show and a way to recover), which the bare
	// `uploadStore.config` truthiness check used to conflate into the
	// same blank <main> (U13 fold F5.8).
	let configState = $state<'loading' | 'ready' | 'failed'>('loading');

	// The session id this selection is actually uploading against, set
	// the moment it is known (createSession's resolved id; a resumed
	// row's own id) rather than read from uploadStore.session, which
	// {@link UploadStore.resetSession} clears between uploads but a lone
	// ZIP selection never populates at all -- see journeySessionId (U13
	// fold F5.7: this is what stops a second album from polling the
	// first album's session).
	let currentSessionId = $state<string | null>(null);

	// Set when POST /api/session itself fails (U13 fold F5.6): a parsed
	// {@link RejectReason} for a server verdict (the session cap, most
	// likely), or 'network' for anything that carries none. While this
	// is set every row otherwise reads a false, permanent "waiting".
	let sessionCreateError = $state<RejectReason | 'network' | null>(null);

	// The album POST /api/session was last called with, remembered so a
	// failed creation can retry the identical request rather than
	// re-deriving it from whatever pendingFiles/tagAlbum hold by the
	// time the user presses Retry.
	let sessionCreateAlbum = '';

	// A resume picked from History carries no File object (only its
	// anchor's name/size/URL survive a page reload), so resuming re-opens
	// a native picker scoped to that session and waits for a matching
	// re-pick before calling the store's documented reselectFile/startFile
	// path. resumeAlbum is journeyAlbum's fallback for a resumed session,
	// which never goes through routeSelection/deriveAlbumLabel.
	let resumingRow: HistoryRow | null = null;
	let resumeAlbum = $state('');
	let resumeInputEl: HTMLInputElement | undefined = $state();

	// Resume's own recovery target: handleResume swaps `view` away from
	// History before the native picker resolves, which unmounts the
	// "Resume" button the user just activated and drops focus to
	// <body> with no announcement (U13 fold F5.10). <main> is a stable
	// focus target across every phase the resume can land in.
	let mainEl: HTMLElement | undefined = $state();

	// The one shared aria-live region (U11 criterion): constructed once
	// here, attached to the element below that renders on every phase,
	// and passed to Journey so its own state/completion announcements
	// land in the same region rather than a second one.
	const announcer = new LiveRegion();
	let liveRegionEl: HTMLDivElement | undefined = $state();

	// The one shared wake-lock toggle: constructed once here so a lock
	// requested during the in-flight transfer below survives the
	// hand-off into Journey instead of dropping and re-arming.
	const wakeLock = new WakeLock();

	onMount(() => {
		if (liveRegionEl) {
			announcer.attach(liveRegionEl);
		}
		void uploadStore
			.loadConfig()
			.then(() => {
				if (uploadStore.signedOut) {
					return;
				}
				if (!uploadStore.config) {
					configState = 'failed';
					return;
				}
				if (!selectedCollection) {
					selectedCollection = uploadStore.config.collections[0] ?? '';
				}
				configState = 'ready';
			})
			.catch(() => {
				configState = 'failed';
			});
	});

	const collectionMode = $derived(
		isPrivateCollection(selectedCollection, uploadStore.config?.privateCollection ?? '') ? 'private' : 'standard',
	);

	// The zip's own filename (minus extension) is the only album label a
	// selection carries on the main lane -- there is no "album name" field
	// in the copy contract, since MusicBrainz resolves album identity
	// during the tag check (U5); a multi-file pick has no comparable
	// signal, so it falls back to a placeholder the match step overrides.
	function deriveAlbumLabel(files: File[]): string {
		if (isZipSelection(files)) {
			const name = files[0].name;
			const dot = name.lastIndexOf('.');
			return dot > 0 ? name.slice(0, dot) : name;
		}
		return 'Album';
	}

	function handleFilesPicked(files: File[]): void {
		pendingFiles = files;
		if (isZipSelection(files)) {
			routeSelection(files);
			return;
		}
		const verdicts = screenSelection(files);
		const rejected = verdicts.filter((v) => v.reason !== null);
		if (rejected.length > 0) {
			preflightRejections = rejected;
			phase = 'preflight-rejected';
			return;
		}
		routeSelection(files);
	}

	function routeSelection(files: File[]): void {
		if (collectionMode === 'private') {
			tagTrackTitles = files.map((file) => deriveTitleFromFilename(file.name));
			phase = 'tagging';
			return;
		}
		beginUpload(files);
	}

	function beginUpload(files: File[], albumOverride?: string): void {
		uploadStore.resetSession();
		currentSessionId = null;
		sessionCreateError = null;
		phase = 'uploading';
		const album = albumOverride ?? deriveAlbumLabel(files);
		if (isZipSelection(files)) {
			const id = uploadStore.addFile(files[0]);
			uploadFileIds = [id];
			uploadStore.startFile(id, { album, collection: selectedCollection });
			return;
		}
		uploadFileIds = files.map((file) => uploadStore.addFile(file));
		sessionCreateAlbum = album;
		createMultiFileSession(album, files.length);
	}

	/**
	 * POSTs /api/session for the current {@link uploadFileIds} and starts
	 * each file once it resolves. Split out of {@link beginUpload} so a
	 * failed creation can retry the same request without re-adding the
	 * files (which would mint duplicate store rows) -- see
	 * {@link retrySessionCreate}.
	 */
	function createMultiFileSession(album: string, expectedFiles: number): void {
		void uploadStore
			.createSession(album, selectedCollection, expectedFiles)
			.then((sessionId) => {
				currentSessionId = sessionId;
				for (const id of uploadFileIds) {
					uploadStore.startFile(id, { session_id: sessionId });
				}
			})
			.catch((err: unknown) => {
				// A 401 already renders the app-wide reload-to-sign-in state
				// (uploadStore.signedOut); nothing further to show here.
				if (err instanceof SignedOutError) {
					return;
				}
				sessionCreateError = err instanceof FetchRejectedError ? (err.reason ?? 'network') : 'network';
			});
	}

	/** Retries a failed session creation against the same selection, without re-adding the already-added files (U13 fold F5.6). */
	function retrySessionCreate(): void {
		sessionCreateError = null;
		createMultiFileSession(sessionCreateAlbum, uploadFileIds.length);
	}

	function submitTagForm(): void {
		// Artist/year/track titles have nowhere to go yet: POST /api/session
		// carries only album/collection/expected_files, and the tus
		// pre-create hook reads no per-track tag metadata (verified against
		// internal/serve/upload.go at U10 time). The album title is real --
		// createSession's own Album field -- so it rides along; the rest is
		// a known gap for the task that adds a private-lane tag endpoint.
		beginUpload(pendingFiles, tagAlbum || undefined);
	}

	function resetSelection(): void {
		phase = 'idle';
		pendingFiles = [];
		preflightRejections = [];
		currentSessionId = null;
		sessionCreateError = null;
		// U13 fold F5.7: without this, the session id/album/journey from
		// the album just finished (or abandoned) outlives it and can leak
		// into the next selection's polling and metadata.
		uploadStore.resetSession();
		// Journey's onDestroy releases the lock on the normal path, but a
		// reset from the in-flight branch never mounts Journey, and the
		// sentinel would otherwise outlive the upload it was armed for.
		void wakeLock.release();
	}

	const currentFiles = $derived(uploadStore.files.filter((file) => uploadFileIds.includes(file.id)));
	const stalledFiles = $derived(currentFiles.filter((file) => file.status === 'failed' && file.failureKind === 'connection'));
	// A server rejection reached mid-transfer (U13 fold F5.5): distinct
	// from sessionCreateError, which fires before any byte moves.
	const rejectedFiles = $derived(currentFiles.filter((file) => file.status === 'failed' && file.failureKind === 'rejected'));

	// The live region's progress voice: the aggregate server-acknowledged
	// percentage across the current selection (never the raw transfer
	// count -- the store's own honest-progress split), announced through
	// the one shared region above as the store's onAcked callbacks land.
	// shouldAnnounceProgress's own threshold silences everything below a
	// 25-point delta, so this effect can run on every tick without
	// spamming the region.
	$effect(() => {
		if (phase !== 'uploading' || currentFiles.length === 0) {
			return;
		}
		const totalBytes = currentFiles.reduce((sum, file) => sum + file.size, 0);
		if (totalBytes === 0) {
			return;
		}
		const ackedBytes = currentFiles.reduce((sum, file) => sum + file.ackedBytes, 0);
		announcer.announceProgress((ackedBytes / totalBytes) * 100);
	});

	// U11's Journey takes over once every file in the current selection
	// has landed: a multi-file session already carries its own ID from
	// createSession; a lone ZIP never calls that endpoint, so its
	// session ID only exists in the completed upload's own URL (see
	// sessionIdFromUploadURL's doc comment for why that is safe to
	// parse).
	const allFilesLanded = $derived(currentFiles.length > 0 && currentFiles.every((file) => file.status === 'ok'));
	// U13 fold F5.7: currentSessionId (set explicitly by beginUpload/
	// handleResumeFilesPicked for this selection) wins over the store's
	// own session field, which resetSession clears between uploads but
	// which a lone ZIP selection never populates -- the completed
	// upload's own URL is that path's only source for its session id.
	const journeySessionId = $derived.by(() => {
		if (currentSessionId) {
			return currentSessionId;
		}
		const zipURL = currentFiles[0]?.tusUploadURL;
		return zipURL ? sessionIdFromUploadURL(zipURL) : null;
	});
	const journeyAlbum = $derived(uploadStore.session.album || resumeAlbum || deriveAlbumLabel(pendingFiles));

	/**
	 * OutcomeAlert's "Upload to Private Recordings" action: the picked
	 * files are still held in memory, so a pushed-back album re-routes
	 * through the same tag form U10 already built rather than asking
	 * for a fresh pick. Guarded against an empty selection (U13 fold
	 * F5.13): a row opened from History is, by definition, a past
	 * session, so pendingFiles is empty after any reload -- without the
	 * guard this opens a zero-track tag form that can never submit.
	 */
	function retryAsPrivateCollection(): void {
		const config = uploadStore.config;
		if (!config) {
			return;
		}
		if (pendingFiles.length === 0) {
			announcer.announceState("That album's files aren't in this browser tab anymore. Pick it again to send it to Private Recordings.");
			return;
		}
		const privateCollection = config.collections.find((name) => isPrivateCollection(name, config.privateCollection));
		if (!privateCollection) {
			return;
		}
		selectedCollection = privateCollection;
		tagTrackTitles = pendingFiles.map((file) => deriveTitleFromFilename(file.name));
		phase = 'tagging';
	}

	function retryStalled(): void {
		// journeySessionId also covers a resumed session, which never
		// calls createSession and so never populates currentSessionId
		// until the resume itself sets it.
		const metadata: Record<string, string> = isZipSelection(pendingFiles)
			? { album: deriveAlbumLabel(pendingFiles), collection: selectedCollection }
			: { session_id: currentSessionId ?? journeySessionId ?? '' };
		for (const file of stalledFiles) {
			uploadStore.retryFile(file.id, metadata);
		}
	}

	/** Moves focus to a stable, always-mounted target and announces message -- {@link handleResumeFilesPicked}'s recovery from the view swap {@link handleResume} makes before the native picker resolves (U13 fold F5.10). */
	function focusMainAndAnnounce(message: string): void {
		mainEl?.focus();
		announcer.announceState(message);
	}

	/** History's "Resume" affordance: opens a native picker scoped to row's session, then hands the re-picked files to {@link handleResumeFilesPicked}. */
	function handleResume(row: HistoryRow): void {
		view = 'upload';
		resumingRow = row;
		resumeInputEl?.click();
	}

	/**
	 * The resume picker's change (and cancel) handler: re-attaches every
	 * re-picked file that matches one of row's resumable anchors via the
	 * store's documented reselectFile path, then starts each one. Every
	 * exit restores focus and announces the outcome (U13 fold F5.10):
	 * `handleResume` already unmounted the "Resume" button this picker
	 * was opened from, so without this the user is left with focus at
	 * <body> and no indication anything happened, silently on a cancel
	 * or a non-matching pick alike.
	 */
	function handleResumeFilesPicked(): void {
		const row = resumingRow;
		const files = Array.from(resumeInputEl?.files ?? []);
		if (resumeInputEl) {
			resumeInputEl.value = '';
		}
		resumingRow = null;
		if (!row || files.length === 0) {
			focusMainAndAnnounce('No file was picked to resume.');
			return;
		}

		const anchors = uploadStore.resumableAnchors.filter(
			(anchor) => sessionIdFromUploadURL(anchor.tusUploadURL) === row.sessionId,
		);
		const resumedIds: string[] = [];
		for (const anchor of anchors) {
			const file = files.find((f) => f.name === anchor.name && f.size === anchor.size);
			const id = file ? uploadStore.reselectFile(anchor, file) : null;
			if (id) {
				resumedIds.push(id);
			}
		}
		if (resumedIds.length === 0) {
			focusMainAndAnnounce("Those files don't match the album being resumed.");
			return;
		}
		// U13 fold F5.7: clears any session/journey state the previous
		// selection left behind before this one adopts row's own id.
		uploadStore.resetSession();
		// U13 fold F5.6 review fix: a failed session creation from the
		// selection just abandoned must not survive into the resumed
		// selection's own uploading phase -- without this, resuming here
		// renders the previous album's "Couldn't add this album" alert
		// over this album's rows.
		sessionCreateError = null;
		resumeAlbum = row.album;
		uploadFileIds = resumedIds;
		currentSessionId = row.sessionId;
		phase = 'uploading';
		focusMainAndAnnounce(`Resuming ${row.album || 'the album'}.`);
		for (const id of resumedIds) {
			uploadStore.startFile(id, { session_id: row.sessionId });
		}
	}

	// passb-copy-final.md section 4: one reason shared by every flagged
	// file collapses to one sentence (with its own singular form); any
	// mix of reasons uses the generic count-aware sentence instead. Takes
	// bare reasons (not PreflightVerdict[]) so the same composition
	// serves both the at-selection pre-flight and a mid-upload server
	// rejection (U13 fold F5.5), which has no PreflightVerdict of its
	// own -- only each file's own server verdict.
	function composeAlertBody(reasons: (RejectReason | null)[]): string {
		const labels = new Set(reasons.map((r) => r?.label));
		const n = reasons.length;
		if (labels.size === 1) {
			const label = reasons[0]?.label ?? '';
			return n === 1 ? `One track is ${label}, and albums go in whole.` : `${n} tracks are ${label}, and albums go in whole.`;
		}
		return `${n} tracks didn't pass the checks, and albums go in whole.`;
	}

	// The reason hint renders once per distinct failing reason code
	// (Pass B conductor ruling 5), not once per flagged row.
	function distinctReasons(reasons: (RejectReason | null)[]): RejectReason[] {
		const seen = new Set<string>();
		const out: RejectReason[] = [];
		for (const reason of reasons) {
			if (reason && !seen.has(reason.code)) {
				seen.add(reason.code);
				out.push(reason);
			}
		}
		return out;
	}

	const preflightVerdictsByFile = $derived(new Map(preflightRejections.map((v) => [v.file, v.reason])));

	function preflightFileVerdict(file: File): FileVerdict | null {
		const reason = preflightVerdictsByFile.get(file);
		return reason ? { Name: file.name, OK: false, Reason: reason } : null;
	}
</script>

<div class="bg-base-100 text-base-content min-h-screen">
	<header class="navbar bg-base-200 flex items-center justify-between">
		<span class="text-xl font-semibold uppercase tracking-wide">
			{collectionMode === 'private' ? 'Private Recordings' : (uploadStore.config?.siteName ?? 'dubplate')}
		</span>
		<button
			type="button"
			class="btn btn-outline btn-sm min-h-11"
			onclick={() => (view = view === 'upload' ? 'history' : 'upload')}
		>
			{view === 'upload' ? 'Your uploads' : 'Back to upload'}
		</button>
	</header>

	<!-- U13 fold F5.5/F5.6: the section-4 "Couldn't add this album" block, shared by the at-selection pre-flight, a mid-upload server rejection, and a failed session creation so the three paths can never drift on wording. preUpload adds the "Nothing has uploaded yet." line the contract reserves for the pre-flight case alone. perTrack gates the count-aware composed sentence (section 4), which only makes sense for a set of per-track verdicts -- a failed session creation carries one session-level reason, not a track verdict, so composing "One track is..." over it is nonsense (U13 fold F5.6 review fix). -->
	{#snippet rejectedAlert(reasons: (RejectReason | null)[], preUpload: boolean, perTrack: boolean)}
		<AlertCard tone="error" title="Couldn't add this album">
			{#if perTrack}
				<p class="text-base-content/80">{composeAlertBody(reasons)}</p>
			{/if}
			{#if preUpload}
				<p class="text-base-content/80">Nothing has uploaded yet.</p>
			{/if}
		</AlertCard>
		{#each distinctReasons(reasons) as reason (reason.code)}
			<p class="text-sm">{reason.friendly} <strong>Next:</strong> {reason.next_step}</p>
		{/each}
	{/snippet}

	<!-- Section 2's "it stopped moving, but nothing is lost" copy, shared by the two ways an upload stalls: a failed POST /api/session before any byte moves, and a connection failure mid-transfer. The zip hint rides along only on the second, where the retry re-sends a whole archive rather than resuming. -->
	{#snippet stoppedAlert(onRetry: () => void, zipHint: boolean)}
		<AlertCard tone="warning" title="Upload stopped">
			<p class="text-base-content/80">
				It stopped moving, but nothing is lost. Check your connection and try again; it picks up where it left off.
			</p>
			<div class="flex items-center gap-2">
				<button type="button" class="btn min-h-11" onclick={onRetry}>Retry</button>
				{#if zipHint}
					<span class="text-base-content/60 text-sm">Retrying re-uploads the zip.</span>
				{/if}
			</div>
		</AlertCard>
	{/snippet}

	<main bind:this={mainEl} tabindex="-1" class="mx-auto flex max-w-2xl flex-col gap-4 p-4">
		<!-- U11's one shared aria-live region: exists empty from first render, across every phase, not only once Journey mounts. -->
		<div bind:this={liveRegionEl} aria-live="polite" class="sr-only" role="status"></div>
		<!-- The resume picker's hidden input: always mounted so handleResume can trigger it from History regardless of the current phase. -->
		<input
			bind:this={resumeInputEl}
			type="file"
			class="hidden"
			accept={acceptExtensions}
			multiple
			onchange={handleResumeFilesPicked}
			oncancel={handleResumeFilesPicked}
		/>

		<!-- The content column's own title (U13 fold F5.2/F5.12): the masthead's site name is a navbar label, not a heading, so without this the document carries no h1 at all and the outline opens on whatever alert happens to render first. Rendered on every view (review fix) so the history view also opens on an h1, not directly on History's own h2. The eyebrow line is upload-only copy, so it stays gated. -->
		<div>
			<h1 class="text-2xl font-bold">
				{collectionMode === 'private' ? 'Private Recordings' : (uploadStore.config?.siteName ?? 'dubplate')}
			</h1>
			{#if view === 'upload'}
				<p class="text-primary text-sm font-semibold uppercase tracking-wide">
					{collectionMode === 'private' ? 'your tags, its own section' : 'add music to the library'}
				</p>
			{/if}
		</div>

		{#if view === 'history'}
			<History
				onResume={handleResume}
				onUploadPrivate={() => {
					view = 'upload';
					retryAsPrivateCollection();
				}}
				onFixNames={() => {
					view = 'upload';
					resetSelection();
				}}
			/>
		{:else if uploadStore.signedOut}
			<AlertCard tone="warning" title="Signed out">
				<p class="text-base-content/80">
					Your sign-in expired. Reload to sign in again, then pick the upload back up from Your uploads.
				</p>
				<button type="button" class="btn" onclick={() => location.reload()}>Reload</button>
			</AlertCard>
		{:else if configState === 'failed'}
			<AlertCard tone="error" title="Couldn't load the uploader">
				<p class="text-base-content/80">Something went wrong loading this page's setup. Reload to try again.</p>
				<button type="button" class="btn" onclick={() => location.reload()}>Reload</button>
			</AlertCard>
		{:else if uploadStore.config}
			<CollectionPicker
				collections={uploadStore.config.collections}
				selected={selectedCollection}
				onSelect={(name) => (selectedCollection = name)}
			/>

			<SpecCard variant={isIOS ? 'mobile' : 'desktop'} {collectionMode} maxAlbumBytes={uploadStore.config.maxAlbumBytes} />

			{#if phase === 'idle'}
				{#if isIOS}
					<div class="flex flex-col items-center gap-2">
						<PickerButton label="Choose from Files" accept={acceptExtensions} onFiles={handleFilesPicked} />
						<p class="text-base-content/60 text-center text-sm">
							To pick several songs at once: tap <kbd class="kbd kbd-sm">⋯</kbd> in the picker, then
							<strong>Select</strong>.
						</p>
					</div>
				{:else}
					<!-- One coherent, single-aligned picker group (U13 fold F5.1): the dashed drop panel, the full-width primary, its hint, and the folder secondary all share one column, none narrower or wider than the rest. -->
					<div class="flex flex-col items-center gap-3">
						<DropZone onFiles={handleFilesPicked}>
							<PickerButton label="Choose files" accept={acceptExtensions} onFiles={handleFilesPicked} />
						</DropZone>
						<p class="text-base-content/60 text-center text-sm">One album at a time. Albums go in whole.</p>
						{#if supportsFolderSelect}
							<PickerButton
								label="Choose a folder"
								accept=""
								webkitdirectory
								emphasis="outline"
								onFiles={handleFilesPicked}
							/>
						{/if}
					</div>
				{/if}
			{:else if phase === 'preflight-rejected'}
				{@render rejectedAlert(
					preflightRejections.map((verdict) => verdict.reason),
					true,
					true,
				)}
				<ul class="border-base-300 divide-base-300 divide-y rounded border">
					{#each pendingFiles as file, index (`${index}-${file.name}`)}
						{@const verdict = preflightFileVerdict(file)}
						<FileRow
							name={file.name}
							size={file.size}
							status={verdict ? 'failed' : 'ok'}
							progressBytes={0}
							ackedBytes={0}
							{verdict}
							failureKind={verdict ? 'rejected' : null}
							variant="verdict"
						/>
					{/each}
				</ul>
				<button type="button" class="btn min-h-11" onclick={resetSelection}>Try another album</button>
			{:else if phase === 'tagging'}
				<AlbumCard
					trackFilenames={pendingFiles.map((file) => file.name)}
					trackTitles={tagTrackTitles}
					artist={tagArtist}
					album={tagAlbum}
					year={tagYear}
					onArtistInput={(value) => (tagArtist = value)}
					onAlbumInput={(value) => (tagAlbum = value)}
					onYearInput={(value) => (tagYear = value)}
					onTrackTitleInput={(index, value) => (tagTrackTitles[index] = value)}
					onSubmit={submitTagForm}
					onCancel={resetSelection}
				/>
			{:else if phase === 'uploading'}
				{#if sessionCreateError}
					{#if sessionCreateError === 'network'}
						{@render stoppedAlert(retrySessionCreate, false)}
					{:else}
						{@render rejectedAlert([sessionCreateError], false, false)}
						<button type="button" class="btn min-h-11" onclick={resetSelection}>Try another album</button>
					{/if}
				{:else if allFilesLanded && journeySessionId}
					<Journey
						sessionId={journeySessionId}
						album={journeyAlbum}
						totalFiles={currentFiles.length}
						{announcer}
						{wakeLock}
						onUploadPrivate={retryAsPrivateCollection}
						onFixNames={resetSelection}
						onUploadAnother={resetSelection}
					/>
				{:else}
					{#if isZipSelection(pendingFiles)}
						<p class="text-base-content/70 text-sm">We check the tracks once the zip finishes uploading.</p>
					{/if}
					<WakeLockBar {wakeLock} />
					{#if rejectedFiles.length > 0}
						{@render rejectedAlert(
							rejectedFiles.map((file) => file.verdict?.Reason ?? null),
							false,
							true,
						)}
						<button type="button" class="btn min-h-11" onclick={resetSelection}>Try another album</button>
					{/if}
					{#if stalledFiles.length > 0}
						{@render stoppedAlert(retryStalled, isZipSelection(pendingFiles))}
					{/if}
					<ul class="border-base-300 divide-base-300 divide-y rounded border">
						{#each currentFiles as file (file.id)}
							<FileRow
								name={file.name}
								size={file.size}
								status={file.status}
								progressBytes={file.progressBytes}
								ackedBytes={file.ackedBytes}
								verdict={file.verdict}
								failureKind={file.failureKind}
								variant="progress"
							/>
						{/each}
					</ul>
				{/if}
			{/if}
		{/if}
	</main>
</div>
