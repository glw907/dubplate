<!--
@component
The mark: a black record with a red label (visual identity spec, "The
mark"). It renders at one of three detail levels, chosen from `size`,
the mark's rendered pixel size: 56px and up gets `mark-large.svg` (with
track bands), 21 to 55px gets `mark-medium.svg` (grooves without
bands), and 20px and below gets `mark-small.svg` (the disc, band,
run-out, and label only). `alt` is always empty: no lettering appears
inside the mark, and every caller sets its own wordmark text beside it.
-->
<script lang="ts">
  import { asset } from '$app/paths';

  let { size }: { size: number } = $props();

  const src = $derived(
    size >= 56 ? asset('mark/mark-large.svg') : size >= 21 ? asset('mark/mark-medium.svg') : asset('mark/mark-small.svg')
  );
</script>

<img {src} alt="" width={size} height={size} />
