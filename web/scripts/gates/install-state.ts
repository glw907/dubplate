/**
 * The install-state comparison. `npm ls` alone misses an in-range
 * lockfile bump that has not been installed yet, because it
 * only checks whether the installed tree satisfies `package.json`'s
 * ranges. This compares `web/package-lock.json`'s resolved versions
 * against npm's own hidden lockfile
 * (`web/node_modules/.package-lock.json`), which is what was actually
 * installed. Node/npm presence, the Node major floor, and `npm ls`
 * itself are checks in the Makefile's `web` target, run before this
 * comparison so it never has to guess why `node_modules` is missing.
 */
import { readFileSync } from 'node:fs';
import { fail, isMainModule } from './lib.ts';

interface LockPackage {
  version?: string;
  optional?: boolean;
}
interface Lockfile {
  packages?: Record<string, LockPackage>;
}

function readLockfile(path: string): Lockfile {
  return JSON.parse(readFileSync(path, 'utf8')) as Lockfile;
}

export interface Drift {
  name: string;
  committed: string;
  installed: string;
}

/**
 * Every entry present in both lockfiles whose resolved version
 * differs, and every entry the committed lockfile marks non-optional
 * but the hidden lockfile omits (reported as installed `"absent"`).
 * The root `""` key is skipped; it carries the project's own metadata,
 * not a dependency.
 */
export function compare(committed: Lockfile, hidden: Lockfile): Drift[] {
  const drift: Drift[] = [];
  const committedPackages = committed.packages ?? {};
  const hiddenPackages = hidden.packages ?? {};
  for (const [name, entry] of Object.entries(committedPackages)) {
    if (name === '') continue;
    const installedEntry = hiddenPackages[name];
    if (installedEntry === undefined) {
      if (entry.optional) continue;
      drift.push({ name, committed: entry.version ?? '', installed: 'absent' });
      continue;
    }
    if (entry.version !== undefined && installedEntry.version !== entry.version) {
      drift.push({ name, committed: entry.version, installed: installedEntry.version ?? '' });
    }
  }
  return drift;
}

function main() {
  const committed = readLockfile('package-lock.json');
  const hidden = readLockfile('node_modules/.package-lock.json');
  const drift = compare(committed, hidden);
  if (drift.length > 0) {
    const first = drift[0];
    fail(
      `install-state: ${drift.length} lockfile entr${drift.length === 1 ? 'y' : 'ies'} out of` +
        ` step with node_modules, starting with ${first.name}` +
        ` (committed ${first.committed}, installed ${first.installed}); run \`npm ci --prefix web\`.`
    );
  }
  process.stdout.write('install-state: package-lock.json matches node_modules/.package-lock.json.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
