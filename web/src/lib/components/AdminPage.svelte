<!--
@component
The shell of an admin page that loads one list: the head title, the
signed-in header, and `main`, the root layout's skip-link target. It
renders the gate states itself: a non-admin's `refused` shows the
not-found state, `signedOut` a link to sign in, and `unreachable` the
page's own title over its failure line with a control to try again.
`loading` renders the header and an empty `main`; only `ok` renders the
page's own content.
-->
<script lang="ts" module>
  export type Status = 'loading' | 'ok' | 'refused' | 'signedOut' | 'unreachable';
</script>

<script lang="ts">
  import type { Snippet } from 'svelte';
  import { browser } from '$app/env';
  import { resolve } from '$app/paths';
  import SiteHeader from './SiteHeader.svelte';
  import { API_REFUSALS, ERROR_PAGE, ROSTER_RETRY, SIGNED_OUT_SIGN_IN, TITLES } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  interface Props {
    /** The page's own title: its `h1` when unreachable, and the head title in every state but `refused` and `signedOut`. */
    title: string;
    status: Status;
    /** The deck line the `unreachable` state shows. */
    unreachableMessage: string;
    /** The list load the `unreachable` state's control runs again. */
    retry: () => void;
    children: Snippet;
  }

  let { title, status, unreachableMessage, retry, children }: Props = $props();

  const siteName = browser ? siteNameFromDocument(document) : '';
</script>

<svelte:head>
  <title>{pageTitle(status === 'refused' ? TITLES.notFound : status === 'signedOut' ? TITLES.signedOut : title, siteName)}</title>
</svelte:head>

<SiteHeader />

<main id="main" tabindex="-1" class="content-measure flow py-8">
  {#if status === 'refused'}
    <h1 class="display page-title text-4xl">{TITLES.notFound}</h1>
    <p>{ERROR_PAGE.notFoundHeading}</p>
  {:else if status === 'signedOut'}
    <h1 class="display page-title text-4xl">{TITLES.signedOut}</h1>
    <p>{API_REFUSALS.unauthenticated}</p>
    <a href={resolve('login')} class="btn btn-primary">{SIGNED_OUT_SIGN_IN}</a>
  {:else if status === 'unreachable'}
    <h1 class="display page-title text-4xl">{title}</h1>
    <div class="alert alert-warning">{unreachableMessage}</div>
    <button type="button" class="btn btn-primary" onclick={retry}>{ROSTER_RETRY}</button>
  {:else if status === 'ok'}
    {@render children()}
  {/if}
</main>
