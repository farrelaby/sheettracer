<script lang="ts">
  import type { Snippet } from "svelte";
  import { Dialog, type WithoutChild } from "bits-ui";
  import { fly, fade } from "svelte/transition";
  import { untrack } from "svelte";
  import { dialogTracker } from "@/lib/stores/dialog.svelte";

  type Props = Dialog.RootProps & {
    title: Snippet;
    description: Snippet;
    contentProps?: WithoutChild<Dialog.ContentProps>;
    open?: boolean;
    confirmLabel?: string;
    danger?: boolean;
    onConfirm?: () => void;
  };

  let {
    open = $bindable(false),
    children,
    contentProps,
    title,
    description,
    confirmLabel = "Confirm",
    danger = false,
    onConfirm,
    ...restProps
  }: Props = $props();

  $effect(() => {
    if (!open) return;
    untrack(() => dialogTracker.push());
    return () => untrack(() => dialogTracker.pop());
  });

  function handleConfirm() {
    onConfirm?.();
    open = false;
  }
</script>

<Dialog.Root bind:open {...restProps}>
  <Dialog.Portal>
    <Dialog.Overlay forceMount>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            transition:fade={{ duration: 150 }}
            class="fixed inset-0 z-40 bg-black/70 backdrop-blur-sm"
          ></div>
        {/if}
      {/snippet}
    </Dialog.Overlay>
    <Dialog.Content forceMount {...contentProps}>
      {#snippet child({ props, open: isOpen })}
        {#if isOpen}
          <div
            {...props}
            transition:fly={{ y: 12, duration: 180 }}
            class="fixed left-1/2 top-1/2 z-50 w-full max-w-sm -translate-x-1/2 -translate-y-1/2 rounded-lg border border-slate-800 bg-slate-900 p-5 shadow-xl"
          >
            <Dialog.Title class="text-sm font-semibold text-slate-100">
              {@render title()}
            </Dialog.Title>
            <Dialog.Description
              class="mt-1.5 text-xs leading-relaxed text-slate-400"
            >
              {@render description()}
            </Dialog.Description>
            {@render children?.()}
            <div class="mt-5 flex justify-end gap-2">
              <Dialog.Close
                class="rounded-md border border-slate-700 px-3 py-1.5 text-xs text-slate-300 hover:bg-slate-800"
              >
                Cancel
              </Dialog.Close>
              <button
                onclick={handleConfirm}
                class={danger
                  ? "rounded-md bg-red-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-red-500"
                  : "rounded-md bg-purple-600 px-3 py-1.5 text-xs font-medium text-white hover:bg-purple-500"}
              >
                {confirmLabel}
              </button>
            </div>
            <Dialog.Close
              aria-label="Close"
              class="absolute right-3 top-3 rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-slate-200"
            >
              <svg
                class="h-4 w-4"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
                ><path d="M18 6 6 18M6 6l12 12"></path></svg
              >
            </Dialog.Close>
          </div>
        {/if}
      {/snippet}
    </Dialog.Content>
  </Dialog.Portal>
</Dialog.Root>
