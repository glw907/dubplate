<!--
@component
The server settings page: every key the server registers, with its
description, its sensitivity, and when a change to it takes effect. A key the
store holds is a field the admin edits and resets. A key only the
server's environment file sets, or one the environment file sets to a
value, is listed beneath with no control, since only that file changes
it. A restart-applied key names its stored value as the one the next
start applies.

Every call goes to `/api/v1` (`#lib/api/v1.ts`). A non-admin meets
`forbidden` and sees the not-found state; a signed-out answer shows its
own state with a link to sign in. A write the server refuses for want of
a recent sign-in keeps the edit in its field and offers a sign-in link
that opens in a new tab, so the edit survives in this one, and a control
to retry it.
-->
<script lang="ts">
  import { onMount, tick } from 'svelte';
  import { call, failureAlertClass, failureText, isSignedOut, loadList, needsReauthentication } from '#lib/api/v1.js';
  import type { ApiResult, SettingEntry, SettingWriteResponse } from '#lib/api/v1.js';
  import { announce } from '#lib/announcer.svelte.js';
  import AdminPage from '#lib/components/AdminPage.svelte';
  import type { Status } from '#lib/components/AdminPage.svelte';
  import WriteNotice from '#lib/components/WriteNotice.svelte';
  import {
    ENVIRONMENT_ONLY_LINE,
    LIST_REFRESH_FAILED,
    SETTINGS_EDITABLE_HEADING,
    SETTINGS_ENVIRONMENT_HEADING,
    SETTINGS_INTRO,
    SETTING_APPLY,
    SETTING_INVALID_VALUE,
    SETTING_SENSITIVITIES,
    SETTING_RESET,
    SETTING_SAVE,
    SETTING_SOURCES,
    TITLES,
    environmentValueLine,
    nextStartLine,
    resetSettingLabel,
    saveSettingLabel,
    settingResetLine,
    settingSavedLine
  } from '#lib/copy/deck.js';

  /** What the alert beneath the list says about the last write. */
  interface Notice {
    text: string;
    alertClass: string;
    politeness: 'polite' | 'assertive';
    /** The key a refusal names, so its field carries the invalid state; null for a success. */
    key: string | null;
    /** Set when the write needs a fresh sign-in: the alert offers the sign-in link and a retry. */
    reauthenticate: boolean;
  }

  let status = $state<Status>('loading');
  // Replaced whole on every fetch and never mutated, so it skips the deep proxy.
  let entries = $state.raw<SettingEntry[]>([]);
  let unreachableMessage = $state('');
  let refreshFailed = $state(false);
  // What the admin has typed into each field, by key, kept across a
  // refused write so the edit survives a sign-in.
  let drafts = $state<Record<string, string>>({});
  let notice = $state<Notice | null>(null);
  // The write the alert's retry control runs again.
  let retry = $state<(() => Promise<void>) | null>(null);
  let busy = $state(false);

  /** A key the store holds and the environment file does not set: a field to edit. */
  const editable = $derived(entries.filter((e) => e.home === 'store' && e.source !== 'environment'));
  /** Every other key: listed, with no control. */
  const environmental = $derived(entries.filter((e) => !(e.home === 'store' && e.source !== 'environment')));

  function fieldId(name: string): string {
    return `setting-${name}`;
  }

  /**
   * Takes the list into the page. Each stored value seeds its field,
   * except where `keepDrafts` is set and the admin has typed something
   * other than the value the field last held: that edit stays.
   */
  function take(list: SettingEntry[], keepDrafts: boolean): void {
    const prior = new Map(entries.map((e) => [e.key, e.value]));
    entries = list;
    const next: Record<string, string> = {};
    for (const e of list) {
      if (e.value === undefined) continue;
      const edited = keepDrafts && e.key in drafts && drafts[e.key] !== prior.get(e.key);
      next[e.key] = edited ? drafts[e.key] : e.value;
    }
    drafts = next;
  }

  async function load(): Promise<void> {
    const result = await loadList<SettingEntry>('list_settings', 'settings');
    if (result.ok) {
      take(result.entries, false);
      status = 'ok';
      return;
    }
    status = result.state;
    unreachableMessage = result.message;
    if (result.state === 'unreachable') announce(result.message, 'assertive');
  }

  // A write's own refresh: it never turns the page's `status` away from
  // `'ok'`, so a refresh failure after a write still shows the list. The
  // key just written then shows what the server now holds.
  async function refreshAfterWrite(key: string): Promise<void> {
    const result = await loadList<SettingEntry>('list_settings', 'settings');
    if (!result.ok) {
      refreshFailed = true;
      announce(LIST_REFRESH_FAILED, 'polite');
      return;
    }
    refreshFailed = false;
    // The field just written shows what the server now holds, an edit or not.
    const written = result.entries.find((e) => e.key === key);
    take(result.entries, true);
    if (written?.value !== undefined) drafts[key] = written.value;
  }

  onMount(() => {
    void load();
  });

  function say(n: Notice): void {
    notice = n;
    announce(n.text, n.politeness);
  }

  /** Shows why a write did not go through, or replaces the page when the session is gone. */
  function refuse(result: Exclude<ApiResult<unknown>, { kind: 'ok' }>, key: string, again: () => Promise<void>): void {
    if (isSignedOut(result)) {
      status = 'signedOut';
      return;
    }
    const reauthenticate = needsReauthentication(result);
    if (reauthenticate) retry = again;
    // Only a refusal of the value itself marks the field: a network or
    // server failure says nothing about what was typed. A value the
    // server cannot take answers `invalid_request` with no detail, which
    // the shared copy blames on the page, so this page words it itself.
    const refusedValue = result.kind === 'refused' && (result.code === 'invalid_request' || result.detail === 'environment_sets');
    const unparsedValue = result.kind === 'refused' && result.code === 'invalid_request' && result.detail === '';
    say({
      text: unparsedValue ? SETTING_INVALID_VALUE : failureText(result),
      alertClass: failureAlertClass(result),
      politeness: 'assertive',
      key: refusedValue ? key : null,
      reauthenticate
    });
    // The alert sits above the list, far from the field a long page
    // puts under the pointer, so the refused field takes the focus.
    if (refusedValue) void tick().then(() => document.getElementById(fieldId(key))?.focus());
  }

  async function write(
    key: string,
    operation: 'set_setting' | 'reset_setting',
    request: object,
    again: () => Promise<void>
  ): Promise<void> {
    if (busy) return;
    busy = true;
    notice = null;
    retry = null;
    announce('', 'polite');
    try {
      const result = await call<SettingWriteResponse>(operation, request);
      if (result.kind !== 'ok') {
        refuse(result, key, again);
        return;
      }
      const apply = SETTING_APPLY[result.value.apply] ?? '';
      say({
        text: operation === 'set_setting' ? settingSavedLine(key, apply) : settingResetLine(key, apply),
        alertClass: '',
        politeness: 'polite',
        key: null,
        reauthenticate: false
      });
      await refreshAfterWrite(key);
    } finally {
      busy = false;
    }
    // A control that removed itself (the retry button, or a Reset whose
    // row is back at its default) leaves focus on the page body: the
    // key's field takes it instead.
    await tick();
    if (document.activeElement === document.body || document.activeElement === null) {
      document.getElementById(fieldId(key))?.focus();
    }
  }

  async function save(key: string): Promise<void> {
    await write(key, 'set_setting', { key, value: drafts[key] ?? '' }, () => save(key));
  }

  async function reset(key: string): Promise<void> {
    await write(key, 'reset_setting', { key }, () => reset(key));
  }

  /**
   * When a change takes effect. A restart-applied key with a stored
   * value names it as the one the next start applies; with no stored
   * value, only the mode is true.
   */
  function applyLine(e: SettingEntry): string {
    if (e.apply === 'restart' && e.value !== undefined && e.source !== 'default') return nextStartLine(e.value);
    return SETTING_APPLY[e.apply] ?? '';
  }

  function sourceLine(e: SettingEntry): string {
    return e.source !== undefined && Object.hasOwn(SETTING_SOURCES, e.source) ? SETTING_SOURCES[e.source] : '';
  }
</script>

<AdminPage title={TITLES.serverSettings} {status} {unreachableMessage} retry={load}>
  <h1 class="display page-title text-4xl">{TITLES.serverSettings}</h1>
  <p>{SETTINGS_INTRO}</p>

  {#if refreshFailed}
    <div class="alert alert-warning">{LIST_REFRESH_FAILED}</div>
  {/if}
  <WriteNotice {notice} {retry} id="settings-outcome" />

  <h2 class="display text-2xl">{SETTINGS_EDITABLE_HEADING}</h2>
  <ul id="settings-editable" class="list" role="list">
    {#each editable as e (e.key)}
      <li class="list-row">
        <div class="list-col-grow flow">
          <div class="flex flex-wrap items-center gap-2">
            <label class="font-mono text-xs font-medium" for={fieldId(e.key)}>{e.key}</label>
            <span class="badge">{SETTING_SENSITIVITIES[e.sensitivity] ?? e.sensitivity}</span>
          </div>
          <p class="first-letter:uppercase">{e.description}</p>
          <p class="text-sm">
            {sourceLine(e)} {applyLine(e)}
          </p>
        </div>
        <form
          class="list-col-wrap flex flex-wrap items-center gap-2"
          onsubmit={(event) => {
            event.preventDefault();
            void save(e.key);
          }}
        >
          <input
            id={fieldId(e.key)}
            name={e.key}
            type="text"
            autocomplete="off"
            data-1p-ignore
            data-lpignore="true"
            class="input {notice?.key === e.key ? 'input-error' : ''}"
            aria-invalid={notice?.key === e.key ? 'true' : undefined}
            aria-describedby={notice?.key === e.key ? 'settings-outcome' : undefined}
            bind:value={drafts[e.key]}
          />
          <button type="submit" class="btn btn-primary" aria-label={saveSettingLabel(e.key)} aria-disabled={busy || undefined}>{SETTING_SAVE}</button>
          {#if e.source === 'admin'}
            <button
              type="button"
              class="btn"
              aria-label={resetSettingLabel(e.key)}
              aria-disabled={busy || undefined}
              onclick={() => void reset(e.key)}
            >{SETTING_RESET}</button>
          {/if}
        </form>
      </li>
    {/each}
  </ul>

  <h2 class="display text-2xl">{SETTINGS_ENVIRONMENT_HEADING}</h2>
  <ul id="settings-environment" class="list" role="list">
    {#each environmental as e (e.key)}
      <li class="list-row">
        <div class="list-col-grow flow">
          <div class="flex flex-wrap items-center gap-2">
            <span class="font-mono text-xs font-medium">{e.key}</span>
            <span class="badge">{SETTING_SENSITIVITIES[e.sensitivity] ?? e.sensitivity}</span>
          </div>
          <p class="first-letter:uppercase">{e.description}</p>
          <p class="text-sm">{e.value !== undefined ? environmentValueLine(e.value) : ENVIRONMENT_ONLY_LINE}</p>
        </div>
      </li>
    {/each}
  </ul>
</AdminPage>
