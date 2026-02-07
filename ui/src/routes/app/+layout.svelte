<script lang="ts">
  import Header from '$lib/components/Header.svelte';
  import Toast from '$lib/components/Toast.svelte';
  import Navigator from '$lib/components/Navigator.svelte';
  import TaskDetail from '$lib/components/TaskDetail.svelte';
  import TaskDetailSkeleton from '$lib/components/TaskDetailSkeleton.svelte';
  import ChatInput from '$lib/components/ChatInput.svelte';
  import NewTaskModal from '$lib/components/NewTaskModal.svelte';
  import NewWorkspaceModal from '$lib/components/NewWorkspaceModal.svelte';
  import { appState } from '$lib/stores/app.svelte';
  import { taskStore } from '$lib/stores/tasks.svelte';
  import { tasksAPI } from '$lib/api';
  import { createTaskSSE } from '$lib/utils/sse';
  import { modalSlideUp, backdropFade, DURATIONS } from '$lib/utils/transitions';
  import AlertCircleIcon from '@lucide/svelte/icons/alert-circle';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import type { Workspace, Task, Message, LogEntry } from '$lib/types';
  import { onDestroy, tick } from 'svelte';

  let { children } = $props();

  // Task detail state for modal
  let currentTask = $state<Task | null>(null);
  let loadingTask = $state(false);
  let taskError = $state<string | null>(null);
  let currentMessages = $state<Message[]>([]);
  let logContent = $state<string[]>([]);
  let eventSource: EventSource | null = null;

  function parseError(err: string) {
    if (err.startsWith('API error:')) {
      const parts = err.split(' - ');
      if (parts.length > 1) {
        try {
          const json = JSON.parse(parts[1]);
          return {
            status: parts[0].replace('API error: ', ''),
            message: json.message || json.error || 'Internal Server Error',
            code: json.code || 'INTERNAL_ERROR',
          };
        } catch {
          return {
            status: parts[0].replace('API error: ', ''),
            message: parts[1],
            code: 'ERROR',
          };
        }
      }
    }
    return { status: 'Error', message: err, code: 'ERROR' };
  }

  // Cache for loaded tasks
  let taskCache = $state<
    Map<
      string,
      {
        task: Task;
        workspace: Workspace;
        messages: Message[];
        logs: LogEntry[];
      }
    >
  >(new Map());

  // Prefetch on hover
  let prefetchTimeout: number | null = null;

  function prefetchTask(taskId: string) {
    if (taskCache.has(taskId) || prefetchTimeout !== null) return;

    prefetchTimeout = window.setTimeout(() => {
      loadTaskDetail(taskId, true);
    }, 150); // Small delay to avoid fetching on quick passes
  }

  async function loadTaskDetail(taskId: string, isPrefetch = false) {
    if (!taskId) return;

    // Check cache first
    if (taskCache.has(taskId)) {
      const cached = taskCache.get(taskId)!;
      if (!isPrefetch) {
        currentTask = cached.task;
        currentMessages = cached.messages;
        logContent = cached.logs.map((log) => renderLogEntryHTML(log));
        setupSSE(taskId);
      }
      return;
    }

    if (!isPrefetch) loadingTask = true;
    taskError = null;

    try {
      const data = await tasksAPI.get(taskId);

      // Cache the result
      taskCache.set(taskId, {
        task: data.task,
        workspace: data.workspace!,
        messages: data.messages || [],
        logs: data.logs || [],
      });

      if (!isPrefetch) {
        currentTask = data.task;
        taskStore.currentTask = data.task;
        currentMessages = data.messages || [];
        logContent = data.logs?.map((log) => renderLogEntryHTML(log)) || [];

        // Set up SSE for real-time updates
        setupSSE(taskId);
      }
    } catch (err) {
      if (!isPrefetch) {
        taskError = err instanceof Error ? err.message : 'Failed to load task';
        console.error('Task load error:', err);
      }
    } finally {
      if (!isPrefetch) loadingTask = false;
    }
  }

  function setupSSE(taskId: string) {
    // Close existing SSE connection
    if (eventSource) {
      eventSource.close();
      eventSource = null;
    }

    // Set up SSE for real-time updates
    eventSource = createTaskSSE(taskId, {
      onAgentUpdate: (data: string) => {
        try {
          const parsed = JSON.parse(data);
          if (Array.isArray(parsed)) {
            currentMessages.push(...parsed);
          }
        } catch (e) {
          console.error('Failed to parse agent update JSON:', e);
        }
      },
      onDiffUpdate: (html: string) => {
        // Diff is now loaded on-demand in TaskDetail
      },
      onLog: (html: string) => {
        logContent = [...logContent, html];
      },
      onStatus: (html: string) => {
        // Status indicator updated
      },
      onComplete: (status: string) => {
        if (currentTask) {
          currentTask = {
            ...currentTask,
            status: status as Task['status'],
          };
          taskStore.currentTask = currentTask;
        }
      },
      onError: (error) => {
        console.error('Task SSE error:', error);
      },
    });
  }

  function renderLogEntryHTML(log: LogEntry): string {
    return `
			<div class="ml-4 relative">
				<div class="absolute -left-[21px] top-1 h-2.5 w-2.5 rounded-full border border-[#0D1117] bg-blue-500"></div>
				<p class="text-xs text-gray-400">${escapeHtml(log.message)}</p>
			</div>
		`;
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
    if (!appState.modalOpen || !appState.modalTaskId) {
      // Close SSE when modal closes
      if (eventSource) {
        eventSource.close();
        eventSource = null;
      }
      logContent = [];
      currentMessages = [];
      currentTask = null;
    } else {
      loadTaskDetail(appState.modalTaskId);
    }
  });

  onDestroy(() => {
    if (eventSource) {
      eventSource.close();
    }
    if (prefetchTimeout) {
      clearTimeout(prefetchTimeout);
    }
  });

  // Expose prefetch function globally for TaskRow
  if (typeof window !== 'undefined') {
    (window as any).prefetchTask = prefetchTask;
  }
</script>

<div class="h-[100dvh] flex flex-col overflow-hidden bg-background">
  <Toast />
  <Header />

  <main class="flex-1 overflow-y-auto bg-background relative pt-16" id="feed-container">
    <div class="px-3 pt-6 pb-40">{@render children()}</div>
  </main>

  <!-- Bottom Navigation Bar -->
  <div class="fixed bottom-4 left-4 right-4 z-20 mx-auto max-w-4xl grid items-end mb-safe">
    {#if appState.showChatInput}
      <div class="col-start-1 row-start-1 w-full relative z-50">
        <ChatInput mode="create" onClose={() => appState.closeChatInput()} />
      </div>
    {:else}
      <div class="col-start-1 row-start-1 w-full relative z-10">
        <Navigator
          activeTab={appState.activeTab}
          onNavigate={(tab) => (appState.activeTab = tab as any)}
        />
      </div>
    {/if}
  </div>

  <!-- New Task Modal -->
  {#if appState.showNewTaskModal}
    <NewTaskModal />
  {/if}

  {#if appState.showNewWorkspaceModal}
    <NewWorkspaceModal />
  {/if}

  <!-- Task Detail Modal -->
  {#if appState.modalOpen && appState.modalTaskId}
    <div
      transition:backdropFade|global={{ duration: DURATIONS.normal }}
      class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm"
      role="presentation"
      aria-hidden="true"
    ></div>
    <div
      class="fixed inset-0 z-50 flex items-end sm:items-center justify-center pointer-events-none"
    >
      <div
        transition:modalSlideUp|global={{ duration: DURATIONS.normal }}
        class="pointer-events-auto bg-popover flex flex-col overflow-hidden w-full h-full sm:h-auto sm:max-w-4xl sm:max-h-[85vh] sm:rounded-2xl border border-white/10 shadow-2xl"
        role="dialog"
        aria-modal="true"
        aria-labelledby="modal-title"
      >
        {#if loadingTask && !currentTask}
          <TaskDetailSkeleton />
        {:else if taskError}
          {@const errorData = parseError(taskError)}
          <div
            class="flex flex-col items-center justify-center h-full p-8 text-center animate-in fade-in duration-500 scale-in-95"
          >
            <div
              class="w-16 h-16 rounded-full bg-red-500/10 border border-red-500/20 flex items-center justify-center mb-6 shadow-[0_0_20px_rgba(239,68,68,0.1)]"
            >
              <AlertCircleIcon class="w-8 h-8 text-red-500" />
            </div>

            <h3 class="text-lg font-bold text-white mb-2 tracking-tight">Something went wrong</h3>
            <p class="text-sm text-zinc-400 max-w-sm mb-6 leading-relaxed">
              {errorData.message}
            </p>

            <div
              class="flex items-center gap-3 mb-8 px-4 py-2 bg-white/[0.03] border border-white/5 rounded-xl"
            >
              <span class="text-[10px] font-mono font-bold text-zinc-500 uppercase tracking-widest"
                >Code</span
              >
              <span
                class="text-[10px] font-mono text-zinc-400 font-medium tracking-tight bg-white/5 px-2 py-0.5 rounded"
                >{errorData.code}</span
              >
              <div class="w-px h-3 bg-white/10 mx-1"></div>
              <span class="text-[10px] font-mono font-bold text-zinc-500 uppercase tracking-widest"
                >Status</span
              >
              <span
                class="text-[10px] font-mono text-zinc-400 font-medium tracking-tight bg-white/5 px-2 py-0.5 rounded"
                >{errorData.status}</span
              >
            </div>

            <button
              onclick={() => loadTaskDetail(appState.modalTaskId!)}
              class="flex items-center gap-2 px-6 py-2.5 bg-violet-600 hover:bg-violet-500 border border-violet-400/30 rounded-full text-sm font-semibold text-white transition-all shadow-lg active:scale-95 group"
            >
              <RefreshCwIcon
                class="w-4 h-4 group-hover:rotate-180 transition-transform duration-500"
              />
              Retry Connection
            </button>
          </div>
        {:else if currentTask}
          <TaskDetail
            task={currentTask}
            messages={currentMessages}
            {logContent}
            isInProgress={currentTask.status === 'in_progress'}
          />
        {/if}
      </div>
    </div>
  {/if}
</div>
