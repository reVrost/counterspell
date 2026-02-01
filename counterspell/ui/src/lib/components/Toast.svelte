<script lang="ts">
  import { appState } from "$lib/stores/app.svelte";
  import { slide, DURATIONS } from "$lib/utils/transitions";
  import CheckCircleIcon from "@lucide/svelte/icons/check-circle";
  import XCircleIcon from "@lucide/svelte/icons/x-circle";
  import InfoIcon from "@lucide/svelte/icons/info";

  const iconClasses = {
    success: "text-primary",
    error: "text-red-400",
    info: "text-blue-500",
  };

  const borderClasses = {
    success: "border-primary/30",
    error: "border-red-400/30",
    info: "border-blue-500/30",
  };
</script>

{#if appState.toastOpen}
  <div
    transition:slide|global={{ direction: "down", duration: DURATIONS.quick }}
    class="fixed top-6 left-1/2 -translate-x-1/2 z-[60] bg-gray-900 border text-white px-4 py-2 rounded-full shadow-2xl flex items-center gap-3 text-sm font-medium {borderClasses[
      appState.toastType
    ]}"
    role="alert"
    aria-live="polite"
  >
    {#if appState.toastType === "success"}
      <CheckCircleIcon class="w-4 h-4 {iconClasses.success}" />
    {:else if appState.toastType === "error"}
      <XCircleIcon class="w-4 h-4 {iconClasses.error}" />
    {:else}
      <InfoIcon class="w-4 h-4 {iconClasses.info}" />
    {/if}
    <span>{appState.toastMsg}</span>
  </div>
{/if}
