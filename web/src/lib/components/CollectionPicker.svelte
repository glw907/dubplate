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
	 * `internal/serve/upload.go`'s `defaultPrivateCollection`, mirrored
	 * here because the wire config carries only the collection name
	 * list, never a flag marking which entry is the private lane --
	 * "Private Recordings" is the documented default name and the
	 * literal string the copy contract renders regardless of config.
	 */
	export const PRIVATE_COLLECTION_NAME = 'Private Recordings';

	export function isPrivateCollection(name: string): boolean {
		return name === PRIVATE_COLLECTION_NAME;
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
