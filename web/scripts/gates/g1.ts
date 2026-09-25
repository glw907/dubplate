/**
 * G1: no hardcoded colors. `src/lib/themes/` is the one exempt path,
 * since it is the layer that defines the palette values every other
 * file must reach through a token instead.
 */
import { readFileSync } from 'node:fs';
import { join, sep } from 'node:path';
import { fail, walk } from './lib.ts';

const SRC_ROOT = 'src';
const EXEMPT_PREFIX = `lib${sep}themes${sep}`;
const SCANNED_EXTENSIONS = ['.svelte', '.ts', '.js', '.css', '.html'];

const COLOR_FUNCTIONS = /\b(rgb|rgba|hsl|hsla|oklch|oklab|lab|lch|hwb|color)\(/;
const HEX_LITERAL = /#[0-9a-fA-F]{3,8}\b/;

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
  if (HEX_LITERAL.test(line) && /(class|style)\s*=|:\s*#/.test(line)) {
    violations.push({ path, line: lineNo, pattern: 'hex literal' });
  }
  if (TAILWIND_UTILITY.test(line)) {
    violations.push({ path, line: lineNo, pattern: 'Tailwind color utility' });
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
}

export function run(root: string): Violation[] {
  const violations: Violation[] = [];
  for (const rel of walk(root)) {
    if (rel.startsWith(EXEMPT_PREFIX)) continue;
    if (!SCANNED_EXTENSIONS.some((ext) => rel.endsWith(ext))) continue;
    const contents = readFileSync(join(root, rel), 'utf8');
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

if (import.meta.url === `file://${process.argv[1]}`) {
  main();
}
