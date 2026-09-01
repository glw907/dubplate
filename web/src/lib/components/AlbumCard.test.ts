import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { describe, expect, it, vi } from 'vitest';
import AlbumCard, { deriveTitleFromFilename, trackRowLabel } from './AlbumCard.svelte';

describe('deriveTitleFromFilename', () => {
	it('strips the extension and a leading track-number prefix', () => {
		expect(deriveTitleFromFilename('01 - Song Name.flac')).toBe('Song Name');
		expect(deriveTitleFromFilename('02.Another Song.flac')).toBe('Another Song');
		expect(deriveTitleFromFilename('03_Third Song.flac')).toBe('Third Song');
	});

	it('falls back to the bare stem when there is no number prefix', () => {
		expect(deriveTitleFromFilename('Song Name.flac')).toBe('Song Name');
	});
});

describe('trackRowLabel', () => {
	it('zero-pads a one-based track number with the "from filename" suffix', () => {
		expect(trackRowLabel(0)).toBe('01 (from filename)');
		expect(trackRowLabel(9)).toBe('10 (from filename)');
	});
});

describe('AlbumCard', () => {
	function baseProps() {
		return {
			trackFilenames: ['01 - Song Name.flac', '02 - Second Song.flac'],
			trackTitles: ['Song Name', 'Second Song'],
			artist: '',
			album: '',
			year: '',
			onArtistInput: vi.fn(),
			onAlbumInput: vi.fn(),
			onYearInput: vi.fn(),
			onTrackTitleInput: vi.fn(),
			onSubmit: vi.fn(),
		};
	}

	it('renders one labeled row per track, prefilled from the filename', () => {
		render(AlbumCard, baseProps());
		expect(screen.getByText('01 (from filename)')).toBeInTheDocument();
		expect(screen.getByDisplayValue('Song Name')).toBeInTheDocument();
		expect(screen.getByText('02 (from filename)')).toBeInTheDocument();
		expect(screen.getByDisplayValue('Second Song')).toBeInTheDocument();
	});

	it('reports edits through the callback props rather than owning state', async () => {
		const props = baseProps();
		render(AlbumCard, props);

		await fireEvent.input(screen.getByLabelText('Artist'), { target: { value: 'The Band' } });
		expect(props.onArtistInput).toHaveBeenCalledWith('The Band');

		await fireEvent.input(screen.getByDisplayValue('Song Name'), { target: { value: 'Renamed Song' } });
		expect(props.onTrackTitleInput).toHaveBeenCalledWith(0, 'Renamed Song');
	});

	it('submits via the callback prop, never a fetch call of its own', async () => {
		const props = baseProps();
		render(AlbumCard, props);

		await fireEvent.click(screen.getByRole('button', { name: 'Upload to Private Recordings' }));

		expect(props.onSubmit).toHaveBeenCalled();
	});

	it('disables the submit button when told to', () => {
		render(AlbumCard, { ...baseProps(), submitDisabled: true });
		expect(screen.getByRole('button', { name: 'Upload to Private Recordings' })).toBeDisabled();
	});
});
