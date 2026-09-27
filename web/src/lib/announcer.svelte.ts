/**
 * The app's one pair of live regions (W7.6), rendered once by the root
 * layout and written by every page through {@link announce}. A live
 * region announces a change, so setting the text it already holds says
 * nothing: each call empties both regions first and writes the new text
 * on the next animation frame, which re-announces a repeated identical
 * outcome.
 */

export type Politeness = 'polite' | 'assertive';

const regions = $state({ polite: '', assertive: '' });

// Only the latest call's text lands: a frame still pending from an
// earlier call finds the sequence moved on and writes nothing.
let sequence = 0;

/** The two regions' current text, read by the root layout. */
export const announcer: { readonly polite: string; readonly assertive: string } = regions;

/** Announces `text` in the region for `politeness`; an empty `text` only clears. */
export function announce(text: string, politeness: Politeness): void {
  const call = ++sequence;
  regions.polite = '';
  regions.assertive = '';
  if (text === '') return;
  requestAnimationFrame(() => {
    if (call === sequence) regions[politeness] = text;
  });
}
