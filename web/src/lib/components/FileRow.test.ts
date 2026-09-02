import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import FileRow from './FileRow.svelte';
import type { FileVerdict } from '../store/upload.svelte';

const notFlac: FileVerdict = {
	Name: 'track01.flac',
	OK: false,
	Reason: {
		code: 'bad-extension',
		friendly: "This isn't a FLAC file, and the library keeps lossless FLAC.",
		next_step: 'Download this album as FLAC and upload that instead.',
		label: 'not FLAC',
	},
};

function baseProps() {
	return {
		name: 'track01.flac',
		size: 30 * 1024 * 1024,
		status: 'pending' as const,
		progressBytes: 0,
		ackedBytes: 0,
		verdict: null,
		failureKind: null,
		variant: 'progress' as const,
	};
}

describe('FileRow', () => {
	it('shows the filename', () => {
		render(FileRow, baseProps());
		expect(screen.getByText('track01.flac')).toBeInTheDocument();
	});

	it('chips "waiting" for a pending file and "sending" for one in flight', () => {
		const { unmount } = render(FileRow, baseProps());
		expect(screen.getByText('waiting')).toBeInTheDocument();
		unmount();

		render(FileRow, { ...baseProps(), status: 'uploading' });
		expect(screen.getByText('sending')).toBeInTheDocument();
	});

	it('renders the acked-vs-total progress readout only while uploading', () => {
		render(FileRow, {
			...baseProps(),
			status: 'uploading',
			progressBytes: 10 * 1024 * 1024,
			ackedBytes: 8 * 1024 * 1024,
		});
		expect(screen.getByText('8.0 MiB confirmed')).toBeInTheDocument();
		expect(screen.getByText('of 30.0 MiB')).toBeInTheDocument();
	});

	it('uses the progress variant\'s "looks good" word for a passed file', () => {
		render(FileRow, { ...baseProps(), status: 'ok', variant: 'progress' });
		expect(screen.getByText('looks good')).toBeInTheDocument();
	});

	it('uses the verdict variant\'s "fine" word for a passed file', () => {
		render(FileRow, { ...baseProps(), status: 'ok', variant: 'verdict' });
		expect(screen.getByText('fine')).toBeInTheDocument();
	});

	it('chips "stopped" (never "waiting") for a transport failure awaiting retry', () => {
		render(FileRow, { ...baseProps(), status: 'failed', failureKind: 'connection', verdict: null });
		expect(screen.getByText('stopped')).toBeInTheDocument();
		expect(screen.queryByText('waiting')).not.toBeInTheDocument();
	});

	it("shows the server verdict's own label for a rejected file in either variant", () => {
		const { unmount } = render(FileRow, {
			...baseProps(),
			status: 'failed',
			failureKind: 'rejected',
			verdict: notFlac,
			variant: 'progress',
		});
		expect(screen.getByText('not FLAC')).toBeInTheDocument();
		unmount();

		render(FileRow, {
			...baseProps(),
			status: 'failed',
			failureKind: 'rejected',
			verdict: notFlac,
			variant: 'verdict',
		});
		expect(screen.getByText('not FLAC')).toBeInTheDocument();
	});
});
