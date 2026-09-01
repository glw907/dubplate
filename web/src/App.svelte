<!--
  @component
  The uploader's root shell. Mounted once from main.ts: loads
  `/api/config` into the shared {@link uploadStore} and lays down the
  DaisyUI navidrome theme (app.css) every later screen renders inside.
  It owns no upload logic of its own -- U10's selection UI and U11's
  journey spine are components that read uploadStore display-only, per
  the plan's "components are display-only" rule; this file is scaffold
  only, not a finished page.
-->
<script lang="ts">
	import { onMount } from 'svelte';
	import { uploadStore } from './lib/store/upload.svelte';

	onMount(() => {
		void uploadStore.loadConfig();
	});
</script>

<div class="bg-base-100 text-base-content min-h-screen">
	<header class="navbar bg-base-200">
		<span class="text-xl font-semibold uppercase tracking-wide">
			{uploadStore.config?.siteName ?? 'dubplate'}
		</span>
	</header>

	<main class="mx-auto max-w-2xl p-4">
		{#if uploadStore.signedOut}
			<div class="alert alert-warning" role="alert">
				<span>Your sign-in has expired. Reload the page to sign in again.</span>
			</div>
		{:else if uploadStore.config}
			<p class="text-base-content/70 text-sm">{uploadStore.config.formatSpec}</p>
		{/if}
	</main>
</div>
