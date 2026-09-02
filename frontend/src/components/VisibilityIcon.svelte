<script lang="ts">
  import type { Visibility } from "./demo-data";

  type Visibility = {};

  let { visibility }: { visibility: Visibility } = $props();

  const paths: Record<Visibility, string> = {
    private:
      '<rect x="3" y="11" width="18" height="11" rx="2"></rect><path d="M7 11V7a5 5 0 0 1 10 0v4"></path>',
    shared:
      '<path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"></path><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"></path>',
    editor:
      '<circle cx="12" cy="12" r="10"></circle><path d="M12 2a14.5 14.5 0 0 0 0 20 14.5 14.5 0 0 0 0-20"></path><path d="M2 12h20"></path>',
    unknown: '<path d="M5 12h14"></path>',
  };

  const color: Record<Visibility, string> = {
    private: "text-red-400",
    shared: "text-amber-400",
    editor: "text-emerald-400",
    unknown: "text-slate-500",
  };
</script>

<svg
  class={`h-4 w-4 shrink-0 ${color[visibility]}`}
  viewBox="0 0 24 24"
  fill="none"
  stroke="currentColor"
  stroke-width="2"
  stroke-linecap="round"
  stroke-linejoin="round"
  aria-hidden="true">{@html paths[visibility]}</svg
>
