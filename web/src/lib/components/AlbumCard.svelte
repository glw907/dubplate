<!--
  @component
  The Private Recordings tag form (passb-copy-final.md section 6 /
  the spec's journey addendum): artist, album, year, and one editable
  title per track, prefilled from each file's name. Renders before
  upload can start for the private lane -- the caller gates its own
  "Upload to Private Recordings" action on this form rather than this
  component calling anything itself; display-only over whatever state
  the caller threads through as props.
-->
<script lang="ts" module>
	/**
	 * Strips a track's extension and a leading "01 - "/"01."/"01_"-style
	 * number prefix, leaving a reasonable starting title -- editable, not
	 * authoritative, since Private Recordings has no MusicBrainz pass to
	 * correct it.
	 */
	export function deriveTitleFromFilename(filename: string): string {
		const dot = filename.lastIndexOf('.');
		const stem = dot > 0 ? filename.slice(0, dot) : filename;
		return stem.replace(/^\s*\d{1,3}[\s._-]+/, '').trim() || stem;
	}

	/** The "{nn} (from filename)" row label, one-based and zero-padded per the copy contract. */
	export function trackRowLabel(index: number): string {
		return `${String(index + 1).padStart(2, '0')} (from filename)`;
	}
</script>

<script lang="ts">
	interface Props {
		trackFilenames: string[];
		trackTitles: string[];
		artist: string;
		album: string;
		year: string;
		onArtistInput: (value: string) => void;
		onAlbumInput: (value: string) => void;
		onYearInput: (value: string) => void;
		onTrackTitleInput: (index: number, value: string) => void;
		onSubmit: () => void;
		submitDisabled?: boolean;
	}

	let {
		trackFilenames,
		trackTitles,
		artist,
		album,
		year,
		onArtistInput,
		onAlbumInput,
		onYearInput,
		onTrackTitleInput,
		onSubmit,
		submitDisabled = false,
	}: Props = $props();
</script>

<div class="card bg-base-200">
	<div class="card-body gap-4">
		<h2 class="card-title">Tag this album</h2>

		<div class="fieldset">
			<label class="fieldset-label" for="album-artist">Artist</label>
			<input
				id="album-artist"
				class="input min-h-11 w-full"
				type="text"
				value={artist}
				oninput={(event) => onArtistInput(event.currentTarget.value)}
			/>
		</div>

		<div class="fieldset">
			<label class="fieldset-label" for="album-title">Album</label>
			<input
				id="album-title"
				class="input min-h-11 w-full"
				type="text"
				value={album}
				oninput={(event) => onAlbumInput(event.currentTarget.value)}
			/>
		</div>

		<div class="fieldset">
			<label class="fieldset-label" for="album-year">Year</label>
			<input
				id="album-year"
				class="input min-h-11 w-full"
				type="text"
				inputmode="numeric"
				value={year}
				oninput={(event) => onYearInput(event.currentTarget.value)}
			/>
		</div>

		<ul class="flex flex-col gap-3">
			{#each trackFilenames as filename, index (`${index}-${filename}`)}
				<li class="fieldset">
					<label class="fieldset-label" for={`track-title-${index}`}>{trackRowLabel(index)}</label>
					<input
						id={`track-title-${index}`}
						class="input min-h-11 w-full"
						type="text"
						value={trackTitles[index] ?? ''}
						oninput={(event) => onTrackTitleInput(index, event.currentTarget.value)}
					/>
				</li>
			{/each}
		</ul>

		<div class="card-actions">
			<button type="button" class="btn btn-primary" disabled={submitDisabled} onclick={onSubmit}>
				Upload to Private Recordings
			</button>
		</div>
	</div>
</div>
