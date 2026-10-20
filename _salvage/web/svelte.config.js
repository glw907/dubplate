import { vitePreprocess } from '@sveltejs/vite-plugin-svelte';

// No SvelteKit: this is a standalone Vite SPA (docs/research/
// passb-stack-brief.md section 1). vitePreprocess is the one thing a
// plain Vite + Svelte 5 + TypeScript setup still needs from the
// SvelteKit toolchain -- it lets <script lang="ts"> and any future
// <style lang="postcss"> block compile through Vite's own pipeline
// rather than Svelte's bundled (and more limited) preprocessor.
export default {
  preprocess: vitePreprocess(),
};
