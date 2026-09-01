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
  can do that page's form doesn't already do better.
-->
<script lang="ts">
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
	<div class="alert alert-warning flex-col items-start gap-3" role="alert">
		<h3 class="font-semibold">No MusicBrainz match</h3>
		<p>
			MusicBrainz didn't find a match for this album, so it didn't go into the main library. You can
			upload it to <strong>Private Recordings</strong> instead. That collection is for music
			MusicBrainz doesn't know, and it shows up in Navidrome as its own section.
		</p>
		<p>
			If it's a released album, a fix usually works: check the artist and title spelling and try
			again, or add the release to MusicBrainz first.
		</p>
		<div class="flex flex-wrap gap-2">
			<button type="button" class="btn btn-primary" onclick={outcome.onUploadPrivate}>
				Upload to Private Recordings
			</button>
			<button type="button" class="btn btn-ghost" onclick={outcome.onFixNames}> Fix the names and retry </button>
			<a class="btn btn-ghost" href="https://musicbrainz.org/release/add" target="_blank" rel="noopener noreferrer">
				Add it to MusicBrainz
			</a>
		</div>
	</div>
{:else if outcome.kind === 'probe-unavailable'}
	<div class="alert alert-warning flex-col items-start gap-3" role="alert">
		<h3 class="font-semibold">Can't check MusicBrainz right now</h3>
		<p>
			The matching service isn't answering. Your album is already uploaded and waiting; no need to
			send it again. Try again in a little while.
		</p>
		<button type="button" class="btn btn-primary" onclick={outcome.onTryAgain}>Try again</button>
	</div>
{:else if outcome.kind === 'queued'}
	<div class="alert alert-info flex-col items-start gap-2" role="status">
		<h3 class="font-semibold">Waiting its turn</h3>
		<p>
			{outcome.ownEarlierUpload
				? 'Your earlier album is still importing. This one starts as soon as it finishes.'
				: 'The library is finishing another import. This one starts right after.'}
		</p>
	</div>
{:else}
	<div class="alert alert-warning flex-col items-start gap-2" role="status">
		<h3 class="font-semibold">Taking longer than it should</h3>
		<p>
			Something behind the scenes is slow. The admin gets notified automatically when this happens,
			and this page keeps checking on its own. Nothing needs to be redone.
		</p>
	</div>
{/if}
