<script lang="ts">
  import VisibilityIcon from "./VisibilityIcon.svelte";
  import { seedSheets } from "./demo-data";
  import type { SheetRow } from "./demo-data";
  import { fly } from "svelte/transition";
  import { selection, type Selection } from "../lib/selection.svelte";
  import OAuthStatus from "./OAuthStatus.svelte";
  import { OAuthService } from "../../bindings/sheettracer/internal/services";

  interface Props {
    onClose: () => void;
  }

  let { onClose }: Props = $props();

  const MIN = 260;
  const MAX = 480;

  function initialWidth(): number {
    const saved = Number(localStorage.getItem("st.sidebarWidth"));
    return Number.isFinite(saved) ? Math.min(MAX, Math.max(MIN, saved)) : 340;
  }

  let width = $state(initialWidth());
  let dragging = $state(false);

  let sheets = $state<SheetRow[]>(seedSheets);
  let query = $state("");
  let linkInput = $state("");
  let addError = $state("");
  let settingsOpen = $state(false);
  let rescanOnLaunch = $state(true);

  const filtered = $derived(
    sheets.filter((s) =>
      s.title.toLowerCase().includes(query.trim().toLowerCase()),
    ),
  );

  function clampWidth(w: number): number {
    return Math.min(MAX, Math.max(MIN, w));
  }

  function onHandleDown(e: PointerEvent) {
    dragging = true;
    (e.currentTarget as HTMLDivElement).setPointerCapture(e.pointerId);
  }

  function onHandleMove(e: PointerEvent) {
    if (dragging) width = clampWidth(e.clientX);
  }

  function onHandleUp() {
    if (!dragging) return;
    dragging = false;
    localStorage.setItem("st.sidebarWidth", String(width));
  }

  function onSelect(s: Selection) {
    selection.set(s);
  }

  function addSheet() {
    const value = linkInput.trim();
    if (!value) return;
    if (!value.startsWith("http")) {
      addError =
        "Enter a Google Sheets link, e.g. https://docs.google.com/spreadsheets/d/…";
      return;
    }
    if (value.includes("denied")) {
      addError =
        "Could not access this sheet — no access, or it doesn't exist.";
      return;
    }
    addError = "";
    sheets = [
      {
        id: `new-${Date.now()}`,
        title: "Just added sheet",
        visibility: "unknown",
        lastScan: "never",
        status: "ok",
      },
      ...sheets,
    ];
    linkInput = "";
  }

  function rescan(id: string) {
    sheets = sheets.map((s) =>
      s.id === id ? { ...s, status: "scanning" } : s,
    );
    setTimeout(() => {
      sheets = sheets.map((s) =>
        s.id === id ? { ...s, status: "ok", lastScan: "just now" } : s,
      );
    }, 1200);
  }

  function remove(id: string) {
    sheets = sheets.filter((s) => s.id !== id);
  }

  function statusDot(status: SheetRow["status"]): string {
    switch (status) {
      case "scanning":
        return "bg-amber-400 animate-pulse";
      case "error":
        return "bg-red-400";
      case "partial":
        return "bg-amber-400";
      default:
        return "bg-emerald-400";
    }
  }
</script>

<aside
  transition:fly={{ x: -200, duration: 200 }}
  class="absolute inset-y-0 left-0 z-20 flex flex-col border-r border-slate-800 bg-slate-900/85 backdrop-blur"
  style={`width: ${width}px`}
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
      addSheet();
    }}
    class="mx-2 mb-2 border border-dashed rounded p-2"
  >
    <input
      bind:value={linkInput}
      placeholder="Paste a Google Sheets link…"
      class="w-full rounded-md border border-slate-700 bg-slate-950/60 px-3 py-1.5 text-sm text-slate-200 placeholder-slate-600 outline-none focus:border-purple-500"
    />
    {#if addError}<p class="mt-1.5 text-xs text-red-400">{addError}</p>{/if}
    <button
      type="submit"
      disabled={!linkInput.trim()}
      class="mt-2 w-full rounded-md bg-purple-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-purple-500 disabled:opacity-40"
      >Add spreadsheet</button
    >
  </form>

  <input
    bind:value={query}
    placeholder="Search sheets…"
    class="mx-4 mb-2 w-[calc(100%-2rem)] rounded-md border border-slate-800 bg-slate-950/50 px-3 py-1.5 text-sm text-slate-200 placeholder-slate-600 outline-none focus:border-slate-600"
  />

  <nav class="min-h-0 flex-1 overflow-y-auto px-2 py-4">
    {#if filtered.length === 0}
      <p class="px-3 py-6 text-center text-xs text-slate-600">
        No tracked sheets yet.
      </p>
    {:else}
      {#each filtered as s (s.id)}
        <div
          role="button"
          tabindex="0"
          onclick={() => onSelect({ type: "node", id: s.id })}
          onkeydown={(e) => {
            if (e.key === "Enter" || e.key === " ") {
              e.preventDefault();
              onSelect({ type: "node", id: s.id });
            }
          }}
          class={[
            "group mb-1 flex w-full cursor-pointer items-center gap-2.5 rounded-md border border-transparent px-2.5 py-2 text-left hover:border-slate-800 hover:bg-slate-800/40 transition-all",
            ,
            selection.value && selection.value.id === s.id && "bg-slate-800",
          ]}
        >
          <VisibilityIcon visibility={s.visibility} />
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm text-slate-200">{s.title}</p>
            <p class="flex items-center gap-1.5 text-xs text-slate-500">
              <span
                class={`h-1.5 w-1.5 shrink-0 rounded-full ${statusDot(s.status)}`}
              ></span>
              {s.status === "scanning" ? "scanning…" : s.lastScan}
            </p>
          </div>
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              rescan(s.id);
            }}
            title="Rescan"
            class="rounded p-1 text-slate-500 opacity-0 hover:bg-slate-700 hover:text-green-300 group-hover:opacity-100 cursor-pointer"
            >⟳</button
          >
          <button
            type="button"
            onclick={(e) => {
              e.stopPropagation();
              remove(s.id);
            }}
            title="Remove"
            class="rounded p-1 px-2 font-extrabold text-slate-500 opacity-0 hover:bg-red-700 hover:text-white group-hover:opacity-100 cursor-pointer"
            >✕</button
          >
        </div>
      {/each}
    {/if}
  </nav>

  <div class="mx-4 mb-3">
    <div class="mb-1 flex items-center justify-between text-xs text-slate-500">
      <span>Scanning 2 of 7 sheets…</span>
      <span>28%</span>
    </div>
    <div class="h-1 overflow-hidden rounded-full bg-slate-800">
      <div class="h-full w-[28%] rounded-full bg-purple-500"></div>
    </div>
  </div>

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
        >{settingsOpen ? "▾" : "▸"}</span
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
          onclick={() => { OAuthService.Connect(); }}
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
    onpointerdown={onHandleDown}
    onpointermove={onHandleMove}
    onpointerup={onHandleUp}
    onpointercancel={onHandleUp}
    class="absolute inset-y-0 -right-1.5 z-10 w-3 cursor-ew-resize"
  ></div>
</aside>

<style>
</style>
