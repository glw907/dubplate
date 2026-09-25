<!--
@component
The root layout. Every route imports its CSS bundle through here, since
a SvelteKit static build has exactly one entry that reaches every page.
-->
<script lang="ts">
  import '../app.css';

  let { children } = $props();
</script>

{@render children()}
