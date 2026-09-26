/**
 * The string deck: every user-facing string of the login, consumption,
 * and members pages, the error page, and the signed-in header, as
 * named constants and outcome-keyed dictionaries.
 *
 * This module has two readers: the pages, through `$lib/copy/deck`,
 * and the rendered harness's `run.ts`, through a relative import
 * under Node's type-stripping loader. So it imports nothing and uses
 * only erasable TypeScript syntax (no `enum`, no `$lib` import): both
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
  'We don’t recognize that address. Ask whoever runs this server to add you.';

/**
 * Every outcome `internal/auth/handlers.go` answers with, one line
 * each, keyed by the outcome value. Covers both the request-link and
 * the consumption flow, since a status like `rate_limited` answers
 * both. `unknown`'s line is the nil-contact case;
 * `unknownAddressLine` below is the form a page renders once it has
 * an admin contact.
 */
export const AUTH_OUTCOMES: Record<string, string> = {
  sent: 'Check your email for a sign-in link.',
  requested: 'If that address is on the roster, a link is on its way.',
  unknown: UNKNOWN_ADDRESS_NO_ADMIN,
  rate_limited: 'Too many requests. Wait a few minutes and try again.',
  invalid_address: 'That doesn’t look like an email address.',
  too_large: 'That request was too large.',
  malformed: 'That request didn’t come through right. Try again.',
  failed: 'Something went wrong with that request. Try again.',
  signed_in: 'You’re signed in.',
  refused: 'That link didn’t work. Ask for a new one.',
  // Never claims revocation: this line describes only this device's
  // session ending, since a magic-link server has no way to end a
  // session on any other device.
  signed_out: 'You’re signed out on this device.'
};

/**
 * The `unknown` outcome's line, given the roster's admin contact
 * (`internal/auth/handlers.go:157-164`'s `admin_contact`, `omitempty`).
 * A nil contact renders `UNKNOWN_ADDRESS_NO_ADMIN` instead.
 */
export function unknownAddressLine(adminContact: string | null): string {
  return adminContact
    ? `We don’t recognize that address. Ask ${adminContact} to add you.`
    : UNKNOWN_ADDRESS_NO_ADMIN;
}

/** The cross-origin refusal a request-link submission can meet. */
export const CROSS_ORIGIN_REFUSED = 'That request didn’t come from this site.';

/** A request that never reached the server at all (a network failure). */
export const NETWORK_FAILURE = 'We couldn’t reach the server. Check your connection and try again.';

/** The sent moment's own text, shown once the link is on its way. */
export const SENT_MOMENT_TEXT = 'Check your email for a sign-in link.';

/** The sent moment's escape hatch back to the request-link form. */
export const SENT_ESCAPE_HATCH = 'Use a different address';

/** The consumption interstitial's statement, before its one control. */
export const INTERSTITIAL_STATEMENT = 'Confirm your sign-in.';

/** The consumption interstitial's one control (I2). */
export const INTERSTITIAL_CONTROL = 'Sign in';

/** The signed-in header's sign-out control. */
export const SIGN_OUT_CONTROL = 'Sign out';

/** The add-member form's field labels. */
export const ADD_MEMBER_LABELS = {
  address: 'Email address',
  displayName: 'Name',
  streamerUsername: 'Streamer username'
};

/** The add-member form's field hints. */
export const ADD_MEMBER_HINTS = {
  streamerUsername: 'Used to sign in to the streamer. Letters, numbers, and hyphens only.'
};

/**
 * Roster-outcome copy, keyed by the value
 * `internal/auth/admin.go`'s outcome constants answer with.
 */
export const ROSTER_OUTCOMES: Record<string, string> = {
  added: 'Added to the roster.',
  duplicate_address: 'That address is already on the roster.',
  duplicate_streamer_username: 'That streamer username is already taken.',
  duplicate_principal: 'That member is already on the roster.',
  invalid_display_name: 'That name isn’t valid.',
  blank_display_name: 'Enter a name.',
  invalid_streamer_username: 'That streamer username isn’t valid.',
  missing_streamer_username: 'Enter a streamer username.',
  admin_not_grantable: 'That member can’t be made an admin here.',
  unknown_field: 'That request carried a field the roster doesn’t take.',
  busy: 'The roster is busy. Try again.',
  write_unconfirmed: 'The roster write didn’t confirm. Try again.'
};

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
  signedOut: 'Signed out'
};
