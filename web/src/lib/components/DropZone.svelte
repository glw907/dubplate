<!--
  @component
  Desktop whole-page drop and paste target (the plan's shadcn-svelte-
  extras Root/Trigger/DragOverlay decomposition, restyled to stock
  DaisyUI). Listens on `window`, not on its own bounding box, so a drop
  anywhere on the page reaches `onFiles` -- drag-and-drop is
  enhancement, never required, so this never blocks its slotted picker
  button from rendering on touch platforms where these events simply
  never fire. Renders {@link DropOverlay} on top of everything while a
  drag is in progress.
-->
<script lang="ts">
	import type { Snippet } from 'svelte';
	import DropOverlay from './DropOverlay.svelte';

	interface Props {
		/** Called with every File from a drop or a clipboard paste; never fetch or tus, just the raw selection. */
		onFiles: (files: File[]) => void;
		children: Snippet;
	}

	let { onFiles, children }: Props = $props();

	// A drop target nested inside this one still fires dragenter/dragleave
	// on window as the pointer crosses child element boundaries; a bare
	// boolean would flicker false on every such crossing, so a depth
	// counter is what makes "still dragging" survive traversing children.
	let dragDepth = $state(0);

	function carriesFiles(transfer: DataTransfer | null): boolean {
		return transfer !== null && Array.from(transfer.types).includes('Files');
	}

	function handleDragEnter(event: DragEvent): void {
		if (!carriesFiles(event.dataTransfer)) {
			return;
		}
		event.preventDefault();
		dragDepth += 1;
	}

	function handleDragOver(event: DragEvent): void {
		// preventDefault is what tells the browser this drop target
		// accepts a drop at all; without it, 'drop' never fires.
		if (carriesFiles(event.dataTransfer)) {
			event.preventDefault();
		}
	}

	function handleDragLeave(event: DragEvent): void {
		if (!carriesFiles(event.dataTransfer)) {
			return;
		}
		dragDepth = Math.max(0, dragDepth - 1);
	}

	function handleDrop(event: DragEvent): void {
		if (!carriesFiles(event.dataTransfer)) {
			return;
		}
		event.preventDefault();
		dragDepth = 0;
		const files = Array.from(event.dataTransfer?.files ?? []);
		if (files.length > 0) {
			onFiles(files);
		}
	}

	function handlePaste(event: ClipboardEvent): void {
		const items = event.clipboardData?.items;
		if (!items) {
			return;
		}
		const files: File[] = [];
		for (let i = 0; i < items.length; i += 1) {
			const item = items[i];
			if (item.kind === 'file') {
				const file = item.getAsFile();
				if (file) {
					files.push(file);
				}
			}
		}
		if (files.length > 0) {
			onFiles(files);
		}
	}

	$effect(() => {
		window.addEventListener('dragenter', handleDragEnter);
		window.addEventListener('dragover', handleDragOver);
		window.addEventListener('dragleave', handleDragLeave);
		window.addEventListener('drop', handleDrop);
		window.addEventListener('paste', handlePaste);
		return () => {
			window.removeEventListener('dragenter', handleDragEnter);
			window.removeEventListener('dragover', handleDragOver);
			window.removeEventListener('dragleave', handleDragLeave);
			window.removeEventListener('drop', handleDrop);
			window.removeEventListener('paste', handlePaste);
		};
	});
</script>

<div class="relative">
	<div class="mb-4 hidden flex-col items-center gap-2 md:flex">
		<p class="text-base-content/70 text-sm">Drop an album anywhere on this page</p>
		<div class="divider w-full max-w-xs self-center">or</div>
	</div>
	{@render children()}
</div>

<DropOverlay active={dragDepth > 0} />
