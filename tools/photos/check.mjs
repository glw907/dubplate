#!/usr/bin/env node
// check.mjs is the one dependency-free check for the photo pool. It has four
// modes: integrity (the pool committed under photos/), balance (the picked
// pool against the share and floor targets), candidate (one candidate record
// or a search agent's directory), and sheet (the round-open compile step).
// It reads photos/meta/vocab.json directly; nothing else in the repo imports
// it. Retires at rung 7, when the Go loader ports both modes' checks.

import { createHash } from 'node:crypto';
import { readFileSync, readdirSync, existsSync, statSync, writeFileSync, mkdirSync } from 'node:fs';
import path from 'node:path';

// The rules a waiver in photos/meta/waivers.json may cover (Pool size and
// balance). `artist` and `landmark` are checked by integrity; the rest by
// balance.
const WAIVER_RULES = new Set([
  'share',
  'region',
  'posed',
  'floor',
  'landmark-floor',
  'surface',
  'artist',
  'landmark',
]);

// Share targets and the plus-or-minus-3pp band (Pool size and balance).
const FAMILY_SHARES = { reggae: 0.30, punk: 0.20, 'hip-hop': 0.20, other: 0.30 };
const SHARE_BAND = 0.03;
const REGION_CAP = 0.5;
const POSED_CEILING = 0.15;
const FLOOR_GRID_FAMILIES = ['reggae', 'punk', 'hip-hop'];
const FLOOR_MINIMUM = 2;
const LANDMARK_FLOOR_MINIMUM = 2;
const SURFACE_FLOORS = { error: 3, empty: 3 };

const SOURCE_ID_RE = /^(commons|flickr|archive):[A-Za-z0-9_.~:@+-]{1,190}$/;
// A candidate filename's <id> spells the source_id's colon as a dash.
const CANDIDATE_ID_RE = /^(commons|flickr|archive)-[A-Za-z0-9_.~-]{1,190}$/;
const SLUG_RE = /^[a-z0-9]+(-[a-z0-9]+)*$/;
const SHA256_RE = /^[0-9a-f]{64}$/;
const SOURCE_EXTENSIONS = ['.jpg', '.jpeg', '.png', '.tif', '.tiff', '.webp'];

// An "unreadable input" the caller cannot recover from: a missing or
// unparsable vocab.json, or (for candidate/sheet) a root that has none.
class UnreadableInputError extends Error {}

function readJSONFile(filePath) {
  const text = readFileSync(filePath, 'utf8');
  return JSON.parse(text);
}

function loadVocab(root) {
  const vocabPath = path.join(root, 'meta', 'vocab.json');
  let vocab;
  try {
    vocab = readJSONFile(vocabPath);
  } catch (err) {
    throw new UnreadableInputError(`${vocabPath}: ${err.message}`);
  }
  return vocab;
}

// A meta array file (waivers.json, rejected.json, pack.json) that is missing
// is tolerated as empty; one that fails to parse is an integrity finding,
// not a fatal error, so the rest of the pool still gets checked.
function loadMetaArray(root, name, errors) {
  const filePath = path.join(root, 'meta', name);
  if (!existsSync(filePath)) return [];
  try {
    const data = readJSONFile(filePath);
    if (!Array.isArray(data)) {
      errors.push({ file: name, rule: 'parse-error', message: 'must be a JSON array' });
      return [];
    }
    return data;
  } catch (err) {
    errors.push({ file: name, rule: 'parse-error', message: err.message });
    return [];
  }
}

function longEdge(width, height) {
  return Math.max(width, height);
}

// Case folded, a leading "the" dropped (Pool size and balance).
function normalizeName(name) {
  return name.trim().toLowerCase().replace(/^the\s+/, '').trim();
}

function artistKey(artist) {
  if (artist.mbid) return `mbid:${artist.mbid}`;
  return `name:${normalizeName(artist.name)}`;
}

function landmarkKey(landmark) {
  if (landmark.wikidata_id) return `wikidata:${landmark.wikidata_id}`;
  return `name:${normalizeName(landmark.name)}`;
}

// The family comes from the primary scene through vocab.json; a record with
// no scene counts toward "other" (Categories).
function deriveFamily(record, vocab) {
  if (!record.scene || !record.scene.primary) return 'other';
  const scene = vocab.scenes[record.scene.primary];
  return scene ? scene.family : undefined;
}

function regionOf(countryCode, vocab) {
  for (const [id, region] of Object.entries(vocab.regions)) {
    if (region.country_codes.includes(countryCode)) return id;
  }
  return undefined;
}

function licenseById(id, vocab) {
  return vocab.licenses.find((l) => l.id === id);
}

function hasWaiver(waivers, rule, key) {
  return waivers.some((w) => w.rule === rule && w.key === key);
}

// Shared field checks for a picked record and a candidate record: the
// facets, the license and public-domain rules, the artwork rule, and the
// size floor. `width`/`height` pick the dimensions to floor-check (a picked
// record's stored size; a candidate's source size).
function validateCommonFields(record, vocab, file, errors, { width, height, allowPackLicense = false }) {
  const push = (rule, message) => errors.push({ file, rule, message });

  if (!record.title || typeof record.title !== 'string') push('title', 'missing title');
  if (!record.photographer || typeof record.photographer !== 'string') {
    push('photographer', 'missing photographer');
  }
  if (!record.source_id || !SOURCE_ID_RE.test(record.source_id)) {
    push('source-id', `source_id ${JSON.stringify(record.source_id)} does not match the pattern`);
  }
  if (!record.source_url) push('source-url', 'missing source_url');
  if (!record.file_url) push('file-url', 'missing file_url');
  if (!record.retrieved_at) push('retrieved-at', 'missing retrieved_at');
  if (!record.verified || typeof record.verified !== 'string') push('verified', 'missing verified');
  if (!record.source_sha256 || !SHA256_RE.test(record.source_sha256)) {
    push('source-sha256', 'missing or malformed source_sha256');
  }

  if (!record.format || !vocab.formats.includes(record.format)) {
    push('vocab:format', `format ${JSON.stringify(record.format)} is not in vocab.json`);
  }
  if (!record.shot || !vocab.shots.includes(record.shot)) {
    push('vocab:shot', `shot ${JSON.stringify(record.shot)} is not in vocab.json`);
  }
  if (!Array.isArray(record.subject) || record.subject.length === 0) {
    push('subject', 'subject must be a non-empty array');
  } else {
    for (const s of record.subject) {
      if (!(s in vocab.subjects)) push('vocab:subject', `subject ${JSON.stringify(s)} is not in vocab.json`);
    }
  }
  if (record.surfaces) {
    for (const s of record.surfaces) {
      if (!vocab.surfaces.includes(s)) push('vocab:surface', `surface ${JSON.stringify(s)} is not in vocab.json`);
    }
  }
  if (record.scene && record.scene.primary && !(record.scene.primary in vocab.scenes)) {
    push('vocab:scene', `scene ${JSON.stringify(record.scene.primary)} is not in vocab.json`);
  }
  for (const s of (record.scene && record.scene.secondary) || []) {
    if (!(s in vocab.scenes)) push('vocab:scene', `secondary scene ${JSON.stringify(s)} is not in vocab.json`);
  }
  if (record.place) {
    const region = regionOf(record.place.country_code, vocab);
    if (!region) push('vocab:region', `country_code ${JSON.stringify(record.place.country_code)} has no region`);
    if (record.place.landmark && record.place.landmark.name === undefined) {
      push('landmark', 'landmark requires a name');
    }
  }

  // License and public-domain.
  const license = record.license ? licenseById(record.license, vocab) : undefined;
  if (!license) {
    push('vocab:license', `license ${JSON.stringify(record.license)} is not in vocab.json`);
  } else if (license.class !== 'shipped' && !allowPackLicense) {
    // A picked record's license is always a shipped class; round close moves
    // a pack candidate to pack.json instead of committing it (Storage).
    push('license-class', `license ${license.id} is not a shipped class`);
  }
  if (record.license === 'LicenseRef-public-domain') {
    if (!record.public_domain) {
      push('public-domain-required', 'LicenseRef-public-domain requires public_domain');
    } else if (!vocab.public_domain_bases.some((b) => b.id === record.public_domain.basis)) {
      push('public-domain-basis', `basis ${JSON.stringify(record.public_domain.basis)} is not in vocab.json`);
    }
  } else if (record.public_domain) {
    push('public-domain-forbidden', 'public_domain is set on a non-public-domain license');
  }

  // Artwork, required for flyer and zine-cover formats.
  if (record.format === 'flyer' || record.format === 'zine-cover') {
    if (!record.artwork) {
      push('artwork-required', `${record.format} requires artwork`);
    } else {
      const artworkLicense = record.artwork.license ? licenseById(record.artwork.license, vocab) : undefined;
      if (!artworkLicense) {
        push('vocab:license', `artwork license ${JSON.stringify(record.artwork.license)} is not in vocab.json`);
      } else if (artworkLicense.class !== 'shipped') {
        push('artwork-license', `artwork license ${artworkLicense.id} is not a shipped class`);
      }
    }
  } else if (record.artwork) {
    push('artwork-forbidden', 'artwork is set on a format that is not flyer or zine-cover');
  }

  // Size floor, on the dimensions the caller names.
  if (Number.isInteger(width) && Number.isInteger(height)) {
    const edge = longEdge(width, height);
    if (edge < 1000) {
      push('size-floor', `long edge ${edge}px is under the 1000px floor`);
    } else if (edge < 1200) {
      if (!record.size_exception || typeof record.size_exception !== 'string') {
        push('size-exception-required', `long edge ${edge}px needs size_exception`);
      }
    } else if (record.size_exception) {
      push('size-exception-forbidden', `size_exception is set at ${edge}px, at or above the 1200px floor`);
    }
  }
}

// --- integrity ---------------------------------------------------------

function scanTopLevel(root, ext) {
  return readdirSync(root, { withFileTypes: true })
    .filter((d) => d.isFile() && d.name.endsWith(ext) && !d.name.startsWith('.'))
    .map((d) => d.name)
    .sort();
}

function runIntegrity(root) {
  const vocab = loadVocab(root);
  const errors = [];

  const waivers = loadMetaArray(root, 'waivers.json', errors);
  for (const w of waivers) {
    if (!WAIVER_RULES.has(w.rule)) {
      errors.push({ file: 'waivers.json', rule: 'waiver-rule', message: `rule ${JSON.stringify(w.rule)} is unknown` });
    }
    if (!w.key) errors.push({ file: 'waivers.json', rule: 'waiver-key', message: 'waiver has no key' });
    if (!w.reason) errors.push({ file: 'waivers.json', rule: 'waiver-reason', message: 'waiver has no reason' });
  }
  const rejected = loadMetaArray(root, 'rejected.json', errors);
  const pack = loadMetaArray(root, 'pack.json', errors);

  const jsonFiles = scanTopLevel(root, '.json');
  const webpFiles = new Set(scanTopLevel(root, '.webp').map((f) => f.slice(0, -'.webp'.length)));

  const picked = new Map(); // slug -> record
  for (const file of jsonFiles) {
    const slug = file.slice(0, -'.json'.length);
    if (!SLUG_RE.test(slug)) {
      errors.push({ file, rule: 'slug-pattern', message: `${slug} does not match the slug pattern` });
    }
    let record;
    try {
      record = readJSONFile(path.join(root, file));
    } catch (err) {
      errors.push({ file, rule: 'parse-error', message: err.message });
      continue;
    }
    picked.set(slug, record);
  }

  // Dedupe keys across picked, rejected, and pack records (integrity).
  const seen = new Map(); // "type:value" -> file
  const claim = (type, value, file) => {
    if (!value) return;
    const k = `${type}:${value}`;
    const prior = seen.get(k);
    if (prior && prior !== file) {
      errors.push({ file, rule: 'duplicate-key', message: `${type} ${JSON.stringify(value)} already used by ${prior}` });
    } else {
      seen.set(k, file);
    }
  };

  const artistCounts = new Map(); // key -> [files]
  const landmarkCounts = new Map();

  for (const [slug, record] of picked) {
    const file = `${slug}.json`;
    if (record.status !== 'picked' && record.status !== 'withdrawn') {
      errors.push({ file, rule: 'status', message: `status ${JSON.stringify(record.status)} is not picked or withdrawn` });
      continue;
    }

    if (record.status === 'withdrawn') {
      if (!record.withdrawn || !record.withdrawn.reason || !record.withdrawn.date) {
        errors.push({ file, rule: 'withdrawn-fields', message: 'withdrawn requires reason and date' });
      }
      if (webpFiles.has(slug)) {
        errors.push({ file, rule: 'withdrawn-has-image', message: `${slug}.webp still exists` });
      }
      claim('source_id', record.source_id, file);
      claim('origin_id', record.origin_id, file);
      claim('source_sha256', record.source_sha256, file);
      continue;
    }

    if (record.withdrawn) {
      errors.push({ file, rule: 'withdrawn-forbidden', message: 'withdrawn is set on a picked record' });
    }

    const webpPath = path.join(root, `${slug}.webp`);
    if (!webpFiles.has(slug)) {
      errors.push({ file, rule: 'missing-webp', message: `${slug}.webp does not exist` });
    } else {
      const actualHash = createHash('sha256').update(readFileSync(webpPath)).digest('hex');
      if (record.sha256 !== actualHash) {
        errors.push({ file, rule: 'hash-mismatch', message: `sha256 ${record.sha256} does not match the file's ${actualHash}` });
      }
    }

    validateCommonFields(record, vocab, file, errors, { width: record.width, height: record.height });

    claim('source_id', record.source_id, file);
    claim('origin_id', record.origin_id, file);
    claim('source_sha256', record.source_sha256, file);

    for (const artist of record.artists || []) {
      const key = artistKey(artist);
      if (!artistCounts.has(key)) artistCounts.set(key, []);
      artistCounts.get(key).push(file);
    }
    if (record.place && record.place.landmark) {
      const key = landmarkKey(record.place.landmark);
      if (!landmarkCounts.has(key)) landmarkCounts.set(key, []);
      landmarkCounts.get(key).push(file);
    }
  }

  for (const slug of webpFiles) {
    const record = picked.get(slug);
    if (!record || record.status !== 'picked') {
      errors.push({ file: `${slug}.webp`, rule: 'orphan-webp', message: 'no picked record claims this file' });
    }
  }

  for (const [key, files] of artistCounts) {
    if (files.length > 1 && !hasWaiver(waivers, 'artist', key)) {
      for (const file of files.slice(1)) {
        errors.push({ file, rule: 'artist-cap', message: `a second image of artist ${key} needs a waiver` });
      }
    }
  }
  for (const [key, files] of landmarkCounts) {
    if (files.length > 1 && !hasWaiver(waivers, 'landmark', key)) {
      for (const file of files.slice(1)) {
        errors.push({ file, rule: 'landmark-cap', message: `a second image of landmark ${key} needs a waiver` });
      }
    }
  }

  for (const entry of rejected) {
    claim('source_id', entry.source_id, 'rejected.json');
    claim('origin_id', entry.origin_id, 'rejected.json');
    claim('source_sha256', entry.source_sha256, 'rejected.json');
    if (entry.reason && !vocab.reasons.includes(entry.reason)) {
      errors.push({ file: 'rejected.json', rule: 'vocab:reason', message: `reason ${JSON.stringify(entry.reason)} is not in vocab.json` });
    }
  }
  for (const entry of pack) {
    claim('source_id', entry.source_id, 'pack.json');
    claim('origin_id', entry.origin_id, 'pack.json');
    claim('source_sha256', entry.source_sha256, 'pack.json');
  }

  return errors;
}

// --- balance -------------------------------------------------------------

function runBalance(root) {
  const vocab = loadVocab(root);
  const errors = [];
  const waivers = loadMetaArray(root, 'waivers.json', errors);

  const jsonFiles = scanTopLevel(root, '.json');
  const picked = [];
  for (const file of jsonFiles) {
    let record;
    try {
      record = readJSONFile(path.join(root, file));
    } catch {
      continue; // integrity reports parse errors; balance only counts what parses.
    }
    if (record.status === 'picked') picked.push(record);
  }

  const total = picked.length;
  const breaches = [];

  // Share bands.
  const shareCounts = { reggae: 0, punk: 0, 'hip-hop': 0, other: 0 };
  for (const record of picked) {
    const family = deriveFamily(record, vocab);
    if (family in shareCounts) shareCounts[family] += 1;
  }
  const shares = {};
  for (const [family, target] of Object.entries(FAMILY_SHARES)) {
    const count = shareCounts[family];
    const share = total > 0 ? count / total : 0;
    const lo = target - SHARE_BAND;
    const hi = target + SHARE_BAND;
    let status;
    if (total === 0) status = 'short';
    else if (share < lo) status = 'short';
    else if (share > hi) status = 'over';
    else status = 'ok';
    shares[family] = { count, total, share, target, status };
    if (status !== 'ok' && !hasWaiver(waivers, 'share', family)) {
      breaches.push({ rule: 'share', key: family, status, message: `${family} is ${count} of ${total} (${status})` });
    }
  }

  // Region cap: at most half of records that carry a place.
  const withPlace = picked.filter((r) => r.place && r.place.country_code);
  const regionCounts = new Map();
  for (const record of withPlace) {
    const region = regionOf(record.place.country_code, vocab);
    if (!region) continue;
    regionCounts.set(region, (regionCounts.get(region) || 0) + 1);
  }
  const regions = {};
  for (const [region, count] of regionCounts) {
    const share = count / withPlace.length;
    const status = share > REGION_CAP ? 'over' : 'ok';
    regions[region] = { count, total: withPlace.length, share, status };
    if (status === 'over' && !hasWaiver(waivers, 'region', region)) {
      breaches.push({ rule: 'region', key: region, status, message: `${region} is ${count} of ${withPlace.length} (over the region cap)` });
    }
  }

  // Posed ceiling, over photographs only (flyers and zine covers excluded).
  const photographs = picked.filter((r) => r.format === 'photograph');
  const posedCount = photographs.filter((r) => r.shot === 'posed').length;
  const posedShare = photographs.length > 0 ? posedCount / photographs.length : 0;
  const posedStatus = photographs.length > 0 && posedShare > POSED_CEILING ? 'over' : 'ok';
  const posed = { count: posedCount, total: photographs.length, share: posedShare, status: posedStatus };
  if (posedStatus === 'over' && !hasWaiver(waivers, 'posed', 'all')) {
    breaches.push({ rule: 'posed', key: 'all', status: posedStatus, message: `posed is ${posedCount} of ${photographs.length} (over the posed ceiling)` });
  }

  // Floor grid: 15 cells, family x floor-grid subject, primary subject only.
  const floorGrid = {};
  for (const family of FLOOR_GRID_FAMILIES) {
    for (const subject of vocab.floor_grid_subjects) {
      const key = `${family}/${subject}`;
      const count = picked.filter((r) => deriveFamily(r, vocab) === family && r.subject && r.subject[0] === subject).length;
      const status = count < FLOOR_MINIMUM ? 'short' : 'ok';
      floorGrid[key] = { count, minimum: FLOOR_MINIMUM, status };
      if (status === 'short' && !hasWaiver(waivers, 'floor', key)) {
        breaches.push({ rule: 'floor', key, status, message: `${key} is ${count} of ${FLOOR_MINIMUM} (short)` });
      }
    }
  }

  // Landmark floor, per named family.
  const landmarkFloor = {};
  for (const family of FLOOR_GRID_FAMILIES) {
    const count = picked.filter((r) => deriveFamily(r, vocab) === family && r.place && r.place.landmark).length;
    const status = count < LANDMARK_FLOOR_MINIMUM ? 'short' : 'ok';
    landmarkFloor[family] = { count, minimum: LANDMARK_FLOOR_MINIMUM, status };
    if (status === 'short' && !hasWaiver(waivers, 'landmark-floor', family)) {
      breaches.push({ rule: 'landmark-floor', key: family, status, message: `${family} landmark floor is ${count} of ${LANDMARK_FLOOR_MINIMUM} (short)` });
    }
  }

  // Surface floors.
  const surfaces = {};
  for (const [surface, minimum] of Object.entries(SURFACE_FLOORS)) {
    const count = picked.filter((r) => (r.surfaces || []).includes(surface)).length;
    const status = count < minimum ? 'short' : 'ok';
    surfaces[surface] = { count, minimum, status };
    if (status === 'short' && !hasWaiver(waivers, 'surface', surface)) {
      breaches.push({ rule: 'surface', key: surface, status, message: `surface ${surface} is ${count} of ${minimum} (short)` });
    }
  }

  // Normalized artist and landmark names and IDs, for --json.
  const artists = new Map();
  const landmarks = new Map();
  for (const record of picked) {
    for (const artist of record.artists || []) {
      const key = artistKey(artist);
      const entry = artists.get(key) || { name: artist.name, mbid: artist.mbid || null, count: 0 };
      entry.count += 1;
      artists.set(key, entry);
    }
    if (record.place && record.place.landmark) {
      const key = landmarkKey(record.place.landmark);
      const entry = landmarks.get(key) || {
        name: record.place.landmark.name,
        wikidata_id: record.place.landmark.wikidata_id || null,
        count: 0,
      };
      entry.count += 1;
      landmarks.set(key, entry);
    }
  }

  const report = {
    total,
    shares,
    regions,
    posed,
    floor_grid: floorGrid,
    landmark_floor: landmarkFloor,
    surfaces,
    breaches,
    scene_to_family: Object.fromEntries(Object.entries(vocab.scenes).map(([id, s]) => [id, s.family])),
    country_to_region: Object.fromEntries(
      Object.entries(vocab.regions).flatMap(([region, r]) => r.country_codes.map((c) => [c, region])),
    ),
    artists: Object.fromEntries(artists),
    landmarks: Object.fromEntries(landmarks),
  };

  return { report, breaches };
}

// --- candidate -------------------------------------------------------------

function findRoot(startPath) {
  let dir = statSync(startPath).isDirectory() ? startPath : path.dirname(startPath);
  dir = path.resolve(dir);
  for (;;) {
    if (existsSync(path.join(dir, 'meta', 'vocab.json'))) return dir;
    const parent = path.dirname(dir);
    if (parent === dir) return undefined;
    dir = parent;
  }
}

function findSourceFile(dir, id) {
  for (const ext of SOURCE_EXTENSIONS) {
    const candidate = path.join(dir, `${id}${ext}`);
    if (existsSync(candidate)) return candidate;
  }
  return undefined;
}

// Validates one candidate record on disk. Returns a list of errors; an empty
// list means the record passes.
function validateCandidateRecord(dir, filename, vocab, errors) {
  const id = filename.slice(0, -'.json'.length);
  if (!CANDIDATE_ID_RE.test(id)) {
    errors.push({ file: filename, rule: 'filename', message: `${id} does not match the candidate filename pattern` });
    return;
  }
  let record;
  try {
    record = readJSONFile(path.join(dir, filename));
  } catch (err) {
    errors.push({ file: filename, rule: 'parse-error', message: err.message });
    return;
  }

  validateCommonFields(record, vocab, filename, errors, {
    width: record.source_width,
    height: record.source_height,
    allowPackLicense: true,
  });

  if (!Number.isInteger(record.source_width) || !Number.isInteger(record.source_height)) {
    errors.push({ file: filename, rule: 'source-dimensions', message: 'source_width and source_height are required' });
  }
  if (record.screen === undefined) errors.push({ file: filename, rule: 'screen', message: 'screen is required' });

  const sourcePath = findSourceFile(dir, id);
  if (!sourcePath) {
    errors.push({ file: filename, rule: 'missing-source', message: `no source file for ${id}` });
  } else {
    const actualHash = createHash('sha256').update(readFileSync(sourcePath)).digest('hex');
    if (record.source_sha256 !== actualHash) {
      errors.push({ file: filename, rule: 'hash-mismatch', message: `source_sha256 ${record.source_sha256} does not match the file's ${actualHash}` });
    }
  }
  if (!existsSync(path.join(dir, `${id}.thumb.webp`))) {
    errors.push({ file: filename, rule: 'missing-thumbnail', message: `${id}.thumb.webp does not exist` });
  }
  if (!existsSync(path.join(dir, `${id}.preview.webp`))) {
    errors.push({ file: filename, rule: 'missing-preview', message: `${id}.preview.webp does not exist` });
  }

  const license = record.license ? licenseById(record.license, vocab) : undefined;
  const wantsPack = license ? license.class === 'pack' : false;
  if (Boolean(record.pack) !== wantsPack) {
    errors.push({ file: filename, rule: 'pack-flag', message: `pack must be ${wantsPack} for license ${record.license}` });
  }
  if (record.flags) {
    for (const flag of record.flags) {
      if (!vocab.people_flags.includes(flag)) {
        errors.push({ file: filename, rule: 'vocab:flag', message: `flag ${JSON.stringify(flag)} is not in vocab.json` });
      }
    }
  }
}

// Lists the candidate record filenames directly inside `dir`: files matching
// the <source>-<key>.json pattern, skipping drops.json, raw evidence, and
// any subdirectory whose name starts with "_" (the sheet's own output).
function listCandidateFiles(dir) {
  return readdirSync(dir, { withFileTypes: true })
    .filter((d) => d.isFile() && d.name.endsWith('.json') && CANDIDATE_ID_RE.test(d.name.slice(0, -'.json'.length)))
    .map((d) => d.name)
    .sort();
}

function collectRootKeys(root) {
  const keys = new Map(); // "type:value" -> source ("picked"|"rejected"|"pack")
  const claim = (type, value, source) => {
    if (value) keys.set(`${type}:${value}`, source);
  };
  for (const file of scanTopLevel(root, '.json')) {
    try {
      const record = readJSONFile(path.join(root, file));
      claim('source_id', record.source_id, 'picked');
      claim('origin_id', record.origin_id, 'picked');
      claim('source_sha256', record.source_sha256, 'picked');
    } catch {
      // integrity reports parse errors on the picked pool itself.
    }
  }
  const errors = [];
  for (const entry of loadMetaArray(root, 'rejected.json', errors)) {
    claim('source_id', entry.source_id, 'rejected');
    claim('origin_id', entry.origin_id, 'rejected');
    claim('source_sha256', entry.source_sha256, 'rejected');
  }
  for (const entry of loadMetaArray(root, 'pack.json', errors)) {
    claim('source_id', entry.source_id, 'pack');
    claim('origin_id', entry.origin_id, 'pack');
    claim('source_sha256', entry.source_sha256, 'pack');
  }
  return keys;
}

function runCandidate(target) {
  const stat = statSync(target); // throws ENOENT if the target is missing.
  const dir = stat.isDirectory() ? target : path.dirname(target);
  const root = findRoot(target);
  if (!root) throw new UnreadableInputError(`${target}: no meta/vocab.json found in any ancestor`);
  const vocab = loadVocab(root);

  const errors = [];
  const filenames = stat.isDirectory() ? listCandidateFiles(dir) : [path.basename(target)];

  const seenHere = new Map(); // "type:value" -> filename, for repeats within this read.
  const rootKeys = stat.isDirectory() ? collectRootKeys(root) : new Map();

  for (const filename of filenames) {
    validateCandidateRecord(dir, filename, vocab, errors);
    let record;
    try {
      record = readJSONFile(path.join(dir, filename));
    } catch {
      continue;
    }
    for (const [type, value] of [
      ['source_id', record.source_id],
      ['origin_id', record.origin_id],
      ['source_sha256', record.source_sha256],
    ]) {
      if (!value) continue;
      const k = `${type}:${value}`;
      const priorHere = seenHere.get(k);
      if (priorHere && priorHere !== filename) {
        errors.push({ file: filename, rule: 'repeat', message: `${type} ${JSON.stringify(value)} repeats ${priorHere}` });
      } else {
        seenHere.set(k, filename);
      }
      const rootSource = rootKeys.get(k);
      if (rootSource) {
        errors.push({ file: filename, rule: 'root-collision', message: `${type} ${JSON.stringify(value)} already in ${rootSource}` });
      }
    }
  }

  return errors;
}

// --- sheet -----------------------------------------------------------------

function listAgentDirs(candidatesDir) {
  if (!existsSync(candidatesDir)) return [];
  return readdirSync(candidatesDir, { withFileTypes: true })
    .filter((d) => d.isDirectory() && !d.name.startsWith('_'))
    .map((d) => d.name)
    .sort();
}

function runSheet(root, outDir) {
  const vocab = loadVocab(root);
  const candidatesDir = path.join(root, '.candidates');
  const rootKeys = collectRootKeys(root);

  const kept = []; // { id, record, agent }
  const dropped = []; // { id, agent, reason: 'duplicate' }
  const seen = new Map(); // "type:value" -> id, first found wins

  for (const agent of listAgentDirs(candidatesDir)) {
    const agentDir = path.join(candidatesDir, agent);
    for (const filename of listCandidateFiles(agentDir)) {
      const id = filename.slice(0, -'.json'.length);
      let record;
      try {
        record = readJSONFile(path.join(agentDir, filename));
      } catch {
        continue; // check.mjs candidate reports parse errors; sheet only compiles what parses.
      }
      if (record.pack) continue; // pack candidates never appear on the sorting page.

      let duplicate = false;
      for (const [type, value] of [
        ['source_id', record.source_id],
        ['origin_id', record.origin_id],
        ['source_sha256', record.source_sha256],
      ]) {
        if (!value) continue;
        const k = `${type}:${value}`;
        if (rootKeys.has(k) || (seen.has(k) && seen.get(k) !== id)) {
          duplicate = true;
          break;
        }
      }
      if (duplicate) {
        dropped.push({ id, agent, reason: 'duplicate' });
        continue;
      }
      for (const [type, value] of [
        ['source_id', record.source_id],
        ['origin_id', record.origin_id],
        ['source_sha256', record.source_sha256],
      ]) {
        if (value) seen.set(`${type}:${value}`, id);
      }
      kept.push({ id, record, agent });
    }
  }

  // Second-image flags, against the picked pool's dedupe keys and the other
  // kept candidates' normalized names.
  const artistOwners = new Map();
  const landmarkOwners = new Map();
  for (const { id, record } of kept) {
    for (const artist of record.artists || []) {
      const key = artistKey(artist);
      if (!artistOwners.has(key)) artistOwners.set(key, []);
      artistOwners.get(key).push(id);
    }
    if (record.place && record.place.landmark) {
      const key = landmarkKey(record.place.landmark);
      if (!landmarkOwners.has(key)) landmarkOwners.set(key, []);
      landmarkOwners.get(key).push(id);
    }
  }

  const pickedArtistKeys = new Set();
  const pickedLandmarkKeys = new Set();
  for (const file of scanTopLevel(root, '.json')) {
    try {
      const record = readJSONFile(path.join(root, file));
      if (record.status !== 'picked') continue;
      for (const artist of record.artists || []) pickedArtistKeys.add(artistKey(artist));
      if (record.place && record.place.landmark) pickedLandmarkKeys.add(landmarkKey(record.place.landmark));
    } catch {
      continue;
    }
  }

  const docs = kept.map(({ id, record, agent }) => {
    const secondImage =
      (record.artists || []).some((a) => {
        const key = artistKey(a);
        return pickedArtistKeys.has(key) || artistOwners.get(key).length > 1;
      }) ||
      (record.place && record.place.landmark
        ? (() => {
            const key = landmarkKey(record.place.landmark);
            return pickedLandmarkKeys.has(key) || landmarkOwners.get(key).length > 1;
          })()
        : false);

    return {
      id,
      agent,
      ...record,
      family: deriveFamily(record, vocab),
      region: record.place && record.place.country_code ? regionOf(record.place.country_code, vocab) : undefined,
      primary_subject: record.subject ? record.subject[0] : undefined,
      second_image: secondImage,
    };
  });

  docs.sort((a, b) => (b.screen && b.screen.score ? b.screen.score : 0) - (a.screen && a.screen.score ? a.screen.score : 0) || a.id.localeCompare(b.id));

  mkdirSync(outDir, { recursive: true });
  for (const doc of docs) {
    writeFileSync(path.join(outDir, `${doc.id}.json`), `${JSON.stringify(doc, null, 2)}\n`);
  }

  return { docs, dropped };
}

// --- CLI ---------------------------------------------------------------

function printErrors(errors) {
  for (const e of errors) {
    process.stdout.write(`${e.file}: ${e.rule}: ${e.message}\n`);
  }
}

function parseFlags(args) {
  const flags = { _: [] };
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === '--root') flags.root = args[++i];
    else if (arg === '--out') flags.out = args[++i];
    else if (arg === '--json') flags.json = true;
    else flags._.push(arg);
  }
  return flags;
}

function main(argv) {
  const [mode, ...rest] = argv;
  const flags = parseFlags(rest);

  try {
    if (mode === 'integrity') {
      const errors = runIntegrity(flags.root || 'photos');
      printErrors(errors);
      process.exit(errors.length === 0 ? 0 : 1);
    } else if (mode === 'balance') {
      const { report, breaches } = runBalance(flags.root || 'photos');
      if (flags.json) {
        process.stdout.write(`${JSON.stringify(report, null, 2)}\n`);
      } else {
        for (const b of breaches) process.stdout.write(`${b.rule}: ${b.key}: ${b.message}\n`);
      }
      process.exit(breaches.length === 0 ? 0 : 1);
    } else if (mode === 'candidate') {
      const target = flags._[0];
      if (!target) throw new UnreadableInputError('candidate requires a file or directory argument');
      const errors = runCandidate(target);
      printErrors(errors);
      process.exit(errors.length === 0 ? 0 : 1);
    } else if (mode === 'sheet') {
      if (!flags.out) throw new UnreadableInputError('sheet requires --out');
      const { docs, dropped } = runSheet(flags.root || 'photos', flags.out);
      process.stdout.write(`sheet: wrote ${docs.length} candidate(s), dropped ${dropped.length} duplicate(s)\n`);
      for (const d of dropped) process.stdout.write(`${d.id}: duplicate: dropped in favor of an earlier agent's copy\n`);
      process.exit(0);
    } else {
      process.stderr.write(`usage: check.mjs <integrity|balance|candidate|sheet> [options]\n`);
      process.exit(2);
    }
  } catch (err) {
    if (err instanceof UnreadableInputError || err.code === 'ENOENT') {
      process.stderr.write(`error: ${err.message}\n`);
      process.exit(2);
    }
    throw err;
  }
}

main(process.argv.slice(2));
