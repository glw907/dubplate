/**
 * G9: contrast. For each composed theme in `label.css` and
 * `label-night.css`, computes WCAG 2.2 relative luminance in sRGB
 * (converting `oklch()` first) and checks a closed list of pairs
 * (DL R2.6):
 *
 * - Each `-content` pair against its own fill, at 4.5:1.
 * - `--color-base-content` against each of `--color-base-100`,
 *   `--color-base-200`, and `--color-base-300`, at 4.5:1.
 * - Each of the eight colors DaisyUI paints as text directly
 *   (`--color-primary` through `--color-info`) against
 *   `--color-base-100`, at 4.5:1, unless a `contrast-exception` marker
 *   directly above the declaration records a lower measured ratio. A
 *   marker above a pairing that clears the floor is itself a failure: a
 *   stale exemption cannot survive an edit.
 * - Each ink in `inks.css`, resolved through `var()` against the same
 *   composed theme's own palette and nothing else, against each of
 *   the three bases, at 4.5:1. No ink admits a marker.
 * - A pinned-value row: each resolved ink compared to the visual
 *   identity spec's ink table (quoted below), so an ink that drifts
 *   from the spec fails even when it still passes contrast.
 * - `surface.css`'s `--keyline-base` against each of the three bases,
 *   at 3:1 (WCAG 1.4.11, a control's edge). A pair under the floor
 *   needs a `contrast-exception` marker above the declaration that
 *   records its ratio and names that base; a marker no failing pair
 *   matches is stale. `--keyline-fill` is not measured: it borders a
 *   fill whose own contrast identifies the object (VI's dark keylines).
 * - `--color-base-content` against `--selection-ground`, the
 *   `color-mix()` in OKLab resolved here, at 4.5:1.
 *
 * `label`'s `--color-secondary` and `label-night`'s `--color-neutral`
 * are fill-only (`FILL_ONLY_EXCLUSIONS` below, VI "Themes"): excluded from
 * the direct-text floor by name, but each still must carry a
 * `contrast-exception` marker reading "fill only, never rendered as
 * text" (W2.6), which G9 checks against the measured ratio without
 * gating the floor.
 */
import { readFileSync } from 'node:fs';
import { fail, isMainModule } from './lib.ts';
import {
  COMPOSED_NAMES,
  type DeclNode,
  declMap,
  type ExceptionMarker,
  INK_NAMES,
  pluginBlocksNamed,
  selectorHasDataTheme,
  singleLayerBaseChildren,
  type RuleNode
} from './g2.ts';

const THEMES_DIR = 'src/lib/themes';

/** The three grounds every ink and `--color-base-content` must clear (DL R2.6). */
const BASES = ['--color-base-100', '--color-base-200', '--color-base-300'];

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

/**
 * DL R2.6's closed list, quoted independently of `DIRECT_TEXT_COLORS`
 * so a deletion from that array is still caught: the two must agree,
 * checked below in `run()`, rather than one being derived from the
 * other.
 */
const DL_R2_6_DIRECT_TEXT_COLORS = [
  '--color-primary',
  '--color-secondary',
  '--color-accent',
  '--color-neutral',
  '--color-error',
  '--color-success',
  '--color-warning',
  '--color-info'
];

/**
 * The two fills VI marks fill-only (visual identity spec, "Themes"): never
 * measured against the direct-text floor, but each still carries a
 * `contrast-exception` marker, the same text as its own file
 * (`fill only, never rendered as text`), that G9 requires and checks.
 */
const FILL_ONLY_EXCLUSIONS: { composed: string; name: string; reason: string }[] = [
  { composed: 'dubplate-label', name: '--color-secondary', reason: 'fill only, never rendered as text' },
  { composed: 'dubplate-label-night', name: '--color-neutral', reason: 'fill only, never rendered as text' }
];

/**
 * The fill-only pair in VI "Themes", transcribed independently of
 * `FILL_ONLY_EXCLUSIONS` so an entry added, renamed, or dropped there
 * is caught rather than silently trusted, the same discipline as
 * `DL_R2_6_DIRECT_TEXT_COLORS` above.
 */
const VI_FILL_ONLY: { composed: string; name: string }[] = [
  { composed: 'dubplate-label', name: '--color-secondary' },
  { composed: 'dubplate-label-night', name: '--color-neutral' }
];

function isFillOnly(composed: string, name: string): boolean {
  return FILL_ONLY_EXCLUSIONS.some((e) => e.composed === composed && e.name === name);
}

/**
 * The visual identity spec's ink table, quoted verbatim
 * (docs/superpowers/specs/2026-09-24-visual-identity-design.md, "Themes"),
 * `label` then `label-night`. An ink whose resolved value drifts from
 * this table fails the pinned-value row even when it still clears the
 * contrast floor.
 */
const PINNED_INKS: Record<string, Record<string, string>> = {
  'dubplate-label': {
    '--ink-primary': '#9d2f26',
    '--ink-accent': '#2d6136',
    '--ink-info': '#2d5a81',
    '--ink-error': '#a4261d',
    '--ink-warning': '#754f00'
  },
  'dubplate-label-night': {
    '--ink-primary': '#ec7b6a',
    '--ink-accent': '#66a46e',
    '--ink-info': '#6d9cc7',
    '--ink-error': '#e37866',
    '--ink-warning': '#e5ad33'
  }
};

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

/**
 * The measured ratio between `text` and `base`, pushing a violation for
 * either color that fails to parse. Returns `null` when either did, the
 * shared preamble `checkFloor` and `checkFillOnlyMarker` both need
 * before they diverge on what a marker requires.
 */
function measureRatio(where: string, variable: string, text: string, base: string, violations: string[]): number | null {
  const textLum = relativeLuminance(text);
  const baseLum = relativeLuminance(base);
  if (textLum === null) {
    violations.push(`g9: ${where}: ${variable}'s value "${text}" is not a parseable color`);
  }
  if (baseLum === null) {
    violations.push(`g9: ${where}: ${variable}'s base value "${base}" is not a parseable color`);
  }
  if (textLum === null || baseLum === null) return null;
  const lighter = Math.max(textLum, baseLum);
  const darker = Math.min(textLum, baseLum);
  return (lighter + 0.05) / (darker + 0.05);
}

/**
 * Checks one text/base pairing against `FLOOR`, pushing a violation for
 * any failure and, on success, a report line to `reportLines`. Returns
 * the measured ratio, or `null` when either color failed to parse.
 */
function checkFloor(
  where: string,
  variable: string,
  text: string,
  base: string,
  rawMarker: ExceptionMarker | null,
  allowMarker: boolean,
  violations: string[],
  reportLines: string[]
): number | null {
  let marker = rawMarker;
  if (marker && !allowMarker) {
    violations.push(`g9: ${where}: ${variable} carries an exception marker on a pairing that admits none`);
    marker = null;
  }
  const ratio = measureRatio(where, variable, text, base, violations);
  if (ratio === null) return null;
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
    reportLines.push(`g9: ${where}: ${variable} = ${formatRatio(ratio)} (exception: ${marker.reason})`);
    return ratio;
  }
  if (!passes) {
    violations.push(
      `g9: ${where}: ${variable} measures ${formatRatio(ratio)} against ${FLOOR}:1 and carries no exception marker`
    );
  }
  reportLines.push(`g9: ${where}: ${variable} = ${formatRatio(ratio)}`);
  return ratio;
}

/**
 * Checks a fill-only color's required `contrast-exception` marker
 * (VI "Themes", W2.6): the marker must be present and name `variable`, and
 * its recorded ratio must match the measured one. Unlike `checkFloor`,
 * the floor itself is never gated here, since a fill-only color is
 * deliberately allowed to fall below it.
 */
function checkFillOnlyMarker(
  where: string,
  variable: string,
  text: string,
  base: string,
  marker: ExceptionMarker | null,
  reason: string,
  violations: string[],
  reportLines: string[]
): void {
  const ratio = measureRatio(where, variable, text, base, violations);
  if (ratio === null) return;
  if (!marker) {
    violations.push(
      `g9: ${where}: ${variable} is fill-only and must carry a contrast-exception marker (VI "Themes", W2.6)`
    );
    return;
  }
  if (marker.variable !== variable) {
    violations.push(`g9: ${where}: the marker above ${variable} names ${marker.variable} instead`);
  }
  const recorded = Number(marker.ratio);
  if (Math.abs(recorded - ratio) >= RATIO_TOLERANCE) {
    violations.push(
      `g9: ${where}: ${variable}'s marker records ${formatRatio(recorded)}, measured ${formatRatio(ratio)}`
    );
  }
  reportLines.push(`g9: ${where}: ${variable} excluded from the direct-text check (${reason}), marker ${formatRatio(ratio)}`);
}

/** The non-text floor a control's edge must clear against its ground (WCAG 1.4.11). */
const NON_TEXT_FLOOR = 3.0;

/** A two-digit hex byte from a 0-1 sRGB channel, clamped. */
function hexByte(c: number): string {
  return Math.round(Math.min(1, Math.max(0, c)) * 255)
    .toString(16)
    .padStart(2, '0');
}

/** Linear sRGB to OKLab, Björn Ottosson's published matrices (CSS Color 4's conversion appendix). */
function linearToOklab([r, g, b]: [number, number, number]): [number, number, number] {
  const l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b);
  const m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
  const s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
  return [
    0.2104542553 * l + 0.793617785 * m - 0.0040720468 * s,
    1.9779984951 * l - 2.428592205 * m + 0.4505937099 * s,
    0.0259040371 * l + 0.7827717662 * m - 0.808675766 * s
  ];
}

/** OKLab to linear sRGB, the inverse of {@link linearToOklab}. */
function oklabToLinear([L, a, b]: [number, number, number]): [number, number, number] {
  const l = (L + 0.3963377774 * a + 0.2158037573 * b) ** 3;
  const m = (L - 0.1055613458 * a - 0.0638541728 * b) ** 3;
  const s = (L - 0.0894841775 * a - 1.291485548 * b) ** 3;
  return [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.707614701 * s
  ];
}

/**
 * `color-mix(in oklab, <a> <p>%, <b>)` with both operands resolved
 * against `palette`, as the `#rrggbb` a browser paints, or `null` when
 * `value` has another shape or an operand is not a hex color.
 */
function resolveOklabMix(value: string, palette: Map<string, string>): string | null {
  const m = /^color-mix\(\s*in oklab\s*,\s*(.+?)\s+(\d+(?:\.\d+)?)%\s*,\s*(.+?)\s*\)$/.exec(value.trim());
  if (!m) return null;
  const a = parseHex(resolveAgainstPalette(m[1], palette) ?? '');
  const b = parseHex(resolveAgainstPalette(m[3], palette) ?? '');
  if (a === null || b === null) return null;
  const weight = Number(m[2]) / 100;
  const labA = linearToOklab(a.map(srgbToLinearChannel) as [number, number, number]);
  const labB = linearToOklab(b.map(srgbToLinearChannel) as [number, number, number]);
  const mixed = labA.map((v, i) => v * weight + labB[i] * (1 - weight)) as [number, number, number];
  const encode = (c: number) => (c <= 0.0031308 ? 12.92 * c : 1.055 * c ** (1 / 2.4) - 0.055);
  return `#${oklabToLinear(mixed).map(encode).map(hexByte).join('')}`;
}

/**
 * The keyline rows: `--keyline-base` against each base at the
 * non-text floor. A marker covers the pairs under the floor only when
 * each matches its recorded ratio and its reason names that base; a
 * marker with no pair under the floor is stale. Returns the pairs
 * measured.
 */
function checkKeyline(
  composed: string,
  decl: DeclNode,
  palette: Map<string, string>,
  bases: Map<string, string | undefined>,
  violations: string[],
  lines: string[]
): number {
  const resolved = resolveAgainstPalette(decl.value, palette);
  if (resolved === null) {
    violations.push(`g9: ${composed}: --keyline-base's value "${decl.value}" does not resolve against this theme's own --color-* set`);
    return 0;
  }
  let measured = 0;
  let covered = 0;
  for (const [baseName, baseValue] of bases) {
    if (!baseValue) continue;
    const where = `${composed} (${baseName})`;
    const ratio = measureRatio(where, '--keyline-base', resolved, baseValue, violations);
    measured++;
    if (ratio === null) continue;
    if (ratio >= NON_TEXT_FLOOR) {
      lines.push(`g9: ${where}: --keyline-base = ${formatRatio(ratio)}`);
      continue;
    }
    const marker = decl.marker;
    const matches =
      marker !== null &&
      marker.variable === '--keyline-base' &&
      Math.abs(Number(marker.ratio) - ratio) < RATIO_TOLERANCE &&
      marker.reason.includes(baseName.replace('--color-', ''));
    if (!matches) {
      violations.push(
        `g9: ${where}: --keyline-base measures ${formatRatio(ratio)} against ${NON_TEXT_FLOOR}:1 and no marker records it for ${baseName}`
      );
      continue;
    }
    covered++;
    lines.push(`g9: ${where}: --keyline-base = ${formatRatio(ratio)} (exception: ${marker.reason})`);
  }
  if (decl.marker !== null && covered === 0) {
    violations.push(`g9: ${composed}: --keyline-base's exception marker is stale, no base measures under ${NON_TEXT_FLOOR}:1 at its ratio`);
  }
  return measured;
}

function readSurfaceRules(violations: string[]): RuleNode[] {
  const text = readFileSync(`${THEMES_DIR}/surface.css`, 'utf8');
  return singleLayerBaseChildren('surface.css', text, violations).filter((n): n is RuleNode => n.kind === 'rule');
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

/**
 * The closed list's size, per composed theme (DL R2.6, R-5ciii3): 8
 * `-content` pairs + 3 `--color-base-content` rows + 7 direct-text
 * colors (the eighth, per theme, is fill-only and carries a marker
 * instead) + 5 inks x 3 bases + 3 keyline rows + 1 selection row = 37,
 * times 2 composed themes = 74. Pinned to the literal rather than
 * derived from `CONTENT_PAIRS`, `BASES`, `DL_R2_6_DIRECT_TEXT_COLORS`,
 * `VI_FILL_ONLY`, or `INK_NAMES`, so a deletion from any of them is
 * caught instead of silently shrinking this count to match.
 */
const EXPECTED_MEASURED_PAIRS = 74;

export interface Report {
  violations: string[];
  /** One line per measured pair, one worst-case line per ink, and one line per fill-only exclusion. */
  lines: string[];
}

export function run(): Report {
  const violations: string[] = [];
  const lines: string[] = [];

  for (const name of DL_R2_6_DIRECT_TEXT_COLORS) {
    if (!DIRECT_TEXT_COLORS.includes(name)) {
      violations.push(`g9: DIRECT_TEXT_COLORS is missing ${name}, which DL R2.6 names as direct text`);
    }
  }
  for (const name of DIRECT_TEXT_COLORS) {
    if (!DL_R2_6_DIRECT_TEXT_COLORS.includes(name)) {
      violations.push(`g9: DIRECT_TEXT_COLORS names ${name}, which DL R2.6's closed list does not`);
    }
  }
  for (const pinned of VI_FILL_ONLY) {
    if (!FILL_ONLY_EXCLUSIONS.some((e) => e.composed === pinned.composed && e.name === pinned.name)) {
      violations.push(
        `g9: FILL_ONLY_EXCLUSIONS is missing ${pinned.composed} ${pinned.name}, which VI "Themes" names as fill-only`
      );
    }
  }
  for (const exclusion of FILL_ONLY_EXCLUSIONS) {
    if (!VI_FILL_ONLY.some((p) => p.composed === exclusion.composed && p.name === exclusion.name)) {
      violations.push(
        `g9: FILL_ONLY_EXCLUSIONS names ${exclusion.composed} ${exclusion.name}, which VI "Themes" does not name as fill-only`
      );
    }
  }

  const inkRules = readInkRules(violations);
  const surfaceRules = readSurfaceRules(violations);
  let measuredPairs = 0;

  for (const composed of COMPOSED_NAMES) {
    const declarations = readPaletteBlock(PALETTE_FILES[composed], composed);
    if (!declarations) {
      violations.push(`g9: ${composed}: no composed theme found, an empty iteration is not a pass`);
      continue;
    }
    const palette = declMap(declarations);
    const bases = new Map(BASES.map((name) => [name, palette.get(name)]));
    for (const [name, value] of bases) {
      if (!value) violations.push(`g9: ${composed}: ${name} is missing, contrast cannot be measured`);
    }
    const base100 = bases.get('--color-base-100');
    if (!base100) continue;

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
      checkFloor(composed, contentName, contentDecl.value, fillValue, contentDecl.marker, false, violations, lines);
      measuredPairs++;
    }

    const baseContentDecl = declarations.find((d) => d.name === '--color-base-content');
    if (!baseContentDecl) {
      violations.push(`g9: ${composed}: --color-base-content is missing, contrast cannot be measured`);
    } else {
      for (const [baseName, baseValue] of bases) {
        if (!baseValue) continue;
        // base-content is always text too: same restriction as the -content pairs above.
        checkFloor(
          `${composed} (${baseName})`,
          '--color-base-content',
          baseContentDecl.value,
          baseValue,
          baseContentDecl.marker,
          false,
          violations,
          lines
        );
        measuredPairs++;
      }
    }

    for (const name of DIRECT_TEXT_COLORS) {
      const decl = declarations.find((d) => d.name === name);
      if (!decl) {
        violations.push(`g9: ${composed}: ${name} is missing, contrast cannot be measured`);
        continue;
      }
      if (isFillOnly(composed, name)) {
        const exclusion = FILL_ONLY_EXCLUSIONS.find((e) => e.composed === composed && e.name === name)!;
        checkFillOnlyMarker(composed, name, decl.value, base100, decl.marker, exclusion.reason, violations, lines);
        continue;
      }
      checkFloor(composed, name, decl.value, base100, decl.marker, true, violations, lines);
      measuredPairs++;
    }

    const inkRule = inkRules.find((r) => selectorHasDataTheme(r.selector, composed));
    const inkDecls = inkRule ? declMap(inkRule.declarations) : new Map<string, string>();
    for (const inkName of INK_NAMES) {
      const rawValue = inkDecls.get(inkName);
      if (rawValue === undefined) {
        violations.push(`g9: ${composed}: ${inkName} is missing from inks.css, contrast cannot be measured`);
        continue;
      }
      const inkDecl = inkRule?.declarations.find((d) => d.name === inkName) ?? null;
      if (inkDecl?.marker) {
        violations.push(`g9: ${composed}: ${inkName} carries an exception marker; inks admit none`);
        continue;
      }
      const resolved = resolveAgainstPalette(rawValue, palette);
      if (resolved === null) {
        violations.push(
          `g9: ${composed}: ${inkName}'s value "${rawValue}" does not resolve against this theme's own --color-* set`
        );
        continue;
      }
      const pinned = PINNED_INKS[composed]?.[inkName];
      if (pinned !== undefined && resolved.toLowerCase() !== pinned.toLowerCase()) {
        violations.push(
          `g9: ${composed}: ${inkName} resolves to "${resolved}", the visual identity spec's ink table pins "${pinned}"`
        );
      }
      let worst: number | null = null;
      for (const [baseName, baseValue] of bases) {
        if (!baseValue) continue;
        const ratio = checkFloor(`${composed} (${baseName})`, inkName, resolved, baseValue, null, false, violations, lines);
        measuredPairs++;
        if (ratio !== null && (worst === null || ratio < worst)) worst = ratio;
      }
      if (worst !== null) {
        lines.push(`g9: ${composed}: ${inkName} worst case ${formatRatio(worst)}`);
      }
    }

    const surfaceRule = surfaceRules.find((r) => selectorHasDataTheme(r.selector, composed));
    const keyline = surfaceRule?.declarations.find((d) => d.name === '--keyline-base');
    if (!keyline) {
      violations.push(`g9: ${composed}: --keyline-base is missing from surface.css, contrast cannot be measured`);
    } else {
      measuredPairs += checkKeyline(composed, keyline, palette, bases, violations, lines);
    }
    const selection = surfaceRule?.declarations.find((d) => d.name === '--selection-ground');
    const selectionGround = selection ? resolveOklabMix(selection.value, palette) : null;
    if (selectionGround === null) {
      violations.push(
        `g9: ${composed}: --selection-ground ${selection ? `"${selection.value}" is not an OKLab color-mix of two palette colors` : 'is missing from surface.css'}`
      );
    } else if (baseContentDecl) {
      checkFloor(
        `${composed} (--selection-ground ${selectionGround})`,
        '--color-base-content',
        baseContentDecl.value,
        selectionGround,
        null,
        false,
        violations,
        lines
      );
      measuredPairs++;
    }
  }

  if (measuredPairs !== EXPECTED_MEASURED_PAIRS) {
    violations.push(
      `g9: measured ${measuredPairs} pair(s), want ${EXPECTED_MEASURED_PAIRS}: a missing declaration would otherwise pass silently`
    );
  }

  return { violations, lines };
}

function main(): void {
  const { violations, lines } = run();
  if (violations.length > 0) {
    fail(`g9: ${violations.length} finding(s):\n${violations.join('\n')}`);
  }
  process.stdout.write(`${lines.join('\n')}\n`);
  process.stdout.write('g9: passed, every pairing clears its floor or carries a live exception.\n');
}

if (isMainModule(import.meta.url)) {
  main();
}
