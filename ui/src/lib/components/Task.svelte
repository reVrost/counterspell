<script lang="ts">
  import { goto } from '$app/navigation';
  import { FolderIcon } from '@lucide/svelte';
  import { Check, ChevronRight, X } from '@lucide/svelte';
  import { cn } from 'tailwind-variants';

  interface Task {
    id: string;
    title: string;
    repository_name?: string;
    status?: string;
    last_assistant_message?: string;
    updated_at: number;
  }

  import { formatRelativeTime } from '$lib/utils';

  interface Props {
    task: Task;
    variant: 'pending' | 'active' | 'review' | 'completed' | 'planning';
  }

  let { task, variant }: Props = $props();

  let elapsed = $state(0);

  function handleClick() {
    goto(`/tasks/${task.id}`);
  }

  const baseClasses =
    'w-full text-left bg-card border rounded-sm p-4 shadow-sm focus:outline-none focus:ring-2 transition-all duration-200 ease-in-out active:scale-[0.98]';

  const variantClasses = {
    pending:
      'border-gray-700/50 hover:border-primary/30 hover:bg-primary/5 focus:ring-gray-500/50',
    planning:
      'border-purple-900/50 hover:border-primary/40 hover:bg-primary/5 focus:ring-purple-500/50',
    active:
      'border-gray-800/50 hover:border-primary/30 hover:bg-primary/5 focus:ring-orange-500/50',
    review:
      'border-gray-800 hover:border-primary/30 hover:bg-primary/5 focus:ring-blue-500/50',
    completed:
      'bg-card border-gray-600/20 hover:border-primary/30 hover:bg-primary/5 focus:ring-green-500/50',
  };
</script>

<button type="button" class="{baseClasses} {variantClasses[variant]}" onclick={handleClick}>
  <div class="flex justify-between items-start w-full gap-4">
    <!-- Left content (Shared) -->
    <div class="flex-1 min-w-0 space-y-1">
      <h4 class="text-base font-semibold text-gray-100 leading-tight truncate">
        {task.title}
      </h4>

      <div class="flex items-center gap-1.5 text-gray-500">
        <FolderIcon class="w-3 h-3" />
        <span class="text-xs font-medium truncate">{task.repository_name || 'Unknown'}</span>
      </div>

      {#if task.last_assistant_message}
        <p class="text-sm text-gray-400 leading-normal line-clamp-3 pt-1">
          {task.last_assistant_message}
        </p>
      {/if}
    </div>

    <!-- Right content (Variant specific) -->
    <div class="shrink-0 flex flex-col items-end gap-1.5 self-start pt-0.5">
      {#if variant === 'completed'}
        <div class="flex flex-col items-end gap-1.5">
          <div class="flex items-center gap-2">
            {#if task.status === 'failed'}
              <div
                class="w-5 h-5 rounded-full bg-red-950/40 border border-red-500/20 text-red-500 flex items-center justify-center shrink-0"
              >
                <X class="w-2.5 h-2.5" />
              </div>
            {:else}
              <div
                class="w-5 h-5 rounded-full bg-green-950/40 border border-green-500/20 text-green-500 flex items-center justify-center shrink-0"
              >
                <Check class="w-2.5 h-2.5" />
              </div>
            {/if}
            <ChevronRight class="w-3.5 h-3.5 text-gray-700" />
          </div>
          <span class="text-xs text-gray-500/70 font-medium tracking-tight"
            >{formatRelativeTime(task.updated_at)}</span
          >
        </div>
      {:else if variant === 'pending'}
        <div class="flex flex-col items-end gap-1.5">
          <span
            class="text-xs text-gray-500 px-2 py-0.5 rounded-full border border-gray-800 bg-gray-900/30 font-semibold uppercase tracking-wider"
          >
            Pending
          </span>
          <span class="text-xs text-gray-500/70 font-medium tracking-tight"
            >{formatRelativeTime(task.updated_at)}</span
          >
        </div>
      {:else if variant === 'planning'}
        <div class="flex flex-col items-end gap-1.5">
          <span
            class="text-xs text-purple-400 px-2 py-0.5 rounded-full border border-purple-900/40 bg-purple-950/20 font-semibold uppercase tracking-wider"
          >
            Planning
          </span>
          <span class="text-[10px] text-gray-500/70 font-medium tracking-tight"
            >{formatRelativeTime(task.updated_at)}</span
          >
        </div>
      {:else if variant === 'active'}
        <div class="flex flex-col items-end gap-1.5">
          <span
            class="text-xs text-orange-400 px-2 py-0.5 rounded-full border border-orange-900/40 bg-orange-950/20 font-semibold uppercase tracking-wider whitespace-nowrap"
          >
            In Progress
          </span>
          <div class="flex items-center gap-1.5">
            <span class="text-xs text-gray-500/70 font-medium tracking-tight"
              >{formatRelativeTime(task.updated_at)}</span
            >
            <span class="text-xs text-orange-500/40 font-mono tabular-nums">· {elapsed}s</span>
          </div>
        </div>
      {:else if variant === 'review'}
        <div class="flex flex-col items-end gap-1.5">
          <div class="flex items-center gap-2">
            <span
              class="text-xs text-blue-400 px-2 py-0.5 rounded-full border border-blue-900/40 bg-blue-950/20 font-semibold uppercase tracking-wider"
            >
              Review
            </span>
            <ChevronRight class="w-3.5 h-3.5 text-gray-600" />
          </div>
          <span class="text-xs text-gray-500/70 font-medium tracking-tight"
            >{formatRelativeTime(task.updated_at)}</span
          >
        </div>
      {/if}
    </div>
  </div>
</button>
