<!--
@component
The consumption interstitial. It renders a static confirm control on
load and consumes nothing until the control is clicked, so a mail
scanner's GET cannot burn the link. The token comes from
`#lib/api/link-token.js`, which the client hook filled and stripped from
the address bar before kit's router read the URL; this page never reads
`location` or kit's `page.url` for it. `h1` carries the page's own state
(never the site name, which the login page's heading already is), and
takes focus once the page settles into a terminal state (W7.16), so a
terminal state is never also announced.

The `signed_in` state offers no link onward: the consume answer does not
say whether the member is an admin, and the members page answers not
found to anyone else.
-->
<script lang="ts">
  import { tick } from 'svelte';
  import { browser } from '$app/env';
  import { resolve } from '$app/paths';
  import { consumeLink, type AuthResult } from '#lib/api/auth.js';
  import { outcomeAlertClass, outcomePoliteness, outcomeText } from '#lib/api/outcome.js';
  import { takeLinkToken } from '#lib/api/link-token.js';
  import { announce } from '#lib/announcer.svelte.js';
  import { AUTH_OUTCOMES, INTERSTITIAL_CONTROL, INTERSTITIAL_STATEMENT, LINK_TOKEN_MISSING, REFUSED_LINK, TITLES } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  const siteName = browser ? siteNameFromDocument(document) : '';
  const token = browser ? takeLinkToken() : null;

  type Terminal = 'signed_in' | 'refused' | 'missing';

  let terminal = $state<Terminal | null>(token === null ? 'missing' : null);
  // Never read into; a fresh answer always replaces the whole object.
  let pending = $state.raw<AuthResult | null>(null);
  let submitting = $state(false);
  let heading = $state<HTMLElement | null>(null);

  const headingText = $derived(
    terminal === 'signed_in' ? TITLES.loginLinkSignedIn : terminal === null ? TITLES.loginLink : TITLES.loginLinkRefused
  );

  // W7.16: focus moves to the terminal heading once it mounts, in an
  // effect keyed on the discriminant plus tick().
  $effect(() => {
    if (terminal !== null) tick().then(() => heading?.focus());
  });

  async function confirm(): Promise<void> {
    if (submitting || token === null) return;
    submitting = true;
    pending = null;
    try {
      const result = await consumeLink(token);
      if (result.outcome === 'signed_in' || result.outcome === 'refused') {
        terminal = result.outcome;
        announce('', 'polite');
      } else {
        pending = result;
        announce(outcomeText(result.outcome, result.adminContact), outcomePoliteness(result.outcome));
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{pageTitle(headingText, siteName)}</title>
</svelte:head>

<main id="main" tabindex="-1" class="content-measure flow py-8">
  <h1 tabindex="-1" bind:this={heading} class="display page-title text-4xl">{headingText}</h1>

  {#if terminal === 'signed_in'}
    <p id="outcome">{AUTH_OUTCOMES.signed_in}</p>
  {:else if terminal === 'refused'}
    <p id="outcome">
      {REFUSED_LINK.before}<a href={resolve('login')} class="link link-primary">{REFUSED_LINK.linkText}</a
      >{REFUSED_LINK.after}
    </p>
  {:else if terminal === 'missing'}
    <p id="outcome">{LINK_TOKEN_MISSING}</p>
  {:else}
    <p>{INTERSTITIAL_STATEMENT}</p>
    <button type="button" class="btn btn-primary" aria-disabled={submitting || undefined} onclick={confirm}
      >{INTERSTITIAL_CONTROL}</button
    >
    {#if pending}
      <div id="outcome" class="alert {outcomeAlertClass(pending.outcome)}">{outcomeText(pending.outcome, pending.adminContact)}</div>
    {/if}
  {/if}
</main>
