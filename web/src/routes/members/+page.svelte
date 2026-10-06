<!--
@component
The add-member page: an admin's roster list and add form, behind the
roster API's own admin gate (`internal/auth/admin.go`'s `RequireAdmin`).
The page ships in public `_app/immutable/` chunks and renders nothing it
did not fetch; a non-admin's GET meets the gate's 404 and sees the
not-found state instead of the roster, with no roster and no form. A GET
that fails for any other reason (a network failure, or a status the
gate itself never answers with) shows its own unreachable state instead,
with a control to try again, and never discards a roster the page has
already loaded.
The signed-in header renders in every state, loading included, and
`main` is the root layout's skip-link target.
-->
<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { browser } from '$app/env';
  import { endpointPath, postForm } from '#lib/api/post.js';
  import { outcomeAlertClass, outcomePoliteness, outcomeText } from '#lib/api/outcome.js';
  import { announce } from '#lib/announcer.svelte.js';
  import SiteHeader from '#lib/components/SiteHeader.svelte';
  import {
    ADD_MEMBER_HINTS,
    ADD_MEMBER_LABELS,
    ADD_MEMBER_LEGEND,
    ADD_MEMBER_SUBMIT,
    ADMIN_BADGE_LABEL,
    AUTH_OUTCOMES,
    ERROR_PAGE,
    NETWORK_FAILURE,
    ROSTER_EMPTY_STATE,
    ROSTER_RETRY,
    TITLES,
    addedMemberLine,
    memberCountLine
  } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  // The roster API's own subtree name (`cmd/dubplate-server/serve.go`'s
  // `rosterRouteName`), composed from the runtime base.
  const ROSTER_PATH = 'roster';

  // Which field a roster-write outcome names, so the field itself can
  // carry the invalid state, the description, and the focus a server
  // refusal earns, rather than only the alert beneath the form.
  const FIELD_FOR_OUTCOME: Record<string, 'address' | 'display_name' | 'username'> = {
    duplicate_address: 'address',
    blank_display_name: 'display_name',
    invalid_display_name: 'display_name',
    missing_username: 'username',
    invalid_username: 'username',
    duplicate_username: 'username',
    duplicate_principal: 'username'
  };

  interface RosterMember {
    address: string;
    display_name: string;
    username: string;
    admin: boolean;
  }

  const siteName = browser ? siteNameFromDocument(document) : '';

  function apiURL(): string {
    return endpointPath(ROSTER_PATH);
  }

  type Status = 'loading' | 'ok' | 'refused' | 'unreachable';

  let status = $state<Status>('loading');
  // A list this size or larger stops being a handful of rows to diff on
  // every fetch, so the roster and the form's transient fields skip
  // Svelte's deep proxy: nothing here reads or writes into their
  // contents through a derived value.
  let members = $state.raw<RosterMember[]>([]);
  let unreachableMessage = $state('');
  // Set only by a refresh after a successful add: the page keeps
  // showing the list and the form, and reports the refresh's own
  // failure beside them instead of replacing the page (svelte #5, a11y
  // W9).
  let refreshFailed = $state(false);

  let address = $state('');
  let displayName = $state('');
  let username = $state('');
  let submitting = $state(false);
  let outcome = $state<string | null>(null);
  let erroredField = $state<'address' | 'display_name' | 'username' | null>(null);
  let addMemberForm = $state<HTMLFormElement | null>(null);
  let addressInput = $state<HTMLInputElement | null>(null);
  let displayNameInput = $state<HTMLInputElement | null>(null);
  let usernameInput = $state<HTMLInputElement | null>(null);

  // W5: every field-specific server outcome moves focus to the field it
  // marks, not only the address field.
  function inputForField(field: 'address' | 'display_name' | 'username'): HTMLInputElement | null {
    if (field === 'address') return addressInput;
    if (field === 'display_name') return displayNameInput;
    return usernameInput;
  }

  // The line for a roster outcome: `#lib/api/outcome.ts`'s shared
  // resolver, except `added`, which names the member the write just
  // added and so needs its own deck function instead of a dictionary
  // lookup.
  function resolveOutcome(o: string): string {
    return o === 'added' ? addedMemberLine(addedMemberName) : outcomeText(o);
  }

  // The added member's own name, captured before the form's reset
  // clears `displayName`, since the outcome alert renders after that
  // reset runs.
  let addedMemberName = $state('');

  function clearFieldError(field: 'address' | 'display_name' | 'username'): void {
    if (erroredField === field) erroredField = null;
  }

  /**
   * One roster GET, read into its three possible shapes: a member list,
   * the admin gate's own 404, or anything else (a network failure, a
   * non-2xx status, or a body that is not the roster's own JSON). The
   * caller decides how each shape changes the page.
   */
  async function fetchRoster(): Promise<
    { ok: true; members: RosterMember[] } | { ok: false; notFound: true } | { ok: false; notFound: false; message: string }
  > {
    let response: Response;
    try {
      response = await fetch(apiURL());
    } catch {
      return { ok: false, notFound: false, message: NETWORK_FAILURE };
    }
    if (response.status === 404) return { ok: false, notFound: true };
    if (!response.ok) return { ok: false, notFound: false, message: AUTH_OUTCOMES.failed };
    let body: unknown;
    try {
      body = await response.json();
    } catch {
      return { ok: false, notFound: false, message: AUTH_OUTCOMES.failed };
    }
    if (typeof body !== 'object' || body === null || !Array.isArray((body as Record<string, unknown>).members)) {
      return { ok: false, notFound: false, message: AUTH_OUTCOMES.failed };
    }
    return { ok: true, members: (body as { members: RosterMember[] }).members };
  }

  async function loadRoster(): Promise<void> {
    const result = await fetchRoster();
    if (result.ok) {
      members = result.members;
      status = 'ok';
      return;
    }
    if (result.notFound) {
      status = 'refused';
      return;
    }
    status = 'unreachable';
    unreachableMessage = result.message;
  }

  // The add form's own post-write refresh: it never turns the page's
  // `status` away from `'ok'`, so a refresh failure right after a
  // successful add still shows the list and the form (a11y W9, svelte
  // #5). The roster the page already has stays on screen.
  async function refreshAfterAdd(): Promise<void> {
    const result = await fetchRoster();
    if (result.ok) {
      members = result.members;
      refreshFailed = false;
    } else {
      refreshFailed = true;
    }
  }

  onMount(() => {
    void loadRoster();
  });

  async function submit(event: SubmitEvent): Promise<void> {
    event.preventDefault();
    if (submitting) return;
    submitting = true;
    // The stale outcome leaves while the request is in flight, so a
    // repeated identical answer visibly arrives again.
    outcome = null;
    erroredField = null;
    announce('', 'polite');
    try {
      const answer = await postForm(ROSTER_PATH, {
        address,
        display_name: displayName,
        username
      });
      outcome = answer.outcome;
      erroredField = FIELD_FOR_OUTCOME[answer.outcome] ?? null;
      if (outcome === 'added') addedMemberName = displayName;
      announce(resolveOutcome(outcome), outcomePoliteness(outcome));
      if (outcome === 'added') {
        address = '';
        displayName = '';
        username = '';
        // The native reset clears the form's own dirty-value flag, so
        // the required username field drops its `:user-invalid`
        // state instead of showing invalid the moment it goes back to
        // empty (the render-and-read's "members-added" defect).
        addMemberForm?.reset();
        await refreshAfterAdd();
        // W1: a successful add returns focus to the address field
        // rather than leaving it on the button the reset just cleared.
        await tick();
        addressInput?.focus();
      } else if (erroredField) {
        await tick();
        inputForField(erroredField)?.focus();
      }
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{pageTitle(status === 'refused' ? TITLES.notFound : TITLES.members, siteName)}</title>
</svelte:head>

<SiteHeader />

<main id="main" tabindex="-1" class="content-measure flow py-8">
  {#if status === 'refused'}
    <h1 class="display page-title text-4xl">{TITLES.notFound}</h1>
    <p>{ERROR_PAGE.notFoundHeading}</p>
  {:else if status === 'unreachable'}
    <h1 class="display page-title text-4xl">{TITLES.members}</h1>
    <div class="alert alert-warning">{unreachableMessage}</div>
    <button type="button" class="btn btn-primary" onclick={loadRoster}>{ROSTER_RETRY}</button>
  {:else if status === 'ok'}
    <h1 class="display page-title text-4xl">{TITLES.members}</h1>

    <p>{members.length === 0 ? ROSTER_EMPTY_STATE : memberCountLine(members.length)}</p>
    {#if members.length > 0}
      <ul id="roster-list" class="list" role="list">
        {#each members as member (member.address)}
          <li class="list-row">
            <div class="list-col-grow">
              <div class="flex items-center gap-2">
                <span class="font-medium">{member.display_name}</span>
                {#if member.admin}<span class="badge">{ADMIN_BADGE_LABEL}</span>{/if}
              </div>
              <div class="font-mono flex flex-wrap gap-x-1 text-xs opacity-70">
                <!-- The separator trails the address, glued on with a non-breaking
                     space, and each part is its own nowrap span: a wrap can only
                     land between the parts, so a wrapped line never starts with a
                     dangling "·", and a blank username leaves no separator at all.
                     The gap on one line is the flex gap, not a text space. -->
                <span class="whitespace-nowrap">{member.address}{#if member.username !== ''}<span aria-hidden="true">&nbsp;·</span>{/if}</span>{#if member.username !== ''}<span class="whitespace-nowrap"><span class="sr-only">{ADD_MEMBER_LABELS.username}{' '}</span>{member.username}</span>{/if}
              </div>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
    {#if refreshFailed}
      <div class="alert alert-warning">{AUTH_OUTCOMES.failed}</div>
    {/if}

    <form onsubmit={submit} bind:this={addMemberForm} class="flow">
      <fieldset class="fieldset">
        <legend class="fieldset-legend">{ADD_MEMBER_LEGEND}</legend>

        <label class="label" for="add-address">{ADD_MEMBER_LABELS.address}</label>
        <input
          id="add-address"
          name="address"
          type="email"
          autocomplete="off"
          data-1p-ignore
          data-lpignore="true"
          required
          class="input validator {erroredField === 'address' ? 'input-error' : ''}"
          aria-invalid={erroredField === 'address' ? 'true' : undefined}
          aria-describedby={erroredField === 'address' ? 'roster-outcome' : undefined}
          oninput={() => clearFieldError('address')}
          bind:value={address}
          bind:this={addressInput}
        />

        <label class="label" for="add-display-name">{ADD_MEMBER_LABELS.displayName}</label>
        <input
          id="add-display-name"
          name="display_name"
          type="text"
          maxlength="200"
          required
          class="input validator {erroredField === 'display_name' ? 'input-error' : ''}"
          aria-invalid={erroredField === 'display_name' ? 'true' : undefined}
          aria-describedby={erroredField === 'display_name' ? 'roster-outcome' : undefined}
          oninput={() => clearFieldError('display_name')}
          bind:value={displayName}
          bind:this={displayNameInput}
        />

        <label class="label" for="add-username">{ADD_MEMBER_LABELS.username}</label>
        <input
          id="add-username"
          name="username"
          type="text"
          autocomplete="off"
          data-1p-ignore
          data-lpignore="true"
          pattern="[A-Za-z0-9][A-Za-z0-9._@+\-]*"
          maxlength="64"
          required
          class="input validator {erroredField === 'username' ? 'input-error' : ''}"
          aria-invalid={erroredField === 'username' ? 'true' : undefined}
          aria-describedby={erroredField === 'username' ? 'username-hint roster-outcome' : 'username-hint'}
          oninput={() => clearFieldError('username')}
          bind:value={username}
          bind:this={usernameInput}
        />
        <p id="username-hint" class="text-sm opacity-70">{ADD_MEMBER_HINTS.username}</p>
      </fieldset>
      <button type="submit" class="btn btn-primary" aria-disabled={submitting || undefined}>{ADD_MEMBER_SUBMIT}</button>
    </form>
    {#if outcome}
      <div id="roster-outcome" class="alert {outcomeAlertClass(outcome)}">
        {resolveOutcome(outcome)}
      </div>
    {/if}
  {/if}
</main>
