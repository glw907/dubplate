/**
 * G2: variable completeness, extended for the inks and the
 * `dubplate-system` copy. Parses the five theme data files under
 * `src/lib/themes/` and asserts:
 *
 * - `dubplate.css` carries two `@plugin "daisyui/theme"` blocks, named
 *   `dubplate-label` and `dubplate-label-night`, each declaring exactly
 *   the eleven structural names at identical values, no `default` or
 *   `prefersdark` key, a `--size-field` clearing the 44px touch-target
 *   floor, and each font value ending in a CSS generic family.
 * - `label.css` and `label-night.css` each carry one block declaring
 *   exactly the 21 palette names, no `default` or `prefersdark` key.
 * - `inks.css` carries one ink block per composed theme, each declaring
 *   exactly the ink set (`--ink-primary`, `--ink-warning`).
 * - `system.css` carries the unconditional and the
 *   `prefers-color-scheme: dark` `[data-theme="dubplate-system"]`
 *   branches, each repeating its composed theme's 32 names plus the
 *   ink set at parity, byte for byte after parsing.
 *
 * This module also exports the low-level CSS scanner `theme-bundle.ts`
 * and `g9.ts` reuse: a generic `@rule`/rule/`@plugin` tree walk that
 * treats a comment as inert text rather than stripping it, so a
 * `contrast-exception` marker's position can be matched back to the
 * declaration it sits above.
 */
import { readFileSync } from 'node:fs';
import { fail, isMainModule } from './lib.ts';

const THEMES_DIR = 'src/lib/themes';

export const STRUCTURAL_NAMES = [
  '--radius-selector',
  '--radius-field',
  '--radius-box',
  '--size-selector',
  '--size-field',
  '--border',
  '--depth',
  '--noise',
  '--font-sans',
  '--font-serif',
  '--font-mono'
];

export const PALETTE_NAMES = [
  'color-scheme',
  '--color-base-100',
  '--color-base-200',
  '--color-base-300',
  '--color-base-content',
  '--color-primary',
  '--color-primary-content',
  '--color-secondary',
  '--color-secondary-content',
  '--color-accent',
  '--color-accent-content',
  '--color-neutral',
  '--color-neutral-content',
  '--color-info',
  '--color-info-content',
  '--color-success',
  '--color-success-content',
  '--color-warning',
  '--color-warning-content',
  '--color-error',
  '--color-error-content'
];

/** The two status inks G9 also reads: the inventory's one home. */
export const INK_NAMES = ['--ink-primary', '--ink-warning'];

export const COMPOSED_NAMES = ['dubplate-label', 'dubplate-label-night'] as const;
export const SYSTEM_NAME = 'dubplate-system';

/** A theme plugin's own authoring metadata: never a completeness-checked name. */
const PLUGIN_METADATA_KEYS = new Set(['name', 'default', 'prefersdark', 'root']);

/** DL's fixed exception-marker form: `/* contrast-exception: <var> <ratio>:1 <reason> *\/`. */
const MARKER_PATTERN = /^\/\*\s*contrast-exception:\s*(--[\w-]+)\s+(\d+(?:\.\d+)?):1\s+(.+?)\s*\*\/$/;

export interface ExceptionMarker {
  variable: string;
  ratio: string;
  reason: string;
}

interface CommentSpan {
  start: number;
  end: number;
  marker: ExceptionMarker | null;
}

export interface DeclNode {
  name: string;
  value: string;
  start: number;
  marker: ExceptionMarker | null;
}

export interface RuleNode {
  kind: 'rule';
  selector: string;
  declarations: DeclNode[];
  start: number;
}

export interface PluginNode {
  kind: 'plugin';
  prelude: string;
  declarations: DeclNode[];
  start: number;
}

export interface AtRuleNode {
  kind: 'at-rule';
  name: string;
  prelude: string;
  children: Node[];
  start: number;
}

export type Node = RuleNode | PluginNode | AtRuleNode;

function findComments(text: string): CommentSpan[] {
  const spans: CommentSpan[] = [];
  const re = /\/\*[\s\S]*?\*\//g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(text))) {
    const parsed = MARKER_PATTERN.exec(m[0]);
    spans.push({
      start: m.index,
      end: m.index + m[0].length,
      marker: parsed ? { variable: parsed[1], ratio: parsed[2], reason: parsed[3] } : null
    });
  }
  return spans;
}

function blankSpans(text: string, spans: CommentSpan[]): string {
  const chars = text.split('');
  for (const span of spans) {
    for (let i = span.start; i < span.end; i++) chars[i] = ' ';
  }
  return chars.join('');
}

/** Every top-level `;`-terminated segment of `blanked`, quotes and `()` nesting respected. */
function splitTopLevel(blanked: string, sep: string): { start: number; end: number }[] {
  const out: { start: number; end: number }[] = [];
  let depth = 0;
  let quote: string | null = null;
  let segStart = 0;
  for (let i = 0; i < blanked.length; i++) {
    const ch = blanked[i];
    if (quote) {
      if (ch === quote && blanked[i - 1] !== '\\') quote = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      quote = ch;
      continue;
    }
    if (ch === '(') depth++;
    else if (ch === ')') depth--;
    else if (ch === sep && depth === 0) {
      out.push({ start: segStart, end: i });
      segStart = i + 1;
    }
  }
  out.push({ start: segStart, end: blanked.length });
  return out;
}

function unquote(value: string): string {
  if (value.length >= 2) {
    const first = value[0];
    const last = value[value.length - 1];
    if ((first === '"' || first === "'") && first === last) return value.slice(1, -1);
  }
  return value;
}

/** Parses a `{ ... }` body's declarations, `prop: value;`, pairing each with a directly-above `contrast-exception` marker. */
export function parseDeclarationBody(text: string, start: number, end: number): DeclNode[] {
  const body = text.slice(start, end);
  const comments = findComments(body);
  const blanked = blankSpans(body, comments);
  const declarations: DeclNode[] = [];
  for (const seg of splitTopLevel(blanked, ';')) {
    const raw = blanked.slice(seg.start, seg.end);
    const trimmed = raw.trim();
    if (!trimmed) continue;
    const colon = trimmed.indexOf(':');
    if (colon === -1) continue;
    const name = trimmed.slice(0, colon).trim();
    const value = unquote(trimmed.slice(colon + 1).trim());
    const marker =
      comments.find((c) => c.marker && c.start >= seg.start && c.end <= seg.end)?.marker ?? null;
    declarations.push({ name, value, start: start + seg.start, marker });
  }
  return declarations;
}

/** The index of the first top-level `{`, quotes and `[]`/`()` nesting respected, or -1. */
function findBraceOpen(text: string, from: number, limit: number): number {
  let bracket = 0;
  let paren = 0;
  let quote: string | null = null;
  for (let i = from; i < limit; i++) {
    const ch = text[i];
    if (quote) {
      if (ch === quote && text[i - 1] !== '\\') quote = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      quote = ch;
      continue;
    }
    if (ch === '[') bracket++;
    else if (ch === ']') bracket--;
    else if (ch === '(') paren++;
    else if (ch === ')') paren--;
    else if (ch === '{' && bracket === 0 && paren === 0) return i;
    else if (ch === ';' && bracket === 0 && paren === 0) return -i - 2; // encode a bare `;` boundary
  }
  return -1;
}

function matchBrace(text: string, openIndex: number): number {
  let depth = 0;
  let quote: string | null = null;
  for (let i = openIndex; i < text.length; i++) {
    const ch = text[i];
    if (quote) {
      if (ch === quote && text[i - 1] !== '\\') quote = null;
      continue;
    }
    if (ch === '"' || ch === "'") {
      quote = ch;
      continue;
    }
    if (ch === '/' && text[i + 1] === '*') {
      const close = text.indexOf('*/', i + 2);
      i = close === -1 ? text.length : close + 1;
      continue;
    }
    if (ch === '{') depth++;
    else if (ch === '}') {
      depth--;
      if (depth === 0) return i;
    }
  }
  return -1;
}

/**
 * A generic recursive scan of rules, `@media`/`@layer` blocks, and
 * `@plugin "daisyui/theme"` option blocks, in source order. A bare
 * `@name ...;` statement (Tailwind's own `@layer base,utilities;`
 * ordering line in a built bundle) is consumed and produces no node.
 */
export function scanNodes(text: string, start: number, end: number): Node[] {
  const nodes: Node[] = [];
  let i = start;
  while (i < end) {
    while (i < end) {
      const ch = text[i];
      if (/\s/.test(ch)) {
        i++;
        continue;
      }
      if (ch === '/' && text[i + 1] === '*') {
        const close = text.indexOf('*/', i + 2);
        i = close === -1 ? end : close + 2;
        continue;
      }
      break;
    }
    if (i >= end) break;
    const boundary = findBraceOpen(text, i, end);
    if (boundary === -1) break;
    if (boundary < -1) {
      // A bare `@rule ...;` with no body: skip past the semicolon.
      i = -boundary - 1;
      continue;
    }
    const prelude = text.slice(i, boundary).trim();
    const closeBrace = matchBrace(text, boundary);
    if (closeBrace === -1 || closeBrace >= end) break;
    if (prelude.startsWith('@')) {
      const spaceIdx = prelude.search(/\s/);
      const name = spaceIdx === -1 ? prelude : prelude.slice(0, spaceIdx);
      const atPrelude = spaceIdx === -1 ? '' : prelude.slice(spaceIdx + 1).trim();
      if (name === '@plugin') {
        nodes.push({
          kind: 'plugin',
          prelude: unquote(atPrelude),
          declarations: parseDeclarationBody(text, boundary + 1, closeBrace),
          start: i
        });
      } else {
        nodes.push({
          kind: 'at-rule',
          name,
          prelude: atPrelude,
          children: scanNodes(text, boundary + 1, closeBrace),
          start: i
        });
      }
    } else {
      nodes.push({
        kind: 'rule',
        selector: prelude,
        declarations: parseDeclarationBody(text, boundary + 1, closeBrace),
        start: i
      });
    }
    i = closeBrace + 1;
  }
  return nodes;
}

/** Whether `selector`'s list carries the compound `[data-theme=<name>]`, quoted or not. */
export function selectorHasDataTheme(selector: string, name: string): boolean {
  const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const re = new RegExp(`\\[data-theme=(?:"${escaped}"|${escaped})\\](?![\\w-])`);
  return re.test(selector);
}

/** Splits a selector list at top-level commas, ignoring commas nested inside `[]` or `()`. */
export function splitSelectorList(selector: string): string[] {
  const parts: string[] = [];
  let depth = 0;
  let current = '';
  for (const ch of selector) {
    if (ch === '[' || ch === '(') depth++;
    else if (ch === ']' || ch === ')') depth--;
    if (ch === ',' && depth === 0) {
      parts.push(current.trim());
      current = '';
    } else {
      current += ch;
    }
  }
  if (current.trim()) parts.push(current.trim());
  return parts;
}

/** Whether `selector`'s list holds nothing but the compound `[data-theme=<name>]`, quoted or not: every comma-separated entry, not just one. */
export function selectorIsExactlyDataTheme(selector: string, name: string): boolean {
  const escaped = name.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const re = new RegExp(`^\\[data-theme=(?:"${escaped}"|${escaped})\\]$`);
  const parts = splitSelectorList(selector);
  return parts.length > 0 && parts.every((part) => re.test(part));
}

/** Normalizes an `@media` prelude for exact comparison (whitespace only; casing already matches ours). */
export function normalizeMediaQuery(prelude: string): string {
  return prelude.replace(/\s+/g, '').toLowerCase();
}

const DARK_QUERY = normalizeMediaQuery('(prefers-color-scheme: dark)');

interface ThemeFiles {
  dubplate: string;
  label: string;
  labelNight: string;
  inks: string;
  system: string;
}

function readThemeFiles(violations: string[]): ThemeFiles | null {
  const names: [keyof ThemeFiles, string][] = [
    ['dubplate', 'dubplate.css'],
    ['label', 'label.css'],
    ['labelNight', 'label-night.css'],
    ['inks', 'inks.css'],
    ['system', 'system.css']
  ];
  const out: Partial<ThemeFiles> = {};
  let missing = false;
  for (const [key, file] of names) {
    try {
      out[key] = readFileSync(`${THEMES_DIR}/${file}`, 'utf8');
    } catch {
      violations.push(`g2: ${file}: no theme file`);
      missing = true;
    }
  }
  return missing ? null : (out as ThemeFiles);
}

export function pluginBlocksNamed(text: string): Map<string, PluginNode> {
  const byName = new Map<string, PluginNode>();
  for (const node of scanNodes(text, 0, text.length)) {
    if (node.kind !== 'plugin' || node.prelude !== 'daisyui/theme') continue;
    const nameDecl = node.declarations.find((d) => d.name === 'name');
    if (!nameDecl) continue;
    byName.set(nameDecl.value, node);
  }
  return byName;
}

/** Names a top-level node for a violation message: its selector, at-rule name, or `@plugin` prelude. */
function describeTopLevelNode(node: Node): string {
  if (node.kind === 'rule') return `rule "${node.selector}"`;
  if (node.kind === 'plugin') return `@plugin "${node.prelude}"`;
  return node.prelude ? `at-rule "${node.name} ${node.prelude}"` : `at-rule "${node.name}"`;
}

/**
 * `inks.css` and `system.css` must each hold exactly one top-level node,
 * an `@layer base` at-rule: nothing else, and not split across two
 * `@layer base` blocks (concatenating two is valid CSS but defeats the
 * "one authored home" shape this checker relies on to read children).
 * Pushes a violation naming `file` and every top-level node found when
 * that shape does not hold, and returns the single block's children
 * (empty when the shape is wrong, so callers see no phantom names).
 */
export function singleLayerBaseChildren(file: string, text: string, violations: string[]): Node[] {
  const nodes = scanNodes(text, 0, text.length);
  const isSingleLayerBase = (node: Node): node is AtRuleNode =>
    node.kind === 'at-rule' && node.name === '@layer' && node.prelude === 'base';
  if (nodes.length === 1 && isSingleLayerBase(nodes[0])) {
    return nodes[0].children;
  }
  if (nodes.length === 0) {
    violations.push(`g2: ${file}: no top-level @layer base block found`);
  } else {
    for (const node of nodes) {
      violations.push(`g2: ${file}: unexpected top-level ${describeTopLevelNode(node)}, want exactly one @layer base block`);
    }
  }
  return [];
}

function checkCompleteness(
  where: string,
  declared: string[],
  expected: readonly string[],
  violations: string[]
): void {
  const declaredSet = new Set(declared);
  const expectedSet = new Set(expected);
  for (const name of expected) {
    if (!declaredSet.has(name)) violations.push(`g2: ${where} is missing ${name}`);
  }
  for (const name of declared) {
    if (!expectedSet.has(name)) violations.push(`g2: ${where} carries an extra name ${name}`);
  }
}

function checkDuplicates(where: string, declarations: DeclNode[], violations: string[]): void {
  const counts = new Map<string, number>();
  for (const d of declarations) counts.set(d.name, (counts.get(d.name) ?? 0) + 1);
  for (const [name, count] of counts) {
    if (count > 1) violations.push(`g2: ${where} declares ${name} twice`);
  }
}

function checkNoMetadataFlags(where: string, node: PluginNode, violations: string[]): void {
  for (const key of ['default', 'prefersdark']) {
    if (node.declarations.some((d) => d.name === key)) {
      violations.push(`g2: ${where} carries a ${key} key`);
    }
  }
}

/** The CSS Fonts Module 4 generic-family keywords a font stack's last entry must be one of. */
const GENERIC_FAMILIES = new Set([
  'serif',
  'sans-serif',
  'cursive',
  'fantasy',
  'monospace',
  'system-ui',
  'ui-serif',
  'ui-sans-serif',
  'ui-monospace',
  'ui-rounded',
  'math'
]);

function checkFontGenericTail(where: string, name: string, value: string, violations: string[]): void {
  const parts = value.split(',').map((p) => p.trim());
  const last = parts[parts.length - 1];
  if (!GENERIC_FAMILIES.has(last)) {
    violations.push(`g2: ${where} ${name} does not end in a CSS generic family (last entry: "${last}")`);
  }
}

/** The 44px touch-target floor, from DaisyUI's own `--size:calc(var(--size-field) * 10)` (`components/button.css`), 1rem = 16px. */
function checkSizeFieldFloor(where: string, value: string, violations: string[]): void {
  const m = /^(-?(?:\d+\.?\d*|\.\d+))rem$/.exec(value);
  if (!m) {
    violations.push(`g2: ${where} --size-field is not a bare rem value ("${value}")`);
    return;
  }
  const px = Number(m[1]) * 10 * 16;
  if (px < 44) {
    violations.push(`g2: ${where} --size-field ${value} yields ${px}px, below the 44px floor`);
  }
}

export function declMap(declarations: DeclNode[]): Map<string, string> {
  const map = new Map<string, string>();
  for (const d of declarations) map.set(d.name, d.value);
  return map;
}

export function run(): string[] {
  const violations: string[] = [];
  const files = readThemeFiles(violations);
  if (!files) return violations;

  // ---- dubplate.css: the match axis, two identical structural blocks.
  const matchBlocks = pluginBlocksNamed(files.dubplate);
  for (const composed of COMPOSED_NAMES) {
    const block = matchBlocks.get(composed);
    if (!block) {
      violations.push(`g2: dubplate.css: no match block named "${composed}"`);
      continue;
    }
    const where = `dubplate.css match block "${composed}"`;
    const names = block.declarations.map((d) => d.name).filter((n) => !PLUGIN_METADATA_KEYS.has(n));
    checkCompleteness(where, names, STRUCTURAL_NAMES, violations);
    checkDuplicates(where, block.declarations, violations);
    checkNoMetadataFlags(where, block, violations);
    const map = declMap(block.declarations);
    for (const fontName of ['--font-sans', '--font-serif', '--font-mono']) {
      const value = map.get(fontName);
      if (value) checkFontGenericTail(where, fontName, value, violations);
    }
    const sizeField = map.get('--size-field');
    if (sizeField) checkSizeFieldFloor(where, sizeField, violations);
  }
  const labelMatch = matchBlocks.get('dubplate-label');
  const nightMatch = matchBlocks.get('dubplate-label-night');
  if (labelMatch && nightMatch) {
    const labelMap = declMap(labelMatch.declarations);
    const nightMap = declMap(nightMatch.declarations);
    for (const name of STRUCTURAL_NAMES) {
      const a = labelMap.get(name);
      const b = nightMap.get(name);
      if (a !== undefined && b !== undefined && a !== b) {
        violations.push(
          `g2: dubplate.css: the two match blocks disagree on ${name} ("${a}" vs "${b}")`
        );
      }
    }
  }

  // ---- label.css / label-night.css: one palette block each.
  const paletteFiles: [string, string, string][] = [
    ['label.css', files.label, 'dubplate-label'],
    ['label-night.css', files.labelNight, 'dubplate-label-night']
  ];
  const paletteDecls = new Map<string, DeclNode[]>();
  for (const [file, text, composed] of paletteFiles) {
    const blocks = pluginBlocksNamed(text);
    const block = blocks.get(composed);
    if (!block) {
      violations.push(`g2: ${file}: no palette block named "${composed}"`);
      continue;
    }
    paletteDecls.set(composed, block.declarations);
    const where = `${file} palette block "${composed}"`;
    const names = block.declarations.map((d) => d.name).filter((n) => !PLUGIN_METADATA_KEYS.has(n));
    checkCompleteness(where, names, PALETTE_NAMES, violations);
    checkDuplicates(where, block.declarations, violations);
    checkNoMetadataFlags(where, block, violations);
  }

  // ---- inks.css: one ink block per composed theme, inside `@layer base`.
  const inkDecls = new Map<string, DeclNode[]>();
  const inkRules = singleLayerBaseChildren('inks.css', files.inks, violations).filter(
    (n): n is RuleNode => n.kind === 'rule'
  );
  for (const composed of COMPOSED_NAMES) {
    const rule = inkRules.find((r) => selectorHasDataTheme(r.selector, composed));
    if (!rule) {
      violations.push(`g2: inks.css: no ink block for "${composed}"`);
      continue;
    }
    inkDecls.set(composed, rule.declarations);
    const where = `inks.css ink block "${composed}"`;
    checkCompleteness(where, rule.declarations.map((d) => d.name), INK_NAMES, violations);
    checkDuplicates(where, rule.declarations, violations);
  }

  // ---- system.css: the unconditional branch and the dark-media branch.
  const sysChildren = singleLayerBaseChildren('system.css', files.system, violations);
  const sysRules = sysChildren.filter((n): n is RuleNode => n.kind === 'rule');
  const sysMedia = sysChildren.filter((n): n is AtRuleNode => n.kind === 'at-rule' && n.name === '@media');

  const lightCandidates = sysRules.filter((r) => selectorHasDataTheme(r.selector, SYSTEM_NAME));
  let lightBranch: RuleNode | undefined;
  for (const candidate of lightCandidates) {
    if (selectorIsExactlyDataTheme(candidate.selector, SYSTEM_NAME)) {
      lightBranch = candidate;
    } else {
      violations.push(
        `g2: system.css: the unconditional branch holds an unexpected selector "${candidate.selector}"`
      );
    }
  }
  if (!lightBranch) {
    violations.push('g2: system.css: the unconditional light branch is missing');
  }

  const darkMedia = sysMedia.filter((m) => normalizeMediaQuery(m.prelude) === DARK_QUERY);
  if (darkMedia.length === 0) {
    violations.push('g2: system.css: no @media (prefers-color-scheme: dark) block');
  }
  for (const media of sysMedia) {
    if (normalizeMediaQuery(media.prelude) !== DARK_QUERY) {
      violations.push(`g2: system.css: unexpected @media (${media.prelude})`);
    }
  }
  let darkBranch: RuleNode | undefined;
  for (const media of darkMedia) {
    for (const child of media.children) {
      if (child.kind !== 'rule') continue;
      if (selectorIsExactlyDataTheme(child.selector, SYSTEM_NAME)) {
        darkBranch = child;
      } else {
        violations.push(
          `g2: system.css: the prefers-color-scheme block holds an unexpected selector "${child.selector}"`
        );
      }
    }
  }
  if (darkMedia.length > 0 && !darkBranch) {
    violations.push('g2: system.css: the dark branch is missing');
  }

  const expectedBranchNames = [...STRUCTURAL_NAMES, ...PALETTE_NAMES, ...INK_NAMES];
  const branches: [string, RuleNode | undefined, string][] = [
    ['light', lightBranch, 'dubplate-label'],
    ['dark', darkBranch, 'dubplate-label-night']
  ];
  for (const [label, branch, composed] of branches) {
    if (!branch) continue;
    const where = `system.css ${label} branch`;
    checkCompleteness(where, branch.declarations.map((d) => d.name), expectedBranchNames, violations);
    checkDuplicates(where, branch.declarations, violations);

    const reference = new Map<string, string>([
      ...declMap((matchBlocks.get(composed)?.declarations ?? []).filter((d) => !PLUGIN_METADATA_KEYS.has(d.name))),
      ...declMap(paletteDecls.get(composed) ?? []),
      ...declMap(inkDecls.get(composed) ?? [])
    ]);
    for (const decl of branch.declarations) {
      const want = reference.get(decl.name);
      if (want !== undefined && want !== decl.value) {
        violations.push(
          `g2: system.css: ${label} branch's ${decl.name} ("${decl.value}") differs from ${composed}'s ("${want}")`
        );
      }
    }
  }

  return violations;
}

function main(): void {
  const violations = run();
  if (violations.length > 0) {
    fail(`g2: ${violations.length} finding(s):\n${violations.join('\n')}`);
  }
  process.stdout.write('g2: passed, the theme data files are complete and at parity.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
