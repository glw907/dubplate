/**
 * The AI-tells rule: a TypeScript port of cairn-cms's
 * `scripts/checks/check-admin-prose.mjs` (MIT, 285 lines; the blocking
 * lists ported verbatim from prose-guard), scanning every user-facing
 * string in `web/src` markup and literals for the same lexical and
 * structural tells: a Svelte component's text nodes and copy-bearing
 * attributes, and a `.ts` module's string and template literals.
 *
 * This gate enforces no house register; a human prose read against
 * the copy register does that separately. `--list` dumps every
 * extracted string for that read.
 *
 *   node scripts/gates/ai-tells.ts          scan and fail on a hit
 *   node scripts/gates/ai-tells.ts --list   print every extracted string, grouped by file
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { fail, isMainModule, walk } from './lib.ts';

const SRC_ROOT = 'src';

const EM_DASH = '—'; // The en dash (ranges) is never flagged.

const BANNED_PHRASES = [
  "it's worth noting",
  'when it comes to',
  'dive into',
  'delve',
  "let's explore",
  'at the end of the day',
  'game-changer',
  'game changer',
  'state-of-the-art',
  'look no further',
  "in today's world",
  'to be honest',
  'to be frank',
  'the honest answer is',
  'in the realm of',
  'in the world of'
];
const BANNED_OPENERS = ['moreover', 'additionally', 'furthermore', 'in conclusion', 'needless to say', 'certainly', 'it should be noted'];
const FILLER_WORDS = ['genuinely', 'honestly'];
const SLOP_WORDS = ['tapestry', 'multifaceted', 'testament', 'seamless'];
const MARKETING_WORDS = ['empower', 'streamline', 'supercharge', 'turbocharge', 'revolutionize', 'effortless', 'effortlessly', 'plethora', 'myriad'];
const JUDGMENT_WORDS = [
  'robust',
  'leverage',
  'comprehensive',
  'dedicated',
  'curated',
  'tailored',
  'foster',
  'elevate',
  'transformative',
  'pivotal',
  'thriving',
  'meticulous',
  'nuanced',
  'embark',
  'harness',
  'bolster',
  'groundbreaking',
  'cutting-edge',
  'innovative',
  'foundational'
];
// The general tier scans the slop and judgment word lists alongside the shared rules above.
const TIER_WORDS = [...SLOP_WORDS, ...JUDGMENT_WORDS];

const SENTENCE_SPLIT = /(?<=[.!?])\s+/;

/** The em dashes in `text`. (The en dash for ranges is never flagged.) */
function emDashCount(text: string): number {
  return (text.match(new RegExp(EM_DASH, 'g')) ?? []).length;
}

interface Issue {
  kind: string;
  hint: string;
}

// (kind, regex, hint). High-precision; ported from prose-guard's STRUCTURAL.
const STRUCTURAL: Array<[string, RegExp, string]> = [
  [
    'negative antithesis',
    /\b(it|this|that)'?s?\s+(not|isn'?t)\b[^.!?;]{3,60}[,;]\s+(it|this|that|but)'?s?\b/i,
    "Explicit 'not X, it's Y' antithesis. Prefer an implicit contrast or a plain statement."
  ],
  ['not just … but', /\bnot\s+just\b[^.!?]{5,60}\bbut\s+(also\s+)?/i, 'The escalation frame reads as AI. State the point directly.'],
  [
    'setup-colon payoff',
    /\b(the\s+)?(point|takeaway|truth|reality|bottom line|catch|kicker)\s*:\s+[A-Z]/i,
    "The 'The point:' setup-payoff is a tell. Fold it into the sentence."
  ],
  ['copula dodge (serves/stands as)', /\b(serves?|stands?|acts?|functions?)\s+as\s+a\b/i, "Use 'is' instead of 'serves as a'."],
  [
    'participial wind-up',
    /^\s*(building on|recognizing|leveraging|drawing on|having established|taking)\b[^,]{0,40},\s/im,
    'Start with the subject, not a participial bridge.'
  ]
];

/** Run the blocking lexical and structural rules over one extracted string. */
export function scan(line: string): Issue[] {
  const issues: Issue[] = [];
  // General/marketing prose: an occasional em dash is human; flag a spray, and flag a
  // clause followed by a short tacked-on fragment after a dash.
  if (emDashCount(line) > 2) {
    issues.push({ kind: 'em-dash spray', hint: 'Keep at most one interruption (a pair) per line.' });
  }
  for (const sent of line.split(SENTENCE_SPLIT)) {
    if (emDashCount(sent) === 1) {
      let after = sent.split(EM_DASH)[1] ?? '';
      after = after
        .replace(/[*_`)\]]+$/, '')
        .trim()
        .replace(/[.!?]+$/, '')
        .trim();
      const words = after.split(/\s+/).filter(Boolean).length;
      if (words >= 1 && words <= 6) {
        issues.push({
          kind: 'em-dash appendage',
          hint: 'A clause plus a tacked-on fragment after a dash is a tell. Use a period, comma, or colon.'
        });
      }
    }
  }
  const low = line.toLowerCase();
  if (low.includes('not only') && low.includes('but also')) {
    issues.push({ kind: 'not only … but also', hint: 'Reword without the correlative pair.' });
  }
  for (const phrase of BANNED_PHRASES) {
    if (low.includes(phrase)) issues.push({ kind: `banned phrase: ${phrase}`, hint: 'Reword in a human voice.' });
  }
  for (const filler of FILLER_WORDS) {
    if (new RegExp(`\\b${filler}\\b`).test(low)) issues.push({ kind: `filler word: ${filler}`, hint: 'Drop the filler intensifier.' });
  }
  for (const word of MARKETING_WORDS) {
    if (new RegExp(`\\b${word}\\w*\\b`).test(low)) issues.push({ kind: `marketing word: ${word}`, hint: 'Cut the marketing verb; say what it does plainly.' });
  }
  for (const opener of BANNED_OPENERS) {
    for (const sent of low.split(SENTENCE_SPLIT)) {
      const head = sent.replace(/^[\s\-*>#]+/, '');
      if (head.startsWith(opener)) issues.push({ kind: `banned opener: ${opener}`, hint: 'Start with a subject, not a connector.' });
    }
  }
  for (const word of TIER_WORDS) {
    if (new RegExp(`\\b${word.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')}\\w*\\b`).test(low)) {
      issues.push({ kind: `banned word: ${word}`, hint: 'Reword in a human voice.' });
    }
  }
  for (const [kind, rx, hint] of STRUCTURAL) {
    if (rx.test(line)) issues.push({ kind, hint });
  }
  return issues;
}

// --- Extracting the user-facing copy. ---

// Attributes whose values are human-readable copy (not class names, ids, or bindings).
const COPY_ATTRS = ['placeholder', 'aria-label', 'title', 'alt'];

/** Keep only the extracted strings carrying a letter, de-duplicated. */
function lettersOnly(found: string[]): string[] {
  return [...new Set(found.filter((s) => /[A-Za-z]/.test(s)))];
}

/**
 * Pull the user-facing strings out of one `.svelte` component's
 * source: the markup text nodes plus the static values of the
 * copy-bearing attributes. Script and style blocks, comments, and
 * `{...}` expressions are removed first so only literal author-visible
 * copy remains.
 */
export function extractSvelteCopy(src: string): string[] {
  const stripped = src
    .replace(/<!--[\s\S]*?-->/g, '')
    .replace(/<script[\s\S]*?<\/script>/gi, '')
    .replace(/<style[\s\S]*?<\/style>/gi, '')
    .replace(/\{[^{}]*\}/g, ' '); // drop mustache interpolations, keeping surrounding text

  const found: string[] = [];
  for (const attr of COPY_ATTRS) {
    const rx = new RegExp(`\\b${attr}\\s*=\\s*"([^"]*)"`, 'gi');
    for (const m of stripped.matchAll(rx)) {
      const v = m[1].trim();
      if (v) found.push(v);
    }
  }
  for (const m of stripped.matchAll(/>([^<>]+)</g)) {
    const v = m[1].replace(/\s+/g, ' ').trim();
    if (v) found.push(v);
  }
  return lettersOnly(found);
}

/**
 * Pull the user-facing strings out of a `.ts` module: every
 * single-quoted, double-quoted, and backtick literal. Line and block
 * comments are removed first so a TSDoc gloss never reads as copy.
 */
export function extractTsCopy(src: string): string[] {
  const stripped = src.replace(/\/\*[\s\S]*?\*\//g, ' ').replace(/\/\/[^\n]*/g, ' ');

  const found: string[] = [];
  for (const rx of [/'([^'\\]*)'/g, /"([^"\\]*)"/g, /`([^`\\]*)`/g]) {
    for (const m of stripped.matchAll(rx)) {
      const v = m[1].replace(/\s+/g, ' ').trim();
      if (v) found.push(v);
    }
  }
  return lettersOnly(found);
}

interface ScannedFile {
  path: string;
  copy: string[];
}

/** Every `.svelte` and `.ts` file under `web/src`, with its extracted copy. */
function collect(): ScannedFile[] {
  const files = walk(SRC_ROOT);
  const out: ScannedFile[] = [];
  for (const rel of files) {
    if (rel.endsWith('.svelte')) {
      out.push({ path: rel, copy: extractSvelteCopy(readFileSync(join(SRC_ROOT, rel), 'utf8')) });
    } else if (rel.endsWith('.ts')) {
      out.push({ path: rel, copy: extractTsCopy(readFileSync(join(SRC_ROOT, rel), 'utf8')) });
    }
  }
  return out;
}

/** Print every extracted string, grouped by file. Backs the `--list` flag. */
function listCopy(files: ScannedFile[]): void {
  for (const f of files) {
    if (f.copy.length === 0) continue;
    process.stdout.write(`\n${f.path}\n`);
    for (const s of f.copy) process.stdout.write(`  ${s}\n`);
  }
}

/** Scan every file, report each tell, and fail on any hit. */
function scanCopy(files: ScannedFile[]): void {
  let hits = 0;
  let scanned = 0;
  for (const f of files) {
    for (const s of f.copy) {
      scanned += 1;
      for (const issue of scan(s)) {
        hits += 1;
        process.stderr.write(`${f.path}: [${issue.kind}]\n  "${s}"\n  ${issue.hint}\n`);
      }
    }
  }
  if (scanned === 0) {
    fail('ai-tells: no user-facing string found under web/src; an empty scan is not a pass.');
  }
  if (hits > 0) {
    fail(
      `ai-tells: ${hits} tell(s) found across ${scanned} scanned string(s). Run \`node scripts/gates/ai-tells.ts --list\` to read every scanned string.`
    );
  }
  process.stdout.write(`ai-tells: passed, ${scanned} scanned string(s) across ${files.length} file(s).\n`);
}

function main(): void {
  const files = collect();
  if (process.argv.includes('--list')) listCopy(files);
  else scanCopy(files);
}

if (isMainModule(import.meta.url)) {
  main();
}
