import { fireEvent } from '@testing-library/dom';
import { render, screen } from '@testing-library/svelte';
import { describe, expect, it, vi } from 'vitest';
import OutcomeAlert from './OutcomeAlert.svelte';

describe('OutcomeAlert', () => {
	it('renders the no-match copy and its three actions for a pushed-back outcome', async () => {
		const onUploadPrivate = vi.fn();
		const onFixNames = vi.fn();
		render(OutcomeAlert, { outcome: { kind: 'pushed-back', onUploadPrivate, onFixNames } });

		expect(screen.getByText('No MusicBrainz match')).toBeInTheDocument();
		expect(screen.getByText('Private Recordings', { selector: 'strong' })).toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Upload to Private Recordings' }));
		expect(onUploadPrivate).toHaveBeenCalledOnce();

		await fireEvent.click(screen.getByRole('button', { name: 'Fix the names and retry' }));
		expect(onFixNames).toHaveBeenCalledOnce();

		const addLink = screen.getByRole('link', { name: 'Add it to MusicBrainz' });
		expect(addLink).toHaveAttribute('href', 'https://musicbrainz.org/release/add');
		expect(addLink).toHaveAttribute('target', '_blank');
	});

	it('renders the probe-unavailable copy with its single retry action', async () => {
		const onTryAgain = vi.fn();
		render(OutcomeAlert, { outcome: { kind: 'probe-unavailable', onTryAgain } });

		expect(screen.getByText("Can't check MusicBrainz right now")).toBeInTheDocument();
		expect(screen.queryByRole('button', { name: 'Upload to Private Recordings' })).not.toBeInTheDocument();

		await fireEvent.click(screen.getByRole('button', { name: 'Try again' }));
		expect(onTryAgain).toHaveBeenCalledOnce();
	});

	it('renders the own-earlier-upload queued variant', () => {
		render(OutcomeAlert, { outcome: { kind: 'queued', ownEarlierUpload: true } });
		expect(screen.getByText('Waiting its turn')).toBeInTheDocument();
		expect(screen.getByText(/Your earlier album is still importing/)).toBeInTheDocument();
	});

	it('renders the library-busy queued variant', () => {
		render(OutcomeAlert, { outcome: { kind: 'queued', ownEarlierUpload: false } });
		expect(screen.getByText(/The library is finishing another import/)).toBeInTheDocument();
	});

	it('renders the stalled copy', () => {
		render(OutcomeAlert, { outcome: { kind: 'stalled' } });
		expect(screen.getByText('Taking longer than it should')).toBeInTheDocument();
		expect(screen.getByText(/admin gets notified automatically/)).toBeInTheDocument();
	});
});
