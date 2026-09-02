<!--
  @component
  The post-upload journey (passb-copy-final.md sections 2 and 3; Pass B
  conductor rulings 4 and 5): drives {@link JourneySteps}'s frozen
  four-step spine plus whichever of {@link OutcomeAlert}'s three branch
  outcomes or the terminal "In your library" card applies, over U9's
  store. Mounting calls `store.finishSession` (the synchronous,
  server-blocking custody-policy probe) exactly once; only once it
  resolves `promoted` does this component start `store.startJourneyPolling`
  -- starting it earlier would read the session's live state mid-decision,
  a transient window `internal/serve/api.go`'s own live-session switch
  does not cover (verified against the landed backend, not the plan's
  earlier draft). `announcer` is App.svelte's one shared {@link LiveRegion}
  instance, already attached to the one aria-live element that exists
  empty from first render (App.svelte's main region, mounted for every
  upload phase, not only this one) -- every step and outcome change
  announces through it once, and completion gets its own distinct call;
  this component never owns the DOM region itself. Journey polling is
  driven by an `$effect` keyed on `outcome?.promoted` and `sessionId`,
  not a variable assigned from `runFinish`'s `finally` (U13 fold
  F6.10): `finally` runs after `finishSession`'s server-blocking await
  resolves, which can be well after this component unmounts, so a
  stop function assigned there could arrive too late for `onDestroy`
  to find it and this component would leak a 5s poll for the rest of
  the page's life. An effect's own returned teardown is guaranteed by
  Svelte to run on unmount regardless of when it started.
-->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import FileRow from './FileRow.svelte';
	import JourneySteps, { type StepStatus } from './JourneySteps.svelte';
	import OutcomeAlert from './OutcomeAlert.svelte';
	import WakeLockBar from './WakeLockBar.svelte';
	import { LiveRegion } from '../announce';
	import { SignedOutError } from '../tus';
	import {
		uploadStore,
		type FileVerdict,
		type FinishOutcome,
		type RejectReason,
		type UploadStore,
	} from '../store/upload.svelte';
	import { WakeLock } from '../wakeLock.svelte';

	interface Props {
		sessionId: string;
		album: string;
		totalFiles: number;
		store?: UploadStore;
		/** The app's one shared live region; defaults to an unattached instance so this component still renders standalone (tests, Storybook-style isolation). */
		announcer?: LiveRegion;
		/** The app's one shared wake-lock toggle (App.svelte's in-flight branch renders the same instance); defaults to a fresh one so this component still renders standalone. */
		wakeLock?: WakeLock;
		onUploadPrivate?: () => void;
		onFixNames?: () => void;
		/** The in-library card's "Upload another album" action (U13 fold F6.11); App.svelte wires this to the same reset the pushed-back/rejected paths already use. */
		onUploadAnother?: () => void;
	}

	let {
		sessionId,
		album,
		totalFiles,
		store = uploadStore,
		announcer = new LiveRegion(),
		wakeLock = new WakeLock(),
		onUploadPrivate = () => {},
		onFixNames = () => {},
		onUploadAnother = () => {},
	}: Props = $props();

	let finishing = $state(true);
	let outcome = $state<FinishOutcome | null>(null);
	// Set in runFinish's catch on anything but a signed-out rejection (U13
	// fold F6.9): finishSession rethrows on a network drop, a 500, or a
	// non-JSON body, and without this outcome simply stays null forever,
	// which viewMode below used to read as the ordinary still-resolving
	// case -- freezing the spine on "Matching" with no error and no way
	// to retry short of a page reload.
	let finishError = $state(false);

	const journey = $derived(store.journey);

	type ViewMode =
		| 'finishing'
		| 'matched'
		| 'queued'
		| 'stalled'
		| 'in-library'
		| 'pushed-back'
		| 'probe-unavailable'
		| 'format-rejected'
		| 'finish-failed';

	const viewMode = $derived.by((): ViewMode => {
		if (finishError) {
			return 'finish-failed';
		}
		if (finishing || !outcome) {
			return 'finishing';
		}
		if (outcome.pushedBack) {
			return 'pushed-back';
		}
		if (outcome.probeUnavailable) {
			return 'probe-unavailable';
		}
		if (!outcome.promoted) {
			return 'format-rejected';
		}
		if (journey?.state === 'in-library') {
			return 'in-library';
		}
		if (journey?.state === 'stalled') {
			return 'stalled';
		}
		if (journey?.state === 'queued') {
			return 'queued';
		}
		return 'matched';
	});

	const statuses = $derived.by((): [StepStatus, StepStatus, StepStatus, StepStatus] => {
		switch (viewMode) {
			case 'finishing':
				return ['complete', 'complete', 'current', 'upcoming'];
			case 'pushed-back':
			case 'probe-unavailable':
			case 'finish-failed':
				return ['complete', 'complete', 'error', 'upcoming'];
			case 'format-rejected':
				return ['complete', 'error', 'upcoming', 'upcoming'];
			case 'stalled':
				return ['complete', 'complete', 'complete', 'error'];
			case 'queued':
				return ['complete', 'complete', 'complete', 'upcoming'];
			default:
				return ['complete', 'complete', 'complete', 'current'];
		}
	});

	// passb-copy-final.md section 4: mirrors App.svelte's composeAlertBody
	// over a finished session's own per-file verdicts (U13 fold F6.2/F6.7),
	// so a post-upload rejection composes the same count-aware sentence an
	// at-selection one does, rather than the section-4 fallback that used
	// to fire 100% of the time here (outcome.detail is only ever populated
	// on the pushed-back path).
	function composeAlertBody(files: FileVerdict[]): string {
		const failed = files.filter((f) => !f.OK);
		const labels = new Set(failed.map((f) => f.Reason.label));
		const n = failed.length;
		if (labels.size === 1) {
			const label = failed[0]?.Reason.label ?? '';
			return n === 1 ? `One track is ${label}, and albums go in whole.` : `${n} tracks are ${label}, and albums go in whole.`;
		}
		return `${n} tracks didn't pass the checks, and albums go in whole.`;
	}

	/** One reason hint per distinct failing reason code (Pass B conductor ruling 5), not one per flagged row. */
	function distinctReasons(files: FileVerdict[]): RejectReason[] {
		const seen = new Set<string>();
		const out: RejectReason[] = [];
		for (const f of files) {
			if (!f.OK && !seen.has(f.Reason.code)) {
				seen.add(f.Reason.code);
				out.push(f.Reason);
			}
		}
		return out;
	}

	/** passb-copy-final.md section 3's "Matched:" line, shown only for a main-lane confident match -- a private-lane promotion carries no matched metadata (internal/serve's own doc comment on sessionOutcome). */
	const matchedLine = $derived(
		viewMode === 'matched' && outcome?.matchedArtist
			? `Matched: ${outcome.matchedArtist} - ${outcome.matchedAlbum} (${outcome.matchedYear}). You can close this page now.`
			: null,
	);

	const inLibraryTitle = $derived(
		outcome?.matchedArtist ? `${outcome.matchedArtist} - ${outcome.matchedAlbum}` : (journey?.album ?? album),
	);

	// The two queued variants share one wire Detail signal
	// (internal/journey.QueuedDetail's own two literal strings), never
	// rendered verbatim -- OutcomeAlert owns the contract copy for each.
	const ownEarlierUpload = $derived(journey?.detail === 'waiting on your other upload');

	onMount(() => {
		void runFinish();
	});

	onDestroy(() => {
		void wakeLock.release();
	});

	// Drives the 5s journey poll off outcome/sessionId rather than a
	// variable runFinish's own finally assigns (U13 fold F6.10): Svelte
	// guarantees an effect's returned teardown runs on unmount even
	// though the effect itself started asynchronously, which a manual
	// onDestroy-reads-a-maybe-still-null-closure could not.
	$effect(() => {
		if (!outcome?.promoted) {
			return;
		}
		const stop = store.startJourneyPolling(sessionId);
		return () => stop();
	});

	async function runFinish(): Promise<void> {
		finishing = true;
		finishError = false;
		try {
			const result = await store.finishSession(sessionId, album);
			outcome = result;
		} catch (err) {
			// store.signedOut already carries the reload-to-sign-in state
			// (App.svelte swaps this component out entirely for that
			// state) -- anything else is a genuine failure to finish,
			// with a retry the user can act on.
			if (!(err instanceof SignedOutError)) {
				finishError = true;
			}
		} finally {
			finishing = false;
			announceForViewMode();
		}
	}

	function announceForViewMode(): void {
		switch (viewMode) {
			case 'matched':
				announcer.announceState(matchedLine ?? 'Adding it to the library now.');
				break;
			case 'pushed-back':
				announcer.announceState('No MusicBrainz match for this album.');
				break;
			case 'probe-unavailable':
				announcer.announceState("Can't check MusicBrainz right now.");
				break;
			case 'format-rejected':
				announcer.announceState("Couldn't add this album.");
				break;
			case 'finish-failed':
				announcer.announceState("Couldn't finish checking this album.");
				break;
			default:
				break;
		}
	}

	// journey.state changing (via the 5s poll) is announced the same way
	// a state change from runFinish is: a fresh, distinct message each
	// time, never a progress-style delta.
	let lastAnnouncedJourneyState = $state<string | null>(null);
	$effect(() => {
		if (!journey || journey.state === lastAnnouncedJourneyState) {
			return;
		}
		lastAnnouncedJourneyState = journey.state;
		if (journey.state === 'in-library') {
			announcer.announceCompletion(`${inLibraryTitle} is in your library.`);
		} else if (journey.state === 'stalled') {
			announcer.announceState('Taking longer than it should.');
		} else if (journey.state === 'queued') {
			announcer.announceState('Waiting its turn.');
		}
	});

	function retryFinish(): void {
		void runFinish();
	}
</script>

<div class="flex flex-col gap-4">
	{#if viewMode === 'in-library'}
		<div class="alert alert-success alert-soft motion-safe:transition-opacity" role="status">
			<div class="flex flex-col items-start gap-3">
				<h2 class="font-semibold">In your library</h2>
				<p>{inLibraryTitle} is ready.</p>
				<div class="flex flex-wrap gap-2">
					{#if journey?.libraryLink}
						<a class="btn btn-primary" href={journey.libraryLink} target="_blank" rel="noopener noreferrer">
							Open in Navidrome
						</a>
					{/if}
					<button type="button" class="btn btn-outline" onclick={onUploadAnother}>Upload another album</button>
				</div>
			</div>
		</div>
	{:else}
		<JourneySteps {totalFiles} {statuses} />

		{#if matchedLine}
			<p class="text-sm">{matchedLine}</p>
		{/if}

		{#if viewMode === 'finishing' || viewMode === 'matched'}
			<WakeLockBar {wakeLock} />
		{/if}

		{#if viewMode === 'pushed-back'}
			<OutcomeAlert outcome={{ kind: 'pushed-back', onUploadPrivate, onFixNames }} />
		{:else if viewMode === 'probe-unavailable'}
			<OutcomeAlert outcome={{ kind: 'probe-unavailable', onTryAgain: retryFinish }} />
		{:else if viewMode === 'queued'}
			<OutcomeAlert outcome={{ kind: 'queued', ownEarlierUpload }} />
		{:else if viewMode === 'stalled'}
			<OutcomeAlert outcome={{ kind: 'stalled' }} />
		{:else if viewMode === 'finish-failed'}
			<div class="alert alert-error alert-soft" role="alert">
				<div class="flex flex-col items-start gap-3">
					<h2 class="font-semibold">Couldn't finish checking</h2>
					<p>
						Something went wrong finishing the check. Your album is already uploaded and waiting; no need to send it
						again.
					</p>
					<button type="button" class="btn btn-primary" onclick={retryFinish}>Try again</button>
				</div>
			</div>
		{:else if viewMode === 'format-rejected'}
			{@const verdictFiles = outcome?.files ?? []}
			<div class="alert alert-error alert-soft" role="alert">
				<div class="flex flex-col items-start gap-2">
					<h2 class="font-semibold">Couldn't add this album</h2>
					<p>{verdictFiles.length > 0 ? composeAlertBody(verdictFiles) : "The album didn't pass its checks."}</p>
				</div>
			</div>
			{#if verdictFiles.length > 0}
				<ul class="border-base-300 divide-base-300 divide-y rounded border">
					{#each verdictFiles as file (file.Name)}
						<FileRow
							name={file.Name}
							status={file.OK ? 'ok' : 'failed'}
							progressBytes={0}
							ackedBytes={0}
							verdict={file.OK ? null : file}
							failureKind={file.OK ? null : 'rejected'}
							variant="verdict"
						/>
					{/each}
				</ul>
				{#each distinctReasons(verdictFiles) as reason (reason.code)}
					<p class="text-sm">{reason.friendly} <strong>Next:</strong> {reason.next_step}</p>
				{/each}
			{/if}
		{/if}
	{/if}
</div>
