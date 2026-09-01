import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { describe, expect, it, vi } from 'vitest';
import CollectionPicker, { PRIVATE_COLLECTION_NAME, isPrivateCollection } from './CollectionPicker.svelte';

describe('CollectionPicker', () => {
	it('renders nothing when there is only one collection', () => {
		render(CollectionPicker, { collections: ['Library'], selected: 'Library', onSelect: vi.fn() });
		expect(screen.queryByLabelText('Collection')).not.toBeInTheDocument();
	});

	it('renders the configured names verbatim when there is a choice', () => {
		render(CollectionPicker, {
			collections: ['Library', PRIVATE_COLLECTION_NAME],
			selected: 'Library',
			onSelect: vi.fn(),
		});
		const select = screen.getByLabelText('Collection') as HTMLSelectElement;
		expect(Array.from(select.options).map((o) => o.value)).toEqual(['Library', PRIVATE_COLLECTION_NAME]);
	});

	it('calls onSelect with the chosen collection', async () => {
		const onSelect = vi.fn();
		render(CollectionPicker, {
			collections: ['Library', PRIVATE_COLLECTION_NAME],
			selected: 'Library',
			onSelect,
		});
		const select = screen.getByLabelText('Collection') as HTMLSelectElement;

		await fireEvent.change(select, { target: { value: PRIVATE_COLLECTION_NAME } });

		expect(onSelect).toHaveBeenCalledWith(PRIVATE_COLLECTION_NAME);
	});
});

describe('isPrivateCollection', () => {
	it('matches only the documented default name', () => {
		expect(isPrivateCollection('Private Recordings')).toBe(true);
		expect(isPrivateCollection('Library')).toBe(false);
	});
});
