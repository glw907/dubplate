import { describe, expect, it } from 'vitest';
import { formatBytes } from './format';

describe('formatBytes', () => {
	it('renders sub-gibibyte sizes in MiB', () => {
		expect(formatBytes(45.2 * 1024 * 1024)).toBe('45.2 MiB');
	});

	it('renders gibibyte-and-above sizes in GiB', () => {
		expect(formatBytes(2 * 1024 * 1024 * 1024)).toBe('2.0 GiB');
	});

	it('rounds to one decimal place', () => {
		expect(formatBytes(1536 * 1024)).toBe('1.5 MiB');
	});
});
