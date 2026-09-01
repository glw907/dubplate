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
	 * the content check, so it has no taxonomy label to show; the
	 * batch-level "Upload stopped" state (passb-copy-final.md section 2)
	 * carries that message, and this row just reads as still-in-flight
	 * until retried.
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
		return (failureKind === 'rejected' ? verdict?.Reason.label : null) ?? 'waiting';
	}

	const chipText = $derived(computeChipText());

	const chipClass = $derived(
		status === 'ok' ? 'badge-success' : status === 'failed' && failureKind === 'rejected' ? 'badge-error' : 'badge-ghost',
	);
</script>

<div class="border-base-300 flex flex-col gap-1 border-b p-3 sm:flex-row sm:items-center sm:justify-between sm:gap-4">
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
</div>
