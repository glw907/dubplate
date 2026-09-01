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

	onMount(() => {
		void store.loadHistory();
	});

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
				return 'badge-success';
			case 'rejected':
				return 'badge-error';
			case 'queued':
			case 'stalled':
			case 'pushed-back':
			case 'probe-unavailable':
			case 'review':
				return 'badge-warning';
			default:
				return 'badge-ghost';
		}
	}
</script>

<section class="flex flex-col gap-3">
	<h2 class="text-lg font-semibold">Your uploads</h2>

	{#if store.history.length === 0}
		<p class="text-base-content/60 text-sm">Nothing here yet.</p>
	{:else}
		<ul class="border-base-300 divide-base-300 divide-y rounded border">
			{#each store.history as row (row.sessionId)}
				<li class="flex flex-col gap-2 p-3">
					<div class="flex flex-wrap items-center justify-between gap-2">
						<button
							type="button"
							class="min-w-0 flex-1 truncate text-left {isPushedBack(row.state) ? 'link' : ''}"
							onclick={() => isPushedBack(row.state) && toggleExpanded(row)}
							disabled={!isPushedBack(row.state)}
						>
							{row.album || 'Untitled album'}
						</button>
						<span class="badge {chipClass(row.state)} shrink-0">{chipForRow(row)}</span>
						{#if row.resumable}
							<button type="button" class="btn btn-sm" onclick={() => onResume(row)}>Resume</button>
						{/if}
						{#if row.state === 'in-library' && row.libraryLink}
							<a class="btn btn-sm" href={row.libraryLink} target="_blank" rel="noopener noreferrer">
								Open in Navidrome
							</a>
						{/if}
					</div>

					{#if isPushedBack(row.state) && expandedSessionId === row.sessionId}
						{#if row.state === 'pushed-back'}
							<OutcomeAlert
								outcome={{
									kind: 'pushed-back',
									onUploadPrivate: () => onUploadPrivate(row),
									onFixNames: () => onFixNames(row),
								}}
							/>
						{:else}
							<OutcomeAlert outcome={{ kind: 'probe-unavailable', onTryAgain: () => onFixNames(row) }} />
						{/if}
					{/if}
				</li>
			{/each}
		</ul>
	{/if}
</section>
