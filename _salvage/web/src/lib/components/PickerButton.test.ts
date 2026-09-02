import { render, screen } from '@testing-library/svelte';
import { fireEvent } from '@testing-library/dom';
import { describe, expect, it, vi } from 'vitest';
import PickerButton from './PickerButton.svelte';

describe('PickerButton', () => {
	it('renders the given label and an extension-first accept list, never bare audio/*', () => {
		render(PickerButton, { label: 'Choose files', accept: '.zip,.flac', onFiles: vi.fn() });
		const input = screen.getByLabelText('Choose files', { selector: 'input' }) as HTMLInputElement;
		expect(input.accept).toBe('.zip,.flac');
		expect(input.accept).not.toContain('audio/');
	});

	it('calls onFiles with the picked files and resets the input', async () => {
		const onFiles = vi.fn();
		render(PickerButton, { label: 'Choose files', accept: '.flac', onFiles });
		const input = screen.getByLabelText('Choose files', { selector: 'input' }) as HTMLInputElement;
		const file = new File([], 'track01.flac');

		Object.defineProperty(input, 'files', { value: [file], configurable: true });
		await fireEvent.change(input);

		expect(onFiles).toHaveBeenCalledWith([file]);
		expect(input.value).toBe('');
	});

	it('sets webkitdirectory only when asked', () => {
		render(PickerButton, { label: 'Choose a folder', accept: '', webkitdirectory: true, onFiles: vi.fn() });
		const input = screen.getByLabelText('Choose a folder', { selector: 'input' }) as HTMLInputElement;
		expect(input.webkitdirectory).toBe(true);
	});

	it('is operable from the keyboard: the input stays in the tab order despite being visually hidden', () => {
		render(PickerButton, { label: 'Choose files', accept: '.flac', onFiles: vi.fn() });
		const input = screen.getByLabelText('Choose files', { selector: 'input' });
		expect(input).not.toHaveAttribute('tabindex', '-1');
		expect(input).toHaveClass('opacity-0');
	});
});
