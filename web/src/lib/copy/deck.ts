/**
 * The string deck: every user-facing string of the login, consumption,
 * and members pages, the error page, and the signed-in header, as
 * named constants and outcome-keyed dictionaries.
 *
 * This module has two readers: the pages, through `#lib/copy/deck.js`,
 * and the rendered harness's `run.ts`, through a relative import
 * under Node's type-stripping loader. So it imports nothing and uses
 * only erasable TypeScript syntax (no `enum`, no `#lib` import): both
 * fail under `node`'s strip-only mode.
 *
 * Every value is plain text. No value carries `<`, `>`, `&`, or a
 * straight apostrophe (U+0027); every apostrophe is the curly U+2019,
 * so no page needs `{@html}` to render one. `deck-coverage.ts` checks
 * both rules against every string this module exports.
 */

/** The login page's subline, under the group's name heading. */
export const LOGIN_SUBLINE = 'Sign in with your email. We’ll send you a link.';

/**
 * The line a member sees for `unknown` when the roster carries no
 * admin: the request-link answer's nil `admin_contact` case.
 */
export const UNKNOWN_ADDRESS_NO_ADMIN =
  'We don’t recognize that address. Check the spelling, or ask whoever runs this server to add you.';

/**
 * The `refused` outcome's line, structured so a page can wrap
 * `linkText` in an anchor without ever restating the sentence around
 * it. `AUTH_OUTCOMES.refused` composes the same three parts, so the
 * two can never drift apart.
 */
export const REFUSED_LINK = {
  before: 'That sign-in link doesn’t work anymore. Each link works only once and expires after a short time. Get a new one from ',
  linkText: 'the sign-in page',
  after: '.'
};

/**
 * Every outcome `internal/auth/handlers.go` answers with, one line
 * each, keyed by the outcome value. Covers both the request-link and
 * the consumption flow, since a status like `rate_limited` answers
 * both. `unknown`'s line is the nil-contact case;
 * `unknownAddressLine` below is the form a page renders once it has
 * an admin contact.
 */
export const AUTH_OUTCOMES: Record<string, string> = {
  sent: 'A sign-in link is on its way to your inbox.',
  requested: 'If that address belongs to a member, we’ll send it a sign-in link.',
  unknown: UNKNOWN_ADDRESS_NO_ADMIN,
  rate_limited: 'Too many requests. Wait a while and try again.',
  invalid_address: 'That doesn’t look like an email address. Check it and try again.',
  too_large: 'That was too much to send. Check what you entered and try again.',
  malformed: 'Something went wrong sending that. Reload the page and try again.',
  failed: 'Something went wrong on our end. Try again in a few minutes.',
  signed_in: 'You’re signed in.',
  refused: `${REFUSED_LINK.before}${REFUSED_LINK.linkText}${REFUSED_LINK.after}`,
  // Never claims revocation: sign-out clears only this browser's
  // cookie and revokes no session.
  signed_out: 'You’re signed out in this browser.'
};

/**
 * The `unknown` outcome's line, given the roster's admin contact
 * (`internal/auth/handlers.go`'s `answer.AdminContact`, `omitempty`,
 * carrying both `display_name` and `address`). A nil contact renders
 * `UNKNOWN_ADDRESS_NO_ADMIN` instead.
 */
export function unknownAddressLine(adminContact: { displayName: string; address: string } | null): string {
  return adminContact
    ? `We don’t recognize that address. Check the spelling, or ask ${adminContact.displayName} (${adminContact.address}) to add you.`
    : UNKNOWN_ADDRESS_NO_ADMIN;
}

/** The cross-origin refusal a request-link submission can meet. */
export const CROSS_ORIGIN_REFUSED =
  'We couldn’t accept that from another site. Go to this site’s address directly and try again.';

/** A request that never reached the server at all (a network failure). */
export const NETWORK_FAILURE = 'We couldn’t reach the server. Check your connection and try again.';

/** The sent moment's own text, shown once the link is on its way: the `sent` outcome's own line. */
export const SENT_MOMENT_TEXT = AUTH_OUTCOMES.sent;

/** The sent moment's escape hatch back to the request-link form. */
export const SENT_ESCAPE_HATCH = 'Use a different address';

/** The consumption interstitial's statement, before its one control. */
export const INTERSTITIAL_STATEMENT = 'To finish signing in, select Sign in.';

/** The consumption interstitial's one control (I2). */
export const INTERSTITIAL_CONTROL = 'Sign in';

/**
 * The consumption page's line when it loaded with no token at all (a
 * reload, or a copied address missing its query): nothing was spent,
 * so the mailed link itself still works.
 */
export const LINK_TOKEN_MISSING = 'Open the link from your email again.';

/** The signed-in header's sign-out control. */
export const SIGN_OUT_CONTROL = 'Sign out';

/** The root layout's skip link, ahead of every page's own content (W3.6). */
export const SKIP_LINK = 'Skip to content';

/** The add-member form's field labels. */
export const ADD_MEMBER_LABELS = {
  address: 'Email address',
  displayName: 'Name',
  username: 'Username'
};

/** The add-member form's own legend, submit control, and admin badge. */
export const ADD_MEMBER_LEGEND = 'Add a member';
export const ADD_MEMBER_SUBMIT = 'Add member';
export const ADMIN_BADGE_LABEL = 'Admin';

/**
 * The roster list's own control when it could not be reached at all
 * (a fetch failure, or a status that is neither a refusal nor a
 * parseable answer): shown beside {@link NETWORK_FAILURE} or
 * {@link AUTH_OUTCOMES}'s `failed` line.
 */
export const ROSTER_RETRY = 'Try again';

/**
 * The add-member form's field hints. `username` states the
 * write path's own shape (`internal/auth/roster.go`'s
 * `usernamePattern`): an alphanumeric first character, then any run of
 * letters, numbers, periods, underscores, hyphens, plus signs, or `@`,
 * up to 64 characters (the field is ASCII-only, so bytes and characters
 * agree).
 */
export const ADD_MEMBER_HINTS = {
  username:
    'The name the member types into a music app. Starts with a letter or number, then any mix of letters, numbers, periods, underscores, hyphens, plus signs, and @. Up to 64 characters.'
};

/**
 * Roster-outcome copy, keyed by the value
 * `internal/auth/admin.go`'s outcome constants answer with. Member-facing
 * lines never say "roster" (W8.2's leak list).
 */
export const ROSTER_OUTCOMES: Record<string, string> = {
  added: 'Added.',
  duplicate_address: 'That email address already belongs to a member.',
  duplicate_username: 'That username is already taken. Choose a different one.',
  invalid_display_name: 'That name is too long, or it has characters it can’t use, such as tabs or line breaks.',
  blank_display_name: 'Enter a name.',
  invalid_username:
    'Start the username with a letter or number, and use only letters, numbers, periods, underscores, hyphens, plus signs, or @.',
  missing_username: 'Enter a username.',
  admin_not_grantable: 'Admins can’t be added from this page.',
  unknown_field: 'This page sent something the server doesn’t accept. Reload the page and try again.',
  busy: 'That didn’t go through. Try again.',
  write_unconfirmed:
    'The member may have been added, but we couldn’t confirm it. Reload the page and check the list before you try again.'
};

/**
 * The roster write's own success line, naming the member the write
 * added. `ROSTER_OUTCOMES.added` stays a plain fallback line for the
 * deck's own coverage check; a page that has the member's name calls
 * this instead.
 */
export function addedMemberLine(displayName: string): string {
  return `Added ${displayName}.`;
}

/**
 * What a refused `/api/v1` call says, keyed by its wire code
 * (`internal/ops/errors.go`'s codes). A code with an `error.detail` reads
 * {@link API_DETAILS} first. Member-facing lines never say "roster".
 */
export const API_REFUSALS: Record<string, string> = {
  unauthenticated: 'You’re signed out. Sign in again to continue.',
  reauthentication_required:
    'This change needs a recent sign-in. Sign in again in this browser, then try it again.',
  forbidden: 'Your account can’t make this change.',
  not_found: 'That member is no longer in the group. Reload the page.',
  conflict: 'That doesn’t match what the group has now. Reload the page and try again.',
  invalid_request: 'The server didn’t accept that. Reload the page and try again.',
  unavailable: 'That didn’t go through. Try again.',
  canceled: 'That didn’t go through. Try again.',
  internal: AUTH_OUTCOMES.failed
};

/**
 * The finer `error.detail` lines, keyed by the value
 * `internal/ops/details.go` declares. `deck-coverage.ts` reads that file
 * and fails on a detail with no line here.
 */
export const API_DETAILS: Record<string, string> = {
  duplicate_address: 'That email address already belongs to a member.',
  duplicate_username: 'That username is already taken. Choose a different one.',
  invalid_address: 'That doesn’t look like an email address. Check it and try again.',
  invalid_display_name: 'That name is too long, or it has characters it can’t use, such as tabs or line breaks.',
  blank_display_name: 'Enter a name.',
  invalid_username:
    'Start the username with a letter or number, and use only letters, numbers, periods, underscores, hyphens, plus signs, or @.',
  missing_username: 'Enter a username.',
  last_admin: 'The group needs at least one admin, so that member can’t be removed or demoted.',
  environment_sets: 'The server’s environment file sets this, so it can’t be changed here.',
  busy: 'That didn’t go through. Try again.'
};

/** The sign-in control a refusal that needs a fresh sign-in offers, and the one for a signed-out page. */
export const REAUTH_SIGN_IN = 'Sign in again';
export const SIGNED_OUT_SIGN_IN = 'Go to the sign-in page';

/** The members page's per-member controls and the confirm step's buttons. */
export const REMOVE_CONTROL = 'Remove';
export const DEMOTE_CONTROL = 'Demote';
export const CONFIRM_REMOVE = 'Remove member';
export const CONFIRM_DEMOTE = 'Demote member';
export const CONFIRM_CANCEL = 'Cancel';

/** A control's accessible name, naming the member its row holds. */
export function removeControlLabel(displayName: string): string {
  return `Remove ${displayName}`;
}

export function demoteControlLabel(displayName: string): string {
  return `Demote ${displayName}`;
}

/** The confirm step's question. A destructive confirm is plain and never playful. */
export function confirmRemoveLine(displayName: string): string {
  return `Remove ${displayName} from the group? They’ll be signed out right away.`;
}

export function confirmDemoteLine(displayName: string): string {
  return `Demote ${displayName}? They’ll stay in the group without admin rights.`;
}

/** The success lines for a removal and a demotion, naming the member. */
export function removedMemberLine(displayName: string): string {
  return `Removed ${displayName}.`;
}

export function demotedMemberLine(displayName: string): string {
  return `${displayName} is no longer an admin.`;
}

/** The members page's one link to the server settings page, and that page's lead. */
export const SERVER_SETTINGS_LINK = 'Server settings';
export const SETTINGS_INTRO =
  'What the server does, and how. A change takes effect when its line says so. Keys set in the server’s environment file are listed beneath, and change only there.';

/** The settings page's two sections. */
export const SETTINGS_EDITABLE_HEADING = 'Settings you can change';
export const SETTINGS_ENVIRONMENT_HEADING = 'Changed only in the environment file';

/**
 * What a refused value says on the settings page: the server answers
 * `invalid_request` with no detail for a value that does not parse, falls
 * outside its bounds, or breaks a rule between settings, and reloading
 * the page cannot help.
 */
export const SETTING_INVALID_VALUE = 'That value isn’t one this setting takes. Check its description and try again.';

/** The settings page's per-key controls. */
export const SETTING_SAVE = 'Save';
export const SETTING_RESET = 'Reset to default';

/** A key's mark, as the badge on its row reads it. */
export const SETTING_MARKS: Record<string, string> = {
  sensitive: 'Sensitive',
  not_sensitive: 'Not sensitive'
};

/** When a write takes effect, keyed by the apply mode the server answers with. */
export const SETTING_APPLY: Record<string, string> = {
  immediately: 'Takes effect right away.',
  next_operation: 'Takes effect from the next request.',
  restart: 'Takes effect at the next server start.'
};

/** Where a stored value came from, keyed by the source the server answers with. */
export const SETTING_SOURCES: Record<string, string> = {
  admin: 'Changed here.',
  'env-seed': 'Taken from the environment file at first start.',
  default: 'The default.'
};

/** A key only the environment file sets, with no value shown. */
export const ENVIRONMENT_ONLY_LINE = 'Only the server’s environment file sets this. Change it there and restart the server.';

/** A key whose value the environment file sets, so it cannot be edited or reset here. */
export function environmentValueLine(value: string): string {
  return `Set to ${value} in the server’s environment file. Change it there and restart the server.`;
}

/** What a restart-applied key's stored value means: it is the one the next start applies. */
export function nextStartLine(value: string): string {
  return `The stored value ${value} applies at the next server start.`;
}

/** The success lines for a save and a reset, naming the key and when it takes effect. */
export function settingSavedLine(key: string, applyLine: string): string {
  return `Saved ${key}. ${applyLine}`;
}

export function settingResetLine(key: string, applyLine: string): string {
  return `Reset ${key} to its default. ${applyLine}`;
}

/** A control's accessible name, naming the key its row holds. */
export function saveSettingLabel(key: string): string {
  return `Save ${key}`;
}

export function resetSettingLabel(key: string): string {
  return `Reset to default: ${key}`;
}

/** The roster list's empty state (no members yet). */
export const ROSTER_EMPTY_STATE = 'No members yet.';

/** The roster list's plural-aware member count. */
export function memberCountLine(count: number): string {
  return count === 1 ? '1 member' : `${count} members`;
}

/** The client router's own not-found and generic error states. */
export const ERROR_PAGE = {
  notFoundHeading: 'That page doesn’t exist.',
  genericHeading: 'Something went wrong.'
};

/** The theme toggle's accessible name (VI `:141`). */
export const THEME_TOGGLE_LABEL = 'Dark theme';

/**
 * Every page `<title>`'s state string, one per page and state where
 * the state changes it (W7.14), in VI's "Sign in" register. `title.ts`
 * composes each of these with the served site name.
 */
export const TITLES = {
  login: 'Sign in',
  loginSent: 'Check your email',
  loginLink: 'Confirm sign-in',
  loginLinkSignedIn: 'Signed in',
  loginLinkRefused: 'That link didn’t work',
  members: 'Members',
  serverSettings: 'Server settings',
  signedOut: 'Signed out',
  notFound: 'Page not found',
  error: 'Something went wrong'
} as const;
