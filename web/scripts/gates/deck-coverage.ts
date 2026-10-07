/**
 * Deck coverage: the string deck's own checks.
 *
 * 1. Every outcome value `internal/auth/handlers.go` and `admin.go`
 *    answer with has a line somewhere in the deck's outcome
 *    dictionaries, read from the Go source itself so a renamed or
 *    added outcome constant surfaces here rather than silently
 *    shipping with no copy.
 * 2. No deck string (nor a title) carries `<`, `>`, `&`, a straight
 *    apostrophe (U+0027), or a straight double quote (U+0022).
 * 3. No sign-out line claims "everywhere" or "revoked": this device's
 *    session ends, not every session.
 * 4. No `{@html` anywhere in `web/src`: every deck consumer renders
 *    by text interpolation.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { stripComments } from './g1.ts';
import { fail, isMainModule, walk } from './lib.ts';
import * as deck from '../../src/lib/copy/deck.ts';
import { AUTH_OUTCOMES, ROSTER_OUTCOMES, SIGN_OUT_CONTROL, TITLES, unknownAddressLine } from '../../src/lib/copy/deck.ts';
import { pageTitle } from '../../src/lib/copy/title.ts';

const SRC_ROOT = 'src';
const HANDLERS_PATH = join('..', 'internal', 'auth', 'handlers.go');
const ADMIN_PATH = join('..', 'internal', 'auth', 'admin.go');

/** `internal/auth`'s own `outcomeX = "value"` declaration shape. */
const GO_OUTCOME_CONST = /\boutcome[A-Za-z]+\s*=\s*"([^"]*)"/g;

/** Every outcome value declared in the Go source at `path`. */
function parseGoOutcomes(path: string): string[] {
  const text = readFileSync(path, 'utf8');
  return [...text.matchAll(GO_OUTCOME_CONST)].map((m) => m[1]);
}

interface Leaf {
  key: string;
  value: string;
}

/** The nine page-state keys `deck.ts` declares on `TITLES`. */
const REQUIRED_TITLES = [
  'login',
  'loginSent',
  'loginLink',
  'loginLinkSignedIn',
  'loginLinkRefused',
  'members',
  'signedOut',
  'notFound',
  'error'
];

/**
 * Every `REQUIRED_TITLES` key resolves to a non-empty `TITLES` line, and
 * the nil-contact line `unknownAddressLine(null)` renders. `deckLeaves`
 * only ever reads `TITLES.login` through `pageTitle`, so a deleted title
 * for any other page state would otherwise pass silently. The reverse
 * check catches the opposite drift: a `TITLES` key `REQUIRED_TITLES`
 * does not name, which would otherwise ship uncovered.
 */
function checkRequiredTitles(): void {
  for (const key of REQUIRED_TITLES) {
    const value = (TITLES as Record<string, string | undefined>)[key];
    if (!value) {
      fail(`deck-coverage: TITLES has no line for "${key}"`);
    }
  }
  for (const key of Object.keys(TITLES)) {
    if (!REQUIRED_TITLES.includes(key)) {
      fail(`deck-coverage: TITLES."${key}" is not in REQUIRED_TITLES; add it so the two lists cannot drift`);
    }
  }
  if (!unknownAddressLine(null)) {
    fail('deck-coverage: unknownAddressLine(null) has no line');
  }
}

/**
 * How to call each function `deck.ts` exports to reach a representative
 * line: a function's own arguments cannot be inferred from a walk, so
 * this is the one hand-kept list the walk cannot replace. `deck.ts`'s
 * own promise (every export is covered) still holds for every plain
 * string and dictionary, which {@link deckLeaves} walks generically; a
 * function `deck.ts` adds with no entry here fails instead of walking
 * past uncovered.
 */
const FUNCTION_SAMPLES: Record<string, () => string> = {
  unknownAddressLine: () => unknownAddressLine({ displayName: 'Root Admin', address: 'admin@example.test' }),
  memberCountLine: () => deck.memberCountLine(2),
  addedMemberLine: () => deck.addedMemberLine('New Member')
};

/**
 * A second `FUNCTION_SAMPLES` argument to check for each function
 * named here, beyond the one {@link deckLeaves} walks: `memberCountLine`
 * has a singular branch (`1 member`) a lone sample at `2` never
 * exercises, so its plural-aware wording could drift with no leaf ever
 * proving the singular line still reads right.
 */
const EXTRA_FUNCTION_SAMPLES: Record<string, () => string> = {
  memberCountLine: () => deck.memberCountLine(1)
};

/**
 * Every plain-text export, every dictionary export flattened one
 * level, and every function export's sample, walked from `deck.ts`'s
 * own module namespace rather than a hand-kept list of its names,
 * so an added string or dictionary is covered without a second edit
 * here.
 */
function deckLeaves(): Leaf[] {
  const leaves: Leaf[] = [];
  for (const [name, value] of Object.entries(deck)) {
    if (typeof value === 'string') {
      leaves.push({ key: name, value });
    } else if (typeof value === 'function') {
      const sample = FUNCTION_SAMPLES[name];
      if (!sample) {
        fail(`deck-coverage: deck.ts exports a function "${name}" with no FUNCTION_SAMPLES entry to walk it.`);
      } else {
        leaves.push({ key: `${name}(sample)`, value: sample() });
      }
    } else if (typeof value === 'object' && value !== null) {
      for (const [key, leaf] of Object.entries(value as Record<string, unknown>)) {
        if (typeof leaf !== 'string') {
          fail(`deck-coverage: deck.ts export "${name}.${key}" is not a string; every dictionary export must be flat.`);
        }
        leaves.push({ key: `${name}.${key}`, value: leaf });
      }
    }
  }
  leaves.push({ key: 'unknownAddressLine(null)', value: unknownAddressLine(null) });
  leaves.push({ key: 'pageTitle', value: pageTitle(TITLES.login, 'Example Group') });
  for (const [name, sample] of Object.entries(EXTRA_FUNCTION_SAMPLES)) {
    leaves.push({ key: `${name}(extraSample)`, value: sample() });
  }
  return leaves;
}

/**
 * `pageTitle` with an empty site name returns the state alone, with no
 * dangling ` · ` separator (`title.ts`'s empty-meta case).
 */
function checkPageTitleEmptySiteName(): void {
  const got = pageTitle(TITLES.login, '');
  if (got !== TITLES.login) {
    fail(`deck-coverage: pageTitle(state, '') = ${JSON.stringify(got)}, want ${JSON.stringify(TITLES.login)}`);
  }
}

/** `internal/auth`'s combined outcome set, or a failure naming the parser or the count. */
function goOutcomes(): Set<string> {
  const handlers = parseGoOutcomes(HANDLERS_PATH);
  const admin = parseGoOutcomes(ADMIN_PATH);
  const parsed = new Set([...handlers, ...admin]);
  for (const must of ['signed_out', 'admin_not_grantable', 'duplicate_username']) {
    if (!parsed.has(must)) {
      fail(
        `deck-coverage: the parsed outcome set is missing "${must}"; the constant parser's pattern may be broken.`
      );
    }
  }
  if (parsed.size < 20) {
    fail(
      `deck-coverage: parsed only ${parsed.size} outcome constant(s) from ${HANDLERS_PATH} and ${ADMIN_PATH}, fewer than the 20 internal/auth declares.`
    );
  }
  return parsed;
}

/** Every outcome key the deck's dictionaries cover. */
function deckOutcomes(): Set<string> {
  return new Set([...Object.keys(AUTH_OUTCOMES), ...Object.keys(ROSTER_OUTCOMES)]);
}

function checkOutcomeCoverage(): void {
  const parsed = goOutcomes();
  const covered = deckOutcomes();
  const missing = [...parsed].filter((v) => !covered.has(v));
  if (missing.length > 0) {
    fail(`deck-coverage: ${missing.length} outcome(s) declared in internal/auth with no deck line: ${missing.join(', ')}`);
  }
}

const FORBIDDEN_CHAR = /[<>&'"]/;

function checkForbiddenChars(leaves: Leaf[]): void {
  for (const { key, value } of leaves) {
    const m = value.match(FORBIDDEN_CHAR);
    if (m) {
      fail(`deck-coverage: deck value "${key}" carries a forbidden character (${m[0]}): ${value}`);
    }
  }
}

function checkSignOutWording(): void {
  const lines = [SIGN_OUT_CONTROL, AUTH_OUTCOMES.signed_out];
  for (const line of lines) {
    const low = line.toLowerCase();
    if (low.includes('everywhere') || low.includes('revoked')) {
      fail(`deck-coverage: a sign-out line claims revocation: "${line}"`);
    }
  }
}

const AT_HTML = /\{@html/;

function checkNoAtHtml(files: string[]): void {
  const violations: string[] = [];
  for (const rel of files) {
    const text = stripComments(rel, readFileSync(join(SRC_ROOT, rel), 'utf8'));
    text.split('\n').forEach((line, i) => {
      if (AT_HTML.test(line)) violations.push(`${rel}:${i + 1}`);
    });
  }
  if (violations.length > 0) {
    fail(`deck-coverage: {@html} found; every deck consumer must render by text interpolation: ${violations.join(', ')}`);
  }
}

function main(): void {
  const files = walk(SRC_ROOT);
  if (files.length === 0) {
    fail(`deck-coverage: ${SRC_ROOT} holds no file; an empty iteration is not a pass.`);
  }

  checkOutcomeCoverage();
  checkRequiredTitles();

  const leaves = deckLeaves();
  if (leaves.length === 0) {
    fail('deck-coverage: the deck exports no string; an empty iteration is not a pass.');
  }
  checkForbiddenChars(leaves);
  checkSignOutWording();
  checkNoAtHtml(files);
  checkPageTitleEmptySiteName();

  process.stdout.write(
    `deck-coverage: passed, ${leaves.length} deck value(s) checked, ${files.length} src file(s) scanned for {@html}.\n`
  );
}

if (isMainModule(import.meta.url)) {
  main();
}
