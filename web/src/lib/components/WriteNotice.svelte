<!--
@component
The alert that reports an admin page's last write, or nothing when
there is no notice. A write refused for want of a recent sign-in adds a
sign-in link that opens in a new tab, so the pending change survives in
this one, and a control that runs the write again.
-->
<script lang="ts">
  import { resolve } from '$app/paths';
  import { REAUTH_SIGN_IN, ADMIN_RETRY } from '#lib/copy/deck.js';

  interface Props {
    notice: { text: string; alertClass: string; reauthenticate: boolean } | null;
    /** The write the retry control runs again. */
    retry: (() => Promise<void>) | null;
    /** The alert's id, which a refused field's `aria-describedby` names. */
    id: string;
    /** The text's id, set when the page moves focus to the text; it makes the text focusable. */
    textId?: string;
  }

  let { notice, retry, id, textId }: Props = $props();
</script>

{#if notice}
  <div {id} class="alert {notice.alertClass}">
    <div class="flex flex-col gap-2">
      {#if textId}
        <span id={textId} tabindex="-1">{notice.text}</span>
      {:else}
        <span>{notice.text}</span>
      {/if}
      {#if notice.reauthenticate}
        <div class="flex flex-wrap gap-2">
          <a href={resolve('login')} target="_blank" rel="noopener" class="btn">{REAUTH_SIGN_IN}</a>
          <button type="button" class="btn btn-primary" onclick={() => void retry?.()}>{ADMIN_RETRY}</button>
        </div>
      {/if}
    </div>
  </div>
{/if}
