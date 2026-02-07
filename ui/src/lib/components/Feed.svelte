<script lang="ts">
  import type { FeedData } from '$lib/types';
  import Task from './Task.svelte';
  import { slide } from '$lib/utils/transitions';
  import { appState } from '$lib/stores/app.svelte';

  interface Props {
    feedData: FeedData;
  }

  let { feedData }: Props = $props();

  // Derived Values
  const projects = $derived(Object.values(feedData?.projects || {}));

  // Unified Active Tasks (combines reviews, in_progress, planning, draft)
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

    if (appState.activeWorkspaceId) {
      const project = projects.find((p) => p.id === appState.activeWorkspaceId);
      if (project) {
        tasks = tasks.filter(
          (t) => t.workspace_id === project.id || t.workspace_name === project.name
        );
      }
    }
    return tasks;
  });
</script>

<div id="feed-content" class="flex flex-col gap-6">
  <!-- Task List -->
  <div class="space-y-3 min-h-[300px]">
    {#key appState.activeWorkspaceId}
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
          <h3 class="text-base font-medium text-zinc-300 mb-1">All Caught Up</h3>
          <p class="text-sm text-zinc-500 max-w-[200px]">
            {appState.activeWorkspaceId
              ? 'No draft items in this workspace'
              : 'Your inbox is empty'}
          </p>
        </div>
      {/if}
    {/key}
  </div>
</div>
