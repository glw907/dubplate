/**
 * Vitest setup: extends `expect` with jest-dom's DOM matchers
 * (`toBeInTheDocument`, `toHaveAttribute`, ...) and unmounts every
 * rendered component after each test -- without this, one test's DOM
 * (and its `window` event listeners) leaks into the next.
 */
import '@testing-library/jest-dom/vitest';
import { cleanup } from '@testing-library/svelte';
import { afterEach } from 'vitest';

afterEach(() => {
	cleanup();
});

// jsdom implements the <dialog> element itself but not its imperative
// methods (https://github.com/jsdom/jsdom/issues/3294); SpecCard's
// detail sheet needs a working showModal/close pair to test the open
// interaction without stubbing it in every test file.
if (typeof HTMLDialogElement !== 'undefined') {
	HTMLDialogElement.prototype.showModal ??= function showModal(this: HTMLDialogElement): void {
		this.setAttribute('open', '');
	};
	HTMLDialogElement.prototype.close ??= function close(this: HTMLDialogElement): void {
		this.removeAttribute('open');
	};
}
