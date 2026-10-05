/**
 * The one outcome-to-text and outcome-to-class resolver, shared by
 * every page and component that reads a server outcome string:
 * `internal/auth/handlers.go`'s auth outcomes, `internal/auth/admin.go`'s
 * roster outcomes, and the three synthetic values `#lib/api/post.ts`
 * answers with when nothing real came back.
 */
import type { AdminContact } from './auth.js';
import { AUTH_OUTCOMES, CROSS_ORIGIN_REFUSED, NETWORK_FAILURE, ROSTER_OUTCOMES, unknownAddressLine } from '#lib/copy/deck.js';

export type OutcomeClass = 'status' | 'warning' | 'error';

/**
 * The state table's own "Class" column: a status announces politely
 * and carries no alert modifier, a warning is something that failed on
 * the server's own end rather than a refusal. Every outcome missing
 * from this table is an error, the default that covers every refusal
 * both dictionaries name.
 */
const OUTCOME_CLASS: Record<string, OutcomeClass> = {
  sent: 'status',
  requested: 'status',
  unknown: 'status',
  rate_limited: 'status',
  signed_in: 'status',
  signed_out: 'status',
  added: 'status',
  failed: 'warning',
  busy: 'warning',
  write_unconfirmed: 'warning'
};

export function outcomeClass(outcome: string): OutcomeClass {
  return Object.hasOwn(OUTCOME_CLASS, outcome) ? OUTCOME_CLASS[outcome] : 'error';
}

export function outcomePoliteness(outcome: string): 'polite' | 'assertive' {
  return outcomeClass(outcome) === 'status' ? 'polite' : 'assertive';
}

/** The DaisyUI alert modifier for outcome, or '' for a plain status alert. */
export function outcomeAlertClass(outcome: string): string {
  const cls = outcomeClass(outcome);
  return cls === 'warning' ? 'alert-warning' : cls === 'error' ? 'alert-error' : '';
}

/**
 * The deck line for outcome, checking both dictionaries and the two
 * named synthetic values before falling back to the generic failure
 * line. `adminContact` resolves the `unknown` outcome's own admin-named
 * form; a caller with none omits it. Each dictionary lookup is an
 * own-property test, so an outcome named after an inherited member
 * (`toString`, `constructor`) falls through to the fallback line.
 */
export function outcomeText(outcome: string, adminContact?: AdminContact | null): string {
  if (outcome === 'cross_origin') return CROSS_ORIGIN_REFUSED;
  if (outcome === 'network_failure') return NETWORK_FAILURE;
  if (outcome === 'unknown') return unknownAddressLine(adminContact ?? null);
  if (Object.hasOwn(ROSTER_OUTCOMES, outcome)) return ROSTER_OUTCOMES[outcome];
  if (Object.hasOwn(AUTH_OUTCOMES, outcome)) return AUTH_OUTCOMES[outcome];
  return AUTH_OUTCOMES.failed;
}
