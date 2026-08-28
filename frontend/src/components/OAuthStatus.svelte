<script lang="ts">
  import { Events } from "@wailsio/runtime";
  import {
    OAuthService,
    type AccountInfo,
  } from "../../bindings/sheettracer/internal/services";
  import type { OAuthState } from "../lib/authstate";

  const initial: OAuthState = { connected: false, status: "disconnected" };
  let authState = $state<OAuthState>(initial);
  let busy = $state(false);

  let accountInfoState = $state<AccountInfo | null>(null);

  $effect(() => {
    const off = Events.On("auth:state", (ev) => {
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

  $effect(() => {
    if (authState.status === "connected") {
      OAuthService.Account().then((v) => {
        if (v)
          accountInfoState = {
            name: v.name,
            email: v.email,
            photo_url: v.photo_url,
          };
      });
    } else {
      accountInfoState = null;
    }
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
  <!-- <div
    class="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-purple-500/20 text-xs font-semibold text-purple-300"
  >
    G
  </div> -->
  {#if accountInfoState}
    <img
      src={accountInfoState.photo_url}
      alt="G"
      class="h-8 w-8 shrink-0 rounded-full object-cover"
    />
  {/if}
  <div class="min-w-0 flex-1">
    {#if accountInfoState}
      <p class="truncate text-sm text-slate-200">{accountInfoState.name}</p>
      <p class="text-[0.6rem] text-gray-400 max-w-lg overflow-x-hidden">
        {accountInfoState.email}
      </p>
    {/if}
    {#if authState.status === "connected"}
      <p class="text-xs text-emerald-400">Connected</p>
    {:else if busy || authState.status === "connecting"}
      <p class="text-xs text-amber-400">Waiting for approval…</p>
    {:else if authState.status === "expired"}
      <p class="text-xs text-rose-400">
        {authState.message || "Session expired"}
      </p>
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
