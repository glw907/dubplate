import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { describe, expect, it } from 'vitest';
import SpecCard from './SpecCard.svelte';

const twoGiB = 2 * 1024 * 1024 * 1024;

describe('SpecCard', () => {
	it('renders the desktop format panel with the substituted max album size', () => {
		render(SpecCard, { variant: 'desktop', collectionMode: 'standard', maxAlbumBytes: twoGiB });
		expect(screen.getByText(/This library is/)).toBeInTheDocument();
		expect(screen.getByText('2.0 GB')).toBeInTheDocument();
	});

	it('renders the mobile format panel and opens the detail sheet', async () => {
		render(SpecCard, { variant: 'mobile', collectionMode: 'standard', maxAlbumBytes: twoGiB });
		expect(screen.getByText('of the album is easiest. Picking the files works', { exact: false })).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'What can I upload?' }));

		expect(
			screen.getByText("Lossless FLAC, CD quality or better. MP3, AAC, and other lossy files can't go in."),
		).toBeInTheDocument();
		expect(screen.getByText('The checks only look at audio files. Booklets and other extras ride along.')).toBeInTheDocument();
	});

	it('swaps to the private lane contract panel regardless of variant', () => {
		render(SpecCard, { variant: 'desktop', collectionMode: 'private', maxAlbumBytes: twoGiB });
		expect(screen.getByRole('heading', { name: 'Private Recordings' })).toBeInTheDocument();
		expect(screen.getByText('your tags, its own section')).toBeInTheDocument();
		expect(screen.getByText(/No MusicBrainz matching here/)).toBeInTheDocument();
	});
});
