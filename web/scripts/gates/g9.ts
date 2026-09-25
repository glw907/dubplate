/**
 * G9: contrast, extended for the status inks. For each composed theme
 * in `label.css` and `label-night.css`, computes WCAG 2.2 relative
 * luminance in sRGB (converting `oklch()` first) and checks:
 *
 * - Each `-content` pair against its own fill, and `--color-base-content`
 *   against `--color-base-100`, at 4.5:1.
 * - The eight colors DaisyUI paints as text directly (`--color-primary`
 *   through `--color-info`) against `--color-base-100`, at 4.5:1,
 *   unless a `contrast-exception` marker directly above the declaration
 *   records a lower measured ratio. A marker above a pairing that
 *   clears the floor is itself a failure: a stale exemption cannot
 *   survive an edit.
 * - Each ink in `inks.css`, resolved through `var()` against the same
 *   composed theme's own palette and nothing else, against
 *   `--color-base-100`, at 4.5:1. No ink admits a marker.
 */
import { readFileSync } from 'node:fs';
import { fail, isMainModule } from './lib.ts';
import {
  COMPOSED_NAMES,
  type DeclNode,
  declMap,
  type ExceptionMarker,
  pluginBlocksNamed,
  selectorHasDataTheme,
  singleLayerBaseChildren,
  type RuleNode
} from './g2.ts';

const THEMES_DIR = 'src/lib/themes';

/** The eight colors DaisyUI paints as text directly rather than through a `-content` token. */
const DIRECT_TEXT_COLORS = [
  '--color-primary',
  '--color-secondary',
  '--color-accent',
  '--color-neutral',
  '--color-info',
  '--color-success',
  '--color-warning',
  '--color-error'
];

/** Every `-content` pair: the content name and its fill. */
const CONTENT_PAIRS: [string, string][] = [
  ['--color-primary-content', '--color-primary'],
  ['--color-secondary-content', '--color-secondary'],
  ['--color-accent-content', '--color-accent'],
  ['--color-neutral-content', '--color-neutral'],
  ['--color-info-content', '--color-info'],
  ['--color-success-content', '--color-success'],
  ['--color-warning-content', '--color-warning'],
  ['--color-error-content', '--color-error']
];

/**
 * The 4.5:1 contrast floor. The design-language standard's palette-pair
 * ratio (DL R2.6) and the web surface standard's ink ratio (W2.7) are the
 * same number, so both pairs and inks share this one constant.
 */
const FLOOR = 4.5;

function srgbToLinearChannel(c: number): number {
  return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4;
}

function parseHex(value: string): [number, number, number] | null {
  const m = /^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$/.exec(value.trim());
  if (!m) return null;
  const digits = m[1].length === 3 ? m[1].split('').map((c) => c + c).join('') : m[1];
  return [
    parseInt(digits.slice(0, 2), 16) / 255,
    parseInt(digits.slice(2, 4), 16) / 255,
    parseInt(digits.slice(4, 6), 16) / 255
  ];
}

/**
 * `oklch()` to linear sRGB, via OKLab, Björn Ottosson's published
 * matrices (adopted into CSS Color 4's own conversion appendix). The
 * result is already linear-light, so `relativeLuminance` below skips
 * `srgbToLinearChannel` for this branch.
 */
function parseOklchToLinear(value: string): [number, number, number] | null {
  const m = /^oklch\(\s*([\d.]+)(%)?\s+([\d.]+)\s+([\d.]+)(?:deg)?\s*(?:\/\s*[\d.%]+\s*)?\)$/i.exec(
    value.trim()
  );
  if (!m) return null;
  const L = m[2] ? Number(m[1]) / 100 : Number(m[1]);
  const C = Number(m[3]);
  const H = (Number(m[4]) * Math.PI) / 180;
  const a = C * Math.cos(H);
  const b = C * Math.sin(H);
  const l_ = L + 0.3963377774 * a + 0.2158037573 * b;
  const m_ = L - 0.1055613458 * a - 0.0638541728 * b;
  const s_ = L - 0.0894841775 * a - 1.2914855480 * b;
  const l = l_ ** 3;
  const mm = m_ ** 3;
  const s = s_ ** 3;
  return [
    4.0767416621 * l - 3.3077115913 * mm + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * mm - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * mm + 1.7076147010 * s
  ];
}

/** WCAG 2.2 relative luminance, or `null` when `value` is neither a hex color nor `oklch()`. */
function relativeLuminance(value: string): number | null {
  const hex = parseHex(value);
  if (hex) {
    const [r, g, b] = hex.map(srgbToLinearChannel) as [number, number, number];
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  }
  const oklch = parseOklchToLinear(value);
  if (oklch) {
    const [r, g, b] = oklch.map((c) => Math.min(1, Math.max(0, c))) as [number, number, number];
    return 0.2126 * r + 0.7152 * g + 0.0722 * b;
  }
  return null;
}

function formatRatio(ratio: number): string {
  return `${ratio.toFixed(2)}:1`;
}

/** `var(--color-*)` resolved against `palette`, or `null` when it names anything else (never silently skipped). */
function resolveAgainstPalette(value: string, palette: Map<string, string>): string | null {
  const m = /^var\((--[\w-]+)\)$/.exec(value.trim());
  if (!m) return value;
  const name = m[1];
  if (!name.startsWith('--color-')) return null;
  return palette.get(name) ?? null;
}

/** The 3:1 tier a marker's reason may cite (WCAG 1.4.11 non-text contrast), one decimal below `FLOOR`. */
const CITED_TIER = 3.0;
const CITES_THREE_TO_ONE = /\b3(?:\.0+)?:1\b/;
/** The largest gap between a marker's recorded ratio and the measured one still counted as the same number, absorbing `toFixed(2)` rounding. */
const RATIO_TOLERANCE = 0.005;

function checkFloor(
  where: string,
  variable: string,
  text: string,
  base: string,
  rawMarker: ExceptionMarker | null,
  allowMarker: boolean,
  violations: string[]
): void {
  let marker = rawMarker;
  if (marker && !allowMarker) {
    violations.push(`g9: ${where}: ${variable} carries an exception marker on a pairing that admits none`);
    marker = null;
  }
  const textLum = relativeLuminance(text);
  const baseLum = relativeLuminance(base);
  if (textLum === null) {
    violations.push(`g9: ${where}: ${variable}'s value "${text}" is not a parseable color`);
  }
  if (baseLum === null) {
    violations.push(`g9: ${where}: ${variable}'s base --color-base-100 value "${base}" is not a parseable color`);
  }
  if (textLum === null || baseLum === null) return;
  const lighter = Math.max(textLum, baseLum);
  const darker = Math.min(textLum, baseLum);
  const ratio = (lighter + 0.05) / (darker + 0.05);
  const passes = ratio >= FLOOR;
  if (marker) {
    if (marker.variable !== variable) {
      violations.push(`g9: ${where}: the marker above ${variable} names ${marker.variable} instead`);
    }
    const recorded = Number(marker.ratio);
    if (Math.abs(recorded - ratio) >= RATIO_TOLERANCE) {
      violations.push(
        `g9: ${where}: ${variable}'s marker records ${formatRatio(recorded)}, measured ${formatRatio(ratio)}`
      );
    }
    if (CITES_THREE_TO_ONE.test(marker.reason) && ratio < CITED_TIER) {
      violations.push(
        `g9: ${where}: ${variable}'s reason cites the 3:1 tier but measures ${formatRatio(ratio)}, below it`
      );
    }
    if (passes) {
      violations.push(
        `g9: ${where}: ${variable}'s exception marker is stale, the pair now clears the floor at ${formatRatio(ratio)}`
      );
    }
    return;
  }
  if (!passes) {
    violations.push(
      `g9: ${where}: ${variable} measures ${formatRatio(ratio)} against ${FLOOR}:1 and carries no exception marker`
    );
  }
}

function readPaletteBlock(file: string, composed: string): DeclNode[] | null {
  const text = readFileSync(`${THEMES_DIR}/${file}`, 'utf8');
  return pluginBlocksNamed(text).get(composed)?.declarations ?? null;
}

function readInkRules(violations: string[]): RuleNode[] {
  const text = readFileSync(`${THEMES_DIR}/inks.css`, 'utf8');
  return singleLayerBaseChildren('inks.css', text, violations).filter((n): n is RuleNode => n.kind === 'rule');
}

const PALETTE_FILES: Record<string, string> = {
  'dubplate-label': 'label.css',
  'dubplate-label-night': 'label-night.css'
};

export function run(): string[] {
  const violations: string[] = [];
  const inkRules = readInkRules(violations);

  for (const composed of COMPOSED_NAMES) {
    const declarations = readPaletteBlock(PALETTE_FILES[composed], composed);
    if (!declarations) {
      violations.push(`g9: ${composed}: no composed theme found, an empty iteration is not a pass`);
      continue;
    }
    const palette = declMap(declarations);
    const base100 = palette.get('--color-base-100');
    if (!base100) {
      violations.push(`g9: ${composed}: --color-base-100 is missing, contrast cannot be measured`);
      continue;
    }

    for (const [contentName, fillName] of CONTENT_PAIRS) {
      const contentDecl = declarations.find((d) => d.name === contentName);
      const fillValue = palette.get(fillName);
      if (!contentDecl) {
        violations.push(`g9: ${composed}: ${contentName} is missing, contrast cannot be measured`);
        continue;
      }
      if (fillValue === undefined) {
        violations.push(`g9: ${composed}: ${fillName} is missing, ${contentName}'s contrast cannot be measured`);
        continue;
      }
      // -content pairs are always text: never a place for a contrast-exception marker.
      checkFloor(composed, contentName, contentDecl.value, fillValue, contentDecl.marker, false, violations);
    }
    const baseContentDecl = declarations.find((d) => d.name === '--color-base-content');
    if (!baseContentDecl) {
      violations.push(`g9: ${composed}: --color-base-content is missing, contrast cannot be measured`);
    } else {
      // base-content is always text too: same restriction as the -content pairs above.
      checkFloor(
        composed,
        '--color-base-content',
        baseContentDecl.value,
        base100,
        baseContentDecl.marker,
        false,
        violations
      );
    }
    for (const name of DIRECT_TEXT_COLORS) {
      const decl = declarations.find((d) => d.name === name);
      if (!decl) {
        violations.push(`g9: ${composed}: ${name} is missing, contrast cannot be measured`);
        continue;
      }
      checkFloor(composed, name, decl.value, base100, decl.marker, true, violations);
    }

    const inkRule = inkRules.find((r) => selectorHasDataTheme(r.selector, composed));
    if (!inkRule || inkRule.declarations.length === 0) {
      violations.push(`g9: ${composed}: no ink block found, an empty iteration is not a pass`);
      continue;
    }
    for (const ink of inkRule.declarations) {
      if (ink.marker) {
        violations.push(`g9: ${composed}: ${ink.name} carries an exception marker; inks admit none`);
        continue;
      }
      const resolved = resolveAgainstPalette(ink.value, palette);
      if (resolved === null) {
        violations.push(`g9: ${composed}: ${ink.name}'s value "${ink.value}" does not resolve against this theme's own --color-* set`);
        continue;
      }
      checkFloor(composed, ink.name, resolved, base100, null, false, violations);
    }
  }

  return violations;
}

function main(): void {
  const violations = run();
  if (violations.length > 0) {
    fail(`g9: ${violations.length} finding(s):\n${violations.join('\n')}`);
  }
  process.stdout.write('g9: passed, every pairing clears its floor or carries a live exception.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
