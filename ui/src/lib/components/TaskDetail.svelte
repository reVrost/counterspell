<script lang="ts">
  import { goto } from '$app/navigation';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';
  import { tasksAPI } from '$lib/api';
  import { cn } from '$lib/utils';
  import { modalSlideUp, backdropFade, DURATIONS } from '$lib/utils/transitions';
  import type { Message, Task } from '$lib/types';
  import TaskActionInput from './TaskActionInput.svelte';
  import TodoIndicator from './TodoIndicator.svelte';
  import Thread from './Thread.svelte';
  import DiffSkeleton from './DiffSkeleton.svelte';
  import DiffRenderer from './DiffRenderer.svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import TrashIcon from '@lucide/svelte/icons/trash';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import EraserIcon from '@lucide/svelte/icons/eraser';
  import GitMergeIcon from '@lucide/svelte/icons/git-merge';
  import GithubIcon from '@lucide/svelte/icons/github';
  import SparklesIcon from '@lucide/svelte/icons/sparkles';

  interface Props {
    task: Task;
    messages: Message[];
    isInProgress?: boolean;
  }

  let { task, messages, isInProgress }: Props = $props();

  // Thread rendering handled by Thread component

  // State declarations first
  let activeTab = $state<'agent' | 'diff'>('agent');
  let confirmAction = $state<string | null>(null);
  let rawDiff = $state<string>('');
  let isLoadingDiff = $state<boolean>(false);
  let containerRef = $state<HTMLDivElement | null>(null);
  let agentScrollRef = $state<HTMLDivElement | null>(null);
  let isDragging = $state(false);
  let startX = $state(0);
  let scrollY = $state(0);
  let showCompactHeader = $derived(scrollY > 60);

  function handleScroll(e: Event) {
    const target = e.target as HTMLDivElement;
    scrollY = target.scrollTop;
  }

  function handleTouchStart(e: TouchEvent) {
    isDragging = true;
    startX = e.touches[0].pageX;
  }

  function handleTouchMove(e: TouchEvent) {
    if (!isDragging) return;
    // Touch events are passive by default, can't preventDefault
  }

  function handleTouchEnd(e: TouchEvent) {
    if (!isDragging) return;
    isDragging = false;
    const endX = e.changedTouches[0].pageX;
    const diff = startX - endX;
    const threshold = 50;

    if (Math.abs(diff) > threshold) {
      if (diff > 0 && activeTab === 'agent') {
        activeTab = 'diff';
      } else if (diff < 0 && activeTab === 'diff') {
        activeTab = 'agent';
      }
    }
  }

  function handleMouseDown(e: MouseEvent) {
    isDragging = true;
    startX = e.pageX;
  }

  function handleMouseUp(e: MouseEvent) {
    if (!isDragging) return;
    isDragging = false;
    const diff = startX - e.pageX;
    const threshold = 50;

    if (Math.abs(diff) > threshold) {
      if (diff > 0 && activeTab === 'agent') {
        activeTab = 'diff';
      } else if (diff < 0 && activeTab === 'diff') {
        activeTab = 'agent';
      }
    }
  }

  interface FileStat {
    filename: string;
    additions: number;
    deletions: number;
  }

  function handleBack() {
    goto('/app');
  }

  async function handleChatSubmit(message: string, modelId: string) {
    try {
      const response = await tasksAPI.chat(task.id, message, modelId);
      if (response.message) {
        appState.showToast(response.message, 'success');
      }
    } catch (err) {
      console.error('Failed to send message:', err);
      appState.showToast(err instanceof Error ? err.message : 'Failed to send message', 'error');
    }
  }

  async function handleAction(action: string) {
    confirmAction = null;
    try {
      if (action === 'retry') {
        const response = await tasksAPI.retry(task.id);
        appState.showToast(response.message || 'Task retry started', 'success');
      } else if (action === 'clear') {
        const response = await tasksAPI.clear(task.id);
        appState.showToast(response.message || 'History cleared', 'success');
      } else if (action === 'pr') {
        const response = await tasksAPI.createPR(task.id);
        if (response.pr_url) {
          appState.showToast('Pull request created!', 'success');
          window.open(response.pr_url, '_blank');
        } else {
          appState.showToast(response.message || 'Pull request created', 'success');
        }
      } else if (action === 'merge') {
        const response = await tasksAPI.merge(task.id);
        if (response.status === 'conflict') {
          appState.showToast('Merge has conflicts - resolve them to continue', 'info');
        } else {
          appState.showToast(response.message || 'Changes merged', 'success');
        }
      } else if (action === 'review') {
        const response = await tasksAPI.chat(
          task.id,
          'Please review the changes made in this task',
          appState.activeModelId
        );
        if (response.message) {
          appState.showToast(response.message, 'success');
        }
      } else if (action === 'discard') {
        const response = await tasksAPI.discard(task.id);
        appState.showToast(response.message || 'Task discarded', 'success');
        goto('/app');
      }
    } catch (err) {
      console.error(`Failed to ${action}:`, err);
      appState.showToast(err instanceof Error ? err.message : `Failed to ${action}`, 'error');
    }
  }

  $effect(() => {
    if (activeTab === 'diff' && rawDiff === '' && !isLoadingDiff) {
      loadDiff();
    }
  });

  async function loadDiff() {
    isLoadingDiff = true;
    try {
      const response = await tasksAPI.getDiff(task.id);
      rawDiff = response.git_diff || '';
    } catch (err) {
      console.error('Failed to load diff:', err);
      rawDiff = '';
    } finally {
      isLoadingDiff = false;
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="flex flex-col h-[100dvh] relative">
  <!-- Floating Controls Row -->
  <div
    class="absolute top-0 left-0 right-0 z-30 px-3 pt-3 pb-2 flex items-center justify-between pointer-events-none transition-all duration-200"
    class:backdrop-blur-md={showCompactHeader}
    style:background-color={showCompactHeader ? 'rgba(13, 17, 23, 0.8)' : 'transparent'}
  >
    <!-- Back Button -->
    <button
      onclick={handleBack}
      class="pointer-events-auto w-9 h-9 rounded-full bg-white/5 backdrop-blur-md border border-white/10 flex items-center justify-center text-gray-400 hover:bg-white/10 hover:text-white transition-all focus:outline-none"
      aria-label="Go back"
    >
      <ArrowLeftIcon class="w-4 h-4" />
    </button>

    <!-- Title (shows when scrolled) -->
    <div
      class="flex-1 mx-3 overflow-hidden transition-all duration-200 pointer-events-auto"
      class:opacity-0={!showCompactHeader}
      class:opacity-100={showCompactHeader}
    >
      <p class="text-sm font-medium text-white/90 truncate text-center">
        {task.title}
      </p>
    </div>

    <!-- View Indicators & Status -->
    <div
      class="pointer-events-auto flex items-center gap-2.5 bg-white/5 backdrop-blur-md border border-white/10 rounded-full px-2.5 py-1.5"
    >
      <!-- Swipe Indicators -->
      <div class="flex items-center gap-1">
        <button
          onclick={() => (activeTab = 'agent')}
          class={cn(
            'w-1 h-1 rounded-full transition-all duration-200',
            activeTab === 'agent' ? 'bg-white/70 w-2' : 'bg-white/20 hover:bg-white/40'
          )}
          aria-label="Agent view"
        ></button>
        <button
          onclick={() => (activeTab = 'diff')}
          class={cn(
            'w-1 h-1 rounded-full transition-all duration-200',
            activeTab === 'diff' ? 'bg-white/70 w-2' : 'bg-white/20 hover:bg-white/40'
          )}
          aria-label="Diff view"
        ></button>
      </div>

      <div class="w-px h-3 bg-white/10"></div>

      <!-- Status Dot -->
      {#if task.status === 'draft'}
        <div class="w-1.5 h-1.5 rounded-full bg-gray-400" title="Draft"></div>
      {:else if task.status === 'in_progress'}
        <div class="w-1.5 h-1.5 rounded-full bg-orange-400 pulse-glow" title="Running"></div>
      {:else if task.status === 'review'}
        <div class="w-1.5 h-1.5 rounded-full bg-blue-400 pulse-glow" title="Ready"></div>
      {:else if task.status === 'done'}
        <div class="w-1.5 h-1.5 rounded-full bg-green-400" title="Merged"></div>
      {:else if task.status === 'failed'}
        <div class="w-1.5 h-1.5 rounded-full bg-red-400" title="Failed"></div>
      {/if}
    </div>
  </div>

  <!-- Floating Todo Indicator -->
  {#if taskStore.todos.length > 0}
    <TodoIndicator />
  {/if}

  <!-- Main Content Area - Swipeable -->
  <div
    bind:this={containerRef}
    class="flex-1 relative w-full h-full overflow-hidden"
    ontouchstart={handleTouchStart}
    ontouchmove={handleTouchMove}
    ontouchend={handleTouchEnd}
    onmousedown={handleMouseDown}
    onmouseup={handleMouseUp}
    onmouseleave={() => (isDragging = false)}
  >
    <div
      class="flex h-full transition-transform duration-300 ease-out"
      style:transform="translateX({activeTab === 'agent' ? '0%' : '-100%'})"
    >
      <!-- Agent View -->
      <div
        bind:this={agentScrollRef}
        class="w-full h-full flex-shrink-0 overflow-y-auto"
        id="agent-scroll"
        onscroll={handleScroll}
      >
        <!-- Header - Scrolls with content -->
        <div class="px-4 pt-20 pb-6 flex flex-col gap-2">
          <!-- Workspace -->
          <span class="text-gray-500 text-[11px] flex items-center gap-1.5">
            <i class="fas fa-folder text-[10px]"></i>
            {task.workspace_name || 'Unknown'}
          </span>

          <!-- Title -->
          <h1 class="text-xl font-semibold text-white/95 leading-tight">
            {task.title}
          </h1>
        </div>

        <div class="space-y-1 pb-44">
          <Thread
            mode="task"
            {messages}
            emptyText="No agent output"
            emptyClass="p-5 text-gray-500 italic text-xs"
            scrollContainerId="agent-scroll"
          />

          {#if isInProgress}
            <div class="flex items-center gap-3 px-8 py-4">
              <p class="text-base font-medium shimmer" style="color: #7950f2;">Thinking...</p>
              <p class="text-base" style="color: #7950f2; opacity: 0.6;">Analyzing context</p>
            </div>
          {/if}
        </div>
      </div>

      <!-- Diff View -->
      <div class="w-full h-full flex-shrink-0 overflow-y-auto" id="diff-scroll">
        <div class="min-h-full pb-32">
          <div class="p-3 diff-container">
            {#if isLoadingDiff}
              <DiffSkeleton />
            {:else}
              <DiffRenderer diff={rawDiff} class="w-full" />
            {/if}
          </div>
        </div>
      </div>
    </div>

    <!-- Swipe Hint - Shows briefly on first load -->
    <div
      class="absolute bottom-24 left-1/2 -translate-x-1/2 pointer-events-none opacity-0 transition-opacity duration-500"
      class:opacity-100={activeTab === 'agent'}
    >
      <div class="flex items-center gap-2 text-[10px] text-white/30 uppercase tracking-wider">
        <span>Swipe for diff</span>
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7" />
        </svg>
      </div>
    </div>
  </div>

  <!-- Unified Action Input - Mobile native design -->
  <div class="absolute bottom-0 inset-x-0 z-20 pb-6 px-3 mb-6 pb-safe">
    <div
      class="absolute inset-0 bg-gradient-to-t from-[#0D1117] via-[#0D1117]/95 to-transparent pointer-events-none"
    ></div>
    <div class="relative mx-auto max-w-4xl">
      <TaskActionInput
        taskId={task.id}
        taskStatus={task.status}
        placeholder={task.status === 'done' ? 'Task completed' : 'Ask for changes or approve...'}
        onSubmit={handleChatSubmit}
        onMerge={() => (confirmAction = 'merge')}
        onCreatePR={() => (confirmAction = 'pr')}
        onDelete={() => (confirmAction = 'discard')}
      />
    </div>
  </div>

  <!-- Confirmation Modal -->
  {#if confirmAction}
    <div
      transition:backdropFade|local={{ duration: DURATIONS.normal }}
      class="fixed inset-0 z-[200] flex items-start justify-center pt-[25vh] bg-black/60 backdrop-blur-sm"
      role="button"
      tabindex="-1"
      onclick={(e) => e.target === e.currentTarget && (confirmAction = null)}
      onkeydown={(e) => e.key === 'Escape' && (confirmAction = null)}
      aria-label="Close confirmation dialog"
    >
      <div
        transition:modalSlideUp|local={{ duration: DURATIONS.quick }}
        class="bg-popover border border-gray-700/50 rounded-xl p-5 w-[320px] shadow-2xl"
      >
        {#if confirmAction === 'retry'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-amber-500/10 flex items-center justify-center">
                <RotateCcwIcon class="w-5 h-5 text-amber-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Retry Task</h3>
                <p class="text-xs text-gray-400">Re-run with the same prompt</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will retry the previous prompt and overwrite any existing changes.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('retry')}
                class="flex-1 h-9 rounded-lg bg-amber-600 hover:bg-amber-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Retry
              </button>
            </div>
          </div>
        {:else if confirmAction === 'clear'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center">
                <EraserIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Clear History</h3>
                <p class="text-xs text-gray-400">Reset memory and context</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will clear the chat history and agent output. The task will start fresh without
              any prior context.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('clear')}
                class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Clear
              </button>
            </div>
          </div>
        {:else if confirmAction === 'pr'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-violet-500/10 flex items-center justify-center">
                <GithubIcon class="w-5 h-5 text-violet-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Create Pull Request</h3>
                <p class="text-xs text-gray-400">Push changes to GitHub</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will create a new pull request on GitHub with all the changes from this task.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('pr')}
                class="flex-1 h-9 rounded-lg bg-violet-600 hover:bg-violet-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Create PR
              </button>
            </div>
          </div>
        {:else if confirmAction === 'merge'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-green-500/10 flex items-center justify-center">
                <GitMergeIcon class="w-5 h-5 text-green-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Merge to Main</h3>
                <p class="text-xs text-gray-400">Apply changes directly</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will merge all changes directly into the main branch without creating a pull
              request.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('merge')}
                class="flex-1 h-9 rounded-lg bg-green-600 hover:bg-green-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Merge
              </button>
            </div>
          </div>
        {:else if confirmAction === 'review'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-violet-500/10 flex items-center justify-center">
                <SparklesIcon class="w-5 h-5 text-violet-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Request AI Review</h3>
                <p class="text-xs text-gray-400">Get code review feedback</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will request an AI code review of the changes made in this task.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('review')}
                class="flex-1 h-9 rounded-lg bg-violet-600 hover:bg-violet-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Request Review
              </button>
            </div>
          </div>
        {:else if confirmAction === 'discard'}
          <div class="space-y-4">
            <div class="flex items-center gap-3">
              <div class="w-10 h-10 rounded-full bg-red-500/10 flex items-center justify-center">
                <TrashIcon class="w-5 h-5 text-red-500" />
              </div>
              <div>
                <h3 class="font-semibold text-[#FFFFFF]">Discard Task</h3>
                <p class="text-xs text-gray-400">Permanently delete task</p>
              </div>
            </div>
            <p class="text-sm text-gray-300">
              This will permanently delete this task and all its data. This action cannot be undone.
            </p>
            <div class="flex gap-2 pt-2">
              <button
                onclick={() => (confirmAction = null)}
                class="flex-1 h-9 rounded-lg bg-gray-700/50 hover:bg-gray-700 text-gray-300 text-sm font-medium transition-colors"
              >
                Cancel
              </button>
              <button
                onclick={() => handleAction('discard')}
                class="flex-1 h-9 rounded-lg bg-red-600 hover:bg-red-500 text-[#FFFFFF] text-sm font-medium transition-colors"
              >
                Discard
              </button>
            </div>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>
