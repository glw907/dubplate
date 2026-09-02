<!--
  @component
  The one alert shell every state message in this app renders through:
  a DaisyUI soft alert carrying a heading and whatever the caller
  slots in below it. It exists because DaisyUI 5's `.alert` is
  `display:grid; grid-auto-flow:column`, which lays each direct child
  out as its own column the moment there are two or more (U13 fold
  F6.1) -- `flex-col items-start` on the alert itself does nothing,
  since those are flexbox properties applied to a grid container. The
  single content wrapper below keeps the alert's own grid to one column
  and lets normal flex layout run inside it. Reproducing that shape by
  hand at each call site is what let eight of them drift apart over one
  fix round; a caller now names a tone and a title and gets the rest.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';

	/** Which DaisyUI semantic color the alert wears (app.css defines all four). */
	type Tone = 'error' | 'warning' | 'info' | 'success';

	interface Props {
		tone: Tone;
		title: string;
		/**
		 * `'alert'` for a message the reader has to act on, `'status'`
		 * for progress they only have to hear about -- an assertive
		 * interruption is wrong for the latter.
		 */
		role?: 'alert' | 'status';
		/** `'roomy'` for an alert whose content ends in an actions row, which needs more separation from the copy above it than paragraphs need from each other. */
		spacing?: 'tight' | 'roomy';
		/** Set to 3 where the alert sits under a section heading of its own, so the document outline never skips a level. */
		headingLevel?: 2 | 3;
		/** Extra classes for the alert element itself; the tone, softness, and layout classes are this component's own. */
		class?: string;
		children: Snippet;
	}

	let {
		tone,
		title,
		role = 'alert',
		spacing = 'tight',
		headingLevel = 2,
		class: extraClass = '',
		children,
	}: Props = $props();

	// Spelled out rather than interpolated: Tailwind's scanner only ever
	// sees class names that appear literally in the source.
	const toneClass: Record<Tone, string> = {
		error: 'alert-error',
		warning: 'alert-warning',
		info: 'alert-info',
		success: 'alert-success',
	};
</script>

<div class="alert {toneClass[tone]} alert-soft {extraClass}" {role}>
	<div class="flex flex-col items-start {spacing === 'roomy' ? 'gap-3' : 'gap-2'}">
		<svelte:element this={headingLevel === 3 ? 'h3' : 'h2'} class="text-base-content font-semibold">
			{title}
		</svelte:element>
		{@render children()}
	</div>
</div>
