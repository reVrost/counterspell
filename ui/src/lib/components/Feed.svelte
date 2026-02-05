<script lang="ts">
  import type { FeedData } from '$lib/types';
  import Task from './Task.svelte';
  import { slide } from '$lib/utils/transitions';
  import { appState } from '$lib/stores/app.svelte';

  interface Props {
    feedData: FeedData;
  }

  let { feedData }: Props = $props();

  // State
  let selectedWorkspaceId = $state<string | null>(null);

  // Derived Values
  const projects = $derived(Object.values(feedData?.projects || {}));

  // Unified Active Tasks (combines reviews, in_progress, planning, pending)
  const activeTasks = $derived.by(() => {
    const all = [
      ...(feedData?.reviews || []),
      ...(feedData?.active || []),
      ...(feedData?.planning || []),
    ];
    return all.slice().sort((a, b) => b.updated_at - a.updated_at);
  });

  // Filter tasks based on selected workspace
  const currentTasks = $derived.by(() => {
    let tasks = activeTasks;

    if (selectedWorkspaceId) {
      const project = projects.find((p) => p.id === selectedWorkspaceId);
      if (project) {
        tasks = tasks.filter(
          (t) => t.repository_id === project.id || t.repository_name === project.name
        );
      }
    }
    return tasks;
  });

  function handleAddWorkspace() {
    appState.projectMenuOpen = true;
  }
</script>

<div id="feed-content" class="flex flex-col gap-6">
  <!-- Workspace Pills Navigation -->
  <div class="flex items-center gap-2 overflow-x-auto scrollbar-hide py-1 mask-linear-fade">
    <button
      onclick={() => (selectedWorkspaceId = null)}
      class="h-8 px-4 rounded-full text-xs font-medium transition-all duration-200 border whitespace-nowrap
      {selectedWorkspaceId === null
        ? 'bg-violet-500/20 text-violet-200 border-violet-500/20 shadow-[0_0_10px_rgba(139,92,246,0.1)]'
        : 'bg-zinc-900/40 text-zinc-400 border-zinc-800/60 hover:bg-zinc-800/60 hover:text-zinc-200'}"
    >
      All
    </button>

    {#each projects as project (project.id)}
      <button
        onclick={() => (selectedWorkspaceId = project.id)}
        class="h-8 px-4 rounded-full text-xs font-medium transition-all duration-200 border whitespace-nowrap flex items-center gap-2
        {selectedWorkspaceId === project.id
          ? 'bg-violet-500/20 text-violet-200 border-violet-500/20 shadow-[0_0_10px_rgba(139,92,246,0.1)]'
          : 'bg-zinc-900/40 text-zinc-400 border-zinc-800/60 hover:bg-zinc-800/60 hover:text-zinc-200'}"
      >
        <div
          class="w-1.5 h-1.5 rounded-full"
          style="background-color: {project.color || '#A1A1AA'}"
        ></div>
        {project.name.split('/').pop()}
      </button>
    {/each}

    <button
      onclick={handleAddWorkspace}
      class="h-8 w-8 rounded-full border border-dashed border-zinc-700/60 text-zinc-500 hover:text-zinc-300 hover:border-zinc-500 flex items-center justify-center transition-all shrink-0 ml-1"
      title="Add Workspace"
    >
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
        stroke-linecap="round"
        stroke-linejoin="round"
      >
        <line x1="12" y1="5" x2="12" y2="19"></line>
        <line x1="5" y1="12" x2="19" y2="12"></line>
      </svg>
    </button>
  </div>

  <!-- Task List -->
  <div class="space-y-3 min-h-[300px]">
    {#key selectedWorkspaceId}
      {#if currentTasks.length > 0}
        {#each currentTasks as task, i (task.id)}
          <Task {task} variant={task.status} delay={i * 40} />
        {/each}
      {:else}
        <div
          class="flex flex-col items-center justify-center py-20 px-4 text-center opactiy-0 animate-in fade-in duration-500"
          in:slide|local
        >
          <div class="rounded-full bg-zinc-900/50 p-6 mb-4 ring-1 ring-white/5">
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-zinc-500"><polyline points="20 6 9 17 4 12" /></svg
            >
          </div>
          <h3 class="text-sm font-medium text-zinc-300 mb-1">All Caught Up</h3>
          <p class="text-xs text-zinc-500 max-w-[200px]">
            {selectedWorkspaceId ? 'No pending items in this workspace' : 'Your inbox is empty'}
          </p>
        </div>
      {/if}
    {/key}
  </div>
</div>
