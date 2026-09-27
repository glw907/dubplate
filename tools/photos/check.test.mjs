// Table-driven tests for check.mjs's four modes. Each failing case builds a
// fixture from a known-good base and mutates exactly one thing; the mutation
// is named in the case so a reviewer can see what one-line change would make
// it pass again.

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, copyFileSync, readFileSync, existsSync, readdirSync } from 'node:fs';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';

const HERE = path.dirname(fileURLToPath(import.meta.url));
const CHECK = path.join(HERE, 'check.mjs');
const REPO_ROOT = path.resolve(HERE, '..', '..');
const REAL_VOCAB = path.join(REPO_ROOT, 'photos', 'meta', 'vocab.json');
const BASE_PICKED = JSON.parse(readFileSync(path.join(HERE, 'testdata', 'base-picked.json'), 'utf8'));
const BASE_CANDIDATE = JSON.parse(readFileSync(path.join(HERE, 'testdata', 'base-candidate.json'), 'utf8'));

function sha256(buf) {
  return createHash('sha256').update(buf).digest('hex');
}

function run(args, cwd) {
  const result = spawnSync(process.execPath, [CHECK, ...args], { cwd, encoding: 'utf8' });
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

function newRoot() {
  const root = mkdtempSync(path.join(tmpdir(), 'photos-check-'));
  mkdirSync(path.join(root, 'meta'), { recursive: true });
  copyFileSync(REAL_VOCAB, path.join(root, 'meta', 'vocab.json'));
  writeFileSync(path.join(root, 'meta', 'waivers.json'), '[]\n');
  writeFileSync(path.join(root, 'meta', 'rejected.json'), '[]\n');
  writeFileSync(path.join(root, 'meta', 'pack.json'), '[]\n');
  return root;
}

// A known-good picked record, loaded from testdata/base-picked.json.
// `overrides` is merged shallowly; nested objects that need a partial
// change should be spread by the caller.
function basePickedRecord(overrides = {}) {
  return { ...BASE_PICKED, ...overrides };
}

// Writes a picked record and its .webp sibling, with `sha256` computed from
// the actual bytes so the pair is hash-consistent unless the case asks for
// a mismatch. `source_sha256` defaults to a hash unique to `slug`, so two
// records in the same fixture never collide on it by accident; a case that
// wants a real collision passes a shared `source_sha256` override.
function writePicked(root, slug, overrides = {}, { webpBytes, sha256Override } = {}) {
  const bytes = webpBytes || Buffer.from(`webp-bytes-for-${slug}`);
  writeFileSync(path.join(root, `${slug}.webp`), bytes);
  const record = basePickedRecord({
    source_sha256: sha256(Buffer.from(`source-bytes-for-${slug}`)),
    sha256: sha256Override || sha256(bytes),
    ...overrides,
  });
  writeFileSync(path.join(root, `${slug}.json`), `${JSON.stringify(record, null, 2)}\n`);
  return record;
}

// Writes a picked-shaped record with no .webp and no hash consistency, for
// balance-only fixtures that never run through integrity's hash check.
function writeBalanceOnly(root, slug, overrides = {}) {
  const record = basePickedRecord({ source_sha256: sha256(Buffer.from(slug)), sha256: sha256(Buffer.from(slug)), ...overrides });
  writeFileSync(path.join(root, `${slug}.json`), JSON.stringify(record));
}

// Writes `count` balance-only picked records under one `prefix`, each with
// a distinct source_id, for share-band and floor-grid pool construction.
// Defaults to no scene (family "other") and no place/landmark, since the
// base fixture's own defaults (reggae, a landmark) would otherwise leak
// into every filled record unless the caller overrides them back.
function fillPicked(root, count, prefix, overrides = {}) {
  for (let i = 0; i < count; i++) {
    writeBalanceOnly(root, `${prefix}-${i}`, {
      source_id: `commons:${prefix}-${i}`,
      origin_id: undefined,
      artists: [],
      scene: undefined,
      place: undefined,
      ...overrides,
    });
  }
}

function baseWithdrawnRecord(overrides = {}) {
  return basePickedRecord({
    status: 'withdrawn',
    withdrawn: { reason: 'License disputed after publication.', date: '2026-01-01' },
    ...overrides,
  });
}

function basePackEntry(overrides = {}) {
  const { screen, ...rest } = BASE_CANDIDATE;
  return {
    ...rest,
    source_sha256: sha256(Buffer.from('pack-entry-source-bytes')),
    license: 'CC-BY-NC-4.0', // a pack.json entry's license is always a pack class (Storage).
    status: 'pack',
    ...overrides,
  };
}

// A known-good candidate record, loaded from testdata/base-candidate.json.
function baseCandidateRecord(overrides = {}) {
  return { ...BASE_CANDIDATE, ...overrides };
}

// Writes a candidate record plus its source image, thumbnail, and preview,
// under `dir/<id>.json` etc. `skip` names files to omit, for the missing
// cases.
function writeCandidate(dir, id, overrides = {}, { skip = [] } = {}) {
  mkdirSync(dir, { recursive: true });
  const sourceBytes = Buffer.from(`source-bytes-for-${id}`);
  // source_id defaults to the filename id's reverse (the dash-for-colon
  // spelling undone), so a caller that only wants a distinct filename
  // never has to also restate source_id to keep the two consistent.
  const record = baseCandidateRecord({ source_id: id.replace('-', ':'), source_sha256: sha256(sourceBytes), ...overrides });
  if (!skip.includes('source')) writeFileSync(path.join(dir, `${id}.jpg`), sourceBytes);
  if (!skip.includes('thumb')) writeFileSync(path.join(dir, `${id}.thumb.webp`), Buffer.from('thumb'));
  if (!skip.includes('preview')) writeFileSync(path.join(dir, `${id}.preview.webp`), Buffer.from('preview'));
  if (!skip.includes('json')) writeFileSync(path.join(dir, `${id}.json`), `${JSON.stringify(record, null, 2)}\n`);
  return record;
}

test('integrity: a fully populated passing pool exercises every kind of record and exits 0', () => {
  const root = newRoot();
  writePicked(root, 'dischord-house-1984');

  // A public-domain record.
  writePicked(root, 'pd-record', {
    source_id: 'commons:2001',
    origin_id: undefined,
    artists: [{ name: 'Unrelated Artist' }],
    place: undefined,
    license: 'LicenseRef-public-domain',
    public_domain: {
      basis: 'pre-1931',
      jurisdiction: 'US',
      us_status: 'public domain',
      fact: '1920 first publication, no notice.',
    },
  });

  // A flyer whose own artwork is separately public domain.
  writePicked(root, 'pd-artwork-flyer', {
    source_id: 'commons:2002',
    origin_id: undefined,
    artists: [],
    place: undefined,
    format: 'flyer',
    artwork: {
      artist: 'The Flyer Artist',
      license: 'LicenseRef-public-domain',
      credit: 'The Flyer Artist',
      source_url: 'https://example.org/flyer-artwork',
      verified: "Quoted from the archive's own catalog entry.",
      public_domain: {
        basis: 'us-government-work',
        jurisdiction: 'US',
        us_status: 'public domain',
        fact: 'A US federal government work.',
      },
    },
  });

  // A size_exception record that reuses the base record's artist, so the
  // artist-cap waiver below has a real breach to cover.
  writePicked(root, 'size-exception-record', {
    source_id: 'commons:2003',
    origin_id: undefined,
    place: undefined,
    width: 1100,
    height: 750,
    size_exception: 'The only known photograph of this session.',
  });

  // A withdrawn record: no .webp, but the record itself still validates.
  writeFileSync(
    path.join(root, 'withdrawn-record.json'),
    JSON.stringify(
      baseWithdrawnRecord({
        source_id: 'commons:2004',
        origin_id: undefined,
        artists: [],
        place: undefined,
      }),
    ),
  );

  // A pack.json entry, with a pack-class license.
  writeFileSync(
    path.join(root, 'meta', 'pack.json'),
    JSON.stringify([basePackEntry({ source_id: 'commons:2005', license: 'CC-BY-NC-4.0' })]),
  );

  // A rejected.json entry.
  writeFileSync(path.join(root, 'meta', 'rejected.json'), JSON.stringify([{ source_id: 'commons:2006', reason: 'passed' }]));

  // A waiver covering size-exception-record's shared artist.
  writeFileSync(
    path.join(root, 'meta', 'waivers.json'),
    JSON.stringify([{ rule: 'artist', key: 'name:test sound system', reason: 'Geoff, round 1' }]),
  );

  mkdirSync(path.join(root, '.candidates', 'a'), { recursive: true });
  writeCandidate(path.join(root, '.candidates', 'a'), 'commons-99999');
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('integrity: the three exit codes', () => {
  const passRoot = newRoot();
  writePicked(passRoot, 'dischord-house-1984');
  assert.equal(run(['integrity', '--root', passRoot]).status, 0);

  const failRoot = newRoot();
  writePicked(failRoot, 'too-small-one', {}, { webpBytes: Buffer.from('x') });
  writeFileSync(
    path.join(failRoot, 'too-small-one.json'),
    JSON.stringify(basePickedRecord({ width: 900, height: 600, sha256: sha256(Buffer.from('x')) })),
  );
  assert.equal(run(['integrity', '--root', failRoot]).status, 1);

  const noVocabRoot = mkdtempSync(path.join(tmpdir(), 'photos-check-'));
  assert.equal(run(['integrity', '--root', noVocabRoot]).status, 2);

  for (const r of [passRoot, failRoot, noVocabRoot]) rmSync(r, { recursive: true, force: true });
});

test('integrity: the default root resolves from the script location, not the caller cwd', () => {
  // Run from a directory with no photos/ at all: if the default were the
  // relative string "photos", this would miss and exit 2.
  const { status } = run(['integrity'], tmpdir());
  assert.equal(status, 0);
});

// Each case starts from the passing fixture above and applies exactly one
// mutation. The comment on each case is the one-line change that flips the
// case back to passing.
const integrityCases = [
  {
    name: 'a duplicate artist without a waiver',
    // Flip: give the second record a different artist name.
    setup(root) {
      writePicked(root, 'record-a', { source_id: 'commons:1', origin_id: undefined });
      writePicked(root, 'record-b', { source_id: 'commons:2', origin_id: undefined });
    },
    wantRule: 'artist-cap',
  },
  {
    name: 'a duplicate landmark without a waiver',
    // Flip: give landmark-b a different landmark.
    setup(root) {
      writePicked(root, 'landmark-a', {
        source_id: 'commons:10',
        origin_id: undefined,
        artists: [{ name: 'Band A' }],
        place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q999' } },
      });
      writePicked(root, 'landmark-b', {
        source_id: 'commons:11',
        origin_id: undefined,
        artists: [{ name: 'Band B' }],
        place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q999' } },
      });
    },
    wantRule: 'landmark-cap',
  },
  {
    name: 'an NC-licensed pick',
    // Flip: license: "CC-BY-2.0".
    setup(root) {
      writePicked(root, 'nc-pick', { license: 'CC-BY-NC-4.0' });
    },
    wantRule: 'license-class',
  },
  {
    name: 'an NC-licensed flyer artwork under a CC BY scan',
    // Flip: artwork.license: "CC-BY-2.0".
    setup(root) {
      writePicked(root, 'nc-flyer', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-NC-4.0',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-license',
  },
  {
    name: 'a flyer with no artwork at all',
    // Flip: format: "photograph", or add an artwork object.
    setup(root) {
      writePicked(root, 'flyer-no-artwork', { format: 'flyer' });
    },
    wantRule: 'artwork-required',
  },
  {
    name: 'artwork set on a photograph',
    // Flip: remove artwork, or set format to "flyer".
    setup(root) {
      writePicked(root, 'photo-with-artwork', {
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-2.0',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-forbidden',
  },
  {
    name: 'a flyer artwork missing credit',
    // Flip: add credit: "..." to artwork.
    setup(root) {
      writePicked(root, 'flyer-no-credit', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-2.0',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-credit',
  },
  {
    name: 'a Commons record and its Flickr original sharing origin_id',
    // Flip: give commons-copy a different origin_id.
    setup(root) {
      writePicked(root, 'commons-copy', {
        source_id: 'commons:1',
        origin_id: 'flickr:5',
        artists: [{ name: 'Commons Copy Artist' }],
        place: undefined,
      });
      writePicked(root, 'flickr-origin', {
        source_id: 'flickr:5',
        origin_id: undefined,
        artists: [{ name: 'Flickr Origin Artist' }],
        place: undefined,
      });
    },
    wantRule: 'duplicate-key',
    // source_id and origin_id share one identity key space ("id"), so the
    // message names the shared value regardless of which field carried it.
    wantMessage: /id "flickr:5" already used by/,
  },
  {
    name: 'a record whose JSON fails to parse',
    // Flip: fix the trailing comma.
    setup(root) {
      writeFileSync(path.join(root, 'unparsable.webp'), Buffer.from('x'));
      writeFileSync(path.join(root, 'unparsable.json'), '{ "status": "picked", }');
    },
    wantRule: 'parse-error',
  },
  {
    name: 'a .webp with no record',
    // Flip: delete the stray .webp, or add a matching picked record.
    setup(root) {
      writePicked(root, 'has-record');
      writeFileSync(path.join(root, 'orphan.webp'), Buffer.from('x'));
    },
    wantRule: 'orphan-webp',
  },
  {
    name: 'a picked record with no .webp file',
    // Flip: add the .webp file.
    setup(root) {
      writeFileSync(path.join(root, 'no-webp.json'), JSON.stringify(basePickedRecord()));
    },
    wantRule: 'missing-webp',
  },
  {
    name: 'a picked record at 1100px without size_exception',
    // Flip: add size_exception: "..." to the record.
    setup(root) {
      writePicked(root, 'edge-1100', { width: 1100, height: 750 });
    },
    wantRule: 'size-exception-required',
  },
  {
    name: 'a picked record at 900px',
    // Flip: width/height back to 1600x1067.
    setup(root) {
      writePicked(root, 'edge-900', { width: 900, height: 600 });
    },
    wantRule: 'size-floor',
  },
  {
    name: 'size_exception set at or above the 1200px floor',
    // Flip: remove size_exception.
    setup(root) {
      writePicked(root, 'unneeded-exception', { size_exception: 'Not actually needed at this size.' });
    },
    wantRule: 'size-exception-forbidden',
  },
  {
    name: 'a picked record with non-integer width',
    // Flip: width: 1600.
    setup(root) {
      writePicked(root, 'bad-width', { width: undefined });
    },
    wantRule: 'dimensions',
  },
  {
    name: 'a picked record with a malformed sha256',
    // Flip: remove the override, so sha256 is computed from the .webp bytes.
    setup(root) {
      writePicked(root, 'bad-sha-format', { sha256: 'not-a-valid-hash' });
    },
    wantRule: 'sha256',
  },
  {
    name: 'year.value is not an integer',
    // Flip: year: { value: 1978, precision: "circa" }.
    setup(root) {
      writePicked(root, 'bad-year-value', { year: { value: 'nineteen-seventy-eight', precision: 'circa' } });
    },
    wantRule: 'year-value',
  },
  {
    name: 'year.precision is not in vocab.json',
    // Flip: precision: "circa".
    setup(root) {
      writePicked(root, 'bad-year-precision', { year: { value: 1978, precision: 'ish' } });
    },
    wantRule: 'year-precision',
  },
  {
    name: 'a public_domain record on an unlisted base',
    // Flip: public_domain.basis: "pre-1931".
    setup(root) {
      writePicked(root, 'pd-bad-basis', {
        license: 'LicenseRef-public-domain',
        public_domain: { basis: 'not-a-real-basis', jurisdiction: 'US', us_status: 'public domain', fact: '1920 first publication' },
      });
    },
    wantRule: 'public-domain-basis',
  },
  {
    name: 'a public-domain license with no public_domain object',
    // Flip: add a valid public_domain object.
    setup(root) {
      writePicked(root, 'pd-missing', { license: 'LicenseRef-public-domain' });
    },
    wantRule: 'public-domain-required',
  },
  {
    name: 'public_domain set on a non-public-domain license',
    // Flip: remove public_domain, or set license to LicenseRef-public-domain.
    setup(root) {
      writePicked(root, 'pd-forbidden', {
        public_domain: { basis: 'pre-1931', jurisdiction: 'US', us_status: 'public domain', fact: '1920 first publication' },
      });
    },
    wantRule: 'public-domain-forbidden',
  },
  {
    name: 'public_domain missing jurisdiction',
    // Flip: add jurisdiction: "US".
    setup(root) {
      writePicked(root, 'pd-no-jurisdiction', {
        license: 'LicenseRef-public-domain',
        public_domain: { basis: 'pre-1931', us_status: 'public domain', fact: '1920 first publication' },
      });
    },
    wantRule: 'public-domain-jurisdiction',
  },
  {
    name: 'public_domain missing fact',
    // Flip: add fact: "...".
    setup(root) {
      writePicked(root, 'pd-no-fact', {
        license: 'LicenseRef-public-domain',
        public_domain: { basis: 'pre-1931', jurisdiction: 'US', us_status: 'public domain' },
      });
    },
    wantRule: 'public-domain-fact',
  },
  {
    name: 'public_domain us_status does not match vocab.json',
    // Flip: us_status: "public domain" (vocab.json's public_domain_us_status).
    setup(root) {
      writePicked(root, 'pd-wrong-us-status', {
        license: 'LicenseRef-public-domain',
        public_domain: { basis: 'pre-1931', jurisdiction: 'US', us_status: 'probably public domain', fact: '1920 first publication' },
      });
    },
    wantRule: 'public-domain-us-status',
  },
  {
    name: 'a hash mismatch',
    // Flip: recompute sha256 from the actual .webp bytes.
    setup(root) {
      writePicked(root, 'bad-hash', {}, { sha256Override: 'b'.repeat(64) });
    },
    wantRule: 'hash-mismatch',
  },
  {
    name: 'a bad source_id: dash instead of colon',
    // Flip: source_id: "commons:12345".
    setup(root) {
      writePicked(root, 'bad-source-id', { source_id: 'commons-12345' });
    },
    wantRule: 'source-id',
  },
  {
    name: 'a bad source_id: non-digit commons key',
    // Flip: source_id: "commons:12345".
    setup(root) {
      writePicked(root, 'bad-source-id-commons-key', { source_id: 'commons:12345a' });
    },
    wantRule: 'source-id',
  },
  {
    name: 'a bad source_id: non-digit flickr key',
    // Flip: source_id: "flickr:12345".
    setup(root) {
      writePicked(root, 'bad-source-id-flickr-key', { source_id: 'flickr:abc', origin_id: undefined });
    },
    wantRule: 'source-id',
  },
  {
    name: 'a bad source_id: archive with no key segment',
    // Flip: source_id: "archive:archive.org:004278".
    setup(root) {
      writePicked(root, 'bad-source-id-archive-no-key', { source_id: 'archive:archive.org', origin_id: undefined });
    },
    wantRule: 'source-id',
  },
  {
    name: 'an unknown status value',
    // Flip: status: "picked".
    setup(root) {
      writePicked(root, 'weird-status', { status: 'draft' });
    },
    wantRule: 'status',
  },
  {
    name: 'a filename that does not match the slug pattern',
    // Flip: rename Bad_Slug.json/.webp to a lowercase-dash slug.
    setup(root) {
      const bytes = Buffer.from('x');
      writeFileSync(path.join(root, 'Bad_Slug.webp'), bytes);
      writeFileSync(path.join(root, 'Bad_Slug.json'), JSON.stringify(basePickedRecord({ sha256: sha256(bytes) })));
    },
    wantRule: 'slug-pattern',
  },
  {
    name: 'an unknown scene',
    // Flip: scene.primary: "reggae".
    setup(root) {
      writePicked(root, 'bad-scene', { scene: { primary: 'not-a-real-scene' } });
    },
    wantRule: 'vocab:scene',
  },
  {
    name: 'an unknown subject',
    // Flip: subject: ["sound-systems"].
    setup(root) {
      writePicked(root, 'bad-subject', { subject: ['not-a-real-subject'] });
    },
    wantRule: 'vocab:subject',
  },
  {
    name: 'an unknown shot',
    // Flip: shot: "candid".
    setup(root) {
      writePicked(root, 'bad-shot', { shot: 'not-a-real-shot' });
    },
    wantRule: 'vocab:shot',
  },
  {
    name: 'an unknown format',
    // Flip: format: "photograph".
    setup(root) {
      writePicked(root, 'bad-format', { format: 'not-a-real-format' });
    },
    wantRule: 'vocab:format',
  },
  {
    name: 'an unknown surface',
    // Flip: surfaces: ["gallery"].
    setup(root) {
      writePicked(root, 'bad-surface', { surfaces: ['not-a-real-surface'] });
    },
    wantRule: 'vocab:surface',
  },
  {
    name: 'a country code with no region',
    // Flip: place.country_code: "JM".
    setup(root) {
      writePicked(root, 'bad-region', { place: { city: 'Nowhere', country_code: 'ZZ' } });
    },
    wantRule: 'vocab:region',
  },
  {
    name: 'an unknown license',
    // Flip: license: "CC-BY-2.0".
    setup(root) {
      writePicked(root, 'bad-license', { license: 'not-a-real-license' });
    },
    wantRule: 'vocab:license',
  },
  {
    name: 'screen still set on a picked record',
    // Flip: remove screen.
    setup(root) {
      writePicked(root, 'has-screen', { screen: { score: 5, note: 'x' } });
    },
    wantRule: 'screen-forbidden',
  },
  {
    name: 'flags still set on a picked record',
    // Flip: remove flags.
    setup(root) {
      writePicked(root, 'has-flags', { flags: ['personality-rights'] });
    },
    wantRule: 'flags-forbidden',
  },
  {
    name: 'pack still set on a picked record',
    // Flip: remove pack.
    setup(root) {
      writePicked(root, 'has-pack-field', { pack: false });
    },
    wantRule: 'pack-forbidden',
  },
  {
    name: 'a withdrawn record missing reason and date',
    // Flip: withdrawn: { reason: "...", date: "..." }.
    setup(root) {
      writeFileSync(path.join(root, 'withdrawn-a.json'), JSON.stringify(baseWithdrawnRecord({ withdrawn: {} })));
    },
    wantRule: 'withdrawn-fields',
  },
  {
    name: "a withdrawn record whose .webp still exists",
    // Flip: delete withdrawn-b.webp.
    setup(root) {
      writeFileSync(path.join(root, 'withdrawn-b.json'), JSON.stringify(baseWithdrawnRecord()));
      writeFileSync(path.join(root, 'withdrawn-b.webp'), Buffer.from('still here'));
    },
    wantRule: 'withdrawn-has-image',
  },
  {
    name: 'withdrawn set on a picked record',
    // Flip: status: "withdrawn", or remove withdrawn.
    setup(root) {
      writePicked(root, 'picked-with-withdrawn', { withdrawn: { reason: 'x', date: '2026-01-01' } });
    },
    wantRule: 'withdrawn-forbidden',
  },
  {
    name: "a rejected entry that repeats a picked record's key",
    // Flip: give the rejected.json entry a different source_id.
    setup(root) {
      writePicked(root, 'kept-one', { source_id: 'commons:777' });
      writeFileSync(
        path.join(root, 'meta', 'rejected.json'),
        JSON.stringify([{ source_id: 'commons:777', reason: 'duplicate' }]),
      );
    },
    wantRule: 'duplicate-key',
  },
  {
    name: 'a rejected entry with no reason',
    // Flip: add reason: "passed" (or any vocab.json reason).
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'rejected.json'), JSON.stringify([{ source_id: 'commons:500' }]));
    },
    wantRule: 'reason-required',
  },
  {
    name: 'a rejected entry with an unknown reason',
    // Flip: reason: "passed".
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'rejected.json'), JSON.stringify([{ source_id: 'commons:501', reason: 'not-a-real-reason' }]));
    },
    wantRule: 'vocab:reason',
  },
  {
    name: 'a pack.json entry with the wrong status',
    // Flip: status: "pack".
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'pack.json'), JSON.stringify([basePackEntry({ status: 'picked' })]));
    },
    wantRule: 'status',
  },
  {
    name: 'a pack.json entry with withdrawn set',
    // Flip: remove withdrawn.
    setup(root) {
      writeFileSync(
        path.join(root, 'meta', 'pack.json'),
        JSON.stringify([basePackEntry({ withdrawn: { reason: 'x', date: '2026-01-01' } })]),
      );
    },
    wantRule: 'withdrawn-forbidden',
  },
  {
    name: 'a pack.json entry that still carries screen',
    // Flip: remove screen.
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'pack.json'), JSON.stringify([basePackEntry({ screen: { score: 3, note: 'x' } })]));
    },
    wantRule: 'screen-forbidden',
  },
  {
    name: 'a waiver on an unknown rule',
    // Flip: rule: "artist" (a rule waivers.json actually allows).
    setup(root) {
      writePicked(root, 'record-a', { source_id: 'commons:1', origin_id: undefined });
      writePicked(root, 'record-b', { source_id: 'commons:2', origin_id: undefined });
      writeFileSync(
        path.join(root, 'meta', 'waivers.json'),
        JSON.stringify([{ rule: 'not-a-real-rule', key: 'name:test sound system', reason: 'Geoff, round 1' }]),
      );
    },
    wantRule: 'waiver-rule',
  },
  {
    name: 'the same waiver with the wrong key',
    // Flip: key: "name:test sound system" (the artist's normalized key).
    setup(root) {
      writePicked(root, 'record-a', { source_id: 'commons:1', origin_id: undefined });
      writePicked(root, 'record-b', { source_id: 'commons:2', origin_id: undefined });
      writeFileSync(
        path.join(root, 'meta', 'waivers.json'),
        JSON.stringify([{ rule: 'artist', key: 'name:the wrong band', reason: 'Geoff, round 1' }]),
      );
    },
    wantRule: 'artist-cap',
  },
  {
    name: 'a waiver missing key',
    // Flip: add key: "name:some band".
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'waivers.json'), JSON.stringify([{ rule: 'artist', reason: 'Geoff, round 1' }]));
    },
    wantRule: 'waiver-key',
  },
  {
    name: 'a waiver missing reason',
    // Flip: add reason: "Geoff, round 1".
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'waivers.json'), JSON.stringify([{ rule: 'artist', key: 'name:some band' }]));
    },
    wantRule: 'waiver-reason',
  },
  {
    name: 'a flyer artwork with a public-domain license but no public_domain object',
    // Flip: add a valid public_domain object to artwork.
    setup(root) {
      writePicked(root, 'flyer-pd-artwork-missing', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'LicenseRef-public-domain',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the archive's catalog.",
        },
      });
    },
    wantRule: 'artwork-public-domain-required',
  },
  {
    name: 'a flyer artwork missing artist',
    // Flip: add artist: "The Artist".
    setup(root) {
      writePicked(root, 'flyer-no-artist', {
        format: 'flyer',
        artwork: {
          license: 'CC-BY-2.0',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-artist',
  },
  {
    name: 'a flyer artwork missing source_url',
    // Flip: add source_url: "https://example.org/flyer".
    setup(root) {
      writePicked(root, 'flyer-no-source-url', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-2.0',
          credit: 'The Artist',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-source-url',
  },
  {
    name: 'a flyer artwork missing verified',
    // Flip: add verified: "Quoted...".
    setup(root) {
      writePicked(root, 'flyer-no-verified', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-2.0',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
        },
      });
    },
    wantRule: 'artwork-verified',
  },
  {
    name: 'a flyer artwork with an unknown license',
    // Flip: license: "CC-BY-2.0".
    setup(root) {
      writePicked(root, 'flyer-bad-artwork-license', {
        format: 'flyer',
        artwork: {
          artist: 'The Artist',
          license: 'not-a-real-license',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'artwork-vocab-license',
  },
  {
    name: "a rejected entry sharing only a picked record's source_sha256",
    // Flip: give the rejected.json entry a different source_sha256.
    setup(root) {
      writePicked(root, 'hash-share-record', { source_id: 'commons:3001', origin_id: undefined });
      writeFileSync(
        path.join(root, 'meta', 'rejected.json'),
        JSON.stringify([
          {
            source_id: 'commons:3002',
            source_sha256: sha256(Buffer.from('source-bytes-for-hash-share-record')),
            reason: 'passed',
          },
        ]),
      );
    },
    wantRule: 'duplicate-key',
    wantMessage: /hash ".*" already used by/,
  },
  {
    name: 'a picked record missing title',
    // Flip: add title: "...".
    setup(root) {
      writePicked(root, 'no-title', { title: undefined });
    },
    wantRule: 'title',
  },
  {
    name: 'a picked record missing photographer',
    // Flip: add photographer: "...".
    setup(root) {
      writePicked(root, 'no-photographer', { photographer: undefined });
    },
    wantRule: 'photographer',
  },
  {
    name: 'a picked record missing source_url',
    // Flip: add source_url: "...".
    setup(root) {
      writePicked(root, 'no-source-url', { source_url: undefined });
    },
    wantRule: 'source-url',
  },
  {
    name: 'a picked record missing file_url',
    // Flip: add file_url: "...".
    setup(root) {
      writePicked(root, 'no-file-url', { file_url: undefined });
    },
    wantRule: 'file-url',
  },
  {
    name: 'a picked record missing retrieved_at',
    // Flip: add retrieved_at: "...".
    setup(root) {
      writePicked(root, 'no-retrieved-at', { retrieved_at: undefined });
    },
    wantRule: 'retrieved-at',
  },
  {
    name: 'a picked record missing verified',
    // Flip: add verified: "...".
    setup(root) {
      writePicked(root, 'no-verified', { verified: undefined });
    },
    wantRule: 'verified',
  },
  {
    name: 'a picked record with a malformed source_sha256',
    // Flip: source_sha256: a 64-hex string.
    setup(root) {
      writePicked(root, 'bad-source-sha256', { source_sha256: 'not-a-hash' });
    },
    wantRule: 'source-sha256',
  },
  {
    name: 'a picked record with an empty subject list',
    // Flip: subject: ["sound-systems"].
    setup(root) {
      writePicked(root, 'empty-subject', { subject: [] });
    },
    wantRule: 'subject',
  },
  {
    name: 'a picked record missing year',
    // Flip: add year: { value: 1978, precision: "circa" }.
    setup(root) {
      writePicked(root, 'no-year', { year: undefined });
    },
    wantRule: 'year',
  },
  {
    name: 'a landmark missing name',
    // Flip: add name: "..." to the landmark.
    setup(root) {
      writePicked(root, 'landmark-no-name', {
        place: { city: 'Kingston', country_code: 'JM', landmark: { wikidata_id: 'Q1' } },
      });
    },
    wantRule: 'landmark',
  },
  {
    name: 'a pack.json entry with an unknown scene',
    // Flip: scene.primary: "reggae".
    setup(root) {
      writeFileSync(
        path.join(root, 'meta', 'pack.json'),
        JSON.stringify([basePackEntry({ scene: { primary: 'not-a-real-scene' } })]),
      );
    },
    wantRule: 'vocab:scene',
  },
  {
    name: 'a pack.json entry with a shipped (CC BY) license',
    // Flip: license: "CC-BY-NC-4.0" (a pack class).
    setup(root) {
      writeFileSync(path.join(root, 'meta', 'pack.json'), JSON.stringify([basePackEntry({ license: 'CC-BY-2.0' })]));
    },
    wantRule: 'pack-license',
  },
  {
    name: 'a focus missing y',
    // Flip: focus: { x: 50, y: 50 }.
    setup(root) {
      writePicked(root, 'focus-missing-y', { focus: { x: 50 } });
    },
    wantRule: 'crop-focus',
  },
  {
    name: 'a focus with x out of range',
    // Flip: focus: { x: 100, y: 50 }.
    setup(root) {
      writePicked(root, 'focus-out-of-range', { focus: { x: 101, y: 50 } });
    },
    wantRule: 'crop-focus',
  },
  {
    name: 'a focus with a non-integer y',
    // Flip: focus: { x: 50, y: 50 }.
    setup(root) {
      writePicked(root, 'focus-non-integer', { focus: { x: 50, y: 50.5 } });
    },
    wantRule: 'crop-focus',
  },
  {
    name: 'a focus with an extra key',
    // Flip: remove the z key.
    setup(root) {
      writePicked(root, 'focus-extra-key', { focus: { x: 50, y: 50, z: 0 } });
    },
    wantRule: 'crop-focus',
  },
  {
    name: 'a fit that is not "cover" or "contain"',
    // Flip: fit: "cover".
    setup(root) {
      writePicked(root, 'fit-bad-string', { fit: 'stretch' });
    },
    wantRule: 'crop-fit',
  },
  {
    name: 'a flyer with fit: "cover"',
    // Flip: fit: "contain", or drop fit entirely.
    setup(root) {
      writePicked(root, 'flyer-with-cover-fit', {
        format: 'flyer',
        fit: 'cover',
        artwork: {
          artist: 'The Artist',
          license: 'CC-BY-2.0',
          credit: 'The Artist',
          source_url: 'https://example.org/flyer',
          verified: "Quoted from the artist's own site.",
        },
      });
    },
    wantRule: 'crop-fit-forbidden',
  },
];

for (const c of integrityCases) {
  test(`integrity: ${c.name}`, () => {
    const root = newRoot();
    c.setup(root);
    const { status, stdout } = run(['integrity', '--root', root]);
    assert.equal(status, 1, stdout);
    assert.match(stdout, new RegExp(`: ${c.wantRule}: `), stdout);
    if (c.wantMessage) assert.match(stdout, c.wantMessage, stdout);
    rmSync(root, { recursive: true, force: true });
  });
}

test('integrity: a matching waiver passes the same breach the mismatched key failed', () => {
  const root = newRoot();
  writePicked(root, 'record-a', { source_id: 'commons:1', origin_id: undefined });
  writePicked(root, 'record-b', {
    source_id: 'commons:2',
    origin_id: undefined,
    place: { city: 'Kingston', country_code: 'JM' },
  });
  writeFileSync(
    path.join(root, 'meta', 'waivers.json'),
    JSON.stringify([{ rule: 'artist', key: 'name:test sound system', reason: 'Geoff, round 1' }]),
  );
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('integrity: two artists with different MBIDs and the same name do not collide', () => {
  const root = newRoot();
  writePicked(root, 'record-a', {
    source_id: 'commons:1',
    origin_id: undefined,
    artists: [{ name: 'Bob Marley', mbid: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa' }],
  });
  writePicked(root, 'record-b', {
    source_id: 'commons:2',
    origin_id: undefined,
    artists: [{ name: 'Bob Marley', mbid: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb' }],
    place: undefined,
  });
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('integrity: a dotted archive source_id (archive.org, digitaltmuseum.se) passes', () => {
  const root = newRoot();
  writePicked(root, 'archive-dotted-name', {
    source_id: 'archive:archive.org:004278',
    origin_id: undefined,
    artists: [],
    place: undefined,
  });
  writePicked(root, 'archive-dotted-name-two', {
    source_id: 'archive:digitaltmuseum.se:021016187835',
    origin_id: undefined,
    artists: [],
    place: undefined,
  });
  writePicked(root, 'archive-dotted-name-underscore-key', {
    source_id: 'archive:archive.org:calauem_004278',
    origin_id: undefined,
    artists: [],
    place: undefined,
  });
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('integrity: a valid focus and fit pass, and their absence passes', () => {
  const root = newRoot();
  writePicked(root, 'focus-and-fit-set', { focus: { x: 40, y: 60 }, fit: 'contain' });
  writePicked(root, 'no-focus-or-fit', {
    source_id: 'commons:9001',
    origin_id: undefined,
    artists: [{ name: 'Another Band' }],
    place: undefined,
  });
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('integrity: two landmarks with different Wikidata IDs and the same name do not collide', () => {
  const root = newRoot();
  writePicked(root, 'record-a', {
    source_id: 'commons:1',
    origin_id: undefined,
    artists: [],
    place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q111' } },
  });
  writePicked(root, 'record-b', {
    source_id: 'commons:2',
    origin_id: undefined,
    artists: [],
    place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q222' } },
  });
  const { status, stdout } = run(['integrity', '--root', root]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

// --- balance -------------------------------------------------------------

test('balance: on the real, empty photos/ every share is 0 of 0 and short, never NaN', () => {
  const { status, stdout } = run(['balance'], REPO_ROOT);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /share: reggae: reggae is 0 of 0 \(short\)/);
  assert.doesNotMatch(stdout, /NaN/);
});

test('balance --json: on the real, empty photos/, the report is well-formed and lists every floor and surface as short', () => {
  const { status, stdout } = run(['balance', '--json'], REPO_ROOT);
  assert.equal(status, 1, stdout);
  const report = JSON.parse(stdout);
  assert.equal(report.total, 0);
  assert.equal(Object.keys(report.floor_grid).length, 15);
  assert.ok(Object.values(report.floor_grid).every((c) => c.status === 'short'));
  assert.equal(Object.keys(report.landmark_floor).length, 3);
  assert.ok(Object.values(report.landmark_floor).every((c) => c.status === 'short'));
  assert.ok(Object.values(report.surfaces).every((c) => c.status === 'short'));
});

test('balance --json: regions always lists all eight, at 0 when the pool has no places', () => {
  const root = newRoot();
  writeBalanceOnly(root, 'no-place-record', { place: undefined });
  const { stdout } = run(['balance', '--root', root, '--json']);
  const report = JSON.parse(stdout);
  assert.equal(Object.keys(report.regions).length, 8);
  assert.ok(Object.values(report.regions).every((r) => r.count === 0));
  rmSync(root, { recursive: true, force: true });
});

test('balance: exits 2 on a missing vocab.json and prints no JSON', () => {
  const root = mkdtempSync(path.join(tmpdir(), 'photos-check-'));
  const { status, stdout } = run(['balance', '--root', root, '--json']);
  assert.equal(status, 2);
  assert.equal(stdout, '');
  rmSync(root, { recursive: true, force: true });
});

test('balance: an empty floor cell reports short for that cell', () => {
  const root = newRoot();
  // Two records fill reggae/sound-systems but nothing fills reggae/clubs-and-venues.
  writePicked(root, 'floor-a', { source_id: 'commons:1', origin_id: undefined, artists: [{ name: 'Band One' }] });
  writePicked(root, 'floor-b', { source_id: 'commons:2', origin_id: undefined, artists: [{ name: 'Band Two' }] });
  const { status, stdout } = run(['balance', '--root', root]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /floor: reggae\/clubs-and-venues: reggae\/clubs-and-venues is 0 of 2 \(short\)/);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a landmark floor short of 2 in a non-trivial pool reports short', () => {
  const root = newRoot();
  fillPicked(root, 1, 'landmark-floor', {
    scene: { primary: 'reggae' },
    place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q1' } },
  });
  const { status, stdout } = run(['balance', '--root', root]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /landmark-floor: reggae: reggae landmark floor is 1 of 2 \(short\)/);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a surface floor short of 3 reports short', () => {
  const root = newRoot();
  fillPicked(root, 1, 'error-surface', { surfaces: ['error'] });
  const { status, stdout } = run(['balance', '--root', root]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /surface: error: surface error is 1 of 3 \(short\)/);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a region cap breach reports over', () => {
  const root = newRoot();
  fillPicked(root, 3, 'jm', { place: { city: 'Kingston', country_code: 'JM' } });
  fillPicked(root, 1, 'gb', { place: { city: 'London', country_code: 'GB' } });
  const { status, stdout } = run(['balance', '--root', root]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /region: caribbean: caribbean is 3 of 4 \(over the region cap\)/);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a posed ceiling breach reports over', () => {
  const root = newRoot();
  fillPicked(root, 2, 'posed', { shot: 'posed' });
  fillPicked(root, 8, 'candid', { shot: 'candid' });
  const { status, stdout } = run(['balance', '--root', root]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /posed: all: posed is 2 of 10 \(over the posed ceiling\)/);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a waived breach does not appear in the breach list', () => {
  const root = newRoot();
  fillPicked(root, 1, 'landmark-floor', {
    scene: { primary: 'reggae' },
    place: { city: 'Kingston', country_code: 'JM', landmark: { name: 'Studio One', wikidata_id: 'Q1' } },
  });
  writeFileSync(
    path.join(root, 'meta', 'waivers.json'),
    JSON.stringify([{ rule: 'landmark-floor', key: 'reggae', reason: 'Geoff, round 1' }]),
  );
  const { stdout } = run(['balance', '--root', root]);
  assert.doesNotMatch(stdout, /landmark-floor: reggae:/, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('balance: a share exactly at the +/-3pp band edge passes; one past it does not', () => {
  const lowOk = newRoot();
  fillPicked(lowOk, 27, 'reggae', { scene: { primary: 'reggae' } });
  fillPicked(lowOk, 73, 'other', {});
  assert.doesNotMatch(run(['balance', '--root', lowOk]).stdout, /share: reggae:/, 'count=27 of 100 is the low edge, inclusive');
  rmSync(lowOk, { recursive: true, force: true });

  const lowShort = newRoot();
  fillPicked(lowShort, 26, 'reggae', { scene: { primary: 'reggae' } });
  fillPicked(lowShort, 74, 'other', {});
  assert.match(run(['balance', '--root', lowShort]).stdout, /share: reggae: reggae is 26 of 100 \(short\)/);
  rmSync(lowShort, { recursive: true, force: true });

  const highOk = newRoot();
  fillPicked(highOk, 33, 'reggae', { scene: { primary: 'reggae' } });
  fillPicked(highOk, 67, 'other', {});
  assert.doesNotMatch(run(['balance', '--root', highOk]).stdout, /share: reggae:/, 'count=33 of 100 is the high edge, inclusive');
  rmSync(highOk, { recursive: true, force: true });

  const highOver = newRoot();
  fillPicked(highOver, 34, 'reggae', { scene: { primary: 'reggae' } });
  fillPicked(highOver, 66, 'other', {});
  assert.match(run(['balance', '--root', highOver]).stdout, /share: reggae: reggae is 34 of 100 \(over\)/);
  rmSync(highOver, { recursive: true, force: true });
});

// --- candidate -------------------------------------------------------------

test('candidate: a fully populated candidate directory passes', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1');
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a missing thumbnail fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', {}, { skip: ['thumb'] });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: missing-thumbnail: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a missing preview fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', {}, { skip: ['preview'] });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: missing-preview: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a missing source file fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', {}, { skip: ['source'] });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: missing-source: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a hash mismatch against the real source bytes fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_sha256: 'b'.repeat(64) });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: hash-mismatch: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a source under the 1000px floor fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_width: 900, source_height: 600 });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: size-floor: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: missing source_width/source_height fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_width: undefined, source_height: undefined });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: source-dimensions: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a missing screen fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { screen: undefined });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: screen: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: an unknown people flag fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { flags: ['not-a-real-flag'] });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: vocab:flag: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: an NC license without pack fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { license: 'CC-BY-NC-4.0' });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: pack-flag: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: an NC license set as pack passes', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { license: 'CC-BY-NC-4.0', pack: true });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 0, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: pack set true on a shipped license fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { pack: true }); // license defaults to CC-BY-2.0, a shipped class.
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: pack-flag: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a filename outside the source prefix pattern fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  mkdirSync(dir, { recursive: true });
  writeFileSync(path.join(dir, 'not-a-source-1.json'), JSON.stringify(baseCandidateRecord()));
  const { status, stdout } = run(['candidate', path.join(dir, 'not-a-source-1.json')]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: filename: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: source_id not matching the filename fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_id: 'commons:99999' });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: filename-mismatch: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: status/sha256/width/height/withdrawn are forbidden on a candidate', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', {
    status: 'picked',
    sha256: 'a'.repeat(64),
    width: 1600,
    height: 1067,
    withdrawn: { reason: 'x', date: '2026-01-01' },
  });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: status-forbidden: /);
  assert.match(stdout, /: sha256-forbidden: /);
  assert.match(stdout, /: dimensions-forbidden: /);
  assert.match(stdout, /: withdrawn-forbidden: /);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a directory ignores drops.json and raw/, and reports a within-directory repeat and a root collision', () => {
  const root = newRoot();
  writePicked(root, 'already-picked', { source_id: 'commons:picked-1' });
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_id: 'commons:picked-1' }); // collides with the picked pool.
  writeCandidate(dir, 'flickr-2', { source_id: 'flickr:2', origin_id: 'flickr:shared-origin' });
  writeCandidate(dir, 'flickr-3', { source_id: 'flickr:3', origin_id: 'flickr:shared-origin' }); // repeats flickr-2's origin_id.
  writeFileSync(path.join(dir, 'drops.json'), JSON.stringify([{ source_id: 'commons:99', reason: 'blurred' }]));
  mkdirSync(path.join(dir, 'raw'), { recursive: true });
  writeFileSync(path.join(dir, 'raw', 'commons-1.json'), '{ this is not read as a record }');
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /commons-1\.json: root-collision: /);
  assert.match(stdout, /flickr-3\.json: repeat: /);
  assert.doesNotMatch(stdout, /drops\.json/);
  assert.doesNotMatch(stdout, /raw[\\/]/);
  rmSync(root, { recursive: true, force: true });
});

test('candidate: a bad source_id fails', () => {
  const root = newRoot();
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_id: 'commons-not-a-colon' });
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /: source-id: /);
  rmSync(root, { recursive: true, force: true });
});

// --- sheet -----------------------------------------------------------------

test('sheet: compiles candidates, dedupes across agents, drops pack records, and skips _sheet/', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', { screen: { score: 5, note: 'Great.' } });
  writeCandidate(path.join(root, '.candidates', 'agent-b'), 'flickr-2', {
    source_id: 'flickr:2',
    screen: { score: 3, note: 'OK.' },
  });
  // A pack candidate never reaches the sheet.
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-3', { license: 'CC-BY-NC-4.0', pack: true });
  // A stray, otherwise-valid candidate record directly under _sheet/ must
  // never be read as an agent's, per the "_" skip rule.
  writeCandidate(path.join(root, '.candidates', '_sheet'), 'commons-9', { source_id: 'commons:9' });

  const outDir = path.join(root, 'out');
  const { status, stdout } = run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(status, 0, stdout);
  assert.match(stdout, /wrote 2 candidate/);
  assert.equal(existsSync(path.join(outDir, 'commons-9.json')), false);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: a duplicate origin_id across two agents is dropped, not written twice', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', {
    origin_id: 'flickr:shared',
    screen: { score: 5, note: 'Great.' },
  });
  writeCandidate(path.join(root, '.candidates', 'agent-b'), 'flickr-shared', {
    source_id: 'flickr:shared',
    origin_id: 'flickr:shared',
    screen: { score: 3, note: 'OK.' },
  });
  const outDir = path.join(root, 'out');
  const { status, stdout } = run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(status, 0, stdout);
  assert.match(stdout, /wrote 1 candidate\(s\), dropped 1 duplicate/);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: two agents both writing commons-1.json for the SAME source_id are deduped to one', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', {
    source_id: 'commons:1',
    screen: { score: 5, note: 'Great.' },
  });
  writeCandidate(path.join(root, '.candidates', 'agent-b'), 'commons-1', {
    source_id: 'commons:1',
    source_sha256: sha256(Buffer.from('agent-b-distinct-source-bytes')),
    screen: { score: 4, note: 'Also good.' },
  });
  const outDir = path.join(root, 'out');
  const { status, stdout } = run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(status, 0, stdout);
  assert.match(stdout, /wrote 1 candidate\(s\), dropped 1 duplicate/, stdout);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: a candidate whose source_id does not match its own filename is skipped, not compiled', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', {
    source_id: 'commons:99999', // does not match the filename "commons-1"
    screen: { score: 5, note: 'Great.' },
  });
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'flickr-2', {
    source_id: 'flickr:2',
    screen: { score: 3, note: 'OK.' },
  });
  const outDir = path.join(root, 'out');
  const { status, stdout } = run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(status, 0, stdout);
  assert.match(stdout, /commons-1: filename-mismatch: /, stdout);
  assert.deepEqual(readdirSync(outDir).sort(), ['flickr-2.json']);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: --out is cleared before writing, so a dropped candidate does not linger', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', { screen: { score: 5, note: 'Great.' } });
  const outDir = path.join(root, 'out');
  mkdirSync(outDir, { recursive: true });
  writeFileSync(path.join(outDir, 'stale-from-a-prior-round.json'), '{}');
  run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(existsSync(path.join(outDir, 'stale-from-a-prior-round.json')), false);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: second_image is set when a candidate matches a picked artist by name only, the artist carrying an MBID', () => {
  const root = newRoot();
  writePicked(root, 'bob-marley-1978', {
    source_id: 'commons:picked-bob',
    origin_id: undefined,
    artists: [{ name: 'Bob Marley', mbid: '5f4dd6f2-a935-478d-a930-e4d8affe1f10' }],
  });
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-2', {
    source_id: 'commons:2',
    artists: [{ name: 'Bob Marley' }], // no MBID on the candidate side.
  });
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-3', {
    source_id: 'commons:3',
    artists: [{ name: 'A Totally Different Artist' }],
  });
  const outDir = path.join(root, 'out');
  run(['sheet', '--root', root, '--out', outDir]);
  const matched = JSON.parse(readFileSync(path.join(outDir, 'commons-2.json'), 'utf8'));
  const unrelated = JSON.parse(readFileSync(path.join(outDir, 'commons-3.json'), 'utf8'));
  assert.equal(matched.second_image, true);
  assert.equal(unrelated.second_image, false);
  rmSync(root, { recursive: true, force: true });
});

test('sheet: second_image is set when two candidates share an artist name, neither carrying an MBID', () => {
  const root = newRoot();
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-1', {
    source_id: 'commons:1',
    artists: [{ name: 'Shared Band' }],
  });
  writeCandidate(path.join(root, '.candidates', 'agent-b'), 'flickr-2', {
    source_id: 'flickr:2',
    artists: [{ name: 'Shared Band' }],
  });
  const outDir = path.join(root, 'out');
  run(['sheet', '--root', root, '--out', outDir]);
  const first = JSON.parse(readFileSync(path.join(outDir, 'commons-1.json'), 'utf8'));
  const second = JSON.parse(readFileSync(path.join(outDir, 'flickr-2.json'), 'utf8'));
  assert.equal(first.second_image, true);
  assert.equal(second.second_image, true);
  rmSync(root, { recursive: true, force: true });
});
