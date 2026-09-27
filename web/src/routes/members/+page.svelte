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
  import { browser } from '$app/environment';
  import { resolve } from '$app/paths';
  import { postForm } from '$lib/api/post';
  import { announce } from '$lib/announcer.svelte';
  import SiteHeader from '$lib/components/SiteHeader.svelte';
  import {
    ADD_MEMBER_HINTS,
    ADD_MEMBER_LABELS,
    ADD_MEMBER_LEGEND,
    ADD_MEMBER_SUBMIT,
    ADMIN_BADGE_LABEL,
    AUTH_OUTCOMES,
    CROSS_ORIGIN_REFUSED,
    ERROR_PAGE,
    NETWORK_FAILURE,
    ROSTER_EMPTY_STATE,
    ROSTER_OUTCOMES,
    ROSTER_RETRY,
    TITLES,
    memberCountLine
  } from '$lib/copy/deck';
  import { pageTitle, siteNameFromDocument } from '$lib/copy/title';

  // The roster API's own subtree name (`cmd/dubplate-server/serve.go`'s
  // `rosterRouteName`), composed from the runtime base.
  const ROSTER_PATH = 'roster';

  // Which field a roster-write outcome names, so the field itself can
  // carry the invalid state, the description, and the focus a server
  // refusal earns, rather than only the alert beneath the form.
  const FIELD_FOR_OUTCOME: Record<string, 'address' | 'display_name' | 'streamer_username'> = {
    duplicate_address: 'address',
    blank_display_name: 'display_name',
    invalid_display_name: 'display_name',
    missing_streamer_username: 'streamer_username',
    invalid_streamer_username: 'streamer_username',
    duplicate_streamer_username: 'streamer_username',
    duplicate_principal: 'streamer_username'
  };

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
  let streamerUsername = $state('');
  let submitting = $state(false);
  let outcome = $state<string | null>(null);
  let erroredField = $state<'address' | 'display_name' | 'streamer_username' | null>(null);
  let addMemberForm = $state<HTMLFormElement | null>(null);
  let addressInput = $state<HTMLInputElement | null>(null);

  // A roster member's meta line joins only the segments it actually
  // carries: an address always has one, but an older member's
  // `streamer_username` can be blank, and a bare address followed by a
  // dangling " · " is the render-and-read's "members lists" defect.
  // Every segment after the first leads with a non-breaking space, so
  // each part renders in its own `nowrap` span (the re-read's item 2):
  // a wrap can only land ahead of a separator, never inside a segment
  // at its own hyphen, and never right after the separator leaving it
  // orphaned at the end of a line.
  function memberMetaParts(member: RosterMember): string[] {
    const segments = [member.address, member.streamer_username].filter((segment) => segment !== '');
    return segments.map((segment, i) => (i === 0 ? segment : ` · ${segment}`));
  }

  // The line for a roster outcome, an auth-shared outcome (a cross-origin
  // refusal or a network failure that never reached a roster answer, or
  // a status like 'failed'), or a last-resort deck line if the server
  // answers with something neither dictionary names. The lookups test
  // own properties, so an outcome named after an inherited member
  // (`toString`, `constructor`) falls through to that last resort.
  function resolveOutcome(o: string): string {
    if (o === 'cross_origin') return CROSS_ORIGIN_REFUSED;
    if (o === 'network_failure') return NETWORK_FAILURE;
    if (Object.hasOwn(ROSTER_OUTCOMES, o)) return ROSTER_OUTCOMES[o];
    if (Object.hasOwn(AUTH_OUTCOMES, o)) return AUTH_OUTCOMES[o];
    return AUTH_OUTCOMES.failed;
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

  function clearFieldError(field: 'address' | 'display_name' | 'streamer_username'): void {
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
    try {
      const body = (await response.json()) as { members: RosterMember[] };
      return { ok: true, members: body.members };
    } catch {
      return { ok: false, notFound: false, message: AUTH_OUTCOMES.failed };
    }
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
        streamer_username: streamerUsername
      });
      outcome = answer.outcome;
      erroredField = FIELD_FOR_OUTCOME[answer.outcome] ?? null;
      announce(resolveOutcome(outcome), outcomePoliteness(outcome));
      if (outcome === 'added') {
        address = '';
        displayName = '';
        streamerUsername = '';
        // The native reset clears the form's own dirty-value flag, so
        // the required streamer-username field drops its `:user-invalid`
        // state instead of showing invalid the moment it goes back to
        // empty (the render-and-read's "members-added" defect).
        addMemberForm?.reset();
        await refreshAfterAdd();
        // W1: a successful add returns focus to the address field
        // rather than leaving it on the button the reset just cleared.
        await tick();
        addressInput?.focus();
      } else if (erroredField === 'address') {
        await tick();
        addressInput?.focus();
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
              <div class="font-mono flex flex-wrap text-sm opacity-70">
                {#each memberMetaParts(member) as part, i (i)}<span class="whitespace-nowrap">{part}</span>{/each}
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
          autocomplete="email"
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
        />

        <label class="label" for="add-streamer-username">{ADD_MEMBER_LABELS.streamerUsername}</label>
        <input
          id="add-streamer-username"
          name="streamer_username"
          type="text"
          pattern="[A-Za-z0-9][A-Za-z0-9._@+\-]*"
          maxlength="64"
          required
          class="input validator {erroredField === 'streamer_username' ? 'input-error' : ''}"
          aria-invalid={erroredField === 'streamer_username' ? 'true' : undefined}
          aria-describedby={erroredField === 'streamer_username' ? 'streamer-hint roster-outcome' : 'streamer-hint'}
          oninput={() => clearFieldError('streamer_username')}
          bind:value={streamerUsername}
        />
        <p id="streamer-hint" class="text-sm opacity-70">{ADD_MEMBER_HINTS.streamerUsername}</p>
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
