<!--
  @component
  The backgrounding note and the "Keep my screen awake" toggle
  (passb-copy-final.md section 2), rendered in both places a transfer is
  visible: App.svelte's in-flight branch and {@link Journey}'s finishing
  and matched views. Both pass the app's one {@link WakeLock} instance,
  so the toggle reads the same state on either side of the hand-off and
  the copy has a single home rather than two that can drift. The toggle
  disappears entirely on a browser without the Screen Wake Lock API;
  the note stands on its own, since backgrounding still pauses there.
  The toggle's class swaps with `wakeLock.on` (U13 fold F6.3): before
  this fix `aria-pressed` was the only thing that ever changed, so a
  sighted user pressing it got no visual confirmation either way.
-->
<script lang="ts">
	import type { WakeLock } from '../wakeLock.svelte';

	interface Props {
		wakeLock: WakeLock;
	}

	let { wakeLock }: Props = $props();
</script>

<div class="flex flex-wrap items-center justify-between gap-2">
	<p class="text-base-content/60 text-sm">
		Switching away pauses the upload; coming back resumes it. If the page closes, pick it up from Your
		uploads.
	</p>
	{#if wakeLock.supported}
		<button
			type="button"
			class="btn btn-sm min-h-11 {wakeLock.on ? 'btn-primary' : 'btn-outline'}"
			aria-pressed={wakeLock.on}
			onclick={() => wakeLock.toggle()}
		>
			Keep my screen awake
		</button>
	{/if}
</div>
