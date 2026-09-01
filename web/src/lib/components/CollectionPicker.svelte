<!--
  @component
  The collection selector (passb-copy-final.md section 1): renders
  nothing when `/api/config` lists one collection or fewer, since a
  single-collection operator has no choice to present. Options are the
  configured names verbatim -- no relabeling. {@link isPrivateCollection}
  is exported so the page can swap the format panel and tag form the
  instant the private lane is chosen, per the plan's "the UI lets the
  user know what's happening" ruling.
-->
<script lang="ts" module>
	/**
	 * Reports whether name is the private lane, against
	 * privateCollectionName -- `/api/config`'s own `private_collection`
	 * field (`internal/serve/upload.go`'s `Engine.PrivateCollection`),
	 * never a hardcoded "Private Recordings" guess. The server default is
	 * the only value production serves today; no config field feeds
	 * `Options.PrivateCollection` yet, so this seam is what a rename
	 * would ride, not a knob an operator can already turn.
	 */
	export function isPrivateCollection(name: string, privateCollectionName: string): boolean {
		return name !== '' && name === privateCollectionName;
	}
</script>

<script lang="ts">
	interface Props {
		collections: string[];
		selected: string;
		onSelect: (collection: string) => void;
	}

	let { collections, selected, onSelect }: Props = $props();
</script>

{#if collections.length > 1}
	<div class="fieldset w-full max-w-xs">
		<label class="fieldset-label" for="collection-picker">Collection</label>
		<select
			id="collection-picker"
			class="select min-h-11 w-full"
			value={selected}
			onchange={(event) => onSelect(event.currentTarget.value)}
		>
			{#each collections as name (name)}
				<option value={name}>{name}</option>
			{/each}
		</select>
	</div>
{/if}
