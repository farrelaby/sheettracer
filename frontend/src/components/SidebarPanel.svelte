<script lang="ts">
  import VisibilityIcon from "./VisibilityIcon.svelte";
  import { fly } from "svelte/transition";
  import { selectionStore, type Selection } from "@/lib/selection.svelte";
  import OAuthStatus from "./OAuthStatus.svelte";
  import { OAuthService } from "../../bindings/sheettracer/internal/services";
  import { spreadsheetStore } from "@/lib/spreadsheets.svelte";
  import { useSpreadsheetActions } from "@/lib/hooks/spreadsheetActions.svelte";
  import { useSidebarWidth } from "@/lib/hooks/sidebarWidth.svelte";
  import { lastScanText, statusDot } from "@/lib/utils/format";

  interface Props {
    onClose: () => void;
  }

  let { onClose }: Props = $props();

  const { addSheet, removeSheet, scanSpreadsheet } = useSpreadsheetActions();
  // const addSheet
  // const removeSheet
  // const scanSpreadsheet
  const sidebar = useSidebarWidth();

  let query = $state("");
  let linkInput = $state("");
  let settingsOpen = $state(false);
  let rescanOnLaunch = $state(true);

  let spreadsheets = $derived(spreadsheetStore.value);
  let loading = $derived(spreadsheetStore.loading);

  const filtered = $derived(
    spreadsheets.filter((s) =>
      s.Title.toLowerCase().includes(query.trim().toLowerCase()),
    ),
  );

  function onSelect(s: Selection) {
    selectionStore.set(s);
  }

  async function handleAdd() {
    if (await addSheet(linkInput)) linkInput = "";
  }
</script>

<aside
  transition:fly={{ x: -200, duration: 200 }}
  class="absolute inset-y-0 left-0 z-20 flex flex-col border-r border-slate-800 bg-slate-900/85 backdrop-blur"
  style={`width: ${sidebar.width}px`}
>
  <header class="flex items-center justify-between px-4 pb-3 pt-4">
    <h1
      class="flex items-center gap-2 text-sm font-semibold tracking-wide text-slate-100"
    >
      <svg
        class="h-4 w-4 text-purple-400"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        ><rect x="3" y="3" width="7" height="7" rx="1"></rect><rect
          x="14"
          y="3"
          width="7"
          height="7"
          rx="1"
        ></rect><rect x="14" y="14" width="7" height="7" rx="1"></rect><path
          d="M3 14h7v7H3z"
        ></path></svg
      >
      SheetTracer
    </h1>
    <button
      onclick={onClose}
      title="Collapse sidebar"
      class="rounded p-1 text-slate-500 hover:bg-slate-800 hover:text-slate-200"
    >
      <svg
        class="h-4 w-4"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"><path d="m15 18-6-6 6-6"></path></svg
      >
    </button>
  </header>

  <form
    onsubmit={(e) => {
      e.preventDefault();
      handleAdd();
    }}
    class="mx-2 mb-2 border border-dashed rounded p-2"
  >
    <input
      bind:value={linkInput}
      placeholder="Paste a Google Sheets link"
      class="w-full rounded-md border border-slate-700 bg-slate-950/60 px-3 py-1.5 text-sm text-slate-200 placeholder-slate-600 outline-none focus:border-purple-500"
    />
    <button
      type="submit"
      disabled={!linkInput.trim()}
      class="mt-2 w-full rounded-md bg-purple-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-purple-500 disabled:opacity-40"
      >Add spreadsheet</button
    >
  </form>

  <input
    bind:value={query}
    placeholder={"Search sheets"}
    class="mx-4 mb-2 w-[calc(100%-2rem)] rounded-md border border-slate-800 bg-slate-950/50 px-3 py-1.5 text-sm text-slate-200 placeholder-slate-600 outline-none focus:border-slate-600"
  />

  <nav class="min-h-0 flex-1 overflow-y-auto px-2 py-4">
    {#if loading}
      <p class="px-3 py-6 text-center text-xs text-slate-600">
        {"Loading \u2026"}
      </p>
    {:else if filtered.length === 0}
      <p class="px-3 py-6 text-center text-xs text-slate-600">
        No tracked sheets yet.
      </p>
    {:else}
      {#each filtered as s (s.ID)}
        <div
          role="button"
          tabindex="0"
          onclick={() => onSelect({ type: "node", id: String(s.ID) })}
          onkeydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              onSelect({ type: "node", id: String(s.ID) });
            }
          }}
          class={[
            "group mb-1 flex w-full cursor-pointer items-center gap-2.5 rounded-md  px-2.5 py-2 text-left border-l-2 border-l-slate-600 hover:border-l-slate-400 hover:shadow-2xl hover:bg-slate-800/40 transition-all",
            selectionStore.value &&
              selectionStore.value.id === String(s.ID) &&
              "bg-slate-800 border-l-slate-300!",
          ]}
        >
          <VisibilityIcon visibility={s.Visibility as any} />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm text-slate-200">{s.Title}</p>
            <p class="flex items-center gap-1.5 text-xs text-slate-500">
              <span
                class={`h-1.5 w-1.5 shrink-0 rounded-full ${statusDot(s.LastScanAt)}`}
              ></span>
              {lastScanText(s.LastScanAt)}
            </p>
          </div>
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              scanSpreadsheet(s.ID, s.Title);
            }}
            title="Scan"
            class="rounded p-1 px-2 text-xs text-slate-500 opacity-0 hover:bg-slate-700 hover:text-slate-300 group-hover:opacity-100 cursor-pointer"
            >{"\u25B6"}</button
          >
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              removeSheet(s.ID);
            }}
            title="Remove"
            class="rounded p-1 px-2 font-extrabold text-slate-500 opacity-0 hover:bg-red-700 hover:text-white group-hover:opacity-100 cursor-pointer"
            >{"\u2715"}</button
          >
        </div>
      {/each}
    {/if}
  </nav>

  <OAuthStatus />

  <div class="border-t border-slate-800 p-2">
    <button
      onclick={() => (settingsOpen = !settingsOpen)}
      class="flex w-full items-center gap-2 rounded-md px-2.5 py-2 text-sm text-slate-300 hover:bg-slate-800/60"
    >
      <svg
        class="h-4 w-4 text-slate-500"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
        ><circle cx="12" cy="12" r="3"></circle><path
          d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"
        ></path></svg
      >
      Settings
      <span class="ml-auto text-xs text-slate-600"
        >{settingsOpen ? "\u25BE" : "\u25B8"}</span
      >
    </button>
    {#if settingsOpen}
      <div class="space-y-1 px-2 py-1">
        <label
          class="flex items-center justify-between py-1 text-xs text-slate-400"
        >
          Rescan on launch
          <input
            type="checkbox"
            bind:checked={rescanOnLaunch}
            class="accent-purple-500"
          />
        </label>
        <button
          onclick={() => {
            OAuthService.Connect();
          }}
          class="w-full rounded px-2 py-1 text-left text-xs text-slate-500 hover:bg-slate-800 hover:text-slate-300"
          >Reconnect Google account</button
        >
        <p class="pt-1 text-[10px] text-slate-600">
          SheetTracer 0.1.0-prototype
        </p>
      </div>
    {/if}
  </div>

  <div
    role="separator"
    aria-orientation="vertical"
    onpointerdown={sidebar.onHandleDown}
    onpointermove={sidebar.onHandleMove}
    onpointerup={sidebar.onHandleUp}
    onpointercancel={sidebar.onHandleUp}
    class="absolute inset-y-0 -right-1.5 z-10 w-3 cursor-ew-resize"
  ></div>
</aside>
