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
	import {
		sessionIdFromUploadURL,
		uploadStore,
		type FileVerdict,
		type HistoryRow,
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

	// A resume picked from History carries no File object (only its
	// anchor's name/size/URL survive a page reload), so resuming re-opens
	// a native picker scoped to that session and waits for a matching
	// re-pick before calling the store's documented reselectFile/startFile
	// path. resumeAlbum is journeyAlbum's fallback for a resumed session,
	// which never goes through routeSelection/deriveAlbumLabel.
	let resumingRow: HistoryRow | null = null;
	let resumeAlbum = $state('');
	let resumeInputEl: HTMLInputElement | undefined = $state();

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
		void uploadStore.loadConfig().then(() => {
			if (uploadStore.config && !selectedCollection) {
				selectedCollection = uploadStore.config.collections[0] ?? '';
			}
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
		phase = 'uploading';
		const album = albumOverride ?? deriveAlbumLabel(files);
		if (isZipSelection(files)) {
			const id = uploadStore.addFile(files[0]);
			uploadFileIds = [id];
			uploadStore.startFile(id, { album, collection: selectedCollection });
			return;
		}
		uploadFileIds = files.map((file) => uploadStore.addFile(file));
		void uploadStore.createSession(album, selectedCollection, files.length).then((sessionId) => {
			for (const id of uploadFileIds) {
				uploadStore.startFile(id, { session_id: sessionId });
			}
		});
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
		// Journey's onDestroy releases the lock on the normal path, but a
		// reset from the in-flight branch never mounts Journey, and the
		// sentinel would otherwise outlive the upload it was armed for.
		void wakeLock.release();
	}

	const currentFiles = $derived(uploadStore.files.filter((file) => uploadFileIds.includes(file.id)));
	const stalledFiles = $derived(currentFiles.filter((file) => file.status === 'failed' && file.failureKind === 'connection'));

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
	const journeySessionId = $derived.by(() => {
		if (uploadStore.session.sessionId) {
			return uploadStore.session.sessionId;
		}
		const zipURL = currentFiles[0]?.tusUploadURL;
		return zipURL ? sessionIdFromUploadURL(zipURL) : null;
	});
	const journeyAlbum = $derived(uploadStore.session.album || resumeAlbum || deriveAlbumLabel(pendingFiles));

	/** OutcomeAlert's "Upload to Private Recordings" action: the picked files are still held in memory, so a pushed-back album re-routes through the same tag form U10 already built rather than asking for a fresh pick. */
	function retryAsPrivateCollection(): void {
		const config = uploadStore.config;
		if (!config) {
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
		// calls createSession and so never populates uploadStore.session.
		const metadata: Record<string, string> = isZipSelection(pendingFiles)
			? { album: deriveAlbumLabel(pendingFiles), collection: selectedCollection }
			: { session_id: uploadStore.session.sessionId ?? journeySessionId ?? '' };
		for (const file of stalledFiles) {
			uploadStore.retryFile(file.id, metadata);
		}
	}

	/** History's "Resume" affordance: opens a native picker scoped to row's session, then hands the re-picked files to {@link handleResumeFilesPicked}. */
	function handleResume(row: HistoryRow): void {
		view = 'upload';
		resumingRow = row;
		resumeInputEl?.click();
	}

	/** The resume picker's change handler: re-attaches every re-picked file that matches one of row's resumable anchors via the store's documented reselectFile path, then starts each one. */
	function handleResumeFilesPicked(): void {
		const row = resumingRow;
		const files = Array.from(resumeInputEl?.files ?? []);
		if (resumeInputEl) {
			resumeInputEl.value = '';
		}
		resumingRow = null;
		if (!row || files.length === 0) {
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
			return;
		}
		resumeAlbum = row.album;
		uploadFileIds = resumedIds;
		phase = 'uploading';
		for (const id of resumedIds) {
			uploadStore.startFile(id, { session_id: row.sessionId });
		}
	}

	// passb-copy-final.md section 4: one reason shared by every flagged
	// file collapses to one sentence (with its own singular form); any
	// mix of reasons uses the generic count-aware sentence instead.
	function composeAlertBody(rejections: PreflightVerdict[]): string {
		const labels = new Set(rejections.map((v) => v.reason?.label));
		const n = rejections.length;
		if (labels.size === 1) {
			const label = rejections[0].reason?.label ?? '';
			return n === 1 ? `One track is ${label}, and albums go in whole.` : `${n} tracks are ${label}, and albums go in whole.`;
		}
		return `${n} tracks didn't pass the checks, and albums go in whole.`;
	}

	// The reason hint renders once per distinct failing reason code
	// (Pass B conductor ruling 5), not once per flagged row.
	const distinctRejectionReasons = $derived.by(() => {
		const seen = new Set<string>();
		const reasons: NonNullable<PreflightVerdict['reason']>[] = [];
		for (const verdict of preflightRejections) {
			if (verdict.reason && !seen.has(verdict.reason.code)) {
				seen.add(verdict.reason.code);
				reasons.push(verdict.reason);
			}
		}
		return reasons;
	});

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
		<button type="button" class="btn btn-ghost btn-sm" onclick={() => (view = view === 'upload' ? 'history' : 'upload')}>
			{view === 'upload' ? 'Your uploads' : 'Back to upload'}
		</button>
	</header>

	<main class="mx-auto flex max-w-2xl flex-col gap-4 p-4">
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
		/>

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
			<div class="alert alert-warning flex-col items-start" role="alert">
				<h3 class="font-semibold">Signed out</h3>
				<p>Your sign-in expired. Reload to sign in again, then pick the upload back up from Your uploads.</p>
				<button type="button" class="btn" onclick={() => location.reload()}>Reload</button>
			</div>
		{:else if uploadStore.config}
			<p class="text-base-content/70 text-sm">
				{collectionMode === 'private' ? 'your tags, its own section' : 'add music to the library'}
			</p>

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
					<div>
						<DropZone onFiles={handleFilesPicked}>
							<PickerButton label="Choose files" accept={acceptExtensions} onFiles={handleFilesPicked} />
						</DropZone>
						<p class="text-base-content/60 mt-2 text-center text-sm">One album at a time. Albums go in whole.</p>
						{#if supportsFolderSelect}
							<div class="mt-2">
								<PickerButton
									label="Choose a folder"
									accept=""
									webkitdirectory
									emphasis="outline"
									size="default"
									onFiles={handleFilesPicked}
								/>
							</div>
						{/if}
					</div>
				{/if}
			{:else if phase === 'preflight-rejected'}
				<div class="alert alert-error flex-col items-start" role="alert">
					<h3 class="font-semibold">Couldn't add this album</h3>
					<p>{composeAlertBody(preflightRejections)}</p>
					<p>Nothing has uploaded yet.</p>
				</div>
				{#each distinctRejectionReasons as reason (reason.code)}
					<p class="text-sm">{reason.friendly} <strong>Next:</strong> {reason.next_step}</p>
				{/each}
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
				<button type="button" class="btn" onclick={resetSelection}>Try another album</button>
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
				/>
			{:else if phase === 'uploading'}
				{#if allFilesLanded && journeySessionId}
					<Journey
						sessionId={journeySessionId}
						album={journeyAlbum}
						totalFiles={currentFiles.length}
						{announcer}
						{wakeLock}
						onUploadPrivate={retryAsPrivateCollection}
						onFixNames={resetSelection}
					/>
				{:else}
					{#if isZipSelection(pendingFiles)}
						<p class="text-base-content/70 text-sm">We check the tracks once the zip finishes uploading.</p>
					{/if}
					<WakeLockBar {wakeLock} />
					{#if stalledFiles.length > 0}
						<div class="alert alert-warning flex-col items-start" role="alert">
							<h3 class="font-semibold">Upload stopped</h3>
							<p>
								It stopped moving, but nothing is lost. Check your connection and try again; it picks up where it left
								off.
							</p>
							<div class="flex items-center gap-2">
								<button type="button" class="btn" onclick={retryStalled}>Retry</button>
								{#if isZipSelection(pendingFiles)}
									<span class="text-base-content/60 text-sm">Retrying re-uploads the zip.</span>
								{/if}
							</div>
						</div>
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
