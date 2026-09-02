/**
 * The one polite `aria-live` region U11's journey view exists empty
 * from first render (passb-copy-final.md's honesty rules; the plan's
 * U11 criteria): a screen-reader user gets a progress readout without
 * the page's visible layout needing a second announcement surface.
 * This file owns the announcement *policy* -- when a progress delta or
 * a state change is worth interrupting a screen reader for -- as pure,
 * DOM-free functions; {@link LiveRegion} is the thin imperative
 * wrapper Journey.svelte attaches its rendered `<div aria-live="polite">`
 * to, kept a plain class (not `$state`) because this file has no
 * `.svelte.ts` suffix and therefore no rune transform.
 */

/** The minimum announced-since-last-time progress delta, in whole percentage points (plan: "announcements at >=25% deltas"). */
export const progressDeltaThreshold = 25;

/**
 * Reports whether a progress announcement should fire: true the first
 * time (lastPercent null), on reaching 100%, or once percent has moved
 * at least {@link progressDeltaThreshold} points from lastPercent.
 * Pure so the threshold logic is unit-testable without a DOM.
 */
export function shouldAnnounceProgress(lastPercent: number | null, percent: number): boolean {
	if (lastPercent === null || percent >= 100) {
		return true;
	}
	return Math.abs(percent - lastPercent) >= progressDeltaThreshold;
}

/** Formats a progress announcement's text: "N percent confirmed", the server-acknowledged figure (never the raw transfer count -- the store's honest-progress split). */
export function progressAnnouncementText(percent: number): string {
	return `${Math.round(Math.max(0, Math.min(100, percent)))} percent confirmed`;
}

/**
 * The minimal surface {@link LiveRegion} writes to: a real
 * `aria-live` element in the app, or a plain object with a
 * `textContent` field in a test that never mounts one.
 */
export type AnnouncementTarget = Pick<HTMLElement, 'textContent'>;

/**
 * Owns one aria-live region's announced text. Journey.svelte renders
 * the element and calls {@link LiveRegion.attach} in `onMount`; every
 * announcement is a plain text write, since an `aria-live="polite"`
 * region announces on text-content change alone, with no imperative
 * screen-reader API to call.
 */
export class LiveRegion {
	private target: AnnouncementTarget | null = null;
	private lastPercent: number | null = null;

	/** Attaches el as this region's live DOM target; call once, from `onMount`. */
	attach(el: AnnouncementTarget): void {
		this.target = el;
	}

	/** Detaches the current target; call from `onDestroy` so a stale region never receives a late announcement. */
	detach(): void {
		this.target = null;
	}

	private write(text: string): void {
		if (this.target) {
			this.target.textContent = text;
		}
	}

	/**
	 * Announces percent if {@link shouldAnnounceProgress} says this
	 * delta clears the threshold; a call that does not clear it is a
	 * silent no-op, exactly as many times as the store's poll or
	 * progress callback likes.
	 */
	announceProgress(percent: number): void {
		if (!shouldAnnounceProgress(this.lastPercent, percent)) {
			return;
		}
		this.lastPercent = percent;
		this.write(progressAnnouncementText(percent));
	}

	/**
	 * Announces a state-change message verbatim (the spine step or
	 * outcome copy the caller already composed) and resets the progress
	 * baseline, so the next {@link announceProgress} call after a state
	 * change always announces rather than measuring against a delta
	 * that belonged to the previous state.
	 */
	announceState(message: string): void {
		this.lastPercent = null;
		this.write(message);
	}

	/** Announces message as the flow's distinct completion event -- never folded into a state or progress announcement. */
	announceCompletion(message: string): void {
		this.write(message);
	}
}
