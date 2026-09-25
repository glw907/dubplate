import { readdirSync, statSync } from 'node:fs';
import { join } from 'node:path';

/**
 * Every file under `root`, recursively, as a path relative to `root`.
 * Directories are not yielded. An absent `root` yields no files rather
 * than throwing, so a caller decides for itself whether "no files" is
 * a pass or a fail (an absent `web/src` and an empty `web/src` are the
 * same failure, per the server-file check's states table).
 */
export function walk(root: string): string[] {
  const out: string[] = [];
  let entries: string[];
  try {
    entries = readdirSync(root);
  } catch {
    return out;
  }
  for (const name of entries) {
    const full = join(root, name);
    const st = statSync(full);
    if (st.isDirectory()) {
      out.push(...walk(full).map((p) => join(name, p)));
    } else if (st.isFile()) {
      out.push(name);
    }
  }
  return out;
}

/** Prints `message` to stderr and exits 1. Every gate script's one failure path. */
export function fail(message: string): never {
  process.stderr.write(`${message}\n`);
  process.exit(1);
}
