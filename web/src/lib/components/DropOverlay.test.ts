import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import DropOverlay from './DropOverlay.svelte';

describe('DropOverlay', () => {
	it('renders nothing when inactive', () => {
		render(DropOverlay, { active: false });
		expect(screen.queryByText('Drop an album anywhere on this page')).not.toBeInTheDocument();
	});

	it('renders the drop cue when active', () => {
		render(DropOverlay, { active: true });
		expect(screen.getByText('Drop an album anywhere on this page')).toBeInTheDocument();
	});
});
