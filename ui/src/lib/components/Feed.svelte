<script lang="ts">
  import type { FeedData } from "$lib/types";
  import Task from "./Task.svelte";
  import { slide, DURATIONS } from "$lib/utils/transitions";

  interface Props {
    feedData: FeedData;
  }

  let { feedData }: Props = $props();

  // Split active into pending and in_progress
  const activeTasks = $derived(feedData?.active || []);
  const planningTasks = $derived(feedData?.planning || []);
  const reviewsTasks = $derived(feedData?.reviews || []);
  const doneTasks = $derived(feedData?.done || []);

  const pendingTasks = $derived(
    activeTasks.filter((t) => t.status === "pending"),
  );
  const inProgressTasks = $derived(
    activeTasks.filter((t) => t.status === "in_progress"),
  );
</script>

<div id="feed-content">
  <!-- Review (needs user action) -->
  {#if reviewsTasks.length > 0}
    <div class="mb-6">
      <h3
        class="px-2 text-10 font-bold text-primary uppercase tracking-wider mb-3"
      >
        Review
      </h3>
      <div class="space-y-3">
        {#each reviewsTasks as task, index (task.id)}
          <div
            transition:slide|local={{
              direction: "up",
              duration: DURATIONS.normal,
              delay: index * 50,
            }}
          >
            <Task {task} variant="review" />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- In Progress -->
  {#if inProgressTasks.length > 0}
    <div class="mb-6">
      <h3
        class="px-2 text-sm font-bold text-orange-300 uppercase tracking-wider mb-3"
      >
        In Progress
      </h3>
      <div class="space-y-3">
        {#each inProgressTasks as task, index (task.id)}
          <div
            transition:slide|local={{
              direction: "up",
              duration: DURATIONS.normal,
              delay: index * 50,
            }}
          >
            <Task {task} variant="active" />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Planning -->
  {#if planningTasks.length > 0}
    <div class="mb-6">
      <h3
        class="px-2 text-sm font-bold text-primary uppercase tracking-wider mb-3"
      >
        Planning
      </h3>
      <div class="space-y-3">
        {#each planningTasks as task, index (task.id)}
          <div
            transition:slide|local={{
              direction: "up",
              duration: DURATIONS.normal,
              delay: index * 50,
            }}
          >
            <Task {task} variant="planning" />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- Pending -->
  {#if pendingTasks.length > 0}
    <div class="mb-6">
      <h3
        class="px-2 text-sm font-bold text-gray-500 uppercase tracking-wider mb-3"
      >
        Pending
      </h3>
      <div class="space-y-1">
        {#each pendingTasks as task, index (task.id)}
          <div
            transition:slide|local={{
              direction: "up",
              duration: DURATIONS.normal,
              delay: index * 50,
            }}
          >
            <Task {task} variant="pending" />
          </div>
        {/each}
      </div>
    </div>
  {/if}

  <!-- No active tasks message -->
  {#if activeTasks.length === 0 && reviewsTasks.length === 0}
    <div class="mb-6">
      <div class="px-2 py-2 text-sm text-gray-600 text-center">
        No active agents running
      </div>
    </div>
  {/if}

  <!-- Completed -->
  {#if doneTasks.length > 0}
    <div class="pt-4 border-t border-gray-800/50">
      <h3
        class="px-2 text-sm font-bold text-gray-600 uppercase tracking-wider mb-3"
      >
        Completed
      </h3>
      <div class="space-y-3">
        {#each doneTasks as task, index (task.id)}
          <div
            transition:slide|local={{
              direction: "up",
              duration: DURATIONS.normal,
              delay: index * 50,
            }}
          >
            <Task {task} variant="completed" />
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>
