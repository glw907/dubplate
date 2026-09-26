/**
 * G1: no hardcoded colors. `src/lib/themes/` is the one exempt path,
 * since it is the layer that defines the palette values every other
 * file must reach through a token instead.
 */
import { readFileSync } from 'node:fs';
import { join, sep } from 'node:path';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';
const EXEMPT_PREFIX = `lib${sep}themes${sep}`;
const SCANNED_EXTENSIONS = ['.svelte', '.ts', '.js', '.css', '.html', '.svg'];

const COLOR_FUNCTIONS = /\b(rgb|rgba|hsl|hsla|oklch|oklab|lab|lch|hwb|color)\(/;
const HEX_LITERAL = /#[0-9a-fA-F]{3,8}\b/;
/**
 * DaisyUI's own status text utilities (W2.7: "a fill is not an ink").
 * `text-primary`, `-secondary`, `-accent`, `-neutral`, `-info`,
 * `-success`, `-warning`, and `-error` reach a raw fill, never the
 * measured on-surface ink; markup reaches a status color only through
 * `text-ink-primary`, `-accent`, `-info`, `-error`, or `-warning`
 * (`lib/styles/components.css`). The negative lookahead refuses the
 * bare utility while admitting its `text-ink-*` sibling.
 */
const STATUS_TEXT_UTILITY = /\btext-(?!ink-)(primary|secondary|accent|neutral|info|success|warning|error)\b/;
/** A Svelte `style:<prop>=` directive, the attribute-shaped sibling of a CSS `prop: value;` declaration. */
const STYLE_DIRECTIVE =
  /\bstyle:(color|background(-color)?|border(-[a-z]+)?-color|fill|stroke|outline-color|box-shadow|text-decoration-color|caret-color|accent-color)\s*=\s*["']([^"']*)["']/gi;
/** SVG's own color-bearing presentation attributes, never routed through G1's CSS-declaration scan. */
const SVG_COLOR_ATTR = /\b(fill|stroke|stop-color)\s*=\s*["']([^"']*)["']/gi;

const TAILWIND_COLOR_PREFIXES = [
  'bg',
  'text',
  'border',
  'fill',
  'stroke',
  'ring',
  'shadow',
  'divide',
  'outline',
  'decoration',
  'accent',
  'caret',
  'from',
  'via',
  'to'
];
const TAILWIND_FAMILY =
  '(slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(50|[1-9]00|950)';
const TAILWIND_BARE = '(white|black|transparent|current)';
const TAILWIND_UTILITY = new RegExp(
  `\\b(${TAILWIND_COLOR_PREFIXES.join('|')})-(${TAILWIND_FAMILY}|${TAILWIND_BARE})\\b`
);

// CSS Color Module 4's named-color list, matched only against a
// color-bearing declaration's value, never a bare word anywhere in a
// file.
const CSS_NAMED_COLORS = new Set([
  'aliceblue', 'antiquewhite', 'aqua', 'aquamarine', 'azure', 'beige', 'bisque', 'black',
  'blanchedalmond', 'blue', 'blueviolet', 'brown', 'burlywood', 'cadetblue', 'chartreuse',
  'chocolate', 'coral', 'cornflowerblue', 'cornsilk', 'crimson', 'cyan', 'darkblue', 'darkcyan',
  'darkgoldenrod', 'darkgray', 'darkgreen', 'darkgrey', 'darkkhaki', 'darkmagenta',
  'darkolivegreen', 'darkorange', 'darkorchid', 'darkred', 'darksalmon', 'darkseagreen',
  'darkslateblue', 'darkslategray', 'darkslategrey', 'darkturquoise', 'darkviolet', 'deeppink',
  'deepskyblue', 'dimgray', 'dimgrey', 'dodgerblue', 'firebrick', 'floralwhite', 'forestgreen',
  'fuchsia', 'gainsboro', 'ghostwhite', 'gold', 'goldenrod', 'gray', 'green', 'greenyellow',
  'grey', 'honeydew', 'hotpink', 'indianred', 'indigo', 'ivory', 'khaki', 'lavender',
  'lavenderblush', 'lawngreen', 'lemonchiffon', 'lightblue', 'lightcoral', 'lightcyan',
  'lightgoldenrodyellow', 'lightgray', 'lightgreen', 'lightgrey', 'lightpink', 'lightsalmon',
  'lightseagreen', 'lightskyblue', 'lightslategray', 'lightslategrey', 'lightsteelblue',
  'lightyellow', 'lime', 'limegreen', 'linen', 'magenta', 'maroon', 'mediumaquamarine',
  'mediumblue', 'mediumorchid', 'mediumpurple', 'mediumseagreen', 'mediumslateblue',
  'mediumspringgreen', 'mediumturquoise', 'mediumvioletred', 'midnightblue', 'mintcream',
  'mistyrose', 'moccasin', 'navajowhite', 'navy', 'oldlace', 'olive', 'olivedrab', 'orange',
  'orangered', 'orchid', 'palegoldenrod', 'palegreen', 'paleturquoise', 'palevioletred',
  'papayawhip', 'peachpuff', 'peru', 'pink', 'plum', 'powderblue', 'purple', 'rebeccapurple',
  'red', 'rosybrown', 'royalblue', 'saddlebrown', 'salmon', 'sandybrown', 'seagreen', 'seashell',
  'sienna', 'silver', 'skyblue', 'slateblue', 'slategray', 'slategrey', 'snow', 'springgreen',
  'steelblue', 'tan', 'teal', 'thistle', 'tomato', 'turquoise', 'violet', 'wheat', 'white',
  'whitesmoke', 'yellow', 'yellowgreen'
]);
const COLOR_BEARING_PROPERTY =
  /\b(color|background(-color)?|border(-[a-z]+)?-color|fill|stroke|outline-color|box-shadow|text-decoration-color|caret-color|accent-color)\s*:\s*([^;{}]+)/gi;

export interface Violation {
  path: string;
  line: number;
  pattern: string;
}

function scanLine(path: string, lineNo: number, line: string, violations: Violation[]): void {
  if (COLOR_FUNCTIONS.test(line)) {
    violations.push({ path, line: lineNo, pattern: 'color function' });
  }
  if (HEX_LITERAL.test(line)) {
    violations.push({ path, line: lineNo, pattern: 'hex literal' });
  }
  if (TAILWIND_UTILITY.test(line)) {
    violations.push({ path, line: lineNo, pattern: 'Tailwind color utility' });
  }
  const statusMatch = STATUS_TEXT_UTILITY.exec(line);
  if (statusMatch !== null) {
    violations.push({ path, line: lineNo, pattern: `DaisyUI status text utility "text-${statusMatch[1]}"` });
  }
  let m: RegExpExecArray | null;
  COLOR_BEARING_PROPERTY.lastIndex = 0;
  while ((m = COLOR_BEARING_PROPERTY.exec(line))) {
    const value = m[m.length - 1].toLowerCase();
    for (const word of value.split(/[^a-z]+/)) {
      if (CSS_NAMED_COLORS.has(word)) {
        violations.push({ path, line: lineNo, pattern: `named color "${word}"` });
      }
    }
  }
  STYLE_DIRECTIVE.lastIndex = 0;
  while ((m = STYLE_DIRECTIVE.exec(line))) {
    const value = m[m.length - 1].toLowerCase();
    for (const word of value.split(/[^a-z]+/)) {
      if (CSS_NAMED_COLORS.has(word)) {
        violations.push({ path, line: lineNo, pattern: `named color "${word}" in a style: directive` });
      }
    }
  }
  SVG_COLOR_ATTR.lastIndex = 0;
  while ((m = SVG_COLOR_ATTR.exec(line))) {
    const attr = m[1];
    const value = m[2].trim().toLowerCase();
    if (value === '' || value === 'none' || value === 'currentcolor' || /^(var|url)\(/.test(value)) continue;
    violations.push({ path, line: lineNo, pattern: `SVG ${attr} attribute "${value}"` });
  }
}

/**
 * `/* *\/`, `<!-- -->`, and `//` (`.ts`/`.js` only, since `//` is not a
 * comment marker in CSS, Svelte markup, or HTML) blanked to spaces so a
 * hex literal or a named color inside a comment never counts as a
 * violation, while every line number and column stays put for the
 * caller's line-based scan.
 */
export function stripComments(path: string, contents: string): string {
  const blank = (match: string) => match.replace(/[^\n]/g, ' ');
  let out = contents.replace(/\/\*[\s\S]*?\*\//g, blank).replace(/<!--[\s\S]*?-->/g, blank);
  if (path.endsWith('.ts') || path.endsWith('.js')) {
    out = out
      .split('\n')
      .map((line) => line.replace(/\/\/.*/, blank))
      .join('\n');
  }
  return out;
}

export function run(root: string): Violation[] {
  const violations: Violation[] = [];
  for (const rel of walk(root)) {
    if (rel.startsWith(EXEMPT_PREFIX)) continue;
    if (!SCANNED_EXTENSIONS.some((ext) => rel.endsWith(ext))) continue;
    const contents = stripComments(rel, readFileSync(join(root, rel), 'utf8'));
    contents.split('\n').forEach((line, i) => scanLine(rel, i + 1, line, violations));
  }
  return violations;
}

function main() {
  const files = walk(SRC_ROOT);
  if (files.length === 0) {
    fail(`g1: ${SRC_ROOT} is absent or empty; an empty iteration is not a pass.`);
  }
  const violations = run(SRC_ROOT);
  if (violations.length > 0) {
    const lines = violations.map((v) => `${v.path}:${v.line}: ${v.pattern}`);
    fail(`g1: ${violations.length} hardcoded color(s):\n${lines.join('\n')}`);
  }
  process.stdout.write(`g1: passed, read ${files.length} file(s), no hardcoded color found.\n`);
}

if (isMainModule(import.meta.url)) {
  main();
}
