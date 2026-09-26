/**
 * W5.8's unlayered-count ratchet. `lib/styles/components.css` (its own
 * header) explains why an unlayered declaration outranks every layer:
 * that reach is exactly what makes an unlayered rule dangerous to add
 * casually, so this gate holds the set to a pinned list rather than
 * merely counting it. It scans every authored `web/src/**\/*.css` file,
 * finds each style rule that sits outside every `@layer` block, and
 * fails unless that set is exactly the pinned list below.
 *
 * A `@font-face` rule is an at-rule, not a style rule (CSS has no
 * selector to key it by), so it never enters the scan. A Tailwind
 * `@utility` block is also an at-rule: its body holds no nested
 * `selector { ... }`, so `scanNodes` finds no rule node inside it
 * either. Svelte `<style>` blocks are scoped by the compiler (W5.7)
 * and are outside this script's subject; `unlayered-font-clobber`
 * (G9) governs those.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { type Node, scanNodes } from './g2.ts';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';

/** One style rule found outside every `@layer` block. */
export interface UnlayeredRule {
  file: string;
  /** The at-rule directly enclosing the rule (its name and prelude joined), or `null` at the file's own top level. */
  enclosing: string | null;
  selector: string;
}

/** `rule`'s identity for the pinned-list comparison: its line number never enters it, so a harmless reflow cannot break the ratchet. */
function key(rule: UnlayeredRule): string {
  return `${rule.file}\u0000${rule.enclosing ?? ''}\u0000${rule.selector}`;
}

/**
 * The one rule this ratchet already admits: W9.3's reduced-motion kill
 * switch (`lib/styles/daisyui-overrides.css`), which must outrank every
 * layer including an operator's own tier 3 CSS.
 */
export const PINNED: UnlayeredRule[] = [
  { file: 'lib/styles/daisyui-overrides.css', enclosing: '@media (prefers-reduced-motion: reduce)', selector: '*,\n  ::before,\n  ::after' }
];

function lineOf(text: string, offset: number): number {
  let line = 1;
  for (let i = 0; i < offset && i < text.length; i++) {
    if (text[i] === '\n') line++;
  }
  return line;
}

/** Every style rule in `nodes` that sits outside every ancestor `@layer` block, recursively. */
function collect(
  nodes: Node[],
  file: string,
  text: string,
  insideLayer: boolean,
  enclosing: string | null,
  out: (UnlayeredRule & { line: number })[]
): void {
  for (const node of nodes) {
    if (node.kind === 'rule') {
      if (!insideLayer) {
        out.push({ file, enclosing, selector: node.selector, line: lineOf(text, node.start) });
      }
      continue;
    }
    if (node.kind === 'plugin') continue;
    // node.kind === 'at-rule'
    const nextInsideLayer = insideLayer || node.name === '@layer';
    const label = `${node.name}${node.prelude ? ` ${node.prelude}` : ''}`;
    collect(node.children, file, text, nextInsideLayer, insideLayer ? enclosing : label, out);
  }
}

/** Every unlayered style rule under `root`'s `.css` files, in file order. */
export function run(root: string): (UnlayeredRule & { line: number })[] {
  const out: (UnlayeredRule & { line: number })[] = [];
  for (const rel of walk(root).filter((p) => p.endsWith('.css')).sort()) {
    const text = readFileSync(join(root, rel), 'utf8');
    collect(scanNodes(text, 0, text.length), rel, text, false, null, out);
  }
  return out;
}

function main(): void {
  const files = walk(SRC_ROOT).filter((p) => p.endsWith('.css'));
  if (files.length === 0) {
    fail(`unlayered-count: no .css file found under ${SRC_ROOT}; an empty scan is not a pass.`);
  }
  const found = run(SRC_ROOT);
  const foundKeys = new Map(found.map((r) => [key(r), r]));
  const pinnedKeys = new Set(PINNED.map(key));

  const unpinned = found.filter((r) => !pinnedKeys.has(key(r)));
  if (unpinned.length > 0) {
    const lines = unpinned.map((r) => `${r.file}:${r.line}: unpinned unlayered rule "${r.selector}"${r.enclosing ? ` under ${r.enclosing}` : ''}`);
    fail(`unlayered-count: ${unpinned.length} unpinned unlayered rule(s):\n${lines.join('\n')}`);
  }
  const missing = PINNED.filter((p) => !foundKeys.has(key(p)));
  if (missing.length > 0) {
    const lines = missing.map((p) => `${p.file}: pinned rule "${p.selector}"${p.enclosing ? ` under ${p.enclosing}` : ''} no longer found; the ratchet is stale`);
    fail(`unlayered-count: ${missing.length} pinned entry(ies) no longer found:\n${lines.join('\n')}`);
  }
  process.stdout.write(`unlayered-count: passed, ${found.length} unlayered rule(s), matching the pinned list of ${PINNED.length}.\n`);
}

if (isMainModule(import.meta.url)) {
  main();
}
