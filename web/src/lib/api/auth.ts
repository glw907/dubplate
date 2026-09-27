/**
 * The client side of the auth flow's two POST endpoints
 * (`internal/auth/handlers.go`), and the shared reading of an outcome
 * into what a page renders. Every URL composes from the runtime base:
 * `apiURL` joins kit's `resolve('/')` with the endpoint name rather
 * than a root-absolute literal, so a deployment under any prefix
 * reaches its own server.
 */
import { resolve } from '$app/paths';
import { AUTH_OUTCOMES, CROSS_ORIGIN_REFUSED, NETWORK_FAILURE, unknownAddressLine } from '$lib/copy/deck';

function apiURL(name: string): string {
  return resolve('/') + name;
}

/** The roster's admin contact, as `internal/auth/handlers.go`'s `answer.AdminContact` field names it. */
export interface AdminContact {
  displayName: string;
  address: string;
}

interface AuthAnswer {
  outcome: string;
  admin_contact?: { display_name: string; address: string };
}

/**
 * One POST's result. `outcome` is the server's own outcome value, or
 * one of two synthetic values this module assigns for a case the
 * server never answers with a body: `cross_origin` for the CSRF
 * defense's 403 (`internal/serve/origin.go`, no JSON body), and
 * `network_failure` for a request that never reached the server at
 * all (a network failure, or an intercepted request the page never
 * meant to make).
 */
export interface AuthResult {
  outcome: string;
  adminContact: AdminContact | null;
}

async function post(url: string, body: URLSearchParams): Promise<AuthResult> {
  let response: Response;
  try {
    response = await fetch(url, {
      method: 'POST',
      body,
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    });
  } catch {
    return { outcome: 'network_failure', adminContact: null };
  }
  // The CSRF defense's deny handler answers 403 with a plain-text
  // body, never JSON, so its status is read before any JSON parse is
  // attempted.
  if (response.status === 403) {
    return { outcome: 'cross_origin', adminContact: null };
  }
  let answer: AuthAnswer;
  try {
    answer = (await response.json()) as AuthAnswer;
  } catch {
    return { outcome: 'network_failure', adminContact: null };
  }
  const adminContact = answer.admin_contact
    ? { displayName: answer.admin_contact.display_name, address: answer.admin_contact.address }
    : null;
  return { outcome: answer.outcome, adminContact };
}

/** Posts address to the request-link endpoint. */
export function requestLink(address: string): Promise<AuthResult> {
  return post(apiURL('request-link'), new URLSearchParams({ address }));
}

/** Posts token to the consume-link endpoint. */
export function consumeLink(token: string): Promise<AuthResult> {
  return post(apiURL('consume-link'), new URLSearchParams({ token }));
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
  return AUTH_OUTCOMES[result.outcome] ?? AUTH_OUTCOMES.failed;
}
