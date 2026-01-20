<script lang="ts">
  import { appState } from "$lib/stores/app.svelte";
  import { taskTimers } from "$lib/stores/tasks.svelte";
  import { FolderIcon } from "@lucide/svelte";
  import { CheckIcon, ChevronRightIcon, Folder } from "lucide-react";
  import { cn } from "tailwind-variants";

  interface Task {
    id: string;
    title: string;
    repository_name?: string;
  }

  interface Props {
    task: Task;
    variant: "pending" | "active" | "review" | "completed" | "planning";
  }

  let { task, variant }: Props = $props();

  let elapsed = $state(0);

  // Hover prefetch
  let prefetchTimeout: number | null = null;

  function handleHover() {
    if (variant === "completed" || prefetchTimeout !== null) return;

    prefetchTimeout = window.setTimeout(() => {
      // Call global prefetch function from layout
      if ((window as any).prefetchTask) {
        (window as any).prefetchTask(task.id);
      }
    }, 150);
  }

  function handleMouseLeave() {
    if (prefetchTimeout) {
      clearTimeout(prefetchTimeout);
      prefetchTimeout = null;
    }
  }

  $effect(() => {
    if (variant === "active") {
      if (!taskTimers[task.id]) {
        taskTimers[task.id] = Date.now();
      }
      elapsed = Math.floor((Date.now() - taskTimers[task.id]) / 1000);

      const interval = setInterval(() => {
        elapsed = Math.floor((Date.now() - taskTimers[task.id]) / 1000);
      }, 1000);

      return () => {
        clearInterval(interval);
        if (prefetchTimeout) {
          clearTimeout(prefetchTimeout);
        }
      };
    }
  });

  function handleClick() {
    appState.openModal(task.id);
  }

  const baseClasses =
    "w-full text-left bg-card border rounded-2xl p-4 shadow-sm focus:outline-none focus:ring-2 transition-transform active:scale-[0.98]";

  const variantClasses = {
    pending:
      "border-gray-700/50 hover:border-gray-600/50 hover:shadow-md focus:ring-gray-500/50",
    planning:
      "border-purple-900/50 hover:border-purple-800/50 hover:shadow-md focus:ring-purple-500/50",
    active:
      "border-gray-800/50 hover:border-gray-700/50 hover:shadow-md focus:ring-orange-500/50",
    review:
      "border-blue-800 hover:border-blue-700/50 hover:shadow-lg focus:ring-blue-500/50",
    completed:
      "bg-card/60 border-gray-800/20 flex-row hover:border-gray-700/30 focus:ring-green-500/50",
  };
</script>

<button
  type="button"
  class="{baseClasses} {variantClasses[variant]}"
  onclick={handleClick}
  onmouseenter={handleHover}
  onmouseleave={handleMouseLeave}
>
  {#if variant === "completed"}
    <!-- Completed Layout -->
    <div class="flex justify-between items-center w-full">
      <div class="flex items-center gap-3">
        <div
          class="w-6 h-6 rounded-full bg-green-900/40 text-green-500 flex items-center justify-center text-xs shrink-0"
        >
          <CheckIcon class="w-3 h-3" />
        </div>
        <div class="min-w-0">
          <div class="text-sm text-gray-400 leading-snug line-clamp-2">
            {task.title}
          </div>
          <div class="text-xs text-gray-600 mt-0.5">
            {task.repository_name || "Unknown"}
          </div>
        </div>
      </div>
      <ChevronRightIcon class="w-4 h-4 text-gray-700 ml-3 shrink-0" />
    </div>
  {:else}
    <div class="flex justify-between items-start mb-2">
      <div class="flex items-center gap-2">
        <span
          class={cn(
            "text-gray-400 opacity-80 w-6 h-6 rounded-lg flex items-center justify-center text-xs",
            variant === "review" && "bg-gray-800/50 border border-gray-700/50",
          )}
        >
          <FolderIcon class="w-3 h-3" />
        </span>
        <span class="text-xs font-medium text-gray-400"
          >{task.repository_name ?? "Unknown"}</span
        >
      </div>

      {#if variant === "pending"}
        <span
          class="text-xs text-gray-400 bg-gray-500/10 px-2.5 py-1 rounded-lg font-medium border border-gray-500/20"
        >
          Pending
        </span>
      {:else if variant === "planning"}
        <span
          class="text-xs text-purple-400 bg-purple-500/10 px-2.5 py-1 rounded-lg font-medium border border-purple-500/20"
        >
          Planning
        </span>
      {:else if variant === "active"}
        <div class="flex items-center gap-2">
          <span
            class="text-xs text-orange-400 bg-orange-500/10 px-2.5 py-1 rounded-lg font-medium border border-orange-500/20"
          >
            In Progress
          </span>
          <span class="text-xs text-orange-400/80 font-mono tabular-nums"
            >{elapsed}s</span
          >
        </div>
      {:else if variant === "review"}
        <span
          class="text-xs text-blue-400 bg-blue-500/10 px-2.5 py-1 rounded-lg font-medium border border-blue-500/20"
        >
          Review
        </span>
      {/if}
    </div>

    <p
      class="text-sm text-gray-200 font-medium leading-snug line-clamp-2"
      class:pr-6={variant === "review"}
    >
      {task.title}
    </p>

    {#if variant === "review"}
      <div class="absolute right-4 top-1/2 -translate-y-1/2 text-gray-600">
        <ChevronRightIcon class="w-4 h-4" />
      </div>
    {/if}
  {/if}
</button>
