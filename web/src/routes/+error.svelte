<!--
@component
The client router's own not-found state: the Go mux answers a served
404 itself, so this page renders only when client-side navigation
lands on a path no route matches. It ships as a plain sentence with
no photo, until the photo curation pass supplies one.
-->
<script lang="ts">
  import { browser } from '$app/environment';
  import { page } from '$app/state';
  import { ERROR_PAGE, TITLES } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  const notFound = $derived(page.status === 404);
  const siteName = $derived(browser ? siteNameFromDocument(document) : '');
  const heading = $derived(notFound ? ERROR_PAGE.notFoundHeading : ERROR_PAGE.genericHeading);
</script>

<svelte:head>
  <title>{pageTitle(notFound ? TITLES.notFound : TITLES.error, siteName)}</title>
</svelte:head>

<main id="main" tabindex="-1" class="content-measure flow py-8">
  <h1 class="display page-title">{heading}</h1>
</main>
