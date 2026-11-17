/**
 * The one place the web app calls `fetch`. Answers the `Response`, or
 * `null` for a request that never reached the server (the `fetch`
 * itself threw), which each caller maps to its own `network_failure`.
 * Body, headers, and the 403 reading stay with the caller.
 */
export async function send(url: string, init: RequestInit): Promise<Response | null> {
  try {
    return await fetch(url, init);
  } catch {
    return null;
  }
}
