<script lang="ts">
  import type { Snippet } from 'svelte';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
  import CheckIcon from '@lucide/svelte/icons/check';
  import SearchIcon from '@lucide/svelte/icons/search';
  import PlusIcon from '@lucide/svelte/icons/plus';

  interface Props {
    open?: boolean;
    align?: 'start' | 'center' | 'end';
    sideOffset?: number;
    contentClass?: string;
    children?: Snippet;
  }

  let {
    open = $bindable(false),
    align = 'start',
    sideOffset = 8,
    contentClass = '',
    children,
  }: Props = $props();

  let workspaceSearch = $state('');

  const filteredWorkspaces = $derived(
    appState.workspaces.filter((p) => p.name.toLowerCase().includes(workspaceSearch.toLowerCase()))
  );

  $effect(() => {
    if (!open) {
      workspaceSearch = '';
    }
  });

  function openCreateWorkspaceModal() {
    open = false;
    appState.openNewWorkspaceModal();
  }
</script>

<DropdownMenu.Root bind:open>
  {@render children?.()}
  <DropdownMenu.Portal>
    <DropdownMenu.Content
      class={cn(
        'w-72 bg-zinc-950/95 backdrop-blur-xl border border-white/10 rounded-2xl shadow-[0_20px_50px_rgba(0,0,0,0.5)] overflow-hidden flex flex-col z-50 p-1',
        contentClass
      )}
      {align}
      {sideOffset}
    >
      <div class="p-2">
        <div class="relative">
          <SearchIcon class="absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 w-3.5 h-3.5" />
          <input
            bind:value={workspaceSearch}
            type="text"
            placeholder="Filter workspaces..."
            class="w-full bg-white/[0.03] border border-white/5 rounded-xl pl-9 pr-3 py-2 text-sm text-zinc-200 focus:outline-none focus:border-violet-500/50 focus:bg-white/[0.05] placeholder-zinc-600 transition-all"
          />
        </div>
      </div>

      <div class="max-h-[320px] overflow-y-auto py-1 custom-scrollbar">
        <DropdownMenu.Item
          onSelect={() => appState.setActiveWorkspace('', '')}
          class={cn(
            'w-full px-3 py-2.5 hover:bg-white/5 cursor-pointer rounded-lg flex items-center justify-between group transition text-left focus:bg-white/5 outline-none mb-0.5',
            !appState.activeWorkspaceId && 'bg-white/[0.08] text-white'
          )}
        >
          <span
            class="text-sm {appState.activeWorkspaceId
              ? 'text-zinc-400'
              : 'text-zinc-100'} group-hover:text-zinc-100 transition"
            >All Workspaces</span
          >
          {#if !appState.activeWorkspaceId}
            <CheckIcon class="w-3.5 h-3.5 text-violet-400" />
          {/if}
        </DropdownMenu.Item>

        {#each filteredWorkspaces as workspace}
          <DropdownMenu.Item
            onSelect={() => appState.setActiveWorkspace(workspace.id, workspace.name)}
            class={cn(
              'w-full px-3 py-2.5 hover:bg-white/5 cursor-pointer rounded-lg flex items-center gap-3 group transition text-left focus:bg-white/5 outline-none mb-0.5',
              appState.activeWorkspaceId === workspace.id && 'bg-white/[0.08] text-white'
            )}
          >
            <div
              class="w-6 h-6 rounded-md bg-zinc-900 border border-white/5 flex items-center justify-center shrink-0"
            >
              <span class="text-[10px] {workspace.color}">
                <i class="fas {workspace.icon}"></i>
              </span>
            </div>
            <div class="flex-1 min-w-0">
              <div class="text-sm text-zinc-400 group-hover:text-zinc-100 truncate transition">
                {workspace.name}
              </div>
            </div>
            {#if appState.activeWorkspaceId === workspace.id}
              <CheckIcon class="w-3.5 h-3.5 text-violet-400" />
            {/if}
          </DropdownMenu.Item>
        {/each}

        {#if filteredWorkspaces.length === 0}
          <div class="px-4 py-8 text-center text-zinc-600 text-sm">No workspaces found.</div>
        {/if}
      </div>

      <div
        class="mt-1 px-3 py-2 bg-white/[0.02] border-t border-white/5 text-sm text-zinc-500 flex justify-between items-center rounded-b-xl"
      >
        <span class="font-medium">{appState.workspaces.length} Workspaces</span>
        <button
          onclick={openCreateWorkspaceModal}
          class="hover:text-violet-400 font-semibold cursor-pointer flex items-center gap-1 transition-colors"
        >
          <PlusIcon class="w-3 h-3" /> New
        </button>
      </div>
    </DropdownMenu.Content>
  </DropdownMenu.Portal>
</DropdownMenu.Root>

<style>
  .custom-scrollbar::-webkit-scrollbar {
    width: 4px;
  }
  .custom-scrollbar::-webkit-scrollbar-track {
    background: transparent;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 10px;
  }
  .custom-scrollbar::-webkit-scrollbar-thumb:hover {
    background: rgba(255, 255, 255, 0.1);
  }
</style>
