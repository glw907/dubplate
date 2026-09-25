// Table-driven tests for check.mjs's four modes. Each failing case builds a
// fixture from a known-good base and mutates exactly one thing; the mutation
// is named in the case so a reviewer can see what one-line change would make
// it pass again.

import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { mkdtempSync, mkdirSync, writeFileSync, rmSync, copyFileSync, readFileSync } from 'node:fs';
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
// the actual bytes so the pair is hash-consistent unless the case asks for a
// mismatch.
function writePicked(root, slug, overrides = {}, { webpBytes, sha256Override } = {}) {
  const bytes = webpBytes || Buffer.from(`webp-bytes-for-${slug}`);
  writeFileSync(path.join(root, `${slug}.webp`), bytes);
  const record = basePickedRecord({ sha256: sha256Override || sha256(bytes), ...overrides });
  writeFileSync(path.join(root, `${slug}.json`), `${JSON.stringify(record, null, 2)}\n`);
  return record;
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
  const record = baseCandidateRecord({ source_sha256: sha256(sourceBytes), ...overrides });
  if (!skip.includes('source')) writeFileSync(path.join(dir, `${id}.jpg`), sourceBytes);
  if (!skip.includes('thumb')) writeFileSync(path.join(dir, `${id}.thumb.webp`), Buffer.from('thumb'));
  if (!skip.includes('preview')) writeFileSync(path.join(dir, `${id}.preview.webp`), Buffer.from('preview'));
  if (!skip.includes('json')) writeFileSync(path.join(dir, `${id}.json`), `${JSON.stringify(record, null, 2)}\n`);
  return record;
}

test('integrity: a fully populated passing pool exits 0 and ignores .candidates/', () => {
  const root = newRoot();
  writePicked(root, 'dischord-house-1984');
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
          verified: 'Quoted from the artist\'s own site.',
        },
      });
    },
    wantRule: 'artwork-license',
  },
  {
    name: 'a Commons record and its Flickr original sharing origin_id',
    // Flip: give record-flickr-origin a different origin_id.
    setup(root) {
      writePicked(root, 'commons-copy', { source_id: 'commons:1', origin_id: 'flickr:5' });
      writePicked(root, 'flickr-origin', { source_id: 'flickr:5', origin_id: 'flickr:5' });
    },
    wantRule: 'duplicate-key',
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
    name: 'a public_domain record on an unlisted base',
    // Flip: public_domain.basis: "pre-1931".
    setup(root) {
      writePicked(root, 'pd-bad-basis', {
        license: 'LicenseRef-public-domain',
        public_domain: {
          basis: 'not-a-real-basis',
          jurisdiction: 'US',
          us_status: 'public domain',
          fact: '1920 first publication',
        },
      });
    },
    wantRule: 'public-domain-basis',
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
    name: 'a bad source_id',
    // Flip: source_id: "commons:12345".
    setup(root) {
      writePicked(root, 'bad-source-id', { source_id: 'commons-12345' });
    },
    wantRule: 'source-id',
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
];

for (const c of integrityCases) {
  test(`integrity: ${c.name}`, () => {
    const root = newRoot();
    c.setup(root);
    const { status, stdout } = run(['integrity', '--root', root]);
    assert.equal(status, 1, stdout);
    assert.match(stdout, new RegExp(`: ${c.wantRule}: `), stdout);
    rmSync(root, { recursive: true, force: true });
  });
}

test('integrity: a matching waiver passes the same breach the mismatched key failed', () => {
  const root = newRoot();
  writePicked(root, 'record-a', { source_id: 'commons:1', origin_id: undefined });
  writePicked(root, 'record-b', {
    source_id: 'commons:2',
    origin_id: undefined,
    source_sha256: 'c'.repeat(64),
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

test('candidate: a directory ignores drops.json and raw/, and reports a repeat and a root collision', () => {
  const root = newRoot();
  writePicked(root, 'already-picked', { source_id: 'commons:picked-1' });
  const dir = path.join(root, '.candidates', 'a');
  writeCandidate(dir, 'commons-1', { source_id: 'commons:picked-1' }); // collides with the picked pool.
  writeCandidate(dir, 'flickr-2'); // a clean second record, no collision.
  writeFileSync(path.join(dir, 'drops.json'), JSON.stringify([{ source_id: 'commons:99', reason: 'blurred' }]));
  mkdirSync(path.join(dir, 'raw'), { recursive: true });
  writeFileSync(path.join(dir, 'raw', 'commons-1.json'), '{ this is not read as a record }');
  const { status, stdout } = run(['candidate', dir]);
  assert.equal(status, 1, stdout);
  assert.match(stdout, /commons-1\.json: root-collision: /);
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
  // Same origin_id as commons-1's implicit none; give it a real duplicate via source_sha256 collision instead:
  writeCandidate(path.join(root, '.candidates', 'agent-b'), 'flickr-2', {
    source_id: 'flickr:2',
    screen: { score: 3, note: 'OK.' },
  });
  // A pack candidate never reaches the sheet.
  writeCandidate(path.join(root, '.candidates', 'agent-a'), 'commons-3', { license: 'CC-BY-NC-4.0', pack: true });
  // A stray _sheet/ directory under .candidates/ must never be read as an agent.
  mkdirSync(path.join(root, '.candidates', '_sheet'), { recursive: true });
  writeFileSync(path.join(root, '.candidates', '_sheet', 'commons-1.json'), '{ this would break sheet if read }');

  const outDir = path.join(root, 'out');
  const { status, stdout } = run(['sheet', '--root', root, '--out', outDir]);
  assert.equal(status, 0, stdout);
  assert.match(stdout, /wrote 2 candidate/);
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
