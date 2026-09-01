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
  this component never owns the DOM region itself.
-->
<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import JourneySteps, { type StepStatus } from './JourneySteps.svelte';
	import OutcomeAlert from './OutcomeAlert.svelte';
	import WakeLockBar from './WakeLockBar.svelte';
	import { LiveRegion } from '../announce';
	import { uploadStore, type FinishOutcome, type UploadStore } from '../store/upload.svelte';
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
	}: Props = $props();

	let finishing = $state(true);
	let outcome = $state<FinishOutcome | null>(null);
	let stopPolling: (() => void) | null = null;

	const journey = $derived(store.journey);

	type ViewMode = 'finishing' | 'matched' | 'queued' | 'stalled' | 'in-library' | 'pushed-back' | 'probe-unavailable' | 'format-rejected';

	const viewMode = $derived.by((): ViewMode => {
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
		stopPolling?.();
		void wakeLock.release();
	});

	async function runFinish(): Promise<void> {
		finishing = true;
		try {
			const result = await store.finishSession(sessionId, album);
			outcome = result;
		} catch {
			// store.signedOut already carries the reload-to-sign-in state;
			// there is nothing further for this view to do.
		} finally {
			finishing = false;
			announceForViewMode();
			if (outcome?.promoted) {
				stopPolling = store.startJourneyPolling(sessionId);
			}
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
		<div class="alert alert-success flex-col items-start gap-3 motion-safe:transition-opacity" role="status">
			<h3 class="font-semibold">In your library</h3>
			<p>{inLibraryTitle} is ready.</p>
			{#if journey?.libraryLink}
				<a class="btn btn-primary" href={journey.libraryLink} target="_blank" rel="noopener noreferrer">
					Open in Navidrome
				</a>
			{/if}
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
		{:else if viewMode === 'format-rejected'}
			<div class="alert alert-error flex-col items-start" role="alert">
				<h3 class="font-semibold">Couldn't add this album</h3>
				<p>{outcome?.detail || "The album didn't pass its checks."}</p>
			</div>
		{/if}
	{/if}
</div>
