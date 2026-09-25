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
import { decodePNG, pixelHex } from './png.ts';

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

const PAGES: ServedPage[] = [
  { name: 'login', path: 'login', heading: 'Log in' },
  { name: 'login-link', path: 'login/link', heading: 'Confirm login' }
];

/**
 * First paint per theme cookie and emulated preference. Every expected
 * pixel is base-100 of the palette the stamp should select, quoted from
 * docs/superpowers/specs/2026-09-24-visual-identity-design.md: `label`
 * `#f4ead3` (:71) and `label-night` `#1c1916` (:85). With no cookie the
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
}

const visits: Visit[] = [];

function requiredEnv(name: string): string {
  const value = process.env[name];
  if (!value) throw new Refusal(`${name} is not set; the rendered leg runs only under its Go test`);
  return value;
}

function report(row: Row): void {
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
  const visit: Visit = { url, page, consoleErrors: [], violations: [], expectedBlock: null };
  page.on('console', (msg) => {
    if (msg.type() !== 'error') return;
    const source = msg.location().url;
    // A document that declares no icon makes Chromium probe the
    // origin's /favicon.ico, outside the owned prefix, where the shell
    // answers 404. That probe is the browser's, not the page's.
    if (source === new URL('/favicon.ico', url).href) return;
    visit.consoleErrors.push(`${msg.text()} (${source})`);
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
 * page's color syntax is never parsed here. The canvas is filled transparent
 * first, so a value the canvas refuses reads back with alpha 0 and can
 * never match an opaque expectation.
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
    return await visit.page.evaluate((): RGBA => {
      const probe = document.createElement('div');
      probe.style.backgroundColor = 'var(--color-primary)';
      document.documentElement.append(probe);
      const resolved = getComputedStyle(probe).backgroundColor;
      probe.remove();
      const canvas = document.createElement('canvas');
      canvas.width = 1;
      canvas.height = 1;
      const ctx = canvas.getContext('2d', { colorSpace: 'srgb' });
      if (ctx === null) throw new Error('the canvas has no 2d context');
      ctx.fillStyle = 'transparent';
      ctx.fillStyle = resolved;
      ctx.fillRect(0, 0, 1, 1);
      const [r, g, b, a] = ctx.getImageData(0, 0, 1, 1, { colorSpace: 'srgb' }).data;
      return [r, g, b, a];
    });
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
        report({
          id,
          fixture,
          property: false,
          detail: `branding.css href ${tier.brandingHref}; content ${JSON.stringify(tier.brandingCSS)}`
        });
        continue;
      }
      const got = await paintedPrimary(browser, target.base, state.cookie, state.scheme);
      const want = hexRGBA(target.want);
      report({
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
    report({
      id: opID,
      fixture: opFixture,
      property: false,
      detail: `branding.css ${JSON.stringify(opTier.brandingCSS)}; operator.css href ${opTier.operatorHref} content ${JSON.stringify(opTier.operatorCSS)}`
    });
  } else {
    const got = await paintedPrimary(browser, endpoints.operatorBase, 'label', 'light');
    const want = hexRGBA(endpoints.operatorPrimary);
    report({
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
        report({ id, fixture, property: false, detail: `served stamp ${stamp}` });
        continue;
      }
      const got = await visit.page.evaluate((names: string[]) => {
        const cs = getComputedStyle(document.documentElement);
        return Object.fromEntries(names.map((n) => [n, cs.getPropertyValue(n).trim()]));
      }, Object.keys(STRUCTURAL));
      const mismatches = Object.entries(STRUCTURAL).filter(([name, want]) => got[name] !== want);
      report({
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
      report({ id, fixture, property: false, detail: `served stamp ${stamp}, want a dubplate-* stamp` });
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
      report({
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
        report({
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
      report({
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
      report({
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
    report({
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

function isExpectedConsoleError(visit: Visit, text: string): boolean {
  return (
    visit.expectedBlock !== null &&
    text.startsWith(`Loading the image '${visit.expectedBlock}' violates the following Content Security Policy directive: "img-src `)
  );
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

async function main(): Promise<void> {
  const endpoints: Endpoints = {
    base: requiredEnv('DUBPLATE_RENDERED_BASE_URL'),
    requestLog: requiredEnv('DUBPLATE_RENDERED_REQUEST_LOG'),
    brandingBase: requiredEnv('DUBPLATE_RENDERED_BRANDING_BASE_URL'),
    operatorBase: requiredEnv('DUBPLATE_RENDERED_OPERATOR_BASE_URL'),
    brandingPrimary: requiredEnv('DUBPLATE_RENDERED_BRANDING_PRIMARY'),
    operatorPrimary: requiredEnv('DUBPLATE_RENDERED_OPERATOR_PRIMARY')
  };
  const browser = await launch();
  try {
    await firstPaintRows(browser, endpoints);
    await g7ComputedRows(browser, endpoints);
    await g10Rows(browser, endpoints);
    await cspCleanRows(browser, endpoints);
    await violationListenerRow(browser, endpoints);
    refuseUnclean();
  } finally {
    await browser.close();
  }
}

main().catch((err: unknown) => {
  process.stderr.write(`rendered: ${message(err)}\n`);
  process.exit(1);
});
