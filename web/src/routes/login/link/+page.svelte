<!--
@component
The consumption interstitial. It renders a static confirm control on
load and consumes nothing until the control is clicked, so a mail
scanner's GET cannot burn the link. `h1` carries the page's own state
(never the site name, which the login page's heading already is), and
takes focus once the click settles into a terminal state (W7.16).
-->
<script lang="ts">
  import { tick } from 'svelte';
  import { browser } from '$app/environment';
  import { resolve } from '$app/paths';
  import { consumeLink, outcomeAlertClass, outcomePoliteness, outcomeText, type AuthResult } from '$lib/api/auth';
  import { AUTH_OUTCOMES, INTERSTITIAL_CONTROL, INTERSTITIAL_STATEMENT, TITLES } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  const siteName = browser ? siteNameFromDocument(document) : '';

  // Before any other work, the token is read from the URL and the
  // history entry is replaced with the same URL less its query,
  // through the native history.replaceState rather
  // than kit's own $app/navigation replaceState. Kit's replaceState
  // re-embeds page.url.href, which still carries the token, into the
  // history entry's own state under [PAGE_URL_KEY]
  // (@sveltejs/kit client.js:2573 sets `[PAGE_URL_KEY]: page.url.href`
  // from the current, still-tokened, URL); the native call instead
  // keeps history.state exactly as kit's own bootstrap left it (its
  // HISTORY_INDEX/NAVIGATION_INDEX keys, no PAGE_URL_KEY, client.js:371-384),
  // so nothing this page does puts the token anywhere but this one
  // local variable. Kit wraps the native history.replaceState only
  // inside `if (DEV && BROWSER)` (client.js:90-114), to print a
  // console warning; a production build never installs that wrapper,
  // so this call is the plain native one there too.
  let token: string | null = null;
  if (browser) {
    token = new URLSearchParams(location.search).get('t');
    history.replaceState(history.state, '', location.pathname + location.hash);
  }

  type Terminal = 'signed_in' | 'refused';

  let terminal = $state<Terminal | null>(token === null ? 'refused' : null);
  let pending = $state<AuthResult | null>(null);
  let submitting = $state(false);
  let heading = $state<HTMLElement | null>(null);
  let politeText = $state('');
  let assertiveText = $state('');

  // The refused line links back to sign-in through its own phrase
  // ("the sign-in page"), split out of the deck's own sentence so the
  // page adds a link without ever writing that sentence's words a
  // second time.
  const REFUSED_LINK_PHRASE = 'the sign-in page';
  const refusedParts = $derived.by(() => {
    const text = AUTH_OUTCOMES.refused;
    const i = text.indexOf(REFUSED_LINK_PHRASE);
    return i === -1
      ? { before: text, phrase: '', after: '' }
      : {
          before: text.slice(0, i),
          phrase: text.slice(i, i + REFUSED_LINK_PHRASE.length),
          after: text.slice(i + REFUSED_LINK_PHRASE.length)
        };
  });

  const headingText = $derived(
    terminal === 'signed_in' ? TITLES.loginLinkSignedIn : terminal === 'refused' ? TITLES.loginLinkRefused : TITLES.loginLink
  );
  function announce(text: string, politeness: 'polite' | 'assertive'): void {
    if (politeness === 'polite') {
      assertiveText = '';
      politeText = text;
    } else {
      politeText = '';
      assertiveText = text;
    }
  }

  // W7.16: focus moves to the terminal heading once it mounts, in an
  // effect keyed on the discriminant plus tick().
  $effect(() => {
    if (terminal !== null) tick().then(() => heading?.focus());
  });

  async function confirm(): Promise<void> {
    if (submitting || token === null) return;
    submitting = true;
    try {
      const result = await consumeLink(token);
      if (result.outcome === 'signed_in' || result.outcome === 'refused') {
        terminal = result.outcome;
      } else {
        pending = result;
      }
      announce(outcomeText(result), outcomePoliteness(result.outcome));
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{pageTitle(headingText, siteName)}</title>
</svelte:head>

<a
  href="#main"
  class="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:bg-base-100 focus:px-4 focus:py-2 focus:text-base-content"
  >Skip to content</a
>

<div id="announce-polite" aria-live="polite" class="sr-only">{politeText}</div>
<div id="announce-assertive" aria-live="assertive" class="sr-only">{assertiveText}</div>

<main id="main" tabindex="-1" class="p-8">
  <h1 tabindex="-1" bind:this={heading} class="display page-title">{headingText}</h1>

  {#if terminal === 'signed_in'}
    <p id="outcome">{AUTH_OUTCOMES.signed_in}</p>
  {:else if terminal === 'refused'}
    <p id="outcome">
      {refusedParts.before}{#if refusedParts.phrase}<a href={resolve('/login')} class="link link-primary"
          >{refusedParts.phrase}</a
        >{/if}{refusedParts.after}
    </p>
  {:else}
    <p>{INTERSTITIAL_STATEMENT}</p>
    <button type="button" class="btn btn-primary" disabled={submitting} onclick={confirm}>{INTERSTITIAL_CONTROL}</button>
    {#if pending}
      <div id="outcome" class="alert {outcomeAlertClass(pending.outcome)}">{outcomeText(pending)}</div>
    {/if}
  {/if}
</main>
