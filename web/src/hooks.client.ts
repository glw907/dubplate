import type { ClientInit } from '@sveltejs/kit';
import { captureLinkToken } from '$lib/api/link-token';

// Kit runs `init` before it resolves the first URL, so the token is off
// the address bar before `page.url` is ever built from it.
export const init: ClientInit = () => {
  captureLinkToken();
};
