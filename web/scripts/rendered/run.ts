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
 * owned prefix with a trailing slash), and `DUBPLATE_RENDERED_REQUEST_LOG`,
 * the URL of the Go side's log of every request the server received.
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
  /** The page raises a violation on purpose, so the final sweep skips it. */
  expectsViolation: boolean;
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

/** The two endpoints the Go test hands the script. */
interface Endpoints {
  base: string;
  requestLog: string;
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
  const visit: Visit = { url, page, consoleErrors: [], violations: [], expectsViolation: false };
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
    visit.expectsViolation = true;
    const before = await violations(visit.page);
    const fixture = before.length === 0;
    let seen: Violation[] = [];
    if (fixture) {
      await visit.page.evaluate(() => {
        const img = document.createElement('img');
        img.src = 'https://example.test/rendered-probe.png';
        document.body.append(img);
      });
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
    }
    report({
      id,
      fixture,
      property: seen.some((v) => v.directive === 'img-src' && v.blocked.startsWith('https://example.test/')),
      detail: `violations before injection ${JSON.stringify(before)}; after ${JSON.stringify(seen)}`
    });
  } finally {
    await context.close();
  }
}

/** Refuses the run when any page but a deliberate one reported a problem. */
function refuseUnclean(): void {
  for (const visit of visits) {
    if (visit.expectsViolation) continue;
    if (visit.violations.length > 0) {
      const v = visit.violations[0];
      throw new Refusal(`${visit.url}: securitypolicyviolation: ${v.directive} blocked ${v.blocked}`);
    }
    if (visit.consoleErrors.length > 0) {
      throw new Refusal(`${visit.url}: console error: ${visit.consoleErrors[0]}`);
    }
  }
}

async function main(): Promise<void> {
  const endpoints: Endpoints = {
    base: requiredEnv('DUBPLATE_RENDERED_BASE_URL'),
    requestLog: requiredEnv('DUBPLATE_RENDERED_REQUEST_LOG')
  };
  const browser = await launch();
  try {
    await firstPaintRows(browser, endpoints);
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
