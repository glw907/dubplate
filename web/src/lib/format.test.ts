import { describe, expect, it } from 'vitest';
import { formatBytes } from './format';

describe('formatBytes', () => {
	it('renders sub-gibibyte sizes in MB', () => {
		expect(formatBytes(45.2 * 1024 * 1024)).toBe('45.2 MB');
	});

	it('renders gibibyte-and-above sizes in GB', () => {
		expect(formatBytes(2 * 1024 * 1024 * 1024)).toBe('2.0 GB');
	});

	it('rounds to one decimal place', () => {
		expect(formatBytes(1536 * 1024)).toBe('1.5 MB');
	});
});
