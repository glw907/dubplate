/**
 * Deck coverage: the string deck's own checks.
 *
 * 1. Every outcome value `internal/auth/handlers.go` and `admin.go`
 *    answer with has a line somewhere in the deck's outcome
 *    dictionaries, read from the Go source itself so a renamed or
 *    added outcome constant surfaces here rather than silently
 *    shipping with no copy.
 * 2. No deck string (nor a title) carries `<`, `>`, `&`, or a straight
 *    apostrophe (U+0027).
 * 3. No sign-out line claims "everywhere" or "revoked": this device's
 *    session ends, not every session.
 * 4. No `{@html` anywhere in `web/src`: every deck consumer renders
 *    by text interpolation.
 */
import { readFileSync } from 'node:fs';
import { join } from 'node:path';
import { stripComments } from './g1.ts';
import { fail, isMainModule, walk } from './lib.ts';
import {
  ADD_MEMBER_HINTS,
  ADD_MEMBER_LABELS,
  AUTH_OUTCOMES,
  CROSS_ORIGIN_REFUSED,
  ERROR_PAGE,
  INTERSTITIAL_CONTROL,
  INTERSTITIAL_STATEMENT,
  LOGIN_SUBLINE,
  NETWORK_FAILURE,
  ROSTER_EMPTY_STATE,
  ROSTER_OUTCOMES,
  SENT_ESCAPE_HATCH,
  SENT_MOMENT_TEXT,
  SIGN_OUT_CONTROL,
  THEME_TOGGLE_LABEL,
  TITLES,
  memberCountLine,
  unknownAddressLine
} from '../../src/lib/copy/deck.ts';
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

/** The seven page-state keys `deck.ts` declares on `TITLES`. */
const REQUIRED_TITLES = [
  'login',
  'loginSent',
  'loginLink',
  'loginLinkSignedIn',
  'loginLinkRefused',
  'members',
  'signedOut'
];

/**
 * Every `REQUIRED_TITLES` key resolves to a non-empty `TITLES` line, and
 * the nil-contact line `unknownAddressLine(null)` renders. `deckLeaves`
 * only ever reads `TITLES.login` through `pageTitle`, so a deleted title
 * for any other page state would otherwise pass silently.
 */
function checkRequiredTitles(): void {
  for (const key of REQUIRED_TITLES) {
    const value = (TITLES as Record<string, string | undefined>)[key];
    if (!value) {
      fail(`deck-coverage: TITLES has no line for "${key}"`);
    }
  }
  if (!unknownAddressLine(null)) {
    fail('deck-coverage: unknownAddressLine(null) has no line');
  }
}

/** Every plain-text and dictionary string this module exports, flattened one level. */
function deckLeaves(): Leaf[] {
  const scalars: Record<string, string> = {
    LOGIN_SUBLINE,
    CROSS_ORIGIN_REFUSED,
    NETWORK_FAILURE,
    SENT_MOMENT_TEXT,
    SENT_ESCAPE_HATCH,
    INTERSTITIAL_STATEMENT,
    INTERSTITIAL_CONTROL,
    SIGN_OUT_CONTROL,
    ROSTER_EMPTY_STATE,
    THEME_TOGGLE_LABEL
  };
  const dictionaries: Record<string, Record<string, string>> = {
    AUTH_OUTCOMES,
    ADD_MEMBER_LABELS,
    ADD_MEMBER_HINTS,
    ROSTER_OUTCOMES,
    ERROR_PAGE,
    TITLES
  };
  const leaves: Leaf[] = Object.entries(scalars).map(([key, value]) => ({ key, value }));
  for (const [dictName, dict] of Object.entries(dictionaries)) {
    for (const [key, value] of Object.entries(dict)) {
      leaves.push({ key: `${dictName}.${key}`, value });
    }
  }
  leaves.push({ key: 'unknownAddressLine(contact)', value: unknownAddressLine('admin@example.test') });
  leaves.push({ key: 'unknownAddressLine(null)', value: unknownAddressLine(null) });
  leaves.push({ key: 'memberCountLine(1)', value: memberCountLine(1) });
  leaves.push({ key: 'memberCountLine(2)', value: memberCountLine(2) });
  leaves.push({ key: 'pageTitle', value: pageTitle(TITLES.login, 'Example Group') });
  return leaves;
}

/** `internal/auth`'s combined outcome set, or a failure naming the parser or the count. */
function goOutcomes(): Set<string> {
  const handlers = parseGoOutcomes(HANDLERS_PATH);
  const admin = parseGoOutcomes(ADMIN_PATH);
  const parsed = new Set([...handlers, ...admin]);
  for (const must of ['signed_out', 'admin_not_grantable', 'write_unconfirmed']) {
    if (!parsed.has(must)) {
      fail(
        `deck-coverage: the parsed outcome set is missing "${must}"; the constant parser's pattern may be broken.`
      );
    }
  }
  if (parsed.size < 23) {
    fail(
      `deck-coverage: parsed only ${parsed.size} outcome constant(s) from ${HANDLERS_PATH} and ${ADMIN_PATH}, fewer than the 23 internal/auth declares.`
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

const FORBIDDEN_CHAR = /[<>&']/;

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

  process.stdout.write(
    `deck-coverage: passed, ${leaves.length} deck value(s) checked, ${files.length} src file(s) scanned for {@html}.\n`
  );
}

if (isMainModule(import.meta.url)) {
  main();
}
