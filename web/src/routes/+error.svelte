<!--
@component
The client router's own not-found state: the Go mux answers a served
404 itself, so this page renders only when client-side navigation
lands on a path no route matches.
-->
<script lang="ts">
  import { page } from '$app/state';
</script>

<svelte:head>
  <title>{page.status === 404 ? 'Page not found' : 'Something went wrong'} · dubplate</title>
</svelte:head>

<main>
  {#if page.status === 404}
    <h1>That page doesn't exist.</h1>
  {:else}
    <h1>Something went wrong.</h1>
  {/if}
</main>
