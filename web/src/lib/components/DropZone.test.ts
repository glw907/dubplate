import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { createRawSnippet } from 'svelte';
import { describe, expect, it, vi } from 'vitest';
import DropZone from './DropZone.svelte';

/** A minimal `DataTransfer` stand-in -- jsdom implements the interface only partially. */
function fileTransfer(files: File[]): DataTransfer {
	return {
		types: ['Files'],
		files,
	} as unknown as DataTransfer;
}

function childButton(): { children: ReturnType<typeof createRawSnippet> } {
	return {
		children: createRawSnippet(() => ({
			render: () => '<button type="button">Choose files</button>',
		})),
	};
}

describe('DropZone', () => {
	it('renders its slotted content', () => {
		render(DropZone, { onFiles: vi.fn(), ...childButton() });
		expect(screen.getByRole('button', { name: 'Choose files' })).toBeInTheDocument();
	});

	it('shows the overlay while a file drag is over the window and hides it on drop', async () => {
		const onFiles = vi.fn();
		render(DropZone, { onFiles, ...childButton() });
		const file = new File([], 'track01.flac');

		const dragEnter = new Event('dragenter', { cancelable: true });
		Object.defineProperty(dragEnter, 'dataTransfer', { value: fileTransfer([]) });
		await fireEvent(window, dragEnter);
		expect(screen.getByText('Drop an album anywhere on this page', { selector: 'p.text-primary' })).toBeInTheDocument();

		const dropEvent = new Event('drop', { cancelable: true });
		Object.defineProperty(dropEvent, 'dataTransfer', { value: fileTransfer([file]) });
		await fireEvent(window, dropEvent);

		expect(onFiles).toHaveBeenCalledWith([file]);
		expect(screen.queryByText('Drop an album anywhere on this page', { selector: 'p.text-primary' })).not.toBeInTheDocument();
	});

	it('extracts files from a paste event', async () => {
		const onFiles = vi.fn();
		render(DropZone, { onFiles, ...childButton() });
		const file = new File([], 'track01.flac');

		const pasteEvent = new Event('paste');
		Object.defineProperty(pasteEvent, 'clipboardData', {
			value: { items: [{ kind: 'file', getAsFile: () => file }] },
		});
		await fireEvent(window, pasteEvent);

		expect(onFiles).toHaveBeenCalledWith([file]);
	});

	it('ignores a drag that carries no files', async () => {
		render(DropZone, { onFiles: vi.fn(), ...childButton() });

		const dragEnter = new Event('dragenter', { cancelable: true });
		Object.defineProperty(dragEnter, 'dataTransfer', { value: { types: ['text/plain'] } });
		await fireEvent(window, dragEnter);

		expect(screen.queryByText('Drop an album anywhere on this page', { selector: 'p.text-primary' })).not.toBeInTheDocument();
	});
});
