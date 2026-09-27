<!--
@component
The signed-in header (visual identity spec, "Themes"): a neutral bar
carrying the mark and wordmark as a link back to the members page, the
theme toggle, and the sign-out control. Every header foreground reads
`--color-neutral-content`, since `base-content` measures 1.00:1 on the
`label` bar. The toggle is a stock DaisyUI `swap`, driven only by its
hidden checkbox; choosing a state writes the theme cookie, stamps
`data-theme`, and updates the one `theme-color` meta the document may
carry. Its icons are Phosphor's `sun` and `moon`, inlined as SVG
(`PHOSPHOR.md`).
-->
<script lang="ts">
  import { browser } from '$app/environment';
  import { resolve } from '$app/paths';
  import Mark from './Mark.svelte';
  import { SIGN_OUT_CONTROL, THEME_TOGGLE_LABEL } from '$lib/copy/deck';

  const THEME_COOKIE = '__Host-dubplate_theme';
  const SIGN_OUT_PATH = 'sign-out';

  // The stamp a page without a theme cookie carries: the composed
  // "dubplate-system" name resolves through the device's own
  // preference, since neither cookie palette applies.
  const SYSTEM_STAMP = 'dubplate-system';
  const NIGHT_STAMP = 'dubplate-label-night';

  function currentlyDark(): boolean {
    if (!browser) return false;
    const stamp = document.documentElement.dataset.theme;
    if (stamp === NIGHT_STAMP) return true;
    if (stamp === SYSTEM_STAMP) return window.matchMedia('(prefers-color-scheme: dark)').matches;
    return false;
  }

  let dark = $state(currentlyDark());

  // Leaves exactly one theme-color meta, without `media`, at the
  // computed --color-neutral: the shape the server stamps for a cookie
  // (R-5ciii9). A document with none (a branding neutral outside
  // `themeColorRE`) is left with none.
  function restampThemeColorMeta(): void {
    const metas = document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]');
    if (metas.length === 0) return;
    const neutral = getComputedStyle(document.documentElement).getPropertyValue('--color-neutral').trim();
    metas.forEach((meta, i) => {
      if (i === 0) {
        meta.removeAttribute('media');
        meta.content = neutral;
      } else {
        meta.remove();
      }
    });
  }

  function onToggleChange(event: Event): void {
    const checked = (event.currentTarget as HTMLInputElement).checked;
    dark = checked;
    const palette = checked ? 'label-night' : 'label';
    document.cookie = `${THEME_COOKIE}=${palette}; Path=/; Secure; SameSite=Lax; Max-Age=31536000`;
    document.documentElement.dataset.theme = `dubplate-${palette}`;
    restampThemeColorMeta();
  }

  // Sign-out clears only this browser's cookie (the deck's own bound):
  // the destination page reads the cleared session fresh either way,
  // so a network failure here still returns to sign-in.
  async function signOut(): Promise<void> {
    try {
      await fetch(resolve('/') + SIGN_OUT_PATH, { method: 'POST', body: new URLSearchParams() });
    } catch {
      // Falls through to the navigation below regardless.
    }
    window.location.assign(resolve('/login'));
  }
</script>

<header class="site-header bg-neutral text-neutral-content">
  <div class="content-measure flex items-center justify-between gap-4 p-2">
    <a href={resolve('/members')} class="flex min-h-11 items-center gap-2 text-neutral-content">
      <Mark size={32} />
      <span class="display text-lg">dubplate</span>
    </a>
    <nav class="flex items-center gap-2">
      <button type="button" class="btn btn-ghost min-h-11 text-neutral-content" onclick={signOut}>{SIGN_OUT_CONTROL}</button
      >
      <label class="swap h-11 w-11">
        <input type="checkbox" aria-label={THEME_TOGGLE_LABEL} checked={dark} onchange={onToggleChange} />
        <span class="swap-off text-neutral-content" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 256 256" fill="currentColor"
            ><path
              d="M116,36V20a12,12,0,0,1,24,0V36a12,12,0,0,1-24,0Zm80,92a68,68,0,1,1-68-68A68.07,68.07,0,0,1,196,128Zm-24,0a44,44,0,1,0-44,44A44.05,44.05,0,0,0,172,128ZM51.51,68.49a12,12,0,1,0,17-17l-12-12a12,12,0,0,0-17,17Zm0,119-12,12a12,12,0,0,0,17,17l12-12a12,12,0,1,0-17-17ZM196,72a12,12,0,0,0,8.49-3.51l12-12a12,12,0,0,0-17-17l-12,12A12,12,0,0,0,196,72Zm8.49,115.51a12,12,0,0,0-17,17l12,12a12,12,0,0,0,17-17ZM48,128a12,12,0,0,0-12-12H20a12,12,0,0,0,0,24H36A12,12,0,0,0,48,128Zm80,80a12,12,0,0,0-12,12v16a12,12,0,0,0,24,0V220A12,12,0,0,0,128,208Zm108-92H220a12,12,0,0,0,0,24h16a12,12,0,0,0,0-24Z"
            /></svg
          >
        </span>
        <span class="swap-on text-neutral-content" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 256 256" fill="currentColor"
            ><path
              d="M236.37,139.4a12,12,0,0,0-12-3A84.07,84.07,0,0,1,119.6,31.59a12,12,0,0,0-15-15A108.86,108.86,0,0,0,49.69,55.07,108,108,0,0,0,136,228a107.09,107.09,0,0,0,64.93-21.69,108.86,108.86,0,0,0,38.44-54.94A12,12,0,0,0,236.37,139.4Zm-49.88,47.74A84,84,0,0,1,68.86,69.51,84.93,84.93,0,0,1,92.27,48.29Q92,52.13,92,56A108.12,108.12,0,0,0,200,164q3.87,0,7.71-.27A84.79,84.79,0,0,1,186.49,187.14Z"
            /></svg
          >
        </span>
      </label>
    </nav>
  </div>
</header>
