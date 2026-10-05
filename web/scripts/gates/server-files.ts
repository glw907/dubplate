/**
 * The server-file check. Server-only SvelteKit files and form actions
 * are forbidden (CLAUDE.md's web stack rule): the Go binary is the
 * only server, so a `+page.server.ts`, a `hooks.server.ts`, or an
 * `instrumentation.server.ts` (which kit enables by its presence alone)
 * would run nowhere. An absent or empty `src` fails rather than passing on an
 * empty iteration.
 *
 * The forbidden set is route and hooks entry points, which is a
 * different set from what SvelteKit treats as a server-only module (a
 * `server` filename segment, or any `server` directory outside
 * `src/routes`). Kit itself refuses to bundle a server-only module
 * into client code; this check refuses the entry points that would
 * need a server to run at all.
 */
import { existsSync, readFileSync, statSync } from 'node:fs';
import { basename, join } from 'node:path';
import type { Config } from '@sveltejs/kit/vite';
import { kitOptions } from '../../vite.config.ts';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';

/** True when a file's basename matches one of the forbidden server-file shapes. */
export function isForbiddenServerFile(path: string): boolean {
  const name = basename(path);
  return (
    /^\+page\.server\.(ts|js)$/.test(name) ||
    /^\+layout\.server\.(ts|js)$/.test(name) ||
    /^\+server\.(ts|js)$/.test(name) ||
    /^hooks\.server\.(ts|js)$/.test(name) ||
    /^instrumentation\.server\.(ts|js)$/.test(name)
  );
}

/** True for a `+page.ts`/`+page.svelte` whose source declares an `actions` export. */
export function declaresActionsExport(path: string, contents: string): boolean {
  if (!/^\+page\.(ts|js|svelte)$/.test(basename(path))) {
    return false;
  }
  return /export\s+(const|let)\s+actions\b/.test(contents);
}

/** `files.hooks.server` from the options `vite.config.ts` passes the kit plugin, or SvelteKit's own default `src/hooks.server`. */
export function hooksServerEntryBase(options: Config = kitOptions): string {
  return options.files?.hooks?.server ?? 'src/hooks.server';
}

/**
 * Whether SvelteKit would resolve `entryBase` to a real module: `<entryBase>.js`/`.ts`, or
 * `<entryBase>/index.js`/`.ts` when `entryBase` names a directory. Mirrors `@sveltejs/kit`'s own
 * `resolve_entry` (`kit/src/utils/filesystem.js`), which the plain basename scan above cannot
 * see: a `src/hooks.server/index.ts` directory, or a `files.hooks.server` entry under a
 * custom name, carries no `hooks.server.(ts|js)` basename at all.
 */
export function resolveHooksServerFile(entryBase: string): string | null {
  for (const ext of ['js', 'ts']) {
    const candidate = `${entryBase}.${ext}`;
    if (existsSync(candidate) && statSync(candidate).isFile()) return candidate;
  }
  if (existsSync(entryBase) && statSync(entryBase).isDirectory()) {
    for (const ext of ['js', 'ts']) {
      const candidate = join(entryBase, `index.${ext}`);
      if (existsSync(candidate) && statSync(candidate).isFile()) return candidate;
    }
  }
  return null;
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
  const entryBase = hooksServerEntryBase();
  const resolved = resolveHooksServerFile(entryBase);
  if (resolved) {
    violations.push(`${resolved}: forbidden server-only file (files.hooks.server entry)`);
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

if (isMainModule(import.meta.url)) {
  main();
}
