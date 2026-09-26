/**
 * The rendered leg's browser half. `cmd/dubplate-server/rendered_test.go`
 * serves the real shell on loopback, runs this script under its own
 * deadline, and judges every row against the manifest it holds: this
 * script owns the browser work only. Each row is one JSON line on stdout
 * (`id`, the fixture-state result, the property's result, a detail
 * string); a row never reports a property it could not first stand up
 * the fixture state for.
 *
 * Page-level failures are not rows. A browser that does not launch, a
 * server that does not answer, a page outside 2xx, a page whose final URL
 * differs from the one requested, a page that never boots, and any Content
 * Security Policy violation or console error each end the run with exit
 * status 1 and a message naming the state.
 *
 * Environment: `DUBPLATE_RENDERED_BASE_URL`, the served app's base (the
 * owned prefix with a trailing slash); `DUBPLATE_RENDERED_REQUEST_LOG`,
 * the URL of the Go side's log of every request the server received;
 * `DUBPLATE_RENDERED_BRANDING_BASE_URL` and
 * `DUBPLATE_RENDERED_OPERATOR_BASE_URL`, G7's two extra fixture
 * servers; and `DUBPLATE_RENDERED_BRANDING_PRIMARY` and
 * `DUBPLATE_RENDERED_OPERATOR_PRIMARY`, the `--color-primary` values
 * those two fixtures declare.
 */
import { chromium } from 'playwright';
import type { Browser, BrowserContext, Page } from 'playwright';
import { decodePNG, pixelHex, pixelRGBA } from './png.ts';
import {
  ADD_MEMBER_LABELS,
  AUTH_OUTCOMES,
  CROSS_ORIGIN_REFUSED,
  INTERSTITIAL_CONTROL,
  NETWORK_FAILURE,
  TITLES,
  unknownAddressLine
} from '../../src/lib/copy/deck.ts';
import { pageTitle } from '../../src/lib/copy/title.ts';

const INSTALL_HINT = 'npx --prefix web playwright install chromium';

/** How long a hydrated page has to show its declared `h1`. */
const BOOT_TIMEOUT_MS = 10_000;

const THEME_COOKIE = '__Host-dubplate_theme';

/** A state that ends the run: the message names it. */
class Refusal extends Error {}

interface Row {
  id: string;
  fixture: boolean;
  property: boolean;
  detail: string;
}

interface Violation {
  directive: string;
  blocked: string;
}

/** One page load, and every problem the browser reported for it. */
interface Visit {
  url: string;
  page: Page;
  consoleErrors: string[];
  /** Read before the page's context closes; empty for a JavaScript-disabled page. */
  violations: Violation[];
  /**
   * The one URL the page loads on purpose to be blocked by `img-src`:
   * the final sweep forgives that violation and its console report,
   * and nothing else.
   */
  expectedBlock: string | null;
  /**
   * True on a visit that deliberately posts to `request-link` or
   * `consume-link` and reads a non-2xx answer: Chromium logs "Failed
   * to load resource: the server responded with a status of NNN" for
   * any such fetch response regardless of whether the page's own code
   * handles it, so the auth rows' many intentional 4xx/5xx answers
   * would otherwise each end the whole run.
   */
  expectsAuthErrorStatus: boolean;
}

/**
 * A document the unauthenticated fixture serves, and the `h1` text it
 * shows once the client boots: `ssr = false` leaves the prerendered body
 * empty, so the heading is the proof that hydration ran.
 */
interface ServedPage {
  name: string;
  path: string;
  heading: string;
}

// Populated in `main`, once the fixture site name is in hand: the
// login page's own heading is the group's name, and the interstitial's
// shared boot heading is the deck's own refused-state title, since
// every row that opens it with no token lands there.
const PAGES: ServedPage[] = [
  { name: 'login', path: 'login', heading: '' },
  { name: 'login-link', path: 'login/link', heading: '' }
];

/**
 * First paint per theme cookie and emulated preference. Every expected
 * pixel is base-100 of the palette the stamp should select, quoted from
 * docs/superpowers/specs/2026-09-24-visual-identity-design.md: `label`
 * `#f4ead3` and `label-night` `#1c1916` ("Themes"). With no cookie the
 * `system` stamp follows the preference; a cookie pins its palette
 * under either preference.
 */
const FIRST_PAINT: {
  cookie: 'label' | 'label-night' | null;
  scheme: 'light' | 'dark';
  stamp: string;
  paint: string;
}[] = [
  { cookie: null, scheme: 'light', stamp: 'dubplate-system', paint: '#f4ead3' },
  { cookie: null, scheme: 'dark', stamp: 'dubplate-system', paint: '#1c1916' },
  { cookie: 'label', scheme: 'light', stamp: 'dubplate-label', paint: '#f4ead3' },
  { cookie: 'label', scheme: 'dark', stamp: 'dubplate-label', paint: '#f4ead3' },
  { cookie: 'label-night', scheme: 'light', stamp: 'dubplate-label-night', paint: '#1c1916' },
  { cookie: 'label-night', scheme: 'dark', stamp: 'dubplate-label-night', paint: '#1c1916' }
];

// Records every securitypolicyviolation event from the document's first
// script onward. An init script runs before any page script and outside
// the page's CSP, so the listener cannot itself be blocked.
const VIOLATION_RECORDER = `
  window.__renderedViolations = [];
  document.addEventListener('securitypolicyviolation', (e) => {
    window.__renderedViolations.push({ directive: e.effectiveDirective, blocked: e.blockedURI });
  });
`;

/** The off-origin image the listener row injects for `img-src` to block. */
const LISTENER_PROBE = 'https://example.test/rendered-probe.png';

/** The endpoints and fixture values the Go test hands the script. */
interface Endpoints {
  base: string;
  requestLog: string;
  brandingBase: string;
  operatorBase: string;
  brandingPrimary: string;
  operatorPrimary: string;
  /** `DUBPLATE_SITE_NAME`, carrying `"`, `<`, and `&`, so a page that rendered it unescaped would leave a stray element behind. */
  siteName: string;
  /** The fixture roster admin's display name, carrying `<` the same way. */
  adminDisplayName: string;
  /** A real minted link for the fixture admin, spent by the token-ordering row. */
  loginLinkOrdering: string;
  /** A second, distinct real minted link for the fixture admin, spent by the `signed_in` state-table row. */
  loginLinkSignedIn: string;
}

/**
 * The fixture roster's own addresses (`renderedEnviron`,
 * `cmd/dubplate-server/rendered_test.go`): the admin, the member, and
 * one address the roster does not carry at all.
 */
const ADMIN_ADDRESS = 'admin@example.test';
const MEMBER_ADDRESS = 'member@example.test';
const UNKNOWN_ADDRESS = 'stranger@example.test';

/**
 * An address a browser's `type="email"` constraint validation admits
 * (its local part is any run of `[a-zA-Z0-9.!#$%&'*+/=?^_\`{|}~-]`,
 * which does not forbid a repeated `.`) but Go's `net/mail.ParseAddress`
 * refuses (RFC 5322 dot-atom text forbids an empty atom between two
 * dots: `mail.ParseAddress("a..b@example.test")` answers `mail: missing
 * '@' or angle-addr`, measured against `go1.27.1`). So `invalid_address`
 * is a real answer here, not a fulfilled one.
 */
const INVALID_ADDRESS = 'a..b@example.test';

const visits: Visit[] = [];

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Refusal(`${name} is not set; the rendered leg runs only under its Go test`);
  return value;
}

/**
 * The owned prefix every request to the main fixture server must stay
 * under, and how far into its request log the last reported row read.
 * Set once `main` has the endpoints.
 */
let prefixWatch: { log: string; prefix: string; seen: number } | null = null;

/**
 * Writes `row`, failing it when any request the server logged since the
 * previous row left the owned prefix: a stray URL in any sink (a
 * `location` write, `window.open`, a dependency's own fetch) reaches the
 * log whichever code path built it, which a source scan cannot promise.
 */
async function report(row: Row): Promise<void> {
  if (prefixWatch !== null) {
    const records = await requestLog(prefixWatch.log);
    const watch = prefixWatch;
    const outside = records.slice(watch.seen).filter((r) => !r.path.startsWith(watch.prefix));
    watch.seen = records.length;
    if (outside.length > 0) {
      row = { ...row, property: false, detail: `${row.detail}; requests outside ${watch.prefix}: ${describe(outside)}` };
    }
  }
  process.stdout.write(`${JSON.stringify(row)}\n`);
}

function message(err: unknown): string {
  return err instanceof Error ? err.message : String(err);
}

async function launch(): Promise<Browser> {
  try {
    // The "chromium" channel runs Playwright's own full Chromium in new
    // headless mode. The default headless shell never requests a
    // favicon, so a favicon row under it would pass on nothing.
    return await chromium.launch({ channel: 'chromium' });
  } catch (err) {
    const text = message(err);
    if (text.includes("Executable doesn't exist at")) {
      throw new Refusal(
        `the browser did not launch: Playwright's Chromium is not installed; run \`${INSTALL_HINT}\`\n${text}`
      );
    }
    throw new Refusal(`the browser did not launch: ${text}`);
  }
}

/**
 * Opens `path` under the base URL in a new page of `context`, refusing
 * a server that does not answer, a status outside 2xx, and a final URL
 * that differs from the one requested (a redirect the harness would
 * otherwise follow and judge the wrong page by). Returns the visit and
 * the served document's bytes.
 */
async function open(
  context: BrowserContext,
  url: string,
  waitUntil: 'load' | 'domcontentloaded'
): Promise<{ visit: Visit; body: string }> {
  const page = await context.newPage();
  const visit: Visit = { url, page, consoleErrors: [], violations: [], expectedBlock: null, expectsAuthErrorStatus: false };
  page.on('console', (msg) => {
    if (msg.type() !== 'error') return;
    visit.consoleErrors.push(`${msg.text()} (${msg.location().url})`);
  });
  page.on('pageerror', (err) => visit.consoleErrors.push(err.message));
  visits.push(visit);

  let response;
  try {
    response = await page.goto(url, { waitUntil });
  } catch (err) {
    throw new Refusal(`${url}: the server does not answer: ${message(err)}`);
  }
  if (response === null) throw new Refusal(`${url}: the server does not answer: no response`);
  if (response.request().redirectedFrom() !== null || page.url() !== url) {
    throw new Refusal(`${url}: the final URL ${page.url()} differs from the one requested`);
  }
  if (response.status() < 200 || response.status() > 299) {
    throw new Refusal(`${url}: answered ${response.status()}, outside 2xx`);
  }
  return { visit, body: await response.text() };
}

/** {@link open}, then waits for the page's declared `h1` after hydration. */
async function openBooted(context: BrowserContext, base: string, served: ServedPage): Promise<Visit> {
  const { visit } = await open(context, new URL(served.path, base).href, 'load');
  const heading = visit.page.getByRole('heading', { level: 1, name: served.heading, exact: true });
  try {
    await heading.waitFor({ timeout: BOOT_TIMEOUT_MS });
  } catch {
    throw new Refusal(
      `${visit.url}: the page settled without its declared identity (h1 "${served.heading}") within ${BOOT_TIMEOUT_MS}ms; the client did not boot`
    );
  }
  return visit;
}

async function violations(page: Page): Promise<Violation[]> {
  return page.evaluate(
    () => (window as unknown as { __renderedViolations?: Violation[] }).__renderedViolations ?? []
  );
}

interface LoggedRequest {
  method: string;
  path: string;
  query: string;
  headers: Record<string, string[]>;
}

async function requestLog(url: string): Promise<LoggedRequest[]> {
  const response = await fetch(url);
  if (!response.ok) throw new Refusal(`the request log answered ${response.status}`);
  return (await response.json()) as LoggedRequest[];
}

/**
 * The records logged from index `from` on, polled until one satisfies
 * `until` or `timeoutMs` passes. A browser fetches an icon on its own
 * schedule after the page loads, so no page event marks the moment.
 */
async function logSince(
  url: string,
  from: number,
  until: (r: LoggedRequest) => boolean,
  timeoutMs: number
): Promise<LoggedRequest[]> {
  const deadline = Date.now() + timeoutMs;
  for (;;) {
    const records = (await requestLog(url)).slice(from);
    if (records.some(until) || Date.now() >= deadline) return records;
    await new Promise((resolve) => setTimeout(resolve, 100));
  }
}

/** A served document's own stylesheet links (R-5ciii6): matches the Go
 * side's `stylesheetLinkRE` (`cmd/dubplate-server/serve_test.go`). */
const STYLESHEET_LINK_RE = /<link href="([^"]*)" rel="stylesheet">/g;

/** The `<html data-theme="…">` stamp, matched the same way `firstPaintRows` reads it. */
function stampOf(body: string): string {
  return /<html\b[^>]*\sdata-theme="([^"]*)"/.exec(body)?.[1] ?? '(none)';
}

/**
 * G7 and G10's fixture state: the branding and operator stylesheets'
 * own served content, read through the served document's own tier
 * links, never a guessed path.
 */
interface TierState {
  brandingHref: string | null;
  brandingCSS: string;
  operatorHref: string | null;
  operatorCSS: string | null;
}

async function readTierState(context: BrowserContext, base: string): Promise<TierState> {
  const pageURL = new URL('login', base).href;
  const page = await (await context.request.get(pageURL)).text();
  const hrefs = [...page.matchAll(STYLESHEET_LINK_RE)].map((m) => m[1]);
  const brandingHref = hrefs.find((h) => h.endsWith('branding.css')) ?? null;
  const operatorHref = hrefs.find((h) => h.endsWith('operator.css')) ?? null;
  const brandingCSS = brandingHref
    ? await (await context.request.get(new URL(brandingHref, pageURL).href)).text()
    : '';
  const operatorCSS = operatorHref
    ? await (await context.request.get(new URL(operatorHref, pageURL).href)).text()
    : null;
  return { brandingHref, brandingCSS, operatorHref, operatorCSS };
}

/** sRGB bytes read back from a canvas: red, green, blue, alpha. */
type RGBA = [number, number, number, number];

/** The bytes a fully opaque `#rrggbb` literal paints. */
function hexRGBA(hex: string): RGBA {
  const m = /^#([0-9a-f]{2})([0-9a-f]{2})([0-9a-f]{2})$/i.exec(hex);
  if (m === null) throw new Refusal(`${hex} is not a #rrggbb literal`);
  return [parseInt(m[1], 16), parseInt(m[2], 16), parseInt(m[3], 16), 255];
}

function sameRGBA(a: RGBA, b: RGBA): boolean {
  return a.every((byte, i) => byte === b[i]);
}

/**
 * Paints `--color-primary` as the document element resolves it onto a
 * one-pixel canvas under cookie and scheme, and returns the sRGB bytes
 * read back (W12.5): the browser resolves and paints the value, so the
 * page's color syntax is never parsed here. {@link paintedColor} does
 * the painting once the resolved value is in hand.
 */
async function paintedPrimary(
  browser: Browser,
  base: string,
  cookie: 'label' | 'label-night' | null,
  scheme: 'light' | 'dark'
): Promise<RGBA> {
  const host = new URL(base).hostname;
  const context = await browser.newContext({ colorScheme: scheme });
  try {
    if (cookie !== null) {
      await context.addCookies([{ name: THEME_COOKIE, value: cookie, domain: host, path: '/', secure: true }]);
    }
    const { visit } = await open(context, new URL('login', base).href, 'load');
    const resolved = await visit.page.evaluate((): string => {
      const probe = document.createElement('div');
      probe.style.backgroundColor = 'var(--color-primary)';
      document.documentElement.append(probe);
      const resolved = getComputedStyle(probe).backgroundColor;
      probe.remove();
      return resolved;
    });
    return await paintedColor(visit.page, resolved);
  } finally {
    await context.close();
  }
}

/** Every shipped theme's own `--color-primary` (all `#c1352b`), quoted
 * from web/src/lib/themes/label.css, label-night.css, and system.css:
 * the without-branding fixture's expectation. */
const G7_UNBRANDED_PRIMARY = '#c1352b';

const G7_STATES: { id: string; cookie: 'label' | 'label-night' | null; scheme: 'light' | 'dark' }[] = [
  { id: 'label', cookie: 'label', scheme: 'light' },
  { id: 'label-night', cookie: 'label-night', scheme: 'dark' },
  { id: 'system-light', cookie: null, scheme: 'light' },
  { id: 'system-dark', cookie: null, scheme: 'dark' }
];

/**
 * G7's computed half (R-5ciii8): the with-branding and without-branding
 * fixtures' rows under every stamp and both emulated preferences, and
 * the operator fixture's one row proving tier 3 over tier 2 (R4.1).
 * Every row's fixture state is the served document's own tier links,
 * read before any color is compared.
 */
async function g7ComputedRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const targets: { label: 'with-branding' | 'without-branding'; base: string; want: string }[] = [
    { label: 'with-branding', base: endpoints.brandingBase, want: endpoints.brandingPrimary },
    { label: 'without-branding', base: endpoints.base, want: G7_UNBRANDED_PRIMARY }
  ];
  for (const target of targets) {
    const tierContext = await browser.newContext();
    let tier: TierState;
    try {
      tier = await readTierState(tierContext, target.base);
    } finally {
      await tierContext.close();
    }
    const fixture =
      target.label === 'with-branding'
        ? tier.brandingHref !== null && tier.brandingCSS.includes(`--color-primary: ${endpoints.brandingPrimary}`)
        : tier.brandingHref !== null && tier.brandingCSS.trim() === '';
    for (const state of G7_STATES) {
      const id = `g7/computed/${target.label}/${state.id}`;
      if (!fixture) {
        await report({
          id,
          fixture,
          property: false,
          detail: `branding.css href ${tier.brandingHref}; content ${JSON.stringify(tier.brandingCSS)}`
        });
        continue;
      }
      const got = await paintedPrimary(browser, target.base, state.cookie, state.scheme);
      const want = hexRGBA(target.want);
      await report({
        id,
        fixture,
        property: sameRGBA(got, want),
        detail: `painted --color-primary [${got.join(',')}] (want [${want.join(',')}], ${target.want})`
      });
    }
  }

  const opContext = await browser.newContext();
  let opTier: TierState;
  try {
    opTier = await readTierState(opContext, endpoints.operatorBase);
  } finally {
    await opContext.close();
  }
  const opFixture =
    opTier.brandingHref !== null &&
    opTier.brandingCSS.includes(`--color-primary: ${endpoints.brandingPrimary}`) &&
    opTier.operatorHref !== null &&
    opTier.operatorCSS !== null &&
    opTier.operatorCSS.includes(`--color-primary: ${endpoints.operatorPrimary}`);
  const opID = 'g7/computed/operator/label';
  if (!opFixture) {
    await report({
      id: opID,
      fixture: opFixture,
      property: false,
      detail: `branding.css ${JSON.stringify(opTier.brandingCSS)}; operator.css href ${opTier.operatorHref} content ${JSON.stringify(opTier.operatorCSS)}`
    });
  } else {
    const got = await paintedPrimary(browser, endpoints.operatorBase, 'label', 'light');
    const want = hexRGBA(endpoints.operatorPrimary);
    await report({
      id: opID,
      fixture: opFixture,
      property: sameRGBA(got, want),
      detail: `painted --color-primary [${got.join(',')}] (want [${want.join(',')}], ${endpoints.operatorPrimary}, over branding's ${endpoints.brandingPrimary})`
    });
  }
}

/**
 * The `dubplate` match block's eleven structural declarations, quoted
 * from web/src/lib/themes/dubplate.css and confirmed against the built
 * bundle's own computed values by a probe on 2026-09-25 (Chromium 153
 * through playwright@1.63.0): both composed identities share one match
 * block, so one expectation map serves both.
 */
const STRUCTURAL: Record<string, string> = {
  '--radius-selector': '.25rem',
  '--radius-field': '.25rem',
  '--radius-box': '.25rem',
  '--size-selector': '.25rem',
  '--size-field': '.28125rem',
  '--border': '2px',
  '--depth': '0',
  '--noise': '1',
  '--font-sans': 'Work Sans, ui-sans-serif, system-ui, sans-serif',
  '--font-serif': 'Alfa Slab One, ui-serif, serif',
  '--font-mono': 'IBM Plex Mono, ui-monospace, monospace'
};

/** The built-in `light` theme's own `--radius-box`, quoted from
 * daisyui@5.7.46's theme/light.css (`0.5rem`) and confirmed against
 * the built bundle's computed value by the same probe: `light`'s
 * minified value has no leading zero. */
const LIGHT_RADIUS_BOX = '.5rem';

const G10_IDENTITIES: { id: string; cookie: 'label' | 'label-night' }[] = [
  { id: 'dubplate-label', cookie: 'label' },
  { id: 'dubplate-label-night', cookie: 'label-night' }
];

/**
 * G10 (R3.4, R-5ciii8): every structural name's computed value under
 * each composed identity, against the `dubplate` match block's own
 * declared value, plus the built-in `light` control that proves a leak
 * through the bare `:root` fallback would be caught.
 */
async function g10Rows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const identity of G10_IDENTITIES) {
    const id = `g10/structural/${identity.id}`;
    const host = new URL(endpoints.base).hostname;
    const context = await browser.newContext();
    try {
      await context.addCookies([{ name: THEME_COOKIE, value: identity.cookie, domain: host, path: '/', secure: true }]);
      const { visit, body } = await open(context, new URL('login', endpoints.base).href, 'load');
      const stamp = stampOf(body);
      const fixture = stamp === `dubplate-${identity.cookie}`;
      if (!fixture) {
        await report({ id, fixture, property: false, detail: `served stamp ${stamp}` });
        continue;
      }
      const got = await visit.page.evaluate((names: string[]) => {
        const cs = getComputedStyle(document.documentElement);
        return Object.fromEntries(names.map((n) => [n, cs.getPropertyValue(n).trim()]));
      }, Object.keys(STRUCTURAL));
      const mismatches = Object.entries(STRUCTURAL).filter(([name, want]) => got[name] !== want);
      await report({
        id,
        fixture,
        property: mismatches.length === 0,
        detail:
          mismatches.length === 0
            ? `all eleven structural names matched ${JSON.stringify(STRUCTURAL)}`
            : `mismatched: ${JSON.stringify(mismatches)}; got ${JSON.stringify(got)}`
      });
    } finally {
      await context.close();
    }
  }

  const id = 'g10/light-control';
  const context = await browser.newContext();
  try {
    const { visit, body } = await open(context, new URL('login', endpoints.base).href, 'load');
    const stamp = stampOf(body);
    const fixture = stamp.startsWith('dubplate-');
    if (!fixture) {
      await report({ id, fixture, property: false, detail: `served stamp ${stamp}, want a dubplate-* stamp` });
    } else {
      const before = (
        await visit.page.evaluate(() => getComputedStyle(document.documentElement).getPropertyValue('--radius-box'))
      ).trim();
      const after = (
        await visit.page.evaluate(() => {
          document.documentElement.setAttribute('data-theme', 'light');
          return getComputedStyle(document.documentElement).getPropertyValue('--radius-box');
        })
      ).trim();
      await report({
        id,
        fixture,
        property: before === STRUCTURAL['--radius-box'] && after === LIGHT_RADIUS_BOX,
        detail: `before ${before} (want ${STRUCTURAL['--radius-box']}); after stamping light, ${after} (want ${LIGHT_RADIUS_BOX})`
      });
    }
  } finally {
    await context.close();
  }
}

/**
 * First paint, per theme cookie and emulated preference. JavaScript is
 * off, so nothing but the served stamp and the stylesheet can choose
 * the palette. Chromium paints no frame before render-blocking CSS, so
 * a screenshot at `domcontentloaded` already shows the post-CSS pixel.
 * The cookie is seeded with `addCookies`, since a JavaScript-disabled
 * page cannot write it; the `url` form is refused for a `__Host-` name,
 * so the domain and path are given instead.
 */
async function firstPaintRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const host = new URL(endpoints.base).hostname;
  for (const state of FIRST_PAINT) {
    const id = `first-paint/cookie-${state.cookie ?? 'none'}/${state.scheme}`;
    const context = await browser.newContext({
      javaScriptEnabled: false,
      colorScheme: state.scheme,
      viewport: { width: 640, height: 480 }
    });
    try {
      if (state.cookie !== null) {
        await context.addCookies([
          { name: THEME_COOKIE, value: state.cookie, domain: host, path: '/', secure: true }
        ]);
      }
      const marker = `rendered-row=${encodeURIComponent(id)}`;
      const url = new URL(`login?${marker}`, endpoints.base).href;
      const { visit, body } = await open(context, url, 'domcontentloaded');

      const stamp = /<html\b[^>]*\sdata-theme="([^"]*)"/.exec(body)?.[1] ?? '(none)';
      const logged = (await requestLog(endpoints.requestLog)).find((r) => r.query === marker);
      const sentCookie =
        (logged?.headers.Cookie ?? [])
          .flatMap((h) => h.split(';'))
          .map((c) => c.trim())
          .find((c) => c.startsWith(`${THEME_COOKIE}=`)) ?? null;
      const wantCookie = state.cookie === null ? null : `${THEME_COOKIE}=${state.cookie}`;
      const fixture = stamp === state.stamp && logged !== undefined && sentCookie === wantCookie;
      if (!fixture) {
        await report({
          id,
          fixture,
          property: false,
          detail: `served stamp ${stamp} (want ${state.stamp}); request logged ${logged !== undefined}; theme cookie sent ${sentCookie} (want ${wantCookie})`
        });
        continue;
      }

      const image = decodePNG(await visit.page.screenshot({ type: 'png' }));
      const samples = [
        pixelHex(image, image.width >> 1, image.height >> 1),
        pixelHex(image, image.width - 1, image.height - 1)
      ];
      await report({
        id,
        fixture,
        property: samples.every((s) => s === state.paint),
        detail: `stamp ${stamp}; painted ${samples.join(', ')} (want ${state.paint})`
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * No served page raises a CSP violation. The fixture state puts a
 * `data:` image on the page before judging: DaisyUI's checkbox paints
 * its `--fx-noise` texture as a `data:` SVG, which `img-src` must
 * admit, so the row cannot pass on a page that loads no such image.
 */
async function cspCleanRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const context = await browser.newContext();
  await context.addInitScript({ content: VIOLATION_RECORDER });
  try {
    for (const served of PAGES) {
      const id = `csp-clean/${served.name}`;
      const visit = await openBooted(context, endpoints.base, served);
      const background = await visit.page.evaluate(() => {
        const box = document.createElement('input');
        box.type = 'checkbox';
        box.className = 'checkbox';
        document.querySelector('main')?.append(box);
        return getComputedStyle(box).backgroundImage;
      });
      const fixture = background.includes('url("data:image/svg+xml');
      if (fixture) await visit.page.locator('input.checkbox').screenshot();
      visit.violations = await violations(visit.page);
      await report({
        id,
        fixture,
        property: fixture && visit.violations.length === 0 && visit.consoleErrors.length === 0,
        detail: `checkbox background-image data: ${fixture}; violations ${JSON.stringify(visit.violations)}; console errors ${JSON.stringify(visit.consoleErrors)}`
      });
    }
  } finally {
    await context.close();
  }
}

/**
 * The violation listener is live: an off-origin image injected into a
 * hydrated page is refused by `img-src` and recorded. Without this row
 * a listener that never fires would make every clean row pass.
 */
async function violationListenerRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'csp-violation-listener/img-src';
  const context = await browser.newContext();
  await context.addInitScript({ content: VIOLATION_RECORDER });
  try {
    const visit = await openBooted(context, endpoints.base, PAGES[0]);
    visit.expectedBlock = LISTENER_PROBE;
    const before = await violations(visit.page);
    const fixture = before.length === 0;
    let seen: Violation[] = [];
    if (fixture) {
      await visit.page.evaluate((probe: string) => {
        const img = document.createElement('img');
        img.src = probe;
        document.body.append(img);
      }, LISTENER_PROBE);
      await visit.page
        .waitForFunction(
          () =>
            ((window as unknown as { __renderedViolations?: Violation[] }).__renderedViolations ?? [])
              .length > 0,
          undefined,
          { timeout: 5_000 }
        )
        .catch(() => undefined);
      seen = await violations(visit.page);
      visit.violations = seen;
    }
    await report({
      id,
      fixture,
      property: seen.some((v) => v.directive === 'img-src' && v.blocked === LISTENER_PROBE),
      detail: `violations before injection ${JSON.stringify(before)}; after ${JSON.stringify(seen)}`
    });
  } finally {
    await context.close();
  }
}

/**
 * Whether a problem the browser reported is the one a page raised on
 * purpose: `img-src` blocking its expected URL, as the violation event
 * and as Chromium's console report of it.
 */
function isExpectedViolation(visit: Visit, v: Violation): boolean {
  return visit.expectedBlock !== null && v.directive === 'img-src' && v.blocked === visit.expectedBlock;
}

const AUTH_ERROR_STATUS_RE =
  /^Failed to load resource: (the server responded with a status of \d+|net::ERR_FAILED) \(.*\/(request-link|consume-link)\)$/;

function isExpectedConsoleError(visit: Visit, text: string): boolean {
  if (
    visit.expectedBlock !== null &&
    text.startsWith(`Loading the image '${visit.expectedBlock}' violates the following Content Security Policy directive: "img-src `)
  ) {
    return true;
  }
  return visit.expectsAuthErrorStatus && AUTH_ERROR_STATUS_RE.test(text);
}

/** Refuses the run when any page reported a problem it did not raise on purpose. */
function refuseUnclean(): void {
  for (const visit of visits) {
    const violation = visit.violations.find((v) => !isExpectedViolation(visit, v));
    if (violation !== undefined) {
      throw new Refusal(
        `${visit.url}: securitypolicyviolation: ${violation.directive} blocked ${violation.blocked}`
      );
    }
    const text = visit.consoleErrors.find((t) => !isExpectedConsoleError(visit, t));
    if (text !== undefined) {
      throw new Refusal(`${visit.url}: console error: ${text}`);
    }
  }
}

/** Refuses the run when a request left the owned prefix after the last row reported. */
async function refuseLateLeaks(endpoints: Endpoints): Promise<void> {
  if (prefixWatch === null) return;
  const watch = prefixWatch;
  const outside = (await requestLog(endpoints.requestLog)).slice(watch.seen).filter((r) => !r.path.startsWith(watch.prefix));
  if (outside.length > 0) throw new Refusal(`requests outside ${watch.prefix} after the last row: ${describe(outside)}`);
}

/** How long a row waits for the browser's own icon fetch. */
const ICON_TIMEOUT_MS = 5_000;

/**
 * The owned prefix's paths, from the base URL the test hands in: every
 * request must stay under `prefix`, and the icon links the server
 * inserts name `icons`.
 */
function prefixPaths(base: string): { prefix: string; icons: string[] } {
  return {
    prefix: new URL(base).pathname,
    icons: ['favicon.svg', 'favicon.png'].map((name) => new URL(name, base).pathname)
  };
}

function describe(records: LoggedRequest[]): string {
  return JSON.stringify(records.map((r) => `${r.method} ${r.path}`));
}

interface ThemeColorMeta {
  media: string | null;
  content: string;
}

async function themeColorMetas(page: Page): Promise<ThemeColorMeta[]> {
  return page.evaluate(() =>
    [...document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')].map((m) => ({
      media: m.getAttribute('media'),
      content: m.content
    }))
  );
}

/**
 * The theme-color metas. The expected colors are the visual
 * identity spec's neutrals, quoted from
 * docs/superpowers/specs/2026-09-24-visual-identity-design.md ("Themes"):
 * "`#1f1b16` for light and `#110f0d` for dark, as a `media` pair when
 * following the system, and as one value matching the stamped palette
 * when the cookie overrides it."
 */
const THEME_COLOR = { light: '#1f1b16', dark: '#110f0d' } as const;

/**
 * The `system` stamp's pair switches with the emulated preference: one
 * page, read under light and then dark, the active meta being the one
 * whose media query matches. A pinned palette carries one meta with no
 * media query.
 */
async function themeColorRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const context = await browser.newContext({ colorScheme: 'light' });
  try {
    const { visit, body } = await open(context, new URL('login', endpoints.base).href, 'load');
    const stamp = stampOf(body);
    const metas = await themeColorMetas(visit.page);
    const fixture = stamp === 'dubplate-system' && metas.length === 2 && metas.every((m) => m.media !== null);
    for (const scheme of ['light', 'dark'] as const) {
      const id = `head/theme-color/system/${scheme}`;
      if (!fixture) {
        await report({ id, fixture, property: false, detail: `served stamp ${stamp}; metas ${JSON.stringify(metas)}` });
        continue;
      }
      await visit.page.emulateMedia({ colorScheme: scheme });
      const active = await visit.page.evaluate(() =>
        [...document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')]
          .filter((m) => !m.media || matchMedia(m.media).matches)
          .map((m) => m.content)
      );
      await report({
        id,
        fixture,
        property: active.length === 1 && active[0] === THEME_COLOR[scheme],
        detail: `active theme-color ${JSON.stringify(active)} (want ["${THEME_COLOR[scheme]}"]); metas ${JSON.stringify(metas)}`
      });
    }
  } finally {
    await context.close();
  }

  const host = new URL(endpoints.base).hostname;
  for (const cookie of ['label', 'label-night'] as const) {
    const id = `head/theme-color/cookie-${cookie}`;
    const want = cookie === 'label' ? THEME_COLOR.light : THEME_COLOR.dark;
    const pinned = await browser.newContext();
    try {
      await pinned.addCookies([{ name: THEME_COOKIE, value: cookie, domain: host, path: '/', secure: true }]);
      const { visit, body } = await open(pinned, new URL('login', endpoints.base).href, 'load');
      const stamp = stampOf(body);
      const fixture = stamp === `dubplate-${cookie}`;
      const metas = await themeColorMetas(visit.page);
      await report({
        id,
        fixture,
        property: fixture && metas.length === 1 && metas[0].media === null && metas[0].content === want,
        detail: `served stamp ${stamp}; metas ${JSON.stringify(metas)} (want one, no media, ${want})`
      });
    } finally {
      await pinned.close();
    }
  }
}

/**
 * Every served page makes the browser fetch its icon from inside the
 * owned prefix, and nothing it loads leaves the prefix. The fixture
 * state is the page's own document request in the same stretch of the
 * log, so the row judges the requests that page caused.
 */
async function iconRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const paths = prefixPaths(endpoints.base);
  for (const served of PAGES) {
    const id = `head/icon/${served.name}`;
    const from = (await requestLog(endpoints.requestLog)).length;
    const context = await browser.newContext();
    try {
      await openBooted(context, endpoints.base, served);
      const records = await logSince(endpoints.requestLog, from, (r) => paths.icons.includes(r.path), ICON_TIMEOUT_MS);
      const documentPath = new URL(served.path, endpoints.base).pathname;
      const fixture = records.some((r) => r.path === documentPath);
      const icons = records.filter((r) => paths.icons.includes(r.path));
      const outside = records.filter((r) => !r.path.startsWith(paths.prefix));
      await report({
        id,
        fixture,
        property: icons.length > 0 && outside.length === 0,
        detail: `icon requests ${describe(icons)}; outside ${paths.prefix}: ${describe(outside)}; all ${describe(records)}`
      });
    } finally {
      await context.close();
    }
  }
}

/** The members stub's declared identity, the client navigation's target. */
const MEMBERS_HEADING = 'Members';

/** How long the navigation row watches for a request outside the prefix. */
const NAVIGATION_WATCH_MS = 3_000;

/** An icon link as the server writes it into a served document. */
const ICON_LINK_RE = /<link rel="icon" href="([^"]*)"/g;

/**
 * After a same-document navigation Chromium re-resolves a depth-relative
 * icon link against the new URL and fetches it from there, outside the
 * prefix (probed 2026-09-25 with a bare `history.pushState`). SvelteKit's
 * router rewrites every icon link to its absolute URL when it starts
 * (`@sveltejs/kit` 2.70.3, `src/runtime/client/client.js`, "fix
 * link[rel=icon]"), so in a hydrated page the fetch after a navigation
 * cannot show the defect; the served bytes are read as well, and every
 * icon href there must already start with the prefix. The row loads the
 * `login/link` stub, injects a link to `../members`, and clicks it. The
 * fixture state proves the navigation stayed client-side: no document
 * request for `members` reached the server, and the address ends in
 * `/members`.
 */
async function clientNavigationRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'head/icon/client-navigation';
  const paths = prefixPaths(endpoints.base);
  const context = await browser.newContext();
  try {
    const servedHrefs = [
      ...(await (await context.request.get(new URL(PAGES[1].path, endpoints.base).href)).text()).matchAll(
        ICON_LINK_RE
      )
    ].map((m) => m[1]);
    const before = (await requestLog(endpoints.requestLog)).length;
    const visit = await openBooted(context, endpoints.base, PAGES[1]);
    await logSince(endpoints.requestLog, before, (r) => paths.icons.includes(r.path), ICON_TIMEOUT_MS);

    const from = (await requestLog(endpoints.requestLog)).length;
    await visit.page.evaluate(() => {
      const a = document.createElement('a');
      a.href = '../members';
      a.id = 'rendered-client-navigation';
      a.textContent = 'members';
      document.querySelector('main')?.append(a);
    });
    await visit.page.click('#rendered-client-navigation');
    const heading = visit.page.getByRole('heading', { level: 1, name: MEMBERS_HEADING, exact: true });
    const arrived = await heading
      .waitFor({ timeout: BOOT_TIMEOUT_MS })
      .then(() => true)
      .catch(() => false);
    const records = await logSince(
      endpoints.requestLog,
      from,
      (r) => !r.path.startsWith(paths.prefix),
      NAVIGATION_WATCH_MS
    );
    const pathname = await visit.page.evaluate(() => location.pathname);
    const membersPath = new URL('members', endpoints.base).pathname;
    const documentRequested = records.some((r) => r.path === membersPath);
    const fixture = arrived && !documentRequested && pathname.endsWith('/members');
    const outside = records.filter((r) => !r.path.startsWith(paths.prefix));
    const hrefsInPrefix = servedHrefs.length > 0 && servedHrefs.every((h) => h.startsWith(paths.prefix));
    await report({
      id,
      fixture,
      property: fixture && hrefsInPrefix && outside.length === 0,
      detail: `served icon hrefs ${JSON.stringify(servedHrefs)}; members heading shown ${arrived}; location.pathname ${pathname}; document request for ${membersPath} ${documentRequested}; outside ${paths.prefix} after the click: ${describe(outside)}; all ${describe(records)}`
    });
  } finally {
    await context.close();
  }
}

/**
 * Chromium probes the origin's `/favicon.ico` only when a document
 * declares no icon, so a probe anywhere in the run means some page lost
 * its icon links. The fixture state is at least one icon fetch in the
 * log, which proves this browser fetches icons at all.
 */
async function noFaviconProbeRow(endpoints: Endpoints): Promise<void> {
  const id = 'head/no-favicon-ico';
  const paths = prefixPaths(endpoints.base);
  const records = await requestLog(endpoints.requestLog);
  const icons = records.filter((r) => paths.icons.includes(r.path));
  const probes = records.filter((r) => r.path === '/favicon.ico');
  await report({
    id,
    fixture: icons.length > 0,
    property: probes.length === 0,
    detail: `${icons.length} icon request(s) in the run; /favicon.ico requests ${describe(probes)}`
  });
}

/**
 * The closed fixture control set (VI's component layer and printed
 * surface), each element carrying the bundle's own DaisyUI classes.
 * The class strings are literals in this file so Tailwind's content
 * scan of it emits their CSS into the built bundle, the same way
 * `cspCleanRows`' injected checkbox already does.
 */
const FIXTURE = {
  btnDefault: 'rendered-fx-btn-default',
  btnPrimary: 'rendered-fx-btn-primary',
  btnSecondary: 'rendered-fx-btn-secondary',
  btnWarning: 'rendered-fx-btn-warning',
  btnNeutral: 'rendered-fx-btn-neutral',
  btnGhost: 'rendered-fx-btn-ghost',
  btnDisabled: 'rendered-fx-btn-disabled',
  btnDisabledClass: 'rendered-fx-btn-disabled-class',
  btnLink: 'rendered-fx-btn-link',
  anchorBtn: 'rendered-fx-anchor-btn',
  input: 'rendered-fx-input',
  wrappedInput: 'rendered-fx-wrapped-input',
  select: 'rendered-fx-select',
  checkbox: 'rendered-fx-checkbox',
  radio: 'rendered-fx-radio',
  toggle: 'rendered-fx-toggle',
  fileInput: 'rendered-fx-file-input',
  inputError: 'rendered-fx-input-error',
  inputSuccess: 'rendered-fx-input-success',
  validator: 'rendered-fx-validator',
  validatorHint: 'rendered-fx-validator-hint',
  card: 'rendered-fx-card',
  alert: 'rendered-fx-alert',
  alertControls: 'rendered-fx-alert-controls',
  alertButton: 'rendered-fx-alert-btn',
  alertInput: 'rendered-fx-alert-input',
  proseLink: 'rendered-fx-prose-link',
  badgeUtility: 'rendered-fx-badge-utility',
  display: 'rendered-fx-display',
  pageTitle: 'rendered-fx-page-title'
} as const;

/** The status suffixes DaisyUI's badge and link modifiers take. */
const STATUSES = ['primary', 'secondary', 'accent', 'neutral', 'info', 'success', 'warning', 'error'] as const;
type Status = (typeof STATUSES)[number];

/** Every badge the badge row measures: the plain badge, then one per status. */
const BADGES: { id: string; className: string; status: Status | null }[] = [
  { id: 'rendered-fx-badge-plain', className: 'badge', status: null },
  { id: 'rendered-fx-badge-primary', className: 'badge badge-primary', status: 'primary' },
  { id: 'rendered-fx-badge-secondary', className: 'badge badge-secondary', status: 'secondary' },
  { id: 'rendered-fx-badge-accent', className: 'badge badge-accent', status: 'accent' },
  { id: 'rendered-fx-badge-neutral', className: 'badge badge-neutral', status: 'neutral' },
  { id: 'rendered-fx-badge-info', className: 'badge badge-info', status: 'info' },
  { id: 'rendered-fx-badge-success', className: 'badge badge-success', status: 'success' },
  { id: 'rendered-fx-badge-warning', className: 'badge badge-warning', status: 'warning' },
  { id: 'rendered-fx-badge-error', className: 'badge badge-error', status: 'error' }
];

/** One link per status modifier, for the link row. */
const STATUS_LINKS: { id: string; className: string; status: Status }[] = [
  { id: 'rendered-fx-link-primary', className: 'link link-primary', status: 'primary' },
  { id: 'rendered-fx-link-secondary', className: 'link link-secondary', status: 'secondary' },
  { id: 'rendered-fx-link-accent', className: 'link link-accent', status: 'accent' },
  { id: 'rendered-fx-link-neutral', className: 'link link-neutral', status: 'neutral' },
  { id: 'rendered-fx-link-info', className: 'link link-info', status: 'info' },
  { id: 'rendered-fx-link-success', className: 'link link-success', status: 'success' },
  { id: 'rendered-fx-link-warning', className: 'link link-warning', status: 'warning' },
  { id: 'rendered-fx-link-error', className: 'link link-error', status: 'error' }
];

/**
 * The plain outline button and one per status modifier, for the
 * outline keyline row: an outline button has no fill, so its border is
 * its only edge.
 */
const OUTLINE_BUTTONS: { id: string; className: string; status: Status | null }[] = [
  { id: 'rendered-fx-btn-outline', className: 'btn btn-outline', status: null },
  ...STATUSES.map((status) => ({ id: `rendered-fx-btn-outline-${status}`, className: `btn btn-outline btn-${status}`, status }))
];

/**
 * One alert per fill, each holding a focusable link, so the focus row
 * measures a ring drawn over a colored fill (the plain alert's fill is
 * base-200). The fixture adds one more info alert holding a button and
 * an input, controls that carry box-shadows of their own, so the row
 * also measures a ring on a fill where the halo must outrank them.
 */
const ALERTS: { id: string; className: string }[] = [
  { id: 'rendered-fx-alert', className: 'alert' },
  { id: 'rendered-fx-alert-info', className: 'alert alert-info' },
  { id: 'rendered-fx-alert-success', className: 'alert alert-success' },
  { id: 'rendered-fx-alert-warning', className: 'alert alert-warning' },
  { id: 'rendered-fx-alert-error', className: 'alert alert-error' }
];

/** The fixture's markup: built from literals and this file's own IDs only, so no page byte reaches it. */
function fixtureMarkup(): string {
  const f = FIXTURE;
  const button = (id: string, className: string, extra = '') =>
    `<button id="${id}" class="${className}"${extra}>${id}</button>`;
  return [
    button(f.btnDefault, 'btn'),
    button(f.btnPrimary, 'btn btn-primary'),
    button(f.btnSecondary, 'btn btn-secondary'),
    button(f.btnWarning, 'btn btn-warning'),
    button(f.btnNeutral, 'btn btn-neutral'),
    button(f.btnGhost, 'btn btn-ghost'),
    button(f.btnDisabled, 'btn btn-primary', ' disabled'),
    button(f.btnDisabledClass, 'btn btn-disabled'),
    button(f.btnLink, 'btn btn-link'),
    ...OUTLINE_BUTTONS.map((b) => button(b.id, b.className)),
    `<a id="${f.anchorBtn}" class="btn" href="#rendered-fx">${f.anchorBtn}</a>`,
    `<input id="${f.input}" type="text" class="input" aria-label="${f.input}">`,
    `<label id="${f.wrappedInput}" class="input"><span>wrapped</span><input type="text"></label>`,
    `<select id="${f.select}" class="select" aria-label="${f.select}"><option>fixture</option></select>`,
    `<input id="${f.checkbox}" type="checkbox" class="checkbox" aria-label="${f.checkbox}">`,
    `<input id="${f.radio}" type="radio" class="radio" name="rendered-fx-radio" aria-label="${f.radio}">`,
    `<input id="${f.toggle}" type="checkbox" class="toggle" aria-label="${f.toggle}">`,
    `<input id="${f.fileInput}" type="file" class="file-input" aria-label="${f.fileInput}">`,
    `<input id="${f.inputError}" type="text" class="input input-error" aria-label="${f.inputError}">`,
    `<input id="${f.inputSuccess}" type="text" class="input input-success" aria-label="${f.inputSuccess}">`,
    `<div><input id="${f.validator}" type="text" class="input validator" aria-invalid="true" aria-label="${f.validator}"><p id="${f.validatorHint}" class="validator-hint">fixture hint</p></div>`,
    `<div id="${f.card}" class="card"><div class="card-body">fixture card</div></div>`,
    ...ALERTS.map((a) => `<div id="${a.id}" class="${a.className}">fixture alert <a class="link" href="#rendered-fx">${a.id}</a></div>`),
    `<div id="${f.alertControls}" class="alert alert-info">${button(f.alertButton, 'btn')}<input id="${f.alertInput}" type="text" class="input" aria-label="${f.alertInput}"></div>`,
    `<p>Body text with <a id="${f.proseLink}" class="link text-ink-primary" href="#rendered-fx">a link</a> in it.</p>`,
    ...STATUS_LINKS.map((l) => `<a id="${l.id}" class="${l.className}" href="#rendered-fx">${l.id}</a>`),
    ...BADGES.map((b) => `<span id="${b.id}" class="${b.className}">${b.status ?? 'plain'}</span>`),
    `<span id="${f.badgeUtility}" class="badge rounded-full">utility</span>`,
    `<p id="${f.display}" class="display">Fixture display</p>`,
    `<p id="${f.pageTitle}" class="page-title">Fixture title</p>`
  ].join('');
}

/**
 * Appends the closed set to the booted page's `main`, under a wrapper
 * with room above it so a focus ring's outer edge never samples
 * outside the viewport. Nothing here ships; the elements exist only
 * for this run.
 */
async function injectFixtureControls(page: Page): Promise<void> {
  await page.evaluate((markup: string) => {
    const main = document.querySelector('main');
    if (main === null) throw new Error('no <main> to inject the fixture set into');
    const wrap = document.createElement('div');
    wrap.style.marginTop = '40px';
    wrap.style.display = 'flex';
    wrap.style.flexDirection = 'column';
    wrap.style.gap = '24px';
    // `flex-start` keeps every fixture element at its own content
    // width, so the shadow rows below always find plain background
    // to sample just past an element's right edge.
    wrap.style.alignItems = 'flex-start';
    wrap.innerHTML = markup;
    main.append(wrap);
  }, fixtureMarkup());
}

/**
 * A stamp and an emulated preference the component rows run under,
 * and the palette whose values the page must then show: a cookie pins
 * its palette, and the `system` stamp follows the preference.
 */
interface ThemeState {
  id: string;
  cookie: 'label' | 'label-night' | null;
  scheme: 'light' | 'dark';
  palette: 'label' | 'label-night';
  stamp: string;
}

const COOKIE_STATES: ThemeState[] = [
  { id: 'label', cookie: 'label', scheme: 'light', palette: 'label', stamp: 'dubplate-label' },
  { id: 'label-night', cookie: 'label-night', scheme: 'dark', palette: 'label-night', stamp: 'dubplate-label-night' }
];

const ALL_STATES: ThemeState[] = [
  ...COOKIE_STATES,
  { id: 'system-light', cookie: null, scheme: 'light', palette: 'label', stamp: 'dubplate-system' },
  { id: 'system-dark', cookie: null, scheme: 'dark', palette: 'label-night', stamp: 'dubplate-system' }
];

/**
 * The fixture contexts' viewport: tall enough that the whole closed set
 * paints without scrolling, so a rect read from the page and a
 * screenshot clip share one coordinate space.
 */
const FIXTURE_VIEWPORT = { width: 1280, height: 4000 };

/**
 * Opens `served` under `state` in a new context, injects the closed
 * set, and reports whether the fixture stands: the document carries
 * the expected stamp and the set fits the viewport unscrolled. Every
 * component row checks it before measuring.
 */
async function openThemedFixture(
  browser: Browser,
  endpoints: Endpoints,
  state: ThemeState,
  served: ServedPage = PAGES[0],
  options: { reducedMotion?: 'reduce' } = {}
): Promise<{ context: BrowserContext; visit: Visit; stamped: boolean; problem: string }> {
  const host = new URL(endpoints.base).hostname;
  const context = await browser.newContext({ colorScheme: state.scheme, viewport: FIXTURE_VIEWPORT, ...options });
  if (state.cookie !== null) {
    await context.addCookies([{ name: THEME_COOKIE, value: state.cookie, domain: host, path: '/', secure: true }]);
  }
  const visit = await openBooted(context, endpoints.base, served);
  await injectFixtureControls(visit.page);
  const { stamp, height } = await visit.page.evaluate(() => ({
    stamp: document.documentElement.dataset.theme ?? '(none)',
    height: document.documentElement.scrollHeight
  }));
  const problems: string[] = [];
  if (stamp !== state.stamp) problems.push(`the document is stamped ${stamp}, want ${state.stamp}`);
  if (height > FIXTURE_VIEWPORT.height) problems.push(`the page is ${height}px tall, over the ${FIXTURE_VIEWPORT.height}px viewport`);
  return { context, visit, stamped: problems.length === 0, problem: problems.join('; ') };
}

/** `--ink-primary`'s two palette values (VI's ink table). */
const INK_PRIMARY = { label: '#9d2f26', 'label-night': '#ec7b6a' } as const;

/**
 * The ink each status's text must paint, per palette, quoted from VI's
 * ink table. `null` marks a status with no ink (`label` secondary and
 * `label-night` neutral are fill-only, and neither palette gives
 * secondary or neutral an ink), which falls back to base-content.
 */
const STATUS_INK: Record<'label' | 'label-night', Record<Status, string | null>> = {
  label: {
    primary: '#9d2f26',
    secondary: null,
    accent: '#2d6136',
    neutral: null,
    info: '#2d5a81',
    success: '#2d6136',
    warning: '#754f00',
    error: '#a4261d'
  },
  'label-night': {
    primary: '#ec7b6a',
    secondary: null,
    accent: '#66a46e',
    neutral: null,
    info: '#6d9cc7',
    success: '#66a46e',
    warning: '#e5ad33',
    error: '#e37866'
  }
};

/** `--color-base-content` per palette (VI's palette tables). */
const BASE_CONTENT = { label: '#1f1b16', 'label-night': '#efe4cc' } as const;

/** `--color-base-100`, the plain page background surrounding the fixture set. */
const BASE_100 = { label: '#f4ead3', 'label-night': '#1c1916' } as const;

/** sRGB relative luminance (WCAG 2.2), from 0-255 channel bytes. */
function srgbToLinear(channel: number): number {
  const c = channel / 255;
  return c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
}

function relativeLuminance([r, g, b]: RGBA): number {
  return 0.2126 * srgbToLinear(r) + 0.7152 * srgbToLinear(g) + 0.0722 * srgbToLinear(b);
}

/** WCAG 2.2 contrast ratio between two painted colors. */
function contrastRatio(a: RGBA, b: RGBA): number {
  const [lighter, darker] = [relativeLuminance(a), relativeLuminance(b)].sort((x, y) => y - x);
  return (lighter + 0.05) / (darker + 0.05);
}

/** Paints a resolved CSS color string on a 1px canvas and reads the bytes back (W12.5). */
async function paintedColor(page: Page, cssColor: string): Promise<RGBA> {
  return page.evaluate((color): RGBA => {
    const canvas = document.createElement('canvas');
    canvas.width = 1;
    canvas.height = 1;
    const ctx = canvas.getContext('2d', { colorSpace: 'srgb' });
    if (ctx === null) throw new Error('the canvas has no 2d context');
    ctx.fillStyle = 'transparent';
    ctx.fillStyle = color;
    ctx.fillRect(0, 0, 1, 1);
    const [r, g, b, a] = ctx.getImageData(0, 0, 1, 1, { colorSpace: 'srgb' }).data;
    return [r, g, b, a];
  }, cssColor);
}

/** The painted color `elementId`'s computed `property` resolves to on `page`. */
async function paintedProperty(page: Page, elementId: string, property: string): Promise<RGBA> {
  return paintedColor(page, await computedProperty(page, elementId, property));
}

async function computedProperty(page: Page, elementId: string, property: string): Promise<string> {
  return page.evaluate(
    ({ id, prop }) => {
      const el = document.getElementById(id);
      if (el === null) throw new Error(`no element #${id}`);
      return getComputedStyle(el).getPropertyValue(prop);
    },
    { id: elementId, prop: property }
  );
}

/**
 * Every element a keyboard `Tab` can reach: standard interactive tags
 * plus an explicit `tabindex`, excluding a negative `tabindex` (removed
 * from the tab order on purpose) and anything with no painted box
 * (`display: none`, `visibility: hidden`, or zero size).
 */
const TABBABLE_SELECTOR =
  'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]';

/** The attribute {@link enumerateTabbable} stamps on each queried element, read back by {@link focusedInfo}. */
const TAB_PROBE_ATTR = 'data-rendered-tab-probe';

/** One {@link TABBABLE_SELECTOR} match with a painted box: `id` is its stable identity, `label` its printable name. */
interface TabTarget {
  id: string;
  label: string;
}

/**
 * Every {@link TABBABLE_SELECTOR} match with a painted box, each
 * stamped with {@link TAB_PROBE_ATTR} so the Tab walk can name which
 * queried element it lands on by identity rather than by position in
 * a cycle that may wrap or revisit.
 */
async function enumerateTabbable(page: Page): Promise<TabTarget[]> {
  return page.evaluate(
    ({ selector, attr }: { selector: string; attr: string }) => {
      const targets: TabTarget[] = [];
      for (const el of Array.from(document.querySelectorAll<HTMLElement>(selector))) {
        if (el.tabIndex < 0) continue;
        const style = getComputedStyle(el);
        if (style.display === 'none' || style.visibility === 'hidden') continue;
        const rect = el.getBoundingClientRect();
        if (rect.width <= 0 || rect.height <= 0) continue;
        const id = String(targets.length);
        el.setAttribute(attr, id);
        const label = el.id || `${el.tagName.toLowerCase()}:${(el.textContent ?? '').trim().slice(0, 24)}`;
        targets.push({ id, label });
      }
      return targets;
    },
    { selector: TABBABLE_SELECTOR, attr: TAB_PROBE_ATTR }
  );
}

interface Outline {
  style: string;
  width: string;
  offset: string;
  color: string;
}

interface FocusedInfo {
  label: string;
  probe: string | null;
  /** The focused element's own outline. */
  own: Outline;
  /**
   * The outline of the DaisyUI control wrapping the focused element
   * (`<label class="input">`, `.select`, `.textarea`), or `null` when
   * nothing wraps it: DaisyUI draws a wrapped control's ring on the
   * wrapper, so the wrapper is the element the ring is judged on.
   */
  wrapper: Outline | null;
  /** The ring-bearing element's box: the wrapper's when there is one. */
  rect: { x: number; y: number; width: number };
}

/** `document.activeElement`'s outline and geometry, or `null` once `Tab` leaves the page (a wrapped cycle). */
async function focusedInfo(page: Page): Promise<FocusedInfo | null> {
  return page.evaluate((attr: string) => {
    const el = document.activeElement as HTMLElement | null;
    if (el === null || el === document.body) return null;
    const outline = (target: Element) => {
      const style = getComputedStyle(target);
      return {
        style: style.outlineStyle,
        width: style.outlineWidth,
        offset: style.outlineOffset,
        color: style.outlineColor
      };
    };
    const wrapperEl = el.parentElement?.closest('.input, .select, .textarea') ?? null;
    const rect = (wrapperEl ?? el).getBoundingClientRect();
    const label = el.id || wrapperEl?.id || `${el.tagName.toLowerCase()}:${(el.textContent ?? '').trim().slice(0, 24)}`;
    return {
      label,
      probe: el.getAttribute(attr),
      own: outline(el),
      wrapper: wrapperEl === null ? null : outline(wrapperEl),
      rect: { x: rect.x, y: rect.y, width: rect.width }
    };
  }, TAB_PROBE_ATTR);
}

/**
 * Waits out any transition the newly focused element (or the control
 * wrapping it) runs on focus: DaisyUI's `.btn` transitions
 * `box-shadow`, so a screenshot taken at once catches a halo still
 * growing in. A cancelled transition rejects `finished`, which counts
 * as settled.
 */
async function settleFocusTransitions(page: Page): Promise<void> {
  await page.evaluate(async () => {
    const el = document.activeElement;
    if (el === null) return;
    const wrapper = el.parentElement?.closest('.input, .select, .textarea') ?? null;
    const animations = [el, wrapper].flatMap((target) => target?.getAnimations() ?? []);
    await Promise.all(animations.map((a) => a.finished.catch(() => undefined)));
  });
}

function describeOutline(o: Outline): string {
  return `${o.style} ${o.width} ${o.color} offset ${o.offset}`;
}

/**
 * `component/focus/<page>/<state>`: every tabbable element the page
 * serves, enumerated by DOM query, plus the injected closed set
 * (the icon-only exception does not apply, since the set carries no
 * icon-only control), under both cookie stamps and the `system` stamp
 * in either preference. A page must serve a nonzero tabbable count
 * and every injected control by role before any ring is judged. Each
 * element's ring is read back at `--ink-primary` by tabbing to it; a
 * control DaisyUI wraps shows one ring, on the wrapper, and none on
 * the focused element inside it. The pixel just inside the ring (the
 * offset gap) and the one just outside it each clear 3:1 against the
 * ring, both sampled above the element's top edge, where the printed
 * surface casts no shadow (the shadow's own offset is bottom-right
 * only); for a control inside a filled alert, those two pixels fall on
 * the alert's fill unless a page-colored halo covers them. The walk
 * also proves coverage by identity: every element
 * {@link enumerateTabbable} finds must be focused at least once (a
 * revisit after a wrap counts once), and the row fails, naming
 * whichever queried element the walk never reached, rather than only
 * comparing a visited count against the query's own total.
 */
async function focusRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const state of ALL_STATES) {
    for (const served of PAGES) {
      const id = `component/focus/${served.name}/${state.id}`;
      const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, state, served);
      try {
        if (!stamped) {
          await report({ id, fixture: false, property: false, detail: problem });
          continue;
        }
        const targets = await enumerateTabbable(visit.page);
        const total = targets.length;
        if (total === 0) {
          await report({ id, fixture: false, property: false, detail: 'no tabbable element found on the page' });
          continue;
        }

        const buttonNames = [
          FIXTURE.btnDefault,
          FIXTURE.btnPrimary,
          FIXTURE.btnSecondary,
          FIXTURE.btnWarning,
          FIXTURE.btnNeutral,
          FIXTURE.btnGhost
        ];
        const missingRoles: string[] = [];
        for (const name of buttonNames) {
          if ((await visit.page.getByRole('button', { name, exact: true }).count()) === 0) missingRoles.push(name);
        }
        if ((await visit.page.getByRole('textbox', { name: 'wrapped', exact: true }).count()) === 0) {
          missingRoles.push(FIXTURE.wrappedInput);
        }
        if ((await visit.page.getByRole('textbox', { name: FIXTURE.input, exact: true }).count()) === 0) {
          missingRoles.push(FIXTURE.input);
        }
        if ((await visit.page.getByRole('link', { name: ALERTS[1].id, exact: true }).count()) === 0) {
          missingRoles.push(`${ALERTS[1].id}'s link`);
        }
        if ((await visit.page.getByRole('button', { name: FIXTURE.alertButton, exact: true }).count()) === 0) {
          missingRoles.push(FIXTURE.alertButton);
        }
        if ((await visit.page.getByRole('textbox', { name: FIXTURE.alertInput, exact: true }).count()) === 0) {
          missingRoles.push(FIXTURE.alertInput);
        }
        if (missingRoles.length > 0) {
          await report({ id, fixture: false, property: false, detail: `missing by role: ${missingRoles.join(', ')}` });
          continue;
        }

        await visit.page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
        const want = hexRGBA(INK_PRIMARY[state.palette]);
        const wantText = `solid 2px ${INK_PRIMARY[state.palette]} offset 2px`;
        const problems: string[] = [];
        const visitedProbes = new Set<string>();
        let visited = 0;
        // A margin of 2 over `total` absorbs a cycle that wraps back
        // to the first element before this bound stops it; a landing
        // back on `document.body` (`focusedInfo` returning `null`)
        // ends the loop early regardless.
        for (let i = 0; i < total + 2; i++) {
          await visit.page.keyboard.press('Tab');
          await settleFocusTransitions(visit.page);
          const focused = await focusedInfo(visit.page);
          if (focused === null) break;
          visited++;
          if (focused.probe !== null) visitedProbes.add(focused.probe);
          if (focused.wrapper !== null && focused.own.style !== 'none') {
            problems.push(`${focused.label}: a second ring on the wrapped control itself, ${describeOutline(focused.own)} (want none)`);
            continue;
          }
          const ringOutline = focused.wrapper ?? focused.own;
          const ring = await paintedColor(visit.page, ringOutline.color);
          if (
            ringOutline.style !== 'solid' ||
            ringOutline.width !== '2px' ||
            ringOutline.offset !== '2px' ||
            !sameRGBA(ring, want)
          ) {
            problems.push(`${focused.label}: outline ${describeOutline(ringOutline)} (want ${wantText})`);
            continue;
          }
          const x = Math.round(focused.rect.x + focused.rect.width / 2);
          const clipY = Math.round(focused.rect.y - 8);
          const shot = await visit.page.screenshot({ clip: { x: x - 2, y: clipY, width: 4, height: 8 } });
          const image = decodePNG(shot);
          // Relative to clipY: the outline band spans offset 2px to
          // offset+width 4px above the element's top edge, so row 7
          // (one pixel below the band) sits inside the offset gap and
          // row 3 (one pixel above the band) sits outside the ring.
          const inside = pixelRGBA(image, 2, 7);
          const outside = pixelRGBA(image, 2, 3);
          const insideC = contrastRatio(inside, ring);
          const outsideC = contrastRatio(outside, ring);
          if (insideC < 3 || outsideC < 3) {
            problems.push(
              `${focused.label}: ring contrast inside ${insideC.toFixed(2)}:1, outside ${outsideC.toFixed(2)}:1 (want 3:1 or better)`
            );
          }
        }
        if (visited === 0) {
          await report({ id, fixture: false, property: false, detail: `tabbing found no focusable element (of ${total} tabbable by query)` });
          continue;
        }
        const missing = targets.filter((t) => !visitedProbes.has(t.id));
        if (missing.length > 0) {
          problems.push(`never focused: ${missing.map((t) => t.label).join(', ')} (of ${total} queried)`);
        }
        await report({
          id,
          fixture: true,
          property: problems.length === 0,
          detail:
            problems.length === 0
              ? `${visitedProbes.size} of ${total} tabbable elements visited by identity, every ring clean`
              : problems.join('; ')
        });
      } finally {
        await context.close();
      }
    }
  }
}

/** Whether `a` and `b` agree within a tenth of a pixel, absorbing subpixel rounding. */
function closeTo(a: number, b: number, tolerance = 0.1): boolean {
  return Math.abs(a - b) <= tolerance;
}

/**
 * `component/pressed` and `component/reduced-motion`: a pressed
 * button's box moves 2px down and right into its shadow (no motion
 * preference), and does not move under `prefers-reduced-motion:
 * reduce`. `transform` alone would miss DaisyUI's own independent
 * `translate` declaration, which composes with it, so this measures
 * the painted box's own position instead of either CSS property. The
 * fixture state itself is `.btn:active`'s `box-shadow: none`
 * (`lib/styles/components.css`), read back alongside the settled
 * position, so a mouse-down that never lands on the button (and so
 * never reaches `:active`) fails the row instead of silently reporting
 * an unmoved box as `fixture: true`.
 */
async function pressedRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const reduce of [false, true]) {
    const id = reduce ? 'component/reduced-motion' : 'component/pressed';
    const context = await browser.newContext(reduce ? { reducedMotion: 'reduce' } : {});
    try {
      const visit = await openBooted(context, endpoints.base, PAGES[0]);
      await injectFixtureControls(visit.page);
      const locator = visit.page.locator(`#${FIXTURE.btnDefault}`);
      const box = await locator.boundingBox();
      if (box === null) {
        await report({ id, fixture: false, property: false, detail: 'the default button has no box' });
        continue;
      }
      await visit.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await visit.page.mouse.down();
      // DaisyUI's `.btn` transitions `box-shadow` (and `transform`)
      // over 0.2s but not `translate`, so the box lands at once while
      // the shadow is still fading; a read immediately after the press
      // would catch the shadow mid-transition. This waits the
      // transition out, then reads the shadow and the box as one
      // settled snapshot of the pressed state.
      await visit.page.waitForTimeout(300);
      const pressedShadow = await computedProperty(visit.page, FIXTURE.btnDefault, 'box-shadow');
      const pressedBox = await locator.boundingBox();
      await visit.page.mouse.up();
      if (pressedShadow !== 'none') {
        await report({
          id,
          fixture: false,
          property: false,
          detail: `the button's box-shadow read ${pressedShadow} while the mouse was held down, so its pressed state was never reached (want none)`
        });
        continue;
      }
      if (pressedBox === null) {
        await report({ id, fixture: false, property: false, detail: 'the default button lost its box while pressed' });
        continue;
      }
      const delta = { x: pressedBox.x - box.x, y: pressedBox.y - box.y };
      const want = reduce ? { x: 0, y: 0 } : { x: 2, y: 2 };
      const moved = closeTo(delta.x, want.x) && closeTo(delta.y, want.y);
      await report({
        id,
        fixture: true,
        property: moved,
        detail: `box moved (${delta.x.toFixed(2)}, ${delta.y.toFixed(2)}) while pressed (want (${want.x}, ${want.y}))`
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * `component/pressed/flat`: a ghost button and a link button carry no
 * shadow to press into, so a press moves neither (VI: "Ghost buttons
 * carry no border or shadow"). The fixture state is each button
 * matching `:active` while the mouse is held on it.
 */
async function pressedFlatRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'component/pressed/flat';
  const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, COOKIE_STATES[0]);
  try {
    if (!stamped) {
      await report({ id, fixture: false, property: false, detail: problem });
      return;
    }
    const results: string[] = [];
    let reached = true;
    let still = true;
    for (const name of [FIXTURE.btnGhost, FIXTURE.btnLink]) {
      const locator = visit.page.locator(`#${name}`);
      const box = await locator.boundingBox();
      if (box === null) {
        reached = false;
        results.push(`${name} has no box`);
        continue;
      }
      await visit.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await visit.page.mouse.down();
      await visit.page.waitForTimeout(300);
      const active = await locator.evaluate((el) => el.matches(':active'));
      const pressedBox = await locator.boundingBox();
      await visit.page.mouse.up();
      if (!active || pressedBox === null) {
        reached = false;
        results.push(`${name} never reached :active`);
        continue;
      }
      const delta = { x: pressedBox.x - box.x, y: pressedBox.y - box.y };
      if (!closeTo(delta.x, 0) || !closeTo(delta.y, 0)) still = false;
      results.push(`${name} moved (${delta.x.toFixed(2)}, ${delta.y.toFixed(2)})`);
    }
    await report({ id, fixture: reached, property: reached && still, detail: `${results.join('; ')} (want (0, 0) each)` });
  } finally {
    await context.close();
  }
}

/**
 * `component/surface/*` (VI's printed surface, `lib/themes/surface.css`):
 * the button and card/alert shadows at their documented strengths, the
 * dark keylines, a control's border, and the ghost button's absent
 * border and shadow. Every expectation is a literal computed from the
 * surface tokens' own committed values by a probe on 2026-09-25
 * (Chromium 153 through playwright@1.63.0), never read from the page.
 */
const SURFACE_EXPECTED = {
  buttonShadow: { label: '#7f796b', 'label-night': '#090807' },
  cardShadow: { label: '#948c7e', 'label-night': '#0b0a08' },
  keylineBase: { label: '#1f1b16', 'label-night': '#706a5f' },
  keylineFill: { label: '#1f1b16', 'label-night': '#000000' }
} as const;

/** An element's own `<top>px <right>px 0px 0px` offset-only box-shadow, with its color token stripped. */
function shadowGeometry(boxShadow: string): string | null {
  const m = /^\S+\([^)]*\)\s*(.+)$/.exec(boxShadow);
  return m === null ? null : m[1].trim();
}

/** The single pixel at document coordinates `(x, y)`, read back from a 1x1 screenshot clip. */
async function paintedPixelAt(page: Page, x: number, y: number): Promise<RGBA> {
  const shot = await page.screenshot({ clip: { x, y, width: 1, height: 1 } });
  return pixelRGBA(decodePNG(shot), 0, 0);
}

/**
 * One `component/surface/shadow/<name>/<theme>` row: parses the
 * element's own computed `box-shadow` geometry against `offset` (the
 * hard, unblurred `<offset>px <offset>px 0px 0px` the printed surface
 * always casts), then samples one pixel 1px right of the element's
 * right edge at mid-height (inside the shadow band, expected
 * `wantShadow`) and one pixel `offset + 1` past the right edge
 * (beyond the band, expected the page's own `base-100`).
 */
async function shadowRow(
  page: Page,
  name: string,
  theme: 'label' | 'label-night',
  elementId: string,
  offset: number,
  wantShadowHex: string
): Promise<void> {
  const id = `component/surface/shadow/${name}/${theme}`;
  const boxShadow = await computedProperty(page, elementId, 'box-shadow');
  const geometry = shadowGeometry(boxShadow);
  const wantGeometry = `${offset}px ${offset}px 0px 0px`;
  if (geometry !== wantGeometry) {
    await report({ id, fixture: geometry !== null, property: false, detail: `box-shadow "${boxShadow}" (want geometry ${wantGeometry})` });
    return;
  }
  const rect = await page.evaluate((elId) => {
    const el = document.getElementById(elId);
    if (el === null) throw new Error(`no element #${elId}`);
    const r = el.getBoundingClientRect();
    return { right: r.right, top: r.top, height: r.height };
  }, elementId);
  const midY = Math.round(rect.top + rect.height / 2);
  const insideX = Math.round(rect.right) + 1;
  const outsideX = Math.round(rect.right) + offset + 1;
  const inside = await paintedPixelAt(page, insideX, midY);
  const outside = await paintedPixelAt(page, outsideX, midY);
  const wantShadow = hexRGBA(wantShadowHex);
  const wantBg = hexRGBA(BASE_100[theme]);
  const shadowOK = sameRGBA(inside, wantShadow);
  const bgOK = sameRGBA(outside, wantBg);
  await report({
    id,
    fixture: true,
    property: shadowOK && bgOK,
    detail: `geometry ${geometry}; (${insideX},${midY})=[${inside.join(',')}] want shadow ${wantShadowHex}; (${outsideX},${midY})=[${outside.join(',')}] want bg ${BASE_100[theme]}`
  });
}

/** A computed color that paints nothing: DaisyUI's `#0000` and CSS `transparent` both resolve to it. */
const TRANSPARENT = 'rgba(0, 0, 0, 0)';

async function surfaceRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const state of COOKIE_STATES) {
    const theme = state.palette;
    const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, state);
    try {
      if (!stamped) {
        for (const row of ['keyline-base', 'keyline-fill', 'keyline-outline', 'keyline-controls', 'border-width', 'ghost', 'flat']) {
          await report({ id: `component/surface/${row}/${theme}`, fixture: false, property: false, detail: problem });
        }
        for (const name of ['button', 'card', 'alert']) {
          await report({ id: `component/surface/shadow/${name}/${theme}`, fixture: false, property: false, detail: problem });
        }
        continue;
      }
      await shadowRow(visit.page, 'button', theme, FIXTURE.btnDefault, 2, SURFACE_EXPECTED.buttonShadow[theme]);
      await shadowRow(visit.page, 'card', theme, FIXTURE.card, 3, SURFACE_EXPECTED.cardShadow[theme]);
      await shadowRow(visit.page, 'alert', theme, FIXTURE.alert, 3, SURFACE_EXPECTED.cardShadow[theme]);

      const border = (elementId: string) => paintedProperty(visit.page, elementId, 'border-top-color');
      const defaultBorder = await border(FIXTURE.btnDefault);
      const cardBorder = await border(FIXTURE.card);
      const primaryBorder = await border(FIXTURE.btnPrimary);
      const alertBorder = await border(FIXTURE.alert);
      const wantBase = hexRGBA(SURFACE_EXPECTED.keylineBase[theme]);
      const wantFill = hexRGBA(SURFACE_EXPECTED.keylineFill[theme]);
      await report({
        id: `component/surface/keyline-base/${theme}`,
        fixture: true,
        property: sameRGBA(defaultBorder, wantBase) && sameRGBA(cardBorder, wantBase),
        detail: `default button border [${defaultBorder.join(',')}], card border [${cardBorder.join(',')}] (want ${SURFACE_EXPECTED.keylineBase[theme]})`
      });
      await report({
        id: `component/surface/keyline-fill/${theme}`,
        fixture: true,
        property: sameRGBA(primaryBorder, wantFill) && sameRGBA(alertBorder, wantFill),
        detail: `primary button border [${primaryBorder.join(',')}], alert border [${alertBorder.join(',')}] (want ${SURFACE_EXPECTED.keylineFill[theme]})`
      });

      // An outline button's border is its only edge, so it paints the
      // status's ink (base-content for a status with none), never the
      // fill keyline a filled button of the same status takes.
      const outlineBorders: string[] = [];
      let outlineOK = true;
      for (const b of OUTLINE_BUTTONS) {
        const got = await border(b.id);
        const wantInk = statusInk(theme, b.status);
        if (!sameRGBA(got, hexRGBA(wantInk))) outlineOK = false;
        outlineBorders.push(`${b.id} [${got.join(',')}] want ${wantInk}`);
      }
      await report({
        id: `component/surface/keyline-outline/${theme}`,
        fixture: true,
        property: outlineOK,
        detail: outlineBorders.join(', ')
      });

      // The selection controls' own edges: DaisyUI draws each at 20% to
      // 50% of base-content, under 3:1 on every base.
      const controls = [FIXTURE.checkbox, FIXTURE.radio, FIXTURE.toggle, FIXTURE.fileInput, FIXTURE.select];
      const controlBorders: string[] = [];
      let controlsOK = true;
      for (const control of controls) {
        const got = await border(control);
        if (!sameRGBA(got, wantBase)) controlsOK = false;
        controlBorders.push(`${control} [${got.join(',')}]`);
      }
      await report({
        id: `component/surface/keyline-controls/${theme}`,
        fixture: true,
        property: controlsOK,
        detail: `${controlBorders.join(', ')} (want ${SURFACE_EXPECTED.keylineBase[theme]} each)`
      });

      const defaultBorderWidth = await computedProperty(visit.page, FIXTURE.btnDefault, 'border-top-width');
      const inputBorderWidth = await computedProperty(visit.page, FIXTURE.input, 'border-top-width');
      await report({
        id: `component/surface/border-width/${theme}`,
        fixture: true,
        property: defaultBorderWidth === '2px' && inputBorderWidth === '2px',
        detail: `default button border-top-width ${defaultBorderWidth}, input border-top-width ${inputBorderWidth} (want 2px, 2px)`
      });

      // A flat button keeps a solid border drawn transparent, never
      // `border-style: none`: forced-colors mode repaints a transparent
      // border in the system color, so the control keeps its edge there.
      const flat = async (elementId: string) => ({
        id: elementId,
        shadow: await computedProperty(visit.page, elementId, 'box-shadow'),
        style: await computedProperty(visit.page, elementId, 'border-top-style'),
        color: await computedProperty(visit.page, elementId, 'border-top-color')
      });
      const isFlat = (f: { shadow: string; style: string; color: string }) =>
        f.shadow === 'none' && f.style === 'solid' && f.color === TRANSPARENT;
      const ghost = await flat(FIXTURE.btnGhost);
      await report({
        id: `component/surface/ghost/${theme}`,
        fixture: true,
        property: isFlat(ghost),
        detail: `box-shadow ${ghost.shadow}, border ${ghost.style} ${ghost.color} (want none, solid ${TRANSPARENT})`
      });
      const flats = [
        await flat(FIXTURE.btnDisabled),
        await flat(FIXTURE.btnDisabledClass),
        await flat(FIXTURE.btnLink)
      ];
      await report({
        id: `component/surface/flat/${theme}`,
        fixture: true,
        property: flats.every(isFlat),
        detail: `${flats.map((f) => `${f.id}: box-shadow ${f.shadow}, border ${f.style} ${f.color}`).join('; ')} (want none, solid ${TRANSPARENT} each)`
      });
    } finally {
      await context.close();
    }
  }
}

/** The WCAG 2.2 text floor (SC 1.4.3) every badge and link color must clear. */
const TEXT_FLOOR = 4.5;

/** The ink a status's text paints under `palette`, or base-content for a status with none. */
function statusInk(palette: 'label' | 'label-night', status: Status | null): string {
  return (status === null ? null : STATUS_INK[palette][status]) ?? BASE_CONTENT[palette];
}

/**
 * `component/badge/<state>`: every badge (the plain badge and one per
 * status) under both cookie stamps and the `system` stamp in either
 * preference. A badge is a squared outline (VI ruling 12), so the
 * pixel painted just inside its left padding must be the page's own
 * base-100, which proves both a transparent fill and no noise layer.
 * Its text and border must paint the status's ink (base-content for
 * a status with none), and the text must clear 4.5:1 against that
 * painted ground.
 */
async function badgeRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const state of ALL_STATES) {
    const id = `component/badge/${state.id}`;
    const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, state);
    try {
      if (!stamped) {
        await report({ id, fixture: false, property: false, detail: problem });
        continue;
      }
      const wantGround = hexRGBA(BASE_100[state.palette]);
      const problems: string[] = [];
      const measured: string[] = [];
      for (const badge of BADGES) {
        const rect = await visit.page.evaluate((elId) => {
          const r = document.getElementById(elId)?.getBoundingClientRect();
          return r === undefined ? null : { left: r.left, top: r.top, height: r.height };
        }, badge.id);
        if (rect === null) {
          problems.push(`${badge.id} is absent`);
          continue;
        }
        const text = await paintedProperty(visit.page, badge.id, 'color');
        const edge = await paintedProperty(visit.page, badge.id, 'border-top-color');
        // 2px of border, then 3px into the 10px padding, clear of every glyph.
        const ground = await paintedPixelAt(visit.page, Math.round(rect.left) + 5, Math.round(rect.top + rect.height / 2));
        const wantHex = statusInk(state.palette, badge.status);
        const want = hexRGBA(wantHex);
        const ratio = contrastRatio(text, ground);
        measured.push(`${badge.status ?? 'plain'} ${ratio.toFixed(2)}:1`);
        if (!sameRGBA(ground, wantGround)) {
          problems.push(`${badge.id}: ground [${ground.join(',')}] (want ${BASE_100[state.palette]}, no fill and no noise)`);
        }
        if (!sameRGBA(text, want) || !sameRGBA(edge, want)) {
          problems.push(`${badge.id}: text [${text.join(',')}], border [${edge.join(',')}] (want ${wantHex} both)`);
        }
        if (ratio < TEXT_FLOOR) {
          problems.push(`${badge.id}: text ${ratio.toFixed(2)}:1 on its ground (want ${TEXT_FLOOR}:1 or better)`);
        }
      }
      await report({
        id,
        fixture: true,
        property: problems.length === 0,
        detail: problems.length === 0 ? `every badge clean: ${measured.join(', ')}` : problems.join('; ')
      });
    } finally {
      await context.close();
    }
  }
}

/** Each palette's `--color-error` and `--color-success` fills and `--ink-error`, quoted from VI's tables. */
const STATE_COLORS = {
  label: { error: '#a8261d', success: '#2f6b3a', inkError: '#a4261d' },
  'label-night': { error: '#e0604f', success: '#5a9e66', inkError: '#e37866' }
} as const;

/**
 * `component/input-states/<palette>`: DaisyUI's error and success
 * input modifiers and its `validator` keep their state borders under
 * the component layer's base keyline, and a shown validator hint is
 * set in the error ink rather than the raw error fill (W2.7). The
 * fixture state is the hint shown: DaisyUI makes it visible only on an
 * invalid validator.
 */
async function inputStateRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const state of COOKIE_STATES) {
    const id = `component/input-states/${state.palette}`;
    const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, state);
    try {
      const hintVisible = (await computedProperty(visit.page, FIXTURE.validatorHint, 'visibility')) === 'visible';
      if (!stamped || !hintVisible) {
        await report({ id, fixture: false, property: false, detail: `${problem}; validator hint visible ${hintVisible}` });
        continue;
      }
      const want = STATE_COLORS[state.palette];
      const border = (elementId: string) => paintedProperty(visit.page, elementId, 'border-top-color');
      const checks = [
        { what: `${FIXTURE.inputError} border`, got: await border(FIXTURE.inputError), want: want.error },
        { what: `${FIXTURE.inputSuccess} border`, got: await border(FIXTURE.inputSuccess), want: want.success },
        { what: `${FIXTURE.validator} border`, got: await border(FIXTURE.validator), want: want.error },
        {
          what: `${FIXTURE.validatorHint} text`,
          got: await paintedProperty(visit.page, FIXTURE.validatorHint, 'color'),
          want: want.inkError
        }
      ];
      await report({
        id,
        fixture: true,
        property: checks.every((c) => sameRGBA(c.got, hexRGBA(c.want))),
        detail: checks.map((c) => `${c.what} [${c.got.join(',')}] (want ${c.want})`).join('; ')
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * `component/link/<palette>`: a prose link is DaisyUI's `.link`,
 * underlined at rest, so it never relies on color alone (SC 1.4.1);
 * every `link-<status>` and `.btn-link` paints an ink (base-content
 * for a status with none) that clears 4.5:1 on base-100; and a
 * hovered `a.btn` stays free of the link underline.
 */
async function linkRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const state of COOKIE_STATES) {
    const id = `component/link/${state.palette}`;
    const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, state);
    try {
      if (!stamped) {
        await report({ id, fixture: false, property: false, detail: problem });
        continue;
      }
      const ground = hexRGBA(BASE_100[state.palette]);
      const problems: string[] = [];
      const color = (elementId: string) => paintedProperty(visit.page, elementId, 'color');

      const proseDecoration = await computedProperty(visit.page, FIXTURE.proseLink, 'text-decoration-line');
      if (proseDecoration !== 'underline') problems.push(`${FIXTURE.proseLink} text-decoration-line ${proseDecoration} at rest (want underline)`);
      const inked = [
        ...STATUS_LINKS.map((l) => ({ id: l.id, want: statusInk(state.palette, l.status) })),
        { id: FIXTURE.proseLink, want: INK_PRIMARY[state.palette] },
        { id: FIXTURE.btnLink, want: INK_PRIMARY[state.palette] }
      ];
      for (const link of inked) {
        const got = await color(link.id);
        const ratio = contrastRatio(got, ground);
        if (!sameRGBA(got, hexRGBA(link.want)) || ratio < TEXT_FLOOR) {
          problems.push(`${link.id}: [${got.join(',')}] at ${ratio.toFixed(2)}:1 (want ${link.want}, ${TEXT_FLOOR}:1 or better)`);
        }
      }

      const box = await visit.page.locator(`#${FIXTURE.anchorBtn}`).boundingBox();
      if (box === null) {
        problems.push(`${FIXTURE.anchorBtn} has no box`);
      } else {
        await visit.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        const hovered = await visit.page.locator(`#${FIXTURE.anchorBtn}`).evaluate((el) => el.matches(':hover'));
        const decoration = await computedProperty(visit.page, FIXTURE.anchorBtn, 'text-decoration-line');
        if (!hovered) problems.push(`${FIXTURE.anchorBtn} never reached :hover`);
        if (decoration !== 'none') problems.push(`${FIXTURE.anchorBtn} hovered text-decoration-line ${decoration} (want none)`);
      }
      await report({
        id,
        fixture: true,
        property: problems.length === 0,
        detail: problems.length === 0 ? `prose link underlined at rest; ${inked.length} inked links clean; a.btn hover plain` : problems.join('; ')
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * `component/utility-wins`: a Tailwind utility in markup beats the
 * component layer, which shares its cascade layer. The fixture state
 * is the plain badge squared by the layer (DaisyUI's own badge is
 * rounded), proving the layer's rule applies at all; the property is
 * `rounded-full` rounding a badge again.
 */
async function utilityWinsRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'component/utility-wins';
  const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, COOKIE_STATES[0]);
  try {
    const plain = await computedProperty(visit.page, BADGES[0].id, 'border-top-left-radius');
    const utility = await computedProperty(visit.page, FIXTURE.badgeUtility, 'border-top-left-radius');
    const fixture = stamped && plain === '0px';
    await report({
      id,
      fixture,
      property: fixture && utility !== '0px',
      detail: `${problem} plain badge radius ${plain} (want 0px); rounded-full badge radius ${utility} (want nonzero)`.trim()
    });
  } finally {
    await context.close();
  }
}

/**
 * The three embedded faces (VI "Type"), their files, and the family
 * name each `@font-face` rule declares, quoted from
 * web/src/lib/themes/fonts.css.
 */
const FACES: { id: string; family: string; file: string }[] = [
  { id: 'alfa-slab-one', family: 'Alfa Slab One', file: 'AlfaSlabOne-Regular.woff2' },
  { id: 'work-sans', family: 'Work Sans', file: 'WorkSans-Variable.woff2' },
  { id: 'ibm-plex-mono', family: 'IBM Plex Mono', file: 'IBMPlexMono-Regular.woff2' }
];

/**
 * `fonts/<face>`: each face loads, and its file is fetched from inside
 * the owned prefix. The fixture state is the face declared in
 * `document.fonts`. Then `.display` sets the slab face, at weight 400,
 * and that face is the one loaded (`component/display-face`).
 */
async function fontRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  const paths = prefixPaths(endpoints.base);
  // A face in use on the page (Work Sans sets the body) is fetched
  // while the page loads, so the log is read from before it opens.
  const from = (await requestLog(endpoints.requestLog)).length;
  const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, COOKIE_STATES[0]);
  try {
    for (const face of FACES) {
      const id = `fonts/${face.id}`;
      const declared = await visit.page.evaluate(
        (family) => [...document.fonts].some((f) => f.family.replace(/["']/g, '') === family),
        face.family
      );
      if (!stamped || !declared) {
        await report({ id, fixture: false, property: false, detail: `${problem}; ${face.family} declared ${declared}` });
        continue;
      }
      // `document.fonts.load` rejects when a face's file fails to load,
      // so the rejection is this row's failure, never the run's.
      const statuses = await visit.page.evaluate(async (family) => {
        try {
          return (await document.fonts.load(`400 16px "${family}"`)).map((f) => f.status);
        } catch (err) {
          return [`rejected: ${err instanceof Error ? err.message : String(err)}`];
        }
      }, face.family);
      const filePath = `${paths.prefix}fonts/${face.file}`;
      const records = await logSince(endpoints.requestLog, from, (r) => r.path === filePath, ICON_TIMEOUT_MS);
      const fetched = records.some((r) => r.method === 'GET' && r.path === filePath);
      await report({
        id,
        fixture: true,
        property: statuses.length > 0 && statuses.every((s) => s === 'loaded') && fetched,
        detail: `load statuses ${JSON.stringify(statuses)}; GET ${filePath} logged ${fetched}; all ${describe(records)}`
      });
    }

    const id = 'component/display-face';
    const family = await computedProperty(visit.page, FIXTURE.display, 'font-family');
    const weight = await computedProperty(visit.page, FIXTURE.display, 'font-weight');
    const loaded = await visit.page.evaluate(() => document.fonts.check('400 16px "Alfa Slab One"'));
    await report({
      id,
      fixture: stamped,
      property: stamped && family.startsWith('"Alfa Slab One"') && weight === '400' && loaded,
      detail: `${problem} .display font-family ${family}, font-weight ${weight} (want "Alfa Slab One" first, 400); face loaded ${loaded}`.trim()
    });
  } finally {
    await context.close();
  }
}

/**
 * `component/page-title`: the label rule sits under the title, a 3px
 * line over a 1px line (VI "Component layer"), and nothing frames it
 * from above.
 */
async function pageTitleRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'component/page-title';
  const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, COOKIE_STATES[0]);
  try {
    const got = await visit.page.evaluate((elId) => {
      const el = document.getElementById(elId);
      if (el === null) throw new Error(`no element #${elId}`);
      const after = getComputedStyle(el, '::after');
      return {
        content: after.content,
        top: after.borderTopWidth,
        bottom: after.borderBottomWidth,
        own: getComputedStyle(el).borderTopWidth
      };
    }, FIXTURE.pageTitle);
    const fixture = stamped && got.content !== 'none';
    await report({
      id,
      fixture,
      property: fixture && got.top === '3px' && got.bottom === '1px' && got.own === '0px',
      detail: `${problem} ::after content ${got.content}, border-top ${got.top}, border-bottom ${got.bottom}; the title's own border-top ${got.own} (want 3px, 1px, 0px)`.trim()
    });
  } finally {
    await context.close();
  }
}

/**
 * `component/reduced-motion/kill-switch`: under `prefers-reduced-motion:
 * reduce`, no element of the closed set runs a transition or an
 * animation longer than 1ms, DaisyUI's own included (W9.3). The
 * fixture state is the default button declaring DaisyUI's transition,
 * so the row cannot pass on a set with nothing to stop.
 */
async function killSwitchRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'component/reduced-motion/kill-switch';
  const { context, visit, stamped, problem } = await openThemedFixture(browser, endpoints, COOKIE_STATES[0], PAGES[0], {
    reducedMotion: 'reduce'
  });
  try {
    const got = await visit.page.evaluate((btnId) => {
      const seconds = (list: string) => list.split(',').map((d) => parseFloat(d));
      const longest: { what: string; s: number }[] = [];
      for (const el of document.querySelectorAll('main *')) {
        for (const pseudo of [null, '::before', '::after']) {
          const style = getComputedStyle(el, pseudo);
          const worst = Math.max(...seconds(style.transitionDuration), ...seconds(style.animationDuration));
          if (worst > 0.001) longest.push({ what: `${el.id || el.tagName.toLowerCase()}${pseudo ?? ''}`, s: worst });
        }
      }
      const btn = document.getElementById(btnId);
      return {
        reduce: matchMedia('(prefers-reduced-motion: reduce)').matches,
        property: btn === null ? '' : getComputedStyle(btn).transitionProperty,
        longest
      };
    }, FIXTURE.btnDefault);
    const fixture = stamped && got.reduce && got.property.includes('box-shadow');
    await report({
      id,
      fixture,
      property: fixture && got.longest.length === 0,
      detail: `${problem} reduce matched ${got.reduce}; default button transition-property ${got.property}; over 1ms: ${JSON.stringify(got.longest.slice(0, 8))}`.trim()
    });
  } finally {
    await context.close();
  }
}

/**
 * An outcome's announcement class (the state table's own "Class"
 * column), read independently of `web/src/lib/api/auth.ts`'s own
 * `outcomeClass`, which cannot be imported here (it imports `deck.ts`
 * through the `$lib` alias, which Node's type-stripping loader cannot
 * resolve outside Vite). Re-deriving it from the state table is what
 * lets this row prove the page's own classification, rather than
 * assuming it.
 */
type OutcomeClass = 'status' | 'warning' | 'error';
function outcomeClass(outcome: string): OutcomeClass {
  if (outcome === 'failed') return 'warning';
  if (outcome === 'sent' || outcome === 'requested' || outcome === 'unknown' || outcome === 'rate_limited' || outcome === 'signed_in') {
    return 'status';
  }
  return 'error';
}

/**
 * Static markup every page must carry regardless of outcome: one
 * `h1`, no `placeholder` attribute, the login page's email input
 * declaring `autocomplete="email"`, none of the forbidden button
 * shapes (`.btn-outline`, `.btn-dash`, a status-modified
 * `.btn-ghost`: DaisyUI 5.7.46 paints their text from an unfixed
 * variable, so their text contrast cannot be trusted), no control
 * painted on `base-300`, and the skip link first in tab order.
 */
async function staticMarkupRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const served of PAGES) {
    const id = `auth/static/${served.name}`;
    const context = await browser.newContext();
    try {
      const visit = await openBooted(context, endpoints.base, served);
      const info = await visit.page.evaluate(() => ({
        h1Count: document.querySelectorAll('h1').length,
        placeholderInputs: document.querySelectorAll('input[placeholder]').length,
        emailAutocomplete: document.querySelector('input[type="email"]')?.getAttribute('autocomplete') ?? null,
        forbiddenButtons: [...document.querySelectorAll('.btn')].filter((el) => {
          const c = el.className;
          return (
            /\bbtn-outline\b/.test(c) ||
            /\bbtn-dash\b/.test(c) ||
            (/\bbtn-ghost\b/.test(c) && /\bbtn-(primary|secondary|accent|neutral|info|success|warning|error)\b/.test(c))
          );
        }).length,
        base300Controls: document.querySelectorAll('[class*="base-300"]').length
      }));
      await visit.page.keyboard.press('Tab');
      const firstFocused = await visit.page.evaluate(() => {
        const el = document.activeElement;
        return {
          tag: el?.tagName ?? null,
          href: el instanceof HTMLAnchorElement ? el.getAttribute('href') : null
        };
      });
      const skipFirst = firstFocused.tag === 'A' && firstFocused.href === '#main';
      const emailOK = served.name !== 'login' || info.emailAutocomplete === 'email';
      await report({
        id,
        fixture: true,
        property:
          info.h1Count === 1 &&
          info.placeholderInputs === 0 &&
          emailOK &&
          info.forbiddenButtons === 0 &&
          info.base300Controls === 0 &&
          skipFirst,
        detail: `h1 count ${info.h1Count}; placeholder inputs ${info.placeholderInputs}; email autocomplete ${JSON.stringify(info.emailAutocomplete)}; forbidden buttons ${info.forbiddenButtons}; base-300 controls ${info.base300Controls}; first focused ${JSON.stringify(firstFocused)}`
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * The login `h1` is the served site name literally, with no element
 * created from it (L2's routed rule proven again in the browser): the
 * fixture carries `"`, `<`, and `&`, so a page that rendered it with
 * `{@html}` would either fail the text match or leave a child element
 * inside the heading.
 */
async function siteNameHeadingRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'auth/login/site-name-heading';
  const context = await browser.newContext();
  try {
    const { visit } = await open(context, new URL('login', endpoints.base).href, 'load');
    const metaPresent = await visit.page.evaluate(() => document.querySelector('meta[name="dubplate-site-name"]') !== null);
    if (!metaPresent) {
      await report({ id, fixture: false, property: false, detail: 'served document carries no dubplate-site-name meta' });
      return;
    }
    let text = '';
    let childCount = -1;
    try {
      const heading = visit.page.getByRole('heading', { level: 1, name: endpoints.siteName, exact: true });
      await heading.waitFor({ timeout: BOOT_TIMEOUT_MS });
      text = (await heading.textContent()) ?? '';
      childCount = await heading.evaluate((el: Element) => el.children.length);
    } catch {
      // leave text and childCount at their failing defaults
    }
    const title = await visit.page.title();
    const wantTitle = pageTitle(TITLES.login, endpoints.siteName);
    await report({
      id,
      fixture: metaPresent,
      property: text === endpoints.siteName && childCount === 0 && title === wantTitle,
      detail: `h1 ${JSON.stringify(text)} (want ${JSON.stringify(endpoints.siteName)}); h1 child elements ${childCount}; title ${JSON.stringify(title)} (want ${JSON.stringify(wantTitle)})`
    });
  } finally {
    await context.close();
  }
}

/** One login-page request-link case: how the answer is produced, and what it spends. */
interface LoginCase {
  id: string;
  address: string;
  mode: 'real' | 'abort' | { status: number; body: Record<string, unknown> };
  /** The outcome the page must classify by; see {@link wantOutcomeText} for the two IDs that share it. */
  outcome: string;
  linkRequests: number;
}

const LOGIN_CASES: LoginCase[] = [
  { id: 'auth/login/sent', address: ADMIN_ADDRESS, mode: { status: 200, body: { outcome: 'sent' } }, outcome: 'sent', linkRequests: 0 },
  {
    id: 'auth/login/requested',
    address: ADMIN_ADDRESS,
    mode: { status: 200, body: { outcome: 'requested' } },
    outcome: 'requested',
    linkRequests: 0
  },
  { id: 'auth/login/unknown-with-admin', address: UNKNOWN_ADDRESS, mode: 'real', outcome: 'unknown', linkRequests: 1 },
  {
    id: 'auth/login/unknown-no-admin',
    address: UNKNOWN_ADDRESS,
    mode: { status: 200, body: { outcome: 'unknown' } },
    outcome: 'unknown',
    linkRequests: 0
  },
  {
    id: 'auth/login/rate-limited',
    address: ADMIN_ADDRESS,
    mode: { status: 429, body: { outcome: 'rate_limited' } },
    outcome: 'rate_limited',
    linkRequests: 0
  },
  { id: 'auth/login/invalid-address', address: INVALID_ADDRESS, mode: 'real', outcome: 'invalid_address', linkRequests: 1 },
  {
    id: 'auth/login/too-large',
    address: ADMIN_ADDRESS,
    mode: { status: 413, body: { outcome: 'too_large' } },
    outcome: 'too_large',
    linkRequests: 0
  },
  {
    id: 'auth/login/malformed',
    address: ADMIN_ADDRESS,
    mode: { status: 400, body: { outcome: 'malformed' } },
    outcome: 'malformed',
    linkRequests: 0
  },
  { id: 'auth/login/failed', address: MEMBER_ADDRESS, mode: 'real', outcome: 'failed', linkRequests: 1 },
  { id: 'auth/login/cross-origin', address: ADMIN_ADDRESS, mode: { status: 403, body: {} }, outcome: 'error', linkRequests: 0 },
  { id: 'auth/login/network-failure', address: ADMIN_ADDRESS, mode: 'abort', outcome: 'error', linkRequests: 0 }
];

/**
 * The text a login-page case must show. `unknown-with-admin` and
 * `unknown-no-admin` share the outcome value `unknown` but not the
 * text, since only the real path's roster carries an admin contact;
 * `cross-origin` and `network-failure` are the two outcomes this
 * script's cases never let the server itself answer, so their text
 * comes from the two synthetic-outcome deck lines directly.
 */
function wantOutcomeText(c: LoginCase, endpoints: Endpoints): string {
  if (c.id === 'auth/login/unknown-with-admin') {
    return unknownAddressLine({ displayName: endpoints.adminDisplayName, address: ADMIN_ADDRESS });
  }
  if (c.id === 'auth/login/unknown-no-admin') return AUTH_OUTCOMES.unknown;
  if (c.id === 'auth/login/cross-origin') return CROSS_ORIGIN_REFUSED;
  if (c.id === 'auth/login/network-failure') return NETWORK_FAILURE;
  return AUTH_OUTCOMES[c.outcome];
}

/**
 * One row per login-page outcome state: a real POST, a
 * `route.fulfill`, or a `route.abort`, per {@link LoginCase.mode}.
 * `#outcome` is the one element either the sent moment's `<section>`
 * or the alert `<div>` carries, so every row reads it the same way
 * regardless of which of
 * the two the outcome renders as.
 */
async function loginOutcomeRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const c of LOGIN_CASES) {
    const id = c.id;
    const context = await browser.newContext();
    try {
      if (c.mode === 'abort') {
        await context.route('**/request-link', (route) => route.abort('failed'));
      } else if (c.mode !== 'real') {
        const { status, body } = c.mode;
        await context.route('**/request-link', (route) =>
          route.fulfill({ status, contentType: 'application/json', body: JSON.stringify(body) })
        );
      }
      const before = c.mode === 'real' ? (await requestLog(endpoints.requestLog)).length : 0;
      const visit = await openBooted(context, endpoints.base, PAGES[0]);
      visit.expectsAuthErrorStatus = true;
      await visit.page.getByLabel(ADD_MEMBER_LABELS.address, { exact: true }).fill(c.address);
      await visit.page.getByRole('button', { name: TITLES.login, exact: true }).click();

      const outcomeEl = visit.page.locator('#outcome');
      let appeared = true;
      try {
        await outcomeEl.first().waitFor({ timeout: BOOT_TIMEOUT_MS });
      } catch {
        appeared = false;
      }
      if (!appeared) {
        await report({ id, fixture: false, property: false, detail: 'the #outcome element never appeared' });
        continue;
      }

      const isMoment = c.outcome === 'sent' || c.outcome === 'requested';
      // The moment's `#outcome` is a `<section>` holding both the
      // heading and the escape-hatch button, so its own textContent
      // includes the button's label too; the heading alone is the
      // outcome text every other case's `#outcome` textContent already is.
      const text = isMoment
        ? ((await outcomeEl.locator('h2').textContent()) ?? '').trim()
        : ((await outcomeEl.textContent()) ?? '').trim();
      const tag = await outcomeEl.evaluate((el) => el.tagName);
      const cls = (await outcomeEl.getAttribute('class')) ?? '';
      const childCount = await outcomeEl.evaluate((el) => el.children.length);
      const outcomeCount = await outcomeEl.count();
      const politeText = (await visit.page.locator('#announce-polite').textContent()) ?? '';
      const assertiveText = (await visit.page.locator('#announce-assertive').textContent()) ?? '';
      const announced = politeText || assertiveText;
      const announcedPoliteness: 'polite' | 'assertive' | 'none' = politeText ? 'polite' : assertiveText ? 'assertive' : 'none';
      const titleText = await visit.page.title();

      const wantText = wantOutcomeText(c, endpoints);
      const wantClass = outcomeClass(c.outcome);
      const wantPoliteness: 'polite' | 'assertive' = wantClass === 'status' ? 'polite' : 'assertive';
      const wantModifier = wantClass === 'warning' ? 'alert-warning' : wantClass === 'error' ? 'alert-error' : null;
      const wantTitle = pageTitle(isMoment ? TITLES.loginSent : TITLES.login, endpoints.siteName);
      const classes = cls.split(/\s+/);
      const classOK = isMoment
        ? tag === 'SECTION'
        : wantModifier === null
          ? classes.includes('alert') && !classes.includes('alert-warning') && !classes.includes('alert-error')
          : classes.includes(wantModifier);
      const childOK = isMoment || childCount === 0;

      let requestsSeen = 0;
      if (c.mode === 'real') {
        const records = await requestLog(endpoints.requestLog);
        requestsSeen = records.slice(before).filter((r) => r.method === 'POST' && r.path.endsWith('/request-link')).length;
      }

      await report({
        id,
        fixture: true,
        property:
          outcomeCount === 1 &&
          text === wantText &&
          classOK &&
          childOK &&
          announced === wantText &&
          announcedPoliteness === wantPoliteness &&
          titleText === wantTitle &&
          (c.mode !== 'real' || requestsSeen === 1),
        detail: `text ${JSON.stringify(text)} (want ${JSON.stringify(wantText)}); #outcome count ${outcomeCount} tag ${tag} class ${JSON.stringify(cls)} children ${childCount}; announced ${JSON.stringify(announced)} as ${announcedPoliteness} (want ${wantPoliteness}); title ${JSON.stringify(titleText)} (want ${JSON.stringify(wantTitle)}); real request-link POSTs seen ${requestsSeen}`
      });
    } finally {
      await context.close();
    }
  }
}

/**
 * After two submits, exactly one outcome element exists: two
 * fulfilled answers in a row, each superseding the last, over one
 * page.
 */
async function loginSingleOutcomeRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'auth/login/single-outcome';
  const context = await browser.newContext();
  try {
    let calls = 0;
    await context.route('**/request-link', (route) => {
      calls += 1;
      const outcome = calls === 1 ? 'rate_limited' : 'too_large';
      return route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify({ outcome }) });
    });
    const visit = await openBooted(context, endpoints.base, PAGES[0]);
    visit.expectsAuthErrorStatus = true;
    const submit = async () => {
      await visit.page.getByLabel(ADD_MEMBER_LABELS.address, { exact: true }).fill(ADMIN_ADDRESS);
      await visit.page.getByRole('button', { name: TITLES.login, exact: true }).click();
    };

    await submit();
    await visit.page.locator('#outcome').first().waitFor({ timeout: BOOT_TIMEOUT_MS });
    const firstText = (await visit.page.locator('#outcome').textContent()) ?? '';
    const firstCount = await visit.page.locator('#outcome').count();

    await submit();
    await visit.page
      .waitForFunction((want) => document.querySelector('#outcome')?.textContent === want, AUTH_OUTCOMES.too_large, {
        timeout: BOOT_TIMEOUT_MS
      })
      .catch(() => undefined);
    const secondText = (await visit.page.locator('#outcome').textContent()) ?? '';
    const secondCount = await visit.page.locator('#outcome').count();

    await report({
      id,
      fixture: true,
      property:
        firstCount === 1 && secondCount === 1 && firstText === AUTH_OUTCOMES.rate_limited && secondText === AUTH_OUTCOMES.too_large,
      detail: `first #outcome count ${firstCount} text ${JSON.stringify(firstText)}; second #outcome count ${secondCount} text ${JSON.stringify(secondText)}`
    });
  } finally {
    await context.close();
  }
}

/**
 * The token never reaches `location.search` or `history.state` once
 * the page has settled, no request carries it, and exactly one
 * `consume-link` POST happens, only after the click.
 */
async function interstitialOrderingRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'auth/interstitial/token-order';
  const context = await browser.newContext();
  try {
    const token = new URL(endpoints.loginLinkOrdering).searchParams.get('t') ?? '';
    const { visit } = await open(context, endpoints.loginLinkOrdering, 'load');
    await visit.page.getByRole('heading', { level: 1, name: TITLES.loginLink, exact: true }).waitFor({ timeout: BOOT_TIMEOUT_MS });
    // The mailed link's own GET necessarily carries the token in its
    // query (that is how the token reaches the client at all); every
    // request logged from here on must not.
    const sinceLoad = (await requestLog(endpoints.requestLog)).length;

    const readState = () => visit.page.evaluate(() => ({ search: location.search, state: JSON.stringify(history.state) }));
    const clean = (s: { search: string; state: string }) => s.search === '' && (token === '' || !s.state.includes(token));

    const afterLoad = await readState();
    await visit.page.locator('a[href="#main"]').first().focus();
    const afterSkip = await readState();

    const consumePOSTs = (records: LoggedRequest[]) =>
      records.filter((r) => r.method === 'POST' && r.path.endsWith('/consume-link'));
    const beforeClick = consumePOSTs(await requestLog(endpoints.requestLog)).length;

    await visit.page.getByRole('button', { name: INTERSTITIAL_CONTROL, exact: true }).click();
    await visit.page.locator('#outcome').first().waitFor({ timeout: BOOT_TIMEOUT_MS });
    const afterClick = await readState();

    const records = await requestLog(endpoints.requestLog);
    const afterClickCount = consumePOSTs(records).length;
    const anyCarriesToken =
      token.length > 0 &&
      records
        .slice(sinceLoad)
        .some((r) => r.query.includes(token) || Object.values(r.headers).flat().some((v) => v.includes(token)));
    const text = (await visit.page.locator('#outcome').textContent()) ?? '';

    await report({
      id,
      fixture: token.length > 0,
      property:
        clean(afterLoad) &&
        clean(afterSkip) &&
        clean(afterClick) &&
        beforeClick === 0 &&
        afterClickCount === 1 &&
        !anyCarriesToken &&
        text === AUTH_OUTCOMES.signed_in,
      detail: `afterLoad ${JSON.stringify(afterLoad)}; afterSkip ${JSON.stringify(afterSkip)}; afterClick ${JSON.stringify(afterClick)}; consume-link POSTs before ${beforeClick} after ${afterClickCount}; any request carries the token ${anyCarriesToken}; outcome text ${JSON.stringify(text)}`
    });
  } finally {
    await context.close();
  }
}

/** The state-table's own `signed_in` row: its own minted link, so the token-ordering row's own link is spent once, on that row alone. */
async function interstitialSignedInRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'auth/interstitial/signed-in';
  const context = await browser.newContext();
  try {
    const { visit } = await open(context, endpoints.loginLinkSignedIn, 'load');
    await visit.page.getByRole('heading', { level: 1, name: TITLES.loginLink, exact: true }).waitFor({ timeout: BOOT_TIMEOUT_MS });
    await visit.page.getByRole('button', { name: INTERSTITIAL_CONTROL, exact: true }).click();

    const heading = visit.page.getByRole('heading', { level: 1, name: TITLES.loginLinkSignedIn, exact: true });
    let settled = true;
    try {
      await heading.waitFor({ timeout: BOOT_TIMEOUT_MS });
    } catch {
      settled = false;
    }
    if (!settled) {
      await report({ id, fixture: false, property: false, detail: 'the signed-in heading never appeared' });
      return;
    }

    const text = (await visit.page.locator('#outcome').textContent()) ?? '';
    const polite = (await visit.page.locator('#announce-polite').textContent()) ?? '';
    const title = await visit.page.title();
    const focusedTag = await visit.page.evaluate(() => document.activeElement?.tagName ?? null);
    const wantTitle = pageTitle(TITLES.loginLinkSignedIn, endpoints.siteName);

    await report({
      id,
      fixture: settled,
      property: text === AUTH_OUTCOMES.signed_in && polite === AUTH_OUTCOMES.signed_in && title === wantTitle && focusedTag === 'H1',
      detail: `outcome text ${JSON.stringify(text)}; polite region ${JSON.stringify(polite)}; title ${JSON.stringify(title)} (want ${JSON.stringify(wantTitle)}); focused element ${focusedTag}`
    });
  } finally {
    await context.close();
  }
}

/** A reload of the stripped URL (no token at all): the refused copy, and no request. */
async function interstitialRefusedReloadRow(browser: Browser, endpoints: Endpoints): Promise<void> {
  const id = 'auth/interstitial/refused-reload';
  const context = await browser.newContext();
  try {
    const from = (await requestLog(endpoints.requestLog)).length;
    const visit = await openBooted(context, endpoints.base, PAGES[1]);
    const text = (await visit.page.locator('#outcome').textContent()) ?? '';
    const linkHref = await visit.page
      .locator('#outcome a')
      .getAttribute('href', { timeout: 5_000 })
      .catch(() => null);
    const records = (await requestLog(endpoints.requestLog)).slice(from);
    const consumeSent = records.some((r) => r.method === 'POST' && r.path.endsWith('/consume-link'));
    await report({
      id,
      fixture: true,
      property: text === AUTH_OUTCOMES.refused && !consumeSent && linkHref !== null,
      detail: `outcome text ${JSON.stringify(text)}; consume-link POST sent ${consumeSent}; sign-in link href ${linkHref}`
    });
  } finally {
    await context.close();
  }
}

/** One consume-link case answered by `route.fulfill`, never the real handler. */
interface ConsumeCase {
  id: string;
  outcome: string;
  status: number;
}

const CONSUME_CASES: ConsumeCase[] = [
  { id: 'auth/interstitial/rate-limited', outcome: 'rate_limited', status: 429 },
  { id: 'auth/interstitial/too-large', outcome: 'too_large', status: 413 },
  { id: 'auth/interstitial/malformed', outcome: 'malformed', status: 400 },
  { id: 'auth/interstitial/failed', outcome: 'failed', status: 500 }
];

/**
 * The interstitial's four fulfilled consume-link outcomes: the `h1`
 * stays "Confirm sign-in" (the deck names no separate title for these,
 * unlike `signed_in` and `refused`), and the control's own alert
 * appears beside it.
 */
async function interstitialFulfilledRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const c of CONSUME_CASES) {
    const id = c.id;
    const context = await browser.newContext();
    try {
      await context.route('**/consume-link', (route) =>
        route.fulfill({ status: c.status, contentType: 'application/json', body: JSON.stringify({ outcome: c.outcome }) })
      );
      const url = new URL(`login/link?t=fixture-${c.id}`, endpoints.base).href;
      const { visit } = await open(context, url, 'load');
      visit.expectsAuthErrorStatus = true;
      await visit.page.getByRole('heading', { level: 1, name: TITLES.loginLink, exact: true }).waitFor({ timeout: BOOT_TIMEOUT_MS });
      await visit.page.getByRole('button', { name: INTERSTITIAL_CONTROL, exact: true }).click();
      await visit.page.locator('#outcome').first().waitFor({ timeout: BOOT_TIMEOUT_MS });

      const text = (await visit.page.locator('#outcome').textContent()) ?? '';
      const cls = (await visit.page.locator('#outcome').getAttribute('class')) ?? '';
      const classes = cls.split(/\s+/);
      const wantClass = outcomeClass(c.outcome);
      const wantModifier = wantClass === 'warning' ? 'alert-warning' : wantClass === 'error' ? 'alert-error' : null;
      const classOK =
        wantModifier === null
          ? classes.includes('alert') && !classes.includes('alert-warning') && !classes.includes('alert-error')
          : classes.includes(wantModifier);
      const politeText = (await visit.page.locator('#announce-polite').textContent()) ?? '';
      const assertiveText = (await visit.page.locator('#announce-assertive').textContent()) ?? '';
      const announced = politeText || assertiveText;
      const announcedPoliteness: 'polite' | 'assertive' | 'none' = politeText ? 'polite' : assertiveText ? 'assertive' : 'none';
      const wantPoliteness: 'polite' | 'assertive' = wantClass === 'status' ? 'polite' : 'assertive';
      const headingUnchanged = await visit.page.getByRole('heading', { level: 1, name: TITLES.loginLink, exact: true }).count();

      await report({
        id,
        fixture: true,
        property:
          text === AUTH_OUTCOMES[c.outcome] &&
          classOK &&
          announced === AUTH_OUTCOMES[c.outcome] &&
          announcedPoliteness === wantPoliteness &&
          headingUnchanged === 1,
        detail: `text ${JSON.stringify(text)}; class ${JSON.stringify(cls)} (want modifier ${wantModifier}); announced ${JSON.stringify(announced)} as ${announcedPoliteness} (want ${wantPoliteness}); h1 unchanged ${headingUnchanged === 1}`
      });
    } finally {
      await context.close();
    }
  }
}

async function main(): Promise<void> {
  const endpoints: Endpoints = {
    base: requiredEnv('DUBPLATE_RENDERED_BASE_URL'),
    requestLog: requiredEnv('DUBPLATE_RENDERED_REQUEST_LOG'),
    brandingBase: requiredEnv('DUBPLATE_RENDERED_BRANDING_BASE_URL'),
    operatorBase: requiredEnv('DUBPLATE_RENDERED_OPERATOR_BASE_URL'),
    brandingPrimary: requiredEnv('DUBPLATE_RENDERED_BRANDING_PRIMARY'),
    operatorPrimary: requiredEnv('DUBPLATE_RENDERED_OPERATOR_PRIMARY'),
    siteName: requiredEnv('DUBPLATE_RENDERED_SITE_NAME'),
    adminDisplayName: requiredEnv('DUBPLATE_RENDERED_ADMIN_DISPLAY_NAME'),
    loginLinkOrdering: requiredEnv('DUBPLATE_RENDERED_LOGIN_LINK_ORDERING'),
    loginLinkSignedIn: requiredEnv('DUBPLATE_RENDERED_LOGIN_LINK_SIGNED_IN')
  };
  PAGES[0].heading = endpoints.siteName;
  PAGES[1].heading = TITLES.loginLinkRefused;
  prefixWatch = { log: endpoints.requestLog, prefix: new URL(endpoints.base).pathname, seen: 0 };
  const browser = await launch();
  try {
    await firstPaintRows(browser, endpoints);
    await g7ComputedRows(browser, endpoints);
    await g10Rows(browser, endpoints);
    await cspCleanRows(browser, endpoints);
    await violationListenerRow(browser, endpoints);
    await themeColorRows(browser, endpoints);
    await iconRows(browser, endpoints);
    await clientNavigationRow(browser, endpoints);
    await noFaviconProbeRow(endpoints);
    await focusRows(browser, endpoints);
    await pressedRows(browser, endpoints);
    await pressedFlatRow(browser, endpoints);
    await killSwitchRow(browser, endpoints);
    await surfaceRows(browser, endpoints);
    await badgeRows(browser, endpoints);
    await inputStateRows(browser, endpoints);
    await linkRows(browser, endpoints);
    await utilityWinsRow(browser, endpoints);
    await fontRows(browser, endpoints);
    await pageTitleRow(browser, endpoints);
    await staticMarkupRows(browser, endpoints);
    await siteNameHeadingRow(browser, endpoints);
    await loginOutcomeRows(browser, endpoints);
    await loginSingleOutcomeRow(browser, endpoints);
    await interstitialOrderingRow(browser, endpoints);
    await interstitialSignedInRow(browser, endpoints);
    await interstitialRefusedReloadRow(browser, endpoints);
    await interstitialFulfilledRows(browser, endpoints);
    refuseUnclean();
    await refuseLateLeaks(endpoints);
  } finally {
    await browser.close();
  }
}

main().catch((err: unknown) => {
  process.stderr.write(`rendered: ${message(err)}\n`);
  process.exit(1);
});
