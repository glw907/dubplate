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
  <div class="flex items-center justify-between gap-4 p-2">
    <a href={resolve('/members')} class="flex min-h-11 items-center gap-2 text-neutral-content">
      <Mark size={32} />
      <span class="display text-lg">dubplate</span>
    </a>
    <nav class="flex items-center gap-2">
      <label class="swap flex h-11 w-11 items-center justify-center">
        <input type="checkbox" aria-label={THEME_TOGGLE_LABEL} checked={dark} onchange={onToggleChange} />
        <span class="swap-off text-neutral-content" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 256 256" fill="currentColor"
            ><path
              d="M120,40V16a8,8,0,0,1,16,0V40a8,8,0,0,1-16,0Zm72,88a64,64,0,1,1-64-64A64.07,64.07,0,0,1,192,128Zm-16,0a48,48,0,1,0-48,48A48.05,48.05,0,0,0,176,128ZM58.34,69.66A8,8,0,0,0,69.66,58.34l-16-16A8,8,0,0,0,42.34,53.66Zm0,116.68-16,16a8,8,0,0,0,11.32,11.32l16-16a8,8,0,0,0-11.32-11.32ZM192,72a8,8,0,0,0,5.66-2.34l16-16a8,8,0,0,0-11.32-11.32l-16,16A8,8,0,0,0,192,72Zm5.66,114.34a8,8,0,0,0-11.32,11.32l16,16a8,8,0,0,0,11.32-11.32ZM48,128a8,8,0,0,0-8-8H16a8,8,0,0,0,0,16H40A8,8,0,0,0,48,128Zm80,80a8,8,0,0,0-8,8v24a8,8,0,0,0,16,0V216A8,8,0,0,0,128,208Zm112-88H216a8,8,0,0,0,0,16h24a8,8,0,0,0,0-16Z"
            /></svg
          >
        </span>
        <span class="swap-on text-neutral-content" aria-hidden="true">
          <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 256 256" fill="currentColor"
            ><path
              d="M233.54,142.23a8,8,0,0,0-8-2,88.08,88.08,0,0,1-109.8-109.8,8,8,0,0,0-10-10,104.84,104.84,0,0,0-52.91,37A104,104,0,0,0,136,224a103.09,103.09,0,0,0,62.52-20.88,104.84,104.84,0,0,0,37-52.91A8,8,0,0,0,233.54,142.23ZM188.9,190.34A88,88,0,0,1,65.66,67.11a89,89,0,0,1,31.4-26A106,106,0,0,0,96,56,104.11,104.11,0,0,0,200,160a106,106,0,0,0,14.92-1.06A89,89,0,0,1,188.9,190.34Z"
            /></svg
          >
        </span>
      </label>
      <button type="button" class="btn btn-ghost min-h-11 text-neutral-content" onclick={signOut}>{SIGN_OUT_CONTROL}</button
      >
    </nav>
  </div>
</header>
