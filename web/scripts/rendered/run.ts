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
 * docs/superpowers/specs/2026-09-24-visual-identity-design.md (:128-131):
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
        report({ id, fixture, property: false, detail: `served stamp ${stamp}; metas ${JSON.stringify(metas)}` });
        continue;
      }
      await visit.page.emulateMedia({ colorScheme: scheme });
      const active = await visit.page.evaluate(() =>
        [...document.querySelectorAll<HTMLMetaElement>('meta[name="theme-color"]')]
          .filter((m) => !m.media || matchMedia(m.media).matches)
          .map((m) => m.content)
      );
      report({
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
      report({
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
      report({
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
    report({
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
  report({
    id,
    fixture: icons.length > 0,
    property: probes.length === 0,
    detail: `${icons.length} icon request(s) in the run; /favicon.ico requests ${describe(probes)}`
  });
}

/**
 * Task 5's closed fixture control set (VI's component layer and
 * printed surface): a primary, secondary, warning, and neutral button,
 * a ghost button, a default button, an input, a card, an alert, and a
 * badge, each with the bundle's own DaisyUI classes. The IDs are
 * literal strings so Tailwind's content scan of this file emits the
 * classes' CSS into the built bundle the same way `cspCleanRows`'
 * injected checkbox already does.
 */
const FIXTURE = {
  btnDefault: 'rendered-fx-btn-default',
  btnPrimary: 'rendered-fx-btn-primary',
  btnSecondary: 'rendered-fx-btn-secondary',
  btnWarning: 'rendered-fx-btn-warning',
  btnNeutral: 'rendered-fx-btn-neutral',
  btnGhost: 'rendered-fx-btn-ghost',
  input: 'rendered-fx-input',
  card: 'rendered-fx-card',
  alert: 'rendered-fx-alert',
  badge: 'rendered-fx-badge'
} as const;

/**
 * Appends the closed set to the booted page's `main`, under a wrapper
 * with room above it so a focus ring's outer edge never samples
 * outside the viewport. Nothing here ships; the elements exist only
 * for this run.
 */
async function injectFixtureControls(page: Page): Promise<void> {
  await page.evaluate((ids: typeof FIXTURE) => {
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
    const button = (id: string, extra: string) => {
      const el = document.createElement('button');
      el.id = id;
      el.className = `btn ${extra}`.trim();
      el.textContent = id;
      wrap.append(el);
    };
    button(ids.btnDefault, '');
    button(ids.btnPrimary, 'btn-primary');
    button(ids.btnSecondary, 'btn-secondary');
    button(ids.btnWarning, 'btn-warning');
    button(ids.btnNeutral, 'btn-neutral');
    button(ids.btnGhost, 'btn-ghost');
    const input = document.createElement('input');
    input.id = ids.input;
    input.type = 'text';
    input.className = 'input';
    wrap.append(input);
    const card = document.createElement('div');
    card.id = ids.card;
    card.className = 'card';
    card.innerHTML = '<div class="card-body">fixture card</div>';
    wrap.append(card);
    const alert = document.createElement('div');
    alert.id = ids.alert;
    alert.className = 'alert';
    alert.textContent = 'fixture alert';
    wrap.append(alert);
    const badge = document.createElement('span');
    badge.id = ids.badge;
    badge.className = 'badge badge-warning';
    badge.textContent = 'fixture';
    wrap.append(badge);
    main.append(wrap);
  }, FIXTURE);
}

/** Opens the login page under `cookie`'s stamp and injects the closed set. */
async function openThemedFixture(
  browser: Browser,
  endpoints: Endpoints,
  cookie: 'label' | 'label-night'
): Promise<{ context: BrowserContext; visit: Visit }> {
  const host = new URL(endpoints.base).hostname;
  const context = await browser.newContext();
  await context.addCookies([{ name: THEME_COOKIE, value: cookie, domain: host, path: '/', secure: true }]);
  const visit = await openBooted(context, endpoints.base, PAGES[0]);
  await injectFixtureControls(visit.page);
  return { context, visit };
}

/** `--ink-primary`'s two theme values (VI's ink table). */
const INK_PRIMARY = { label: '#9d2f26', 'label-night': '#ec7b6a' } as const;

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

interface FocusedInfo {
  label: string;
  probe: string | null;
  outlineStyle: string;
  outlineWidth: string;
  outlineOffset: string;
  outlineColor: string;
  rect: { x: number; y: number; width: number };
}

/** `document.activeElement`'s outline and geometry, or `null` once `Tab` leaves the page (a wrapped cycle). */
async function focusedInfo(page: Page): Promise<FocusedInfo | null> {
  return page.evaluate((attr: string) => {
    const el = document.activeElement as HTMLElement | null;
    if (el === null || el === document.body) return null;
    const style = getComputedStyle(el);
    const rect = el.getBoundingClientRect();
    const label = el.id || `${el.tagName.toLowerCase()}:${(el.textContent ?? '').trim().slice(0, 24)}`;
    return {
      label,
      probe: el.getAttribute(attr),
      outlineStyle: style.outlineStyle,
      outlineWidth: style.outlineWidth,
      outlineOffset: style.outlineOffset,
      outlineColor: style.outlineColor,
      rect: { x: rect.x, y: rect.y, width: rect.width }
    };
  }, TAB_PROBE_ATTR);
}

/**
 * `component/focus/<page>/<theme>`: every tabbable element the page
 * serves, enumerated by DOM query, plus the injected closed set
 * (Ruling 4's icon-only exception does not apply, since the set
 * carries no icon-only control). A page must serve a nonzero tabbable
 * count and every injected control by role before any ring is judged.
 * Each element's ring is read back at `--ink-primary` by tabbing to
 * it, and the pixel just inside the ring (the offset gap) and the one
 * just outside it each clear 3:1 against the ring, both sampled above
 * the element's top edge, where the printed surface casts no shadow
 * (the shadow's own offset is bottom-right only). The walk also proves
 * coverage by identity: every element {@link enumerateTabbable} finds
 * must be focused at least once (a revisit after a wrap counts once),
 * and the row fails, naming whichever queried element the walk never
 * reached, rather than only comparing a visited count against the
 * query's own total.
 */
async function focusRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const theme of ['label', 'label-night'] as const) {
    for (const served of PAGES) {
      const id = `component/focus/${served.name}/${theme}`;
      const host = new URL(endpoints.base).hostname;
      const context = await browser.newContext();
      await context.addCookies([{ name: THEME_COOKIE, value: theme, domain: host, path: '/', secure: true }]);
      try {
        const visit = await openBooted(context, endpoints.base, served);
        await injectFixtureControls(visit.page);

        const targets = await enumerateTabbable(visit.page);
        const total = targets.length;
        if (total === 0) {
          report({ id, fixture: false, property: false, detail: 'no tabbable element found on the page' });
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
        if ((await visit.page.getByRole('textbox').count()) === 0) missingRoles.push(FIXTURE.input);
        if (missingRoles.length > 0) {
          report({ id, fixture: false, property: false, detail: `missing by role: ${missingRoles.join(', ')}` });
          continue;
        }

        await visit.page.evaluate(() => (document.activeElement as HTMLElement | null)?.blur());
        const want = hexRGBA(INK_PRIMARY[theme]);
        const problems: string[] = [];
        const visitedProbes = new Set<string>();
        let visited = 0;
        // A margin of 2 over `total` absorbs a cycle that wraps back
        // to the first element before this bound stops it; a landing
        // back on `document.body` (`focusedInfo` returning `null`)
        // ends the loop early regardless.
        for (let i = 0; i < total + 2; i++) {
          await visit.page.keyboard.press('Tab');
          const focused = await focusedInfo(visit.page);
          if (focused === null) break;
          visited++;
          if (focused.probe !== null) visitedProbes.add(focused.probe);
          const ring = await paintedColor(visit.page, focused.outlineColor);
          if (
            focused.outlineStyle !== 'solid' ||
            focused.outlineWidth !== '2px' ||
            focused.outlineOffset !== '2px' ||
            !sameRGBA(ring, want)
          ) {
            problems.push(
              `${focused.label}: outline ${focused.outlineStyle} ${focused.outlineWidth} ${focused.outlineColor} offset ${focused.outlineOffset} (want solid 2px ${INK_PRIMARY[theme]} offset 2px)`
            );
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
          report({ id, fixture: false, property: false, detail: `tabbing found no focusable element (of ${total} tabbable by query)` });
          continue;
        }
        const missing = targets.filter((t) => !visitedProbes.has(t.id));
        if (missing.length > 0) {
          problems.push(`never focused: ${missing.map((t) => t.label).join(', ')} (of ${total} queried)`);
        }
        report({
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
        report({ id, fixture: false, property: false, detail: 'the default button has no box' });
        continue;
      }
      await visit.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
      await visit.page.mouse.down();
      // DaisyUI's `.btn` transitions `transform` over 0.2s; a read
      // immediately after the press lands mid-transition, so this
      // waits out the transition before reading the settled position
      // and the box-shadow, which does not itself transition but is
      // read alongside the box for one settled snapshot of the
      // pressed state.
      await visit.page.waitForTimeout(300);
      const pressedShadow = await computedProperty(visit.page, FIXTURE.btnDefault, 'box-shadow');
      const pressedBox = await locator.boundingBox();
      await visit.page.mouse.up();
      if (pressedShadow !== 'none') {
        report({
          id,
          fixture: false,
          property: false,
          detail: `the button's box-shadow read ${pressedShadow} while the mouse was held down, so its pressed state was never reached (want none)`
        });
        continue;
      }
      if (pressedBox === null) {
        report({ id, fixture: false, property: false, detail: 'the default button lost its box while pressed' });
        continue;
      }
      const delta = { x: pressedBox.x - box.x, y: pressedBox.y - box.y };
      const want = reduce ? { x: 0, y: 0 } : { x: 2, y: 2 };
      const moved = closeTo(delta.x, want.x) && closeTo(delta.y, want.y);
      report({
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
    report({ id, fixture: geometry !== null, property: false, detail: `box-shadow "${boxShadow}" (want geometry ${wantGeometry})` });
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
  report({
    id,
    fixture: true,
    property: shadowOK && bgOK,
    detail: `geometry ${geometry}; (${insideX},${midY})=[${inside.join(',')}] want shadow ${wantShadowHex}; (${outsideX},${midY})=[${outside.join(',')}] want bg ${BASE_100[theme]}`
  });
}

async function surfaceRows(browser: Browser, endpoints: Endpoints): Promise<void> {
  for (const theme of ['label', 'label-night'] as const) {
    const { context, visit } = await openThemedFixture(browser, endpoints, theme);
    try {
      await shadowRow(visit.page, 'button', theme, FIXTURE.btnDefault, 2, SURFACE_EXPECTED.buttonShadow[theme]);
      await shadowRow(visit.page, 'card', theme, FIXTURE.card, 3, SURFACE_EXPECTED.cardShadow[theme]);
      await shadowRow(visit.page, 'alert', theme, FIXTURE.alert, 3, SURFACE_EXPECTED.cardShadow[theme]);

      const defaultBorder = await paintedColor(visit.page, await computedProperty(visit.page, FIXTURE.btnDefault, 'border-color'));
      const cardBorder = await paintedColor(visit.page, await computedProperty(visit.page, FIXTURE.card, 'border-color'));
      const primaryBorder = await paintedColor(visit.page, await computedProperty(visit.page, FIXTURE.btnPrimary, 'border-color'));
      const alertBorder = await paintedColor(visit.page, await computedProperty(visit.page, FIXTURE.alert, 'border-color'));
      const wantBase = hexRGBA(SURFACE_EXPECTED.keylineBase[theme]);
      const wantFill = hexRGBA(SURFACE_EXPECTED.keylineFill[theme]);
      report({
        id: `component/surface/keyline-base/${theme}`,
        fixture: true,
        property: sameRGBA(defaultBorder, wantBase) && sameRGBA(cardBorder, wantBase),
        detail: `default button border [${defaultBorder.join(',')}], card border [${cardBorder.join(',')}] (want ${SURFACE_EXPECTED.keylineBase[theme]})`
      });
      report({
        id: `component/surface/keyline-fill/${theme}`,
        fixture: true,
        property: sameRGBA(primaryBorder, wantFill) && sameRGBA(alertBorder, wantFill),
        detail: `primary button border [${primaryBorder.join(',')}], alert border [${alertBorder.join(',')}] (want ${SURFACE_EXPECTED.keylineFill[theme]})`
      });

      const defaultBorderWidth = await computedProperty(visit.page, FIXTURE.btnDefault, 'border-top-width');
      const inputBorderWidth = await computedProperty(visit.page, FIXTURE.input, 'border-top-width');
      report({
        id: `component/surface/border-width/${theme}`,
        fixture: true,
        property: defaultBorderWidth === '2px' && inputBorderWidth === '2px',
        detail: `default button border-top-width ${defaultBorderWidth}, input border-top-width ${inputBorderWidth} (want 2px, 2px)`
      });

      const ghostShadow = await computedProperty(visit.page, FIXTURE.btnGhost, 'box-shadow');
      const ghostBorderStyle = await computedProperty(visit.page, FIXTURE.btnGhost, 'border-style');
      report({
        id: `component/surface/ghost/${theme}`,
        fixture: true,
        property: ghostShadow === 'none' && ghostBorderStyle === 'none',
        detail: `box-shadow ${ghostShadow}, border-style ${ghostBorderStyle} (want none, none)`
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
    operatorPrimary: requiredEnv('DUBPLATE_RENDERED_OPERATOR_PRIMARY')
  };
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
    await surfaceRows(browser, endpoints);
    refuseUnclean();
  } finally {
    await browser.close();
  }
}

main().catch((err: unknown) => {
  process.stderr.write(`rendered: ${message(err)}\n`);
  process.exit(1);
});
