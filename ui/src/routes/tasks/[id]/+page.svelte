<script lang="ts">
  import { taskStore } from '$lib/stores/tasks.svelte';
  import { tasksAPI } from '$lib/api';
  import { createTaskSSE } from '$lib/utils/sse';
  import type { TaskResponse, Message, Task, LogEntry } from '$lib/types';
  import TaskDetail from '$lib/components/TaskDetail.svelte';
  import Skeleton from '$lib/components/Skeleton.svelte';
  import { page } from '$app/stores';
  import { onDestroy } from 'svelte';

  let task = $state<TaskResponse | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let messages = $state<Message[]>([]);
  let diffContent = $state('');
  let logContent = $state<string[]>([]);
  let eventSource: EventSource | null = null;

  function applyTaskData(taskData: TaskResponse) {
    task = taskData;
    messages = taskData.messages || [];
    diffContent = taskData.git_diff
      ? renderDiffHTML(taskData.git_diff)
      : '<div class="text-gray-500 italic">No changes made</div>';
    logContent = [];
    taskStore.currentTask = taskData.task;
  }

  async function loadTask(taskId: string) {
    loading = true;
    error = null;

    try {
      const taskData = await tasksAPI.get(taskId);
      applyTaskData(taskData);
      setupSSE(taskId);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load task';
      console.error('Task load error:', err);
    } finally {
      loading = false;
    }
  }

  function setupSSE(taskId: string) {
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    eventSource = createTaskSSE(taskId, {
      onAgentUpdate: (data: string) => {
        try {
          const parsed = JSON.parse(data);
          if (Array.isArray(parsed)) {
            // Update messages in-place
            messages.push(...parsed); // append only
          }
        } catch (e) {
          console.error('Failed to parse agent update JSON:', e);
        }
      },
      onRunUpdate: () => {
        loadTask(taskId);
      },
      onDiffUpdate: (html: string) => {
        diffContent = html;
      },
      onLog: (html: string) => {
        // Do nothing
        // logContent = [...logContent, html];
      },
      onStatus: (html: string) => {},
      onComplete: (status: string) => {
        if (task) {
          task = { ...task, task: { ...task.task, status: status as Task['status'] } };
          taskStore.currentTask = task.task;
        }
      },
      onError: (err) => {
        console.error('Task SSE error:', err);
      },
    });
  }

  function renderDiffHTML(diff: string): string {
    if (!diff) return '<div class="text-gray-500 italic">No changes made</div>';

    let html = '';
    for (const line of diff.split('\n')) {
      const escapedLine = escapeHtml(line);
      if (line.startsWith('+')) {
        html += `<div class="px-3 py-1 bg-green-500/10 text-green-400 font-mono text-sm border-l-2 border-green-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith('-')) {
        html += `<div class="px-3 py-1 bg-red-500/10 text-red-400 font-mono text-sm border-l-2 border-red-500/50">${escapedLine.substring(1)}</div>`;
      } else if (line.startsWith('@@')) {
        html += `<div class="px-3 py-1 bg-gray-800 text-gray-500 font-mono text-sm">${escapedLine}</div>`;
      } else if (line.trim() !== '') {
        html += `<div class="px-3 py-1 text-gray-400 font-mono text-sm">${escapedLine}</div>`;
      }
    }
    return html;
  }

  function escapeHtml(text: string): string {
    return text
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#39;');
  }

  $effect(() => {
    const taskId = $page.params.id;
    if (!taskId) {
      loading = false;
      error = 'Missing task id';
      return;
    }

    loadTask(taskId);
  });

  onDestroy(() => {
    if (eventSource) {
      eventSource.close();
    }
  });
</script>

<svelte:head>
  <title>{task?.task.title || 'Task'} - Counterspell</title>
</svelte:head>

<div class="min-h-screen bg-background flex flex-col">
  <!-- Task Detail Content -->
  <div class="flex-1 overflow-hidden">
    {#if loading}
      <!-- Skeleton screens matching TaskDetail layout -->
      <div class="flex flex-col h-full">
        <!-- Tabs skeleton -->
        <div class="p-2 flex justify-between">
          <div class="w-6"></div>
          <div class="flex gap-2">
            {#each [1, 2, 3] as _}
              <Skeleton variant="rounded" class="h-8 w-16" />
            {/each}
          </div>
          <div class="w-6"></div>
        </div>

        <!-- Content skeleton -->
        <div class="flex-1 p-4 space-y-4">
          {#each [1, 2, 3, 4] as _}
            <div class="space-y-2">
              <Skeleton variant="text" class="w-full" />
              <Skeleton variant="text" class="w-5/6" />
              <Skeleton variant="text" class="w-3/4" />
            </div>
          {/each}
        </div>

        <!-- Bottom actions skeleton -->
        <div class="px-4 py-3 border-t border-white/5">
          <div class="flex gap-2">
            <Skeleton variant="rounded" class="flex-1 h-12" />
            <Skeleton variant="rounded" class="flex-1 h-12" />
          </div>
        </div>
      </div>
    {:else if error}
      <div class="flex items-center justify-center h-full">
        <div class="text-center">
          <p class="text-base text-red-400 mb-2">{error}</p>
          <button
            onclick={() => loadTask()}
            class="px-4 py-2 bg-violet-500/20 border border-violet-500/30 rounded-lg text-sm text-violet-300 hover:bg-violet-500/30 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    {:else if task}
      <TaskDetail
        task={task.task}
        {messages}
        {logContent}
        isInProgress={task.task.status === 'in_progress'}
      />
    {/if}
  </div>
</div>

<style>
  :global(summary::-webkit-details-marker) {
    display: none;
  }

  :global(.shimmer) {
    background: linear-gradient(
      90deg,
      rgba(255, 255, 255, 0.05) 25%,
      rgba(255, 255, 255, 0.1) 50%,
      rgba(255, 255, 255, 0.05) 75%
    );
    background-size: 200% 100%;
    animation: shimmer 2s infinite;
    -webkit-background-clip: text;
    -webkit-text-fill-color: transparent;
  }

  @keyframes shimmer {
    0% {
      background-position: 200% 0;
    }
    100% {
      background-position: -200% 0;
    }
  }

  :global(.pulse-glow) {
    animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 1;
      filter: drop-shadow(0 0 2px rgba(139, 92, 246, 0.5));
    }
    50% {
      opacity: 0.7;
      filter: drop-shadow(0 0 5px rgba(139, 92, 246, 0.8));
    }
  }
</style>
