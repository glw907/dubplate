/**
 * The root-absolute check. The server mounts the app under an operator's
 * prefix, so every page link, fetch, navigation, and asset reference must
 * compose from the runtime base; a URL that starts at `/` escapes the
 * prefix. `src` is scanned for root-absolute `href`, `src`, `srcset`,
 * `action`, and `formaction` attributes, and for the navigation and
 * request sinks that take a URL literal: `fetch`, `goto`, a `location`
 * write, `window.open`, `new URL`, `new Request`, and `sendBeacon`. A
 * URL built in a variable escapes any source scan, so the rendered leg
 * also fails every row on a request that leaves the prefix. `src`'s
 * CSS is not scanned, because an `@font-face` source must be
 * root-absolute for Vite to rewrite it; the built stylesheets are
 * scanned instead, since a root-absolute `url()` that survives the
 * build is the real defect. An absent or empty `src`, or a build with
 * no stylesheet, fails rather than passing on an empty iteration.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { stripComments } from './g1.ts';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';
const BUILT_CSS_ROOT = join('build', '_app', 'immutable', 'assets');
const SCANNED_EXTENSIONS = ['.svelte', '.ts', '.js', '.html'];

/**
 * Each opening quote may follow a Svelte expression brace, so
 * `href={'/x'}` and ``href={`/x`}`` count as well as `href="/x"`. Each
 * match runs on to the URL's end, so a failure names the URL. A
 * `srcset` lists candidates separated by commas, so its second pattern
 * catches a root-absolute URL in any later candidate.
 */
const SOURCE_PATTERNS: RegExp[] = [
  /\b(?:href|src|srcset|action|formaction)\s*=\s*\{?\s*["'`]\/[^"'`\s]*/g,
  /\bsrcset\s*=\s*\{?\s*["'`][^"'`]*,\s*\/[^"'`\s]*/g,
  /\b(?:fetch|goto|sendBeacon|window\.open|location\.assign|location\.replace)\s*\(\s*["'`]\/[^"'`\s]*/g,
  /\bnew\s+(?:URL|Request)\s*\(\s*["'`]\/[^"'`\s]*/g,
  /\blocation(?:\.href)?\s*=\s*["'`]\/[^"'`\s]*/g
];

const CSS_URL = /\burl\(\s*["']?\/[^"')\s]*/g;

interface Violation {
  path: string;
  line: number;
  match: string;
}

function scan(path: string, contents: string, patterns: RegExp[]): Violation[] {
  const violations: Violation[] = [];
  contents.split('\n').forEach((line, i) => {
    for (const re of patterns) {
      for (const m of line.matchAll(re)) {
        violations.push({ path, line: i + 1, match: m[0] });
      }
    }
  });
  return violations;
}

function scanSource(path: string, contents: string): Violation[] {
  return scan(path, stripComments(path, contents), SOURCE_PATTERNS);
}

function scanBuiltCSS(path: string, contents: string): Violation[] {
  return scan(path, stripComments(path, contents), [CSS_URL]);
}

function main(): void {
  const sources = walk(SRC_ROOT).filter((rel) => SCANNED_EXTENSIONS.some((ext) => rel.endsWith(ext)));
  if (sources.length === 0) {
    fail(`root-absolute: ${SRC_ROOT} holds no markup or script; an empty iteration is not a pass.`);
  }
  const stylesheets = walk(BUILT_CSS_ROOT).filter((rel) => rel.endsWith('.css'));
  if (stylesheets.length === 0) {
    fail(`root-absolute: ${BUILT_CSS_ROOT} holds no stylesheet; run the build first.`);
  }

  const violations = [
    ...sources.flatMap((rel) =>
      scanSource(join(SRC_ROOT, rel), readFileSync(join(SRC_ROOT, rel), 'utf8'))
    ),
    ...stylesheets.flatMap((rel) =>
      scanBuiltCSS(join(BUILT_CSS_ROOT, rel), readFileSync(join(BUILT_CSS_ROOT, rel), 'utf8'))
    )
  ];
  if (violations.length > 0) {
    const lines = violations.map((v) => `${v.path}:${v.line}: ${v.match}`);
    fail(
      `root-absolute: ${violations.length} root-absolute URL(s); compose each from the runtime base:\n${lines.join('\n')}`
    );
  }
  process.stdout.write(
    `root-absolute: passed, read ${sources.length} source file(s) and ${stylesheets.length} built stylesheet(s).\n`
  );
}

if (isMainModule(import.meta.url)) {
  main();
}
