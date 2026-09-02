<!--
  @component
  The large native-picker button that is the primary selection control
  on every platform (drag-drop is enhancement; this is not). A real
  `<input type="file">` sits inside the label, visually hidden but
  still focusable and clickable -- `display:none` would drop it from
  the tab order entirely, so hiding is via `opacity-0` positioned to
  cover the label, with `focus-within` carrying the visible focus ring
  onto the label itself. Fires `onFiles` with the picked files and
  never calls fetch or tus; the caller decides what happens next
  (pre-flight screening, then the store).

  Always full width (U13 fold F5.1): the ratified mockup's idle screen
  never shows this button snapping to a content-sized chip at a wider
  viewport, and the caller constrains the column's own max width
  instead. Always `lg`-sized and uppercase to match Navidrome's own
  Material button convention (app.css's design-language note) -- every
  picker on the idle screen, primary or secondary, reads at the same
  weight rather than one at 48px and the other at 40px.
-->
<script lang="ts">
	interface Props {
		label: string;
		/** Extension-first accept list (gate 3: bare `audio/*` greys out FLAC files on iOS Safari, WebKit bug 242110). */
		accept: string;
		multiple?: boolean;
		webkitdirectory?: boolean;
		emphasis?: 'primary' | 'outline';
		onFiles: (files: File[]) => void;
	}

	let { label, accept, multiple = true, webkitdirectory = false, emphasis = 'primary', onFiles }: Props = $props();

	const emphasisClass = $derived(emphasis === 'primary' ? 'btn-primary' : 'btn-outline');

	let inputEl: HTMLInputElement | undefined = $state();

	function handleChange(): void {
		const files = Array.from(inputEl?.files ?? []);
		// Reset so picking the same file again still fires a change event.
		if (inputEl) {
			inputEl.value = '';
		}
		if (files.length > 0) {
			onFiles(files);
		}
	}
</script>

<label
	class="btn btn-lg {emphasisClass} focus-within:outline-primary relative w-full uppercase tracking-wide focus-within:outline-2 focus-within:outline-offset-2"
>
	<input
		bind:this={inputEl}
		type="file"
		class="absolute inset-0 h-full w-full cursor-pointer opacity-0"
		{accept}
		{multiple}
		{webkitdirectory}
		onchange={handleChange}
	/>
	{label}
</label>
