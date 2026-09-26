<!--
@component
The add-member page: an admin's roster list and add form, behind the
roster API's own admin gate (`internal/auth/admin.go`'s `RequireAdmin`).
The page ships in public `_app/immutable/` chunks and renders nothing it
did not fetch; a non-admin's GET meets the gate's 404 and sees the
not-found state instead of the roster, with no roster and no form.
`main` and the skip link give the page a keyboard-reachable landmark.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/environment';
  import { resolve } from '$app/paths';
  import {
    ADD_MEMBER_HINTS,
    ADD_MEMBER_LABELS,
    AUTH_OUTCOMES,
    CROSS_ORIGIN_REFUSED,
    ERROR_PAGE,
    NETWORK_FAILURE,
    ROSTER_EMPTY_STATE,
    ROSTER_OUTCOMES,
    TITLES,
    memberCountLine
  } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  // The roster API's own subtree name (`cmd/dubplate-server/serve.go`'s
  // `rosterRouteName`), composed from the runtime base.
  const ROSTER_PATH = 'roster';

  // The deck names every field's label and hint but not the form's own
  // legend, submit control, or the admin badge, so these three stay local.
  const ADD_MEMBER_LEGEND = 'Add a member';
  const ADD_MEMBER_SUBMIT = 'Add member';
  const ADMIN_BADGE_LABEL = 'Admin';

  interface RosterMember {
    address: string;
    display_name: string;
    streamer_username: string;
    admin: boolean;
  }

  const siteName = browser ? siteNameFromDocument(document) : '';

  function apiURL(): string {
    return resolve('/') + ROSTER_PATH;
  }

  let status = $state<'loading' | 'ok' | 'refused'>('loading');
  let members = $state<RosterMember[]>([]);

  let address = $state('');
  let displayName = $state('');
  let streamerUsername = $state('');
  let submitting = $state(false);
  let outcome = $state<string | null>(null);
  let politeText = $state('');
  let assertiveText = $state('');

  function announce(text: string, politeness: 'polite' | 'assertive'): void {
    if (politeness === 'polite') {
      assertiveText = '';
      politeText = text;
    } else {
      politeText = '';
      assertiveText = text;
    }
  }

  // The line for a roster outcome, an auth-shared outcome (a cross-origin
  // refusal or a network failure that never reached a roster answer, or
  // a status like 'failed'), or a last-resort deck line if the server
  // answers with something neither dictionary names.
  function resolveOutcome(o: string): string {
    if (o === 'cross_origin') return CROSS_ORIGIN_REFUSED;
    if (o === 'network_failure') return NETWORK_FAILURE;
    return ROSTER_OUTCOMES[o] ?? AUTH_OUTCOMES[o] ?? AUTH_OUTCOMES.failed;
  }

  // 'added' is the only roster outcome that is not a refusal; 'failed'
  // and the two transient write states read as a warning rather than a
  // hard failure, as auth.ts classifies 'failed'.
  function outcomeAlertClass(o: string): string {
    if (o === 'added') return '';
    if (o === 'failed' || o === 'busy' || o === 'write_unconfirmed') return 'alert-warning';
    return 'alert-error';
  }

  function outcomePoliteness(o: string): 'polite' | 'assertive' {
    return o === 'added' ? 'polite' : 'assertive';
  }

  async function loadRoster(): Promise<void> {
    let response: Response;
    try {
      response = await fetch(apiURL());
    } catch {
      status = 'refused';
      return;
    }
    if (!response.ok) {
      status = 'refused';
      return;
    }
    try {
      const body = (await response.json()) as { members: RosterMember[] };
      members = body.members;
      status = 'ok';
    } catch {
      status = 'refused';
    }
  }

  onMount(() => {
    void loadRoster();
  });

  async function submit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (submitting) return;
    submitting = true;
    try {
      let response: Response;
      try {
        response = await fetch(apiURL(), {
          method: 'POST',
          body: new URLSearchParams({
            address,
            display_name: displayName,
            streamer_username: streamerUsername
          }),
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
        });
      } catch {
        outcome = 'network_failure';
        announce(resolveOutcome(outcome), outcomePoliteness(outcome));
        return;
      }
      // The CSRF defense's deny handler answers 403 with a plain-text
      // body, never JSON, so its status is read before any JSON parse.
      if (response.status === 403) {
        outcome = 'cross_origin';
        announce(resolveOutcome(outcome), outcomePoliteness(outcome));
        return;
      }
      let answer: { outcome: string };
      try {
        answer = (await response.json()) as { outcome: string };
      } catch {
        answer = { outcome: 'malformed' };
      }
      outcome = answer.outcome;
      announce(resolveOutcome(outcome), outcomePoliteness(outcome));
      if (outcome === 'added') {
        address = '';
        displayName = '';
        streamerUsername = '';
        await loadRoster();
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{pageTitle(status === 'refused' ? TITLES.notFound : TITLES.members, siteName)}</title>
</svelte:head>

<a
  href="#main"
  class="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:bg-base-100 focus:px-4 focus:py-2 focus:text-base-content"
  >Skip to content</a
>

<div id="announce-polite" aria-live="polite" class="sr-only">{politeText}</div>
<div id="announce-assertive" aria-live="assertive" class="sr-only">{assertiveText}</div>

<main id="main" tabindex="-1" class="p-8">
  {#if status === 'refused'}
    <h1 class="display page-title">{TITLES.notFound}</h1>
    <p>{ERROR_PAGE.notFoundHeading}</p>
  {:else if status === 'ok'}
    <h1 class="display page-title">{TITLES.members}</h1>

    <p>{members.length === 0 ? ROSTER_EMPTY_STATE : memberCountLine(members.length)}</p>
    {#if members.length > 0}
      <ul id="roster-list" class="list">
        {#each members as member (member.address)}
          <li class="list-row flex flex-col gap-1 sm:flex-row sm:items-center sm:justify-between">
            <div>
              <div class="flex items-center gap-2">
                <span class="font-medium">{member.display_name}</span>
                {#if member.admin}<span class="badge">{ADMIN_BADGE_LABEL}</span>{/if}
              </div>
              <div class="text-sm opacity-70">{member.address} · {member.streamer_username}</div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}

    <form onsubmit={submit}>
      <fieldset class="fieldset">
        <legend class="fieldset-legend">{ADD_MEMBER_LEGEND}</legend>

        <label class="label" for="add-address">{ADD_MEMBER_LABELS.address}</label>
        <input id="add-address" name="address" type="email" autocomplete="email" required class="input" bind:value={address} />

        <label class="label" for="add-display-name">{ADD_MEMBER_LABELS.displayName}</label>
        <input
          id="add-display-name"
          name="display_name"
          type="text"
          maxlength="200"
          required
          class="input"
          bind:value={displayName}
        />

        <label class="label" for="add-streamer-username">{ADD_MEMBER_LABELS.streamerUsername}</label>
        <input
          id="add-streamer-username"
          name="streamer_username"
          type="text"
          pattern="[A-Za-z0-9][A-Za-z0-9._@+\-]*"
          maxlength="64"
          required
          class="input validator"
          bind:value={streamerUsername}
        />
        <p class="validator-hint">{ADD_MEMBER_HINTS.streamerUsername}</p>
      </fieldset>
      <button type="submit" class="btn btn-primary" disabled={submitting}>{ADD_MEMBER_SUBMIT}</button>
    </form>
    {#if outcome}
      <div id="roster-outcome" class="alert {outcomeAlertClass(outcome)}">
        {resolveOutcome(outcome)}
      </div>
    {/if}
  {/if}
</main>
