<!--
  @component
  The format-requirement panel (passb-copy-final.md section 1): visible
  on every selection surface before any picker opens (Geoff's
  format-prevention layer 1), never tucked behind a link. `variant`
  switches between the desktop and iPad/iPhone registers; `collectionMode`
  swaps the whole card for the private lane's contract panel (section 6)
  the instant that collection is chosen, including on desktop, since the
  contract change applies everywhere the picker does. `maxAlbumBytes`
  is the only numeric substitution the copy makes ({max_album_size});
  every other fact is fixed prose, not derived from config.
-->
<script lang="ts">
	import { formatBytes } from '../format';

	interface Props {
		variant: 'desktop' | 'mobile';
		collectionMode: 'standard' | 'private';
		maxAlbumBytes: number;
	}

	let { variant, collectionMode, maxAlbumBytes }: Props = $props();

	let detailSheet: HTMLDialogElement | undefined = $state();
</script>

{#if collectionMode === 'private'}
	<div class="card bg-base-200">
		<div class="card-body gap-2">
			<h2 class="card-title">Private Recordings</h2>
			<p class="text-base-content/70 text-sm">your tags, its own section</p>
			<p>
				No MusicBrainz matching here. Private Recordings is for music MusicBrainz doesn't know:
				this album keeps <strong>the tags you enter</strong> and goes into its own section in Navidrome,
				separate from the main library.
			</p>
		</div>
	</div>
{:else if variant === 'desktop'}
	<div class="card bg-base-200">
		<div class="card-body gap-2">
			<p>
				This library is <strong>lossless FLAC</strong>, CD quality or better. Upload a whole album: a
				.zip or its files, up to <strong>{formatBytes(maxAlbumBytes)}</strong>. Cover art is optional.
				Include one if you have it; the library finds one otherwise.
			</p>
		</div>
	</div>
{:else}
	<div class="card bg-base-200">
		<div class="card-body gap-2">
			<p>
				From a phone or iPad, a <strong>.zip</strong> of the album is easiest. Picking the files works
				too. Cover art is optional; the library finds one if you don't include it.
			</p>
			<div class="card-actions">
				<button type="button" class="btn btn-ghost btn-sm" onclick={() => detailSheet?.showModal()}>
					What can I upload?
				</button>
			</div>
		</div>
	</div>

	<dialog bind:this={detailSheet} class="modal modal-bottom sm:modal-middle">
		<div class="modal-box">
			<h3 class="text-lg font-semibold">What can I upload?</h3>
			<ul class="flex flex-col gap-3 py-4">
				<li>Lossless FLAC, CD quality or better. MP3, AAC, and other lossy files can't go in.</li>
				<li>
					Whole albums only, up to {formatBytes(maxAlbumBytes)}. Send a .zip of the album, or pick the
					files themselves.
				</li>
				<li>
					Cover art is optional. If the album comes with art, it's used. If not, the library finds a
					cover on its own.
				</li>
				<li>The checks only look at audio files. Booklets and other extras ride along.</li>
			</ul>
			<div class="modal-action">
				<form method="dialog">
					<button type="submit" class="btn">Close</button>
				</form>
			</div>
		</div>
		<form method="dialog" class="modal-backdrop">
			<button type="submit">close</button>
		</form>
	</dialog>
{/if}
