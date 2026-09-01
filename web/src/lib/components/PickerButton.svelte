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
-->
<script lang="ts">
	interface Props {
		label: string;
		/** Extension-first accept list (gate 3: bare `audio/*` greys out FLAC files on iOS Safari, WebKit bug 242110). */
		accept: string;
		multiple?: boolean;
		webkitdirectory?: boolean;
		size?: 'default' | 'lg';
		emphasis?: 'primary' | 'outline';
		onFiles: (files: File[]) => void;
	}

	let {
		label,
		accept,
		multiple = true,
		webkitdirectory = false,
		size = 'lg',
		emphasis = 'primary',
		onFiles,
	}: Props = $props();

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
	class="btn {emphasis === 'primary' ? 'btn-primary' : 'btn-outline'} {size === 'lg'
		? 'btn-lg'
		: ''} focus-within:outline-primary relative w-full focus-within:outline-2 focus-within:outline-offset-2 sm:w-auto"
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
