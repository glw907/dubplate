<!--
@component
The client router's own not-found state: the Go mux answers a served
404 itself, so this page renders only when client-side navigation
lands on a path no route matches. It ships as a plain sentence with
no photo, until the photo curation pass supplies one.
-->
<script lang="ts">
  import { browser } from '$app/env';
  import { page } from '$app/state';
  import { ERROR_PAGE, TITLES } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  const notFound = $derived(page.status === 404);
  // The document's site-name meta is fixed for the page's lifetime, so
  // this reads once rather than tracking a dependency that never changes.
  const siteName = browser ? siteNameFromDocument(document) : '';
  const errorTitleState = $derived(notFound ? TITLES.notFound : TITLES.error);
  const heading = $derived(notFound ? TITLES.notFound : ERROR_PAGE.genericHeading);
  // The not-found body is the same sentence the members page's own
  // refused state shows (S10): the two never drift into two different
  // "page not found" wordings.
  const body = $derived(notFound ? ERROR_PAGE.notFoundHeading : '');
</script>

<svelte:head>
  <title>{pageTitle(errorTitleState, siteName)}</title>
</svelte:head>

<main id="main" tabindex="-1" class="content-measure flow py-8">
  <h1 class="display page-title text-4xl">{heading}</h1>
  {#if body}<p>{body}</p>{/if}
</main>
