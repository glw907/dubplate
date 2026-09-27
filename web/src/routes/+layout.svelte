<!--
@component
The root layout. Every route imports its CSS bundle through here, since
a SvelteKit static build has exactly one entry that reaches every page.
It also carries what every page shares: the skip link, first in tab
order ahead of any page's header (W3.6), targeting each page's own
`<main id="main">`, and the app's one pair of live regions
(`$lib/announcer.svelte`), rendered unconditionally so a page writes
into regions the browser already tracks.
-->
<script lang="ts">
  import '../app.css';
  import { announcer } from '$lib/announcer.svelte';

  let { children } = $props();
</script>

<a
  href="#main"
  class="skip-link sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:bg-base-100 focus:px-4 focus:py-2 focus:text-base-content"
  >Skip to content</a
>

<div id="announce-polite" aria-live="polite" class="sr-only">{announcer.polite}</div>
<div id="announce-assertive" aria-live="assertive" class="sr-only">{announcer.assertive}</div>

{@render children()}
