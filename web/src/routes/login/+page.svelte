<!--
@component
The login page: a single column (no photo, until the photo curation
pass supplies one), the group's name as the heading. The `h1` is the
served site name and never changes with an outcome; the "sent" and
"requested" outcomes replace the form with a centered moment instead,
per W7.15 and W7.16. Every other outcome shows an alert beside the
still-editable form. `main` and the skip link satisfy W3.6.
-->
<script lang="ts">
  import { tick } from 'svelte';
  import { browser } from '$app/environment';
  import Mark from '$lib/components/Mark.svelte';
  import { requestLink, outcomeAlertClass, outcomePoliteness, outcomeText, type AuthResult } from '$lib/api/auth';
  import { ADD_MEMBER_LABELS, AUTH_OUTCOMES, LOGIN_SUBLINE, SENT_ESCAPE_HATCH, TITLES } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  const siteName = browser ? siteNameFromDocument(document) : '';

  let address = $state('');
  let submitting = $state(false);
  let result = $state<AuthResult | null>(null);
  let momentHeading = $state<HTMLElement | null>(null);
  let politeText = $state('');
  let assertiveText = $state('');

  const moment = $derived(result && (result.outcome === 'sent' || result.outcome === 'requested') ? result : null);

  function announce(text: string, politeness: 'polite' | 'assertive'): void {
    if (politeness === 'polite') {
      assertiveText = '';
      politeText = text;
    } else {
      politeText = '';
      assertiveText = text;
    }
  }

  // W7.16: an async transition that unmounts the focused control
  // moves focus deliberately, in an effect keyed on the discriminant
  // plus tick(), rather than leaving it to drop to <body>.
  $effect(() => {
    if (moment !== null) tick().then(() => momentHeading?.focus());
  });

  async function submit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (submitting) return;
    submitting = true;
    try {
      result = await requestLink(address);
      announce(outcomeText(result), outcomePoliteness(result.outcome));
    } finally {
      submitting = false;
    }
  }

  // The escape hatch back to the form keeps the typed address (W6.11):
  // `address` is never cleared, only the outcome that replaced the form.
  function useDifferentAddress(): void {
    result = null;
    announce('', 'polite');
  }
</script>

<svelte:head>
  <title>{pageTitle(moment ? TITLES.loginSent : TITLES.login, siteName)}</title>
</svelte:head>

<a
  href="#main"
  class="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:bg-base-100 focus:px-4 focus:py-2 focus:text-base-content"
  >Skip to content</a
>

<!--
  W7.6: one persistent, unconditionally rendered live region per
  politeness. The visible alert below carries no role or aria-live of
  its own, so only these announce; a repeated identical outcome would
  need to clear and re-set on the next frame to re-announce, which no
  outcome here repeats back to back since a fresh submit always
  replaces the previous result.
-->
<div id="announce-polite" aria-live="polite" class="sr-only">{politeText}</div>
<div id="announce-assertive" aria-live="assertive" class="sr-only">{assertiveText}</div>

<main id="main" tabindex="-1" class="content-measure flow p-8">
  <div class="flex items-center gap-2">
    <Mark size={32} />
    <span class="display text-lg">dubplate</span>
  </div>

  <h1 class="display page-title">{siteName}</h1>
  {#if !moment}
    <p>{LOGIN_SUBLINE}</p>
  {/if}

  {#if moment}
    <section id="outcome">
      <h2 tabindex="-1" bind:this={momentHeading} class="display">{outcomeText(moment)}</h2>
      <button type="button" class="btn btn-ghost" onclick={useDifferentAddress}>{SENT_ESCAPE_HATCH}</button>
    </section>
  {:else}
    <form onsubmit={submit}>
      <fieldset class="fieldset">
        <legend class="fieldset-legend">{TITLES.login}</legend>
        <label class="label" for="address">{ADD_MEMBER_LABELS.address}</label>
        <input
          id="address"
          name="address"
          type="email"
          autocomplete="email"
          required
          class="input validator"
          bind:value={address}
        />
        <p class="validator-hint">{AUTH_OUTCOMES.invalid_address}</p>
      </fieldset>
      <button type="submit" class="btn btn-primary" disabled={submitting}>{TITLES.login}</button>
    </form>
    {#if result}
      <div id="outcome" class="alert {outcomeAlertClass(result.outcome)}">{outcomeText(result)}</div>
    {/if}
  {/if}
</main>
