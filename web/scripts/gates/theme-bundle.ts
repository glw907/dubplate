/**
 * The built-bundle theme row: runs after `npm run build`, beside G3's
 * built-bundle half, gate-resident on every commit rather than a
 * one-time proof. Reads the built stylesheet's cascade and asserts, for
 * `dubplate-label`, `dubplate-label-night`, and `dubplate-system`, that
 * every stampable name's *effective* value (the last declaration in
 * source order across every rule in the same media context whose
 * selector list carries the compound `[data-theme=<name>]`, quoted or
 * not) matches the theme data files' value once both are parsed. It
 * also asserts the bundle's lone `prefers-color-scheme` block is the
 * system rule, that no `:root:not([data-theme])` selector exists, and
 * that `web/src` carries no `dark:` utility, which would style by the
 * device rather than by the stamp (Tailwind 4's default `dark` variant
 * is media-based: `tailwindcss@4.3.3` `dist/lib.js`, `r("dark",["@media
 * (prefers-color-scheme: dark)"])`).
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { isPlaceholderOnly } from './g3.ts';
import {
  type AtRuleNode,
  COMPOSED_NAMES,
  type DeclNode,
  INK_NAMES,
  type Node,
  normalizeMediaQuery,
  PALETTE_NAMES,
  pluginBlocksNamed,
  scanNodes,
  selectorHasDataTheme,
  selectorIsExactlyDataTheme,
  splitSelectorList,
  STRUCTURAL_NAMES,
  SYSTEM_NAME
} from './g2.ts';
import { fail, isMainModule, walk } from './lib.ts';

const BUILD_ROOT = 'build';
const ASSETS_DIR = 'build/_app/immutable/assets';
const THEMES_DIR = 'src/lib/themes';
const SRC_ROOT = 'src';

const DARK_QUERY = normalizeMediaQuery('(prefers-color-scheme: dark)');

interface FlatRule {
  selector: string;
  declarations: DeclNode[];
  media: 'none' | 'dark' | 'other';
}

/** Every `@media` node in the tree whose prelude mentions `prefers-color-scheme`, recursively. */
function collectPrefersColorSchemeBlocks(nodes: Node[]): AtRuleNode[] {
  const out: AtRuleNode[] = [];
  for (const node of nodes) {
    if (node.kind !== 'at-rule') continue;
    if (node.name === '@media' && /prefers-color-scheme/i.test(node.prelude)) {
      out.push(node);
    }
    out.push(...collectPrefersColorSchemeBlocks(node.children));
  }
  return out;
}

function flattenRules(nodes: Node[], media: FlatRule['media']): FlatRule[] {
  const out: FlatRule[] = [];
  for (const node of nodes) {
    if (node.kind === 'rule') {
      out.push({ selector: node.selector, declarations: node.declarations, media });
    } else if (node.kind === 'at-rule') {
      const nextMedia: FlatRule['media'] =
        node.name === '@media' ? (normalizeMediaQuery(node.prelude) === DARK_QUERY ? 'dark' : 'other') : media;
      out.push(...flattenRules(node.children, nextMedia));
    }
  }
  return out;
}

/**
 * Every rule whose selector list carries `[data-theme=<name>]`, in source
 * order, later writes winning. The `none` context reads only unconditional
 * rules; the `dark` context reads both unconditional and dark-media rules,
 * since the cascade applies both once the media query matches, and an
 * unconditional rule that lands later in the bundle still wins.
 */
function effectiveValues(rules: FlatRule[], name: string, context: 'none' | 'dark'): Map<string, string> {
  const map = new Map<string, string>();
  for (const rule of rules) {
    const applies = context === 'none' ? rule.media === 'none' : rule.media === 'none' || rule.media === 'dark';
    if (!applies) continue;
    if (!selectorHasDataTheme(rule.selector, name)) continue;
    for (const decl of rule.declarations) map.set(decl.name, decl.value);
  }
  return map;
}

/** Compares two CSS values as parsed data, never as raw bytes: hex case and a minifier-dropped leading zero both normalize away. */
function normalizeValue(value: string): string {
  const hex = /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.exec(value.trim());
  if (hex) {
    const digits = hex[1].length === 3 ? hex[1].split('').map((c) => c + c).join('') : hex[1];
    return `#${digits.toLowerCase()}`;
  }
  const joined = value
    .split(',')
    .map((p) => p.trim())
    .join(', ')
    .toLowerCase();
  return joined.replace(/(^|[\s(,])0\.(\d)/g, '$1.$2');
}

function readReferenceDeclarations(): {
  match: Map<string, Map<string, string>>;
  palette: Map<string, Map<string, string>>;
  inks: Map<string, Map<string, string>>;
} {
  const match = new Map<string, Map<string, string>>();
  const dubplateText = readFileSync(`${THEMES_DIR}/dubplate.css`, 'utf8');
  for (const [name, block] of pluginBlocksNamed(dubplateText)) {
    match.set(name, new Map(block.declarations.map((d) => [d.name, d.value])));
  }
  const palette = new Map<string, Map<string, string>>();
  for (const [file, composed] of [
    ['label.css', 'dubplate-label'],
    ['label-night.css', 'dubplate-label-night']
  ] as const) {
    const text = readFileSync(`${THEMES_DIR}/${file}`, 'utf8');
    const block = pluginBlocksNamed(text).get(composed);
    if (block) palette.set(composed, new Map(block.declarations.map((d) => [d.name, d.value])));
  }
  const inks = new Map<string, Map<string, string>>();
  const inksText = readFileSync(`${THEMES_DIR}/inks.css`, 'utf8');
  for (const node of scanNodes(inksText, 0, inksText.length)) {
    if (node.kind !== 'at-rule' || node.name !== '@layer' || node.prelude !== 'base') continue;
    for (const child of node.children) {
      if (child.kind !== 'rule') continue;
      for (const composed of COMPOSED_NAMES) {
        if (selectorHasDataTheme(child.selector, composed)) {
          inks.set(composed, new Map(child.declarations.map((d) => [d.name, d.value])));
        }
      }
    }
  }
  return { match, palette, inks };
}

function checkStampedValues(
  effective: Map<string, string>,
  reference: Map<string, string>,
  names: readonly string[],
  where: string,
  violations: string[]
): void {
  for (const name of names) {
    const got = effective.get(name);
    const want = reference.get(name);
    if (got === undefined) {
      violations.push(`theme-bundle: ${where}: ${name} is not effective`);
      continue;
    }
    if (want === undefined) continue; // a reference gap is g2's finding, not this row's
    if (normalizeValue(got) !== normalizeValue(want)) {
      violations.push(`theme-bundle: ${where}: ${name} is "${got}", the data file says "${want}"`);
    }
  }
}

/** Every theme-stampable name G2 and G9 already track, the inventory a rogue rule could set. */
const THEME_TOKEN_NAMES = new Set<string>([...STRUCTURAL_NAMES, ...PALETTE_NAMES, ...INK_NAMES]);
/** A `[data-theme=dubplate-*]` compound, quoted or not: the only selector shape a theme-token rule may confine itself to. */
const DUBPLATE_THEME_COMPOUND = /\[data-theme=(?:"dubplate-[\w-]+"|dubplate-[\w-]+)\]/;
/** DaisyUI's stock `light`/`dark` theme compound, quoted or not. */
const STOCK_DATA_THEME = /\[data-theme=(?:"(?:light|dark)"|light|dark)\]/;
/** DaisyUI's `theme-controller` checkbox machinery for its own stock `light`/`dark` themes. */
const STOCK_THEME_CONTROLLER = /input\.theme-controller\[value=(?:"?light"?)\]|\[value=(?:"?dark"?)\]/;

/**
 * True for a bare `:root` and/or `html` selector part: the shape a
 * hand-authored global override takes (the review's own probe,
 * `:root{--color-primary:#0000ff}`). `:host` is excluded on purpose:
 * Tailwind's own `:root,:host{--font-sans:...}` preflight reset uses
 * this shape legitimately (verified in the built bundle, `@layer
 * theme`, ahead of `@layer base` in Tailwind's own layer order, so it
 * can never win the cascade over a stamped theme regardless of source
 * order) and is not the hazard this check targets.
 */
function isBareRootPart(part: string): boolean {
  return part === ':root' || part === 'html';
}

/**
 * True for a selector part that confines a theme-token rule to a real
 * theme: a `[data-theme=dubplate-*]` compound, DaisyUI's own stock
 * `[data-theme=light]`/`[data-theme=dark]` compound, or DaisyUI's
 * `theme-controller` checkbox machinery for those stock themes. `:host`
 * rides along with `:root` in Tailwind's preflight reset, never alone.
 */
function isConfinedPart(part: string): boolean {
  return (
    DUBPLATE_THEME_COMPOUND.test(part) ||
    part === ':where(:root)' ||
    part === ':host' ||
    STOCK_DATA_THEME.test(part) ||
    STOCK_THEME_CONTROLLER.test(part)
  );
}

/**
 * Fails a rule that sets a theme-stampable name unless every
 * comma-separated part of its selector either confines it to a real
 * theme (`isConfinedPart`) or is a bare `:root`/`html` riding alongside
 * at least one confined or `:host` part (Tailwind's preflight shape).
 * A rule whose every part is bare `:root`/`html` and nothing else (no
 * `:host`, no confined part) is the classic global-override hazard and
 * always fails, regardless of `@layer`. A rule that does carry a
 * confining part but sits outside `@layer base` also fails: it competes
 * with our own `@layer base` theme rules by source order instead of by
 * a settled layer order.
 */
function checkNameConfinement(nodes: Node[], insideLayerBase: boolean, violations: string[]): void {
  for (const node of nodes) {
    if (node.kind === 'rule') {
      const names = node.declarations.map((d) => d.name).filter((n) => THEME_TOKEN_NAMES.has(n));
      if (names.length === 0) continue;
      const parts = splitSelectorList(node.selector);
      const confinedParts = parts.filter((p) => isConfinedPart(p));
      const looseParts = parts.filter((p) => !isConfinedPart(p) && !isBareRootPart(p));
      if (confinedParts.length === 0 && parts.every((p) => isBareRootPart(p))) {
        violations.push(
          `theme-bundle: rule "${node.selector}" sets ${names.join(', ')} with no [data-theme=dubplate-*] compound at all`
        );
      } else if (looseParts.length > 0) {
        violations.push(
          `theme-bundle: rule "${node.selector}" sets ${names.join(', ')} through unconfined selector part(s) ${looseParts.join(', ')}`
        );
      } else if (confinedParts.some((p) => DUBPLATE_THEME_COMPOUND.test(p)) && !insideLayerBase) {
        violations.push(`theme-bundle: rule "${node.selector}" sets ${names.join(', ')} outside @layer base`);
      }
    } else if (node.kind === 'at-rule') {
      const nextInside = insideLayerBase || (node.name === '@layer' && node.prelude === 'base');
      checkNameConfinement(node.children, nextInside, violations);
    }
  }
}

export function run(): string[] {
  const violations: string[] = [];

  if (isPlaceholderOnly(BUILD_ROOT)) {
    fail('theme-bundle: build holds only placeholder.txt; run `npm run build` first.');
  }
  const cssFiles = walk(ASSETS_DIR).filter((f) => f.endsWith('.css'));
  if (cssFiles.length === 0) {
    fail(`theme-bundle: no built stylesheet under ${ASSETS_DIR}.`);
  }
  const bundleText = cssFiles.map((f) => readFileSync(join(ASSETS_DIR, f), 'utf8')).join('\n');

  const rules = flattenRules(scanNodes(bundleText, 0, bundleText.length), 'none');
  const reference = readReferenceDeclarations();

  const paletteNamesExceptScheme = PALETTE_NAMES.filter((n) => n !== 'color-scheme');
  for (const composed of COMPOSED_NAMES) {
    const effective = effectiveValues(rules, composed, 'none');
    const matchRef = reference.match.get(composed) ?? new Map();
    const paletteRef = reference.palette.get(composed) ?? new Map();
    const inkRef = reference.inks.get(composed) ?? new Map();
    const combinedRef = new Map([...matchRef, ...paletteRef, ...inkRef]);
    checkStampedValues(
      effective,
      combinedRef,
      [...STRUCTURAL_NAMES, ...paletteNamesExceptScheme, ...INK_NAMES],
      composed,
      violations
    );

    // Checked by name and by "never normal" both: a theme whose
    // effective color-scheme falls back to the match axis's default
    // still passes name-completeness while missing its own value.
    const wantScheme = composed === 'dubplate-label' ? 'light' : 'dark';
    const gotScheme = effective.get('color-scheme');
    if (gotScheme !== wantScheme) {
      violations.push(`theme-bundle: ${composed}: color-scheme is "${gotScheme}", want "${wantScheme}"`);
    }
  }

  const lightEffective = effectiveValues(rules, SYSTEM_NAME, 'none');
  const darkEffective = effectiveValues(rules, SYSTEM_NAME, 'dark');
  const branches: [string, Map<string, string>, string][] = [
    ['light', lightEffective, 'dubplate-label'],
    ['dark', darkEffective, 'dubplate-label-night']
  ];
  for (const [label, effective, composed] of branches) {
    const matchRef = reference.match.get(composed) ?? new Map();
    const paletteRef = reference.palette.get(composed) ?? new Map();
    const inkRef = reference.inks.get(composed) ?? new Map();
    const combinedRef = new Map([...matchRef, ...paletteRef, ...inkRef]);
    checkStampedValues(
      effective,
      combinedRef,
      [...STRUCTURAL_NAMES, ...PALETTE_NAMES, ...INK_NAMES],
      `dubplate-system (${label})`,
      violations
    );
  }

  const mediaBlocks = collectPrefersColorSchemeBlocks(scanNodes(bundleText, 0, bundleText.length));
  if (mediaBlocks.length !== 1) {
    violations.push(
      `theme-bundle: the bundle carries ${mediaBlocks.length} prefers-color-scheme block(s), want exactly 1`
    );
  } else {
    const [block] = mediaBlocks;
    if (normalizeMediaQuery(block.prelude) !== DARK_QUERY) {
      violations.push(
        `theme-bundle: the prefers-color-scheme block's query is "${block.prelude}", want "(prefers-color-scheme: dark)"`
      );
    }
    for (const child of block.children) {
      if (child.kind !== 'rule') {
        violations.push('theme-bundle: the prefers-color-scheme block holds a non-rule node');
        continue;
      }
      if (!selectorIsExactlyDataTheme(child.selector, SYSTEM_NAME)) {
        violations.push(
          `theme-bundle: the prefers-color-scheme block holds an unexpected selector "${child.selector}"`
        );
      }
    }
  }
  if (/:root:not\(\[data-theme\]\)/.test(bundleText)) {
    violations.push('theme-bundle: the bundle carries a :root:not([data-theme]) guard');
  }

  checkNameConfinement(scanNodes(bundleText, 0, bundleText.length), false, violations);

  const srcFiles = walk(SRC_ROOT);
  for (const rel of srcFiles) {
    if (!/\.(svelte|ts|js|html|css)$/.test(rel)) continue;
    const contents = readFileSync(join(SRC_ROOT, rel), 'utf8');
    contents.split('\n').forEach((line, i) => {
      if (/\bdark:/.test(line)) {
        violations.push(`theme-bundle: ${rel}:${i + 1}: a stray dark: utility styles by the device, not the stamp`);
      }
      if (/\btheme-controller\b/.test(line)) {
        violations.push(
          `theme-bundle: ${rel}:${i + 1}: a theme-controller reference bypasses the stamp DL R4.6 bans`
        );
      }
    });
  }

  return violations;
}

function main(): void {
  const violations = run();
  if (violations.length > 0) {
    fail(`theme-bundle: ${violations.length} finding(s):\n${violations.join('\n')}`);
  }
  process.stdout.write('theme-bundle: passed, every stampable value is effective and at parity.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
