<!--
@component
The server settings page: every key the server registers, with its
description, its mark, and when a change to it takes effect. A key the
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

The signed-in header renders in every state, loading included, and
`main` is the root layout's skip-link target.
-->
<script lang="ts">
  import { onMount } from 'svelte';
  import { browser } from '$app/env';
  import { resolve } from '$app/paths';
  import { call, failureAlertClass, failureText, isSignedOut, needsReauthentication } from '#lib/api/v1.js';
  import type { ApiResult } from '#lib/api/v1.js';
  import { announce } from '#lib/announcer.svelte.js';
  import SiteHeader from '#lib/components/SiteHeader.svelte';
  import {
    API_REFUSALS,
    AUTH_OUTCOMES,
    ENVIRONMENT_ONLY_LINE,
    ERROR_PAGE,
    REAUTH_SIGN_IN,
    ROSTER_RETRY,
    SETTINGS_EDITABLE_HEADING,
    SETTINGS_ENVIRONMENT_HEADING,
    SETTINGS_INTRO,
    SETTING_APPLY,
    SETTING_MARKS,
    SETTING_RESET,
    SETTING_SAVE,
    SETTING_SOURCES,
    SIGNED_OUT_SIGN_IN,
    TITLES,
    environmentValueLine,
    nextStartLine,
    resetSettingLabel,
    saveSettingLabel,
    settingResetLine,
    settingSavedLine
  } from '#lib/copy/deck.js';
  import { pageTitle, siteNameFromDocument } from '#lib/copy/title.js';

  /** One registered key as `list_settings` answers it. */
  interface Entry {
    name: string;
    home: string;
    description: string;
    apply: string;
    mark: string;
    value?: string;
    source?: string;
    seed_value?: string;
  }

  /** What `set_setting` and `reset_setting` answer with. */
  interface Written {
    key: string;
    apply: string;
    source: string;
  }

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

  const siteName = browser ? siteNameFromDocument(document) : '';

  type Status = 'loading' | 'ok' | 'refused' | 'signedOut' | 'unreachable';

  let status = $state<Status>('loading');
  // The list skips Svelte's deep proxy: nothing here reads or writes
  // into an entry through a derived value.
  let entries = $state.raw<Entry[]>([]);
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

  async function fetchEntries(): Promise<
    { ok: true; entries: Entry[] } | { ok: false; state: 'refused' | 'signedOut' | 'unreachable'; message: string }
  > {
    const result = await call<{ settings: unknown }>('list_settings');
    if (result.kind === 'ok') {
      if (!Array.isArray(result.value.settings)) return { ok: false, state: 'unreachable', message: AUTH_OUTCOMES.failed };
      return { ok: true, entries: result.value.settings as Entry[] };
    }
    if (result.kind === 'refused' && result.code === 'forbidden') return { ok: false, state: 'refused', message: '' };
    if (isSignedOut(result)) return { ok: false, state: 'signedOut', message: '' };
    return { ok: false, state: 'unreachable', message: failureText(result) };
  }

  /** Takes the list into the page; each stored value seeds its field unless the admin is mid-edit on it. */
  function take(list: Entry[], keepDrafts: boolean): void {
    entries = list;
    const next: Record<string, string> = {};
    for (const e of list) {
      if (e.value === undefined) continue;
      next[e.name] = keepDrafts && e.name in drafts ? drafts[e.name] : e.value;
    }
    drafts = next;
  }

  async function load(): Promise<void> {
    const result = await fetchEntries();
    if (result.ok) {
      take(result.entries, false);
      status = 'ok';
      return;
    }
    status = result.state;
    unreachableMessage = result.message;
  }

  // A write's own refresh: it never turns the page's `status` away from
  // `'ok'`, so a refresh failure after a write still shows the list. The
  // key just written then shows what the server now holds.
  async function refreshAfterWrite(key: string): Promise<void> {
    const result = await fetchEntries();
    if (!result.ok) {
      refreshFailed = true;
      return;
    }
    refreshFailed = false;
    take(result.entries, true);
    const written = result.entries.find((e) => e.name === key);
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
    say({
      text: failureText(result),
      alertClass: failureAlertClass(result),
      politeness: 'assertive',
      key: reauthenticate ? null : key,
      reauthenticate
    });
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
      const result = await call<Written>(operation, request);
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
  }

  async function save(key: string): Promise<void> {
    await write(key, 'set_setting', { key, value: drafts[key] ?? '' }, () => save(key));
  }

  async function reset(key: string): Promise<void> {
    await write(key, 'reset_setting', { key }, () => reset(key));
  }

  function sourceLine(e: Entry): string {
    return e.source !== undefined && Object.hasOwn(SETTING_SOURCES, e.source) ? SETTING_SOURCES[e.source] : '';
  }
</script>

<svelte:head>
  <title>{pageTitle(status === 'refused' ? TITLES.notFound : status === 'signedOut' ? TITLES.signedOut : TITLES.serverSettings, siteName)}</title>
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
    <h1 class="display page-title text-4xl">{TITLES.serverSettings}</h1>
    <div class="alert alert-warning">{unreachableMessage}</div>
    <button type="button" class="btn btn-primary" onclick={load}>{ROSTER_RETRY}</button>
  {:else if status === 'ok'}
    <h1 class="display page-title text-4xl">{TITLES.serverSettings}</h1>
    <p>{SETTINGS_INTRO}</p>

    {#if refreshFailed}
      <div class="alert alert-warning">{AUTH_OUTCOMES.failed}</div>
    {/if}
    {#if notice}
      <div id="settings-outcome" class="alert {notice.alertClass}">
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

    <h2 class="display text-2xl">{SETTINGS_EDITABLE_HEADING}</h2>
    <ul id="settings-editable" class="list" role="list">
      {#each editable as e (e.name)}
        <li class="list-row">
          <div class="list-col-grow flow">
            <div class="flex flex-wrap items-center gap-2">
              <label class="font-mono text-sm font-medium" for={fieldId(e.name)}>{e.name}</label>
              <span class="badge">{SETTING_MARKS[e.mark] ?? e.mark}</span>
            </div>
            <p>{e.description}</p>
            <p class="text-sm opacity-70">
              {sourceLine(e)} {SETTING_APPLY[e.apply] ?? ''}
              {#if e.apply === 'restart' && e.value !== undefined}{nextStartLine(e.value)}{/if}
            </p>
          </div>
          <form
            class="list-col-wrap flex flex-wrap items-center gap-2"
            onsubmit={(event) => {
              event.preventDefault();
              void save(e.name);
            }}
          >
            <input
              id={fieldId(e.name)}
              name={e.name}
              type="text"
              autocomplete="off"
              data-1p-ignore
              data-lpignore="true"
              class="input {notice?.key === e.name ? 'input-error' : ''}"
              aria-invalid={notice?.key === e.name ? 'true' : undefined}
              aria-describedby={notice?.key === e.name ? 'settings-outcome' : undefined}
              bind:value={drafts[e.name]}
            />
            <button type="submit" class="btn btn-primary btn-sm min-h-11" aria-label={saveSettingLabel(e.name)} aria-disabled={busy || undefined}>{SETTING_SAVE}</button>
            {#if e.source === 'admin'}
              <button
                type="button"
                class="btn btn-sm min-h-11"
                aria-label={resetSettingLabel(e.name)}
                aria-disabled={busy || undefined}
                onclick={() => void reset(e.name)}
              >{SETTING_RESET}</button>
            {/if}
          </form>
        </li>
      {/each}
    </ul>

    <h2 class="display text-2xl">{SETTINGS_ENVIRONMENT_HEADING}</h2>
    <ul id="settings-environment" class="list" role="list">
      {#each environmental as e (e.name)}
        <li class="list-row">
          <div class="list-col-grow flow">
            <div class="flex flex-wrap items-center gap-2">
              <span class="font-mono text-sm font-medium">{e.name}</span>
              <span class="badge">{SETTING_MARKS[e.mark] ?? e.mark}</span>
            </div>
            <p>{e.description}</p>
            <p class="text-sm opacity-70">{e.value !== undefined ? environmentValueLine(e.value) : ENVIRONMENT_ONLY_LINE}</p>
          </div>
        </li>
      {/each}
    </ul>
  {/if}
</main>
