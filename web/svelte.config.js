import adapter from '@sveltejs/adapter-static';
import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

/** @type {import('@sveltejs/kit').Config} */
const config = {
  preprocess: vitePreprocess(),
  kit: {
    // No `fallback`: every route prerenders (root +layout.ts sets
    // `prerender = true`), so adapter-static's default `strict: true`
    // has nothing dynamic to reject.
    adapter: adapter(),
    csp: {
      mode: 'hash',
      directives: {
        'default-src': ['self'],
        'style-src': ['self', 'unsafe-inline'],
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
