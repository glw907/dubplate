<!--
  @component
  The three branch outcomes (passb-copy-final.md sections 5, 7, 8) as
  one DaisyUI alert beside the frozen {@link JourneySteps} spine --
  never a fifth step (Pass B conductor ruling 4). `outcome`'s `kind`
  picks the copy verbatim from the contract: `pushed-back` is section
  5's no-MusicBrainz-match state (with `probe-unavailable` as its own
  distinct sub-kind, never folded into the same heading); `queued` is
  section 7, its `ownEarlierUpload` flag choosing between the two
  variants internal/journey.QueuedDetail's Detail string distinguishes
  server-side (the wire Detail itself is an internal phrase, not
  contract copy -- Journey.svelte maps it to this flag rather than
  rendering it); `stalled` is section 8's fixed copy. "Add it to
  MusicBrainz" links out to MusicBrainz's own add-release flow rather
  than calling back into the app, since there is nothing this session
  can do that page's form doesn't already do better. Every branch
  renders through {@link AlertCard}, which owns the DaisyUI grid
  mechanic each of these alerts otherwise has to reproduce by hand.
-->
<script lang="ts">
	import AlertCard from './AlertCard.svelte';

	type Outcome =
		| { kind: 'pushed-back'; onUploadPrivate: () => void; onFixNames: () => void }
		| { kind: 'probe-unavailable'; onTryAgain: () => void }
		| { kind: 'queued'; ownEarlierUpload: boolean }
		| { kind: 'stalled' };

	interface Props {
		outcome: Outcome;
	}

	let { outcome }: Props = $props();
</script>

{#if outcome.kind === 'pushed-back'}
	<AlertCard tone="warning" title="No MusicBrainz match" spacing="roomy">
		<p class="text-base-content/80">
			MusicBrainz didn't find a match for this album, so it didn't go into the main library. You can
			upload it to <strong>Private Recordings</strong> instead. That collection is for music
			MusicBrainz doesn't know, and it shows up in Navidrome as its own section.
		</p>
		<p class="text-base-content/80">
			If it's a released album, a fix usually works: check the artist and title spelling and try
			again, or add the release to MusicBrainz first.
		</p>
		<div class="flex flex-wrap gap-2">
			<button type="button" class="btn btn-primary" onclick={outcome.onUploadPrivate}>
				Upload to Private Recordings
			</button>
			<button type="button" class="btn btn-outline" onclick={outcome.onFixNames}>
				Fix the names and retry
			</button>
			<a class="btn btn-outline" href="https://musicbrainz.org/release/add" target="_blank" rel="noopener noreferrer">
				Add it to MusicBrainz
			</a>
		</div>
	</AlertCard>
{:else if outcome.kind === 'probe-unavailable'}
	<AlertCard tone="warning" title="Can't check MusicBrainz right now" spacing="roomy">
		<p class="text-base-content/80">
			The matching service isn't answering. Your album is already uploaded and waiting; no need to
			send it again. Try again in a little while.
		</p>
		<button type="button" class="btn btn-primary" onclick={outcome.onTryAgain}>Try again</button>
	</AlertCard>
{:else if outcome.kind === 'queued'}
	<AlertCard tone="info" title="Waiting its turn" role="status">
		<p class="text-base-content/80">
			{outcome.ownEarlierUpload
				? 'Your earlier album is still importing. This one starts as soon as it finishes.'
				: 'The library is finishing another import. This one starts right after.'}
		</p>
	</AlertCard>
{:else}
	<AlertCard tone="warning" title="Taking longer than it should" role="status">
		<p class="text-base-content/80">
			Something behind the scenes is slow. The admin gets notified automatically when this happens,
			and this page keeps checking on its own. Nothing needs to be redone.
		</p>
	</AlertCard>
{/if}
