import { createHash } from 'node:crypto';
import { readdirSync, readFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

const webRoot = dirname(fileURLToPath(import.meta.url));

/**
 * Every file under `dir` (relative to web/), as a `/`-separated path
 * relative to web/.
 *
 * @param {string} dir
 * @returns {string[]}
 */
function filesUnder(dir) {
  return readdirSync(join(webRoot, dir), { withFileTypes: true }).flatMap((entry) => {
    const path = `${dir}/${entry.name}`;
    if (entry.isDirectory()) return filesUnder(path);
    return entry.isFile() ? [path] : [];
  });
}

/**
 * The build's version name: a SHA-256 over every input that shapes
 * the bundle, so two builds of the same source emit the same bytes.
 * Kit's default is the build's timestamp, which reaches a chunk and
 * so renames chunks and moves the inline script's CSP hash on every
 * build. The paths sort by their UTF-8 bytes, because `readdirSync`
 * promises no order and the order must not depend on the filesystem;
 * each path is hashed with its length-prefixed bytes, so no two input
 * sets share a digest by concatenation. The source is hashed rather
 * than asking git for a commit, so a source tarball without `.git`
 * still builds. `scripts/` counts as an input: Tailwind's automatic
 * source detection scans it, and the rendered leg's fixture markup
 * there is what puts the classes it measures into the bundle.
 *
 * @returns {string}
 */
function sourceVersion() {
  const inputs = [
    'package-lock.json',
    'svelte.config.js',
    'tsconfig.json',
    ...readdirSync(webRoot).filter((name) => name.startsWith('vite.config.')),
    ...filesUnder('scripts'),
    ...filesUnder('src'),
    ...filesUnder('static')
  ];
  inputs.sort((a, b) => Buffer.compare(Buffer.from(a), Buffer.from(b)));
  const hash = createHash('sha256');
  for (const path of inputs) {
    const bytes = readFileSync(join(webRoot, path));
    hash.update(`${path}\0${bytes.length}\0`);
    hash.update(bytes);
  }
  return hash.digest('hex');
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // No `fallback`: every route prerenders (root +layout.ts sets
    // `prerender = true`), so adapter-static's default `strict: true`
    // has nothing dynamic to reject.
    adapter: adapter(),
    version: { name: sourceVersion() },
    // The meta policy these directives write must match the header
    // internal/serve/csp.go sends, frame-ancestors aside (a meta policy
    // cannot carry it); a Go test over the built documents fails when
    // the two differ. img-src admits data: for DaisyUI's --noise texture.
    csp: {
      mode: 'hash',
      directives: {
        'default-src': ['self'],
        'style-src': ['self', 'unsafe-inline'],
        'img-src': ['self', 'data:'],
        'frame-ancestors': ['none'],
        'base-uri': ['none'],
        'object-src': ['none'],
        'form-action': ['self']
      }
    },
    typescript: {
      // The gate scripts under web/scripts/ are plain Node, outside
      // src/test/tests, so svelte-check's generated tsconfig never sees
      // them unless this hook appends them. `config.include.push`
      // mutates kit's generated list rather than replacing it, since a
      // child tsconfig's `include` replaces the one it extends instead
      // of merging.
      config: (config) => {
        config.include.push('../scripts/**/*.ts');
        config.compilerOptions.allowImportingTsExtensions = true;
        return config;
      }
    }
  }
};

export default config;
