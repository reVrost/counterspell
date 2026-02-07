<script lang="ts">
  import { tasksAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import type { Task, FeedData } from '$lib/types';
  import SearchIcon from '@lucide/svelte/icons/search';
  import XIcon from '@lucide/svelte/icons/x';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import ClockIcon from '@lucide/svelte/icons/clock';
  import CheckCircleIcon from '@lucide/svelte/icons/check-circle';
  import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
  import PlayCircleIcon from '@lucide/svelte/icons/play-circle';
  import FileTextIcon from '@lucide/svelte/icons/file-text';
  import Loader2Icon from '@lucide/svelte/icons/loader-2';

  let searchQuery = $state('');
  let isLoading = $state(true);
  let allTasks = $state<Task[]>([]);
  let error = $state<string | null>(null);

  // Load all tasks on mount
  $effect(() => {
    loadTasks();
  });

  async function loadTasks() {
    isLoading = true;
    error = null;
    try {
      const feed = await tasksAPI.getFeed();
      // Combine all tasks from feed
      allTasks = [
        ...(feed.active || []),
        ...(feed.reviews || []),
        ...(feed.done || []),
        ...(feed.todo || []),
        ...(feed.planning || []),
      ];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load tasks';
      console.error('Error loading tasks:', err);
    } finally {
      isLoading = false;
    }
  }

  // Filter tasks based on search query
  const filteredTasks = $derived.by(() => {
    if (!searchQuery.trim()) return allTasks;
    const query = searchQuery.toLowerCase();
    return allTasks.filter(
      (task) =>
        task.title.toLowerCase().includes(query) ||
        task.intent.toLowerCase().includes(query) ||
        task.workspace_name?.toLowerCase().includes(query)
    );
  });

  function getStatusIcon(status: Task['status']) {
    switch (status) {
      case 'done':
        return CheckCircleIcon;
      case 'in_progress':
        return PlayCircleIcon;
      case 'review':
        return FileTextIcon;
      case 'failed':
        return AlertCircleIcon;
      case 'planning':
        return ClockIcon;
      default:
        return InboxIcon;
    }
  }

  function getStatusColor(status: Task['status']) {
    switch (status) {
      case 'done':
        return 'text-emerald-400';
      case 'in_progress':
        return 'text-violet-400';
      case 'review':
        return 'text-amber-400';
      case 'failed':
        return 'text-red-400';
      case 'planning':
        return 'text-blue-400';
      default:
        return 'text-zinc-400';
    }
  }

  function formatTime(timestamp: number) {
    const date = new Date(timestamp * 1000);
    const now = new Date();
    const diff = now.getTime() - date.getTime();
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));

    if (days === 0) {
      const hours = Math.floor(diff / (1000 * 60 * 60));
      if (hours === 0) {
        const minutes = Math.floor(diff / (1000 * 60));
        return minutes <= 1 ? 'Just now' : `${minutes}m ago`;
      }
      return `${hours}h ago`;
    } else if (days === 1) {
      return 'Yesterday';
    } else if (days < 7) {
      return `${days}d ago`;
    } else {
      return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' });
    }
  }

  function handleTaskClick(task: Task) {
    appState.openModal(task.id);
  }

  function clearSearch() {
    searchQuery = '';
  }

  let inputElement: HTMLInputElement | null = null;

  $effect(() => {
    // Auto-focus search input on mount
    if (inputElement) {
      inputElement.focus();
    }
  });
</script>

<div class="animate-in fade-in duration-300 space-y-6">
  <!-- Search Header -->
  <div class="sticky top-0 z-10 -mx-3 px-3 py-4 bg-background/80 backdrop-blur-md">
    <div class="relative">
      <SearchIcon
        class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground pointer-events-none"
      />
      <input
        bind:this={inputElement}
        bind:value={searchQuery}
        type="text"
        placeholder="Search tasks..."
        class="w-full h-12 pl-12 pr-12 bg-card border border-border/50 rounded-xl text-foreground placeholder:text-muted-foreground focus:outline-none focus:border-violet-500/50 focus:ring-2 focus:ring-violet-500/20 transition-all text-base"
      />
      {#if searchQuery}
        <button
          onclick={clearSearch}
          class="absolute right-4 top-1/2 -translate-y-1/2 p-1 rounded-full hover:bg-white/10 text-muted-foreground hover:text-foreground transition-colors"
        >
          <XIcon class="w-4 h-4" />
        </button>
      {/if}
    </div>
  </div>

  <!-- Results -->
  <div class="space-y-2">
    {#if isLoading}
      <div class="flex items-center justify-center py-20">
        <Loader2Icon class="w-8 h-8 text-violet-400 animate-spin" />
      </div>
    {:else if error}
      <div class="text-center py-20">
        <AlertCircleIcon class="w-12 h-12 text-red-400 mx-auto mb-4" />
        <p class="text-muted-foreground">{error}</p>
        <button
          onclick={loadTasks}
          class="mt-4 px-4 py-2 bg-violet-600 hover:bg-violet-500 rounded-lg text-sm font-medium transition-colors"
        >
          Retry
        </button>
      </div>
    {:else if filteredTasks.length === 0}
      <div class="text-center py-20">
        {#if searchQuery}
          <SearchIcon class="w-12 h-12 text-muted-foreground/50 mx-auto mb-4" />
          <p class="text-muted-foreground">No tasks found for "{searchQuery}"</p>
          <p class="text-sm text-muted-foreground/70 mt-1">Try a different search term</p>
        {:else}
          <InboxIcon class="w-12 h-12 text-muted-foreground/50 mx-auto mb-4" />
          <p class="text-muted-foreground">No tasks yet</p>
          <p class="text-sm text-muted-foreground/70 mt-1">Create your first task to get started</p>
        {/if}
      </div>
    {:else}
      <!-- Results count -->
      <p class="text-xs text-muted-foreground px-1 mb-3">
        {filteredTasks.length}
        {filteredTasks.length === 1 ? 'task' : 'tasks'}{searchQuery
          ? ` matching "${searchQuery}"`
          : ''}
      </p>

      <!-- Task List -->
      <div class="space-y-1">
        {#each filteredTasks as task (task.id)}
          {@const StatusIcon = getStatusIcon(task.status)}
          <button
            onclick={() => handleTaskClick(task)}
            class="w-full text-left p-4 bg-card/50 hover:bg-card border border-border/30 hover:border-border/60 rounded-xl transition-all duration-200 group"
          >
            <div class="flex items-start gap-3">
              <!-- Status Icon -->
              <div
                class={cn(
                  'w-8 h-8 rounded-lg flex items-center justify-center shrink-0 mt-0.5',
                  task.status === 'done' && 'bg-emerald-500/10',
                  task.status === 'in_progress' && 'bg-violet-500/10',
                  task.status === 'review' && 'bg-amber-500/10',
                  task.status === 'failed' && 'bg-red-500/10',
                  task.status === 'planning' && 'bg-blue-500/10',
                  task.status === 'draft' && 'bg-zinc-500/10'
                )}
              >
                <StatusIcon class={cn('w-4 h-4', getStatusColor(task.status))} />
              </div>

              <!-- Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-start justify-between gap-2">
                  <h3
                    class="font-medium text-foreground truncate group-hover:text-violet-300 transition-colors"
                  >
                    {task.title}
                  </h3>
                  <span class="text-[11px] text-muted-foreground shrink-0">
                    {formatTime(task.updated_at)}
                  </span>
                </div>

                {#if task.workspace_name}
                  <p class="text-xs text-muted-foreground mt-0.5 truncate">
                    {task.workspace_name}
                  </p>
                {/if}

                {#if task.intent}
                  <p class="text-sm text-muted-foreground/80 mt-1.5 line-clamp-2 leading-relaxed">
                    {task.intent}
                  </p>
                {/if}
              </div>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </div>
</div>
