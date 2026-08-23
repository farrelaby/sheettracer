<script lang="ts">
  import { Events } from "@wailsio/runtime";
  import { OAuthService } from "../../bindings/sheettracer/internal/services";
  import type { OAuthState } from "../lib/authstate";

  const initial: OAuthState = { connected: false, status: "disconnected" };
  let authState = $state<OAuthState>(initial);
  let busy = $state(false);

  $effect(() => {
    const off = Events.On("auth:state", (ev: any) => {
      if (ev?.data) {
        authState = ev.data;
        busy = false;
      }
    });
    return off;
  });

  $effect(() => {
    OAuthService.Status()
      .then((s: any) => {
        if (s) authState = s;
      })
      .catch(() => {});
  });

  async function connect() {
    busy = true;
    try {
      await OAuthService.Connect();
    } catch {
      busy = false;
    }
  }

  async function disconnect() {
    try {
      await OAuthService.Disconnect();
    } catch {}
  }
</script>

<div
  class="mx-2 mb-2 flex items-center gap-3 rounded-lg border border-slate-800 bg-slate-950/50 px-3 py-2.5"
>
  <div
    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-purple-500/20 text-xs font-semibold text-purple-300"
  >
    G
  </div>
  <div class="min-w-0 flex-1">
    <p class="truncate text-sm text-slate-200">Google account</p>
    {#if authState.status === "connected"}
      <p class="text-xs text-emerald-400">Connected</p>
    {:else if busy || authState.status === "connecting"}
      <p class="text-xs text-amber-400">Waiting for approval…</p>
    {:else if authState.status === "expired"}
      <p class="text-xs text-rose-400">{authState.message || "Session expired"}</p>
    {:else}
      <p class="text-xs text-slate-500">Not connected</p>
    {/if}
  </div>

  {#if authState.status === "connected"}
    <button
      onclick={disconnect}
      class="rounded border border-slate-700 px-2 py-1 text-xs text-slate-300 hover:bg-slate-800"
    >
      Disconnect
    </button>
  {:else if busy || authState.status === "connecting"}
    <span class="h-2 w-2 animate-pulse rounded-full bg-amber-400"></span>
  {:else}
    <button
      onclick={connect}
      disabled={busy}
      class="rounded border border-slate-700 px-2 py-1 text-xs text-slate-300 hover:bg-slate-800 disabled:opacity-50"
    >
      Connect
    </button>
  {/if}
</div>
