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
-->
<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { resolve } from '$app/paths';
  import { call, failureAlertClass, failureText, isSignedOut, loadList, needsReauthentication } from '#lib/api/v1.js';
  import type { ApiResult, Member } from '#lib/api/v1.js';
  import { announce } from '#lib/announcer.svelte.js';
  import AdminPage from '#lib/components/AdminPage.svelte';
  import type { Status } from '#lib/components/AdminPage.svelte';
  import WriteNotice from '#lib/components/WriteNotice.svelte';
  import {
    ADD_MEMBER_HINTS,
    ADD_MEMBER_LABELS,
    ADD_MEMBER_LEGEND,
    ADD_MEMBER_SUBMIT,
    ADMIN_BADGE_LABEL,
    CONFIRM_CANCEL,
    CONFIRM_DEMOTE,
    CONFIRM_REMOVE,
    DEMOTE_CONTROL,
    LIST_REFRESH_FAILED,
    REMOVE_CONTROL,
    ROSTER_EMPTY_STATE,
    SERVER_SETTINGS_LINK,
    TITLES,
    addedMemberLine,
    confirmDemoteLine,
    confirmRemoveLine,
    demoteControlLabel,
    demotedMemberLine,
    memberCountLine,
    removeControlLabel,
    removedMemberLine
  } from '#lib/copy/deck.js';

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

  const FIELD_ID: Record<Field, string> = {
    address: 'add-address',
    display_name: 'add-display-name',
    username: 'add-username'
  };

  const OUTCOME_ID = 'roster-outcome';
  const OUTCOME_TEXT_ID = 'roster-outcome-text';

  /** What the alert beneath the form says about the last write. */
  interface Notice {
    text: string;
    alertClass: string;
    politeness: 'polite' | 'assertive';
    /** Set when the write needs a fresh sign-in: the alert offers the sign-in link and a retry. */
    reauthenticate: boolean;
  }

  let status = $state<Status>('loading');
  // Replaced whole on every fetch and never mutated, so it skips the deep proxy.
  let members = $state.raw<Member[]>([]);
  let unreachableMessage = $state('');
  // Set only by a refresh after a successful write: the page keeps
  // showing the list and the form, and reports the refresh's own
  // failure beside them instead of replacing the page.
  let refreshFailed = $state(false);

  let address = $state('');
  let displayName = $state('');
  let username = $state('');
  // True while any write is in flight, an add or a remove or a demote,
  // so a second one cannot start until the first settles.
  let submitting = $state(false);
  let notice = $state<Notice | null>(null);
  // The write the alert's retry control runs again.
  let retry = $state<(() => Promise<void>) | null>(null);
  let erroredField = $state<Field | null>(null);
  let addMemberForm = $state<HTMLFormElement | null>(null);

  // The confirm step: the change the member has asked for and not yet
  // confirmed, the modal that holds it, and the control that opened it,
  // which gets the focus back when the modal closes.
  let pending = $state<{ kind: 'remove' | 'demote'; member: Member } | null>(null);
  let confirmDialog = $state<HTMLDialogElement | null>(null);
  let opener: HTMLElement | null = null;

  function clearFieldError(field: Field): void {
    if (erroredField === field) erroredField = null;
  }

  async function loadRoster(): Promise<void> {
    const result = await loadList<Member>('list_members', 'members');
    if (result.ok) {
      members = result.entries;
      status = 'ok';
      return;
    }
    status = result.state;
    unreachableMessage = result.message;
    if (result.state === 'unreachable') announce(result.message, 'assertive');
  }

  // A write's own refresh: it never turns the page's `status` away from
  // `'ok'`, so a refresh failure right after a successful write still
  // shows the list and the form. The list the page already has stays on
  // screen.
  async function refreshAfterWrite(): Promise<void> {
    const result = await loadList<Member>('list_members', 'members');
    if (result.ok) {
      members = result.entries;
      refreshFailed = false;
    } else {
      refreshFailed = true;
      announce(LIST_REFRESH_FAILED, 'polite');
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
    if (result.kind === 'refused') {
      erroredField = Object.hasOwn(FIELD_FOR_DETAIL, result.detail) ? FIELD_FOR_DETAIL[result.detail] : null;
    }
    return false;
  }

  /** Moves focus to what a refused write shows: its field, else the alert's text. */
  async function focusRefusal(): Promise<void> {
    await tick();
    document.getElementById(erroredField ? FIELD_ID[erroredField] : OUTCOME_TEXT_ID)?.focus();
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
      if (result.kind !== 'ok') {
        settle(result, addMember);
        await focusRefusal();
        return;
      }
      say(addedMemberLine(result.value.member.display_name), '', 'polite');
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
      document.getElementById(FIELD_ID.address)?.focus();
    } finally {
      submitting = false;
    }
  }

  function submit(event: SubmitEvent): void {
    event.preventDefault();
    void addMember();
  }

  // The modal opens once `#confirm-question`, its accessible name, has rendered.
  async function ask(kind: 'remove' | 'demote', member: Member, event: MouseEvent): Promise<void> {
    opener = event.currentTarget as HTMLElement;
    pending = { kind, member };
    await tick();
    confirmDialog?.showModal();
  }

  /** The modal closed, by Cancel, Escape, or a confirm: nothing is pending and focus goes back. */
  function closed(): void {
    pending = null;
    opener?.focus();
  }

  async function change(kind: 'remove' | 'demote', member: Member): Promise<void> {
    if (submitting) return;
    submitting = true;
    notice = null;
    retry = null;
    erroredField = null;
    announce('', 'polite');
    try {
      const result = await call(kind === 'remove' ? 'remove_member' : 'demote_member', { user_id: member.user_id });
      if (!settle(result, () => change(kind, member))) {
        await focusRefusal();
        return;
      }
      say(kind === 'remove' ? removedMemberLine(member.display_name) : demotedMemberLine(member.display_name), '', 'polite');
      await refreshAfterWrite();
      // The row that held the control is gone or changed, so focus goes to
      // the member count, which the change just moved.
      await tick();
      document.getElementById('member-count')?.focus();
    } finally {
      submitting = false;
    }
  }

  function confirm(): void {
    const chosen = pending;
    // A confirmed change moves focus itself, once it has settled.
    opener = null;
    confirmDialog?.close();
    if (chosen) void change(chosen.kind, chosen.member);
  }
</script>

<AdminPage title={TITLES.members} {status} {unreachableMessage} retry={loadRoster}>
  <h1 class="display page-title text-4xl">{TITLES.members}</h1>
  <p><a href={resolve('server-settings')} class="link link-primary inline-flex min-h-11 items-center">{SERVER_SETTINGS_LINK}</a></p>

  <p id="member-count" tabindex="-1">{members.length === 0 ? ROSTER_EMPTY_STATE : memberCountLine(members.length)}</p>
  {#if members.length > 0}
    <ul id="roster-list" class="list" role="list">
      {#each members as member (member.user_id)}
        <li class="list-row">
          <div class="list-col-grow">
            <div class="flex items-center gap-2">
              <span class="font-medium">{member.display_name}</span>
              {#if member.admin}<span class="badge">{ADMIN_BADGE_LABEL}</span>{/if}
            </div>
            <div class="font-mono flex flex-wrap gap-x-1 text-xs">
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
                class="btn"
                aria-label={demoteControlLabel(member.display_name)}
                onclick={(event) => void ask('demote', member, event)}
              >{DEMOTE_CONTROL}</button>
            {/if}
            <button
              type="button"
              class="btn"
              aria-label={removeControlLabel(member.display_name)}
              onclick={(event) => void ask('remove', member, event)}
            >{REMOVE_CONTROL}</button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
  {#if refreshFailed}
    <div class="alert alert-warning">{LIST_REFRESH_FAILED}</div>
  {/if}

  <form onsubmit={submit} bind:this={addMemberForm} class="flow">
    <fieldset class="fieldset">
      <legend class="fieldset-legend">{ADD_MEMBER_LEGEND}</legend>

      <label class="label" for={FIELD_ID.address}>{ADD_MEMBER_LABELS.address}</label>
      <input
        id={FIELD_ID.address}
        name="address"
        type="email"
        autocomplete="off"
        data-1p-ignore
        data-lpignore="true"
        required
        class="input validator {erroredField === 'address' ? 'input-error' : ''}"
        aria-invalid={erroredField === 'address' ? 'true' : undefined}
        aria-describedby={erroredField === 'address' ? OUTCOME_ID : undefined}
        oninput={() => clearFieldError('address')}
        bind:value={address}
      />

      <label class="label" for={FIELD_ID.display_name}>{ADD_MEMBER_LABELS.displayName}</label>
      <input
        id={FIELD_ID.display_name}
        name="display_name"
        type="text"
        maxlength="200"
        required
        class="input validator {erroredField === 'display_name' ? 'input-error' : ''}"
        aria-invalid={erroredField === 'display_name' ? 'true' : undefined}
        aria-describedby={erroredField === 'display_name' ? OUTCOME_ID : undefined}
        oninput={() => clearFieldError('display_name')}
        bind:value={displayName}
      />

      <label class="label" for={FIELD_ID.username}>{ADD_MEMBER_LABELS.username}</label>
      <input
        id={FIELD_ID.username}
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
        aria-describedby={erroredField === 'username' ? `username-hint ${OUTCOME_ID}` : 'username-hint'}
        oninput={() => clearFieldError('username')}
        bind:value={username}
      />
      <p id="username-hint" class="text-sm">{ADD_MEMBER_HINTS.username}</p>
    </fieldset>
    <button type="submit" class="btn btn-primary" aria-disabled={submitting || undefined}>{ADD_MEMBER_SUBMIT}</button>
  </form>
  <WriteNotice {notice} {retry} id={OUTCOME_ID} textId={OUTCOME_TEXT_ID} />

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
        <button type="button" class="btn" autofocus onclick={() => confirmDialog?.close()}>{CONFIRM_CANCEL}</button>
        <button type="button" class="btn btn-error" onclick={confirm}>
          {pending?.kind === 'demote' ? CONFIRM_DEMOTE : CONFIRM_REMOVE}
        </button>
      </div>
    </div>
  </dialog>
</AdminPage>
