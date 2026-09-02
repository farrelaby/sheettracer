<script lang="ts">
  import { fly } from "svelte/transition";
  import { toast, type ToastPosition } from "@/lib/toast.svelte";

  const kindStyles: Record<string, string> = {
    success: "border-l-emerald-400",
    error: "border-l-red-400",
    info: "border-l-slate-400",
  };

  const iconPaths: Record<string, string> = {
    success:
      '<path d="M20 6 9 17l-5-5" stroke-linecap="round" stroke-linejoin="round"></path>',
    error:
      '<circle cx="12" cy="12" r="10"></circle><path d="m15 9-6 6" stroke-linecap="round" stroke-linejoin="round"></path><path d="m9 9 6 6" stroke-linecap="round" stroke-linejoin="round"></path>',
    info: '<circle cx="12" cy="12" r="10"></circle><path d="M12 16v-4" stroke-linecap="round" stroke-linejoin="round"></path><path d="M12 8h.01" stroke-linecap="round" stroke-linejoin="round"></path>',
  };

  const containerClass: Record<ToastPosition, string> = {
    "bottom-right": "bottom-4 right-4",
    "bottom-left": "bottom-4 left-4",
    "top-right": "top-4 right-4",
    "top-left": "top-4 left-4",
    "bottom-center": "bottom-4 left-1/2 -translate-x-1/2",
    "top-center": "top-4 left-1/2 -translate-x-1/2",
  };

  const flyDir: Record<ToastPosition, { x?: number; y?: number }> = {
    "bottom-right": { y: 20 },
    "bottom-left": { y: 20 },
    "top-right": { y: -20 },
    "top-left": { y: -20 },
    "bottom-center": { y: 20 },
    "top-center": { y: -20 },
  };

  let pos = $derived(toast.position);
</script>

<div
  class="pointer-events-none fixed z-50 flex flex-col gap-2 {containerClass[
    pos
  ]}"
>
  {#each toast.toasts as t (t.id)}
    <div
      transition:fly={{ ...flyDir[pos], duration: 200 }}
      class="pointer-events-auto flex w-72 items-start gap-2.5 rounded-md border border-slate-800 bg-slate-900/90 px-3 py-2.5 shadow-lg backdrop-blur {kindStyles[
        t.kind
      ]}"
    >
      <svg
        class="mt-0.5 h-4 w-4 shrink-0 text-{t.kind === 'success'
          ? 'emerald-400'
          : t.kind === 'error'
            ? 'red-400'
            : 'slate-400'}"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        {@html iconPaths[t.kind]}
      </svg>
      <p class="min-w-0 flex-1 text-xs leading-relaxed text-slate-200">
        {t.message}
      </p>
      <button
        onclick={() => toast.dismiss(t.id)}
        aria-label="Dismiss"
        class="shrink-0 rounded p-0.5 text-slate-500 hover:text-slate-300"
      >
        <svg
          class="h-3 w-3"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"><path d="M18 6 6 18M6 6l12 12"></path></svg
        >
      </button>
    </div>
  {/each}
</div>
