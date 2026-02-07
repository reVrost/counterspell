<script lang="ts">
  import { tasksAPI } from '$lib/api';
  import { appState } from '$lib/stores/app.svelte';
  import { cn } from '$lib/utils';
  import { Input } from '$lib/components/ui/input';
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
</script>

<div class="animate-in fade-in duration-300 space-y-8 max-w-3xl p-4">
  <section>
    <h2
      class="text-sm font-bold text-muted-foreground uppercase tracking-wider mb-3 px-1 flex items-center gap-2"
    >
      <SearchIcon class="w-3.5 h-3.5" /> Search Tasks
    </h2>
    <div class="relative">
      <SearchIcon
        class="absolute left-4 top-1/2 -translate-y-1/2 w-5 h-5 text-muted-foreground/50 pointer-events-none"
      />
      <Input
        bind:value={searchQuery}
        type="text"
        placeholder="Search by title, description, or workspace..."
        class="w-full h-12 pl-12 pr-12 bg-background/50 focus:bg-background transition-all border-border/50 focus:border-violet-500/50"
      />
      {#if searchQuery}
        <button
          onclick={clearSearch}
          class="absolute right-4 top-1/2 -translate-y-1/2 p-1 rounded-full hover:bg-white/5 text-muted-foreground hover:text-foreground transition-colors"
        >
          <XIcon class="w-4 h-4" />
        </button>
      {/if}
    </div>
  </section>

  <section>
    {#if isLoading}
      <div class="flex items-center justify-center py-24">
        <div class="flex flex-col items-center gap-3">
          <Loader2Icon class="w-8 h-8 text-violet-400 animate-spin" />
          <p class="text-base text-muted-foreground">Loading tasks...</p>
        </div>
      </div>
    {:else if error}
      <div class="bg-card border border-border/50 rounded-xl p-8 text-center shadow-sm">
        <div
          class="w-12 h-12 bg-red-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-red-500/20"
        >
          <AlertCircleIcon class="w-6 h-6 text-red-400" />
        </div>
        <p class="text-muted-foreground mb-4">{error}</p>
        <button
          onclick={loadTasks}
          class="px-6 py-2.5 bg-violet-600 hover:bg-violet-500 rounded-full text-base font-medium transition-all shadow-lg shadow-violet-500/20"
        >
          Retry
        </button>
      </div>
    {:else if filteredTasks.length === 0}
      <div class="bg-card border border-border/50 rounded-xl p-10 text-center shadow-sm">
        {#if searchQuery}
          <div
            class="w-12 h-12 bg-zinc-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-zinc-500/20"
          >
            <SearchIcon class="w-6 h-6 text-zinc-400" />
          </div>
          <h3 class="text-foreground font-medium mb-1">No results found</h3>
          <p class="text-base text-muted-foreground">Try adjusting your search or filters</p>
        {:else}
          <div
            class="w-12 h-12 bg-zinc-500/10 rounded-full flex items-center justify-center mx-auto mb-4 border border-zinc-500/20"
          >
            <InboxIcon class="w-6 h-6 text-zinc-400" />
          </div>
          <h3 class="text-foreground font-medium mb-1">No tasks yet</h3>
          <p class="text-base text-muted-foreground">Create your first task to get started</p>
        {/if}
      </div>
    {:else}
      <h3 class="text-sm font-bold text-muted-foreground uppercase tracking-wider mb-3 px-1">
        {filteredTasks.length}
        {filteredTasks.length === 1 ? 'result' : 'results'}{searchQuery
          ? ` for "${searchQuery}"`
          : ' total'}
      </h3>

      <div class="space-y-2">
        {#each filteredTasks as task (task.id)}
          {@const StatusIcon = getStatusIcon(task.status)}
          {@const statusBg =
            task.status === 'done'
              ? 'bg-emerald-500/10 border-emerald-500/20'
              : task.status === 'in_progress'
                ? 'bg-violet-500/10 border-violet-500/20'
                : task.status === 'review'
                  ? 'bg-amber-500/10 border-amber-500/20'
                  : task.status === 'failed'
                    ? 'bg-red-500/10 border-red-500/20'
                    : task.status === 'planning'
                      ? 'bg-blue-500/10 border-blue-500/20'
                      : 'bg-zinc-500/10 border-zinc-500/20'}
          <button
            onclick={() => handleTaskClick(task)}
            class="w-full text-left p-4 bg-card/60 hover:bg-card border border-border/50 rounded-xl transition-all duration-200 group hover:border-violet-500/30 hover:shadow-sm"
          >
            <div class="flex items-start gap-3">
              <!-- Status Icon -->
              <div
                class={cn(
                  'w-9 h-9 rounded-lg flex items-center justify-center shrink-0 border',
                  statusBg
                )}
              >
                <StatusIcon class={cn('w-4.5 h-4.5', getStatusColor(task.status))} />
              </div>

              <!-- Content -->
              <div class="flex-1 min-w-0">
                <div class="flex items-start justify-between gap-3">
                  <h3
                    class="font-medium text-foreground text-[15px] leading-snug group-hover:text-violet-300 transition-colors"
                  >
                    {task.title}
                  </h3>
                  <span class="text-[11px] text-muted-foreground shrink-0 font-medium">
                    {formatTime(task.updated_at)}
                  </span>
                </div>

                {#if task.workspace_name}
                  <div class="flex items-center gap-1.5 mt-1">
                    <div class="w-1 h-1 rounded-full bg-muted-foreground/40"></div>
                    <p class="text-sm text-muted-foreground font-medium">
                      {task.workspace_name}
                    </p>
                  </div>
                {/if}

                {#if task.intent}
                  <p class="text-base text-muted-foreground/70 mt-2 line-clamp-2 leading-relaxed">
                    {task.intent}
                  </p>
                {/if}
              </div>
            </div>
          </button>
        {/each}
      </div>
    {/if}
  </section>
</div>
