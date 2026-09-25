// adapter-static has no `fallback`, so every route must prerender;
// `ssr = false` keeps the client bundle in charge of navigation, since
// the Go server never runs SvelteKit's Node renderer.
export const prerender = true;
export const ssr = false;
