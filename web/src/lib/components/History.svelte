<!--
  @component
  "Your uploads" (passb-copy-final.md section 9): every session GET
  /api/history returns, in-flight resumables first (U9's `HistoryRow`
  already carries that ordering and the `resumable` flag), each rendered
  as one status chip. A `pushed-back`/`probe-unavailable` row is a
  disclosure: opening it renders {@link OutcomeAlert}'s section-5 copy
  and its actions, the same "no need to hunt down the old alert" path a
  fresh push-back gets. `review` has no contract copy of its own -- the
  custody-policy decision retires that state from this surface entirely
  for a real uploader session (a pushed-back album never reaches the
  ledger; ledger `review` only exists for the importer's admin-timer
  orphans going forward) -- so this file gives it one plain, undramatic
  label rather than inventing prose the contract never ratified.
  `probe-unavailable`'s "Try again" is this file's own concern, not a
  reroute through `onFixNames`: the copy above it promises the album is
  already uploaded and waiting, so retrying means re-issuing the same
  finish probe (`store.finishSession`), never sending the caller back to
  the idle picker. The title only renders as a `<button>` for a
  pushed-back/probe-unavailable row, with `aria-expanded`/`aria-controls`
  pointing at the revealed detail div; every other row's title is a plain
  `<span>` (U13 fold F6.12/F6.15) -- a `disabled` button here used to
  read to a screen reader as an unavailable control over the row's own
  primary content, and the enabled case carried no expanded/collapsed
  state at all.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import OutcomeAlert from './OutcomeAlert.svelte';
	import { uploadStore, type HistoryRow, type HistoryState, type UploadStore } from '../store/upload.svelte';

	interface Props {
		store?: UploadStore;
		onResume?: (row: HistoryRow) => void;
		onUploadPrivate?: (row: HistoryRow) => void;
		onFixNames?: (row: HistoryRow) => void;
	}

	let { store = uploadStore, onResume = () => {}, onUploadPrivate = () => {}, onFixNames = () => {} }: Props = $props();

	// U13 fold F5.14: a genuine loadHistory failure (never a signed-out
	// one -- the store already handles that by setting store.signedOut)
	// used to be an unhandled rejection that left "Nothing here yet."
	// rendered, actively lying about the caller's uploads.
	let loadFailed = $state(false);

	onMount(() => {
		void store.loadHistory().catch(() => {
			loadFailed = true;
		});
	});

	// Set to the row whose retry just failed (U13 fold F5.14): "Try
	// again" otherwise silently does nothing on a rejection, with no
	// feedback that the attempt even happened.
	let retryFailedSessionId = $state<string | null>(null);

	/** Re-probes row's session outcome in place and refreshes this list; never touches `onFixNames`, which stays scoped to the pushed-back re-pick path. */
	async function retryProbe(row: HistoryRow): Promise<void> {
		retryFailedSessionId = null;
		try {
			await store.finishSession(row.sessionId, row.album);
			await store.loadHistory();
		} catch {
			retryFailedSessionId = row.sessionId;
		}
	}

	let expandedSessionId = $state<string | null>(null);

	function toggleExpanded(row: HistoryRow): void {
		expandedSessionId = expandedSessionId === row.sessionId ? null : row.sessionId;
	}

	function isPushedBack(state: HistoryState): boolean {
		return state === 'pushed-back' || state === 'probe-unavailable';
	}

	const chipText: Record<HistoryState, string> = {
		uploading: 'uploading',
		uploaded: 'checking',
		checking: 'checking',
		tagging: 'adding now',
		'in-library': 'in library',
		review: 'needs a look',
		rejected: 'wrong format',
		queued: 'waiting its turn',
		stalled: 'slow going',
		'pushed-back': 'pushed back',
		'probe-unavailable': 'pushed back',
	};

	function chipForRow(row: HistoryRow): string {
		const base = chipText[row.state];
		return row.state === 'uploading' && row.resumable ? `${base} · resume` : base;
	}

	function chipClass(state: HistoryState): string {
		switch (state) {
			case 'in-library':
				return 'badge-soft badge-success';
			case 'rejected':
				return 'badge-soft badge-error';
			case 'queued':
			case 'stalled':
			case 'pushed-back':
			case 'probe-unavailable':
			case 'review':
				return 'badge-soft badge-warning';
			default:
				return 'badge-ghost';
		}
	}
</script>

<section class="flex flex-col gap-3">
	<h2 class="text-lg font-semibold">Your uploads</h2>

	{#if loadFailed}
		<div class="alert alert-error alert-soft" role="alert">
			<div class="flex flex-col items-start gap-2">
				<h3 class="text-base-content font-semibold">Couldn't load your uploads</h3>
				<p class="text-base-content/80">Something went wrong checking your uploads. Reload to try again.</p>
				<button type="button" class="btn min-h-11" onclick={() => location.reload()}>Reload</button>
			</div>
		</div>
	{:else if store.history.length === 0}
		<p class="text-base-content/60 text-sm">Nothing here yet.</p>
	{:else}
		<ul class="border-base-300 divide-base-300 divide-y rounded border">
			{#each store.history as row (row.sessionId)}
				<li class="flex flex-col gap-2 p-3">
					<div class="flex flex-wrap items-center justify-between gap-2">
						{#if isPushedBack(row.state)}
							<button
								type="button"
								class="link min-h-11 flex min-w-0 flex-1 items-center truncate text-left"
								onclick={() => toggleExpanded(row)}
								aria-expanded={expandedSessionId === row.sessionId}
								aria-controls={`history-detail-${row.sessionId}`}
							>
								{row.album || 'Untitled album'}
							</button>
						{:else}
							<span class="min-w-0 flex-1 truncate">{row.album || 'Untitled album'}</span>
						{/if}
						<span class="badge {chipClass(row.state)} shrink-0">{chipForRow(row)}</span>
						{#if row.resumable}
							<button type="button" class="btn btn-sm min-h-11" onclick={() => onResume(row)}>Resume</button>
						{/if}
						{#if row.state === 'in-library' && row.libraryLink}
							<a class="btn btn-sm min-h-11" href={row.libraryLink} target="_blank" rel="noopener noreferrer">
								Open in Navidrome
							</a>
						{/if}
					</div>

					{#if isPushedBack(row.state) && expandedSessionId === row.sessionId}
						<div id={`history-detail-${row.sessionId}`}>
							{#if row.state === 'pushed-back'}
								<OutcomeAlert
									outcome={{
										kind: 'pushed-back',
										onUploadPrivate: () => onUploadPrivate(row),
										onFixNames: () => onFixNames(row),
									}}
								/>
							{:else}
								<OutcomeAlert outcome={{ kind: 'probe-unavailable', onTryAgain: () => void retryProbe(row) }} />
								{#if retryFailedSessionId === row.sessionId}
									<p class="text-error text-sm" role="alert">That didn't go through. Check your connection and try again.</p>
								{/if}
							{/if}
						</div>
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</section>
