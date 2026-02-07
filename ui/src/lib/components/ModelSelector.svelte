<script lang="ts">
  import { MODELS } from '$lib/types';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import ZapIcon from '@lucide/svelte/icons/zap';
  import { dropdownPop } from '$lib/utils/transitions';

  interface Props {
    position?: 'left' | 'right';
  }

  let { position = 'left' }: Props = $props();

  let modelOpen = $state(false);
  let modelMenuRef = $state<HTMLDivElement | null>(null);

  const canSelectModel = $derived(
    !appState.settings || appState.settings.agent_backend === 'native'
  );

  $effect(() => {
    if (!canSelectModel && modelOpen) {
      modelOpen = false;
    }
  });

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as Node;
    if (modelOpen && modelMenuRef && !modelMenuRef.contains(target)) {
      modelOpen = false;
    }
  }

  $effect(() => {
    document.addEventListener('click', handleClickOutside);
    return () => document.removeEventListener('click', handleClickOutside);
  });
</script>

{#if canSelectModel}
  <div class="relative" bind:this={modelMenuRef}>
    <button
      type="button"
      onclick={() => (modelOpen = !modelOpen)}
      aria-label="Select model"
      class="w-10 h-10 flex items-center justify-center rounded-xl text-gray-400 active:text-white active:bg-white/10 transition-all duration-150"
    >
      <ZapIcon class="w-5 h-5" />
    </button>
    {#if modelOpen}
      <div
        transition:dropdownPop|local
        class={cn(
          'absolute bottom-full mb-3 w-52 bg-popover/95 border border-white/10 rounded-xl shadow-2xl overflow-hidden py-1 z-50 backdrop-blur-xl',
          position === 'left' ? 'left-0' : 'right-0'
        )}
      >
        <div
          class="px-3 py-2 text-[10px] text-gray-500 font-bold uppercase tracking-wider border-b border-white/5 mb-1"
        >
          Select Model
        </div>
        <div class="p-1.5 space-y-0.5">
          {#each MODELS as m}
            <button
              type="button"
              onclick={() => {
                appState.setModel(m.id);
                modelOpen = false;
              }}
              class={cn(
                'w-full flex items-center justify-between px-3 py-2 rounded-lg cursor-pointer transition text-left group',
                appState.activeModelId === m.id
                  ? 'bg-primary/10 text-primary border border-primary/20'
                  : 'hover:bg-white/5 text-gray-400 hover:text-white border border-transparent'
              )}
            >
              <span class="text-sm font-medium">{m.name}</span>
              {#if appState.activeModelId === m.id}
                <div
                  class="w-1.5 h-1.5 rounded-full bg-primary shadow-[0_0_8px_hsla(270,72%,78%,0.8)]"
                ></div>
              {/if}
            </button>
          {/each}
        </div>
      </div>
    {/if}
  </div>
{/if}
