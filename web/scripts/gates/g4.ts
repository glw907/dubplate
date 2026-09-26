/**
 * G4: font delivery. Reads `web/static/fonts/` and
 * `web/src/lib/themes/fonts.css` and asserts:
 *
 * - Every file directly under `web/static/fonts/`, skipping
 *   `licenses/` and `FONTS.md` (the license texts and the manifest),
 *   ends in `.woff2` and no other font format is present.
 * - The file count stays within the four-woff2-file budget.
 * - Every `@font-face` rule in `fonts.css` sets `font-display: swap`,
 *   `font-weight`, and `font-style` explicitly, and its `src` names
 *   one `url()` in `format('woff2')` and nothing else.
 * - Every file's SHA-256 equals its `FONTS.md` row, and every
 *   `FONTS.md` row names a file that exists.
 */
import { createHash } from 'node:crypto';
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { fail, isMainModule, walk } from './lib.ts';

const FONTS_DIR = 'static/fonts';
const FONT_FACE_FILE = 'src/lib/themes/fonts.css';
const MANIFEST_FILE = 'static/fonts/FONTS.md';
const MAX_WOFF2_FILES = 4;

const LICENSES_PREFIX = 'licenses/';

/** Every font file under `FONTS_DIR`, excluding `licenses/` and `FONTS.md`. */
export function fontFiles(root: string): string[] {
  return walk(join(root, FONTS_DIR))
    .filter((rel) => !rel.startsWith(LICENSES_PREFIX))
    .filter((rel) => rel !== 'FONTS.md');
}

/** A file extension outside `.woff2`, read from a font file's own name. */
export function nonWoff2Files(files: string[]): string[] {
  return files.filter((rel) => !rel.endsWith('.woff2'));
}

export interface FontFaceRule {
  /** The rule's own text, for a violation message. */
  text: string;
  /** Each declared descriptor, its name lowercased, to its trimmed value. */
  descriptors: Map<string, string>;
}

/** Every `@font-face { ... }` block in css, with its descriptors. */
export function parseFontFaceRules(css: string): FontFaceRule[] {
  const rules: FontFaceRule[] = [];
  const re = /@font-face\s*\{([^}]*)\}/g;
  let m: RegExpExecArray | null;
  while ((m = re.exec(css))) {
    const descriptors = new Map<string, string>();
    for (const decl of m[1].split(';')) {
      const colon = decl.indexOf(':');
      if (colon === -1) continue;
      descriptors.set(decl.slice(0, colon).trim().toLowerCase(), decl.slice(colon + 1).trim());
    }
    rules.push({ text: m[0], descriptors });
  }
  return rules;
}

/** A `src` value naming exactly one `url()`, in the woff2 format and no other. */
const SINGLE_WOFF2_SRC = /^url\(\s*(['"])[^'"()]+\1\s*\)\s+format\(\s*(['"])woff2\2\s*\)$/;

/** The problems with one rule's descriptors, each a phrase naming what is wrong. */
export function fontFaceProblems(rule: FontFaceRule): string[] {
  const problems: string[] = [];
  for (const descriptor of ['font-display', 'font-weight', 'font-style', 'src']) {
    if (!rule.descriptors.has(descriptor)) problems.push(`is missing ${descriptor}`);
  }
  const display = rule.descriptors.get('font-display');
  if (display !== undefined && display !== 'swap') problems.push(`sets font-display: ${display}, want swap`);
  const src = rule.descriptors.get('src');
  if (src !== undefined && !SINGLE_WOFF2_SRC.test(src)) {
    problems.push(`sets src: ${src}, want one url() in format('woff2')`);
  }
  return problems;
}

export interface ManifestRow {
  file: string;
  sha256: string;
}

/**
 * `FONTS.md`'s `- File:` and `- SHA-256:` lines, paired by the fixed
 * order the manifest writes them in: a `File:` line, then that face's
 * `SHA-256:` line before the next `File:` line.
 */
export function parseManifest(text: string): ManifestRow[] {
  const rows: ManifestRow[] = [];
  let currentFile: string | null = null;
  for (const line of text.split('\n')) {
    const fileMatch = /^-\s*File:\s*`([^`]+)`/.exec(line);
    if (fileMatch) {
      currentFile = fileMatch[1];
      continue;
    }
    const shaMatch = /^-\s*SHA-256:\s*`([0-9a-f]{64})`/.exec(line);
    if (shaMatch && currentFile) {
      rows.push({ file: currentFile, sha256: shaMatch[1] });
      currentFile = null;
    }
  }
  return rows;
}

function sha256Hex(data: Buffer): string {
  return createHash('sha256').update(data).digest('hex');
}

function main() {
  const violations: string[] = [];

  const files = fontFiles('.');
  if (files.length === 0) {
    fail(`g4: ${FONTS_DIR} is absent or empty; an empty iteration is not a pass.`);
  }
  const badFormat = nonWoff2Files(files);
  for (const rel of badFormat) {
    violations.push(`g4: ${join(FONTS_DIR, rel)}: not a .woff2 file`);
  }
  const woff2Count = files.length - badFormat.length;
  if (woff2Count > MAX_WOFF2_FILES) {
    violations.push(`g4: ${woff2Count} .woff2 files under ${FONTS_DIR}, want at most ${MAX_WOFF2_FILES}`);
  }

  let css: string;
  try {
    css = readFileSync(FONT_FACE_FILE, 'utf8');
  } catch {
    fail(`g4: ${FONT_FACE_FILE}: no such file`);
  }
  const rules = parseFontFaceRules(css);
  if (rules.length === 0) {
    violations.push(`g4: ${FONT_FACE_FILE}: no @font-face rule found`);
  }
  for (const rule of rules) {
    for (const problem of fontFaceProblems(rule)) {
      violations.push(`g4: a @font-face rule ${problem}: ${rule.text.replace(/\s+/g, ' ').trim()}`);
    }
  }

  let manifestText: string;
  try {
    manifestText = readFileSync(MANIFEST_FILE, 'utf8');
  } catch {
    fail(`g4: ${MANIFEST_FILE}: no such file`);
  }
  const manifestRows = parseManifest(manifestText);
  if (manifestRows.length === 0) {
    violations.push(`g4: ${MANIFEST_FILE}: no manifest row found`);
  }
  for (const row of manifestRows) {
    const path = join(FONTS_DIR, row.file);
    let data: Buffer;
    try {
      data = readFileSync(path);
    } catch {
      violations.push(`g4: ${MANIFEST_FILE}: row for ${row.file} names a file that does not exist`);
      continue;
    }
    const actual = sha256Hex(data);
    if (actual !== row.sha256) {
      violations.push(
        `g4: ${row.file}: SHA-256 is ${actual}, manifest row says ${row.sha256}`
      );
    }
  }
  const manifestFiles = new Set(manifestRows.map((r) => r.file));
  for (const rel of files) {
    if (badFormat.includes(rel)) continue;
    if (!manifestFiles.has(rel)) {
      violations.push(`g4: ${join(FONTS_DIR, rel)}: no ${MANIFEST_FILE} row names it`);
    }
  }

  if (violations.length > 0) {
    fail(`g4: ${violations.length} finding(s):\n${violations.join('\n')}`);
  }
  process.stdout.write(
    `g4: passed, ${woff2Count} .woff2 file(s) under ${MAX_WOFF2_FILES}, ${rules.length} @font-face rule(s) complete, ${manifestRows.length} manifest row(s) verified.\n`
  );
}

if (isMainModule(import.meta.url)) {
  main();
}
