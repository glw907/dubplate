/**
 * The client side of the auth flow's two POST endpoints
 * (`internal/auth/handlers.go`). The shared reading of an outcome into
 * what a page renders lives in `$lib/api/outcome.ts`.
 */
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
