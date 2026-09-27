<!--
@component
The login page: a single column (no photo, until the photo curation
pass supplies one), the group's name as the heading. The `h1` is the
served site name, falling back to the deck's own login title when the
site-name meta is absent (a11y W6), and never changes with an outcome;
the "sent" and "requested" outcomes replace the form with a centered
moment instead, per W7.15 and W7.16: a short slab title plus the full
sentence in body copy (S1), never a full sentence set in Alfa Slab. The
moment's heading takes focus, so the moment is never also announced.
Every other outcome shows an alert beside the still-editable form, as
does the signed-out state a sign-out lands on, and is announced through
the layout's live regions. `main` is the root layout's skip-link target
(W3.6).
-->
<script lang="ts">
  import { tick } from 'svelte';
  import { browser } from '$app/environment';
  import { page } from '$app/state';
  import Mark from '$lib/components/Mark.svelte';
  import { requestLink, outcomeAlertClass, outcomePoliteness, outcomeText, type AuthResult } from '$lib/api/auth';
  import { announce } from '$lib/announcer.svelte';
  import { ADD_MEMBER_LABELS, AUTH_OUTCOMES, LOGIN_SUBLINE, SENT_ESCAPE_HATCH, TITLES } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  const siteNameFromMeta = browser ? siteNameFromDocument(document) : '';
  // a11y W6: an absent site-name meta falls back to the deck's own
  // login title, so the `h1` never renders empty. The fallback is for
  // the heading only: `pageTitle` gets the meta value as-is, since an
  // empty site name is its own documented case (the state alone, no
  // separator).
  const siteName = $derived(siteNameFromMeta || TITLES.login);

  let address = $state('');
  let submitting = $state(false);
  let addressInput = $state<HTMLInputElement | null>(null);
  // Server outcomes that name the address field specifically; every
  // other outcome renders only the shared alert beside the form.
  let addressInvalid = $state(false);
  // A sign-out lands here through `goto` with `signedOut` in the page
  // state (`SiteHeader.svelte`), and the page opens on that outcome.
  const signedOut: AuthResult | null = page.state.signedOut ? { outcome: 'signed_out', adminContact: null } : null;
  if (signedOut) announce(outcomeText(signedOut), outcomePoliteness(signedOut.outcome));

  // Never read into; a fresh answer always replaces the whole object.
  let result = $state.raw<AuthResult | null>(signedOut);
  let momentHeading = $state<HTMLElement | null>(null);

  const moment = $derived(result && (result.outcome === 'sent' || result.outcome === 'requested') ? result : null);
  const titleState = $derived(moment ? TITLES.loginSent : result?.outcome === 'signed_out' ? TITLES.signedOut : TITLES.login);

  // W7.16: an async transition that unmounts the focused control
  // moves focus deliberately, in an effect keyed on the discriminant
  // plus tick(), rather than leaving it to drop to <body>.
  $effect(() => {
    if (moment !== null) tick().then(() => momentHeading?.focus());
  });

  function clearAddressError(): void {
    addressInvalid = false;
  }

  async function submit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (submitting) return;
    submitting = true;
    // The stale outcome leaves while the request is in flight, so a
    // repeated identical answer visibly arrives again.
    result = null;
    addressInvalid = false;
    announce('', 'polite');
    try {
      const answer = await requestLink(address);
      result = answer;
      if (answer.outcome !== 'sent' && answer.outcome !== 'requested') {
        announce(outcomeText(answer), outcomePoliteness(answer.outcome));
      }
      if (answer.outcome === 'invalid_address') {
        addressInvalid = true;
        await tick();
        addressInput?.focus();
      }
    } finally {
      submitting = false;
    }
  }

  // The escape hatch back to the form keeps the typed address (W6.11):
  // `address` is never cleared, only the outcome that replaced the
  // form. a11y W2: focus moves to the reinstated email field, since the
  // control that had it just unmounted.
  async function useDifferentAddress(): Promise<void> {
    result = null;
    announce('', 'polite');
    await tick();
    addressInput?.focus();
  }
</script>

<svelte:head>
  <title>{pageTitle(titleState, siteNameFromMeta)}</title>
</svelte:head>

<main id="main" tabindex="-1" class="content-measure flow py-8">
  <div class="flex items-center gap-2">
    <Mark size={32} />
    <span class="display text-3xl">dubplate</span>
  </div>

  <h1 class="display page-title text-4xl">{siteName}</h1>
  {#if !moment}
    <p>{LOGIN_SUBLINE}</p>
  {/if}

  {#if moment}
    <section id="outcome">
      <h2 tabindex="-1" bind:this={momentHeading} class="display text-xl">{titleState}</h2>
      <p>{outcomeText(moment)}</p>
      <button type="button" class="link link-primary" onclick={useDifferentAddress}>{SENT_ESCAPE_HATCH}</button>
    </section>
  {:else}
    <form onsubmit={submit} class="flow">
      <fieldset class="fieldset">
        <legend class="fieldset-legend sr-only">{TITLES.login}</legend>
        <label class="label" for="address">{ADD_MEMBER_LABELS.address}</label>
        <input
          id="address"
          name="address"
          type="email"
          autocomplete="email"
          required
          class="input validator {addressInvalid ? 'input-error' : ''}"
          aria-invalid={addressInvalid ? 'true' : undefined}
          aria-describedby={addressInvalid ? 'address-hint' : undefined}
          oninput={clearAddressError}
          bind:value={address}
          bind:this={addressInput}
        />
        <p id="address-hint" class="validator-hint" hidden={!addressInvalid}>{AUTH_OUTCOMES.invalid_address}</p>
      </fieldset>
      <button type="submit" class="btn btn-primary" aria-disabled={submitting || undefined}>{TITLES.login}</button>
    </form>
    {#if result}
      <div id="outcome" class="alert {outcomeAlertClass(result.outcome)}">{outcomeText(result)}</div>
    {/if}
  {/if}
</main>
