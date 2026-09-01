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
  independently of upload phase.
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
	import { supportsDirectorySelect } from './lib/platform';
	import { isZipSelection, screenSelection, type PreflightVerdict } from './lib/preflight';
	import { sessionIdFromUploadURL, uploadStore, type FileVerdict } from './lib/store/upload.svelte';

	/** The picker/drop accept list: extension-first per gate 3 (audio/* greys out FLAC on iOS Safari). */
	const acceptExtensions = '.zip,.flac';

	// A capability, not reactive state -- computed once rather than
	// re-probed (a throwaway <input> element) on every render.
	const supportsFolderSelect = supportsDirectorySelect();

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

	onMount(() => {
		void uploadStore.loadConfig().then(() => {
			if (uploadStore.config && !selectedCollection) {
				selectedCollection = uploadStore.config.collections[0] ?? '';
			}
		});
	});

	const collectionMode = $derived(isPrivateCollection(selectedCollection) ? 'private' : 'standard');

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
	}

	const currentFiles = $derived(uploadStore.files.filter((file) => uploadFileIds.includes(file.id)));
	const stalledFiles = $derived(currentFiles.filter((file) => file.status === 'failed' && file.failureKind === 'connection'));

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
	const journeyAlbum = $derived(uploadStore.session.album || deriveAlbumLabel(pendingFiles));

	/** OutcomeAlert's "Upload to Private Recordings" action: the picked files are still held in memory, so a pushed-back album re-routes through the same tag form U10 already built rather than asking for a fresh pick. */
	function retryAsPrivateCollection(): void {
		const privateCollection = uploadStore.config?.collections.find((name) => isPrivateCollection(name));
		if (!privateCollection) {
			return;
		}
		selectedCollection = privateCollection;
		tagTrackTitles = pendingFiles.map((file) => deriveTitleFromFilename(file.name));
		phase = 'tagging';
	}

	function retryStalled(): void {
		const metadata: Record<string, string> = isZipSelection(pendingFiles)
			? { album: deriveAlbumLabel(pendingFiles), collection: selectedCollection }
			: { session_id: uploadStore.session.sessionId ?? '' };
		for (const file of stalledFiles) {
			uploadStore.retryFile(file.id, metadata);
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
		{#if view === 'history'}
			<History
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
			<div class="alert alert-warning" role="alert">
				<span>Your sign-in has expired. Reload the page to sign in again.</span>
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

			<div class="hidden md:block">
				<SpecCard variant="desktop" {collectionMode} maxAlbumBytes={uploadStore.config.maxAlbumBytes} />
			</div>
			<div class="md:hidden">
				<SpecCard variant="mobile" {collectionMode} maxAlbumBytes={uploadStore.config.maxAlbumBytes} />
			</div>

			{#if phase === 'idle'}
				<div class="hidden md:block">
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
				<div class="flex flex-col items-center gap-2 md:hidden">
					<PickerButton label="Choose from Files" accept={acceptExtensions} onFiles={handleFilesPicked} />
					<p class="text-base-content/60 text-center text-sm">
						To pick several songs at once: tap <kbd class="kbd kbd-sm">⋯</kbd> in the picker, then
						<strong>Select</strong>.
					</p>
				</div>
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
					{#each pendingFiles as file (file.name)}
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
						onUploadPrivate={retryAsPrivateCollection}
						onFixNames={resetSelection}
					/>
				{:else}
					{#if isZipSelection(pendingFiles)}
						<p class="text-base-content/70 text-sm">We check the tracks once the zip finishes uploading.</p>
					{/if}
					{#if stalledFiles.length > 0}
						<div class="alert alert-warning flex-col items-start" role="alert">
							<h3 class="font-semibold">Upload stopped</h3>
							<p>
								It stopped moving, but nothing is lost. Check your connection and try again; it picks up where it left
								off.
							</p>
							<button type="button" class="btn btn-sm" onclick={retryStalled}>Retry</button>
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
