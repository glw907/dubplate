<!--
  @component
  The four-step DaisyUI spine (passb-copy-final.md section 3; Pass B
  conductor ruling 4: no fifth step -- a confident match completes step
  3 and moves step 4 to current with "adding it now", and the frozen
  spine's outcome renders as {@link OutcomeAlert} beside it, never a
  fifth step). Purely a renderer: `statuses` is the one thing this
  component reads to decide color and glyph, so every branch (queued,
  pushed back, probe unavailable, stalled) is Journey.svelte's mapping
  to make, not this file's. `steps-vertical sm:steps-horizontal` keeps
  an iPad-portrait viewport from ever overflowing the spine
  horizontally (U11 criterion).
-->
<script lang="ts" module>
	/** One step's render state; `error` freezes the spine on a branch outcome (pushed back / probe unavailable) without inventing a fifth step. */
	export type StepStatus = 'complete' | 'current' | 'upcoming' | 'error';
</script>

<script lang="ts">
	interface Props {
		/** How many tracks the album carries -- step 1's `{n}` token. */
		totalFiles: number;
		/** One status per step, in spine order. */
		statuses: [StepStatus, StepStatus, StepStatus, StepStatus];
	}

	let { totalFiles, statuses }: Props = $props();

	interface StepDef {
		label: string;
		sublabel: string;
	}

	const stepDefs = $derived.by((): StepDef[] => [
		{ label: 'Uploaded', sublabel: `all ${totalFiles} track${totalFiles === 1 ? '' : 's'} received` },
		{ label: 'Checked', sublabel: 'every file is CD-quality FLAC' },
		{ label: 'Matching', sublabel: 'checking MusicBrainz, just a moment' },
		{ label: 'In your library', sublabel: '' }, // step 4's sublabel depends on its own status; see fourthSublabel below
	]);

	const fourthSublabel = $derived(statuses[3] === 'current' ? 'adding it now' : 'ready to play');

	function colorClass(status: StepStatus): string {
		switch (status) {
			case 'complete':
				return 'step-success';
			case 'current':
				return 'step-primary';
			case 'error':
				return 'step-error';
			default:
				return '';
		}
	}

	function glyph(status: StepStatus): string | undefined {
		if (status === 'complete') {
			return '✓';
		}
		if (status === 'error') {
			return '!';
		}
		return undefined;
	}
</script>

<ul class="steps steps-vertical sm:steps-horizontal w-full" aria-label="Upload progress">
	{#each stepDefs as def, i (def.label)}
		{@const status = statuses[i]}
		{@const dataContent = glyph(status)}
		<li
			class="step {colorClass(status)}"
			data-content={dataContent}
			aria-current={status === 'current' ? 'step' : undefined}
		>
			<span class="flex flex-col items-center gap-0.5 py-1 text-left sm:items-center sm:text-center">
				<span class="font-semibold {status === 'current' ? 'motion-safe:animate-pulse' : ''}">{def.label}</span>
				<span class="text-base-content/60 text-xs">{i === 3 ? fourthSublabel : def.sublabel}</span>
			</span>
		</li>
	{/each}
</ul>
