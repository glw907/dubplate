<!--
  @component
  One file's selection/upload row (passb-copy-final.md sections 2 and
  4): filename, size, a status chip, and -- while transferring -- the
  honest acked-vs-transferred progress readout. `variant` picks the
  chip's passed-word register: `'progress'` for the live per-file
  streaming verdicts section 2 describes ("looks good"), `'verdict'`
  for a finished batch's summary list section 4 describes ("fine").
  A failed row always shows its own verdict's `{label}` rather than a
  generic word -- the taxonomy strings are the API's, never a client
  table (U4's reject wire format). Display-only: no fetch, no tus,
  no retry action of its own -- retry is a batch-level control, since
  the plan's honest-retry copy differs by upload path (multi-file
  resends only failing files; a ZIP resends the archive), a decision
  this row has no path context to make.

  Root is `<li>`, always rendered inside a caller's own `<ul>` (U13
  fold F5.11): a `<div>` root here made every caller's list an invalid
  `<ul>`/non-`<li>` nesting that a screen reader repairs by dropping
  the list semantics entirely. Separators are the parent `<ul>`'s own
  `divide-y`, not this row's -- a border here would double every rule.
-->
<script lang="ts">
	import { formatBytes } from '../format';
	import type { FailureKind, FileStatus, FileVerdict } from '../store/upload.svelte';

	interface Props {
		name: string;
		size: number;
		status: FileStatus;
		progressBytes: number;
		ackedBytes: number;
		verdict: FileVerdict | null;
		failureKind: FailureKind;
		variant: 'progress' | 'verdict';
	}

	let { name, size, status, progressBytes, ackedBytes, verdict, failureKind, variant }: Props = $props();

	/**
	 * A transport failure (`failureKind` `'connection'`) never reached
	 * the content check, so it has no taxonomy label to show; it chips
	 * "stopped" instead (passb-copy-final.md section 2, ratified
	 * post-build 2026-09-01) rather than "waiting", which the contract
	 * reserves for a file that has not started -- a failed row must
	 * never read as not-yet-started.
	 */
	function computeChipText(): string {
		if (status === 'pending') {
			return 'waiting';
		}
		if (status === 'uploading') {
			return 'sending';
		}
		if (status === 'ok') {
			return variant === 'progress' ? 'looks good' : 'fine';
		}
		return (failureKind === 'rejected' ? verdict?.Reason.label : null) ?? 'stopped';
	}

	const chipText = $derived(computeChipText());

	const chipClass = $derived(
		status === 'ok' ? 'badge-success' : status === 'failed' && failureKind === 'rejected' ? 'badge-error' : 'badge-ghost',
	);
</script>

<li class="flex flex-col gap-1 p-3 md:flex-row md:items-center md:justify-between md:gap-4">
	<div class="min-w-0 flex-1">
		<p class="truncate">{name}</p>
		<p class="text-base-content/60 text-xs">{formatBytes(size)}</p>
		{#if status === 'uploading'}
			<p class="text-base-content/60 text-xs">
				{formatBytes(ackedBytes)} confirmed <span class="opacity-70">of {formatBytes(size)}</span>
			</p>
			<progress class="progress progress-primary w-full" value={progressBytes} max={size}></progress>
		{/if}
	</div>
	<span class="badge {chipClass} shrink-0">{chipText}</span>
</li>
