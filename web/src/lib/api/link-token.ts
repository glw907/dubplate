/**
 * The mailed sign-in link's token, held in this module's memory and
 * nowhere else. The client hook's `init` moves it off the address bar
 * before kit's router first reads the URL, so kit's own `page.url`
 * never carries it. That matters because kit writes `page.url` back
 * into the address bar and history on its own schedule: its focus
 * reset on a link with a fragment (`#main`, say) replaces the history
 * entry with `page.url` after the page has rendered, which would put
 * the token back if kit had ever seen it.
 */
import { resolve } from '$app/paths';

/** The query parameter `internal/auth`'s `LinkURL` carries the token in. */
const TOKEN_PARAM = 't';

let held: string | null = null;

/**
 * Moves the token from the current URL into memory, then replaces the
 * history entry with the same URL less its query, through the native
 * `history.replaceState`. Runs only on the consumption page's own
 * path, so a `t` on any other page is left alone.
 */
export function captureLinkToken(): void {
  if (location.pathname !== resolve('/login/link')) return;
  const params = new URLSearchParams(location.search);
  if (!params.has(TOKEN_PARAM)) return;
  held = params.get(TOKEN_PARAM) || null;
  history.replaceState(history.state, '', location.pathname + location.hash);
}

/**
 * Returns the captured token and forgets it, so a second reader (a
 * remount, a later client navigation back to the page) finds none.
 * `null` when the page loaded with no token at all.
 */
export function takeLinkToken(): string | null {
  const token = held;
  held = null;
  return token;
}
