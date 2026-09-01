/// <reference types="vitest/config" />
import { defineConfig } from 'vite';
import { svelte } from '@sveltejs/vite-plugin-svelte';
import tailwindcss from '@tailwindcss/vite';

// dubplate serve's basePath constant (internal/serve/serve.go) mounts
// every route -- the embedded app, /api/*, and tus's /files/ -- under
// this exact prefix; base here must match or every emitted asset href
//404s the instant it's reached through Access (plan-time decision,
// 2026-09-01 conductor ruling).
const basePath = '/upload/';

// The port `dubplate serve --dev-identity <email>` binds locally in the
// documented dev loop (DUBPLATE_SERVE_PORT in a workstation .env, not
// hardcoded on the Go side -- adjust here to match your own).
const devBackend = 'http://localhost:8080';

export default defineConfig(({ mode }) => ({
  base: basePath,
  plugins: [svelte(), tailwindcss()],
  // Under Vitest (which runs with mode "test") only: its default
  // "node" resolution condition would otherwise pick Svelte's
  // server-side runtime (`mount()` throws lifecycle_function_unavailable)
  // instead of the client one component tests need. Scoped to test-only
  // so the real production build keeps Vite's own default conditions.
  resolve: mode === 'test' ? { conditions: ['browser'] } : undefined,
  build: {
    outDir: 'dist',
    emptyOutDir: true,
  },
  server: {
    proxy: {
      // Matches internal/serve/serve.go's buildMux routes exactly:
      // POST /api/session[/*], GET /api/journey/*, GET /api/history,
      // GET /api/config, POST /api/stream/probe all live under
      // /upload/api/; tus's PATCH/HEAD/DELETE protocol lives under
      // /upload/files/. Plain path-based proxying is enough for both --
      // tus is ordinary HTTP, and Vite's proxy streams SSE through
      // unmodified.
      '/upload/api': { target: devBackend, changeOrigin: true },
      '/upload/files': { target: devBackend, changeOrigin: true },
    },
  },
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/setupTests.ts'],
  },
}));
