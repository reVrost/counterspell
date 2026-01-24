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
  }

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
    'w-full text-left bg-card border rounded-sm p-4 shadow-sm focus:outline-none focus:ring-2 transition-transform active:scale-[0.98]';

  const variantClasses = {
    pending: 'border-gray-700/50 hover:border-gray-600/50 hover:shadow-md focus:ring-gray-500/50',
    planning:
      'border-purple-900/50 hover:border-purple-800/50 hover:shadow-md focus:ring-purple-500/50',
    active: 'border-gray-800/50 hover:border-gray-700/50 hover:shadow-md focus:ring-orange-500/50',
    review: 'border-gray-800 hover:border-blue-700/50 hover:shadow-lg focus:ring-blue-500/50',
    completed:
      'bg-card border-gray-600/20 flex-row hover:border-gray-500/30 focus:ring-green-500/50',
  };
</script>

<button type="button" class="{baseClasses} {variantClasses[variant]}" onclick={handleClick}>
  {#if variant === 'completed'}
    <!-- Completed Layout -->
    <div class="flex justify-between items-center w-full">
      <div class="flex items-center gap-3">
        {#if task.status === 'failed'}
          <div
            class="w-6 h-6 rounded-full bg-red-900/40 text-red-500 flex items-center justify-center text-base shrink-0"
          >
            <X class="w-3 h-3" />
          </div>
        {:else}
          <div
            class="w-6 h-6 rounded-full bg-green-900/40 text-green-500 flex items-center justify-center text-base shrink-0"
          >
            <Check class="w-3 h-3" />
          </div>
        {/if}
        <div class="min-w-0">
          <div class="text-base leading-snug line-clamp-2">
            {task.title}
          </div>
          <div class="text-sm text-gray-400 mt-0.5">
            {task.repository_name || 'Unknown'}
          </div>
        </div>
      </div>
      <ChevronRight class="w-4 h-4 text-gray-700 ml-3 shrink-0" />
    </div>
  {:else}
    <div class="flex justify-between items-start w-full gap-4">
      <!-- Left content -->
      <div class="flex-1 min-w-0 space-y-1">
        <h4 class="text-base font-semibold text-gray-100 leading-tight truncate">
          {task.title}
        </h4>
        
        <div class="flex items-center gap-1.5 text-gray-500">
          <FolderIcon class="w-3 h-3" />
          <span class="text-xs font-medium truncate">{task.repository_name ?? 'Unknown'}</span>
        </div>

        {#if task.last_assistant_message}
          <p class="text-sm text-gray-400 leading-normal line-clamp-3 pt-1">
            {task.last_assistant_message}
          </p>
        {/if}
      </div>

      <!-- Right content (Badges) -->
      <div class="shrink-0 flex flex-col items-end gap-2">
        {#if variant === 'pending'}
          <span class="text-sm text-gray-400 px-2 py-1 rounded-sm border border-gray-800 bg-gray-900/30 font-medium"> 
            Pending 
          </span>
        {:else if variant === 'planning'}
          <span class="text-sm text-purple-400 px-2 py-1 rounded-sm border border-purple-900/50 bg-purple-900/10 font-medium">
            Planning
          </span>
        {:else if variant === 'active'}
          <div class="flex flex-col items-end gap-1">
             <span class="text-sm text-orange-400 px-2 py-1 rounded-sm border border-orange-900/50 bg-orange-900/10 font-medium whitespace-nowrap">
              In Progress
            </span>
            <span class="text-xs text-orange-400/60 font-mono tabular-nums">{elapsed}s</span>
          </div>
        {:else if variant === 'review'}
          <div class="flex items-center gap-2">
            <span class="text-sm text-blue-400 px-2 py-1 rounded-sm border border-blue-900/50 bg-blue-900/10 font-medium">
              Review
            </span>
            <ChevronRight class="w-4 h-4 text-gray-600" />
          </div>
        {/if}
      </div>
    </div>
  {/if}
</button>
