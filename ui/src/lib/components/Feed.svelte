<script lang="ts">
  import type { FeedData } from '$lib/types';
  import Task from './Task.svelte';
  import { slide, DURATIONS } from '$lib/utils/transitions';
  import * as Tabs from '$lib/components/ui/tabs';

  interface Props {
    feedData: FeedData;
  }

  let { feedData }: Props = $props();

  // View state
  let view = $state('active');

  // Unified Active Tasks (combines reviews, in_progress, planning, pending)

  const activeTasks = $derived.by(() => {
    const all = [
      ...(feedData?.reviews || []),
      ...(feedData?.active || []),
      ...(feedData?.planning || []),
    ];
    return all.slice().sort((a, b) => b.updated_at - a.updated_at);
  });

  const doneTasks = $derived(
    [...(feedData?.done || [])].sort((a, b) => b.updated_at - a.updated_at)
  );

  const currentTasks = $derived(view === 'active' ? activeTasks : doneTasks);
</script>

<div id="feed-content" class="flex flex-col gap-6">
  <!-- Custom Segmented Control -->
  <div class="flex justify-center">
    <Tabs.Root bind:value={view} class="w-full max-w-[400px]">
      <Tabs.List
        class="grid w-full grid-cols-2 rounded-full bg-secondary/50 p-1 h-10 border border-border/40 backdrop-blur-sm"
      >
        <Tabs.Trigger
          value="active"
          class="rounded-full data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm transition-all duration-300 font-medium text-xs tracking-wide"
        >
          ACTIVE ({activeTasks.length})
        </Tabs.Trigger>
        <Tabs.Trigger
          value="completed"
          class="rounded-full data-[state=active]:bg-background data-[state=active]:text-foreground data-[state=active]:shadow-sm transition-all duration-300 font-medium text-xs tracking-wide"
        >
          COMPLETED ({doneTasks.length})
        </Tabs.Trigger>
      </Tabs.List>
    </Tabs.Root>
  </div>

  <!-- Task List -->
  <div class="space-y-3 min-h-[300px]">
    {#if currentTasks.length > 0}
      {#each currentTasks as task, index (task.id)}
        <div
          in:slide|local={{
            direction: 'up',
            duration: DURATIONS.normal,
            delay: index * 50,
          }}
        >
          <Task {task} variant={task.status} />
        </div>
      {/each}
    {:else}
      <div
        class="flex flex-col items-center justify-center py-12 px-4 text-center"
        transition:slide|local
      >
        <div class="rounded-full bg-secondary/50 p-4 mb-3">
          {#if view === 'active'}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-muted-foreground"><path d="M22 12h-4l-3 9L9 3l-3 9H2" /></svg
            >
          {:else}
            <svg
              xmlns="http://www.w3.org/2000/svg"
              width="24"
              height="24"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              stroke-width="2"
              stroke-linecap="round"
              stroke-linejoin="round"
              class="text-muted-foreground"><polyline points="20 6 9 17 4 12" /></svg
            >
          {/if}
        </div>
        <p class="text-sm text-muted-foreground font-medium">
          {view === 'active' ? 'No active agents running' : 'No completed tasks yet'}
        </p>
      </div>
    {/if}
  </div>
</div>
