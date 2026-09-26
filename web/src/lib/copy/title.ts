/**
 * Composes a page `<title>` from one of the deck's state strings and
 * the site name the server stamps into every document. Reads the meta
 * as text, so `deck.ts` stays import-free: this is the module's one
 * caller under Node's type-stripping loader, and the pages' other
 * caller through `$lib/copy/title`.
 */

/** The meta the document stamps the group's site name into (`internal/serve/document.go`). */
const SITE_NAME_META = 'dubplate-site-name';

/** The `<title>` text for `state`, given the served site name. */
export function pageTitle(state: string, siteName: string): string {
  return `${state} · ${siteName}`;
}

/**
 * Reads the site name from `doc`'s `dubplate-site-name` meta as text.
 * Returns `''` when the meta is absent, so a caller decides for itself
 * how to degrade rather than this module throwing on a document that
 * has not stamped it yet.
 */
export function siteNameFromDocument(doc: Document): string {
  return doc.querySelector(`meta[name="${SITE_NAME_META}"]`)?.getAttribute('content') ?? '';
}
