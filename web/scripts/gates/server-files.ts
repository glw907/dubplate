/**
 * The server-file check. Server-only SvelteKit files and form actions
 * are forbidden (CLAUDE.md's web stack rule): the Go binary is the
 * only server, so a `+page.server.ts` or a `hooks.server.ts` would run
 * nowhere. An absent or empty `src` fails rather than passing on an
 * empty iteration.
 */
import { readFileSync } from 'node:fs';
import { basename, join } from 'node:path';
import { fail, walk } from './lib.ts';

const SRC_ROOT = 'src';

/** True when a file's basename matches one of the forbidden server-file shapes. */
export function isForbiddenServerFile(path: string): boolean {
  const name = basename(path);
  return (
    /^\+page\.server\.(ts|js)$/.test(name) ||
    /^\+layout\.server\.(ts|js)$/.test(name) ||
    /^\+server\.(ts|js)$/.test(name) ||
    /^hooks\.server\.(ts|js)$/.test(name)
  );
}

/** True for a `+page.ts`/`+page.svelte` whose source declares an `actions` export. */
export function declaresActionsExport(path: string, contents: string): boolean {
  if (!/^\+page\.(ts|js|svelte)$/.test(basename(path))) {
    return false;
  }
  return /export\s+(const|let)\s+actions\b/.test(contents);
}

export function run(root: string): { readCount: number; violations: string[] } {
  const files = walk(root);
  const violations: string[] = [];
  for (const rel of files) {
    if (isForbiddenServerFile(rel)) {
      violations.push(`${rel}: forbidden server-only file`);
      continue;
    }
    const contents = readFileSync(join(root, rel), 'utf8');
    if (declaresActionsExport(rel, contents)) {
      violations.push(`${rel}: exports \`actions\`, and form actions are forbidden`);
    }
  }
  return { readCount: files.length, violations };
}

function main() {
  const { readCount, violations } = run(SRC_ROOT);
  if (readCount === 0) {
    fail(`server-files: ${SRC_ROOT} is absent or empty; an empty iteration is not a pass.`);
  }
  if (violations.length > 0) {
    fail(`server-files: ${violations.length} forbidden file(s):\n${violations.join('\n')}`);
  }
  process.stdout.write(`server-files: passed, read ${readCount} file(s) under ${SRC_ROOT}.\n`);
}

if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}
