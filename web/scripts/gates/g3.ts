/**
 * G3: no external origin. Every asset ships embedded in the binary, so
 * a `url()`, `@import`, `<link>`, or `@font-face src` naming an
 * outside host would 404 on a box with no outbound fetch for it, or
 * silently depend on one. Reads `web/src` and the built bundle under
 * `web/build`.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';
const BUILD_ROOT = 'build';
const SCANNED_EXTENSIONS = ['.svelte', '.css', '.html', '.js', '.ts'];

// A scheme or a protocol-relative `//` sitting inside a `url()`,
// `@import`, a `<link href>` directive, an `src`/`srcset`/`poster`/`data`
// attribute, or an `import()`/`fetch()` call, never a bare scheme
// anywhere in a file (a comment or a string elsewhere is not a load).
const EXTERNAL_ORIGIN =
  /(url\(\s*['"]?|@import\s+['"]?|<link\b[^>]*\bhref\s*=\s*['"]|\b(?:src|srcset|poster|data)\s*=\s*['"]|\b(?:import|fetch)\(\s*['"])(https?:)?\/\//gi;

export interface Violation {
  path: string;
  line: number;
}

export function scan(root: string, extensions: string[]): Violation[] {
  const violations: Violation[] = [];
  for (const rel of walk(root)) {
    if (!extensions.some((ext) => rel.endsWith(ext))) continue;
    const contents = readFileSync(join(root, rel), 'utf8');
    contents.split('\n').forEach((line, i) => {
      EXTERNAL_ORIGIN.lastIndex = 0;
      if (EXTERNAL_ORIGIN.test(line)) {
        violations.push({ path: rel, line: i + 1 });
      }
    });
  }
  return violations;
}

/** True when `root` holds only the placeholder file, the never-built tree's shape. */
export function isPlaceholderOnly(root: string): boolean {
  const files = walk(root);
  return files.length === 0 || (files.length === 1 && files[0] === 'placeholder.txt');
}

function main() {
  const srcFiles = walk(SRC_ROOT);
  if (srcFiles.length === 0) {
    fail(`g3: ${SRC_ROOT} is absent or empty; an empty iteration is not a pass.`);
  }
  const srcViolations = scan(SRC_ROOT, SCANNED_EXTENSIONS);
  if (srcViolations.length > 0) {
    const lines = srcViolations.map((v) => `${v.path}:${v.line}`);
    fail(`g3: ${srcViolations.length} external origin(s) in ${SRC_ROOT}:\n${lines.join('\n')}`);
  }

  if (isPlaceholderOnly(BUILD_ROOT)) {
    fail(`g3: ${BUILD_ROOT} holds only placeholder.txt; run \`npm run build\` first.`);
  }
  const buildViolations = scan(BUILD_ROOT, ['.css', '.html', '.js']);
  if (buildViolations.length > 0) {
    const lines = buildViolations.map((v) => `${v.path}:${v.line}`);
    fail(`g3: ${buildViolations.length} external origin(s) in the built bundle:\n${lines.join('\n')}`);
  }

  process.stdout.write('g3: passed, no external origin found.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
