/**
 * The JSON client for the cookie mount's `/api/v1` operations
 * (`internal/ops/opshttp/mount.go`, contract `openapi.yaml`). Every URL
 * composes from the runtime base through `endpointPath`, as the form
 * POST's does, so a deployment under any prefix reaches its own server.
 */
import { API_DETAILS, API_OUTCOMES, API_REFUSALS, CROSS_ORIGIN_REFUSED, NETWORK_FAILURE } from '#lib/copy/deck.js';
import { endpointPath } from './post.js';
import { send } from './request.js';

/** One member as `list_members` and `add_member` carry it. */
export interface Member {
  user_id: string;
  address: string;
  display_name: string;
  username: string;
  admin: boolean;
}

/** One registered key as `list_settings` answers it. */
export interface SettingEntry {
  key: string;
  home: string;
  description: string;
  apply: string;
  sensitivity: string;
  value?: string;
  source?: string;
  seed_value?: string;
}

/** What `set_setting` and `reset_setting` answer with. */
export interface SettingWriteResponse {
  key: string;
  apply: string;
  source: string;
}

/**
 * One call's result. `refused` is a JSON error body, so `code` and
 * `detail` are the server's own; `cross_origin` is the origin check's
 * plain-text 403, the only 403 that is not JSON; `network_failure` is a
 * request that never reached the server; `failed` is every other answer
 * that arrived without the JSON shape the contract promises.
 */
export type ApiResult<T> =
  | { kind: 'ok'; value: T }
  | { kind: 'refused'; status: number; code: string; detail: string }
  | { kind: 'cross_origin' }
  | { kind: 'network_failure' }
  | { kind: 'failed' };

/** The `{ error: { code, detail? } }` body of a refusal, or null when `body` is not one. */
function readError(body: unknown): { code: string; detail: string } | null {
  if (typeof body !== 'object' || body === null || !('error' in body)) return null;
  const error = body.error;
  if (typeof error !== 'object' || error === null || !('code' in error) || typeof error.code !== 'string') return null;
  const detail = 'detail' in error && typeof error.detail === 'string' ? error.detail : '';
  return { code: error.code, detail };
}

/** Posts `request` as JSON to the operation `operation` and reads the answer. */
export async function call<T>(operation: string, request: object = {}): Promise<ApiResult<T>> {
  const response = await send(endpointPath(`api/v1/${operation}`), {
    method: 'POST',
    body: JSON.stringify(request),
    headers: { 'Content-Type': 'application/json' }
  });
  if (response === null) return { kind: 'network_failure' };
  let body: unknown;
  try {
    body = await response.json();
  } catch {
    // Only the origin check answers 403 without a JSON body.
    return response.status === 403 ? { kind: 'cross_origin' } : { kind: 'failed' };
  }
  if (response.ok) {
    return typeof body === 'object' && body !== null ? { kind: 'ok', value: body as T } : { kind: 'failed' };
  }
  const error = readError(body);
  return error ? { kind: 'refused', status: response.status, ...error } : { kind: 'failed' };
}

/**
 * Reads a list operation's `field` array. A failure names the state the page changes to: a
 * non-admin's `forbidden` is `refused`, a gone session `signedOut`, and anything else, a body
 * whose `field` is not an array included, `unreachable` with the deck line to show.
 */
export async function loadList<T>(
  operation: string,
  field: string
): Promise<{ ok: true; entries: T[] } | { ok: false; state: 'refused' | 'signedOut' | 'unreachable'; message: string }> {
  const result = await call<Record<string, unknown>>(operation);
  if (result.kind === 'ok') {
    const list = result.value[field];
    if (!Array.isArray(list)) return { ok: false, state: 'unreachable', message: API_OUTCOMES.failed };
    return { ok: true, entries: list as T[] };
  }
  if (result.kind === 'refused' && result.code === 'forbidden') return { ok: false, state: 'refused', message: '' };
  if (isSignedOut(result)) return { ok: false, state: 'signedOut', message: '' };
  return { ok: false, state: 'unreachable', message: failureText(result) };
}

/** The deck line for a refusal: its detail's line when it has one, else its code's, else the generic failure. */
function refusalText(code: string, detail: string): string {
  if (detail !== '' && Object.hasOwn(API_DETAILS, detail)) return API_DETAILS[detail];
  if (Object.hasOwn(API_REFUSALS, code)) return API_REFUSALS[code];
  return API_OUTCOMES.failed;
}

/** The deck line for any result that is not `ok`. */
export function failureText(result: Exclude<ApiResult<unknown>, { kind: 'ok' }>): string {
  switch (result.kind) {
    case 'refused':
      return refusalText(result.code, result.detail);
    case 'cross_origin':
      return CROSS_ORIGIN_REFUSED;
    case 'network_failure':
      return NETWORK_FAILURE;
    case 'failed':
      return API_OUTCOMES.failed;
  }
}

/** Codes that mean the server, not the caller, failed: they alert as warnings, every other refusal as an error. */
const SERVER_SIDE_CODES = ['internal', 'unavailable', 'canceled'];

/** The DaisyUI alert modifier for a failed result. */
export function failureAlertClass(result: Exclude<ApiResult<unknown>, { kind: 'ok' }>): string {
  if (result.kind === 'refused') return SERVER_SIDE_CODES.includes(result.code) ? 'alert-warning' : 'alert-error';
  return result.kind === 'cross_origin' ? 'alert-error' : 'alert-warning';
}

/** Whether a failed result asks for a fresh sign-in. */
export function needsReauthentication(result: Exclude<ApiResult<unknown>, { kind: 'ok' }>): boolean {
  return result.kind === 'refused' && result.code === 'reauthentication_required';
}

/** Whether a failed result means the session is gone. */
export function isSignedOut(result: Exclude<ApiResult<unknown>, { kind: 'ok' }>): boolean {
  return result.kind === 'refused' && result.code === 'unauthenticated';
}
