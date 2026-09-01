import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { describe, expect, it, vi } from 'vitest';
import CollectionPicker, { isPrivateCollection } from './CollectionPicker.svelte';

const privateCollectionName = 'Private Recordings';

describe('CollectionPicker', () => {
	it('renders nothing when there is only one collection', () => {
		render(CollectionPicker, { collections: ['Library'], selected: 'Library', onSelect: vi.fn() });
		expect(screen.queryByLabelText('Collection')).not.toBeInTheDocument();
	});

	it('renders the configured names verbatim when there is a choice', () => {
		render(CollectionPicker, {
			collections: ['Library', privateCollectionName],
			selected: 'Library',
			onSelect: vi.fn(),
		});
		const select = screen.getByLabelText('Collection') as HTMLSelectElement;
		expect(Array.from(select.options).map((o) => o.value)).toEqual(['Library', privateCollectionName]);
	});

	it('calls onSelect with the chosen collection', async () => {
		const onSelect = vi.fn();
		render(CollectionPicker, {
			collections: ['Library', privateCollectionName],
			selected: 'Library',
			onSelect,
		});
		const select = screen.getByLabelText('Collection') as HTMLSelectElement;

		await fireEvent.change(select, { target: { value: privateCollectionName } });

		expect(onSelect).toHaveBeenCalledWith(privateCollectionName);
	});
});

describe('isPrivateCollection', () => {
	it('matches only the configured private collection name', () => {
		expect(isPrivateCollection('Private Recordings', 'Private Recordings')).toBe(true);
		expect(isPrivateCollection('Library', 'Private Recordings')).toBe(false);
	});

	it('follows an operator-renamed private lane rather than assuming the documented default', () => {
		expect(isPrivateCollection('Home Recordings', 'Home Recordings')).toBe(true);
		expect(isPrivateCollection('Private Recordings', 'Home Recordings')).toBe(false);
	});

	it('never matches on two empty strings (config not loaded yet)', () => {
		expect(isPrivateCollection('', '')).toBe(false);
	});
});
