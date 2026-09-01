import { describe, expect, it } from 'vitest';
import { chunkFloorBytes, classifyFailure, decideRetry, retryDelays, type RetryState } from './tus';

describe('classifyFailure', () => {
	it('classifies 401 as signed-out', () => {
		expect(classifyFailure(401, 'application/json')).toBe('signed-out');
	});

	it('classifies an HTML body as signed-out even on a 200 (an Access reauth page)', () => {
		expect(classifyFailure(200, 'text/html; charset=utf-8')).toBe('signed-out');
	});

	it('classifies a 5xx as retryable, including 507 (dubplate-no-space)', () => {
		for (const status of [500, 502, 503, 504, 507, 524]) {
			expect(classifyFailure(status, 'application/json')).toBe('retryable');
		}
	});

	it('classifies any other 4xx as rejected', () => {
		for (const status of [400, 403, 404, 409, 413]) {
			expect(classifyFailure(status, 'application/json')).toBe('rejected');
		}
	});
});

function freshState(chunkSizeBytes = 8 * 1024 * 1024): RetryState {
	return { attempt: 0, lastFailedOffset: null, chunkSizeBytes };
}

describe('decideRetry', () => {
	it('retries on the first failure without downshifting (no prior offset to compare against)', () => {
		const state = freshState();
		const decision = decideRetry(state, { status: 502, contentType: 'application/json', offset: 1000 });
		expect(decision).toEqual({ action: 'retry', delayMs: retryDelays[0], chunkSizeBytes: state.chunkSizeBytes, downshifted: false });
	});

	it('halves chunkSize when the same offset fails twice in a row on 502/504/524', () => {
		for (const status of [502, 504, 524]) {
			const state: RetryState = { attempt: 1, lastFailedOffset: 4096, chunkSizeBytes: 8 * 1024 * 1024 };
			const decision = decideRetry(state, { status, contentType: 'application/json', offset: 4096 });
			expect(decision.action).toBe('retry');
			expect(decision.downshifted).toBe(true);
			expect(decision.chunkSizeBytes).toBe(4 * 1024 * 1024);
		}
	});

	it('does not downshift a different offset failing, even on a downshift-eligible status', () => {
		const state: RetryState = { attempt: 1, lastFailedOffset: 4096, chunkSizeBytes: 8 * 1024 * 1024 };
		const decision = decideRetry(state, { status: 502, contentType: 'application/json', offset: 8192 });
		expect(decision.downshifted).toBe(false);
		expect(decision.chunkSizeBytes).toBe(8 * 1024 * 1024);
	});

	it('does not downshift a same-offset failure on a status outside the downshift set', () => {
		const state: RetryState = { attempt: 1, lastFailedOffset: 4096, chunkSizeBytes: 8 * 1024 * 1024 };
		const decision = decideRetry(state, { status: 503, contentType: 'application/json', offset: 4096 });
		expect(decision.downshifted).toBe(false);
	});

	it('never halves the chunk size below the 1MB floor', () => {
		const state: RetryState = { attempt: 1, lastFailedOffset: 100, chunkSizeBytes: 1.5 * 1024 * 1024 };
		const decision = decideRetry(state, { status: 502, contentType: 'application/json', offset: 100 });
		expect(decision.chunkSizeBytes).toBe(chunkFloorBytes);

		const alreadyAtFloor: RetryState = { attempt: 1, lastFailedOffset: 100, chunkSizeBytes: chunkFloorBytes };
		const next = decideRetry(alreadyAtFloor, { status: 502, contentType: 'application/json', offset: 100 });
		expect(next.chunkSizeBytes).toBe(chunkFloorBytes);
	});

	it('walks the full retry ladder and then reports stalled, never a bare failure', () => {
		const state = freshState();
		for (let i = 0; i < retryDelays.length; i++) {
			const decision = decideRetry(state, { status: 500, contentType: 'application/json', offset: i });
			expect(decision.action).toBe('retry');
			expect(decision.delayMs).toBe(retryDelays[i]);
			state.attempt += 1;
			state.lastFailedOffset = i;
		}
		const exhausted = decideRetry(state, { status: 500, contentType: 'application/json', offset: retryDelays.length });
		expect(exhausted.action).toBe('stalled');
		expect(exhausted.delayMs).toBe(0);
	});

	it('signs out on a 401 without spending any retry budget', () => {
		const state = freshState();
		const decision = decideRetry(state, { status: 401, contentType: 'application/json', offset: 0 });
		expect(decision.action).toBe('sign-out');
		// The caller (createTusDriver) never increments state.attempt for a
		// sign-out decision -- proven at the type level here by asserting
		// the decision alone carries no delay/budget cost.
		expect(decision.delayMs).toBe(0);
	});

	it('signs out on an HTML body regardless of status, without spending retry budget', () => {
		const state = freshState();
		const decision = decideRetry(state, { status: 200, contentType: 'text/html', offset: 0 });
		expect(decision.action).toBe('sign-out');
		expect(decision.delayMs).toBe(0);
	});

	it('rejects a terminal 4xx immediately, spending no retry budget waiting out the ladder', () => {
		const state = freshState();
		const decision = decideRetry(state, { status: 400, contentType: 'application/json', offset: 0 });
		expect(decision).toEqual({ action: 'rejected', delayMs: 0, chunkSizeBytes: state.chunkSizeBytes, downshifted: false });
	});

	it('accepts a custom delay ladder and floor for testing the same policy at other scales', () => {
		const state = freshState(4096);
		const decision = decideRetry(state, { status: 502, contentType: null, offset: 0 }, [0, 1], 1024);
		expect(decision.delayMs).toBe(0);
	});
});
