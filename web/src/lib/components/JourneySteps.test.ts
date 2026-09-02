import { render, screen } from '@testing-library/svelte';
import { describe, expect, it } from 'vitest';
import JourneySteps, { type StepStatus } from './JourneySteps.svelte';

describe('JourneySteps', () => {
	it('renders the four contract labels and the track-count sublabel', () => {
		render(JourneySteps, {
			totalFiles: 12,
			statuses: ['current', 'upcoming', 'upcoming', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		expect(screen.getByText('Uploaded')).toBeInTheDocument();
		expect(screen.getByText('all 12 tracks received')).toBeInTheDocument();
		expect(screen.getByText('Checked')).toBeInTheDocument();
		expect(screen.getByText('Matching')).toBeInTheDocument();
		expect(screen.getByText('In your library')).toBeInTheDocument();
	});

	it('uses the singular track count for a one-file album', () => {
		render(JourneySteps, {
			totalFiles: 1,
			statuses: ['current', 'upcoming', 'upcoming', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		expect(screen.getByText('all 1 track received')).toBeInTheDocument();
	});

	it('marks exactly the current step aria-current="step"', () => {
		render(JourneySteps, {
			totalFiles: 3,
			statuses: ['complete', 'current', 'upcoming', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		const current = screen.getByText('Checked').closest('li');
		expect(current).toHaveAttribute('aria-current', 'step');
		const notCurrent = screen.getByText('Uploaded').closest('li');
		expect(notCurrent).not.toHaveAttribute('aria-current');
	});

	it("shows step 4's pending sublabel before it is current, and the current sublabel once it is", () => {
		const { unmount } = render(JourneySteps, {
			totalFiles: 3,
			statuses: ['complete', 'complete', 'complete', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		expect(screen.getByText('ready to play')).toBeInTheDocument();
		unmount();

		render(JourneySteps, {
			totalFiles: 3,
			statuses: ['complete', 'complete', 'complete', 'current'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		expect(screen.getByText('adding it now')).toBeInTheDocument();
	});

	it('sets a check-mark data-content on a complete step and an exclamation on an errored one', () => {
		render(JourneySteps, {
			totalFiles: 3,
			statuses: ['complete', 'error', 'upcoming', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		const complete = screen.getByText('Uploaded').closest('li');
		const errored = screen.getByText('Checked').closest('li');
		expect(complete).toHaveAttribute('data-content', '✓');
		expect(complete).toHaveClass('step-success');
		expect(errored).toHaveAttribute('data-content', '!');
		expect(errored).toHaveClass('step-error');
	});

	it('renders the vertical spine at every viewport, with no responsive switch to horizontal mode (U13 fold escalation)', () => {
		render(JourneySteps, {
			totalFiles: 3,
			statuses: ['current', 'upcoming', 'upcoming', 'upcoming'] as [StepStatus, StepStatus, StepStatus, StepStatus],
		});
		const list = screen.getByRole('list');
		expect(list).toHaveClass('steps', 'steps-vertical');
		expect(list.className).not.toMatch(/steps-horizontal/);
	});
});
