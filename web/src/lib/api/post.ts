/**
 * The one form POST every page's write goes through, and its reading
 * of the answer into an outcome. Every URL composes from the runtime
 * base: `resolve('/')` joined with the endpoint name rather than a
 * root-absolute literal, so a deployment under any prefix reaches its
 * own server.
 */
import { resolve } from '$app/paths';

/**
 * One POST's result. `outcome` is the server's own outcome value, or
 * one of three synthetic values for an answer that carries none:
 * `cross_origin` for the CSRF defense's 403 (`internal/serve/origin.go`,
 * no JSON body), `network_failure` for a request that never reached the
 * server (the `fetch` itself threw), and `failed` for every other
 * answer that arrived but carried no JSON outcome (a proxy's error
 * page, a 404, any status). `body` is the parsed answer, empty for a
 * synthetic outcome.
 */
export interface PostResult {
  ok: boolean;
  outcome: string;
  body: Record<string, unknown>;
}

function synthetic(outcome: string): PostResult {
  return { ok: false, outcome, body: {} };
}

/** Posts `fields` form-encoded to the endpoint `name` under the runtime base. */
export async function postForm(name: string, fields: Record<string, string>): Promise<PostResult> {
  let response: Response;
  try {
    response = await fetch(resolve('/') + name, {
      method: 'POST',
      body: new URLSearchParams(fields),
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' }
    });
  } catch {
    return synthetic('network_failure');
  }
  // The CSRF defense's deny handler answers 403 with a plain-text
  // body, never JSON, so its status is read before any JSON parse.
  if (response.status === 403) return synthetic('cross_origin');
  const notJSON = synthetic('failed');
  let body: unknown;
  try {
    body = await response.json();
  } catch {
    return notJSON;
  }
  if (typeof body !== 'object' || body === null || !('outcome' in body) || typeof body.outcome !== 'string') {
    return notJSON;
  }
  return { ok: response.ok, outcome: body.outcome, body: body as Record<string, unknown> };
}
