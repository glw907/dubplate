<!--
  @component
  The full-viewport cue {@link DropZone} shows while a drag is over the
  page. Presentational only: `active` is the sole input, so a
  `DropZone` unit test can drive it directly without staging a real
  `DragEvent`.
-->
<script lang="ts">
	interface Props {
		active: boolean;
	}

	let { active }: Props = $props();
</script>

{#if active}
	<div
		class="bg-base-100/90 pointer-events-none fixed inset-0 z-50 flex items-center justify-center backdrop-blur-sm"
		role="presentation"
	>
		<div class="rounded-box border-primary bg-base-200 border-4 border-dashed p-12 text-center">
			<p class="text-primary text-xl font-semibold">Drop an album anywhere on this page</p>
		</div>
	</div>
{/if}
