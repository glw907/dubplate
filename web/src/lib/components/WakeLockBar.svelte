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
		<button type="button" class="btn btn-outline btn-sm" aria-pressed={wakeLock.on} onclick={() => wakeLock.toggle()}>
			Keep my screen awake
		</button>
	{/if}
</div>
