/**
 * The client side of the auth flow's two POST endpoints
 * (`internal/auth/handlers.go`), and the shared reading of an outcome
 * into what a page renders.
 */
import { AUTH_OUTCOMES, CROSS_ORIGIN_REFUSED, NETWORK_FAILURE, unknownAddressLine } from '$lib/copy/deck';
import { postForm, type PostResult } from '$lib/api/post';

/** The roster's admin contact, as `internal/auth/handlers.go`'s `answer.AdminContact` field names it. */
export interface AdminContact {
  displayName: string;
  address: string;
}

/**
 * One auth POST's result: the outcome as {@link postForm} reads it, and
 * the admin contact the `unknown` answer may carry.
 */
export interface AuthResult {
  outcome: string;
  adminContact: AdminContact | null;
}

function adminContactOf(body: Record<string, unknown>): AdminContact | null {
  const contact = body.admin_contact;
  if (typeof contact !== 'object' || contact === null) return null;
  const { display_name: displayName, address } = contact as Record<string, unknown>;
  return typeof displayName === 'string' && typeof address === 'string' ? { displayName, address } : null;
}

function authResult(result: PostResult): AuthResult {
  return { outcome: result.outcome, adminContact: adminContactOf(result.body) };
}

/** Posts address to the request-link endpoint. */
export async function requestLink(address: string): Promise<AuthResult> {
  return authResult(await postForm('request-link', { address }));
}

/** Posts token to the consume-link endpoint. */
export async function consumeLink(token: string): Promise<AuthResult> {
  return authResult(await postForm('consume-link', { token }));
}

/**
 * An outcome's announcement class (the state table's own column): a
 * warning and an error both announce assertively, a plain status
 * announces politely. `failed` is the one outcome that mails or
 * mints nothing having gone wrong on the server's own end, so it
 * alone is a warning; every other non-status outcome is an error.
 */
type OutcomeClass = 'status' | 'warning' | 'error';

function outcomeClass(outcome: string): OutcomeClass {
  switch (outcome) {
    case 'failed':
      return 'warning';
    case 'sent':
    case 'requested':
    case 'unknown':
    case 'rate_limited':
    case 'signed_in':
    case 'signed_out':
      return 'status';
    default:
      return 'error';
  }
}

export function outcomePoliteness(outcome: string): 'polite' | 'assertive' {
  return outcomeClass(outcome) === 'status' ? 'polite' : 'assertive';
}

/** The DaisyUI alert modifier for outcome, or '' for a plain status alert. */
export function outcomeAlertClass(outcome: string): string {
  const cls = outcomeClass(outcome);
  return cls === 'warning' ? 'alert-warning' : cls === 'error' ? 'alert-error' : '';
}

/** The deck line for result, resolving the two synthetic outcomes and the roster-absent answer's admin-contact form. */
export function outcomeText(result: AuthResult): string {
  if (result.outcome === 'cross_origin') return CROSS_ORIGIN_REFUSED;
  if (result.outcome === 'network_failure') return NETWORK_FAILURE;
  if (result.outcome === 'unknown') return unknownAddressLine(result.adminContact);
  // An own-property test, so an outcome named after an inherited member
  // (`toString`, `constructor`) falls through to the fallback line.
  return Object.hasOwn(AUTH_OUTCOMES, result.outcome) ? AUTH_OUTCOMES[result.outcome] : AUTH_OUTCOMES.failed;
}
