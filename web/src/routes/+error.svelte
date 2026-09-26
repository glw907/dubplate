<!--
@component
The client router's own not-found state: the Go mux answers a served
404 itself, so this page renders only when client-side navigation
lands on a path no route matches.
-->
<script lang="ts">
  import { browser } from '$app/environment';
  import { page } from '$app/state';
  import { ERROR_PAGE, TITLES } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  const notFound = $derived(page.status === 404);
  const siteName = $derived(browser ? siteNameFromDocument(document) : '');
</script>

<svelte:head>
  <title>{pageTitle(notFound ? TITLES.notFound : TITLES.error, siteName)}</title>
</svelte:head>

<main>
  {#if notFound}
    <h1>{ERROR_PAGE.notFoundHeading}</h1>
  {:else}
    <h1>{ERROR_PAGE.genericHeading}</h1>
  {/if}
</main>
