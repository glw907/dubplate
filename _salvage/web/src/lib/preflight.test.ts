import { describe, expect, it } from 'vitest';
import { badExtensionReason, isZipSelection, screenSelection } from './preflight';

function testFile(name: string): File {
	return new File([], name);
}

describe('isZipSelection', () => {
	it('reports true for a single .zip file', () => {
		expect(isZipSelection([testFile('Album.zip')])).toBe(true);
	});

	it('reports false for multiple files, even if one is a .zip', () => {
		expect(isZipSelection([testFile('Album.zip'), testFile('track01.flac')])).toBe(false);
	});

	it('reports false for a single non-zip file', () => {
		expect(isZipSelection([testFile('track01.flac')])).toBe(false);
	});
});

describe('screenSelection', () => {
	it('passes FLAC files', () => {
		const [verdict] = screenSelection([testFile('track01.flac')]);
		expect(verdict.reason).toBeNull();
	});

	it('flags a lossy audio extension with the ReasonBadExtension mirror', () => {
		const [verdict] = screenSelection([testFile('track01.mp3')]);
		expect(verdict.reason).toEqual(badExtensionReason);
	});

	it('lets a non-audio extra ride along unflagged', () => {
		const [verdict] = screenSelection([testFile('cover.jpg')]);
		expect(verdict.reason).toBeNull();
	});

	it('screens a mixed batch file by file', () => {
		const verdicts = screenSelection([testFile('track01.flac'), testFile('track02.mp3'), testFile('booklet.pdf')]);
		expect(verdicts.map((v) => v.reason === null)).toEqual([true, false, true]);
	});

	it('is case-insensitive on extension', () => {
		const [verdict] = screenSelection([testFile('track01.MP3')]);
		expect(verdict.reason).toEqual(badExtensionReason);
	});
});
