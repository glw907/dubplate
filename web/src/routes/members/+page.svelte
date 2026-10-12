<!--
@component
The members page: an admin's list of the group, each member with a
control to remove them or, for an admin, to demote them, and the form
that adds one. Every call goes to `/api/v1` (`#lib/api/v1.ts`), which
answers a non-admin `forbidden`; the page then shows the not-found
state with no list and no form. A signed-out answer shows its own state
with a link to sign in. A list call that fails any other way shows its
own unreachable state with a control to try again, and never discards a
list the page has already loaded.

A removal or a demotion asks first, in a modal that opens with Cancel
focused, and calls nothing until the member confirms. A write the
server refuses for want of a recent sign-in keeps what the member
entered or chose and offers a sign-in link that opens in a new tab, so
the pending change survives in this one, and a control to retry it.

The signed-in header renders in every state, loading included, and
`main` is the root layout's skip-link target.
-->
<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { browser } from '$app/env';
  import { resolve } from '$app/paths';
  import { call, failureAlertClass, failureText, isSignedOut, needsReauthentication } from '#lib/api/v1.js';
  import type { ApiResult, Member } from '#lib/api/v1.js';
  import { announce } from '#lib/announcer.svelte.js';
  import SiteHeader from '#lib/components/SiteHeader.svelte';
  import {
    ADD_MEMBER_HINTS,
    ADD_MEMBER_LABELS,
    ADD_MEMBER_LEGEND,
    ADD_MEMBER_SUBMIT,
    ADMIN_BADGE_LABEL,
    AUTH_OUTCOMES,
    CONFIRM_CANCEL,
    CONFIRM_DEMOTE,
    CONFIRM_REMOVE,
    DEMOTE_CONTROL,
    ERROR_PAGE,
    REAUTH_SIGN_IN,
    REMOVE_CONTROL,
    ROSTER_EMPTY_STATE,
    ROSTER_RETRY,
    SIGNED_OUT_SIGN_IN,
    TITLES,
    addedMemberLine,
    API_REFUSALS,
    confirmDemoteLine,
    confirmRemoveLine,
    demoteControlLabel,
    demotedMemberLine,
    memberCountLine,
    removeControlLabel,
    removedMemberLine
  } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  type Field = 'address' | 'display_name' | 'username';

  // Which field an add refusal's detail names, so the field itself can
  // carry the invalid state, the description, and the focus a server
  // refusal earns, rather than only the alert beneath the form.
  const FIELD_FOR_DETAIL: Record<string, Field> = {
    duplicate_address: 'address',
    invalid_address: 'address',
    blank_display_name: 'display_name',
    invalid_display_name: 'display_name',
    missing_username: 'username',
    invalid_username: 'username',
    duplicate_username: 'username'
  };

  const siteName = browser ? siteNameFromDocument(document) : '';

  type Status = 'loading' | 'ok' | 'refused' | 'signedOut' | 'unreachable';

  /** What the alert beneath the form says about the last write. */
  interface Notice {
    text: string;
    alertClass: string;
    politeness: 'polite' | 'assertive';
    /** Set when the write needs a fresh sign-in: the alert offers the sign-in link and a retry. */
    reauthenticate: boolean;
  }

  let status = $state<Status>('loading');
  // The list and the form's transient fields skip Svelte's deep proxy:
  // nothing here reads or writes into their contents through a derived
  // value.
  let members = $state.raw<Member[]>([]);
  let unreachableMessage = $state('');
  // Set only by a refresh after a successful write: the page keeps
  // showing the list and the form, and reports the refresh's own
  // failure beside them instead of replacing the page.
  let refreshFailed = $state(false);

  let address = $state('');
  let displayName = $state('');
  let username = $state('');
  let submitting = $state(false);
  let notice = $state<Notice | null>(null);
  // The write the alert's retry control runs again.
  let retry = $state<(() => Promise<void>) | null>(null);
  let erroredField = $state<Field | null>(null);
  let addMemberForm = $state<HTMLFormElement | null>(null);
  let addressInput = $state<HTMLInputElement | null>(null);
  let displayNameInput = $state<HTMLInputElement | null>(null);
  let usernameInput = $state<HTMLInputElement | null>(null);

  // The confirm step: the change the member has asked for and not yet
  // confirmed, the modal that holds it, and the control that opened it,
  // which gets the focus back when the modal closes.
  let pending = $state<{ kind: 'remove' | 'demote'; member: Member } | null>(null);
  let confirmDialog = $state<HTMLDialogElement | null>(null);
  let opener: HTMLElement | null = null;

  $effect(() => {
    if (pending && confirmDialog && !confirmDialog.open) confirmDialog.showModal();
  });

  function inputForField(field: Field): HTMLInputElement | null {
    if (field === 'address') return addressInput;
    if (field === 'display_name') return displayNameInput;
    return usernameInput;
  }

  function clearFieldError(field: Field): void {
    if (erroredField === field) erroredField = null;
  }

  /** A read of `list_members`, in the three shapes the page changes on. */
  async function fetchRoster(): Promise<
    { ok: true; members: Member[] } | { ok: false; state: 'refused' | 'signedOut' | 'unreachable'; message: string }
  > {
    const result = await call<{ members: unknown }>('list_members');
    if (result.kind === 'ok') {
      if (!Array.isArray(result.value.members)) return { ok: false, state: 'unreachable', message: AUTH_OUTCOMES.failed };
      return { ok: true, members: result.value.members as Member[] };
    }
    // A non-admin meets `forbidden` and sees the not-found state.
    if (result.kind === 'refused' && result.code === 'forbidden') return { ok: false, state: 'refused', message: '' };
    if (isSignedOut(result)) return { ok: false, state: 'signedOut', message: '' };
    return { ok: false, state: 'unreachable', message: failureText(result) };
  }

  async function loadRoster(): Promise<void> {
    const result = await fetchRoster();
    if (result.ok) {
      members = result.members;
      status = 'ok';
      return;
    }
    status = result.state;
    unreachableMessage = result.message;
  }

  // A write's own refresh: it never turns the page's `status` away from
  // `'ok'`, so a refresh failure right after a successful write still
  // shows the list and the form. The list the page already has stays on
  // screen.
  async function refreshAfterWrite(): Promise<void> {
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

  function say(text: string, alertClass: string, politeness: 'polite' | 'assertive', reauthenticate = false): void {
    notice = { text, alertClass, politeness, reauthenticate };
    announce(text, politeness);
  }

  /**
   * Reads a write's result into the alert beneath the form. It returns
   * true when the write went through. A signed-out answer replaces the
   * page; a refusal that needs a fresh sign-in keeps `again` for the
   * retry control; a refusal that names an add field returns its field.
   */
  function settle(result: ApiResult<unknown>, again: () => Promise<void>): boolean {
    if (result.kind === 'ok') return true;
    if (isSignedOut(result)) {
      status = 'signedOut';
      return false;
    }
    const reauthenticate = needsReauthentication(result);
    if (reauthenticate) retry = again;
    say(failureText(result), failureAlertClass(result), 'assertive', reauthenticate);
    if (result.kind === 'refused') erroredField = FIELD_FOR_DETAIL[result.detail] ?? null;
    return false;
  }

  async function addMember(): Promise<void> {
    if (submitting) return;
    submitting = true;
    // The stale alert leaves while the request is in flight, so a
    // repeated identical answer visibly arrives again.
    notice = null;
    retry = null;
    erroredField = null;
    announce('', 'polite');
    try {
      const result = await call<{ member: Member }>('add_member', { address, display_name: displayName, username });
      if (!settle(result, addMember)) {
        if (erroredField) {
          await tick();
          inputForField(erroredField)?.focus();
        }
        return;
      }
      say(addedMemberLine(displayName), '', 'polite');
      address = '';
      displayName = '';
      username = '';
      // The native reset clears the form's own dirty-value flag, so the
      // required username field drops its `:user-invalid` state instead
      // of showing invalid the moment it goes back to empty.
      addMemberForm?.reset();
      await refreshAfterWrite();
      // A successful add returns focus to the address field rather than
      // leaving it on the button the reset just cleared.
      await tick();
      addressInput?.focus();
    } finally {
      submitting = false;
    }
  }

  function submit(event: SubmitEvent): void {
    event.preventDefault();
    void addMember();
  }

  function ask(kind: 'remove' | 'demote', member: Member, event: MouseEvent): void {
    opener = event.currentTarget as HTMLElement;
    pending = { kind, member };
  }

  /** The modal closed, by Cancel, Escape, or a confirm: nothing is pending and focus goes back. */
  function closed(): void {
    pending = null;
    opener?.focus();
  }

  async function change(kind: 'remove' | 'demote', member: Member): Promise<void> {
    notice = null;
    retry = null;
    erroredField = null;
    announce('', 'polite');
    const result = await call(kind === 'remove' ? 'remove_member' : 'demote_member', { user_id: member.user_id });
    if (!settle(result, () => change(kind, member))) return;
    say(kind === 'remove' ? removedMemberLine(member.display_name) : demotedMemberLine(member.display_name), '', 'polite');
    await refreshAfterWrite();
    // The row that held the control is gone or changed, so focus goes to
    // the page rather than to nothing.
    await tick();
    document.getElementById('main')?.focus();
  }

  function confirm(): void {
    const chosen = pending;
    // A confirmed change moves focus itself, once it has settled.
    opener = null;
    confirmDialog?.close();
    if (chosen) void change(chosen.kind, chosen.member);
  }
</script>

<svelte:head>
  <title>{pageTitle(status === 'refused' ? TITLES.notFound : status === 'signedOut' ? TITLES.signedOut : TITLES.members, siteName)}</title>
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
    <h1 class="display page-title text-4xl">{TITLES.members}</h1>
    <div class="alert alert-warning">{unreachableMessage}</div>
    <button type="button" class="btn btn-primary" onclick={loadRoster}>{ROSTER_RETRY}</button>
  {:else if status === 'ok'}
    <h1 class="display page-title text-4xl">{TITLES.members}</h1>

    <p>{members.length === 0 ? ROSTER_EMPTY_STATE : memberCountLine(members.length)}</p>
    {#if members.length > 0}
      <ul id="roster-list" class="list" role="list">
        {#each members as member (member.user_id)}
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
                     dangling "·". The gap on one line is the flex gap, not a text space. -->
                <span class="whitespace-nowrap">{member.address}<span aria-hidden="true">&nbsp;·</span></span><span class="whitespace-nowrap"><span class="sr-only">{ADD_MEMBER_LABELS.username}{' '}</span>{member.username}</span>
              </div>
            </div>
            <div class="list-col-wrap flex flex-wrap gap-2">
              {#if member.admin}
                <button
                  type="button"
                  class="btn btn-sm min-h-11"
                  aria-label={demoteControlLabel(member.display_name)}
                  onclick={(event) => ask('demote', member, event)}
                >{DEMOTE_CONTROL}</button>
              {/if}
              <button
                type="button"
                class="btn btn-sm min-h-11"
                aria-label={removeControlLabel(member.display_name)}
                onclick={(event) => ask('remove', member, event)}
              >{REMOVE_CONTROL}</button>
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
    {#if notice}
      <div id="roster-outcome" class="alert {notice.alertClass}">
        <div class="flex flex-col gap-2">
          <span>{notice.text}</span>
          {#if notice.reauthenticate}
            <div class="flex flex-wrap gap-2">
              <a href={resolve('login')} target="_blank" rel="noopener" class="btn btn-sm min-h-11">{REAUTH_SIGN_IN}</a>
              <button type="button" class="btn btn-sm btn-primary min-h-11" onclick={() => void retry?.()}>{ROSTER_RETRY}</button>
            </div>
          {/if}
        </div>
      </div>
    {/if}

    <dialog class="modal" bind:this={confirmDialog} onclose={closed} aria-labelledby="confirm-question">
      <div class="modal-box">
        {#if pending}
          <p id="confirm-question">
            {pending.kind === 'remove' ? confirmRemoveLine(pending.member.display_name) : confirmDemoteLine(pending.member.display_name)}
          </p>
        {/if}
        <div class="modal-action">
          <!-- Cancel opens focused, the safe default for a destructive confirm. -->
          <!-- svelte-ignore a11y_autofocus -->
          <button type="button" class="btn min-h-11" autofocus onclick={() => confirmDialog?.close()}>{CONFIRM_CANCEL}</button>
          <button type="button" class="btn btn-error min-h-11" onclick={confirm}>
            {pending?.kind === 'demote' ? CONFIRM_DEMOTE : CONFIRM_REMOVE}
          </button>
        </div>
      </div>
    </dialog>
  {/if}
</main>
